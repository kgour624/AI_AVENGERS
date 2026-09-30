134

Harsh Bharadwaaj (00:00:00)  
There's a difference between you have access and you have a ⁓ There's a difference between you don't have access and you have access to only a few things. ⁓ And then of course you have access to the world. Right now we say you have access to everything, ⁓ but often you want to be able to have a little bit more fine-grained control over the different things that you're giving access to a third party to. ⁓ And so that is what we're gonna be doing in this exercise. It has to do with scopes.

So here's how this works. The client is going to make a request with an access token. ⁓ The MCP server is going to ⁓ do the introspection with the auth server and find out what scopes that auth server or that token represents. ⁓ It's going to validate that that ⁓ those scopes include at least what the ⁓ at least something that the MCP server requires. And so we're going to have like a minimum number of scopes or minimum combinations of scopes that are required to make the MCP server actually useful.

If it does, then we can go ahead and do the thing and we can come back and everything's fine. If it doesn't, then we send back a 403\. And then we can also protect certain aspects of our MCP server and say, yeah, you have sufficient scope to make use of the MCP server, but you don't have sufficient scope to do this. So we're not even going to expose that tool for you. We won't even register it for you. ⁓ And that is what you're going to be doing in this exercise: adding a couple of utilities and protecting different parts of the server ⁓ so that the user can't.

just do everything. They can just do the things that they set their ⁓ auth token scopes to include. ⁓ Now that's actually something that they kind of get to choose. When they connect to the server in the first place, if we clear our auth state, ⁓ they get to select the scopes that are available. ⁓ You do have some control over ⁓ the scopes that the user can possibly select, of course, because you are in charge of this authorized page, or maybe your auth provider is and you have some configuration abilities.

But then the ⁓ user often you're going to give them the option to say, no, I don't wanna give it access to all of this stuff. ⁓ Depending on the situation, ⁓ maybe you're the one building both the service and the MCP server. That's more most likely. So ⁓ you want ⁓ them to just be able to do everything. But there are definitely situations where you want ⁓ to allow the user to say, you know what, I'm gonna give this to an LLM.

Harsh Bharadwaaj (00:02:23)  
I think I'm good with them reading, but I don't want them to write anything. Or maybe they can write tags, but I don't want them to write my entries or whatever. Like it kind of depends. And it might just make the user feel a bit a little bit more comfortable to make sure that the LLM can't do something ⁓ that would be really catastrophic for them. ⁓ So having scope control is really important in the world of MCP. So let's get into it.

—-------------------

131

Harsh Bharadwaaj (00:00:00)  
Now that we have the user information inside of our MCP server, we can start using it in some of our tools. Now we're already using it with the client, and so all of the tools we had already are accessing user-specific information. But I thought it would be kind of useful to have a who am I tool. This is a pretty common thing, like who am I currently logged in as? ⁓ That helps the LLM determine what to do and with certain questions and stuff, and even a ⁓ effectively a who am I resource. And so

Kelly actually added those for us already. Here's the Who Am I tool, ⁓ and you can also check for the resource. And so when you're all finished, this is what the experience should be like. Let's come here, we'll do Kelly, ⁓ and let's list our resources. ⁓ We've got user is one of our resources here now. ⁓ And also tools. We've got a who am I tool, and that will return Kelly. So that is your job in this exercise, is to make that work. It should be pretty quick and painless. Have a good time.

—------------

107

Harsh Bharadwaaj (00:00:00)  
So we're in the MCP inspector. We're going to go to open auth settings and quick OAuth flow. And we're going to get an error here because it failed to discover OAuth metadata. It needs to be able to ⁓ find where the authorization server URL is and all of that information. So if we look at our console, we'll see that the reason this failed ⁓ is because of ⁓ what's called the cross-origin resource sharing ⁓ policy. That's ⁓ shortened to cores.

we have it set to no access control origin header. So ⁓ basically it's trying to make a request ⁓ and that request is being blocked. If I keep my ⁓ developer tools open, when I click this, it will see all the requests that the client is trying to make. The first request it makes is to dot slash well known slash OAuth protected resource slash mcp. And the reason that it's making that request is because of the resource.

That I ⁓ am trying to request here. If if I had changed this to ⁓ my MCP and then tried to do a quick OAuth, then we'll see we we're getting my MCP showing up there as well. So ⁓ that's what's going on. It's trying to discover the ⁓ different endpoints that it needs to hit to ⁓ complete the OAuth flow. ⁓ And that whole process ⁓ is actually as part of this.

guided flow that we get in the MCP inspector in this version. So it's gonna discover the metadata, then do client registration, prepare authorization and get the token, all of that stuff. And so all of that is ⁓ requires some metadata discovery that needs to happen. ⁓ But for us to be able to do that, our client needs to be able to make a request to ⁓ the that is a cross origin request. The origin here being localhost 4000, we need to make a request to localhost

5600 or 56,000. And so we need to configure our server to allow any origin, like localhole 4000, ⁓ to be able to access ⁓ our origin, which is a a different origin. ⁓ Normally or often, you're going to have a web application that's at like myExample.com ⁓ and it's going to be making requests to the same domain. And so you don't run into this. But when you're doing OAuth stuff, you're going to have a one client that's at one domain and your server that's in another domain.

Harsh Bharadwaaj (00:02:18)  
And for security reasons, you need to establish that ⁓ sharing mechanism, sharing of resources over that ⁓ that origin. ⁓ So we're we're happy about this. This is a good thing that we have this security built into the browser. ⁓ But we do need to kind of sidestep it or at least instruct it that it is okay and expected for these dot slash well or or slash dot ⁓ well-known ⁓ endpoints to be requested from any origin.

So that's what you're gonna be doing in this exercise. We've got a utility that we have written for you so you don't have to worry about the intricacies of actually adding those headers and stuff. It's really not all that important for your learning. The most important thing here for you to understand ⁓ is that those headers need to be added. So you're just going to use the utility and hopefully that ⁓ that transfer of knowledge comes and you understand what these utilities are doing and how that's solving the problem for us.

making it so that we can actually perform this discovery. When you're done with this exercise, ⁓ the whole flow is not going to be finished. We got still a number of things we need to do to make all of this work. ⁓ But this first barrier will be knocked down. So ⁓ have a good time with this one. We'll see you when you're done.

—-------------------

135

Harsh Bharadwaaj (00:00:00)  
All right, let's talk about scope. So ⁓ if I go to ⁓ yeah, let's just connect, ⁓ and I say, you know what, I don't want this application to be able to write to entries or read entries or read tags. I just want it to write tags. That's all. ⁓ Okay, so that's my scope. I'm gonna authorize ⁓ as codey. And then here I come to prompts, I list my prompts and I suggest tags for entry ID ⁓ one.

That seems wrong. Like I shouldn't be able to read that. ⁓ And in fact, look, ⁓ I literally have tags and I have entries. So ⁓ yeah, that's that's wrong. I shouldn't be able to do that at all. ⁓ same thing with resources and tools. Like I should be able to lock these things down so that ⁓ we respect the scopes that the user has selected. And in fact, Kelly added a handy little ⁓ piece of information to ⁓ the who am I ⁓ resource or tool. ⁓

So you can see what scopes there are, just so you can double check. Yes, I only should be able to write tags. ⁓ So in this exercise, ⁓ you're going to lock down sampling and prompts. You can feel free to lock down the rest of them. Kelly's gonna do that for us. It's all pretty much the same ⁓ if you want to ⁓ go for it. But ⁓ mostly I just want you to lock down prompts. You should only be able to use the suggest tags prompt if you have read for both the ⁓ tags and entries.

⁓ and technically like there's a pretty good argument that you have to be able to write as well, but ⁓ maybe they want to suggest tags for some other reason than actually applying those tags. So ⁓ yeah, you should just be able to use suggest tags if you have read for both write and tags. ⁓ and so if as you're testing this out, just go through this and uncheck these and make sure that you don't have access to the prompts.

And then check them and make sure that you do. ⁓ And that is what you're going to do in this exercise. Have a good time with it.

—-------------

108

Harsh Bharadwaaj (00:00:00)  
So let's come over here to our editor. We'll open up the index right here. And we're gonna bring in this with cores utility. You can feel free to dive into that and figure out how it works. But basically, we just need to make sure that any request that comes to ⁓ /dot well known ⁓ is going to have the proper course headers to say, yeah, any of you like ⁓ use use this however you like, that's fine. ⁓ So we're going to grab this, we're gonna say with cores, ⁓ and that's gonna be our handler.

So handler, there we go. ⁓ And then we're going to have our get cores header. This is going to accept a request and it's gonna return the cores headers that we want to have applied. So we're gonna say if the request URL includes slash dot well known, then it's totally fine. Or we can maybe ⁓ do path name, ⁓ like URL path name and it starts with or whatever if you wanna be a little bit more specific, but this works just fine. ⁓

And so one of the headers we want is access control allow origin. So we're fine with any origin accessing ⁓ these resources. And then we also want to specify which methods. We don't want somebody to try and post or put or delete at these ⁓ these routes. There's not really a a reason they would need to do that, so we'll specify those. You can do a get, you can do a head, or you can do an options request. The options ⁓ request is important specifically because ⁓ that type of request is ⁓ just trying to say, hey,

Do you allow cross-origin requests? And so we're gonna keep that one. Head request ⁓ is useful just to to get back a status and stuff. Sometimes ⁓ clients are going to make that request just to make sure that the domain works and everything. And then the get request is the more common type of request. ⁓ And then we'll have our access control. ⁓ access control. ⁓ I'm kind of waiting for my AI assistant to kick in here, but it's not. So ⁓ we'll just do this ourselves like ⁓ good old times of old. ⁓

Protocol, protocol version. ⁓ And what this does ⁓ is ⁓ you you can control through this ⁓ the different headers that are allowed to be on that request that's coming over to your ⁓ resource. ⁓ Maybe you don't want to accept any headers that include like a cookie or a token or something like that, just you know, for plausible deniability, like an I never received your ⁓ your private things, like just don't don't send those to me. It's not gonna work anyway.

Harsh Bharadwaaj (00:02:27)  
But ⁓ in the MCP spec, ⁓ it does say that clients will typically include the MCP protocol version in requests. And so ⁓ there's no harm in us accepting that header. And so it's ⁓ we're not really gonna use it for these particular requests. ⁓ but because they're going to include it, we don't want it to fail just because they included that extra header. ⁓ you might check and like test with the clients that you're interested in supporting, making sure that.

They're not sending any extra headers, and maybe you'd include those in here if that was important to your use case. ⁓ But we're gonna try and keep it locked down generally to avoid ⁓ potential problems, getting too much information that we shouldn't have or ⁓ accepting requests ⁓ that are of a type that we shouldn't really accept. ⁓ So with that all set up now, if I try to do a quick OAuth flow, here let's try this again. Here we go. Failed to discover OAuth metadata.

Let's see what the problem is now. ⁓ it's just to make sure that the cores problems are gone. ⁓ And yeah, now it's a 404 not found. That's what we're going to be dealing with next. So we're no longer getting the ⁓ access control ⁓ problems. ⁓ And in fact, if you look at the response headers on ⁓ this resource, you are gonna see the access control allow headers and allow methods and allow origin. And that was your objective here, is just to to show those.

So with that, we now are ready to continue and make ⁓ this metadata discovery actually work. Good job.

—---------------

133

Harsh Bharadwaaj (00:00:00)  
Hey, good job on that one. Now it's time for a break. It's good for your brain to just take a break, write down what you learned, get a drink, go compliment somebody's outfit or something. ⁓ And also let's hear a joke. Did you know crocodiles could grow up to fifteen feet? But most just have four. Haha. ⁓ all right, great. Now is a good time for you to ⁓ take your break, make sure you write down what you learned, ⁓ and then come back because we're not done. I still have more things to show you and teach you and

You have more things to do, so when you're ready, come on back, I'll be waiting for ya.

—-------------

110

Harsh Bharadwaaj (00:00:00)  
Alright, let's start in our handler here. When we get a request to well-known authorization server, we are going to want to handle the OAuth ⁓ authorization server request. ⁓ And so with that, ⁓ we're gonna need to bring that in from our authors. And then of course we need to implement that. So let's export ⁓ the async function, handle ⁓ OAuth authorization server.

