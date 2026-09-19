Sneha Mehra (00:00:00)  
computer. Nice. ⁓ Okay, today we'll discuss, sorry, yesterday we discussed single agent flow. Today we'll discuss multi agent systems and memory. Three prototypes, today we'll go slow. And what we'll do is we'll discuss, we'll start with memory and then we go into multi agent stuff. ⁓ So memory ⁓ is again the hottest topic. Or rather to be honest, easiest topic. Like again, memory is complex, I agree.

but everybody feels that memory is something that they can build because it's the thing that comes closest to database and extracting information like the kind of curds that we have been building. So there are you see so many memory startups, context layer startups, people building their own memory have coming because this is where you typically, this is like your classic API that you build that talks, fetches that relevant information and serves it. So now when what's

like different like memory one of the stuff that we use which is memory is nothing but like your context that we add and we keep adding to your conversation chat that is nothing but your in context memory ⁓ because that is part of your context window now given it is part of your context window it does not require any retrieval that you would want to get out of it like it's not an external system from which you are extracting something it's literally part of the system it's really part of the chat so one thing that we have been kind of touching upon

is if your context becomes big, your LLM starts to struggle with lost in the middle problem. ⁓ So most attention is given to the beginning of the stuff, beginning of the context and end of the context. At the middle, things suffocate. Right? That's always a problem. So that's why you typically do not want your context to become very massive and just keep ⁓ what's required in your context window. But it has its downside as well. As we kind of touched upon it, where if you

are having like extremely large context window. Lost in the middle is one of the problem. Cost is another because now you are paying for all the tokens because every LLM call is like full context being passed there. Now if you are compressing it or whatever you doing with that memory, if you don't put it then your prompt caching takes a hit. Then your caching token consumption reduces and then you pay more. So now that's why it's a very tricky problem to deal with. Now how should you deal with it? Again the answer is it depends.

Sneha Mehra (00:02:27)  
but depends on what scenarios is what we'll focus on in a bunch of prototypes and the systems that we discuss later. ⁓ But first type of memory is your current context window that you have filled with ⁓ everything you put the files that you load, the fragments of your snippets of the code that you load or the instructions that you give all of this parts of is part of your in context memory. Now, when this becomes big, of course there are strategies to evict some horizonal we'll discuss it in some time.

Second is external key value storage as a memory. This is where you classically use your Redis's of the world, your DynamoDB's of the world to hold. Now you can choose to hold your entire conversation as this ⁓ or you could extract facts and decisions and important aspects and store it externally. But the whole idea is that ⁓ when it grows very big, you can store everything in external system and it's a classic trade off that you pick

what's most important, what's most relevant, like how we discussed BM25 and basically cosine similarity based lookup, or basically your vector database, SNSW part, you retrieve it and that becomes part of your context, like that becomes your next LLM call. Okay, now who decides what to retrieve and how to retrieve and what's relevant, again, it depends on your use case. But I'll take two examples. First ⁓ is it's not.

that it's always going to be a semantic lookup or a semantic retrieval that you are doing. ⁓ Sometimes the information like what was the rate limit of this customer? Don't think this as a prompt. Think of this as a requirement that you want to bring in. ⁓ Now, ⁓ you, this could be in your agentic loop. It could be a tool call. It says, ⁓ hey, I want to retrieve the rate limit for this customer, do whatever reason, right? And then there's a tool call which is written. The tool call goes, ⁓ makes a call, ⁓ literally,

creates that key that user ID rate limit. ⁓ Again, you have to provide it in the tool description, what key format it is and how you get that information. It goes there, makes a call, gets a response and adds it to your context. ⁓ That is what you are doing. So now here information is not just with respect to chat that you are doing or your agent flow that you are doing. This information is also could be user configuration, user preferences. ⁓

Sneha Mehra (00:04:50)  
or some past historical relevant chats or some key decisions that you think should have been persisted ⁓ over a ⁓ long chat. ⁓ All of that that you can store externally but the whole trade-off is your ⁓ good part is your context is small, your working context is small but the bad thing is you have to retrieve it which means your key needs to be very deterministic in way or the query that you are firing or the MCP tool that you exposing needs to be very deterministic in way.

So you can fetch the relevant information, which means your tool definition, input schema and all becomes very critical. Now, what are these Excel systems? Although I took examples of Redis, Dynamo, B2B, but think of your elastic search also as this. Like not just like don't just think of Excel key value memory as vector. ⁓ Any database that supports vector also is an external memory, but we'll cover this in vector memory, which is a semantic lookup. But Excel key value store is like named facts. Given this, give me this.

⁓ That is our external key value store. also it need not be just key value store. You can make it anything but key value store is most popular because you're typically looking for pointed stuff. I want A, B, Let me get that information by doing a multi get, et cetera, et cetera and proceed. So this is where you're again, if you look at it, is where a classic system design typically comes in. Where

how much of database to provision, how do you make sure availability of it, how do you decide key patterns, now are you doing range lookup or appointed lookup, et cetera, et cetera, you have to be aware of that. Then comes elastic search to do keyword based lookup, which is a BM25 or TFIDFs of the world. ⁓ There you need to know what analyzer you are choosing, how you are querying the data, how your elastic search query is being formed, what kind of data you need, what your ranking strategy is, and then what you're, if you are using.

⁓ LL rank ML rank whatever that thing is that does re-ranking using a deep learning algorithm within Elasticse I forgot the name of it. If you're doing that then that entire feedback loop needs to be so this is your entire information retrieval as a domain. ⁓ These are external key value memory external memory that you are stored or that you are holding. Then comes your most popular nowadays which is your vector memory where you are essentially doing a semantic lookup. Now here

Sneha Mehra (00:07:06)  
I kept these two separate, is external key value store and vector, because I wanted to emphasize on configurations, on preferences, on graphs that we are trying to fetch. And then cosine similarity. Now there are databases which has merged both of these two seamlessly. Like Elasticsearch also supports vector lookups, your Postgres supports your transactional queries and your selects and your wares and indexes, et et cetera, and also supports vector, like through a PGVector extension, it also supports vector lookups. But this is where the whole idea is,

your embedding of your document or your query determines what is more relevant to you. ⁓ Similar to what we discussed in previous weeks. ⁓ So this is where you're doing semantic lookup. But again, lines are now becoming blurry because every database is a vector database, but you can treat them separately so that you are not just always doing a semantic lookup. ⁓ point here is don't treat that every rack solution or everything that I'm building has to just blindly go into a vector database and just retrieve it and that's all you need.

Always remember, treat ⁓ what you like, think, look at your problem statement, understand the flow and then decide what you would want to do. Next up is now getting closer to how human brain works which is episodic memory. Fancy word, think of episodic memory as important facts, important decisions is what you trying to hold like sequence of events that happened. ⁓ That is an episode that has happened.

and you are extracting from your conversation this stuff again you use LLM to extract this information and then you put it into again this episodic memory can be stored in your Redis, DynamoDBs or ⁓ Postgreses of the world or your graph database depending on how you choose to query it. So these are key thing is you are extracting the key facts out of it or key events out of it and then ⁓ you are storing it somewhere making it easy for you to retrieve.

Now episodic memory is also a good way to summarize the information. Like for example, if a chat has happened, it has gone really big. Now you want to summarize it. Either you make an LLM call, say, hey, whatever is there is summarize it. Or you can write a more sophisticated prompt and say, ⁓ but hey, this is a conversation history that I have. ⁓ Please make sure you are logging or you are summarizing it by not losing A, B, C, which is your key facts, key events, key decisions you made, ⁓ et cetera, et cetera, et cetera.

Sneha Mehra (00:09:31)  
⁓ And that is a good summarization prompt rather than just blindly saying it. ⁓ this is again summarization is not the only use case but imagine episodic memories like hey I used to like pizza now I like burger. ⁓ That is also a key change in preference because of some event that happened that you went to a place you tried burger for the first time and you loved it. So this is what you have to understand again there is no generic way you have to look at your system and see given the system

Are there any episodic traits that I could extract out of it? Are there any events, insights, decisions that I could extract out of it which I certainly don't want to lose. You can convert it into embeddings if it is verbose text but very likely you would be extracting it as facts and very likely you would be querying it as a structured data rather than just a semantic lookup because for that you have vector memory where you have verbose of like high verbosity, text, incidents, etc etc that you are

trying to retrieve. But episodic memory in a gist is more about the key facts, the key events that you would want to extract. Now, let's dig deeper into in context window because it's the most important one because that directly impacts your cost. Now, your key or your working memory is ⁓ your main context window that you have. That is the most important thing for your agent because what you have in the context is what's getting passed into the prompt. Now, that has limited capacity.

Right? You may have 1 million token, 2 million token. Remember the old days where you are 100,000 tokens as limit. ⁓ So 1 million, 2 million is 1 million is what most people use. But again, it's still in for some order, it's still even lesser. So it does not mean as I said, it does not mean if your context window is 1 million and if you pass 1 million tokens to it, it does not mean that every token gets equal attention. I'm reiterating. Remember this, there's a lost in the middle problem. ⁓

First thing gets more attention, last thing gets more attention. There's a very interesting paper which says if you copy paste your entire prompt twice, it gives you better result. Literally. Like by just doing nothing, you just copy paste your prompt twice and you'll get a better result. So it's it's like kind of acting as a forcing function. It is giving your task enough attention that it deserved in the first place. Right? Okay. So now what is the challenge with context window?

Sneha Mehra (00:11:56)  
First, it's limited, that's a constraining factor. Second is what you put into context. You put your system prompt. Yesterday we saw with paper how big our system prompt became. ⁓ If not, when you're reading the entire paper and putting it into context, every iteration of your next next prompt or next next chat, all of that is going into the context. Another thing that goes into context is tool definitions that we saw. And we saw how verbose our tool description was and how non-ambiguous we had to make. Now all of that goes to your context.

Then goes user prompt which is things entered by your user. Then goes even responses that your LLM generated or the user responded to an LLM query goes into your context. Then your file content. Yesterday we saw how the code that was written was part of the context. It fixed it, ⁓ it, et cetera, et cetera. Then you have intermediate observations. This is what your reasoning loop that we spoke about is added. That the thinking of your model is outputted as a text.

which is again added as a context in the chat. Right? Okay. Now, given this is a fundamental thing that our agent is working with, the only thing that our agent can work with, ⁓ how do we manage it? Now, there is no one right way to do it. It's not always summarization. So, there are different strategies to deal with it. First one is eviction. ⁓ For example, if a context window has become really big, you can choose to remove the content. But as we kind of briefly touched upon it yesterday,

Which content to remove? don't know. And imagine you remove the oldest content, then your prompt caching takes a hit. That's another problem. Right? So then you have to be very mindful ⁓ of what you are revicting and why you are revicting. Some examples. So for example, ⁓ once the decision is taken, any reasoning step that you took to reach to the decision can be removed. Or if you have verbose tool outputs, that could be removed.

One simple example of that we kind of saw yesterday was like your research paper stuff. Either we put your entire paper into context, then it decided what all section it had to read ⁓ or ⁓ once that is done, that step from there we extended what are the sections ⁓ and then this step finished, we hold it in external storage. Now this entire context can be ⁓ removed and evicted because now we don't need it. So you decide that in this single

Sneha Mehra (00:14:22)  
agent loop that you have, you first reading the paper, give full paper, the ⁓ sections and the text of it. ⁓ Once that job is done, you remove everything from the context and just add that one simple bit to it. That makes your life simple. So ⁓ again, like always, there is no one right way to do it, but you decide what ⁓ is something you don't need in the context. It could be your

subtask whose result is there. example, your calculation is done that 0.1 % population of India are like calculation that we did. Imagine that calculation is done. Now we have the answer. Now we don't need the tool response. We can choose to remove the tool response. Because our job for that is done. Now imagine the complexity of you tracking what is important and what is not. Because the biggest problem when you're trying to evict something from your working memory, which is your context window, which is your conversation is

We don't know if at step 8 we would need the output of step 3, like output of step 3 at step 15\. So we are at step 8, step 3 outputted something, but we don't know in future step because we only have visibility to the 8th step. We don't know if in any of the future step I would need output of step 3\. So until and unless you are very sure, don't evict it. And once you're evicting it, that's why.

what your reasoning traces were, you can choose to have it once the job is done. You could choose to have a subtask result once the job is done. Now imagine the complexity you need to build around it to decide what should I remove, what should I not. Right? Okay. One specific example of this is imagine you're building a ⁓ LLM based trading agent. Nobody should build it, but imagine you're building this. Now imagine in context window, the past

trades of a particular stock are added. ⁓ You don't want your LNM to act on your outdated signals. as and when you keep moving ahead you keep evicting stock from your context because you would want to deal with your latest signals that you are getting and not the outdated ones. Again a trivial example but you see the reason for you to remove it. Or if you take example of a coding agent where you have a long running coding agent then you kind of

Sneha Mehra (00:16:43)  
or you're kind of doing a massive refactoring. Once a particular file is refactored and it's working fine, you can choose to remove it from your context. You don't need it anymore. Right? So what you remove and how you remove completely depends on your use case. There is no one right answer, but hence you need to understand your use case really, really, really well. And don't just blindly. That's why my job here is to serve the buffet.

and tell you places where you need it. Now what you choose and how you choose is something that depends on your use case. And now that you see how tricky evictions can get, hence now here think of it that why people rely on summarization mode. Because it is pretty general purpose that way. And hey, once I summarize the result, I make sure I capture all the key information, but there is still risk of it being lossy, it removed or it skipped.

something which was more important, let's say a output of a tool call that you need later. ⁓ So, but it's usually safer to do it and very general purpose. ⁓ So that's why you see a lot of people by default inching towards summarization, but you can choose to do ⁓ eviction of things that you are sure of that you would not be needing. ⁓ So if you have a very deterministic workflow that you have built, then you know that once this step is executed, so I'll

Give a concrete example. One of the agents that we are building at workplace has 15 steps. And after first five steps, ⁓ all it matters to us is the output of that fifth step. After that, we don't need it first five step. So we just got rid of it and that helped us save the cost. Very simple example. So that's why if you have an idea of what your use case looks like, if you have an idea how the chat is going to, or not just chat, it's not just human who's chatting.

But if you know that this agentic loop, this is it's going to behave. Until this point, I would not need this thing later. Then you can choose to remove it. So that is where imagine how the implementation, you have a list of strings, which is your chart. And for each one, instead of just string, make it a dictionary. And each one you tag it with what it is doing. If you have a very deterministic agentic workflow or loop that you're trying to build over.

Sneha Mehra (00:19:06)  
Now that we know why summarization is so general purpose and important. ⁓ here the whole idea is it takes your older messages, compresses it. It tries to retain the summary, but this is always going to be lossy. Right. And it slow because it involves an LLM call, but where is it okay? It's okay where you want to first not let your context window explode to save costs and like not hallucinate, et cetera, et cetera. But more importantly,

Practical examples of it is let's say your customer support chatbot. Where the conversation has flown a lot and because it's human and AI kind of talking like human language, there is not much decisions that are being taken. You could then summarize it and just replace our entire context with that one and then let the chat continue. Because in most cases you are just doing chit chat or like humanish chit chat and you could save yourself some tokens. Then another is your

legal discovery system where you are hunting for a case which which supports your argument imagine you are a lawyer and you are fighting a case and for that case for you to prepare for that case you finding a supportive argument a similar case in which court gave a certain kind of order ⁓ once that is discovered let's say you have multiple steps into it and let's say one of the steps found the most relevant one and now it will continue to hunt once that most relevant one is

found you can remove everything from it or you could summarize and say okay this was a case this is exactly what happened this is this is the final verdict of it and this is a supportive argument so what it took as multiple steps to figure that out and reason about it get summarized into this was a case this when it happened this was a party this was a verdict and this is a supporting argument that you could make so this way you could

it seems lossy because now you have shrunk the text to a very small size or very small length but what it did help you save is some tokens similarly think of research assistant when your research is done there a of exploratory questions that you ask you ask it to summarize it would summarize and just shorten your window so but key thing to remember is you should be okay with it being lossy number one

Sneha Mehra (00:21:30)  
for you to have a summary of whatever the chat has happened up until now is okay. It's good enough. ⁓ That's number two. ⁓ And you make sure you ask your prompt to capture the details and those details are something that a prompt would not miss. So you may have like multiple loops to check. You are certainly not missed any if you are slightly unsure that your existing model might miss a few things. So then you trade off by making multiple calls, ⁓ multiple NLM calls to summarize it so that you don't

lose out on that information. Next part is you have eviction, ⁓ have summarization. Next part is moving to external storage as we discussed, right? Where external could not just be another database, it could also be disk. Imagine you are just building something that just works on single machine, you can just dump it on the disk and make your life simple. Right? Now here, for example, one of the most popular agent that does this is a coding agent.

The coding agent you would have observed creates temporary Python files or temporary test files to do this, to test your changes. These are nothing but your disk bits. Like now when it needs to load it, it can load it. Otherwise, it's just storing it on the disk and like using it. Right. And then it just forgot to do that many times where it forgots to delete it. But if you choose to, you want to delete it, you can delete it yourself or you can ask your agent to do it. But the idea is you can move this part and

store it externally and then retrieve it. So either you evict, so this is similar to eviction, this is a strategy to evict, but it's not that now here imagine this you can extract key facts and store it in a structured format in a database ⁓ or you just take verbatim conversation and store it in the database, et cetera, et cetera. But the whole idea is you're managing your context window. You want to keep it simple. Now what I'll do is I'll take an example. After this we'll take questions. I'll take an example ⁓ of managing context window.

So what I want to demonstrate with this is number one, how we know that we have hit this limit where this code gets plugged in, what are the things that we do and what are the repercussions that we see. ⁓ So I've injected or induced an artificial budget of 2000 tokens. I'm doing a 10 step DV migration plan, very simple. And I'm mocking a lot of tool outputs. I want to keep things very simple. But what I want to demonstrate is what do we gain and what do we lose by doing what we do.

Sneha Mehra (00:23:56)  
So let me share the screen. ⁓ This is the working memory part. I will go to working memory code. ⁓ OK. ⁓ Sure.

Okay, so here what I'm doing is I'm taking all the three strategies, eviction, summarization and moving to external storage. I'm doing all three, right? And the whole idea is because my context is finite, I have to deal with that. So instructions, you are a senior database migration analyst, you receive investigation finding one step at a time, analyze each step that you provided carefully, note exact command, constraint numbers, action items. You may be asked to recall any specific details later. So you obviously what.

kind of eval is what I building here. will at the end of the summarization or whatever I will ask it some questions about it which should have been in context or not in context depends on if your summarization is lossy or you did eviction and that information got lost. Now what I did over here ⁓ is here so I have multiple steps which says I want to take schema snapshot, want to do row count and volume so now these are like tool outputs that I have walked. So this is step, this is the question that I have.

Given this data, do this. And this is imagine this is a mock tool output. ⁓ It's too hot, no. Okay. Then you have index inventory. Then you have foreign key audits, ⁓ stored procedure audit, application query scan, et cetera, et cetera. ⁓ And here you have your development checklist. Now there's a multi-step operation that I'm doing. Now let's look at summarization, summarize history. How does the code gets involved?

Now here, if you look at this, I'm iterating this step by step and I'm saying every summarize after every five steps summarize what has happened. Replace the context and replace it with summary. It goes over here. It summarizes history and then it replaces over here. Compressed summary goes over here. ⁓ And then it continues the chat messages here. There's a user message. It formed over here. Step number one tool output, et cetera, et cetera. Here you get this output.

Sneha Mehra (00:26:06)  
Okay, now here I'm multiple phases, phase two, phase one. Let's go to eviction. ⁓ eviction. Now eviction, I'm doing simple sliding window based eviction that I'm trying to evict as per my context budget. So the moment my current tokens exceed my context budget, I evict evict evict evict, literally removing it from my context window. So this way, my messages are getting shortened. Here, look at this, my messages, you say I popped bunch of stuff.

to be well within the limit. So I'm hard coding it for now to be, where it go, summarize token, token count context budget. So I'm artificially inducing the limit of 2000 tokens. Artificially, you can go till 1 million because it can go further, but I'm artificially limiting it so that I can show a demo. This is my token bar, okay. Then we saw, now we saw that we saw summarization. Now, summarization.

summarize after, tokens only, summarize history. Now look at the summarization prompt. Summarization prompt looks like this. Comprehensive following investigation step. Now here, it's not a generic summarization that I wrote. It's specific to what my task was. So that I explicitly mention what is important for me and what is not important for me. Given what we are doing is a database migration. So what is important?

is exact commands, numbers, table names, constraint name, procedure, breaking changes, action labels. That is important to us. Right? If I would have written generic, summarize this. Very likely it would have skipped a lot of this, ⁓ a lot of this important stuff. Hence it's important to be very prescriptive in nature. How your doctor is prescriptive, it says one medicine morning, afternoon, evening. In afternoon, if there is no medicine, it writes X, zero, X. Right?

