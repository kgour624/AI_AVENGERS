55

Harsh Bharadwaaj (00:00:00)  
What do you call a f your friend who st I can't I I read it already. It's I think this is pretty funny. What do you call your friend who stands in a hole? Phil. ⁓ anytime I hear the name Phil, I I apologize if your name is Phil. I hope this doesn't offend you, but when I hear the name Phil, I think of Phil Connors from Groundhog Day. ⁓ Phil Connors, I thought that was you. ⁓ that ⁓ Ned Ryerson? Okay, I hope that some of you get that. ⁓

all right, so this is good. I hope you had a really good time with this exercise. Now is the time to get a drink of water, go get a snack, go high-five somebody, compliment somebody on their outfit, whatever you want to do to make the world a better place ⁓ and take care of your physical body. ⁓ And when you're finished with that, then come right on back because I'll be ready to take you on to the next one.

—------------

54

Harsh Bharadwaaj (00:00:00)  
This exercise is a little bit different than the others because it's a little bit of prompt engineering, is what we're doing in here. So I'm going to stick in the prompt that I think is going to generate what I want. ⁓ And to be like really specific, we have some examples right here ⁓ of the types of output that we really are expecting to get from the LLM. ⁓ Additionally, like we might be able to control the system prompt in some tools that we're using. I'm using code name Goose.

And it works really well. It's really configurable and stuff. I might be able to control the system prompt, but this is sufficient. Just send a message. ⁓ It's going to say, okay, yeah, I understand. What whatever. Yeah, send me the information. ⁓ And so we're also going to send ⁓ resources. So here is all of the information that you need. And then hey, I verified it is sending me back ⁓ some ⁓ like proper responses. Here's ID2, 3 and 4\. ⁓ And here's Beach Day and Family Fun. Okay, I can parse that.

I can do something with that. So my prompt is good. ⁓ As a result now, I can come over here and we're gonna update our system prompt. And I'll stick in you are a helpful assistant, suggest relevant tags ⁓ to make them easier to categorize and find later. So kind of explaining what the purpose is. I I kind of see the system prompt as this is your reason for existing. Like there's that that meme of the little robot, what is your purpose? What is my purpose in life? You pass the butter, or whatever, right?

So it's like this that's what the system prompt is. What is my purpose in life? This is your purpose in life. ⁓ And so here's the purpose. Here's what you can expect. I'm gonna give you a journal entry and the current tags, an existing one, ⁓ suggest tags that aren't already applied. The journal entries, they should have four to five ⁓ tags. It's perfectly fine if they don't have any. Feel free to suggest new tags that are not currently in the database and they will be created. Okay, great. With that information, ⁓ now

I want you to respond with JSON only. Here are some examples. The examples are really, really helpful. This is key. ⁓ So if you have no suggestions, again reiterating, you don't have to suggest anything if you're happy with the tags that are already applied or whatever. ⁓ So here's an empty array. And then if you do have suggestions, here is how it should appear. This is referencing an existing one. This is asking for a one to be created. This is referencing yet another existing one.

Harsh Bharadwaaj (00:02:22)  
And so now we just need to update this to be text.json. ⁓ And we need to stringify the ⁓ whoops. We need to stringify the ⁓ entry, the current tags, and the existing tags. So then let's get rid of that. And so the system prompt is like: here's your purpose for existing. The message is here's the specific implementation of that purpose. That's the way I think about it. It seems to work out pretty well.

max tokens also is something you need to play around with. You need to make sure that you give it enough max tokens for the amount of response that you're expecting. 100 tokens is probably fine for what we're expecting, but if you're expecting it to like generate a book or something, then you're gonna need more tokens for that. So you'll wanna kind of play around with the max tokens. So if we go with that, then I also gave you something you could copy paste.

as a response so that you don't have to hand code it or hand write it yourself. ⁓ So let's copy that and we'll come over here to our tools, create an entry. Here's the title, here's the content, run the tool, we get the sampling request. We set that up last time. I'm gonna paste in my response, we'll approve, ⁓ and boom, we got a notification. Here it is. ⁓ Added tags to entry. Here are the added tags: ⁓ personal, health, and beach. ⁓ And here's the entry. That's entry ID of six.

There are three tags, beach, health, and personal. ⁓ we could go look that up and it should have all those tags. So ⁓ success, right? Like we we have a successful sampling request. If you want to dive in deeper, feel free to look at the code and how we did this. Every implementation is gonna be a little different though, that's why I didn't have you do all of that. ⁓ the basic idea is you're going to ⁓ wire all of this up to make the sampling request, you're gonna ⁓ fine-tune your system prompt and the messages that you provide.

And then you're going to parse out the prompt response so that you can then make use of it in however you need to use. ⁓ And that ⁓ is your sampling.

—----------

57

Harsh Bharadwaaj (00:00:00)  
So it's the end of the year and users want to see their wrapped video. It's the video at the end of the year that summarizes their experience writing and all of these journal entries and everything. So we have this tool, Create Wrapped Video. And so I'm going to create a wrapped video for the year of 2025\. We're gonna run the tool.

Harsh Bharadwaaj (00:00:21)  
And we're gonna sit here waiting. ⁓ okay, there we go. It finally, you know, sometimes things take a while, alright? That's that's fine. Okay, now we've got a video and we can go take a look at that video. In fact, it is saved in videos and we can open it up ⁓ and it will in fact play. ⁓ my goodness, it's my wrapped video\! This is so great, and it has my username, it's all dynamic data. This is s ⁓ it's September 8th, 2025\. I literally just dated my video. ⁓

That is when I recorded this. ⁓ Here is your epic me wrapped for video for 2025\. Isn't this so fun? ⁓ I have a wrapped video. I wrote five entries in 2025\. ⁓ Okay, so my point is we we made I made this video the way that it is, ⁓ being really slow and kind of boring intentionally because I wanted to demonstrate long running tasks. Sometimes we have tasks that take a while to ⁓ to generate, like generating a video. ⁓ And as you're working on this and and testing things out.

If you don't have FFmpeg installed, then this might not work. ⁓ And so you can actually set a mock number of milliseconds. That'll take three seconds. This will take 30 seconds. You get the idea. ⁓ And so if you want to, you can use that. ⁓ But my point is, that experience is not great. ⁓ I want to be able to give the user some sort of feedback ⁓ of progress. And this is built into the MCP spec. So if we look at the solution when you're finished with this, if I do that same thing again, you'll notice we're getting

⁓ these messages coming across, these notifications ⁓ of progress coming across for how what our current progress is. And so then the client can show the user some sort of progress bar ⁓ or ⁓ at least some sort of indication that yes things are happening, like progress is being made. Don't you worry about it. Eventually your amazing wrapped video will be finished. ⁓ So your job in this exercise ⁓ is to implement that those progress notifications.

Send that those notifications that include the progress and the total. We're gonna have a total of one, and then the progress is gonna be ⁓ somewhere between zero and one, ⁓ as well as a video, a message saying create video. ⁓ so ⁓ that is the the basic idea here. You're gonna be ⁓ learning a little bit in this exercise about ⁓ signals and aborts and abort controllers and stuff like that. ⁓ Some of this is going to kind of be done for you.

Harsh Bharadwaaj (00:02:45)  
Some of this you're gonna have to work around ⁓ the ⁓ mock time versus ⁓ the actual creating of FFM peg. ⁓ But all of the emoji are there to kind of guide you through this process. So hopefully it's not too much of a pain. ⁓ All right, I think that's enough for you to get going. Have a good time.

—----------  
51

Harsh Bharadwaaj (00:00:00)  
Okay, it's time to borrow the user's LLM. This is actually really, really cool that we can do this. So, ⁓ what this exercise is going to be about is the create entry tool. So when you create a new entry, we're going to suggest new tags for the user. You know what's cooler than the prompt that we have for suggesting tags? We what's cooler is actually just proactively doing that. And so that's what the ⁓ our sampling request is going to be when you create a new entry.

Let's go ahead and just ask them if we can go ahead and use their LLM to suggest tags for that entry they just made. ⁓ And if they allow us, then we can be a lot more helpful MCP server without actually having our own LLM. We'll just borrow the one the user has. ⁓ So ⁓ I'm going to type in here, type in here, and run the tool, and we just create the entry. Nothing really special happens ⁓ in the ⁓ your case when we're just getting started. So your job in this exercise step.

is to make it so that it creates a sampling request. ⁓ As a part of that, we're going to actually ⁓ use logging as well so that we can log ⁓ some server server notifications to let ⁓ us know as we're looking at this ⁓ what is going on. ⁓ And that it can be quite helpful. So we're going to be turning on logging. ⁓ You can switch the log level if you so choose. That's all handled by the SDK for us automatically. So we can switch this to emergency that sets the log level

And now the SDK is only going to log really, really emergent things. But we're gonna leave it on debug ⁓ and we'll go to create entry. We'll type in a nice title and some content here. ⁓ Do you hear that truck? ⁓ Unbelievable. I'm gonna leave that in because it's gonna make some of you chuckle. ⁓ Alright, so then we're gonna run the tool. And there we go, success.

⁓ it might look like nothing has happened, but when you're when you've got it done right, you'll get this sampling request right here. So we're gonna go sampling, ⁓ and you'll see we've got our messages. Here's the message that we want to send to the LLM, here's the system prompt, here's the max tokens, all of this stuff. ⁓ All you're going to do in here when you're finished is just say hooray or whatever you want to say in there, and then hit approve. And then you should get another notification from the server that ⁓ is logged from the tag generator. That's what we're gonna call it.

Harsh Bharadwaaj (00:02:21)  
and the data will say message response ⁓ from the model, hooray. ⁓ And that is what we're going to be logging. Once you see that, you know that your sampling request was successfully processed. Now we're not actually going to be ⁓ generating tag suggestions or anything like that just yet. ⁓ we're just wiring up the sampling. So your job in this exercise step is just get things wired up, and then we'll see you in the next one.

—---------

70

Harsh Bharadwaaj (00:00:00)  
Hey, good job. You did it. What a great workshop. I hope you had a really good time with this one. And now you know some of the more advanced features of the Model Context Protocol, which I think is really valuable for ⁓ the really ⁓ interactive and dynamic ⁓ MCP servers that you are hoping to create for your users. We want to give our users the very best user experience as they're interacting with our server, ⁓ or rather, as their agents are interacting with our server.

and those client applications to be able to keep track of the resources as they're changing, to ⁓ know what the tools are all about, to be able to elicit some additional ⁓ information from the users and to borrow the LLM, all of this really cool stuff. I hope you had a really good time with this one. ⁓ Good job on this workshop. You should feel proud of yourself. Now go out into the world and make it better.

—----------------

40

Harsh Bharadwaaj (00:00:00)  
Alright, let's talk about some advanced things that you can do with tools. We'll start with the ⁓ ability for you to specify annotations. So, annotations include the title, which we actually have already talked about. The SDK has a special place ⁓ for you to put the title, ⁓ and this makes a lot of sense. ⁓ But then we also have a couple other flags that you can specify on tools so that LLMs and the clients know what to expect before even calling the tool. ⁓ So the flags that you can

provide are read-only, disc ⁓ the read-only hint, destructive hint, item potent hint, and open world hint. So ⁓ it's kind of interesting here because if the read-only hint is set to fal or to true, then the client knows that it's can't be destructive and it can't be itempotent. And so ⁓ that can be fine. ⁓ that and so you don't need to specify those. ⁓ but if it's set to false then it can technically be destructive and et cetera.

