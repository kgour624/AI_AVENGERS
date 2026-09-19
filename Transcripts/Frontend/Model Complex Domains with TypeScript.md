Sneha Mehra (00:00:00)  
So let's look at the roles we have, and then I want us to discuss like how we might implement ⁓ additional additional rules here. So we have ⁓ don't put your plant out of bounds.

Don't plant plants on top of other plants. ⁓ we've got an interesting like parametric rule here. And I say that because unlike these first two, it takes arguments. And so what we're saying, this one's a little weird, but planting beds can never be more than 80% full. And so we can see that one in action if we go and just try to copy, copy, copy, copy. And eventually it's gonna say you've crossed the threshold.

This is one that we should probably remove. It's just sort of an example of something we could do. Like this doesn't make sense in the context of if I'm putting my gardener hat on, like I absolutely want to plant plant things everywhere in the raised bed that I'm allowed. And then ⁓ we could restrict plants in zones. We have specific plant like here we're looking at sunlight requirements where each zone has a sun level. ⁓ And in this case, we're saying, well, like,

Tomatoes for some reason are the only thing, like there's a special case rule for tomatoes and sun. So ⁓ and then here is an adjacency, ⁓ an adjacency-based rule. And what we're looking for here ⁓ is if I can scan through real quick, ⁓ you can plot pass in ⁓ a array of plant incompatibilities.

and a reason. And if we look how this is wired up, I'm just gonna have to grab the usage of this ⁓ in the garden page.

Sneha Mehra (00:01:57)  
So ⁓ tomatoes and potatoes can't be planted next to each other for some reason. And so if we were to drag ⁓ do we have a potato here?

Sneha Mehra (00:02:12)  
Nightshade. I don't see it, but ⁓ let's let's just change it. Let's say carrots.

Sneha Mehra (00:02:23)  
And I should be prevented.

Sneha Mehra (00:02:29)  
From like putting this here. Tomatoes and carrots should not be planted together. Great. So we have like kind of a flexible framework here.

Let's model some constraints. So I'm gonna shift into my my gardener mode. ⁓ And

Given ⁓ what we already have, which is ⁓ inbounds, ⁓ not on top of each other. Well, like that's the very basics. Like that is sort of adhering to what it means to be things on a grid. ⁓ we've got adjacency, we've got ⁓ characteristics of the bed versus plant needs.

Sneha Mehra (00:03:18)  
So

If we put our gardener hat on, what are what are some other rules that we may want to to think about here? I mean, realistically, let's say I know nothing and probably know next to nothing about gardening. I I would ask you the gardener, considering that these are what we've talked about so far, right? Like these are the things you don't do. ⁓ Just ask the simple question like when you're planting, what you know which plants don't you want to plant next to each other and

Why? Are there any other circumstances where you wouldn't plant a given plant that you haven't listed? It's a little general, I think, but I I don't without getting the answer to that, I don't know how to suss out any more specifics for No, I think that's that's a great way to start here. So ⁓ we're i in this in this commit here, we're already modeling this concept of ⁓ something called companion and antagonist plants.

And so we can sorry, I'm I'm trying to arrange something very specifically here. So these are these little circles that are showing up. I'm trying to get a more interesting one. This that's a green one there. And I want to drop garlic in and mess everything up.

So these indicators, like the little red semicircle, ⁓ this is where we can say, like, there's something about one plant that that hurts the other. And right now, all we're doing is we're lighting up indicators. But you could imagine like a strict mode here where we could say, look, clearly we already have this data, like we have these indicators. Maybe we just prevent.

Sneha Mehra (00:05:01)  
Maybe it's just like an invalid placement to drop anything where you're creating a problem like this.

Sneha Mehra (00:05:10)  
Garlic literally does, ⁓ you know, it it ruins things around it. Whereas this, ⁓ like, these are ⁓ these are kind of like companion effects. Now, granted, we're talking about blocking actions, but you could imagine like a validation rule that's shaped a little bit differently that says, all right, like there is a more optimal way to arrange this garden. Like you have ⁓ inadvertently like defeated all of the beneficial effects from plant placement.

But that that might fit more into sort of an optimization ⁓ optimization concept. So like for now, let's let's go with with where this conversation led. Like we clearly have plant interactions that are harmful. And those are clearly like evaluated, you know, as as we move things around, like you can see on the corner, on the side. So let's ⁓ rather than just sort of like lighting this up, let's

Reject the placement of the plant if that happens. ⁓ We can build a rule for that. So let's jump in. So what we'll do is create ⁓ a new ⁓ we're gonna create a new validation rule in this sort of namespace for them, and eventually we'll end up adding it here. So we'll go back to this ⁓ point here, we'll add it right on top, and we'll call this ⁓ prevent antagonist.

Sneha Mehra (00:06:38)  
⁓ plant adjacency. ⁓

Sneha Mehra (00:06:46)  
And this is a validation rule that operates ⁓ on this ⁓ the same like plant item.

Sneha Mehra (00:06:59)  
⁓ And

Sneha Mehra (00:07:12)  
And it's just telling us we need to return the appropriate thing. So first let's like return ⁓ the object, the name, ⁓

Sneha Mehra (00:07:28)  
⁓ well just call it.

Something like that. I this might be in a tool tip so I don't want to get too long.

Sneha Mehra (00:07:40)  
And we can start out with like ⁓ sorry, fat arrow. We can start out with saying we're going to always ⁓ reject this thing. Now what we what we need to return is something called a validation result. And this this just is a type that that kind of represents this ⁓ returning an error rather than throwing it. It's just, is it valid? What's the reason?

Sneha Mehra (00:08:17)  
So let's let's just like see if something gets thrown. And ⁓ we'll add a comma. We'll wire this rule up. That's fine.

Sneha Mehra (00:08:31)  
Gonna go back to garden page ⁓ and just like that, prevent antagonist plant adjacency. And what this should do is basically like anything we try to do, we should see no because I said so. Zooming out.

I said so. All right. So we're we see that clearly we're falling down the path of like this is invalid. Now let's worry about like making a real determination of whether, you know, whether this is this should be allowed or not. So ⁓ what we can do in terms of finding this information.

Sneha Mehra (00:09:11)  
It happens to be on the workspace. ⁓ because it kind of makes sense from from a domain modeling standpoint to say, like this is this is just like ⁓ known information about plants, right? It's not like this one garlic plant, like one of seven, should really not be placed next to tomato plant number six. This is really on a on a category by category basis, like garlics.

And tomatoes should not be together. And so this almost is like ⁓ it's it's at the same level as seed packets and the information about seeds and like the way tomatoes grow and the way tomatoes work. So it's not, it doesn't have to do with like positioning on the grid. Now there's something that's derived from it, but in terms of the interactions between plants, we we might model it that way and just say category A and category B are helpful or harmful.

And and ⁓ here's how that works. Let's explore how this is modeled. So we've got an indicator and there's an ID for it, ⁓ just so we can keep track of ⁓ like when when they're presented in the UI, we can go look up information to display about, you know, is this harmful, is this helpful? ⁓ And we've got an in ⁓ an interaction effect, and this is ⁓ an item type like

Tomatoes or basil, and then the nature of the effect, which is beneficial, harmful, or neutral, and then a description to present in a tooltip. So these are these are sort of directional effects. If you have things that are mutually beneficial, they would just have one in either direction. ⁓ And ⁓ these can kind of stack up and they can be used to drive this kind of UI ⁓ that we get here, where like in this case, lettuce is good for tomatoes.

Tomatoes are good for lettuce, and we have that bi-directional beneficial effect. So each of these green boxes is is one of those effects. So let's let's use that and evaluate whether the placement will result in ⁓ effects being, you know, being felt.

Sneha Mehra (00:11:29)  
Great, so we've got our indicators. ⁓ And yeah, that's great. We'll extract that out.

structure it. ⁓ And ⁓ we also want to ⁓ for the current placement we want to kind of look at ⁓ look at neighbors and so I'm gonna just borrow a pattern from down here and say ⁓ I want which one would be a good one? Incompatible plants? This seems

This seems like a good place to start. In fact, what we could do ⁓ this is a little bit of a cheat.

Sneha Mehra (00:12:15)  
How do we regard plant one and plant two here?

Sneha Mehra (00:12:21)  
Yep, they're already based on plant categories. So all we would need to do is kind of leverage this already existing rule. ⁓ and and we can ⁓

Sneha Mehra (00:12:35)  
In fact, I want to structure this differently now. Because we have plant validation rules.

Sneha Mehra (00:12:44)  
⁓ incompatible plants ⁓ we can then say indicators ⁓ dot ⁓

Sneha Mehra (00:12:58)  
Map.

Sneha Mehra (00:13:03)  
And we can just generate a bunch of rules that are sort of stemming from this incompatible plants ⁓ function that already exists.

Sneha Mehra (00:13:19)  
Yeah, this will be this will be fairly simple. So ⁓ prevent antagonist plant adjacency. We'll go over here. Great, that type checks. Now, ⁓ what we want ⁓ is ⁓ like let's let's assume there there are other things that already detect to see like is this an invalid placement or not. We've got a target X and Y, and like let's just start by seeing like is there a plant to my left that

I'm an antagonist with. Or let's let's start with even more basic ⁓ an even more basic concept. Like is there anything in my raised bed that I shouldn't be planted near? We can we can begin there. ⁓ And we need some more things off of context.

Sneha Mehra (00:14:05)  
like ⁓ the the target zone, the target X and Y.

Sneha Mehra (00:14:21)  
And ⁓ then we can kind of cycle through the indicators and see if like there's anything else in the zone that ⁓ might like the plant that's attempting to be placed, which is gonna be one more thing we need, the item, right? Something that I don't want to be near.

Sneha Mehra (00:14:43)  
And we'll type this appropriately in a second.

Sneha Mehra (00:14:59)  
Okay, effect has a ⁓ target item type ID. So for now we're gonna be selfish and we'll say like if there's anything that affects me badly, let's make sure that ⁓ I'm not plants that planted there. ⁓ So if the effect is equal to ⁓ item dot category.

Sneha Mehra (00:15:29)  
Relevant indicators ⁓ push effect or something like that.

Sneha Mehra (00:15:42)  
And we'll say is valid, ⁓ is

Sneha Mehra (00:15:48)  
⁓ based on whether we found any relevant indicators.

Sneha Mehra (00:16:14)  
So we're saying like some number of indicators ⁓ poorly affect this plant.

And we need to scope this ⁓ actually within ⁓ within the bed. So rather than just pushing, we need to look through the target zone.

Sneha Mehra (00:16:40)  
Placements.

Sneha Mehra (00:16:44)  
And see if ⁓

Sneha Mehra (00:16:49)  
Any of these ⁓ placements in the bed actually like without adding this condition here, we're just gonna say like in this world, do there exist negative indicators that affect me? And so now we're saying like in the target zone, are there plants?

Sneha Mehra (00:17:20)  
Okay, so just to add some comments like

For all indicators for each effect of the indicator.

⁓ if it affects the plant being dragged.

Sneha Mehra (00:17:51)  
That's the right arrangement. ⁓ and the plant ⁓ and the plant like creating the negative effect hurts ⁓ me.

Consider it to be relevant as a blocker for placement. And one last thing we have to throw into this. So ⁓ if effect. ⁓

And

Sneha Mehra (00:18:32)  
If the effect is harmful ⁓ and the item being placed

Sneha Mehra (00:18:40)  
Yeah, and and the the the sorry, not the item being placed. The thing we already found to exist in the bed is something that creates that effect. ⁓ Is valid we have our condition flipped. We're saying it's only valid if there are indicators, ⁓ are relevant indicators that are greater than zero. Let's let's log real quick.

Sneha Mehra (00:19:20)  
⁓ one more thing we could do here ⁓ is logged indicator or log each effect.

Sneha Mehra (00:19:42)  
So we should at least see a bunch of stuff cycle through here. Psh. ⁓ Okay.

Sneha Mehra (00:19:51)  
Plant onions garlic plant broccoli beneficial.

Sneha Mehra (00:20:01)  
⁓ and let's ⁓ let's log the item being placed as well, just so we can see all that information real clear.

Sneha Mehra (00:20:29)  
And I just wanna focus in on like the one the antagonist one here.

Sneha Mehra (00:20:39)  
There it is.

Rockley beneficial.

Sneha Mehra (00:20:46)  
Ps. All right. So this is ⁓ if target item type ID is peas.

Sneha Mehra (00:21:07)  
We'll log there. ⁓ And this should be item

I think it's item ID, not item category. That would explain it.

Sneha Mehra (00:21:26)  
All right. One more try with feeling.

Sneha Mehra (00:21:33)  
Broccoli moves over here. Fine. Garlic moves up here. Peas

Sneha Mehra (00:21:44)  
One indicator negatively affects sugar snap peas. So now we have this like nice database of or nice representation of like which plants don't get along with each other. And you could imagine how we could do it in the other direction as well, where like if you're a plant and you introduce negative indicators into the bed that you're being placed into, then you know you could have this flag as well. But like part part of the value here is ⁓ this idea of

Bothering to formalize something like validation logic into its own data structure so that you could imagine like if we have 30 rules that stack up for some reason, like say we get into modeling irrigation and sunlight and things like that, you're just you're able to like very effectively unit test these things and pass that context object in and make sure that like this this ⁓ you know each piece of logic is modular and they kind of all stack up and it doesn't become, you know, ⁓

a big sort of monolithic ⁓ part of your code base that's like a a a a bunch of this ⁓ a bunch of these essential constraints that that matter to your user.

—--------------------

Sneha Mehra (00:00:00)  
So this is what we're trying to do. Now, if you notice in our example app, we've got ⁓ we've got quite a few seeds coming through here. And this is because we're actually reading that file and passing some of the most basic possible data through. ⁓ And now we can start to, but you know, you see there's like I net weight, I have no idea. Maybe that's a string we can search for and put something meaningful there ⁓ in in the future. And we're just kind of

We're gonna try to fill in some of this information that's on the back here. ⁓ days to harvest is an interesting one. Like that's that's another time-related thing that kind of came up. And I don't know, let's say for the purpose of this, that's our starting point. We're gonna build that version of the software, we'll show it to our user, and let's see, let's see what ⁓ they think about that. ⁓ So ⁓ great. So we're gonna we're gonna build this, ⁓ and ⁓ we're gonna have to re

touch a couple different places. Well first off, there's a seed packets service. Remember, these are our domain services. You're gonna find that this thing has some methods ⁓ and it mostly deals with just seed packet. All of the request response stuff is sort of above this part of our server. ⁓ And ⁓ while we may end up like making sure data threads through properly, we're not gonna be ⁓ hopefully spending too much time there. So we need to touch a couple methods in there.

Parse seed packet so that as we read this data file, we can grab pieces of data that are interesting, that are not quite represented on this model yet. ⁓ And ⁓ we also may want to take a look at get all seed packets, although I think things should just sort of pass through. But this is the list where when we're when we're like our UI is requesting all the seed packets to generate those components, this is what it's hitting. ⁓

In our types package, we have request and response shapes. And this is the ⁓ API route that handles ⁓ the actual incoming requests and then delegates the fetching of the list of seed packets to that domain surface service. So again, a nice separation between like request response ⁓ responsibilities ⁓ and the sort of core thing that is dealing with.

Sneha Mehra (00:02:25)  
The seed packet domain.

⁓ we also have a couple UI components, which we'll we'll come back to this ⁓ so because like let's let's start with the back end and like trickle it trickle it up. But there are two UI components where we can sort of thread through some of the data once we kind of incorporate it into the model and see that it's coming through in the HTTP response. So ⁓ let's begin. And don't don't overlook this point. So we we have a generic like

packet concept that the UI understands, but it has ⁓ a a ⁓ metadata object that lets us ⁓ add additional information. And this this is a good example of like the API contract looks very different potentially than the way we're going to articulate our data model. We may have a rich seed packet class ⁓ in our back end, but our UI looks at things a little bit differently.

And so we're going to show how we can sort of adapt between those two contracts without letting any of that request-response stuff bleed into our ⁓ server-side data model, which is sort of the, you know, it's the it's the thing that should be like the most clear representation of of ⁓ of this concept. Great. So where we're gonna start is loading, sorry, we're gonna start with the model. Let's begin diving into the implementation.

By ⁓ extending the seed packet model. And here's here's where we're gonna look.

Sneha Mehra (00:04:01)  
In our types package.

Remember, this is where we have our value objects and our entities ⁓ and the resources, which are the request response shapes. ⁓ let's look at entities ⁓ and seed packet metadata. ⁓ And like I wanna I want you to show you how this is sort of threaded through. We've got a generic packet type, which just has like a name, a description, a category, and then this thing called a presentation, which if we look at it, it's like.

Path to an icon and an accent color of some sort. If we look at our UI, like that is ⁓ the accent color is sort of the background behind the icon, and it varies from plant to plant. It's sort of this is a ⁓ GPT estimated color that represents this plant. And so we have accent colors there. But but like there's not much here. It's just purely

This could be like a packet of snacks or something. Like there's nothing plant-ish about this yet. ⁓ But we have a function here which takes in a metadata schema as an argument ⁓ and it creates a new type with that metadata schema on it. So if we look at ⁓ the seed packet type, you can see we're doing exactly that. We're creating a packet type.

And we have this metadata schema that's being sort of tacked onto it. This this ⁓ is our seed packet. And so all that's left for us to do is to enrich this. So we already have a quantity here, and we can start to add on, ⁓ add on other things. So ⁓ we want, just to refresh my memory here, days to harvest. We want days to harvest. And that's usually, by the way, the time from planting the seed.

Sneha Mehra (00:06:02)  
to when you expect to get something to eat. Well, the best time to get something to eat. I guess you can eat small plants, but usually not worth the work. So great.

Sneha Mehra (00:06:19)  
And in this case, ⁓ we're already saying days. ⁓ We could we could make it time to harvest, but real we just want days to harvest here.

Sneha Mehra (00:06:31)  
Or even int, because that would be an integer. Great. Anything else? Has to be greater than zero. Great.

Sneha Mehra (00:06:44)  
I'm not even gonna say less than three sixty-five. You can grow you can grow a fig tree and it takes ten years before you're gonna actually get figs for the first time. So but negative time is not a thing. If you find a seed that does the har you can harvest in negative time, let me know. That seems like ⁓ solving world hunger kind of discovery.

Sneha Mehra (00:07:07)  
Now this this is squirrely, like, I don't know, do ⁓ an empty seed packet, does that belong in our collection? We'll leave that problem for another day. But like you could say ⁓ one is the minimum here, so that you actually hit a validation error and potentially you could use that to drive business logic and say, we're removing the seed packet from the collection. As soon as you mutate this entity and say, okay, quantity zero now, it's like, well, something else has to happen.

we're actually removing the record. ⁓ let's look back at the the the back of our packet here. ⁓ let's let's pick one of these instructions. How about the sew instructions? Because I do recall ⁓ on the back of our seed packet, ⁓ we had

Sneha Mehra (00:08:01)  
Seed depth an eighth eighth of an inch, ⁓ plant spacing 24 inches apart. That seems pretty important. So like those are like planting instructions. You can think of them that way. Also, like we might find some day some something here. So let's let's see if we can fish in our our raw data model that represents that YAML file and dig up something interesting there. ⁓ I do want to add one more thing.

Sneha Mehra (00:08:39)  
I wanna I wanna add planting distance and I'm doing this because I know where we're going. I know we're gonna aim for this like s gardening thing, gardening app where you can drag plants ⁓ onto grid ⁓ and I know that we're gonna wanna know what the size of those plants are so that we can sort of plan for for that. And we're gonna wanna have planting distance.

as as part of that. ⁓ Especially if ⁓ as as we were talking about, like a plant probably derives from a seed. You want to have information on this that is ⁓ very useful in terms of modeling a plant. ⁓ just checking packet real carefully make sure we don't duplicate our duplicate our work. So we have a category, let's call that like we had a plant family in our in our drawing here.

So we've already got quantity. Let's say like category kind of meets this need, at least for the API contract. ⁓ we'll we'll get that expiration date and name. Name is already on packet as well as category. So like this and this kind of get get what we want. ⁓ and then expires at, we can add that here.

Sneha Mehra (00:10:03)  
And because this is an API contract, I'm gonna make this a string. So that because we can't send date objects over the wire as it is.

Great. So that's that's done. Now let's turn our eye towards the server. And that is where we're gonna build like kind of the real ⁓ entity here. So we're gonna go into server, source, entities. ⁓ And we already have a seed packet here. And like there's already kind of a starting point of some sort. We have a basic representation of a plant. ⁓

Many plants come from a single seed packet, so this is just part of our starting point modeling. Not all of this needs to be exposed through our API, right? This can be a richer model. ⁓ we've already got a name, a description, a quantity here. Now remember, this is going to be like metadata.quantity when when we're dealing with that that ⁓ serialized shape that that we pass back to the UI. Category, that's gonna be on the top level because.

It's it's part of category here, right? ⁓ it's got the presentation, and this is a like take it from me, this is a one-to-one mapping, it's the icon, it's the accent color. And then planting distance. So it it's it's fine to represent all of this that way. ⁓ we need one more thing, an expiration date, I think. We can always come back here if we need ⁓ need more. And ⁓

Sneha Mehra (00:11:47)  
We'll just make this like a a standardized timestamp. ⁓ easiest easiest way to go about doing this. In fact, we don't even need this. It will infer based on string that ⁓ it should make it a text field.

Sneha Mehra (00:12:11)  
what I I think we should be good here, but let's see. SQL light constraint.

Sneha Mehra (00:12:19)  
That might have been at the top of my buffer there. Nope, it's at the bottom. So what's going on here? Insert into, and then it's got the SQL insert statement with these parameters. And it says not null constraint failed. Seed packets expires at. So somewhere we're trying to create a seed packet, but we've just added this expires at field. ⁓ And it appears that by default, this is not a nullable field. Now we could make it one.

And the error should go away. But I kind of want every seed packet to have an expiration date. So let's not make it nullable. And let's fix this error by having ⁓ like that field get populated with data as we load it from a file. We'll leave this at smallest level so we can kind of see see it when it gets fixed. So where we're gonna go next is ⁓ the seed service, ⁓ seed packet service.

All right, so ⁓ at a high level, like two things are happening here. ⁓ we've got ⁓ a function called parse seed packet. ⁓ And it takes in a repository. Remember that's the same concept we were dealing with before, but apparently I am provided a repository when this function's called. And then it's got this thing called raw seed packet info. ⁓ And this is ⁓

This is the monster type with everything that's in that YAML file. This is ⁓ a very, very, very detailed seed packet type. And the point, we'll we'll sort of fish around because we have nice little type aheads and TypeScript for the data we need, or we'll look at the YAML file to figure out where to where to go and fish for it. But ⁓ what we have to return is the seed packet entity. So this function represents getting ⁓ one ⁓ seed.

As raw information from that YAML file, and then returning ⁓ a persisted record, it looks like. So this is ⁓ creation of the packet. ⁓ interesting. I bet I bet that the persistence happens outside of this, ⁓ right? Because pr like the packets are ending up in the database. This must be how it works. So we've got repo create and then a bunch of the existing things, ⁓ and we need an expire Zad.

Sneha Mehra (00:14:54)  
And

Sneha Mehra (00:14:58)  
So let's let's look at our YAML file and see if we can spot a good place for an expiration date.

Sneha Mehra (00:15:06)  
Actually we're we're riffing a little bit here, so I'm not sure that we're gonna find one. We can always just vegetate.

Sneha Mehra (00:15:15)  
Bread, mature size. ⁓

Sneha Mehra (00:15:23)  
Viability in years. Let's use that. Let's say I bought an enormous number of seeds ⁓ in ⁓ February of 2024 and we'll use like seed packet info viability in years.

So we'll do ⁓ new date and then

