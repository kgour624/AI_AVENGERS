8  
Sneha Mehra (00:00:00)  
Hello everyone. In this lesson, we want to present you an in-depth ⁓ architecture design ⁓ walkthrough of the Brown ⁓ writing workflow. And to make this even more interesting, we will present you two versions of Brown, the writing workflow. Right? ⁓ And this happened because initially we actually had another implementation and design of what we will present you in the course in future lessons.

And this happened because we weren't really sure how to properly design it to have like a good balance between performance, ⁓ latency, and cost. ⁓ And yeah, basically we did what we thought was right. And we realized we made a lot of bad design choices. ⁓ And we initially basically we had to rewrite everything from scratch based on what we learned from all our failures. ⁓ And

We think that this is super interesting to teach you and to show you the first version which doesn't work properly and then to show you the good version with which which will actually teach you because this will happen so many times in industry while we work in real world scenarios. ⁓ And like this, we can walk you through our thought process and show you that ⁓ these scenarios are really natural and pretty common in in real life.

So before digging into the actual architecture, we want to like to rescope and refresh your memory a bit on what we actually want to implement here, right? ⁓ So at a very high level, like a black box expected behavior, we have something like this. So we have the research as an output from Nova, the research agent, right? Which we presented in previous lessons in in in that.

So you now know how to implement this ⁓ and everything about Nova the research agent. As for the Brown writing workflow, we expect the following. So we want, as input from the user, the article guideline and the research, where the research is a byproduct of the article guideline plus Nova the research agent. Right? So we want to input this plus some profiles that dictate.

Sneha Mehra (00:02:27)  
How the writing should be done, plus some examples as few shot examples. ⁓ And all of this will be as input to brown the writing workflow. And we expect ⁓ as the output the article. So we have four types of inputs: the article guideline, the research, the profiles, and the examples. And we have one output the article. Okay, and you might ask, ⁓ why can't we just prompt to an NLM something like

Write me an article on AGIS versus workflows, right? Why did this problem of a writing workflow ⁓ should have a capstone project on itself? Why is an interesting problem to solve, right? ⁓ And the thing is that in reality, you could write something like write me an article on ages versus workflow. You could, but you shouldn't. Why? Because to actually output high quality results, you have to do a lot of

Context engineering, prompt engineering around it. For example, just for this article use case that we'll present to you, we have to keep in mind constraints such as a mechanics. For example, we want to have an active voice. We want to always enforce a point of view like the second voice, like to always to refer to you, the reader, as you, to use only specific emojis and not overuse them, to to use abbreviations and acronyms.

only sometimes and only when we want to use them, right? So you want some control on how the words are actually ⁓ used. And on the other side of the spectrum, we want to avoid a common AI slope like ⁓ using too often words such as delve, streamlining paramount and these annoying words that are are are are so obvious that this is output made made by an AI that is not human, right?

Or the classic usage of dashes or the the semicolons, right? Which kind of break the narrative flow of the sentence and doesn't sound that good. Of that that classic formulation that you see think of ⁓ X as Y and ⁓ again, which some A slot that is super common and annoying when when when you say it in text. And on on the other side of the spectrum, again, we want to enforce a specific structure such as how we format code, because

Sneha Mehra (00:04:48)  
We write technical articles. So we want to ensure that the code, some media items are formatted as we want to, or every media item has some captions. You know, the text under the the image is always present there. ⁓ Or other very article specific stuff, such as this format of an introduction, conclusion and and sections, right? We always want that, plus the references at the end. So I won't dig too much on the the the art of writing articles.

because this is not the scope, but what I wanted to to to highlight is that ⁓ to actually output articles in a specific format, in a specific style, with a specific voice, specific structure, and like to chat all the bullet points that you want from this, there is a lot of prompt engineering around it, ⁓ and many multiple LLM calls that enforce this. So you cannot just ask

an LLM, write me an article on agents versus workflows and basically expect to read your mind and have an output ⁓ as you you ⁓ really want to. And this is an interesting problem because ⁓ you will be able to apply these patterns in many ⁓ use cases when building vertical AI agents. For example, when you want to output reports in a specific structure with with a specific format and all of that, you can apply very very similar techniques.

Or when you want to ⁓ build conversational chatbots that impersonate specific characteristics, ⁓ again you can apply very similar techniques. So you you can take all these learnings and all these patterns and apply them ⁓ almost everywhere in the LLM world. Okay, now before actually going into the architecture itself of those ⁓ two versions, let me quickly refresh your mind and show you some.

concrete examples of how these inputs and outputs actually look like. Okay, so now let's open the inputs from the brown writing workflow and see how how they look like, like to build an intuition on why we take some design choices.

Sneha Mehra (00:07:05)  
so we will go here into inputs, tasks, and open this this sample, ⁓ this zero two sample medium. ⁓ And let's first open the article guideline. And as you can see here, we walk people through how the article should look like. So this is an input directly from the user, which will be unique per each article, right?

So whenever you want to write a new article, you have to write such a guideline. Where from our test, usually you need like to write an outline, like basically the outline of the article, and then go in each section ⁓ and describe what you expect from the section and the section length. And as you can see, even from you, the user, you have to input quite quite many, many ideas if you want the final output to actually reflect your ideas.

And not some hallucinations made made by the LLM. Sometimes, as you can see here, maybe even your input is larger than the actual number of words that you expect from a specific section. And that's fine because you want like here to to dump all your thoughts without thinking too much ⁓ into how they sound, how they're formatted, and how they're actually written, right? That's the job of the

Writing workflow to actually take this ⁓ thought dump, you might think of it, and write it in in a nice format, in a nice presentation, right? Okay, so this is the article guideline. Again, we have basically the outline ⁓ and then we detail each each section per se. Then we have the research where we basically ⁓ just gather all the information.

That we have either in the article guideline here at the bottom as links, ⁓ or the research agent found finds ⁓ by itself. Now we won't go too much in into the research file because we already went through all the lessons that presented everything on how this this particular file is created. But what what's more interesting for us as inputs are the profiles, right? So

Sneha Mehra (00:09:28)  
We open here the profile directory, and as you can see, we have many many files here. ⁓ So we have six types of profiles. And each file is important. ⁓ And ⁓ we we we we name them as profiles because this is basically a way to dictate ⁓ different aspects of the writing ⁓ process ⁓ on different dimensions. As you can see, each file kinda

represents a different dimension of of what you want as an outcome from the writing workflow. For example, for the mechanics, ⁓ here we we we kinda explain how we expect the the wording to look like up to sentences and paragraph, right? So active versus the point of view, punctuations, emojis, and stuff like that. For example, for the structure, we we kinda dictate things such as sentence leg.

Paragraph structure, bullet lists, ⁓ and ⁓ everything that we expect about the structure of the writing itself. But for example, if we go to the article profile, here we have things that are very particular just for the article itself. Now I I won't go into each file and the details of it because we will do that in in in future lessons. But I want you to get a sense of what the pro

These profiles are ⁓ and their importance. Basically, here we dictate on multiple dimensions what the writing workflow should respect while it writes the article. Okay, and ⁓ as you can remember, as input, we have the article guideline, the research, the profiles, and we also have the examples. For example, which we have here. ⁓ And for this particular use case, we use these course lessons examples.

And here basically we actually have just articles from end to end, right? From from the actual title ⁓ up ⁓ to the end where we put the references. So everything end-to-end about an article with the code. And basically, we just take the output article end-to-end and put it in an example. And these will act as few shot examples during our

Sneha Mehra (00:11:52)  
Article generation workflow. ⁓ So yeah, these are the the four inputs. Again, we will detail more on the importance of each and how we pick them in future lessons. But these are the four inputs, and the only output is the article itself, which is this one for the sample. And as as you can see, we have the title.

And yeah, this is the article. And just just let me like open the preview and look at it like this. ⁓ So this is the final output. ⁓ And as you can see, we have the title, the subtitle, the introduction, then we go through through through the sections. ⁓ And we have submerged diagrams, which are our like media assets for this particular example. ⁓ And yeah, we have the conclusion, the references, and as you can see.

This is an article that I I never touched manually. So this is fully generated end-to-end. I haven't changed the word. And as you can see, you don't really have dashes, you don't have like semicolons. ⁓ the wording is pretty pretty nice, pretty humane. Like it starts with when building applications, engineers face a critical architectural decision early on. Yeah. You have this critical architectural decision thingy.

But a as you go through you will see that ⁓ it's a good balance between an AI engineetic text ⁓ and a human written text. But I want to highlight that these articles don't use like the full fledged ⁓ power that we actually use in production. Like i it was written with Gemini Flash, which is a smaller model, not with Gemini Pro.

It did not go through multiple editing and reviewing processes that I will show you soon how they work. ⁓ It was just like a one-shot ⁓ example. But again, let's look at this paragraph. In an AI agent, an LLM dynamically decides a sequence of steps and actions needed to achieve a goal. Sounds pretty pretty decent. ⁓ again, here you f you can find an M-dash. But as you can see, there are not that many bullet points.

Sneha Mehra (00:14:07)  
The paragraphs have have a nice structure. We have these references. ⁓ the tr there are transitions between each structure. So it's a pretty good written text. It's like an eight out of ten. Not perfect, but not far from something that you can just take and use as is. Okay, so enough chit chat, let's jump into the system design. Let's start with version one.

of the brown writing workflow, right? The version that ⁓ we decided not to go not to go to, but we learned a lot from it. So before digging into the actual steps of the design, let's go on how the article structure should look like, right? Because it's important to understand the article structure before understanding this design. So like a standard article looks like this. We have the title, we have then an introduction, then we have multiple sections and then the conclusion.

plus some CO stuff, sometimes like a different title, plus a description, ⁓ and things like to optimize the the searchability of the article. And you can complicate more this like each section can have multiple subsections ⁓ and all of this. But at a very high level, this is kind of like each article looks like or every written long form piece. It looks something something like this. Okay, so when when when we send this article

Structure, we naively thought about the following ⁓ design pattern, right? And the system design, the architecture of the first version of Brown looked something like this. So we we split it into three stages. So we have a stage one, stage two, and stage three. For stage one, we took ⁓ all these inputs which are a bit different from what I explained to you.

What I explained to you now are like the final version of the inputs. ⁓ here are a bit different. I will focus on that a bit later. But let's look into these ⁓ three stages for now. So we have stage one where we have as as input the article guideline, ⁓ and we do a one-shot generation plus some reviewing and editing on top of it using the orchestrator worker workflow pattern. ⁓ And again, we will dig more into this.

Sneha Mehra (00:16:31)  
into this pattern itself into next lessons. But for now, what is important for you to understand is the following. So usually the evaluator optimizer pattern has these two stages where you have a generator, where in our use case is the article writer, and then you have an evaluator, which in our use case will be an article reviewer. So the generator generates the article and then the reviewer, the evaluator generates some reviews.

And then the generator takes back all these reviews, these evaluations, and updates its initial input. Basically, our article writer takes the reviews and edits the article. And then it generates a new version, which is passed again to the evaluator, in our use case, article writer. And you have this loop until usually a specific threshold, a specific score is reached.

Right. For example, if you score it from zero to one until ⁓ zero eight of the score is reached or until a number of iterations is reached. And for now to understand the high level architecture, this is enough, more than enough to understand about this pattern. So what we did here basically ⁓ is we did a one shot generation, plus we applied the this evaluator ⁓ optimizer pattern until a specific score was reached.

But this was not like a naively ⁓ one-shot article generation. And what we did was we thought it was smart to decouple the article guideline from the actual final outline of the article. Because we thought that, hey, we don't want the user every time to define the whole article ⁓ outline ⁓ of the final piece. So we thought it was necessary to ask the LLM first to ⁓ decide.

On the outline and based on that, start to write based on this generated outline, right? Which is an introduction, section one to end, and the conclusion plus the guideline. But we saw that it's it's not that great. But anyway, ⁓ I will explain that a bit later on. But to to move on to the next section, so now we have one-shot article which was edited in this one-shot way. ⁓ And then we thought that.

Sneha Mehra (00:18:56)  
To actually specialize and to actually refine this article, we have to apply the evaluator optimizer pattern to each section of the article, where each ⁓ like evaluator that checks the the those sections and each editor that edits those sections should be specialized for for each section. For example, for the introduction, you want to have an

like an LLM that's prompt engineered to write something more engaging, right? To make the the reader actually write everything. For each section, we wanted to have an agent more an LLM prompt engineered to write better code descriptions, to go into more descriptions and things like that. And for the conclusion, again something that's more straightforward and gives you some next steps ⁓ and things like that. So basically

We thought that for each d for each type of section, you need a specialized LLM with some specialized prompt engineering. ⁓ And that's why we we we we took this division. Plus, we thought that by doing this separation of concerns, the performance will be higher because the LLM will be focused just on editing that particular section and not the whole article, right?

So we avoided this needle in a haystack problem, and we we we were able to keep the the contest more contained and more granular and and all of that. So we thought that hey, let's do this specialized ⁓ stage. ⁓ And here we actually computed an evaluation score for each section. And based on that, until ⁓ each ⁓ section passed that evaluation score, we we kept running this.

This evaluation of optimizer pattern. So you can start seeing that this this starts to get tricky, but we managed to implement this and it kind of worked. ⁓ And the final stage, we thought that, hey, after we do this refinement on each stage, we saw that each stage started to get a bit isolated and ⁓ not connected with with the whole global article. For example, when we had acronyms, for example, the acronym REG.

Sneha Mehra (00:21:17)  
⁓ in multiple sections. DLM taught that this is the first time the acronym is introduced. So for example, when you had the second introduction of the RAG acronym, instead of writing just RAG, ⁓ it always written retrieval augmented generation and then in parentheses RAG. ⁓ And basically because we did that editing at the section level, it we had we started to see discontinuities ⁓ at the global

level level or the article. So we decided, hey, let's just run this evaluator optimizer pattern again over the whole article to like refine and glue ⁓ all the ⁓ sections together ⁓ back again. So again, we had another iteration of the evaluator optimizer pattern. And lastly, we thought that generating the Titan and CEO independently with some prompts that are super specialized just for the Titan CO.

⁓ should again be a good idea. ⁓ And as you can see here, ⁓ we had too many steps ⁓ with too much complexity and too many LLM calls. And this resulted in a huge workflow that was super slow and was super costly. ⁓ Because like we had so at minimum we had like three LLM calls just in the first stage, then ⁓ two LLM calls just per section.

