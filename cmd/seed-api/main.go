package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-faker/faker/v4"
)

const (
	defaultAdminEmail    = "admin@test.fr"
	defaultAdminPassword = "admin@test.fr"

	defaultClassCount        = 3
	defaultStudentsPerClass  = 20
	defaultBonusChance       = 0.30
	defaultPenaltyChance     = 0.25
	defaultPunishmentChance  = 0.15
	defaultMaxBonuses        = 3
	defaultMaxPenalties      = 3
	defaultMaxPunishments    = 2
	defaultRequestTimeout    = 20 * time.Second
	defaultExecutionTimeout  = 5 * time.Minute
	defaultYearLabel         = "2025-2026"
	randomUserCreateAttempts = 5
)

type seedConfig struct {
	BaseURL                string
	Email                  string
	Password               string
	ClassCount             int
	StudentsPerClass       int
	BonusChance            float64
	PenaltyChance          float64
	PunishmentChance       float64
	MaxBonusesPerStudent   int
	MaxPenaltiesPerStudent int
	MaxPunishments         int
	Timeout                time.Duration
	ExecutionTimeout       time.Duration
	CreateRandomUser       bool
}

type apiClient struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
}

type registerStatusResponse struct {
	RegisterAllowed bool `json:"register_allowed"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type apiErrorDetail struct {
	Field string `json:"field"`
	Error string `json:"error"`
	Value string `json:"value"`
}

type apiErrorResponse struct {
	Error        string           `json:"error"`
	ErrorCode    int              `json:"error_code"`
	ErrorDetails []apiErrorDetail `json:"error_details"`
}

type apiRequestError struct {
	StatusCode int
	Path       string
	APIError   string
	Body       string
}

func (e *apiRequestError) Error() string {
	msg := fmt.Sprintf("request failed: status=%d path=%s", e.StatusCode, e.Path)
	if e.APIError != "" {
		msg = fmt.Sprintf("%s error=%s", msg, e.APIError)
	}
	if e.Body != "" {
		msg = fmt.Sprintf("%s body=%s", msg, e.Body)
	}
	return msg
}

type idResponse struct {
	ID string `json:"id"`
}

type paginatedResponse[T any] struct {
	NextPage *int `json:"next_page"`
	Data     []T  `json:"data"`
}

type managedTypeResponse struct {
	ID string `json:"id"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := parseSeedConfig()
	if err != nil {
		slog.Error("invalid seed-api configuration", "error", err)
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ExecutionTimeout)
	defer cancel()

	client := &apiClient{
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	if err := seedAPI(ctx, client, cfg); err != nil {
		slog.Error("seed-api failed", "error", err)
		os.Exit(1)
	}

	slog.Info("seed-api completed successfully", "base_url", cfg.BaseURL)
}

func parseSeedConfig() (seedConfig, error) {
	defaultBaseURL := strings.TrimSpace(os.Getenv("SEED_API_URL"))
	defaultEmail := strings.TrimSpace(os.Getenv("SEED_API_EMAIL"))
	if defaultEmail == "" {
		defaultEmail = defaultAdminEmail
	}

	defaultPassword := strings.TrimSpace(os.Getenv("SEED_API_PASSWORD"))
	if defaultPassword == "" {
		defaultPassword = defaultAdminPassword
	}

	cfg := seedConfig{}

	flag.StringVar(&cfg.BaseURL, "base-url", defaultBaseURL, "Target API base URL (ex: http://localhost:8080)")
	flag.StringVar(&cfg.Email, "email", defaultEmail, "Login email used for seeding")
	flag.StringVar(&cfg.Password, "password", defaultPassword, "Login password used for seeding")
	flag.IntVar(&cfg.ClassCount, "class-count", defaultClassCount, "Number of classrooms to create")
	flag.IntVar(&cfg.StudentsPerClass, "students-per-class", defaultStudentsPerClass, "Number of students per classroom")
	flag.Float64Var(&cfg.BonusChance, "bonus-chance", defaultBonusChance, "Probability [0..1] to create bonuses for a student")
	flag.Float64Var(&cfg.PenaltyChance, "penalty-chance", defaultPenaltyChance, "Probability [0..1] to create penalties for a student")
	flag.Float64Var(&cfg.PunishmentChance, "punishment-chance", defaultPunishmentChance, "Probability [0..1] to create punishments for a student")
	flag.IntVar(&cfg.MaxBonusesPerStudent, "max-bonuses", defaultMaxBonuses, "Max number of bonuses created when bonus chance hits")
	flag.IntVar(&cfg.MaxPenaltiesPerStudent, "max-penalties", defaultMaxPenalties, "Max number of penalties created when penalty chance hits")
	flag.IntVar(&cfg.MaxPunishments, "max-punishments", defaultMaxPunishments, "Max number of punishments created when punishment chance hits")
	flag.DurationVar(&cfg.Timeout, "http-timeout", defaultRequestTimeout, "HTTP request timeout (e.g. 20s)")
	flag.DurationVar(&cfg.ExecutionTimeout, "exec-timeout", defaultExecutionTimeout, "Global seed timeout (e.g. 5m)")
	flag.BoolVar(&cfg.CreateRandomUser, "create-random-user", true, "Create an extra random user when register is allowed")
	flag.Parse()

	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.Email = strings.TrimSpace(cfg.Email)
	cfg.Password = strings.TrimSpace(cfg.Password)

	if cfg.BaseURL == "" {
		return seedConfig{}, fmt.Errorf("base-url is required")
	}
	if !strings.HasPrefix(cfg.BaseURL, "http://") && !strings.HasPrefix(cfg.BaseURL, "https://") {
		return seedConfig{}, fmt.Errorf("base-url must start with http:// or https://")
	}
	if cfg.Email == "" || cfg.Password == "" {
		return seedConfig{}, fmt.Errorf("email and password are required")
	}
	if cfg.ClassCount < 0 || cfg.StudentsPerClass < 0 {
		return seedConfig{}, fmt.Errorf("class-count and students-per-class must be >= 0")
	}
	if cfg.MaxBonusesPerStudent < 0 || cfg.MaxPenaltiesPerStudent < 0 || cfg.MaxPunishments < 0 {
		return seedConfig{}, fmt.Errorf("max values must be >= 0")
	}

	for key, value := range map[string]float64{
		"bonus-chance":      cfg.BonusChance,
		"penalty-chance":    cfg.PenaltyChance,
		"punishment-chance": cfg.PunishmentChance,
	} {
		if value < 0 || value > 1 {
			return seedConfig{}, fmt.Errorf("%s must be in [0,1]", key)
		}
	}

	if cfg.Timeout <= 0 || cfg.ExecutionTimeout <= 0 {
		return seedConfig{}, fmt.Errorf("timeouts must be > 0")
	}

	return cfg, nil
}

func seedAPI(ctx context.Context, client *apiClient, cfg seedConfig) error {
	registerAllowed, err := client.getRegisterStatus(ctx)
	if err != nil {
		return err
	}

	if registerAllowed {
		if err := client.registerUser(ctx, cfg.Email, "Admin", "User", cfg.Password); err != nil {
			var apiErr *apiRequestError
			if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
				slog.Info("admin user already exists", "email", cfg.Email)
			} else if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized && apiErr.APIError == "register_not_allowed" {
				slog.Warn("register was disabled while seeding, continuing with login only")
			} else {
				return fmt.Errorf("failed to register admin user: %w", err)
			}
		} else {
			slog.Info("admin user registered", "email", cfg.Email)
		}
	} else {
		slog.Info("register disabled on target API, skip admin creation")
	}

	if err := client.login(ctx, cfg.Email, cfg.Password); err != nil {
		return fmt.Errorf("failed to login with provided credentials: %w", err)
	}

	slog.Info("authenticated", "email", cfg.Email)

	if registerAllowed && cfg.CreateRandomUser {
		if err := client.registerRandomUser(ctx); err != nil {
			slog.Warn("failed to register random user", "error", err)
		}
	}

	bonusTypeIDs, err := client.ensureManagedTypes(ctx, "/v1/bonus-types", []string{
		"Participation", "Devoir rendu", "Aide aux camarades",
	})
	if err != nil {
		return fmt.Errorf("failed to ensure bonus types: %w", err)
	}

	penaltyTypeIDs, err := client.ensureManagedTypes(ctx, "/v1/penalty-types", []string{
		"Retard", "Bavardage", "Oubli materiel",
	})
	if err != nil {
		return fmt.Errorf("failed to ensure penalty types: %w", err)
	}

	punishmentTypeIDs, err := client.ensureManagedTypes(ctx, "/v1/punishment-types", []string{
		"Retenue", "Mot aux parents", "Exclusion",
	})
	if err != nil {
		return fmt.Errorf("failed to ensure punishment types: %w", err)
	}

	for classIndex := 1; classIndex <= cfg.ClassCount; classIndex++ {
		classroomID, err := client.createClassroom(ctx, classIndex)
		if err != nil {
			return err
		}

		for studentIndex := 0; studentIndex < cfg.StudentsPerClass; studentIndex++ {
			studentID, err := client.createStudent(ctx)
			if err != nil {
				return err
			}

			if err := client.addStudentToClassroom(ctx, classroomID, studentID); err != nil {
				return err
			}

			if roll(cfg.BonusChance) {
				if err := client.createRandomBonusesForStudent(ctx, studentID, bonusTypeIDs, cfg.MaxBonusesPerStudent); err != nil {
					return err
				}
			}

			if roll(cfg.PenaltyChance) {
				if err := client.createRandomPenaltiesForStudent(ctx, studentID, penaltyTypeIDs, cfg.MaxPenaltiesPerStudent); err != nil {
					return err
				}
			}

			if roll(cfg.PunishmentChance) {
				if err := client.createRandomPunishmentsForStudent(ctx, studentID, punishmentTypeIDs, cfg.MaxPunishments); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *apiClient) getRegisterStatus(ctx context.Context) (bool, error) {
	var resp registerStatusResponse
	if err := c.doJSON(ctx, http.MethodGet, "/v1/auth/register/status", nil, &resp, http.StatusOK); err != nil {
		return false, fmt.Errorf("failed to get register status: %w", err)
	}

	return resp.RegisterAllowed, nil
}

func (c *apiClient) registerUser(ctx context.Context, email, firstName, lastName, password string) error {
	body := map[string]string{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"password":   password,
	}

	return c.doJSON(ctx, http.MethodPost, "/v1/auth/register", body, nil, http.StatusCreated)
}

func (c *apiClient) registerRandomUser(ctx context.Context) error {
	for attempt := 1; attempt <= randomUserCreateAttempts; attempt++ {
		email := faker.Email()
		firstName := faker.FirstName()
		lastName := faker.LastName()
		password := email

		err := c.registerUser(ctx, email, firstName, lastName, password)
		if err == nil {
			slog.Info("random user registered", "email", email)
			return nil
		}

		var apiErr *apiRequestError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusConflict {
			continue
		}

		return err
	}

	return fmt.Errorf("could not create random user after %d attempts", randomUserCreateAttempts)
}

func (c *apiClient) login(ctx context.Context, email, password string) error {
	body := map[string]string{
		"email":    email,
		"password": password,
	}

	var resp loginResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/auth/login", body, &resp, http.StatusOK); err != nil {
		return err
	}
	if strings.TrimSpace(resp.AccessToken) == "" {
		return fmt.Errorf("login succeeded but access_token is empty")
	}

	c.accessToken = resp.AccessToken
	return nil
}

func (c *apiClient) ensureManagedTypes(ctx context.Context, listPath string, defaultNames []string) ([]string, error) {
	existing, err := c.listManagedTypeIDs(ctx, listPath)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}

	created := make([]string, 0, len(defaultNames))
	for _, name := range defaultNames {
		var resp idResponse
		if err := c.doJSON(ctx, http.MethodPost, listPath, map[string]string{"name": name}, &resp, http.StatusCreated); err != nil {
			return nil, fmt.Errorf("failed to create managed type %q on %s: %w", name, listPath, err)
		}
		created = append(created, resp.ID)
	}

	return created, nil
}

func (c *apiClient) listManagedTypeIDs(ctx context.Context, listPath string) ([]string, error) {
	var allIDs []string
	page := 1

	for {
		path := fmt.Sprintf("%s?page=%d", listPath, page)
		var response paginatedResponse[managedTypeResponse]
		if err := c.doJSON(ctx, http.MethodGet, path, nil, &response, http.StatusOK); err != nil {
			return nil, fmt.Errorf("failed to list managed types on %s: %w", listPath, err)
		}

		for _, item := range response.Data {
			if strings.TrimSpace(item.ID) == "" {
				continue
			}
			allIDs = append(allIDs, item.ID)
		}

		if response.NextPage == nil {
			break
		}
		page = *response.NextPage
	}

	return allIDs, nil
}

func (c *apiClient) createClassroom(ctx context.Context, classIndex int) (string, error) {
	mainTeacher := fmt.Sprintf("%s %s", faker.FirstName(), faker.LastName())
	body := map[string]string{
		"name":         fmt.Sprintf("Classe %d", classIndex),
		"year":         defaultYearLabel,
		"main_teacher": mainTeacher,
	}

	var classroom idResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/classrooms", body, &classroom, http.StatusCreated); err != nil {
		return "", fmt.Errorf("failed to create classroom %d: %w", classIndex, err)
	}

	slog.Info("classroom created", "classroom_id", classroom.ID)
	return classroom.ID, nil
}