Sneha Mehra (00:15:53)  
And let's see, it's possibly undefined.

Sneha Mehra (00:16:17)  
Great. Sorry, just needed that formatting to work out. So we've got we've got a date, we're passing in year, a month index, ⁓ and then ⁓ a ⁓ a day, right? And like we're gonna forget about the time. In fact, we could probably forget about the day too. ⁓ now we're dealing with this. So unexpected constant nullishness on the left-hand side of a double question mark ex

Operator. I think you need to enclose the coalescing operator and ⁓ yeah. I I'm not sure that's two as well question marks two. I ⁓ see.

Great. So we've got 2024 plus the viability in years. ⁓ And let's scroll to the bottom and look, ⁓ no more SQL SQL light constraint there. ⁓ so we're already getting the category, the quantity. ⁓ let's just double check. Planting distance, is that already calculated? Yep, it is. So really we just had to we we had to add expire as that. Now let's look at what's going over the wire and see if.

if that data looks like what we need. So just open up good old DevTools Network tab, search for the ⁓ hackett's request. So here's your response. ⁓ And I'm just scanning through. We've got category big boy tomato presentation ⁓ metadata. All right, there's our quantity and our planting distance. So those are already wired up somehow. We're gonna have to look at that.

but I don't see any sort of expire that coming through. And this is where I would I would want to see it. Remember, like metadata is the place for any plant-specific things to be trickling through. To do this, we're gonna have to get out of our domain service and get into the route handler, because that is where ⁓ we kind of convert between our internal representation ⁓ and ⁓ the ultimate ⁓ HTTP response that we're returning, which

Sneha Mehra (00:18:30)  
you know, looks like this. ⁓ And here we're gonna have to say expires ⁓ at his packet.

Sneha Mehra (00:18:41)  
Expires at. And if we hit save.

Sneha Mehra (00:18:47)  
Packets.

Sneha Mehra (00:18:50)  
There it is. A timestamp that comes through. Now we can go and thread it through to our ⁓ our seed packet component, like in the back. So we're gonna close a bunch of these files.

Seed packet back is what this is called. ⁓ And ⁓ because we're sharing types like that at types package. ⁓

All this information is kind of trickling through, and although it's not shown in the tooltip here because ⁓ it gets trunk truncated.

wait, I expected that to kinda come through. What is this type?

Sneha Mehra (00:19:36)  
You know what, I wanna bump the TypeScript server just to the language server just to make sure we're getting the latest types after that change.

Sneha Mehra (00:19:50)  
Hmm. That's interesting.

Sneha Mehra (00:19:56)  
⁓ gotta rebuild the tips package. Okay, no worries. So I'm gonna just put this server in the background.

Sneha Mehra (00:20:12)  
It'll npm run build so that we get the right types in the disk folder. And then we're gonna bring that server back into the foreground. ⁓ And I think we should get something more here. No?

Sneha Mehra (00:20:27)  
This is in the entities dist.

Sneha Mehra (00:20:38)  
All right, let's try one more thing.

Sneha Mehra (00:20:43)  
Gonna blow away this disc folder just for good measure.

Sneha Mehra (00:20:52)  
Build this package directly, see if there are any errors. All right, what's what is in your dist folder types?

Entities. Seed packet.

Sneha Mehra (00:21:10)  
And seed packet metadata.

Sneha Mehra (00:21:15)  
There we go. Quantity plane distance days to harvest. All right, let's make sure that all percolated through the way it should.

Sneha Mehra (00:21:41)  
Maybe things are still warming up. Like ⁓ maybe it's svelt. So for people following along live, I I figured out what the problem was. It had to do with our our ⁓ the audible we called to try to fix the import paths earlier. ⁓

What was going on here ⁓ is ⁓ we in our TS config had changed or had set root dir to dot. Right? We'd said the root directory of this project ⁓ is ⁓ effectively the root folder of this package. ⁓ And ⁓ then we ran a build. ⁓ And what what happens when we do that is you're gonna see in this folder here.

⁓ if we I think refresh.

Sneha Mehra (00:22:32)  
Did I save that? I didn't save it. ⁓ We're gonna see a source folder pop up here. So what happened was we had like two versions of the build output for this package. ⁓ One in the root level of the disk folder, and that that one had ⁓ only quantity in that metadata object. That's why we were seeing just that one option in the type ahead. ⁓ And meanwhile we had like the real changes we just made inside this source folder. So if you just set this to source, save, and then

remove disk and build again, you should end up with a nice dist folder that just has everything in the root here. And we should be back in action.

—-----------------

Sneha Mehra (00:00:00)  
So now let's work on the plant. We need to model this as well. Or we already have a plant here. So we can we can look at this. We've got a seed packet, ⁓ an accent color, an icon path, bunch of good stuff here. And it turns out that when this app boots, if you look at the logs, like we create ⁓ a plant for every seed packet here. ⁓

You could think of these as ⁓ a little different than the way we've sort of set up our domain model, right? Here we've said a plant has a position. ⁓ really what what I've done here is I've said, all right, seed packet refers to a plant, kind of like this. ⁓ And then we've got like a plant placement here. Sort of what this lets us do is if we wanted to, and I think I'll continue with this course since I've got the beginnings of the solution anyway. ⁓

What it would let you do is say, you know, seed packet can be seed packet. ⁓ we have kind of a freestanding domain model here where sorry, it's a zero too many. Where where this is sort of entirely self-contained, where we could say, you know, there there's a chance plant, we we might need to model ⁓ different things here. ⁓ like it's very similar, but you know, we're gonna have a little bit of duplication here.

If if I'm honest. Like it's ⁓ the the key difference is gonna be we have something that represents ⁓ the core data about the plant, and then we've got something that refers to it as being in an XY position. And what this lets you do is like you can kind of like clone the plant and you end up with another plant, but like all this this still represents what we were what we were describing before. Like there are many plants that refer to a seed packet. So ⁓

We can carry that through. If we look at plant, it still refers to a seed packet. ⁓ we will just need kind of something that relate that's like a placement.

Sneha Mehra (00:02:18)  
And we're just gonna grab bed as a starting point. And this will be ⁓

Place plant placement.

Sneha Mehra (00:02:39)  
And this once we we want to refer to a plant. ⁓

Sneha Mehra (00:02:54)  
So we're saying on plant I'm gonna introduce this concept of placements.

Sneha Mehra (00:03:01)  
We don't need width and height. So let's let's finish that relationship there. Over here, we're gonna say ⁓ no, I think we've actually got it here. Presentation. We've already got the many to one with the seed packet. You know what? We can just use plant here. I'm gonna back this up. I like I like the URL's domain models a little bit better. So we still have the concept of a seed packet referring to many plants.

We're just gonna have to get this right as we feed it up to the UI that wants to see this in terms of item, ⁓ in terms of item placements. So what we're missing on plant though is some concept of a coordinate.

I don't see something here yet. Alright, no problem. We already have a value object for this. Column.

Sneha Mehra (00:03:53)  
XY coordinate.

Sneha Mehra (00:04:01)  
Position ⁓ is ⁓ xy coordinate. So that's the xy. ⁓ I don't see anything here.

Sneha Mehra (00:04:13)  
Sorry, I'm gonna get rid of this too. I don't see anything here that relates to like this plant belonging to a bed. So we need to add that.

Sneha Mehra (00:04:25)  
That's more of a relationship. I actually see everything's column, column, column except for this. So we'll add it here. ⁓ And ⁓ this is ⁓ similar. It's it's also gonna be let's think about this. Is this many to one? One bed, many plants. Yep.

Sneha Mehra (00:04:47)  
Bed Bed Bed And then

Sneha Mehra (00:04:58)  
We'll also call it plants.

that a one to many.

Sneha Mehra (00:05:06)  
⁓ so ⁓ this ⁓ this is the other side of a one to many. That's why we said many to one. Sorry, Seth. I I take your take your suggestion right to heart here. ⁓ there's gonna be a one to many on the other side. So when we do bed, we gotta do

Sneha Mehra (00:05:28)  
One to many.

Sneha Mehra (00:05:33)  
Plant.

Sneha Mehra (00:05:39)  
And how do we fetch that relationship? Bed.

Sneha Mehra (00:05:48)  
Kinda like that. So we've got a many to one because ⁓ the garden above this bed has many of me, the beds. And then I the bed have many plants below me. So the first word here is always like, what are what are you in this relationship?

⁓ And ⁓ there's something else we had to wire up here. ⁓ no, I think we got it. Many to one. And then we've got ⁓ many to one here as well. So you can think of plant as almost being like a many, it's it's sort of a think of it almost like an edge in a many-to-many relationship between seed packets and raised beds. Like many seeds of a given type can go in a single raised bed.

But and a raised bed ⁓ you know can have plants of many that came from many different kinds of seeds. And so this is this is really like that join table, but with extra data ⁓ attached to it. Great. ⁓ now let's see if we can wire this up in our API contract to feed it back up to the UI.

Sneha Mehra (00:07:02)  
⁓ actually we gotta go back to our garden service and let's create a couple of plants in here.

Sneha Mehra (00:07:11)  
So we've got a six by six ⁓ raised bed.

Sneha Mehra (00:07:17)  
We're gonna need a plant repo.

Sneha Mehra (00:07:24)  
and import that and then we're gonna say plant repo dot create

Satisfies Deep Partial.

Right.

And what needs to go in here? Position.

Sneha Mehra (00:07:49)  
X and Y are zero. ⁓ the bed.

Sneha Mehra (00:08:06)  
We need ⁓ an accent color. So I think I think what we want to do here is ⁓

Let's let's hard code this for now, and then we can pull it from real data. We'll make this bright red.

Sneha Mehra (00:08:28)  
There's our accent color.

Sneha Mehra (00:08:35)  
⁓ And

Sneha Mehra (00:08:39)  
You know what? Maybe we should just grab it from a seed packet. Yeah, that's what we should do.

Sneha Mehra (00:08:51)  
That has all the data that we kinda want here.

Sneha Mehra (00:08:57)  
So what we can do is say we've got our seed packet repo, we could say seed equals seed repo, proceed packet repo dot find ⁓ one where ⁓

ID is ⁓ and then

Sneha Mehra (00:09:22)  
Or let's not use the ID. Let's say find where the ⁓ let me get rid this. So we're not scrolling through a tiny little slit there.

Sneha Mehra (00:09:39)  
I just want something that'll survive the reload. Category tomato? Yeah, category is tomatoes. And we'll just grab the first one. I like that. Find ⁓ one where category is tomato.

Sneha Mehra (00:09:55)  
And that we probably need to await, indeed.

Sneha Mehra (00:10:12)  
Great, so we've covered that case and now here we can say seed dot presentation accent color ⁓ and icon path.

Sneha Mehra (00:10:29)  
And we can add a whole bunch of other information. Variant. We can say seed.category. ⁓ name. We can say seed.name. ⁓

Sneha Mehra (00:10:44)  
⁓ We need to what? Await the ⁓ not not on the create because this is just kind of creating it in memory and then we have a separate step step to persist it. So this is what would allow you to say, I'm gonna create a bunch of things and I want to maybe as one transaction ⁓ persist them all at the same time. Why wouldn't we just relate that plant to the seed? Yeah, we totally create. ⁓

Putting all this data in the data. You're absolutely right. Let's let's do that. So we've already got presentation and planting distance and description. Yeah, I like that. Accent color. Certainly need the bed. I think it really just comes down to this, right? Coordinates and the relationships. That's great.

Sneha Mehra (00:11:35)  
We've got the bed, we've got the seed, we've got the position.

Sneha Mehra (00:11:44)  
And we're not using this because ⁓ well first let's save that bed up here. ⁓

Sneha Mehra (00:12:01)  
So we've got this is the persisted version of the bed. Really, this is the pattern I'd I'd use for real. Something like this. Where you're like, I've got one version of that that's sort of my params, and then here, ⁓ we're saving it. ⁓ Rename the bed one there. Yeah, thanks. And so this, it's just gonna get its ID populated and all the things that that sort of happen when it's really stored in the database.

And then down here we've got bed. ⁓ we can make this even slicker.

Sneha Mehra (00:12:40)  
Call this seed packet. So we can even do that. And now we just need to persist plant one.

Sneha Mehra (00:13:00)  
This ⁓ is ⁓ wait. Cool. Just checking to make sure yep, we can get rid of all this crap here, because it's the seed is the source of truth here.

I love it. ⁓ And looking at our logs again, make sure that everything seems copacetic.

Sneha Mehra (00:13:22)  
Sequel light constraint.

⁓ plant position X. ⁓

Sneha Mehra (00:13:32)  
Not not set. Hmm. What could this be? Like I see we've got position X, position Y. These are the only parameters that are being passed in. The first thing I would check is ⁓ in our ⁓ app or data source.

XY coordinates in there, plants in there, so that makes sense. ⁓ positions.

Sneha Mehra (00:14:01)  
That is xy coordinate, which is an entity.

Sneha Mehra (00:14:18)  
This is quite odd.

Sneha Mehra (00:14:22)  
Let's check XY coordinate, make sure that this is set up right. X is a number, Y is a number. Can't really get too much simpler than that. Hi folks, ⁓ figured out what the problem was pretty quickly. So ⁓ you may have ⁓ noticed that I mentioned ⁓ like the my my starting point code had this concept of a plant that was sort of very redundant with this concept of a seed packet, where ⁓ on boot.

I don't know if you'd you'd been paying attention to sort of the the scroll of logging of object creation here. ⁓ a plant is created for every seed packet that is read from that seeds YML file. But now that we've stated, well, plants have a position and they refer to a seed packet. And you can think of a plant as representing the placement of a seed ⁓ in a bed, which like that is some excellent domain modeling there. That's that's like

A plant represents a placement of seed in a bed is something that gardeners would certainly understand and agree with that like that is that is the correct way to think about this. ⁓ we're gonna have to disable that code. So we have this file ⁓ in our server module, it's called seed-db. And you'll see like here's where we ⁓ you know we find ⁓ all of the seed packets.

And this happens this this whole function here, in fact, is like kind of what we have to disable. Generate plants. ⁓ So we'll disable that, we'll find where it's used down here. ⁓ And like instead of ⁓ after loading the seeds into the database, we generate one plant per seed. We're now saying this concept of a plant models the xy coordinate of where we're putting our our planting, right? Our our seed that has grown up. And so if we hit save there.

And we can get rid of some of these imports because we don't need them anymore. And we go back to our plant service. There's also like generate plant from seed packet. We can get rid of that as well because that that doesn't really define ⁓ exactly how this is gonna work anymore. And we can get rid of a bunch of this stuff because that was mostly necessary for that purpose. ⁓

Sneha Mehra (00:16:35)  
Now we can go back to our garden service, and what you should see is that error has gone away. ⁓ so the the thing that was like a little ⁓ tricky there is of course we're just working on this plant. Clearly, clearly we have an xy coordinate where we're placing this plant in position, but that error message was from one of those C one of those plants created on app initialization that did not have xy coordinates. And so, of course, this ends up being persisted.

position x ⁓ just comes from plant dot position dot x. And if we look at our database, ⁓ we can see plants, we've got one record in there, and it is in fact at ⁓ X0Y zero. And so we've now got garden, beds, plants. Let's not forget the relationships bit. So when we're getting all plants ⁓ up here, we've got this relations object. And we can say, well.

In this case, like true is sort of the termination of ⁓ a f of a family of paths, if you want to think about this. So what we're saying is go get beds ⁓ and this is the end of what you should go and get. But we can instead say

Plants. And so now we're gonna get when you fetch a garden, we're gonna get all the beds. ⁓ And we're gonna get all the plants that are in those beds. And we should ⁓ start to see that coming through. But remember we hard-coded that placement item when we were preparing our API response. And so once we do some work there, we're gonna go to our browser, we're gonna hit the API, and we should see that our ⁓ tomato plant, whatever the first plant that we found.

is ⁓ you know being presented in the placements array.

—-------------------------------------

Sneha Mehra (00:00:00)  
Welcome to Domain Modeling for Humans and AI. My name is Mike North, and I'm a principal staff engineer and product architect at Stripe. And I've been a front-end master's instructor for over 10 years. Today I'm going to be talking to you about taking a complex problem, ⁓ studying it, working closely with domain experts and AI agents that you're using to sort of structure your code and define your architecture. ⁓

How do we take those complex business problems and translate them into manageable ⁓ units of code that are self-contained, where you can evolve over time quickly and easily as the requirements of your software change? So ⁓ what what am I what am I focusing in on here? Like what is domain modeling and why am I here to talk to you about this? A big part of what I do at Stripe ⁓ is I work closely with teams to make sure that the foundational building blocks we're creating layer up.

In a way where we can offer low-level and high-level things and we don't end up having to change our APIs all the time and change our products all the time. So at the lowest level, you can think we have the concept of money moving between two places, because Stripe is at its core a payments processing company. On top of that, we have this concept of a payment with the payment method.

And you can layer up and up, and eventually you get to these drop-in checkout forms that are sort of turnkey and ready to go, and they have concepts like shipping options and all of that. And so being able to really think about all of those building blocks and how they layer up is a valuable skill. ⁓ And Stripe really focuses in on this. Part of this is involving is studying your problem space to discover the key entities, their relationships with each other, and

any constraints that are at the essence of the problem that you're trying to solve. And then of course the process of taking that mental model and translating that into a software architecture.

Sneha Mehra (00:02:00)  
So ⁓ in this course we're gonna be talking a little bit about domain-driven design. Now, this is a 20-year-old concept. ⁓ there are many good books about this. Eric Evans has a great book on this topic. we're going to borrow some ideas from this that are particularly useful. Like any software architecture concept, you don't want to be adhering to these ideas too rigidly. You should think about them as tools in the toolbox and you pull them out when they're useful to you. But there are a couple gems.

That are ⁓ at the core of domain-driven design that I think every software engineer should be applying, particularly in a world where ⁓ we're increasingly using agentic coding tools to help us author software, right? You're collaborating with product owners and customers, and you're collaborating with an AI as well. And if you bake the the same language and the same concepts into your code as the way the business is referring to these things.

You have a better shot at working together sort of collaboratively to develop this solution, even though you're the person that's writing the code. ⁓ Another point here, like why why domain modeling is important, at least for me, this represents ⁓ a sort of a part of the ⁓ engineering career ladder. So you can think of becoming a staff engineer ⁓ as ⁓ you know learning how to deal with technical ambiguity very well. Like that's often

You know, part of what it means to be a staff engineer. Someone takes ⁓ a well-defined problem ⁓ and gives it to you, and your job is to sort of figure out, like, all right, what what frameworks are available, what libraries, what off-the-shelf tooling can I use. You know, you're it's it's sort of up to you to define like how that solution takes shape. Domain modeling is a great tool for handling business ambiguity, and that is.

Ambiguity in the problem space, not the solution space. So if someone just, you know, gives gives you a vague problem, like, you know, I've got a ⁓ video course website to teach front-end engineers. Like, how should I set this up? Like what what do people need? What are the important entities that are at the essence of solving that problem in any meaningful way? And this is this is a really important part of sort of ⁓

Sneha Mehra (00:04:24)  
Growing your career at that more senior level, as sort of the lines between product manager and engineer ⁓ start to blur together a little bit, and you start to think about like this architect track of your career. So, ⁓ why do I feel like domain modeling is a useful concept to teach in TypeScript? Well, TypeScript types are incredibly expressive. ⁓ They're not the most expressive in the world, like if you look at Rust.

They have this concept of lifetimes, and like right in the types, you can sort of see some representation of, you know, are you allowed to mutate this object that's passed to you or not? ⁓ TypeScript can't do that because underneath it is the JavaScript programming language, and we don't have those concepts. But TypeScript strikes this really nice balance between being incredibly expressive and relatively simple to read, as well as being widely used, which means it doesn't matter if your company is using.

protobuffs or JSON schemas or whatever representation you choose for representing like contracts between system components of whatever you're building. Usually you can start with TypeScript interfaces and types and you can generate whatever you need based on that. So increasingly this is being used as sort of a source of truth for shapes of objects, ⁓ contracts between things, ⁓ even even public APIs, there are some nice frameworks like I think it's like type API that ⁓ really let you

Use a TypeScript set of TypeScript interfaces as the sort of source of truth, and then you can generate like SDKs based on those types. So to study domain modeling, we're going to need a complex problem space to dive into. ⁓ And ⁓ selfishly, I'm gonna have you all ⁓ work with me to solve a real problem that I have. So I I have a fairly substantial ⁓ vegetable garden, and here it is, I took a picture.

a couple days ago before I flew out to Minneapolis to film this. ⁓ And you can see I've got a lot of these like these metal things here. These are called raised beds. So they're just like metal sort of bins, they don't have a bottom. And you fill them with soil and you put some rocks in the bottom so they drain well. But like ultimately, you're ⁓ you're planting things in them. So you can see here, like I've arranged them. You can see a couple little like ⁓ these are tomato plants that are growing, ⁓ and I've got

Sneha Mehra (00:06:46)  
Close to 40 of these things now. It's it's a lot to manage. ⁓ the story starts with, you know, figuring out what kind of seeds you want to order. And this, once you start ordering seeds from some of these companies, they will send you these catalogs where it's just like this one here, the whole seed catalog is probably, you know, two inches thick. And you can flip through this and see like every variety of tomato you could ever imagine. And you order seeds.

And then you end up with a collection like this. And it's it's a lot. Like this is a collection that I've messed over maybe three, four, five years. Some of these are old, which means they're less likely to be effective, right? It's not like food that expires necessarily where it like goes goes bad, but the older the seeds are, the less likely they are to work. And so figuring that out is challenging. These seeds all need to be ⁓ they they're

Turn into plants which want to go into the ground at different times, right? Some of these they're fine to plant in the dead of winter. Like it's totally fine. Like broccoli, you can do that. But ⁓ basil with like you know, like sweet basil with really soft leaves, ⁓ that if if it gets anywhere near a frost, it's gonna absolutely get wiped out. And so just figuring out like when you plant things, what's still good to use, how far apart do you plant these things?

There is a lot of complexity here. ⁓ after I start with the seeds, ⁓ like during the winter, I have these little hydroponic bins where I drop seeds into each of these little baskets. There's a little s growing sponge in there, and you put a seed in, ⁓ and it'll grow purely in water. So there's water in this black bin and you put a little some nutrients in there and it'll grow. So so now we have like an additional level of complexity here. Like these have to sort of get started indoors at the right time.

So that they go outside whenever it's right for the plant, whatever that means, we're gonna have to figure that out and model that somehow. And then they get planted outside. And you can see here, when I do plant these, I I separate plants ⁓ by some ⁓ you know certain distance here. So this is a zucchini plant, it's going to get very big. And these are little peppers, and they'll stay relatively small. And so, like.

Sneha Mehra (00:09:09)  
How far apart do we put these things? There are a lot of properties of these plants that we need to model if we're going to solve the problem of organizing my raised beds and my seed collection effectively. ⁓ So we're going to work on a piece of software today to make this a lot easier. And we're going to apply domain modeling in order to keep the complexity contained. We're going to have some business rules like what should be planted next to other things, or how far apart should things be spaced. And we're going to see how.

It's easy to ⁓ contain this business logic and make it modular so that hopefully you can see that over time, if we were to add 20 rules to this, like how much water does the plant need, how much sunlight does it need, what kind of fertilizer should it be using? Like all of that could be modeled and added onto this in a way where the overall complexity of the software is not going to ⁓ skyrocket ⁓ as the user's needs evolve. So

⁓ there are kind of two main pages we're going to focus on here. ⁓ One is the seed collection. So this is going to be solving kind of this problem here where we're incredibly disorganized. Like, how do we model one of these seed packets? ⁓ What information goes on the back? ⁓ I'm gonna play sort of your user here, and you're gonna learn how to sort of tease information out of me so that you can identify like what are the most useful aspects of this to ⁓ to flesh out.