What this is going to do is ⁓ it actually we don't need the request in this case, so that's fine. ⁓ and yep, we're gonna need the URL for ⁓ our ⁓ authorization server. ⁓ We're gonna construct a new URL. So we're gonna go to ⁓ the ⁓ OAuth authorization server endpoint ⁓ on that ⁓ that domain. And then we're gonna make a fetch request to that URL. We actually don't need toString, so there you go. That's nice.

and then we're gonna return the response.json. Or ⁓ you could take the ⁓ yeah data, ⁓ await that, and then return response.json of the data. That works just as well. ⁓ and with that then ⁓ we can come back here. Let's make sure that we're not passing the request. There we go. ⁓ And now ⁓ if we do the quick OAuth flow, that's gonna take us to ⁓ authorize.

And the reason it did that is because we sent back this configuration. ⁓ And that said, hey, the issuer is here, the authorization endpoint is here. And so it's like, great, I know what to do with that. Let me take you over to the authorization page, to that page with ⁓ the scopes and all the other like response type, client ID, all of that stuff. ⁓ So ⁓ and in fact we also have our registration endpoint. So that's something that it did as well. We didn't quite see that. So let's back up and let's do the guided flow. ⁓ actually it shows us.

right here, just hitting back. It just went so fast. ⁓ so yeah, it started out with the metadata discovery. ⁓ And ⁓ it ⁓ said, there was a problem getting the resource metadata with the protected resource. ⁓ here's what you should read about that. And we'll deal with that here in a little bit. ⁓ But I did manage to get the auth OAuth authorization server stuff. And here's what I got from that. That should look kind of familiar. ⁓ And ⁓ from that it was able to register a client. And so here we've

Harsh Bharadwaaj (00:02:19)  
registered our client. We now ⁓ have our our client ID and so the client can save that and submit that as part of the ⁓ authorization request and all of that. ⁓ And then it just automatically opened this for us, which took us to this page. ⁓ and actually, if you go to the homepage ⁓ of our auth server, you'll now see that you've got ⁓ a client in here. You might have one or two, depending on how many times you went through the registration process.

⁓ But yeah, so you can dive into here. This is what's saved ⁓ in our OAuth server storage for each one of the clients. As the clients are dynamically registering themselves and saying, hey, I'd like to get tokens for your users, we save those in our database ⁓ on the OAuth server side. That's the authorization server's job. You don't need to worry about that as an MCP server author ⁓ for the resource server side of things, but it can be useful to understand how that all is working behind the scenes.

On the authorization server side, so you you can build a better resource server. ⁓ So that is being stored for us. And now we can come over here. And ⁓ if we select one of our users that we want to authorize as, so we'll do Cody at KCD. ⁓ That will ⁓ actually now, here, let's go here now. ⁓ we have a grant that has been added to our database as well. So we've got our ⁓ the user ID that it's associated to, the client that it's associated to, here's the scope.

here are ⁓ like all of the other important things ⁓ for this token. ⁓ And now we go to continue OAuth flow and that'll give us ⁓ our authorization code. Come back here and paste that in. And now if we hit continue and then continue, boom. We are now authenticated. We have a valid access token. ⁓ we can see when it expires, we see its scopes and the refresh token, all of that now ⁓

Whatever client we're using, whether that be here the MCP inspector or Goose or VS Code or Cursor or ChatGPT or Claude or whatever, ⁓ they will all manage that token from that point on and send that as an authorization header to your resource server, which you can then ⁓ validate and all the stuff, which is where what we're gonna get to here in future exercises. ⁓ So in review, all that we had to do to make ⁓ this piece work was add a handler.

Harsh Bharadwaaj (00:04:41)  
For clients that are going to treat us as an authorization server, we effectively ⁓ proxy that request to the actual authorization server ⁓ and respond with the same metadata, and then it's able to complete the authorization flow for us and get us an auth token.

—--------------

109

Harsh Bharadwaaj (00:00:00)  
If we take a closer look at the errors that we get when we go through the quick auth flow, you'll notice ⁓ it actually is making a couple of requests, which is kinda interesting. So first it checks this protected resource to see whether it's gonna send anything back. It doesn't right now. We'll get our server doing that here in a little bit. And then when that doesn't work, it checks the OAuth protected resource. So it it just basically removes the slash MCP off of the end of this ⁓ to see, hey, is

Their access to the entire server, not just the MCP resource, but everything. Like, is that been configured? We haven't configured that either, so that's not going to show up here. But then it makes another request to OAuth authorization server. And what this is doing is at the beginning, it treats our server as the ⁓ as a resource server. ⁓ And once we don't have the correct endpoints for those protected resource endpoints, then it's going to check, hey, maybe you're the authorization server.

And the resource server. Maybe you're both. So it's going to make a request to s ⁓ slash well dot well known slash OAuth authorization server. And this should include some metadata for ⁓ like all of the metadata that it needs to perform the authorization, assuming that we are the authorization server. ⁓ This is actually ⁓ something that ⁓ normally we shouldn't be doing with MCP because MCP servers should actually be the resource server.

Though they can still be the authorization server as well and kind of be both. But ⁓ this is a common fallback for clients and you should be handling this in your MCP server because they may skip these protected resource requests altogether and just assume you are the authorization server. So it's important that you support this particular endpoint. Then finally it tries to do an open ID configuration. We're not gonna go quite that far. ⁓ So we're gonna stop with this one and we're going to implement this.

Now, we are not actually the authorization server in this context. Our MCP server is a resource server. But because this is so common, what we're going to do ⁓ is we'll implement this endpoint, but get the configuration from our authorization server. ⁓ Now, as you're running this workshop app, you should be also running ⁓ localhosts 7788\. That is the EpicMe app. ⁓ That is at the root of this project. If you're running the workshop app, this will be running automatically as a sidecar process for you.

Harsh Bharadwaaj (00:02:20)  
And this is the.well known slash OAuth authorization server that that server is already serving. ⁓ So what our job in this exercise is to do is to ⁓ handle the request to our MCP server. ⁓ And in that handler, we're going to make a request to the 7788, our authorization server, to get this configuration and send it back. So we're effectively acting as a proxy for our authorization server.

And that way we handle clients that aren't up to date and are treating us as an authorization server. This isn't technically the ⁓ the most correct way to do things according to the spec, but it's a pragmatic and practical way to deal with ⁓ clients that assume that we are the authorization server, ⁓ and it will work just fine. In future exercise ⁓ step we will handle the protected resources and so we'll kind of handle both of these. ⁓ So

That is your objective is to create an endpoint that handles this and effectively proxies the result of this back ⁓ as a response to that. Okay? It's I I talked a lot in this exercise like leading up to this, but it's actually not a ton of code that you're gonna be writing. Hopefully it's pretty straightforward. Have a good time with this one.

—-------------------

138

Harsh Bharadwaaj (00:00:00)  
Hey, it's like Christmas time. Kelly gave us a gift. All right, let's talk about this. It was just a little bit difficult to describe exactly what was needed here. And so Kelly just gave this to us, and hopefully, it makes sense ⁓ without having to write it yourself. So basically, these are the minimal scope combinations that make this server actually useful. ⁓ And in our case, any one scope will be perfectly fine. In some cases, maybe you need at least two scopes.

For the server to be useful ⁓ for this case, but maybe one scope for this case, whatever. ⁓ And so with this, then we have this utility that says, do I have sufficient scope? So we're gonna take the auth info and we're gonna go through every one of these combinations of scopes. ⁓ And if one of them, at least one of them, the user has all of the scopes necessary in that combination, then they have sufficient scope to do something useful in the server.

So all we need to do now ⁓ is make sure that ⁓ or ⁓ create another utility function that returns a www ⁓ authenticate header with the error of insufficient scope and then include that in the error description. ⁓ So we're gonna export a function ⁓ and we're gonna grab that auth header. We know that they have the auth header. We're going to ⁓ create a URL for the protected resource. We've got the bearer ⁓ epic ⁓ umme, we've got the resource metadata.

And then if they have the auth header, ⁓ then we have insufficient scope ⁓ as the error. And then if ⁓ and then we also include as the error description that they need at least one of these scopes. ⁓ Now you might be tempted, I was tempted, to also add a scope that takes all the supported scopes and joins them and says, hey, this is the scope that you should have to access this resource. ⁓ That like technically, if you want to access the full resource, then yeah, this is the scope you should have.

I'm not sure I whether I'm being pedantic or just like trying to be like follow the spec well, but it feels ⁓ right to me that we don't do that. ⁓ because it ⁓ you don't necessarily want the user to just give the keys of the kingdom to the MCP server. I don't know, maybe you do. ⁓ but it just seems to me that ⁓ just because we support all of these scopes doesn't necessarily mean we should include all of those.

Harsh Bharadwaaj (00:02:20)  
we have actually another mechanism to communicate what scopes are available, and lots of clients will just use that anyway. ⁓ So instead we use the error description to ⁓ communicate. Here's the different combinations of scopes that you could have that ⁓ that would make sense. All right, sweet. So with that done, we still need to go and add that ⁓ handling to our index. ⁓ So let's come here, ⁓ right here.

And before we let them into the MCP server, let's just make sure that they have sufficient scope. ⁓ And if they do not, then we will ⁓ send them on back. ⁓ And with that, now we can ⁓ come here, try to connect. We will remove all scopes. We have not sufficient scope. In fact, we could even add right user write, because ⁓ that is not one of our scopes that makes our server useful. ⁓ So click that, come here, ⁓ and we get success.

Here, ⁓ technically we do have an auth token, but we get an error here, and the error is going to indicate ⁓ YOIK. ⁓ error posting forbidden. And inside of the client here, it's gonna see those errors and ⁓ a pop proper client would probably show specifically what the problem was, and maybe it'd even trigger another auth flow, which I think is probably what it should do. ⁓ one thing I just noticed here that the AI did that we should not be doing.

Is the resource metadata. You don't put resource metadata on ⁓ this forbidden ⁓ response here. Doesn't make any sense. And on top of that, ⁓ if we've gotten this far, we do have an auth header, so we don't actually need to handle the auth header either. ⁓ whoops, we'll just get rid of that guy, get rid of that guy, get rid of that guy. ⁓ and that also means we don't necessarily need the request, ⁓ which means we can come back here and get rid of the request.

And like it wanted to, we can do it on a single line if we really want. If ⁓ some of you are like, gross. You're looking at my code like, I can't, how dare you. ⁓ Look at the lack of semicolons. my goodness. Ha. It's my course. I get to write it the way I want to. ⁓ all right, so that's what we do in here. ⁓ this is how I do it. ⁓ you have your minimal combinations of scopes, you have a utility to

Harsh Bharadwaaj (00:04:40)  
Determine whether they have sufficient scope to make use of the server at all. And then you add a 403 forbidden with the w authenticate with an error that says insufficient scope. And then the client can do what it will with that. ⁓ And then we verify, once we have the auth info, that they have sufficient scope. And if they do, then they can get in. Let's actually verify that that also ⁓ does work. ⁓ Auth settings, let's clear auth state, connect, ⁓ and connect is Olivia. Ta-da. ⁓

And now we can access the server and we're good. ⁓ All right, great job verifying sufficient scopes. ⁓

—--------------------------

119

Harsh Bharadwaaj (00:00:00)  
Good job in this exercise. Now it's time for a little dad joke for that chuckle that'll help you live longer. And ⁓ time to get a drink of water and move your body, get the blood flowing, write down what you learned so you remember it better. Here's the dad joke. To be frank, I'd have to change my name. Ha ha. ⁓ All right, go ahead and take your break and we'll see you when you're done.

—--------------  
136

Harsh Bharadwaaj (00:00:00)  
Let's start out this by making ourselves ⁓ a couple of utilities. So we're going to make a supported scopes ⁓ that's going to have ⁓ user read, entries read, entries write, tags write, and tags ⁓ tags read and tags write. ⁓ these are the scopes that our MCP server requires to do its job. This is not necessarily all of the scopes that our OAuth provider does ⁓ support. These are just the ones that we care about, the ones that are gonna be important to us.

