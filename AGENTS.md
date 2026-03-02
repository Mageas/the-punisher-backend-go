# AGENTS.md

## Objectif
Définir des règles de travail obligatoires pour ce repository.

## Portée
Ces règles s'appliquent uniquement aux tâches de développement ("dev").

## Règles obligatoires

### 1) Issue GitHub en fin de dev si demandé par l'utilisateur (et seulement en fin de dev)
- Quand une tâche est une tâche de dev et demandé par l'utilisateur (fait moi une issue Github), créer une issue GitHub à la fin du dev.
- Utiliser le template le plus adapté dans `.github/ISSUE_TEMPLATE/`:
  - `bug.md`
  - `feature.md`
  - `refactor.md`
  - `test.md`
  - `chore.md`
  - `docs.md` (si la tâche est majoritairement documentation)
- Ne pas créer d'issue GitHub pour les tâches non-dev.
- L'issue doit être basée sur le template choisi et complétée proprement (contexte, besoin, critères d'acceptation).

### 2) Mise à jour de la documentation pendant les devs
- Toute modification de code qui impacte le comportement, l'architecture, les contrats API, la configuration ou les workflows doit inclure une mise à jour de la documentation associée.
- Mettre à jour `README.md` et/ou les fichiers concernés dans `docs/` dans la même intervention.
- Si aucune mise à jour de documentation n'est nécessaire, l'indiquer explicitement dans le compte-rendu final.

### 3) Message de commit en anglais
- Comme pour la règle d'issue GitHub, cette règle s'applique uniquement aux tâches de dev.
- Le message de commit doit être écrit en anglais.
- Utiliser un format de type Conventional Commits:
  - `feat: ...`
  - `fix: ...`
  - `refactor: ...`
  - `test: ...`
  - `chore: ...`
  - `docs: ...`

## Définition pratique de "dev"
Une tâche est considérée comme "dev" si elle inclut au moins un des points suivants:
- modification du code source
- ajout/modification de tests
- changement de schéma/migration/base de données
- changement de configuration runtime/CI impactant le produit
- changement de contrat API ou de comportement fonctionnel

Une tâche n'est pas "dev" si elle est uniquement:
- réponse à une question
- lecture/analyse sans modification
- revue sans changement de code
- tâche administrative sans impact code
