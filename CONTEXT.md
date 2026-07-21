# org-archiver

A CLI that moves completed items out of org-mode files into a companion archive
file, keeping the working file focused on what's still open.

## Language

**Done item**: A headline whose TODO status is `DONE` or `CANCELLED` — the unit
that gets archived. The set is fixed to these two keywords. _Avoid_: completed
task, finished item, closed item.

**Subtree**: A headline together with all its descendant headlines and content.
Archiving always moves a whole subtree, never a lone headline.

**Source file**: The `.org` file passed on the command line, edited in place to
remove done items. _Avoid_: input file, working file.

**Archive file**: The companion `<source>.org_archive` that receives done items.
Same directory as its source. _Avoid_: output file, backup.

**Run**: A single invocation of the tool. Each run that archives anything adds
one archive section.

**Archive section**: The `* archive <timestamp>` heading that groups all items
archived in one run. _Avoid_: archive entry, archive block.

**Re-leveling**: Shifting the heading levels of an archived subtree so its root
becomes level 2, sitting directly under its archive section.

## Architecture Decisions

**Done set is fixed, and enforced over a file's own `#+TODO:`.** `DONE` and
`CANCELLED` are always treated as done, regardless of what a file's `#+TODO:`
line declares. go-org only assigns a `Status` to keywords the file registers, so
`forceDoneStatus` reclassifies leading title text to catch our keywords when the
file's config omits them. This diverges from Emacs org, where an unregistered
keyword is plain text — but it makes archiving predictable without per-file
configuration. Alternatives rejected: deriving the set from `#+TODO:`
(unpredictable across files) and a CLI flag (unnecessary surface).

**Editing is a full go-org round-trip, not line-surgery.** Modified files are
parsed and re-serialized with go-org's `OrgWriter`, which pretty-prints the
_entire_ file — so archiving a single item can produce a large, noisy diff in
untouched sections. Accepted for implementation simplicity over a line-based
edit that would preserve byte-for-byte formatting. A file with nothing to
archive is left untouched, so this reformatting only ever hits files that
actually changed.

**Archive format: dated grouping, level-2 normalization, no provenance.** Each
run appends one `* archive <timestamp>` section; every archived subtree is
re-leveled so its root is `**` beneath it, preserving internal structure and
document order. No per-item provenance is recorded (no `:ARCHIVE_TIME:`,
`:ARCHIVE_OLPATH:`, etc.) — the dated section is the only breadcrumb. This is
deliberately unlike Emacs `org-archive-subtree`, which writes flat items each
carrying an `ARCHIVE_*` property drawer.