Right, you can go in so many directions, but like part of part of what we're trying to learn here is like how do you get to the essence of what's most important? And then the the more ⁓ involved part of this workshop, after we kind of solve the seed collection in a basic way, ⁓ we have waiting for this, waiting for us a drag and drop user interface where you can grab these tiles, each of which represents a plant, and you can see like carrots has this little down arrow, like there are many types of carrots here.

You can drag them into the raised bed, and we we're gonna think about like validation logic and how do we how do we model the idea of ⁓ putting a plant in a bed, moving it within a bed, moving it across beds. And we'll talk a little bit ⁓ later about like transactionality, like which of these operations should be atomic, meaning it either all succeeds or all fails, versus which ones is it okay that like things are temporarily out of sync and will eventually kind of

Sneha Mehra (00:11:34)  
converge on a on a consistent result.

—--------------------------

Sneha Mehra (00:00:00)  
Let's switch to our backend here. So where we're gonna go is in our server package.

We're gonna go into source ⁓ entities. ⁓ And we've got a location here, which kind of matches like what we already have described in our location. We we kind of ended up like undoing ⁓ our change there. So this file hasn't been touched. ⁓ but we need ⁓ a monthly temperature range and we'll need a temperature value object. So let's start from the bottom up. So we've got in values.

Sneha Mehra (00:00:39)  
called this monthly temperature range dot TS. ⁓ We'll grab our RGB color and we'll use that as our starting point. ⁓ And what we can do here is say monthly

Temperature range, ⁓ and we want something that ⁓ matches. ⁓ I mean, if if we wanted to, like we have to think about whether we want something that matches the representation in the in the types package. Like we have the option of persisting this in our database in a way that's a bit different from how is a how it's exposed to our API. But I don't really see a need for that in this case. Let's let's see if we can keep it simple.

Sneha Mehra (00:01:29)  
Alright, now I'm running into this problem where the types package ⁓ is ⁓ not resolving here.

Sneha Mehra (00:01:40)  
Let's see. Import from

Sneha Mehra (00:01:49)  
P-shoot types, obviously you have to aliate alias it.

Sneha Mehra (00:01:55)  
⁓ And ⁓ it's just saying the value is never read. That's fine.

One last step here. We have a single entry point for this P shoot P shoot types package. So if we were to click on this, you can see like we export a bunch of stuff here, and we haven't introduced monthly temperature range.type.js. And if we if we hover over that, we can see like this this resolves. And we also have to worry about the value objects. Temperature.

Dot type.js. So we we did this ⁓ and then we did this one up here. So this ensures that like this package exports the files that we just created. Where we left off ⁓ is if we look at the changes that we've made. First, in our ⁓ types, we declared this monthly temperature range schema and we exported the the type for it. We've got the month, the min and the max. ⁓ we had a location here. I removed it for now.

And we're we're gonna see why. ⁓ and in the value objects, you can see we've got like basically we've articulated this shape. Value is a number and the unit is Fahrenheit. And then we made sure to export these two newly created modules from our types package. Now we're turning our focus back to the server side. And so what we're what we've done here ⁓ is I've got an empty class here.

And we're getting ready to articulate the ⁓ this this database entity, right? And ⁓ if we remember from the slide with like going through type ORM, we have to add entity ⁓ to this. ⁓ And there are options you could pass here, like you could say, I wanna describe the table of like the name of the table that'll be created in the database, but we're not gonna, no need to get fancy here. And now

Sneha Mehra (00:03:59)  
It's time for us to define columns. ⁓ And ⁓ I've I've actually borrowed our interface from that types package so that in this case we're going to keep this ⁓ well aligned, like when this object is serialized into an API response, ⁓ or when we're working with it in our business logic, or when we're storing it in the database. It's all the same representation. And for a simple object like this, at least

When you first ⁓ start working with it, you often start out this way. And like inevitably you make breaking changes internally because you need something about your database representation to change, and you can separate these things. But that that is what this represents. Like we're kind of linking the API representation ⁓ and our ⁓ server representation of this of this concept. And so just refreshing ourselves on what this what defines this, we've got month.

And min and max. So here we go. We've got a column ⁓ and month. ⁓ And we're going to use the definite assignment operator here because our ⁓ RRM will take care of making sure that this exists whenever we load it from the database. And just so like if we didn't have this, you would see TypeScript complaining that, like, look, if you create a new instance of this thing, like I have no guarantees that.

month is going to be populated with something. And so this exclamation mark is basically telling TypeScript, it's telling the compiler there is some other process by which this will be populated. ⁓ And ⁓ we can rely on that. And so when it comes to the type checking, assume this will get initialized. And type ORM is going to take care of that for us. And then we have a minimum X

Sneha Mehra (00:06:00)  
So I've auto-imported ⁓ temperature from our types package, right? Remember we we exported this type. ⁓ And I'm gonna do the same thing with max. Min and max. And these both also need a definite assignment operator. Or ⁓ yeah, operator. ⁓ And we're gonna add column to those as well.

Sneha Mehra (00:06:29)  
Now, ⁓ there is something special we're gonna have to do here because ⁓ all we've stated here is a type. And like type ORM, sorry, we we have a complex type here, and all we've done is say max is this sort of compound field, right? It's it's like this object. And SQLite, the database we're using here, doesn't have a like a column type that matches exactly this. And so we have to do some work to see.

like ⁓ to sort of make sure that the ORM knows how to persist this thing. And in order for that to work, we have to create the value object on the server side. It's gonna look a lot like this. So I'm gonna copy the contents of this file and we're gonna go into this values folder and this is where we're gonna create temperature.

And I'm just going to use this as the starting point. So just so you can see the file tree, it's in the source values temperature. And ⁓

Sneha Mehra (00:07:33)  
We'll just do this.

Sneha Mehra (00:07:37)  
We'll call this eye temperature.

Sneha Mehra (00:07:54)  
Being used up influence.

And ⁓ because this is a value object, right? This is not going to have an ID, we're gonna get rid of this.

It's not an entity. This is a value object. ⁓ And we see that we're incorrectly implementing this interface. So here, here we really, really want to keep our API representation and our server representation aligned. We need a number and we need a unit that's either C or F. So we need value.

And we need unit. ⁓ And this is a ⁓ it's gonna be a string, but really we want temperature unit. We already defined this, right? We got that that nice union type with the C or F. And this is coming right out of our Zod schema and it's sort of trickling through, and we can get rid of this. So now we have a temperature class.

And what we're gonna see is that when this gets persisted in the database, we'll have ⁓ it'll sort of be embedded with its own columns and some prefixing so that it does it does end up being sort of flattened in the column layout. But when we're working with this in memory, it's gonna be this nice nested object, and it it just feels like a nice value object there. So now we're gonna grab this class, ⁓ and instead of referring to this temperature as just sort of like the type.

Sneha Mehra (00:09:24)  
We want to refer to the entity. And so be careful when you pick from these two. This up here is just the type information. This ⁓ in value slash temperature, this is the class. ⁓ And one more last thing we have to do is this.

And what we're doing with this here is we're saying, ⁓ when you want to create like ⁓ when you need to instantiate like the ⁓ a a type for this, right? You need to ⁓ when you're reading a row from a database, you need to create like this temperature object that we just defined. This is the function you call to create that thing.

Sneha Mehra (00:10:08)  
And so now we have our monthly temperature range. And we should be able to build

Sneha Mehra (00:10:18)  
And there it goes. Now we have one last thing we have to do, and that's associating with a location.

Sneha Mehra (00:10:28)  
And we're gonna do that with

⁓ another another annotation here from

Sneha Mehra (00:10:42)  
gosh.

Sneha Mehra (00:10:47)  
Snow belongs to. What the heck is it?

Sneha Mehra (00:10:52)  
It is ⁓ it's one one to many to one, that's what they call it.

Sneha Mehra (00:11:05)  
Or no, sorry, one to many. One to many is what we want. So this is a one to many relationship between

⁓

The monthly temperature range, let me make sure location is coming from the right place, should be the entity, not the types package that you're getting this from. It needs to be definitely assigned like anything else. Right? So you should see class location, not to be confused with like window.location or anything like that. ⁓ And ⁓ we need arguments here. So this is where we can look at ⁓

Some other examples that exist here. ⁓ Let's see. Plant has one of these, I think. ⁓ look at this. ⁓ We've got a many to one. So we need something to sort of instantiate the record as we had before, and then some way to sort of like fetch how this relationship works. So here we go. Back to monthly temp.

Location?

Sneha Mehra (00:12:17)  
We've already imported it. And then the last thing we need is given a location, like what's the other side of this relationship? And we can wire that up on the other side. So for now, ⁓ let's just say it's location dot monthly. ⁓

Temps, ⁓ something like that. Now this field doesn't exist yet. We're about to create it on the other side. So we're gonna go over to the location and say many to one.

Sneha Mehra (00:12:52)  
Actually, I think this side is the one too many. And this is gonna be monthly temps.

Sneha Mehra (00:13:03)  
Monthly temperature range.

Sneha Mehra (00:13:12)  
We need definite assignment. We need to make sure this gets imported.

Sneha Mehra (00:13:21)  
And this will need ⁓

Something like range dot location.

Sneha Mehra (00:13:31)  
So we've we've made this side happy. Just just so you can see like the two components here. This top one is like create the record. And then this one is ⁓ other like ⁓ walk the association. Right? Like how do I from a range get location? And then on the other side, we're gonna need from a location, how do I get the range? So we've got a one to many here. Sorry, this we're gonna change to like a many to one.

Sneha Mehra (00:14:07)  
Right, so we've got the location, and then this should be actually should be like this. Monthly temps, we have many, and on the location side, we have one location going back. So what what we've done in summary here, we've set up the has many relationship, we have a value object of temperature, we have this monthly temperature record, ⁓ and ⁓ we're we're getting really well positioned to ⁓

you know, incorporate this in the AP API contract for the ⁓ the the temperature calculator. Let's just do one more build to make sure that this all works.

Sneha Mehra (00:14:49)  
Great. And I'm gonna make a ⁓ I'll make a git commit here. So we're gonna call it like DDD1.

Sneha Mehra (00:15:00)  
⁓ monthly. ⁓

Sneha Mehra (00:15:08)  
Monthly temperature ranges.

Sneha Mehra (00:15:28)  
So if you want to pull down the D D D progress branch, you'll get all the code changes that I just made. ⁓ and it should work.

Sneha Mehra (00:15:45)  
The next thing we're going to do is turn our attention to getting that drop-down working with the list of locations. And then we can incorporate this sort of temperature checking logic. So if you if you try to run the app, if you run npm rundev in the project, you'll see that entity metadata for location monthly temps was not found. ⁓ one last thing we need to do here. The ⁓ there is sort of the entry point for

the way our server deals with different entities in our database and that is in ⁓ our ⁓ data source.ts file and in here you're gonna see we've got a list of entities and so like important to know these ⁓ are ⁓ I think this is the full set of things that need to be kind of like pulled from

From different tables, and it'll include the value objects as well. So we're gonna add monthly temperature range, making sure we pick the correct one here, right? And temperature. Format it, hit save, ⁓ and this error should go away unless we've ⁓ we need a primary key here. I have a good solution for that. What we can do is extend from this base class. So if you look at some of these other types, you can see.

I extend P Shoot Entity plant and I have this constructor. ⁓ this will take care of our our primary key for us. The conventional way to do this ⁓ would be to add one more column to monthly temperature range, like this. ⁓ ID. ⁓

Sneha Mehra (00:17:33)  
And we would say this is auto incrementing.

Sneha Mehra (00:17:38)  
Yeah.

Sneha Mehra (00:17:43)  
Primary job generated column. And you could say, type is a UUID ⁓ or something like that. Or I think you can just do UUID. Yep. But we have a base class that'll take care of this for us and give us a nice prefix for this URL. This this alone would be ⁓ you know, should be enough ⁓ to ⁓ to to make this work here. But let's

Let's borrow this concept that we see elsewhere in our code base.

Sneha Mehra (00:18:26)  
Something's weird here, but

Sneha Mehra (00:18:34)  
We're gonna implement it we're gonna fix our imports here. ⁓

Sneha Mehra (00:18:54)  
And we should so I just deleted this interface here just to keep keep things going smooth, but like something odd where it's like treating this as both the class in an interface and still I think this is basically like a little a caching error that I have locally. ⁓ you shouldn't have to do that step, but ⁓ we'll we'll factor that back in. I'm gonna just copy what we have from plant. ⁓ And ⁓ what we're doing here is we're basically picking a nice ID prefix for this.

So we're gonna just call it MTR for monthly temperature range.

Sneha Mehra (00:19:31)  
⁓ And looks like that didn't quite make it happy yet. Data type object and monthly temperature range unit is not supported by the SQLite database. So ⁓

Sneha Mehra (00:19:45)  
The temperature range unit.

Sneha Mehra (00:19:50)  
Unit belongs on temperature. Let's check that out and see if

Sneha Mehra (00:19:59)  
See if that makes sense. So we've got entities, we've got values, temperature, ⁓ unit. ⁓ we can call this text. And now it starts working. So temperature range unit, like it's basically saying I don't have a column type for this union type of like C or F. And we can just say, look, store this as a string. Great. So now ⁓ we could.

create these temperature ranges ⁓ and store them in our database, which is exactly what we're going to do next.

—-----------------------

Sneha Mehra (00:00:00)  
So ⁓ we're going to look in our location service. This is source services location.ts in our in our server. ⁓ And we've got a couple to-dos here. We've got like the need to parse some data into a model. And then ⁓ we're going to ⁓ implement the actual calculation. ⁓ And up here we can see ⁓ we've got ⁓ our get all locations.

⁓ you know field here. So let let's see what this is printing out.

This location's find, so this returns a promise that resolves to an array of locations.

Sneha Mehra (00:00:45)  
And we'll keep returning that.

Sneha Mehra (00:00:54)  
And we'll log those locations to the console. Let's see what we get.

Sneha Mehra (00:01:01)  
So we're gonna have to in our UI.

Sneha Mehra (00:01:06)  
⁓ load this. We want to see that we're making a network request to locations. And here's the request ⁓ that's being made. Well, it's just a kind of like a naked list request. And we're getting an array of empty locations here. So we're gonna trace this through. we we can also see locations is empty. So ⁓ part of this is because we're going to end up creating these locations as part of reading.

A data file. And if we if you look, I'll show you where it where it is, but it's already being ⁓ loaded just so you can ⁓ kind of see see where this is coming from. Temperature ranges dot yml. And so here we've got like a name, region, country, monthly temperatures, and we've got a temperature range, and there's a little bit of a tuple here with the the value and either C or ⁓ F. In this case, ⁓ we

have a little bit of an easier job. Everything's in Celsius. We don't need to worry about converting quite yet. So where is that coming from? Well, we've got parsed data here.

Sneha Mehra (00:02:17)  
And let's see what that looks like. Great. So the file's being read, and you should see in your console we've got name, region, ⁓ country, and then monthly temperatures. But this is just as JSON. We're just reading this file and it's coming through as JSON. So what we need to do ⁓ is ⁓ create ⁓ like we can start by just creating the list of locations, and we should start seeing those pop pop up in the drop down because we can already see the requests being made. We can see like what we're getting back.

Is an empty array here and we should see that start to populate with something. So ⁓ let's let's get to it. Here it looks like we've got a top-level object. ⁓ And actually just to avoid driving myself crazy here, I'm gonna ⁓ get rid of my other log just so we can tell which is which, ⁓ like which is the contents of the file versus which is that location's array. All right, so we've got locations.

Which makes sense if we look at our temperature ranges file, top level property locations there. So ⁓

We could say, ⁓ sorry, going to the right place here. It's the first of the two do's. So parse data is currently of type JSON value. ⁓ what we can do is feed that through a JSON schema to sort of validate that it it looks the way it's supposed to.

Sneha Mehra (00:03:51)  
So we can do ⁓

Up here we could say let's just keep it real simple for now.

Sneha Mehra (00:04:02)  
So we've got location file data is Z object. And we'll import Zod. And we've got locations, which is a Z object. And within that object, what do we have? Name, region, and country, all strings.

Sneha Mehra (00:04:31)  
And then we've got ⁓ monthly temperatures. It's an array of something. I can't really see what's in there in my console. But if we look over here, we've got month. ⁓ And it looks like we're we're going from one through twelve, so we should probably make sure we convert that back, like subtract one. Like this, we want to get a line with the date constructor, so like the JavaScript date object.

So we've got month, which is an integer, and then temperature range. Let's just copy this and make it real easy for us to see up here.

Sneha Mehra (00:05:08)  
Really it's like that. So month. ⁓ and sorry, this was monthly temperatures. Z dot array. And what's in the array? Each member has ⁓ a month, which is a number.

Sneha Mehra (00:05:30)  
Better yet, an integer. ⁓

And we could add some of the validation that you you prompted earlier, Seth. Like we could say it's a min of ⁓ one and a max of twelve the files one through twelve, for sure. ⁓ then we've got ⁓ temperature range.

Which is another object. ⁓ And we've got a min and a max. ⁓ And they're of the same type. And that type is a Z dot Oop. I'm not gonna get very helpful autocomplete if I do it that way.

We could say Z is a tuple, which lets us like who here knows what a tuple is? What's a tuple? ⁓ A pair of two values together, similarly. Could yeah, it could even be more than two. But the the significance is like positionally, you know the types of these things. Like if we were to model this as just an array, it would be array of strings or numbers ⁓ and

That's not quite right. Like we know the first position is a number, and we know the second position is going to be C or F. And so what we can do there is say we've got a tuple ⁓ and we're going to provide the two items. The first is a number.

Sneha Mehra (00:06:55)  
And then the second, we could say it's a string, although to get to line up really nicely with our type checking, what we really want is C or F, right? The union type of literally C or literally F. Well, Zog can do that.

Sneha Mehra (00:07:16)  
Think this has to be an array, but I'll do that in a sec.

Sneha Mehra (00:07:23)  
Here we go. Boom. Little formatting here. ⁓

Sh what is going on? It's a Z, it's an array of these objects.

Month does not exist.

Sneha Mehra (00:07:45)  
We can get rid of this and see what happens.

Sneha Mehra (00:07:51)  
That's the element type. ⁓ classic classic mistake here. It's this. Like I can't just pass an object in here.

It's a zod object like that. You have to give it that nice little wrapper. Oop. Sorry, is that sufficiently wrapped? ⁓ Yeah, duplicate that line.

Sneha Mehra (00:08:16)  
let me take a step back and make sure. Yeah, I I believe you. I just want to make sure I'm doing this in the right order here. There you go. So there's my object. Parentheses, close parentheses, format. Okay, great. So we've got name, region, country. Here's our temperatures. It's an array. Here's the schema for a member of the array. And it's got all this stuff going. So what we can do now is say.

⁓ sorry, what did we call it? Location file data. Parse.

Sneha Mehra (00:08:53)  
Parse data. And what are we gonna get out?

Sneha Mehra (00:08:58)  
file data and ⁓ it's never read, but here's its type. And look at that. And so let's print that to the console.

Sneha Mehra (00:09:12)  
I'm gonna do JSON stringify just so that we can see it ⁓ in its full depth and it doesn't get ⁓ truncated.

Sneha Mehra (00:09:29)  
⁓ and let's see. Get rid of this console log so it doesn't kinda interfere. And look at this. We've got it it kind of looks the same.

Sneha Mehra (00:09:42)  
interesting. Locations expected object received array. ⁓

Sneha Mehra (00:09:52)  
So this is a validation error here. Just checking this. We've got locations, checking our schema.

Sneha Mehra (00:10:03)  
Interesting. Is that Z dot object supposed to be Z dot array? Yes. ⁓ Well but it's an array of objects, so we just want to even round. Yep. Easiest way? Like that. Good catch.

Sneha Mehra (00:10:25)  
⁓ so that's the close of the yeah, just of run.

Sneha Mehra (00:10:31)  
Wonderful. Look at that. We've got like our full depth data there. So let's now turn this into records.

Sneha Mehra (00:10:41)  
So we can close that and we're gonna go like use our little outline here so we can get to the right function. So we're gonna want to get into load temperature data. Here we are, we've got our file data, we've logged it to the console. Now what we wanna do is save ⁓ file data.

Sneha Mehra (00:11:03)  
Locations, ⁓ map, ⁓ and we'll say

Sneha Mehra (00:11:11)  
And then here we're gonna take in one loc data and we're going to what we want to return ⁓ is ⁓ a promise that resolves to ⁓ a location.

Sneha Mehra (00:11:30)  
Is this the right location? Yep. There's the entity right there. ⁓ And of course we're getting all s all kinds of ⁓ nagging because we certainly have done nothing to live up to that ⁓ pr that contract that the function has. And here's our location data. So ⁓ what we wanna do is we're gonna say ⁓

Actually up here we could just get get what's called the repo for the location. This is in in the DDD world and in the type ORM world, that you can think of this as the way that you engage with your collection ⁓ of ⁓ a diff like a certain entity type.

Sneha Mehra (00:12:21)  
And we're gonna say app data source. And remember, this is the file that contains all of our entities here. This is all the things it knows about. And we're gonna say get repository for location. And if we hover over that, we can say it's a repository that's generic over location here. And in here, we can say.

LoCrepo.create and we can pass parameters in here. And what does this return?

Sneha Mehra (00:12:55)  
It returns a location. So we're going to create it and then we're going to save it, at which point we will have a record and we should see things in that drop down. So what do we need in here? We need name.

Sneha Mehra (00:13:15)  
We need region. Or what do we call it?

Sneha Mehra (00:13:23)  
Think I called it the same thing here. Yeah, region. We need country. And most importantly, we need commas.

Sneha Mehra (00:13:35)  
And then let's let's see if this is enough.

Sneha Mehra (00:13:43)  
⁓

Sneha Mehra (00:13:47)  
So what we're doing here is we're sort of like cre we're creating the record. And you can think of this as like you're just creating an instance of the class. And then when we save it, that's where it goes to the database, that's where it gets its ID. ⁓ And you can look at save, it returns a promise that resolves to its location. This is what actually persists things. ⁓ And now you've got a promise that resolves, or an array of promises that resolve to a location. ⁓ now.

This is gonna be interesting. We'll probably end up with like a race condition here, which we'll we'll refactor our way out of it. But we're gonna parse this file and we're gonna kick off a bunch of attempts to save to this database and let's see what happens. It's gonna be like a lot of contention over this table. interesting.

Sneha Mehra (00:14:35)  
Looks like it may have worked. I'm surprised SQLite is able to keep up with that. We'll look at the right way to do it. ⁓

Sneha Mehra (00:14:47)  
We'll look at the right way to do it ⁓ in a sec here. But let's let's take a look at our database and see if if anything happened, what happened. So location. Hey, look at that. We've got seven locations here. So creating seven things concurrently is fine for this, apparently. What you could have also done is like a four a for loop with an await where one by one you're creating each of those locations.

I tried to tease out some lock contention here where we're trying to do many, many writes at the same time. But like here we go. We've got our list of locations. ⁓ our job is to return them. So we're gonna await promise.resolve locations. ⁓ And let's see what happens in our UI. Didn't you want to do promise.all? ⁓ thanks.

Sneha Mehra (00:15:44)  
Hey, look at that. Got a lot of stuff coming back and sure enough, here's our drop down.

—----------------

Sneha Mehra (00:00:00)  
Let's discuss ⁓ this problem of ⁓ organizing seed packets. And I've deliberately given you all more information than you should carry forward because I want I want you to practice this sort of identifying the core value, teasing that out, building what's necessary without going overboard. So ⁓ I'm gonna give you a couple pieces of information to use.

And then I want you to spend a couple minutes just thinking about questions to ask. And then we'll have ⁓ we'll have a discussion. ⁓ And by the way, if you're ⁓ so let's go through the information. I also have a way you can have a discussion if you're watching this course on video. I've built an LLM prompt for you that can sort of simulate a a gardening expert. And you can ask Chat GPT or an LLM of your choice the same kinds of questions we're gonna discuss today, and it should give you something.