⁓ And there are a couple of defaults in here. ⁓ I ⁓ have some help that I'm gonna give you because it can be a little confusing. Like, do I need to specify this property or ⁓ is read-only hint setting that to true making setting these other properties unnecessary? Whatever. So in the exercise, I give you a little bit of help ⁓ through typings to make sure that you're setting everything properly.

⁓ but a lot of this has some nuance to it as well. And if you're not very familiar with things like what's itempotent and what counts as destructive, then this might be a little bit of a challenge to you. But here's the basic ⁓ description of each of these annotations. So read-only hint defaults to false. And actually I should say that all of these hints default to ⁓ the most ⁓ permissive or or I suppose conservative ⁓ thing. So we're assuming the worst, I suppose.

We're assuming it's not read-only. We're assuming it's gonna destroy stuff. We're assuming it's not idempotent and assuming that it's open world. So that way we can ⁓ with those assumptions be the most protective of ourselves as possible. ⁓ So that's what the determines what the defaults are. So we're assuming it's not read-only. If it's true, then we're going to say it's not gonna modify anything. If it's destructive, ⁓ then we're going to assume that this performs some sort of destructive updates or or dis ⁓ it's going to destroy some existing.

Harsh Bharadwaaj (00:02:19)  
Data. ⁓ That one's kind of interesting because if you update a record, does that count as being destructive? I'll let you think on that a little bit. ⁓ And then item potent means ⁓ if it's true, calling the tool repeatedly with the same arguments will have no additional effect. ⁓ And ⁓ open world, if that's true, then may ⁓ interact with the open world of externalities outside of the tool's own domain. ⁓ So we want to be a little pro pragmatic about this. ⁓

We want to use destructive true if the tool is gonna delete anything, and destructive false if it's just gonna modify existing records. ⁓ even if original content's getting overridden. ⁓ the the spec is not ⁓ totally ⁓ the the language is not totally clear on whether this is the correct imprint interpretation. I could be you could argue either way, but this is the most pragmatic approach as far as I'm concerned. Otherwise, every update is gonna be destructive and it just doesn't communicate effectively between

Will this delete a record or will it just update one? ⁓ I I think that distinction is important. And so that that's why we make that distinction here. ⁓ item potent, if it's gonna pr ⁓ create the same logical output. So even metadata changes like timestamps and stuff, we're not going to take those into consideration for whether something is itempotent. So if it's gonna produce the same basic result, yes, the record was updated, then we're good with that. ⁓ And for open world, we're ⁓ focused on

Systems external to our own. So we've got a database, and that's not technically part of the open world. We're we're gonna think about okay, we're gonna make a search on the internet or something like that. ⁓ So that is some of the annotation stuff that you can look forward to in this exercise on the advanced tool stuff. ⁓ Now let's talk about the structured content stuff. ⁓ So this is ⁓ a tool definition. It includes our input schema. ⁓ it also can include an output schema.

So with an output schema, you're able to let the LLM and the clients know that, hey, ⁓ when you call me, this is what you can expect to get back. And so if you were to call this generate fantasy character, you can expect to receive the ⁓ the name and the species and the character class and its abilities. So if you need any of those things, then go ahead and call me because I'm gonna give those to you. That's the benefit of having ⁓ structured content and this output schema.

Harsh Bharadwaaj (00:04:42)  
So here's what an example of that might look like. ⁓ Once that has been established, you're going to call the tool with the arguments and you're going to get the response. ⁓ And in the content, you're going to have the same thing that you have in this structured content. So you'll have ⁓ this sibling property to the content that's called structured content. And that's a JSON representation ⁓ of ⁓ something that matches the schema that you said you were going to return. ⁓ But then you also are going to provide content.

That includes the serialized version of that ⁓ structured content. And that is for backward compatibility reasons. ⁓ There are some clients that don't support that. And so the specification recommends that you include that. And in fact, the MCP inspector also validates that you do include that. And so ⁓ for the time being, we're going to include that. Eventually, maybe all the clients will take structured content into account. ⁓ And then we can remove that from our content, which would be nice. ⁓

So taking a look at the inspector, ⁓ when you're all finished, you should have the annotations listed, ⁓ spoiler alert, ⁓ and ⁓ output schema listed on each one of these. And then you'll be able to ⁓ call the tool. And when you do, actually, you'll see the output schema right here. You'll call the tool and you'll get ⁓ valid according to output schema and structured content matches text blocks and other content. ⁓ and so that is what we're gonna work on in this exercise, some advanced.

things that you can do to make your tools more helpful for the clients. ⁓ I hope you enjoy this one. Let's get right to it.

—-----------

43

Harsh Bharadwaaj (00:00:00)  
There's a feature of tools in MCP called structured output. What is structured output? Well, it's basically a ⁓ specified way to output your tool results. ⁓ And what ⁓ makes this special is that it's predictable. So it comes as a part of your tool definition. ⁓ And so the client or even the LLM can know what to expect back when the tool is called. Let me show you what I mean.

So when I call this get entry with ID1, I have no way of knowing what's going to come back before I actually hit the run tool. ⁓ Once I get it, then I know, okay, so this is a JSON object and here are the properties and things. ⁓ But if we add structured output and an output schema, then I actually have the output schema before I even call the tool. So here if we ⁓ go to get entry ⁓ and I say I want to get entry number one, before I even call the schema or call

run tool, I see this output schema. Here it says I'm gonna get an object of type or yeah, a thing that is an object ⁓ with the following properties. There's gonna be an entry, it's a type object. It has these properties. Each of these properties are of these types ⁓ and ⁓ potentially even have these descriptions because this is all just JSON schema stuff. ⁓ And here are the required properties that you can expect. And so then we can start making plans or we can ⁓ potentially

⁓ know, okay, so this is the tool that I need to call if I need to get the location ⁓ of the this ⁓ entry or whatever. So it allows me to know what I'm going to get before I try to go and get it. ⁓ and that is what we get by having an output schema. And then when I run the tool, once you've finished this exercise, this is the result, you'll get the structured content that matches that output schema. So it's got an entry, it's got all these properties and these have these different types.

And the MCP inspector will actually validate that the output matches the schema. And it also validates that the structured content matches some unstructured content as well, as a backward compatibility thing. ⁓ And so ⁓ in this exercise, you are going to be adding structured content to your tool results, and you're also going to be providing that exact same structured content ⁓ as one of the text content ⁓ output ⁓ as a result of your tool call.

Harsh Bharadwaaj (00:02:25)  
For that backward compatibility. ⁓ Because of that, we're actually going to be going from embedded resources to resource links so that we still get the benefit of the resource link, but we also get the benefit of structured output. ⁓ You'll also notice here when we do ⁓ the tools list, this is where that ⁓ output schema shows up. So this is going to be one of the configuration options you add ⁓ that ⁓ shows, hey, here are the properties you can expect when you call this tool.

And you'll notice ⁓ especially important is the fact that ⁓ that shows up in the tool list call and not in the tool call. So we know before we even call the tool what we're going to ⁓ be getting. Not the specific data, but the shape of that data. And that's what makes it valuable. So the LLM knows, if I call this tool, I'm gonna get this ⁓ this result. And that can be even more helpful when the LLM is making decisions ⁓ or if there are clients who are ⁓ building specific integrations with a specific server or something that can be

useful as well. ⁓ So that is your exercise. There's ⁓ you can apply this to all of the tools. It's a pretty ⁓ boring process once you figure out how to do it for the first tool and then you just kind of apply it to all of them. So feel free to stop once you kind of get the idea and then we can move on to the next one. So give it a whirl.

—--------------

68

Harsh Bharadwaaj (00:00:00)  
Hey, so I've got some good news. We can finally update our capabilities in a way that the SDK wouldn't do for us automatically. And that is the SDK does not yet know, maybe it will in the future, but it doesn't know that I have the subscription capability, or I'm about to. ⁓ And so I'm going to add subscribe explicitly. Ha ha. ⁓ And the way that we add subscriptions is we're going to have this initialized subscriptions. So we can add our listeners.

For the subscribe ⁓ callbacks and all of that stuff. So we're gonna import those. We'll import ⁓ some utilities for subscribing to the videos. ⁓ And then we're going to ⁓ listen for all of the ⁓ URI ⁓ subscriptions. So ⁓ as ⁓ subscribe requests come in, they're gonna come in with a URI that they wanna subscribe to. I need to keep track of those and then they'll unsubscribe, so I need to remove those as well. So we're gonna add both of those.

In this server, this is a standard I.O. server, so I can just stick URI subscriptions in memory like this, and it will be just fine. But if you were to deploy this to a server, you'd probably need to segment this. Maybe you'd persist it somewhere, something like that, ⁓ so that it survives redeploys. In Cloudflare ⁓ with durable objects, you'd probably put this in state, and each one of the durable objects is individual for each client, and so you don't have to worry about segmenting in that context.

Because it's segmented for you automatically in the durable object, which is pretty cool. ⁓ So anyway, we can do this at the module level and it's working just fine for us. ⁓ Let's start by calling agent server server. We're ac accessing the underlying server implementation ⁓ because this is a more advanced API you don't typically use ⁓ to set the request handler for the subscribe request schema. So when a subscribe request comes in, ⁓ this function will be called. And it will be called with params.

We're gonna take the URI and stick that in the URI subscriptions. Then we'll return an empty object that's just like, yep, got your message. You don't need to know anything about that. ⁓ And then we'll do a similar thing with the unsubscribe ⁓ request schema. We're gonna delete ⁓ the ⁓ URI from our subscriptions. They don't care anymore, they don't need to be notified of changes. ⁓ And then we need to ⁓ subscribe to the database. So we're gonna say agent db subscribe.

Harsh Bharadwaaj (00:02:18)  
We'll get the changes. ⁓ And for each of the changes that are relevant to an entry, we'll construct that URI. ⁓ If that URI subscriptions has that URI, then we will wait for the ⁓ server-to-server notification to happen. So we're going to send a notification, say, hey, the resources have been updated. ⁓ and here's the the resource ⁓ URI ⁓ that was ⁓ updated, it was updated to. ⁓ And then we can come over.

⁓ further here and do the same thing for tags. ⁓ And then we're gonna subscribe to video changes. ⁓ We're gonna gr when a video changes, we're gonna go get all the videos, and for each one of those, we'll check if ⁓ somebody's subscribing to it. ⁓ And if they are, then we'll send a resource update. ⁓ And with that, then we are happily able to connect to our server ⁓ right here.

We'll go to ⁓ list our resources. We'll go here, let's go to food this time. ⁓ this is ID of eight. We'll subscribe to that one. We'll go to our tools, we'll list our tools, we'll go update tag, tag with the ID of eight, ⁓ and ⁓ food. ⁓ And then we'll run and we get ⁓ resources list changed. That makes sense. ⁓ resources did change. ⁓ we'll also get resources updated, and ⁓ that has our tag and everything that we want.

There. And so then we can come over to our resources, we can refresh, and ta-da. And then if we unsubscribe, then we shouldn't get another notification. So we'll go over here to tools, we'll update tag 8 with ⁓ just lowercase food. We'll run the tool. We do get list change, that still is relevant in this context, but we don't get that ⁓ this was updated, that's our previous notification there.