So this is was dependent on the number of sections. But let's say that is an article just with one section, one introduction, one conclusion, six ⁓ LM calls, ⁓ and another one, another two here at the end, and another two here at the end. So we had two, four, ten, thirteen LM calls just for a small article. And these two for for for for example, on average, like

while we started using it, this whole thing was was taking like 30 minutes to run and like one to dollars per run while using Gemini flash and pro depending depending on the situation. So one to dollars per run ⁓ writing workflow is a lot. Like a lot, a lot. This clearly doesn't scale if you want to run ten runs is already ten dollars. So this was clearly not sustainable.

Sneha Mehra (00:23:42)  
And other issues that we had with this design before like digging ⁓ and presenting the the fix is that ⁓ we try to follow the book, where the book teaches us that you have to isolate context ⁓ and you have to write specialized LLM or specialized agents for very particular things. And we took took that to an extreme and we we written the specialized LLMs. ⁓

for each section, for the introduction, for the conclusion, for example, here for generating the the outline ⁓ for editing and all everything was a specialized agent. And we realized that that's not good ⁓ and because the the context started to be too fragmented. ⁓ And there is a balance that you want to think about between like having specialized agents and fragmenting context.

Because context actually is valuable, right? It gives you information about everything that is going on. For example, here when we write the title just based on the article itself, you might miss on some very interesting perspective rather than writing the article based on the article itself, based on the research, based on the article input from the user, right? So you might miss on some very interesting insights when you do that.

And that's very similar when you edit the article, when when when you write particular sections. ⁓ As you remember, for example, when we edited a particular section, because we did not add in context other sections of the article, ⁓ it started to to to go off rails and introduce and move away from the whole picture, right? So we we tried many techniques like introducing

actually introducing the whole article along that section, but ⁓ at some point we managed to make it work. But as I said, the code ⁓ became very complicated, very clunky, the latency was was huge and up to 30 minutes, the costs were high, ⁓ and we had many, many issues with this design because we tried to apply by the book principles. ⁓ And one important issue

Sneha Mehra (00:26:04)  
That we also had with this, and I I think this is one of the most important ones, is that we couldn't probably introduce human in the loop here. Right? Remember, when you build workflows and even agents, you want to create this generation validation feedback loop, right? Where the AI generates something from you, ⁓ a human looks at output, gives some feedback, and then the AI starts to improve that.

Think of it ⁓ something super similar to the evaluator optimizer pattern, but in the generator is the whole AI system and the reviewer is the actual human in this use case, which gives feedback, reviews to the to to the AI to further improve them, improve on itself. We kind of manage ⁓ to do that here by, for example, stopping the generation here after the ⁓ article was generated with a one shot.

We we we kinda managed to do that for example by stopping the workflow here ⁓ after stage one, where we had the one shot the first one shot generation ⁓ and giving some feedback directly in the CLI and we also tried to do it here ⁓ and again we we we kinda tried to force it, but this was not natural because here you could add the the feedback ⁓ only while the article ⁓ was actually generated, not after. So

You actually want ⁓ to leave ⁓ the AI system do its own thing and that the reviewing should be completely decoupled from that. And then pass the initial input plus the review to the system and update the the output again. A very good example is is what we have here in in cursor or or in any other coding tool. I I think they've done a great job at this.

So for example, we have this article which was generated yesterday, and we have time only today to actually ⁓ refine it. ⁓ And what we can do here in this native ⁓ coding experience, we can just take this ⁓ or pass this to cursor, ⁓ or do some feedback like make it shorter, right? ⁓ And this is our feedback, right?

Sneha Mehra (00:28:26)  
And it takes care of it and it updates this per particular output just on what we wanted ⁓ to ⁓ improve. ⁓ And we kind of took this design into the second version of the Brown Ranking Frowflow and improved it using this this option. And as you can see here, now you you can like similar to any coding experience, ⁓ it ⁓ suggested you the improvements and you took what you think it makes sense.

⁓ And what it doesn't, you can just undo. I will just undo in this particular use case everything. Okay, now now I want to wrap up this with other some poor design choices, which weren't necessarily a deal breaker, but there are some good learnings that that I think are important to highlight. Okay, so because we use the evaluator optimizer pattern, we thought it was a good idea to use a different set of evaluation rules it should respect. ⁓

Basically relative to the the the style guide, ⁓ the writer profile, the examples, and and all the other inputs that ⁓ guided ⁓ the workflow. Here I won't go into why the style guide and writing profile and all of this is different from the profiles I showed you because it's not irrelevant. ⁓ But this this particular design of evaluation rules is interesting because it just duplicated.

The context, it just duplicated the logic and and all of that. Because in reality, you want the evaluator, in our use case, the article reviewer, to check that your style guide, your writing profile, your examples are respected. You don't want to have something on top of it that adds more complexity. ⁓ And another poor design choice was ⁓ this ⁓ separation of creating a new outline.

Based on article guidelines relative to what the user inputted. It just added more non-determinism ⁓ and ambiguity, and it was confusing that you expect one outline and the final output was completely something different. ⁓ And the third thing was that we decided here on when to stop basically the evaluator optimizer loop.

Sneha Mehra (00:30:50)  
based on this evaluation score. So for example, here on ⁓ on the introduction, we computed the evaluation on the existing introduction. ⁓ And we have the evaluation scores. And based on that, we we took this decision if we further edit the introduction or or we just stop. And I think is that with computing a score is very ambiguous, right? So ⁓ it's very hard to compute a score that you can rely on

And ultimately, what's good writing? You know, you you count the number of slopes, you count how many paragraphs respect a particular ⁓ particular length, like it gets super complicated on what exactly you want to respect. And that's why ultimately we decided just to run it for a particular of number of iterations, refine it on all those rules that that we want, and that's it. And don't rely on scores. And again, I want to highlight that.

You have these patterns, you have the how they should look like by the book, and but you should always ⁓ think ⁓ by yourself if do they really make sense as they are on their own in your particular use case, or you should take them and refine them. And that will lead me to the final bad design choice that we took with this ⁓ design, is by using Langgraph and Langchain too hard. Because if you rely too much on an AI framework, right?

You don't have the freedom to customize all these patterns as you want. ⁓ So we were kind of forced into respecting that their own utilities. So we decided to write everything from scratch. And we also started using the Lang Graph Graph API with those nodes and edges ⁓ which was which made all of this super complicated. So when you first look at it,

You think it it it really makes sense to be a graph, right? It kinda looks like a graph itself. You see this diagram, you think it will fit very naturally. But it doesn't. Like to to to take those decisions with those ⁓ nodes and edges ⁓ and it just overcomplicated the code more than it should be. So instead of writing some s simple if-else branches, ⁓ we had to use those edges logic and everything became so so complicated. And we just decided to drop it.

Sneha Mehra (00:33:18)  
So ⁓ this was our first design choice. Again, I might have lost you ⁓ in some particular sections, but that's the whole idea. It was too complicated, too complex, very, very hard to extend, ⁓ too much duplicated code, it ran too slow. ⁓ the cost of running this thing once was too high. ⁓ we couldn't add a proper human in the in the in the loop here. The evaluation was

kind of ambiguous, like we had a score that we couldn't rely on. The input wasn't properly normalized, right? So this is normal when you first try to build something that is not already built because the the problem is not super clear in your mind, right? So that's why you do a first iteration of how you think on how you might think things should look like. ⁓ And that's more like the prototype where you learn a lot on

what works, what doesn't, on how things should actually ⁓ be implemented and designed. And that's completely normal. Like from my experience, very often you you end up doing a a first iteration, which is kinda clunky. ⁓ you just want to see if it works or doesn't. In this use case performance wise, it worked, right? So the final output was a good article, but at the expense of all this mess that was going on.

So that was a strong signal that ⁓ we can do this ⁓ and we can actually rewrite this in in a better way. Before digging into the second version of the Brown writing workflow, I want to give a quick note on how can we actually know that when we transition from this version to this version, how we do we know that the performance stands, right? So okay, we had all these issues with the versions, but the performance was good.

How do we know that after we implement the second one that's I don't know faster, cheaper and all of that? How do we know that the performance still stands? ⁓ How do we know just more than looking at the article and ⁓ giving saying that okay, it looks good. ⁓ I'm I'm cool with it. So how do we know that the the new architecture ⁓ is better than just five checks? Then just just just looking at the final outcome and thinking that okay it

Sneha Mehra (00:35:42)  
might be better or might be not be better. So the trick here is to implement EIE valves, right? So basically EIE vals should be a layer on top of your application that are independent from your actual architecture. And they're dependent just on the data, such as the input and output to your system. And like that you can compute ⁓ a benchmark, right? A score, a metric ⁓ on this version of your architecture. And then

After you implement your second version based on your inputs and outputs, you can ⁓ run the same EIE valves, right? On this new version, and you can very easily compare the two methods and decide performance-wise which is better and which is worse. But we will dig ⁓ deeper into the EIE valves problem and how to frame it, how to implement it and all of that into the part three ⁓ of this course.

Which leads us to the to the second version ⁓ of Brown the writing workflow. Okay, so you might think that there's a lot going on, but in reality, a lot of this is duplicated here. ⁓ So we have three workflows. So this time we decided to split all of this into three workflows. We have just a generate article workflow, then we have an edit article workflow, and an edit only a selected

piece of text workflow, right? Similar to to what we we we showed you here where we select a piece of text and we want to edit just this instead of the whole article. So every workflow will be exposed as an MCP tool. So ultimately when we will serve this, we'll have three MCP tools. A generate article, an edit article, and an edit selective text mcp tool. Okay, so this is like ⁓

Bigger picture of the system, and this is like our final outcome. And now let's dig into the into the generate article workflow. ⁓ And actually, the other two workflows will be quite similar with small changes. So as you can remember, the user input is the article guideline and the research. ⁓ And we also have these other inputs which are static, such as the profiles, right? So we have six profiles as you can remember.

Sneha Mehra (00:38:07)  
These ones ⁓ and ⁓ I I will ⁓ explain a bit later what and what not we can customize. Okay, and now to zoom out a bit to to to like the generate article architecture, we have three we have three larger pieces, right? We have this orchestrator worker pattern which we applied to generate various media items based on the article guideline and research.

We have the actual article writer, which is actually just an LLM call this time, so we do everything in one shot. And we have the evaluator optimizer pattern again, which was used to review and edit the the article. But this time we applied it to the whole article, which is a lot easier to do. And now let's walk through this ⁓ step by step and see how how this works.

So we have the user inputs, we have the orchestrator worker pattern which work as follows. We take the article guideline and the research, apply it to the asset generator, which in reality is ⁓ is like the orchestrator which understands these inputs ⁓ and generates two calls for every media item that should be generated based based on this this input.

For for example, if in the article guideline we have three requests to generate three Mermaid diagrams, it will generate three tool requests for media diagrams with particular inputs based on this these user inputs. If it finds three to generate three Mermaid diagrams and one image, it will do that as well. So per media asset we will have ⁓ one one tool call which we will run in parallel.

And we'll gather here in a list of media media items. Right. So now we have this specialized orchestrated worker which just generates media. And this is like the only specialized element that we have because we've seen that if we try to do it to generate media from one shot, ⁓ it it performs really poor because ⁓ for example, the mermaid diagram code specifics, ⁓

Sneha Mehra (00:40:28)  
are are really specific and they started to miss very small details that were annoying. And also, for example, if we want to extend this pattern to generate images or videos or whatever other min media items, you cannot do that in in a one-shot LLM call. You need you need to to do that through mo other LLM calls and specialized models, right? And this is optimal because in one go we just

We understand what we need to generate. And then we call all these tools in parallel. So the time to execute all of this will be ⁓ the maximum time of the like the longest tool that takes to run. Right? Okay, so now we have the media items which go into the article writer. We have the user inputs, right? Which go into the article writer. And then we also have these profiles and examples, which go into the article writer.

And now let me explain more on what's going on in here. As I said, these profiles are basically these files that we have here. They're also markdown files that are inputted as static files. So these are not user inputs that are different for each article, but they kind of define how an article in general should look like. And we actually have this terminology, tonality, mechanics, and structure profiles, these four.

That are in reality independent from our article. They just dictate how your output should look like in general, right? They talk about general stuff such as avoiding AS law, having a particular tone, right? Following some specific mechanics such as active voice and the second person. Structure-wise, here we we mostly talk about how you should output the media items and code and things like that.

So nothing very specific to an article. That's why we have ⁓ a completely different article profile where we specify very particular things on what we expect from articles, such as the article structure, right? Introduction, sections, and conclusion. And also a writer profile, which kind of dictates very particular things about a particular writer, in this use case ⁓ or myself, right? Because that that's what I knew how to talk about. ⁓

Sneha Mehra (00:42:54)  
Where I just give give the LLM some background about me, some niches that I talk about, and some similar personas that I want to sound like, some styles and things very particular to me. While in this terminology and tonality profile, we we we just want to prompt the LLM to write in a general way and on how we expect, mostly like not not to sound too AI-ish, and then

We put on top this writer profile that adds our particular voice. And also we have the examples, which kind of control how the output should look like. So that's why I highlighted that we can customize this tree: the writer profile, the article profile, and the examples. Because, for example, if we want to switch from this to writing social media posts, we have to write a social social media profile ⁓ and inject different examples, right? You don't want to

Add examples of how to write an article. You want to add examples on how to write social media posts. And for example, if you want to switch from LinkedIn to X, you will probably need different examples because the posts themselves are different, right? So for each output, you will probably need to configure this. And also you can configure the writing profile, even for you, from social media ⁓ from social media platform to social media platform because you have to sound different. ⁓ Or I don't know.

you can do something more general and write a writer profile that sounds like ri Richard Feynman or or something like that, right? So you can take this and go go wild with it. Okay, so at at this point we have the first ⁓ level of the article. But as you will see in future lessons, at this point the article the LLM ⁓ and the article writer doesn't really respect everything that that you put it here. Because

If we start to open all these profiles, I won't go into what's going on in here, but that's kind of the point. You will see that there are many, many rules to respect, right? So expecting the LLM to

Sneha Mehra (00:45:05)  
reason through all of this from one shot ⁓ is kind of tricky. It's a a lot to ask, right? At least at this point of of the LLMs and AI in general. And that's basically the idea of ⁓ using the evaluator optimizer on top of these profiles, plus asking it to also check if the article guideline is respected, right? Because here in in the article guideline

It also needs to reason on respecting all these bullet points that you asked. Basically, it it needs to follow this article guideline as you instructed it in here. So there's a big chance to miss a particular bullet point or a particular section or or something, right? So we here during the review, we basically check if this article guideline is respected. We check if the profile is respected, and also one.

