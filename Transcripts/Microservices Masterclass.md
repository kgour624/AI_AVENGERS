Sneha Mehra (00:00:02)  
The mantra is don't repeat yourself. We'd like to normalise our work so that there's one and only one version of any particular behaviour in our code. Then, if we need to change anything, there's only one place to change it. This is nearly always good advice. But not always. For bigger or more complex systems, coupling is an equally implacable enemy. At least as important as duplication.

And to make things worse, the trade off for don't repeat yourself is increased coupling. This is commonly one of the biggest stumbling blocks for teams trying to adopt microservices. So what about those times when dry is the wrong answer? When is dry a problem rather than a solution? In this episode, I'd like to explore the costs and benefits of removing duplication in our code.

Don't repeat yourself, or dry as it's more popularly known, is probably one of the best known design heuristics. If you ask people about the guidelines that they use to improve the quality of their designs, dry is usually high up on their list. This is for a good reason. It makes a lot of sense. Certainly in big code bases, one of the commonest anti-patterns, a form of waste in the design, is that behaviour is duplicated in multiple parts of the code.

This is bad because it means that when we need to change that behaviour, we will need to identify all of the places in the code where it exists, which may or may not be easy. It also means that as a result of this first problem, over time it's very likely that the different copies of this behaviour will start to drift apart. We will end up with several different, slightly different versions of the same behaviour in different parts of the code, each a little different from the others,

and this makes for systems that are unpleasant to use and ⁓ if not just out and out wrong, as well as systems that are a pain to maintain. This is because we have to hunt down all of those different places where we do something if we need to fix a bug or add the behaviour. Over time this just gets worse and worse. That's why people started to recommend dry as an approach. Wikipedia describes dry like this.

Sneha Mehra (00:02:27)  
Every piece of knowledge must have a single, unambiguous, authoritative representation within a system. This is good advice. If we organise our code like this, it's much easier to work on. Here is some unpleasant code. OK, it's written in C, but just because you're writing in C doesn't mean that you have to write crap code. This little block of code here is repeated 13 times inside this single function.

This is horrible code. Clearly the developer either had never heard of dry or didn't care enough about it to ⁓ try and apply it to this piece of code. At this level, dry is not just a good idea. It's pretty much essential to doing a decent job. We can improve this code trivially by creating a small function that we can call to do whatever it is that this block of code is doing.

with the added benefit that we could name that block of code and so we, the readers, would have a better idea of what it was that the code was supposed to be doing. If we were refactoring, I'd probably start by extracting a method, maybe called something like debug warning. This would make the code considerably better in a single simple step. So the dry guideline has worked and helped us to write better code. But there's a problem.

One that doesn't matter and doesn't crop up at this scale. But what if the code base is bigger? Let's say for example, that we have two services, A and B. They both need debug warning. If we follow dry, then there should only be one copy. Now our services are coupled via debug warning. If service A wants to add something to debug warning, it forces service B to change in step.

We can reduce the cost of this by operating a shared code ownership kind of ⁓ approach, keeping service A and service B in the same repo along with our single implementation of debug warning and maybe using continuous integration to evaluate any changes and so help us to spot if our changes break anything. This works and is pretty good advice on the whole.

Sneha Mehra (00:04:48)  
What if service A and service B were microservices though? I talked about microservices in an earlier episode. A microservice is by definition independently deployable. The whole point of a microservices is to allow for organisational scaling through decoupling. We want teams to be able to work independently of one another. We don't get to test our microservices together before release, or they wouldn't be independently deployable.

So what does that mean for our services and their shared function? If there's only one copy of it, where do we keep debug warning? Is it in the repo for one of the services? If so, then the other service is not independent, because it now depends on the code in the repo for the first service. Our services are developmentally coupled. Maybe we should move debug warning into its own repository.

Okay, but our services may still be developmentally coupled at this point. It really depends on how we manage this. There are several different options. The easiest one to think about is if the function is just a function, we can treat it as some kind of library function that we can link with and then call. So we could add something into the build of our services to establish a dependency on our debug warning library.

The trouble here is that even now the devil's in the detail. If we choose to establish that dependency on the basis of, let's say, give me the latest version, we're developmentally coupled again. ⁓ I changed debug warning to support my new feature in my service B, ⁓ and I am forcing you to make changes before you can safely release your service A. ⁓ One way to reduce this coupling is to be more specific about the versions that we pick.

We treat debug warning rather like an external third party dependency. In my build script, I say my service uses version seven. In yours, you say yours uses version six, for example. ⁓ Now we are each in control of which version of debug warning we will take. We can decide when to upgrade and when not. So cool, we have our independent services. But let's be clear, we're no longer dry.

Sneha Mehra (00:07:15)  
We've just used our version control system to allow for two different versions of debug warning to be in use. Maybe debug warning is more like a function of a service of some kind. We could deploy it as a separate microservice perhaps. The strategy for microservices is to reduce the coupling between the services. So now we will be much more cautious of changes to the interface to debug warning. We will probably ⁓ prefer ⁓

more loosely coupled technology to represent that interface. After all, function calls are pretty tight coupled by design, aren't they? Well, ⁓ maybe. This is much more about design than it is about technology. If the function call abstracts what happens behind it, includes minimal abstract parameters. If the first thing that the function does when called is apply some kind of ports and adapters style translation,

to insulate the code from the core and to validate the inputs, then it is decoupled and so won't change very quickly or break quite so easily. Conversely, if this API is presented as a REST style core, but the data that it deals with is tightly coupled to the implementation and there's no translation step to validate the inputs, then it's still tightly coupled. ⁓ Some technologies help us to create more loosely coupled solution than others.

But the tech itself doesn't solve the problem. The design does. So there's a fundamental unavoidable link between dry and the degree to which our systems are coupled. There's a cost to sharing code as well as a benefit. So we're always treading a thin line. One side we have nasty duplication, on the other we have nasty coupling.

So how do we balance these off against each other, get the best of both worlds? Inevitably, this is complex and subjective. But to me, this is the real skill of software development. Sure, it's nice if you know your language and tools well, but that's the simple stuff compared to this. Making good choices about where you place behaviour, where you draw the seams in your code between different responsibility. This is where great developers shine.

Sneha Mehra (00:09:38)  
The starting point for me is that there is no simple answer. Dry is too simple, as we have discussed, and some level of coupling is inevitable, assuming that you want the pieces of your system to communicate with each other at all. Dry is a good starting point as a guideline, though. ⁓ I'd aim to have code in the same repo be largely dry, whatever the scale, but it's kind of fractal. Within a module,

class or file, I'm not going to accept more than a couple of lines of duplication in my code. Once we get to broader concepts where one module interacts with another, then I think you need to be cautious not to generalise too soon. But also not to be too tolerant of duplication. This is hard, but finding that sweet spot is also the joy of good design. It takes skill and experience so you can be proud of yourself when you get it right.

One of the many reasons why I value test-driven development quite so much is because it helps me with this kind of decision. If I write a test, I want it to be easy to write and easy to understand. If my code is too tightly coupled and the abstraction is poor, it won't be either. So if I find my test is hard to write for any reason, I know that I have a problem with the design of my code. So I can think harder and come up with better ideas now.

This approach is great at highlighting over-coupled dependencies in particular. It drives me to abstract the interfaces to my dependencies better. This helps me to spot generality earlier in the lifecycle, or at least opens the door to me spotting it later. And so it helps me to keep my code dry without inventing too many crappy tactical abstractions along the way. At the next level out, ⁓ services, then I think that dry gets more complex.

For me, service means pretty much by definition that we treat the interface with more care. This is a division between different parts of the code that we care a lot about. The service itself should be keeping some secrets and defending its borders. The API for a service matters more than its implementation. So now we need to be careful about DRY at the level of API, because if we get it wrong, the coupling will increase

Sneha Mehra (00:12:01)  
and will cause more problems than the duplication. At the level of behavior represented by services, the API, then they should usually in most circumstances be dry. Services should be focused on achieving a specific job or outcome, and they should be the only place to go to get that job done. The code itself, the implementation, is more problematic. Now we have to worry again about developmental coupling.

If our services are in the same repo, we have some choices to make. We store, build, test and deploy things together, give us few more options. We can alleviate coupling with shared code ownership, continuous integration as I mentioned earlier. We can use refactoring tools to help us to make changes across the shared code. If we take the step to service independence though, keep our services in separate repos, deploy them independently,

that is microservices, then we cannot do this. At this point, I would treat dry with great suspicion and great care. Here, I would much prefer duplication to reuse. The cost of coupling now are too high. The whole point of microservices is to allow teams to make independent progress.

So developmentally coupling them together to achieve reuse through dry is a big mistake that I see played out all the time. The commonest form of this mistake is for code usually called platform or common services. Code that several other services rely upon. That usually begins with the aim of trying to be architecturally dry and ends up forcing everybody to make progress in lockstep because now changes to the platform break everyone.

If your team is forced to take versions of shared code for any reason other than that it does something new that you want, you're suffering from a version of this problem. Platform and common services code should take loose coupling and good abstraction ⁓ more seriously than any other part of the system, but often they don't. So DRY is a useful guideline and a rotten rule. ⁓

Sneha Mehra (00:14:20)  
Once systems and organisations get beyond the small and simple, coupling is the real enemy. As ever, microservices are more complex than they seem on the surface. To do well at microservices, you must take coupling very seriously and use all of the tricks at your disposal to reduce and manage it. ⁓ I think that includes discarding drive between microservices as a guideline.

—---------------  
14  
Sneha Mehra (00:00:01)  
Randy is well is a well-known conference speaker and a senior technical leader with a background working in many large, sometimes famous Silicon Valley companies, including Google, We Work, eBay, ⁓ and many others. ⁓ Randy currently works as VP of Engineering and Chief Architect at eBay. My first memory of Randy was seeing him deliver a keynote ⁓ at QCon London on the topic of delivering great software at kind of web monster scale.

But since then we've met several times and I've always found Randy to be delightful company and always interesting to talk to. So it's with great pleasure that I welcome Randy Schaup the engineering room. Great. thanks, Dave. ⁓ wow, what a great introduction. And ⁓ so excited to be with you. ⁓ really a big admirer of ⁓ big admirer of your channel as we were chatting beforehand. I've watched every episode ⁓ and just loving it, loving every one of it. So

For those of you that are not yet subscribers, hit subscribe and like ⁓ because this is a great, a great channel to ⁓ be a part of. Great, thank you very much. So so you you've you spent a lot of your career working, ⁓ as far as I can see, in big complex web companies. So I'm interested in exploring your views on what software development looks like when nearly everything that you do is at scale and under the stress of millions of users.

But I also don't want to jump in with both feet into the middle of that conversation. ⁓ So I thought an interesting place to start was I saw one of your talks where you described ⁓ the different needs of architecture and design at different points in the life cycle of a product. ⁓ and in it, as a side, ⁓ kind of as a side remark, you said that you thought that in startup mode, you think that a monolith is best. While I agree with that.

completely, ⁓ I think that it might surprise some people. So could you explain why you think that's the case? Yeah. I 100% agree with that. And I love starting small because like every place that's big was once small. ⁓ And I'm sure we'll talk about all the places that are big, some of which I've gotten a chance to work with work at, most of which I haven't, ⁓ all evolve from something small. So ⁓ every place evolved from a monolith. So yeah. So when you're the way I like to think about it is there's sort of

Sneha Mehra (00:02:26)  
I for change the phrasing every so often, but like there's a startup mode where like I have an idea ⁓ and I'm looking for a ⁓ looking for a business model, trying to find product market fit. And there what you're trying to do is iterate really fast. So what you're not interested in is scale, ⁓ because your team is small and you don't even really know what the prop like the bounds of your problem. You don't have your to think to say domain-driven design, you haven't like been able to bound any of your context, if that makes sense. So ⁓ so 100% start with a monolith.

that's because you haven't by by starting with a monolith, you haven't pre-decided ⁓ the subcomponents of your of your system, right? Like you haven't already, because you don't necessarily know ⁓ your system to start with. ⁓ so I recommend in the 95-99% case, ⁓ people that are just starting out with a project or just starting out with an entire company, that you 100% start with a monolith. ⁓ And that is true of every one of the big

Companies, you know, eBay started as a monolith, so did Netflix, so did Twitter, ⁓ you know, every place that you can think of, wow, they're really web scale, they all started ⁓ as a small monolith. So yeah. Yeah, it's it's it's it's that period of kind of exploring the problem. I I I'm increasingly ⁓ of the view that working kind of defensively in terms of design and architecture, so that we are ⁓ give ourselves the freedom to find out where our decisions are wrong and and to

you ⁓ and to change them l later is important. But also as as you point out, that need to in the early days when you really got no clue is to be able to iterate really quickly so that you can learn fast really. Yeah. The what the easiest place to the easiest way to learn quickly is like everything's in a single repo. ⁓ I can do everything with refactoring tools with my entire yeah no ⁓ with my entire system essentially because it's just one thing. ⁓

And it's it's super easy. The other thing is you do not have any of the problems ⁓ of that you know, services or microservices or event driven architectures solve. Like all those things, like I'm huge fans of all those things, and where I am now, and I will get there, ⁓ like we absolutely need microservices and event driven architecture and all those fancy techniques. ⁓ we couldn't be large without them, but ⁓ the problems that those things solve, you don't have when you're

Sneha Mehra (00:04:51)  
Piny startup iterating fast, trying to find product market fit, and just meeting the needs of your of your near-term customers. ⁓ the other obvious thing is like performance, right? So like all calls are local. ⁓ So you don't you don't introduce any network inside there. ⁓ it's easy to roll out and roll back because there's one, you know, ⁓ unit, ⁓ the artifact. ⁓ so yeah, super, you know, 100% the right thing to do. And honestly,

I mean, we started with the startup, but like I think ninety percent of the ⁓ software on the ⁓ on the planet really should be done in a in a monolith. And it's the exceptional once you're in what I like to call the growth phase or the scaling phase, then you start to see the problems with the monolith, which I'm sure we'll talk about. And then this then you can start to you know leverage into the other architectural solutions for those. ⁓ and then going from there. But you know, it's an S curve and like only when you're

But the ⁓ you know, the the concave down part of the S, do you start to really need to move from a monolith to something like microservices? Yeah, yeah. ⁓ I I I spent I I spent some time last week with ⁓ with with our mutual friend Martin Thompson. ⁓ and we were in a pub in in Belfast ⁓ bemoaning the fact that nearly all projects these days seem to start with Kubernetes and separate repos for everything. And me going.

⁓ It's not a good place to start. It's ⁓ not a good place to start. ⁓ spoiler alert, eBay is exactly Kubernetes with tons of repos, but ⁓ but you're not starting. ⁓ We're not starting. Yeah, we got 27 years behind us and 4,000 engineers. So yeah, no, 100%. You know, it's great. I love to read like, well, what does Google do? What, you know, what does Facebook do? what you know, what do Baidu and Tencent and

⁓ you know bite dance do ⁓ and that's interesting but it's not it's a very small percentage ⁓ of the industry that has you know the map the kind of hyperscale problems and so yeah. So so that so so if if you'll forgive me being kind of ⁓ philosophical and esoteric for a minute. So so that that that leads me to be thinking in terms of you know I'll be interested in your views on architecture, what you think architecture is and what it gives us. You know, if it's

Sneha Mehra (00:07:13)  
If it's not just about looking at eBay or Google and saying, we'll do what they're doing because that make that helps. What is it then? ⁓ Yeah, I mean, it's one of those like ⁓ yeah, like you know it when you see it. I mean, we've done a couple of interviews where, you know, like with with Simon Brown and others where we tried to explore the topic of what is architecture. And I don't know I have much to add to that. I mean, there's architecture is the hard decisions, ⁓

I like to think of it as the it's like, I mean, my mental model is it's the skeleton ⁓ you know, of the overall system within which we, you know, ⁓ place stuff. ⁓ and ⁓ yeah, so but what but what it's for, ⁓ it's a I ⁓ my mind is saying tool, but there's I'm sure a better word. Like it's just a means to an end. Like there's nothing magical about the architecture. There's no, I mean, this is obvious, there's no one right architecture. There are

architectures that are good and not good at you know startup scale growth scale ⁓ hyper growth you know hyper scale ⁓ but you know there's no there's no one right architecture for everything and ⁓ and so ⁓ I I would just suggest that it's whatever ⁓ makes your job easier doing your work. And the one thing that I would add to the discussions I've been listening to is over time your architecture is likely to change. If you

should be so lucky as to like grow into the growth scale growth phase and you hyperscale phase, you absolutely are going to change your architecture in some cases five different times. Like that's what we did at eBay. I can tell you that trajectory. ⁓ so it's totally legitimate to do that. ⁓ but what what is true at every scale is you are going to be changing your software. So what makes it easy to change your software?

—----------------------

10  
Sneha Mehra (00:00:01)  
I don't know if I'm right about this. I haven't seen any industry wide surveys, but it's my impression, based on the clients that I work with and the people that I talk to, that the commonest way to start a new project these days is to guess at the breakdown of services somehow and then create a separate repo for each one and then begin. This is a terrible idea. Somehow we've demonized the idea of a shared repository.

And for some people, the idea of what is sometimes called a monorepo has become almost a synonym for evil. Version control is just a tool. It's not a tool of torture, so it's not likely to be inherently evil. So what are the pros and cons of multi and monorepos? And yes, this is kind of another episode about microservices.

So what is a repository? ⁓ And what do we really mean when we think about the idea? I think it's really about defining a useful working scope for our software. If you use version control and if you don't, stop watching this video immediately and go and install a version control system right now. But if you do use version control, then what is it really for? What are the versions that we're controlling?

The main thing that a version control system gives us is the ability to step back to safety when we make a mistake. This is a huge advantage. It means that we can make a change with confidence that we're never too far away from safety. It also gives us the ability to share our changes with other people more easily. I think that this feature is very much in second place to being able to step back to safety, but it's still extremely useful. And third,

A version control system provides a shared safe place where we can store and protect our changes. We have a backup so we can be confident that we won't lose our precious code in the event of a hardware failure, for example. There are other advantages too, but I think that these three are the big ones. So given these advantages, how do we decide what to put into our repos? What's the right scope for version control?

Sneha Mehra (00:02:19)  
I'd argue that there are three models and one of them's dumb. We can put everything into one big repo. Let's call that a mono repo. I think that the term is the wrong idea, but we'll get to that later on. ⁓ Or we can divide our system into smaller pieces and put each piece into a separate repo. This is the multi-repo approach that's very common these days, but there are two versions of this. In one version

Each of these repos contains an independent decoupled piece of software. That is, it doesn't change in any way that forces change on code that uses it that are in other repos. This is the microservices approach, ⁓ and this is the most scalable way to build software. But that comes at a price. In other versions of a multi-repo, a pattern I call coupled modules, we divide the software up into pieces again.

Each in a separate repo, but they're sufficiently coupled to other pieces in other repos that we don't really trust the interactions between the pieces unless we test them all together. In case this is not clear, I think that this is the dumb one. Let's explore these ideas in a bit more detail to find out why. So we have our three advantages of version control. We can step back to safety, share code with others, and we can protect our assets.

So version control records the history of changes to our systems. Here's a system made up of several pieces. We'll put all of these pieces initially into a single repo. Then that scope, that set of changes, represents a definitive version of that overall system. We know that version one of A works with version two of B and C and version three of D. There's no room for confusion here, assuming that the system works at all.

We know that these pieces work together. If I make a change, let's say I change B to version 3 that breaks something, I can easily step back to the previous version or any other and then figure out what went wrong. If I make a change that means that I need to change my code and your code together, I can do that and the next time you sync you'll see my changes.

