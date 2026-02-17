# Plan d'enrichissement API — Réduction des requêtes frontend

## Contexte

L'API actuelle retourne des réponses "plates" contenant uniquement des UUIDs pour les relations (ex: `student_id`, `bonus_type_id`).
En analysant les mockups, chaque page frontend nécessiterait un nombre excessif de requêtes pour résoudre les noms et calculer les agrégats.

Ce document liste **page par page** les données manquantes et les changements à apporter.

---

## Règle de cohérence — Un DTO = une forme unique

**Principe** : toute route qui retourne une entité donnée retourne **toujours** le DTO enrichi, qu'il s'agisse d'un GET unitaire, d'une liste, d'une création (POST), d'une mise à jour (PUT) ou d'une action (use/resolve).

Concrètement :

| Entité | Routes concernées (toutes retournent le même DTO enrichi) |
|--------|-----------------------------------------------------------|
| **Student** | `GET /students`, `GET /students/{id}`, `POST /students`, `PUT /students/{id}`, `GET /classrooms/{id}/students` |
| **Classroom** | `GET /classrooms`, `GET /classrooms/{id}`, `POST /classrooms`, `PUT /classrooms/{id}` |
| **Bonus** | `GET /bonuses`, `GET /bonuses/{id}`, `POST /bonuses`, `POST /bonuses/{id}/use`, `GET /students/{id}/bonuses` |
| **Penalty** | `GET /penalties`, `GET /penalties/{id}`, `POST /penalties`, `GET /students/{id}/penalties` |
| **Punishment** | `GET /punishments`, `GET /punishments/{id}`, `POST /punishments`, `POST /punishments/{id}/resolve`, `GET /students/{id}/punishments` |
| **Rule** | `GET /rules`, `GET /rules/{id}`, `POST /rules`, `PUT /rules/{id}` |

Ainsi, le frontend peut toujours s'attendre à la même forme de réponse, peu importe la route utilisée.

### Décision : `points` sur `bonus_types`

Le mockup `types-bonus.html` affiche un champ "points" par type de bonus. Cependant, **les points restent uniquement sur `bonuses`** (pas sur `bonus_types`). Le mockup affiche en réalité les points définis au moment de la création du bonus. Le champ `points` dans le `RequestBonusDto` est à fournir par le frontend lors de la création. `ReturnBonusTypeDto` ne change pas.

---

## Sommaire des problèmes

| Page | Requêtes actuelles estimées | Après enrichissement |
|------|----------------------------|---------------------|
| Dashboard | ~50+ (KPIs + 3 listes avec résolution de noms) | 1 |
| Liste élèves | 1 + 3×N (classes, bonus, pénalités par élève) | 1 |
| Profil élève | ~10+ (KPIs + punitions + bonus + historique + résolution noms) | 1 |
| Liste classes | 1 + N (étudiants + agrégats par classe) | 1 |
| Détail classe | 1 + N (agrégats par élève) | 1 |
| Liste bonus | 1 + résolution noms (étudiant + type par bonus) | 1 |
| Liste punitions | 1 + résolution noms (étudiant + type + règle par punition) | 1 |
| Liste règles | 1 + résolution noms (2 types par règle) | 1 |

---

## 1. Dashboard — Nouvelle route

### Route : `GET /v1/dashboard`

**Problème** : Cette page n'existe pas dans l'API. Le frontend devrait faire :
- `GET /students` → compter
- `GET /bonuses?state=unused` → sommer les points, compter
- `GET /penalties` → compter total
- `GET /punishments?state=pending` → compter
- Puis pour chaque item des 3 listes : résoudre `student_id` → nom, `*_type_id` → nom type

**Paramètre optionnel** : `?classroom_id=uuid` (filtre par classe)

**Réponse proposée** :

