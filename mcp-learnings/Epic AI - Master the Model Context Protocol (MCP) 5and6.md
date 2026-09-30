101

Harsh Bharadwaaj (00:00:00)  
So the boss found out that you could literally just go to any one of these URLs and you could see journal entries. Sweet, journal entry there. And you can even go ⁓ journal viewer and boom, now I've got the like all of the journals and I can use those now. Of course, like none of these buttons work because that's like we're communicating with a parent frame, we're not a parent frame, so that that's they don't actually work. But ⁓ you can see it all. ⁓ And of course that's kind of been the case the entire time with this MCP server.

We haven't done any authentication or authorization and stuff. ⁓ Well, the boss decided that we want to start locking things down. And so the problem is that our ⁓ these routes are actually reading the database ⁓ when we arrive at them. And it's just grabbing this ID and it's going to read the database with that ID. ⁓ and so there are a couple of ways that we could ⁓ could deal with this ⁓ particular issue. ⁓ when we get into authentication, we're gonna have a token and maybe we could have the

LLM pass the token through or something. I don't feel like that's not very safe. ⁓ maybe we could put it as a query param to the iframe. Also doesn't feel very safe. ⁓ One thing that we could do is we already acknowledge that the LLM can access all of this data, right? And so what if we pass this data ⁓ through the the ⁓ MCP client and have the MCP client send it along to our iframe? And so the basic idea here is ⁓ the LLM calls our tool.

That tool responds with UI, and that response also includes the data that needs to be passed into the UI. And then when the iframe is created and it says, hey, I'm ready, ⁓ the host application can send ⁓ that data that we sent from our tool, send it into the UI. So that's what we called render data. And it's we just need to configure our ⁓ server tool to send that rend render data, and then we need to configure our UI to accept that render data.

And ⁓ then when we're all done, you'll be able to just go to UI slash entry viewer and that will just have a spinner waiting for that ⁓ that data. And so you won't be able to just see it. ⁓ it won't look like this, it'll just have a little spinner when you're all done. ⁓ So that is the objective in this exercise. we're going to start locking down things so that you can't just view it, but you ⁓ it you have to be authenticated through the MCP server. We'll get to the auth stuff in another workshop. Don't don't get tied up with that.

Harsh Bharadwaaj (00:02:25)  
But this gets us into a place where we could have UI that ⁓ only shows authenticated data because it's coming through from the parent ⁓ via the render data stuff. So you're gonna write a little utility that will help this will resemble a lot of the tool calling stuff like post this message, wait for the response, and all of that. ⁓ so you'll also be actually adjusting the routes because right now ⁓ we are expecting an ID, but ⁓ in the future, ⁓ and in this exercise,

We are not going to expect an ID. So you'll have to adjust the route. You'll have to adjust a little bit of some of the loader stuff. Actually, some of that's already been done for you. ⁓ And then yeah, you'll just need to expect that data. ⁓ Okay, that should be enough for you. And then when you're all done, you should be able to ⁓ still click on view ⁓ entry and it will still show. But this time ⁓ it will ⁓ not just be ⁓ getting the entry ID from the URL, but it'll actually get the entry data.

from our tool response. ⁓ All right, I've said enough. Have a good time with this one.

—---------------

89

Harsh Bharadwaaj (00:00:00)  
Time for a joke. Here it is. The first time I got a universal remote, I thought to myself, this changes everything. Haha. ⁓ Actually, here's the thing about universal remotes: they stink. They're really awful because they basically are like, we ⁓ we are useful for nobody because we're useful for everybody. We like we're gonna put all sorts of buttons on there that make sense in some context and not others. They're basically least common denominator controllers.

And they are awful. I do not like universal remotes. ⁓ not fun. Anyway, ⁓ there's there's probably an analogy to programming in there. ⁓ it's time to get a drink, time to write down the stuff that you learned so that you can remember it better, and and then go give somebody a high five or ⁓ I don't know, ⁓ kiss your spouse or something, ⁓ and then come back and we can learn something together. We'll see you ⁓ when you're back from your little break.

—-----------------

81

Harsh Bharadwaaj (00:00:00)  
For this one, we've got a new tool. It's called View Journal. It allows us to view our entire journal. ⁓ And this one is a little bit different. Now you have a ⁓ a URI that points to view journal slash ⁓ sum number. The only important thing here is that that is ⁓ r ⁓ unique, and so we actually just ⁓ use date.now and that works sufficiently well. ⁓ you could ⁓ create a a random y ⁓ ID, but anyway, ⁓ it

So we've we have this view journal URI ⁓ that has the UI prefix. That's the important part to make it MCP UI. ⁓ the MIME type is a text URI list. ⁓ And ⁓ and then the text is a list of all the URIs that ⁓ are okay to try. And so you typically just have one, but you could actually have multiple and ⁓ it will just try ⁓ until the first one actually works. And so it's good for fallbacks and s stuff like that. ⁓ But here's the interesting thing ⁓ is that it's pointing.

To localhost ⁓ 59021\. I'm running ⁓ the ⁓ the solution on this example. So your port number will be different. ⁓ But ⁓ yeah, so ⁓ it's the same as the MCP server. So the MCP server is also serving the UI, ⁓ umI journal viewer. And so if we ⁓ copy this and paste it over here, ta-da, there it is. It's sometimes you're gonna get errors like this. ⁓ it's a VT thing and it's super annoying.

⁓ literally just click anywhere or hit refresh and that should go away. You'll sometimes actually also have problems starting the ⁓ or initializing that connection ⁓ with the inspector or other tools ⁓ when we're running in dev mode like this. Once you've deployed it, then it's not a problem. But during dev, ⁓ the first time you make a request, sometimes it's a pain. hopefully I can make it so that you don't ever see that problem ⁓ when you're working through it, but if you do, just ignore it, it's fine. But anyway.

here is that UI. So that is pretty neat. ⁓ sometimes you want to build something that's a little bit more complicated than what you can do with ⁓ raw HTML. ⁓ And so we're using a full-on framework. This is all React, a React router, the full, like ⁓ even server rendering and all sorts of stuff. It's pretty cool. ⁓ and so your objective in this exercise is to ⁓ make it so that the view journal will actually return ⁓ that. And the tricky thing is this

Harsh Bharadwaaj (00:02:22)  
The actually the host name. We want it to be the same as ⁓ the MCP server. ⁓ We want that to work whether we're deploying to staging or running locally or it's in production or whatever. So we're gonna actually derive what ⁓ this portion of the URL should be based off of ⁓ the where the request came from. ⁓ Some of that is going to involve s working with Cloudflare and how their workers environment works and passing props through context and stuff like that.

The emoji will be there to guide you through that. But some of that might be a little bit kind of ⁓ interesting ⁓ on like what we're doing. And so that's what we're doing is just making it so that you don't have to hard code the URL for ⁓ your resource. Creating the resource itself is actually pretty straightforward. So that'll be pretty easy once you get everything ⁓ set up so that you can pipe ⁓ the ⁓ the base URL through to your tools. ⁓ So with that, you should be ready to go.

And once you're done, you should be able to use Goose or Postman or Nanobot like I have here ⁓ to say just show me your my journal and it should show you that fancy UI. None of the buttons or anything will work yet. We'll get to those later. But why don't you get ⁓ started on this and we'll see you when you're finished.

—----------------

078

Harsh Bharadwaaj (00:00:00)  
Alright, do you love making elements with ⁓ JavaScript? No, me either. ⁓ it's not the best, and there's probably easier ways or abstractions that could be built to make it a lot easier that you could write it almost like it's JSX or something. ⁓ but for now we're just raw dogging it, and that is ⁓ yeah, so you get a really good idea of what's going on ⁓ at the the base of the protocol. ⁓ So here first we're going to switch from raw HTML to remote DOM.

And then we're going to update this from the HTML string to script. And we're also going to add the framework React ⁓ right here as well. So ⁓ the reason that we choose framework React is for this. ⁓ Your agent is going to be the one in charge of deciding what each one of the element types looks like and how it's implemented. ⁓ And so ⁓ maybe your agent is built using React, maybe it's built with web components, maybe it's built with Vue or Svelte or whatever.

And so it's going to be in charge of that and some agents might actually support multiple frameworks. ⁓ it's ⁓ not entirely clear to me why it should be up to me to decide which framework is best to use ⁓ for resolving those ⁓ the ⁓ UI text and UI button, all that those things into ⁓ UI elements that get rendered to the screen. ⁓ But that option does exist and it actually is required. So ⁓ I think ⁓ if I were to guess this is actually ⁓

pretty likely to change in the future that you don't have to specify this and then the host application just decides which one. ⁓ And typically I would actually expect host applications to only support one. ⁓ This just feels a little bit like an implementation detail for me. ⁓ So anyway we're going to go into here ⁓ and we have to construct our elements. So let's start with our stack. We're going to say document create element UI stack. ⁓ And this is going to do our spacing 20 center alignment. So direction is vertical, spacing all of that.

then we're gonna create the title. We're gonna say content is ⁓ JSON stringified ⁓ tag name. So this is ultimately ⁓ going to ⁓ once this string has been constructed, it's gonna be whatever the tag name was, right? ⁓ but we are interpolating a string and so we're gonna JSON stringify that. ⁓ I should mention that ⁓ the reason for that comment is because I have a an editor extension ⁓ that

Harsh Bharadwaaj (00:02:24)  
⁓ syntax highlights this string. So it makes it like slightly better ⁓ to deal with. ⁓ But we are not ⁓ working in JavaScript. We're working in we're writing JavaScript inside of a string. ⁓ It's a little bit wonky. ⁓ But yeah so then we've got our description. ⁓ We are going to append that to the root which is just exposed as a global ⁓ for us so the root element, whatever it is, we don't get to choose and ⁓ they'll put that our stack on the root.

