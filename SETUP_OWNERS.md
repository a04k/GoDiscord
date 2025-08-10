# Guild Owner Setup Tool

This tool helps you manually set guild owners for existing guilds where the bot is already a member.

## Usage

1. Make sure your `.env` file is properly configured with the `DATABASE_URL`
2. Run the tool: `go run setupowners.go`
3. Enter the guild ID and user ID when prompted
4. The tool will set the specified user as the owner of that guild
5. Type 'quit' when you're done

## Example

```
Guild Owner Setup Tool
======================
This tool helps you manually set guild owners for existing guilds.

Enter guild ID (or 'quit' to exit): 584955644155789329
Enter user ID to set as owner: 174176306865504257
Successfully set <@174176306865504257> as owner in guild 584955644155789329

Enter guild ID (or 'quit' to exit): quit
Done!
```

## How It Works

The tool directly updates the database to mark the specified user as the owner of the specified guild. This is useful for cases where:

1. The bot was already in the guild before the automatic owner detection was implemented
2. You need to manually override the guild owner for any reason
3. The guild owner changed and the automatic detection didn't catch it

## Security Note

Only the bot owner (you) should run this tool. Make sure to keep your `DATABASE_URL` secure and never share it with others.