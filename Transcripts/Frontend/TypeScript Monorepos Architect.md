24  
Sneha Mehra (00:00:00)  
If I want to build an open source library composed of multiple packages like lib slash course, lib slash utils, lib slash react, all inside of one monorepo using Learna or NX, what's the best way to structure it for long-term maintainability? Should it just be ⁓ each feature in its own package, or is it better to keep everything in one package with an internal structure? Yeah, that's a great question. ⁓ I there is a trap of trying to split things apart.

too too much too early. And what what I would say is ⁓ look for like I would rather have a fat like core package ⁓ that I later understand I need to like factor some things out than start in a world where every single feature is its own package. ⁓ and then have to worry about like, all right, now people are using these things. It's sort of very difficult to claw that back, right? Once people have like

installed these things into their apps, you sort of I mean you can just stop publishing packages, but that loss of continuity of like a particular piece of functionality being available in a certain package, it's much it's much harder to ⁓ to correct the problem where you fragmented too much and need to consolidate than like, all right, you've ⁓ not fragmented enough. And so you factor something out into its own package, but then you can always re-export it from the original package and it's still

Still feels okay. ⁓ One exception, and that would be if whatever you're building ⁓ is intended to be ⁓ something that involves plugins, ⁓ and in that case, like maybe your features are plugins, that's where I would say, all right, well, invest some of your mental energy in defining a good plugin interface, and then you will want to have some world where.

You know, maybe your internal features, some of them are modeled as plugins, and then you can also let other people write libraries that also act as plugins. That would be the case where maybe you sort of end up with a lot of packages earlier, ⁓ rather than sort of waiting for a need to ⁓ introduce more packages to arise before before factoring things out. Have you ever had to override the NX Cloud?

Sneha Mehra (00:02:26)  
To store the task caches in your company, whatever companies cloud provider you ⁓ cloud storage use, whether it's like cloud storage or S3 or anything like that? I'm s sorry, like have I ever done that? Yeah. Have you ever had to like override, instead of using NX Cloud, like their cloud service, storing the cache in your S3 bucket or something like that? ⁓ not while using NX. So ⁓ I I kind of these days do two sort

Kinds of development. I do open source stuff, in which case I am happy to use NX, and I'm perp I'm fine with everyone being able to see the build artifacts. They would be able to see them in GitHub action logs anyway. And then on the other side, ⁓ monorepos that aren't just JavaScript, and that's where we're using Basil. And ⁓ we have a similar concept of being able to, you know.

Take advantage of distributed build artifacts, but it's not it's not something we'd use NX for. And that would be a like at Stripe.

Does that sense? And like in that case, absolutely we don't want we don't want to be tossing our not yet released build artifacts onto another person's ⁓ infrastructure. I mean like Amazon, sure, but you know, you a cloud provider is who we'd we'd be trusting there, not

Not that the NX team is not trustworthy, but I think like many enterprises would make a similar decision. Yeah, that's why that's why I looked into it because my company was like, ⁓ So I was like I was curious if you've had to use that in the corporate side. Only not on the corporate side, but ⁓ the ⁓

Sneha Mehra (00:04:11)  
Yeah. Sorry, I can't really provide a better path forward for you there.

This isn't necessarily related to monorepos, but I was curious if you've heard of Arctype before. No, I have not. No, it's similar to Zod. I'm basically ⁓ trying to get into some data validation ⁓ stuff for TypeScript since we don't really do any of that right now. So I know Zod is the one you covered on ⁓ last ⁓ last one.

But also looks into archetype as well. So ⁓ I mean sod is used ⁓ extensively. And I think like I haven't seen archetype, but ⁓ when evaluating libraries like that, I would look at, you know, how likely is it that you can hire someone and they already know how it works, or how likely is it that when you run into a problem, you can go look on Stack Overflow, or let's be real.

Ask an LLM for help. And like, how many examples do they have to pull from? ⁓ and if you've like, don't don't forget this, as a TypeScript or JavaScript developer, there is an enormous amount of TypeScript and JavaScript code that LLMs have ingested. But if you try to get an LLM to help you with a more esoteric programming language like Elixir or or something like that, it's a lot less ⁓ on target ⁓ because it just has a lot fewer examples to pull from.

And similarly with libraries, like there is if you're using AI coding help, there is an advantage to going with an option when you're making choices that where there's like just a ton of usage out there and a ton of stuff in GitHub and GitLab open source. That all helps build ⁓ I'd say an understanding, but really like a database of autocomplete results that

Sneha Mehra (00:06:15)  
That will ⁓ you know, be more likely what you want, you know, what you want to implement. Like a more helpful answer if that makes sense. Arct type seems kinda cool because it literally uses ⁓ TypeScript syntax at runtime. So you can take TypeScript syntax and it'll take that to runtime and then do data validation. So you don't have to learn any new APIs, which might make it an easier cell. But also Zod is more proven, I guess.

Well, and in that case, maybe what I just said is less relevant because if if the way you express these things is in TypeScript types, then there is plenty of information out there on that. And it makes sense. Like there are lots of libraries where you can extract a JSON schema or a protobuf or whatever you want from a TypeScript type. And it stands to reason you could ⁓ use that to compile a validation of some sort. I'm gonna piggyback off of Pepe's question here in the chat.

Do you have any best practices for building NX generators? Or do you have any favorite ones that you or your company have built and use? I I d bluntly I use the off-the-shelf stuff. I do sorry, I do have a best practice. It's not a best practice. I can offer my practice. ⁓ for for something like our dev script, right? Like, and I would just know about this. ⁓ more as a, you know, don't overlook this as a possibility thing. Not

That it's the answer to everything. But ⁓ a lot of people ⁓ don't seem obvious when when you think about it, but like you can create a package that ⁓ NXCs as a as a package. And it has like a project.json file. Like you could just create a folder with a project.json file in it. And it's kind of a fake package. Like it has nothing to do with software that's being imported.

Into other packages in your monorepo. But that lets you, it has the effect of letting you almost like create commands that are of the shape that you want to create. Like if I wanted to create something that was like ⁓ nx devserve, well, all I need to do is create really that's following the pattern of nx ui test. Oops, sorry, I got it backwards.

Sneha Mehra (00:08:39)  
NX test UI, yes, terminal too small, gotcha. Right? So really what you're saying is I want to create a target that is called dev and a project that is called serve. And so this lets you, if you if you lean into that, like these are just two words, and so you can create a combination of projects and targets that let you make a very ⁓ semantically relevant command.

If you're just like you're willing to break this ⁓ sort of implicit pattern, but it doesn't have to stay that way of like, ⁓ well, the project JSON goes in my models folder, and there are absolutely no other places I can put this project JSON. Like you can make NX aware of projects that have nothing to do with your TypeScript ⁓ monorepos here. It could just be like another folder in your

in your repo, like another top level folder that's not even packages. Does that make sense? Like this this could become like noun verb or whatever you want it to be, because these are really just task definitions, like targets, ⁓ and then

This last word is a project. And a project is just a project JSON file. It does not have to be mapped one-to-one with a package JSON.

Sneha Mehra (00:10:04)  
So like ⁓ usually ⁓ nx commands are sometimes not incredibly ergonomic. And so I will I will reach for that when I have something that's really common that needs to be run, and I want it to be like stupid simple. You know, like ⁓ nx ⁓ format manifests ⁓ or nx ⁓ like lint CSS. And you could do that if you just set it up.

this way. And you're not you're not abusing the tool, you're just cleverly naming the targets and the projects.

While benefiting from from like the the the target dependencies and all of that, right? Like of course you could do P NPM whatever you want. Like you could use NPM scripts, but ⁓ NX can do that too. It just requires some more JSON. So this this one's tangentially related, but you seem to be a big advocate for P NPM. If I were to go back to my company, how do I sell P NPM to them to replace MPM? ⁓

So the most obvious benefit is really like a developer productivity thing, where

Like every install is faster, which makes CI faster. Local build startups are faster. ⁓ it means less ⁓ less compute happening ⁓ on your your dev machine. And so that that means, you know

Sneha Mehra (00:11:41)  
I mean, I'm sure I'm sure Stripe is not the only company where this happens where like developers get issued at the time. They feel like like these really powerful laptops. ⁓ And then you get to a point a couple years later where you're like, I only have I only have sixteen gigs of RAM on this thing. Like, God, that's that's I've got two dozen Chrome tabs open and I've got an NPM install going in the background and it it sucks. So yeah, I would say like mostly the developer productivity angle is

is the the main thing there. ⁓ alternatively, like if the company is paying for storage of build artifacts, just the size on disk with Pnpm is ⁓ so much smaller. I mean, sorry, for a sizable project, it's so much smaller because of just the way the the linking works and the flat the flat structure ⁓ of the

Of that like cached set of dependencies. Gotcha. Interesting. Thank you. How much a lift do you think it is? ⁓ Like switching at scale? ⁓ w what what dimension of scale? I don't know, like a hundred packages or something like you know. A hundred I mean, that th that would be a dimension of scale where I would say doesn't matter. I mean like that should be fine.

You'd start out with Pnpm, you can point it at your package lock.json or your yarn lock file. It'll generate like the right PNPM lock file that doesn't drift any of your ⁓ dependencies. That's fine. Here's the other dimensions of scale. What are developers used to? Like, ⁓ how you kind of have to train them up a little bit. You're gonna get people who are a little curmudgeonly about, like, there was this one feature of this thing that I really liked. ⁓

Where there is a little bit of a trap is if you're using ⁓ some custom tooling that like

Sneha Mehra (00:13:46)  
Doesn't that ⁓ relies on a different layout of node modules, ⁓ that's the thing I would look at first. Like if you have wacky scripts that kind of depend on certain things being in the node modules at the workspace level and certain things at the works at the at the package level, those are the things that are likely to break. But the good the good news is that's sort of a fail-fast like task there. Like you try it out.

And it'll be obvious if it breaks. Like you just need to exercise all those tools and ⁓ see see if something produces ⁓ a a wacky result of some sort. But really, like the scale that I would say is ⁓ what to worry about. It's like it's not the hundred dependencies, it's the hundred engineers. ⁓ And kind of helping them understand all right, here's how to think about these things, workspace colon is what you do. Like that's it's more about.

Like any tool requires this kind of thing. Sort of the cultural rollout.

Sneha Mehra (00:14:50)  
It's I mean, these things are ⁓ it's they're not their implementation is not simple. But they all read a package.json, they all have a lock file, they have differing articulations of how workspace works, but you know, that's a relatively small surface f for them to be entangled with your code base. And so you know, it depends. Like i but i

Granted, you could have a bunch of custom stuff that's doing a bunch of npm link everywhere, and that's not going to be friendly with P NPM. So maybe maybe the scale dimension that makes this really hurt is how much custom software has been built on top of like the conventional use of these commands, like that more tightly integrates you with one package manager. So I hope you enjoyed learning about ⁓ TypeScript monorepos and some of the great tools you can use to make your builds fast.

To make developing on a big project feel small in a good way. And how, ⁓ even in a world where you have many packages inside the same Git repo, you have ⁓ great ways to trim down your dependencies, ⁓ to ⁓ enforce consistency standards across your code base, and generally to make sure that you you get to reap a lot of the benefits that the concept of a mod repo promises.

—----------

7

Sneha Mehra (00:00:00)  
Now let's get into the TS configs. So in this next step, we're going to set up ⁓ the right or a working TS config setup so that Visual Studio Code still understands when to give us the red squiggles or whatever editor you're using. Like the language server needs to be able to engage with the project. You need to make sure that you're building the right thing. And you need to make sure you can type check the right thing. And what I mean by the right thing here is like when you're type checking.

You should type check your tests ⁓ and your source code. But when you build, obviously you don't want all of your test files to be part of like a production build. There's no point. So we're gonna have to create a couple TS configs in order for this to work. so the first thing that we're gonna do, and we'll we'll come back to this models folder when we're wiring everything back up. ⁓ we're gonna go back to our UI package, and we have

A bunch of different TS configs here. This was literally what ⁓ starting this Vite project auto-generated for me. ⁓ And we've got ⁓ a tsconfig.json here. ⁓ And this this represented like this is what a language server is typically gonna look for. Yes, you can configure it, you can provide some settings and point it to a different thing. But if you just like open up a project, there's a ts config in the root. That's that's what your language server is gonna look at. And before we

Started the course, this is what was working here. So we're gonna take this ⁓ and we're gonna drag it into the root of our project. We're gonna move it to the root of our project. And we're gonna have to change a couple things. ⁓ Well, first off, source, everything in source and tests, this is not like there are no top level source and tests folders now.

But we can do this. We can say, all right, like any any subpackage, subfolder of packages, ⁓ look for source and tests, ⁓ and we're gonna be tracking those files. And then ⁓ I'm gonna go, we we need a TS config for ⁓ for the UI here. So I'm gonna create a new file here. And in this case, I this is going to be regarded as a not a rename. We're changing these files substantially. So creating a new file here is fine.

Sneha Mehra (00:02:23)  
And ⁓ what we'll say is this file extends

Sneha Mehra (00:02:36)  
Extends TSConfig.json. And if we click there, we get to our root level tsconfig. That's how we know we have the path right. And then we'll say includes. ⁓ And here we'll have ⁓ sorry, I'm gonna close the package json's here just so we don't get confused. So the left we've got the root one, and this is for our UI. So here I'm gonna kind of ⁓ grab all this.

Sneha Mehra (00:03:08)  
Paste in here. ⁓ And now my job is like I'm going to sort of prune this down a little bit. So we'll we'll sort of undo what we just did for packages, ⁓ right? Because ⁓ relative to this file here, there is a source ⁓ and a test subfolder. And then we've got all this stuff that's like tailwind specific, post CSS specific, vite, svelte, yes, lint. We'll deal with linting in a later chapter. But for now, like

ESLint config lives ⁓ here. So it's in the UI folder. So great. So this is what we want. It's kind of this was the starting point. Like this these were the original contents of includes in this file. And over here we can get rid of this stuff. So just a summary of where we ended up. ⁓ this is looking for all TS, JS, and Svelte files ⁓ in the source and tests.

Folders ⁓ of any of our packages. And then over here, we have a complete set of within this package. What does this look like? Now, one benefit we have here is like there are no strictness settings in this file anymore. And what we want to start doing is working towards a place where across our whole monorepo, we can just kind of like point to this file and say, look, no implicit any. That's a rule I want to enforce everywhere.

In all packages. You don't want to be going around and like, you know, poking at individual TS configs eventually, right? Like in reality, sometimes you're tightening things up and you want to do that on a package-by-package basis, but there is some benefit to having like a baseline layer of config, particularly around strictness. This is almost like linting settings. If TypeScript is a fan fancy linter, which it kind of is.

Yes. So aren't we including with the includes and the base TS config pointing down to the package ⁓ directory? Aren't we including stuff twice then? ⁓ Yeah, let's get rid of this and see if it works. Bluntly, I I forget whether we're replacing the includes array or whether it is ⁓ it is ⁓ appended to. Like I let's let's delete that and let's see where we end up. We can always add it back later. But

Sneha Mehra (00:05:31)  
We are absolutely specifying it twice here. So we're gonna see real quick like, is this sufficient? And all that all that you'd need to do in each monorepo package is whatever's extra, like whatever's not in the source and test folder. And in this case, I want to type check all these config files in the root of the package. Good good question. Okay, ⁓ now we need to turn our attention back to models. So we'll close our UI folder here.

We're going to need two files in here: tsconfig.json, tsconfig, dot build.json. ⁓ And the rule of thumb here is the t sconfig.json, this is the file that governs like where you want your authoring feedback. tsconfig build.json is about compiling. Now we want to make sure we're still leveraging a common set of strictness settings and

module type and our target our target language level, but certainly the includes are going to be different here. One of these should point to your tests and the other should only point to your source folder.

So ⁓ in our new tsconfig.json, it's going to again extend from the base.

Sneha Mehra (00:06:52)  
the root level tsconfig.json ⁓ and ⁓ let's see if this is sufficient.

Sneha Mehra (00:07:03)  
Yep, that looks good.

And ⁓ TS config build.

This should only include source.

In this case, don't worry about extensions. ⁓ It's all TypeScript files.

We don't want the tests to be compiled.

Sneha Mehra (00:07:29)  
Alright, let's try our build command again. ⁓ wait, sorry, before I do that, we need a couple things. Compiler options. Alright, no emit false. This is a double negative. What we're saying here is yes, emit true. ⁓ and we're going to say the out directory is dist. Without this, ⁓ you're gonna end up with a js file and a

DTS file right next to your input TypeScript files. And personally, it's kind of messy to me. ⁓ Root dir is source. ⁓ Now what this means is ⁓ in our dist folder, we want to see the compiled results of our source folder at the top level ⁓ of the dist folder. We don't want to see dist slash source slash.

index.dts and index.js. We just want to see dist slash index.js. And so this is sort of like the the the root of what is to be compiled. ⁓ And this is a library. So let's make sure we build declarations. We have to opt into that. ⁓ And four options. That looks right to me. ⁓ one more thing. I want to do this. So why do I want to do this? Why not point to the base package JSON?

Well, this still means that if I wanted to have some strictness settings here, if I wanted to say compiler options, ⁓ no fall-through cases in switch, ⁓ this lets me have a convention where like it's always the regular tsconfig.json where per monorepo package strictness settings are are set. And so you can think of it almost like you've got the base level tsconfig.

That threads up to each ⁓ monorepo's ⁓ sort of like authoring TS config. And then the compile TS configs, like it's gonna look very much like this in every single library that you have. Because all this is doing is saying, okay, forget the test folder. Also, we're actually compiling stuff, and here's where to put the output. But like these files are just commoditized. We're gonna see when we create other monorepo packages.

Sneha Mehra (00:09:55)  
We're copying and pasting these things. Like it's just, it's a very uninteresting file. And if you were doing a large scale monorepo, this would be the kind of thing that you would ⁓ you'd template, like you'd use TS Poet or something like that to just crank one of these out. ⁓ And ⁓ you know, it's it's ⁓ not something that a human should really be messing with.

Alright, let's try to build.

Ooh sorry. ⁓ Out or not out. ⁓

Hey, there's something. ⁓

What happened? The thing that I said that makes things messy happened. ⁓ did I get confused here?

Sneha Mehra (00:10:45)  
Outdoor is disk. There is stuff in the disk folder. Let me blow this away and see what happens.

Sneha Mehra (00:10:59)  
And I'll blow this way.

Sneha Mehra (00:11:06)  
Let me just check my build.

Sneha Mehra (00:11:11)  
Script here, make sure that's right. Build tsconfig.build dot JSON. Yep.

And I'm in the models folder.

Sneha Mehra (00:11:28)  
You know what? It was probably my my the last time I ran the build command and ⁓ maybe I just wasn't paying attention to my sidebar. And but like this is the expected output. Maybe maybe those ⁓ files I just deleted were from when I attempted the build and there there wasn't any ⁓

