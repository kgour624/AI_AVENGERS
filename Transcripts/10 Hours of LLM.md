4  
Sneha Mehra (00:00:09)  
Welcome to the fourth session of this course. Let's dive right in. ⁓ So, today we talk about workflows and probably the most hyped term ⁓ this year: agents. ⁓ So, what's the plan for today? ⁓ We will start with workflows and transition towards agents. And mostly explain why agents are actually relevant and not just hype, but also why they are not useful for anything. We'll then touch on

Reasoning models, finally, deep research, popular tools, and also discuss what we think is very important for the future, ⁓ MCP, and some other things like the A to A protocol by Microsoft and a very useful case study that Omar will go through.

So why is 2025 the year of agents? Tons of CEOs of important companies like Demis Hassabis, the CEO of DeepMind, and also the CEO of Microsoft, or even Greg Brockman from OpenAI, all say that this year or 2025 is the year of agents. Why is that? Well, that's what we will see in this course. But first, why are we seeing agents come out?

Well, one of the main reasons is that LLMs aren't enough. We sometimes want to augment it. ⁓ First, an LLM is just text in, text out. You ask a question, you get a reply. But sometimes, as we've seen, you need additional data. You need to add some kind of retrieval component to give either updated information or access to information it wasn't random, like company specific private information.

You may also want to use different tools or ⁓ allow it to do some other things than just generate text. ⁓ Whether it is to use a calculator or to run some code or query a database, whatever, you may want to give it access to some other things than generate tokens and generate text. And lastly, you may want to add a memory to the model, just because oftentimes just asking a question and getting a response isn't enough. The user will

Sneha Mehra (00:02:25)  
Keep asking questions and also come back later on. So you usually want a long-term memory and a short-term memory, which is something LLMs don't have ⁓ naturally. And likewise, ChatGPT has all those that they implemented using engineering principles. And it's not just the models like GPT-4.0 that has these things embedded. They actually have to implement them inside ChatGPT.

Which you also have to implement inside your own applications. And just to keep in mind, this is not an agent. It's actually a workflow. And it can be even more complex than that. You can have multiple queries and you can deal with multiple queries from users, which we call prompt chaining. And it's quite useful when you have a too challenging task that you want to break down into simpler cases. Still, this is not.

Agentic. We reduce the latency of a complex task by chaining the task in clear successive subtasks. It's not something agentic yet. It's still a workflow. Also, ⁓ sometimes just chaining them isn't enough. You may want to ⁓ decide on which task to do next. And in that case, we often do that using what we call routers ⁓ or an orchestrator.

Which is basically just another LLM call where you give it a prompt with specific ⁓ if-else case statement statements where you say, for example, if the user asks about ⁓ something in your enterprise, query this database, ⁓ else if it ⁓ asks about running this kind of query, ⁓ use the code interpreter, whatever, it will be ⁓ hard-coded or hard-prompted into

The router, which will itself select the right LLM call to do next. And here, for example, it can be also to ⁓ use a smaller model for easy task or classify results with or without an LLM. For example, sentiment analysis could be done without an LLM. It could be ⁓ to perform intermediate tasks, as I said, to ⁓ query a database or run code.

Sneha Mehra (00:04:46)  
It could be to use a fine-tuned model for a specific case, and another fine-tuned model for another specific case can be lots of things. ⁓ We can also do them in parallel ⁓ to be more efficient ⁓ or ⁓ to improve results with ⁓ something like majority voting, which is basically to, ⁓ for example, ask the same query to ⁓ four LLMs and take ⁓ the and combine the results to either take the the components that

Are repeated in the ⁓ multiple results, or just to have a better answer as a whole. And then you aggregate the results, or you combine, or you do whatever to try to improve the results. This is still a workflow. It's not something agentic or something magical. It's something we can hard code and use LLMs to help us do that. Even more, we can add loops to that. We can have an LLM to generate an answer, then have an evaluator or a judge.

as we thought in the the last course, to evaluate this generation, give feedback, reject it, and iterate until we have an accepted answer that is better than usual. This is similar to our LLM as judge, and ⁓ sometimes we call that LLMs in a loop. ⁓ It's very useful when we when we can evaluate what we want with like for example the checklist that we referred to in the last

Course or have measurable outputs ⁓ and ideally have human feedback to improve the LLM and to improve our judge. Still, this can become a quite advanced workflow with ⁓ loops, with tasks executed in parallel, with prompt training, with tools, memory, but it is all still a workflow, even though it can be super advanced. So, what makes an agent an agent?

Well, we have several points that we need to have in order to be an agent. First, it needs to be autonomous. It needs to decide what it needs to do to answer the query. And it does that with a plan. So it just like the deep research, if you've used it, ⁓ it starts by thinking okay, what this is the question of the user. What should I do next? And how can I answer it? Like, for example, I need to research on the internet.

Sneha Mehra (00:07:12)  
Then based on what I find on the internet, I need to do that. Then based on that, ⁓ I need to do this, etc. So that comes to the third point that it needs to react according to the environment. So for example, if it finds what it needs on the internet or not, it will do different things. This is what we would call an agent. ⁓ It's basically something autonomous that will do what it needs to answer the query. ⁓ So it itself decides.

If it uses tools or not, ⁓ how to respawn, which LLM to use, because it can be implemented with various LLMs. So as is shown in this graph here, ⁓ it does actions in its environment, gets feedback from the environment, whether it is the the links, the summaries of the internet pages that it searched for, or whatever, it gets feedback and then adapts based on this feedback from the environment. And that's the key.

Architectural pattern. ⁓ We have a plan, validate the plan with either heuristics, human feedback, AI judges, whatever, and then execute it. That's what we would call an agent. And just to mention, to avoid super costly or incorrect actions based on a bad plan, ⁓ we want a human in the loop. But we don't really want to use agents at our cost.

Unlike ⁓ most companies are trying to sell you. Instead, we always want to use the simplest possible solution, ⁓ such as following the process we've seen ⁓ in the second course, starting with the simplest thing, prompt engineering, just trying it out, then implementing a rag system, embedding models, fine-tuning, etc. ⁓ And ⁓ agents come into consideration when we need more than what a workflow can bring us.

When we need more variability between tools, when we need autonomous decision making, this is a strong suit for agents. When we have a very complex problem that we cannot build predetermined workflow, but ⁓ this is the caveat and the problem, is that this ⁓ problem that we have needs to be not too complex either, because otherwise ⁓ the even the agent cannot ⁓ fix it and you will end up

Sneha Mehra (00:09:35)  
Spending a lot of money trying to fix the issue that is just simply too complex for any LLM to do that yet. ⁓ It's also ⁓ into consideration when you have enough budget ⁓ and ⁓ when we don't know the style of response that we want. So something that is much more adaptable to the environment to the user. Here's a a checklist ⁓ that I I found on an interesting entropic talk.

That we linked in the show notes, where basically Enthropic says that you should build an agent if the task ⁓ is ⁓ complex enough, if it's valuable enough, that's the most important part. Because if the task is super complex but doesn't really get you enough money, well the agent costs a lot. So you definitely won't get back on what you spend because agents ⁓ build a plan, execute the plan, iterate on loops, etc. ⁓

In short, they just spend a lot of tokens and so a lot of money. You also want to ensure that all of the tasks, ⁓ when split into subtasks, ⁓ are doable. Because otherwise, it may just be a too challenging problem and you may want to ⁓ reduce the scope. And lastly, you want to try agents if the cost of error is low. Because ⁓ since you will be giving agents ⁓ a decision-making power.

You don't want its decisions to ⁓ heavily impact the company or anyone at all. So if the cost of discovery is low, such as, for example, as we will see the deep research agent, the goal here is just to produce a report. And so if it does an error, either through an internet search or a summarization, it's not really ⁓ a big deal. It's just a report. There's no ⁓ life or death decision or whatever to

important in a monetary sense.

Sneha Mehra (00:11:35)  
So we talked about ⁓ workflows to agents and when to use them or not use them, but we haven't defined them yet. So what is an agent? Well, here it is, according to the AI community. It's something that has an intention, it has goals and must act towards that goal. ⁓ Then it has delegated authority, so it can act on our behalfs.

It has a long-term memory, as we mentioned, so it can interact with the world. And for example, if it searches on the internet and has cookies that pop up or a password that has to log in, it needs to remember that in the future. Another point is that the LLM has a flow control. It will decide on the flow of the application and it will be hard-coded as in a workflow. ⁓ An agent is a multi-step planner. It will

As we said, make a plan, act on it, and then edit the plan and follow the plan. It can perform non-trivial multi-step operations that basically previously would have required either some very complex hard-coded workflow in your application or human interaction. Here it will be done by an agent. And lastly, LLMs are tool users. An agent needs tools. It cannot just

Generate text, else it won't do anything. It needs internet access, calculators, code executors, etc. Needs tools. Here's just another definition of agents, this one from OpenAI, where they say that basically an agent is a model, so whatever model, instructions like the just the system prompts and prompts in general, tools and ⁓ runtime, which is basically the environment.

So it's it all comes together. ⁓ we agree that these are agents. And so let's dive into some quick examples of what are real agents to give you an idea if you should build one or another. The first example would be Cloud Code, which is a very good example because it has access to tools like Bash, Read, Write. It has an environment, it evolves in the terminal of your computer, and it has

Sneha Mehra (00:13:56)  
system prompts, which is its goal of creating new features, fixing GitHub issues, whatever. And how does it work? Well, it works by having a human or yourself, a user, asking a question. Then the LLM refines the question, taking clarifications from you. ⁓ Once this is done, it acts with its environment to search for files, ⁓ do tests, run code, etc. And once it's complete, it gets back.

To the human with ⁓ displaying the results. ⁓ I just want to note that this is an agent and it's a good example of an agent, but ⁓ we very more often need workflows. ⁓ And and why I say that it's a good example of an agent is because of multiple reasons from the same presentation from Entropic that we see the link below. Four reasons here, because ⁓ the complexity is there.

It's very hard to fix an issue and do a pull request fixing that issue. ⁓ The value is very high because you are basically replacing developers that need a lot of money, or ⁓ a developer is using it to save time and be more efficient. So either way, it can save a lot of money. Then ⁓ Cloud is great at coding, so it makes it viable, or other models are great at coding. So this is a use case that has potential.

And lastly, the cost of error is pretty low because you have unit tests in place and other systems to ensure that all the code isn't deployed right in your application. You have checks in place to double check and to ensure everything's fine. Another great example ⁓ is computer use, where the agent has access to a browser with clicking possibilities ⁓ or using the keyboards, ⁓ taking screenshots, etc. ⁓

Its environments would be the computer interface, and the goal, the prompt, would be to do some task for us, such as, for example, asking to install Microsoft Word, and it would go on the internet, try to do that, locate things, ⁓ react to the environment if there's a pop-up or whatever, close it. And so that's a true Agentic behavior. Another one would be what we said in the second course with RAG, or what we call agentic rag.

Sneha Mehra (00:16:23)  
Which would be a whole advanced rag system that is autonomous to use database or not, code interpreter or not, ⁓ and ⁓ act in its ⁓ own ⁓ rag environment to accomplish system prompts like answering the user ⁓ correctly. And lastly, we have deep research, which we'll cover a bit more in depth in the ⁓ in a few slides, because we find this one very interesting. That is a true agent because it does.

web search on Google on the internet, ⁓ uses the keyboard and has a system prompt, which would be to do market research on something or produce a big report on something, depending on what the user wants. Okay, so those are examples. But why are we speaking about agents now and why 2025? For a few reasons. Firstly, because models are becoming more powerful than ever.

But we also see some kind of saturation on the popular benchmarks. All models ⁓ are becoming quite powerful, and the new versions don't really hit that hard as it used to do. Another reason is that there are many models. ⁓ And actually, most of them are pretty similar and have pretty similar results. They have reached some kind of soft limit for now.

And even though ⁓ the graphic here is small and but it's not really important, it's just to show that even though ⁓ Gemini are the most powerful models, it's very similar to GPD from OpenAI, to Cloud from Enterpic, ⁓ or or Deep Seek, whatever, they all have very similar results, even though some are a bit better than others, but you can interchange them in your application without much trouble.

Even though you need to make some small adjustments, it's just to say that they are more than one offers compared to when OpenAI first started. We also see larger input context, just like Gemini, in the millions of tokens, as we've seen in the first course and second course. Another point is that models are also cheaper and cheaper to use, and that coupled with larger input context, makes agent systems.

Sneha Mehra (00:18:45)  
Super interesting. Tokens are always cheaper, and agents need lots of tokens to provide good results. And similarly, tokens are generated faster and faster, which also makes agents super interesting because we generate lots of tokens and we still want the system to be somewhat efficient. Another great reason is the new reasoning suite of models, ⁓ or what we call test time compute or inference time compute.

These what we call thinking models have improved the results a lot and also helped a lot for the agency system just because they intrinsically think or plan before acting. So it's a very good fit for agents. And likely we have better tools systems like better web search, memory. ⁓ We have MCP, which we'll cover in a few slides, and also we have better use of these tools.

Like structured output, that we now have a hundred percent confidence compared to before that we just hoped to get ⁓ a perfect structured output. But even though this makes it all perfect for agents, they still have limitations and they still cause ⁓ issues and specifically new issues, new failure modes that we just to come back on evaluations also want to evaluate. ⁓ It's still super important, just like in course three.

Evaluate your whole rag system. Here you want to evaluate your agents. You want to check if the plans are good or not, if it makes the good plan before acting, ⁓ if ⁓ it uses the tools correctly, and also if there are efficiency failures. ⁓ If the agent system succeeds but uses way too many steps, is too costly, too slow, you want to check that. And here it's

Almost more important than ever to evaluate because even if you have 95% accuracy or even 99% accuracy, ⁓ the agent uses the LLM ⁓ so much that you will eventually face an error. And so you need to be as close as 100% as you can, or at least have fallbacks in case you have some kind of errors. And we must ⁓ build powerful tests to evaluate this and to check for this.

Sneha Mehra (00:21:09)  
Here are some recommendation metrics to check for all these steps. For example, for planning, you may want to check the rate of valid plans generated for a set of tasks. So just how many plans are good ones. You want to check what is the average number of plans needed ⁓ to ⁓ produce a valid one, and ⁓ of course, get that closer to one. You want to compare the rate of valid versus invalid tool calls in your plan.

And ⁓ similarly, the rate of correct tools called with the wrong parameters or values. Those are all errors that can happen. Then we can analyze tasks with ⁓ where failures are frequent to find the common pattern errors, the tools that seem harder to use, etc. etc. All that to then know what to work on to improve your agent. Just like we talked about in the last course. ⁓

Whole planning process will help us know if we need to fine-tune a model to do something better, to change the tool prompts and the tool descriptions to be used more efficiently, etc. And then ⁓ we also have metrics for efficiency, just like checking the average number of steps to complete a task and see if this goes downward with each update. Check the average cost or time and resources for each task. Check the average delay per action.

And identify the slowest or most costly actions to work on them and improve them. So this is this was just a parenthesis on evaluations, just because it's always crucial, always super relevant. But we were talking about reasoning models and agent system as a whole. And I said that reasoning models were a perfect suit for agent because of how they worked. So just quickly, reasoning models, where are they? If you remember in the

First course we talked about the problem with the query, the number of R's in the word Strawberry, where models, because they use tokens and not letters, had difficulty find counting the numbers of Rs in a word like Strawberry. But if you asked it to first spell the let the word and then count the letters, it will achieve that correctly and even know where the letters are. So we kind of

Sneha Mehra (00:23:37)  
Have to make the plan for them, but then it could ⁓ answer a complex query that ⁓ usually it couldn't. And so that's what reasoning models try to do by itself. For example, here I asked the same question. It tries to reason ⁓ about your query. For example, here it says it reasoned for seven seconds and it makes a plan. ⁓ It tries to understand, okay, the user wants to do that. ⁓ Let me do that.

carefully. ⁓ For example, here it's it itself types each letter to know where the R appeared. It does ⁓ multi-step ⁓ and thinking ⁓ through the ⁓ question before it answers. And so what's the goal? It's to reason ⁓ according to the request. Here we ask how many R's are in Strawberry.

It needed to reason just a bit because it needs to spell the letters of the word and count the number of hours. So it's seven seconds. And here, ⁓ even though it's in English, I asked what's the difference between the training of a reasoning model and the training of a regular model. And it had to reason ⁓ for a minute and two seconds. ⁓ It can be even worse if you ask a very complex ⁓ question, for example, with ⁓

Deepseek, where you can see the whole thinking process compared to OpenAI that hides it and gives you a summary, it can be super long in the ⁓ multiple ⁓ of minutes. And so it uses a huge amount of tokens and thus increases the cost of using models and yeah, of using models by a lot. And so resuming models go beyond prediction. It moves from

Next token prediction, as we described in the first course, to multi-step reasoning. It basically implements the chain of thought prompting that we have talked about in the first course. But it does that at all times. And what's interesting here is that it should do that in an adaptative way, where ⁓ for a simple query, it wouldn't think a lot, and for a complex query, it would produce many thinking steps.

Sneha Mehra (00:25:57)  
It also has advanced capabilities. ⁓ Typically, just because of this thinking process, it can handle complex problems like mathematics programming because it was trained ⁓ with those in mind, because they have measurable final goals. Basically, we train reasoning models ⁓ with, for example, math proofs, just because we have a final answer and a whole proof to train the model to ⁓ produce a good.

Reasoning a good proof before giving the final ⁓ results. Some examples are O3, O1, Deep Seeker 1, etc. And of course, the goal is to maximize the test time compute or inference time compute here. That is basically the ⁓ compute that we do after training. So we may we discussed in the first course that a model is trained on lots of data and requires enormous amount of

Compute, money, and time to train them and to be so powerful, which is why we emphasize on the fact that you should use APIs and not train your own your own models just because it's so expensive. But here, reasoning models also do that process of training. ⁓ And they also use way more tokens when directly exchanging with the user. And that's the process we call ⁓ reasoning.

It's basically just the model generating tokens to help itself generate a final answer later on. And we can see how that can work by understanding the fact that models generate one token at a time and uses each of these generated tokens to provide the next one. So if instead of asking ⁓ what is ⁓ 10 divided by 3 and just ⁓

Asking for the answer directly, if you allow it to do it step by step and generate all intermediate calculations. Obviously, not here, this example is too simple, but if you ask a complex mathematical query, ⁓ if ⁓ for example, I give you a paper and ⁓ let you do it by hand, there's a much better chance for you to give a good answer than just trying to mentally come up with the answer ⁓ very quickly. So that's a bit the same thing here.

Sneha Mehra (00:28:23)  
Since the model has access to all the generated tokens, it should help it give a better answer. And so how is that done exactly? Well, as we said, we start with a powerful model that we pre-trained, so we still have the training step. Then we create a database of prompt with chain of thought. So we typically do that with mathematical concepts or code because they have ⁓ chain of thought by default, but you can do that with text as well, if ⁓ you have the data to.

Do that. ⁓ Then we retrain it with ⁓ fine-tuning again using the chain of thought examples. Then we do reinforcement learning, just like when training regular models, but here is to reward the correct and structured reasoning. And here we can use synthetic data. ⁓ So just a generation of diverse examples of reasoning chains and ⁓ manually eliminate the wrong ones. So basically we just

Teach it to try to use as much or as little compute as needed to correctly answer the question. And then finally, there's a balancing ⁓ phase that I pretty much just described, but you just ⁓ ensure in the training phase that advanced queries or complex problems use more tokens and simpler ones don't use as much. So you teach the model to use more or fewer tokens. And here, because I thought it was

Quite interesting and can change the way you use models. Here's how OpenAI provides reasoning models to users. If you exchange with the model, you will, for example, here enter your prompt, which would be your input, and then it will generate a reasoning that they don't fully provide you, and an output, an answer. Then, if you ask it a follow-up question, the answer and the input, ⁓ the previous prompt, will be sent.

Along with your new question. So it won't have access to the reasoning that you see. So for example, if I ask how many R's are in Strawberry, and in its reasoning, it spells out the letters and does lots of thinking and then just replies in its answer three hours. Then if you ask it the follow-up questions, assuming it knows about each letter of the word, ⁓ it won't, because this was in the reasoning that you've read, but not in the output.

Sneha Mehra (00:30:48)  
So this is important to take in consideration when you exchange with a model. And so ⁓ I described it, but technically, here's how ⁓ reasoning models work. Basically, we have ⁓ our tokens being generated one at a time token one, token two, token three. And then ⁓ usually we have the end of text token which tells the model I'm done generating ⁓ it's ⁓ and sends the reply to the user.

This works and it's in fact completely autonomous by the model because we've trained it to provide answers. And in our databases, we created these end of text tokens at the end of each of our examples. So the model learned that after each example, it needs to generate this end-of-text token. And so after training, when it finishes its reply, it will generate this end-of-text token.

And we cannot really know when in advance. This is what the model has been trying to do, and itself only knows when your answer is finished. And by the way, this is how we write a token usually with these symbols. But that's before reasoning models. Since reasoning models, we have new tokens that work exactly the same way. Basically, we have pre-token ⁓ during thinking, and we have an

End of thinking token to then tell the model, okay. Now I'm done thinking, I will generate my answer and then end of text to send a user. And so this is why OpenAI can do whatever they want with the thinking tokens, because they are clearly separated from the ⁓ response tokens. ⁓ And basically they've trained the model with, as we said, chain of thought examples, ⁓ mostly mathematics and code, and they just

Manually added this end-of-thinking token at the end, for example, of mathematical proof right before giving the answer, the final answer. But that doesn't really explain why they work. Well, there's a few reasons. First is because it's obviously a structured problem-solving approach. Reasoning models just do like us and split a complex task into simpler ones. So it's just normal that it works better.

Sneha Mehra (00:33:16)  
Than a regular model. Then it's just the natural evolution of scaling loss. We saw the scaling law of the more training data that we had, the better the results. The more ⁓ time we spent training, the better the results, the bigger the model, the better the results. And now just the more compute that we use when serving the model to our users, the better the results.

So it's just another new scaling laws that companies are aiming for. Also, we have robust verification. We use synthetic reasoning chains and automated checks, ⁓ whether it is through code executions or mathematical proofs, as I mentioned, to ensure that the model is trained properly for reasoning. And so that makes ⁓ the creation of somewhat large data sets quite easy because.

Mathematical proofs, for example, have this intrinsic behavior of having a whole reasoning before answering, which we can give to the model and specifically trying to replicate this process. And obviously, just like embeddings, if we give a better context and are clearer in our text, and so to what we give the model, well, the model have better details to know which token would be the right one for the

for the answer. And so we are basically just making it clearer and clearer for the model, with itself generating the reasoning to provide a better final answer.

So these models are very powerful and they do the thinking before answering, which leads to a new possible training concept that we've seen in session two, reinforcement learning fine-tuning, which is basically a way to automatically retrain a model ⁓ using a precise evaluation function, or what we call a grader, with a handful of examples.

Sneha Mehra (00:35:22)  
I quickly talked about this in the second course, but we we didn't really enter into how it worked. And here we believe it's relevant because it's a new possibility that we can do. And we can, in fact, improve models a lot by doing that. But how does it work? Well, it works quite simply. You mainly need a grader. So a way to evaluate the answer. If it's a mathematical proof, it's simple. It will be if the response is good or not.

Else it can be by using an ⁓ LLM as a judge and giving a grade, ⁓ or following a checklist with true or false and then combining it, etc. There are many ways we can do that, but you need a function, a way to quantitatively evaluate ⁓ the type of responses that you want. Then you have multiple steps, as usual, when you want to fine-tune, where the first step is obviously to have your training data, which here would be a handful of examples of.

Questions and responses, then you have the validation data, which is even less ⁓ examples, but that wouldn't be shown on training, just to know that you are not fine-tuning your model towards your exact training examples. You are also generalizing a bit. Then we have the grader that we just mentioned, ⁓ and this grader would give the feedback.

To the models saying if the answer was good or not, and how close are we to a good answer? And so basically, you just need your training data set, which is ⁓ the question and answers, and then the reasoning model would fine-tune itself to reason to better answer the questions that you have. So instead of ⁓ improving the model little by little by measuring each token generation if it's good or not, here

We don't measure the thinking tokens, we just tell it it's a good answer or it's a bad answer. And then it itself needs to ⁓ understand how to change its thinking to fit the proper answer. So, anyways, it's just to say that this is why reinforcement fine-tuning is much different than fine-tuning. It's because instead of the training for optimizing each token generation, we instead train on optimizing the answer only. And so we need

Sneha Mehra (00:37:46)  
A quite powerful model to understand just from the answer how to improve it. And so for now, only thinking models are able to do reinforcement fine-tuning like this. And ⁓ at the time of recording, ⁓ it's coming soon to OpenAI and to many other platforms. ⁓ and when to use it. Well, we've seen it in the second course, but basically, ⁓ it's since we are not training ⁓ on each token generation, but just on the answer, it's mostly to improve or change the user experience.

And ⁓ when you don't have access to lots of data. It's a very good approach to use if you have a powerful model or a fine-tuned model already and want to just improve it for your own application, but it won't really work to teach it the new programming language, etc., where you would need to change all tokens. Here, you just want to teach it to give a better answer based on things it already knows, but doesn't really ⁓ formulate properly.

Obviously, it allows to replicate an expert reasoning, and that's how we should see it. One last thing ⁓ that we find very interesting, which is deep research, which is basically the perfect next step after reasoning models, where we have retrained it to ⁓ do better internet search and better summarization, better iterative process with such searches, have specific tool used like code interpreter.

Or tables, etc., and have internet access, which all leads to deep research that many companies provide, like OpenAI, but also Gemini, Grok, and others. I already mentioned the tools, environment, and sister prompts, so I won't say that again here. But basically, deep research is a combination of web browsing, planning to know which ⁓ website you want to go to and what query you want to make, etc., multi-step reasoning.

And synthesis. ⁓ So it's more than just a chatbot, it's an autonomous research assistant that can be extremely useful for many things. For example, recently I just ⁓ searched ⁓ for a family member to help them find the perfect electric bicycle for them. And so they had a budget in mind and just preferences, for example, battery, longevity versus speed, etc. And so I just entered everything in the search box.

Sneha Mehra (00:40:14)  
And the Deep Research made ⁓ a very in-depth research browsing, I think it was like 30 or more websites, ⁓ and spending 15 minutes to try ⁓ and find the best bicycle for our needs. And it ended up with a good list of bicycles with no hallucinations because it used lots of internet websites, and it was super useful helping them find the right bicycle for them. ⁓

Deep Research is a good agent take example because it has a structured multiprocess system that is actually automatically determined by itself. It starts with a plan, decomposing a complex query from a user into sub questions. ⁓ And ⁓ for example, in the case of OpenAI, ⁓ if you ask Deep Research a query, it will reply to you with questions to ensure that it makes the right plan. So it's just to show that.

It's all automatic by the agent. Then it does search using tools like browsing the web, ⁓ analyzes the web, the results from the web searches, and pivot if need be. For example, in the case of my bicycle, it searched on a few top 10 bicycle websites, then narrowed down towards prices and towards battery longevity, and made some other researches based on the top 10\.

resources it found, etc. So it it had to do things ⁓ based on the results of the previous step. ⁓ And lastly, it synthesizes everything to give you a good structured detailed report. ⁓ Here are the deep searches that we can have access. For example, just quickly you can screenshot this if you want, but ⁓ th there are multiple ones that you can use. And ⁓ basically I use OpenAI for all very complex queries because it usually does

Longer and more in-depth analysis, but gives very good and detailed report. And on the opposite, if you want a good grounded answer, Perplexity is much faster, much more efficient, and gives a very shorter answer, but ⁓ allows to have very reliable sources and summaries. And also we have Gruck from Twitter. If if you like seeing ⁓ X or tweet citations in your responses, it's also free, which is

Sneha Mehra (00:42:40)  
Quite nice. There are also open source alternatives like GPT Researcher, for example, and others. And so here's to show if you didn't try Deep Search for now, it's much more in-depth than even the more complex query that I showed with O3 earlier today, where it used one minute of compute to answer the question. ⁓ instead, here if I asked the same question, here ⁓ again it's in French, but

I'm basically just asking what's the difference between training a reasoning model versus a regular model. Here we see that the first red highlighted thing on the left, that it asks ⁓ if I'm talking about a specific model or not, like GPT or ⁓ Gemini, and if I want the explanations to be more theoretical or more applied. And so I said I want both, and I want, to example, compare O3 versus GPT-40.

And so then it searched for nine minutes and through 18 sources to give me a very detailed report. And you can then ⁓ look at these sources ⁓ at the top right to ⁓ see if they made sense or not. So, anyways, it's a very powerful tool and we think it's actually one of the best ones to explain what an agent is and to demonstrate the agentic capabilities. All right, now let me just quickly finish this first part.

With ⁓ some popular tools that we may want to ⁓ use versus develop when we build such a genetic system, just because ⁓ it's important to ⁓ not spend a lot of time ⁓ recreating the wheel, basically. And so for knowledge, for example, you definitely want to use some kind of web search as for deep research ⁓ instead of developing your own web search system.

It's better to use an API that does that for you quite optim efficiently. On the opposite, for file search, as we've seen in the second course, it's quite simple to integrate. And so it's much better to build your own system ⁓ instead of using some kind of third-party file search system. For capability extensions like Cone Interpreter, you definitely want to implement it, use ⁓ like the OpenAI API or whatever. You don't want to code that yourself.

Sneha Mehra (00:44:59)  
And likewise for the calculator, just because it's too complex for nothing, it already exists. For actions like computer use, it's the same. You don't want to build this computer use agent tick system unless it's the goal of your company, just because it's a very challenging ⁓ task to tackle. And then we have function calling. That is basically the way to implement your own tools. So you definitely want to do that yourself. And basically

The tool is a function, and you describe the function schema to the LLM so that it can know when to use the tool and how to use it. And then it will generate the arguments in a JSON format, which you can then ⁓ use and ⁓ give back the answer of the function to the LLM, etc. So function calling is basically tool calling, it's the same thing, and it's ⁓ a way to implement tools yourself. we

Of course, advertise to mostly make your own tools. This is basically the IP of your company. You want to build tools specific to your applications, to your needs. ⁓ and lastly, a bit ⁓ the opposite, but ⁓ if it already exists, we definitely promote to use it. And then ⁓ if it works in your case and you have a proof of concept working, etc., you can consider developing it yourself.

Unless it's extremely complex where ⁓ you need to have good return on investment. Anyways, ⁓ these were just some small words on the popular tools. We will now dive into the second part with my colleague Omar sharing about ⁓ MCP, a very interesting recent protocol, and also talking about ⁓ agent-to-agent and a super interesting case study to better.

Recreate what someone would do in a typical ⁓ workflow or agent situation. ⁓ Hello everyone. So I will start with the second part of this session. ⁓ And like Louis mentioned at the beginning, I will be talking about two different protocols today. So first we have MCP, ⁓ a protocol made by Entropic, and then A2A, a newer protocol made by Google. I will mention what are the benefits of using them.

Sneha Mehra (00:47:18)  
And then I'll end today's session with a case study. So basically, I will show the different steps ⁓ that we can take ⁓ when we are creating an application that uses AI. And in this example, we will be creating a translation service, ⁓ basically. So first, what is ⁓ MCP? What's the model context protocol? It's an open standard created by Anthropic.

This past November 2024, ⁓ and it basically defines how ⁓ we can provide context to ⁓ these AI models. ⁓ So what they define is three different ways to ⁓ give context. So first we have the ⁓ data, then we have prompts, and we can also give them tools. What how does it work?

Well, basically, MCP works in a client server ⁓ architecture. So we have here an image basically showing all the different components of how we can ⁓ use MCP. So we have the host application. So this is any type of app that basically implements ⁓ the MCP client. So ⁓ I list here some examples that actually

Use right now the MCP client. So we have cloud desktop, we have different IDEs such as VS Code, cursor, we also have extensions for VS Code that ⁓ use the MCP client. And we also have CLI tools like Cloud Code that all ⁓ well where all of them basically implement the MCP client. And this is necessary when we

Will be connecting to ⁓ the MCP server. So ⁓ these MCP clients ⁓ connect or communicate with the servers via the MCP protocol. ⁓ And ⁓ once they are in communication, connected to the MCP servers, now the MCP servers can ⁓ expose specific ⁓ or ⁓ yeah, specific different capabilities. So, like I mentioned,

Sneha Mehra (00:49:41)  
It can be data if we are doing, for example, retrieval of information. It can be simply prompts. So if someone developed wrote ⁓ a very nice good prompt, ⁓ and that person wants to share that prompt to ⁓ someone else ⁓ on a team, ⁓ here MCP can be useful for that. And these MCP servers can also expose tools that are going to be very useful.

For giving different types of tools to ⁓ LLMs. ⁓ So here are some examples. We can connect ⁓ LLMs ⁓ with MCP to, ⁓ for example, ⁓ SQL databases. We can share prompts. So for example, if someone on the team, on the marketing team, wrote a very good prompt, then that ⁓ prompt can then be shared via MCP to other team members.

With MCP, we can also easily connect ⁓ the LLMs to third party ⁓ services. So for example, if we want to do web search, web scraping, we can wrap those tools in ⁓ put them in a NCP server and expose those tools to the MCP client. And then the the LLM on the on the on the client side will have access to those tools.

So we can consider MCPs as basically a specification for AI microservices. ⁓ Here I have another figure that summarizes summarizes well what I'm ⁓ explaining. So this you can find this figure on a very good talk done by the creators of MCP. So I recommend you check out this talk on YouTube. ⁓ But yeah, here we can see the different

Components of MCP. We have the client that invokes the tools, the data, the prompts. Then we have the MCP server that exposes tools, prompts, and data. And then on the bottom we can see what we mean by by tools. So any function that can basically ⁓ be exposed to an LLM is a tool. And for the data, yeah, it can be anything from

Sneha Mehra (00:52:06)  
Accessing a database to using ⁓ retrieval services, web search services, and then the prompts where it's very convenient if we want to share ⁓ a well written prompt to anyone else on the in the company or in the team. I have another figure here. So we have the MCP clients ⁓ that I listed before. We have the servers in the center and the

Basically, the ⁓ the use case, the the tool connected to each server. So databases, it can be anything, basically. Anything you imagine, you can basically connect to a ⁓ MCP server. And so what are the benefits? Why should we use MCP or or why it's convenient to do so? ⁓ Well, ⁓ it it basically ⁓ can simplify the integration of tools.

The two systems that use ⁓ LLMs. So ⁓ if you developed applications before, you know that you can give any tool to LLMs, but you need to develop your own integration each every time you actually develop a new application. So here on the left, you would be connecting your APIs, for example, your third-party APIs to your applications.

So it can be Slack, your local file system, GitHub. And then on the right, you can see that we add ⁓ a newer component. ⁓ So we have the MCP, basically the MCP server and client. But here ⁓ it will allow us to do is basically reuse these integrations ⁓ and easily be able to share them basically. So when we write an integration.

It's basically within this ⁓ red box. And then if we are we want to use those same tools in another application, then it will be very easy to basically move up move the code or connect the same MCP server to ⁓ a different application. And yeah, everything about integrating third-party services becomes a little bit easier in this case. So some people on

Sneha Mehra (00:54:33)  
On LinkedIn, ⁓ basically compare this to the USB protocol. ⁓ So here this is just an illustration example of what MCP is when compared to basically the USB protocol. ⁓ So you can see that here MCP is basically ⁓ like the connector, the connection between the computer and ⁓ the third-party services, or here basically ⁓ a third-party device connecting to a computer.

And because we have USB, because we have MCP, we don't need to reinvent the integration every time we develop a new application. Here we can just reuse MCP to make the connection between our application and different MCP servers or different third-party services. So ⁓ it, like I said, it simplifies the integration of tools. ⁓

To systems that use LLMs, it allows more flexibility in the architecture. So we can change hosts. So if I'm using, for example, cursor and I'm connecting to a specific MCP server, ⁓ and then I choose to use another IDE, for example, VS Code or or even a different application like Cloud Desktop, I can connect to this very same MCP server and it's going to be very useful.

I will not have to write a new integration code. Also, the MCP standard will ensure compatibility between all of these components. And like I said, it's very easy now to share them because we don't have to ⁓ write any integration code. ⁓ And we can think that in the future it would be possible to monetize basically tools ⁓ using MCP. ⁓

As a way to easily connect to those tools. Another reason ⁓ to ⁓ use MCP or maybe future benefits ⁓ are because different LLM providers ⁓ are going to be integrating MCP ⁓ within their applications, their APIs. So I here I have Sam Altman and Demis Hassabis saying that they they will be integrating.

Sneha Mehra (00:57:01)  
MCP within their their libraries or their even their own applications like ChatGPT desktop. So I just want to quickly mention that ⁓ if you already developed applications with LLMs, maybe you you already actually use a protocol before it. So we already had ⁓ OpenAPI. ⁓ So ⁓ whenever we connected to

For example, third-party services, we would use OpenAPI to connect to let the model know what are the different functionalities on this third-party service. So, why the need to introduce a new protocol? ⁓ And here I have actually within ChatGPT, we could also use Open API to connect to different ⁓ actions or third-party tools if we wanted to.

So ⁓ here I just want to mention that unlike OpenAPI here, ⁓ MCP will allow ⁓ the client basically to dynamically discover what are the different different tools ⁓ in the MCP server. So this is ⁓ a little bit different. ⁓ So with MCP, you have two endpoints.

First, the MCP client would call tools list to discover what are the available tools. And this is done every time the MCP client starts. And then there once it knows what are the tools and it wants to use a specific tool for accomplishing a task, ⁓ then it can use the other endpoint, which is tools call, to execute a specific endpoint. So

Here I do I'm just putting ⁓ the ⁓ the JSON ⁓ basically description of a tool. ⁓ So here we have the name. So here we would specify the name of the tool, the description of what each tool does. This is very important so the LLM on the client side can choose in a good way what tool to use ⁓ and make sure that every tool has a unique and very ⁓

Sneha Mehra (00:59:23)  
Descriptive description, and then the basically the types ⁓ of the input parameters to that specific tool.