It's very prescriptive and it exactly tells you which medicine to take and at what time. Very similar to this. Given my use case, I'm exactly telling what it needs to do and how. Now, I just added use extreme abbreviation, max 100 words. Every token must carry information. So I'm making it super dense. Now, why I did it? Ideally, I should not add this. But the reason I added this is to show it's lossy. I wanted to quickly converge to its lossy behavior. That's the only reason why I did it. ⁓

Sneha Mehra (00:28:33)  
Okay, so summarization is done ⁓ and eviction, summarization and external storage. So let's look at external storage. So for external storage, I'm holding this data now into ChromaDB. I'm using simple semantic lookup out of ChromaDB to find relevant part. That's it. That's all. So here the query goes, ⁓ collection.query, query text and results I would want. And I proceed further with that information to all user. This is my user message that gets formed, which is that context from my ChromaDB.

that I get in results. ⁓ Okay, let's run and see what it does. So when I run, I have some output, but at 730 ish something time, I ran this code first time ⁓ and I got a mini heart attack. I got this 503\. Well, if Gemini is down, by the way, this is because Gemini was down for some time. I'm like, so which is why I always have.

Output saved and for this example I did not have any output. not this example. Working memory I had one. Working memory here. I always save output. Always. So that if Gemina is down, I have my fallback plan ready. ⁓ Okay. Sorry. Now here it's running, running, running. By the way all this fancy is just my ⁓ terminal output skill that I have written. ⁓ I write the code first. Ask a realm to modify it. And then I ask it to apply my terminal output skill. So it renders it in a very beautiful way.

So that's why my code looks very verbose, but it's just filled with this terminal output stuff. So that looks neat and clean. Okay. Now here, what it did ⁓ is we take the first phase here. ⁓ yeah. Phase one eviction. The moment it went over budget here, it went over budget. It evicted few things, right? ⁓ And then it proceeded further. Then I have eviction sliding window. Now here we'll go exact. We'll go at the meeting. Hey, Gemini down again.

⁓ What's happening? Okay, let it run. ⁓

Sneha Mehra (00:30:37)  
this one is still ready. I suppose hold on. Okay. Here. So here, if you look at this, now what I'm doing is I've given it 10 step stuff. I have token budget of 2000\. I'm captured for each step, how many tokens are getting consumed. Right. Now here you see the moment it crosses 2000, if I reduce it even further, goes like, okay, before we go into that. If you look at it here, it summarizes. So here the summarization happened every fifth step.

250, 428, 591, 988, 1213 because we asked it to summarize every 5th step it summarized then my context window shrunk and it increased further now here eviction we were evicting when my budget hit 90 % and then the eviction happened look at this increasing and before the 10th step because it hit the 90 % of my total context budget which is 2000 it triggered an eviction ⁓ of historical message the first message that it wrote it would have evicted that message

and here as and when we kept taking decisions I kept putting it into chroma dp. ⁓ Now when I tried to fire a recall for the first phase which was eviction that we did it could not find the result because eviction is lossy. Second for summarization it was lossy it could not find. ⁓ Now what it tried to find recall evict

⁓ Do ⁓

Sneha Mehra (00:32:07)  
called check record here. Answer. Chat, pull chat, where did go question, eviction page, evict message, verification question here. Okay. Now this is what the first step or the command in between it does. So I want to make sure that it doesn't lose that information. What is the exact rollback command and exact rollback time window that was documented in the initial schema analysis, give the exact command string. So

I just said the expected keywords I looking for are these. If it's not there, it's a problem. ⁓ So you could use this in your evals if you trying to build something where a critical information in your system should not be lost and given that you know what kind of system you are building so that you know what is important for you, you try to capture that information and add evals so that if tomorrow someone changes your eviction strategy or other your context management strategy, you have your evals in place. ⁓

to hunt that, to find that error for you. ⁓ So to report that error for you. ⁓ So this is what I tested and here we see what works and what doesn't work. For external storage it was retained because what it was looking for was already there in ChromaDB. It went, queried, got the thing and it could output what we are looking for. Here it's there, based on this assessment, step final workflow. Okay, here, if you look at this model response, the exact role that command documented in the initial schema analysis is this.

The ruleback time window is four, source, post migration, wall segments are pushed after that. Again, it went through that. It could figure out the command and it's outputting. ⁓ It will run, it will take some time. But what I wanted to certainly demonstrate is like different strategies to operand where you would put your code. Now you can make, ⁓ one thing that I did is ⁓ token budget.

Here that I had token budget.

Sneha Mehra (00:34:04)  
context budget here the context budget that I added is here I'm just making my best guess. So ideally it's not the exact like count tokens only if I look at this here if I would want to count tokens I can do here I'm making a call to model to count it again it does not make call to model still internet call it's still a network call that you are making but you could still do it in a very ⁓ simple way you could just say

my tokens is equal to number of words that I have in the text rather than wanting like rather than having to make a network call to do that or you can use any local model that ⁓ bakes in the information that gives you a rough estimate to do it because not everything needs to be a network call to get exact precise information. Like imagine you are doing it for summarization even if you do it like 100 tokens early or if you are off by 200 tokens or 300 tokens not the end of the world for you. Right? So you could make a fair assumption and just

work with that. ⁓ Okay. Any questions up until this stage while this runs, it will take some time, but while it runs any questions for our own context management. ⁓ Good so much.

Sneha Mehra (00:35:15)  
Arpit, we discussed summarization and ⁓ LLM models also give something called compaction. Are these both same different? Summarization and compaction is same. And so when you say LLM models gives you compaction, is the Claude agent SDK is what you're talking about, right? Or your Claude code is what you're talking about. So that is nothing but your infinite while loop that we say and it picks up. Now, there is a logic that is written in Claude. It says, how do you compress it? How do you comp

packed it, right? So that is a different branch on what strategy to apply, how to apply given the session. You can add a lot of smartness and go down the rabbit hole on how to best compact your working session. ⁓ I think you had a demo for that, right? Where you made a network call and saw it. ⁓ Yeah, I did. also, so you already have that. Can you go back to your prompt that summarizes? So you have given instructions on how to summarize. So, yes, ⁓ Claude has the same thing.

In addition to that, Claude also gives you an option to give your own prompt, which is, know, the conversation that has happened when you run slash compact, you can add another prompt as a parameter there saying that always remember X and Y facts. So ⁓ it also has this capability to inject your prompt, but by default, what Arpit has in the POC is what Claude does because it has a specific prompt that it will run whenever it tries to compact.

This one. So preserve only critical packs, commands, numbers, table names. So again, as I mentioned that this is my, and I didn't know Claude, we could give our stuff to it. How do we get that Pratik? ⁓ So it's slash compact and then you ⁓ can, it's not just slash compact. So one way to run is slash compact. It will run the compaction by default. Then there is slash compact and you can follow it with your own prompt. So slash compact, preserve key information, which is user related. So then it would focus on that information in the...

⁓ history and try to represent that. Yes, Subat hope that answers. Yes, yes. And one follow up question is in cloud documents, says sonnet doesn't support compaction and ⁓ opus supports. Does it means we cannot summarize? This is a custom summarization, right? Which we can do any realm ⁓ in this case. Yes.

Sneha Mehra (00:37:40)  
So given this is our flow, we can summarize and compact however we like. ⁓ Deepesh, do have a question? ⁓ Yeah. So when you're talking about the eviction, right, context eviction, do we evict the prompts which were given by user also? Or is only the results coming out of the like either tools or tool calls or some other network calls? So like do.

Can we evict both things or can we only the results? ⁓ imagine this. you are evicting, ⁓ imagine this you just evicted messages of zero. That was a system prompt. ⁓ You just evicted a system prompt. ⁓ Ideally you should not. But a naive eviction strategy which is a sliding window based, you evict your system prompt. Now you can make it smarter and say, I won't evict my first three prompt or first three messages because it contains my system prompt and what user wants to do. That is your use case specific. ⁓

and then you say after this I will keep shortening stuff like I will keep evicting stuff that I want. which is where you typically have a system prompt place explicitly your conversation messages is part of a different array so when you are evicting only a conversation messages gets affected but there it could also be like you might accidentally like because of our eviction strategy you might evict your actual task that you just started with. But now imagine how difficult it gets if you have a general purpose agent how difficult it would get like

What do I evict? What do I not? Now, if your task is very specific and you know exactly what you want to do and what you want to get, that's much easier. ⁓ But if you don't know what's important, that's why you see almost everybody or almost all Frontier Labs start with summarization as the best way to do it. Because you don't know what needs to be evicted. Imagine you evicting something which is super important just because you use this naive sliding window based approach. ⁓ So answer is it depends on how you do it. ⁓

But a good practice is system prompt separately, ⁓ messages stored separately in an array and when you're passing for your next prompt, ⁓ you concatenate this and this and then you pass it. ⁓ Got it. ⁓ Another question on this summarization, right, which Sumanth was asking. ⁓ So like he mentioned Opus. ⁓ Opus can do but Sonet cannot do summarization, right? ⁓ I'm unsure on that. I'm slightly unsure on that. We are unsure but like just assume it's true.

Sneha Mehra (00:40:02)  
Let's assume it's false. ⁓ Because if you look at it, ⁓ summarization is nothing but your historical messages. ⁓ That's what I was... So that's why I would by default assume it to be false. ⁓ But again, ⁓ because how difficult it would be to summarize the stuff. It should not. Because we are using Sonnet in our broad LLM agent. ⁓ It's our investigation agent. So I was just wondering how our system is able to compact it. ⁓ I was like, bro...

You have evidence, have anecdotal evidence. So you should have assumed it to be false. ⁓ Thank you. Thanks. Ariyant, go ahead. Can we again go through the retrieval part? I'm very fascinated, you know, why the context consumptions are so less in that part. ⁓ how is it? So ⁓ there are three parts to it, right? Eviction, summarize and external DB. ⁓ Yeah. So in that part, like what is really happening?

So eviction, if you look at this, this is my cumulative context that I'm passing. Yeah. In every call, right? What was our eviction strategy? The ones I hit 90 % budget, goes away. Yeah. So as soon as I hit here, if you observe, it never hit 90%. It ⁓ never hit. So there is no ⁓ eviction that got ever got triggered. That's why you see retained. ⁓ That's why non-intervening waiver is so good. I could see different outputs. Yeah, before it was lost. Right? ⁓

For summarization it was lost because we were expecting the exact command. Although we said it in the prompt that preserve commands, it could not due to whatever reason. And that's why you see the moment it hit ⁓ this, which was ⁓ we asked it to summarize every fifth step, we got this and then it summarized. So then the summary was literally shrunk from 1000 to 220\. Now it's also fault of our summarization prompt where what we asked it to be dense something here. ⁓

look at into a dense and every token must carry information. Now this is enforcing it to be very ⁓ succinct, extremely short, extremely abbreviated. Yeah, actually my question is on the external DB part. So, let's go. ⁓ So then literally every single line is getting added to external DB. So every time the call happens, fetches the relevant semantic information and then running it. I'm not even sending it. If you observe,

Sneha Mehra (00:42:28)  
the usage is not growing much. Yeah, yeah, that's the most fascinating part. ⁓ Because I'm pushing everything to external TV. Here, ⁓ I will show you. And is it making a search call ⁓ at every step? Here, receive context, query text. If I get that one, I query, I get the retrieved documents and then I'm adding it and it goes. And then I immediately add a document to this. So depending on how many relevant document it could find, depending on the step, it's fetching it.

from the database every time and every subsequent conversation or message once it's outputted it goes into my DB. So I literally it's a nice implementation where I'm literally fetching this is kind of we discussed yesterday or last week I forgot when that what if we have this I think it was yesterday when we were building this coding agent what if we retrieve everything in the Ralph Lou what if we retrieve everything and then we run it right this ⁓ is an example of that where every time we are literally making a call

except for the first one because there is nothing to be made. Every time I'm making a call, getting the result and then passing it as my context and getting the result and then adding it back to context. ⁓ This was that example. Yeah, yeah. No, this is pretty interesting. Thank you. Sure. Thanks. Ankit, Kevin, I'll take your question in some time. Okay. Awesome. Next up, again, it's not very heavy, heavy session, so don't worry. We'll have time with you. Okay.

So next, it's already one hour, shit. ⁓ Sorry, it's a heavy session. My bad, ⁓ one hour again. I didn't even realize. Okay, next up ⁓ is ⁓ memory. Aha, round figure. Seal function I applied. Math.seal. ⁓ Okay, ⁓ memory compression and summarization strategies. So now, we saw ⁓ summarization, how it happens, right? But that was a...

prompt that we provided. ⁓ different kinds of prompt. This is very similar to the prompt that I just gave on how it needs to summarize. Like you are summarizing an agent conversation for memory compression, preserve all decisions, numerical values, et cetera. You saw how similar the prompt looks. I just specified in addition to this a specific output format. Given my use case, given my use case, would want to do this.

Sneha Mehra (00:44:47)  
So now here you could play around with what is the best way to represent this information for your use case for agent to consumer. could be decisions. ⁓ Imagine you don't need to have any decisions to be captured, but let's say metrics are important for you or what some numbers are important for you. You try to get that. Right? More interestingly, your hierarchical memory consolidation. Now this is where your graph database is typically kicking. Where as and when the conversion is happening and let's say I'm extracting

Decisions open items. Imagine you are doing let's say ⁓ agent ⁓ you are building an agent for an agentic SDLC. In that case open items should be going into your GIA and individual tickets needs to be created. that is where at end of the conversation you want to do it. You don't want to make tool call on every ticket. It would be wrong. Let's say you want to create tickets ⁓ after your entire thing is reviewed by a human. Okay, this is all good. Now I create.

So in that case this sort of structured information helps. This is also very true for your Google notes type system. I'm just making notes so that I could example. So let's say note taking apps that you have, it comes in handy for that or your JIRA tickets. But that you would want to do at the end, at the end of it after human in the loop is done.

So which means when you're summarizing, you'd want to summarize all the key decisions were made, all the open items separately, so that when you're showing the final Google Meet summary that Gemini sends you, has this in a very structured format. Like, hey, these are the next steps, what are decided is a summary, who said what, key decisions made. Now here, because my use case wants me to do this in a structured format, I'll ask it to do this.

⁓ The answer is always it depends but how it depends is what I try to bring in. ⁓ The third one is the more interesting one. The more human the use case, if you are trying to make your agent mimic your human behavior then you have to have a weighted retention. We as humans don't retain all memory equally. Not every memory is equally important otherwise we would go nuts. What is important and what is not?

Sneha Mehra (00:47:06)  
is typically what is more recent is what is more important for us. Everybody's talking AI because that's a recent take. ⁓ Everybody's talking, hey, Fable did this, Fable did this, one week later, nobody will remember. So ⁓ what was an important information is exponentially decayed and eventually forgotten. So here you could add a weighted retention. For example,

depending on the time when the information was registered and the time it spent it follows an exponential decay e raised to minus x or any variation of that you apply and it decays as per that right and once it reaches it would never go zero but once it goes below let's say 0.001 it is forgotten so this way you're kind of mimicking human behavior where something was important i like pizza three years later i what was the dish that i used to like shit i forgot

It was round and had this and then you try to recall. you remember something but you don't remember that exact information. So you're trying to do this. Now where this comes in handy, this comes in handy where you're kind of building system that closely works with human and is human like. Think of personal assistant. So what was it? Because if a personal assistant captures everything that you want, it should not keep every information ⁓ as is because your preferences change, your memory evolves.

You used to like X, let's say used like Java but you worked on Go, now you like Go more. ⁓ So eventually you forget Java. Same stuff goes over here. So now imagine you had things, you had chat which says I like Java. Three months later you say I like Go. One year later you ask your agent what is my favorite language? It should say Go and not Java because Java has an exponential decay and you forgot about

So now this is also a way for you to reduce your memory. Again if you look at it you are kind of mimicking how we behave as humans. But in this case it's not always true that everything will be eventually forgotten. There are certain instances, ⁓ certain memories that are there with you forever. Let's say you getting married. ⁓ I hope you don't forget that and have an exponential decay on that. ⁓

Sneha Mehra (00:49:21)  
or you have a baby please don't forget your baby. ⁓ So whatever the critical event is, this is what you would write in your LLM prom to figure out hey this is critical this is not like how your brain processes it to be important similarly your LLM will process it to be important. ⁓ thesis I have I have not verified it but how I don't know if somebody has written a paper on it but my thesis is there will be a need

for a PageRank like algorithm ⁓ in memory. For example, if I keep referencing a memory, its weight should increase, right? And the things that are linked to that are more important versus others or some variation of this. Like what Google did with web searches with PageRank algorithm, something can be done with memory. Because if it is, ⁓ if I'm talking about, let's say my baby on

chat with someone because that ⁓ is ⁓ related to marriage my marriage should bump up so things that this is linked to gets higher attention as and when it works as a kind of page rankish stuff i've not checked i just got the thought as like when i was preparing for this session i got that hey this could go into this page rankish direction but again if you'd want to pick that up pick that up right to explore so ⁓ one of the easiest way to do is

Something very simple is like access multiple times kind of like your LFU or like anti LFU stuff. If something is access multiple times of more importance and it's least access time below certain threshold in a time window, you remove it. Or if someone is explicitly, please remember this you agent, this is very important that you explicitly remember it forever. Okay. Next up is when do we write memories? We kind of saw after every five steps,

you write memory to an external storage or you try to persist it somewhere or try to compact or try to summarize whatever. But when we look at write strategies, let's look at few of them, like not just strategies, but when do you choose to write a memory? When do you register a memory? First is we decide what to persist. Not every agent decision, not every single conversation that you have is important because the cost of noisy memory is very high. Imagine we saw

Sneha Mehra (00:51:45)  
how we can just do semantic lookup to find stuff. Imagine you have lots of random conversation with that same word repeated again and again, it's semantically similar, just bloats of your context ⁓ and reduces your precision because what you got, most of that was junk, ⁓ only one of the few point was important. So if you're just bluntly saving every single chat in your external storage or your memory, problem.

your SNR ratio which is signal to noise ratio degrades. Your ⁓ retrieval becomes irrelevant and very likely contradictory because you just said some stuff in a contradiction fashion and it somehow was ranked higher. Your output goes berserk. So good rule of thumb again not always a golden rule a good rule of thumb is you try to or you should persist decisions that would change the behavior of the agent or future agent execution.

In the case of a personal assistant, could be user preferences. Hey, ⁓ now I like this more. This shows that I used to like that more, but now I like this more. So this is a signal for me to replace the edge that I just created. ⁓ Arpit likes pizza. ⁓ And now let's say next chart, say, ⁓ but now I like burger more. ⁓ This is an indication that I need to replace. ⁓ Now imagine why by the way, we'll take a look at example of this next, which is evolving memory of user preferences. ⁓ Where what we do?

is we try to ⁓ not always insert or append, sometimes we have to upsert. That depends on what we are trying to do, what we ⁓ gathered as an information out of it. That is the older state of this graph and edge or this node and edge relationship irrelevant. Should I replace or should I append? Let's say I say I like pizza and then you say I like burger. This is a case of append.

But if I say I like pizza or let's say you say I used to like pizza but now I like burger. ⁓ This is a case of replace. ⁓ Again imagine these two sentences happened in two different chats for the same user in the span of one week or two week. This is where it comes in. Okay. So what to persist, how to persist again it depends on your use case. You need to understand what you're building, ⁓ what information is critical. You extract that information and persist it. But remember don't just blindly persist everything. ⁓ Okay. Now.

Sneha Mehra (00:54:11)  
Few things to keep in mind around right canonical keys descriptive let's say user underscore user ID underscore preference underscore food should be if you're doing it in a key value store. ⁓ Be very verbose because your agents understand you just don't like PREF because it's easier for your agent to output preference versus PREF because it might get confused with something else. ⁓

Then you tag your memory that just got created at your writing with the task ID, agent ID or whatever. You need that observable system that if something goes corrupt, why did it go corrupt? If something got replaced, why did it got replaced? Which is the latest state that is getting reflected in your memory for the task, for that agent. ⁓ Just remember this thing in mind when you're building the system. Just don't blindly take things on its face value. We'll take a concrete example of this. The concrete example of this is Evolving Edge. This is what most

memory startups that you see in the world are doing. They're just blindly making call to LLM to figure something out and say, hey, tell me what's important. So I'll first show you the demo and then we take it from there. Here I have to type something. Okay. So chat. So this is literally me chatting with the agent and now you will see evolving memories over here. So I copy paste this. I am Arpeth, I'm engineer for Bangalore. So it went.

⁓ It could not find anything. It should have found. am an engineer in Bangalore. Nonsense. Let me run it again. Let me run it again. It could not. ⁓

Sneha Mehra (00:55:54)  
This is 2.0 by 2.5. 2.0 is not deprecated. No, it's unavailable. That's why it failed. ⁓ And fine, this happens. ⁓ by the way, remember this. Your models can get deprecated. So have those evals in place that when you get a notification from Google or whichever model you're using, you change your code. Otherwise this happens. You get mini heart attacks in middle of a session. Huh, ⁓ good that this wasn't production. Okay, I do this. I copy paste this.

