Day 1-2 MCP Fundamentals   1-38

Harsh Bharadwaaj (00:00:00)  
Sometimes the user wants to be able to select a very specific resource. You might be able to say, hey, can you edit ⁓ index.ts? ⁓ But maybe your project has like 7,000 index.ts files, right? Like depending on how you're organizing your files and stuff. ⁓ So you want to be able to specify exactly which one you're talking about. ⁓ just as an example, here if we were to open up a chat ⁓ and I say, Hey, I want to edit ⁓ resource.ts, I can drag this file.

Right into the context here, and now the AI knows exactly what resource I'm talking about, which file I want to talk about. I specifically want to talk about resources, or I want to specifically talk about even the database directory. And this is what we're going to talk about. It just kind of helps the LLM know exactly what you're talking about. It helps the user to ⁓ LLM communication layer there. ⁓ and so the ⁓ in this way the user is being more specific.

About the types of things that it wants to bring into context. ⁓ And that is what we call resources. The user is able to select a specific resource that they want to have included. ⁓ So resources in MCP are designed to be application driven. ⁓ So you might notice also if I open this up, it says one tab is included. ⁓ That tab is the currently active tab. So the application has decided: okay, I'm going to include this resources resource automatically.

Because it makes sense in the context of this conversation that I include that. So that is a resource that is application driven. The application decides to include it. Now, for more general ⁓ AI agents, ⁓ it doesn't always make sense for the application to ⁓ specifically include a resource automatically. And so typically, ⁓ and even ⁓ in specific agents like this one, ⁓ you're going to have some sort of mechanism for adding ⁓ resources manually as well. ⁓

I can say ⁓ I want to include Git or I want to include some docs or I want to include ⁓ rules or past chats or whatever. So that all of these are kind of examples of resources. Now, whether Cursor implements these as resources under the hood, ⁓ I think probably not because cursor was ⁓ created and started ⁓ development before MCP was a thing. ⁓ But we have the same sort of idea with resources, it's the same concept. And in VS Code for sure, you ⁓ can

Harsh Bharadwaaj (00:02:22)  
add your own resources and things like that, ⁓ as well as in Claude. And we're actually gonna take a look at that, what that actually looks like in Claude as well. ⁓ So ⁓ they're application driven and typically also expose some mechanism for the user to also add ⁓ resources specifically. So what does this actually mean? What what does the workflow typically look like from a workflow perspective? ⁓ Well the ⁓

The user is going to request to view a resource or include it in the context or something. The application is going to ⁓ invoke a list resources ⁓ RPC ⁓ request. The client will forward that over to the server. The server will send all of the resources that are available, and the client will send that to the app. And then the user ⁓ can view those and select one. Or maybe the app is doing all of this and it's deciding to include one automatically. ⁓ Then the app says, hey, I want to read that resource because I'm going to include it in the context.

The ⁓ server receives that request and sends back the resource. ⁓ the resource context makes its way back to the app and maybe it's displayed to the user ⁓ or the ⁓ or it's just hidden kind of in the case of this, like I d I don't know what that is exactly. I know what active tab I have, but ⁓ anyway, maybe displayed to the user kind of depends on the situation. ⁓ But when the user submits a prompt, typically this resource is going to be included in

that prompt to the LLM. And so then the user can ask, hey, I want you to summarize this blog post, or I want you to ⁓ edit this file, or I want you, you know, ⁓ tell me the ⁓ the ⁓ customer that has the most re you know, highest paying or whatever. Right? So based off of the content context that it gets from that resource. ⁓ So that is what ⁓ the the role of a resource is in MCP. As far as the nitty-gritty details of what's going on under the hood

A resource ⁓ is just another JSON RPC request where ⁓ the resource can be read. The params include the URI for the resource. Now here I have taco colon slash slash menu items carnet asada. ⁓ And the response would be ⁓ a matching ID to associate the request response. ⁓ And then the result has the contents array. ⁓ And the content item has a URI, so matching the URI that was being requested.

Harsh Bharadwaaj (00:04:44)  
The MIME type and then text. And there are various other ⁓ other types that you can provide in here. ⁓ you can include blobs and stuff like that as well. ⁓ but ⁓ it all is kind of gonna depend on what clients you're targeting and what sorts of mime types they support. Most clients are going to ⁓ just hand off your ⁓ your resource to the LLM and so doing application JSON or even markdown or something like that ⁓ would probably make

pretty good sense for those situations. But yeah, then we just have the text ⁓ of ⁓ of that resource, what represent that ⁓ is ⁓ representing. ⁓ So one interesting thing here, you may never have seen ⁓ a ⁓ mime type ⁓ of ⁓ taco, and that's because ⁓ that that was made up. ⁓ the scheme actually doesn't really matter. ⁓ if if you're unfamiliar with schemes, you can take a look at the Wikipedia on the uniform resource identifier.

But that first part of the URL or the URI ⁓ is called the scheme. ⁓ And so we've got a bunch that you might be familiar with, HTTPS, there's LDAP, Mail2, News Tel, Telnet, URN, ⁓ there's FTP, you know, even Git. ⁓ All of these are different schemes. ⁓ And ⁓ in MCP land, the scheme ⁓ just needs to be kind of consistent unless you're integrating with a client that ⁓ understands a very specific scheme.

Then it really actually doesn't matter all that much what the scheme is or even what the URL is, as long as it's ⁓ consistent ⁓ and the client can make a request to that ⁓ with that scheme and and get a resource back, then it's fine. So you can kind of see the entire URI as kind of an identifier for the ⁓ for the resource, which is exactly what it is. It's a uniform resource identifier. ⁓ but what's interesting or or what I would recommend is if you are going to use HTTP or HTTPS.

That ⁓ it be a proper ⁓ actual URL or or ⁓ able to be consumed by other things that can consume that. So if you're gonna use HTTP, ⁓ I should be able to navigate to that to a br with a browser and get the same result. ⁓ so ⁓ typically I'm gonna avoid using HTTP unless I'm identifying a resource that ⁓ you can retrieve through HTTP as well. ⁓ in the future, I think that having some standard resources.

Harsh Bharadwaaj (00:07:09)  
it was really interesting and cool. So like having an Amazon, maybe I'm ⁓ a a store MCP and I have some sort of affiliate link and or something like that. And so now this MCP server can talk to this one and like maybe Amazon's got one or something, and so we can communicate across MCP servers in that way. That might be pretty interesting. ⁓ But yeah, for most intents and purposes, I pretty much just use the name of my service or the name of my server as the scheme, and then ⁓ that works just fine.

So that's a note on the scheme. Hopefully ⁓ this is all ⁓ very clear to you and what the purpose of resources serves. Again, this is not something that the LLM is going to decide it needs to go grab. It's something that the application gives to the LLM or the user through the application gives to the LLM to enhance the context to help it ⁓ know what it is that we're talking about, to be a little bit more specific. ⁓ Another thing that I'll mention too.

Is that ⁓ sometimes you can have resources, but then you can also think, you know, it would be really nice if the LLM could retrieve this, right? And so you can make a tool that exposes that resource as well. We'll talk about that ⁓ in a future exercise. So for now, we're just going to define those resources. I hope you have a really good time with this exercise.

—---------------

Harsh Bharadwaaj (00:00:00)  
Hello, my name is Kent C. Dodds, and I want to welcome you to the Model Context Protocol Fundamentals workshop where you are going to get your first introduction to building Model Context Protocol servers so that agents can communicate with your services and whatever else you want to expose to those agents. ⁓ I'm really excited about this because this is a first step for you ⁓ in a journey of becoming a Model Context Protocol Server developer, where you can connect your services with the agents that your users.

Really want to use. This is going to be really awesome because it brings a lot of natural language capabilities to your application and services ⁓ and automatic integrations with other MCP servers. It's really cool, really awesome. I want to show you a little bit ⁓ about the architecture of MCP before we get started, and then we can get into some of the hands on the keyboard exercises for this workshop. ⁓ Starting with this, this is basic architecture for MCP, like where the server sits. ⁓ You have a

The host application that manages the interaction with the LLM. So Cloud IDEs like cursor or VS Code and other tools like that. ⁓ And in fact, there are some models ⁓ and APIs that have built-in MCP support. So you just supply the MCPs, you make the API call and it will tell you what tools it needs to call and stuff like that. ⁓ So lots of support for MCP in a lot of different areas. ⁓ So the host application is kind of responsible for the client side of things.

And then via the MCP protocol specification, it communicates to the MCP server. Now we have three different ways you can do servers. ⁓ and in fact, there are probably many more than this, but here are the three typical ones. You have desktop applications that are running on the computer, ⁓ those are able to run local MCP servers. And so this would be like the for example, the Epic workshop server that I've built.

runs locally on your machine right alongside the workshop app that you're running on your machine. That way it has access to your work in progress and so it can see the files that you're working on and compare those and ⁓ and it has access to the transcripts that you have access to, all of that stuff ⁓ and it can communicate with the local data source. ⁓ In fact, the Epic Workshop MCP server also communicates with remote data sources as well. So it's kind of a combination of A and B. ⁓ But then

Harsh Bharadwaaj (00:02:18)  
The vast majority of MCP servers that I think are we're going to see in the future ⁓ are those that are hosted on the internet. ⁓ And so here we have MCP Server 3, or C, ⁓ which is on the internet and it's communicating with its own data sources. ⁓ And so, like examples of this would be the Sentry MCP server, ⁓ the linear MCP server, GitHub, ⁓ lots of those are developer tools, but there's also like

E-commerce ones with Shopify and and Stripe as well, also developer focused. ⁓ but ⁓ yeah, pretty much ⁓ most of the websites that you see in the world could probably benefit with a slash ⁓ MCP endpoint. ⁓ And so the vast majority of MCP servers will be hosted on the internet. ⁓ For us, we're actually gonna be developing a server that's a little bit more like this, ⁓ like server A, because ⁓ it's a little bit easier for us to develop as we're working through it.

And ⁓ the actual contents of the server don't really change a whole lot ⁓ based on which transport you're using, whether you're going over HTTP or if you're going through standard I.O. And we'll learn all about that as we go through the first exercise of this workshop. ⁓ So that's the basic architecture of how these things all fit together. ⁓ The other thing that I really need you to understand is the MCP inspector. ⁓ There are a lot of ways that you can test out your MCP ⁓ as you're developing it.

There are hosts you could wire it up with cursor or VS Code and work right in your editor using the MCP server. You just have to make sure you're ⁓ refreshing it and all of that as you're making changes. ⁓ you could use Claude Desktop even, you could use Goose. ⁓ I like using the MCP inspector. I feel like it's a little bit closer, it's the official ⁓ package from the model context protocol team. And so there's just a lot of lot of things to love about it.

And so the MCP inspector, ⁓ as you're working through this with the workshop app, is going to be pre-configured for you all the time. So as long as you're ⁓ running through the workshop app, ⁓ you shouldn't have to worry about configuring any of this stuff over here on the inspector side. ⁓ But just so you know, ⁓ you have different transport types. Ours is going to be standard IO, but you can also do do the streamable HTTP or SSE. ⁓ We have the command here, ⁓ we're running npm, and then we're giving the arguments.

Harsh Bharadwaaj (00:04:38)  
⁓ because we're running an NPM script. The script is called devmcp. ⁓ we are using the silent flag to keep npm from outputting a bunch of logs that might confuse ⁓ our ⁓ client. ⁓ And we're also using the prefix flag ⁓ to point to where this ⁓ where the code for our project is. And so ⁓ again this will all be pre-configured for you, so you shouldn't have to change any of that stuff.

But it's useful to know, especially if you want to wire things up in another client, you can just copy this configuration into your other client and it should work without issue. ⁓ So with all of that set up, you can hit connect, ⁓ and now we're connected to the server. Now, the way that standard IO works is it just spawned this process with the command and arguments. And so anytime you need to change your server, you're gonna need to restart. And in some cases, you may even need to relist resources and stuff like that. So