So that takes care of subscriptions and unsubscriptions. Great job on this one. Now our clients can subscribe to resources ⁓ and keep things updated in their context. Awesome work.

—------

46

Harsh Bharadwaaj (00:00:00)  
Sometimes when you're in the middle of your tool call, you realize, ⁓ I think I need to have some additional information from the user. I need to ask them for some confirmation. I need to ask for some extra information. ⁓ That is what elicitation is for. It's kind of a funny name, but you're going to elicit information from the user. You're going to ask them information. So you are on the server. You need to make a request to the client.

Typically the user is the one who's going to be answering this, but I could see some situations where an LLM might answer for the user if the LLM is confident ⁓ that it knows what the answer is going to be. ⁓ and as f as far as I'm aware of all ⁓ clients, they actually just present a form for the user to fill out the information. So here's an example: we've got ⁓ pizza order, ⁓ and so we've made this ⁓ request to ⁓ create a pizza, and then while you're in the middle of the tool call, you notice, hey, they didn't add a drink to their order.

And we are capitalists and so we want more money. So we're going to elicit ⁓ information asking them if they want a drink. And maybe this isn't just capitalist. Like lots of times people forget to add a drink to their cart. ⁓ and so you can ask, hey, do you want a drink? ⁓ this is your request schema, this is a JSON schema, and you're saying, Hey, I require that you tell me what drink you want. ⁓ maybe maybe don't do that. Maybe like it's okay. Or here, we have no drink. ⁓ That is fine.

⁓ but yeah, all JSON schema stuff that you can generate a form out of. ⁓ And the user has the option to accept or decline or to cancel. ⁓ And each one of these is kind of different, so they provide the the requested data, they explicitly refuse to give the data or they just ⁓ dismiss it without a choice. I kind of see decline and cancel as kind of basically the same thing, but there is a distinction, I suppose.

And then the response is going to be like, hey, the action was to accept. And here is their response. So this is the content. And from there, you can use that to create their order with their drink. Hooray, that's good. ⁓ So ⁓ the flow diagram for this, just to see this visually, is we submit the pizza order. The order is forwarded to the LLM, and then the LLM ⁓ decides to call the pizza order tool. So that's submitted to the client that goes to the server. The server is in the middle of processing the order when it realizes, they don't have a drink yet.

Harsh Bharadwaaj (00:02:15)  
So let's send an elicitation request to the client, forward the elicitation request, present that to the user, the user makes a selection, that gets forwarded to the client, which forwards to the server, and then the server can continue with its tool call. ⁓ Now, elicitation isn't just about handling extra information during tool calls. You can make an elicitation request at any time. ⁓ I can't really think of any situations where you'd want to do that outside of a tool call, but ⁓ stranger things have happened.

So ⁓ that is what you're gonna be building in this exercise. Here's a little demo of what this is ⁓ what this is gonna look like. You are going to be able to stop people from accidentally deleting stuff. So if I run this tool, it's gonna take me over to elicitations. I can confirm that, and now we've verified yes, you do want to delete this thing, and so we can delete it. ⁓ So that's the idea. We're gonna be building that in this exercise. I hope you have a really good time with it. We'll see you in the exercise.

—-------------

52

Harsh Bharadwaaj (00:00:00)  
We're gonna start in the create entry tool. So this is where we want to perform the sampling request. We want to right before we ⁓ return, ⁓ but not make the return wait. ⁓ We want to perform the sampling request. And so ⁓ after we've created, we'll say suggest tag sampling ⁓ from the sampling file, we're gonna pass the agent and the created entry. ⁓ But we're getting a red underline, and the reason that we are is because this is an asynchronous operation.

So it's telling us, hey, you need to await this. ⁓ But if we await it, then that means the user's not going to get the response that the creation was successful until they've approved the sampling. Maybe they don't notice the sampling. Maybe they decide to ignore the sampling. We don't want them to have to wait for that. And so instead, we're going to put void, which is just a JavaScript way to say, hey, ⁓ I know that this returns something, but I don't care about it. ⁓ and so we're just ⁓ it it's kind of a it's actually not necessary. You could just do this, and this technically will work.

But by putting void, you're communicating to your coworkers that you didn't forget the wait, you actually intentionally don't care about this as a fire and forget situation here. ⁓ So with that, we're gonna jump into suggest ⁓ tag sampling, ⁓ and we're going to get the capabilities of the client. We want to make sure that they are capable ⁓ of performing a sampling request. If they're not, we'll just log that and return. ⁓ And sorry, they don't get suggested tags, and that's fine.

⁓ if they do, however, we want to send a message ⁓ to the from the server to the client. So this is another like elicitation. It's another ⁓ request from the server to the client. Now before we do that, let's fix this. Get client capabilities capabilities. ⁓ And ⁓ this could potentially be undefined, and we're this is not an await situation there, because we get that during initialization. ⁓ So there we go. Now we can create a message. Let's say agent. ⁓

Server server, ⁓ we've got to access the underlying server. It's it's ⁓ funny, that's just the way the SDK is. ⁓ And then create message. Now, how does create message ⁓ expect messages? Well, we're gonna take params. We've got our create message params, and then we also take options. So there's request options on progress, signal, timeout, bunch of cool stuff that's interesting, but we're focused ⁓ on the create message request params specifically. So we've got messages, has roles and content.

Harsh Bharadwaaj (00:02:23)  
It also has, if we scroll through all this stuff, ⁓ max tokens, meta, system prompt, include context. This one's interesting. ⁓ so I guess what it's saying is, hey, I want you to include the context from this server. ⁓ I'm not sure what that really means. ⁓ especially all servers, I'm not sure that's really advisable. Do you want somebody to request sampling and have access to all of your other servers? Probably not. ⁓ that just seems dangerous. So ⁓

I don't specify this ever. I think maybe it makes sense if you're building your own clients and there's a lot of trust involved with all the servers that are involved. But for a ⁓ general tool ⁓ or a general client, I think that they probably should just ignore this altogether. ⁓ So I don't ever specify that. ⁓ you can also ⁓ give hints around the temperature and stop sequences, model preferences, all of that stuff. Feel free to go hog wild on that. I don't really bother with that. I'm just like, ⁓ give me the

⁓ a sampling request, the do the best that you can. ⁓ But there could be situations where you have a really specific need. And so being able to specify the name of the model that you would prefer and your priority on cost versus speed versus intelligence, ⁓ that could be useful. And there could be clients that take advantage of that ⁓ as well. So let's get back into this. We're going to start with ⁓ our messages. ⁓ And our messages are going to ⁓ include a message from the user.

That says, hey, you just created this new entry ⁓ respond with accommodation. ⁓ And we'll also have a system, whoop, whoop, system prompt that says you are a helpful assistant, because we all know that that is really effective. ⁓ and then we'll have our max tokens. Now, ⁓ I'm clearly missing something. ⁓ we're missing ⁓ our result. We're gonna await that promise. ⁓ And when that result comes in, we're going to send a logging message that includes

That ⁓ server ⁓ server or let's see ⁓ response ⁓ from the model, which is gonna be our result dot content dot text. And that's what we want to send as our response. ⁓ this is also a fire and forget thing, so I'm gonna add a void there. A void, not a void. ⁓ And ⁓

Harsh Bharadwaaj (00:04:47)  
that should get us in a working state. So let's just clean up this because ⁓ no reason. ⁓ And we'll restart our server. We'll go to tools, we'll create an entry. Here's the title. Here's the content. We'll run the tool. We get our sampling request. Ta-da\! And we can say huzzah and approve it. And we should get a message. ⁓ but we didn't. ⁓ And ⁓ I know why. Because we didn't add that we support logging. So let's go back.

To our index. And ⁓ no, we don't need to do that. ⁓ let's ⁓ add logging right there. There we go. ⁓ And with that, now it should work. Let's restart and ⁓ yes, we now have the logging level. We can go to our tools, create an entry. Let's do this again, do do do, ⁓ and run the tool. We get our sampling, we say yay, approve it, and we got our logging message. ⁓ It gave us yay\! Hooray\! good. ⁓

And that is how you wire up sampling.

—---------

48

Harsh Bharadwaaj (00:00:00)  
So here we are in our delete tag tool. We want to verify that the user actually wants to con ⁓ to delete this tag. So we're going to first find out the capabilities of our client, because maybe the client doesn't support elicitation. If they don't, then sorry, we're just gonna go ahead and delete the tag and hopefully you click the right one. So let's get our capabilities. ⁓ And we're going to say agent server.server ⁓ get ⁓ client capabilities. ⁓ And what this is going to do is ⁓

verify or or give us all of the things that the client gave us ⁓ on initialization. And actually if you want to, you can take a look at the initialization ⁓ call here. And of course, it doesn't show us the capabilities that it sent. But I promise it sent capabilities during this initialization phase. It just isn't sending for us. So if you want, you can add a console log of the capabilities ⁓ right here. And we'll take a look at that here in a little bit. But I can tell you and the types will tell you

That ⁓ capabilities has a couple of things on it. We've got illicit ⁓ or elicitation, roots, and sampling as well as experimental. ⁓ we're gonna use elicitation. So if the capabilities has elicitation, now we know we can elicit input. Otherwise, we'll just skip the if and we'll delete the tag. ⁓ So now we're going to ⁓ do perform our elicitation. And the AI thinks it knows what we're doing, so let's see ⁓ if it gets it right.

So we're gonna ⁓ elicit input. Here's our message. Are you sure you want to delete tag? That's the name with this ID. ⁓ the schema, ⁓ the thing that we're requesting, the thing we want back from the client, ⁓ and that's actually a a fun thing to call out. This is the first time we're actually ⁓ making a request from the server to the client. So this is kind of fun. ⁓ the request schema is gonna be an object. We're expecting properties ⁓ confirmed. That is a type Boolean in the description.

Whether to confirm the action or whether to ⁓ delete the tag. ⁓ and then ⁓ we're what once we get that result, then this ⁓ promise will resolve ⁓ and we can continue. If the result action is accept and they confirmed, then ⁓ or and the confirmed is true, then we can proceed with the deletion. ⁓

Harsh Bharadwaaj (00:02:20)  
⁓ not sure I want to write it this way. So instead we're gonna say confirmed ⁓ is equal to ⁓ the result being actioned and the confirmed being true. ⁓ And I'm gonna say if it's not confirmed, then we're going to exit early. So we'll create a structured content. Success is false, it did not delete the tag. Here's the existing tag, ⁓ and ⁓ tag deletion canceled. That works just fine. So with that.

And actually let me call out also that accept, decline, or cancel. So in some situations you might say, they just canceled, so we'll do something else. Or they explicitly decline this, they do not want ⁓ to ⁓ to even answer us. ⁓ maybe they canceled because they didn't see it or timed out or something. ⁓ and then here ⁓ confirmed is unknown because the types aren't nice here, and it would be cool if we had Zod and we could have ⁓ good typing on that. So you might have to deal with some typing issues if you have a more complicated

elicitation schema. ⁓ Okay, great. So let's make sure that this is all working properly. We're gonna restart. I'm gonna s ⁓ scrunch this down. We'll go to our tools, we'll list our tools. And now ⁓ we're going to go to delete ⁓ tag and we'll delete tag with the idea of four. Run that. That's gonna send us right over to elicitations. ⁓ And here you sure want to delete that one and I'll confirm it and we'll submit it. ⁓ And boom, it was deleted successfully.