It goes calculates memory actions. ⁓ what I'm doing is I'm instructing LLM to generate actions for me. ⁓ It generated action key extensions are pith engineer Bangalore subject is Arpit works as engineer Arpit lives in Bangalore. Right now what I ask you to do I ask you to output operations like absurd absurd. Because if I tomorrow say I live in Gurgaon this is a case for a

date not append. ⁓ So keep this thing in mind. Again, I'll show you the prompt how I'm generating it because that is important one. Now I'm writing actually I just moved to Bangalore from here it would extract and say Upsert Arpit lives in Bangalore. So this did as is Arpit works as engineer because this was earlier as well. I'm I'm just I'm just color coding it so it becomes easy visually and Arpit lives in Bangalore little it didn't Upsert. ⁓ Next up.

I say I am now a principal engineer 2 working at Razorpay. It says it extracts and says Arpit works as. Now here it was engineered. ⁓ it evolved and said Arpit works as principal engineer 2 at Razorpay. It did not do append. It realized that ⁓ this is an absurd case because it is works as. Again here we are relying on intelligence of LLM to output the operation. ⁓ LLM can go wrong.

These are evals command which we will discuss next week Saturday. Either way on how to write evals. And then lives in Bangalore stays as is the yellow one stays as is. ⁓ Okay. Next up is I primarily code in Python for AI. Now you look at this, this is kind of graph that is doing. Now it says add. If you observe it did add for Arpit codes in Python because I could write Arpit codes in Java also. In this case, it did not.

Sneha Mehra (00:58:17)  
It did not ⁓ do an absurd. It's not asking me to replace. It's adding it because it it intelligently identified. This is an additional use case because it's codes in because this user can code in multiple languages. Right. So the new line that got added. Right. Next, I'm starting to use Rust for performance critical goals. A performance critical parts. ⁓ Ideally, it should append. Yes. ⁓ Yes. ⁓ Bro.

Because of non deterministic you get so joy that it worked fine. But this is what also happens in production. Here you see codes in Python codes in Rust. Right because it identified it's an ad not a upset. Okay next up I'm meeting Sam at Sardo Sofia today. Okay so meeting with Sam meeting at Sardo Sofia meeting date today. ⁓ It extracted and added.

The meeting with sam has moved to next day. Now imagine next day I write this meeting with sam has moved to next day Tuesday. Now meeting date change this remains as is meeting date change to next Tuesday.

⁓ Here it identified it to be absurd and that's the important part and because meeting date is something that could change now I change same can't make it I'm meeting Sarah instead now meeting with would update look at the operations absurd or pith meeting with Sarah everything else remain as is Now what I just demonstrated is most memory AI startups This is what they are doing for every conversation. Of course, they are not doing it for every line of command or every

conversion well ⁓ some of the heavily funded ones were actually doing that and then people call them out and then blah blah blah will be happened but you see how easy it gets to build your ⁓ It's a toy layer of course, but here there is a lot of ⁓ opportunity for you to save cost by using a weaker model or a cheaper model to expect that information and do it in batches for a set of common rather than for every single conversation Now you decide how you want to do you can club it with the standard

Sneha Mehra (01:00:26)  
Named entity recognition etc etc to make your life easier ⁓ and you bring more CPU bound versus more LLM bound. ⁓ This is an example of working memory. Let's go through prompt. The prompt is very easy. ⁓ Here you are a memory extraction engine. Your task is to analyze user message and derive specific memory actions like add, ups or delete. This is what I asked it to

that because I wanted operations to be outputted so that I can make changes to my graph, the graph that I'm doing is all in memory. ⁓ There's no graph data as it appears, I just built a graph in memory that is storing, it's actually a list, not even a graph. It just literally does a linear scan, finds the entry and then it replaces or adds, that's what it does. ⁓ Critical rules, do not try to fix or reconcile with current memory, simply extract actions based on new user messages. Action type, add, when to use it, ⁓ upset, when to use it, so that we hear.

lives in, works as, current state, likes, has skills. So you are being very prescriptive to agent on what constitutes ad, what constitutes absurd. You'll be like, but Arvind, I am building general purpose agent. ⁓ Don't try to build general purpose agent. You will most likely be building agent for a specific use case. For that use case, use relevant examples and let it act as a prescription for your agent to see what is important and like when it should call ad and when it should call absurd. Then you have delete.

Then use crisp uppercase relations. Now what I ask it to do ⁓ is when I first build this application, it did not have this thing. So sometimes it write live space in, sometimes it did work space as, sometimes it extracted as this word from my text and put it there. So then this instruction was important because now I'm giving it a uniform structure on how a relation would look like so that duplicate detection becomes easy for me. Then respond only with valid json object. This is how I want. And then when I get it, I update my action.

Now this is where I get memory actions and then I act on it. Memory actions and apply actions. If you look at apply actions, I'm literally look at this changes ⁓ memory. Look at this memory. It is literally a list. It's not even a graph. It's literally a list because it's a toy example. ⁓ Ideally, could what you could do is you could write it into a graph database. When you extract memory, you could write it to a graph database and then use a cyber query or whatever to retrieve it if you want to. But it was a toy example. I just use list to do it.

Sneha Mehra (01:02:49)  
dead simple. ⁓ I just wanted to keep things simple for demo purposes. ⁓ But this is how you build your evolving memory. For example, now you can make it as sophisticated as per your use case. ⁓ Any questions till this stage?

Sneha Mehra (01:03:07)  
Good luck.

⁓ First is that, this is, I know this is a toy example, ⁓ but normally we would not, in our prompt, we will not mention that written in JSON, right? We will institute by identic. ⁓ Of course, of course, of of course. You use identic model, ⁓ pass it as metadata information in your tool call or general tool call into your prompt where you define it. Let this how my output schema should look like and it would respect that. Yes.

I just wrote it because it an easy choice for me to write. Right, right. I got ⁓ it. Another thing is that ⁓ when you change the person you are meeting with ⁓ from Sam to Sara, ⁓ it still kept that ⁓ meeting place as Saudi Sofia. Yes. Because that's what I asked. Somehow wrong, right? ⁓ No, no. just asked. Sam can't make it. I'm meeting Sara instead. I never said where I'm meeting. But, ⁓ okay. I'm meeting Sara instead. Must be at the same place. Right?

could be at the same place. Not must be. Must be. Could be at the same place. So now if the if now imagine if the meeting gets scheduled Google calendar invite in can flood in and can replace the event. Let's talk. ⁓ No. ⁓ Otherwise we can we have like more relations like with whom you are meeting. Right now you have ⁓ subject relation and object. ⁓ Can we have more columns.

No, typically you do a three-tuple format, subject-object verb. Your graph data is also that, node-node connected with the relation. Okay. Right? It's not a SQL database. Correct? What you're doing is a graph relationship, subject-object verb kind of stuff. Okay? Okay. Right? Now here, now I am also meeting, let's say Prati ⁓ at McDonald's.

Sneha Mehra (01:05:06)  
⁓ Next wait a stay. don't know it should do good things here. It should add rather than upset it up started ⁓ it upsetted it removes are all together. So that needs to be added when I say also it needs to do this now. See it's very susceptible very error prone ⁓ now imagine how complex your prompt will become. ⁓ No do not. Remove meeting with Sara.

Let's see what it does.

Sneha Mehra (01:05:41)  
We did not do anything. ⁓ Hey, did not do anything. ⁓ User said something and it did not even find anything. ⁓ Let's do it.

It also tells that these things are not silver bullet. ⁓ I know, ⁓ I know. I'm just trying. ⁓ It's a toy, right? It's a good toy to play with. Why not? Again, this shows and this also shows importance of evals. ⁓ Ideally, it should not have replaced. ⁓ But I think it would have been in the prompt meeting at or something. No, I don't think it's prompt.

But ⁓ add absurd is for the fact that we should have only one value at a time lives in works as current state. If the user mentions new value for this system will replace the old one. I never mentioned it, but it's not a as you said, it's not a silver. ⁓ It's all an actual thank you. Thank you. Good Pratik. ⁓ What change? Tell me. I haven't changed. ⁓ I ⁓ wanted to understand like so based on this toy.

I could have any number of relations created, right? So, but ideally in production, you would want to have a limited corpus of relations that you want to because otherwise every statement can lead to a relation that is sort of adapted. Yes. So best way to implement in pyDentic enum data type, define all the relations that you have so that it respects that schema. And then LLM should also get that schema so that it knows about that it is a thing. that, ⁓

I think one other aspect I was thinking of is if you also wanted to link to the actual memories that you had it somewhere, but ⁓ the output didn't show it. You are linking it to the actual memories also, right? ⁓ Sure. It had it here. Cell.memory. ⁓ Here. then I get changes. ⁓ yeah. Here it's a cell.memory is what I have. ⁓ And linking. ⁓

Sneha Mehra (01:07:43)  
Linking as in to the text. ⁓ Yeah, ⁓ so the actual text that produced it. ⁓ For debuggability on what is producing what memory produced. ⁓ yeah, yeah, yeah, yeah, yeah, yeah, So that was an explanation not in the code. ⁓ That is an explanation. ⁓ And again, this shows why it's important because now you need to know why that bug happened. Exactly. So I was thinking if the last

⁓ There were three columns right? The last column could also contain ⁓ like object. Yeah, it's an object. it's chat ID. The chat conversation ID from which it ⁓ came. Thanks Prateek. Akash, good.

Yeah, two questions. So can we use something like spaCy to ⁓ do the NER and save some tokens? ⁓ What's the assumption if you're using spaCy to do NER? That it's a legit English text. Right? And LLMs are smarter. But if you know that you're getting, imagine you're doing with news articles. ⁓ You could rely on spaCy.

But if you're talking to a human and human agent extraction, they will write in different languages in the same chatting conversation, right? Then it becomes flicky. As said, you could always fall back to or you could, you should always prefer classic ML, classic DL, classic NLP strategies to do things. If you know that input is, it's well formatted, it's legit, it's what those tools expected to do or expected to have. But

If you're dealing with a lot of uncertainties, ⁓ imagine a customer support agent, ⁓ then you typically would want to rely LLM more. Right. One more question. ⁓ So this very much looks like Mempalist. ⁓ I think it's an open source memory tool. ⁓ OK. Yes, very similar to that. But again, now you see where most of them do this but with higher efficiency, like with respect to.

Sneha Mehra (01:09:49)  
⁓ token consumption, multiple messages. So they typically do five or 10 conversion messages and then extract one out of it. And then they have a filter or with respect to number of tokens so that they don't like they have striked a balance video because they have a system prompt ⁓ and then they provide the input and this stays within the limited budget because they have their evals which says that this is the optimal number of tokens at which it would run fairly well. ⁓ And then they decide as per that. Right, cool. Thank you. Yeah, Abhishek, good.

Yeah, but I know it is a toy example, but so the thing that we are doing here, ⁓ so actually, isn't it too costly? ⁓ yes. ⁓ Yes. And companies called out people who are doing more deep research on building memory layer. They did call out other startups that were doing LLM calls on every invocation, like every chat message. Right. So now think of it. This

gives you your playground to play with. And now you can go and for example, as Akash was also mentioning, if I can assume that my input is like pure English text, well formatted like news article, I could just rely on my spaCy or whatever NER library like Stanford NER library or spaCy to exit name ready recognition. Because double extraction from text is a solved problem for good English text. ⁓ It's already a solved problem.

Right? Pronoun disambiguation is already a solved problem. ⁓ It's just that LLMs are just like people are just using it because they can. The underlying expectation that why people are using it is because it made their code simpler. Number one, more accurate. Number two, and token cost was consistently reducing. ⁓ Once it hits the floor, it cannot go below this and token efficiency becomes a thing.

Then you would start seeing a traditional MLDL algorithms coming into fashion again. Got it. Thank you Suryansh. Good.

Sneha Mehra (01:11:57)  
Yeah. I mean, I just wanted to see like how in this example you are doing an absurd. it like the first world match? ⁓ First world match. Yeah. I think also explain why, ⁓ why it would, ⁓ yeah. Replace. ⁓ Yeah. So, yeah. I mean, in, in, in, in ⁓ actual, ⁓ production ice cases, there would be a different, I think, ⁓ method to absurd or something.

Yeah. So that's why you would use a graph database to do it. Yeah. Yeah. Okay. ⁓ And another thing is that this ⁓ update is like whatever we are adding to this graph database is also a tool call in itself, right? ⁓ Or so here I'm not adding graph database, but you don't need a tool call because it's imagine in real practical ⁓ solution. This would come as a stream of events through Kafka to your system. You're not doing it in an agent loop. Okay. ⁓

The AI response is basically subject verb object that way. ⁓ No, no, no, no. So you would not do it in agentic loop. What you would do because here I did it synchronously. ⁓ But in production system, imagine you're building a customer support agent. ⁓ Correct? ⁓ Or let's say your personal assistant agent. ⁓ Would you want to the subject of the verb synchronously and update in your database? Not really. ⁓ Because at the end this conversation is also going into a database.

You can have a CDC pipeline, consumer consumes, extracts and updates asynchronously. You don't need a synchronous update for that. Right? So now you can use a separate model to do this. You don't need to do it in your in context window that you have. Okay, that's why it won't be a tool call as such. But the out ⁓ LLM is in the format of subject verb object. ⁓ So that's what I did over here. But ideally, you're using a cyber query, you can ask it to a cyber query and just run it. You shouldn't do it.

You should make it output an RDF tuple. That's a better way to do it. Like subject object verb is a better output format because LLM is more structured into outputting a subject object verb because it's easy to extract that information. ⁓ Right. Thank you. was helpful. ⁓ Thank you. ⁓ Okay. Sajjal Pratik and all I'll take questions sometime. I just don't know how long it would take for me to cover the multi-agent but as I'm just being cautious. ⁓ But again, we have lots of data. Okay.

Sneha Mehra (01:14:23)  
Thank you. We'll move to the next part and now we'll discuss multi-agent. Multi-agent, relatively very simple. The idea is that one agent could not do things. ⁓ We are hoping multiple agents would do things. Okay. We kind of touched upon it yesterday where ⁓ we just kind of like your ⁓ plan and execute thing where ⁓ we discussed that, there could be things that I could do in parallel. Right? That is what the multi-agent use cases that

is one doing parallel. ⁓ So multi agent is you having multiple ⁓ single agents running. ⁓ Because if you're just doing everything in your one agent, you see how complex your agentic loop gets. Treat your multi agent as a microservices architecture. That's a loose analogy that you can apply. Would you have a monolith? Would you have microservices? Each service doing its own thing really well. Each service can use its own database. Each service can use its own architecture.

Here is your agentic loop, database is your model. ⁓ That is your ⁓ multi agent flow. ⁓ Which is where it is interesting how that handoff happens that leads to a problem which we will discuss in some time. ⁓ So now think of multi agent is that you have a single agent, has a large context, treat monolith, start drawing analysis to monolith and the tools good enough for most use case. But if your task can be split into multiple sub tasks, if your multiple sub tasks can execute in parallel.

If by doing those things in parallel and those are like pretty independent tasks you can speed things up. You want specialized model to solve this one problem. example, planning by Opus, execution by Sonnet. Right? Or let's say Gemini is great with English text. I would use Gemini and let's say I might use some model for Hindi output or whatever multilingual output that you would want to do.

That is how, or for example, if for a task I have longer context, I would use a particular model and for other tasks I would use a different model. I would want to do that. That is how you decide you'd want to do multi-agent. On other case, you can also think of it as single place of like single piece of responsibility. Like, hey, this agent does this work, only this work, this agent does this work. For example, this agent just takes care of refund part. This agent gives me everything about shipping and takes care of shipping. This agent does planning.

Sneha Mehra (01:16:45)  
This agent writes code in Go. This agent writes code in Java. You could do this. This does a code review in Go. This does a code review in Java. Imagine you have a monorepo where changes are both in Go and Java. You would invoke Go code review agent to do this and Java code review agent to review Java code. So it's all about splitting of responsibilities. So very loose analogy for that is monolith versus microservices. ⁓ Okay, let's look at first pattern. For this we'll look at prototype as well.

Orchestrator specialist pattern is what you would very likely be using to ⁓ build multiagents. It's not the only one, but very likely you will be using orchestrator specialist. Idea is very simple. The idea is I would want to split my task. I have an orchestrator, which is orchestrator. This is your literal your LLM task that you are given. Orchestrator splits kind of step and kind of plan and execute. It splits into multiple sub tasks.

hands it off to specific agents, waits for them to respond. All of this agent does its thing, responds to orchestrator, orchestrator keeps adding it to its context window. Not every single message, but output from the first agent. ⁓ Imagine it ⁓ wants to process refund ⁓ or something it wants to process refund, it hands it off to refund agent, which runs its own loop to verify this, that and outputs. Yes.

that refund was already given for this order and this one final output because this is an agentic loop that is running for this sub task. It gathers that information just that one information final information and gives it over here. It's not a tool call. Tool call is external call to a system. This is literally an agentic loop that it runs multi-step reasoning on its end and generated the final output. This output goes to orchestrator

Orcasator keeps track of everything that it has and decides what to do next, what to do next. Orcasator can then make a few tool calls to get subsequent information. Even a subagent can make few tool calls to get subsequent execution, like subsequent information. Right? Okay. Now, let's take a look at practical example for this one. Practical example, two usages and then one prototype. Two usages, imagine you're doing a customer ticket triage. So imagine your ticket can be anything.

Sneha Mehra (01:19:05)  
This is your main agent that received a ticket. That agent needs to see until this ticket is resolved, resolve. ⁓ That is a loop that you write that goes and figure out what needs to be done, raising token output and decides, hey, this is where I could want to, ⁓ this is a task that I think is payment related or rather this ticket is payment related. Let me hand it off to payment agent. Now, how does the handoff happen? Handoff is nothing but,

you decided I want to hand it off to this one. So you can create a sub process, ⁓ a Python sub process and make it run the other function which starts the sub process and it executes that part. It could be a separate file, it could be a separate thread, whatever. If it go, ⁓ write a call a separate go routine, right? That just calls this as an agentic loop and gathers all that permission, solves it and it gives output. ⁓ And so here it's a sub process or a go routine.

you could also do everything in a single flow. You might not want to do anything in parallel. Given that a ticket that you are receiving ⁓ is, I want to hand it off to payment agent. You're literally just calling the payment agent's code and it's adding that stuff either in this context, actually you should not do that, but or in this, its own context. And it runs, it outputs output. If you are using a separate model, then you would have a separate go routine or something, its own context window, et cetera. That's a better practice. It outputs.

⁓ Similarly, you look at incident triaging, where I might have an agent which is specialized in ⁓ looking at metrics, this agent also just has access to metrics tools. Now another benefit of using multi-agent is tool isolation, which is metric agent should have access to metric tools, it should not have access to log tools. My log agent, which is specializing in analyzing large amount of text to extract the critical information and

narrow down the bug. So I can when an incident happens, I can just fork it off and say metric agent go find relevant metrics for this one. It would run its own agentic loop, find relevant metrics, generate a final output, return it to the orchestrator Python sub process or go rotate whatever right then ⁓ logs go find it out figure it out generate a final output send it source code go find out the relevant pieces of information and give me

Sneha Mehra (01:21:25)  
Now all of this goes into context and this main loop orchestrator is reasoning, reasoning, reasoning, reasoning. And in reasoning it could make tool call, could make hand it off to other subagent. Let's take an example of a deep research agent. How to write one. This is the final prototype for today. Again, we have a system to cover, but prototype wise only three prototype for today. So here I take a deep research example where if I run, it asks some, me just close it, OCD, sorry.

Python main.py. And I would open another window where it runs here ⁓ and copy, copy, copy, paste. I have a watch. ⁓ Okay. So now here should fix this. ⁓

So now here what I am doing is am doing a deep research topic. Research topic is... ⁓ Tau Ceti... Sorry. Let's say... ⁓ Why was Pluto... I made this shit up. Why was Pluto removed from ⁓ planets in solar system? Sorry. Hail Mary effect. I watched it second time today. Okay. So now what it's doing...

It's doing my lead planner agent. is kind of orchestrator. Then I have a specialist agent. One makes web search. Second goes and does ⁓ concept specialization. Then I have a research critic agent. And then I have a lead synthesizer agent. Fancy name. Does nothing fancy. Right? But now here it's split into subclass. Here you saw it was just one. But now you see three Python processes running. ⁓ Once that job is done, you would start seeing this chopping off. So now it goes, it figures out it needs to do this. I will...

Let's look at the prompts as well. I want to show when this ⁓ process is complete, you would see this going away. ⁓ So it goes, ⁓ your lead planner agent, first it created all the subtasks. Then it realized it needs to do concept, ⁓ which agent it should assign to, and it assigns a task to that corresponding agent. So research, all of this is happening in parallel. ⁓ Now all of the task is done. ⁓ These two tasks are done. It goes, does a web search, et cetera, et cetera.

