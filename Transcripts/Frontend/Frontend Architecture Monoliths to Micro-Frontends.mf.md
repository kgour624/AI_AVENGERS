Sneha Mehra (00:00:34)  
And you're back. ⁓ Alright. So as a recap, what we wanted to do in this exercise was to ⁓ decouple the state. ⁓ or to be to be clearer, we want to move the state so that this microfront that we have here, this order status component, ⁓ owns the state instead of depending on the state that is passed down from above. So we're gonna use these nanostores library for

For doing this, so we're gonna start by installing the library. ⁓

Sneha Mehra (00:01:08)  
And we're also gonna install both nanostores and nanostores React, which gives us a nice React hook to use here with with this state. ⁓ I'm gonna just save that ⁓ in our analytics package at JSON.

With that, we're gonna go to our order distribution chart component, ⁓ which was receiving the state ⁓ and ⁓ the handler as props. So we're gonna remove both of those. We're no longer gonna accept this. ⁓ and now we're gonna replace with a piece of state. ⁓ Before we do that, we have to declare the store. So I'm gonna create a folder here in source inside of my analytics app called ⁓

We can c we can call it whatever we want, call it stores. Oops, I'm on the

Sneha Mehra (00:02:01)  
stores. ⁓ And we'll create a file called ⁓ we'll just call it count ⁓ dot state.ts, ⁓ which can will contain the atom, the atomic piece of state that we want to share. ⁓

It's very simple, ⁓ just import the atom ⁓ property from nano stores and you declare a ca a ⁓ a piece of state using that atom, which in this case will be a number and it starts at zero.

Now in our component we're going to import that count. So let's say import count. ⁓

From

stores ⁓ count.state ⁓ and we are going to use the ⁓ React hook ⁓ to ⁓ actually ⁓ use that piece of state. We're gonna find count ⁓ as use ⁓

Sneha Mehra (00:03:03)  
Tor, sorry. ⁓ yeah, use Tor.

Sneha Mehra (00:03:10)  
We have to import this U store. ⁓

Sneha Mehra (00:03:18)  
React, nanostorse react. ⁓

Sneha Mehra (00:03:23)  
And now that we have the count, we can render that count. And this handle click we need to update ⁓ because actually we want to do is we have to ⁓ set ⁓ this to the value of count plus one. Oops, that will be ⁓

Syntax ⁓ yeah we can do ⁓

Sneha Mehra (00:03:47)  
Let's test if this works first of all. Let's run MP and run dev. ⁓ Let's make sure that this is working. And yes, it's increasing my account correctly. But as you can see, it's detached from this other account. These are two separate pieces of state now. ⁓ Now, what we want to do is we want to make sure that this piece of state, whenever it changes, ⁓ we want to do something in the host application, in the dashboard. In this case, we're just gonna do a console log.

So let's go to our dashboard application ⁓ where we had the piece of state in this component. ⁓

I am going to remove the counter ⁓ that I had here just to avoid confusion that this is not the counter that we are going to be using or looking at. ⁓ And now I'm going to import that atomic piece of state from my micro that my microphone is exposing. ⁓ First of all, my microphone is not exposing anything at this point, so we need to expose it ⁓ in my ⁓ analytics app ⁓ RSBuild.json.

We need to make sure that we are exposing ⁓ the store. ⁓ So ⁓ let me see what ⁓ I'm I use this ⁓ here. ⁓

So we're exposing this ⁓ this ⁓ remote that is called count at state that points to this piece of state that we are that we have in the code base.

Sneha Mehra (00:05:18)  
And now here we're gonna import it. So again, this is evaluated at build time, sorry, at runtime, which means that we have to import it asynchronously. We're gonna use an asynchronous import. ⁓ And we can do this anywhere here because we only want to console log. So let's import from analytics. ⁓

What do we name this? ⁓ Count.state. ⁓

Sneha Mehra (00:05:44)  
And ⁓ when this promise resolves, ⁓ we're gonna grab ⁓ from here ⁓ the value of count, which is what this module is exposing. You can see that this is the file that is exporting, ⁓ if we look at it stay here, it's exporting ⁓ dollar sign count, ⁓ which is what I get access to here.

And now we've this is an atomic source. We're just gonna listen to changes ⁓ and ⁓ those changes will ⁓ will just console log the value. So whenever this value changes, we're gonna ⁓ console. ⁓

Sneha Mehra (00:06:26)  
Business type. ⁓

Sneha Mehra (00:06:30)  
And this needs ⁓ sorry, it's value, not count. ⁓ And this as always needs to be declared in the federation.tile that we have, ⁓ and we will see one way in which we can make this ⁓ much better. ⁓ But for now, let's just do that.

Okay, let's ⁓ see this is still running. So now if we did everything correctly and open the console, we should be seeing that ⁓ every time I increment this, ⁓ the value of count is logged here in the console. ⁓ So now we have this opposite direction. ⁓ the state belongs to my microfront end. And from outside of my microfront end, all I'm doing is listening to that state and how that changes. ⁓

the only problem that we have with this approach is that this still allows my consuming consuming app to mutate this state. So I could say count that sets one hundred.

And you see that now I like kind of mutated the value from outside from from my host application, it up it mutated the value that belongs to really to my microphone. And I shouldn't be able to ⁓ do that. ⁓ So what we can do to prevent this from happening is we can we can choose what to expose, right? In this case, we're exposing the entire atom, but maybe we we just want to expose a listener to that atom, right? So what we can do to do that is ⁓

We're going to create ⁓ a new file in our stores. This could be anywhere, of course, ⁓ called count.listener.ts. ⁓ And this will export a function that we're just gonna call ⁓ listen listen to listen to count. We can name this whatever we want. ⁓ And it will take a callback.

Sneha Mehra (00:08:29)  
Here we're gonna import our piece of state.

Sneha Mehra (00:08:34)  
count that state ⁓ and here we're gonna do the listening we're gonna listen to changes to count ⁓ and we're gonna invoke that callback. ⁓ We will have to type this probably so let me ⁓ type this as ⁓ value number

That's where we're trying.

Work, let me see dot T S. ⁓ I'm missing the From ⁓

Sneha Mehra (00:09:04)  
So now we can choose to only instead of importing the piece of state directly, we're just gonna export the listener to it. So we're gonna replace it with count.listener. This will export the count.listener. ⁓

Sneha Mehra (00:09:19)  
And now from the dashboard, ⁓ the count the state doesn't exist anymore. We're gonna import the listener, ⁓ which will export ⁓ a function called ⁓ listen ⁓ to count.

And we are going to call that function with our callback.

Sneha Mehra (00:09:39)  
Again we need to ⁓ make sure that this is declare here. ⁓ So just gonna update ⁓ my declaration. ⁓

Sneha Mehra (00:09:48)  
And now this should continue to work as before. I can still listen to count. Count ⁓ this is still being broadcasted. ⁓ But the difference is that I don't have access to mutate it. This is not exposing the count directly. This is exposing just a value, a function that I can use to listen to count. ⁓ But the only function available is just listen to it. I can't mutate it in any way. ⁓

Sneha Mehra (00:10:33)  
All right.

Now let's move on to the last thing really that we're gonna be talking about, which is ⁓ Module Federation 2.0. So this might sound like that is super new, but it's actually been around for a couple of years at this point. ⁓ and this is was about the time where they split module federation from Webpack and made it available as a plugin for all different libraries. They also made a couple a couple of ⁓ they added new features to ⁓

To the plugin to mostly also to differentiate it from users using import maps. ⁓ So we now have dynamic type safety. They introduced a manifest that they are proposing to use for standardizing microfronts usage ⁓ called MF manifest at JSON. The goal of this manifest is to have a standard way of deploying microfronts. When you have

Hundreds or even thousands of microfines of these little remotes, ⁓ each exposing different components, ⁓ and you want those to be versioned, it can be kind of a challenge to just to manage all of that. So this manifest will help with with that. ⁓ and they added some other features like ⁓ registering a remote at runtime rather than at build time.

So what we're gonna do now is we are going to upgrade our application from ⁓ module federation ⁓ one point five, which is the version that we're using now, is the version that's built in in RS build. We're gonna upgrade to ma module federation two point zero.

Sneha Mehra (00:12:07)  
So the first thing we need to do here is we're gonna install ⁓ the, ⁓ in this case it's the RS Build plugin because we're using RS Build. ⁓ This plugin includes the latest version of module federation.

Sneha Mehra (00:12:22)  
And now in our RS build config from our share folder, this is the RS build config that is shared by all of our applications, ⁓ just this one. We are going to ⁓ make some updates here.

Sneha Mehra (00:12:43)  
First of all, we're gonna import the module federation two point plugin, which is in ⁓

The RS build plugin package. ⁓

Sneha Mehra (00:12:58)  
Next we are going to replace our plugins here. We're actually going to put those up into ⁓ a plugins ⁓ array here at the top. ⁓

Sneha Mehra (00:13:13)  
Because we want to conditionally use the model federation plugin if our packages are actually using module federation. So ⁓ if the package that I'm looking at ⁓ has a module federation config, ⁓ which is ⁓ the case of, for example, ⁓ our analytics app ⁓ has a model federation config, so this ⁓ this will fall into that statement. ⁓ Then in that case, we want to add to the plugins array. ⁓

The initialization of the module federation plugin, ⁓ passing that module federation config.

Sneha Mehra (00:13:55)  
Finally we want to remove this. This is the built-in configuration from RS build, which is the one we were using before. ⁓ we're not using that anymore, so we can just delete that.

Sneha Mehra (00:14:09)  
And others on the build.

Sneha Mehra (00:14:14)  
You'll see here, I'm gonna stop the ⁓ bill so we can see because this will keep running. ⁓ That this is now running the DTS to generate these these files. ⁓ And it it does this in two steps. First, it extracts the files and then ⁓ it it it creates the the types definitions. And we will find them in our host application, which is the AppShell.

Sneha Mehra (00:14:47)  
It's in the analytics app. Let me see what's ⁓ maybe I stop the wheel too soon. Let me let me let this run ⁓ for a few ⁓ seconds. ⁓

There we go. It's gonna keep running because it's it's a problem with the we're we're generating these files and this ⁓ RSBL is detecting ⁓ the generation and the watcher of dev mode is running again and it's generating the files again and so on. That's why I have to stop it here and we're gonna fix it in a minute. ⁓ But for now I wanna show you that it generated the types automatically. It created this ⁓ MF types folder with the types of all of my remotes, right?

So you can see all of the components that we exposed in the remote. Now they have ⁓ their own type definitions. These are the screens and the components that we exposed. Here's the count listener that we just created. ⁓ And you can find ⁓ the specific type definitions here. ⁓ And you can see that these are fully typed. We don't have to declare anything. So now we should be able to ⁓ remove, let's say we go to ⁓

This application, if I remove this federation.ts where we were ⁓ maintaining these types manually, ⁓ I first of all I should get a bunch of errors now from TypeScript ⁓ in my application. ⁓ let's see. ⁓

Sneha Mehra (00:16:16)  
⁓ There you go. We're getting the error from TypeScript because ⁓ we lost those files. But now we can add this MFTypes folder to our ⁓ to the place where TypeScript will look for those types. ⁓ And ⁓ and it will it should fix them. ⁓ So ⁓ to do that, let's go to rtsconfig.json. And here I don't want to write configuration code by hand anymore, so I'm just gonna copy ⁓ this. ⁓

⁓ and basically we're saying go to the ⁓ at mf types folder to find ⁓ more more types to use, right? And as you saw, we got rid of those errors. This is ⁓ this is passing like nicely here. We have access to all those types, ⁓ and we don't have to maintain them manually anymore. So ⁓ for this feature alone, if you're using TypeScript, this is a good reason to ⁓ to upgrade to module federation 2.0. ⁓

Sneha Mehra (00:17:33)  
Okay, the the microfronts ⁓ world is ⁓ vast with a lot of ⁓ features that ⁓ not a lot of teams are gonna need. For ⁓ cases when you need more sophistication for microfronts, there are teams that are working on some really interesting stuff that that might be useful for you. So a lot of what's being worked on is about

Management managing microfronts ⁓ with different versions of each other. So in this case we don't have any version in form of microfronts, everything just we just have handle one version of each of our packages that we deploy. But if you have multiple versions, if our analytics app, for example, can expose multiple versions and deploy multiple versions, and your remotes can consume those different versions at runtime, then you might need something to coordinate those those versions. ⁓ So ⁓ there's been a lot of work trying to sort of standard

Like this. And a lot of companies are creating different schemas. ⁓ The front-end service discovery one is something that I think it's come from AWS. ⁓ So front-end service discovery. ⁓ And it's a way, it's a it's a proposal to generate a schema that lets you manage manage microphone and deployment. ⁓ I thought it was ⁓ here is ⁓

The actual front-end discovery pattern that they're proposing. ⁓ And this is what a schema would look like, right? And you it's almost like a package.json or package log.json, I should say, for your microfrontends, where you you define the different versions that and ⁓ you may have ⁓ you may find different dependencies of different versions between your remotes and your hosts. ⁓ Consider that this can also become ⁓

more complicated as you have remotes that depend on other remotes, which is entirely possible. ⁓ but of course it will make things more more complicated. If you have those needs, looking into some of this might help.

Sneha Mehra (00:19:34)  
Vercel is also working on a schema or a standard using microforend.json, which they support in Vercel if you want to deploy microfronts in Vercel. But they're proposing this as a way for ⁓ for deploying to other platforms as well. And we have the module federation ⁓ mfmanifest.json, which we didn't see, but now that we upgraded to ⁓ module federation 2.0, now ⁓ our applications should be including ⁓

Let see ⁓

Sneha Mehra (00:20:08)  
Looks like we're not including in the distribution. Let me see the app show. ⁓

Sneha Mehra (00:20:17)  
Well they're not included in the distribution by default, but it has an option to create these manifests ⁓ for you. So at when you build your applications, it will automatically create this manifest. And if you're hosting your microfinance in ⁓ in a platform that supports can read these manifests, then you get a ton of features out of them. Yes. ⁓ So on the topic of turbo repo versus modern mod hang on, let me try that again. ⁓

On the topic of Turbo Repo versus module federation, if the front-end shell has two apps, in this case kind of like the Analytics and Orders app, and you want to be able to use two different versions of the same component library or or design system that aren't necessarily compatible. ⁓ for example, if you want to be able to update the apps to a newer version of a design library one at a time, do you then have to move to module federation or is there a way to solve this with turbo repo?

I will say if you let me think about the so the question is if you're have two versions of the same dependency ⁓ and different components depend on two different versions and you want to avoid conflicts, is that is that what we're trying to And I and I think they're asking if maybe you want to update ⁓ one component's version of that dependency without the other one, ⁓ you know if they're sharing or if they're both including like for example in the case of a d design system and they're

Stuck on two different versions of the design system and you just want to be updating one of them. Is module federation really the only way to keep those separate, or can you solve that with Turbo Repo? Well with TurboRepo, you by itself it don't you don't have a way to have two copies in your source code or two versions in the source code, right? So if you're ⁓ if you're doing if you're consuming the source code, if you're treating your design system as a as ⁓ as an internal package where where you're consuming

just the source code, ⁓ then there is no way of doing of of managing two different versions. Now if you treat your design system as something that gets built, ⁓ in even if you don't deploy it anywhere, you just you just build it and you can build two different versions.

Sneha Mehra (00:22:31)  
Then you have the option of consuming either one of those builds. ⁓ So there's definitely a a way using TurboRepo to consume two different versions ⁓ of the design system. You have to make sure that those two versions are built separately, right? So you can think of having a build script in your package.json that has ⁓ two build scripts, right? Build v1 and build v2, those will get in two different folders, and then your apps can consume either either one, right? ⁓ But of course, module federation is. ⁓

is ⁓ an option as well. Okay, the final character sheet for our MF, MF, MF SPA, so module federation microfronts. The third one is my name, it's Maxi Ferrari. This is Maxi's Ferreira's version of microfronts. ⁓

what we are the main trade-off here is that we now we now have a ton of like scalability. ⁓ we don't have as much simplicity as we just ⁓ in this very simple example I run into some issues with module federation configuration. So we are definitely taking on a lot of more complexity, but we gain a lot of scalability. Teams can move faster, ⁓ many teams can move faster with this setup once you get the setup right, right?

⁓ and what we talk about the the the rest haven hasn't changed ⁓ much, but you will need to make decisions about how to ⁓ handle this type of distributed architecture ⁓ and it's it's it's best if you document them somewhere.

So we reached the end of the course. Let's wrap up with what we went through. ⁓ We started with some foundational ⁓ concepts about software architecture. We learned that architecture is all about decisions. Specifically, the decisions are important to the structure of your system. We saw how ⁓ the way you make those decisions is by listening to the architectural drivers or the things that will influence your architecture.

Sneha Mehra (00:24:22)  
We saw the frameworks of the four pillars that you can use to describe and evaluate different architectures against each other. ⁓ And we we touched on the relationship between architecture and AI. Then we saw a bit about monoliths. We saw that all monoliths naturally will evolve into this big big ball of mud.

But that you want to d avoid distributing that big bullet of mud as much as possible. You always want to first invest in increasing the modularity of your monoliths before you distribute them. ⁓ And we saw that the key to modularity is having good organization plus good encapsulation.

We saw how monorepos can help us scale these modular monoliths. ⁓ we migrated an application from a single repo to using monorepo strategy. And we saw how tool tooling with TurboRepo ⁓ or an X or Pants, if you want to use Pants, they can help us ⁓ scale it even further. And then finally, we talk about microfrontens. And the key takeaway of microfundance is that ⁓

This the problem that they solve is not like a technical problem, but it's more like an organizational problem. You have multiple teams that want to move independently, ⁓ yet your technical solutions are not allowing them to do that. So that's at the point that you reach ⁓ out for them. ⁓ But if you don't need them, exhaust every other option because they will just increase the complexity of your copies. ⁓

One other takeaway of microphones is that you should try to avoid treating them as components because they will start to get coupled together, and microfonents are all about independence. So you want to keep them as decoupled as possible. And finally, to find out what's the right moment to adopt microfonents, take a look at your costs. Adopt them when the cost of coordinating your monolith is higher than the cost of maintaining a distributed architecture. And that is all I got. ⁓

Sneha Mehra (00:26:18)  
Yeah.

Sneha Mehra (00:26:24)  
Awesome, thanks. I'll ⁓ leave the stream up just for a minute here while people ⁓ in case there's any final questions in the chat, but feel free to ask them away.

next week we'll have Chris Oliver ⁓ in the studio doing a getting started with Rails workshop and then after that ⁓ or the the next day we have TypeScript in the age of AI which kinda dives deeper into some more complex types and the kind of stuff that your ⁓ AI friends are generating for you and make sure you understand what they're generating. And then yeah, Kubernetes week two weeks after that. So it's just a a range of things happening here. Nice over the next few.