We're also going to export a type called supported scopes ⁓ to make things a little bit easier for some other utilities. ⁓ And we're going to make a validate scopes utility. And what this does ⁓ is it takes ⁓ the auth info ⁓ and an array of scopes and it tells us if that auth info scopes includes every one of the items in that scopes array. ⁓ And then we're going to go over to our index and add this other utility to our agent that kind of builds on top of that.

so we're gonna grab our ⁓ supported scopes, we're gonna take any number of scopes as an array of supported scopes, and then we're gonna validate those with our auth info. ⁓ And it will return true if we have all of those scopes. ⁓ So now we can go to prompts ⁓ and we can say if the agent does have entries read and tags read, then it can suggest tags. ⁓ And with that now, whoops, ⁓

The prompts should be protected. ⁓ So ⁓ it doesn't really make sense to or or you shouldn't be allowed to suggest prompts if you can't read and ⁓ read entries and tags ⁓ because that's part of what comes back from that prompt. So we gotta protect stuff. ⁓ Sampling is a similar sort of thing, except here we're gonna exit early if we can't do all of these things. You need to be able to ⁓ read entries, tags, and write entries and tags.

Because scopes or because we're going to read those ⁓ bits of information, give it to the LLM, and then ⁓ get that back and then write ⁓ information. So we're adding tags to entries and we're possibly creating tags. ⁓ So with that set up, we should be able to come over here, connect. Let's ⁓ remove our ability to read tags alone, just that one. And we'll pull up Olivia. And now if we go to prompts and list prompts.

Harsh Bharadwaaj (00:02:25)  
⁓ It's not gonna work because we don't have any props. And it just occurs to me now ⁓ that that actually ⁓ we can do better than that. ⁓ And I am now eating my words. ⁓ you might recall in our fundamentals workshop where I insisted that you should just list things in your capabilities, ⁓ like up front. ⁓ actually, if you've gotten it this far, you've probably already had it updated in the instructions where I say, no, actually don't do that. ⁓

⁓ but ⁓ yeah, so I just it felt really good to me and felt right to me that you would actually list your capabilities. ⁓ Yeah, that I think is probably unadvisable and just trust the ⁓ that the MCP server is gonna keep these things up to date. That I I kinda have a mixed feeling about this because you could start the N MCP server without prompts and then later add a prompts. ⁓ I'm not sure what to what to do about this. So ⁓

I feel like this might be a gap in the spec. So just be aware of that. ⁓ One thing that we could do ⁓ is add some logic here that initializes the MCP server, ⁓ determining whether to include the prompts based off of the scope. ⁓ And so yeah, we could we could do that. Feel free to ⁓ continue with that. I'm gonna leave the capabilities defined here because we can dynamically change what prompts and and tools are available.

And I just would hate to have a situation where you start with no prompts and that's what gets initialized. And then ⁓ as you continue, ⁓ you end up with prompts. ⁓ And now the client doesn't know that you have prompts. ⁓ So yeah, again, seems like a little bit of a gap in the spec. ⁓ We'll we'll work on that, don't worry. ⁓ Let's also test the sampling. ⁓ So when we create a tool or create an entry, rather, ⁓ then it should suggest sampling except

We don't have the ability to read tags, and so no sampling is requested. ⁓ Awesome. Now let's clear the auth state and let's verify that if we do have the necessary permissions, we can do all of those things. So now we do have prompts. We can list the prompts. We can get an entry ID. no entry found with that ID. well, I wonder what entries we've got. Hey, look, autocomplete. Amazing. ⁓ get that prompt. Boom. Let's go. And then tools. ⁓ we can now create an entry. Bloo. ⁓

Harsh Bharadwaaj (00:04:47)  
And run the tool and boom, we do not have a sampling. ⁓ I am not sure why that didn't work out. So ⁓ I do have an inkling though. Here, let's reconnect. Let's go to tools ⁓ and create an entry, daloo, doo, and run the tool. Yeah, still no sampling. I think I know what's going on here, and this is actually a bug in the Cloudflare implementation ⁓ of the MCP agent. So this is gonna be fun. Why don't we go do this together?

we'll go to sampling ⁓ and it's in the client capabilities. So let's add a console log. ⁓ we have permission. ⁓ And then let's console log the client capabilities. And this is going to be the problem right here. In fact, let's also console log the client capabilities when we init. So ⁓ it's this dot server dot ⁓ server get client capabilities. So init.

Client capabilities. Okay. Great. So here we are with this. Let's actually, ⁓ I'm gonna restart the server just to be super duper super sure that this is good. Alright, so we've got this. We're going to disconnect. ⁓ And we're all set there. Okay, good. Now we're gonna connect. ⁓ And this is busted. ⁓ I know why. Here, hold on. ⁓ Let's refresh. Okay, connect. Boom.

Alright, so we haven't actually initted the server yet, so there there shouldn't be any of those logs for us. ⁓ Let's go to Cody, continue authflow. That was successful. ⁓ yeah, where's the our log capabilities? ⁓ there it is. Init client capabilities undefined. So this is ⁓ like I said, a bug in the ⁓ implementation of MCP ⁓ with Cloudflare, and so that's why the ⁓

The sampling didn't work for me. This is how I debug it, so I'm gonna leave this in the video. Hopefully that is helpful to you. ⁓ feel free to ⁓ actually hopefully by the time you watch this, that has been fixed in the Cloudflare implementation. ⁓ if not, ⁓ or if it has, then hooray, here's like a little ⁓ tip or a piece of history for you. If it has been fixed, then yeah, awesome.

Harsh Bharadwaaj (00:07:10)  
Okay, great. So you can continue and ⁓ protect our tools ⁓ and ⁓ resources as well if you want to add ⁓ scopes. But if we just go ahead and review this, ⁓ all we did was add the scopes that we support in our MCP server. We added a utility to validate whether ⁓ the auth info has those necessary scopes. We added another utility to our agent to make it even easier for clients or our ⁓ tools and server resources and prompts to validate those scopes.

And then we validated ⁓ prompts ⁓ so that it verifies that we do indeed have the necessary scope to suggest tags. ⁓ And ⁓ then we explored a little problem with the agent ⁓ or with the s the ⁓ MCP implementation from CloudFlyer. ⁓ Hopefully that was at least entertaining. All right, we're done. Thank you. ⁓

—-------------------

138

Harsh Bharadwaaj (00:00:00)  
Let's start out this by making ourselves ⁓ a couple of utilities. So we're going to make a supported scopes ⁓ that's going to have ⁓ user read, entries read, entries write, tags write, and tags ⁓ tags read and tags write. ⁓ these are the scopes that our MCP server requires to do its job. This is not necessarily all of the scopes that our OAuth provider does ⁓ support. These are just the ones that we care about, the ones that are gonna be important to us.

We're also going to export a type called supported scopes ⁓ to make things a little bit easier for some other utilities. ⁓ And we're going to make a validate scopes utility. And what this does ⁓ is it takes ⁓ the auth info ⁓ and an array of scopes and it tells us if that auth info scopes includes every one of the items in that scopes array. ⁓ And then we're going to go over to our index and add this other utility to our agent that kind of builds on top of that.

so we're gonna grab our ⁓ supported scopes, we're gonna take any number of scopes as an array of supported scopes, and then we're gonna validate those with our auth info. ⁓ And it will return true if we have all of those scopes. ⁓ So now we can go to prompts ⁓ and we can say if the agent does have entries read and tags read, then it can suggest tags. ⁓ And with that now, whoops, ⁓

The prompts should be protected. ⁓ So ⁓ it doesn't really make sense to or or you shouldn't be allowed to suggest prompts if you can't read and ⁓ read entries and tags ⁓ because that's part of what comes back from that prompt. So we gotta protect stuff. ⁓ Sampling is a similar sort of thing, except here we're gonna exit early if we can't do all of these things. You need to be able to ⁓ read entries, tags, and write entries and tags.

Because scopes or because we're going to read those ⁓ bits of information, give it to the LLM, and then ⁓ get that back and then write ⁓ information. So we're adding tags to entries and we're possibly creating tags. ⁓ So with that set up, we should be able to come over here, connect. Let's ⁓ remove our ability to read tags alone, just that one. And we'll pull up Olivia. And now if we go to prompts and list prompts.

Harsh Bharadwaaj (00:02:25)  
⁓ It's not gonna work because we don't have any props. And it just occurs to me now ⁓ that that actually ⁓ we can do better than that. ⁓ And I am now eating my words. ⁓ you might recall in our fundamentals workshop where I insisted that you should just list things in your capabilities, ⁓ like up front. ⁓ actually, if you've gotten it this far, you've probably already had it updated in the instructions where I say, no, actually don't do that. ⁓

⁓ but ⁓ yeah, so I just it felt really good to me and felt right to me that you would actually list your capabilities. ⁓ Yeah, that I think is probably unadvisable and just trust the ⁓ that the MCP server is gonna keep these things up to date. That I I kinda have a mixed feeling about this because you could start the N MCP server without prompts and then later add a prompts. ⁓ I'm not sure what to what to do about this. So ⁓

I feel like this might be a gap in the spec. So just be aware of that. ⁓ One thing that we could do ⁓ is add some logic here that initializes the MCP server, ⁓ determining whether to include the prompts based off of the scope. ⁓ And so yeah, we could we could do that. Feel free to ⁓ continue with that. I'm gonna leave the capabilities defined here because we can dynamically change what prompts and and tools are available.

And I just would hate to have a situation where you start with no prompts and that's what gets initialized. And then ⁓ as you continue, ⁓ you end up with prompts. ⁓ And now the client doesn't know that you have prompts. ⁓ So yeah, again, seems like a little bit of a gap in the spec. ⁓ We'll we'll work on that, don't worry. ⁓ Let's also test the sampling. ⁓ So when we create a tool or create an entry, rather, ⁓ then it should suggest sampling except

We don't have the ability to read tags, and so no sampling is requested. ⁓ Awesome. Now let's clear the auth state and let's verify that if we do have the necessary permissions, we can do all of those things. So now we do have prompts. We can list the prompts. We can get an entry ID. no entry found with that ID. well, I wonder what entries we've got. Hey, look, autocomplete. Amazing. ⁓ get that prompt. Boom. Let's go. And then tools. ⁓ we can now create an entry. Bloo. ⁓

Harsh Bharadwaaj (00:04:47)  
And run the tool and boom, we do not have a sampling. ⁓ I am not sure why that didn't work out. So ⁓ I do have an inkling though. Here, let's reconnect. Let's go to tools ⁓ and create an entry, daloo, doo, and run the tool. Yeah, still no sampling. I think I know what's going on here, and this is actually a bug in the Cloudflare implementation ⁓ of the MCP agent. So this is gonna be fun. Why don't we go do this together?

we'll go to sampling ⁓ and it's in the client capabilities. So let's add a console log. ⁓ we have permission. ⁓ And then let's console log the client capabilities. And this is going to be the problem right here. In fact, let's also console log the client capabilities when we init. So ⁓ it's this dot server dot ⁓ server get client capabilities. So init.

Client capabilities. Okay. Great. So here we are with this. Let's actually, ⁓ I'm gonna restart the server just to be super duper super sure that this is good. Alright, so we've got this. We're going to disconnect. ⁓ And we're all set there. Okay, good. Now we're gonna connect. ⁓ And this is busted. ⁓ I know why. Here, hold on. ⁓ Let's refresh. Okay, connect. Boom.

Alright, so we haven't actually initted the server yet, so there there shouldn't be any of those logs for us. ⁓ Let's go to Cody, continue authflow. That was successful. ⁓ yeah, where's the our log capabilities? ⁓ there it is. Init client capabilities undefined. So this is ⁓ like I said, a bug in the ⁓ implementation of MCP ⁓ with Cloudflare, and so that's why the ⁓

The sampling didn't work for me. This is how I debug it, so I'm gonna leave this in the video. Hopefully that is helpful to you. ⁓ feel free to ⁓ actually hopefully by the time you watch this, that has been fixed in the Cloudflare implementation. ⁓ if not, ⁓ or if it has, then hooray, here's like a little ⁓ tip or a piece of history for you. If it has been fixed, then yeah, awesome.