And then we also want to handle the case where ⁓ something does not exist. So ⁓ yep, there we go. ⁓ We'll create a ⁓ stack. ⁓ I guess that works. ⁓ however you really want to do it. This is just so that you get exposed to ⁓ these different elements. ⁓ so yeah, we'll create a stack, sure. we'll have a title, ⁓ tag not found, and then ⁓ we'll append that to the stack and then append the stack to the root. Same sort of thing. ⁓ Okay, great. So

Let's just test out. ⁓ Could you show me tag one visually? ⁓ And this should call our view tool. ⁓ And would you look at that? There it is. ⁓ And what makes this interesting ⁓ is that even though it technically is using an iframe to keep things safe, so you see all that iframe stuff right in here, ⁓ our elements are not appearing in the iframe. It uses the iframe to ⁓ run our untrusted code in an isolated environment, and then the results of our code run.

actually get copied over to the actual running environment. So there's the span and ⁓ for our title and the description. ⁓ And so what makes this really neat is that we can ⁓ adopt whatever it is that the host application does for what a UI what constitutes a UI stack or UI text. Now Nanobot hasn't actually invested a whole lot into customizing all of those things, but it ⁓ eventually it likely will. ⁓ And another interesting

thing here is I actually have in my clipboard, ⁓ and cursor knows ⁓ that I'm gonna do this, ⁓ this ⁓ UI button, which hopefully will ⁓ give you a better idea. We're doing some stuff in here that I haven't taught you yet. ⁓ that's for future steps. And so we're not gonna look too closely at what this is doing. But ⁓ yeah, with that now I can say, okay, show me tag two now.

Harsh Bharadwaaj (00:04:44)  
And here it is. Boom. We've got time spent with family, family, do-do-do. And if I click on this, ⁓ it actually does perform the tool call, and that was just deleted. ⁓ So that's actually pretty cool. This is ⁓ I I didn't decide what this button looks like or whatever. This is the value ⁓ of ⁓ remote DOM. So if I ran this same thing in a different agent, then it's going to use whatever that agent decides is a good looking button. ⁓ And so that's the value of React DOM or Remote DOM is that it allows you to just

Construct your UI. You kind of describe what your UI should look like. ⁓ It's pretty imperative, to be honest. ⁓ You can even do event handlers and stuff like that. ⁓ And then the host application will take that description and turn that into the elements for its rendering, ⁓ which I think is pretty interesting. So there you go. That's remote DOM. Honestly, not my favorite as far as the authoring experience. I think that there are abstractions that could be made to improve it.

But it does allow you to do something that you can't do with anything else, and that is creating elements that actually are rendered ⁓ right in the context of the agent in a way that is safe and secure.

—--------------

100

Harsh Bharadwaaj (00:00:00)  
So if we take a look at the tool that we've registered here, we have this output schema. We are expecting to return success Boolean ⁓ and the entry as well. ⁓ All we actually need is just the success Boolean. So let's create a schema ⁓ that expects success ⁓ of ⁓ whether or not it's a Boolean. But we actually were expecting more than just that. We ⁓ are actually expecting our response to be this entire thing.

So it's gonna have structured content, and that structured content is gonna have our success status. It will also have content, but we don't care about that. So let's take this a step further. First, I'm gonna tell this to import Z from Zod. ⁓ and we're gonna say it's a structured content, and there's our success Boolean. So we're gonna expect the response to include structured content, and that will be an object that includes success. It can include more than that, which it will, but we don't care about that. We just care about

the ⁓ structured content success. ⁓ And then with that, ⁓ we're gonna replace this with a await send mcp ⁓ message ⁓ tool. Okay, something is right. ⁓ Gotta not throw the error so we can make it here. ⁓ and our tool call. So we're gonna expect that. We're gonna get our result back. ⁓ And ⁓ I'm a little confused why we're getting an error here. So let's see.

Not ⁓ yeah, it yeah, you're right. It's not assignable to link because it's not link, it is tool. So that's kind of confusing. What's going on here? ⁓ tool. And then ⁓ we've got tool name and params as an object, and then we've got ⁓ aha schema. ⁓ There we go. Type script. ⁓ It's funny. ⁓ Okay, sweet. So we're defining our schema, delete entry schema. ⁓ And because we have that schema.

We actually are going to get type safe results here. So assuming that the result that comes through the iframe ⁓ post message process, ⁓ assuming that that is matches our schema, then we will get a result with structured content. If not, then ⁓ we're we're not. ⁓ It will actually throw an error. It will reject the promise and we'll end up showing the error boundary, which is what shows that error. ⁓ and ⁓ if it was not deleted, then we'll just say fail to delete entry. ⁓

Harsh Bharadwaaj (00:02:25)  
We could actually in our tool send back more specific error message and then parse include that in our Zod schema here and use that to display a more helpful error message here. We could we could do all that. ⁓ I'm gonna leave that as an exercise for you. ⁓ So have fun doing that. And with that now, ⁓ we should be all set. Coming back over here, let's say, please show me my journal.

And let's hope that it actually works and we don't need to restart anything.

Harsh Bharadwaaj (00:03:08)  
Okay, now I've got everything set up again, so we'll say, Please show me my journal.

And there's my journal. It's beautiful. And now I'm gonna say delete the ⁓ planning summer vacation. ⁓ And it's deleting, deleting, deleting, and boom, deleted. Isn't that cool? Yeah, that is pretty cool. And we got that because we ⁓ are are ⁓ accepting that response. ⁓ if we dive into this a little bit deeper, you might not recall, so let's just look into how all this works. So we ⁓ we send this post message for the tool call.

And then we ⁓ add an event listener for handling that message. We're gonna verify that it is this ⁓ the response for the request we made with this message ID. ⁓ and if it is, then we'll remove the event listener for ⁓ memory re reasons. We don't want a memory leak. ⁓ And then we're gonna get the response and the error from the payload. If there was an error, we'll reject and then we can show that error boundary. ⁓ if ⁓ there wasn't, then it was presumably successful. There should be a response. ⁓ if it we

weren't given a schema, then just resolve and they'll have to deal with that on their end. ⁓ but if we were given a schema, which in our case we are giving a schema, then we'll parse that. And if it w parses successfully, or it doesn't parse successfully, then we'll reject with the parse error. ⁓ And if it parses successfully, then we'll resolve with the data. And that's what keeps it type safe and and we're able to ⁓ do this logic off of that.

And you could do all kinds of things. I'm gonna again another ⁓ thing for the the viewer, but ⁓ now you can come over here and instead of telling the LLM, hey, go call this tool to get the entry and then summarize it for me, you can actually do a tool call yourself right here ⁓ to get the entry and then include that as part of your prompt. Say here is a journal entry, please summarize it for me. ⁓ so you can do all kinds of stuff with this, and I think that is pretty cool.

Harsh Bharadwaaj (00:05:09)  
So ⁓ good job on this one. Ha I hope that you had a lot of fun with it.

—-------

98

Harsh Bharadwaaj (00:00:00)  
So let's talk about some advanced use cases. ⁓ we've got two things I want to talk about. The first is the tool results flow. So you can actually get the response from a tool call, which is pretty cool. So ⁓ the user interacts with the UI, it wants to delete an entry, wants to do something, sends a tool call the MCP server, processes the request and sends that back. It's like all normal stuff. ⁓ and then the results actually come through on the post message.

and then we actually in this exercise we're gonna add a ZOD schema so that we can validate because every single one of these where the data is passed from one to the next, ⁓ you just ⁓ it's it's a network boundary. You don't really know whether the data that's coming back is what you expect, maybe it's an old version, like who knows? And so we're gonna add a ZOD schema to validate that. and then we can update the UI based on that result. So here's a a quick and simple example. ⁓ If we send a tool call with analyze code and here's the file path.

we can provide this schema to say, hey, this is the result I expect. If it's not that, then I don't know what to do with it, so throw an error or something. ⁓ with that, then we can say, hey, results structured content. I expect metric suggestions and complexity. And if the complexity is greater than 10, then I'm gonna show refactor warning or something like that. ⁓ Or update the code metrics or whatever. ⁓ And so, yeah, being able to ⁓ get the results back allows you to update the UI in some way that could be useful.

as I mentioned earlier in one of the earlier exercises, you could use this to ⁓ go and retrieve the entirety of the contents of the entry and then send that for your prompt when you're asking ⁓ to summarize an entry or something like that. And you can include all of the context and not require the LLM to ⁓ know that it needs to call the tool or whatever. So just to be a little bit more specific. ⁓ i in addition, you could actually just make a fetch call, ⁓ but authentication is

a tricky beast with that. And that is ⁓ the next thing with this render data flow. ⁓ So ⁓ when your ⁓ when you return your resource for your UI resource, you can include this UI metadata ⁓ that includes the initial render data that ⁓ could be, yeah, sure, your bed inventory or whatever. And then on the client you can receive that. And so here's what that would look like. The user requests some sort of ⁓

Harsh Bharadwaaj (00:02:20)  
resource, maybe they want to view a journal entry, for example. ⁓ the app says, okay, let's go get that. ⁓ it returns that resource. We create the iframe ⁓ for showing that journal entry or whatever it is, the the bed inventory. ⁓ and ⁓ the iframe is going to i emit that lifecycle event, hey, I'm ready. ⁓ And as a result, the app is going to send back ⁓ the render data that was sent initially as part of this ⁓ UI metadata.

And with that then we can process that data. Like we're gonna parse it again because again, we're going across boundaries, ⁓ and then we can display that content. ⁓ all the while we can show like a loading spinner, but it happens so fast that you the user won't ever see the loading spinner ⁓ because the iframe isn't even created until that render data is inside of the ⁓ host application anyway. ⁓ So it creates the iframe and then ⁓ it technically it's gonna wait for your lifecycle event, but then immediately it's gonna send you the render data.

So you shouldn't see any loading spinner in ⁓ that situation. ⁓ So there you go. That's those are the two advanced things that we're going to do in this exercise. ⁓ I think that you're going to have a good time with this one. So let's get to it.

—-

88