⁓ I'll also mention we just added ⁓ one last workshop in June yesterday. ⁓ Scott Moss is gonna come and do harness engineering and agent orchestration. ⁓ so that'll be ⁓ a good one on on June twelfth. That was newly added to the workshop schedule if anybody hasn't looked at it recently. ⁓ we got one question here in the chat I'll ask. ⁓

What are the main drawbacks of using module federation if you don't have lots of teams? ⁓ Can't there be some CSS issues with the cascade when using module federation or depending on which remote is loaded first, different CSS is applied? How do you detect and solve issues like that?

Yeah, definitely. There are tons of drawbacks of adopting module federation if you don't need it. Right. Like ⁓ that's why we started with monoliths and we spent ⁓ a lot of time like figuring out how do we scale this monolith before I have the need to reach for module federation because you won't exhaust every option, because they come with a lot of drawbacks. ⁓ Styling is one issue. ⁓ I will say Tailwind, which is you know of course is very popular and it really helps with microfronts because it helps ⁓

Sneha Mehra (00:28:16)  
each each component can have its own standard. You don't depend on like a global style sheet, for example, or you don't depend ⁓ on the cascade as much. That's the the in the same way that these atomic pieces of state help ⁓ help the decoupling for your state, ⁓ the atomic utilities of Tailwind help ⁓ the decoupling of your styles. ⁓ So I would say ⁓ use Tailwind is probably my best advice using microfunds. ⁓

If not you're gonna have some conflicts and you have you're gonna have one additional problem. Yeah. ⁓

Okay, I think that's ⁓ everything I've seen. Thanks again everyone online and we'll see you next week. Thank you, thank you. Thank you. Appreciate it. Thank you. Thank you. Thanks for

—---------------------

Sneha Mehra (00:00:18)  
Sure. ⁓ All right, you're live. ⁓

Alright, welcome everybody to the five hundred and forty sixth front end masters workshop. ⁓ so that's yeah, we've been around for a little while. ⁓ so today we have Maxi and Maxi has been a customer for ⁓ quite a few years, so it's pretty cool that you've gone from you know customer to the podium. ⁓ so this this first workshop ⁓ you know, obviously on front end architecture and ⁓ folks have been asking a lot about ⁓ what architecture should I use.

⁓ especially microfront ends is the the buzzword of the day. ⁓ so we'll be excited to get into that. And then ⁓ logistics wise, ⁓ Dustin, how many people do we have in chat? ⁓

Got about ⁓ seventy online right now. Cool. ⁓ so yeah, seventy people from around the world. Thanks for joining. ⁓ And if you have any questions, feel free to throw them in chat. either at me or ⁓ Dustin will ⁓ relay your questions to the chat. ⁓ We also have ⁓ six folks in the room, and ⁓ you guys obviously ⁓ ask your questions throughout the day and and ⁓ folks on the live stream will be able to hear you as well, so that's pretty cool. ⁓ so yeah, to get started we'll probably

just do a round of introductions and ⁓ if ⁓ you have ⁓ fun facts maybe throw that in. ⁓ yeah but yeah Max. Sounds good. Perfect. Well it's great to be here. huh it's nice to meet everyone. ⁓ would you like to to start Rodrigo? ⁓ Sure. hi everyone I'm Rodrigo ⁓ I have my own company ⁓ and fun fact I love kickboxing and I lost forty pounds last year. I gained some pounds back this year and I'm fighting myself to ⁓ lose those pounds again this year. Nice.

Sneha Mehra (00:02:05)  
⁓ Hey I'm Jeff. ⁓ I work at the University of Minnesota. Thank you. And ⁓ fun fact, I one time my lung ⁓ spontaneously c collapsed. ⁓ well. Okay. That's ⁓ gonna give people ⁓ nightmares. ⁓

Hi, I'm Dylan. I'm a front-end software engineer at Veradim, and my fun fact is I like board games. My favorite one is ⁓ Katan. Nice. Hi, I'm Kate. I'm a mobile engineer at a little nonprofit called Fields Kit. ⁓ and my fun fact is I lived in Kenya for three years and I know Swahili. A little bit, not fluent, but enough to get by, and a kunabatata does mean no worries. So ⁓

Hey morning y'all, my name is Trake. I'm a front developer at the Federal Reserve Bank in Minneapolis. ⁓ fun fact is I'm a big Mortal Kombat fan and I named my son Jax after Jax Briggs with the metal arms. Okay. Nice. ⁓ Hi, I'm Shelby. I'm also at the Federal Reserve Bank. ⁓ fun fact is I physically can't burp. ⁓ wow. Yeah. ⁓

Okay. Some stomach issues, but like it just never actually Yeah. Do you have to like take something in order to get the like stays in my chest? ⁓ It does make a noise, but like it doesn't actually come up. ⁓ Does that cost you any like trouble? Like in like any ⁓ I I'm guessing it's is it painful or like painful. Yeah. Sorry. ⁓

So I think it's the most fascinating one. Yeah. Some of these fun facts are not fun at all. ⁓ Well you have a fun one from chat. ⁓ when I was a kid I had a goat for a pet. ⁓ Okay. And it lived in the kitchen for a while. So that's that's pretty good. ⁓

Sneha Mehra (00:04:09)  
A kitchen goat sounds pretty odd. That that does sound fun. ⁓ well I'm I'm Maxi. ⁓ I work at a company called Help Scout. ⁓ a staff engineer there. Always work on the front layer. ⁓ and we're gonna talk a bit more about that today. Fun fact, I also love board games. ⁓ it's one of my favorite things to to do as a hobby. I don't have many hobbies, but board games I'll say is one. ⁓ and another one is my mom's birthday today. So she's probably not watching, but happy happy birthday ma.

⁓ all right. ⁓ With that, let's ⁓ ready to get started. ⁓ Okay.

So, welcome to front-end architecture monoliths to microfrontends. ⁓ We're gonna spend the next few hours ⁓ talking about architecture and sort of covering the fundamentals and also scaling a code base from ⁓ a very simple monolith all the way to using microfronts. ⁓ as an introduction, very quickly about me, my name is Maxi. I am a front-end ⁓ staff engineer at Help Scout, where I work as a front-end architect there. ⁓

On nights and weekends I also write this blog newsletter called Fortnite at scale, which ⁓ started as a way for me to sort of explore the realm of software architecture and software design ⁓ and see how ⁓ those concepts apply to the front layer. ⁓ So I ⁓ I write a lot in there about the things that we're gonna be covering today, so if you're interested. ⁓ and you can find me online as Charka on GitHub, Twitter, LinkedIn, and that's my my username. ⁓

What we what I'm hoping that we'll cover today and by the end of the day that what what you get out of this course is ⁓ first of all, ⁓ a very ⁓ high level fun ⁓ very high level version of the fundamentals of software architecture. So we will ⁓ come back to these concepts throughout the rest of the course, but I think it's important to ⁓ set like a base layer of understanding and also to make sure we're all on the same page when we talk about architecture. ⁓

Sneha Mehra (00:06:14)  
We will touch on the relationship between architecture and AI and how that has been evolving ⁓ over the last couple of years. ⁓ And I will also share a framework that I like for describing and evaluating different types of architectures. ⁓ And then the core of the course we're gonna spend most of our time ⁓ seeing or evaluating different methods for how we can scale ⁓ a code base, a front-end code base, incrementally from a monolith to microfronts.

⁓ and overall I I want to keep the course as practical as possible, which is hard to do with an architectural course because we're always talking about things higher level, but I wanna as much as I can see how show how those concepts translate into the actual code. So a very quick overview of the different parts of the project. We're gonna be ⁓

⁓ going through four different parts. So I'm I'm ⁓ sorry I nerded out a bit with this with the slides. I'm I'm a big fan of you know RPGs and fantasy ⁓ themes, so ⁓ I like to apply some of that here. ⁓ But we're gonna start our journey in with with some foundations. This is where we're gonna ⁓ explore some of these fundamentals that I that I was talking about. We're gonna define what we mean by software architecture. ⁓

And we'll talk about these four pillars of software architecture framework that ⁓ that I like to use. From there, we will go to ⁓ our very first architectural style, which is going to be monoliths. We're gonna start with a simple monolith, ⁓ and we're gonna see what what sort of growing pains that we'll have as that monolith scales over time. We'll touch on model monoliths, ⁓ and we'll and we'll talk about choosing our architectural style, using a pattern, ⁓ and also setting boundaries for your monolith.

From there, ⁓ we will take a detour and we'll go to we'll talk about monorepos, which isn't technically an architectural style by itself. You can do microfronts and you can do monolith with monorepos, but it can help solve some of these growing pains that we'll see with with monoliths. So I think it's important to cover some of that. ⁓ so yeah, we will talk about migrating the monolith to using a monorepo and how we can scale that with with some toolings.

Sneha Mehra (00:08:24)  
And finally, we will reach the microfonits area where ⁓ we will we'll talk about how how we can we can evaluate using a framework the the version of microfonents that you need for your project. ⁓ And we'll see what different versions or flavors of microfonents exist. We'll talk about ⁓ the simplest approach for me, which is to doing like a do-it-yourself version, and then we'll talk about module federation, which is this ⁓

pattern and library that we can use to to make it to make it scale even better. We'll also talk a bit about communication between between microframes. ⁓ So that's our journey. Please feel free to interrupt with any questions that you may have at any point. I'm happy to ⁓ to stay to go off the of the set path here. ⁓ So yeah, let me know if you have any any questions about anything. ⁓

The common thread that you will see throughout the workshop today is that we're gonna be trading off complexity for scalability. We're gonna as we scale this code base, we're also gonna be taking on more complexity. So this is just to say that none of these are better than the other necessarily.

Because they eel they each solve a specific problem and each one has ⁓ strengths and weaknesses. So our job ⁓ as as architects today is to find what's the best architecture for each scenario. And if we can't find the best, we should at least try to find the least worst architecture, which is ⁓ what we typically end up doing. ⁓ to find the resources for the course, there you can go to this website, charka.dev slash fm dash course. ⁓

That should be the only URL I'll probably need to share there because the this is contains the links to all the other repos. So I'm gonna open this now.

Sneha Mehra (00:10:11)  
We are going to be using three different repositories, each ⁓ one for each of the architectures that we're gonna see, the monolith, the monorepo, and microfinance. ⁓ and there is a script here that you can copy to clone all those repos at the same time. So I'm going to do it now.

Sneha Mehra (00:10:29)  
You can go to any folder and the script will create a folder called front-end architecture workshop. It will CD into that folder and then it will clone all these three repos ⁓ in one go. So ⁓ hopefully the setup should be nice and easy. ⁓ So now we're in the front-end architecture workshop, and as you can see here we have ⁓ our ⁓ our three repositories. ⁓

and the on the website you can also find links to ⁓ series of diagrams. We're gonna be covering those in in the later sections. And you can also find the link to the same slides that I am showing here in case you wanna go over them.

But yeah, those are the repos. ⁓ And ⁓ with that, let's get started with the first part, which is the foundational concepts. ⁓

⁓ before we get started, I want to make sure that we all sort of have a s the same understanding of what do we mean when we talk about architecture. ⁓ I know that we all probably have different things in our heads, ⁓ but I'm curious to know what are what are some of the the words or the concepts that come to mind when you think about software architecture?

Sneha Mehra (00:11:42)  
Yes. ⁓ Integration between different systems. Mm-hmm. Integration, okay. Yes. ⁓ Anyone else? ⁓ Boundaries. Boundaries. Yes, definitely. ⁓ extensibility. ⁓ Extensibility, right. Scalability or structure. Those those type of con they all come to mind when we think about architecture. And ⁓ it's architecture is hard to define because it's really all of those things. ⁓

The word that like to use is decisions. For me, architecture is the decisions that you make about how you do all these things, how you make it more extensible, how do you create their nice the better structure for your code base. ⁓ Ralph Johnson, who ⁓ is one of the authors of the original design patterns book, he ⁓

He has this this definition, which is that is the decisions that you wish you could get early right early in a project, which is probably one of my favorite definitions. ⁓ It's not ⁓ very descriptive. ⁓ so I have another definition that I like to use about architecture, ⁓ which ⁓ comes from a book called Design It by Michael Keeling, and is that it's the decisions about how we organize ⁓ a software project to promote desired quality attributes, which we're gonna talk about what those are.

So in order to make these decisions, we need to know like what how do we make what influences those decisions? And there are a bunch of things that can influence your architecture. ⁓ The main one should be like the business goals. What are we trying what are we trying to do here? The business goals are like the reason the software exists in the first place. So we should ⁓ we should do our best to at least understand it and see how those can have an impact on your architecture.

Quality attributes, this comes from the definition we just saw, and these are sort of the traits of your system. ⁓ These are things like performance, maintainability, scalability, ⁓ and defining which ⁓ which of those are more important to you and how those impact your architecture is also something that will influence your decisions.

Sneha Mehra (00:13:41)  
Constraints are in a way, if you can think of constraints, are decisions that have already been made for you, right? So these are like non-negotiable aspects. We can have technical constraints, like I don't know, you can't we can't we have to deploy to AWS, right? We can't use Vercel or Cloud or anything anything like that. So that will definitely have an impact of

what decisions I make in in how in the technology technology that I choose and things like that. ⁓ Or there could be business constraints. The most common ones are time and money. ⁓ Which is why like all of these that we're talking about today, we're talking about from the perspective of an ideal world. In an ideal world, I will do all of these things, I will consider all these aspects.

But of course, in the real world, we have constraints. We have time, we have money, we have resources to to take into account. So we really can't ⁓ create the perfect, the most ideal architecture. ⁓ The functional requirements are sort of the features of the product, right? This is what your product does. If you have an e-commerce application, some of those may be adding ⁓ you know adding an item to a shopping cart.

But not all of those influence your architecture. You have to look for what are the what are called the architectural significant ones, which are things that will have a say in your architecture. If you need real-time support, you need streaming or something like that, those things might influence your architecture. The team's experience, what technologies ⁓ your team is ⁓

is has has experience with will also definitely have an impact. Technology trends of the time, you know, AI, of course, ⁓ and and other things that though all of those things will influence your architecture. All of these things that we're talking about here as what are known as architectural drivers. And though from those you will come up with a set of requirements. ⁓ Your requirements are like a the formal description of your ⁓

Sneha Mehra (00:15:26)  
of your the of your the drivers that you have. So for example, if for your quality attributes you can define that you care about performance, but you have to be more specific. You have to say, okay, we care about performance and we have one requirement that says that every website must must load in less than one second or something like that.

so you'll typically put all of those in a document, and those are the things that will drive your architectural decisions, which you'll also like end up putting in documents. We're not gonna dive into like any of this, ⁓ but there are some like resources out there you can use to to see what an architectural decision document looks like. ⁓

Now we need a framework to use to evaluate and compare different architectures. And one that I really like is I don't think it has a name. I call it the four pillars of software architecture. It comes from a book called The Fundamentals of Software Architecture. And it essentially breaks down ⁓ an architecture, any architecture, in these four components. The architectural style, which is what the sort of the shape of the architecture. ⁓ For example, you can have microfonents, you can have monoliths, single-page applications.

The characteristics, which is just another name for the quality attributes that we care about, like performance, maintainability, and so on. The architectural decisions, which is what we just talked about, ⁓ like what are the decisions that you make about your architecture, and then logical components, which are sort of your building blocks. Now, like I said, ⁓ I'm a nerd and I like to ⁓ the way that I like to remember what these four pillars are, is to imagine that I'm building an architecture in the same way that I'm building

A character in a in like a role-playing game. So if you think about building a character, if you played a game, could be like a board game or a role-playing game where you're creating your character, you define at least a few things, right? You define the class, so what gives the overall shape to your character. A series of stats which will play to the strengths of the character. so ⁓ different car different classes will have different stats to care about, like strength, speed, and so on.

Sneha Mehra (00:17:25)  
Characters usually have a background which tells sort of the backstory and the where the character come from and so on. And also they have like a series of skills ⁓ to do what they do, right? Which we typically see represented ⁓ in a diagram like this one. ⁓ So I think we can map all of these four things to these four pillars of architecture we saw before.

I think of the class of the character as sort of the architectural style, which can be like we said, microfronts. So we typically refer to an architecture using just the style, which is good. It gives us a good idea of what the architecture kind of looks like, but it doesn't paint the entire picture, right? That's why we have all the other three. ⁓

for the architectural characteristics, we can define okay, these are the things that I care about in this architecture. ⁓ I care more about reliability than I care about performance, or I care more about scalability ⁓ than I care about agility and things like that. ⁓ this will come out of discussions with your team. And the point of doing this is that it helps clarify like the trade-offs that you will you have to make.

The sort of the background of the character, I think of these architectural decisions, which is like decisions about how things work in your in your code base, how do you set your boundaries and so on. And then I think of the skills as these logical components. What are the building blocks of your architecture? Your modules, features, screens, and so on.

So we'll come back to this. We'll talk about w when we evaluate each of the architectures that we're gonna see. We're gonna see them through the lens of this framework. We're gonna see we're gonna evaluate them using using this framework here. I want to touch a little bit on AI. I think ⁓ it's you know it's 2026, we can't ⁓ we can't talk about anything without talking about AI. And I think architecture is ⁓ I think architecture is interesting because

Sneha Mehra (00:19:08)  
A lot of the fundamentals ⁓ have only grown more important with AI, which we can say about every aspect of software development. I think we should always be sort of re-evaluating the what we think are the fundamentals. So for example, in software design we have ⁓ readability of code and refactoring. And those things, even though we can think of them as fundamentals, a lot of them they just just don't matter as much in software design, in the world of AI. ⁓

It's not just that refactoring doesn't matter, it's just that knowing all the little tricks of how to refactor your code base is not as important because the AI can do it much better than us, right? But for software architectural ⁓ fundamentals, I think they even grow more important. So if you think about context engineering, which is providing the context that your agents need to make decisions.

well a lot of the things we talked about, like ADRs, which are these architectural decision records, or the architectural requirement documents, design specs, these all fit into the context of IAI. ⁓ harness engineering, which is providing like this this guidance so that ⁓ and guardles so that agents can work autonomously for long periods of time. This also like ⁓ we can think of this as a representation of your architecture in in a harness.

And then your architectural drivers can help you like find what gaps you have in your in your code base. Like security is a big one with AI today. Which ⁓ is ⁓ it ⁓ became more popular with AI. I ⁓ think it's useful like defining an entire spec before you you ⁓ you give it to a coding engine coding agent to implement. But ⁓

I think there there is a problem. I think we should use it with caution. Because if you think about a spec driven development, what you're essentially doing is if you think of that the entire timeline of your project or how much time you're gonna implement, is spending a lot of time at the beginning defining like the entire spec in a lot of detail. And then from there you like you you go and do task one and then you do the task two and the agent does all this work, right? But this leads to ⁓ something that resembles a waterfall.

Sneha Mehra (00:21:19)  
And it's not so much about like the sequencing that is that is ⁓ bad or dangerous about this, is that in this model we can't use the learnings ⁓ that we learned during implementation to improve our design, right? We do all of the design up front, which can lead to to to a waterfall like this, which leads to not ideal solutions. Also, this doesn't take advantage of the LLM's ability to sort of connect the dock for us, connect the dots for us, which is something that LLMs are really good at.