Harsh Bharadwaaj (00:07:10)  
Okay, great. So you can continue and ⁓ protect our tools ⁓ and ⁓ resources as well if you want to add ⁓ scopes. But if we just go ahead and review this, ⁓ all we did was add the scopes that we support in our MCP server. We added a utility to validate whether ⁓ the auth info has those necessary scopes. We added another utility to our agent to make it even easier for clients or our ⁓ tools and server resources and prompts to validate those scopes.

And then we validated ⁓ prompts ⁓ so that it verifies that we do indeed have the necessary scope to suggest tags. ⁓ And ⁓ then we explored a little problem with the agent ⁓ or with the s the ⁓ MCP implementation from CloudFlyer. ⁓ Hopefully that was at least entertaining. All right, we're done. Thank you. ⁓

—--------------

128

Harsh Bharadwaaj (00:00:00)  
I don't know if you noticed, but this whole time we have not done anything with tools, resources, or prompts or anything. And that's because it until this exercise, it ⁓ didn't work. ⁓ and what and it still doesn't work until you finish with this exercise. But when you're done, you will not only be able to use the tools as usual, but you'll also have a new tool called Who Am I, which allows you to get information about the currently logged in user. ⁓ But to be able to do that, we have to be able to get that information into our MCP server.

And to do that, we are going to be doing some conflare specific stuff. ⁓ So here we're going to ⁓ inside of our request handler set the context props to have the auth info. And that will give us access to that inside of our ⁓ class on this.props. ⁓ That's gonna be like possibly undefined, so we're gonna have to do some ⁓ checks and use some invariant stuff and and all that stuff. ⁓ but it's gonna be good.

And then we're going to use that information to set up our database client to have the user's auth token. That way, ⁓ we every request that we make to the database, to our our backend includes the auth token and it can resolve the user and ⁓ use that user's token to perform whatever database ⁓ requests are necessary. Every application is gonna be different. Maybe your MCP server has ⁓ direct access to the database, and so you're fine once you know the

Who the user is and what their ideas, and you can make those sorts of ⁓ queries to the database. ⁓ We have a separate application that is handling these requests through our ⁓ client. And so we have to provide the auth token to that. ⁓ So those are the the two things you're gonna do. And once you've done that, then you can start accessing ⁓ all of the things that the ⁓ database client can access, and we can add these additional tools. ⁓ I wanna talk really quickly about the flow as far as

The Cloudflare worker is how to get from the ⁓ request into our durable object or our agent. So the client makes a request with the authorization header. That gets to our Cloudflare worker. ⁓ We resolve that ⁓ with the auth server. That comes back with our auth info. We set the context props auth info. And then the Cloudflare worker makes a request to our durable object, ⁓ which is going to include those props.

Harsh Bharadwaaj (00:02:25)  
⁓ And then from there we can ⁓ get the auth info from the props and we can make ⁓ create our database client which ⁓ will include that auth info token or the OAuth token. ⁓ And then any requests that we make using that database client are going to include that user's information, and then we can send the response. ⁓ So that's the basic workings of everything. It's not a ton of code that you have to put in here, but there's a little bit. So I hope you have a good time with this one.

—--------------

118

Harsh Bharadwaaj (00:00:00)  
So we need to find the URL that effectively takes us to this endpoint. If we take a look at our index, that's gonna be this URL. So we're gonna copy that ⁓ and we'll come over here to get first the request. And then we'll yeah, there we go. ⁓ Exactly that. ⁓ Thank you, AI assistant. ⁓ And then our resource metadata, ⁓ which is ⁓ called an auth param. So each one of these is authparams. ⁓ we have our realm and then comma.

Resource metadata and then here is the URL toString. ⁓ Technically, I guess you could have not do that, and a two string is called automatically, but however you want to do it. ⁓ And that ⁓ tells the receiver of this response where to go to get resource metadata. So if you have some special URL, some special place they need to go to get it, ⁓ then you could change that here. ⁓ And also this is going to take care of clients that are a little lazy and aren't just gonna guess. ⁓ So that's nice.

⁓ coming over here, we have to add the request here so that we can derive that URL. ⁓ And if we save this and try to connect, just to be sure, it does still work. We could double check ⁓ with the the network tab, click on this, and when it ⁓ runs that, we actually, yeah, you're not actually going to be able to see this because the ⁓ request is all happening.

On the server side of the MCP client. So ⁓ this request is coming from the server side. So you won't see it in the in the network tab here, ⁓ which ⁓ hopefully is useful that I showed you that. ⁓ Because maybe some of you are like, wait, where is the request coming for slash MCP? You know, I want to see that ⁓ that dub dub dub authenticate response. You're not gonna see that because that's all happening on the server side of our MCP ⁓ inspector. So

If you want to dive into that, you could go into the source code and dig around, and you would be able to add a log that would show up in here. ⁓ here you can actually see the log for our 401, but it's not going to show us the ⁓ the www header. ⁓ but it's in there, and that's how it knew to make this request right there. So ⁓ anyway, there you go. That is how you set the resource metadata URL ⁓ in a www authenticate header for

Harsh Bharadwaaj (00:02:16)  
your MCP resource server.

—-----

122

Harsh Bharadwaaj (00:00:00)  
You know what? Let's go ahead and go for these bonuses. I'm gonna create ⁓ export ⁓ a type called authinfo ⁓ that extends the auth info from the SDK. So ⁓ let's bring in the auth info type, and we're gonna call this the SDK auth info. ⁓ from here, ⁓ and we'll say SDK auth info ⁓ and add to that ⁓ the ⁓

user ID on the extra property. So we'll have all these things, but we'll also have that extra property. ⁓ And this is not too happy with the auth info. So let's find out where that auth info comes from. Let me show you how I do that. ⁓ Let's get rid of that. And then we'll say auth info. ⁓ And we're ⁓ here let's see type ⁓ A equals ⁓ auth info. Here we go. See now it's autocompleting ⁓ and I can bring that in. There we go. Okay. And we'll call that ⁓ SDK auth info. There we go.

⁓ Great, so we've got our our auth info type. Benefit of doing this is the SDK auth info does ⁓ not have anything specific for the extra property, and we are going to be adding the user ID to our extra property. So I want to have that. ⁓ But we also want to maintain ⁓ consistency with the rest of the SDK because it makes it easier to integrate with other things. So we're that's what we're going to need to return from our resolve function.

Okay, great. And then let's also create a Zod schema to parse the introspection ⁓ return value. So let's get our ⁓ yeah, introspect response works just fine. Let's bring in Zod. ⁓ And we're going to expect a client ID, a scope, and a sub. It actually sends back more, and we can add a console log to see what it sends back. ⁓ but this is all that we need right now. So we're just gonna expect those fields. ⁓ and the reason we need these is because auth info.

expects we need to have at least token, client, ID, and scopes. We can also add or expires at and resource, but we're not gonna get into that right now. ⁓ So with that now we can make ⁓ export a function, an async function called resolve auth info that takes a request. ⁓ And yeah it's ⁓ the AI is doing like pretty okay so let's go and review. ⁓ first it's getting the auth header. Perfect. ⁓ if the auth header starts with bearer,

Harsh Bharadwaaj (00:02:24)  
that works fine. ⁓ you ⁓ technically a lowercase b is okay as well. So let's switch this ⁓ to be ⁓ bearer ⁓ as a ⁓ we don't want starts with okay, let's see. ⁓ yes, that's right, I am cheating. I'm referencing my ⁓ here we go. Token. I'm referencing my notes. ⁓ auth header header dot replace.

There we go. ⁓ There we go. That's what we're looking for. So we're replacing the bearer portion. The rest of it is going to appear ⁓ as part of the string. So ⁓ the auth header might look like something like that. There you go. It'll look like that. ⁓ in this case, yeah, it's either it's not undefined, it's gonna be null. ⁓ but it's either a string that starts with a bearer, and it could technically start with a lowercase b. It normally doesn't. ⁓ but yeah, it could be uppercase bearer, and then the token, ⁓ or ⁓ technically null.

because that's how headers work. But there we go. ⁓ So now ⁓ now that we have the token, everything else in the string was replaced, ⁓ we're gonna say if there's no token, then resolve auth info returns null. There was no token. ⁓ Not going to try and introspect something that doesn't have a token. ⁓ Now we go to the validate URL and ⁓ per our instructions ⁓ we were given that the validate URL is at slash OAuth slash introspection.

And so that's what where we're ⁓ making that URL. It's on the EpicMe auth server URL. ⁓ we make a fetch to there. We're gonna ⁓ do a post with the headers content type is ⁓ form URL encoded. This is a pretty standard for ⁓ introspection APIs. That's typically what they're going to expect. ⁓ not always, but so read the docs on whatever ⁓ authorization server you're using. ⁓ And ⁓ yeah, then we send the token as the body.

If the respen response comes back not okay, there was like a 500 error or something. ⁓ You might want to add some additional error handling here ⁓ so that you can send back a r a more helpful error message or something. But we're right now just gonna keep it simple. Say return null, we weren't able to get ⁓ a token or resolve this token into the auth info. ⁓ Assuming that everything was okay, we're gonna take that data from the JSON that we're expecting.

Harsh Bharadwaaj (00:04:49)  
We're going to use the introspection response schema and parse that data. ⁓ And that's going to get us our client ID, our scope, and our sub, which is the subscriber. I don't think that the term sub is short for subscriber. I'm not actually sure what it's sh short for. ⁓ But ⁓ the ⁓ it's ⁓ what this token represents, who what account this represents, or who owns the token. ⁓ and then we'll stick that in our user ID.

Now if we go back to our SDK ⁓ that's expecting token, client ID, and scopes, ⁓ our AI was ⁓ kind enough to say client ID scope, ⁓ but it's not giving us ⁓ it's it gave us sub on there, ⁓ but we don't need that. Remember, we are looking for token, client ID, and scope. So we need to handle the scopes ⁓ or we need to include the token ⁓ itself, which we have right here. So we'll stick token right there.

we don't need to include the sub there, that's fine. And then let's also make this a little ti happier with the types validate our return type. We can do that in one of two ways. We can either say satisfies auth info, and now see there we go. That's kind of why I wanted to do this. ⁓ so it should be client id is the client ID, and then the scopes should split. ⁓ or we could say this returns a promise that is either the auth info or null. So you can do it either way.

I like satisfies a little bit because it feels satisfactory, ⁓ but ⁓ yeah, ⁓ however you want to do it. ⁓ Okay, so that's our auth info resolution. So now that we have this all implemented, we can come back to our index ⁓ and we can call that ⁓ right here. So ⁓ yep, there we are. ⁓ so we determine whether we have the auth header. If we don't have it at all, you gotta authorize. If we do have it, we're gonna make sure that it does represent ⁓ auth info.

And otherwise we're gonna handle unauthorized again. You're still not authorized. You have a bad token or something. And we'll take care of a better error message for that later. ⁓ But from there, if you've got it and it it's valid, then you've got auth info and we'll be able to provide that to our MCP server, which we'll do a little bit later too. ⁓ Let's go ahead and add a console log of the auth info. ⁓ And then we can come over here, we can run through connect. ⁓ and you know what? I actually practiced this earlier.

Harsh Bharadwaaj (00:07:09)  
And we already have auth state. So let's clear that out. ⁓ Let's disconnect and connect again. There we go. ⁓ And I'll click this, continue auth flow, and then come back over here. There we go. We've got our auth info. Here's our token, our client ID, here are the scopes we've got, ⁓ and extra. And if you want to play around, you could ⁓ say, okay, let's clear the auth info. ⁓ and we'll go ⁓ reconnect ⁓ and let's ⁓ remove all scopes and see that that ⁓ logs out.

Scopes are empty. Hooray\! It does work. ⁓ So ⁓ let's just review this really quick and then we can move on. So the first thing that we did was we created our special auth info type based off of the SDK auth info where we could specify: hey, we have an extra property with the user ID. ⁓ And then we have our introspect response schema that's got a client ID, the scope, and the sub. That's what we're expecting as a response from the introspection endpoint.

we grab the token from the auth header. If there's no token, we return null. We ⁓ build up our validate URL and post to it with ⁓ URL ⁓ or a form URL encoded. ⁓ we provide that as URL search params as our body with the token. ⁓ If the response is bad, then we just say ⁓ return null. So response.okay would be false if it's a non 200 or non-300 ⁓ response, so 400 or 500 response.