Harsh Bharadwaaj (00:00:00)  
So this isn't a React React workshop, and so you may have noticed a couple of Marty the Money bags in here just telling you exactly what to do, because I don't think that you really necessarily need to know React to be successful here. So we've got our use ref. We need to get a handle on this element so that we know how big it is, how much space it's going to take. Is it gonna take 800 pixels or is it gonna be a little smaller? And I deleted a couple of my journal entries so we can ⁓ see what that looks like.

so we do need to get a reference to this div, and so that's why we're using this ref prop. ⁓ and we need to use ref to ⁓ create or to get a reference to that. ⁓ And so now we're going to ⁓ grab the root. ⁓ If that root doesn't exist, then we'll just get out of here, but it should at this point. And then we're going to grab the height and the width of that element, and we're going to let our parent know, hey, this is how big I need to be to like display without.

⁓ doing any scrolling or whatever. So we're saying UI size change. ⁓ And so with that now, ⁓ with our ⁓ ref added to the div, we should be able to just say, ⁓ could you show me my journal please? ⁓ You always want to be kind and polite. ⁓ and of course we've got a bunch of ⁓ errors here. So let's try it one more time. Reload frame. There we go. Development stinks. So let's ⁓ let's do it as a fresh thing.

Could you show me my journal, please? And this time it should work out of the gate because ⁓ Vit has already optimized all the stuff and well that build stuff nonsense. But yeah, it works. It's perfect. ⁓ And so it's not too big and not too small. ⁓ So there you go. ⁓ And ⁓ of course, we do say the max height is 800 right here. And so if I were to go to localhost 7787, ⁓ this is the sidecar app that's running, that's got all the data and everything.

I'm gonna reset the database and that's going to ⁓ now we have five entries in here. ⁓ and so if I say again, can you show me my journal please? ⁓ There should be five entries and it so it should be taller and now we are going to scroll. So ⁓ the my word. What in the world? Okay, ⁓ not scroll height. This is client height.

Harsh Bharadwaaj (00:02:18)  
My bad. okay, so hopefully you didn't make that mistake, or if you did, you can ⁓ enjoy a fun laugh with me. Let's try it one last time. ⁓ Show me my journal, please. And this time it's going to work. There we go. Okay, whew\! My goodness. See, I added a console log in there. You may not have noticed that I did a little edit so I could figure out. I was about to re-record this video, but now you see me and my dub debugging. Yes, I do console log occasionally.

⁓ but yeah, so now we've we do still have it scrolling. If we have too many, we're not gonna take up the whole ⁓ area, ⁓ but ⁓ we will expand to fit ⁓ the number of entries that we've got. ⁓ Okay, ⁓ if let's just review because I was kind of jumping around a little bit there. ⁓ the the crux of all of this is that you can send a UI size change and tell ⁓ the agent, the the parent frame, ⁓ hey, I've changed my size. So

If you could like adjust accordingly, that would be great. Here's the height and width that I would like you to ⁓ set me to. ⁓ And most agents will probably just set you to ⁓ that width and width and height. ⁓ and ⁓ yeah, and then the the agent, in this case nanobop, nanobop ⁓ nanobot ⁓ will be able to ⁓ set that height and width to the height and width that we have. ⁓ So there you go. That's that's the crux of this.

The way that we did this was using ⁓ React Ref, ⁓ and ⁓ we passed that to this div, so now we get a handle on that DOM node, and then we use the client height and width, ⁓ not the scroll height. This is why it's important for you to review your AI-generated stuff. But we want the client height and width to get the actual ⁓ height and width of that, not including the scrollable area, ⁓ and then we let the parent window know that that's how big we need to be. ⁓ All right, good job on dynamic sizing.

—-------------

82

Harsh Bharadwaaj (00:00:00)  
This one has a couple of moving pieces, but it's really not all that complicated. This is the entry for our Cloudflare worker, which is running our MCP server. ⁓ And every request that comes in, we're going to pipe the ⁓ /mcp into our ⁓ MCP agent, ⁓ and ⁓ and then mcp takes over from there. But we need to get the request URL into the MCP agent so that when we come around here and we want to construct this URL.

we know ⁓ what the base url should be. We know it we need to go to slash UI slash journal viewer. ⁓ that we can find in ⁓ right here. So that's where this URL is ⁓ or what's what's gonna handle that route. ⁓ but we need to find out how to like give that base URL in the beginning. So ⁓ we're gonna start right here at this request. We have our URL right here. The URL.origin ⁓ is ⁓ that first part of our URL. So what we're gonna do ⁓

is we'll set our ⁓ props ⁓ our context props base url to that origin. ⁓ Now we're getting type errors because ⁓ it doesn't realize that we have a base url that we're expecting on our props. ⁓ And so I'm going to add mcp props right here and we're gonna bring in props as mcp props from our mcp index. So what that means then is we need to dive into ⁓ this props definition and add base url.

And with that all constructed, now types are happy, we're happy, we're getting the props into the context ⁓ of our MCP server. ⁓ And so now we can dive into the EpicMe MCP implementation. ⁓ And if we say this ⁓ this dot props, ⁓ base URL shows up right there. ⁓ All type safe and nice. So to make things a little bit easier for ourselves, we're gonna make a nice utility on top on our agent.

That checks for the existence of base URL. And if it doesn't exist, then we're gonna throw an error. ⁓ This should not happen. We shouldn't even get into here without setting the base URL. But the types say that it's possible. ⁓ Maybe the types are at fault here, but I don't mind being a little bit safe in ⁓ things. So what I'm gonna do actually is we'll grab the base URL this way. ⁓ We'll add an invariant and say base URL is not set. ⁓ This could also be

Harsh Bharadwaaj (00:02:24)  
⁓ props could be undefined, so we're gonna just say or empty object and then we'll destructure off of that and grab the base URL. ⁓ You could do that in many ways, but the basic idea is you say require base URL ⁓ and you're gonna get that base URL. And if it doesn't exist, we'll throw an error because we shouldn't that shouldn't be possible. ⁓ So now we have a mechanism ⁓ for going from this original request, ⁓ getting that base URL into our MCP server, and now our agent has a mechanism for grabbing that easily. ⁓

And so with that now, ⁓ we can get our base URL ⁓ from that and we can construct our iframe URL this way. ⁓ I actually like doing it this way. Watch this. New URL. ⁓ not quite almost there, almost there. And then base URL. So you actually ⁓ give it the path name and then ⁓ the origin or the base URL, and that will construct the URL as well. I like doing it this way.

Okay, great. So now we can create a UI resource. So create UI resource. ⁓ And we're going to specify our UI ⁓ is our ⁓ view journal. Date.now. So I ⁓ typically you want this to be a unique ID. ⁓ Date.now works just fine. ⁓ It's not like your user is going to be requesting it faster than one millisecond at a time. And it it ⁓ to me it communicates effectively. This is what it looks like right now.

⁓ there are other situations where it could make more sense, like if if you had a very specific ⁓ instance of of something and it doesn't ever change, or I I don't know. but date.now seems to work pretty well for me. So ⁓ then we'll say our content type is external URL, our iframe URL, this does need to be a string, so we're gonna say iframe URL. ⁓ URL.two string. There we go. ⁓ And our encoding is text. There we go. So

Assuming that's all working, let's just say right here, could you please show me my journal?

Harsh Bharadwaaj (00:04:27)  
⁓ And ⁓ let's see. There it is. And again, that ⁓ VT issue ⁓ during development. Hopefully by the time you go through this, this isn't a problem, but ⁓ yeah, it's a problem for me right now. ⁓ what's happening is ⁓ Vit tries to ⁓ optimize dependencies as needed, and it's super, super annoying. So here, we're just gonna reload the frame and boom, now it's gone. ⁓ So great, we've got our journal. We've

Five entries, we've got our post thing here. None of this is going to work yet, because the links are not supported yet. ⁓ viewing details, that's not gonna work. Summarize, delete, that's not gonna none of these things are gonna work yet. we'll get to those soon enough. But ⁓ that is ⁓ how I ⁓ combine a like full stack framework ⁓ into MCP. So the full stack framework can handle ⁓ these routes, ⁓ and you can point ⁓

The MCP server that's also s ⁓ handling that ⁓ full stack framework, ⁓ that same server can point to itself. and that's ⁓ what I am doing for my MCP UI right now. ⁓ if we're gonna review it really quick, because there was a little couple couple pieces here. ⁓ when a request comes in, we're gonna grab that request, get that ⁓ URL from the request. If the r URL matches slash MCP, ⁓ then we're going to ⁓ add to the context props.

the base URL, and then ⁓ have our Epicme MCP server serve ⁓ the MCP ⁓ endpoint and we'll call fetch on that. And that basically is doing ⁓ an internal ⁓ fetch request to the durable object that is managing ⁓ our instance for ⁓ our ⁓ our server. ⁓ And this context props is going to be passed along. So that's how you get the base URL into ⁓ our instance of this agent. So ⁓

Then we can access this.props with that base URL. ⁓ And then we use that to construct our iframe URL. And so now we are ⁓ able to create a resource that is to an external URL, which is technically to our own URL, ⁓ with that iframe URL ⁓ that is pointing to that. And so now you can create all sorts of routes in your application ⁓ and ⁓ use MCP UI with those routes. I think that's pretty cool.

—----------------

94

Harsh Bharadwaaj (00:00:00)  
Sweet. So let's start out in here. We no longer want to throw an error. Instead, we're going to await send MCP message. So we've made our ⁓ the link that we had earlier. Now that's a little bit more general, and you specify what type of message you want to send. We want to send a tool message. ⁓ The tool name is ViewEntry. The params are ID and the entry ID. ⁓ Now we need to make this type safe, and so let's dive into here.

And we get ⁓ this type definition on our function. We're gonna add an override. So if you've never done an override before, ⁓ your LLM should hopefully be able to help you with that. ⁓ so right up here, yep, we're gonna add that. So we've got these ⁓ MCP message types. We're gonna add a message type for a tool. We're gonna expect the tool name ⁓ and ⁓ params. And then down here in this override, we're going to say for type tool, the payload should be of ⁓ the message type tool.