Really important thing is we also we also check if the piece is anchored into the research, right? And and this is one way to avoid hallucinations. A really important thing because we always want the article to be written based on the guideline and the research. The LLM should never but never ⁓ use information from its ⁓ own internal knowledge. Because sometimes it ⁓ the information may be correct.

And sometimes should can be incorrect. But we we don't want to bother with ⁓ this decision making. So we enforce it to always and always rely on facts from these user inputs. And we check that as well during the reviewing process, right? So as you can see during this article reviewing process, we do three things. We see if everything adheres to the profiles.

So basically the structure, how everything is written, the islob and all of that is respected. We check if the article follows the article guideline, ⁓ the user input as you want. And also we check if the facts from the whole article ⁓ is are taken only from the research and from the article guideline. If not, we consider them automatically as ⁓ hallucinations and we double fact check them.

Sneha Mehra (00:47:28)  
Okay, so this this is what happens during a single article briefing process. And just to highlight relative to the previous architecture, this now happens on the whole article at once. So it's a lot easier, right, to to check for examples that are done on the whole article. So these examples are relative to the whole article. Not ⁓ we we we have a fragmented ⁓

By sectioning the article where we we previously had examples, for example, just on how a particular section should should look like, just on a particular introduction and conclusion should look like. Now everything is scoped globally, and it's a lot easier to gather examples, to manage, to everything. Okay, so ⁓ now we have the the reviews from the article reviewer. We have the article writer, which takes as input.

All the input that we had previously here in the article writer plus the reviews. Because this is yes, this is the same article writer as here. It is the same class that we use, the same prompt, ⁓ the same LLM, the same everything as before, plus those reviews. And there's some custom logic into this article writer that takes ⁓ everything as before plus the reviews, and it knows that now it has to add it.

a previously written article based on these reviews. So now we have the reviewed article. And the question is when do we stop? Right? Because I as you can remember, I said that now we are not computing any score anymore because it's very ambiguous. Like how do you compute scores based on this profile, based on the article guideline? You could think of something, but it it can quickly become complicated and not reliable, ambiguous and all of that. So

So, because we want in the long run to rely on human feedback and on an iterative process, ⁓ we just dropped completely the scoring system ⁓ and we just relied on the max number of iterations. And to keep this light, we added only two iterations of reviewing and editing, ⁓ except of like I don't know, three, four, five, six. So we keep this very light because

Sneha Mehra (00:49:47)  
If the article is not yet perfect, because we don't expect after just two number of refinements the output to be perfect, but we will let the human decide decide on that. And I will explain how that works a bit later. So we repeat this, for example, for two times in our particular use case. And then we have the final final article. And that's it. So ⁓ now nothing too fancy, nothing too too too crazy. ⁓ And now basically in ⁓ LLM call.

If we count all our LLM calls, now we have just ⁓ an LLM call per each media asset, which can be a small LLM because you don't need a lot of power to generate these media assets. And you usually, for example, have I know three, five media assets per article. So let's say on average three LLM calls of small LM LLMs, plus the actual asset generator which generates all these tool calls. So let's let's say four.

LLM calls here, another one here, five now, ⁓ and another two per each refinement step. So we have seven for one iteration and nine for the second iteration. And that's it. So just on on average, just nine llm calls, except of like on on average 20 lm calls ⁓ from the previous strategy, like almost double and

The previous strategy was highly dependent on the number of sections, right? So that was like for three sections. If you want to do five, six, seven sections, that number can go crazy. And you usually want ⁓ that for an article, right? ⁓ And now the article is a lot simpler, is more modular, is more flexible, is is customizable, ⁓ is easier to implement, ⁓ and yeah, it it's better in almost every dimension possible.

So we reduce the running time from thirty minutes on average to like five minutes. The costs are reduced from like one two dollars to zero to like ten cents or something like that. So everything is better. And you might think that because we always write the article and do everything at the article level, right? Even the the title, we write it here in the article writer at the article level, everything.

Sneha Mehra (00:52:10)  
that the performance will be worst because like we input the whole context ⁓ at once ⁓ and there is like this this huge input prompt ⁓ and everything is done into into a single shot. ⁓ And you might think that but in in reality ⁓ if you're careful ⁓ the final output ⁓ is is as performance right so

In our use case, the final outcome was as performant with lower cost ⁓ and lower latency. And before highlighting other ⁓ good aspects of this design, let me quickly walk you through the ⁓ edit article ⁓ and edit select text from this design. Okay, so this edited article workflow and edit selected text workflow actually rely on the same building blocks.

That we used before. As you can see, we have the evaluator optimizer with article reviewer and article writer. ⁓ And there's like ⁓ logical-wise, there's nothing new. The only difference is how we we glue things together, right? So for example, in this use case, we are we care about only applying this pattern once. So we don't care about this this loop anymore at all. Because we want to make it very iterative ⁓ and

rely a lot on adding the human in the loop, right? So here the core part is having the this editing logic on one end and then the human on the other end the and the human calls these workflows ⁓ as often he thinks is is is is necessary. So that's why the in most interesting part is seeing what are the inputs to this to these workflows, what are the outputs and how can we actually apply this.

To to the final outcome and what the user already sees. So on the on the input side, we have the article guideline and the research, what we already ⁓ used to. We have the old article and some user feedback because now the the user can actually see what the article looks like ⁓ and can provide some feedback, instructions on what it should be improved, what should be expanded, what should be removed, ⁓ and all of that, right? ⁓ So

Sneha Mehra (00:54:36)  
As you can see, we we get this article, we apply the evaluator optimizer as before. It uses the same profiles, the same, same everything as before. ⁓ we have the reviewed article, and then we have this step where we take this reviewed article, we take the old article, and we apply these changes to the old article and ask the user ⁓ if it if it wants the new changes, right? Because maybe.

it likes the new changes, maybe it doesn't. Exactly as as as you've seen at the beginning, how cursor did it ⁓ with code, actually in that article that I've showed you, right? So we will ⁓ mimic that exact same behavior. So we have the new version of the article, we have the old version of the article, and we apply these changes and let the user decide ⁓ what it wants and what it doesn't. And then we have the final reviewed article.

And on the other side of the spectrum, when you have this edit selected text item, is pretty similar. As you can see, they're almost the exact same elements. But the only difference is that this time we add this input along the whole article, the selected text that you want to edit, which it's the MCP client's job to extract this text and pass it to this workflow, where the MCP client usually is like ⁓ a chatbot, right?

What we have in cursor is a chatbot that can do that and call this additive selected text as a tool where you have as input the selected text and it basically ⁓ extracts it and passes it to the evaluator optimizer. So now we run this pattern only on the selected text. We have the reviewed section as output from this pattern, ⁓ and this time we have the review section and apply it to the whole article. ⁓

We have again using the MCP client, we promptly engineer it to take the section and apply it as a patch to the original article. And then as before, the user can decide what ⁓ it wants to accept, what it doesn't, ⁓ so everything is super interactively with with the human in the loop. And again, we have the f a reviewed article at the end.

Sneha Mehra (00:56:58)  
So, because we created these specialized edit article and edit selected text workflows, which are exposed as MCP tools, we can very easily plug them to cursor or cloud code or any MCP client that you want to use in reality. ⁓ And because of that, we can really easily create this ⁓ AI generated human feedback loop.

Which is super important in today's applications. Like that is the pattern that you want to follow in any any AA product in reality. And again, to repeat, the first step is to use the this this workflow to generate the whole article, right? So the start. Then a human looks at the output, this output, and then depending ⁓ on what he thinks.

It can apply the whole ed edit article to the whole article plus some user user feedback or without it. And then it will just apply this optim this pattern again based on these profiles, based on its adherence to the article guideline, and based on it adherence to the research, right? Right. ⁓ Basically mimicking the exact same logic as it does here. ⁓ Or plus the human feedback where it still does everything that it does here.

plus ⁓ looking at the human feedback as well. And again, it can choose if not editing the whole article, it can choose just to edit some piece of text where the same logic is repeated, but just looking at that piece of text. And actually this is the only thing that is inspired from our old architecture, right? Where we did this checking automatically here when we wrote article. But we took this

⁓ And implemented it only optionally where it makes sense and only on sections that makes sense. And we we can do this even not on the whole section, right? We you can do that ⁓ just on three words on a paragraph or ⁓ or just on a diagram or whatever it makes sense to you. You you can use that. So like this, we just decoupled this ⁓ very long running and costly job to run it just on

Sneha Mehra (00:59:20)  
When it makes sense. And just just to refresh your mind and show you how, for example, how this edit selected text command works ⁓ in cursor, let's let's move to cursor and see and put it in action. Let's pick, I don't know, let's pick this this this section. So this is a section, a subsection from the understanding the spectrum from workflows to a agent sections. ⁓ And we do control L or Command ⁓ L to pass it here.

To the chat and then we do slash ⁓ brown edit selected text prompt. Now here at the top we pass some some feedback. Let's say reduce this to one paragraph. I know. Something not that

Sneha Mehra (01:00:12)  
Imaginative and l let's see what it happens.

So as you can see, now we haven't still called the Brown ⁓ workflow. We are still in cursor. And this is a prompt that instructing cursor ⁓ what tools should it call from ⁓ Brown and how. And basically, now we have this edit selected text tool call. And cursor asks us ⁓ to accept it or not. ⁓ We can take a look at the parameters and we can see we have this article.

That is opened. We have the the human feedback, which is exactly what we wrote. We have the selected text, which is exactly what we inputted. So from AI agents up to this own adventure, on adventure, ⁓ and the line numbers of of of this piece of text. And then we can hit run ⁓ and basically only now we are calling ⁓ okay. So it pick up that because it was a relative.

tool is not working, which is amazing because because of this folder structure, it works only with absolute absolute paths, but that's okay. I I stand by what I said. Only after I accept this tool, ⁓ our problem writing workflow actually starts to run. ⁓ And we we execute the edit selected text workflow, which takes this text and applies the evaluator optimizer just to it.

Okay, and let's let's wait for a bit for for this to end. Okay, so this this ⁓ edit selected text workflow just just finished. Just let me quickly show you like what's the final output. So we instructed it to reduce everything to to one paragraph. That's exactly what it did. ⁓ and it now it asks us if we want to accept this this new option or not. ⁓ And

Sneha Mehra (01:02:14)  
How is this different just from directly asking cursor to do this? It's because instead of using directly the LLM from cursor that is isolated from our whole application, now we actually load all this context, like all the profile, the guideline, everything that I spoke to you about. It loads automatically all that context and applies exactly our prompt engineering and context engineering to it, and it adapts the text based on those instructions, not on just

your human feedback that reduced this to one paragraph, right? So it knows exactly what you want, like this, which is important if you want this to actually make sense, sound good, and and all of that. Okay. I will just undo for for I did this for the sake of the example. ⁓ And that's mostly it about the architecture. So we have these three steps expose NCP tools ⁓ and with everything that ⁓ I explained.

Okay, and before wrapping up, I want to end this this lesson with one last question that you might ask yourself, which is ⁓ why are we injecting the research as a markdown file and we don't do reg on top of it, right? Because this this reach the research file can be quite quite big. So here we have a trimmed version of it because we wanted to keep it small for the samples.

Which is already quite quite big, right? From our examples, we we had files that can go up to like tenk, twenty k, thirty k lines. ⁓ And in reality, the question is you could do rag on top of it because we just grab factual data from from there where where it makes sense. But let me explain you why for this particular use case it's not necessarily a good idea. So let let me go to this.

to this ⁓ sketch that I did here. ⁓ So why haven't we done rag to inject the research? Because this can also be reduced to the to this ⁓ question of cag versus rag versus agentic rag, where cag is basically the idea of injecting stuffing all your your context in into the prompt of what what we did, plus some other techniques to optimize it. But that that's not the point for this conversation.

Sneha Mehra (01:04:35)  
Rag is just retrieving what you what you need when you need. And genetic reg is letting an agent using particular tools to retrieve ⁓ stuff from your database when you need it. So here what what we did was our research agent generated this research in a markdown file, and our writing workflow just takes the the markdown file and uses it as is. We throw in everything into them the writing agent, and that's it. ⁓ While

If we wanted to use rag, things will start to get complicated. ⁓ And basically, instead of just dumping everything into a markdown file, the research agent will have to ⁓ own its own rag ingestion pipeline, which starts starts to have all these rag problems of chunking, embedding, ⁓ and storing. This is not a rag curve, so I won't dig into what this necessarily means, but

What's important for you to understand now is that you have more logic, more complications, ⁓ more complexity, more things to manage, more things to write. And then you will need to put your research into a database, which usually has this vector and document support because you want to do semantic search ⁓ and also do text search for documents and storing them. And this happens just on the research agent side. And then on your writing workflow side, we'll probably need to write.

Another agent that has some rag retrieval tools, as I said, that do semantic search, text search, or whatever techniques make sense to retrieve all that data when you need it. So this starts to become you also introduce a retrieval quality problem. And you also start to have multiple LLM calls, introducing again latency costs and going back to our initial problem that's we added too much complexity for no reason. And I want to conclude this idea with okay.

The idea of reg is beautiful, but before going into reg, you should always think about how big is your context. Because in reality, let's say that on average we will have around 10 long articles in our research of around 40 words, which 10 times or 4000 times on average, like ⁓ one word is ⁓ one ⁓ and thirty-three ⁓ tokens.

Sneha Mehra (01:07:00)  
So this on average results ⁓ on about like 53,000 tokens, which is very good. Like Gemini in in theory supports up to 1 million input tokens. But from our experience, starting from 150,000 tokens, it starts to get sketchy. The DLM calls start to fail, are not as reliable, not the performance. But here we are only at like 53,000, so up to 150\.

thousand there's a long way to go. So basically the for most of our use cases this will do. ⁓ And in reality for many vertical AI agents use cases, you don't work with big data, you work with small data and with some simple of filters, you can reduce your context to something like this and you can just completely drop your whole rag ecosystem and use CAG.

Which will make your whole design a lot easier, faster to iterate on, faster maybe even to run and and everything. Of course, there there is a lot more to talk about this ⁓ on this topic. ⁓ And unfortunately, ⁓ this this this this lesson all already got too long. But I just wanted to highlight ⁓ the thing why we decided to go with CAG, not with rag. And yeah, ⁓ in future lessons we will see how to actually implement this.

We will implement everything that you see here. So in the next lesson, ⁓ we will focus on this writing workflow. Basically, without the evaluator optimizer part, we'll focus on implementing the orchestrator worker ⁓ and the article writer and seeing how these profiles actually work, how it follows the guidelines, and ⁓ all the prompt and context engineering behind writing this high quality article.

