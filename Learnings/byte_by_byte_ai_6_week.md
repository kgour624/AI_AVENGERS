# Byte-by-Byte AI Cohort — Dense Knowledge Distillation

> **Source:** `Transcripts/byte byte ai - 6 week cohort.md` (~3,969 lines), instructor Sneha Mehra.
> **Focus:** LLM foundations, adaptation (fine-tuning / prompt engineering / RAG), structured output, grounding, evaluation.
> **How produced:** read end-to-end; only reusable principles, patterns, anti-patterns, concrete techniques, thresholds and quotable lines kept. Classroom fluff, intros, repetition and demo narration dropped.
> **Coverage caveat (important):** the file documents **Week 1 (LLM foundations)** and **Week 2 (adaptation)** in depth, plus their project deep-dives. Weeks 3–6 (agents/tool-use, multi-agent, reasoning models, image/video, capstone) are only *announced* — no body content exists in this transcript, so those themes are **omitted rather than invented**.

---

## 0. Core mental model / vocabulary

- "Whenever we hear LLM, it means it's just a **decoder-only transformer** trained on internet data."
- All modern LLMs share the same architecture; they differ only in **hyperparameters** (layers, hidden dim, heads) — bigger hyperparameters = more capacity but more expensive.
- **Base model** = completion model (continues whatever you give it). **Post-trained/instruct model** = answers + formats.
- Two-stage training: **pre-training → post-training**. "Both pre-training and SFT are exactly the same. Just the training data is different."
- **Parameters** are learned during training; **hyperparameters** are fixed before training.

---

## 1. LLM Foundations

### 1.1 The four training stages

| Stage | Data | Volume / quality | Compute | Objective | Output |
|---|---|---|---|---|---|
| Pre-training | Raw internet text | Trillions of tokens; **large quantity, low quality** | Thousands of GPUs, ~1 month, $100Ms+ | Next-token prediction (cross-entropy loss) | Base model |
| SFT / instruction fine-tuning | Curated prompt→response demos | 10k–100k pairs (**low quantity, high quality**) | 100s GPUs or fewer, days | Same next-token loss, **same code** | SFT model |
| Reward modeling (only if unverifiable) | Human rankings of SFT outputs | Win/lose pairs | 100s GPUs, days | Margin ranking loss → a score | Reward model |
| RL (PPO / GRPO) | Tens–100k prompts | — | Fewer GPUs, days | Reinforce high-scoring responses | Final deployed model |

- Cost anchors: GPT-3 175B <\$10M; GPT-4 / Gemini Ultra ~hundreds of millions; 405B LLaMA-3 needed FP32 ≈ 1.6 TB (≥880–100 GPUs just to fit, realistically 2000+), checkpoints 2–5 TB.

### 1.2 Pre-training mechanics
- Sample a text span → predict next token → compute **cross-entropy loss** vs. known next token → **optimizer updates all weights**; repeat over the corpus.
- High loss = prediction far from truth; low loss = close. Result is a statistically accurate next-token predictor holding **implicit world knowledge**.
- Base behavior is statistically correct but assistant-blind: `"I like machine learning because..."` continues internet-like text; `"How is the weather?"` continues instead of answering.

### 1.3 Data preparation (crawl → clean → tokenize)
- **Crawl:** seed URLs → fetch → extract links → queue → loop, with a `visited` set. Common Crawl: since 2007, ~2.7B pages, 200–400 TB HTML/crawl, new crawl ~monthly.
- **Clean:** extract real page text (H1/P), dedupe (repetition → memorization), remove unsafe/irrelevant sites, strip **PII**. FineWeb order: URL filtering → text extraction → language filtering → dedup → more filters → PII removal ⇒ 44 TB, ~15 T tokens.
- Named clean datasets: **C4, Dolma (~3T tokens), RefineWeb, FineWeb**. Rule: don't rebuild a cleaning pipeline — start from an open clean dataset.

### 1.4 Tokenization
- Text → sequence of IDs. Two phases: **training** (split + build vocab) and **inference** (encode/decode). "A tokenizer … is just a list and a vocabulary. It's nothing more."
- | Tokenizer | Split | Vocab | Flaw |
  |---|---|---|---|
  | Word-level | whitespace | huge | expensive; many OOV |
  | Character-level | per char | tiny (~105) | very long sequences |
  | Sub-word | merged pairs | ~50k–200k | best balance; all modern LLMs |