useful enough for you to you to engage with. So ⁓ first let's look at what a seed looks like in a catalog. So I've taken this is a huge image for this kind of thing. ⁓ this is a little snippet ⁓ of ⁓ one of those seed catalogs. So we've got like this is for a tomato, we've got San Marzano, Heirloom, 80 days, then a description, and then this is a wait, and there's some words here, and then we've got different ⁓

Quantities that you can buy for a price. ⁓ And here's the front and back of a seed packet. So we've got the front where ⁓ got some information up here. And then in the back, ⁓ this is even information. And then we've got some measure of days, ⁓ some planting instructions.

And then don't forget to look on the side here, this vertical like edge information here. Sprouts in seven to fourteen days. There's an ideal temp, hey, a temperature range. ⁓ seed depth for planting, whether it's frost hardy, something called plant spacing, right? So ⁓ do not attempt to model every detail of the real world. Don't lose like you have to keep track of the fact that you are.

Sneha Mehra (00:02:22)  
Trying to build a useful software product. You're not trying to create a simulate like this is not the matrix where you're trying to create a seed packet and a seed that will actually grow a plant and that is a perfect representation of this thing as it exists ⁓ in the real world. And I've given you a seriously overkill data source here that you're going to have to pick and choose from. This represents ⁓ one seed. So we've got like some spacing stuff, we've got some environmental needs.

Like how much sun does it like? How much water does it like? We've got some soil characteristics that it wants, some temperature ranges, right? So you're gonna you're gonna wanna look at this. And let's say like you had access to a government data source that that p has a ton of this information about seeds. I want us to work on like what do we really need to tease out in order to make this useful? Like clearly, I mean there's some stuff back here. This is mostly placeholder stuff we can

We can tweak it based on our discussion. But like you're not going to fit all of that back here on the back of the seed packet. So we'll have to refine that.

Sneha Mehra (00:03:32)  
All right, finally, if you're watching remotely, ⁓ or ⁓ if you're watching this ⁓ as a a video course, ⁓ this is an LLM prompt that you can plug into ⁓ a a ⁓ chat, you know, AI of your choice, ⁓ and ⁓ it will ⁓ it will be your gardening expert. ⁓ it'll it'll have a garden that's similar to mine, so you'll get hopefully like some kind of similar answers that I would give. ⁓

But this is a way you can sort of have a discussion, ⁓ you know, even if you're watching this course at home. So with that, let's take a few minutes ⁓ to sort of look through some of this information I provided you, and then let's come up with some good good questions. And what we're gonna do is go back to this tool, and we're gonna try to model like what exactly is a seed packet and how do we how do we break that down? Like what are the value objects, et cetera.

⁓ there is a ⁓ file in your repo that I'll give you a path to.

Packages, server, data, seeds dot y ml.

Sneha Mehra (00:04:48)  
And that is ⁓

This path is where you're gonna find a huge data file that's kind of our source of truth for for seed data. So if you look at that, like here's that tomato plant, ⁓ and ⁓ you know, VS Code makes this nice because you can still see the IDs in the path, but like it it goes pretty deep. Like this is a 12,000 line YAML file with a ton of data. All right.

So take take let's say two or three minutes ⁓ and just like process some of this and let's let's prepare to have a conversation.

Sneha Mehra (00:05:36)  
I'm now putting my gardening expert hat on. So the scenario is ⁓ I have a problem. I have a ton ⁓ of of envelopes of seeds at home. ⁓ And ⁓ I have reported to you that it's just it's like a unmanageable pile. And that that is painful for me. That is difficult. So how can we go from that to like

Identifying kind of like a a problem to solve, like a more specific problem to solve, and then going over here and figuring out how that translates into entities, relationships, ⁓ values, constraints. Can you describe ⁓ what specifically you struggle with a little bit? Do you have trouble knowing what seeds you have on hand? Do you struggle with ⁓ remembering what specific time you need to plant certain seeds?

Maybe some more information like that. So ⁓ the question was like, can it go into more detail? Like are there specific things that I struggle with? You you ⁓ you had some good prompts there. Like, are there things I need to know at a particular point in time? So you're you're fishing at like, is there a life cycle here? Is there a sequence of ⁓ of time based operations? And it turns out there are. So I'm gonna use a different color here because these are not these are not entities, but there's ⁓

There's like, I wanna plant stuff in the ground. ⁓ And this happens ⁓

When the time is right. You should probably ask me more about that. ⁓ so we need to we we plant in the ground and then like before that we we harden. Harden plants, and this means like gradually expose them to the outdoors. ⁓ They're really delicate when you start them inside, and you want to give them a few hours of real outdoor sun and wind so they get strong. And then there's a part below that, or before that, sorry, which is ⁓

Sneha Mehra (00:07:45)  
Grow indoors. ⁓ And there is ⁓ time here. ⁓ Right? So I want to know like how long these things take to mature. Well, it's almost like time that works backwards, right? Like, I want to know when does this plant go in the ground? ⁓ And this one's pretty known.

Seven days. Harden for a week. This one depends on the plant. Some things ⁓ like ⁓ mint will just shoot up. ⁓ and some things take a really long time ⁓ to to develop. ⁓ like peppers. Peppers take a lot longer to grow than squash, which kind of just like ⁓ takes over everything.

Like you start it inside and it's like unwieldy as you're carrying it out. So like there's a timeline here. There's a sequence of things I want to do. And like because there's like a variable here, we can call it like T, right? Like the time it takes to grow, and then there's a variable here, let's say that's temp. Sorry, too many T's. There's a time and there's a temp. And so like based on when this goes outside, I'm looking for particular things.

at different points when I'm growing indoors. It's almost like a work back plan, if that makes sense. Like we want to launch on a particular date.

There's a fixed time thing here. And then like I need to work backwards ⁓ and figure out like when do I start this inside? Some things I have to start really early. And then some things I can start later, you know, ⁓ and and it'll be fine. So that's that's like one of the things I want to do. ⁓ I also so let's let's capture this ⁓ as ⁓ just like that. There's a life cycle. There's a time-based sequence.

Sneha Mehra (00:09:53)  
Alright, and there's something else. Like, seeds also expire. I happen to have like sometimes I order too much. ⁓ part of it's like I'm legitimately disorganized with this collection of seeds. I I forget that I have a thing and I order more of it. ⁓ And now I have an old seed packet and a new seed packet, and I want to go through the old stuff first. So I there's some sense of like, show me if I'm if these things are getting close to where the seeds aren't gonna grow.

As easily. And I want to go through those. I want to make sure that I use those up ⁓ rather than use the new thing. And then finally, if I I want to look through what I have so I don't double or triple order things. ⁓ So just being able to like easily find what I have quickly rather than sort of rifling through a big box of envelopes, that's pretty important. So there's like search by name or plant type.

There's ⁓ a time based sequence, and so I want to understand.

When to start indoors. And then ⁓ and then like there's sort of a favorites concept here.

Stock are you saying? How much you have stock of them? Yep.

Sneha Mehra (00:11:18)  
That's great. How many seeds of a given type do I have? Do ⁓ seeds tend to expire sort of at the same is there like different expiration dates for different seeds? How long they'll last? Yeah, that's a good that's a very good question. ⁓ I'll give you like the first level of insight, which is I just there's the date that's on the packet, and I'm not sure what the answer to that is. But here's if if we dug deeper.

It's something like

Well, the way you get seeds ⁓ is you grow a plant and you wait for it to mature. And sometimes it'll like it'll fruit, right? Like you get a pepper. But instead of like eating the pepper, well, maybe you also eat the pepper, but you collect the seeds from inside the pepper. And so that's happening at a predictable time. Like that's usually happening with with seasons. And so when I think about like expiring, it's it's usually less about, you know.

It's on March 27th that this expires and I don't want to use it after that point. But it's a softer signal there. Like, look, these are getting old, and I'm really looking at expiration dates, kind of around this part of the time period. Like when I'm planting things indoors, I will totally plant seeds that are past their expiration date. But I'll like I'll put two or three of them in those little hydroponic ⁓ those those baskets that we we talked about.

And like, I don't know, maybe one grows. Sometimes none grow. But like it's it's it's a it's sort of a I'd say we're kind of honing in on it's more of a year that it expires on, and it's something I only evaluate at the beginning of the season. Someone online said, Yeah, reminders or out of I need a reminder. That's great. That's great. So ⁓ we do want a reminder anytime there's a time based thing. Like remember, I've got like

Sneha Mehra (00:13:15)  
40 of those raised beds. I keep saying it's like 38 or 40, I legit have lost count. ⁓ it's a lot to manage, a lot of different plants going on, a lot of things starting at the same time. And really I'm I'm well beyond the point of just putting a calendar event somewhere manually and being like, this is the day where I'm performing this activity for my garden. It's just all these things have different timelines. So yeah, great reminders.

Yeah, yeah, there's a time-based sequence. This is the reminder right here. I want to understand when to start indoors. I want to understand when to transplant. Okay, I think we're good. I'm going to stop us here. So, back to seed packet. And as we look at these things, like, what are some pieces of data that we need to associate with the seed packet? ⁓ And let's think about are they value objects? Are they other entities? Like, let's explore.

Like okay, s maybe you understand like the pain. Let's start thinking about how to describe the related concepts here.

Not not in code yet, but just you're still on the whiteboard with the gardener or you're using whimsical with the gardener.

—----------------------

Sneha Mehra (00:00:00)  
And we've already got plant in there. So let's take a look at what plant is. ⁓ It's another like metadata schema wrapped thing. Ultimately, it comes down to this class where we have planting distance already. ⁓ And ⁓ let's let's see if we can sort of work our way up from you know building some of these server-side models and see how far we can get. Let's like get to a point where we've persisted some records and we can see we can console log a data structure out.

So ⁓ so a lot a lot is already here wired up for us, but we're gonna have to add some more interesting data as we as we choose to tackle more. So I'm gonna focus back in on server ⁓ and in my entities folder. And we've got garden, ⁓ and it's just got a name and description. So we're gonna need a garden bed, or we called it just bed. Bed in this in this bounded context means garden.

And then we'll call this a well we already have plant, but let's get garden bed implemented. So let's use garden as a starting point. And we're just gonna say bed. ⁓ And ⁓ this is us just customizing the name of the database table. You don't have to do that. And we'll change the ID prefix to bed. And now ⁓ we want to ⁓ establish a relationship.

And this is ⁓ many to one.

Sneha Mehra (00:01:34)  
This is gonna be garden.

picking the entity here, not the thing from the types package. It's definitely assigned. ⁓ And it's gonna need the the inverse relationship. So we'll have to set that up over here. And here we'll have ⁓ one too many.

Sneha Mehra (00:01:56)  
There's bed, and this is how to get from bed to garden. ⁓

Sneha Mehra (00:02:07)  
And resolve the import. Let's see.

Sneha Mehra (00:02:14)  
we just return the type there. Not we don't invoke a constructor with parentheses. Right, now we can finish the other side. So this bit the factory is garden, not invoking it. And then ⁓ this is garden.

Sneha Mehra (00:02:33)  
and had like the the opposite side of that relationship. So great. We we've got this set up. ⁓ let's go back to our domain model here. ⁓

We need a width and a height.

Sneha Mehra (00:02:49)  
So ⁓ let's ⁓ let's implement that.

Column.

Sneha Mehra (00:03:06)  
⁓ let's let's skip it a lot.

We've got a width, we've got a height.

Sneha Mehra (00:03:18)  
⁓ And plant has a position. Plant has a relationship with the seed packet. Let's see if we can just get garden beds rendering on the screen. Like remember, we've had sort of we got the workspace with zones, and that's an array. Right now it's hard-coded to an empty array. And then each zone has like placements ⁓ of items within the zone. But let's see if we can just get zones rendering on the screen. And then and then we'll look back and like get those get those plant tiles coming through.

So ⁓ to make this work, we're going to have to go to ⁓ the domain service, right? That's that's always like the source of truth for producing ⁓ gardens, ⁓ producing whatever the entity is. So we'll go into our services, gardens service.

Sneha Mehra (00:04:15)  
Great. We got ⁓ get all gardens. So this this will work like once we have gardens in our database, this will return them. We've got get garden by ID.

And then we've got ⁓ create example garden. ⁓ And I can see here we're using the garden repo. We've got this like ⁓ assertion against the deep partial of the you know of the type, so we get nice field level errors that pop up instead of the whole shape turning into red squiggles if something's wrong. And then we save it, and like presumably this is what's passed up into the UI. ⁓

⁓ create example garden is called when the app boots. This is sort of just like getting us our starting point. ⁓ And ⁓ now we have a beds, ⁓ a beds array that we could create. And let's let's start our server or resume it.

Actually, let's restart it and see if there are errors. We should be good. ⁓ what's going on? ⁓ I know what's going on. Entity metadata for garden beds not found. So this means go into your data source and add these new concepts. Bed and import it like that. Save. ⁓ And there we go. Look at all the plants servers running on localhost 3000\. So now our data source knows about the concept.

bed ⁓ and ⁓ you know we're we're setting this up now we haven't really done anything that makes this like any different like that's would have been an empty array either way but ⁓ at least we're sort of exercising that field great so let's ⁓ after we create the garden and save it

Sneha Mehra (00:06:14)  
Let's create some beds. ⁓ And we need a bed repo.

Sneha Mehra (00:06:25)  
Like that.

And we can say bed repo dot create.

Satisfies Deep Partial Bet.

And then

Sneha Mehra (00:06:45)  
Calls bad one.

Sneha Mehra (00:06:55)  
⁓ bed repo dot save bed one. ⁓ And that should save the bed.

Now if I hit save, we're gonna see a bunch of errors because like we have not given the bed what it needs. Like it it there's a lot, a lot we need to define there. Let's see what the error says. Okay, it needs a width and it needs a height. So let's make this like a six by six.

Sneha Mehra (00:07:27)  
All right. ⁓ And ⁓ it also should be associated with ⁓ a garden.

Sneha Mehra (00:07:38)  
Use a little shorthand there. Great. So ⁓ and let's let's ⁓ see what the API looks like when we hit it now.

Sneha Mehra (00:07:55)  
⁓ sorry.

Sneha Mehra (00:08:00)  
Interesting. So we're still not quite there yet. Let's let's follow this through from our domain service into the routing layer.

So that's gonna be in our application. Gardens router. ⁓ look at this. ⁓ Nothing here, right? So this is where we need to say, ⁓

Sneha Mehra (00:08:28)  
Actually.

Sneha Mehra (00:08:33)  
Did we remodel bed yet in our types package? I think we did.

Here, types, source. We may not have exported them yet. There's the garden type.

Sneha Mehra (00:08:50)  
Yeah, I don't see a bed here yet. So we'll need we'll need to wire that up. But for now, let's let's see if we can just like thread something through as just a very generic ⁓ zone.

Like an array of zones. And so that means

Sneha Mehra (00:09:11)  
We're gonna say ⁓ gardens. no. It's garden.beds dot map.

And this returns the zone.

We don't need this anymore.

Sneha Mehra (00:09:36)  
And this will be a bed. Bed dot ID. Oops.

Sneha Mehra (00:09:45)  
With

Sneha Mehra (00:09:49)  
Hi.

Sneha Mehra (00:09:53)  
Alright, and what are we missing? We need ⁓ name, description, and placements.

Sneha Mehra (00:10:10)  
Leave that empty. Placements. Let's leave it empty as well. Again, our goal is like come come up for air here after we've ⁓

After we've introduced this concept of a bed. Now what's going on here? Types of property zone are incompatible.

Placements are incompatible.

Sneha Mehra (00:10:38)  
Well we can ⁓ let's see.

Satisfies. Deep partial. We're just trying to get a an error message that's more local.

Sneha Mehra (00:10:56)  
Great. Okay, so it's tracking us down to zones. ⁓ item dot metadata are incompatible between these types. So let's see if we can add some metadata here.

Sneha Mehra (00:11:10)  
Nope, that's not gonna make it happy. So what we can do instead is say ⁓

Sneha Mehra (00:11:21)  
Can we cast it like that?

Sneha Mehra (00:11:31)  
Type of property placements are incompatible.

Sneha Mehra (00:11:40)  
You know, we're gonna cast our way ⁓ out of this for now.

Sneha Mehra (00:11:51)  
No. Interesting. So it's on the item metadata, so it's really opinionated about this. What if we get rid of it? Nope.

Sneha Mehra (00:12:12)  
Yeah, that's fine.

Sneha Mehra (00:12:17)  
Is it still gonna yell at me about placements?

Sneha Mehra (00:12:24)  
it's the planting distance. It needs a planting distance on the items. So let's let's just create like a very ⁓ a very generic placement object here.

Sneha Mehra (00:12:39)  
ID ⁓ one position ⁓ X ⁓ one ⁓ Y zero.

Sneha Mehra (00:12:53)  
⁓ And

metadata.

Sneha Mehra (00:13:01)  
Well we can remove the metadata for now if if it's too much in our way here. ⁓ It

Sneha Mehra (00:13:12)  
It needs an ⁓ item and a source ID.

Sneha Mehra (00:13:43)  
Hmm. ⁓ sorry. That's it. And this is gonna be our ⁓ like a bed ID. But let's see if we can leave it as empty here ⁓ and the item.

Sneha Mehra (00:14:04)  
Okay, let's see what's coming through now. ⁓ cannot read properties of undefined reading map. Garden, beds. So let's serialize this.

Sneha Mehra (00:14:19)  
Just print it to the console.

Sneha Mehra (00:14:26)  
See what we get.

Sneha Mehra (00:14:31)  
So there's our garden. Beds is still undefined. Interesting. I'm gonna check our database to see if we've actually got a garden bed in there. We do. The garden ID should be set. So it's with 03\. ⁓

Yep, that's correct. ⁓ the relationships. We need to traverse the relationship here. So we're saying get all gardens ⁓ and here we go.

Sneha Mehra (00:15:04)  
Now it'll fetch the beds.

Sneha Mehra (00:15:10)  
And our API should return something more interesting now. Yep. So there we go. We have a zone. We have a placement in there with X and Y. And let's see how our UI responds to this. Yep. I I expect like it's it's still gonna need a little bit more ⁓ because we've very much fudged this placement's ⁓ object. So we're gonna need to keep carrying this through and do these do these plants.

—-------------------

Sneha Mehra (00:00:00)  
Let's take a step back now and think about how we want to model this. Like right now, if we look at location, it's ⁓ sort of boring. Like it has a name, a region, which could be like a province or a state in the US, ⁓ and it has a country. But there's there's nothing here that deals with this concept of like ⁓ answering the question we're trying to answer. Like, when can I put my plants outside? When is it going to be 50 degrees or greater? So ⁓ let's think about how we would want to model this.

Does anyone have any ideas? Like if we were to start with location and this has things on it like name, ⁓ what do we call it, region?

Sneha Mehra (00:00:45)  
Oops, sorry. We've got name, we've got region. Guess it's not gonna let me do multiple bonds space things.

Sneha Mehra (00:00:59)  
But like, what are what are the other things that probably need to be part of this story?

How would you all think about this? Like well t a temperature doesn't exist yet, so we need something like a temperature. What should a temperature have on it?

Sneha Mehra (00:01:20)  
Use Unit Max.

Unit. Unit? Yeah. What kinds of things should I have for units? Yeah, we're we're sort of you're looking at like the API contract there, which is good. There's are there's some existing software. Like unless you're starting Greenfield and building something entirely new that touches no other systems, like sometimes you can look for clues in terms of what already exists. And then ⁓ I heard Value Value. And that's a number. Great.

So we have that. Let's pret pretend I'm the gardener. I'm I've asked for this. Like I want something that makes this work. I want I want this tool to work. And I and I hired a designer that made this for me. So like S ⁓ that interaction between the date and the temperature. The date and the temperature.

Say say Mona, you're on to something. ⁓ I'm not sure. I mean a min and a max would be the start of it. Absolutely. A min and a max. So what should we call this thing?

Limit? A limit? You can call it that? Is that is that the thing with the min and the max or is it range? ⁓ range. ⁓ A range has a min and a max, sort of the upper and the lower limit. So great. Why don't we call this a temperature range?

Sneha Mehra (00:02:50)  
And we could say, I heard men and mechs.

Sneha Mehra (00:02:57)  
Okay, ⁓ and so clearly like we've got something here. We've got min and max. So we'll represent those with two two arrows. ⁓ And it's it's not quite like a belongs to because remember, like this here is probably

Sneha Mehra (00:03:18)  
It's probably a value object, like temperature, because I don't think we're gonna store like 56.3 degrees Fahrenheit in our database somewhere. It's probably just sort of something that gets embedded on something stored in a database. So I'm gonna make a nice little tag here. So this we're gonna say is a value object. ⁓ And we'll make this a different color.

Sneha Mehra (00:03:49)  
Alright, ⁓ what about temperature range? Is that a value object or is it an entity? And I'm gonna check chat here.

range and location, great. Yeah, Santiago, you're on the right track. Entities are in the database and values are sort of embedded on the things in the database. So temperature range, is this is this something that will have like an ID? Like, how would we articulate this in in ⁓ spoken language? I think it would always be associated with an entity. I think so. Right? This is sort of like a weather report range. Like the low is ⁓

The low today is 58 and the high apparently is 82 in Minneapolis today. So this is this is probably also a value object. Like it has to it has to be embedded on something. ⁓ And ⁓ so could can we just do this?

Can we say like a location has a temperature range? Like Minneapolis has this low and this high. Like does that get us all the way to being able to estimate a date? Pardon? Like some measure of time, like season or month? We need we need some measure of time, right? We need to know, like

At this point, here's the expected min and max. And at another point, here's the expected min and max. That would almost let us think of this as a curve, right? Where we're trying to find the like if you think of the temperature curve throughout the year, let's say, and we want to have a threshold or like start it's above 50 degrees starting on this date. Like where does my line intersect with the temperature curve? Well, we need some measurement of time. And so let's call this a ⁓

Sneha Mehra (00:05:36)  
Monthly.

Sneha Mehra (00:05:40)  
Temperature.

And then this will refer l let's just say this is like a month.

Sneha Mehra (00:05:51)  
And a r a range.

Sneha Mehra (00:05:56)  
And this is going to end up being an entity.

Sneha Mehra (00:06:06)  
Right? This this we'll store in a database. Like you could imagine us reading from some data source. We have our locations, we get our temperature data, and then at runtime, we're trying to figure out like, all right, you've asked for for this this ⁓ 50-degree temperature, like where where exactly does that fit? And so really, it's gonna be something like this. Sorry, let me rearrange these slightly.

Sneha Mehra (00:06:36)  
This is a nice quick and dirty tool here, but sometimes the arrows get a little bit weird. Okay, so ⁓ so monthly temp has a temperature range, which ⁓ which sort of or you can think of it as embedding a temperature range, and that embeds two temperatures. What's the relationship between ⁓ location and monthly temperature? It seems like there's a relationship here.

Is it one to one? Has many. Has many. Has many. ⁓ And we can represent that.

Let's say it's one or many. Really, we're in trouble if we have like less than twelve months of data here. So location, which is an entity.

Has many monthly temperatures, ⁓ each of which embeds a temperature range, each of which embeds two temperatures. That's a that's a very good, very good point there. Like we have multiple ways we could represent this data. And this, you're getting into what's called normalization. ⁓

Right? Like if we let's say we're got this temperature data from different sources. Some of them are Fahrenheit, some of them are Celsius, some of them are are ⁓ Kelvin, ⁓ which ⁓ maybe like do you think Kelvin is what a f vegetable gardener is probably thinking in terms of? Yeah, I mean you don't have negative numbers with Kelvin, right? Right, right. Like that's this where like absolute zero is the lowest possible temperature. That's a zero Kelvin. So that's you don't see those that that