and then if we're good, then we'll parse the JSON data, which is what we're expecting. I suppose you could add some extra error handling. Maybe it didn't come back with the JSON response. And that would be unexpected, and so maybe handle that error ⁓ a little bit better. ⁓ and then we're gonna parse the ⁓ data response to get the type safe ⁓ version of that JSON data and make sure that it returned what we expected. If it didn't, then things are really bad and we should probably throw an error. ⁓ We get the client ID scope.

And sub and then we have ⁓ we turn that into our auth info. One other thing I want to do just for the fun of it is ⁓ let's log the data. What what does this introspect endpoint give us? So let's ⁓ clear our auth info right here. ⁓ query that and reconnect. ⁓ And here, let's do Olivia this time. Ta-da\! ⁓ And here we are. ⁓ if ⁓ no, it's right here. so this includes active true, client ID.

Harsh Bharadwaaj (00:09:35)  
Scope, sub, exp, ⁓ and issued at IAT. ⁓ So ⁓ that's spoiler alert on active true. We'll get to that here in a little bit. Some of these things ⁓ we don't really need to worry too much about, especially like expires at. That's gonna be handled by the client mostly. ⁓ the client is gonna handle the refresh token and all that stuff. So we don't really need to worry so much about ⁓ expiration or issued at typically ⁓ on our side of things. We do care about ⁓ active true, ⁓ though.

Yeah, we'll we'll talk about that in a little bit. ⁓ but mostly ⁓ and and actually the client ID doesn't ⁓ typically matter so much for resource servers either. ⁓ It's mostly the scope and the sub ⁓ that matter most ⁓ from introspections. There are other properties that you could have here. AUD is audience, ⁓ also not typically ⁓ necessary for ⁓ the ⁓ MCP servers that you're gonna be building. ⁓ So we're just focused on scope and sub right now. ⁓ All right.

I think you did a good job with that one. So pat yourself on the back. See ya.

—----------

106

Harsh Bharadwaaj (00:00:00)  
Welcome to this exercise. We're gonna get some metadata going on here. So, ⁓ when a client wants to connect to a protected resource like our MCP server, ⁓ it needs to be informed of where to get the authorization tokens. ⁓ And that is what we're going to be doing in this exercise. So here's a quick flow in a sequence diagram. The client, ⁓ this could be cloud.app or chatgpt.com, or it could be their desktop apps, or it could be just about

And Goose, VS Code, like so many clients, ⁓ one of those clients, ⁓ once the user is configured or or it has decided it wants to connect to an MCP server, ⁓ the first thing that it's going to do is ⁓ it's going to determine whether or not authentication needs to happen. And when it has determined that, it's going to say, okay, well, get me the protected resource ⁓ on this MCP server. The MCP server, ⁓ first of all, ⁓ needs to be able to accept a request like that from a client.

And this is where we're going to get into cross ⁓ cross-origin resource sharing. I always forget what CORES stands for. ⁓ But ⁓ if you've ever seen a CORES error, this is because the browser wants to keep the user safe and it doesn't want ⁓ a ⁓ a website from ⁓ one ⁓ domain to be able to request ⁓ make requests to ⁓ resources on another domain. That could be very, very bad. ⁓ like malicious site requ makes requests on your banking site, right? Like that would not be good.

So ⁓ by default, if the response from that server does not include ⁓ the certain headers for cores for cross-origin resource sharing, then it that request will be rejected by the browser ⁓ outright. And so that's one of the first things we have to do in this is set up cores, which I know is everybody's favorite thing. ⁓ but assuming that is all set up correctly, the MCP server is going to include ⁓ or send back resource metadata, which includes

Authorization servers. So it's the URL for where the authorization server is. So this is a really important point that you need to understand is that there's a difference between the server that's responsible for validating who the user is and handling their login and handling the access tokens and stuff like that. There's a difference between that server and what we call the resource server, which is ⁓ basically just accepting those access tokens ⁓ and then using that and maybe talking with the

Harsh Bharadwaaj (00:02:24)  
the back end that resolves those authorization tokens or whatever. ⁓ So there's the authorization server that generates the tokens and manages the clients and all of that stuff. That's where actually the real complexity of OAuth lives is in that authorization server. And then there's the resource server, the MCP server in our case. So this is the thing that's going to be able to accept those access tokens and then perform actions on behalf of the user ⁓ using those access tokens. And so the client only knows about the resource server.

And that's why it performs this request first, to discover, okay, your resource server, great, but who's your authorization server? Because I need to get a token. And so we're going to send back ⁓ that information, and then the client will say, Great, now I know who the authorization server is. Let me go make a request to that authorization server ⁓ to get some more information. Like, I need to know how I register myself as a client for the authorization server. ⁓ Because the authorization server isn't going to just accept ⁓ token requests from anybody.

It needs to know about the clients that are going to be asking for ⁓ for tokens. ⁓ And so ⁓ by default, that's actually ⁓ not super common for ⁓ OAuth servers to just accept ⁓ a client registration from anybody. But as part of ⁓ MCP, ⁓ the dynamic client registration is an important piece. And so when the client says, Hey authorization server, could you get me some metadata about yourself?

It is going to expect to have some information about how to register itself dynamically. ⁓ Normally you like you may have actually experienced this before if you've ever made like a a sign in with Google or sign in with you know with Apple or Microsoft or something, where you have to register your app as a client with with that service. ⁓ with MCP, the setup is a little bit more complicated because you ⁓ you could literally ha be registering with any number of clients. And so that's why ⁓

MCP spec uses dynamic client registration, which is a part of the OAuth spec. ⁓ But ⁓ you might notice with some services they don't yet support dynamic client registration, ⁓ and that's an important thing that they need to be able to support to work with MCP authorization. ⁓ Anyway, so the client is gonna request the information for like how do I authorize users, how do I register myself as a client, ⁓ and then that's gonna come back. ⁓ It's going to do the client registration flow.

Harsh Bharadwaaj (00:04:47)  
And then it's going to ⁓ perform the OAuth flow to get a token. ⁓ So that will go over to the authorization server. Presumably the user is signed in on that server. So they're gonna like open up to that page ⁓ and they'll see the authorization flow. ⁓ In our example, we actually have ⁓ a ⁓ demo app that is managing all of this ⁓ authorization server for us. ⁓ And so you don't have to build any any of this stuff.

This is ⁓ outside of the scope of this workshop is the authorization server. We're focused on the resource server. ⁓ but ⁓ this might be kind of familiar to you. You've seen this with ⁓ authorizing apps with Discord and with Twitter ⁓ or X and with ⁓ Google, and like you ha have been through pages like this before. ⁓ Ours is a little bit more complicated because I want you to be able to control your permissions, the scopes that will matter later in the workshop.

I want you to be able to choose user accounts. So we've got three user accounts in the database automatically, so you can kind of switch between them ⁓ and play around with that. ⁓ And then I also allow you to look at the search parameters and even update them if you want to. So we're doing a little bit more ⁓ in ⁓ on this page than you typically would do. And in fact, even when you select a user ⁓ that ⁓ actually does create the OWASP token and everything.

But then you have to click this continue OAuth flow to to finish that entire flow. ⁓ whoops, there we go. ⁓ and so once you finish with that whole flow, then you're gonna have the access token and everything is is gonna work just fine. So ⁓ one other thing that I'll mention about that as well is that should be running for you automatically as you run the workshop app. If it's not, then you need to ⁓ CD into the epic shop directory in the epic me directory. That's where the source code for this app is.

Feel free to look into that. You need to have that running. ⁓ It should definitely be running as you're running the Workshop app, though. So you shouldn't have to have any trouble with that. ⁓ It will run as a sidecar process. ⁓ But anyway, you can go to the homepage and you can see ⁓ the users and their data. And then you can also see ⁓ some of the back-end information around the auth tokens that are being generated. So you'll see here are the dynamic client registrations that were made, so the different clients.

Harsh Bharadwaaj (00:07:06)  
Here are the grants that's part of the OAuth token flow process. And then here is when that token has been resolved into, or that ⁓ that code, that auth code, is resolved into a token. Then you can see ⁓ those tokens and who they represent and all of that stuff here as well. There's a lot of complexity with the authorization server, and that's like a like two-week workshop by itself building an authorization server from scratch. ⁓ I did not build this one from scratch. I'm using a library ⁓ from Cloudflare.

for the ⁓ OAuth provider in a Cloudflare worker. That's what I use for my own website and for its auth flow. So ⁓ it's not terribly complicated if you're using a library like that, or there's also better auth. ⁓ but ⁓ there there is some complexity there. And so using a third-party service for the authorization server side of things, ⁓ not terri not a terrible idea. ⁓ it it is a pretty complicated thing. ⁓ We're focused, like I said, on the MCP server, which is our resource server in this case.

So once the complicated OAuth flow has finished and the client has re ⁓ registered itself ⁓ and then received a code for the ⁓ for the user, then it ⁓ converts that code into an ⁓ OAuth token and then it can use that token. So that's where this authenticated request with ⁓ token works. The MCP server can now ⁓ r ⁓ reveal the protected resources or whatever else is going on.

it can make further downstream requests to the auth server or to any other server that can resolve those tokens ⁓ and talk to the database or whatever on behalf of the user, whatever it needs to do, because now it has the auth token to be able to do that. And that is the the basic flow of what you can expect for ⁓ this exercise. ⁓ pretty much we're focused on just the the first part of this, which is the protected resource, ⁓ the

⁓ authorization server stuff and then ⁓ cores and i said that in reverse order actually so that's what you're going to be doing in this exercise it's a really good one ⁓ i ⁓ I do kind of like it hurts me a little bit inside that we have to start with cores but it's actually a really important thing that you're absolutely gonna hit when you start implementing stuff like this. ⁓ So ⁓ that's why it's included. ⁓ And then we've got a couple other things to handle like maybe legacy clients and and different things like that. So we're gonna handle those error cases. ⁓ And when you're all finished you should be able to

Harsh Bharadwaaj (00:09:29)  
go through the auth flow on this server. ⁓ You ⁓ we'll we'll get to actually connecting to the server and and doing stuff ⁓ in a future exercise. So don't worry about being over here on the connection side. You should just be ⁓ focused on the auth settings in this exercise. ⁓ And also I'll mention that you'll want to have your ⁓ developer tools open in the network tab ⁓ and see the requests that are being made as you're going through this ⁓ because

you're gonna see a couple of ⁓ interesting requests and when you're working through it the first couple exercises, you're gonna see a bunch of errors and stuff. That's where you're gonna be fixing. ⁓ All right, I think I've talked enough. You'll have a really good time with this one. So we'll see you in the exercise.

—-------------------

125

Harsh Bharadwaaj (00:00:00)  
There's one more thing that we need to do just to cover our bases on this, and that is dealing with the active value that can come back from our introspection. ⁓ So ⁓ if a token has expired or ⁓ it's been revoked or something, the active property will be set to false, and we don't want to let that user through. Now, depending on the introspection token, ⁓ the active property will be the only thing that comes back, but it's possible that it they could include all the other properties, and then you'll just be in trouble.

So you always want to make sure you check the active property. It's pretty easy. The response, you get the JSON, check the active token. And in our case, we're just gonna return null. ⁓ However you do it, maybe you'll show a special error message or something else like that. But ⁓ for ⁓ typical OAuth ⁓ stuff, you just r ⁓ are going to ⁓ have that be included in the www authenticate header like we already have set ⁓ that we just ⁓ literally just did a second ago. ⁓ So

The other thing that we're going to do in here is make our types a little bit nicer. And so we're going to do something called a discriminated discriminated ⁓ union. ⁓ This is a TypeScript thing just to make it a little bit easier on ourselves. And I know that not all of you are super familiar with that, so I have this example for you. If the API response ha can either be a status success ⁓ with a data property or status error with a message property. And TypeScript knows how to handle this. ⁓ If the status

is success. Now inside of this if ⁓ statement ⁓ it's going to ⁓ know that response data exists ⁓ and inside of here it knows that response message exists. But I couldn't move this up here. I would get a type error if I did that. So that's the beauty of discriminated unions. They're quite nice. ⁓ And you can do that with Zod. ⁓ So this is a bit of a hint for you ⁓ with ⁓ using the discriminated union you say what property is the discriminated union.

