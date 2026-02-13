# Exemples et Cas d'Usage

Ce document fournit des exemples concrets d'utilisation de l'API The Punisher.

## Table des matières

- [Scénarios complets](#scénarios-complets)
- [Exemples de requêtes cURL](#exemples-de-requêtes-curl)
- [Cas d'usage typiques](#cas-dusage-typiques)
- [Gestion des erreurs](#gestion-des-erreurs)

---

## Scénarios complets

### Scénario 1 : Configuration initiale d'un professeur

**Étape 1 : Créer un compte**
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "prof@lycee.fr",
    "first_name": "Jean",
    "last_name": "Dupont",
    "password": "monmotdepasse123"
  }'
```

**Réponse :**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "prof@lycee.fr",
  "first_name": "Jean",
  "last_name": "Dupont",
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

**Étape 2 : Se connecter**
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "prof@lycee.fr",
    "password": "monmotdepasse123"
  }' \
  -c cookies.txt  # Sauvegarder les cookies (refresh_token)
```

**Réponse :**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Étape 3 : Créer des classes**
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Créer la classe 3ème A
curl -X POST http://localhost:8080/v1/classrooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "3ème A",
    "year": "2023-2024",
    "main_teacher": "M. Dupont"
  }'

# Créer la classe 4ème B
curl -X POST http://localhost:8080/v1/classrooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "4ème B",
    "year": "2023-2024"
  }'
```

**Étape 4 : Créer des étudiants**
```bash
# Créer Alice Martin
curl -X POST http://localhost:8080/v1/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Alice",
    "last_name": "Martin"
  }'

# Créer Bob Durand
curl -X POST http://localhost:8080/v1/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Bob",
    "last_name": "Durand"
  }'
```

**Étape 5 : Associer les étudiants aux classes**
```bash
CLASSROOM_ID="class-uuid-3a"
ALICE_ID="alice-uuid"
BOB_ID="bob-uuid"

# Ajouter Alice à la 3ème A
curl -X POST http://localhost:8080/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "'$ALICE_ID'"
  }'

# Ajouter Bob à la 3ème A
curl -X POST http://localhost:8080/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "'$BOB_ID'"
  }'
```

**Étape 6 : Créer des types de bonus**
```bash
# Créer type "Participation active"
curl -X POST http://localhost:8080/v1/bonus-types \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Participation active"
  }'

# Créer type "Devoir rendu en avance"
curl -X POST http://localhost:8080/v1/bonus-types \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Devoir rendu en avance"
  }'
```

---

### Scénario 2 : Consultation des données

**Lister tous les étudiants (avec pagination)**
```bash
curl -X GET "http://localhost:8080/v1/students?page=1" \
  -H "Authorization: Bearer $TOKEN"
```

**Réponse :**
```json
{
  "page": 1,
  "item_per_page": 20,
  "total_count": 28,
  "previous_page": null,
  "next_page": 2,
  "data": [
    {
      "id": "alice-uuid",
      "first_name": "Alice",
      "last_name": "Martin",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "bob-uuid",
      "first_name": "Bob",
      "last_name": "Durand",
      "created_at": "2024-01-15T10:31:00Z",
      "updated_at": "2024-01-15T10:31:00Z"
    }
  ]
}
```

**Voir les détails d'un étudiant**
```bash
curl -X GET http://localhost:8080/v1/students/$ALICE_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Lister les classes d'un étudiant**
```bash
curl -X GET http://localhost:8080/v1/students/$ALICE_ID/classrooms \
  -H "Authorization: Bearer $TOKEN"
```

**Lister les étudiants d'une classe**
```bash
curl -X GET http://localhost:8080/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN"
```

---

### Scénario 3 : Mise à jour de données

**Mettre à jour un étudiant**
```bash
curl -X PUT http://localhost:8080/v1/students/$ALICE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "last_name": "Martin-Dupont"
  }'
```

**Mettre à jour une classe**
```bash
curl -X PUT http://localhost:8080/v1/classrooms/$CLASSROOM_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "3ème A - Section Européenne"
  }'
```

---

### Scénario 4 : Suppression

**Retirer un étudiant d'une classe**
```bash
curl -X DELETE http://localhost:8080/v1/classrooms/$CLASSROOM_ID/students/$ALICE_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Supprimer un étudiant (définitif)**
```bash
curl -X DELETE http://localhost:8080/v1/students/$ALICE_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Supprimer une classe (définitif)**
```bash
curl -X DELETE http://localhost:8080/v1/classrooms/$CLASSROOM_ID \
  -H "Authorization: Bearer $TOKEN"
```

---

### Scénario 5 : Renouvellement du token

**Quand l'access token expire, utiliser le refresh token**
```bash
curl -X POST http://localhost:8080/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt  # Mettre à jour les cookies
```

**Réponse :**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## Exemples de requêtes cURL

### Variables d'environnement
```bash
export API_URL="http://localhost:8080"
export TOKEN="votre-access-token"
```

### Authentification