So basically, before ⁓ MCP, we would hard code what were the different tools in a third-party API service. But here the tools are no longer defined statically. Here the it's basically dynamic ⁓ and it will change every time you restart MCP client. And what's also nice with MCP is that

By default, the client can also chain multiple tool calls within a single completion. So if you're giving the LLM ⁓ the tools to check the weather and to send a message on WhatsApp, for example, ⁓ and then ⁓ you instruct the LLM to check the weather and send a message within the same completion here with MCP. ⁓

These tools will be allowed to be executed in within the same completion. So this is very very very useful when using these AI applications. So yeah, I put two examples. It could do retrieval in a database. So this would be the first tool call. And in the same completion, DLM could choose to use the ⁓ Python analysis tool. So ⁓

A sandbox where the LLM could write Python code ⁓ and have it executed within this sandbox and then then get the results of the analysis within the same completion. So this is very useful for that. Now I will talk about the current limitations of MCP because since it's a relatively new protocol, of course, right now it's not perfect. ⁓

Sneha Mehra (01:01:32)  
It's open source, ⁓ and many companies want to integrate it within their ecosystem. So we can expect the protocol to improve over time. But I just wanted to mention ⁓ at least a little bit what are the current limitations so you know ⁓ what to expect when using MCP. So, what's missing right now? Well, what's missing is basically something very important, which is better security.

Right now, MCP is very vulnerable to what we call prompt injection. Prompt injection is when ⁓ users don't notice ⁓ that hidden messages, hidden prompts are sent to the LLM, and yeah, users don't actually know what's going on. So here I have an example highlighted by ⁓ Simon Wilson.

So if you're interested, you can go and read the the blog post because it contains ⁓ more details about about this specific ⁓ example. ⁓ But ⁓ if if I try to summarize a little bit, he's basically showing that if we were to use, for example, ⁓ the MCP server that can connect to WhatsApp, and this MCP server has two tools.

So, for example, we have the list chats tool that will retrieve all the chats ⁓ we have in our ⁓ profile. ⁓ And we have the send message tool that can allow users to send messages using the MCP server. What could happen is that whenever a user wants to send a message, for example, the ⁓ description

Of the tool can basically tell something different to the LLM. So here in the ⁓ basically the description ⁓ of these tools, we have this prompt where we are telling the LLM to forward a copy of all the messages ⁓ in the list of messages to a different phone number. And this can happen even if you didn't actually instruct this as a user.

Sneha Mehra (01:03:57)  
Here, this is written ⁓ in the ⁓ tool descriptions. So you didn't write that, but yeah, ⁓ this is what actually will be read by the LLM. And yeah, the LLM could choose to execute this instruction ⁓ and basically send all your messages to a phone number that is not yours. ⁓ We so we also have the possibility of

Yeah, basically installing malicious ⁓ MCP servers. So we have to be very careful about that. Here I have another example, ⁓ and you can also go check it out. ⁓ it's another blog post written by invariantlabs.ai, where they also ⁓ exploit ⁓ the ⁓ WhatsApp MCP server. So again, here

This is just an example, but here we basically have a malicious ⁓ prompt that will be read by the LLM. And this is written in the tool description. So this is not written, ⁓ this is not basically shown to the users of the ⁓ of the MCP server. So this will not be shown in the MCP client. ⁓ And ⁓ using this tool description, we can try to ⁓ basically do malicious.

things. ⁓ So someone could write this and try to gain access to things ⁓ private ⁓ stuff. So ⁓ yeah. So for example here we want to basically read private information and send that via HTTPX to basically ⁓ a malicious ⁓ malicious person.

So ⁓ given these tools that I explained before, list chats and send messages.

Sneha Mehra (01:05:54)  
What could happen ⁓ is that here we have a different tool. ⁓ And whenever this ⁓ tool ⁓ is executed, it will change the behavior of the other two tools that were totally good and innocent before. So whenever ⁓ someone decides to execute this get fact of the day tool.

Because of the description, the malicious description of this tool, it will change the behavior of the other tools that have actually ⁓ very innocent tool descriptions. So this could happen without your awareness. ⁓ so if you were to use the other tools before, they would work, they would be total totally ⁓ work as intended. But ⁓ once you basically call this third.

Tool, this description basically changes the behavior of these other tools that you executed before. So this is something that could also happen with MCP right now. ⁓ And yeah, it would be it will be important to improve this ⁓ as companies or developers continue to adopt MCP. ⁓ Now we have something else that's missing, and this is actually something that is getting

Better over time, but we would want basically more a better ⁓ a more granular authentication. ⁓ So let's say you have an MCP server ⁓ and you want different permissions for different types of users. Right now it's not possible to do. And that's something that would be very actually useful to implement in the future. So yeah, basically.

Better authentication. Now, ⁓ what's ⁓ missing is better context management. ⁓ So ⁓ if you reach the context length of an LLM, ⁓ right now it's not it's not doing anything smart about how to manage this context that is getting very, very long and getting to the limit of the LLM. So right now ⁓ on the protocol layer.

Sneha Mehra (01:08:22)  
This is not something that is very well executed. Basically, ⁓ once you reach ⁓ the context limit, basically the earlier context is just cut off ⁓ and ⁓ yeah. So you you basically lose context of what happened at first in the conversation.

Also, what's ⁓ really hard, not just ⁓ for the protocol, but just in general, is how to handle a lot of tools. So we know that if you give a hundred tools, for example, to an LLM, then it's going to be actually quite difficult to ⁓ make sure that every time the LLM wants to execute a task ⁓ or you give it a task.

That the LLM would will actually choose the correct tool to ⁓ achieve the goal. So this is not something that is actually solved in the ⁓ LLM ⁓ within the LLM layer. It would be nice if it was solved within the LLM layer, but yeah, right now, even in the protocol or the application layer, it's very hard to implement. ⁓

Yeah, this is something that could actually be improved either in the LLM itself or in the ⁓ orchestration system around the LLM. ⁓ So yeah. Also, what could be nice to add within this protocol layer is a better ⁓ standard when it comes to logging and monitoring. So right now there's a ton of different services you can use.

To monitor and log your AI applications. Many of them are actually very useful. But here it would be very nice to have it as a standard, like ⁓ just one way to actually monitor and log things within your AI application, because it would then make it very easy to ⁓ basically change ⁓ monitoring services or actually add these month monitoring services very easily.

Sneha Mehra (01:10:39)  
So it would be nice if we had this standard monitoring within the protocol layer. Yeah, it would facilitate debugging and tracking performance when ⁓ it it it basically this is what it allows. ⁓ It allows to better track performance when we deploy AI applications in production. So ⁓ where can we find these MCP servers? Well, right now, ⁓ here I just put three sources, ⁓ but there is right now.

There's probably tons of different sources. You can find examples of good MCP servers. So ⁓ I listed the official GitHub repo of MCP with different examples. We you have you also have third-party or you know people that ⁓ basically collect and provide a good list of curated MCP servers, and you also have websites.

Like smithy.ai that try to ⁓ make it easy to explore and use these MCP servers. How can we build MCP servers? Well, the first way to do it is simply to just use the different SDKs that are made available by Entropic and that are open source. So if you want to contribute.

To making these SDKs better, you can do so. So you have the Python, the TypeScript, the Java, and the C sharp SDK ⁓ that you can use to develop MCP servers or clients. So here I'm just going to ⁓ show a very short example that you can actually find in the MCP documentation, but but that is very very useful to ⁓

Basically, just have a good idea of what it looks like in code. So, this is what I'm going ⁓ to show here quickly. ⁓ So, let's say we want to create ⁓ a get weather MCP server. So, basically, a server that implements ⁓ a getWeather tool. You can do it very fast. Here, I'm just showing the Python SDK that you can use. Here

Sneha Mehra (01:12:59)  
You simply just initialize the fast MCP server with one line. ⁓ then you of course define some constants that you will use in your getWeather tool. Here you ⁓ write the the tool that will actually call the GetWeather API. So here you notice that it's written in async format, just to make you you can have

And a sync MCP server to have multiple people ⁓ use the MCP server at the same time. ⁓ And here is the actual tool that will use the previous function. So you can see the decorator mcp.tool that you need to write at the top of each ⁓ function. And then below ⁓ that, ⁓ actually just below, you have the tool description in the triple.

Yeah, in the in the quarrett quoted string. So here we have the description, get weather forecast for a location, and then a quick description of the two input parameters you need to give to the tools in order to get ⁓ a weather ⁓ result. So yeah, you have here ⁓ another function get alerts, and then ⁓ you have ⁓ again.

What it looks like to have a tool defined in a JSON ⁓ object. Basically, this is what the MCP ⁓ client ⁓ sees and reads whenever it starts. ⁓ Basically, it will read a list of tools that ⁓ every connected MCP server can provide. Then we have, of course, the run command, so MCP run to actually start.

The MCP server. And here we are using the STDIO transport. You can also use the SSE transport if you're ⁓ communicating via ⁓ basically the internet, so a remote MCP server. ⁓ And ⁓ for example, let's say I'm using Cloud Desktop and I want to use this new MCP server. I will need to edit the JSON config here.

Sneha Mehra (01:15:25)  
In this is the the path if you're on macOS. So you open the cloud desktop config.json file, and then within this file, you will basically paste ⁓ every mcp server you want the mcp client to use. So in this case, I only have one mcp server, and what we are providing here is just a way to start this mcp server. So

The command to start this MCP server is UV. This is just ⁓ a tool, a Python tool that can ⁓ launch ⁓ different ⁓ Python applications. ⁓ And here we have the arguments of that command. ⁓ So directory and the path ⁓ to that tool, to that ⁓ Python file, Python project, run ⁓ and the ⁓

Script that we are launching. So this is ⁓ how you add these MCP servers to these MCP clients. And so if you launch the MCP client, it will start in the background the MCP servers. ⁓ And ⁓ here you can notice at the bottom that you have two different connected tools. ⁓ And if you ask for the weather.

The LLM will know that it has two different tools for the weather, and so it will try to ⁓ actually execute them in order to answer you. So ⁓ you can see how it executes the tool and how it ⁓ answers your question. So this is very useful if you're ⁓ developing AI applications. Now, what if you want to basically

Use MCP servers but not directly in an existing application. Well, you can connect these MCP servers for your agents for your current AI applications. There are different frameworks right now that you can use to develop AI applications. So you have, for example, OpenAI agents SDK. ⁓ And within this framework, you can see that.

Sneha Mehra (01:17:50)  
When you are defining an agent right now, you can actually just ⁓ define a list of mcp servers the agent ⁓ can use to achieve a task. So this is very ⁓ very useful. ⁓ If you are working with Langchain, there's what ⁓ we call MCP adapters. So this is a library that will allow you to convert ⁓ MC MCP clients into ⁓

Langchain tools. So if you're if your app ⁓ is written with Langchain, then these adapters will make it so that you can use all of these different MCP servers. You can have the the equivalent in Lama Index. So if you ⁓ use Lama Index for your app, you you can use the ⁓ this ⁓ function to convert that to a something Lama Index can recognize basically.

And then you have all of those different ⁓ all different frameworks you can use. You have QAI, Langgraph, all of them also make it provide a way to use MCP servers for your applications. Now we I will switch ⁓ topics. Now I will talk about this newer protocol that is very future focused. Here, Google is thinking about how to connect.

Agents together. So agents that actually don't live in the same environment, agents that ⁓ don't share ⁓ memory, don't share the same tools. So this is basically a protocol where Google imagines a future where different ⁓ agents, for example, you would have an agent working for you, ⁓ and this agent ⁓ can could communicate.

To a company agent. So for example, ⁓ you would use the Gemini agent ⁓ and it could ⁓ connect to a third-party agent. So let's let's say a different company, ⁓ let's say ⁓ Amazon ⁓ or ⁓ yeah, any any any type of ⁓ a different company would have ⁓ a different agent. So let's say Amazon here with this protocol.

Sneha Mehra (01:20:16)  
This protocol defines how the communication can be done between our agent that we are using, we are using as client, to these other types of agents that live ⁓ on a different company server, for example.

Or just developed using a different framework and yeah, basically how you would communicate to another human here ⁓ we are thinking about how different agents could communicate ⁓ between each other.

So ⁓ yeah, it aims to standardize exchanges between agents. ⁓ So here you are the end user, you are using a client. So you would have an agent ⁓ that lives in your client, and this client can then talk to remote agents to delegate ⁓ different kinds of tasks ⁓ and get to a an answer that. ⁓

That you will be satisfied with. So ⁓ here ⁓ it's for agents that actually don't know each other. So ⁓ another way ⁓ to define them is to say that they are black box agents or opaque agents. Here, my agent will communicate ⁓ to other agents to use delegate task to ⁓ another company or service agent.

These agents, like I said, don't share the memory, don't share tools, they don't share thoughts, ⁓ they are completely independent, but they can communicate with each other to accomplish more complex tasks. So here we have ⁓ a figure that tries to communicate how this communication works. So here on the left, we have one agent ⁓ that we ⁓

Sneha Mehra (01:22:18)  
Where we have ⁓ local agents ⁓ that live in the same environment. Here, MCP is used ⁓ as ⁓ basically a way to connect tools to these agents. Then we have this line in the middle that ⁓ defines or shows that this ⁓ is a division. So the agent on the right here would be in a different organization, different.

⁓ server here it's a black box agent remote agent ⁓ and the 82a protocol is to basically communicate with those kinds of agents and those agents would have their own local agents their own LLMs written ⁓ with their own agent frameworks yeah completely different thing here like you saw ⁓ 82a ⁓

Is complementary to MCP. So Google here is trying to not do the same thing as MCP. Here, MCP will serve to connect the LLMs ⁓ or the agents here to different tools. So, like you like you saw here, MCP is mainly used to connect the agents to different APIs or tools. ⁓ And also ⁓

⁓ Besides connecting them to tools, here what Google recommends is using MCP for agent discovery. So let's say you have a remote MCP server, ⁓ and this ⁓ MCP server can have an endpoint called slash resources, and this ⁓ endpoint would then provide information for different. ⁓

Black box agents ⁓ available ⁓ to connect to for ⁓ to do more complex ⁓ more complex applications. So here, ⁓ once ⁓ we call the resources endpoint, ⁓ we will get the information about those black box agents. The information here is written inside the ⁓ agent card. And once ⁓ we discover these agents,

Sneha Mehra (01:24:42)  
With the MCP server. Now ⁓ we can then connect directly to those agents using the A2A protocol. So this is an ⁓ basically a second way to use MCP within this ⁓ future that ⁓ Google is ⁓ is ⁓ imagining.

So ⁓ right now there's a since it's a very new protocol, there's not many implementations yet. But if you look at the ⁓ the GitHub ⁓ repo, you can actually find some implementations of A2A. So you have implementations using Langgraph, QAI, and also one implementation using the Google ADK library. That is also very new.

And that you can check out. So if you're interested about using A2A, you can go and check the ⁓ the GitHub repo. So now let's finish this session with a case study. So this case study, we mainly wanted to include this ⁓ in today's session just to show to ⁓ people that have never developed AI applications what are the main steps that we can ⁓ follow to develop.

these apps. Basically what's what are the thought processes, what are the questions we need to ask ourselves when we are developing these ⁓ these apps.

So ⁓ in this case study, ⁓ we are showing an example of creating a domain-specific translation service. So here we want to create an app that can translate ⁓ basically technical technical terms. So what's the objective here? So yeah, you you need to think this ⁓ just like what are the the thought processes.

Sneha Mehra (01:26:48)  
And yeah, ⁓ just to follow along. So what's the objective here? Well, we want to develop a service that translates ⁓ text or documents, ⁓ and these documents can be either.txt, ⁓ Word documents, PDF documents, and we want to translate them between different or multiple languages.

Here, the challenge ⁓ is that we must correctly translate these ⁓ all of the terms inside of those documents. But we have ⁓ basically domain-specific terms that might change over time, ⁓ and we want the application to actually ⁓ be able to stay up to date. So if we change a definition over time, we want the application to.

⁓ follow along so that the results so that users actually have ⁓ good results all of the time we also want to handle ⁓ different formats so like like i said we want ⁓ pdf word documents ⁓ and we also want to ⁓ keep as much as possible the default layout within those documents so ⁓

What we recommend all the time is that you start with the simplest version of your application at first, and then progressively add complexity ⁓ whenever you're satisfied or your current implementation is good enough given your evaluation process.

So ⁓ what's ⁓ the initial requirement? The simplest thing we can implement at first? Well, given that we want to translate to multiple languages, what we can do at first, we can just start with a specific pair of languages. So let's say we want at first just to make sure that English ⁓ is correctly ⁓ translated into French. ⁓

Sneha Mehra (01:29:00)  
We can ⁓ start with that. Now, given this requirement, we need to ask ourselves some questions. So, ⁓ is this something LLMs can actually do? Yes, ⁓ translation is actually something that most LLMs can do ⁓ once they are you know smart enough. ⁓ given give given the size of an LLM, it can actually improve its performance. So

And now we know that most LLMs can do translation, but can ⁓ current powerful LM LMs accurately translate special instructions for ⁓ base basically technical jargon jargon terms. Okay, can can can the current powerful LLMs do that?

Can should we use a ⁓ private API, ⁓ like a proprietary proprietary model, or should we use a local open source model? We can actually start with a ⁓ API because it's easier to implement. So we should always try to do the simplest thing at first. So yeah, it makes sense to start with an API at first. And then should we do complex ⁓ things like include?

⁓ A rag or a few shot prompting? Well, yeah, probably, but let's try zero shot at first. So, what zero shot means is simply to test without examples. And what's the budget? Well, right now it's low because we are only testing via API. Do we need an agent or a workflow? Well, for now it's simply an LLM call, so we we will not be implementing.

Neither a workflow or an agent. So let's do an initial test. We will send a text prompt that contains technical terms to an LLM API, ask the LLM to translate this into French, and check the output for accuracy for ⁓ all of the technical terms. So now when we test the current implementation.

Sneha Mehra (01:31:19)  
We can see by either evaluating the system or just observing what it does, that with this implementation basically it doesn't really work. Some of the technical terms are actually not ⁓ translated correctly. It might be very generic translations. We can think about using ⁓ some examples in the prompt, so including a few shot examples ⁓ within the prompt, but this wouldn't be ⁓

really the solution because we want the ⁓ the trans the translations to be updated over time so this is not something that could work so what's what's the ⁓ the decision that we can make here well we we need ⁓ a way to correctly provide the correct translations for domain specific terms to ⁓ to the ⁓ to the context of the prompt of the of the llm

We can do it with fine tuning, but this would require fine tuning every time we want to update the the translations. So this can be a little bit more costly in time and money. So here, what would be actually like the best ⁓ option is to just include a retrieval augmented generation system where before we do the translation, the system correctly retrieves the correct ⁓

pair of trend translation for for the ⁓ correct input terms ⁓ and we include the retrieve results in the context of the LLM so the LLM can actually know what is the correct translation for those specific technical terms. So now our simple LLM call gets transformed into a what we can call a workflow. So ⁓ now that we have a retrieval component

we need to actually create ⁓ this ⁓ vector ⁓ or whatever we want to use database. So, first now we need to create a glossary. In this case, since we still want to just translate English to French, we can just include all of the terms ⁓ in our ⁓ domain ⁓ of English and have the translation, the correct translations to French.

Sneha Mehra (01:33:43)  
⁓ And now before we do ⁓ the API call to an LLM, we do the retrieval ⁓ of the correct terms, ⁓ and then we add those ⁓ translations directly ⁓ into the LLM prompt before it does a translation. ⁓ So we are basically giving the LLM the answer before it actually ⁓ generates an answer. ⁓ So now this is a very simple workflow.

Because we have ⁓ a fixed number of steps here, we only have the retrieval part. We then add this, ⁓ the results of the retrieval in the prompt, and then we do the call to the LLM API to retrieve the final answer. Now, if we are satisfied with the results, ⁓ either by evaluation or by testing a lot of back and forth with the system.

We can now keep adding complexities. So now ⁓ we don't only want to translate from one language to another, ⁓ from English to French. Now we want to do it with multiple various languages pairs. So it could be English to ⁓ Spanish, French to ⁓ Italian, whatever. And we want to include all of those types of translations into our glossary.

That is in the database. So now what are the challenges? Is that how does the system know what languages ⁓ need to be retrieved to provide good translations? ⁓ And how the how does the basically how the does the retrieval part of the system correctly retrieve the correct language pair? So

Now we need to basic basically think of solutions to do it. So now we can we can do it by ⁓ improving the glossary structure basically. And we need to also modify the current workflow so that maybe a classifier model or even a very small LLM detects what is the current ⁓ input language and the target language that the user wants to ⁓ translate to.

Sneha Mehra (01:36:12)  
So now we can have a better enhanced glossary. So we include all of the possible translations ⁓ into the database. ⁓ And we now enhance the workflow. So like I said, now we need to have a step where we need to identify what is the input, ⁓ what is the input language and the target language. ⁓ So

This can either be by using a very simple model or using a very small LLM if we want to keep things fast. ⁓ And now, since we are able to ⁓ identify ⁓ what are the languages that we want to ⁓ use, then we can do a ⁓ more targeted rag implementation. And then of course we need to ⁓ update the prompt.

so ⁓ that the the model can actually know what is the ⁓ current language in of the input and the actual target language because it's no longer English to French.

So this is still a workflow. We have a little bit more components, but ⁓ everything is still fixed. It it's it's still a fixed number of steps. So this is not something we can call an agent.

Now let's say we have evaluated the system. Once again, here we are not describing ⁓ how we can implement an evaluation system for this specific application, but you can think of ⁓ something very simple where you create ⁓ a long list ⁓ of 50 to 100 input examples with the correct translation, ⁓ and you test the system.

Sneha Mehra (01:38:11)  
To see if the system is able to give the correct answer for all of those ⁓ examples in the basically the evaluation data set. ⁓ So when we are satisfied with the current implementation, we can then think of adding more features, more complexity. So here, what we want to do is to be able to upload Word documents, so files that end in docx. ⁓ And we also want to ⁓

Not only translate but also give back to the user a Word document. So we want the app ⁓ to be able to process ⁓ the Word documents and also be able to create new ones with the correct ⁓ translated text. So what are the challenges by adding this new functionality? Well, we know that for all LLMs, they all basically just process

Basic text, so plain text, they don't take ⁓ Word documents as input. ⁓ And ⁓ we also know that we need to provide ⁓ Word documents as output. So, how how can we try to keep the ⁓ original layout in the final output? So those those would be the the main challenges when implementing this new requirement. What are are the tools that we need now?

And how does the workflow now integrate those new tools? So part of the solution now is to ⁓ create ⁓ tools for our system. So now for the first tool, we can ⁓ actually just implement a very basic Python script that will this Python script will be able to read the ⁓ Word documents using a library like python.x to ⁓

Read the documents of the file. ⁓ We can also create a second tool. This one can also be a simple Python script that tries to write ⁓ with the same layout the translated text back to the user. Now ⁓ the updated workflow will look like this. ⁓ Now instead of just receiving text, we now receive a ⁓ Word document.

Sneha Mehra (01:40:38)  
Then the workflow directly calls the first tool that will extract the text from the Word document. Then we will do the retrieval, the translation, ⁓ and then give the translated text back to the second tool that will then ⁓ create this new ⁓ Word document.

Now we have a ⁓ a a little a little bit more complexity. Now we're our workflow ⁓ uses ⁓ what we call external tools, but it's still defined in a sequential mana ⁓ manner. So yeah, this is still a workflow.

Now, if we are satisfied with that, now again I repeat, ⁓ we you need to be able to ⁓ say exactly if the system is able to ⁓ correctly do the task giving these new requirements. So if you are satisfied, then you can think of adding more ⁓ features to the system. So here now ⁓ we didn't just want to process Word documents, we also want to process PDF files. So now

What we're going to do is basically try to add these new capabilities. ⁓ So, what are the challenges now? Well, the challenge is that we need to be able to extract text from those PDFs. We need to be able to preserve the layout when we're writing back the translated text into a PDF. And yeah, that's ⁓ that's not something that is very easy to do with PDFs.

So, what are the tools that we need to use to do it? ⁓ And how does the workflow will change to accommodate this new feature? So, yeah, we need to add ⁓ a PDF reading tool. We can keep using just Python because it's very easy to ⁓ implement. So here we will use Python with a ⁓ library ⁓ that is able to read PDFs.

Sneha Mehra (01:42:49)  
So we have different choices. We can either go with PyPDF2 or PyMUPDF to extract the text. And then for ⁓ writing PDFs, we can use other types of libraries already available in Python libraries basically. So now ⁓ if ⁓ we update the workflow implementation, we now receive the file.

And then we have to add a routing step. Here we want to detect if we are dealing with a text file ⁓ or a Word document or a PDF document. ⁓ And given the results ⁓ of this routing step, we can either go with the branch that uses the Word tools, the PDF tools, or the text file tools. So now we have a little bit.

Of a conditional ⁓ component. So ⁓ we have different branches, and given the type of file we have in the as input, we go through different paths. But this is still basically a predefined workflow with just a little bit more complexity. Now, ⁓ if we are ⁓ satisfied with, for example, text within the PDF files.

And now how do you handle images within those PDS files? Because the tools that we use could only extract text from those files. So ⁓ what are the challenges now? ⁓ Well, we need a way to convert images that contain text to ⁓ text.

So we can either go with OCR. ⁓ So these are tools that are able to extract text from images, or we can do ⁓ use LLMs, multimodal LLMs that ⁓ can do that. Now, ⁓ if we ⁓ use multimodal LLMs, we need to keep in mind that it will increase costs, it will increase complexity ⁓ and ⁓ yeah, latency ⁓ and ⁓

Sneha Mehra (01:45:09)  
By reading these images, ⁓ it will then become very hard to preserve the layout of the original document. So, ⁓ questions to address: what are the tools that we can use? So, do we use OCR? Do we use ⁓ LLMs that are vision ⁓ that can actually see images? How do we integrate these tools to the current workflow? And what are the current budget implications?

So let's say that we have some choices for these tools. So for the first tools ⁓ tool, we can use ⁓ an OCR tool. So you can use something like Google Cloud Vision that is very cheap and very easy to use. Or you can go to the second option, option B, ⁓ that uses a multimodal LLM to extract information from images. ⁓

Let's say we update the ⁓ workflow with by using option B because we think that ⁓ we want ⁓ to ensure ⁓ high performance in our application, and just using OCR doesn't provide that ⁓ high performance, ⁓ high fidelity for our translation, then the workflow will look like ⁓ the following.

Of course, now we need to receive the PDF, we need to read the PDF, and we need a way to know if there are images within the PDF. So one way to do it is ⁓ if the PDF reading tool doesn't provide information, ⁓ then we can assume that it contains a lot of images. So if that's the case, then we can use option B, the tool B.

that uses a multimodal LLM to extract the text from all the information from those images. And then we can pass ⁓ the extracted information back to the retrieval, the retrieval component ⁓ and then do the translation.

Sneha Mehra (01:47:19)  
So, of course, in terms of the budget, it's going to cost a little bit more money and time because we are now using something way more resource intensive than just an OCR tool. ⁓ And in terms of the type, of course, this is still a workflow. We are using more advanced tools, but this is still a predefined ⁓ workflow with predefined number of steps in order to achieve a task.

Right now, the current system is a complex workflow. It has different branches, it uses different tools, the logic is already predefined, ⁓ and ⁓ DLMs here don't make very complex decisions. They only they're only used to either identify the languages or do the translation, but they they are not ⁓ deciding about how the app should behave. So when does the app

Become an agent. Well, the ⁓ app would need to use LLMs to do basically dynamic tool selection. So ⁓ here, ⁓ one way to do it would be to add an LLM that decides to use either the OCR tool or the or the multimodal extraction with an LLM. ⁓ It could try to do more complex steps ⁓ such as web search. It could try to use

More complex analysis of the input. It needs to have at least ⁓ the possibility to do self-correction. So if the system, the L a final LLM is able to ⁓ review the output, and if the D LLM decides that it's not good enough, then the the process can be restarted again with ⁓ some correction, some pointers to to then get to the correct answer.

And of course, ⁓ if we are letting the LLM choose ⁓ everything about the application, then here it would be very useful to have a planning step because LLMs ⁓ perform better when we instruct them to do some planning before they actually start doing actions with the different tools. So, conclusion, ⁓ we design

Sneha Mehra (01:49:45)  
By adding ⁓ more complexity every time, ⁓ translation service. We tried to keep it very simple at first and then added ⁓ more complexity, more ⁓ decisions to handle different features over time. ⁓ And ⁓ it was mainly to show you how ⁓ these types of applications can be basically developed. And yeah, we we hope that this was useful to you and.

That you can use this information to develop your own applications. ⁓ And yeah, see you for the next session that ⁓ we have prepared for you.

—---------------------------  
3  
Sneha Mehra (00:00:09)  
Welcome to this third session of the course, where in this one we will cover evaluations, which we think is the most essential thing to know in the field, mostly for a few reasons that we will cover in the first few slides. The first one being that without evaluations, we don't know what to do and where to go next. ⁓ This is a direct continuation of the first two sessions where we explored the limitations of LLMs and then how to build on top of them.

So when we are at the step of building on top of them, we need to understand what's happening and understand where where to go next. So that whole session is about this exact goal of helping ⁓ you understand what to do next and how to do so. So that's through evaluations, which we'll cover at different levels and different ⁓ complexities, let's say. So today's plan is to cover the necessary theory to know that we believe is most important.

Just like the usual courses, the two previous courses, we will start with a first hour where I will cover the why, when, and how to evaluate with different levels of evaluations and at different steps, ⁓ whether it is through ⁓ evaluating your prompts, your retrieved information, or the generation part. We will also talk about ⁓ one of the most interesting steps of evaluations, the LLMS judge or the human feedback part.

of it. And then Omar will continue with more practical and code examples with two specific notebooks that we built for this course, the evals for retrieval and the evals for LLM responses, which is a notebook with an LLM as a judge example. You can get both of these notebooks in the description of the course and in the slides that we provide. ⁓ If you want to run them throughout, they are colour notebooks, so you can just ⁓ execute them

Right away if you enter your ⁓ OpenAI API keys. And so let's start with why evaluate. The first reason is because making a prototype is pretty easy, pretty fast. You can build ⁓ a quick chatbot in 10 minutes, but optimizing it will take days or even months to get ⁓ to the level of performance that you want. You may even never get to that point if the task is too complex. So we want to evaluate if it's feasible and

Sneha Mehra (00:02:38)  
How much we are improving ⁓ towards our goal. Second point is because ⁓ AI isn't magic. It's just like software development and it and it requires the same rigor, if not way more, because you are playing with, you are working with much more powerful systems ⁓ that are not completely controlled by you. Just like, for example, using OpenAI's GPT models, you don't.

Really fully understand or know what's behind these APIs that you're using. Also, evaluations allow for many things, including debugging, doing relevant and continuous improvements, aligning with real user needs, having confidence in our model or our system as a whole, the fight-finding the edge cases and silent failures that we may not think about a priori, and it helps with what we call the March of 9th.

So this is basically going from ⁓ the easy ⁓ quote unquote 90% accuracy that we can get with the 10 minute prototypes and then to 99%, 99.9%, and etc. until we basically never reach 100%, but 99.9999. So it helps towards that route because it's impossible to know the real progress you are making ⁓ without evaluating properly your system.

Here is an example also ⁓ for ⁓ why evaluate. ⁓ LinkedIn first developed a chatbot for ⁓ a skill fit assessment, which they implemented, I think, in 2023, pretty early on, and they discovered that the users didn't really appreciate the chatbot. And why? ⁓ Well, it's because they built the chatbot to answer ⁓ the users to know if they were a good fit for the role or not.

And so they didn't have really good evaluations in place, and they just built this chatbot to be able to find ⁓ if ⁓ from the user's resume and from the job description it was a good fit or not. But then the chatbot answered all users ⁓ if they were a good fit or not, and just saying, for example, you are a terrible fit. And the users were quite disappointed, they didn't like that. And so that's just to say that.

Sneha Mehra (00:05:02)  
Even if you build something that works really well, it says if they are a terrible fit or not. It might not be what the user wants. And you cannot really know that first without asking the users, but also without ⁓ building the evaluation system that will evaluate if ⁓ your system replies the way the user wants. And so this wouldn't have happened if they would have surveyed the users first and then built evaluations that would ⁓ assess if.

The chatbot answered the way they want, such as instead of saying you are a terrible fit when ⁓ it's the case, ⁓ saying that ⁓ you are probably not the best fit, you would need to work on this, on that, etc. Basically, the people wanted to know why they weren't a bit a bet ⁓ a good fit and how to become a good fit. So that's ⁓ basically the the use of evaluation is basically to

Understand and know ⁓ how to bridge the gap between what the user wants and what you offer. And you can just know that by ⁓ evaluating your system and doing that ⁓ many times until you see a convergence. ⁓ Another example here of why evaluating is because ⁓ for this chatbot, LinkedIn mentioned to achieve 80% of the experience that they wanted in the first month.

But then needed an additional four months to surpass 95%. So this is just to demonstrate the March of 9 that can be pretty long, pretty exhaustive. And ⁓ to go faster towards that route, you want to evaluate your system at every update to know directly if this update helps or not, instead of walking around and testing many things without necessary improvements.

It's also ⁓ pretty essential just because if you are evaluating at each step and ⁓ documenting this evaluation and better understanding what led to improvements or ⁓ downgrades in your system, if someone else comes into your team or if the team changes, the new person will know what has been tested, what worked, and what to do next. So that's pretty essential in a team that changes, or even just for yourself, to not

Sneha Mehra (00:07:28)  
Try the same things over and over again, or to do many changes and don't know what affected what. ⁓ When to do evaluations? Well, pretty much at each step, ⁓ at each iteration of the pipeline. ⁓ Whether you update the model version, you change the model, you change the prompt just a bit, you change the database, the types of chunks that you have, the re-ranker, the embedding model, whatever you do or change, ⁓ you want to evaluate to know.

How did it affect the the whole system?

Secondly, you want to compare objectively or do A-B test between ⁓ the system, the two parts of the system, or two models, or whatever. If you do these tests, you want to evaluate as well, obviously. So if you compare ⁓ GPD 4.1 to Cloud 4, you want to know which one does best and how. Well, that's only through evaluations. Also, obviously, you want to do that before each big push or each deployment ⁓ or

Do that in production as well from time to time in case there is ⁓ what we call model drift or any kind of drift between what your system does versus what the ⁓ user wants. And so you do evaluations lots of time and pretty much all the time. So there's a good need for automating this whole evaluation system as much as possible. And we will give a few tips for that during these ⁓ next two hours.

So ⁓ we want to quickly iterate. ⁓ we want to evaluate often, ⁓ just like as in regular development, we want to continuously evaluate quality, ⁓ quickly debug small changes. You don't want to make tons of prompt changes and then don't know what caused what. And unfortunately, many companies focus on fine-tuning and building very huge prompt right from the start, where they

Sneha Mehra (00:09:28)  
Don't even have a baseline or know where to ⁓ go to make improvements from their application. And as I said, without evaluations, ⁓ it's pretty easy to go in circles. ⁓ There's also benchmarks that we mentioned in the previous courses. And ⁓ while they are important, they are not the best ⁓ way to measure performance in your system. So when ⁓ do you want to use benchmarks or why?

use benchmarks. Well, it it's because it's it allows for objective comparison with between different approaches. So just to select between two models or two ⁓ chunking techniques or whatever benchmarks can be useful. They provide a common reference for evaluating the pipeline when even common amongst ⁓ everyone in the world. They help validate that the system works in

General cases, but not maybe your own cases. And ⁓ there are ⁓ numerous ⁓ number of options that you can use ⁓ through Lama Index or whatever library you may want to use. But benchmarks are not enough. They don't know ⁓ your domain, your product, your users. Public benchmarks are pretty much ⁓ generic. They won't be measuring what you want to measure.

They capture neither the specific use cases or the edge cases in your application. They are a good start. For example, as I mentioned, to know which LLM might be best for now, ⁓ to just to start with, but they are not the best way for optimizing the system. They also do not measure the right metrics that we want to track. The solution is to move towards personalized evaluations that you build yourself, ⁓ first that understand.

Our or your users and their needs. And to do that, you first need to understand the users and their needs. So you need to look at the queries in your systems. As soon as you deploy it, ⁓ you want to keep track of that and look at them. And beforehand, you may just want to test it yourself and take the time to read the answer and what's wrong or what's not perfect in them. You also want to define your important metrics because there's no

Sneha Mehra (00:11:54)  
Two applications that use the same metrics, you need to find the ones that you want to track and improve on. You want to create a data set, which we will see shortly. You will want to integrate domain experts when creating a data set and ⁓ at pretty much every step of the way here, ⁓ especially in your evaluation loop, where ⁓ we will cover it, but basically you want the person that knows the best about the domain of the chatbot to

Check, analyze, and give feedback to the model. You finally also want to integrate user feedback, obviously. ⁓ So let's cover the multiple steps, ⁓ the evaluation process between a proof of concept and after a proof of concept. So typically for a proof of concept, you often start with just vibe checking. You try to have the clearest prompt possible, the answers make sense, and so you refine the prompt.

Look at more results. That's perfect. It's super useful and it's essential to do. You need to look at the results, refine the prompt, check it, read it, look at the data you have and everything. But in this case, everything is done approximately without real comparisons or real metrics. It's based on your subjective impressions. It affects several factors at the same time. Either you change many things in the prompt or the data ⁓ that is retrieved.

As well as the LLM and the prompt, ⁓ you don't measure what does what. So that also leads to a lack of reproducibility and standardization. Just because, as I mentioned, if someone new joins the team, they won't know what you have tested or what worked or not, and they will just redo the same mistakes. And that's ⁓ for proof of concept. And during and after this proof of concept, we want to transition towards objectivity.

So we need to have examples of what we want to generate and see if our results match ⁓ what we want. So we will compare the current outputs of our systems to ideal answers, which are our examples in our database. And we want to keep the good and the bad examples and note pretty much everything that we can find about them. Is the tone good? Is the answer complete? Is there anything you you want to have in a typical answer?