We hadn't put these compiler options here yet. ⁓ Anyway, this is what we're looking for. So we've got our index.dts, right? We've got our index.js, which is like the same things here. I mean really this is just type information. And so

interesting. Were these classes?

Sneha Mehra (00:12:13)  
⁓ we have enums. That's why we're getting some compiled output. Enums are not purely type information. They are values. Right? And so if we look in. Yep, there you go. There's this is our our string-based enum being created. So great. So we've got some meaningful JavaScript. We've got some meaningful TypeScript. ⁓ let's try dev mode.

Sneha Mehra (00:12:38)  
And we started compiling in watch mode. And so what that means is if we were to go here and say, I'm gonna add a new property, save, and you can see if you're watching closely.

Sneha Mehra (00:12:55)  
Every time I save a file. Things update if I were to have any build errors.

We'd see those pretty clearly down here, right? Get a nice link we can click to, go to the line. So great. So we've got a nice script for rapid development. We've got a script for building. One more thing we have to give attention to is like exactly how how is this package exporting what it has in its disk folder? So go into your package JSON ⁓ and we're gonna add a couple new fields. Types. ⁓ And we'll say dist index.d.ts ⁓ and module.

This is dist.index.js. So think of these as the type ⁓ and runnable code entry points for this library. So when another package imports this, we're we're saying like, where do you go? Where is the build artifact?

—------------------------  
16

Sneha Mehra (00:00:00)  
Next, we're going to apply a TypeScript compiler feature called TypeScript Project References, ⁓ which is related to composite projects. And this will improve the performance of our TypeScript builds. Now, in in this workshop project, the builds are fast. There's just not much code here. But this will make a huge difference in a monorepo of significant size. And it's really easy to set up. So the first thing we need to do is think about ⁓

Like what we need to think a little bit about a dependency graph and we need to understand like what's what's depending on what. So

We will need to in the models folder in the tsconfig.build.json, because this is really a build feature. We're gonna have to go into the compiler options and say composite true. What we're doing here is we're enabling this ⁓ enabling this project to be compiled in a way where you can have a piece of build information for this package, or really for for ⁓

Like TypeScript would regard this as a project, like a TS config refers to some collection of source code. And anyway, this project is gonna have a build file that can be stitched together with other pieces of build information to allow ⁓ rebuilds to happen more incrementally. And just remember what our baseline is. Yes, we are sharing some code, right? Like, or we're benefiting from ⁓ some consolidated compile stuff. Like there is one dist folder.

For the models package, and anyone using the models package points to that dist folder. But the downside here is anytime we touch anything ⁓ in a package, that whole package is being rebuilt. And so this ⁓ composite project and project references and these TS build info files, they allow ⁓ even more incremental compiling ⁓ at the module level rather than at the package level. And that's again, like

Sneha Mehra (00:02:06)  
Your project gets big enough, you will notice a big difference here. In fact, this is if you're not using this and you have a s sizable monorepo, this is absolutely the first thing you should invest in. So composite true. In fact, we're gonna add this to all the package.jsons of everything in our monorepo. Sorry, the tsconfig build JSONs, not the package.json. So there's the server, and we'll add it to the UI. ⁓ And just because of how Vite works.

This TS config is what's used for build. And so we're going to ⁓ we'll add it here.

Sneha Mehra (00:02:44)  
Great. Now we need to ⁓ establish references. And so these have to do with like edges on your dependency graph. And I want you to think about them as going hand in hand with something like this. If you have a workspace dependency, you should also be establishing a project reference to go along with that. Now, your build won't fail if you're missing this project reference, but this is what you need in order to get that kind of speed up factor. Here's what it looks like.

You're going to have ⁓ it as a top-level property ⁓ of your the TS configs you're using to build. And again, we're in the UI one, so there's no build here, but that's what it's for. And we're gonna create a references array with one item in it, and it's the path to ⁓

Sneha Mehra (00:03:41)  
Build.json. This file doesn't exist yet, but it will in a moment.

Sneha Mehra (00:03:48)  
This is going to be sort of our builds, our incremental build artifact, if you will. I'm going to copy this because I'll need something very similar in the server package JSON or TS config build JSON. ⁓ And it's a top-level property. Too many commas. ⁓ And ⁓ turns out, same relative path.

Sneha Mehra (00:04:14)  
Models ts config build.json.

Sneha Mehra (00:04:23)  
And now I just ran ⁓ PnPM builds and it built everything. Here's our models, our server, and our UI. ⁓ And we get these files now. ⁓ And you can see just there's a lot of information here about which declaration files are used and what your dependencies are and what version of everything is being used. ⁓ This effectively is ⁓ what it means to have these proxy references. ⁓

Unless we're working with something of substantial size, we're not going to see a speed up factor here. But ⁓ trust me, it's there. ⁓ And you will know, you will know it's working for you when you apply this to a large project. And you can see, especially that incremental rebuild time is faster. ⁓ And the incremental build time. ⁓ sorry, the way incremental compilation affects the performance of your language server, that's also going to be important, right? That's affected by this.

And so that's where you know you're gonna hover over things ⁓ or you know, in those cases where your language server is lagging behind a little bit and you're seeing the red squiggles and it's still figuring out that you installed this thing, like it it really tightens that up significantly. So the idea here is when we edit the server, it's not reevaluating anything in ⁓ models, right? It's just using cache. Yeah, ⁓ sorry, let me.

I'm gonna do a little experiment here. We're not gonna commit this, because I typically don't commit these build info files. But I wanna I wanna see if we can spot what changes. That might be a good way to look at this. models. Perfect.

Sneha Mehra (00:06:16)  
Alright, so I just made a fairly trivial change and let's look at the build ⁓ and

Alright, so first off, you can see like two things happened. One is you can see server was affected. There's something happened here, and something happened in models. ⁓

There's a sense that servers like exposed to this change. gosh, we're not gonna make any sense out of this. All right, you're gonna have to take my word for it. ⁓ what's the difference in terms of what's happening here is ⁓ in this build info file is more of a sense that like this one module changed. And we're preserving all of the compiled results of whatever can be preserved within this package, as opposed to.

Rebuilding this entire package from scratch every time you hit the save button. And just think about what's happening in in when we run our dev script, where like anytime we change something and we hit save, what's happening is that whole that whole package is being rebuilt from from scratch. Like there's no state that's being preserved between builds. This represents state being preserved between builds and an opportunity to an opportunity to reuse.

some of the state that can be safely re reused given the scope of the change that you made when you hit save. And so it's it's really like slimming down to you know closer to the minimum amount of work that's necessary in order to create an updated build output. If you didn't ⁓ add that reference, what would the behavior be in the original TS config to the models? It would rebuild the whole thing, right? It would rebuild the whole project. Like you'd you'd end up with something new. So

Sneha Mehra (00:08:04)  
The contents of the disk folder are going to be the same in either case. But the difference is what what is the work required in order to get there? Before I added project references, it's similar to compiling it for the first time. Like it's as if you have nothing in your disk folder. It's just building everything from scratch. You can think of this as almost having like an intermediate result where certain files that I didn't touch, and it's not really by file, because obviously like this this is changing too.

⁓ but like certain parts of the package that were left unperturbed, we can reuse the build output from the last build as sort of an advanced starting point for creating that that new build. So the increment of work that's required to get the same build output is significantly smaller.

Does that make sense? Yeah. It's more like an edit to the build as opposed to throwing away and recreating that. I think my confusion came from ⁓ aren't you already importing a built ⁓ disk of models ⁓ like into the server? Like they're already separated and that's already built, so I'm just struggling to understand where the savings come from. That's a great that's a great point. Remember, when we're saying we're importing ⁓

When we're in ⁓

Load data, and we're saying I've got the seed packet collection model from here. I want you to think of this more as like instructions of where to look when compiling. There's nothing here in terms of preserving work. And so, like ultimately, this is just directions for finding the thing that we're interested in. Like, where where can I find this thing? If you remember when we were messing with ⁓ with this folder here, like

Sneha Mehra (00:09:56)  
Remember when we were messing with these? And we s when I took them away, we c the the it couldn't find the module. Like the the dependency link was sort of broken. It's because all all this is all this is, all the import statements are, it's really just like instructions for finding something that exists in a node modules folder somewhere. Now

So that's one thing. Another thing is, how does what is the work required to produce an updated disk folder based on changes to source code? And what we just did here by adding TS project references, it's the difference between deleting that dist folder, starting completely fresh with no knowledge of previous builds, and doing the exact same amount of work to recompile ⁓ models as

we did to build it the very first time. So that that was our starting point. And now we're more at the point where, well, we have a lot of like bits of combined compiled data for other things that might be in that package. And some portion of that we can reuse. And so those that that r that represents work that is like already done and that ⁓ reduces the the amount of new work that we have to do. And the amount of new work we have to do is

Much more closely related to sort of the scope of the change, which is like it's still always at the file level. So you're not, you're not like recompiling a function within the file. You're compiling, you know, you're creating a new compiled output for that module. But it's it's sort of like incremental compile at the module level instead of at the package level. And the bigger your packages are, and the more of them there are in your monorepo, the more this is going to make a difference for you.

one last thing. Git ignore these. As you can see, they are just ⁓ junk. They're those are files for programs to understand.

Sneha Mehra (00:12:03)  
Don't commit these at you of bedtime. It they'll just always change all the time.

—---------------

12

Sneha Mehra (00:00:00)  
In this next section of the course, we're gonna work on getting linting and code formatting working throughout our repo. It's something that we had in the beginning state of our project, and we have since lost it, in part because ⁓ the prettier and the linting stuff is only happening within the context of our UI package at this point, right? We kind of left left the ⁓ left those concerns there. So ⁓

We also want to take on some other ⁓ developer experience niceties. So one of these is if we look at ⁓ one place where we're missing a developer experience nicety is if we pull up a component that that depends on something from ⁓ like seeds slash models, and we command click on this.

We'll end up in a declaration file. Like we're in the dist folder right now of our dependency. ⁓ And if we really want to deliver on the promise of, you know, the way you navigate code, the way your tools work, the way your scripts work, it's the same. It's the same at the end of the day, once we've converted this to a monorepo. Like we had better be able to jump to ⁓ these source files and be able to sort of seamlessly ⁓ step through and navigate. In particular, this is a particular

This is likely to be a a potential hazard because this, ⁓ especially with the building out to modern modules, ⁓ it looks a lot like a TypeScript file, like a TS file. ⁓ and it'd be very easy to kind of make the mistake that you think you're editing the source, but you're editing the build output and you drive yourself crazy, wondering like, why does my change keep getting overwritten? Why hasn't it taken effect? So we're gonna address that too.

And then ⁓ finally, because we're we're sort of ⁓ going to be working with some some config files in this area, we're going to apply this concept of TypeScript project references, which will have each package generating a build information file. And that that will improve the performance of our build, it'll improve the performance of our language server because TypeScript will have a much richer source of information in terms of ⁓ what has been changed.

Sneha Mehra (00:02:23)  
since the last build, what needs to be incrementally recompiled. it it won't be on a package by package basis. It'll more be on a module by module basis. And we'll take a look at the inside of these files and it should should make sense why this works so well.

Sneha Mehra (00:02:41)  
So ⁓ first off, let's let's focus on prettier. ⁓ if we were to run this command at the project root, ⁓ we should get this error.

Prettier plugin Svelte is not found. ⁓ And a reference to a file here, which ⁓ is a generated thing, so it does that doesn't help us too much in the air. ⁓ But if we look at our Prettier RC, we can see, hey, we've got this Pretier Plugin Svelte. ⁓ So this is a cue that we have ⁓ like we haven't yet pulled something that's kind of like a

project-wide tool up to the top of our monorepo. So when you think about linting, when you think about code formatting, things like that, you you generally want those to be applied across your whole repo. Now it would have been fine if we if we left PurdyRarC inside UI and that's the only place we wanted code formatting to happen, that would have worked perfectly well too. We would have only been able to run this task ⁓ when we're within that package. But we want

the benefits of a modern repo and having the same code style applied broadly throughout all of the packages that we have we're authoring together. So we're gonna go into our UI package JSON.

Here it is. ⁓ And ⁓ there is ⁓ a Prettier plugin Svelte. In fact, we can grab Prettier and take it out of here as well. ⁓ it's gonna be running, Preter will be running at the project root. It's not a dev dependency of any one project anymore once we do this. ⁓ And we can add those here.

Sneha Mehra (00:04:29)  
There we go. If I was using some opinionated ⁓ formatting here, maybe we'd end up

Making these more alphabetical. ⁓ okay, now let's try this again. ⁓ sorry. We touched a package.json, therefore, ⁓ PNPMI.

Sneha Mehra (00:04:53)  
And there we go. We can see ⁓ if you're if you're looking closely, it's probably more clear to me on my screen than it is if you're watching at home. ⁓ some of these files, you can see unchanged is missing next to some of them. ⁓ formatting was applied. So now we're formatting in the models package, in the server package, in the UI. We have sort of hoisted this ⁓ this tool up so it's now working at the Monorepo level. What we can do is ⁓

Create a nice NPM script for this. So instead of saying the way we format across the monorepo is by invoking the format script on each monorepo package, we're gonna replace this with something that operates at the monorepo level. And it's gonna look like prettier.

Right.

⁓ packages, any any package, right? And then ⁓ either the file source or tests.

And then any subfolder TS. And ⁓ if you copy from the workspace notes, so this should be a basic version of this that'll get all of your source and tests. ⁓ If you go to the workspace notes, I've got a string here that you probably don't want to figure out ⁓ yourself that also will format the config files. So we'll grab that and we'll paste it here as well, getting the comma in the right place.

Sneha Mehra (00:06:27)  
Pnpm format. ⁓ And it does basically the same thing. Looks like ⁓ some of those new files that were picked up this time ⁓ didn't end up getting affected. So here you can see this Vite N file was in the last invocation that was picked up, but now we've got things that are not in the source folder, ⁓ like these down here. So great. Purtier formatting is working across our monore.

—------------------------

21  
Sneha Mehra (00:00:00)  
I'm going to ⁓ go into models ⁓ and I just wanna add a trivial change somewhere, like

Like here. What is this? Nope, that's in a noom. That's actually gonna mess things up. ⁓ we've got human toxicity, pet toxicity, livestock toxicity. Great. We're gonna have ⁓

Sneha Mehra (00:00:32)  
Alien toxicity with affected parts. ⁓ We're worried about how our vegetables will affect aliens. Alright, so I've saved this. ⁓ And you can see this yellow X. I've got

I've got something s like a code alteration that's been made. So what I can do is say PNPM, learn a run test since ⁓ equals

Course progress.

And what it's done is it can identify that ⁓ I have made gosh.

Sneha Mehra (00:01:17)  
There go. That's more satisfying to look at. ⁓ well, I've touched something in the models folder and I've made a change. ⁓ And ⁓ it it kind of like looks at the delta between my current working state ⁓ and any git ref here. So this could be a SHA, it could be ⁓ it I could have pointed to like origin, course progress, whatever you want to do. And so like if this this allows you to run tasks on a subset of

Not only what you just touched, but what is downstream in the dependency graph from what you just touched. So it's a lot more sophisticated than if you've ever used something like lint staged before, where it that that's simply just like looking at git diff, which files did you touch? Let me run things based on those files. ⁓ here we're getting something very important back that we had at the beginning of the class when this was a ⁓ when this was a

Monolith in a single repo. And that was we could make a low-level change to our models, ⁓ and then the entire test suite ran on the server, on the UI, the tests that run on the models. And it's very easy when you start separating things out into a monorepo. If you don't have the right tooling in place, you can fall into this trap where it's almost like the unit test trap, where you know, gosh, it seemed fine. We we made this code change, and like the tests in that package.

Past or the linting ⁓ on that package past or this type this small part of the project type checked, this lets you get back to that point where you're saying, all right, I have the benefits of a monorepo, meaning I'm not like boiling the ocean and saying, run ⁓ all the tests over again for the everything that exists in this git git repo. But we are getting the benefit of saying, ⁓ I touched something, ⁓ use your knowledge of like where was the thing I touched.

what are the downstream dependencies that it could affect and now we can end up ⁓ running a subset of the tests there. So let's let's see if

Sneha Mehra (00:03:29)  
Let's let's see if we can show an example of this actually working. What I should do now is revert this change.

Sneha Mehra (00:03:38)  
Let me touch something just in server. And the hope is this just runs.

There we go. Trivial change.

Sneha Mehra (00:03:53)  
Oop. Just the server wait, am I in the server folder? Could have fooled me. There you go. That's all that's running.

Sneha Mehra (00:04:04)  
The only thing that changed since my head commit was files and server. So we don't need the models to run again. We don't need that test suite to pass. That's upstream. Presumably those tests pass in master. ⁓ Like if I check things out and I haven't touched it, it's validated already. So this is a really cool way, ⁓ a cool benefit, a way to reap this benefit of being able to operate on, test on, type check on.

Lint on small portions of your project. And this is where we're starting to see some of the the lightweight promise of working in a monorepo come to life. But you need a a a build tool that's aware not just of file structures, ⁓ like file locations ⁓ in a directory, but the dependency graph of how all of these things are related to each other.

When we're doing learn a run test, is that our old like test ⁓ script we had that it's running? Yep. This is the this is the test script. And so how is it able to break up down to individual tests if it's just running a script? To be clear, and I'm gonna I'm gonna prove this. It is not in our root package.json, like

Sneha Mehra (00:05:24)  
I'm renaming that. ⁓ This should still run. This is going into each individual package ⁓ and running PNPM test. But it's making a determination of which packages it does that in based on what was changed and its knowledge of the dependency graph. So this is very much like the for each ⁓ of testing. And so what I would do here is I'd say, great, we've got like the test thing in CI. Yeah, we're going to run that.

But we can also have like test changed ⁓ and this is where we would say, look, PNPM is not quite specific or sophisticated enough to do this, and so I'm gonna grab this and I'm gonna put it up here.

Sneha Mehra (00:06:09)  
And we can just have this be like, this is what you would periodically be running over and over again. Like I made some changes, ⁓ lint it for me. Lint only the stuff that's changed. Build only the packages that that need to be rebuilt as a result of the code I have not committed yet.

And it's super powerful to be able to point to any git ref. Like, ⁓ you want to look at ⁓ like is this an incremental change that is being added on top of an existing PR? And you know that build already passed. And so you can just look at the last three commits that you've made. So ⁓ let's just do it dash dash since origin slash PR branch or whatever that is. So really the benefit with this is gonna kind of compound as you have more and more.

packages, I guess, in a monarchy, but right. If you just had one package, this wouldn't really do anything. ⁓ is that correct? Yeah, it's it's almost like the more this benefits you, the more you're able to break a sizable project up that already took a long time to build, to lint, to test. And the more you you're able to break that up, this approach is what lets you pay the cost, the computational cost.

