# Fantasy Premier League (FPL) & English Premier League (EPL) Integration

This integration adds FPL and EPL functionality to your Discord bot.

## FPL Commands

### .fplstandings (or .fpl)
Shows the standings for your guild's Fantasy Premier League league.

Usage:
```
.fplstandings
```

### .setfplleague
Sets the Fantasy Premier League league ID for your guild (Admin only).

Usage:
```
.setfplleague <league_id>
```

## EPL Commands

### .epltable
Shows the current Premier League table.

Usage:
```
.epltable
```

### .nextmatch
Shows upcoming Premier League fixtures. Use with a club name to see their next match.

Usage:
```
.nextmatch
.nextmatch <club_name>
```

## Setup

1. To use FPL commands, an admin must first set the league ID:
   ```
   .setfplleague 123456
   ```

2. Once set, any user can view the standings:
   ```
   .fplstandings
   ```

3. EPL commands work without any setup:
   ```
   .epltable
   .nextmatch
   .nextmatch arsenal
   ```