Sneha Mehra (01:23:41)  
The research critic agent is conducting cross report gap analysis and audit completed review. Now it does critical gap analysis, is this one critic agent lead synthesizer. It does critic. It says as a research critic, sorry, review the parallel reports and find gaps in conflict. This is a problem that was sent to that agent. It does this, this, this, this outputs. And then I have everything over here. And this is my preview of the report. Right. Now look at this. Everything is done.

So literally forked of Python subprocesses do things independently. But now how does the handoff and all happen? Let's look at it. ⁓ here. Research subagent task, research plan. So what I'm doing is this like simple pidentic models what I have generated. You are a chief research planner. ⁓ I found this prompt somewhere and asked Gemini to modify it.

Given a research topic, your job is to break it down into 2-3 distinct highly focused subtasks that can be researched in parallel. For each subtask, determine whether it requires web search or self-knowledge. For theoretical concepts, fundamental physics etc etc etc etc etc etc etc etc etc etc etc etc etc are complimentary and do not overlap. ⁓

So I did web search for astronomical discovery challenge, Pluto status. It figured it out and I need for this Pluto discovery in early and the resolution that happened itself contained part. ⁓ Now you could hand it off to agents like how exactly let's do it. Run research task. Here I'm doing you are a web search research specialist. Your job is to do this. So now my system instruction, if you look at this changes. If my task is web search, I'm doing this. Otherwise I'm doing this.

Now what I doing is this runResearchTask I am saying what is a subtask, what is the task number and this research task ⁓ is here where did it go? Here. runResearchTask. So I did plan.subtask when it got a subtask I am calling the same function for each of the subtask. Here I did task.append and here I called asyncIO which created multiple subprocesses kind of and it scattered and gathered and then it passed it to

Sneha Mehra (01:26:05)  
Run critic, run synthesis, generate final report. Here's the final report. From ninth planet to dwarf ⁓ world. ⁓ classic GBT. ⁓ Classic LLM. The definitive account of Pluto's reclassification and the evolution of planetary sciences. And then you get this final output, which is the final research. ⁓ Some things it came from, what do we say, web search. Some things came from self knowledge.

Now you can add more sub digits. The whole idea is to do scatter gather. The magic sauce is this async io.run. It calls this main function ⁓ where it gets task ⁓ and ⁓ gather here. This is where it gathers all the tasks. It's wait for it to complete. Once it gets everything, then it proceeds further. So all the things that I wanted, I just added it over here in this one and my job is done. So this is where imagine this go routine and this is your

WG dot wait or your Java and you have something to wait until all the threads are done doing its thing Kind of that. They are simple deep research agent that you can build. You can make it as sophisticated as you like I just kept it short enough for us to look at a demo in the session itself And it's a different report in general. I can ⁓ see different report earlier report was ⁓

Good. Doof.

Research ⁓ report. Yeah, earlier was unveiling the inner workings of LLM. I asked her to do how LLMs work on that topic. But now I asked her to do Pluto. Just to show it just works. ⁓ OK, any questions on this?

Sneha Mehra (01:27:51)  
how the agents are splitting, scatter together. You why threads, processes, OS, important. That's why. ⁓ Aniket, go ahead.

Sneha Mehra (01:28:03)  
Yeah, this is ⁓ basically agents inside the same code base, right? But if let's assume we are having multiple code bases and you're talking about A2A in that case, right? are we going to discuss that as well? No, so A2A protocol I'm not covering, but one of the easier way to implement this is put it into a shared memory, which is let's say your database in which ⁓ one of the worker is pulling the database. If something needs to be done, then it calls up a process.

A2A handoff ⁓ is very ⁓ problematic as of today. It's still not mature enough. Most people would rely on having a shared state because everybody wants it high observability and state management and like checkpoint and resume. ⁓ So simpler way to do this is to just upload it or like update it in the database when the subtasks are created and then someone pulls it and then forks off different thing within the same machine or some other processes listing. So depending on who the

owner is that corresponding machine pulls and forks it off. ⁓ basically via state. Okay. Okay. Thank you. Which is where you see temporal and airflow getting fractured with LLS in place, right? Because now every machine can independently pick that up and run it. Yep. Thank you. Sajan, ahead. So ⁓ can you go to the part where like you're forking off sub agents?

I just wanted to understand ⁓ how does the orchestrator like tell like what exact thing are you parsing for in that response from the orchestrator to fork off? Okay, here my orchestrator plan has subtasks. If you look at this, I asked it to this research plan was provided ⁓ as my pyrantic schema so that I get subtasks. That's doing the heavy lifting. So the subtask splitting is literally a structured output is what I got here.

And that is part of my research subclass, which has a title, a description, and agent type, which is a web search or what. So then it goes and picks it up and executes as different async IO loops. Got it. And I'm guessing the subagent in this case does not have access to the conversation that the orchestrator has had. ⁓ Yes. Yes. So that's why orchestrator should not be. It's like literally, ⁓ if you have given your thing to ⁓

Sneha Mehra (01:30:29)  
⁓ your SD one, you should not look at it. You should just care for the output that you're getting. It ⁓ doesn't peek into it. Once it's done, it outputs. And again, that end-to-end management, if it's different machine, ⁓ via shared storage you can do. If it's within the same machine, you can just use a shared memory. In this case, I just use the shared memory to get the output. ⁓ Got it. Got it. Thank you. you. ⁓ Sumanth, go ahead.

Sneha Mehra (01:30:57)  
⁓ My main question is when to use agents and when ⁓ we are okay with skills itself. ⁓ Agents is when RBAC is needed between agents. We don't need tools to be shared across different ⁓ work. So ⁓ agent is better for kind of RBAC between agents. Otherwise skills in the same agent is okay. ⁓ But now how are you executing skills? That's why you need agent SDKs.

if you're doing a programmatic. ⁓ So here the loop is implicitly done by your agent SDK, which is your cloud agent SDK or anti-gravity SDK. But then they are running the loop on your behalf, you're just passing your skill file to it. ⁓ Sure. ⁓ Thank you. Saurabh, go ahead. ⁓ Yeah, just one question. So here, ⁓ in case of multi-agents, ⁓ the common layer you're saying is the state management are doing on the same machine because you know that

all these are going to run on the same system. Yes, it's a toy example. That's we do it. But as we discussed, right in a distribute setup, ⁓ user central database is a state manager. So then how means every agent needs to know right like where to look for state or like no, no, no, no, no, no, no, no, won't be part of your agent loop. These are regular Python loop that you will write to pull a database to see if is there any task for me, any task for me, any task for me, any task for me or a queue. ⁓

But then it won't be orchestrator this pattern, would say. Let's say if you have a pattern where you're first doing the planning and there's an orchestrator that needs to keep executing agents one by one. ⁓ even your orchestrator loop, your agentic loop is still Python code? ⁓ Yeah. Where it can make a tool call. ⁓ Imagine there is a tool call which says wait for all subtask agents to complete. ⁓ The tool call is looping and checking the database until are all subtasks completed or not. ⁓ Correct?

It's your agent doing, it's not your agent English language thing that you wrote that keep polling until all the subtasks are done. Correct? So what you can do is you call, you provide it this step that plan all the subtasks to be executed, hand it off, make a tool call and wait for all subtasks to be executed. So agent will automatically make tool call. You will have a call which calls that function, which polls the database and that function you are invoking as part of an agent telling you to make a tool call.

Sneha Mehra (01:33:23)  
Rather you could just avoid that LLM taking that call and you can add your own thing. ⁓ Imagine this. I'll show you how. ⁓ So imagine your subtasks are created. ⁓ Subtask. Look together. ⁓ Okay. Imagine here. ⁓ Your subtasks are created. This is where your job is done. Your first research task. ⁓ Sorry. Plan.subtask. Planet came over here. Plan came over here. Here. ⁓ Generate plan. ⁓ Right? ⁓ Now, Generate plan is your orchestrator and loop.

created a sub task, right? If it wants to run a loop, it can run a loop, right? Otherwise it could be just one LLM call and be done with it. Correct? Now here you can write, because you're executing, you can write literally, pull to like, not literally, but literally for a db.call, you pass the query, select star from whatever database to do whatever state where ⁓ agent ID is whatever, right?

and then you have a loop until all subtasks are finished. You are literally making this call to do it. Sleeps 1.0, for example. So then after you spun up all the tasks, then you're waiting for it to come. So here that async.io.gather that did that stuff for you, you could add a for loop. This is like a classic shot polling example. This is basically build your distributed job scheduler. ⁓ That's it. ⁓

That's it. Now you could replace this with temporal. ⁓ Yeah. ⁓ temporal replace this with Airflow DAC. ⁓ Correct? ⁓ entire workflow management because here we knew we are creating a task. ⁓ Like we are creating a plan and a bunch of sub tasks and we have to wait for all of them to finish. ⁓ But you could put your workflow. There's a kind of a workflow. The workflow can become very complex. It needs to be executed in a distributed fashion. Who can execute what? You have different worker pools. ⁓ Each worker pool, imagine it's an agent worker pool.

Yeah, now can we get as sophisticated as you like? ⁓ That's a temporal and DAGs and like a extra DAGs are doing wonders. And then each of those would also need to know right each independent. ⁓ Yes, that's a code you write. ⁓ Yes, yes, because each of those are also Python code in itself. ⁓ correct. ⁓ Where it should persist and there can be interdependencies between those subagents as well. ⁓ Yes. ⁓

Sneha Mehra (01:35:45)  
That is where your DAG comes in, that is the final system that we will discuss in this course where we build a natural language workflow engine. And so that is literally your workflow that you are designing which can go scatter, gather, weight and then fork off multiple things and do so this is like your N8n workflow sort of stuff. Got it. And I did not get every what you said about A2L like because ⁓ It's a protocol if you are not aware Yeah I know know I follow at Google prepare but like in this case where would it be useful? A2L ⁓

Here one agent wants to hand off tasks to another agent within the same machine you do it. I'm not sure if they've added it in a distributed fashion, but I found it to be very buggy at least till this stage. So that's why I'm still rely on my classic old school database scatter gather format for me to operate with. Again, I'm just waiting for that A2A handoff to become mature enough for like everybody to use. That's why you just see you equal announcing one week, everybody talking about it. Then it's like pretty quiet.

Everybody relied because now temporal and air flow are becoming the like temporal and any 10 are becoming the preferred choice for do to do this workflow because it's much easier and much more deterministic and much more reliable. ⁓ This A to A is still iffy at this stage. ⁓ Gaurab, it would be more distributed right kind of or more generalized when you want to create that. That's the problem. You don't need to over generalize it. Your temporal inflow depth works just fine.

So it's like when you, you would always try to create a like, deterministic kind of workflow means you know that what you're doing for a particular task and you code that in temporal rather than making it very generic and figuring things out on the fly. Yes, Perfect. That is what you like. If you know that you have to wait for all sub tasks to complete, why to let it be a tool call? Why can't you just be done with it and like write a while loop like this, right?

That's the same thing. Let your agent decide that it needs to wait or you know it needs to wait to buy waste tokens of your agent. Pratik, add please. Yeah. When you said ⁓ that it needs to be deterministic, if you think about it, you wouldn't let your production workflows be like, I will call one agent which can spin 1000 agents or 100 agents or 10 agents depending on the task. There is no determinism there. You cannot

Sneha Mehra (01:38:08)  
possibly have an SLO against that endpoint anymore. Because you don't know. I was going to add this. The robustness and the reliability of the system is still in your hands. Like you cannot let your agent say, what our agent does is right. Right now, everybody unfortunately is just handing off to Cloud Agent Nest. can expect it to do wonders. But I think the reliability of your production system is in your hands. You are accountable for it. Are you sure what it is doing? That's why.

Use LLM for intelligence, not orchestration. ⁓ If your task is fixed, you exactly know how it needs to execute. ⁓ By the way, we are struggling with the same thing with agent studio at Razorpay. ⁓ We, two months ago, we evaluated that can we have from a natural language, create a deterministic workflow, ⁓ generate code for each one so that our merchants, ⁓ when they run, ⁓ they should not expect one time it's working fine, second time it did not. That's a bad user experience, right? So,

they should be very sure what it's going, ⁓ how that behavior is. ⁓ So you would tend to go towards more deterministic path and you would see in coming months, ⁓ people moving towards more deterministic path and using LLM for not like entire orchestration, no matter how much of loop engineering comes in. ⁓ Your production systems are way ahead, ⁓ are too far away from your production systems. ⁓ Where, ⁓ I said production, production, whatever. ⁓ So your toy systems, I was about to say something.

So your toy system that you're doing, your looping and a full agent orchestration versus what your user is doing. Because your user is not saying, wow, this company is using AI. I'm OK with some non-determinism. ⁓ Definitely. They say, I am paying money. I want output. ⁓ So be very mindful of that. That's what I want to highlight. ⁓ OK. Thank you. Perfect. ⁓ Thank you. And what you can build is the

they can orchestrate a platform, it's a generic. ⁓ Agents have specific roles defined at a higher level, but then you have this workflow where you can connect these agents in whatever fashion you want for your specific use cases. ⁓ And then you can have specific prompts, ⁓ like user prompts that you can define

Sneha Mehra (01:40:30)  
That way, then you can build different multi-agent systems from a platform agent system, essentially. ⁓ Yeah. So what Kevin is saying is it's a deterministic flow. ⁓ let's assume you have a flow where you are calling an agent to identify the language of the user. ⁓ Then you're moving ahead based on that choice of language.

You can then call another agent which ⁓ understands English, another agent who understands something else. But this flow is always deterministic. There will always be one path to be taken for a particular thing that can happen. It will not be that that agent spawns thousand other agents. ⁓

Awesome. Okay. We'll take few more. Rohit, Yeah, we don't use any ⁓ AI in our backend system. So ⁓ most of the use cases are like just like looking at logs or such things and we just use skills. And should we think about using code to do such workflows or are skills enough? If you have like a simple workflow, not like a simple. ⁓

If it's simple and skill is deterministic and you okay willing to spend few extra tokens then skill is good enough. If you are trying to manipulate or edit it and it's easy to describe it in a very natural language and you know that Claude for example Claude is a specific example is good enough for it to figure out just use skill because it makes your GTM faster. Correct? Yeah. But ⁓ if your customer if a user is expecting that full determinism ⁓ and or you are chasing efficiency

So right now all workflows of agent studio are skills. ⁓ But what we are not doing is we are not optimizing on cost at the moment. ⁓ But the moment cost factor kicks in and when we roll it out to production at scale, right now it's still in production but for very small subset of merchants. But the moment it gets rolled out to production for a large scale of merchants, things would change. ⁓ Then we would convert this skill into deterministic workflow.

Sneha Mehra (01:42:37)  
and then run and only make LLM calls when it's essential. Right now our skills is doing orchestration for us. Which says, if I'm building an agent, which is doing, let's say subscription ⁓ recovery, or let's say, let me take an example, ⁓ abandoned cart, which means if someone paid, added to the cart, went to Razorpay, ⁓ added to cart, went to Razorpay, did not complete the payment, we give that user a call, negotiate it. That entire workflow is written in a skill.

that runs periodically that figures out what needs to be called, where it needs to be called, who needs to be called, how it needs to be called, how it needs to be negotiated, ⁓ etc. ⁓ Everything is a scale. ⁓ All that entire scale and because right now we are iterating, ⁓ we are getting more customer requirements, ⁓ our skills are becoming mature. ⁓ Once we know that this is what we want and token efficiency and cost efficiency needs to be taken because once we roll it out to millions of merchants that we have, ⁓ that cost will balloon. We don't want that.

then it gets converted into deterministic. So start simple and evolve. But make sure you know how to ⁓ convert into a deterministic workflow and just use LLM. You can use LLM for intelligence, not denying that. There you need it. But orchestration can be ⁓ waved off to your temporals of the world. Temporals, Bboss, Agno, whichever one you prefer to use. ⁓ Sure. Thank you. ⁓ Suryansh, go ahead. I the topic. Sorry. Go ahead, Suryansh. ⁓

and quickly then ⁓ like I just had a question based on the discussion you guys were having with Saurabh. So like I was looking into like we definitely have these different agents SDK like with pre agents SDK and all of that but in production like whoever's building like you are with and Pratik basically are you ⁓ more ⁓ relying on the response API and you know owning the loop to dispatch and handling state yourself or are you guys using agents SDK as of now ⁓ in production?

I am full agent SDK right now. So again, because it's all GTM, it's not operating at scale at the moment for us. But once we hit that limit, right? Once we scale, once we scale, things will change. Right? So we are still figuring it's like MVP phase, right? ⁓ You're using agent SDK where your iteration on skills is easier. It's just English, English, English, English, right? You rely on Claude's intelligence to do orchestration for you. ⁓

Sneha Mehra (01:44:59)  
But the moment you go for scale, this will become very costly because imagine if you know you have to wait for five minutes, why you want an LLM call to decide that I need to make a tool call that makes me wait for five minutes. ⁓ I can decide. I know my book is going to wait for five minutes at that point. Correct. Got it. Got it. Sure. ⁓ These agents are also like compact, like can you, can you build temporal workflows with these agents SDK? They're smart enough to do it by the way. can have like temporal, ⁓

like way to generate temporal code out of it because it is just Python code at the end of the day. ⁓ You can just add a decorator, would generate a temporal code for you. You can have a human in the loop that validates it once, once you, and again, that's the production issue. But that's the last ⁓ design decision I took last week ⁓ around how do we do ⁓ this natural language workflow split and making it into deterministic flow. That's literally my last Friday discussion with the team. I did a prototype with Agno. Agno is a framework that

makes this mix like that creates an abstraction to build agentic workflows. So AG, I know in case anybody wants to take a look at it. ⁓ Aniket, you've been waiting for last year, I'll pull you in and then we'll pause for this one. Go ahead, Aniket. So I mean, continuing the discussion that you were having with Sourav, right? I mean, there are cases where you don't need a deterministic workflow and that's where, for example, we're evaluating a middle of agent framework, agent fabric.

where you have an agent card, multiple agents are registered as an agent card. And then on the fly decides which agent to call. So it's not a deterministic ⁓ workflow, but you have a catalog of agents where it gets decided which one to call at what point of time. That's where they ⁓ are using ATO to do that, but that's where I was trying to understand. mean, ⁓ deterministic workflow that is always needed, that could be dynamic ⁓ calling that might be needed in some cases. Or most of the cases, because now...

It is still deterministic. As I gave the example, identify language based on the language you call different agents. It's the same thing what you're doing. So it is still deterministic because the number of agents that you can call is fixed. also in a particular request, only one of them will be called. It's not that I will reach a decision point and then spawn hundreds of things. It's always I will reach a particular point and choose one of the possible ⁓ options.

Sneha Mehra (01:47:22)  
So my agent, my workflow will only behave in one of the few ways depending on the number of agents I have available. ⁓ Okay. Probably I will continue that discussion once we are done with this. ⁓ Yeah. ⁓ Next up is critic refiner. Now, no more prototypes. We directly go into system, but I have three more topics to cover. ⁓ So first is critic, like how we saw orchestrator specialist, ⁓ similar to that, there is one more part of it is critic refiner.

fancy word for you to just revalidate and critique what has been outputted and refine if required. ⁓ example, where the whole idea is whatever is the output of the first agent you have a critic in it which is critiquing the output that hey for example the article that this agent done it now this is not just one LLM call that is critiquing this full agentic loop the final output which is imagine

the deep research agent, ⁓ deep research report is not in the format that we write. ⁓ So this is a critic. ⁓ So not critic. This is a main agent that wrote the article. ⁓ critic, critic writes critic critiques the article. ⁓ Why did I write this critic?

Critics article and refiner then refines it. Because now imagine your refiner in case of article is someone who knows how to write the markdown file or how to structure or how to format the markdown or this is simple formatting. This could be another which is taking a look at into factual consistency of it because a critical is finding the flaws and refiner is refining basis the findings of the critical to do it. So refiner is just a doer, critical is critiquing like hey is this factually correct?

Now this critic can be its own loop, it does web search to find factual correctness in that. ⁓ Similarly, SQL query optimizer, critic can find slow queries and then refiner can optimize it. Again, the lines are blurry. ⁓ can let say, my critical is also refining the queries and then your refiner might just go and like validate it. ⁓ Again, it's all English at the end of the day. The idea is to have whatever is the output of an agent, if it is sensitive enough, ⁓ critical enough.

Sneha Mehra (01:49:39)  
you may want to make it take a second look. That's the whole idea. ⁓ Right? Okay. This is a critical final look, very similar to what we discussed in first week, where we had this thing around re-evaluating ⁓ what is our, when we were doing this fact checking system, where we relied on self-knowledge. Now imagine instead of self-knowledge, you make tool calls and do web search and check if it's all correct. Now that is its loop.