**Inscription**
```bash
curl -X POST $API_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nouveau@prof.fr",
    "first_name": "Nouveau",
    "last_name": "Prof",
    "password": "password123"
  }'
```

**Connexion**
```bash
curl -X POST $API_URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nouveau@prof.fr",
    "password": "password123"
  }' \
  -c cookies.txt
```

**Refresh**
```bash
curl -X POST $API_URL/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

### Étudiants

**Créer**
```bash
curl -X POST $API_URL/v1/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Charlie",
    "last_name": "Brown"
  }'
```

**Lister (page 1)**
```bash
curl -X GET "$API_URL/v1/students?page=1" \
  -H "Authorization: Bearer $TOKEN"
```

**Récupérer**
```bash
STUDENT_ID="uuid"
curl -X GET $API_URL/v1/students/$STUDENT_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Mettre à jour**
```bash
curl -X PUT $API_URL/v1/students/$STUDENT_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Charles"
  }'
```

**Supprimer**
```bash
curl -X DELETE $API_URL/v1/students/$STUDENT_ID \
  -H "Authorization: Bearer $TOKEN"
```

### Classes

**Créer**
```bash
curl -X POST $API_URL/v1/classrooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "5ème C",
    "year": "2024-2025",
    "main_teacher": "Mme Durant"
  }'
```

**Lister**
```bash
curl -X GET "$API_URL/v1/classrooms?page=1" \
  -H "Authorization: Bearer $TOKEN"
```

**Ajouter un étudiant**
```bash
CLASSROOM_ID="uuid"
STUDENT_ID="uuid"
curl -X POST $API_URL/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "'$STUDENT_ID'"
  }'
```

**Retirer un étudiant**
```bash
curl -X DELETE $API_URL/v1/classrooms/$CLASSROOM_ID/students/$STUDENT_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Lister les étudiants d'une classe**
```bash
curl -X GET $API_URL/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN"
```

**Lister les classes d'un étudiant**
```bash
curl -X GET $API_URL/v1/students/$STUDENT_ID/classrooms \
  -H "Authorization: Bearer $TOKEN"
```

### Types de Bonus

**Créer**
```bash
curl -X POST $API_URL/v1/bonus-types \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Aide aux camarades"
  }'
```

**Lister**
```bash
curl -X GET "$API_URL/v1/bonus-types?page=1" \
  -H "Authorization: Bearer $TOKEN"
```

**Mettre à jour**
```bash
BONUS_TYPE_ID="uuid"
curl -X PUT $API_URL/v1/bonus-types/$BONUS_TYPE_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Aide exceptionnelle aux camarades"
  }'
```

---

## Cas d'usage typiques

### Cas 1 : Gestion d'une classe en début d'année

1. Créer la classe
2. Créer tous les étudiants
3. Ajouter tous les étudiants à la classe
4. Créer les types de bonus personnalisés

```bash
#!/bin/bash
TOKEN="your-token"
API_URL="http://localhost:8080"

# 1. Créer la classe
CLASSROOM_RESPONSE=$(curl -s -X POST $API_URL/v1/classrooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "2nde 1",
    "year": "2024-2025"
  }')

CLASSROOM_ID=$(echo $CLASSROOM_RESPONSE | jq -r '.id')

# 2. Créer les étudiants
declare -a students=("Alice Martin" "Bob Durand" "Charlie Petit")

for student in "${students[@]}"; do
  first_name=$(echo $student | cut -d' ' -f1)
  last_name=$(echo $student | cut -d' ' -f2)
  
  STUDENT_RESPONSE=$(curl -s -X POST $API_URL/v1/students \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "first_name": "'$first_name'",
      "last_name": "'$last_name'"
    }')
  
  STUDENT_ID=$(echo $STUDENT_RESPONSE | jq -r '.id')
  
  # 3. Ajouter l'étudiant à la classe
  curl -s -X POST $API_URL/v1/classrooms/$CLASSROOM_ID/students \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "student_id": "'$STUDENT_ID'"
    }'
done

echo "Classe $CLASSROOM_ID créée avec tous les étudiants"
```

### Cas 2 : Recherche et consultation

**Trouver tous les types de bonus**
```bash
curl -X GET $API_URL/v1/bonus-types \
  -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, name}'
```

**Voir toutes mes classes**
```bash
curl -X GET $API_URL/v1/classrooms \
  -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, name, year}'
```

**Compter mes étudiants**
```bash
curl -X GET $API_URL/v1/students \
  -H "Authorization: Bearer $TOKEN" | jq '.total_count'
```

### Cas 3 : Mise à jour en masse

**Renommer plusieurs types de bonus**
```bash
#!/bin/bash
TOKEN="your-token"
API_URL="http://localhost:8080"

# Récupérer tous les types de bonus
BONUS_TYPES=$(curl -s -X GET $API_URL/v1/bonus-types \
  -H "Authorization: Bearer $TOKEN")

# Extraire les IDs
IDS=$(echo $BONUS_TYPES | jq -r '.data[].id')