```json
{
  "kpis": {
    "student_count": 34,
    "available_bonus_points": 14.5,
    "unused_bonus_count": 12,
    "penalty_count": 47,
    "pending_punishment_count": 5
  },
  "recent_penalties": [
    {
      "id": "uuid",
      "student_id": "uuid",
      "student_first_name": "Thomas",
      "student_last_name": "Martin",
      "penalty_type_id": "uuid",
      "penalty_type_name": "Bavardage",
      "created_at": "2026-02-17T14:32:00Z"
    }
  ],
  "recent_bonuses": [
    {
      "id": "uuid",
      "student_id": "uuid",
      "student_first_name": "Emma",
      "student_last_name": "Bernard",
      "bonus_type_id": "uuid",
      "bonus_type_name": "Participation",
      "points": 1.0,
      "created_at": "2026-02-17T10:15:00Z",
      "used_at": null
    }
  ],
  "pending_punishments": [
    {
      "id": "uuid",
      "student_id": "uuid",
      "student_first_name": "Lucas",
      "student_last_name": "Dubois",
      "punishment_type_id": "uuid",
      "punishment_type_name": "Retenue",
      "triggering_rule_id": "uuid",
      "due_at": "2026-02-20T17:00:00Z",
      "created_at": "2026-02-17T08:00:00Z"
    }
  ]
}
```

