- make a new tab, call it Appliances, for tracking appliance information; add
  an appliance field to Maintenance that links (by foreign key) to this new
  dataset; second, make it so that when i hit enter on cell that is linked
  (maybe also find a way to indicate that a column is linked to another table
  [including its relationship 1:1, 1:m, m:1 or m:n as the indicator would also
  │ be fuckin' dope]) that will move you to that cell in the other tab (if it's
  1:1 of course. if 1:m go to the first match, if m:1, there's only a single
  match, if m:n not sure, help me   figure out the ux for that )
- [RW-HOUSE-UX] redesign house profile collapsed/expanded views: remove chip borders,
  use middot-separated inline text, cleaner section layout in expanded view
- add the ability to sort by a column with a key stroke, it should toggle
  ascending, descending, and no sort. sort by primary key by default
- for maintenance items, compute the default ghost text for next due date from
  the last serviced date + the maintenance interval and default to that

## Completed

- refactor forms.go and view.go: deduplicate submit/edit pairs, centering, inline edit boilerplate, form-data converters (9851c74)
- scrap the log-on-dash-v approach, just enable logging dynamically (and allow changing log level) with a keyboard shortcut and bring up the logger ui component when that key is pressed (it's a toggle obviously) (75b2c86)

- remove the v1 in Logs; remove the forward slashes; the ghost text should read type a Perl-compatible regex; put the log lines themselves in visually separate components (1c623d4)
- build a search engine (and sweet embedded UI for it) that MUST run locally, with spinner and selection (1c623d4)
- make a global search interface that works like google, pop up a box, show matches, select and jump to row, runs locally with background indexing and spinner (1c623d4)
- highlight the part of the string that the regex matched in log lines (4289fb7)
- i can't edit existing entries, make that work please (a457c44)
- with no logging the keystroke info is really tight up against the bottom of the data (a457c44)
- can you make the keystroke info always appear all the way at the bottom of the terminal? (a457c44)