func (c *apiClient) createStudent(ctx context.Context) (string, error) {
	body := map[string]string{
		"first_name": faker.FirstName(),
		"last_name":  faker.LastName(),
	}

	var student idResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/students", body, &student, http.StatusCreated); err != nil {
		return "", fmt.Errorf("failed to create student: %w", err)
	}

	return student.ID, nil
}

func (c *apiClient) addStudentToClassroom(ctx context.Context, classroomID, studentID string) error {
	path := fmt.Sprintf("/v1/classrooms/%s/students", classroomID)
	body := map[string]string{
		"student_id": studentID,
	}

	if err := c.doJSON(ctx, http.MethodPost, path, body, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to link student %s to classroom %s: %w", studentID, classroomID, err)
	}

	return nil
}

func (c *apiClient) createRandomBonusesForStudent(ctx context.Context, studentID string, bonusTypeIDs []string, maxCount int) error {
	if len(bonusTypeIDs) == 0 || maxCount <= 0 {
		return nil
	}

	count := rand.Intn(maxCount) + 1
	for range count {
		body := map[string]any{
			"student_id":    studentID,
			"bonus_type_id": bonusTypeIDs[rand.Intn(len(bonusTypeIDs))],
			"points":        randomBonusPoints(),
		}
		if err := c.doJSON(ctx, http.MethodPost, "/v1/bonuses", body, nil, http.StatusCreated); err != nil {
			return fmt.Errorf("failed to create bonus for student %s: %w", studentID, err)
		}
	}

	return nil
}