Of like performing these build, lint, and test tasks on the increment of what you touched and what could be affected by what you touched, as opposed to the whole thing. So it pays more dividends. ⁓ Like, if you want to think of it as the relative difference between those, this is more valuable as your project gets larger. And as productivity starts to look worse and worse, when you look at like

What would the cost be of running the entire build from scratch over and over again? Like bluntly, I mean this took seven hundred and sixty-two milliseconds. Let's see. What does the whole thing take? I mean it's not it's not zero. Like this was 372\. This was 1.1. 411\. Like it's not the end of the world, but let's try build.

Sneha Mehra (00:08:22)  
learna run build since.

Sneha Mehra (00:08:30)  
Oop. PNBM learner run build. Great, so that was 726\. But what if we just said

PNPM build.

Alright, but the remember, these are in sequence. So we've got half a second here, and then another 0.7. So we're over a second. Like this is call it close to double? It's almost double, yeah. Close to double. And this is for a teeny tiny little project. And mainly it's coming from the fact that I avoided ⁓ building, like spinning up Vite to build this at all, which by the way is an incredibly fast tool. Like you're lucky if you're using PNPM and Vite.

And this latest version of LearnOt, this is all we're leaning towards like the fastest possible things available that you could be using to build. And so where where Learner really is useful, like say you're using regular NPM, ⁓ and you know, you've got a lot more files on disk as a result, and you're you're not using like turbo or you're not using well, in this case you'd be using NX, right? So it it's it can add up and it gets better with where we're going next.

—------------------------

Sneha Mehra (00:00:00)  
I'm going to ⁓ go into models ⁓ and I just wanna add a trivial change somewhere, like

Like here. What is this? Nope, that's in a noom. That's actually gonna mess things up. ⁓ we've got human toxicity, pet toxicity, livestock toxicity. Great. We're gonna have ⁓

Sneha Mehra (00:00:32)  
Alien toxicity with affected parts. ⁓ We're worried about how our vegetables will affect aliens. Alright, so I've saved this. ⁓ And you can see this yellow X. I've got

I've got something s like a code alteration that's been made. So what I can do is say PNPM, learn a run test since ⁓ equals

Course progress.

And what it's done is it can identify that ⁓ I have made gosh.

Sneha Mehra (00:01:17)  
There go. That's more satisfying to look at. ⁓ well, I've touched something in the models folder and I've made a change. ⁓ And ⁓ it it kind of like looks at the delta between my current working state ⁓ and any git ref here. So this could be a SHA, it could be ⁓ it I could have pointed to like origin, course progress, whatever you want to do. And so like if this this allows you to run tasks on a subset of

Not only what you just touched, but what is downstream in the dependency graph from what you just touched. So it's a lot more sophisticated than if you've ever used something like lint staged before, where it that that's simply just like looking at git diff, which files did you touch? Let me run things based on those files. ⁓ here we're getting something very important back that we had at the beginning of the class when this was a ⁓ when this was a

Monolith in a single repo. And that was we could make a low-level change to our models, ⁓ and then the entire test suite ran on the server, on the UI, the tests that run on the models. And it's very easy when you start separating things out into a monorepo. If you don't have the right tooling in place, you can fall into this trap where it's almost like the unit test trap, where you know, gosh, it seemed fine. We we made this code change, and like the tests in that package.

Past or the linting ⁓ on that package past or this type this small part of the project type checked, this lets you get back to that point where you're saying, all right, I have the benefits of a monorepo, meaning I'm not like boiling the ocean and saying, run ⁓ all the tests over again for the everything that exists in this git git repo. But we are getting the benefit of saying, ⁓ I touched something, ⁓ use your knowledge of like where was the thing I touched.

what are the downstream dependencies that it could affect and now we can end up ⁓ running a subset of the tests there. So let's let's see if

Sneha Mehra (00:03:29)  
Let's let's see if we can show an example of this actually working. What I should do now is revert this change.

Sneha Mehra (00:03:38)  
Let me touch something just in server. And the hope is this just runs.

There we go. Trivial change.

Sneha Mehra (00:03:53)  
Oop. Just the server wait, am I in the server folder? Could have fooled me. There you go. That's all that's running.

Sneha Mehra (00:04:04)  
The only thing that changed since my head commit was files and server. So we don't need the models to run again. We don't need that test suite to pass. That's upstream. Presumably those tests pass in master. ⁓ Like if I check things out and I haven't touched it, it's validated already. So this is a really cool way, ⁓ a cool benefit, a way to reap this benefit of being able to operate on, test on, type check on.

Lint on small portions of your project. And this is where we're starting to see some of the the lightweight promise of working in a monorepo come to life. But you need a a a build tool that's aware not just of file structures, ⁓ like file locations ⁓ in a directory, but the dependency graph of how all of these things are related to each other.

When we're doing learn a run test, is that our old like test ⁓ script we had that it's running? Yep. This is the this is the test script. And so how is it able to break up down to individual tests if it's just running a script? To be clear, and I'm gonna I'm gonna prove this. It is not in our root package.json, like

Sneha Mehra (00:05:24)  
I'm renaming that. ⁓ This should still run. This is going into each individual package ⁓ and running PNPM test. But it's making a determination of which packages it does that in based on what was changed and its knowledge of the dependency graph. So this is very much like the for each ⁓ of testing. And so what I would do here is I'd say, great, we've got like the test thing in CI. Yeah, we're going to run that.

But we can also have like test changed ⁓ and this is where we would say, look, PNPM is not quite specific or sophisticated enough to do this, and so I'm gonna grab this and I'm gonna put it up here.

Sneha Mehra (00:06:09)  
And we can just have this be like, this is what you would periodically be running over and over again. Like I made some changes, ⁓ lint it for me. Lint only the stuff that's changed. Build only the packages that that need to be rebuilt as a result of the code I have not committed yet.

And it's super powerful to be able to point to any git ref. Like, ⁓ you want to look at ⁓ like is this an incremental change that is being added on top of an existing PR? And you know that build already passed. And so you can just look at the last three commits that you've made. So ⁓ let's just do it dash dash since origin slash PR branch or whatever that is. So really the benefit with this is gonna kind of compound as you have more and more.

packages, I guess, in a monarchy, but right. If you just had one package, this wouldn't really do anything. ⁓ is that correct? Yeah, it's it's almost like the more this benefits you, the more you're able to break a sizable project up that already took a long time to build, to lint, to test. And the more you you're able to break that up, this approach is what lets you pay the cost, the computational cost.

Of like performing these build, lint, and test tasks on the increment of what you touched and what could be affected by what you touched, as opposed to the whole thing. So it pays more dividends. ⁓ Like, if you want to think of it as the relative difference between those, this is more valuable as your project gets larger. And as productivity starts to look worse and worse, when you look at like

What would the cost be of running the entire build from scratch over and over again? Like bluntly, I mean this took seven hundred and sixty-two milliseconds. Let's see. What does the whole thing take? I mean it's not it's not zero. Like this was 372\. This was 1.1. 411\. Like it's not the end of the world, but let's try build.

Sneha Mehra (00:08:22)  
learna run build since.

Sneha Mehra (00:08:30)  
Oop. PNBM learner run build. Great, so that was 726\. But what if we just said

PNPM build.

Alright, but the remember, these are in sequence. So we've got half a second here, and then another 0.7. So we're over a second. Like this is call it close to double? It's almost double, yeah. Close to double. And this is for a teeny tiny little project. And mainly it's coming from the fact that I avoided ⁓ building, like spinning up Vite to build this at all, which by the way is an incredibly fast tool. Like you're lucky if you're using PNPM and Vite.

And this latest version of LearnOt, this is all we're leaning towards like the fastest possible things available that you could be using to build. And so where where Learner really is useful, like say you're using regular NPM, ⁓ and you know, you've got a lot more files on disk as a result, and you're you're not using like turbo or you're not using well, in this case you'd be using NX, right? So it it's it can add up and it gets better with where we're going next.

—---------------------

11

Sneha Mehra (00:00:00)  
The next tool I'm going to show you is called SyncPack. ⁓ And what this helps you do ⁓ is ⁓ detect variants ⁓ in the versions of packages. Sorry, the variants in the versions of external dependencies that your packages depend on. So this is the tool that will help you identify how many versions of React are in your micro front ends. ⁓ And ⁓ potentially it will help you consolidate down so that.

you know, your your like ⁓ checkout times or so you're you're like NPM install or P NPM install times are are faster and you know you're you're starting to benefit from the idea of a monorepo, which means there's a greater sense of sort of the whole repo advancing together. ⁓ there there is an alpha version of this. ⁓ The like they have a stable release that's like I think they're V

⁓ and we're using their V13. So by the time you take this course it might not be alpha, but this this will work today. It's just a faster version of the tool. ⁓ So let's install it.

Sneha Mehra (00:01:16)  
And let's ⁓

Let's see what it did just now. ⁓ just package JSON and the lock file. Makes sense. All right. Now we're gonna do pnpm sync pack.

Sneha Mehra (00:01:32)  
Lint.

Sneha Mehra (00:01:37)  
No issues found, huh? ⁓ I have to create the issues. Sorry, I intended to to sneakily drop in a bunch of problems here that we would be able to find. And if you're copying and pasting from the course notes, you will find those things. But here's what we're gonna do. I'm gonna say we're using TypeScript 5.7 here. And in their server, we're gonna say we're using ⁓

Sneha Mehra (00:02:07)  
Head script five eight one. ⁓ And then elsewhere we have five eight three. So P N PMI.

Sneha Mehra (00:02:19)  
⁓ five eight three is wait, what is it telling me? No version five eight one.

Maybe that ru that package was pulled.

Sneha Mehra (00:02:36)  
Really?

Sneha Mehra (00:02:41)  
Did I type a number that was greater than five? Or greater than eight?

All right, so it's saying in server package JSON.

Sneha Mehra (00:02:54)  
It's there is not making any sense to me. Like, ⁓ we'll try that one. That's fine. A release candidate. Great. ⁓ now it's telling me about the UI package. Sorry, I'm I'm floundering a little bit as I try to ⁓ get us onto a varied range of different TypeScript versions. Great. So now I've got like some 5.7x version in my UI package, some 5.6x version in my server package, and in models.

I'm going to ⁓ float as far as I can in the 5.8 ⁓ release stream. We should get a lint a lint hit here. And we do. So we can see that ⁓ it's telling us ⁓ we've got three copies of TypeScript. ⁓ And these ⁓ these are the things that it has identified. Now personally I like the the check, like the lint command here.

It also has a sync pack fix, ⁓ which what that will do is go and forcibly bump these up. And if you're, I guess if you're super confident that that's gonna work out okay, maybe that's what you wanna do. But in a project of any significant complexity, like I would do I would be doing these one at a time. ⁓ And I might have ⁓ a check like this be part of a a nightly build or a weekly build where I could sort of periodically get a report.

and you know kind of keep tabs on like how much floating is actually going on, how much variance on certain packages is actually going on ⁓ within within my monorepo. So still still a good tool and like imagine the case where you have 200 packages in your monorepo. Like this ⁓ really will help you detect ⁓ you know where things are getting a little weird.

And especially for patch versions. Like I think they have some arguments. ⁓ Lint ⁓ help.

Sneha Mehra (00:05:02)  
⁓ where you can change like the specifier types, you can ⁓ you know, you can just like look for specific packages.

Sneha Mehra (00:05:20)  
Whether they're runtime dependencies or dev dependencies or peer dependencies. So ⁓ you know, this would be like if your mission is ⁓ we're reducing the variance in React, and we finally get everything on one React version, then you might have something like this, and in our case it would be.

Sneha Mehra (00:05:43)  
Something like that, where like I will inline this into CI. Like you have just introduced a new version of React. You can't do that. Like we've we fought hard to to dedupe and we're holding the line. And so this would be this is a nice tool that you can sort of inline into that and provide a very actionable failure message. I just love like this is a CLI that someone put a lot of care into.

—----------------------  
3

Sneha Mehra (00:00:00)  
We a question online. Are monorepos and micro front-end architecture interrelated at all? they are. Often you'll have micro front ends ⁓ arranged into ⁓ packages within a monorepo, right? So microfront ends is the concept of like instead of having a UI monolith, you're sort of composing together ⁓ sub subparts of it. And that that's like a great example of

Where you might want to model each of those as a distinct package. So there's strong encapsulation or encapsulation that you have a high degree of control over, you know, as it relates to ⁓ the contract between two micro front ends, or their contract with some core part of the UI that sort of cuts horizontally across all the micro front ends. Are monorepos a good setup for creating a library, for instance?

If I have one package with multiple projects and I want to serve that as a library.

If you have one package with multiple projects, I I would say I would flip that around and say it's probably if you have one project with multiple packages. So say you had ⁓ say you had a project where it was consumable as a library, but also you have a CLI for people that just want to use it, use some part of it in a a terminal. Say it's for image resizing or something. Well you could say, look, there's this core thing that's the library, and then on top of that, I'm gonna build the CLI, but I want to

develop and evolve these things ⁓ together. Like when I add a new feature or I I make some change to the way I represent data, I want this to all be something where I can have like a a single git commit and kind of keep these ⁓ synchronized over time. And in that case I'd break them up into two packages within one project. Your opinions on TurboRebo versus NX ⁓ my my opinion is they both meet ⁓ like I like them both.

Sneha Mehra (00:01:58)  
And they work a similar way. This this concept of having ⁓ well, like there are two two concepts there, which we'll get to more deeply when we start talking about NX. ⁓ one is the idea of like cached build steps. And you can think of this as almost it's almost like a pure function where you can state, like my linting task, it's the result of running that task, like both in terms of

Did it pass or did it not pass? And what was the output to the CLI? Like what was the output in the terminal or what was the file it generated? If you have your linter like piping out to a file. If that's the output and the input's the source code, and you can state, yeah, this is ⁓ like for a given input, the output will always be the same. Well, then you can cache that. And instead of ⁓ actually running your linter every single time, you could say, you know.

We haven't really touched anything since we ran this, including the version of ESLint that's being run. ⁓ so I'm just gonna spit the same output out. Like I'm gonna I'm gonna skip the step where we run exactly the same thing, and I'm just gonna give you the output. So that's that's this concept of like a cached task, if you will. And then what TurboRepo ⁓ or TurboBuild and NX both do is they will then have a cloud service where you can

Push these cache results. And what that means is even across your development team, you get the idea of sort of getting a fingerprint on the inputs and preserving the same output. And you can, you can sort of benefit from that across your team or in CI across multiple jobs. ⁓ So it's sort of like that second step is more second step is more about like more than just you sharing the same cache and getting the benefits of the same cache. And we're gonna, we're gonna do that.

We're going to ⁓ as we dive into NX and Learna as well, we're gonna see the benefits of that. Like it's it's an incredible ⁓ speed up for those tasks where the same ⁓ inputs ⁓ will always yield the same outputs. And it's up to you to state that that's the case for a given task. Given all of these benefits of monorepos, we're kind of going through, it almost sounds like you should always choose a monorepo, but obviously that's not the case.

Sneha Mehra (00:04:22)  
What do you kind of look for when you're like trying to decide if a project should be a monorepo versus like poly-repo? Like what are the indicators you're kind of looking for there? Great question. ⁓ for me, the the most significant bit is really about like what is the interrelated relatedness of the packages that would be within this project? Like the project, and like let me answer this in two ways. Like ⁓ if you're thinking about

A project on GitHub. Like let's take Babel, for example, for like taking very modern JavaScript and transpiling it so that ⁓ older runtimes can execute it. Well, you can have a nice set of Babel plugins in there. You can have the Babel core package, you can have the Babel CLI, and there there's a lot of sort of entanglement. Like there would be a lot of extra work if those were each modeled as a totally separate thing. Now ⁓ at at a company level.

where the the commonality factor comes in, like the the reason for grouping things together, it's usually about having some fairly sophisticated tooling and conventions that are that are applied that would be otherwise like very difficult for ⁓ in a polyrepo ecosystem to enforce. So that's probably number one. And then number two, if there's a lot of pain that some community, be it a company or people working on an open source project,

Like if there's pain associated with keeping everybody upgraded, were you to go the polyrepo path, often a monorepo is something to aim for. But like here are examples of things where I would not necessarily reach for a monorepo. Like if you're building a little node library that's sort of self-contained ⁓ and it has ⁓ you know, it has like a clear reason for existence, there's really no convenient place to sort of start dividing things up.

Or if you were dividing things up, it would just be for your own like internal purposes to get like encapsulation boundaries. Well, like there are other ways to do that. And I think at that point, you know, you're gonna see that it takes a little bit to get these set up. ⁓ now, once you are set up, it's it's fairly easy to manage them with some of the tools I'm gonna show you. But at that point, it it wouldn't necessarily be worth it to me. Like if you have a CLI tool you're building that's for image manipulation or something like that, like maybe.

Sneha Mehra (00:06:47)  
Maybe it makes sense to just consider that to be a like a monolith tool and it not divided into ⁓ sub packages.

—--------------

20

Sneha Mehra (00:00:00)  
Next, we're going to dig into some more advanced Monorepo build and task running tools. ⁓ The first is Lerna, and then we're going to get to using NX. So Lerna ⁓ in some ways it's the original JavaScript Monorepo tool. ⁓ It has been around for about a decade. But what's important to know is that the way it works today is very, very different than.

⁓ what you might have seen if you if you're poking at it a couple of years ago. And part of this is because ⁓ Lerna as a project is now maintained by the same team that works on NX. So to some degree, it's a little bit of a wrapper around NX that allows for sort of a simplified usage pattern. So let's begin by installing Lerna, and we're gonna make sure we're doing so while in the workspace route. PNPM DLX Lerna.

init. So again, DLX is download execute. That's like running npx if you're familiar with the npm equivalent. And so we're downloading Lerna and we're running the init command. And what it's going to do is create a learna.json file. ⁓ Gives us a JPM an npm schema. Sorry, a JSON schema. It detects that we are using Pnpm. Lerna works really, really well with Pnpm. It does work well with npm workspaces.

and with yarn workspaces, but just the level of configuration you need in the PNPM world, it's almost nothing because we already have this workspaces ⁓ YAML file. ⁓ And like you would otherwise have to be adding more information into this file if your ⁓ package manager of choice didn't make it so easy to parse that out. ⁓ similarly, Lerna, because part of what it does ⁓ is

allow you to install workspace dependencies. ⁓ Learner knows we're using PNPM and it will follow that workspace colon star convention. So it it really doesn't this is not interfering with all of the the norms that we've been discussing so far.

Sneha Mehra (00:02:16)  
I'm gonna install Lerna as a dev dependency as well.

Sneha Mehra (00:02:25)  
Just to make sure I always have it. Alright, so ⁓ let's look at what it what what it ⁓ like let's let's use learner to run builds here. And we'll do that by saying pnpm learna run build.

Sneha Mehra (00:02:43)  
And you see Lerna powered by NX and it's going to each of these packages and it's ⁓ it's building each of them. ⁓ So it can build, it can test, it can lint.