In afterwards, we will focus on this evaluator optimizer pattern and ultimately we will focus on these two workflows used to edit the article or the select text, plus exposing everything as an MCP server and plugging them into a CL application and cursor to see how we can introduce this human in the loop and replicate this experience that that I showed you here. So yeah, I hope you learned a lot in this lesson.

Sneha Mehra (01:09:27)  
probably this is this is the one that excited me the most. That's why I I made it so long. We learned a lot from implementing this this this project, from iterating many, many, many times over it. ⁓ And ⁓ yeah, I hope you did too. ⁓ And see you in next lessons.

—------------------------------  
3  
Sneha Mehra (00:00:00)  
A major leap in AI progress is the rise of agentic reasoning, an approach that enables AI systems to plan, execute, and adapt like dynamic problem solvers, like humans, basically. It allows these models to break down complex problems, gather information, and respond in context, ⁓ all while autonomously learning and refining their approach. Traditional LLMs could not do this kind of work.

Give them a complex task like research edge AI deployment and write a solid report, and they'll often treat it like a single writing prompt. It will do as much work as with a super basic question. They will basically start answering right now and draft something that sounds structured but isn't grounded or verified. The problem isn't intelligence in the abstract. ⁓ It's that there's no built-in habit of planning, checking, and adjusting mid flight.

Mid-execution, mid-prompt. The first fix people discovered was chain of thought prompting. You tell the model to think step by step before answering. For larger models, that simple instruction can noticeably improve performance because the model first lays out a plan, search recent papers, skim abstracts, compare claims, and then synthesize. And for our research assistant, the prompt would sound like: before answering, think step by step about how.

How you'll research and verify sources, then write the report. But this still wasn't enough, because in a single pass chain of thought setup, the model's thinking and its final answer come out as one uninterrupted stream. There's no pause to run tools, fetch fresh data, or verify claims before the model moves on and gives a final result. As a result, the model can write a plan and immediately draft a report without ever executing it.

Engineering-wise, this also makes the output hard to control and parse. Because there's no separation or interrupt boundary, you cannot build an iterative loop where the LLM executes a step, observes the result, and uses that observation to guide the next step. Tool calls can't be executed, results can be inspected, and claims can't be verified before the answer is produced. To enable verification, tool use and adaptation.

Sneha Mehra (00:02:22)  
We need to use separate planning and reasoning from the action-ansoring phases. That separation is the foundation for two core patterns you'll see everywhere: React and Plan and Execute. React is a short for reason and act. It was designed to blend two things LLMs used to do separate: thinking through a problem and taking actions in the world. Humans do this naturally.

You decide what to do, you do it, you see what happens, and that new information changes what you decide next. That's React in a nutshell. The agent cycles through thought, action, observation ⁓ over and over until it can finish a task. For example, when a research assistant is asked to research AJI deployment, it starts with a thought like I need recent, trustworthy sources.

So I'll begin with a targeted search across credible academic and industry sites. That thought triggers an action. It runs the search. The observation is what comes back. Titles, links, and a short list of candidate papers. Now the agent thinks again using that observation. It thinks something like: I should filter for more relevance and credibility, pick a few strong sources, check the venue, and confirm dates. That thought leads to the next action.

Fetch and extract the abstract and metadata. The observation is the extracted summaries and structured details that the agent can actually compare. With those in hand, it thinks, now I can synthesize. Let me compare claims, especially adapting numbers and flag inconsistencies. Then it acts again, running a compare and summarize step. The observation here is the conflict. One paper reports a 40% ⁓ adoption rate.

Another says 25\. And here's where React really matters. That conflict doesn't get ignored or hand-waved. ⁓ It becomes the trigger for the next loop. The agent thinks again that it needs a tiebreaker source. So it takes an action by searching for credible market analysis and observes a stronger third party report. At that point, the final thought is basically I can now resolve the discrepancy and produce the report. And the agent finishes by generating a structured write up.

Sneha Mehra (00:04:41)  
With citations, including a short explanation of how it handled the conflicting statistics. React's big advantage is that it can steer as it learns. Each tool result becomes a new context so the system can pivot, verify, and recover from surprises ⁓ instead of charging ahead on assumptions. The trade-off is that this loop can become expensive and slow when the task doesn't actually require much adaptation.

That's where plan and execute fits a bit better. It's designed for tasks with a predictable structure where the sequence from input to output is mostly known in advance and doesn't depend heavily on intermediate results. A writing pipeline is a good example. Outline, draft sections, render diagrams, extract images, and then compile the final article. Plan and execute splits the process into two phases.

First, the system generates a complete plan, and then it executes that plan step by step. This is often more efficient because it reduces how often you need to call the most powerful reasoning model. You use a planner, typically a capable LLM, once to produce a detailed order checklist, then an executor or a worker, which can be a lighter weight agent or even a simple loop that runs tools and carries out the steps.

You only go back to the planner if something breaks badly enough that the plan needs to change. That separation tends to lower latency and cost compared to re-reasoning after every single tool call. In plan and execute, you can think of it as committing to a route upfront. Just like looking at Google Maps before going, then driving it fully rather than stopping after every turn and replanning.

One quick clarification before we go further is that plan and execute can sound similar to orchestrator worker. But the key difference here is where the orchestration lives. ⁓ In orchestrator worker, the developer wires together workers with external rules. ⁓ One worker searches, another summarizes, and another writes. ⁓ In plan and execute, the model generates the plan and the system executes it. Both have a planner.

Sneha Mehra (00:06:57)  
And both have executors, but in Orchestrator Worker, the coordination logic is written outside the model. While in Plan and Execute, it's generated by the agent ⁓ as part of the reasoning pattern. Now let's take the same research task and see how Plan and Execute would handle it. You need to first prompt the agent to create a complete plan to fulfill the request. It will output a structured list of tasks like this. Then comes execution.

The system works through step one, then step two, and so on, storing outputs as it goes so each step feeds the next. The system re-enters planning only if something major goes wrong, such as no relevant sources being found, or a step failing completely, or a downstream step can't proceed because the expected input wasn't produced. That's also the main risk of plan execute. You are committing to a plan before you've seen the messy parts in reality.

It's efficient when the world behaves as expected, but it can break when assumptions don't hold. That's why this pattern works best when you have a stable pipeline and clear interfaces between steps. And it's also why we usually want to build in safeguards for error handling and replanning. Choosing between react and plan and execute comes down to three constraints: uncertainty, task structure, and latency and cost budget.

If you expect surprises, conflicting sources, missing data, or weird tool outputs, you want an approach that can adapt midstream, like React. If uncertainty is low and the steps are unlikely to change, you can use a more committed approach like Plan and Execute, where you decide the route up front and then follow it. Now let's bring the task complexity into it. If the task is small and simple, React often wins on speed.

Because stopping to build a full plan beforehand, ⁓ as in plan and execute, ⁓ is just a lot of extra overhead. If the task is complex but predictable, same steps every time, plan and execute is usually faster and cheaper because one good plan up front replaces a lot of repeated reasoning along the way. But if the task is complex and unpredictable, React tends to win again because the ability to adapt is the whole point here.

Sneha Mehra (00:09:20)  
With plan and execute, you are more likely to hit surprises that force replanning, then that's where the cost and time pile up. Cost and quality can trade off too. ⁓ We ran a little comparison on a structured data task of analyzing a CSV and generating a report, and React was slightly cheaper, around 6 to 9 cents, but less accurate, around 85%, while plan and execute cost more, 9 to 14 cents, and achieved slightly higher accuracy, around 92%. ⁓

But that's just one example. In practice, you rarely pick one pattern and stick to it in any case. The most useful systems borrow the strength of both. You use a structure to stay on track and use loops to handle uncertainty when it shows up or when you expect them to show up. That combination is easiest to see in what people often call deep research workflows. Deep research systems are built for long horizon tasks.

Like market analysis, literature review, or financial diligence, all jobs where one model call is never enough because the work spans many sources, many steps, many tools, and a lot of verifications. ⁓ Instead of treating the prompt as write a report, the system needs to first turn it into a set of sub-goals. So if the task is to analyze the competitive landscape for quantum computing,

It breaks that down into things like identify the key companies, collect recent research, scan patents, track recent announcements, and summarize sentiment and momentum. Once that high level structure is set, the system executes step by step. But inside each step, it behaves more like React. It searches, pulls data, compares sources, and checks claims. If it hits something messy,

or unclear, like conflicting statistics that we had previously, it doesn't just pick one and move on. It loops locally, finds a tiebreaker source, and only then continues. So the mental model is simple. Plan execute gives you the global scaffolding, what the sub goals are, in what order to tackle them, and what being done for a task really looks like. And React style loops handle the local uncertainty

Sneha Mehra (00:11:36)  
Missing data, conflicts, and verification inside each subagent or subgoal. Some deep research systems scale this further by running multiple research workers in parallel under a lead coordinator. So the different subgoals are handled at the same time, and the coordinator stitches the results into a single final cited report. Deep research systems are basically React and Plan Execute stitched together at the system level.

What's changed recently is that the same IDs are showing up inside the models through what we now call reasoning models. Reasoning models are models that can spend budget on planning and, in many cases, interleave that planning with tool use. Instead of relying entirely on external orchestration frameworks, some of the reasoning structure is built into the model's behavior and the API's exposure. In practice, you'll see two common reasoning styles show up.

The first is think, then answer. The model takes the prompt, does an internal planning phase, and then produces the final answer. Some APIs let you control how much thinking it does, like setting a higher or lower reasoning effort. For example, OpenAI exposes a reasoning effort setting from low, medium, to high to trade off speed and cost versus depth. Cloud exposes something similar through Extended Thinking, which enables thinking and sets a token budget for it.

Which is really interesting. The second is interleaved reasoning, where the model alternates between thinking and acting. It can decide to call a tool, read the tool result, then think again and repeat until it's done. That's the React loop embedded into the interaction. Thought, action, observation, thought, and repeat. This is also where tool calling becomes much more natural. Tool calling is still a back and forth between your app

And the model. You send tools, the model asks to call one, you execute it, then you send the results back. But the model can handle the decision of when and why to use each of your available tools much better when it's in a reasoning model. So models are evolving and that's really cool. But what's driving this shift in reasoning capability? ⁓ A big piece is training. Many reasoning focused models are post trained using reinforcement learning to improve their ability.

Sneha Mehra (00:14:02)  
To produce, correct, and verify solutions and to choose tool use strategies that actually work. Okay, that's cool. But what does it mean for you as an agent developer? It doesn't mean you stop caring about patterns like plan and execute or react. You still need them as mental models for how the system should behave because you are still responsible for the environment around the model.

Tool design, tool permission, stop conditions, evaluation monitoring, guardrails, etc. What's changing is where the boundary sits. More of the which tool should I use next, logic can live inside the model, while your job shifts toward making sure the tools are well scoped, the objectives are clear, and the system is safe when the model makes a bad call. ⁓ I have a very simple example. If you ask, check the weather in Paris tomorrow, if it's sunny, book me a flight.

Even with a strong reasoning model, you still need to provide two things: a weather tool and a booking tool. The model can decide the sequence, check the weather first, then book, but your system has to enforce constraints, like don't book without confirmation, don't spend above X amounts of dollars, and log everything. Tool intelligence doesn't remove system responsibility. It just changes how much manual orchestration you have to write yourself.

or Cloud Code has to write for you. When you build around LLMs, you are not really adding intelligence to a product. You are deciding how much uncertainty you are willing to tolerate. A workflow tries to eliminate uncertainty by forcing the world into a predictable shape. ⁓ React accepts uncertainty and manages it step by step. Plan and execute bets that the shape is stable enough to commit early. Reasoning models just change where some of that thinking happens

Not whether you need it. So the real skill isn't choosing the fanciest architecture. It's choosing the smallest amount of structure that keeps the system honest about what it knows, what it checked, and when it needs to keep going. As always, we want to use the simplest system that works for our task.

—---------------------------------  
11  
Sneha Mehra (00:00:00)  
Hello everyone. In this lesson, I want to show you how to create a dataset for offline AI evaluations. So, as usual, we will start with a more theoretical section where I show you how conceptually how you can create this dataset. Next, we will move to a notebook where I will show you how we can actually do this and evaluate the Brown writing workflow. So in this lesson, we will evaluate only Brown.

But because I want first to like build out the foundations, you will have the knowledge and transfer these skills, for example, to evaluate Nova or other custom agents. ⁓ And ultimately, I will also show you how to actually see this dataset in OPIC ⁓ and hook it to an LMOps tool. So yeah, let's start with the theoretical part. Okay, so.

I want to start like with a simple basic question and actually explain you what's data set, right? So for for people who come from like data science or more like machine learning backgrounds, this is an obvious question. ⁓ But for people who come from software engineering or maybe even data engineering, maybe it's not that obvious on how you actually see a data set for evaluations, right? So

Usually, like in its simplest form, a dataset ⁓ consists of multiple samples. So this is ⁓ it's basically a list of items if you see it very in a very abstract way, right? So every sample within this dataset contains usually the following items. In one form or another, you will see these four items in a dataset sample. So you usually have the input to your system.

The context that you pass along that input that kind of ⁓ can change right the perspective of the of the input. Then you have the expected output, which is usually also known as the label ⁓ or as the ground truth. And then you have the generate output, which is not necessarily part of your data set, but you need this generated output to actually do something with your dataset, like compute metrics, right?

Sneha Mehra (00:02:30)  
So this is a very abstract, so I won't insist more under its form, but I want to show ⁓ show you how you actually can put this to work and how these four elements are used to actually do AI evaluations in concrete systems, right? So the next logical step is to answer a question such as how can we actually use these four elements, right? These four elements to compute metrics ⁓ or

when you build an AI agent or a workflow or any other AI application for for that matter, right? So ⁓ in reality, AI evaluations are are very omnipresent into the whole AI engineering, machine learning, data science field. So in reality you can actually leverage all these concepts in all these areas and these skills are are ⁓ usable regardless of what model you use or or or regardless of how

develop your application because in reality your AI valve stack is sits on top of your application or or or on top of your model. So ⁓ it's an application in itself, ⁓ which means that it is completely independent from the rest of the application. So it doesn't really care ⁓ what model you use or or how your system looks like. It actually cares only about the inputs and outputs. So

This is a very important to understand, right? So in reality, if you actually want to compute metrics on top of a data sample that that has these four elements, we do the following ⁓ process, right? So we pass the input and context, right? So we pass the input and context to our AI app. So we don't really care what AI this AI app is. As I said, we just pass it ⁓ into it and we treat it as a black box, right? And then we get this generated output.