So what I mean by this is that you don't have to provide like a spec that is super detailed for an LM to do to do a good job. If you think about, for example, vibe coding on one end of the spectrum, where you provide no instructions at all, you just say, okay, I want a React application. That's the same as saying, okay, ⁓ you provide some details, like I want a React application. You provide ⁓ one you just use Tannstack query and TandStack Rudder, and you provide like some of these details, right? And it's possible that ⁓ the

architecture or the the software that you oops sorry, I'm ⁓ not very good at this. ⁓ But it's possible that it will resemble something like this, what you asked the the AI to implement. But it's also possible that it will connect it in like some other way, right? ⁓ it might not be whoops have been trouble here. It might not be like the shape that you were expecting.

So spec driven development says, okay, I'm going to provide more dots. I'm going to provide more detail. But it's that's saying like I'm gonna provide every single detail about how the thing is should be implemented, which as we said, like this this ⁓ doesn't take doesn't take advantage of the fact that AI can fill the gap for us. So what I like to do is to find like the middle the middle spot, which there's this concept of just enough architecture, which is ⁓ just pro finding finding where to stop for an initial version to provide sort of the

sort of the foundation for what you wanna do. ⁓ the important details of what the foundation should look like, implement that and then iterate on that. Go ⁓ build it in stages. Don't try to build everything as at the same time, but build it in stages. And finally the the last thing I'll say about AI is that when you talk about architecture, you just always think of it as having a conversation with the model rather than just providing instructions. ⁓ they they

Sneha Mehra (00:23:36)  
They have a lot of knowledge and some a lot of times they can be really useful in making ⁓ in helping us make decisions.

All right. ⁓ we're gonna start with now with our actual ⁓ the what the first architecture we're gonna take a look at, which is Monoliths. So for that we're gonna start with a quick overview of the project that we're gonna be working on. So throughout the entire course, we're gonna be looking at this fictitious company called Commerce OS, which ⁓ it's ⁓ an e commerce platform you can think of.

As a mini version of Shopify. ⁓ And in this scenario, the very first version, we are a small startup. Maybe we have a handful of engineers. ⁓ And what we're trying to do is just to get started. And the best way to do that ⁓ is with the monolith. I think monoliths really shine in their simplicity. And when you're just getting started with a project, it's really ⁓ they can be really, really useful. So you find a link here, which we will also find it in the course.

with a series of diagrams. We're gonna take a look at the V one diagram.

This diagram is ⁓ using something called the C4 model. If you're not familiar with C4, I'm not gonna go into details in C4, but you can go to cformodel.com, and that explains like they have a lot of ⁓ really good resources on on how this works. But it's essentially a way to visualize your architecture at different levels of detail. So it gives you four types of diagrams. The level one diagram is the one that we're looking at right now, which is the context ⁓ system context diagram. And this takes a

Sneha Mehra (00:25:15)  
completely zoomed out view of your architecture. So ⁓ and this view you can't really see softwares yet. You can't see APIs or databases or anything like that. You just see ⁓ like the systems that you have, which makes this type of diagram really great for sharing with people who are not technical. And then from there you can zoom in. Zoom in to if we zoom into level two, we can take one of those con those containers that we have, sorry, one of those systems that we had.

And we can expand it and we can look into what's inside of them. So this diagram, which is the level two, and let me zoom in a little bit. It's the same thing we were seeing before, but now that that ⁓ that admin system box is expanded so we can see what's inside of it. And we can see here that we have a very simple architecture. We is we're starting with a web application, ⁓ which is a React ⁓ React single page application. We have a core API, and we have a database, like the most simplest of architecture.

And what we're gonna be focusing on is on this admin web app. So this is what we're building today, only the admin web application.

So with that, let's ⁓ jump into this the first repository that we're gonna cover, which is this one: front-end architecture monolith. So let's see into that.

Sneha Mehra (00:26:45)  
You can run npm install to run to install the dependencies. ⁓ this is a simple React single-page application. I don't think you're gonna run into ⁓ any troubles. ⁓ but if you do please let me know. And then you can run the server with npm rundub. Let's take a peek at what this ⁓ application looks like. ⁓ We have a fake ⁓ auth system. You can just click the sign in button here. ⁓ but this is just to show that we have different levels of systems, we have a permission system in place.

This application sort of try to simulate the complexities of a real application. So it it does have a ton of features, like and everything is like fully functional, mocked. We're not really saving any data, storing any data in a real database. But everything should be fully functional. You can add ⁓ you can add products to the catalog, you can ⁓ you know, you can touch everything here. ⁓ and this is ⁓ again, this is represents sort of the admin portion of ⁓

a tool like Shopify where you have all these different ⁓ you can you can create a store and you can upload your products, you can you know ⁓ sell them and manage the inventory and so on.

Sneha Mehra (00:27:57)  
let me first let me open the covet so you can take a peek also how the covis looks like.

Sneha Mehra (00:28:16)  
Can never remember how to close the ⁓ the agent chat. Let's see it. there you go. ⁓ So if we take a look at the code, you'll see that we have a very simple structure, which is what you probably get out of every monolith that you create. You have, you know, ⁓ not much ⁓ separation of concerns at all. We have everything in a components folder. We have concepts such as routes ⁓ and you know hooks and things like that. So these are

This is like a technology split of our architecture. It doesn't tell us anything about what this application is about. We can't see that this isn't like an e commerce application from this example.

Sneha Mehra (00:28:57)  
⁓ so we're gonna we're gonna see how we can improve it. ⁓ If we evaluate it, ⁓ this is how I'll create a character sheet for this architecture. It's a monolithic SP SPA. Some of the ⁓ architectural characteristics it promotes is simplicity and ease of deployability. It's very easy to deploy this type of architecture. It's also very easy to maintain. Performance, not the best. It's a client-side render SPA, so it can definitely do better, but it's also not not the worst, at least not at this stage.

The big drawback here is a scalability. This mo this architecture that we have right now is not gonna scale very well as as the project grows. ⁓ And for that reason, we're gonna start to run into a bunch of different growing pains with this type of architecture. The the main one is like there's a a lack of organization, which makes things harder to find, harder to maintain. We don't have any clear boundaries in our system to to say where the feature where one feature ends and the other one begins. Everything is gonna

mess. ⁓ let ⁓ me ⁓ start over. We don't have any clear boundaries in ⁓ in our application. Everything is like together in one single one single context, ⁓ which we're gonna see how that can can cause some trouble. ⁓ and we we'll touch about some of these other growing pains, but you can probably imagine like ⁓ as the team grows and as the cobalt grows, we're gonna start to run into some some some of these challenges.

So let's actually see an example of what I mean by ⁓ particularly this unclear boundaries problem. Let's say we have, let me go back to website. ⁓ we have this orders feature right here, where you can ⁓ go into one of the orders and you can see all the items that are in this order. And then from here you can see you can click on one of the items and it takes you to the catalog so you can see the details of that item. Let's say we want to implement a feature, because

Right now here, if I click back to the back button, it will take me back to the index of the catalog. But I came from my orders, right? I came from here. So let's say we want to implement a feature that says if I enter my product from my orders, I want this this button to say back to orders. And when I click back, I want to go back to orders instead go back to the catalog. Very simple feature. And the way we're gonna implement it is we're gonna grab this URL link and we're gonna append the order ID, which is you can see it here in the URL.

Sneha Mehra (00:31:25)  
As a query param, so we can then go back. So it's it's it's ⁓ something that you might have done like a million times before. So let's go ahead and implement that. ⁓ we to do that we are gonna open the ⁓ one component called order line items ⁓ table is the table here that we want. ⁓ This renders ⁓ this table right here with my line items. So we wanna update this link to send the query param.

to ⁓ actually ⁓ to include the order ID here. So we can do that since we're using ⁓ Tanstack Radware in this case, ⁓ we can grab the order ID ⁓ from ⁓ and have cursor helping me here, ⁓ using the use params ⁓ and ⁓ and we're gonna save that into this order ID and then we're gonna pass it here ⁓ as a search param. ⁓ My link. Perfect.

So with that let's I'm not running my project, so let me run that again.

Sneha Mehra (00:32:33)  
So let's see if that works. ⁓ and as you can see, this ⁓ clicks here includes the order ID in the URL. So now we can use that to say, if I have an order ID in the URL, I can click go back to to the order instead of going back to the catalog. Pretty straightforward stuff. Let's say we this looks good to us and we ship it. Now, 20 minutes later we hear from another team that's working on the customers app and says the cost the customer's app is broken. But maybe we test we maybe we even tested the customer app.

And we saw, okay, this goes to order, but this link works. So what is it broken? Well, it turns out that if you hover over this link for a little bit, everything blows up. And this is also like the type of tricky case that it doesn't it might not show up in a test, because in here what I have is like I have a I have a tooltip that shows up on on a timeout. So it this might not even show up on the test, but it's breaking my application. And the problem here that I have here is that this ⁓ component

The the component that renders the the little tool tip that you can see. Actually let me let me roll this back so you can actually see what's going on here.

Sneha Mehra (00:33:44)  
This is the tooltip that is supposed supposed to show up in this customer's application. And this re this uses the same exact component, the same a table component that the orders uses, right? This component right here. So we're crossing sort of the boundaries between the customers app and the orders app. And there is no way for us to know that until that hits production and everything breaks.

So that's one example of this these unclear boundaries. ⁓ these unclear boundaries growing pain. Or we can see like examples of others because they will it will show up. And like I said, the problem that we have here is that we have, let me actually do this.

Sneha Mehra (00:34:27)  
See I can this works now. The problem that we have is that oops. Well, I ⁓ I spoiled the surprise, but ⁓ I was gonna ⁓ I was gonna like visualize how how we end up with this big ball of mud, which is sort of the natural evolution of the the COVID, because we have ⁓ one module ⁓ that

Sneha Mehra (00:34:50)  
We have one module that is dependent on another module in an implicit way. We don't know that this relationship ex exists because there is supposed to be a boundary around this module, right? Around my customer's logic that doesn't exist. So ⁓ we have this relationship. We are allowing this relationship. And we can have the same sort of the same type of relationship across all the different parts of a code base, which is how we end up with this big ball of math ⁓ architecture. And this is the point where.

Teams will say, okay, our monolith has become a big bottom of map. The monolith is a problem. So we need to decompose it. We need microphones, right? A lot of people will advocate for microfonents at this point. But that is a ⁓ that is a trap, and I will tell I'll tell you why. ⁓ We have to look at these architectures on these two different ⁓ under these two two different dimensions. The first one is the number of deployment units, which is ⁓ of course, in a monolith we just have one. With microphones, we have many deployment units.

But the other dimension is that the degree of modularity of your code base. So when we have the this big ball of mud, we have both low modularity, everything there are no boundaries, and we have ⁓ just one deployment unit. Microfronts are on the other end of the spectrum, right? We have ⁓ lots of deployment units and everything is nicely encapsulated. But the problem is that if you take the big ball of mud and you try to increase the number of deployment units without imp imp improving the module.

Modularity first, you don't end up with microfurns. You end up with a distributed big value of mod ⁓ or a distributed modernity, which is much worse, right? It's much worse because you have all of the same challenges we had before, and we had all of the same challenges of a distributed architecture, because now we need to care about ⁓ communication and ⁓ handle different versions and so on. So this is ⁓ the worst of all the worlds. We want to avoid this at all costs. The alternative is to invest in increasing the modularity first, and that's how we end up with a modular module.

⁓

Sneha Mehra (00:36:51)  
Yes. So how do we create a module monolith? The first thing we need to do is need to choose sort of like an architect architectural pattern so we can define what those modules look like.

And we have many different options. In the world of full stack ⁓ architecture or backend architecture, there are there are many different ⁓ no well-known ones. The most simplest one is a layer architecture where I divide my code base into different layers. We have the clean architecture one, which is sort of divides into ⁓ the logic into different buckets where you have your entities and use cases and so on. Hexagonal architecture, which is sort of like a precursor to clean architecture and uses the concepts of ports and adapters to

To handle to separate your logic from the different clients that you may have. Domain-driven design is also like a very popular way of defining these different modules. And then if you ⁓ are ⁓ if you're using like a pure front-end application, there's also you have the option of using something like atomic design, which sort of fell out of fashion in the last few years, but it fits really well the model of components that we use on modern applications. ⁓ There is also this one called feature slice design, which

⁓ it it it's very interesting because it gives you enough enough ⁓ structure so you can find where your logic fits, but it's not as strict as some of the others.

⁓ some of these architectures, like I said, they were designed for back end applications or full stack applications. So trying to fit clean architecture, for example, in a pure front-end code base can feel clunky at some time. So I don't recommend like trying to follow one of those to to the T. What I like to do for when adopting module monoliths is I like to use the concept of subdomains from domain driven design to identify those modules. ⁓ And then ⁓ from there define a folder structure and then create my boundaries. And that's how we end up with

Sneha Mehra (00:38:42)  
What I consider to be a modulum.

Sneha Mehra (00:38:47)  
How do we identify subdomains? So this is a whole thing, I have to say. If you go to a domain driven design workshop, it will take us an entire weekend to try to do the exercises of identifying the subdomains. So you can think of them for the purpose of this workshop as sort of a slice of your main business problem, right? And you typically divide them into three three types the core subdomains, which are the things that make your product unique ⁓ or competitive.

⁓ in the case of our application, could be the order management ⁓ subdomain. You can you can have supporting subdomains, which are also important, but they are not ⁓ the thing that make you unique or different from the competition. And then generic ones which are the undifferentiated ones like authentication or or user management, things that every application has.

⁓ there is one resource I want to share. If you want to do the more formal process, which we're not gonna not gonna do today, but of identifying your subdomains, there is a good resource called DDD starter modeling process, ⁓ which is a GitHub repo, which has a ton of resources on DDD and how to do this this process properly or formally. This is sort of the whole process, which again we're not gonna go into, but if you're interested in this kind of stuff, this is a good way to do it.

For the purposes of our project, we're gonna start with an exercise about identifying the different subdomains of a little e-commerce application. So since we have the code base already implemented, we're gonna cheat a little bit and we're gonna browse the UI and the code base for for hints on how to find this subdomain. So look at things like the navigation, the different routes that we have, and things like that, and try to identify those domains. And s if you can, like you can split them into

What are a core subdomain of our application, a supporting subdomain, and a generic subdomain? So if you wanna take a few minutes, five, ten minutes, we can ⁓ try to go through this exercise and I'll show you my solution in the next few minutes. Should we maybe do ten for this first one just to give people ⁓ everything works? Sounds good. Cool. We'll be back in ten.

—----------------------  
Sneha Mehra (00:00:05)  
⁓ Alright, we're back. ⁓ Alright, so here is the ⁓ why this wasn't running before. ⁓ I just ⁓ I just made a typo here. This shouldn't be remote, but should be remote. ⁓ Because we may have more than one.

Let's run the build again and this time it should pass. ⁓ But if we try to open the app ⁓ localhost three thousand, ⁓ we will get a run a blank page because we have a runtime error. ⁓ The runtime error is saying that ⁓ we're trying to load one of these shared dependencies that we declared before ⁓ a s ⁓ asynchronously, and we need them synchronously.

So this is coming from our app shell. ⁓ And what we're when we declare a shared library, we're also creating a separate bundle for the library in case we want to share that with other remotes. And the problem is that this bundle, by declaring it share, will be loaded asynchronously. And React and React DOM are things that we actually need to be available on the page as soon as we load the components. So we can't just wait for React, at least in the way that we have it set up right now. If we set it up in a way where we ⁓

We do it asynchronously rendered, then it will work. But in this case, since our app requires React to be loaded when I load my my app, I need to define that these are ⁓ what they call eager dependencies, which means that they will be loaded at runtime. ⁓ as soon as my page ⁓ runs at runtime.

So with that, I shouldn't need to refresh or anything. ⁓ And my app should be back to work. ⁓ And let's see if my analytics app is working. ⁓ It's not, it's failing now. Let me go back. ⁓ Everything in my app shell continues to work, but when I go to the analytics route, which is the one loading my analytics remote, ⁓ we're working around around an error. ⁓ Now, ⁓ this is running, this is failing for

Sneha Mehra (00:02:08)  
Probably the same ⁓ reason, let me see. ⁓ What's the actual error here?

Sneha Mehra (00:02:33)  
Let me see if this next step, which is what we're gonna what I was gonna do next, fixes this. ⁓ singleton ⁓ eager ⁓ true. ⁓ So defining Tansac query also has a shared library. ⁓ And I'm gonna do the same thing in my ⁓ analytics. ⁓

Let's run the build again. ⁓

Sneha Mehra (00:03:04)  
We are still not working. Hmm, okay. ⁓ Let's try to debug this in real time. This is waiting for localhost remote entry, which is not ⁓ loading. ⁓

Sneha Mehra (00:03:23)  
Because it needs React, which is interesting. ⁓ Let me try to define React and React DOM also as ⁓ eager true here. ⁓

Sneha Mehra (00:04:05)  
There's something clearly. ⁓

Not working nice here. ⁓ I'm getting an error on the app shell now, even though I didn't touch the actual config, which is ⁓ this might be sorry, the up the actual error might be ⁓ because of this. ⁓

Sneha Mehra (00:04:33)  
That ⁓

Sneha Mehra (00:04:50)  
Then me. ⁓

Sneha Mehra (00:04:55)  
move to a branch that I have prepared where this is actually working. Sorry about that. ⁓ could it be you have to specify the version of React? No, it's not necessary to specify the version.

Sneha Mehra (00:05:09)  
Let me see if I miss something. ⁓ Is the analytics app running? ⁓ The analytics app is running, localhost ⁓ in localhost three thousand ⁓ and one. ⁓ And it should ⁓ I I've lost my my monitor again. ⁓ Again? ⁓ Try plugging it, plug it back in. If it keeps doing it, we can swap it out for a different monitor. It might be ⁓

Sneha Mehra (00:05:38)  
The remote entry is not defined. Maybe it do you have a typo here with the remote entry? ⁓

Sneha Mehra (00:06:36)  
Do it that so

Sneha Mehra (00:06:53)  
⁓ Sorry about that. Let me change to temperance ⁓

Sneha Mehra (00:07:28)  
Okay, there we go. ⁓ I ⁓ don't know what happened. ⁓ What was the difference? ⁓ Because if you look at ⁓ our configuration it looks exactly the same. I might

Sneha Mehra (00:07:48)  
Just comparing the before and after, ⁓ it looks the same.

Sneha Mehra (00:07:55)  
Okay.

Sneha Mehra (00:08:02)  
We can debug that ⁓ later. I'll try to leave a note about what happened there. ⁓ but ⁓ this is now ⁓ working correctly as we expected. Let's just make sure that this is actually loading a remote entry file. ⁓

Sneha Mehra (00:08:21)  
Fort three thousand and one. ⁓