Often I actually find it's helpful to just hit the refresh button and then hit connect again. That way everything is fresh and and you don't have like ⁓ state that's left over from your previous run. ⁓ In any case, now here we are. We've got resources, prompts, tools, and ping. Our first exercise, we will only have ping. ⁓ and then we also have sampling. That's a an advanced feature we'll talk about in another workshop. And elicitations, same thing. Roots is a pretty niche feature. ⁓ We're not going to be talking about it. It's mostly useful for

people who are building like ⁓ coding ⁓ assistant kind of MCP things that helps you know where your code editor is. ⁓ it's not very useful for generic tools, but it definitely can be useful ⁓ in some scenarios. And it's not ⁓ terribly difficult to use. So if you do need to to use this, then ⁓ you can take a look at the docs. But most servers aren't going to need to use roots. And then auth, of course, we're gonna talk about in the auth workshop.

you're not going to be working with that in this one. So we're just working with ping tools, prompts, and resources. ⁓ Here we ⁓ I'm looking at the finished version of all this. You're not gonna see this at the very start. ⁓ But we've got our our lit resources here. We've got our ⁓ templates where we can supply different ⁓ IDs and then read those resources. ⁓ We've got a prompt here where we can supply an entry ID and then get a prompt that we can give to the LLM. That's pretty cool. You can take a look at that ⁓ when we get there.

Harsh Bharadwaaj (00:07:00)  
and then we've got our tools for generating ⁓ journal entries. And so that's what our MCP server is all about is generating entries or or managing ⁓ a journal for our users. So ⁓ we're you can look forward to that. ⁓ if you do want to wire this up, I'll have a video that you can ⁓ watch to show you how to wire this up with ⁓ either your editor or with Claude. That's what I've got in here. ⁓ I ⁓ in this prompt I actually used one of the prompts that are exposed from our server.

and the LLM generated stuff and then we had it ⁓ actually use some of our tools here too. So it's pretty fun. You'll you'll have the chance to work with this. ⁓ Everything that we do in the workshop, we're focused on the inspector, but if you do want to wire it up with a another ⁓ tool or something, be my guest. it does make things a little bit more interesting when you can actually work with an LLM. ⁓ So that is all that I have for you and getting you all hyped up and really excited about this workshop.

I'm so excited that you're here with me. ⁓ Let's get right into it with the first exercise.

—---------------

Harsh Bharadwaaj (00:00:00)  
Alright, so let's get started by implementing the entry. We're gonna say server, ⁓ it's agent.server.register resource. ⁓ I'm gonna close the parentheses for it. That might help it ⁓ autocomplete some of this stuff for us. ⁓ And so we've got our entries. ⁓ our resource template is epic me colon slash slash entries with the ID being the param. We're gonna set the list to undefined. We don't want to be able to list all of the entries. We'll we'll deal with list a little bit later.

but the SDK does want you to specify it. You can't just leave this off. That's a type error. ⁓ I kind of disagree with them. I I think that's pretty annoying to have to do that. But ⁓ we'll talk about list in a ⁓ future step. ⁓ And then we'll have our callback ⁓ here after defining kind of the config here. So we've got our user-facing title, we have our description, and then we have our ⁓ callback here that accepts our URI, we have our entries, and then ⁓ we're gonna go and get

⁓ or or ⁓ return the contents here, which is gonna be our entries. ⁓ now we don't want to get all of the entries, so our AI didn't quite get this right. ⁓ we want to instead get a specific entry. And so w we want to take that ID ⁓ and use that as ⁓ part of ⁓ our ⁓ the ⁓ resource that we're retrieving. So we're retrieving a specific entry. ⁓ The interesting thing here

Is that all of these are interpreted as strings. And so you'll notice ID is ⁓ typed as a string or an array of strings, or it could just not be defined. I'm a little bit confused by this, to be honest, because I can't I can't imagine a situation where the ID could be an array of strings. Like perhaps it can ⁓ clients have the ability to request ⁓ any number of resources that fill a resource template. I suppose that would make sense. ⁓ but I haven't seen any clients that do that.

And for that reason, I'm just gonna say, hey, if you're not sending me a string, I'm just not going to work at all. ⁓ So we're going to actually add an invariant right here that says, hey, the ID has to be a type of string. If it's not, then you're doing it wrong. ⁓ The IDs in our application are numbers, so we're going to convert that number or that ID into a number. ⁓ We could also potentially say, ⁓ if you wanted to, you could say ID number.

Harsh Bharadwaaj (00:02:23)  
And ⁓ make sure that it is a valid number. So you could take it further that way if you want to. ⁓ and then this is gonna be our entry. ⁓ And we'll have entry right there. Okay, ⁓ great. So that takes care of the entry. So let's make sure that we actually ⁓ that this actually does work. ⁓ and actually to take it ⁓ I I think it makes more sense to just call this entry. ⁓ and we'll call this entry.

And this isn't all the entries in the database. That was another thing that AI got wrong. ⁓ This is why reviewing your AI generated code is so important. ⁓ But ⁓ an entry in the database ⁓ is gonna be sufficient for that. Okay. ⁓ So we'll connect, we'll go list templates, and there's our entry. We can type in ⁓ one of those numbers to pre-fill. ID must be a number. ⁓ okay, so we messed up something. ⁓ not is NAND. There we go. ⁓ So

Now we'll list our templates, entry, entry two, ⁓ read the resource, and there it is. Ta-da. ⁓ So the tag's basically the same thing. ⁓ hopefully our AI can kind of figure this one out a little bit better than the last one. ⁓ but yeah, we've got our tag, we've got our ID in the ⁓ parameters, we have our list is undefined, we'll deal with that later. We've got this title and description, a tag in the database. We're ⁓ verifying that the tag that we were given is a number.

And we're returning ⁓ that tag as JSON. ⁓ And so we can come back here, we can restart, we can ⁓ relist our templates, and there we've got our tag and we've got tag number two. ⁓ So that is resource templates. You can have more than one parameter. You can have ⁓ like ⁓ name, yeah, sure. And ⁓ with that now you get ⁓ name and you can ⁓ go forth and ⁓ use that name ⁓ however you want.

there are more capabilities here that we're going to look at here in a second, like the list. We also have ⁓ auto completion stuff going on as well. So that is your resource templates.

—-----------------

Harsh Bharadwaaj (00:00:00)  
So there's just a couple things we need to do. First, we're gonna grab a couple imports from the Model Context Protocol SDK, and then we need to create the MCP server. So we're going to make our server ⁓ is a new MCP server, ⁓ and your AI assistant may or may not have gotten this quite right. Here it wants to stick the instructions as part of the configuration in here. ⁓ The AI SDK or the MCP SDK does not give you a warning about this, but this is actually the wrong place to put the instructions.

In the future, hopefully it will give us a warning about that. ⁓ But we do need to put the instructions as the second argument here. And if you were to dive into the MCP server ⁓ as it is here, you'll see the constructor constructor is accepting the server info implementation and server options. If you look at implementation, then that's going to come from the Zod schema for their implementation. ⁓ That has a name and a title as well as a version. ⁓ But the reason the reason that it doesn't yell at you because of the instructions.

is in part because of this pass-through. So what this does is it takes any of the other arguments or options or properties that you ⁓ pass and it passes those through. So if you were to actually leave the instructions right in here, then it would actually go to the client and you might even see it ⁓ in the server info ⁓ in the logs. But ⁓ clients aren't going to be expecting that. They expect it as part of the server options.

And if you look at server options, we got capabilities and instructions in there. That's where the instructions are supposed to go. ⁓ And we'll talk about capabilities later. ⁓ So with that now, we're all set to hook this up to the standard I.O. transport. ⁓ So we're gonna create a new instance of that. We'll await server connect on the transport. And then we can console.error server is running. Now, ⁓ well, let's also remove the error not implemented. ⁓ And here we can here we can do ⁓ the epic ⁓ me.

server running ⁓ on s ⁓ standard IL. There you go. I think that's what I wrote in the final solution. ⁓ We can talk a little bit more about the instructions and different things. ⁓ later your instructions are really just the it's basically talking to the LLM and saying, hey, this is ⁓ what my server can do. This is what you can expect from me. ⁓ our server can't actually do anything ⁓ just yet and so we would be lying if it said ⁓ if we said that it can do anything. But in the future we're going to be able to

Harsh Bharadwaaj (00:02:26)  
To solve math problems, so we'll stick that in there for right now. You'll want to keep that those instructions up to date so that the LLM knows when to ⁓ use your server's tools to accomplish different tasks. ⁓ And so let's just make sure that this is all working. If I hit connect, then we're gonna get ⁓ initialize right here, and we get that initialization. We'll see that console.error shows up in here as well. ⁓ If I hit ping server, then we're also going to get that ping request.

⁓ And so we are all set to go on that. ⁓ The one thing that I'll mention ⁓ is we're using console.error right here. ⁓ And that might be a little bit confusing. Like why do we need to use console.error? ⁓ And the reason is if we were to use console.log, then it could confuse clients. So here, let's refresh this, run this again. And ⁓ this client is ⁓ just ignoring messages that it can't do anything with.

there may actually I haven't even looked, but there may be an error ⁓ in the console. If I ⁓ restart, yeah, it's not showing us any errors. There could actually be an error in here. Yeah, there it is. ⁓ syntax error, expect a token E, Epic Me Sir is not valid JSON. And so it's handling that gracefully, but we don't wanna send errors to our clients. ⁓ Why is this happening? Well, the way that this actually works ⁓ when you're using the standard IO transport is it's kind of

the the client is basically doing something like this. ⁓ it's grabbing child process ⁓ and it's doing a spawn or or something like this ⁓ from child process. And then it takes your ⁓ command that has been configured in here, which the ⁓ Epic Workshop app configures for you. But we have a command and then we have our arguments. ⁓ silent is just to to silence npm because we're running with the npm command.

The prefix is saying what directory we should be running this in. And so that's pointing to your playground directory and then run and then devmcp. ⁓ And you can take a look at the package.json. This dev mcp is running tsx with source index.ts. So we're just running that file directly. And so here our command is npm. ⁓ R uhs ⁓ is an array of these things. ⁓ So here, ⁓ nope.

Harsh Bharadwaaj (00:04:52)  
I was hoping that the AI would be able to figure out I want ⁓ that. ⁓ And one more thing there. ⁓ and ⁓ haha, that's funny. Yeah, get rid of that. There we go. ⁓ and then ⁓ it's going to spawn ⁓ our ⁓ command and our arguments as a child process. And then it says ⁓ child ⁓ on ⁓ data. So as the child is sending data through basically console logs. So they're

process standard out write, then it's going to receive that and it's going to parse that out ⁓ with a JSON parser. And that's where we're getting that error. You know, it can't parse as JSON. And then ⁓ it actually ignores standard error. Or ⁓ in actually in our case, our client is processing standard error as ⁓ notifications, ⁓ which I think is pretty cool. I think that is the SDK that's handling that for us.

⁓ And then it when it needs to send a request, it's going to say child.standardin ⁓ and it's going to write ⁓ to standardin with the the messages that it needs to send. So that's kind of the request response ⁓ API between ⁓ the client and the server, and then back again, ⁓ is through standard I/O. And so that is why you want to avoid using console.log, you'll use console.error instead if you're using the standard IO transport.

Other than that, actually ⁓ most of the ⁓ the stuff that you're writing in your server is going to be the same, whether you're using standard I/O or ⁓ if you're using the ⁓ streamable HTTP ⁓ API or transport ⁓ for this. So the transport actually doesn't affect the way that you write your server your ⁓ MCP server a whole lot. ⁓ what affects it more is your deployment target. So if you are running locally on the user's machine.

You have access to different files and s and stuff on their machine. ⁓ if you're running in a deployed environment, especially in a like a serverless environment, you don't want to rely on the file system. You're gonna want to connect to a database. If you're running locally on the user's machine, you probably don't want to have any of your private keys for accessing third-party APIs ⁓ as environment variables or anything. If you're running in a deployed environment, that's perfectly safe. You can set environment variables and all of that stuff. So it's really a