Sneha Mehra (00:02:58)  
And behind the scenes, NX is doing the lifting here.

Sneha Mehra (00:03:06)  
⁓ we have a lint error somewhere. ⁓ we we must have forgotten something a little ways back. These are files in a dist folder. We shouldn't be linting things in a disk folder. Let's let's ⁓ go ahead and fix that since we've got this error here. ⁓ there has to be an ignores already. Perfect. We'll do this ⁓ and dist. That should work.

Sneha Mehra (00:03:39)  
Great.

Sneha Mehra (00:03:45)  
So ⁓ an X detected a flaky task. I can I can tell you what this means. Can't tell you exactly why this one's flaky. I'm I'm not sure to be blunt. But ⁓ part of what NX does, it's cached task artifacts. Meaning ⁓ for a given set of inputs, ⁓ a a deterministic task will give you a consistent set of outputs.

And what NX is detecting here is that ⁓ something's changed. You know what it could be? Let's see if we get another flaky task notification here. I was gonna guess that it's the fact that we changed our config, but I'm just speculating at this point. So anyway, it's it's running these things. ⁓ and ⁓ let me show you ⁓ something that PnPM has a little bit more trouble doing, or at least you'd have to run a more elaborate command for this. Pnpm.

Learna, run, build, lint, test, check, stream. Stream means like if you noticed when we were previously executing tasks, we kind of just saw the green check marks for them completing. We weren't seeing the output. You only if things error will use output. But if we run this, you can see, all right, it's doing a bunch of stuff. We're actually seeing the build output and it all completes. If we didn't use stream,

it'd be a much more abbreviated thing where we just see a bunch of stuff happening in parallel.

You have a lot of control over ⁓ how tasks are executed. We could say I only want two tasks running at the same time. And you can see I've I've used dash-concurrency too. So this would be if you've got other things happening on your machine and you don't want, like say you're ⁓ Twitch streaming while you're coding and you don't want to just peg every core on your CPU for build and like your video will start to stutter. This would be how you'd sort of throttle it down.

Sneha Mehra (00:05:47)  
And we can just ⁓ the the default here is to let it loose ⁓ and run everything in parallel, but you could you can explicitly specify that running dash dash parallel, which we don't need to do. So great. It's sort of an abbreviated syntax, it's kind of cool that we can like comma separate a bunch of tasks that we want to run and it runs them on ⁓ a bunch of different ⁓ packages in our repo. There's also like ⁓ scope.

Sneha Mehra (00:06:18)  
Sorry, this has to be ⁓

Something's like seeds seeds UI. So here we were saying I want to run ⁓ all these tasks, but only in Seeds UI. Or we could do this, right? It's any any text fragment here, basically, that that can resolve to some subset of your packages. So that's great. ⁓ But let's let's look at something, and like PNPM can run tasks similar to this. Like it still is dependency aware. But here's here's something that's a little bit different.

—--------------------

4

Sneha Mehra (00:00:00)  
Let's get started. So, first, you're gonna wanna make sure that you have PnPM installed. In this course, it is actually going to be important that you use this specific ⁓ package manager. npm and yarn also have kind of some similar capabilities, but we're going to be using some PnPM specific concepts here. And so you'll need this. ⁓ you can either ⁓ install it directly by clicking on this link here and following the installation instructions, or if you happen to have Volta installed.

⁓ you can just run Volta install PnPM and it will set it up globally for you. Remember after, like if you have to install Volt Volta, here's how to do it. And then like after you complete this step, remember to close and reopen your browser. It sets up an environment variable for you, and that will need to be read. Or if you're clever and you want to like source your ZSHRC to to load that variable, you can do that too. ⁓ and that's that's PNPM.

And then you're gonna want to check out the GitHub repo. And here's the name of the repo: it's Mike North, TS Monorepos V2.

Sneha Mehra (00:01:09)  
And then you'll want to enter the repo, install the dependencies, and the way you'll know things are working well is these three tasks should complete successfully. You should be able to build, lint, and test.

Sneha Mehra (00:01:24)  
And then finally, if you want to test that the dev script works, ⁓ this is going to start the back end, start the front end, and you should see a URL you can click on. ⁓ Well, you you can test the API by hitting this endpoint, or you can load the UI and you should see something that looks like this. If you have a monorepo containing packages of several different languages like Ruby, Python, and Java, how would you integrate those type of packages in a single repo?

⁓ it's it's certainly possible to do that. Like ⁓ I would at that point reach for a tool like ⁓ Basil, which is kind of a language agnostic build tool that produces ⁓ you know, her hermetic build artifacts and and is quite fast at building bluntly like most of the languages you care about can work with with Basil. But ⁓ that's that's kind of what I would work work towards.

Like you'd you'd pick tooling like that that is suitable for polyglot ⁓ monorepos. Now that you've set up the workshop project, ⁓ let's let's start it up and make sure that everything works. So the first thing you want to make sure you do is make sure you have pnpm ⁓ installed and then it's in a ⁓ you know it's in a relatively recent version. I'm on 10.12.1. You can run pnpm i to install all of the packages. I already have them cached here.

so you might see some some spinning going on there as it pulls pulls things down. And then you can run pnpm build. You should see some sort of build up, but that looks like this. So build lint test should work. And then if you run pnpm ⁓ dev ⁓ and you go to the workshop project, so this this is the UI ⁓ the UI endpoint, this is the API endpoint.

So if you click the API endpoint, you'll see like localhost 3000\. At the root, I am not serving things. But if you go slash API slash seeds, ⁓ you should see some JSON. And similarly, if you go to the UI, you should see some visual representation ⁓ of that JSON bunch of data coming through that is just like it's enough that we have a back end, front end, and something that's shared between them, and that is like plenty sufficient for us to think about monorepos.

Sneha Mehra (00:03:47)  
So with that, ⁓ let's talk a little bit about what's inside ⁓ what like what this project consists of. And this is a ⁓ this is really important to do like before you begin your monorepo journey. And I really what I mean is if we're setting out on a journey to turn this this single project, right? We only have a root package.json here. Here's our source code, all this stuff. Like we're gonna start.

begin a process that normally you would conduct in cycles where you'd say, All right, like I've got something that is ⁓ a little bit too monolithic and I have reason, so like some reason to sort of break things out and I want to like factor out a package. What is that package? Like can I identify things that are ⁓ separable? And like this is sort of the step, like before we even start talking about the Monoribo, you want to have an understanding of like

What are we dealing with here? What is logically separable ⁓ in this app? And I've I've kind of done that work ahead of time because you're gonna have to do this yourself in in your own ⁓ in your own app. But like we have a fairly obvious case here where we've got some UI components ⁓ in this ⁓ lib folder, right? We've got some Svelte components and some ⁓ something like ⁓ dealing with state ⁓ in in the Svelte world. We're not gonna worry about the contents of those files ⁓ for now.

And then we've got a folder in source called server. And if I look at those, we've got like some Express stuff, ⁓ like Cores Middleware for Express. We've got something called load data, which seems to be engaging with the file system in some way and loading like loading a file. And then we've also got models. And these are just TypeScript interfaces that represent ⁓ what is a C?

packet, like what is the data structure here? And so those those are some good candidates here. So we've got our models. ⁓ we've got our server. And in our server, ⁓ you know, we have this express stuff as well as let me remember where I put it. ⁓ this is important here. We have this folder called data up here, which we're gonna want to keep track of. This is the source of truth for the data. So the back end just oops, it's a huge file, more than 5,000 items. So

Sneha Mehra (00:06:14)  
The YAML VS Code plugin is barking at me. ⁓ And ⁓ so so like that's that's really what this consists of a data, some models, a UI, and a server. And that already gives us some a good starting point for ⁓ starting to peel some of this apart. as well, we have we have this like utilities thing, which is just ⁓ some formatting stuff. Mostly this is used in the UI. I think exclusively it's used in the UI. ⁓ And converting an RGB color to

Like a CSS value for RGB. ⁓ so that's it. That's all this app consists of. Now we're gonna begin the process ⁓ of ⁓ taking this package as it exists today and taking the smallest step we can to making this into a P N PM workspace. At the end of this step, we're gonna be able to run the same tasks, right? Build lint test. ⁓ just just so we we

Capture the starting point here.

Sneha Mehra (00:07:18)  
I can also run prettier with like to format things.

And I can run this check command, which ⁓ it's it's type checking, which ⁓ requires a special command in the Svelte world as well. So you can think of them like I've got a bunch of commands here, and they're all in the package.json. You know, we've got build, lint, test, some test variants here, lint format. What we want is ⁓ to see we that we get most of these working, but in a world where ⁓

We're we're we've restructured the way the code is in this project and we can sort of see that we've got a workspace, but with only one package in it. And then we're gonna begin the process of factoring things out.

—------------------------------------  
22  
Sneha Mehra (00:00:00)  
Alright, so everything's going well. You're using Lerna. We can we have like ⁓ we we're able to operate on small portions of our project. We're able to run tests or lint or build only what's changed and what's affected by what's changed. ⁓ let's let's see if we can use the sort of the full fat build tool that's underneath this, which is NX. ⁓ And we're gonna need to globally install NX. And we're gonna I'm gonna really just ⁓ sort of show give you a little.

⁓ kind of preview or an intro into what this can do. But there's so much depth with this tool. And I would encourage you to learn more about it. It's it's ⁓ it's quite powerful. ⁓ So you're gonna want to install a global version of the tool ⁓ and you can either s do like an npm install-g for global. I use Volta. So you could say Volta install nx and this will like set up a global version for you. You could use PnPM if you want.

⁓ but ultimately you just want have something that lives outside of the project. Okay, and now you're gonna run pnpm dlx nx init. And this this is like npxnx init. Basically, what we're saying is ⁓ run this init task, sorry, run the init command using the nxcli and it's gonna create a config file for us.

Well, actually, this is more of a choose your own adventure thing. So, making this ⁓ the main thing we're looking at. So which scripts need to be run in order, for example, before building a project, dependent projects must be built. So this is telling us like which of these tasks need to be sequenced versus which don't. And I would say, sure, let's go with test. I mean, we could test everything if we wanted to.

all like in parallel, but I kinda like to see that ⁓ my low level things pass their tests ⁓ because otherwise you can have like low level tests that fail and then of course higher level tests fail because something was broken at a lower level. So ⁓ let's say that there's that. ⁓ I'm not gonna worry about these three and linting, I feel like linting can happen in parallel across my whole project. There are no there's no sense of like a build order there ⁓ necessarily.

Sneha Mehra (00:02:26)  
no, sorry, I'm wrong. ⁓ For linting to work properly, we need declaration files to exist for projects that we depend on. So there is some sequencing there. Let's leave it unchecked for now and let's see what happens. ⁓ Okay, which scripts are cacheable? This means they produce the same output, given the same input. Build, lint, and test are usually ⁓ falling into this category and the others are not. So I'm gonna say test.

Yes. Lint and build, yes. Check, yes. ⁓ And format. Sure, we can say format, yes. The rest, no. And here's here's the mental model I want I want you to use. Like you have a code base and then you run a task. And there's some combination of like standard out and standard error that that task produces, and there may be files that it produces.

And if you could say running that task over and over and over would have exactly the same result, it's gonna make the same files, it's gonna have the same standard out. You should check these boxes. Dev is different because there's another thing that happens there, and that's like I'm engaging with a UI or I'm making requests and log lines are coming out. I don't want that to be cached. I want that to be live data that's happening. Because what's gonna happen when we complete this?

I want you to imagine, because this is ⁓ in in essence what's happening, when you run test and you haven't changed your code at all, instead of running your test, ⁓ nx will just spit out the same test results that it already knows ⁓ it should have, based on the fact that you already ran them and you haven't touched anything. And it will appear like the command is running instantly. So that's like when you understand that that's how this is working, that apply that.

Mental model to which of these things you should be ⁓ taking off. The test UI and the watch thing, like these are more ongoing, ongoing things. Coverage is fair. This generates a code coverage report. It's going to be the same code coverage. If I don't add a test, the code coverage report should look exactly the same. ⁓ So let's leave it at those.

Sneha Mehra (00:04:47)  
Does the test script create any outputs? ⁓ If not, leave blank, otherwise provide a path relative to the project route. It does not. The test coverage task does. Yes, this does.

Sneha Mehra (00:05:04)  
Does the lint script create outputs? It does not. There's no report it's creating, just standard out. ⁓ dist. ⁓ And

⁓ well, we're gonna find where this is and we'll edit in the config file. I'm not sure how this is gonna behave if I add comma separated things here. But it's the dist folder and it's the tsbuild info files. Does the check script create any outputs? No, it does not. Does the format script? No, it does not. ⁓ Okay, now it's going and doing its thing. Installing a bunch of dependencies. Do we want remote caching?

I'm gonna say yes. Now eventually NX will charge you for these for for remote build caching. ⁓ give you a generous free free amount. So ⁓ it's worth worth checking out and in my opinion it's ⁓ it's worth paying for if fast builds are some something you're really chasing. It's really just to store, not necessarily like it's storing the console output or the files that are created.

And a hash of the inputs. So that and that means that like if you build something on your machine, and then I build something on my machine and the the inputs are exactly the same, ⁓ I get to benefit from your pre-existing build result. Like that is work that is already done. It it does not to be need to be done on my machine if we're really honest about those builds being cacheable. And it it it is only as good as your judgment around whether builds are truly cacheable or not. Like I'm using an

M4 MacBook, somebody else might be using an X86 processor, and there's some native dependency that we both need. And maybe that's gonna screw things up. But like for CI machines where you can routinely rely on them being like, you know, there like we're running this in Docker. It's just gonna be the same thing no matter what. Like ⁓ no hardware access, ⁓ like no direct access to hardware. So it's like gonna be very, very predictable. ⁓

Sneha Mehra (00:07:10)  
That touches on what I was gonna ask. Would you trust the caching enough for like if you had like a test suite that was running before a deploy? Is there is this reliable enough that you would consider caching that or would you after it's been run on local machines ⁓ before the main deploy ⁓ actually run it? I would trust this like i I if I were I'm putting my from from remember where I work, like to

We gotta make sure that we're actually running tests ⁓ right before deploy, for sure. But if you were saying, well, the PR builds, ⁓ like the ⁓ validating code in those PR branches, ⁓ knowing that ultimately, before things are deployed, we're running the pipeline on the main branch anyway. I would totally put something like this in place. And especially upstream of that for like developer builds, like keeping that nice and productive. Basically, the further

The closer to authoring you get, the more I'm willing to tolerate using a cached build. But and and it also depends on the task. Like if it's linting, I'm much more okay with that. But if we're saying the build output, like the actual compiled output of the TypeScript, I would want that to be created fresh. You know. Does that make sense?

Yeah. I think just any time you're caching things, I've run into stale caches enough that Totally. I'm skeptical, but it seems like a cool cool idea. It's a cool idea and I think it's it's a no-brainer for some kinds of tasks. Like, would I be fine with prettier being cached? Absolutely. Like single quotes, double quotes. I mean, I'm sure you could someone can find like a significant security vulnerability that occurred because someone used double quotes or they should have used single quotes, like

It generally is not going to

Sneha Mehra (00:09:06)  
All right. Which plugins would you like to add? And it's giving me an opportunity to check things off. ⁓ NX is pretty good at inferring what already exists in your project and setting up plugins for itself so that it can engage with ⁓ with these kinds of things. Really w ⁓ and what are these plugins for? Think of them as like replacements for your N ⁓ NPM tasks, right? Where it knows how to invoke Vite. It allows you in

an nx config file, instead of passing arguments to the Vite CLI, you can have configuration that you check in. And it's a little bit more maintainable that way because you're not trying to look at like 12 flags that you're passing to the CLI in some shell script somewhere.

All right, we're gonna install those. ⁓ do you want to start using nx in your package JSON scripts? I want to say no to this. If I were to say yes, what it'll do is it will reach into my package.json, grab the existing scripts, ⁓ move them into an nx task that is effectively like their shell out task, right? Just like run this command. And then it would replace ⁓ everything in the package JSON.

with an NX-based invocation of that task. In fact, we could do it both ways. Let's try it this way, and then we can reset and try it the other way if we want. But I would rather show you like one off how this is going to work.

—------------------------  
19

Sneha Mehra (00:00:00)  
Next, we're gonna take the output of API Extractor and use it to make some simple API documentation for our models Monorepo package. ⁓ And this is quite simple. ⁓ we first need to ⁓ install API documenter. If it's not installed already, let's check. Microsoft.

API documenter.

And doing this within the models package is right. Great. Okay, dependency added. ⁓ And now I'm going to run PNPm API documenter markdown. ⁓ Temp Dash O for output docs. So we're saying I want to invoke API documenter.

I want my docs in markdown. The other option here is a YAML file, which you can feed into other documentation systems like Docusaurus. ⁓ I've personally never figured out how to do that, but it it is it is ⁓ just a matter of aligning those two formats. And then ⁓ pointing to this temp folder, right? So if you if you customized where this temp folder was being stored, ⁓ you would want to keep that consistent. And this, by the way, is something you don't want to commit to git. So while we're saying that.

Let's let's go to our gitignore, make sure we're ignoring that.

Sneha Mehra (00:01:32)  
Temp. Save. And we want to see that get grayed out. Perfect. And then the output folder is docs. Let's let it run.

Great. ⁓ deleted the old output from docs. Well, none was there. And then ⁓ what was written was a new feeds models package. And look, a bunch of mark markdown files. Let's look at our index.md in VS Code, get rid of that. And you can see here how ⁓ if we had more than one library, you could lump these things together. And you could have like all of the different packages you might have in your mono repo.

And then clicking into one of those, you can see all right, we've got like a seed packet collection model. There wasn't a lot of commenting here, but this this is all just coming from code comments. Seed packet model. Here are the different types and the descriptions. This is all derived from comments and ⁓ type information. So, like, particularly if you're ⁓ if you're using this mostly for your for

like internal purposes. It's it's kind of nice to have this in here. Like you can browse this right within VS Code. ⁓ you don't have to go to some other Docs website. You don't have to switch out of this tool. I I really like the idea, particularly if if you're not making this a product and you really care about the visual appearance of the DOCs greatly. ⁓ these end up being just a very browsable thing. And ⁓

If there happen to be code comments on things like that gets percolated through, but mostly it's type information ⁓ and ⁓ it does have a nice sort of multi-page structure where you could easily track the changes to things over time. So like that's another aspect I like about this. ⁓ If if you check this into Git, this lets you look at the history of these things over successive builds, and you can see them kind of as ⁓

Sneha Mehra (00:03:38)  
It's like almost like a time machine of the surface without having to worry about, you know, the implementation and refactoring within the methods, like it's a nice sort of ⁓ distilled piece of information.

So this is API documenter.

—------------------------------

18  
Sneha Mehra (00:00:00)  
So we've made some changes.

⁓ and ⁓ we can run the report again.