Sneha Mehra (00:08:30)  
⁓ So my host app ⁓ that lives in localhole three thousand is loaded in the remote entry file from three thousand and one ⁓ and it's running the it's it's this analytics screen is coming from from there.

Everything is still in the same app, but this portion, this portion right here, ⁓ is coming from a different application that is built and deployed ⁓ independent of the others. ⁓

Sneha Mehra (00:09:24)  
One thing that I have on this branch that I didn't touch on ⁓ before is that I've also added this import ⁓ config here to my shared dependencies. ⁓ And what this is doing, it's ⁓ preventing the remote. This is on my analytics page. ⁓ And if I don't say anything, by default, ⁓ the remote will still bring a version of React to use as a fallback. Because as we as we said earlier, it's possible that I'm using a remote.

on ⁓ an app that doesn't have React installed, right? Maybe it's a plain HTML app, but ⁓ I can still use this remote that depends on React because it brings a fallback to use whenever the host doesn't have that dependency. ⁓ the problem is that we are loading that version of React ⁓

by default, even if the host already has React. So if we want to prevent that, if we know for sure that this remote will always be used ⁓ on a React application, React will be available, then we can say import false, and that will remove that extra import. ⁓

Okay.

If you want to get to the same place where I am right now, you can check out ⁓ the module dash federation-1 ⁓ branch, which will take you exactly here. ⁓ If you had any other changes, you can stash them or you can commit them to a temporary branch. ⁓ But you can look at what we have here in this branch. ⁓ And that brings me ⁓ to the next exercise, ⁓ which is ⁓

Sneha Mehra (00:11:02)  
To create a new microfront using module federation using the config that we saw before. ⁓ But now the analytics remote, which is the one that we we configured here, is gonna expose another component. Not just it's gonna export the screen, but we also want it to export an additional component, which is the order status distribution chart.

You can find it in the code base if you look order status distribution chart. ⁓ It's this component right here. ⁓ We want to expose that, and then we want to replace the component that we're using in our dashboard ⁓ with that component. ⁓ So, ⁓ to show you what I mean exactly, ⁓ the analytics app has one chart here that is called ⁓ order status distribution. It's this graph right here. ⁓ If we look at the dashboard. ⁓

⁓ I don't have my app running, sorry about that. ⁓

Sneha Mehra (00:12:06)  
If we look at the dashboard, we have a similar component. It's showing kind of the same thing. So we wanna do is we wanna replace this one, this order distribution chart, from the one that we use in the analytics app. This analytics app, we want we want this app to expose this component as a remote, and then we wanna import it here in the dashboard and show it in place of this one.

Sneha Mehra (00:12:33)  
So ⁓ with that we're gonna take a quick break and we'll come back. Well do you ⁓ you feeling good on time? Well actually can't even can't even see the time because clocks on. ⁓ like ten minutes? Yeah, ten minutes is ten minutes works. Yeah. Sounds good. ⁓ Be back in ten. ⁓

Sneha Mehra (00:22:57)  
Alright, let's see how we can implement this. ⁓ So, the first thing we need to do is we need to expose a new component from ⁓ our analytics remote. ⁓ So, analytics is still ⁓ one single ⁓ application that gets deployed, ⁓ but it will expose now another component in addition to the screen, ⁓ which is this component or the distribution chart. ⁓ This is a regular JSX component that just renders ⁓ a graph. ⁓

Now we need to consume that component, ⁓ and we're gonna do this from the dashboard ⁓ overview ⁓ component. ⁓ We are going to ⁓ do something similar to what we had before, ⁓ where the existing order distributions component. ⁓

Sneha Mehra (00:23:51)  
this one right here, which is this ⁓ entire card, we're gonna replace this with the component that we we are importing from the remote. ⁓ So let's just get rid of this this card. ⁓ And here is sort of a placeholder for where we're gonna import that.

As we did before, since this is now ⁓ evaluated at runtime, that means we can't import it directly. We have to import it ⁓ with an asynchronous import. And the way we do that ⁓ with React is we use the React Lacy function ⁓ to import the components from the analytics remote. ⁓ And then ⁓

Sneha Mehra (00:24:34)  
When we're gonna use that component, ⁓ we have to have ⁓ a suspense boundary around it, just like we did before.

Here I am rendering that distribution chart component, which comes from my other application. Remember, we are here on the app shell essentially, and we're rendering this component from the other application, and we are passing some data to it. ⁓ and the last thing we need to do is we need to make TypeScript happy by defining. ⁓

⁓ federation. ⁓

.d.ts file. It doesn't have to be called like this, you just need a ⁓ type definition file ⁓ to make sure you don't. ⁓

Sneha Mehra (00:25:20)  
You can distype scarrels.

So let's run this now, NPM rundev. ⁓

Sneha Mehra (00:25:31)  
Our app should work. And ⁓ this component now, it looks kind of similar, but this component now is coming from the analytics app. It's not the same component that we were using before. ⁓ And we can test that this is actually being deployed independently and the app doesn't depend on it.

By bringing down the analytics server and checking that the app continues to run, the rest of the app continues to run smoothly. ⁓ So, ⁓ as a reminder, my application shell is running on port 3000\. ⁓ The analytics microfurnet is running on port 3001\. So now let's go to my dev script in package.json. ⁓ I'm going to remove the analytics run, I'm not going to run it anymore on my dev script. ⁓

Dap.

I'm not running the analytics app, only run the app shell on the API. This will likely blow up because we don't have t any type of error error handling here. So let's add some of that. ⁓

Sneha Mehra (00:26:37)  
And the way we can do that is ⁓ in React we have these error boundaries that we can use. ⁓ But since I don't have defined this is something you have to define kind of on your own or bring in a plugins. Since I don't have one defined here, I'm just gonna do the error handling at this level. I'm gonna replace this with a try-catch statement to catch my errors. ⁓

Sneha Mehra (00:27:02)  
So I'm gonna do the import here, I'm gonna return this import, which means this needs to be ⁓ an asynchronous function now.

And for ⁓ your return, ⁓ you have to ⁓ return what this import will return, which is an object that at least has one default. If we had name exports, you would put them here. Since we're using the default export, ⁓ we need to define the default. And here this can be anything you want, it can be ⁓ React component or anything. ⁓ let's just say, oops, ⁓ in this case. ⁓

Sneha Mehra (00:27:42)  
See what's happening here. ⁓

Sneha Mehra (00:27:57)  
I'm sorry, this is the same error that we've been seeing before. ⁓

Sneha Mehra (00:28:07)  
Should be ⁓ fine.

Sneha Mehra (00:28:42)  
Sorry about this, I don't know why I can't demonstrate the error handling now. ⁓

Sneha Mehra (00:29:11)  
Let me try to remove ⁓

Sneha Mehra (00:29:23)  
Import from the op show. ⁓

Sneha Mehra (00:29:47)  
Actually let me ch switch to another branch. ⁓

Sneha Mehra (00:30:42)  
still getting this error from ⁓ the analytics ⁓ application share package react. ⁓

It's ⁓ interesting.

Sneha Mehra (00:31:05)  
Let's try this one more time. ⁓

Sneha Mehra (00:31:09)  
Nope, still getting the error. ⁓

I don't know. ⁓ This might be a problem. Let's actually try ⁓ to run ⁓ the build instead of ⁓ this because I'm I'm wondering if this is ⁓ those kinds of errors that ⁓ are caused by the ⁓ by the build by the sorry by the dev version and not exactly ⁓ the build version which is what customers will see in production. ⁓ So let me try to run the build, which will probably ⁓ fail at this point because we have to define ⁓

Sneha Mehra (00:31:56)  
In our dashboard, ⁓ we also have to define, since we're building now the dashboard separately, we also have to define ⁓ our configuration here ⁓ for our remote. So let me just copy that very quickly. ⁓

From the upshore. ⁓

Sneha Mehra (00:32:21)  
Dashboard.

Sneha Mehra (00:32:28)  
Build. ⁓

Sneha Mehra (00:32:34)  
Okay, and now we can run ⁓ npm run preview instead of npm rundev. ⁓ We have everything running in the same port. ⁓ Again, this sort of is this simulates more what we'll see in production because this is actually running the ⁓ it is consuming those files from the distribution folder instead of consuming the apps from the dev server. ⁓ And now we should be able to ⁓ go to our package.json. ⁓

Sneha Mehra (00:33:07)  
remove the analytics ⁓ app from here. ⁓

Sneha Mehra (00:33:18)  
Did you mean run preview or run dev? ⁓ Thank you. I meant run preview. ⁓ Here run the build, npm run build, ⁓ npm run preview. ⁓ That's what I meant. Thank you. ⁓ Okay, ⁓ now we're getting the error that we're not catching from ⁓ our error handling. So let me go back here and ⁓ re add ⁓ the try-catch that we had before. ⁓

Sneha Mehra (00:33:50)  
Place this with a tri catch. ⁓

Sneha Mehra (00:33:57)  
Seems to be async. ⁓

Sneha Mehra (00:34:02)  
Okay.

Let's run the build.

Sneha Mehra (00:34:14)  
Review.

Sneha Mehra (00:34:18)  
So they run wrong.

Sneha Mehra (00:34:30)  
not entirely sure where this error I know it's coming from the fact that you can find the remote entry but it should be ⁓ I might be missing some config, but this should be able to ⁓ to catch that error here and show us something else instead of ⁓ blowing up. ⁓

Sneha Mehra (00:34:51)  
I will find what is whatever it is that I'm missing here. ⁓ Let me see.

Sneha Mehra (00:35:14)  
If this ⁓ contains ⁓ what I need. ⁓

Sneha Mehra (00:35:26)  
This one doesn't contain that example either. So ⁓ I'm sorry I can't sh I can't show that to you, but ⁓ there might be something in my configuration that it's requiring this at low time, and this is why I can't run the app ⁓ if I don't have the other remote server running. Which shouldn't be the case because that's ⁓ the whole reason we w we want might want to do this is that we want to be able to deploy these apps independently. ⁓

I will find the actual code example that I thought I had prepared here to show you the error handling, but for now we're just gonna stay with this. ⁓

Sneha Mehra (00:36:12)  
Let's ⁓ talk a little bit about ⁓ communication between microphones.

So when we depending on which ⁓ horizontal split solution we use, we have different options. With iframes, really the only option that we have is using the post message API, which is a browser API that let us send messages between ⁓ one one iframe or one document running in a frame and another frame. It could be between iframes, it could be between the host and the and the actual iframe content. ⁓ when we use a module federation, we have more options. As we saw, we can just pass problem.

Properties to it, ⁓ to the to the our federated component. ⁓ And we should be able to share the state between both the host application and the remotes because both are running on the same instance of React, ⁓ which means that they ⁓ they both have access to the same context providers. ⁓ This also means that we can use the traditional method of just

Pulling the state up the tree and pass it down to a mic to one of our microfront ends, even if that code for the microfront end runs in a separate service. ⁓

We have the options to do custom events or using a message bus. ⁓ And I guess the difference here depends on the nature of what is what we're trying to do. So if you were trying to piece a ⁓ piece of client-side state that has a source of truth, maybe you would just want to use state. But if you're trying to communicate something that happened, you can use an event. ⁓

Sneha Mehra (00:37:40)  
And finally, we have the options of using something like signals ⁓ or atoms, which is a nice way to decouple your microfinance from ⁓ the rest of your application. And the reason this is important is that if we use a lot of props ⁓ to to communicate with our microfinance, those will start to get coupled together. And you may start to treat these microfronts as regular components. ⁓ And ⁓ you shouldn't do that because the ⁓ the they have they serve different purposes. ⁓

Component is about reusability. You want to be able to reuse that in multiple contexts. But ⁓ a microphone is about independent independence. ⁓ And the more you rely on props to make sure that that component communicates with other microfronts, then the less independent your microfront will be. So if you want to keep it independent, don't rely too much on props. We are going to see an example, we're just gonna see a simple example of using ⁓

Counter for example. Let me see here. ⁓ I think I have one implemented here, so we're just gonna go over it. ⁓

Sneha Mehra (00:38:49)  
Let's ⁓ confirm that the application is running and that I can render both of my components. And I have it already implemented here. This is a very simple counter example. So on my dashboard, which runs on the host, ⁓ I've defined ⁓ a piece of state using use state, which is my account. And then I am showing the account, I have an increment button that increments the account.

And I'm also passing those pieces of data to my federated microfront here. This is my federated ⁓ chart component that I'm importing from the analytics app. And I'm passing that as regular props, right? Passing the count and the function that increment. And the nice thing about doing this is that ⁓ no matter how I increment this, I can increment it from here.

Or I can argument it from here, ⁓ and they will share the same state, even though these are two different applications and the code is deployed separately. ⁓ We have the option of sharing the state because ⁓ module federation is ⁓ making my microfront part of my overall React tree, right? ⁓ And because ⁓ they share the same instance of React, they don't have any problems doing that. ⁓

Sneha Mehra (00:40:17)  
Okay.

we're gonna do one final exercise about this, ⁓ which is about ⁓ using a different method for sharing state. Now we're gonna be using a tool called nanostores. And this is the the whole motivation for doing this, for sharing state in another way to what we have now, is that in this scenario, the state is owned by ⁓ my my host application. It's if the state lives here in the dashboard. And I'm passing it down to ⁓

to the to the application here to my microfront. ⁓ In some cases we may want the microfront to own this the state, right? We might want an acro a microphone to hold its own state and and ⁓ and share that with the outside world. Like we we want the the sort of the opposite direction.

And with that, the one good option since we want to keep things nice and decoupled is using some sort of signals or atoms. And there are different ways of doing this. ⁓ we're gonna use a tool called nanostores, which is it's a pretty popular library because it's very small and simple to use.

That let us define these ⁓ atomic like stores that we can use to share pieces of data or pieces of state between one component and another, or in this case, we're gonna share it between one microfront ⁓ and the it could be other microfinance, or it could be other hosts. And what we're going to do is we're going to replace these counts that we defined here. ⁓

Sneha Mehra (00:41:54)  
With the the account coming from the atom. The state won't come from the component itself that lives in the dashboard. The state will come from the federated component, right? The state will live there. And we're gonna consume that state from here. It could be from the dashboard, for example. And for ⁓ simplicity purposes, what we're gonna do is we're gonna register, we're gonna console log all the changes to that state.

Sneha Mehra (00:42:25)  
One thing to keep in mind here is that now that state will become asynchronous because we're importing it from a microfront end at runtime, which means that we ⁓ to access it, we will need to do the same things we do when we import a microfront end and importing using the an import statement. ⁓ Once we have that, we can console log the changes to that store. So what you'll need for this is you'll need to install the nanostores and the nanostores React ⁓ libraries.

And you can use as a baseline the this branch called module dash federation dash two point five. This contains sort of the latest version of what we were seeing up to this point. So you can use this as a baseline and then replace that piece of state with state ⁓ coming from your macro front end rather than state that is owned by the ash the dashboard. ⁓

Do about ten minutes? Yes. ⁓ Sounds good. ⁓ man.

—-------------------------------  
Sneha Mehra (00:00:50)  
We are back. ⁓ All right. ⁓ how did that go? What types of ⁓ subdomains did you did you find? ⁓

Yes. ⁓ Orders, it's like a core domain. Mm-hmm. Probably catalogue. ⁓ Okay, yes. ⁓

Yeah, ⁓ I mean for generic I say ⁓ auth and analytics. ⁓ Mm-hmm. ⁓ Mm-hmm. ⁓ Yes. And there's also like a design system, which is probably generic. Okay, yes, definitely. ⁓

Sneha Mehra (00:01:27)  
Alright, these are the ones that I have ⁓ that I d identify. ⁓ Core, I have two order management and product catalog. These are the things that ⁓ I that make my application sort of unique and the things that I'm gonna compete on. Then supporting, I have inventory, customer management. ⁓ I put analytics here. And if for something like a design system, I will say it depends on whether that ⁓ that is

If you have a custom design system, I would put it probably maybe in a supporting subdomain. ⁓ If it's a generic one, we'll probably put it on the generic subdomain. ⁓ And then the generic I have user management, roles, permissions, authentication, and so on. ⁓

We we probably won't agree on the the split here, all of us, because we really don't have enough information about what makes something a core subdomain supporting or not. And for the purposes of this workshop, it doesn't really matter if something is core core or supporting. ⁓ but in a real application it might help you define, okay, I if something is core, then I might need more tests. I might need better boundaries around that than if something is is supported. ⁓ Or when we move to microfon.

⁓ The core ones are the ones that are more likely to have their own teams working on them. So these are the ones that are more likely to be sort of split into their own, into their own independent applications. The ⁓ next step was defining a folder structure. And here you can, I don't have like a one-way that I ⁓ well, I have one way that I like, but I don't think it matters so much as long as you keep it consistent.

So ⁓ for the one that I like to use is defining first of all my modules and which is this we have one module per subdomain that we identified. ⁓ I like to split that into screens. ⁓ talking obviously about a front-end application, I like to split that into different screens. Then I like to split those screens into features and then into components. ⁓

Sneha Mehra (00:03:28)  
In a module, ⁓ you're gonna have other things other than UI, of course. I don't have a prescription on how you should do this. I like to keep those things ⁓ in an API folders, hooks folders, and so on. So ⁓ you you can choose whatever works works for you. So I have a branch here. There is a branch in the in the monorepo, sorry, the monolith repository that you can check out to see.

⁓ to see how the structure of this code base looks in a modular monolith. So you can check out the branch modular dash monolith. Let me stop this just in case. ⁓

Sneha Mehra (00:04:12)  
And that will ⁓ that has the folder structure that we were just describing. Well we have now we have a modules folder. And inside of those module folder, we have all these different modules that we identify out of our subdomains. ⁓ We have analytics, we have our catalog, our orders module, and so on. When you look at one of the modules, you can see like all the different pieces from our folder structure, like the different screens that we have, different features and components.

Again, the structure itself doesn't matter as long as you keep it consistent, because keeping it consistent will help you sort of define ⁓ what rules you want.

Sneha Mehra (00:04:50)  
A quick question. Yes. Would customer management be considered core since orders depends on it? Customers management core since orders depend on it. I will say it's not really about the dependencies, it's more about how ⁓ unique or how important is that it for your core business problem. So if customer management is something that you compete on, if you want to compete on being

The e-commerce platform that has the best customer management, ⁓ you could put that as core. In our e-commerce case, I wouldn't say it's core, ⁓ but if you were building a CR CRM, it would definitely be a core subdomain. Yeah, good question. Okay, so we ha we're in the monolith ⁓ module monolith one branch. Now we only have one problem, which is that

We really haven't improved anything, right? We still have the same problem we have before, where if I try to reproduce the bug that I showed you earlier where I f I changed something ⁓ in ⁓ I think it was I changed something in orders and that and and ended up impacting customers, ⁓ that can still happen here. ⁓ And that is because we only took care of one part of modularity, which is the organization. But we're missing the other piece, which is encapsulation. And you need both to to have a good modular system.

Now encapsulation ⁓ is something you're probably already familiar with. It comes from object-oriented programming, but we have a lot of levels of encapsulation in code. So a classes is the best example, probably, because we can define ⁓ some things that are private, some attributes are private, or functions, methods, and those are encapsulated inside the class. There is nothing outside the class that can that can really touch this private variable.