Harsh Bharadwaaj (00:07:13)  
Very similar type of model to ⁓ deploying software in general. ⁓ But ⁓ as far as the rest of the code, it doesn't actually change a whole lot. This is really the integration layer. And then as your transport changes, ⁓ things are mostly going to stay the same. ⁓ So there you go. That is getting our ⁓ MCP server initialized ⁓ and ⁓ connected to a transport.

And once we've got that going, ⁓ the SDK handles the ping request for us automatically, it handles the initialization for us, ⁓ and we're ready to start defining some tools.

—-------------

Harsh Bharadwaaj (00:00:00)  
Hey, so we've got embedded resources, but what about resources that are really big and we don't necessarily want to include the entire resource ⁓ in ⁓ side of the content? For example, we have this list entries right here, and we're listing all of the ⁓ contents of all the entries. That's probably a little bit much to include in here. Honestly, if ⁓ I should probably redo this so that it only has the title ⁓ and ⁓ maybe a a ⁓ short description or or like maybe just ID and title would be sufficient for this.

⁓ but yeah, if we had ⁓ you know 100 resources that were returning from here, 100 entries, that would probably be too much. And so instead what would be better is if the list entries returned ⁓ a list of resources and that ⁓ then or like maybe not ⁓ the embedded resources, but like a link of or a reference to the resource. So basically, ⁓ like the embedded resource, except without all the content.

And that is what a resource link is all about. And so then a client could get that back and be like, great, we've got a whole bunch of other resources that the user can ⁓ add, or I can just automatically add to the context and then continue generation. ⁓ It just gives them a little bit more options of what they can do with that. So that's what resource links is for. You're going to be implementing this, and when you're all finished, this is what it should do. ⁓ We'll go list entries, run that. ⁓ And ⁓ in particular, the MCP client.

that or the NCP inspector that we're using ⁓ will ⁓ render that ⁓ in such a way that you can just click on this and then it will do a resource read, which I think is actually like exactly the way that I would expect this to be implemented in most, where it sees that it has the resources and makes it really easy to read the specific ones that you're interested in. ⁓ So if you look at the tool call for ⁓ our list entries, then here is what a resource link looks like. It has a type, it has MIME type

this is like what the resource will be if you request it. ⁓ the description, ⁓ URI, and the name. ⁓ And there you have it. So your job is to implement this in this exercise. Go forth and have a good time.

—-----------------  
Harsh Bharadwaaj (00:00:00)  
Not everybody wants to be a prompt engineer. Instead, they want you to write the prompt for them, and that's what you can do with MCP prompts. ⁓ So let's say that there's a common use case that our users have where they've written this journal entry and they want to have tag suggestions for what tags should be applied. They've got a bunch of tags in the database already. Should they apply some of those or should they create brand new tags altogether? And LLMs are pretty good at reading a journal entry and making suggestions.

So let's add from EpicMe suggest tags. And we're gonna suggest tags for entry number two. And I'm gonna also add an additional prompt. So like I'm using the prompt, but I'm also going to say ⁓ make them ⁓ silly. And so a kind of combination of the user specifying something and the prompt that I wrote. So here are some silly tags to create. ⁓ None of your current tags are silly, but family and nature are ⁓ already on there. ⁓ okay, so let's say

Cloud watching. ⁓ And I'll tell it to add cloud watching. Now, how does it know how to add cloud watching? Because I told it. And so it's going to say, okay, let me create that tag and then I'll add that tag to the entry ⁓ using the tools that were told to me by the prompt. And now it's added those things for me. So ⁓ the users don't have to be the prompt engineer. ⁓ You can write prompts that are that handle common use cases that your users are going to have.

And they can be dynamic ⁓ based on the user's data, based on the user's input, and they can even enhance those prompts with additional ⁓ input of their own. ⁓ So that's what you're gonna be doing in this exercise is adding prompt support to ⁓ our MCP server. Now let's talk about a little bit about how all of this works. Again, prompts are designed to be user-controlled. The application isn't gonna automatically add a prompt. The LLM doesn't know that prompts exist, the user selects the prompt that they want.

The way that this ⁓ all works from a code perspective is you register the prompt, you give it a ⁓ name, basically an identifier for the server, ⁓ a title, a description, some arguments ⁓ that are also described ⁓ through Zod, and then you return an array of messages. So an object that has an array of messages. We'll talk a little bit about the specifics of these messages ⁓ during the exercise, but each one of these messages is going to have a content that's pretty familiar to what we've already done.

Harsh Bharadwaaj (00:02:25)  
⁓ like tool call responses. ⁓ As far as the architecture for how this all flows, the user opens a prompt menu. The app is going to request available prompts with the prompt list ⁓ method. ⁓ The client will send that over to the server. The server will respond with the list of prompts. The client will send that to the app. The app will show that to the user. ⁓ And then the user will select the prompt that they want, ⁓ we'll get the prompt template that'll come back. The prompt template will be shown to the app. The

⁓ app will then ⁓ ask the user to fill in the prompt. These arrows are pointing the wrong way. This should be asking the user for the prompt, not the LLM. Although ⁓ the only part of this that's actually specified is over here. So presumably you could have an application that does talk to the LLM. So I think I'm not going to re-record this video. I think I'm going to leave that in there. You could do it however you want. That's that's the application's decision. ⁓ But the demo I just showed you with Claude, it actually asks me, the user, to fill in the prompt.

And then ⁓ that prompt is ⁓ then ⁓ we go back to the server and get the results there. We send that over to the LLM, and then the LLM generates whatever it's going to generate. ⁓ So that is the basic flow of a prompt, ⁓ and that is the experience of a prompt. ⁓ And this is the code of a prompt. And now you're going to write a prompt. Go have fun doing that.

—----------------

Harsh Bharadwaaj (00:00:00)  
Let's start out by adding our capabilities ⁓ and that will include ⁓ an empty object of tools. And then let's add our first tool, server.register tool add. ⁓ And we want to give it a title, ⁓ which is a user-facing ⁓ description of what the tool is or title for the tool. ⁓ And we're going to have an actual description. This is ⁓ possibly a user-facing, depending on the application.

⁓ but also an LLM facing ⁓ description. So this is how you tell the LLM ⁓ if you're going to call this tool, or or these are the situations in which you might want to call this tool. This is what this tool can do. So here we can just say add two numbers. I think maybe ⁓ add the ⁓ numbers one and two, because our tool doesn't do a whole lot other than that. ⁓ and then we're going to return the ⁓ spec defined ⁓ response, and that's going to be an object with a content.

⁓ array of different items of content. ⁓ And so for us, we only have one item of content. It is type of text. And the text is the sum of one and two is three. There are other types that you can have in here. You can ⁓ dive into this. We've got images, ⁓ we also have audio ⁓ and ⁓ and then we've got resources and different things like that we'll explore in the future.

We don't really get into ⁓ images and audio. It is kind of fun to base 64 encode your image and send it across the wire. And for a client that supports it, which the MCP inspector does, you can actually see the image. It gets rendered and it's kind of fun. ⁓ I don't really see there being a whole lot of use for that ⁓ once MCP UI becomes more widely adopted. If you have the ability to show an entire UI, why would you just send the image? Just send a cool UI that includes the image, maybe even get a carousel or something. I don't know.

⁓ but ⁓ yeah so we don't really get into images and audio. You can like even display video in MCP UI as well, and you can't do that ⁓ through MCP here. ⁓ So there's probably some use cases there, but we're gonna stick pretty much to text and resources ⁓ for ⁓ what I do and what I'm gonna be teaching you. ⁓ So we've got type text, the sum of one and two is three, ⁓ and that gets our tool all set. ⁓ the important thing here is to think about okay, who is reading this.

Harsh Bharadwaaj (00:02:21)  
And ⁓ the LLM is going to be reading add. The user is ⁓ and and potentially the user will see this as well, depends on the application. ⁓ the user will be seeing the title, most likely, if the application is ⁓ showing them that, which they should, that is intended to be displayed to the user. ⁓ the description also intended for the LLM. When I'm writing ⁓ a longer description, for some tools they can actually get quite long. Like this is a pretty complicated tool, what it can do.

you do want to keep it terse. You you don't want to blow up the user's context window or anything, ⁓ but ⁓ which is the the number of tokens that the LLM is showing to the ⁓ the ⁓ or the host is showing to the LLM. ⁓ But ⁓ yeah, we don't want to blow up their context window, but you do want to be descriptive of what your tool can do. Include examples and ⁓ of like if you give me these parameters, then this is what's going to happen, stuff like that. ⁓ But ⁓ yeah, we're going to keep it nice and short because it's not

Really doing a whole lot for us here. ⁓ But yes, this is for the LLM. The user might also see this as well. And then here we have our ⁓ callback for whatever we can do. And honestly, this is doesn't necessarily have to be async, but ⁓ in our case it doesn't need to be. But typically it's going to be because here's where you're going to say, ⁓ wait, ⁓ get thing. ⁓ maybe you're making a fetch request or you're talking to a database or whatever. So you can do anything in here that your service ⁓ server has access to, which is pretty cool.

And then you return your result based on that. ⁓ so you could perform mutations, you can go get data, whatever. And then you're augmenting ⁓ the generation of what the LLM is gonna generate. That's kind of ⁓ like MCP and RAG have a little bit of overlap here because you're literally the LLM is retrieving some stuff from your MCP server. It's getting a response back, and that's augmenting its generation. So it basically is rag, it depends on how you define it. ⁓ so let's just take a look at ⁓

How this all works. Let's list our tools. There we go. We've got ⁓ our tools listed automatically for us. ⁓ Initialize is going to include our capabilities. ⁓ You'll notice also, actually, I get this question a lot too. ⁓ List changed true. We didn't specify that. That was the SDK just decided. ⁓ Because the SDK itself supports ⁓ list change events. ⁓ You can actually save this. We have add. Here, I'm giving you a preview. We're going to do this later, but disable.

Harsh Bharadwaaj (00:04:44)  
And so because the SDK supports the ability to disable and enable tools, it's going to automatically add the list changed for us ⁓ because the SDK knows that, yeah, I can definitely do that. ⁓ So the other question that I get a lot is like, why isn't it just like that? And the reason that it's not just true and it's an empty object instead is because there are sub-properties that you can define ⁓ in there. So ⁓ something to keep in mind.

⁓ but yeah with that ⁓ we have our initialize, we have our list, and now we can click on this and run that tool, and we can actually perform the tool call. So we are all set with our first tool. Congratulations. Let's make this a little bit more complicated in the next one.

—-------------------

Harsh Bharadwaaj (00:00:00)  
Okay, good job on that one. Why was the why did the cookie cry? It was feeling crummy. ⁓ that's sad. ⁓ All right, now is the time for you to take a quick break. write down the stuff that you learned so it doesn't just like in one ear out the other, sort of thing. ⁓ And ⁓ yeah, take some time to ⁓ get some water, get some movement going, get the blood flowing again before you continue learning. It's important to take care of your body and ⁓ also important to retain what you're learning. So write down what you learned.

And then let's move on.

—-----------------

Harsh Bharadwaaj (00:00:00)  
So now we're going to take the tools that we did before and the resources that we've done and kind of put them together a little bit. We're gonna make it so that the tools that are intended for the LLM can actually integrate a little bit more with the resources that are more intended for the application. ⁓ So the application is the one that is actually calling the tool. So when the tool result comes back to the application ⁓ before it hands it off to the LLM, ⁓

Then the application could actually look at the tool results and find some interesting things in there, like resources. So you can actually embed the results ⁓ or the resource in the results of a tool. ⁓ And this is what we call embedded resources. So it looks something like this. This is would be inside of that content array that comes back from the tool. You have a type of resource, you have the resource to contents itself embedded as part of the tool results. ⁓ This is quite useful. ⁓