That's why I not have the prototype for this. So you add this where your correctness is more important than speed. That's the most important part because now ⁓ once the output is generated, having this additional loop is costly. It's time consuming, it's token consuming. So where your correctness is of utmost importance, you would want to spend more tokens, do more iterations until it's refined to the best that you would want as an output from this agent. ⁓ Then you have another which is mixture of agents.

It's very similar to what we discussed what the mixture of what something we discussed where you make the same task you give to multiple agent ⁓ and then you combine the result and give the best output. So imagine you give deep research agent like you have deep research agent using Kimi deep research agent using Claude the deep research agent using Gemini and then you have output output output and then you have a synthesizer phase

which takes all these three reports and creates the best version of it. On top of it, now you can have a critic refiner if you want. So it's not this or this or this. It could be combination of it depending on what you are building. So there is no one way to do it. ⁓ Remember, ⁓ this is not just one NLM call that we talking. This full agentic look that is doing a deep research. It's not just one. It's full agentic look that doing. Second agent doing the same task. Third agent doing the same task. And you either picking the best of them.

or you combine all of them and generate a final synthesized report. This is a mixture of agents. Again you see, it's only done by people having lot of money. If you don't have money, are okay with the output, good enough. ⁓ Which is where eval comes at play. Next week, first thing we discuss that. Then the interesting one. This is something that ⁓ I have suffered ⁓ very poorly. ⁓ Which is deadlocking agents. Yeah, OS chapter number 3 again.

Sneha Mehra (01:51:56)  
Deadlocks. now imagine this you have multiple agents ⁓ and your agent A produces an output that goes to agent B and agent B decides that whatever I am doing it needs to go to agent A. Same thing. ⁓ And it's stuck. It keeps sending huge huge huge huge and nothing comes out of it. I'll give a concrete example of it. And how to fix of course same OS concepts comes in handy. Now imagine you have two agents. One is refund agent, one is shipping agent. ⁓

customer is asking for refund on a damaged item and there is a text line written in the agent clue prompt for refund agent which says if the item was damaged escalate to shipping agent for fault confirmation before processing

And on shipping agent you wrote if refund is requested, confirm with the refund agent that the refund is approved. Because imagine the engineer in the team decided if the refund is not approved, why do I even need to check? As a quick short circuit evaluation. And that would lead to what? So if refund is requested, confirm with the refund agent that refund is approved before logging for the ⁓ shipping fault. Now because

These two agents are owned by two different teams. And if someone wrote this and you wrote this, so what would happen? It would go over here. It would hand it off here. It would go here. It would hand it off over here. So that's why you have to be very mindful when your agents are distributed. Super mindful. When your ⁓ agents ⁓ are ⁓ owned by, let me be specific, owned by different.

Sneha Mehra (01:54:09)  
or cancel it or return something right or you can decide that if it's going ⁓ more than two times explicitly or you say hey if you cannot decide then just assume that you have to give a refund right or you have a supervisor agent which takes a look hey what are you guys doing now imagine what is the supervisor agent ⁓ running on the input to the supervisor agent would be imagine this agent is exactly solving one ticket

one ticket that is raised in your system, which means it would have a database entry. The agent delegation that it did, it would be logged somewhere in a table. So imagine there is a supervisor agent that is running, which is polling the database, making a simple LLM coin and say, do you think there's a possibility of a deadlock? As simple as that. Or you can do it deterministically if you detect like cycle in that, right?

You can do it ⁓ deterministically with LLM but the idea is that do you see a risk of running an agent? So this is literally a simple process, simplest example, which pulls the DB, queries, checks for deadlock, and be done. This is what it short-circuiting the entire ticket resolution agent. Your life becomes simple. And that is what we should do. Do not unnecessarily make things complex. Like that's why these are the things where

These are the common production pitfalls that people run into. just be mindful. Don't take everything with a pinch of salt when you're dealing with LLM because it's fully non-deterministic and worst case it's English. So it's subjective to interpretation. So hence your prompts needs to be super unambiguous. ⁓ Perfect. We have one thing to discuss, which is system. We'll take a break and then we come back and after system we'll discuss again. Not very huge, but this has some interesting nuances that I want to touch upon on building incident auto remediation system. Right?

So it's 9.56, we take break for 10 minutes, sorry, eight minutes, we come back at 10.04 ⁓ and we resume this discussion, right? With this system and then we take questions all together in one shot. See you folks in eight minutes at 10.04.

Sneha Mehra (02:04:49)  
Awesome. It's 10-0-4. And now we start with the system where we design an incident auto remediation system. The idea is when you get an outage, now a lot of companies are actually trying to build this, right? So when you get an outage, by the time the on-call engineer comes, you should have the first set of triages ready so that your life becomes simpler, right? That your on-call life becomes simpler and your

your incident triaging system should also recommend commands to run, et cetera, et cetera. Now, ⁓ over time, the companies, first it will all be the human who will be executing it. But once you have high confidence, you would see ⁓ agents automatically taking actions for high confidence triage that has come out. Now, what we'll do is we are trying to build a system. So we'll look at it from the lens of system design. ⁓ And the agent loop, we saw the drill, how it happens.

So if you look at it, what we always covered in this course is the fundamental patterns. Now we're just using them together to solve a problem. So essentially now you'll see how this is all just a regular classic LLD, HLD or system design problem that we have all solved always, but with just sprinkling of AI. So we'll take concrete numbers. What we'll say is we have 5,000 alerts per day is a steady state. ⁓ We have to do time to first action.

P99 will be 90 seconds for a P1 result or P1 incident, five minutes for P2 incident, 15 minutes for P3 incident. Imagine what you're getting from the systems which is alerting and sending you ⁓ alerts like say Zen duty, Pager duty, incident IO, et cetera. They all come to you via webhook that is becomes an ingestion. ⁓ And from there we proceed ⁓ and put everything out. So idea is to first,

focus on your triage report on steps that uncle can take to output this. Now here you will start seeing like how system doesn't comes at play and things that we have to worry about that how robust the system is because if ⁓ after AI does the triaging if you have to take subsequent actions that should be very apparent. Right? So here these kinds of systems what I felt

Sneha Mehra (02:07:13)  
improves your overall empathy in the system like how you design the system. ⁓ One simple example that I give which is kind of a spoiler is to imagine linking all the critical references in your triage report so that user does not have to go through internal documentation again and search for it. Make it everything that you're giving as a triage report has everything in it so that you don't have to look for pieces of information elsewhere.

Plus it's not just that but also ⁓ this act as this citations acts as the place from which this output or this agent got its information. So in case it refers to something which is outdated you will flag it and that becomes your feedback loop into your system. I'm just giving few spoilers and few things to worry about and why these kind of systems it's about how you are structuring the information.

Agentic loop that reasoning, reasoning, et cetera, et cetera, like this classic rag and reasoning system that we are doing. But around that, what we build is what we are discussing. ⁓ Okay. So we'll start with the first thing. I'll start with functional and non-functional requirements. I'll quickly go through functional because non-functional onwards is interesting. ⁓ functional requirement, is system receives the alert. What each alert contains is important.

Now what we expecting the system like Zen duty, Pager duty, everything else giving you in that alert which is more importantly it should give you priority of the incident. Because when you configure an alert you know what priority it is if you can, if not every alert becomes equal. Second it gives you incident lifecycle like what does an incident lifecycle or sorry it not give you, you have to decide on incident lifecycle because it's very particular to an organization.

where if I am seeing this incident is it snoozed, it retried etc etc or like it reappeared etc etc and it is in fixing, ⁓ in triage, AI triage, ⁓ human approval needed etc etc. So you have to design that state machine around it on what your incident life cycle typically looks like. So for example you may look at this system or any other system who has built it and say hey this is the exact same flow I want to copy but you should not do.

Sneha Mehra (02:09:31)  
You want a system like this to integrate in your existing workflow. You cannot change your workflow because an AI solution exists. Because if your company, your team is used to operating in a certain way during an incident, it cannot change overnight. So think of it as augmentation, not as a replacement. And hence most of the AI system augment ⁓ not replace.

So most of the AI systems that we see, people try to replace their existing scheme of things and say, hey, this is a better way to do it. I'll replace everything. ⁓ But the thing is, then it never gets solved. Then fulfills like a forced adoption. ⁓ Right? ⁓ Okay. ⁓ Another thing to worry about is remediation action that we're doing. That is the most important stuff that we have to do. And verification. ⁓ This is the ultimate thing. ⁓ That's where your critical refiner loop, ⁓ your fact checking stuff, all of that would come in handy. ⁓ That more or less critical refiner. ⁓ That whatever I'm suggesting,

is true to its needs. Now let's get into non-functional requirements. So in terms of non-functional requirements, I'll ask a concrete question. Imagine I'm getting keep your calculator ⁓ ready. If I have 500 alerts per day, is my steady state, what is the number of alerts I would get per second or per minute?

Let's start with that and then I'll ask a lot of probing questions on that. Whoever wants to chip in, their hands, calculate this and we'll start the brainstorming around this.

Sneha Mehra (02:11:06)  
5000 alerts per day steady state. What is the peak load? ⁓ What is the average load? How you want to deal with that Pratik? ⁓ Yeah, ⁓ so I think my I did calculation by hand. So I think it's around 3 to 4 per seconds. ⁓ How many? 3 to 4 per seconds? ⁓ Yeah. ⁓ How many seconds in a day? ⁓ 3600\. ⁓ 3600 seconds in a day? ⁓ I know it. ⁓ What? ⁓ 60 seconds.

16 minutes into 24\. ⁓ So, ⁓ that's why. ⁓ Zero point. So, ⁓ now if it's less than one, I'll take a concrete number, which is let's say it is where did it go.

Okay, it is 0.006 alerts per second. 0.006 alerts per second. ⁓ So is this your peak concurrency for this system? This is my average concurrency, right? So ⁓ now my peak. would be the peak? Let's assume in the worst case scenario, everything going down, we can say 10x this, but it's still less than like our peak would be ⁓ one to two maybe.

That should be at peak. That's it? 1 to 2 alerts per second? Yeah. Based on 0.006 alert per second, if I... At an organization, you would not get more than 5000 incidents in a day. Even at a, like not at Amazon scale, but at decent organization scale. 5000 is very high, yeah. Exactly, right? So even at 5000 alerts per day, you're saying your peak alerts that you would receive in a system is 1 to 2\. Per second. ⁓ Peak, at that second, peak number of alerts I'm getting.

No, right? Because when one goes down, ⁓ you have cascading effects. ⁓ okay. You mentioned it. ⁓ You mentioned it. But I'm assuming that there is, okay, we have something called grouping where if one goes down and they're cascading. then all of this, correct, but all of this is coming as a raw event from incident. Okay. ⁓ Correct. So you would still get those events. ⁓ This means that you need not act on every single one. Act on every one, yeah.

Sneha Mehra (02:13:26)  
This is super important. Otherwise you're just wasting your tokens and your entire computer infra. Because if it's a cascading effect that has emerged, you have to have that grouping logic that typically these tools have. If they send you raw events, or if multiple systems are that they're integrating with, you have to know that these are related. ⁓ And if I fix this, then I take over. So that is one of the important choices to make. Okay, so I have 5,000 alerts per day. Peak certainly would be, let's say,

50 alerts if and you group it then it would have like two to three unique incidents that you're dealing with. Let's say three. Right. And then you would want to still scale it because if let's say three different system went down due to one reason it would be difficult for you to group them under one because you have to treat them individually. You multiply it by let's say 10 so roughly 30 worst case peak peak you would want to handle at one runtime. Yeah. Okay. Now given this ⁓ is

⁓ what we're dealing with, ⁓ what your very high level overview of the system looks like. So you typically have your ⁓ pager duties, end duties, everything else of the world. They're sending you over Webook, you ingest it in Kafka, and then you ⁓ run it with your executor. These are classic standard day zero architecture for the same, right? And there's where our agentic loop runs. Yes. ⁓ I would also split based on priority.

Okay, the splitting plays on priority. So you get it into different Kafka topics depending on the priority. Great. Then your executors will be per priority based? Yes. Okay. What does information your executor needs on runtime? So logs, ⁓ metrics, ⁓ code references, the GitLab repo and I'm guessing so how we configure PagerDuty is we have each PagerDuty alert.

containing this information along with OpDocs. So ⁓ the alert has an associated OpDoc with it. So that comes, ⁓ so I'm guessing all of this information is flowing. ⁓ Metrics will obviously have the Grafana alert because of which it is failing. So it will go look at that particular dashboard. ⁓ Code will just be the GitLab reference to the, or the GitHub reference to the code base. ⁓ Logs is basically, I know the pod or the, ⁓

Sneha Mehra (02:15:51)  
⁓ the namespace in which it is failing and I can go to the application or deployment for that. ⁓ This one other thing I would need is memory ⁓ in the system because I can. This is a great system to actually learn from what has happened in the past ⁓ and for me this system should then store memories of incidents that have happened in the past and then use them in the future for more incidents similar instance. And also improve your Op books ⁓ improve your Op docs as well. Yes, right.

Okay. Now one of the most important one is deployment. Yes. Because very likely most incidents happen because of a recent deployment. ⁓ So if you check there is a recent deployment that has happened, your first remediation step would be ⁓ to going back. Okay. Now where is this entire correspondence happening? ⁓ What do you mean? Now this, what is this executor doing? Like where the output of this executor is flowing? So the executor output is flowing to

Basically different integration points. So one easy would be I write back to the ⁓ PagerDuty itself in comments, whatever the ⁓ this is doing the other what we have or what we are trying to build is slackbot integration. So it sends a message to the channel on which the PagerDuty incident comes so that we have the on call in the loop on what the agent is doing. Or a classic Jira ticket. right. So typically every organization has a place which has a

where the entire correspondence happens. ⁓ Typically, JIRA, if you assume, then JIRA or linear or whatever. ⁓ So ⁓ we need essentially this entire triage is one place where that on-call person is coming because you don't want to change the on-call person's behavior that, now we to log into a different portal. So this connectors needs to be there, which could be internal tool calls. ⁓ So whenever an incident happens, you need to have a database.

which holds this entire set of information, which says, this is my incident message, sorry, this is my incident ID, this is the description of the incident that I have, ⁓ and this is its relevant Slack thread, or your PagerDuty ID, or your Jira ticket number, so that it becomes easier for your agent to make a tool call and add correspondence to it. ⁓ So all of those will be tool calls. ⁓

Sneha Mehra (02:18:16)  
Like as I was mentioning, in PagerDuty configuration, we add all of these things. So in PagerDuty configuration, there is Slack. In PagerDuty configuration, we have all the additional stuff ⁓ that you... Because even as a developer, if I'm going without an AI also, it helps me out ⁓ finding out these things. So yeah, I would keep it there instead of a separate database because then the database can go still. Yes. But again, you would need that corresponding reasoning trail so that you do debug, observability. This could be...

So again, this database would help you drive your memories. ⁓ okay. Yeah. Memory, anyway, I wanted to keep a separate database. So yeah. So again, all of that would flow into the system. You could use ⁓ Postgres with PGA interactions or whatever to ⁓ then take a look at it later and improve on your OpDocs and runbooks, et cetera. ⁓ Okay. Are you happy with this overall flow? We discussed at very high level and it has a lot of tool calls integrated. Anything specific that you would do apart from this?

⁓ One thing is having an alert on this system because this system is obviously our own thing. ⁓ So if our queue size is becoming larger, then we obviously might breach the SLAs that we have in terms of response time. ⁓ if we get multiple P1 incidents at the same time ⁓ and our time to process each one is ⁓ maybe 30 seconds, maybe 40 seconds, ⁓ then we might not be able to

solve the third incident within the 90 second time frame. Correct. We have a P1 pin 90 second within which we have to output a triage, which means your average latency should be within 30 seconds. Now within 30 seconds, how many LLM calls can you make? ⁓ Max five to six. ⁓ Sequential. Max five sequential. Max five sequential. And parallel? Parallel I can make... ⁓

any number again if there is no not everything is in parallel then just it will take five seconds six seconds for me to complete all of them so ⁓ now given you can make five sequential calls right so assume if you have five ⁓ assume you have like four or five agents ready which means each agent in its sequential iteration can make in five LLM calls it need to be it should be able to deduce

Sneha Mehra (02:20:40)  
Yes, what it needs. Yes, so you do not have more than that, which means now your system prompt your user prompt that you're passing that needs to be sharp enough. To output this to all you have is this five or even let's say double it. Let's say 10 LLM calls per agent. That's why we cannot do all of this agent calls like all of this ⁓ calls to let's say if you're doing with a critical agent which it goes and like one agent that takes look at.

your logs and tries to deduce what takes a look at metrics and tries to deduce one looks at let's say code and tries to deduce they cannot have more than five to seven to eight iterations in order to come up to this final piece that hey this is where the root cause lies right so but all these agents can run in parallel which means you need to have this provisioned capacity of five into let's say I have three agents running in parallel to do this five threes are fifteen calls in parallelism making for one incident

Yes. Given that you are getting ⁓ roughly, let's say peak pay, are getting 30 incidents or not 30 seconds, 1 to 2 alerts at scale. So let's say 10, 10 you are getting. So you would be making roughly 150 LLM calls. Yes. During peak time. I can give an LLM calls spans a certain duration. There will be some interleaving, but 150 LLM calls peak ⁓ LLM calls ⁓ in a second. ⁓

in a second is what we will be making. So which means this will constitute your top level LLM rate limiting that you are having. ⁓ Typically one of the things that we don't appreciate enough is imagine you are a big company imagine Amazon and in at Amazon imagine that an outage happened and the system that deals with this outage that alert system that deals with this outage or the system which

raises the alert raises the page of duty or whatever that itself is unreachable or is down that is problematic so here also given ⁓ applying the same concept over here to make sure that you have enough capacity enough rate limits that at peak time you should be able to make calls to LLMs more importantly have fallbacks yeah because if let's say Gemini is down what do you do

Sneha Mehra (02:23:07)  
You cannot say, I cannot fix my issue because Gemini is down. You need to have fallback, which means your agentic loop cannot be a model specific. It has to span at least two providers, at least two providers. I'm just making notes here as well. It is at least two providers because uptime of this is important. Apart from the uptime of the infrastructure that we typically discuss in system design stuff, uptime of LLM, like having that fallback is super important.

and having enough capacity, rate ⁓ limits is super important. Typically, you might just want to use a separate account for this if possible, or you want to use a separate key depending on how you're negotiating your contract with your LLM provider. Okay, some of these are decisions that people typically keep on just trying to highlight that this ⁓ is nothing but your harness, like how reliable and robust your harness is.

Like other loop you can write but these are the stuff that we typically don't think about. Awesome. Thanks Pratik. Anybody else who wants to chip in? For the next part. Where we discussed functional requirement we keep things kind of we discussed keep things to keep in mind. ⁓ And we discussed number of peak LLM calls. Right? Now capacity estimation. We'll take a look at it. Very simple calculations we'll do. Capacity estimation. Deepesh.

With respect to capacity, even I'm dealing with 10 incidents max in parallel, we certainly would not need to worry about database. Do think we need to worry about database at all? What do think?

Sneha Mehra (02:24:47)  
⁓ So Arpeth, ⁓ as you are building this investigation team, agent and it's dumping the whole investigation in a dock right for the on-call person. ⁓ So do you want to keep some sort of stepping limits right? How many steps you want to go your agent to go deep in right? ⁓ Because it's possible like the skills. You are always limited, that's what we discussed right? You are always limited by so your steps, your LLM calls that you making, ⁓ it cannot go beyond five steps.

So that is why this number is very important. Okay understood yeah. Right so this is your multi reasoning step that you are doing right. That that number is very important right. But now you talking about we are not dumping. So ⁓ one more one very important thing is given that all this information will be consumed by what your on caller and on caller let's say goes through Jira ⁓ UI to ⁓ do it. This information cannot be very confusing.

You cannot give like massive prescription that it takes time for him to read. You cannot just output what cloud outputted as is to that. ⁓ Give them very succinct thing. This happened, this happened, this is the command that you run that would ⁓ fix this. Then you run this command, then you do this, then you do this with hyperlinking in place. So do not unnecessarily keep things verbose. This needs to be very crisp and short. So this is one of the most important things that you will provide in your LLM prompt. The triad that you are posting.

cannot be too verbose for your on-caller to go through because by the time on-caller goes through it, gone. ⁓ You cannot let that happen. Okay. ⁓ Ripesh, apart from this, ⁓ you said ⁓ large amount of data, we are putting every correspondence in database. ⁓ Don't you think database will have too much data? Can you calculate? Make a rough estimate of it. ⁓

If you are putting the whole doc in the database right? Which doc? ⁓ The investigation doc. ⁓ What is investigation? There is an investigation doc. ⁓ We have discussed investigation doc. Are you talking about OpDoc? ⁓ These are runbooks. ⁓ Yeah, not the runbooks. The context of the 5 steps which our LLM has done right? ⁓ So if you want to give our on-call person