With functions, we have similar concepts with closures, where the closure kind of makes it impossible for something outside of this function to access my animal variable that is defined in this function. And also at the module level, if you think of this as an ES module, we also have some level of encapsulation because unless we export something, we don't make it available to the outside world. ⁓ But we don't have the same type of things ⁓ for when we're talking about architecture.

Sneha Mehra (00:07:02)  
So we need we need to create these boundaries ourselves. ⁓ in other languages we have things like Arch unit and things like that that help us define these architectural rules. ⁓ we're gonna see ⁓ some options we have in in front in JavaScript code bases. But essentially the types of boundaries you want to create are typically you wanna create a module boundary, so we wanna make sure that ⁓ your modules, your sorry, your your entire module can depend on something that is ⁓ defined in another module. You wanna ⁓ you wanna

make a rule that sort of disallows this type of of ⁓ of dependency. You can also ⁓

Have boundaries around your layer. So if you think of your folder structure as having your application split into different layers, right? We have only like features and components here, but you can split them different layers and you can say ⁓ that ⁓ the dependency should only go in one direction, which is down. But you don't want ⁓ a component to import from a feature, right? You want to have a rule or a boundary that's that prevents something in this layer to depend on something on a layer above.

⁓ for information hide hiding, what you can do is because let's say you actually want some communication with these modules.

In an ideal world, they will be separated completely, it will be completely independent and decoupled. But in the real world, you actually ⁓ you might have to depend on some parts of it. ⁓ But instead of letting one module reach out and grab whatever they want, ⁓ what you can do is you can define like an interface, right? So instead, let's say you have your module here, another module here, and you have a piece of it that you need, but instead of trying to do this and reach out, you can define an interface on one of your modules and say, okay, you have access to this. So and the interface

Sneha Mehra (00:08:44)  
interface is available. ⁓ So you can actually depend only on the interface, nothing, nothing inside of the module, right? ⁓ So you can define those rules and we're gonna see some examples of how those rules translate into code. ⁓

Like I said, in other ⁓ languages we have things that are built into the languages or the frameworks. In JavaScript, ⁓ unless we use a particular framework, we don't have anything built in, so we have to like create these things ourselves. So ⁓ the one of the most popular approaches is to use ⁓ ESLint rules, ⁓ ESLint rules, which also work with BIOM, OXE, Vitplus, anything that supports ESLint.

no restricted imports is a built-in rule that you can use to restrict an import. even it says no, it's a confusing name, but you can actually restrict an import. They say you have a deprecated module, you can say, okay, you you are not allowed to import this module. And then we have a series of plugins. there is the import plugin that helps you manage these dependencies, and the boundaries plugin, which is the one we're gonna use today, to ⁓ to define those boundaries. ⁓ Then ⁓

There are other alternatives. If you don't want to use USLint, ⁓ a library that I really like is called Dependency Cruiser. ⁓ And also when we're working with monorepos, then you ⁓ workspaces and monorepo tooling help us there.

Sneha Mehra (00:10:08)  
So let's see an example. We're gonna we're working now ⁓ on this modular monolith branch. ⁓ let me make sure this is

Sneha Mehra (00:10:20)  
And let's start by installing, we're gonna install the ESlint plugin boundaries library, which is the library we're gonna use to ⁓ define these rules ⁓ of modularity in our system. ⁓ So you can do that by ⁓ npm install ⁓ esLint plugin boundaries. And you can say that that should be installed nice and easy. ⁓ And now we're gonna update our ESLint config ⁓ to define some of those rules.

So we're gonna start by importing ⁓ the boundaries plugin.

We're gonna ⁓ use that plugin here. So in plugins we have to ⁓ define that we actually wanna use that plugin.

And then in our settings, ⁓ we there are two pieces to use in this library. One is defining the elements of your architecture. ⁓ These are sort of your building blocks that you're working with. ⁓ And ⁓ and the other one is defining the actual dependency rules. So let's start with defining. Let me ⁓ turn off the cursor tab so that it doesn't distract me. There we go.

Let's start by defining Sorry, what was that command again? I missed ⁓ it. ⁓ sorry. Because I was ⁓ No, the command, ⁓ the import command. I missed that. this one? Sorry. Yeah. Import boundaries from ESLint plugin boundaries. Okay. ⁓ You don't need to you don't need to install that? Yes, we you can install it with we have the script here.

Sneha Mehra (00:11:54)  
MPM install. Okay, I got it. ⁓

Sneha Mehra (00:12:04)  
So in our settings, we're gonna start by defining these different elements. So ⁓ the way we do that is we define these boundaries elements ⁓ object. ⁓

Here we're gonna define two types of elements. ⁓ One to represent all of our modules, that will represent all the files in the module folder here. And another one to represent our shared dependencies, everything that is in shared. And the rule that we want to say is that we don't wanna allow anything in shared, anything in shared, because that is in a on a layer below to depend on something from the modules layer, right? This is the cross layer dependency, one of one of the cross-layer dependencies that we talked about that we wanna ⁓

We wanna disallow that that dependency. ⁓ So we have to define two types. We're gonna start with the type module ⁓ that represents all the files in my modules folder. So we call this modules. ⁓ And then you have to put a pattern to match those files. So we're gonna match all the files in ⁓ source, modules. ⁓

Sneha Mehra (00:13:12)  
One pet pip that I have with this library is that the default pattern matching algorithm it's it's weird. So ⁓ what I recommend to always do is also set mode full here, because that is the way that you will ⁓ it will work the way you expect it when finding files in in your in your code base. Otherwise it might only match folders but not files, or it might match

files in the directly in the folder but not within subfolders and so on. It's it's kind of weird. ⁓ and the other one we're gonna just copy this and we're also gonna define I'm missing something here. this is an array, sorry, not an object.

Sneha Mehra (00:13:58)  
And the other we're gonna define another element for our shared folder.

Sneha Mehra (00:14:07)  
And now we can define rules between those two elements. ⁓ and the way you find the rules is in the rules section of your yes link config you set boundaries, ⁓ dependencies. ⁓

Sneha Mehra (00:14:24)  
The first argument is what you want to do when you catch one of these ⁓ dependency ⁓ violations and in this case you wanna throw an error. And let's say the default ⁓ to disallow.

So by default, we're gonna disallow relationships between these two these two modules. ⁓ this won't disallow any dependency, only the ones ⁓ in your module. So this will disallow modules to shared, shared to modules, modules to modules, and shared to share. Like all those different combinations, every everything will be disallowed. So let's test. This should give us a bunch of errors because we do have dependencies from modules to share, which is allowed. But ⁓ if we run npm run lint.

We should see a bunch of errors. We have two hundred and ninety-three ⁓ errors, which is exactly what we want, right? Because we do want to allow modules to depend on shared.

Sneha Mehra (00:15:21)  
There are two ways we could go about this. We could s we could say by by default I'm gonna allow everything, which will get rid of all the errors, right? And then I'm gonna keep a list of everything that is disallowed. ⁓ you can go that way is easier for sure, but the problem with that is that this won't catch as you introduce new elements later, this won't catch all of those different ⁓ combinations that you can have. So I recommend keeping it disallowed by by default because it

When you add a new element, it will force you to think about ⁓ the relationship between ⁓ other elements you already had ⁓ with the new one that you just added. So we're gonna keep everything disallowed. ⁓ And now we're gonna define actual rules. ⁓ So there's a rules property you can define, ⁓ or you can define an array of rules. ⁓ So I'm just gonna copy this so you have to see me type.

Sneha Mehra (00:16:15)  
But I'll leave it up in case you want to follow along. And I have two rules. I'm saying that from elements ⁓ that are of type shared, that's all my files in my share folder, I can only I'm only allowing importing from other share elements. ⁓ And from modules, I want to allow importing from modules and from share as well. So if we run the lint now, pm run lint. Oops.

Sneha Mehra (00:16:47)  
We only get one error, which is an actual problem that we have, because we have a file in a share folder. So the error says that elements of type share cannot depend on elements of type modules, right? This is the boundary exactly that we're that we're checking. And that's because we have a file here in the share folder.

Sneha Mehra (00:17:07)  
That is depending on something in my authentication module, right?

So this is the type of relationship we wanna disallow because I might change something in my authentication module and that ends up breaking authentication for everyone because this is being used in a sharefolder. Can you slow down a little bit? I thought it was just me, but I think people in the chat are saying the same thing. I'm I'm sorry. Yes, definitely. Okay. Thank you.

Is there anything I can I should go ⁓ over again? ⁓ I think just a little bit on that screen. ⁓ and then why you're doing the certain things that you're doing would be helpful. Of course, yes. So what the rules that we are defining here ⁓ are essentially these dependencies between these two elements. So if ⁓ I I'll leave this up here.

But what we're saying here is that our elements

Sneha Mehra (00:18:08)  
This

Sneha Mehra (00:18:12)  
We have our modules here in one layer, like if this is like conceptual layers, right? We have our modules and everything that is ⁓ shared, which can include our design system, it can include our share folder that we have here. This live in another layer, right? And we want modules which live in the layer above to be able to depend on things in shared, right? That makes sense. But we don't want to allow this relationship from share.

To depend on a module because that could end up ⁓ causing some of these troubles where I share something in a module and that ends up impacting something in my share folder, which a lot of different modules depend on. That is the rule that we are trying to prevent here from happening. We want to prevent something in the share folder or our share layer to depend on something that is specific to a particular module.

Sneha Mehra (00:19:11)  
Does that make sense? Every anyone has any questions? Yes. I have a different question. yes, go ahead. Okay. ⁓ is there any reason we couldn't use like MPM packages ⁓ to enforce ⁓ the dependencies? ⁓ we we can and we will see an example of that when we move to monorepos. Okay, cool. at this point we don't have any pack we just have one package, right? Because we are in the monolith, we don't have that split yet. When we move to monorepos, we're gonna see how that can help us. ⁓ but yeah, for sure.

Sneha Mehra (00:19:44)  
Any other questions? Yes. Good question. using this plugin, ⁓ do you do you feel that it also helps when you need to do some refactoring? Helps the AI in that refactoring effort because it has these rules that you're trying to enforce. So it it it makes the AI code that it outputs better. yes, absolutely. Definitely. This helps a lot because it's these are the those guardrails that AI needs, right? By the a lot of these coding agents like Claude and Cursor, they will by default without

Too much instructions, they will do a change and they will run the lint or the type check or something, right? This always happens. And ⁓ if you have these rules, it will let them know ahead of time that they are that what they just did is breaking one of your rules of your architecture. So yeah, this is definitely, definitely useful for that.

Sneha Mehra (00:20:41)  
And please tell me to slow down any time. I'm I'm sorry if I'm moving moving too quickly. ⁓ all right, so hopefully these rules ⁓ make sense. ⁓ d so this is the representation of that diagram that that was just drawing, where we wanna disallow dependencies going in the wrong direction.

Sneha Mehra (00:21:05)  
And this is what the error that we'll get, which again like this this can happen to an AI. The AI will read the error and we'll try to fix it. ⁓ So the way that we can fix this error ⁓ is I'll say whenever this type of error happens, the easiest way to fix it is by moving whatever was defined in a particular module, which in this case was

Sneha Mehra (00:21:30)  
This authentication auth storage ⁓ module that I have here, ES module, ⁓ to move that to the share folder, right? To make it part of the shared folder so that to make it clear that this is a shared dependency. It doesn't belong to authentication module, it belongs to the share folder. There are some others ways other ways to handle this and we're gonna see some of those ways later. But for now, let's just do the easy way and let's just move that to the share folder. So we're gonna

Grab this and move this out of modules of the authentication module ⁓ to ⁓ the ⁓ shared folder. There's a leave folder there. And you can just drag it ⁓ and ⁓ move it out. ⁓

Yes, let's update our imports. ⁓

Sneha Mehra (00:22:21)  
And if we run the lint again, we should see that we don't have any errors now, right? ⁓ Because now that file ⁓ is in the share folder. It's no longer in the authentication folder. We don't have ⁓ this ⁓ problem anymore.

Sneha Mehra (00:22:49)  
Alright, ⁓ we're gonna do another exercise, ⁓ this time in the code. ⁓ And the exercise is to implement ⁓ one rule, like the one that we just implemented, to prevent ⁓ modules to depend on each other. Right? This is if we go back, let me show you this diagram. ⁓ The rule that we just implemented was the cross there was this cross layer dependencies where we have our share, like let's say our shared folders live here on a layer below.

And we are allowing all of this to depend on share, but we don't want the share to depend on anything above it, right? Anything in our modules folder. Now the rule that I want that we're gonna implement next in this exercise is to disallow this type of relationship with my customer's module, depending on any other of my modules, right? We wanna disallow this type of relationship.

We are going to use ⁓ the ⁓ ESLint boundaries plugin to do this. ⁓ And we want to disallow this relationship. ⁓ I will ⁓ have you go to the jsboundaries.dev website, ⁓ which is the which contains the documentation for this code, ⁓ and you can look up ⁓ the capture ⁓ properties.

Sneha Mehra (00:24:13)  
Which will give you the essentially will give you what you need to implement this type of relationship. Because what the tricky part about this is that if you look at our rules as they are right now, we are allowing from modules to depend on any other modules. ⁓ If we remove this.

This will give us a lot of failures.

Sneha Mehra (00:24:43)  
But some of these shouldn't be failures because we are allowing, for example, ⁓ if I go to my users module and I look at one of my components.

Sneha Mehra (00:24:56)  
Greens

Sneha Mehra (00:25:03)  
Let me find an example of

Sneha Mehra (00:25:14)  
Well, I guess these are examples, right? Where we have this is a file in the user's module, and it's depending on another file, this component. I'm sorry, this component that is part of the same module. But this type of relationship will be forbidden by our rules because we are not allowing modules to depend on other modules, even if those two modules are the same one, right? That's essentially what we're doing here. We're saying modules can import anything from the modules folder, even if we're in the same module.

So what we want to do is we wanna update this rule so that ⁓ one module can depend on another module, but your the module can depend on itself, right? On on features ⁓ implemented inside of the same module.

Sneha Mehra (00:26:01)  
⁓ so yeah, that's what we're gonna do. So let ⁓ me know if you have any questions ⁓ and yeah, take a few minutes and and try to implement this yourself. ⁓ What how much do you wanna do for this one? ⁓ Let's say ⁓ ten ten minutes, okay? Yeah, ten minutes. Ten minutes. I'll leave your screen.

Sneha Mehra (00:42:35)  
And ⁓ you're back. ⁓ Alright. ⁓ How did that go? ⁓ Were you able to implement that ⁓ dependency rule ⁓ from modules to other modules? ⁓ I'll show you the the way that I've implemented. Like I I gave you a hint of using the ⁓ excuse me, let me find it where it's here. ⁓

Sneha Mehra (00:43:07)  
Of using the capture ⁓ property when defining your elements, so you can capture ⁓ what module you're talking about here. Because this is matching all files in my module folder. But I want to capture what module I'm working on. ⁓ And the way ⁓ one way you can do that is you can update this to say that you wanna capture.

Sneha Mehra (00:43:32)  
In this case, I'm gonna capture the module name. This can be anything. I'm just saying, I'm just calling this property module name. And what I'm gonna grab here is gonna I'm gonna grab the first directory under modules. ⁓ And this first asterisk will become my module name ⁓ once this library sort of evaluates my my file directory. So with that, ⁓ you can now update your rules here.

And instead of saying that modules are allowed to import from ⁓ all other modules, ⁓ you can define a specific rule. ⁓ So we can turn this allow ⁓ into an array. And we'll keep a rule that modules are allowed to import from share, right? That we we still want that. ⁓ And then we add another rule. ⁓

That's saying that imports from modules to elements of type ⁓ module ⁓ should also be allowed, but only ⁓ if ⁓ what I capture ⁓ matches the ⁓ if the module names ⁓ of the file that I'm importing from ⁓ and the module name of the file that I'm importing to ⁓ match, right? That's that's what essentially what we're doing here. ⁓ So we're gonna say module name. ⁓

matches and here the ⁓ syntax is ⁓ a bit like this. It's ⁓ from ⁓ capture ⁓ module ⁓ name.

Sneha Mehra (00:45:10)  
If we run the lint now, npm ⁓ run lint. ⁓

Sneha Mehra (00:45:18)  
We still get one hundred and six ⁓ problems, which is not ⁓ what I'm expecting, so let me see what happens. ⁓

Sneha Mehra (00:45:39)  
I don't know if I have any syntax errors here. ⁓

Sneha Mehra (00:45:50)  
Yes, I can. ⁓

Sneha Mehra (00:46:00)  
I think that looks right. Let me I don't know what I was typing wrong here, but there was a difference between these two things. ⁓ To ⁓ type ⁓ I said module instead of modules. ⁓ That was the error ⁓ that I ran into. It's not type module, it's type modules. ⁓

Sneha Mehra (00:46:20)  
And that gives me 31 errors. ⁓ So ⁓ this is compared to 106, I think we had before, when I disallowed the imports from every other from modules to any other modules, including themselves. So we filtered those down to 31\. Now most of these errors are coming ⁓ big are coming from the fact that we have an authentication module, which it's in the modules folder but belongs to ⁓

The belongs to the shared layer because shared is sorry, authentication is a shared concern of my application. ⁓ But I would still consider it a module, which is like a subdomain of my application. So it's a module. Now we could fix this the same way we did before. We moved something to the shared folder, and now since authentication will be in shared, ⁓ then ⁓ anything, any modules can depend on it, and we wouldn't get any errors.

But ⁓ if you wanna keep this structure, if you wanna keep authentication in our module folder, we we can apply like a ⁓ an exception for authentication. So we can say ⁓ that in addition to ⁓ allowing modules ⁓ to import from themselves, you also want to allow modules to depend on authentication. So ⁓ what there are a few ways in which we can do this. One way that I like is to define authentication as its own element, right? So

We can define a type for authentication. ⁓

Sneha Mehra (00:47:51)  
And that will match ⁓ all my files in ⁓

Source modules. ⁓

Sneha Mehra (00:48:03)  
authentication. ⁓ So all the finds in there.

Sneha Mehra (00:48:11)  
Again we use mode full. ⁓

Sneha Mehra (00:48:24)  
And now we can update our rules now that I have this type authentication. ⁓ So since authentication is part of our elements now, it will start matching every relationship between authentication ⁓ and ⁓ itself, and authentication and modules, and authentication and shared. All of these relationships are checked, which means that we have to provide one rule for authentication.

So for my type authentication. ⁓

Sneha Mehra (00:48:55)  
What I wanna allow is ⁓ importing two ⁓ elements of type.

Authentication. I want to allow authentication to depend on itself and also ⁓ anything defined in shared. And now we can add our exception here in modules to say ⁓

Sneha Mehra (00:49:22)  
that from modules we want to allow both importing from shared ⁓ and ⁓ authentication. ⁓ So shared and authentication belong to this shared layer, even though authentication is in a different folder, right?