func (c *apiClient) createRandomPenaltiesForStudent(ctx context.Context, studentID string, penaltyTypeIDs []string, maxCount int) error {
	if len(penaltyTypeIDs) == 0 || maxCount <= 0 {
		return nil
	}

	count := rand.Intn(maxCount) + 1
	for range count {
		body := map[string]any{
			"student_id":      studentID,
			"penalty_type_id": penaltyTypeIDs[rand.Intn(len(penaltyTypeIDs))],
		}
		if err := c.doJSON(ctx, http.MethodPost, "/v1/penalties", body, nil, http.StatusCreated); err != nil {
			return fmt.Errorf("failed to create penalty for student %s: %w", studentID, err)
		}
	}

	return nil
}

func (c *apiClient) createRandomPunishmentsForStudent(ctx context.Context, studentID string, punishmentTypeIDs []string, maxCount int) error {
	if len(punishmentTypeIDs) == 0 || maxCount <= 0 {
		return nil
	}

	count := rand.Intn(maxCount) + 1
	for range count {
		body := map[string]any{
			"student_id":         studentID,
			"punishment_type_id": punishmentTypeIDs[rand.Intn(len(punishmentTypeIDs))],
			"due_at":             randomFutureDueAt().Format(time.RFC3339),
		}
		if err := c.doJSON(ctx, http.MethodPost, "/v1/punishments", body, nil, http.StatusCreated); err != nil {
			return fmt.Errorf("failed to create punishment for student %s: %w", studentID, err)
		}
	}

	return nil
}