And here we care a lot about this generated output, but not necessarily about this AI application. And then we compare this generated output with our expected output. So this is like basically the core idea between AI evaluations, right? So you need your input based on which you generate your output, and you compare this generated output of your application with the expected output.

Sneha Mehra (00:04:57)  
And based on this comparison, you can compute ⁓ all sorts of metrics. Right? So this is very similar to for, for example, how data set looks look for training ⁓ or evaluating, as I said, your particular model during training. And yeah, they're they're very similar concepts used in different scenarios. ⁓ And another question that I want to answer here is: what if we don't have this expected output? ⁓

In many scenarios, you might not have it because to create this expected output, you need to go through a process of labeling, which is very ⁓ time and resource consuming, right? And that's why there are like some smart tricks, especially on the EIE vals ⁓ realm, where you you can actually leverage the generate output and compare it, for example, with the input or with the context, right?

Which you would always have because if you don't have the input and context, you you cannot generate the output. So this is a lot easier to gather than the expected output. So basically, these these are like the two core ways in which you can actually compute metrics using your data set. You have your generator output, and you either compare it against the expected output or against your input and context. But again,

This is still pretty abstract, right? So if this is the first time you're you're doing A vals, ⁓ it can be quite tricky. So I promise that as we go deeper into our Brown use case, things will get a lot more ⁓ concrete and easier to follow. So the next step will be right to to map this into our Brown use case. So basically this can be mapped to any AI application. And when we think

About our Brown application, when we want to translate that to our Brown use case, we have the following setup. So our input is the article guideline, right? So everything relates to the article guideline. And based on that, we have a research file, which is ultimately our context. We pass that to Brown, we generate the article, ⁓ and that's it. Like ⁓ at a very high level, this is our application looks like, right? And we if we want to compute metrics.

Sneha Mehra (00:07:19)  
We have an expected article that's also generated based on this input, right? So we have this article guideline and in this research. And while we create it, we have in our mind how our ideal expected article will look like. And this is basically it, right? This is what we expect from brow to output. But in reality, they they they never like generate exactly what you what you expect. And that's why

you do this comparison, right? So basically you compute this metric where if the generated article is exactly as the expected article, you have like a big metric. And as they go further and diverge from one another, you have like a a lower metric. So that that's a basic idea. And that's how you can very easily optimize and improve your system because now you have a clear way to quantify the quality of your output of your system, right?

And again, if we don't have this expected article, because you not always have the ideal article because that's a lot of work to put into, right? You you you you can do some tests also against the article guideline, for example. So you can compute metrics where you want to check that the generated article actually follows the article guideline one on one, right? Because may maybe it will not follow the bullet points ⁓ or the narrative ⁓ as you instruct it over there.

Or if it contains only information from the research, right? Because you want to make sure that the generated article doesn't hallucinate and contains only facts that you have within your research. And again, these are very important signals that in reality you will you will do both. You will also compare the generated article against the expected article, but also ⁓ against your article guideline and research.

And we will see this in in action actually in our in another lesson where we will compute the metrics using a similar strategy. But for now, I I want in this lesson to focus just ⁓ on how we will actually create the data set that will later on be used for the EI evaluations, for the offline AI evaluations per se. ⁓ So before going into the code, ⁓ one last thing that I want to show you is.

Sneha Mehra (00:09:45)  
How can you actually gather this data in reality, right? Because we have this this guideline, this research, this expected article. How can we actually gather these items to create the dataset per se? Because in reality, you will not download this data set from Hideface or from the internet, but you will have to create it on your own. So for example, when you fine-tune a model, you might get away and download some data sets plus.

A small custom data set and fine-tune it on your use case. But the trickiest part is in how you actually create this data set because you want it to very to be very scoped ⁓ on your particular use case. So basically using generic ⁓ datasets won't help you a lot, or in many use cases, won't won't help you at all, right? So yeah, let me go over that. Okay, so we have this section where

I walk you through how you can gather each piece individually, where you have basically the input, where in our brown use case is the article guideline. You have the context, which in our use case is the research, and you have the expected output, which is the expected article. ⁓ And this process is also known as labeling, which you for sure you you heard in one way or another about, right? Okay, so let's dive in into this one. ⁓

Okay, so as you can see from from this graph, the process ⁓ is pretty similar, right? So we have the guideline, the research, we pass this to the Brahm, we get the generated article. And let's focus just on this part for now before for how we gather the input. So the key idea is that this input needs to be as diverse as possible because our end goal is to compute metrics ⁓ in as many edge cases and scenarios as possible, right? And to do that.

You actually need to find as many scenarios that are as close to production as possible. Because ultimately you you don't care to evaluate on scenarios that will never happen in reality, ⁓ but to find as many scenarios that might happen in production. So to get this article guideline, there are basically three ways on how you you you can do that. So the most obvious one is just to manually create it, right? So

Sneha Mehra (00:12:10)  
Sometimes you have no other way than just manually creating it. And for example, in our use case, we were pretty lucky because as we actually created the course and we actually use this product, we actually required to write these article guidelines manually. So in our use case, manually creating it was a viable solution, right? And the second option is usually possible only ⁓ once you deploy the product to production and you start having users.

So, as I said, your end goal is to create a data set that captures ⁓ Smatch edge cases and use cases that actually happen in production, which means that it's very intuitive to actually monitor the production and capture real-world user inputs ⁓ and you use those, right? Because those users actually want to use brown or other AI on those particular inputs. So that's what you should.

test your product against, right? And because we already integrated this with OPEC and we already monitor all the traces, it's very easy, right, to capture these inputs and ⁓ put them in your evaluation data set and actually compute metrics against the the the these inputs. ⁓ And to find edge cases, ⁓ usually you will you will go through a process of error analysis, right? Where you look over the the data inputted by

people and you see usually where a brown or other AI app ⁓ doesn't perform as expected. For example, it it can introduce ⁓ real errors or it can ⁓ introduce more subtle errors where basically brown doesn't perform as expected, right? So these are called like regression tests where you want to capture as many samples where your application doesn't still perform.

Good enough, put them in your evaluation data set and start optimizing your system against these ⁓ samples. ⁓ And based on these samples, you guide your system ⁓ on improving on top of them and getting better and better on use cases that it did not did well in the past, right? And like this, because you keep those samples in the evaluation data set whenever you run the EI evaluate process again, you make sure.

Sneha Mehra (00:14:36)  
That the system still performs well. ⁓ And the third option, if like manually creating the inputs or you're not still in production, ⁓ is to synthetically create these inputs using another LLM. And here you actually don't completely get away with not manually creating items at all or gathering something from production. Because what you do usually you ⁓ have a small sample ⁓ of items that are

Created by humans. And then you is use this as seeds, as few short examples to an LLM, plus some prompting to create, to expand that data set. And this is a very powerful technique when you're in the beginning, right? When you're not yet in production and you don't have data at all. And it's very powerful in the beginning to ⁓ kick off your EIEVOL's flywheel and start this optimization process. Okay. ⁓ So

Enough talk about the input, let's let's move to the next piece of the puzzle, which is the context.

And in in many use cases, ⁓ actually in all use cases, the context is a byproduct of the input. For example, in our particular use case for Brown, our context is the research which is generated by the Nova agent, right? So once you have the input, the article guideline, we get this research ⁓ by calling Nova. ⁓ And that that's basically it. So basically your ⁓

Context is usually a byproduct of your input. And in other use cases, the disease context can be like ⁓ tool calls, basically the results from particular tool calls or from other LLMs, or basically whatever makes sense in a particular scenario based on an input. You want to capture that and put that as context. So always the context will be a byproduct of your input based on and generated by other parts.

Sneha Mehra (00:16:39)  
of your system. In our use case it was Nova, but it can be generated by any other parts of of even the system that you want to evaluate, right? And basically capture a particular context, a particular state of your application when you actually want to to to evaluate it. ⁓ And if you ask me, this is one of the hardest parts to to to solve because like capturing particular states and particular outputs from all sorts of tools.

⁓ or adul LMs and all that can it can get complicated if the system is not ⁓ programmed and engineered properly. And another way to to capture this is again by by monitoring production. Where you usually if you monitor it properly using tools such as OPIC or Langfuse or LangSmint or whatever, you capture the inputs plus the context as well, right? So you can actually see

That for a particular input, you you get that particular context, as we've seen in in the lesson on observability. But ⁓ let me actually quickly open OPIC ⁓ and show this to you. So we have Brown. In the Brown project from OPEC, let's search based on the tag. So we you can do a filter such as like ⁓ we pick the text column contains and we say generate. And then we have all our generate steps, and let's pick ⁓ the second.

So basically we we have here our input, right? So we know that we use the context engineering ⁓ directory.

And then basically in this trace, you you you can gather whatever you want from from from your context. So when we run this generate media items tab, we actually have access to everything, right? To to to the all to all the contacts.

Sneha Mehra (00:18:38)  
So we basically have probably we also have the research here. No, we actually have to search by this up.

Sneha Mehra (00:18:49)  
So we have the article guideline here, right? We have the the research here. And basically all the context is inside inside here. In our particular use case, we haven't engineered this to gather this this context in for EIE valves. But the idea is that you have access to all this information in every trace. And for every particular input, you can gather whatever whatever you need from here.

⁓ And you you can do this for particular tool calls, ⁓ right? As I said before, for particular model calls, ⁓ for particular input you can freeze other elements and do AI evals on particular output.

Okay, so this is how you get your context. And I want to highlight that you don't want to synthetically create this context, right? Because the context always has to be a byproduct of your input, which let's say it was created synthetically using an LLM, but this context needs to represent the reality as a byproduct of the article guideline. Okay, and the next element would be how.

We get the expected output from from these three elements. Because the fourth one is the generated article, which is like created during the process of evaluations per se. ⁓ And then the third item that we need to actually put in our data set is the expected output. In our use case, it is the expected article. So we have our input, our context, we pass that to our AI application, we get the generated output, and then a domain expert reviews.

This generated output, right? So reviews this generated output ⁓ and transforms it into the expected output. ⁓ And you do that because you don't want to write it from scratch, because you most probably your model is already good enough, like 80% good enough, and you want just to adjust it ⁓ and make it the ideal output. ⁓ And this is ⁓ what you can do in most generative AI. ⁓

Sneha Mehra (00:21:00)  
scenarios or or or in many many ⁓ ai applications ⁓ another very similar scenario is where for example you want to use a smaller model right for a application but during this ⁓ labeling step you use very powerful LLM so for example let's say that in production you want to use brown only with flash but while you

Do this labeling and you're more in the beginning, it's fine like to use a Gemini 3.0 Pro ⁓ and you want like to maximize the chance that your generate article is as close to the expected article. But you know that in production you will use, for example, only Gemini 3.0 flash. And this is fine because this generation, this labeling process happens only once.

And it's offline and you don't care that much about latency and costs because then you can reuse this. But after you have this AI Evals dataset, you have like your North North Star, ⁓ right? You have your evaluation dataset, ⁓ and on top of which you can actually optimize your system and you can compute metrics and start the AI Evals optimization flywheel. So again, this this is like another way. Usually you want to put all

more resources into this labeling when you when you run your system. Okay, so let me repeat that. So to conclude, you can either just use the system, for example, Browne as is to generate the article, use the domain expert to refine the expected article, or if you're more in the beginning or your system doesn't necessarily perform that well or you know that you want to like to make it to optimize it and use ⁓

lore models or or things like that. When you do this step, when you create your EIVAS dataset, it's fine like to use the most powerful models out there just to reduce the domain expert effort here, right? Because usually a human doing manual work will always be more expensive than the most expensive lab out there that's accessible to you. In most scenarios that that will be true. Okay, so

Sneha Mehra (00:23:18)  
This is the theoretical part. Now I want to actually go over the code and over our dataset and show you how this looks in our concrete brow use case. Okay, so we have here a notebook. It's again into the lessons repository. ⁓ And we have this creating dataset for AI Evals notebook, right? Over here. To set it up is very similar to how we set up notebooks for from previous lessons.

Just be sure that you actually ⁓ picked up the virtual environment created by UV here here at the root. And then I will just restart it to run this together with you. ⁓ So we have this standard setup step. Then we need to make sure to have the OPIC API key set up, right? So you need to go to OPIC, create an account, put it in your file, and then it will know to pick it up.

We need this because we will load at the end the dataset to OPEC. We do some extra imports for our pretty print. And this is pretty standard what we did so far over the course. Then we also download the configs for brown and the inputs. Here in the inputs, we actually have our EIE valve dataset. We check what's actually in the inputs. ⁓ And we can also see here, right, the configs.

And the inputs. Okay, and here we actually have our evals directory, which we'll use in this lesson. ⁓ And we check that everything is alright. So we have our inputs there and our inputs evals there. ⁓ So everything looks alright. I went really fast over this because it's pretty much a copy paste over all our all our notebooks so far. So nothing interesting. And only here we actually start to dig into the particularities of this notebook.

So basically, the data set is split into two big components. We have this metadata.json. ⁓ And we also have, right? So here after we downloaded the inputs folder, we have the eval ⁓ and the dataset folder. And we have this metadata.json file ⁓ and this data folder with the samples per se. Right? So this metadata.json file consists, well,

Sneha Mehra (00:25:44)  
metadata about the dataset, right? ⁓ And what does that mean? So basically ⁓ we list all our items that we want to gather from this data folder over here. And we specified its name, which is like our unique identifier and also like a way for us as a human to easily understand what that sample is about. ⁓ And the directory where we actually have this

The sample. So basically we just listed all our items that we want in this AI Evelse dataset. And as you can see, for example, in the first one, we have the name, which is our identifier, where it's located on disk, a path to the article guideline, which is our input, to the research, which is our context, and to the ground truth article, which is our expected output. Right? So remember, for the expected output, we also have like the this naming code actions of the label or ground truth is the same.

And we actually wrote a function where we don't necessarily need to specify all of these, they're optional. ⁓ if you use the same naming conventions, right, they're optional. So basically we can just specify the name and directory to the metadata and and that's it. And here we also have this is few shot examples, which will be used for the LLM judges to compute metrics using LLM judges method methods.

And they will be used as QShot examples in these element judges, but we'll learn more on that in the offline metrics lesson. ⁓ So I won't dig more into that over here. Okay. ⁓ And basically, in our samples directory, right? In the this one, where we have all our samples. So as I said before, here we have our dataset directory, here we have our metadata file.