Sneha Mehra (00:00:10)  
And you know, take as a given that some of these have been ⁓ have gone away. Now I want you to look at another folder that was created here. It's this temp folder. So what we have here is this thing that I referred to as an API report. Here's what it looks like.

So, this is the API report. This is a marked on file. If you commit it to GitHub, it will render beautifully on the website. And it shows exactly what is being exported from this module. And it also encodes warnings. So there's a lot of stuff here. Like this is a complicated ⁓ type. Now I want you to imagine you're reviewing a pull request, and instead of having to

Piece through and see, like, you've changed this type and it is used in a function signature that happens to be public. But the where you changed this type is like deep in some internal code. Like having something that's so clear to see, that's this sort of a roll-up of the exposed API surface of this library, is super useful because you can just see it in terms of a diff here. It becomes clear as day whether whatever you changed in a pull request.

Has kind of ripple effects that downstream users of the library will feel in a way that is either breaking or non-breaking. But that that's a good reason to bake something like this into your build process or to build a bot around it so that it like comes into a pull request and you know says, here's the updated API report. Makes it super, super clear to spot when a given code change has ripple effects that affect.

The library's outer surface.

Sneha Mehra (00:02:03)  
There's another file here. It's called models.api.json. And this ⁓ is going to lead into our next topic, where we talk about generating API documentation based on this.

—-----------------------

1

Sneha Mehra (00:00:00)  
Welcome to TypeScript Monorepos V2. I'm Mike North. I'm a principal, staff, engineer, and product architect at Stripe, and I've been a front-end master's instructor for over 10 years now. And today I'm going to talk to you about monorepos. So, first off, ⁓ what is a monorepo? It is the concept of having a single Git repository that contains multiple packages. Sometimes ⁓ companies will have

Like a language level monorepo. So Stripe Stripe is one of these. We have a monorepo for all of our Java code. All of our Ruby code is in a different repo. And then we have a big JavaScript monorepo as well. ⁓ And there are also open source projects. Like if you've used Babel before, there are many others where ⁓ it's kind of one project, but if you look into it, you can see there are multiple distinct packages, and you will find each of these packages.

on NPM and they'll each have a version. Sometimes you're using just a couple of these things, not the entire contents of the repo in order to do something. ⁓ So ⁓ what what are the benefits of organizing your code this way? Like why would you why would you aim for a monorepo instead of kind of the polyrepo approach where you would treat every library you have as a separate git repo? Well first off there's the issue of dependency management. If you've ever worked on something, say in the open source, you know

JavaScript ecosystem, you'll make a change in the library, like you'll open up your PR and it'll get merged, and then you have a whole bunch of work where you have to go to other packages that depend on this and then run npm install or pnpm install or whatever it is, and you know, float that lock file pinned version forward a little bit, and you kind of have to pull it through ⁓ all of the things that depends on or that depend on it. And this has to happen in multiple layers, right? You could

You could be several layers deep in a dependency graph, and a small change could result in you having to do a lot of work beyond making the code change just to propagate through a polyrepo ecosystem. Well, the idea of monorepos is ⁓ you are kind of you're often locally referencing other parts of this software project, ⁓ and you're kind of evolving all packages ⁓ at once. You could open up a single pull request that

Sneha Mehra (00:02:21)  
touches three or four different packages within one Git repository, get that merged, and there's really no follow-on step that you have to take. So that's part of what makes it attractive. And especially for like a large company with a lot of different things going on, a lot of different packages that they have ⁓ set up, it's kind of attractive to say we don't have to deal with this problem of ⁓ you know, somebody made an improvement, but it's a whole new set of work to sort of propagate that that improved

Whatever it is, that that chunk of ⁓ value through our ecosystem. And you end up with people that are sort of like two, three years behind on some dependency just because they haven't gotten around to doing that work. A additional benefit of this, and it's possible to do to get this benefit in a polyrepo ecosystem, it's that in a monorepo, it's generally easier to be able to say, I've made a change in a library that's say three levels deep in a dependency graph, right? Like say it's some UI.

That depends on some library, that depends on some other library, that depends on some other library, and you're right there at the bottom. Well, it's it's a lot easier when everything's in one Git repo to sort of run the entire test suite, not just for the the library you changed, but everything that depends on it, and ensure that you're not just validating against some small set of unit tests, you're validating the whole project when you know composed together. Did this actually work? Am I going to be able to roll this out?

—--------------------

5

Sneha Mehra (00:00:00)  
So the first thing we're gonna do is ⁓ we wanna create a folder for this new package.

So we're gonna say make dir-p, so we create all subfolders necessary to make this work. And we're just gonna call this like UI. And I'm I'm creating a folder here called packages. This is where I'm gonna put all of my packages. There's ⁓ another approach that some people like to take here, where they'll have two folders, one called apps and one called libraries. And they'll sort of separate out their leaf level dependencies, like CLIs, UIs, servers from.

Things that are sort of in the middle of the dependency graph. But it it's just purely an organizational construct and we're gonna keep things simple for now. All right, next. ⁓ we're going to take all of the code ⁓ and except for a couple files.

And if you're following the coast course notes, I have this list of files too. So we're gonna select everything, deselect packages, ⁓ and what we're gonna leave is the git ignore, the nvmrc, which is just make sure we're all using the same node version. We're gonna leave the prettier config, we're gonna leave the pnpm lock file. ⁓ we're going to ⁓ leave the readme for the whole project. ⁓ And let's

Let's move everything into ⁓ oops.

Sneha Mehra (00:01:30)  
Let's move everything into packages UI, except for those things.

Sneha Mehra (00:01:37)  
Okay.

And I'm definitely hit no here. Like this is ⁓ gonna wreak havoc if you update imports for known modules.

And the next thing we'll do, and I think I actually had a version of this I I accidentally checked in here, but but ⁓ no no, this is fine. P npm workspace, you want to have this also in your root level. Now, I've added some things here. Like normally when you'd start this journey, you wouldn't have this file. All all we're seeing here is ⁓ these files, sorry, these npm modules have a post-install step that requires like a little bit of a build.

Once it's installed and it's on disk. And this allow lists those two packages so that they can do their post-built step. Normally you'd be blocked and you'd see a message and it'd say, like, do you want to let these two things run? So I've taken care of that for you. But the meaningful thing we have to add to this PnPM workspace YAML file is this: packages, and then we're going to say ⁓ we're gonna s in packages say ⁓ packages.

Sneha Mehra (00:02:52)  
Slash star. And what we're saying here is like we're telling PnPM that this is where it can look for subfolders, ⁓ each of which represents ⁓ a package. And now ⁓ we will need so if we close up packages slash UI, ⁓ we're gonna need a package JSON at the top level of the project.

And this is going to be really thin. ⁓ we're gonna give this package a name and it's gonna be called seeds. ⁓ And we're gonna give it a version, and we'll just say ⁓ 001\. ⁓ And we can say it's private. True. Well, private means is ⁓ like your your package manager will f refuse to publish this to npm if it's private. And there's one more thing that I want us to move here.

And that is ⁓ if you go to your package.json, you'll see there's a little Volta object at the bottom. This is kind of like the NVMRC, just make sure that if you have Volta installed, we're all using the same version of Node. This only really matters, or at least using a monorepo the right way, in my opinion, you'd say, we want one version of node that you're using across development of every single thing in this repo. So I'm gonna cut that, bring it to our root-level package JSON.

Paste it and save. One last change we have to make. If we go back to our UI package JSON, the one with all the dependencies in it, we're gonna want to rename this, right? Like originally this was the seeds project, and we're going to rename it to this. Now, what I've done here is I'm kind of leaning into this concept of ⁓ npm scopes. So if you wanted to put this on npm, if I wanted to publish this project.

I could go and reserve, if it's available, the npm scope seeds. ⁓ And that means this is kind of a prefix that I can use for any packages that I wish to publish. So for an open source monorepo consisting of TypeScript libraries, it's very, very common for you to have an NPM scope that kind of says like it's it's another, it's sort of in the package registry the monorepo concept. It's saying these are all interrelated packages for.

Sneha Mehra (00:05:11)  
you know, for some reason. Sometimes they're used for other things too, like ⁓ at Stripe we have an at Stripe scope ⁓ and we have multiple repos, but it's sort of an indicator that like this is a real package ⁓ from from Stripe. Or a lot of our some of our real packages are there.

Great. Okay. ⁓ this is going to be a recurring theme. I have touched I've touched some package.json files, therefore I must run PNPMI. And before we do, I just want us to examine the state of the world. There's no node modules folder anywhere here yet. And there there is one, of course, here, like it was part of what we dragged.

Sneha Mehra (00:06:00)  
And we should see ⁓ a little node modules folder up here. Now it's not very interesting. ⁓ there's not much in here yet, but we'll we will get there. You can see that there's some some evidence of some things happening. Like, hey, look, this is a little a state file, and like we've recognized that there is this package, and like I have the 001 here. So clearly ⁓ you know, something's taking place. So now ⁓

Something unfortunate's happened here though. Like what I wanted to do was say this still works and the build command's not found. And so we have to figure out a way to sort of wire this up. The solution to this is going to be going back to this ⁓ package.json file at the root of the project, and we're going to be creating some scripts. ⁓ And the first script we're going to create is this build pnpm color.

run dash r build and i'm writing this in the most explicit way pros possible so that we you know we can see see how this works and now if we were to run pnpm build ⁓ we can see hey look there's you know vite doing its thing and it's building the project so what what does this mean? Well first we're saying ⁓ I'm running this build task and I'm doing so recursively like I am

Going into each package that I find in the workspace as defined in this workspace YAML file, right? Like I'm going to everything I find here, and by the way, you this is how you do multiple oops, except with real YAML.

Well, you'd have multiple listings here. I think you just can't have the same thing over and over. ⁓ but like it's going through every package that that it can find, and where a build task exists, like as long as a package, like our UI project, has some build script, it that is going to be run. And this, all this does is make sure that the colorful output is is preserved. So to to get everything wired up nicely.

Sneha Mehra (00:08:17)  
It's gonna be super easy. We're gonna create a bunch of these. ⁓ And we're gonna say this one is lint and this one is ⁓ test and this one is dev. And ⁓

Sneha Mehra (00:08:38)  
This one will make format ⁓ and last one is check.

Again, this is this snippet is in the course notes if you want to make that real easy, but it's just like all we we haven't really changed what any of these things mean. ⁓ we have simply now said, all right, like there's sort of a fan out. For for the all all of our one packages, ⁓ we're going to run these things. And so now you should be able to do pnpm test and it should work. Pnpm lint and it should work.

the NPM dev and we should see the startup where like there's the client, there's the server. Everything works well.

Sneha Mehra (00:09:27)  
⁓ yeah, one more thing I want to show you all that's important about how we just did this.

Sneha Mehra (00:09:36)  
Let's look at our git diff and let's add everything. ⁓ And sorry, I can get rid of that. ⁓

⁓ and I want to see this as a tree so that I can close the tests. ⁓ In fact, the whole UI package. So look at what's happened here. Like we we have a PnPM lock file. Because we preserved this, because we didn't just sort of like nuke it or or leave it in the UI sub package. All that's happened here is like a little bit of restructuring up at the top. And there are obviously like about 5,000 lines that haven't been touched.

Very important, like this is one of the benefits of ⁓ Pnpm. The lock files are much more human readable. And ⁓ like I if you had done this with like a yarn lock file or a package lock.json, like you might not see such a clean result here that would give you confidence. We have accomplished this reorganization of our code without releasing any of the locks that we had on all of our packages. Like you don't want to be doing that.

At the same time, you're sort of reorganizing your code because that's like we're trying to do one thing here, and that's move some files around and get some NPM scripts wired up. So ⁓ so that's that's nice. And then furthermore, if we look at ⁓ the way that this code is being treated, ⁓ it's important at this step, like you want to see ⁓ all of the changes you've made is being regarded as a rename.

Who why why is why is having this as a rename a good idea, as opposed to like a deletion and an add? Which you would get if I was messing with these files, ⁓ if I was making changes in the files of a substantive nature, as like in addition to making this move. You you get to preserve your git history. sorry. Right? Like you you you don't want sort of the the old history of this file to end and then a deletion happens and then a new file pops up into existence and like there goes your git blame.

Sneha Mehra (00:11:41)  
Because you've lost all of that history. It's very valuable to preserve. And so this is why when you're doing this step, I would advise like this this is a good first step to do. Don't continue refactoring here and separating things out into different packages as part of that. Because when you're doing that, you're introducing ⁓ a much stronger possibility that you're gonna lose history on some of these files. In this case, we're getting really well, in this case, it's such a simple operation we're performing that.

you've got a ⁓ very good shot at these all just being ⁓ sort of a reorganization, which is exactly what you're doing.

—--------------------------

2

Sneha Mehra (00:00:00)  
So in this course we're gonna be working with a small project that is kind of the the the minimum example or close to the minimum example ⁓ of a monorepo that has like some some layering and in dependencies. So we're gonna start with everything being together. It's it's a s a simple Svelte app with a node server and you know, some things that both the server and the UI need, and we're gonna factor it out into a monorepo.

⁓ we're going to show how we can automatically link ⁓ different packages within the monorepo. So you can just sort of edit whatever you need to edit, and those changes take effect immediately. You can just sort of still treat this. ⁓ You get the same benefits of the monolith, despite it being broken up into a bunch of packages. ⁓ we're going to ⁓ towards the end of the course get to this benefit of ⁓ dependency aware tech.

Dependency aware task execution for build, lint, and testing. And what that means is I touch a low-level dependency. I want to run ⁓ only those tests that could have potentially been affected by my changes. Not just the package that I touched, but the package I touched and anything that depends on it. Like let's run just those lint rules. Or sorry, let's lint just that subset of packages, or let's test just that subset of packages.

⁓ we're going to ⁓ explore the idea of monorepo packages. Like sometimes some packages are more coupled than others, where you might think of there being some sort of private API between packages, right? Where sometimes things are just more intimately entangled in in some cases than others. And I'll give you some tools where you can deliberately create those boundaries and make sure that your TypeScript types are correct and helpful in cases where.

You know, you you deliberately want like a lot of deep private access between packages. And then in other places you can say, hey, look, there's this sort of blessed surface for the package. Like in the general case, you should only be, you know, integrating against these points. ⁓ we'll also look at ⁓ automatic code documentation as well, right? How can you have some some nice generated docs that don't require a human to go in and keep updating things?

Sneha Mehra (00:02:27)  
that explain, you know, all of the packages in this Monorepo project ⁓ and you know they're kept up to date automatically because they're derived from from code comments and from the types themselves.

So ⁓ these are some of the tools that we're gonna be exploring today. We'll use PNPM, which is ⁓ in my opinion, like the package manager or yeah, the package manager that you wanna use for TypeScript moderos today. ⁓ we'll be using the latest version of Lerna, which ⁓ behind the scenes it it is leaning on NX for a lot of things. And so we'll we'll start with Lerna and then we'll eject out to NX when we have ⁓ you know a need to do so.

So we'll get through the package managers ⁓ and the build tools. We'll talk about ⁓ we'll use API extractor and document documenter to create a very deliberately controlled API surface for each package in the monary, but it'll only matter for like libraries in this case. ⁓ but we'll get auto-generated docs and we'll get what's called a ⁓ called the declaration file rollup, which is kind of like the a single DTS file that represents the whole library.

And you'll be able to tag individual methods and fields, anything that's exported as you know, internal or alpha or beta maturity, and you'll get the right roll-up generated depending on whether you're using alpha or beta features in this library, or you're using sort of the stable version. ⁓ we'll talk a little bit about GitHub code owners, right? This if you've ever used this before, at your root repo level, you can have a code owners file that sort of ⁓ governs who is automatically assigned to review.

⁓ a PR ⁓ and ⁓ we'll talk about how to use that in in a monorepo context. Like how can you store this same information ⁓ in ⁓ individual packages and have that sort of synchronized and pulled up to a root level code owners file that that is purely kind of a build artifact. Like you don't have to touch it because the source of truth lies within each package. And that gives this gives you a lot of control over like maybe you have, you know, an infrastructure team that's like

Sneha Mehra (00:04:36)  
responsible for some small set of packages and then you have other people that should be reviewing code in other places. And then finally, like and this even if you're well experienced with ⁓ with monorepos, I'm gonna walk through a couple kind of like standalone tools that help ⁓ detect and address common monorepo pitfalls, like the quality of each package's package JSON. Like are are you missing things in there? Does each have a home page like

Are you are you like pointing to the correct place in a repository like on GitHub where you could see the README for that package, not for the whole monorepo? ⁓ deduping versions of packages across your monorepo. Sync pack is a great, a great tool for that, right? If you have 16 versions of TypeScript across 20 packages in your monorepo, this this is what will help alert you to that ⁓ and potentially like in a in a quite automated way.

Where you can just use dash dash fix, it'll like consolidate you onto ⁓ a version that you can use across the entire workspace. ⁓ and we'll also work with a library called Kn KNIP, KNIP, which is great for ⁓ identifying like ⁓ needless exports from libraries that that ⁓ at least within the the monorepo, nobody is seems to be consuming. ⁓ And it it helps to identify unneeded packages. ⁓

in your package JSON, like unneeded dependencies in your package.json. And like why this is important for Monorepos? Well, we're gonna end up with a bunch of package JSONs. So that work, the like curating your list of dev dependencies and dependencies, it's sort of it's multiplicative now. And these tools help make sure that you're not ⁓ you're not having to go into each package in your in your project and you know personally you know pare it down where you can

see that there are things that like nobody seems to be importing within the package.

—-------------

8

Sneha Mehra (00:00:00)  
We need to now thread models into the UI code base. So if we look, like we got some problems. ⁓ see these like red squiggles happening? These files aren't in the UI package anymore. And we need some way of ⁓ stating that a dependency exists ⁓ where the UI package needs this new models package that we created. So we're gonna go ahead and take care of that. And we'll do this by going to the UI packages, ⁓ package JSON.

And we're gonna go into ⁓ dependencies.

And here we'll name our monorepo dependency that we depend on, seeds, models. ⁓ And the the specifier here is gonna be something interesting. This is a PNPm specific construct.

Sneha Mehra (00:00:57)  
So what we're saying here ⁓ is this is a package coming from the workspace, and I will tell I will tolerate ⁓ any version of that package. And what's going to happen is we will say, like the when when PNPM needs to resolve this dependency, it will first bias towards ⁓ identifying whether the local package meets this version specifier. In this case, we'll take anything. So that will always be met. But you could also put a number here.

And what that would mean. And like somebody in class was talking about how there are sometimes people do have monorepos with versioned inter monorepo dependencies. You could put something like that there, ⁓ and you would pull down, you'd go to NPM and you'd grab that ⁓ the appropriate version of the package here. But like I advise like go with this. This is where you get a lot of benefits from the monorepo. We have a question from chat.