Sneha Mehra (00:08:16)  
The kind of thing on a weather report, but it's totally a valid way to represent temperature. So we have to decide like how are we storing this information? And then how are we representing it in the UI? Like ultimately, I I see I get a little like Fahrenheit Celsius thing here, and presumably that's gonna define like I I noticed when I hit this, like look at the little label here. Like it's it's changing. So at the very least, this is how I'm describing my input.

Presumably, I want to see my output in the temperature I asked for, like in the units I asked for. So that is certainly something to consider. But like where it would matter most is if you wanted to have charts or something that showed that temperature curve. If you're storing data sometimes in Fahrenheit, sometimes in Celsius, sometimes in Kelvin, you're creating a lot of work for like anything that is built on top of this. They will have to perform this like normalization task, right? They'll have to translate it into something that's very consistent.

across all of the locations, irrespective of you know what the data source was.

Okay, let's jump in and and like I I buy that this, like as as your resident gardener, I would say this makes sense. Like I'm looking for like a day, I don't need hour by hour, minute by minute, like when the temperature is gonna be what it is. And especially if I'm like in January getting trying to figure out when do I start my plants, like there's no way, like last year's day by day temperature ranges are not going to be accurate.

I want the monthly averages. That seems fine. Right? That's that's it's a prediction with some error bars around it at best. I'm not really following what makes monthly temp different than temperature range, thus that we want to make it an entity versus a value. ⁓ Right. ⁓ we totally could do that. We could say.

Sneha Mehra (00:10:15)  
That's a very that's a very good point. Here's an alternate representation.

Sneha Mehra (00:10:24)  
Does anyone see anything better or worse about this?

I'll give you a better thing.

It's one less thing to build. That's so by default by default it's better.

Sneha Mehra (00:10:43)  
Is there a downside? How would you talk to your user and and understand whether you should do what we modeled before or this? Like how would you answer that question talking to your customer? Let's say you're the you're the you're the tech lead on the team, and it's like it's up to you to figure out like which which is the right pattern. Are there some questions you might ask? If there are other places you might use temperature range, ⁓ you're gonna reuse that type for temperature range and not have to recreate a new ⁓

value object for temperature range. That's absolutely true.

Could be. And at this point, it kind of comes down to like, how confident are you that you're going to run into that? Or, you know, is it okay to start here and sort of tease out that abstraction ⁓ in the future? ⁓ And I mean, at this point, temperature range doesn't really add a lot of value here. Yeah, my question was more not like consolidating those two types, but

What where the line between a value and an entity really lives? Because ⁓ obviously I think you don't want to be storing temperatures as like individual like columns in a database, their own IDs, it doesn't make any sense, especially if they're afloat. But you get to a point with like temperature range, like how granular are the locations and do you really if each location's gonna have its own range, should that be an entity or should that be a value?

'Cause that kind of seems like an entity should maybe be reused, but I'm not entirely sure. ⁓ all all great all great points to think about. Like do you do you agree that or do you think we need this here? This one to many range. Like sorry, one to many relationships.

Sneha Mehra (00:12:33)  
Like I I would argue any valid solution has to have this relationship. Like locations ⁓ have temperature information at a point in time. So no matter what, we need this. ⁓ I think it's kind of it at this point, we're just really discussing ⁓ kind of like the structure of whether there's this additional level of one more value object that represents this this range. It's it's almost like a tuple.

Of temperatures that represent like the low and the high. We could have other places where temperatures exist. Like, I don't know, maybe the tomato seed packet has like minimum outdoor temperature and it's just 50 degrees. Maybe there are times where we have, we'll find either a temperature or a temperature range. Some seed packets say 50 degrees and some say between 45 and 55\. But at this point, I think this is.

This is a simple representation, as kind of as simple as we can get away with for now. And we should bias for that, right? Like there's sort of the rule of find two or three places before you tease out an abstraction. Otherwise, you can get into ⁓ attempting to model all of the details of the real world. And that's sometimes a trap. Like you can end up overbuilding things ⁓ for that purpose. ⁓ Nikita asks, ⁓ sh

Should an entity have an ID? Does it make sense for temperature ranges to have IDs? You know, I I forgot about this.

Entities should certainly have IDs. We need an ID there. ⁓ I was I was sort of taking that as a given, but we should we should list it out. And then in this case, monthly temperature ranges should have an ID.

Sneha Mehra (00:14:18)  
Or string. ⁓ Whatever we choose to use.

—------------------

Sneha Mehra (00:00:00)  
Now let's dig into another domain modeling exercise. We have this concept of a garden, ⁓ and ⁓ we have this concept of raised beds. Just, you know, remember that photo I showed where it's like those metal things filled with dirt, plants in them. ⁓ and I have some statements here. These are like user testimonials. These are things, ⁓ pieces of pain, or like elements of pain that a user has to deal with.

or they they reflect, you know, something they wish to do that is difficult for them to do today. ⁓ I've made these intentionally vague, ⁓ and I have used maybe some ⁓ terms like square foot garden. I'm not sure everybody who's taking this course understands exactly what that means. Perhaps some of you in the room don't understand what that means. ⁓ Like, this is a cue for you to tease out more, ask questions, see if this needs to be part of the common language.

language we use in this gardening domain model.

⁓ so your task, let's let's do that ⁓ the the domain meddling exercise around this ⁓ this app. Sorry.

I mean we kind of already know like l a little bit about where we're going. Oops sorry, I keep going back to the other one.

Sneha Mehra (00:01:26)  
We kind of know that we're we're going here. But actually better still, let me take you back to this. ⁓ I'm trying to get organized around these things. And so let's let's see if we can have a conversation. How do we represent this ⁓ in software? What are the what what what's at the essence of this? The things that we need to capture. And let's start with the things that we feel are ⁓ necessary for any valid solution to this problem. And and

To be clear, like we're departing from seed packets now and we're focusing on the layout of plants in a garden that is is mostly focused around raised beds.

Sneha Mehra (00:02:11)  
So I'm gonna I'm gonna start us off.

Sneha Mehra (00:02:16)  
Clearly.

We're gonna have a garden. Alright, that's that's the easy part. And this is ⁓ for sure an entity. Like a garden has an ID, maybe I have many gardens. ⁓ Maybe this turns into like a ⁓ garden organization as a service app where neighborhoods can rally around this.

How come how can we try to represent in entities, value objects, relationships, constraints?

Sneha Mehra (00:02:50)  
Raise beds. ⁓

Sneha Mehra (00:03:01)  
Each race would have plant plants. ⁓ Yep. What do you mean by plant? A array of plants.

Sneha Mehra (00:03:12)  
So Claire, like we'll need a plant. What does what does a plant mean? ⁓ Dropping a seed and then like you can add multiple seeds in a raised bed and then So th this is like one plant planted in the ground at a particular time. Great. Not the abstract concept of like there exists this concept of a ⁓ blueberry plant and I can sell you many of these things, but this is like I've handed you one, right? ⁓

How about a planting area for the raised bed? ⁓ Is that sorry, a new concept or renaming this concept? That would be a new concept of value entity the raised bed has. A planting area. Okay. A value object or entity?

Tell me about the difference between planting area and raised bed.

I mean, I guess the width and height

Could just be with heightened depth, could just be properties of the raised bed. Unless you're planting things in the garden outside of a raised bed, and then a planting area could apply to anything that isn't a raised bed in addition to describing the areas contained within. Now we're now we're thinking critically about this this problem space. So I'm gonna I'm gonna say, you know, I I actually call those things on the ground, I call them beds. Not like yeah. ⁓

Sneha Mehra (00:04:39)  
Some I have some raised beds, but I also have like these these little beds on the ground. Like it's sort of a a ⁓ space that plants go into. So maybe the fact that it's raised doesn't need to be modeled. Now this really makes things interesting. Depth. Like ⁓ what is the depth of a ⁓ bed on the ground? Well, like, is it the the radius of the planet Earth? Like, what are we talking about here? ⁓ so I I love this. And

Because I actually intend to use this app, and like up ⁓ at the end of this, I'm gonna show you some some like tickets that I've I've prepared if people want as like a final project to do something interesting in the context of this app. But like depth starts to become really interesting if you're planting something like bulbs, where you can have different layers. You have ones that can be planted deep, and then ones that can be six inches above, and they'll sort of all pop up at different times. This is like

A lasagna of bulbs, if you want to think about it that way. ⁓ let's stick with width and height right now. I ha I I notice like it's a top-down view, ⁓ as far as we're we're concerned right now, and we're not trying to to get into depth quite yet. But a real solution, ⁓ a robust one would certainly model it that way. I would

So ⁓ without without like making final decisions here of of any kind, like what what other things do we want to incorporate? So like garden clearly has many raised beds.

Sneha Mehra (00:06:15)  
I guess location ⁓ could be on the garden. ⁓ There might be some sort of like X and Y situation for the beds within the garden. I don't know if that's important to this. Yeah, well I absolutely.

Right? Like thinking back to the first exercise we did, like if I want to answer some of these questions around like when is the right time to plant these things, location is absolutely critical. I'm gonna bend your X and Y point here. I I'd argue at the very least we need to know on plant, like in the context of a bed, where is it?

Sneha Mehra (00:06:59)  
And we're gonna need a position here. We happen to have already an a a value object called XY coordinate, which will will give us a good indication there. ⁓ so does a planting area have a position within a bed? ⁓ In which case would you wanna tie the plant to the planting area and position? ⁓ Yeah. So stepping into my subject matter.

Expert role here. I'm gonna say I can't really tell the difference between these two things. Like to me, a bed is a way to describe a planting area.

Is the planting area not one of the tiles within a bed we can ⁓ drag it into? ⁓ that's a good question. Well, let's forget that we have seen a peek at a solution there. But if you were to ask me that same question in the context of like how I think about a garden. Well, I have my garden, it's like the big photo I showed you, and then I've got beds, some of which are on the ground, some of which are in those ⁓ metal containers, and then

Plants are positioned within those beds. And like while it's true that each of those beds has a position within my garden, really what I'm trying to the the problems I'm trying to solve here have to do with like tracking ⁓ how much sun each bed gets. Or am I planting things too close together? ⁓ Or am I needlessly planting things too far apart and I'm wasting space? And if you just show me like ⁓ some

List of little bed diagrams, like a lot of my needs are met and I don't need it to be I don't need to be at like a a a two scale real diagram of of my garden. Although that would be t very cool. I'd say we're going well beyond the basics if we go all the way there. Would we have any plants that exist outside of a bed? Like like seedlings ⁓ or ⁓ yeah, great question.

Sneha Mehra (00:09:00)  
I think I can simplify that by saying what whatever a plant exists in, I'm gonna call that a bed. If it won if it's one of those like hydroponic things, ⁓ that's a bed too. If it's in a red party cup, maybe that's a a bed of width one and height one, and I'm just gonna call that call that a bed. A good it it's a like it's a good starting point. So I think what what maybe what we can get out of that is like plant always.

always has ⁓ a pointer to a bed, right? And beds can only exist within the context context of a garden. So what I'm gonna do is I'm gonna combine these because your s your subject matter expert says I'm I'm confused here. the plant would have a size. ⁓ What's okay, say more? ⁓ like how big it gets. ⁓ How big it gets. Maximum

⁓ How many positions with then ⁓ the planting distance. ⁓ The planting distance, right? Yeah. Now wouldn't the plant just have kind of plant metadata, which is all of the stuff that the plant had to turns out we're going to use the same the same met plant metadata concept. I'd say yeah, absolutely. It has a planting distance.

For sure.

Sneha Mehra (00:10:25)  
And you know.

Sneha Mehra (00:10:29)  
I'm gonna tell you I only care about feet here because I mentioned ⁓ I want to embrace the concept of a square foot garden. So really, like in reality, when we say there's this plant distance, ⁓ it's a circle. It it kind of represents like how big the plant's gonna get, ⁓ how competitive it is for nutrients, ⁓ you know, with the things around it. But gosh.

A real simple way to draw this out is squares. And so we're gonna approximate it. ⁓ but that's the like square foot garden is is a term people use where like you can allocate a square foot, and and sometimes that means you can plant like nine carrots in there in a three by three grid. Or if it's a tomato, it's like a two foot by two foot square in your garden. It takes up all four of those spots. And that that's sort of like the the crude way of getting it pretty much right.

And a lot of people do it this way.

Alright. ⁓ plant has a position. It has a planting distance. ⁓ I'm actually gonna do this.

Sneha Mehra (00:11:45)  
come on. You're not gonna let me scroll? I'm gonna say a plant comes from a seed packet.

Sneha Mehra (00:12:00)  
Ultimately it comes from a seed back and I I I think we already put planting distance on here. So I'm gonna refine.

Sneha Mehra (00:12:14)  
I'm gonna refine our domain model here. We've made a change. We actually added this already. And ⁓

Sneha Mehra (00:12:25)  
We can just refer to the seed packet as sort of like the source of truth for ⁓

That's that's sort of our like it's our plant factory, right? It's the template for a plant. And you know, that's we don't need that data to be stored on every single plant that's in the ground. Like if I plant forty snow pea vines, I can point to the same seed packet. Maybe outside the scope of the exercise, but like with I don't know a lot about gardening, but I have heard the concept of like permaculture, right? Where d certain plants planted near each other can be more like cooperative and effective and grow better.

Do you as a master gardener with the Master Gardener hat on, do you care about that relationship for plants in how you plant things? Absolutely, but I think it goes beyond the basics here. I would I would love for that to be eventually part of this app. But for now, let's like let's focus on plant spacing as the core problem. Plant spacing and how that is dealt with in the layout of plants ⁓ in a bed of some kind.

We can also assume the beds are rectangular for now, you know, ⁓ and and we'll we'll see where this takes us. So I I think this is a good starting point. We'll revisit this if as we're implementing we discover that there are things that we're we're missing.

—--------------------------

Sneha Mehra (00:00:00)  
Let's jump into our project and start ⁓ and start implementing this. So where we're gonna start is ⁓ in the types folder. We're gonna go into I'm gonna close server for now. And we're in packages, types, source, ⁓ and ⁓ we've got a folder here called entities.

And we've got a folder here called value objects. ⁓ So I think we have a temperature value object to implement here. And let's start there. I'm going to create a new file. Let me close some of this ⁓ stuff, which we'll pull up later. You can leave those tabs open ⁓ because those are files we're going to need to touch. Temperature. And I'm just going to follow the convention here because we're going to have another temperature TS that exists elsewhere. Temperature.type. ⁓

⁓ And ⁓ let's let's start with like copying and pasting some code from an existing value object. We've asked should we switch branch? Or what branch should we be on? ⁓ you should be on you should be on the DDD start branch. ⁓ D start.

Sneha Mehra (00:01:19)  
I'll make sure that we keep that where it is, even as this project evolves. ⁓ So we've got a temperature type here. I'm going to start by grabbing this distance type, because it's kind of got some interesting stuff in it. And I just want to get a nice little starting point. I'll just grab this because a lot of these will end up looking the same. So we're importing Zod, ⁓ this library for defining schemas. ⁓ And we're going to want a ⁓ temperature.

unit schema ⁓ and let's ⁓ let's just focus on Celsius and Fahrenheit. Just C and F.

Sneha Mehra (00:02:05)  
And ⁓ this is kind of our source of truth for ⁓ what it what it means to be a temperature unit, but this isn't a TypeScript type.

Sneha Mehra (00:02:22)  
Not gonna risk spelling that. And you can see like ⁓ temperature unit schema refers to a value, but I'm use I'm trying to use it as a type. So how do we get the type out of this?

Sneha Mehra (00:02:40)  
Z infer.

Sneha Mehra (00:02:47)  
Just like this. So we've got a temperature unit, and look at this down here. This is what we want. So this lets you define a Zod schema and extract the TypeScript type that matches it. Why is this so important? This ensures that your compile time type checking and your runtime validation of objects that sort of flow through this schema are very well aligned. It's useful to have a source of truth where like both things originate from the same ⁓ structure.

And we'll export this ⁓ as a symbol so other parts of our code base can use it. Yes, question from online. Why are we adding validation logic inside the type.ts file? That's a good question. ⁓ maybe I should have called this like common or core. ⁓ if you look at other modules in this file, it's kind of like this is kind of a module that contains the the structures of the data that that we'll be working with. But this does not include just like.

Pair type information. There, there are some ⁓ you know, some type guards here. Like this is checking to see if something's a valid distance unit, and we may have similar things. So perhaps a poor name of the package on my part. But this is, you can think of this as like conceptually representing the shapes of the value objects, the shapes of the entities, the shapes of the request and response shapes. ⁓ like the HTTP request response shape pairs, but like ⁓

We have a schema here that's not just pure type information. And we may have type guards that that make sense to sort of put, you know, in in the the closest place we can put them, you know, with the objects that they relate to, the shapes they relate to.

Sneha Mehra (00:04:33)  
So we've exp we've exported the schema, because that's gonna be useful for whoever's using this to do runtime validations. We'll export the type, but this is just the C and the F, right? It's just the unit. So we're gonna need, ⁓ and just to prove that ⁓ this works. Like now, temperature unit, like the only error we're getting is this is an unused variable. So this is this is the type that we can kind of carry forward in our code. We're gonna do something very similar.

With ⁓ the monthly temperature range, but we're gonna create that as a different file here. So I'm just gonna format this really quickly and hit save. ⁓ And we're gonna create an entity now, or the the shape of the entity at least. And so I'm gonna create a new file in this in this entities folder. So you're in the types package, source entities, new file, ⁓ and monthly.

Sneha Mehra (00:05:33)  
Temperature range dot type dot ts. You can name it whatever you like, but this name it like I did and your code will stay in sync. ⁓ we'll need Zod just as before.

Sneha Mehra (00:05:51)  
And

Here we're defining our schema. And this ⁓ is going to be an object type. Like conceptually, what we want is something that kind of looks like this.

Sneha Mehra (00:06:16)  
Something like that. So like we're gonna look at this example and let's see how we can articulate that here. Question from ⁓ the chat. ⁓ Why does it make sense for a temperature range to have an idea? It seems more like an attribute of the actual entity. That's that's a good question. If we if we look back here, like going back to this discussion, we we could have said we have a temperature range that's a value object. And you can think of this as just

The concept of having like a min and a max, ⁓ and maybe we find multiple uses for temperature range in the future. ⁓ at this point, I think it's it could go either way, where we either say we're kind of like embedding this concept of a temperature range with this concept of a month, you know, like a ⁓ monthly piece of temperature data with the low and a high. We can always tease that out later if we need to. But it really this just comes down to.

sort of whether you want to make the choice to sort of tease that abstraction out right from the start, or whether you want to see it emerge as something like a temperature range emerge as something that's useful beyond this monthly temperature data. And and we'll see how that goes. Like we we still have a choice to evolve this in the future because ultimately we're going to have this API contract we create, like a request and response tape. So we can internally start to model this.

As a temperature range, and then have like a month that embeds a range. Like we could totally set up our ⁓ data model that way, but preserve the existing API contract if we had to. As we're defining our schema, we're gonna start with z.object. And that ⁓ all we're doing is like this part here. Like this ⁓ shape starts with be being an object. ⁓ And ⁓ we've got an ID, which is a string. We've got a month.

Which is a, we could do a number, but that also lets us do an integer. Remember, there's runtime validation that happens here. We can actually like, as we check this data, this is a constraint that goes beyond just the number type that exists in in JavaScript. And then ⁓ we've got the temperatures, right? Min and max.

Sneha Mehra (00:08:38)  
And what we're gonna do here is ⁓ use ⁓

Sneha Mehra (00:08:48)  
The temperature schema, which I think we f we forgot to do a piece here. But we want the temperature schema.

Now, we never finished our work on the previous file, so now's a good time to finish it. This is what we want. So all we have is temperature unit here. Sorry, sorry to switch back and forth a little bit. ⁓ we can finish and say the concept of a temperature is. ⁓

Sneha Mehra (00:09:18)  
It starts with an object, and it's got a value, which is a number, and it's got a unit, which is a temperature unit. And you can point right to the schema that we just defined for sort of the C or the F. And then similarly, we can just copy the same thing that we used before, but instead of temperature unit, it's just temperature.

So now we have like the unit, and then we're wrapping it in this object, so it's the value and the unit. And in both cases, we export both the schema and the type. So when we go back to our monthly temperature record, now this will resolve ⁓ and we'll use it for the min and the max with a comma.

and export it.

And then finally grab the t get the type.

Sneha Mehra (00:10:33)  
And there we go. And if we look at this, there's our monthly temperature range. Sorry, let me get rid of the terminal so everyone can see. ⁓

Sneha Mehra (00:10:51)  
That's the structure that we came up with. We have an ID, we have the number of the month, and then we have the min and the max. Yes. Should we limit month to be 1 through 12? Good point. Yeah. Does it make sense to have month be like ⁓ negative 16? Probably not. ⁓ And ⁓ Zod lets us do this. We can say min is ⁓ sh what ⁓ should we use here?

One through twelve or zero through eleven? ⁓ Zero zero. Let's do zero through eleven. Why do we want to use zero through eleven? In JavaScript, I think that's how it is. It's gonna line up really well with the date object, right? We can just grab like get the month number and it should work.

So there you go. So now now when we use this to validate data, ⁓ we can we can ensure that it's it is like an integer. Here's the min, here's the max, and you can even add things like you know ⁓ a description.

Sneha Mehra (00:11:56)  
No, is it describe or description? ⁓ yes, there's the description. So, and this this will like when there are error messages, or if you use this to generate a JSON schema of some sort, these will be like what is applied on the tooltips as you're sort of hovering over a file that's conforming to the schema. So ⁓ we have we have our types. We have to handle this piece here. ⁓ And the way we'll do that is we'll say.

Sneha Mehra (00:12:28)  
Monthly temps. ⁓

Monthly temperature range schema ⁓ and we don't want just one. We want an array. If we were to do this, we'd get an error because that's just treating it like it's type information. We want a z dot array. And that is the member type of the array.

And so now if we were to look at what this location type ends up being being like, you can see it's sort of ⁓

embeds all of this nice information here. So we've got like the location, the monthly temps, the min and the max. We're really articulating a reasonable structure here for how we think about these these temperatures.

Sneha Mehra (00:13:21)  
Mike, wouldn't we wanna put a like location ID on that monthly temperature range schema? Yeah. That's a good that's a good question. We could do it both ways. In fact, what t tell me how you think about Well you're gonna store it in a table, and so you it would each of those would have one location, and then ⁓ when you get the locations you resolve from the other table. Yep. But you're not gonna store an array of IDs on the location table for monthly temperature ranges, you're gonna resolve that from the database. ⁓

That's a great point. So we could either say the location schema embeds the monthly temperatures. Or we could say, well, wait a minute, those monthly temperatures, they're entities. ⁓ And we want sort of the belongs to relationship. We want the the monthly temperature to belong to ⁓ a location. And we can totally do that. In fact, I think that's a better way to do that. Like just looking at

Looking at the UI we're trying to provide here, like ultimately we want to get this working. We want to have a list of locations that pops up. And we don't need to pass all those temperature ranges up to the UI. So if this is part of like something that we seek to embed in an API contract of some sort.

That's kind of unnecessary to include. So the other way to do this is to go back to monthly temperature range schema and say,

Sneha Mehra (00:14:51)  
Something like that. And now we can kind of like behind the scenes find those. Like this never really ends up being exposed through to our API. So this is more of sort of a back-end concept. This would let us query by location and sort them by month. And then we have the points on our line. That's that's great.

Sneha Mehra (00:15:16)  
So there you go. So we've kind of like created a little bit of a soft edge on that on that relationship there where we know that this has to be a valid location ID.

Sneha Mehra (00:15:31)  
And then the location type just ends up being this. Right? Great for a drop down.