And yeah, we accept all the same options and stuff. So now we can send an MCP message, whether it's a tool ⁓ or a link. Now, if you're not super familiar with how overrides work in TypeScript and all of that stuff, ⁓ it's really not that important. The most important thing for you to take away from this is that communication across this barrier between the child iframe that we're running in and the parent iframe that is the agent that is ⁓ using our MCP UI is ⁓ basically the

The same thing that we did with the link. So you have a message ID, you send the post message with the type of message you want to send. Before it was link, now it's tool. ⁓ And then you wait for a response ⁓ with the mess UI message response with that same message ID. And then you take that ⁓ response in the area, you resolve reject it, whatever. ⁓ Now we also added to this utility a schema ⁓ which we'll be using a little bit later.

So that you can parse that schema and get some type safe stuff on the response. We'll use that later, ⁓ but for now, let's just make sure that this is actually working. So show me my journal, please.

Harsh Bharadwaaj (00:02:07)  
I think the LLM is going to get tired of us asking that. ⁓ There we go. Okay, so we'll say view details, and boom, there it is. Ta-da\! Isn't that cool? ⁓ And so we've scrolled right down to the bottom ⁓ and we can view our ⁓ journal. So that is great. ⁓ And that ⁓ is ⁓ using tool calling with MCP UI. I think that's actually pretty cool.

—--------------------

95

Harsh Bharadwaaj (00:00:00)  
So we've got another one. This time I want to be able to hit summarize, and that should send a prompt to ⁓ the host application ⁓ to tell it, hey, could you please look up this journal entry and summarize it for me? ⁓ So ⁓ we can link, we can ⁓ call tools, and we can send prompts ⁓ with MCP UI. So that's your job in this one is to ⁓ take that same utility, enhance it further to be able to send a request for a prompt.

And then write out the prompt to ask it to summarize a particular journal entry. You can even in that prompt tell it to call a tool to retrieve ⁓ information about that journal entry ⁓ that we don't have in the data for this UI. So pretty cool. Go take a look.

—---------------

77

Harsh Bharadwaaj (00:00:00)  
So we've got another one. This time I want to be able to hit summarize, and that should send a prompt to ⁓ the host application ⁓ to tell it, hey, could you please look up this journal entry and summarize it for me? ⁓ So ⁓ we can link, we can ⁓ call tools, and we can send prompts ⁓ with MCP UI. So that's your job in this one is to ⁓ take that same utility, enhance it further to be able to send a request for a prompt.

And then write out the prompt to ask it to summarize a particular journal entry. You can even in that prompt tell it to call a tool to retrieve ⁓ information about that journal entry ⁓ that we don't have in the data for this UI. So pretty cool. Go take a look.

—------------

102

Harsh Bharadwaaj (00:00:00)  
Sweet. So let's start with the routes and we'll make sure that we don't accept the entry ID here. And then let's go over to our tools and we'll not send that ID anymore. Instead, our tool is going to get the entry. ⁓ come on, AI assistant. Await. There we go. ⁓ Sometimes you're like, what's going on? ⁓ little hiccup on the network or something. so we get our entry. Let's make sure that it exists. ⁓ if it doesn't, we'll throw an error and then the

⁓ LLM can be like, sorry, there's no entry there. ⁓ and then we're going to ⁓ use UI metadata to pass our initial render data. ⁓ and we're gonna pass this object. So this is effectively going from the server to the host application. It probably won't be added to the LLM context. It might, but like that's not relevant for us. So it goes to the host application.

Well, let's let's say it this way. It goes server to client to host application, and then the host application renders the iframe. The iframe says it's ready, and then it it passes this render data. ⁓ And so when we get to the iframe part of this, we need to ⁓ wait for that render data. ⁓ So the way that this is set up in the entry viewer is we have this client loader. This is going to ⁓ prevent this page from rendering anything ⁓ other than what's called this hydrate fallback. And so this is gonna be our loading spinner.

And it's gonna show this while we wait for ⁓ our render data to become available. ⁓ So here and once it is, then we have our loader data, gets our entry, and and then we're all good. We can ⁓ show the entry title and all that stuff. ⁓ So ⁓ with that now we need to wait for the render data. So let's get rid of that, delete all this, get rid of that. So the last thing we need to do is actually implement wait for render data. How do we wait for the render data?

Well, let's go into here, and it's actually gonna be pretty similar to send MCP message. ⁓ just a little bit different. So we're gonna export, ⁓ wait for render data. ⁓ And yeah, sure, let's just let the LLM ⁓ do what it thinks it knows it needs to do. my goodness. Okay. Hopefully this is right. ⁓ So wait for render data. We're we've made this a generic, and that is going to be inferred ⁓ based off of the schema that we're given. ⁓ So we're given ⁓

Harsh Bharadwaaj (00:02:18)  
This, where is that? we're given this schema right here. This is what we're expecting to get. If we don't get this, then we can't render this UI. Like we don't you sent us something, we don't know what to do with it. and so we're gonna be really strict and say, I expect an entry, and it should have an ID and a title and its content, all this stuff. ⁓ So ⁓ and that is what we're sending right here. We're getting that entry, we're sending that along. So we're just validating that we got what we thought we were gonna get because we're going across a network boundary and then the iframe boundary.

so there's a lot of passing of things going on here. So we're gonna validate that with that schema. ⁓ we're going to say, hey, this wait for render data is gonna return a promise that it is of the type render data. ⁓ So this this stuff is like some a little bit more advanced TypeScript stuff. If you didn't quite get this, don't feel ⁓ bad about it. It's not something that you do a whole lot anyway. ⁓ and again, like this will probably be a built-in utility into MCP UI in the future, because it's pretty useful.

So then we're gonna add ⁓ a message handler. So I I can already tell we're missing one one thing here. I'm gonna bring back the comments here and that might help. ⁓ So yeah, we're gonna add a message handler. So when the UI lifecycle iframe render data ⁓ event shows up, ⁓ then we remove the event listener and we can get the render data. We're going to reject if there was an error, parse the data based off of the scheme that we were given, and if it was successful ⁓ or if it wasn't successful, we'll reject. Otherwise, we'll resolve with that data. ⁓ So the part we're missing is

Is telling the parent frame, hey, I'm ready to ⁓ receive this event. ⁓ now we technically already ⁓ well no, this this won't work ⁓ because ⁓ I was going to say we we technically ⁓ send the event, the lifecycle event right here. And so this says, hey, I'm ready. ⁓ But the way that we have this structured is this won't render until this resolves. And so we need to actually send that event.

At this point to trigger ⁓ this ⁓ render data to be received. ⁓ So we're gonna add a window ⁓ window dot post message UI lifecycle iframe ready. So we're like, hey, we're ready to receive the render data and any other events that you want to send to us. ⁓ And then we add this event listener. And with that, it should work. I hope it works, because this would be annoying to have to re-record for you all, but I I would.

Harsh Bharadwaaj (00:04:44)  
If if I had to, I would. It would be a pleasure. Okay. So let's see if this works. Can you please show me my journal? ⁓ and it blew up. So give me just one moment to get everything set up again. Okay, so we restarted the app, we restarted ⁓ Nanobot, and now let's try it again. Can you please show me my journal?

And here we go, journal call, did to do. Of course, this is the first run, so V is like, I don't know what to do. So we're gonna reload the frame. ⁓ and here we go. Now I should be able to click view details. ⁓ And we're waiting for journal entries, but we never get the journal entries. ⁓ So for your benefit, we ⁓ I'm not gonna stop the recording and re-record. Instead, we're going to see if we can debug this together. ⁓ So let's do that. ⁓ right here. Actually, first let's reload the frame.

Yeah, okay, so that's definitely not working. ⁓ let's add a debugger ⁓ right here. And then we're gonna reload ⁓ the frame again. Okay, so we are ⁓ we're at least verified that we are changing the code we think we're changing, which is always a good thing to verify. ⁓ and then we're gonna come in here. I'm pretty sure we're not gonna get even this far. ⁓ okay, so we do get an event. ⁓

UI iframe ready. That's interesting. Yeah, okay. ⁓ haha, I know what it is. Window post message. I need to post this to the parent. ⁓ Goodness gracious. ⁓ yep. Parent. ⁓ So we were just sending ourselves a message. Haha. ⁓ No, we want to send the parent a message. So let's try this once more. ⁓ if I reload, I bet it'll work actually. Yep, there it is. So let's just make total sure if I go to ⁓

weekend hike with family, click that, boom, there it is. ⁓ and I can view any one of these, and each one of these is going to work. it's going to send the data through. ⁓ If I ⁓ here, let's just double double double check that the iframe URL does not include ⁓ that ID. So right here, entry viewer, there's no ID. So the only way that it could get ⁓ the data ⁓ is ⁓ through render data. ⁓ So that's neat.

Harsh Bharadwaaj (00:07:04)  
Congratulations. ⁓ You have now successfully waited for render data, passed some ⁓ data which could be privileged. And here's the other cool thing: is that you could build a UI that accepts data that you don't control. Maybe the LLM generates something and then passes that to you. ⁓ I think that would be kind of interesting. So ⁓ yeah, lots of really cool things that you can do with this. ⁓ And that is MCP UI ⁓ accessing potentially private data. Great work on this.

—----------------

076

Harsh Bharadwaaj (00:00:00)  
Being able to create UI that appears in the context of the chat or in some sort of sidecar or something, that's really cool. It's really awesome to be able to control that entire experience. But what if you were to be tasked with, like, hey, I want a stopwatch? Do you really need to be able to control that entire experience? Or do you want it to feel a lot more integrated into the agent conversation that your user is having? ⁓ There are good arguments for some situations where being totally integrated makes sense.