- **BPE** = start from characters, **iteratively merge the most frequent pair** until target vocab. GPT-2/LLaMA-3 use BPE variants. Rare/misspelled words become meaningful pieces (`un`+`believ`+`able`).
- **Special tokens:** `pad`, `unknown`, BOS/EOS. Too many `<unk>` = lost information.
- Use **tiktoken** / HF `AutoTokenizer`; **customizing a tokenizer is usually not worth it** (start from a recent ~200k tokenizer).

### 1.5 Architecture
- Neural net = "a sequence of parameterized transformations mapping input to output"; linear layer `y = Wx + b`.
- **Transformer** (Attention is All You Need, 2017) = a unique stacking of layers; **decoder-only** variant used for text generation.
- Full LLM = tokenizer → **embedding layer** (ID → vector; similar words land nearby) → N decoder blocks → final **linear** to vocab size → **softmax**.
- For generation you **discard all but the last vector** and read next-token probabilities.
- GPT-2 12 blocks/768 dim; GPT-3 96 layers/~12,288 dim; heads are a hyperparameter. Each size is **trained separately**.

### 1.6 Text generation & decoding
- Generation is an **iterative loop**: pick token → append → re-run → stop at `EOS`/`max_new_tokens`.
- | Algorithm | Class | Verdict |
  |---|---|---|
  | Greedy | deterministic | highest prob; **never in production** (repetition loops) |
  | Beam search | deterministic | tracks top-K paths; better, still repeats |
  | Multinomial | stochastic | can sample unlikely/wrong tokens |
  | Top-K | stochastic | fixed K fails across differently-shaped distributions |
  | **Top-P (nucleus)** | stochastic | dynamic K by cumulative probability; **"commonly used in practice"** |
- **Temperature** smooths/sharpen distributions; **no one best value** — tune per task. Code: low top-P (~0.1); creative writing: higher (~0.7+); captions: low temperature.

### 1.7 Post-training
- **SFT:** prompt→response demos with special tokens; identical loop to pre-training (InstructGPT ~14,500 private; Alpaca/Dolly/FLAN open).
- SFT isn't enough: multiple valid-but-poor answers exist → RL moves correct → **better, safer, helpful**.
- **Verifiable** tasks (math, coding): auto-check → RL (PPO/GRPO). **Unverifiable** tasks (writing): need **RLHF** — train a **reward model** from human win/lose pairs (**margin ranking loss** maximizes score gap); reward model = **proxy for humans**; then RL optimizes against its scores.

### 1.8 Evaluation (LLM level)
- **Offline:** perplexity ("no longer too helpful" — measures exact reproduction); **task-specific benchmarks** (datasets with correct answers); **human expert eval** (capable but biased).
- **Online:** thumbs up/down; **LMArena** pairwise comparison → Elo-style leaderboard (~211 LLMs).
- **LLM-as-judge** scores output quality for tuning (e.g., temperature sweeps).

### 1.9 Chatbot system design (beyond the model)
`input guardrails → prompt enhancer → response generator → output guardrails → user`, with a **rejection response** path.
- Guardrails block unsafe input **and** output.
- Prompt enhancer fixes ambiguity, typos, grammar, punctuation.
- **Session management** stores chat history and prepends it (per user, per session) and must be **bounded** (trim/compress) or the context window overflows.

---

## 2. Why adapt an LLM (Week 2)
- General LLMs fail domain questions (`"What is your refund policy?"` → answers for OpenAI, not your store).
- Three techniques: **fine-tuning, prompt engineering, RAG**.
- Startup rule: **start with RAG**, move to parameter-efficient fine-tuning **only if RAG is insufficient**.

### 2.1 Fine-tuning
- **Full fine-tuning:** update all params — very expensive at billions of params.
- **PEFT:** adapters and **LoRA** (freeze `W`, add low-rank `A`,`B`, learn only those). On a 3B model only **~3M params / ~0.1%** trainable. Library: HuggingFace **PEFT**.

### 2.2 Prompt engineering
- = logic that **transforms the user query into a new prompt**. Final prompt = **system prompt** (hidden engineered instructions) + **user prompt** (visible query).
- **Few-shot:** example Q/A pairs force a format (works even on base models).
- **Zero-shot:** no examples — put instructions directly ("return only JSON-like: `{years_experience: <n>}`").
- **Chain-of-Thought:** few-shot CoT (Google Brain, Jan 2023); **zero-shot CoT** = append `"Let's think step by step."` — PaLM 540B math **18% → 57%**. Foundation of thinking models.
- **Role-specific:** `"You are an expert tax advisor."`; also meta-prompting, prompt chaining, tree-of-thoughts.
- Context-window limit is exactly what RAG solves.
- Anti-pattern: applying CoT where it adds no value ("unnecessary complexity").