So awesome job team. That is elicitation. ⁓ Not the most fun thing to write out. I I really don't enjoy writing out a JSON schema by hand. So hopefully one day we can get Zod working in here ⁓ at some point ⁓ and get some type safety on some of this stuff. ⁓ But hopefully this gives you an idea of what is possible here. You don't necessarily have to do an elicitation during a tool call, but that's the most ⁓ common scenario where you would do that that I can think of.

Maybe in the future there will be other situations where it makes sense to elicit some input from the user. ⁓ The other thing that I'll mention here too is that this is not necessarily a mechanism for ⁓ a secure transfer of information. So you wouldn't want to ask for a credit card or anything from here because ⁓ somebody could install your MCP server ⁓ inside of some other ⁓ proxy or something like that, some other orchestrator. And so now ⁓ that content is going through that proxy and

Harsh Bharadwaaj (00:04:44)  
Yeah, you're in trouble. So ⁓ I wouldn't recommend using this as a mechanism for secure transfer of information unless you have total control over the environment in which this is being used. ⁓ There are other mechanisms for that, and hopefully in the future there will be even more. ⁓ That's like a ⁓ explicitly for secure transfer of information. But that's called out in the spec, and so I thought I'd mention that to you as well. There you go. That is elicitation, can be quite helpful.

—----

69

Harsh Bharadwaaj (00:00:00)  
Hey, well done on this exercise. You deserve a break, and you also deserve a chuckle. Here's a good one. Which side of the chicken has more feathers? The outside. Ha ha. ⁓ man, that that's kind of gross actually thinking about that too much. ⁓ okay, great work. ⁓ go ahead and move your body a little bit. Go give somebody a nice compliment, give your loved ones a hug, and get yourself a drink, write down the stuff that you learned, and whenever you're ready, you can come back and I'll be ready to teach you some more stuff about this. ⁓

—--------

56

Harsh Bharadwaaj (00:00:00)  
Sometimes your tasks take a little bit longer than your users want them to, maybe all the time. ⁓ And when they do, you want to be able to give the user some good feedback. If you can't make it any faster, at least give the user some feedback. That's what we're going to be talking about in this exercise. So, first, with long-running tasks, this is what the MCP spec allows for us to do. ⁓ So when a client makes a request,

that it wants to be able to be notified of progress updates on, it will include a progress token. ⁓ The server can then take that progress token and to send notifications back to the client ⁓ as the ⁓ progress continues. And that can include the progress, the total amount. So if you're going ⁓ from ⁓ zero to one or from zero to a hundred or whatever, the progress would be a fraction of that. ⁓ the total amount would be

The one or the 100, and then a message indicating what step we're on or or what ⁓ the ⁓ what is currently going on or whatever. ⁓ And with that, then the client can forward that along to the host application, the host can do something useful with it. ⁓ show a status bar or or something like that. That just kind of helps the user know that yes, things are happening, I'm not frozen or anything. We're we're processing your request. And that can be quite useful. And then finally, the server sends the response.

So that's ⁓ and that typically is for ⁓ for tool calls. ⁓ So ⁓ as far as the actual contents of the requests here, the client will send the request with a progress token under the meta property. And then ⁓ the server will send a response ⁓ with the progress token and then the progress total and the message. ⁓ You don't have to worry so much about this. The SDK is gonna help you manage that progress token. You'll just get that from

⁓ the in your request handler and then you can send those notifications and the ⁓ SDK actually gives you a utility for sending those progress notifications. So it makes it pretty straightforward. ⁓ The biggest challenge is knowing how to retrieve the progress. And we'll be doing that in this exercise. ⁓ And then finally when the server is all done then it sends its its normal ⁓ tool response or whatever.

Harsh Bharadwaaj (00:02:13)  
Then there's the cancellation side of things. So sometimes the user might start something and then decide, no, never mind, I don't want to do that anymore. And you're still chugging away at their request and wasting resources for something that they don't care about anymore. ⁓ And so ⁓ there is an a ⁓ an API or a spec for this specifically as well. The client sends a request to the server with an ID, ⁓ and ⁓ it is ⁓ that that request is in progress. You actually don't necessarily have to implement

the progress token and stuff for you to be able to handle cancellation. That's just a fun fact. ⁓ And so that's not pictured in here. ⁓ At any point in time the client can send a notification canceled that will tell the server, hey, you know what, I don't I don't really care about this anymore. ⁓ And so then the server can say, okay, that's fine. I'll I'll stop whatever it was that I was doing ⁓ as a result of that request. And now we're saving ourselves some resources.

And it doesn't have to send any sort of response to that notification. ⁓ So that is pretty neat. ⁓ And the ⁓ way that that works, the way that looks, it's just the client sending this notification. ⁓ Here's the request ID that I no longer care about. Here's the reason ⁓ that I don't care. ⁓ And maybe you want to log that or something ⁓ and see why are people always canceling stuff? It could be.

So, one thing that this ⁓ uses is something called an abort controller and abort signals. And some of you may not be very familiar with this. It's not a super common, it's relatively common, but not super common in the JavaScript space. So I've got this example here if you need to look into that a little bit ⁓ further. So you create ⁓ an abort controller, you get that controller, and that gives you a signal. ⁓ And then ⁓ you're not managing the abort controller, the SDK is going to manage that for you. ⁓ and it has

Who knows how it's controlling all that? All you get is the signal. ⁓ And with that signal, you can add an event listener for the abort event. And if that happens, then you can send a sig term to the child process or you can ⁓ do whatever else. ⁓ like you ⁓ made a really long database query or something, you can ⁓ stop the database query, whatever it is that you're doing ⁓ as a result of that ⁓ abort event.

Harsh Bharadwaaj (00:04:29)  
⁓ and here we have this ⁓ do long task. We've got our output, all of this stuff going on, ⁓ handling all of that stuff. We don't have to worry about ⁓ all of that stuff running anymore. ⁓ and here th this is like a complete example. So this is how ⁓ the ⁓ abort controller actually triggers the abort event ⁓ is by taking that controller and saying dot abort. So you could just copy paste this somewhere and run it and see like all the console logs and stuff.

I hope that's helpful to you if you are unfamiliar with abort controller and signals and stuff, because we're going to be using that in part two ⁓ of this exercise. ⁓ So, with that, hopefully this gets you up and running on how to handle long tasks and give the user some really good ⁓ user experience with that, ⁓ and how to save yourself some resources if the user cancels on you. Have a good time with this one.

—------------  
64  
Harsh Bharadwaaj (00:00:00)  
I know it's not technically necessary, but I just get good feelings when I add list changed here. The reason I say it's not technic technically necessary is because the SDK itself supports being able to ⁓ send list changed events. And so by default, if you look at the ⁓ our initialize ⁓ request response here, you'll see we already have prompts list change and resources and tools list change, but I just it feels good to be

Explicit about that. Yes, I have prompts and I do change things sometimes. ⁓ I don't know, maybe you you'll disagree with me, but I like to do that. ⁓ So here now we're going to assign the result of register prompt to suggest tags prompt, and we're going to use that to enable and disable ⁓ the prompt. ⁓ And now for us, we want to enable and disable this when we ⁓ have or don't have tags and entries. So let's make a function update prompt.

Prompts. ⁓ And here we're going to call ⁓ here, we're going to need this to be async. ⁓ We're going to call the database to get all the entries. And if the entries length is greater than zero ⁓ and the suggest tags prompt is not enabled, then we're going to enable it. So now we have entries ⁓ and it's disabled right now. Let's enable it. ⁓ Otherwise, ⁓ if the ⁓ suggest tags prompt is already enabled, then we're going to disable it.

Now I think that the SDK hopefully will soon make it easier for us to do this, so we can just ⁓ do this. ⁓ But as of the time of this recording, if you do that, it will ⁓ trigger a list changed event ⁓ even or like the prompt list changed, even if ⁓ nothing actually changed. So I like to be a little protective here. ⁓ something to check on ⁓ at the time of ⁓ you're watching this video on whether or not that's still the case. ⁓ So

Now we have a mechanism for updating the prompts based off of the current entries. We just need to ⁓ add our subscription. So whatever database you are using at ⁓ at the like for the server that you're building, ⁓ you probably for you to be able to do something like this, you're going to need to have some mechanism to subscribe to, updates to the underlying data. ⁓ And we've got that with our little database here.

Harsh Bharadwaaj (00:02:19)  
⁓ yeah, everybody is gonna have a different scenario. And so I'm not going into the details of how that works. You feel free to dive into it if you want. ⁓ but yeah, every database is gonna have its own mechanisms for doing that, or maybe you have a data resource that's external, you're not talking to a database directly. You're gonna need to have some sort of subscription mechanism to be able to do this sort of thing. ⁓ And with that now, we can come over here to our ⁓ prompts. We can list prompts.

And you'll see we don't have any prompts. We'll look at the list. The prompts, list of prompts is empty. ⁓ And so now if I go to list tools, we'll create an entry. Let me get rid of that. We'll say ⁓ here's a new entry. We'll run the tool, and you'll notice we have a new notification. Notification prompts list changed. Ta-da\! And so now ⁓ the the ⁓ MCP inspector doesn't actually do this automatically for us, ⁓ but

⁓ the client might see that notification and be like, great, let me go and trigger a list prompts. And then it will get the latest prompts. And now we can actually ⁓ supply, you know, get our suggested tags and stuff. ⁓ So that's the basic idea of the list changed event. You have the same sort of thing with resources ⁓ and with with tools. It's quite nice. Resources have kind of a special ⁓ side of this though. So we're gonna look at that here in just a second. But hopefully that gives you a pretty good idea of ⁓

what you can do with the list changed event as you know making a really interactive dynamic server can be quite useful. Have a good time ⁓ adding all of this to the tools and resources if you want to, and we'll see you in the next one.

—----------  
50

Harsh Bharadwaaj (00:00:00)  
Alright, so check this one out. Could you please write a fictitious journal entry about a day skiing in the mountains for me?

⁓ and save it in EpicMe, ⁓ just in case. It might just like output it. ⁓ So let's see. It's going to normally you're gonna write your own journal entries, but here, yeah, it's unforgettable skiing. Awesome. We're gonna allow this to create the entry. And then what is this? interesting. We we still get the completion right there. So the tool is still finished, ⁓ but we're getting this pop-up, allow in this session the MCP server advanced MCP features. That's the workshop we're doing right now.

Has issued a request to make a language model call. Do you want to allow it? Yeah, sure, I'll allow it. ⁓ And okay, yeah, what happened? Well, VS Code could maybe do a little bit better job of showing you what happened ⁓ and giving you an idea of the experience. ⁓ But I'll tell you what happened, and that is we had a sampling request. So if I go over here and we go to show sampling requests, then I'm gonna see we had a sampling request. What does that even mean? Hold on a second, it's talking about ⁓

We've got a name scheme. That looks like a tag. Okay, hold on. ⁓ Let's look at ⁓ this. It created the entry. It didn't give it any tags though. So we're talking about tags. What tags do does that entry have that you just made?