Sneha Mehra (00:14:22)  
And then you want to ask your expert or your users ⁓ what they need in a perfect answer and break it down into ⁓ potentially a checklist to then use that checklist to give a manual score according to the checklist by experts or yourself. So, for example, instead of just asking the expert if the answer is good or not and explaining why, which is definitely useful and crucial, you want the why, this won't really help you build.

Ideal prompt. Instead, you may want to have some kind of checklist that makes the answer actually good. And you can use an LLM for that if you, for example, ask your expert to describe why the model's answer is good and why it's bad for many examples. You can then ask an LLM to build a checklist out of these ⁓ feedback to find just true and false or yes or no checklist points.

That would make ⁓ a good response. If, for example, the answer needs to mention the username and then needs to conclude with thank you, whatever, those are all things that you can put in a checklist to see if ⁓ it's there or not. And based on how many you have out of the ⁓ 10 points in the checklist, you would have a ranking out of 10\. So that would be quite easy ⁓ to have more of a quantitative node.

Even if it's a somewhat subjective answer. And then we can try to automate this with some kind of system that I just mentioned, but a bit more advanced that we'll see in a few slides. And then after the proof of concept, we really want this type of quantitative data even more. ⁓ We want to use standardized metrics in addition to the subjective ones. We want to use benchmarks if they are applicable, but that's rarely the case.

We will obviously be creating our personalized evaluations where the goal is to prevent problems, resolve limitations without ⁓ creating new ones, obviously, and also understand our system's performance. And so these personalized evaluations that we will build will allow for many things that we already mentioned, but quickly again for doing A-B testing, understand what to work next, what to work on next, and allowing for optimizing the system's.

Sneha Mehra (00:16:49)  
One thing at a time, knowing what ⁓ caused what. And so ⁓ they allow to create a continuous improvement loop and align the matrix that you track with your users. But how do we do those evaluations? ⁓ There are different types of evaluations, ⁓ and here they are ⁓ in order of complexity, ⁓ difficulty of implementation, and cost. So the first level is the unit test.

They are implemented at all stages. ⁓ First, the data preparation. For example, you can have just ⁓ quick if cases to ensure that all documents have titles, text, and a source before the ingestion. ⁓ At the prompt or code level, for example, having regex conditions ensuring ⁓ no IDs are ⁓ disclosed. ⁓ Or you can even have them at model changes. For example, you could have a simple Boolean rule in place.

To ensure that the SQL data is actually returned correctly. ⁓ For example, just ⁓ comparing if the number of products for a client returned is the actual one in your database. Next, you want to create those test cases and with test data for each of these unit tests. ⁓ Those can be generated directly with an LLM. And you want to make these tests as difficult as possible.

While still following what the user would want, obviously. ⁓ Also, the pass rate depends on the performance you want. ⁓ It's not necessarily 100% that you may ⁓ want to ensure to go to the next step or to ⁓ push the deployment. For example, ⁓ you know about ⁓ DALI and Midjourney. ⁓ they are ⁓ image generation systems and they are super popular, people love them.

Yet you may have to send ⁓ the same prompt five to ten times to get one good image. And so that's an accuracy of ten to twenty percent. And still it's a huge success. So that's just to share that it's not necessarily a hundred percent or perfect accuracy that you may want to have.

Sneha Mehra (00:19:05)  
Also, here you obviously want quantitative results, a percentage per error type. That would be ⁓ super useful. And you want to execute and monitor these tests. Here, for example, we have ⁓ two steps ⁓ of the model changes where we show the different errors that we have grouped by type of error. And for example, we see on the left that there's a big ⁓ yellow part which is related to empty responses, and so we could.

⁓ understand that this is a problem in our system, work on that problem, ensure ⁓ there is no longer empty responses. And then in the second step, you see that this type of responses and percentage disappeared. And so that's just to show that monitoring, checking, and evaluating can help fix issues that we didn't even know we had, just like empty responses here that were ⁓ grouped together. ⁓ And so, for example, you may want to.

Use GitHub Actions and continuous integration to collect the results ⁓ and current state of the prompts and the whole system and test to see your progress over time. And you also may want to set up visualization tracking ⁓ right from the start, ⁓ as we will see later on in ⁓ the next lessons, because it's quite easy to get that ⁓ started. And obviously, we want to run these unit tests every time the code changes. After unit test,

We can move on to more advanced ⁓ evaluations with manual evaluations ⁓ that ⁓ cannot be tested using conditions like regex or simple Boolean functions. And they are still important to log traces and to look at our data. There are multiple tools that do that, like LangSmith and Langwatch. You can use whatever you may want. We also want to ⁓ classify the type of error as we did with the unit tests.

Which will be super useful later on. And then with ⁓ this kind of manual evaluation, even if it's very painful to do, we can then use the errors corrected by humans to do fine-tuning or add, modify, or improve our prompts. So that they are extremely useful in many situations, even though they are long to do, because you do manual evaluation. So you read the prompts and the responses and you give feedback yourself or the expert in that case. But that's crucial to do.

Sneha Mehra (00:21:34)  
And here again, we want scores. ⁓ Ideally, we want binary if it's a fail or pass, because it's ⁓ much more consistent and less subjective than just a score out of a hundred or ten. ⁓ For example, to illustrate that, we used to have a project related to color palettes where we basically wanted to build a system that can say if a color palette of five color ⁓ is beautiful or not. We initially asked designers that we had in our hands.

To rate palettes from one to five ⁓ on how beautiful they were. But then we discovered that just by analyzing the rating data, that there were lots of variability between two raters for the same palettes and even between the same rater for the same palette. So that's quite problematic. And it's because ⁓ it's very unclear to know what's a three or a four and what's a two or a three, just like when you go to the

physiotherapist or or whatever, and they ask you ⁓ how much does it hurt from one to ten? what makes it a six or a seven? It's it's pretty hard. And so it's the same here. ⁓ And this is why you want to rate in binary as much as possible. And that can be done, as we mentioned, through some kind of checklist or just a true or fail, a pass or fail rating that is already better than just a subjective score.

Here we will have our domain experts, which is either ⁓ the doctor or whoever has the expertise in our company or in our field, or ourselves if that's the case, to ⁓ score these results. But ⁓ this is very tedious, and you actually don't really have to do all that 100% manually. You can instead do that using LLMs. ⁓ Yes, we can automate them with LLMs.

So that's pretty cool because you can use LLMs to improve your work. And you can also use LLMs inside your work, inside your applications to improve the application itself. And so that's ⁓ way more scalable than doing everything by hand. ⁓ You still need to ask human evaluators or yourself for justifying their answer of mainly explaining ⁓ why the answers are good or bad in our ⁓ in your evaluation database. And

Sneha Mehra (00:24:01)  
ideally from varied human experts, just because of what I mentioned with the core palette example where ⁓ two humans have ⁓ subjective biases, and you want to limit that as much as possible unless your application ⁓ relies on following someone's ⁓ IDs and ⁓ intuition. So the goal here is to replicate the decision making of our experts.

It's always the case with LLMs. And then we do prompt engineering, we test, ⁓ we analyze the results, and then we loop all that until we finally ⁓ reach some kind of good results. ⁓ So the process here is to first for the prompt, you ask the LLM to give a pass or a fail according to certain predefined criteria that you find ⁓ exchanging with your experts in the domain.

Ideally, you ask the LLM to justify its answer before giving it. That should help with the results, ⁓ either through a reasoning model or a chain of thought prompting. We can then have a predefined question for a test. We ask the LLM ⁓ who are your current system to answer this question. You have the LLM, ⁓ the most powerful possible ⁓ LLM, here a reasoning one ideally, critique the answer.

According to the predefined criteria that we have. So that's just like a kind of replica of your expert, ⁓ which would be here a judge, ⁓ analyzing your system's response automatically. Then ⁓ it assigns a binary score, pass or fail, following the critiques that it just did. And this would allow to automate this whole process. However, you still need to look at the differences between your

human expert and your automated ⁓ LLM expert. So here it's just about tracking the correlation between the responses from the LLM judge and from the human responses. So even if this is mostly automated, you still need at least 30 to 50 human examples to compare how much the LLM responses resemble the human responses. There are multiple metrics that allow you to track that, like the cohen's

Sneha Mehra (00:26:27)  
Kappa, for example, ⁓ this is basically just to ⁓ measure the agreement between the LLM and the human, and you want to keep following this metric to ensure it's fine, or even if it to ensure it's better and better. ⁓ Also, you obviously still want to ensure that your metrics reflect what user wants, even if it's automated by an LLM, you want it to reflect what the user wants, not necessarily what you want to generate.

here a few metrics is way better than many generic metrics that you may find online. And we want to run these ⁓ manual quote unquote evaluations for each ⁓ more ⁓ significant change, not for each push or for each change of prompt, but for each larger ⁓ update once a week, once per feature, once a month, whatever. ⁓ it depends on your application, but you want to do them ⁓

in a recurring manner to ⁓ ensure there's no drift between the agreement ⁓ between the human and the LLM and ⁓ because it measures the performance of your system in a way better way than the unit tests do. And obviously you always want to keep the human feedback on a portion of the test and compare its alignment with, as I said, 30 to 50 examples to just ensure you're not drifting away from what your real expert would say.

So now there's a third level, but before diving into that, which is our LLM as judge and using LLMs in our evaluations, ⁓ I want to note that we want to evaluate at all stages of our pipelines. ⁓ Whether it is the prompt, which is the input from our user, the retrieval or intermediate values in the system, or even the generations, which would be our outputs. Why is ViCheck not enough? Well, it's because LLMs are not predictable.

As we said, they are probabilistic. Their performance varies depending on the task, the model, the model version, even. And so a prompt that works today ⁓ may not work tomorrow. And ideally, we want to start with obviously the most robust prompt, but also a robust prompt that could ⁓ attempt to fix future error cases. So anticipate possible mistakes.

Sneha Mehra (00:28:50)  
possible errors ⁓ in a future version of the model, integrate clear application constraints, and ⁓ specify the format, the output format as much as possible. So the usual process would be the one ⁓ showed on the right, but typically it's to write a prompt, you test it on a concrete case or a few concrete cases, you evaluate the performance with both subjective and objective.

Evaluations, which is ⁓ using the level one to level three evaluations, and the three will be covered in the next few slides. Then you identify the failures, the edge cases, and everything that didn't work. You modify the prompt, you retest on these several cases, ⁓ you compare performance ⁓ to avoid regressions, so you measure ⁓ your evaluations to ensure that your changes didn't ⁓ hurt in some other ⁓

Aspect of the system. And obviously, you want to keep a version history, whether using Git, Excel, or whatever you decide to ⁓ use, and document well ⁓ everything: the failure, the success, the reason of failure, and reason of success. ⁓ That would be super useful as well. And so for these tests, for the prompts, we want ⁓ inputs from real data, real user data, or

As close to it as possible. You can ask an LLM to generate those if you don't have access to real users yet. We want to cover different scenarios and ⁓ intentions as much as possible, cover our edge cases, situations with contradictory constraints ⁓ from the user, ⁓ malformed inputs from the users if they forget to ask something that they need to ask, if they miss any information. And ⁓

Obviously, you want to annotate everything that's wrong or not with the answer of the LLM here, which, as I said, can be generated synthetically, ⁓ which means via ⁓ using another language model. Before checking the generation, ⁓ we need to know if we are giving the right context. So you want to evaluate retrieval, obviously. And we can do that ⁓ using metrics to verify if we find

Sneha Mehra (00:31:08)  
The most useful information for our LMM. I will go over them quickly, the most relevant metrics, because you can implement them very easily with Lamaindex or Langchain with just ⁓ one command line, it's just one variable. And we will see them ⁓ in the second part of this course with Omar. But basically, what are those metrics that are useful here to measure if we find the relevant information in a retrieval pipeline? So here ⁓ we evaluate our prompt.

Previously, we evaluated our prompt. And so we know that the prompt is good. But then if we want to answer the user, we may need to ⁓ access additional information. And before evaluating if the answer is good, we may want to evaluate if ⁓ we give the relevant context so that the LLM generates the good answer. And so to do that, you can measure the hit rate of the retrieval pipeline, which is basically.

To know if, for example, if you ask ⁓ the system to retrieve the 10 most relevant sources, which here we would refer as as K. So the top K here would be the top 10\. The hit rate would measure if there is a document that is relevant ⁓ in the top K. ⁓ But more precisely, it would measure the percentage of queries for which at least one relevant document was found. So that's just a metric that we can use to.

know if you find relevant documents. You can also use the mean reciprocal rank MRR to know at what average position do we find this right document in our top key. If the hit rate is good, ⁓ you may still not have a good retrieval pipeline because it could be the 10th each time. And so you give nine useless document text chunk to your system for the ⁓ one good document that is only at the at the end.

And so the MRR helps you measure that as well. ⁓ Then you can use the recall ⁓ to know if you retrieve all the relevant document ⁓ that you may have in your database in ⁓ your top 10 documents. That's ⁓ pretty useful as well. You may have the precision ⁓ to know ⁓ if all the retrieve documents are relevant or not. So that helps you disregard or remove the ⁓ uses document that may implement.

Sneha Mehra (00:33:33)  
That may add noise to the system and make the responses worse. There's also ⁓ NDCG or the normalized discounted cumulative gain, which will take pretty much everything into account to ⁓ measure relevance and the order of documents. So it's it's just like the precision, but also taking into ⁓ into account at which ⁓ order in this top 10 responses ⁓ are the good responses. So for example.

NDCG would be higher. If you have five good documents and you try to retrieve the top 10, if these five documents are in the first five positions, the NDCG would be better than if it would be in the five last positions. So that helps you ⁓ better understand if you ⁓ are maximizing your research. And to do that, we also need a data set for the evaluation, the retrieval part of the evaluation. Here, this data set would be.

⁓ Some associated questions and relevant chunks, ⁓ one or more chunks, but basically ⁓ you could have questions that you generate with an LLM. You have, for example, a big database with lots of text chunk that we created in our second course. And ⁓ you want to find the most relevant chunks. And so how do you do that? Well, you you can ask an LLM to generate questions based on the text chunk that you have.

So it would be fake ⁓ or synthetic questions. And then you can measure if ⁓ by asking the same question, your system retrieves the same chunk. There are many useful libraries to help you do that, ⁓ whether it is from Laman index or Langchain again or tons of others, you can find them. ⁓ You can also ask the ask us ⁓ for something specific in your use case.

In this dataset, you also want the version ⁓ of the chunks and indexes with the associated questions or prompts. ⁓ Because chunks evolve over time, you may change the size of the chunk, you may add data to the dataset, you may remove some, etc. You want versioning to ensure you compare apples with Apples. ⁓ Idle, you also want a golden context data set that we see, which basically would just mean that you have.

Sneha Mehra (00:35:59)  
Real data or as close to real user data as possible. So, for example, pairs of questions and chunks containing the answer where the question would come from a user and the chunk is the ideal context that you want to give to the model. Finally, you also want to evaluate the generation, obviously, which is the most important thing to know if the output we give our user is actually good or not. Therefore,

We want to ensure that the answers are relevant, faithful to the context and sources that we give, and that are factually correct. There are three metrics that you can use to do that, but the most important metrics that you want to have, ⁓ as we said, are the personalized ones. So the ones that will track ⁓ what your users actually want. Here we will share just three basic ⁓ general metrics that you should be using just to ensure the generation part is great.

But it doesn't ensure it's great for your users. It just ensures that it makes sense from a general perspective. The first ⁓ metric is relevance, ⁓ which is just measures if the generated answer properly addresses the question asked. It's just to measure if ⁓ the answer uses the right ⁓ context ⁓ to address the query. So you need a good retrieval for that. Then

There's the faithfulness to context, obviously, which based on the context that you give it and the answer, you evaluate if the answer explanation ⁓ is correlated with the sources that you gave it. And finally, there is the correctness, which can be used to measure if the answer ⁓ is factually correct compared to a reference answer or a ground truth that we may have. Alright, so those are all general metrics that you want to measure, ⁓ which you can also.

Just ask Lama Index and Langchain to do ⁓ very easily with just a quick function. ⁓ But as we said, it's very important to build personalized evaluation matrix instead. And so ⁓ how do you measure the matrix and how do you create matrix to measure what the user wants? Well, the classic met the classical method would be to rely on human judgments and to read and ⁓

Sneha Mehra (00:38:23)  
document everything, every exchange with your user and LLM, just to note down what was wrong, what wasn't, ⁓ which is still necessary. You want to do that here and there, but it will be way too long and expensive to do that at scale. And so instead you can use the modern method that we call, which would be to use an LLM as a judge. And here the LLM would replace your human judgment to score answers according to

Define criteria as we mentioned earlier and measure everything that we talked about for generation. This can be automated on large scale with LLMs, obviously. You also want binary scores, as we said ideally. And all this ⁓ should be automated for the most part. And so here, still in level two, the advantages of using an LLM as judge instead of ⁓ doing the manual evaluations.

Would be first that with a powerful system in place for this automated judge, you would have evaluations aligned with human judgment. It would be excellent for detecting hallucinations or subtle errors compared to unit tests, for example. It will allow for rapid iteration on the RAG pipeline compared to having to review everything manually. ⁓ You don't need to wait for your expert to evaluate and to check and give you feedback.

It's also ideal for frequent and low-cost evaluations compared to asking the expert. But there are still some limitations. They can introduce biases specific to LLMs. For example, ⁓ we know that LLMs are susceptible to the position of the information in the text prompt that we send it. So, depending on the LLM, if you send a long prompt and the information that you really want to the ⁓ LLM to know about is at the third of this prompt.

It may be possible that the LLM skips it ⁓ and instead focuses on the information at the beginning or at the end of the prompt. This is a proven fact that LLM do that. And so the position of the text in the LLM is ⁓ in the LLM prompt, in the prompt that you give the LLM, is quite important. And it's a limitation here because this don't really happen ⁓ when you do human manual evaluations. Also, the length of the answers might.

Sneha Mehra (00:40:47)  
Hurt the results as well. ⁓ As we know, the more we give to the model, the more susceptible it is to fail because of the noise that is added in ⁓ the larger context. And there is a final bias where the model ⁓ usually has self-preference, which means that if you use ⁓ a GPT-4 model as a judge, it has been shown that it would generally prefer GPT-4 answers instead of, for example,

Cloud answers. That's mostly because they use the same training schema. And so it sees the generation of the GPT-4 model ⁓ as the same thing they would generate. And they they basically we could say that they think they they they see it as a more sensical answer. ⁓ They are not always reliable. ⁓ For example, sometimes prompts are poorly designed ⁓ and ⁓ this can lead to some

Weird problems or various issues where the LLM isn't as ⁓ confident or ⁓ useful as a human. You want ⁓ scores, but the scores aren't enough. Obviously, you want binary to have more secure ⁓ and ⁓ objective feedback, but it's way better if you have justifications. And here ⁓ a warning you can have justifications asking the LLM, but it's often ⁓ way worse, let's say, than.

The actual human expert, the domain, the human that has the domain expert. But still, you want to ask for the justifications from the LLM to later on use it to understand why it failed or not. But also, ⁓ if you ask the LLM to justify its answer before giving it, it helps it give a better answer by itself providing the context that it needs to give a better answer. And obviously, we also we already mentioned it, but you always want to combine.

This automated feedback with targeted human validation. You want your domain expert to always have a few examples, like 50 or so, that you quickly review to ensure the responses are great. ⁓ Here are some tips for using this ⁓ LM ⁓ judge effectively. First, you want to use very good prompts with chain of thought reasoning or a reasoning model ⁓ if you want.

Sneha Mehra (00:43:13)  
But you you want really good advanced prompts that would make the model detail intensive before responding. You want to give the judge concrete examples of good and bad cases to have ⁓ some kind of few shot prompt and justifications with them. Also, just to note that you want to alternate ⁓ good, bad, bad, good, good, good, etc. ⁓ because if you always say good, bad, good, bad, good, bad, it may

Be biased towards doing exactly that. One good, then one bad, then one good. Just like in exams where you have, ⁓ for example, lots of questions with multiple choices from one to five, and there's four twos in a row. ⁓ You are not inclined to selecting two for the next question. That's pretty much the same thing with LMs here. ⁓ They will be biased by the structure ⁓ you give it. So just be careful and mindful of that.

you want to ask for a binary score or compare two outputs against each other if it's too difficult to come up with a good binary score, pass or fail. This comparison is called pairwise comparison, and it's pretty good for subjective comparison where you give you ask the model to generate two or three versions, and then your judge would evaluate these two or three versions to just find its favorite one and explain why.

⁓ By the way, you also want to provide an equality ⁓ option in this case, just because it's definitely possible that the two answers are ⁓ completely reasonable and provide value. ⁓ So you obviously want to favor pass or fail judgments with detailed critique rather than a more subjective feedback. Then you want to iterate under prompt until you converge with a high correlation rate with the domain expert.

And just note that for subjective tasks, you may want to prefer pairwise comparison that we just described. And for objective tasks, you obviously want to go with direct scoring ⁓ that can be more reliable. And now, how do we build a good eleven judge? The first step is to create a representative data set with your outputs to evaluate. So you start, you have your inputs, which is your prompt, then ⁓ expected outputs of the system. So the ideal case.

Sneha Mehra (00:45:38)  
Then you want to generate your answers. And finally, you want the human annotations for these answers to measure the failed past conditions and check for the critique. Ideally, you want to include here real user cases, as I mentioned, and ⁓ ask several experts, not just one to annotate, to reduce biases. You also want to aim for case diversity and not just ⁓ known obvious errors. ⁓

The second step is to have a domain expert judge a sample. You can start with, as we said, 30 to 50 well-annotated examples and then annotate them, pass or fail, with critique from the expert or yourself to form the basis of the prompt. ⁓ Based on these 30 to 50 annotations and critiques, you can use them to create the best prompt possible for the judge. After that, you adjust the prompt, trial and error loop.

for the LLM judge to reproduce the judgment of the expert as much as possible. After that, you compare the LLM's decisions with human judgment and repeat the steps four to five until ⁓ convergence on the whole database. And why does it work? Because it allows for outsourcing, business reasoning, or the decision-making process of the domain expert that we have to an LLM, which is basically what they are good at.

If they have the right data, it facilitates error analysis and root cause investigation. ⁓ It helps standardize evaluation criteria within your team. ⁓ There's less ambiguity than vague grids with scores of 0 to 10 or 0 to 100 with these binary evaluations. It brings out users' implicit expectations, while the explicit expectations are in unit tests or more easily defined.

And it can also serve as guardrails in production ⁓ once you have a good ⁓ judge in place. You can use it later on, as we will see. ⁓ Finally, there's a third level that I will go very quickly over. It's the A-B testing that we all know about. It's the same for ads or other products or software. You basically want to compare models, compare versions of prompts or whatever ⁓ one at a time.

Sneha Mehra (00:48:02)  
And then ⁓ use them live. ⁓ show both versions to real users. Well, evaluate both versions first, obviously. And if you are not sure which one is the best, you just push both in production to compare what seems best with your users. And this obviously is reserved for a more major product once it's in production. Okay, so this was a lot. We discussed these three levels, going very quickly over the third one.

⁓ Let's do ⁓ a short recap just because it may be redundant, but it's crucial information to know and to ⁓ to not be afraid of because it's the only way you have to actually know if your system is good or not, and on what task to do next. So it's super important, and everyone in the team should know about the evaluations, mostly because they need to tune in.

and especially a Drummond Expert need to be working closely with you as the developer.

So, to do a short recap, our datasets here ⁓ are very important because they allow us to measure performance quantitatively, to compare versions of our code, to identify the edge cases and the errors, to prioritize the next improvements to work on, to automate quality and monitoring, which is something quite nice to automate. And here, a good dataset ⁓ must absolutely include the real or realistic users' questions, as we mentioned. ⁓

The golden data set, ⁓ the expected chunks ID. If you are working with a RAG pipeline, you want to know ⁓ what type of data should be returned in ⁓ for which query. ⁓ You also want, obviously, the expected answers, which would be the ground truth of your ⁓ system. You want to know the versions of each prompt, model, chunk, ⁓ or ⁓ your code in general, just like in regular software development. And the the

Sneha Mehra (00:50:08)  
Better the dataset, ⁓ the faster your system can improve. And optionally, but ideally, you want justifications for these expected answers, either from your LLM judge or from your human experts. The better this data set is, the better your system can improve. So this is very crucial. Here's a short strategy to create ⁓ your dataset.

The initial phase would be to generate synthetic questions from the chunks that you may already have with an LLM. This is to quickly launch your evaluation and already start evaluating. Then you want to gradually replace those generated questions with real user queries, ideally 50 or so, and associated chunks that you know will answer this question. Then you will manually annotate the correct answers and sources, which would be more reliable and allow for.

More representative data. So this is just to ensure that your golden data set from your real users also have a good ground truth with good explanations of why the answer was good, etc. The dataset must be improved over time, obviously, with feedback loop and as the product evolves. So that's why we have an initial phase and an advanced phase. And you don't have to wait to have real users to start monitoring, tracking, and evaluating.

You start right away with an suboptimal ⁓ evaluation pipeline that you gradually improve along with your product, which will, even if fully synthetically generated, help you understand your current system and its weaknesses. We know that's it's long and boring, but building your evaluation data set ⁓ is the most important part of the pipeline else.

You will just work in circle and work on many improvements without knowing what works or not. We also talked about a critique a lot and giving feedback. So ⁓ this is just a quick example that you can ⁓ make ⁓ press pause and read if you want. But basically, ⁓ a critique is just an explanation of why was it a success, a pass, and ideally ⁓ even contain how to do better.

Sneha Mehra (00:52:31)  
So the critique would be just the perfect reply from your expert that would measure how good was the AI response and how could it be even better. Likewise, when it's a fail, when it's not a good response, ⁓ the critique would say why and I identify the key missing elements ⁓ to become a good response instead. And we will discuss this further in session five.

But I just want to know that it's super important to have a good UI to monitor what happens over time. Even for these critiques, failed pass tests, you want your domain experts to look at these data and look at how your model behaves. Just because the domain expert, which is not necessarily a programmer, will not be in the loop very often, usually. And so there can ⁓ be a drift between.

What the expert says and would have said versus what the developer or yourself ⁓ work, how how you work on the prompt and edit the prompt and make changes to the whole system. And so this can create a drift between what the expert expects and ⁓ would answer versus what the how the model actually behaves. And this is just to say that you want this expert to look at your ⁓ system and your system's answers.

As often as possible, which means that you have to build a nice UI for him or her to have a look and to ⁓ enjoy looking at it. So for example, you don't want to send the expert just your JSON file with lots of examples. You want to show it ⁓ at least in a clear Excel sheet or something that is more approachable to ⁓ regular human beings.

That's very important to keep in mind because your expert will be the key factor to make your application succeed with proper evaluations. And just to motivate you to build this LM judge and this better evaluation pipeline, ⁓ it's not just a one-time use. The the judge will be super useful ⁓ throughout your whole lifecycle just to measure the improvements and the system as a whole, but also you can even use it.

Sneha Mehra (00:54:54)  
In the future to train or fine-tune your model if you have to do that. ⁓ That's through something called RLAIF, which is ⁓ basically the same thing as RLHF that we discussed in the past, but uses an artificial intelligence judge, so an LLM, to help during the fine-tuning to give proper feedback and review rather than a human expert. And if you have built your judge properly,

Well, you can use it to ⁓ replace ⁓ your human judge. So that would be quite beneficial to have an already working judge ⁓ in this case. And can be used in many things, ⁓ as we will see in course five, to, for example, ⁓ act as guardrail within your live application, just to ensure your system actually answers properly to the user and to the proper requests, etc. ⁓

Okay, so that was a lot. It was pretty much all the theory that we wanted to share about evaluation as a whole. I will now leave Omar to share a bit more of a practical view ⁓ of some ways to build evaluations with our notebooks to you. ⁓ Hello everyone. So in this second part, ⁓ I will be showing some ⁓ code implementations of what we just talked about in the first part of the session.

So I will showcase ⁓ and notebooks around how to evaluate a very simple chatbot, rag chatbot. Then I will try to show you how it looks like when we try to do a more comprehensive iterative optimization process. So a more like a ⁓ like a scientific approach to optimizing a ⁓ a system like this. And then

⁓ finish this session on ⁓ how to evaluate not the retrieval part but also the final answer part of is the answer good enough or not. So yeah let's go let's start. So here I have the notebook ⁓ that I will show you. So if you remember in the last session, I showed how to create very basic, very simple ⁓ rag chatbots using

Sneha Mehra (00:57:21)  
LAMAINDEX THRAMERWEC. ⁓ And it's actually quite easy to do it with LAMAINDEX. We you just give it ⁓ data, some documents, and then with a few lines of code you can have a vector store index. So here I'm not going to go into how to repeat how to create one, but here I'm just actually creating a ⁓ a ⁓ a chatbot again. ⁓ The only thing that is new here is that

We are adding a function that will set the IDs in a deterministic form, the IDs of each chunk in the database. Because if you don't do that, then each time you create a new database, a new vector database, then Lama index just randomizes the ID. ⁓ And in this case, since we want to evaluate the database with the same questions with the same data, basically.

If we restart the notebook and we create the database again, we want to keep the same IDs as we had before. So we don't, yeah, we don't get ⁓ mismatch ⁓ on the IDs basically. Here, this is just a different way to create a vector store index. Here we use the ingestion pipeline. We just ⁓

specify what model to use, which vector store to use, and then the documents to use.

Then here I specify the model I want to use. So Gemini 1.5 flash. ⁓ And ⁓ then ⁓ I can ⁓ create a query engine using this LLM and the index. And I want the chatbot to use five chunks to enter me. So it's going to retrieve the highest five ⁓ nodes, chunks, that the system can retrieve. And if I ask

Sneha Mehra (00:59:22)  
question the chatbot is able to answer me correctly using the information within the ⁓ vector store database.

Now, for ⁓ evaluating the system, I want to make sure ⁓ the system is actually able to retrieve information correctly, the correct chunks, the correct nodes. So here, since we don't have a dataset, an evaluation dataset, we are just going to use DLLM to come up with questions to use. ⁓ And we are basically going to show every chunk in the database to the model.

⁓ And ⁓ we are asking the model to come up with a question for that chunk specifically. So ⁓ at the end, we will have questions directly related or associated with ⁓ a specific chunk. So when we ask ⁓ one of the questions that we have in the dataset, we actually expect to see that chunk.

That ⁓ the question was created from, ⁓ we expect to see that chunk in the sources retrieved by the chatbot, the system. ⁓ So this is what we are doing here. This is the prompt that we are using to create the questions. So here we put the chunk text ⁓ and we ask the model to be a professor to come up with.

A specific number of questions. So for example, one question. And ⁓ if we decide to have two questions per chunk, then these questions should be diverse and ⁓ use all the document. ⁓ And ⁓ also the questions, ⁓ the answers should be all within the chunk. They shouldn't use a different context.

Sneha Mehra (01:01:29)  
So, this is the prompt to create the data set, the questions. ⁓ So, this was to show you what was the prompt. Here we are actually using the prompt. So, we have this function here: generate question context pairs. ⁓ And ⁓ yeah, here we are just looping ⁓ through ⁓ our chunk database for all the chunks in our database. And here, for example, I'm

Defining the model to use that will come up with the questions. So here it's Gemini 1.5. And here I'm I'm actually not going to use all of the nodes in my database because there are too many of them. I'm just going to use 25\. And I want ⁓ one question ⁓ for all of these 25 chunks. ⁓ And so that's how I am creating my retrieval.

Evaluation data set. This is just for the retrieval part. I want to see if my system is actually able to retrieve the chunk, the specific node that I that can answer my question or the generated questions. And so ⁓ once I have the data set, then I can ⁓ specify the metrics. ⁓ Here ⁓ we are going to use two metrics. ⁓ One is the hit rate. So

The hit rate measures ⁓ if the ⁓ actual chunk, the correct chunk, is in the list of retrieved sources, retrieved chunks. ⁓ If the correct ⁓ chunk is retrieved, then it's ⁓ it's good. For that question, we have a good ⁓ result. And we are going to do this across the data set and then compute the hit rate to know for how many questions we retrieved.

the correct document.

Sneha Mehra (01:03:28)  
Then we have the MRR, which ⁓ measures or tries to give us the score or the average position of the correct chunk in the list of sources. So here I show that if the correct chunk is in the basically in the third position, then I'm going to have a score of 1 over 3\. And of course, the goal is to get to the first position because as we mentioned.

In the first session, the LLMs like to read, like to understand the first part of the context and maybe the last part, but not the middle part. So we want to make sure that the correct information, the relevant context is in the beginning of the context, the beginning of the prompt or at the end basically. So either at the beginning or or the end. And here we're just measuring the ⁓ the MRR to know how the system is.

Working. So here this function will help me ⁓ display the results. And here I am evaluating. So here I want to evaluate, ⁓ and at the same time, I want to change some configuration ⁓ values. So here the only thing I'm changing is the number of chunks I retrieve. So we can see that for two chunks that I retrieve.

This is the score. So we have a hit rate of 72%. So for 72% of the questions, I have the correct information, the correct chunk retrieved. And the MRR is also at is closed at 72\. Then we can see that for the last value of 10, we can see that the hit rate is at 96%. ⁓

So this is almost 100% correct. And it makes sense ⁓ that if you retrieve many more chunks, that the probability of retrieving the correct chunk is actually higher. It makes sense. But the goal is not to retrieve too much content because then you're trying you're giving the model actually incorrect information. ⁓ So you're increasing the context.

Sneha Mehra (01:05:54)  
Length, you're increasing the context size with maybe irrelevant information. ⁓ But the the actual answer is in that context. But as we know, the the models need optimized context. Like if you give it ⁓ a million tokens of context, then of course the answer is in the context, but is the model going to be able to?

answer and to find the inf the correct information in that context. So ⁓ this is why we don't want to put too many tokens in the in the prompt because then the model is full with ir irrelevant information and that's not something good either.

So here I'm just showing that you can also use two other metrics. So you have the faithfulness and relevancy metric. So here we are actually using a different model. So in this case, we're going to use either GPT-4.0 or GPT-4.0 mini to tell ⁓ if the answer is faithful or relevant to the question. So if if it's faithful, then

It means that the answer ⁓ is not hallucinated, that the answer actually uses the context. And the if the answer is relevant, it means that it actually answers the user question. The the answer. So this is what we we measure here using these ⁓ the the the models. ⁓ And if I ⁓ change the number of chunks again ⁓ from two to ten.

We can see that actually GPT-4.0 Mini thinks that it's not GPT-4.0 Mini, it's ⁓ for the faithful, it's actually GPT-4.0. It it thinks that it's actually quite faithful and relevant. So we don't need to worry about the answer. We just need to worry about the retrieval part, which usually usually is the most ⁓ hard to optimize and ⁓ the most important part actually, the part where we need to retrieve the correct information.

Sneha Mehra (01:08:07)  
Is actually harder than if the model is actually able to use the prompt, the information in the prompt to answer the question. So that's why we see high scores here, because the answers are basically all correct or faithful and correct. ⁓ So I'm going to skip that and just explain you how a more comprehensive process looks like, more like a scientific approach, let's say. ⁓ So here I'm

Pulling up a lesson from ⁓ one of our courses, which you can check out. But I thought it would be way more easier to ⁓ show you how this process looks like this way, because it's actually ⁓ yeah, it's it's not that it's complex, it's just that it's there there's many important parts I wanted to show. So here for context, we have the ⁓ AI tutor chatbot.

Which ⁓ should be able to ⁓ retrieve information to answer student questions. And I want the chatbot to retrieve information from all of these different libraries, like transform hugging face transformers, PEFT, TRL, even Langchain, Lemma Index. I want to have in my vector database all of those documents.

And I want the chatbot, the system to retrieve those documents to answer questions to the students. And so here I just want to show you how optimizing this chatbot can look like. So here, like I showed, ⁓

—---------------------------  
Sneha Mehra (00:00:09)  
Welcome to this third session of the course, where in this one we will cover evaluations, which we think is the most essential thing to know in the field, mostly for a few reasons that we will cover in the first few slides. The first one being that without evaluations, we don't know what to do and where to go next. ⁓ This is a direct continuation of the first two sessions where we explored the limitations of LLMs and then how to build on top of them.

So when we are at the step of building on top of them, we need to understand what's happening and understand where where to go next. So that whole session is about this exact goal of helping ⁓ you understand what to do next and how to do so. So that's through evaluations, which we'll cover at different levels and different ⁓ complexities, let's say. So today's plan is to cover the necessary theory to know that we believe is most important.

Just like the usual courses, the two previous courses, we will start with a first hour where I will cover the why, when, and how to evaluate with different levels of evaluations and at different steps, ⁓ whether it is through ⁓ evaluating your prompts, your retrieved information, or the generation part. We will also talk about ⁓ one of the most interesting steps of evaluations, the LLMS judge or the human feedback part.

of it. And then Omar will continue with more practical and code examples with two specific notebooks that we built for this course, the evals for retrieval and the evals for LLM responses, which is a notebook with an LLM as a judge example. You can get both of these notebooks in the description of the course and in the slides that we provide. ⁓ If you want to run them throughout, they are colour notebooks, so you can just ⁓ execute them

Right away if you enter your ⁓ OpenAI API keys. And so let's start with why evaluate. The first reason is because making a prototype is pretty easy, pretty fast. You can build ⁓ a quick chatbot in 10 minutes, but optimizing it will take days or even months to get ⁓ to the level of performance that you want. You may even never get to that point if the task is too complex. So we want to evaluate if it's feasible and

Sneha Mehra (00:02:38)  
How much we are improving ⁓ towards our goal. Second point is because ⁓ AI isn't magic. It's just like software development and it and it requires the same rigor, if not way more, because you are playing with, you are working with much more powerful systems ⁓ that are not completely controlled by you. Just like, for example, using OpenAI's GPT models, you don't.

Really fully understand or know what's behind these APIs that you're using. Also, evaluations allow for many things, including debugging, doing relevant and continuous improvements, aligning with real user needs, having confidence in our model or our system as a whole, the fight-finding the edge cases and silent failures that we may not think about a priori, and it helps with what we call the March of 9th.