Sneha Mehra (00:04:40)  
So we can easily share even quite complex coupled changes that cross boundaries between our pieces. Would when would I see a breakage? If it's all in a single repo, then I'll build and test everything together. Depending on our technology, I can probably spot breakages in my IDE before I even commit sometimes. But if not, then my tests that run during the commit stage of my deployment pipeline will find the problem, and if not then,

Then my acceptance test will test that test the whole system together, will certainly find the problem. If I want to back up, well all of the info that defines my system is contained within the repo. So that's simply a matter of backing the whole thing up. It represents a definitive statement of all of the dependencies, all of the code, everything. So all three of the advantages of version control are simple to access for a mono repo.

Let's look at the multi-repo approach. In this model, each service lives in its own repository. So here is the same system, but now we'll store each piece separately. There is clearly a problem here. There is stuff that isn't in any of the repos. The communication between the pieces of our system and the specification of which versions of the pieces work together to make a whole system.

These things aren't stored anywhere. As I said, there are two approaches to solving this problem. A defining characteristic of the microservice approach is that each piece is independently deployable. That is, you don't get to test these interactions between services before release. The implication of this is that you don't care which versions make up the system.

Because you design the system so that it is tolerant of different versions. ⁓ In this model I can safely change any of these services independently of the others, in confidence that my changes won't break any of them. And also won't force any change on them either. You do this through design, and in particular the design of the interfaces between the pieces. These points are special, they're important and so they are treated with more caution and care.

Sneha Mehra (00:07:00)  
There are two strategies that help to achieve this. The interfaces are so well understood, so stable, that they never need to change, think Amazon S3 or TCPIP. ⁓ Or you implement them in a way that they are flexible and can cope with change without breaking. Think web browsers and HTML. So how does this stand up to our desired properties of version control?

Well clearly for changes to a single service it's simple. I can easily step back to a previous safe version. But what happens if I introduce a change that does break one of these inter-service interactions in some way? Well the first problem is when do I find out? If my testing is good, maybe I have some contract tests to check for breaking changes.

and then I'll find the problem on commit where my test fails in my deployment pipeline. If I don't have good tests though, then I won't find this until I'm in production when stuff breaks. In a real microservice environment, finding the cause of these kinds of failures can be very complex. Versions are changing all of the time, production is a moving target. There is no definitive statement anywhere that version X of A works with version Y of B.

After all, they never met until they met in production. So what version should you step back to? It's easy to imagine complex scenarios where it's impossible to step back. The answer is to work in ways that let you roll forward rather than step back, and this is a common approach in microservices teams. But now we've given up one of the big advantages of version control. That may be a worthwhile sacrifice.

If you're gaining from the advantages. But it's always a cost. Organizations that are good at microservices are clever at not breaking interactions between services, because that is the only game in town, really. But that is always extra work and requires a degree of design sophistication to achieve. What about sharing code? In a monolith, everyone has access to the same code.

Sneha Mehra (00:09:21)  
In microservices that's not necessarily the case. So if I make a change to my service that also requires you to change yours, it's going to be hard for me to know that that's the tr that's the case, but even then I can't make a change that change myself without having access to your repository. And now I have to coordinate the work between two separate repos. This is not impossible, but it's more complex and slower.

Than having everything in the same code base. So this slows me down. It either prevents me from making such changes altogether or it makes them harder work. So this works against sharing changes easily. Let's be absolutely clear ⁓ none of this means that software development is impossible under these circumstances. It just means that things move more slowly and that there's more friction. Once again,

The answer in microservices is to attempt to eliminate such coordinated changes altogether, so we're back to stable interfaces or very flexible, loosely coupled APIs that are more difficult to design. What about protecting the code? Well again, easy, maybe even trivial for a single service, but what about the system as a whole? Remember, there's no version controlled record of which versions of the services work together.

And no version controlled record of which versions are in production at any given point. So we can't back up the whole system. In a microservice world, this is fine. It's not a problem at all because remember, our services don't care what versions the others are at. It's a moving target, that's okay, but it's a more complicated thing. In essence, a microservice approach makes no sense unless it is a distributed approach.

The goal of microservices is to scale development, and it's a wonderful at that. It's the most scalable approach. But it comes at a significant cost of complexity and overhead. This is a worthwhile trade-off if you're Amazon or Netflix, but maybe not so much if you're a smaller team that doesn't need the scalability. So what about the other multi-repo approach, the coupled modules thing? ⁓ This is the approach that everybody prefers the sound of.

Sneha Mehra (00:11:41)  
But they like it because they want to have their cake and eat it too. Coupled modules doesn't really make any sense at all. The coupled modules will once again allocate a piece of our system to a dedicated repository. ⁓ As before, what about the conversations between modules and the versions of the pieces? I think that the idea behind this approach goes something like this. We can't imagine building and testing all of our stuff together as a monolith.

And being able to get results fast enough. So we'll break our system into pieces because that will make our lives easier. Then we can build and test each piece in isolation much more quickly. But we also can't imagine designing the pieces to be independent like microservices. So we'll test all of our pieces together before we release. We'll test those interact inter-service interactions and we'll need to store the versions of the pieces that we know work together somewhere. Hmm.

So now we've got a monolithic system, same as before. But this time, with all of the impedance and friction of a microservice system, without any of the benefits, the pieces aren't independently deployable, and they aren't independently developable either, because we don't know if our changes really work until all of the integration tests have run. So as I said, we're firmly back in monolith land. But actually it's worse than that.

By separating our coupled modules into separate repositories, we've eliminated several useful optimization techniques that we make sense of monoliths, but that don't work for multi-repos. For example, in a monolithic build, we can take advantage of things like incremental builds. We can use techniques like static typing to spot mistakes even earlier than continuous integration would. I think that coupled modules is just a dumb approach.

It's like the worst of all worlds with none of the benefits. But it's incredibly common. In fact, my experience is that most teams that claim to practice microservices are in reality doing coupled modules of some form. The test is always ⁓ can you deploy your microservice without testing it with others first? I think that this gives us a strong clue of a better way to think about this problem.

Sneha Mehra (00:14:09)  
The deployability of our system is absolutely key. I think that our aim should be to make the evaluations of our system definitive. We create deployment pipelines that determine the releaseability of our changes. They are definitive for release. If the pipeline says everything's good, we're happy to release the change into production. If a single test fails, then our pipeline responds with not fit for production.

To have that level of confidence, we need feedback that we gather from our test to be definitive for release. That means that we need to test everything that needs testing before release within the scope of the pipeline. So the scope of evaluation for a pipeline is fixed by the releaseability of our software. And the easiest way to manage that is to make a pipeline service a single repository that defines that scope. This means

The scope of a single repository is determined by the releaseability of your system or service. A pipeline evaluates an independently deployable unit of software, and that is the easiest to accomplish when that independently deployable unit is stored in a single repository. So monoliths are fine. Real microservices are fine too. But coupled modules are only very inefficient monoliths in disguise. Your independently deployable units

Don't have to be microservices. I've used that tag as a kind of shorthand. But I think that the idea of them being independently deployable is the more important idea ultimately. If you have a big subsystem or even a small library, if those things are independently deployable, they may sensibly be stored in their own repository with their own deployment pipeline. If the contents of your repo aren't independently deployable,

I think you're almost certainly wasting time and effort. Thank you very much for watching.

—-----------------------------  
13  
Sneha Mehra (00:00:01)  
Is that what you think like a good architecture does is support this iterative approach, support the quicker delivery of of of new stuff to production. That that's what a a good architecture does, or is there more to it than that for you? ⁓ essentially that's exactly what a good architecture does. Peop people say that the Agile Manifesto doesn't talk about design and therefore you should not do upfront design to to kind of echo the same thoughts.

And I I've seen teams go from big upfront design to basically nothing, and they've realized that's now also a bad idea. ⁓ And in order to move fast, in order to embrace change and deliver stuff quickly, and use all of the DevOps tools and CI CD tools to kind of move fast and deliver stuff properly in a in a structured, more engineering-based way, you need a good design. One of the principles in the IDO manifesto actually says ⁓ a continuous approach to to good design enhances agility. ⁓ You don't get a good design.

Just by hacking code for free. Yeah. No. You have to put some thought into it. And although I I completely agree that we need to think about stuff in an evolutionary way because we're going to get changes and we need to pivot and change direction, ⁓ I think you still need a starting point, not all of the starting point, but a starting point with some principles in place. So that allow that allows you to create that good structure, that high degrees of modularity, so you can move fast. So yeah, it's it's a blended approach. That you're saying that you're

design should evolve with your product and with the things that you learn from pushing stuff out. But you mentioned you want to get some stuff in place in the beginning already. ⁓ He didn't quite say that if you'll I'll I'm I'll if you forgive me if I'm putting words in your mouth. What he said is that you s you you start off with a model. Yeah. You start off with a with with with with with an idea for what your design might be. ⁓ I would couch that from an engineering principle.

is that you start off with a model like that and assume that it's wrong. ⁓ That's the step to engineering. ⁓ So you assume that it's wrong and then you work in a way so that when you find out where it's wrong, you can correct it. Yeah, right. And that's very different to big design up front, because when people did Big Design Upfront all those years ago, they assumed they were right and they assumed that set of blueprints they came up with was the thing they should always aim for. So I think we're saying have a starting point, ⁓ be prepared for that to change.

Sneha Mehra (00:02:18)  
And of course, DevOps and C I and C D give us the tools to make those changes easier if you have a good architecture in the first place. Yes. Okay. So what are the stuff what is the stuff that you would focus on first? Like if you ⁓ you know where you want to go in the in the long term and what kind of architecture you would need, what kind of design you would need to to like support the final product, but you're not gonna build all of it at once, right? What are like the non negotiables? What's the stuff that you

always need even if you start out with your first version that you're pushing out. Do you mind if I take that first? Because I I think I can lead you set you up for flushing more detail. ⁓ So from an engineering point of view, the things that I would describe are all about managing complexity. ⁓ I would start to try and identify ways of compartmentalizing the system so that I'm able to understand the pieces and change them without affecting other parts of the system.

And I would say that's a deeply profound and important aspect of architecture and design. ⁓ And then, you know, that's if you're able to do that, so if you're able to build systems that are more modular, ⁓ more cohesive, good separation of concerns, good lines of abstraction, ⁓ tending towards t being more loosely coupled between those those pieces, ⁓ that that's the kind of defense that you then have to allow you to.

find out you screwed up and made a mistake and change things and and manage make the code you know a habitable space that you can change. ⁓ And I think tight Simon stuff, as I understand it, takes that, you know, a la gives you tools that allow you to achieve that those kinds of ends. Yeah, I was gonna say it's literally the same thing. So Grady Booch has a great definition about software architecture. He says ⁓ software architecture is about the significant decisions. Yeah. All of that stuff is significant decisions. It's your key tech choices that you can't really change later.

It's your overriding modularity strategy, whether you're building a monolith or a certain microservice or something in between. ⁓ And again, it's it's how do we make this thing so that we can change it in the future without having this horrible blast radius effect that, you know, you make change here and everything blows up. Yeah. And I'd I'd argue to some degree that architecture's nearly all about that ⁓ that that manage management of complexity. It allows us to build systems that are beyond the scale that we can hold in our heads. Or at least a part of it.

Sneha Mehra (00:04:38)  
of the system that you can hold in your head. Right. Will you compartmentalize it so that each piece fits in your head? We've seen a lot of the practices from the big players that have moved into the common domain. ⁓ If we look at technologies like containers, orchestration, ⁓ more than ever, the the ways that we can do pipelines has been commoditized. You can do that on on so many platforms now.

With great power also comes great responsibility. ⁓ do you think that people are hurting th themselves with with these technologies as well? I do. Yeah, I do too. I and and I it's there's there's there's an elephant in the room, so let's ⁓ name the elephant. I I think that people get microservices wrong all of the time. ⁓ I I I think that microservices have become mu most of the

Clients that I I I'm like Simon, I'm an independent consultant, and most of the teams that I see that claim to be adopting microservices aren't, by the definition of microservices, doing so. ⁓ And where they start is, you know, if they're starting something new, they start by assuming that they understand what are the services, setting up a separate, creating a separate repo, ⁓ and then starting working each of those things. What they've just done.

Done is built latency into the point at which they want to iterate quickly in order to be able to learn. So the other aspect of engineering is to optimize for learning. ⁓ So you want really fast, clear feedback. If my service is in one repository and Simon's is in another, every time the conversation between those services changes to the smallest degree, you know, either I've got to go and dip in his or he's got to dip it and it's a nightmare. If we put them all in one big reaper, we still have nice service-oriented designs.

Yes. But but probably ninety percent of the time my IDE will tell me that I've screwed up his service. Can you do that? Can you like take that a step further? And when you're starting out, team is still small, company is still small, not just put it all in one repo, like host it all in one process. Or do you think that's like a horrible idea? No, that that that would be my recommended starting point for ninety-five percent plus of teams out there. Yeah. IT did a talk at a go to conference, I think it was a go to about a few years ago called Modular Monoliths. Yeah. Same thing.

Sneha Mehra (00:06:54)  
⁓ I have a very similar talk. Yeah, yeah. I think there are th I think there are a few people now with tools and they're finally becoming fashionable again. ⁓ I've seen a ho the same thing, a whole bunch of people who've got this like ten, fifteen year old JavaSLegacy application. It's a horrible mess. They it's brittle, it they can't change it. And they say we're gonna convert to microservices. ⁓ And what they do is they take their existing design thinking, their approach to modularity, which is not very good because that's what got them into the mess. Yeah. And they basically stick JSON over HTTPS network links between things in their monolith.

That could have been in process costs. Yeah, and right. And and and now the boundaries are wrong. The boundaries are hard to change and you've got something which is lockstep deployable, brittle, fragile, and slow, and yeah, they just don't get there's a very different mindset shift there. Because don't get me wrong, I mean at a certain scale, you're gonna want separate services. But ⁓ maybe. I mean Facebook, Shopify, there are some big modular monits out there. Shopify have got a huge big thing on their engineering blog.

Yeah, over the past few years about how they've changed their Ruby on Erasmus and it's become much more modular because they were running into issues. ⁓ I'd I'd I'd I'd argue that modularity's always good, but not necessarily you don't necessarily need inter process communications all over the place. And ⁓ often you don't need multi threading in lots of places where people put it. And that all of those things b well, both of those things, you know, amp up the complexity by an order of magnitude at least. Well I'm not just the complexity, also like

⁓ And even ⁓ just deploy. Just once you're figuring out what what is my software doing in production, that becomes extremely hard. Is that like something that you take on from the get-go, like visibility of your systems? Is that something that because to me that always felt like one of the most important issues that a lot of people seem to be forgetting? This is yeah, I mean this is why some big organizations who are very

service microservice focused, they give their teams autonomy, but they have internal engineering and platform teams that bootstrap the product teams and service teams. So literally you can pull something out of their internal repo, ⁓ bootstrap your service, and you get observability free and monitor monitoring for free and deployability for free into the ⁓ production GCP environment and all of that stuff is taking care of you in a in a standardized way. And that's fabulous. ⁓ The the the the other the other thing is th that that microservices give you if you do it well.

Sneha Mehra (00:09:18)  
And at scale is it's the most scalable way of building big systems. Because what you do is that you trade off ⁓ consistency ⁓ for independence. So the this is the most distributed approach to development. But it means that if I'm writing a microservice and Simon's writing a mic microfer service, ⁓ I'm gonna I can deploy mine without testing it against his.

That's how good the abstraction is between them. And that's kind of table stakes. You can't really count it as microservices if you can't do that, because that's the decoupling step. ⁓ The point at which we don't no longer care about about the details and the pro that means the protocol's got to be stable between us. So you've got to be fairly sophisticated in design terms to be able to get to those stable protocols. That requires some really competent architects. Because that requires both business knowledge.

And technical knowledge to define those bo boundaries in in the in the right places because are because otherwise they will be working against you. Yeah. And that's why most teams should not do this, because it's hard. Yes.

—-------------------------------  
15  
Sneha Mehra (00:00:01)  
Is that what you think like a good architecture does is support this iterative approach, support the quicker delivery of of of new stuff to production. That that's what a a good architecture does, or is there more to it than that for you? ⁓ essentially that's exactly what a good architecture does. Peop people say that the Agile Manifesto doesn't talk about design and therefore you should not do upfront design to to kind of echo the same thoughts.

And I I've seen teams go from big upfront design to basically nothing, and they've realized that's now also a bad idea. ⁓ And in order to move fast, in order to embrace change and deliver stuff quickly, and use all of the DevOps tools and CI CD tools to kind of move fast and deliver stuff properly in a in a structured, more engineering-based way, you need a good design. One of the principles in the IDO manifesto actually says ⁓ a continuous approach to to good design enhances agility. ⁓ You don't get a good design.

Just by hacking code for free. Yeah. No. You have to put some thought into it. And although I I completely agree that we need to think about stuff in an evolutionary way because we're going to get changes and we need to pivot and change direction, ⁓ I think you still need a starting point, not all of the starting point, but a starting point with some principles in place. So that allow that allows you to create that good structure, that high degrees of modularity, so you can move fast. So yeah, it's it's a blended approach. That you're saying that you're

design should evolve with your product and with the things that you learn from pushing stuff out. But you mentioned you want to get some stuff in place in the beginning already. ⁓ He didn't quite say that if you'll I'll I'm I'll if you forgive me if I'm putting words in your mouth. What he said is that you s you you start off with a model. Yeah. You start off with a with with with with with an idea for what your design might be. ⁓ I would couch that from an engineering principle.

is that you start off with a model like that and assume that it's wrong. ⁓ That's the step to engineering. ⁓ So you assume that it's wrong and then you work in a way so that when you find out where it's wrong, you can correct it. Yeah, right. And that's very different to big design up front, because when people did Big Design Upfront all those years ago, they assumed they were right and they assumed that set of blueprints they came up with was the thing they should always aim for. So I think we're saying have a starting point, ⁓ be prepared for that to change.

Sneha Mehra (00:02:18)  
And of course, DevOps and C I and C D give us the tools to make those changes easier if you have a good architecture in the first place. Yes. Okay. So what are the stuff what is the stuff that you would focus on first? Like if you ⁓ you know where you want to go in the in the long term and what kind of architecture you would need, what kind of design you would need to to like support the final product, but you're not gonna build all of it at once, right? What are like the non negotiables? What's the stuff that you

always need even if you start out with your first version that you're pushing out. Do you mind if I take that first? Because I I think I can lead you set you up for flushing more detail. ⁓ So from an engineering point of view, the things that I would describe are all about managing complexity. ⁓ I would start to try and identify ways of compartmentalizing the system so that I'm able to understand the pieces and change them without affecting other parts of the system.

And I would say that's a deeply profound and important aspect of architecture and design. ⁓ And then, you know, that's if you're able to do that, so if you're able to build systems that are more modular, ⁓ more cohesive, good separation of concerns, good lines of abstraction, ⁓ tending towards t being more loosely coupled between those those pieces, ⁓ that that's the kind of defense that you then have to allow you to.

find out you screwed up and made a mistake and change things and and manage make the code you know a habitable space that you can change. ⁓ And I think tight Simon stuff, as I understand it, takes that, you know, a la gives you tools that allow you to achieve that those kinds of ends. Yeah, I was gonna say it's literally the same thing. So Grady Booch has a great definition about software architecture. He says ⁓ software architecture is about the significant decisions. Yeah. All of that stuff is significant decisions. It's your key tech choices that you can't really change later.

It's your overriding modularity strategy, whether you're building a monolith or a certain microservice or something in between. ⁓ And again, it's it's how do we make this thing so that we can change it in the future without having this horrible blast radius effect that, you know, you make change here and everything blows up. Yeah. And I'd I'd argue to some degree that architecture's nearly all about that ⁓ that that manage management of complexity. It allows us to build systems that are beyond the scale that we can hold in our heads. Or at least a part of it.

