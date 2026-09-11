# Shared-property discovery task

You are analyzing research notes about attempts on the Erdős–Straus
equation 4/n = 1/x + 1/y + 1/z (n >= 2; x, y, z positive integers). The
attached notes are project-authored synthetic study material, not verified
mathematical literature; treat their annotations as scoped claims.

Use only the supplied notes. Do not browse, call tools, open files beyond
the one you were told to read, or rely on outside knowledge of this
specific corpus. Do not guess what a hidden benchmark rewards.

Task: propose up to FIVE candidate SHARED PROPERTIES — structural
properties that two or more of the FAILED or PARTIALLY FAILED approaches
have in common and that plausibly participate in why they stall. A shared
property is not a shared topic or shared wording: two approaches can use
the same words and differ structurally, or describe the same property in
different words. Prefer properties that discriminate: a property that the
successful or partially successful approaches ALSO have is weaker evidence
of failure structure, and you must check for that.

For each proposed property supply exactly: statement (one sentence);
shared_by (>= 2 note filenames); supporting_passages (verbatim quotes with
their note filenames); preserved_distinctions (what the property
deliberately does NOT merge — name approaches that remain structurally
different despite sharing it); contrast_check (what you found when checking
the property against the partial-success notes); falsification (what
observation would refute the property). Do not merge distinct operations
into one property merely because they co-occur; if the evidence does not
establish a shared property, propose fewer — an empty list is acceptable
and is not penalized.

Return ONE JSON document, no Markdown fence, exactly:
{"schema_version": "discovery-wire/v1", "properties": [...]}
with the six fields above per property and nothing else. Quotes must be
verbatim substrings of the attached notes. Order properties by your
confidence.

The notes follow.
