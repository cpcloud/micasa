- remove the v1 in Logs; remove the forward slashes; the ghost test should read
  type a Perl-compatible regex; put the log lines themselves in visually
  separate components
- i don't see a spinner spinning or any activity when i begin searching; also the search indexing seems to never complete and so i never get search results even when i konw i should
- highlight the part of the string that the regex matched in log lines
- i can't edit existing entries, make that work please
- add the ability to sort by a column with a key stroke, it should toggle
  ascending, descending, and no sort. sort by primary key by default
- make a global search interface that works like google. it should pop up a box
  or something and show any of the data that matches and allow selecting it and
  then take you to that row. a hard requirement is that it must run locally (so
  probably building a search index in the background, if you  go that route
  make sure to show a spinner while the index is building
- with no logging (ie no -vv) the keystroke info is really tight up against the bottom of the data
- for maintenance items, compute the default ghost text for next due date from the last serviced date + the maintenance interval
- can you make the keystroke info always appear all the way at the bottom of the terminal?