Sneha Mehra (00:04:38)  
of the system that you can hold in your head. Right. Will you compartmentalize it so that each piece fits in your head? We've seen a lot of the practices from the big players that have moved into the common domain. ⁓ If we look at technologies like containers, orchestration, ⁓ more than ever, the the ways that we can do pipelines has been commoditized. You can do that on on so many platforms now.

With great power also comes great responsibility. ⁓ do you think that people are hurting th themselves with with these technologies as well? I do. Yeah, I do too. I and and I it's there's there's there's an elephant in the room, so let's ⁓ name the elephant. I I think that people get microservices wrong all of the time. ⁓ I I I think that microservices have become mu most of the

Clients that I I I'm like Simon, I'm an independent consultant, and most of the teams that I see that claim to be adopting microservices aren't, by the definition of microservices, doing so. ⁓ And where they start is, you know, if they're starting something new, they start by assuming that they understand what are the services, setting up a separate, creating a separate repo, ⁓ and then starting working each of those things. What they've just done.

Done is built latency into the point at which they want to iterate quickly in order to be able to learn. So the other aspect of engineering is to optimize for learning. ⁓ So you want really fast, clear feedback. If my service is in one repository and Simon's is in another, every time the conversation between those services changes to the smallest degree, you know, either I've got to go and dip in his or he's got to dip it and it's a nightmare. If we put them all in one big reaper, we still have nice service-oriented designs.

Yes. But but probably ninety percent of the time my IDE will tell me that I've screwed up his service. Can you do that? Can you like take that a step further? And when you're starting out, team is still small, company is still small, not just put it all in one repo, like host it all in one process. Or do you think that's like a horrible idea? No, that that that would be my recommended starting point for ninety-five percent plus of teams out there. Yeah. IT did a talk at a go to conference, I think it was a go to about a few years ago called Modular Monoliths. Yeah. Same thing.

Sneha Mehra (00:06:54)  
⁓ I have a very similar talk. Yeah, yeah. I think there are th I think there are a few people now with tools and they're finally becoming fashionable again. ⁓ I've seen a ho the same thing, a whole bunch of people who've got this like ten, fifteen year old JavaSLegacy application. It's a horrible mess. They it's brittle, it they can't change it. And they say we're gonna convert to microservices. ⁓ And what they do is they take their existing design thinking, their approach to modularity, which is not very good because that's what got them into the mess. Yeah. And they basically stick JSON over HTTPS network links between things in their monolith.

That could have been in process costs. Yeah, and right. And and and now the boundaries are wrong. The boundaries are hard to change and you've got something which is lockstep deployable, brittle, fragile, and slow, and yeah, they just don't get there's a very different mindset shift there. Because don't get me wrong, I mean at a certain scale, you're gonna want separate services. But ⁓ maybe. I mean Facebook, Shopify, there are some big modular monits out there. Shopify have got a huge big thing on their engineering blog.

Yeah, over the past few years about how they've changed their Ruby on Erasmus and it's become much more modular because they were running into issues. ⁓ I'd I'd I'd I'd argue that modularity's always good, but not necessarily you don't necessarily need inter process communications all over the place. And ⁓ often you don't need multi threading in lots of places where people put it. And that all of those things b well, both of those things, you know, amp up the complexity by an order of magnitude at least. Well I'm not just the complexity, also like

⁓ And even ⁓ just deploy. Just once you're figuring out what what is my software doing in production, that becomes extremely hard. Is that like something that you take on from the get-go, like visibility of your systems? Is that something that because to me that always felt like one of the most important issues that a lot of people seem to be forgetting? This is yeah, I mean this is why some big organizations who are very

service microservice focused, they give their teams autonomy, but they have internal engineering and platform teams that bootstrap the product teams and service teams. So literally you can pull something out of their internal repo, ⁓ bootstrap your service, and you get observability free and monitor monitoring for free and deployability for free into the ⁓ production GCP environment and all of that stuff is taking care of you in a in a standardized way. And that's fabulous. ⁓ The the the the other the other thing is th that that microservices give you if you do it well.

Sneha Mehra (00:09:18)  
And at scale is it's the most scalable way of building big systems. Because what you do is that you trade off ⁓ consistency ⁓ for independence. So the this is the most distributed approach to development. But it means that if I'm writing a microservice and Simon's writing a mic microfer service, ⁓ I'm gonna I can deploy mine without testing it against his.

That's how good the abstraction is between them. And that's kind of table stakes. You can't really count it as microservices if you can't do that, because that's the decoupling step. ⁓ The point at which we don't no longer care about about the details and the pro that means the protocol's got to be stable between us. So you've got to be fairly sophisticated in design terms to be able to get to those stable protocols. That requires some really competent architects. Because that requires both business knowledge.

And technical knowledge to define those bo boundaries in in the in the right places because are because otherwise they will be working against you. Yeah. And that's why most teams should not do this, because it's hard. Yes.

—------------------------------

13  
Sneha Mehra (00:00:01)  
Is that what you think like a good architecture does is support this iterative approach, support the quicker delivery of of of new stuff to production. That that's what a a good architecture does, or is there more to it than that for you? ⁓ essentially that's exactly what a good architecture does. Peop people say that the Agile Manifesto doesn't talk about design and therefore you should not do upfront design to to kind of echo the same thoughts.

And I I've seen teams go from big upfront design to basically nothing, and they've realized that's now also a bad idea. ⁓ And in order to move fast, in order to embrace change and deliver stuff quickly, and use all of the DevOps tools and CI CD tools to kind of move fast and deliver stuff properly in a in a structured, more engineering-based way, you need a good design. One of the principles in the IDO manifesto actually says ⁓ a continuous approach to to good design enhances agility. ⁓ You don't get a good design.

Just by hacking code for free. Yeah. No. You have to put some thought into it. And although I I completely agree that we need to think about stuff in an evolutionary way because we're going to get changes and we need to pivot and change direction, ⁓ I think you still need a starting point, not all of the starting point, but a starting point with some principles in place. So that allow that allows you to create that good structure, that high degrees of modularity, so you can move fast. So yeah, it's it's a blended approach. That you're saying that you're

design should evolve with your product and with the things that you learn from pushing stuff out. But you mentioned you want to get some stuff in place in the beginning already. ⁓ He didn't quite say that if you'll I'll I'm I'll if you forgive me if I'm putting words in your mouth. What he said is that you s you you start off with a model. Yeah. You start off with a with with with with with an idea for what your design might be. ⁓ I would couch that from an engineering principle.

is that you start off with a model like that and assume that it's wrong. ⁓ That's the step to engineering. ⁓ So you assume that it's wrong and then you work in a way so that when you find out where it's wrong, you can correct it. Yeah, right. And that's very different to big design up front, because when people did Big Design Upfront all those years ago, they assumed they were right and they assumed that set of blueprints they came up with was the thing they should always aim for. So I think we're saying have a starting point, ⁓ be prepared for that to change.

Sneha Mehra (00:02:18)  
And of course, DevOps and C I and C D give us the tools to make those changes easier if you have a good architecture in the first place. Yes. Okay. So what are the stuff what is the stuff that you would focus on first? Like if you ⁓ you know where you want to go in the in the long term and what kind of architecture you would need, what kind of design you would need to to like support the final product, but you're not gonna build all of it at once, right? What are like the non negotiables? What's the stuff that you

always need even if you start out with your first version that you're pushing out. Do you mind if I take that first? Because I I think I can lead you set you up for flushing more detail. ⁓ So from an engineering point of view, the things that I would describe are all about managing complexity. ⁓ I would start to try and identify ways of compartmentalizing the system so that I'm able to understand the pieces and change them without affecting other parts of the system.

And I would say that's a deeply profound and important aspect of architecture and design. ⁓ And then, you know, that's if you're able to do that, so if you're able to build systems that are more modular, ⁓ more cohesive, good separation of concerns, good lines of abstraction, ⁓ tending towards t being more loosely coupled between those those pieces, ⁓ that that's the kind of defense that you then have to allow you to.

find out you screwed up and made a mistake and change things and and manage make the code you know a habitable space that you can change. ⁓ And I think tight Simon stuff, as I understand it, takes that, you know, a la gives you tools that allow you to achieve that those kinds of ends. Yeah, I was gonna say it's literally the same thing. So Grady Booch has a great definition about software architecture. He says ⁓ software architecture is about the significant decisions. Yeah. All of that stuff is significant decisions. It's your key tech choices that you can't really change later.

It's your overriding modularity strategy, whether you're building a monolith or a certain microservice or something in between. ⁓ And again, it's it's how do we make this thing so that we can change it in the future without having this horrible blast radius effect that, you know, you make change here and everything blows up. Yeah. And I'd I'd argue to some degree that architecture's nearly all about that ⁓ that that manage management of complexity. It allows us to build systems that are beyond the scale that we can hold in our heads. Or at least a part of it.

Sneha Mehra (00:04:38)  
of the system that you can hold in your head. Right. Will you compartmentalize it so that each piece fits in your head? We've seen a lot of the practices from the big players that have moved into the common domain. ⁓ If we look at technologies like containers, orchestration, ⁓ more than ever, the the ways that we can do pipelines has been commoditized. You can do that on on so many platforms now.

With great power also comes great responsibility. ⁓ do you think that people are hurting th themselves with with these technologies as well? I do. Yeah, I do too. I and and I it's there's there's there's an elephant in the room, so let's ⁓ name the elephant. I I think that people get microservices wrong all of the time. ⁓ I I I think that microservices have become mu most of the

Clients that I I I'm like Simon, I'm an independent consultant, and most of the teams that I see that claim to be adopting microservices aren't, by the definition of microservices, doing so. ⁓ And where they start is, you know, if they're starting something new, they start by assuming that they understand what are the services, setting up a separate, creating a separate repo, ⁓ and then starting working each of those things. What they've just done.

Done is built latency into the point at which they want to iterate quickly in order to be able to learn. So the other aspect of engineering is to optimize for learning. ⁓ So you want really fast, clear feedback. If my service is in one repository and Simon's is in another, every time the conversation between those services changes to the smallest degree, you know, either I've got to go and dip in his or he's got to dip it and it's a nightmare. If we put them all in one big reaper, we still have nice service-oriented designs.

Yes. But but probably ninety percent of the time my IDE will tell me that I've screwed up his service. Can you do that? Can you like take that a step further? And when you're starting out, team is still small, company is still small, not just put it all in one repo, like host it all in one process. Or do you think that's like a horrible idea? No, that that that would be my recommended starting point for ninety-five percent plus of teams out there. Yeah. IT did a talk at a go to conference, I think it was a go to about a few years ago called Modular Monoliths. Yeah. Same thing.

Sneha Mehra (00:06:54)  
⁓ I have a very similar talk. Yeah, yeah. I think there are th I think there are a few people now with tools and they're finally becoming fashionable again. ⁓ I've seen a ho the same thing, a whole bunch of people who've got this like ten, fifteen year old JavaSLegacy application. It's a horrible mess. They it's brittle, it they can't change it. And they say we're gonna convert to microservices. ⁓ And what they do is they take their existing design thinking, their approach to modularity, which is not very good because that's what got them into the mess. Yeah. And they basically stick JSON over HTTPS network links between things in their monolith.

That could have been in process costs. Yeah, and right. And and and now the boundaries are wrong. The boundaries are hard to change and you've got something which is lockstep deployable, brittle, fragile, and slow, and yeah, they just don't get there's a very different mindset shift there. Because don't get me wrong, I mean at a certain scale, you're gonna want separate services. But ⁓ maybe. I mean Facebook, Shopify, there are some big modular monits out there. Shopify have got a huge big thing on their engineering blog.

Yeah, over the past few years about how they've changed their Ruby on Erasmus and it's become much more modular because they were running into issues. ⁓ I'd I'd I'd I'd argue that modularity's always good, but not necessarily you don't necessarily need inter process communications all over the place. And ⁓ often you don't need multi threading in lots of places where people put it. And that all of those things b well, both of those things, you know, amp up the complexity by an order of magnitude at least. Well I'm not just the complexity, also like

⁓ And even ⁓ just deploy. Just once you're figuring out what what is my software doing in production, that becomes extremely hard. Is that like something that you take on from the get-go, like visibility of your systems? Is that something that because to me that always felt like one of the most important issues that a lot of people seem to be forgetting? This is yeah, I mean this is why some big organizations who are very

service microservice focused, they give their teams autonomy, but they have internal engineering and platform teams that bootstrap the product teams and service teams. So literally you can pull something out of their internal repo, ⁓ bootstrap your service, and you get observability free and monitor monitoring for free and deployability for free into the ⁓ production GCP environment and all of that stuff is taking care of you in a in a standardized way. And that's fabulous. ⁓ The the the the other the other thing is th that that microservices give you if you do it well.

Sneha Mehra (00:09:18)  
And at scale is it's the most scalable way of building big systems. Because what you do is that you trade off ⁓ consistency ⁓ for independence. So the this is the most distributed approach to development. But it means that if I'm writing a microservice and Simon's writing a mic microfer service, ⁓ I'm gonna I can deploy mine without testing it against his.

That's how good the abstraction is between them. And that's kind of table stakes. You can't really count it as microservices if you can't do that, because that's the decoupling step. ⁓ The point at which we don't no longer care about about the details and the pro that means the protocol's got to be stable between us. So you've got to be fairly sophisticated in design terms to be able to get to those stable protocols. That requires some really competent architects. Because that requires both business knowledge.

And technical knowledge to define those bo boundaries in in the in the right places because are because otherwise they will be working against you. Yeah. And that's why most teams should not do this, because it's hard. Yes.

—--------------------  
15  
Sneha Mehra (00:00:01)  
Let's move on a bit. I ⁓ one one one of the the other things that I saw ⁓ recently from you, I think, ⁓ one one of your recent talks, you were talking about contract testing and PACT. ⁓ and I I I I I I really liked your presentation and ⁓ and w one one of the ideas that stuck with me that resonated with the ⁓ that you you mentioned

You mentioned naive approach ⁓ and decoupled approaches to ⁓ to to thinking about the contracts between different services. Could you just cover a little bit that and and in particular maybe touching on

⁓ I think I think something a view that we both share is that microservices is one of those really good ideas more frequently practiced in the breach than the observance. ⁓ Yes. ⁓ Yeah, so s so I mean the idea of contract testing again is that it it's not really about testing. ⁓ again, yeah, just like T D D isn't really about testing and BDD isn't about

Testing the ⁓ contracts are about trying to ⁓ specify the ⁓ dependencies between components in a way that is both understandable to the people who need to understand them, but is also enforceable or at least checkable by software so it so it can be ⁓ the the compliance can be ⁓ automatically discovered. ⁓ and

One of the challenges ⁓ is that in the in the microservice ⁓ environment ecosystem that we live in at the moment, you don't just have component A talking to component B. You've got chains of interaction and chains of dependency. ⁓ And ⁓ the the challenges, ⁓ you know, I'm maybe we're skipping ahead a bit too quickly, but the challenge in in the the microservice ecosystem is that

Sneha Mehra (00:02:12)  
promise of microservices is that each microservice should be independently deployable ⁓ rather than having to naively release ⁓ multiple ⁓ microservices in lockstep ⁓ and one of ⁓ the the challenge there is that if you have multiple if you have got a pipeline ⁓ as your your wonderful book which i see on the shelf behind you talks about

then you'll have you'll have different versions, potentially have different versions of each component in each environment ⁓ of the pipelines. And you'll want to run the ⁓ run the tests or ensure compliance ⁓ as the component moves its way through from doing the unit tests, for instance, all the way through to deployment and making sure that it works in live.

And now we've got a lot of it's not just ⁓ it's not just one component talking to another, it's not not just one component potentially talking to many other components, it's ⁓ a version of a component that has to work with many other components that might also be in many other versions. And we've got a cross-product going on here, and quite quickly it becomes mind-bogglingly complicated to try and visualize it for a human, not so difficult for a machine. ⁓ and this is where PAC really excels.

So PACT, ⁓ which is a free open source piece of software, ⁓ ships with ⁓ a component within it called the PACT broker, which essentially tracks each version of each component, what its dependencies are on other components, and which environments each version is deployed in. ⁓ And you only have to run the test between or run the contract test between component A version

73 and component B version 212 ⁓ once. ⁓ And now you can say that whenever you are about to promote component A into a new environment, you just check have I run it, have run a test with this version of this component against the version of the component that it depends on in the environment I'm just about to deploy it to. So you don't have to run the test at the point that you want to do the deployment, you just have to look up

Sneha Mehra (00:04:34)  
in your the broker's matrix, what was the result of that test? And if the test is passed, then it's for the for the purposes of that dependency, it might be safe to promote. ⁓ and then then you scale it up to multiple dependencies and many, many components, and you get a very, very quick way of being able to ascertain whether it's safe to promote any given version of a component into a specified environment. ⁓ And that's a major win.

⁓ Beth Beth Scurry, who's one of the developers, ⁓ original developers of PACT, ⁓ has a wonderful saying which is ⁓ if if you can't promote ⁓ microservices in independently of other microservices, you don't have ⁓ you don't have microservices. What you've got is a distributed monolith. Yeah. And ⁓ there's nothing wrong with a distributed monolith, it's just it doesn't deliver

⁓ some of the beneficial properties that a microservice architecture ⁓ promises to people. And so ⁓ there are many organizations who've jumped headfirst into microservices, ⁓ maybe without understanding all of the constraints about them, but certainly without understanding ⁓ the need to be able to treat each microservice as an independent deployable that can be ⁓ evolved, ⁓ promoted and deployed

without having to worry about changing other ones. If you're introducing a breaking change, you wanna know that you've introduced a breaking change and then maybe you do have to deploy them the same time. But as ⁓ as your book ⁓ lays out in some of the pattern sections, you know, there are ways of keeping backward compatibility and then phase phasing it out later. So there are yeah, I'm not sure is that what is that what you wanted is that the question you wanted me to answer. No no that's that's that that's perfect. Thank you. And and and

And i it's perfect on mu multiple fronts in the in in that it's really understandable. You've talked about a new project and you're reinforcing my prejudices. So it's ticking all the boxes. I I I I agree entirely. I I I I often talk about ⁓ I I'm I'm ⁓ I'm a fairly I'm I'm I'm a fairly s strong proponent of a distributed monolith in the right circumstances. It has some nicer properties, it makes dependency management a lot easier.

Sneha Mehra (00:06:59)  
for some kinds of systems than microservices do. But microservices have some wonderful properties and you need to architect your system to get the gains that you want to get rather than just ⁓ following a fad. Well so I think ⁓ I think Matt Wynne did a a a talk it must be a decade ago now called Mortgage Driven Development. Yeah. Basically

I mean, it was a satire, right? So nobody listening should take it seriously. But where you try where you practice a particular skill because you know it'll look good on your C V and it'll get you a another job ⁓ later on. ⁓ Yeah, yeah. I yeah, it's it's at risk of being one of those at times. But but I I it and ⁓ I I I and contract testing is certainly, you know, a an important part of the answer. I think the other thing that we've alluded to throughout our conversation ⁓ is

the idea of more deliberate design, thinking about what those contracts are and you know, abstracting appropriately to to to reduce coupling to a sensible level, a sensible manageable level, and that gives you then opportunities to mitigate the chances of ⁓ of of a change in one part of the system adversely affecting another and your con your contract tests can, if they're good, can can detect those those the th the times when you get that wrong.

—--------------------------  
08  
Sneha Mehra (00:00:01)  
It's hard to argue that observability is anything but a good thing, but it can also be a challenging thing, particularly in distributed and even more so microservice based systems. It's also often left as something as an afterthought in the development of new features and systems. So what do we mean by observability? What are the challenges that it poses in development and specifically what are the extra challenges and importance for microservice based systems? That's our topic for today.