Why add types and modules explicitly? Currently our code works in the same way. Interesting. So if you go up to ⁓

Sneha Mehra (00:02:06)  
Well let's let's get let's explore that idea. I'm gonna get the UI package working, and then I'm gonna remove those two entries ⁓ from the package.json of the models folder and let's see if things still work. I would be surprised if they do. It's possible, but ⁓ let's let's see. Like what what that would mean is ⁓ something is inferring that

That declaration file is the entry point for my package. Because if we go back to the models ⁓ package JSON, like

There's no information in here except in these in these places. Like, what's the entry point of this library? Is it main.js? Is it dist slash source.index.js? So typically you have to specify this. I would I would assume it's kind of dangerous to have a convention like that. Like, what if you had a main and an index? Like this kind of it's a bold assumption to make that if you just name something a particular way. But we'll ⁓ let's let's poke at that.

And we'll experiment and see if in fact these are necessary components. I as far as I know, they are absolutely necessary. Did you have a question? Yeah, so are we we're pointing to a local directory with ⁓ that workspace ⁓ string, basically, right? Yes. You're saying two things. One, this dependency can be found.

Inside this PNPM workspace, and I am willing to work with any version that you have for me. But like what that happens to mean is because there's this part of the resolution algorithm where like you first try to see if this requirement, which will accept anything, if that versioning requirement can be met within the local workspace, and of course it will, because you'll accept anything, then you'll end up linking.

Sneha Mehra (00:04:10)  
To that workspace copy ⁓ of that monorepo dependency. And so this is a Pnpm specific feature? This is. Other package managers have similar ⁓ concepts here, but this index here ⁓ is ⁓ a Pnpm construct. And is it crawling kind of the directory structure, looking for other package.json files to know where to find ⁓ that dependency? It does, but

Y yes, it does this part of PNPMI, but remember, it kinda already knows.

It it's like crawling a very small set of things, ⁓ right? It's crawling the folder that we've set our packages are in. And it's looking at each package and seeing, like, where is this? And and if you look if you're adventurous and you want to spelunk through here, like as part of setting up that package, there's some internal state that PNPM has where it already knows, like, here's the folder.

This is exactly where like it has already crawled that. It has already crawled that. So what's happening is as it evaluates each of the packages that it finds in that packageslash star folder, it's creating these this projects object. And then when it comes time to perform that resolution algorithm, it like absolutely knows what it has to work with in the latest the latest state it has about this monorepo. So like yeah, the crawling is happening, but it's

⁓

Sneha Mehra (00:05:49)  
It's happening more ahead of time than you think. Right? It's it's sort of the the state is already established and it's just sort of reading from that state. Alright. Well, let's see if this works. We touched a package.json, therefore we P N P I.

Sneha Mehra (00:06:10)  
Okay, so something interesting happened here. It says like it's we can see it's got three workspace projects. This is counting the root of the workspace because there exists a package.json there as well. ⁓ and let's take a look ⁓ at kind of what ended up happening here. So, first off, we're gonna need to update this import. That's the first thing that we're going to we're gonna need to do here. So this is gonna come from seeds slash models.

And look, that that resolved nicely. ⁓ And ⁓ let's run the build and let's see if there are other places where this needs to be updated. ⁓ and we want to run it in ⁓ let's just get out of our packages here.

to the root of the project.

PNPM build, so this is building everything. And note that it's building models before it builds the UI. This is not an accident. It understands there's a dependency between those. It understands that we have to have things in that dist folder of the models package in order for the UI compile to have a chance of working. ⁓ And this looks like it worked, but remember there's this separate type checking thing that Svelte has to do, and that's really where ⁓ we're gonna surface.

More more useful feedback here.

Sneha Mehra (00:07:36)  
And we get an error. I and I suspect it'd be what. ⁓ okay, we're building UI. We've got this formatting thing. I'm gonna run it again just so that I can see the full line. That's kind of what's screwing with me here. Hey, there we go. Real file names. So we've got our formatting TS. All right. So we're importing stuff that was from Seed Packet Model. This needs to come from. I'm gonna leave it on my clipboard, because I'm gonna use it over and over. Seeds slash models. Seed packet back. This is coming from seeds models.

Sneha Mehra (00:08:10)  
Seed packet state. This is going to come from Seeds Models. ⁓ And now this is actually ⁓ an additional ⁓ like import from the same place because it's not two separate modules, ⁓ one for the collection, one for the model. It's all being exported through that index.ts at the root of ⁓ our library now.

Sneha Mehra (00:08:37)  
And I know this needs a type here, like this is ⁓ this is just type information. I'm leaving it as as a little ⁓ Easter egg for us to find when we re-enable linting and getting it working across our monorepo. Let's try building again. Or checking again rather. All right, couple more. Seed packet state. ⁓ maybe we didn't save our file. no, we're getting it already.

We need that type. It's just encouraging us. ⁓ this helps with tree shaking, where we're trying to eliminate dead code, like unused dependencies as we're building for production. ⁓ if you sort of force yourself ⁓ to say, look, if we're just importing the interface, we import it this way. That lets build tools kind of walk through just the import statements and say, look, at runtime, there is no TypeScript. ⁓ we're just we can just like

avoid including this package entirely because we're only using it for type information in this case. And then our load data function, this is in our server. And again, same deal. Oop. Seeds, models, and there we go. Check once more.

And the build passes.

Sneha Mehra (00:09:57)  
So to loop back to that question about types, I'm gonna go back to our models package.json.

What happens if we get rid of these? Save. ⁓ And I'm going to, just to keep ourselves honest here.

Sneha Mehra (00:10:18)  
I'm gonna blow away the previous build.

Sneha Mehra (00:10:27)  
And look, like now the UI package can't resolve this. It's it's like it what it's saying is it can't find the module. But really what it's saying is ⁓ like sorry, a a a specific ⁓ a more specific statement here. It's like you're the module's there, but I have no idea. Well sorry, the the dependency is there, but it can't find.

Whatever we're trying to import here, like where is the specific JavaScript file or the DTS file to get types from? It can't find that. And so if we put it back, oops, wrong package JSON. If we go back to the models package.json and we add it again and hit save and then build.

We can see that everything goes back and we can type check and we can build again. So that's the purpose of those two fields. ⁓ This one is for type checking. This one is for actually executing code at runtime. Like, what's the entry point for the JavaScript itself? But they both ⁓ are you as the library author specifying like what is the entry point for this package. So is it kind of like

Without it, it'd be like saying, Hey, go to front end masters that's in this building. And you're like, Yeah, okay, front end masters exists in this building, but where? Yeah, it almost it'd be like saying, Go to front end masters and take the course. And you're like, Well, so I ⁓ I appear to have a front end masters and there are many courses, but like, what do you mean? Like which one? You know, there's a bunch of stuff in here. Like, where do I enter? And especially if you're like, this is almost like s remember, when when you're saying

I want to import something like this, like this is a package. ⁓ These are exported symbols that come from somewhere in that package, but without a module bridging that gap. Like when you when you say like there's almost an implicit something here.

Sneha Mehra (00:12:33)  
I think of it kind of like this. There's kind of something implicit there, but like what exactly do you mean? It's ⁓ what I'm sorry, it is it is ⁓ not implicit. But you're you're saying like I want something from this path, but it needs to know, like it needs to resolve, like once it says, all right, yeah, like there's a dependency here for sure, but like within that, like I've got a folder in node modules for this thing, but like what specifically?

And without that types field and without this module field, ⁓ that's where attempting to type check or attempting to run in the case of those enums where there's actual JavaScript, you know, in the compiled output, those will fail. Have you run into TypeScript LSP performance issues in large monorepos? And as the monorepo grows, what are some best practices to keep the LSP footprint lower? It's ⁓ so we're

I have some tips which we're gonna touch on, and that is the idea of using project references, ⁓ which creates like these TS build info files, ⁓ and it allows ⁓ type checking both in terms of the language server doing its job and build. It allows those to be done in a much more incremental way where TypeScript has more keen awareness of like these were the specific modules that you touched.

In the project. And when you rebuild, it's it has a lot more information that it can reuse instead of recompiling ⁓ every package from scratch. So that's one thing that helps here. But like bluntly, TypeScript, the the TypeScript compiler is really ⁓ complicated, is a heavyweight project, and this is part of why the TypeScript team is in the middle of a Go rewrite, where they're they're building both the compiler and probably like the

guts of the language server that ⁓ is the main thing people will use. ⁓ They're trying to use a language that that sort of lets them accomplish all of the analysis that they need to do in a much more efficient way that is more suitable for for large TypeScript modorepos. It is it is a real problem. Like it's, you know, Buntley, I mean, you know this if you if you're you know work on something big ⁓ you know for work.

Sneha Mehra (00:14:58)  
Type checking slows down when there's a lot of ⁓ a lot of weight. Especially at at ⁓ at Stripe, like our dashboard project is in a modern repo with a lot of TypeScript ⁓ and it's like 3.7 million lines of code or even more now. It's ⁓ it's challenging. So we would use like SWC for the build, but that doesn't really help with the language server. And so that's you know.

We're always just trying to tune it up more, use these project references, ⁓ et cetera. But the Go, you should take the Go rewrite ⁓ as ⁓ an indicator that there's not there there's kind of a ceiling in terms of like the compiler doing its job and being written in TypeScript itself. There's a limit to how performant it can get. And in particular, like it's not just the speed, but it's a memory issue as well. So if you're like

If your type checking is running slowly, just run run a little like a top or an H top and check out like how much state is being stored and how much memory it's consuming. And it's ⁓ it is ⁓ not trivial, gigabytes. Like this is like why you need to have a lot of RAM. Even if you're trying to run things like do do your build elsewhere, if your language server is running locally, it's it's a big ⁓ it's a lot of compute.

—--------------------

9

Sneha Mehra (00:00:00)  
In this next section, ⁓ what we're going to take on is breaking, like factoring another concern out of the UI, ⁓ the UI package. We have a server. We have that express server and the data that it needs in order to ⁓ produce an API response when you ask for it. And we're going to factor that out of the UI project so that we'll end up with three packages. At the end of this, we're gonna have the models package.

And both the server package and the UI package will depend on models. So we will now have a nice ⁓ a nice setup where we can examine like how does a shared dependency work? And we can start to poke at, you know, ⁓ poke at different tasks and how they can operate while having well being aware of that dependency. So ⁓ step one. ⁓

Let me give us a little clean folder structure here. So we're gonna make a server folder. Make dir dash p packages, ⁓ server, source, ⁓ and tests.

There we go. ⁓ and step two, we're going to move the we're gonna go into the UI project and we're gonna move the data folder into server.

Sneha Mehra (00:01:26)  
We're going to grab

All of the j the TypeScript modules in source server. We're gonna grab all those and we're gonna bring those into source up here.

Sneha Mehra (00:01:43)  
And I'm not gonna update imports. We're gonna like ⁓ we'll we'll figure that out. And then finally, tests. ⁓ There is a server test. And I'm gonna move that into server slash tests.

Sneha Mehra (00:01:59)  
So we've moved over all the source code and the tests and the data. We can delete these two server folders if you want, just to be nice and clean, but it won't matter if you if you leave them alone. ⁓ Cool. All right, we need a very basic ⁓ package.json for our new ⁓ server project. And I'm just gonna borrow. I'm gonna grab actually all three of these things. Both TS configs and the package.json from

our models project because a lot of it's gonna be the same. These essentially are just like they're both sort of node libraries ⁓ that don't involve any fancy UI of some sort. And ⁓ we should be able to to sort of make all that work with some minor adjustments. So I'm gonna oop sorry I'm gonna copy them. ⁓ And we're gonna make some adjustments. So here we'll say this is server. ⁓ We'll keep our convention here, although it

Bluntly it matters less because this server kind of has an NPM script to start. If it exported a function that started the server, we would care about downstream dependencies. But this is a leaf level dependency.

maybe we delete it just in case to prevent someone from accidentally importing stuff. ⁓ All right, looking at tests. ⁓ This all still applies, linting still applies, build still applies, this is all exactly what we want here. ⁓ And ⁓ it's still like we have excess dependencies here. I'm leaving them because we're we have a we have a tool for that and it'll help us clean all that up. But I think this is a good starting point here. Now we depend on ⁓ models.

Right. So if we look in our code.

Sneha Mehra (00:03:46)  
It's not there, it's in load data for sure. This is now broken. We just had it working and it's broken now. And the reason is we haven't described that inter-workspace dependency where server depends on the models package. So just as before, we have to go and take care of that. And it's not a dev dependency, it is a dependency.

Sneha Mehra (00:04:15)  
Again the P NPM thing, workspace, star. Great.

⁓ And ⁓

Sneha Mehra (00:04:28)  
Let's see if this works. So I'm gonna go into packages. Server.

Sneha Mehra (00:04:37)  
⁓ wait, first we touched a package JSON, we gotta run PNPMI.

Sneha Mehra (00:04:46)  
Okay, if we look at our node modules folder, we've got some nice simlinking here.

Sneha Mehra (00:04:55)  
Great. Hey look at there's seeds. ⁓ So it's showing up. It's actually, interestingly, it's gonna be if we look, take a peek in our PNPm state here, just going for seeds, whatever that is. S. ⁓

Sneha Mehra (00:05:16)  
it's not here. There's probably some other mechanism where that was happening. I was curious, like is it is it going to actually simlink it in this node modules folder? But importantly, like here, if we were to go into source.

Sneha Mehra (00:05:33)  
And do this. ⁓

Sneha Mehra (00:05:38)  
Like, it's the same file. So it's simlinking right into our workspace. Same file but different undo redo stack, apparently.

Great, so we did PNPMI, PNPM build.

Sneha Mehra (00:05:58)  
And ⁓ we have ⁓ a a build output in dist. ⁓ let's try to start the start the server and taking a look at our npm tasks. ⁓ we have a dev thing here, but this is more about like watch the build and do a rebuild when files change. Like dev means something different in the context of our server. What we want

is to use ⁓ TSX, which if you haven't if you use TS Node and haven't checked out TSX yet, give give it a look. It's does same thing, better, more support for ⁓ modern module systems. TS Node, ⁓ you can use experimental loaders for things, but TSX is pretty, pretty sweet. ⁓ There is a task that was defined in our original project that takes care of starting our server. ⁓ And it's in our UI projects package JSON. Here it is.

So we're saying TSX watch preserve watch output and it's it's just running the TypeScript file ⁓ natively. So we're not we're not worried about like an incremental build happening, like all of that's taken care of for us already. So I'm gonna grab this line here. It doesn't it doesn't really mean anything in the context of ⁓ the UI project anymore. And granted, this task depends on it, so we've sort of broken our dev.

our holistic PNPM dev experience. We're gonna get around to fixing that. So ⁓ what what we are gonna be able to do is start this server up and see that it listens on localhost three thousand and we can ⁓ we can hit it and we will work our way towards actually being able to see data.

Sneha Mehra (00:07:48)  
Okay. Error, file not found. Well, what does that mean? File. Server source server index.ts. So that clearly means this, right? Like it's just source index.ts now. This was a server subfolder of source in the UI project. Now it's just indexed index.ts. ⁓ Try again. And we're going to click on this link.

And we're gonna say slash API seeds ⁓ and we get our seed seed data there. Now we benefited from something nice here, and that is ⁓ we dragged our data folder in from from ⁓ we have placed our data folder in a spot where the relative path to that data folder is unaltered between the load data.

file here. So like if you put that data folder in a different place, this line of code here, the data file path, is what to troubleshoot if you didn't see that seed data coming through. But now, now we have a working UI. We have a working ⁓ models package. We can see the dependencies between all of them. And if we back out

Sneha Mehra (00:09:11)  
We can see that we ca we have like a holistic build process that first starts in models and then builds the UI and then builds server. And this is PNPM having an awareness of the ⁓ sort of the the the d the DAG, right? The directed graph of our dependencies. And it knows to build the things everything depends on first and then works its way towards the leaf level dependencies.

Sneha Mehra (00:09:39)  
So we should be able now to run P NPM build, P NPM.

⁓ check PN PNPM test.

Sneha Mehra (00:09:57)  
A lot of dopamine right there. Lots of green check marks. Everything looks good.

—-----------------------------

14

Sneha Mehra (00:00:00)  
We're going to use a tool called KNIP, ⁓ KNIP, ⁓ to identify exports ⁓ from our modules that don't appear to be used from within our monorepo ⁓ and dependencies that we don't appear to be importing from. This is a really useful tool in general for any JavaScript project because ultimately, like that dependencies list in your package.json starts to get big, your dev dependencies and your dependencies, ⁓ and

you know, it's t it's tough to know what of those things is actually still being used. It sort of ⁓ ends up being hard to hard to garden. Is this called tree shaking? This is this is not called sh well, ⁓ it is conceptually similar to tree shaking. What most people mean when they say tree shaking, it is the concept ⁓ during a build process ⁓ of ⁓ identifying dead coat.

Or like unused code and eliminating it from the build. So ⁓ for example, ⁓ if ⁓ if you were ⁓ Lodash is a great example. You can consume Lodash as one big library if you want. You can say, I want to install Lodash. But they also ⁓ let you ⁓ yeah, there you go. So here are a bunch of NPM packages. Like you just need Lodash dot memoise.

you can install that as an individual package. And so if you were to use this, what this lets you do is ⁓ well, first off, you can install only the modules you need, sorry, only the packages you need, but this also shakes away any modules within those packages that you don't happen to be using. Now what we're doing is related to that in that we're trying to prune unused things away. like dependencies that are in our package JSON

For our various Monorepo packages where we see no evidence, those are being used within that respective package. But we're doing this sort of to benefit ⁓ install times and build times and just to take away stuff that that's ⁓ somebody's factored some code away, but they forgot to remove the library. Like, yeah, if you have tree shaking in place, it would get it would take care of that eventually. But like ultimately it's good to clean it up ⁓ as well. So that's that's what KNIP does from

Sneha Mehra (00:02:28)  
⁓ the the dependency side of things and then unused exports as well. Like you're exporting 12 things from a module, but we can only see three of them being imported from anywhere. And this will help you make sure you're not overexposing your code to the outside world. We had some interesting conversations on about like how do you how do you make sure that ⁓ in a monorevo, how do you make sure you can evolve things that are deep in the dependency graph.

And part of that is just being very deliberate about the API surface you expose. If you export the whole world, people have access to your internals, they start using those things. Well, yeah, evolving that's gonna be really tricky. And this helps you make sure you're you're more deliberately exporting things when they're actually being used by something. And it's not just sitting there waiting for somebody to to grab onto. So with that, ⁓ let's jump in.

So the first thing we'll do is we're gonna install ⁓ nip PNPM install dev dependency knip. And I'm in the root of my monorepo. This is another ⁓ like workspace level tool. ⁓ so there ⁓

Like this is most appropriate to store it at the workspace level. Now we're going to install ⁓ knip and then to do that we have to run pnpm i dash d knip.