And here we have our data directory where we actually have our concrete data where we have 10 samples for our AIE valve use case. And this is enough for us because an article per se ⁓ contains multiple sections where each section kind of contains its own partic particularities. So, for example, an article with 10 sections can contain one section with only text, one section that's more heavy on.

Sneha Mehra (00:28:07)  
like media data one section that is more heavy on on code. So ⁓ usually like let's say that on average we have an article with like 10 sections. So this can be translated like to 10 times 10, like to 100 scenarios that we will actually test, which is quite a big AEVALS dataset already. ⁓ Okay. And if we go over each directory, you can see that basically every sample has this ⁓ ground truth

Article guideline and research. Basically, basically it has these three elements while the generated output will be generated by Brown when we do the the EI Evalve, right? So we have the article guideline, research expected article, while the generated article is generated on demand. Okay. So basically, this is our data set. We actually used our lessons from the first part to create our eval dataset.

And that's why this is actually an EIVAS data set that reflects reality and it's not something really mocked that's super ⁓ specialized and synthetic and is not that useful. So you will actually see how a real world project looks like and how the metrics after you you test it on real world data look like and how they're really far from perfect, right? And that's perfect perfectly normal. ⁓ every ⁓ a project starts like this.

⁓ And actually before ⁓ going to our last section from from this, I want like just to show you that basically here we have our article guideline, which is the same as in our demos. So we this is actually the the input that we use to generate the lesson. Then we have our research, which is just the output from Nova. So it's nothing special. That's why I won't go into in all the details because we are already used to this.

And we have the article ground truth, which is actually lesson five that you read in part one. So we just took that lesson and exported it in Markdown and put it here. And that's it. ⁓ so we already went through all that effort, right? Of through all this effort, right, where we wrote the article guideline, we got the research, we joined the article and we did this huge domain expert effort of refining the expected article to look our ideal article.

Sneha Mehra (00:30:33)  
would look like like in a professional manner.

Okay, so these are not again, these are not mocked articles, these are like the real real thing. Okay. So this is our data set looks like you're already familiar with it from previous sections. And now let's see how we actually like model this into Python and upload it to OPEC. So before going into this code, let me actually show it to you how it looks like in OPEC. So we go to OPEC.

We go to so this is like the dashboard, and we go to the dataset sections, and we have this brown course lesson dataset. ⁓ And let's open it. So we have this ⁓ each one of these is like a sample, which in our use case is a lesson, right? So based on the name column, you can very easily identify it, and that's why we chose it like to be both an ID and like a humanite way to ⁓ look and understand.

Which is which. And here in the columns, we can actually like simplify it a bit. So let's say that we don't care about this one, about this one, about this one. So we actually don't care about many of these columns at this time. So this is our data set. We for example, we miss lesson four and lesson seven because we use them as few shot example for the LLM judges, right? As explained here in the

in the metadata. So if you go to four, no, not here, my bad. Here into the metadata JSON file. If you go to to lesson four, it's flagged as is few shot example true. And lesson seven is flagged again as is few shot example true. So we don't upload that because we will actually use it to build our LM judges, which we'll see in next lessons. Okay. So this is our data set and we can we here in OPIC we can actually like open

Sneha Mehra (00:32:34)  
Open one sample and you can see everything about it. This is a beauty of having like a GUI ⁓ for this. So we can see all our media items that are within the research. We can see the article guideline, which is ultimately a markdown file, right? ⁓ And then you can see the directory, the ground truth article, if it's a few shot example or not, the name.

the research as context and that's it. So basically everything that was as a column before is here ⁓ as a dataset element.

Okay, and next you could use this dataset to compute experiments, compute metrics, and so on and so forth, which we will do in future lessons. But for now, this is our final outcome to have these brown course lessons datasets here in OPEC. So now let's see how we can actually do that. So we go here in section three. ⁓ So we actually like exported the code that we we care about over here. So

We model each sample in a piedentic ⁓ model entity, right? As we did everywhere inside the brown code. So we have this ⁓ eval sample pydentic model, which actually has all the fields that you've seen in in ⁓ OPIC, right? So we have the name, the guideline, the research, and everything that we've seen here in OPIC. We have here present. Next, we create this eval dataset pydentic model, which basically aggregates all this eval.

eval samples into a dataset. So we have a list of eval samples and also have like a name of and a simple description of the dataset. And here in the load dataset method we ⁓ read that metadata json file and then iterate over the directory, right? Over this directory and basically just load all of this ⁓ data into eval samples.

Sneha Mehra (00:34:44)  
⁓ And create ultimately this python model, right? So that that's it, nothing so super complicated. And this load markdown file method just reads stuff from markdown file, right? As the name suggests. Okay, and the next few functions that are really interesting or relevant, let's say, ⁓ are the upload dataset method, which takes this eval dataset object that we just created here.

and filters out the Aval samples that we want to actually upload based on this ease few shot examples flag and keeps the training samples ⁓ one locally. Basically it doesn't upload them, it just prints that to the to the console output. So when you run this you are aware that hey we actually did not upload this to the OPEC. And then we have this custom update or create dataset function in our OPQtils module.

⁓ that we highlighted here that basically takes the items of the dataset, the description and the name. Again, we model this ⁓ one-on-one with this eval dataset by the object, but here we pass them as individual attributes because this is how OPIC expects them. Also here in the upload dataset, it's important, probably the most important line is here where we take the evaluation dataset and dump it.

Into a JSON serializable ⁓ dictionary. ⁓ And this is how we actually pass the EVA samples to to opic, not as a Pythonic model, because it doesn't work with Pythonic models. Usually every time you pass something from your Python application to an API, you kind of need to translate that to JSON, right? Okay. ⁓ And yeah, that's that's basically it. So as I said, it the code is very simple.

What we do is we just ⁓ take this data from here from disk and pass it to this function, which ultimately what it does, it uses this opic client ⁓ and it has this get or create dataset where the name of the dataset is like an ID. So if we already have it, we just get it. If not, we create it. We clear everything that we have in it and we insert the new items to it. So we did that out of simplicity for our use case.

Sneha Mehra (00:37:11)  
where ⁓ to avoid having duplicates because we we we kept changing and adapting our our our samples. We just wanted to clear everything and insert from scratch. This works because we have like a very small dataset. So this operation of cleaning and inserting back again is very light, almost instant. But ⁓ some next steps will be to have a more more careful ⁓ eye on how you want to insert new items and don't delete what what you had.

before. So this will be like a next step of on actually improving this this method.

Okay, and the last section of this ⁓ of this notebook is on actually running ⁓ what I showed you before, right? So so we have our eval dataset here, this directory. We scope down everything to the dataset directory, right? We call our dataset name ground course lessons, as we saw here in OPIC, right? ⁓ And this is a simple dataset description.

Like the Brown evaluation dataset on course lesson format. This is important because you can actually evaluate your your AI application on multi formats, right? So let's say that we want to adapt Brown also for social media posts, right? So you might want to have a different data set for that as well because ⁓ it's ⁓ a different problem, right? So you you need to optimize the output in a completely different way. So you can actually have multiple data sets for the same application.

So okay, so let's actually call and load our dataset in this eval dataset pyden model.

Sneha Mehra (00:38:59)  
Right, so as we've seen before, we have a dataset called Brown Course Lessons. We have this data set description, we have ten dataset samples, ultimately loaded from from here.

And then we call called the upload dataset ⁓ function. If we look here into the dataset, it says that it was created ⁓ on the on this date. But we call we clear and insert this is like actually the first time I created it. So just for this example, let me delete it, delete it ⁓ and recreate it from from scratch.

Sneha Mehra (00:39:44)  
Right? So we have it here back again, and it was created in today's date. Okay. ⁓ And yeah, that's that that's that's basically it. I think some ⁓ nice next steps on what you can further do and further improve your understanding on dataset is to actually extend the dataset with more diverse samples, or even change the dataset with a new set of articles or

On a completely new format. For example, if you adapted ground to generate social media posts, it would be nice to exchange the data set to do ⁓ evals on social media posts. And also remember that I said that this method needs to be improved and not run this clear insert ⁓ step. ⁓ So you could further improve that method ⁓ and to not do that clearing, but just add on top of it.

And a byproduct of that would be to actually add versioning on your data set. ⁓ And I want as a final idea on this is to realize that the code here is very basic. There's there's nothing complex about it. But the real hustle, the the the real struggle is it actually gathering your data, right? Gathering your inputs, gathering your research, and gathering your expected article.

and creating the data data set itself. ⁓ That's the most complicated part over here. Then how you structure it and load it and work with it. It's pretty easy. And that's why ⁓ when you actually download a dataset from Huggy Face or other platform and you use it ⁓ in a notebook is a bit of a bit of cheating because it doesn't show the complexity of the real world at all. So

It's just like the hello world of EAEvels when you do that. So this is a pretty similar scenario because you haven't gone through the hustle to the pain of actually generating this. So that's why I strongly suggest you that if you want to understand how this works, e is to go through the steps ⁓ and try to create an article guideline, generate lesson with it, like

Sneha Mehra (00:42:03)  
And that's why I strongly suggest you to ⁓ go through the steps, for example, right? Write an article guideline, do the researcher Nova, generate an article, tweak it until you like it, generate the expected article, and extend this data set. Do this ⁓ on generate articles or social media posts or whatever other use case, and that's when you will really understand how how this process looks like.

And yeah, I hope you you learned a lot about Avalations and creating datasets. ⁓ And see you in the next lesson.

—------------------------------------  
1  
Sneha Mehra (00:00:00)  
Welcome to the first video of our new course, Agentic AI Engineering, built in collaboration with Tour ZI and Decoding AI. I'm Louis François, CTO and co-founder of Tour ZI, and you are about to gain a competitive advantage that will redefine your career. Why? Because you'll soon master the creation of large language model workflows and autonomous AI agents capable of reliably automating complex real-world tasks.

This new specialized AI engineering course moves you beyond simply calling an LLM API, transforming you into a real AI engineer who designs, deploys, and maintains sophisticated AI systems trusted by users. If you decide to enroll, or if you have already, over the coming weeks, you will build, evaluate, and ship a complete research agent and writing workflow system, gaining a professional AI engineering certification.

That validates your new expertise. But even better, you'll graduate with a robust project portfolio showcasing a real deployable multi-agent system that you can even use yourself for researching and crafting anything from blog posts to technical reports, emails, and documentation. Super useful in my daily work. After completion, you'll also gain lifelong access to our private Slack community connecting you with expert instructors from the Tour ZI team member.

And the decoding AI team members and successful alumni who have already launched thriving careers or startups in AI. We really focused on teaching the engineering skills to create production-grade AI systems that deliver real business value. Our approach mirrors how we build agents in the real world for our clients, iteratively, pragmatically, and with a focus on what actually works, not just hyped approaches.

While we aim to cover the fundamentals, our courses do not focus on endless theory or fragmented toy projects to build superficial demos. We, in fact, built this course with a single project in mind, teaching the student to replicate it end-to-end. We've been focusing on AI education since 2019, now teaching over 500,000 learners globally. But we are not just educators, we are also builders.

Sneha Mehra (00:02:18)  
Our teams have spent the last four years building LLM applications from our AI tutor to automated job boards and many custom agents for our clients. And you are about to learn to solve all challenges we've encountered when building these systems, from the mistakes we've made to the best practices we've developed about agent tick AI. ⁓ All into this single course. Before starting, let's get real about what LLMs can and cannot do on their own. At their core,

LLMs understand, prompts, and generate remarkably human-like text. This seemingly simple capability triggered unprecedented hype, as you know, as well as genuine and explosive adoption. In just the past year, advances in reasoning enhanced models like GPT-5 alongside increasingly sophisticated agent tech products such as Deep Research and coding agents have dramatically accelerated adoption.

OpenAI's revenue surge as the number of ChatGPT users increased, reaching nearly a billion users. Nvidia's dominance in AI GPUs has catapulted it to become the most valuable company in the world. Google's Gemini model, Family Alone, processed an incredible 980 trillion tokens just in June 2025\. Enough data to read the Lord of the Rings series 1.5 billion times. That's a hundred times growth.

From just over a year prior. Yet, despite all this, the raw capability of LLMs to simply generate plausible sounding text on their own and without additional code infrastructure falls short of solving real-world problems. Yes, even GPT-5. For example, if you ask an LLM alone to find a flight from New York to Paris next month with a hotel under $400 per night, the LLM will provide detailed flights, hotel names, prices, and availability.

It sounds impressive, but none of these details will actually be real. Without integration into booking, APIs, real-time data access, internet, or persistent memory, the LLM is simply guessing based on patterns learned during training. It predicts the next token. That's it. Even reasoning models only predict the next token. When you follow up to request cheaper options, compare prices across different sites, or adjust travel dates, the problems with barebone LLMs become even more apparent.

Sneha Mehra (00:04:41)  
The LLM will forget the original request, repeat previously answered questions, and confidently present outdated or entirely fictious information. It cannot dynamically verify availability, proactively pivot to alternative booking options, or reason effectively through real world constraints as your travel planner would. This example illustrates the core limitations of raw LLMs. Without additional engineering infrastructure, ⁓ LLMs produce plausible sounding text.

But not actionable solutions. They lack persistent memory, real-time data integration, multi-step reasoning capabilities, adaptive decision making, and a robust ability to take actions in the real world. They can mimic a travel planner's email responses, but can't mimic their actual work. And so we need someone to make that happen and enable LLMs with all these tools, making them actually viable for real applications. And this person will be you after this course.

A true AI engineer. AI engineers build features on top of models. Even the big labs have AI engineers to release increasingly sophisticated systems within ChatGPT, Cloud, and Gemini, which integrate features such as prompt chaining, built-in memory for conversation persistence, agentix search via web tools, image generation via other models or the same one, and sandboxes for code execution. ⁓

Even these complex products often fall short of the reliability, capability, and customization needed to perform specialized tasks across industries. They cannot do everything from a single general application. Central AI labs simply lack the proprietary data, domain expertise, software connectors, time, or granular problem knowledge to tailor solutions for every customer or enterprise need. Off-the-shelf systems.

Can't fully anticipate the quirks of each company's datasets, use cases, and edge cases. This is where our work begins. The AI engineer doesn't just prompt these models. They build complete systems around them. They create the infrastructure that gives an LLM memory, allows it to use tools, and enables it to execute complex multi-step tasks. They also decide where human expertise needs to be brought in the loop.

