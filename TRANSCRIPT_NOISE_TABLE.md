# Transcript Noise Table — SCALER DSA

> **Purpose:** Reference table of all noise patterns found in SCALER DSA transcripts.
> **Source:** Manual analysis of 5 transcript files (~87,000+ lines).
> **Used by:** `transcript_cleaner.py` — automated cleaning pipeline.
> **Rule:** Before adding a new noise pattern, verify it appears in actual transcript files.

---

## How to Use This Table

When a new transcript is uploaded, `transcript_cleaner.py` applies these rules in order:
1. **REMOVE** — entire line deleted
2. **CLEAN** — line kept but specific pattern removed/replaced
3. **KEEP** — line always preserved regardless of other rules

KEEP rules override REMOVE rules.

---

## Category 1 — Speaker Attribution Lines

Lines that only contain speaker name + timestamp. Zero DSA content.

| Pattern | Example | Action |
|---|---|---|
| `Name (HH:MM:SS)` | `Sneha Mehra (00:04:33)` | REMOVE entire line |
| `Name (HH:MM:SS)  ` | `Sneha Mehra (00:04:33)  ` | REMOVE entire line |
| Regex | `^[A-Za-z\s]+\(\d{2}:\d{2}:\d{2}\)\s*$` | REMOVE entire line |

---

## Category 2 — Filler Greetings

Class start/end social filler. Zero DSA content.

| Noise Pattern | Variants | Action |
|---|---|---|
| `Hello everyone` | `Hello everyone, good evening`, `Hello everyone, ⁓ good evening` | REMOVE |
| `Good evening` | `Good evening to all`, `good evening everyone` | REMOVE |
| `Hello` | Standalone `Hello.`, `Hello, hello, hello` | REMOVE |
| `Thank you` | `Thank you, Brahma`, `Thank you everyone` | REMOVE |
| `All right, I'm back` | — | REMOVE |
| `Just give me one minute` | `give me one minute okay` | REMOVE |
| `I'll just set the iPad screen` | — | REMOVE |
| `Am I audible to all` | `Is my voice clear` | REMOVE |
| `Yes, no, yes` | Student response noise | REMOVE |
| `It is slow` | Audio quality check | REMOVE |

---

## Category 3 — Hesitation Markers

Speech disfluencies captured by transcript tool.

| Pattern | Example | Action |
|---|---|---|
| `⁓` symbol standalone | Line containing only `⁓` | REMOVE entire line |
| `⁓` mid-sentence | `We have ⁓ a question` | CLEAN — remove `⁓`, keep sentence |
| `⁓` at line start | `⁓ Hello everyone` | CLEAN — remove `⁓ `, keep rest |

---

## Category 4 — Administrative / Logistics Talk

Class schedule, breaks, team announcements. Zero DSA content.

| Noise Pattern | Example | Action |
|---|---|---|
| Schedule talk | `ending on 14th September`, `next week or next to next week` | REMOVE |
| Break announcements | `we'll try to have a break`, `team's call` | REMOVE |
| Future class mentions | `we will discuss in the next class`, `advanced content will start` | REMOVE |
| Curriculum changes | `old curriculum`, `classes were created some extra` | REMOVE |
| Batch/module info | `last week of the essay`, `next module that is final` | REMOVE |
| `I am not sure when` | `I am not sure when the advanced content will start` | REMOVE |
| `Don't worry about` | `Don't worry about anything`, `Don't worry about the future` | REMOVE |

---

## Category 5 — Student Interaction Noise

Chat responses, quiz mechanics. Not DSA teaching content.

| Noise Pattern | Example | Action |
|---|---|---|
| Quiz instructions | `please answer the quiz`, `take your time then click` | REMOVE |
| Quiz results | `88% people got it right`, `92% people got it right` | REMOVE |
| Student name mentions | `Kujbhush says`, `Tejas says`, `Seemant will come` | REMOVE |
| Thumbs up requests | `please give a thumbs up if this is clear` | REMOVE |
| Answer privately rule | `answer it privately`, `use question tab` | REMOVE |
| Top 10 winners | `Top 10 winners` | REMOVE |
| Volunteer selection | `Let's take Shiva as a volunteer`, `Everyone will get a chance` | REMOVE |
| `% people got it right` | Any line with this pattern | REMOVE |
| `click yes` / `click no` | `just click yes`, `click on the answer` | REMOVE |
| `3, 2, 1, go` | Quiz countdown | REMOVE |