With those updated rules, if we run the lint again. ⁓

Sneha Mehra (00:49:46)  
we will get only eighteen problems, right? We we're filtering those down. Let's see what we got here.

Sneha Mehra (00:49:57)  
So a lot of these problems are coming from types. ⁓

Sneha Mehra (00:50:04)  
So I have this ⁓ users module. This is a file in my users folder. And it's importing a type from the settings folder, from the setting module, right? We can also fix this by saying, okay, I don't wanna like this ⁓ relationship, so I will move this out of to a shared types folder. ⁓ but we can also say,

Types are okay. We can say importing types around the module is okay because if there is a problem, I will get those errors at build time, right? Types serve as a way of ⁓ checking these things before we actually build the we actually build the application. So we have the option to allow all dependencies of kind type, ⁓ which is the way we're gonna do. ⁓ We're gonna fix these problems. ⁓ you can fix this

Also, like I said, we're moving the types to a different place, but for now we're just gonna we're gonna allow those types of dependencies ⁓ by saying that we're gonna allow dependency

Of kind ⁓ type.

And that will give us ⁓

Sneha Mehra (00:51:18)  
Only two problems, right?

And these two problems are the problems that were causing the bug I show you at the beginning, right? That ⁓ or at least one of them is, which is that

This file ⁓ in my ⁓ customers folder ⁓ is depending from ⁓ something important in the orders ⁓ folder, or in the order or the orders. ⁓

Sneha Mehra (00:51:47)  
in the orders module.

Sneha Mehra (00:51:56)  
And we can fix this. We're not gonna fix this in this case, but ⁓ and we want to fix this, we can use some of the strategies we saw before. We can define an interface ⁓ to say, okay, orders. You're not allowed to depend on any on everything inside my orders for you. I'm sorry, yes, there's a question. ⁓ how does it know it's a type kind? Is that just like checking if it's ⁓ a odd? Because it's the import yeah, it's the import type. ⁓

It it's this kind of imports. ⁓ import. So we're allowing this type of import of types of other other modules, right? Gotcha. So it's not based on like file structure. Exactly. Yes. So by default it will also check types. You say no, you're not allowed to import a type from another module. That's why we were getting all those errors before. ⁓ but we have the option if you wanna have ⁓ not as strict ⁓ rules, we can say okay, types are okay because if something breaks ⁓ in

I will catch it at build time. ⁓ Right. ⁓ and then can you go back to your config just to like Yes. ⁓

Sneha Mehra (00:53:05)  
This is my updated config. This is to allow types. ⁓ And this defines relationships between authentication and everything else.

between shared and everything else ⁓ and ⁓ modules and everything else. ⁓

Sneha Mehra (00:53:31)  
But yeah, this is catching the ⁓ cross-boundary problem that was causing the bug that we had the beginning. So at least now we have a way to know at build time in CI or for an AI agent to if this is a feature that an agent implements, they will get this error ⁓ so they can they can fix it, right?

Sneha Mehra (00:53:54)  
I want to show you ⁓ a ⁓ different li like I mentioned earlier, ESLin rules are not the only way we can do this. ⁓ one library that I like is called Dependency Cruiser, ⁓ which is used to build, let me Google ⁓ the dependency ⁓ cruiser. ⁓

Usually people use this to build graphs, like ⁓ diagrams of all the different dependencies, how how they how how they show up. Of course, this only works if you have a small ⁓ application or a small folder where you have only a handful of files. If we try to run this in our code base, it will we will have so many things that it will it wouldn't make sense. ⁓ But this also serves for doing these types of checks. ⁓ So ⁓

I have one example here in the there is one example configuration file that you can use if you want to see how these different dependency cruiser rules look like. ⁓ This is the same rule that we implemented in the exercise of this allowing to from one module to depend on the other. This is how we implemented in dependency cruiser. And we're also here making the exception for authentication. ⁓

And then we have other types of of rules that you can you can see for different the different types of rules we can we can create in architecture. ⁓

There is an NPM run ⁓ lint ⁓ depth what was it? Lint depths ⁓ that runs dependency cruiser. ⁓ And you'll see that we get the same two ⁓ errors that we ⁓ that we were getting with DSlint config. ⁓ It's just reported in a different way, right? Because we have the same rules.

Sneha Mehra (00:55:37)  
There are some interesting rules as well in dependency which you can do only with dependency cruiser, which is defining a shared ⁓ set of minimum dependence. So if you have something like let's say you want to share, ⁓ you want to make sure that files in your shared folder are actually shared and not just you put it there just in case, you can check that at least two ⁓ at least two files depend on that file in the shared folder. And if not, this will give you an error.

And there are others you can ⁓ you can check as well.

Sneha Mehra (00:56:14)  
Yes. What's the impetus behind like number of dependencies? ⁓ Is it to check that it shouldn't be in shared, instead belongs in like a module? Yes, exactly. Because if you have something that ⁓ is in your share folder, but only one module is importing, that probably belongs in that module, right? From a folder structure perspective perspective, it doesn't really matter. But it helps to keep your module more cohesive in a way, right? Because

that module is likely to change together with sorry, that shared file is likely to change together with the the other module. And when things change together, they probably should be together, right? So we should probably keep those those two things cohesive. ⁓ yeah, great question.

Alright, to finish with this, we're gonna this is our updated character sheet for this modular monolith architecture that we have. ⁓ we have different sets of logical components now. We're not talking about just hooks and ⁓ and ⁓ and routes, we're talking about modules. We're talking about screens and features. So we changed the building blocks. We also have ⁓ the characteristics switch a little bit. Now this application is not as simple as before, but it's a bit more scalable.

And we made some architectural decisions, right? We set the rules of how ⁓ how the communication between these modules should work, we set some boundaries.

Sneha Mehra (00:57:43)  
And we fix some of these growing paints, right? We fix the ⁓ some of the pain some unclear boundaries ⁓ and lack of organization because now we have a nice modular organized code base, but we still have some of the other paints. ⁓ the mainly this shared folder will become like a damping ground for all the stuff if we keep fixing things in this way, of moving everything to shared. ⁓ we still have some problems on monoliths, like ⁓ updating a dependency becomes an all-or-nothing kind of thing.

And ⁓ as the team grows, the CI pipelines are starting to become stuck in a way. ⁓ Start to take longer and test as well because the code is growth, you have more things to test, but also you have this problem of multiple people want to release at the same time. And now we have this queue of releases ⁓ because everything is deployed ⁓ in the same way.

And with that, we we're gonna take a look at how monorepos can help us ⁓ solve some of those those pains.

Sneha Mehra (00:58:48)  
A little update on our project. So now our company grew a little bit. We have not just a handful, two or three engineers, we have ⁓ you know teams ⁓ that are working ⁓ on ⁓ on different parts of the application. We've

kind of starting a pivot two AI maybe where you know AI is ⁓ it's consuming SaaS business, so ⁓ it's very likely that we're gonna have some impressions with AI. ⁓ and that requires you take a look at how we can improve this architecture to fit those demands. Yes. ⁓

When a module depends on itself, is there a chance of circular dependencies to get formed through barrel file imports? What's the recommendation for avoiding and fixing that? ⁓ Yes, great question. There is a chance of circular dependency. And there are there are tools we're gonna see now in in this monorepo section, we're gonna see a tool that exposes those circular dependencies. But ⁓ it's not always the case that ⁓

That you will have a circular dependency when you have a module that ⁓ depends on itself, because ⁓ that module may be in a completely different tree. Let's let's say, for example, let me go back to the slide where I have my tree of my modules, right? ⁓ So ⁓ if we have the case that this component depends on this component, yes, we have a circular dependency right here. ⁓ But if this component depends on something that is in a different branch of the tree. ⁓

we don't we we're not gonna have that problem, right? ⁓ So ⁓ yes, there is a chance, but there are tools that can help us find those ⁓ those problems with this the with circular dependencies. ⁓

Sneha Mehra (01:00:30)  
It's another question I'm trying to ⁓ kinda parse it.

Sneha Mehra (01:00:48)  
⁓ so when you have a utility file with ⁓ some functions and those functions use something like ⁓ a dependency like date Fn, ⁓ when you include that ⁓ dependency or that utils, ⁓ it'll include date Fn correct or as well, right? ⁓ And ⁓ is there a way to like tree shake? ⁓

out, you know, kind of dependencies that go up and up. Mm-hmm. Yes, yes, for sure. ⁓ in this case of this since we have monolith, we don't have to do anything special about ⁓ handling ⁓ shared dependencies, right? Because if I have if my customer module depend on ⁓

on Lodash and something in my sharefolder also depends on Lodash. ⁓ since I have one single repository and one single build, I'm not gonna end up with two versions of Lodash. They will be the same version. They will install from the same packages at JSON, it will be the same thing.

If these were different packages, which is gonna be the case with when we talk about monorepos, then yes we have to take care of that. ⁓ we have to make sure that ⁓ those packages they don't end up each one including its own version of Lodash, unless that's what we want. With monorepos, you also have the same problem because ⁓ yeah, we'll see we'll see the example of monorepos, but we will ⁓ we're gonna have this problem. This is one of the advantages of the monolith. You don't have to care about that stuff because ⁓ everything is built together. Mega, great question. ⁓

Sneha Mehra (01:02:26)  
Okay, let's take a peek at the updated C four diagram for this stage. ⁓

Sneha Mehra (01:02:36)  
The V two diagram.

From this perspective, ⁓ things look the same, right? Nothing really changed except for the fact that we're now like we have a this this arrow is not, I don't know what happened there. ⁓ But we we're now integrating with another service because we have an LLM provider, we have some AI features built into our product, but that's an external service. We don't we don't care about how that works. If we take a look at ⁓ our ⁓ our actual system, we now see that it grew more complex, right?

The ⁓

Or the core API got split into now we have an orders API that has its own database. So they're using some version of microservice, ⁓ microservices on the back end. We now have an AI gateway that let us talk to the LLM providers. ⁓ and we're also exposing a public API and we have a mobile client. We have different types of clients, right? So this is what you typically see as as the company grows. From the perspective of our admin web application, which is what we're working on, things haven't changed too much though.

We're still ⁓ have one single user and we're still ⁓ talking to a single API. ⁓ So at this point we we don't see like a big effect on ⁓ from the architecture on a little application. ⁓

Sneha Mehra (01:03:58)  
We're gonna switch to the front end architecture monorepo ⁓ repo which sho you should have ⁓ installed by the plugin script. ⁓ So let me ⁓ go ⁓ there. ⁓

Front end ⁓ architecture monorepo. ⁓

C D into that and ⁓

Open it with cursor.

Sneha Mehra (01:04:27)  
And at this stage, this is the exact same stage where we leave where we left the other application on. So we're still have we don't have a monorepo at this point yet. We have still the modular monolith that we were working on before, with the same file structure. Nothing ⁓ really changed from what we were just looking at.

Sneha Mehra (01:04:54)  
So now we're gonna see wha how we can transform this monolith into a monorepo. There are two ⁓ ways in which you can arrive at monorepo. ⁓ One is from ⁓ the way we are like going through right now, which is we have a monolith and we wanna decompose that into different apps and packages. And the other way might be because your your code base might already be ⁓ spread all over different repositories and you wanna bring them together.

Both of these approaches are ⁓ valid, you can come from either way, you get to the same result. And the the approach that I like to use for both of these approaches is to sort of break down the adoption into three steps. The first one is bringing all the repositories together into the single workspace. Then we're gonna start to break those break those into ⁓ break the parts into multiple packages within the monorepo.

and then we can adopt monor monorepo tooling like TurboRepo for example to make it make it scale better.

So I'm gonna share now, you can follow along how I will do step one, which is ⁓ bringing together everything into a single workspace. ⁓ In this case, all of our code already exists in a single repo, so we don't have to worry about about it, but it helps to create a workspace ⁓ because that will make it easier to then decompose it.

Sneha Mehra (01:06:20)  
So we're gonna start by let me make sure I install my dependencies.

Sneha Mehra (01:06:29)  
we're gonna start by creating a couple of folders. We're gonna create ⁓ an apps folder.

I'm gonna create an admin folder ⁓ within that apps folder that will contain my entire workspace ⁓ and then ⁓ a packages folder for all my packages. ⁓

Sneha Mehra (01:06:51)  
Then we're gonna move ⁓ pretty much everything that we have here. We're gonna move it into the apps ⁓ admin folder. ⁓ This will be our workspace, so everything needs to be there. ⁓ I'm gonna start, I'm gonna do it in pieces. I'm gonna start by doing ⁓ using git mv to move ⁓ the public ⁓ folder into my apps admin folder. ⁓ And then I'm gonna do the same thing ⁓ for my source folder.

Sneha Mehra (01:07:22)  
So now source and public should both be here inside of admin. ⁓ I like to use git mv because that maintains ⁓ the commit history of these files in these folders, which ⁓ may not always be the case if you just drag and drop things. So I like to just in case use git mv. For the rest, you can ⁓ you can drag and drop pretty much everything. I have a script here, ⁓ but I'm just gonna use it, but you don't have to. That essentially what it does is

It moves everything to my admin folder, right? ⁓ Except for ⁓ gitignore and the package log.json. ⁓ You can do this manually or you can ⁓ do use the same script that I have here, ⁓ but it will do th those things for you.

So the end result should be that everything except ⁓ gitignore and package log.json should now be in your apps admin folder.

Sneha Mehra (01:08:28)  
Once we have that, we are going to create ⁓ a package.json file at the root of our repository. And this will be the root of our of our monorepo.

And here we can define let's define just the basic stuff that you need for this, which is a name for your repository. We'll call it Commerce OS. ⁓

Sneha Mehra (01:08:53)  
We'll say that this is a private repository. ⁓ And then we'll define the workspaces. Workspaces is a concept that all package managers ⁓ support. ⁓ The syntax looks different depending on which package manager you're using. In this case, we're using MPM, so this is how it looks. But if you use PMPM or if you use Bun, Dino, all of those support the same concept of workspaces, ⁓ which is how you keep

those folders, not just folders, but also like an independent application that can be built by itself. ⁓ And here you have to define where your bo workspaces are, which is in my apps folder and in my packages folder. Those folders will contain all of my workspaces.

Sneha Mehra (01:09:42)  
We will port some of the scripts that we have. Let's start with just the ⁓ the dev script, ⁓ just that we can run the application.

And what we'll do when they run the dev script, we will just run the dev script of ⁓ my admin workspace. ⁓ So you can filter bar workspace and say ⁓ that I want to run ⁓ the npm run dev script of ⁓ my workspace in apps admin, which is the one that we just created here and has its own package adjuston.

Sneha Mehra (01:10:18)  
And that should be enough to ⁓ really transform this into a monorepo. ⁓ And let's run npm install to make sure ⁓ everything is ⁓ in the right place. ⁓ And now ⁓ let's run npm run ⁓ dev to make sure everything works correctly.

It looks like our app is up and running.

Sneha Mehra (01:10:42)  
Perfect. ⁓

Sneha Mehra (01:10:53)  
So in this case, this concludes step one. If we were moving from the other direction where we had multiple repositories, we will have to do sort of repeat the process for each one of our repositories. So since we only have one, we are essentially done here.

And now we're gonna go into ⁓ step two, which is start to pull up out the monorepoint to multiple packages.

Sneha Mehra (01:11:33)  
Right. ⁓ And to start with, this is gonna be the next exercise for you. ⁓ To ⁓ pull the shared folder, which is inside of our admin folder, into a separate package, into ⁓ its own package. ⁓ To do this, you'll have to create a folder ⁓ in ⁓ our packages folder. So we have an empty packages folder at this point. So we are going to move the source code from ⁓ admin source shared.

Into that package packages folder. You will have to create and package the JSON inside that folder to actually make it a works workspace. And then you will have to update all the imports because we have a lot of stuff that is important from shared, ⁓ but you can use search and replace. and I will you can you yeah, you can just use search and replace. ⁓ let me just show you an example of ⁓ how.

You have to search for this pattern ⁓ and actually change it to the however whatever name you gave your mono re your new package in the package.json. ⁓ So let's say you create a package.json ⁓ that names your package ⁓ commerce oslash ⁓ shared, ⁓ then your search and replace will just simply replace this, the ⁓ at slash shared with commerce OS.

That's the the what we want to do. ⁓ yeah, don't forget to install it. And you will probably run into a tile wind, staling issue when once you do this. ⁓ Don't worry about this, we'll we'll we'll fix it later. So with that, please take a few minutes and we'll come back. what do you think? Five or ten?

Say can we do ten? Yeah. Yeah.

—--------------------------------

Sneha Mehra (00:00:58)  
And you're back. ⁓ All right. ⁓ I hope that went well. ⁓ So to recap what we wanted to do, we wanted to ⁓ make sure that we have some boundaries in place into between the different layers of our architecture of the different packages. ⁓ So what we have right now, we have, ⁓ as we talked about before, we have some sort of

layer at the top where we have all of our modules and then we have a layer below where we have our share dependencies. This is our share package. ⁓ We may have a UI ⁓ design system here. And we could have other things in this layer, right? And we want to make sure that they die the the dependencies go in one direction from our modules layer and not ⁓ from the this base layer up to the modules. ⁓

So this is the relationship we want to avoid. We saw how to do this with ESLint boundaries before, ⁓ and now we're gonna see how we can apply these TurboRipo boundaries to have these layers now that these are separate packages and they don't belong to the same repository.

Sneha Mehra (00:02:07)  
So TurboRepo, as we saw, has this boundaries plugin and we can config ⁓ some ⁓ specific boundaries here in our TurboRepo configuration. ⁓ So ⁓ we can define ⁓ a boundaries object ⁓ here.

That will define the rules just like we define the rules of our ⁓ ES ESLin boundaries. ⁓ And in order to do that, we need to tag specific packages with different ⁓ with different tags for our different layers. ⁓ The way that I have it set up here is that I'm going to tag all of my modules, everything that belongs ⁓ on the the apps folder, with the exception of authentication, because as we mentioned earlier, authentication is really something that belongs.

on the the base layer, the platform layer, right? So we what I'm gonna do is I'm gonna grab one of my modules on the application layer, ⁓ users, and I'm gonna create a toolboard.json file inside of this module.

Sneha Mehra (00:03:10)  
And here all I'm gonna do is I'm gonna define a tag for you. I'm gonna tag it with a tag that I'm just gonna call platform. Or sorry, in this case the tag will be ⁓ modules.

Because this is one of the modules of my application, it belongs on the the layer on top with all my modules are. ⁓ And now on my share folder, I'm gonna do the same thing. ⁓ Create a new turbo turbo.json file where I define a tag, and this I'm gonna call platform.

Sneha Mehra (00:03:45)  
So ⁓ I will do the same for, for example, the UI ⁓ package will also belong to the platform layer, authentication package will belong to the platform layer, and all of all of the others will belong to to my module layer. But for now with these two it should it should be enough to to do this exercise.

So now that we have those two tags, we can define the rules between these two tags. So the boundary that I want to set is that