—-------------------------------

Sneha Mehra (00:00:00)  
Next, let's check out a branch where we have a lot of our Raisebed layout app already built out for us. So go ahead and check out the branch called DDD-client-validation. ⁓ And if you have your server running in the background, make sure you kill it and restart it. And then you should be able to start it back up, npm run dev. ⁓ And you should see a bunch of stuff starting up.

Sneha Mehra (00:00:32)  
Great. So if you go to ⁓ your local host, colon 5173 slash garden, you should see something that looks like this. So here we have a bunch of our, you know, our plants ⁓ categorized a little bit differently. And we can see some of the data is threaded through. There's a lot of placeholder data that we could add to if we wanted to. ⁓ And you can grab these tiles ⁓ and you can drag them into the raised beds ⁓ and place them.

And what you may notice is that some types of actions ⁓ are not allowed. For example, we see placement is out of bounds. Like if we try to put this tomato here, we can see we get ⁓ a validation failure. Not just an error message, but like this is a completion of a validation task of some sort. ⁓ Similarly, I think this bed ⁓ doesn't have enough sunlight for tomatoes. So I want to show you how we've set up some.

Some business validation rules. This is sort of the constraint part of how we think about these bounded contexts, right? We've got our entities, we've got our value objects, the relationships, and and the important constraints. And so in this branch, what you're gonna wanna take a look at is something called the workspace controller. ⁓ And this is ultimately where kind of like the brains of

Plant placement in the garden beds happen. And so if you want to look at like what what kinds of things we have in here, we've got like validate item move, validate item removal, validate item placement. ⁓ And ⁓ these are these are sort of like this is the interface that the rest of the app engages with as it relates to ⁓ this workspace controller class. Now, looking near the bottom of this class, you'll see something that's interesting.

And that is ⁓ plant validation rules. ⁓ And so here we've got something called check boundaries, ⁓ right? And if we look, we've got like the target zone. Remember in the UI, we're thinking about workspace zone ⁓ item. ⁓ And we've got the plant item. ⁓ And we're first getting the planting distance in feet ⁓ or the number. ⁓ And I would argue here we could change this even to size because our plant.

Sneha Mehra (00:03:00)  
You know, has a size on it. This is if ⁓ this is where we were doing that. You know, we have a normalized version of size that uses ⁓ math.seal to like round the planting distance up to the nearest whole foot, make sure we're at least at one cell by one cell. And then ⁓ we really just look at kind of the lower left corner of where the plant tile is, and that's that's like the

Well, that's the target X and Y. So in in the case of this ⁓ tomato plant, like you see the little coordinate system on the left and the bottom of the raised beds. So all we have to do is say, like, this bottom left cell of the four cells it would occupy, if if we take that coordinate and walk upwards and we exceed the height of the bed, or if we walked rightwards and we exceed the width of the bed, this is out of bounds.

Sneha Mehra (00:04:01)  
And that's this logic here. Target X, target Y plus size. So if we were to make this more interesting, and maybe these these tiles aren't squares, this is where we would go and change that. Now, what we do here ⁓ is ⁓ this is kind of the concept of returning an error rather than throwing an error. So a lot of programming languages, like Golang is a good example that follows this. ⁓ returning errors as values. And there there's

There's something very important about articulating things this way, in in this case. There's a difference ⁓ between ⁓ I threw while attempting to perform a validation versus I successfully performed a validation and it failed. So this this in and of itself is an important part of how you might model a validation task. There's like, I completed and this is allowed, or I completed, but this is disallowed.

And then there's like I exploded, or you attempted to ⁓ you attempted to place me in a garden bed that as far as I know doesn't exist. So a a good mental model I use here, because we we already have to deal with this, ⁓ HTTP requests, right? There's like a difference between an HDDP request that's like, am I allowed to place this here? Like, would you ever say you're gonna say like a 400 bad request?

If if it's sort of like not enough sunlight for the plant, you want to say, yes, we successfully completed the validation operation, respond 200\. And the answer is no, you can't place the the plant there. Does that make sense? And what this lets us do is we can also like ⁓ add more information to the validation failure, which then bubbles up into the UI and eventually it's presented in that ⁓ you know, in that little red, red

thing that you that you saw. So ⁓ we've got a couple rules here. We've got check boundaries, ⁓ where you can't place an item outside of the zone's boundaries. We've got no overlaps, and this gets ⁓ more interesting because we kind of have to, as we go through, like we get the the target zone, we look through all of the placements of the zone ⁓ and ⁓

Sneha Mehra (00:06:26)  
In case we're moving an item within a raised bed, we want to ignore that, right? Like if we're if we're over here and we're saying, I just want to move this led the spinach over by two, like we're gonna treat the original location of the move as vacant, because it's also a valid operation to just drop it right where it was. ⁓ And that shouldn't be considered an overlap.

Sneha Mehra (00:06:52)  
And then we've got ⁓ you know, the the logic where we're going through and we're ⁓ we're checking to see if there exists, ⁓ like, you know, if it's not, if the placement of all of the placements in the zone is not ⁓ ourselves, right? Like if if we're dealing with a tile in the bed that is not the thing being moved and and dropped, then we're going to examine its size and see if the width.

and height of these two things ⁓ overlap with each other. And what that might look like ⁓ is ⁓ is this. ⁓ Well sorry, I already can't drag that over here, but I shouldn't be able to do this. Right? Like this ⁓ overlaps with big boy tomato.

Okay, so we we have a nice little rule framework here. We've got a context object, ⁓ and ⁓ context has on it a lot of the information that we would need. Like where is this coming from? What's the operation type? Is this an addition, a removal? Is it ⁓ moving within a zone or moving across zones? And then we've got

access ultimately to the whole workspace in case we needed to go and grab sort of the root node ⁓ of this this whole diagram that we have in front of us.

—------------------------

Sneha Mehra (00:00:00)  
In this next part of the course, we're going to focus on solving the problem of a disorganized collection of seeds and get back into the mode of domain modeling and interviewing a domain expert or collaborating with a domain expert or an LLM to figure out like what is the most productive direction we could go ⁓ for solving this problem for our user. Before we get into this though, let's talk a little bit about ⁓

Some tips for having conversations with domain experts, and then we're going to have a conversation. So, first, ⁓ it's very important to develop a shared vocabulary with ⁓ all of the people you're collaborating on discussing a problem space with. ⁓ what this helps, what this makes sure that is avoided, ⁓ is ⁓ like things that end up being lost in translation and ⁓

you want to make sure that you have an opportunity or you give less technical people an opportunity to correct misconceptions you have. So if you're if you're talking to your gardener about ⁓ you know, your plans to like help organize their seed catalog, and you say, Well, like what we can do here is we can ⁓ take these objects and we can put them into an elastic search collection and we'll have like different dimensions that we can use to you know, search them, you know.

Put some denormalized data into Elasticsearch, like you've you've already lost them. ⁓ And ⁓ it's it's actually it's not that difficult to sort of take the names of classes and the names of variables and tie them to this common language that you share with your subject matter experts. When you're working with agentic coding tools, this also helps make sure that you're encoding more of the business problem right into the source code of your app.

And so you can have conversations that are much more in line with like, tell me about this seed packet. Like, do we have an estimated weight for the seed packet? And it's going to be able to go and find some terminology that joins the terminol like ⁓ the words used in the problem space with the words that are used in the solution to the problem. So ⁓ tips ⁓ for for kind of having discussions that are conducive.

Sneha Mehra (00:02:25)  
It's developing a shared language. ⁓ You want to understand terminology that's being used. ⁓ Ask clarifying questions. Make sure you really pin these things down and take the time as these as these words are mentioned for the first time or the second time, like coming up with a nice glossary of like what do we mean when we say this is very well worth it. ⁓ Use visual tools ⁓ like this little ⁓ diagram tool that I've been using, which is called whimsical.

Use a whiteboard, you could use a notebook, whatever you want to use. Slides are great for diagrams. ⁓ this helps make visual concepts ⁓ more tangible. And try to avoid technical logic and ⁓ sorry, technical terminology, such as referring to like database indices or instances of classes. ⁓ Like you you want to see if you can articulate kind of the shape of how you're thinking about the problem ⁓ in in solution agnostic terms.

Tip number two, ⁓ stay curious and open. So there there's a failure mode for these conversations, ⁓ which you know well-intentioned people that want to collaborate can fall into. So if if you're if you're talking to a gardener ⁓ and you listen to them ⁓ say a few things, ⁓ and then immediately you propose, you're like, here's how I think we could solve it, and then you sort of blurt something out. What's gonna happen is ⁓ the gardener's gonna listen to this ⁓ and they will pattern match against it and they'll assume.

Like you're a software engineer, you're probably pretty smart. The words you're saying mostly make sense to me, but ⁓ sometimes you get into this trap where they'll kind of like accept that description of a solution at a point in the conversation where you haven't reached alignment, where you haven't ⁓ gotten to the point where you can both say, yes, the way we're thinking about this is the same, and we've actually explored all of the details, right? Like in this case.

Alright, I can give you a list of your seed packets. ⁓ And you can use like Command F and you can find basil. Great. I I guess I can find all of the basils. I can sort of advance. If I have like you know, Thai basil and ⁓ sweet Genovese basil and all of this, like I guess I could control F and Alright, that's fine. But like what if what if there's incremental data loading and everything's not on the same page or something like that? Like you've

Sneha Mehra (00:04:51)  
You've offered a solution where sometimes people accept it without knowing that there will be inherent limitations to what they're what they're sort of validating to. ⁓ So ⁓ tips for avoiding falling in this trap. focus on the business problems. But what we mean here is like focus on the problem space. The business here is doing something useful for a gardener. ⁓ ask open-ended questions, ⁓ right? Like ⁓ you you want to understand

The why behind the ask. Like what is the core pain that is experienced, which the solution is aiming to develop? It's like aiming to mitigate that pain or take that pain away entire entirely. ⁓ Document and respect nuances, right? Sometimes ⁓ edge cases and quarter cases turn out to be really important. And while like there is a point when you're implementing like a minimum viable solution, you can start to say,

Alright, like we know we will eventually need to do this, but for now, let's not worry about it. Like, an example of this would be, ⁓ well, we've got this like carnation plant in our database. ⁓ And ⁓ we want to represent that like you can affect the color of a carnation based on the factors of the soil. So if you say I want a blue carnation or red carnation, that's not necessarily about planting the right seed. That's about

Additives in the soil. And so, like, all right, but like, let's worry about that when we come to it. Like, for the most part, we at least need to capture this dimension of like there are different ⁓ species of plants. Let's worry about that first. Soil additives can come later, it'll layer on. But like, it's important to not lose track of it. ⁓ request concrete examples for complex scenarios. And what I mean by this is ⁓ have your domain expert walk through what they do.

In the absence of your solution, step by step. So, if, for example, this is ⁓ an accountant trying to perform some calculation, like literally get on a whiteboard or get in a doc with them and say, take me through, like step by step. You start with this number, then what happens? Then what happens? Okay, you consult this spreadsheet. Who makes that spreadsheet? And just walk through that process. That can become what is called a user journey.

Sneha Mehra (00:07:20)  
Right? That is ⁓ that is the user journey of today, the process your user goes through to get from the beginning of ⁓ having the problem to like scratching their own itch, if there is a way to scratch that itch today. And that lets you avoid missing things. That lets you make sure you at least understand the world of today. ⁓ And ⁓ especially complex scenarios. Like as you hear about these nuances, dig into them, understand them.

And then obviously like show active interest. Sometimes you'll be dealing with like a dry subject, like dry topic. You know, some people get really excited about ⁓ well, I'll say, like, I get really excited about ⁓ financial concepts, which is great. I work at Stripe, it fits really well. I'm genuinely excited when I go to work to sort of wrap my heads around different business models and things like that. Other people would consider that to be incredibly dry. ⁓ but you do want to.

Make sure that you're you're trying to tease out these ⁓ these details in a way where you're sort of feeding energy into the conversation. And remember, like, presumably you all are excited about building software, and ultimately that's part of this. And so try to tap into that enthusiasm. ⁓ iterate and validate regularly. So a key part of domain-driven design that I think is definitely worth keeping is the idea of refining knowledge. ⁓ I know we all

Have probably in our careers ⁓ dealt with like, you know, somebody puts a big document together and they say, This is the plan. And then as you start working on the plan, it changes course. ⁓ And that document is like, no one's updating it. It's just sort of rot. ⁓ It's important to sort of take a look at the information you have as you make progress ⁓ on implementing things and thinking through things. And some, like sometimes you have to write some code in order to figure out like,

Well, I have a set of new questions now. Refine that knowledge, go back and and keep a nice, concise source of truth that you're willing to keep up to date as you as you iterate. This could be just like a one-page document with your glossary of terms, ⁓ or ⁓ you know, a simple diagram like this, where like we changed temperature range, we combined that into a single monthly temperature range entity instead of having that intermediate value object here.

Sneha Mehra (00:09:44)  
Have something you're willing to keep up to date, and that is usually not ⁓ a ⁓ you know 30-page design doc. So ⁓ a big, a big tip here is ⁓ summarize ⁓ and reflect concepts back to experts. This is easier than ever with LLMs. If you have like taken a ton of notes, you can feed it into something like Chat GPT and say, give me the most important bullet points here. ⁓

And focus is important. So at the ends of these conversations, ⁓ there's an important step to take. And that is like you'll have a bunch of notes. ⁓ And I want you to to sort of like if you have a half-hour discussion with someone, what you want to do is save the last three or four minutes ⁓ to with your collaborators ask the question: what if anything we've discussed is worth keeping and carrying forward? It's likely not all of it. You know, you've gone into

some avenues where it didn't yield anything or you have a very elaborate carnation related topic here that could really be boiled down to for now we're focusing on plant species, ⁓ although other variations of plants will exist. Maybe that's all you need to do. And and that's that's the important ⁓ gem to mind in the auror of the full conversation that you had.

—-------------------------

Sneha Mehra (00:00:00)  
Let's finally get to the ⁓ calculating the date. So we've already got ⁓ a you know something here that's using this location repository, which we there's one declared in an outer scope here. ⁓ or sorry, in the in the class. But like we're finding one record where the ID is ⁓

you know, is of this format. So we're gonna make sure that that looks right. We're gonna just like console log, like what is this request thing coming in? We'll take a look at it. We'll make sure the ID looks right. We'll log the location. And then we're just gonna quickly we're gonna fetch the list of temperature ranges and to keep things simple, let's just like find the first month that appears to satisfy the range. And we're just gonna return like the first of the month as the date. You could interpolate, but that's that's not the point of what we're trying to focus on here.

So ⁓ first let's ⁓ let's say ⁓ s sorry, let me make sure we're focusing on the right thing here.

Sneha Mehra (00:01:09)  
I actually don't think that this is gonna be needed. We're gonna let's let's focus on the calculate date function. This it's it's the same same thing here, but this is where we really need to focus. So I wanna log.

The ⁓ location ID ⁓ and the temperature. And I can delete this underscore because it's now it's a used variable. And you can see we have a little mini schema here because we hadn't ⁓ created that temperature type for the starting point code. ⁓ but like let's let's see what actually comes through when we hit that calculate date button. So we're gonna go back to our UI, sorry, hit save, go back to our UI, ⁓ and ⁓

Let's select a location, Minneapolis, ⁓ and I want to know when when we're ⁓ over 122 degrees in Minneapolis. No, let's go with 50 degrees Fahrenheit. Estimate date. And I I expect this to fail because we haven't built this out yet. But let's take a look at what the request looks like going out.

Sneha Mehra (00:02:18)  
So we've got a location ID. It already has this LOC prefix, so we're gonna keep that in mind. And then we've got a value. This is just our temperature type coming through again. ⁓

Going going back to our code here, what we can do is say, well, we don't need this anymore. Actually, sorry, we can cast it to this type.

Because all we're doing here is we're telling TypeScript that this prefix exists. This is a template literal type where it's like any string that begins with L-O-C underscore. ⁓ And ⁓ let's see. ⁓

Sneha Mehra (00:02:59)  
Let's see if we get the location out.

So we're gonna hit

Estimate date again. ⁓ I keep going the wrong way. And what are we seeing? Location with ID, this not found. All right. So the first thing we want to do, remember I said the database is recreated each time we hit save. We want to load that database, ⁓ load the app again.

Sneha Mehra (00:03:29)  
I'm gonna put these next to each other so I don't have to swipe across too many things. We're gonna load this app again because like these location IDs keep changing. So, Minneapolis, fifty degrees ⁓ Fahrenheit.

Sneha Mehra (00:03:46)  
Estimate date. All right, what do we get back? We get back December 31st. So this is just like the hard coded date that we get back. So what this tells me is we're gonna go back to our code here, and we should see we've loaded a location. Here's the instance of the class. ⁓ interesting, monthly temps undefined.

So there are two ways we could handle this. One would be, well, we can go and like grab the location ID, we can query that table, get get all the monthly temps that match the location, but there's a shortcut.

Sneha Mehra (00:04:37)  
If just do this and I'm gonna have to reload. We don't need this anymore.

Sneha Mehra (00:04:48)  
Estimate date and let's go back to our code. We should see that that is now populated. And it is.

⁓ sorry, that's ⁓ other logging that's happening there. Let me get rid of some of the other logging so it's super clear. ⁓ it's in our floating here.

Sneha Mehra (00:05:18)  
We don't need that anymore.

I think this is this is it. Look at that. An array of monthly temperature ranges with commas in between them. And if we scroll up high enough, there we go. We have a location. And there are our monthly temps. So we've basically just traversed this relationship and it's turned into two queries. One to get the location, ⁓ one as a batch get to fetch all of the monthly temps ⁓ that we're interested in. Now all we have to do is iterate over those temps.

Sneha Mehra (00:05:54)  
⁓

Sneha Mehra (00:06:06)  
And what's the temp that the user passed in? Where are we getting that? Temperature. Perfect. ⁓ If ⁓

If ⁓ Mt dot min value ⁓ is greater than temperature dot value return new date.

Sneha Mehra (00:06:42)  
We could just do that. So this should give us something maybe something other than December thirty-first. Let's check it out. So reloading. Minnesota, Minneapolis, Fahrenheit, 50\.

Interesting, we still get twelve thirty-one. What we're missing here is our unit conversion. And I think we already have ⁓ in temperature.

Sneha Mehra (00:07:14)  
Well, let's let's save time. You could do some unit conversion here, but for now I'm just gonna make the equivalent request in Celsius, which is ten degrees.

Sneha Mehra (00:07:31)  
And look at that, we've got 6-1. So we've modeled monthly temperature range, we've modeled this value object of a temperature, and we've wired it up so that it's all working in the temperature date calculator. And next we're gonna turn our attention towards seed packets ⁓ and really focus in on value objects ⁓ and have a discussion about ⁓ what is most meaningful ⁓ in order to deal with this user interface.

Where we can have the back of a seed packet and structure some data for ⁓ you know, what's what's useful for your gardener in order to keep this well organized.

—----------------------  
Sneha Mehra (00:00:00)  
Do we need like a value object that represents ⁓ or that can contain information about a given stage and a duration of time? Cool. Okay, tell me more. ⁓ Well, if ⁓ you know, you could represent ⁓ the indoor growing time and the hardening time with the same value object, those two entities can be, or maybe could be discrete things per plant, unless hardening is like always seven days for every single plant.

Can use the same sort of value object to represent those things. Okay. ⁓ And I'm I'm not a gardener anymore, so use ⁓ using the word value object here is fine. But like what what would we call this thing? Like how do we how do we give this a name that makes sense to our gardener? The plant's name. ⁓ the ⁓ this is the class. This is the class. Like I think what you're trying to capture here is there's ⁓ some something here, right? Like I've drawn little boxes here.

That's a signal, right? If your gardener gets up and they say, like, this is a thing, and they circle a thing and they say, Here's what I'm describing here, like that's teasing that in a planting stage or something like this. Yeah. Like ⁓ plant plants plant stage. Yeah, we could call it that until we come up with a better name. But those are words like

Do we need to know more? Are there more stages than really just the like would you call it the germination stage where it's inside? Yeah. ⁓ there could be. Let's start with this though. They I mean, realistically, there are. There's like you plant the seed and then it's in the hydroponic thing, and then you put it in a red party cup because it's too big for the hydroponics. Like it goes deeper, but like I think we we have enough here. So we've got a plant stage, and then we've got like a time.

Time period on it probably, right? Yeah.

Sneha Mehra (00:02:00)  
All right, how should we represent that?

Sneha Mehra (00:02:06)  
Like what is the time period? Days are probably the easiest. Cool.

Sneha Mehra (00:02:17)  
And ⁓ describe to me in words like what's an example of a time period? Is it a range range of days? ⁓ It it's could be a range of days. Like specific dates? Like ⁓ March twenty seventh through April ninth? ⁓

So a start date and a length? It's I think it's a start date and a length, right? Like this is this is a this is a a duration. It's a time duration.

Sneha Mehra (00:02:51)  
It might already have duration in there somewhere, but like let's let's say. And it's a it's a ⁓ value and a unit.

Sneha Mehra (00:03:05)  
Days or whatever.

Okay, great. And so this is a time duration. So now we can model a plant stage. We could say, all right, here's the germination time. Here's when you're planting things. All right, now what what do we need on this?

Well like great, we can model durations. Maybe give it a name.

Sneha Mehra (00:03:37)  
Okay, so great. You could say here's ⁓ some kind of tomato, and there's like a germination duration that's this many days, and then a hardening duration that's this many days, and then we plant it in the ground. Presumably then it's growing and then it fruits and then I can harvest, so it's that many days. And then it's continuing to fruit until it gets cold and gets knocked out. This is great. Let's let's pop back up to seed packet though. ⁓

Hmm, interesting. On seed packet? ⁓ On the plant stage. Plant stage has a start date and the duration from that start date. I'm so glad you brought this up. So this is where, like, we're talking about the seed packet here. I'm not the there's also gonna be a concept of a plant. Like you put a seed in and the plant's growing. And that's that that has ⁓ there's some state there, right? Like there are 200 seeds in the tomato seed packet.

And those can become each an individual plant. But we're we're trying to model the packet here. Does that make does the difference make sense? Mr. Gardener, does every plant have a indoor planting, a hardening, and a plant in the ground phase? ⁓ Mm. Well, the the start date would depend on the location. Start date is very all of those things, plant stage and time duration, are very much derived state from the seed packet. They're not their own entities. You're gonna pick the

date, the germination time, all the information in the seed packet and create those things as value objects based upon that data. We're not going to store that. If everything has ⁓ those three stages, or if we and if we have the data to determine if what stages they have.

This is great. This is great. So, so ⁓ a very productive conversation here. So seed packet is kind of like it's almost like the template for a plant, if you want to think about it that way. Like when we say a time we have a time duration for germination, that's like that is ⁓ sprouts in seven to fourteen days. That's what we're trying to capture here. Now, once it's planted, there's a time where it's planted.

Sneha Mehra (00:05:50)  
And there might be a specific, like a date time, a date and a time where like we've entered the beginning of that range, we've left that range. And that that's that's gonna be a different ⁓ a different concept that we're gonna model. But like these seeds ⁓ sprout in seven to fourteen days the first year, midway through the year, the next year, it's still seven to fourteen days, right? It's it's sort of ⁓ does that does that difference make sense? All right.

Let's jump back into this though. Like what what else do we need to know? So I like I like this. Like we have plant stage, and there's kind of a many relationship there. Actually, it's at least sorry, I keep picking the wrong one. ⁓ it's it's one or many plant stages, and then each of these has a time duration. What else goes on the seed packet?

Quantity. Let's let's start there.

Sneha Mehra (00:06:48)  
That's a number. Variety.

Fighting.

Sneha Mehra (00:07:00)  
Great. Plant family.

⁓ I like that word better.