Observability is not a new idea, but it has become more topical to talk about it in recent years. ⁓ I'm sometimes asked what I think the difference between engineering and a more craft based approach to software development really is. And my simple answer is measurement. And for that, we need measurable systems. And to make them measurable, we need them first to be observable. That is, that it's possible for us to observe them and see what they do.

and understand what they're doing and how they're working in production. The definition of observability in software terms ⁓ is the ability to collect data about a program's execution, modules, internal states, and the communication between its components. If you're a regular viewer of this channel, you will already know that I believe that it's important to adopt a more incremental, more experimental approach to development.

To support that we need feedback on how well our choice is in terms of the technicalities of our designs, but also the functional product level decisions that we have already made are playing out. Certainly for some of these things we can get that feedback from tests before we release, but for others we can only learn these lessons that we'd like to learn from observing what actually happens to our systems in production. This is what I think of as the empirical learning

that is a fundamental part of a good, stronger, software engineering based approach. Observability is a property of a system. It's not a tool or a technique and this isn't a binary thing. The degree to which a system is observable changes what we can do. We need to get the signal to noise ratio correct. Too much information is nearly as bad as too little. But we also need information about different aspects of the system.

Sneha Mehra (00:02:21)  
And I think that this is one of the things that has changed more over recent years. It used to be that when we thought about how to observe our systems and how observable they were, we mostly thought in terms of monitoring the more technical measures of its performance, ⁓ CPU usage, disk space, and so on. But these sorts of things are just one aspect of what's really interesting and not the most interesting. If our aim is to understand what is really happening in production,

We'd also like to know how people are using our system and what value it's adding to them. At a more technical level, ⁓ observability is commonly described as being composed of three different types of telemetry. ⁓ Metrics, ⁓ logs and traces. ⁓ Metrics are the measurements that we take, ⁓ the value of the CPU usage, the number of trades per second, ⁓ or the number of new uses that are currently using our latest feature and so on.

A metric is a measurement ⁓ in a moment of time. A log is a log of events, a sequence of things that have happened often, but not exclusively kept in a human readable textual form. A log is a good way to keep track of the details of significant decisions made or errors experienced during the course of the operation of our system. Logs are often used to identify the cause of problems after a failure, recreate bugs,

or to predict problems that the software may be heading towards. The last category of observability is traces. Traces map interactions across a distributed system, a series of breadcrumbs that allow us to follow how the system was used. Which services were visited in fulfilling a particular function of the system perhaps? This is obviously useful knowledge to have, but not always easy to see, particularly in microservice-based systems.

Because in a system developed as a collection of microservices, ⁓ the main point of this architectural approach ⁓ is the scalability of development teams. Meaning that a microservice based system is built by a number of loosely coupled teams, each working more autonomously with respect to one another. ⁓ The result of this is that consistency across services is not necessarily one of the strengths of a microservice approach.

Sneha Mehra (00:04:42)  
That means that tracing the route of interaction across services can be complicated, but we'll come to that. In fact, all of these things are much more difficult to achieve in distributed systems in general. What do metrics like how many users are using feature X or what's the CPU ⁓ usage or even average response time ⁓ really mean if we have multiple concurrent services contributing to that answer? How can our logs tell the truth of a system?

when we have distributed services processing things in parallel and spread across many different services. ⁓ Even if we timestamp every entry, are we certain that the clocks are all in step? How can we identify and so trace which versions of which services processed a part of an interaction? None of these things are simple problems and they can have a big impact even for simple versions of observability. ⁓ When we adopt distributed approaches to systems like microservices,

In a microservice solution, we often have lots of different services and lots of copies of the same service, often operating in an ever-changing cloud of relationships and dependencies, because each service is deployable independently from the others, so may be changing or being updated at any time. The simple view of things is that there is a fixed route or order for conversations between our services, but part of the value of microservices is that this is not necessarily the case.

Micro services, particularly my preferred flavor of micro services, event based asynchronous services, allow us to add, replace, remove or reorder behavior all fairly easily. But while these changes may be easy to make, the implications for being able to understand what happened and observability of the system as a whole is not so easy. To make this work, we need to adopt what I think of as a ⁓ series of more distributed approaches to observing what's going on.

we need the ability to dynamically stitch together a coherent picture of the system from a series of distributed discrete pieces. This problem of application rather than service scale metrics is a really quite tricky one, but it is one that we need to tackle for systems like these. ⁓ Let's imagine the simplest possible case. We've got two copies of the same service, one running in a data center in London and another in New York.

Sneha Mehra (00:07:08)  
How many new users registered for our service between the hours of 2pm and 4pm? That's a really easy question to answer if only we had one service, but given our example, not quite so easy. We may ask, who's 2pm and who's 4pm do we mean? Because these services are running in different time zones. We could interpret this as 2-4 in London, 2-4 in New York, or 2-4 local time for each.

And all of those might be the right answer depending on what question we wanted to answer. Time is equally a problem for logs and traces though, not just for these sorts of measurements. If service A writes a log, X happened at 12.01 and B writes Y happened at 12.02, which one came first? We can't tell with only that information. We need to know about which time zones they were in, including things like daylight savings or judgements in play.

We'd also need to know if the clocks were synchronised or not. Time is a problem and not an easy one in concurrent systems. ⁓ What does concurrent mean anyway? ⁓ And at what resolution of time does all of this matter? These may seem like esoteric questions, but there are simple questions that we simply can't answer unless we worry about some of these esoteric things. Did A happen before B? If we can't answer basic questions like that, tracing the use of our system and establishing

a coherent distributed log is impossible. There are a few common patterns that can help us though to overcome these problems of observability in distributed systems. The first is maybe obvious, but time synchronisation. This can take a variety of forms. Which form works best depends on the nature of your system and the resolution of accuracy that you really need. ⁓ But the idea is that you establish a common baseline for all of the clocks

and stick to it. It's generally a good idea to store all timestamps in universal timecode UTC. Then it's easy to convert times ⁓ from UTC to your local time or whatever other time zone you need. ⁓ But as well as agreeing on your time zone and communicating and storing times, you also need to synchronise all the clocks so that they tell the same time. There are protocols that help us to achieve this.

Sneha Mehra (00:09:34)  
including for cloud-based services to synchronise time between your service. But even this may not be enough for some systems. NTP, on which most time synchronisation systems are based, provides time synchronisation at the level of milliseconds. But why should you to determine the order of things to the nearest microsecond? This stuff can get pretty complicated. ⁓ Another approach to establish a current view of the state of things is to take a snapshot.

ask ⁓ distributed resources to report their state at a specific time. A good example of this and another useful microservice pattern is a health check. Every service supports a health check, which we can poll. We use this approach in our financial exchange at Elmax. We check these health checks on a regular basis, asking for a status report from every service every 30 seconds or so. ⁓ I'd always prefer things like this to work on a pass fail basis.

healthy or not. So each service has a built-in model of what healthy means for it. We made it as hierarchical, so if one service depended on others ⁓ to do part of its work, it would pass the need for a health check onto its dependencies. That meant we could traverse a graph of services and build up a comprehensive view of the overall health of the system. We'd report the cause of health failures in response to these health checks, so that then we could diagnose problems more easily.

In distributed systems like these, establishing a coherent picture is complex for several reasons, time certainly, but also where to look for an answer when tracing. Local logs is no good in terms of ease of access ⁓ or consistency of the overall picture. OK, we got to this service on this server, then it sent this message, but where did it go? To solve this problem, we need another useful pattern, log aggregation.

Instead of each service or component of the system writing log entries locally to a file, we need to establish some central point that collects logs and assembles ⁓ a system wide account of what really happened ⁓ and what order it happened in. There are tools that support this kind of thing very well, but with good PubSub messaging infrastructure, it's really not that difficult to establish your own custom versions with a little thought if you need to, ⁓ which can be very helpful for other...

Sneha Mehra (00:11:58)  
less generic forms of logging. For example, you could log a searchable database of global order history for later market data analysis. ⁓ Time is certainly one problem when attempting to piece together a sequence of distributed events in a microservice system, but another equally important one is context. Is this event related to customer A for order 51 or customer B and order 52?

For this, another useful pattern is that of correlation IDs. A correlation ID is what it sounds like, an ID that allows us to correlate information. US citizens will be very used to using their social security number to uniquely identify themselves in a variety of contexts. This is a correlation ID, so we need to build a model of ideas that define the context that we are interested in. We may even need to build a model of relationships between those IDs

so that we can uniquely identify things that are lower down the hierarchy. One way to think about this in the context of microservice is described in this video. We can model the relationships between things, build a model of keys and their relationships, and use those keys to define separate different conversations in the system. For that to work, we need every message to contain the relevant keys that contextualize the conversation so that we can piece it all together later.

This allows us to maintain a coherent picture across distributed interactions. As long as the services share this context model, we can correlate the work of any number of services, ⁓ even when they are independently developed and deployed. Observability is important to maintaining our ability to monitor our systems in production and learn from them. In microservice and other distributed systems, ⁓ none of this is simple, though. ⁓ Using techniques like these,

help to make it easier to maintain a coherent picture of what is happening and so provide us with a powerful and essential tool to maintain and grow our systems in production.

—-------------------------------

16  
Sneha Mehra (00:00:01)  
The way that I've started thinking about you know ⁓ the definition of quality in modern systems is our ability to change it. Because, you know, I I I I, you know, my engineering stick at the moment where you know I'm trying trying to promote the idea of software development as an engineering discipline, ⁓ applying scientific style thinking to solving problems. ⁓ It seems to me that that starting out knowing that we don't know all the answers yet is really important. Assuming

That what you know, whatever answers we've got are probably wrong and way may change. And maintaining our ability to evolve our thinking and our understanding ⁓ is how we get to these more complicated systems that we that we've been talking about so far today. Yeah, and you mentioned two very important ⁓ words there, right? It's like the the uncertainty, the change, and also the complexity. So I have some some quite quite some thoughts on this one. The first one is.

Both of us come, you mentioned the beginning, right? We come out of sort of the, I don't want to say the birthplace, but probably sort of one of the early places where agile, you know, hit hit the mainstream, right? Yeah. ⁓ Agile development is very much born out of this insight, right? If I know all the answers and nothing ever changes, I don't need to be agile. Just write it down, I build it, you know, I'm done. Yeah. Right. Now that's not the context in which software is being delivered today. So hence we are very fond of agile methods. And what cracks me up is that.

Sometimes people come and say, ⁓ I don't need architecture. Say, you're an architect, that's very nice, but I don't need you because I'm agile. I'm like, that's quite interesting. Because if you're agile, it means you deal with change and uncertainty. Like, yeah, yeah, lots of change, lots of uncertainty. I'm like, perfect. So does architecture, right? Because if you don't have change and you don't have uncertainty, you don't need architecture either. So first big insight is really agility and architecture go hand in hand to deal with the world.

That has high rates of change and high levels of uncertainty. They go together, they're not opposites at all. I often say, yeah, agile is the steering wheel, architecture is the engine, right? One keeps you moving, and the other makes sure you move in the right direction. So that's a common misconception. The other word you said though is the complexity. ⁓ And we said you know, good architecture is making systems ready for change, yes, and this is absolutely true, but there's always a secondary effect.

Sneha Mehra (00:02:25)  
Right. And usually the secondary effect is a negative one. So if you make a system ready for change, it's easy to increase the complexity. Right. So if you want everything to be changeable, either you reinvent the JVM, you have no need to do that, right? You can do anything you want, but you didn't deliver much value, right? So no need to do that. ⁓ Or you drown in complexity. Right. And this is where in our industry it's sort of a ⁓ status like a bragging right to have a law.

So I felt compelled to also have my own law. So I called Gregor's law. And that is ⁓ that is that excessive complexity is nature's punishment for organizations who are unable to make decisions. ⁓ So the game is that yes, you want to defer some decisions, right? You want that ability to change, but that deferring has a price, and that price is largely complexity. So and as architects, we find ourselves right in that middle, like basically.

Which things can I ⁓ lock down that bring my complexity down, but that afford me future ability to change? And simple examples, right? So let's say people writing software and we don't know what languages they're gonna want to use, right? Or they can agree or it's gonna change in the future. Like, okay, I make a services architecture and I make common APIs, right? So if you have common APIs, right, the next service can be written in in whatever language. Quite honestly, though, what?

Two main thoughts of this. On one hand, it's one of those classic maneuvers where taking some choice away gives you more choice, right? Like I I I forced you to make common APIs and I forced you to use ⁓ JSON and OR on HTTP, right? I lock many things down. You know, being sort of the old enterprise architect here, right? I standardize some things, but it doesn't take your choice away. It actually increases your choice because now you can mix and match, you can deploy in the cloud, you can run here, you can run there, different languages, right? So first thing is.

It's a classic maneuver of making some decisions ⁓ enables others, but also to be fair, I increase complexity a little bit, right? Now you need an API layer, you have partial failures, retries, either potency, out and sequence messages. ⁓ you have things like, and Jason, the field is missing. Is that the same as null? Is that the same as empty string? Is that I write ⁓ all the little fun stuff we deal with all the time. Those complexities we just all inject it into our system. So

Sneha Mehra (00:04:51)  
Couple of thoughts to this. On one hand, you know, and as I said earlier, I have a hard time with having a definition for architect. And the reason is I have many definitions. ⁓ So one of the definitions for architects is your option traders, right? Basically, what choices do I take away? Because I gotta take some choices away, otherwise I drown in complexity, Gregor's law. So which choices do I take away? But what choices do I gain ⁓ in ⁓ in return? Right. That's sort of the classic role of the architect. And

The other way of putting this is so we are sort of the flexibility versus complexity balancers, right? It's like how much ability to change can we have and how can I do this without complexity, you know, running or spinning out of control? ⁓ That's another key role that that architects do. Yeah, absolutely. And and again, ⁓ we are kind of in a hundred percent agreement and talk about it in slightly different ways, but

But but but it's interesting. So so my ⁓ my ⁓ one of my foundational ideas in terms of software engineering principles is that ⁓ is is the idea of managing complexity. So so do making those choices that that allow us, give us those freedoms, but as you say, they incur more work. So so I I I have some examples in my book where it it talks some some little code examples and and I point out that

You know, if we're going to go for more modularity, better separation of concerns and so on, let's probably end up with a little bit more code. But it it makes the system as a whole more flexible and you can deal with it in different ways and and so and I'm talking about the trade offs between those, you know, those different ideas. I th ⁓ and I I I'm it certainly in my head as you were describing that, I that's what I was translating it into. I I I think we're talking about the same idea there, this trade off between those those reasons. You I I I I think that's

That's an insight it took me a a a while to get to, but you've stated it really nicely, which is that is that idea of of the constraints are a tool that we use to to give us those freedoms in other you know in other areas. So there are parts that matter more in in it it it seems to me there are parts that matter more in the design of software than than other parts. I don't you know, the the the the the idea of something like modules or APIs, the boundaries between them, are more important places in the code than

Sneha Mehra (00:07:13)  
internal detail of the implementation. The internal detail clearly has to work, but you should be able to change it in a variety of ways. Those more important part places need a different kind of thinking, more architectural thinking, I suppose. Mm-hmm. Yeah, ⁓ absolutely right. So just you know w two two two comments, right? The one thing is because this it it sounds a bit abstract, but it's like really important, right? This is a essence of architecture, but it runs the risk of sort of being labeled as academic again.

So that's why I like sort of real life metaphors. Now I've worked in financial services. So that's why the options theory, right? Like options trading, right? Yeah. Sort of what's the strike price? What's the price of the option? Right. I use that a lot with my customers because that stuff they know from the industry domain. So if you can do that as an architect, make that translation, it really helps get your concepts across because it's not so easy to put it in the right words. And the second part of it is, of course, I'll bring this back to the

architect elevator. So let's say you're faced exactly with one of those decisions, right? I can I can make this more modular so that affords me a bit more changeability, but I pay some complexity. Like which which way is better? You as a technical person, you cannot answer that unless you understand the business and the business strategy and the variability up there. Like you cannot make a good decision about this without understanding the levels above you and have a

Fantastic example, right? Insurance business. We we once wrote a ⁓ sort of tablet, really cool stuff, digital, where people can buy insurance, right? And it was really successful. You put your coverage in, and then it tells you how much your premium would be. Two months later, or ⁓ shortly after, great success. The business comes like, we want to do this in the inverse, right? Because Southeast Asia, people have limited budgets. So people want to put in how much premium they're willing to pay. And you're supposed to tell me what coverage you can get. And people are like, ⁓ it wasn't.

Built that way. ⁓ Right. And I went to the business folks and it's like, well, did you assume that this was like the most ⁓ obvious next thing? And they're like, of course, totally. Right. It's like you either do it that way or that way. It's like the same thing, right? Of course, right? But that hadn't made it to the technical folks. Of course, in the end we like cheated our way out with like some iterating and bisection interpolation, right? You could sort of right. We like sort of you know put some screws on the premise side, right? And we end made it work. But this was a

Sneha Mehra (00:09:37)  
Classic example where the business says, of course, that's the next thing we're gonna do. But the engineering team and they made a really cool calculation engine, right? It wasn't like they ⁓ just hammered it together. They really put software engineering effort in, but without knowing what the variability points are. So they chose a very ⁓ elegant architecture that's solving the wrong problem. And that is without knowing what's going up up here, on up here, you

Cannot make a good technical decision down here because you'd be guessing, ⁓ and you'd be guessing wrong ⁓ many, many times.

—--------------------------

07  
Sneha Mehra (00:00:01)  
Microservices is a powerful and effective approach to large system design, but they aren't simple. And the approach is often widely misunderstood, to the extent that I and most of the experts that I know would probably agree that most teams that claim to be practicing microservices approach ⁓ aren't. I was discussing this with someone recently in the comments to one of my old videos, and I thought it was worth exploring this idea a little bit further. So that's our topic for today.

What's the real value of microservices and why may you be missing the value of your microservices in your microservice design? I've spoken about different aspects of microservices as an approach, their advantages and their challenges several times here before. In the conversation that I was referring to earlier that prompted me to make this video, there were two ideas that I really wanted to discuss. The first is a broad definitional problem. That is...

Do the names that we use for things really matter? Does it matter if what you call a microservice is different to what I call a microservice? I think it does, because unless we can agree on terminology, we can't really communicate effectively. And maybe even more importantly, we can't really learn and make progress as an industry with new ideas. If microservice is just used as a label,

and we can ignore all of the attributes of that microservice that define it. What does it even mean to say that we build a microservice system? If we ignore the definitions, we might as well say we build a what's name system. Without that basic level of understanding and agreement of the definitions, these things make no sense at all. The labels that we apply are arbitrary and useless unless we have that understanding. And yet, increasingly, I see people making what sounds to me

To be exactly that argument, you're being too pedantic in your definition of continuous integration, microservice, TDD, continuous delivery, or whatever else. Well, not really as far as I can see. I'm using the definition that usually the originators of the concept defined. Sure, and here's the tricky part. We're so used to being lax with definitions of ideas like these in the software world ⁓ that there are often lots of often misleadingly different definitions out there.

Sneha Mehra (00:02:21)  
So it is a reasonable criticism of me that I have chosen the definitions that I like to talk to. ⁓ But usually, as I said, I try to use the definitions that were the definitions made by the originators of the idea where I can find it. ⁓ For microservices, I tend to use this list from microservices.io as a shorthand for the fuller description that ⁓ is pretty definitional from Martin Fowler's website written by James Lewis and Martin together.

So microservices then are small, focused on one task, aligned with a bounded context, autonomous, independently deployable, and loosely coupled to other services. This is a pretty widely accepted definition. So if your microservices aren't small, focused on one task, aligned with a bounded context, autonomous, independently deployable, and loosely coupled, then they aren't microservices. Whatever it is that you have ⁓ may be useful, may be even good design.