And let's see. ⁓ no tags. ⁓ Could you actually look it up for me, please? So it doesn't know, but I think it has some tags. ⁓ that's fascinating. Huh. How cool. So that's what the sampling request did. ⁓ we have a prompt that allows you to suggest tags for a particular entry. But what's even cooler than having the user use a prompt is just automatically creating the tags for them.

Harsh Bharadwaaj (00:01:53)  
when they create the journal entry in the first place. ⁓ And that requires an LLM for understanding what the journal entry is all about. ⁓ And I don't want to pay some other LLM and make my user pay for it. The user is already using an LLM. So let's just have them like borrow theirs and ⁓ use their tokens. They're already paying for it. So we'll just borrow it and enhance their experience with the LLM they're already using. So that's what that request was was asking if we can

use their LLM to generate the tags that are ⁓ most relevant to this journal entry. And yeah, we've got adventure, nature, skiing, and travel. That all makes sense for this journal entry. ⁓ So that works out pretty well. And VS Code supports sampling. Not a whole lot of clients do at the time of this recording, but more of them are all the time. So ⁓ and as you could see the ⁓ experience wasn't altogether like the best. I couldn't see what was it going to request and ⁓ what was the response or what was the result or whatever.

we do actually log ⁓ and kind of send a notification that it was finished and what was added, but we didn't see any of those logs in our output in the UI here. It's kind of a hard problem, so hopefully the user experience does get better. ⁓ But it is pretty cool that ⁓ we do have the ability to ⁓ to borrow somebody else's LLM to ⁓ enhance their experience with the LLM that they're already using. ⁓ So that is what sampling is all about, and you are going to be implementing it in this exercise.

So it is a little bit ⁓ different kind of exercise. Let's talk a little bit about how all of these pieces work together. ⁓ so actually, first we'll look at ⁓ here it is. This is the workflow for how all of this is ⁓ gonna work. So ⁓ notice interestingly that the server is the one initializing the request. Normally, with a workflow like this, you're gonna start with the thing on the left that triggers everything. But all of our diagrams have had this format, and so we're gonna start with server on the right. Things are gonna start.

From the server. So we're going to create the sampling request that's going to include the system prompt and the messages that we want to ⁓ give to the LLM. ⁓ That's going to go to the app. The app is going to send it to the user to ask, hey, are you cool with this with them borrowing your LLM? ⁓ Because the user will probably have to pay tokens for that and things. ⁓ And then the app, if the user approves, the app is going to make a request to the LLM. ⁓

Harsh Bharadwaaj (00:04:17)  
The that's going to include the system prompt and the messages. It's also going to take into account potentially ⁓ some hints around the model that should be used and different than like ⁓ preference on ⁓ speed and cost and intelligence ⁓ and temperature, stuff like that. So you can configure all of those things. ⁓ and then the LLM is going to generate the completion, send it to the app. The app will send that sampling result to the client, which then sends it to the server.

And of course the user can reject, so if it's denied, then that goes to the server. And the server's like, well, ⁓ that's fine if you don't want to have this nice ⁓ thing that I was gonna make for you, whatever. ⁓ So that is the flow. The way that the ⁓ the messages are structured are just JSON RPC. We've got the sampling create meth message. Remember, this is actually a request from the server to the client. So sampling create message, here are the messages, here's the system prompt, here's the tokens that I expect this.

response to ⁓ come back with. ⁓ And this is the message is just a list of messages, very similar to what we have with prompts. ⁓ And then the response from the client ⁓ in the event of a successful generation ⁓ is going to include a result that has a role of assistant. And here's our content, the type ⁓ is text and here's the text. ⁓ here's the model that was used, here's the reason we stopped. That's the typical reason it stops, is because the ⁓ end turn. There are a couple other stop reasons, but this is really the only one that we are ever going to be concerned with.

And then ⁓ we also are going to be talking about logging, because we want ⁓ we this is kind of a side effect, it's just gonna go off. but we do want to let ⁓ you know the client know, hey, this is what happened as a result ⁓ of this sampling request. And so we're going to enable logging in our server ⁓ and send a notification. And that notification is gonna include some data that says, hey, here's what ⁓ what happened as a result of that sampling. ⁓ and this is that's what the logging message looks like.

We've got some utilities for making that easy from the SDK. ⁓ So I think that's everything you need to know about this. I'll show you how that all works in the inspector. You're gonna be the LLM in the context of the inspector. ⁓ So ⁓ with that, ⁓ you should have everything you need to get going on this exercise. Sampling is way cool. So I hope you enjoy it.

—-------------  
67

Harsh Bharadwaaj (00:00:00)  
Let's say that I'm talking to the LLM and I want to talk about this prompts file. Like in prompts I do ⁓ this. And then you know what? I need to change this ⁓ thing first. ⁓ And ⁓ then I'm gonna have a more conversation. And then I'm gonna change another thing. And I'm going back and forth. By the time I actually submit ⁓ my conversation, I wanna have the latest version of the file. And in addition, as I continue to have a conversation about it, I wanna make sure that the LLM and I are talking about the same file and that that file is kept up to date.

And in the same way, if I'm if I've got a resource in my conversation with the LLM, I want that resource to remain up to date as well. So what we're talking about here is subscriptions. I want to be able to subscribe to a particular resource ⁓ and have the client be able to ⁓ keep that updated in the context of the conversation. And so as a server, I need to make sure that I support those types of subscriptions. So that's what we're gonna do in here. And when you're all finished, then you'll be able to list resources.

We'll go to family, and here you'll notice we have a subscribe. So we click on subscribe. That sends the resources subscribe request. ⁓ And then we can come over to our tools. We'll list our tools. We're going to update ⁓ tag number five. That was the one we subscribed to. And we'll capitalize the family and make an exclamation point. We're going to spell it right too. ⁓ Make an exclamation point. ⁓ Run the tool. ⁓ And there we go. We have an updated family. You'll notice we get a resources updated.

⁓ server notification right here. And so then the client can say, you updated, let me go and retrieve the latest version of this resource, ⁓ which we can ⁓ symbolize here by hitting refresh. Now we get the latest version of the resource. ⁓ And then we can keep that ⁓ history up to date. And so the conversation is always referencing the latest version of the resource. This is important for really ⁓ advanced and dynamic and powerful interactive ⁓ experiences with our agents.

And so that's why we're going to learn about it right now. So I'm gonna send you off on that one and we'll see you when you're done.

—-----------

47

Harsh Bharadwaaj (00:00:00)  
All right, let's pretend that I'm a user and I come over here and I'm like, I want to delete a tag. Let's go to the delete tag thing and I'm gonna say delete tag two, run the tool. ⁓ shoot, no, that was not the tag I wanted to delete. no, I'm in trouble. So this is one situation where elicitation could really help us out. And so when you're finished with this exercise ⁓ and somebody comes over here and they say, I want to delete tag number three, we're gonna run that tool.

And ⁓ we're taken over to the elicitations right here. And here it's saying, are you sure you want to delete tag travel with ID three? ⁓ And ⁓ this is the request schema, what it's expecting to get back. I'll get this little form. Thank you, MCP UI. We'll hit yeah, check that checkbox confirmed. We're gonna submit that, ⁓ and then we get success. It actually did work. Hooray, we're in a good spot. ⁓ we can also, you might have noticed, let's just try this again. ⁓ we can also decline or cancel.

Both of those are distinct operations that we might take a different direction based on those. ⁓ So there are actually like four different things. We can submit with confirmed, we can submit with not confirmed, we can decline, and we can cancel. ⁓ And each one of these technically has a different ⁓ intent from the user, and we may take a different path. For ours, it's pretty simple. Like if we are only going to delete ⁓ if they confirm that they're deni ⁓ that they don't want this.

The other thing actually to consider here too is whether the client is capable of handling elicitations. ⁓ Because some clients might not. ⁓ And so we've got a couple of if statements we're going to be putting in here throughout. ⁓ And I think you're going to have a good time with this. You can also, of course, ⁓ accept different inputs and things like that. It's more than just confirmations, ⁓ but confirmation pretty straightforward. The one problem here.

Is that we are not able to unfortunately ⁓ use Zod to configure our schema. So this we're gonna have to configure manually, which is a bit of a pain, but it's it's gonna be just fine, I promise. ⁓ Okay, with that, I think you're ready to go. So let's get right into it.

—------------

49

Harsh Bharadwaaj (00:00:00)  
Alright, awesome job. Let's ⁓ take a break, write down the stuff that you learned here, ⁓ and let's have a dad joke. What is the advantage of living in Switzerland? Well the flag is a big plus. Haha. Yeah, because the flag looks like a big plus sign. ⁓ Also, ⁓ different shape than lots of other flags. ⁓ so that's interesting too. ⁓ so yeah, that I hope you had a good time with that exercise. Now is the time to write down what you learned so you don't forget.

get yourself a drink of water, get your body moving or something, so that you can get f blood flowing and ⁓ your brain ⁓ doing good things. ⁓ And yeah, when you're all ready for the next one, I'll be ready for you.

—------------  
60

Harsh Bharadwaaj (00:00:00)  
Alright, let's start out in here ⁓ by accepting that signal, and we've got that abort signal right there. ⁓ And if we start the function and it's already been aborted, then we're going to ⁓ throw an error. ⁓ In the mocking case, we're going to throw an error anytime as we're going through this loop. If it was aborted, okay, let's just stop going through the loop and throw an error. ⁓ And then we can get to the actual like real-world implementation sort of thing. So we'll start with the ⁓ on abort handler. ⁓

This is going to take the ⁓ the ffmpeg subprocess and send the sig kill to it. ⁓ So ⁓ then we can add the ⁓ on abort handler to the abort event on the signal. So if that signal's abort controller ⁓ has the abort event, then we can abort the ffmpeg ⁓ generation and save ourselves some resources.

we definitely do also want to clean up after ourselves. So we're gonna add a void to this. ⁓ so whether or not the FFmpeg promise resolves or rejects, it doesn't make a difference. We want to remove the event listener here so that we don't have to ⁓ worry about memory leaks and all of that stuff. This is you know good ⁓ JavaScript hygiene related stuff that we gotta do there. ⁓ We're not quite finished though. ⁓ we also need to ⁓ reject the promise if the ⁓ signal was aborted. And so

That can happen here where we say, okay, send the sig kill command, and then we're gonna get our close. ⁓ so we'll reject the promise so we know, hey, yeah, video creation was canceled. ⁓ So now that we've got all this wiring in place for our FFM peg stuff, let's go back to the tools and we'll grab the signal ⁓ that we get as the second ⁓ parameter to our handler, ⁓ and this is going to be attached to abort an abort controller.

Which should abort ⁓ when the request is canceled or whatever the case may be. And the SDK is gonna handle that for us. ⁓ So then we forward that on to our create wrapped video. ⁓ And ⁓ now that's all wired up together. So if the user decides, hey, you know what, I don't care about this video anymore, then we don't have to continue processing that when we ⁓ send the sigk command to the ffmpeg subprocess.

Harsh Bharadwaaj (00:02:15)  
And like I said, this is kind of difficult to test. There is not yet a button here for canceling an ongoing request, but hopefully there will be soon, or maybe you can contribute that. Or maybe by the time you're going through this, there is a cancel button. So you can try it there. ⁓ But that is how you ⁓ wire together the signal that you get from ⁓ your request handler ⁓ for this ⁓ for your tool. ⁓ you take that, you add some ⁓ event listeners to it, you check for its aborted status.