Sneha Mehra (00:07:00)  
They transform LLMs into reliable agents that act in the world. This is a distinct new role that is different from both software developers and machine learning engineers. Unlike ML engineers who train models from scratch, AI engineers work with pre-existing foundation models. Unlike traditional software developers who write deterministic code, AI engineers design systems that gracefully handle the inherent non-determinism of LLMs.

To deal with deterministic coding infrastructure. The same prompt might generate different responses each time, requiring a fundamentally different engineering approach that embraces iteration, experimentation, and scientific mindset. Yet, the software around the LLMs still expect deterministic communications. This bridge is where AI engineers lie. The modern AI engineering stack comprises three interconnected layers.

But more importantly, it requires a specific set of skills that we will teach you to apply strategically throughout this course. First, the application layer. This is where AI engineers bring systems to life through workflow design, tool integration, evaluation systems, and user-facing interfaces. They master specialized skills such as retrieval augmented generation, or RAG, advanced prompting, context engineering, strategic model selection.

Data collection and data engineering. Then the model layer. This is provided by AI Labs training huge foundation LLMs. AI engineers interact with this layer via fine-tuning and strategic model selection skills through APIs, for instance, navigating trade-offs between different model types, sizes, and capabilities. Lastly, the infrastructure layer provides the foundation through tools such as cloud services, API endpoints.

Vector databases, frameworks, and observability systems. Don't forget this last one. While some of these skills are explored in greater depth in our foundational course Full Stack AI Engineering, especially anything RAG related, this current course focuses explicitly on turning these techniques into autonomous AI agents and workflows designed for real-world deployment. By the way, AI engineers also generally use Python, or in some cases, TypeScript.

Sneha Mehra (00:09:20)  
So a programming basis is a requisite. If you don't yet know how to code or aren't confident with Python, you can also take our beginner Python for AI engineering course first, which I'd strongly recommend. Otherwise, if you are good to go with Python, let's keep going. Beyond these technical skills, AI engineers also need to develop something equally valuable the intuition for integrating domain expertise directly into their systems. Generic LLMs lack understanding of specific industries.

So, you need to embed domain knowledge through careful prompt design, data set choice, validation processes, and UX design. We'll also guide you through the product thinking aspect of the AI engineer role. Since the cost difference between models can be thousandfolds, you'll need to match solutions to business value. Always asking what tangible outcome justifies this approach. Essentially, you'll develop the intuition for what to try next. You'll learn when to improve prompts, when to add RAG.

When to consider fine-tuning, when to build agents instead of workflows, or even when it's good enough based on actual business needs, and learn when it's time to stop developing. As you may know, it's relatively easy to get an impressive demo together. But it takes increasing amounts of work, knowledge, iteration, and complexity to solve more and more edge cases, reduce failures, and get the system ready to deploy in the world. ⁓

To directly answer a popular redundant concern of am I too late to become an AI expert, I'd like to highlight that a key reality of this field is that there are no true expert AI engineers yet. As the models and techniques have only existed for a few years and are still evolving week to week, this creates a massive opportunity for you to take an early lead and become one of the first experts in this key new field. We've seen why AI engineers exist.

Basically, to fix LLM's limitations and integrate them in real products. But here's why it's time to jump in and lead in the AI engineering space. This year, we have seen the most explosive growth in the application layer on top of these foundation models. All the tools building on top of models like Cursor, Perplexity, and Base44 just exploded in valuation. But these aren't simple wrappers around LLMs.

Sneha Mehra (00:11:42)  
There are sophisticated systems that combine models, data, tools, and careful engineering to solve real problems reliably. The current AI Gold Rush is about building the infrastructure that makes these models actually useful. And this is where all of us can create value. Custom agent and workflow systems are measurable business assets. They increase revenue through automation, reduce costs through reliability, and create entirely new product categories.

That simply weren't possible a year ago. The engineers who build these systems will become indispensable. Before we let you loose to build AI agents on your own with our course, let's be clear about what happens when you give an LLM autonomy. The results can be spectacular, but they can also spiral out of control. And these aren't just hypothetical risks. ⁓ We are already seeing documented disasters from real deployment. Recently, a coding agent from Replit

Had a simple task to do under a code freeze to prevent accidents. Instead of seeking clarification on an empty query result, it panicked, ignored the freeze, and proceeded to delete the entire production database. It lied to its human operator, confessing only when cornered. Fortunately, the user was able to recover their data in this case due to safeguards built into the system. This is just one example.

But there are tons of failure cases when giving too much control to the agent without proper infrastructure built around them. We give the agent a simple command, but without the right controls, it can flood the entire castle. This is the core challenge of the AI engineer, particularly when building agentic capabilities. This course teaches you to create sophisticated, production-ready LLM workflows and autonomous AI agents through structured hands-on engineering practice.

And how are we doing that? We start by mastering foundational skills with practical exercises, chaining multiple LLM calls to perform complex tasks, implementing conditional logic for routing inputs to different workflows, structuring LLM outputs reliably with tools like PythonTick and JSON modes, and integrating external tools using function calling. You'll also build robust memory and knowledge access systems.

Sneha Mehra (00:14:02)  
Learning RAG techniques for reliable information retrieval and context management. Then, using Langgraph as your primary agent development framework and FastMCP, the most popular Python implementation of the MCP protocol, you'll combine these foundational elements into advanced agentic systems. You'll build agents capable of advanced reasoning and planning by implementing core patterns like React.

or reasoning plus action loops, plan and execute, or complex goal decomposition, and reflection loops, basically self-improvement through self-critique. ⁓ Using Langraph's functional API, you'll learn to orchestrate complex workflows and using FastMCP, you'll learn to expose your AI agent to MCP clients such as Cursor, Cloud Code, or any other custom application. We decided to use Langraph's functional API instead of their Graph API as it's easier to understand

And work with. Doesn't lock you too deeply into the Langchain ecosystem either, while providing most of LineGraph's advantages, such as orchestrating multi-steps, retries, and memory. We think it's the best trade-off between writing from scratch and using an AI framework. But building agents is just the first step. Deploying and running them reliably in production is equally important.

We'll dive into the LLMups discipline or large language model operations, where you'll master custom evaluation frameworks to ensure your agent aren't just impressive demos, but robust measurable systems that businesses can trust. Using open source specialized observability tools such as OPIC, you'll learn to monitor, debug, evaluate, and continuously improve your agent's performance.

You'll optimize production efficiency by balancing costs, latency, and quality through strategic system design, model selection, and careful trade-offs around inference scaling. In production, each user request triggers potentially expensive multi-step model interactions. You'll become skilled at tracking and optimizing these interactions, learning techniques to identify bottlenecks, debug complex reasoning paths, and handle errors gracefully.

Sneha Mehra (00:16:12)  
Comprehensive logging, real-time monitoring, and evaluating your AI system against your key business metrics will become second nature. Our first few lessons will be taught using notebooks, while building components of our central comprehensive production grade capstone project. Building interconnected research and writing agents using FastMCP and Langgraph, which will then be hosted in your own GitHub repository and have a final UI to showcase it. At the end of the course,

Your research agent will autonomously accept any user-provided topic or question, identify no-less gaps, formulate targeted research queries, and autonomously explore the web or APIs, automate elements of data engineering, including gathering, scraping, and synthesizing relevant information into structured actionable notes, and then your writing agent.

Will ingest the multimodal structure research notes from your research agent, transform the raw research into polished production-ready content, follow detailed user-provided input guidelines, adapt outputs dynamically based on user-specified requirements, blog posts, reports, social media content, generate supporting visuals and graphics automatically, implement self-correction loops to autonomously review, critique, and revise drafts.

This modular architecture provides exceptional flexibility and extensibility, enabling you to customize and integrate your agent systems into any workflow or industry scenario. By the end of the course, you'll have built a professional quality AI agent system suitable for showcasing in job interviews, impressing stakeholders, or even launching your own AI initiatives. To support your certification and final project submission, we provide you with a complete professional grade.

Template repository, which allows you to either extend our sophisticated agent pipeline or innovate and create a novel solution ⁓ uniquely suited for your interests and professional goals. You don't have to replicate what we did. You can always diverge from it and build around it. This course is your structured path from cautious LLM API thinkerer to a confident agent builder architecting systems that solve complex real-world problems.

Sneha Mehra (00:18:28)  
You'll join a community of practitioners building the systems around AI models that will allow them to actually start to deliver on their hype. Let's transform you into an AI engineer.

—-----------------------------

5  
Sneha Mehra (00:00:00)  
Hello everyone, I want to shoot a quick video on showing you how to run the Nova and Brown agents together, right? ⁓ So before going into the actual lessons and explaining everything on how they work on and how they are architectured, I want to like quickly run them and build up an intuition of what's the final outcome ⁓ of those capstone projects, right? So

Like this, you'll have a strong feeling on how they work ⁓ and you will know what to expect at the end. So it will be easier for you to be anchored into the final output and the final expectations of the capstone project while we actually go through the lessons and explain the code. ⁓ So to do so, ⁓ you will have to clone the repository with all the lessons, right? So

You will follow the standard instructions from the admin lessons. There's nothing specific to accessing those projects. But if you go into the lessons folder, instead of going through the actual lessons per se, we'll have this research agent and writing workflow folders that contain the whole project. Basically, the research agent contains the Nova project and the writing workflow.

folder contains the broad project. They're independent agents that communicate through some files and we will see that in action right now. So before running the Nova agent from the CLI, I want to quickly show you how to install it. So we have the MCP client and mcp server folders, right? Both are managed by UV. So I already installed them but for example

When you install a project through UV, it creates those VM folders. And let me show you how how to actually do that. So let's start with the MCP server. So we move to the server and we run just UV Sync. And now it will install all our dependencies and create this.vm folder which actually contains a virtual environment of our ⁓ Python project.

Sneha Mehra (00:02:22)  
And if we do the same thing for the C client.

We run uv sync. As you can see, it created a v dot.m folder with everything we need to to to run this. So if we want to activate it in our terminal and always run the code using this virtual environment, we can do something such as source.

Vm bin activate. And now as you can see, it it started using this ⁓ MCP client virtual environment from here. Just let me show you. For example, if we do which Python, which points us to the Python executable that's ⁓ used inside this ⁓ this terminal session, we can see that it's using the MCP client VM bin Python ⁓

Python version. And like this, you can use multiple Python versions, different Python versions with different instances for each of your Python projects. Now the final step would be to actually fill in the.env files. I won't open them because it contains my credentials. But basically, what you have to do is go and copy these dot env.example files, right? Copy them.

and fill in all those credentials, right? So that's that's super important to make ⁓ the agents have access to things such as OpenAI, Gemini, Perplexity, Firecall, and all our dependencies. But the thing is that we have a lot of documentation on this. So that's why I want ⁓ insist more on how to install UV, how to install specific packages, ⁓ and how to

Sneha Mehra (00:04:13)  
fill in all those credentials we have in the README and in the less admin lessons, all the documentation you need to do this. So let's start by actually running the Nova agent. So as as stated previously, inside the research agent folder we have this ⁓ we have multiple other folders. So we will explain ⁓ what each folder contains in more depth and how they structure soon enough. But for now I just want to focus on

How to run, to run it, like right, to build the intuition. So to do so, we will begin by explaining the input to this agent. ⁓ And here in the data folder, we prepared multiple samples that we can play around with and run the agent. But for now, we will focus just on this workflows versus agent sample. So the actual input is this article guideline markdown file, which actually contains ⁓ what we expect from the final article, right? Because

Yes, this is a research agent, but this is a research agent specialized for writing content. Which means that we want to input an article guideline that explains what we want to write. Right? So, for example, in this article guideline we ⁓ at a very high level explain what we are planning to share, why we think it's valuable, ⁓ other details, ⁓ like about

The perspective of writing and things like that that are not that interesting for now. But what's important for the research agent is this outline of the article. Because you don't want the agent to actually decide what to write, right? That's that is still like the job of the writer, of the creator, ⁓ or ⁓ what you want to communicate to the reader, right? ⁓ And to do that, we we we created this outline where you basically specify everything that you expect.

From this article, right? So basically, we we can leverage this for the research agent to help it guide him ⁓ on what ⁓ to actually research, right? Because here in this guideline we actually explain everything that we expect to write. So it has enough information for the research agent to know what questions to ask and what to research, extremely aimed for this specific article because there

Sneha Mehra (00:06:38)  
As you can see here, ⁓ is a lot of information. And this is actually inspired from what we do in production. ⁓ So it actually works. ⁓ And on top of this, ⁓ more flexible free room area, we also added at the bottom these more concrete resources, right? Because we read a lot, we research a lot on our end as well. ⁓

Here we pass some links that we already know are super relevant to this particular article. ⁓ And the research agent in this use case will just crack them and put them in other markdown files in our research file, which will actually be the expected output. ⁓ And we can use that during writing. ⁓ And for example, we can put here things such as normal ⁓ article links, we can put ⁓ GitHub URLs.

Right, it it can scrape all GitHub repositories, ⁓ it can also look for video transcripts from YouTube ⁓ and notebooks ⁓ and everything that we considered is important to support as a research during ⁓ the article writing process. Okay, so now let's ⁓ run the Nova agent from using the CLI. ⁓

And then we will run it from cursor, right? To run it from the CLI, we actually have to move into the specific research agent ⁓ folder. ⁓ And now we are at the root. ⁓ And now we will move into that particular folder.

⁓ it's it's important to understand that we have to go into the MCP client folder, not the server, because the client will communicate to the server and we're actually running the client that will call the server.

Sneha Mehra (00:08:35)  
And to run the client, we are using UV, like to manage our Python project. So we have to do UV run. ⁓ And we we have to call the client.py ⁓ Python script, which is the entry point to the project. So to do so, we will call it as a Python module ⁓ and call that file relative to the source. And because we use it as a Python module and not as a Python file, we don't have to

hit the.py ⁓ extension. Now we run it. As you can see, now it started to run the project. ⁓ And this client is like a simple React agent that kind of mimics the cloud code experience, but super simplified. We thought is an interesting example to to see how these things work. ⁓ And which means ⁓ we can call specific commands with this slash

utility right so if you run slash tools we can see all the tools that are supported if we run

Slash prompts, ⁓ we can see all the prompts that are supported. Because again, this is kind of like a normal React agent that is hooked to our Nova MCP server, which means it has access to all the prompts and tools available from the MCP server. But we will dig in into this a bit later. So now I just want to show how how to run it properly from the CLI. So for example, if we want to call this particular prompt.

We do slash prompt and just copy that the name of the prompt and hit enter. And now it will actually call the prompt, the full system prompt from behind it. And now it ⁓ it's processing it. So as you can see, it basically went through the prompt, it it created a research plan.