---

## Category 6 — Incomplete / Broken Sentences

Transcript artifacts — mid-sentence cuts, single words, OCR errors.

| Pattern | Example | Action |
|---|---|---|
| Single word lines | `you`, `and`, `Graph` (standalone) | REMOVE |
| Lines with only `⁓` | `⁓` | REMOVE |
| `+\|` separator lines | `+\|` or `\|` only | REMOVE |
| Trailing `\\` on numbers | `5\\.`, `n minus 1\\.` | CLEAN — remove `\\` |
| Lines < 20 chars, no tech terms | `Okay.`, `Yes.`, `Sure.`, `Nice.` | REMOVE |
| `in Japanese. ⁓` | Random language reference | REMOVE |

---

## Category 7 — Language / Programming Disclaimers

Repeated boilerplate. Not DSA content.

| Noise Pattern | Example | Action |
|---|---|---|
| Language disclaimer | `I cannot go specific in any programming language` | REMOVE |
| Pseudo code disclaimer | `I'm writing pseudo code`, `pseudo words` | REMOVE |
| Language-specific note | `In C++, it's a different way. In Java, it's a different way` | REMOVE |

---

## Category 8 — Motivational / Off-topic Filler

| Noise Pattern | Example | Action |
|---|---|---|
| Movie references | `Shah Rukh Khan's movie`, `Mehmuna` | REMOVE |
| Motivational talk | `don't stop solving questions`, `big journey is about to be over` | REMOVE |
| `I am there` type | `I am here, Then why are you worrying?` | REMOVE |
| `Enjoy the present` | `Let's enjoy the present` | REMOVE |
| `Future will take care` | — | REMOVE |

---

## Category 9 — Section Headers (Standalone)

Single topic words as standalone lines — no context, no sentence.

| Pattern | Example | Action |
|---|---|---|
| Standalone topic word | `Graph`, `Arrays`, `Hashing` as a lone line | REMOVE |

---

## KEEP Rules — Always Preserve

These override ALL remove rules. If a line contains any of these, keep it.

### Technical Terms (always keep)
```
O(n), O(log), O(1), O(n^2)
array, arrays, index, indexing
hash, hashing, hash table, hash map
tree, binary tree, BST, binary search tree
graph, DFS, BFS, depth first, breadth first
sort, sorting, merge sort, quick sort, bubble sort
recursion, recursive, base case, stack overflow
dynamic programming, DP, memoization, tabulation
backtracking
linked list, node, pointer, next, prev
stack, queue, deque, priority queue
heap, min heap, max heap
binary search
time complexity, space complexity
algorithm, data structure
for loop, while loop, function, method
O(, n log n, log n
Union Find, DSU, disjoint set
two pointer, sliding window
subarray, subsequence
prefix sum, suffix
matrix, 2D array
string, character, ASCII
bit manipulation, XOR, AND, OR
greedy
topological sort
Dijkstra, Bellman
MST, Kruskal, Prim
```

### Code Patterns (always keep)
```
int , void , return , public , private
for(, while(, if(, else
[], {}, ()
= , ==, !=, <=, >=
new , null , true , false
```

---

## Cleaning Statistics (Expected)

Based on analysis of SCALER DSA transcripts:

| Metric | Before Cleaning | After Cleaning (Expected) |
|---|---|---|
| Total lines | ~87,000 | ~35,000-45,000 |
| Noise lines removed | — | ~45-55% |
| Avg chunk quality | Low (noise mixed) | High (DSA content only) |
| Reranker score | < 0.35 (fails) | > 0.35 (passes) |

---

## Adding New Noise Patterns

When you find a new noise pattern:
1. Verify it appears in actual transcript (not assumed)
2. Add to correct category above
3. Add corresponding regex/string to `transcript_cleaner.py`
4. Test on a sample before re-ingesting

---

## Change Log

| Date | Change |
|---|---|
| 2026-09-13 | Initial noise table created from analysis of 5 SCALER DSA transcript files |
