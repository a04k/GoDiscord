# Permission System Overhaul

This document outlines the changes made to the bot's permission system for better flexibility and management across multiple guilds.

## Summary of Changes

The old permission system, which used boolean flags (`is_owner`, `is_admin`, `is_mod`) in the `users` table, has been replaced with a role-based access control system. This new system is managed through a dedicated `permissions` table in the database.

## New Permission Hierarchy

1.  **Owner**: Has full permissions. Can appoint Admins.
2.  **Admin**: Can manage the bot's settings, including enabling/disabling commands. Can appoint Mods.
3.  **Moderator**: Has basic moderation permissions.

## Database Changes

-   A new table, `permissions`, has been created with the following schema:
    -   `guild_id` (TEXT): The ID of the guild.
    -   `user_id` (TEXT): The ID of the user.
    -   `role` (TEXT): The user's role in the guild (`owner`, `admin`, or `moderator`).
-   The `is_owner`, `is_admin`, and `is_mod` columns have been removed from the `users` table.
-   The `dbedit.go` script has been updated to reflect these changes.

## Code Refactoring

### `bot/bot.go`

-   The in-code database schema has been removed. The database schema is now solely managed by `dbedit.go`.
-   New permission-checking functions have been added:
    -   `GetUserRole(guildID, userID)`
    -   `IsOwner(guildID, userID)`
    -   `IsAdmin(guildID, userID)`
    -   `IsMod(guildID, userID)`
-   A `guildCreate` event handler has been added to automatically assign the `owner` role to the server owner when the bot joins a new guild.

### Admin Commands

-   The `setupowner.go` and `setowner.go` command files have been deleted.
-   The following commands have been updated to use the new role-based permission system:
    -   `disable.go`: Requires Admin role.
    -   `enable.go`: Requires Admin role.
    -   `setadmin.go`: Requires Owner role.
    -   `setmod.go`: Requires Admin role.

### `utils/helpers.go`

-   The old permission-checking functions (`IsOwner`, `IsAdmin`, `IsModerator`) have been removed to avoid conflicts with the new system.

### `main.go`

-   Removed obsolete logic for setting the owner on `messageCreate`, `guildCreate`, and `guildUpdate` events, as this is now handled by the `guildCreate` handler in `bot/bot.go`.