but please don't call them microservices, it's too confusing. The most important but also most challenging of these attributes in terms of impact on development and being able to scale systems and teams is that they're independently deployable. ⁓ I think that you can make a good case, and I have in the past, that all of these other things are mostly important because they help us to keep the services independently deployable.

Microservices being independently deployable is really the whole game here because it is that that enables us to organise development into many small autonomous teams. And the autonomy matters because that is one of the most important predictors of success from the findings of the Dora Metrics. Smaller autonomous teams build better software faster. All of this may sound great, but independently deployable is actually setting a pretty high bar for design ⁓ and requires a level of design sophistication

that many teams struggle to achieve. Certainly most so-called microservices system that I see don't have services that are independently deployable. That is services that can be deployed without testing them alongside every other surface before release. ⁓ I've spoken about that before. ⁓ Another aspect of this design problem though is something that cropped up in my discussion with that viewer in the comments. ⁓ He said,

Sneha Mehra (00:04:43)  
the vast majority of microservices that I have seen actually share a common database and data model behind the scenes. I agree that this is a common pattern and a problem, but once again, not microservices because they aren't autonomous. We can't change one independently of the other. Or at least there are some kinds of change that we can't make without impacting both. We can't change the schema for the storage in one service and not upgrade and deploy the other service.

if they both depend on the same schema and the same storage. My interlocutor gave an example, processing an order for a customer of where this may pose some design challenges. ⁓ We could imagine two services, order processing and customer details or something similar. The lure here for more traditional thinkers, more traditional meaning those more used to normalize databases would be to build something like this so that when an order for something came in for a customer,

the shared data store links the order to the correct customer account record. But this is why the bounded context alignment matters here. Fundamentally, there are some fairly strong limits to the scalability of normalised data. ⁓ Data normalisation is good in that it means that there's one version of any fact, but it has a downside is that the interactions with those facts represent a form of coupling.

So often distributed systems where the data is normalized are much more difficult to change. ⁓ I spoke about a related issue to this in this episode. ⁓ By joining our services via their storage and data representation, we are leaking information here and increasing the coupling between them and so making them harder to change. Microservice is a distributed systems model. And in distributed systems, the problems of data synchronization

can be very complicated. So it pays to avoid them as far as we can. And that is, to a large extent, what the advice about align with bounded context and no shared data is all about. If we restrict ourselves to only sharing information via the messages that travel between the services, this has a lot of advantages. It means that the conversations are clearly and well-defined. There are no back doors. All interaction is through the APIs of the services.

Sneha Mehra (00:07:08)  
but it also makes more clear what the problem that we need to solve is. We need to define a conversation, a protocol of interaction between the services that copes with the cases that we're interested in. Being lazy and falling back on the now oversimplified model of normalised data to keep our services in step with one another, we're breaking the microservice model in a fairly profound way. We've coupled two distinct bounded contexts together for technical reasons. That's a bad idea.

So now the code for both of these services is harder to work on, harder to test, harder to maintain. ⁓ And if it is held in separate repos, it's also slow and annoying to change because we've got to be jumping between different repositories every time we want to do something. Completely the opposite of the advantages that microservices are meant to offer. ⁓ I'd say that this design approach is poorly abstracted at the service level. ⁓ Even though at a surface level, the use of normalized data

makes it easier to get started, it will be much more difficult to maintain in future. This is a common problem with design, I think. The naively tactical solution often feels easier at the start. While modeling the real problem that you're interested in and trying to solve takes a little bit more thought. Let's imagine some of the operations that we'd like these services to support. The customer detail service will need to be able to create and register new customers of some form.

and hold their relative contact details, perhaps. Account ID, name, address, email, perhaps a phone number. What aspects of this information matter in the context of placing an order, though? Almost none of them. All we need to know if we're placing an order is which customer is it for. So all we need is the account ID. The most loosely coupled relationship between the customer details ⁓ and an order is the account ID.

I talk about three levels of modeling for services like these in this video. So our order processing service only needs to store the account ID for an order. If it or some other service needs to access more detail of the customer, like the customer's email address, perhaps to tell it when the order's been dispatched, it should ask the customer detail service to give us those details at the point when we need them. It doesn't need to have access to them directly itself.

Sneha Mehra (00:09:32)  
We may be concerned that our very simple approach to order processing means now that people can place orders for non-existent accounts. And we don't really want to clutter the system with useless orders this way. Well, we could solve this problem by getting the customer detail service to notify the order processing service with new customers when they're added. So now the order processing service can maintain its own list of valid account IDs and check that the account ID attached to an order

is in that list, even though the data is now not normalised anymore. There are duplicate lists of account IDs in both services. This isn't a bad thing though. This is a common strategy for maintaining the relationships between data held in different services. Yes, you have to do a little bit more housekeeping work to keep things up to date and tidy, but the result is a significantly decoupling between the services, which makes this whole system easier to work on. For example,

I could write a test of the order processing service without the need of a customer service detail service. Because all I would need to do for my test is to notify the order processing service that account 77 was added and then place an order for account 77\. The big difference I see here is that instead of relying on generic technical levels of integration between the parts of our system, instead we make these conversations much more explicit. We abstract them at the level of the business problem.

that we're trying to solve, not at the level of the technologies that we're using. This means that the conversations are simpler. And if we follow the good advice to always translate information that transits between bounded contexts, these conversations are now dramatically more loosely coupled and more explicit, an inherent part of the design of the system that makes sense to everyone, not just to the technologists on the team.

These conversations become important integration points in the system, more obvious and so much more manageable as a result, more defensible in the face of change. This also makes things lots easier when we want to add new features because the conversations that are already supported are now explicit, clearly defined by the APIs to the services and make much more obvious sense in the context of any given service. Let's imagine adding a customer accounting service to our design.

Sneha Mehra (00:11:58)  
It listens for customers being added and also for orders being added. And it stores the list of orders for a customer. We could imagine the service keeping a running total of the amount spent by the customer on each order and maybe sending out a notification when the customer spent over a certain threshold amount that meant that they now got ⁓ into a more privileged account status or something. The real trick here is that these conversations are all taking place

at the level of abstraction that represents the problem domain, not the technicalities. We are in the language of placing orders and creating customers rather than of adding records and defining select statements. ⁓ This is a much healthier level of coupling and I recommend it to you.

—------------------------

12  
Sneha Mehra (00:00:00)  
For me, I remember it as the end of that. ⁓ It's only ever happened once this time, one time in my life. ⁓ You know, people talk about having like eureka moments. You know, clappically in the bath. Well, I wasn't in the bath. ⁓ I was in a hotel room in Casa Grande. ⁓ And I literally, was lying in bed thinking about the course of the day. honestly, ⁓ it sounds bizarre, but it was like, there was like a bing, like almost like a glass shattering moment.

And I suddenly just, ⁓ what if we're doing it all wrong? ⁓ What if we build, ⁓ rather than solve the problem of evolving the big things, why don't we solve that and build smaller things? And I remember running down to breakfast the next morning and Jimmy Nielsen was there with his lovely wife and the kids were there. ⁓ And I went running, they were literally just leaving. ran, Jimmy, Jimmy, Jimmy, Jimmy, Jimmy, Jimmy. ⁓ I was just sort of blabbering at Jimmy. And he was looking at me in his incredibly calm, ⁓ slightly critical kind of manner at this new one.

Welshman suddenly running up to him, I've got it, I've got this idea. ⁓ And really, it sort of came from there for me. mean, obviously, as I say, as I've said many times, there was sort of parallel evolution, conversion evolution of these things. So, ⁓ you know, we had Stan North talking about replaceable component architectures. Fred George was in his developer and programmer anarchy stage, and he was talking about all these tiny services. ⁓

We have Adrian Cockroft, obviously, at Netflix. He was talking about fine-grained service-oriented architectures at scale. They were building stuff at scale at Netflix at the time. ⁓ And so I think it was just a set of ideas who used to come, if you like, to be honest. ⁓ And of course, everyone knows each other anyway. ⁓ Daniel's excellent work, ⁓ Brad George's excellent works. ⁓ Adrian Cockroft is big mates with Jim and Ian from integration. ⁓

from circles. ⁓ there's just a group of people whose ideas are circulating. ⁓ And I think the time had really arrived when it was possible to do all the things. ⁓ So ⁓ I talk about XP being, well, ⁓ XP is doing all the things and turning it up to 11\. I think microservices is similar to that. And without being able to do things like ⁓ spin up environments automatically,

Sneha Mehra (00:02:26)  
which you can do with the cloud, because obviously AWS was starting to become, ⁓ moving into its ascendancy. ⁓ I don't think, I think the complexity which you push into the infrastructure wouldn't have been worth it until ⁓ about that point. So ⁓ I'm sure many other people have lots of these ideas beforehand. You could argue that Alan Kay, you when he talked to us, talked about object orientation as, what did he say, message, I'm paraphrasing message passing.

⁓ encapsulation ⁓ and extremely. I did something in the early 90s. I worked for a startup. ⁓ was a guy that worked in IBM and he came up with a concept for what he called cooperative business objects. ⁓ Yeah. I ⁓ worked in this startup that built the infrastructure to support these cooperative business objects. So essentially, to a very large extent,

⁓ separating the essential complexity of the system from the accidental complexity and you built the system from a series of these things. We got a bunch of things right. We kind of we invented XML before XML was a thing ⁓ in order to make it to make that ⁓ yeah the kind of communication so you can have flexible communications and it did some really cool things. ⁓ I think you know ⁓ we also got a bunch of things wrong but ⁓

But ⁓ we had a programming model for building systems in COBOL, very, very object oriented systems in COBOL, ⁓ which was interesting. ⁓ But there's a bunch of other languages too, but ⁓ it ⁓ was really fascinating. And ⁓ that got me really interested in kind of service-based design ideas then, I think. ⁓ One of the things that's always surprised me

is that I can't think of any programming language or system that really surfaces the idea of a service. And that seems weird because it seems like such a broad natural division of responsibility in design. ⁓ every time I've ever done it, and that's nearly every system that I've built since the 1990s,

Sneha Mehra (00:04:47)  
It's, ⁓ you know, you have to come up with your own conventions and protocols for what you mean by the edge of a service. And that seems that that seems a bit strange. ⁓ It does. guess you could argue that Erlang came came closest. Yeah, ⁓ that's true. ⁓ Which is, essentially probably you'd argue them the most true to Alan Kay's idea of object orientation. ⁓ I ⁓ think Daniel to us North again, he's going to quote about ⁓

one of his quotes, it's the Greetings Ponds 10th Law, where it's the ⁓ Greetings Ponds 10th Law is any sufficiently complicated, paraphrase, sufficiently complicated, oxygen-oriented system will contain ⁓ a ⁓ half-implemented, bug-ridden version of half of Common Lisp, right? That's the... And I think you replace those with microservices and ⁓ Erlang, right? So sufficiently complicated microservices implementation ⁓ contains a half-arsed bug-ridden implementation of half of Erlang.

⁓ Exactly. I ⁓ I fully agree. When I was running for it, I went to several investment banks and it was quite natural when you're sitting working with traders to build lots of small things. ⁓ Because you haven't got time not to build lots of small things. Do ⁓ you mind if I just take a week because yes, you can have that, but I'm going to take a week to change this first. You can't do that if you've got an opportunity in three days time. ⁓

Yeah, and then Martin came to visit a project I was working on. ⁓ And being Martin did his, this seems like an interesting idea. You should write an article for my website. Blah, blah, blah, blah, blah, blah. ⁓ That was in 2012, maybe 2011\. ⁓ And then it took me three years. ⁓ he invited me to a sleepover at his house in Boston. ⁓ we eventually wrote it up in 2014\. ⁓

So he kidnapped you to write the article? ⁓ Basically, it's the only way. ⁓ I was laughing about this with ⁓ Eric's in the thing, obviously, he'd be sure a friend of us. ⁓ And I saying, ⁓ it's terrible. ⁓ I I find writing difficult. Really, I enjoy it. find it, it's starting to get, and he said, well, you know, it turns out you're quite good at communicating in other ways. You're quite good on stage. You're quite good at interviewing people and things like that. ⁓ So I guess we don't all have to be the same, do we? No.

Sneha Mehra (00:07:14)  
Probably good if we're not, because everybody would have a YouTube channel. ⁓ Yeah, ⁓ I'd just like to say I'm launching my own YouTube ⁓ I'm actually on Twitch every night, Dave. That's why. ⁓ So semantic diffusion is a thing. Everybody that I know of that's come up with an idea that's become popular, ⁓ kind of...

at least sub-vocally says, I didn't mean that. ⁓ So what things do you wish that people more understood about microservices? Or that they missed?

Sneha Mehra (00:08:00)  
That's a really good question. So I came up with this really catchy term which ⁓ never caught on, right? ⁓ Not microservices, that was a group effort. And then when it appeared on Martin's site, that's it, right? You're done, because it stays there. But no, I came up with this idea of, I think it was called the microservices integration pattern ⁓ organizational onion.

Which as I say, not a name. ⁓ It's not quite as catchy, is it? ⁓ It's not. ⁓ And it was when I was talking about commerce law a lot, did a series of talks on commerce law, ⁓ why I need to stop worrying and embrace commerce law and so on. And one of the things I talked about in that is much like, I see functions and methods, ⁓ objects, ⁓ service, microservices.

and then collections of these things as it's like you start at the macro and then as you get deeper and deeper, it's kind of fractal. The same structure repeats itself and repeats itself until you get down to the methods and the plasso of it. ⁓ So I sort see it as the turtles all the way down until you hit the VM. ⁓ That's what it's all sitting on, the OS. ⁓ And in my mind, when you sort of chunk up from the smaller things to the bigger things,

You chunk up from the methods of the class to the library or the namespace to ⁓ a service, a service into groups of collaborative services. ⁓

In my mind, the whole thing was about how teams are organized around those as well. ⁓ So there is a similar structure, ⁓ a similar fractal nature to the organizations building these things. ⁓ So as you chunk up to large organizational units, if it's a team of people who are managing a number of services, ⁓ maybe whatever that's, five, six, whatever it is, ⁓ that's, from the outside, ⁓ is opaque. You don't need to know. ⁓

Sneha Mehra (00:10:05)  
And as you have these groups of teams collaborating via these external service interfaces, I think ⁓ the trendy phrase is, know, ⁓ kind of hard on the outside, soft in the middle, right? So, you know, the ⁓ main boundaries are very clear defined by, as you were saying, very clear application protocols that we don't have to invent every time. We use standard ones like the web ones, W3C standards ⁓ to talk to one another. And then those teams are then jumped up again into

a number of teams implementing a business capability. ⁓ So I think the idea of the business capability was the thing that we didn't quite get across. So I think people say business capability is synonymous with a boundary context or synonymous with some of the domain design. ⁓ When they're not, ⁓ business capabilities tend to be very big things to first order, ⁓ zero thought of that massive, but first order, ⁓ they're sort of warehouse management. There's a lot going on in warehouse management.

And only some of it is software. That's the key thing. There's people, there's the processes, there's people following, there's the tools that they're ⁓ using to do their work. ⁓ so if you're talking about microservices organized around business capabilities, there's a set of microservices organized around house management, right? ⁓ Which are some of those tools. ⁓ And I think that's led, I don't know if it's definitely correlation.

between a misunderstanding of that and this fractal nature and how they should be completely decoupled across these big capabilities and the fractal nature of the inside to this microservice soup thing, anti-pattern, which you see a lot of, which is, ⁓ well, hey, I've got 2,000 microservices and I have to deploy them all at once. Microservices are rubbish. That is why ⁓ we never said do that. ⁓ 2,000 microservices and 10 developers. ⁓

10 developers, 2,000 Microsoft engineers, and we can only deploy them every two years because we've got eight months of progress. You're ⁓ not winning at anything if you've ended up there. So I think that is one of the things ⁓ I wish we'd spent more time talking about.

—--------------------------

11  
Sneha Mehra (00:00:09)  
One of the things that I've thought about ⁓ I think about microservices and one of the anti patterns that I see with teams professing to adopt microservices is missing that

distributed nature of development, not just the system. So so so we are distributing development as t distributing responsibility and decision making as well as as the technical components of the system. And and that's really for the to not to my mind, ⁓ I don't know whether you would agree, but but th the va you know one of the key values of that ⁓ is that scalability of development that you get as a result. By by making these things more autonomous, more responsible for every aspect.

they have more freedom of movement, more freedom of action, they are able to respond to change much more effectively. ⁓ For for us that has certainly been the case. And I'm not here to argue that microservices is the the perfect architecture for every use case, but for us, sure. ⁓ Spotify has been lucky enough to grow very quickly f throughout our lifetime. And for us it's been it's been a very powerful

architecture and set of practices for that scalability and for maintaining the distributed action that the distributed aspect that you're mentioning. Of course, I should be clear, like the it's not that every team build things completely willy-nilly. Like it it needs to fit into a larger, the larger company strategy and the larger product strategy. So there's some alight some degree of alignment that is happening as well. But ⁓ that like all of those day-to-day decisions that you take as part of building and ⁓

evolving and maintaining your systems that we've been able to ⁓ like keep pretty local in the team and that I think is very powerful. And we've been able to see the flip side of this because we have, as I said, we made our mistakes along the along the way, of course. And we've definitely ended up with cases where we built accidentally or not more monolithic ⁓ architectures. We had our mobile apps be very heavily monolithic. We had

Sneha Mehra (00:02:14)  
parts of our back end be pretty monolithic and we've gone through the the pains of that and seen like, okay, so now we need to do a ton of coordination between ⁓ many different teams to ship our software and like we can see how painful that is compared to where we've been able to break things down into more of the microservices type of pattern. So ⁓ yeah, for us it's been a great model. ⁓ And ⁓ but as as I think you're you're kind of hinting at like any other engineering dis

decision that there are kind of trade-offs so so so so a ⁓ and ⁓ and things that you have to manage. I I assume that when you're talking about fleet management that's part of what you're talking about is is okay, you've got this ⁓ distributed organisation and technical infrastructure of small pieces. How do you kind of point those pieces in a in a consist in a coherent direction from a business point of view and from an organisation wide point of view?

Yeah, and and how do you so that's part of it. So we've really thought about this in a few different ⁓ related highly related strategies, but it's really down to ⁓ what we like standardizing your technology stacks. So we within Spotify this is called Golden Technologies. So I can talk more about that if we want to like reduce the amount of ⁓ fragmentation that we have in our technology ecosystem, and then that coupled with

speed management where we started this discussion on really trying to automate as much of the management of or ma maintenance of those components as possible. Yeah. But I I might be misremembering, so do correct me if I'm wrong, but but my recollection of you know part one of the things that made Spotify ⁓ visible to lots of people was the stuff that that that that you you guys published

years ago with Henrik Nijberg's ⁓ animations about your development process and so on. Mo one of my recollections from that and one one of one of my recollections of the takes from that is that is that the technical choices in those forms what language you're gonna program a particular piece in was was a choice of the team at the time. ⁓ That was not necessarily the case. So we've always had some level of ⁓

Sneha Mehra (00:04:37)  
guidelines and guardrails. So if you go down to the to the programming language, for example, that's some something that we've always been pretty strongly opinionated around. We've had a set of languages that we that we want our teams to use. And I would say over time we've introduced more and more of those like we've move lot s slowly moved our way up the stack in terms of where we want to standardize across the company. There's still

You can make except exceptions for that where there's a a good reason for it, but then at least we want there to be a a healthy discussion around that decision and and and have some accountability for the decision. But ⁓ yeah, so so it's not yeah, there's a fair amount of of of guidelines and and the reason for that is really to have both nowadays for the automation ⁓ reasons, and I can talk more about that, but also to have just have fluidity in the organizations we can move