Sneha Mehra (02:27:12)  
⁓ handy run book or would you this is not run book right these are not run books are separate run books are already defined steps for you to do things correct this is just what LLM did you don't want to express everything to on call on call is already frustrated he or she woke up in the middle of the night to fix this not overrun correct this data is like every single step every single every single reasoning should not be visible to your if they want to they can exploit but

I don't think Oncall would ever do that. Oncall will try to just fix it as fast as possible. Correct? And then go back to sleep. Right? Yeah. ⁓ Okay. Any other considerations? Let me ask a very specific question. How will I generate my first triage report for my Oncall agent? So we discussed multi-agent where each one goes and figures regular metrics, etc. Let's say once it got all the response. Now what?

So what all things it is calling, we kind of have an outline for that and then how it is synthesizing the information and what are the important things to provide in the prompt.

Sneha Mehra (02:28:24)  
⁓ Maybe the one thing I think of here is like which alerts are actionable, right? Which alerts require ⁓ action from any human. ⁓ So all of that can happen at due duplication at this level. At this level. At this level. So actionable, grouping, etc. All of that will go over here. Okay. ⁓ Then?

What all things is executor looking at?

Code we discussed, metrics we discussed, right? Logs we discussed. How is it going to check runbooks?

Sneha Mehra (02:29:10)  
Ideally yes, if one of the other is right, they don't know right, so they have to go to runbook to see ⁓ how is it going to runbook.

How is your agent going to run book and checking it? Where is the run book? There are thousands of run books.

Sneha Mehra (02:29:29)  
We have to build a map of it somewhere, ⁓ With type of alerts, ⁓ which runbooks should be referred for. But what if that runbook does not contain the information, but some other runbook contains the information, right? You can't just rely on one runbook. ⁓ What if organization starts with zero runbooks or has very limited set of runbooks? Because, ⁓ and what if, worst case, you saw an issue which is not tagged to a runbook, right? It's a fresh issue. Then what do you do?

So it's an open ended retrieval problem, right? So this is where your rack system comes in.

The entire Rack system that we discussed is your vector database. You need this where all the run books.

Sneha Mehra (02:30:17)  
are indexed and your agent makes a call to get the relevant stuff. So this is another lookup that you would do. You'll find everything that is relevant. ⁓ So the entire rack system database is just copy paste over here. I just replace it with the vector database. That will be required. So once it has all the information, now what do you do? What is your key things that you provide in the prompt that provides you output for it to fix?

Sneha Mehra (02:30:48)  
⁓ Let me ask Sibyl, how does a prompt look like the page?

Sneha Mehra (02:31:00)  
There is a board at the end. Given all this information, give me steps to fix it. Right? You could start something as simple as this. Yeah, understood. Right? But what's the next level? Is it good enough?

Sneha Mehra (02:31:17)  
We saw so many prompts and we are being oddly very specific.

Sneha Mehra (02:31:27)  
Right? We go to Kevin. Wait, thanks Dipesh. Kevin, how does my prompt at executor look like? So I just paid all the tool calls, gathered all the information. So it created sub-agents, went, gathered the data. Oracle Sutter has everything. Now it needs to dissect. It needs to come up with a triage. What do I do? What do I write? ⁓ What are things I can take care of? ⁓ Yeah, so basically like we are thinking from the perspective of an on-call ⁓ engineer who's going to take a look at it, right? And it needs to be...

⁓ very clear ⁓ and what actions they need to take. So specifically, ⁓ if the ⁓ agent is going through ⁓ all the run books, ⁓ and then all the ⁓ matrix, ⁓ the logs, ⁓ and all of that, right? So then ⁓ one thing that I can think of it, like if it has like some deterministic ⁓ knowledge, like what, ⁓ so then the first step would be like, you know,

providing a brief description of what, ⁓ description of like this is the issue and this is the ⁓ potential ⁓ remediation ⁓ and mentioning about like how it came up with that ⁓ information. So it could be like a query to ⁓ your logs, like what query ⁓ it ran and then it got that information specifically ⁓ pointing to the specific ⁓ book. ⁓

⁓ and also providing, let's say if it was a particular deployment ⁓ that caused that issue then... But it would not know, right? It would just give you a list of past deployments on the service. So now it's your logic to say in past what time? 5 minutes, 10 minutes or you say past 5 deployments and pick do you think it could be the issue? So it would say what are the chances that it happened because of a recent deployment? ⁓ LMS smart enough to figure out

that the deployment happened last week, cannot happen. It cannot change anything today. But imagine you have a change flag, ⁓ your feature flag that got flipped. ⁓ So based on the logs, ⁓ it can potentially, ⁓ if there is a specific change that happened and that triggered this issue, ⁓ then potentially it can you have deployment, you got past deployments. ⁓ Deployment happened two weeks ago.

Sneha Mehra (02:33:52)  
But your feature flag turned on today. So you need to infuse your feature flag as well into this that audit is pure feature flag. Fair? Can you repeat one more time? Feature flag you know a feature flag. Yeah, yeah. Deployment happened two weeks ago. ⁓ LLF will say this is not a cause of issue. Correct? Because you twist, you changed the feature flag today. You turned it on today. The code was like 20 days ago or two weeks ago.

Right, but feature flag turned on today. So today it went live to all the users. So that log also needs to be there. Right. Right. So audit history of feature flag relevant to this service. Now imagine now all the metadata that you need to hold for a service. These are relevant feature flags for this service. These were the feature flags that were used in the deployment. All this information needs to be consumable. It should be in a consumable format. That's why people say your agent ⁓ is as good as the information that you provided.

the better the information you provide in a better format you smarter your agent will be. ⁓ I am just opening up different possibilities, will say why is this system so 10,000 feet to you but these are the things which is very AI related, like what are we providing for it to proceed ⁓ and these are the unknowns that we have to think about and this is where our intuition typically kick in. ⁓ Okay, so you wrote that from

It outputted. One of the most important things you forgot to provide this as an input. So you provided tool access to get things. You provided vector database with runbooks. Right? You forgot one big one super important thing. Other two very important things.

Sneha Mehra (02:35:48)  
So would the system ⁓ act ⁓ remediation or are we saying the on-call would do the remediation? On-call will fix it. Okay, great question. What will you prefer? Depending on the severity ⁓ and potential of the issues that have been happening ⁓ based on that, know, some of them, yeah, the confidence, right? Essentially, if we are confident on certain low hanging issues, then

then potentially those could be remediated by the system based on the history. ⁓ And if there are like, yeah, high priority systems and you would want your own call to take a look at them. you're considering these are the... But imagine this, ⁓ where on-call is anybody going to be there for an incident? ⁓ Right? Sorry? ⁓ Your on-call will always be there for the incident when the incident happens. ⁓ They're just aiding it. This is assisting the on-call. Right? You don't want AI to take actions.

without a human because it's outage. even if they take action, right? I mean, would be more like, say if it's a PR ⁓ change that they need to make, ⁓ the agent will create the PR, but still the reviewing would be human. You're assuming PR, I'm assuming database config changes, feature flag reversal, right? All those actions are also there. You're assuming it's code fault. I'm assuming it's subsystem faults. Yeah, even if it is a config change that can still be a PR.

Do you have a config? You are assuming your config changes the PR driven change. ⁓ I imagine you have a config system that has an MCP tool expose which allows agent to flip any bit that it wants. Right? Problem. Yeah, that I would get because that you wouldn't want to do that. ⁓ Hence, the idea is by the time person comes online, all the critical that it needs to take action should be there because the person is there is attending to it. Right?

the person can take actions. ⁓ So I am just saying how severe the given its outages, outages because if AI messes up something you go bazack problem. So that another thing that we should add is past incidents that has happened. You do a semantic lookup on that or a cosine similarity on that. This is similar system that happened and what steps agent took. There should be one more agent.

Sneha Mehra (02:38:13)  
that is running from this where an issue comes that takes similar incidents that has happened in the past and actions agent took. That's what memory is. ⁓ Even like run book we kind of mentioned so run book has these past incidents. ⁓ not all this. Run book has limitations. ⁓

⁓ No, we have two runbooks, right? One runbook is remediation and one runbook is essentially ⁓ the on-call has kind of kept like the past incident record like each day. ⁓ Right? So that is, ⁓ that would, I would consider still as a runbook. So if you have it, great. I'm just being explicit, right? Refer to your past systems and how agent tried to navigate what steps it took so that agent has a guiding light, how to proceed with that, right? Rather than it hallucinating in different directions.

It can leverage that, but it has some weightage, not full weightage, it has some weightage for it because it could be there that we don't want it to take the exact same path. It would be different thing. So it still needs to be reasoning. It still needs to be reasoning, but hey, you might consider these steps, but feel free to take differences if you feel like kind of this. ⁓ That is an important thing. So that this way you're still letting it ⁓ converge a bit. You're helping, you're guiding the agent a bit.

but you are not forcing agent to take the same path. ⁓ So because it is an exploratory use case, it's still a reasoning use case. ⁓ Okay. ⁓ Next up, sorry, the second important part is format or style of triage.

So ideally you should have how the tonality that structure the command how it needs to be formulated if it's written ⁓ in a very similar way up corrosion ⁓ or there's a less cognitive load. Again I'm just making the life of my on-caller simpler. ⁓ Now we'll go to the next part. The next part is something which is highlights of this that we have to worry about which is your remediation actions can be irreversible.

Sneha Mehra (02:40:26)  
For example, if you are doing a pod restart, it's an irreversible action. Once you trigger the restart, you cannot do it. ⁓ The reason is why I'm saying this is because should you let your agent take these actions or not? That's the point I'm debating. Because it could be reversible action, that could be irreversible actions. So now for all the actions, for all the MCP tool definitions that you have, add a flag which says if it's a reversible or a reversible action. And then you can have a staggered rollout and say, hey, if these are reversible action, I'll let my agent take.

Here is an irreversible action, I won't let my agent do it. Okay, then next up, Capacity Planning we discussed, this is very simple, the same five questions that we did the number of LLM calls that we are trying to make, etc, etc. Storage is anyway never going to be a problem over here, never. It's going to be peanuts, right? Whatever we are storing. Now, this was high level architecture we discussed, all the access that it needs is what we have. Okay, ⁓ here one interesting calculation. ⁓ In...

90 seconds span that was our p99 for ⁓ a p1 issue. So in 90 seconds we have 90 into 10, 10 is a peak alerts that we calculated that we would receive. In 90 seconds worst case we'll get 900 incidents. Some of them would be redundant agreed, right? What we need is let's say if we assign one incident to one server, then we'll need 900 servers to do it. Problem, we don't need this.

Hence the deduplication logic is essential so that you don't deal with 900 instances. ⁓ Plus even if we have to, worst case we have to, given all of your use cases ⁓ IO bound, not CPU bound or memory bound, it's still memory bound but not ⁓ CPU bound or most of this is LLM calls that you're dealing with, with some memory and you can still offload the memory to external storage, you can have ⁓ one core per incident.

So if let's say you have a four core machine, ⁓ You have a four core machine, but it can handle four incidents. So this way you slash your number of servers requirement to 225\. But it is still large, but that is peak. And if you do de-duplication and all grouping, grouping, et cetera, you can bring it on your word for the, this is the worst case, ⁓ absolute worst case analysis that we are doing. ⁓ That's where your short circuiting of your incidents is essential. If you think that, hey, ⁓ it's not worth it, I'm already.

Sneha Mehra (02:42:48)  
solving an existing current ongoing issue which is on the similar lines I will chop like I'll kill myself because I know it will be done by someone else because someone else is already some agent is already working on it right so then that's one of the things that we have to deal with so roughly 80 to 100 servers is what you would need okay now let's look at very high level architecture of this and then we'll look at human taking action so you have your pager duty and zine duty coming in

Again, I'm using GRIs my entire ticketing system and correspondence. It can change depending on organization, but you typically or you can use pager duty or incident or whatever to do it. ⁓ Now, whenever you get, whenever you get alerts, your information goes in over here. It first creates ticket, et cetera, cetera. It's by default or integration. Then it goes in here. Your deduplication happens. Now, what is the deduplication logic? The deduplication logic is for this service.

for this ⁓ service for this condition if something is already in progress you run it with a detail of let's say 1 hour or 5 minutes or whatever then it does a deduplication over here ⁓ once it filters it out then it goes into kafka from this kafka the executor picks up then from here you have your postgres where your memories and your ⁓ incident state management takes care of

This incident is ⁓ triage done, this done, that done, et et You take care of this. Then you have your orchestrator, is essentially scaling your queue, scaling your consumers and executors depending on the queue length. That's what it does. These executors have access to metrics, log, deployment history, run book index, code. And another thing that we discussed was feature flag.

Sneha Mehra (02:44:40)  
and similar incidents.

Sneha Mehra (02:44:45)  
⁓ And depending on this, runs your loop. It runs whatever you'd want to do, multi agent, because you would be checking this in parallel, right? And within that, you would have your reasoning steps to reason, reason, reason until it converges. Everything should have a max step attached to it. You cannot go beyond five to seven LLM calls or within the time. Time is a better thing. You cannot go beyond that to come up with ⁓ your... ⁓

your diagnosis. ⁓ Okay. Now what executors output also goes as correspondence. So now this would need to have connector setup with this Jira so that it adds that correspondence. So the idea is within 90 seconds if I have my correspondence over here, so my on-caller can take a look at it. Now given this is just initial triaging which helps this on-caller issue like fix the issue very quickly. Now there are two important things. First thing is that

This is not the end of the job for my agent. This is not the end of the job because given that this guy did so much of heavy lifting into figuring out all the important stuff, creating citations and links. Citation and links and hyperlinks and everything. This on caller should be able via chat. This is the same same database. This Postgres.

and this so imagine executors continuously pushing the state over here. So given this it did a lot of heavy loading and kept updating, ⁓ kept the state updated. ⁓ I should have a simple chat interface with the same set of tool integration that Oncall can talk to using the same database as this one and continue the chat. Given all the context that it used to generate this first try is already loaded and then can pull it more if required.

So this becomes a standard chat based interface, which is essentially a rag application, ⁓ which is figuring out what it needs to build, how it is relevant, what needs to be done, et cetera, et cetera. ⁓ This standard rag implementation. ⁓ Add vector database, et cetera, et cetera, the entire downstream subsystem as is. ⁓ That's why. ⁓ Second important part is whatever this triage it uploaded over here in correspondence. ⁓ Again, as I said, this job is not done of executor. Your triage agent should keep running in background.

Sneha Mehra (02:47:10)  
and should critique ⁓ whatever is happening. When a user adds a new correspondence, your triage should validate is this correct or why this is happening. It can continue doing its own exploratory analysis on the same set of stuff until it finds that out. If it finds out because ⁓ all it did is first 90 seconds it did help you triage. But what is the long-term fix? If you need any help, that's first is chat.

but it can proactively push what it found by redoing this analysis or digging deeper or digging deeper. This is you use a different model and different ways. Now here you can just use low hanging fruits. You can move this code part to critic.

You will say that my critic will look at code and again all these metrics again but code in much more detail to figure out what has happened and will help you raise the PR or help you change the PR or if the PR is created will help you modify it and review not just review not code review but critic the actions that you are taking because you will keep adding your correspondence this critic is adding its correspondence on this part now depending on how sophisticated you would want to make you would choose to add a critic. ⁓ Okay.

Next up, now here what has happened is we wanted up until now this user is taking all the action depending on what is written over here and it's using his or her own brain to decide what action it should be taken. But now imagine if you want AI to even execute, if you want AI to even execute then your executor notes here what I did I separated few things out. I do not want this executors

to also execute directly when it's strategy because my first step is to figure out the root cause, help onColor and because most of the stuff will be taken care by onColor. But if I want AI to execute, I would never want that AI to execute ever. Right? But if it's no ⁓ action, imagine on the correspondence, you have a button which says ⁓ run and the command is given. You press that run button and it runs it.

Sneha Mehra (02:49:27)  
⁓ So think of it as human in the loop implementation, ⁓ fancy word, imagine I would rather still implement it as a API call where in my correspondence I get this run button which says take this thing and run over here. It's not, it doesn't require agent to do it. ⁓ So it literally just makes an MCP tool called to an executor. This is your command, let me get a separate name. This is your command executor, which has connectivity.

to K8S, AWS, internal feature flag, all other subsystems and can run commands. So the output of what you got in your JIRA correspondence out of your critique and your correspondence, the command, it could be shell command, it could be a simple MCP tool called invocation. So that command execution, I should not be having to ⁓ log into a server to do something. It should be done via this executer node. So you define all the tools

that requires you to fire bash commands. That's it. If it is used to fire a command on a particular instance, which abstracts out creating of shell of SSH ⁓ connection to that server via bash churn and fire a command, get the response and do it. Fires a command on its deployment tool, fires a command on let's say some other system. Let's say you'd want to take down your auto scaling group. If it figures out the step, then you can just have a, let's say a highlighter. You select the text and say, run it.

that goes via mcptool and it executes over here so this has access to all the system ⁓ and it runs the command and it exposes the mcp server for you to do it now you can be as creative with this as you like and say that hey ⁓ I want to ⁓ execute let's say deletion of this autoscaling group you just say in natural language it's written in triage you pick that and you say run as mcptool your IR your ⁓

Instead of auto remediation agent ask you which auto scale group you want to delete. You copy paste the name and say delete this one. It makes a tool call, waits for human in the loop for you to take that action and then you run it. So what I doing is I making life of this oncaller easier. Rather than going and going to AWS console clicking clicking and figuring out I might have all the tools exposed here at mcp level and this might make call to multiple mcp servers and I have this entire human in the loop that we discussed yesterday to be implemented here.

Sneha Mehra (02:51:49)  
that I just say do this, it figures out multiple SAP tools, invoke, invoke, invoke in a choppy style. ⁓ But here it's still human who is using his or her ability to decide what to run if this command is correct and then taking the action. There's not some noob who is doing it. He's a pro, he's a proper, proper on-call engineer who is there to fix this issue as it happened. ⁓

Again, the ideas you can make, if you observe, it's just literally imagination is the limit. You can make any kind of sophistication that you like. But ⁓ more importantly, the things that I found people not talking about is this sort of stuff like where... There's this notion that MCPs are dead and CLIs everything, etc. But this is a very good example of why MCP makes your life simple. Because if this has access to everything, I can just expose an MCP server over here.

I can just say in natural language that do this it will find out multiple actions to be taken with a reasoning step for example it takes that action makes my life easier. So your whole job is to make this life easy like this person's life easy. ⁓ You can make everything as sophisticated as you like but if you building this for organization start very small. Solve one problem at a time. Make sure your entire sanity is intact. Sanity with respect to

the commands that is outputting that is correct you have this critical loop and then you enhance, enhance, however you would want to Right? Awesome. This is some very high level overview of how you design an incident auto remediation system. Right? That very high level overview. But again, the loop implementation we have discussed it separately. That's why I skipped it. But on high level part, how the flow should look like is what we discussed. Next week ⁓ is next week, Saturday evals. But

Now it will be system design heavy, ⁓ where we design two systems. ⁓ First is code reviewer. Everybody is kind of doing that and self-updating AI documentation system, ⁓ which identifies a doc drift between your docs and your API where it gets updated. ⁓ It fixes the drift, ⁓ gives it a structure, et cetera, et cetera, how to design it. ⁓ I'm designing the same system at Razorpay. So copy paste of it, right? So first we'll discuss evals, right? Where they come in handy. We'll see ways to breach.

Sneha Mehra (02:54:08)  
everything and show demo of that. Then we discuss two systems which is code reviewer and self-updating AI documentation system and in this process we will look at two prototypes. ⁓ On Sunday we discuss one big system which is I don't know I still haven't counted zero. We will discuss so all this stuff I was going to initially cover all this stuff but we have kind of covered this with this this will kind of cover either covered this or will cover it in eval's part.

So all this has moved to appendix. I'll see what I would want to pull this up over here, but we'll go in depth of this system. This is nothing but agent studio that every company is trying to build in some shape and size. We kind of touched upon it when we were discussing this multi-agent loop. Kind of touched upon it today. ⁓ And your product, this is one thing that I'm trying to brainstorm on, which is about production scenarios, because in first session, what...

lot of people resonated with this like people sharing their production stories and like openly saying hey this is the situation that I'm facing what do we do so it's more of probing probing probing so very concrete examples I'm trying to come up with in which we'll probe probe probe that if this is an AI system what all things could go wrong what all things could we do better so kind of amalgamation of what we discussed with sprinkle of system design to it and like because they are like long running systems so we have to take care of lot of stuff going down and so this will be

⁓ So we might start with natural language workflow engine to start with and then we'll discuss production scenarios and wartime stories. ⁓ But pre-reads, ⁓ keep an eye on it. I'll make sure that I'll send an email as well for that. ⁓ And very likely I'll be adding some more periods for six sessions that we are having. So for this session, depending on how the curriculum goes or the topics goes that I will be covering. ⁓