The problem is that it's not secure for an agent to allow you to just ⁓ use all of its components and render like ⁓ run code inside of the agent, ⁓ because then you could access other things. ⁓ And so ⁓ for that reason, when you render some HTML in the agent, you're ⁓ rendering inside of an iframe to isolate you from the rest of the environment. The problem is when you're running running inside that environment.

That isolated environment in the iframe, you don't have access to ⁓ the different components from the agent. So you can't say, hey, can you take your button and stick it in here and take your text content and stick it in here, whatever, ⁓ to make it visually consistent with the rest of the UI. ⁓ So there's a project from Shopify called Remote DOM that kind of solves this. It still uses iframes to isolate you, ⁓ but it takes the result of your running code and sticks that into

The ⁓ DOM ⁓ using the components from the host application. It does this in a secure way that allows you to even add interactivity to those elements. And so when you create that ⁓ stopwatch, you can actually use the buttons from the host environment, which is really cool. And MCP UI supports this. So here, if I say, could you please show me tag one visually? ⁓ And here, now it's going to show me tag one. And would you look at that?

I have a button in here. And this is the native button native to ⁓ this agent. If I were to run run this in another agent, ⁓ some other application, ⁓ that button would look differently. ⁓ and it even supports indication or or ⁓ tool calling and all of that stuff. We're gonna cover tool calling from MCP UI a little bit later, but I wanted to show you that ⁓ you can actually have

Harsh Bharadwaaj (00:02:14)  
An integrated experience. And of course, like this doesn't look all that that great right here. ⁓ but we're going to help ⁓ give you an idea of how all of this is put together. ⁓ It's not my favorite thing to author, ⁓ it is ⁓ very imperative. ⁓ it's very similar to constructing ⁓ document or elements with this document create element and all of this stuff. You've got different attributes.

⁓ we are using ⁓ UI stack and UI text and UI button. All this stuff is going to be very specific to ⁓ the remote DOM implementation of MCP UI. So you're gonna want to ⁓ look at the docs if you ever are are actually going into building ⁓ this sort of an experience. ⁓ MCP UI is managing all of this stuff, and you can take a look at the docs ⁓ about ⁓ remote DOM, remote DOM right here, and it'll show you how to create these.

different scripts and hopefully the authoring experience will get better because this is just awful. ⁓ I do have ⁓ a an extension in my editor that allows me to add a comment like this. It doesn't work in Markdown here, but ⁓ adding that comment will ⁓ add a syntax highlighting to all of the code here. And so if you want to, ⁓ I'll put a link to that extension ⁓ if you're using a VS Code ⁓ editor or similar

because it is at least a little bit nicer, even though it is you're still like authoring ⁓ JavaScript in a string, which is not fun. ⁓ but yeah, this is ⁓ pretty cool. it does allow you to basically use the elements from the agent. You don't have to worry about styling it, and you can make sure that it stays consistent wherever it's rendered. R Remote DOM is pretty cool, and so we're gonna get you into this one. Have a good time.

—------------------

092

Harsh Bharadwaaj (00:00:00)  
So we're going to build this send link mcp message ⁓ utility. So let's go ahead and import that now. And then we're gonna replace this throw error with a send link mcp message. Yeah, I don't know what this is talking about. Yeah, we want to do send link mcp message to that URL. So basically, ⁓ this is the URL we wanna go to, x.com intent post, and it's gonna say, I have this journal entry, and it's so exciting, whatever. ⁓ but we need to ⁓

Ascend that we're actually we want to await that because it's gonna be asynchronous. We're gonna ⁓ make the request and then we're gonna wait for the ⁓ our parent, the host application, to respond ⁓ with an event that says, yeah, this ⁓ was successful or whatever. ⁓ So coming over to our MCP utilities, we're gonna make a function ⁓ that handles this for us. So send link mcp ⁓ message. This is going to accept a URL.

And then we're gonna return a promise. So that way we can await that and we'll know whether or not it was successful. ⁓ we are going to create a random UUID. ⁓ This is how ⁓ we communicate across this boundary between the child and the parent. We're gonna say, I'm making this request, and then when the parent's ready with a response, it'll send a response ⁓ with that same ID. So we can associate those. ⁓ That way, if you were to like send a whole bunch of requests at once and then get those responses and

You don't really know whether like what order they're gonna come in. They might not come in the right order. And so having an ID will make sure that we associate the right request with the right response. So we're gonna make that ID. We're gonna say, hey, window parent post message. The type of message we want to send is a link. We're trying to navigate to a link and the client is gonna decide what to do with that, ⁓ or the that host application. ⁓ and then we're gonna send that message ID so it can associate that.

We're going to send the URL as the payload, and we're going to ⁓ send this to whoever our parent is. We don't want to limit it because we don't ⁓ necessarily know if we're like building a general use thing. We just send it to whoever the parent is. That's what that star is all about. ⁓ And then we're going to add an event listener for that message, ⁓ for a message from our parent. ⁓ And if the event data type is UI message response.

Harsh Bharadwaaj (00:02:18)  
And the data message ID equals our message, then this is a response to our ⁓ request. And then we can resolve this. Now we can make this a little bit better. We want to handle rejecting and stuff like that. So let's ⁓ and also we want to clean up after ourselves because we're responsible like that. And so let's extract this ⁓ here and we'll call this ⁓ handle event. And I just can't I can't not do this. So we're gonna ⁓ convert this to a name function. I don't know. I'm I'm weird that way.

I also don't like return types if I can help it. So there you go. ⁓ And so we add that, we restructured it this way because ⁓ if we end up in this situation, we want to remove that event listener. So say window.remove event listener. I'm kind of hoping for my AI assistant to finish that. There we go. ⁓ Message handle event. So now we're no longer listening for those events. ⁓ And

Actually, I like to write this a little differently. Don't don't hate me. ⁓ but I'm gonna say if the type is not, then return. ⁓ And if the message is not, then return. That way we don't have to worry about ⁓ doing everything inside of these c ⁓ additional like it just it makes your code go like in this direction. I like that. Just exit early and then the rest of it we can assume everything is what we want it to be. ⁓ okay, so now we can get our error and response.

⁓ from the payload. And if there is an error here, actually, either way, we're going to remove the listener, so let's get rid of that. ⁓ If there's an error, then we'll reject with that error. If ⁓ otherwise, we'll resolve with the response. Now, ⁓ presumably, ⁓ this should be fine. But what if there's no error and there's no response? ⁓ that would be unfortunate. And so we're going to ⁓ remove this if statement will just say ⁓ resolve with the response. So if there's an error, reject with the error. Otherwise,

We'll just resolve with the response and maybe the response is undefined. I don't know, but we're at least gonna resolve at some point. So there we go. That's how I write ⁓ this code. Now of course there's more that you could do in here. You could do ⁓ adding timeouts and cancellation and stuff like that, but this is gonna handle us pretty well, I think. ⁓ So with that, I actually think that I should be getting hot module replacement and hot reload here. So I should be able to just click this and it'll work. Boom\! ⁓ Let's go. It totally worked, so that's nice.

Harsh Bharadwaaj (00:04:43)  
so I've got five journal entries in my Epic Me journal. Should I post that? ⁓ I'm gonna post it and all of you go look for that ⁓ post one day and go favorite it later. ⁓ because yeah, that is great. ⁓ okay, great. So ⁓ let's just review what's going on here. First, ⁓ back here, where like the easy easy stuff where we're using it, we want to s ⁓ when the ⁓ post on X ⁓ button is clicked.

We're going to handle that and say send link MCP message ⁓ with this URL. So we effectively want to ⁓ tell our parent, hey, could you direct us over there? And I th I think I didn't no don't know that I made this clear enough, but you can't do this ⁓ any other way. Like you might think, why don't we just use like a regular link? Maybe some of you asked that question. If you do that, then it's actually gonna change your ⁓ iframe to that URL, and then you kind of lose all control.

And users don't have back buttons in here. You can right-click and hit back. ⁓ you know what? You can't. If you hit back right here, it's gonna hit back in the parent. So ⁓ yeah, there's not like really an easy way to do that. And so ⁓ this is why we're communicating across the iframe, even for a link. ⁓ Okay, so we send the MCP message, we're gonna create an ID so we can associate the request and the response. We're going to send the request, we're going to listen for the response, we're going to verify that the events we're getting are ⁓

response because you can get ⁓ message events ⁓ like all kinds of different types of message events from the parents and so we're gonna filter it down make sure that this is a UI message response ⁓ and that the message ID matches the request ⁓ that we made and if that's the case then we're gonna remove the event listener so we clean up after ourselves we avoid ⁓ memory leaks for doing that ⁓ and then we're gonna grab the error and the response and we're gonna reject if there's an error and resolve that there is a response. ⁓ And that worked. So

Good job on this one. And now we can actually build on top of this to add support for tool calling and prompts. ⁓ And ⁓ I feel like there was another one, but yeah, like all of the asynchronous stuff in here. ⁓ So great job.

—------------------

087

Harsh Bharadwaaj (00:00:00)  
So this is really helpful, but we're kind of just guessing ⁓ on that preferred size. And it would be nice if we had the ability to say, hey, you know what? I want to set the size after I've rendered, once I know how big I'm supposed to be. Maybe I don't want it to be ⁓ really, really tall if I only have a couple entries. Or maybe it's like it's not gonna take up the whole space, but like just a little bit more, and then I won't have to scroll. So ⁓ that's what we're gonna be doing in this exercise is making it so that you can actually communicate from

the child iframe from your UI to the parent iframe and say, hey, listen, this is my size. Like could you set it to be that size? ⁓ So that is your objective in this exercise. Have a good time.

—----  
85

Harsh Bharadwaaj (00:00:00)  
Have you noticed how this ⁓ doesn't look great? ⁓ Like there's a lot of content in here. And here's the problem that ⁓ you typically run into with iframes: the parent doesn't actually know how big to make this UI view for the child. And so they kind of like do their best guess. And so what we can do is actually help a little bit ⁓ in the definition of the resource. We can just say, hey, here's the external URL.