So this is basically going from ⁓ the easy ⁓ quote unquote 90% accuracy that we can get with the 10 minute prototypes and then to 99%, 99.9%, and etc. until we basically never reach 100%, but 99.9999. So it helps towards that route because it's impossible to know the real progress you are making ⁓ without evaluating properly your system.

Here is an example also ⁓ for ⁓ why evaluate. ⁓ LinkedIn first developed a chatbot for ⁓ a skill fit assessment, which they implemented, I think, in 2023, pretty early on, and they discovered that the users didn't really appreciate the chatbot. And why? ⁓ Well, it's because they built the chatbot to answer ⁓ the users to know if they were a good fit for the role or not.

And so they didn't have really good evaluations in place, and they just built this chatbot to be able to find ⁓ if ⁓ from the user's resume and from the job description it was a good fit or not. But then the chatbot answered all users ⁓ if they were a good fit or not, and just saying, for example, you are a terrible fit. And the users were quite disappointed, they didn't like that. And so that's just to say that.

Sneha Mehra (00:05:02)  
Even if you build something that works really well, it says if they are a terrible fit or not. It might not be what the user wants. And you cannot really know that first without asking the users, but also without ⁓ building the evaluation system that will evaluate if ⁓ your system replies the way the user wants. And so this wouldn't have happened if they would have surveyed the users first and then built evaluations that would ⁓ assess if.

The chatbot answered the way they want, such as instead of saying you are a terrible fit when ⁓ it's the case, ⁓ saying that ⁓ you are probably not the best fit, you would need to work on this, on that, etc. Basically, the people wanted to know why they weren't a bit a bet ⁓ a good fit and how to become a good fit. So that's ⁓ basically the the use of evaluation is basically to

Understand and know ⁓ how to bridge the gap between what the user wants and what you offer. And you can just know that by ⁓ evaluating your system and doing that ⁓ many times until you see a convergence. ⁓ Another example here of why evaluating is because ⁓ for this chatbot, LinkedIn mentioned to achieve 80% of the experience that they wanted in the first month.

But then needed an additional four months to surpass 95%. So this is just to demonstrate the March of 9 that can be pretty long, pretty exhaustive. And ⁓ to go faster towards that route, you want to evaluate your system at every update to know directly if this update helps or not, instead of walking around and testing many things without necessary improvements.

It's also ⁓ pretty essential just because if you are evaluating at each step and ⁓ documenting this evaluation and better understanding what led to improvements or ⁓ downgrades in your system, if someone else comes into your team or if the team changes, the new person will know what has been tested, what worked, and what to do next. So that's pretty essential in a team that changes, or even just for yourself, to not

Sneha Mehra (00:07:28)  
Try the same things over and over again, or to do many changes and don't know what affected what. ⁓ When to do evaluations? Well, pretty much at each step, ⁓ at each iteration of the pipeline. ⁓ Whether you update the model version, you change the model, you change the prompt just a bit, you change the database, the types of chunks that you have, the re-ranker, the embedding model, whatever you do or change, ⁓ you want to evaluate to know.

How did it affect the the whole system?

Secondly, you want to compare objectively or do A-B test between ⁓ the system, the two parts of the system, or two models, or whatever. If you do these tests, you want to evaluate as well, obviously. So if you compare ⁓ GPD 4.1 to Cloud 4, you want to know which one does best and how. Well, that's only through evaluations. Also, obviously, you want to do that before each big push or each deployment ⁓ or

Do that in production as well from time to time in case there is ⁓ what we call model drift or any kind of drift between what your system does versus what the ⁓ user wants. And so you do evaluations lots of time and pretty much all the time. So there's a good need for automating this whole evaluation system as much as possible. And we will give a few tips for that during these ⁓ next two hours.

So ⁓ we want to quickly iterate. ⁓ we want to evaluate often, ⁓ just like as in regular development, we want to continuously evaluate quality, ⁓ quickly debug small changes. You don't want to make tons of prompt changes and then don't know what caused what. And unfortunately, many companies focus on fine-tuning and building very huge prompt right from the start, where they

Sneha Mehra (00:09:28)  
Don't even have a baseline or know where to ⁓ go to make improvements from their application. And as I said, without evaluations, ⁓ it's pretty easy to go in circles. ⁓ There's also benchmarks that we mentioned in the previous courses. And ⁓ while they are important, they are not the best ⁓ way to measure performance in your system. So when ⁓ do you want to use benchmarks or why?

use benchmarks. Well, it it's because it's it allows for objective comparison with between different approaches. So just to select between two models or two ⁓ chunking techniques or whatever benchmarks can be useful. They provide a common reference for evaluating the pipeline when even common amongst ⁓ everyone in the world. They help validate that the system works in

General cases, but not maybe your own cases. And ⁓ there are ⁓ numerous ⁓ number of options that you can use ⁓ through Lama Index or whatever library you may want to use. But benchmarks are not enough. They don't know ⁓ your domain, your product, your users. Public benchmarks are pretty much ⁓ generic. They won't be measuring what you want to measure.

They capture neither the specific use cases or the edge cases in your application. They are a good start. For example, as I mentioned, to know which LLM might be best for now, ⁓ to just to start with, but they are not the best way for optimizing the system. They also do not measure the right metrics that we want to track. The solution is to move towards personalized evaluations that you build yourself, ⁓ first that understand.

Our or your users and their needs. And to do that, you first need to understand the users and their needs. So you need to look at the queries in your systems. As soon as you deploy it, ⁓ you want to keep track of that and look at them. And beforehand, you may just want to test it yourself and take the time to read the answer and what's wrong or what's not perfect in them. You also want to define your important metrics because there's no

Sneha Mehra (00:11:54)  
Two applications that use the same metrics, you need to find the ones that you want to track and improve on. You want to create a data set, which we will see shortly. You will want to integrate domain experts when creating a data set and ⁓ at pretty much every step of the way here, ⁓ especially in your evaluation loop, where ⁓ we will cover it, but basically you want the person that knows the best about the domain of the chatbot to

Check, analyze, and give feedback to the model. You finally also want to integrate user feedback, obviously. ⁓ So let's cover the multiple steps, ⁓ the evaluation process between a proof of concept and after a proof of concept. So typically for a proof of concept, you often start with just vibe checking. You try to have the clearest prompt possible, the answers make sense, and so you refine the prompt.

Look at more results. That's perfect. It's super useful and it's essential to do. You need to look at the results, refine the prompt, check it, read it, look at the data you have and everything. But in this case, everything is done approximately without real comparisons or real metrics. It's based on your subjective impressions. It affects several factors at the same time. Either you change many things in the prompt or the data ⁓ that is retrieved.

As well as the LLM and the prompt, ⁓ you don't measure what does what. So that also leads to a lack of reproducibility and standardization. Just because, as I mentioned, if someone new joins the team, they won't know what you have tested or what worked or not, and they will just redo the same mistakes. And that's ⁓ for proof of concept. And during and after this proof of concept, we want to transition towards objectivity.

So we need to have examples of what we want to generate and see if our results match ⁓ what we want. So we will compare the current outputs of our systems to ideal answers, which are our examples in our database. And we want to keep the good and the bad examples and note pretty much everything that we can find about them. Is the tone good? Is the answer complete? Is there anything you you want to have in a typical answer?

Sneha Mehra (00:14:22)  
And then you want to ask your expert or your users ⁓ what they need in a perfect answer and break it down into ⁓ potentially a checklist to then use that checklist to give a manual score according to the checklist by experts or yourself. So, for example, instead of just asking the expert if the answer is good or not and explaining why, which is definitely useful and crucial, you want the why, this won't really help you build.

Ideal prompt. Instead, you may want to have some kind of checklist that makes the answer actually good. And you can use an LLM for that if you, for example, ask your expert to describe why the model's answer is good and why it's bad for many examples. You can then ask an LLM to build a checklist out of these ⁓ feedback to find just true and false or yes or no checklist points.

That would make ⁓ a good response. If, for example, the answer needs to mention the username and then needs to conclude with thank you, whatever, those are all things that you can put in a checklist to see if ⁓ it's there or not. And based on how many you have out of the ⁓ 10 points in the checklist, you would have a ranking out of 10\. So that would be quite easy ⁓ to have more of a quantitative node.

Even if it's a somewhat subjective answer. And then we can try to automate this with some kind of system that I just mentioned, but a bit more advanced that we'll see in a few slides. And then after the proof of concept, we really want this type of quantitative data even more. ⁓ We want to use standardized metrics in addition to the subjective ones. We want to use benchmarks if they are applicable, but that's rarely the case.

We will obviously be creating our personalized evaluations where the goal is to prevent problems, resolve limitations without ⁓ creating new ones, obviously, and also understand our system's performance. And so these personalized evaluations that we will build will allow for many things that we already mentioned, but quickly again for doing A-B testing, understand what to work next, what to work on next, and allowing for optimizing the system's.

Sneha Mehra (00:16:49)  
One thing at a time, knowing what ⁓ caused what. And so ⁓ they allow to create a continuous improvement loop and align the matrix that you track with your users. But how do we do those evaluations? ⁓ There are different types of evaluations, ⁓ and here they are ⁓ in order of complexity, ⁓ difficulty of implementation, and cost. So the first level is the unit test.

They are implemented at all stages. ⁓ First, the data preparation. For example, you can have just ⁓ quick if cases to ensure that all documents have titles, text, and a source before the ingestion. ⁓ At the prompt or code level, for example, having regex conditions ensuring ⁓ no IDs are ⁓ disclosed. ⁓ Or you can even have them at model changes. For example, you could have a simple Boolean rule in place.

To ensure that the SQL data is actually returned correctly. ⁓ For example, just ⁓ comparing if the number of products for a client returned is the actual one in your database. Next, you want to create those test cases and with test data for each of these unit tests. ⁓ Those can be generated directly with an LLM. And you want to make these tests as difficult as possible.

While still following what the user would want, obviously. ⁓ Also, the pass rate depends on the performance you want. ⁓ It's not necessarily 100% that you may ⁓ want to ensure to go to the next step or to ⁓ push the deployment. For example, ⁓ you know about ⁓ DALI and Midjourney. ⁓ they are ⁓ image generation systems and they are super popular, people love them.

Yet you may have to send ⁓ the same prompt five to ten times to get one good image. And so that's an accuracy of ten to twenty percent. And still it's a huge success. So that's just to share that it's not necessarily a hundred percent or perfect accuracy that you may want to have.

Sneha Mehra (00:19:05)  
Also, here you obviously want quantitative results, a percentage per error type. That would be ⁓ super useful. And you want to execute and monitor these tests. Here, for example, we have ⁓ two steps ⁓ of the model changes where we show the different errors that we have grouped by type of error. And for example, we see on the left that there's a big ⁓ yellow part which is related to empty responses, and so we could.

⁓ understand that this is a problem in our system, work on that problem, ensure ⁓ there is no longer empty responses. And then in the second step, you see that this type of responses and percentage disappeared. And so that's just to show that monitoring, checking, and evaluating can help fix issues that we didn't even know we had, just like empty responses here that were ⁓ grouped together. ⁓ And so, for example, you may want to.

Use GitHub Actions and continuous integration to collect the results ⁓ and current state of the prompts and the whole system and test to see your progress over time. And you also may want to set up visualization tracking ⁓ right from the start, ⁓ as we will see later on in ⁓ the next lessons, because it's quite easy to get that ⁓ started. And obviously, we want to run these unit tests every time the code changes. After unit test,

We can move on to more advanced ⁓ evaluations with manual evaluations ⁓ that ⁓ cannot be tested using conditions like regex or simple Boolean functions. And they are still important to log traces and to look at our data. There are multiple tools that do that, like LangSmith and Langwatch. You can use whatever you may want. We also want to ⁓ classify the type of error as we did with the unit tests.

Which will be super useful later on. And then with ⁓ this kind of manual evaluation, even if it's very painful to do, we can then use the errors corrected by humans to do fine-tuning or add, modify, or improve our prompts. So that they are extremely useful in many situations, even though they are long to do, because you do manual evaluation. So you read the prompts and the responses and you give feedback yourself or the expert in that case. But that's crucial to do.

Sneha Mehra (00:21:34)  
And here again, we want scores. ⁓ Ideally, we want binary if it's a fail or pass, because it's ⁓ much more consistent and less subjective than just a score out of a hundred or ten. ⁓ For example, to illustrate that, we used to have a project related to color palettes where we basically wanted to build a system that can say if a color palette of five color ⁓ is beautiful or not. We initially asked designers that we had in our hands.

To rate palettes from one to five ⁓ on how beautiful they were. But then we discovered that just by analyzing the rating data, that there were lots of variability between two raters for the same palettes and even between the same rater for the same palette. So that's quite problematic. And it's because ⁓ it's very unclear to know what's a three or a four and what's a two or a three, just like when you go to the

physiotherapist or or whatever, and they ask you ⁓ how much does it hurt from one to ten? what makes it a six or a seven? It's it's pretty hard. And so it's the same here. ⁓ And this is why you want to rate in binary as much as possible. And that can be done, as we mentioned, through some kind of checklist or just a true or fail, a pass or fail rating that is already better than just a subjective score.

Here we will have our domain experts, which is either ⁓ the doctor or whoever has the expertise in our company or in our field, or ourselves if that's the case, to ⁓ score these results. But ⁓ this is very tedious, and you actually don't really have to do all that 100% manually. You can instead do that using LLMs. ⁓ Yes, we can automate them with LLMs.

So that's pretty cool because you can use LLMs to improve your work. And you can also use LLMs inside your work, inside your applications to improve the application itself. And so that's ⁓ way more scalable than doing everything by hand. ⁓ You still need to ask human evaluators or yourself for justifying their answer of mainly explaining ⁓ why the answers are good or bad in our ⁓ in your evaluation database. And

Sneha Mehra (00:24:01)  
ideally from varied human experts, just because of what I mentioned with the core palette example where ⁓ two humans have ⁓ subjective biases, and you want to limit that as much as possible unless your application ⁓ relies on following someone's ⁓ IDs and ⁓ intuition. So the goal here is to replicate the decision making of our experts.

It's always the case with LLMs. And then we do prompt engineering, we test, ⁓ we analyze the results, and then we loop all that until we finally ⁓ reach some kind of good results. ⁓ So the process here is to first for the prompt, you ask the LLM to give a pass or a fail according to certain predefined criteria that you find ⁓ exchanging with your experts in the domain.

Ideally, you ask the LLM to justify its answer before giving it. That should help with the results, ⁓ either through a reasoning model or a chain of thought prompting. We can then have a predefined question for a test. We ask the LLM ⁓ who are your current system to answer this question. You have the LLM, ⁓ the most powerful possible ⁓ LLM, here a reasoning one ideally, critique the answer.

According to the predefined criteria that we have. So that's just like a kind of replica of your expert, ⁓ which would be here a judge, ⁓ analyzing your system's response automatically. Then ⁓ it assigns a binary score, pass or fail, following the critiques that it just did. And this would allow to automate this whole process. However, you still need to look at the differences between your

human expert and your automated ⁓ LLM expert. So here it's just about tracking the correlation between the responses from the LLM judge and from the human responses. So even if this is mostly automated, you still need at least 30 to 50 human examples to compare how much the LLM responses resemble the human responses. There are multiple metrics that allow you to track that, like the cohen's

Sneha Mehra (00:26:27)  
Kappa, for example, ⁓ this is basically just to ⁓ measure the agreement between the LLM and the human, and you want to keep following this metric to ensure it's fine, or even if it to ensure it's better and better. ⁓ Also, you obviously still want to ensure that your metrics reflect what user wants, even if it's automated by an LLM, you want it to reflect what the user wants, not necessarily what you want to generate.

here a few metrics is way better than many generic metrics that you may find online. And we want to run these ⁓ manual quote unquote evaluations for each ⁓ more ⁓ significant change, not for each push or for each change of prompt, but for each larger ⁓ update once a week, once per feature, once a month, whatever. ⁓ it depends on your application, but you want to do them ⁓

in a recurring manner to ⁓ ensure there's no drift between the agreement ⁓ between the human and the LLM and ⁓ because it measures the performance of your system in a way better way than the unit tests do. And obviously you always want to keep the human feedback on a portion of the test and compare its alignment with, as I said, 30 to 50 examples to just ensure you're not drifting away from what your real expert would say.

So now there's a third level, but before diving into that, which is our LLM as judge and using LLMs in our evaluations, ⁓ I want to note that we want to evaluate at all stages of our pipelines. ⁓ Whether it is the prompt, which is the input from our user, the retrieval or intermediate values in the system, or even the generations, which would be our outputs. Why is ViCheck not enough? Well, it's because LLMs are not predictable.

As we said, they are probabilistic. Their performance varies depending on the task, the model, the model version, even. And so a prompt that works today ⁓ may not work tomorrow. And ideally, we want to start with obviously the most robust prompt, but also a robust prompt that could ⁓ attempt to fix future error cases. So anticipate possible mistakes.

Sneha Mehra (00:28:50)  
possible errors ⁓ in a future version of the model, integrate clear application constraints, and ⁓ specify the format, the output format as much as possible. So the usual process would be the one ⁓ showed on the right, but typically it's to write a prompt, you test it on a concrete case or a few concrete cases, you evaluate the performance with both subjective and objective.

Evaluations, which is ⁓ using the level one to level three evaluations, and the three will be covered in the next few slides. Then you identify the failures, the edge cases, and everything that didn't work. You modify the prompt, you retest on these several cases, ⁓ you compare performance ⁓ to avoid regressions, so you measure ⁓ your evaluations to ensure that your changes didn't ⁓ hurt in some other ⁓

Aspect of the system. And obviously, you want to keep a version history, whether using Git, Excel, or whatever you decide to ⁓ use, and document well ⁓ everything: the failure, the success, the reason of failure, and reason of success. ⁓ That would be super useful as well. And so for these tests, for the prompts, we want ⁓ inputs from real data, real user data, or

As close to it as possible. You can ask an LLM to generate those if you don't have access to real users yet. We want to cover different scenarios and ⁓ intentions as much as possible, cover our edge cases, situations with contradictory constraints ⁓ from the user, ⁓ malformed inputs from the users if they forget to ask something that they need to ask, if they miss any information. And ⁓

Obviously, you want to annotate everything that's wrong or not with the answer of the LLM here, which, as I said, can be generated synthetically, ⁓ which means via ⁓ using another language model. Before checking the generation, ⁓ we need to know if we are giving the right context. So you want to evaluate retrieval, obviously. And we can do that ⁓ using metrics to verify if we find

Sneha Mehra (00:31:08)  
The most useful information for our LMM. I will go over them quickly, the most relevant metrics, because you can implement them very easily with Lamaindex or Langchain with just ⁓ one command line, it's just one variable. And we will see them ⁓ in the second part of this course with Omar. But basically, what are those metrics that are useful here to measure if we find the relevant information in a retrieval pipeline? So here ⁓ we evaluate our prompt.

Previously, we evaluated our prompt. And so we know that the prompt is good. But then if we want to answer the user, we may need to ⁓ access additional information. And before evaluating if the answer is good, we may want to evaluate if ⁓ we give the relevant context so that the LLM generates the good answer. And so to do that, you can measure the hit rate of the retrieval pipeline, which is basically.

To know if, for example, if you ask ⁓ the system to retrieve the 10 most relevant sources, which here we would refer as as K. So the top K here would be the top 10\. The hit rate would measure if there is a document that is relevant ⁓ in the top K. ⁓ But more precisely, it would measure the percentage of queries for which at least one relevant document was found. So that's just a metric that we can use to.

know if you find relevant documents. You can also use the mean reciprocal rank MRR to know at what average position do we find this right document in our top key. If the hit rate is good, ⁓ you may still not have a good retrieval pipeline because it could be the 10th each time. And so you give nine useless document text chunk to your system for the ⁓ one good document that is only at the at the end.

And so the MRR helps you measure that as well. ⁓ Then you can use the recall ⁓ to know if you retrieve all the relevant document ⁓ that you may have in your database in ⁓ your top 10 documents. That's ⁓ pretty useful as well. You may have the precision ⁓ to know ⁓ if all the retrieve documents are relevant or not. So that helps you disregard or remove the ⁓ uses document that may implement.

Sneha Mehra (00:33:33)  
That may add noise to the system and make the responses worse. There's also ⁓ NDCG or the normalized discounted cumulative gain, which will take pretty much everything into account to ⁓ measure relevance and the order of documents. So it's it's just like the precision, but also taking into ⁓ into account at which ⁓ order in this top 10 responses ⁓ are the good responses. So for example.

NDCG would be higher. If you have five good documents and you try to retrieve the top 10, if these five documents are in the first five positions, the NDCG would be better than if it would be in the five last positions. So that helps you ⁓ better understand if you ⁓ are maximizing your research. And to do that, we also need a data set for the evaluation, the retrieval part of the evaluation. Here, this data set would be.

⁓ Some associated questions and relevant chunks, ⁓ one or more chunks, but basically ⁓ you could have questions that you generate with an LLM. You have, for example, a big database with lots of text chunk that we created in our second course. And ⁓ you want to find the most relevant chunks. And so how do you do that? Well, you you can ask an LLM to generate questions based on the text chunk that you have.

So it would be fake ⁓ or synthetic questions. And then you can measure if ⁓ by asking the same question, your system retrieves the same chunk. There are many useful libraries to help you do that, ⁓ whether it is from Laman index or Langchain again or tons of others, you can find them. ⁓ You can also ask the ask us ⁓ for something specific in your use case.

In this dataset, you also want the version ⁓ of the chunks and indexes with the associated questions or prompts. ⁓ Because chunks evolve over time, you may change the size of the chunk, you may add data to the dataset, you may remove some, etc. You want versioning to ensure you compare apples with Apples. ⁓ Idle, you also want a golden context data set that we see, which basically would just mean that you have.

Sneha Mehra (00:35:59)  
Real data or as close to real user data as possible. So, for example, pairs of questions and chunks containing the answer where the question would come from a user and the chunk is the ideal context that you want to give to the model. Finally, you also want to evaluate the generation, obviously, which is the most important thing to know if the output we give our user is actually good or not. Therefore,

We want to ensure that the answers are relevant, faithful to the context and sources that we give, and that are factually correct. There are three metrics that you can use to do that, but the most important metrics that you want to have, ⁓ as we said, are the personalized ones. So the ones that will track ⁓ what your users actually want. Here we will share just three basic ⁓ general metrics that you should be using just to ensure the generation part is great.

But it doesn't ensure it's great for your users. It just ensures that it makes sense from a general perspective. The first ⁓ metric is relevance, ⁓ which is just measures if the generated answer properly addresses the question asked. It's just to measure if ⁓ the answer uses the right ⁓ context ⁓ to address the query. So you need a good retrieval for that. Then

There's the faithfulness to context, obviously, which based on the context that you give it and the answer, you evaluate if the answer explanation ⁓ is correlated with the sources that you gave it. And finally, there is the correctness, which can be used to measure if the answer ⁓ is factually correct compared to a reference answer or a ground truth that we may have. Alright, so those are all general metrics that you want to measure, ⁓ which you can also.

Just ask Lama Index and Langchain to do ⁓ very easily with just a quick function. ⁓ But as we said, it's very important to build personalized evaluation matrix instead. And so ⁓ how do you measure the matrix and how do you create matrix to measure what the user wants? Well, the classic met the classical method would be to rely on human judgments and to read and ⁓

Sneha Mehra (00:38:23)  
document everything, every exchange with your user and LLM, just to note down what was wrong, what wasn't, ⁓ which is still necessary. You want to do that here and there, but it will be way too long and expensive to do that at scale. And so instead you can use the modern method that we call, which would be to use an LLM as a judge. And here the LLM would replace your human judgment to score answers according to

Define criteria as we mentioned earlier and measure everything that we talked about for generation. This can be automated on large scale with LLMs, obviously. You also want binary scores, as we said ideally. And all this ⁓ should be automated for the most part. And so here, still in level two, the advantages of using an LLM as judge instead of ⁓ doing the manual evaluations.

Would be first that with a powerful system in place for this automated judge, you would have evaluations aligned with human judgment. It would be excellent for detecting hallucinations or subtle errors compared to unit tests, for example. It will allow for rapid iteration on the RAG pipeline compared to having to review everything manually. ⁓ You don't need to wait for your expert to evaluate and to check and give you feedback.

It's also ideal for frequent and low-cost evaluations compared to asking the expert. But there are still some limitations. They can introduce biases specific to LLMs. For example, ⁓ we know that LLMs are susceptible to the position of the information in the text prompt that we send it. So, depending on the LLM, if you send a long prompt and the information that you really want to ⁓ the LLM to know about is at the third of this prompt.

It may be possible that the LLM skips it ⁓ and instead focuses on the information at the beginning or at the end of the prompt. This is a proven fact that LLM do that. And so the position of the text in the ⁓ LLM is ⁓ in the LLM prompt, in the prompt that you give the LLM, is quite important. And it's a limitation here because this don't really happen ⁓ when you do human manual evaluations. Also, the length of the answers might.

Sneha Mehra (00:40:47)  
Hurt the results as well. ⁓ As we know, the more we give to the model, the more susceptible it is to fail because of the noise that is added in ⁓ the larger context. And there is a final bias where the model ⁓ usually has self-preference, which means that if you use ⁓ a GPT-4 model as a judge, it has been shown that it would generally prefer GPT-4 answers instead of, for example,

Cloud answers. That's mostly because they use the same training schema. And so it sees the generation of the GPT-4 model ⁓ as the same thing they would generate. And they they basically we could say that they think they they they see it as a more sensical answer. ⁓ They are not always reliable. ⁓ For example, sometimes prompts are poorly designed ⁓ and ⁓ this can lead to some

Weird problems or various issues where the LLM isn't as ⁓ confident or ⁓ useful as a human. You want ⁓ scores, but the scores aren't enough. Obviously, you want binary to have more secure ⁓ and ⁓ objective feedback, but it's way better if you have justifications. And here ⁓ a warning you can have justifications asking the LLM, but it's often ⁓ way worse, let's say, than.

The actual human expert, the domain, the human that has the domain expert. But still, you want to ask for the justifications from the LLM to later on use it to understand why it failed or not. But also, ⁓ if you ask the LLM to justify its answer before giving it, it helps it give a better answer by itself providing the context that it needs to give a better answer. And obviously, we also we already mentioned it, but you always want to combine.

This automated feedback with targeted human validation. You want your domain expert to always have a few examples, like 50 or so, that you quickly review to ensure the responses are great. ⁓ Here are some tips for using this ⁓ LLM ⁓ judge effectively. First, you want to use very good prompts with chain of thought reasoning or a reasoning model ⁓ if you want.

Sneha Mehra (00:43:13)  
But you you want really good advanced prompts that would make the model detail intenser before responding. You want to give the judge concrete examples of good and bad cases to have ⁓ some kind of few shot prompt and justifications with them. Also, just to note that you want to alternate ⁓ good, bad, bad, good, good, good, etc. ⁓ because if you always say good, bad, good, bad, good, bad, it may

Be biased towards doing exactly that. One good, then one bad, then one good. Just like in exams where you have, ⁓ for example, lots of questions with multiple choices from one to five, and there's four twos in a row. ⁓ You are not inclined to selecting two for the next question. That's pretty much the same thing with LMs here. ⁓ They will be biased by the structure ⁓ you give it. So just be careful and mindful of that.

you want to ask for a binary score or compare two outputs against each other if it's too difficult to come up with a good binary score, pass or fail. This comparison is called pairwise comparison, and it's pretty good for subjective comparison where you give you ask the model to generate two or three versions, and then your judge would evaluate these two or three versions to just find its favorite one and explain why.

⁓ By the way, you also want to provide an equality ⁓ option in this case, just because it's definitely possible that the two answers are ⁓ completely reasonable and provide value. ⁓ So you obviously want to favor pass or fail judgments with detailed critique rather than a more subjective feedback. Then you want to iterate under prompt until you converge with a high correlation rate with the domain expert.

And just note that for subjective tasks, you may want to prefer pairwise comparison that we just described. And for objective tasks, you obviously want to go with direct scoring ⁓ that can be more reliable. And now, how do we build a good eleven judge? The first step is to create a representative data set with your outputs to evaluate. So you start, you have your inputs, which is your prompt, then ⁓ expected outputs of the system. So the ideal case.

Sneha Mehra (00:45:38)  
Then you want to generate your answers. And finally, you want the human annotations for these answers to measure the failed past conditions and check for the critique. Ideally, you want to include here real user cases, as I mentioned, and ⁓ ask several experts, not just one to annotate, to reduce biases. You also want to aim for case diversity and not just ⁓ known obvious errors. ⁓

The second step is to have a domain expert judge a sample. You can start with, as we said, 30 to 50 well-annotated examples and then annotate them, pass or fail, with critique from the expert or yourself to form the basis of the prompt. ⁓ Based on these 30 to 50 annotations and critiques, you can use them to create the best prompt possible for the judge. After that, you adjust the prompt, trial and error loop.

for the LLM judge to reproduce the judgment of the expert as much as possible. After that, you compare the LLM's decisions with human judgment and repeat the steps four to five until ⁓ convergence on the whole database. And why does it work? Because it allows for outsourcing, business reasoning, or the ⁓ decision-making process of the domain expert that we have to an LLM, which is basically what they are good at.

⁓ If they have the right data, it facilitates error analysis and root cause investigation. It helps standardize evaluation criteria within your team. There's less ambiguity than vague grids with scores of 0 to 10 or 0 to 100 with these binary evaluations. It brings out users' implicit expectations, while the explicit expectations are in unit tests or more easily defined.

And it can also serve as guardrails in production ⁓ once you have a good ⁓ judge in place. You can use it later on, as we will see. ⁓ Finally, there's a third level that I will go very quickly over. It's the A-B testing that we all know about. It's the same for ads or other products or software. You basically want to compare models, compare versions of prompts or whatever ⁓ one at a time.

Sneha Mehra (00:48:02)  
And then ⁓ use them live. ⁓ show both versions to real users. Well, evaluate both versions first, obviously. And if you are not sure which one is the best, you just push both in production to compare what seems best with your users. And this obviously is reserved for a more mature product once it's in production. Okay, so this was a lot. We discussed these three levels, going very quickly over the third one.

⁓ Let's do ⁓ a short recap just because it may be redundant, but it's crucial information to know and to ⁓ to not be afraid of because it's the only way you have to actually know if your system is good or not, and on what task to do next. So it's super important and everyone in the team should know about the evaluations, mostly because they need to tune in.

and especially a Drummond Expert need to be working closely with you as the developer.

So, to do a short recap, our datasets here ⁓ are very important because they allow us to measure performance quantitatively, to compare versions of our code, to identify the edge cases and the errors, to prioritize the next improvements to work on, to automate quality and monitoring, which is something quite nice to automate. And here, a good dataset ⁓ must absolutely include the real or realistic users' questions, as we mentioned, the

The golden data set, ⁓ the expected chunks ID. If you are working with a RAG pipeline, you want to know ⁓ what type of data should be returned in ⁓ for which query. ⁓ You also want, obviously, the expected answers, which would be the ground truth of your ⁓ system. You want to know the versions of each prompt, model, chunk, ⁓ or ⁓ your code in general, just like in regular software development. And the the

Sneha Mehra (00:50:08)  
Better the dataset, ⁓ the faster your system can improve. And optionally, but ideally, you want justifications for these expected answers, either from your LLM judge or from your human experts. The better this data set is, the better your system can improve. So this is very crucial. Here's a short strategy to create ⁓ your dataset.

The initial phase would be to generate synthetic questions from the chunks that you may already have with an LLM. This is to quickly launch your evaluation and already start evaluating. Then you want to gradually replace those generated questions with real user queries, ideally 50 or so, and associated chunks that you know will answer this question. Then you will manually annotate the correct answers and sources, which would be more reliable and allow for.

More representative data. So this is just to ensure that your golden data set from your real users also have a good ground truth with good explanations of why the answer was good, etc. The dataset must be improved over time, obviously, with feedback loop and as the product evolves. So that's why we have an initial phase and an advanced phase. And you don't have to wait to have real users to start monitoring, tracking, and evaluating.

You start right away with an suboptimal ⁓ evaluation pipeline that you gradually improve along with your product, which will, even if fully synthetically generated, help you understand your current system and its weaknesses. We know that's it's long and boring, but building your evaluation data set ⁓ is the most important part of the pipeline else.

You will just work in circle and work on many improvements without knowing what works or not. We also talked about a critique a lot and giving feedback. So ⁓ this is just a quick example that you can ⁓ make ⁓ press pause and read if you want. But basically, ⁓ a critique is just an explanation of why was it a success, a pass, and ideally ⁓ even contain how to do better.

Sneha Mehra (00:52:31)  
So the critique would be just the perfect reply from your expert that would measure how good was the AI response and how could it be even better. Likewise, when it's a fail, when it's not a good response, ⁓ the critique would say why and I identify the key missing elements ⁓ to become a good response instead. And we will discuss this further in session five.

But I just want to know that it's super important to have a good UI to monitor what happens over time. Even for these critiques, failed pass tests, you want your domain experts to look at these data and look at how your model behaves. Just because the domain expert, which is not necessarily a programmer, will not be in the loop very often, usually. And so there can ⁓ be a drift between.

What the expert says and would have said versus what the developer or yourself ⁓ work, how how you work on the prompt and edit the prompt and make changes to the whole system. And so this can create a drift between what the expert expects and ⁓ would answer versus what the how the model actually behaves. And this is just to say that you want this expert to look at your ⁓ system and your system's answers.

As often as possible, which means that you have to build a nice UI for him or her to have a look and to ⁓ enjoy looking at it. So for example, you don't want to send the expert just your JSON file with lots of examples. You want to show it ⁓ at least in a clear Excel sheet or something that is more approachable to ⁓ regular human beings. ⁓ That's very important to keep in mind.

Because your expert will be the key factor to make your application succeed with proper evaluations. And just to motivate you to build this LLM judge and this better evaluation pipeline, ⁓ it's not just a one-time use. ⁓ The judge will be super useful ⁓ throughout your whole lifecycle just to measure the improvements and the system as a whole, but also you can even use it ⁓ in the future to.

Sneha Mehra (00:54:56)  
Train or fine-tune your model if you have to do that. ⁓ That's through something called RLAIF, which is ⁓ basically the same thing as RLHF that we ⁓ discussed in the past, but uses an artificial intelligence judge, so an LLM, to help during the fine-tuning to give proper feedback and review rather than a human expert. And if you have built your judge properly, well, you can use it to.

Replace ⁓ your human judge. So that would be quite beneficial to have an already working judge ⁓ in this case. And can be used in many things, ⁓ as we will see in course five, to, for example, ⁓ act as guardrail within your live application, just to ensure your system actually answers properly to the user and to the proper requests, etc. Okay, so that was a lot.

It was pretty much all the theory that we wanted to share about evaluation as a whole. I will now leave Omar to share a bit more of a practical view ⁓ of some ways to build evaluations with our notebooks to you. ⁓ Hello everyone. So in this second part, ⁓ I will be showing some ⁓ code implementations of what we just talked about in the first part of the session.

So I will showcase ⁓ and notebooks around how to evaluate a very simple chatbot, rag chatbot. Then I will try to show you how it looks like when we try to do a more comprehensive iterative optimization process. So a more like a ⁓ like a scientific approach to optimizing a ⁓ a system like this. And then

⁓ finish this session on ⁓ how to evaluate not the retrieval part but also the final answer part of is the answer good enough or not. So yeah let's go let's start. So here I have the notebook ⁓ that I will show you. So if you remember in the last session, I showed how to create very basic, very simple ⁓ rag chatbots using

Sneha Mehra (00:57:21)  
LAMAINDEX THE FRAMEWORK. ⁓ And it's actually quite easy to do it with Lama Index. We you just give it ⁓ data, some documents, and then with a few lines of code, you can have a vector store index. So here I'm not going to go into how to repeat how to create one, but here I'm just actually creating a ⁓ a ⁓ a chatbot again. ⁓ The only thing that is new here is that

We are adding a function that will set the IDs in a deterministic form, the IDs of each chunk in the database. Because if you don't do that, then each time you create a new database, a new vector database, then Lama index just randomizes the ID. ⁓ And in this case, since we want to evaluate the database with the same questions with the same data, basically.

If we restart the notebook and we create the database again, ⁓ we want to keep the same IDs as we had before. So we don't, yeah, we don't get ⁓ mismatch ⁓ on the IDs basically. Here, this is just a different way to create a vector store index. Here we use the ingestion pipeline. We just ⁓

specify what model to use, which vector store to use, and then the documents to use.

Then here I specify the model I want to use. So Gemini 1.5 flash. ⁓ And ⁓ then ⁓ I can ⁓ create a query engine using this LLM and the index. And I want the chatbot to use five chunks to enter me. So it's going to retrieve the highest five ⁓ nodes, chunks, that the system can retrieve.

Sneha Mehra (00:59:21)  
And if I ask a question, the chatbot is able to answer me correctly using the information within the ⁓ vector store database.

Now, for ⁓ evaluating the system, I want to make sure the ⁓ system is actually able to retrieve information correctly, the correct chunks, the correct nodes. So here, since we don't have a dataset, an evaluation dataset, we are just going to use DLLM to come up with questions to use. ⁓ And we are basically going to show every chunk in the database to the model.

⁓ And we ⁓ are asking the model to come up with a question ⁓ for that chunk specifically. So ⁓ at the end, we will have questions ⁓ directly related or associated with ⁓ a specific chunk. So when we ask ⁓ one of the questions that we have in the dataset, we actually expect to see that chunk.

That ⁓ the question was created from, ⁓ we expect to see that chunk in the sources retrieved by the chatbot, the system. ⁓ So this is what we are doing here. This is the prompt that we are using to create the questions. So here we put the chunk text and we ask the model to be a professor ⁓ to come up with.

A specific number of questions. So for example, one question. ⁓ And ⁓ if we decide to have two questions per chunk, then these questions should be diverse and ⁓ use all the document. ⁓ And ⁓ also the questions, ⁓ the answers should be all within the chunk. They shouldn't use a different context.