⁓ But it will be very pragmatic in any case whatever stuff I was going to cover will be added in the appendix any which way. that you can, okay very easy like it's very classic system design like cost attribution, open telemetry, regression harness we kind of touched upon it. Architecture decision framers was literally we should do sync or in parallel. It felt very redundant and because we almost always touch in almost all the discussions on what goes sync, what goes async.

Sneha Mehra (02:56:29)  
I'm like just making sure that I cover the widest ⁓ range of topics I could possibly cover. Hands a little shaky on the final session, but I'll figure out a way to do it in next couple of days. ⁓ I'm a fully transparent. I want to just optimize for the breadth that I'm trying to cover. Also, folks want to drop off, feel free to drop off. This is what I wanted to cover. We'll take questions. We'll start with Kevin. Kevin, good.

Yeah, one thing I wanted to add on this system design is if you want to add like the postmortems are typically very painful and very time consuming. ⁓ towards the end. that's a good one. Postmortem generation. ⁓ Yeah. Yeah. Is that that will help kind of complete the Lula. ⁓ Yes. ⁓ Very nice. ⁓ Yeah. So you have all the things you have correspondence, ⁓ have triage, ⁓ you have memory, ⁓ everything to create your postmortem. Yeah. ⁓

Awesome. Thanks Kevin for adding that. Yeah. ⁓ Abhishek, ahead. Actually for post-mortem you need a feedback loop as well. If you want post-mortem because AI is not taking all the decisions. There's still a human in the loop. That is a triage thing. ⁓ so you're assuming, so I was assuming all the command execution goes with MCP. So I have a command execution log. ⁓ okay. ⁓ I like, I was assuming the initial thing where we don't execute everything via... And the user executes it. Okay. ⁓

But if user is executing then user should ideally put it in comments that these are the commands I executed. ⁓ But if that's not happening, this provides you a central place to log all the commands for that issue. ⁓ Central command log.

Sneha Mehra (02:58:14)  
better. Perfect. And now this justifies decision for doing this. ⁓ Very important. Okay. Anshul, sorry. Someone else I forgot who I pulled in. So Arpit, I'm actually on call for the current sprint and doing your course. was thinking of something similar. ⁓ Yeah. A question here. So I was also thinking to integrate it with Jira. Okay. ⁓ How to ⁓ do the chunking?

⁓ should it be the GID with the description, but the comments, right? A lot of times the description is a bit weak and this is the place where the answer is. And then we say, okay, so how, how should it be stored sort of, ⁓ so the chunking is perfect for me. Why is chunking a problem? Like I was suggesting that ⁓ your GIDA correspondence is way too long for it to provide it in one context. I'm suggesting that.

Yes. I'm thinking of that. That could be a scenario, right? Is that a case in that case? would, I would still say that summarization is better. Like you would still have like different comments as different things, like different entities in the Rose. You would anyway have it, but given it's incident, you would likely like ⁓ one million cortexes, anyway, too large, right? Like how, how big is your G.R. correspondence? No. So I was thinking that, ⁓ should it be, ⁓ for one G.R. ID, the whole

⁓ like a whole comments with the description, but sometimes the comments get, ⁓ so what happens is it goes to one team then the other. okay. no, but this is on call. Yes. So on call ⁓ is more about incident that happened, right? I'm talking about triaging and fixing ⁓ that team doing correspondence. It's there. It's a different problem. Right? So that falls under a postmortem thing. Correct. ⁓

That does not, and again, that is like you initially, you need, you initially remediate the issue and then you work on a long-term fix. Okay. So like, like split it to ⁓ split it to this is on-call incident remediation. ⁓ Then when your long-term fix comes there, the long-term correspondence comes. And then if you'd want to create a final summary document of it, then you go step by step, you say, you imagine you start with the blank doc. ⁓ say these pick first 10 comments, create a doc.

Sneha Mehra (03:00:42)  
Then pick next 10 comments, update the doc. Next 10 comments, update the doc. So this doc will be ever growing and it will summarize and rewrite the doc every time.

⁓ So then you create a final summary document, for your own call, don't need that your own color correspondence will not, because you're, you're prioritizing fixing of issue. You take 10 minutes or 15 minutes to fix that issue. won't wait for multiple tips to add like ⁓ lengthy stuff during ⁓ the outage during that, when that incident is happening. Yeah. I was thinking more on the lines of on-call plus support. ⁓ So support is long term. Right? So there it's like more of a summarization problem.

You'll this final summary and actionable. think of it like, ⁓ Google meet, ⁓ Gemini notes that we typically take. ⁓ And when you have like multiple minutes, it breaks it into five minute terms and keeps updating the dog extracts the important information. Next set of actions, open items, et cetera, et cetera, et cetera, et cetera. And then it creates a final document kind of what we discussed with multi-agent today. ⁓ And then another question, the phase two that you mentioned is this diagram, right?

⁓ background process that would be running. So are we implying like sort of till, ⁓ till the JIRA ticket is moved to maybe a done, right? ⁓ Yeah. So there's a state. now there just, it would be, it cannot be done because done might mean my long-term fix is dead. So you might say AI triage complete ⁓ something like this or, or, or, ⁓ out of outage. It's kind of, we used to have out of how it is. It sounded cool out. That's why out of outage.

So we had this out of our status. So that's it's very specific to our right. So it is like, am out of outage, but my long-term fix is not there. I still want to trick it, track it as an open item, I just want to mark my outage is not it's remediated at this moment. Long-term fix is not yet implemented. But, ⁓ so ⁓ what should be the point till it should happen? Like, ⁓ won't it be too costly for us to run it in background sort of? ⁓ So that depends on your company's bank balance, bro. ⁓

Sneha Mehra (03:02:51)  
⁓ Don't run if it's costly. ⁓ If you think it's helping, each one of these components, like for example we discussed 4-5 tool calls, if you think these tool calls are helpful, then only you do it. If you think that these agents are helpful, then only you run. If you think that critic is never helping. Imagine so many companies have turned on their AI code reviewer, ⁓ we as humans have started ignoring them. I correct. Nice. ⁓ Right? ⁓ At least I've done it, I've made a hypothesis.

⁓ And I see many people ignoring AI comments, they're respecting human comments now. ⁓ You see the typical AI slop in them. ⁓ So ⁓ same thing happens if you don't see value with critic, which is like running in background trying to see keep an eye on correspondence and trying to dig deeper on a more exhaustive task like code. ⁓

Like subsystems like downstream dependencies. you split responsibility into kind of real time and non real time use case, right? Real time, which means during outage and on real time is post outage, like post outage remediation or during outage remediation. like not like out of SLA part, like beyond 90 seconds. What do you do? You do this. Yeah. It would also be like one for on-call and maybe another for support chat or something like that. You are fixated on support. ⁓ You're fixated on support.

Yes, kind of that. ⁓ Yeah, and you mentioned here that when we did some estimations, right for a Foco machine, the maximum we could handle is maybe four incidents. But here, ⁓ most of the things that we would be doing is say tool calls. And then these sort of things, right. So it is not at all CPU intensive, it would be the IO intensive, right? Yes. ⁓ Why?

that pessimistic of a view in that we took. As in four only I could do 40\. Yeah. ⁓ Okay. So why four is because there is still some CPU involved. Correct. Yes. So you would want to give a dedicated CPU to each incident so that it doesn't plus if you do 500 there is still memory involved. There's a context window that can still bloat up. Right. So you don't want that to overwhelm in any way because if that going down to undo pressure. ⁓

Sneha Mehra (03:05:09)  
is more problematic because it's right now helping your on-call agent. And this is going to work when like this is going to work during crisis time. ⁓ you have to deal with a lot of pampering. ⁓ Okay. ⁓ Again, that's just justification. ⁓ If cost is becoming a factor, can of course make it eight. ⁓ I'm just saying like, you don't want this to go down due to whatever reasons. So you'd add more buffer. Right. ⁓ Thank you. ⁓

Good Anshu. ⁓ So the first question I have is like in general what I observed is whenever a service goes down or like there is some issue with that and there are some upstream and downstream services which get impacted with that right. ⁓ So should we send ⁓ like for a particular team what are the open incidents should we share it to LLM as well so that it can know like this is the related

⁓ Yes, that is what we discussed. If this is related, then I kill myself quickly. We discussed that during first iteration, you would have the short circuit evaluation, executors, that's how we came from 900 to 225 and then the ways to further reduce it was this one. So if you know that this is related, so that identity should be our first step. Is this a primary incident or a secondary incident? Kind of vague terminology, but is this primary incident? ⁓

So for that as well, like how will we even know that? Same service, similar issues. Same service at the time, within two minutes you getting the tissue.

But it will be different services, right? there will be different services which are getting... then different services you don't typically do that. You ⁓ work on it independently. You would typically do the same service different types of issues that... different type of incidents that are being raised. ⁓ Because you don't know. Because you are not ⁓ sure. Like you are not sure that these are related across services. Right. That's why I was thinking we should share all the active incidents for a particular team.

Sneha Mehra (03:07:16)  
If those are low in number then ⁓ that LLM poll will be in right order. You don't want in this case a LLM ⁓ having a false positive. ⁓ It will have a true negative if it goes in which is like hey I think they are rated but turns out they are not. ⁓ Probably. Because that is literally kind of suppressing a alert unnecessarily which someone should have looked at. ⁓

So given that this is something that is going to be called during crisis time, just do a deduplication or a grouping within a service. Okay. And another question is I have this related to chat, like user is interacting with chat to increase instances and all those things. ⁓ So it is interacting via MCP, right? So the question is how is it exactly happening? Like suppose there is a recommendation saying

increase 30 instances or increase memory or something like that. Good example. ⁓ So now chat will share this information and user will say or take action on top of that like increase the instances by 30 or maybe something like that. ⁓ So now user might have some permissions on MCP using which these instances can be increased. ⁓ But the issue here is suppose due to some issue like there will be again another LLM call which will

convert our human instruction into tool call or MCP No, no, no, no, no, no, no, no, no, no, no, ⁓ is the point of difference. So here what you have is you have an MCP tool defined which changes the number of instance count in an autoscaling group. ⁓ Right. ⁓ Get it. So you have this very specific tool definition defined in your thing that is literally says change instances in autoscaling group which requires an input which is an autoscaling group and says what is your

final number that you want the number of instances to be. ⁓ This LLM will convert, ⁓ it will generate the arguments for you that this is the auto scaling group and this is the desired number of instances. ⁓ Right, but there is an issue where LLM can convert that number like LLM is also not good. ⁓ So that's possible. That is where when it makes a tool call, you can add if you are unsure, you can add a human in the loop that this is the argument, do you approve this argument?

Sneha Mehra (03:09:43)  
or you can change it on the fly and say that it was not 25, was 26 stupid. So you change it to 26 literally and say run. ⁓ That's the human in the ⁓ Deterministic approach. ⁓ And how do we ensure that the human which is executing has those, I think that permissions will be sorted because they will already be in some group or something like that. ⁓ And none of the agent is able to execute those tools. Yes.

And also I think ⁓ there is something related to confidence as well. ⁓ in phase one, I think we decided like up to five iterations, whatever findings we have, we can return that. And in phase two, will dig deep into that. So is there any stopping condition or something like that? Like, okay, ⁓ at this point, ⁓ I will ⁓ exit the loop or something like that. is point of diminishing return. ⁓

Let's you see it getting stuck in the loop or it breached SLA. Let's say it breached twice the SLA. Then there is no point running because your on-call is already up and fixing it. Like it's like saying that you won't be able to do it. ⁓ Leave it. Like for example classic happens when you use a non-cloud model and think, ⁓ let it be. You cannot code. Leave it to cloud. Right? Kind of that situation. Now what is that criteria that depends on what kind of situations you see and you keep modifying.

your behavior. this is like very continuously improving system is what we are trying to build here. We start small, we observe and then we modify. Like what are the criteria? We see stuck in the loop, you take more than this much time. We see some third reason next time failing or not being able to fix it. You keep modifying.

Thank you. Thank you. ⁓ So ⁓ in the previous section, we were discussing this ⁓ example of evolving memory. ⁓ Where you had this like, I'm going to meet Sam, Sarah and everything, right? So there we covered like how we're maintaining that memory. I was just trying to understand how, how does the recall happen in that actual conversation where we need to reference that? That is a little graph query.

Sneha Mehra (03:12:03)  
So you provide all the relations that you have. You provide your graph schema. So different node types, where you provide your ID format. This is how my node type and node ID looks like. These are the potential relation. That's what Pratik also mentioned, right? That you Pratik, I think Pratik only thought. Which has, you cannot have like very verbose. You cannot give a free hand to LLM to come up with relations. So depending on your use case, you would limit the number of relation types that you have.

works as, lives in, et cetera, et cetera, et cetera. ⁓ You would not let it go berserk with creating different types of strings every day. You would want to limit it as per your use case. So that once you limit the number of relations, now it's easy for you to craft a Cypher or a graph query to give you the result that you wanted, what's relevant. Because you know that, for example, who am I meeting tomorrow? You have date node, you have a user node, you have a meeting as relation, you can craft a...

query that gives you people you are meeting on this date.

So is that query we are harness creating or is that? ⁓ LLM is creating graph query. LLM is good at creating Cypher queries. ⁓ LLM is good Okay, okay, okay. So we just give it like the context that you have this access to this thing. So you give context like these are my different node types. This is what a node ID looks like. These are the list of relations that I have. ⁓ And now you create a graph query. ⁓ Got it, got it. And just one more question. Before that, you were covering this summarization, right? Mm-hmm. ⁓

I think every five steps or every five turns you're creating a summarized thing. So in that step, ⁓ in the code, are you like evicting five turns? In eviction, I'm evicting five. And in summarization, I'm removing those five and then adding the summarization. And then summarization goes as like a role user itself? Like say that? Role user itself. Role user itself. This is the summarized output I want to continue. So it's not assisted, it's user. Okay. Okay. Thank you. ⁓ Thank you. Suryansh.

Sneha Mehra (03:14:05)  
Good. Yeah, the same question as ⁓ one more question. ⁓ In this diagram, the executor has a connection to this Postgres. ⁓ Does this also imply it the, ⁓ it is using the rack pipeline that we are building on the runbooks and all of that? ⁓ Yes, yes, yes. So it's memory plus state management. ⁓ I would want, and again, sorry, runbook, sorry, no, not runbook.

I had a tool call for run book to find relevant run books. If you see, have a fourth one, which is query run book index. ⁓ So that tool call goes to vector DP, gets the relevant run books and adds it to the run books. So that tool basically makes use of the rack pipelines to get the correct. ⁓ So again, it would might just do because you would have rack pipeline, which is keeping the vector database updated. Yeah. This tool just queries the vector database. Okay. Okay. ⁓

⁓ Now why? I'll add just one more point. Why? Because you might look for a run book. So there your prompt is very important. Like what kind of run books to expect? ⁓ Does this run book is about like even do I have a run book to execute a particular command on my infrastructure because it's an internal tool. So you give them list of run books that you have or types of run books that you have so they can create a corresponding query for it to get to fetch a relevant run book.

Now you can make it as sophisticated. ⁓ You see each one of these directions has its own set of sophistication. Now you can go ⁓ as ⁓ wild as you would like into defining how to find the most relevant run book given this as my correspondence. ⁓ Thanks again. We were good.

Yeah, just wanted to share something, question. Please, please, please. So, ⁓ in session one and two, I was thinking about ⁓ the whole chain of thought and if I could reproduce that. I looked into it, turned out chain of thought is something else, but then I realized it would form a workflow if I try to mimic it. And I tried doing that ⁓ and I started thinking about

Sneha Mehra (03:16:22)  
how long the loop will go and all those things. And it perfectly overlap with the two sessions we have had about the loops in previous session, ⁓ Ralph and everything and the memory management. ⁓ So that sort of helped a lot, even though was like COT is something different, ⁓ but mimicking, trying to mimic it actually helped a lot. ⁓ probably that is something. ⁓ Perfect. ⁓ We're going to sort up. ⁓ Go ahead, sort up.

Yeah, so you mentioned that we cannot rely on one LLM and one provider, right? ⁓ we have to use but earlier you also mentioned like the output changes from model to model. So how will this? Yes. So ⁓ there is output does change model to model, ⁓ but there are two ways to do it. Either you have separate agent ⁓ that

is having that crafted query ⁓ for a particular model. That's what we do. So different prompts for different models. That's what we discussed in first session when Pratik mentioned about ⁓ prompt registry that you would have, right? So, or a prompt repository that you would have in which one of the columns could be your model. That this is my prompt template, this is my prompt variables, is meant for this model ⁓ below or above this particular version.

because we have tested it, it works fine for that. ⁓ So that is one way to do it, where you have different prompts written for different ⁓ agents. ⁓ Otherwise, if you have the right set of evals, then you say, can I write one prompt that works for both the models? If you'd want to go that way, then you would want to take, ⁓ then you need to make sure that you are not.

degrading output of one just to have just one from deal with it. ⁓ Prompt repository is a better solution to deal with this problem versus writing one from that works on all models. If you can, if it's simple enough, all good. But if you say, you know, yeah, like I am ⁓ using Claude versus Gemini, Claude and Gemini. Gemini expects very different type of prompts. You have to be very explicit with Gemini and you can be vague with Claude. So what works at one word doesn't work at other.

Sneha Mehra (03:18:44)  
Worst case, I say Cloud Agent SDK or any Agent SDK, skill files, how are these skill files supported in those Agent SDK run loop? ⁓ Cloud Agent SDK supports, but OpenAI ⁓ SDK doesn't support yet. Skills, the notion of skills. ⁓ That changes the For different models at the same time, like when we are creating an application or then we have to test it with multiple models and keep it ready.

Yes. Now how ⁓ deep you'd want to go, see now this is also a hole. Because there is a point of diminishing return where you know that I'm not going to go away from ⁓ Claude at all and I'm okay taking up this dependency that if Claude is down, I'm okay being down because my on-call is anyway going to wake up and fix it. Then why put that effort? ⁓ Right? So you have to do this benefit analysis. ⁓ But again, like going back to the first session where ⁓ I spoke about like, you you can

still use, for example, say, log or chat, ⁓ GPT models, right? ⁓ But they could be different providers ⁓ in sense when Azure is providing and AWS is providing, right? In that sense, you still have ⁓ that ⁓ capability, right? ⁓ So even if one Azure goes down, AWS is up, right? ⁓ So you could still use the same model ⁓

Okay. But different. ⁓ Like bedrock bedrock solves this problem. Okay. Pratik, ahead. Sorry. Sorry. have one question. So ⁓ you mentioned like if we give the long prompt, then it takes the initial part and the last part. And then today you also mentioned like if we are asking any question and if we will give the same prompt twice.

It will actually give the better result and it will enforce it to go through the middle content as well. But ⁓ is there any other or better way like alternate way to deal with this? Like how to overcome this?

Sneha Mehra (03:20:54)  
No, but again, the one thing that I mentioned, is prompt, concurrently with the same prompt again, that just doubles up your input token utilization. ⁓ Are you okay with that? ⁓ But lost in the middle problem, that's why evals come in very handy. Like how big of a context it to provide in a model respects what is supposed to respect. ⁓ So we just said, even you for this event, there's incident remediation. You would have evals written for it like, Hey,

Given these are my sample run books, it should output these string. If it doesn't output, then something's wrong. ⁓ Then you iterate on that prompt until your eval succeed. If that eval succeeds, then you assume that even there is a loss in the middle problem, it won't affect you. So evals are like unit test. ⁓ If evals work, you can be very certain depending on the quality of your evals. course, you can be fairly certain that this would work just fine. ⁓

Thank you Pratik. I wanted to go to ⁓ the initial sections rolling summarization hierarchical memory. ⁓

Sneha Mehra (03:22:11)  
⁓ Yeah. ⁓ Yes. So ⁓ in this case, ⁓ when we say rolling summarization, so every end turn replace with summary. If you think about it ⁓ in cases of ⁓

Like chat applications, let's assume. ⁓ So what happens in case of chat application is that your API receives ⁓ one request at a time, right? So someone put in a request, your API gets it, it creates a agent response, sends it back, done. So now every end turn becomes difficult because you might not really have the notion across API. ⁓

rolling summarization can only happen in case of like long running agents or coding agents type of scenarios. ⁓ So ⁓ that is one ⁓ limitation that I see here. ⁓ The second is in case of hierarchical memory consolidation. My question was like, ⁓ goes back to how we define ⁓ the graph and the relationships, how we define the entities here.

So this will be very use case specific. Now here, like for example, if you're doing Google Meet transcription and actionable, you then you exactly know that I want to gather evidences of key decisions being made, who said what, next steps that people recommended as separate section. ⁓ So it's very use case specific, but that this is not graph. This literally structured data extracted out of it, which will then be merged because for each 10 minutes I'm processing this gathering.

