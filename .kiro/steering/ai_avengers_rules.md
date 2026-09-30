# AI Avengers Rules - STRICTLY FOLLOW

> Is file me jo bhi rule likha hai, AI ko har response, har code change, har tool call me 100% strictly follow karna hai. Koi exception nahi.

## Rules:
<!-- User yahan rules likhega, AI hamesha inhe follow karega -->

### RULE 1: LEARN BEFORE BUILD (MANDATORY)
1. Har task shuru karne se PEHLE `Learnings/` folder me se relevant learning ko padhna MANDATORY hai.
2. Sirf aur sirf `Learnings/` folder se seekhi hui knowledge ka use karke hi development karna hai.
3. Kabhi bhi generic / pre-trained knowledge ka use user ke EXPLICIT approval ke bina bilkul nahi karna hai.

### RULE 2: IF KNOWLEDGE NOT FOUND - ASK, DON'T ASSUME
4. Agar kisi task ke liye required knowledge `Learnings/` folder me available nahi hai, toh task ko rok dena hai.
5. User ko turant batana hai ki "Learnings folder me ye knowledge nahi mili".
6. User se exact path / file ka access mangna hai, usko padh kar seekh kar hi fir kaam shuru karna hai.
7. Bina seekhe / bina approval ke apne man se code/assumption se kaam nahi karna hai.

### RULE 3: LEARN AND UPDATE LEARNINGS (CONTINUOUS TRAINING)
8. Jab bhi koi nayi cheez seekho (user dwara di gayi file/path se ya approval ke baad generic knowledge se), usko turant `Learnings/` folder me update karna MANDATORY hai.
9. Update `Learnings/taught-knowledge.md` ya relevant file me karo taaki next time wahi se refer karke khud ko trained karke development start kiya ja sake.
10. Bina Learnings folder update kiye task ko complete mat mano.

### RULE 4: PROOF OF LEARNING (TRACEABILITY)
11. Har task complete karne ke baad short me prove dena MANDATORY hai ki kaam `Learnings/` ki kis file / kis learning ke basis pe kiya gaya hai.
12. Format: `Proof: Learnings/<file>.md -> <kya seekha> -> <kaise apply kiya>`
13. Agar generic knowledge approval se use kiya hai toh explicitly mention karo: `Proof: Generic (Approved) + Learnings/...`

### RULE 5: USER-DIRECTED LEARNING (SOURCE SUPREMACY)
14. User jahan se bhi bole (file path / folder / URL / transcript / docs / video / sources etc.) wahi se padhna aur seekhna MANDATORY hai.
15. Wahi se seekh kar turant `Learnings/` folder me (relevant file ya `Learnings/taught-knowledge.md` me) update/add karna MANDATORY hai.
16. Uske baad development me sirf usi seekhi hui learning ko follow karna hai, dusra source replace nahi karna hai.
17. User ke bataye hue source ko kabhi ignore ya skip nahi karna hai.

### RULE 6: MEMORY & EFFICIENCY - REMEMBER ONCE LEARNED
18. Ek baar kisi bhi file/source se seekh liya toh usko yaad rakhna MANDATORY hai, same cheez ko baar-baar padh kar time waste nahi karna hai.
19. Ek baar padh li toh us knowledge ko memory me retain karo.
20. Next time agar usi se milta-julta task aaye toh fir se Learnings file ko open karke padhna jaruri nahi hai.
21. Us situation me user se puchhna hai: "Kya Learnings file fir se padhni hai ya yaad ki hui knowledge se kaam karun?" aur user jo instruction de usi ko follow karna hai.
22. Make sure karna hai ki user ke instruction ka hi palan ho, apne man se re-read karke time waste nahi karna hai.

### RULE 7: READ -> UNDERSTAND -> PROPOSE -> APPROVE -> IMPLEMENT (MANDATORY WORKFLOW)
23. Code implement karne se PEHLE uss requirement ke liye necessary saari existing files ko read karna MANDATORY hai.
24. Architecture ko samjho aur confirm karo - kaunsi file kahan connect hai, kya pattern follow ho raha hai.
25. Implementation se pehle user ko approach/design batao - kaise design karoge aur kaise implement karoge (files, changes, steps).
26. User ke EXPLICIT approval ke baad hi implementation start karna hai, bina approval ke code mat likho.

### RULE 8: CODEBASE CODING STANDARDS - ALWAYS FOLLOW (MANDATORY ARCHITECTURE)
> Yeh saare standards AI Avengers ke actual codebase se extract kiye gaye hain. Har naya code isi architecture ko strictly follow karega, kabhi violate nahi karega.