**Notes d'implémentation** :
- Les listes `recent_penalties` et `recent_bonuses` sont limitées aux ~10 derniers éléments (pas de pagination, c'est un aperçu).
- `pending_punishments` retourne toutes les punitions non résolues (ou les 10 premières si trop nombreuses).
- Si `classroom_id` est fourni, tous les KPIs et listes sont filtrés aux élèves de cette classe.
- Le champ `triggering_rule_id` non null indique que la punition est automatique (badge "Auto" dans le mockup).

---

## 2. Liste élèves — Enrichir `GET /v1/students`

### Route : `GET /v1/students?page=1`

**Problème** : Le mockup affiche par élève :
- Nom complet ✅ (déjà présent)
- Badges des classes (noms) ❌
- Points bonus disponibles (somme) ❌
- Nombre de pénalités ❌

Actuellement, `ReturnStudentDto` ne contient que `id`, `first_name`, `last_name`, `created_at`, `updated_at`.

**Réponse proposée** — enrichir `ReturnStudentDto` dans les listes :

```json
{
  "id": "uuid",
  "first_name": "Thomas",
  "last_name": "Martin",
  "classrooms": [
    { "id": "uuid", "name": "6ème A" },
    { "id": "uuid", "name": "3ème Latin" }
  ],
  "available_bonus_points": 3.0,
  "penalty_count": 5,
  "created_at": "...",
  "updated_at": "..."
}
```

**Champs ajoutés** :
| Champ | Type | Description |
|-------|------|-------------|
| `classrooms` | `[]{id, name}` | Classes associées (id + nom seulement) |
| `available_bonus_points` | `float64` | Somme des points des bonus non consommés (`used_at IS NULL`) |
| `penalty_count` | `int` | Nombre total de pénalités |

**Impact SQL** : Une seule requête avec `LEFT JOIN` + agrégats, ou bien une requête dédiée (`ListStudentsEnriched`) qui fait les sous-requêtes corrélées.

---

## 3. Profil élève — Nouvelle route

### Route : `GET /v1/students/{id}/profile`

**Problème** : Le mockup du profil affiche beaucoup d'informations que le `GET /students/{id}` actuel ne fournit pas :
- Classes (noms)
- KPIs : points bonus dispo (+ nombre), pénalités totales, punitions en attente
- Punitions en attente avec nom du type + badge auto
- Bonus disponibles avec nom du type
- Historique unifié et chronologique

Le frontend devrait actuellement faire :
1. `GET /students/{id}` — infos de base
2. `GET /students/{id}/classrooms` — classes
3. `GET /students/{id}/bonuses?state=unused` — bonus dispo
4. `GET /students/{id}/bonuses` — tous les bonus (pour historique)
5. `GET /students/{id}/penalties` — toutes les pénalités
6. `GET /students/{id}/punishments?state=pending` — punitions en attente
7. `GET /students/{id}/punishments` — toutes les punitions (pour historique)
8. Pour chaque bonus → résoudre `bonus_type_id`
9. Pour chaque pénalité → résoudre `penalty_type_id`
10. Pour chaque punition → résoudre `punishment_type_id`

**Réponse proposée** :

```json
{
  "student": {
    "id": "uuid",
    "first_name": "Lucas",
    "last_name": "Dubois",
    "created_at": "...",
    "updated_at": "..."
  },
  "classrooms": [
    { "id": "uuid", "name": "6ème A" }
  ],
  "kpis": {
    "available_bonus_points": 3.0,
    "active_bonus_count": 2,
    "total_penalty_count": 5,
    "pending_punishment_count": 1
  },
  "pending_punishments": [
    {
      "id": "uuid",
      "punishment_type_id": "uuid",
      "punishment_type_name": "Retenue",
      "triggering_rule_id": "uuid",
      "due_at": "2026-02-20T17:00:00Z",
      "created_at": "2026-02-17T08:00:00Z"
    }
  ],
  "available_bonuses": [
    {
      "id": "uuid",
      "bonus_type_id": "uuid",
      "bonus_type_name": "Participation",
      "points": 1.0,
      "created_at": "2026-02-15T10:00:00Z"
    }
  ],
  "history": [
    {
      "type": "punishment",
      "id": "uuid",
      "punishment_type_name": "Retenue",
      "triggering_rule_id": "uuid",
      "resolved_at": null,
      "due_at": "2026-02-20T17:00:00Z",
      "created_at": "2026-02-17T08:00:00Z"
    },
    {
      "type": "penalty",
      "id": "uuid",
      "penalty_type_name": "Bavardage",
      "created_at": "2026-02-17T08:00:00Z"
    },
    {
      "type": "bonus",
      "id": "uuid",
      "bonus_type_name": "Participation",
      "points": 1.0,
      "used_at": null,
      "created_at": "2026-02-15T10:00:00Z"
    }
  ]
}
```

**Notes** :
- `history` est un tableau unifié trié par `created_at` décroissant, qui mélange penalties, bonuses et punishments.
- Chaque entrée a un champ `type` discriminant (`"penalty"`, `"bonus"`, `"punishment"`).
- `pending_punishments` et `available_bonuses` sont des sous-ensembles de `history` préfiltrés pour un affichage direct (sections dédiées du mockup).
- Pas de pagination pour les sections `pending_punishments` et `available_bonuses` (en général peu nombreux).
- `history` peut être paginée si nécessaire (paramètre `?history_page=1`).

---

## 4. Liste classes — Enrichir `GET /v1/classrooms`

### Route : `GET /v1/classrooms?page=1`

**Problème** : Le mockup des cards affiche par classe :
- Nom ✅
- Année scolaire ✅
- Nombre d'élèves ❌
- Avatars/initiales des élèves ❌
- Points bonus total ❌
- Pénalités total ❌

**Réponse proposée** — enrichir `ReturnClassroomDto` :

```json
{
  "id": "uuid",
  "name": "6ème A",
  "year": "2025-2026",
  "main_teacher": null,
  "student_count": 12,
  "students_preview": [
    { "id": "uuid", "first_name": "Thomas", "last_name": "Martin" },
    { "id": "uuid", "first_name": "Lucas", "last_name": "Dubois" },
    { "id": "uuid", "first_name": "Emma", "last_name": "Bernard" }
  ],
  "total_bonus_points": 14.5,
  "total_penalty_count": 23,
  "created_at": "...",
  "updated_at": "..."
}
```

**Champs ajoutés** :
| Champ | Type | Description |
|-------|------|-------------|
| `student_count` | `int` | Nombre total d'élèves dans la classe |
| `students_preview` | `[]{id, first_name, last_name}` | Premiers ~5 élèves (pour les avatars/initiales) |
| `total_bonus_points` | `float64` | Somme des points bonus disponibles (`used_at IS NULL`) de tous les élèves |
| `total_penalty_count` | `int` | Nombre total de pénalités de tous les élèves |

---

## 5. Détail classe — Enrichir `GET /v1/classrooms/{id}/students`

### Route : `GET /v1/classrooms/{id}/students?page=1`

**Problème** : Le mockup affiche par élève dans la classe :
- Nom complet ✅
- Points bonus disponibles ❌
- Nombre de pénalités ❌

Actuellement, cette route retourne des `ReturnStudentDto` standard (sans agrégats).

**Réponse proposée** — retourner le même `ReturnStudentDto` enrichi que `GET /v1/students` :

```json
[
  {
    "id": "uuid",
    "first_name": "Thomas",
    "last_name": "Martin",
    "classrooms": [
      { "id": "uuid", "name": "6ème A" },
      { "id": "uuid", "name": "3ème Latin" }
    ],
    "available_bonus_points": 3.0,
    "penalty_count": 5,
    "created_at": "...",
    "updated_at": "..."
  }
]
```

**Champs ajoutés** (identiques à la section 2) :
| Champ | Type | Description |
|-------|------|-------------|
| `classrooms` | `[]{id, name}` | Toutes les classes de l'élève (pas seulement la classe courante) |
| `available_bonus_points` | `float64` | Somme des points bonus disponibles de l'élève |
| `penalty_count` | `int` | Nombre total de pénalités de l'élève |

**Pourquoi inclure `classrooms[]` ici aussi ?** Un élève peut être dans plusieurs classes. Afficher toutes ses classes même dans le contexte d'une classe spécifique maintient la cohérence du DTO et évite un aller-retour supplémentaire.

De plus, les KPIs de la page détail classe (nombre d'élèves, total bonus, total pénalités) peuvent être récupérés via le `GET /v1/classrooms/{id}` enrichi (voir section 4 — appliquer le même enrichissement au détail).

---

## 6. Liste bonus — Enrichir `GET /v1/bonuses`

### Route : `GET /v1/bonuses?page=1`

**Problème** : Le mockup affiche par bonus :
- Nom de l'élève ❌ (actuellement seulement `student_id`)
- Nom du type de bonus ❌ (actuellement seulement `bonus_type_id`)
- Points ✅
- Date ✅
- Statut (déduit de `used_at`) ✅

**Réponse proposée** — enrichir `ReturnBonusDto` :

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Emma",
  "student_last_name": "Bernard",
  "bonus_type_id": "uuid",
  "bonus_type_name": "Participation",
  "points": 1.0,
  "created_at": "2026-02-14T10:00:00Z",
  "used_at": null
}
```

**Champs ajoutés** :
| Champ | Type | Description |
|-------|------|-------------|
| `student_first_name` | `string` | Prénom de l'élève |
| `student_last_name` | `string` | Nom de l'élève |
| `bonus_type_name` | `string` | Nom du type de bonus |

**S'applique aussi à** : `GET /v1/students/{id}/bonuses`

---

## 7. Liste pénalités — Enrichir `GET /v1/penalties`

### Route : `GET /v1/penalties?page=1`

**Problème** : Le mockup du dashboard affiche les dernières pénalités avec le nom de l'élève et le nom du type.

**Réponse proposée** — enrichir `ReturnPenaltyDto` :

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Thomas",
  "student_last_name": "Martin",
  "penalty_type_id": "uuid",
  "penalty_type_name": "Bavardage",
  "created_at": "2026-02-17T14:32:00Z"
}
```

**Champs ajoutés** :
| Champ | Type | Description |
|-------|------|-------------|
| `student_first_name` | `string` | Prénom de l'élève |
| `student_last_name` | `string` | Nom de l'élève |
| `penalty_type_name` | `string` | Nom du type de pénalité |

**S'applique aussi à** : `GET /v1/students/{id}/penalties`

---

## 8. Liste punitions — Enrichir `GET /v1/punishments`

### Route : `GET /v1/punishments?page=1`

**Problème** : Le mockup affiche par punition :
- Nom de l'élève ❌
- Nom du type de punition ❌
- Badge "Auto" (déduit de `triggering_rule_id`) ✅
- Description du déclenchement (ex: "3 bavardages cumulés") ❌
- Statut (déduit de `resolved_at`) ✅
- Date d'échéance ✅

**Réponse proposée** — enrichir `ReturnPunishmentDto` :

```json
{
  "id": "uuid",
  "student_id": "uuid",
  "student_first_name": "Lucas",
  "student_last_name": "Dubois",
  "punishment_type_id": "uuid",
  "punishment_type_name": "Retenue",
  "triggering_rule_id": "uuid",
  "triggering_rule_name": "3 bavardages => retenue",
  "created_at": "2026-02-17T08:00:00Z",
  "due_at": "2026-02-20T17:00:00Z",
  "resolved_at": null
}
```

**Champs ajoutés** :
| Champ | Type | Description |
|-------|------|-------------|
| `student_first_name` | `string` | Prénom de l'élève |
| `student_last_name` | `string` | Nom de l'élève |
| `punishment_type_name` | `string` | Nom du type de punition |
| `triggering_rule_name` | `string\|null` | Nom de la règle déclencheuse (null si manuelle) |

**S'applique aussi à** : `GET /v1/students/{id}/punishments`

---

## 9. Liste règles — Enrichir `GET /v1/rules`

### Route : `GET /v1/rules?page=1`

**Problème** : Le mockup affiche par règle :
- Nom du type de pénalité déclencheur ❌ (actuellement seulement `penalty_type_id`)
- Nom du type de punition résultant ❌ (actuellement seulement `resulting_punishment_type_id`)
- Mode ✅
- Seuil ✅
- Toggle actif/inactif ✅

**Réponse proposée** — enrichir `ReturnRuleDto` :

```json
{
  "id": "uuid",
  "name": "3 bavardages => retenue",
  "penalty_type_id": "uuid",
  "penalty_type_name": "Bavardage",
  "resulting_punishment_type_id": "uuid",
  "resulting_punishment_type_name": "Retenue",
  "threshold": 3,
  "due_at_after_days": 7,
  "mode": "every",
  "is_active": true,
  "created_at": "...",
  "updated_at": "..."
}
```

**Champs ajoutés** :
| Champ | Type | Description |
|-------|------|-------------|
| `penalty_type_name` | `string` | Nom du type de pénalité déclencheur |
| `resulting_punishment_type_name` | `string` | Nom du type de punition résultant |

---

## 10. Modales de création — Données nécessaires pour les selects

Les modales de création (bonus, pénalité, punition depuis le dashboard ou d'autres pages) ont besoin de listes pour peupler les `<select>` :
- Liste des classes → `GET /v1/classrooms` (déjà ok, non paginé ou 1ère page)
- Liste des élèves d'une classe → `GET /v1/classrooms/{id}/students` (déjà ok)
- Liste des types de bonus → `GET /v1/bonus-types` (déjà ok)
- Liste des types de pénalité → `GET /v1/penalty-types` (déjà ok)
- Liste des types de punition → `GET /v1/punishment-types` (déjà ok)

**Aucun changement nécessaire** pour ces types catalogues. Ils sont chargés une fois et mis en cache côté frontend.

---

## Résumé des changements

### Nouvelles routes (2)

| Route | Description |
|-------|-------------|
| `GET /v1/dashboard?classroom_id=` | Agrégat KPIs + listes récentes pour le dashboard |
| `GET /v1/students/{id}/profile` | Profil complet d'un élève avec KPIs + historique |

### DTOs à enrichir (6)

| DTO actuel | Champs à ajouter |
|------------|-----------------|
| `ReturnStudentDto` (listes) | `classrooms`, `available_bonus_points`, `penalty_count` |
| `ReturnClassroomDto` (listes + détail) | `student_count`, `students_preview`, `total_bonus_points`, `total_penalty_count` |
| `ReturnBonusDto` | `student_first_name`, `student_last_name`, `bonus_type_name` |
| `ReturnPenaltyDto` | `student_first_name`, `student_last_name`, `penalty_type_name` |
| `ReturnPunishmentDto` | `student_first_name`, `student_last_name`, `punishment_type_name`, `triggering_rule_name` |
| `ReturnRuleDto` | `penalty_type_name`, `resulting_punishment_type_name` |

### Toutes les routes impactées (GET, POST, PUT, actions)

| Route | DTO enrichi utilisé |
|-------|---------------------|
| `GET /v1/students` | `ReturnStudentDto` enrichi |
| `GET /v1/students/{id}` | idem |
| `POST /v1/students` | idem |
| `PUT /v1/students/{id}` | idem |
| `GET /v1/classrooms/{id}/students` | idem |
| `GET /v1/classrooms` | `ReturnClassroomDto` enrichi |
| `GET /v1/classrooms/{id}` | idem |
| `POST /v1/classrooms` | idem |
| `PUT /v1/classrooms/{id}` | idem |
| `GET /v1/bonuses` | `ReturnBonusDto` enrichi |
| `GET /v1/bonuses/{id}` | idem |
| `POST /v1/bonuses` | idem |
| `POST /v1/bonuses/{id}/use` | idem |
| `GET /v1/students/{id}/bonuses` | idem |
| `GET /v1/penalties` | `ReturnPenaltyDto` enrichi |
| `GET /v1/penalties/{id}` | idem |
| `POST /v1/penalties` | idem |
| `GET /v1/students/{id}/penalties` | idem |
| `GET /v1/punishments` | `ReturnPunishmentDto` enrichi |
| `GET /v1/punishments/{id}` | idem |
| `POST /v1/punishments` | idem |
| `POST /v1/punishments/{id}/resolve` | idem |
| `GET /v1/students/{id}/punishments` | idem |
| `GET /v1/rules` | `ReturnRuleDto` enrichi |
| `GET /v1/rules/{id}` | idem |
| `POST /v1/rules` | idem |
| `PUT /v1/rules/{id}` | idem |

---

## Priorité d'implémentation suggérée

1. **Enrichir `ReturnRuleDto`** (`GET /v1/rules`) — simple `JOIN` pour noms des types
2. **Enrichir `ReturnPenaltyDto`** (`GET /v1/penalties`, `GET /v1/students/{id}/penalties`) — simple `JOIN`
3. **Enrichir `ReturnBonusDto`** (`GET /v1/bonuses`, `GET /v1/students/{id}/bonuses`) — simple `JOIN`
4. **Enrichir `ReturnPunishmentDto`** (`GET /v1/punishments`, `GET /v1/students/{id}/punishments`) — `JOIN` + règle optionnelle
5. **Enrichir `ReturnStudentDto`** (`GET /v1/students`, `GET /v1/classrooms/{id}/students`) — agrégats + classes associées
6. **Enrichir `ReturnClassroomDto`** (`GET /v1/classrooms`, `GET /v1/classrooms/{id}`) — agrégats par classe + `students_preview`
7. **Créer `GET /v1/dashboard`** — nouvelle route agrégée avec filtre `classroom_id` optionnel
8. **Créer `GET /v1/students/{id}/profile`** — nouvelle route composite + historique unifié
9. **Modales de création** — aucun changement backend (catalogues déjà suffisants)

---

## Notes techniques

### Message important d'implémentation

Même les enrichissements dits "simples" (rules, penalties, bonuses, punishments) nécessitent de **modifier les requêtes SQL** dans `db/sqlc/queries.sql` (ajout de `JOIN` + alias des nouveaux champs), puis de régénérer sqlc.  
Il ne faut pas faire cet enrichissement uniquement dans les services (N+1) afin de respecter la règle de cohérence : **Un DTO = une forme unique**.

### Approche SQL recommandée

Pour les enrichissements simples (ajout de noms), utiliser des `JOIN` dans les requêtes sqlc :

```sql
-- Exemple : bonuses enrichis
SELECT b.*, s.first_name AS student_first_name, s.last_name AS student_last_name,
       bt.name AS bonus_type_name
FROM bonuses b
JOIN students s ON s.id = b.student_id
JOIN bonus_types bt ON bt.id = b.bonus_type_id
WHERE b.user_id = $1
ORDER BY b.created_at DESC
LIMIT 20 OFFSET $2;
```

Pour les agrégats (students list, classrooms list), utiliser des sous-requêtes corrélées ou des `LEFT JOIN` avec `GROUP BY` :

```sql
-- Exemple : students enrichis
SELECT s.*,
  COALESCE((SELECT SUM(b.points) FROM bonuses b WHERE b.student_id = s.id AND b.used_at IS NULL), 0) AS available_bonus_points,
  COALESCE((SELECT COUNT(*) FROM penalties p WHERE p.student_id = s.id), 0) AS penalty_count
FROM students s
WHERE s.user_id = $1
ORDER BY s.last_name, s.first_name
LIMIT 20 OFFSET $2;
```

Les classes associées à un élève nécessitent un traitement séparé (array aggregation PostgreSQL ou requête complémentaire) car un élève peut avoir plusieurs classes.

### Pas de breaking change

Les champs ajoutés sont **additifs** — les clients existants peuvent ignorer les nouveaux champs. Aucun champ existant n'est modifié ou supprimé.