And ⁓ you act accordingly.

—-------  
63

Harsh Bharadwaaj (00:00:00)  
I'm gonna do something a little bit tricky here. I'm gonna come over here and delete ⁓ the whole SQLite database. ⁓ technically, this will still work because SQLite is amazing. But I'm gonna come over here to Tools. We're going to get an entry, or here, let's list entries. We'll run the tool. There are no entries. What if I ⁓ okay, so does it make any sense to get an entry? Or does it make sense to update an entry or delete an entry?

⁓ no, it does it and it certainly doesn't make sense to get or ⁓ update or delete tags. maybe it still makes sense to list entries and tags, possibly, but there are no entries and no tags, so why do we have those? In if we were building a web UI, we would have some sort of ⁓ empty interface. We w certainly wouldn't show buttons for editing and deleting things, because there's nothing to edit or delete. ⁓ And so in a similar way,

Having ⁓ some mechanism for dynamically updating the tools that are available ⁓ seems like a pretty good idea. ⁓ you would save on the context window that ⁓ shows up in the LLM as it's trying to decide what tools to use. ⁓ Why provide the context for a tool that they just can't use? Now your instructions should be able to in insinuate or or kind of help the LLM know what tools would be possible cons ⁓ if ⁓ we had some entries and state and things.

But we don't have any, and so we may as well not include the tools themselves. ⁓ Same thing goes with ⁓ resources and prompts. Doesn't make any sense to be able to suggest tags if you don't have any ta ⁓ things to suggest tags for, or even any tags to be suggested. So ⁓ what we're gonna do in this exercise step ⁓ is ⁓ disable those things so that they don't even show up if you don't have the data to back them up. And if you want to test that out, you can just delete the database.

And ⁓ you'll be all ready. If you ⁓ want the database back, then you can just reset the playground ⁓ and you should get that database back ⁓ so that you can test ⁓ before and after. ⁓ we're gonna be using the list changed event for this. This is going to be ⁓ handled for us. we're gonna focus on the prompt side of things ⁓ just to keep it easy. ⁓ it's pretty much the same for tools and resources as well, so feel free to take it even further if you want to and apply it to those.

Harsh Bharadwaaj (00:02:19)  
context, ⁓ but we'll just do prompts to make it easy. There is a mechanism for ⁓ being notified of changes to ⁓ things in the database ⁓ and you'll have all of the code comments and stuff that you need to be able to be successful doing that. ⁓ So I'm gonna leave you to it. I'll see you when you're done.

—-----

58

Harsh Bharadwaaj (00:00:00)  
I'm gonna come here into the video and add progress reporting from the video TS, and then we can use that to send notifications. So here we're going to add an on progress callback, ⁓ and that's gonna be ⁓ on progress there. And we'll accept that there. And then we've got two things to think about. First is the mock time. So if you specify a mock, we need to call on progress here, and we'll wanna call on progress when it's all done to say, yep, we're all finished.

And then ⁓ here below, this is all the FFmpeg related stuff. We're gonna spawn the process, all of that. So we do have a mechanism for parsing out the progress right from FFmpeg's output. ⁓ I didn't think it was all that useful for you to ⁓ figure that out on your own. And so Marty the Moneybag is in here giving us exactly how to compute that from the output, ⁓ and then we can just simply ⁓ use that ⁓ progress and report that progress back.

And then of course when FFmpeg is done, then we'll report it's completely finished. ⁓ So now we can come over here to our tools where we have our create wrapped video. ⁓ And we're going to need to grab the send notification and underscore meta for progress reporting. So we're going to grab those. That's coming as the second argument to your handler in the SDK. There are a number of things that come through here. I think this is the first time for ⁓ many of you to ⁓ see this, but you also get request ID. ⁓

You get send request, a bunch of other things that come straight through this. So it's kind of handy. ⁓ so we're gonna say send notification is what we need for what we're doing. ⁓ we need to add an on progress handler. And it's almost got it. ⁓ we've got all this. Here we go. ⁓ Okay, and then we can get rid of this. And let's just see if this does that. ⁓ the red underlines here are because ⁓ send progress is actually async.

And so we could await this, but this is a fire and forget thing, so we're gonna do ⁓ void instead and just say, ⁓ I don't care about the promise that comes back from this. I'm not gonna do anything with it anyway. So I'm just letting you know. So with that then ⁓ we've got our ⁓ notification. Here, I think the red underline here is coming because ⁓ the progress token has to exist. We can't just like

Harsh Bharadwaaj (00:02:18)  
⁓ send a notification if it doesn't exist. So what we're gonna do is get our progress token from Meta. ⁓ And if that doesn't exist, then we're just gonna return. Then we can use the progress token like that. ⁓ There we go. So if you have sent me a progress token, that means that you're interested in getting progress updates. If you are, then anytime there's a progress ⁓ update, then we're going to send a notification with notifications progress. ⁓ Our params will include the progress token.

The progress amount and the total that that amount is out of, ⁓ and then a message creating video. ⁓ we could potentially, I suppose, also have a message here, and then you could have one that says video created or something like that. But ⁓ we're gonna just stick with the easy thing. ⁓ And that is it. So now we come over here to our ⁓ server. We can ⁓ restart, we'll go to tools, we'll go to create ramped video.

I'm gonna set this to take five seconds. We'll run the tool, ⁓ and we'll see we're getting those progress notifications coming through. So that works out nicely for the user. It's a better user experience to get those progress notifications. And all we had to do for that was make sure that whatever asynchronous thing that we're doing supports progress notifications. So we had to add that manually ourselves. Of course, we have the mocked thing. You typ don't typically do that. ⁓

But the more typical scenario is you're spawning out to a process, you have to have some mechanism for getting updates. ⁓ Maybe you're doing multiple step process and so you can kind of make up the percentage if you like you just do the very best that you can. Maybe you're making a fetch request and there are some progress notifications that you can get from that or XHRs. ⁓ just have some mechanism to be notified of the progress as you're going through this.

And then once you have that, then you can send those notifications as you get them.

—--------------

59

Harsh Bharadwaaj (00:00:00)  
Let's say that the user has asked you to do something that's really computationally expensive for you, and then they decide, you know what, never mind, I don't want to do that. And they close the tab or or whatever, and now you're left generating this thing that the user doesn't actually care about or need anyway anymore. ⁓ And there is a built-in mechanism for this, which is an abort controller and an abort signal. ⁓ And if we can take that signal and listen for when that signal has ⁓ been aborted.

Then we can stop whatever processing it is that we're doing. So in this exercise step, we're going to have you ⁓ listen for those events ⁓ and respond so that you ⁓ kill the subprocess or or stop ⁓ the mock listening ⁓ loop that we've got in place so that you don't ⁓ expend resources on things that the user doesn't care about anymore. ⁓ this one's a little bit harder to ⁓ simulate or or ⁓ demonstrate in here, so you're just gonna have to take my word for it. ⁓

So good luck on this one. We'll see you when you're finished.

—------------------

42

Harsh Bharadwaaj (00:00:00)  
Tool annotations are quite ⁓ useful, but they're a little bit confusing as well. So Kelly, the coworker, ⁓ put together this handy annotations type that we can ⁓ use as we're specifying our annotations. What this does is it ensures that you're only setting a configuration that is meaningful. ⁓ So here the open world hint, that can be changed to false. ⁓ It defaults to true. So if you're specifying it, then there's like, why are you doing that? You should only be specifying it to false.

And read-only hint ⁓ is ⁓ defaults to false, so you can only set that to true. And if you do, then there's no reason to specify destructive or idempotent because those things don't make any sense in the context ⁓ of a read-only true. If you do set one of these things, then you can't set read-only true. right. So ⁓ this type is going to help us make sure that we're setting all of our annotations to ⁓ values that actually are meaningful. ⁓ So starting with create entry.

Does create entry destroy anything? Let's let's start like from the top. Is it open world? ⁓ No, it's ⁓ it's not open world. ⁓ this is only interacting with our service. It's not going out and making Google searches or anything like that. So we don't need to worry about ⁓ this being like an open world sort of situation. ⁓ let's think about the next one. So is it read-only? ⁓ and that's that's actually the next decision. If it's read-only, then we're good.

If it's not, then we have to ⁓ maybe configure these other ones. So ⁓ is creating an entry read-only? No, we're actually creating an item in the database. So now we we need to decide, is this destructive? No, we're literally not destroying something, we're creating something. ⁓ And then the last thing, is it itempotent? If it is itempotent, ⁓ that means that we can call it over and over and over again and expect the same result. ⁓ I g ⁓ not really. We're creating new entries in the database. So if you call it over and over and over again,

the number of entries we have is going to increase. So we're not going to end up with the same result. Item point is a little bit confusing and and if it's a new concept for you, then ⁓ I understand spend many years studying the ⁓ magic of the the programming and stuff. And maybe it'll make sense. ⁓ it's a little bit confusing to me as well. But effectively it's just you can call this over and over again. It's not necessarily a pure function or anything. It just means you'll ⁓ the ⁓

Harsh Bharadwaaj (00:02:23)  
The final results will not change regardless of the number of times it's called. That's ⁓ basically the idea. So creating an entry is it itempotent? No, that's not. So ⁓ but that's the interesting thing here. So if we were to do ⁓ to specify these, we're gonna say, no, it's not destructive and itempotent, it's not itempotent, but we don't need to specify that because that's the default. And there you go, it does not satisfy our tool annotations type. We do not need to specify it. So that is quite nice.

Thank you very much, Kelly, the coworker, for that. So then we can continue on to the rest of these tools. But a lot of this is just gonna be the same thought process. So ⁓ I'll go through it, but you can feel free to skip to the end if you're ⁓ if you kind of get this. ⁓ So get entry. Is that open world? No, it's not gonna be open world. ⁓ and is it read-only? Actually, yes, it is. So we don't need to specify the other ones because this one is read-only. ⁓ we're only reading from the database. ⁓

The list entries, same deal. We're in our own world, we're only reading entries, so that's fine. Now, updating entries. ⁓ this is also not open world. All right, so let's talk about destructive and item potent for a second. Destructive, I think, ⁓ that not necessarily is this destructive. ⁓ while you are technically losing the previous content if you're updating it to new content, I don't think that is in the spirit of what the destructive hint is about. So we're going to ⁓ have that one be configured as false.

Now, item potent, ⁓ I think this one actually might be true ⁓ true because while we're ⁓ like if we call it a million times, we are technically continuously updating the updated at ⁓ time period. So like you call it three times versus four times, that updated at is going to be slightly different potentially. But again, in the spirit of what this is intended to be, ⁓ if you call it ⁓ once or a hundred times.

The end result is basically the same. The the updated at time, ⁓ I guess it kind of depends on ⁓ the length of time. But like what if you are talking about a time frame of three seconds ⁓ and you update ⁓ one time at the very end, or you update six times throughout that period, as long as ⁓ like the end result of the item being updated is the same, then I think from a pragmatic standpoint we can call this one ⁓ item potent. So we're going to configure it.

Harsh Bharadwaaj (00:04:44)  
such or thusly. ⁓ So here, the AI thinks it knows what we want to do here. Let's let's see if it got it right. ⁓ Deleting entry. ⁓ open world. No, that's not open world. Destructive. Yes, that actually is destructive. We are literally destroying an entry. But we're getting a type error because our type in it or our tool annotations is telling us, hey, this is actually the default. You do not need to configure this. So let's get rid of it. ⁓ Now is it item potent? ⁓ actually no it's not because here in our case we ⁓