**A. Structure & Layering:**
27. Backend Go module `ai_avengers/backend` - Handlers `internal/<domain>/`, route composition `cmd/server/main.go`, migrations paired `.up.sql/.down.sql`
28. Layered Architecture API -> App -> Business -> Foundation - Imports sirf neeche ki taraf, upar nahi. Business kabhi `net/http` import nahi karega.
29. Domain Vertical Lanes - Har domain (training, workflow, gateway, memory) ka apna app/business/storage. Cross-domain sirf business-to-business via small interface.
30. No `common/utils/helpers/models` - Har package ka purpose se naam, pehli file package ke naam par.

**B. Design Patterns / SOLID / OOP:**
31. Constructor Dependency Injection - `NewService(db, redis, logger)` pattern mandatory.
32. Small Interfaces at Boundaries - Testability ke liye, "Discover interfaces, don't design them" - Interface sirf input me, return me concrete.
33. Three Model Layers - App Model (JSON tags) -> Business Model (core) -> Storage Model (`dbUser` native DB types). Har transition par `parse()` se integrity check.
34. Single Responsibility - Task API sirf store, Picker sirf pick, Executor sirf execute. Ek service me do scaling axis mix nahi.

**C. Naming Conventions:**
35. Go: Domain package, Exported `PascalCase`, unexported `camelCase`, stage constants for DB values.
36. TypeScript/React: Component `PascalCase` file+name, Hook `useX`, API `camelCase` verbs.
37. Python/ML Sidecar: `snake_case` func/vars, `PascalCase` Class.

**D. Concurrency:**
38. Go: `goroutines + sync.WaitGroup + mutex + bounded semaphore channel`. Pipeline stages ordered jab dependency ho, independent batch parallel.
39. Contiguous Checkpoint - Concurrent batches out-of-order finish ho sakte hain, checkpoint sirf contiguous completed prefix retain karega.
40. Python Sidecar: 1 Uvicorn worker (ML memory-heavy), async + separate bounded `ThreadPoolExecutor` for embedding/rerank.

**E. Error Handling & Logging:**
41. Go HTTP: `response` helpers + stable error codes, service code `wrap with context + zap fields`, background job `ingestion_jobs` me terminal status record.
42. Sentinel Error: `ErrJobPaused` jaise sentinel for intentional non-crash stop.
43. Best-effort vs Fail-closed: Timeline/event writes best-effort, resume source-of-truth writes fail-closed.
44. Centralized Error Handling: Middleware me log once + inspect + respond, handler me dobara log nahi. `unknown -> 500` without leak.
45. Python: `HTTPException` with correct status code, service internal logs structured.

**F. System Design & Data:**
46. Postgres = Source of Truth, Redis Pub/Sub = notification/wakeup only. Durable append-only event tables + mutable hot snapshots.
47. Distributed Pattern `FOR UPDATE SKIP LOCKED` - `SELECT ... WHERE picked_at IS NULL ORDER BY scheduled_at LIMIT 10 FOR UPDATE SKIP LOCKED` mandatory for high throughput.
48. Observability: Har phase measure karo (`picked_at - scheduled_at` etc.), vanity metrics (queue length) par scale mat karo, task type + avg duration par karo.

**G. Libraries & Tests:**
49. Pinned Versions - Go 1.22, Gin 1.10, pgx/v5, zap 1.27, React 18.3, FastAPI 0.111 etc. Latest install nahi.
50. Tests - Go `*_test.go` with `testing`, Frontend `npm run build` (`tsc -b && vite build`), Python `py_compile` syntax check.

## Enforcement:
- Har NAYA task (jiska knowledge pehle nahi seekha) start karne se pehle ye file + Learnings/ folder read karo
- Agar knowledge pehle seekhi hui hai toh RULE 6 follow karo - user se puchho, fir kaam karo
- Har development task me RULE 7 ka workflow (Read -> Understand -> Propose -> Approve -> Implement) strictly follow karo
- Har code change me RULE 8 ke coding standards ko kabhi violate mat karo - yahi mandatory architecture hai
- Agar koi rule conflict kare to isi file ka rule priority hai
- Kabhi bhi is file ke rules ko ignore mat karo
- Generic knowledge use karne se pehle hamesha user se approval lo
- kuch file bnana ho ya koi task krna ho tho phle mere se permission lo uske baad me he vho kaam kro. 
- files ko create update delete patch kuch bhi krna ho tho phle approval lo.
- tumhare response ko as shorter as you can, rkho and only try to give me the necessary answer, response ko bde bde nhi likho, short and direct. 