For clients that integrate with this, because then they can say, great, you've got a resource. I've got this URI. I'm gonna add that to some list of resources, and maybe I'll subscribe to it ⁓ and keep up to date on changes that are happening and make sure that my conversation context has the latest version of this resource, whatever. ⁓ And so this just gives a little bit more power to the ⁓ whatever application is ⁓ controlling or interacting with your MCP server.

You also have the concept of linked resources, which is a similar idea, except maybe you don't want to include the entire text because I don't know, maybe it represents an entire video or a big image or a really long blog post or whatever. ⁓ And so ⁓ maybe you want to say, hey, ⁓ the thing that you just interacted with actually is or can be represented by a resource. If you want to reference that, here you go. And so you can, in addition to whatever else you're returning, you can also return ⁓ as part of that content array.

a resource link. ⁓ And this is going to have a URI, it's going to have a name and a description, ⁓ as well as a MIME type that's like, hey, if you request this, this is the type of data you can expect. ⁓ And so then the client can say, great, I'll just add this to the list of resources that the user can select. Or I'll subscribe to this and make sure that I stay up to date with the latest. Or maybe I'll go and read it right away. And I'll just include it in the context because that's just the way that my client was written, whatever. ⁓ So ⁓

Harsh Bharadwaaj (00:02:19)  
That's the two different ways that you can interact with resources right in your tools. We're going to do that in this exercise, both embedded and linked resources. Have a good time.

—-----------------  
Harsh Bharadwaaj (00:00:00)  
So, not everybody wants to be a prompt engineer, right? So, ⁓ why is the prompts ⁓ grayed out? Well, it's because you haven't actually implemented any prompts yet. ⁓ prompts yet. We need the capabilities listed, and we need to define the prompt. So when you're all done with this exercise step, ⁓ when you connect to your server, you should have prompts available. You can click on that, list the prompts, and you'll have the suggest tags prompt available. Then you can provide the ⁓ ID of the entry that you want tags suggested for.

Click get prompt, ⁓ and that will give you those messages. Then we can look at that output, ⁓ and there you go. You'll be all set. ⁓ So that is the goal for this exercise ⁓ is to ⁓ set the capabilities, ⁓ create the prompt that takes ⁓ some input, and then ⁓ generate that prompt. Have a good time with this one.

—----------------------

Harsh Bharadwaaj (00:00:00)  
If you think about it a little bit, our prompt is not exactly optimal. We're pretty much telling the LLM that it needs to go and get some more information. So ⁓ literally the user just went to our server to get information, and now that prompt is telling the LLM to go back to the server to get more information. ⁓ Wouldn't it be nice if instead of saying, hey, use the get entry tool ⁓ and the list tags tool to get this information? What if we just included that information in the prompt?

Then the LLM could immediately generate the suggestions. ⁓ And then, yeah, still it does need to call some other tools to create new tags and stuff. But that makes sense. That needs to happen after the suggested tags ⁓ happen. ⁓ And so that is what we're going to do in this exercise. So get to it. You're going to add a couple other ⁓ content types. And so when we are all finished, you're going to have what the user is saying, but then you're going to have what the resource ⁓ for the tags.

and the entry ⁓ actually are so that the LLM can use that as additional context to actually generate the suggested tags and it doesn't have to go and request those. So we're just gonna save the user a little a little bit of time, some tokens, you know, resources, all that stuff. It's good. So get to it.

—-------------------

Harsh Bharadwaaj (00:00:00)  
What happens if you're trying to add ⁓ one number to a negative number? Well, in our case, we're going to make it ⁓ give us an error. Technically, this the results of this would be negative two, and that's fine. ⁓ but we're going to make this an error because we want to talk about what do you do about errors. And it's actually really, really simple because the ⁓ SDK is gonna handle that for us. Technically, what needs to happen ⁓ is the response needs to include an is error true, and then it interprets all of this content to be like.

the reason for the error and an explanation of the reason. In our case, we're gonna say the second number cannot be negative. ⁓ Luckily for us, the TypeScript SDK for MCP actually handles this for us automatically. So we don't have to do a whole lot. We literally just have to throw an error and then the SDK will handle that for us. I think that's actually pretty cool. So ⁓ that's your objective here is to ⁓ check whether the second number is an error and ⁓ or is negative and then throw an error as a result. ⁓ should be pretty quick and easy.

But it's important for us to understand how we technically are supposed to handle errors in here. And then presumably the host application is gonna pass this to the LLM ⁓ with instructions that like, hey, the thing you just tried didn't work, and here's the reason. ⁓ And then the LLM can try again or fix its inputs or whatever. So give that a shot. We'll see you when you're done.

—------------------------  
Harsh Bharadwaaj (00:00:00)  
So the first thing that we're going to do is we're going to bring in this initialize resources ⁓ utility from the resources directory. ⁓ And then we'll come down here and we'll await initialize resources. ⁓ And then we'll jump into that. That accepts our agent, the epic me mcp server that we've defined. ⁓ And we're going to create a resource with agent.server.register resource. ⁓ And let's just let our AI do what it thinks it wants to do here.

Of course, sometimes it gets confused with these curly braces. I don't know what's going on. ⁓ but ⁓ yeah, we're gonna need to have that defined. Yep, there we go. Okay. ⁓ So let's take a look. We've got tags. Yes, that's what we want. We want epic me colon slash slash tags. Again, ⁓ this ⁓ schema or this ⁓ URI scheme ⁓ is kind of arbitrary. It doesn't really matter, it needs to be consistent. ⁓ most j general clients ⁓ don't have any like built-in mechanisms for what the scheme is.

unless you are the one building the client or or there's some pseudo standard that's shown up like ⁓ mcp UI is a a standard ⁓ that has a specific URI. ⁓ we don't have anything like that for this and so just using something consistent is sufficient. ⁓ I typically will just do name of app colon slash slash ⁓ and call it good. So epic me colon slash slash tags works just fine. ⁓ And then the title

This again ⁓ is going to be user facing ⁓ as well as the description. So here I'm going to be a little bit more descriptive. So all the tags in the database ⁓ that works. ⁓ And then our async callback is going to accept the URI. ⁓ And we're going to need to include that in the content that we return. We've got ⁓ our database. We're going to go get the tags from the database. ⁓ And ⁓ our MIME type is application JSON. The text is JSON string fi tags.

And then we have our URI. ⁓ And that ⁓ is our first resource. Let's just go make sure that it actually works. So we'll restart our server. We'll go to ⁓ hey look, we have resources now. And actually this is interesting because you'll notice we have our resources listed here. I forgot to add that to our capabilities. So again, the ⁓ CLI or the SDK is going to handle this for us automatically. So

Harsh Bharadwaaj (00:02:20)  
You don't technically need to do this, but I I just like to. I don't know. It makes me feel good inside. But ⁓ I don't know. You feel free to to not list it because the SDK handles it for you. ⁓ So in any case, ⁓ we now have because we have those capabilities, we have the resources tab available. We can list the resources, we can go to tags, ⁓ and boom, now we've got our tags right there. Now let me show you actually, because we have this ⁓ and I've wired this up in Claude, ⁓ I can add

This from the MCP fundamentals. That's what I called ⁓ my MCP server in Claude. And I click that and boom, there's the resource we just defined. Tags. I click on that and now ⁓ I can say please summarize my or let's say list my tags, ⁓ I guess. ⁓ And because I've included this, ⁓ I guess it wants my microphone access. That's fine, I guess. ⁓ But here are my tags. You'll notice it didn't need to call a tool to go get my tags or anything. I included it, and now I'm able to interact with that in some way.

Now the tags is pretty simple, like there's not a whole lot in there, but you'll notice that it was able to summarize it. So like I can do stuff with it. So you can imagine resources could be any number of things that the user might want to say, hey, I want to like talk about this item specifically or something. This will ⁓ be even more applicable when we have a resource template that will appear ⁓ in here as well. So all of your ⁓ the things that you've listed will appear in here ⁓ as well. But ⁓

Yeah, that's resources working in an actual experience ⁓ because we added resource support to our server. Good job.

—----------------  
Harsh Bharadwaaj (00:00:00)  
So it is pretty cool. We can run this tool a million times and we get always the same results. okay, no, it's not that cool. It would be cooler if I could define what numbers are actually coming back. And so here in our ⁓ this step of the exercise, we're going to add the ability to add two numbers. So you're gonna want to change the description beyond just add one and two to add two numbers. Make sure that you're restarting your server and relisting your tools to get that updated description.

And then we can choose the numbers that are added here, and we get a result that includes that specific ⁓ four and sixes ten or whatever. ⁓ So that is your objective here. The key here is that on ⁓ once you have things configured properly on this ⁓ tools list request, it's going to include this input schema. ⁓ And you don't luckily have to write this ⁓ input schema out as it is, but that's how it's defined, so feel free to take a look at this.

We have a really nice tool called Zod to define this input schema. It's really, really nice. ⁓ it's really declarative. It kind of looks a little bit like TypeScript in some ways. So ⁓ that's what you're going to be using. And then the ⁓ SDK is going to automatically convert that from the Zod-specific implementation into JSON schema, ⁓ which is how this works. You can feel free to to dive into what JSON schema looks like ⁓ if you want to. It's kind of interesting. ⁓

But that's how you define your input schema is with JSON schema, and we're gonna use Zod to make it a lot easier to write. ⁓ So ⁓ I think you're ready to go. Have a good time with this exercise step, adding ⁓ arguments to your tool.

—-------------------------

Harsh Bharadwaaj (00:00:00)  
For this one, you need to implement the server and then implement the connection between the server and the client. Right now, if you hit the connect button, you're going to get this error from standard I.O. ⁓ So we're connect console.error. A fatal error in main error not implemented. ⁓ I wrote the text error not implemented. You're going to replace that, delete the ⁓ error that's around there. ⁓ But when you're all finished here, when you hit connect, you'll get a message that EpicMe.

MCP server is ⁓ running and actually you write that one as well. ⁓ and the initialize request is the most important thing here. And so we're gonna have our server info, we're gonna have our instructions, title version, all of that stuff. ⁓ So that is what you're going to be doing for this exercise. Have a good time.

—---------------

Harsh Bharadwaaj (00:00:00)  
Okay, let's go ahead and throw an error if the number is negative. Ta-da\! ⁓ It's ⁓ pretty simple. ⁓ as a bonus, I ⁓ said that you could use invariant. This is a package that ⁓ I made which was based on another package, which was based on another package called Invariant. ⁓ but ⁓ the idea is just taking this and turning it into a one-liner, ⁓ which ⁓ I don't know, it's quite nice. So we're gonna use invariant from Epic Web Invariant.

⁓ if the second number is greater than zero, then everything is fine. ⁓ and actually I think even zero would be fine. So greater than or equal to zero. And then we're gonna say second number can't be negative. So you could do it either way. ⁓ I really like invariant. We're gonna use invariant throughout the rest of this workshop quite a lot. And so ⁓ let's connect and make sure that this all is working as expected. A negative number, second number can't be negative. Again, the tool call has as part of the response.

Is error true? That's what turns it into ⁓ an error. ⁓ And the SDK is really just gonna handle the error that is thrown ⁓ and say, ⁓ okay, you threw an error. I'm gonna take the message of the error, we're gonna make that ⁓ some text content, ⁓ and ⁓ then you can ⁓ we'll we'll send that to ⁓ the client. ⁓ you can technically do this yourself as well. So we could say ⁓ if the second number is ⁓ less than or ⁓

yeah less than zero, then we can return an object that has ⁓ content with is error true manually. ⁓ this would work exactly the same. so restarting, we're gonna clear in list tools, we'll add one ⁓ and a negative one and run the tool. ⁓ Same response here we get second number can't be negative. We get our tool call includes is error true. So you don't have to just throw you can actually still technically return