For my packages that are tagged as platform tags or as platform packages.

Sneha Mehra (00:04:23)  
I wanna set a rule on this in on it there are dependencies ⁓ to deny imports to

The module layer.

Right. So platform layer, that relationship is not allowed. I'm sorry, platform module, that relationship is not allowed. ⁓ And if we run MPM ROM boundaries ⁓ now, well I have no issues because I don't have any instances of ⁓ I don't have my my share package is not depending on any of my modules, but if I try to install in my share package, let's say I want my share package to depend ⁓ on

The

users package which belongs in a different layer. Just gonna ⁓ set this here. And if I run MPM inst ⁓ NPM run boundaries again, ⁓ I will get an error. ⁓ And I will actually detect it two different things. ⁓

Sneha Mehra (00:05:27)  
Well it actually just d d detec detected one thing, ⁓ which is w that we have a c a circular dependency here. Let me see why it didn't detect ⁓ the other problem, because it should have detected

True problems here. And this probably because ⁓ I made the same mistake I did before. ⁓ Yes, I I called this layer modules and here I said module. So ⁓ there we go.

Alright, two errors. This is this is what we are expecting. The two errors come from the same place. ⁓ One of the errors is the the ru the set the rule that we just set that ⁓ we have the share package is dependent on one of my modules, right? And here's my rule and we deny that. And the other one is the problem of circular dependencies, which ⁓ it's nice that it can detect those, which we have

In this case we have the shared package depending on the user's package, which depends on the shared package, right? So we have this circular dependency which can cause trouble in the future. ⁓ So it's nice that we at least get those warnings, right?

Sneha Mehra (00:06:43)  
Other things you can do with TurboRepo. So TurboRepo comes with a bunch of different scripts that you can use. ⁓ one of the cool ones is that you can do generate graphs of your of your how your package looks like. Oops, let me go back.

So we can try this ⁓ graph script that essentially w can tag one of the one of the tasks that you have and it will generate that graph so you can visualize your dependencies and how they work.

I will go back ⁓

Sneha Mehra (00:07:16)  
We'll actually fix this dependency that I have here.

It doesn't cause any trouble.

Run the boundaries again, just no problems. ⁓ And I'm gonna run that script which is mpx run build. So this is the same build ⁓ script that we used before, ⁓ but I'm gonna pass the graph ⁓ script and I can pass here the name of a file, just call it graph.mermaid. ⁓ You can use HTML and we'll generate an HTML graph. ⁓ there are other options as well. I'm just gonna use mermaid here.

⁓

Sneha Mehra (00:07:58)  
Having no trouble. This is not what I want. It's MPX ⁓

Sneha Mehra (00:08:05)  
Turbo ⁓ RAM build. There we go. I was I forgot that but ⁓

Sneha Mehra (00:08:20)  
Hm. No extends key file. I think that's ⁓ maybe running into a problem with ⁓ Let me just remove this for now.

sure what that means we'll just ⁓ create those two files ⁓

Okay. I'm probably missing something in those turbo.json files that we created. we're probably missing the schema or something that is required. ⁓ so ⁓ so I just in this case just gonna delete them.

This is the graph that is generated. It probably doesn't make sense. So you can you can use any Mermide ⁓ previewer. This is supported in GitHub as well. I have a ⁓ I have a browser extension that lets me preview the diagram here.

And this is ⁓ sort of the visual representation of how the different packages in the Monorepo look like and how they depend on each other. ⁓ One nice thing about this is that we can kind of visualize our layers here, right? Even though we can't like we if we have too many packages, it might become too confusing. But still, if we look at this from the top down, we can see kind of our different layers. We have the application shell layer, and then we have our modules layer here, and then everything below.

Sneha Mehra (00:09:37)  
It becomes like sort of sort of part of shared infrastructure, right? ⁓ Here is authentication. Sorry, I can't ⁓ there we go.

Sneha Mehra (00:09:47)  
Here's authentication layer, our share package, our UI package, and tooling finally, we everything depends on. So we can draw our layers here and that will help us define sort of the boundaries that we were defining before.

The Turbo Repo DevTools is another great way to visualize this if you don't want to generate a graph. So we can run MPX Turbo ⁓ run dev tools. ⁓ sorry, not run, just DevTools.

And this will run a server that it opened in a different browser.

Sneha Mehra (00:10:26)  
This will run this, which is another way of visualizing your your ⁓ package hierarchy. And you can also filter by one of your packages and you can see the dependency, sorry, the dependency graph of just that one package. So you can see what depends on what. If you want to look at just one of our modules, we can see that discount depends on everything here. And we don't have any cross layer dependencies, right? Nothing in this row should be highlighted.

Because that's one of the rules that we don't want. We don't want ⁓ anything at this layer to depend on anything here.

Sneha Mehra (00:11:05)  
So that's another nice utility. ⁓ generator, the generate tool it's help helps you helps you if you wanna generate apps or packages in a standard way. ⁓ And query you can use it to have access to the GraphQL query of your your dependencies, ⁓ in case you wanna query by something. You can find what packages have no dependencies, for example, things like that. So it's pretty powerful.

Sneha Mehra (00:11:32)  
Here is how our character sheet looks like for this monorepo architecture. ⁓ so I call it triple SPA. We have a modular monolith in a monorepo. ⁓ And the logical components we now are dealing with apps and packages. In terms of characteristics, we are trading off a bit of simplicity. We are adding more tooling ⁓ and we're getting more scalability. This scales much better.

In terms of deployability, this is still simple to deploy because we're still dealing with one package or one single application. And we've made some architectural decisions about boundaries around our packages.

Sneha Mehra (00:12:13)  
So, an update on the growing pains you might see with this type of architecture. As you can see, this solves a lot of problems. You don't have to reach out for microfront ends at this stage if the growing pains are in this list. But you might still run into some paints that might force you to look for different solutions. Even though we are we fixed the problem of builds and tests taking ⁓ forever ⁓ because of ⁓ because of the COVID grew too much.

And we fix that using Turbo Ripple remote caching, we still might have a problem in a CR pipeline ⁓ if we are ⁓ if we're releasing things too often and the builds take too long. Let's say a build takes fifteen minutes, but you can only release one at a time. So that might cause your teams to sort of create a queue of releases. And if you have many teams contributing to the same code base, that's where you start to see the sort of the growing pains of this architecture.

Another problem is that we can't really share components at runtime with this architecture. We've made the decision that we're using we're composing everything at build time. So if we have the n the need to share components at runtime, this is this architecture is just not gonna work.

So to see how we can solve some of those problems, now we're gonna take a look at microfronts. Yes. Is there still an issue having two separate apps like the Analytics and Orders apps each have a dependency on Rechart but with different versions? For example, three point eight point one and three point five point zero? ⁓ it could be you could run into a conflict if those versions are not compatible with each other.

what will happen if the versions don't match and you have two different apps depending on two different versions that you end up with two the code will be duplicated in your bundle, right? You end up with the two versions of of Richards. And that might be okay ⁓ if they are compatible with each other, but if the versions are incompatible, you might run into issues at runtime as well. From a build perspective, you're not gonna get any errors in this case. ⁓ but things might break at runtime. Yeah.

Sneha Mehra (00:14:27)  
We're gonna see how microphones help with some of that as well. Those cases where you have the same package but different versions, we're gonna see how we can negotiate those things with microphones.

So a project update. ⁓ we've evolved again, now we're becoming an AI first, agent first kind of company, and we also happen to do e commerce. ⁓ let's take a look at the updated C4 diagram of our architecture.

Again, again at this level, we still don't see much difference at the context level because we're still dealing with the same pieces. But if you if we take a pic here, we see that things have grown a bit messy. And we have a ton of different ⁓ services and so on. ⁓ The main difference or the main thing that affects us is now our web application is starting to talk to to many different ⁓ APIs. It's not just a commun a simple communication with a core API. Now we're talking to many different services.

And that means that the back-end logic ⁓ is ⁓ split ⁓ across different services. That's probably because we have different teams working on different parts of the application. So we no longer have a front-end team that handles the web app. Now the web app is managed by different teams, each of which have their own teams of front-end engineers. So when you start to see this type of ⁓ split of your logic across different services, that's when you might

you might want to look at microforens as a solution because each one of those teams ⁓ that manage different parts of the overall architecture may be able to move independently of each other.

Sneha Mehra (00:16:08)  
For this we're gonna be using the front end architecture microfront ends repo, so you can go ⁓ and check that out.

going to ⁓ C D into that repository.

Sneha Mehra (00:16:24)  
And open it in cursor.

Sneha Mehra (00:16:32)  
He said open.

Sneha Mehra (00:16:36)  
there it is.

Sneha Mehra (00:16:40)  
And this is basically the same ⁓ the the same ⁓ version that we were seeing before of our monorepo where we have our apps and packages. ⁓ But ⁓ I made two differences. There are two differences here. ⁓ One is that we now have an API package as well, which contains this is the con this contains a back end API. In this case it's using Hono. ⁓ I have installed my dependencies, which is ⁓ why I'm getting those errors.

Sneha Mehra (00:17:10)  
This is a an an API using Honom. In theory, it doesn't really matter, but the previous version that we saw, it was using mock service worker, which ⁓ runs only on the client. ⁓ And ⁓ migrating that to microfurns c can cause some things that are hard to deal with. So I didn't want to deal with with with those problems. So now we have a former API. It's still mocked, the data is still saved in a JSON file on the file system. ⁓ but it's a real API making real network requests.

And the other difference is that I've changed the build from Vit, which we were using before, to RS build. This doesn't mean that you have to do this, but ⁓ since we're gonna be looking at module federation in this section, module federation has ⁓ better support for RS build than it does for Vit. For Vit, some things ⁓ don't work ⁓ exactly the same. For example, the dev script in build is not supported with module federation. You have to use the

preview build, which means that every time we want to test something, we have to like rerun the preview build and so on. So RS build comes with with support for dev, which makes it much much simpler to work with. So I will say for the best developer experience, ⁓ RS build is the one that has better support. And also because ⁓ module federation comes from like the Webpack team, which is sort of the inspiration for ⁓ RSP, the Rust version of Webpack.

which is what RS build is built on top of. So we ⁓ there's a nice like s ⁓ integration there between module federation and RS build. We're gonna talk about the specifics once once we get to the module federation section.

Sneha Mehra (00:18:56)  
Okay, so what are some of the growing pains that we might solve with microfronts? As we talk as we talk about before, like the the main things are that your team now needs to move independently from each other. And the if you have a single release pipeline, that will become the bottleneck for your team. Teams ⁓ they they don't want to wait on other teams to release their work, so they wanna be able to deploy independently. That is a great that is a great excuse for use microfronts. ⁓

You should exhaust every other option before you get there because that will ⁓ this will just add complexity. But it's a great it's a great reason for having to use microphones. ⁓ another growing pain that might cause you to look at these distributed solutions is that ⁓ some areas of the code base may have run like legacy code and you want to update them ⁓ one by one without having to do this whole ⁓ this whole update of React in in one go. You may want to update parts independently.

⁓ and you might also find that different teams may care about different different things now. If you remember what we talked about before about these architectural drivers, in the at the beginning we might have just ⁓ one set of rules for our architecture. But now different teams have may have different requirements, different things that drive their architecture. ⁓

and insights, then they have more specific constraints and requirements for the whole analytics portion, which might not be incompatible with the rest of your architecture. So when these ⁓ architectural drivers diverge, when different teams care about different things, that might also be a reason why you might want to split your monolith and have each team have their own rules applied to each one of their their applications.

Sneha Mehra (00:20:53)  
⁓ as a rule of thumb for deciding when is the right person where in the right ⁓ time to adopt microfronts, I would say that you should try to look at at the time where the cost of coordinating your monolith becomes higher than the cost of managing the distributed architecture. Both codes will increase over time as your application grows in complexity. And at one point you might cross that threshold where it might make more sense to

⁓ incorporate those complexities of microfronts because it will simplify your coordination in the

Sneha Mehra (00:21:32)  
Some questions you should ask yourself ⁓ to decide what flavor of microfront you want. ⁓ These are not all of them, but some things you should ask at the beginning is first of all deciding how you want to split the app. ⁓ The vertical split means essentially in ⁓ a web application like the one that we have, essentially the ⁓ splitting by route. Like each one of our main sections, like the catalog section, will become its own microfront, ⁓ and and so on, right?

Horizontal split is when you have a single view and you wanna split the the parts of it, right? So if we look at our application here, the vertical split will say, Okay, the whole dashboard is my is my microphone and and I have a different microphone for catalog, for example.

Sneha Mehra (00:22:21)  
And the horizontal speed could be actually I just want this part ⁓ to be some microphone. I want to be able to deploy this independently. And the rest can ⁓ just live in the app shell, right? So defining what type of breakdown do you need, that's probably the same the first question that you should ask. ⁓ other things to consider where does the routing happen? Is this ⁓ is this the routing on the server or on the client, or it happens on the edge? ⁓ same with the rendering. Typically it's the same.

Is the same place where the routing happens, but not always. ⁓ And ⁓ where do you want to compose these microfrontets? Do you want to compose them ⁓ on the client or at runtime, on the server at runtime, or at the edge, or at build time? Build time ⁓ is some people consider build-time microfronts as a thing. For me, I don't think there's a description ⁓ a distinction between build-time microfronts and what we have right now, right? What we have right now is essentially

What some people consider build-time microphones. But we still have ⁓ one single deployable unit, so I really don't consider that. ⁓ And finally, how much ⁓ isolation those components need, which is particularly important from the perspective of the horizontal split where you may have ⁓ you may not want one part of the app to to even know about what's happening outside of it, and you don't want anything what's happening outside to leak into what's happening in your microphone.

you need full isolation, that's something you need to consider as well.

So, what are some of the different ⁓ flavors that we have with microfronts? ⁓ one, there there are some built-in ways that that you don't need any frameworks. you can use iframes to, especially for the horizontal split, you can use iframes to separate your your part. And each of those iframes will point to a different application which can be deployed independently. And you have microfronts there. There's no ⁓ no hidden magic behind it. Web components help with the isolation piece as well.

Sneha Mehra (00:24:21)  
You might want to com ⁓ have a runtime composition of those components, but you want them to be isolated as well, then that's a good solution. There are some other ⁓ methods that use native native browser or native ⁓ platform features like edge size includes and import maps. ⁓ and then there is a whole section of ⁓ of ⁓ libraries that support what I call fragment orchestration.

SinglesPa or singles PA is one of the pro probably the most popular ones. At this point, I think that library, I wouldn't recommend looking into that ⁓ because it's largely unmaintained at this point. At least that's what I've been seeing in the in in GitHub, like exploring the how people are maintaining that library. And you if you need some compatibility ⁓ issues, you have some compatibility issues and you want to look into that, you need to support singles PA. That's a good option. But

If you're building something new, I'll try to stay away from it from it. ⁓ There are some newer versions, new newer frameworks like Pyrol ⁓ and Cloudflare has its own version of fragments that you can use. ⁓ And then if you need route-based orchestration, Cloudflare also you can use Cloudflare workers. And NextJS has something that what is called multisones that support microfronts. And finally, module federation or native federation, which is almost like the Angular specific version of module federation. ⁓

Are one of the most most popular ⁓ alternatives for microfronts. ⁓ My cheat sheet is ⁓ if you only need a vertical split, you should do try your best to do it like do-yourself ⁓ solution, which we're gonna see next, ⁓ how that looks like. ⁓ with just ⁓ orchestration at the route level. It can be on a server, like an you may have an N an Nginx server, you may have a node server or whatever.

And then you just point different routes to different applications, right? If that's what you need, if you need ⁓ the teams working on different routes to in deploy independently, that's probably enough. ⁓ if you need full isolation, iframes are the best way to have that full encapsulation that you know that nothing is gonna affect them. ⁓ so even though they're kinda clunky to work with, that's that's a good option if that's what you need. If you're using NextJS, multi-zone is a great way to have like this built-in support. ⁓ this is

Sneha Mehra (00:26:44)  
Inverse we're talking ⁓ a versal feature here, but it's a great way to support those multiple applications in the same server. And then if you want to experiment with something new, I really like the spiral framework that you can use. Otherwise, module federation is always a ⁓ safe bet. It's widely supported, it works everywhere. ⁓ it works it supports SSR, it supports a bunch of different features. So that's what we're gonna we're gonna be using today.

Sneha Mehra (00:27:15)  
Before we move on to module federation, let me show you

a branch that I have prepared here where I show you how this do-it-yourself approach ⁓ looks like. So I'm going to check out a branch, the branch DIY. ⁓

Sneha Mehra (00:27:35)  
And I'm gonna run it.

Sneha Mehra (00:27:44)  
You'll notice a few differences here ⁓ compared to what we had before. I'm running now four different processes. This first one is my API. This is the new thing that I told you about before. This is a simple REST API. The app shell is running on port 3000\. And then I'm running the dashboard ⁓ app and the analytics app in different ports as well. ⁓ So this means that these applications are not being ⁓ served independently and

The app shell doesn't depend on them. We can render the app shell by itself without depending on ⁓ these other applications. ⁓ And the way that I have it set up here is I have what could be considered a horizontal microphone end ⁓ inside of an iframe here. This is simply an iframe that points to whatever is where my dashboard app is located. These two applications can be deployed independently.

If the dashboard team needs to move faster, they can do so and they don't affect the the don't affect the CI pipeline of the AppShell or the main application. And that might be what you want. If this is what you need, then ⁓ stop here. Don't go into like more complicated architectures. ⁓ If you need the horizon the vertical split, there's also a very simple way to do this. And as you can see here, when I change my routes, everything is happening client-side because I'm changing between routes of my main AppShell application.

And we if I go to analytics route, this looks the same. I'm still even on port three thousand. But I have a setup that I can show that it's very simple. It's just a proxy that says that ⁓ I should proxy the request to analytics to a different application that lives in a different port.

My analytics app lives on port 3001, which is why I can open it like here by itself. And this is another application that can be deployed independently. And all I'm doing here ⁓ is I am proxying these requests from this ⁓ this service to this other one. And again, this is these are two apps that are completely independent from each other, even though I can still ⁓ navigate between them seamlessly in the UI. ⁓ Yes, one question.

Sneha Mehra (00:30:02)  
⁓ I think in the so the analytic analytics apps a separate app. How are you managing ⁓ the selector for like the account owner and preserving that state between apps? The selector here, this is ⁓ both apps are being deployed independently, they run independently, but they they consume the same packages, right? Both apps consume the same packages. And what I'm doing here is that this is I'm not just running the analytics portion that you see here.

The entire view is the analytics app. So ⁓ the layout that you see here also comes from the analytics app. It just happened to be a shared component between my two applications. The main application, which is what runs all of these other apps, and analytics app. They both consume the same shell, the s the same layout. Does that make sense? Yeah. I guess I'm just curious, like is that