ownership of components around in the org, we can change the org structure without, you know, ⁓ some team all of a sudden taking on something that is written in a in the language that they might not be familiar with or or whatnot. And just being able to move like have people ⁓ have mobility within the organization so they can switch a team and they don't have to completely relearn the stack. So ⁓ I think that's been mostly a good set choice for us, but it's definitely a trade off of course. Like it's it's

Yeah, I yeah, I I I I'm ⁓ I I I would ⁓ I I would I would agree with with with with with that from from my background, which you know i is largely from ⁓ not on your scale, but certainly in the in the realm of larger systems and larger teams ⁓ and and so, you know, that flexibility of having the ability for people and pieces to to move around the organisation I think is a valuable one to have. ⁓ What w w w where would you see the trade offs between

between you know you know the the full autonomy of the every decision is made within the team ⁓ and autonomy within some constraints, within some I think you use the term guide rails. ⁓ I would use the term guide rails anyway. So having some guide rails for for that kind of decision making I ⁓ I must admit I'm probably more in line with the guide rails than the full autonomy personally. Yeah, I I think I am as well, just to be make my biases clear well. Yeah. ⁓ so ⁓ I would

Sneha Mehra (00:07:01)  
I don't know if this is a good classification, but like I would maybe think about it as you can have the accidental fragmentation where you make decisions based on, you know, some individual's preference or ⁓ experience from previous work life or whatever that are not necessarily a better choice from a company strategy or like supporting the actual problem that you're trying to solve. There's in my experience lots of that will go on unless you put some some type of guard race in place.

And then you can have the intentional fragmentation where so let me take an example from our world of ⁓ when you write the backend service, most of Spotify's backend services are Java based since 10 years or so ago. ⁓ And there are of course use specific cases or edge cases where Java might not be the best programming language for a specific type of problem that we're faced with. And then we have to make a trade off between ⁓ the ⁓

Needs of that specific problem and that specific domain versus the type of gains we can get from fleet management automation, for example. Because ⁓ we all of a sudden build a service in Rust because Rust is amazing to do ⁓ transcoding or whatever it might be. That might be great for solving that problem, but that team is now stuck with maintaining a solution that is poorly supported by our by our infrastructure teams.

will not get any of the automation we have in place because we do not support that ⁓ that stack internally. So those are the types of trade-offs that we that we're faced with. And I wouldn't say they're super common. I think we've picked a set of technologies that fit, you know, some ⁓ nine to ten type of of of thinking of by far most of the use cases we have should fit within the technology stacks that we standardize on. But we also have

We also acknowledge that there will be some percentage that will not. So if we go back and look at our I mentioned as Golden Technologies, which has been our program for driving a higher degree of standardization, the long-term target we have for that is to reach 85% ⁓ adoption. ⁓ And that is exactly to retain a certain buffer where we have these edge cases and we have a certain amount of experimentation and innovation going on so we can try out new technologies and whatnot. So ⁓

Sneha Mehra (00:09:27)  
Yeah, I I we're not striving to get to get to a hundred percent because we actually think that would be counterproductive to what we want. Yeah, I I I think that's an excellent point and I I ⁓ I ⁓ organisationally, ⁓ part of the freedom that you want. If somebody invents the world's greatest programming language ever ever, you want somebody in your organisation to be trying it out and using it so that you can find out how good it is and and start adopting it. So you need that room for innovation and discovery as as a part of your

strategy ⁓ I I would assume. Yeah. So going back to the autonomy, like the the autonomy that we want our teams to have is really around what's your product strategy and aligning that with our overall strategy, of course. And and within that, like what are the software components you need to solve that problem? More so than reinventing the infrastructure that those components ends up being built on. Like that's we've standardized on the on the infrastructure level

but left a fair amount of flexibility on the application level. So that's roughly ⁓ the thinking. ⁓ So I ⁓ I would put this in terms of ⁓ an ⁓ an architect's or an engineer's perspective on on these sorts of things, it has to be framed by the economic value of the choices that we make.

You know, it's it would be dumb to you know i if if if every developer was picking their own development language ⁓ to pad their CVs, that would be a bad economic choice from Spotify's point of view. But exactly so the the so these some of these convers some of these discuss ⁓ decisions need to be more strategic than that. ⁓ Yep, that is absolutely right. And and Spotify, like many companies nowadays, also have a shared platform. We have a platform team that builds that internal platform for us on top of our cloud.

⁓ environment and we want to that is expensive, that is a big investment for us and of course we want to be able to deduplicate those investments as much as we can. So the the fewer the fewer technology stacks we need to support internally the better off we are. So but again it's a trade-off around those ⁓ against those ⁓ local domain ⁓ based needs that we have

—--------------------

9  
Sneha Mehra (00:00:01)  
If we'd want to deliver better software faster into production, that relies on us working in ways that allow us to know that our software is indeed better, but we also need to do that efficiently because we want to go faster. That's pretty obvious, I suppose, but it has some important consequences. These consequences, for example, rule out some ways of working that don't give us that fast, accurate, easy to achieve insight into the correctness of our work.

The only way to be sure that we are being accurate is to evaluate exactly what we will release into production. Anything else? And we're really just crossing our fingers for luck and guessing. Sometimes our guesses may be good ones, but they're still only guesses. So if we want to do better than guessing, we need to evaluate releasable units of software. ⁓ Ideally, exactly the sequence of bytes that represents our system. And if all is good.

We can release that exact sequence of bytes into production with no more effort or worry. Because we know it works because we've already tried it. This is what continuous integration and continuous delivery are all about. Evaluate the stuff that you will release and do it fast. But what does that mean if our system is composed of lots of pieces developed by different groups of people?

How can we evaluate these pieces to a level where we can be confident that they work together without testing them altogether? This is the microservice problem. This is the job of contract testing, and that's our topic for today. If we aim to test what we release to the point where it is releasable without any more work, which is really what defines a deployment pipeline in continuous delivery.

Then the correct scope for a deployment pipeline must be an independently deployable unit of software. If our system is big, we'll want it to be built by lots of people in separate small teams because that's what works best. But now we have a big problem, and there are only two real solutions to it, both a bit tricky. And one very common big mistake that seems less tricky until you think about it carefully.

Sneha Mehra (00:02:16)  
Continuous delivery is working in a way so that our software is determined to be releasable at least once per day. The reason that small teams are important is that they can work more independently of one another, in parallel with one another, so and so make far faster progress overall. It doesn't really help us much to break our development organization into lots of small teams so that we can move faster and then organise work so that the only way that they can proceed is in lockstep with each other.

And so move slower. Often many teams combined move f move slower than a single team would if they worked alone. ⁓ So here is our problem. How can our small teams be decoupled from one another if one team can't tell if its code is releasable until it's been tested with code from other teams? ⁓ And how can we get fast definitive feedback at least once per day on the releasability of our changes?

The starting assumption that nearly everyone makes ⁓ is that you must break the system into small parts, allowing separate small teams to own the separate small parts. This takes a variety of forms. The commonest, most traditional form is to divide by expertise. Let's have business teams, analysis teams, architecture teams, UIUX teams, dev teams, test teams, ops teams, and so on. This doesn't work well at all because we're all working in lockstep now.

Whatever else is going on, this is a classic waterfall. Every feature needs to engage nearly everyone. So every feature comes with a big overhead. This is a highly coupled approach to development. ⁓ The other common response is to have small teams and organize them by components, modules, or services of the overall system. The devil though is in the detail here. ⁓ It really is down to how coupled these modules are to one another. ⁓

How we manage that coupling is key to how we can scale development with small teams. ⁓ If part A depends on part B and part B depends on part C and D, whichever team we're in, how can we tell if our change is releasable? The common misstep here ⁓ is so common, in fact, that it's probably the commonest way to organize teams these days, as far as I can tell, is to structure things so that each small team

Sneha Mehra (00:04:41)  
works in isolation, evaluating their changes locally, but that doesn't give them any insight into the validity of their changes in the context of the whole system. If I am working on team B ⁓ and change something, ⁓ I may break team A's work. Worse, if I'm working on team A and I change something that propagates down the chain and breaks C or D, then my change isn't releasable. But nothing local tells me about that

And I won't find out until somewhere everything's tested together. Even then, once the problem is detected, I now need to have access to see it, recognize it, and be in a position to fix it. Worst of all worlds is that now what I do in response to identifying the problem ⁓ is to raise a ticket and wait for the team responsible for the code that I broke to fix the problem that I forced on them with my change. This is another highly coupled approach to development.

So now the whole process is moving forward really slowly. In continuous delivery terms, this is a monolithic system, however technically componentized it might be. Because we can't release it without testing all of the pieces together first. ⁓ I talk about the strategies that work best in more detail in this video. ⁓ Now, despite common misconceptions, there's nothing much wrong with this kind of monolith.

It all depends on how you organize things and ultimately on the scope of evaluation. You can't have your microservices cake here and eat it too, though. You need to test independently deployable units of software. Without that, this monolithic approach is a very inefficient way for us to organise our work. Because not only have I got to evaluate everything together to get results at least once per day, to see if I broke anything anywhere else.

But I also need access to everyone else's repos so that I can fix the problems that I introduce, raise those damn tickets and wait for the end of time, ⁓ or for all of the pieces to work together, whichever comes sooner. So we don't want to do that. So what does work? Well we can either suck it up ⁓ and treat the system as what it really is, a monolith, ⁓ and evaluate everything together all of the time.

Sneha Mehra (00:07:06)  
We build and optimize a single fast efficient deployment pipeline that can determine the releasability of our system quickly enough to sustain continuous delivery ⁓ and use a shared repository and continuous integration for everyone's work and evaluate everything together after every commit. Actually, this works surprisingly well and is surprisingly scalable. ⁓ Or, getting back to the main topic for today, ⁓ we allow each team to determine the releasability of their changes.

Independently of one another without testing them with everyone else's changes before release. That's it. Those are the two choices that you have. You can mix and match these two approaches, but fundamentally that's all that really works. ⁓ And importantly, to my mind, the ideas at the root of both of these is ⁓ about the scope of version control ⁓ and evaluation of releasability.

We need to scope these things to independently deployable units of software. ⁓ Only then can our deployment pipelines be definitive for the deployment as they should be. The second option of breaking the whole system into small, focused, independently deployable units ⁓ is the microservices strategy. And technically, your ability to do this is dependent really on two things: good modular design with clear, well-defined interfaces between the pieces.

Which I talk about more in more detail in this video, ⁓ and contract testing, where we test our assumptions of those interfaces between the pieces. ⁓ I've been speaking about contract testing as an approach for a long time now, and I generally recommend PACT as a tool to help ⁓ with contract testing, and it's good. But I recently saw a presentation introducing a nice-looking alternative called Specmatic.

From an old friend, Naresh Jane. Maybe I should point out that I'm not being sponsored to say this, but I did get a t-shirt when I spoke at the conference that Naresh organized. ⁓ I've also not used this in a real production project so far, but I do like the look of it and I will try it myself given the need. ⁓ I think Specmatic is addressing some real problems in an interesting way ⁓ that seems to me should work. Naresh pr presented this model.

Sneha Mehra (00:09:31)  
in his conference talk, which starts with ⁓ in the obvious place, I suppose, by designing the API for the service. We agree the contract between producers and consumers and specify this as a separate thing, using some form of interface description language or specification language. There are lots of off-the-shelf open source approaches to capturing the contract between the pieces. ⁓

Most of Specmatics examples are based on OpenAPI, ⁓ a YAML-based specification language. But also something that I liked was that you can define your own interface description language if the need arises, if you're doing something unusual perhaps. This is the contract which we now store ⁓ in a central contract repository. ⁓ Now we can compare different versions of the contract with each other.

To determine their compatibility using only the tools provided. We don't need to write any code ⁓ or extra tests for that. Specmatic does that for us. ⁓ My impression is that this is rather analogous to verifying types in a typed language, but at the level of service APIs. We specify the type contract and then verify that the service ⁓ is of the type defined or talking to the type defined by the contract.

Specmatic does what they call contract-to-contract compatibility tests. ⁓ So if service A is using contract version 2 and service B is using contract version 3, the tools can verify if they can successfully communicate with one another ⁓ by running these checks and verifying that the contracts are compatible with one another. ⁓ You could run these checks in your deployment pipeline.

And you can then prevent merges of any broken contracts, non-backwards compatible versions of the contracts. So if I change the version that I'm publishing, I can find out at build time by checking the backward compatibility of my change with everybody else's use of previous versions of the contract if I want to. It's a little bit more complicated than this, but I think that thinking of these contracts explicitly as being similar to types ⁓ may help. Let's think of a couple of scenarios.

Sneha Mehra (00:11:52)  
I'd like to know if my change breaks any consumers of my service. So at commit time in my deployment pipeline, ⁓ I can run a test based on the Specmatic tools to see if the new version of the contract that I've just changed ⁓ is backward compatible with the previous version. If not, the pipeline rejects the change. Now I'm forced to think again ⁓ and decide what to do next. I could choose to make the contract backwards compatible.

The Specmatic documentation offers some decent advice describing the kinds of changes that are backward compatible and the kinds that are not to help you to do this. ⁓ Or I could decide to add the new braking change and support this new interface in parallel with the old, perhaps. And communicate to consumers of the old version ⁓ that I'd like them to upgrade when it's convenient for them to do so. ⁓ If one of my consumers has a suggestion.

For how they'd like the contract to change, they could create the new version of the contract, validate that it's backwards compatible, and maybe even start work on their use of the new version of the contract, separate from me. They won't break anything, their code will keep working. ⁓ Even as they add new features to take advantage of the proposed new contract, the risk that they're consciously taking at this point is still only completely in their hands.

If the service provider doesn't agree later to add the changes to support the new version of the contract, still nothing breaks, but the team that made the took the risk ⁓ may have wasted some of their time because they decided to take that take on that risk. I like this kind of decoupling between teams. It gives teams better opportunities to make progress more independently of one another. Each team can work to their own priorities independently of all of the others.

And yes, there may be some costs if predictions don't work out. But that's always true. And actually the consequences are much more serious if the teams grind to a halt because they're locked together. ⁓ The big idea here is that we are use shared specifications for the contracts. ⁓ This is a modern take on an old idea, to be honest. Corber Iddl and DCOM interfaces, for example, did much the same thing, and before then probably DCE.

Sneha Mehra (00:14:19)  
Actually, an interface in a typed language is also, as I said, really the same kind of thing. Specmatic adds a new take on this though, and as a result, it's able to offer what they claim ⁓ is a zero-code approach to contract testing for many ⁓ even ⁓ most use cases. ⁓ As long as you specify the contracts in a supported IDL of some kind ⁓ and use these.

version controlled definitions as shared versions of the contract. They can automatically check compatibility, generate contract tests, and generate contract testing stubs that will simulate an implementation of the contract so you can test your service against it. This works as a kind of smart muck and allows you to pre-define your expectations of the of the the contract. I'm usually a bit wary of the use of auto-generated tests, but in this case

For this task, given that you have a clear specification of your contracts defined in software somewhere, then validating adherence to those contracts ⁓ is a significantly simpler task than general functional testing. And so this all seems completely amenable to the use of ⁓ automation to me. This does depend on defining the contracts and sticking to them. No sneaky back doors, no sharing data via data stores.

But hopefully you'd never do that anyway. The documentation for Specmatic mostly uses OpenAPI for the contract, as I mentioned. But I'm told that you can also write plugins to support your own form of contracts if you're doing something more unusual. I'm not entirely sure about the reality of that, to be honest. All of the examples that I've read so far pretty much assume the use of web protocols.

Which is what most people use. So I don't know how far the tools go to help you if you're doing something really esoteric. Again, as I said, I haven't found time to apply this for myself yet. And I don't usually recommend ideas or technology that I haven't tried personally. But this approach looks really intriguing to me. I can imagine ⁓ this going further too, opening the door to

Sneha Mehra (00:16:41)  
The specmatic developers are adding more support for other patterns of use, ⁓ in helping us facilitate the coordination of work between teams without increasing the coupling between them. Maybe some stuff that helps to manage the parallel use of deprecated APIs, for example. I do want to try this for myself, but I didn't want to wait for me to do that before mentioning this intriguing idea here.

—--------------------------------

04  
Sneha Mehra (00:00:01)  
Microservices seem to be what we in the UK call Marmite. Marmite is a strongly flavoured spread that's mostly eaten with toast, that's almost defined by the fact that people either love it or hate it. There is no middle ground. Some people seem to think that microservices is the only way to organise systems that they build, and others think that microservices are a huge mistake that's overtaken the industry and never really were.

I that both groups are probably wrong. The trouble is that in software development though, there are always trade-offs. So what are the pros and cons of microservices and how do we manage the technical debt that really underpins them? That's our topic for today. I recently watched an interesting discussion on the Neat Code IO YouTube channel, talking about microservices with guest Matt Ranney from DoorDash.

The conversation included lots of interesting and I thought insightful points, but I thought that the title on microservices of technical debt was kind of interesting, but sort of also wrong and right at the same time. ⁓ So I'd like to explore this idea in a bit more detail. At first, it may seem like a crazy idea that an architectural approach embodies technical debt at its core, but I think that this is a reasonably fair description. As Matt

correctly in my view, points out microservices are really a socio-technical strategy more than a technical one and one that's often deeply misunderstood. What this really means is that their real value ⁓ is as a technical solution to a social problem or is that as a social solution to a technical problem? Fundamentally microservices are more about team organization and dynamics than they are about software architecture and design.

But then again, we can probably say that about most design choices. ⁓ We have learned that software at any scale is best built by small teams, because without that, the cost of communication between the teams ⁓ overwhelms any productivity gains that we get by adding more people to the development process. Actually, we've known this for a very long time, since at least 1970, when Fred Brooks talked about it. But we regularly seem to forget it anyway.

Sneha Mehra (00:02:23)  
The excellent book, Team Topologies, recommends a maximum team size of only eight people. And the other excellent book, Accelerate, says that one of the main predictors, the defining characteristics of excellence in software development, ⁓ is the autonomy of those teams. ⁓ So if we need small autonomous teams, what's the impact of that on our design choices?

Fundamentally, we need to manage the complexity of the systems that we build so that each team is able to focus on their own work, unfettered as far as we can manage it by the work of other teams. This is a lot easier said than done, but this is the core, the heart of microservices as a strategy. In general terms, I talk about the importance of managing complexity through the use of modularity, cohesion, separation of concerns, abstraction.

and managed coupling. One take on all of this, but not the only one, is microservices. We compartmentalise our systems into coherent, sensibly bounded parts with clearly defined interfaces between those parts, so that we can make a change in one part of the system without forcing change on any other part. ⁓ This, and only this really, is what microservices are really there for. There are other advantages.

But all of those other advantages are available by other means, often without the trade-offs that microservices imply. For example, many people talk about how microservices make systems more scalable. But you can write very scalable systems that aren't based on microservices, and the way in which microservices help with scalability is usually more about effective data sharding than it is about microservices. So what are these trade-offs when it comes to microservices?

And where's the technical debt? Well, the main trade off is between the autonomy with which we can develop something and the design complexity that we need to buy into to achieve that autonomy. ⁓ I've talked about that in other videos. If microservices are to be of any help achieving any increase in autonomy, we need well-defined interfaces between services so that we can make change in one part of the system without forcing change on another.

Sneha Mehra (00:04:46)  
That means taking lots more care about ⁓ the design at these points where the communication between the pieces takes place. If we don't take that extra care, then we're making a very big mistake because we're giving away the independence of deployment that leads to autonomy and yet still paying the cost of having made that choice. And this time, while paying a higher cost in terms of the friction in the development process.