And then you use a z dot literal. Ours is not a string, so it there's a little bit of challenge there. ⁓ But yeah. Go ahead and give that a whirl. It's really quick, just one line change for ⁓ the check on active and then a little bit for the introspection ⁓ schema if you want to play around with that. Okay, have fun.

—------------------

112

Harsh Bharadwaaj (00:00:00)  
Let's start here in our handler and we'll handle any request that is made to our protected resource. So it ends in slash MCP because that is the resource that we're pointing clients to is slash MCP. So if this was ⁓ anything else, then you would do ⁓ anything else ⁓ here. ⁓ right. So that's the resource that we're protecting. That's the one that we're pointing these clients to. So MCP, there we go.

And so that's gonna handle our protected resource. In this case, we actually do want the request. So let's come up here, we'll add our import, and we will come over to our auth module to export that. So we've got our handle OAuth protected request. Yep, there we go. Thank you, ⁓ AI assistant. It accepts a request, and then we're gonna get the resource URL. So this is the URL for which we want to ⁓ protect or or the the resource that we are protecting. ⁓ and so that is the URL.

That we're protecting. It's the slash MCP ⁓ on this domain. ⁓ And so this this is going if you're not familiar with this, this is basically gonna be ⁓ that. That's what we're constructing. So it request.url ⁓ will actually be ⁓ this whole URL by the time it gets here. ⁓ So request URL ⁓ is ⁓ gonna be ⁓ that. There we go.

That is a request URL right there. ⁓ And using new URL slash MCP request URL, it's going to just ⁓ grab the domain ⁓ and or the the host name here, ⁓ the host, and then it's going to swap out the path name for this first argument. So that's not really common. That's why I thought I'd take a second to explain what was going on here. So this is what u resource URL will be set to. So that's one of the things we need to return. ⁓ we are not actually gonna fetch this because that is l like that would lead.

us to right here, which is not what we want to do at all. We are not going to fetch it. Instead, we're going to return response.json that includes the resource, ⁓ and that's our resource URL. That's why we constructed this in the first place. ⁓ And authorization servers. ⁓ And that can be an array of multiple ⁓ options. That's not very common as far as I'm aware to actually have multiple authorization servers, but it is possible. ⁓ We are going to point to our EpicMe ⁓ auth server.

Harsh Bharadwaaj (00:02:24)  
And once it gets this metadata, then it's not going to make this request anymore, ⁓ not to our server anymore. It will make that request to the auth server. And we can see that happen in real time. So let's just try that out. ⁓ And instead of doing the quick OAuth well, we'll do the guided one. So we'll hit continue for the metadata discovery. ⁓ Looks like maybe something broke, so let me just get that running again.

Here we go. Starting that up. Refresh this and we can go to auth settings, guided auth flow. Let's go continue. Boom. We got our OAuth metadata. So first it made a request to the protected resource. ⁓ Interesting, it doesn't add the slash MCP here. So let's just double check that the right requests are being made. So let's refresh the page, close this out, go to ⁓ OAuth settings, do guided auth flow, we'll do continue here.

And you'll notice the request starts out with slash MCP. You can see the preview. This is what we returned. So that's exactly what we want. ⁓ And see, what did I tell you? The resource is what I said it would be, even with that URL nonsense. ⁓ But yeah, then we go here and you'll notice this is a request to our authorization server, the 7788\. And so that is now ⁓ working as expected. ⁓ And the response is what we saw in previous steps. So that is all working.

Perfectly. And again, all of this ⁓ supporting cores and everything. ⁓ So now we aren't going to get errors for clients that are up to date and making ⁓ the protected resource requests properly. And ⁓ they're going to be able to identify who the authorization server actually is and make that request directly to the authorization server. ⁓ Not treat us like an authorization server. ⁓ With this, it gets the registration endpoint so it can register itself as a client with our authorization server.

Now the auth server knows that this is a client that wants to get tokens for users. ⁓ We can ⁓ create that authorization URL. We can go through this flow. We can select our scopes and select the user we want to select. ⁓ That's not typically what you would do in an ⁓ if you were building an authorization server. Normally users have to log in. ⁓ But we're making it easy for the demo purposes. So we select this one, we continue the auth flow, and now we can grab that and paste this into the debugger.

Harsh Bharadwaaj (00:04:49)  
And continue, we get the token request. Continue with that. And that request is successful. ⁓ That token request, of course, happening on the authorization server side. And that's why we didn't have to build that f as the resource server. ⁓ And then the authentication complete. We get our access token. ⁓ So there you go. We have just improved the the flow of things ⁓ by adding the protected resource request handler.

all we did was add a handler for the protected resource URL. And then we handle the OAuth protected resource request. ⁓ And that ⁓ is success. Well done.

—-----------

139

Harsh Bharadwaaj (00:00:00)  
Okay, this is gonna be the easiest exercise of them all because you're literally making a single line change. ⁓ so I will be brief. ⁓ you can add scope supported to the protected resource. ⁓ And that's nice because clients may then just include all of those scopes ⁓ in the request or the authorization request, or they might do something else that is useful for it. But because we can't send back our scope auth param on the 403 forbidden.

This is at least some way to give kind of a hint to the client what scopes this MCP server actually supports and cares about ⁓ so that it can make the right request for the right amount of scopes in in most cases. So ⁓ go ahead and add that one line and I'll see you in just a second.

—--------------  
137

Harsh Bharadwaaj (00:00:00)  
So we've got another problem with scopes. Let's connect. And this time I'm going to uncheck all of the scopes. ⁓ No scopes for you. ⁓ And we're gonna say ⁓ login as Kelly, continue, ⁓ and we are now authenticated. ⁓ But we have no scopes. ⁓ So we can literally do nothing. There is nothing useful that we can do because all of these have now I can ping the server. That's nice, but all things have now been protected against us because we don't we we didn't give enough permissions.

And so now the user has to go in, they have to ⁓ okay, gotta clear my auth state. They don't even know what that means, ⁓ and reconnect so that they can get the right scopes. ⁓ So your job in this exercise is to ⁓ make it so that ⁓ they find out right away as they're connecting that they have insufficient scopes. And then the client can say, ⁓ I need to go get some more scopes. because there's no reason that they should ⁓ have to like

discover on their own that they can't actually do anything useful. ⁓ The interesting thing about this is that you could actually have a number of combinations of scopes that would technically be allowed and and make the server useful to you. And so we can't say, hey, you have to have all the scopes, because that that doesn't make any sense. ⁓ You don't necessarily have to have all the scopes. ⁓ You just have to have any number of combinations of scopes. In our case, ⁓ y any one of those scopes, here let's go connect. Any one of these

Is going to be fine and give you something useful, except for the right user. ⁓ But ⁓ in some situations, maybe you would need both read entries and read tags to do anything useful. ⁓ And so what I am doing is I actually ⁓ have a ⁓ an array of arrays that is all of the combinations of scopes that would be valid. ⁓ And then I include that in the error description. Typically you would do this.

With ⁓ a scope auth param on your www authenticate header, ⁓ but we can't do that because ⁓ there's not any single ⁓ set of scopes that would work. We have multiple combinations. So we do the best we can with the error description ⁓ and hope that the client makes the right decision, not just the client, but also the user as they're going through and selecting their permissions. ⁓ So with that, ⁓ I hope that you have a good time with this exercise. Let's go ahead and make sure users don't make

Harsh Bharadwaaj (00:02:22)  
Big mistakes. ⁓

—----------------

129

Harsh Bharadwaaj (00:00:00)  
So now that we have the auth info, what we need to do is get it into our MCP so that we can access the user's information and stuff. ⁓ And our ⁓ database client right now is not working at all. In fact, if you were to try to perform any kind of action, you may have tried this already, it maybe not, but if you try to ⁓ maybe list entries, you're gonna get invalid token here.

not because we have an invalid token, but because we're not passing that token to our database client. And the way that we do that ⁓ is ⁓ we update the get client to accept the user's OAuth token. ⁓ And the abstractions for making those database calls and stuff that's outside of the scope of this workshop and so that's already built for you. You just have to get that token into that client and then it can make authenticated requests. Feel free to dig into the implementation details if you can

really care but everybody's situation is gonna be a little bit different. The point is you need to be able to pass that auth token to wherever ⁓ n needs it for the request. So maybe you've got downstream APIs that you're calling. That's what we're doing in this situation. ⁓ Or maybe you're just talking to the database directly right in your MCP. Also a perfectly good option. ⁓ but whatever the case is, the the thing that talks to the database needs to have your token. ⁓ and it needs to ⁓

Be able to find out who that user is. So ⁓ that's one thing that you're going to be doing in here is enhancing our client to be able to handle that. And then the other thing that you need to do is get the token to that ⁓ part of the code. And that part of the code is inside of our MCP agent class. ⁓ And so to do that, ⁓ we communicate through props, not React props, different kind of props. ⁓ So we're going to define our state and our props. We'll pass those to

our MCP agent implementation. So here's my MCP. inside of our init then we'll have our props. ⁓ To out of an abundance of caution, ⁓ this ⁓ props property ⁓ is possibly undefined. And so you have to do some undefined checking ⁓ to to check for that. ⁓ and then inside of our ⁓ handler that has the request, this is where we're doing the resolve auth info and stuff. You're going to add ⁓ to the context props ⁓ our ⁓

Harsh Bharadwaaj (00:02:19)  
specific auth ⁓ info ⁓ or or token stuff there. So ⁓ that is your job in this exercise. You're gonna be doing a couple of typing related things. You're gonna be doing some prop passing ⁓ and ⁓ and yeah, and then initializing the client. So should be pretty quick, but I think that you will enjoy it. So have a good time. And actually, yeah, when you're finished with this step of the exercise, you should be able to start making ⁓ tool calls. So that'll be nice.

Have fun.

—---------------------  
105  
Harsh Bharadwaaj (00:00:00)  
Hey, welcome\! My name is Kent C to odds, and I am so excited to go through the MCP Auth workshop with you. ⁓ This is fantastic because most of the things that we're building are kind of behind a login. Like you need to know who the user is and make sure that their data is segmented from somebody else's data, so not everything ⁓ is public. ⁓ And this is really powerful because this like this is where things kind of get magical, and the user feels like the natural language model understands them and knows knows what.

Their specific needs and requirements are. ⁓ And so this Auth workshop is going through the OAuth 2 spec for the resource server side of things. ⁓ So the MCP spec has ⁓ authentication laid out where the MCP server represents a resource server, and then you can have an authorization server over here. ⁓ Luckily for us, the authorization server is the really complicated part of OAuth, ⁓ and you probably want to either use

A an existing library to handle all of that, or a third-party service, or maybe you're at a big company and you've got an entire team dedicated to just OAuth because it is a really complicated subject. ⁓ But the resource server side of it is actually ⁓ quite straightforward. ⁓ And so ⁓ that's that's good for us. So as we go through this, we're going to be talking about the differences between the authorization server and the resource server. ⁓ and we're gonna just make our MCP server use as much of the

⁓ OAuth 2.1 spec as it can. We're gonna be using the MCP inspector because it has all the authentication stuff built in. And it will give us a really good foundation for making things work in any client that the user is using, whether it be ChatGPT or Claude or Cursor, VS Code, or whatever. ⁓ And so it's very exciting stuff that's going on here. We're gonna talk about how to do the metadata discovery and how to in fact we're even gonna do some backward compatibility stuff where

Before the authorization spec required the MCP server be the authorization server. So things were a lot harder back then. ⁓ but yeah, so we're gonna do some backward compatibility stuff for that. We're going to handle initializing the OAuth flow with the www authenticate ⁓ header and 401 and 403s. We're going to handle getting that auth info, ⁓ resolving that token into the user object, and then getting that into

Harsh Bharadwaaj (00:02:19)  
our ⁓ Cloudflare worker and our our ⁓ durable object which is our agent and that's where our MCP stuff is living is inside of that. And then we're even gonna get into scopes and ⁓ protecting certain parts of the application based off of what the user said that they're cool with us actually ⁓ using ⁓ as a part of this MCP integration. ⁓ So I'm really excited about this. I think that you're going to love this and when you're all done you'll have a lot of ⁓ really critical knowledge