---

## 3. Structured output, grounding & validation
- Enforce structure with **zero-shot format instructions** or **few-shot format examples**; make the format explicit in the system prompt.
- Grounding rule that matters: "Use **only** the information in context… if the answer is not in context, respond *'I'm not sure from the docs.'*" — "**This line is very important. If we remove this, the model may hallucinate more.**"
- **Citations** need deliberate engineering: build context with numbered/tracked chunks + a metadata map, and specify citation format explicitly.
- **Hallucination is "not a solved problem."** Mitigations by leverage: (1) switch to a bigger model; (2) fine-tune (RAFT); (3) more inference-time budget to verify; (4) system-level grounding via RAG/web search.
- OpenAI hallucination paper (as summarized): models are penalized for wrong answers but **not** for "I don't know" → RL teaches them to *guess* when unsure. Fixing the **evaluation/reward mechanism** is the proposed direction.

---

## 4. RAG (Retrieval-Augmented Generation)

### 4.1 Pipeline
`documents → parsing → chunking → indexing` (offline); `query → embed → ANN search → chunks+query → prompt engineering → LLM` (runtime).
- RAG's core job: **narrow a huge DB to top-K relevant chunks** that fit the context window.
- **Chunking hyperparameters:** `chunk_size` (~300) and `chunk_overlap` (~30) — overlap prevents cutting meaningful boundaries.
- Algorithms: length-based (cuts mid-sentence), regex/sentence (no semantics), **specialized HTML/markdown splitters at element boundaries (best)**.
- Parsing: **layout detection → text extraction (OCR)**; Unstructured/Docling, LayoutParser; prefer AI-based over rule-based.

### 4.2 Embeddings
- Map text/images to a semantic vector space (Word2Vec origin; queen→king ≈ woman→man).
- Examples: OpenAI `text-embedding-3-large` (3072 dims), Gemini (top of leaderboard), Cohere, GTE-small (512 seq, 384 dims).
- Embedding models are **much smaller than LLMs** — usually affordable to use the best. **"There is no rule of thumb"** linking embedding size to LLM size; evaluate independently.
- Images: (a) shared-space **CLIP** (must use its *text encoder* for the text index too) or (b) captioning model → any text encoder.