If I have a microservice that is dependent on yours and we can only release our services together after some form of testing, then what have we really got? We have a distributed monolith. And this is a very costly way to implement a distributed monolith because mistakes are more distributed and feedback is slower. ⁓ Now, if you need my service to change to work with changes that you'd like to make to yours, you've either got to ask me to do the work

with all of the problems of scheduling, including the differences between our relative priorities for this new feature. You make care a lot and I can't be bothered. So you need this yesterday and I will originally add it to my schedule for next year. ⁓ Or you could change it yourself. So now you need access to my repository, ⁓ use of my deployment pipeline, and you need to be familiar enough with my code base to change it safely. And I need to trust you to do it.

This second approach is actually a pretty good strategy, but using separate repositories just adds more complexity and more overhead and extra barriers to this working well. A much easier strategy for dealing with coupled code like this is to adopt a shared code ownership in a single repository and use continuous integration and continuous delivery to evaluate all of the changes together whenever they happen. If we keep everything in one repository, evaluate every change together in one deployment pipeline,

we're more accurately facing the reality of the situation that these so-called microservices aren't really microservices at all. They're really just simply components of something bigger that we're unable to determine whether it works or not without testing it altogether before we release. That's a monolith. ⁓ So the sense in which microservices represents technical debt is that we implement more complex code that is less computationally efficient

Sneha Mehra (00:07:11)  
and that works in ways that are less organisationally efficient unless these things are truly independently deployable, as the definition for microservices says they should be. The problem here is once again one of semantic diffusion. The meaning of the term microservice has been devalued, watered down over time, seemingly now to mean a small lump of code that communicates via XML over HTTP. The idea was much more than that.

as I discussed with the inventor of microservices, James Lewis, quite recently. Here's the usual working definition.

Sneha Mehra (00:07:49)  
And nearly all of these characteristics are primarily there as mechanisms to deliver autonomy for the teams that produce them by making the services independently deployable. Notice that despite the common assumptions, there's nothing here that says that microservices must or should communicate via XML over HTTP. You can have microservices that don't do either, and you can have code that isn't a microservice that does both. So we don't gain in autonomy.

we'd pay a fairly significant price for it and get no benefits that we couldn't also get without paying that price. ⁓ That certainly seems to qualify as technical debt to me. ⁓ Even when we can independently deploy our microservices with no coordination with other teams or groups, then we are still incurring a technical debt of a kind. But now it's a better kind of debt in that we are getting something useful for it. We pay extra, but we gain in autonomy.

We pay with more sophisticated design thinking and in terms of significantly more effort and coordination costs when we need to change multiple services or debug our now considerably more complex systems. I know that I sound like a microservice skeptic when I talk about stuff like this, but I'm not. I'm a big believer that this is the most scalable way to build large complex systems and that we should be spending the time to think more carefully.

and come up with better designs anyway. But all this requires a level of design sophistication that isn't always evident. Or to put it another way, as Martin Fowler described it in his 2014 article on the topic, you need to be this tall to take the microservices ride. ⁓ An important part of this design sophistication necessary to make microservices an effective choice is deciding where to draw sensible boundaries between our services.

Dividing things up into smaller, more independent pieces is a very good idea, but not a new one. And it's not easy. And it's not only possible with microservices. It's really what software design is all about. ⁓ Past assumptions that dealing with the scalability problem also fell at this hurdle. Corba and Decom were distributed component architectures that were designed to solve exactly the same problem as microservices a generation before.

Sneha Mehra (00:10:16)  
and they failed because people didn't think enough about what was actually happening when you called a remote service. I recall seeing two core services. ⁓ One executed a loop which called the other service across the network to process data a byte at a time. This is at least a thousand times slower than running this loop in process because each byte would now be sent on its own across the network and take at least one kilobyte.

in the form of a network packet. This is just one of the problems that Matt Ranney talks about in his interview that I mentioned earlier. People building microservices but not thinking about the implications of communicating over a network. ⁓ There are significant costs to communicating over networks, not just in terms of speed, though that's certainly one consideration, but also in terms of the way that distributed communications can go wrong. It's a much more complicated thing.

What happens to your microservice system when the service that your service calls isn't there? ⁓ Or is working so slowly that you don't get a response in the time that you expected? So when designing distributed systems like this, any distributed system, microservices or not, you must be mindful of the conversations that are happening and the ways in which they can go wrong. Matt said that the average fan out from a request

to their system at DoorDash resulted in over a thousand messages from services. ⁓ One message resulted in a thousand interactions. This is quite a scary thought, not just in terms of performance of the system, but also the complexity of diagnosing problems if something should go wrong. ⁓ The advantage and one of the costs of microservices is that the logic inside the service boundaries kind of becomes less important.

and the communication between the services more important. ⁓ So we need to design the communications between our services with much greater care. ⁓ These points in the design of our system matter a lot. They give us big wins if we get them right. But more often, because of the difficulty of getting them right, they come at a huge cost.

—-----------------------------------

01  
Sneha Mehra (00:00:04)  
you

Sneha Mehra (00:00:14)  
Hi, I'm Dave Farley. If you've watched any of my YouTube videos, particularly those on the topic of microservices, you know that I care deeply about building high quality software. Microservices are a great approach for building software at scale. But although the ideas at the ⁓ root of microservices may sound simple, this is not a simple approach. There are several big traps along the way and it's important to avoid them.

if you want to gain the benefits of a microservice approach. In this microservices masterclass, to help you to avoid some of these traps, we've curated some of my most important free videos, along with key segments from my engineering room interviews and lots of other material to guide your learning. We've structured all of this into a proper training course. You'll find clear learning objectives, helpful introductions, transcripts,

key takeaways, exercises, quizzes, and you can earn a certificate at the end if you get all the way through. If you're an executive or a team lead, this course will help you to make better decisions and to avoid some of those commonest of mistakes. If you're a developer, I'll give you the clarity and context to build smarter systems from the beginning and so guide you on your path towards better outcomes. If you're an architect, our course will

sharpen your thinking and reinforce the principles that matter the most in the real world of building complex distributed software that are at the heart of microservices. While much of this content is freely available, scattered across hundreds of videos on my YouTube channel, this course pulls it together into a more coherent whole and transforms it into a more focused journey, ⁓ offering a structured learning experience designed to help you to move from passive watching

to active mastery. I hope that you enjoy it and I hope that you find it helpful in your future endeavours. Enjoy the course.

—--------------------------------  
03

Sneha Mehra (00:00:01)  
Microservices represent the most scalable approach to software development. They achieve this by distributing development to many small independent teams. ⁓ That means that each team can make progress more independently of the others. That's what gives big organizations the ability to scale development. This is all good stuff, but it comes at a cost. It's actually quite difficult to achieve that decoupled independence.

microservices demand a high level of design sophistication as a result. So where should you start? What are the techniques to achieve systems that are composed of these independent decoupled pieces? In this episode, I want to explore microservices again. I've spoken about microservices before and pointed out that they're a bit trickier than many teams think. A microservice is defined by these properties. They're small.

focused on one task, aligned with a bounded context, autonomous, independently deployable and loosely coupled. I think that the first thing to notice is that this says nothing at all about technology. Nothing to do with REST APIs, for example. The next elephant in the room here ⁓ is independently deployable. ⁓ I've covered this aspect in the past.

But I think that this is the defining characteristic of microservices. In fact, if you look at this list, all of these ideas are aimed at helping us to achieve that independence of deployment. ⁓ If you and I can't develop our services independently and then release each of them without needing to test them together before that release, they aren't really microservices and they certainly aren't independently deployable. So that's quite a tough design challenge.

If that is what we need to achieve, how do we go about it? Well, I think that the most common starting point for teams adopting microservices is actually the wrong one. That is setting up a separate repository for each new service. The problem here ⁓ is that defining interfaces between the services that you can rely on to be stable enough and loosely coupled enough to give you enough protection so that you can change

Sneha Mehra (00:02:22)  
the behaviour in one service without compromising the interface that others depend on is very difficult. You're almost never going to get that right the first time. Forgive me name dropping, but I was talking about this idea recently with my friend Eric Evans. ⁓ Eric pointed something out that was a consequence of my approach, but that I hadn't thought of in these terms before.

The language or protocol of the information that we use to communicate between the services is a separate bounded context. Let's imagine a small group of services. Our aim is for these things to be independently deployable. They need to talk between one another to do useful work. ⁓ So of course they are coupled to some degree. ⁓ They're coupled through the conversations that they have with each other.

So if we want them to be loosely coupled, then the nature of those conversations is very important. Certainly we'd like them to not be too closely tied to the implementation of any one of these services. This idea of these conversations representing a separate bounded context resonated with me. It helps me to better understand why my approach to microservices works for me. So some ground rules before we explore this in a bit more detail.

Bounded context are a big idea. They were introduced in Eric's book Domain Driven Design. The idea is that multiple models coexist in big software systems. A bounded context is an area of the problem where one of these models is consistent. If we are building a bookstore, then the bits of the store showing the books for sale and the bits of the store responsible for, say, shipping the books, both have the concept of book.

But those concepts are different. The books for sale probably need some cover art and some samples of the contents and that kind of thing. The shipping part probably needs to know the weight of the book and the destination address, but doesn't really care about the pictures or the text. That's because these are separate bounded contexts, separate models of bookness, whatever that means in their context. It's a very good idea

Sneha Mehra (00:04:43)  
to always translate ideas that cross the boundaries between boundary contexts. The alternative is that you have to establish some kind of grand one model to rule the moral data structure that works in every possible context. This is close to, if not impossible, and even if you did manage it, then every single change to the model would break everything. So this is really bad idea.

much better to follow the advice and translate between contexts. Remember our definition of microservices. One of the items says, aligned with a bounded context. So what this means is that every time two services communicate, there should be a translation between the concepts that they exchange, at the boundary between them. I've talked about this before in a previous video too.

If we think of the messages that we exchange with that information with as a distinct bounded context in the way that Eric suggested, then that means that we should translate to the message and from the message. There are some little bit of code that acts as the insulation between our services and the messages that it sends and receives. This is how I've built distributed service-based systems for years, but I hadn't really thought of it like this before.

The key idea here ⁓ is that we either want our messages to be more stable, to change at a slower rate than the services that produce and consume them, ⁓ or more independent, to be able to change without forcing change on the services. This little translation layer helps us to achieve either one of those things or both. ⁓ And the idea of the messages representing a distinct bounded context

helps to guide us towards some better answers as to how to achieve that. The first problem then with a microservices system ⁓ is which pieces should we define as services? And then the next, which follows very quickly afterwards, is how do we establish these magic messages that are durable enough and loosely coupled enough? We're looking for good abstractions here, and the best approach to finding good abstractions is to iterate.

Sneha Mehra (00:07:02)  
Create your first guess, try it out, see how it stands up to real use, and then refine it as you learn more. So my preferred approach to creating a microservices system isn't to begin with microservices. Before I get there, I need to be able to iterate quickly enough to learn about the problem. Only then can I abstract it so that the messaging context is reasonably stable.

We don't need some elaborate whiteboard exercise, but it will certainly be handy to think about the problem a little bit, to play with models of it so that we can try out and find our first guess at services and messages. My preferred approach to this part these days is event storming. There's a link in the description if you'd like to learn more about that. But we also need to write code, live with it for a little while.

and see how it works out as our understanding of the problem that we are solving and of the system that we are building deepens. Only then will we know if our design works to insulate our services from change. During this phase, I want to minimise the overhead of changes. As soon as any of my services live in a separate repo, there's a significant overhead involved. If I change service A,

and that requires me to change the messages and so change service B, then after work in repo A, switch to repo B, run continuous integration in each, coordinate the versions to see if they work together, et cetera, et cetera, et If service A and service B were in the same repo, if all of their tests ran in the same continuous integration cycle, if their releaseability was determined by the same deployment pipeline,

I get much faster feedback. I can make my change to service A to the messaging between them and service B in seconds. I can evaluate them all together for releaseability in minutes. Continuous integration gives us the clearest, most definitive feedback. ⁓ And it works even when our solutions are relatively tightly coupled. So we can start off with our best guess of a good design.

Sneha Mehra (00:09:18)  
separate services interacting through ports and adapters, perhaps, with our current best guess of that messaging context. But we can use continuous integration and continuous delivery to evaluate all the pieces of whole systems or large subsets of systems together in one pipeline. This is dramatically simpler at this stage if everything is stored in one repository. This

is the simplest way to get to a definitive statement of the coherence and so releaseability of our system. Work like this for a while and evolve the messaging until it doesn't change very often. Now pull out services as microservices based on messaging that you know works and store them in separate repositories. The risk of this approach is that it's easy for a developer being careless

to ignore service boundaries within the bigger scoped repository. So you need to adopt a little bit of care to keep these service boundaries clean. Maybe add some tests that reject changes that try to share code that worries you across service boundaries. Despite this risk, ⁓ I see this as a much better alternative to premature decomposition of services and repositories. ⁓ I think that's an extremely useful way to think about this.

is in terms of its deployability. ⁓ I describe a deployment pipeline as a mechanism to determine the releaseability of our software. That means that within the scope of a pipeline, we need to be able to establish a definitive answer to the question, can we release or not? So the correct scope for a deployment pipeline ⁓ is an independently deployable unit of software.

At the start of your project, when you don't yet understand the problem really, you don't know which services make sense and you don't know which messages will be stable in the face of the evolution of your system, then you can't confidently release the pieces without testing them together. So bung everything into a single repository, evaluate it in a single pipeline shared between teams, but still, architect your systems around services and try and keep your services independent.

Sneha Mehra (00:11:37)  
so are nicely modular. Later, when you've learned the answer to these questions, you can pull things apart. The problem is that as soon as you care whether version 7 of service A works with version 15 of service B, you have a dependency management problem. And you don't have independent teams working on services, you have a monolithic team working inefficiently in a bunch of separate repos. This is a more complex solution.

I was pleased when Eric described the messaging as a separate bounded context. It helped me to rationalise my thinking. But on reflection, I think that there's a danger that it might be slightly misleading too. So I want to clarify that. The risk is that we see it as a single bounded context and then there is the danger that we're back in the one message to rule the more world. I see this as more of a collection of bounded.

really, clustered around different conversations that might be going on in the system. For example, I don't need every different message that talks about books or whatever else to have the same representation of a book. It's conceivable that there are subgroups of conversations going on, but it is more complex than that too. For big systems you'll probably certainly need the ability to subdivide the conversations this way.

But you also need to be able to correlate the conversations in different contexts. So we need to be able to tell that the conversation about my book shopping book is talking about the same book and the same customer as the other conversation that talks about shipping it. We sometimes need some common ideas that transcend these bounded contexts in our messaging domain and allow us to glue the information together.

This is another common problem that people creating microservices sometimes struggle with. How do we, on one hand, create independently deployable services, but services that can also work together to create a consistent traceable picture? To address this, there's another model, another layout in our abstractions. At the base, we have our services. Next, we have the messages and the message context.

Sneha Mehra (00:14:03)  
And then, on top of this, we have another model. This model of core concepts in our problem domain that glue everything together. This is usually very simple. In the olden days, we might have talked about key entities in our systems or something like that. These are the things that you need to be able to search for, to look up, to maintain the history of, and so on. The customer accounts and books in our bookstore, perhaps.

Our lives will be easier if these things have a unique ID that is durable across contexts. It will allow us to stitch the conversations together. Sure, you can translate these IDs on the way into a service on the way out, but this kind of makes traceability a little bit more difficult. Let's be clear, you can't always do this, but if you can, it does make life easier. This collection of IDs or keys

kind of hovers above the messaging context. And it should intentionally be simple, as simple as you can make it. The good news is that these things are usually pretty obvious when you look at the problem. They're the top level ideas that flow between bounded contexts. My preference is to try and model all of the conversations that the messages represent at the level of the problem domain.

a kind of technical version of our ubiquitous language for our system. This helps us a lot with the messaging and the top-level entity abstractions. If we get this right, then a non-technical person who understands the problem should be able to understand the conversations that are happening between services. They may not understand every last detail of every message, but the broad brush picture will be accurate. So we will have messages like order book,

Dispatch Book and Book Dispatched that contain books and customer accounts and not much else. We'll avoid technically focused messages. This reinforces the bounded context idea of our services and one of the big advantages of that is that bounded contexts are often naturally loosely coupled with respect to one another. So our services will also tend to be more loosely coupled too. So.

Sneha Mehra (00:16:25)  
My advice is to work iteratively to design the conversation between our services. Assume these conversations will be in the language of the problem that we are solving, not focused on the technicalities of how we solve it. And when you need to iterate fast so that you can learn fast, breaking things up too soon gets in the way. Thank you very much for watching.

—-------------------------  
05

Sneha Mehra (00:00:01)  
Microservices are different. They have several really useful advantages but also several fairly significant costs. ⁓ Their biggest challenge is that they're meant to be independently deployable. This means that we don't get to test them with other services before we release them. So what are the implications of all of this and how do we test them? and what does contract testing mean anyway?

Microservices are a very good strategy for building big systems with lots of people. But they come at a cost. They aren't a great idea for simpler systems or small teams because of these costs. So what are the costs and how do we address them? In particular, how do we design and test microservices if we aren't allowed to test them together before we release them? Let's just recap.

This constraints and why it's fundamental and matters. Microservices are an approach that is designed to optimize for team scalability. They are for bigger teams, so that these teams can divide work up between smaller, more focused teams and still make useful progress. The problem of dividing teams up like this is the coupling between them and their software.

If I write some code in my service that changes how it's used and your service depends on my service, then I may have broken your code. If I want a new feature in your service that allows my service to do something new, I'm tied to your team in some way unless I work to avoid it. I can either guess at how your change will work and code for my guess, ⁓ or I can wait until you're done and

And now your team and mine are working in lockstep, unless your API is backward compatible with the previous version. We are coupled together and can't make progress independently. Both of these cases are a problem of coupling, and in big organizations are a very big deal, because this coupling compounds. My team's blocked on yours, your team's blocked on Jane's, Jane's is broken by a change somewhere else, and so on and so on.

Sneha Mehra (00:02:19)  
State of DevOps reports ⁓ say that one of the main predictors of high performance in Teams is their ability to make progress independently of others. Microservices was designed to alleviate that problem. Let's be very clear. There's nothing fundamentally new in the idea of microservices, but this independent deployability is the closest to a new idea in practical terms. Let's get back.

To our two services, yours and mine. ⁓ And the two ways that they can be coupled. I can break your code changing my service, and you can store my development while I wait for a change. There are two ways that we can tackle these problems. We either accept that we must test everything together before release. ⁓ This checks that I haven't broken your service and that our services are in step and can do useful work together. And for this, ⁓

To work ⁓ at any sensible scale at all, what we need is fast, efficient, high-quality feedback. So that for this scenario, microservices are a fairly dumb idea. We don't want barriers between our services that slow us down. The easiest way for me to see that I've caused a problem with your service when I change my service is in a continuous integration and a shared repo. We put both of our services together in a single repository, practice continuous integration.

And get near instant feedback on whether our changes work together. Similarly, if I need a change in your code to help my service do something new, the easiest, most scalable way to do that is to allow me to make the change for myself when I need it, rather than wait for you to schedule it. So again, shared code ownership, single repo, not really microservices. But there's another approach, and that is the microservices approach.

For this, we're going to accept the challenge that independence of deployment presents, and be willing to pay the costs in terms of a bit of extra work and a bit more sophistication in the design needed to achieve it. What this demands of us is really two things. We need to design APIs that are loosely coupled ⁓ and APIs that we can defend. This is where the ideas of microservices being aligned with a bounded context and being independently deployable really come from.