Sneha Mehra (00:10:38)  
Again, we don't we won't focus on the exact steps of what's going on right now. I just want to build the intuition of what's going on. ⁓ So I will quickly walk you over like how a research plan looks like. So we do some setup steps where we, for example, we extract all the URLs from the article guideline file, right? So GitHub, YouTube, and other web links. ⁓ Or we do some pre-processing where we extract this and clean them. Then we actually start.

Research loop where we start doing queries based on the article guideline and then do some extra filtering and cleaning on top of everything and write the final research file. But we will dig more into all the details and everything that's interesting a bit later. For now, ⁓ it just says that we actually need to point the research agent to the research directory. So ⁓

We have in the data folder multiple samples ⁓ of examples that you can play with and run the research agent locally, right? So we have the sample workload versus agents that we will use in this particular example ⁓ and the article guideline from it, right? So we copy the path to the directory, ⁓ not to this particular folder. And actually, just let me

delete this dot nova folder first. So we copied absolute path to this this directory and state here is the path.

And as you can see, it's an absolute pad from my local computer. We hit enter.

Sneha Mehra (00:12:25)  
And now it it will actually start doing the whole processing. So we it calls first the process local files with this research directory ⁓ folder ⁓ and it starts doing its own thing. But I I won't let ⁓ the whole agent research agent run because it will take a while, and I want to show how to do it from the cursor as well, and there I will let it run entirely and see what's going on. So we'll just ⁓ interrupt.

This one kill the terminal and let's move on to actually running it from cursor, right? So we open the AI tab from cursor ⁓ and the MPC server is already hooked to cursor to this mcp.json folder. So as you can see, here we have the Nova agent. ⁓ And we basically just pinpointed it the

the MCP configuration to the to the exact folder, right? So we pointed to the lessons research agent MCP server, this time to the MCP server folder. ⁓ And here

Sneha Mehra (00:13:38)  
Basically, we have the server.py file ⁓ which contains which spin ups ⁓ the NPC server, right? And we also have to point it to an environment ⁓ file which contains all the credentials of things such as Gemini, perplexity, and things like that. ⁓ And that that's basically it. ⁓ it uses the standard input-output ⁓ transport, which basically states that when I want to run this into cursor.

It will spin up a process containing this MCP server. So it doesn't run remotely. It still runs on your computer, but as a different process from this cursor one. So before running the Nova agent to double check that this works, you can go to the settings of the cursor, then to Tools and MCP, and you should see the Nova agent here ⁓ and with the green light and with some tools, prompts, and resources enabled.

And then you know basically that it's working. Okay, so now actually let's run the Nova agent from cursor. So to do that, we will hit slash, similar to what we did in the terminal. Hit slash, and we will search the

Nova agent and hit this full research instructions prompt, right? Basically, exactly what we did in the CLI. Hit enter. And now again it processes the prompt retrieved from the MCP server. We again see the workflow overview as before. And now it asks us for the research directory as before. ⁓ So let's copy the absolute path. ⁓ And state here is the path.

And now it should start actually calling tools from Nova based on on this this prompt. Okay, so it asks us to confirm that everything is okay. We just say yes.

Sneha Mehra (00:15:43)  
And now as you can see it started with the X ray guidelines URL tool. So every time if you don't specifically say that you allow all these commands to run with not you as a human checking them, you will see this prompt with the all the arguments ⁓ that are input to that particular tool, right? So you can actually check if it's okay or not. And we'll just hit run.

⁓ And now it will continue calling multiple tools as you might be quite used to using any other MCP server or tools from cursor or cloud in general. Okay, so now the whole process will take a couple of minutes. So meanwhile, I just want to highlight that you can use the exact same strategy to call the MCP server from other.

popular tools such as Cloud Code, right? Gemini CLI or ⁓ whatever other MCP client you have, right? Because again we have all this logic into the Nova MCP server, so it it can be plugged in into any other tool that has MCP support. In this particular example we just use it cursor, but it's really not ⁓ limited to that.

So now I will just leave this running ⁓ and return when we actually have the output from from Nova ready to go and show you how the research looks like.

Okay, so the Nova research agent finally completed. It took a while, so I paused it, but let's let let me quickly walk you over what happened. So as you can see, it called quite a few tools, right? ⁓ And yeah, basically here it started. So

Sneha Mehra (00:17:39)  
Here it basically scraped everything from our article guideline. Here we have like a quick report of the URLs, GitHub URLs, YouTube videos. ⁓ And after it wrapped up ⁓ the actual scraping, it started the actual research, right? Where it started asking all kinds of queries based on the article guideline and the current context. ⁓ And it searched those queries on perplexity.

And it repeated that process three times ⁓ and ultimately it outputted the whole research into research file. ⁓ And here we have a quick summary of everything that happened under the hood. And what's actually the most interesting is in this.nova directory that I kept deleting, we have like ⁓ all the files that it it created.

during this research folder, right? It it created all these files like ⁓ local disk memory. You can say it like that. For now, as it works only on your local computer, that works perfectly fine. But we will see in future lessons how to do that ⁓ using a database such as SQLite, if you want to ship it to a server. ⁓ And here we have this research.md file where it actually aggregates

Did everything that it researched started from what it scrapped based on the given URLs to the actually ⁓ files that it found on Perplexity ⁓ using those research queries ⁓ and it aggregated everything in here. As you can see, it's a huge, huge file. In this particular use case, it has around 5000 lines. It's particularly fine because we are using Gemini to generate all our content, which has an input context of ⁓ 1 million tokens.

And as you can see, we will show you that using just inputting, stuffing all of this into the agent, ⁓ it works pretty well up to up to some some extent. And in our use case, it is more than enough to keep it simple ⁓ and actually make it work, right? That that's the end goal of all of this. So now we have this research, which is the output final from the Nova agent.

Sneha Mehra (00:19:57)  
Now our goal is to take this and actually write the article itself. To do so, we will move to the brown writing agent, which is a completely independent project. So to do that, we have to move to this writing workflow ⁓ folder. So we open up the terminal and we will move to the ⁓ writing workflow folder. ⁓ Right?

And this again is a folder managed by UV. ⁓ So to actually install all the dependencies that create the virtual environment, we'll do UV sync, which installed everything we need. We already have everything installed, so it's fine. We also need these.m files with all the credentials filled in. ⁓ basically, we need these credentials filled in. I won't go into the details, we have

All the instructions that we need to install this ⁓ in our documentation. But for now, I just want to show you how to run everything, right? That's the most interesting part to see in video. ⁓ So to run this, we actually will start running it from the CLI as before, and then move to running it from cursor. ⁓ So here in the browne writing agent, we are running everything through a makefile, which helps us aggregate all the commands that interface our.

Brown writing agent into a single place. So if we open the make file, as you can see, we have ⁓ all kinds of commands that again are using UV to interface this project. So for example, if we run brown ⁓ generate article with make, so we need to have make installed, but on ⁓ Unix systems, I think it comes pre-installed. So we do make, we do brown generate article, which will actually run this command from here.

And we need to actually input as an argument the DIR path, which again is the same DIR path ⁓ where we actually have our research and article guideline. This time we need both. Both article guidelines, which is gu which will guide the ⁓ article writing process, which actually doesn't need to be necessarily an article, it can be anything that is in written format. ⁓ And the research, which will be used to support ⁓ all the factual data.

Sneha Mehra (00:22:17)  
From the article from the article generation, right? So the article generation is anchored only in the research. The agent cannot use ⁓ any other factual data than what's in the research. And like that, we can reduce hallucination close to zero. ⁓ So we force that it will never give answers that are not in the article guideline or research. That's super important. Okay, so we take the absolute path from this folder.

And we put it here. We click start. And this is not an agent, this is a workflow. So basically everything will start working out of the box without any questions, without any other queries. So we have this progress bar, which basically will inform us where we are. But again, I will stop this ⁓ now ⁓ and show again how to run it from the cursor interface.

I think that in the cursor interface everything is more interactable and prettier and more interesting to see, right? So again, we do / and similar to what we did for Nova. ⁓ And this time we pick this brown generate article prompt, where again we need to input the whole ⁓ path to the directory containing article guideline and research. Hit enter ⁓ and

Again, this prompt will instruct the brown MCP server ⁓ on what tools to call from the MC server, right? So as you can see, it calls the generate article tool with this this deer pad as input. We click run ⁓ and it will start running. So as I said, this is a workflow, not an agent. So it will call just one tool and it will do its own thing.

We have again this progress bar that is informing us where the article generation is with some description. And while this is running, I realize that I haven't shown you how to plug this in to cursor. So we did it quite similar to Nova. Again, we go to the mcp.json file and we along the Nova ⁓ config, we have the brown config, which is following the exact same pattern. We use UV, we point it to our directory.

Sneha Mehra (00:24:38)  
And this time we pointed to brown mcp server file, right? So let me show you. So we have this writing workflow brown mcp server file, which contains our mcp ⁓ instance. More on this later. ⁓ Now we will just focus on the functionality. ⁓ And again, I just want to point out that it's super important to give this

MCP server instance the right credentials through the MV file. Okay, so again this will take a few minutes to run. So as before, I will pause it ⁓ and we will return when we have the final output and I will show everything that happened during that process. Okay, so the final generation of the article finally completed. ⁓

instruction point point of view we don't have much going on it just instructed us where everything was generated so we go to that particular folder over here and we can see we have multiple files. So we have this article 000, 001 and ⁓ 002, which basically are multiple stages of the article. We will explain this ⁓ further but the core idea is that the article goes automatically to multiple reviewing processing stages

Using the evaluator optimizer pattern, right? ⁓ So like this, we can see how the article looks like after each review process, and this is the actual final article. As you can see, it it follows a clue quite standard pattern where we have in Markdown the title, the subtitle, the introduction, and multiple sections over here of the article.

Right. So it's quite comprehensive. And you have all these references used to write the article. Because the article doesn't necessarily need to use ⁓ all the references from the research. It has the freedom to choose whatever is useful based on the article guideline and reference only that. So it has the freedom and it should have the freedom to pick only what's useful ⁓ to write the article.

Sneha Mehra (00:27:04)  
Based on the user intent from the article guideline. ⁓ And yeah, that's that's basically the core idea. This is the final output. ⁓ And as you can see, and this is really interesting to show you in future lessons, ⁓ is how we managed to make the article writing more or less follow a readable format, a more human format without all the dashes and a lot of like semicolons and a lot of bullet points. And basically to show him.

How to format all of this as we want, right? That's ⁓ the secret sauce of this problem. Yeah. And I think the last thing that will be interesting for you to look like is ⁓ to see the similarities and the similarities in structure between the article and the article guideline. For example, the first section is called understanding the spectrum from workflows to agents. ⁓ And if we go here on the first ⁓

Sneha Mehra (00:28:07)  
Section that we expected, it has the exact same title because this is the introduction, right? So this doesn't have a title in the article, and this is the first section that we have within the article. Because this is the the introduction. And for example, the next one will be choosing your path, choosing your path. And if you are really curious at this point, you can see that basically what's inside here will be inside ⁓

the article as well. ⁓ And for example, sometimes ⁓ inside this guideline, you can even see that the guideline has is more verbose than article per se, and that's fine. Right? So the idea that in the article guideline you can dump whatever thoughts you have, unstructured on unstructured. ⁓ And the secret juice of the article agent is to actually distill that information and

And combine it with what's useful from the research to actually have this clean output at the end, right? So yeah, that that's basically it. The final output of this whole multi-agent system is the article, which we can later take ⁓ and even further edit it, right? I I think that's ⁓ one less thing that's interesting to to see with the writing agent workflow ⁓ is how we can further

interact with this because ⁓ most of the time you're actually not satisfied with this final article. You will see as you read it that you might forget some stuff or the agent is not perfect and it might forget some stuff or it messed up with some writing. And that's why we created these commands to actually allow to hook the human in the loop. And while you read it to actually be able to easily further call the agent to edit it. And actually during the edit,

It just runs against the evaluator optimizer pattern, right? Which actually runs another review and editing process. So for example, to do that, we can either edit the whole article or edit just a piece of text that we know we want to further refine. So for example, let's run the editing process just on a piece of text. So for example, similar to how you write code with cursor, you can select this, this ⁓

Sneha Mehra (00:30:37)  
Lines and now we do something like brown edit selected text prompt and it asks us at the top for some human read feedback. ⁓ let's say that I know something completely random. ⁓ I'm not an AI engineer, I am a data engineer. Also make the introduction ⁓ shorter into ⁓ one.

paragraph. Also make the selected text shorted one paragraph because we selected only ⁓ the first two paragraphs, right? So you still have ⁓ think about what you want to instruct the agent. It's not completely free of constraints, right? That that's one trick in actually doing successful AI applications. Now we hit enter ⁓ and again this is just the prompt which is instructed to call the right tool.

To handle our request. So we call the edit selective text tool from the Brown MCE server with this article path, which now it could infer it automatically with the human feedback, ⁓ which is basically just what we said earlier. The selected text, which is the selected text from the article, and the lines from five to seven, which is correct, from five to seven. Right. Now we click run ⁓ and now this will be a lot faster than writing the whole article because it will

Actually, write just this piece of text and it will also show us the div. So this tool will actually output ⁓ the edited text. And cursor with some prompt engineering will take care of actually applying that div to our article so we can select only what we care about. ⁓ And what's happening behind the scenes is doing a review process with our out of the box logic that was also used while generating the article, plus

This human feedback applied to the reviewing process, similar to how a writer ⁓ and a reviewer would work together, right? So as you can see, it already finished because this is a lot a workflow a lot lighter. ⁓ And it applied the diff. So as you can see, it followed our request from two paragraphs, it reduced it to a single one. ⁓ And yeah.

Sneha Mehra (00:33:04)  
And we hit Keep and that's it. So this is basically what we will build, this multi-agent system from Nova, where it scrates and it does the research for us that outputs the research.mt markdown file to the Brahm writing agent that takes us input the article guideline and research and outputs the final article. Okay, so that's in a nutshell the capstone project that you will build.

As you can see, it's quite complex. It's a multi-agent system that has two ⁓ big agents: the Nova agent that takes us input the article guideline and outputs the research ⁓ MD file. And the Brown writing workflow, which takes us input both the article guideline and the research and outputs the final article. So I hope you will have a great time digging into these projects because they're actually really production ready.

We actually use them to write articles and we carefully refine so everything actually works. And that's why I think this will be super interesting ⁓ to see. So in future lessons, we will start by digging into the project structure and design of the Nova agent and then digging into the actual functionality of it.