are checking whether there's an existing entry. So if you call it once, there's an existing entry, we delete it. Call it again, there's not an existing entry, we have an error. And so in our case, no, it's not item potent, it's false, but again, that is not part of our types. So we'll just not configure that and we'll get the default of it not being itempotent. ⁓ Okay, continuing, we've got create tag. So this one, again, not open world. It's inside of our own service. ⁓ We're creating something new, literally not destroying anything.

but it is also not itempotent. So yeah, ⁓ it's it is not destructive. If we say item potent is false, we're gonna get a type error because that is the default. So we'll just get rid of that. Not destructive, but yeah, it's open world. Okay, now we'll configure get tag. This is actually gonna be really similar to what we had above with getting entries. Yes, it or no, it's not open world, but yes, it is read-only.

And same thing with listing. It ⁓ is not open world. It's in our own little world. ⁓ But it is read-only. Now updating, ⁓ same story as above. ⁓ Well, ⁓ nah, ⁓ you almost had it. ⁓ So ⁓ no, it's not open world. But it is ⁓ and it ⁓ it's not technically destructive, as we've decided. ⁓ and it is item potent, as we've decided with the update entry. So ⁓ it wants us to go back and delete that. I'm not gonna go back and delete that.

Okay, so with delete, we're gonna have open world is false. It's our own little world. ⁓ And no, it's not item potent because we're checking whether the tag exists. And if it ⁓ if it does the first time, it doesn't the second time. And so the result is going to be different. We're gonna get an error. So no, this one's not idempotent. ⁓ it is destructive, but again, that is the ⁓ default, and so we're not going to specify that. ⁓ okay, now add tag to entry. ⁓ That's not open world.

Harsh Bharadwaaj (00:07:06)  
⁓ is it destructive? No, we're literally creating a new record in the database to associate these entries with or this tag to this entry. So no, not destructive. ⁓ and is it itempotent? ⁓ Yeah, like we're we're not gonna create ⁓ a bunch of entries in the database over and over and over again. ⁓ it's going to just be the one entry, so it'll be fine. so yeah, that's itempotent. ⁓ okay, so now we've got the wrapped video. ⁓ in this case

it's not open world, though like we are using FFmpeg. Is that outside of our service? No. We're we're we're gonna include that in our service. ⁓ it's certainly not read-only, so we do want to configure the others potentially. Is it destructive? No, this isn't a destructive thing. We're creating something new. ⁓ Is it item potent? ⁓ so like the ⁓ it kind of depends maybe on the implementation. ⁓ In our implementation, it just creates a file with the file name set to the year.

And so it's just gonna override that file. and the the final result will be the same, ⁓ assuming that nothing else has been changed in the environment, which item potent that's satisfies that. So yeah, this is item potent. We're gonna create the same video every single time, ⁓ assuming that you're not like changing your ⁓ changing the full year or whatever. ⁓ so ⁓ that covers all of our annotations. ⁓ I know it was a lot, but hopefully that gives you an idea.

Clearly there's some like nuance and wiggle room to some of this stuff. you may even disagree with me on some of these things, but remember that the goal of annotations is just to communicate well to ⁓ the client to help the client make good decisions ⁓ about like how it presents different things to the user. Maybe on a destructive thing it's gonna show a little red button in and if you say it's not destructive, then it'll be blue, and maybe if you say it's read-only, then it's gonna be green or something like that, right?

So just little things that the ⁓ client can do to improve the user experience. That's all what annotations is all about.

—----------------------

61

Harsh Bharadwaaj (00:00:00)  
Great work on that one. Here's a dad joke for you. Why did the worker get fired from the orange juice factory? ⁓ Lack of concentration. Haha. ⁓ Not from concentrate. ⁓ All right. Now is a good time for you to go grab yourself a cup of orange juice ⁓ or ⁓ some other beverage that will revitalize you and get you ready for ⁓ whatever else you have on the plan for the day, whether it be the next exercise or whatever else you're gonna be doing. But ⁓ you should probably ⁓ get your body moving in some way to get blood flowing.

and ⁓ write down the stuff that you learned because that's important for your learning and retention. ⁓ And we'll look forward to seeing you next time you start another exercise very soon, I hope.

—--------------  
65

Harsh Bharadwaaj (00:00:00)  
Alright, I deleted the database again, and so if I say list resources, we're not getting any resources because we added the ⁓ list changed support to our resources. This makes sense. We don't have any tags. But if I go in here and I say, hey, let's create a tag, we can do that. We're going to create a tag called one ⁓ and the ⁓ first. Hooray, run the tool. Now we get list changed for our resources. Awesome. ⁓ we can take a look at resources now, list the resources, and

Boom, we've got tags and we've got one. And we also have a template ⁓ for tags. ⁓ And so we can say ⁓ one and boom, read the resource. Perfect. This is exactly what we want. However, what's interesting is ⁓ we we now have access to these resources and the resource templates. But what if I create another ⁓ tag? So let's create tag two. This is the second.

And run the tool again. ⁓ there was no list changed event. But if I go to resources, ⁓ I probably should have gotten a list change because this list should change, right? We we now have another resource. And of course, if I clear this and list the resources, it's gonna be there. But there was no notification that that was going to be the case. And so there's a difference between ⁓ the resource ⁓ category is available ⁓ and the resource ⁓ like

⁓ a an instance of a resource template is available. ⁓ And so we need to have notifications for when instances of a resource are available. ⁓ And that is what you're gonna do in this exercise. So ⁓ go to the resources, add support for subscribing to those types of changes and ⁓ notifying of those types ⁓ of list changed events. ⁓ And we'll see you when you're done.

—------------------

41

Harsh Bharadwaaj (00:00:00)  
For clients and LLMs to be able to work with our tools effectively, it can be really helpful for them to know whether or not they can call it as many times as they want without causing any sort of side of or bad effect. ⁓ Or whether they ⁓ when calling it are going to be destructive, or if it's going to reach outside of ⁓ into the outside world or or whatever. It can be really useful for the LLM or for the ⁓ client to know what is going to happen when they call a tool.

So that they know maybe how much the human needs to be in the loop or or whatever the case may be. Now, these annotations that we're going to be adding aren't necessarily ⁓ like you shouldn't really rely on them as being like a security measure. It's mostly a hint to make the user experience better. ⁓ But that is what we're gonna be doing in this exercise: going through all of the tools and adding appropriate annotations. Now, one of the tricky bits of this is just the default behavior.

And you might say, well, let's just specify the annotation for every single one of these tools. Every annotation is specified. ⁓ I don't do that in it here, but ⁓ you might think to do that once you get through it and see like, wow, if this is if this is the default, then I don't need to specify this, whatever. ⁓ But yeah, let's take a look at what this is going to look like when you're finished, and then I can get you off and running on this one. ⁓ So here we're going to connect to our server. If we look at our tools and list of tools, that's the point where we'll we should see the annotations.

Right now we just see the name, title, description, and input schema. ⁓ When you're all finished with this, you should be able to connect to your server and list the tools. And this will also include annotations. And each one of the tools will include annotations. You might kind of get bored ⁓ in ⁓ adding these. It's kind of repetitive. So feel free to just stop when you feel like you get the idea ⁓ and continue on. But annotations on tools is really helpful to make sure that the clients are able to use your tools.

To the greatest effect and give the best user experience possible. So that's why we're doing that in this exercise. Let's get started.

—-------  
53

Harsh Bharadwaaj (00:00:00)  
This step is a little bit different than the others. We've actually updated all of this code to handle the sampling response, parsing things out and creating the new tags automatically and applying them automatically and everything. ⁓ So your job is really just about fine-tuning the ⁓ message that we are sending, the system prompt ⁓ and the information that we're sending to the client ⁓ and to that LLM. So ⁓ it's

l a bit of what I would recommend is that you pull open y whatever assistant that you use and you work on a prompt that seems to ⁓ get you the result you need. The biggest and important thing about this is that the LLM only generates JSON. ⁓ And ⁓ that is sometimes a tricky thing. So this is kind of a prompting ⁓ tips ⁓ example or exercise step. So go ahead and give that a whirl. Your goal is to make it so that it gives you back

⁓ the right structure of content based off of ⁓ the way that we've implemented it in here. When it's all done, if it's successful, then it's going to tell us what tags were added. ⁓ And so ⁓ you will be required to kind of be the LLM ⁓ on the the sampling page. You are kind of acting as the part of the LLM. ⁓ And so when you're all finished, ⁓ I've got some things that you can copy paste so that ⁓ you don't have to ⁓ hand write some JSON

but go ahead and give this a shot. We'll see you when you're finished.

—----------  
44  
Harsh Bharadwaaj (00:00:00)  
Part of the output schema requires that you have a schema ⁓ for the output that you're going to be providing. So we're going to grab a couple of schema definitions that we have in our schemas, ⁓ and we're going to use that for our output schema. ⁓ So starting with this first one, we're going to create an entry. We're expecting to return an entry ⁓ or return that created entry. And so that's what we're going to ⁓ have as part of our output schema. But we're going to make this be an object that has an entry property.

And that is going to be our entry with tags schema, which is exactly what we get from the create entry. It's our entry with the tags it was created with. ⁓ So then we can ⁓ create a whoops const structured content that has that entry in it. And then we'll add right next to our content, we'll add the structured content. So you've got your unstructured stuff and you've got your structured stuff. This has to match the ⁓ type that you have here ⁓ in the output schema. ⁓ And if it doesn't,

you're not gonna get a type error. Hopefully in the future ⁓ they make that give you a type error because that would be nice. Otherwise you're gonna get a runtime error. ⁓ So ⁓ we're gonna pass that structured content and then ⁓ there's a little bit of duplication here now because we were creating an embedded resource before and that has all of the content but now we actually are are providing it as structured content and we also need to for backward compatibility reasons provide it as ⁓ unstructured content as well. So we're doing this

Create text. If you dive into that, that's just gonna JSON stringify. So that's exactly what we're doing with and entry resources as well. JSON stringify. So now it's gonna appear three times. Now I'm not sure how long we're going to need to do this as ⁓ backward compatibility thing. ⁓ Hopefully LLMs and or host applications pick up on the structured content and we can kind of get rid of this. ⁓ but for now the inspector validates this, the spec recommends that you do this, so we're gonna do this ⁓ and have that duplication.

In what we send back. ⁓ If that's a real problem for you, then maybe you can do some testing with the clients that matter to you and make sure that they take structured content into account, and then you can get rid of this backward compatibility. ⁓ That said, we definitely don't have to have both ⁓ embedded ⁓ resources as well as ⁓ the structured and unstructured version of the content. So we're gonna swap the embedded resource out for a resource link.

Harsh Bharadwaaj (00:02:23)  
And let's assume that we've reached that inevitable day where we no longer need the backward compatibility. ⁓ This is how I would do this. You would have your structured content, you would have some pros that explains to the LLM what happened, and then you would have a resource link to say, hey, by the way, this thing that I just created, there's actually a resource for that now, so you can show that to the user if that's useful or interesting for you. ⁓ That's how I would do it. But for now, since we do care about backward compatibility, we're gonna