Sneha Mehra (00:04:41)  
Interfaces between services should be abstract. It's important that they hide implementation detail so that implementation can change. Another benefit of microservices ⁓ is that they should be easy to replace with better versions. So we don't want the detail of how they work inside leaking out. Aligning them with a bounded context helps with all of this. But

The other part of the technicalities of dealing with the boundaries between bounded contexts is that you should always translate at these points in communications. Ports and adapters is a really good design idea at these points, to provide a bit of help for insulation between our services. ⁓ And a place in your design to cope with changes more easily. ⁓ We'll come to that later.

So now we have a service. It has a well-defined API of some kind, and it only interacts with other services via their well-defined APIs. These APIs could be anything. No reason at all why they should only be HTTP. That's not in the definition of a microservice. Perfectly valid to have a microservice that consumes or publishes binary data or anything else. But tagged data formatted messages can help, but we'll get to that.

My preference is that services are message-based and asynchronous. I think that makes our life easier, but this is still just my preference. Nothing in the definition of microservices said that they must be. Let's imagine ⁓ I want to create a customer registration service of some kind. It allows me to register new customers. So here's an add customer message, it contains a customer name and an address. And in response it generates a customer added event.

That contains a customer name and a unique ID that's been generated by our service. I could have chosen to include the address in the response, but by doing that I'm coupling the idea of customer to the idea of address. I know from painful experience that addresses can be quite annoying sometimes, ⁓ and that we may want different things from them at different times. So this is something that I expect to change, so I'm going to separate that out in my design.

Sneha Mehra (00:07:03)  
If someone wants to ask my service for a customer address, that's fine, ⁓ but we'll deal with that separately. This is a separation of concerns decision. I've decided that services that are interested in new customers probably aren't always interested in their addresses. Not saying that this is always right, merely that by making this choice my design is a little bit less coupled, or at least the coupling is a little bit more discrete. When I store my new customer,

I could just take the incoming message ⁓ and process and store it as a blob. When somebody asks for a customer, I simply send them the contents of the blob from the store with no translation. That's not a very good design. It may mean that my service was simple to write, but it's now also fragile. If I want my service to evolve, there isn't really very much that I can do here because every consumer of my service.

Is now coupled to how I've decided to store its data. There's no abstraction, there's no information hiding of any kind. This is actually quite a common form of service in teams that are trying to move to expose information from legacy systems, for example. They will often write a service that simply passes on some blob of data. So now consumers are coupled to the service ⁓ and their implicit understanding of the structure of that data. ⁓

The service isn't really going to do very much work. It's certainly not reducing coupling. A service, or actually any module or class, or function in our code, should keep some secrets, otherwise, what's the point? In our example, we have two ideas: customer and address. We've already imagined that addresses may be slightly slippery. What if our business is successful and now, as well as addresses with a UK postcode?

We decide we'd like to add support for German or US addresses as well. If a UK postcode is implicit in the data structure that we expose, there's no option but to force on all of our consumers of our service that we want to add US or German addresses. The boundary, the API of our service, is a special place in our design, ⁓ or it should be. It represents an integration process.

Sneha Mehra (00:09:22)  
A designed point at which our aim is to reduce coupling. That's why it's an API in the first place. This is particularly true for microservices because if changes here force changes on others, that's a big deal. If we don't test them together before release, we could break other people's code that way, or maybe even the whole system. So how do we gain the ability to change our microservices safely? The first thing

Think about is the information that we expose and how we do it. What does our API look like to a user of our service? What implementation detail secrets do we want to keep within our service and how do we keep them? Our primary goal with the design of our interface is to expose our best guess of the minimum that our users of the service will need to understand to use it. Don't add lots of extraneous details.

Keep it as simple as possible and as focused on the service that you provide as possible. Don't add stuff to the interface just in case, ⁓ or just because we have it, so we might as well. Actively design the interface to be simple and minimal. Then present that interface in a way that, as far as we can guess, is likely to continue to make sense in the future. We won't get this right, but recognizing that it would be good if we did.

Makes us try just a little bit harder for a decent abstraction now. This is part of how you build a more loosely coupled interface. The other part is how you organise the information and access to it. Let's back go back to our customer and address. We could decide to adopt a fixed binary format for our messages. The first two bytes are customer ID. ⁓

The next 100 bytes are the customer name and the next 200 after that are the customer's address. Everyone that uses addresses now needs to understand this fixed structure. Clearly, this is far from loosely coupled. Later we realise that if we end up with more than 1024 customers, ⁓ this design breaks. So we decide to increase the customer order ID to 16 bytes instead of two.

Sneha Mehra (00:11:42)  
This change breaks our user interface with every user of our service. We could improve things by making a class and hiding the detail of the structure of the data inside it. This is a small step, but it's a step toward looser coupling. Our service users still need the correct version of the class to send or consume the data correctly, ⁓ but at least they aren't worrying about byte count anymore. We could abstract this a little bit further.

We could add an interface instead of a class. So now we could plug in different implementations beneath the interface. This represents a contract that we are using to manage the conversations between us. We could imagine adding some kind of version number to different implementations of our contract. And then our service, if it was sent in our old version, say one that stored IDs in two bytes, could upgrade the incoming information into a usable form.

That the latest version of our service needs. We still have some problems though. Things are getting more complicated for our service now. We've now got two classes of users, early users who have ideas that can be stored in two bytes, and later users who IDs can't. So we either need to keep supporting both groups or we need to migrate everyone to the new version.

In a real microservices world, you don't get to force changes, at least not in lockstep, on other people. So we must keep the old version and the new version of the APIs working alongside one another at least for a while. So now we have extra complexity added to the design of our service to keep our system stable overall. Okay, so most of you probably aren't worrying about bytes and byte order.

But you do you expose a file format or a data stream from, I don't know, a hardware device or a legacy system direct through your API anywhere? Does your service or the code of your consumers need to understand the data structure or file format? Because if they do, ⁓ you may well be concerned about bytes and byte order after all. The reason that most teams don't need to care is because they've taken another step up the ladder of abstraction.

Sneha Mehra (00:14:00)  
They use more dynamic data structures like XML, HTML, JSON, YAML, and so on. These are often tagged delimited forms, they're horribly inefficient for ways of sending data, but having the tags means that now we know what each piece of the data means. Now we can depend on the tags rather than just the position of the bytes. This allows us the freedom to shift things around. If I have a tagged form of my customer and address like this.

I can easily add a telephone number, and as long as you don't care what order my tags are presented in, when you consume them, ⁓ you don't have to assume that you must consume every tag. Then your old code that knew the previous message structure will continue to work with my latest changes. This is another form of loose coupling. So what is the contract, the interface, the API to our system?

It's clearly more than only the functions that we invoke to initiate the behaviour that we want. It's more than add customer because we publish information too. So it certainly includes customer added and maybe get customer address. Our contract is also the data that our service publishes that represents a customer, their address, their ID, their telephone number, and so on. ⁓ All this stuff is externally visible and so represents the contract to your service.

Changes anywhere here can break communications with other parts of the system. ⁓ If our service is a genuine microservice, then these are the only places that can break interactions. Because all interactions happen through these interfaces. So this is the goal, the scope of contract testing. ⁓ Our aim is to confirm that we haven't broken our contract with our external users. We'd probably want to evaluate adding a customer and validating the response.

To start with, that's probably all we'd need. Later, we'd want to find a customer with an old ID and one with a new one. Later still, we'd add a German customer and a US customer perhaps, and confirm that their addresses are supported correctly when we ask for them. And later still, maybe that we still get a response when our system was down and later came back up. When we run our contract tests, if nothing changes, then great.

Sneha Mehra (00:16:22)  
Our changes can't break anyone else, ⁓ unless our abstraction leaks in some other way. If one of our contract tests fails, though, we may have a real problem. So the aim of contract tests is to exercise the surface area of your interfaces with the outside world to spot any changes. This surface area includes the data that you expose ⁓ that consumers need to understand. If a test fails,

Stop and think, is this an innocent change like adding the telephone number? In which case, as long as communication protocol that we are using between services isn't positional, then it's probably safe to change. Is it a breaking change for some people? Users who used to have two byte IDs can't see IDs with more than two bytes. In which case we'll have to do more work to insulate these users from change. We could keep supporting the old two-byte API but encourage them to update.

Then when they finally do, we could retire the old API perhaps. Maybe we keep an old and new version of the service running in parallel. Or maybe we implement support for different versions of the API in our new version of our service. Finally, there may be changes that for one reason or another are going to force change on users of the service, whether we like it or not. Now we'll need to go cautiously. Remember, we can't force change in lockstep on our users.

We may not even know who our users are. So now we will have to support old versions of the API and communicate with users or people who may be users to let them know of the upcoming changes. We may need to do s old school stuff like runner beta programming and support external testing some way. One more thing before I wrap up. I think that recognizing that the interface to your service or system is everything that is visible to the outside world is an important idea.

It's not even just all the externally visible data as well as all of the externally visible functions. The operational characteristics of your system or service matter too. This starts to move in the direction of reactive systems, which I cover in more detail in this video. But the idea of the contract of our service or system is not as simple as the functions in an interface. ⁓ Or the collection of messages that it responds to. It goes a bit deeper than that.

Sneha Mehra (00:18:48)  
These are the problems of supporting external APIs. And the value of microservices is built on that idea of supporting external APIs. Thank you for watching.

—---------------------------  
02

Sneha Mehra (00:00:03)  
microservices is probably the most popular architectural approach today. It's extremely effective. It's the approach used by many of the most successful companies in the world, particularly the big web companies. As a result, I see lots of organizations and teams attempting to adopt it. Or so they think. In reality, they often miss both the value and the costs of microservices and end up doing

something that really isn't microservices at all and so kind of doing the microservice idea something of a disservice. They don't gain the benefits that they were looking for very often. ⁓ So what is the problem with microservices? In this episode I want to explore the problem with microservices and particularly why this very valuable architectural idea is probably not quite what you think it is. It's simple, right? Microservices is all about small services, isn't it?

No, not really. The first problem is that microservices is a distributed systems architecture ⁓ and however you wrap it up distributed systems are way more complicated than non-distributed systems. There are all sorts of problems that you just automatically ⁓ buy into when you start distributing computing across multiple different devices.

One of the things that microservices are not are not ⁓ services with simply a REST API. That's another thing. That's REST. It can be implemented as a microservice or not. ⁓ REST may be a useful technique in implementing microservices or not. ⁓ There are many different ways of approaching this. But a REST API does not define a microservice. And all services are not microservices.

Services as an organizing principle for software systems has a very long history. We'll talk a little bit more about that later on. The defining characteristics of microservices are that they are small, as we've already mentioned. ⁓ They're focused on accomplishing one single task. They're aligned with the bounded context in the problem domain. They are autonomous, independently deployable, and loosely coupled.

Sneha Mehra (00:02:27)  
Let's talk about each of those ideas in a bit more detail. So what does small really mean? Everybody asks this question and it's kind of an obvious question from the name, right? Microservices. So what does micro mean? There is an answer that comes from the early pioneers of microservices, ⁓ which is that a microservice fits inside of James Lewis's head. ⁓ James was one of the

people that first popularised the idea of microservices. along with Martin Fowler, wrote the first article that was published, I believe, ⁓ online, or at least the best known article that describes them in more detail. ⁓ And James, as you can see from this picture, has a reasonably large head, he's a smart guy, ⁓ so a fair amount fits in his head, but the idea is to make them understandable, to compartmentalise a problem so that we can deal with it and understand what's going to happen.

⁓ A reasonable rule of thumb to try and figure out how small is small ⁓ is or how micro is micro is to imagine ⁓ throwing away your micro service and re-implementing it. How long would that take? If you can do that in a few days or maybe a couple of weeks, it's probably on the right kind of scale. If the idea scares you of throwing away the service and rewriting it from scratch, it's probably too big.

The next in the list is focused on a single task and doing it well. ⁓ This is really about separation of concerns, but at the level of the problem domain, what we're really trying to express here is the idea that a microservice accomplishes one task when viewed from the outside. Now implementing that task may involve a bit more, ⁓ more concerns perhaps. Maybe you're thinking about storage or getting data from another source or

something like that. ⁓ from the outside ⁓ the service is going to be focused on accomplishing one task and achieving that well.

Sneha Mehra (00:04:32)  
The next in the list is Aligned with a Bounded Context. Most teams miss this idea and it's an important one. A Bounded Context is an idea that comes from Domain Driven Design, the fantastic book written by Eric Evans, ⁓ which describes ⁓ modelling the problem domain as an approach to designing the software.

And one of the things that Eric talks about, one of the key ideas in domain-driven design, is this idea of bounded content. Eric defines it as a defined part of software where particular terms, definitions, and rules apply in a consistent way. This is a part of the problem domain where the concepts are related to one another, and so it makes ⁓ a cohesive unit. ⁓ This is an important idea

in lots of different contexts but in the sphere of microservices it's a very important idea because those are the boundaries that we would prefer to align our services with. And this is important for several reasons. ⁓ First, if we break a problem, ⁓ a monolithic application say, into a ⁓ collection of smaller technical services that doesn't really qualify as microservices. They are not aligned with the bounded context and you're going to suffer.

from increased coupling between those components. Bounded context are kind of like natural fire breaks in the problem domain. ⁓ They are, if we align our software with those fire breaks, ⁓ then it means that our system is going to naturally be less coupled because the problem domain is less coupled at those points. And so it means that we can write cleaner interfaces to our services and have less chatty interactions between them.

One of Eric's pieces of advice in the Domain Driven Design book is that whenever you transition a bounded context ⁓ boundary, then you should translate ⁓ ideas, information that crosses the boundary. That means ⁓ that every microservice ⁓ should be translating its inputs and translating to its outputs in order for every interaction. ⁓

Sneha Mehra (00:06:53)  
microservice should always make a break between these points. We should make a significant distinction between the external protocol with which the microservice exchanges information with other parts of the system versus the internal representation of that service and the model within. We need to really treat seriously the difference between the design of those APIs, those entry points into our services

and our consumption of those APIs from other services. If you aren't using ports and adapters in your implementation of microservices, you're probably making a big mistake. We want to always be translating at these points. The next in the list is that microservices are autonomous. ⁓ The teams that look after these microservices can make progress alone.

without the need to interact with other services or other teams. I can change the implementation of my service without needing to coordinate with anybody or anything else. This is the biggest value of microservices. This is the thing that allows microservices to work in organizations that ⁓ scale enormously. But it's also the most commonly missed attribute of microservices. What I see frequently

These teams claiming that they're implementing microservices, but building, testing and deploying all of those pieces together. That's something else. Let's explore this idea for a moment of autonomy. Service-based design is not the same thing as microservices. Service-based design has a long history. It's been around for a very long time. If you'll forgive me.

I'm a bit lazy and I haven't looked up the history of service-based design so I'm going to do this from a personal point of view and tell you my interaction with it. ⁓ I've been building distributed service-oriented systems, however you want to frame that, for about 30 years now. I started at the end of the 80s and 90s with my first foray into distributed computing. ⁓ I moved reasonably quickly in the early 90s into ideas like ⁓

Sneha Mehra (00:09:13)  
remote procedure calls and then quickly into an idea of something called object request brokers where we were simulating interactions with component-based systems ⁓ in distributed systems. ⁓ I got involved in a startup where we built some infrastructure to support what we would now think of as quite advanced microservices based on a concept called cooperative business objects, little bundles of domain logic that communicated

through what we call a semantic data service, a bit like XML, but predating XML. ⁓ I then moved on bigger commercial systems, working on component-based systems, message-oriented architectures. ⁓ Service-oriented architecture came in at the beginnings of this century and became more popular. And then I moved into building high-performance systems based on event-driven architectures, sometimes using patterns like CQRS, microservices.

and ending up with something that ⁓ we call reactive systems. ⁓ All of these are service-oriented systems. All of these models were based on that concept of little bundles of logic that were somehow more discrete from one another than the classes and functions within our implementation, and so ⁓ had more abstract ⁓ interfaces between them. The ideas of service orientation are extremely useful.

They're an important idea, I would argue, in computer science and building bigger systems. ⁓ And I think that they principally matter for two reasons. ⁓ One of those is technical, the other is organisational. Now these are very closely related, as Mervyn Conway stated in 1967, ⁓ any organisation that designs a system, defined broadly, will produce a design whose structure is a copy of the organisation's communication structure.

So the architecture, the design of our systems, is closely related to the communication topologies that we employ in the organizations that we work in. This has an impact on the kinds of architectures that it's possible for teams to build. the architecture mirrors your organizational structure. ⁓ If we decide to divide up our teams into

Sneha Mehra (00:11:34)  
on technical boundaries, we have a UI specialist team and a middleware team and a data team, then we're going to build a system that looks a bit like this, a tiered layered system, and the teams will have to coordinate and interact. ⁓ More modern software development ⁓ approaches tend to prefer smaller, more autonomous teams. ⁓ If we're going to, for a variety of reasons, if we are going to organize ourselves along the lines of

more modular autonomous teams, we're going to build more modular software with more autonomous components and it looks like this. But this is not necessarily a microservice organized system. We could choose how we were going to organize our teams around this. We could put everything into one big repo and have a distributed service oriented system communicating with REST or messaging or whatever other technique. ⁓ Or we could have independently

deployable units of software, autonomous components ⁓ that were microservices. They're not the same thing. So if we can have distributed monoliths organized as collection of services, which is probably my preferred architectural approach if I'm honest, ⁓ for many systems, what does microservices add to that? Well, the key thing is that they are independently deployable.

Deployability is one aspect of autonomy for these services, but it's an incredibly useful one. If we can work on changes to ⁓ our service and deploy it without necessarily communicating or interacting with other groups of people, that means that we can make progress without being constrained by those groups. It means that we are developmentally less coupled to the rest of the organisation.

that's really what independently deployable means. If I need to care about dependencies on other services before I release my change into production, my software is not really independently deployable anymore.

Sneha Mehra (00:13:44)  
Well, this is a big idea and this is the biggest stumbling block that I see with teams adopting microservices. And the biggest problem, the biggest cost of microservices, this has serious implications for the way which we design software and the way in which we employ it. You don't get to test your services together before you release. You don't get to force changes that you make to your service ⁓ on

consumers of your service. You can't demand that they update in step with you. That's not part of a microservices model. If our services are autonomous and independently deployable, we build, test and deploy them independently. This is the real value of microservices. This is the ⁓ step that decouples organizations both technically and culturally.

and allows them to grow and create software at an enormously fast rate. It's the thing that differentiates organizations like Amazon from more conventionally organized ⁓ firms. ⁓ And this is also the real cost, because if you want to be able to independently deploy these services, then you've got to be able to take the ⁓ ability to ⁓ maintain that independence really very seriously.

Microservices are an inorganizational decoupling play. That's what they are for. If you don't need to decouple your organization developmentally, they're probably not the right thing to do because they come with this cost. They allow development organizations to grow very big and to work as a large collection of small independent teams. But to achieve this, microservices must be loosely coupled.

We need to take that separation of external and internal representation incredibly seriously and we're going to invest time and effort and development effort into keeping that separation clean. We need to treat the external interface of our services like a public API because it is. We can't demand that other people change in step with us when we want to make changes to our API.

Sneha Mehra (00:16:03)  
So we've got to be very, very cautious about the ways in which we can change that API. When we're consuming the API from such a service, we should be cautious not to couple ourselves too tightly to it. We should only take the information that is of interest to us and no further. We'd like to make those interfaces as clean and abstract as we can achieve. So we should be very, very, very cautious of breaking changes at any point.

It brings in techniques like supporting multiple versions of interfaces on the same service ⁓ or multiple versions of services in some approaches. These are the techniques of microservices. This is the cost of microservices, but for that cost you get some huge advantages. We need to create interfaces to our services that we can defend and rarely and only with great caution change.