And actually, by the way, like you might want to size it about this because this is how big ⁓ I'm expecting ⁓ this component to be. Now the the agent might of course say, well, I'm too skinny to fit in that. And so like you can't always 100% ⁓ depend on ⁓ the fact that you're telling them how big you want it to be, but like with pretty good confidence, you can at least give it kind of a hint and say what your preferred frame size would be. And so that's what you're gonna do in this exercise. It's really quick.

And when you're all finished, this should be sized appropriately.

—---------------

97

Harsh Bharadwaaj (00:00:00)  
I used to be a banker, then I lost interest. Ha ha ha. ⁓ okay, that's a pretty good one. ⁓ great. So now is the time for you to take a break. Write down what you learned so you remember it better. ⁓ go like take a nap or something. ⁓ Naps are good for you. ⁓ do whatever you need to do to get blood flowing in your brain again. ⁓ and make sure you retain everything that you're learning. ⁓ And then ⁓ when you're ready to learn something new, I'll be here ready for you.

—--------

74

Harsh Bharadwaaj (00:00:00)  
Let's start out by bringing in create UI resource. ⁓ And then we'll come on down to our view tag tool. And first we'll grab the tag from the database. And then we'll create some HTML. So if there is a tag, then we'll have our div with the H1 and the and the P with the description. ⁓ Otherwise, we'll say tag not found. ⁓ And then we'll use create UI resource ⁓ to ⁓ create our ⁓ UI ⁓ resource.

we need to start with UI colon slash slash, and then we'll say view tag slash id. ⁓ and then we can set our ⁓ content. Here we go. Whoops. ⁓ our content needs to be an object. ⁓ And we can do a couple of things in here. ⁓ we want to do a raw HTML string. So we're going to say HTML string is our HTML. ⁓ our type ⁓ is ⁓ raw HTML. ⁓ And then we have our encoding ⁓ is just text.

And that is enough to create our ⁓ HTML that we're looking for, that HTML resource. ⁓ So if we reconnect and go to our tools, we can go to view tag ⁓ and run the tool, ⁓ and there it is. And of course, we head over to ⁓ some client that supports MCP UI and it will render that as we saw earlier. ⁓ So that is how you create your first UI resource with a basic HTML. ⁓ It's pretty basic. You can extend this and you know add a

even styles and all of the stuff you would expect ⁓ for something like this. Feel free to to go nuts on how you do this sort of thing. But if you're gonna go a little too nuts, then maybe keep on going through the workshop because we have better ways to handle more complex scenarios. And we can get really interactive here. So ⁓ look forward to that. But right now, you've just rendered some UI right in the conversation ⁓ with an agent. And I think that's pretty powerful. Good job.

—------------

71

Harsh Bharadwaaj (00:00:00)  
Hey, welcome. My name is Kent C. Dodds, and I am so excited that you're joining me on this journey to learn MCP UI. ⁓ So MCP UI is really exciting because this is where we ⁓ get to merge the UI stuff that we've been doing for all this time with the natural language stuff that we're ⁓ experiencing ⁓ with ⁓ agents and that sort of thing. So ⁓ when when I started getting into MCP, I was kind of like, man, all this UI stuff I've been learning over the last decade and building UIs.

It's kind of going out the window now. ⁓ because ⁓ like we're just gonna use natural language for everything. But the more I used it, the more I realized, you know what, there are some cases where UI is just better. Like if I say, hey, I want directions to the nearest taco stand, ⁓ I don't want it to like say, okay, turn left here and then do that and whatever. Like I want to see it. I want to see a map. ⁓ Or if I say, hey, I need a a you know, a lift, I want to see a map of the the ⁓ car coming to me. ⁓ or if I say, hey, I need a stopwatch.

And I I don't want it to have to type start and stop or say start and stop. ⁓ I want to be able to click the start and stop buttons. ⁓ And I would like that to all feel like a really cohesive experience as part of ⁓ the agent that I'm communicating with. And so being able to ⁓ have an MCP server that can serve UI, ⁓ and not just UI that is specific to it, but actually we're going to talk about ⁓ having an MCP server that takes the UI elements.

From the agent and pieces those together, puts those where they're supposed to go ⁓ with remote DOM. It's very interesting. ⁓ But all of this ⁓ new user interaction stuff is just fascinating to me. And I think that we're going to see a lot of new patterns emerge over the next ⁓ several years as we're learning about and using MCP UI. Now, at the time of this recording, there are only a handful of agents that actually support MCP UI.

⁓ but MCP at the also at the time of this recording is not even a year old yet. So it's amazing how far it is already. So it's gonna be a little bit rough as you get into this because it's so new, ⁓ but this is your absolute best place to learn about MCP UI, and you can jump in because we're so early, you can make a huge impact on the future of user interaction ⁓ by participating in MCP UI and being one of the people on the planet who knows it best. ⁓ So I'm excited to take you through this.

Harsh Bharadwaaj (00:02:22)  
There's going to be a ton of stuff. We're going to start with really, really simple stuff, and then we'll get to ⁓ progressively more advanced things. ⁓ We're going to be using React and React Router, but don't worry if you aren't familiar with those things. ⁓ I really try to focus on the MCP part of ⁓ all of that stuff and kind of guide you through the React and React Router stuff. ⁓ And then you can of course apply what you learn to whatever framework you're familiar with using as far as UI is concerned.

So anyway, I'm really excited about this. I'm super jazzed that you've joined me, and I hope you enjoy this as well. See you in the workshop.

—-----------

083

Harsh Bharadwaaj (00:00:00)  
This one's pretty quick, but I wanted to make sure to have you do this one, even though there's not actually gonna be a visual ⁓ change when you do this, because it's important to communicate from your ⁓ mini application that's inside this iframe to the parent that you're ready to receive events and all of that stuff. Otherwise the host application is gonna be waiting forever. ⁓ And so you're not actually going to see a visual change here with this one at all. ⁓ but really all

This is just establishing how to communicate from the child to the parent. And we're going to be doing a lot of that in ⁓ the coming exercises. So like I said, you won't notice anything different here, ⁓ but it is important for you to do. So ⁓ let's get to it. It will be pretty quick. The emoji will tell you exactly ⁓ how to send that post message and what to send. ⁓ And then we'll see you when you're done.

—----------

90

Harsh Bharadwaaj (00:00:00)  
So now we're going to make this interactive. This is one of the reasons I wanted to get us out of a string of HTML because there's actually some logic that needs to go around to make things really nice and interactive. And while you can technically do a string of HTML and include a script tag and all of that stuff, ⁓ using a full framework that ⁓ has like state capabilities and stuff like that just makes this a lot, lot easier. ⁓ and at least manageable. So ⁓ we are going to be building a utility in this exercise.

⁓ for ⁓ being able to click on links and have those open up. So we're gonna start with a send link MCP message where you can pass a URL and it will go open that up. ⁓ And then we're also we're gonna evolve that into a send MCP message, more generic ⁓ utility that can send a tool and prompt ⁓ things. And so here ⁓ we're going to be able to click on post and that'll open up a post here on X. I've got five journal entries. Hooray\!

⁓ And you'll be able to click on View Details and it will show you your details ⁓ right here. And you'll also ⁓ click on the summarize and you'll get the ⁓ LLM to summarize your journal entries ⁓ using the prompt that's generated for you. ⁓ So those three things are what we're gonna we're going to be doing in this exercise. And here to kind of explore how all of that goes together ⁓ and like how that's all structured.

The user is going to click on the button ⁓ and the iframe will generate a message ID and send a post message to the host application. That's the the parent application, the parent iframe or the parent frame. So you're the iframe, you're the child, and we're gonna send it to the parent. The parent is going to handle that request. So in this case, we're going to ⁓ click on a link and it's gonna open up a page. ⁓ So it's gonna handle the navigation request and maybe open it up in a new tab or whatever.

And then send a response to say, hey, I opened it up, congratulations, and then you can handle that and show that in your UI. Like, good job, you clicked on the link or whatever. ⁓ The user might click on generate haiku, and that would be a tool call. So it's gonna send it's kind of goes through the same thing. It actually still generates a message ID here as well, sends a tool message via the post message, executes the MCP tool, ⁓ and or it sends that to the client. The client will send that to server, server will send the response.

Harsh Bharadwaaj (00:02:23)  
Client sends that to the host, and then the host sends that back to the iframe ⁓ in the UI message response ⁓ with and then the UI can update according to that. And then it's the same sort of thing with a prompt. ⁓ The user clicks on some button or there's some interaction and the iframe generates a an ID, ⁓ sends the prompt to the host, and then the host puts that into its LLM and the LLM does whatever it wants to with it. ⁓ So that is what we're going to be doing in this exercise to make everything interactive.

There's gonna be a little bit of TypeScript stuff going on in here. If you're not very comfortable with TypeScript, hopefully your LLM can kind of help guide you. We're gonna be doing overrides on ⁓ the this send MCP message so that depending on ⁓ what value you put in here, whether it's prompt or tool or ⁓ link, ⁓ it will determine what ⁓ possible values can go as the second argument. It's pretty cool and ⁓ I wouldn't say it's advanced TypeScript, but it's definitely not something that I do every day.

And so if that's unfamiliar to you, just prepare yourself emotionally for a little bit of struggle on that. But hopefully the emoji are there to guide you and help you be successful. And then of course you've got your ⁓ AI assistant in your coding editor. So I hope that helps. And I hope you have a really good time with making our UI interactive. Have fun.

—-----

104

Harsh Bharadwaaj (00:00:00)  
Hey, welcome. My name is Kent C. Dodds, and I am so excited that you're joining me on this journey to learn MCP UI. ⁓ So MCP UI is really exciting because this is where we ⁓ get to merge the UI stuff that we've been doing for all this time with the natural language stuff that we're ⁓ experiencing ⁓ with ⁓ agents and that sort of thing. So ⁓ when when I started getting into MCP, I was kind of like, man, all this UI stuff I've been learning over the last decade and building UIs.