Sneha Mehra (01:01:29)  
So, this is the prompt to create the data set, the questions. ⁓ So, this was to show you what was the prompt. Here we are actually using the prompt. So, we have this function here: generate question context pairs. ⁓ And ⁓ yeah, here we are just looping ⁓ through ⁓ our chunk database for all the chunks in our database. And here, for example, I'm

Defining the model to use that will come up with the questions. So here it's Gemini 1.5. And here I'm I'm actually not going to use all of the nodes in my database because there are too many of them. I'm just going to use 25\. And I want ⁓ one question ⁓ for all of these 25 chunks. ⁓ And so that's how I am creating my retrieval.

Evaluation data set. This is just for the retrieval part. I want to see if my system is actually able to retrieve the chunk, the specific node that I that can answer my question or the generated questions. And so ⁓ once I have the data set, then I can ⁓ specify the metrics. ⁓ Here we are going to use two metrics. ⁓ One is the hit rate. So

The hit rate measures ⁓ if the actual chunk, the correct chunk, is in the list of retrieved sources, retrieved chunks. ⁓ If the correct ⁓ chunk is retrieved, then it's ⁓ it's good. For that question, we have a good result. And we are going to do this across the data set and then compute the hit rate to know for how many questions we retrieved.

the correct document.

Sneha Mehra (01:03:28)  
Then we have the MRR, which ⁓ measures or tries to give us the score or the average position of the correct chunk in the list of sources. So here I show that if the correct chunk is in the basically in the third position, then I'm going to have a score of 1 over 3\. And of course, the goal is to get to the first position because as we mentioned.

In the first session, the LLMs like to read, like to understand the first part of the context and maybe the last part, but not the middle part. So we want to make sure that the correct information, the relevant context is in the beginning of the context, the beginning of the prompt or at the end basically. So either at the beginning or or the end. And here we're just measuring the ⁓ the MRR to know how the system is.

Working. So here this function will help me ⁓ display the results. And here I am evaluating. So here I want to evaluate, ⁓ and at the same time, I want to change some configuration ⁓ values. So here the only thing I'm changing is the number of chunks I retrieve. So we can see that for ⁓ two chunks that I retrieve.

This is the score. So we have a hit rate of 72%. So for 72% of the questions, I have the correct information, the correct chunk retrieved. And the MRR is also at is closed at 72\. Then we can see that for the last value of 10, we can see that the hit rate is at 96%. ⁓

So this is almost 100% correct. And it makes sense ⁓ that if you retrieve many more chunks, that the probability of retrieving the correct chunk is actually higher. It makes sense. But the goal is not to retrieve too much content because then you're trying you're giving the model actually incorrect information. ⁓ So you're increasing the context.

Sneha Mehra (01:05:54)  
Length, you're increasing the context size with maybe irrelevant information. ⁓ But the the actual answer is in that context. But as we know, the the models need optimized context. Like if you give it ⁓ a million tokens of context, then of course the answer is in the context, but is the model going to be able to?

answer and to find the inf the correct information in that context. So ⁓ this is why we don't want to put too many tokens in the in the prompt because then the model is full with ir irrelevant information and that's not something good either.

So here I'm just showing that you can also use two other metrics. So you have the faithfulness and relevancy metric. So here we are actually using a different model. So in this case, we're going to use either GPT-4.0 or GPT-4.0 mini to tell ⁓ if the answer is faithful or relevant to the question. So if if it's faithful,

Then it means that the answer is not hallucinated, that the answer actually uses the context. And if the answer is relevant, it means that it actually answers the user question, the the answer. So this is what we we measure here using these ⁓ the the the models. ⁓ And if I ⁓ change the number of chunks again.

From two to ten, we can see that actually GPT-4. Mini thinks that it's not GPT-40O Mini, it's ⁓ for the faithful, it's actually GPT-4.0. It it thinks that it's actually quite faithful and relevant. So we don't need to worry about the answer. We just need to worry about the retrieval part, which usually usually is the most ⁓ hard to optimize and ⁓ the most important part actually.

Sneha Mehra (01:08:03)  
The part where we need to retrieve the correct information ⁓ is actually harder than if the model is actually able to use the prompt, the information in the prompt to answer the question. So that's why we see high scores here, because the answers are basically all correct or faithful and correct. ⁓ So I'm going to skip that and just explain you how a more comprehensive process looks like, more like a

Scientific approach, let's say. ⁓ So here I'm pulling up a lesson from ⁓ one of our courses, which you can check out. But I thought it would be way more easier to ⁓ show you how this process looks like this way, because it's actually ⁓ yeah, it's it's not that it's complex, it's just that it's there, there's many important parts I wanted to show. So here for context, we have

The AI tutor chatbot, which ⁓ should be able to ⁓ retrieve information to answer student questions. And I want the chatbot to retrieve information from all of these different libraries, like transform hugging face transformers, PEFT, TRL, even Langchain, Lemma Index. I want to have in my vector database all of those documents.

And I want the chatbot, the system to retrieve those documents to answer questions to the students. And so here I just want to show you how optimizing this chatbot can look like. ⁓ It's a good way to show you how it's done ⁓ and how it can be expanded and optimized even further. So here, like I showed in the notebook.

We want to focus on the retrieval part because ⁓ the retrieval is basically the most important part. ⁓ And especially when it comes to being able to answer specific types of questions. ⁓ And so that's why we here we are focusing on the retrieval part. So ⁓ and to measure that performance, we are using once again the hit rate and the MRR. Now, just to give you more context here.

Sneha Mehra (01:10:30)  
The dataset for this AI tutor is right here. This is all the documents that the vector database contains. So we have almost 400 markdown files. These are complete markdown files for the Transformers library. ⁓ And you can see all the ⁓ numbers being really different because there's less documentation for.

Other libraries, like for example, TRL. You only have 34 markdown files, which is very small compared to other libraries. ⁓ And we also added this is the basically the most content here, 5,000 blog posts from the Towards AI website, which all talk about different topics or aspects of AI.

Artificial intelligence. So in total, we have almost 8,000 markdown files or documents. And we want to evaluate this retrieval system. So to do so, we need a dataset. Like I said before, we need a dataset to evaluate this retrieval part. And we are going to use the same ⁓ approach.

We saw in the previous example, we are going to use an LLM to come up with questions from each of those documents. ⁓ So at the end ⁓ of this process of this step, we now have almost 8,000 question and document pairs. So this is our evaluation dataset. And just a little note here: ⁓ when we do this.

When we actually generate questions like this, we need to remember actually that this ⁓ is not really comprehensive because we don't know if the end users of the system of the chatbot will actually ask the same type of questions. So ⁓ by generating questions like this, we assume that our data set or evaluation is comprehensive, but not really, because

Sneha Mehra (01:12:51)  
These are not questions from real users. We we would need to actually collect real questions to actually make ⁓ create a real evaluation in the real world. But since we don't have users or we don't have ⁓ launched the product or the service, then this is a good way to start to use an LLM to come up with all of these different questions. And we can actually go ⁓ quite far with this. We can tell the model to act.

in different ways. You can we can tell the model to act like a beginner, like a person that doesn't know how to code, ⁓ or we can tell the model to ask, to come up with a question that that someone that is very good, an expert, could ask. So this is this could be something really that go that goes far, but in the end, we need to actually also in the end implement ⁓ real user questions.

And now, since ⁓ we have 8,000 question document pairs, ⁓ if we want to ⁓ go fast in evaluating, then ⁓ we could actually just you know take ⁓ a subset of those questions like here ⁓ and use these to evaluate the system. So here I'm just going to use a hundred questions for the ⁓ Towards the eye blog.

documents and 100 for length chain and or a hundred for lemma index and so on. ⁓ And this way I can make faster evaluations because the the experiments are going to end up and faster. ⁓ And also, like if you notice at the end when you have users that they mostly ask questions about length chain, for example, then ⁓ you could, for example, here put more.

questions for lang chain to make sure that you are actually evaluating Langchain a little bit more because that's what users like. So here we are putting pretty much everything even at a hundred questions and yeah that's that's it. And we cannot go more than three four for TRL because we don't have more than ⁓ 34 documents for TRL.

Sneha Mehra (01:15:12)  
⁓ And what we say here is that we can run ⁓ these ⁓ experiments three or four times in a row, just because each time it's going to take different questions. So if I run this evaluation ⁓ once, then it's going to grab a hundred questions randomly. But if I run the evaluation again, then the a hundred questions will be different. It's not going to be the same.

The same questions ask again. So it's here, it's useful to run the evaluation multiple times in a row just to get a better sense of the performance. We can either do this or just increase the number of questions. Either run multiple times in a row or just increase the number of questions that we use for the evaluation. Now for the baseline, here I'm just choosing some default values for the chatbot. So ⁓ I'm choosing to use 100\.

Tokens long ⁓ chunks ⁓ and with zero overlap, ⁓ I'm choosing the coheres embedding 3.0 model. We are using the simple cosine similarity function for the score to ⁓ to compute the scores. ⁓ And I'm starting with five chunks ⁓ to answer my question.

And so if I run the evaluation on this configuration, this baseline, I'm getting 45% ⁓ of ⁓ hit rate ⁓ and 37 for the MRR. Now I also put the scores for each individual source. So this can be useful just to see what source ⁓ is doing better than others. So for example, we can see that for transformer questions.

It's doing way better than, for example, lengtchain questions somehow. ⁓ And if we try to interpret the results or try to come up with reasons why this is happening, ⁓ you can come up with so just just to ⁓ clarify here, we have an MRR of 37, which means ⁓ that the proper document appears on average at the dot.

Sneha Mehra (01:17:36)  
2.7th position. So this is just to clarify that. ⁓ And yeah, for the hypothesis on why some sources are doing better than others, we can think that for newer libraries like Lama Index and Langchain, they didn't exist when these embedding models were trained, or the ⁓ training data for those embedding models didn't include.

Specifically these libraries. So we think that since ⁓ Langchain and Lama Lama Index are newer, that this could be a reason why they are less performant than, for example, Transformers, which is a library that is ⁓ did exist for a couple of years before these newer ones. ⁓ And also ⁓ we think that, for example, the content overlap could confuse the embedding model.

So, since the Transformers library is large and has many different topics, the ⁓ retrieval can be easier than, for example, a library where the documents are too ⁓ similar within compared to each other. So, if you have, for example, lang chain that explains what are what chains are in multiple files, in multiple ⁓ markdown files, then the model

The retrieval system is not going to be able to retrieve the correct document. Just the simple cosine similarity function will not be able to correctly retrieve the correct document. Now, if I try to optimize this system, the first thing we can do, of course, is to increase the number of chunks in ⁓ the retrieval. So we can see here that if we increase the number of retrieval chunks.

To 15 instead of five, then we get really high scores compared to before. So now we have an improvement of 23% for the hit rate ⁓ and almost 10% for the for the MRR. So this is good. It means that we should keep the number of chunks really high. ⁓ And if we need to reduce, actually, if if this is too much, this is too many tokens.

Sneha Mehra (01:20:01)  
Then we could introduce a re-ranker later on ⁓ to actually re-rank and then cut down the number of sources so that the context is not too long. So that's ⁓ what we say here. Now, if we compare the Coheres embedding model to the OpenAI's text embedding 3 large model, we can actually compute the results here. ⁓ And for this experiment,

We actually see a decrease in performance. We see almost 20% less accuracy for the hit rate ⁓ and 20% decrease for the MRR. So this shows that for this specific data set, which is libraries in the domain of artificial intelligence, the cohere embedding English model is actually more aligned with our content. Maybe because it's simply newer.

Yeah, not sure about this, but this experiment showed that we should use or stick to Coherce embedding model. ⁓ And this is where we don't go completely to the end of the optimization process because here we only compare to one other model and we keep the one we are using since it's better. But ⁓ if we needed to actually optimize this ⁓ further, then we could try.

Four or five other embedding models, see what's better. We can even try to fine-tune an embedding model if the content we have is too new ⁓ or the technical content is too unique ⁓ and it's not really well represented on the internet, for example, then we can consider also fine-tuning this embedding model. So this is something ⁓ that can be done for like more ⁓ advanced optimizations.

Now, if we implement the re-ranker that we talked about, we can see that ⁓ if we retrieve 15 documents ⁓ and we cut down to five to preserve a short context, a very short ⁓ number of ⁓ low number of tokens in the prompt, we can see that it actually ⁓ decreases a little bit the performance on the hit rate for the re-ranker.

Sneha Mehra (01:22:28)  
But it actually increases almost 10% the MRR. So the position, the position of the correct document is actually higher in the list compared to not using a rebanker. So that's what we actually want. So even if the hit rate is a little bit lower, it could be due to random factors, or maybe the type, the question we use here were not. ⁓ maybe if we had used all of the questions.

We wouldn't have seen such a ⁓ decrease. But we think it's ⁓ it's a good trade-off to make a little bit of less hit rate for way more ⁓ MRR performance. So we are willing to introduce here the re-ranker. ⁓ And I just want to mention that we can also fine-tune the re ranker if we don't want, and this is using the Coheres fine-tuning ⁓ platform.

So this would be actually easy to do because you don't need to worry about the hardware, how to rent, ⁓ or here, here, the only thing you you would need to worry about is the build, building the dataset, the training dataset, and making sure it's in the correct format to send it on the platform. So here ⁓ this is something that could be really ⁓ interesting to explore, fine-tuning the re ranker.

Now, if I try to optimize the chunk size and the chunk overlap, here I choose some values, some configuration values, and I measure the performance of the system with all of those options. I actually see that the highest performance is with the baseline configuration. So 800 chunk size and zero overlap. So

This confirms to me that we should keep the baseline configuration. But of course, these are just some of the configurations we can try. We can try to go with a higher number and actually test every single possible combination. But just to make things not too complex and not too long, ⁓ here we just test with a few of them and we keep the highest performance with the baseline configuration. Then another

Sneha Mehra (01:24:53)  
Thing we can introduce ⁓ and that can actually improve the results by a lot in some cases is the hybrid search ⁓ feature ⁓ of some of these libraries, some of these vector database bases. You can ⁓ also implement that yourself, but many of these frameworks databases all already include this functionality in their features. So here, if I use the hybrid search

Feature from Lama Index, I actually get some a little bit of an increase. So 2% for the hit rate and 2% for the MRR, which is very ⁓ good. We can also notice that for the actually for the OpenAI cookbooks, which are mostly Jupyter notebooks converted to ⁓ Markdown, we see a lot of ⁓ gain in performance here. This could mean that.

For this specific library or sources ⁓ of information, searching by keywords ⁓ is better. Maybe there are more questions with keywords for this source in particular. ⁓ So this is good to know. Let's say that for each ⁓ chunk, we can actually filter the source before we do the retrieval.

So let's say that in the ⁓ user interface, someone already knows that they want to ask a question about Langchain. They don't want to search all of those other libraries, they want to search ⁓ lang chain. We can do this, ⁓ it's something that can be done. So here we simulate someone wanting to know something about lang chain, ⁓ and the system here.

Will only retrieve documents, ⁓ will only search within those lang chain documents. And so we see that the performance is actually really high because the system doesn't have to worry or get confused ⁓ by these other sources. We go from 27% ⁓ to 46% and the MRR score from dot 23 to dot 38\. So this is normal.

Sneha Mehra (01:27:17)  
This makes sense. We are only searching within Langchain documents. So ⁓ it makes sense that the performance is better. Now, if we try something more a little bit more advanced, so let's say we do the contextual retrieval. This is a technique that was popularized by Anthropic. They released a blog post a few months ago.

And September, ⁓ and if you actually just implement what they did, we can see a performance increase. So ⁓ the main thing that we do here is that for each chunk in our vector database, we add a summary of the whole document in the chunk. So that means that the chunk

We'll have a better context, will be more ⁓ basically better information in the chunk because the the chunk is well contextualized within the bigger document. So ⁓ we can see that it pretty much ⁓ increases the performance overall. ⁓ We see two percent, three percent for the MRR, which is good. ⁓ Now

If we go to ⁓ question decomposition, we actually see a decrease in performance. So, what question decomposition is, is that when the system receives a question, it tries to decompose the question in simpler questions and then run the retrieval on those two ⁓ simpler questions. But we see here that it doesn't work ⁓ and in

⁓ And you might wonder why it doesn't work. Well, because we think that since the questions that were generated by the LLM at the beginning were already pretty simple, well, it means that if you take something simple and try to decompose it, you get very generic questions, sub-questions. And if you run those generic sub-questions, then of course the retrieval will be less accurate.

Sneha Mehra (01:29:39)  
So we think that actually, since the questions were already simple, this adding this ⁓ decomposition step actually hurt the evaluation or or the system actually. So ⁓ one way to fix this would be to have an LLM detect if the question is complex. If it's complex ⁓ or if the user wants to know multiple things, then the model

Could then trigger the decomposition step. But yeah, here we didn't have that. We simply decomposed directly after ⁓ receiving the user input. So this is good to know. Now, this was all for the scientific approach ⁓ to optimizing a rag chatbot. I I guess the main, the most important part is just to measure the performance.

And since we are measuring the retrieval performance, then it makes sense to measure the hit rate for this specific task. But yeah, there are many other metrics, but for this specific ⁓ system, it it made sense to use hit rate and MRR. But but yeah, the the most important part is to just find what are the metrics that are ⁓ important for the system and then optimize against ⁓ those metrics.

So if I chose recall instead of ⁓ headrate, then I would be optimizing for recall instead of head rate. ⁓ And ⁓ yeah, measuring ⁓ is better than just ⁓ reading the answers. ⁓ And if you don't have any users, then it makes it makes a lot of sense to generate ⁓ all of the questions. And then once the product is deployed,

Then you can start using real-world user questions for your system evaluation. But yeah, this is pretty much ⁓ what it looks like. ⁓ after all of those improvements, we see that we increase the hit rate by almost 30% and ⁓ the MRR by 40%, which is pretty good.

Sneha Mehra (01:32:04)  
We could go further, we could try better embedding models, we can try fine-tuning an embedding model, we can try fine-tuning the re-ranker, we can we can keep adding more steps. They will increase latency, but the system can be more accurate. So yeah, this was ⁓ just to show you how this the process can look like. ⁓ Now, for the last notebook I want to show.

This one is about how to evaluate the answers of an LLM. So let's say you are working on a feature that generates marketing content ⁓ and you want to make sure that the marketing content is actually good. You can actually do this by using what we call the LLM as judge technique, where you where you use another model to evaluate the answers of

your system. So this is what I will show you ⁓ how we can we we can create a very simple LLMS judge. But first let's start by creating once again a llama index rag chatbot. So here I quickly create one, I ask a question, and then the the chatbot is able to answer me. Now I can then create a dataset

So once again, I ⁓ I I I prompt a model to give me a dataset of questions ⁓ associated to one document, and I get this synthetic rag evaluation dataset.

Then with those questions, of course, we want to read them and curate them. So here I'm just simply displaying all of the questions that were generated, and I can check to choose the ones I want to keep, for example. So, and then I can clean ⁓ that way. So if that if some questions are too weird or don't make any sense, then I can remove them.

Sneha Mehra (01:34:18)  
So, right now we don't have the answers ⁓ of those questions. We only have the questions and the documents. So here we are going to complement the dataset by gathering the actual answers. So here I reuse the dataset above and I generate 27 answers for those 27 questions. Now, what you want to do once you have

This data set is to have a human expert ⁓ rate those questions. So that person, it can be either you or someone that has a lot more knowledge about this specific task. So, like I said, if you are ⁓ prompting a model to write marketing content, here it would be beneficial to ask an actual marketer to review the questions and the answers.

For those questions ⁓ and have that person, that expert, review the dataset. So ⁓ now let's say we we did this. ⁓ Now I'm loading the labels. So I have my data set that we created above, and we have the labels and the critiques ⁓ of the expert. So we have ⁓ the questions, we have

The document that contains the information to answer. And we have the actual answer that the LLM ⁓ generated. Then we have someone else or you label the dataset, which then the the the answers can be either good or bad. ⁓ And then we have a critique that about why it's good or why it's bad. The most important ones are the bad.

Because the expert can then guide ⁓ the optimization about why this answer was bad. So if we count the number of questions that were answered correctly, we have 18 of them, but nine were quite ⁓ bad according to the expert. ⁓ Now, if we create an LLM as judge here, so simply asking another LLM to do the same thing, to label the answer as bad.

Sneha Mehra (01:36:44)  
or good. Here we can ask GPT-4.0. ⁓ We can ask it to label good, bad and give a critique.

Sneha Mehra (01:36:58)  
We can see that these ones were given by GPT-4.0 instead of the human expert. And we have ⁓ these results. So right now, ⁓ the LLM as judge is not aligned with the expert. So this is not good. We cannot use this as our LLM as judge. We need to optimize the prompt for GPT-4.0 to make it so that the answer of GPT-4.0. ⁓

matches as much as possible the answers of the human expert, the labels.

Sneha Mehra (01:37:38)  
So here we measure the the agreement. So we can use the ⁓ sum metrics that measure the agreements. So Cohen's kappa is one of them. We measure 0.6 ⁓ and then we have the simpler accuracy metric at 63\.

⁓ And ⁓ if we actually read the critiques ⁓ and we improve on the GPT-4.0 prompt, then we can actually improve ⁓ the agreement. So this is here ⁓ a newer version of the GPT-4.0 prompt with clearer requirements that actually match ⁓ the ⁓ the human expert labels.

So after ⁓ optimizing the prompt for the LLM as judge, we see that the coherence kappa, which measures agreement, actually went up. So now we are 0.25, which is really better. And yeah, that's it. That's it. Then we can use, we can keep improving this so that the LLM as judge labels just like our human help expert.

And then we can use this LLM as Judge in our evaluation process and make sure that our marketing content is actually ⁓ good. So LLM as Judge is a really good way to scale the evaluation to like a thousand generations to make sure that ⁓ every possible ⁓ user input ⁓ is ⁓ well answered for any type of application. So

You can take this and apply it to any type ⁓ of ⁓ task basically. And here below, I just I'm just showing the ⁓ llama index ⁓ enter relevancy and faithfulness metrics that we saw before. So yeah, you you can see that faithfulness is quite high, relevant, not so. You can use these as well.

Sneha Mehra (01:39:51)  
These ⁓ metrics which are just prompts to another model. Bas basically this is just ⁓ another form of LLM as job.

—--------------------------  
2

Sneha Mehra (00:00:09)  
Welcome to the second session of this course. In this one, we dive into building on top of LLMs. If you missed the first one, we covered the foundational knowledge and using LLMs. So it was a bit more of a theory recap, and we discussed the limitations and weaknesses, and most importantly, why does this course exist? We share a lot of tips, useful tips for using closed versus open models, prompting, benchmarking models, etc. In this one, we will focus on.

Customizing and building around LLMs. And more specifically, here's the plan. We start with still a bit of essential theory, let's say. So first, we just do a small comeback on LLM's limitations. So the problem with context window, the knowledge problem, and more. Let me dive into that shortly. Then we talk a bit more about embeddings and encoders, what we can do with long context, rag, cag, fine-tuning, ⁓ and finally reinforcement learning. Then

Omar comes in with a bit of practice with structured outputs and practical demonstrations implementing rag, cag, and fine-tuning. All the cool stuff. ⁓ So let's start right away. ⁓ LLM limitations. There are many of them. I guess the biggest one is hallucinations when the model just ⁓ invents stuff on the fly because it's just a predicting machine that's predicting words, doesn't know about truth.

Or lies, as we saw in the first session. And so there's a problem we call hallucinations, maybe not the best word for it, but it just means that LLM is generating nonsense. A second limitation is the context window of models. While some models like Gemini give ⁓ millions of context window, it's still not that performant. And as we saw recently, if you heard of it, Lamafor had ⁓ pretended to have 10 million.

⁓ tokens in the context window ⁓ and ⁓ it wasn't performing so well. So anyways, we can scale to millions of context windows, but the models aren't performing that well at this range. And otherwise, ⁓ there's a true limitation of skills and potential when you give more and more tokens to the LLM. So you always want to restrict it and give it the most essential stuff.

Sneha Mehra (00:02:37)  
So we will cover this in the course. Another limitation is that the knowledge you need is sometimes not present in the training data set. If you want to ask questions in a different language or about a programming language that is proprietary to your company, might not be in the LLM ⁓ knowledge base when it was trained. And so ⁓ it won't be able to answer correctly. ⁓ If you are not sure why, I invite you to check out the first course again, where we discussed the training of LLMs.

Another limitation is the no-ledge cutoff. So this basically means that when we train LLM, as we saw in the first course, the training ends at a time t. ⁓ And after that time you can serve the model to users. But then the model isn't updated anymore. So if the precedent changes or anything happens, it won't know about it unless you do some kind of fix, which is, for example, to give it access to internet and other things, which we will see in this course.

But this is a limitation, an intrinsic limitation to LLMs, that its knowledge is fixed in time. Another limitation is the reasoning of LLMs. I'm not talking about reasoning models like O1 or O3 or O4. I'm talking about models not being ⁓ powerful ⁓ logical ⁓ machines. Like they're not conscious and humans, they make very dumb mistakes, as we saw. And so a very

Fundational limitation is their capacity to reason and understand the world. ⁓ So solutions. Well, first we add knowledge. That's the easiest fix, I guess. And that's that can be done through ⁓ many ways. ⁓ And these many ways we will cover them ⁓ in this course and in what order we have to do them. Another solution is to control or steer the generation.

And this can be done through many approaches as well, that we'll cover ⁓ in the next hour. So, what can we do to fix ⁓ LLM limitations? Should we fine-tune ⁓ or retrain-tune, ⁓ do fine-tuning? There's many synonyms, but basically, yeah, just retrain our model. Should we do that from the start? ⁓ No, ⁓ never first. ⁓ why we never want to do that first? Firstly, it's because it's

Sneha Mehra (00:05:04)  
Quite costly. You need thousands, if not millions, of examples of questions and answers you want, because you are basically retraining your model to understand the language and the world in different ways. So you are not just trying to tweak a bit how it responds, you are rather trying to ⁓ reshape its whole understanding of the world specifically to your needs. So it takes a lot of data, it takes a lot of time with experimentation.

To do that correctly, ⁓ which means it's also super complex. You need to do that many times. There are training problems like failure runs where we need to restart or start ⁓ from an RDR checkpoint. It's very complicated. ⁓ Plus, it means you need ⁓ expertise, you need some ML engineers, ML research scientists, you need to spend a lot of money and get the right people on board to help you do that.

Another ⁓ limitation of doing that from the start is that you become model dependent. And with such rapid progress that we see with all the state-of-the-art models, this is quite a big limitation because you stay fixed in time with, for example, if you fine-tuned Lama 2, you still have Lama 2, whereas Lama 4 recently came out and some other ⁓ very great model, even open source models, that are much better than Lama 2\. So you need to retrain for every new model.

It's even more expensive and ⁓ complicated to do. Also, another thing is that the model might not be the problem. So if you do that from the start, you ⁓ guarantee costs ⁓ and wasting time, and you don't guarantee the results. The model itself may already be capable of what you want it to achieve, but with a little help. Here's ⁓ what we can do instead of fine-tuning right from the start.

So I I will ⁓ draft this whole first hour around this plot. So I will just take some time to explain it. But basically, there are two axes, four ⁓ quadrants, to represent the space of all possible solutions and things we can do. ⁓ The two axes are from low to high, left to right, or bottom to top. ⁓ And the y-axis, the one ⁓ vertical, ⁓ it is about external knowledge needed, whereas the

Sneha Mehra (00:07:31)  
A horizontal axis, the x-axis, is about the model adaptation needed. So how much you need to tweak the model itself to achieve what you want. So for example, if we ask what is a transformer, the model already knows that. You don't need to retrain it to modify its weights, its parameters, you don't need to add external knowledge. It knows what it is. This is fixed in time. ⁓ What a transformer is doesn't really change. But if you ask it,

What is the current best LLM or current best transformer? Now this has changed since 2023 or whenever the model was trained. So you need to add external knowledge, whether it's like the top five query from internet or just your own database, you need to give it updated information so that it knows what is the current best LLM and then can ⁓ compute and generate you a good answer.

If you ask it to write it in a language that it doesn't know, then you need external knowledge ⁓ and you need to adapt the model because just giving it some information on the new language won't make it speak the new language. You instead need to really change its inner workings so that it can talk this language. So you need both external knowledge ⁓ and adapt the model. And lastly, ⁓ you might want to have the model write in a specific way.

compliant to a specific norm or follow some kind of structure. In this case, you will not need external knowledge because the model already knows about this, but it doesn't really do it every time. Instead, you want to adapt the model to do it every time, but it already knows how to do it. So you don't have to really fine-tune each parameter of the model and really retrain the way it thinks. You just need to train it a bit more, a bit superficially.

And what are each of these quadrants called? The first one, the basic one, is prompt engineering. The second one is our rag quadrant, so giving more information, retrieval and content generation. ⁓ The third one is fine-tuning, our the thing that most people want to start with. And the last quadrant is reinforcement fine tuning, that we will all dive into in the course. ⁓

Sneha Mehra (00:09:56)  
Just to show a roadmap of what we will do, ⁓ here is a typical path of what each ⁓ company should do when implementing LMs in their application and in which order. So I will go over them very quickly, but just be reassured that I will ⁓ go over them one by one with lots more details in a few slides. So this is just to show what we will do, but we will go over them in detail. So don't worry if it's a bit fast.

So, first you want to start with prompting, specifically zero shot prompting. Then you want to try few shot prompting, so giving it a bit of examples of how to do it. Then implement some basic rag feature, more advanced ones. Try fine-tuning, but not the generator, the encoder. Then try fine-tuning a generator, and finally, fine-tune a generator again, but using reinforcement learning. But this was quite fast, and we didn't really explain what.

Each of these steps mean. So let's rewind back a bit. We want to start with prompting. This means maximizing or leveraging the LLM's base knowledge as much as possible. This is because LLMs are extremely powerful, they know a lot of things, so why not try to extract the most out of it before ⁓ we spend money? So, how do we do that? First, we choose the model according to our needs.

So the right model for the text. For example, if you have very long context and long queries all the time, you better use Gemini for now. If you have complex requests that will require several steps, back and forth, follow-up questions, I'd ask a reasoning model, like O3, for example. And then for more intellectually simple requests, or like very basic questions, I'd try ⁓ 4.1 mini or nano Gemini Flash.

Or 4.0 mini, depending on ⁓ where you're using the models. ⁓ Then you want to optimize your instructions, your prompts. You want to refine them, ⁓ test them on evaluations that you build, which we will see in the next course, and ⁓ ensure you maximize the instructions. Finally, you want to maximize the context you are giving to the model in your instructions. ⁓ That's another part that is done ⁓ doing RAG, which we will cover shortly.

Sneha Mehra (00:12:19)  
But first to optimize the context, you want to give the model a few examples of ⁓ showing what you want to achieve. So for example, here I give it just two examples of how I want to translate text from English to French. And here, when I ask translate this text and then enter the English query, it translates it in French in the right format just by sending it two examples. We recommend sending at least three to five examples minimally.

Then you want to add the missing context from these examples. And to do that, you would have pre-selected examples added to the prompt of the user to enrich the prompt, send the LLM, generate the response, etc. So this is all from the three to five examples that you you give. You add it to the context to add the context. Next, you may want to adapt your examples according to the request of the user.

Instead of sending always the same three to five examples, you may want to build some kind of JSON file or whatever type of mini database that you could host locally easily, and instead have a dozen, ⁓ 20, 30, 50 examples, and find the best examples to give to the user based on what they ask. So you again enrich the prompt, but you select

The three most relevant examples based on all your examples in the question. So here, ⁓ this is what we call the retrieval part. And there are several steps to find the relevant information in this retrieval. For example, in our JSON, we want to find the three most relevant examples from the 50 we have. So how do we do that? Well, we have to understand the user's request. Then this user request, we have to translate it into a language the machine can understand. This is

Done using an encoder and embeddings, which we covered in the last course. But quickly, ⁓ it's basically at this step ⁓ where we send the transformer model. So after ⁓ taking our sentence, making our tokens, and then making the embeddings, this is what we use to compare text. Why? Because it's our numbered representation of each word and of each concept, meanings, sentences. ⁓ So

Sneha Mehra (00:14:48)  
It basically allows the model to understand the sentence and the words and represent it in its own space. So here, for example, we have its space showing that cat and kitten is pretty close, whereas Doug is a bit farther away. And so thanks to these numbers and this space, we can, since they are numbers and not just words, compare them to each other. This means that, for example, here, if we we show the cat and kitten points, we see that they are much closer than cat. ⁓

And dog R. This means that if the user enters cat and you select the most relevant embedding that we have in our dataset, it would be kitten or cat if we allow to give the back the same embedding, but otherwise it would be kitten because it's closer to it. So this is basically what we do. We use embeddings or vectors because they are in the same space, the same language the model understands. And so we can

The model can compare them. Then we find ⁓ these most similar examples with approaches with calculations like the cosine similarity, which is basically just if we get back on these arrows ⁓ representing cat and kitten and dog, ⁓ it just measures the difference between them. So the distance between them, which in simple terms will be the angle between the two vectors. So this cosine similarity measures exactly that.

Now, if we get back to our path, we have been implementing basic rag. ⁓ We are starting to implement that. We have a JSON with many examples and we try to retrieve the most relevant ones in them. But basic rag has obviously some limitations. For example, retrieval can become pretty slow with many examples. If you have a JSON with millions of lines and trying to compare all of them one by one to find the two most relevant,

It will be pretty slow. You will also have to constantly maintain and update the JSON, read it with each request, etc. It's super inefficient globally. ⁓ Also, you cannot have other types of data. You cannot have images, tables, and ⁓ sounds or whatever from this basic setup. Embeddings are also not always the best way ⁓ to find what we want. Sometimes we might actually want to find

Sneha Mehra (00:17:14)  
Exact keywords like a specific name or a specific place instead of asking ⁓ the model to give us information about different ⁓ places, different people. That's just because we may want to have everything referring to someone specifically in our dataset or some concept rather than just getting the most similar meanings, which won't guarantee us to have exactly this name ⁓ get back to us.

Every time. ⁓ That's because embeddings understand the global context of the text. So if you embed a sentence, it will understand the global context of the sentence, but not the specific words or exact facts. For example, also, if you want only in your dataset the results of ⁓ men aged of 50 and above, you better use filters rather than trying to do that with an embedding. Or likewise, ⁓ more general questions about the database itself, like

How many clients do you have that like sports, or how many Python programmers do you have, etc.? These could be done with just filters. The user's question doesn't necessarily have the context needed as well to find the useful data. This means that sometimes a user's question is ambiguous and doesn't really contain ⁓ everything you need to know to answer ⁓ this user, which means that you may need more detail.

And you need a more complex setup to ask it the details that you like. For example, if the user asks to give specific things for again men aged of 15 and above, well you might want to ask how many of them, from what location, etc. Everything that makes this query ambiguous, you want to clarify it. And the solution to do that is to build some kind of more advanced rag system or retrieval augmented generation. So we are now here.

In our typical path. And the difference ⁓ between basic and advanced rag is basically that in basic rag, we have on the left a user asking a question, then we retrieve the useful sources, add it back to the prompt, give it to the LM, give a response. That's pretty much all we have. In a more advanced setup, it will rather look something like this. And we will go through all these steps one by one. ⁓ So first, the data.

Sneha Mehra (00:19:40)  
We need to start with good data. Remove the junk with filters, small models, regex, or even ideally, human review. So you check your data before preparing them and make them as good as they can be. After that, you perform the crucial steps: chunking the text, embedding the text, and putting it in your database. So chunking is basically just splitting your text into smaller bits so that it's best to

Give the model only what it needs and not additional context which would add noise and make the model not able to answer or trump the model a bit. And there are many ways to ⁓ chunk the text that you have as optimally as possible. And when I say optimally, I mean ⁓ in optimal sizes. So you don't want to cut where it's in the middle of a sentence or in the middle of an ID.

even if it has many paragraphs, etc. Basically, if it's too small, you will lack context. If it's too large, you will add useless noise. You generally want to keep your chunks between 200 and 500 tokens. ⁓ It's a good compromise depending on the encoder and LLM you are using. And the most important part is that you test and evaluate and measure for all the different scenarios that you may have.

So there's no one ground truth ⁓ solution. It really depends on your data.

You also may want to try chunk overlapping, which means to use an overlap of let's say 10 to 20% ⁓ to avoid information loss between segments. So that means ⁓ one of the chunk, one of the segments, will have plus 10% ⁓ of tokens. So if you have 100 tokens in your segment, you will also take the next 10 tokens, ⁓ and the following segment of 100 tokens will take the previous 10 tokens as well. So they will have an overlap.

Sneha Mehra (00:21:44)  
Of 10 tokens that goes to both segments, just to ensure that we didn't cut through something relevant for the context. Another thing we can do is intelligent segmentation. ⁓ You don't want to cut randomly, you want to rather favor complete sentences, coherent paragraphs, respect the structure of the text. Then you can also use variable context levels, ⁓ which is often called material scar chunking.

And embedding where you can do that at two levels. So here I say chunking and at the level of embedding, ⁓ which means that at the chunking level, you would create nested chunks like sentence, ⁓ embed sentence, embed paragraphs, then embed complete section, and do retrieval on the context levels adapted to the query. And that is how to retrieve context levels adapted to the query, finding the most similar information ⁓ which would ideally find.

The right ⁓ level of information it needs. Then you can do that, this Matrioshka type of compressing with embedding. ⁓ This is a bit more complex, but basically you would just have multiple encoders that will encode at various levels the same chunks. So one chunk would be ⁓ as we said in the first course, would produce an embedding which has thousands of values, for example.

And instead of just having one with a thousand value, you would have embeddings with a thousand value for all your chunks. Then you will have another embedding with, for example, a hundred value for all your chunks, your chunks, then another embedding with 10 values for all your chunks. If you have a huge data set, you can start with comparing only the 10 value chunks, finding the hundred most relevant, and from these hundred,