For you to build really excellent experiences for your users. So, what are we waiting for? Let's get into it.

—----------------------

118

Harsh Bharadwaaj (00:00:00)  
So when that request comes into MCP, then we need to check whether it's authorized. If it's not, then we shouldn't be sending it to our MCP server. We should instead send it a 401 response. So let's do that. If ⁓ the request headers ⁓ get authorization. So if it does not have an authorization header. ⁓ And in fact, I'm gonna give this ⁓ a auth header variable name, just for the fun of it. ⁓ And so if there's no auth header.

⁓ then we want to handle unauthorized. And with that we don't actually need the request. We're going to come up here, bring in handle unauthorized, ⁓ and we will export a function, handle unauthorized. ⁓ and it's pretty simple actually. So we've got ⁓ a new response, status 401\. Here's our headers, a w ⁓ w ⁓ authenticate header with the bearer being the realm. So the bearer realm.

equals Epic Me is just indicating the the domain of this ⁓ this resource that they're trying to access. And so we call it EpicMe, you pretty much just call it the name of the app. ⁓ And that's all. That's it. Like ⁓ you ⁓ really just check, hey, do you have an auth header? You don't? Then you're not authorized. You need to go get an auth header first. Here's ⁓ and ⁓ we're not actually in telling it any more than that. If from there

Most clients are going to know, okay, let me do the well-known process. That's what makes it well-known, as most clients know it. Haha. ⁓ So with that, now we can hit connect, and boom, now we're going through the OAuth flow. ⁓ So there you go. That is a standard response to an unauthorized request ⁓ for an OAuth authenticated resource.

—-----------------

141

Harsh Bharadwaaj (00:00:00)  
Alright, break time, get a drink, go tell somebody something nice, write your mom a letter or something, ⁓ and ⁓ here's a joke for you. I used to work at a stationery store, but I didn't feel like I was going anywhere. ⁓ So I got a job at a travel agency, now I'm going places. ⁓ Or I know I'll be going places. ⁓ Alright, there you go. ⁓ chuckles are good for the soul. ⁓ get a drink of water, ⁓ write down what you learned that's important for your attention, ⁓ and ⁓ yeah, have a nice break.

—------------------

113

Harsh Bharadwaaj (00:00:00)  
Ben and Jerry's really need to improve their operation. The only way to get there is down a rocky road. Ha ha ha. ⁓ that is definitely a dad joke. ⁓ all right. Great job. This has ⁓ been a good exercise. Now it's time for you to write down what you learned so you don't forget it. This really important process for retention is writing down what you learned. So go ahead and write it down, and when you're all finished with that, you can come on back and we can learn more. ⁓ actually, before you do that, you should probably take a drink of water.

because that's important for your body to function. And moving your body also is important for your brain. Your brain needs blood flow, you'll remember things better, have a better experience. So make sure that you're moving around ⁓ and we'll see you when you're ready.

—------------------

111

Harsh Bharadwaaj (00:00:00)  
While it is technically neat that we can do the whole auth flow already, we should probably handle that request that it's trying to make right here to get our protected resources. ⁓ in part because it's telling us right here in the metadata discovery that ⁓ there's a problem with resource metadata. ⁓ so it does not implement OAuth 2.0 protected resource metadata. And while that's not technically required.

For the MCP inspector, it could be required for other clients that you want to support. And so that's your job in this exercise is to implement ⁓ this endpoint so that it doesn't return not found, but it ⁓ instead returns the proper metadata to describe this resource. There's not a whole lot in there, and so the emoji are there to guide you. Go ahead and give it a whirl. And when you're done, you should be able to still do the OAuth flow, but you won't get this warning anymore. So have a good time.  
—---------------

120

Harsh Bharadwaaj (00:00:00)  
Let's get started by grabbing the auth header ⁓ and determining whether the user has an auth header. Our AI assistant here wants to get the auth header and check whether it equals null. That works fine, but there's actually a has method that you can call directly on the headers, and that just returns a Boolean whether it has it. So that's even better. ⁓ So now we know whether we have the auth header. So we can come down here and update this. Now I'm

I'm gonna change this a little bit because I don't want it just to be like some massive string. Like that's no ⁓ look at that. Like, no, no, no, no, we're not gonna do that. So instead we're gonna turn this into an array, ⁓ which we will join with commas. ⁓ And then we'll just do this and we'll do that. There we go. ⁓ And then we can say, hey, if you've got an auth header, then ⁓ error invalid token, otherwise ⁓ and if you have an auth header, error description. ⁓

And here we're saying undefined. I like to say null, ⁓ but we do need to get rid of those. We don't want to join null. So we'll do this filter Boolean trick, which basically just filters out falsy values, which null is a falsy value. So that will filter out those, then we'll end up with an array of just two items that'll be joined, ⁓ or we'll end up with an array of four items, and those will be joined by the comma. So ⁓ we should have a proper ⁓ www authenticate header.

That gives the client more information about ⁓ why they're getting this 401 response. So well done.

—---------------  
115

Harsh Bharadwaaj (00:00:00)  
So we've got the OAuth flow working right now ⁓ if we go through the OAuth settings, but we need to tell the server that they need to authenticate if they want to access anything. So if we would go to Connect, which is effectively what you would do if you were adding an MCP server to any of the clients, ⁓ it's gonna just let them on through. And ⁓ then if you try to list tools and do anything, like here, let's go list our entries, we're gonna get invalid token. That's not a great experience. You know what's better?

is if we click connect and it takes us through the OAuth flow. ⁓ And so we need to tell it when that first request comes in that hey, you actually need to be authenticated to access this server. ⁓ And so that's what you're going to be responsible for doing in this exercise. It's actually not a lot of code. It's all nice and well standardized and everything. ⁓ But you do need to first really check for the authorization header and then give a response ⁓ that is one of the standard OAuth responses.

if they are unauthorized. If that's the case, then the client will know, okay, unauthorized header. Here, let me look at the WWW authenticate header. That has the information I need. Boom, I'll send them over ⁓ through the OAuth flow and then it gets the metadata and everything else that we've done already. ⁓ So that is your job in this exercise. Get the O the auth header ⁓ and set the the or ⁓ if it's not there, then send a 401 response ⁓ with the proper headers, and then it should kick them into the OAuth flow.

Have a good time.

—-------------

114

Harsh Bharadwaaj (00:00:00)  
So we've got our resource server set up to handle an auth flow, but we need to trigger that auth flow. We need to initialize it. When a request is made that requires authorization and there's no authorization header, we need to tell the ⁓ requester, hey, you need to go get an authorization header. That's what this exercise is all about. So here's what this might look like. We've got our MCP client in our server. There's a request to slash MCP with no auth header.

The server needs to send back a 401 unauthorized response with the www authenticate header, and that's going to have a number of auth params that we'll talk about. ⁓ And then the MCP client ⁓ can derive from that where to get the metadata, and we get the metadata back. So that's what kicks off that OAuth flow. ⁓ So ⁓ it's basically like this: check if there's an auth header. If there's not one, send a 401 with the authenticate header. ⁓

The ⁓ interesting thing about the authenticate header is that it can take a number of what are called auth params. So there's different types of auth headers, and then we've got ⁓ we're gonna use bearer, we've got realm error, error description, ⁓ and resource metadata. We're gonna focus on the realm and the resource metadata. ⁓ error and error description we'll get into a little bit later. ⁓ the resource metadata for us is is just that same well-known OAuth protected resource that ⁓ our

Client just happens to guess ⁓ is ⁓ where to find that protected resource. And that is a well-known address for it, but we can be more explicit about it. So that is your job in this exercise is to kick off the auth flow so that ⁓ the server knows right when it tries to make a request that, okay, I gotta go through the auth flow. ⁓ So have a good time with this one.

—--------------

140

Harsh Bharadwaaj (00:00:00)  
So we've got our resource server set up to handle an auth flow, but we need to trigger that auth flow. We need to initialize it. When a request is made that requires authorization and there's no authorization header, we need to tell the ⁓ requester, hey, you need to go get an authorization header. That's what this exercise is all about. So here's what this might look like. We've got our MCP client in our server. There's a request to slash MCP with no auth header.

The server needs to send back a 401 unauthorized response with the www authenticate header, and that's going to have a number of auth params that we'll talk about. ⁓ And then the MCP client ⁓ can derive from that where to get the metadata, and we get the metadata back. So that's what kicks off that OAuth flow. ⁓ So ⁓ it's basically like this: check if there's an auth header. If there's not one, send a 401 with the authenticate header. ⁓

The ⁓ interesting thing about the authenticate header is that it can take a number of what are called auth params. So there's different types of auth headers, and then we've got ⁓ we're gonna use bearer, we've got realm error, error description, ⁓ and resource metadata. We're gonna focus on the realm and the resource metadata. ⁓ error and error description we'll get into a little bit later. ⁓ the resource metadata for us is is just that same well-known OAuth protected resource that ⁓ our

Client just happens to guess ⁓ is ⁓ where to find that protected resource. And that is a well-known address for it, but we can be more explicit about it. So that is your job in this exercise is to kick off the auth flow so that ⁓ the server knows right when it tries to make a request that, okay, I gotta go through the auth flow. ⁓ So have a good time with this one.

—------------------------

117

Harsh Bharadwaaj (00:00:00)  
Let's go ahead and open up the console here. We'll go to our network tab ⁓ and just clear out everything. ⁓ And see what requests are being made when we hit connect. I'm gonna have preserve log on here because it's gonna redirect us as soon as I do that. Okay, great. So now we're over on on the authorized server. But what happened here ⁓ was first it did a health ⁓ endpoint request. That's ⁓ not important here. ⁓ and then it made this request. ⁓

To to connect. That's all ⁓ implementation detail stuff in there as well. The thing that I want to point out specifically is that it made this well-known OAuth protected resource request. ⁓ And the reason it did that was to get metadata about ⁓ how to authorize ⁓ this server. ⁓ It did that just because it kind of knows or it assumed where that ⁓ that well-known protected resource URL would be.

it just made a guess. ⁓ And most clients probably are going to guess that right, but not all. And so being able to supply exactly, hey, y ⁓ when when we send that 401 response, we can say, hey, this is an unauthorized request. You're not allowed to get here. But also we can tell it where to go to get this metadata, where that resource metadata URL is. And so that's your job in this exercise. It's pretty quick, but it is important, especially to make sure that you have good compatibility with all kinds of clients.

That may not want to do a guess like this, but want to be more specifically directed. ⁓ This is also really helpful if you for some reason cannot follow the standard ⁓ process with where you locate your ⁓ metadata about your ⁓ authorization server and stuff, ⁓ and where your protected resources. And so you can be very specific about, hey, I want you to go here. So ⁓ this should be pretty quick. Have a good time with this one.

—-----------------

121

Harsh Bharadwaaj (00:00:00)  
Now that we have an authorization token, we have to actually verify that it is a real authorization token, not something that somebody just made up, ⁓ and that it it's actually currently active. they're not making a request with an outdated or an expired token. There are a couple of ways to do this. ⁓ A really popular and common way to do this is with a JWT, or that's pronounced JOT. I didn't make that up, that's really what it's pronounced. ⁓

But ⁓ a J JWT is a special type of token that is serialized and encoded in such a way that you can decode it to get the information that it represents. Things like the user ID and when it expires, ⁓ things like that. ⁓ But it is tamper-proof because it comes along with a signature ⁓ that you can then ⁓ verify without necessarily making a request to the authorization server. ⁓ This is a pretty common approach, but another approach that we're going to be doing is called introspection.

Where the authorization server exposes an endpoint, you pass the token to that endpoint, and then it returns that information to you, who the user is, the when it was issued, who it was issued to, like which client ID, all of that stuff. ⁓ And and more importantly, ⁓ is it still active? ⁓ And so ⁓ we're going to go with the introspection route, and this is a sequence diagram of how that works. So the client is gonna post to the MCP ⁓ endpoint with the authorization bearer token.

