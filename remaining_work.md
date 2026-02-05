- i can't edit existing entries, make that work please
- scrap the log-on-dash-v approach, just enable logging dynamically with
  a keyboard shortcut and bring up the logger ui component when that key is
  pressed (it's a toggle obviously)
- add the ability to sort by a column with a key stroke, it should toggle
  ascending, descending, and no sort. sort by primary key by default
- with no logging (ie no -vv) the keystroke info is really tight up against the bottom of the data
- for maintenance items, compute the default ghost text for next due date from the last serviced date + the maintenance interval
- can you make the keystroke info always appear all the way at the bottom of the terminal?

## Completed

- remove the v1 in Logs; remove the forward slashes; the ghost text should read type a Perl-compatible regex; put the log lines themselves in visually separate components (1c623d4)
- build a search engine (and sweet embedded UI for it) that MUST run locally, with spinner and selection (1c623d4)
- make a global search interface that works like google, pop up a box, show matches, select and jump to row, runs locally with background indexing and spinner (1c623d4)
- highlight the part of the string that the regex matched in log lines (4289fb7)