### 4.3 Search / ANN
- Encode query with the **same encoder**; nearest neighbors; metrics **cosine / Euclidean** (use the metric the model was trained with).
- **Exact NN too expensive** at millions–billions of points → **Approximate NN** (clustering, tree, LSH, graph); small accuracy loss, large speed gain. Libraries: **FAISS**, Google **ScaNN**.
- `k` = number of chunks stuffed into context (default 4; demo 8). In production, **post-process retrieved items** (drop weak ones so they don't confuse the LLM).

### 4.4 Generation + RAFT
- Generation LLM = same LLM + prompt engineering (role, CoT, few-shot, user-context).
- If **retrieval is poor, the LLM has no idea** → **RAFT (2024):** label retrieved docs relevant/irrelevant and continue training so the model relies on relevant ones and ignores irrelevant ones.

### 4.5 RAG evaluation
| Aspect | Between | Measures | Typical metrics |
|---|---|---|---|
| Context relevance | query ↔ context | retrieval quality | hit rate, MRR, NDCG, precision@K |
| Faithfulness | result ↔ context | grounded vs hallucinated | humans / fact-check / consistency |
| Answer relevance/correctness | result ↔ query | correct for the question | similarity vs reference |
- **Debug method:** **evaluate each component separately and independently BEFORE chaining.** Localize the failure (bad phone number with correct context ⇒ the **LLM**, not retrieval).

### 4.6 RAG system design
`input guardrails → query rewrite/expansion → embed → ANN search (text+image) → prompt engineering → LLM (general or RAFT) → output guardrails → user`.
- Easy to add/remove documents (re-index) and to inject **live web data** into retrieval.

---

## 5. Cross-cutting principles & anti-patterns

| Principle | Rule / quote | Anti-pattern |
|---|---|---|
| Simplicity first | "Start with the simplest possible way… then gradually introduce complexity." | Jumping to agents/tools before a simple pipeline |
| Component-wise eval | Evaluate retrieval separately from generation, pre-chain | Evaluating the whole chain, not knowing what broke |
| Model strength | "Most issues can be easily fixed by just switching to a more powerful model." | Blaming the sampling algorithm when the model is weak |
| Cost/latency | Bigger model ⇒ more cost + latency; **VLLM** prod, **Ollama** local | Assuming an upgrade is free |
| Sampling | Top-P for chat; low temp/top-P for code & captions; **never greedy** prod | Greedy/beam in production |
| Context | Bound chat history; don't stuff all docs | Unbounded history / window overflow |
| Tokenizer | Start from ~200k tokenizer; avoid many `<unk>` | Word/char tokenizers |
| Versions | Pin library versions | Installing latest → breakage |
| Licensing | Open-weight ≠ free commercial use | Assuming open source is usable commercially |
| Grounding | Explicit "only use context / say I'm not sure" | Trusting grounding without instruction |

Operational notes (Q&A): LLM latency (10–20 s) breaks millisecond API assumptions; per-user/tenant vector filtering is an **open problem**; right depth of AI engineering = between math-heavy research and shallow API-only content.

---

## 6. Relevance to AI Avengers

> Mapping uses only concepts that appear in the transcript. Where a target task depends on content absent here (agents / multi-agent / expert synthesis), that is flagged as a gap rather than invented.

### A13 — Gate-1 vagueness check (cheap **zero-shot structured-output LLM** as a *second* check; problem-solving domains skip it)
- **Cheap zero-shot structured output is exactly the transcript's pattern:** explicit format instruction ("return only JSON-like") is the low-cost way to get a machine-parseable verdict.
- **Why it's a *second* check:** the transcript's recurring rule is **verifiable vs unverifiable**. Math/coding are *verifiable* (auto-checkable); soft judgments ("is a request vague?") are where an LLM scorer is justified. This **justifies skipping the LLM gate for problem-solving (verifiable) domains** and running it only for soft domains.
- **Double-check precedent:** "evaluate each component separately and independently"; the reward model acts as a **second scorer / proxy for humans** after the primary signal.
- **Anti-patterns:** CoT/heavy prompting for a cheap gate ("unnecessary complexity"); relying on a tiny model for a soft judgment (upgrade if it mislabels).
- **Fail-safe to borrow:** the "if not in context, say I'm not sure" pattern ⇒ gate should default to *not vague / unknown* rather than guessing.

### B1 — LLM synthesis of multiple expert answers
- **Supported:** RAG generation is the canonical synthesis step — the LLM combines query + retrieved items via prompt engineering into one grounded answer.
- **Gap:** the transcript does **not** cover synthesizing *multiple independent expert answers* (multi-agent week only announced). Use RAG-generation/prompt-engineering patterns; treat multi-expert orchestration as out of scope for this file.

### B8 — Verification-first (claim → evidence)
- **Strong fit:** **faithfulness** (result ↔ context) is the verification axis; grounding to provided context; **citations** (numbered contexts + metadata map + explicit format instruction); "use only the provided context" as a system rule; RAG grounding is the **system-level** hallucination mitigation.
- Rule: an answer not traceable to retrieved context is unverified / "I'm not sure".

### C3 — Evaluation harness / golden set
- Task-specific **benchmarks = datasets with correct answers** → the golden-set pattern; offline eval on held-out data + model comparison.
- For RAG harnesses: **context-relevance ranking metrics** (hit rate, MRR, NDCG, precision@K), **answer correctness** vs a reference, faithfulness checks.
- Process rule: evaluate each component independently *before* chaining.

### C3/C10 — Evals, observability, cost budgets
- **Evals:** LLM-as-judge when ground truths exist; component-level metrics; online thumbs + pairwise human comparison.
- **Observability/debugging:** check whether retrieved items actually contained the needed fact before blaming the LLM; LangChain logging; **context engineering** (trim/compress).
- **Cost budgets:** explicit cost/latency tradeoff of bigger models; embeddings are cheap relative to LLMs; per-task sampling params; test-time budget as a verification lever.

### B6 — Answer quality scoring & regeneration
- **Quality-scoring precedent:** the **reward model** scores quality vs human preference (margin ranking loss) — a template for scoring *quality* rather than pass/fail. **LLM-as-judge** scores output for parameter sweeps.
- **Metrics:** answer relevance/correctness; faithfulness for grounding.
- **Regeneration:** iterative decode + tunable sampling + "fix the failing component / stronger model, then re-run" support regenerate-on-low-score; explicit automatic regeneration is not described here and should be designed from these primitives.

---

*End of distillation. Themes absent from the source (agents & tool use, multi-agent orchestration, dedicated evaluation week, week-6 capstone) were intentionally omitted.*