# Ajouter un préfixe à chaque nom
for id in $IDS; do
  curl -X PUT $API_URL/v1/bonus-types/$id \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "🌟 Participation active"
    }'
done
```

---

## Gestion des erreurs

### Erreur : Token expiré
```bash
curl -X GET $API_URL/v1/students \
  -H "Authorization: Bearer expired-token"
```

**Réponse :**
```json
{
  "error": "jwt_expired",
  "error_code": 401
}
```

**Solution :** Utiliser l'endpoint refresh
```bash
curl -X POST $API_URL/v1/auth/refresh -b cookies.txt
```

---

### Erreur : Validation échouée
```bash
curl -X POST $API_URL/v1/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "A",
    "last_name": ""
  }'
```

**Réponse :**
```json
{
  "error": "validation_failed",
  "error_code": 400,
  "error_details": [
    {
      "field": "first_name",
      "error": "validation_min_length:2"
    },
    {
      "field": "last_name",
      "error": "validation_field_required"
    }
  ]
}
```

---

### Erreur : Ressource non trouvée
```bash
curl -X GET $API_URL/v1/students/non-existent-uuid \
  -H "Authorization: Bearer $TOKEN"
```

**Réponse :**
```json
{
  "error": "student_not_found",
  "error_code": 404
}
```

---

### Erreur : Conflit
```bash
# Ajouter deux fois le même étudiant à une classe
curl -X POST $API_URL/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "'$STUDENT_ID'"
  }'

# Deuxième fois
curl -X POST $API_URL/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "'$STUDENT_ID'"
  }'
```

**Réponse (deuxième appel) :**
```json
{
  "error": "student_classroom_relation_exists",
  "error_code": 409
}
```

---

### Erreur : Email déjà utilisé
```bash
curl -X POST $API_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "existing@email.com",
    "first_name": "Test",
    "last_name": "User",
    "password": "password123"
  }'
```

**Réponse :**
```json
{
  "error": "conflict",
  "error_code": 409,
  "error_details": [
    {
      "field": "email",
      "error": "validation_email_already_exists"
    }
  ]
}
```

---

## Scripts utiles

### Script de test complet
```bash
#!/bin/bash
set -e

API_URL="http://localhost:8080"

echo "=== Test de l'API The Punisher ==="

# 1. Créer un compte
echo "1. Création du compte..."
REGISTER_RESPONSE=$(curl -s -X POST $API_URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@'$(date +%s)'@example.com",
    "first_name": "Test",
    "last_name": "User",
    "password": "testpassword123"
  }')

echo "Compte créé: $(echo $REGISTER_RESPONSE | jq -r '.email')"

EMAIL=$(echo $REGISTER_RESPONSE | jq -r '.email')

# 2. Se connecter
echo "2. Connexion..."
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'$EMAIL'",
    "password": "testpassword123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
echo "Token obtenu"

# 3. Créer un étudiant
echo "3. Création d'un étudiant..."
STUDENT_RESPONSE=$(curl -s -X POST $API_URL/v1/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "Student"
  }')

STUDENT_ID=$(echo $STUDENT_RESPONSE | jq -r '.id')
echo "Étudiant créé: $STUDENT_ID"

# 4. Créer une classe
echo "4. Création d'une classe..."
CLASSROOM_RESPONSE=$(curl -s -X POST $API_URL/v1/classrooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Classe Test",
    "year": "2024-2025"
  }')

CLASSROOM_ID=$(echo $CLASSROOM_RESPONSE | jq -r '.id')
echo "Classe créée: $CLASSROOM_ID"

# 5. Ajouter l'étudiant à la classe
echo "5. Ajout de l'étudiant à la classe..."
curl -s -X POST $API_URL/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "'$STUDENT_ID'"
  }'

echo "Étudiant ajouté à la classe"

# 6. Vérifier
echo "6. Vérification..."
STUDENTS_IN_CLASS=$(curl -s -X GET $API_URL/v1/classrooms/$CLASSROOM_ID/students \
  -H "Authorization: Bearer $TOKEN")

echo "Nombre d'étudiants dans la classe: $(echo $STUDENTS_IN_CLASS | jq '.total_count')"

echo "=== Test terminé avec succès ==="
```

---

## Notes importantes

### Isolation des données
- Chaque professeur ne voit que ses propres données
- Le `user_id` est automatiquement récupéré du token JWT
- Impossible d'accéder aux données d'un autre professeur

### Pagination
- Par défaut : 20 éléments par page
- Toujours utiliser le paramètre `page` pour naviguer
- Le champ `total_count` indique le nombre total d'éléments

### Suppressions
- Toutes les suppressions sont **définitives** (hard delete)
- Les relations en cascade sont gérées automatiquement
- Attention : les données supprimées ne peuvent pas être récupérées

### Sécurité
- L'access token expire rapidement (ex: 15 minutes)
- Utiliser le refresh token pour obtenir un nouveau token
- Le refresh token est stocké dans un cookie HTTP-Only
- Ne jamais exposer les tokens dans les logs ou le code côté client
