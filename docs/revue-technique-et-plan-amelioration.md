# Revue complète du projet (implémentation, sécurité, cohérence routes/DTO)

## Portée de la revue

Cette revue couvre :
- l'architecture et l'implémentation globale ;
- la sécurité applicative (auth, cookies, multi-tenant) ;
- la cohérence des routes exposées ;
- la cohérence des DTOs de retour ;
- un plan d'amélioration priorisé.

---

## 1) État général de l'implémentation

### Points solides
- Architecture claire en couches : `handler -> service -> repository(sqlc)`.
- Tests présents et nombreux sur handlers/services.
- Validation DTO systématique via tags `validate`.
- Pagination homogène via helpers web dédiés.
- Isolation multi-tenant systématique par `user_id` dans les services/repository.

### Points à surveiller
- La logique métier de création de pénalité + déclenchement de punitions automatiques est riche : il faut maintenir des tests d'intégration transactionnels pour éviter les régressions subtiles.
- Le projet expose des routes `force delete` pour les types : c'est utile en maintenance, mais nécessite une gouvernance stricte côté produit.

---

## 2) Revue sécurité

### Bonnes pratiques déjà en place
- Middleware d'auth Bearer sur toutes les routes métier.
- Vérification JWT (signature + expiration), et codes d'erreur dédiés.
- Cookie refresh token `HttpOnly` et `SameSite=Strict`.
- Retour `404` pour ressources hors tenant (bonne discrétion de données).

### Risques identifiés
1. **Cookie refresh en clair en prod potentiellement**
   - Le cookie est posé avec `Secure: false` actuellement.
   - Risque : interception sur HTTP non chiffré.
   - Recommandation : rendre ce flag configurable par environnement et forcer `Secure=true` hors local.

2. **Absence de rotation/invalidation active des refresh tokens à l'usage**
   - Le refresh est vérifié en base, mais la rotation n'est pas visible dans ce flux.
   - Recommandation : rotation à chaque refresh + révocation de l'ancien token.

3. **Durcissement observabilité sécurité**
   - Ajouter métriques et alertes sur erreurs auth (pics de `401`, erreurs JWT).

---

## 3) Cohérence des routes

### Constat
- Les routes du routeur Chi sont globalement cohérentes et bien regroupées par ressource.
- Les conventions REST sont majoritairement respectées (`POST` create, `GET` list/get, `PUT` update, `DELETE` delete).
- Cas spécifiques métier cohérents :
  - `POST /bonuses/{id}/use`
  - `POST /punishments/{id}/resolve`
  - routes relationnelles élève/classe.

### Ajustements documentation
- Les endpoints `DELETE /{type}/{id}/force` existaient dans le code mais n'étaient pas explicitement listés partout dans la doc API.
- La route exemptée d'auth doit être notée explicitement en `/v1/health`.

---

## 4) Cohérence des DTOs de retour

### Constat principal
- Les DTOs de retour sont globalement homogènes.
- Incohérence documentaire identifiée : le champ `automated` est présent côté code dans les réponses `Punishment` et dans l'historique élève, mais absent de la doc API.

### Ajustements documentation effectués
- Ajout de `automated` dans les exemples `Punishment`.
- Ajout de `automated` dans `StudentHistoryItem` (type punishment).
- Ajout de `automated` dans la définition `ReturnPunishmentDto`.

---

## 5) Plan d'améliorations priorisé

## Priorité P0 (sécurité)
1. Rendre `Secure` du cookie refresh configurable et forcé à `true` en non-local.
2. Ajouter rotation des refresh tokens (token unique actif par session).
3. Ajouter tests de non-régression sécurité :
   - cookie flags selon env ;
   - rejet refresh token révoqué.

## Priorité P1 (robustesse)
4. Ajouter tests e2e centrés sur la cohérence métier : `CreatePenalty -> Rules -> Punishments`.
5. Standardiser un contrat explicite pour les routes `force` (quand les utiliser, impacts).
6. Ajouter audit log métier pour actions sensibles (`force delete`, `resolve`, `use`).

## Priorité P2 (maintenabilité)
7. Générer une table de routes automatiquement depuis le routeur pour éviter la dérive doc/code.
8. Ajouter une checklist PR "DTO de retour impacté => docs API mises à jour".
9. Renforcer la documentation des états métier (`pending/resolved`, `used/unused`).

---

## 6) Résumé exécutable

- Implémentation : solide et bien structurée.
- Sécurité : base saine, mais `refresh cookie Secure` et rotation des tokens sont des améliorations importantes.
- Routes : cohérentes, avec quelques écarts documentaires corrigés.
- DTOs retour : cohérents ; documentation maintenant alignée sur le champ `automated`.