Use the 100 embedding size or the thousand embedding size to just find the three most relevant chunks. So that's ⁓ specifically to optimize ⁓ the research process. But it's a more advanced ⁓ feature. And while you are preparing your chunk, you also want to do other things in your dataset. You want first to extract the metadata or any information you may have about the chunk, add relevant context to these chunks, relevant information about the global document.

Sneha Mehra (00:24:09)  
Main topics or just synthetically generate section summaries or other relevant information that could help the model. ⁓ And then you may want to do anonymization as well, depending on your type of data. Finally, after that, you want to embed your text chunks into ⁓ embeddings. ⁓ And for that, you use an encoder, ⁓ either open source or private. Cohere OpenAI has great ones. You can use open source ones as well.

After that, there's the research part. And as we said, retrieval with vector search isn't the best search method. Embedding comparison ⁓ isn't the best search method. So you can do different things. The first one could be keyword search ⁓ with techniques like BM25, BM42, etc. You may want to just compare keywords and find the same keywords in your dataset. ⁓ Second one would be using

Filters either by metadata or other type of data that you may have. ⁓ Or then you can try embedding search, ⁓ vector search, which we describe that use two important things: the vector database and indexing. A vector database is just to store the embeddings efficiently, like ChromaDB for example. It will save the text chunk and the embedding together for easy retrieval.

It works differently from classic databases like SQL and NoSQL that use tabular structure. It really uses embedding, puts them in order, etc. It allows for scalability to millions of vectors without brute force comparison or comparing, which means comparing one by one each vector as we described in Basic Right. ⁓ And some popular examples here ⁓ are Pinecone, ChromaDB, Wave 8, etc.

Then an important part is indexing. This simply means to group and optimize embeddings to accelerate research. ⁓ This is always to avoid brute force ⁓ search across the entire database to reduce latency. So instead of having ⁓ for each user query, having to compare it to all your database and then find the two most similar, you will have them in order already and just easily be able to know which one.

Sneha Mehra (00:26:30)  
Is ⁓ more similar super rapidly. If we go back here, we have our embedding search, and the next step would be to try hybrid search. So combining keywords and embeddings. This would be super efficient. You take the three best documents for each and combine the results to have more information and the best one possible. Then you can add other ways to find ⁓ information like an agent connected to your SQL dataset, a CSV. ⁓

A graph database, you can add whatever you want. And lastly, you can do an intelligent routing system. So adding them all together. ⁓ Basically, ⁓ here you will have another LLM, which we call a router, having specific prompts to orient the retrieval to the right ⁓ type of retrieval. So if it understands it needs to find the most

Relevant exact words, it will use keyword search, for example. If it understands it needs very good ⁓ structure to understand links between two entities, etc. It will use SQL or GraphRag. Or if it needs to understand the global context and just reply with a good answer, it might use just cosine similarity with vector search. So it does that intelligently, ideally. So now we are still here with the retrieval. And if we add our user, it looks like this.

User yes ask a question, we do the retrieval from the database that we built. There is nothing new much here other than the whole data preparation steps. So what can we add? Well, we can improve those information a lot. ⁓ So ⁓ when we retrieve information, we can reorganize them, rank according to the quality of the retrieve chunk, reorganize the data to make more sense of them.

Which can also all be done by a model that we can fine-tune to do that. We also may want to remove duplicates, ⁓ or ⁓ depending on what we find, summarize the information if we have too much information in our chunks, or otherwise augment it with a loop to add ⁓ more chunks and find more relevant information by searching differently in the different database or ⁓ the changing the request to find different ⁓ responses.

Sneha Mehra (00:28:52)  
And then once we optimize the chunks that we receive, we can add it back to our prompt, give it to the LLM, and provide the response to the user. But ⁓ it's not over. First, we want to evaluate our pipeline and quantitatively improve at each of these steps. We don't want to just try these steps and implement them one by one and have this thing at the end ⁓ without having tested beforehand. ⁓ When chunking, we want to test.

When adding the re-ranker, when adding the summarization, when adding each of these steps. ⁓ So testing, which we'll see in the next course, is crucial here. We also have limitations with this advanced system. ⁓ There's no LLM to improve the retrieval itself. The user questions can still be vague and problematic here because it will retrieve nonsense.

We do not allow for images and other modalities still. We may have difficulty finding the right chunks. ⁓ We may have difficulty providing a good answer, even if we find the right chunks, like speaking a new language, for example, especially with the mini models that aren't so good at ⁓ generalizing quote to quote. So, next, we might want to add a few things, a few interesting next steps for a real advanced rag.

And this is all, by the way, optional, depending on your needs. Again, you want to evaluate each step to know if you have reached a good point or you want to still keep improving the system. So the first thing we want to do is to incorporate a verification loop between the generator and the retriever. This is just to make sure that we don't fail at the retriever step and we give the best information possible that we have access to in our database to the generator.

So we want to answer that before the generator generates the answer. ⁓ And to do that, we may also want to rewrite the user's query for better search context. So if the query is ambiguous, we may want to either automatically, based on our knowledge, modify the query and add the essential information before doing the search, ⁓ or ⁓ ask the user to refine, clarify, just like OpenAI's deep search does if you've tried it. If you ask it

Sneha Mehra (00:31:16)  
A question to OpenID search, it will, I think, in all cases get back to you with questions to clarify what if it has understood what to do and ask for clarifications. You then may want to try new ways to retrieve and understand context, like Graphrag, for example, ⁓ where you ⁓ you you'd use that to manage global queries about your data set to

Improve the understanding of relationships in your dataset and to extract ⁓ and structure knowledge in different ways. So that's an advanced technique you can use ⁓ if ⁓ you see relationship issues when ⁓ the model generates its answer. So if you see it didn't really understand everything between two ⁓ things, ⁓ either two personas or or whatever two concepts, you may want to use RAG to ensure you retrieve

All the information that links the two concepts together. ⁓ You may want to add modalities, ⁓ which means adding image search, like using visual embeddings with ⁓ relevant images. ⁓ This can be done using a model like clip to embed also images. So have embeddings of our images that can be compared to text embedding, textural embeddings. ⁓

You may want to use audio and video ⁓ analysis, ⁓ either convert speech to text, ⁓ or transcript videos to use them in your text database as well. You can use structured data integration with SQL or CSV and generating charts via LLMs beforehand. You may also want to ⁓ do an adaptation of your embeddings to specialized models like clip, for example, to process or convert specific modalities from image to text or

Audio to text, etc. And we are now at this point in the graph where after advanced rag or during doing advanced rag, you may want to start trying to fine-tune actual models. And here start with the encoder and our re ranker here. And for example, fine-tuning our the encoder means that ⁓ the encoder would be able to have a better understanding of the of the of the domain.

Sneha Mehra (00:33:41)  
So here I just have blue and orange, but to illustrate, I will give a specific example. ⁓ If we are in the medical field, and I'm not an expert at all in the medical field, ⁓ we may be talking about a specific ⁓ illness or a specific cancer or ⁓ whatever with very specific medical terminology. And if I hear them right now, I will just put them all together under it's something medical. So here, that's what we see on the left.

Where the orange and blue dots are bit ⁓ together. But to an experienced ear, like a doctor, they know ⁓ when they hear a word, they don't classify it as just as just mystic stuff. They classify it as, ⁓ this is related to the prostate, this is related to the tongue, whatever. They know what it means. And so they can cluster the topics in much better ways. So here that's what we will try to teach our

Embedding model is to do better embeddings by having him better understand the domain ⁓ and thus create ⁓ create better groups of data that resemble each other instead of just having like a ⁓ global group of ⁓ this is related to medicine. We would have a specialized medicine embedder which in it would have groups for all the specialties that.

a usual specialist doctor knows about. So it's always to replicate the expert and here giving the model expertise to a domain, if we if that's our current issue, would help with ⁓ having the next, the generator and the retriever and everything, a better understanding ⁓ of ⁓ the answers as well. Without having to fine tune the generator itself. Yeah.

Then another thing we can do in this whole process, still in advanced rag, would be to implement CAG or context augmented generation, context caching, which is again ⁓ optional, but ⁓ quickly, ⁓ it's just a Boolean you turn true or false when you use an API, or you can implement it locally. But basically it will just cache the embedding of specific things.

Sneha Mehra (00:36:05)  
So if, ⁓ as we said in the first course, you ⁓ ask a question about a report and you send the whole report to the LLM and then you ask a follow-up question, you don't want to send the whole report again. It will just process the whole words, ⁓ again tokeni, again do the embeddings, all before sending the actual model. Here instead, we will save ⁓ the embeddings just before the attention model, ⁓ attention layer that we see here that we covered in the past course.

And ⁓ so the whole report will be already ready to process by the model. And the only thing we'll have to process the new question ⁓ that we do the tokenization, the embedding, and then we add it ⁓ to the right of the of the report, and ⁓ which is the query on the image on right here, and we get our answer. So it's basically just ⁓ a thing to make it much more, it's basically a technique to make the generation much faster and

Efficient.

Then we may want to finally try fine-tuning. So as you see, it's not from the start, it's actually at slide 115\. So it's quite ⁓ down the path of improving the LLMs ⁓ in your application. And when we say fine-tune our generator ⁓ and our models, which is the model that answers the user, we mean to do that as efficiently as possible. Here in fine-tuning, we often just try to retrain the deeper layers.

This is because it has been studied, and the deeper layers and transformers are the ones that understand the context and ⁓ the sentence and everything more conceptually, whereas the first few layers focus on grammar, syntax, and more language stuff. And so we do this to add knowledge to our LM. And this is all done ideally with a technique we call LoRa that we will see in the next courses, which is about.

Sneha Mehra (00:38:05)  
Fine-tuning LLMs efficiently, where we keep the same big large LLM, the large model, large language model, and instead we just use tiny adapters that would modify all the model without having to retrain all the models. We just train those tiny adapters. Anyways, we will see that in the next courses. But it's just to say that there are ways to fine-tune models ⁓ very efficiently. ⁓ And finally, if this still doesn't really fix

All your issues, ⁓ and there's still a problem with UX and with the way the LLM generates answer to your users. You may want to try reinforcement fine-tuning, ⁓ which is new to the OpenAI platform and other platforms, ⁓ or ⁓ reinforcement learning fine-tuning that is often called. ⁓ And so this is mainly to improve user experience, as we say.

Because here, instead of retraining the model or adapters for the model to generate tokens in different ways, you actually retrain the model to speak in different ways, but with the same knowledge. We covered it in the last course quickly, and we will cover it in the next few courses. But basically, what we do here, instead of retraining to generate tokens one at a time and changing the way it generates tokens.

You actually train ⁓ the model again, but not per token, but rather per answer. So you give it examples of a question, a query, and a response. And then you get ⁓ you give it a system to rate the answer, whether it's an LLM as judge ⁓ or ⁓ a specific calculation that you do, a function that you use, whatever. You give a rating to the answer, and this rating

will be used to tweak the model's ⁓ weights, the model's parameters. So you retrain it just once per response instead of for each token as done in in regular fine-tuning.

Sneha Mehra (00:40:14)  
And so we've covered lots of things, but when do you do which one? Between RAG, CAG, using long context models, fine-tuning, reinforcement fine-tuning, reason using reasoning models, which we didn't even cover here because we will talk about it in the fifth course, which is available on the Academy platform. But quickly, here's a quick recap on when you want to do each. Typically, you want to follow the trajectory that we covered here, but

Quickly, you may want to use prompting for initial testing, or if you see the model already contains the knowledge you need. ⁓ Or if you just need to add a little bit of knowledge so that you get a good answer. You definitely want to just test and try it because it's super low in cost and easy to know if the model will be good enough or not. Then you may want to use RAG for larger database that you have and for faster response time and reduced cost. Because

You can use long context, send the entire code base or whatever report to the model. ⁓ But this will use lots of tokens, increase the cost, and even add a problem because longer contexts aren't perfect and the models ⁓ generate ⁓ random things sometimes because of the needle in the haystack problem that we saw in the last course. So, anyways,

Rag is the most efficient ⁓ system. But if you have, for example, a monthly report and you just want to ask two or three questions, you don't want to build a whole rag pipeline about the the report and repeat that each month. You can use a long context, but when you do that, ⁓ try to use CAG ⁓ turned on for when using ⁓ a Gemini model or whatever, for example, just because it will save you money ⁓ by

Computing, pre-computing the report once, and then you can ask your question several times and it won't ⁓ compute these embeddings every time. ⁓ For fine-tuning, you may want to use it for very specific tasks. ⁓ Also, if you have existing substantial data set, because it takes ⁓ thousands, if not millions, of examples to fine-tune, retrain a model the way you want. And that's really to optimize performance or teach new domains. It's something you do at the end.

Sneha Mehra (00:42:40)  
Or in very specific cases. For reinforcement fine-tuning, ⁓ that's actually super useful to improve the way you reply to users based on their feedback. And ⁓ ideally you do that with an already good fine-tuned model. You don't train one from scratch using reinforcement fine-tuning. You do reinforcement fine-tuning at the very last step of the whole pipeline. And also the plus here is that you don't need much data ⁓ in this case. You just need

Based on what OpenAI says, a dozen or hundreds of examples rather than thousands, if not millions, to have very good tweaks in the models and get ⁓ very good responses in the style that you want. And for reasoning models, which we'll cover in the next courses, but that you probably have already tried before, you want to use them for more complex tasks, obviously, because they spend more tokens, they think before they answer.

And ⁓ that's also ⁓ pretty good for multi-step and logical questions like math code planning, mostly because, as we will see in the fifth course, they were trained mostly on math and code to generate plans and have specific responses to follow. Anyways, we will get back to that on the fifth course. You may also want to use that for ⁓ when you need a very detailed and structured explanations, when you need a detailed report or a very thorough.

response. If you know that you already need all those things, you can use the reasoning models right away. ⁓ The most important thing, as we said, is to measure and evaluate everything. ⁓ Prompting, rag, long context. Just have a base evaluation data set and try it. Try everything, compare everything and know what's best. Don't just try blindly. And we'll see that in the next course.

Now I leave Omar to cover structured outputs and practical notebooks for Rag, CAG, and PineTwein. ⁓ Hello everyone. So let's start with the second part of this session where I'll be showing a little bit more practical tips ⁓ on how we can build on top of LLMs. So in this ⁓ section, I'll be presenting what are structured outputs, why it's useful, ⁓ and

Sneha Mehra (00:45:05)  
I'll show some code on how we can implement structured outputs. And then I'll go over some notebooks. So, some code on how we can implement a basic chatbot that uses retrieval to answer questions. ⁓ And some notebooks about context-augmented generation. ⁓ And finally, some an example on how we can fine-tune a model, ⁓ a closed source model.

So let's start. So here we have an example of some text that can be generated but by an LLM. ⁓ And ⁓ as you can see here, it's a paragraph. And if we had asked about specific elements in the ⁓ text, here it's very hard to know if the model actually generated all the information that we wanted to see.

So, for example, if we wanted to generate a ⁓ like a description of some people working at different companies, here we wanted to see the name, the age, the role of the people, ⁓ and other types of information. ⁓ And ⁓ it's very hard to know if the model actually generated those elements. Here I highlighted ⁓ the main points we wanted to see in the answer.

And also the format in which the information is presented here, it's ⁓ it's within a paragraph, but the model could have put the same information in Markdown or in ⁓ JSON ⁓ schema. We we didn't give it the instructions to do so, so the the model will ⁓ kind of surprise us. We're not sure ⁓ how the information will be presented and if all the information is actually within the text.

So ⁓ that's why we like structured outputs, because it's a way to ⁓ manage ⁓ these kinds of ⁓ factors. And so ⁓ what is structured outputs? Well, it's a it's just a technique where we can make the models follow a predefined schema, for example. So if if we want the

Sneha Mehra (00:47:31)  
The answers to be within a JSON schema, ⁓ or if we want the answers to be within XML, for example, we can use structured outputs to do exactly that. Also, if we already know ⁓ the types of answers that we want from the models, we can ⁓ tell the model to use the ⁓ specified tokens. So if we already know that we want to know if it's if something is true or false.

Here we can make the model only generate ⁓ those two tokens. So ⁓ it's gonna be very easy to then parse the answer and know what the model answered. So for example, if we see ⁓ this example here, this string that says the meeting, the meeting is Monday at 1 pm with Francois, ⁓ the structured ⁓ answer output.

Would be here a JSON schema with the correct or like the the most important information of the string above. So if instead of having to deal with strings with characters and having to parse that string and not being sure what's ⁓ in the string if all the information is ⁓ is included here with the JSON, it's very easy to parse because we already have libraries that can parse.

JSON ⁓ formats, ⁓ and then it's going to be very easy to know exactly what type of information is missing. Yeah, it's gonna make ⁓ working with LLMs way more easier. So ⁓ just to summarize, the advantages of using structured outputs ⁓ is number one, the consistency. So ⁓ if we choose to work with JSON schema, then all of the answers will

Have to follow that format ⁓ and all the answers will be then consistent. It's going to be very efficient because we ⁓ no longer need to parse ⁓ strings, for example. It's going to be more way more reliable because we can ⁓ see if a value is missing or if if the model actually answered with the complete information we wanted. And so when ⁓

Sneha Mehra (00:49:56)  
Should we use structured outputs? Well, for example, it's going it's it's very useful if we want to extract information from a paragraph or ⁓ just unstructured text. So here we can get JSON schemas like we see on the screen where we have exactly what type of meeting, the time, the date, and ⁓ who is going to be included in the meeting.

It's also very useful when we want to ⁓ integrate our apps ⁓ with external systems. So let's say I want to use a web search API, like a Bing Search or Brave Search API, then ⁓ it's going to be very easy to use structured outputs in this case because I already know the fields I want to ⁓ send ⁓ to the ⁓ web search API.

And I already know all the other attributes, all the other ⁓ options I want to send to the API. So it here is very easy and very convenient to already get from the bottle, from the response, the ⁓ correct request that I can send then to the web search API. ⁓ And it's also very useful when you want to create structured content. So, for example, let's say you want to write financial reports.

And you already know that you want an introduction or some type of analysis and then a conclusion, then you can use structure outputs to make sure that all the sections are included in the model answer. ⁓ And it's going to be very ⁓ convenient to use in structure outputs in this case. How can we obtain these ⁓ structure outputs? ⁓ Well, there's two main methods that.

People developing these models ⁓ do to ⁓ allow us to use structured outputs. The first method is simply to ⁓ train the model to do it ⁓ and also include the instruction within the prompt. ⁓ And the second method is to use what we call grammar based constraints or context-free grammars to constrain ⁓ the ⁓

Sneha Mehra (00:52:21)  
The type of tokens the ⁓ model can output. So I'm going to dive deep within these two methods. But before I do that, I just want to specify that these two methods can be actually combined. ⁓ And when we use most ⁓ LLM providers, they give give us ⁓ access to both methods, for example. Here, OpenAI, we already know that we can use.

Method number one and method number two. So method actually we don't see the method, we just use ⁓ structured outputs, but in the background that's what is ⁓ happening. ⁓ And to ⁓ make it easy to ⁓ get structured outputs, ⁓ then we we can use tools like Pydentic if if we are working ⁓ with Python, for example, to easily create ⁓ the JSON schemas.

For our ⁓ the types of answers that we want to see in the response. So I'm going to ⁓ just give a little bit more details for each method. So the method number one, which is just to train ⁓ to fine-tune the models to produce ⁓ JSON schemas. Here, ⁓ OpenAI or whoever is training ⁓ these models just has to

Create a data set of prompts. So for example, respond or extract the information within this text. And I want the answer to follow a specified JSON schema. And then that will be in the input, the input, and the output will just be the JSON schema with the correct information. So if we train the models to do that with all of those examples, then the model will

Learn ⁓ to do that. So, for example, here ⁓ we want the model to extract the information of the string. The meeting is Monday at 1 pm, 1 pm with Francois, and then we give it a ⁓ a JSON schema that is empty with the fields we want to see in the answer. Then the model will be able to fill in the values ⁓ and we will have the ⁓ completed.

Sneha Mehra (00:54:48)  
Answer the the JSON with the correct values that we want to see. Now for method number two that uses grammar based constraints. Here we are not training the model to do it. Here we are mainly restricting the model's choices for the next token. So ⁓ on top of

Model, let's say we are removing all of the token options the model can generate. And for example, if we want only the tokens true or false, then the model will only have the choice of using those two tokens. So this is something that ensures that the answer will actually follow the instruction or follow the format ⁓ at a very reliable

Rate. It's going to be ⁓ basically ⁓ good all the time. So here, for example, we have an LLM and we know that ⁓ LLMs produce probabilities for the next token. So let's say here the LLM for the next token will be the next token here will be here at 22%. ⁓ What we can do is to apply a mask and here

⁓ Put all the other tokens at zero and only leave the probabilities of the token ⁓ that we want to see in the answer. ⁓ So for ⁓ those two tokens, the tokens false and true, we leave the probabilities as is. ⁓ And since we put all the other probabilities to zero, then ⁓ here ⁓ the token that the model will generate or that we

Basically, here the token that will be chosen by default is the highest number. And since all the other ones are at zero, then false will be the next token that the model would generate, basically. And that's what we call constrained sampling because we are constraining the number of tokens that the model can use to gener ⁓ to output ⁓ a response. ⁓

Sneha Mehra (00:57:08)  
To do this ourselves, we need to have access to the probabilities of the model. So this is very easy or easier to do with open weight models because we have complete access to those probabilities. We know exactly what's happening in the in the model. ⁓ But for closed source models, not every company is willing to give that information. So so for example, with Gemini or other companies, it's not something that we can

Do ourselves ⁓ this method of grammar-based constraints. This method can become actually very complex. So let's say the model generates these three function calls at the same time ⁓ within the same answer. ⁓ Then ⁓ the model has to make sure that the brackets are correctly ⁓ in ⁓ the right position or in the right order, the commas.

Everything has to be in the correct order for these tool calls to work. So ⁓ doing this constraint sampling method can become very complex because you have to keep track ⁓ of what you generated before to make sure that you actually complete the string, you complete the bracket, you complete the ⁓ the values correctly. So to make it

Easier or or one way to do it in an easy way is to use what we call we call context free grammars. So the this thing ⁓ is the ⁓ context-free grammar is basically a list of strict rules that the model will need to follow when it's generating a response. So for example, I want to make sure that the rules

Will constrain the model to only generate a valid JSON. So here I put a very simple example of what these rules can look like. For example, here I start, and then the next token will need to be a greeting. ⁓ And here the green greeting is defined as hello ⁓ or hi. So here the model has two available choices, and the model will choose depending on what has the highest probability.

Sneha Mehra (00:59:32)  
Then ⁓ the greeting is followed by a comma, then by a space, and then by a name. ⁓ And then here, name is also defined by two strings. So there's Alice and then there's Bob. The model will choose ⁓ whatever has the highest probability, and then we will end ⁓ the answer with the ⁓ exclamation point. So this is what we ⁓

called a strict rule ⁓ or a context-free grammar to have ⁓ an answer that that follows exactly what what we want to see in the answer. Then ⁓ after we define the rules, ⁓ we need a decoding algorithm that follows the progression of the text according to this grammar.

So for this ⁓ decoder, it can either be a parser or a finite state machine or an automaton. And at each step, so ⁓ while it's answering, ⁓ only the tokens that follow the rules will be allowed for the model to use. And I'm not going that much into the details because this process or these algorithms are all

Already implemented and are very fast already, entirely automated. And you can use them when you use, for example, libraries like Outlines or Lama CPP, ⁓ or even the Open API if you're using the one of the OpenAI models. So ⁓ behind the scenes, ⁓ these libraries will create ⁓ the grammar, the list of rules to follow.

And then have a decoder algorithm that will ⁓ make it so that the model can only output the tokens it needs to output. So if you want to see a complex JSON response with many, many indented keys or many layers, then the the model will have to follow the rules, and then at the end you you will have an answer that actually follows exactly the rules that you that the ⁓

Sneha Mehra (01:01:54)  
The library ⁓ set at the beginning. I just want to ⁓ say that different implementations ⁓ can have varying speeds. ⁓ And for example, outlines, ⁓ their implementation is quite faster than the OpenAI1. So that's interesting to note. ⁓ And so if you're you mainly using open source or open weight models, then I recommend you use outlines.

If you have to use OpenAI models, then yeah, you will have to use the OpenAI implementation. Then you have libraries like Instructor that will use whatever implementation is already available from the different providers. Instructor is very useful when you ⁓ are defining the different JSON schemas that you want to see in the answer.

⁓ And I I will come back to this in the ⁓ in the coding example. So how can we use these methods in practice? Well, like I said before, we need to ⁓ specify the answer we want to see. ⁓ And to do that, we can use libraries like Pydantic if we are using Python. ⁓ So what's Pydenc? ⁓

So basically Pydantec is ⁓ is a Python library that can help us define and validate data models. For alternatives, you can look at Zod for JavaScript ⁓ or libraries like Fluent Validation for C sharp. And so ⁓ what does it look like when we use Pydantec? Well, basically, here we can create what they call a base model.

So let's say I want to create a user. I give the user an ID and I specify that it's an integer. And I also say that this user has to have a name, ⁓ and this name has to be a string. So then I can create one user with an ID ⁓ and a name. And then I have the final pythantic type object. ⁓ This is very useful because then it's going to validate.

Sneha Mehra (01:04:16)  
that the ID is actually an integer and that the name is actually a string.

Then ⁓ when we create ⁓ this ⁓ class, this user, we can then use the Pydantic library model dumped JSON to create the JSON representation of an object ⁓ or a base model. So this is ⁓ going to be very useful for ⁓ defining the LLM answers.

So for example, I can say I want to see what type of event, the date, the hour, and the people that will be in the event. So let's say I have this. I can then pass ⁓ the JSON representation of this ⁓ base model, this event, and I can prompt a model to answer me and to put the information in the ⁓ exact format that I am specifying.

And when I get the answer, here Pydantic can validate that the ⁓ LLM answer actually follows the schema, the types that I set above. And it's going to be very useful for that. ⁓ So if I show you a complete example, here I'm showing something that's already in the OpenAI documentation. So here I have a calendar event.

And I want to see the name of the event, the date, and a list of participants. And I'm actually also giving ⁓ types for each ⁓ attribute or each variable that I want to see in the calendar event. And then when I make a request to get an answer from ⁓ OpenAI, I can then give it the instructions. So for example, extract the event information.

Sneha Mehra (01:06:17)  
I give the model the string that contains the information I want to extract. And then within the same API call, the same request, I also give the ⁓ the model, the system, the calendar event. So this is the Piedentic object that I define here above. And when I receive the answer here in the event variable, I can then easily access.

The name, the date, and the participants variable within this ⁓ response here, the event. So this is very useful when you're working on top of these LLMs. So now I'm going to show a quick example that it is basically the same thing, but ⁓ here I instead I'm going to show how to do it with the ⁓ instructor library because

it's it's very easy to ⁓ to then swap whatever llm provider i'm using so here this is ⁓ a notebook and i'm using collab to run the code it's if you don't know about collab it's very useful for for this for these kinds of demonstrations so here first i i'm i'm going to show you the text that contains information ⁓ and from this ⁓

Text unstructured text, I want to extract specific information ⁓ that I want to then ⁓ either ⁓ save to a database or to do some type of analysis ⁓ or whatever. I just want to from this text extract the relevant information that I want to see. So to do that, I'm going to use Pydantic ⁓ and I'm going to ⁓

For example, I will define a person object, Pynantic object, and then ⁓ this person will have a name ⁓ with a string, and the string I want ⁓ the size of the string to have a maximum length of 50 characters. So this can be very useful if you were validating that the ⁓ information that is extracted is correct. Then I want to see the age. So there's

Sneha Mehra (01:08:40)  
DH can can either be an integer or a string. This is just for demonstration ⁓ purposes, because here ⁓ I will either want to have everything in integers or everything in strings. It doesn't make sense to ⁓ here to let the model choose the type within. So yeah. Here I'm just giving ⁓ a choice. Then for the role, I want it to be a string. And then here I'm giving it.

description. So this is very useful because sometimes if you're giving this ⁓ information to the models, the the JSON schema that is empty, sometimes the models don't know exactly what information should be in the key. So for the role key here I'm specifying that it's the role of a person within a company.

And I also give it an example. So, for example, a developer. Then for the company, here, ⁓ since I know that I'm working within a specific number of companies, ⁓ for whatever reason, I already know ⁓ the number of companies that will be in the text. Here I can specify the list that the model will be able to choose from. And if

A company is not within that list, then the model can choose to put the other string for the as answer.

Then, ⁓ as we saw in the first ⁓ session, it's very ⁓ useful ⁓ for non-reasoning models to instruct them to ⁓ reason before they answer, to think step by step before they actually answer the question or the instruction. So that's what I'm ⁓ doing here. I'm forcing the model, I'm telling him to

Sneha Mehra (01:10:46)  
To write a reasoning string to write how ⁓ the how the model needs to approach, how it will approach solving this task. ⁓ And finally, here as second key, I want the actual people list. And this is going to be a list of person, ⁓ person which I defined here above. ⁓ And I want a minimum of

one actually three ⁓ people here it's because i already know that in the text there there's three people that are mentioned in the text. So now for the second step ⁓ once ⁓ I finish defining the ⁓ the pygantic models then I can create the API client that I'm going to use to make the requests. Here I

Define the text, the unstructured text. And here as fourth step, I write the prompt. So I say to the model, extract all the people mentioned in this text, and I put the text in the prompt here as a variable. And here on number five, I make the make the requests, and here I display the answer.

I just want to specify that I used Gemini 2.0 flash. So this is a very cheap, very small model that you can use. And here for the answer, we can see the people extraction variable that the model answered with. We can ⁓ actually with code see the all of the keys that are defined above, all the variables.

So if I go to the reasoning, ⁓ I print the reasoning, I can see that the model said, ⁓ I will extract the name and the age and the other things. ⁓ I will iterate through the text. ⁓ And if the company does not exactly match the list, I will use other. And then I print within this for loop all the different people that the model extracted.

Sneha Mehra (01:13:11)  
And I can see for person one, I have the name, the role, the company, the age, ⁓ or ⁓ I can also ⁓ have the string representation of the same information. So as you can see, ⁓ using structured output, working with LLMs is much more easier than trying to ⁓ parse the unstructured answer from the models.

Here I'm able to ⁓ directly see how many people, what the different attributes of all the different people. ⁓ I can ⁓ make the model think before it answers. I can do ⁓ basically everything that I can do without structured outputs, but here I'm ⁓ I'm using structured outputs to actually make it very easy to work ⁓ in code with LLMs. So as you can see, it's very very useful to do.

Now, if I show ⁓ some ⁓ more base model examples. So let's say I want to extract the location information from ⁓ job posting. So let's say you have a job posting on LinkedIn, for example. If you're you're looking at job job postings, you see a lot of unstructured text, a lot of information. ⁓ And if you want to, for example, ⁓ extract key.

things from on structured text then you can use like I showed pydantic to define the information you want to see. So here I'm just showing an example of trying to get locations from job postings. So here I have the chain of thought attribute, the variable that will contain the ⁓ the the step of step-by-step reasoning of the model. Then I want to see a list of cities, list of regions,

I specify the type, so list ⁓ or string. ⁓ I specify the number of items the list ⁓ should have. So for example, I want to have at least one ⁓ item in the list, one element. ⁓ I show examples, I show description. So all of this ⁓ is just extra information ⁓ so that the model can extract the correct information and put that into a JSON schema.

Sneha Mehra (01:15:37)  
So everything here is just to make it easier for the model to actually so that the model is able to extract the information they want to extract. So ⁓ if we have the salary, for example, we we want to know exactly what salary ⁓ is given by each job posting. We can ⁓ do the same thing. So we have the reasoning at the top, is because the model

Generates from top to bottom. So that's why it's always at the beginning of the answer, basically. And for the salary, ⁓ we can either have a list of floats ⁓ or strings, because sometimes the information is not contained in the structured text. So we want to give the model the choice to put ⁓ if the salary information is not present in the in the text.

Then I have more examples of using literal. So this is ⁓ something specific to Python. But here I'm I'm saying to the model, these are the options, so that I don't see different types of ⁓ answers that ⁓ later on will be harder to work with. So if I'm already know that I I want to see hourly, monthly, annually salaries, then the model will not try to come up with.

Other ⁓ time time frames for salary, for example, like a daily salary or weekly salary. I don't want to see that, I want to see hourly, monthly, annually. And then this is where I make the request. ⁓ So ⁓ here would be the Bydantic model that I defined above. Here is the ⁓ prompt.

With the ⁓ the query, so the the query here will contain the struck the unstructured text that contains the information, the the paragraph, the page, ⁓ the whatever information you have. ⁓ And ⁓ yeah, that's how you make the request. And just to show you quickly, I recommend you check out all of those examples. ⁓

Sneha Mehra (01:18:00)  
If you want to learn how to use unstructured outputs, because ⁓ yeah, it can make working with LLMs way more easy. Besides using you know frameworks, more complex frameworks. So let's go to ⁓ RAG. So in the first section, Louis talked a lot about how we can ⁓ implement a RAG system within our applications. And here I just want to show

With simple examples, how we can ⁓ implement such systems. We're not going to show a complicated or complex examples, but it's just to show you how simple it can be, how yeah, and quick it can it can be using Python or some other library. ⁓ So if I show an example that uses the Google Gemini API. ⁓ So here let's go to the top.

Here I will be ⁓ using two APIs because the LLM I'm going to be using is actually Gemini. But in order to create these vector representations, these embeddings, I'm going to use ⁓ an embedding model created by OpenAI, and we can access it by doing API call requests.

And so that's why I'm here. I'm setting ⁓ I will be setting two API keys and not just one. Then here I'm just importing all of the data that I want to work with. Here to just put some background context. ⁓ we want to build a chatbot that is able to look ⁓ at different articles that

Talk about the Meta Lama model. So if we have questions about the Lama model, then this chatbot needs to look at those articles, see if the information is there. And if it's there, I want to see that in the final answer of the chatbot. So here I'm just downloading these articles in the locally here in the Indie Colab.

Sneha Mehra (01:20:24)  
⁓ And we can see here that we have actually ⁓ 14 articles. ⁓ And if we ⁓ actually split ⁓ those articles into chunks, we have a hundred and seventy-four chunks. Just to make it clear, here is our chunking function. So here basically we divided the articles by chunks of

1024 characters. So this is not the optimal way to go about it, but it's one way to go about it that is very simple and ⁓ works in most cases. But yeah, this is very ⁓ keep in mind that this is just for demonstration and ⁓ and and just just to show quickly how we can build a chatbot.

Now, once we have the articles that are chunked, now I want to create here for this demo ⁓ a pandas data frame. ⁓ And I'm going to put all of those chunks within a single column. So ⁓ let's imagine that in within my pandas data frame, I own right now I only have one column and it's the chunk.

Column and for each row I will have ⁓ all of the different chunks in my dataset. Then here I'm ⁓ defining what is the the get embedding function. Here we are going to be calling the text embedding tree small model, which is a very small model from ⁓ available from in the OpenAI API. So

This embedding model just needs ⁓ text ⁓ and it will output a ⁓ an embedding with which is just a list of numbers that ⁓ represents ⁓ this text specifically.

Sneha Mehra (01:22:36)  
So, what we do here is we just loop over all the rows, ⁓ and then we create the embedding for each chunk in the data frame. So now we have two columns. We have the ⁓ chunks, ⁓ columns, and the embedding columns. So for each row, we will have the text, the strings, and the embeddings for each of each one of those chunks. ⁓

If we create write a question, so for example, I write how many parameters does Lama 2 has, and then I get the embedding ⁓ of the question I just wrote, then ⁓ I can see exactly the number of values that are within this embedding. So we have 1536 number of values for that specific question.

So this this will not change. This is specific to that embedding model. This is the the number of ⁓ values for each vector representation, ⁓ the number of dimensions ⁓ in the the output. Now let's show how we can ⁓ use cosine similarity. So here I have two questions. ⁓ And I have, for example, the sky is blue ⁓ and lemma 2\.

Model has a total of two billion parameters. So these are two different strings. Now let's compute the the embedding for both questions. So I I will have two embeddings. And now ⁓ using the cosine similarity function, which I import from this library.

I can see the ⁓ the similarity score of the question ⁓ with the ⁓ question embedding ⁓ and the question above. So we have this question embedding, ⁓ the how many parameters Lamatu model has. And I compute the score ⁓ of this question above with the two ⁓ answers below.

Sneha Mehra (01:24:55)  
And we can see that for the answer, the sky is blue. We can see that the score is actually really, really low because ⁓ the answer doesn't match ⁓ the question within the ⁓ the embedding space. ⁓ And ⁓ in the contrary, for the ⁓ answer, the second answer here, we have a high score for the similarity.

of eighty three percent. So this is just to showcase that similar pieces of text will have a higher score than ⁓ two pieces of text that don't mean the same thing, basically.

Now now, if I do the same thing ⁓ of taking the question and comparing that to each of the chunks in my dataset above, then I I will have actually 174 scores logically, naturally, ⁓ and I will be able to sort ⁓ the answer to get the chunks that get the higher score ⁓ for my question.

So I can print ⁓ the indices that have the highest score. ⁓ And we can see that the chunk with the index ⁓ with these three index. ⁓ These three chunks actually have the highest similarity score. If I print the chunks for each of those, I we can we can see them here. ⁓ And

One of them ⁓ should con contain the answer to my question. Now when I ⁓

Sneha Mehra (01:26:45)  
Build the request. This is basically the chatbot right now, the the rag chatbot. Here I have the system prompt. I have the prompt. I have the ⁓ the model that I want to use. And I have the ⁓ yeah, basically the request to ⁓ to Gemini. Then if I if I print the request.

I'm going to see the answer. So Lama 2 is available in four sizes: 7 billion, 13 billion, 34 billion, and 70 billion. This is basically what ⁓ what a rag chatbot is. So basically, here I have the question. ⁓ I ⁓ I filter the indices, the chunks that have their the highest ⁓ similarity scores. ⁓ Then I take the text.