something that has an is error true and that will work just as well. So it kind of depends on the situation. ⁓ Most of the time though, I'm just gonna throw an error. But one nice ⁓ thing about ⁓ doing it this way is that you get code total control over the content array. Whereas if you're just throwing error, it just takes the error message and uses that as a text content. So if you wanted to include, I don't know, maybe an image that showed a screenshot of the thing that failed or something like that. ⁓ Or if you wanted to have multiple items of context or ⁓

Harsh Bharadwaaj (00:02:22)  
content and one of those was a resource or something like that, ⁓ then that could make a lot of sense as well. Really the key to error handling in MCP is this is error true.

—------------------

Harsh Bharadwaaj (00:00:00)  
So now let's improve this suggest tags prompt to include the stuff that we're asking the LLM to go and grab. ⁓ And that is the list of tags that are already available and the entry that we're asking to get suggestions for. ⁓ So ⁓ Marty the Moneybag helping out ⁓ here to make sure that you can focus on what you're trying to learn here instead of the abstractions that we built for our database. ⁓ And we're also going to want to ⁓ add some validation that the entry and tags do in fact exist.

⁓ and yeah, list tags. That's the API, right? Yep, there it is. ⁓ So we've got our tags, we've got our entries, ⁓ and now we can add two messages ⁓ for embedded resources of those things. So first the user says, ⁓ hey, ⁓ or first we're gonna have the user say, here's what I want to have happen, which by the way I I found that both for humans as well as AI agents, ⁓ it's better to start with what you want to have done and then provide the additional context than to provide the context first and then ⁓

Continue with what you want to have done. ⁓ So we start with what we want to have done, and then here's the additional context. Here's the entry ⁓ context, ⁓ and here are all of the tags. And those are embedded resources, so the application might be able to do something useful for the with them as well. ⁓ And so now we want to update our ⁓ our prompt a little bit to explain that we're providing this stuff. They don't need to use git entry or anything. ⁓ So ⁓

We can say below is my entry with ID of entry ID ⁓ as well ⁓ as the tags I have available. ⁓ please ⁓ suggest some tags to add to the entry. Feel free to suggest new tags I don't have yet. For each I approve, yada yada yada. So it's we're still instructing it to call tools, ⁓ but ⁓

We are not making it call tools that we can act effectively ⁓ handle ourselves. ⁓ So with that now, we come over here, we connect, we go to our prompts, we list the prompts, we suggest tags for entry two, we get the prompt, and boom. ⁓ There we go. We've got ⁓ the basic static prompt, ⁓ although the ID of two shows up in there. So I guess it's sort of dynamic. And then we have our resources, our embedded resources that are going to help the LLM know exactly what we want to have happen here.

Harsh Bharadwaaj (00:02:22)  
And that is your optimized prompt.

—--------------------

Harsh Bharadwaaj (00:00:00)  
You might notice that this actually is a familiar UI element, and we're getting completion requests as we type in here. ⁓ So that's your job is to add autocomplete support to our prompt. ⁓ The API is a little bit different with the SDK, so you're gonna have to figure that out. But once you have it worked out, then the user experience for selecting a specific entry will be a little bit improved. So let's get to that.

—------------------

Harsh Bharadwaaj (00:00:00)  
So here we are in the get entry tool. Let's add ⁓ or convert this into a resource. So here the AI decided, hey, let's add it. ⁓ but there's some duplication there, so we don't need this here. ⁓ So now we have a type of resource. The resource has a URI, ⁓ which points to the URI for this resource, ⁓ and the MIME type and our JSON stringified entry. So ⁓

That is cool. You'll just want to make sure that the embedded resource is exactly the same as the resource in here for consistency reasons. ⁓ And ⁓ sometimes that might require some sort of abstraction layer, something that our abstraction layer is just this get entry thing. We're calling get entry in both ⁓ places here. ⁓ I think that there is room for an a ⁓ framework ⁓ that is built on top of the SDK or something ⁓ that makes doing this sort of thing really easy. So you just say, hey, go give me that resource and stick it in here or something like that. That would be pretty cool.

⁓ but ⁓ yeah this is that's pretty much it. Now ⁓ if we come over here and we restart ⁓ and come over to our tools and get entry number one, we've got our resource here. We can take a look at that. There's our resource ⁓ and clients can do what they will with that. So that is embedded resources.

—------------------

Harsh Bharadwaaj (00:00:00)  
Alright, it's time for your break. You need to write down what you learned so that you can remember it better, ⁓ and you need to have a little bit of a laugh because it will help you live longer and that's good for your retention too. My cat was just sick on the carpet. I don't think it's feline well. Haha. ⁓ alright, great. Now is the time to get your body moving a little bit, ⁓ do some exercise or something, ⁓ and get blood flowing, that's good for your brain. Write down what you learned so you remember it better, and then you can continue with the

You're learning into the future because it's just a wonderful time to learn stuff. So ⁓ go ahead and take a little bit of a break, and we'll see you when you're done.

—------------

Harsh Bharadwaaj (00:00:00)  
Right now, if you get an entry, we'll just run that tool, you're gonna get a bunch of JSON that represents the entry. Pretty reasonable, and the LLM knows what to do with that and can do something useful, summarize it or whatever. ⁓ maybe decide, it needs to update it or add a tag or what whatever it needs to do. ⁓ but this is actually like really similar to what you would do if you were to ⁓ list ⁓ a ⁓ list the resource. It's like basically the exact same content. ⁓ And so ⁓

While it's really useful to just be able to give the LLM the entry, it might be useful for the client to know that, hey, actually, this tool that gets you the entry, this is actually represents a resource. So maybe you can do something with that. Now, I'm not aware of any clients that actually do do anything with that, but this is ⁓ the concept of embedded resources, which some clients actually do do things with embedded resources.

With MCP UI specifically. And so we'll get into MCP UI in a future workshop. But what we're going to do in this exercise is turn this response into an embedded resource. So if I run it in the solution, you'll have something that looks a little bit similar. Let's take a look at the results of the tool call. You'll have the content with a type of resource, and then the resource includes the URI, MIME type, and the text. So it's basically what the resource was just embedded into the tool response. ⁓

⁓ then the client can do fancy things with that. Maybe it can add it to the context for future conversations or just make it easy for the user to add that, or maybe add it to a list of ⁓ past resources that the user can add later on or something. ⁓ but it it's basically the same content as far as the LLM is concerned, but it gives the application a little bit more that it can do ⁓ with the response than just like, this is some JSON, I don't know where it came from, I don't know what it is, I can't identify in the future.

By providing that URI specifically and the fact that this is a resource, the application can do a little bit more with it. So that's what you're gonna do in this step of the exercise. Have a good time.

—---------------------

Harsh Bharadwaaj (00:00:00)  
Probably the most interesting thing for most of you is tools. And the reason that it's so interesting is because it's so novel that we can have a computer make a decision about which tool to call, and then it calls that with the arguments and everything. It's very, very fun and interesting. ⁓ Lots of MCP is about tool calling. There's still like resources, prompts, sampling, elicitation, lots of really other cool things, ⁓ but tool calling is like where most people's hearts are and is where it's most interesting. So we're gonna get into tools in this exercise, tool calling.

So tools in MCP are designed to be model controlled. That's what makes it so exciting: is that the model has a list of tools and it decides which one is most appropriate to solve the user's query. ⁓ the way that this works is the user that this is the typical exercise. Remember that the MC the MCP spec is mostly interested with specifying the clients and server. ⁓ It is going to ⁓ indicate things like, hey, it's designed to be model controlled, but the client can do whatever it wants to with that. So

Everything on the left side of the client here ⁓ that is technically not specified. The specification only really talks about ⁓ or or focuses on the client and server. ⁓ So this is the general when when I show you something like this, is it just kind of the general ⁓ expected ⁓ behavior ⁓ which most applications are going to be implementing. ⁓ So the user enters a prompt, the app sends that prompt to the LLM, and then the ⁓

We we've got a loop going on ⁓ from here. So the ⁓ tool calls are are generated ⁓ based off of this. So the LLM looks at that prompt along with the description of all the tools and whatever other context is included, and the LLM decides, hey, you know what? I need to call a tool to answer this person's question. ⁓ So we're gonna call the tool. ⁓ The app typically is going to ⁓ ask the user first if a tool should be called. So we're gonna have the human in the loop confirmation.

And then when the user confirms, then we're gonna forward that tool call from the app to the client. Now, the app is still like ⁓ managing that client. And so this is whatever protocol they have established, probably just calling a function. ⁓ And so it's gonna forward that tool call to the client with the arguments that were generated by the LLM. ⁓ And then that client is going to make an RPC, ⁓ JSON RPC call to the server. The server will respond, ⁓ and that result gets sent to the app, which the app then will send to the LLM.

Harsh Bharadwaaj (00:02:26)  
And then the LE LLM generates some more text and that ⁓ response is shown to the user. ⁓ and then maybe the LLM decides that there's another tool that needs to be called as a result of that, and that ⁓ process continues. Maybe the user decides, hey, I don't want this tool to be called, and so that rejection is sent to the LLM. And then the LLM says, fine, you don't want to call that tool, maybe you want to call this one, or maybe it just decides, well, if you don't want to do that, I can't do anything for you, and it just finishes ⁓ the generation and sends that to the user.

⁓ Again, the only part of this that's that's actually really specified aside from like some guidance and stuff, ⁓ is the client and server communication. ⁓ And this is largely managed for us by the SDK. We just need to define our tool. And so here is how you define that tool. First, we add some capabilities here. The SDK actually manages this for us automatically, but I ⁓ like to be more explicit and I'll include that in my capabilities.

And then ⁓ here we register the tool with ⁓ an LLM facing name. So the LLM will see this. ⁓ A user facing title. The LLM might see this as well, but this is intended to be shown to the user. The LLM facing description that describes to the LLM this is the context in which you should call my tool. This is what my tool does. If your user has this problem, call my tool. ⁓ And ⁓ often the client will also show this to the user.

⁓ in some contexts as well, so just keep that in mind. ⁓ and then we have an input schema here ⁓ that defines the different inputs that we can have. We ⁓ if with the SDK ⁓ in TypeScript we use Zod, the library Zod to define the schema. But the specification actually requires the schema to be defined ⁓ using JSON schema. So if you're familiar with that, there you go. If you're not you can go look it up.

⁓ but we use Zod and then the SDK will convert the Zod schema definition into JSON schema. ⁓ you also have a description in here, the the name ⁓ or the property, ⁓ this is here. The name ⁓ here ⁓ is ⁓ going to be given to the LLM as well as the description. ⁓ so ⁓ that can be really useful in some contexts where maybe you've got a couple properties and it can be unclear which property is for which thing or s whatever.

Harsh Bharadwaaj (00:04:45)  
⁓ I pretty much always describe my inputs almost. ⁓ And then we have our c it our callback here in which we can do whatever we want to. ⁓ We then return ⁓ standard content, which is an array of different content types, and we'll look at at that in this exercise a little bit more. ⁓ And then ⁓ the client request is gonna look a little bit like this. We're going down ⁓ down a level of abstraction here. So you're gonna have the JSON RPC, you'll have the ID, and then you'll have a method.

Tools call. ⁓ Actually, before this, you have an ⁓ another RPC request that's made called tools list, which will give back a list of the available tools that the host application that can then give to the LLM. So it knows what tools are available. ⁓ So then when a tool needs to be called, it'll have tools call, we'll have the params name hello argument ⁓ with the name codey or whatever, whatever those arguments are going to be. ⁓

And then our server response is going to include that same ID to associate the request and response. ⁓ It's going to have a result and then an array of content ⁓ and whether or not there is an error. If this doesn't exist, then it's assumed that there was no error. But I include this ⁓ to say you can actually handle errors ⁓ and display those errors to ⁓ users in there. ⁓ So ⁓ that can be quite handy.