select a shared state between apps. Like if I select the second option and I go to dashboards, ⁓ you know, is the intended experience persistent. I see what you mean. The persistent of this. ⁓ not in this ⁓ did it work? ⁓ Maybe I think it worked. it worked because the way that I'm persisting the state in this particular example, which is just using local storage. Okay. Right. So since this is right which is a great a great ⁓ callback because th the only reason this works is they these are running on the same

service in the same server, ⁓ which is the difference between ⁓ this, what we have now, and ⁓ clicking on analytics and this takes me to the other port. ⁓ That wouldn't work in that case, right? Because they're running on different servers. In this case, in so I have a proxy, they both are running on the same server. Does that make sense? Yeah. ⁓ So they're both pointing to local host three thousand there ⁓ for the user information? Am I if I'm getting that right?

The user information is coming from from an API in local for four four thousand. For four thousand is where the API is. okay. So both point to four thousand for the API. Okay. ⁓ What the what we're saying here about local storage is that here I'm setting when I log in, for example, and it would be let me just log out. Let me look at local storage.

Sneha Mehra (00:32:22)  
⁓

Sneha Mehra (00:32:26)  
Yeah, the system theme also should also work, right? I have this ⁓ system ⁓ that this manages my my theme if whether I use dark mode or not in this application. ⁓ And ⁓ the idea is that I can change this.

To light ⁓ and then go to analytics ⁓ and it preserves that because both are consuming local storage ⁓ from the same place. And that's the benefit of running it both in the same in the same server, in the same origin, I should say. ⁓ This will not be persisted if I open the app by itself. Like if I run open it here in three thousand and one, which is where the analytics run by itself, ⁓ I'm not gonna see light mode here because this doesn't have access to local storage of the other of the other one, right?

Sneha Mehra (00:33:20)  
But yeah, this is just to show that this is a way in which we can dip keep the applications separated without using like any frameworks. We're not using my module federation, we're not using anything here. ⁓ I can try to run just

Sneha Mehra (00:33:37)  
My dev server on ⁓ let's run the dev server ⁓ only on

Sneha Mehra (00:33:49)  
My API ⁓ and

Sneha Mehra (00:33:56)  
My application shell. Let's see what happens. ⁓ that didn't work, that ran everything.

Let's see what am I missing here?

Sneha Mehra (00:34:19)  
right, yeah, let's move ⁓ dashboard analytics from here. ⁓

Sneha Mehra (00:34:26)  
Now I'm only running my app shell and my API. ⁓ The other two are down. ⁓ And of course I'm getting an error here because the dashboard is down. But the rest of the app continues to work, right? ⁓ And if I go to analytics, ⁓ well, in this case, I'm not I'm not doing any error handling, but in this way it w we have the option at the route level where we're we're doing this proxy to say, if analytics failed.

show a fallback screen or something like that, right? ⁓ But we have options in in in this in this scenario.

Sneha Mehra (00:35:03)  
⁓ to show yes. ⁓ What are like alternatives to not using local storage? 'Cause I imagine polluting that with a bunch of state is probably not the best approach, right? For sharing state? Yeah. Between these different ⁓ so I guess it depends on the nature of the state. ⁓ If it's server state, you should probably share via the server, right? Like if it's if ⁓ I switch this and I change something in a database that says which account I'm looking at, then

the other apps running in different remotes will automatically get that because it the s the state is in the database, not in the client. ⁓ if it's client state, ⁓ when you move between ⁓ applications, ⁓ do you have an example of some state you wanna like keep ⁓ here ⁓ when you move to a different application?

Sneha Mehra (00:35:55)  
No. Because I I don't I when when sharing state in between vertical applications, I when I'm doing like a route change completely, I don't often find that I need to share client state. Right. ⁓ I guess it'd be like yeah, in a ho in a horizontal on the dashboard. Actually, never mind. ⁓ No, no, i it's it's a good question. We we're gonna talk about like ways to communicate between different microphones.

But I'm always thinking about I wanna communicate between these horizontal splits, right? I may wanna have, let's say, my table here is its own separate microphone, and that might want to communicate with the app shell. ⁓ we have options there to do that type of communication, right?

Sneha Mehra (00:36:39)  
I guess I'm just wondering with local state like the security concerns there as well. ⁓ yeah. Absolutely. Yes, of course. Like the w the method that we're using here is not secure at all. We're using we're saving s what is essentially an authentication token in local storage. ⁓ but you may want to use ⁓ if you use a cookie, a secure cookie, for example, you should still be able to access that cookie from another application as as long as it's running on the same remote.

Right. ⁓ sorry, the same origin. Does that make sense? So if when I log in to local host three three thousand, I set a cookie, I save a cookie with my with my user credentials. And then I go to analytics. If analytics, I just go to a different URL, that won't have access to my cookie. But if analytics is still running on on port three thousand, then it will have access to your cookies as well. Right. So that's kind of the difference between, these are two separate

applications in two different URLs and these are two different applications but they same this ⁓ they share the same origin. And I'm doing just proxy which like I said, we could you can use something like Next.js to say, ⁓ sorry, ⁓ not Next.js, I mean Nginx to say my analytics route actually points to this other service, but the rest of my routes point to my application shell. Right? ⁓

The way that I have it set up here is using I think I have ⁓ if we look at the tooling config and we look at my RSBL config.

Sneha Mehra (00:38:20)  
This is the way that I have it set up here, where I have ⁓ the proxy ⁓ of my ⁓ of my server.

That is ⁓ proxying request to analytics to a different server running somewhere else, which is I don't remember where it takes this value from, but ⁓ that should be essentially the value of localhole three thousand and five or whatever was the the the port where ⁓ analytics ran, right? ⁓ But you can do this with really anything. We can do this with Nginx, you can do this with ⁓

Let's say you have a Larabel application, right? You have a Laravel application where you have your controller or your routing in Laravel that has different routes, and those routes they just happen to point to different could be different index.html files, right? ⁓ And that will be essentially two independently deployable applications just running on the same server. Make sense? So this is sort of the do-it-yourself approach. It's a bit clunky, but it works. It works and this doesn't require you to install any

any libraries or ⁓ use module federation or anything like that.

Sneha Mehra (00:39:34)  
So we're gonna talk about module federation now, which is ⁓ the way that you handle w whether you have ⁓ more requirements than whatever this simple do-it-yourself approaches can give you.

So the thing about module federation is that it's kind of confusing at times because ⁓ it's both the name of like a some like an architecture or a pattern and is an implementation as well. So you may be referring to module federation as the concept, but maybe not using the plugin, right? ⁓ the it comes in it it it was born in Webpack 5, ⁓ as a feature of Webpack 5, and then they split it into its own thing. ⁓ So ⁓

You can Google module federation. Module dash federation IO, it's the the the official homepage. ⁓ And ⁓ it supports now since ⁓ after it was split from ⁓ from Webpack, it supports ⁓ plugins for different build build tools. ⁓ so if you go to build plugins here, you'll see that it supports RS build, Rspack, ⁓ Webpack, Vit, everything.

Thing with Vit, ⁓ this is what we were talking about before. It doesn't support the dev option for Vit, which is why we're not using it now. ⁓ And with RS Build, it comes with sort of built-in support for RS Build. ⁓ And ⁓ the if it comes with built-in support for conf for the configuration of module federation. ⁓ But it doesn't come with the runtime. We're going to install that manually. I'm going to show you how that looks like in the code.

Sneha Mehra (00:41:12)  
So the main benefit ⁓ I'll say or the reason why a lot of people reach for reach out for module federation is that it allows you to for two different deployed applications, not only to be able to run them independently and deploy them independently, but to share code at runtime. I can compose an application using components or code defined in another application at runtime. So I don't need to deploy anything. ⁓

module federation supports both client side and server side rendering. ⁓ And ⁓ you can do a lot of the things that module federation does with import maps, but ⁓ this has the benefit that it runs anywhere you can run ⁓ JavaScript, not just in the browser, and also it comes with some dev tools that makes the developer experience much better.

So now we're gonna work on this ⁓ repo that I have here to set up module federation. We're going to go back.

to the main branch.

This DIY branch was just to show you how ⁓ that might look like without using any frameworks. But now we're gonna go back to my

Sneha Mehra (00:42:35)  
Okay. The first thing we're gonna do to set up module federation is we're gonna install it. We're gonna install the runtime. Like I said, since we're using RS build, we already have ⁓ one other piece, which is the the plugin installed, but we need the runtimes, which is the tools that run in the browser. So to install it, you can run npm install add module federation slash runtime tools.

Sneha Mehra (00:43:09)  
And now we're gonna create ⁓ a split. We're gonna create a split for the analytics ⁓ application. What we're gonna do is we're gonna say that ⁓ our application shell, the app shell app, ⁓ is ⁓ what is known in module federation as the host. This will be the one that loads all all of my microfronts. And the analytics app will be what is called a remote. And that's the thing that we're gonna sort of embed into the host, right? So

Analytics will become our remote, AppShell will become our host. So we're gonna start by updating the config of analytics.

Sneha Mehra (00:43:53)  
Here's my RS build config, I'm just setting the port. I have some other configuration here, but it's not important for this example. ⁓ And here, since we're using RS build, we can simply do module federation.

Sneha Mehra (00:44:07)  
And we have to define a name for it, ⁓ which is gonna call it analytics.

Sneha Mehra (00:44:16)  
We have to define a file name, which is a file that's gonna be generated that contains information about how to load this remote in a host. By default, the the convention is to call this file remoteentry.js. So we're gonna see how that's generated in a minute. You can name this whatever you want, of course. And then we're gonna define what this exposes, what this remote exposes. ⁓

We may not want to expose everything out of this module. We maybe only expose ⁓ only one component, for example. ⁓ In this case, we're just gonna import one file, which is the index screen of the analytics module that lives in ⁓ screens ⁓ analytics index at TSX. This is just the only component that we're gonna be exposing. ⁓ So we're just gonna copy the path to that. ⁓

Sneha Mehra (00:45:17)  
And the key in this object, this is how the host or anyone consuming this microphone will refer to this file. As you can see, it doesn't have to match, ⁓ but you have to define both.

And then the last thing we're gonna do is we're gonna define which dependencies are ⁓ shared between the I think something

Sneha Mehra (00:45:44)  
I lost my second monitor. ⁓ Did you

Sneha Mehra (00:45:51)  
Yeah.

Sneha Mehra (00:45:58)  
I think it's let's log in, ⁓ think in.

Sneha Mehra (00:46:05)  
there go. It's back. It's back. All right.

start with that. Now we're gonna define in the share object which dependencies ⁓ I'm willing to share between the remote and the host, right? And we'll see what that is important I mean. So we're gonna define React ⁓ and ⁓ React React DOM to start with.

Sneha Mehra (00:46:33)  
In this subject, you can pass some configuration about the way you want to resolve these dependencies at runtime. ⁓ These dependencies, don't think of them as the externals of your package.json. In package.json, you can define externals, which means when you have defined externals, you say this package is not gonna bundle with React or React DOM. So whoever uses this package needs to provide React, React DOM, and I will use whatever version the

Shared is not exactly the same. This is about like negotiation of who's bringing React. If you have two, for example, the host and the remote, they both depend on React. ⁓ And you wanna negotiate which versions of of the library you use. And you wanna say, okay, I'm gonna use the same version that you have, but I'm not gonna share the same instance, for example. I'm gonna reuse the code that you have, but I'm not gonna share the same instance. So we have different ways of

⁓ of configuring this. So we're gonna we're gonna see that that will become clearer in a minute.

One thing that we're gonna define for both of these libraries is that I want to define these two as singletones, which means that I'm gonna reuse the same instance of React between the host and the remote, which is important for things like React, because in React, whenever you have, for example, context, if you have a context provider defined on the host, and then you're using that context in the remote, that would only work if those two are sharing the same instance of React.

not just the same the same code, right? So it's important for certain libraries to be defined as singleton. In the case of React is a requirement, otherwise things things won't won't work.

Sneha Mehra (00:48:22)  
So now we define the configuration of our remote. Let's go to the host, which is in the app shell RS build.

And here the configuration will will look kinda similar in the sense that we're gonna define a name.

Sneha Mehra (00:48:40)  
gonna op show.

Instead of exposing something, this this particular ⁓ application is gonna expose anything. But since we are using it as a host, we need to we need to define which remotes it has access to. So we're gonna define that we have one remote called analytics.

And this comes from the name that we define here.

Sneha Mehra (00:49:10)  
And that points to ⁓ analytics ⁓ at ⁓ in this case ⁓ and the analytics app will be be running on port three thousand and one. So let's go.

Sneha Mehra (00:49:25)  
And it points to that remote entry.js file. That will have the instructions of how to use this analytics remote. You have a small typo there on analytics.

Thank you.

And finally we are also going to define the shirt, ⁓ not the same. Which is gonna be the same that we have here.

Sneha Mehra (00:49:51)  
Again, this doesn't mean necessarily that this the these two applications are gonna share React and React DOM. It means that they are willing to sort of enter a negotiation about they're willing to share the dependencies with each other, right? Which may or may not be the case. Because, for example, this application I might deploy it somewhere and then I might be including this remote on a host that doesn't use React. Maybe it's a view application, right? And in that case it will come and it will say, you don't have React, so I have React with me, I'm gonna bring it with.

Right? That's that's w what we're saying here. ⁓ And here what we're saying is ⁓ I have React and whatever remotes I'm loading, I'm willing to share my version of React with them. ⁓ Right? Does that make sense? Yes? You're saying in module federation the host can be in a different framework than the remotes? ⁓ Yes. ⁓ Yes. Is there any like runtime issues?

There might be, yes, there might be depending on the frameworks. Typically it's su the support is pretty good, I will say. I haven't run when I tested with different applications written in different frameworks. ⁓ I have I didn't run into any any any runtime issues. ⁓ I'm guessing some frameworks may not like the fact that you have, let's say, a div ⁓ in a React application and that div is not controlled by React in in in a way, right?

So whenever try to do re-rendering, it might not like that. ⁓ But ⁓ I think a lot of frameworks now support that. Like they they support okay, this div, it's out of the tree of the reactory. I'm not gonna touch it. So they you can like you can import a ⁓ component building another framework.

⁓ which makes this also like a good solution for migrations. That's one of the one of the benefits of module federation or microfonents in general, is that if you have a legacy application built in backbone, you can migrate that to React incrementally, right? Because you can do this thing of ⁓ loading multiple frameworks at runtime on the same application.

Sneha Mehra (00:52:03)  
One more thing that we're gonna run. Let's let's run the app l we're probably gonna run into an error at this point. But let's run the dev script. ⁓ And we're only running the API and the app shell at this point, which is probably wrong. We're gonna fix that in a minute, but I'm gonna see if we run into

Sneha Mehra (00:52:38)  
Me update my dev script. My dev script is only running the API and the app shell. I'm also gonna run

The analytics up.

Sneha Mehra (00:52:53)  
How it works. ⁓

Sneha Mehra (00:53:01)  
Hm. ⁓ I was expecting an error at this point. We'll see if I run into that error ⁓ later on.

Sneha Mehra (00:53:11)  
Otherwise it might mean that we haven't set it up correctly, but we'll see we'll see how how that goes. ⁓

Sneha Mehra (00:53:26)  
Okay. ⁓ So we have the setup. We have the host and the remote setup. We're not actually using it. So now we're gonna update our ⁓ our router to actually ⁓ use the screen that we expose. If you remember when we defined our analytics exposes, we are exposing one screen and we actually wanna use that ⁓ that screen instead of importing the the package directly. ⁓

So let's go to the router.tsx, ⁓ which imports, since this was a like a single page application, all the apps are being imported here, all the screens. And now this analytics page, let's actually remove that. ⁓ We're not gonna import this directly. We can also uninstall the dependency now. ⁓

Sneha Mehra (00:54:20)  
The next thing we need to do is we need to import that component that we are exposing. ⁓ And with that, this will be an asynchronous import. We can import like like this because this will this needs to run at runtime, and all these imports that you have here will be resolved at build time. So ⁓ we need to have ⁓ a runtime import, right? So we're gonna define ⁓ a new constant called federated.

Analytics ⁓ page. ⁓

We're gonna be using React Lazy, oops, ⁓ which is how we can load ⁓ components asynchronously in React. ⁓

If you're using a different framework, ⁓ framework have ⁓ all frameworks have a way to to load components asynchronously as well. This is the way we do it in React. And here I'm just gonna start by importing ⁓ that analytics. ⁓

Sneha Mehra (00:55:19)  
index.screen.

Sneha Mehra (00:55:29)  
And finally I'm gonna define with that component that I imported, I'm gonna define my analytics page, ⁓ which is the one the import that I removed that I was consuming directly from the from the analytics package. I'm now gonna recreate it here and I will just ⁓ use React suspense. ⁓

Sneha Mehra (00:55:54)  
Just okay.

Sneha Mehra (00:56:00)  
To render that federated analytics page ⁓ that we just that we just imported. ⁓ So if you're not familiar with React, essentially when you are loading or you're importing a component asynchronously, you need to add a suspense boundary around it so that you can provide a fallback in case it doesn't load correctly or it's still loading and things like that. Because it will be evaluated at runtime. ⁓ Now this is failing my type check here because

Just come find it. ⁓ We are going to look how module federation two point zero gives us a way ⁓ of ⁓ generating types so we don't have to maintain types manually. ⁓ But for now, what we're gonna have do to fix this error is just gonna create a new file called ⁓ federation.tsd.ts. ⁓ and here I just gonna declare ⁓ module ⁓ just copy this. ⁓ This is just to get ⁓

rid of that type check error.

Sneha Mehra (00:57:03)  
We can be as ⁓ specific about what type of component it ⁓ it takes, what type of component this is, what props it takes, and so on. But since we're gonna see how we can do this automatically later, I'm just gonna not gonna spend some time doing that. ⁓ and now, analytics page, I believe that we're using that already in the analytics route, so there should be nothing else that we need to change. ⁓ So let's run NPM rundev.

Sneha Mehra (00:57:37)  
Someone in the chat asked if we had the key module federation in the config of App Shell.

⁓

I missed doing that? Maybe I missed doing that. ⁓ yes. ⁓ Great catch. That is exactly why we weren't getting the error before. ⁓ thank you. Yes. So that is in RS build sorry. RS build in the app shell. ⁓ All of this should be in module federation.

Sneha Mehra (00:58:14)  
Thank you, great catch. ⁓ Okay. ⁓ Is that what this was failing? ⁓ It was not. Let's see what's happening here. ⁓ It can't find analytics in the screen. ⁓

Might be ⁓

Sneha Mehra (00:58:37)  
Let me see one thing. I didn't put one thing on purpose and I'm wondering if that's why this is failing. ⁓

Sneha Mehra (00:58:53)  
This is not failing because of that. ⁓ Sorry, give me one second. Analytics ⁓ index that screen. ⁓

Sneha Mehra (00:59:08)  
Am I missing something here? ⁓ One second. ⁓

Sneha Mehra (00:59:17)  
We're due for a tech break, so we could do a little five minute break. You could debug it ⁓ and then we come back. ⁓ okay. Let me try this. Let me try this. Yes, let's do that. Let's do Yep. ⁓