Within those chunks, and I just append that to the prompt. So within the same prompt, I'm also adding the context that contains the information. And then the because of that, the chatbot will be able to answer me correctly. ⁓ So this is a very, very simple example of what a retrieval-based ⁓ system looks like. You do the retrieval, ⁓ you get the text.

And then you add that to the prompt so that when someone asks a question, then the model can check the information in the prompt in the context and answer correctly. So here I have another chunk with some information that wasn't in the articles. Here ⁓ I ask how many parameters the newer model has, the model 3.1. Here I add ⁓ this. ⁓

Example chunk in the prompt. So this is the correct information taken from the meta documentation.

Sneha Mehra (01:28:50)  
And if I and we can see that we get the correct answer here once we do the request. And if I do the same thing ⁓ without adding the correct context, then we can see that the model is not able to answer me because when it was trained, it didn't have that answer. ⁓ So ⁓ we can see that by adding the correct context, these models are able to read the context in the prompt and answer ⁓ reliably.

So this was a very basic example. Now let's look ⁓ at the same example, but now let's use a framework. There's many of them. You can use ⁓ Lama Index, for example. They specialize in ⁓ retrieval for these exact purposes, ⁓ for these ⁓ kinds of chatbots. So if I do the same thing, I download the data set.

I open the CSV files with the 14 articles. ⁓ Here ⁓ I'm going to create what we call ⁓ Lamma index documents. And I'm going to put all of the different content content of the articles within each ⁓ document. ⁓ And each document has a text ⁓ attribute. So I'm just putting the content, the content of the articles ⁓ in the text variable.

And here I'm printing the contents of the first article. ⁓ And this is where I create the index. So this is what we call ⁓ a ⁓ vector store index. ⁓ Here, Lama index makes it very, very easy to create one ⁓ from a list of documents. So here I'm just specifying the list of documents that I'm I want to use. ⁓ I'm saying exactly what the chunk size should be.

What should be the chunk overlap? What should be the embedding model? ⁓ I just want to specify that these options should be ⁓ optimized based on evaluations. So here we are just showing how we can create a ⁓ very quickly a vector store index, but the next step would be ⁓ to try to optimize it. ⁓ So once we have the index, we can then

Sneha Mehra (01:31:21)  
So this is going to be very quick. And then once we have the index, we can specify the model that we want to use for the ⁓ answer generation. Then if I do a quick test here, I do a query ⁓ using the index as query engine. I ask a question. So how many parameters Lama 2 has? And it's able to answer me correctly.

And if I ask a question about Lama 3, then it's it's not going to be able to answer me because the articles above do not have the information for Lama 3\. You can you can see that ⁓ to create the same chatbot as I showed in the in the other notebook, here it was really, really easy to create. But ⁓ as I said, the the next step.

That we are actually going to show in the next session will be to evaluate and iterate ⁓ over this ⁓ rag system. ⁓ Now, what embedding model to choose? Well, there's many, many different embedding models. Ideally, you want the embedding model to know about the type of data that you are working with. So here I'm just showing ⁓ a list of models taken from a leaderboard.

You can try to use some of them. So, as you can see here, there's Gemini embedding model that is okay. Then you have the ⁓ the Snowflake embedding model. You have many, many different ones to ⁓ play with, to try. But as I said, to really know ⁓ what model you should choose, you need to to evaluate your system to know exactly the ⁓

the the scores that your retrieval s system is is able to get. And also ⁓ it can be very beneficial to actually train an embedding model because sometimes the type of data that you have ⁓ is not in the training data that was used to tr to train ⁓ these ⁓ open public models. So ⁓ some of them actually can be fine tuned. So that that's that will be like a

Sneha Mehra (01:33:44)  
A different optimization avenue if you want to increase your retrieval scores. But we will talk more about this in the next session. So, like I said, it's it's good to show ⁓ to look at the leaderboards, but it's way more useful to check to actually evaluate those systems. So we'll want to measure things like the precision of your retrieval system.

The recall of your rag ⁓ app, the recall MRR, ⁓ and other ⁓ measurements if they are relevant to your task. So now let's change topics. Now let's look at context augmented generation, which is the same thing, but here we are mostly talking about bigger context. So let's say you have a book.

Or you have a big number of documents that will remain the same. ⁓ Here using ⁓ a GAG will be very useful. So different providers ⁓ make the ⁓ functionality available. You have Gemini context caching. So for example, if you have ⁓ a specific number of PDF documents that you want to cache in the system.

You can do so and just pay by the number the number of tokens in those documents. The default time that the content will be cached is one hour, but you can specify that in the API. Then you have automatic systems. So for Gemini, it was manual. You had to specify these are my documents. I want to cache them here with OpenAI. ⁓

It's done automatically. So whenever your prompt is above around a thousand tokens, the system will automatically cache the prompt and the cache will be remain will remain active for at least 10 minutes of activity. So it means that if you ask a a very ⁓ a long a question with a long context ⁓ and you ask the same context an hour later, then ⁓ it's

Sneha Mehra (01:36:10)  
⁓ the system will need to recache ⁓ the the content in in your documents. And you have authentropic prompt caching, which ⁓ works similarly to the OpenAI prompt caching feature. Right now ⁓ they they pretty much all have the same type of features and they really help in in reducing costs because once the content is cached, then you don't have to pay the compute for those tokens.

And it's also something that can speed up ⁓ the answers so their latency can be reduced. If I show quickly ⁓ how to use Gemini context caching, here I have a dataset that contains documentation from Lama Index. ⁓ And it has a hundred and fifty K tokens. So this fits ⁓ within.

⁓ Inside the context window of Gemini at a hundred and fifty to ⁓ K tokens.

Now, here I'm going to create again a rag bot ⁓ using context caching without and also comparing to a chatbot that doesn't use context caching. So here that's why I'm creating ⁓ Lama Index documents. And here I will create once again an index ⁓ and query ⁓ this chatbot. And you can see that.

For the chatbot it took four seconds to answer. So I asked how to set set up a query engine in code and it answered me in four seconds.

Sneha Mehra (01:38:02)  
Now let's do the same thing with Gemini context caching. So here ⁓ I create ⁓ the file upload. So I here I'm converting the JSON L file into.txt because at the time it it didn't ⁓ support JSON. I'm not sure if if it's the case right now, but basically it's the same content but it in a text file.

Here I specify the model that I want to use for ⁓ the cache ⁓ and later for the ⁓ chatbot. So I create the cache. ⁓ first I upload the document and then I create the cache using this Gemini SDK for the Gemini API. ⁓ And here I'm going to ask your question ⁓ using the cached tokens.

And I want I I also say that I want a maximum of thousand tokens for the answer.

And I can see the answer, ⁓ I can see the metadata of the answer. So in total we have 200,000 tokens used. Actually, ⁓ the the number of tokens cached was actually 200,000. So it different from ⁓ the file name ⁓ because it was using a different tokenizer basically.

But it shows that ⁓ here it answered me in ⁓ two seconds ⁓ instead of four seconds. So we can see that it started answering me way faster than the ⁓ Lamaindex chatbot that didn't have the context cached. So this is just an example of using context caching. Now, how we can how can we fine-tune LLMs? As Louis said, ⁓ it's

Sneha Mehra (01:40:04)  
Something we choose to do mostly after we have tried optimizing everything else. So when we don't have really the choice, we actually go into trying to fine-tune a model. And to make things even easier, we can start with closed source models if we we can use ⁓ closed source models, if the data can be shared with OpenAI or if if the company we are working with.

⁓ actually allows to use closed source models. ⁓ And here ⁓ to make things easier, we can actually use the services that are provided by these closed source models. So, for example, we can use the OpenAI platform. And just to put some context on ⁓ what we want to do, let's say we have a system, we have an AI tutor that is able to answer questions re regarding documentation of various.

AI libraries. So let's say hugging face, llama index, lengtchain. Right now the system is using ⁓ GPT-4.0 and then some retrieval to be able to retrieve the correct information, add that to the prompt, to the context of the prompt, ⁓ and then get answers that have the correct information. So right now the system is working pretty well. ⁓ It gives us complete and useful answers. But now

We wonder what if we use GPT-4.0 mini? Can we have the same quality quality of answers? And if that's possible, if we get the same ⁓ level of quality, then we can save a lot on tokens, on costs for ⁓ for the system basically. So now let's see how we can ⁓ actually

Use the OpenAI fine tuning service. Here I basically already created the dataset for the fine tuning ⁓ example. So ⁓ basically, we we have a JSON L file with an input and an output. We have questions ⁓ and the answers, the output ⁓ is ⁓

Sneha Mehra (01:42:28)  
Written by our current system that uses GPT-4.0 and that also includes the retrieval part done before. So here we already have that data set ⁓ of questions ⁓ and correct answers. ⁓ And we want to fine-tune GPT-4.0 with this dataset. And here, the only thing that ⁓ we are doing here is downloading that dataset from Hugging Face. So right now it's

Stored in the hogging face ⁓ and I am downloading it ⁓ here in the in the notebook. Now if I create here again ⁓ a quick chatbot using Lama Index, ⁓ I can see that it's able to answer me. So I create the vector store, I create the LLM, and then I create the query engine.

We can start asking questions and see the results. ⁓ And we can see the sources ⁓ used for each for each question. So here I have ⁓ all of these chunks. Lama index calls them nodes. And each ⁓ chunk, each node has ⁓ a URL. The URL that points to the correct content, basically.

So I can this is useful because then I can cite whatever information the model is ⁓ using to answer. I I can tell the model please cite your answers using the URL, for example. And then people can then ⁓ click on the URL and confirm that the information that the model ⁓ answered with is correct. So here I have another question and another example.

Here I'm just validating ⁓ that the that the dataset is in the correct format. This function validate dataset will validate that my dataset contains ⁓ the the key messages. ⁓

Sneha Mehra (01:44:44)  
The roles, the content. So ⁓ this is important because we want the inputs ⁓ to be assistant or ⁓ user inputs. So this is the role key. So this is just basically making sure that the data set has a correct open AI format. So this is not important to ⁓ memorize. Here I'm just counting the number of tokens in each.

example each sample ⁓ and here here is where I will ⁓

format the ⁓ the dataset. So this is what I was talking about. My my dataset that is currently stored in hugging phase and right now in the notebook has questions and answers. But I want my dataset to fit within the correct format that OpenAI ⁓ needs ⁓ to

for for this data set to work for the training ⁓ for the training run. And so I'm converting ⁓ basically every question and answer into this ⁓ message ⁓ in into this ⁓ JSON, basically JSON object. ⁓ And at the end, I'm going to have a JSON lines file where each line will be a JSON with these ⁓ with this exact content. And so I'm going to

Basically prepare the dataset ⁓ and upload that to ⁓ the OpenAI fine tuning service. So you can either go directly in the OpenAI fine tuning service and drag and drop your dataset, or you can actually do it with the OpenAI API. You can use code to upload your file to the fine-tuning service. So let's just go back.

Sneha Mehra (01:46:47)  
To the presentation. Here I have some screenshots to make it easier to follow. So here ⁓ we are going to go to the platform. So ⁓ you see the URL, platformopenai.com. ⁓ Or you can go to the fine tuning section here in the platform. And then you're going to create a new fine-tuning job. Here you can

Basically, choose what which model to fine-tune. So you can fine-tune here in our case we want to fine-tune GPT-40 mini, so that's why we choose that. Then we select an existing dataset. ⁓ So you can either, ⁓ like I said, upload this ⁓ using the ⁓ the website, or you can do it in code using the API. Here I'm we can also create

⁓ specify what is the training data and what is the validation data. So this is whenever you are training AI models, you need to measure performance. ⁓ And you cannot measure performance on the data you are training with. Because of course the model will be will get better on the training data. But the point is that the model needs to learn how to generalize to other types of questions. So

We need this validation dataset which contains other questions that the model will not see during training. ⁓ And we want to measure the performance on that validation dataset. ⁓ So here we also specify what will be the suffix. ⁓ So when when we will use the model afterwards in our application, instead of putting GPT-40 mini slash ⁓ and the rest of the string, we will

Use this string but also append the suffix so that we can use the fine-tuned model and not the general one. Then here we need to set the number of epochs and batch size ⁓ and the learning rate. These values are given to us automatically, ⁓ but ⁓ during training, during our experimentations, we will try to optimize them to get the most performance.

Sneha Mehra (01:49:13)  
So usually you don't train ⁓ one time, you actually train multiple times, and you play with these values to actually get try to get the most performance, the highest score, the lowest loss on your validation ⁓ evaluation. So ⁓ once finishes training, you will have this succeed status, and you will be able to see.

The the graph of the loss, the training loss, and the validation loss. As I said before, you want to mainly look at the validation loss because these are the questions the model never got to see. ⁓ And also you can see that for the ⁓ training loss, it's usually a little bit lower because of course ⁓ the model got better on the questions it actually saw during training. So.

So this is just one training run. ⁓ And ⁓ you want basically to make sure that you have the lowest loss possible for your task. ⁓ And I think that's ⁓ that's it. Here I'm just yeah, you can you can also check on the status of the training job in code. You can look at the OpenAI API, it's very simple.

And you can start using actually the model directly in in the with the API. So this is not very important, but yeah, it was mainly to show you how easy it is when you train these models using services like the OpenAI one. It requires a little bit more complexity when you're dealing with open source, open weight models, but yeah, it's basically the same process.

But we recommend that you actually start with these services that make it a little bit easier to start with. ⁓ And that's it for the second session. ⁓ I hope you learned something. For the next session, it's going to be all about evaluations. ⁓ So, how we can take these chatbots, these retrieval systems, how we can measure the performance to actually try to optimize it for whatever task you are doing. ⁓

Sneha Mehra (01:51:31)  
See you for for the next session. Goodbye.

—-------------------  
1

Sneha Mehra (00:00:09)  
Good morning everyone. This is Louis Francois. I'm the co-founder and CTO of Tourzi, but I'll introduce myself and the second speaker, my colleague Omar Solano, in a few slides. We just wanted to start with a quick introduction on why we are here and why this course first. So, this course, what is it? Well, it's called Operation and Development with Large Language Models. This is just the name of the slide, but it changed on the website, obviously. So, the first session here.

Is the foundational knowledge to know for working with LLMs. So why are we here? It's to learn large language models. Why? Well, it's for many reasons. The first one is because it's no longer just buzzwords anymore. Even though it's full of buzzwords like agents, agent tick AI, generative AI, whatever, they are not just buzzwords anymore. We can do real things, real products, real applications. ⁓ Here, just a ⁓ a quick

⁓ example of what it can do. ⁓ Just in in recent reports, the company Klarna and Amazon both reported to save many millions of dollars thanks to artificial intelligence or language models in this case being implemented in their workflow. And likewise, it saves me and my team tons of hours. Another thing is that there are increasing ways to avoid mitigate errors, which we'll all cover in this course. Also,

Early adapters gain relevant expertise before it becomes mainstream, which is super important for anyone's portfolio when working in the industry. Also, I just want to note that it's really bad, it can be really bad at first, but just give it some time and learn to use it, especially with a paid tier. For example, I I've been in the field since 2019, and even though ChatGPT came out in November 2022, I think I started really using it.

Just in June 2023\. So it took me maybe six or seven months of hearing about it, learning about it, and even trying it out to finally figure out, okay, I can use this for that, etc. So it can take time, and you have to figure out how it can be useful to you. ⁓ Another reason why is that for from a recent Anthropic study, so Enthropic, the company behind the cloud series of models, share that.

Sneha Mehra (00:02:34)  
37.2% of all the requests they receive ⁓ are for coding. And that's from only 3.4% of people that are developers from their audience. So that's just to show how useful LMs can be for developers, which is why we build this course for developers to better use LLMs. So it just proves it's good. So just as an example, sales and marketing are only 2.3% of all requests entropic.

receives in the same study. But that comes from 8.8% of people. So it comes from even more people, but still it's fewer requests. Which means the people in marketing and sales don't really like ChatGPT. So there's definitely an opportunity to build something there and for many other industries as well where ⁓ you can implement domain expertise and many other useful skills

Or ⁓ useful ⁓ systems around the LLMs to make them more powerful. And that's for why to learn to use LLMs. But why learning how to build with LLMs as a developer? Well, it's for many reasons again. But the most important one is because it can actually be linked to existing products and provide real value, which is something that is

Relatively new in the LLM space and even in the AI space. So it's very important to get going right away and as soon as possible. ⁓ Also, LLMs are not good enough out of the box. Well, they are becoming better and better, but you can't often use them right away for your task or in your application. And fortunately, there are increasing ways to avoid or mitigate errors like hallucinations.

Including ⁓ everything we'll cover in this course, like better prompting, retrieval, fine-tuning, implementing tools like web search and more. ⁓ It also brings a competitive advantage to your app if you implement it correctly. And you don't need huge budgets as you used to with AI. Big companies already train huge models. You can just leverage them easily. This is just something quite new but incredible.

Sneha Mehra (00:04:56)  
Companies like OpenAI spend millions and even billions in training and making these models as best as possible, and they make it super cheap to use right away. So why not leverage that? ⁓ Also, what's I guess the best point here is that any software developer can do it with some upscaling and practice, which we hope to provide here. Another reason on why to learn to build, or rather

What we try to teach here is pretty much what is called the AI engineering stack. So this comes from Swix, ⁓ an interesting person to follow in the AI space. If you don't follow him already, he's quite everywhere, in fact. And he has a newsletter called Latent Space. It's also a great podcast ⁓ to listen to if you want to hear about the news and better understand the field in general. But in any case,

The AI engineer stack is pretty much a new subfield between a full stack engineer like ourselves, ⁓ people working on product developers, and ML researchers, research scientists, ML engineers, the people that prepare data, train models, test models, do evaluations, etc., maintain models, etc. So basically, AI engineers.

⁓ is is like the bridge between the two. They use APIs or sometimes ⁓ models directly. So they need to understand a bit of everything. And so our goal here is to teach this bit of everything to the ⁓ software engineers. But the course is not ⁓ only to software engineers, obviously. ⁓ It can be pretty beneficial to anyone. ⁓ So we definitely invite you to keep listening and check out the other courses, even if you are not a developer.

The reason why it's it's for developer is because we are talking about APIs and ⁓ have some programming example using Python, but you can still follow along ⁓ even if you are not a programmer or you don't use Python in a day-to-day. That's just because ⁓ the whole, as we will see pretty shortly, the whole course is built in a half-and-half ⁓ manner where the first half is always a bit more theoretical.

Sneha Mehra (00:07:21)  
But always useful theory to know ⁓ in the field. It's not like math or ⁓ less relevant concepts for an industry. ⁓ We always try to just keep it the bare bones ⁓ of things you need to know. ⁓ And the second half is a bit more code related, but it's still a ton of industry-specific tips and things we discovered while working in the field. So who we are to teach this?

⁓ myself, I'm Louis François Bouchard. As I said, I'm the co-founder and CTO of Tour CI, and I used to be a PhD student at Mila ⁓ and Polytechnic Montreal, that was affiliated. ⁓ I'm from Montreal, and I was a student at Mila. I've also been an educator in AI since 2019, starting on YouTube under the name What's AI, where I used to share and explain research papers. ⁓

Which I did every week since I think late 2019 to early 2020, since ⁓ maybe last year, where I left the PhD to focus fully on towards AI and thus had a bit less time to read papers and cover papers. But, anyways, I've been creating educational content around AI ⁓ s for many years, and now I'm fully dedicated into creating courses and educational content.

Still for artificial intelligence, but more focused on the industry now. And also I've I've started working in the field since 2020 at a company, a startup, where I worked with Omar as well, called Design Stripe, which no longer exists. And after that, ⁓ I worked for Toward the Eye, ⁓ with Toward the Eye, my company, ⁓ where we give consulting, we give trainings, etc. I will now leave my friend Omar Solano.

Colleague to introduce himself. ⁓ Hello everyone. So ⁓ I will be the second presenter of this course. My name is Omar Solano. ⁓ And just for a quick background, I'm currently doing a master's degree at ETS in Montreal, an engineering school. I've also worked in the industry since 2021, ⁓ so mostly as a machine learning engineer.

Sneha Mehra (00:09:50)  
and research for product development. And since 2023 I've been ⁓ working with Towards AI on ⁓ mostly technical writing and also some consulting on the side.

And who is Two R ZI? Two RZAI has been a publication on Medium. If you know about Medium, it's a place where people publish articles. ⁓ And basically, ToRZI was a publication there where we have thousands of writers explaining AI concepts. It has been there since 2019 as well. But now we are transforming our company ⁓ not into a media company only, but also an education company with.

Our 2RZI Academy ⁓ online learning platform, but also industry coaching trainings and consulting ⁓ trainings like this one, for example, that we also give to companies directly in other styles and formats where we adapt the course to the company. ⁓ We also have tons of free resources like the 2RZI blog post that I just mentioned, ⁓ my YouTube channel, the Whatsi YouTube channel, a few weekly newsletters, including the bigger one.

That is now followed by over a hundred thousand people. And as a fun fact, ⁓ one of the followers is Jensen Wang, the CEO of Nvidia, which was quite funny to learn recently when we were looking just for fun to ⁓ company emails. So that's ⁓ that's quite cool to know that Jensen frequently reads our newsletter. ⁓ We also have the LearnAI Together Discord community with ⁓ now over 75,000 learners. So

That's always pretty cool to go to explain. So, what is this course exactly? And it's five sessions of two hours, where each session, as I said, consists of one hour of like the essential theory to know given by me. And the second half ⁓ is ⁓ some applied theory that we call with some code examples ⁓ given by Omar. ⁓ By the way, if you have any questions, feedback, or whatever, please ⁓ feel free to email me with the email.

Sneha Mehra (00:11:59)  
Right here, ⁓ Louis at towards the i.net. We are happy to help you out, build courses, give coaching. We we take on many projects. And what's this course? It's five sessions to cover the fundamentals. The session one is this one, where we will cover foundational knowledge and how to use LLMs. So it's just a quick, I guess, recap or summary of how LLMs work, why they are so good.

Their limitations, their weaknesses, some prompting tips, coding assistance, security, etc. A lot of good basis to know when first starting. Then session two is about building on top of LLMs. We discussed RAG and the whole pipeline one company ⁓ usually has to follow to build around LLMs efficiently. Then we go out about evaluations in a single course.

covering everything you need to know about evaluations, which is, I think, or we think the most crucial course of this whole series. Then we cover the one I think people love the most: workflows and agents with case studies, where we cover when to use LLMs, when to use workflows, how to build agents, and lots of stuff in this ⁓ in this style. And in session five is kind of everything else you need to know.

more about optimizations, monitoring, and advanced techniques, the things to know afterwards that ⁓ are still somewhat crucial, but ⁓ that didn't fit into the session one-to-four. So ⁓ we talk about optimizations that we may not necessarily use as an industry worker, but are still very relevant to understand and you might encounter this in the near future. ⁓ we'll also talk about vile coding, how to stay up to date, and a lot of other.

Useful topics. So when are we giving this course? Well, it's today, you are following it already. And today's plans is the, as I said, AIPRE, a quick review. We'll discuss the limitations, different ways to interact with LLMs, some tips for prompting, coding assistance, when to choose which LLM, etc. So it's it's it's very much the the foundation ⁓ for starting.

Sneha Mehra (00:14:18)  
If you are already well aware of that, you can skip to the lesson two, which I think is extremely relevant for ⁓ pretty much any ⁓ future AI engineer. In session two, we take you in the path ⁓ of someone that is building a product using an LLM. Not using, but implementing an LLM in their application. And so ⁓ we cover all the steps that they should be trying in order. ⁓ And I think it's very relevant and very interesting.

Before we start, I just want to mention that there's ⁓ tons of jargon. Note that the slide is maybe not up to date here because you cannot really interrupt. But ⁓ if something's unclear, let me know by email if you want, or you can just ask ChatGPT. ⁓ I may often use different words for the same thing. For example, when we say that ⁓ ChatGPT has one trillion parameters.

this is often referred at as parameters, as I said, but also as weights, neurons, ⁓ and other terms. So there are many terms that are interchangeable a bit. But I will try to make it as easy to follow as possible. So, some theory for understanding large language models without map or without code. So LLMs fit in the AI landscape ⁓ and inside machine learning. They learn to do things and are not programmed to do so.

just like expert systems were where we experience we basically had a bunch of if cases. So if something happens, do this or else do this, etc. And just a bunch of them until we reach a decision. And this is all decided by the programmer or like the expert ⁓ in the domain that tells the programmer what to do. Here instead we use neural networks.

something ⁓ vaguely inspired by the brain to try to learn for us what to do in these kind of if cases. So it's basically ⁓ replicating our expert. That's the whole goal ⁓ at all times. It's also now leading the NLP field, but also unifies vision, speech, code, and everything. It's the first model, the first type of model, the the LLMs, the transformers that ⁓ works for.

Sneha Mehra (00:16:41)  
Many modalities and super good well. ⁓ Here a modality is just like vision is a modality, speech is one other, text is another, etc. So it works with a lot of them. ⁓ LLMs also use the deep learning architectures. So here deep ⁓ mainly refers to the neural networks type that have ⁓ multiple layers. So I won't enter into the details here specifically because it's not that.

Relevant, but basically we have our input, which is basically what we send the model, our prompt, which will be here decomposed into each of these words. For example, here in red we will have three words. Then we have our little arrows that process the information to the next layer. Without entering into the details, the technical details, here we are simply doing math, mathematical products ⁓ of

A weight, so ⁓ a number associated with each of these arrow ⁓ to the next neuron. So, for example, to the next layer, ⁓ each of these ⁓ bubble would be called neuron here. So, for example, ⁓ the first top left ⁓ red ⁓ input bubble would be the first word like I, and then it will be sent to all the next ⁓ bubbles in the the hidden layer one in blue, just to tell.

Each of these, how much the ⁓ I ⁓ word is relevant. And then this process is repeated again. And it's just basically numbers multiplying numbers again and again. And so each of these lines that we see here ⁓ is what we call a parameter or a weight. It's just a number, ⁓ between usually between minus one and one or zero and one that multiplies the previous numbers to.

Process the information in different ways. ⁓ It's not really important to understand the full details and the full mathematical concepts and everything, but it's important to know that it only works with numbers and it's basically just scalar numbers ⁓ between minus one and one that decides of everything with multiplications and additions. So it all is pretty simple and just does multiplications.

Sneha Mehra (00:19:09)  
In parallel for for each layer, and then layer by layer, ⁓ one at a time, sequentially. So it does many calculations in parallel, so you need lots of memory, and it also does them in sequence, so it takes time. So that's that's the deep in the deep learning. It's because we have multiple of these layers, hidden layers here, until we have our output layer, which is basically just the prediction of the next word.

Or the next word to generate, the next word to say. ⁓ LLNs also generate stuff, but they also understand it pretty well, obviously. This is why it's called generative AI. It's because it's not the first, but ⁓ the first type of models to really be good at generating stuff, versus previously, where AI ⁓ was super good at predicting stuff. In both cases, ⁓ predictive or generative, they learn to represent data. ⁓

a distribution of data and understand it quote to quote. Just ⁓ trying to do predict ⁓ what's the the the shape of the data, etc. With the goal of yeah, first predicting it, but also ⁓ to create new data similar to the ones that we have. ⁓ for example, the the predictive ⁓ type of AI is like image classification or sentiment analysis. And the generative example, well.

You know it already by generating text, generating images, etc. There are four core components behind large language models that really make them work. And we'll cover them one by one right now. ⁓ The first one is data. It needs two types of data. The first one is basically everything you can find. So it's the whole internet that per people usually usually refer to, but it's Wikipedia articles, books, news.

Lots of code, etc. And the second type is something that we that we call fine-tuning, but we will dive in that soon. But it's basically curated examples of what we want our model to follow. And this is what is expensive to ⁓ produce. This type of data is what is expensive to produce. The second core component is are the tokens. So basically it's our input, which isn't text, because models

Sneha Mehra (00:21:30)  
Don't know text really, they just know numbers. And those numbers are called tokens. We basically just have like an English dictionary, for example, ⁓ and each word has ⁓ an index where it is in the dictionary, and that's that's it. That's what the models see. It just sees these numbers that don't mean anything. So they need to understand it. And to understand it, they use embeddings. So these embeddings take the words.

You make the token. So for example, here if you follow cat, the cat is a token 9059\. And then it produces an embedding. Here, each embedding is just three numbers. Like for example, cat it's 0.256, 0.121, minus 0.552. That's not really important, but it just shows that there are multiple numbers that represent now each word.

And why there are multiple numbers is because here there are just three numbers, but in reality it's thousands of them. And each of those numbers represents a characteristics that the model learned by itself to represent our world. So for example, the the the first number could be is it big or small? The second number could be is it alive or not? The third number could be is it

it its color, like from from zero to one, the whole s spectrum of colors, etcetera. So it's all characteristics that ⁓ ideally, when well trained, the embeddings will accurately be able to represent anything in the world. And if ⁓ we plot it just in two dimensional here for example, or three-dimensional, we see that similar topics are near each other, like cat and kitten, but dissimilar topics are far away from each other.

Like cat is far away from garden or from the end and width. The embeddings is basically the language of the model. So it's how they learn to represent our world. And so by using and understanding embeddings, we better understand LLMs and we can leverage embeddings to do other things, as we will see. ⁓ Anyways, it's just to say that embeddings are super essential here because it's just we are not even inside the model right now. ⁓

Sneha Mehra (00:23:55)  
It's we are building its its future language. So ⁓ it's basically its whole understanding of our universe. So it needs to be pretty damn good. And then it will process it to understand multiple embeddings together. So ⁓ yeah, that's just to show how useful and important are embeddings. Then we have a deep neural network. Here it's always called the transformer. ⁓ we always use this model as an LM, but there are

Many others that exist. So what we've seen was the we have the text, we send it to our tokenizer to make our token, then we transform it into our embeddings so that now it's ready for the model to take it. And our model is ⁓ a series of ⁓ decoder blocks that we call, which ⁓ includes two crucial steps. First is the attention step and then the feedforward step. ⁓

For example, we can have ⁓ 20 or more of this this block that are that repeats both these steps. ⁓ And as a concrete example, DeepSeq that was somewhat recently released has 61 of these blocks. And ⁓ yeah, we will dive into those two crucial blocks because ⁓ they are what makes transformers what they are. So for the first one, the attention mechanism. The attention mechanism is crucial.

For the model to understand how each token should use each other token to better understand its context and where it is. For example, here we have ⁓ the sentence he sat on the bank and watched the river flow. ⁓ And here, if we send just the word bank to the model, it might think that we are referring to a financial bank, but it's not the case, just based on the sentence. So that's the role of the attention mechanism: ⁓

Put the sentence with the word itself together to have a global understanding of what's going on and be able to generate the next words by understanding all the others and their interrelations. Next is the feedforward layer. This is also extremely crucial because it's kind of the brain of the model. It's basically a tiny neural network that we described before that will take each token and kind of ⁓ process it.

Sneha Mehra (00:26:18)  
To make a new version of this token. So it kind of refines the token ⁓ the deeper we go into the model. And here, ⁓ just remember that each token has been transformed into an embedding. So our embedding at first is, for example, ⁓ when we add the word bank, it would be the embedding of a bank. And bank can mean many things. So why do we want to transform it? Well, it's because we want to better understand that.

It's actually the bank of a river, not the financial bank. So we need those feedforward layers that will transform each word representation, so the word bank here, into what it actually means in its context. This specific layer is pretty much the brain of the network and it will be repeated with the attention mechanism to always complement understanding the sentence, the global context to improving the current understanding of each local aspect.

And so we repeat that many times to finally obtain our prediction, which would be again a token, so our index. ⁓ And then we find the word associated with it, within this case would be in, ⁓ which is our output. This is what we call our output. And then what happens? ⁓ Now we have ⁓ for example, if we read the sentence ⁓ below to the right, is a long time ago, in. ⁓

So we send this whole sentence again to the tokenizer embedder. Then we process in ⁓ the transformer with the mini decoders, and then it generates again a second token. So we repeat the whole process for each new token that we want to generate. And this is why we say that transformers are auto-regressive. It's because they need to generate a token to generate the next one. So they cannot just generate a sentence one shot.

They need generate one word at a time and use the generated words to generate the next ones. So that's why it's a lot of compute, a lot of time, and why you receive ⁓ token in a streaming manner that we say, ⁓ where you see one word appearing at a time as if it was speaking to you, but it's because it's generating one word at a time. ⁓ And ⁓ how do we get these models? Well, it's by training them ⁓ on the data that we mentioned.

Sneha Mehra (00:28:46)  
And there are two main ways of training them that we'll cover right now. The first is pre-training. So we basically want to learn them how to write, how to understand the language, or like English in this case, and how to predict next words or ⁓ the next sentences and next reply to say. ⁓ And as a consequence of how it's made with

Basically, the whole internet, lots of Wikipedia, etc., ⁓ we are basically teaching it a lossy compression of the internet. Why lossy? Because the while the models are enormous, ⁓ the internet and the data we have is way more huge, and we cannot memorize everything. It has to compress a bit. But this isn't enough. ⁓ just training on the internet isn't enough. Why?

Because ⁓ if we show here an example with GPT-3 that was just a model pre-trained like this, when we say the prompt write a short poem about a wise frog, the model just give more instructions instead of replying or ⁓ doing what we want it to do. And that's because in its training data, when it saw this type of sentence of writing something, it saw probably lists of

In such instructions. I don't know why, but that can happen. And so the model is just doing completion based on what it f it has seen on internet. So it's basically an internet simulator where ⁓ for example, if for with GPD3 you gave the a Wikipedia article and you gave it like half of it, it would normally generate the rest of the article almost as is.

And that's because it has been trained on the internet and on these articles specifically, and not to ⁓ understand what to do and how to answer. ⁓ So this is where post-training comes into play. And here there are two things, main things that we do in post-training. The first one is called fine-tuning, where we use this data that we mentioned ⁓ that is basically just a lot of examples of questions or queries pronounced by the by users and ideal responses we want to have.

Sneha Mehra (00:31:07)  
This is used to refine our models for our needs or to follow ⁓ what the user ⁓ wants. It's the same thing as training, but we start with a pre-trained model rather than a ⁓ from scratch. ⁓ We start with a very powerful model that already understands English and some things to do, and we just retrain it to make it much better at following what we want to do.

And also, as I mentioned, it needs a curated dataset. So now, if we show the example, ⁓ the the version fine-tuned of GPT 3 is called instruct GPT, and now it gives ⁓ the short poem, the real one, instead of just a list ⁓ of other instructions. ⁓ There's also the the a relatively new ⁓ form things that we do.

That is called reinforcement learning from human feedback or for from AI feedback, as with ⁓ DeepSeek and Entropic ⁓ has been doing, which is mainly used to align the model. How? ⁓ We do two main steps. ⁓ One is to ask many questions to the model and ask it to answer multiple times. So we ask it to answer, for example, here ⁓ we have ABCD, so four times.

And then we pay people ⁓ or we use models, ⁓ language models, to rate these responses from worst to best. ⁓ And ⁓ once we have this whole data set dataset built, we use another model and we train another model ⁓ that we call our reward model to give this rating. So we basically train a model that can ⁓ say if a response is what we want or not.

And then we can use this new reward model to train ⁓ our existing model with ⁓ reinforcement learning and algorithms that we won't discuss in this course, but ⁓ reinforcement learning as algorithms like PPO and ⁓ newer methods. And for example, it's it's no longer true, but ChatGPT used to always answer like this a short introduction, bullet points, a short conclusion.

Sneha Mehra (00:33:27)  
Now it uses lots of tables, emojis and other stuff, but basically that's what reinforcement learning did. It changes the way the model answers and gives ⁓ responses and is mostly used ⁓ for better UX. ⁓ It's also how many companies like OpenAI make the model learn to refuse malicious queries. It just gives lots of examples of when ⁓ it cannot reply and how it should reply instead.

And then we have a final part after the pre-training and post-training, we have the inference where we serve the model to users. Here we have many important components ⁓ to make the model better for users, like streaming the tokens, as we say. So since we generate one token at a time, why not show it one at a time when it's ready so that the user already feels like it's getting a reply instead of waiting for the whole generation?

there are tons of chat features as we see in ChatGPT, like the history, memory, etc. ⁓ there are multiple types of prompts that we can ⁓ use and provide the model, like the system prompt, the assistant, the user prompt, that we will all dive into in this course. And basically, these prompts, ⁓ what we send, ⁓ for example, the OpenAI API, are dictionaries, so ⁓ a list of dictionaries if we have multiple interactions with ⁓

The current role, whether it is ⁓ assistant or user, whatever, and the content, so just the prompt itself. Then ⁓ at inference, we need to manage quality, cost, efficiency, small versus large models. We need ⁓ also to augment LLM, as I said, with ⁓ memory or other tools, ⁓ code interpreters, calculator, etc. So that was a lot. So let's recap what's going on. We first start with a general no-ledge LLM.

Something that has learned everything on the internet. And ⁓ we have a tokenizer, which is, by the way, specific to each model. So that's pretty important. And so we are now able to take text, prompts, transform it into numbers, process it into our model, and generate new text. But this isn't ideal. So we need to fine-tune it to get our instruction model. So something that can follow instructions. Then we can, if you want.

Sneha Mehra (00:35:52)  
Train it further to better align it with our needs and add things like reasoning, for example, that we'll cover later in this course. ⁓ We can also ⁓ do as Meta is doing with Lama, so destinying models, bigger models, into smaller ones, ⁓ which we will cover how it works pretty soon. ⁓ That's used to retain knowledge and increase efficiency. And finally, we serve it to user, where each new query is tokenized.

Embedded and processed to generate a new token, and that is repeated sequentially, ⁓ which means autoregressively, until the whole sentence is generated, the whole answer is generated. So the model generates new tokens sequentially, ⁓ as I said, which means it simply generates one token based on the previous sent and its whole training data. It also means an LLM is a statistically powerful machine.

It's not conscious and not intelligent. They are probabilistic, not deterministic, and and so answers may vary even if you send the same prompt. It also remembers most facts if seen often in its training data, but it doesn't have memory or real memory by default of what you say it and ⁓ your all your exchanges. This has to be programmed into it with many tricks that we will also cover in this course.