func (c *apiClient) doJSON(
	ctx context.Context,
	method string,
	path string,
	reqBody any,
	out any,
	expectedStatus ...int,
) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var bodyReader io.Reader
	if reqBody != nil {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body for %s %s: %w", method, path, err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to build request %s %s: %w", method, path, err)
	}

	req.Header.Set("Accept", "application/json")
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s failed: %w", method, path, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body for %s %s: %w", method, path, err)
	}

	if containsStatus(expectedStatus, resp.StatusCode) {
		if out == nil || len(responseBody) == 0 {
			return nil
		}
		if err := json.Unmarshal(responseBody, out); err != nil {
			return fmt.Errorf("failed to decode response body for %s %s: %w", method, path, err)
		}
		return nil
	}

	apiErr := apiRequestError{
		StatusCode: resp.StatusCode,
		Path:       path,
		Body:       strings.TrimSpace(string(responseBody)),
	}

	var errorResponse apiErrorResponse
	if err := json.Unmarshal(responseBody, &errorResponse); err == nil {
		apiErr.APIError = errorResponse.Error
	}

	return &apiErr
}

func containsStatus(expected []int, status int) bool {
	for _, value := range expected {
		if value == status {
			return true
		}
	}
	return false
}

func randomBonusPoints() float64 {
	steps := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0}
	return steps[rand.Intn(len(steps))]
}

func randomFutureDueAt() time.Time {
	days := rand.Intn(14) + 1
	hour := rand.Intn(9) + 8
	targetDate := time.Now().AddDate(0, 0, days)
	return time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), hour, 0, 0, 0, targetDate.Location())
}

func roll(chance float64) bool {
	return rand.Float64() < chance
}