Sneha Mehra (00:04:00)  
And ⁓ now that that install is completed, we want to create ⁓ a config file. So this is gonna be in the root of your project, and we're gonna call it ⁓ knip.json. And you'll know you got the right name because there's a little, at least in this VS Code icon pack, ⁓ apparently they have a cute little icon. All right, workspaces. You gotta tell ⁓ tell this tool where things are, and we're gonna say packages.

slash star ⁓ and what we next need to describe is like what what is the source code for each package like what's the the pattern to be ⁓ that you could use to find source code versus ⁓ you know what's the entry point for each package so here we'll say entry and that's gonna be it's pretty consistent for us we've got source slash index well it's really index or main

⁓ And the reason is ⁓ in our server and our ⁓ models packages, source slash index is what we're using. Main is what's used in the UI package. And like normally you don't care about the entry point for something like a web UI, but but the point of what this is trying to do is it's going to start at the entry point and then it's going to walk all of the imports ⁓ to get an understanding of like what are those ⁓ modules importing.

And it's it's doing this to figure out like, are you actually using all of the things that are in your package.json. So next we need to do the second property, which is project. ⁓ And we're gonna say source. ⁓ yes, this is all relative to a package to be clear. ⁓ So it's gonna be source anything dot ts tsxvelt now.

You don't really need TSX, it's not in this project, but you get the idea. Whatever extensions ⁓ are part of the analysis that should happen here.

Sneha Mehra (00:06:06)  
⁓ okay, we can leave it at that. If you're following in the workspace notes, you you'll see there's some other stuff that we'll add there, but I wanna hit the problem that necessitates adding that second field. So let's run PNPM.

Sneha Mehra (00:06:23)  
Nip.

And what's it gonna spit out? A bunch of stuff. So it's telling us a lot of good information here. We apparently have unused dependencies. ⁓ Looks like we left some server stuff in our UI package.json in the dependencies object. Similarly, in models, like here's the Express ⁓ dependency again. I mean, after all, we were just copying and pasting those objects. We were being kind of sloppy about it, but it's okay. We have a tool now.

They can help us prune those things away. You get the same thing for dev dependencies. Now, important to realize, there are some things here that we want to keep, which do not get imported by something else. In fact, these two things are CLIs. This is something that's used for a linting task. These are ⁓ well, these are actually, ⁓ sorry, we could get rid of these because ESLint is ⁓ at the workspace root now.

We don't need the types for these packages because we're getting rid of the packages. You should be able to see we're not being asked to eliminate those packages in this server group. So this tool even has knowledge over these like this ambient type information convention, like the at types packages. It will see that, like, no, you're not directly importing from at type slash express, but it is in use and it's intended to sort of layer on top of the express package. ⁓

⁓ let's let's go around and play whack-a-mole and and ⁓ address some of these. This is gonna be the easiest group for now. So let's go into our ⁓ UI package.json.

Sneha Mehra (00:08:02)  
And get rid of bluntly everything except our models dependency. So that's that was all server stuff. A YAML parser, a logger, the Express HTTP server, and course. Get rid of that. And let's go into models package.json. ⁓ And similarly, like we don't need any of that. ⁓ And let's see where we're at. Run that command again.

And now we're down to unused dev dependencies. So these two I want to treat differently. ⁓ We'll go back to our ⁓ config file here and we're going to add a new top level field called ignore dependencies.

And it's an array and we can put ⁓

Mini package, CLI, and sync pack in there. And if we run the command again, we'll see we're not being yelled about yelled at about those two things anymore. So this is this is where your dependencies that are tooling or CLIs or you know ⁓ I don't know, some some ⁓ plugin that you use for testing infrastructure. Like this this is a good example.

Like test coverage, let's add that. You know, that's that's not something that that ⁓ this tool can see. Or it it both does not have special awareness of like how this plays into our test command, ⁓ nor can it see an import.

Sneha Mehra (00:09:44)  
So we'll ignore that one. Any other ones? TypeScript ESLint. Well that should be pulled up to the top of the repo. So it genuinely does not belong in there. There's coverage V8. ⁓ Great. All right. ⁓ let's do some more pruning. Get rid of some of this. So

Sorry, let me make sure I'm on my most recent invocation of the command. There we go. So we took care of a couple of those. ⁓ let's go in our models package.json, giving us nice links, by the way. I love this. Click it, and it takes you to a row and a column of exactly what you need to eliminate. So we get rid of cores. We get rid of express. ⁓ we need the node types. ⁓ This is fine.

This we already said we're ignoring. ⁓ the UI it's not complaining about. Concurrently, we don't need in here. Prettier, we don't need in here. TSX we don't need. TypeScript ESLint, yep, we don't need that either. Really slimming it down.

So there that's what we're left with. This is much more reasonable for just like a very plain ⁓ TypeScript library. It's it's some testing stuff. In fact, ⁓ ESLint, do we even need that?

Nope, I don't think we do.

Sneha Mehra (00:11:12)  
Great. So running it one more time.

Sneha Mehra (00:11:18)  
Okay, so we're out of the models ⁓ section. Now let's clean up our server package.json. So we've got ⁓ eslint.js ⁓ concurrently prettier typescript eslint ⁓ and save. And then in our UI, ⁓ this testing library Svelte seems important.

I'm gonna chalk that one up as like let's let's keep it in here. Let's add it to our ignore list.

Sneha Mehra (00:11:56)  
⁓ And the TS config svelt, let's let's say that's the same thing. that ⁓ where that's being used in our UI TS config. ⁓ sorry.

Sneha Mehra (00:12:14)  
This one.

There's a TS config npm scope now that contains like very framework-specific settings for ⁓ you know popular configurations and stuff. So, but this is not part of what KNIP is trying to analyze. So it's not aware of that. And then ⁓ going back to our UI thing, just these cores ⁓ and express ⁓ things here. And let's see where we're at. Getting much closer. All right, now we've got.

Unresolved imports. ⁓ Great, it's finding more things. ⁓ This is ⁓ just removing the word models. Because that's within the same package. And here, ⁓ this one doesn't need to be referring to this package. Seeds models. Great. So we will have taken care of those. And that represents kind of like the next class of errors. Now we're down to unused exports.

So we've got server config. And if we look here, this is just like a configuration object that holds a port and a logger, and it's for our express server. Now, here's here's a little bit of a gotcha. I personally I wish these two families of errors were sorted differently. What it's really telling us here is we've exported this thing twice. Like ⁓ when it says unused export, I want you to remember that's that's an unused.

like export site. The ⁓ what I'm trying to disambiguate between is server config is a class. It is both exported as the default export of this module and it's exported as a named export here. This is a understandable thing to do. ⁓ And ⁓ there are things that use this ⁓ like in our ⁓ app TS

Sneha Mehra (00:14:15)  
You can see we do import it, but we're importing it as the default export, right? And so it's telling us ⁓ it's it's really this that's unused. So ⁓ I'm gonna decide, you know what? I want the named export to be the thing that I ⁓ I preserve here. So I'm gonna do this. ⁓ I'll go back to this server config thing, and I'll just delete the default export.

Sneha Mehra (00:14:44)  
And let's run this again.

Great. And so that went away. So really like this is a trap to sort of process these first. Because sit similarly here, load data pipe default, that's saying this thing, there's a duplicate here, but you're also seeing this lineup here, which has to do with the duplicate, right? So similarly, I'm gonna bias towards ⁓ named exports. ⁓ And where would load data be? ⁓

Well, what we could just say find references.

Sneha Mehra (00:15:22)  
⁓ Only in this routes package. And good. We have another thing to fix here. Load data and server config. So both of these will consume as named exports. Everything lines up. We get rid of the default export. We can run more one more time.

Great. And we're whittling this down and we're whittling this down. Now, what do I love about this tool? This would have just been incredibly difficult information to track down. Just think about all of the manual work that you'd go through, like trying to scan through. You'd like grab each dependency and like scan through import paths, and you'd have to worry about you'd have to build this tool to do this job right, especially in a large monorepo. ⁓ it would be really, really challenging. All right.

Get seed packet ID. ⁓ I happen to know, like I I left this in here ⁓ so that it was here for us to delete. It turns out nobody needs this function. Somebody built this ⁓ and it's just it's just like throwaway code. Maybe at some time someone was using it, but we can just entirely get rid of it now. Nobody was even depending on it. And then these two as well. Like we have a bunch of formatting things, which are for the backs of the seed packet if you click click and turn them over, but

We're not displaying light preferences and we're not displaying water needs. And so we can get rid of those as well. And we should be able to run this. And it passes. So in summary, NIP is good for ⁓ two things. Eliminating things in your dependencies, dev dependencies, peer dependencies of your package.json that do not appear to be used, and giving you an automated tool that lets you continually scan for these things. You could incorporate it into your build process as well.

Somebody introduces a new dependency, they'd better use it, or state that it should explicitly be retained. And then similarly, this makes sure that when you're exporting things, you're doing so deliberately. You're not just saying, I needed, ⁓ you know, maybe it's like I needed to test this thing and so I had to export it or something. I mean, maybe that's a good reason to export. We'll talk about how you can do that safely later. ⁓ but you know, this this lets you find things that are just simply dead.

Sneha Mehra (00:17:38)  
Like dead code. And tree shaking would not have helped with this part, by the way. Tree shaking is always done at a module by module level. So so like these things here that nobody was using, they ⁓ it would have been very difficult to detect. And this absolutely would have been code that you're sending to production that nothing appears to use. So check this tool out. Try it. I bet you'll find some stuff.

It's just unused in your project. Monorepo or not. And the but like monorepo, especially because of the contracts you have between your monorepo packages. Like this this problem, as bad as it as it is with one JavaScript or TypeScript project, it is that much worse with a monorepo because you're you know, you have all of these different ⁓ relationships between packages.

—-----------------------  
10

Sneha Mehra (00:00:00)  
In this next section, we're going to take a close look ⁓ at several tools that we can use ⁓ to work with our package manifests, establish consistency across our Monorepo, ⁓ detect varied use of dependencies, meaning you have six versions of React across your monorepo. Like how do you deal with that? As well as ⁓ trimming away ⁓ unused dependencies and exports. So

A bunch of things that sort of relate to working with your package manifests and your external dependency graph. ⁓ And these are some really nice, ⁓ really nice kind of standalone projects that you could layer on top of PNPM or an NPM workspace or a yarn workspace. ⁓ they all they all sort of work with any kind of ⁓ any monorepo that sort of aligns with those conventions. So, ⁓

First, we're gonna take a look at ⁓ many package. And the the main purpose we're going to use many package for is linting ⁓ our ⁓ package manifests. So this is if you wanna have a really opinionated formatting of your package.json, like alphabetizing your dependencies, for example. ⁓ this is this is kind of like prettier for package.json, if you wanna think about it that way. Not just for formatting JSON and

getting your quotes right and your your indenting, but more ⁓ having a standardized way of formatting things so that as people change, go in and make changes to package JSON, like you're you're keeping the increment of change ⁓ very small, like nice readable diffs instead of seeing that you know somebody moved six lines around or something like that. Let's begin by installing the the mini package CLI. So we'll go back to our environment, our dev environment, ⁓ and

Well install this package and I'm in the workspace, very important here. Like this is where you want your workspace level tooling.

Sneha Mehra (00:02:10)  
So we've installed at many package CLI. By the way, there's an NPM scope. I know you've been working with packages like this before, but like this is an example of something that has a programmatic interface for like consuming this as a library as well as being able to use a CLI that they provide. The next step we're gonna take is we're going to run Pnpm MiniPackage check. And by the way, if you're if you're still using npm and you have to use like npm run or npx, like this.

I this is the thing I really enjoy about PNPM, the fact that like Yarn does this as well, where if you have something that's in your like node modules bin folder, something here, you can just invoke it, ⁓ invoke it directly. So menu package has a check command. Remember, this is different from the check that we've been running that relates to Svelte and type checking. ⁓ And ⁓ let's see what it finds. Workspace is valid. That's interesting.

Am I the right spot? ⁓ I thought I left some some problems here. Well, let's let's look at the documentation and let's create a problem for ourselves and see if we can we can detect that. Maybe I fixed some things on the fly.

Or fix some things as we were going. So ⁓ many package check runs checks against your monorepo. ⁓ And ⁓ an example of this ⁓ is ⁓ having ⁓ you know peer dependencies ⁓ that are not dev dependencies ⁓ and having their range specifiers meet. So we we don't have any of those. ⁓ root has production dependencies. We could certainly do that. So let's go to our package JSON here.

And let's say many packaged CLI is a production dependency.

Sneha Mehra (00:04:01)  
This will tell us if it's working. Hey, great. The root package.json contains dependencies. This is disallowed. Dependencies versus dev dependencies in a private package does they don't affect anything and it creates confusion. This is this is important, right? And it's not like.

I want to be clear, private package in this case, it literally, it it kind of means two things. One is literally this, ⁓ which is, you know, is this published to NPM or not? But the other is, is this a leaf level dependency or not? So for example, in in this package here, like if we were ⁓ let's say

Let's say we had a step where we built built things, ⁓ and then we want to slim down the set of dependencies we need to run the UI in in production. Like it it does matter that we have some runtime dependencies here. So I would argue this is a little bit of a imprecise way of describing this. It's not it's not about private packages. It's about being a leaf-level dependency versus being a

dependency that has dependencies, right? Like if if there ⁓ if we need Express as a transitive dependency, despite this being a private package, like if we were to move this up, all of these and say, these are all just needed for dev, that's gonna be a problem, right? Because there remember there are build commands with with each of these, you know, each of your build tools, they have a mode where you could say, ⁓ only install my prod dependencies. Sorry, not build tools, your package managers have

Like only install prod stuff. So that's an example of something that ⁓ that it can detect. And if you were to say dash dash fix, ⁓ but I think it's not dash dash fix, it's fix. ⁓ And you can see that it it went through and it actually did this rename for us. So the kinds of things that it can detect, ⁓ they they have a you know bunch of different rules here, invalid package names. This this is about like you have chosen

Sneha Mehra (00:06:10)  
Gotten too cute with your Monorepo package name, you've put an emoji in there, like a double width UTF-16 character, and it's just not gonna npm will be ⁓ unhappy with that. ⁓ this is the one. ⁓ This is the one that I was trying to run us into and I forgot a step. So let's do this. I'm gonna have my Monorepos v2 URL here. ⁓ and I wanna go back to my code and I'm gonna say in my root package.json.

I'm gonna have repository and then ⁓ URL.

Sneha Mehra (00:06:49)  
And now let's run this. Ooh, wait, is it just this?

Sneha Mehra (00:06:59)  
Yep, there it goes. So ⁓ this is a nice thing that many packages going to detect for us. Like in an ideal world, you could say I have one git repo for my monorepo, but ultimately, like when you publish to npm, if we go to npm.js, oops, feature belling.

Sneha Mehra (00:07:24)  
and let's go to Babel. ⁓

Sneha Mehra (00:07:30)  
Babel preset TypeScript. ⁓ You kind of want to have like a home page or or something that like, well, maybe this is a bad example. Here's the point. There's a repository URL on each package you publish. And one option for doing this, and I can switch to a branch that I have here, ⁓ where I've got all of the work that we've already done today, or that we will get through today.

⁓ if I go to packages ⁓ and then UI, like you could have your own README for each package. And what you'd want to do is say, All right, well, I have a a repo for the whole mono repo, and I have a URL there. And now what many packages trying to get us to do is say, you know, there's a UI path here that we can go and add in each path.

packages, package JSON, not just for the whole repo. So if we would go up here, repository, and paste this in and grab that, and we're just gonna change the name, the last little token in the string. So we'll add this to server as well. I know I'm going a little fast here, but

Server and then models.

Sneha Mehra (00:09:00)  
Oops. I think I copied the whole line.

Sneha Mehra (00:09:05)  
Great. And now if we run this check again, it's gonna say oop. ⁓ It has a repository field of this when it should be.

Sneha Mehra (00:09:23)  
You have the branch name and the you have tree slash subs instead of tree. ⁓ yes. ⁓ Like this is even more sophisticated than I hoped, right? Like you you wanna have ⁓ this is steering me towards a mistake that I could have made here. So it's it's pretty sophisticated. And it it's a small number of checks that it that many package enforces, but ⁓ you know, generally generally a good thing to use. Now, there are other things that this does.

Which is ⁓ it it like allows you to you know run scripts ⁓ for all packages within your monorepo, but bluntly like I would use PNPM for this. And ⁓ in future steps, we'll use Lerna. ⁓ because this is not going to be as aware of the dependency graph. There like you'll see that a lot of monorepo tools have have this ⁓ feature where like you're

You're gonna install a bunch of these tools and you'll have 12 different ways you can like run a task in each package. But what you wanna lean towards is the ones that let you do interesting things. And we're gonna see later in the course, like Learna and NX is where that gets really sophisticated and what you what you should lean on. But this still has as a unique feature of linting those package JSONs and giving you a very ⁓

You know, some some helpful feedback where you can even see like the stuff I'm doing, it's busting me on because I'm I'm pasting in strings that are not quite right.

—-----------------------------------

15

Sneha Mehra (00:00:00)  
In this next section, we're going to get our dev script ⁓ working again. So if you remember at the beginning of the course, we had this nice project level script when package and project were the same thing at the beginning of the course. We could run ⁓ PNPM dev and our back end would start up, our front end would start up, ⁓ and importantly, ⁓ any code that we touched would trigger a rebuild of some sort.

And so we could just sort of like keep the app running, both the front end and the back end, and then go and edit our code, and we could see the changes that we made take effect. This is part of what's awesome about being ⁓ a web developer, either on the front or the back end. Like often you get that fairly immediate feedback loop that you don't you don't quite get if you're, you know, you have a big ⁓ meaty compile step that that can't be done incrementally. Let's let's get started.

The first thing we're gonna need to do is go to our UI's package.json.

And this needs some work, right? ⁓ really what we have to think about here is in the context of our UI, what does dev mean? Now, because like this is a UI scope to package JSON at this point, I would argue this is the real dev for the UI. Just spinning up bites, ⁓ you know, dev server mode.

And we're gonna grab this as the starting point for the dev script we will hoist up to the top of the workspace.

Sneha Mehra (00:01:38)  
Open up our workspace package JSON ⁓ and we're gonna paste it. And it needs some alteration here. We we can get rid of this thing here. Actually, let's call this dev dev one for now. I want to show you why this is not going to work.

If we run this, because remember, like Pnvm can run a command in each package in the repo, ⁓ and it can produce colorized output. But here's what we get: packages slash models, ⁓ it's running the dev script, ⁓ and Pnvm showing this. But Pnpm run is ⁓ it's good for tasks that exit.

Right now, what we're running into ⁓ is there are dependencies here. Like as far as PnPM is concerned, dev is some arbitrary thing. ⁓ And it knows that models is the first thing it should try this on, because it's the it's the lowest level thing in the dependency graph right there. It has to happen first. And so we're kind of stuck here. And if you s if I control C this, you're gonna see some evidence that, okay.