It's kind of going out the window now. ⁓ because ⁓ like we're just gonna use natural language for everything. But the more I used it, the more I realized, you know what, there are some cases where UI is just better. Like if I say, hey, I want directions to the nearest taco stand, ⁓ I don't want it to like say, okay, turn left here and then do that and whatever. Like I want to see it. I want to see a map. ⁓ Or if I say, hey, I need a a you know, a lift, I want to see a map of the the ⁓ car coming to me. ⁓ or if I say, hey, I need a stopwatch.

And I I don't want it to have to type start and stop or say start and stop. ⁓ I want to be able to click the start and stop buttons. ⁓ And I would like that to all feel like a really cohesive experience as part of ⁓ the agent that I'm communicating with. And so being able to ⁓ have an MCP server that can serve UI, ⁓ and not just UI that is specific to it, but actually we're going to talk about ⁓ having an MCP server that takes the UI elements.

From the agent and pieces those together, puts those where they're supposed to go ⁓ with remote DOM. It's very interesting. ⁓ But all of this ⁓ new user interaction stuff is just fascinating to me. And I think that we're going to see a lot of new patterns emerge over the next ⁓ several years as we're learning about and using MCP UI. Now, at the time of this recording, there are only a handful of agents that actually support MCP UI.

⁓ but MCP at the also at the time of this recording is not even a year old yet. So it's amazing how far it is already. So it's gonna be a little bit rough as you get into this because it's so new, ⁓ but this is your absolute best place to learn about MCP UI, and you can jump in because we're so early, you can make a huge impact on the future of user interaction ⁓ by participating in MCP UI and being one of the people on the planet who knows it best. ⁓ So I'm excited to take you through this.

Harsh Bharadwaaj (00:02:22)  
There's going to be a ton of stuff. We're going to start with really, really simple stuff, and then we'll get to ⁓ progressively more advanced things. ⁓ We're going to be using React and React Router, but don't worry if you aren't familiar with those things. ⁓ I really try to focus on the MCP part of ⁓ all of that stuff and kind of guide you through the React and React Router stuff. ⁓ And then you can of course apply what you learn to whatever framework you're familiar with using as far as UI is concerned.

So anyway, I'm really excited about this. I'm super jazzed that you've joined me, and I hope you enjoy this as well. See you in the workshop.

—---------

80

Harsh Bharadwaaj (00:00:00)  
Okay, now we're gonna get to a little bit more complicated and ⁓ useful situations here. So I wanna take you on a flow of how we're going to do things now. We don't we don't really want to just do ⁓ an inline string ⁓ of text, right? That like isn't the best. And it's sticking it inside of an iframe anyway. So what if we could just instead give it an external URL and point to that iframe URL? I feel like that would be a little bit better. We can even control the frame size and stuff. We're gonna get into that in this exercise.

So let's look at kind of the architecture of how this all works. The user's gonna request something like a dashboard. The host will call, ⁓ tell the client that it wants to call the get dashboard tool, the server will send that. ⁓ it's going to create the iframe URL. We're gonna actually construct that and I'll show you how to get ⁓ the ⁓ request URL so we know what domain we're on and so we can serve the right domain and everything ⁓ for that iframe URL. Because we don't want to hard code our domain right here.

Like we've got like what if it's you know a pro or ⁓ staging ⁓ or localhost? Like we don't want to hard code that. So I'm I'm gonna show you how to find ⁓ the ⁓ the UR or the origin of ⁓ the server that's receiving the request and ⁓ send that through our CloudFlyer stuff so that we can get it into our tool so we can construct this URL based on that. ⁓ and then ⁓ we're going to s package that up as a UI resource.

We're going to return that UI resource to the host, and then the host will create an iframe element using MCP UI on there ⁓ or following the spec for MCP UI. ⁓ And that's going to load the iframe. ⁓ Our iframe is gonna ⁓ send a special event called the UI lifecycle iframe ready to let it know, hey, I'm ready to start receiving ⁓ data or receiving other events, communicating ⁓ between the agent ⁓ and the UI. That's that's one of the really cool things, is that like

Your agent can actually talk back to the UI. The UI can talk back to the agent. ⁓ it's yeah, pretty grad. We're gonna be doing a little bit of that today. ⁓ and then we're gonna render the dashboard with data or whatever. ⁓ we're gonna also handle sizing. And so we've got like our preferred frame size, we'll look at that. ⁓ but you also have the ability to ⁓ dynamically change the size, and so we can send ⁓ those types of events. The user could interact with the data dashboard and then ⁓

Harsh Bharadwaaj (00:02:23)  
That dashboard could send ⁓ requests for tool calls or prompts or even open links, ⁓ stuff like that, which we'll look at in the next exercise. ⁓ the ⁓ host will then execute that requested action and forward that onto the server. If it's like a tool call or something, return the results and that goes back to the host and then send the result back to iframe. ⁓ So w ⁓ lots of this stuff we're gonna get into in the next exercise. ⁓ it's just really from this line up that we're doing in this exercise.

So that is that. Now, one thing I want to talk about before I send you off into this exercise is we're actually changing things considerably because before we just had a really simple ⁓ Cloudflare hosted or Cloudflare setup MCP server. ⁓ But now I want to have like a full-on UI with a UI framework, and I can use a React and React router and server rendering and all of that stuff. ⁓ And that all actually works in Cloudflare, and you can run it right alongside.

your MCP server ⁓ in that same worker. ⁓ And so I've done all of that work for you because this is not a full stack React workshop. ⁓ And so ⁓ you may be doing things differently at your work or whatever. ⁓ But I thought you would be interested in seeing what changes were necessary to make all of this work. And so we're actually just gonna look at the diff together. So we're not gonna go too deep into the specifics of the React router setup. That's all the stuff inside of app

we've got routes configured in here. You're actually going to be working in here in a future exercise. This is the journal viewer. This is what you're going to be working in for some of our ⁓ steps in this exercise. ⁓ we've got a couple other routes defined in here in the server entry and stuff like that. ⁓ not super critical for you to understand all of that stuff. Also the React Router config and TS configs were ⁓ adjusted and changed. So we have ⁓ TS config for the Cloudflare environment and TS config for the node environment.

⁓ that's actually mostly for development like vtest and the react router config and stuff or or your vtconfig and stuff. ⁓ then we also have some types that are defined here. Again, mostly for the React router integration. ⁓ you don't really need to worry yourself about this stuff unless you happen to be building a React router application on Cloudflare workers, then all of this will be very relevant and you can use it as an example. ⁓ But for the rest of you, ⁓ which I expect is probably the majority of you.

Harsh Bharadwaaj (00:04:51)  
⁓ all of this is going to need to be adapted to how for you're delivering your MCP server. So I just wanted to show you how we're doing it here, ⁓ just briefly, but we're not going to do any exercises on this stuff because most of you, it won't be relevant. ⁓ Okay, yeah, we've got our vconfig. That also is a thing. ⁓ Most of our MCP stuff actually doesn't change a whole lot. So here are ⁓ worker, most of the changes in here are just setting us up for the first exercise.

I believe that's actually the case for all of our stuff inside of ⁓ worker. Yep, more first exercise stuff. ⁓ yeah, we have a new tool, ⁓ the View Journal tool. So that's some more first exercise stuff. ⁓ I actually switched us from remote DOM back to raw HTML. ⁓ that's kind of irrelevant at this point. ⁓ and then I yeah, deleted the tag remote DOM UI script because man, this scares me. ⁓ just having JS in that string like that. ⁓

⁓ so yeah, I I I moved it, I I guess I should say. ⁓ So ⁓ that ⁓ is ⁓ that's that. It's pretty like there's it's pretty light-handed in how all of this works. ⁓ I I kind of ski skimmed over this, but we do have this create request handler that's coming from React Router and it's got this virtual module and stuff. ⁓ Again, all very specific to the way that React Router works and not very relevant to you if that's not what you're doing. ⁓ So ⁓

For the most part, ⁓ it's all pretty much the same, just with the addition of React Rider. All the MCP stuff is unchanged. ⁓ So hopefully that helps ⁓ orient you as we're bringing in something like pretty significant and new here. If you're not into React, you don't use React at all, don't worry. All of this, the same concepts and things transfer just as well as ⁓ any other framework that you're using, even if it's not JavaScript, ⁓ whatever you use to generate HTML, these concepts will transfer there too.

So I hope that is helpful to you, and I hope you're excited to learn all about building ⁓ even more complex UIs with MCP UI. Okay, let's get into it.

—------------

96

Harsh Bharadwaaj (00:00:00)  
So this is the summarize entry button. ⁓ And right here we need to replace that throw with a wait send MCP message prompt. And here's our prompt. Please use the Epicme get journal entry tool, or get entry tool, ⁓ to get entry, entry ID, and provide a concise and insightful summary of it. Now, I'm gonna say this right now because some of you are asking: hey, I feel like I remember when we talked about prompts in MCP that it's actually better to

⁓ retrieve all of the information and include all of that information in the prompt. ⁓ And yes ⁓ and no ⁓ for this case. So ⁓ what you're going to see here in a second when we do this is the user will actually, and I think that this is typical of most agents, they'll actually see the prompt that you send. ⁓ And ⁓ if you include all of that information, it just kind of feels like a lot. So that's one reason.

⁓ and the other reason is we haven't yet implemented the ability to call tools and get the response yet. That'll be in the next exercise. So once we do that, you can come back here and add that capability to ⁓ retrieve the entry and then you can just have the prompt include all of the context so the LLM doesn't have to go and call the get entry tool. ⁓ and so you could say, ⁓ please summarize this journal entry and then just include the whole journal entry rather than being specific and telling it to go call a tool to get more context.

Okay, great. So with that then, ⁓ we're gonna dive into here and add prompt support. So ⁓ the prompt message message type should accept a prompt ⁓ input. So the prompt message type accepts a prompt input. That's what that is. ⁓ And then we'll come down here, we'll add another override ⁓ for send mcp message that says if the type is prompt, then the payload needs to ⁓ have that.