include the JSON representation of our structured content as well as including it in the structured content. ⁓ All right, that's pretty much it. Now we just apply the same thing over and over and over again. So we're gonna say output schema for get entry, same sort of thing, structured content, and then ⁓ get our resource link and all of that stuff, ⁓ as well as our creating our text. you'll notice I actually don't have any pros ⁓ as a part of this. I could say here is your

journal with the entry, yada yada, but like we didn't actually do anything. We're just retrieving it. The LLM knows what it's doing here, and if it comes back, then it knows it's was successful. So ⁓ I don't always necessarily feel like I need to include some text that explains what's going on. ⁓ But most of the time it's quite useful. It's just occasionally it doesn't seem to make much sense to me. ⁓ Okay, so output schema here, ⁓ we're going to have a list of entries ⁓ and we'll make our structured content. ⁓ And there's that.

⁓ and yeah, we've got our resource links ⁓ taken care of there already. We don't need to change that from embedded because we didn't do embedded anyway. ⁓ All right, structured content again, and there's our resource link. ⁓ Again, like over and over again. This one's interesting because we wanted to ⁓ let it know whether the deletion was successful, because we are sending the old entry, but that doesn't necessarily confirm that it was successful. And so we're being a little bit more explicit here.

in ⁓ our output schema. ⁓ okay, so with that, ⁓ we're gonna create that text ⁓ and ⁓ the resource link. And then again, just keep on ⁓ chugging along. It's like it's so easy, even an LLM can do it ⁓ to take care of that one. And then we go on to this one. ⁓ It's a tag. Here's our structured content, here's our resource link and here's our text for backward compatibility. Here's our list of tags

Harsh Bharadwaaj (00:04:44)  
And our create text for the structured content. Here's our output schema for ⁓ updating a tag, and let's go ahead and ⁓ give that link. Here's deletion, we've got the success boolean, ⁓ there's our structured content, here's our link, and then fine ⁓ we have add to tag, ⁓ and there's our structured content we're gonna include. We've got our resource link. Here we're doing a link for both the entry and the tag, which were both affected, ⁓ and then our text for that. And then finally our ⁓

Video ⁓ and we've got our structured content ⁓ and ⁓ here. Now this one's interesting. So we want to keep the resource link. We're already doing the resource link, even though it's in the structured content, ⁓ the and the structured content has ⁓ or it is the video URI, ⁓ clients may not look for it. So we're gonna keep the resource link here. ⁓ we are going to add ⁓ the structured content ⁓ as a a part of this as well for that backward compatibility.

and this, yeah, this video URI is all that we're gonna do for this one. ⁓ and ⁓ actually yeah, we called it video. There we go. ⁓ And that's gonna be our video URI. That one's kind of a special one. Okay, great. So let's just make sure that all this worked. ⁓ coming over here, we're gonna restart, we're gonna go to our tools, we're gonna get an entry, entry ⁓ to run the tool, ⁓ and ⁓ we've got ⁓ that working. Awesome.

kind of confused here a little bit because it's not giving me the confirmation that it it matches, but we do get tool result of success. Let me refresh. I think I didn't ⁓ relist the tools maybe. So let's try that once more. ⁓ with number one, success. Yeah, there we go. So valid according to output schema and structured content matches a text block ⁓ with other comment content. And then we've got our link right here if you want to subscribe to it or whatever you want to do with it. And there's the content for that too.

Alright, there you go. That is structured content. Can be useful, so good job. ⁓

—-------------  
62

Harsh Bharadwaaj (00:00:00)  
⁓ Changes. ⁓ Alright. ⁓ We are gonna talk about dynamic things. ⁓ On a website, ⁓ you ⁓ interact with the the website with the data, the data is changing. The website is constantly changing. The things you have available to you are going to change based off of the data that is in there. If you have no data, you're gonna have some sort of empty state. If you have some data, you're going to have like delete buttons and different things. ⁓ And I think that the web is dynamic. Dynamic MCPs should be a thing.

as well. ⁓ And that is possible with the spec. We have the list changed event as well as some subscription stuff. So let's talk about these things. ⁓ actually let's go straight to the diagram and then we'll talk about the specifics. ⁓ So when the user starts the application, ⁓ the application is going to initialize the client. That'll send an initialization request to the server. And the server says hey here are my capabilities. And then so now we're we're running and and things are going great. But let's say that the

At some point the tools are called or or something else happens in the database. For some reason, tools, resources, or prompts may change. ⁓ The available tools, the available resources, available prompts. ⁓ And so we want our client to be as up to date as possible. And we want what the user is using, we want the what the LLM sees, all of that, we want that to be as up-to-date as possible. And so we send these notifications from the server. ⁓ The client sends that over to the app, and the app decides what to do with it. ⁓

Typically, the app is going to want to fetch the updated definitions. And so it's going to go do ⁓ a list call on the tools, prompts, and resources. That will send back the updated definitions, and then it can update the UI, update the LLM, whatever. ⁓ And then we also have ⁓ subscriptions that can change. ⁓ And so the application can decide, you know what, I have this resource and I want to make sure that I'm using the latest version of this. So it makes a subscription call to ⁓ the server.

Then the server says, Yep, you are subscribed and it's handling those subscriptions. ⁓ And when resources are changed, then it can proactively send notifications back to the app, and the app can say, great, I need the latest version of that, so let me go read that. And it comes back and it shows the UI with the latest resource, or it show ⁓ sends that to the LLM with the latest version of the resource. ⁓ All of that is important. ⁓ And so there are three different types of changes that can happen here. There is the availability of

Harsh Bharadwaaj (00:02:25)  
tools, resources, and prompts. So whether or not they are available and and what the list of those things are. ⁓ And then for resources, there's a special kind of availability where you have the resource template that represents a bunch of different resources ⁓ and the resource template can have a list of its own. ⁓ And so if that list specifically changes, like you add another ⁓ entry to the database, and now that list would return something new, that is also a list changed event.

And then finally, ⁓ updates to the subscriptions, ⁓ those types of events are what we're going to be covering in this exercise as well. ⁓ So a couple of ⁓ pseudocode and some other helpful tips here. So if we decide, hey, there are ⁓ maybe we're we're managing a doc module tool or something like that. So if there are ⁓ n ⁓ free docking ports, ⁓ more than zero, and it's not already enabled, then we want to enable that tool. You can now dock your module. ⁓ If there are no available docking ports,

And it's not a and it's enabled right now, we want to disable it no more. You cannot dock, ⁓ don't even try because it's gonna fail and it's gonna be bad, and everybody's gonna just things are gonna crash and off, awful pandemonium. ⁓ So ⁓ that is how you handle that in code. And under the hood, ⁓ the SDK is going to trigger a list changed. In this case, if this is a tool, it's gonna change ⁓ list ⁓ send a list changed for the tool. ⁓ and then we have the resource.

So the the thing that is returned by the list callback can change as a result of the underlying database. ⁓ And ⁓ this can also change as a result of anything else. And so these are two different types of list changed for the resource. This is whether that type of resource is even available, ⁓ and this is ⁓ elements of that resource being available or implementations of that resource. ⁓ And then ⁓ finally we have resource subscriptions. So the client says, hey, you know what? I want to know.

When this space station changes. Okay, great. We'll keep track of that. And if it ever changes, we'll let you know. ⁓ so that's what this response is: is like, okay, great, we've got your ⁓ request and we'll we'll let you know when that changes. ⁓ And then later on, the server says, hey, guess what? That changed. Here's the URI, here's the title of the thing that changed. ⁓ just so you know. And then the client can go and make a request to get what what is the latest now. Now that I know that it changed, well, what is the latest now?

Harsh Bharadwaaj (00:04:49)  
So that is what you're gonna be implementing in this exercise, maybe not with space stations and docking ports and stuff, but ⁓ with our journaling app. I think you're gonna enjoy this one. So let's get to making our MCP server a little bit more dynamic.

—----------------  
45

Harsh Bharadwaaj (00:00:00)  
Why does a chicken coop only have two doors? Because if it had four doors, it would be a chicken sedan. Ha ha. ⁓ Alright, great. So now it's break time. I hope you had a really good time with that exercise. It's time for you to write down what you learned ⁓ and ⁓ get your body moving. Get the blood flowing, get a drink of water or whatever you need. Get go grab a snack. Go give somebody a high five and tell them something nice about themselves. That'll make you both feel better. ⁓ and then you can come back for the next one.

—------------

66  
Harsh Bharadwaaj (00:00:00)  
So we actually have two resource templates that have that implement the list callback. So it's that one right here with the tags. And then we also do this for the video resource. We implement the list callback. So this is applicable to any resource template that implements the list callback. We do not need to do this for the entries because we do not implement the list callback. ⁓ So for both of these, we have this mechanism for subscribing to video changes, ⁓ and we already have the database for subscribing.

To those changes. So ⁓ when the database changes, we're going to ⁓ trigger a resource list changed. And when ⁓ videos change, then we'll ⁓ send that as well. ⁓ And so that way we can always be made aware of when those changes happen and clients can update automatically as that happens. ⁓ So let's ⁓ just test this out again. We'll delete the SQLite database just to be doubly, doubly sure. We're gonna restart. We're going to actually let's refresh, clear all the state of everything.

We'll list resources, list templates. We do have the templates. ⁓ could argue that maybe those could be disabled also. ⁓ but we'll go over here to our tools, we'll create a tag. Here's the tag name, here's the description, we'll run the tool, and boom, we've got those ⁓ resources now available. ⁓ And I come over here again, we're going to ⁓ do another one and run the tool, and we get another list change event happening ⁓ right there. So that is happening.

So clients can now say, ⁓ list change event on resources. Let me go and ⁓ list the resources again. ⁓ So hopefully that helps you understand there's there's the distinction between the resource itself is available and an instance of that resource is now available or it's changed. Maybe we can have the same sort of thing if we get a deletion. So if I go over here and here, let's ⁓ refresh our tools now. ⁓ and I say delete tag, we'll delete tag one, run the tool. ⁓

And here we've got elicitation. Woo-hoo\! And we s submit that. Now we're also getting another ⁓ list changed event for that as well. So hopefully ⁓ that all is checking out for you and tracks. Thank you for having a good time with this one. We'll see you in the next.

—--------  
39

Harsh Bharadwaaj (00:00:00)  
Alright, in this workshop, we are going to learn about the advanced MCP features that are available to us to make our MCP servers ⁓ really interactive, really dynamic, really awesome. And so we're going to be learning about how to enhance our tools so that clients know a little bit more and LLMs know a little bit more about what our tools do and what they offer to the user and give the user a much better experience with that.

We're gonna learn about elicitation and how you can ask for additional information from the user even as they're going through some tool call or something like that. ⁓ And then we're going to borrow the user's LLM to do some really awesome things for them ⁓ using sampling. That is actually a really cool feature that I think ⁓ is one of the things that I love about MCP the very most.

And then we also need to handle long-running tasks that our users are running. Sometimes we need to give progress updates and we need to handle cancellation and that sort of thing. ⁓ And then we want to make sure that the ⁓ resources and tools and prompts that are available to our users ⁓ are kept up to date and as dynamic as you would expect a website to be. As the data changes, the website changes. And same sort of thing for MCP servers. So we're gonna learn all about all of that stuff to make your servers the very best that they can be.

Let's jump right into it. Thank you for joining me today.

—--------