Trial processing gathering and then I consolidated at the end. ⁓ Got it. And weighted memory was? Can you put on? ⁓ That's it. Sorry. That's the last one. ⁓ Weighted retention.

Sneha Mehra (03:24:16)  
yeah, this one, ⁓ I think it is there with everything like any memory that you are storing has to be weighted by default. ⁓ not by default. are like memories can be categorized into facts and preferences. Facts remain as is. ⁓ Facts stay as is and preferences are the thing that always have a weightage that will go with or have a decay.

Yeah, New Delhi Capital of India cannot have be forgotten. ⁓ It's a fact, ⁓ not opinion or preference. ⁓ Correct. ⁓ So, and then there is also this other type of memory called procedural. You covered it, but you did call it procedural, which is in the system design when we talked about agent remembering what it did such that it can recommend better that is procedural. So, I knew it had a name. ⁓ I literally did not know it had a name. Okay. ⁓

like naming convention wise there are like so there is episodic, ⁓ there is semantic, ⁓ there is procedural and then there is short term. ⁓ So episodic is basically when you can create a memory and tie it to when and when it happened basically so you can reference back to this was the situation. Semantic is you can take multiple episodes and then extract ⁓ facts sort of not facts but ⁓

⁓ commonly occurring themes around it. So let's assume a user had had five conversations with me about booking a hotel, and every single time they have mentioned that they want a hotel with swimming pool. So I can extract that as a preference of the user that they would want to have swimming pool. So that becomes a semantic memory. Within the same conversation, it could also be an episodic memory linked to that conversation. ⁓ So got it. So semantic is more of an aggregation over multiple episodic memories.

⁓ So semantic is more powerful in the sense that episodic can happen once and you can forget about it. Like it's never mentioned again. But if the same thing is repeated multiple times, it gives me a stronger signal. that's where episodic comes and plays a role. then procedural is basically for self-improvement of agents. what actions agent took ⁓ and if it did something wrong, it's like a good behavior. Yes. Getting reinforced. ⁓

Sneha Mehra (03:26:39)  
That is procedure. ⁓ we have defined a procedure and I kind of can choose to stick to it is a good behavior exam. Yes. Remember. Thanks. Akash good. I am a bit off topic. ⁓ you created a video on how to build a second brain in Obsidian. So how do you ⁓ like use it now with LLMs? ⁓ Gemini. I have created in Gemini. That's why I use Gemini. But right now

It still has this like nothing changed. ⁓ But I have now LLM tools that I built that helps me summarize my learnings, my notes and create slightly detailed ⁓ notes and adds to my second brain. That's the only thing that has changed. So I have now my own CLI tool called brain. ⁓ I can show a demo. I can show a demo but it's very simple tool that I built which does this. Hey, I have Twitter open.

Okay, so this is what the tool looks like.

Sneha Mehra (03:27:48)  
So I do brain add and I can do thought. So all the social media posts that I have, it's a brain add thought, D-H-O-U-G-H-P. So here what it does is ⁓ I can add a thought that I have, which is essentially my social media post. I'm calling all my social media posts as thought. And then it classifies it into stuff and it adds it. So for example, all my thoughts that I see over here.

⁓ So all these are thoughts on AI that I posted. These are literally my social media posts that I wrote. It goes over here. These are my career growth related thoughts that I made. It goes over here. Then I have all my funny posts that I made. Goes over here. So every single time when I write something, ⁓ it ⁓ creates tags, everything. I don't use tags for navigation. Just did it because I could. ⁓ Now then all my blogs are here. Thoughts is the major one. All my talks are here.

Now what I do ⁓ is this is one thing. So I have my hardcoded text that I'm adding. I can also do brain add URL and then what it expects me to pass is a URL. Now this URL could be a website URL. So let's say I go news.hackerjose.com ⁓ and I say fancy word. Let's click on this.

⁓ I assume I call this only when I know I have already learnt it. ⁓

I have not done it. If I'm not ready, so I'll call ⁓ add URL instead of say brain add run. I said brain learn from this URL. So then what it does is it does the same exact step, but it shows me this information on my terminal for me to learn from it. So for example, it just give it a minute. So this is when most of the stuff I do like this brain learn. And then I pass in the URL. This URL could be YouTube URL, website URL, whatever, whatever, whatever.

Sneha Mehra (03:30:08)  
It goes and fetches it. I have not started reading a lot of stuff in markdown format on my terminal. Because terminal renders markdown very beautifully. So let it fetch, let it fetch. Now it fetches, now it process. It process, it extracts bunch of information that I want. And then it renders it nicely over here. When I call add, it literally adds it to my knowledge base. My knowledge base is here. Where I have ⁓ notes here, notes. So there all the stuff that I read and notes from it. So my notes.

and this notes combined goes over here and this is LLM part. So this is mostly for my recall purposes, right? Because I've already learned it. I already know what this is going to talk about. So I can refer to this and like read again if I want to. ⁓ That's all, right? ⁓ Let it process. It will take some time. So now here is the output. So on the left side, this is a little actual blog post written in markdown format on my terminal. ⁓ Because now

all of my articles that I am reading looks the same. Less cognitive overhead. I don't want to remember different styling of different websites and different format style of different websites. And on right side I see summary, ⁓ I see key concepts, see ideas worth writing about, insights, maxims, ⁓ analogies and mental models if I want to form. Again, I don't go through all of them, but some of them I go through. But more importantly, I do go through this blog. From this document ⁓ is what I go through.

And this is my learning flow. ⁓ Thanks a ⁓ lot. But this is the only thing now when I want to hunt for a when I was applying for at a particular conference and I want to hunt for it. So this brain utility has a settings file. So ⁓ .brain slash config no settings is what is called brain slash rejected topics. ⁓ So ⁓ there are some of the rejected topics that I never want to talk about.

So, ⁓ hey, title, ⁓ no, these are not rejectable topics. Where is my brain setting, wait. Brain, hey, sorry, I don't know where that file, ⁓ let me do LS. Hey, then how is it adding? There was supposed to be a file called config.yaml, sorry, config.json, which had the topics that I'm interested in. I will need to find where that file is. But, how can you bring?

Sneha Mehra (03:32:35)  
There was a file which had the topics that I'm interested in so that when I reading a topic it would flag that hey this is something that you are not interested in. Let's say CSS giving an example. They'll just flag that stuff that this article is about CSS you will never want to read about this article move on. So this is what it does. And then I have one more command called read. So what I do is I read books on terminal.

So I three books and I maintain state management over here that which book I would want to read and it resumes from the point. All this data is maintained on S3. I built a database called S4DB to store and that's where S4DB idea came from. S4DB. So it's a simple embedded database in Python that helps me. ⁓ That helps me that does a key value store ⁓ that does a key value lookup on S3. ⁓ All the state goes to S3.

and it's an embedded database. ⁓ I just have to point to bucket and my prefix and then it stores. ⁓ it's a simple key value database. But it helps me put the data and get the data directly from S3. It has a local copy, not entire copy of it, but if it is cash, it doesn't go to S3 and it helps me save some money. So I built this database to power my brain. ⁓ So this brain is all obsidian.

But just the read part of it goes to S3 because I have to store each page of the book separately and to keep track till what page have I read. So that goes to S3. ⁓ This is what my flow is. Thanks for asking. I got to revisit. Now I need to find where the setting file went, but my code is still working. ⁓ I don't know, AI wrote it. ⁓ I don't know. ⁓ Searching. this utility provided or you have created that? Like, is it kind of alias?

⁓ brain is my software ⁓ I built that software again I promted LLM I promted Gemini to build lot of stuff but I built it for my needs ⁓ everybody should have their own software I have one for my daughter like that is super cool super cool like I can show you ⁓ very recently I hosted it online but

Sneha Mehra (03:34:50)  
That is what I do with my daughter. So I built something called as Rumi because my daughter's favorite character is Rumi because Demon Hunters rocks. ⁓ She's there. Right, so here I have parent login and a child login. So what I do is I have topics for my daughter. These are all Claude and Jaina generated. So I create topics and for each topic I create questions. So my daughter is five and a half. So I create, like I'm pushing her to do maths more.

and so she does word problem really well now. So I define what kind of word problems I want to create. So this is like literal prompt, ⁓ a single variable equation word problem, a simple word problem suitable for a child that can be translated into a one variable linear equation. The equation should involve basic operations. ⁓ This is, this is, this. ⁓ It should look like this output, right? ⁓ Now imagine I have this prompt and what I can do is I can create a notebook. So this is a notebook for UKG for her. I create a new quiz.

I select what type of questions I need, how many questions for each type. So let's say I want seven questions of this and eight questions of this. That's a generic quiz. So I picked two types of questions. It created quiz for me. Just 15 questions that will wrap up quickly. ⁓ And so this is a parent login. Now I have a child login as well, where ⁓ she sees active quiz. ⁓ And then from that quiz,

She can give quiz on her iPad if she wants to but she loves using this blackboard. So I keep telling her questions and she keeps solving it. So this is our bonding moment. Sort of stuff. Hey fail to generate. Hey was it 2.0 ⁓ shit. Very likely 2.0 on Wait let me check. But I can show demo on different stuff if I want to. Let's see.

Wait, I'll go to different questions. So now imagine this is an existing quiz, right? So these are like 10, 12 questions. Now what I can do is I can do print PDF. So it outputs a very nice PDF format. I can take printout, give it to her in the morning by evening she is done with the questions. So she gets this motivation done, done, done, done, done. And then Papa will give me star and sticker and this and that. Right? So I do this and again, this is simple HTML. HTML page did a control P event using JavaScript. Nothing fancy.

Sneha Mehra (03:37:12)  
no pdf parser nothing. asked it to output html. Given I know that I can just do an HTML output and do a control p on it. Again the idea is you need to know otherwise LLM will go in very random direction. ⁓ then 14 June. It ⁓ just created this quiz 14 June today. ⁓ So now it also has latex.

I used to have ⁓ an over here which is to explain I could do regenerate. So this way I had again it ⁓ I used to store this data on ⁓ Postgres. It ran a stupid migration deleted all the data. I had more than 70 quizzes that my daughter solved out of this which included basic addition, single digit addition, single digit subtraction, double digit addition, four digit addition, multiplication, division. ⁓ Then we started equations, single variable equations.

Then word problems. Now next up is, now she did word problems with addition. Now addition subtraction. Now she'll do word problems with multiplication. And then we did science facts. Out of this, where I asked her to do science questions. So rather than doing this, I asked her speed, distance, time questions. She relates to it. Then we measure diameter of earth. So she ⁓ then knows how big the earth is and like how, like how, like how small we are in the universe. So.

This is my doing it, I lost a lot of data out of this. But now I won't because it's DynamoDB. when I blocked ⁓ the APEC restriction is on deletion. It cannot delete anything. But yeah, this is another tool that I built. This is why I love AI. Like this was not possible. It would have taken me ages to build this. I have something very similar. ⁓ I was onwarding to this new team, right? And they gave me 20-25 documents. I was like, fuck this. ⁓ So I built an onwarding buddy.

which basically created milestones for me ⁓ and has a test for me at end of each milestone to test my knowledge. ⁓ On body body that's a good one. This very good idea. That quiz part is very interesting. Do you really understand it? Because it has all the corpus of the knowledge. It can ask me the quiz. So I just use it enough. ⁓ Every time it generates multiple choice questions.

Sneha Mehra (03:39:36)  
and as ⁓ and done. Yes, that's that's ⁓ that's how and most people don't know what to do with AI like they have super power you could do so much stuff but again people are bad at identifying problems like that's my grudge people should be good at identifying problems like these are the problems worth solving like if I just made I could make it like not open source but I can just offer it as a login I added because Claude gave me the way like because I just a Claude prompt for me to add I did not even have login

It was there but now I had to host it because I was running it locally on docker so I had to start my computer and do this and let it be on cloud. ⁓ Then I was going to add this one very cool feature which is my daughter loves when the iPad rings and she has a fact to read because she is just 5 and a half so she just started reading and she loves reading facts. So what I have done is I built a pusher channel and I let Ellen generate a quiz depending on the history doesn't repeat.

⁓ I played it once when I was out in office. ⁓

and then I was like she reading so happily yes this is what I wanted so just making sure she remains curious throughout her life ⁓ and it's she is my neural network did you ⁓ so this Rumi app like did you do you open it up or why why why should I open it up people can build their own software stuff see I was going to ⁓

⁓ I was going to ⁓ but the effort of opening up number one second why are you pressurizing your five-year-old to do this ⁓ I don't want to fall into that ⁓ let me in it's between me and my daughter right if she is enjoying word problems what I learnt in sixth standard ⁓ at age of six I don't care ⁓ the moment the day she cries about it I'll stop doing it

Sneha Mehra (03:42:02)  
But if she is happy, ⁓ I see that joy. But people don't understand. ⁓ no, back to back. ⁓ Why do do it? But I see, in cohort I show, ⁓ and you go and build it. ⁓ It has done wonders to my daughter. ⁓ She has built that intuition, ⁓ that love for math, that love for science. ⁓ The other thing I am struggling with is screen time with my kids. Therefore, I think...

⁓ When did you introduce Screen Time? ⁓ One year old. I gave her phone. ⁓ I'll tell you my reasoning behind it. My reasoning behind it is when my parents were forcing me to not watch TV, I used to watch it more. The moment they stopped forcing, the moment they said go watch it, I got bored of watching it.

So now I'm seeing the same trait in my daughter because what I realized my daughter is carbon copy of me in terms of behavior. So what worked for me is working for her. So I literally gave her that Jake, Joe, whatever you want to watch, watch. She gets bored after 15 minutes and she comes with Papa lets all maths. Sometimes she's in mood, she watches for two hours. But after that, ⁓ after that watching for two hours, when I asked her, hey, let's study, she yes, dad, was anyway getting bored. Let's study.

So she might need a trigger point. But because that's exactly how I behaved when I was a kid, I'm ⁓ just using the positive and negative reinforcement that I received into positive reinforcement for my daughter. Like finding ways to do it. But ⁓ screen time there is no restriction from my side on her. ⁓ As long as she does not, I guess if she herself is getting bored and wants to do it, then it works out.

But even if you look at it, I look at it this way that watching TV or having a screen time is a positive reinforcement for her given if she solving 25 to 40 questions a day. Yeah, but again depends like what type of screen time, right? I mean, it is just like ⁓ TikTok videos which are just like short for ⁓ YouTube. YouTube, YouTube, YouTube. She was YouTube 6, 7, ⁓ all of that stuff. This Rumi is what? K-pop D1 Hunter, bro. ⁓ I...

Sneha Mehra (03:44:23)  
I have memorized all the kpop demon hunter songs now. Like I get excited when I see someone playing demon hunters like hey Ruby Ruby I like Zoe more but yeah Ruby Ruby I'll just start screaming and like why that? Also one more reason sorry for this parenting pep talk but one more reason that when she sees me being interested she is so proud of it that my dad is interested in like girl stories ⁓ and like every time that song plays doom doom doom voila

how it's done done done she comes running to me and says papa your favorite song is here and like that is important like for me I can watch all sorts of brain rot stuff that she is watching like Mikey and JJ and Minecraft videos and all just for that one moment that I get every day it's ⁓ it, I'm happy, doesn't matter but again she does 25 to 40 questions a day her maximum is 125 in a day she was in a fab mood that day

125 questions, she studied for 4 and a half hours non-stop. Okay, this was like addition subtraction, but all in her head. She was in that mood. I was giving like ⁓ a single digit, double digit addition and subtraction. Subtraction without borrow, but addition with carry. ⁓ And she was doing it in her head. She was in that zone that day. ⁓ I ⁓ that day. ⁓ I have like 50 photos I clicked of her that day because she was on fire.

4 and a half, 5 hours non-stop we studied. ⁓ This entire blackboard became white that day. ⁓ Because every time she was storing 5, Papa give me a star. Every time she stored 5, Papa give me a star. ⁓ I was there, I made stars that day. Full stars. ⁓ My best investor. ⁓ Best 600 rupees ever spent. ⁓ Ever. ⁓ Best 600 rupees. ⁓ I just, ⁓ the day she was born.

The day we moved over here within a week I ordered this wall sticker that converts a wall into a blackboard. Two stickers, stuck it in my room. ⁓ Best decision ever.

Sneha Mehra (03:46:30)  
It's so much I do so much fun on that white board, ⁓ that black board. But the way I also put it is I'm just raising my co-founder. She's my neural network, I'm training her. And I'm training her to be my co-founder in the future. Like what I wished I had as a co-founder, how I would want to be a founder because education system has zero trust. So what do I do? Raising a co-founder. ⁓ Fallback plan, she'll become an engineer like me. But best case scenario?

After 10th, start up with me. ⁓ Till then, we make ⁓ We will

It's experiment. ⁓ got very much inspired by Polgar. If you have read about him. Polgar was chess grandmaster and he had a belief that geniuses are not born, they are made. He had three daughters. all three grandmasters. ⁓ What is Polgar? Polgar is the surname of a person who was a grandmaster. Chess grandmaster. So P-O-L-G-A-R. I read about him about 8 or 9 years ago.

And it stayed with me. It's not a forcing function for me, but it stayed with me. That geniuses are not born, they are made. And like, let me try. Like, if my daughter reciprocates, then why not? I'm trying. I'll see where it goes. It's an experiment. Also the YouTube, she also recorded two YouTube videos. She sees me doing YouTube videos. She said, Papa, I also want to do it. Because she sees whenever I step out.

One or two people come and we here pick fan there is that they click photo. She gets jealous. ⁓ And she asked me Papa how come everybody comes to you and not mama? ⁓ Papa is famous. Papa I also want to become famous when I am 20\. What do I do? Papa makes YouTube videos. Papa can I make YouTube videos? We have a video recorded before where she was explaining ⁓ prism effect. She was explaining ⁓ life that entire water cycle thing.

Sneha Mehra (03:48:33)  
She can explain prism effect really well. she kind of mugged it up initially, but now she can understand it. Like why that like splits and how it splits and like has wavelengths. I literally had a demo of sound is physical sort of stuff. So she tries to explain it and then how plants grow like that entire thing. ⁓ We have three, three videos recorded and like I still don't want to post it, but it's she gets happy when I play it for her. She thinks it's live on the-

Wait for two years people will come to greet you as well. ⁓ So that's an incentive. ⁓ You'll be stupid and I'll be stupid too. ⁓ It's working out. ⁓ Aid your random stuff. ⁓ Again, I see that same traits that are there in me. Like she doesn't look like me. But those are traits in me. So what worked for me, that external validation and that lack of confidence and... Everything is copy. ⁓ I'm copy So I'm happy.

one way of course but he. ⁓ Chalo, ⁓ awesome sorry this always happens but use LLM to build stuff like find problems worth solving that you think is problematic like onboarding buddy Pratik build do it when you become parent or if you are a parent try to build stuff like this and like just get your kid excited and one thing sorry one more thing I made her call AI as AI auntie

And then she gets a personal attachment ⁓ with AI auntie ⁓ and she wipes code sometimes. I asked her, what do want it to build? So ⁓ I let her imagination go wild. When that Gemini and Nano Banana came in, gave free to everyone. ⁓ made her imagine, imagine what you want to imagine. So she imagined, Papa, what if there is an animal with body like lion and like face like giraffe? I asked AI auntie to build it.

Papa this is it would look. I wait now I will show you something. What if there is an animal called TORSQUITO. ⁓ Which is tortoise and a mosquitos combination. ⁓ And then she started imitating me. Then she started merging multiple animals. Then I had to nudge him. That this more positive reinforcement. ⁓ Then I gave a different example. So then it went in that direction. And then I made her wipe code a nap. I what kind of game you would want to play? Then she said Papa I want to play a game where I can like jump and swim both. I said where should it?

Sneha Mehra (03:50:59)  
I asked her to build it. We called it Claude Opus 4\. ⁓ whatever. ⁓ I just bought Claude subscription that day. I asked her to build it. Papa can can A \&T do this? ⁓ I don't know. She asked. ⁓ And the best part is AI understands child tone. ⁓ That was surprising. ⁓ It does a very good job at it. ⁓ And then she understood what to say and what not to say. During second iteration it became much better. ⁓ So she's like prompt engineer now.

She writes better prompts. Hey, make this, make it look beautiful. Not round, square I want. I am giving full exposure to tech. I ⁓ am not restraining. She uses calculator. ⁓ So, like full scientific calculator. Like, I don't want her to do mental maths. Like division.

Put it in, do 21 divided by 6\. She knows 6 is 24\. ⁓ Why waste time? ⁓ Because if she can articulate text to equations, good enough. There is computer Now the AI is there, we are saying we are prompt engineers. The whole thing is intuition and problem solving. I focusing on that. Everyone's thesis is running away from here to there. But yes, that's situation. Let's go.

Awesome\! Anything, Everyone sleepy, Go to He's gone. Charlam Okay, awesome. Thanks ⁓ for the time, Thank Bye. See you folks next week. I'll share the periods and questions sometime. Bye. ⁓