Your MCP server is going to ⁓ post to the auth server at whatever ⁓ endpoint that it has set up for introspection. ⁓ And the auth server is going to respond with two different types of responses. If the token is active, it'll return the sub, which is ⁓ the ⁓ for in our case it's the user ID, but it this is going to be ⁓ the ⁓ if essentially the ID of ⁓ the account that it the token represents. ⁓ The client ID ⁓

the scope that ⁓ this has, those the permissions, and a number of other things, audience, stuff like that. ⁓ and then that will come back to ⁓ the MCP client. well the server will get that, it'll do whatever it needs to with it and then send the the proper response. ⁓ If it's not valid, it's in expired or something, then we return an error with the inactive status. And we'll get to how we go about doing that ⁓ later. ⁓ So your job in this exercise ⁓ is to implement this POST request.

Harsh Bharadwaaj (00:02:21)  
And ⁓ when you're all finished, there's not actually gonna be any result ⁓ that is different in the inspector or anything. ⁓ you will be able to hit connect, you'll go through this, you'll ⁓ go through the auth flow, ⁓ and that's it. ⁓ you'll probably wanna add a console log just to make sure that you're getting that auth info. ⁓ and then we'll do stuff with that auth info later. So yeah, let's get to it.

—-----------

130

Harsh Bharadwaaj (00:00:00)  
Let's get started in our index. So, right up here, we're gonna grab the auth info type because we're going to need that for our props. So we've got our state. We don't have anything we're storing in our ⁓ durable object state. ⁓ And then we've got our props with the auth info. Then we gotta add that to the generic on our MCP agent. ⁓ And then we need to ⁓ pass this.props.auth token to ⁓ our get client. So this props authoken.

this is getting a red underline for kind of a couple of reasons. ⁓ one, the ⁓ client doesn't support the auth token yet. So let's grab that, stick the auth token parameter in there, and then our db client can accept the URL for our ⁓ auth server ⁓ and the auth token. ⁓ And with that saved, ⁓ we are now not getting any underlines. ⁓ this can technically be undefined at this point.

But if it is, that's like a pretty big problem. We shouldn't be able to get to the point where this dot props is not defined. So what I'm going to do is add a require auth info ⁓ function. It's not actually going to return a promise or require a request. Instead, what it's going to do ⁓ is we're going to s ⁓ get our props or our ⁓ yeah, auth info from this.props, ⁓ or that'll be ⁓ an empty object.

Then we're gonna add invariant, auth info not found. Let's bring in ⁓ invariant, and then we'll return the auth info. And what this does is it ⁓ ensures that we never get into a situation where we are trying to get auth info, but we don't have it. That would be a really bad situation. ⁓ so we're just gonna make that impossible by adding this invariant there. ⁓ Now, ⁓ we actually technically are in that situation. So if we were to try connecting right now, we're gonna get an error.

You look at the logs. Let's see. Let's look at the logs. Yep, there it is. Initialization fair failed. Auth info not found error. ⁓ So let's fix that. We'll come back over here and we'll pass that to ⁓ the props. And with that now, we can connect ⁓ and we can go to tools and list tools. Let's list our entries. Ta-da\! We've got entries. Hooray\! That's fantastic. And here's the real test. Finally, after all of this, can we get user-specific info?

Harsh Bharadwaaj (00:02:23)  
That's the question. So are these entries actually Cody's entries? ⁓ We could check with seven seven eight eight. If we look at Cody, first day at Epic Web. First day at Epic Web. There it is. Okay, awesome. Now here, let's try this. What if we want to come over to our auth? We're going to clear auth state, reconnect. We're gonna go through here. Let's ⁓ grab Olivia, continue, ⁓ and come over to Tools, List Tools, List Entries, boom, Olivia has Night Watch and

In the forest. that's fun. ⁓ Localhost77. Whoops, ⁓ eight, eight. Is that Olivia's post? ⁓ Yes, it is. So we are set. We can finally actually make requests on behalf of our users inside of our MCP. And what that took for Cloudflare specifically was adding the auth info to our props before we do the fetch on our MCP. And then ⁓ we also added some types to make this all nice and type safe.

⁓ We updated our client to accept an OAuth token so that all the requests that it's making include that OAuth token. Feel free to dive into the details if you're curious how all that works. ⁓ And ⁓ now our database is making requests that include the OAuth token. ⁓ And we made sure that we have a token by verifying that the auth info ⁓ is indeed inside of the props. And if it's not, then we throw an error. ⁓ Because that should never happen. ⁓ All right.

Awesome work on this one. We've made some really great progress. ⁓

—----------------

142

Harsh Bharadwaaj (00:00:00)  
Hey, you made it\! Good job\! We made it through the whole workshop where you learned all about authorization in the MCP spec. There's already so much that we can do, and I I can see some things that could be improved in the spec, and I'm looking forward to those improvements in the future. But right now, you can do some really awesome things with MCP, allowing users to ⁓ give access to their data in your service to their natural language ⁓ application of preference.

Which is super duper awesome because ⁓ making this user specific is what makes this so powerful. This is what's going to enable users to say, Hey, I want to do this, that, and the other, and all these services integrate that are ⁓ specific to them. It can text their mom and it can ⁓ because they love their mom and it can ⁓ send the a ⁓ a message to the school class or what you know, whatever it you need to do, not just message sending, ⁓ but whatever your service does, you probably have user specific stuff.

So this is really critical for you. ⁓ I hope you had a really good time with this, and I'm looking forward to seeing all the stuff that you build. Take care.

—-----------------

		123

Harsh Bharadwaaj (00:00:00)  
There's a difference between having a token that's not valid and not having a token at all. And right now, the way that our code is written, ⁓ we're sending a response that's exactly the same in either case. So we need to enhance the www authenticate token so that the client knows what it it what action it should take. Like, I do have a token. I guess I should throw that one away and get a new one. ⁓ Or I don't have a token, I guess I need to go get one. And so right now, our ⁓ authenticate header that we send back.

Just as the barrel bearer realm and the resource metadata, we need to include the error and error description. So when the user makes a request and they post with an invalid token, the server comes back and says, hey, this is ⁓ invalid token, it returns nothing. ⁓ we can come back and say, okay, well that's an invalid token 401\. Here's the error description, this is what you need to do, and then the ⁓ the client can show ⁓ something more useful to the user. So

Pretty quick ⁓ little thing we need to do to our handle ⁓ unauthorized callback here. ⁓ I think that you can do it. So get to it.

—--------------

132

Harsh Bharadwaaj (00:00:00)  
Let's start out by adding a utility to our agent. So we're gonna go to index.ts ⁓ and we're going to add a require user. So async require user. We're going to call getUser from the database, and that's gonna get the current user because the database client has been set up with our auth token. So it knows who is requesting the current user. ⁓ And now ⁓ with that, ⁓ we also have an invariance in there to make sure that if the user isn't found, then like, well, we're in a bad

Place because our MCP server ⁓ is blocked. So if we can't find the user, they were deleted or something. So we probably shouldn't let them do anything that they're about to do. ⁓ So with that utility, then it actually becomes pretty trivial to update these tools ⁓ and this resource. So ⁓ yeah, we simply grab the user. It's gonna ⁓ our output schema is set to ⁓ the user schema. So that all is ⁓ working just fine.

And then we can come over to our resources, do the same thing here. We use agent require user. ⁓ And now we're all set. ⁓ So let's just make sure that that is working. We'll connect, we'll go list resources. Here's our user. ⁓ I don't remember. Last time I was in here, I think I did authenticate with Olivia. I'm gonna just believe that's the case. ⁓ And we'll go to who am I? That should be Olivia also. So there you go. That is ⁓ all it takes to actually now start ⁓ using the user ⁓ is

Just use the database, which is all hooked up with our ⁓ OAuth token, ⁓ and that gets resolved into who the user is and we get the user data.

—------------

124,

Harsh Bharadwaaj (00:00:00)  
Let's get started by grabbing the auth header ⁓ and determining whether the user has an auth header. Our AI assistant here wants to get the auth header and check whether it equals null. That works fine, but there's actually a has method that you can call directly on the headers, and that just returns a Boolean whether it has it. So that's even better. ⁓ So now we know whether we have the auth header. So we can come down here and update this. Now I'm

I'm gonna change this a little bit because I don't want it just to be like some massive string. Like that's no ⁓ look at that. Like, no, no, no, no, we're not gonna do that. So instead we're gonna turn this into an array, ⁓ which we will join with commas. ⁓ And then we'll just do this and we'll do that. There we go. ⁓ And then we can say, hey, if you've got an auth header, then ⁓ error invalid token, otherwise ⁓ and if you have an auth header, error description. ⁓

And here we're saying undefined. I like to say null, ⁓ but we do need to get rid of those. We don't want to join null. So we'll do this filter Boolean trick, which basically just filters out falsy values, which null is a falsy value. So that will filter out those, then we'll end up with an array of just two items that'll be joined, ⁓ or we'll end up with an array of four items, and those will be joined by the comma. So ⁓ we should have a proper ⁓ www authenticate header.

That gives the client more information about ⁓ why they're getting this 401 response. So well done.

127,   
Harsh Bharadwaaj (00:00:00)  
Okay, that's time for a break. Now go ahead and give somebody a high five or tell them something nice about themselves. It'll make both of you feel better. ⁓ And hopefully ⁓ you have enough water close by. If you haven't been drinking any water, then what are you doing? Your body needs water. Your body also needs blood flow. So get your body moving. ⁓ And with that, we also need laughter in life. And so here's a joke for you. I really want to buy one of those supermarket checkout dividers, but the cashier keeps putting it back. ⁓ Ha ha.

all right, awesome. Thank you so much for this exercise. Good job on it, and we'll see you in the next one.

126

Harsh Bharadwaaj (00:00:00)  
Let's get started by grabbing the auth header ⁓ and determining whether the user has an auth header. Our AI assistant here wants to get the auth header and check whether it equals null. That works fine, but there's actually a has method that you can call directly on the headers, and that just returns a Boolean whether it has it. So that's even better. ⁓ So now we know whether we have the auth header. So we can come down here and update this. Now I'm

I'm gonna change this a little bit because I don't want it just to be like some massive string. Like that's no ⁓ look at that. Like, no, no, no, no, we're not gonna do that. So instead we're gonna turn this into an array, ⁓ which we will join with commas. ⁓ And then we'll just do this and we'll do that. There we go. ⁓ And then we can say, hey, if you've got an auth header, then ⁓ error invalid token, otherwise ⁓ and if you have an auth header, error description. ⁓

And here we're saying undefined. I like to say null, ⁓ but we do need to get rid of those. We don't want to join null. So we'll do this filter Boolean trick, which basically just filters out falsy values, which null is a falsy value. So that will filter out those, then we'll end up with an array of just two items that'll be joined, ⁓ or we'll end up with an array of four items, and those will be joined by the comma. So ⁓ we should have a proper ⁓ www authenticate header.

That gives the client more information about ⁓ why they're getting this 401 response. So well done.

—-------------

Harsh Bharadwaaj (00:00:00)  
Let's get started by grabbing the auth header ⁓ and determining whether the user has an auth header. Our AI assistant here wants to get the auth header and check whether it equals null. That works fine, but there's actually a has method that you can call directly on the headers, and that just returns a Boolean whether it has it. So that's even better. ⁓ So now we know whether we have the auth header. So we can come down here and update this. Now I'm

I'm gonna change this a little bit because I don't want it just to be like some massive string. Like that's no ⁓ look at that. Like, no, no, no, no, we're not gonna do that. So instead we're gonna turn this into an array, ⁓ which we will join with commas. ⁓ And then we'll just do this and we'll do that. There we go. ⁓ And then we can say, hey, if you've got an auth header, then ⁓ error invalid token, otherwise ⁓ and if you have an auth header, error description. ⁓

And here we're saying undefined. I like to say null, ⁓ but we do need to get rid of those. We don't want to join null. So we'll do this filter Boolean trick, which basically just filters out falsy values, which null is a falsy value. So that will filter out those, then we'll end up with an array of just two items that'll be joined, ⁓ or we'll end up with an array of four items, and those will be joined by the comma. So ⁓ we should have a proper ⁓ www authenticate header.

That gives the client more information about ⁓ why they're getting this 401 response. So well done.