Sneha Mehra (00:07:08)  
Plant family. You mean a name then, right? Name.

Sneha Mehra (00:07:17)  
Great. So so like let's let's use our tip of like going through a concrete example here. ⁓

Sneha Mehra (00:07:27)  
So that's a name. 200 seeds in a packet. Tomato. And then some date. Some date. That's a problem for Tuesday. That's one I would probably choose ⁓ not to model. Because expire date is what matters to me. Like maybe seeds have different shelf life. And like really what I care about is ⁓ like there was this story ⁓ of

we didn't document it. Like

Sneha Mehra (00:08:03)  
I wanna plant seeds before they expire. But like purchase date, I I would argue there we're getting a little bit close to that trap of just like modeling more than is necessary on the seed packet in order to solve ⁓ solve this problem. Now if I were accounting for my seed purchases and I wanted to like deduct them from taxes or something, absolutely that matters. But we didn't that that wasn't wasn't part of the part of the pain there. You bought a whole bunch of seeds last year at a discount that were almost expired.

It doesn't really matter what There you go. Yeah. Yep. Yeah, or I could have bought them from a friend. Yeah, like exactly. They're about to expire. And so the purchase date is unknown to me. It, but I do know. Like, I do look on this real ⁓ seed packet, and they have a sell ⁓ by hair. and and like here you can see evidence of that. Remember, I said it's like a year-based expiration kind of thing?

Packed for 2024\. There's a clue right there. It's packed. These are made for a season, ⁓ and it's less like look, 1231st, 24\. Like, really? Does it expire on New Year's at 1159? Like, it's really about sort of the seasonal rollover there. So this is why looking at concrete examples often yields some some great insight here. Like it would have been totally easy for us to ⁓ overdo the expiration date and been like,

You know, let's let's filter by expiration date and let's like let you put in a day, month, year. But really, like if we were building this feature, you'd just have a drop down with years and it would be fine. Like that's that's probably as complicated as it needs to get.

Okay, let's shift into taking some of what we've learned ⁓ and ⁓ implementing it.

—----------------  
Sneha Mehra (00:00:00)  
So we've parsed some basic information about location, but we need to s we need to do a very similar thing with these monthly temperature records. And I'm gonna do I'm gonna, while we're doing that, convert to that four await, you know, it's not really a four-await loop, but it's a four loop with an await inside each turn of the loop. Because now we're not just doing seven things, we're doing seven times twelve. And that is surely enough to ⁓ warrant.

you know, not ⁓ racing all of those rights to a database concurrently. So ⁓ instead of this map here, we're gonna do just a standard for loop. And this is, I believe, yep, already an async function, so we're great here. So we can say ⁓ locations is a ⁓ it's an array.

Of locations ⁓ initialized to an empty array. Const declaration is fine. We're not reassigning this array to a different array, although we will be filling it.

We can just at the bottom return locations, because we're going to handle the asynchrony ⁓ differently within the function.

And just to give us a little bit of focus here.

Sneha Mehra (00:01:20)  
And ⁓ here's how we'll change things.

Sneha Mehra (00:01:38)  
An old school for loop. ⁓ And we're gonna do the exact same thing in here that we were doing before.

Sneha Mehra (00:01:48)  
There's my loc repo. ⁓ let's see. We need a loc data.

Sneha Mehra (00:01:59)  
So there's my loop data and ⁓ there's no return in the t each turn of the for loop, but

Sneha Mehra (00:02:11)  
If we do this, like the way the way we've restructured this. ⁓ Sorry, let's we could totally do a for of loop, but I like I like the logging opportunities of having that index in there. ⁓ so instead of quickly mapping over everything that we find in the file and then waiting for all those promises in parallel to resolve, we're waiting for each save operation to resolve. Only then when that's resolved, do we push it into the array.

We should end up with a solution that works ⁓ equally well here. I mean, we're we're sort of serializing ⁓ all of those location persistences, but it's like microseconds, it's fine. What this lets us do ⁓ is ⁓ after we create the location.

Sneha Mehra (00:03:05)  
We can start to do other things, like create a monthly temperature.

Sneha Mehra (00:03:14)  
And to do that, we'll need the other repo.

Monthly temperature range. Again, being careful to choose the entity, not just the type. MT repo. Right, so just like the same same deal here. This is just the other another instance of that repo class that deals ⁓ in these monthly temperature ranges. ⁓ And we're gonna do a very similar thing like what we did here.

Mt repo.create and here are my params and then we'll fill those in and then ultimately we're gonna do this.

Sneha Mehra (00:03:59)  
Right? So this is the monthly temperature. ⁓ And we'll see if we even need to go beyond that, whether whether we're just sort of awaiting this, or whether we'll end up using that variable. At the very least, we could log it out for now and then delete that. ⁓ but it's our time to perform this association. So the first thing is location, loc. So that's that's great. Like we've got that that relationship wired up now. And what are the other things we need? We need a month, a min and a max.

So let's let's worry about the month. Now ⁓ we're gonna have to do this inside a loop, right? Because each location has multiple pieces of temperature data, one for each month.

Sneha Mehra (00:04:46)  
⁓ we'll call it month data.

Sneha Mehra (00:04:54)  
We'll do a 408 loop here to make ESLn happy. ⁓ Loc data, monthly temperatures. ⁓ Here we go. Wrap the whole thing in a loop. Boom. Okay, now what should the month be?

We could just do this, ⁓ but we're gonna do that. Keep it consistent with the date object.

And then min. So what do we want for min? We've got a value.

And we've got a unit. And unit can be C C ⁓ R F mm. It's not type checking against it yet, but it will.

Sneha Mehra (00:05:42)  
There you

Sneha Mehra (00:05:47)  
So ⁓ yeah, we could say it's a deep partial of this. That's what it's type checking against. So you really want to see that that ⁓ type checking come through. You can use this satisfies ⁓ keyword. ⁓ And even without this now, you should see.

Sneha Mehra (00:06:13)  
This is just ESLint being touchy. ⁓ we'll just keep going here. So this is gonna be month data dot temperature range dot min. ⁓ And we can see the tuples coming through here. We've got number ⁓ and ⁓ a ⁓ a C or an F. So we can pull this apart up here.

Sneha Mehra (00:06:43)  
Something like this. ⁓

Sneha Mehra (00:06:51)  
Min value, min unit, max value, max unit.

Sneha Mehra (00:06:59)  
And wired up.

Sneha Mehra (00:07:05)  
And basically the same thing for Max.

Sneha Mehra (00:07:20)  
let's see. What am I doing wrong? interesting. Did we forget a piece of a bit of our type information? We did. There's the men. We need a max two.

While you're on this, ⁓ is it possible to ⁓ pull in the type for the temperature range unit and then make a Zod object out of that type? You absolutely could help it in the future? In fact, that's why we do this.

⁓ we could say ⁓ we could even pull in the whole temperature schema, given our data file sort of ⁓ no, we couldn't because of the tuple representation.

Temperature unit schema we could use for that. So here's how that would work. It's just for this union type piece.

Sneha Mehra (00:08:13)  
There you go. Using little chunks of Zod schema, absolutely.

There we're aligning with the same C and F because that's a common piece between our API contract and the way data is stored in this file. Great. Now we've got our min and our max. ⁓ sure. ⁓ And we ⁓ we save each location's ⁓ data. So I'm gonna hit save. Again, like when you're in dev mode, ⁓ your server restarts every time you hit save. The database is deleted.

And it's recreated again. And so we should be able to see if this is working just by looking at our SQLite database. ⁓ And let's see if we've got table entries. We do. And we've got our months. There's month zero. Here's a location ID. And you can see like this key up here.

indicates you can't see this tooltip very easily, but it says it's a foreign key. So that means if we needed to do a join on this later, like this is an index column, it makes traversing that relationship between monthly temperature data and location really, really fast.

—------------------

Sneha Mehra (00:00:00)  
So the first thing we're gonna do, we're gonna check out the project, and then we're gonna sort of take a take a quick look at how data flow works in this app. So I'm gonna check out the repo here. It's mic-north slash p shoot. So this is the git repo. You're gonna check this out. Make sure you're using at least node.js23 because of the way module imports work. So in your in your authoring environment, you can use node-v and you should get 23\.

Or greater. ⁓ And you're going to want to check out the DDD start branch. So that should look like.

Git clone. Git at github.com. ⁓

Sneha Mehra (00:00:48)  
Something like this.

Sneha Mehra (00:00:52)  
I'm just adding this so it creates a new folder.

Sneha Mehra (00:01:02)  
And then you should be able to run npm install or npm i.

Sneha Mehra (00:01:15)  
And then you should be able to run npm run dev.

Sneha Mehra (00:01:23)  
⁓ sorry. You want to build it before you run npm rundev. Part of this is because the types package needs to have things in its ⁓ its build output folder in order for this dev server to run. So we're gonna run npm rundev. Great. And you should see some feedback here. Server running on port three thousand. ⁓

And if you scroll up, you should be able to see. Wow, it's a lot you can see we have a lot of data in this app. ⁓

There go. But right before it did all of that. Localhost 5173\. If this doesn't work on your machine, if like you see that something that looks like that successful build output and it doesn't work, sometimes like if you have something already on this port, it'll try 5174\. But you should be able to click this and you'll have to do a little bit of a refresh. And then you should see an experience kind of like what we've got going on here. If you see that your project

Isn't rendering ⁓ images. What you're gonna want to do ⁓ is ⁓ well the the way you'll know that you're in this situation is if you look at ⁓ any PNG or JPEG in this project, and here here's an example path of one: packages, client, public, plant icons, apple something.png. Like if you select one of those files and you don't see an image, ⁓ what likely has happened is ⁓

⁓ you you don't have git lfs installed and set up in this repo. So you'll want to ⁓ install git lfs, which I think in in on a Mac it's just brew install git dash lfs. And then you're gonna want to ⁓ set it up in the repo. What did you do to set this up in the repo? Git LFS install. Git LFS install in the repo. Sorry, I did this a while ago. And then git lfs pull. And that should replace all these little

Sneha Mehra (00:03:29)  
What what would have been text pointers in here, just like little three-line text files that sort of refer to an image, you're gonna actually be downloading downloading the image and ⁓ you should then be able to select these things and see that something's rendered for you. Before we get into ⁓ implementing this sort of temperature calculator, let's think about the concepts that we'll we will have to introduce. All we have so far in this project, and

On the course website, there's a little outline of how does data flow work in this app. But like there's a place where the request initiates from the client. There's a UI component that represents like what's shown on the screen. And then the request is initiated from this file. ⁓ and there's a calculate date method. So like let's go look at this.

Just gonna do command P, paste that link, and there's a calculate date method. So here it is. And you can see we have a calculate date request, ⁓ and it looks like we're just like fetching ⁓ against our API with a particular path, ⁓ and we're sending the request as the body. And then if the response comes back and it looks okay, we're apparently you know parsing it and returning a date. And like from that point.

Going back into the UI component and the appropriate thing is rendered on the screen. So this is sort of a an isolated piece that represents the client side of this data flow. Now, where is it happening on the server? Well, first, like let's worry about the contract here. We've got a locations calculate date ⁓ TypeScript file. Locations calculate date.

And here you can see we've got like a request and a response. And here are examples of those Zod schemas. So we have a location ID, we've got a temperature with ⁓ what looks like a number and then a union type. And so if we were to look at this, ⁓ like this is what that sort of amounts to. So we can just see it all all on the same screen.

Sneha Mehra (00:05:43)  
Location ID, and we've got a value, temperature value, and in temperature unit, and you can see here, like this is the way in Zod you would articulate that.

And then what's the response we're looking for? It's a stringified date. And that makes sense. Like if we go back to our client side code, you can see whatever this returns, we're just kind of like shoving it into the date constructor. And so if this is like an ISO 8601 date string or like whatever the date constructor is ready to ⁓ receive, this should just work. Otherwise, ⁓ we throw an error. So ⁓ where does this land on the server? So we

This is the introduction of what we call like our first domain service. And this is a similarly kind of like isolated and purposeful piece of code, kind of like what we were just looking at on the client side, where we're dealing with ⁓ locations and dates. ⁓ So it's like ⁓ in the server package, it's source services location. And here you can see we've got like a location service.

And we've got a couple methods here. Let me fold them up so we can see it at a high level. We've got a method for getting a location, getting all locations, and calculating a date, and it returns a promise that resolves to a date. So all of these are async. And an important thing I want you all to notice here is we're this ⁓ service here is dealing really in the concept of a location. Like the HTTP request and response stuff is handled elsewhere. This is part of what it means to have like

A domain service. ⁓ it it is it is separated from our ⁓ route, you know, for like handling the requests ⁓ you know in this code base. That's what deals with request and response shapes. But part of part of like what's useful to take away from DDD here is like this service should just deal in the concept of like whatever models, whatever entities are important for the tasks that we're performing here. And that way.

Sneha Mehra (00:07:50)  
You can test this very effectively without worrying about like ⁓ HTTP requests and response. You're you're just sort of thinking about the date objects and what what's what is stored in memory in terms of the dates. And then you can test your HTTP request and response ⁓ part of your app by kind of mocking this and saying, All right, like I've got a stub location service that's just gonna spit the same stuff out. ⁓ And ⁓

Now it's very easy to sort of in isolation make sure the request and response stuff ⁓ works. So that's kind of how the data flow works. Start from a UI component. In this case, we're using Svelte because I wanted to stay as close to vanilla TypeScript as we can. And you shouldn't, we should just be able to set values there and things should just work, just like it's vanilla TypeScript. We go into sort of the the data layer piece of our client, which is this repository, to make the request.

And then ultimately downstream, we end up in this ⁓ service ⁓ and we call calculate date, or we call get all locations.

—-------------------

Sneha Mehra (00:00:00)  
So the first thing we're gonna do, we're gonna check out the project, and then we're gonna sort of take a take a quick look at how data flow works in this app. So I'm gonna check out the repo here. It's mic-north slash p shoot. So this is the git repo. You're gonna check this out. Make sure you're using at least node.js23 because of the way module imports work. So in your in your authoring environment, you can use node-v and you should get 23\.

Or greater. ⁓ And you're going to want to check out the DDD start branch. So that should look like.

Git clone. Git at github.com. ⁓

Sneha Mehra (00:00:48)  
Something like this.

Sneha Mehra (00:00:52)  
I'm just adding this so it creates a new folder.

Sneha Mehra (00:01:02)  
And then you should be able to run npm install or npm i.

Sneha Mehra (00:01:15)  
And then you should be able to run npm run dev.

Sneha Mehra (00:01:23)  
⁓ sorry. You want to build it before you run npm rundev. Part of this is because the types package needs to have things in its ⁓ its build output folder in order for this dev server to run. So we're gonna run npm rundev. Great. And you should see some feedback here. Server running on port three thousand. ⁓

And if you scroll up, you should be able to see. Wow, it's a lot you can see we have a lot of data in this app. ⁓

There go. But right before it did all of that. Localhost 5173\. If this doesn't work on your machine, if like you see that something that looks like that successful build output and it doesn't work, sometimes like if you have something already on this port, it'll try 5174\. But you should be able to click this and you'll have to do a little bit of a refresh. And then you should see an experience kind of like what we've got going on here. If you see that your project

Isn't rendering ⁓ images. What you're gonna want to do ⁓ is ⁓ well the the way you'll know that you're in this situation is if you look at ⁓ any PNG or JPEG in this project, and here here's an example path of one: packages, client, public, plant icons, apple something.png. Like if you select one of those files and you don't see an image, ⁓ what likely has happened is ⁓

⁓ you you don't have git lfs installed and set up in this repo. So you'll want to ⁓ install git lfs, which I think in in on a Mac it's just brew install git dash lfs. And then you're gonna want to ⁓ set it up in the repo. What did you do to set this up in the repo? Git LFS install. Git LFS install in the repo. Sorry, I did this a while ago. And then git lfs pull. And that should replace all these little

Sneha Mehra (00:03:29)  
What what would have been text pointers in here, just like little three-line text files that sort of refer to an image, you're gonna actually be downloading downloading the image and ⁓ you should then be able to select these things and see that something's rendered for you. Before we get into ⁓ implementing this sort of temperature calculator, let's think about the concepts that we'll we will have to introduce. All we have so far in this project, and

On the course website, there's a little outline of how does data flow work in this app. But like there's a place where the request initiates from the client. There's a UI component that represents like what's shown on the screen. And then the request is initiated from this file. ⁓ and there's a calculate date method. So like let's go look at this.

Just gonna do command P, paste that link, and there's a calculate date method. So here it is. And you can see we have a calculate date request, ⁓ and it looks like we're just like fetching ⁓ against our API with a particular path, ⁓ and we're sending the request as the body. And then if the response comes back and it looks okay, we're apparently you know parsing it and returning a date. And like from that point.

Going back into the UI component and the appropriate thing is rendered on the screen. So this is sort of a an isolated piece that represents the client side of this data flow. Now, where is it happening on the server? Well, first, like let's worry about the contract here. We've got a locations calculate date ⁓ TypeScript file. Locations calculate date.

And here you can see we've got like a request and a response. And here are examples of those Zod schemas. So we have a location ID, we've got a temperature with ⁓ what looks like a number and then a union type. And so if we were to look at this, ⁓ like this is what that sort of amounts to. So we can just see it all all on the same screen.

Sneha Mehra (00:05:43)  
Location ID, and we've got a value, temperature value, and in temperature unit, and you can see here, like this is the way in Zod you would articulate that.

And then what's the response we're looking for? It's a stringified date. And that makes sense. Like if we go back to our client side code, you can see whatever this returns, we're just kind of like shoving it into the date constructor. And so if this is like an ISO 8601 date string or like whatever the date constructor is ready to ⁓ receive, this should just work. Otherwise, ⁓ we throw an error. So ⁓ where does this land on the server? So we

This is the introduction of what we call like our first domain service. And this is a similarly kind of like isolated and purposeful piece of code, kind of like what we were just looking at on the client side, where we're dealing with ⁓ locations and dates. ⁓ So it's like ⁓ in the server package, it's source services location. And here you can see we've got like a location service.

And we've got a couple methods here. Let me fold them up so we can see it at a high level. We've got a method for getting a location, getting all locations, and calculating a date, and it returns a promise that resolves to a date. So all of these are async. And an important thing I want you all to notice here is we're this ⁓ service here is dealing really in the concept of a location. Like the HTTP request and response stuff is handled elsewhere. This is part of what it means to have like

A domain service. ⁓ it it is it is separated from our ⁓ route, you know, for like handling the requests ⁓ you know in this code base. That's what deals with request and response shapes. But part of part of like what's useful to take away from DDD here is like this service should just deal in the concept of like whatever models, whatever entities are important for the tasks that we're performing here. And that way.

Sneha Mehra (00:07:50)  
You can test this very effectively without worrying about like ⁓ HTTP requests and response. You're you're just sort of thinking about the date objects and what what's what is stored in memory in terms of the dates. And then you can test your HTTP request and response ⁓ part of your app by kind of mocking this and saying, All right, like I've got a stub location service that's just gonna spit the same stuff out. ⁓ And ⁓

Now it's very easy to sort of in isolation make sure the request and response stuff ⁓ works. So that's kind of how the data flow works. Start from a UI component. In this case, we're using Svelte because I wanted to stay as close to vanilla TypeScript as we can. And you shouldn't, we should just be able to set values there and things should just work, just like it's vanilla TypeScript. We go into sort of the the data layer piece of our client, which is this repository, to make the request.

And then ultimately downstream, we end up in this ⁓ service ⁓ and we call calculate date, or we call get all locations.

—-------------------

Sneha Mehra (00:00:00)  
Next, let's turn our attention to the UI component that represents the back of the seed packet. And we're going to open up seed packetback dots felt. ⁓ And ⁓ you can see we've got ⁓ a bunch of stuff here. I've already added expires at here, and we could change this to expires ⁓ at to something like that. ⁓ What else is interesting on here? Planting distance. So we could add another little piece of UI here that's like.

Planting distance.

Is there already another one? Nope. So good. Distance.

Sneha Mehra (00:00:42)  
⁓ And ⁓ this is ⁓ always a number. So let's say it's always feed. And we should be able to go back. ⁓ we should be able to start our server. ⁓ And then this should start to appear ⁓ on the back of the seed packet. And by start our server, I mean not just the tapes package, the actual whole monorepo, run npm run dev. ⁓ And

Sneha Mehra (00:01:13)  
Sorry, it takes a little second to come up.

Sneha Mehra (00:01:19)  
Interesting.

Sneha Mehra (00:01:24)  
What's going on here?

Sneha Mehra (00:01:32)  
⁓ looks like we have a validation failure here. So ⁓ what we should do is look at those request response shapes and see if this data that we're threading through ⁓ is ⁓ in alignment with sort of the the API contract between these two ⁓ these two types. But this is Zod at runtime, keeping us honest. And it's saying metadata dot planting distance. Like it's it's an invalid type, like.

Expected number, received object.

Now, the reason for this ⁓ is if we look in our ⁓ seed packet router, we've got planting distance here, and look at this. It's a distance object with a unit and a value. And what's expected is a number here. ⁓ thankfully, yep, I've got a little helper function already built for us. Convert distance to feet. And if we hit save there, I suspect.

Sneha Mehra (00:02:36)  
Interesting. Planting distance and days to harvest undefined. Wait, let's let's look at these carefully. New new request there. So expected number received undefined for days to harvest. Let's see if we can look at that one.

Sneha Mehra (00:02:55)  
Days to harvest. We didn't we didn't wire this one through. Let's just set this to thirty or something.

Sneha Mehra (00:03:05)  
That took care of that one. And now expected number received object for metadata dot planting distance.

This point

Sneha Mehra (00:03:23)  
There's the response. ⁓ Wait, what's going on here?

Sneha Mehra (00:03:35)  
Types of metadata planting distance are incompatible. We've got all right, so it wants ⁓ it wants the full value there.

Let's let's keep pulling at this. So that makes the type happy. Maybe it was just the other error that was in our way. Let's see.

No.

Sneha Mehra (00:04:04)  
All right, let's let's let's track this down. So w we are making this list packets response contract. No, we're not making it happy yet. What's it telling us? Distance is not assignable to type number.

Sneha Mehra (00:04:21)  
It wants that to be a distance.

Let me bump the TS server just to make sure I feel like I'm getting conflicting information here. The types of metadata.planting distance are incompatible. Distance is not assignable to number.

Sneha Mehra (00:04:40)  
There's the distance.

you know what we could do here? Satisfies ⁓ seed metadata, seed packet metadata. There. So now like this is a great example of where this satisfies keyword is really helpful. It doesn't actually change the type like using the as keyword, but it does help sort of push this error in ⁓ quite a bit lower lower.

Sneha Mehra (00:05:10)  
Expectative comes from property planting distance, which is declared here. Distance is not assignable to number. So it really does want a number. Is packet planting distance in the YAML?

That's a c good question. ⁓

Sneha Mehra (00:05:29)  
no, I know what I'm doing here. Moving too fast. So all this all this function does is it it normalizes the distance to ⁓ to feet. So if we do dot value, right, it still has a unit and a value here. So if we do value, this should make everything happy. Yep, and it does. So all all we've done here is we've taken the plending distance, whatever it is in, centimeters or whatever, and we've converted it to

An equivalent distance value object that is expressed in feet. And then because our UI can operate purely within the dimension foot, we we're just passing this back as a scalar. And hopefully now when we turn this around, yep, we get two feet. And hard coding feet in there is appropriate now because part of the contract between the back end and the front end is we're always ⁓ operating in terms of feet.

And we'll get into why in the next chapter.

—-----------------

**Sneha Mehra (00:00:00)**  
**The last topic I want to cover today is ⁓ dealing with this concept of ⁓ aggregates in domain-driven design. And this has to do with thinking about like ⁓ thinking about consistency. By that I mean data consistency. So imagine imagine this scenario. Like we're we ⁓ we're able to move plants within a raised bed. And let's say we build a system ⁓ that kind of is modeled this way.**