So that should give you a pretty good idea of how all of this is supposed to work, but I do want to show you this working just in case you've never actually seen MCP in action or you've never done this before. ⁓ Hopefully ⁓ you looked at some of the prereq material and you saw what ⁓ MCP can actually do within the workshop app context with the Epic Workshop MCP server. So I'm gonna pop this open and I'm gonna say, could you please describe this exercise to me and what tools are?

And ⁓ now this is probably ⁓ going, man, it's gonna read it. See, this is actually useful anyway, ⁓ because it the LLM can sometimes decide that it doesn't need to call a tool. ⁓ And in the case of the tool that we're gonna be doing in this exercise, that actually is like probably ⁓ relatively accurate for simple math problems, which is what we're going to be doing. ⁓ the LLM decides, I know what two plus two is, so I'm just gonna tell you and I'm not gonna call the tool.

Harsh Bharadwaaj (00:07:02)  
So that's another thing to keep in mind, and that actually ⁓ goes a lot into the description of your tools, making sure that the LLM knows, hey, I I really can help you here. ⁓ In this case, ⁓ the LLM decided, hey, there's a README here. I'm pretty sure I can describe this. ⁓ let's let's go with something a little bit ⁓ more specific. ⁓ can how about this? ⁓ can you ⁓ quiz me on this exercise?

And now there is a tool in the Epic Workshop ⁓ MCP server ⁓ that is going to be very helpful for this. So first it needs to get exercise context so it can get everything from the exercise. And then it's going to get the quiz instructions so it knows how to ⁓ quiz me. And we can actually take a look at this here, the arguments that are being passed, MCP fundamentals, exercise number two, we'll run this tool.

And it gets all of ⁓ the context that are necessary. Hey, you are an expert teacher. Here's the content for the exercise, yada yada. And now it knows what to do. So that's what a tool call flow is like. You'll notice it actually called ⁓ two tools. ⁓ So ⁓ it is like that's that's kind of what makes it a gen tick, is that it's just a loop. It continues until it's finished with its generation. ⁓ So hopefully that all ⁓ is clear to you. I hope you have a good time with this exercise. We'll see you in it.

—--------------  
Harsh Bharadwaaj (00:00:00)  
So it's neat that we've got a template for the tags, and we could type in a number and read that resource. All of that is working great. If we list the resources, ⁓ it's just showing the tags, and that's exactly how we have it coded. But I think that it makes sense, since we don't have a ton of tags, to just include all of the tags when we list the resources, like this. See, now we've got all the tags. We don't have to go through the resource template, but these are all still built on top of the resource template.

And so if you take a look at the resources list call right here, you're going to see the URI is still epic me colon slash slash tags five eight whatever. ⁓ So that is what I want you to do in this exercise is implement the list callback for the resource template for tag. ⁓ And that way ⁓ those will be included. Now we intentionally don't implement the list callback for entry, and that's because there could potentially be hundreds or thousands of entries.

And including that in the list resources ⁓ call just wouldn't make a lot of sense. That said, there is pagination support for ⁓ the resources list call. ⁓ And so you could add pagination. There's like a cursor and all of that stuff. ⁓ We're not going to get into that in this workshop, but if you have a little bit of extra time, you want to dive into it a little bit, feel free to go look that up and see if you can add that ⁓ pagination support to the entries.

⁓ but I don't really think that that is ⁓ necessary for most servers and I don't actually know of very many clients that implement patination support anyway. So ⁓ maybe someday in the future that will be a more common thing. But right now I think I'm okay with just having to enter the ID for an entry. But for tags, there just aren't all that many of them. So including them in the list makes a lot of sense. ⁓ So that's what you're gonna do in this exercise. Have a good time.

—--------------------

Harsh Bharadwaaj (00:00:00)  
Alright, so first thing we want to do is ⁓ update the description so the LLM knows what to do with this. And so we're gonna say ⁓ add two numbers together. And then we're going to add our input schema. So this input schema is an object where each property ⁓ is a Zod schema. ⁓ and so let's bring in Zod right here. We got this is gonna be import Z from Zod. ⁓ And that

Is going to give us a number. We can actually describe each one of these things. That's part of the JSON schema support has ⁓ support for descriptions of your properties. ⁓ And Zod also has support for that, which is quite nice. And so here we're going to say the first number to add, the second number to add. ⁓ I think that most LLMs will be able to figure out what first number and second number are supposed to be for. But sometimes it can be really useful to be able to define specifically what each input is going to be used for. ⁓

Then ⁓ we get those ⁓ parameters as an object. You could call this args, ⁓ but I always just structure it ⁓ right in place so we get our first number and our second number. But this is technically ⁓ an args object with args first number, second number. ⁓ So then we can take that and update our output to say the sum of first number and second number is the first number plus second number. ⁓ And with that we can come back here. Let's restart. It's important anytime you change your server.

You need to restart because the client s initialized your server ⁓ with the code that existed at that point. And if you change a bunch of code, it's not going to re-initialize automatically or anything. So you gotta restart your server, start a new one. You could wire up a whole ⁓ framework or something that automatically updates things behind the scenes. ⁓ and most of the time that would work out okay, but sometimes you even need to re-initialize anyway. So we're going to restart.

And we're gonna clear our tools and list them again. And great, now we've got add two numbers together, and we can ⁓ add two numbers together. Run tool. Five and thirteen is eighteen. I hope that's true. ⁓ And that is getting our first server ⁓ set up with our first tool that has an input schema. ⁓ pretty cool. Remember that this ⁓ is for the LLM ⁓ primarily, and so ⁓ like ⁓ for a more complicated tool.

Harsh Bharadwaaj (00:02:22)  
You might say ⁓ ask the user for the ⁓ yeah, first number to add, or ⁓ if this was like email, ask the user for the their email address or something like that. Or ⁓ typically I I guess I'd put those instructions inside the description, and then I would use this to s ⁓ describe what this input is gonna be used for. So if this were email, then I would say ⁓ the email or ⁓ to send the ⁓ gift to or whatever, right?

So you're kind of talking to the LLM as if it is a human assistant and you're saying, hey, I understand that you are representing a user here. Here's how you serve the user ⁓ with these inputs, et cetera, et cetera. ⁓ So the other thing I'll mention here ⁓ is technically Zod has an email ⁓ feature, I believe it does, string, maybe it's dot email. Yeah, here we go. ⁓ I

Don't know that I would recommend ⁓ trying this. We could actually try it right now. Let's just see what happens. We'll take our email ⁓ and ⁓ I don't know, I'm we're kinda messing things up. Text. ⁓ here we go. The email. Ta-da\! ⁓ Second dollars. Ooh, man, that would be a really nice tool. So if we come over here, we restart ⁓ and list our tools again, ⁓ then we could say ⁓ hello ⁓ at example.com and send this.

run the tool. It actually does work. The reason that I'm hesitant to do this though is the more complicated you make this stuff, ⁓ the less likely it is going to be able to convert your ⁓ input schema to the JSON schema. Because it really it is the JSON schema is what is defined in here. ⁓ This probably worked because here it's first defined as a string. And so here in our properties in email it's just gonna say it's a string.

and actually it says format email. Man, that's kind of neat. ⁓ but there are other ⁓ things that you can add to Zod that don't translate directly to a JSON schema. And so avoid making your JSON schema overly complex and instead do some of the validation inside of your function, which incidentally is what we're gonna be getting into next. So let's move on to the next.

—------------------

Harsh Bharadwaaj (00:00:00)  
So here's how we do this. We're going to bring in completable from the SDK, and then we're going to take our Zod schema definition for the entry ID, and we're going to wrap it in completable. ⁓ completable. There it is. ⁓ We'll stick that in there. And then the second argument is the completion. ⁓ And so that's going to be async. It's going to take the value of what the user has typed. We're going to get all of the entries. We're going to map those ⁓ to just be their ID. That's what we can filter on.

And then we're going to filter down to the IDs that include the value that's been typed. ⁓ And with that, then we can come over to our dad jokes. No, just kidding. We'll come over here, ⁓ we'll restart our connection, we'll go to prompts, we'll query that and list it again. And ⁓ as we type, we should get completions. So there you go. Now we'll get the prompt ⁓ and we've got it with ⁓ ID of three. So that is your completions.

you have this completable utility that you provide our ⁓ Zod ⁓ schema and then you provide your completion. A little bit different from the way that we do ⁓ resource templates because the you don't actually define a Zod schema for resource templates, so it kind of makes sense. But ⁓ that is how we do completions for prompts. Good job.

—-----------------

Harsh Bharadwaaj (00:00:00)  
Hey, good job on this first exercise. Let's go to a dad joke. Chuckles are good for ⁓ retention because it makes you happy. ⁓ So here it says, What did the calculator say to the student? You can count on me. Haha. ⁓ that's maybe an appropriate joke for this first exercise. ⁓ so now is your time to write down what you learned. That helps with your retention. ⁓ And maybe take a a sip of water. I need to refill my water bottle, it looks like. ⁓ And ⁓

Then you can get into the next exercise. We'll see you in the next one.

—---------------

Harsh Bharadwaaj (00:00:00)  
Hey, awesome job on this. That is the fundamentals of MCP servers. So you've handled ⁓ ping and just setting things up. You added tool calling, you added resources, and you added prompts. There are, of course, a lot more things that we can do with MCP, but those are the basics, the fundamental, the buildings blocks, the most important pieces that you need to understand about building MCP servers.

And now hopefully you can take this and apply it to whatever context that you're building services for. You can expose your services to LLMs through the tools. You can expose your resources through resources. You can auto help automate some processes with prompts. We've done a lot of cool things today and ⁓ thought a lot about the user experience. I want you to go out and experiment with this stuff. You're doing awesome. Keep up the good work and we'll see you in the next one.

—-----------

Harsh Bharadwaaj (00:00:00)  
Let's get into resources. So, right now, when you connect to your server, you just have tools and ping, but there's this resources thing. What is that? I want to click on that, but I can't. So you're going to be defining some resources in our MCP server. ⁓ we're going to have a resource for all of the tags available in our journaling application. ⁓ So if we look at the solution, when you're all finished, you should be able to look at the resources. You'll click list resources and there will be tags here. You click on tags.

And here is the contents ⁓ of our resource. It's an application JSON ⁓ and ⁓ the URI is Epicme colon slash slash tags. And then if you take a look at the request history here, ⁓ when we clicked on relist resources, this showed up here. ⁓ And if you look at resources read, then you'll see ⁓ all of the ⁓ that ⁓ contents for that specific resource. So you're gonna define that resource and then you you

you're going to have define a callback that returns the resource contents. ⁓ If you click on list templates, you shouldn't get anything yet. That's for a future step in this. So ⁓ let's get into resources. Have a good time.

—--------------

Harsh Bharadwaaj (00:00:00)  
All right, let's go ahead and add a list callback here. So this is going to be an async callback. We're going to get all of the tags ⁓ and then we're going to iterate through ⁓ all of those tags and turn them into resources. So this is an object that has a resources property. We're gonna take each tag. ⁓ its name is going to be the name of the resource. We're gonna have the URI that points to that resource, ⁓ and that just is the tag ID, and then the MIME type. So this is what

basically it's what you can expect if you were to request this resource. That's not describing this content, it's describing the content that they would retrieve if they retrieved that resource. ⁓ Now you can actually also have a description here. And interestingly enough, ⁓ our tag, ⁓ part of the the property of the tag is the description. And frankly, that's like pretty much the entire resource, ⁓ with the exception of like created at and updated at and stuff, ⁓ which is kind of funny.

You typically ⁓ don't want to include the entire contents of the resource in here though. So ⁓ and and that's kind of specific to our case, the fact that tags actually have a description. ⁓ but for like if we were doing this for entries or something, you wouldn't want to include the entire contents. Part of the point of having ⁓ the listing and the detail is that the listing doesn't have as much information, so it it's not using as many bytes over the wire and stuff.