Great, now now it tries the server and now it tries the UI. So so what you're seeing is, ⁓ well, finally I exited this ⁓ with a control C and now it's starting up the other things. So it's falling short here. ⁓ and this is why I don't use dev for this. ⁓ I like to use this is a really nicely scoped NPM package, good at one thing.

And that's just running multiple commands concurrently. And it's called concurrently. So we're going to modify this. First, we're gonna have like you've got this first set of things. These are names of different tasks that will be printed to the console. It's a prefix. So we could say server models client. These are colors ⁓ that are used to sort of color code this is coming from the server, this is coming from the client.

Sneha Mehra (00:03:50)  
Remember, this is all this is going to be like an interleaved output from ⁓ many different things that are running. And so this is to stay sane and to know like which output is coming from which of these tasks, color coding is pretty, pretty nice. And then ⁓ we have individual tasks that'll run ⁓ the server and the client. But we kind of want to do it differently. Like these are still left over from the beginning of the course where we had dev server and dev client.

But we can do this differently. We're gonna say ⁓ PNPM.

Sneha Mehra (00:04:28)  
Dash dash filter. And this is a way to say ⁓ this should run ⁓ on one package.

So the and important to get them in the right order. Ordering matters. Notice we've got three things, three things, and now we need three things. So we want server, models, client. So here we are gonna have server.

Bottles.

Sneha Mehra (00:05:00)  
And client.

Sneha Mehra (00:05:08)  
This is definitely the point, like how many columns wide are we at this point? Like we're already at the point where I'd say write a shell script. We're gonna do that in a sec, because good God. ⁓ but let's see this work.

Sneha Mehra (00:05:24)  
⁓ what's happening here? Concurrently, command not found. We gotta install it.

Sneha Mehra (00:05:37)  
Let's try it again.

Nice. Okay. ⁓ I'm noticing some things. Well, first, models doesn't have a color code. Pink might not be the right, might not be available. There we go. We've got servers, model, and client. So you see how important it is to have that nice little prefix ⁓ so you can easily see what's happening. But look, here's our back end. ⁓

Interesting. No projects match the filters in that. Did we call it client instead of UI? Of course I did. Because I was saying the word client ⁓ while I was typing. And there we go. There's Vite actually starting up. So in the end, oop.

Interesting. We'll figure that out. But that clearly clearly the client's running. ⁓

Sneha Mehra (00:06:38)  
Clearly the API is running. Let me see if that's a super easy thing to fix. ⁓

Well race condition or something. Maybe that error popped up because the client started first before the server was up or something like that. Anyway, ⁓ here we are. Here's real data coming through from that data file that's being read, ⁓ all coming from the JSON. So now we're back to ⁓ but here's here's your server logs coming in. If I go back to a Svelte component, like the seed packet. ⁓ interesting. So first, like

⁓ on the website, and I'm just showing you that you could do it either way. You can always, whenever you're expressing filters like this, you can refer to sort of the unambiguous package name, or you can refer to it by the folder name in that package's folder. So this should work similarly well. Yep. There it goes. Yeah, I'm s I ⁓ to me this looks like it's ⁓ at least from this standpoint it's working fine. Hmm.

What's happening here? Did I delete something that was important from package JSON? Is Svelte? Svelte's in here.

⁓ restart language server.

Sneha Mehra (00:08:07)  
That was it. Some stale language server state. So sorry, the thing I was trying to prove here was ⁓ if we go back ⁓ and we find, I don't know, some interesting text. That's a lot of gradient stuff. ⁓ that's an icon. look, there's some text being shown here.

I hit save. There we go. I've ruined every seed packet with some text. Everything's like, so this is happening immediately. We know that the dev server is working there on the UI. We know it's working on the server. ⁓ because I could go ⁓ here. ⁓

And change the startup. Sorry, that's an index.ts. You don't have to follow along with me here. Let's delete the word port because we're showing a full domain here, right? So ⁓ gonna scroll to the bottom so we can all see it. Save. ⁓ And look, we just bumped the server, so changes happen instantaneously there. The tricky bit ⁓ is of course models. And if we went here ⁓ and said version.

Sneha Mehra (00:09:18)  
And hit save. You can see there's a file change detection, incremental compile. That means that in the disk folder, we're getting, you know, getting a new, a nice new build. So now we're back to this great ⁓ productive dev script. I can touch any file, those changes take effect immediately. I have a working client server set up, despite it all being in a monorebo. Now, ⁓ the only downside here, just I'll I'll be blunt, and ⁓ you can script your way out of this.

Is this a place where you do have to keep up to date with all of the different packages that are in your repo? But ⁓ there are ways to get around this with other tools. If you really wanted to use concurrently for this, what I would do is I would just make a script that lists the packages and then builds the concurrent concurrently shell script, which I told you I would do, and I'm gonna do it.

Sneha Mehra (00:10:18)  
'Cause I think it's a little crazy to have something that that complicated in an NPM script. So I'm gonna in the root of my f project, make a scripts folder.

Sneha Mehra (00:10:31)  
Make a shell script there.

Sneha Mehra (00:10:36)  
did I get that wrong?

Sneha Mehra (00:10:41)  
⁓ other way.

Sneha Mehra (00:10:45)  
⁓ And get these lined up real nice just so we can see what's going on.

Sneha Mehra (00:11:02)  
And I think we can get rid of these escaping things too.

Sneha Mehra (00:11:13)  
Continue, continue, continue. So again, just like three commands running at the same time, color coded.

⁓ schmudd.

Sneha Mehra (00:11:28)  
This if you're still learning your l Unix commands, we're just making this an executable script.

Sneha Mehra (00:11:35)  
⁓ one more thing. ⁓ We ⁓ have to say PNPM DLX, which if you're an NPM user, this is PNPM's version of NPX. ⁓

Download, execute, DLX.

Sneha Mehra (00:11:58)  
And there we go. And that first little blurb you saw here, this is like this is the downloading on the fly, even if you didn't have concurrently. Right? And so now we can also just do PNPM dev and it calls the script.

Sneha Mehra (00:12:15)  
So now we have a working dev script.

—-----------------------

13

Sneha Mehra (00:00:00)  
Next, let's get linting working across our project. So the current state ⁓ of the world is we have this ESLint config, but it is within the UI package. Like this is we haven't touched this since we moved things around. ⁓ And if we went into the UI.

lint still works. ⁓ Does it still work? interesting. I need to add my ⁓ vite config to ⁓ my TS TS config here. You know what? This will get shaken out as we as we make progress here. I can explain what's going on though. ⁓ so our current setup, this this has to do with like TypeScript and Linting more than

More than monorepos. But ⁓ what what we have going on here is we're saying I'm using ESLint and I have some type-aware linting rules. This this these are linting rules that ⁓ are ⁓ alerting the developer to problems ⁓ that

Like detection of those problems involves ⁓ using type information itself, right? ⁓ And the way you set this up is you have to enable the project service, ⁓ and you have to point to a a tsconfig file or a tsconfig root directory. And right now, that's gonna be this file here. ⁓ And let's see. We've got the post CSS config stuff.

Sneha Mehra (00:01:42)  
Huh. Kind of expected that we would not see these files here. Maybe. ⁓ You know what? We're gonna keep pushing forward. ⁓ And let's let's see if this still shakes out when we move things around. But I'll tell you that the thing I immediately look for here is like based on these inner error messages, ⁓ you you like if you've left this file out, if it's not being type checked, it's not part of your config, you either have to do something like this.

Sneha Mehra (00:02:16)  
Right, and we could put put a bunch of things in here like

Sneha Mehra (00:02:23)  
Right? You could build up an array and say, like, these are just things that won't get linted for me. ⁓ Or you have to ⁓ be referring to a TypeScript project that includes type checking on those files. Otherwise, TypeScript does not know, or ESLint doesn't know what to do with them because it's ultimately producing kind of like the AST that represents the code that will be executed and an equivalent data structure that represents the types on that code. And if there's no types,

Like it can't do its job. So ⁓ let's let's do some refactoring and see if we we end up fixing this ⁓ in part as part of that process. So first ⁓ let's hoist some of our dependencies up to the workspace level that relate to linting. And so what's that gonna be? Certainly TypeScript ESLint is gonna be one of them. ⁓ we'll have ⁓ ESLint, there it is. So we'll grab that too.

⁓ And I think that might be it. Just these two?

Maybe TypeScript itself. We'll grab that.

So these three lines. ⁓ And we're gonna bring those up to a root level package.json as dev dependencies. Save. We touched a package.json. We must PNPMI.

Sneha Mehra (00:03:53)  
Did I save that? Yep. Great. Okay, now ⁓ we should be able to do.

sorry, we need to move this lint the the lint config up to the root of our project. So we're gonna move that up.

And whenever I do this, I think we might be in good shape here, but I I do want to check that paths look correct. ⁓ interesting it can't find this module. Let me try restarting my language server just to see if that's real. Hmm. ⁓ you know what? That's legit. I think we forgot to bring that over. Let's check our UI, package.json. There it is.

Helpful tooling, making sure that we do all the things we need to do. ⁓ PNPMI.

Sneha Mehra (00:04:51)  
Great. Looks like that was successful ⁓ and goes away. Fantastic. ⁓ I want to change this from dir name to ⁓ process oops.cwd.

And ⁓ we need types.no types node for that to work. So let's let's install that at the workspace level.

Sneha Mehra (00:05:21)  
That's for sure the YesLint config is being evaluated at ⁓ you know, in in a node context.

Sneha Mehra (00:05:37)  
this is interesting.

Save?

Check in the package JSON, make sure everything looks right.

Sneha Mehra (00:05:51)  
Seems good.

no, it didn't add the tapes note.

Sneha Mehra (00:06:05)  
That's really strange.

Let's try something else. We can grab it ⁓ and bring it in and then just run

The install process.

Sneha Mehra (00:06:23)  
Okay, something about I guess looking up with the latest version of at types node. It's a little weird, but so now we've got types node. We go back to this file. There we go. We've got process current working directory. ⁓ and because this is a script that we should be running at the workspace level, this will always be the root of the project, right?

Sneha Mehra (00:06:52)  
And here's Pnpm Lint. All right. So we got some errors. ⁓ and this the like this represents us having some legitimate things that we need to catch. ⁓ so we can start with first this ⁓ invalid template literal.

Sneha Mehra (00:07:21)  
Restrict temporal literal expressions. ⁓ That's what it's called. So we're gonna go rules. here it is. ⁓

I already had that in there.

Well, let's let's look at a couple other of these. Like this this to me screams an import's not resolving.

And there it is. So now for the first time we're we're catching that ⁓ some of these ⁓ some of these imports aren't resolving. The reason why this wasn't being detected before is we've been doing a lot of like PNPM builds, which remember, just looks at the source folder. And this is why it's useful to have that check command so that you're also type checking against your tests. But this just needs to be like you see all the red over here, ⁓ and it's just gonna go away once we resolve.

Svelte or seeds and models. Great. ⁓ so there's that that file.

Sneha Mehra (00:08:25)  
We can run again. Oops. And we should end up with just two errors left. These two, tailwind config, so the Svelte config, the tailwind config, and the vite config. ⁓ one more.

Sneha Mehra (00:08:44)  
That ⁓ should be fine. What do we got here?

Sneha Mehra (00:08:51)  
Okay, so now I'm seeing a rule, ⁓ just verbalizing my my debugging steps here. I know I have a rule in my TS config. Sorry, in my ESLint configuration that allows ⁓ numbers. Like we ran lint at the beginning of the project and it for sure passed. We all saw that. So that tells me, mmm, that tells me I need to look at paths here. Look.

We're look we're relative to the root of the project, we're looking for source and tests. So here we go. We're gonna add to both of these places packages slash anything slash source ⁓ and tests. And you should see some of those things allowed. There we go. So now we're down to a set of errors that all are are similar in nature. And they're saying, ⁓ there's a parsing error here.

I'm not included in a TS config or or something like that. So two two places this could happen. One is a TS config is not covering these things. But another would be just making sure that you have some representation of these files here. So like at a high level, I want you to think about this as either ⁓ the ts config contained ⁓ the ts config referred to files, which are

neither ignored ⁓ nor described in terms of how they can be linted, ⁓ or your linting files that your TS config doesn't cover. So let's see, let's see what we can do about that.

Sneha Mehra (00:10:38)  
⁓ I think sorry, I'm just gonna check my notes real quick here.

Sneha Mehra (00:10:56)  
Packages UI. All right. Let's just double check the TS config one more time.

Sneha Mehra (00:11:09)  
UI TS config. There it is. ⁓ Okay. ⁓ And

Sneha Mehra (00:11:18)  
Try that. No, that shouldn't matter. this can go away because this file is not present anymore. But there's the Svelte config, there's the Vite config. What if we add these to our root TS config? That would be the other thing to experiment with. So this would be saying packages, ⁓ or we could do star star slash that. And let's see if this happens.

If if if this works, what that tells me is we're only ESLint's only looking at that root TS config and it's not sort of cascading into each project.

There, it works. So so basically, like ⁓ just through debugging this, let's think I think we can make an assumption now. Like, although the TypeScript compiler is fine with the include array sort of being ⁓ added onto, if you will, in the UI package, right? Like it's it's still going to be including source and test because of what it's extending from. ESLint, it appears, does not work that way.

And we're having to say, you know, in our config, when we're saying I'm pointing to this ⁓ folder, like that's where you can find the TS config that's going to be used for linting. That's where it wants to be able to find everything, ⁓ everything that is being asked to lint. And linting passes now. And just to convince ourselves that it's actually working, ⁓ we can have a lint error of some sort, and let's make sure that it ends up being picked up. Like

⁓

Sneha Mehra (00:12:55)  
What would we do?

Sneha Mehra (00:13:01)  
we could do this.

Sneha Mehra (00:13:11)  
So that should be a lint error.

No confusing void expression. Save.

And we see the errors pop up. Great. Now we have ESLint working across our workspace. ⁓ And this is the advisable way to do it, by the way. Like having ⁓ one lint task that works everywhere, it's just gonna be a lot more ⁓ a lot more efficient than having linting happen in each monorepo package. If you look at what the ESLint team says ⁓ in their documentation, like once you start getting to even something like 10 packages.

⁓ especially if the linting is happening in parallel. Like if you don't want to have an ESLint file in each package at that point. It's it's just a lot for it to parse and just has to do with the way it's implemented. But this this scales up much better. Having one ESLint.mts at your root and have this be the central place where all of your ⁓ where everything's happening. And of course you get the added benefit of saying, well,

Just like we want one place for TypeScript compiler strictness settings, we want one place to look where we can state across the entire monorepo, here are the roles. And yes, you might have deviate deviation between different packages where let's say you're incrementally tightening things up, like applying a new rule and getting everything fixed in one package at a time and rolling that out. Well, you can still do that in this file. Like you can have as many of these objects as you want, ⁓ and this would be another great place to have like.

Sneha Mehra (00:14:48)  
Per package configuration, but at least there's one place to look, and you're not opening up a dozen files to figure out like what the heck is going on. Next up, ⁓ let's get that jump to definition working. ⁓ and so just to to refresh what this problem is, in any place where ⁓ yeah, this will work. In any place where we're importing ⁓ like something from either the UI or the server package, and we're importing from.

Seeds models. If we command click, we end up going to these declaration files. There's a very simple fix for this, and that is ⁓ in your TS config build, you want not just declaration true, but declaration map true. Think of these as the source maps for declaration files. So this one's super easy. If we build across the whole project, it'll result in a new a new build output.

I sorry, I'm in the UI folder running the build command in the UI package only.

Gonna get out of here and I'm gonna run build across the whole Monorepo, at which point that there, that's what I was looking for. These declaration map ⁓ files, which you can see, all right, like here's here is the declaration file, ⁓ and this is the source that it comes from, or the sources. And so as a result, now if we go back to that formatting TS and command click, we're now in TypeScript source.

Right. So very important, declaration map true. There's really no downside to building those ⁓ you know, anywhere you care about declarations. Any library should have these declaration maps in place. Even if it's in your node modules folder, it's still valuable to be able to sort of ⁓ you know, like if we're here and we're going into Svelte, like it's it's nice to be able to to jump into, you know, some something that's ⁓

Sneha Mehra (00:16:51)  
You know, more readable here. In this case, we're we're benefiting from source maps with JS doc types, but you know, it's the original source code with the comments and it allows you to spelunk into your dependencies ⁓ without having to sort of walk back up and be like, I'm in a dist folder. Let me get into the original source code. Yes. If you were publishing something to npm for that ⁓ property, would you then have to include your

TypeScript in the ⁓ DIST, or would the jump to just jump to the JavaScript code? That's a that's a good good question. When I publish TypeScript libraries, I will typically leave my source code in the library for that reason. So, like when if I were to publish models models, you would see something that looks exactly like this. You'd see a dist folder, and that's part of the tarball that ends up going up to npm.

⁓ but I'll absolutely leave my source folder in there and I'll even leave the tests. Like it's it's more text, but ⁓ there's still a lot of value in somebody being able to go in and understand ⁓ exactly exactly what's going on. And especially if you want to look into a concept like ⁓ like creating a patch. So all the all the popular JavaScript package managers, NPM Yarn and P NPM.

They support this concept of like a patch command, which lets you really reach into a node modules folder, ⁓ make a little adjustment, and then you'll you you create a git diff effectively, like a git patch that represents that adjustment having been made. And then when you install that package again, your patch, which is checked into git along with your project, that's applied. So if you've ever

Run into a problem where like some dependency has a bug. Like I I run into this all the time where somebody has a very old way of representing types. Like they're using in their ambient type information both the declare keyword and the export keyword. And like at one time the TypeScript compiler was okay with that, and it is not anymore. And so you can just go in and you can change those files and you create a patch. And you don't you don't have to fork the library.

Sneha Mehra (00:19:18)  
And to publish it to NPM and then pull it down. ⁓ still good to open a PR and to see if you can help fix it for other people too. That's just sort of a good ecosystem citizen thing to do. But the idea that you can sort of make that adjustment and check it in as a Git patch, it's it's pretty powerful. And you do it once, and then it's really it's like persisted with your source code. It's in Git. And so, ⁓

And that's that's like that's a good reason to include the source code in your library. So if somebody ever needed to do that with something I'd published, well, they would have the ingredients necessary. They'd be able to go in there into their node modules modules folder and like npm install from within that node modules and use my build script to they could adjust the source code and then they could, you know, ⁓ run a build ⁓ on

What is to them a dependency, and then they'll get both the source code changes and the dist, and they can check that in as sort of a patch with their repo, which is pretty cool. And in the monorepos world, you can do that in a way where that patch is applied ⁓ for any ⁓ package ⁓ in your monorepo that uses the same version of that dependency. So if you had to patch React or you had to patch something else.

Like you can do it once ⁓ and then everywhere it ends up being used.