⁓ that object shape. ⁓ And then our ⁓ thing here will just handle that just beautifully. ⁓ So let's try this again. Please show me my journal.

Harsh Bharadwaaj (00:02:04)  
Okay, great, there we go. And now let's summarize. ⁓ First day, yada yada. Please use the EpicMe get entry tool, get entry one, and provide a concise summary. So it went and got the entry and then it summarized. Ta-da\! So there you go. So ⁓ now you can do tool calls, you can link, and you can send prompts to the agent. ⁓ the world is your oyster. This is pretty neat. And all it takes is this ⁓ one utility for managing that communication.

I believe that in the future ⁓ MCP UI will probably ship with some of these utilities. I think that they're pretty useful. ⁓ but now you know exactly how to ⁓ do that yourself. So ⁓ you're really smart. ⁓ Okay, good job with this one, ⁓ and ⁓ I hope that you ⁓ use this to great effect.

—----------

75

Harsh Bharadwaaj (00:00:00)  
Hey, good job on that one. It is now time to take a break so that you can just write down what you learned and make sure that ⁓ you're ⁓ giving them your body the rest and stuff that it needs. So here's a joke for you to give you a little chuckle. It's good for retention as well. People are shocked to discover I have a police record, but I love their greatest hits. Ha ha ha. ⁓ Alright, ⁓ awesome. So now take your break and come on back because I've got more to teach you.

—--

73

Harsh Bharadwaaj (00:00:00)  
In this exercise, we're going to add a new tool called ViewTag. This is similar to get tag, except this one returns ⁓ HTML. ⁓ And it returns actually a resource with UI colon slash slash view tag slash one, so the ID of the tag, with the MIME type of text HTML, and then with the text of raw HTML. ⁓ And why does it do this? Well, because we've got some really cool.

Agents now that can display that HTML for us automatically based off of this standardized resource type. So as an example, we can head over to Goose and we can say, please show me ⁓ tag two ⁓ in ⁓ Epic Me Visually. ⁓ And now it's connected to the EpicMe Workshop and boom, there's a visual representation of that HTML. Not all that impressive, I know, but

The fact that you can do HTML opens up the opportunity for a lot of really cool things. So that's your task in this exercise is to go to the view tag tool ⁓ and add support for MCP UI. Have a good time with this one.

—---------

86

Harsh Bharadwaaj (00:00:00)  
Like I said, this one was pretty quick. You're simply gonna add a UI metadata and then a preferred frame size, which is an array where the first ⁓ item is the width and the second item is the height. ⁓ And this is of course gonna be very dependent on the specific UI that you're rendering, but ⁓ you are better at determining what the r ⁓ preferred frame size should be than the agent is. So we know that we're gonna view the journal.

We wrote the journal UI or whatever. ⁓ and so we can see, yeah, max height, 800 pixels. So it's it's structured to support that. Okay, ⁓ great. So let's try this out. Can you please show me my journal? ⁓ And with that, would you look at that? Doesn't that look great? And now we don't have two scroll bars or anything. Like it ⁓ is scrolling just nicely. ⁓ you probably don't want to take up the entire ⁓

Chat ⁓ experience. So like if if it's a really long list, it probably makes sense to ⁓ have it scroll. And so that's why I ⁓ set the max height to 800 right here. ⁓ And so your preferred frame size should probably not be like infinity. Like that would probably not be great. ⁓ So yeah, just be a little cognizant of the user experience there. But this is how you can control ⁓ that initial frame size for ⁓ your applications.

—------

103

Harsh Bharadwaaj (00:00:00)  
Alright, it's break time, and this is a pretty good joke. I read it already. ⁓ Today my son asked, Can I have a bookmark? And I burst into tears, 11 years old, and he still doesn't know. My name is Brian. ⁓ my gosh. Yeah, that one's a pretty good one. Okay, great. Great work on this one. It's a good time to write down what you learned and ⁓ take a break, get blood flowing in your body and brain, and go give somebody a high five, tell them what you accomplished and feel good about yourself.

Good work on this one, and I hope that you had a really good time with it.

—-----------  
091

Harsh Bharadwaaj (00:00:00)  
So we want to make this post actually post. We want to make it work instead of just throw an error and give us this error right here. So what you're going to do in this exercise is implement a function which will handle sending the request to post to actually open up a link and then wait for the response to make sure that that was like that post request was successful. ⁓ And then we'll be able to build upon that utility to actually implement the rest of the functionality. So

Get to work on that utility, the emoji you're there to guide the way. Have a good time with this one.

—-------------

099

Harsh Bharadwaaj (00:00:00)  
So we actually have some cool UI that we show if a deletion is successful. So here I'm gonna say, I'm gonna delete the gym session, confirm, ⁓ and that is successfully deleted. But if we come back up to the UI, would you look at that? It's deleted, success. It like is all blurred out and stuff. That looks a little bit nice. To be able to do this though, we have to be able to ⁓ find out whether it was actually successful. And our tool sends back structured content that includes whether or not it was successful.

And that's what we use. So I'm showing you the solution here. Your job is to make it look like the solution. ⁓ Until we get the solution, we don't actually know whether or not the deletion was successful. ⁓ So ⁓ your job in this exercise is to not only make the tool call, which we're doing already, but also parse the response to verify that the tool call ⁓ was ⁓ successful. ⁓ And then we can ⁓ say that it was actually deleted. So you're going to be parsing the response. There's actually not a lot of code going on in here.

But it is ⁓ useful for you to understand like where things are going and ⁓ how to parse that response. We're using Zod for this because it is going across a network boundary and the iframe boundary. It's and so you wanna type-safe your boundaries, and so we're using Zod for that. ⁓ and I think that you're gonna have a good time with this one. So let's get into it.

—--------

79

Harsh Bharadwaaj (00:00:00)  
Okay, good job on that one. Let's get a random dad joke in here. What happens when you anger a brain surgeon? They will give you a piece of your mind. Ha ha. ⁓ Alright, great. So now is the time for you to take a quick break. Get yourself a drink of water, go give somebody a nice compliment or a hug or something. ⁓ And ⁓ all of those ⁓ things are good for your retention. Write down the stuff that you learned, of course, ⁓ and then when you're ready, come on back because there's still more for us to do.

—--------

84

Harsh Bharadwaaj (00:00:00)  
So now that we have actually rendered, we need to let the parent know hey, we've rendered and we're ready to start receiving events. We're gonna do this in a use effect. ⁓ I know some people are afraid of use effects. ⁓ But yeah, let's uncomment ⁓ our use effects call right here. ⁓ And we're gonna call window parent post message ⁓ UI lifecycle iframe ready. So the parent is expecting this to be called and it's waiting to send data and to send other events.

until this event is called. And so we're just saying, hey, we've rendered, we're ready to go, everything is good. And now the client can ⁓ know that okay, the UI is ready and displaying for the user. There could be some agents that ⁓ wait to continue generation until this message is received, different things like that. So it is important to do that. Like I said, doesn't actually ⁓ change anything in this UI. You're not gonna notice any change in the inspector or anything, ⁓ but ⁓ it is an important

thing to do in the UI that you are ⁓ serving up. ⁓ And we're doing that just in our journal viewer ⁓ route where we're rendering all this stuff. So there you go. That is how you send a post message to your parent to let them know that you're ready.

—-----

93

Harsh Bharadwaaj (00:00:00)  
So now that we've got things set up for being able to ⁓ send links to the parent, we also can send tool calls to the parent. And we're going to use lots of the same ⁓ structure because it's a very similar sort of thing. ⁓ And we w want to make it so the view details will work. And so when view details works, ⁓ it is going to trigger a tool call. ⁓ it's the view entry ⁓ tool call. So if I say, could you please show me my weekend hike with family journal entry, please?

I said please twice. I sometimes do that. and the AI that transcribed did not say please twice. So that's nice. ⁓ So here it's listing the entries so it can find the the one about the hike with with the family, and then it does the tool call, and there we are. ⁓ So I should be able to ⁓ get that just by clicking view details, and then it will immediately call the view entry tool, and then I can see that. ⁓ So

⁓ that is your objective for this exercise. Have a good time with that. You'll be able to build on top of what we already have. And ⁓ actually what that utility that we made has been slightly refactored to be a little bit more general. ⁓ So some of what you're gonna be doing is actually some TypeScript stuff just to make TypeScript happy. ⁓ but for the most part, it's gonna be pretty straightforward. So anyway, have a good time with this one. We'll see you when you're finished.

—----------

72

Harsh Bharadwaaj (00:00:00)  
All right, let's get our first example of rendering some UI to the client so that our users' agents can display something beautiful for our users to view. ⁓ We'll get into the interaction stuff a little bit later. We're going to use create UI resource from MCP UI Server to render just a raw HTML string. You've got your UI URI, you've got the content. That URI needs to start with UI colon slash slash.

⁓ you've got your content with raw HTML and encoding of text. And here's the HTML. You can have style, you can have like whatever, all the HTML you want in the world. ⁓ And it works great. ⁓ And what we're ultimately going to be able to do is something like this inside of a an agent that supports UI, like codename Goose here. ⁓ And we are ⁓ going to pro create this kind of user flow. So this is how things work. ⁓ The user wants to request weather information, for example.

The application, like Goose or some other ⁓ application that supports MCP UI, like Postman or something, ⁓ is gonna call get weather ⁓ on the client. The client will forward that onto the server. ⁓ The server is going to ⁓ fetch the weather data, it's going to create HTML content and then package that up as a UI resource, send that back as part of the tool call response, ⁓ and the client will send that to the app. The app is going to render that as HTML content and also display the visual weather card to the user.

So that is what you have to look forward to in this exercise. It's pretty simple. ⁓ There's just one step for this one, but I think it's a really great one. So let's get into it. We'll see you in the exercise.

—-------------------------------------