It also cannot tell the difference between truth and a lie, which can be a huge problem and limitation, which we will address in this course. And I put here asterisks everywhere just because that's mostly for now. ⁓ Models are evolving incredibly fast. They are always better and better. ⁓ And ⁓ just for example, reasoning models came out a few months ago and it's it improved models drastically. And so, anyways.

It's always improving pretty quickly. So those limitations are just for now. By the way, it's also all these reasons, the fact that they are not intelligent and they are just processing numbers. This is also why models are bad at jokes. Since they generate word by word, even if they have the the context in mind of what to say next, they mostly work on the next word, the next token. So they are pretty bad at building suspense, anticipation, and creating a punch.

Sneha Mehra (00:38:20)  
Even if they can do that, ⁓ they are not the best ones, mostly because they don't plan much ahead. But that has been fixed with many workarounds like the reasoning that we see, ⁓ or ⁓ implementing a memory directly, or like some planning steps. And so, this ⁓ all leads to our limitations. Since models are so powerful, ⁓ one of the problems is that they are extremely powerful.

Calculators, they process numbers, crunch numbers, understand numbers, they are not conscious, not intelligent. So ⁓ that's one of the main limitation or weakness of LLMs is the what we call the hallucinations. That happens when ⁓ the model produces confident yet factually incorrect outputs because of the way they work with next token prediction, which is ⁓ statistics. We had the example of the number of R's in Strawberry.

It's now fixed, by the way. But if you use to ask how many R's are in Strawberry, the model used to answer two or a wrong number. But ⁓ there are many fixes, and this is all due, by the way, because ⁓ as we said, ⁓ LLMs work with tokens and embeddings. So they were representing for Strawberry as one or two tokens. And so it doesn't have the concept of letters. It cannot really count letters in each word. But

There are fixes ⁓ like asking to do it step by step or to spell the letters beforehand, etc., or use tools like ⁓ a code interpreter, where you just ask it to verify with code and it will ⁓ spell the word and count the errors. There are also biases that all come during the training. ⁓ it basically reflects all the stereotypes that are present in your training data.

So every bias on the internet. ⁓ For example, if you were to prompt the to list typical professions for men and women, well, you would have the classic ⁓ men stereotype jobs and women stereotype jobs. ⁓ there's no real fix to that. You just want the best data possible ⁓ and basically mitigate the biases that you don't want ⁓ and ⁓ focus on the biases that you want.

Sneha Mehra (00:40:47)  
Since every human, everything is biased, you can just try to have the best biases possible, like being nice, respect others, etc. Those are all biases that you may want to have ⁓ in your application. And to do that, you need good data and you need to pay people to produce them. Another limitation is the knowledge cutoff, which is basically the last training data point ⁓ beyond which a model is no longer updated or trained.

So when we do the pre-training and post-training, as we said, we ⁓ do it once or multiple times, but we do it. ⁓ And once it's done, there's no input anymore into the model. It's not modifying its weights, its parameters, as we said. It's just fixed there in time. A quick fix is to give it access to more data, like the internet, for example. ⁓ likewise, the knowledge that you ⁓

Need to use may not have been in the training data of DLM, like speaking another language or trying to have answers in another programming language. Well, if the model hasn't seen it during training, it might not be able to answer pretty well. And to fix this, you either need to retrain your model or fine-tune it, what we call fine-tuning, or use ⁓ what we call rag. So add ⁓ retrieval to your system with your own data.

Another limitation is the limited context window of models. The context window is actually the number ⁓ of tokens that you send to the model. So we have a limit of the maximum number of tokens that you can send at once to the model, which is usually in the hundred of thousand of tokens ⁓ up to 1 million or 2 million in the case of Gemini. Just to note that larger windows do offer more context, but are costlier, slower, and ⁓ they are not perfect.

For example, Gemini or Google with Gemini uses Infini Attention. And it basically works like usual, processing a hundred thousand or so tokens at a time, but then at each step it condenses information to send it to the next one. So we lose this information just like zipping a file or ⁓ compressing an image to make it a bit blurry. ⁓ We save space and we can send more information, but it's of lesser quality.

Sneha Mehra (00:43:10)  
And so the fix here, ⁓ the problem here rather, ⁓ is called the needle in a haystack, where ⁓ this is a type of ⁓ benchmark that we use to measure how well the model does with long context. But the problem is that needle, that this needle in the the haystack is basically like an answer that you are trying to find in the whole book. Well, this needle appears quite flashy in the the context. It's not like a subtle thing you are looking for.

It really flashes and the model can find it more easily. And another issue is that there aren't multiple related needles ⁓ that we can evaluate on to understand if the ⁓ long context actually works. So, for example, if we send a book and ask everything to know about Robert, models are pretty bad at that. Just because there are things to know about Robert at page one, page 10, page 500, ⁓ and they cannot really do that pretty well. And to do that,

We instead need ⁓ to divide and concure. So we need ⁓ what we call rag that we will cover in the full second course. Another limitation is that ⁓ LLMs ⁓ are not intelligent. They struggle with basic logic, like for example, counting the letters R in Strawberry. ⁓ Or here, very basic example: ⁓ asking it to add 1 plus 0.9, then it says the right answer. But if you say

⁓ isn't it 1.8? It will agree and say that it's 1.8. So it doesn't has like an opinion and truth. It just follows what you say. So that can be pretty dangerous, by the way. Another example to show that it's a bit stupid or or not conscious and intelligent is ⁓ when you ask which is bigger, 9.11 or 9.9, it says that 9.11 is bigger because basically 11 is bigger than 9\. So that's

Totally false. And it's because of multiple reasons, but most importantly that the model is overfitting to memorized examples rather than doing true reasoning. And this is fixed with what we call reasoning or ⁓ using external tools and obviously improving models. Just as a quick example of how someone can fix those limitations, ⁓ I have here how OpenAI or ChatGPT addresses these limitations.

Sneha Mehra (00:45:39)  
With ChatGBD. So first it's by having lots of curated dataset and then by paying ⁓ lots of humans to do RLHF or reinforcement learning from human feedback as well as possible. Then they integrate many tools like live web search, they integrate ⁓ RAG or retrieval augmented generation, ⁓ which is basically document ingestion. They allow for custom fine-tuning for domain-specific expertise if they don't know your language, for example.

They are increasing token limits over time, which lots of companies do, and they have bright patches for other problems, like ⁓ for example, summarizing chat history to make it more condensed, but keeping the crucial information instead of ⁓ relying on the the math inside the model to do it. They ⁓ do advanced techniques like prompt caching, which is basically to ⁓ pre-compute the text that you send often. So as I said.

When you send a prompt, it's always tokenized and embedded. So here we will do it once. And if you send the same prompt, big prompt, like sending a PDF multiple times, it will do once and then reuse the computed embeddings ⁓ when you ask new questions. They implemented user memory for better user fit, which by the way can be turned off. ⁓ That can be interesting because in my case, it finally knew that I was.

A YouTuber, and it started, I don't know why, to always answer as if it was a YouTube video. So it they always said like, Good morning, this is my name, or whatever. It always had a style of a YouTube video, even if I just asked a random programming question or whatever, so it was pretty annoying. I turned it off and it was all fixed. So ⁓ if Chat GPT starts acting weirdly to you, that might be a nice fix.

They implemented many useful tools like the co-interpreter for better math. ⁓ They implemented Canvas, which I don't use anymore, but when it first came out, it was pretty interesting for code and text. I believe it may not be available for O3 and reasoning models, but I would need to confirm. But in any case, ⁓ this used to be interesting where I, for example, here ⁓ highlighted a word ⁓ asked.

Sneha Mehra (00:48:06)  
Do something here, it's in French, but ⁓ bear with me. I was just asking ⁓ to better explain JSON before mentioning it, and so it edited the only paragraph and ⁓ provided an explanation of what a JSON is ⁓ before referring to it. So it's pretty useful to edit only specific places in your text or code. ⁓ And finally, they are working on reasoning models that fix a lot of issues that we'll talk about in the

session five. And now quickly, ⁓ I want to cover the AI landscape just because I always hear like things like ⁓ the model chat GPT is super powerful or whatever. ⁓ I don't I just want to say that first ChatGPT isn't a model, it's an interface. And so this is why ⁓ this whole slide. So first we have companies in the field. So the main one would be OpenAI, but there's also Google, Enthropic, DeepSeek, and others.

Which ⁓ usually have two types of interface: the user interface, which ⁓ is the one that most people know about, ChatGPT, or Gemini, Cloud, DeepSeek, etc. ⁓ And ⁓ there's a developer interface like the playground or the AI Studio, the console, etc., where developers use APIs and communicate with models. And those models that are also in ChatGPT.

That are that are both in the user interface and developer interface ⁓ are the actual model. ⁓ For example, GPT 4.0, ⁓ 4.0 mini, 01, now it's like 03, ⁓ 04 mini. ⁓ We also have developer-and-models like the GPT 4.1 series that is super useful. But there's also, as I said, Gemini models, cloud models, deep sea models, etc. So this was the last slide from me. I'll now leave Omar to talk about the second part of this first session.

Alright, so for this second part of the presentation, ⁓ we will start with the different ways to interact with LLMs, then some prompting tips you can use, some ⁓ tips around how to use LLMs as programming assistants, how we can choose LLMs depending on the task. ⁓ It's going to be ⁓ some high-level information which we will go more in depth in the next few classes.

Sneha Mehra (00:50:32)  
And finally, ⁓ some points around security and ⁓ confidentiality. So, for the five ways we to interact with these models, of course, we have the public-facing chatbots, so ⁓ services like ChatGPT, Cloud.ai, or Google Gemini, for example. They all have some sort of free tier where you can use a service for free, but your

Pretty limited in terms of how many times you can interact with a chatbot. So they all they also offer these subscription ⁓ models where you can pay and increase the number of interactions you can have with the ⁓ the services. Then if you're a developer, you can also use ⁓ these AI models with public API endpoints. So you have, for example, companies like

OpenAI, Google, or ⁓ or even third-party inference companies like Together AI or GROK, where you can basically use these models through APIs ⁓ and then instead of paying by subscription, you you pay by the number of input or output tokens. If you need a little bit more privacy, so this is better for enterprises that need more

more security and privacy for their use cases. Then you can ⁓ choose these private API endpoints. So you can, for example, use Azure OpenAI or even Amazon Bedrock to ⁓ use these models in a more ⁓ private way. So that's another ⁓ way you can ⁓ use these models. And of course you can also deploy your own models to different GPU

Cloud providers. So you can either go with Google Cloud, Amazon Web Services, Azure, or even smaller companies like Lambda Labs, where you can deploy open models. ⁓ And since we have the access to the weights, the model weights, we can deploy the models. And then we can, for example, with this way to use LLMs, we're going to pay by mostly by the hour. So we are going to be renting these GPUs.

Sneha Mehra (00:52:55)  
And here I give ⁓ a small example. So for example, if we use the full precision Lemma 70B, then we are going to need around 80 gigabytes of VRAM, and that's about around $2 an hour on Lambda Labs. So that's another way we can interact with these models. Then of course, the other way we can use the models is to simply deploy them locally.

if you have the ⁓ the the hardware to support it mostly it's not very cost effective to do that unless you already have a product and you have a lot of users that need the where you can ⁓ serve those those users and in this case for the same model so Lama 70B you could buy two A100 cards

With 40 gigabytes each, and it's going to cost you around 23k. So it's not it's not ⁓ cheap. Now, if we talk about prompting, of course, ⁓ you have the simplest technique, which is just zero shot prompting, so just an instruction. Here, for example, I say translate this statement into French, and the chatbot, the the AI model is going to be able to translate the string.

So, this is the simplest way you can ⁓ basically prompt these models. Then, ⁓ more performant ways to use them is to basically ⁓ do what we call few shot prompting, where we add some more instructions, maybe some relevant context of the task, ⁓ and examples of the task being done within the input prompt. So in this case.

We can see that following the simple instruction here, we add more examples. ⁓ And we also add the correct ⁓ output format that we we want the answer to be in. So this is very useful when you have a well-defined task. ⁓ And yeah, basically we always recommend to add more instructions and to well define your tasks whenever you're prompting these AI models.

Sneha Mehra (00:55:19)  
Then you have role prompting. This can be useful in some cases. You're telling the model to react or to write in in the form of a of a persona. So for example, here we we can tell it to act like a like a pirate. And so here we don't need to tell it exactly what ⁓ a pirate does because it already knows what a pirate is. So the model can already ⁓ act in the correct way.

So, for example, you can also tell it to be an expert in marketing. ⁓ And instead of trying to write all the possible ⁓ correct instructions, if the model already knows what a good marketer does, then the model is going to be able to ⁓ write correct marketing copies, for example. So this can be useful in many different ways. Then you have chain of thought prompting.

This is more useful for models that are non-reasoning models. So we will explain what a reasoning model is in the next few presentations. But basically, here we want to give these non-reasoning models the instruction to reason about how to find the answer before answering it. So, for example, if we tell it to do ⁓ a math computation, we tell it to write step by step.

What will be the different steps to solve the problem, and then we will tell it to answer ⁓ at the end. This is very useful because ⁓ it gives the model more tokens, more time to find ⁓ the correct answer. It's the same you you can think of it like humans, where we need to write down the different steps to get to the correct answer. So maybe we can do it ⁓ first shot, but it's

The task becomes way easier when we simply write down the different steps and ⁓ reason ⁓ with the steps to find the correct answer. So whenever you're ⁓ prompting these models, we recommend that you always tell these models to think, ⁓ to write down the steps before they give the correct answer. Then the fifth ⁓ and last prompting tip we can give you is to use advanced frameworks to

Sneha Mehra (00:57:42)  
⁓ basically automatically optimize the prompts that you give to different AI models. Here you can use tools like DSPI ⁓ or Adult Flow to do it. It's a little bit more complex to set up. You need, of course, to code the optimization flow and also to provide a way to evaluate the prompt. So an evaluation function. But using this, you can ⁓

Basically, make sure that you're optimizing to the limit whatever task you're you're doing. So we we recommend that you check out these tools. So for some key principles, we always want to treat prompts like code. So you might want to use version control to ⁓ optimize your the prompts you give to different AI models.

By doing that, you're going to be able to ⁓ iterate on them. It's going to be easier to roll back ⁓ or optimize in a more scientific way, let's say, the input prompts. You can ⁓ try different tools to do this. You can either use more, ⁓ let's say, ⁓ established services like Microsoft Prompt Flow, or you can use simpler.

Tools like Prompt Layer or even ⁓ LangSmit to do these kind of things. Then ⁓ of course it's very useful ⁓ to ⁓ write ⁓ w whatever instruction you're giving to these models, to write them in the in a way that is very clear ⁓ and very concise. So ⁓ the models don't deal with ambiguities. ⁓ Here a good communication is your is always going to be good, basically.

And also, whenever you're ⁓ working on writing a prompt, to not forget to add all the necessary context. So ⁓ you can think of it like giving ⁓ trying to get help from a coworker, ⁓ and you give it as much context as possible ⁓ to be able to be helped and so that they don't get stuck or don't really understand the task you you want to complete. So here it's going to be very useful to add.

Sneha Mehra (01:00:08)  
All the context that they need. And also, of course, to be explicit, to ⁓ avoid implicit assumptions. So these models, maybe even if they're they they seem very intelligent, sometimes they can make ⁓ wrong assumptions. ⁓ And if you're ⁓ an expert in a field, then then it's going to be very useful to be explicit about them ⁓ and guide the model as close as possible to.

To extract ⁓ the most performance you can from these ⁓ LLMs. Alright, so now let's talk about how we can use these LLMs as coding assistants. But before we do that, I thought it would be interesting to see exactly ⁓ how programmers think about these tools and how they are using them since ⁓ they basically came out. So here we have five insights.

From a survey that GitHub did back in September 2024\. And for the first insight, we can see that for ⁓ 97% of them, so basically all of them, they have already used AI coding tools. So we know we ⁓ we basically see that AI tools are no longer a niche thing. It's now something mainstream among developers.

For the second insight is that in terms of how they boost productivity or creativity, ⁓ here we see that for American developers, they basically now spend more time collaborating with their team members ⁓ or ⁓ designing better systems for their product instead of spending time, for example, doing code reviews or even taking breaks.

And for the other countries, we see that it enables developers to spend more time learning ⁓ or even researching and experimenting with new technologies. So AI enables developers to ⁓ basically make it very easy to adopt new programming languages or trying to understand new code bases. And that's what we see here. So for

Sneha Mehra (01:02:36)  
The majority of all the developers in in the USA or even in Brazil and the others, we can see that developers find it very easy or even easier to understand new code bases. So that's how ⁓ AI is helping helping them in their work. Then for security and quality, for 90% of them, they think that it enables

them to write better secure code ⁓ and that ⁓ yeah basically it's gonna it's going to be better for code security. Then ⁓ in terms of adoption we can see that the American developers are more encouraged by their companies to ⁓ use this new technology in comparison to ⁓ countries like Germany where developers are a little bit less encouraged to use them.

So we we see here a gap between these countries. Very interesting. ⁓ And ⁓ the ⁓ final insight from the survey is that for basically 100% of them, they believe that using these AI tools will make them more employable over time. So very important to start ⁓ using them right now, ⁓ so we can be more

Yeah, basically employable, better performing using these tools. ⁓ Now, for the first ⁓ strategy or method that we can employ to get better performance out of these tools, of course, is just to use the automatic code suggestions tools that we can use. So basically, many services already provide them. So you have GitHub Copilot.

You have CursorTab ⁓ and Windsurf and many, many other tools. They basically allow you to ⁓ accept automatic suggestions ⁓ and it ⁓ it it can make it faster to write code. So we we always ⁓ suggest that you turn on this feature and it will ⁓ make it a little bit faster when programming, basically. The second way, which is not automatic.

Sneha Mehra (01:05:01)  
Is to basically prompt directly the model to help you write ⁓ a function or whatever you're working with. So here instead of getting the automatic suggestion, we write a prompt, a question, whatever you want ⁓ to see, and the model will try to answer the best it can. So here for ⁓ what model to use.

Here we recommend that you check out benchmarks or even leaderboards. So here I put the example of the Web Dev Arena leaderboard. ⁓ I'm going to come back to this later. But basically, whenever new models come out, sometimes new models are way better for programming, so you can try to experiment to use them ⁓ and then get ⁓ better performance out of these tools.

so yeah we recommend that you check out the ⁓ web dev arena leaderboard for for that specifically. Whenever you're chatting with these models, it's very useful to ⁓ basically ⁓ understand different codebases. So if you're trying to understand a new codebase here, the cursor chat with codebase will make it very, very easy to ⁓ get up up to speed basically.

So then here, another way to use ⁓ these models more specifically is to use them as code and documentation explanation tools. So here, tools like cursor with the chat with code base feature is very very useful for that. The tool will try to retrieve all the relevant context. ⁓ And ⁓ then the model will have all the context to answer your question basically. So

Whenever you're trying to understand a new code base or you're trying to understand the ⁓ documentation written by your coworkers, here ⁓ with ⁓ with AI models, it becomes really easier to to do that basically. Then of course, you can use them to write ⁓ documentation. So here it's very important that you you actually understand how the the code functions.

Sneha Mehra (01:07:27)  
So you can fix whenever ⁓ a model didn't write the documentation correctly. But ⁓ using these models as first draft ⁓ or ⁓ even to just start writing the documentation here, it becomes very useful to do that. So ⁓ you can write markdown, whatever you want to write. Of course, you can use them for refactoring or even improving.

⁓ The ⁓ the code you're writing. So here you can ask it to provide different ways to go about implementing the feature. You can ask it to make a class or a function more concise. Here it's very important that you still read the code, whatever the model is writing, to really try to understand what is being written so you can still modify the code and understand it.

And of course, you can use more ⁓ agentic tools. ⁓ So ⁓ we will go more in depth in the third session on how agents work. ⁓ agents are some something more recent that that you can use. So here I list some of them, but basically, these tools allowed you ⁓ to give, to hand off some task, some implement implementation to

These agentic tools, and they will try to these agents will try to implement the feature the best they can basically. But we need to keep in mind that since they they are writing more code and you're not actually ⁓ giving them more guidelines or instructions, there's more probability of errors. But they're getting better over time. ⁓

Even if they're not perfect now, we can ⁓ expect better performance ⁓ in the following months, following years, basically. So if you're trying to ⁓ use them, ⁓ then we recommend that you still prompt them in ⁓ in a way that ⁓ reduces the amount of errors. So, first you're going to ask for some planning. So before they write any code.

Sneha Mehra (01:09:49)  
To ask them to come up with a plan, then ⁓ they can start implementing the feature. Then you're going to read whatever they wrote to validate that ⁓ the code looks correct and the feature is ⁓ looks well implemented, and then you're going to execute the code to see if there's any errors in the generated code. ⁓

These kinds of tasks that require require more autonomy and less human input. ⁓ You can check these kind of benchmarks. So you have Suite Bench, for example, that you can check out. Basically, ⁓ they're trying to solve the agents are trying to solve GitHub issues. And here you you can see that the first system solves about 64% of all of those issues.

So we can see that it's not perfect, but it's it's getting better over time. Now, some best practices to to ⁓ to make it clear. We want to prompt them with clear requirements and iterate and refine the prompts. And of course, we want to read, understand, and validate whatever the models generate here because we don't want to.

End up with a code base we don't understand that then we cannot debug. So here, of course, human expertise is still valuable. ⁓ and yeah, what we want to avoid, we want to avoid generating large code bases that we don't understand. ⁓ yeah, basically, even if the tools are very eager to provide a lot of code and implement implement things right away, ⁓ yeah, it

We always recommend to not do that, to try to generate as ⁓ small chunks of code to read it, understand it, understand the code, and then keep going. So now if we switch to ⁓ another topic ⁓ on basically what's different between open or closed source models ⁓ and going ⁓ about it by seeing different metrics. ⁓ yeah, first we can

Sneha Mehra (01:12:09)  
Categorize these two types by three different ⁓ categories, basically. So we have the proprietary models. So all the models you cannot deploy yourself, you don't have access yourself to the model weights. So ⁓ all the models by OpenAI, for example, or Anthropic or XAI. Then you have companies that are willing to share the weights of those models to everyone.

So here we have companies like Meta, Lama, Mistral, DeepSeek that ⁓ provide all the model weights. And then of course you have the models that are completely open source. So we're going to go point by point now. Okay, so let's switch to another topic now. Now we're going to see basically what's different between open and closed source models by looking at

different metrics and categories. ⁓ So we can categorize these models by three diff different types. So you have of course the proprietary models, the open weight models and the completely open source ones. So we're we're going to go one by one.

So for the proprietary, here these models you can only use them with paid APIs or web interfaces. So you cannot host the model yourself. These services ⁓ are usually not completely private. So for example, if you use the ⁓ the models of OpenAI through the API, ⁓ then OpenAI is going to keep all the data for 30 days just to monitor for abuses.

And if you want to use ⁓ proprietary models but in a more private way, then you can consider the private API endpoints that I mentioned at the beginning of the presentation. So you can check out Azure OpenAI or Amazon Bedrock. ⁓ And for ⁓ people using ⁓ services like ChatGPT, then ⁓ I just want to mention that you can opt out.

Sneha Mehra (01:14:28)  
Of letting OpenAI use your your converse conversations for model training. So if you look into the settings, you're going to find this toggle, improve the model for everyone. And you can either check it or uncheck it if you want OpenAI to use ⁓ your conversation basically. Now for the open weights ⁓ category, ⁓ here are the models that you basically can host yourself.

By hosting them yourself, you can ensure complete privacy, of course. But here it's important to note that ⁓ model replication, here I mean creating ⁓ creating a replica of the model is not possible because these companies do not share the training data that they used ⁓ and the training recipe that they they basically use to create the model.

For these kinds of models, you can either ⁓ host them yourself ⁓ on ⁓ GPU cloud providers, or you can use third-party LLM providers. Here I have a graphic showing all of the different companies that you can basically ⁓ use if whenever you want to use an open weight model. And if you want to ⁓ want to see more updated ⁓ figures, you can go to this website.

Artificial analysis.ai, which is which does a very good job benchmarking and comparing either LLMs ⁓ or these third-party LLM providers. So we it's ⁓ it's a very useful source. Now for the third category, ⁓ you have the completely open source models. Here you you have the training data, you have

A detailed technical, ⁓ basically a di ⁓ a detailed paper, a report. ⁓ And by reading the report using the training code and the training data, you could replicate the models. So create a clone or replica of the the model they released. ⁓ And this is mostly useful for academics or people trying to improve on these ⁓ AI models. So this is

Sneha Mehra (01:16:53)  
Very useful distinction between open weight models and open source models. Now, if we look at the models themselves, ⁓ we can look at different metrics. So we have the cost per token, we have the context length, the ⁓ model performance ⁓ on different key benchmarks, we have their features, their capabilities, so if they can process images, voice.

We also have the latency. So that's more about like the latency during inference. You can also look at if you are able to fine-tune them or not. And then ⁓ you can ⁓ try to look at ⁓ things like tokenization or even vocabulary to ⁓ differentiate these kinds of models. So if we look at the cost per token.

Here we can ⁓ I show a basic graph that shows how the cost per token has diminished over time. So ⁓ we can see that a year ago, two years ago, GPT-4.0 was quite expensive and now it's almost 240 times cheaper. So whenever you're trying to choose a model based on cost, it's always going to be something temporary because.

By seeing the the trend over time, we can see that these tokens are getting cheaper over time. So ⁓ this is something to keep in mind whenever you're either building or using these tools. So here I show another ⁓ figure where we can see that ⁓ actually the Gemini models are quite cheap relative to their intelligence. So for the most intelligent model, we can see that it's

Right now it's Gemini 2.5 Pro and it's quite cheap. So the X axis ⁓ shows how cheap ⁓ it is compared to two other models. So we want to be on the right here to be cheap and high ⁓ on the ⁓ y-axis to be smart, ⁓ and you can see basically that all the other Gemini models are cheap compared to.

Sneha Mehra (01:19:17)  
Their intelligence. So this is something that is very like snapshot in time. So whenever you're choosing something on a model in the moment, you can look up these kinds of figures and ⁓ make your ⁓ decision based on the on these ⁓ figures. Now, if we look at context length, we can see the the same thing basically. Back when ⁓ ChatGPT was launched, we only had

GPT 3.5 with 4K tokens of context length. Now with GPT 4.1 it can process 1 million tokens, and you also have Gemini 2.5 Pro with 1 million tokens of context capacity. So yeah, again, when we're building either applications or just using them as tools, we can ⁓ predict

That in a few months, in a few years, the context length will continue to increase. But we need to keep in mind that, ⁓ as we saw in the first part of the presentation with Lui, that the performance can actually decrease as the number of tokens ⁓ becomes large. So ⁓ even if the models can process a lot more tokens, we don't want to include irrelevant context, irrelevant information.

Because it could confuse the models, it could reduce the performance ⁓ of whatever task you're trying to get done. So that's something else to keep in mind. ⁓ Now, in terms of model performance, here you can look at different key benchmarks. So this is just trying to see. This is not about whatever feature you're building. This is just trying to see: okay, we have all of these models available.

Which ones are the ones that seem the most ⁓ smart, intelligent for ⁓ whatever you use case I'm trying to ⁓ to build. So here we recommend that you check out the LM chatbot arena. ⁓ So this is basically just a way ⁓ they they provide a way for people to vote ⁓ on what model provides the best answer. So you you give it an instruction.

Sneha Mehra (01:21:44)  
And then you get two answers, and you can choose which answer you you like the most without seeing what model is answering you. ⁓ And by ⁓ gathering all of this data but by a lot of people, then the people at ⁓ LM Arena they can compute this ELO score that you can then use to try to see: okay, here ⁓ for the web dev arena, we can see that most people right now ⁓ is.

prefers to use clot 3.5 sonnet. So yeah what whenever you're for example coding ⁓ and you want to know what model to use within ⁓ VS Code or cursor or anything like that, you can check out the web dev arena or things like that. Now for these benchmarks like MMLU Pro, these are questions ⁓ on

Related to ⁓ the STEM field. This is just to see how intelligent the model is in mostly scientific type of questions. ⁓ It might not be relevant for your use case, but it's always, I guess, good to see what type of model is most ⁓ intelligent in these kinds of questions. Then we have more

Specific benchmarks. So you can see, for example, the benchmarks provided by Scale AI, the company at San Francisco, where, for example, they measured things like agentic tool use ⁓ or coding more specifically. ⁓ And ⁓ these are questions that the the models never saw, so they you can be sure that the models were were not trained on those questions.

And yet, and then yes, you can use all of these kinds of benchmarks to see what are the ⁓ models that are most intelligent. ⁓ And here, for example, you can look at the things like the artificial analysis intelligence index, where they try to gather all of these different benchmarks and compute a single score. ⁓ And it can be useful ⁓ if you don't want to look at all of the individual benchmarks. So here you can have a single score.

Sneha Mehra (01:24:12)  
To ⁓ to compare all of the different models. Then, if you're ⁓ actually ⁓ building something related to images or voice, then ⁓ programming tasks or coding tasks might not be useful. So you can look at things like the vision arena, where people actually prompt the models with images, for example, and here you can see that.

Actually, Gemini here is the best model for that. So ⁓ it will depend on your task, but it's it's here it's ⁓ useful to compare ⁓ all of these models on different ⁓ benchmarks ⁓ or arena here in this case. Now, in terms of latency, here it's more mostly related about how fast you can get an answer from these models.

Here you can see that O3 is the slowest model to answer you because it it produces a lot of reasoning tokens. We will come back to reasoning models in the next presentation, but ⁓ it's ⁓ it's normal to see that ⁓ on the on the right. ⁓ And here, if we are building ⁓ applications that require require a lot of latency here.

B basically yeah, very low latency, then we want to choose models on the left and look at the ⁓ y-axis for the intelligence metrics. Now if we look specifically at lambda 3 the three.

Then we can compare all of the different companies that are able to provide this model. So we can see that Cerebras is actually the fastest company to provide for Lamatrida 3, but it's also the most expensive. So here it's ⁓ all of these figures are from artificial analysis, and yeah, I invite you to check them out if you're trying to choose an LLM for your use case.

Sneha Mehra (01:26:27)  
Then, ⁓ if fine-tuning ⁓ is something that ⁓ will be required, then the capability, the possibility to fine-tune ⁓ LLMs becomes something important to consider. So, with fine-tuning, we can benefit from higher quality results. We are able to ⁓ reduce.

The size of the prompt because the model already kind of knows what the task will be. And so by reducing the size, we can save on token costs ⁓ and by consequence also lower the latency. But yeah, we will talk more about this specifically in the next session, but we recommend that you only consider fine-tuning after you tried simpler techniques like prompting or even rag.

Before ⁓ trying to train ⁓ in an LLM. And why we say that, it's because the time, the the cost of training can be quite large, ⁓ just in time, for example. And we also know that newer and intelligent models keep coming out, token costs keep getting lower. So yeah, fine-tuning become can become something.

not very efficient to do. So if you're trying to fine-tune ⁓ LLMs, we actually recommend that you use the fine-tuning services of proprietary models like the OpenAI fine tuning service where you don't need to set up all the the hardware or renting the the instance with a GPU to do so. You can just provide the training data and then iterate

On the hyperparameters to get the best performance. But it becomes easier ⁓ to train to fine-tune a model using these kinds of services. So, for example, if you use the OpenAI fine-tuning service and you have a hundred tokens ⁓ in your training dataset, then you're going to pay by the tokens. You're not going to pay by the number of hours ⁓ renting a GPU.

Sneha Mehra (01:28:53)  
So if you train GPT-4.0 mini, for example, over three epochs, so three times ⁓ over the training data set for all the examples, it's only going to cost you ⁓ 90 cents. ⁓ It's going to ⁓ be actually quite cheap to ⁓ fine-tune your model. So in comparison, if you rent ⁓ a ⁓ an instance, ⁓ a machine, a server with a GPU.

It could cost you about two dollars an hour on Lambda Labs. And if you multiply that by the number of hours, ⁓ it can become really quite expensive compared to the ⁓ other service. Now, if we look at tokenization or vocabulary, and more specifically, ⁓ you you can consider basically this metric, this ⁓ factor if you're

Task requires you to ⁓ use LLMs in a different language other than English. Because, for example, if you look at the French ⁓ arena, so if you go back to the LM arena, you can ⁓ choose the category of French or whatever language you're trying to see. And here you can see that different models become ⁓ the the rankings ⁓ can really be different here. So

Here in in on this figure on this screenshot, we can see that actually GRUC3 is the best in the French language, followed by GPT-4O latest, and then the other ones. So whenever you're you're optimizing for ⁓ or you're you you will use the the models on a different language here, it can become very useful to look at these kinds of ⁓ benchmarks. So here I just list

Quickly, the most intelligent, the most used key closed models right now, but it's going to be changing in the following months, following years, of course. ⁓ And the the the models that are open weight. So all the Meta, Mistral, Google, DeepSeek models. And then for some key trends, we see that a larger context.

Sneha Mehra (01:31:20)  
Is a trend. So models keep getting ⁓ more and more ⁓ context, they can process a lot more context ⁓ each for each release. And then if we combine that with context augmented generation or cache to augmented generation, it can reduce cost and latency. So we will go back to that in in the next session, but ⁓ we we see that.

Larger context become ⁓ something ⁓ that keeps getting bigger. Also, the models no longer ⁓ work with only text. Most of them now ⁓ can process images, videos, audio, and that enables them to become way more useful for more types of application. Here I listed two examples. ⁓ So you have OpenAI operator that can control a web browser.

Or the anthropic computer use ⁓ tool that can let a model take over a computer. So this is very interesting, ⁓ and ⁓ yeah, basically it enables more types of applications. Then, of course, you have the reasoning models that enable the ⁓ latest agentec type of application.

You can see that with the OpenAI Deep Research on ChatGPT, that is very good. ⁓ Basically, the models become way more autonomous. They are able to use tools on their own, ⁓ reason about the results of the tools, choose to do ⁓ different actions, ⁓ and ⁓ these kinds of tools they can be very ⁓ powerful and save you a lot of time. So, this is something to keep in mind.

And if you're trying to build an application, then these reasoning models can really be ⁓ useful for that. To build agentic applications. Then you have these small distilled models. So even if the models, the most intelligent ones right now are very big and you cannot use them ⁓ on your local hardware, you still see some small.

Sneha Mehra (01:33:41)  
Models that can be used on small devices like your smartphone, ⁓ and they can be actually be good enough for simple tasks. So if you're just trying to translate something ⁓ or rewrite something, then these small models can be good enough for that. And yeah, companies still release these kind of small models. Now, if we look at some ⁓ points about security and privacy.

So, ⁓ in terms of data privacy and security, you need to keep in mind that services like ChatGPT or Gemini are not actually private. So you need to keep be aware that if you're sharing private information, then it's going to be seen by those companies. So if you're using your the ChatGPT app and you're providing private salary information, for example, then ⁓

⁓ yeah, of course, OpenAI is going to be able to see that, ⁓ and ⁓ you need to be aware of it so you don't actually provide that information. Then, like I said earlier, ⁓ if you use the model through through the API, for example, the OpenAPI, then OpenAI will retain the data for 30 days. And if you use tools like GitHub Copilot, then all the engagement data. ⁓

The accepted or discarded suggestions will be retained for 24 months to improve the service, basically. So the code it's suggesting, the code that you're writing basically is not private. Then, of course, if you're building your own application, it's going to be important to control ⁓ what the LLMs can see or not. So ⁓ if you're

Building a rag application, a chatbot application that can connect to a database, for example. We want to put some controls in place so the model that cannot access ⁓ all the data within the database. ⁓ And it's going to be important to log the interactions that the DLM ⁓ is performing. We will come back to this in the in the final session where we talk about jailbreaking ⁓ and ⁓

Sneha Mehra (01:36:07)  
Yeah, jailbreaking. Basically, what are the jailbreaking methods right now? So, yeah, it's going to be important to keep an eye on these access control topics. Now, in terms of guidelines, if you're in a company, it's going to be important to have a guideline. So ⁓ be aware of what is allowed or not to be shared with tools like ChatGPT. ⁓ Then for

Output management it's going to be important to ⁓ keep an eye that you're responsible ⁓ if you're building ⁓ an app on top of these of these models, that you're responsible for ⁓ handling the al hallucination problem. So you you want to be clear with your users if you're building an application that these kinds of models have limitations ⁓ and you that you're actually documenting ⁓ and

Yeah, keeping an an eye on these types of risks and try to mitigate as much as possible the bad outputs ⁓ of that the models can output basically. So ⁓ if you're building on top of models, you're actually responsible for the AI outputs. So you you need to be complying with the privacy regulation. And also you need to be aware that the models have

vendor terms or licensing requirements. So if you're using models by Meta, for example, you need to make clear in your application that the application is powered by Lama, it's built by Lama. ⁓ And if you're a big company, you have more than 700 million monthly active users, you will need to ask permission ⁓ from Meta before using the model yourself.

So, yeah, like I said in the previous slides, we are responsible ⁓ for what we show to the users, even if the system doesn't we don't own the actual LLM that the that the system uses. So we we always need to make sure that the system, the application, the feature always comply with security policies if we have security policies, that the system maintains consistent quality, that we detect.

Sneha Mehra (01:38:31)  
Potential quality issues very early in the pro in the development, and that of course that we meet the business requirements. So we will see this more in depth in the evaluation session that we will do in this course. But basically, we need to implement at least a basic structured testing approach. So

May making sure that we have some sort of evaluation process to catch those issues early. ⁓ Of course, we need some ⁓ metrics, ⁓ so those will depend on the type of task that you are doing on your application. ⁓ And of course, we can also build automated validation pipelines. So things like using other LLMs to judge ⁓ or anything really to make sure that.

the outputs the system is producing are correctly evaluated and validated before they are sent to the users. ⁓ And so that's it for ⁓ today's session. I hope you like it and I hope you learned some useful things for whatever you you want to build in the future.