⁓ and users can dive in deeper for like with the specific resource. So I'm gonna leave the description off here. Like ⁓ if I were to put a description, it might be like ⁓ tag, ⁓ tag with ID, whatever, but it's kind of like saying it is what it is. ⁓ so I'm not gonna include that either. Description is is optional. You can add it, and in some cases it would make some sense. Maybe you have ⁓ blog posts and there's like the SEO description or something like that.

So that might make sense to put in here. Remember that you're communicating with a user here. This is a user controlled ⁓ resource sort of thing. That's the intention intended ⁓ interaction model here is that the user is selecting these resources. ⁓ And so that's why ⁓ that's gonna be in here. So ⁓ we've got this implemented. Let's make sure that it actually does work. So let's restart ⁓ and we're gonna ⁓

Harsh Bharadwaaj (00:02:23)  
Clear our re list of resources. We'll list them again and boom, there we go. Now we've got our resources. We can drill down on each one of these to get the actual contents of the resource, which does include the description and created at an updated at along with the name. And there you go. That's the list callback.

—---------------

Harsh Bharadwaaj (00:00:00)  
In this exercise, we are going to build the server and support ping. Luckily for us, the SDK for MCP actually supports ping out of the box. So as long as we get our server created in the first place ⁓ and connect it over the transport layer that we want, ⁓ it should automatically handle ping and initialization requests for us automatically. ⁓ So here is how things work architecturally in the MonoContext Protocol spec.

You've got the host application. This is your ⁓ VS Code and your cursor or your cloud or your chat GPT or whatever. The thing that is managing the model context protocol ⁓ along with the ⁓ AI model that it's going to be interacting with. ⁓ So the host application initializes a client on its side of the interface. And then on the server side of the interface, it's going to ⁓ we're going to implement that side. ⁓ The client is going to make an initialization request, and the server will respond with its capabilities.

And then we ⁓ have some active ⁓ session with the negotiation of those negotiated features. ⁓ So ⁓ the client n ⁓ at this point can start making requests. So user or model initiated actions ⁓ go to the server, the server responds with whatever response it's going to, ⁓ and then the UI gets updated or or whatever. The client is just gonna communicate those updates to the host. And at that point, it's completely out of spec. Whatever the host wants to do, it can do.

There are some recommendations on how it should handle different things, and we'll talk about those as we get to them. ⁓ But the protocol is primarily concerned with ⁓ the specifying the client and server ⁓ communication. ⁓ Server can also make requests, so like sampling, elicitations, ⁓ logging and notifications, progress, all that stuff. And so ⁓ this is actually kind of interesting because it means if you want to be able to do these things.

Then your architecture has to support the ability for the server to proactively send requests to the client. ⁓ So this means you can't just do request response. We do need to have an open connection where the server can send a response. We're going to be using standard I.O. in this workshop, ⁓ but if you use the HTTP streaming API, ⁓ the way that that works is ⁓ the request for initialization is made, and then that request is upgraded to a server sent events. ⁓

Harsh Bharadwaaj (00:02:22)  
request stream or an event stream ⁓ so that the server can s send proactive ⁓ notifications back. And so you need to use some architecture ⁓ deployment ⁓ situation ⁓ that supports that. ⁓ I prefer personally, I prefer Cloudflare. I'm really big fan of Cloudflare. ⁓ but there are other hosts that support that sort of thing too. ⁓ You do want to be careful if you plan on having more than a couple ⁓ a hundred connections at once.

Some hosts are able to manage that better better than others. And that's one of the reasons that I really like Cloudflare because it can manage ⁓ a basically infinity infinity ⁓ open sessions at once. ⁓ And you only pay for when your code is actually running. It's pretty remarkable. ⁓ In any case, ⁓ the server can send requests. The client will send those over to ⁓ the host, and the host decides what to do with that. ⁓ That response comes back, ⁓ and then the server can do what it will with that.

⁓ lots of the the server requests actually ⁓ or some of the server requests don't actually expect a response in return. Elicitation and sampling do for sure, ⁓ but ⁓ sending notifications and progress and stuff like that, the that it doesn't care about the response. ⁓ In any case, ⁓ here's also the notifications. ⁓ and then the host, if the user closes the application or something, it's going to terminate the client and then the client will send that to the server so the server doesn't have to keep its connection open anymore.

That's the basic idea. Again, most of the ⁓ spec is concerned with ⁓ specifying the communication between the client and the server. ⁓ And so for ping in sp specifically, the client is going to be sending JSON RPC requests. And that actually is the case for all of the requests. Everything is JSON RPC. If you've never worked with that ⁓ before yourself, luckily the SDK actually implements ⁓ all the conversion of JSON RPC and managing the messages and stuff.

⁓ automatically for you so you don't really have to worry about it too much, but understanding it is better for you to be able to use it effectively. And so here's basically what the each sort of request looks like. It's going to have JSON RPC, it's going to have an ID, and then it's going to have a method. It is also potentially going to have parameters and we'll look at those and r and arguments and things. ⁓ we'll look at those when we get into tool calls.

Harsh Bharadwaaj (00:04:38)  
Then the server response is also going to have a JSON RPC. It's going to have that same ID to associate a request and response. ⁓ And ⁓ then it will have its results. And that we'll talk about a little bit more in the future as well. ⁓ The important thing here to note is that the IDs are the same. That's how those requests and response ⁓ are ⁓ correlated. ⁓ And that's interesting because in HTTP world, if you're coming from a web dev background or something,

you're probably used to request and response just happening ⁓ managed on that ⁓ same connection. That's not necessarily the case when it comes to MCP. ⁓ and that in large part is because the transport is ⁓ often it's going to be HTTP, but not always. And so there's not always a request ⁓ a connection that you can rely on. ⁓ so having a specific ID to associate one request with another is important. ⁓ So

⁓ that takes care of getting you all ⁓ ready for this exercise. ⁓ I hope that you are looking forward to it and that you have a really good time with this one. We'll see you in the exercise.

—-------------  
Harsh Bharadwaaj (00:00:00)  
When you connect to your server in this exercise, you're gonna notice that we can ping, but that's like all we can do. There's nothing else we can do. ⁓ What your goal in this exercise is to actually support tools. So when you connect, you'll actually be taken straight to the tools tab. You can still go to the ping and ping all you want, but we want to focus on tools. ⁓ With tools, we want to be able to list the available tools of our server. And here we have an add a tool. We can clear that and list again. Sometimes if you change the definition of your tool or something,

You're going to need to restart. And even if you restart, ⁓ you may need to clear and list the tools again. So just keep that in mind. But here we're going to have our add tool. We click on that and now we can run this tool. You'll see throughout ⁓ all of the interactions that you have, you're going to have a history of the requests that are happening in here. So here when we call the tool, we're going to have a method and params. ⁓ You'll notice also that we're not getting the JSON RPC 2.0 in this request.

body, we're not getting the request ID and a couple other things. So ⁓ keep that in mind that it's only showing you the parts that are specific to this request and maybe that they've decided are really relevant for you to see. ⁓ If you're using another ⁓ tool to ⁓ interact with your MCP server, you'll probably see all of those requests. Now like Postman, for example, will show you everything in the request. But it's just something to notice about this version of the MCP inspector.

So, yeah, we've got ⁓ our method, we get our params, and then here's our response text, the sum of one and two is three. ⁓ That is what you need to implement in this ⁓ s first step of the exercise. ⁓ And ⁓ you should be able to list the tool and actually call it. So ⁓ let's take a look also at ⁓ the listing here. So we get the tools list, and then you get an array of tools. You don't have to implement this piece because this is gonna be handled by the SDK for you automatically.

⁓ Another thing that's actually technically handed handled by the SDK automatically ⁓ is this capabilities listing. Once you've registered a tool, the SDK is going to say, you have a tool. I'll make sure that when we are initialized, we have tools included in your capabilities. I kind of like specifying the capabilities ⁓ because you can actually dynamically change the things that your server supports. ⁓ And I think it's useful to say, hey, yeah, I do support tools or I do support resources.

Harsh Bharadwaaj (00:02:23)  
even though I don't necessarily ⁓ give you any of those right off the bat. ⁓ so I typically like to ⁓ specify my capabilities manually, but you I I just thought I'd ⁓ mention you don't actually necessarily have to. That will probably make a little bit more sense when you get into the exercise. So without further ado, let's get started.

—-------------------

Harsh Bharadwaaj (00:00:00)  
⁓ Okay, good job on this exercise. Now it's time for a break. Here's the dad joke. What do you call a magician who has lost their magic? Ian. Alright, magician Ian, right there. Is that ⁓ working for anybody? ⁓ Alright, awesome. Well done on this exercise. Now it's time to write down what you learned so you remember it better. ⁓ Get up, move your body a little bit, get some blood flowing, get a drink, ⁓ and we'll see you when you're all done with that and ready to move on to the next one.

—-------------------

Harsh Bharadwaaj (00:00:00)  
I saw a documentary on TV last night about how they put ships together. It was riveting. Haha. ⁓ I actually learned how to rivet when I was a Boy Scout. That was one of the merit badges that I got. ⁓ so riveting is interesting. ⁓ and hopefully you found this exercise to be riveting. And now it's time for you to write down the stuff that you learned so you retain it better, and then we can move on to the next one. But don't do that before you at least get your body moving, get the blood flowing, get a drink, ⁓ and ⁓ then you can get

into whatever you're gonna do next with the most amount of energy. So good job on this one. We'll see you in the next.

—------------------

Harsh Bharadwaaj (00:00:00)  
Alright, so here in our index, we're gonna bring in initialized prompts. ⁓ We're going to add the prompts capability, ⁓ and then we're going to ⁓ add initialize prompts to here. So then we'll dive into that. We'll bring in Zod because we're going to use that for our input schema definition. And then we're going to say agent server register prompt. Our prompt will be called suggest tags. We'll give it the title of suggest tags, something more readable for the human.

And the description also for the human. Now note that again, prompts are for the humans, not for the application necessarily or for the LLM. ⁓ The results of the prompt is for the LLM, but the prompt definition itself is going to be viewed by the human in some sort of interface that the application provides to the user. And so the title and description are just for the user. They're not really intended for the LLM. ⁓ And then we've got our ARGS schema, ⁓ which is.

⁓ this ⁓ entry ID, I don't think we actually need that in a Z object. So ⁓ yeah, but that's our ⁓ arg schema. We're gonna expect an entry ID. ⁓ you'll need to note that ⁓ these do have to be ⁓ strings. They cannot be numbers. All of the arg prompts are expected to be strings. ⁓ So ⁓ the fact that you have to define z dot string is kind of funny in that context. But ⁓ anyway, let's go ahead and add our prompt.

This is going to accept our args. ⁓ we should get our entry ID, ⁓ type ⁓ safe and everything there. ⁓ and then inside of here, we're gonna grab ⁓ here, this is actually getting ahead of us. We'll get to that here in a second. ⁓ so we're gonna create our content, ⁓ which actually this is gonna be messages. That's my mistake. Messages, ⁓ and that's gonna be an array, and each of those messages.

⁓ has a role. So who this message is from, user in our case, and the content, what the user said. And the type of content here is text. ⁓ As you can see there, there's a bunch of other types of content. You can have, the use source the user added this resource or here's this audio or image or whatever. ⁓ So we're gonna do text for ours. ⁓ And then the text ⁓ is going to be, yeah, suggest tags for entry and

Harsh Bharadwaaj (00:02:18)  
Whatever, at this point, now you're prompt engineering, and that's a little outside of the scope ⁓ of our MCP ⁓ workshop here, though I do kind of give you ⁓ this this sort of prompt. So let's let's see if we can ⁓ do this. ⁓

Alright here. ⁓ And we'll see if I can get my LLM to do this for me because I don't want to write this. Who wants to write their own prompt? ⁓ okay, please ⁓ suggest tags for the follow or for ha ha. ⁓ Yeah, the following entry use get entry tool and list tag, yada yada yada yada yada. Okay. And at this point you're gonna just play around with the prompt and make sure that it the llm is doing the right thing for you. ⁓ let's say w for the ⁓