**Or we would say**

**Sneha Mehra (00:00:35)**  
**We've got like a task that needs to be befor performed.**

**Sneha Mehra (00:00:47)**  
**And we build a nice little like a semantic action, right? Instead of saying like I have this big update API call and I'm gonna like set the new ⁓ array of your items within your bed, and here are the x, y coordinates. Let's say we edit like a really, a really nice thing that kind of looks like ⁓ like this.**

**Sneha Mehra (00:01:17)**  
**Right? Something like this. Spiritually.**

**And ⁓ let's say our app's really popular ⁓ and we end up having to get big, beefy databases to handle all of this. And we decide that, you know, what we can do is we can say, ⁓**

**We implement something called like called charting, which is the idea of saying, you know, we actually have multiple databases and like ⁓ Seth's raised beds and my raised beds, they can actually be on different databases. It's fine. We can co kind of like spread the data out a little bit, like there's no ambiguity, like Mike's raised beds are always on database A, Seth's are on database B, and everything's fine.**

**And and maybe in fact we can sort of mix them up a little bit. We're like, I might have some beds on database A, some on database B, but it doesn't affect this because this you can think of this as like one write happening in one place. But what happens when we want to write across those different concepts? You could imagine how we would say we've got bed one and bed two, ⁓ and we've got a plant.**

**Sneha Mehra (00:02:36)**  
**Like when we start getting into this**

**It gets really interesting where we're trying to move this ⁓ into another bed. Now, if we were to say, look, there's an update plant position, or let's say it's ⁓**

**Sneha Mehra (00:02:55)**  
**Something like this.**

**Sneha Mehra (00:03:04)**  
**Think about what might happen if like ⁓ one ⁓ side of this fails and the other succeeds. This inevitably happens once in a while. You know, we think about our like big distributed systems as having some number of nines of reliability. And sometimes like errors are thrown. But imagine a world where like I'm trying to move plant from bed one to bed two, and the remove plant operation fails, but the add plant operation succeeds.**

**And so now I've cloned this plant somehow. Like there are two of it. ⁓ Or you could have it disappear entirely where you remove it and then add plant fails. And you could say, well, like ⁓ let's check the success of the first call before proceeding with the second. But sometimes, like, you know, sometimes you lose state if you do that. Like, okay, you check to see if ⁓**

**Like imagine remove plant is successful, add plant fails. Are you gonna like add it back into the prior raised bed? Well, you're trying to do there, if if you go in that direction, is you're trying to create this like illusion that there is an atomic operation that we're like both of these things, the removal from the old bed, the addition to the new, ⁓ is ⁓ a single operation that either all succeeds or all fails. And this is really important when we think about like our**

**Like what is the aggregate here? ⁓ And I would say in this case, it's this.**

**Sneha Mehra (00:04:38)**  
**Garden. Because what you could do is you could say, no, no, no. It's really garden.**

**Sneha Mehra (00:04:53)**  
**It's really like this. ⁓ And what you could do, ⁓ if you get deeper into databases and things, you can create what's called a database transaction, which is it's basically doing the ⁓ in one atomic operation that either all succeeds or all fails, it's the removal of the plant from bed one and the placement of the plant in bed two. Inevitably, like especially if you work on something with significant complexity, you're going to have to make these choices. You're gonna have to decide.**

**Where ⁓ like what are the atomic ⁓ operations you can perform? And ⁓ usually what that means is like we would say ⁓ this ⁓ actually I'm I'm gonna change the model here a bit. We'd say this whole thing, and I'm gonna have to move this to back. This is the aggregate.**

**Sneha Mehra (00:05:47)**  
**that whole larger rectangle. ⁓ And**

**The garden, we would say, is kind of like as l as like a lot of aggregates have. Can I remove this to the front, please? Yep, perfect. And this to the front.**

**Sneha Mehra (00:06:13)**  
**Well, we can move it in here, it's fine. ⁓ you'd say, well, garden ⁓ is ⁓ kind of like the root node. And so often when you pick an aggregate, you have to say, like, what is the entity that's sort of the main purpose of this thing? And yes, there may be like a lot of other things embedded within it. But this the the important thing is like choosing the transactional boundary. And in doing that, you get some you're you're making choices. You're saying, Well, within this boundary**

**Brown box here, that's where we are internally consistent. Like you you're never gonna be able to like load the page at a weird time and see a plant has been added to a new zone, but not removed from another zone. And like that would be internally inconsistent. But it it does also mean like you're when you make choices like this, you're designing for in inconsistently, inconsistency, at least momentarily. ⁓**

**To appear in other places. Where it may be okay in a neighborhood where you're like moving a bunch of plants around, like, all right, two gardens, ⁓ two different gardens may be like one one is totally fresh data and the other is like somewhat stale, but at least**

**The totality of data within each of those two gardens will be like the same level of freshness, if that makes sense. So thinking about this as part of your domain modeling ⁓ is really important, especially like working at Stripe when you're thinking about like ⁓ transactions and refunds and ⁓ ledgers that have to all add up so that like at any moment in time when you load this page, like you're not seeing that the balance in your account is like different than what you're seeing in another page.**

**So sometimes like ⁓ these can be really important to your user. ⁓ And ⁓ personally, like I find this to be kind of one of the trickiest areas to have those relatable discussions with users. ⁓ and often you want to hone in on, you know, this idea of of freshness. Like, is it okay to view this data if it like if it's stale, but at least it's all the same amount of stale and you're never seeing like a partial.**

**Sneha Mehra (00:08:26)**  
**A partial state there. And if you talk to an accountant about that and they're like, you're gonna show me like a ledger where like certain items are ⁓ still waiting to sort of percolate through, like that's useless. That's that's gonna be a real problem there. ⁓ So this is the concept of aggregates and designing transactional boundaries with intent.**

—-------------------

Sneha Mehra (00:00:00)  
The last topic I want to cover today is ⁓ dealing with this concept of ⁓ aggregates in domain-driven design. And this has to do with thinking about like ⁓ thinking about consistency. By that I mean data consistency. So imagine imagine this scenario. Like we're we ⁓ we're able to move plants within a raised bed. And let's say we build a system ⁓ that kind of is modeled this way.

Or we would say

Sneha Mehra (00:00:35)  
We've got like a task that needs to be befor performed.

Sneha Mehra (00:00:47)  
And we build a nice little like a semantic action, right? Instead of saying like I have this big update API call and I'm gonna like set the new ⁓ array of your items within your bed, and here are the x, y coordinates. Let's say we edit like a really, a really nice thing that kind of looks like ⁓ like this.

Sneha Mehra (00:01:17)  
Right? Something like this. Spiritually.

And ⁓ let's say our app's really popular ⁓ and we end up having to get big, beefy databases to handle all of this. And we decide that, you know, what we can do is we can say, ⁓

We implement something called like called charting, which is the idea of saying, you know, we actually have multiple databases and like ⁓ Seth's raised beds and my raised beds, they can actually be on different databases. It's fine. We can co kind of like spread the data out a little bit, like there's no ambiguity, like Mike's raised beds are always on database A, Seth's are on database B, and everything's fine.

And and maybe in fact we can sort of mix them up a little bit. We're like, I might have some beds on database A, some on database B, but it doesn't affect this because this you can think of this as like one write happening in one place. But what happens when we want to write across those different concepts? You could imagine how we would say we've got bed one and bed two, ⁓ and we've got a plant.

Sneha Mehra (00:02:36)  
Like when we start getting into this

It gets really interesting where we're trying to move this ⁓ into another bed. Now, if we were to say, look, there's an update plant position, or let's say it's ⁓

Sneha Mehra (00:02:55)  
Something like this.

Sneha Mehra (00:03:04)  
Think about what might happen if like ⁓ one ⁓ side of this fails and the other succeeds. This inevitably happens once in a while. You know, we think about our like big distributed systems as having some number of nines of reliability. And sometimes like errors are thrown. But imagine a world where like I'm trying to move plant from bed one to bed two, and the remove plant operation fails, but the add plant operation succeeds.

And so now I've cloned this plant somehow. Like there are two of it. ⁓ Or you could have it disappear entirely where you remove it and then add plant fails. And you could say, well, like ⁓ let's check the success of the first call before proceeding with the second. But sometimes, like, you know, sometimes you lose state if you do that. Like, okay, you check to see if ⁓

Like imagine remove plant is successful, add plant fails. Are you gonna like add it back into the prior raised bed? Well, you're trying to do there, if if you go in that direction, is you're trying to create this like illusion that there is an atomic operation that we're like both of these things, the removal from the old bed, the addition to the new, ⁓ is ⁓ a single operation that either all succeeds or all fails. And this is really important when we think about like our

Like what is the aggregate here? ⁓ And I would say in this case, it's this.

Sneha Mehra (00:04:38)  
Garden. Because what you could do is you could say, no, no, no. It's really garden.

Sneha Mehra (00:04:53)  
It's really like this. ⁓ And what you could do, ⁓ if you get deeper into databases and things, you can create what's called a database transaction, which is it's basically doing the ⁓ in one atomic operation that either all succeeds or all fails, it's the removal of the plant from bed one and the placement of the plant in bed two. Inevitably, like especially if you work on something with significant complexity, you're going to have to make these choices. You're gonna have to decide.

Where ⁓ like what are the atomic ⁓ operations you can perform? And ⁓ usually what that means is like we would say ⁓ this ⁓ actually I'm I'm gonna change the model here a bit. We'd say this whole thing, and I'm gonna have to move this to back. This is the aggregate.

Sneha Mehra (00:05:47)  
that whole larger rectangle. ⁓ And

The garden, we would say, is kind of like as l as like a lot of aggregates have. Can I remove this to the front, please? Yep, perfect. And this to the front.

Sneha Mehra (00:06:13)  
Well, we can move it in here, it's fine. ⁓ you'd say, well, garden ⁓ is ⁓ kind of like the root node. And so often when you pick an aggregate, you have to say, like, what is the entity that's sort of the main purpose of this thing? And yes, there may be like a lot of other things embedded within it. But this the the important thing is like choosing the transactional boundary. And in doing that, you get some you're you're making choices. You're saying, Well, within this boundary

Brown box here, that's where we are internally consistent. Like you you're never gonna be able to like load the page at a weird time and see a plant has been added to a new zone, but not removed from another zone. And like that would be internally inconsistent. But it it does also mean like you're when you make choices like this, you're designing for in inconsistently, inconsistency, at least momentarily. ⁓

To appear in other places. Where it may be okay in a neighborhood where you're like moving a bunch of plants around, like, all right, two gardens, ⁓ two different gardens may be like one one is totally fresh data and the other is like somewhat stale, but at least

The totality of data within each of those two gardens will be like the same level of freshness, if that makes sense. So thinking about this as part of your domain modeling ⁓ is really important, especially like working at Stripe when you're thinking about like ⁓ transactions and refunds and ⁓ ledgers that have to all add up so that like at any moment in time when you load this page, like you're not seeing that the balance in your account is like different than what you're seeing in another page.

So sometimes like ⁓ these can be really important to your user. ⁓ And ⁓ personally, like I find this to be kind of one of the trickiest areas to have those relatable discussions with users. ⁓ and often you want to hone in on, you know, this idea of of freshness. Like, is it okay to view this data if it like if it's stale, but at least it's all the same amount of stale and you're never seeing like a partial.

Sneha Mehra (00:08:26)  
A partial state there. And if you talk to an accountant about that and they're like, you're gonna show me like a ledger where like certain items are ⁓ still waiting to sort of percolate through, like that's useless. That's that's gonna be a real problem there. ⁓ So this is the concept of aggregates and designing transactional boundaries with intent.

—-------------------

Sneha Mehra (00:00:00)  
The first domain modeling concept we're going to dig into, the first pair of concepts, are value objects and entities. What are value objects? These are ⁓ things that you have to create some sort of data structure for. It could even just be like some special kind of string that has ⁓ a convention to it. Doesn't have to be fancy. But these are things that don't possess a unique identity. What do I mean by this? Like

If ⁓ if you had something like an RGB color, which already exists ⁓ in in this app, like here's the path in the workshop project, which we'll check out in a moment, ⁓ you can see that I'm modeling an RGB color. And I've got a red channel, a a green channel, a blue channel, and an alpha channel for transparency. Now, we're not going to end up storing in in a database in our little SQLite database that's

Holding all the seed packets and things behind the scenes. We're not gonna be storing like a table of colors where each color has an ID and referring we're referring to colors by ID. I g I guess you could if you were doing some sort of theming thing, but this you want to think about as being kind of embedded ⁓ onto more interesting things. So we also have a concept of a distance that already exists in the in the app. ⁓ there's a value, like three, and then there's a unit, like feet or centimeters, or meters.

Those are value objects. And an important aspect of value objects is you should be able, through some code that you write, to be able to like compare them against each other and check whether they're equal, ⁓ not by just checking like ⁓ ID equals ID, but you're you're trying to figure out are these things equivalent. So this gets a little bit interesting when you start thinking about units. Like, is three meters equal to some precise number of feet?

Maybe those are equal. Like that would be up to you to think about. Like, talk to your user. Does ⁓ like does this make sense? In that case, I'd argue probably. So there are a couple things going on here in this code. First off, we're using a library called type ORM. ⁓ don't worry if you've never used this. We're not gonna have to worry about it too much, and we'll incrementally get into this. But this is ⁓ a tool for object relational mapping, and this just means.

Sneha Mehra (00:02:22)  
We don't have to set up a database table and we don't have to write SQL queries. There's nothing wrong with that if you if that's the way you like to engage with the database. But ⁓ simply by creating a class like this and denoting that like these these fields represent columns, right? Where we're saying alpha is a column and if it's not specified, let's assume it has a value of one, all the database stuff happens ⁓ behind the scenes for us, and we don't have to worry about it.

The second thing is we've got this pshuot slash types package. We'll dive into this a little bit more, but I have I've factored out ⁓ some shared types that our front-end and backend ⁓ both depend on. And in there you'll find common things that relate to sort of the shared contract between these two system components. So when our UI sends a request to our backend, there is a clear type for what that request shape and the response shape looks like.

As well as any other objects that are embedded within that request and response. So that's that's what you can find in this types package. In the types package, you'll see we're using a library called Zod. And again, if you haven't used this before, don't panic. We're going to get into it incrementally. But this is a way of defining schemas. And so this adds on top of TypeScript ⁓ runtime type checking. So by defining an RGB color like this, we get something that we can use at runtime to validate that like.

Yes, the red, green, and blue properties are on this object, and they are in fact numbers. ⁓ And you know, you can see that we can mark something as optional or required. There's a lot here. And the I want you to ⁓ start with the assumption, it's not strictly two, but start with the assumption that if you can articulate a type in TypeScript, Zod you can use to sort of articulate a a schema and extract a type from that schema so that you can have both.

Compile time type checking from TypeScript and runtime type checking from Zot. Of course, this doesn't come for free. Unlike TypeScript's types, this doesn't drop away as part of your compile process. ⁓ But that's that's kind of the point. So that's value objects. They're sort of ⁓ comparable things that typically get embedded into other resources, and they don't have ⁓ a an identity field like an ID. Now let's talk about entities.

Sneha Mehra (00:04:47)  
So ⁓ entities are ⁓ mutable objects. Value objects are generally immutable. You create a new one. It like if you want to change the color of something, you create a new color that's a slightly different color. Entities are different. These represent ⁓ things that you will persist. They are mutable. They can change. ⁓ And generally, at least in this app, they're all going to have an ID of some sort.

So if we're gonna take a peek at what our database looks like, and you don't have to be a database expert here, but like you're gonna see that each of these corresponds to a table. ⁓ And the way you can tell the difference between a value object like this and an entity ⁓ is this entity decorator that you're adding to the class. This is what signifies that this is sort of a top-level thing that ends up being persisted in a database. ⁓

When you start working with entities, relationships get really interesting. You can have a has-many relationship or a one-to-one relationship, ⁓ whereas that color concept just ends up kind of being embedded in the object, if you will. We're going to walk through the process of checking, checking the get repo for this gardening app out. ⁓ And we're going to kind of warm up and do some very basic modeling, like figure out what entities are important and value objects. ⁓ For a part of this app.

Called the temperature date calculator calculator. So ⁓ the way ⁓ the the problem to solve here is ⁓ as a gardener, I want to be able to select a location, type it in a temperature, and I want to know ⁓ what is the date where I can be assured the the temperature in my location will be ⁓ at or above ⁓ that level. So this is like a concrete example here is

Tomato plants, they want to go outside when ⁓ the nighttime temperatures are above fifty degrees Fahrenheit. And so I need to know like what is that date? That's very important to me because I work backwards from that date to sort of like how long do they have to grow inside, and then when should I plant my seeds? And sort it's it's a really important part of sort of working back from the date they go outside.

—----------------

Sneha Mehra (00:00:00)  
Let's transition back to implementing this. We're gonna need some new types, ⁓ and we're going to need ⁓ some new entities, ⁓ and let's get cooking. So we're gonna go back to our types folder. We have entities to create, not our dist folder, our source entities folder. And I'm gonna close out some of the stuff from the last exercise. ⁓ And ⁓ let's see, we we kind of have a little bit of a starting point here. We have a garden, and I think it's it looks like

There's something complicated here, a function that generates a type. So we're gonna create a workspace schema. Oops.

Sneha Mehra (00:00:39)  
this is ⁓ this is actually not right, but it's a good opportunity for us to go and fix this. ⁓ what we want is ⁓

Build on this concept of zone. So, zone, if we think about our other bounded context, I'm gonna go back ⁓ to this. Right? We've got workspace, that's the garden. Zone, that's the bed within the garden. Item placement, we'll call that the plant. ⁓ Item is kind of the seed packet, right? It's it's the source of information for like what is this

What is this thing all about? So ⁓ let's look at workspace, not zone. Workspace type. It says create workspace schema for ⁓ item type. And we can see, great. It takes in some arguments. That's like the item type, the metadata for the workspace, the metadata for the zone. ⁓ And it's got an array of zones. ⁓ And then it's got some metadata. So we have opportunity for metadata on the garden, and then a zone below that. We've got

Placements. ⁓ So th those are those like item placements. This is the pla like a placement would have to have like an xy coordinate of some sort. And let's see if that looks right. Hey, look, we've got a position with an xy coordinate and an item. ⁓ And this might this might be a seed packet or something like that. And then what does it belong to? Yes. ⁓ Have you tried ⁓ using

Have you tried using Cursor to scaffold some of these types before? Absolutely. Cursor is quite good at scaffolding these types. And especially, like my my big pro tip here would be: if you have like an entity that's related to this, ⁓ leave that information in the code comments. Like leave yourself a little like a ⁓ a relative path or sorry, a path relative to your project route to follow. So that if

Sneha Mehra (00:02:48)  
Cursor finds itself updating this. Like one of the first things it does is it reads the entire file. And if you've left notes, like whenever you update this, make sure to go and check that it threads through this other thing. Cursor will be pretty good at at figuring that out. I'm not doing that right now because I would just start typing a couple characters and it would just be ⁓ spitting out, spinning out a lot of things, and it gets a little bit more difficult to to kind of follow what's going on. So ⁓

We've got item placement, we've got zone, and we've got workspace. So let's let's back up here and go to garden. great. So we've got create workspace schema for item type. We pass in the plant schema. We pass in some object that represents metadata on the zone. I believe that's what this is. ⁓ zone metadata schema is the first one, and then workspace metadata schema is the second one.

Just double checking that. Yep. Zone and workspace. Great.

—----------------------------  
15  
Sneha Mehra (00:00:00)  
The next topic we're going to talk about is what domain-driven design describes as bounded contexts. And you could think about a bound a bounded context as kind of an area of a problem space or a domain model where a ⁓ a ubiquitous language, right? A common shared language applies ⁓ and there is internal consistency within a bounded context. So for for example,

If you had a very generic model name like item, where, you know, maybe we consider there's like an item planted in a raised bed and it happens to be a plant. Or maybe item means other things ⁓ in other contexts. Like the the purpose of this bounded context is really to create a bubble where you don't have to add all these namespace qualifiers where you're like, this is a grid placeable.

Item with ID and metadata. Like that's what we're going to call our class for for a particular entity. And part of the value of this is it's almost like identifying a related family ⁓ of ⁓ of like entities and their relationships and their constraints that you can kind of like draw a circle around and say, these are this is a neatly bundled thing. ⁓ And in and ⁓ like ⁓ the TypeScript monorepos course, like we might consider making this a separate.

package because there's some self-contained complexity there. And we can use simple terminology while while always noting we're operating within that context. ⁓ Well, it turns out that ⁓ there's another context in this app beyond sort of seeds and plants. ⁓ there's the context that has to do with this user interface, which we're about to start working on, for being able to drag plants into raised beds and arrange them and get some feedback based on our planting. ⁓

Our UI has a ⁓ model that looks like this. ⁓ We have workspaces, ⁓ zones, ⁓ item placements, and those item placements relate to items. So you could think of this as sort of it's a very generic UI. Like you could use this same drag and drop UI to build a chess game or something like that. Like it in essence, it's a grid with

Sneha Mehra (00:02:25)  
Draggable tiles that are placed at XY coordinates on a grid. ⁓ And maybe they relate to a chess piece, so we might have a model of like bishop or knight or pawn. And that would be part of the visual representation of that thing on the grid. And like maybe there are multiple grids and you can move things across chess boards for some reason. The analogy is breaking down. But like the UI has not no idea ⁓ about this gardening concept.

And so what we would say is, well, that's a bounded context. Item in this case is just like in this world. ⁓ Well, we've since ⁓ since interfered with our plant construction here. But in in the world of the gardening app here, it's just sort of like these things that ⁓ you can drag onto the grid. Like it's it's the pepper plant or it's the cherry tree or whatever it is. ⁓ And so ⁓

We're about to create a a gardening related bounded context of our own that will involve, like sort of at its root, we may have this concept of garden, but there are different beds ⁓ in that garden and plants within those beds. And so we're gonna have to model all of that, but we can say like the terminology we're going to use when speaking garden belongs in that bounded context. And when we're speaking about grids and workspaces and zones.

That belongs in the other context.

Alright, ⁓ but how do we bridge those worlds? Well, we need what's called ⁓ an anti-corruption layer. This is like D D terminology, but you have built these before. It effectively is a way to keep kind of clean modeling in in two areas. Or sometimes you have clean modeling and very dirty modeling in like a legacy code base or something. The anti-corruption layer is a very ⁓ closely scoped layer.

Sneha Mehra (00:04:26)  
That you use to kind of convert between the important domain models ⁓ that are at the essence of the app you're building and whatever else they need to talk to, right? This this is what's going to avoid kind of contaminating our garden and our raised bed and our plant and our seed packet with like the drag and drop UI concepts. ⁓ and that we we have this already, right? Today it it kind of exists in the route handlers.

Where we're adapting between what our domain services spit out, like a seed packet, and the HTTP response that needs to be passed back up to our user interface. So you might argue that's kind of an anti-corruption layer. It would be even cleaner if we were to say, you know what, we're actually gonna turn some of that transformation. We're gonna like refactor it into functions, and we might create like a new package that's that purely serves as this layer to adapt between our ⁓ service.

or domain services and what the UI wants to get. And then the request handling is just sort of leveraging, ⁓ leveraging that. And we can further separate concerns that way.

—---------------------------------  
23  
Sneha Mehra (00:00:00)  
Thanks for watching this course about domain modeling and collaborating with humans and AI. I hope you learned some tips and tricks for having useful conversations with users and aligning around a shared language so that you can bring that language into your software and increase the likelihood that you will solve important problems for your user and you'll avoid lossiness in communication, which can result in overly complicated or misaligned ⁓ software products.