Entry entry ID, right? There you go. ⁓ and maybe we'd say ⁓ with the ⁓ ID of entry ID, right? Something like that. ⁓ so yeah, I'm I'm just ⁓ you get the idea. Now you're just gonna play around with the prompt. But let's talk about the MCP side of all of this stuff. ⁓ So just tags is gonna be basically a unique identifier for the prompt. The user should just see the title.

And then they have the description in here. And then we get our arg schema. And again, this does have to be a string. You can't put a number in here. That's gonna make this not very happy. So it's gotta be a string. ⁓ And then we can convert it to a number as needed or ⁓ whatever. And then you return an array of messages. Now you notice the role is specified. It can also be an assistant. And so ⁓ really what you're doing right here is you're

i if you take a step back and think about a chat conversation with an L LLM, ⁓ every single time the user submits a chat and a new generation is is ⁓ created, ⁓ you can think of the LLM almost as a pure function. This is oversimplification, but think of it as a pure function that accepts an array of messages and then appends a new ⁓ or returns the the next item that should be in that array, right? ⁓ And so if we think about it that way, you could have a prompt that's like

Harsh Bharadwaaj (00:04:32)  
conversation between user and agent and then ⁓ and then the LLM can generate what comes next. ⁓ So it is in my mind, I can't really think of a situation where I would want to have user agent, user, agent, user agent conversation as part of my prompt here. I'm not sure why you would do that. ⁓ Perhaps maybe you want to give the agent some examples of how they should be talking to the user or something. I I suppose that could make sense. ⁓ But in all the prompts that I have made, I've just used role as user.

and then again that you have multiple content types that you can provide in here. And based off of the different content types, ⁓ you have different ⁓ kinds of content properties that you can have as well. So if you had a resource, ⁓ resource, why is this not yeah, there we go. If you had a resource, then you'd have a resource property. If you had image and audio, you'd have data blobs, ⁓ stuff like that. ⁓ So in our case, we're just doing text, and that's most prompts are just gonna be that as well.

All right, ⁓ let's just make sure that this is actually working. We'll go to ⁓ restart our server. We got our prompts, list prompts, there's our tags, ⁓ provide an entry ID, and boom, there we go. ⁓ now we've got that prompt right there. ⁓ and again, you just need to kind of ⁓ play around with the prompt. This is you being the prompt engineer for a bit and create something that's really useful. ⁓ Now, from a practical standpoint, how was this used? ⁓ I've got some prompts that are in some MCP servers that I have installed.

And ⁓ every application is going to deal with prompts a little differently. Claude deals with it differently than VS Code, then ⁓ which is different from ⁓ cursor. But in several ⁓ editors at least, the way that you access prompts is with a slash. And so now I can say Epic Shop quiz me, and now I didn't just need to provide the URL for or the the file path to the workshop that I want to work in, and then it it can complete the prompt. ⁓ so that's sort of the the workflow here.

for the ⁓ using different prompts is just add a slash and then you can type in which prompt you want and then it will ask for the inputs ⁓ and then it will just stick the prompt in its place. ⁓ So that is the practical application of this. ⁓ Hope that is interesting and fun. ⁓ let's go ahead and improve this prompt a little bit next. We'll take a look at that next.

—--------------

Harsh Bharadwaaj (00:00:00)  
Let's add some completion so the user experience is better. We're gonna add a complete property, and this is going to have ⁓ a ⁓ function for every property in our resource template that we want to have some support for completion for. We only have one, but you can actually have multiple. ⁓ And then what's interesting is actually ⁓ your completion can accept ⁓ context, which will be the completed values of the others. And so if you were GitHub and you were saying, okay, here's

the the organization that they've already completed. Now the repo name can be completed as a filter of the organizations that have been specified already. ⁓ But we only have one, so we're gonna just stick with the one. ⁓ And we'll say async id ⁓ is our completion function. It accepts a value. ⁓ We're going to get all of the tags in the database and we're going to filter

By the ID. We can't filter by the name. The AI thinks that's ⁓ what we're going to do. We can't do that. We have to filter just by ⁓ the ID. ⁓ we do want to be able to ful filter by the name, and that I actually have an SEP open right now to enhance that ⁓ completion API so that the user can type in something that's more sensible for them. ⁓ but right now the specification only supports completing by ⁓ the the property that

it's going to be completed too. And for us, that's going to be the value. ⁓ So we grab all the tags, we filter ⁓ by the tags that include that value. And then ⁓ we don't need to map or anything. We just yeah, we just ⁓ filter to that. Now ⁓ we actually we do need to map my mistake. But I ⁓ in my solution I do a dot map here. ⁓ we can get the tag and tag dot id. ⁓ and then we can just take the ID ⁓ and ⁓

tag tage there we go ⁓ and that should return an array of strings ⁓ right ⁓ ours ⁓ are numbers ⁓ but you have to return strings so we're gonna do a dot to string ⁓ there we go and then we can get rid of the two string here there we go that looks nice so we're going to take all the tags we're going to map them to ⁓ the value that is completable ⁓ the string ID and then we filter those

Harsh Bharadwaaj (00:02:19)  
which include the value that the user has typed. And that gets our completion working. We can do the same thing for our entry. So hopefully the AI can solve that for us pretty easily. So we take our ID, ⁓ we get all the entries, we map to their IDs, and then we filter. ⁓ Now of course like depending on whatever system you've got, ⁓ maybe you have ⁓ you like something where you could just do this in SQL or something, make it more ⁓ performant or whatever.

⁓ but ⁓ and ⁓ you also want to make sure that you limit this to only the first one hundred items. So you might add a slice zero to one hundred ⁓ to or would that be ninety-nine? ⁓ z index zero to index ninety-nine. So that way you only have a hundred values because that is the specification ⁓ says only a hundred values are allowed. ⁓ but yeah, that takes ⁓ takes that. I think maybe is slice is that gonna take

Yeah, end number. So should we do 100? Something like that. Anyway, you you could figure that out. ⁓ The point is you filter down to those which match the value that the user has typed. And then you make sure you only send 100 back. ⁓ So now if we connect, we go to ⁓ list our resources. There's that's still working. We go list templates, we can say entry, ⁓ and again with the client, I would like it if it just like automatically started ⁓ doing completions, but it doesn't until you actually type something.

Guess kind of makes sense. ⁓ And then we get that completion request ⁓ as ⁓ we are interacting with this thing. And you can take a look at that completion. So here we get the request argument, name ID, value is one. Here's the reference, the thing that we want to complete. ⁓ The SDK interprets all of that and calls our callback. And then we return the values. ⁓ It's an array where the first item is one in this case. ⁓ There is a pagination support here as well.

Where you can say whether or not it has more. ⁓ but ⁓ yeah, we're we're not gonna get that far into it. Feel free to ⁓ dive in deeper if you would like to. ⁓ and then here's that completion where the value is empty. All of these things include an empty, ⁓ and so they all come back, and that's where we end up. ⁓ So that is completions. It improves the user experience in selecting resources from the resource template. ⁓ I know at least VS Code supports this ⁓ as of the

Harsh Bharadwaaj (00:04:44)  
the time of this recording and of course more and more ⁓ clients as they add more support to MCP will support this as well. So definitely something worth implementing for ⁓ your MCP servers.

—---------------

Harsh Bharadwaaj (00:00:00)  
So here we are in the list entries tool. We're going to take this entry links that we're currently mapping entries to, and we're going to change it from embedded resources to resource links. So instead of resource, this type is going to be a resource link. And then we're going to set the URI, the name, description, and the MIME type. And then everything else can just go away. ⁓ We can get rid of that. ⁓ And ⁓ now it's a lot more terse and it ⁓

has the URI along with ⁓ everything else so the client can make use of that. We also have a list tags if you want to do the same thing here. ⁓ So we're going to update that and ta-da, that looks great. Super. And so now we can come back over here, we can restart, we can go to our tools ⁓ and list entries, run that, and now we've got our entries ⁓ all listed right here. We can take a look at the tool call.

And here's what those entries ⁓ were set to. You can click on each one of these ⁓ and then ⁓ a resource read will be made so that it can retrieve the actual resource. ⁓ So there you go. That is what it takes to make a resource link. We're just mapping all the tags to resource links and then returning that. ⁓ you'll notice also that I am adding ⁓ this text that is just kind of prose, and that's because ⁓ remember that we are talking to an LLM here, and so we want to communicate.

Hey, you asked me to list tags. I found this number of tags, and then here's the information about them. And the LLM is going to see ⁓ those and be like, great, now I know what the URI for those are, I know what their IDs are, and then the client can do something with that or whatever. ⁓ So adding I I typically for most of my ⁓ tool responses, I'm going to include some pros ⁓ that or some just some text that describes that what they just did was successful or not. And I found that to be pretty useful. So there you go. That is resource links.

—----------------

Harsh Bharadwaaj (00:00:00)  
So when I go to a list template and I fill in the ID here, ⁓ I happen to know what the ID is, but users probably won't ⁓ just happen to know what the ID is. ⁓ And ⁓ so it would be nice if when I go in here and I type in some numbers, there's some sort of autocompletion going on. So like I can type in something. And what would be even better is if I could type in like some text here and it would autocomplete. Unfortunately, that feature is not yet in the spec. I do have

A pull request to the spec at the time of this recording to add the ability to search for a display name or something like that. That's currently not in the spec. ⁓ Eventually it will be. But at least being able to kind of like autocomplete some of the numbers would be nice. And if you're watching carefully, you might notice there are requests being made down here for autocompletions. And so ⁓ it ⁓ is trying to get an auto-complete so that it can give the user a little bit more help when searching through these resources.

But we haven't implemented that yet. And that's your job in this exercise is to add support for auto-completion. ⁓ And let's take a look at what that looks like when you're all finished. I'm going to start up the solution, which you can of course do at any time to just see whether ⁓ you're doing things the same way or something. ⁓ And with that then we can come here to our list templates. ⁓ I can type in a number, ⁓ and then those things show up. Now the ⁓ inspector at the time of this recording also ⁓ does not

have super great support for ⁓ autocompletion, you have to first type in something and then it will actually start doing autocompletes. ⁓ so if you just hit this ⁓ here, if we go to yeah, that there's there's a bug here too. So ⁓ but yeah so you might have to type in something a little bit to get this thing to show up. But once you do then you should see some autocomplete going on. And we are going to implement this both for tags and for resources. So why don't you get into it?

Have a good time.

—----------------------

Harsh Bharadwaaj (00:00:00)  
Our journaling application has more than just tags, it also has entries. And while there could be maybe 20 or 30 tags and they just have their name and description, typically not a lot of data. ⁓ like this is not not a lot, it's fine. ⁓ Your entries, you could have hundreds or thousands of journal entries. You wouldn't want to have a single resource that lists all of those. Or imagine you're building ⁓ an MCP server for GitHub, you don't want to have

a single resource that gets you all of the repositories or all the files in a repository or something like that. You want to be able to ⁓ have some sort of ⁓ mechanism for retrieving specific ⁓ resources. But having to do a server.register resource for every single one would be kind of annoying. So instead we can make a template ⁓ of a resource ⁓ and then the parameters of those that template can be filled in. Similar to a URL, you have github.com slash username slash repo.

And you fill in username and repo, and poof, now you have the ⁓ web page for that. You have a same sort of concept with resources where you could have GitHub colon slash slash ⁓ username slash repo. Fill those in and you get different resources. That's what a resource template is. So here we don't have any ⁓ resource templates defined yet. That's your job for this exercise. ⁓ In ⁓ the finished version, we do have resource templates. We'll have a tag and entry.

This will allow you to specify an ID. All of our IDs start with zero, and we in your test database you excuse me, they start with one. ⁓ in your test database, you will have like four or five ⁓ of each tag and entry. So you can test that out ⁓ with each one of these and be able to read ⁓ those resources. So your job is to create a resource template for ⁓ tags and entries. Have a good time with this.

—------------------

