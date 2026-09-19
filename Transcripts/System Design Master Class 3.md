2

Sneha Mehra (00:00:01)  
Great week 6th, 6th, day 1st, we will talk about Hyderabad system. Actually, it's not really Hyderabad. Hyderabad is an extension to storage. So what happened is earlier, ⁓ first three or four cohorts I'm talking about, ⁓ I used to cover S3 in storage week itself. ⁓ So with 5 itself, we used to cover S3. But then I found

Like, hey, I should go a little more deeper into the way I'm ⁓ the kind of things I'm covering and because people are comfortable with those concepts. So that's where I moved S3 to this week, giving enough space for other discussions. ⁓ yeah, it's like over time, the course has evolved really well. the agenda for today ⁓ is ⁓ we'll design S3 that ⁓ is going to be a huge discussion and which is where you understand that

Hey, the systems are not as simple as we think they are. yeah, that's the beauty of it. ⁓ And ⁓ we'll talk more about tiered storages because I want to go really deep into tiered storages because that is one of the most underrated way to get performance out of a system while keeping the cost on. ⁓ So we'll talk about S3 after half an hour. But before that, I'll just switch the order and cover

a cost efficient order storage system. ⁓ This is we would start with first. This is where you get glimpse of data engineering, ⁓ quote unquote data engineering and how those systems are built. And then we start with S3. ⁓ We basically cover it to a 20 by 30 % on brainstorming. Then we take a break and then we cover complete testing. ⁓ So around 11.30, 11.40 input go and then obviously open questions at the end. ⁓

Okay, so before that just want to touch upon ⁓ bunch of continuous feedback that I got. ⁓ I would request you to split out class to two ⁓ three hour class at a time is very lengthy and able to focus with the same energy till the end. I agree but ⁓ that but that's the beauty of it. I'll still think because if I split it, it becomes way too short does not seem ⁓ like we are actually covering decent enough. For example, if I cover ⁓

Sneha Mehra (00:02:21)  
that ⁓ what was that ⁓ real time communication that itself went for two hours 20 minutes and S3 will go for two and a half hours. ⁓ So splitting that would become way too small. But the way the courses evolved, I felt this way, but point taken, I'll try to see like if I can split something into two. In those cases, I'll give longer breaks like 10, 15 minutes breaks and we'll try to manage something. ⁓ point taken.

⁓ It's hard. can understand that totally. And ⁓ can you please request viewers to not answer questions asked by you while discussing any topic in the chat box? It's not my control. It's fun to have like chat. So folks who ⁓ are distracted by chat box, just stay away from it ⁓ and brainstorm on your own. But it's fun when people post things on chat boxes. ⁓

⁓ I cannot stop that because it would make things very monotonous. ⁓ It's better. And in case you are someone who gets affected by seeing the prompted answers in the chat box, just keep the chat in the midst. ⁓ And when you feel like that, then open. I, take it, like to each its own. ⁓ But just keep your chat closed. ⁓ There is an option to do that. This way it would work out well. ⁓ Okay.

These were the two. I have one more I received last week which I forgot to cover is when are we going to have one on one. So I'll not schedule it. You have to schedule because it's on demand. You have to schedule. So whoever wrote that question in feedback, please make a note that link is there on the log. Go and log by calendar for that. ⁓ Okay. Great. Fine. Let's move to the technical side of things. Cost efficient order storage system.

So this is where ⁓ for next half an hour, I'd want to talk about like why multi-tiered data stores are so important and how they come in really handy. ⁓ it most like, and why orders I'm taking as an example, because it fits this use case really well. ⁓ So ⁓ I touched upon fraction of it last week, but we'll go deeper into how a system like this is built and how you should build a system. ⁓

Sneha Mehra (00:04:41)  
And it's very much inspired from a lot of other systems that already exist. ⁓ Something very similar. also worked on. So I can very closely relate to it. Okay. This is the kind of architecture that typically happens. Okay. Let's start with the first one. So transactional systems, you typically have like whenever you ⁓ let's say building an e-commerce system in the e-commerce system, you take orders. So you need to, when you take orders, you need to store orders somewhere. So that is what you say. Let's say if you store those

orders that are coming in in a transactional database. Now when these orders are being accepted, they are being stored in the transactional database. Now what happens ⁓ is other services need this order information. For example, your payment service needs this order information so that it can update its own status and basically collect payment and EMIs and whatnot flow. The logistic service needs this order information so as to know, okay, which order am I delivering?

And it would update that corresponding entry in those database. ⁓ Hey, because in the order space, you see that, Hey, you need this order. ⁓ This is where it is at the stage of delivery, right? ⁓ It's shipped, ⁓ it's packaged, ⁓ dispatch ship out for delivery and whatnot. ⁓ Then customer support also needs this order information in order to solve your queries. Right. When you reach out to, let's say Amazon customer support, they would have your order. So a lot of systems are dependent on orders, but

Here, when a lot of systems are dependent on this one single subsystem, you realize that this system is not just handling, this system is not just handling end user requests, but also request from other internal services. And then you say, okay, ⁓ when I'm getting these many requests, scaling of API servers is fine. We all know just put a load balancer, add more servers, define an auto scaling policy, your API servers are stateless and it can very easily scale.

Things become tricky because of database that you have. And you say, okay, now my database, need to scale my database. So what do you do? You apply three very common ways to do it. First, you try to scale up. You try to vertically scale the database until a certain limit. Then you say, ⁓ have large amount of reads, which is very much there because customer, why would customer support write to your orders? It would not. It would only read from it. Right? Logistics. Why would it write? It would just read that which, which package am I delivering? Right? Payments.

Sneha Mehra (00:07:02)  
would also read. So you see a very high read traffic happening on your order service because of internal services that are coming in. Right? Right. So typically from the customer side that are happening, right? Few rights internal also there, but you get the idea. say, Hey, let me just add read replica. Right? Let me add read replica on my master database and let I'll just transfer my reads to read replica. This is I be able to scale scale the reads request that I'm having. So that works well. Now what happens is

Now you get even more. You are getting more orders, which means there are more write requests, which one note cannot handle, which one vertically scaled up node cannot handle. ⁓ What do you do? You shard it. Right? So you break your master into subsets or into end databases with end mutually exclusive subsets of data. ⁓ Right? ⁓ For example, orders for from user one to a hundred over here, a hundred to two hundred over here, through a hundred to three hundred over here. ⁓ Let's say there are only three hundred users. So you create mutually exclusive groups.

and shard it up. Right. We saw this approach is far too many times up until now. Right. But now when you do this, it's not always a good choice that you are doing it because you are now it's a little more complex sharding, especially when it comes to relational database, you have to typically manually shard the database and have your API server know where to forward the request to. Right. ⁓ So sharding is OK. You can still do it if you have good engineering team. But now think about it.

I'm like, how long would you just keep on horizontally scaling the data? Right? Because now after a certain point in time, you would have large number of these databases, large number of unnecessarily large number of these database. So what can you do? It's like, no, but after starting, what else, what else can we do? The thing is, when you have, like, if you look closely at this stage,

of your database architecture, where you have read replicas. At this stage, you must figure out that this one database was not able to handle lot of write requests, which is where you were ⁓ doing sharding. But now what you should look at over here ⁓ is that what is the root cause of degradation? Why is your DB not able to handle the load? ⁓

Sneha Mehra (00:09:27)  
crude optimization that you can do over here is hey, ⁓ whatever data I have in my database, ⁓ am I accessing all of them? Do I need all of the data to be present in my main transactional database? This is the question that we need to ask because now what happens is let's say you have orders table, you have huge orders table. Let's say Amazon takes orders, let's say 100,000 orders per minute. And I'm not sure what the numbers are, but let's assume that 100,000 order per minute.

So now what happens is your orders table will have huge number of entries. ⁓ When it has huge number of entries, then you, whatever you are querying on it, even though you have indexes, ⁓ Huge table will have huge index. Huge index, need to fit this huge index in memory. Now, which means that let's say there are 10 million orders on Amazon and you created an index of the 10 million.

that index of 10 million orders needs to be fitted to memory. Otherwise spill on the disk and whatnot would be there. When you have this much of data, ⁓ are you actively querying all of the data? Not really. Because this is where what you see is you see degradation in performance because there is huge amount of data that you are putting into your database. Into this one node that you have, the amount of data that you are putting in is huge enough.

Do you really need all of the data? This is a critical question that you should ask because in most cases, whenever you are storing any information anywhere in the database, every single row, every single object has a life cycle to it. At some point in time, that object is accessed very frequently and then the pattern deteriorates. Then it degrades. Then the amount of the number of accesses that would happen.

is much lesser than what they used to be. So every object in every system, literally every single system has a life cycle to it. And the life cycle depends on the use case. In some cases, you'd see that most recently posted stuff gets very high traction. After that, no one gives a damn about it. But for example, you tweeted something, it's very likely that you would get very high traction during the first phase of it, like first one or two days.

Sneha Mehra (00:11:51)  
And then no one cares about it. So you see very less engagement into it. ⁓ And similarly for orders, ⁓ when the order is very fresh, ⁓ you're doing a lot of reads and writes onto it because just imagine the order is just placed. ⁓ Immediately your payment service would be querying it. You might be just reaching out to customer support to get things done. ⁓ Then your logistic would be constantly querying this to know what you're delivering, how you're doing, ⁓ updating the database and what. So lot of information would be getting read and updated when an order is newly created.

But once your order is delivered, logistics is out of question because logistics is done, right? Order is delivered. Now logic has no like no reason to query your orders. So then this goes out of window. Payment is done. Payment is done in one one or in 10 minutes after the payments also does not need to do anything. Then what happens is once your order is delivered, then for let's say ⁓ one month, you have your warranty period, which is where you're likely to access

which is you are likely to call your customer support to get something. So that's where you would see only load from query, sorry, from the customer support team and user in case user is going through historical orders. ⁓ But you see that in classic case of orders, there is a very high access that you're doing reads and writes both when order is very new. Before order gets delivered, you see a lot of frequent reads happening.

After that, no one cares. ⁓ If order is let's six months old, no one cares on what's happening. Tickets are made some order. Maybe during once a year, during your filing of taxes, you go through your orders and see, okay, this is huge amount of money that I spent and you go under. So every single system in the world, wherever you are storing this data, whatever data you are storing, try to understand the life cycle of the data. How?

Frequently it is being used when it is being used when that access frequency drops. Now you'll say, but how do I know? ⁓ First metrics, right? You can either clock where the particular item is accessed, how frequently it is accessed that otherwise you can just guess. Right? That also works like in case of orders, you can very easily guess that, Hey, this is how this access pattern should be. But if you haven't quantifiable metrics, good enough. If not even a guess works fine. Right? This is, this holds true for almost

Sneha Mehra (00:14:19)  
all systems out there. Social networks. How likely are you to access your Facebook post made in 2008? ⁓ Very unlikely. ⁓ How likely are you to access your own tweets that you made one month ⁓ ago? ⁓ Highly unlikely. ⁓ Given that, it's very unlikely for you to access it. So which means that there would be an access pattern for anything. ⁓ Now, obviously there are aberrations, there are anomalies in it. Because there are cases where a historical

post suddenly gets traction and it goes viral. And that is exception. But in most cases, for most people, it's a 99.99 % of cases, access pattern would remain fairly like this, which is where now what you can do is you can reduce the load ⁓ on the database by taking the data out of this database. This would help you save cost for the organization.

So typically ⁓ building a multi tiered storage is a way for organization to save cost by reducing the size, the amount of data that you are storing in the database. That's where the highlight is. That if I want to store, let's say 1 million orders in my database, out of which I'm only accessing 10,000 at this stage. Why the hell am I storing this much of data? Let me move it somewhere else. Now this database is small. The amount of data is small. The amount of data that I'm storing in a table is small.

which means indexes are small, which means indexes can fit in memory, which means my database is performed. This is where tiered databases or tiered data store comes in. So there ⁓ is a very common nomenclature around it, but it's not a hard bowed nomenclature. Hot, warm and cold, it's related to temperatures, ⁓ people just exchange, people say warm store is nothing and hot and cold people refer, but...

you'll get the idea when I talk about it. So hot store is where, which is a main transactional database, right? For example, for an order system, the MySQL database where when user places an order, the database where it goes in and creates that entry is a transactional store. That's where your user facing transactions are happening. So that may be MySQL, maybe DynamoDB, whichever database you are picking, that is where your reads and writes happen, right?

Sneha Mehra (00:16:43)  
These databases are low latency databases like in MySQL. When you write into it, it's low latency database. ⁓ It gives a very strong consistency. ⁓ Because that's what you need when you are doing end user basic transactions. But they're expensive. say, but MySQL is cheap, ⁓ right? Caches are expensive. This is relatively expensive. ⁓ When you're seeing at scale of Amazon, if I take example of that, ⁓ the amount of data, the amount of transition that they are handling is relatively more expensive. What do you do then?

This is hot tier. This is your first tier of data, which we typically design. Then comes warm tier. So warm tier is typically a tier which is there for your read-only workloads. Now this read-only workloads that you have, it means that ⁓ once your data moves from hot tier to warm tier, after warm tier, no one would be doing any rights on this data. You assume that.

Once your data is moved from hot to warm, it becomes read-only. It is non-transactional, which means no one can write to the database directly, like your end user queries will not write to this database. It will always go to transactional, but this is there to support frequent read-only data. For example, you are more likely to read your last six months ⁓ of orders. Right? Beyond that, very unlikely.

So last six months of orders can be put into warm storage when you are frequently reading that particular data. It can be a little slower, little slower, not as fast as hot stored, but it can be a little slower, but it needs to be scalable. ⁓ And that because now you would take data from a hot tier, ⁓ move into a warm tier. So it needs to be distributed, needs to be hot stored and scalable. ⁓ In your hot tier, you can have one database to handle that load.

I'm just using that as an example, but in warm tier because you are putting in a lot of data ⁓ and you want to read them, it needs to be horizontally scalable. And then comes your cold tier. Cold tier is where you have very infrequent reads. Let's say once a year sort of reads that you can dump it on S3. Once a year, if I get even very less reads on that or very slow reads on that, it's fine. ⁓ So cold store is also read only.

Sneha Mehra (00:19:02)  
You typically store ⁓ it ⁓ in a block, something like S3, in that too in a glacial tier. No one gives a damn about it. there. ⁓ It's just there for accounting, legal, when a customer files a lawsuit against an organization, then you say, let's bring out that data, right? All of that. So you are very unlikely to access that data, which is what you put into cold tier. ⁓ So this is multi-tiered storage that almost all big organizations have.

Almost whatever system you are building. Try to see if you can archive something. I'll give very practical examples for that. Messaging apps. ⁓ Imagine, Facebook pay you put in tons of messages to your friends back in 2008 to 2015 for example. How like do you even read them? Does it make sense for Facebook to store it in their hot storage and not even wearing the data? What Facebook would have done? It would have taken the data, put it into

⁓ cold storage like S3, when you access it, then the first read would be very slow, but subsequently then puts into this database and then you access it. But in most cases, it doesn't even allow you to access it. ⁓ So why to waste your expensive storage on things that are not read or queried actively? This is the crux of it for which people build multi-tiered storages. ⁓

So let me talk about a bit more detail into practicalities of how the systems are built. This is this kind of architecture that I'm showing you. It looks overwhelming, but it is not. And this is what any data engineering pipeline looks like. ⁓ So with this, want to touch upon the key concepts of data engineering where with technology fits in and ⁓ how data is moved and whatnot. I'll go a little more deeper into it. And ⁓ again,

really practical way of doing it. Nothing's just like for the sake of it. So what we have is we have orders DB. sorry. We have order service. This is the main transactional database of orders, which is your hot tier, which is transaction. So my SQL Postgres, January, pick your favorite database where all everything hot goes in. Now from this hot, whenever ⁓ your order is new, the data is put into it.

Sneha Mehra (00:21:25)  
orders, payments, logistics, what's not, they all query this particular database to get their answers. Once your order is delivered, which is where you have a job that takes what you want. You want this data to taken from your hot tier and moved to warm tier. ⁓ Because you are not likely to access order after six months, for example. So after six months, you may want to take your orders and move it to warm tier. ⁓ So which is where you need a data deletion policy.

Basically a job that says my data is six months old. So let me copy the data into here and delete from hot tier. ⁓ All encompass into data deletion policy. And ⁓ for every table, the data deletion policy might be something different. Let's say for Snapchat messages, it could be as small as two days because the messages are disappearing. So it's gone two days, you can delete the data. Amazon order six months for something it would be one month. ⁓

So depending on use case, know for how long you want to keep the data in your hot storage. And then what you want to do is you want to take the data from a hot storage and move to warm storage. Now the property of warm storage is that it is read-only. It supports read-only workloads, right? Because here rights cannot go from your end user. The rights would not go into your warm storage at all. It is only meant for read-only use cases, right? ⁓

The way you would be doing it, the way you would be structuring this information would be that it would be very efficient for you to serve the orders. I'll give a small example. ⁓ Once your order is placed, the payment is collected, the logistics is sorted out. Now what you need for you given an order, sorry, given an order ID, you want to fetch all the information from it. Right? So when you're doing that from an order ID, you want to get

all the information about the order. You just don't need order information. You need payment, logistics and everything in one response. Right? So what you want is you are not just storing your order details, but along with the payment, notification, ⁓ payment, logistics and everything in one gigantic JSON document and keeping it handy in your warm storage. So what you do is in your warm storage, you want to take data from multiple sources, aggregate them.

Sneha Mehra (00:23:52)  
and store it. So which is where from order service you get data and you store it into a staging storage. It just a staging storage. Typically it's S3, S3 or any blob storage. So periodically you take data from hot storage, put it into staging storage. ⁓ So ⁓ in this staging storage you get data from your order service. Then similarly from payment service you get data into the same staging storage. From logistic service you get data into the same staging storage.

Data from all these three services, they come to this staging storage. ⁓ Now what you do, you merge them. Because what you want in the end JSON that you are sending, you want to send for a given order, the logistics information, the payment information, the order details, everything. Not just order details, you want to send everything. So you want everything to be part of the same document that you are storing over here. So that because this is anyway costly, you don't have to make multiple calls over

So given we have to ⁓ merge the data, payments, logistics, orders linked to a particular order, you dump all the data into staging storage using a dumper. The dumper is just a name that I given which periodically reads the data and puts it into staging storage. But now here assume there are three files, one for orders, one for payments, one for logistics. Now you will write a simple Spark job that would

go through these three files, join them by order ID and put them into Worms. So this is where data, this is the data pipeline framework like Spark comes in. Spark does fabulous job into loading the data over here, transforming it and putting it over here. Classic ETL definition. So here you have data from orders, payments and logistics, which is read by Spark. The data is joined or aggregated and pushed.

into this warm storage. ⁓ Once your data enters the warm storage, your order service can now, if it's old data, it can read from warm storage. If it's new data, it can go to transactional storage, your hot storage. So this is what your dumper with dumps to staging area goes in. From staging area, your loader loads it, which is typically a Spark job, and it writes to the warm storage. ⁓ Now warm storage, the data is structured such that ⁓

Sneha Mehra (00:26:19)  
given an order ID, I'll give you all the information in one shot. You don't have to query separate data, separate payments, ⁓ logistics and whatnot. Everything is there because this information is already done. Once order is delivered, the logistics change. No, once payment is made, would payment be changed? No, once order is read frequently, now it's going to be little infrequent. Would it make any difference? Because it's all it's literally read only. So what you're doing is you are just structuring the data such that it becomes very efficient.

for you to just make one API call to this warm tier, get the data, serve it to the user. Right? Okay. Now the question is, how would this ⁓ order service know if I want to go to hot tier or cold tier or warm tier? How would it know? ⁓ Right? Because all you have is an order ID. Given that you only have order ID with you, how would you know that order is hot or warm?

because either data is over here or data is over here. Right. But if for every request you make two calls, very expensive. Right. So you need a way that given an order ID, if it's warm or if it is new or not, this is where the ID generation logic comes in. Remember in ID that is snowflake ID, if you're talking about the first 23 bits were epoch milliseconds, that is time. You can use that

Because if you know your deletion policy, you would know that, ⁓ it has been this long. So then where should I make my query to either warm storage or hot storage? It is possible that data might not be in one that you make, then you do a fallback on hot storage. Or if you're making a call on a hot storage, then you do a fallback on warm storage. ⁓ But that first 23 bits of your IDs, which is why having timestamp in your ID and your ID generation logic is very important. It just optimizes the entire flow. ⁓

Otherwise you may want to have another database to store. When was this order place and all, then this needs to scale up. None of that. Stir timestamp in your ID, which is why ID generation is so important. ⁓ Right. Having that time information in your ID is pure gold for use cases like this, because that would clearly tell you if my, like, what are the chances of my data to be there in warm tier and hot tier. Right. Okay. So this is how your data moves from orders.

Sneha Mehra (00:28:45)  
to hot tier to warm tier. Now similarly what we do from warm tier, I want to move data to cold tier. Cold tier is S3, highly unlikely data would be accessed, very large latency, people don't like really query them. ⁓ So this is where your cold tier is. It's typical of our accounting, legal compliance stuff. So what do you do? Similarly, what you have dumpers, staging and loader, you do something very similar and put data onto S3. ⁓

When a data goes into S3, it's very unlikely that you would be directly querying the data. But in case you want to do it, you can use PrestoDB to do it. Now these are all tools that are part of data engineering ecosystem that are very widely popular. So they might seem jargon to you, but just do Google search and play around with these tools because 60 % of infrastructure cost comes from data analytics. So you should know a bit of data analytics, like how these tools perform, how big data analytics tools work. ⁓

When you are storing data on S3, ⁓ either, now let's say if I define that after six months, ⁓ my data should go into cold tier, right? So after six months, my data is in S3. But now let's say I want to query the data. How would I query the data? ⁓ For us to query the data, can I directly make a call to S3 and read the actual pointed data? S3 is blob storage. ⁓ We studied bird dictionary.

In Word dictionary, we were able to fire queries on S3, pointed reads on S3. So for example, the way we define the data format in which we are storing the data on S3, it may open up things for us to directly read the pointed records from S3. Which is where we covered Word dictionary last week, where we defined that somehow it is possible for us to make pointed reads

on S3. We just need basic indexing file, basic data file and we can still do that. And that's not the only format you can define on custom format, which is where your Apache Hoody ⁓

Sneha Mehra (00:31:03)  
Apache Hoody is the format in which data engineering teams write data on S3, giving them an ability to query directly from S3, which is where people typically use PrestoDB or AWS Athena you might have heard of, those things to directly query the data from S3. So if you are making infrequent one-off reads from S3, you can do that. Let's say just want a one-off information. It's not frequent, it's just one-off that, hey, I just need one information.

It's not that in bomb storage. It's not dead in hot storage. It is there in cold. So I just did one of information. One of queries. can still query like this, right? But there are cases, there are cases where you have to fire a lot of queries ⁓ on your old data. For example, let's say there's a police complaint file against your organization. Right? ⁓ Let's say some case happened one year back.

Now your data is in cold cold air and police is asking you a bunch of questions. Now, how do you fire those queries on S3 if you make pointed queries, it would be very slow, right? Plus very expensive because S3 has its own cost. Your infrastructure running PrestoDB has its own cost. So which is when you know that you are firing a lot of frequent queries on cold data. What you do, you load the data into a temporary database, which is queryable. Let's say there was a criminal.

who are rather there is a person who is accused not a criminal there is a person who was accused of a crime you know that person's ID you are downloading all the historical records of that person it's there on S3 you load it into a temporary database and later in the legal analytics and all just query on it because you are making a lot of bulk queries for that historical data so instead of making direct queries to S3 it is better if you load it put it into a temporary database and then query it

So which is why you need a small loader that loads this data and puts it into a temporary storage. Once this job is done, you can delete this data. ⁓ The data is still residing on it. It's not a problem. You are just loading it into a temporary database so that internal teams for legal purposes and all can query it. And this is the most basic data. And this is where like almost encompasses everything around data engineering. The way you are

Sneha Mehra (00:33:26)  
taking data out, putting into staging, ⁓ reading from that and putting it into multiple tiers is what data engineering team does. This is where your spark comes in. This is where your presto DB comes in. This is where your hoodie comes in. So now if you are not accustomed to data engineering as a domain, I would highly recommend take a week, like take one or two weeks of your time and just explore data engineering stuff. There are so many amazing tools, ⁓ libraries, ⁓ frameworks, ⁓ concepts that you would be finding. ⁓

fascinating world out there. ⁓ So, but this is what you would find almost everywhere. So what I did through this thing, I, ⁓ we spoke about the implementation details on how those things are covered, how those data are moved, where components of data engineering like Spark Apache Hoody, would come across this terms. So where it would come in Spark over here Apache Hoody is the format in which it's stored. A lot of people don't like, unfortunately, when you look at the ⁓

the documentation of these tools you don't understand where can I use it which is why just making it little simpler while also covering how the systems are built. ⁓ So this is where you'll find Apache Hoody. This is what Spark would do, load it, PrestoDB you would find it over here which is actively querying on your quote unquote data lake. This is kind of your data lake written in Apache Hoody format. ⁓ Dump and load through staging, putting it on S3, deletion policies and what not. ⁓ Just one caveat.

from our hot storage, how would data be moved over here? This is where the CDC's come in. Right? So from your hot storage, you're constantly taking data or using CDC and putting it there. So air bite, air flow, those two, sorry, air bite and DBZM comes in reading from CDC, putting it out. So what I would recommend is take, and it's big domain, very massive domain. So you need at least two weeks of time to just understand what's happening in it. But this would give you a very nice idea on how, this is how

Most like almost 99.99 % data engineering thing is just encompassed over here. This is what you need typically when you are writing any like if you ask any data engineering, this is typically what they would expect. Right. And now it's your responsibility to pick one component and go really deep into it. How to use it, where to use it, right? Spark jobs, right? Presto queries, see what high is, see what hoodie is, see what fling flume, what not, that's tons of them. Right.

Sneha Mehra (00:35:54)  
This is how you design cost efficient Amazon's order system where you clearly see a dip on the axis of your cost which is where you have your hot tier, warm tier and cold tier. ⁓ Before we move forward, any questions on this? Payments and logistics data which we are putting in dumper for later merging ⁓

Payments and logistics. So in the hot DP or they have their own hot DPS. They have their own basic payments have their own database. Right. ⁓ So they also have the same problem, right? You need to move the data from hot to warm. Exactly same. So same dumper like you will have different pipelines to read from the database and put it into stating storage. From there you merge it and you put it into warm store by pipelines. You mean ⁓ spark jobs, spark jobs. Okay.

So they pull the data, but they have the policies over there. ⁓ And also policy is like policy could change. Right. I mean, let's say policy change. ⁓ Then I'm you, you, you have some metrics and you figure out of the six months is still very, ⁓ mean, not Amazon, but you are in an early stage and you see ⁓ six months is this lot much view could go to three months. So that means you also need, we need to make a cashier, right? So that if any change comes up in the ⁓ policy cash, you have a policy idea over there.

such that if any change comes into the policy, the services should actively fetch the new policy and then look up, do some processing based on that. ⁓ if you are that frequently changing policy because changing policy is not so frequent, right? So you can do a redeployment in that case to take your policy. ⁓ Because it's not, it's not life, right? It would happen once, ⁓ once a year or something. ⁓ And you would not change it that frequently. ⁓

You can just do a redeployment of your stuff. So can I take it in a general formula that whenever a conflict change happens infrequently, basically, ⁓ you do a restart of the box. Yeah, always don't over engineer. I mean, if something is changing, let's say once a month, why would you want to take care of doing it in court and put real time? need. ⁓ for ⁓ that. ⁓

Sneha Mehra (00:38:17)  
So I'm going to voice this too low. Now it's better. ⁓ So I have a very quick query. Like can we consider staging storage as a cold tier? Staging storage as cold tier. Yeah, you can consider that. So long as you are able to query on the data. There is no hard and fast rule. This is how it should be. You can also use this staging storage as this and like directly read Hoody files over here if you are okay doing that. ⁓

You can do that. And then it's your job. which is what I said, there is no hard and fast rule to anything in tech. It's all up to you, your implementation. ⁓ But yeah, you can do that. Yeah. that one point is not clear to me that in the very last ⁓ bottom that, yeah, what is this temporary storage? So for example, if query is coming to historical data, ⁓

So, okay. So for example, as I gave you an example of, let's say, there's a police complaint file against a person. You are not making one of queries to S3. You are making a lot of bulk queries. Now we are investigating that. Right? So what you do, you would download all the records of that user from S3 and put it into a database so that because firing a lot of queries directly to S3 is very costly, very costly. Okay. Right. So what you would want to do is you want to just dump all the data for that user.

loaded into a template database and then query on that directly. ⁓ So that you save the cost of making pointed queries on ST. Okay. Yeah. It's clear now. Thanks. Yeah. ⁓

Sneha Mehra (00:39:58)  
Yes. Yeah. So I am a bit, ⁓ I'm just about, why do we really need the staging storage as in payments data would be also related to some other table. ⁓ like all of those would come over here. So then what is OLAP exactly? Like it's used for analytical ⁓ things. Use a specific example. are ⁓ you talking about big queries, like state shift or a data warehouse?

Yeah, I mean, anything useful, analytical there do we like store it just combining everything? Yeah, but that is variable. ⁓ S3 is not variable. Correct? What you're talking about OLAB BigQuery Redshift, that's this. Correct? Where you are actively allowed to query the data in a relatively faster manner. Correct? Yeah. Right. Will you be able to query with that frequency on S3? No, S3 does not support that.

because it's not meant to serve that use case. So this is what your BigQuery Redshift is all about. You're still doing a lot of reads on that, which is costly, which is cheaper than your transactional, but all the data from all the places come into this, your data warehouse, which is our OLAP. And then you're querying on it. Whoever wants to query on it can actually query on it. So do we have to keep it joined? It's up to you. In most cases, yes.

In most cases, you have to pre-join the data and store it because otherwise your queries on this, if you're doing the same kind of joints over here, that would be this very expensive. You would require massive compute to do that. So you typically do pre-join pre-aggregation of data and store it in a very consumable format. Okay. So it's all depends, right? Basically, again, as I said, there is no hard and fast rule over here. Whatever keeps your cost down, you do that. Whatever

fits your use case, you do that. Okay. Thank ⁓ you. ⁓ have a two question. ⁓ first ⁓ on the crown side. So as soon as we let's say pass the six month ⁓ of our work, ⁓ you will see that orders are there always. And for one of the user or another user, the time is passed and every second you have to remove something.

Sneha Mehra (00:42:23)  
Is there other way other than cron where we can achieve this thing in a bit more efficient way? No, so cron ⁓ implies anything which is repetitive. It does not mean a simple cron job. So you can also do this that once your data is taken out by this dumper and is loaded into warm storage, this loader deletes the data from hot tier. This cron only signifies that there is something repetitive which is happening, which is pulling data over here. As simple as that.

It does not mean that the data is automatically getting deleted from hot tier because you need to guarantee that your data is there in the warm tier before you delete it from hot tier. Correct? So what you can also do is you can let your loader delete the data from hot tier. Yeah, I'm thinking ⁓ a load on a DB that ⁓ all the time load on the DB. You can do it once a day. Like you can, you can bulk the, you can batch the deletion and run it once a day on during off peak hours.

Okay. So I mean, ⁓ best strategy looks like in place of doing this thing continuously and putting a load on a hot, you can take like on a day and or whatever time you try to get all the data, which are six months old or something and it's a bad job. good. Better. ⁓ Maybe, maybe not. Okay. That's when I say it's more about having some process that is doing it. ⁓ Okay. ⁓

Maybe then second question is towards this order data. We need it from two purpose now in future. One is the transactional form where we want to know, ⁓ for this user's data around seven months ago or something where you know the order and you need the order ID and then based on that you patch it. ⁓ Another is for all the analytical purpose. ⁓ need, let's say one to know ⁓

our analytics team wants to know the all orders of from users which has greater value than X. ⁓ On those cases, we need that columnar form of the data where you do not try to store on a ⁓ row based but in a formula format because the queries are suited ⁓ at kind of. So in this picture where I should kind of ⁓ plug the ⁓ analytical DB, what I'm thinking

Sneha Mehra (00:44:45)  
Plugging with here in the worm, what you have mentioned, ⁓ looks ⁓ not so good. I'm not, ⁓ I'm not feeling comfortable on this is ⁓ the purpose of analytics is always you, as soon as you get the data, okay. It is in hot. Okay. And you also need to do analytics on hot also. Warm reads based on whatever your cron or bad job is those data. So if that is a risk case, then

Your rights can directly go to our analytics DB. You don't have to wait for the double to cupcake in key only six months old data. Now what you are doing is you are, you are adding a new constraint. You're adding a new requirement that you want to do deep analytics on a real time data. And which is what I have not drawn that because it's what not part of our ⁓ specification. Now what you're specified at this moment ⁓ is by hot data. I'm not able to do analytics on my hot data. Correct. Now what you can do

is your orders when it writes to hot data. You have another thing which is dumping it into let's say your Amazon redshift and whatnot like this right there and then Kafka, Kinesis pick your favorite tool, CDCs and whatnot. ⁓ All the data goes into this hot tier analytics when you are doing which is kind of your Amazon redshift, Google, BigQuery and all. And if you want real time analytics then you build that pipeline that in real time flushes that it does not wait for six months time period. What I was describing

was at six months duration ⁓ while a put up flow. And if you want to do that, that, that unified analytics ⁓ on that also includes your hot data. Then you have to immediately push the data onto this site ⁓ or to your data warehouse like Amazon Redshift and it's basically Amazon Redshift and BigQuery. Okay. So got it. I mean, the solution is to use a separate path from order from order service to put it in ⁓ a well. Yeah. ⁓ And

⁓ Maybe just a follow up on that. From the perspective of load ⁓ or scalability, ⁓ what should be the pattern? Should I put it in a sync ⁓ way? because it's not immediately needed. ⁓ Okay. Yeah. The order which are coming. Okay. Writing in a transactional DB is a ⁓ sync job. will say, but orders writing on ⁓ a

Sneha Mehra (00:47:13)  
⁓ OLAP is a kind of a... Okay. ⁓ So, is it ⁓ advisable ⁓ that from orders it goes to, let's say, Kafka ⁓ and from there we write in a batch because writing on OLAP is a costly affair. ⁓ CDCs come in handy for that. ⁓ Okay. So that's what my question is. We should use CDC or... We should use CDC which is plugged onto this database and takes the data, transforms it and puts it wherever you want.

Because it is, this is what kicking immediately, right? Whenever there are changes, it is constantly reading from the database at this moment. ⁓ Right? So you can use CDCs for that. ⁓ Don't plug it in from here to Kafka and all. ⁓ Whenever it's about data, you directly plug CDCs and read it and then store it wherever you want. ⁓ So even this, even this can be a CDC. ⁓ Yeah, actually my, my thought started from that. Can it be CDC then? ⁓ Yeah, yeah, yeah. Whenever it's data.

They don't really from the database when you're dealing only with data because it does not require application context at all. Right. Yes. So you did it really from CDC rather than relying on some Kafka event coming in from the API side because not everything would have a Kafka event. Right. That is where CDC comes in and you read and you put it wherever you want. So CDC can have two different destinations. cannot like once CDC receives the data, can go and write in two different one on the dumper side for this and another for analytics. ⁓ Both are different.

But that solves the purpose with very clean architecture. You have a TDC running on the hot and it stops both of your cold, warm storage and also the analytics parts. Okay. Thank you. Thanks for it. And this is where CDC shines, right? Which is where you should not rely on EPS server to push it because you are dealing with raw data. You want data to be there no matter what table you add, no matter what kind of updates you make, you are not expecting to have Kafka events for everything. So CDC is the best place from which you can pull the data out.

from the database and do whatever you want to do with it. Yeah, I got this idea yesterday when I was reading a db log ⁓ from Netflix. They created a CDC, not open source as of now. In 19, a paper came out. ⁓ They're trying to do the same thing because they have all these kinds of things. Okay, super. Thanks. Thanks for that. I'll take one more question. Go ahead, Mohan. Yeah, so ⁓ from the relational databases typically, right? So the transactional databases, we are creating the merging

Sneha Mehra (00:49:41)  
⁓ Is there any loss of information if you want to reconstruct the original picture? ⁓ No, why do you think there would be a loss? I'll address that.

Because we are creating ⁓ like a from relational store, right? We are flattening out the structure, most likely, because we want to ⁓ put JSONs and things like that, right? ⁓ So obviously, when you are flattening other structure, so there might be a case where, know, maybe, you know, I was just wondering if do you kind of capture every each and every attribute of ⁓ literally every single column of every single table?

of this database is stored in the staging area, literally everything. And from there it is merged. what happens is literally imagine end tables having ⁓ columns, all of them as is stored on this S3 being read by loader merged on real time. And this is where that JSON object is created in the memory of this loader and written over here. So you would not have any loss of data at all that entire table with that entire structure.

is stored over here. Okay. In case you want to construct this, you can do that. So there will not be any loss of it. And you should not design it so that there is a loss of information anyway. So here you literally have every single column copied. Okay. And second thing, is there any flattening out ⁓ in the sense that for example, orders and items, right? ⁓ If you imagine if you

flatten out in an Excel, right? The same order ID will be repeated across, right? Something like that. So does something like that happen or because of JSON, it is ⁓ efficient structure. with JSON, you don't need it, but in case you want to flatten it out still, you can do it over here. For example, you have a table like rather here also, I'm just drawing it like one S3. Here also you have multiple stages of data. Like you have your raw data, then you have processed data, then your semi-processed data, then you have optimized data and whatnot.

Sneha Mehra (00:51:49)  
Right. It goes deeper into this. It's literally how you form data links. But the idea is that have a place where you are dumping your tables and all data as is ⁓ one place on S3. Then if you'd want to flatten it out, if you'd want to make it ⁓ denormalized, you can do it. Right. So there is no rubb. But have one place where your actual raw data is stored from read that and create a flattened data. That flattened data you can read and create JSON out of.

So whatever you're proposing in Worm is typically a structure like that or it based on... in Worm I'm proposing to store everything that you need to be served to an end user, which is queryable as is. Like given an order, I want all the information for that order in one JSON. So it's literally structured for me to just like directly send to the end user. Okay, understood. So I might want to then add additional formats. Correct. ⁓

⁓ But you see how this architecture makes it pluggable. Where you can just, if you want different types of use cases, you just have to change the staging, right? Pipelines that are just writes back to the staging from which you can consume and fork other things out. That's all that other use case well. ⁓ But the overflow would still remain the same. you. Great discussion. Okay. Now ⁓ addressing. ⁓

⁓ Before we start designing it, I think we'll take break before we even start designing it. I was thinking that this would wrap the 30 minutes into dot. So we'll take break in five minutes. But before we brainstorm on S3, which would be a massive discussion before you bring someone S3, I just want to leave you all with thoughts ⁓ on ⁓ things that we discussed last week because they will all be, they will all be coming in handy when we design S3. So last week we built word dictionary on S3. Word dictionary that we designed that one

format with index and data. It's kind of like an embedded database. We stored on S3, which was dirt cheap, but it had very costly access. It can be used as cold storage with infrequent access. just designed an order system leveraging that. So when ⁓ a ⁓ tool, remember this, whenever a tool says, we give you an ability to query directly on S3, like AWS Athena. Athena gives you this tool.

Sneha Mehra (00:54:14)  
Presto gives you those abilities hive queries. write hive queries to do that. We had a single with this powered by spark. ⁓ All of that. Then we do data is structured on S3 has to be such that it has that information to make that pointed query somehow. Right. The big data engineering tools like Apache hoodie format ⁓ or spark SQL and what not. Won't go into those details, but ⁓ all of that create data when they write on S3, they create it so that they have that meta information.

And then so that they can make pointed queries on it. ⁓ okay. Another thing that we studied last week ⁓ is BitCask. In BitCask we saw how lock ⁓ structured storages are important, how they give you performance. ⁓ And we understood where it comes in handy. It gives you very good performance on cheap commodity hardware, like magnetic disk storage. ⁓ Kafka or any similar tool that you pick, whichever this say that they are

high throughput system behind the scenes are powered by lockstructured storage. So if you go to Kafka internals, when Kafka accepts the right, the right, the message that is there is literally written in one append only file. There is no magic behind it. It reads it and it writes an append only file. It reads another right gets it writes an append only file. So for every single partition in a particular topic in Kafka, simply assume this like oversimplification, but you'll get an idea. Every topic

has any partitions. Each partition is one file on the disk. When a read or when a write comes in, it knows that this for this topic, for this topic, for this partition, it goes to the corresponding file and appends it at the end of it. When reader reads it, it knows that from this file, I want to read from this offset. The byte offset of this is maintained ⁓ in Kafka. That is, this is the consumer which was reading from this partition and has read till this offset. So you can literally

build a micro version of Kafka in a matter of one day using lock structured files, simple up and only fights. And I would highly encourage you to do that. Like this is how I built my understanding of Kafka really well in the internals of it before I started diving into the source code, because this is what it's all about. So what I would recommend ⁓ is take this as an exercise, ⁓ it out. For every, like just create one topic.

Sneha Mehra (00:56:38)  
One topic and partitions create and files open all of them in a part on the board. When you get the right pick one file, write the data there, write a reader that is reading from this files step by step because in Kafka you read sequentially one after another from a partition. This is where you're doing something very similar. Don't think of making a distributed. Don't go into those complex people at least build a small prototype. You'll get very good idea on how Kafka works internally. And you, when you build it, you get that joy. Okay. Then

where you use lockstructured storage now because it gives you very high throughput ⁓ at ⁓ on commodity cheap hardware. So think of lockstructured storage wherever you are getting very high ingestion rate really high ingestion rate and you are doing write heavy workloads not read heavy write heavy workloads where your system is having write heavy workloads. You want to accept a lot of writes coming into your system and you want to write it as fast as possible.

which is where lockstructure storage shines. Few examples, metrics, analytics, clickstream data, any system you pick New Relic, you pick Datadog, you pick Google Analytics, ⁓ all of them are backed by lockstructure storage. Not as a primary storage, but the place which accepts it. Most of them, say that we write to Kafka. Kafka itself is lockstructure storage because it can support very high write throughput. ⁓

So this is the magic sauce behind any system that claims to be a high write throughput system. It's always backed by a log structured storage. Kafka is an example for that. So when you say I'll use Kafka over here, you're effectively using a log structure storage, which gives you an ability to write on a cheap commodity hardware at a very high performance, which is why BitCast is so important. Okay. This is what a glimpse ⁓ of the previous things.

All of this will come in handy when we design S3. That's why I just gave you a refresher for that. Okay. So now what we do, we take a break because now we have to get up for a very huge discussion for one and a half hour or almost two hours. We'll talk about S3. Right. We'll talk about S3, we'll design S3. let you all folks have some downtime. ⁓ We'll sync back in at 10.05, six minutes from now. At 10.05 we'll resume a very deep brainstorming on S3 and very interesting points to cover. First of all,

Sneha Mehra (00:58:59)  
This is where you'll see why consistent hashing is not the magical solution. Consistent hashing fails while scaling. No one uses consistent hashing. We'll talk about it. Why? Right. Then we'll talk about storage, how to make elastic storage. Then we talk about handling hot partition problem and access and lot of other granular details of it. Right. Every single decision will be very realistic. The baby like the baby talk about it like, okay. So we'll see back at 10 or five, six minutes from now. And then we.

design S3 with that of Brainstorm. ⁓ See you folks in 6 minutes.

Sneha Mehra (01:04:53)  
Let's start with designing S3. discussion. ⁓ So ⁓ S3 ⁓ is a key value.

So yeah, so what happens on S3 is your key ⁓ is the file path that you have defined. The value is the file content that you're storing on S3. So S3 is actually, and I'm not even making it simpler. It is actually just a key value store because the notion ⁓ of ⁓ directories on S3 is logical. Like on real world, when you

create a directories on your file system, does not mean that on S3 when you're creating it is getting created that way. It is at that path in the file system. The concept of directories ⁓ and folders on S3 is actually logical and virtual. For them, it's simple. This is the path. This is the file. That's it. Right? So S3 is a gigantic

key value store where values can be massive files, terabytes big files. ⁓ So the bitches I like you would have seen that throughout this course, we designed so many key values stores because key value stores are actually the foundational building block for so many systems. ⁓ So ⁓ S3 is just key value store or values can be gigantic files that needs to be stored on this. Right. So you cannot store it in memory. They are gigantic with one terabyte is also a common

⁓ like a common length of ⁓ first. So now which means that what we have to do ⁓ is we have to think of the storage side of things that hey, if we are just building a key value store, then all this, remember this, right? It's just a key value store that given this file path, I want to write this content. This is, let's say, let's say you, let's say I just make it simple. I say that, hey, in an HTTP request,

Sneha Mehra (01:07:05)  
⁓ In an HTTP request what I am getting is I am getting that hey put this this file which is the content of the file at this particular path. ⁓ This request came to our API server and now our API server needs to write it to storage. ⁓ Obviously you cannot use a transactional DB like MySQL to do it because there is huge amount of data you are writing GBs and TVs of data.

Right. So now let's talk about how do we design or how do we define storage layer for S3. Now the requirement side of things. The storage of S3 has to be elastic.

which means it should grow when you need it. Right? One key thing when you write a file to S3, you always pre tell S3 that this is the size of the file. You would never say, ⁓ this is a stream of bytes and keep on appending to a file. You always say that the file that I'm writing has length of let's say one GB and then you write

into that file. You never do this that you say I'll just send keep sending you bytes and keep appending to file whatever, however big it can get. You never do that. You always pretell S3 that this is the size of the file that I'm writing and now comes the bytes of that file. Right. So now what you need to do is you need to or rather we need to design a storage layer for S3 which is elastic in nature, which can grow as we want it to.

because S3 is an ocean of data. So how would you design the storage side, literal raw storage side of S3? The access is key value based, directories are logical. They don't hold any significance there and it needs to be like it needs to grow on its own. How would you go about it? Raise your hands. I'll pull you in.

Sneha Mehra (01:09:13)  
Apply first principle thinking. Don't think in a very complicated way. Keep it very simple. And typically the simpler you think, the better are the chances that that's how it would be implemented. ⁓ Rahul, go ahead. So basically one way I can think of this as is we will just provide a pointer as a value and allocate memory to that particular block of like 1 GB basically.

But what's the pointer that it's low disk, right? Pointer disk. Yeah, this Yeah, this like some sort of like ⁓ values like disk, ⁓ disk identification number and the address on it. So saying that this is your disk. This is disk one, right? On which you say, okay, please reserve one GB file for me. Let's say this is where it results. So it means it has some method. So some file path on the disk.

where you are storing this and this file path can be separate from the other file path that you have. Okay. So some file path you have where you are storing this on the disk may be similar to that some ID and ID and what not you added. No, ⁓ like there is also one more thing that can be done is like ⁓ I split these bytes ⁓ into some partition. need, no need. Let's keep it simple. And S3 also does not chunk everything.

Don't over complicate. Keep it simple. That is performance optimization. Now, first, let's decide the basic one. ⁓ Then we'll go into advanced. And also like S3 has ⁓ some amount of copies of the data that we send. That is different. Let's just think of one. Let's solve one problem. Now, those are additional features. Let's solve this one problem. Is your storage because now this disk would have some capacity, let's say 100 terabyte disk. What happens when this fills out? We need to think of that. ⁓

We are still not done with simple storage. Then we'll go into durability and other, but I have all of that covered, right? But let's first solve this problem. ⁓ basically, ⁓ okay. So basically it would be like this guy identification dash the address or something like that. Okay. Let's say I have one, ⁓ 1000 GB of storage. So now you'll store this part. Now what happens when you get more rights, you would fill out on this disc. Now what? Yeah. So

Sneha Mehra (01:11:39)  
That's why I said disk identification and doesn't disk one and this is the address on which you ⁓ need to read to get this file. And like ⁓ and when this disk fills out, we will create ⁓ another disk with the new identification ⁓ and same thing ⁓ pointers. ⁓ Okay, fine. Thanks. for that. Let me pull in Heman. let's build on top of this. Now what would happen? Let's go into those micro decisions.

when you're storing files on this disk. ⁓ Yeah. ⁓ So maybe the first thing what Mural says is we need index which identify the disk and the location of that. ⁓ Now, ⁓ if we go and start understanding that we have to allocate disk space, ⁓ we need to worry about how much we can allocate and how if we say like, I have a thousand GB.

and I'm looking for up to a 1 GB file, then I can have a pointers only for 1000 specific key value pairs. ⁓ It seems like a overkill to me. We, so not every file will be 1 GB, right? It depends, ⁓ right? initially when we are saying we need to have, you cannot start with the assumption because people can write small files also. Yeah. That's why the problem is we have to ⁓ figure out a solution, how we allocate a size.

for a given file. Correct. Correct. So there we may need to ⁓ think about the chunking as you said in between that no need, no need to keep it very simple. Think of it because chunking you don't need chunking at all. To be honest, you don't need checking at all. Now ⁓ on this disk, what? Okay, let's start with this. What is your key design decision when you define this disk? What type of disk are you using and why? Okay, it

Definitely, ⁓ whatever we have learned from previous systems, we are saying that these are the key value pairs where you are going to put it something for ⁓ as a value ⁓ and the value if it changes, we can kind of provide a separate pointer. So, these should be kind of a write only which can be a sequential. Like, okay, you start writing from start to a point and then you do not need to modify. So, we can use ⁓

Sneha Mehra (01:14:04)  
⁓ something which is a continuous storage where you go once and append only. So whole thousand GB disk starts from sector one block one and ⁓ you start adding one by one each of the pointers. You started finished. What you describing is a block structured file system. Yeah it is saying I started from zero zero location of the disk. I reached to 10 then I

From then I will give another pointer and let's say that goes to 35 then I will ⁓ got it. But why did you go for lockstruct file system? Okay, but but but basically before that, let me brief for brief folks about it. Okay, but even before that, what kind of disk are you using?

What kind of disc are you using? ⁓ Are you talking about SSD? Yeah. Yeah. Yeah. Definitely. Here. What it makes sense is like a, desk where I can go and write once, ⁓ easily by seeking to a given location. ⁓ So that, that is good. think for SGD also, ⁓ what, what, what the problem with SGD is to seek a location and then write it. ⁓ fair SSD science because you do not need to do that.

So here I think HDD is good enough because the way we are designing the system. ⁓ But what other factor will you consider in picking HDDs and SSDs? Magnetic disk versus SSDs? ⁓ Read and write throughput of the system of that disk that how big files. ⁓ So let's say I have support like a 1 GB file. If my file format system is not supporting, then it is not making sense. ⁓ That plus what kind of a cost it gives.

on me. Everything else is secondary ⁓ because imagine on S3, the amount of data that we are storing on S3 is so huge, ⁓ so huge that if we store it on SSDs, that would shoot up the cost like anything. ⁓ And the amount of data, ⁓ like it has multiple exabytes of data, which is ⁓ insane. I think it's seven or eight exabytes of data at this moment.

Sneha Mehra (01:16:18)  
It's huge amount of data that S3 is storing at this moment. Given that, now what would happen is if you store that much of data on SSDs, again, this is like hot tier and cold tier, kind of that. If you're on SSDs, you get very fast credits, very fast rates, but your cost would shoot up. Now, which means that you would want to have a very nice margin where it becomes a profitable business. You need to go with a cheap commodity hardware.

So which means the first thing that you fix when you are designing an S3 ⁓ is that you need to go with magnetic disk storage. Now you start with that because that is cheap and commodity hardware you'll get better margins. Business comes first engineering, engineering everything second performance, performance will try to get the max out of the constraints that we have. So we start with this concern because we are building an ocean of data. We start with HDD. Now as soon as we start with HDD ⁓

Now starts to kick in the point Hemant you said about performance that HDDs are not performant in doing random writes. So Jyothinder what would we do? If random writes we cannot do random writes because HDDs are bad rated because of magnetic disc movement. What's the solution to that? To get similar performance as SSDs? Like we just do append only writes. Right? So that's where we want to do append only writes. But it's not

Like we talk about append only we typically talk about within a file append only here then that file system is append only that whenever you write it will always write to the next sector. It was never that your first right would go over here then right over here then over here then over here then over here. It would never be like this. The entire file system is lock structured storage, which means that whenever you write it would always write from the first sector to the nth sector.

and would fill out data like this ⁓ no matter what. some caveat ⁓ the file systems that we typically use FAT, NTFS, EXT, BTRFS any when your operating system the file system that you have they have the core responsibility of the file system is to figure out that when someone asks for a particular file path where on the disk will I allocate that file.

Sneha Mehra (01:18:41)  
And it's the implementation of file system that dictates that this is where you need to store it. This is where you need to store it. This is where I can store it and then defragmentation and whatnot. ⁓ Right. All of that comes in. So that is what your file system does it for you. So on a normal file system, ⁓ the things that you have is that this is the file path. This is the I node. ⁓ I node is a very common concept which tells you

that ⁓ basically path is part of inode, but you'll get the idea that I'm just making it, ⁓ I'm just simplifying it for everyone to understand. So let's say if I have a path slash home slash at the slash one dot txt, right? If this is the part on this part, I'll have an inode like, so my inode would store this information that where exactly on my desk is this file stored and how big is the file?

So it would know that on this disk, this file is stored from this location to this location. So that information will be part of my inode. ⁓ Now this is what the key value that we are talking We don't have to explicitly maintain this. Your file system implicitly does it for us. You define a path that at this path I'm writing it. Your file system itself. So when you browse through your local file system, your Mac, your Windows laptop, you open C drive, you double click on that folder. How does it know?

that these are the files within that it needs to store that somewhere, which is where your file system comes in the file system stores that information in a directory, which is what is called as directory in which there are multiple I nodes entries that I nodes contain the meta information about the file like updated at body ⁓ created at modified at file size file format and other information like who created a file who can access this file all of that information. So when you run

LS minus L the output that you get is typically the meta information of the file that is stored in the inode ⁓ and the way it is getting that information is making a file system API call getting that information and rendering it. This is what happens. Now. This is what the key value mapping that we're talking about is stored implicitly file the file system. We don't have to do anything that you just have to in this file system in this hard disk just create this hard disk in the

Sneha Mehra (01:21:07)  
just use any lock structured file system and you start writing file at a particular part. It would internally maintain this mapping for you. You don't have to do anything for that. Right. But when you're doing it because it is locked, you need to tell them that, this is the size of the file, which I'm writing. This is the next file. This is the next file, whatever you're getting. Right. And then it would start writing on those locations, but it will always move as Haman said, always move from sector zero sector one sector to sector three sector four sector five. So on and so forth.

in a sequential fashion, no matter what. ⁓ This is a very important design decision that we are making. Again, we don't have to maintain this mapping. It is done by your file system implicitly. You just need to know which hard disk your file resides in. That's it. And then you make a call that give me file from this part, it would give it. Okay. So this is what you have at this moment. ⁓ now, Jyotinder, my question to you is what happens

When this fills up, what would you do?

⁓ you need to start sending write requests to another like disk or data node. So another disk you are adding it. Okay. But now how would you ⁓ know or rather ⁓ let's say ⁓ let's say even this filled up now what

Sneha Mehra (01:22:29)  
You add one more. Yeah. ⁓ And then this filled up. You add one more. Correct? Yes. Now this is a very common way to do it. But then what kind of and now what you're doing is where are you writing? So you need to have a way to know which is your current hard disk where you're writing. Correct? Right. That becomes let's say that is nothing but a

head pointer kind of linked list. If you look carefully, this is actually being formed as a linked list where we have a head pointer that tells you which is the active disk where you have to write. Which is what we do in S3 or in any blob storage. Everyone knows that this actually now let me go to a bit more detail. Now this actually is nothing but a storage rack.

The storage rack that you have, this is exactly what a storage rack looks like. A storage rack has a huge number of disks, which has in 100 to 200, not that much, 100 to 200 hard disks. They're attached, they're plugged into the storage rack. ⁓ And when you are writing to that storage rack, you know that this is the active hard disk in the storage rack where I would write it. And it writes from there. Once a hard disk gets filled up,

this pointer moves below and now the other hard disk becomes an active hard disk and this can be maintained by a proprietary storage rack software who maintains that this storage rack and this is where hardware level changes embedded system not embedded system but basically hardware level device driver changes comes in so most in most cases storage racks comes with a device driver that maintains this pointer for you

on this is the active hard disk at this moment. If your storage rack knows that this is going to be a gigantic lock structured file system. ⁓ Now going bit more, I've seen little abstract, but this is how it is. And just to give, you want to know bit more details about it, just go to Dell website. You'll find so Dell has amazing set Dell ⁓ verbatim.

Sneha Mehra (01:24:57)  
So Dell and Verbatim are market leaders and even WD like how can I forget WD. Dell, Verbatim, Western Digital, these are market leaders in storage space. They give you or they sell storage racks like this. So when Amazon, Google and all of these companies, they build it, they either they choose to like now they've started manufacturing it, but they typically get this cheap hardware from these companies and they put it there. ⁓ So now what happens is the

Storage racks, the device that keeps this information. It's now pluggable over there. Which is the active one, it starts to forward to that once the hard disk will back up and then it re-shockballs and what not. So you can directly interface with this storage ⁓ rack ⁓ driver and ask it to do whatever you want to do. So it abstracts out the complexities of this hard disk. ⁓

The storage rack that we're talking about, the storage rack having multiple hard disks. This is one of the configuration where your storage rack knows that the underlying hard disks that you have is lock structured format kind of connected in a linked list. It's all maintained by the storage rack itself. So it's typically masked out from the implementation data. Like you don't really have to worry about it, but it takes care of that. But I'm going bit more details on how it actually functions. ⁓ Right. So

In a, a, I got one place. ⁓ So during my B tech, ⁓ I used to manage, ⁓ I was part of the managing team, which used to manage data centers for my college. ⁓ So some of the professor loved me. So he always used to invite me to data centers, ⁓ which is where I first saw a storage rack kept in a very cool temperature, roughly 12 or 10 degree AC room, ⁓ bunch of storage racks.

where every internet request made from entire college used to go through that. It was the proxy. It used to go through that. It used to make request on the internet, get the response, store the webpage on the storage rack and send it back. And so literally every single webpage, ⁓ every single student or teacher ⁓ access in last 30 days were stored in this gigantic storage space. ⁓

Sneha Mehra (01:27:24)  
We got to know about it. So it started all the funny activities that we used to do because this is very risky. Right. But whenever you are behind a proxy, remember anyone can do anything with it. So you need to know which network you are part of, which is when I got exposure to this thing, like how the storage ⁓ rack works, what kind of things they does, what kind of interfaces it provides. So one of the responsibilities that I had at that time ⁓ is if one of the storage rack gets filled,

I because it's just temporary data. I was allowed to delete that which is why I got to know that how to interface with the storage rack and how to ask you to delete something if I if we want here on S3 we don't delete anything but I'm just giving you a glimpse that if you'd want to know deeper details of it just go through Delverbatim Western Digital pick your favorite go through their product specification of storage racks. You'll find all of this information.

will be a bit more technical, ⁓ but go through that. It's very fun exploring those parts. ⁓ But here you see how beautifully this fits together as an purely elastic storage. ⁓ Now, this is one storage rack that we were talking about. ⁓ Rides are coming in from your users, going into the storage rack. ⁓ One storage, ⁓ one hard disk builds up, it moves to the next hard disk, it moves to the next hard disk, it moves to the next hard disk. Now I can add as many disks as I want.

But storage rack has a limit because storage rack is a physical hardware in which you can, you are inserting ⁓ the disks there, right? Because there are ⁓ limited number of slots, let's say 20 or 30 slots, you cannot go beyond that. So which means that you would have multiple such storage racks.

And then you would do one storage rack is done. I'll move to another, then it's sort of like this and I'll move to another, then another storage rack is done and move to another and so on and so forth. All right, so this is what you do. So I simplified it by explaining one storage rack, but you would have multiple storage rack in your data center storing all of this information. ⁓ Okay, now how big one hard disk would be? ⁓ Let's start with that. ⁓ Jyotir will continue with that and then I'll pull it out of the book. How big one hard disk would be?

Sneha Mehra (01:29:47)  
⁓ Like it depends, right? ⁓ It can be as big as possible. ⁓ can be 100\. ⁓ Guess ⁓ a number. ⁓ 100 terabytes. 100 terabytes. 100 terabytes is very costly. ⁓ So you chop it down. 10 to 20 each hard disk is 10 to 20 terabytes big. And in one storage rack, you will have 20 to 30 nodes like this. ⁓

⁓ So the amount of data that one storage rack can handle is

200 to 600 terabytes. Huge enough space. Right? This is what one terabyte would have. Sorry, this is what one storage rack would handle. Right? Okay. Now, one small caveat over here. We think, it's just 600 terabytes. It's not huge. ⁓ I... ⁓ A lot of us would think that because we are so accustomed to hearing 1 TB

100 GB, 200 GB, it's nothing. But for a business workload, for a corporate workload, for any mid-size organization, even big companies, the kind of things that we are storing on S3 is always some kind of text data, either dump of our database, either ⁓ some ⁓ files that we are temporarily creating.

and some processing that we did in most cases it is textual data that we are storing typically dumps of our databases in most cases apart from that we also use it to store media objects ⁓ but in most cases you would see that 100 terabyte is a huge space for almost any and every organization there are very few who go beyond that for most people for most organization out there this is big enough storage space

Sneha Mehra (01:31:44)  
for their lifetime. ⁓ Don't go into extremes. Extremes are like you cannot design a system considering anomalies are normal behavior. ⁓ You have to design systems for your normal customers, extreme customers separate, like handle them separately. ⁓ So for your regular customer, this 600 terabytes of storage space is good enough. What else do you need? So what you can do as a bare minimum,

Nice solution. Give one storage rack to every customer that you have. ⁓ So let's say I have company ABC, give them one entire storage rack. This is your storage rack, do whatever you want to do with it. ⁓ And you can just abstract like this is how you allocate your resources in S3 for them. No matter how many buckets they create, no matter how many files they create, it's their sort of like whatever they would want to source. This way you get isolation ⁓ of the load.

across your customers. But is that efficient enough? No, because what if when you assign a storage rack to a company, they just don't have one GB of data and then they never do anything. You are wasting so much of space. So which is why it is not a good utilization of your underlying hardware. So now we have to make it better. So now let's talk about how do you make your

Hardware utilization better. So up until now what we discussed is we discussed around Whenever write request comes in we store it in series of hard disks connected kind of like linked list abstracted by the storage ⁓ basically abstracted by the device driver of the storage rack which keeps track of this is the head of my storage rack, which is I'm writing

You can also choose to implement it on your API side if you want to, but you are keeping track. Okay, this is the head of the storage rack. Either way it's possible. ⁓ And when one disk fills up, you add to another one and then to another one, then to another one. Right. And everything is stored in a log-structured format. Right. Okay. Now my next question is, given this is the structure, right. Now, how would you know

Sneha Mehra (01:34:06)  
and pull it and above and above. How would you know that a disk is full? And now you have to go to the next one. Someone has to tell it. Now the disk is full. Go to the next one. Who would do that?

I think the file system driver or not the file system. Yeah, the file system driver would have to do it. So the file system driver is ⁓ is limited to this. Right. ⁓ maybe storage rack. ⁓ Rack is one, right? Because file system driver is limited to one hard disk. So the storage rack driver can do it. If that does not give you. So then what you do, you add your own small component whose job is to monitor it.

So you add a monitor.

who is monitoring this and switching the head if you want to. ⁓ Right? So you may want to have now, when would you switch it from one to another? What do you mean when you say it is filled? Let's say I have one, let's say I have hard disk of 1000 GB. When do I say move to the next one? ⁓ When I fill entire 1000 GB, ⁓ then I say,

No, I don't think so. In that case, like we would be impacting throughput for rights because it gets filled and then some of the rights would be in pending states. So we should do it a bit before. ⁓ Brilliant point. ⁓ Because now what would happen, the way you put it was very nice that now imagine now this is where you would say, Hey, let's start with Zippad. When I complete my 1000 GB, then I start writing to the next one.

Sneha Mehra (01:35:52)  
Now when you do this, let's say just completed 1000 GB and now what you are doing is you changing the head. Now the right should come over here. So while this is happening, you cannot accept any other rights into the system because there is no space to fill over here. Now this is yet to take effect. So this is not a good, good idea. So which is what you do is you define a threshold typically 70 % that when your hard disk

is filled 70 % you move to the next one. Right? You change the head you move to the next one. So you give yourself enough headroom to complete the existing rights that are coming in while the other hard disk is becoming the new head of the linked list. You are still accepting the rights happening over here. Once this entire right is done, then the newer set of rights can come to the second hardest. This way you are not impacting your any existing rights.

And that is...

really important for when you are building a system that is six times or seven nine availability they give 99.99997 nine availability that provide right and you're giving that much of availability you need to enter no matter what how many rights that are coming in you are able to serve them so you cannot wait until the last moment for you to do it for you to push or you want to write start adding to the next hard disk you would always need to do

that when it is 70 % I'll start writing to it. So any existing rights that are happening can still happen. You have enough room available for you to complete existing rights. And then the newer rights can anyway start happening over here. ⁓ So this way you do it. Okay. Now on S3, you can also delete a file. So when you delete a file, you have to remove the entry and then do a cleanup. So this is a classic case ⁓ of merge.

Sneha Mehra (01:37:51)  
and compaction.

Sneha Mehra (01:37:56)  
So now what we do is because this hard disk are just 70 % filled and there will be some entries that would be deleted like that does not hold any stance at all. What we can do is we can ⁓ merge the hard disk and these are terabytes of data I'm talking about, ⁓ Literally terabytes of data. You are merging this hard disk, ⁓ removing the stale entries or rather skipping the stale entries, removing the deleted files and what not.

and you are merging them into set of hard disk. So you are basically freeing up the space doing defragmentation ⁓ and making them re-available for re-consumption. This is why merge and compaction becomes important for you to ensure an efficient utilization of the system. Now it was easier for me to say I will do merge and compaction. Now imagine writing code for it. ⁓ You have to literally start

alternating from one hard disk, which is why order n complexity is really essential when you're doing this. ⁓ You will have to check for every file that you have, which is stored in this hard disk, see if that file is active or not. If it is active, you keep it. If it is not, you skip it. And then you copy those things into other file, or into other ⁓ hard disk. Now, this is where one more interesting thing crops up.

is around what if

Sneha Mehra (01:39:27)  
I update a file. Now what happens over here? Let me put in Siddesh. Siddesh, what happens when you update a file on S3?

I mean, I don't know what happens, but I can say what I do. Yeah. ⁓ I would write a completely new file. Log structured file system, right? Yes. It would always write to the new sector, which means that when you are updating a new, when you're updating an existing file, you are always giving ⁓ S3 the complete new file object. Correct? You always do put, you give the entire file to be written on S3. Even though it is same path, it would go to a new

consecutive location in the hard disk. So no matter how many times you update a file, ⁓ are always getting, ⁓ you're always writing to new locations on the disk. Otherwise you do random writes and your performance would degrade. So which is why you're always doing sequential writes. So even if you write to the same file again, it would go to the newer location on the disk, which is how ⁓ S3 is able to do ⁓ object.

versioning. See how beautiful the system is as a side effect of having lockstructured file system. When you're writing to new location every time you are not deleting the old version, you can change. can go and access the historical version of the file. This is what when S3 says we give you object versioning. This is how they're implementing that because their file system itself is lock structured. So let's say

Just a detailed example of that. Let's say I wrote my file a.txt over here. Then at the same path I am writing the file. And now this side it is written over here because it is locked structured file system. It's append only. And now I wrote it the third time. So I have three versions of my same file. In which first I wrote hello world, then hello world one and then hello world two. I have three versions of my file. But because they are stored separately because

Sneha Mehra (01:41:37)  
every time you are writing it is writing to the newer location on the disk and across the disk you can actually go back and access historical version of the file which gives you object versioning out of the box. So when you enable object versioning on S3 it just does not delete these tail entries during merge and compaction. And if you disable object versioning

You say, I only want the latest version of the file during merchant compaction. ⁓ would skip those entries while it is copying the data. Such a beautiful side effect of that. Imagine this is a bug and S3 made it a feature that a allow in modern compression. don't, have to do less work, but I'll charge customer for the storage method. They're incurring and you can have all the versions of your file that you have ever created on S3. And a really beautiful side effect.

of having lock structured file system, writing it to a new location every time you can access historical virtual software. Now obviously you need to maintain somewhere that this file path or rather not file ⁓ path, S3 it is bucket.

and key. So this key is stored at this disk at this file path and so on so forth. And for all the versions, you would store this information and then you clean up during merge and compaction. So this is the information that you would store ⁓ in some database.

some database if you want to store it in some database it was because they are supporting multi versioning you need to give them a way to access historical versions so we need to store this information somewhere that these are all the versions that you have in most cases you would say Aaron I like this do it how many things with this storage that I try were doing right you need this is your business logic so you need to handle it on your own you need to know where it is writing it would tell you when your right is getting accepted it would tell you that this is where my right is getting accepted this is where I'm writing on this storage rack

Sneha Mehra (01:43:44)  
⁓ at this disk you would store this information somewhere which would allow you to go back and read it whenever you want it. And when you supporting multiversely of a particular file you would just keep this information handy that these are all the locations where all the versions of my file is stored. So version 1 stored over here, stored 2 over here, 3 over here so that you can access them at will. ⁓

⁓ Right. spending so much time on storage because it is such there is so such beautiful byproduct ⁓ of having this one constraint in place out of nowhere. We just got object versioning. Right. Okay. So now we know how margin compaction happens. Now we know how disk space is reclaimed when you delete a file. Does this entry from this database is removed. So when during margin compaction, it does not find an entry. So it would skip that entire file and move on and on.

and on and on. Right. And this is how your storage side of things function in case of S3. Right. Okay. ⁓ Okay. Before we go on routing, let me take some questions so that this gets cleared up. Then we go into routing. Then we go into hard partitions. Okay. Sitesh, have any questions? ⁓

Sneha Mehra (01:45:04)  
on the storage side of things, things that we discussed up until now.

Yeah, okay. ⁓ So just want to understand one thing on this ⁓ disk. Are we seeing ⁓ the location of ⁓ where our file pointers are? That is managed by the software given by a storage rack. Yes, ⁓ it provides where it correct. When you write it, you interface with this.

It writes it to the corresponding hard disk and returns you in which hard disk gets stored and where it's stored. So it is part of the same file system where I have these. It's out of one of the disk. It is. ⁓ Okay. ⁓ I just want to understand ⁓ this. Why? Because ⁓ like last year, one paper came out of S3 ⁓ with Manses. ⁓ This is now out of this ⁓ disk. ⁓

⁓ I read that. Yeah. So it's, both ways. So now obviously systems are evolving, right? ⁓ There are new advancements being made, which is where I now added this part of the story. I guess now abstracting a lot of things. ⁓ Del has put in a ton of effort into ensuring this. Yeah. It is not a still in production that paper. ⁓ but, ⁓ one of my friend who works at Dell works on this team. So that's why I these details. ⁓

got from that paper only now they are saying that lsm trees node where all the extented are there are out of that ⁓ thing okay okay so they are ⁓ i am trying to recall that yeah there's something mention of it i'll i'll see what our thanks thanks for obviously refreshing my memory i'll go through it today once again but thanks for that ⁓ got

Sneha Mehra (01:47:00)  
⁓ Hi, Hi, so can you, I think I missed it. Why do we store only 70 % of the data and then skip to another disk? Because this disk, why 70 %? Because it is still accepting the rights that are coming in. Remember when we store a file, we store off, we define that, hey, I'm storing this file of this size. Yeah. Right. So when you are doing that, so it is still accepting the rights. If you wait until the last moment,

the existing rights that are being sent onto this disk at this moment would be affected. ⁓ Plus when you are merging and compacting, you need some extra temporary storage. ⁓ Imagine the way you merge and compact two gigantic hard disks, you may have to store some temporary information when you're writing it. ⁓ So if you are filling a hard disk to the fullest, you would not have extra space to do that. ⁓ So the 30 % buffer that you have would help you in your merge and compaction.

and would also help you continue to accept the rights that are in transit before the complete cut over happens to the next one.

⁓ Thanks. ⁓ Yeah. ⁓ Rahul. ⁓ Just a simple question. ⁓ Like what would be a disadvantage of using chunked storage here? ⁓ Like let's just say we are not storing chunks in random blocks. We are just adding chunks. ⁓ Like it can be random hardest, ⁓ but it would be like a chunk, ⁓ we wrote a chunk to a file.

Any other chunk in a particular hardest will come after that only for a file. ⁓ So then your file becomes non-continuous. So your access may become slower because now when you're accessing it, you're doing a lot of back and forth movement to read it. This is cheap commodity hardware magnetic bliss. Head movement would come in. So your random reads would be very costly.

Sneha Mehra (01:49:03)  
Okay. Plus there would be ⁓ like hardest degradation also like because yeah. Plus, now this is little physical movement that that the ⁓ correct. So hard disk degradation is the correct term for that. That ⁓ the also called as basically the variant error of hard disk. Now the lifespan of your hard disk would decrease because you are doing a lot of, you are making our cheap commodity hardware do a cut of work then. ⁓

Yeah, got it. One last question before we move forward. ahead. ⁓ Here I had a question about logs that should file system. So last time we discussed, I think logs are structured files. Yes. ⁓ Yes. So when it comes to file system, do we need like we can just sequentially keep writing to a list. We would need some ⁓ three like data structure to support it like file systems are hierarchical trees.

So that's just meta information. ⁓ imagine those, those meta information is hierarchical trees, which is storing your directory structure. But the data is always stored sequentially one after another. ⁓ Okay. Got it. So you still have tree, but it does not mean that that's how physically your, your files are laid out physically. Your files are laid out one after another. ⁓ But logically they're arranging a hierarchical

Very similar to indexes in databases. Very similar to indexes in data. Got it. great. ⁓ Basically, Nidin, I'll take your question at the end. But this is the first time I've spent so much time talking about storage, but I'm really happy we went into depth of it. Why are we doing what we are doing? We now have answers to most questions. Obviously, I'm not a storage expert. I try to gather as much information as I could to make it understandable. So I might have...

⁓ I will have some limitations in understanding. Some things might be little too abstract, but if I were this engineer, I would definitely would have spoken about it. ⁓ things that I could pull off from papers, from talking to my friends who work in very similar teams, I tried to pull that off on the storage ⁓ side. Whew, storage done. Now let's talk about routing. ⁓ things is when we talk about routing, very interesting thing that crops up is

Sneha Mehra (01:51:29)  
How do we know, ⁓ again, classic routing problem. How do we know where to go and look for the data? So I have API servers over here. So I have user API servers and then I have a bunch of storage racks. So I have users, ⁓ I have your S3 API servers and then I have bunch of storage racks that I've stored.

So now how would you know where to go and read the data from? You take the day I need to know where I'm putting the data when I'm reading it from a classic case of routing. ⁓ Let's say I got a request for a particular S3 bucket. It's a bucket B1 key K1. And I want to know where should I write this request with storage rack should I write this request? ⁓ Because I also need tenant isolation. I also need a way to tell that hey,

This storage rack or this set of storage rack are for this organization. This set of storage rack are for this organization, but you still want to have, but you cannot just assign one storage rack for one organization because then what if the data or what if an organization does not use it? It's waste of hardware. You're chopping down your margins, right? Which means you want multiplexing ⁓ of you'd want one storage rack to be there for multiple

S3 buckets, but still you have to ensure that you are still isolating the load of the customer. both the problems are there. So now first challenge that would come in is how would you know which storage rack should I forward the request to? Right? A classic problem, a classic routing problem. So now a way to do it ⁓ is you may think, Hey, let me just use hash based routing. And the most common thing that would come to our mind is let's say I use hash based routing.

which is where ⁓ S3 is a key value store. You just specify that hey this is the bucket, this is the key. You take the hash of it, find where you want to store and you store it in that thing. Right? So for example, for let me take concrete names so that it becomes easy. Let's say I have Amazon Prime. Let's say this is their bucket name, Amazon Prime. ⁓ And for that, this is the key that they're writing. Let's say I have Netflix. This is their bucket and this is their key.

Sneha Mehra (01:53:54)  
And then let's say ⁓ hot start. I'm just using three computer names so that I can ⁓ help you understand ⁓ the importance of tenant isolation. ⁓ As you have hot start and they have key three. Now, depending on the request that you got, that hey, store this file, one GB file somewhere, because S3 is abstracting it for you. You don't know where it would store it, right? These are S3 API servers. Now, how do you decide where to store this data?

First part that comes to your mind, that, let me just hash. Like I have, let's say thousands of storage racks. What I'll do ⁓ is given a bucket and the key, I would take the hash of it. I would find a storage node to store and I would store it. Let's say you store it over here. Right. So for prime, ⁓ fortunately they have separate names. Superb. Prime and for key one, I stored it over here.

Then let's say for prime in that same bucket for Amazon prime in that same bucket. I'm having key two, which I'm storing over here for prime same bucket. I'm storing third file, which is being stored over here. And from P3, K3, K4 stored over here. And now Netflix came in, it's stored over here. Netflix came in, it's stored over here. Netflix came in, it's stored over here. So now you see that your one

Hard disk is now catering to request from multiple people and there is no locality. Now, what's the problem with this? For example, if ⁓ I am having a hard disk, which is being shared by your prime ⁓ and your Netflix. Now what would happen if prime sees a lot of load? The hardware is a physical. ⁓ The hardest is a physical entity. has physical limitations, which means that

If it gets a lot of requests coming in, it would not be able to handle all of them. ⁓ if I'm ⁓ if let's say prime, if let's say this prime video was let's say what the show that came in, was that rings of power or something, right? So this was rings of power, very widely used episode one, everyone saw that episode. ⁓ And here you have, let's say, ⁓ a Netflix show.

Sneha Mehra (01:56:19)  
pick any some Netflix or is a lie don't work with let's say sacred ⁓ games or what was that? Let's say sacred games. Not sure what that is, but I've heard something about it, whatever that is. Right? So these two are very popular shows part of the same artist. Now what would happen when you're making a lot of reads onto that the hardest will start to throttle to very high

to very frequently access data stored in the same hard disk is affecting one other. ⁓ That is bad. This leads to a hard partition problem. We would see ways to solve it. Hard partition problem is there, but you also see how heavy load on prime can affect Netflix. So this is a place which tells you that, we need tenant isolation. That somehow

We need to ensure that load from one customer does not affect another. Right? You say, simple na. Let's just handle it over here. Because all the requests goes through your S3 API server. Let S3 API server throttle. That is there. ⁓ S3 API server can throttle and will throttle. But still if you're having two important things.

or to highly frequently access things in same hardware that would still cause a problem. Let's say you do a limit. Let's say one customer can request 100 requests per minute at max, no matter what. ⁓ But if those 100 requests per prime and 100 requests per minute, both are under that constraint, but both are going to that same location, then that's ⁓ bad because one...

traffic from one is affecting another which is where hash based routing ⁓ is ⁓ not a good way to design system like this because you are ⁓ letting the control go in someone else's hand. What is someone else's hand? The hash function.

Sneha Mehra (01:58:32)  
Now this is where you need to understand why, where you cannot use hash based stuff, routing, partitioning, pick your favorite, whatever it is, right? Now understand this very well. With hash based routing, you take something, pass it through the hash function, you get something and then you go and do activity on that. Be it API server, be it in this case, it is storage, in API server you go to that API server and handle the request.

determine ownership that we spoke about consistent hashing last week, right? With hash based what happens is you rely on the output of the hash function that it would do it uniformly. Yeah, it is distributing the data uniformly, but across everywhere it is doing it uniformly. So because it is not under your control that where the data would go to it is wrong because there is a chance.

where two highly frequently accessed things come and reside on the same database. This is wrong on many levels. So wherever you need tenant isolation, you typically don't see companies going for hash based stuff, routing, partitioning, whatever you want to use, because it's not in your control. ⁓ Ideally, what should have been there? You know that these two terms or these two things are famous. So ideally, you should be the one who should be defining that, hey,

All the Netflix thing. Let's be at one place. All the primes thing. Let's be at one place. That would have been best for you. Right this way the load of one is not affecting others. Right. So which is where ⁓ S3 does not use ⁓ any kind of hash based outing which include consistent hashing because consistent hashing first step of consistent hashing is what you take the hash.

find one in the clockwise direction and you store it there suffers from this exact same problem. The consistent hashing hash functions are no different, right? Because then they're all determining ownership. So this is where when you need understand this, when you need control in your hands, you don't use hash functions. ⁓ As soon as you use hash function, you are leaving it to the fate. You are leaving it to whatever output of hash function is, ⁓ go and write there. It's okay.

Sneha Mehra (02:00:53)  
⁓ But when you need that control in your hand, you don't use hash based routing, partitioning, anything around hash based. So a better way to decide this is going with

range based routing because it gives you the control that you need that you define the range within this bucket from this key to this key this is where it would go from this key to this key this is where it would go you define that range now you can do that balancing on your end because now you are the one who are dictating that everything around prime video or not really prime video but everything around

In this bucket from this key to this bucket, this key, everything will go over here at the end. At the end on S3 your bucket slash key ⁓ is the unique key in which your data is stored. Correct? So this ⁓ is the factor on which you are doing range based partitioning. And when I say range based partitioning, let me use simple terms to help you understand what range based partitioning is on the board. So just to take a simple example,

Let's say I have keys from A to Z. ⁓ And then this is bucket names. So now if you talk about prime, ⁓ it would be somewhere L, O, P, Q, P over here. So somewhere it would be prime, ⁓ then shows, ⁓ then rings of power, then season one, episode one, something like this. So this is where that prime one thing would be stored over here. ⁓ This is where L, N, O, P, N is also over here. ⁓ So prime also over here, Netflix also over here. ⁓ You would have everything stored.

between prime and Netflix over here. So because this is where you are defining a range of keys and where it would be stored on physical storage rack or hard disk. Right? This is where it is being stored. But now you have this control. Now what you can do is you can split this. ⁓ Let's say this is where a lot of hot partition is all about. ⁓ A lot of traffic has come into it. Now I need tenant isolation. What do you do? ⁓ You split the range because all the

Sneha Mehra (02:03:07)  
prime video things is there in the prime video bucket and bucket names are unique. You would always have this part where you can define a range and association to a particular storage here. So for example, if you're ⁓ L2M or L2R is very hot and you want to split it, what you can do is you can have two nodes, create one node one and node two, move half data over here, move half data over here. You can do it at physical size. ⁓

your physical storage here. You can move the data here. It's a big process. That's why you don't do it often, but you need that control to pull that off. Right. But if you want it, you want to move the data. It's easy enough to do it. In most cases, if you are arranging the data, because this way your data is almost sorted by the keys that you have, you can literally take your data and move it to other places just by moving the hard disk here and there and some data movement here or there. That's initial complexity, but you get the idea. Right.

So here, if L2R becomes the hot node, I can literally split it into L2M and N2R in those ranges. And I just keep on splitting it by ranges that I have. And now, if let's say there is L2L, very short range, but within that also, ⁓ L say there will be multiple buckets and paths that are there. You can split within that because now you have the control to where to split and where to move the data.

And that's the beauty of doing range based purchasing. So the idea is wherever you need, wherever you need this things under your control, the way you are splitting, the way you are passing, you should need things under your control. have to do Right. So that's where you cannot rely. You cannot leave it to the fate of the hash function. That a very bad, put it, you take it in your own hands and do it that way. Now a very good side effect of this is

all the things you get locality ⁓ all the piece like all the prime video file would be very close to each other in one or two nodes they would be part of it all netflixes will be close to each other so now the load by prime video will not affect netflix because everything around prime would be one stored one after another over here or rather stored in the same object may not be physically one after another because block structured file system but it would be part of this

Sneha Mehra (02:05:34)  
similar set of hardware like a group of hardware. ⁓ And that's how you get with range based routing or range based partitioning, you get almost very good tenant isolation. ⁓ is one place where I always wanted to talk about that why you cannot use consistent hashing because you are leaving it the fate of yours into the hands of the hash function, which you should not.

Right. Okay. Now, now that we are on topic of hot partitioning, let me talk about another way. It's a small detour from our discussion of S3, a small detour, but it's important to understand how hot nodes are sought. So this is how S3 does it. And almost all Blob Studio does it, which is why you go through S3 document, you see they use range based partitioning, you go through S3 document, like their documentation. They clearly say that they don't use any hash based stuff. They use range based partitioning.

And they arrange the data such that they get tenant isolation. And so that load from one does not affect another. But just a small detour, not related to S3, but another way to solve hot partition problem that would come in handy when you're doing it. I think we have repeated this same thing in the past. I just want to spend one more like five minutes to this so that it's very clear in your head. ⁓ So another way to solve hot partition problem is when

by how you create logical shards of your data. ⁓ So in Instagram's ID generation, what we did is Instagram ID generation, saw how Instagram had multiple logical databases and fewer physical servers. ⁓ They created 8,000 logical databases, but they had three Postgres server to do it. ⁓ So this is each of that database. And this is that server. Now, when a

Database becomes hot because there's a lot of load on one of the data or one of the partition. What you do, you take that partition and move it over here because it is partition. It is very easy to quickly take a dump of it. You don't have to go through row by row and see which one is part of this partition. You can literally take them in one shot and put it there. Just do a DB dump. Every database has a DB dump on PG dump command or my SQL dump command.

Sneha Mehra (02:08:01)  
Every database gives you very easy way to dump an entire database, right? Because they are stored together. You don't have to go through each table and then each row and then take it out, which is why people typically prefer this way ⁓ of doing load balance in hot partition case. ⁓ here you have three nodes in which you have these partitions distributed. If one of the nodes gets hot, you take that partition.

and move it to other. You literally take dump load ⁓ and resume. Classic way of doing it. Now, not just that elastic search. If you have used elastic search and elastic search, when you create an index, you specify the number of shards that you want in that. The number of shards that you have is how in your elastic search cluster, the data would be distributed using the elastic search head plugin in elastic search that is called as a head plugin, which is very amazing plugin.

You literally can drag and drop a shard from one physical node to another physical node and it would do that movement for you. It would dump it ⁓ and then load it into another node and do it with almost no doubt. Right? This is what elastic search gives us. This is another way to solve hot partition problem that where you know you can easily dump and load it to certain place. This is how you also tackle hot node or hot partition problem.

Not related to S3, but a small detour to solidify the data. There are two ways to solve hot partition problem. I just piggyback that discussion over here. So now what you can do ⁓ is try this on an elastic search server. Try this on, try this head plugin. See how we can do the data movement. It's very easy. Like this is something that I did literally as a fresher. ⁓ What happened was ⁓ I was part of DevOps team.

We had massive elastic search question. I had no idea what it was, what I was doing. ⁓ Once in an engineer explained what is happening and just sharing those things with you. Right. ⁓ So I broke down production twice because of that, but that's part of learning. ⁓ So what happened is one of the elastic search node that we had was very hot because we were seeing degradation in performance using metrics that it tells us. ⁓ We know that we have to move this data to other node. ⁓

Sneha Mehra (02:10:27)  
Now how to do it, that's where head plugin we used, ⁓ which visualizes the entire Elasticsearch cluster on the UI. It looks very similar to this. You can literally drag and drop one and move it to other. The movement would happen. That is done by Elasticsearch on its own. We don't have to do anything. Elasticsearch takes dumping and loading and whatnot. It does that for us. A small script had to be written at that stage, but now Elasticsearch does it for you out of the box. ⁓ But this is...

One way to solve hot partition problem. ⁓ I just piggyback multiple ways to solve hot because it's a very common thing to ask like how do you solve hot partition problem? ⁓ Now you know how to solve hot partition problem. ⁓ First, ⁓ ensure tenant isolation, ⁓ range-based partitioning. ⁓ Where you need tenant isolation, you go for range-based partitioning because you want control in your head. ⁓ You still got hot partition problem. This is one way to do it. You literally move the data from one to other so that the request comes on that side. ⁓

Okay, we discussed that. ⁓ Let me go through the sub, but this is exactly what we discussed. Let's get storage and all keeping it simple. S3 API we discussed. Consistent searching, why we should not do that? ⁓ although it gives us minimal data transfer, but this is not what we need. We need a tenant isolation. That's the main concern. That's why we go for range based partitioning. It does give you uniform data distribution. We saw that, but

It may lead to that problem where multiple competitors are sharing the same underlying hardware. Not a good idea. That's where you go for range based partitioning. And I talk about splitting another way to solve it. Okay. ⁓ Now we'll talk about API side of stuff. We spoke about storage. We spoke about routing. So if you see, we are, we are basically coming down the over here, right? We talk about storage. We talk about routing. Now let's talk about API side of things. Now here.

There are a few very interesting things on how S3 or in general block storages make such decisions. ⁓ So here we need to know, ⁓ first of all, ⁓ when it comes to API server, ⁓ cannot, ⁓ like a very basic way of building it is I have user, ⁓ I have bunch of API servers of S3 and I directly plug storage to that, right? ⁓ Something like,

Sneha Mehra (02:12:53)  
this, you may think that, hey, let my ⁓ user request comes to S3 API server and it directly talk to the direct storage that are there. Now the problem with this is the request can go to any of the S3 API server, right? So every API server is equally likely to act to handle any request and it can directly talk to the underlying storage layer. Now the problem with this is that ⁓ I will in reality, I will have

100,000 hard disks or rather thousands of storage racks. ⁓ This S3 API server would need to know every single one of them, which is not feasible. Plus every S3 API server is potentially can handle every single request. That is also a problem. Why? ⁓ Because again, request from Prime, ⁓ request from Netflix affect each other. ⁓ Right? What you want is you want a clear separation of it. You want

that the kind of request the servers that are accessing the data from the underlying storage for prime are different from servers that are accessing the data from the underlying storage ⁓ of Netflix. ⁓ So which is where you see how this is prompting everything around tenant isolation, ⁓ which is what the crux of this statement is all about. ⁓ So what you do ⁓ is you have defined a table called

a partition map table. The partition map table stores that these are the partitions of your entire S3 range, entire, right? These are the partitions and this is where each partition is stored physically on the disk. And you have that mapping somewhere like you have huge enough racks to store that information, but you have this information stored and this is the partition map. ⁓ And this is partition. It serves from this to this stored in this disk and rack.

and you store that information in your partition map table. Now this partition map table will ⁓ need like this partition map table when ⁓ we get a request. This is where we get to know that who ⁓ owns this partition. So when we get a request, this is your actual storage. This is your partition server. Now ⁓ S3 API server cannot directly talk to storage because storage is very complex.

Sneha Mehra (02:15:17)  
storage racks, interfacing with reverse drivers and whatnot. So ⁓ S3 API servers, which are web servers, they are not ⁓ meant to talk to direct physical hardware because there's a lot of abstraction that needs to go into. So who talks to physical hardware? It's the partition server that talks to physical hardware. Partition server holds the logic who knows how to talk to physical hardware and getting some. Now in this partition server, so on physical storage,

The data is there, right? StorageDirect knows how to access the data from a particular Hattis. But the partition server is the one who knows which partition does it own. So for example, in my partition map table, I have some partitions that it's just entries being made. The data is physically stored over here, multiple storage racks and what not. But when a request comes in for a particular partition from the partition map table, we would know

that which partition server will this request or rather this partition server handles partition from partition A to B. So everything around A to B goes to partition server one, which in turn knows it is part of node one. So it will go to node one and read the data and what not. So this is the kind of metadata that is stored in partition map table. Now here what would happen is from partition server one, because I would know I got a request for a particular, let's say,

a S3 bucket that starts with letter B. ⁓ I got to know this, is handled by Partition Server 1\. I'll make a request to Partition Server 1\. ⁓ Partition Server 1 owns that partition. It would go to storage layer, read the data, send it back to the user. ⁓ Now if a Partition Server 1 is overwhelmed because it is handling multiple partitions, you can just move this over here to here, which is logical ownership. You just change this. Let Partition Server 1 owns this partition.

Partition server now also now it does not own a particular partition. This is logical. This is not physical data movement. It's just logical ownership of a partition to a partition manager. So the thing is one partition is handled by one partition server. But one partition server can handle multiple partitions. One partition server can handle or own multiple partitions, but one partition

Sneha Mehra (02:17:41)  
is with only one partition server. ⁓ Physically the data is stored in your storage layer that is now abstracted for us. This is what we talking about who talks to the storage layer. It's a partition server that talks to the storage layer. Who talks to partition server? S3 API server will talk to partition server. But which partition server will S3 API server talk to is determined by the partition map table. This way on the compute side itself there is a clear separation that

Primes with request does not interfere with network with Netflix request. Storage layer also that same separation. Compute layer also that same separation over here. Clear separation of responsibilities. ⁓ Right? Which is why range based partitioning is such a beautiful thing over here. Right? Okay. Now these are partition servers, but obviously these are compute. They can go down because now partition servers can go down or rather sorry. ⁓ If

Let's say one of the partitions ever goes down. Someone has to know orchestrator flow that we discussed so many times in previous system. You put that in that there is an orchestrator who is keeping an eye on partition servers. If one of them goes down, ⁓ would over change the ownership of this partition to that other server and whatnot. Right. That entire thing. Copy paste works like that. Right. But now this partition manager now who is doing that? ⁓ Orchestrator. ⁓

I forgot about it. Sorry. bad. bad. Partition manager is doing that for you. Right? Partition manager is the orchestrator. I forgot about it. Sorry for that. Orcus. Traitor. ⁓ So partition manager itself is the orchestrator who is doing that part for you into ensuring that your load is well distributed between the partition server. Partition servers are interfacing with the actual physical layer and getting the data, sending it back. Right?

But now this partition manager who monitors the partition manager, right? So you cannot have monitor for a monitor and then who monitors that monitor another monitor who monitors that. So there has to be an end to that, which is where leader election comes in. So between the partition manager, just like how we had leader election in orchestrator, same thing over here. You would have leader election over here that there would be one master partition manager.

Sneha Mehra (02:20:06)  
There will be multiple worker partition manager. The worker partition managers are actually checking the health checks of the partition server and balancing the load between them. While the master partition manager is managing the worker partition manager, if one of them goes down, it spins back up. ⁓ But if the master partition manager goes down, one of the worker is elevated to become the master. ⁓ This way your system becomes self-sufficient. ⁓ And this is where leader election comes in.

So again, reiterate, leader election ⁓ is there to make your system auto recover from failures without having any human intervention. That's why leader election comes in. So basically it's who monitors the monitor. When you have this question, understand you need leader election ⁓ as simple as that, right? Because everywhere, because the followers that are there are doing the actual correct work. If leader goes down,

Another one becomes a leader, classic how we work in corporate. When manager, manager manages a team, ⁓ if manager goes down, one of the team member conducts stand up that day. Right? That's it. Otherwise, manager's manager, director, director will not come and take stand up. Like there is no such thing like that. So you just promote one of them and then your team becomes self sufficient. That even if manager is down, someone else is becoming manager for that time, takes that stuff. And on next day, when manager comes back up, it

he or she resumes the responsibility. That same thing goes over here. So let's say the world of computer science is very much is very much influenced by how we operate, how humans operate. ⁓ You would see those things in action over here. So this is where little relation comes in to build self-sufficient systems. ⁓ Okay. Now one very interesting design decision. ⁓ And I'll just probe you all to think on those lines. Now here the flow is

Right. And now you folks need to tell me why this flow is very important and how it, you know, how it is one of the most interesting design decisions. Right. It's very counterintuitive, but it's really good one. So here the request from your user comes to your S3 API server. Right. S3 API server first consults the partition map table to know that for this bucket, for this path, which partition server owns this partition.

Sneha Mehra (02:22:32)  
And then S3 API server talks to the corresponding partition server. ⁓ Partition server then talks to underlying storage, ⁓ reads the data, ⁓ sends it back to the user. ⁓ Partition manager is sitting over here, rebalancing the load between partition servers via health checks. ⁓ If one of them is down, does that. ⁓ The master partition manager manages the partition manager as well with the leader election and whatnot. ⁓ Now, in most cases,

We have seen systems where you would see something like this request from user to API server. goes to partition manager. Then it talks to database gets it done forwards the request to server. ⁓ goes to database ⁓ storage layer gets the data sends it back. Why are we not proxying the request to partition manager? We typically never let other microservice

talk to a particular database owned by some other service. But here we are doing it. Why?

Sneha Mehra (02:23:41)  
We typically see this separation, right? Where we have one microservice, ⁓ not because this is partition map table, typically part of a partition server and partition manager thing. We typically tend to have like a, why S3 API server should talk to the, let S3 API server talk to my partition manager, let partition manager talk to partition server and get the data and send it back. Why is S3 API server directly talking to partition server by consulting partition map table? Why?

Sneha Mehra (02:24:13)  
Yes, it is. I think it's a data flow. As it ⁓ we don't know to see you get the information ⁓ metadata information with server you're going to write to another that ⁓ you don't want your whole data to flow through the whole system. So you want minimum house for data ⁓ and ⁓ but but why do we do it in other microservices? No, it depends on the data right here. It is a blob storage. You have storage. ⁓

huge amount of ⁓ size of data in transaction systems, ⁓ It's hardly one KB, ⁓ the At Max one KB, right? It's typically rows that you transfer here and there. That's very minimal amount of data that you are doing. You can still afford to do that. ⁓ Here, just assume that user is writing one GB file. ⁓ You let one GB transfer to a partition manager, ⁓ then partition server, then over here. That becomes costly. So that's what you do. You talk to this database, get this

partition server who owns it sends data to the partition server which writes to the storage. So you try to minimize the hop over here going counter intuitive to what we typically do that everything goes to everything is a micro-solid and it goes through that. So this is at one place where all the theoretical concepts that we have is thrown out of the window and we build system in a very practical way because it's okay to share the data because what are the chances if schema of this data is changes? Not much. It's a partition map table.

It would start this partition server owns this partition and is present in this rack. That's all. Right. Why should you even worry about that? Because it's not going to change. If it's not going to evolve, you can directly talk to the database, which just reduces the unnecessary hops that you have in your system. Right. And this is what, you know, really good practical systems are designed. Like instead of just going by that very hard bound rule. No, no, no.

It's microservice. Everything has to go through server no matter what. ⁓ Consider the situation that you have at hand. Consider the constraints that you have at hand and then you design system accordingly. So this is and one more thing. The amount of data is there plus the number of requests that this would need to handle will be same as number of requests a partition manager would need to handle unnecessary. Why? Because if you let every request go through partition manager,

Sneha Mehra (02:26:40)  
Data is there, but apart from that, number of requests also, because now this part manager, if it goes through this, this is just a proxy. It has no business logic in it. Why do we even have that? Let's just directly talk to this and get it done. ⁓ Minimizing the hops and reducing the size of the infrastructure to increase your profit margin is a way to make good businesses. ⁓ We talk about storage layer. ⁓ This would give you that idea around.

Who talks to storage layer? Partition server talks to storage layer. ⁓ Who talks to partition server? S3 API server talks to partition server. How we have designed low level storage details? We went through at the beginning of the presentation. That's exactly what I covered. ⁓ Now let's talk about other things that people don't think about. Durability. ⁓ So durability and integrity we'll talk about in depth. So data durability.

For SC is one of the most important properties because what you do on S3 is you write and forget. You trust S3 that because I've written a file, it will always be there no matter what. So you are typically in the write and forget mode. ⁓ with data durability, which means that for you durability is very important. That no matter if a storage node goes down, no matter if a rack goes down, no matter if your data center submerges underwater.

No matter if there is national calamity, like crashing everything, the data ⁓ should still be there. ⁓ This is the kind of guarantees companies have to provide. It is basically called as business. ⁓

continuity.

Sneha Mehra (02:28:28)  
plan. This is what every company works for. That hey, in case of a national calamity, in case of disaster, in case of outage, in case of attacks, what is your business continuity plan that no matter what happens, your business is always running. This is the first thing that your investors will ask you. What's your business continuity plan? When they do due diligence, when your company is raising funds, ⁓ SRE head is typically ask this question, give me a business continuity plan. And you have to provide if you're leading a site.

I had written three documents. know it very well kind of stuff that I wrote, which is why durability is so important because you cannot say that AWS gives you those guarantees. I had to take written consent from AWS team which says that they have a business continuity plan in place and this is what we as customers have to do and we attach it to our business continuity plan and share it with investors. It's all processes everywhere. ⁓ So data durability is one of the most important things that ⁓

No matter what happens, ⁓ your data should survive. ⁓ Because data is everything for everyone. ⁓ So what you do is when you are writing from a storage node, ⁓ when you are getting a write request from sdpia server to partition server, from partition server to your storage layer, you are not just writing to one hard disk, you writing to two hard disks at the time.

This way, if one hard disk goes bad, you have your data in other hard disk. You may also choose to store to write it on separate rack. So if one rack goes down, you have a data return in another rack. You may also choose to write across data center in the same region. Today, I have two data center in Singapore. I wrote it in one and asynchronously I write in another data center. ⁓ Otherwise, you can also choose to do it across geography. And I am writing one.

in Singapore, other in Australia. ⁓ So if Singapore submerges in water, I have data in Australia. ⁓ These are the things that you have to think about. So what S3 or any block storage gives you, and obviously you can anyway envision that when you're writing at so many places, your write would be slower and they are slow. ⁓ But what you typically do is up until you're writing within the same rack, you typically are okay. So in most cases, when you write onto S3,

Sneha Mehra (02:30:52)  
You're writing at two places and not one in two hard disk and not one so that you outlive the hardest failure. ⁓ So your data is still there. If one hard disk doesn't have it, other hard disk could serve the data easily. ⁓ That is durability. But if that data center goes down, you need to have data in another data center. If that another data center goes down, need to have data in geography. That is asynchronous and chargeable. You have to pay huge amount of money for AWS in order to do this. So you don't typically do this for all.

data that you have but for most of the data like for me for all critical data you have a copy across geography you always have it and it's part of your ⁓ due diligence that your investors do on the company that one copy of data has to be there in another region no matter what and you may do it periodically like per day backup or something like that but it has to be there in another region because if one data center goes down your business should still function

No one cares about anything else but business because money is everything, everything is money. Right? Okay. So now you may ask this question that we talk about racks and all and hard disk and all, but what if a partition where I wrote the data, it got corrupted. You say I have data in another rack. I can leverage it. You can leverage it, but you can go a step further.

and configure RAID. You might have heard the term RAID, redundant array of independent disk in your database course. ⁓ If not, not from CS background, just Google RAID, R-A-I-D. You will find what RAID does is there are multiple RAID formats. 0, 1, 2, 3, RAID 1, RAID 2, RAID 3, RAID 4, RAID 5, RAID 1, 0, multiple of them. ⁓ You can read about it, but the idea is that within the same hard disk, instead of writing to just one sector,

You write it to sectors. So if one sector gets corrupted, your data can serve from the same hard disk from another sector that is read for you in simple terms. for anywhere, always remember this. Wherever you hear the term durability, the only way to guarantee durability is by making your data redundant. ⁓ Any place that you can make a data redundant is a potential solution. So first level, when you're writing to a hard disk, you write at

Sneha Mehra (02:33:19)  
two sectors instead of one, which means you are literally making your disk storage half because if you have 1000 GB of storage because you are writing at two places, you are only leveraging 500 GB, but it's okay. Data durability is important. You do that. That is rate. You can do it to hard disk within the same storage rack. You can write across storage rack. You can write across data center. You can write across job. So the only way to guarantee durability

100 % durability is to have multiple copies of your data and having a simplest way to access them. ⁓ Depending on how sophisticated you want your system to be, you define those things. Some things would be free, some things would be chargeable. AWS typically charges you when you're doing it across data center and across geography. Within data center, it's typically free of cost and it it out of the box for you. ⁓ Now next, as a final thing, before I take question, is data integrity.

Now this is something that ⁓ no one talks about for some reason, but integrity of data is very important. Now imagine you're writing a file, you're writing a file from your user space onto S3 and the content of the file is BAT. Right? But when you are writing over the network, a bit flipped, when a bit flipped, your B became a C.

because I didn't just binary the edit distance between B and C is just one bit. If that bit flipped. So instead of writing back, you just wrote cat on the disk. And now when your user reads it, user would read cat. I did not, I never stored cat. I stored bat. Where is my bat? Right. Which is where maintaining integrity of your data is really important across the layers. So what you do the only way

error not on you. The most common, most performant way of doing integrity checks is using checksums. So what you do is you have to ensure whatever you got from the user is exactly what you are storing on the disk. You are not storing anything funky on your disk. You have to ensure that. Now the way to ensure that ⁓ is by adding checksum to everything that you're doing. ⁓ So at every single layer of your writing,

Sneha Mehra (02:35:43)  
So from user to S3 API server, checksum is passed. ⁓ S3 API server to partition server checksum is checked. From partition server to disk checksum is checked. From disk when you read it, checksum is checked. From ⁓ partition server to S3 checksum is checked. From S3 API to user checksum is checked. Even on user AWS SDK checksum is checked. Even in HTTP request checksum is checked. Everywhere. Because data correctness is very important.

what you stored or what you think you stored is what you should definitely get when it is written you have to ensure it especially for a use case like S3 because it's a source of truth for a lot of things you cannot just corrupt data right so data integrity is really important if you are doing checksum check for so many places wouldn't it be slow no it's not checksum is simple 32 bytes at a time you process ⁓ it

you XOR XOR XOR XOR XOR that's the simplest way to do checks out like you can just take XORs of 32 bytes you read the data 32 bytes thing or a batch of 32 bytes and you keep on XORing it one after another so it's not really expensive but it is much needed for you to have your integrity check so in case of integrity fails while storing the data you would reject the write and say that if it fails please read right ⁓ but you have to ensure that what you writing is what a customer thinks it is writing

That is integrity for you. And you have to have to have to have to ensure that this is something that people overlook when we design systems. We never think of integrity. We never think of checksum. We never think of data correctness, but these are very important thing when you are designing an actual system. ⁓ So where did I get all of this information from? The first source of information for me was this paper, which I would highly recommend and which is my

The names that I gave was very weird, partition map table, partition manager. I typically don't use those names because these are taken as is from this paper called Windows Azure Storage, a highly available cloud storage service with strong consistency. Read this paper. Given that we have covered this part, it will become very easy for you to read this paper. But S3 has not made that information public. Windows have, Azure Storage have. So I use that as a reference to cover it because I've never built S3. So I don't know what it is all about, but...

Sneha Mehra (02:38:08)  
I tried to gather as much information as I can by writing, pre-prototage, talking to my friends, working in similar domains to gather that information. But this is the best that I could offer at this moment. ⁓ But I'm sure I do have some gaps in my knowledge, which I'm constantly improving on. ⁓ Hopefully I'll find something new ⁓ to add in the future. read this paper, Windows Azure Storage, a highly available cloud storage service with strong consistency. You'll get an idea on how that, and this is exactly what like Windows version of S3 is.

And another data, another paper that I want you to read is this building a database on S3. This is actually the name of the paper, right? Building a database on S3. This would solidify your understanding of how to make things variable on S3 and how we did with word dictionary, something very similar. There are very interesting hacks over there. So read this and the paper. Third one is a blog post. ⁓ There is also a paper on it called Scuba by Facebook. It is Facebook's blob storage Facebook.

has its own private infrastructure block storage with the use to store media files and whatnot. They also have paper on this. So read this paper to build a very real understanding of S3. These two to build S3, these two how to leverage S3. ⁓ And ⁓ one book that is great for you to understand database internals is called Database Internals by Alex Petrov. One of the best books out there. Get a copy, read it through and through and you'll love it. He has covered some amazing

details around most commonly used databases in very much in depth. So go through these books called database internals. It does not cover relational database internals around query optimization phase and whatnot, but it still covers a wide range of databases and things you should know about it. It's pretty brilliant book in any case. Highly recommend you to read that book to build a very detailed understanding of databases. And yeah, this is all what I wanted to cover as part of S3.

Folks who want to drop off, please drop off. I'll take questions and it will be all part of it. God. I bet I had so many questions. I forgot some because of ⁓ the content. I don't remember and ask again. But the I mean, if you could reverse manner, the replication part ⁓ now here, we thought that we were writing ⁓ on a sequence of bases in the rack ⁓ with the head moving. ⁓ The question which came to my mind is if you want to replicate in the same rack, the data

Sneha Mehra (02:40:34)  
You have only one currently active hard disk on you which writing right? ⁓ How would you replicate to a different disk in the same rack data? No, so the head movement is not within the rack head movement is within hard disk. You can still issue two rights on this thing. But head movement was like the active disk, right? On which we are ready. So now you have two pointers, right?

One is primary, one is secondary and you write it two places. ⁓ Okay. Two pointer thing was not okay. ⁓ Okay. So one second thing was that the ⁓ hot ⁓ key partition thing, right? So you told that if you are, so you told that the node is like hot, you can move partitions, but how would we know which partition to move? Mediocre partition metrics. would know which one is causing the problem. You'd have analytics on top of it, right? You know, for which key are you getting a lot of requests?

And then whatever partition is blocked, you need observability. ⁓ need to make those. So whatever system you're building, you need very high level of observability so that you know what the problem is. So you remember, as you put the full question, the first or second week that what is the actual formula to to ⁓ sharding? Right. So this is the formula you observe. You have a metric and then the metric is the source of truth for you to balance things out. Correct. Correct. Because unless you know

⁓ like unless you know what's causing the problem, you cannot fix it. You would just be fixing random stuff. ⁓ So that's where observability is really important. In this case, ⁓ assume that your request logs are that thing because your request log would content that for this key, we are accessing it. So you have that observability that how many requests one S3 bucket is receiving ⁓ and you would get, now this S3 bucket is a paid one. So I'll take this data and move it to other place. Okay. So before we move from this slide,

durability means like we thought about sector level replication and node mean, hardest level application. ⁓ But then eventually data center can go down as you told right business continue to learn. ⁓ at least we should make sure that ⁓ at least one separate location the data is actually replicated right? Yeah, don't act back. may be for a day. It may not be very frequent. It may not be like every right is going in like everywhere. But at least once a day complete backup of the database into a different region is something which everyone prescribes.

Sneha Mehra (02:42:57)  
And the rights to let's say a user supporting a file rates and you're replicating to a different data center. ⁓ Let's say the replication part the light part again to a different happens actually to the same way as it happens to the primary one, right? Yeah, yeah, but it will happen after a delay. It's not real time. ⁓ yeah. Yes, in case right. And then you have to maintain a map also. Okay. What is the primary secondary partition? ⁓ Yes. Okay.

And as soon as you add durability complexities, a lot of things that we discussed changes and now a lot more additional information you need to store. ⁓ All of that is stored. And the ⁓ point around partitions, key partitions of bucket and the key is the input on which we actually put a range. Now, if you take a real life example, bucket ID is like a bucket for a particular ⁓ user who creates

⁓ account basically. So if Netflix has his own account, let's say one account will be simple to have some simple observation ⁓ on that. There's so many videos. So many videos are there. Right. ⁓ So if it is like ⁓ if it is like a ⁓ file is the key, right? So if the file is the key, then they could be, I mean, the starting point would be like Netflix. Right. So N is the starting character. So you did a character ⁓ alphabet based ⁓ party. ⁓ That means the Netflix itself has so many movies. Right.

So in itself has so many data, right? ⁓ So if, ⁓ you end up in a situation where one alphabet has a lot of data to within that, right? You go and be NC. ⁓ Yeah. That's I was to ask. ⁓ Right. You go and which is, said, ⁓ it's not a try based approach and any, any like try based approach. Yeah. I simplified it with single characters, ⁓ but do it at that path level where you'd want to draw that light that this is where this would go. This is where that would go to other partnership. ⁓

⁓ Okay. And you're putting it all alphabetically. What happens is all things belonging to the same bucket would stay closer to each other. Yeah. that I know the clarity of reference. ⁓ the last question here is the place we were trying to write it on, on a note, right. And a disc, basically you can go to the page, ⁓ maybe writing the data to the, ⁓ disc, the first slide. think we were writing data to the desk where we have

Sneha Mehra (02:45:16)  
So we were trying to write data to a disk, right? So this one, this two, we have, had the headers. No, not ⁓ the previous one to this, not this one.

Okay, let's take any any example. ⁓ Yeah, even even here itself. ⁓ Let's say you were you said that no, you introduce a concept or two pointers. Let's say we have one pointer, which is active disk. ⁓ Now you said that, okay, let's say we have a file of 10 GB and we know that okay, our current this size is 90 GB and by the time the next slide comes 100 GB 100 GB is the size of the disk and it will be over and you just move the move the head to a different disk, right? You calculate that

But in point of time, let's say five GB written and then some reason the right field, the upload field basically, ⁓ and your header at the point that has moved. ⁓ Right? ⁓ Now, how would you do a fault? ⁓ How do you recovery? mean, how you will ask the uploader to upload the file again? That's what happens at the moment. That's what happens also at this moment. And that is where to reduce that you can do chunking and water. Those are enhancements that you can.

Okay. Yeah. Chunking means the I know also, right? If you have multiple files, I don't actually go out. ⁓ So ⁓ yes. Right. You need to know, because if you try to solve one, you would lead to another problem. So either you ask the customer to re upload that part. If it's small enough, then three upload would work. Right. But in any case, you're, you can still write it as an exception.

You can still write to the old team because you still have 30 % of extra space to continue that right. can do that. ⁓ Yeah, okay. ⁓ These are very, very intricate details of it. When you implement, ⁓ this is you would discover that part. But it's good that you folks are thinking on those lines. ⁓ Now think of implementation, which is where you see far more challenging stuff there. ⁓ And it's not difficult to prototype. can start prototyping it. You would start stumbling upon very interesting edge cases.

Sneha Mehra (02:47:14)  
The fat case is like you are uploading tens of GBs non 10 GBs like sensor. I'd be discussed. It's 100 GBs. Right. So network thing was you have to break down files and chunks and then send out. Yes. Yes. And now you have to look carefully. Now this chunk is just an extension to this. Instead of storing one file of this part, you implicitly chunking it and storing it so that your user doesn't have to resume the like, sorry, the user can resume the rights from the last point. ⁓

Now that is an extension, in the first version of S3, they would not have not, they would definitely not have built that. But in future version, they would have added that feature. ⁓ But if in case you do a chunking, so you have a chunk manager as well or the partition manager takes care of this. I mean, have a DB that's implemented. I don't know. I have not put some thought around it, but I would let hard disk tackle it and maybe in partition in the file, I would have met information around what all chunks are there, where all are they placed and whatnot.

Because now I have to still provide a way to seamlessly access the entire file. You have to accumulate all the chunks, right? And you replicate also all those things. That is different. That is you are going in very different direction now. If you are still combining all the chunks and storing it in one place, that also becomes a problem. ⁓ No, ⁓ chunks will be in different places because anyway, ⁓ I mean, not different places. ⁓ have to be in a single place. Stacking it just basically to reduce on the network load and resume the, the, the point you want to upload. ⁓

Okay, it's more complex than we can what we can discuss in five hours. ⁓ It's more complex than that. ⁓ Okay. so how does the read path work? Like, I think we discussed in the storage racks, you have, ⁓ you have basically head pointer, which is just keep on appending items to there. But let's say you get a read request for some item, ⁓ which is now past the head.

Now, ⁓ like, does not mean that, see your head can always move back and read stuff. Random reads are allowed. Random writes are not. ⁓ S3 is not a read heavy system. ⁓ FII. ⁓ S3 is a write heavy system. It's very contrary to think S3 is a read heavy system ⁓ because S3 in reality, there is very high amount of data which has been written and very less data which is read from S3. ⁓ S3 is a write heavy system. So,

Sneha Mehra (02:49:39)  
In S3, ⁓ the write to read ratio is huge. ⁓ The number of writes are more than the number of reads that you make on S3. ⁓ So when you are in such situation, in your read path, ⁓ it's not always the case that I would only like because my head move forward, I cannot go back. Your random reads are allowed. ⁓ And similar to what we did in BitCast, ⁓ where we were still doing random reads, ⁓ our head used to move back, read the thing, send it back to the user and then move forward. That's what we do over here. Now,

the because we moved from one hard disk to another does not mean you would not read from previous hard disk. Reads can still go to the previous hard disk that are written.

Right. And ⁓ like the margin compaction part, think that would change your, some of the mappings of where the data resides. So yes, ⁓ as part of that process. ⁓ Okay. And when you say that, like, I'm still not able to digest that when you say that S3 is more of a write heavy system than a read heavy, like, ⁓ like most probably you would be getting like you'll be

issuing a lot of gets on S3 like be it configuration data, be it static data, ⁓ be it file like you maybe write maybe some ⁓ database files once, but you are reading it quite often. So if you could maybe elaborate. how many, but how, are you see the amount of data you have written on S3 versus amount of data you read from S3. There would be a few files that you are reading more often. In most cases, there are smaller files, configuration files as you gave an example.

Right? it comes to large files, do you read large files often?

Sneha Mehra (02:51:24)  
No, no, no, ⁓ write a lot. ⁓ There are a files that obviously there are a few files that you are reading more often than others, but it does not make S3 read heavy database because the amount of data that is being written and the amount of data that is being read, ⁓ the amount of data that is being written is more than the of data that is being read from us because S3 gives you very slow reads. So it does not make sense for you to read it in transactional use cases.

And in most cases, the high traffic things are transactional. ⁓ Analytics is not that read-heavy. ⁓ analytics is not that frequent anyway. ⁓ Because S3 gives you very slow reads, people don't use it to read things often, ⁓ except for a few smaller configuration files here and there. ⁓ So that is why S3s are right heavy use cases and not read heavy uses. You get it? Okay. Yeah, got it. ⁓ Because even when I was diving deep into this,

I was going through, I think it's just documentation itself. ⁓ I think they mentioned somewhere in the start that S3 is right heavy. ⁓ I like, no, I'm reading so much from S3. But then I spoke with a couple of folks. We were just brainstorming on this. And then we found some paper, I don't really remember, ⁓ six or seven years back, we found a paper which was not six, five years back. Yeah, five years back, 2018\. Yeah, yeah, 2018\.

In 2018, I was going through this ⁓ while I was in middle of the job switch. I was going through this and I found a statistic which had amount of data being written versus amount of data being read from S3. The amount of data being written was more than amount of data being read from S3. And that's when I then we started probing and we found out that the amount of data in the organization that was part of the amount of data we wrote was much more than the amount of data we read from it.

which is why S3s are right heavy. People typically write and forget. In most cases, it's backups that you're storing on S3. ⁓ In most cases, have like obviously there are cases where you're serving images, you are serving videos and whatnot. But what are the odds that every video is equally likely to be accessed? Take example of Netflix. One or two video very popular being accessed. Everything else is just sitting idle then. So it's very right heavy system, not a read. So a system is read heavy when you have 90 to 10 ratio for reads.

Sneha Mehra (02:53:44)  
So that ⁓ for some places, for most places. Yeah. And I think for media files, the CDN will also come into picture like when the traffic is there. okay. Okay. Siddish. Thanks. Hey, I just wanted to discuss like, so this system that we developed has a few major limitations, right? Like one is the size of the file. Yes. Number. ⁓ Sorry, I didn't get it.

Yes, so this is the file is the bad. Yeah. ⁓ And second is ⁓ the amount of data. Like ⁓ it's, I'm ⁓ having hard time understanding that it being a cold storage, ⁓ are not deduplicating any of the data, rather we are actually ⁓ replicating it. So users facing five terabyte data is actually getting converted to something like RF3 also it's 15 terabytes now. ⁓

So ⁓ I was just thinking like, this seems a ⁓ lot to me because ⁓ we want that organization. Organizations are going to dump petabytes and petabytes. It's going to get amplified. Yeah. 100%. I totally agree with that, but that's the kind of guarantees AWS has to provide. If you ⁓ if you, if you delete the data and if

They lost their original copy of data. You don't have another copy of the data. Let's say, because it's a very common API. It's a very common risk case ⁓ where S3 folks get tickets. Like say, hey, I accidentally deleted my data. Can you recover it for me? It's very common. To be honest, it's one of the most common requests that S3 folks get. One of my friend works in S3 team ⁓ and they analyze this part. And they said that almost at one stage, almost 30 % of requests.

that they got 30 % of support cases that they got was about recovering that data. That's when they added into that part ⁓ on the UI, but through which you can recover the data. Like you can see your recently deleted files and you can recover it from there. And which is, which is insane because it's so accidental to delete data. And then you realize that I should not have done it or maybe a bad script happened and you deleted it. ⁓ So that's why having another copy of data is very important. Second,

Sneha Mehra (02:56:06)  
In most cases, people are now not using S3 to just dump their data because data lakes are built on top of S3s nowadays. So it is not just ⁓ one place to which is just like my backup. In a lot of cases where now companies are building data lakes on top of S3, given that now S3 is HDFS compliant, any workload that used to run on HDFS can now seamlessly run on S3. If it is HDFS compliant, people are just using it as their massively distributed storage file system.

and data lakes are built on top of it. So now that is not just becoming your backup storage, but also somewhere you are relying your business analytics on top of it. Now your business decisions are being made on top of it. Having copy of data is very important now. ⁓ So, and they still charge you very handsome on four cents or ⁓ 0.4 cents per TV is a decent amount at a job, other four cents per TV is a very decent amount at a job. Right. So

They're minting money because of that. But we think it's, ⁓ we refer to it as dirt sheep. It's actually costly when you see the lifespan of an organization and the amount of data you are storing there without deleting. The moment that's, that's the thing, right? Once you move it, it's very hard to move it out. ⁓ And that's the trick. Yeah. Anyway, the end of another limitation I felt was ⁓ this system is going to be heavily fragmented eventually. Right. Because once you start accepting.

gigabytes of files. ⁓ That is where I was wondering to like initially, there to solve the fragmentation problem? But how much can like in the sense that's I was going to ask like, is it solving per partition? ⁓ Or is it solving? ⁓ Like, do you do this? Like you actually when you merge, you actually ⁓ defragment the whole disk? defragment the whole disk. That's what happens. ⁓

Defragmatic the whole disk to get the maximum available space on that. And then you move the data and then you reclaim the space and then you basically reuse those space. ⁓ You do that. do that. That's why modern competition is a very heavy process in case of S3. Very heavy. ⁓ Yeah. That's ⁓ another question was ⁓ in this partition server. ⁓ have like confusion in this diagram. ⁓ One partition server is, is it like a logical partition inside this rack space?

Sneha Mehra (02:58:35)  
Or is it pointing? ⁓ It's a logical partition that is exclusive, right? Mutually exclusive. So this partition is different from this partition. So one partition server can own multiple partitions. So which means any read request for a particular partition that comes in, it always goes to that particular partition server. And because this partition server owns those partitions, it knows how to read it from the disk. It reads it and serves it back. ⁓ And if this partition goes down, we move this to other partitions. So this is just a logical ownership of a partition.

Physically data is stored over here. Got it. Okay. Yeah. The reason I asked is because we did not go into like this system is simple enough to not go into concurrency. ⁓ exactly. Yeah. And you have kept it actually. Yes. Yes. And we don't need it to be concurrent. We don't need it. Yes. Yeah. But there is one problem though. Yeah. One problem that maybe we it's a complicated problem, but when we move partitions, right. ⁓ Biggest issue is

like ⁓ generally now two servers there is an intermittent time when you may have it accessible one and two is there could be some stale requests ⁓ which got yeah but files are not getting updated it's read only

I see when you move. perfect. ⁓ I was missing this. This is very important. That's what I say. It's three is right and forget. You are not. It's not a transactional DB in which your data is being changed on the back. It's written. It's written. Yes. It simplifies a lot. Exactly. That's why it's like a lot of things that we think. How to handle this? It does not need it. The problem with this is I spent a lot of time talking about storage site.

because everything else is oversimplified for S3, given that does not have to care about concurrency. Once file is written, you are not updating it. Any updates is a new file. ⁓ Yeah, that's also a huge amount of ⁓ problems. Yeah. And another thing is this, when we talked about replication, I think, I don't know if we missed it, but everything needs to be replicated, right? Even your metadata table needs to be replicated. So this partition map table itself has replicas and whatnot.

Sneha Mehra (03:00:50)  
All those complexities, right? Everything we studied previously, copy paste over here. How do you ensure uptime of the database, failovers and whatnot? Everything, everything fall taller.

Okay, good. ⁓ Yeah. Yes, go ahead.

Sneha Mehra (03:01:09)  
so I have question on the very first thing that we discussed that data back, right? ⁓ So I'm thinking in terms of how are we storing the file? Like what kind of file system we are using? What kind of like, what's the maximum limit of what first file? So, ⁓ okay. So first of all, can we shed some light on like what should be the file system over here? ⁓

Okay. Like what's the ⁓ maximum? Are you talking about EXTFS, BTRFS? So similar to that, is lockstructured file system. There are two very popular implementations of it. not, I really forgot the name of that. ⁓ But you are formatting the disk in non-lockstructured file system. First thing, right? Now the limitation of file, that's a brilliant question. Like fat files, ⁓ like if you have a fat file system, apiT file system, ⁓ the limit is one GB. You cannot have more than one GB file.

Right? Right for NTFS is this I think 4 GB or something and for EXTFS it is 16 GB or something like that. On S3, ⁓ I think S3 tells that the limit is 1 terabyte or something. Obviously S3 internally does chunking and all, but let's say we don't think of chunking at the moment, right? Then we'll introduce the, the complexities of chunking. Let's say we don't think of chunking at the moment. Now what we can do at this stage, let's say we define a limit. We define a limit that, my file cannot be more than 1 GB.

hypothetically, right? So which means that whenever a write request comes in, as I said, when a write request comes in, you have to tell on this bucket at this path, ⁓ I want to write a file which is this big. You may not provide the entire bytes, but you are reserving the space there. Right? This is what you do in your write operation. S3 write operation. Either you send an entire file or you request S3 to reserve a space for you. Correct? In multi-part download. Right? Okay.

So now in up until now, what you have is you now reserved some space to store that file. Let's say this is maximum file, one GP, which is other. ⁓ And then you write data over here. You say that, Hey, this is the, this is the chunk right here at this offset. This chunk right here at this offset. So it gets written over here. And then you complete the file. is how your multi-part upload works. Right now ⁓ S3 can give you like on this file system. So there is one limitation of

Sneha Mehra (03:03:33)  
file system that how big this file can be. It depends on which file system are you using. Let's say lockstor file system type of the example. Let's say it gives you a limit of ⁓ one GB file, but I need more file. need big file. So let's say if S3 supports, I'll support one terabyte file you can have on S3. So what S3 can do is S3 can accept one terabyte file, but store it as 1001 GB files. ⁓

Okay. And maintain this metadata. So S3 has a limit which might not represent the limit of the, I'm so glad you asked this question, which might not represent the limit of your underlying file system. Because lock social file system cannot block a huge chunk of file. Right? So this is where your inherent chunking comes in. And they would need to anyway maintain that this is file path one, two, three, four, five, six stored at this location maintained in inode.

But together it is one big file. Right. So first limitation might be one GB and I'm just giving random example, but as three can say, can support 1000 GB file as well. Right. But then how is three would be implementing it would be something like this. Okay. Yeah. ⁓ another thing, ⁓ like one kind of like suggestion. ⁓ so basically let's look at this. ⁓ Okay. And what if we ⁓ split our file, whatever size it is.

into smaller chunks and distribute ⁓ across the HDDs ⁓ and so that we can implement a parallelization ⁓ among these HDDs. So can we do that? Like what's the pros and cons of that? Great question. Why do you need parallelization? Because HDDs, ⁓ mean, ⁓ let's say one, ⁓ one GP file is being, ⁓ is it a request for one, one GP file ⁓ for read? Okay.

The whole load would be on one particular CD and it will sequentially read it and serve it. But what if we split it into 10 chunks and split it across the whole CD racks. ⁓ Okay. ⁓ So as a D one, two, three, ⁓ will be badly reading it and serving it to the partition. ⁓ Whatever you call it, the partition computer, which is on the rack. Yeah. ⁓ So apparently all those tongues will be this, ⁓ I mean, ⁓ right.

Sneha Mehra (03:05:59)  
serve it to the partition ⁓ computer and then ⁓ it will be assembled and sent it to. Okay. Now my counter to that is simple. Is S3 a read-heavy system? Let's say, let's say, let's say. No, no, wait, wait, wait. Is S3 read-heavy system? No, right? People are not going to read frequently huge items because huge cost is involved on that. Right? ⁓ S3 is a write-heavy system. So that solves that problem partially. But in any case,

You have multiple copies of data stored. If you want parallelism, can leverage multiple copies of data that you have any best store durability. Yeah. But that's, ⁓ that's stored across different regions. ⁓ Even racks within rack, within hard disk, within rack and within data center. You have that, right? You can leverage that if you need, if you need little faster reads, can leverage parallelism using durability.

because you have multiple copies of data. ⁓ You don't need to chunk and store everywhere because that makes things much more complex because now imagine someone want now for you to read this entire file, you have to read different regions from there. ⁓ Now user is waiting for it, reading from multiple places will take time. ⁓ Let's say user is sequential and in most cases, ⁓ our file is on S3 is always sequentially accessed. ⁓ It's not parallelly accessed, it's always sequentially accessed.

because you are not sending everything in one response. Correct? Imagine S3, the way you are accessing the file, you are asking to download the entire file. You asking to download the entire file, are iterating bytes by bytes, sequentially and then accessing it. Right? That is very common. So you always optimize for common use cases. For anomalies, you don't optimize the system for anomalies. ⁓ You find hacks to solve anomalies. This is where if you want parallelism,

that S3s can do parallelism through the durable copies of data that it created within the hard disk, within rack, within data center. No more latencies there. So you have three places where you can read those same things in parallel if you want to. ⁓ In any case, it's not reading. So it's okay. ⁓ Does that answer? But I'm not stopping you from implementing this. ⁓ what you're doing is you're building a read heavy block storage system. If market needs it, you build it.

Sneha Mehra (03:08:26)  
Yeah, I mean, let's do what you suggested. ⁓ Let's say I'm building a Netflix then ⁓ this would just case. No, Netflix CDN server. ⁓ Let's say I don't have that CDN technology and all. Let's say I'm solely relying on this S3. So then you create thousands of copies on S3s and then basically balance the load ⁓ between them. Okay, so multiple tracks are serving different different chunks.

Not chunks, different same file copied at 10 different racks and you're balancing between those racks. ⁓ Okay. ⁓ I would always find other hacks rather than chunking and store because I want that file because even in case of movies that you take as an example, the ⁓ way you will be accessing the movies is always going to be sequential access. If you're storing different things at different places, is additional overhead for you to do that computation.

which chunk is stored where now I read this byte now I have to read the next byte or this but the first five bytes are in this chunk next 50 bytes are in that chunk that is additional computation keep it very simple got it? ⁓ So we can so basically we have a storage we can we compute this data and save it in the partition server then the partition map becomes big that additional metadata itself becomes huge for you to access it yeah so we are

like kind of trading of the memory utilization over the time complexity. But this is not it's not I'll give an example. I'll give an example where it is there. ⁓ And I'm not denying your answer. Your answer is that I'm just proving it to think in different ways. Wait, let's you chunk the file into 10 bytes. I'm taking simple example. Let's say you chunk the file in 10 bytes. So this is zero to nine. This is 10 to 19\. Right. Now you are iterating this file byte by byte.

rather three bytes or rather four bytes at a time. Now what would happen? You are reading four bytes at a time. ⁓ Now what you would have, you would need in meta information that this is chunk one, which is stored in this file over here, which is of 10 bytes. This is chunk two, which is stored over here. ⁓ Now we are iterating file four bytes at a time. Now what would happen? You would read four bytes, you read zero, one, two, ⁓ You read it. ⁓

Sneha Mehra (03:10:49)  
Then you read next four bytes, four, five, six, seven, then over. Then you read next four bytes. Now what would happen? You would have to read two bytes from here and understand that, hey, I'm missing two more bytes. Then you would go to this database and find where the next chunk is. You would go to there and then read two bytes from here. Correct? This is additional complexity because now when you are requesting to read a chunk,

And now on S3, ⁓ can provide this, you can fire a query that says from in this file, read from this offset, these many bytes. It's a very common use case on S3. ⁓ Given that you can fire queries like this on S3, which means there is a possibility where your request can span multiple chunks. Now you have to do this computation that this offset is presented this chunk at this part and this offset is in this part. I have to read these many things. Not worth it.

Right? Because it's not a read-heavy system. What you're proposing is good for read-heavy systems where you are doing micro reads. And micro reads does not mean just very small byte levels up. I'm also talking about MBs. Right? When you're doing micro level reads, that is good enough. But when you're read-heavy system, this is not a read-heavy system. Blob storages are write-heavy systems and not read-heavy system. To do this, to get over this complexity, just create multiple versions of the file because hard disks are cheap.

You could creating multiple copies because files are not getting updated. ⁓ Files are completely written every time, which means that you create multiple copies and extract parallelism if you want to, rather than doing this. This is much more complex. Keep things simple. Okay. Right. But what you're suggesting is something that is doable for read heavy transactional systems. can like read heavy transactional blob storage systems. can do that.

⁓ I was kind of thinking in terms of Netflix streaming video streaming. So ⁓ So another question is that, ⁓ so what if ⁓ so in our system, ⁓ let's say one particular artist becomes hot. Okay, a lot of requests are being directed towards that only. ⁓ So ⁓ let's take an example of Netflix. ⁓ There is a movie one and movie two.

Sneha Mehra (03:13:16)  
there, there are they both are very frequent. mean, they're both are very popular and a lot of people are watching concurrently parallelly. ⁓ So ⁓ both are both of them movie decides in one heart is rack. Okay. So do we have to ⁓ implement such kind of system that, you know, balances the ⁓ I mean, balances this thing. Let's say if two of the movies becomes too hard,

and they are being, they should be shifted to another rack so that the ⁓ one ⁓ hardest rack shouldn't crash. You typically don't do that. Why? Because in real world, first of all, there is CDN to serve it and I'm not just basically ⁓ dodging the problem statement, but you typically don't do that in real world because you have CDNs to do that. But if you're still assume that let's say all requests coming in on to S3 to do it, right?

So in that case, S3 anyway has a limit on the number of requests you can make per access token or other per customer. It would not go beyond that. So you are playing in this limit. You are not saying that, hey, whatever request I get it all, I'll handle all of them because you are playing by the constraints that you have. You are a S3 designer. You can put the limit on the number of requests that you would handle. And now if other system wants to access you, you say that, this is our limit. We cannot go beyond that. You figure out another solution.

You don't have to add just to every single ⁓ request that is coming. I'm talking, I'm talking within the constraints only. Let's say there are 10k concurrent. ⁓ You would define the constraints such that it would never become. No, no, no, no, no, that's correct. I mean, ⁓ we have a constraint 10k ⁓ and we have, let's say 10 ⁓ racks. Okay. And one of the rack is serving like 7k or 8k, the 80 % of the load.

and other nine drags are ⁓ not serving that much. ⁓ do we have to implement such a mechanism limit per rack? That's how you would define it. You define per key limit per key limit. How many times can you access it in a minute?

Sneha Mehra (03:15:28)  
Okay. So S3, if you go to the specification, they have per key limit as well. That per key, can access these many times, like these many data ingress, it's in GBs, the dispatch of data you can take it from. And that is derived from the hardware limit that they have, the hard disk limit that they have. Right? Otherwise you can, if you want to ever build it, you put a cache in front of it, ⁓ which actually becomes your CDN, to be honest. Right? Those are other ways to build it, but you can extend the system the way you want.

But typically you don't cross see as soon as the hardware is added, you cannot go beyond that. Right? What are the hardware limit is, but if you'd want to move that as an exception, you can copy the data and move it to other hard disk and then serve it. But then you need to maintain that additional metadata, which says that where which lies because now you're breaking range based partitioning by moving data into some other place. Now that's the risk that you're taking. Right? So now that is additional metadata that you are storing not needed. ⁓

If they want to want to use it, they find other ways to do it. Got it? So you define the constraints such that your system doesn't break and your foundational rule that you started with, it doesn't break. But if that customer is big enough, you put a cache in front of it and cache all the data in some other storage and serve it as an exception, ⁓ but not as a general practice.

⁓ Right. ⁓ I think you are trying with this, ⁓ kind of things that you are putting it, ⁓ is big making S3 into a transactional database. ⁓ Because S3 is meant to serve different kinds of problems. ⁓ And you would define your own concept. If you go through S3 specification, you know, you would go to know the kind of constraints that it imposes. You cannot preach that.

Sneha Mehra (03:17:20)  
But ⁓ everything that you said, can make it as you can put the data into other hard disk and have a mapping that stores that had this file all this present in this track, but this file is not presented this track. So instead of going for range based partitioning, you go for range based partitioning for most cases, but for exceptional cases, you go for static partitioning. You say that this file is present in this part, this file is present in this part and this hard disk. ⁓ So you first check your static partition map. If it's there, good enough. If not, you go for range based partitioning and then you access the file. That's also a problem.

Hmm. Correct. So exceptional cases goes to static partitioning. You first check that if it's there, you go with that. If not, you fall back to rage based partitioning and just, and you just access the file the baby. ⁓ another thing is that, ⁓ can we talk some numbers over here? Like, ⁓ about a whole system, ⁓ like how many concurrent requests, ⁓ can it serve like, I'm not, ⁓ I don't know the numbers.

It depends on how fast you had this car. Let's brainstorm on this thing. ⁓ Just like we do it in an interview, we are given a problem statement. at all interview. ⁓ Go on. me a random question without saying interview. No, no. I'm not saying that. that, ⁓ I mean, ⁓ that so ⁓ why I'm asking is that ⁓ we would ⁓ like it will build our intuition like how much load Ask me a question. ⁓

Ask me questions.

How would you start? How would you start with that? Without you knowing the throughput of our hardtest, you cannot start with the number. What is the throughput of this hardtest?

Sneha Mehra (03:19:04)  
I don't know like what kind of model and all we are using exactly. Unless you put the load on it, you cannot tell. Right? You to the load on it. But this the bottom of this. This is the bottom of the approach, right? ⁓ The truth we are thinking from the lower level, like what's the throughput would hard this and then we are building up the numbers to the end. ⁓ Let's say we are ⁓ a poor end user, we are saying let's let's say we have a contract with

our user that we're going to solve this number of ⁓ without knowing a physical limitation, you cannot define a contract.

Sneha Mehra (03:19:43)  
You cannot just claim I would go for 1 million requests per second where your system cannot handle 100 requests per second, right? It's always whatever. It's never top down. It's never that I would want to handle this many requests and then I would define this. It's always what your physical hardware limitation is, which then propagates upwards. And then you say, okay, this is what we are guaranteeing. It's never top down. In storage systems where your underlying thing is storage, the hardware comes

Like hardware limitation comes at the cost. Right? You start with that and then you move up. So in lock so that you assume whatever throughput is, let's say 32 GB per second is what you're getting from your hard disk as a throughput. You start from there. Your hard disk is giving you 32 GB per second, which means you cannot go beyond that on that hard disk. 32 GB per second is what you start with, which means on an average file size is, let's say one GB. How many files you can read in one second? 32 files in one second.

And then you move upwards and then you define your boundaries. ⁓ Right? You start with it. Assume it's 32 GB per second. It's roughly, ⁓ it's basically the network boundary, but assume that it's GB per second because you anyway have to read it over network. Right? Start with that. And then you say how many files you can read pan parallel. Right? You may say let's say 32 files in parallel. 32 files in parallel, how many requests you can handle? So assume that your average request size is of let's say 100 MB. How many requests you can handle in parallel?

You do that division, division, division, you come up with that number. And then you know how many things you can serve in per second. Now multiply that with the amount of storage racks that you have, the number of distributions that you have. ⁓ That would be the guarantee that you would provide to your user. It's always portable. So 32 GBPS hardware limit, start with that. Then you can compute those numbers.

Okay. ⁓ But we also have to, this is a, ⁓ we also have to make, think about the network bandwidth between. ⁓ Okay. ⁓ Okay. I thought it was a SGT throughput. When, I put that because hard disk can read more, ⁓ but if you cannot send it over network, then what's the job of it? Correct? Correct. ⁓ That's why I went with 32 GBPS. Hard disk can read much faster than that. ⁓ I went with 32 GBPS or a high throughput network.

Sneha Mehra (03:22:05)  
because it is capped because it is bottle-legged by network. Correct? ⁓ That's why they do GBPS, said. ⁓ And also we have to take in consideration about the partition servers, ⁓ like each and every layer, ⁓ including the network bandwidth and the throughput of each and every system. Like how much time does it take for a network partition if there are concurrent requests of let's say 100k and so how do like

How do we build that intuition? ⁓ That's the division that you would do, right? Because now you have to define the access pattern of it that you cannot access these many things at once. You assume that, hey, this client would make at least 100 requests per second. ⁓ And then you cap it at that and see if you are able to sustain that load or not. ⁓ Assume like you have to make a lot of assumptions when you are doing this. ⁓ Start with that, play around with it, spend some time. We cannot do it live. Spend some time thinking about it. More than happy to discuss tomorrow. ⁓ Right? But go through that. Start with 32 GB.

per second as a limit of what you can read from your hard disk to your partition server. And then now everything is just compute. So now you can define what your per token access limit should be. We'll start with 32 GB per second. And then you define those constraints and how you are allowing people to access your data. How you are allowing people like what are files they can access, how per file limit, per bucket limit, per partition limit.

Right? Per access token limit. You define that because once you know your numbers, once you know your unit tech economics, then only you can define ⁓ what your users can access it. Yeah. 32 GBPS is a fair enough assumption to start because most intra data center high throughput network are 32 GBPS. Got it. Superb. Gaurav. Let's say, so ⁓

does is each partition assigned one rack or can one partition have multiple racks? think one partition one partition is small enough to be part of one rack. ⁓ But it can still span it can still span multiple racks. There is no limit to that because a partition is a logical partition of range is a range based partition. So what if I write a new file in that range, which is now going to another rack. So partition can span multiple racks, which is where a defragmentation comes in. It tries to it if it

Sneha Mehra (03:24:32)  
it would come back to single. ⁓

And will each vendor be given one whole rack? No, no, no, no, no, that's very inefficient. Very inefficient. ⁓ Imagine if you give one rack to every customer. Yeah. And they don't use it. Isolation. ⁓ Won't have vendor isolation. Or we just live with it. But it's small thing, right? When a customer, when a bucket becomes big enough, then you assign a separate track to that. That's where you do that migration of data. One time migration of data for that.

But to start with, you would start multiplexing them. ⁓ Once a certain thing grows beyond a certain scale, let's say more than 500 GB. ⁓ Let's say that's your threshold. Once a customer crosses 500 GB in a bucket, then you assign a separate rack for that bucket. ⁓ That too not separate rack, but you move it into a separate sense because you would have some buckets which contain small enough data. Everyone starts there and then moves to that other part. ⁓ But it still does not mean that one customer gets this one entire rack.

that would never happen because then AWS margin would drop down.

And I have one quick question on CDC. Can I ask it right now? So is the data application then to CDC itself? Let's say we use master slave architecture for our databases. ⁓ Is that done by a CDC kind of? Yes. Because the because true for some databases falls for other databases. Okay. Because in some cases, the database does not have been locked files or commit lock files. Then

Sneha Mehra (03:26:08)  
CDCs will iterate in the rows explicitly. Why? If some database gives you binlog file ⁓ replication, the default replication leverages those binlog file and commit log file to do the replication. ⁓ So CDCs is superset of replication. Okay. ⁓ So typically the way replication happens is by reading the binlog file and commit log file from the master on the replica. And a normal informative note that

One bucket name cannot be used by any other accounts. Yes. Yes. Yes. Yes. Yes. Yes. Bucket names are unique. ⁓ Like I didn't know that it was all over the world, all over the world. All over the world. It's unique. Bucket names are unique. ⁓ Thanks. Yeah. Anupam.

Yeah. So I was following the discussion around chunking and around read that since our system is not read every, ⁓ should not like chunking does not make much sense. But what about since the system is right? Do you think chunking makes sense for parallelizing rights? ⁓ long as users send you parallel rates, ⁓ if user itself, if user is not sending you parallel rights, you can't do much. Correct? ⁓ Let's say your chunk size is 128 MB.

user is not sending a parallel user is sending a sequential rights. You cannot leverage that correct? Yes, that's right. So you are assuming that user sends you parallel rights, which you can accept and then write in parallel. Then you can leverage it. And these are, these are supplements to the core feature. And which is where ⁓ here, when we talked about it, that your physical limit might be one GB, but your logical limit of a file can be thousand GB because you are implicitly chunking and storing and storing that meta information internally.

And then you can leverage. Hmm. Right. ⁓ to close the loop on that by user, like our client code should support that functionality of that as a feature to be able to go to. So if it can send you in parallel, you can accept it. But let them go to now. But then you are risking a random rights ⁓ on your storage site. So be wary of the fact that their random rights would come in and random rights. But because they would be close enough.

Sneha Mehra (03:28:24)  
So the head movement will not be massive. It would be a little bit, but it's enough that your throughput would go down on the storage. Yeah. Yeah. Makes sense. ⁓ And our crunching numbers on this part is always putting load and seeing how your system reacts. ⁓ Because unless we have built it, unless we have seen it, we cannot comment on it. So the way I would have done it is I would have built a system on a storage and I would have. ⁓

actually monitor those metrics on how it is behaving and then come up with that compass key. How like, if ⁓ what are features would I need to support? How much support would I be able to handle? Is it good enough benefit or not? And then we would take a call to release that feature.

So for a quick follow up on that, is it like possible to prototype this locally on our machine? ⁓ just add, just add, ⁓ just buy a Mac. ⁓ If you have SSDs, you will not see that big of a difference. But if you have hard disk magnetic disk storage, you can do that. But apart from that, apart from this, can anyway prototype S3 very easy. You want to implement checking, can implement checking. Storing file, you just store it on local hard disk path.

store that mapping, so that partition mapping and whatnot. Get very well, very well. Is that difficult to scale is different, but at least writing that hello world of history, get very well. Yeah, one month project. if you are, you'll have much deeper understanding post that what gets stored in partition map table. How are you moving data? How is modern complexion happening? ⁓ You can mimic all of that. ⁓ This is something I want to also try out like one month of time.

Go for it and go for it. Take one.

Sneha Mehra (03:30:13)  
⁓ Yeah. ⁓ So I was just thinking when you were having discussion on that numbers. ⁓ I just wanted to like this HDD is not that fast, right? It's not that, ⁓ did you say like, ⁓ like I was just, ⁓ understanding like with five milliseconds to 10 milliseconds latency, right? It can't go beyond 500 MBPS. Yeah. I'm giving 32 MBP like 32 GBP as a starting point.

⁓ see. drops down below that because why 32 Gbps? Because 32 makes calculation very easy. ⁓ Now the only parameter you need to change is that 32 like that part. No, ⁓ Yeah. Because I did not know that that was a random number. It's fine. No, ⁓ it's C. ⁓ So by the way, it's not very random. ⁓ 32 Gbps is the maximum that a network that a intra data center network can support. ⁓ Yes, I was going to get there. So that's the max. That's the max that you can get.

And which means if you change your hard disk to the fastest one, this will become the portal that the study to GBBs. And this is what I started with that. Now, today's HDD tomorrow, let's say you get the fastest hard disk out there or let's you store it in RAM. Right. Your RAM can give you much faster than that, but you're still bought a link by your network, which can send at max 32 GBBs. That's why I started with that. It's not random number.

It's the Mac. It's a physical limit of the connected network of the of the connecting network cable in your data center. It's that limit. So no matter how I'm not just restricting it to HDDs. Take SSDs, take RAM, take L1 cache. Let's have a distributed L1 cache the fastest that you can, but you're still bottoming by the 32GPPS. Right. So if you start with this assumption, change the underlying storage to as fast as you can, but you cannot go beyond 32GPPS.

It just falls back everything. Right. So you start with that and then you just change this number to see, okay, now I can get much less than this. That's right. And that's a good insight because if you are designing a storage system, our bottleneck should be storage, not network. Another question I had on the same line was ⁓ that ⁓ like, if we do those calculations right now, the hard disk is fine. Like my MAPS. ⁓ Let's go to the network.

Sneha Mehra (03:32:45)  
High level. Yeah. Yeah. So now our, yeah, our data hub is going to be most probably API to partition server to the rack, right? API. correct. ⁓ Yeah. Now our API server, let us suppose how many, like now the limit is going to be, what I was trying to get to, calculate those numbers that it is going to be the API server because partition server also has lesser requests. Correct. The throughput that

needs to be delivered needs to be delivered by the API server and client side is fine because there could be multiple clients ⁓ or the internet it could be less. But at the end of the day, our API. So basically I'm assuming the reason we are doing all of this is the network cables attached to rack to partition to API ⁓ are heavily like mostly optics, but the control plane can be normal network. is ⁓ okay. So that that are okay. I think your discussion helped me understand this. This leads to no peel. think first

If this is lightning fast, this would be very fast. ⁓ This you cannot control, ⁓ but whatever is within the data center, can assume that it's best in the, ⁓ it's basically best out there. ⁓ Right. ⁓ And then you build on top. ⁓ yeah. ⁓ Yeah. That, that, that, that made a lot of sense to me because in terms of like, it's very difficult to design this in terms of, so my, another question was going to be, you remember the Instagram discussion we had, we directly uploaded through that.

So my question was like in terms of security and all of this, how is this handled? Because if you through the data, ⁓ now, how do you make sure that you are still guaranteeing the public cloud security and still your data like you'll have to expose at minimum servers out to the public, right? ⁓ That's what I'm confused here. Like most probably like our control, clear requests were fine. We could go through DNS and all of like all the proxies and authentication servers and everything.

But how do you make sure ⁓ someone who is pushing data can be authenticated on every request? Okay, that is where ⁓ the S3 signed URL that we're talking about. The signed URL is nothing but a certificate. So when you get a request like that, you just validate the certificate over here. You don't need to make a database call. It's just a validation of certificate. Right? You just do a quick validation of certificate, basically decrypted and check the validity, check if the signature matches or not. Once you

Sneha Mehra (03:35:07)  
to that, that's good enough for you to let the request go through the partition server and whatnot. So I mean, I can still think of a hacks for man in the middle for this, but the private key is with the customer. ⁓ can only validate a certificate. Okay. It's not, not just the link. It's the whole private key, right? So the customer to do that. So it's part of the SDK that it ships with. So private keys with the customer to

Basically that ends very short-lived. So someone having that, would be very difficult to bypass it. So that's the signature that people got it. And another point I wanted to ask you is which maybe we have not covered it in like, it's fine to ignore the question if you're going to cover it ⁓ but about the compression and encryption, the things like this. Now I like when we were discussing this hot, cold, warm storage, ⁓ how does compression encryption and those things play in the picture? Because

One normal question I had was when we had this whole data engineering pipeline, ⁓ I was assuming when you do joins and like basically if you're doing in JSON formats, you're duplicating data now. So your warm storage is going to have a larger space than your hot storage and same goes for cold, but it's fine. Like it's about cost and queryable. Correct. The compression thing you talk about will not go for out of the box compression. The format that we typically go for is a columnar format when we store this data. Right. Okay.

that format, example, Parquet if you use Parquet natively compresses the data and stores it. ⁓ You leverage the format and use those things out of the box. don't rely on compression done by S3. It typically does not do any compression there. You rely on your site, the format in which you are writing to the staging ⁓ S3 can be a columnar format like Parquet, which internally compresses the data and stores it. That makes sense. Okay. Okay. Okay.

Yeah, because now this whole public cloud thing, I'm also assuming that the whole data that resides on all any of those disks are encrypted because now such huge data as also CPU overhead. So this API servers, which are going to do that. It's not just network and disk now. And encryption is very common, but every bucket that you create on S3, can provide a, you can provide a custom encryption key. Exactly. Yeah. That's, that's the trickiest part, right? Yeah. So encryption decryption.

Sneha Mehra (03:37:30)  
decryption, decoding and then sending it to the user. ⁓ Encryption, and storing it on storage. Yeah. And I was reading about S3 throughput. It seems like they can give GBPS and that's where I was wondering like giving GBPS. Now ⁓ it means, ⁓ what I'm thinking is it definitely means they're doing parallel writes because that cannot come through a single disk. ⁓

It is also possible that they accept the rights but then write it as synchronous way. That is also possible. ⁓ They can also chunk and do it. So it's more about how quickly they can accept the ⁓ rights. There may be tiering as well. There ⁓ are so many complexities that are involved. depending on how S3, because S3 has now basically 15 years of evolution.

They started with something with time progressing. are making changes to their features and all. There will be tons that will be coming and it's not as simple as I do, but you get this idea now. No, no. Yeah. I'm discussing just for the same thing ⁓ to get the general idea of things, right? because once you try to tweak one, there are three things that's come out, which are problems. Thank you. Yeah. Yeah. Rajiv. on. ⁓

⁓ you can go to the diagram ⁓ and open that data pipeline is something which I wanted to know what, I I am not very familiar about. I'll explain now what data pipeline is. want to know about data pipeline? Yes. Yes. Yeah. Why it is used. ⁓ Okay. Data pipeline is a term that people use, which calls it, we take data from one place, put it into another. In reality, it is just a spark job, which is running, which is reading data from one place and putting it into another. ⁓

process is called as pipeline. It doesn't mean we are automatically fetching it or putting it there. It's literally a Spark job that you write a query that you need to fire on this database. You write a query through which you'd want to store this into this database and a Spark job runs and does that for you. That is what a data pipeline is. Data pipeline is basically a simple process that reads data from one place, modifies data in some other place, basically transforms the data and puts it into different places. That is ETL, extract, transform,

Sneha Mehra (03:39:52)  
and load. That's what our ETL pipeline or data pipeline is all about. Throughput also increases like multiple reads, transforms. That is features. That depends on how you write your Spark job. Throughput, parallelism, whatever you want to get it, that is how you write your Spark job. But ⁓ the core idea is you extract, you transform and you load. Okay. Understood. Right. So whenever someone says data pipeline, it's literally something, a process that connects two data sources.

takes data from one place, have apply some transformation store. Yeah. mean, spark job. mean, I have read about map reduce, ⁓ and all those things. So spark is similar to map reduce or how would get linked spark job is similar to map reduce like that's a oversimplified spark job is similar to map reduce where it distributes the load across multiple machines. does that cluster management for map and reduce distributed computing for map reduces different spark is spark does it spark does it in memory?

⁓ It still moves the data across, basically shuffles the data and whatnot, but it basically abstracts out distributed computation for you. For example, if you're reading one terabyte big table, ⁓ one node will not be able to process that much of data. Let's say you're reading two tables, joining them, both tables are two terabytes. Two terabytes RAM you have on your machine? No, you need to distribute it. Right? So the jobs, the way the job would be distributed, each server would pick up a small fragment of data from this table, from this table would

join it and put it there. This sort of distributed computation and then some job would take all of them and basically combine it and then serve it and then load it into another data. This abstraction is done by Spark. ⁓ The distributed computing part, Spark takes care of that. I want to read more about this. mean any good resource you can... Orally book on Spark is good one. ⁓ You told once yeah. Yeah that one. I put the link in discord as well. In February 2023\.

cohort. I put the link in, I think it's called effective spark or something. read that to build an understanding. That's a good one. One book was there. Yes. I saw. Yeah. Yeah. Okay. That's the one I referred that I, and more importantly, don't just read, code it out unless you put it out, you would have no idea what's happening. So code it out to get a deeper understanding. ⁓ It's fun. ⁓ world is very cool. The storage by default is history, right? I mean, when you say cool storage is history. ⁓

Sneha Mehra (03:42:22)  
S3 or any blob story. You mentioned the term glacier also right? So what is glacier here? In S3 there is multiple tiers. ⁓ In S3 there is an infrequent, ⁓ there is a standard tier, ⁓ there is an infrequent access and there is a glacier tier. So it's one, one is more cheaper than the other. Glacier is the cheapest one. It gives you very slow reads as compared to standard tier. ⁓ Infrequent access is in middle and glacier is even slower than that. And you also have a policies for that as well where you want to go. Yeah, where you move it. So on S3 you can configure that.

I'm so happy you are listening. ⁓ I just tried to make note. mean, it's very difficult to remember all those things, sir. Three, three, three, what to do when system design is very vast. ⁓ We can't have, I mean, you are on the floor, right? And you, have some questions. So it's very difficult to go back to the, don't write questions. I just listen because if I write question, I just break the flow of mind this thing. Right. ⁓ So, ⁓ here's the mean gap, which is there. I mean, I have to write things, but you should write, you should take notes always like

at your job also just keep taking out it does not mean descriptive notes even a word in a small question even a small concept just make a note of it it's very helpful so that you don't ⁓ have to put all things in your cache you're flushing it onto a disk and reading it from there ⁓ yeah got fine great thanks anyone else any other question? no how many survived? 11 folks super including me 10 folks survived thanks so much

Thanks so much. It's quite late. Well, 12.45 already. But ⁓ I hope you found this discussion amusing. I tried my best to cover it in depth, answer all the questions. But in case there is any, feel free to ask me anytime in the course, one on one on Discord, wherever. More than happy to answer. I'll upload the recording by evening. Thanks so much. One thing more, let's say after the course is over, one on one is also over. you want to ask and then... Discord, Discord is the Discord or email. Yeah. Async. ⁓

Yeah. Thanks folks. I'll upload the recording by evening. I'll see you tomorrow. Bye bye.

—-------  
6

Sneha Mehra (00:00:02)  
The third day, second, it's all about distributed ID generators. Like ⁓ this is one of the most misunderstood topic out there. ⁓ auto increment works giving ID. Why is this such a big problem? ⁓ Most people don't really face this unless they hit the scale. But nowadays because of the mass internet penetration that we have.

Most companies hit that scale sooner than they think. ⁓ So which is where distributed ID generators really come into the picture. That's one. So that's one reason why you should know them in depth. Second ⁓ is there's some beauty of computer science going behind the scenes. You just fall in love on how amazing engineers are in the device solutions like this. is surely going to blow your mind. Third, when

We understand ID generation. We see ⁓ some really interesting side effects of that, which most people didn't even think about. ⁓ So we'll touch upon all of that today. ⁓ And what I'll do in first one, I'll take you through the journey ⁓ on basics of ID generation. ⁓ And through that, we build or we formulate some really foundational things that we should know about. And then we take a look at

How Amazon does it? ⁓ sorry. How Amazon does it? How ⁓ Flickr does it? How Twitter does it? How Instagram does it? Right? And implement every single thing. Without implement, and trust me, there are people from PasscodeSwap, ⁓ once they implemented, they got much deeper understanding. Theoretically, it's all okay. But implement every single one of these to get the deep understanding, like why that works, because it...

There are such beautiful things out there. ⁓ And I want all of you to, you know, ⁓ experience that same joy, ⁓ obviously early engineers, ⁓ that is very different, but at least you get that kick of you knowing those things really in depth. ⁓ Okay. So we'll do that. ⁓ So we start from really simple and we touch upon some really interesting concepts. ⁓ doing this, I'll show you a very interesting demo, ⁓ not really a demo, ⁓ but some really

Sneha Mehra (00:02:29)  
some really interesting side effects of that. Okay. So let's start with this. So ID generation, first of all, IDs, IDs we all just use so often, like basically identity of something. We use ID word very often. it's more about, if I may want to put it in just one line, it is more about giving uniqueness to a particular row, document, object, event, whatnot. Anything that would help you uniquely ⁓ identify something that is its ID.

So what our problem statement is, that assign a globally unique ID to anything. So whoever asks for an ID, we assign it to them. ⁓ Now here what we do is we write a function that spits out something unique every time it is invoked. So that's what we have to do. As simple as when someone wants it, we invoke a function that function to spit out something unique every time. ⁓

Now here, the function we are writing is not a central service. Don't directly jump into that. It's a function which is part of your business logic and not a central service. A lot of people are directly, I'll start writing a new service. Don't think of that. What we are right now being constrained by, because this is where the foundation, like this is what matters when you are building a foundation, is we are just writing a function.

whose job is to spit out something new every time it is input as simple as that. ⁓ I'll take you through that evolution so that you understand it really well. ⁓ So it is part of our application logic and not a central ID generation service. So when I get a request, I myself am generating an ID by invoking a function call locally function call. I get that idea and I use it. That's what we are doing at the moment. Then we'll build a central ID generation service kind of thing, whatever it's at.

⁓ Okay. So here, the first hint that we get is because we want to write a function that generates something unique every time it is involved. We know time is unique. We know time always moves forward. So why can't our ⁓ ID generation function just return my current epoch millisecond? And that could be the, that could be the simplest starting point that we can think about that because we want to generate ID.

Sneha Mehra (00:04:57)  
Every time a function is invoked, know time always moves forward. So let's just, whenever someone asks for an ID, let me just generate an epoch millisecond, which is the current timestamp in millisecond. We send it. ⁓ That ⁓ should work fine, but obviously something would break because if it was the case, then everyone would be using this. ⁓ So here we are taking epoch millisecond.

which means we are okay with the millisecond level granularity, which means this function cannot be invoked twice within the same millisecond. Otherwise, you have to go to microsecond and nanosecond level. So now what we do ⁓ is that's okay. Let's start with that. With that constraint, it's fine. So what if there are multiple machines? Our function for ID generation was as simple as generating epoch millisecond and just returning it. What if there are multiple machines?

Because if we even assume that a single machine, is not possible for our use case to invoke the same function twice within the same millisecond, it is very much possible I will have multiple machines in which that same method is getting executed within the same millisecond. ⁓ So when that happens, what would happen? Collision. Because both the machine at that very same time, let's say 1729 is the epoch minutes, and just using a smaller number over here.

But imagine it's a epoch millisecond. When this function is getting invoked at the very same millisecond on those two machines, what would happen? There would be collision. Both would generate the same ID because what we were just returning was 1729, the current epoch millisecond. So which means it is not globally unique. So when you have multiple machines, one thing that you can do in order to resolve this collision ⁓ is to concatenate machine ID

and apoc.md.second. So machine, which machine generated which ID that becomes your ID. So this is the function that would be executed on both the machines now. So when a request comes to a particular machine, it invokes this function and the ID generates is not just 1729, but m1-1729. And I'm just using some nomenclature. You can change it however you want. But the ID remains the same m1-1729 and m2-1729. So here what is happening is

Sneha Mehra (00:07:21)  
We resolved the collision. Now we see ⁓ globally unique IDs. This is m1-172-0, this is m2-172. So that is a very common strategy. Whenever you see a collision, you try to add something that led to that collision in order to resolve the collision. It's a very standard practice. ⁓ So here we saw ⁓ IDs are basic, primitive ID generation logic colliding. It collided because we had two different machines executing at the same time.

So in order to resolve the collision or to handle this conflict, we added machine ID because machine ID led to that. our ID generation logic now changes to machine ID and get a pop millisecond concatenation. ⁓ Okay, this works fine. But what if our program has threads? ⁓ If our program has threads, which means that within that same machine, my function can be executed. ⁓

That same function can be executed multiple times ⁓ within the same millisecond. Which means if I have two threads, let's say t1, t2 and m1 and t1 and t2. These are typically integer thread IDs which the operating system assigns to the thread. ⁓ I'm just using t1, t2, m1, m2 to simplify the explanation. So here what would happen is it is very much possible. It is very much possible that your program has multiple threads running on multiple cores.

which would invoke this exact same function within the same millisecond. So now there is again a chance of collision because two threads executing at the same time within the same machine would spit out that exact same ID m1-1729. We don't want that. ⁓ So what do we do? We have two approaches. First approach is like how we added a collision resolution thing by attaching or by concatenating machine ID. We do something similar by attaching thread ID.

So what we do is we write our ID generation logic that looks something like this. Concatenate machine ID thread ID and get a pop milliseconds. So when my ID generation function is now invoked, it returns m1-t1-1729, m1-t2-1729. There might be other thread IDs on m2, which is p1 and t3. So m2-t1-1729.

Sneha Mehra (00:09:44)  
and m2, 3, 1, are the actual IDs that are that will be used for those rows and columns and what not. ⁓ That is one way to do it. Other way to do it ⁓ is by adding a counter that resets every few minutes. It's a very common practice. instead of you always adding a thing that ⁓ in order to resolve the conflict, you always adding the thing that led to conflict.

leads to gigantic that which leads to basically gigantic IDs. What you can instead do, you can go for a static counter. So what is a static counter? Two threads share a RAM. Static counter is globally allocated variable on heap. Everyone can access that. So both the threads can access that. So what you can do is you can atomically increment that counter every time the function is invoked and that becomes a resolution logic. So

And here what we do is we globally allocate a counter variable, get ID function, in all this concatenation of machine ID, timestamp and a static counter. This static counter becomes plus plus every single time the function is invoked, something like this. And as if this is an atomic increment, I did not write that locks, works, mutex and all, but as if this is an atomic increment, like the increment is, and it's safe. Atomic increment.

So what we get is every time the function is in both, ⁓ my counter will become plus plus plus plus plus plus plus plus plus. So which means I would never see collision within the same machine because it's atomic increment. So no matter how many threads I have, increment will always be atomic and it would be fine. ⁓ So no matter which thread handles it, my ID would look something like this. M1-1729-01, ⁓ M1-1729-02, M1-1729-03.

as my millisecond moves forward, ⁓ becomes m1-1730-0404 and so on and so forth. Every few minutes, you can reset the counter or every few seconds you can reset the counter. That's your business logic. ⁓ But the idea is you would still have something unique because now with every millisecond this is anyway moving forward, you can choose to reset this at any point in time. That's fine. So long as you don't exit the integer limit. But you see how we...

Sneha Mehra (00:12:12)  
now got two different ways to resolve collisions when it comes to your program having multiple threads. ⁓ One is by appending thread ID and second is by using a static counter. So static counter no matter how many threads you have it would automatically atomically increment that's what this assumes. ⁓ But if you look carefully here given that the counter

is atomically increasing time is redundant time plays no role because every time your function is getting invoked you are always doing count plus plus atomically so no matter how many threads you have because the increment is atomic it would always be one two three four five six seven eight nine ten and so on and so forth right so time is really not playing any part

So can we remove that? We for sure can. Because now what would happen is our IDs would become shorter. So ⁓ another ID generation logic might look something like this. We have a global counter variable. GetID function does a concatenation of machine ID and counter, that's it. So wherever the request goes to a particular machine, it is handled by one of the threads. And because the count plus plus

is atomic like we will make it atomic you don't like it's not by default atomic you wrap it in locks and whatnot or do an atom or use a native atomic library to do that but the idea is ⁓ assume an atomic increment of countless of the variable cow and you get ids like m101, m102, m103 and so on and so forth this kind of solves the problem but what it has a limitation on that when it hits the integer limit that's a four billion

or if it's true it is to 64, that is a massive number, you might never hit that particular element. But still, please end it up. You would still write a good ID generation because of this. And it is unique, whatever you wanted, you get it. ⁓ Machine ID concatenated with a static counter, which is atomically increment. Now this solves the problem. But, okay, what next? This is good. So we saw multiple ways.

Sneha Mehra (00:14:33)  
of or multiple common ways to build this ID generation or to implement this ID generation. Now what's next? Now here we look at one very interesting factor, which is this static counter. IntCounter, this is a globally allocated variable. Now what is happening is you have a static counter. This is what your logic is, ID generation logic. The problem with this is what if your machine reboots?

As soon as your machine reboots, what would happen? This counter would again be reset to zero because that's what would happen. ⁓ There is no layer of persistence at all over here. There is nothing. So when my process reboots or when my machine reboots, my counter will again start from zero. And when the request comes in, it would again start generating M101, M102, M103, M104, which means that

My ID generation logic is not unique because when my machine reports, it starts regenerating the IDs that it did after the previous reboot.hub. So how do you solve it? The root cause of this problem is that

the place at which we are storing the counter is actually a volatile memory. ⁓ So how do we solve it? Very quick answer. ⁓ Alok.

Back it up with the file locally. Okay. You write it to the file locally, right? Because ⁓ volatility is the problem. So you start writing it to the file. So it looks something like this. So what we are doing over here is we store the counter to the disk every time. And when my process starts, I load it from the disk, which means my code would look something like this. ⁓ So here what we have ⁓ is our counter is equal to a load function. We basically assume that it loads from the disk.

Sneha Mehra (00:16:31)  
I get ID function, if I first do count plus plus, I save the counter to the disk. I concatenate with machine ID and counter. This way, what we did ⁓ is we made, we started storing the counter on the disk and not in some volatile memory. So here, whenever my program starts, I first load the counter like this, wherever a function is invoked, I do count plus plus and immediately save the counter to the disk.

Maybe in a text file, I'm just dumping the counter every time opening the file, writing it there. ⁓ And my business logic remains the same. ⁓ But here there is a problem. ⁓ That in here, what we are doing is when my function or when my machine reboots, when my process reboots, I'm doing a disk IO, which is fairly okay. ⁓ One time at startup of a process or boot up of a machine, it's fine. But every time that my ID function is getting invoked, I am doing a disk IO.

that would put first of all, I would put a lot of pressure on the disk 100 % because if this is a high throughput thing, if it is getting invoked at a very high pace, it would lead to so many disk IOs, ⁓ micro disk IOs happening on the disk. ⁓ Problem number one. ⁓ Problem number two is it might reduce the latency that you are looking at because you are doing disk IOs every time or you were just serving it from memory. So how do you solve this?

We making massive micro discolours. How do you mitigate this problem?

Check.

Sneha Mehra (00:18:05)  
⁓ How about ⁓ adding another counter to the ID and you basically pick this counter up every time you start up the server and then immediately incremented and write it back to five. you'll pick up a solid problem of frequent disk IOS. You do it once at server startup only and then use that ⁓ counter as part of as a third part of this.

I know you did that already. You loaded it from the desk, but now whenever you are generating this ID, every time you are storing it on the desk, correct? This is no, I mean, mean a third part in the ID apart from this counter we are already using. No, but this is, but you're still doing same counter every time. Right? No, it's just once on server startup. How would each service startup is represented by. ⁓

a new counter. How will this solve your problem? ⁓ okay. So every time a server starts, that's a new counter and then let your static counter start from zero. Yeah. Let us take out a start from zero and that booting up. Okay. Okay. That's a, that's a good one. That's a good one. That's a good one. But still, so then you will not be storing this at all. You will be storing. You'll ought to basically you're not changing over here or changing over here. So you're not storing anything over here.

Yeah, no, that wouldn't be needed. But then what's the downside of this?

Sneha Mehra (00:19:42)  
When you increment ad server startup that will have to be atomic. That's right. That can be made atomic. ⁓ A longer ID, longer than we have built here. Longer ID. Right? Yeah. Now, are we so particular about it? We'll talk about disadvantages of having long ID, but your approach is really good. But that ⁓ long ID...

might become a problem. Not as big, but your approach is really nice. I loved it. I loved your approach. Thanks.

I was just thinking of that case, that would also not be that brilliant approach. I liked it like every time because you are now actually tackling this problem head on that instead of saving it every time, you're just starting a third part. You're adding a third part to this. Brilliant. Very interesting. ⁓ Very interesting. And this is something I will definitely be adding to this going forward. Thanks. Thanks. Thanks. Thanks for that. Okay. ⁓ harsh.

What else can we do to minimize this? ⁓ think it is something similar that I was thinking. ⁓ Maybe we can have some bank of IDs. For example, first time I read it is one just add a thousand to it and store it. So once I reach two thousand, ⁓ then again, read it and increase it by a thousand. So that thousand iteration we can do it in memory. ⁓ Got it.

So what you're effectively doing is you are minimizing the number of times you have to write to the disk by writing in a periodic manner. ⁓ Instead of writing it every time, ⁓ you're writing it at a certain frequency. ⁓ And this way you are slashing down the number of disk IOs that you have to do. ⁓ Okay. Let me walk you through on what, ⁓ and I'll tell you why I'm a little bit stressing on this because this comes in handy at two different places.

Sneha Mehra (00:21:43)  
but I still love the first approach where every time the machine starts, you start with a third thing. That's a really good one, really good one. Okay. So ⁓ what we do ⁓ is instead of, we, let's say, we, let's say, define a flush frequency. Now this flush frequency would be something like this, that instead of writing, instead of saving counter every single time, what we do is we define a flush frequency, let's say every thousand times, every thousand times.

we write like after generating every thousand IDs, ⁓ I'll write it to the disk. I'll write my counter to the disk. So we are literally slashing a disk I by a factor of thousand, which is a massive performance boost. But what we just have to take care of, ⁓ instead of just loading from the last point, I have to just add the plus frequency over here so that we don't get any collision at all. ⁓ So our logic now becomes counter equal to load plus the plus frequency. Let's say every thousand IDs we are writing, ⁓ do count plus plus.

counter mod equal to 1000 equal to zero, then you do write on the disk. ⁓ So we are not doing every time we are writing to the disk, but every thousandth time we are writing to the disk. Now this way, now just to give you how this would look like, just to explain how this looks like, let's say hypothetically, my flush frequency is three. ⁓ So every third value, I flush to the disk. What? My counter. ⁓ So my machine ID is M1.

My counter is 1, 2, it's arrow 1, 2, 3, 5, 6, 7, 9, So M1 happened when it created disk, generated disk, generated disk. When it generated disk, your disk invoked 3 mod 3 equal to 0\. Yes. So I saved the counter. So it flushes the counter to the disk. So 0, 3 got written on the disk. Then the next iteration, now rather next time the function call happened, 4, 5, 6, 6 paid also started, wrote it to the disk. Right. Then 7, 8\.

and then the machine crashed. When the machine crashed, the last value that was written was six. The ID that was generated was already eight. So now when my machine reboots, if I start from six, if I see that, hey, six is written, if I start from this place, I would be generating seven, eight again, which is wrong because it was already used by something. Right? So the safest value that you can start with is the last value that was flushed.

Sneha Mehra (00:24:08)  
plus the flush frequency plus one. So nine plus one, 10\. This is the safest value that you can start with. So M1 happened 10, M1 happened 11, M1 happened 12 is what and then 12 you do flush again. This way, what happens is what you are trying to do ⁓ is you're trying to ensure by adding this small like by leaking or by not utilizing every single integer out there, but

My dude, because here we never used 8 and 9\. We directly starting with 10\. We lost those two values, but that's so fine. That's okay. Given that the correctness of the system matters more, then you utilizing every single integer out there. ⁓ So always remember, whenever you are trying to do something with this, think of correctness. Think of the safest value that you want to start with. So key takeaway from this ⁓ is first,

not doing IO every time so you can add a periodicity factor. Now don't think of a thread and timer and I'll do it. Think of very simple approach as simple as a mod operation. So n mod this then you write it. ⁓ And second is whenever your machine recovers quote unquote you start from the safest value. This becomes the key design decisions or key design constraints when we are deciding a distributed ID generator. ⁓ Okay, now ⁓

This is okay. ⁓ Now let's move to some other constraint. ⁓ Let's say instead of just having ID, if you look carefully, these IDs have no particular order like M1 can be literally anything. ⁓ I have hundreds of thousands of machines. ⁓ It like there is no particular order in which we are generating an ID. ⁓ What if I want to enforce and a monotonically increasing basic constraint to my IDs? ⁓ So what we want

is no matter how many machines I have in my infra, no matter how many machines on which my IDs are getting generated, where the request goes, the IDs that should be generating, should be monotonically increasing globally, monotonically increasing IDs. So even if I have 10 machines, the ID like first request should always get ID, second request should get ID, third request should get ID, let's say five, next request should get ID six, no matter which machine it goes to. Right? Now let's say if we want

Sneha Mehra (00:26:31)  
monotonically increasing IDs. will say again, first of all, why someone would need this kind of monotonicity? I'll give you a practical example for that. Database ⁓ in ⁓ second week when we were building airline checking system, what we did ⁓ is we saw importance of locking. ⁓ We always said, Hey, we have two transactions, transaction T1 and transaction T2. ⁓ then we said that hey, transaction T1 ran this query and then transaction T2 ran this query. How did we know?

The transaction T1 came before transaction T2. ⁓ 1 and 2\. This 1 and 2 because they are numbers they tell you which one came first, which one came second. It helps you in resolving conflicts. ⁓ For example, ⁓ when T1 wrote something, T2 wrote something, ⁓ one of the conflict strategy would be the newer right overrides the older right. Now this 1 and 2 way it tells you that 2 is the newer right. So you can

actually accept during the conflict resolution strategy. You can accept the right made by the newer transaction. ⁓ So monotonically increasing IDs play a very important role in doing conflict resolution, which is where we need this. ⁓ And that is one of the places, one of the places, right? Not all the way, but one of the places. So now what do we do is here in order for us to get monotonically increasing IDs. Now, first of all, monotonically increasing IDs are

which increases monotonically does not mean they are purely sequential like 1, 2, 3, 4, 5, 6 is strictly monotonically increasing IDs. But ⁓ 1, 2, 5, 7, 8, 9, 32, 33, 34 that is monotonically increasing which means that my next ID that is generated is guaranteed to be greater than all the previous IDs that are generated. That is monotonically increasing IDs. ⁓

Now if we were to have this as a feature in our thing, what we can do is our logic that we wrote earlier, where we concatenated machine ID and time, we did machine ID and time. So m1-1729. Given that the time always moves forward. It is always increasing. What if we flip the order? What if we flip the order?

Sneha Mehra (00:28:58)  
and have concatenation of time and then machine ID. Given that the time always moves forward, if I do it at millisecond level, every millisecond I would have a new value that I'm writing, it is always moving forward. These are the most significant bits ⁓ of my number. Machine ID can be literally anything. This way what I get is I get monotonicity.

It is not strict monotonicity but it is pretty decent monotonicity that I am getting. Pretty decent. ⁓ It's not strict. To some extent it would falter but it's fine. So what we are getting over here is something like this. This is what my ID would look like. So this is my time. This is my machine ID. So 17290001, 17290002\. These are machine IDs. You can also change this to counters. Perfectly fine.

But the idea is putting time on the left most side gives you this natural order to IDs may not be very may not be ⁓ always monotonically increasing, but to a very good extent, it is monotonically increasing. And because it is possible that within the same millisecond, let's say these are machine IDs. Hypothetically, let's say these are machine IDs. It is very much possible that the first request within the same millisecond went to machine ID to

And then the next request went to machine ID 1\. So in that case, what you get is 17290002 first and then 17290001\. But within that millisecond, it would happen. But ⁓ as soon as millisecond passes by, you get almost monotonically increasing IDs across your system. Few things here and there is okay. So you are not ⁓ worrying about like here you see that it is still stateless.

Request and go to any machine because every machine is just doing the time concatenated with machine ID. Assume machine ID is integer and then it's just bits. What you are getting is you are almost getting monotonicity. Almost. There might be incorrectness in the system. That's fine. Without having to maintain a central state of things in order to get those things like strictly monotonically increasing. But getting pretty decent throughput. Pretty decent. Excellent throughput.

Sneha Mehra (00:31:20)  
at very minimal cost. But here you see how we beautifully put time on the left most side of things to get almost monotonically increasing IDs. Because as your time moves forward, the number would always be greater because it is ⁓ 17,290,001 17,290,002 17,300,000 ⁓ You see this is always

going to be increasing because every millisecond this is moving forward. So this is a very key highlight on why time is always on the left most side. You would see a ton of ID generators. You would always see time being on the left most side of your ID generation logic. Pick any. You would always see time sitting on the left most side. This is the reason why. Now, ⁓

Although it's a good solution, but it still doesn't guarantee monotonicity. ⁓ One of the reasons that we talked about was that it's not guaranteeing monotonicity. ⁓ But at least within milliseconds you have a fair bit of monotonicity. ⁓ But there is another thing, ⁓ another problem that would crop up is here what we talked about was the part where this

Like the first request went to machine 92, then the next request went to machine 91\. So my IDs will be 172902 and then the next ID will be 172901\. So that is wrong. But there is a problem even on the time side of things. ⁓ If we even let go the part that it's okay that within a millisecond, I might have a little bit of miss ordering, but at seconds or at time level, it should not be there. ⁓ So here what would happen is because our

ID generation logic has time component ⁓ and machine ID component in that. Here we up until now we assume that the time across all of my machines is exactly synchronized. Perfectly synchronized, but basically perfectly synchronized, which is not the case. I'll tell you about. So clocks go here and there. It's very common.

Sneha Mehra (00:33:38)  
If you have worked in SRE team or DevOps team, you know you have a process that is running, is continuously seeking NTP clocks of your servers. It's very common for your servers to go out of sync. So what do you do is you have a job running job as in a normal process running, which periodically syncs your machine clock to NTP servers. Right? So here what happens is, and first of all, why does this, why do clocks go out of sync? That's a very fascinating thing. So

The clock that ticks on your machine is a hardware clock. Although we see this digital time, it's actually a hardware clock. There's a quartz crystal vibrating in your machine, in your laptop, on your server, in your CPU. It's vibrating. The vibrations are measured and that tells you that, one second has passed or one millisecond has passed because it vibrates at a very particular frequency. Given that, given that it's a physical movement,

that you are measuring in order to see the time in order to measure the time. What happens is that because it's a physical movement, we learned in physics that when you throw energy at a thing, it vibrates faster. So when you are having a water, when you are having water, you provided energy with fire, the water vibrates and it starts boiling. ⁓ And when you suck the energy out of it, you take the temperature down.

the water freezes, it stops moving, it becomes ice. It holds true for every single element in the world. With Quartz Crystal, what happens is, ⁓ if in most cases, when your machine, let's say you are playing a very ⁓ high intensity game which requires a lot of CPU power, your CPU heats up. This energy ⁓ affects the vibration of the Quartz Crystal. Quartz Crystal would start to vibrate faster and your clock would move faster.

If you're putting your machine in an extremely cold environment, the vibration of Fox crystal would be slower, so your time would move slower. Physical phenomena affects the vibration of crystal, which slows down or makes your ⁓ computer clock go faster. Which is why there is a need of a periodic job that keeps the clock in sync and it runs more frequently than you think. It's really fascinating how

Sneha Mehra (00:36:04)  
physics and this thing works together, right? You need to take care of these all sorts of phenomena when it comes to this. There is a very famous ⁓ article in case you have not read it, ⁓ Jeff Dean and Sanjay Himavat of Google, the friend, just Google, friendship that made Google special. ⁓ Most of you ⁓ should know it in case you don't, just search for this article, friendship that made Google special or something like this, in which Sanjay Himavat and Jeff Dean were facing a bug.

which happened because ⁓ a bit got flipped because of an explosion or something or also because of a solar storm that happened. It flipped a bit in one of the data and they were scratching their heads into understanding that why that even happened. How did this bit got flipped automatically? It was a natural phenomena that caused the bit to flip. You cannot run away from nature. Right? So just Google, ⁓

the friendship that made Google special Jeff Din and Sanjay Ghamavad brilliant folks like I just love the papers that they wrote ⁓ Jeff Din and Sanjay Ghamavad just read about them in case you are are hearing the name for the first time. physical phenomena affecting this thing right so even if you assume that our clocks will periodically eventually be synced with NTP there up until then there is a chance that your clock is out of sync when your ID is getting generated

So hypothetically, ⁓ let's say I have four machines, ⁓ machine ID 2, machine ID 4, machine ID 7, machine ID 9\. And the time on those machines is 23, 24, 23, 23\. It's a pockmarked second, I'm just using smaller numbers over here. Now here what would happen is, ⁓ let's say my request first went to machine ID 2, then it went to machine ID 4, then it went to machine ID 7, and then it went to machine ID 9\. ⁓ Now when I fire getID function, what we are doing is we are concatenating machine

like the time and the machine ID. ⁓ So the ID that would be generated on M2 would be 232\. Then the next ID would be generated would be 244\. ⁓ Then 237 and then 239\. If you look carefully, this is not motor-totically increasing. The first ID that is 232, ⁓ then 244, then suddenly it's 237 and then 239\. ⁓

Sneha Mehra (00:38:30)  
you were expecting monotonicity, but you don't have it. Right? So this is a problem where you need monotonicity. Even the clock skewness is what you need to take care of. Right? Somehow, somehow we'll basically take a look at how people do it. But this is actual thing that affects the in quote unquote, if that is the correctness of our system, it affects the correctness of the system. Right? Okay. So now what we understand by this is

ensuring monotonicity across distributed systems is almost impossible. Almost impossible. You might still see some strands here and there. I did not find a theoretical proof of that, but you can very clearly see practically there are so many things that could break things that you cannot guarantee strict monotonicity in a distributed system, right? Where your ⁓ ID generation logic is distributed. Right? So what do you do? Let's say

Instead of putting it on independent machines, let's have a central ID generation service because there is one machine. ⁓ whoever basically idea is simple, ⁓ whoever wants an ID talks to this machine. This machine generates an ID and sends it out. ⁓ Don't worry about throughput right now. We'll talk about it. So let's say Clad1 wants an ID. There is a central ID server. The idea is that whoever wants the ID makes a network call.

ID server, channel ID and sends it back. Whoever wants it makes a call, channel ID sends it back. ⁓ Now there is an ID generation service that running over here because in a single machine, ⁓ clock skewness would not affect. I will tell you why. I'll tell you why. ⁓ Clock on a machine, ⁓ assume that you are this machine, this particular machine, and you know that your clock is moving faster. ⁓ When you sync with the NTP server, ⁓ you found out, hey,

Current time is not 24 but 23\. So your clock does not change to 23\. Your clock slows down the counting and then it catches up and then it starts from 24\. So instead of the time going back on your machine, it just slows itself down and stretches that one second to become like 1.2, 1.3 and then it receives. So always remember that clock never goes back in time. ⁓ Never.

Sneha Mehra (00:40:58)  
This means that on a single machine, you will never experience your clock moving back and forth. It can slow down or you can just pass it up. So it would always move forward no matter what. And that's such an important thing. Now, which means that if you have your ID generation logic stick to one server, what would happen ⁓ is that no matter where the request comes in, the clock on this machine is always moving forward. will always have monotonically increasing IDs.

that would work just fine, just fine. having this one machine ⁓ is a single point of failure. So now we need to solve it. So you cannot have one machine, which means you need to have fault tolerance to add one more machine or adding a load balancer so that anytime the request comes in, it goes to either one of them. Now, whenever, whether it's this request goes over here or here, it needs to connect to talk to other IDs of the to see what was the last.

ID that it generated and then it would increase to plus plus or some buffer and then would generate a new ID. So which means that these two server need to constantly gossip about the IDs that they generated in order to generate a monotonically increasing ID because they both have to agree upon that, hey, this is definitely larger than the other one that we ever generated problem. So from a very simple ID generation logic, we made our things super.

And you can very clearly see that this gossip you will either lose on throughput when you're doing this high throughput continuous gossip it becomes very messy. at last clock skewness problem is there. So what we clearly see is that when you want strict monotonicity there is no way ⁓ no way for you to distribute. So you cannot have

⁓ an ID generation service which is distributed in nature which also guarantees strict monotonicity at high throughput. Not possible. Because you see every time you think of something there is new problem or a new constraint that pops up. Hey you can't do this, you can't do this. Always, always that will happen. So there is no way, ⁓ no way for you to distribute this. So how do people go about

Sneha Mehra (00:43:22)  
That is where people relax of course, because not everyone needs monotonicity. One question that might come to your head is, ⁓ but my, my SQL server has ⁓ monotonic IDs, single machine. It's not distributed. It's single machine. The ID generation logic is part of your database in general. It's a single machine. That's how you see ID one, two, three, four, five, six, so on and so forth. And if it would have been distributed, multi-mastered, what not.

You would never have it. Try putting mySQL server in multi-master mode with an auto increment ID. See the number of conflicts you get because each one of them is its own brain. ⁓ So keep those things in mind that ⁓ everything in computer science, everything in software engineering is a trade off. You want some, you have to lose some. ⁓ You get something, you lose some. All right. And there are so many factors that we look at. So in most cases,

Whenever you are having a need of ID generation or something similar, always see that do you really need these constraints? Do you really need what you have been asked for? Try asking critical questions, challenging it. Hey, can we not relax these constraints? What if, what if ⁓ I don't care about integer IDs? Is that okay? What if my IDs are random like UUIDs or GUIDs?

What if my IDs are not monotonic? They don't follow monotonicity. ⁓ Is that okay? ⁓ Because in most cases, the requirement says you need monotonicity. But after asking a couple of critical questions that automatically changes because not everyone thinks it through. ⁓ So asking these critical questions is extremely important whenever you are designing. Okay. ⁓ One final thing before we take questions. Why do we need IDs and questions? Why? ⁓ Obviously, that should be the first thing that we should answer. But now that we have

the context. Let me answer it for you. You'd think that hey, MySQL is there now. MySQL or any database using it has ID. MongoDB has ID. Why are we, why do we have to write an ID generation? Because imagine, imagine you are having a sharded database. Multi-master. Multi-master, which means you have two master requests can go to anyone. Right? When request can go to anyone, ⁓ if

Sneha Mehra (00:45:49)  
I am let's take an example of my SQL server itself. So let's say I have a my SQL server to my SQL servers. Both are running in master mode, which means right can go to anywhere. Now, when you create a table, you have an auto increment ID, you have an ID column, which is auto increment. You have either ID column, which is auto increment. Now, if you do not have an ID generation logic of your own, you would configure typically this auto incrementing. Now,

These two server do not communicate or do not talk to each other. Correct. So when this request goes over here, the auto increment ID kicks in. When auto increment ID kicks in, would start, let's say your post table, you start sending ID 1, 2, 3, 4\. Here also 1, 2, 3, 4\. Because they don't know about each other. Rights can go anywhere. Right? If you go for auto increment, this is where the problem would start to creep in. So which is why

Now we understand why we need concatenate machine ID and this and that. So that is a unique identification. So now the IDs that you generate, cannot just rely on database IDs, but you need to have your own custom ID logic where you cannot do this, but just have something that would help you uniquely identify those entries. So whenever you have your databases operating in multi-master mode, you will typically see a custom ID generation logic in place. ⁓

And that's when you need this word ID generators will talk about very practical use cases of it in order to understand how they are leveraged. Like all four examples will go in depth and understand. Okay. Any questions? We'll take it and then we go on. Yeah. ⁓ Yeah. So I was just wondering that the place where we ⁓ opted for a shared service.

to generate the IDs ⁓ instead of a shared service if we just have a common database and we take locks on it would that work? That's exactly shared service your database is a shared service there. Correct? ⁓ Yeah, but if we take locks on it wouldn't it be unique then? No, it will be unique but that is still a single machine right that is that database itself becomes a service right? Yeah. Right? Okay. ⁓ That's the same thing that's one and the same thing. Rajiv? ⁓

Sneha Mehra (00:48:10)  
⁓ The places where we want to resolve conflicts ⁓ based on IDs now, since we are not sure whether the IDs on distributors system would be unique. how do we tackle that problem? We'll talk about that. We are still only ⁓ basically 30 % done. And we'll talk about that. How do we resolve what happens? How should we approach that? We'll talk about ⁓ that. Pankaj.

Yeah, so when you were ⁓ discussing the last word, right about auto incrementing by ⁓ one, two, three, four in both the servers. So can't be like auto increment by like the number of servers we have basically if you have two, then zero, two, four, and one will like one, three, five. Yeah, we can do that. But there is a slight disadvantage to that. We'll discuss in the second half. I haven't covered. Okay. Okay. ⁓

So instead of static counter, why can't we use a random ID generator here and append it to the epoch? will talk about it. ⁓ But great that you folks are already thinking about this. am so loving this much. You folks are continuously thinking. But which is where the static counter would pick because we wanted monotonicity. ⁓ That's why we went for that. ⁓ Random ID does not give you monotonicity. That's why I said that if you need monotonicity, like you may not always need monotonicity. You may think you need monotonicity, but in reality,

You can relax those constraints and say, Hey, what if I don't have monotonicity? Then you can go for a random ID like UUID. ⁓ Right. So that is my second question. Why can't you use a UUID instead of all this complexity? ⁓ can. We can. That's what I said. Relax the constraint. If it's okay for you to relax the constraint then use UUID. Don't rely on strict monotonicity because not everyone needs that. Correct. But in that case, I think we will not be able to, you know, give that ordering if we are using UUID.

Got it. You get some, you lose some. Got it. ⁓ But we'll talk about UUIDs. We'll talk about MongoDB IDs, UUIDs and whatnot. We'll see where you can use it. Advantages of it, disadvantages of it. How monotonicity would help you. All of that. Like real practical situations where ID generation you really need, need kind of thing. We'll go in that. I've got it. Yeah. Rahul? So basically my question is about the central ID generation.

Sneha Mehra (00:50:34)  
So what I was thinking is like a central server which generates ID. ⁓ You told us about the distributed thing also. Just a second.

You also told us about the distributed ID server also. So what we can basically do is basically we can have like a particular ID, like ID for each server that we have like in the distributed system in the ID servers. And ⁓ just like they can have their own logic for like creating ⁓ IDs. And they append, they ⁓ just append their ⁓ ID

their own ID server like ID server ID to that logic which is generating IDs. But you're losing on monitor. That's fine. That's fine. That's okay.

There is nothing wrong in it. That's what I'm saying. There is no one right way to do it. For some it works in one way, for some it works for other ways. ⁓ But your approach is also correct. You get some, lose some. ⁓ I'm just discussing on the poor idea behind ID generation. Now when we go into those specific examples, you'll see that hey, if you are getting advantages on a certain part, you'd be losing on certain part. But at the end, it's a trade-off that you would have to make.

what you want and what you can give up by maximizing the throughput and performance and efficiency of your system.

Sneha Mehra (00:52:08)  
Sure. ⁓ Shreya? I wanted to propose one more solution for ⁓ the part where we wanted to reduce the disk IOs. ⁓ So ⁓ can do ⁓ something like when the server restarts, we read from the database, get the last counter saved in the database and then start incrementing ahead. So we wouldn't need those ⁓ periodical saves as well. So then you are writing the database every time?

No, so we'll have some last row ⁓ which was saved to database. That's what I'm saying. This is not where we are using it and storing it in database. No, no, no, not the IDs. Let's say we are saving messages to database. Now that message ID, ⁓ certain row has a message ID which was last saved to the database. That's exactly what this load is actually doing. This load is actually doing that itself. Writing into the starting into the database is one and the same thing.

Right. You start, you're starting from that particular point. Correct. ⁓ But then how are you, but here you are assuming that you are writing this, this, whatever you're doing using this get ID for you are writing it back to database that row is getting inserted. I'm not even assuming that. Okay. ⁓ Got it. You are assuming that an entry is getting created in the database for every single thing. And then you are picking the max row and then starting from that. I'm not even assuming that you're using this.

Like in this approach, I'm not even assuming that you are writing back to the database because that is anyway expensive. Right. I'm just focusing on this ID generation logic. But then what you're proposing max row ID, and then you start from the, that's exactly what this load function is all about. Yeah. I'm tracking about loading from the disk, but with this, you're assuming that when you get an ID, then you are doing something and then storing things into database. ⁓ And then that's an expensive call that you're making.

I'm assuming not even assuming that you're sitting back to the database. You're just attaching something at this, basically pushing it forward. Let's say request ID card.

Sneha Mehra (00:54:13)  
Okay. Yeah. Let's take a quick break for five minutes. Once we come back, what we'll talk about is we'll talk about how Amazon does it using central ID generation service and then other use cases. I'll show you a very nice demo of, know, why people should like why people go for a certain way and not the other way and how to make those particular trade off decision. Right. So we'll take a quick break for five minutes. We'll sing back at 10\.

And we'll talk about ⁓ Amazon, Flickr, Twitter, Instagram, ⁓ and pagination. A very interesting side effect of having a good ID generation system. ⁓ we'll talk about. See you folks in five minutes.

Sneha Mehra (00:59:55)  
Let's start with the second half. So now what we'll talk about is we'll talk about how Amazon does it. ⁓ FII, I know ⁓ this because the team that did it sat right next to me. So I used to attend a bunch of tech talks there, but I had a bunch of questions to them and kept on asking it, how do folks do that? And they were quite enough to explain it ⁓ in a public tech talk. So, okay. So what Amazon does is for every order that you place,

⁓ is they assign a global ID to that order and that is unique. ⁓ Given the high throughputness ⁓ of the system and plus not just this that they provide ID to a lot of other internal services. ⁓ One of them happens to be order which we all use on a day to day basis. So that had to be unique. So what they had is they you leverage batching. Batching is one of the most common ways to get performance out of any system. ⁓

So what they do is they have something like this. ⁓ They have a bunch of app, they have a bunch of API servers. Like you can literally write this thing in 150 lines of code, not even exaggerated. ⁓ It's that simple to write, right? ⁓ Scaling, you might just have to handle a other cases here and there, but it's not difficult for you to write this particular service. ⁓ Hardly 150 lines of code, not even kidding about anything else. ⁓ Python, if you write 150 lines, you can write this entire service. So.

Recommending you to write it. ⁓ Okay. So ⁓ API, say you have basic API server exposing your ID generation endpoints, putting behind a load balancer classic thing and have a database. I have a database. Now this database is a classic relational database. No fancy non-relational database. ⁓ I don't like to unnecessary complicated things. Normal relational database would work over here. So what we do ⁓ is similar to how we did not

right to the disk every time we generated an ID, we did it in batches. ⁓ Once every thousand ID generated things. Similarly, we do things over here that instead of calling your remote service every time you need an ID, you get a batch of IDs in one shot. That's the core idea. ⁓ You would have seen a ton of videos, few videos on YouTube about this, but let me go into implementation details of it. Given that we covered

Sneha Mehra (01:02:21)  
airline checking system this becomes a piece of cake. So now what would happen is let's say I have two services orders and payments when order service wants IDs so these are the servers that are running the order service these are servers that are running the payment service. So the idea is they don't care about monotonicity FYI they don't care about monotonicity they need uniqueness as simple as that. So now what do you do?

is when order service wants an ID, any server of that makes a call to this API, to this ID generation service and says, Hey, I just booted up. Can you give me 500 IDs? So it just requests for 500 IDs. The request comes to this API server. API server fires a SQL query on this deep. What kind of SQL query? That, Hey, I need 500\. So it wrote 500 over here. So

0 to 500\. So the range is sent to this server. So 0 to 500 range. You don't have to send individual IDs. No, start and end. It's just integers. Right? So start and end, just send it that this range belongs to you. And you update it in this database. Now you can very quickly start thinking, ⁓ exclusive. ⁓

you would do it for update like select start from this for update do it in a transaction so that it does not like multiple requests for a particular service does not affect all of that comes together. Right? So you would automatically update that updated in a transaction ⁓ of the final value that is handed already handed over. So let's say server one came up. It's an AI. I don't have any idea to work with. So I'll go to this and get 500 IDs to this server updates this to 500\.

and responds back with this range, not independent IDs, no needed. Why to send so much of data over network? Just send start and end, problem solved. ⁓ Then let's server two came up and said, hey, give me another 500\. I am new one, give me 500 more IDs. It would make a call, call would go over here, it would go over here. It would say 500 plus plus, but a plus plus equal to 500\. It get thousand. So it sends 500 to 1000 in response over here. So now server two has this range.

Sneha Mehra (01:04:41)  
Server 1 has range from 0 to 500, this is from 500 to 1000\. Now, order service, when they start getting the request from external world about orders, when the request comes to this order services web server, they already have in memory this range. Now when they would want to generate, they would have their own databases, orders database or something. They can start putting data to orders database, not really having to worry about anything. ⁓

The idea being that because they already have the range of IDs in memory, they can start pulling that out. ⁓ And that becomes the globally unique IDs for that particular order. ⁓ And they can prefix it whatever they would want to do in case they follow those conventions. But the idea is that this way, this central ID generation service just decouples the system. Now you can have any number of surveys, order service can have any number of databases without having any problem because ID generation service is separate.

and others database separate. So I can short order database and scale it horizontally without having any problem, not needing an any, every one of those database can start accepting the rights. This way, what happens is that your server does not have to make call every time to this ID services service to get a set of IDs. I use 500 as a small example. In reality, is 10,000, 50,000, 100,000, depending on the throughput of the service. ⁓ Disadvantage of this ⁓ is

In case your server got this arrange and then it server crashed this range is gone. ⁓ putting this range keeping this range too small means you are making too frequent network calls to get the set of IDs. If it is too large and if a server crashes you're losing the entire range. So it's up to you your service your throughput in order to fine tune how many service at max will you get it in a batch. ⁓ It's up to you. ⁓ You make that call you get it. ⁓ Similarly

You had one entry for orders. Similarly, you would have one entry for payment service. So when your payment service handles the request, it would not have to talk to the central ID services service every time. Instead, ⁓ your server can load it once or once again, ⁓ about to be exhausted, it would load it and start using it whenever it wants to do that. ⁓ Whenever it wants to process the payment. ⁓ Once any of the server sees that my IDs are getting exhausted, let's say it got 500, ⁓ 490 are already exhausted.

Sneha Mehra (01:07:03)  
What it would do, it would make call and get another 500\. It might get some, let's say 3000 to 3500\. So now it would have those rates and would start using from that particular range that it has. So they are still not guaranteeing monotonicity across servers, but you are handing them those IDs in batches. ⁓ This is how Amazon does it. Really simple, very simple to build. can very clearly see how we can do this in transaction. Airline checking system, control C, control V over here.

⁓ Read, in transaction update, atomically do plus equal to range was or another, however ID was requested, you get the old value, you get a new value, you send it as a range over here and server starts ⁓ using those IDs over here. So here, the single source of truth, because this is single server, it doesn't have to gossip with anyone. You can very easily scale the system and just, it's just one update. You don't even have to read all of it. Just one.

miniscule tiny update that you're making in a database every time someone wants this site. Right? Okay, this is how Amazon does it. Any questions?

Sneha Mehra (01:08:14)  
⁓ I was wondering why was ⁓ in this architecture database not used for sorry. Hello. ⁓ I'm audible now. ⁓

Sneha Mehra (01:08:31)  
Yeah, Arpit, I have a question. ⁓ wait, wait, sorry, sorry. It must be a problem with my mic. Wait, wait, wait, wait.

Sneha Mehra (01:08:48)  
Yeah, I've got. Audible. ⁓ I was wondering why was that ⁓ a database not used to keep track of information? How many how much out of range server has used and then when it comes up, it uses that information to start consuming from there. ⁓ Why would you want to I mean, I know that you have the range is appropriate enough to decide. These would get wasted. I mean, 100 ideas were used. 54 out of 500 for it got wasted.

If with this thing, but, ⁓ let's say we go for 10,000 now for the same thing. So 9,900 ideas got lost. I mean, there's no way to again go back and use them. Right. Imagine the word that you would have to keep in order to keep track of every single one of them. That's a massive overhead. Why? Okay. And imagine big integer two days to 64\. ⁓ Will you ever exhaust that number? ⁓ Amazon is $500 trillion worth.

company then. ⁓ Right? ⁓ It's okay. okay. Was under practical usage of it then. ⁓ Yeah, ⁓ obviously like think about practically did every step. Don't just think you have to use every single, every single ID that is out there. No one cares about it. It's just a few range of numbers because we tend to underestimate the numbers. For example, ⁓ 1 million to 1 billion. If I talk about it, we just use 1 million like it's nothing.

1 million is a huge number. 1 billion is huge number. The difference between a million and a billion is almost equal to a billion. And that's huge. Right. And we just use it like this. ⁓ 1 million, 4 billion, 5 billion. It's massive, massive, massive numbers. And 44 by integer, 4 billion, 4 billion orders. Imagine 4 billion orders of Amazon. ⁓

If you exhaust that, make it 64 bits, make it eight by 10 because two risks to 64\. ⁓ Not very much. ⁓ It's big enough range for you to not worry about it. Okay. ⁓ Mohit. ⁓ Yeah. I feel like if monotonicity was not a requirement, then why not just go ahead with a UUID kind of approach. Let's talk about it. Let's talk about it. Like why not UUID? ⁓ So if UUID already exists, ⁓ why people are not, ⁓ why companies at scale don't prefer it?

Sneha Mehra (01:11:12)  
⁓ I also don't know what UUID is. UUIDs are 128 bit integers. Integers are 128 bit strings which can be represented as gigantic integers. They are extremely inefficient at scale. ⁓ My company is using UUID, are we inefficient at scale? Yes, you are. Okay, let me break it up. in the ⁓ second week pre-reads, I shared a video on how

indexes make database read faster. In that we saw that indexes are also just tables with indexed values ⁓ and the IDs of the rows that it points to. ⁓ For example, post table has user ID and post ID and post title and what not. If you creating an index on user ID, it is effectively a new B plus three or a new table which is created in which just two columns are there ⁓ just to visualize.

in which you have first column as post user ID, second value as post ID. It's internally managed by database FII. So what happens over here is because in index for a dense index at every single row of your database or of your index, you are also storing the row ID. If this row ID is big, it would make your indexes bulky. For example, 128 bit integers is equal to 16 byte. ⁓ Now imagine,

If your index is first set, if your index took ⁓ one GB, if your ID was just a four byte integer, but now you are changing it to UID which is 16 byte integer. So now you just bloated your index by a factor of four. So your index, which was one GB in size is now suddenly four GB in size. So always remember your database is as fast as like you get the max performance out of your database. It should be able to fit.

all the indexes in RAM. If it is not able to fit all the indexes in RAM, ⁓ no matter how many optimizations you do, ⁓ your database will never get speed. ⁓ Never. ⁓ Because whenever your SQL query executes and not just Rupert SQL, ⁓ any database in the world, wherever indexes are there. ⁓ First set of, whenever you send any query, ⁓ the first set of things that it does is does this query plan, query execution plan it creates while doing that.

Sneha Mehra (01:13:35)  
All the operation it does on indexes because indexes are really tiny, really tiny indexes. ⁓ When it fits in RAM, let's say you're doing joins of two tables on the indexed columns. It can just do it in memory. It doesn't have to read from the disk anytime. Once it knows the result, then it goes to the disk and reads the entire row. So that's the last step. Right? So if your indexes are in RAM, entire query execution happens in RAM.

Just the things that you need from the row from the disk is the detailed row that was setting in the result. So this way, execution becomes slightly fast when your indexes can fit in RAM. So when ⁓ I am ever optimizing a database ever, when someone comes to me and says, Arpit, my database is slow. What to do? The first thing I do is check their index sizes. That's it. In most cases, index sizes are the problem.

random index has created on the database and this and that I see the amount of spills that happens on the disk. Simple because if your query execution for in order for you to even read your index and it has to go to the disk to read it, that would be very slow. So that's why it does not mean that your database keeps your indexes in RAM. It still flushes it to the disk. That is the main source, but it loads it in RAM during boot up or whenever you are requesting it to when you're finding a particular query and it keeps in RAM. So

The RAM of your database server should be equal to or should be more than what your total index size is. Then you get the max out of your database. Otherwise you cannot. Otherwise your database query execution would be pathetically slow. And now is it just over exaggerating? I'll tell you other designs where because of this people have not used UUIDs. Not used UUIDs. Okay. So UUIDs, they are gigantic integers. 16 bytes, 4x larger.

That's huge. So one GB index becomes four GB index now. So they do not index well and they load up the index. That's right. But they're good for security because you cannot have like, you cannot predict what's next in terms of UID. They seem very random ⁓ because you cannot predict the next one. So it's hard for anyone to scrape out the entire website. So for example, if you're, ⁓ if you're, ⁓ if you're building Facebook and you know that you are

Sneha Mehra (01:15:54)  
exposing their database IDs 1, 2, 3, people can just fire a request and start from for i equal to zero i is less than 1 billion i plus plus and start scraping the entire website. ⁓ So because UUIDs are very random, they're good for security because it's hard for anyone to predict and penetrate or just predict other IDs. In most cases they would get four. ⁓ So which is why people don't use UUIDs at scale.

It does not mean you should not use it. It works in some cases. It does not work in other cases. ⁓ some cases where you're just grabbing it out. You just, know that your application is small enough. You would never hit that table. You never hit that scale. The UI is just fine. Right? Not unnecessarily go for the best solution out there. See what works for you the best. MongoDB is just happy with 96 bits logs. So MongoDB IDs are 12 byte integers. Where? Look carefully. First four bytes is epoch. Time.

always comes first. ⁓ Then five bytes of random and three bytes of counter. Counter for each shard that it has. ⁓ This way is 12 bytes is what they are favoring right now. They didn't go for 16 for UID. They weren't stopping them from going for UUIDs. Nothing. To reduce by four bytes, they went for share. ⁓ MySQL auto-increment IDs is four byte integer. You are very happy. That's very tiny index. ⁓ Four bytes, eight bytes is still doable. Beyond eight bytes, you should worry about

the performance at scale at scale. ⁓ Okay. ⁓ So this is what MongoDB does. ⁓ You would see a lot of ID generation sitting things on the left more like having time on the left most side and trying to keep the ID smaller, which is why in the first half of when you're brainstorming, IDs becoming longer is a problem because they do not index well. ⁓ Let me take one more example to help you understand this entire thing better. ⁓ Every single company, every single

quote unquote case study we look at every single one of the further case studies mentioned that they didn't go for UUIDs because they do not index well because this is the harsh reality. ⁓ So let's take a look at what Flickr did. So Flickr is pre Instagram. ⁓ Flickr was Instagram of 2010 to 20, ⁓ basic kind of 2010\. A lot of things got posted there. Then obviously Instagram took over a lot of stuff. So Flickr needed its own ID generation. It's like, why?

Sneha Mehra (01:18:16)  
because database was sharded, which is what I always say. Because your database is sharded operating in multi-master mode, which means that rights can be going to anywhere. They need a central ID generation thing to maintain the uniqueness, right? To guarantee uniqueness, they need a central ID generation service because they cannot just use auto increment everywhere because that would result to conflict. Right? So because of which they needed a central ID generation service. Now you know where you did it. Flickr did it because of this exact same reason.

⁓ And in their blog they mentioned why not UUIDs because they index badly every single one is mentioning UUIDs index badly because 16 byte is a huge thing huge thing because it just forex's your index x almost rough estimate don't quote me on that because your integer was 4 byte now you're taking 16 byte so that's an additional thing that you are taking up and that forex is a wrong calculation actually depends on the value but you get that

⁓ Okay, so this is what it's written in their block, their official blog. If you cannot fit indexes in memory, you cannot make your database pass. And it's true. It's true. Anytime a serial engineer, technical architect, stop-and-jump principle engineer ever optimizes the database, the first thing he or she looks at is the index size. Always remember this. Because if your indexes spill on disk, your database would always be slow. Always be slow. And this is one of the ways through which you can actually provision the database. If someone asks,

Ask you not during interviews. Interviews are crappy. In your real world, if someone asks you, hey, what should be the size of this database for it to be performed? The bare minimums that it should be, what should be the summation of sizes of all the indexes that it holds? Simple. Bare minimum. That's the bare minimum that it needs to be. Beyond that, anything is good, but that's the bare minimum that you should have. ⁓ Always keep these things in mind. That's why database fundamentals so, so, so important because

That's the, as I always say, database is the most brittle component of an infrastructure. Most essential, but most brittle component of an infrastructure because scaling API servers is easy. They are mostly stateless. Every other thing is stateless. Database is the one where everything is getting stored. Everything is getting queried. Entire load is handled by database. So understand your database, whatever, database you're using really well. ⁓ Okay. So how did Flickr did it? So Flickr

Sneha Mehra (01:20:41)  
built a central ticket server, basically IDs and this is right. ⁓ And they built it on top of MySQL. So what they do is whoever wants to get an ID, they make a query to MySQL and they get an ID. Don't think of performance. It's their decision. We can't say much about it, right? But understand the beauty of this design. That's what matters, right? As an engineer, focus on the beauty of design rather than pointing out the faults. It's okay to point out the faults, but pick the good parts of it.

Right? Because it was their decision. We can't interfere. Like it's like your neighbors are fighting. You know that both of them are wrong, but you should not say anything to them. It's their life. Whatever they want to do it, let them do Right? But pick the best parts of it. So what they did is they created a table called tickets. In this table, they added a column called ID, big integer, unsigned, not null and auto increment. Right? You clearly see big integer 2 raised to 64\. Right?

They added a column called stub will come to this what stub is all about. ⁓ They made ID the primary key, unique key to stub. The idea is that if MySQL already gives me auto increment IDs, why can't MySQL be my ID generation service? Exactly what we were discussing some time back. ⁓ Someone asked that, but I think Shreya was that, that asked that part. Why can't we just use my existing ID generation? ⁓ I can't use my database as my ID generation service.

That's exactly what they're doing. They're using database itself as separate database. Photos are being stored in other sharded database. Your ID generation itself is a separate database in which you are leveraging the auto incrementedness of the database as your ID generation file. So what happens? ⁓ Let's say you want to naively ⁓ use a database as that. You know database gives you auto increment. Whenever someone needs an ID, it talks to this database and you want ID from that. Basically what would you do? You have this column.

ID auto increment ID you would delete this and add a new row. When you add a new row, what would you get? It would be the next ID. Then when you need one, you delete one and you add one. It would get the next ID. Right. But for you to add one, you can just cannot insert a row with nothing. You need to pass some values to them. So you had another column called stuff. It's a normal column. The name seems fancy. Just a normal dummy column, some dummy value, nothing special about it. Right.

Sneha Mehra (01:23:09)  
But what we are effectively doing is we are just inserting ⁓ this row. We're deleting and reinserting, deleting and reinserting, deleting and reinserting. This way, whenever you're inserting into this table without providing the ID, the database would auto increment the ID and would use that as a default value while inserting. And then the next one, and then the next one, and then the next one. Classic auto increment ID use case. ⁓ Right? So either we delete or increment, ⁓ delete or insert in one transaction, ⁓ or we can use

upsurge because delete plus insert is equal to upsurge. Right? So because they were using my SQL, I'm talking a bit about my SQL here, but my SQL because they're using my SQL, my SQL provides two ways to do it. First is insert on duplicate key update. So it's very simple. What it does is it tries to insert whatever you provided over here, insert into this, this, this values. This is this. It tries to insert that, but if it fails,

due to unique key constraint then it triggers the update that you passing that is classic upset so you are trying to insert so what is upset you are trying to insert something if it exists then update if it does not exist then insert this is an upset statement for mysql so there are two ways to do upset first is this the second word is replace into similar to how we have insert into

there is replace into we saw this in key value store on relational database this replace into right? replace into does absurd but what replace into does internally it actually deletes and inserts the new row so you can very clearly see it is slower it is basically 32 times slower than on duplicate key update so you can use either replace into or insert but replace into is 32 times slower than this I personally benchmark this that's why am

coding this number 32x slower. I personally benchmarked it on a MySQL server to see the performance difference between the two. So it is 32 times slower over here. So pick one, anyway, what Mattress is doing upset. So what you're actually doing over here is something like this. Whenever ⁓ a server, API server wants an ID, so they can insert a photo in photos database, what it does is it fires a SQL query on the central ID service, the ticket server, something like this.

Sneha Mehra (01:25:33)  
Insert into tickets stub values a because you have to pass some values so you use stub a random value. matters. You put any value there. But it is a constant value. Insert into tickets stub a on duplicate key update id equal to id plus one. Simple. This is what happens is it would try to insert this a but it would see that this already exists. So on duplicate key update that would update 74 to 75\.

and return that. ⁓ Then next are the verifiers 76, 77, 78 and so on and so forth. So you get ⁓ IDs like this. You leveraging MySQL transaction, MySQL atomicity in order to get the IDs that you want.

Right? So this is what Flickr does as part of like implementation logic. But now what you see is like something like this. So this are your photo service server. This would make it clear. This is your ID generation server, MySQL that is different and photos database is different, which is sharded. So what they're doing is whenever there is a request to post a photo request comes to this API server, API server fires a request to this MySQL to get the ID. You get the ID, then you store it in PhotosDB. Right? This is what is happening.

But if you look carefully, this is a single point of failure. Because there is one server, if this server goes down, then what? But how to solve it? You put two servers, but this is very similar to that problem now. Corsive, monotonicity or what not. Here we don't care about monotonicity. So what we do, we put both database behind a load balancer ⁓ and make ⁓ one

Someone also recommended this approach. is exactly what we are doing. So what we do is we have two databases. Now, even if one of the database goes on other database can handle the load. But how do you ensure uniqueness? How do you ensure that they don't conflict? What we do is let's say I have two databases. I'll have instead of doing increment plus plus I do increment plus equal to two. Let's say one database generates odd IDs other database generate even IDs. Right? So how you can configure this?

Sneha Mehra (01:27:51)  
is auto increment underscore increment these are actual configuration that you can provide in your mysql server auto increment increment is set to 2 and auto increment offset is set to 1 so increment increment this means by how much you would want to increment every time ⁓ and increment offset is what's the starting point of it right so this is what you would configure and now the load variances that you put over here is configured in round robin port so first request goes over here

It generates ID bar. Second request goes to other server. Third request comes to first server. Auto increment plus equal to 2, 1 plus 2 is 3\. So get ID 3\. Next request goes over your 2 plus auto increment increment is 2\. So 4, 5, 6, 7, 8\. ⁓ Because you doing it in round robin, would obviously first request you have second here, third here, fourth here, fifth here, sixth here, seventh here, eighth here, here. You almost get more identity. Right? You have solved

problem of spoff, single point of failure. Now you have two servers. If one of the server goes down, it's still fine for you because your ID generation is still continuing to work without any hiccups. ⁓ But what happens when it goes down? What do you do? ⁓ If one of the server goes down, ⁓ let's say my odd server went down due to some reason. Now all of my requests will go to even server. So the IDs that would spit out would be 2, 4, 6, 10, 12, ⁓

Now assume that after 202 ID was generated, my server, my odd server came back up. Now when my odd server came back up, either it can start from the point where it left up because my SQL is a persistent storage. It's not storing in memory. So let's say generated ID till five, I would generate seven, 204, ⁓ nine, 206, 11, 208\. That way. Here you see

a huge discrepancy in IDs. It's not that it's a constraint that you are playing with. It just doesn't look aesthetically pleasing. Much more for aesthetics rather than correctness. You still getting unique ID. Well, there seems to be very massive parity between. So it's optional to just reset the one that just spun up to the ID. Let's say this generated ID till 202\. You just add some buffer, let's say 240\.

Sneha Mehra (01:30:14)  
And you set here 241 and 242 and then it resumes from that. So 241, ⁓ 242, 243, 244, 245, 246 and so on and so forth. ⁓ It just for aesthetics, nothing about correctness, nothing about strict monotonicity because we're still not getting strict monotonicity. ⁓ But you still have that as like nearly similar. So it's good. It's a good byproduct to have. ⁓ So this is how Flickr does it. It's a really simple approach that we right now looked at.

It's pretty simple, but it's pretty brilliant in its own sense. Two servers, order even and resume backup. That's one thing that you'd say who would do this? Who would do this? Who would find this max buffer and then start with? So we are starting with the safest value, right? Who would do this? Obviously, when a server goes down, ⁓ SRE engineer DevOps engineer, you know, they would fix it. They can configure it. They can change it. You can write an automated script, but typically database going down is a big thing.

There will be an engineer looking into it. He or she will do this. It's fine. ⁓ In case you can also, in case you want to automate, you can automate. You just need to find the active server, find the max IDs, add some buffer to that and then set the values and then just do a quick process report on each server. ⁓ Any questions on this?

Sneha Mehra (01:31:38)  
Yeah. ⁓ Can you just go back to the previous page where you have created that query? I'm not able to understand how it's creating the unique value at this one. Can you just elaborate maybe one more time? is 75, 76, ⁓ insert into tickets, stop values A. So what we are trying to do, we are trying to insert a value. Something we have not provided ID. So we're trying to insert this value, right? ⁓ But stop is having what?

Unique key. Correct? What would happen? This insert would fail. You cannot have two same values over here in this column. Correct? So on duplicate key update, update to what? ID equal to ID plus one. ⁓ That I understood. Step part where it is coming, ⁓ not able to make sense of that. It's a, it's a just a common value. It's just a random value that you chose. It's a hard coded value in your query. Hard coded. ⁓

It's a hard code, it can be literally any string that you would want to It's just that you want to trigger this. You want to break the unique constraint. Correct? You just trying, are making your database break the unique constraint so that you can trigger an update. Atomic. ⁓ so for that purpose only a stub exists. That's it. That's it. There is no other reason for stub. Okay, got it. ⁓

Yeah, if you can go to the last page, I'm not able to understand how ⁓ when the server comes back, how you're resetting the values with the buffer value. And this one, I said, no, one engineer will do that. This, this configuration, these configurations, it's my SQL specific check how you can change auto increment offset on the platform. But, but, but why, why do we need to do that? Because let's say one server said aesthetics. ⁓

I said it's for aesthetics, nothing more. Even if you don't do it, your system is still running correctly. It would just look bit because the last value only will keep on updating it, right? ⁓ it doesn't work Yeah, but it would. But that's what I said. Let's say it died off at 5\. And this value is 202 right now. ⁓ If odd server comes back up, what would be the next ID that would be generated?

Sneha Mehra (01:33:59)  
But this is an even server. ⁓ is page 202 odd server came back up. What is the ID generated on the odd server now? ⁓ From the last one to whatever value was there. have 3135 what's the next value that would be generated? ⁓ 137 7 and on the even server. ⁓ Okay 204 so just to maintain that. ⁓ Correct. ⁓ Just for aesthetics. That's what I meant. Just for aesthetics.

Okay. There's nothing special. Even if you don't know your system would still function correctly. ⁓ It just for aesthetics, you can just reset the value on both sides and make it look better. ⁓ That's it. ⁓ Right. Okay. So after reset, what will be the right value? ⁓ to you, ⁓ start with 240, ⁓ 239, ⁓ then let it round rob. Got it. Got it. Got it. Right. Pick some value. ⁓ It's up to you. You are the design. You are the architect of the system. It's up to you. Very good. ⁓

I got reset to do the safest value of ⁓ that. ⁓ That's why we discussed safest value in the first place so that we understand the importance of starting from a safest value, not conflicting with the IT's in the past. ⁓ Okay. One more question before we move forward. Yeah. ⁓ So I think I'm a little confused here in this one. You said it, the max plus buffer thing is just for aesthetics. ⁓ But let's say if the odd server goes down, let's add five.

And the server goes on till 202\. If you don't do this max bus buffer thing, and the odd server starts producing IDs from, let's say seven, ⁓ then they are not monotonically increasing. So we are not, where did we enforce monotonicity? Okay. We're not ⁓ enforcing that here. Okay. ⁓ That's what I started saying. We are not enforcing monotonicity. It's a nice by-product to her. Okay. ⁓ Because that's what I said. Not everyone needs monotonicity.

⁓ And it's okay to not have social media. would you care about what I'm saying? ⁓ Okay. Now the next part ⁓ is we discussed like the two approaches of artilla that we saw Amazon and Flickr. They were central ID generation service. Now what we take, we take a look at distributed part rather not distributed decentralized ID generation logic, not service logic. So Twitter

Sneha Mehra (01:36:23)  
still uses this one of the best IDs generation system out there. So what Twitter uses an algorithm called snowflake. ⁓ I, ⁓ two words back, I never thought why they are calling it snowflake. Now remember just a little backstory. Snowflake ⁓ is like ⁓ almost every single snowflake in the world that you see is unique. That's why the ID generation logic they're calling it a snowflake.

So pick any two snowflakes, would no way be similar. They would no way be exactly same. They would always be dissimilar. That's the beauty of snowflake. In case you're interested, the ⁓ number file has a great number. Mostly number file has a great video on why snowflakes are unique and not ID generation but the actual snowflakes. Unfortunately, computer has to like when we have windows, we think about operating system and not the physical windows. Unfortunately, we are now used to that. So snowflake not ID generation but the actual snowflake.

Every single snowflake is unique because of which they named this particular algorithm, if I may call it, as snowflake. So what they do ⁓ is Twitter uses snowflake for their IDs. Snowflakes, this algorithm is also called Discord and Instagram. We'll talk about Instagram, Discord is very similar. We'll also talk about Discord also and Instagram also. So snowflakes are 64 bits in DGIS. 64 bits.

keeping it small. That's what they also went for. They did not want to keep it large. If they would have kept it large, imagine at Twitter scale, amount of data that is getting ingested, the amount of data that is getting indexed. Imagine instead of having 64 bits, you are having 128 bits, which means 16 bytes. You are literally doubling the space that you need. That's why efficiency at scale, storage efficiency at scale, top-grade efficiency at scale. ⁓ So what snowflakes are?

Slowflakes are very simple. are very similar to MongoDB's ⁓ ID, but very small, very nice. ⁓ So what they have is they have first 41 bits ⁓ as epoch milliseconds. Left side, always type. 41 bits, epoch milliseconds, which means every millisecond, this would always move forward. So you almost get ⁓ ordering in IDs. That's a very nice by-product. We'll take a look at pagination because of this, how it is very efficient.

Sneha Mehra (01:38:48)  
Then the next 10 bits for machine IDs. Machine IDs is server that is generating this ID. Not database server. It's the API server of Twitter. ⁓ imagine you're tweeting something. So what do you have? This is you. This is tweets.tb, which is partitioned, massively partitioned. And these are the API servers. Normal API servers handle REST requests to create a

Here you have this ID generation logic called snowflake. It's not a library. It's not a service. It's a library. Normal, normal integer, normal bit manipulation. When you get a request to tweet something, it does basic validation there. Tweet ID, user ID, it extracts that. It generates a random tweet ID using snowflake ⁓ and it just stores it into this database. These are API servers. So generated. So this 10 bits

are for these machine IDs. Now because of this 10 bit constraint, you know the maximum number of API servers that Twitter would have had for tweet service would be AtomX 2 raised to 10, which is 102 for. It cannot go beyond that. Correct? Otherwise this would overflow. Correct? That's what happens. So you, it reserved 41 bits for epoch milliseconds, 10 bits for machine IDs,

and 12 bit per machine counter that static counter that we discussed at the first. It will press, press, press, press, press, press, press, time. Now, this is what it generates on demand. When you get this, see, when you get this in the business logic, we are generating a new snowflake ID. You need current time you have it. Machine ID, you would know that this is my machine ID. It would be in range of 0 to 1, 0 to 1, pick one. Every machine will get a unique ID to start with and a 12 byte machine counter per machine counter.

What does this tell you? This tells you that per machine counter is 12 bits which means that in ⁓ 1 millisecond ⁓ one machine can generate 2 raised to 12 unique IDs before seeing any collision. Per millisecond per machine 2 raised to power 12\.

Sneha Mehra (01:41:07)  
2 raised to 10 is 1024, 2048, ⁓ 4096, 2 raised to 12 is 4096, which means in one millisecond, ⁓ one machine can generate 4096 unique IDs without seeing any conflict, ⁓ one millisecond. ⁓ That's a humongous throughput that they have made the system tolerant enough that hard to attend that scale. ⁓ And so they are prepared for, well prepared for the future. ⁓

because at max 1024 machines so when entire collision would happen when snowflake would see when when would snowflake even see a collision where same IDs are getting generated it would happen when in one machine in one machine in one millisecond you are handling more than 4096 requests no one can do that one millisecond more than four more one millisecond more than 4000 requests per millisecond

implies 400,000 requests per second. No one does it. No one does it. So they're very much safe from the scale that they are handling and they can at max have 10 machines. Right? So this is normal. You can even write snowflake in three lines of code. That too for aesthetic. I can write it in one line also. Three lines of code because you know the current time, you know the machine ID, you know a static counter to do atomically plus plus. Just bit arrange those bits like this.

Like left take a number left should by 40, ⁓ basically left should by 23 and set the time left should by 12 and then set bitwise or basically create this idea as simple as this normal bit manipulation. Right. But the kind of benefits you get out of this is because now there is no moving component. There is no need of a central ID, surveys and whatnot to talk to. ⁓ It's literally a request comes, ⁓ you generate a snowflake ID, ⁓ you store it in the database.

Done. This is why Twitter scales. is no central... because ⁓ firing tweets is one of the most high throughput things that the world is handling right ⁓ now. People keep on tweeting stuff whenever there's a big event as if they're missing out on something. ⁓ They keep on tweeting. ⁓ These folks need to handle the scale. If they would have had a central ID service, they would need to scale that part and this and that and that and that. Screw it. We don't do that. ⁓ Okay. So...

Sneha Mehra (01:43:36)  
The beauty of snowflakes is that they are roughly sortable. As I said, within the same millisecond, assuming that the machine that has lesser IDs, smaller, the machine has larger IDs, larger within the same millisecond, get roughly sortable IDs. That's brilliant. First, 41 bits of epoch millisecond means after every millisecond, your number is massively moving forward. ⁓ And it gives you an ability of getting a tweet.

before and after a certain time because the first 41 minutes. So if you go to Twitter right now, click on any tweet, look at the link that we tied itself is an integer. This is no longer talking about. Now what would happen is the first 41 bits of that number is actually a box millisecond when the tweet was made, which means that. Given you how it is, you can very clearly tell if this tweet was made before this time or after this thing that is brilliant because now it makes

your pagination super efficient. Let me talk about it. Now this is a very interesting byproduct of having roughly sortable IDs. So if you ever take a look at Twitter API, unfortunately it is paid right now. No, they reverted it, but now I think they already, they made it again paid. Not sure what the current state is, but ⁓ if you go through Twitter's API documentation, what you will find interesting ⁓ is the way you can paginate a tweet. For example, you want to list

all the tweets of a person's profile of a made by a particular account. What you can do, there two ways to page management. One most common limit offset based page management. So you fire a SQL query, ⁓ use SQL as an example, but it holds true for every database. FII select star from tweets where user ID is this limit 10 offset 0 then for the next page limit 10 offset 10 for next limit 10 offset 20 and so on and so forth.

That is one way to pageant. Limit offset, skip limit, whatever you may want to call it. That is one way of pageantation. But this has a problem. When you do limit offset based pageantation, it is pathetically slow. In case of Twitter, imagine it is very common for people to scroll down the tweets of a particular person. Let's say Arpit made 10,000 tweets. Going through all of his tweets is a very slow process. ⁓

Sneha Mehra (01:46:03)  
go down deeper into pagination like if I go level deep into it third page is okay, fourth page is okay, fifth page is okay, sixth page would be slower, seventh would be even slower, eighth would be even slower but it is a very common use case to scroll through it you'll say, ⁓ why would someone scroll through person's profile? Okay, let me give you another example let's you are writing a batch job for twitter in order to compute trending topics, right? trending topics in last two you want to process last one hour of tweets

Processing last one hour of tweets in one hour twitter produces millions of ⁓ Imagine doing pagination across those million rows. That would be so slow. Let me give you a practical example of that. I never thought this would come in this handy. ⁓ I wrote a quick demo. ⁓ Back in 2016 I was getting bored at Amazon. Didn't know what to do. What I did is I was heavily into MongoDB back then. So...

I wrote a simple benchmark. found something interesting and I wrote a benchmark back in 2016\. Good that I got boarded. So, efficient pagination MongoDB, MongoDB benchmark. You want this. ⁓ So, this holds true for MongoDB, this holds true for SQL, this holds true for literally every single database out there. Right? Okay.

Let me explain what's happening here. How pagination, ⁓ ID generation, pagination are linked together. ⁓ So there are two ways to paginate as I said. First way to do it is limit offset. This is limit and offset. MongoDB is example, skip and limit. The idea is that wherever you are firing or want to get the next one, you fire queries like this. You'll skip something, you limit something. You'll something, you limit something. Here it should be 10 FYI. But that's fine. You do that. Now what happens behind the scene when you

execute. Now, when you execute a query like this, limit ⁓ n offset n, when you execute something like this, your database has to evaluate the entire query, literally find the entire result set until the point that you requested, ⁓ skip those many of the results set and pick the one that you're interested in. That is extremely expensive because every time you are executing the query, you have to compute the entire result set

Sneha Mehra (01:48:30)  
then you are skipping and then you are picking the things that you requested for. It becomes linearly expensive. ⁓ I'll walk you through that. ⁓ But other way to do pagination is ID limit based pagination. Where let's say you just want to iterate through a table. ⁓ That's it. ⁓ No advanced filters, ⁓ no advanced var classes. You just want to iterate through the table. That's it. In Twitter's case,

Let's say you want to read tweets, as a Twitter wants to read tweets made in last one or so that it can do processing of running trending topics and whatnot, ⁓ doing recommendation, putting it into Twitter speed and whatnot. ⁓ Very common use case for Twitter. So it's very common for them to iterate through all the tweets that made in last N ⁓ minutes. ⁓ In order for you to do that, they need very efficient pagination, which is where you go for ID limit based pagination. First of all, what is ID limit based pagination?

when your IDs are comparable. For example, MySQL auto increment IDs, they are comparable. One, two, three, four, five, six. You know which one came first, which one came second. Right? IDs are comparable. In case of MongoDiki ⁓ and Snowflake and anywhere where you see ⁓ time on the leftmost side, you can actually compare the IDs. You can tell which one is smaller, which one is larger. Right? So now what you can do is instead of paginating by limit and offset, you paginate it.

by ID something like this db.students.find where ID is greater than the last ID so you first get the first 10 rows then you find the last ID from this while iterating you find the last ID and then you do db.students.find where ID is greater than the last ID limit 10 then ID greater than the last ID limited then ID greater than the last ID limited and so on and so forth given that your collection MongoDB collection SQL table are all

physically stored by ID. That's what how databases are stored. They are physically stored order by primary key. This becomes a very efficient operation. Let me show you the benchmarking graph. ⁓

Sneha Mehra (01:50:44)  
So here you see the blue line is the skip limit or limit offset based pagination. An ID limit is the ID based pagination. Obviously I cannot apply advanced wear clauses over here, but if you don't have to, you can get a constant time performance while doing deep pagination. Here you see for smaller ⁓ range or for smaller depth of pagination, your time does not vary much.

But as you go deeper and deeper and deeper and deeper and deeper, you see the time increasing linearly. It holds true for every use case. ⁓ This is when I had a massive disk flush. ⁓ I was so keen why there is a sudden spike over here. ⁓ My disk started to throttle at that time. And I was so happy that I made my disk throttle because of it, there is a sudden spike, but it is actually linear. ⁓ And it was fun to understand that I...

got to put so much pressure on my desk that it started crying and I was so happy. ⁓ I always wanted to saturate my IO and I was able to do that. So here what happened is, you can very clearly see a linear growth in time, but it is constant performance when I'm doing ID based page mention because it is a property of database to keep things ordered by primary key. ⁓ Always, always. ⁓ Okay, by the way, in case you're interested to know, I have a members only video on my YouTube.

last week posted not promoting, in case you'd want to know deeper, I'm covering a bunch of database engineering topics there on YouTube. ⁓ But it's members only. So you may have to make that other purchase there. But I went into how database actually stores like why database chooses B plus three to store it and how it actually physically stores the data, how it makes this operations. So I treating by primary key ⁓ is very efficient because your database is actually physically

storing the data in that exact same manner. So for you to give me 10 from this particular ID, literally log in time, it reaches to that location, then sequential scan after that. Very efficient for your data. No matter, and because it is B plus, it's always log in to reach to that particular ID and then a sequential scan after that. Log in sequential scan, log in sequential scan. That's why no matter how deep you go, your time would not increase.

Sneha Mehra (01:53:05)  
at all. You see a constant time, very evident from this graph, ⁓ very evident. You see a constant time performance always for ID based pagination. Disadvantaged, you cannot apply complex web clauses, but it's okay. That's what your use case is. Right. But that is a very nice, ⁓ but basically that is a very nice side effect or the by-product ⁓ of using Snowflake or any iterations or any ID generation logic where timestamp

is on the left most side. will always see that. So MongoDB I showed you an example holds true for Snowflake holds true for even MySQL. Even in MySQL, it would want to do that deep pagination and you don't want fancy var clauses. Try doing this. You would get much, much, much, much, much larger performance when you are going deeper into pagination. But this does not mean you go and replace all of your iteration logic ⁓ because it has limitations that if you have complex var clauses, you cannot do ID based pagination. ⁓

but where you can go for it. Right. Okay. So before I take questions, let me just cover this. Snowflake at discord. And after this, we'll take question about Snowflake and then we'll talk about Instagram software. So ⁓ here what happens is there is no enforcement. We say that the first 41 bits over here is epoch millisecond. So epoch millisecond is basically timestamp from epoch, epoch is 1st January, 1970\. But if my company is starting then

Let's say 2011 or 2015\. Why? Why should I start my time from 4th January, 1970? Can I not get a custom app? So this way I can get a larger reach. Right? For example, the time that I lost from 4th January, 1970 to let's say 2015\. Those big of an ID range, I would get it too. It would just enlarge my ID. I can go on for even longer and like heads up for

20, 30, 30, for 35 years heads up. That's huge. Like 35 years of IDs, you will get it. Right? So what Discord does, it has the exact same structure, exact same structure as Twitter, but it does that the epoch time starts from first second of 2015 because in 2015 the company started. ⁓ Snowflake is also used by Sony, the TV company, Sony. They also use Snowflake. They have a Golang based implementation. is open source. You can find it on GitHub.

Sneha Mehra (01:55:32)  
Go through the source code, five lines of code of course snowflake and you'll see it's literally bit manipulation. That's it. ⁓ And we also see it in struggle. How they also manipulate those bits. Very easy. Right. But go through that. It's very easy. ⁓ And this snowflake get very nice traction because of its decentralization. There is no need of centralized ID generation service. There is no moving part required moving part. I no external surveys that needs to be scaled up and scaled out and whatnot. It's very counter intuitive.

Whenever you think of ID generation, the first thing comes up. I'll write as microservice that does that. Good engineers never think like that. Good engineers look at problem. They look at the most efficient way to find a solution. Not every solution requires you to build a service. For some reason, everyone is doing that. Very stupid. ⁓ Do things that matter to you and your business, not just because someone else is doing it. And ⁓ everything, every system, every solution is contextual.

It depends on the context of your organization, the expertise of your engineer, the support from your management. But you are the one who can bring that change in case you are unhappy. If your solution is far superior than others, everyone else is proposing microservice. But if you think Snowflake works well for you, just go for Snowflake. You need to have, you need to make data driven decisions. If you show them the data, why would they not adopt it? Right? It's really simple, but just don't fall into this trap of building microservices after

⁓ like for every single problem on there. Microservices are stupid in most cases. ⁓ In most cases, they are stupid. They are an overkill. ⁓ So understand the context and then build what you really need to build. Keep things simple because simple systems scale. Always simple. And if you look here, it is just a library. Normal three lines of leads you in generating a Snowflake ID.

You don't even have to make a network call for that. Because of its decentralization, because of not having to make a network call and build a service and scale it individually, it's just so, so, so efficient for you to leverage. This is the beauty of Stoflack. ⁓ Any questions of this? Before we go to Instagram.

Sneha Mehra (01:57:48)  
Yes, Rajiv. ⁓ I read that graph, which you showed us. ⁓ How did you came to know that the disc is throttling there? ⁓ I happened to have a problem with us running at that point of time, which was exporting my node metrics into somewhere. ⁓ I happened to have that happen. It was all luck. ⁓ So what I saw was my disk IO, I was continuously monitoring my disk IO. Actually the node was monitoring it. ⁓ I heard those metrics being exported somewhere else. It was run.

⁓ on ⁓ all that was run on my local machine, but I was capturing those metrics for some other thing. I was writing a Wikipedia. I was writing a search in Wikipedia. I wanted to capture a bunch of metrics ⁓ and I happened to have node exporter running and it captured it. It showed my disk IO utilization ⁓ as part of my node exporter. You get this disk utilization metric. That's how I got to know it was 100%. Okay. ⁓ One more thing that you told about this. ⁓

Agination logic works for all DB. You talked about B plus tree of my SQL or SQL based thing. ⁓ But then when it comes to noise school part, ⁓ primary key ⁓ reading is ⁓ same for both the person. ⁓ It well. You saw it in Mongolia. Mongolia is a no SQL database. Correct. ⁓ Anyone that gives you sequential scanning, this would hold true for that. The dynamic DB also. But if you're having completely decentralized hash pasting, then you cannot.

⁓ If you have data locality, then you get this benefit. If your data is completely delocalized, then you would never get this performance benefit. So understand how your database stores your data on the disk, which is why knowing database internals is essential for you to build performance ⁓ system. ⁓ would have like, ⁓ everyone would have done with limit also because you know how your database is actually storing the data. You can make those performance optimizations.

without having to worry about is this true or not? Because you know it is how it is told. ⁓ But MongoDB works. ⁓ And just one more example to that. ⁓ We saw in that graph that for smaller value, ⁓ is okay. ⁓ Elastic search gives you two ways to page edit. It's the exact same thing. ⁓ Elastic search is also no SQL. ⁓ It's also a Charlotte partitioned work part. ⁓ Elastic search puts a limit.

Sneha Mehra (02:00:15)  
of 10,000 documents in a limit offset based pagination. You cannot go beyond 10,000. It's a hard coded limit. Try spinning up Elasticsearch server, try finding a query with ⁓ offset, ⁓ skip, ⁓ I'm not sure what they call it. I think they call it offset only. ⁓ 10,000, more than 10,000. It gives you an error that you cannot go beyond this. Please use cursor API to go beyond this. Classic example, where you are iterating through almost all the documents, there more than 10,000. ⁓

It's better to go for a cursor based approach versus going for a limit offset based approach. For smaller values, it does not matter. But as you go deeper, deeper, deeper, that's where these things start to creep up. Okay. Yeah. Harsh. Yeah. ⁓ So what is counter here in this snowflake approach? Same static counter. Atomically increment every time you generate ID. Per server counter. Per server.

All right. Yeah. Which is exactly why I stress so much in the first time we were discussing this. ⁓ It all comes together in all of these discussions. ⁓ Yeah. I didn't understand like how in discord, ⁓ epoch is just a number, right? So even if we start from today's date, how is it like saving us? ⁓ Okay. What is epoch? Let's say I'm starting with, ⁓ wait, let me get the current.

Sneha Mehra (02:01:49)  
The current epoch is 1676784716\. This is the current epoch second. If I add three zero, it becomes current epoch millisecond. Right? This is the current epoch millisecond. ⁓ If I start, if I'm starting my company at this moment, ⁓ I'm losing out on this big operate because this is my first 41 bits of it. ⁓ Right? So now because I have 41 bits only the maximum value that I can store in those 41 bits is ⁓

⁓ Correct. will get exhausted pretty quickly because now I'm not starting from zero. I'm starting from this value. ⁓ I could have gotten that big of a range. If I start from zero, I from this moment and let my time move forward from here. One, two, three, four, five, six, one and so forth. ⁓ we're actually subtracting the start. Correct. So this way we get a massive range handy.

Because the rate at which your time is moving forward is exactly the rate at which the time is moving forward. So this time and the actual time. So you get 35 years of time. If you start today, you get 45 years of time. ⁓ Right? Yeah. Right. So ⁓ it would take more 45 years for you to exhaust that range of 41 minutes. Right? ⁓ you're just gaining that particular ⁓ point. Okay. Now let's go into

How Instagram does it and this is why you should love early engineers, not current, early engineers of Instagram. You should like, they deserve to be millionaires, like multi-millionaires. They deserve. Look at this beautiful, when I first ⁓ read that, I'm like, ⁓ God, I started cursing because this is ⁓ such a brilliant piece of optimization. Like it just boggled my mind, like how someone having a deep knowledge of data, not really deep.

Someone having decent knowledge of database can do this much out of it. So what Instagram did ⁓ is Instagram wanted to use snowflake. Right. So, but what they did is they did snowflake while inserting in DB. So Twitter did snowflake ⁓ at API servers. Instagram does it over here. API server does not have snowflake logic. Database has it.

Sneha Mehra (02:04:15)  
see the beauty of the representation. So what Instagram needed? They wanted IDs sortable by time. Same ⁓ use case as Twitter. They wanted to process, find trending hashtags, push it to feeds of the users and whatnot, like do processing on the recent items. So they needed IDs sortable by time. 64 IDs so that they are efficient on indexes and they did not want to build a new service. Look carefully how good companies they ⁓ are trying

continuously trying to not have microservices. ⁓ We on the other hand are told to build microservices every now and then. That's why I always say good engineers never think in microservices. ⁓ They think in problems and solutions. They did not go for a building microservices for every single thing out there. That's an overkill for most organizations. ⁓ So they had these requirements. So what they went with? ⁓ They tweaked the snowflake a bit. ⁓ What they did? They put first 41 bits at epoch millisecond, but from

first January, 2011\. So they get massive rage to play with now. Then the next 13 bits is DB shard ID. So what they had is the database was already sharded. We'll talk about the sharding structure also in a couple of minutes. So what snowflake for Twitter did 41, 10 for machine ID and 12 for this. What Instagram did 41 bits 13 and 10\. So 13 for DB shard ID and 10 is per shard sequence number, ⁓ static counter. ⁓

Let's first go into and understand what logical shards and physical servers are for Instagram. So what we do ⁓ is ⁓ I hope you got to share it in the period, but I hope you saw that video on sharding and partitioning that I put on my YouTube and shared in the previous. If not do watch it. I'll give you ⁓ a glimpse of it. So what Instagram does Instagram users, Instagram uses a postgres as a database. Right. At 10\.

What Instagram had is Instagram created a bunch of logical shards also called as partitions and physical DB servers. Physical DB servers is actual physical servers on which your database is running, right? Which means your MySQL servers that you are spinning up, right? Within the MySQL server, you create database by firing create database command, right? That create database is a logical partition, right? So what you do is you create logical shards, which is thousands of them.

Sneha Mehra (02:06:41)  
at max to raise to 13, which is 8192 at max and you have fewer physical servers, let's say 10 or 15\. So on each server, you create, create database, Instagram, one, create database, Instagram, to create database, Instagram, three, other server created database, Instagram, 2000 created database, Instagram, 5000\. They're creating multiple logical partitions in your database. Every partition will have the exact same tables with the exact same schemas.

So this is one database in which you have four tables, same tables in other database, but different data. Table same, schema same, data different. So this is logical partition. This is physical server. ⁓ So MySQL server that you spin up is a physical server within which you when you do create database that is a logical partition. So when you run a migration to add a column, you have to run it on ⁓ all the databases that you have, ⁓ all the physical database servers.

against which all the logical databases that you have. When you do create database, we always, when we create an application, we tend to create database and your company name and you create user stable, odd table and this. They do this multiple times, like 8,000 times. But advantage that they get because of this is that now they can take a dump of a database and put it into some other database. For example, for example, let's say

One of my database is hot. ⁓ One of my databases is hot. It can literally, we'll go into this one more time in the fifth week. ⁓ But I'm just giving you a glimpse on what this is all about. They can literally take this small database, logical partition and move it over here. SQL dump commands, they do this, they dump it and they load it in different place. ⁓ This way they can balance. So for them, horizontally scaling a database is easy. They spin up a new physical server.

Take this logical partition and move it there. Why this logical partition? Because now they can like literally dump it very quickly and load it very quickly. If they were not having this logical partition, they have to go through every single row and then extract the row that needs to be moved and then move it. Instead, they just create separate logical partitions, making their movement very easy. So this is what Instagram has as their structure. ⁓ So they have 10 MySQL servers within each one of them thousand, thousand, thousand

Sneha Mehra (02:09:06)  
MySQL databases, which is Create Database Command that you fire, that are being created, which has logical partition and physical database server. Now, this database shard ID that here we are talking about is this.

⁓ They can be at max 2 raised to 13 which is 8192\. And per short sequence number like static count. But now what Instagram does is bridge it. What they do? They use stored procedures. Things we neglected in college PLSQL no one studied that. Stored procedures what are they? We are not going to use it. Like we write raw SQL queries. What Instagram did is something really fancy, really basic. ⁓ I did study this in college but never thought it would be used in this way.

PL SQL, I love PL SQL, stored procedures and triggers and whatnot, but see how beautifully they're using it. ⁓ their photos table. So as I said, there are multiple physical servers in which there are logical partitions or logical databases that you are creating. Each database has same set of tables with same set of schemas. So let's say for each partition that it has table photos, it creates something like this. So create database in-store five.

Create table insta-fi.photos Schema of the photos table, other column, whatever you want to put it. But id column looks something like this. ⁓ id big integer not null default. Default value is when you don't provide any value, the default value is this. This is a function called a stored procedure, a function. insta-fi.next id. So when you're inserting a photo in the table, ⁓ if you don't provide an id, invoke this function.

That's what it is doing. When you don't provide an ID while inserting the photo in this database or in this table, invoke this function, whatever the output is, use that as the ID. Now what this function is? This function is this. Create or replace function insta5.nextid. So every single database has its own next ID function. Because it is on fifth shard, I'm having those 555 suffixed everywhere. So next ID of this data.

Sneha Mehra (02:11:17)  
Next ID or sorry, a next idea of Insta 5 photos table of Insta 5 created database Insta 5\. ⁓ And how I'm defining this function is like this. Insta 5.NextID is the function name, which outputs result, which is of type big integer. And it declares four variables. First is an epoch variable of type big integer. And this is the value which is first January, 2011 epoch time stamp epoch millisecond ⁓ and

Sequence ID something now millisecond big integer something shard ID is fine this five this five this five this five this five same right it's hard coded there this for shard ID five similarly you have to write the same you have to create a stored procedure for every single database that you have right this is for shard five that I'm taking as an example right now this is the business logic this is actually how you create a snowflake ID so what you do every database has a sequencer

Like how does data by generate auto increment IDs? Behind the scene there is a sequencer or sequencing algorithm whose job is to obviously table sequence. are multiple names for that but basically sequence is what you hear the general term. So you create a sequence which atomically increments and spits out the ID or the job is it does atomic increments. Simple. Every time you invoke it, it increments. So you invoke a function called nextValue. It's a standard function. ⁓

insta-py.table-id sequence mod mod 1024 by 1024 2 raised to 10 this is limited by 10 per shot sequence number. ⁓ So no matter how many transactions come in this will be an atomic increment and you do mod mod 1024 so it restricts it to 10 bits ⁓ as sequence ID so it stores it in this sequence ID over here. ⁓ Then you get now millisecond now millisecond you get normal epoch millisecond of your current database. ⁓

Now you want to create what? You want to create Snowflake ID. So what do you do? In result, which is a big integer, you first 41 bits is what? Epoch millisecond since your custom epoch, which is 1st January ⁓ 2011\. Now millisecond minus epoch left shift by 23\. So you compute the time left shift by 23\. When you left shift by 23, you would set the first 41 bits. ⁓

Sneha Mehra (02:13:45)  
to be epoch milliseconds. ⁓ Then the next is shard ID. You have your shard ID which is five result equal to result bitwise or shard ID left shift 10\. So next 13 bits are set to five, which is shard ID. And then you do result equal to result or sequence ID. And every time you invoke this function or every time you insert this photo without passing in any ID,

It triggers this particular function which generates the snowflake ID returns it as result which gets inserted in this table. And that's how Instagram does snowflake ID in the database stored procedure. Highly highly highly recommend you to implement this. Don't just look at it in theoretical side. Very easy to implement. Just start your favorite server, MySQL, Postgres, anything and write this and see how beautiful this implements are. Folks from the previous batch did it and they found it

They are so amused when it first time because it's so easy. It's so easy. It's so efficient because now you don't need to have logic on the API side. It's all offloaded to data. You just do normal insert like you always do without passing an ID. Earlier if you don't buy you to auto increment. Now instead of auto incrementing, it would invoke this function. Problem solved. This is such a beautiful thing that Instagram does. It still does that. You can still see

Insta post having integer IDs which are snowflake generated exactly like this. Right. Okay. Any questions on this?

Sneha Mehra (02:15:23)  
⁓ Rajiv, in this line, ⁓ dot table underscore add it to a sequence. This should be a column on stuff. I read. ⁓ no, no, no, no, this is not a column. This is a, this is a depends on database. ⁓ It's a sequencer. Okay. Okay. ⁓ you can create as many sequence by default, you create an ID column with auto increment, ⁓ it creates a sequence. ⁓ It basically creates a sequence behind the scene with a unique name.

and it assigns it internally. So when you generate new IDs, it just does a next one and gets a new integer and it uses it as the ID for that row that you're inserting. Here, we are just replacing the default one with this. So this is what you would basically custom create as a new sequencer. ⁓ It's an auto incrementing sequence, like how to increment IDs are similar. You don't need this to be accepted.

One more question. We just lifted the logic from API server and just substitute it and put it in a DB, right? Apart from logical partitioning, we just lifted and placed the logic in DB, which means that we are like abstracting the logic inside DB. DB takes care of this. So now you have large number, you can have more than 10,000, 50,000 API servers ⁓ also. No limit. Twitter is limited by that. Twitter can only have two to 10 API servers. ⁓ Because of machine ideas. Okay. Yeah. ⁓

and you have you have have you on the physical server, we have a lot many partitions. And here we are we can have at max 8000 logical databases, ⁓ which is huge number of 8000 databases. ⁓ What else can you do? Scaling actually helped you. ⁓ Yeah. ⁓ So across machines, how are we comparing the ideas? It's just that one if the machine ID one and machine ID two are generated at the same time.

One is given priority is that one is smaller, two is larger. Okay. ⁓ And that machine ideas like we are giving it ⁓ by any machine that can be a separate thing. Another let's say zookeeper is assigning you IDs when your machine spins. Okay. Okay. ⁓ Okay. Well, by the way, that's all what I to cover as part of ID generation. What I highly recommend is implement every single thing that it's easy to implement. It's nothing fancy over here.

Sneha Mehra (02:17:44)  
but you can very easily implement and see these things working in action. It's really easy, but do that. You'll have fun implementing all of this. And trust me, because I implemented, I got much deeper understanding of how they are doing, what they are doing, what advantages, what disadvantages, logical partitioning that we discussed, logical databases, why do they need it, export and dump and load and shard and whatnot. Do that. One life, engineering is all we have, would help us thrive in our careers. ⁓

Good. Any other questions anyone?

Yeah, one more. ⁓ I just talked a bit earlier that if we want to number the IDs, ⁓ we have some logic in which networks are unreliable. The time we insert in database is time we get right where the entry got in. ⁓ but with this approach, it's no fake and all those things to be guaranteed that the insert in DBA will always be more tonic. I mean, if you have a business case where you want to serve that

What do you monotonic? Snowflake is not monotonic. ⁓ It does not guarantee strict monotony. If you need strict monotony, auto increment is the way to go. You can still see some things out of order. ⁓ That's okay. Because imagine social media app like Instagram and all. Do you really care about it? Not much. But a financial application where you are listing the transactions, imagine bank. Would bank use Snowflake ID? No.

⁓ You need strict order in which the transactions have happened. Auto increment is the way to go for them. ⁓

Sneha Mehra (02:19:25)  
Anubhav? Yeah, I wanted to ⁓ clarify my understanding of the ⁓ which is ⁓ insta 5.table id sequence and that is model 0 to 4\. ⁓ So ⁓ my understanding is that for every table we have a sequence id that is generated when we create that table and then we have... No, that is done when you have an auto increment id column then otherwise you can explicitly create any sequencer that you want with any name.

any name. ⁓ And that sequencer is doing this responsibility of ⁓ atomically increment. It's, it's responsible to the menu invoke next valid atomically increments ⁓ it and returns to the value. When you create a table with an ID column with auto incrementing, this is implicitly getting created, but here we don't have that. We explicitly create one with the name that one. I just use table ID sequence. It's a name that I

And can change it. by that, this is exactly the name that Instagram also has given. So that's why I just stick to that. ⁓ But the thing is that you can name it anything that you want. You can create as many sequences as you want. It's just so that you can limit it to just one per table. You can have as many as you want. You can have other columns which are auto increment. It's not a hard limit where it says only ID can be auto increment. ⁓ FII. You can create another column. You can have three columns with auto increments in your table. You can have that.

And the idea is you create a sequencer whose job is to atomically increment and give it to you, which is exactly what a static counter did back in that other logic. ⁓ So we just leverage that. Because we need thread safety. We need transaction safety over there. So we get that because of

Yeah, I understand. Thank you. Akshay. ⁓ Yeah, for the gossip part that we had in the distributor, ⁓ for the gossip, do they use a Lampert clock implementation usually? They did not specify anything for their using, but Lampert clock is very different for it is not used for gossip as far as I know. And gossip is normal protocol that you are ⁓ using. ⁓

Sneha Mehra (02:21:33)  
Which one I used? I used one. ⁓ Forgot the name.

What did I use for gossip? It has very fancy name. I forgot about it. Shit.

Sneha Mehra (02:21:48)  
very gossipy name like if I remember I'll tell you I need to do a Google search for that. So what ⁓ they do it's just they abstracted like they just told in an abstract way that they need to do gossip. How we do gossip is up to you. It can be as simple as making network calls to check what the greatest ideas and then agree upon something and add some buffer to that which both of them agrees on. Right? May not be an explicit algorithm per se. You can write your own simple gossip thing that works fine.

But I'm not sure what Flickr actually did to implement that particular gossip ⁓ or that other logic that did to complete a gossip because it's just a normal thing to say. You'd have gossip protocol. Pick your favorite, write your own. But I tell you, the idea for that is at the end of the gossip that you're doing, you need to know that you both need to have a consensus on the value that the other one would be picking. ⁓ So long as that is there, it holds true.

Okay, and also in the Amazon implementation, if we could like go to that page. ⁓

Sneha Mehra (02:22:56)  
⁓ Yeah, in here like, can't we like, can the servers ⁓ read from a queue that the API generator actually publishes ⁓ to? not? Keep generating the range, keep pushing it into the queue. Let's send a read from the queue, but not one ID at a time, but batch of IDs at a time. And we can guarantee like something like exactly once delivery. So you are doing

you are offloading the atomicity to Q readness. So when you're reading from the queue that is atomic in nature. At then the core property you're still playing with is atomic reads. Correct? You get it from the database, you get it from the queue. Anywhere it's fine. Okay. It's not just, I tend to start as a focusing on the core property that you need. Q gives you atomicity, well, works for you. Right? Database gives you that atomicity, well, works for you. I can replace this with DynamoDB. Atomic instruction.

I can represent with redis also. That's not a problem at that core property. What you need, if you get from any system out there, it's good enough. Yeah. ⁓ I fit in your blogs graph that you showed, ⁓ value shoots up in first when the result ⁓ is equal to 6,000 number of ⁓ records in a search. is good to 6,000 and then comes down. Why is it? Shouldn't it be higher? ⁓

⁓ It depends on what other, what other processes are running on my machine. I did it on my local machine. I did not use my employer's machine. They don't have my local machine. are tons of other processes which are running. You don't know what all things are contending for that resource. That's why you see those spikes. Okay. ⁓ And for the Instagram implementation, which you showed the last slide, table ID sequence variable. I think it should load in cash, right? Not a desk, which means not in the table. ⁓

What cash cash is locally rather than, ⁓ making a call to desk database. It's database implementation. We don't really have to, by the way, it has to be persistent. If not, ⁓ if not, what if your database reboots? It would start from zero. That's wrong. Okay. Yeah. Right. Let it be persistent. Right. And it's database. So database would optimize it. It would do ⁓ a periodic flash and what not. ⁓ Let database take care of.

Sneha Mehra (02:25:20)  
⁓ We just create a sequencer and ask it to generate by this one. Any other question anyone?

No? Brilliant. Next week is all about social networks. I'll upload the recordings and share the questions by the end of the day. Mostly by 3 or 5 today, I need to go out somewhere. ⁓ But next week is all about social networks. ⁓ We'll go in depth of... Harsh, don't spoil the fun, man. ⁓ Wait, I'll pull you in. Wait, ⁓ What I'm saying?

I'll upload the recordings ⁓ by evening. Do read the pre-reads. Next week we'll go in depth of CDNs, how they are used, what they are used. We'll build a bunch of social network services which are essential. I'll now talk about now that three weeks are gone, we have built enough context. Now we'll think from senior engineer's perspective. Now I'll start putting you folks into situations where you have to wear multiple hats in order to implement different systems. Instead of just

Now systems will become more and more ambiguous. You have to come up with requirements. You have to come up with design. You have to do a bunch of fancy stuff there. So a lot of interesting things planned for fourth, fifth, sixth week. ⁓ You'll have quite, quite a bit of fun next week. Thanks folks. See you next week. on, Harsh. Yeah, actually I thought just we are taking question on that stored procedure. okay, Everyone is raising hands. What's up? ⁓

⁓ So in one of the sentences you said that pagination works fine when that column is comparable. ⁓ So does it mean that our columns are primary key or especially those IDs should be numeric? No, that's not true. The strings are comparable, lexicographical strings. If an ID is a string, then your database would be stored in terms of lexicographical ordering of the strings. ⁓ Yeah, strings are also comparable. ⁓ So, okay, both are fine. ⁓

Sneha Mehra (02:27:23)  
All right. That's it. Thank you. Thanks. Simran?

Sneha Mehra (02:27:29)  
Yeah, so I have a question from last class. I ask now? Yeah, so I have a question regarding the CDC. ⁓ So yesterday, you have to talk about it. CDC yet to talk about it. Fifth will go in depth. Meanwhile, read, read about it. Because others would miss it. That's why I'm yet to cover CDC. I just sprinkled it. If I talk something about it, people will get confused. I just don't want that. Right. We'll talk about CDC in depth. I'm yet to give you an introduction to what CDC is.

I would highly request you to wait for your question for until 50 more than happy to answer it when I cover series formally. ⁓ It just that discussion will digress. I'll start splitting out terms. People will get confused. I'm just trying to minimize the confusion that might happen. ⁓ Thanks. ⁓ Anubhav. ⁓ I had a question about Twitter's snowflake implementation. ⁓

If my understanding is correct, the request must be round robin to increase. Not necessarily. Why? ⁓ You're not guaranteeing atomicity. ⁓ You're not guaranteeing monotonicity. Are you? ⁓ Yeah, that was my follow up. If we want to guarantee monotonicity, you cannot. ⁓ Clocks can go out of sync. ⁓ Yeah. ⁓ You cannot guarantee monotonicity in distributed setup. ⁓ That's the thing. ⁓ That's holy grail. ⁓ That's written in stone.

⁓ Because it becomes extremely inefficient for you to guarantee monotonicity when it's in distributed setup. ⁓ So you give up on those concepts, but you still get roughly sorted. So you try to get the max that you can with rough sortability. ⁓ Got it. And ⁓ then we talked about banks and all where we have use cases for strict monotonicity. Then we have to ⁓ reduce our throughput requirements a bit. That's what you do.

⁓ Thanks folks for tuning in. ⁓ Simran, I'll definitely ⁓ take your questions seriously. It's not that I'm not taking it up right now. It's more about I don't want to have people to be confused about new terms that I spit out. That's right. More than happy to do it when I actually form a series. ⁓ Thanks ⁓ folks for tuning in. ⁓ See you folks next week.

—----------------------

8

Sneha Mehra (00:00:02)  
Great week four day second after this will be completely perfect of the course interesting. Okay. ⁓ So we'll continue our discussion ⁓ on ⁓ social networks, but like how we digressed yesterday on CDN side, but that's important. That's very essential for us to do ⁓ today. What we'll do is we'll digress a little bit more in some other direction.

to touch upon something really interesting. Yesterday we fogged a discussion around, ⁓ hey, Instagram private images are not really private. What if we want to make one? ⁓ We'll talk about that. We'll talk about how to make a purely private image sharing platform, for example. You'll get the gist of it. ⁓ Yesterday, we just spoke about, hey, CDN gives us image optimization, why to build one. Today, we'll build one.

And we'll see how it is actually prevented. It's hardly going to be 10 minute discussion, but we'll see how simple it becomes to implement an on-demand image optimization. ⁓ Then we take a break and then we discuss. from now on, given that we are almost 50 % done with the course, from now on, I'll start throwing you all into different situations ⁓ on typically what ⁓ senior engineers are supposed to, how they are.

supposed to ⁓ wear multiple hats, participate in multiple discussions, ⁓ take care of multiple things at once, ⁓ think of extensibility of the system, think about cross-term collaboration. ⁓ Those sort of factors will take care today. We'll understand ⁓ how should we even approach this. So you'll see across three or four different systems that we discussed, I'm forcing you all to think in different lines.

So that is where we would pick up our discussion with tagging photos. Seems like a really simple statement. But when we ⁓ ask those critical questions, we realize that it is not as simple. ⁓ It touches upon a lot of interesting factors. That and we'll end our ⁓ session for today with a very interesting system that we use every day, but we don't give it enough credit, which is newly unread message. I could not come up with a better name. Really sorry for that.

Sneha Mehra (00:02:23)  
But newly entered message indicator is how we use our social network every single day or social media every single day. And we just fail to acknowledge that, hey, this system or this one small button can be so complex. ⁓ But again, through this, what I would like to do is I'd like to touch upon a very interesting high level pattern that you can ⁓ use or that a lot of company uses in production to prevent major outages. ⁓ So these are the four things that we'll discuss today.

But apart from that, I covered this one system just once, designing notifications for Instagram on my first ever quote. I never covered it again because it was out there on YouTube. But ⁓ do watch it in case you haven't. Do watch it in case you haven't because there is ⁓ one thing, one key takeaway from the lens of very naive, because you would realize how naive I was when I was explaining that. even ⁓ people found it amazing, which shows the massive gap that exists. ⁓

One key thing from this discussion notification system is how we bridge the gap of Kafka. ⁓ Yesterday also we saw that how the parallelization of the parallelism of Kafka is limited by the number of partitions in Kafka. How do you get past that? Because you may have to send a million notifications or notification to a million users. Let's say Justin Bieber posted something you want to send notification to 10 or 15\. ⁓

many million new followers he has, there is a notification to those many folks. How do you do that? Because if you use Kafka, it's pathetically slow. If you have limited number of partitions, adding large number of partitions becomes a problem. So what do you do? Which is where I spoke about a very interesting high level pattern where you club Kafka with SQS. So message stream with message broker to get very high degree of final. ⁓ That's an interesting concept. You should definitely watch that.

just for that one bit. Everything else you can now design on your own. But that one bit on why are we making that decision? And basically that's how notification systems are built. If you just draw, and I'll have a message broker on which I'll send message input broadcast. No, that's not what happens at scale. That's not how you manage or handle scale. To build such systems you need, because I had to do it. There was no option for me to send notifications to, not even talking about million. I'm just talking about few.

Sneha Mehra (00:04:46)  
50, 60,000, but in the span of five seconds, we had to do it. So in that case, when we are being this aggressive in marketing, we had to pull that up. Right? So this is what we did. This is how we stitched our system together. A message broker, a message stream with a message broker to get a high degree of pan out. So what's that video in case you haven't, it's that it's the first ever video I posted on YouTube. So ⁓ what's that? If you've sorted by oldest video first, you'll get it as the first. Right? So

That's why don't cover it again, but what's the video to get how we club, ⁓ s ⁓ we club, ⁓ kindnesses or Kafka with SQS or Abitemq to get high degree viral. First ever cohort, March 20th, goes back two years. Yeah, two years, two years back. covered it. ⁓ Okay. So this is the agenda for today. Let's start with the first one. But before we jump to designing Gravatar, let's spend a bit of time.

to understand how images are served. Now, serving images is really important. We all do it on a day to day basis. But how are these images actually served? Because ⁓ we just say images are served. How? When I ask this question, the first thing people say without even thinking about it is, hey, will you see it and to serve it? But someone has wrote that code. Someone has to read image and serve it. What?

How do we serve a static file? Right. Let's start with that because that's what would help you build a shit ton of features. I'm not even ⁓ exaggerating on that. We'll see a real practical demo where this comes in handy because I don't really want to just cover for the sake of it. Right. You'll understand how ⁓ beautiful this thing is. can understand how companies ⁓ mind-blowingly like mind-blowing awesomely use this particular feature. Right. Okay.

So whenever you hit a particular URL, your browser does not know it's an image URL. Let's say you have this URL and you hit it. Your browser has no idea what's happening. What browser would do? Let's say ⁓ it's ⁓ for some reason, four or five quotes back, some people thought that it's because of this extension, the browser knows it's an image. ⁓ No.

Sneha Mehra (00:07:09)  
browser does not have a clue browser is a dumb stupid piece of software. has nothing in it. What you you gave. So browser has a fixed defined behavior for everything. And when I say browser also applicable to apps. ⁓ Right. So always remember whenever something is a platform browser is a platform app. Android app iOS app is a platform. ⁓ Right. So sorry the Android OS iOS

thing, the SDKs that you get, they're all platform, they're all building blocks. So they cannot be very smart. They have to be basic enough so that you can use it with hundreds of other things. So for you to render an image, what do you typically do? You take some URL and you put it in an IMG tag, IMG, SRC, and you put the image URL there. ⁓ What happens when you, when you send this HTML to browser and browser renders it. So browser starts

rendering the XML, the HTML that you said it stumbled upon IMG SRC tag. What browser simply does whatever you provide in IMG SRC tag, it makes an HTTP get request on that particular URL. If you provide this URL, would make HTTP get to this. If you provide some other URL, it would make HTTP get to that. Browser does not have any idea what it would get in return. It is expecting an image, right?

So what it does, it makes HTTP get requests to whatever URL you specify in the AMG tag, ⁓ whatever you get in the response, it tries to interpret it as an image. How do you do that? It's set up by yesterday. saw when GitHub sent, when we uploaded the image on GitHub issues, what happened? We saw how that PNG files bytes were sent in that that header PNG and some RDF data and what not was shared.

That is the file format. The header of the PNG file tells you that it's a PNG file. ⁓ JPG has a header. Bitpap has a header. TIFF has a header. ⁓ Every single format in the world has a header. So what it does, ⁓ whatever bytes you get in return, ⁓ if it's a 2xx response, whatever bytes you get in the response, ⁓ it checks the header. ⁓ It sees, okay, this is a PNG image. So it would interpret those bytes as PNG. So every browser ships with image encoders and decoders.

Sneha Mehra (00:09:29)  
But other mostly decoders, not really encoders, but decoders for every type of format, PNG, JPG, tape, ⁓ SVG and whatever. Whatever you get something as a response and it looks at header and she say, kind of format is this? it's a PNG file. So it would invoke the 3ng decoder ⁓ or basic PNG decoder code and it would decode that and render it as a pixelated image on your browser. That's what is up. So given that

What is what would ⁓ we know now that IMG tag makes an HTTP get call right to get it. So when HTTP get call is made on this URL, what would happen? The request would come to server pointed by this at the end of me in my server. ⁓ It's a web server. What it would do is it would see what's the handler of this part. You may have a handle at say.

You typically what you do is you map your static folder. This is my static folder. Right. As one of the parts of slash static maps to this particular static folder. You configure this on NG next copy that is configured this on springboard, configure this on flask Django, which framework you use. Every framework supports a static folder path. What happens when you say that, Hey, slash static maps to this folder on my local desk. It means that whenever a path like slash static slash something is hit.

interpret whatever is there after slash static as a relative file path go to my static folder on the disk read this path read the file present in this path and send it as a response that's it it does not care if it's image or text or whatever literally whatever is there at that path so for example slash static slash img slash rpg.jpg let's say I map this static folder to this static url

So what I'll do is whatever I get after this is what the piece of code is there. It is not automatically does that. There's a piece of code written in your web server so that you don't have to re implement it again and again. ⁓ What it would do that slash IMG slash at the top of the page. would try to hunt that same path in the static folder mapped on your direct. ⁓ It would find this file. It would read this file, ⁓ read the bytes of it and send it as a response as part of this HTTP request.

Sneha Mehra (00:11:53)  
Once this response is received by the browser, browser interprets it as an event because you put it in the IMG tag ⁓ and it renders it. So this is how render happens. ⁓ Right? So here we spoke about how static folders ⁓ are mapped. So slash static path map to slash static folder on your directory. Right? Where your ⁓ code is deployed, where your static site is deployed. Right? So this is what happens. Right? Now,

What does this tells us that this is what is happening as part of your code. It reads the URL. This is part of a backend code. It reads the URL. It extracts the path after slash static. It goes to that corresponding folder like that corresponding relative path in your static folder present in the directory. It reads the content of invites and sends it back, which means that ⁓ I

Instead of relying on slash static, I can do anything there. I can write my own handler that instead of reading file from local test, it reads file from S3 for example. I can very well do that. ⁓ let's say hypothetically, hypothetically, what I want to do is we saw how slash static path are exposed, but let's say I want to expose a path slash raw. ⁓ Right? So slash raw slash

path. Anything I gave over here as a path, what I want to do is whatever path is over here, I want to go to this S3 bucket at this path, read the file and return that as a response. That's it. So now what would this do? This is ⁓ exact. This is very similar to slash static, but slash static reads that same relative path from local disk. What we just did is we read that same thing from S3.

So which means that we don't have to configure for every single pilot S3, I'll configure one URL over here. We just take a generic path over here. Whatever is given over here, we're literally ⁓ using that as a key of S3 present in this bucket, reading the file and sending it to the user. So with this three, four lines of code, what we just built is we built a proxy for S3. So we provide, we make any HTTP request like this.

Sneha Mehra (00:14:21)  
HTTP color slash slash localize 5000 slash raw and whatever is there after that, it would use this ⁓ as this path read this path from S3 present in the config bucket whichever bucket you have ⁓ and serve this file. Not a picture, I am not designing a new CDM. I am just telling you the possibility ⁓ of how you can ⁓ return images as part of your normal HTTP response.

A lot of people for some reason think that we can only send JSON responses from the web server. No, you like you can do lot more than that. And one of the key things that you can do over here is to send the images from your HTTP handlers that you have. So I just literally just exposed a path, exposed a ⁓ URL, sorry, exposed a route on my web server.

which literally just reading files from SRE and serving it to the It's like, why are you doing it? It's more about understanding what you can possibly do. Now we'll see practical applications. ⁓ Okay. So given that now it's in our control, ⁓ whatever input is it, so it is not just that the image ⁓ already, what does this tell us? It tells us that it is not ⁓ really needed.

that your image already exists for it to read and serve. We are just making a network call to S3 and reading it. ⁓ I can have my own custom logic over here and then what matters is the bytes we send to the end user. That's it. If we are sending a properly structured image in the response, byte serialized version of it, the browser would be rendering it.

Right now, where do we see this in action? I'm sure you would have seen a lot of images on social platforms where people share about their GitHub repositories or people share their GitHub repositories. Right. Now, let's take a look at how that works as a very interesting side effect ⁓ of understanding this. I'm just sharing my complete screen. I'll show you them. So I have this project DiceDB. Right. Now, if I share this project.

Sneha Mehra (00:16:45)  
on social platforms. I want like very interesting things about this project like because when you share it on Twitter or LinkedIn or Facebook, you see a very nice way of representing it. ⁓ Do we have any people who have shared? ⁓ I think I want to turn back. ⁓ I'll show you. ⁓

Sneha Mehra (00:17:16)  
See, this is not this. This is the one I'm talking about. ⁓ So I shared the link. DiceDB slash dice, github.com slash dice. What it render? It rendered a social snippet of it. It has ways to render a link. It patches the data information, the title, the site name, the description that you provide ⁓ and the image. Now this image is not some generic image. This image is specific to DiceDB.

Right? DiceDB slash dice. This is the description, number of contributors, number of issues, number of discussion, number of stars and number of folks. This image. This image is generated ⁓ on the fly by GitHub, not every time, but on the fly by GitHub. So let's take a look. And I'm happy that I have five notifications. ⁓ Okay. So, okay. So when I share this thing, if you go and by the way,

How is this image rendered? This is part of open graph protocol called OG. So you in the HTML that you serve, you have to specify OG colon image, whatever the URL you provide in OG, OG colon image, that URL is rendered as an image over here. And that's what happens. So if I go through the code of this and I do view page source and I search for OG image, you see this content, this thing.

And now if I do paste, you see this exact same image. ⁓ Now what is this? What is this particular path? If you look carefully, it does not end with ⁓ .jpg, .png, anything like that. ⁓ But what it is indeed returning is it is indeed returning a valid image as a byte response, like reading the bytes of the image and sending it back. ⁓ But now,

This does not mean that this image is stored on S3 somewhere by GitHub. Imagine GitHub unnecessarily creating images on the fly. Because if you look carefully, the data that you see in the image, it's very customized to what a current state of information is of my repository. If you look carefully, 4KAR 74, 275 contributors and what not. ⁓ This is very recent. If in a day it changes,

Sneha Mehra (00:19:38)  
This would change again. If you let's say become a contributor and ties DP, the count will change to 11\. Maybe it would take one day or two days to reflect, but it's very recent. It's not that it is pre-generated by get a lot of people think everything is pre-generated. No, it is not because imagine GitHub has millions and millions and millions of repositories. ⁓ Why would get up spending the computation power generating one for each every single day?

So what happens is you generate image on the fly ⁓ on demand, whenever it is requested. Right. And this is what you do. Right. So here, what is happening is if you look carefully at the path, this part, this part shows us that for you to read, if I did just take this image and put in the IMG tag, it would render this image in my browser without any hiccup. Right. So what is happening is this is literally

The ⁓ route handle that we wrote the one that we wrote slash raw slash raw slash path. is kind of that. And it is put behind CDN. This is another CDN that it has open graph.githubassets.com. Right? This is another CDN that they hosted similar to how we saw edge.adv.me yesterday. This is ⁓ one more CDN. This is CDN on the GitHub site. We saw one on the GitHub as well. User ⁓ assets, GitHub something, GitHub user content.com.

Right. This is another CDN that they've configured. Right. So here this CDN looks very specific to OpenGraph. OpenGraph OG that I was talking about that is OpenGraph. So this is what is you are invoking. ⁓ How this is built is simple. You ⁓ have this key would remain constant for a repository or something. You can generate it on the fly as well. But the idea is pretty simple that

When this URL is sent from the backend to the frontend, frontend renders it in the img tag. This request goes to CDN posted at opengraph.githubassets.com. CDN says if it has the file or not. If it has cached the file, which is image file, good enough, it would serve it back to the user, which is what is now happening. See how quickly it gets loaded. This means that it does not have to go to there every time. It doesn't have to go to backend server every time. ⁓

Sneha Mehra (00:22:01)  
If CDN does not have this file, what CDN would do? It would make call to origin. What is origin? Origin is the API server of GitHub which handles this request, which is generating this image ⁓ on the fly. That hey, this is where my title should be. This is the font of my title. This is where my repository icon should be. This is what I would have fetched the key statistics about repository and put it over here. This is a template. No matter which image you or which repository you pick.

you would see this exact same format. It is a templatized version. So what is happening over here is the API server of GitHub, when that particular API call is made by the CDN, it is generating that image on the fly using image magic or any library or like there is a library called PIL, there is a library called Image Magic and tons of libraries out there to generate images on the fly. It generates image on the fly and responds back to whom CDN metadata requests it sends response to CDN.

CDN gets the response, CDN caches the file. And then with some TTL, it keeps it there and serves it to the user. The subsequent requests are served on the CDN. After a day, when that thing exhausts, ⁓ when the file expires, the next request that comes in, there's a cache miss on CDN, goes to backend, a new image is generated on the plate, is generated on demand.

So it's not that GitHub is creating and storing this image for every repository order. is doing it on demand as an inventory student because the image itself contains the recent information, which is why GitHub cannot pre-generate those images and store it. And now you can see this feature at hundreds of places. There's a company called Banner Bear, ⁓ always specializing in doing this. ⁓ This is how you make the social sharing stuff

really easy, really interactive, really real time. for example, just imagine if Twitter tomorrow wants to do this, I'll give an example. If Twitter tomorrow wants to do this, would you see LinkedIn quote unquote influencers taking a screenshot of their Twitter and point dot so and creating a very nice, beautiful image and sharing it on LinkedIn. Imagine that as a native feature of Twitter, where when you share a Twitter URL, the OGM that is generates

Sneha Mehra (00:24:18)  
It looks very beautiful like this. What it could do is when that request comes into CDN, CDN doesn't have the data. goes to backend. Twitter's API server would generate that image on the fly using that same template, ⁓ like, or using the templatized version of it. It captures the tweet, puts it in it, nicely formats it and sends it to the response. Right? Cache on the CDN, sub to the user. Twitter can build this as a name. So when you share your Twitter link, it can automatically render your tweet.

in a very nice looking image as a social graph and a very social image. ⁓ And that's ⁓ very powerful when it comes to experience, when it comes to designing great user experience. Because user sees personalized things. User sees something which is very recent. It just does not see ⁓ the static wrap up, GitHub, and something, something, something. They see something which is valid. What's the current likes and shares and comments and retweets and reposts and all those things. ⁓ Things that matter.

right there in the image which is shown. And you can literally build this in hardly one hour. Hardly one hour. Lest a local prototype. Highly recommend you to do this. Highly, highly, highly recommend you to do this. It's not difficult. Install Python. I worked with Python for in order to build this. Java also might have something or whatever language you're using would have something. But pick a library that helps you generate images on the fly. Typically,

PIL, Python Image Library is the best out there. It internally uses image magic. It has a C++ binding to that. ⁓ But it just gives you native Python interface to create images on the fly. And just do it. Expose an API endpoint that does this. You can generate the same thing for any GitHub repo in the world. This exact same thing. Just make a call to GitHub, scrape that information out, create image on the fly, and serve it via your local host. ⁓ You learned so much from doing that. ⁓

And this is how images are actually like these sort of dynamic images ⁓ are generated and so giving user a great experience earlier. If you remember, ⁓ when you shared images, you have used to render a very pathetic avatar of that user. So if I shared images, so my ugly image would just come there on the screen. Big flashy image. That's a very, that was a very poor experience. So GitHub implemented this to make

Sneha Mehra (00:26:42)  
sharing of repositories on social media, very pleasant. And through this, you could see that how whatever we discussed up until now, whatever we discussed across these many weeks, every single thing, can see it in production used by some of the other companies. Nothing is theoretical as I always say. You can prototype literally every single thing we discussed. There is no such thing that a

I need 50,000 servers to do that. No, you don't. You can, you just want your local machines, spin it up and try it out. Right. Nothing with practical. ⁓ Okay. Given we have built enough context on this, let's talk about that. Now this becomes an easy problem after this, after we design, after we discuss this, we'll take. Okay. So what is gravitar? So gravitar is a very interesting concept ⁓ on gravitar. What happens is you would have seen.

that on some websites, even though you do not sign up with your ⁓ Twitter or Gmail or Microsoft or whatever account you are using, it is still able to render your image somehow. And which is where Gravatar comes in. You might have logged into Gravatar in order to enable that, but the idea of Gravatar is pretty neat. So what Gravatar does ⁓ is ⁓ it creates your single embeddable ⁓ URL

for your profile picture. For example, if let's say your email ID, let's say my email ID is arpit.directgmail.com. It's not that, but let's say it is arpit.directgmail.com. So what I could do is I could go to gravitar.com, log in via arpit.directgmail.com and I can upload a bunch of profile pictures there. I can mark one of them as active. ⁓ Now my URL, what it would generate,

It would not generate. It's basically very simple. It's an MD5 hash of my email ID. it pretty much is constant. Pretty much. It is definitely constant hashing. ⁓ If you pass the value rp.gmail.com through the same hash function, get the same output always. So it's gravita.com slash hash of my email ID. ⁓ So when I go to this URL, this URL actually renders my current profile photo, which I can literally

Sneha Mehra (00:29:06)  
put in any IMG tag, something like this. So, ⁓ https://gravata.com slash zero E-A-F-D-1-7-2. Let's say this is the hash of my email ID. I think it's something like, this is something very similar to this. So, https://gravata.com slash zero E-A-F-D-1-7-2. When this is put in IMG-SRC, it actually renders the image in the browser. It's not that this takes me to somewhere and that renders the image. No, this is

actually returning the image as bytes ⁓ over here. What we just looked at when we reading from S3 and survey kit similar. ⁓ It is taking, it is literally the single URL to render your profile photo. Now what happens is let's say every company in the world adopts it, which means that whenever you sign up on any platform or any social platform of the world, what you get is what they could do is they have your email ID.

They can pass that email ID to the hash function and use this to render your profile photo wherever they want to. So in the post and the profile page and wherever they could just use this one URL to render it. ⁓ If every company starts supporting this, what would happen? If tomorrow you want to change your profile photo, you don't have to go and upload on every social platform. What you can do is you can just go to gravitar.com, upload your new photo, mark that other one is active.

When you mark that other one is active, the next request comes in for this URL. Go to Gravatar backend. It will find that other photo is active. Pick that up, turn it out. Every video profile photo is changed without having without any social media have to do any social platform. Any social network have to do anything. You don't have to go to every website and do that, but they don't obviously do that. But it's a very novel concept. What we'll do today is we'll build this very safe. What we want to do is support

this efficient ⁓ building around gravita.com. It's a really simple system that you can very well build in three or four hours. Not really good. Scale is difficult. Building a product is not. that's why I recommend build prototypes as many as you can. Right. So what are our requirements? Our requirements from Gravitar. What we are designing is first user can upload multiple pictures on Gravita. Second user can mark one of them is active.

Sneha Mehra (00:31:32)  
The one which is active whenever the API is requested, when one of them is active, it returns the active one in the response. This is what we have. So let's kick off our brainstorming. Then let's see how we should approach this problem. ⁓ Given that we built so much of context, let's directly kick off brainstorming. How would you start? How would you approach this problem? Raise your hands, I'll poll you in.

Sneha Mehra (00:32:08)  
Okay, let's start with database. Let me simplify. Let's start with database. What kind of table you have? What kind of schema would you have? What would you store?

Yeah, the hash of the email ID of the user. ⁓ What would be your table would you have with each table that you have? ⁓ With the user table. ⁓ User gives his email ID and we store the hash for it. Will you store email ID or not? The actual email ID.

Sneha Mehra (00:32:41)  
⁓ Sure.

Sneha Mehra (00:32:45)  
Why if you are just using hash y to store actually militee. ⁓ Yeah, right. No, it's not right. How would they log into Gravatar?

Sneha Mehra (00:32:57)  
⁓ We can directly ⁓ hash their email id. ⁓ ID is a derivable attribute. ⁓ So storing one of them makes ⁓ sense. Why all social platforms take your email id and store your actual email id instead of storing the hash of it?

Sneha Mehra (00:33:23)  
No, they can also do that. would save them a lot of space. I think the use case there is a little different because you might need to display that email ID somewhere, but at Gravita, you don't need to this. ⁓ Why not? If you open Gravita, you go to the profile page. You want to see your email ID. Got it.

How are you authenticating people there? While authentication people pass it because in case of hash collision you are locking in some other's account also problem. Correct?

It's better to store email id. Right? Okay. So id, email id. What else?

not hash probably because hash is a derivable attribute okay hash is derivable attribute so i'll just write hash and cross it off okay you're not you're deliberately not storing hash because it is derivable okay any other thing that you would add to user that's fine user stable is fine ⁓ which other

Sneha Mehra (00:34:24)  
I think that just one, I don't know whether it should be a different table or within the same one to store the image parts, ⁓ not the active one, but just the image parts. How would you say multiple part in a single column of relational data? ⁓ Yeah, we won't be able to do that. ⁓ So let's say you have a photos table. Yeah. ⁓ In with what do you store?

the email ID of the user ID. ⁓ Obviously the photos ID.

Sneha Mehra (00:35:04)  
What is photos ID? The primary key the photo step? The primary key, yes. ⁓ Okay. And what else?

Sneha Mehra (00:35:15)  
The parts of the images, photos that you Do we really need to do that? ⁓ Recall yesterday's discussion. Can you not derive it?

You can write, but they would be multiple. didn't think of the discussion, but they would be multiple photos. ⁓ So there would be multiple rows over here. What's the problem? And obviously you need to start multiple photos, right? Because user can upload multiple pictures on it and mark one of them as active. ⁓ So you need to render all of them. Let's say you're pretty good it. You need to render that photos. ⁓ Right? So all the photos uploaded by the user will go over here. Right?

ID, user ID and what else would you store?

Sneha Mehra (00:36:01)  
The photo which is active. Okay, so let's call it isActive.

Anything else?

I think that's it. it. Hey man, let's write queries. Let's write queries with which you would want to find all photos of a user. What would that query look like?

because on the UI you would have to render all the photos of a user. What would that query look like? So maybe select a star from photos table where user ID equal to given user. ⁓ We may need to filter it by our ejective if we want to sort the active ones. ⁓ No, no, we'll show all of them because only one of them is active, right? ⁓ User ID is equal to something. Okay.

Not sure if a star is needed if we need only the ID because we are looking for that. Let's select that. Simplify. ⁓ Here, ⁓ do we have the user ID of the user?

Sneha Mehra (00:37:06)  
Look at this. No, we need to kind of join it from the user stable because we do not have. So let's join. So select star from photos where ⁓ select star from photos. You join it ⁓ on ⁓ users ⁓ on photos dot user ID equal to ⁓ user dot ID something users dot ID where

user ID equal to this. Correct? Yes. Okay. But you still don't have user ID.

Right? Because what you are getting over here is a hash of email ID. Yeah. So that, that we need to ⁓ kind of a ⁓ reverse hash and get the ID because we are not storing the hashes which are derived. So that drive part will come in computation. So, so what will you do? So what's your query now? You tell me the query, exact query. Okay. So let's just start from what was this. We are, ⁓ we can write a decrypt ⁓ of ⁓ input, which is coming as has.

equal to you cannot get email from the hash. It's not encryption. It's hashing. So it's lossy. You cannot get the actual email ID back. ⁓ Actually, in that case, think, yes, got it. Because there will be a collision. ⁓ Two different things can go. So we have to have a hash stored because in it, we cannot go in a reverse way from email. ⁓ But can you not just do pass the hash function where hash of email ⁓

email is equal to the one that you provided. Yes, that we can do hash of email equal to ⁓ users. Whatever your input is. can work. But what's anything wrong with this?

Sneha Mehra (00:38:58)  
The wrong thing, can be like two different email ID if ⁓ they generate the same hash. That we have accepted. ⁓ We have to accept that, right? That two people having same hash would not work because that's what we are going with, right? So if two people are the same, how will you disambiguate anyway? ⁓ So ⁓ assume that hashes are ⁓ you mean? ⁓ Actually, ⁓ it's a bit of more computation added in the query. Why? ⁓

doing hash of email every time each query you are adding this computation we can try to avoid that but hash is very lightweight very lightweight yeah i mean it's lightweight so we can say it okay not but ⁓ at least i see this as a one problem that it's not a simple query we are adding some computation on top of it okay ⁓ let's say i have 10 000 rows right how many rows would this scan this query scan actually scanning is

pretty bad because you have to do a cross join of both the tables where the user IDs matches. ⁓ will, if you have index on user IDs, then you would have, ⁓ you have index on user IDs. So you get index of photos, you get index of user. ⁓ And then you do a scan. Actually it is pretty bad in evaluation because we do not know the user ID. So a well-functioned try to figure it out the ⁓ user ID. ⁓ no, no, it will not.

⁓ It has to do the whole whole scan of looks like if we have the exact user ID that would have ⁓ reduced the ⁓ scan part of table up to those user ID. In this case, since we do not have that handy, we have to go and figure it out this value, do the whole scan and then do the join and then get those filter. That's how you can think of. ⁓ Why, why does it have to do that entire scan?

because of this hash. Yeah, for each email ID, we have to ⁓ create that ⁓ hasing and evaluate that hasing value. ⁓ Now you got the email ID for that you have. ⁓ So once you do the whole scan, you read each and every row of the table for each email, get that the hash. Now out of those has you will do the search or like filter it which email ID we are talking.

Sneha Mehra (00:41:24)  
Now once we have that, then you can do the join only for those rows which are in ⁓ matching in the very close. Yeah. Yeah. Brilliant. This is what the pain point is. Okay. Let me contradict. Thanks so much for putting that in such a very, in, in, in a really nice way. ⁓ yesterday we discussed now today, will contradict that yesterday. We discussed that, whatever is derivable, we should not store that.

But if you think carefully over here, this query ⁓ is very, very, very inefficient. Because the where clause itself contains a computation, which means that ⁓ it cannot leverage indexes over here because unless it goes through every single email, it cannot say that it would have to the hash of it and then find which matches with your

given value. So it would have to literally because there is no way for your relational database to know, okay, this is the email ID that is pointing to that, that would evaluate to this hash. You cannot leverage indexes over here. This will result in a full table scan every single time. This is why you need hash column over here, because then you can index it and see where ⁓ users.hash is equal to the one that is passed.

This makes a query order one, almost all. And because now it's very pointy. It's a very pointy very, because now what it can do is it could literally just do the user that hash equal to something. It would filter out that one row from it and then join, get the user ID, join it over here. It's very quick, very, very, very, very quick. So it's not always true that what is derivable you should not store. It depends because if it is slowing things down for you, you have to

store that redundant part. This is where we are adding this one redundant column. are actually, whenever a user is getting created, we are not only storing the email ID of the user, but we are also computing the hash of the email ID and storing it and creating an index ⁓ on literally every single column. ID is the primary key, email is a unique key, hash is a unique key. But we have to do it. We have to do it. Otherwise this query, which is more

Sneha Mehra (00:43:51)  
which is very common of a query ⁓ that would be very expensive because there is no ⁓ one efficient way for your database to figure out which all rows it would need to converge on because there was an evaluation in the where clause itself, hash of email. Right. This implementation detail I want all of you to focus on. Again, as I say, drawing boxes is easy. This is where the magic happens. This is what makes you

Like when you scratch the surface, you see the ugliness of the system. ⁓ So devil, sorry, the devil lies in the details. Like we all have heard it. This is a classic example of that. ⁓ Unless you write that query, unless you figure out how that would execute, you would think, we don't need this. But then you realize that your queries ⁓ not really performant enough. Then you add. Right? Okay. So this is what I wanted to very ⁓ key point. wanted to highlight that.

Right. This is why that color hash color is really important. Okay. So we talked about rendering all the photos of a person. Now let's talk about marking someone as marking a photo as active. So then what would be the query to mark a particular photo as active? ⁓ So in photos table, ⁓ we can keep like, no, at ease that two column is already there. Right. ⁓ that row we can keep like no flag true or false. ⁓

⁓ So that's okay. That's a boogie. So what would your query look like? Okay. So this is just to pull ⁓ only active photo, right? No, it's marking a photo as active. Marking photos as active. Yeah. Okay. ⁓ Then it's an update query. ⁓ Update photos.

set is active equal to ⁓ true where user id equal to

Sneha Mehra (00:45:50)  
Hmm.

Sneha Mehra (00:45:55)  
And ⁓ photo ID also we have to pass. ⁓ On VH photo, when user click to make it active. Do you really need, because if ID is primary key, do you need to pass user ID over here? No, not required. Okay. Not required.

So you are passing this. Okay. What's the disadvantage of this query? What could go wrong?

Sneha Mehra (00:46:21)  
you're asking me. Yeah. ⁓ You gave this query. What could go wrong?

Hmm.

Sneha Mehra (00:46:32)  
See, this is what I love. ⁓ This is such a simple query. You could see what's wrong in this query. There is a big problem in this point, a big problem. So we'll try to spot that.

Sneha Mehra (00:46:51)  
Okay. So if you give user ID also, ⁓ then the scan will be like, less on the table. No IDs primary keys can would anyway be logger. So that's very efficient anyways. Okay.

Sneha Mehra (00:47:12)  
three line query, ⁓ one of the most simplest query that you could write ⁓ almost hello world of update. And that is where like ⁓ need to think about all this stuff when we implement. And this is one aspect that almost every engineer, even senior engineers neglect, which you should not, never.

Let me pull in Jyothik. Jyothik Dargav, what's wrong in this query? So you don't mark the previously active image as inactive. That is, think, one issue. OK. So how do you do that? So let's start with that. That's a good point. You're not marking the other previous one as active. So what's your query now? So a very simple, I think, highway would be just to get the currently active thing and mark it as inactive or just ⁓ do something like that. ⁓

But I feel that is like adding getting active marking that is inactive and marking others active. ⁓ Yes. ⁓ Simplified. ⁓ I mean, I can think of sort of like a hack to do this. ⁓ Just have a counter sort of like a counter or like the update. ⁓ I some just stay with me for a minute. Just have like a counter on the photo with on each row and increment. ⁓

No, that's not possible. Try ⁓ writing an update query for that in order to mark the current one, the current active one as inactive. Yeah. ⁓

Yeah, initially I was thinking it might be possible to do it in one query, but I'm not sure if I'm able to get that. It would not be. ⁓ That's not possible. So yeah, you'll probably want to ⁓ make like, think ⁓ the naive approach would be to have three queries then. Why three? You. ⁓ Okay, actually you can do two queries. You can just mark all as false and then mark one as true. Like. ⁓ Yeah. ⁓

Sneha Mehra (00:49:14)  
⁓ What's wrong in this? ⁓ So we did exactly what you said. We marked all of them as false and then we marked one as true. ⁓ What can go wrong?

Sneha Mehra (00:49:38)  
that share what can go down. ⁓

I mean, I had a solution that we shouldn't mark everything as false just add and in the where clause and see is activist to do so it will only pick up the row which is currently active and then mark it as false. Okay, so you're saying that user ID is this ⁓ and ⁓ is activist. Okay, what is wrong in this?

What could go wrong? ⁓

Sneha Mehra (00:50:19)  
It seems obvious, but I want all of you to basically explicitly tell that. I mean, the other query should ⁓ should basically run only when the first one is completed. Correct. Correct. Correct. So what do you do? ⁓ there's a dependency. ⁓ So basically what we can do is when we get a response from the first query that this has been executed.

then you send another or we can make it as a transaction something like that. Transaction. Yeah. ⁓ It sounds obvious, but we don't call it out explicitly. And that's the problem because what if you update this and your process crashed? ⁓ Then every photo is marked as inactive and you don't have any active photo. Facebook is trying to render a profile photo and there is 404 not found coming up. Right. So

This needs to explicitly be wrapped in a transaction.

⁓ Very simple thing, but people skip it. We don't, we should be very focused on every intricate detail, especially in the simple systems because it seems very obvious, but it is not. Okay. Thanks for, ⁓ spelling out transactions for that. ⁓ What's wrong in this? ⁓ What can go wrong? Now we also had a transaction.

once we have added transaction it becomes atomic in nature right ⁓ so I don't see that we will fail in concurrent request ⁓ so if you don't so concurrency taken care of what other factors can go wrong

Sneha Mehra (00:52:20)  
Look at the query very carefully. ⁓ Very carefully. Especially the second query. That's a hit.

Sneha Mehra (00:52:38)  
This is the simplest update query I could have written. What's wrong in this? There is a big problem in this query. Big problem.

Sneha Mehra (00:52:52)  
I'm not able to. ⁓ That's right. That's right. See, it's very hard to spot because you see it's like literally you wrote one print up and you say, yeah, why my program is not running. ⁓ It's literally that it's literally that like figuring out what's wrong in this. Okay. ⁓ yes, Priyanka. What could go wrong in the second. ⁓ So in the first query, said, you know, we are setting the active flag to true. Whatever is

existing active we are already setting it to a false. ⁓ Yes. ⁓ And ⁓ so now we have to set another one to true but again you have to figure out which photoid is that. You have you are you are already passing it. That's fine but what ⁓ is missing over here. There's a big piece missing over here.

User ID. Bye.

If you have the photo ID, why do you need to pass user ID? Photo ID is the primary. This is the primary key. Okay, yeah.

Sneha Mehra (00:54:02)  
I'll move to Mohit. Mohit, God. Yeah, I user IDs required Arpit because a user should be allowed to change only his or her photos, not everyone. Brilliant. ⁓ Brilliant. This is why. ⁓ You need this. need this. Imagine ⁓ someone from the front-end passing a very random photo ID. If you don't have this checked, you are marking someone else's photo as active. You are upgrading your own photos as inactive.

But someone else's photos active. So your photo is still an active like you're like this. None of the photos active, but I'm looking someone else's another photo is active. creates a problem. So you should be you as a user should be allowed to update or mark a photo as active only yours only your supporters active not someone else's security correctness of the system. And it seems and this is why I love covering this problem segment because this tells you that

When you think of implementation, how a very simple use case, really simple use case has so many hidden intricacies. This is what makes you a better engine. ⁓ Going into these details, like understanding why do we need a certain field, a certain attribute, ensuring correctness of the system in all cases. Here we just saw that in all possible cases. Now, with this wrapped in a transaction,

With this thing in place, with having to store redundant value over here, we are ensuring our system is efficient ⁓ and correct ⁓ no matter what. It's such a simple use case, but I just wanted to highlight that part through this design that it seems ⁓ very simple at first. When you go into those details, you realize ⁓ something's fishy. Great. Okay. One more thing.

See, that's right. It's, it's, this is exactly like code reviews. When you write five lines of code, will get 50 comments. When you write 500 lines of code, you'll get, looks good to me. Same thing. ⁓ When you have a simple system at hand, that's where people would start to basically nitpick every single thing. ⁓ Okay. There is ⁓ one small minor, very minor optimization that you can do very minor.

Sneha Mehra (00:56:24)  
can anyone spot it? Very minor, like it's literally insignificant, but still if you do it, it's slightly better. Heman.

Sneha Mehra (00:56:50)  
⁓ Any hint for are you talking about the first part of query? Okay, let's let's write a query. Let's have with it. Let's write a query to pick an active photo of a user because that's what would happen when you hit this URL, right? When you hit this URL, you have to pick the active photo of the user. What your query would look like. So select a star from ⁓ photos ⁓ where ⁓ has of

⁓ has what we have provided as the user ID ⁓ and ⁓ join on ⁓ user ID from photos table.

Sneha Mehra (00:57:29)  
So exactly this query, but just with one more word class. Yes. Okay, let me just copy this. ⁓

So where user hash is this. Yes and user ⁓ user photos.id yeah photos.id

and one more thing is no no no no not photos id but photos dot is active photos dot is active true

Sneha Mehra (00:57:59)  
⁓ Okay, okay. This is your query. This would give you for all the ⁓ one photo because photo active true only one we ensure the correctness over here user hash and we join on this user table. Okay, can we optimize on something in this query? Very micro optimization. I'm literally nitpicking, but just it. ⁓

Sneha Mehra (00:58:23)  
Are you talking about left right join? No, no, that's okay. Inner join would work exactly the same performance as left, left inner join. So it's fine.

Sneha Mehra (00:58:35)  
future has equal ⁓

I can see only as I well part first is the joint thing. ⁓ Then the ⁓ conditional filtering, conditional filtering. Once we join on user ID, ⁓ we've got all the rows for user ⁓ matching user ID. then ⁓ user has ⁓

So I think eval should work not on that. Where the first is user has equal to user ID. So it will figure it out. What are the rows with the user? Which will go one row from user stable. And then I got the rows. So I got the user ID. Now that user ID will do the join on photos, which will give N number of, let's say rows, which are there for each user, ⁓ for that user. ⁓ And on that you are saying,

⁓ Which one is true? ⁓ It ⁓ seems very, but ⁓ I'm extremely nitpicking over here. ⁓ In practical use cases, you might see very minimal performance difference, very minimal, but ⁓ it just, I just wanted to highlight that. ⁓ just it's, it's a very, it's a stretch, but I still wanted to highlight that optimization. That's why. ⁓ Okay. Yeah. ⁓

Sorry, sorry, go on, go on. ⁓ Okay, should we... The only thing is ⁓ the only ⁓ extra work that we have to do is on ejective true. So is there any ordering ⁓ or something which... ⁓ It's anyway indexed because you would be ⁓ anyway narrowing down on ⁓ that one row that would be indexed. So it will then index lookup very fast.

Sneha Mehra (01:00:33)  
So that's, that's a lot of problem. ⁓ Okay. So I was, I was able to think that only, ⁓ yeah. Maybe I'm giving up. ⁓ Yeah. ⁓ Go ahead. What would you optimize? ⁓ I can't optimize something on the right hand side, but left hand side one optimization I can think is don't update if it is exactly is true. ⁓ And one, one, one described that. ⁓ Yeah. And one more thing is if I just zoom out, ⁓ I would store active photo ID in the user's table.

because it would be a ⁓ heavy system. So writing this whole transaction would get us out. Which one? This one? I'm saying in the users table, I would like to store ⁓ active photo ID. ⁓ gotcha. This was the one I wanted. So now what would it do? So basically, like, see, on a high level, this is a read-heavy system.

The whole transaction thing, the whole photo scan table, everything would like this is a single scan in the photos table. So we are saving everything. All of these transaction queries, all of this will be nullified.

⁓ You can very well know, but you would still need this transaction. You would still need this because now you would want to update on the other thing. You would still need those transactions. Your queries would change. But in this case, for this query, you are avoiding the joint. Yes. No, but my question was. ⁓ But now when you're updating, you would also need to update in that same transaction. Update user stable. Set. ⁓

Active Photo ID. ⁓

Sneha Mehra (01:02:17)  
equal to the one that you get from the above. So select from this and update over here where user ID equal to this where user ⁓ ID equal to this like whatever the user ID was there. This is what happens is if you maintain an active photo ID over here, you are more frequent query. The most frequent query on your platform would be this right getting the active photo ID and then rendering it on the UI. Given this is the most active query.

Here you can avoid this joint. Although that's what I said, it's I'm very nitpicking, extremely nitpicking business because it's not going to be a massive performance difference, but we can still avoid this very small joint that we are incurring by just replacing this query to be select star from photos, select star from users. Their ID is equal to this. That's it. Because in star you get active photo ID, active photo ID and user ID computation. You can create S3 path, read the file and send it to the user.

as simple as that. ⁓ So you don't need to complicate at all. ⁓ One question, why do you need to like, is there any other query that we are supporting? Why do we need to store is active in the photos table? No, it just that we have multiple photos. You have to mark one of them as active, right? That's what I mean. If you are, if you are supporting a query that is gay on a photo, tell me to active or not. You can join with the user, right? But that's not, now we can do that. Now we can do that.

Yeah, now that we are storing it over here, so there are two designs. Either you store it as active in the photos table or you store active photo ID in the users table. This is what happens is your most frequent query in your system does not have to go and join two tables. It is literally a primary key lookup. ⁓ As simple as that. And now when you're marking a photo as active, you don't need to mark others as inactive, current one as active. You can simply do this.

⁓ It just simplifies your entire thing.

Sneha Mehra (01:04:20)  
And, but it is very counterintuitive because you think is active is a property of photo, not a user. This is another thing that I want to always bring out that we typically think like whenever we are doing class design or like the low level design, ⁓ the basic class designer, when we're designing scheme of it, we tend to think in terms of attributes of a particular entity here. you think.

in a normal human way, you think that he is active as property of photos, which is where in most cases we typically read is active in photo stable. But it again, as I said, extreme nitpicking. But if you think a little out of the box, you realize that if you store active photo ID over here, it just oversimplifies it because now we don't need this. Now we don't need this. All you need is this with that with those two additional checks. Right. So now you are

Update or marking a photos active becomes just single query. No need to have two or three queries put in a transaction and all. If you have auto commit a single query is a transaction in itself that would solve it. Single query would do the job. This query becomes really simple because all you're doing is just doing a normal select over your select start from users where ID is equal to this. You get a user ID, you get photo ID, you can create S3 path, read the image, serve it to the user. And you're just oversimplifying.

A lot of stuff. A little counterintuitive, but it works. ⁓ I am purely highlighting these points purely from the perspective that what seems very simple, it's not as simple as it seems when you go into those details. When you go into those details, a lot of things can be challenged. There is no one way to build a system. ⁓ Right? We just moved his photo, his active.

as a property of photo to being a property of user, which is showing the active photo. ⁓ We are taking away, we just chopped off the complexities of the system by doing this. ⁓ instead of directly jumping to this, we went because it's a natural progression. ⁓ is 15th of bike. ⁓ In every single cohort since the time when I started taking this question, ⁓ every single cohort people start with is active over here because we are trained to do so.

Sneha Mehra (01:06:45)  
we are trained to think in a certain way, but that might not be the most optimal. And I'm not saying this is like, that is like my approach, like humongously more optimal. I'm just probing you folks into thinking that, ⁓ some other solution, ⁓ which might be a little more optimized can exist. ⁓ Right? So always be handy in challenging our own beliefs, ⁓ unlearn, relearn and become a better. Right? ⁓ Okay. So these are the key things that I wanted to cover.

Let me formally cover it before we take questions and it's already one or 10 minutes and we did cover on-demand image optimization yet. ⁓ Okay. Nothing to worry about. have enough. ⁓ We'll come back. We'll come back. We'll come back. Okay. ⁓ So for our gravitar.com, given that we dive very deep into the SQL query side of things, how we think about performance, how we think about architecting, ⁓ like think about other implementation details.

Let's talk about how will you design it in general. So we take the photo upload service that we designed yesterday, plug and play as it over here. No changes at all. Like your photo upload service, pre-signed URL and whatnot remains as is. ⁓ Right. But now when user is publishing the photo, you just make an entry into the photos table that we just designed. Right. That's how your photos upload would work. Then when your ⁓ user puts that even here usually this is a photos ID.

user ID is active as true schema users to store email and the hash storing derived value contrary to what we studied yesterday. Here we are storing the derived value because it makes index look up faster. ⁓ That's why we are doing it. Rendering the photo is really simple. Whatever you have now here just walking you through the actual thing because now we do not want to serve photos from our backend servers.

We want to serve your CDN. Imagine at the scale of Facebook, at the scale of Instagram, if they start using Gravita, right? They're doing millions and millions and millions of hits coming in. If that's the case, then how will you handle the traffic? All so much of traffic coming to our backend servers, which in turn gets pushed onto S3 that would create chaos. Your AWS will shoot up. So what do you do? You have to use CDN for that to give a better user experience. So now what we are doing is.

Sneha Mehra (01:09:09)  
we are writing API.gravita.com and then we put everything behind the CDN. So now observe this how CDN comes into the picture making things efficient. ⁓ we expose an API endpoint called API.gravita.com slash photos slash hash whatever the hash we provide over here we get the active photo we read the file from S3 and we return the response standard procedure. ⁓ So we derive the S3 path on the fly because we know the user ID and we know the photos ID.

We take that and we generate this part and we send it over here. Sorry, but basically whenever they do that, would fetch the photo from S3 and send it this way. The photo is under, but now what we have to do is we want CDN to set the photos. So what we do is we configure a CDN ⁓ and in CDN configuration, we create a property called gravita.com. So our main root domain itself is behind the CDN collection gravita.com ⁓ gravita.com ⁓

points to this as my origin, api.gravita.com slash photos. Because what we implemented just as the API endpoint which serves the photo reading from S3 and serving it back, looks something like this. api.gravita.com slash photos slash hash. So what we do is we configure gravita.com and origin to that is api.gravita.com slash photos.

So which means that now when the request for gravita.com slash hash comes in this request would go where CDN because in CDN we configured like how I configured edge.arpidbani.me here I'm giving gravita.com ⁓ as my CDN property directly my root domain itself is my property on CDN. So this way, when I make a request like this gravita.com slash hash, the request will go to CDN. If CDN has a data, it would serve it.

If it does not have the data, it would go to origin. What's the origin? API.gaveta.com slash photos. So what it would do? It would do API.gaveta.com slash photos slash hash, which is exactly what we have handled up until now. Because this is what we have handled. It would return the photo. It would go to OSD, read the photo, send it to the CDN. CDN will cache it and serve to the user. Now the subsequent request.

Sneha Mehra (01:11:29)  
would be all sort of the same. So now this is what our flow would look like. A very basic flow, ⁓ nothing super complicated about it. So we have a standard photo upload service. ⁓ You upload a photo using pre-signed URL and S3. ⁓ Once that is done, ⁓ you make a call to API server and add a photo over here, which registers it in your MySQL server. Pick your favorite database. ⁓ That's fine. ⁓ Now when the photo is published or when a photo is marked updated, ⁓ what we do,

We updated in the DB and push an event to Kafka. ⁓ Now why Kafka? Because we want to invalidate the cache. Because your CDN, where you configure gravita.com to api.gravita.com, let's say user B, user two opened a profile of user one and it rendered photos, something like this, this URL and rendered this photo. This request is served via CDN to all the users out there. But now when you as user one updated your photo, you want it to be updated right there and then.

So what do you do in order to do this? You have to invalidate the CDN cache. So because photos don't change that often, you might have a detail of, let's say one day or two, there's something like that. But now because you change it as to give a good user experience, you want to invalidate that photo right there. And then so that the new photo is rendered. So what you do when the photo is updated, you push an event to Kafka, the photo updated, it goes to a bunch of consumers whose job is to invalidate the CDN. That's it.

that event is consumed by these workers. They make API call to CDN. CDN gives you those APIs to invalidate a particular path. So you would invalidate a path slash H1 from here. So that file is deleted from CDN. Next request that comes in, CDN does not have the file. It would go where? It would go to the API server. API server will talk to DB, get the S3 path. API server will go to S3, read the file and send it to CDN. CDN will cache that file and serve it to the user.

Subsequent request will just serve from the CDN. ⁓ And this is how you can efficiently serve photos at scale. And this is how you can design your own gravita. ⁓ Really simple, but covering those details is the key. How we leverage CDN, how we leverage database transactions, how we made, how we evolved our database schema over time, answering those critical questions, how we thought about security and correctness of the data, no matter what. This

Sneha Mehra (01:13:54)  
are the details that I want all of you to remember no matter what. It's a very simple system, but these details make it very interesting. ⁓ Any question on this? ⁓ We'll take for five minutes question and then we'll cover on-demand image optimization. Then we take some more questions and then we go on a break. ⁓ Go on, Hemal. ⁓ Can you please go back to the screen where we have written the transaction three queries? ⁓

When we were writing ⁓ the transaction before we added this query, ⁓ there was ⁓ the second query which says their ID equal to something for which we have ⁓ got the photo. ⁓ I was thinking at that time, if let's say somebody passes a wrong photo ID. ⁓ How we have this user ID to prevent that right? So no rose updated would happen. That's okay.

That came that user ID came bit late when we start initially thinking. ⁓ then we added this. ⁓ Let's say the user ID is not there. I'm just talking about maybe a bit, bit general problem. ⁓ I'm writing in a transaction, some query and we are passing one parameter, which has to be used to find the identifier row. ⁓ Now due to any reason, it may be by mistake or something. ⁓ I passed the wrong value. ⁓ Now that part that inside the transaction, one query is not successfully handled.

Should we worry about writing something ⁓ as a fallback here inside the transaction ⁓ or should we leave it? What should be the right practice? Ideally, ⁓ if you were expecting something to be updated, but it was also let's say a general behavior is something gets updated, but you saw that nothing got updated. That is an alarm for you. In most cases, in most cases, which means that there might be a bug in the front end because here what you said is correct.

In that case, if let's say this does not match anything, what would happen for that user? All the rows are false. Like another photo is active, right? And this does not make anything active. Yes. That's a problem. was that's a, that's a photo for that's a photo for user. Correct. So if that is your internet behavior, that no matter what, no matter what this get out, or this should have like a user should have exactly one row as active.

Sneha Mehra (01:16:17)  
Exactly when was active. If that is what your, ⁓ if that is what your need is, right. Then you have to have that check in this transaction that at least one is active about it. Okay. If this does not change anything, then you don't do anything. So you either you rearrange the statement or you first do selects and ensure that everything is there. And then you do update either. Right. So what you can do is you can select it. This, this user have at least one photo is active. It would mad meets that then you update and update.

That would solve your problem. Right. So the photo and this user, so you first do a get of the same thing. Select is active true. Sorry. Select is active false. use where photo ID is this and user ID is this. If that row exists, then only you do this two things. Otherwise you don't. Okay. Because that's normally a requirement that you should have one active ID. Yeah. And if we are, if we are not taking care of this thing, then ⁓ I mean, it will be too late when

⁓ You've got to that client is at the first place itself. So it is always better to just validate it. Whatever changes that you are doing will not let your data go in inconsistent state. So you first check and then you do these two. Yeah, maybe there only we can save the row zero rows return then application logic take care of saying that ID was not when I do not find that the one. ⁓ But ⁓ I mean the logic has to be divided in two parts.

⁓ Checking the result is coming as zero row. So we had to buy app application layer and this transaction not going through. I mean, rollback in itself. So it'll be taken care of itself here. think that's all. ⁓ Yes. Yes. ⁓ Yeah. Could you explain the CDN mapping that we were doing there? Because like we did yesterday, right? ⁓ We saw yesterday how that CDN is mapped. ⁓ That same thing. What we are doing is we're configuring a CDN. Yeah. And the specific properties gravita.com.

Like how I specified edge.acquipani.me. ⁓ So any request for edge.acquipani.me goes to CDN. Correct? ⁓ Here any request to gravitar.com goes where? ⁓ To CDN. ⁓ No, no. Goes to CDN and CDN's origin is configured to api.gravitar.com. So if it doesn't have it, what it is doing is going to API. And because only API server knows. ⁓ Here one common question that I ask is why CDN not pointing to S3 directly? ⁓ Because S3 start public. Only

Sneha Mehra (01:18:43)  
My database knows which photo is active. ⁓ doesn't know which photo is active. Right? Only databases source of truth for it to know which photo is active. So which is why, what we have to do is CDN falls back to apr.grata.com, reads the database, find active photo, goes to SP, gets the image and sends it to the CD. Right. Makes sense. I was sort of confused. Like I was taking it in the reverse order, the mapping that was there. ⁓

Sneha Mehra (01:19:14)  
Yeah, I have it. So the first question is related to the hashing. So for hashing, you have said that, you know, ⁓ we cannot do we should not do the runtime hashing. ⁓ I was not able to understand why it was ⁓ there. We are there. We cannot do it because if your query looks like this select start from users where hash of email ID equal to right ⁓ now for your database to evaluate email ID matches to

It would have to literally go through every single row, compute the hash of the email ID, see if it matches with the thing that you passed. That is a full table scan. Got it. Got it. And in the, in those queries also, when you are joining, you are joining on the user's ID and then you are putting ⁓ a where clause of hash is equals to this. So when we are joining on the user ID, why can't we just search basis on the user ID itself? Why do we need to get it? You can do it because

The input that we are getting from the URL is the hash. That's the input that we are getting. Correct? But still we are not aware about the user ID then. ⁓ Okay, user ID is because the user is logged in. Is the join key, right? So that's what you are joining the two tables on. ⁓ So that is our problem. From there we are getting the user ID. We are not aware about the user ID here. We don't need to get the user ID, right? We don't need user ID. It's just a join criteria. ⁓ Because it's a foreign key.

This is just a joint criteria. ⁓ Okay. Okay. Got it. Got it. Got it. Got This is not filtering. We're not filtering on the basis of this is just a joint criteria. ⁓ Yeah. What is happening is you're filtering on this. So the first set of filter would be applied on this. Then the joint would happen. ⁓ It would filter it would just narrow down. So it will keep on narrowing down. ⁓ First, a huge narrow down would happen on the hash of the user from which you'll get just one row as candidate match from the user's table, ⁓ which would be joined on that one user ID.

to this photos table from the photos table. you write as a 10 photos uploaded for mass 10 rows out of which like this is where we getting all the photos. It would return all the 10 rows here because we are making one is active out of the 10\. would filter out where we just act. So active photo. This is how your flow would happen. ⁓ And your SQL query or in most cases your database query optimizer takes care of which one to evaluate first here. It would do because it would find there is a unique key on user hash. This is the first thing.

Sneha Mehra (01:21:37)  
that your SQL query evaluation. What is that? That evaluation tree that it creates, it would have the first thing to be done as evaluate this. Then it would define the joint criteria. Then we'll do that. This is basically called as predict. This is called as predicate push down. If you just Google that predicate push down, there's how SQL query SQL engines optimizes queries. So whatever you need at the bottom of the day, because a tree evaluates from the bottom to the top. So this

is what would be evaluated the first thing. It would be at the bottom node at the leaf node of your evaluation tree and then the joint and then it goes up. ⁓ it. ⁓ And second question would be, ⁓ so in the CDN, ⁓ if you can go to the CDN diagram ⁓ there, how you are invalidating the cache when the image is updating. So by making an API call to CDN. So CDN has APIs exposed Cloudflare, Akamai, every CDN has APIs exposed to invalidate the cache.

Got it. Got it. So whenever the new image is uploaded, you are invalidated in the cache directly by calling the CDN. Correct. ⁓ Synchronously. Yes. ⁓ It's like, it's like when user is, so if we are not depending on the TTL here, we are just invalidated in the cache beforehand. ⁓ Because user experience really matters over here that when the photo is changed, ⁓ I want to immediately see the uploaded photo everywhere. ⁓ Got it. Got it. Yeah. That's right. Thank ⁓ For richer user experience. ⁓

Chandran, know I'll take your question in some time. Let's cover a few parts of other by the way, this is what the next part was possible optimizations adding active photo ID over your own covered. Anyways, ⁓ what we'll do is we'll talk a bit about our demand image optimization. And then we take questions and then we go. So on demand image optimization as a problem statement is very interesting and very simple. Now that you know how CDN works, you know how Gravatar works.

All-demand image optimization becomes a piece of cake. ⁓ Now, let's assume that you have to optimize the images. Let's say, Cden doesn't give you that feature. And you have to optimize the images, which means that on typically your Facebook, you see profile photo rendered at this tiny teeny dot 16 cross 16, 32 cross 32 pixels. ⁓ But on other pages, it is like 256 by 24\. On some other pages, would be 512 by 512, something like that.

Sneha Mehra (01:24:02)  
So user can upload image in any resolution, but you want to render the small one or the large one depending on what is requested. So now you want to build your own ⁓ on-demand image optimization. This does not mean what we are not talking about is a, I'll upload the image, I'll post it as a message and have a bunch of servers optimizing it and creating multiple variants of it. No, we are talking about on-demand image optimization. Right? So now in case of on-demand image optimization,

Just like how CDN accepts query parameters, we would also accept query parameters like w equal to 32\. It means that transform the image to a 32 cross 32 pixel, not 32 cross 32, but 32 width image over here. For us to implement it in our Gravatar use case, just extending the discussion, it is now very simple. What we do is just do this. We were anyway reading the URL, we check if the file

exist on this or not. This is how your CDN flow would be by that. This is how your CDN flow would be. Where it reads the URL if it has the file it returns it. If not it reads the file from the origin. Right. In the origin when this request comes to origin. This request comes to origin. What origin would do? Origin would go to S3, read the image. In that it would then apply the transformation pass as the query parameter and send it. This is where your library is coming.

Your image manipulation or image processing libraries come. Image processing libraries like ImageMagic. This is the world's most popular image processing library out there. Python image library, PIL, has two bindings. One is native Python binding. Second is ImageMagic. ImageMagic is purely written in C++. Because it is written in C++ and almost every single language in the world has C++ bindings, which means that you can natively invoke C++ code in your ⁓ preferred language.

What you can do is you can actually just use image magic APIs to manipulate the image. For example, resizing that it has a native resize function. Image.resize and you pass the new size. It would resize and create and give you an in-memory bytes of it. Just send it to the user. So that's the only thing that you have to do. Like that's the only thing that changes in the flow. So now because you are writing your own demand image optimization.

Sneha Mehra (01:26:25)  
Whatever path you get in the query parameter pass the transformation, whatever you would want to support grayscale resizing and what not. Let's talk about resizing over here to simplify the explanation. The request would come to our origin server with those exact query parameters. From that, you see if you have the image, if not, you go to S3 retain image. And from that you then use image magic as a library to resize the image and apply the transformations. And then you send it as a response. This is exactly what CDN also does. When we talk, when we spoke about it yesterday.

CDN also does this, gives a transformation as ⁓ logic, sorry, a transformation as query parameters. We are implementing the same thing. So now what's the difference? How do you scale the system? This one thing always remember, image processing, encryption, video processing, these are all very heavy CPU based, ⁓ it's very CPU intensive ⁓ workloads, right? So what you need.

is you need very large amount of CPU to get things done. ⁓ you throw money at your cloud provider, get bulky machines with large number of CPUs and that is what you do. Whenever you're doing on demand, it's not asynchronous processing, it is on demand processing. Whenever you're doing on demand processing, you need to have bulky machines to do that because large number of CPUs multiple, you get parallel, lot of stuff and get things done very quickly. And this is

purely how every single on-demand image optimization or on-demand ⁓ XYZ processing is built. Nothing fancy over there. because ⁓ just to highlight one thing, it cannot be made async because this is a synchronous request. Wherein right now when you're given this URL, you're given this query parameter, you can say I'll do it processing asynchronous and then send it to you. It's not like that. ⁓ When this ⁓ URL is loaded,

You have to write there and then send the response with 240 pixels image. So you cannot do it as synchronous. So what you do is when the request comes to your origin server, you transform the image and send it. And here I'm assuming CDN does not give you this feature. So if you are building your own CDN, this is exactly how you do it. ⁓ Cloudflare, Cloudflare or Cloudinary. ⁓ One of them has a blog on it, right? They did exactly the same thing. It's literally the shortest blog, like one of the shortest blog I've read. ⁓

Sneha Mehra (01:28:47)  
We use even magic library this and run on massive servers, 30, 32 cores or 64 cores or something like that. Right. They just run on massive servers with this, this thing and that, and we just scale the fleet depending on the requests that are coming in. Right. And we just scale it to a larger one. That's it. I'd read it like six or seven years back. I just recalled it. Right. That's it. That's all it's all about. So handling scale in today's day and especially in cases synchronous thing, that's no, no brain. It's just that.

You need large servers to handle that particular part. is nothing magic when it comes to on-demand image optimization. ⁓ Our synchronous processing, ⁓ we'll talk seven, we'll talk in six weeks when we discuss video streaming a bit of it. You'll understand when you do a sync, ⁓ this is where we talk about image processing, synchronous, on-demand image optimization. ⁓ So what you can do, trust me on this, ⁓ this. It's not really difficult. ⁓ It's not really difficult. ⁓ So the sequence in which

We spoke on Pantella from 9 a.m. to 10 30 a.m. The sequence in which we covered things is exactly how take two days of your time and implement those exact same things the exact same way step by step. Just one local prototype, just one local prototype and you'll learn everything that you have to. You will never have any questions about CDNs, about image processing, about building the synchronous things. Piece of cake. Right? Just one. You are just one implementation.

Right. Okay. There's nothing much to cover to be honest in image optimization, on-demand image optimization because it's all about having bulky servers to do that and auto scaling them depending on the requests that are coming in. Right. Cloud providers give you those facilities anywhere out of the box. Very few chances that you and I would be working on a ⁓ team that manages its own EBRRA. Typically it is all abstracted out. So we typically don't worry about it in day and age. We focus more on things that matter. Right. Okay. Any questions in general?

⁓ anything that we covered from 9 a.m. to 10.30 a.m. We'll take questions for five minutes and then we go on a break. Yanukov, you had a question.

Sneha Mehra (01:30:56)  
Can you please go back to the page where we had those three update statements in our transaction and we did an optimization with active photo ID. ⁓ So my verification is that the new query after this active photo ID optimization to get the active photo, ⁓ should be sell it start from users where hash is equal to hash of email, right? Instead of ID. Yeah, yeah, yeah. So that's the exact. So let me do that.

update ⁓ users, set ⁓ active photo ID equal to something ⁓ where ⁓ hash is equal to something. Yes. Yeah. What I wanted to clarify. Thank you. ⁓

Yeah, I just wanted to check like whenever we have like requirements like this, should we have like Boolean variables like is active or should we generally store them as a type and enum as active because that gives us like more flexibility in terms of, know, changing if something comes up in the future, right? So long as you're doing it with tiny int it's fine. So eight byte integer if you're taking because Boolean anyway takes eight bytes to store the data. ⁓

8 bit is not even 8 bits. I wanted to say 8 bits. So you can go for enum and integer based enum that does it for you. That's also fine. ⁓ And that gives you flexibility on the fact that ⁓ you can then add one more state. So active and active you can add a third state if you want to. So that gives you that flexibility. ⁓ But just go with a tiny int rather than a big one. Don't store string as is in the returns.

Yeah, Harsh. I was just thinking, ⁓ should we improve this third update statements with some join on the other table that at least that ID is correct. ⁓ Yeah, those checks are there. Like I have to rewrite a lot of stuff there, right? But you can check that, that if that photo is there or not and those parts. So you can do a get and then set. So get an update, you can check in that part. ⁓

Sneha Mehra (01:33:23)  
Yeah, wanted to say that photo ID belong to this. Yeah, 100 % you should do that. Like if we take care of security over here and we lapse over here, that would not work. ⁓ But then it would have required me to write a lot of queries again and again. But you get the gist. I'm so happy that you folks are taking on those lines. But think on those granular details. ⁓ So we first do get and check and then update. ⁓

So it's like my question is related to this schema as well as a real text messaging we discussed in Slack, right? So we had user table and we had a channel table, but there we didn't make any foreign key primarly. We created a relationship, ⁓ a membership table, right? Membership table, yeah. Right. But here we didn't go with that concept. Is it because of one to one and one to many concept?

Brilliant. You gave the answer. ⁓ But here also it's one too many. But there we were storing other details as well on that. ⁓ In case of membership, ⁓ you could store last checkpoint up until this you write a message. ⁓ This channel is marked as star for a particular user. ⁓ Here you don't have that complex of a use case. ⁓ That's it. ⁓

I ⁓ will take a question at the end. It's 10.35. We'll take a quick break. We'll take a break for five minutes. We'll come back at ⁓ 10.41 ⁓ and then we'll have discussion around tag people in or tag people in photos and we'll wear a different hat. Now, will brainstorm on requirements. We'll not design system. Design system is easy. This system is very easy. What I want to highlight through this system ⁓ is how this simple system affects so many other systems.

Right? Because as senior engineers, ⁓ should be thrown into, we are thrown into situations where ⁓ we are not just responsible for building this one system today, but rather seeing how it affects, how it interacts with other system, how it could affect other systems and all. So we'll be wearing that hat to understand how this simple thing, and obviously we'll go to the implementation details of it a bit, but more importantly, understanding, like thinking around the requirements itself will be the part of this. Right?

Sneha Mehra (01:35:49)  
And then we call newly 100 % indicator post that. ⁓ See you folks in five minutes at 10.41, we resume our discussion on tagging people. Thanks folks.

Sneha Mehra (01:40:39)  
⁓ Let's start with the second half for today and we'll discuss tagging in photos. ⁓ So let's say what typically happens in a startup. ⁓ day someone from the senior management sub typically a CXO comes to you and says, Hey, let's will let's allow people to tag people in photos. That's all. That's your product briefing. You have to go ahead and implement. Right. So now a senior engineer.

You need to think of things that your senior management has not thought of. ⁓ Unfortunately, that's true. More common than you think. ⁓ So now it's your responsibility because they trust your instinct. They just give you problems. Now you have to figure out ⁓ what all things to do. It's not just high level. High level is still easy. ⁓ It's more about understanding how it fits into the scheme of things, how it fits into that entire puzzle.

of yours. Right? So now what we do is we think around requirements, but not talk about solution, not talk about solution at all. We'll talk about the requirements of it. Where the hat of a product manager as an engineer, you have to as senior engineer. So I always say this. I'll also repeat it next week. Next to next week. The senior engineer wears three hats. ⁓ It does not really wear one hat of engineering. Senior engineer always wears three hats. ⁓ First is a hat of a product manager.

to understand the product requirements, things that others have not thought about, bringing in that clarity. ⁓ Second is a hat of an architect, where you're supposed to architect a system considering extensibility of it in the future. And third, the hat of an engineer, which goes into those low-level nitty-gritty details of it and ensures that it is optimal across all the layers. And this is what makes a good senior engineer. You have to wear three hats. Today, we'll wear the first hat.

are out thinking around requirements where we are just given a briefing, a one-line briefing of the statement, hey, let's allow people to tag people in for us. ⁓ Okay. ⁓ What questions would you ask not to manage it, to yourself, to your team, or how would you brainstorm about it? Raise your hands, I'll pull you in one question, one person, and then we move to the next ⁓ one. Go on.

Sneha Mehra (01:43:06)  
Hemant. Yeah. Maybe the first question I will ⁓ try to ⁓ figure it out what should be the schema of storing ⁓ tags. Nothing around engineering. Nothing around engineering. Don't think on engineering at all. Because all you are given is this one line briefing. Just one line briefing. Allow people to tag other people in photos. That's it. Think on the product side here.

Sneha Mehra (01:43:37)  
Okay. Then maybe ⁓ the most important thing will be around ⁓ who can tag ⁓ and is there a way that I can mark tagging to be correct? ⁓ Who can tag is one thing. ⁓ So what is the significance of who can tag? What would it allow? Like what would it enable you to do? It enables that

⁓ I'm the owner of ⁓ let's say ⁓ photo then only I can take ⁓ I should so nobody else can kind of attack the wrong thing wrong people wrong object or something like this. No, ⁓ but someone else can tag you or not that is also important right.

⁓ I mean that that is okay but ⁓ we also need to think about if it goes through ⁓ a check otherwise let's say some eye comes and not part of not not a owner of a image and tries to ⁓ tag somebody in a wrong way and ⁓ normally there are repercussions around that. no no no, you are not the owner of the image still you are doing that is different problem.

But you are tagging someone else in your photo that's different. what you should more focus on ⁓ is authorization part where who can tag me in their photo. Obviously that's a check that you can only tag people in your photos. Right. Don't go into the correctness of it. Think about ⁓ our is other person allowing you to tag himself or herself in your photo or not. That is more important. Yes. That's, that's I mean, I know that's what you did, but you put it in a different way. ⁓

Got it. But here it's more important that it should be in your control that do you want to have this feature where you are dictating who can tag you in their photos? Yes. Right. Which means that now you have to have a separate authorization layer, which keeps track of things that, Hey, am I allowed to tag? Like who is allowed to tag me in their photos? So you see this feature day in and day out on Instagram, where you are not on even on Twitter, where you are not allowed to tag people.

Sneha Mehra (01:45:50)  
⁓ in their post, in their photos and what not. So that is an important point. Okay. Anubhav, what other part? ⁓ Yeah, we can also have lot of other considerations around ⁓ how many people can be tagged in a photo. Can they remove their tags themselves or? One, one, one, one, one, one, one, one, one. Okay. And then I'll, and then I'll circle back you in. That's right. But one, which one would you want to pick? I would prefer how many people can be tagged in a photo.

So you are setting, you're going for the max limit, like the max number of people, max limit of people tag ⁓ in ⁓ a ⁓ Also, whether it is a photo level or at a broader, ⁓ like maybe at an album level or at a collection level. Photo, photo. Let's keep it photo. But ⁓ how would it define your system? What, what would you have like, what changes can you make depending on this?

If let's say the max limit is 5 versus 10 versus 100 versus infinite.

Okay, 510 100 infinite. So we would need some sort of ⁓ mechanism to limit the number and given how like we can tag we would also need to manage concurrency a bit. So maybe if we have a limit of 100 and a limit of five, we want to

concurrently update tags if I'm updating or if someone else is updating, we also might want to say someone else would update in your photo. You are the one who are tagging in your photo. Someone else cannot come and tag you like tag in your photo. ⁓ Okay. ⁓ that makes sense. ⁓ So concurrency is not a challenge here.

Sneha Mehra (01:47:36)  
But how would this define ⁓ what aspect of the system would this determine the max limit? Let's say, let me simplify 5 versus 10,000. Let me go to extreme 5 versus 10,000. How would a system change if your ⁓ max limit is 5? How would a system change if your max limit is 10,000?

Sneha Mehra (01:48:03)  
I discussed this already.

Sneha Mehra (01:48:12)  
I'll tell you if it is just five people, you can go and keep them in an array directly. Right. But if it is 10,000, you would need to have separate entries for them because you cannot keep that entire thing chung together in that one document. Like what we discussed yesterday. Right. Yeah. This is what it would affect because if you're allowed 10,000 people, then your document size grows out to be really big. And then you want her to do pagination because no one will give you 10,000 in one shot unnecessary because no one's going to

going through 10,000 in one shot. So you would have paginated endpoints. So you would not store everything in one document. This is what determines your storage schema. ⁓ That this max limit of people, like max limit of people who can be tagged in a phone. Right? ⁓ This is what I want to probe. ⁓ Every single question that we ask leads to a change or affects or rather should positively influence our design. ⁓ That's right. ⁓ Okay. ⁓ Vikram.

One point from your side. Yeah. So, so should we allow to send notifications to the back? Great. If I say yes, then, ⁓ then, ⁓ miss, we should also consider the above point at max limit that, ⁓ notification should go to everyone who attacked, isn't it? Yeah. Otherwise, otherwise just imagine someone tagging you in a very shitty photo.

And you don't even know about it. ⁓ Right. That should not happen. So what does this mean? That you have to interface with notification team that, Hey, we are rolling out this feature. You would see a huge spike on your traffic as soon as people start to use this feature, because it's very likely for people to tag other people in their photos. So ⁓ get ready to scale up your systems. Yeah. Right. Because if they are, if they are not ready, right, if they're not ready,

with scaling of the systems and you push notifications on their end, the system gets overburdened, other critical notifications might get affected. So you have to interface with the notifications. As senior engineers, have to interface with the notification team. If it's a yes on sending notification to people, which in most cases would be, you have to interface with the team and say, and just ensure that they're prepared for us to handle, to see a step, to see a rise in the notifications that are being sent.

Sneha Mehra (01:50:39)  
This would also lead to the fact that there is a throttling. Now this throttling is different. This is basically product level throttling. So typically people don't love products or the people don't like when apps send them ton of notifications. People hate it. People, good apps only send three notifications a week or five notifications. If someone, if someone app is sending you notification every single day, it's a big turn off.

people are very likely to unlock or to basically uninstall your app. So if this is a new type of notification that is sent in, let's say you are being tagged in many posts, getting notifications for every single post you are tagged in is a bad user experience. So now you also have to talk to the interfacing team, the notifications team, that, hey, how are you handling this? Are you having any kind of throttling? If not, let's have that.

You need to ensure that you are not bombarding other users with notifications for every photo that they are tagged. Club them. Do something like make product better on that site. Give them a rich notification experience. ⁓ Yeah. ⁓ It could also happen that let's say a very famous celebrity posted one photo and other people are tagging the him that Justin Bieber. ⁓ You would get lots of notifications. ⁓ sounds good. Correct. ⁓ Correct. So

all these things because it is very much possible that a senior management has not thought this through. They're just they're just saying one line statement to you. It's your responsibility to dissect it and go into those details and point them. Right. Okay. What else, Vishal? One point from your side. Yeah, maybe user level setting whether anyone can take in. did that first point. First point with who can take. ⁓ okay.

⁓ No, meant like a weather. ⁓ am a user, so I should have like a prop in the like profile settings, whether anyone can take me or not. if I have this, that's covered on the first part. Any other part? ⁓ No. Self removal of time. Like ⁓ anyone time me and I don't want to. ⁓

Sneha Mehra (01:53:03)  
Let's say in the obscene photo people tagged you. You should have an option to remove yourself from that. So now this makes authorization really interesting because now authorization is not just about who can tag you, but even more granular level that you all the photo is of user one, but you are allowed to remove yourself from that photo.

a little more complex than we anticipated. now this requires you to build a very robust authorization layer. Very important feature. Right. Thanks for bringing that up. Anubhav, which other feature?

⁓ Yeah, I was thinking from user experience perspective, whether the tagging would be self driven as an I tag someone knowing them or whether we would have suggestions. Okay, this ⁓ photo looks like this person from your friends list, ⁓ something that Facebook does, I think. So with for this, we will have some recognition and recommendation ⁓ machine learning systems that provide us these inputs, we would need that. So now what would this entail? Like how would this affect

your design and peripheral teams. ⁓ With my design, have like my systems will have to interface with taking inputs from these ⁓ peripheral teams. ⁓ Whenever a photo upload event is triggered, I have to get these inputs from the ⁓ both for the face recognition and suggestion inputs from these systems from the periphery teams ⁓ and display them to the user. which means that, okay. So then front-end changes for you to display that

And more importantly, ⁓ when you're displaying them, the response time of that service needs to be slow. ⁓ It should be fast enough. ⁓ If they are not fast enough, your user, then sometimes user sees it. Sometimes user doesn't gives user a very poor experience. Yes. So you need to ensure that whenever you're rolling this out, that ML team, ML algorithm, whatever they are running, it's very fast. It maintains a certain SLA. So what this brings to the point.

Sneha Mehra (01:55:10)  
is notion of SLA because you have a synchronous depend. Now you have a synchronous dependency on the team because ⁓ as soon as you clicked on the photo started tagging, you have to immediately show them the suggestions. So given that the ML teams API, which is suggesting you things on the image, they need to have a very strict SLA and they need to adhere to that because you have a synchronous dependency on that. ⁓ And now you also need to create a fallback plan. What if that API does not respond in time? What would you do? Will you not allow user to tag?

Or will you allow user to manually tag? It's still a product call. It sounds straightforward, but it's still a product call because it might be possible in some cases that you don't even allow people to tag unless it's coming from your MLT. For some reason, you want to do that. ⁓ But it's important to think on those slides. ⁓ One more point before I cover one other aspect of it. Aalok sir, on. What's that one feature that you would to consider? ⁓

One question should we update the user's activity page or where the user can see what happened?

profiles as activity. Now, how would this affect your system?

Sneha Mehra (01:56:23)  
⁓ If I say yes to this, what does this end in? ⁓ We would be maintaining this information somewhere that what happened, what happens to user profile. So we have to actually start pushing this information there also asynchronously. Correct. That's what and.

Sneha Mehra (01:56:42)  
And also who can see this activity, then that is probably the users itself will be will be seeing that activity. That's fine. But now I'll just add to that. What does this affect? This affects one very interesting part that this the way you are storing this information that in this photo, these people are tagged. ⁓ They need to be indexed. Somehow you need to store this information in a database where you can now you should be able to vary that. Now you cannot just have

just things stored in an array which is unindexable. You have to store it in a database like the tagging. In this photo, these people are tagged in a way that you should be able to query that who, all, or rather in which photos is this particular user tagged. Which means the database that we pick has to support that query. You just cannot pick any random database here.

You need to be very particular about a database that supports this kind of queries very efficiently. Very seemingly simple problem statement. But this is what happens typically in startups and mid-scale companies where senior management has not thought it through. But when we start brainstorming on it, when we start diving deep, we realize that it affects so many things. It can interface with so many other systems. Some honorable mentions ⁓ are feed.

For example, do you want to send or to like, let's say someone tagged Arpit in a photo. Do you want Arpit's follower to get that event in their feed? If yes, ⁓ then product call feed team needs to gear up for that scale because every time someone tags you push it into a feed. Then similarly moderation needs to happen on feed that hey, you don't show this kind of post every time. ⁓ Only to a few users you show and whatnot.

and those kind of things would come up. Now ⁓ on a bit of mathematical details of it, what I want to bring out is this point that in most cases when people are tagging in photo, there are two ways to tag it. First is you just say, Hey, I just stored X and Y. But if you look carefully, you are not just storing X and Y of the photo of the coordinate where you are tagging it, but rather what

Sneha Mehra (01:59:09)  
If you do X and Y, let's say your image is 720 by 720 pixel. ⁓ And let's say user tagged at this particular location. ⁓ Let's say this is 320, 320, 120\. ⁓ Now when you do this, ⁓ you cannot store 320, 120 in the database. Why? ⁓ Because image resizing. ⁓ What if on a low end device, ⁓ the image is 360 because 360\.

And then it would tag at this location x comma y 320 comma 120 would come somewhere over here, which is incorrect. Right? So that is where whenever you are tagging a bit of granular detail is that you have to tag the relative positioning of it rather than the absolute one. And it's a relative positioning. It means that you store instead of storing 320 comma 120, you store 320 divided by 720 and 120 divided by 720\. This way

It's relative, which means no matter how this image is resized, so long as the aspect ratio is maintained, your tag position will be perfect. No matter how small the device is, no matter how large the devices. So relative, so storing relative positioning is a very important implementation decision over here. Okay. Apart from that, what you store is instead of just storing this one point because

In photo you are not just literally your mouse is not hovered on exactly this point for it to show who is stacked over here. It's typically a bounding box. Which means that instead of just storing x,y relative position of it, you store the x,y of this point and this point or x,y of this and width and height. So these are called bounding box that you would store. So you would store either left top right bottom or x,y width, height.

So LTRB is left top and right bottom coordinates of these two points, the relative coordinates. One way to store it. Second way to store it is to store the coordinate of this and width, relative width and relative height over here. Either way, whatever your front end is more comfortable. But these are really important when you are storing this information. Because this way you would be able to properly place your mapping and when mouse is forward, you show it.

Sneha Mehra (02:01:26)  
Even when your screen is like shrinking or growing and you resize and what not. Right. Okay. And just using things that you're interfacing with, search, page detection, location, notification, role-based access control, R back, role-based access control, the authorization that we spoke about and anything that you're interfacing with, you will put it in Kafka no matter what. So search, notification, analytics, profile, feed, activity and what not. Right. So some things that senior management typically has not thought about.

We have to think on that behalf and figure those details out because devil lies in the details. Always remember this. Right? ⁓ Okay. So this is just a quick brainstorming that I wanted to do around this system because it tells you that a very simple statement allowing people to target photos leads to so many interesting things out there and you can extend it in many ways and

It is not that engineer cannot propose product ideas, engineers can. It's just that we have to be a little more thoughtful about it because we don't have product experience, but we can still pitch in ideas and make them and ask them the critical questions which they typically would not have thought of. Very possible. So whenever you're working with product team, whenever you're working on something that is user interfacing, think as an end user, put them, put yourself into their shoes and think and propose. That's how you create a cross team influence.

and a positive impact in the organization leading you to have a good career path. Now the final system for today is we design newly unread indicator. Listen to the problem statement clearly. ⁓ So the problem statement is that you see this day in and day out ⁓ on LinkedIn, Twitter, Facebook, even Instagram, every social platform out there. So it is possible.

that you are offline, you receive a message. ⁓ You receive a message and when you log in the next time, you see the counter of it. I think I can show you a demo. Let me quickly check if I received any messages on LinkedIn or Twitter. If yes, I will show it to you.

Sneha Mehra (02:03:40)  
Uh-oh, no message on ding ding.

message on Twitter, Instagram. Do I have a message? yeah, I have messages. Right. Okay. I'll show you. I'll show you. There were, it would clear out the requirements because it's something that we see every day, but we take it typically, we typically take it for granted. Okay. So the statement is that this message is three. I don't know when I last checked it.

But these three messages that you see, these are three messages. These are not three unread messages. These are three ⁓ newly unread messages because it is possible that now when I click on this, the count becomes zero. It does not mean that I have read the messages. You can still see unread messages over here. Right. It doesn't mean that I have read the messages. It just means that I

I'm acknowledging that I have messages. This is not behaving the way I wanted to, but you get the idea. This count should reduce to zero. ⁓ Right? So what we are doing is if I would have received any message over here, I can just click on it. I'm flying taxi. Okay. ⁓ I get like, ⁓ have one message, but when I click on it, that count disappears, ⁓ but you could still see that there are a lot of messages, ⁓ which are unread. Unread messages are there.

But this does not mean that here I would say 99 or 10,000 or 50,000 count. What I'm actually looking at over here is actually I have so many unread messages. ⁓ But count there was zero because these are not acknowledged by me. It's just that the presence of it is gone. ⁓ So this is newly unread. So you can choose ⁓ not to read the message, but it indicates that there are new unread messages.

Sneha Mehra (02:05:33)  
This is what we have built. ⁓ So let me clear out. ⁓ Now the promise statement is bit more clearer. ⁓ Let me walk you through the requirements and then we bridge off. There is a very interesting high level pattern that I want to introduce through the system. ⁓

Sneha Mehra (02:05:56)  
⁓ Okay, the statement is very simple that we need to inform the user about presence of new messages. Terminology like the words that are used very carefully presence of new messages not under it not unacknowledged. So on the message icon when you see this three this implies that you have three newly unread messages. It does not mean you have three total unread messages. So you click on this and this country goes to zero.

It does not matter if you read the messages or not. This is what we are building. Now I am just twisting it. Twisting the problem statement and I just add a pseudo constraint where this 3 does not indicate that you have 3 newly unread messages but it indicates that you have messages from 3 unique people. ⁓ So if I say 3 over here, means I might have got 300 messages. But from

three unique people. So I'll show three only. ⁓ So this is what a requirement is. And when someone clicks on the messages, this icon, irrespective of the user read all those messages or not, the count would disappear. That's what we have to do. Requirements near real time update message when received newly, new Android is equal to demurrer of unique people from which you received the messages while you were offline. How would you approach this problem? Let's brainstorm.

Sneha Mehra (02:07:25)  
Yeah, so first and foremost, when we say newly, right? I think I would like to say newly means less than 24 hours. ⁓ is also last time when I clicked on the message. can since then. ⁓ Okay. So then what we need to do is, ⁓ okay. First, first and foremost, the messages that we are getting, right? We need to have an indication whether they are red or not. That is the first, first indication.

just to know how many messages I have read and how many messages have not been read. And then there should be a way in which I can also know when was my last read time, when I last clicked ⁓ that message icon. So these are the two indicators. And the third indicator is once I have that, there should be a kind of a filtering mechanism based on the last read time so that I can get the number of unique users that have messaged me. And using that,

I can populate that count on that, on that icon. These are the three aspects that I would focus on. ⁓ So the first thing would be, ⁓ first I want to start with a data model, ⁓ like the message data model that is received from ⁓ the message itself and the timestamp because timestamp is very critical here.

Sneha Mehra (02:08:48)  
received. ⁓ You have your message. Yeah. ⁓ You have from which user. Yeah. And you can also put two users that could. ⁓ That's that. That's ⁓ that. Two users also let's put. So from and to. Yeah.

and a time step. ⁓ And then there should be another data model, which would be like last read time. So that is very important for me because I need the count.

last wave time.

Sneha Mehra (02:09:24)  
⁓ Now? Yeah. now whenever, ⁓ I'm, so I, you want me to talk about how the messages would be read and all that stuff or that is already? Yeah, everything, So what, so obviously because this is a messaging thing, right? So there will be a messaging queue and this user will be subscribed to that queue and based on the number of, ⁓ and the partitioning would be on the from user that would be. Okay. Let's start. Let's, let's go straight back. What is the input to the system?

Are you storing all the messages that your user received in this table?

Sneha Mehra (02:10:00)  
Am I? In this messages table, are you storing all the messages that a user ever received?

Sneha Mehra (02:10:10)  
I can store, there's no harm in storing, but I think that would not ⁓ be a scalable solution. So then what are you storing in this? What is this messages table?

Sneha Mehra (02:10:23)  
Okay. So I was just thinking one snippet of a message, like, you know, one message one like, hi, how are you? If I send it to you. So that message unit, I was thinking in that way. How will you store that? How will you figure out what that unit is? What, what, exactly are you storing? ⁓ That's a user be sent a message to user a. ⁓ Okay. So then this table needs to store either all the messages or part of messages. What exactly are you storing over here? That's the question.

So I would be storing pretty much all the, so this data model, like for, will store all the messages that a user that has sent to me. So for example, it could be like, have sent high messages. ⁓ Yeah. think there's no arm in that way. Yeah. This is a message. Yeah. Yeah. It is a messaging table. Yes. That's correct. this is all your messages that is shared over here.

So the text part from this to this and timestamp. Okay. You have the messaging table. So now the scale of this is pretty huge, but we will touch upon that later. That's fine. Okay. Now in your other thing, what is this data model called? That other one, where you're storing the last red hat. This is basically for me to get the ⁓ latest messages. So this can be... ⁓

latest message, ⁓ like some kind of a metadata information. So probably you need to have something substantial over here because you might want to store something more. So let's say right now I'm calling it messages or other, let's just store it in a user activity. ⁓ It's basically one on one mapping with user. So every user will have this ⁓ one entry over here where you are storing that last read ⁓ at other details. ⁓

⁓ Let's say this is your user activity table. ⁓ Now what you are doing over here? What's your query to ⁓ get the unique user scout part?

Sneha Mehra (02:12:32)  
Okay. So now, ⁓ the unique user part would be like, ⁓ you know, select messages, ⁓ from messages where timestamp ⁓ is ⁓ greater than last read at. ⁓ that is okay. And then I need to, ⁓ order them by the, ⁓ I need to order them. You are not, you don't want to send a message. You just want to render the count. Why order?

Yeah, but you want the unique users counter. You don't want the message count. So if I do this way, then I would be getting all the number of messages, right? You're doing select star from messages. Yeah.

Sneha Mehra (02:13:17)  
where two is equal to your user ⁓ and timestamp is greater than last readout. ⁓ As in join wine is that what is the problem in this query? See, the problem is, for example, I have sent. this messaging table will store three rows. If I send three messages to you, hi, Arpit, how are you? So I've sent three messages. So three rows it will store. Now in the previous requirement, you said you want from unique users.

How many, if I go with this approach, I would be getting three rows as my input, right? ⁓ But I want three rows as input. just want one, ⁓ one, ⁓ so I'm mixing, I'm, ⁓ accumulating all those three message rows into one. So that's the reason I said, ⁓ I, are you doing that? ⁓ count count from direct to this now.

count of unique of from, correct? count unique from messages where 2 is equal to your ID and time set is greater than this, right? Yeah. Okay. ⁓ What's problem in this?

Sneha Mehra (02:14:36)  
⁓ I think the indexing would be the problem ⁓ because I think we need to index it properly. Otherwise it will go through. It's like one of the time consuming queries because you're pretty much getting all the messages. So what I would do is ⁓ basically we can quit. We can index ⁓ on ⁓ from messages, right? So maybe we can index it ⁓ using a key that has from and message.

would message, no, not message actually, but maybe from the from, ⁓ from the from field, I can index it ⁓ or ⁓ no, not even that actually from timestamp timestamp would be a good indexing factor because we want all the messages after time this time. Right. So yeah, I think we can index it on the timestamp. That's what I'm thinking. Just on timestamp. So it will be like across all users, all data is sorted by timestamp.

Sneha Mehra (02:15:36)  
times. Yeah, that is what I'm thinking. ⁓ But this is still an expensive query because if let's say you have received 1000 or other 10,000 messages for a particular user, you're going through 10,000 rows of your table and then doing a unique of it. Now imagine doing it every time the page loads. Very expensive. ⁓ Yeah. Yes. Very expensive. I, I'm not doubting the correctness of the system system would work. ⁓ But at millions of people,

there in the millions of users currently active on this platform at this very moment doing this file in this query is super expensive and if I just to clear out you will create an index on from and time stamp or composite index on from and time stamp so that all the entries of a particular from a particular sorry from to and time stamp a composite index on this so for entry for and it will be ordered by

index on to, from and timestamp in this particular order. So all the entries of two will be clubbed together, then grouped by from and then ordered by timestamp. So this way, ⁓ you are filtering it out, you are directly going to that location where two of your interest starts within which two timestamp and from. ⁓

I think it should be two timestamp and from, right? Yes. Yeah. Because if you that, can very easily ⁓ cut short that. So it will be two timestamp and Yeah. So two timestamp and from. Right. But although ⁓ you are trying to, you know, answer every single performance from this by rearranging this bit, but the worst case is imagine you having to iterate through these many index entries to just

get the count of it. ⁓ And in case, ⁓ let's say a user, ⁓ example, worst case, ⁓ 10,000 messages are there. ⁓ You have to go through 10,000 entries of it and then do a unique and it turns out the count is just one. ⁓ Let's say, hypothetically, one user sent you 10,000 messages, you're going to 10,000 index entries, we will just figure out that the unique count is just one or two or three or something like that. ⁓ That is expensive. ⁓ Yeah, actually, I'm just thinking instead of going messages, we can have another column.

Sneha Mehra (02:18:05)  
not another column, we can have another table that tells how many messages have been sent, just a count. ⁓ And that because I really don't care what messages are being sent, right? All I care is our messages being sent. ⁓ So I don't care about the actual messages. So now we are coming to ⁓ a more efficient implementation of it. Right. Okay. So this is a crude way and this is not an incorrect way. Again, I'm saying this is not an incorrect way.

This is a crude way of building things which gives you which is not very efficient but very correct when you fire this because this messages I can get this goes for everyone. This is your source of truth for messages. This is your source of this is your source of truth to say when you last clicked the message icon doing this will exactly tell you how many people from which you got those messages exactly it will be 100 % correct no matter what. But the problem is

it is just a little, a little, not a little but rather inefficient.

Sneha Mehra (02:19:12)  
That's the problem. ⁓ Right. It's just inefficient because do we really need to iterate through all the messages of a user that it received and then figure out, okay, these are the unique ones. ⁓ Can we do better? ⁓ Should I continue with something? ⁓ Sandeep, ⁓ just keep your heads raised. ⁓ I'll just do it so that I don't forget. I'll just do it because you added some brilliant points that I just want to pick others' brains.

and see where we can go with this. Right. Yeah. But just basically keep your hands raised. Okay. Rahul, how will you optimize this? Actually, like, ⁓ I will not go with the system. Instead of this, I will try to go with non-relational database because like, ⁓ that's similar query. You'll fire that, right? Yeah. So basically the system would, that's the system that I'm thinking is like, in the non-relational non-relational database.

There is a user key. Basically there is a user key and it has sub hash maps in it. Like it has more hash maps in it. Okay. ⁓ So ⁓ user key ⁓ and ⁓ in this we have total, ⁓ total count basically. ⁓ Just, just a count.

Sneha Mehra (02:20:33)  
and ⁓

⁓ No, not not this not this. ⁓ Just just just count with a number. Count with a number. What do you mean count with a number? I mean count ⁓ count variable. Then is this even a table? ⁓

No, no, no, no. Count. ⁓ It's the name of the, it's the name of the key. ⁓ That's what I was saying. ⁓ It's the name of the key. ⁓ And yeah. ⁓ So these are actually, ⁓ instead of saying count B, ⁓ it's actually total or other duly and read count. ⁓ Yeah. ⁓ Yeah. Something like that. It's just a nomenclature and ⁓

⁓ In this HashMap also, we are also having ⁓ one other HashMap that is unique users.

So it's kind of like a set actually. is like a set. ⁓ Now ⁓ whenever ⁓ there is a new message, ⁓ first what will happen is like ⁓ the API will try to fetch this object. Just one question. What is the difference between this and this?

Sneha Mehra (02:22:03)  
Okay. So basically like a newly unread count, like how will you know a uni user is messaging you? Yeah. basically like, basically like there are two ways we can do this. Basically we can have a, just a unique user site. Okay. ⁓ And it will have like a list of unique, unique users and we can, ⁓ we can just get

get the length of this unique user set ⁓ and show it on the ⁓ notification box. ⁓ I was actually thinking like it was kind of like an O-N of complexity. I was thinking like this. no, ⁓ no, no. Set length is not O-N. Everyone keeps a track. Every set gives a track of the length of the set. It's O-N. Don't worry. Okay. Okay. Sure. ⁓

And so when whenever any user ⁓ user ⁓ like messages this user ID, it will just it will just put the user in that set. Okay, correct. Now, ⁓ whenever user click that button, like, like in the Instagram, ⁓ whenever you click that button, we can just delete this thing at that point of time. And that deletion will be over and also

So I think this system is much more efficient. Correct. Hmm. Yeah, that's about it. That's what we will need in this. ⁓ So what you went with, I'll just talk about it a bit. Delete said when user clicks ⁓ the message cycle. Okay. So what you did is here, this is a classic case. And again, you would see this more than everyone. I'm busy talking to everyone. ⁓

This is what you would see almost in every single system when you are implementing. You either can query the source of truth or you can keep things pre-computed. A classic way to build any system. Either you directly fire source of truth when you need strong consistency ⁓ or you may build or to get more efficiency you may pre-compute the data or store it in a way which is consumable directly because here it will look carefully. Here you can directly

Sneha Mehra (02:24:28)  
get the length of this set and see the number of unique users that sent you the message. Here you had to fire an expensive query, SQL, NoSQL implementation, anything that does not matter. ⁓ But with this, you get this very nice optimization if I may call it. Because now much more, but obviously now it's our responsibility to keep this updated. Here it was by nature, it was updated because it is actually the source of truth.

all the messages ingested into this table, we pick it out, we index and then fire the query and get the output. But here we have to have another flow in which we are also putting this data in this other approach that we are building. And again, as I said, neither this is wrong, neither this is wrong. It depends ⁓ on what you are prioritizing, what your scale is. As a startup, this would work just fine for you folks.

But as your company grows, when millions and millions of people are concurrently active on your system, you might be tempted to go with this approach. Right? OK. Let's elaborate on this a bit more. ⁓ Hemant, which database would you pick to store this? ⁓ I think it should be a NoSQL system, because what I'm trying to put are variable length data against the user ID.

So one user ID against one user ID, I need to explore, let's say, 5, 10, 100 or any number of ⁓ user IDs. having a document key value store where value is a set, ⁓ maybe MongoDB kind of a database, where key is a user ID and value is a set ⁓ where ⁓ each element is a user ID. ⁓ That's it. That should be sufficient. ⁓

Let's use Redis for this. And I'm just oversimplifying it for Redis because almost everyone is aware of Redis and its functionality. ⁓ But you can pick any that gives you set functionality. That's what the key highlight is. ⁓ Any key value store that gives us set functionality in the value so that you don't have to check and set and whatnot. ⁓ You just reuse what database already gives you. Right? ⁓ Okay. So we use Redis to store this data. ⁓ Okay. ⁓ Now we, I'll just draw a few elements over here.

Sneha Mehra (02:26:53)  
Okay. Now when your user tries to get the message, so how would first let's talk about rendering the number that we are showing to the user. then it might be a get ⁓ by might be a get status ⁓ API, which what it would do is it would talk to this red is ⁓ and

count the elements and return the response. So get is done. And similarly, when we are deleting from Redis, we can have a clear API that goes to Redis and clears things up.

So we have clear API, it goes to Redis and it deletes the key that we just said. So get done, clear done. ⁓ And if you look carefully, ⁓ read path is done, a very trivial write path is done. ⁓ Now comes the interesting part, that how will you update this Redis with this data that you need?

This one, this unique set because someone has to put the data in that these are the unique user from which you received the message while you were offline. So what, how would that work? Pankaj?

Yeah. So ⁓ two things here at like what I was thinking is the client also needs to be smart because if the user is already on the messaging tab, ⁓ even though he's reading some other users message ⁓ and he is getting the message at the same time, it still will not be counted as ⁓ you know, newly unread because he's already on the tab, right? So we don't need to show him that notification at that time. the client is offline. So ⁓

Sneha Mehra (02:28:46)  
Yeah. So the second thing, what I was talking about is like if we have a web socket connection for messaging, so that backend can check that if the socket connection with the client is off, then I can update the Redis, but the client also needs to be smarter that when it is on, ⁓ it should not send it when the user is ⁓ already on that tab. So if I'm browsing through Instagram or watching some reels, then if somebody sends me a message, then I can trigger that update API.

But if I'm already on the tab, then the backend should take care that, okay. If the socket connection is off, then yeah, I should update the Redis. Okay. So the input to your system brilliantly put input to this system, which updates things in Redis looks something like this. Okay. Now that the cat is out of the pack, I can directly cover this. Right. And we'll talk about a high level pattern that I wanted to, right. Because it's not just one database that we would have, we would have

few more databases to support it. We'll just talk about it. Okay. So ⁓ the input to the system will be our messaging service, our core messaging service, because our messaging service knows when a message is not sent to the user. Right. And here I'm simplifying it by saying that when user is offline, then only do it in real world. see very other use case as ⁓ a very well explained where you are on some other tab. still receive a message. So your count increases.

Those cases are there, but I'm just simplifying it. You can add other cases where you would want to pop up that notification. But here the idea is very simple that if the user is offline, but the message is unsent. Now, whatever the definition of unsent is user not on the tab user offline or user doing something else. And whatever the definition is of unsent that comes from a messaging service. Why messaging service? Because messaging service would know the message is delivered to you or not.

If you're on that tab or not, if your WebSocket connection is on or not. In either case, this is the source of trouble. Your messaging service. ⁓ Now, messaging service can take help of online offline indicator service to see if you are online or not. Or it might own on its own have this information because it has the the connections, persistent connections are with that edge server. ⁓ So that edge server is the one that would know if you're online or offline or connected or not or on which tab. So in either case,

Sneha Mehra (02:31:10)  
that becomes our source of problem. That hey, if my message is set up, whenever messages is unsent, ⁓ I'm creating an umbrella of everything that I want to pop notification on into this message called onMessageUnsent. But depending on ifs and else's where you are, what you are doing, you just implement those logics over here from the front end. Okay, now this all of this comes to this Kafka onMessageUnsent which becomes the input to our system.

And in that what we do is we do this in each event. send data like from this user, this user, this message was sent because we want to store the number of unique users. We store it in a set. ⁓ And what we do is we typically store it in redis. You can pick any key value database, which is partitionable by default. All of them are and provide set as value any database. Right. Okay. So get status straightforward, clear status.

⁓ So when a user clicks on the message button you delete the entry from the set. When you want to get the number of unread messages count, you make a call to this, count the length of the set and you return. ⁓ For you to update things into Redis, your right path, go to something like this. OnMessageUnSend, coming from your messaging service, leveraging online, offline and what not. Your consumers consume it and write to this Redis cluster. The name, new user, new user.

You received a new message. You received a new message. So all of that goes into this Redis cluster. ⁓ A lot of people would stop designing at this stage. ⁓ You'd say only this much is single. Right? You have a read path, you have a write path, you are almost converging onto the design. But no, ⁓ this is not good. Why? ⁓ Because imagine the situation. ⁓ This database, and now this is the high level pattern that I want to talk about in the next 10 minutes.

This is one thing. This Redis cluster that you have ⁓ is serving a huge number of reads. Imagine millions of users on your platform constantly refreshing the page or going from one page to other, you making the same pay call to get the unread messages out. ⁓ So this database is anyway handling large number of reads. Plus whenever user clicks on the message icon, you see large number of writes happening over here as well. Plus anytime there is message unsent, you are updating into this database.

Sneha Mehra (02:33:39)  
This database is overburdened. When this database is overburdened, as I always say database is the most little component of an infrastructure. See if you are doing some unnecessary operations on this database. Are you doing any unnecessary operations? Let's take a look. Get status, no option. You have to go to the database because that is a source of growth. You have to go to the database, read it. Read replica, you don't need it. It's an in-memory DB. ⁓ So get status, bare minimum.

Clear status. have to go and delete the key over here. You can't do much over there. You are not doing any undo operation, any unnecessary operation in this path that you are making the call. Okay. Let's talk about this part. Are you doing any unnecessary operation? You might think that hey, every time I'm seeing an every time I receive a message unsent event, I'm just updating this database. Hmm. I'm doing the bare minimum it seems, but no given that

In this Redis cluster, the value is a set. If I receive one message from user B or 10,000 message from user B. My message, the state of my database is not changing for 9999 times. Right. Because for the first message itself, the entry is being added into the set for everything else. There is no change happening in the database. So can we reduce that? ⁓

If you look carefully, look, go through your messages, see all of your, all of your friends. When you are offline, they don't send you just one message. They send you five, 10, 15, 20, 30, whether a ⁓ guy, what are you doing? Come here. We are going for the dream. What are they send you a lot of short messages, right? Especially when you're open habitat, right? Given that first on the first thing you are writing it into the DB.

DB state changes for all other N minus one times. There is no changes happening in the DB, but you would still make a call to DB trying to write to the set. Set does not exist. sorry. Set already exists ⁓ and ⁓ there is no changes happening in the DB anyway. So you are wasting a lot of network IO happening over here and not as network IO, but while making by letting this database handle that request. Now, how would this database handle the request?

Sneha Mehra (02:36:04)  
It would get the request. It would read the set, see if the entry exists in the set. It would discard if it does not exist, it would add. There's too much of work going into this DP doing nothing for those n minus one times. So this cluster ⁓ is in most cases doing nothing, but unnecessarily handling operations leading to no state changes in the database. This is what I want you to focus on. ⁓

Whenever you are looking at a database, again, one more high level pattern. Here is what I'm talking about. Whenever you find a database that is doing a lot of unnecessary operations, leading to no state changes in database, try to minimize them because database is the most brittle component of your infrastructure. No matter which cloud provider you pick, if a database is done, everything is done. Protect it. It's like Voldemort and it's basically Deathly Hallows.

Keep it protected no matter what, keep it like well protected by all the magical charms, everything that you can put in a really good mood for some reason. so protected, protected however you can. And given that we are doing a large number of unnecessary operations over here, how do we minimize it? By using an auxiliary database. This is what I want to talk about. Having an auxiliary database whose sole job, so here what you are doing is,

This is your primary database. You want to protect it. Network IO cost is not much. In this case, preventing your main database from going down by increasing, by adding a bit of network cost is a good trade off to make. Always remember this, right? It's all about trade off. So you are incurring one more network IO, but if you can reduce the load on your database, it's worth it. It's worth it. Every single time it's worth it. ⁓

So what we do is we add an auxiliary redis into the scheme of things. The job of this redis is very simple. Whenever the write happens, the write first happens into this redis. ⁓ If an it write happens and it checks if the entry exists or not. If the entry exists, good enough. I would not make a run to this database. If entry does not exist over here, then I'll write to this database. This way for n minus one times

Sneha Mehra (02:38:31)  
You are not even calling this main DB, reducing the load on this database. But all of this load is handled by this. This is the one that is bearing everything. Checking if I've already, so which means it's basically checking if I've already have an unread message from this particular user or not. Right? That's what it is all done. This way what happens ⁓ is this takes a bulk of the load. ⁓ N minus one queries are firing over here. ⁓

while one query goes over here. This way, this database is protected while this takes all the hit eggs. Right? Now, what's the worst case? The worst case is this database goes down. See, you are protecting this database. If this goes down, then everything is done in any case. Worst case is what if this goes down? If this goes down, what would happen? No right would happen over here. All the rights would go over here. So it would have to handle all n-1 rights. That anyway it was anyway doing. Right?

So by adding this auxiliary database, ⁓ you are just trying to shed the load that was there on this main database. ⁓ This is the high level pattern. ⁓ want to talk about. This is really important that whenever you have a very brittle component in your infrastructure, ⁓ this typically your database. And if it is doing a lot of unnecessary operations leading to no state changes in the database, ⁓ try to add an auxiliary database and reduce the load on the main. This way, ⁓ your service will be up and running. Everything will be happy. Right? So now

Just future this by adding this thing, your status check API. you clear the status, it not only has to delete from here, but also delete from here. ⁓ That's it. Some here and there, it glitches on consistency would come, but not much to worry about because it's not transactional. It's not something that deals with money and all. ⁓ It just unread messages. No one even gives anything about. ⁓ Right. So don't worry much in case it goes into slightly inconsistent state. It's not an end of the world for your product. So you can do that, but

If that entire system is done, that is the end of your product because it's very bad experience. ⁓ So you need to pick your battles. ⁓ What you are trading off for what. ⁓ And this is what I wanted to cover as part of this design. This high level pattern and how we are approaching this situation. ⁓ And this is what I wanted to cover for today. ⁓ Folks who may want to drop off, ⁓ I'll take questions around anything as part of the system and then across all the systems that we covered today.

Sneha Mehra (02:40:57)  
I love data recording by three slash four today. I will share the previous for the next week. ⁓ Next week is all about it's my personal favorite week storage engines where we'll build our own storage. It will clear a ton of concepts about databases, not just how to use it, but internal software. It would build that intuition around writing your own database. And I'm not just talking about just for the sake of engineering, but more importantly, knowing those internals would build you that intuition on how

systems like Kafka work with because I didn't know internals of Kafka. I guessed it and it was exactly how it was implemented. I want you all to form that same kind of intuition, that exact same intuition so that whenever you come across a new system, it's very easy to guess ⁓ and you can guess anything. But in this case, your intuition is built, ⁓ 99 % of the guess would be correct, which is what we want to do. So next week is all about storage engines.

where we built three very interesting storage engines. Two of them were 100 % lawyer. Why don't we? Because we're going to those granular details, which I think everybody. Right. Okay. So that's what we'll cover next week. So folks who want to drop off, drop off. I'll take questions. Yeah. ⁓ So I would like to add one more optimization to this system. ⁓ So let's say on the front end. So let's say there are thousands of messages. ⁓

let's say 10,000 messages, then instead of showing 10,000 messages, we should display more than, let's say 100 plus messages. ⁓ And at a point that user is not, ⁓ don't care that he had 10,000 messages. ⁓ So, so let's say on the front end, if we have a hundreds of a hundred messages, then, ⁓ then we, we only show that a hundred plus and in next steps, subs of some messages. We don't write to the database itself, ⁓

We know that it does not matter. It does not matter. ⁓ Yeah. True. True. True. True. That's a good optimization. So that's where you can also leverage this. That it's not that every time you have a unique entry, you do this. If this length is more than let's say a hundred, that's the threshold. You don't even write a positive to this. We even say that one extra right. That is happening over there. ⁓ Thanks for adding that in. ⁓ Okay. ⁓ Anything else? Any question? Anyone?

Sneha Mehra (02:43:18)  
Was it really that clear? ⁓ Yeah. So in the case of auxiliary red is what, how you're making the entries because see in the case of red is cluster, are storing a set like B, ⁓ we can, you know, we have CD EFG, ⁓ but in the office, ⁓ are storing it as separate entries, like B maps to see ⁓ anything. picked this so that we have a very clear, different key model setup.

⁓ I would pick this like for example, B received message from C B received message from D as a key and value is just true. I am just reducing the overhead of storing an extra set comparing if the key exists or not. You can do that way also. You can do this way also either way is fine. And when you're going to so ⁓ let's say at the nth movement till n minus one, you are storing the values here in auxiliary radius. Not storing, I'm just checking, just checking, not storing, just checking.

⁓ Let's say user B sent A, let's say user B sent A 10,000 messages. I'll be writing one time over here, one time over here and other 99 checks would happen over here. So one time right over here, one time right over here, but everything else I'm just checking it over here, correct? Because if it exists already, so let's say I take this example. ⁓ If B sent a message to C, if this entry already exists, why should I write it over here? I should not, correct?

Right, right. here I'll just check. ⁓ I don't have to create 9999 entries over here, correct? ⁓ Got it. Yeah, there was a gap in my understanding. was thinking you are just incrementing the count here also. No, no, no, there is no count. There is no count over here. This is just existence. ⁓ Just existence. Got it. Got it. Got it. ⁓ Thank you. ⁓ I had a question related to clear status because somewhere when we always keep B to C true, ⁓ right?

⁓ I read the message by clicking and then again B to C happens. ⁓

Sneha Mehra (02:45:27)  
So when you clicked on the message icon, when you clicked on the message icon, this set got deleted. These entries got deleted. Okay. When you, when another message was sent, then the entry would be created again. What is the HKS at all? Help me understand that. No, as long as both are getting synced, there is no ⁓ issue. I was. Yeah. As you very rightly said in case, let's say you wrote over here, but this right way.

or other this right succeeded in this right field and there is a problem. Not even denying that, right? Not even denying ⁓ that. But it's not the end of the world. And it's very likely, very likely that the words that would happen in the count you would see too, but in reality you would have 300 messages. Right? Yeah. And that's okay. It's still okay. Right? Because it's not ⁓ dealing with a financial system or really critical things.

some inconsistency where ⁓ one right succeeds other right fails. That is a problem. But introducing that complexity of doing distributed transaction is not worth ⁓ it. You just accept bit of inconsistency. ⁓ Thanks. So should we build it build this functionality where we're doing the status check API as a transactional base like updating both the databases like the auxiliary and the Redis cluster at the same time? ⁓

Not worth it. You take up some inconsistency responses in favor of high throughput. Not worth it because it's not very significant. Got See what's the worst that would happen. You see two messages count of two, but in reality you have three new messages. You're anyway clicking on it, going on that screen, you would see unread messages. If that person has another thing and you did not respond and it was something that would send another message to you. That's okay. Right. So given how humans work.

It's okay to, in this case, it's okay to have a bit of inconsistency, but why do I add that unnecessary complications of distributed transaction? You really don't need that. Got it. Got it. Yeah. ⁓ And this is again, ⁓ that one thing that as you think you feel that you are cheating on by not having this human transaction, making a data code and inconsistency, but in the real world, it really does not matter. You think from user experiences perspective, ⁓ what's the worst that would happen? Would it be an end of the world for them?

Sneha Mehra (02:48:01)  
Yeah, but I'm just weighing with why like instead of building auxiliary radius, why don't we do a read before write in the right part? ⁓ Here? ⁓ before write over here? ⁓ No, in the right part. ⁓ Even read is okay. You're saying we don't even want to want to read on the database. On this database, ⁓ right? ⁓ No, no, no. What I'm saying is I don't even want to read here.

Okay, you don't even want to get. Because see every instruction you fire on this DB consumes CPU consumes memory. Even a little bit it consumes that. I'm just protecting this. This is like that, that ⁓ beauty in the beast car rose that you are placed in this glass jar no matter what you don't want it to crumble. Got it. it. Yeah, makes sense. And one detail I might have missed in the Gravtar thing ⁓ was there like

Why did not we provide a user ID in the graph that is that is privacy or that's by design how graph that is like I. Okay, ⁓ you are saying over here why we don't have user ID. Yeah, like or whenever you generate something, generate even in the hash or like just scramble it with user ID that will save a lot. No, yeah. okay. So you're saying that, okay, somehow I inject user ID over here, but then the problem is

that this hash is universal function MD5 hash. Right? Now Facebook, let's say Facebook wants to integrate your Gravatar. It does not need to know your Gravatar user ID. Got it. Yeah. That's what I was trying to do. All it just has, if it has your email ID, it can just try to see if you have Gravatar or not. can start directly rendering it. Yeah. It's okay. That's a simple requirement that people will feel easy. Yeah. Yeah. ⁓

⁓ I had a question about on message and send I think I did not standard. ⁓ If a user ⁓ is active and then they receive messages from any of the charts, we don't want to trigger that event. Only if the user is an active, whether we have an active web software connection or online offline service, ⁓ only then we want to trigger that ⁓ on message answer. ⁓

Sneha Mehra (02:50:24)  
⁓ I want to understand what unsent means I'm not able to. It's an umbrella that I created when I was not able to send a message to this user in real time. That's it. Right now, implementation of it depends on if user is on other tab, then also it is considered to be unsent. Right? Or if user is completely offline, then only it is considered to be unsent. That is our product. Right? But the idea is, whatever flows into this,

on message unsent. It's like on message unsent in real time, which means that this is when you have to update account. ⁓ Thank you. Any other question? Anyone?

Sneha Mehra (02:51:10)  
⁓ Nope.

Sneha Mehra (02:51:14)  
Tick tick one, two, three, four, ⁓ five. Okay, thanks folks ⁓ for tuning in. ⁓ I'll update, I'll upload a recording in some time by three and upload the notes as well. See you folks next week. It's all about storage engines. Quite fun, we'll dive deep into database details and implement a couple of other systems. Thanks folks for tuning in. It was great session, great questions. I have a small query here. Sorry, sorry to interrupt. ⁓

⁓ So, ⁓ see, initially we talked about the, know, ⁓ query optimizations and database optimizations for my skill or let's say any, any, ⁓ any relation database. Can you suggest something how to improvise on this? Because I personally feel that, you know, I'm really not good at this. So what are the responses? So what I would do, what I would recommend is, ⁓ I actually, to be honest,

How I became good at it is I focused very well during my college days on that subject very deeply. ⁓ So that is one, that is my solid foundation, but I cannot obviously recommend you to read that entire textbook. So what you should do ⁓ is understand how database works, how relational database works, how data is structured, those things you need to go really deep into it. ⁓ And how you do that, ⁓ I am right now running a series ⁓ on my YouTube members only thing.

around database and in case you are interested, go through those videos and see if you are interested into that. ⁓ I cannot claim that I will cover all of that, but it is definitely going to be a big value. ⁓ Because I'm distilling the information from the textbooks, from the courses that I did, going into those nitty gritties of databases, not really internal into it, into the paid structures and all, but more importantly, what you use in your day to day life, what you should definitely know in your day to day life about it. That is what I'm covering. I'm literally

Going through my college textbook, I'm not even exaggerating. I'm literally going through my college textbooks, ⁓ three of them, to just ensure that I've covered all the key details. Very soon I'll be putting out videos on ⁓ Phantom Reads is obviously coming next week. After that it will be SQL query optimization. After that, ⁓ indexing evaluation and whatnot. It's all going to be around databases in general, right? Because that's my strong quota. That's what I realized. So that's what, but in case you still want to explore, ⁓

Sneha Mehra (02:53:37)  
Log into your production console, ⁓ whatever your production console is, ⁓ log into your production console, your MySQL thing, ⁓ and ⁓ start firing explain statements. E-X-P-L-A-I-N. In case you are unaware, Google explain statements MySQL. It gives you the query plan of it. ⁓ Every single word of it, Google that word of it. ⁓ What it does, how the query plan is created. ⁓

and refer to actual MySQL document. And I'm just using MySQL as an example, pick up every database, go through the official documentation of it and not some blogs or videos about it. Go through the official documentation. ⁓ You would see how dense those documentations are because that would give you those nitty gritties that typically videos overlook, that typically blogs over. ⁓ Do that, ⁓ But pick one query at a time, run the explain statement, get those terms.

look out in the official documentation and re-interact. ⁓ And over time, three or four months will be well-versed in that. Now, when you're doing this, if you spot any optimization in your company's workflow, suggest. On your staging environment, make your stages and see how it behaves. See if it improves something or not. Or in any case, you don't have to change in production or change in staging. You can just dump that database into other table.

fire that exact same query with your changes and see what kind of improvement you got. Companies, ⁓ in price, a very good, is a, is a very good playground for us. ⁓ any case, if we improve, it's better for organization and you learn stuff as well. Right. have done it ⁓ more than I would want to add. Right. So that's how I became good at it. Yeah. should create a video. It's a good topic. Yeah. So thanks. Yeah. ⁓

I think like a reduced cluster is there. Do we need auxiliary reduced cluster also? ⁓ It will be cluster. It will be cluster. But single node will not be able to handle. You need a cluster for that. It just became too dense over here. That's why I just. Yeah. Great. Thanks for doing it. See you folks next week. Bye.

10

Sneha Mehra (00:00:01)  
We will create week 5, day 2 and will continue our discussion on storage index. ⁓ The agenda for today are two very interesting problem statements that would not only be interesting but will also solidify a lot of core understanding that you need when you trying to build an interesting system or you are trying to get the maximum out of existing system. So there are a lot of building blocks, ⁓ lot of foundational concepts.

I would help you ⁓ think in a very different way. ⁓ And both of them are actively used in production. It's not something that just theoretical. ⁓ You will find practical use of both of the systems to a huge extent. Let's start with the first one. The problem statement itself, and we'll directly start with the brainstorm. The problem statement in itself is very quote unquote tricky. What do I do? What will do?

is we will build a word dictionary without using any traditional database like no mysql, no postgres, no mongoDB, no redis, nothing you have to be creative you cannot use any traditional database to build this word dictionary ⁓ so you have to be creative into thinking how will you do this this thing needs to be scalable which means you should not be worrying about scale but then

Obviously some constraints would be there, but it needs to be quote unquote scalable. ⁓ And on both the sides, so typically when you design any system, you can bifurcate into two, one is storage, other is compute, compute basically your API servers and storage means your database. Here you cannot use a database, so that's ⁓ an interesting problem to solve, but it needs to be ⁓ scalable with respect to both storage ⁓ and API servers. It's ⁓ okay to have

high response time because that's what that's not what we are trying to solve over here. Getting performance is different. We have done it in past at lot of systems. What we are more focusing on is this constraint that you cannot use a traditional database. So we are okay to have a little high response time because we know hundreds of ways to improve it like caching matching and all what we can improve. But let's not focus on that. Let's focus on the core part of it that it's okay for us to have a little high response time. But what we want

Sneha Mehra (00:02:29)  
is we want portability. ⁓ Portability, come when we speak of our brainstorming, would explicitly point out that hey, this is not making our system portable enough. And you'll understand the importance of portability in the next system ⁓ and next week. ⁓ So I just compiled this problem statement in order to cover a lot of critical building blocks that you see in day in and day out in real world. So we want portability. Our response time can be high. ⁓

So you cannot use a traditional database. You have to be creative. Now here, one key point in the dictionary, it's a standard English dictionary. You have English words, you have English meanings. ⁓ When you do this, but when you're serving this dictionary, it's ⁓ mostly read only. You are not going to update. Imagine how frequently Oxford dictionary gets updated. Not much. Right. So what we are doing is we don't want to support continuous update. What we are doing is

the words and the meanings will be updated weekly. ⁓ It's not that it will be updated anytime you want. ⁓ Weekly, ⁓ you will be given a changelog file. In this changelog file, you will have the updated words and the updated meanings. ⁓ You will have this information. ⁓ You have to take this changelog file and apply it to whatever storage engine or whatever database you are building. ⁓ So you don't have to support pointed updates right there in the link, live pointed updates.

These will be given to you weekly basis, let's say every Sunday. ⁓ Every Sunday you are given this new changelog file over email, you have to apply those changes over here. ⁓ The lookup in this dictionary will always be a singular word. You don't have to think about autocomplete and all. Give an award, give me the meaning. That's it. If the word does not exist, return 404\. The entire size of the dictionary ⁓ is ⁓ 1 terabyte. ⁓ And the dictionary in total

has 170,000 words. ⁓ So total dictionary size is 1 terabyte, 170,000 words you have in the dictionary. So you have to build a system like this, ⁓ build a word dictionary without using any database. ⁓ So no MySQL, no MongoDB, no Elasticsearch, no Redis, no Memcache, nothing, nothing, Be creative in your approach. Now think about it.

Sneha Mehra (00:04:54)  
How would you go about and things that we would brainstorm on is around storage query, whatever it is, similar subjects. Right. So let's start with the storage one. Go on. So basically there are two ways that I can think of. We can store this. One is a text file ⁓ and other is a JSON format basically. And one terabyte big dictionary, JSON format. Yeah. That's a bad way to go. But like

Just as a reference. do you, how do you query a JSON format? So basically it would be like marshalling and marshalling and that would be a very expensive process. ⁓ That's why I'm like more inclined towards text file. ⁓ So basically is it okay if I like tell about updates also? Wait, wait, wait, wait, wait, wait, wait. Let's cover this one part in 10\. ⁓ Okay.

You mentioned text file, you mentioned JSON file. So when you said JSON file, you would want to store all the words and all the meanings in this one JSON file. Given that you brought up this topic, it's ⁓ amazing. First of all, thank you for bringing this topic up. But now what would happen? A lot of us put a lot of data in JSON files. Now the disadvantage of storing data in JSON files, especially in this use case, ⁓ is that for you to check if a word

exist in this JSON file or not, you anyway have to read the entire JSON file, deserialize it, create this in-memory object, and then check like in-memory hash map or something and then check. That becomes really costly. Size of the dictionary is one terabyte big. There is a direct way to access a particular thing in this JSON without reading the entire JSON file because that's what the format says. But which is why we cannot go with JSON file, a singular JSON file.

containing all the data, which means that what we can pick is we can pick a simple text file. Raul, what will be the format of the text file? So basically it would be like, ⁓ it will be something like write a head log, will have like, which will create some logs like this word is updated. This word is like this word is updated with this meaning. It's a dictionary. How are you storing? Yeah. So that's, that's, that's how, like basically like this word equals to ⁓

Sneha Mehra (00:07:16)  
Meaning it goes to anything kind of CSV. Yeah, you can say that. Right. So let's say Apple, comma, ⁓ fruit and then you like, and Amazon and it can have repetitive columns like ⁓ it can have like a columns with the same name. There are no column. You only have words that meanings like, like, like Rose Rose with the same same words.

⁓ No, no, no, we don't want same words in the dictionary. Imagine this file opened in any text editor. People will find it difficult. Right? You portability. You cannot have duplicate things over here.

like it is related to the updates that I thought like, ⁓ that's not now. Now I'm adding, now you have to think out of the box, right? Yeah. I know where you're going with this right now. Then you'll create an index and what not on top of it. No, I was actually thinking about dirty reads and something like that in that context, but let's give it simple because this one constraint that I'm adding that you cannot have repetitive things in this file will make things very, very, let me add it.

⁓ no. ⁓ Repetitive entries. Okay. ⁓ Now things will become interesting. Okay. ⁓ So now what will you do? Is this format according to you? Okay. Enough to start with. Yeah. Or started at. Or started at. It is okay. Okay. Just one question before I moved to Heman. ⁓ Will you keep this file sorted in some, some fashion?

Okay, if there is a sorting constraint then there will be No, there is no constraint. Will you keep it sorted?

Sneha Mehra (00:09:07)  
No, I don't think so since we don't have any like, ⁓ like, I, don't have any restrictions on the response time right as of now, like we are not bound to just give the answer quickly. So I will not prefer sorting here. Okay. So words can come and fit in any order here. Okay. Okay. Okay. ⁓ Hey, man, as, as I go with the trees or something like that. Wait, did it jump?

Wait, wait, wait, wait, wait, we'll come to that. We'll come to that. Right. ⁓ Because that's because it's fun to discover it step by step. Right. Because that solidifies the understanding for him. That's right. ⁓ Okay. ⁓ So, ⁓ Hemant, ⁓ anything that you would want to change in this file format? Yeah, I, maybe I would like to start from scratch the way I have to think it. ⁓ so if I say the biggest

requirement it is first ⁓ searchable ⁓ dictionary because the first most important and the heavy thing is it's a it should be a searchable via key let's say i consider this thing as a key value store where key is a word value can be a two line three line four line explanation of that word ⁓ searchable via key plus it's a read heavy operation

The dictionary is always maybe have a thousand to one ratio of read and write or maybe something. ⁓ is Q2? Much more than that to be honest. ⁓ Much more than thousand to one. Yes. ⁓ So read-heavy searchable system is one thing. ⁓ The only third thing which I have to take care when I'm going to add or modify a word, ⁓ I should not kind of boil the ocean for changing a small thing. ⁓ That's a second thought I'm having. What did you say about the last thing?

When you're modifying something. ⁓ So, so I'm saying it should not kind of a disturb my whole system by doing a very small change. I should be able to update or add. ⁓ that's third thing I have to keep in my mind when I'm designing it. ⁓ If I design in a way where each ⁓ add or each update is too costly, then also not a very good design. It may be a good, not very good. That's why my, but just to, just to probe you on that.

Sneha Mehra (00:11:31)  
Are updates are going to be weekly once? Yes. ⁓ When I say ⁓ it should not disturb means it should not need to do too much heavy on IO side or read side on desktop. If you put it in such a structure, then I have to go and scan each and every file and then go and figure it out where it is and then do a change or add. ⁓ Then it ⁓ does not seem like a good design. ⁓

But then in any case, our updates are, I totally agree with your point that each update, each single word update cannot be very costly. Yes. But given that my total, I'm getting updates in batch. ⁓ can still do ⁓ one ⁓ huge processing. Yes. One where I'm merging. Processing is okay. Okay. So now let's, let's, I think we have understood the fair bit of biggest requirement of the system. Let's think about making it, ⁓ how we can do it. ⁓

When we say we need something searchable, ⁓ definitely a sorted via key looks like the first thing that we have to thought. ⁓ If you do not start it, ⁓ there is no way you can do a search in a very better way. I mean, in an efficient way, if I have to use the right word. So starting by key should be the way. ⁓ Now, how you will store it so that if ⁓

I have to ⁓ search it. can efficiently do it means in place of having a single text file, maybe I need to start it or partitioned it where there are multiple of files. ⁓ if you use yesterday's pattern of figuring out the routing, whatever way you you sent a word ABC and you figure out out of hundred files that I am having ⁓ this file is responsible for ABC.

Let's say, so I have not to go and search it and that file is sorted. So ⁓ if I go and know that this is the way we are up to search it, I can immediately go and read that specific place in the file. that disk I and seek and all those things in minimal. ⁓ So ⁓ just to, to, ⁓ just to stop you there. So you are proposing we don't have just one file. We can split it into multiple files. Right. Just one thing. ⁓ I added this constraint of portability.

Sneha Mehra (00:13:58)  
We cannot have multiple files. Okay. ⁓ Because when you have multiple files, what you are killing is you are killing portability. What we want is literally this one file in some format. ⁓ do a copy like it should be one file which we should be easily able to ship. We should not be able to copy tens and hundreds of files and then only a specific software can read it. Right. What we would want to have is we'd want to just have this one file containing everything we need to read and write to.

Okay. I have a ⁓ bit different of Arpit there. ⁓ When we say multiple files, that's a logical ⁓ way of saying it. We may think about like having a one single zip which have hundreds of folders and each folder has something which makes. ⁓ No, but then you would still have to extract it to use it, right? ⁓ So, so I'm, I'm, I'm also against it. I'm saying have just one.

⁓ And you'll understand why I'm adding this constraint when we reach the end of the solution. Because it forms the foundation for hundreds of systems that I'll talk about. That's why I'm just probing you to think in it might seem like Arpita is adding like random constraint, but these are all, this will all come in very handy. Right? I got it to your point. Maybe I will think with this. So when I say multiple files, my thought came from the logical division.

not necessarily we need to have a multiple files in one single file we have a logical division which says like okay first 100 pages or first 100 is for this part and then ⁓ up to this to this area ⁓ so we can have a proper header or something which is a second file which talks about up to this sector and this place are for this area and we can do it by a single file idea about making it searchable ⁓ via sorting so that's why partitioning is necessary

Starting is necessary when you store it. ⁓ Now, now what we need to think about is when, when we are ⁓ going to do a ⁓ representation in the system, we need a key and maybe ⁓ a big size of ⁓ variable size of ⁓ value. ⁓ In file, if you take ⁓ line by line approach or ⁓ any approach ⁓ when we are writing it,

Sneha Mehra (00:16:19)  
maybe that will not be so good. So we need to think about it encoding it in a way. ⁓ So in place of directly Jason, if you use any other format of encoding when we are storing it, that may be helpful, but it will definitely ⁓ add the ⁓ cost of decoding it. So if I, in place of Jason, if I use any famous barricade or any other format that that will add a bit of a ⁓ work there.

But means that that should be not too much in my understanding. may need to think about the ⁓ real facts and numbers to figure it out. it is a bad ⁓ con, that's not doing it. Thanks. Thanks. Thanks so much for adding so much. I just that I'm not ⁓ I'm trying not to digress into parquet, Avro and all this format. They're trying to build a base understanding of this while still keeping it simple. Right. Because ⁓

What you suggested that we have multiple partitions in a file and we should be able to look up within that. That is still a complex enough solution that will touch upon later. ⁓ I'm trying to go in a very simpler direction so that the base concepts that we need to build anything queryable works. And just to add to a real world use case on why ⁓ I'm just trying to keep it simple ⁓ is on ⁓ AWS

AWS has S3. We all know AWS has S3. ⁓ On S3, we can upload any CSV file. And we can query it using a service called Athena. ⁓ know. ⁓ In Athena, ⁓ when you're things on S3, you're just storing simple CSV file, but you're still able to seamlessly query on Athena. That's what I'm trying to simulate. ⁓ With this example, with this basic understanding,

We are trying to build AWS S3, not at that scale, that, that basically that particular compute engine, how they would have thought of this, how they would have implemented this. And it has a ton of other applications as well. But this is what I'm trying to simulate in a very simple sense. Can we make, and obviously the points that you said, will all incorporate that. Right. But not with that specific, like let's say Parquet and this, because not a lot of people would know about it, but simplifying it to

Sneha Mehra (00:18:45)  
have an understanding, okay, this is what we need. You can use Spark and whatnot to build it, but instead of going into that complexity, we'll just keep it really simple on that. ⁓ Just maybe a last line to say on that. ⁓ terms of the thought process, how Athena has came or ⁓ other systems work in a such way, we need to think about something like Lucille's way of tokenizing and putting here. The only thing is in the key,

⁓ Lucene tries to figure out the document ID and name and all those things against it to save it here. We have to save it, maybe a continuous, ⁓ data of two, three lines, which is the meaning of that word. ⁓ So, ⁓ so we, need to think about, ⁓ building something like this and ⁓ that, that's, that, that looks near real because if you see the Lucene syntax, they, they are creating a single file. So, I mean, one ⁓ index is good enough in a

one single file, which is quite large. Even can go to one terabyte, whatever you are saying. ⁓ That's all I can think of. No, but you added some great part. just that I'm trying to keep it ⁓ one level down and then, and then jump the level. Right. That's it. Okay. So this given this file that we have, right. It will contain all the words that we have seen for now. ⁓ How would you search on this?

Yeah. So I was going to come to that, ⁓ given that we have a single file and let us suppose this is Terabit big file. ⁓ What I would do is basically like, imagine I will start with binary search and then we will go ahead. ⁓ what I would do is basically take some offsets in this store it as indexes and given a word. ⁓ So basically, ⁓ now, okay. I'll get there where I'm getting is basically I'm getting into a B plus tree. ⁓ want into keys and values.

and basically create blocks of pages in memories. ⁓ So let us construct a first level index first. ⁓ Okay. Now let us suppose you have a one terabyte file and you have eight kilobyte block. ⁓ don't go into blocks at all. Keep it very simple. ⁓ simple for a five year old to understand. ⁓ Okay. ⁓ Let us suppose you can keep 1000 keys in memory. ⁓ Now wait, wait, wait, wait, wait, wait, wait, ⁓

Sneha Mehra (00:21:07)  
⁓ I can keep a lot. No, why I'm saying this ⁓ is we are over exaggerating. We are over engineering this. This is a very simple word dictionary. Right? Keep it very simple. No need to add B plus trees. No need to add some, some, anything fancy. Keep it very simple because that's where you see the power of the system. And I'm just trying to bring it down to a very simplistic level because you would see the beauty of the system. Right? That's right. That's Given that you have, what would be the first intuition?

The first way of making or let me cover that. I know you folks would go into very fancy direction. I'll just pull you in, give me a minute. ⁓ Okay, so the nine way, let's say we discussed that, hey, we cannot use a traditional database to store this particular dictionary. Dictionary contains key and the meaning. Now for us to ⁓ make this, now given that we cannot use a traditional DB, the simplest thing that we can come up with is, let's just store things in a simple text file.

Given that we are storing in things in a simple text file, what format would you store? The normal format that we can think of is let's say we store it in a CSV format. But first thing is key, then you have this gigantic meaning for that particular key. ⁓ Okay, given we have this, we need to now search for this, like because that's what the use case is. Your use case is to do a key lookup. Now, given that we have this one gigantic one terabyte file, and now we want to do key lookup in this.

What is a naive way if you have just this one file and you want to do key lookup, what will you do? You'll open this file, read this file line by line and you find the particular key that you're looking for. Once you find it, you stop the iteration, read the meaning and return the response. Right now, doing this for every single read is very expensive because dictionary is a very read heavy application. So linear search is not a way to go.

Because when you search linearly for every read request that is coming in, ⁓ would be, worst case, would be effectively reading one terabyte file every single time. ⁓ Worst case, ⁓ on an average, you would be reading half a terabyte for every read, which will make your things, first of all, very slow. ⁓ Second, ⁓ it would degrade the performance of your disk. Third, ⁓ your disk sectors will go bad because it is very heavily overutilized at the moment. ⁓

Sneha Mehra (00:23:30)  
It has hardware problems, it software problems, it has time problems. ⁓ So whenever you are in the situation where your reads are slow, ⁓ always, always in order to make things faster, in order to make ⁓ reads faster, the only thing that should come to your mind is indexing.

⁓ So can we somehow index the data? Can we somehow create a small index through which we would be able to not, through which we we avoid doing a sequential scan of this entire file to hunt for the meaning that we are doing or to hunt for the key that we are looking for. ⁓ So can we create an index on top of it? So Siddhash, what would be the index that will create? Just one level index, simple.

What would this look like? ⁓ Yeah. So basically ⁓ offset ⁓ and the name like basically Apple and the ⁓ key and the offset and the offset will point to what? ⁓ offset is basically offset inside the offset is basically inside the file. So first line of this file is basically, let us suppose we have organized it line by line ⁓ is offset zero and so on. So is offset. So offset is the

nth byte from the beginning of the file, ⁓ up until that line starts. ⁓ So apple is done, Amazon is offset to which points over here. Then maybe you might have a third word, let's say book with offset tree which stores over here. And offset is the byte position of that word in that file. ⁓ Now let's estimate the size of this index. So this is a classic index. ⁓ Now how would the read happen? ⁓

First of all, would now what we'll do now if you were to given that we have an index, given that we have this gigantic file and now if let's say I'm searching for a word, if I'm let's say searching for Apple, what I can do is I can read this file. I can open this file index file that I have. I'll iterate this file line by line at the moment. We have still not done what, how should we read this file? But assume that we are reading this file line by line and we

Sneha Mehra (00:25:48)  
find the key that we are interested in. Let's say we are interested in apple. We got it. We got the offset around it. We went to the corresponding offset in this gigantic 1TB file and then we read this entire line from it. Get the meaning and send it to the user. It's like, but you are doing linear scan over here. You still doing linear scan over here. How is this better? This is better because here when you are doing the linear scan, you are also going through the meanings of the word.

which are going to be huge. While you're doing linear scan over here, what's happening is you are just for each key that you have, you are just having a four byte offset that you're storing against which, right? So you're not editing around MVs of data, but rather this just four byte offset. So it's very fast to read this small file sequentially, worst case sequentially as compared to this gigantic file. Because when you do file IO, you always go in a sequential

way over there. Right? Okay. Now can we like, should we do this sequentially? Let's see, can we fit this index in memory? Let's do some basic computation to see how big this index file would be. Right? Harsh, I'll pull you in. ⁓ How we need to estimate the size of this index. The number of words that we have is 170,000 words.

Sneha Mehra (00:27:16)  
and what next? How will you compute the size of this index?

Sneha Mehra (00:27:22)  
Harsh. hi Arbit. So I was just thinking one very small way to do it. Like in this index, can we have that 26 entries like ABCD, ⁓ EFG something and their indexes or location? But then we only have 26\. But then still you would have to go to those those all the words for Apple in order to find that one word, right? Yeah, but we at least know about ABCD. But then isn't this better?

given that you already know Apple, you know the offset to directly go to that line and read it. ⁓ Here means we have all the words, right? That's concern. But let's assume, let's, let's compute the size of the index now, how big that index is. The size of each entry in the index is how big the length of the key and the length of the offset, correct? Correct. So what's the length of the key is the length of the word that you have.

The length of the word on an average in English dictionary is 4.3 something, 4.3 characters, average length. So average length is 4.3 plus what you are storing an offset. Offset is integer, integer will take up four bytes. ⁓ So 4.3 plus four is roughly the size of one index record this much. Correct? Correct. ⁓

And now let's say you store it in a new line. So you add two more buffer characters, one for comma and one for new line that you have. ⁓ So ⁓ these many characters or these many bytes per index, you are having ⁓ one entry for each word that you have. You have total 170,000 words. So this would be the size of your index. Okay. Do this multiplication and tell me what the size of the index is.

Open a calculator and do this multiplication for

Sneha Mehra (00:29:24)  
Yeah, take a look.

Sneha Mehra (00:29:42)  
Now you 1,70,000 into 10.3 ⁓

Sneha Mehra (00:29:53)  
17,51,000 17,51,000 17,51,000 Tell me the exact number. ⁓ 1751000\. These many bytes. ⁓ So if these many bytes is the size of the index file, it seems too big. Let's convert it into KB. It will be 1751000\. ⁓

KB which is equal to 1.751 MB. That's it. Okay. Now, will you worry about it? 1.751 you can fit it in RAM very easily. Right. ⁓ So this is what I wanted to showcase that this is why ⁓ number crunching is very important when you are designing any system. This way, what seems to be very huge 1 TB big file

or 170,000 words when you do this computation you found out hey this file is indeed one terabyte big but the index that we creating is like can fit in just 1.751 MB which is peanuts which is nothing you can literally load this index in memory store it in disk while processing load this in memory so now the things that you're trying to do like A to Z ⁓ you don't really have to

because you can fit this index right in your memory. Given a word, just look it up. It's like create a hash map out of this in memory, given a word, check the hash map, get the offset, go to this file, read the corresponding line from that offset and return the meaning to the user. ⁓ Life becomes so simple. And this is why I wanted to show like this one simple number crunch that we did. It helped us make our system very better.

Because what has happened is instead of overcomplicating things, ⁓ we just keep it very simple. ⁓ That one level of indexing. ⁓ first of all, just to converge on this discussion and give you a quick summary of it. We started with storing this entire thing in one gigantic file in a simple CSV format that we had. ⁓ Then because now storage part is done. Now when you're trying to query it, when you're trying to query it, what we found out here, we have to do linear scan every time.

Sneha Mehra (00:32:19)  
For every query, for every read request, if I'm doing this linear scan, sequential scan in this file, on an average, I would be reading 0.5 terabytes, like half of it, 0.5 terabytes on every request, rough estimate. Then for every request, 0.5 terabytes is very slow, 0.5 terabytes is 500 GB for every read request. It's going to be very slow. So whenever our reads are slow, the only thing that we can think of is indexing, and that's the good way to approach it.

reads were slow to make it possible to create an index. In index what we store? We store the key and offset at which it is present in the file. ⁓ Now we think of doing let's say let's keep it simple. Let's start to do linear scan on this thing every time we get a request on this file. If we do that every time that we get a request, we do a linear scan on this file to figure out a key and then get the offset and then query it. What would happen? We think a isn't this slow, but in order to see

If this linear scan is going to be slow or not, we find the size of this index. The size of this index is the number of entries that we would have, which is equal to the number of words that we have multiplied by the average record size. ⁓ It's English dictionary that we are building. In English dictionary, the average length of a word is 4.3 or mostly 4.7, either one of this, but it's not more than five definitely. That's the average one. I actually downloaded an entire dictionary and then computed the average length. That's how I got to know.

⁓ So it's 4.3 as if it's out of 4.3, 4.3 something like that. But 4.3 plus 4 by 2 store this integer. So 4.3 plus 4 plus 1 for comma and 1 for newline. So it became 8.3 plus 2 is 7, 10.3. 10.3 into 17, sorry, 170,000, which gives you just 1.75 MB. Given this, we can very easily see we can literally store this index in women.

Like we can load this index in memory when my API server starts, request comes in. I check this in memory hash table where this index is. I get the key, check the offset, go to this file, read the value, read the meaning and send it back to the user. And this makes, ⁓ this now makes our get ⁓ much faster because we are not doing linear scan over here. We directly doing a pointed read on this file with the help of this index.

Sneha Mehra (00:34:44)  
Right. Okay. Okay. This is the first part where we sorted out our storage and our querying side of things. Right. Okay. Now let's talk about portability. Now here you see we have now two files. Data.dictionary. Right. We have the two files. We have index.index of dictionary and data.dictionary. Right. In index file we have the index. In data file we have the data.

Now for us to ship this quote unquote ship this dictionary to someone, anyone. I am having two files at the moment. This breaks the portability thing. We want this to be just one file. How can we go about this? Anyone except Rahul and everyone because these folks are quite active. I'm just probing someone else to participate. Anyone else who wants to do

We have two files. We don't want to file. We want one file. ⁓ Alok. ⁓ Do we need to create a file for the index? ⁓ I mean, ⁓ on the system startup, can we just build it up in memory? ⁓ So which means on every startup you have to iterate through this entire file once at least. ⁓ Yeah. We can avoid it, right? Because why do we even do that? Because this is not changing anyway. ⁓ Right? This will change once every week. Right? This file.

I mean, then basically it is about ⁓ merging these two files and having it ⁓ in one file. ⁓ So having some kind of partition between ⁓ some kind of partition between using which we know that till this line, this is the index and after this is the dictionary. ⁓ So now what we do, we have this. So we merge it. Okay. But now that's now a new challenge would come in. So now you are merging this two file into one gigantic file.

in which you have index, you literally concatenate this file and you have this gigantic data that you have. Right? Okay. But how would you know when your index ends and when your data begins?

Sneha Mehra (00:37:00)  
There has to be some line or some... ⁓ Some special line? Yeah. Okay.

Sneha Mehra (00:37:11)  
Any other approach?

Sneha Mehra (00:37:15)  
So, I mean we can pre-allocate the space for index because the words are, ⁓ I mean we know the size of the words and only few words will be ⁓ added. What do you mean by pre-allocate? Pre-allocate in the sense like we can fix the size for the index. You are giving it a fixed size. So for example that I know that my dictionary but what would be your fixed size? Right now it's 1.75 and MD. Yes. ⁓ I would be giving more to the index more than this. ⁓

that is currently there, but ⁓ because the new words will be added very less. correct. So you are saying that, Hey, instead of me worrying about to have a separator or a special line that denotes that, my index ended and now my data begins. What you are doing is you are giving a fixed weight to this index that when my application starts, when it reads this gigantic file, it knows that the first two MB is going to be my index. Yes. Right. Okay. That's a good approach. That's a very nice hack.

But it's not space, although you're wasting very little space, but let's be a little more frugal about it. ⁓

⁓ Got off. ⁓ What else can you do? ⁓ These are two good approaches. ⁓ What else can you do? ⁓ Building on top of it, we can have like data is stored from top to bottom, bottom within the file and index is stored from bottom to top within the file. ⁓ How would you know where one ends and other begins? You still want to know,

Sneha Mehra (00:38:47)  
up until where should I read to load the index. So let's say you store index like this. How would you know when to stop reading the index? Like when to stop so that you have built this in memory hash table.

Sneha Mehra (00:39:03)  
⁓ Imagine I'll just give a quick example. So let's say this is how you have concluded. So first index and then data all in CSV format. Right. Now when you start reading this line by line, you created hash table, hash table, hash table, hash table, hash table. You put it entry there and then you start adding over this. First you had an hash table, Apple comma O one. And then when you start, I did it this, you got Apple and then you replace it with a fruit or something. Right. So you need to be a little more frugal about it that

You need to know when to stop. So what I meant was that ⁓ index can grow like a stack, which means from the nth byte to n minus one and minus two, so on. And data can start from zero, zero. So then you are wasting this much of space. So this was a space might be an unutilized. ⁓ Today it might be enterprise, but someday. But one, but the day when these two ⁓ are right next to each other, how would you know when one finish and other begins?

Yeah, having a separator. ⁓ So you are also going with separator, ⁓ right? Yes, yes, yes. ⁓ Yes, Amrita.

So basically I was thinking that since the data is always going to ⁓ grow, but ⁓ we can choose a specific index to store that information. So maybe at the zero index or the first set, we can choose to store that separator information and then update it as we add on more data. Okay, that's interesting. So what you're saying is that add something in the beginning of this and specify what the separator is.

Correct? ⁓ Okay. So now if you do this, way ⁓ to read this entire file would be I first read a separator, then I start reading the data until I find it separator and that much becomes my index beyond which becomes my data. Now ⁓ I'm saying, can you change this to store something else?

Sneha Mehra (00:41:09)  
That would be much more beneficial for you.

Instead of storing the separator, can you not directly store the byte offset?

Yes, we can. Yes, it eradicates the need of a separator. And then ⁓ the advantage of it is let's say you store the byte offset.

of data. So place where the data begins. ⁓ If you store this information, you exactly know when your file loads or when you are trying to read this entire file, the first step is to read this first chart. And this is a classic way ⁓ of having a header. In any file format of the world, PNG, JPG, MP3, AVI, ⁓ everywhere

the first few bytes, it depends on the protocol. It depends on the format that the creator of that ⁓ image format or whatever format specifies what that should be. That first few bytes denotes or contains the meta information about everything else in the file. So what you just spoke about Amrita is to add this header whose job would be to store this meta information about the file that you are creating, the singular file that you're creating.

Sneha Mehra (00:42:32)  
because it contains multiple sections, are storing the information about each of the section in this header. I'll give a small example. So in this header, we might just store that this is ⁓ our dictionary, fancy with some ⁓ version number. ⁓ And now this has to be a fixed size header. It cannot go beyond that. And you have start offset ⁓ of data.

or other you may insert store index size and how big your index is. Right? So let's say this requires you one, two, three, four, five, six, seven, seven or other eight bytes to store this like fixed eight bytes. So first eight byte would represent O U R D I C T for example, it's a name. ⁓ the, in the PNG format, the first three bytes is P N N G. So that tells the

⁓ decoder that this is a PNG file. ⁓ MP3 has some other bytes and then it is MP3 and so on and so forth. ⁓ So the first eight bytes will denote that this is our dictionary like whatever name you want to give it. You can give it. Let's say I build my own database. I want to call it AODB, Arpids outstanding database for some reason. So this would be my first four bytes and it's my decoder who would understand it's AODB file. So I know the format. I know what components would it hold.

⁓ Then a version number, maybe you might have multiple variants of this file. You may want to specify if it's version one, you have index first and then data. Maybe in version two, you have data first and then index. So depending on that version, the decoder would understand what to do with it. ⁓ Then third, you might want to store the index file. So depending on this, are choosing that the next, let's say version number is four bytes as an integer. The next four bytes is index size. That how big is the index?

When you do this, you store how big is the index in this one, what you'll store, you'll store the value one points like this, this value, one, seven, five, one, triple zero. This is the integer that you will store in this space as the index size. So now your header is a fixed with eight plus four plus four equal to 16 byte. So this becomes your header. So this is the header that you just added, which is 16 bytes. So the flow of our application would be

Sneha Mehra (00:45:00)  
When your application server boots up, assume it has this file. When it has this file, it first ⁓ reads the first 16 bytes of this file. It sees the header. It sees that, hey, it is ⁓ our dictionary. First eight bytes it read. It said, hey, it's our dictionary. sorry, what first byte? We know that header is fixed. ⁓ We know that the header is fixed ⁓ 16 bytes. So it would read the first 16 bytes.

Out of which it would find the first 8 bytes and understudy it as our dictionary. The next 4 bytes is version and the next 4 bytes is the index size. Once it knows that this is the index size, it would read those many bytes starting from this location, these many bytes and it would read it as index. Then it can iterate it and create a hash table out of it. And from there, everything is data. ⁓ And which is what these offsets would be storing over here. So this

makes your entire dictionary shipable. And this ⁓ is a classic way to build any custom file format where you are storing a header and information about different sections that you have. Here the different sections are index and data. In case of PNG, it is something. In case of JPG, it's different frames that you have. In case of MP4, it would be where is the video part? Where is the audio part? Right.

And in that video format, you would know that this is the first one. This is the second term. This is the third frame in audio part. You would have the waveform information, right? ⁓ This is literally creating compartments in the file that you have and having a header that tells you where, ⁓ which compartment likes. ⁓ And this is the one file that you can very easily ship across systems where you'd want or give it to your end user. ⁓ Anyone having this decoder can read this.

⁓ And this is the point that I wanted to bring out that how this particular format becomes really important. And it's very general wherever you see anything. For example, tomorrow, what Heman bring out that word, if you are exploring something around data engineering and you come across Parquet format, ⁓ Parquet format also has some headers and information. ⁓ Jason does not have any

Sneha Mehra (00:47:22)  
If you look carefully in the JSON file, JSON file that we all use, it's a simple text file, right? Because it does not have any headers, you have to read the entire file. It is in a specific format and then you have to decode. So the JSON decoder decodes curly braces, codes have to its string, if it's not then it's integer and then converts it into it and then it creates a memory dictionary and gives it to you. Because it doesn't have a header, you have to read the entire file to get the information out of it. CSV in general does not have a header.

Right? But, tpt format, pdf format, everything else has an header. And it depends. You can define your own custom format. I tried to write my own image format for some reason I did that. ⁓ So that's how I got to know this. But this is again really important. When you know, when you see a file, whatever file, whichever file you are picking up, it has its own encoder.

and decoder. So encoder creates that file in that specific format and decoder decodes it. Image, video, audio, pick your favorite format, holds true for everyone. So when we say, ⁓ you might have seen this error, early days of internet to be honest, but hey, this web browser does not support PIFF format. What does this mean when you say that this web browser does not support a format? It implies that

it does not have an encoder or decoder for this file format. What does it mean? It does not have a decoder for this file format. Given this set of bytes, I don't know how to interpret it. But what we just wrote is for our format, the decoder that our decoder, what our decoder would do, it would know for 16 bytes is header through which I would get the size of the index and from which I'll get the data. ⁓ And the offset in this will contain the offset in this data file. ⁓

This is the decoder logic that we would write. This holds true for every file format in the world. ⁓ But now because of that one constraint that I added portability, you see that how we were able to create this ⁓ as just one file instead of shipping multiple files. ⁓ So that does this. Now we'll touch upon infrastructure side of things. So now here what we build, we build an understanding ⁓ of

Sneha Mehra (00:49:45)  
We built an understanding ⁓ of something really interesting. Our file format is importance of portability, headers where they come in handy, indexing to make reads faster. ⁓ Now let's come to the next part. How do you push it? How do you serve the actual user request? ⁓ I'll lay the seed foundation and then we'll improve on the dessert. So ⁓ what we may think that our end user who will be querying on this obviously

we would have a load balance or something where after which you get a bunch of... I had to create a load balancer.

Sneha Mehra (00:50:26)  
So let's say we create a load balancer request from an end user comes to load balancer. Behind the load balancer, we have bunch of API servers whose job is to serve the incoming request that are there. So now what would be the request flow look like? So this is a load balancer on which you got the request. These ⁓ are the API servers, API server one, API. ⁓

not a good place to write.

Sneha Mehra (00:51:04)  
Okay, right. A classic case, a user, a load balancer and an API server. ⁓ This is a file that we have. Now, where would we store this file? For us to serve this request or handle this request, we are exposing a get endpoint. So slash word is we are exposing and we are doing a get on it and we get the meaning around it. Right. So this is the request that user fires and I want to get the meaning from this.

The request comes to a load balancer, load balancer forwards it to one of these instances. Now what would this instance do? Where is this file queried? ⁓ We defined a file format, now we need to provide a way to query it. ⁓ And way to fire HTTP request on that. How does that flow look like? Anyone?

Sneha Mehra (00:51:56)  
Yanoba. ⁓ So can we have a shared ⁓ disk storage for all the API servers that they can query to ⁓ get the file? But why not store this one file on every server? What's the harm in that? ⁓ That would make the update process a bit more tedious, but if we are fine with that, then we can also do that. OK, it would make update tedious. Why?

So whenever there is a change in let's say a few hundred words, we have to make that change happen on each of the APS servers. Some of the APS servers might not be available at the time when we are trying to make that update. Then we have to wait for them to be available. updates are updates are not real. The updates are weekly. Is that still a problem? Every Sunday you are making one update in bulk.

Sneha Mehra (00:52:55)  
Right. It's not a key value store where you are right now. can update any time you like. Correct. It's a weekly update. Every Sunday you'll get a file from Oxford. You would say, okay, these are the new words. These are the new meanings. Go and update your database. Right. Yes, that that's great. But like what another angle that I'm thinking of is what if like the updates fail for a particular set of a file, then we would have data inconsistency. So ⁓

Let's say what what drastic would change if there is data inconsistency because meanings are not dramatically changing. ⁓ Point is valid. You may have inconsistency in data in case your update partially fails, but it's a meaning. So Apple, you might have a meaning which says a fruit tomorrow. You might have a meaning that says around fruit. Is this an end of a word? No, right. ⁓ It's okay. ⁓ But it's a valid point. I'm not denying it.

It doesn't have probing you more to think on those like, what else is wrong with this approach? And why are you going? Because what I'm trying to do is when you are, when you are proposing a solution, you should have thought of other disadvantages, like, ⁓ sorry, you should have thought of disadvantages of other part in depth, which is what I'm trying to prove. ⁓ What else could go wrong? So you said update tedious, correct? You said data inconsistency, but in both cases, both of them are not a problem yet.

What else is the problem?

Sneha Mehra (00:54:33)  
If you store the files, the same dictionary file across all the three servers, what else is the problem? Like we are having a lot of duplicated costs of saving the file, 1db file on a lot of servers. So which we can optimize for.

biggest people, costs.

Biggest inconsistencies, okay. Updates, tedious, okay. Cost. A real world system design is all about doing things, doing the bare minimum things at the bare minimum cost. Always remember this. If you're not thinking about cost, then you are not building good systems. you cannot, because ⁓ in order to make anything faster, just throw a lot of money at a cloud provider, get a bunch of RAM, store everything in memory, problem solved. Right?

If I can have one terabyte RAM over here on each of this machine, I can just load the entire dictionary in memory. No need of creating index, nothing just one hash table containing one terabyte of data. But that's not a good solution. Why? Because it's not cost effective. So cost is something that should come to your mind. ⁓ the first thing that is ⁓ the solution that you're designing is this cost efficient or not. ⁓ Okay. Why is this cost inefficient?

because you ⁓ are duplicating the data that you have in external disk that you are storing and you would be storing it in SSDs. SSDs or any disk is very costly. Cost of one terabyte SSD that you need. And let's say if your dictionary is a huge disk and you have hundreds of servers, so you are having hundred terabytes of redundant data. Problem, right? So there has to be a better way to do it. So this is why.

Sneha Mehra (00:56:20)  
You can think of not going with storing over here and the thing that started with underquark that have a central place. Where you are storing the dictionary. Okay, now what this central place would be how will you access it. ⁓

Sneha Mehra (00:56:42)  
⁓ So ⁓ can we use something like Amazon S3 in this case to store this file? Okay. Why S3? So, I mean, it's very cheap and optimized for blob storage like blob files. So I'm thinking like since we have a single file, we can just store it there. probably like ⁓ one more thing that I thought of was that

⁓ since it's a big file ⁓ in S3 like I have seen that the get object query ⁓ get object has some query params where we can actually pass the bytes that we want to read. So since we already know the index size, we can just pass that these many bytes I want to read and get that and finally get the offset and then ⁓ again ⁓

get the same query with that particular offset byte and read that and give it to the user back.

Great, great, great, great, great. Okay, so I'll just pull it in a minute. Okay. So what Pankaj said ⁓ is something very interesting. What he said is can we because we what we have defined is it's not that we have converged on S3 yet, right? We still want to know what are the core properties we want from the system and then we would pick a particular technology, right? So what we want, what first Anubhav said, hey, let's put it in a shared storage.

Why do we want to share so is because we are not ready to obligate a data across why because it is cost inefficient and other problems come with that. that's okay. So given that we cannot do it. We wanted in a central case, but remember our discussion last week. What is common across multiple servers ⁓ network, which means you need ⁓ a common place, which is accessible over network. ⁓

Sneha Mehra (00:58:41)  
And what is that one common place that is accessible over network? Any block storage or any database, but we cannot use database. So we have block storage and anyway, we are storing a gigantic file. We can just store this file on S3 for example, right? We're still not giving up on portability. We still have portability. We still have this one file on S3. But now, but is this the best choice? What is the feature that we want? If this was a normal disk, assume it was a normal disk. We were, designed our entire approach.

thinking it's a normal file and I would read a set of bytes and then another set of bytes and then do ⁓ a read at or do a random read in the data section, right? Because we are reading this header to understand how big the index is. We're reading this entire thing to build an in-memory hash table and then doing random reads on this ⁓ data section, right? So this should be supported by whatever we use over here. So generic term for this is

A network?

attached storage. So what we were looking for is any storage that is accessible ⁓ over network ⁓ and it gives file system semantics.

And the key things that we need is we need random reads. ⁓ That's a bare minimum we need over here. ⁓ That we should be able to read what I've been doing in my random read. ⁓ It means that I could provide a thing that says that from this file at this offset read n bytes. I need to have somewhere to fire this. ⁓ A normal disk gives you this. You can open a file, seek to that location, ⁓ read those many bytes. ⁓ Now this needs to be provided by

Sneha Mehra (01:00:31)  
any network attached storage that we are using. ⁓ So you can just use the network attached storage that gives you this thing. ⁓ S3 gives you all the file system level semantics because of which it can be used as a distributed file system as well. ⁓ So what you get this feature on S3, where S3 you are not just there to read the entire file. You can say that in this file at this offset read these many bytes. That would give you exactly those many bytes that you are needing. ⁓

You could choose, it's not that I prefer S3, it's just that you can replace S3 with any network attached storage. For example,

and HDFS. HDFS gives you this exact same semantics. Next, can use ⁓ a literal physical hard drive connected in your data center acting as this thing. So literal network

hard disks that you get like Dell network hard disks are very famous. ⁓ You'll get this. You can use anything that gives you file system level semantics to do this. S3 gives you that, everyone understands this, that's why I'm going with S3. So now on S3 what we'll have is we'll put this file having header, index and data in this one file and we'll store it on S3. And now what our reads would look like. Now the entire flow, end to end flow, what it would look like is

When any of my API server bootstrap, what it would do? It would read the first 16 bytes of this file. So first of all, now this comes one step back. What each API server need to know is needs to know the file path on S3. So that can be a static configuration. This is the dictionary that we are serving. It can be a static configuration, but we'll go deeper into this because we still have to do deployment of this. ⁓

Sneha Mehra (01:02:27)  
So what we want is we want a file path on S3 that where exactly is the dictionary stored on S3. Right? So let's say we pass it as a configuration. That this is a configuration. This is the path where it is stored on S3. Given that APS server starts with that, what it would do it from this path, it would read the first 16 bytes. It would read the first 16 bytes, interpret it as header. It would understand it is a dictionary file like whatever our format is.

from there it would know the index size it would read then would make another S3 call to read those many bytes ⁓ and create a hash table in memory hash table over here. So now what we would have we would have an in memory hash table in memory hash table in memory hash table over here. Then it would start serving the request whenever a get request comes in goes to any APS ever it would refer to this in memory hash table if the word exists good enough if it does not exist 404 right if it exists

In the in-memory hash table, it would have that this word is present at this offset. So what it would do? It would take this, know the offset, go to that particular offset on S3 and read this file. How would we go? ⁓ S3 API gives you that. In the S3 documentation, you can do a random read ⁓ and it looks something like this. On this file, at this offset, read these many bytes. You can issue a command like this. But in case you are replacing S3 with

Let's say HDFS, HDFS has the exact same semantics, network hard disk has the exact same semantics. Right? So you can seamlessly read the bytes that you're interested in and nothing before that, nothing after that. So you're doing the bare minimum I over here. And this is how your normal gets would look like. But if you look the beauty of the system, the beauty of the system ⁓ up right now is what you were just being able to do was to have

pointed reads on S3 very efficient reads. If you look carefully, what we just built is we are now able to on a ⁓ specific record from S3 just that much without reading the entire file. So we are just built our database because what a database is key value store, right? We just build a key value store that supports huge amount of gets.

Sneha Mehra (01:04:54)  
on S3. So we got this ability to fire pointed queries on S3 without us having to read this entire file and that is magical. It has a ton of applications and we'll talk about them. We'll talk about how to extend this part into building various applications out of. Right. Okay. Now this is still not done. There is still a big chunk remaining. Updates. Right.

Now for us to have seamless updates, I'll just copy this diagram because we are still not discussed on updates. So we discussed on storage, we discussed on query, we discussed on high level infra that it would look like. But now what about updates? Now the updates that we are having, they are weekly updates, right? Updates are weekly, which means every week,

I'll get a file from Oxford which has the words that have changed and the new meaning for that. Apple, chain, let's say ball, change, and let's say zoo change. These three words changed. Right? Now what would we do? We have to apply these changes to the file that we have on S3. How would we go about it?

⁓ Yeah. ⁓ as this is a weekly update, so we can have a bad job. ⁓ so, ⁓ so, that will, what it will do is it will add those newly words to the end. ⁓ Yeah. ⁓ can have a cron job also. ⁓ See bad job is batch processing. ⁓ Yeah. ⁓ Periodic. Yeah. Correct. Periodic job. ⁓ There is a update from Oxford.

We can trigger that job or we can also have the job also. And what it will do is it will update the dictionary and at the same time it will update the index and so it will repackage the whole file, upload it to the S3 and where, where, where, it would update what? ⁓ Yeah. So we can have, ⁓ on another machine. ⁓ Now we're talking. Okay. You're spinning up another server. ⁓

Sneha Mehra (01:07:11)  
Now and in that file whenever we receive a new ⁓ update so we can ⁓ recompute this whole file. Do you have change log on this server? ⁓ I will add these change log files to the dictionary. So where is dictionary? ⁓ Dictionary means we have to pull it from S3. Correct. You'll download it from S3. ⁓ Then update the version of it. ⁓ How? ⁓ Because we know that the ⁓

⁓ In the header file we have the version store right at the initial buy. That's okay. That's the that's the last thing that we'll do. ⁓ But how are you updating the dictionary? This one. How are you updating? ⁓ So we are so are we getting only change log or the whole okay we are only only change log only change log right. Right. So we can append it at the end because our dictionary is not sorted.

Sneha Mehra (01:08:04)  
But then in that case, what would happen is the dictionary would be ever increasing. Plus it would have duplicate entries. And I just added that you cannot have repetitive entries in the dictionary that you're serving. Okay. ⁓ So now. we can iterate the whole dictionary once ⁓ and ⁓ means check that if the file, if the ⁓ change log has duplicate, what is keeping ⁓ or if the

or meaning or the meaning of the word is changed and replace that. ⁓ Or is it really that simple? ⁓ Think of it. Think, go into a little more details of it. ⁓ We can also have a hash of it and compare the hash with the existing one and new one. ⁓ You directly have the see change lock anything has what needs to be changed. Simple.

So you don't need hash sort of comparison. see again, as I always say, it's easy to build complex systems. It's very difficult to build simple systems. ⁓ Think all the things that you really need. Hash, hash, you don't need it because change log directly contains the changes that are that needs to be made on the dictionary. ⁓

So how will you apply the changes that you have with the change log in this file? What are the challenges that would come in? Challenges is to generate the disk actually. ⁓ Why do you want to generate a disk? Change log already has this, right? You just want to apply the changes over here. What do you mean when you say you to apply these changes? Yeah. So, ⁓ so we, ⁓ we already have an index right here. ⁓ We already had it this. ⁓

And we have a change log. what we can do is, ⁓ so let's say change of meaning of Apple, right? So we can use that index file ⁓ and ⁓ use the index ⁓ and apply those changes ⁓ or replace that the whole line of ⁓ the PSP. Okay. You're saying that if I have this file and let's say I have the word Apple over here and I have meaning of fruit ⁓ and then let's say I have ball.

Sneha Mehra (01:10:17)  
and I have a round object. Right? Okay. Now let's say my change log contains, I should have drawn it. Let's say my change log contains that apple and now the new meaning is a round fruit. Now how will this changes be applied over here? You say that using index, let's say you know the offset when it should be changed. Now what will you do? Will you replace this thing by this? Yes. Okay.

What happens when you replace this by this? Will this entry move down? Yeah, we have to. Yeah, correct. We have to do that. that happen in file IO?

Sneha Mehra (01:11:01)  
What happens in file IO? Let's say if I write around fruit, what would happen? It would look something like this. So this is how, this is the limitation of how this works. Right? ⁓ So when you go to a particular location and start writing, it would blindfoldedly write the bytes like in the IDE, when we add, can hit enter and move other lines down because it is in memory. In memory, can do that on this.

You cannot. So when you go to this particular offset and start writing, Apple ⁓ around.

⁓ fruit you are overriding the existing one so you cannot just say I would replace it because it's not as trivial so how do we go about

Sneha Mehra (01:12:06)  
Thanks, but thanks, Vikram for bringing this up. is see, these are the details that I want all of you to focus on. And because this is what makes us better engineers, because it's easy to just say I would replace it. But when they are going to write the code, that's what I always say. ⁓ An engineer, ⁓ a good engineer or a senior engineer always wear three hats. ⁓ Product manager, ⁓ architect and an engineer. This is what an engineer hat looks like.

You go because imagine you are the one implementing it. What code will you write to implement this? And when you think of implementation, you realize the things that would break. Okay. Let me pull in someone who has not spoken today. Yeah, Jyotindra. How would you solve this problem? Can we have sort of like graveyard entries, ⁓ which sort of replace the currently existing entries there? For example, if we have a new meaning for Apple, we

appended to the end of the file and the previous no repetitive entries. Okay, so ⁓ I was thinking like you just sort of waste this waste this space for some time ⁓ and ⁓ later run some sort of compaction process after. But we don't see it would work well when you are doing live updates. This is anyway weekly updates. So have it very clean up it directly. ⁓ Yeah. ⁓

Hello. Yeah. Yeah. Very naive approach would be to, ⁓ just, yeah. A very naive approach would be to look for every word, ⁓ get the bite where we are supposed to like enter it. ⁓ And then, ⁓ so what we are doing is we are creating a new file now, ⁓ add that change in that file ⁓ and then append all the rest file to it.

And then ⁓ I mean, would be a very naive approach. No, no, it's not naive. It's not naive. It's kind of like merge of merge sort. Correct? Yeah. Yeah. And that's exactly why you should do it. That's what the best approach is. I'm not even kidding. It's not an naive approach to be honest. It's not a naive approach given the design that we have. We're not using fancy B plus C's over here. We're going with very simple approach, but that's how it works. Right. You think of a simple approach.

Sneha Mehra (01:14:28)  
And that's what is going to be really efficient when you make things complex, it has its own set of disadvantages. Right. Okay. So this is my end. The requirement to do that merge of my sort to be to be efficient is that your file needs to be sorted. I just quickly walk you through on what this is all about. So what we'll do is we'll create a new file ⁓ out of the existing type. How? By merging the two files. So we got a change log.

Now change log will be a very short file. ⁓ Assume it is sorted. Even if it is not sorted, you load it in memory, ⁓ sort it and dump it because it's going to be very small. is changed log. It's not going to be gigantic file anyway. ⁓ So that becomes your change log. ⁓ Then you have an entire dictionary that you just downloaded right now on that new server that you spin up. ⁓ Now what do you do? This is sorted. This is sorted. You create a new dictionary out of this, a new file out of this by iterating the two. ⁓

⁓ that famous ⁓ linked list problem merge two sorted linked list ⁓ into third sorted array. This is that. This is merge of merge sort. How you are merging two sorted arrays together is you have a pointer over here, you have a pointer over here. You iterate both files line by line. You see the first entry is a a dash. So these are words and these are meanings. I'm just simplifying it over here. So these are not

sections like ABCD, but these are words ABCD, FJ just explain, it's easy. ⁓ So these are words, these are meanings. These are words, they are updated meanings. ⁓ So what we do, we have two pointers, we added both the files line by line. And then we say it's AA dash CC dash A is small, I'll write AA dash over here. ⁓ Then I move this pointer down, BB dash CC dash B is small than C, I add BB dash over here. I move this point over here, CC dash

CC double dash. This is the updated one. Both matches. This is the updated one. So I write CC double dash and I move both of them forward. Then DD dash EE dash DD dash, because that is small one. Then F F dash EE dash E is small one. I write EE dash and then move forward. ⁓ F F dash F F double dash. I pick the updated meaning right over here. I completed my, I exhausted my change doc. So every entry after this just adds over.

Sneha Mehra (01:16:56)  
A classic merge of merge sort. Nothing fancy like merging two sorted arrays into third big sorted array. This is literally that. But here, one point, you don't have to load this entire file in memory. All that I'm referring to merge of sorted array. Here you don't have to load it in an array. You literally iterating file line by line and creating a third file out of it. Given that the way file IO works,

is when you write it may start to override the next line. If you go beyond that range, you have to always ensure that you are creating new entries over there. So we created a new now with this, we create a new file from the existing dictionary plus the change law. Now this puts a constraint on us that because this dictionary is one terabyte big, this dictionary will also be one terabyte big. The disk of that server that we just paid up to run this job should be at least two terabytes.

at least two terabyte, right? But that is just one server that we are doing that, right? Okay, so this is what we will do to create a third file over here. Okay, so let's say in this server itself, now we created the third file. This is the new file that we have. Now what? Now how would the changes that are there in this server

go to S3 ⁓

Okay, so now we have to start from we have the merged file available. ⁓ And now just just a bit, ⁓ which means that now that we have this, we also have headers, headers, index all sorted for this file. So this is the new dictionary that we are ready to upload to ST. So now now the critical part, critical thing is that

Sneha Mehra (01:18:56)  
We have to tell the clients who are using this thing that we have a new file from where it is created. So it's like you have to swap the ⁓ existing file, which was supporting to this new file, considering this new file has ⁓ previous data plus the Delta that came from channel. ⁓ So what it means is ⁓ we have to ⁓ simply, ⁓

change the pointer, ⁓ but looks like ⁓ it will have a downtime because that time and you will say, okay, this is the URL that was available when you put a new S3 ⁓ file, there'll be a new URL. ⁓ So ⁓ the API service which were hitting NAS or anything where we are doing this thing, we have to figure it out.

deployment process, how we have to upgrade. ⁓ Now, another deployment process, which we know is either you simply say, I will change everything in one go. ⁓ Or we have to say a rolling upgrade where we'll say, okay, I will change few of the files in few places, but it is a single place. ⁓ I don't think that ⁓ that kind of a deployment will work.

what it looks like a completely a to b when ⁓ you are there ⁓ you can have the new file updated having a new s3 URL created and once it is ready this has to be given to the API servers saying this is the ⁓ new ⁓ URL or new address you have to consume it. ⁓

Now API server has to ⁓ change that point and ⁓ they can do ⁓ the normal deployment either they can do a first canary where ⁓ one server only does a change and sees everything working fine and then only moves to others. ⁓ think these all things I am getting in my mind to solve it. Okay. Let me prove you on this one point. Why do you want to upload it to a new S3 path? What would happen if you upload to the same path? Won't it simplify your solution?

Sneha Mehra (01:21:23)  
Okay, we ⁓ then ⁓ for that for the time when we are uploading, I mean, the sort period of time, ⁓ maybe we are not able to serve it. do not know how as three internally works. So let's say you operate it file at the same location because you're the things that were making your system complex, but when you were adding it to a new path, then you have to tell API servers to go and update a configuration and whatnot. Right. What if you upgraded to the same path? What?

Could go down. Yeah. the first thing is ⁓ always when you think about a deployment, you need to think about the rollback part. are going ahead. You need to worry about what will be the my process ⁓ if let's say that deployment fails. if in a same path I'm changing and during the uploading of the file on network, let's say the file got corrupted or something like this, I have to proactively observe it.

using some observability of monitoring and then immediately will provide it. Okay. Go to the previous file only. ⁓ and, ⁓ and that, that can be a ⁓ one ⁓ critical thing to take care. ⁓ Other than that, I, I, I'm not able to think anything that we have to worry about, but ⁓ definitely making sure that the transition, I mean, this file transfer goes properly. ⁓ And if it not go, we have to roll back and provides

the previous file is necessary. having a temp file, mean having a temporary storage where you put this file and then upload it, check it, is working fine. And ⁓ then then market successful. That's ⁓ that's an important thing. I required. Thanks. Thanks for adding all those points. But now we see like you as it everyone now everyone sees that again, as I said, drawing boxes is easy when you go into these details. This is where

things become really interesting. You need to solve this problem where you are, where it's easy for us to say, we'll just upload it on S3. ⁓ But where and how and what challenges would we encounter? ⁓ Will the transition be smooth enough? Are we serving garbage data to user? These all details are super critical. Super critical when we design an actual system, which is I'm just proving you all to think on those lines. It's already one and a half hour.

Sneha Mehra (01:23:48)  
but it still not converges. Such the problem statement looks really simple. I'm sure when we started with this, ⁓ was just saying, what kind of stupid problem statement is this? But when we go into these details and again, as I say, these are all building blocks. You can apply to other systems that you are designed. It's not just limited to this one system. Right. So be very open to how you are absorbing things, what problem we had and how we solved it. Right. Okay.

Let me pull in, ⁓ okay, so, ⁓ but ⁓ before I pull in the next person, let me just converge to what Heman said in a simple order, right? So uploading to the same path on S3 had a lot of problems. First, what if upload fails? What if there is file corruption, right? When we are uploading to the same path on S3, there is one more challenge that

Now that the file that was uploaded onto the same path, now we have to tell this server to reload the index. Why? Because offsets would be mismatched. I'll give a quick example for that. Where is it? This. So for example, if we upload it to the same path on S3, this is what is going to happen. So let's say in index, index is loaded in memory on the API server. Right? API server first made call to S3, loaded the index already.

⁓ And then all the requests that are coming into the API server, there's just from index, is checking the offset, going to the location, reading the meeting and sending it back to the user. Now here what would happen is the old data file, it had that index, index was read from that particular thing. API server has this loaded in memory. So Apple is let's say offset 100 with byte 1024\. This is there. But now because

You uploaded a new file, assuming everything went well, uploaded a new file at the exact same location. In our, if someone queries Apple, it will go to byte offset hundred and read one zero two four bytes, which might be some other random word like the word before Apple that to starting in the middle or somewhere. So what user would see is some garbage because ⁓ the real offset of Apple in this new file, which is uploaded at the same part is different.

Sneha Mehra (01:26:13)  
So when you upload it to the same path on S3, you have to immediately tell APS to reload the index. But for that duration, ⁓ users will see garbage data. Very poor user experience. Right? Okay. So these are the things that I'm basically trying to probe you on. Right? ⁓ Think of these lines. So that is one problem. Now you would think, ⁓ let me

So uploading to the same path on S3, this is one of the problem. So now it looks like it makes sense for us to upload it to a new path. Now when we do it, when we upload it to a new path, what would happen? I'm just going back to brainstorm. So now let's say we upload it to a new path on S3. Let's say my path one on S3 was S3. slash, slash word dictionary slash dictionary one dot dictionary something, right? Let's say this was first part.

Now let's say when you upload this, you upload it to the new path. S3 colon slash slash word dictionary. And now it is dictionary two dot dictionary. Now how do you tell APS I was at this is the new path you should start serving. How do you communicate this? Got it?

Sneha Mehra (01:27:30)  
⁓ Rather than this, ⁓ what we can try for is rather than dict2.dict, we can have wd.1.1.1 slash the path. I mean, we can have the IP address. difference does it make? And then the path has changed, right? Yeah. ⁓ Right. You have to convey this path. Whatever path is, have to convey this to the API server that this is the new path that you need to reload. ⁓ Right? How do you tell this to API servers?

invalidating their cash. How?

and interrupt on, let's say a command which invalidates the cache. But now ⁓ everyone would have to tell APS to do this. ⁓ for ping all APS must invalidate the cache. Sounds too complex, isn't it?

So since we know what time our town job will be running, can't we have... That's unreliable. What if upload fails?

And what if you're not able to successfully do that? Right? ⁓ No. So what, what's the harm in that? Because API servers won't be ⁓ taking the index from the cache. would be going to S3. It's simply adding the latency, but not at the harm of anything else. Why not? What if your upload failed completely? What if your file was corrupt? You didn't check a lot of stuff then. Right? You cannot because

Sneha Mehra (01:28:53)  
If you do a cron job over there, no matter what Sunday at 1pm, everyone update. What if your job ran beyond 1pm? The upload was not handy. Now what? What if the file that was uploaded was corrupt? What if you didn't get time to validate? Because now this is an independent mind over here. And no matter what that would run and update it. That's wrong. ⁓

Sneha Mehra (01:29:19)  
No idea beyond. Yeah. Thank you. Okay. Yes, what will you do? ⁓ well, I, I came, in the middle where we ⁓ already set up the API servers, but, ⁓ I had an idea. ⁓ What if we had the each word in, in one file? No, we then we kill portability. That's why we are not having each word in one fight. Okay. Okay. So I, was, yeah. Yeah.

That was a constraint that we played with. We want portability. We want just one file having everything. Okay. I don't have anything. Okay. ⁓

Sneha Mehra (01:30:01)  
⁓ So I was thinking that we can have some extra file which will have like a metadata. That's the same file again, like it will be in S3 again. ⁓ And file, let's say ⁓ once we know that we have updated our index, we updated the new file ⁓ and uploaded it to S3.

Like the put request which will make in the success callback of that we can update this metadata. Perfect. What our API servers will do is like they will keep on pulling this metadata let's say at some frequency like 5-10 minutes. And they will at my index path has updated so they will start fetching it from a new file and upload their you know update their in memory index also. Okay first question will this path be static or will it be versioned?

So the path of the metadata file will always be static. Perfect. Okay. So that everyone knows that this is the only path where we have to go. Okay. Perfect. Okay. ⁓ Then you're saying that you pull the metadata file at let's say five minutes, 10 minutes. Yeah. What, ⁓ why would you do that given that your job runs weekly?

⁓ Again the same issue, ⁓ It's running weekly but we don't know like when it will ⁓ get updated and when it is getting, you know, uploaded to S3. I mean, ⁓ it might happen that we received updates on Sunday but because S3 was down, we were not able to update it to the S3 and we uploaded it on Monday. ⁓

So then basically API server will not know that now I have to fetch it on Monday. So it needs to continuously pull to find it. Now my index is updated. yeah. But we studied a second week, third week that we can make things wherever we have poll, we can make things reactive.

Sneha Mehra (01:32:00)  
Right? So instead of pulling, we can make things reactive where we can tell APS IOS, Hey, go pull something has changed. ⁓ How do we go about it? Remember Redis PubSub? Yeah. So we can use that Redis PubSub.

So all APIs have subscribed to this.

all API servers subscribe to Redis PubSub ⁓ and the new server that, sorry, I should have used different one. ⁓ And the new server that would be there, it would push an event into this Redis PubSub whenever it is done uploading. ⁓ And now all servers can receive the updates. That makes things reactive. Instead of you pulling it. So your update is happening once every week. Why would you pull 5,000 times before that?

unnecessary. Right? So we can make things reactive, but this seems too complex. ⁓ Having another infrastructure component running and I have really a pair of servers listening to this and this can we be smart about

⁓ So ⁓ I had one more thought. ⁓ that was to basically ⁓ I was initially thinking that in the initial file we can upload in the header itself we can have one key like let's say whether it's the latest one or not zero or one and how would you know the latest one you would never know it's the latest one or not. No so when we are uploading the new file we can make the old files header to be latest equal to zero.

Sneha Mehra (01:33:40)  
But again that required ⁓ the API server to pull it and see that okay. Yeah, yeah. So we had pulling the changes with reactive. Can we do better than this because this requires another infra component and persistent connections over here to tell them something changed. It I'll pull in. ⁓ Yes, Amrita.

So in the header we were keeping a version. ⁓ maybe by the indexing we hold a reference to version and we see if the version is updated. invalidate. you are still polling for that. ⁓ Like not exactly while polling well when we receive a request and. ⁓ Behid on the indexing and then we check with the. Back in when we receive a request. ⁓ But how would you know that something changed? Unless you pull it.

Okay.

Sneha Mehra (01:34:37)  
This is very interesting, right? This is this is what I want all of you to think about like can we make it simpler even more simpler than this because polling is good. I'm not against polling. doesn't why to waste like why to make 5000 requests when you can when you but someone can tell you or even someone can trigger something that's a hit. ⁓

Yeah, I would either expose a rest and point. That's like this. ⁓ You're making it react. Same thing. You expose a reset point and this server calls that. ⁓ Yeah, basically because we'll have to do one server at a time, right? Like basically we are upgrading. ⁓ if we're upgrading, what we will do is we'll do it as a, let's choose a server, ⁓ ask it, ⁓ either send an event to it.

which you can anyways do like basically, whenever you have a job, first, ⁓ let all the servers register with you, like basically all the API servers. So you have a close connectivity with them and send an event like you, you may have during upgrade, you may have to send them multiple events. ⁓ Now you send the first event wherever you're ⁓ doing with an API server. ⁓ Now, once you send that, that server is going to, so I have two ideas. Let's go with a simple idea. First, you have to

Collect the index, give it the path ⁓ and create, create that path, create the index. Why would you want to create a path and tell it to the server? Yeah, that's fine. Yeah, that's right. Yeah. You have a metadata and just read the part. But the point is you'll have to first create indexes in memory. is okay. So the APS I was would need to load it. Yeah. That's the normal stage. ⁓ Yeah. Now you will create it. You will test the API server and everything is fine. ⁓ And then you will.

basically tell your load balancer that this ⁓ like basically if you either give a new port to it so that that API server changes the port and the load balancer knows it. complex complex complex complex complex complex. ⁓ You see we are trying to change port of the server and this and that. This can be done in a very simple way. Very simple way. ⁓ I'm see I'm not against your approach. I'm not I'm just trying to make it

Sneha Mehra (01:36:56)  
utter simple like something that anyone would be able to understand. And because that's where the simplicity like that's where the beauty of system design lies. And okay, let me give hints so that we can converge a little faster. The hint is that given that we now have a static file MetaJSON, which contains the path ⁓ of the ⁓ dictionary that needs to be served. We want all of the servers to read this new Meta.JSON file. ⁓

But if you think about it, then what happens when this API server boots up? When this API server boots up, it by default reads this Meta JSON file, sees the path, goes to that path, reads the dictionary, builds the index and starts serving it. So this is literally like a boot up of this API server because that's exactly what it would do. So what you're doing is when you're updating this and updating this part, you have to trigger a reload.

of this API server.

This is the hint. Can we do something about it? Rahul? So basically what I was thinking was basically like we can upload to a same part posting. Like you cannot upload to the same part. That's the wrong thing that we, that we converge. Just a second. Just a second. Okay. We can upload to a same part and like there needs to be versioning enabled. That's specific. Right. ⁓

any back like anything that supports backup will support versioning like that's exactly what we did with creating two files separately. Yeah. That's like what I'm So I'm, I'm, I am saying this so that other don't get confused. That's right. Okay. Let's just say like there is a matter.JSON. Okay. So whenever like there is an update, we can just update matter.JSON. ⁓ And ⁓

Sneha Mehra (01:38:56)  
not only the parts we can also have like we can also check matter.json attributes like ⁓ updated ad that is still polling that is still polling ⁓ we are doing it on like any request that we are like that we are having okay but on every request made making a meta call to meta json file expensive why why would you want to do that that's still polling okay okay let's just say okay not matter.json we can

talk about like that, our dictionary file, the dictionary file that we are having, which, which is having all the entries. Okay. We can check like anyway, we are reading that file and whenever any request is coming. Okay. ⁓ Before that we can just check updated at or anything like that tag. Like, can't. I said. That's what I said. Right. Because when you are reading that you would get garbage data. Then you would make another read, reload the index and whatnot. Right.

No, like not the garbage data. I'm talking about like before, reading the file itself, before reading a file itself, we are just checking like, ⁓ we have like any, any file as like metadata, like updated ad created ad and all of those things. Where is that stored? ⁓ So ⁓ like ⁓ those in information size stored at the top of the file, think like for the text file.

for the text file. No, that are not stored there. That's what that's what. Okay. Let me convert this. This discussion will go into crazy directions. So let me convert it. Let me convert faster. Otherwise we would spend till 1130 on this. Right. But it's good that you folks are thinking why, why are we making certain decisions? Now we understand, right? The, what you're talking about this Rahul on ⁓ bursting and all. This is exactly that where you're uploading it. Different files having a meta file. Meta file contains

which one is the active one, right? That's the latest version. It really is we can abstract it out for us, right? But I'm laying out in a very simple terms. Okay. Now, if you look carefully, this is exactly what would you do when you reboot a server. Because when a server reboots, it would root the metadata JSON file, understand what path it is, go to this path, read the file, read the index, build this hashfab and start serving the request. Right? So what we're effectively trying to do

Sneha Mehra (01:41:19)  
is when we are done uploading this file to this path, updating the metadata JSON file, we just want to trigger a server reload, which is equivalent to doing a deployment. So what do you do?

You don't need this. You just include your ⁓ AWS SDK and trigger a deployment using your deployment pipeline and bot. It would do rolling upgrade. You just automatically when you trigger a deployment, when you have a rolling deployment configured, it would take down one server, spin up another machine. When this new machine spins up, it would read the method because that's what you have written when the server boots up, when the API server boots up. That's what you are doing.

You are reading this metadata file. Now it would get the new path. It would read it and submit. The problem is with inconsistency, but you're still not serving garbage. You would have a time while this deployment is happening that sometimes the request goes to one server that has already built index on that old data while other server that is new server that is the new data. But meanings does not drastically change. You are at least not serving garbage data. You are serving eventually consistent data, but that's fine.

because meanings anyway does not change drastically. Right? So you avoid surveying garbage data because you are reading from another file now and you're just triggering a rolling deployment. Any deployment strategy is fine. You're just triggering your deployment using the cloud provider that you're using and it would seamlessly take its effect. This is how you don't need anything. Just one deployment invocation call, whatever cloud provider you are using Jenkins job, whatever you are doing.

just trigger a simple deployment because when any API boot up, would read the metadata file in the metadata file. Now the path is the new path, load the file and serve it. ⁓ So instead of exposing API points, reactive redis, pubsub, all of that, throw it out of the window, trigger a simple deployment and a job is done. ⁓ And this is how, and by the way, that's exactly what I've written in this part. ⁓ This entire thing that I've done. ⁓

Sneha Mehra (01:43:30)  
This thing is exactly that. ⁓ I just been a little more elaborative on what we just covered. This is exactly that. Change log, how it would happen, merge that we just discussed, uploading it to the same path, transition, reactive pop sub, parallel setup. This is what we are doing as parallel setup. ⁓ Portability by adding header. And this is what would happen behind the scenes. That's exactly what I've covered. We went very in depth when you are brainstorming it. ⁓ But now...

Why did we discuss it? Let's first understand that. What we just ⁓ see with this is we are now able to seamlessly fire your get request ⁓ on S3, not put request, get request on S3, pointed reads, reading the bare minimum amount of data that you have to. This has huge applications. You can create multi-tiered storage because of this. So let's say you have gigantic amount of data in your database.

on which let's say historical data is not even queried. I'll give an example, Amazon orders. Amazon orders, how likely are you to access orders that you made six months back? Highly unlikely. Highly unlikely you would do that. So what if you take the six months old order and push them to S3? But this does not mean user would never do that. User can still go back in time and check, hey, what was the order that I made? So if your order is reset,

You go to your hot database, let's say my SQL. And if it is your old order, you go to S3 and read it. But how do you know? On S3, you cannot just dump everything in a JSON file. You need an ability to make pointed queries on S3. This is how you are doing it. This gives you an ability to fire pointed queries on S3. Because now literally, just replace WordDictory with OrderID. You're getting order details from S3. That simple.

And this is the beauty. This one system that we just discussed as example as simple as word dictionary on S3. It solves a ton of problem. It helps us. It lays the foundation for multi-tiered storages because we are not losing an ability to query the data that we dump on S3. In most places you think that hey when I dump on S3 it's archival. Here it's a cheap storage for us. We reduce the cost of our main database.

Sneha Mehra (01:45:53)  
by keeping the data that is very infrequently accessed on S3 by trading off the time because in hot database you can get it very fast but on S3 you would get it relatively very slow but that's okay because that's very infrequently accessed data but because you move data out of a hot database into your cold storage you are reducing the cost of your hot database which is significant same significant and this lays the foundation

Another example, Athena. Athena does exactly this. On Athena, you can load a CSV file and fire queries on this. This is exactly what Athena is doing behind the scenes. It is creating an index, keeping it handy along with that file. When you fire the request, it goes to that file, reads the records and starts it back. internally does exactly this. having an ability. So what we just built is we built an ability to query data.

Simple key value format. can make it more complex, but to start with, let's have simple to fire a simple get key request on S3 without having to read this entire thing. We just built it. It has a ton of applications around it. ⁓ AWS S3 Athena, ⁓ entire data lake. ⁓ If I've heard of data lake, data engineering folks would definitely know it, but data lakes are where you're storing huge amount of data.

in a very cheap storage, but having an ability to query Apache Hoody does exactly this behind us. The format that Hoody table format is exactly this index that we are storing. Hive partitions, exactly this. This is the foundation for having an ability to query huge amount of data on a very cheap storage. ⁓ And this is what I wanted to cover. It comes to building a vertex tree on S3 considering from deciding

⁓ to development, to deployment, seamless transitions and everything and keeping things really simple rather than over complicated. Right? Okay. Any questions on this? I'll take questions till 55 and then we go on a break till 11\. Hemant. Yeah. Maybe a few things I'm not able to make sense. I'm just asking ⁓ what we are trying to say ⁓ at the time of when we created the merged file, which is

Sneha Mehra (01:48:14)  
the current file plus the change. Now we have a file. ⁓ Now, once we create the new file, this new file has to go and create a new path in S3 and that has to be known to the API servers. That's what we are trying to achieve. Yes. Yes. Now, ⁓ if we are sending, let's say, but I understood the solution is we send a request to ⁓ S3 that, okay, this is the new file ⁓ and

we go and update in a new path, ⁓ which is being mentioned in the Meta file. ⁓ on that call itself, we are sending API servers to go and do the redeployment. ⁓ Once that file is... Correct, correct, correct. So in that new server, ⁓ Meta.json file is updated, ⁓ we are just triggering a deployment of our API servers. ⁓ Okay. So we are making sure first that Meta.json

Jason is updated. Yes. ⁓ in the deployment of my, like in the change log process, I'm making sure file is created, then meta is updated. Then only I'm sending this thing to the deployment server. Yes. Yes. Okay. ⁓ Just I'm thinking about, can we ⁓ do bit better in place of creating a new VM whole deployment? It's a redeployment ⁓ and

It is just like a process that was reloaded. ⁓ Can we create something which is invoked? ⁓ whole thing is like going, heading from meta, creating the index and start serving up the hash table from server. This is the whole thing. In place of doing the whole redeployment, can we avoid ⁓ of

⁓ Create a new VM doing the whole redeployment. Is there a ⁓ better way we can do it? ⁓ maybe I am like, no, there are, there are, can imagine three ways to do it, but it's just, just make things a little more complicated. For example, I'll give an example. So what you can do ⁓ is instead of doing redeployment of it have like what, ⁓ I think this mentioned that just have API endpoints that would trigger reload of the API servers or trigger API servers, make it reactive.

Sneha Mehra (01:50:34)  
Instead of doing redeployment, if redeployment seems big enough, ⁓ just have API endpoints experts that you invoke, which tells them to now go and pick the new file from S3. That's right. That is one way to do it. ⁓ But I'm just making things simple because this, ⁓ this is an out of the box approach. Thinking of triggering a deployment as part of an all server reload is an out of the box approach, which a lot of engineers just don't think about. I've built, so it's just that I'm proving you all.

to think in that direction that can be come up with something as simple because you already have a deployment set up the sticker diploma because it's anyway the cost remains exactly the same but not having to have pulled the metadata JSON file or do reactive or have reactive pubs of again now those sort of complexities you can totally avoid. Right? And just that I chose this to prove you in the direction that hey we can we control our own infrastructure so we can do something with it if we want. That's it. That's the only thing.

Okay. But, it can be done in a hundred percent. I'll tell you one more thing. Why should API servers read index from S3? What if in the updates that we are sending to a person was what they be sending the actual index data. It's just 1.7 MB. Right. No need to reload the problems. ⁓ Right. You can change a lot of stuff. It doesn't take one. So that I can make a point on ⁓ utilizing our infrastructure calls to get something.

Okay, got it. Thank you. clear. Yeah. ⁓ Okay. So let's folks will go to break. We'll take a break till 11\. It's 10 to already will take a break till 11\. And then we come back. We'll discuss super fast DB. Again, very similar to what we discussed on word dictionary, but we'll build a super fast DB, a very practical example. Everyone uses it unknowingly. ⁓ And we'll make order one reads, order one writes, order one deletes.

with persistence, right? And this is where we'd go a little more deeper into what the file format should be because right now we just work with simple CSVs, but that's not how this is actually stored, right? We'll go deeper into it and build another database, which is lightning fast database that gets the max performance out of your hard disk. And that's the highlight of it. And we all are using it, unknowingly, right? So we'll talk about it in depth when we come back from the break. We'll resume at 11 o'clock.

Sneha Mehra (01:53:00)  
like seven minutes from now. Thanks folks, see you folks at 11\.

Sneha Mehra (01:59:52)  
Okay. Let's start with the second half. Second part for today ⁓ is we'll now talk about building a super fast key value store on top of very commodity hardware, cheap magnetic hard disk. Right. We want best of everything. Read, write, update, delete, persistence, everything we want. So first, before we start our brainstorming, ⁓ what we, what I would want to talk about is how magnetic disk works. So first of all,

If you ⁓ were to use, we are all, we all are so used to using SSDs. We want to say this is very fast at all, but they are very costly. So if I were to put cache like your CPU cache, your L1 cache, your SSDs, your magnetic disk storage, your tape storage, right? So if this is your memory, if this is your ⁓ magnetic disk storage, this is where your SSD lies.

So SSDs are more closer to memory as compared to hard disk with respect to speed and with respect to cost. So getting an SSD is very costly as compared to getting ⁓ a magnetic disk storage. So which is where you are storing huge amount of data. You don't store it on SSDs. You store it on magnetic disk storage. ⁓ Now magnetic disk storage, what are they? It's that classic disk based storage.

where you have a head and the head rotates like this in this motion while your disk spins. ⁓ So when you issue a read of something like read of a particular byte offset that thing gets translated to a particular disk on that particular disk a particular sector in that particular sector a possible offset.

⁓ So you say on this file at this byte this exists it then translates into on this disk in this sector ⁓ at this offset it lies in. ⁓ So this is how it works. So now when your disk is moving and you are issuing a read command, let's say read from this from this sector from this byte from this particular thing. What happens is your magnetic head movement. So this moves

Sneha Mehra (02:02:19)  
to that location magnetic head comes to that point read that particular part and sends it to the device driver and then it goes to the user space and you consume it. So there is this actual physical movement which is happening for you to read this part over there. ⁓ So the head needs to be moved the disk needs to rotate you read that then some other read comes into it moves forward. So this mechanical movement that happens on the disk

is the most time consuming process. ⁓ So this is how ⁓ magnetic disk works. Now we are building a super fast key value store that works on these cheap magnetic disk storage. ⁓ So now how would you build it? ⁓ Let's start with, ⁓ given what we want to do is we want to support operations like put, get, delete, that's it.

⁓ But the key thing is we don't want to miss out on persistence. So we have to persist everything. ⁓ So how would a put look like? Hemant?

Okay, mean, whatever the requirement you have given, ⁓ it looks like ⁓ hitting in the way that ⁓ the structure of a MAM table or SS table works, you have something in memory as a sorted map key and then immediately I want to process it immediately when I say put it is to immediately flush to the desk.

⁓ Okay, then ⁓ if you have to immediately ⁓ plus it. ⁓

Sneha Mehra (02:04:06)  
Okay. So let me understand maybe we are saying that as soon as the request comes, ⁓ we will able to persisted on the persistent at the desk. ⁓ Each each protocol. ⁓ Yeah. ⁓ Yeah. Okay. I mean, then then the then the best thing is like you create a ledger writer WL where you start adding

each and every, but it has its own challenge in terms of when you have to ⁓ delete. Wait, wait, let's, let's let's by step. You're saying right ahead. What is right? Like you have a file, which, ⁓ as, ⁓ like each and every change that you are doing, ⁓ are like inserting as like appending in the, in place of going to figure it out where it has to go. You start adding at the end so that

the magnetist discard best for that. I'm thinking why, why, why, magnitude this got best for append only thing? Because you are, you, you, so when, when, when you know ⁓ on this sector where your head was last, okay, you have to add more blocks on top of it. ⁓ So you have not to look for your cylinder. I mean, you, your heads to go and seek it. So neither, neither you need a seek time.

neither you need ⁓ that your ⁓ cylinder has to spin and reach to the right sector. ⁓ both the movement of like ⁓ moving your spinning your disc. is all gone. ⁓ All is gone. I'm at the right exact point where I have to add more blocks. ⁓ That's the best I can do to get the benefit out of the characteristic of the disc. Perfect. ⁓ So ⁓

So that is, and now what I'm worried about is the cons of it when I have to delete it. No, no, I mean, first let's convert and put, what happens when you put something to this. So let's say this is that append only file. So this is the file that you opened up, right? Now let's say I got a request that says put K1 V1. What will happen? What will you do? ⁓ So put K1 V1, K1.

Sneha Mehra (02:06:33)  
⁓ is a key which is ⁓ I need to add it so I will add that key and corresponding value I will add so that much data is written in the disk ⁓ and I need to have K1 ⁓ and wherever the K1 is started that ⁓ sector address offset is in somewhere in my memory table or somewhere else so that will basically come to that

We'll come to that. Let's first get the max performance out of it. When we hit get, ⁓ First, let's focus on putting, right? Are we doing the best when we are putting? Then we come to get and then we come to delete, right? So you'll cover the entire thing around put, then I'll move to the next one. Then we'll cover get, also then we'll cover delete and then we'll cover get. Okay. So to cover the put part, you said that every time I'm issuing or I'm getting a put operation. Yes. ⁓ I will ⁓ just create new entries into this file. A2, V2.

Yes. then K3 v3. And then let's say if I get another request, let's say now one changes, someone issued a put on K1 with v1 dash. Now, okay, I'm coming to that. So that's the next thing where we have to think about because we cannot go and figure it out where K1 was and change the value. So we will append it. ⁓ We will continue.

Append it only, ⁓ we will not go and change it. We need to figure out another process which goes and market as like obsolete data or market for deletion or something like that. ⁓ What they say, process or whatever to yeah, ⁓ Tom, the stone process. have to do it. ⁓ but idea is always any put comes, ⁓ you go and add no matter what, no matter what that, ⁓ that seems the best way.

Perfect. So this is how you get a max performance out of put operation on a cheap magnetic disk storage. Right. So put operation is started. Right. Okay. ⁓ I think you had your hands raised. I was going to pull you in, but let me go to Pankaj. Pankaj. ⁓ How will you do deletes? Let's say I got a delete request and I want to delete a particular key. How will you delete? Let's say I delete K2. What will you do now? Yeah. So, ⁓

Sneha Mehra (02:08:56)  
I mean, what I was thinking is somewhere from our previous discussion again that we can go to this key and set the value to be something which will never occur. Like if all the values are always positive, we can set it to minus one. So you're going back to that key. You're going over here and then deleting this and basically setting this to zero, zero, zero, zero, some special value. Yeah. But this is up and only file. You cannot write to any other location.

because it's append-only file. You cannot go back and do something. Like you can only read from the previous stuff, but you cannot write to a previous stuff. That's what append-only file does.

We are not allowed to go back and write anything over here.

Hmm. So, okay. I mean, that case, we will have to maintain some version sort of thing saying that K2 has been deleted now. ⁓ So what can you do?

Sneha Mehra (02:10:00)  
⁓ So which means now let's think from the first principles. ⁓ We know how port works. For us to delete, it's an append only file, but we have to write something that tells that K2 is deleted. So all we can do is append. ⁓ Yes. Correct. So I mean, can append something like K2 with minus one that shows it's been deleted. Correct. Any special value. So effectively your delete operation

is equal to put operation with some special value. Yeah. Yes. Because delete was because put was order one like bare minimum, delete also becomes super efficient. ⁓ Yes. ⁓ Correct? So this is sorted. So we get the max out of output. We get the best out of our delete. These two are sorted. Right? Okay. Now the now comes the next part. Get. Right? ⁓ Jyotindra, how would you power get? Now that

in this file structure that we have, we have multiple entries of the same key. ⁓ Now for us to perform get what is the naive way of doing it? ⁓ Just start looking at all the entries and return the latest one. Return the latest one that seems very poor. ⁓ But you want the best performance out of get as well. So what will you do when your reads are slow you make? ⁓ Yeah, ⁓ okay, ⁓ so let's do that.

Sneha Mehra (02:11:32)  
So how will you make it faster then? How will the gets work? ⁓ Yeah. So ⁓ my mind like overall is still going towards what Hemant had mentioned earlier, like around having an ss table sort of structure here. Hash table, simple hash table. Right. Ss table is very different. Hash table is very different. Yeah. So ⁓ for a hash table, we can, like once we have put something into a right ahead log, we maintain a hash table.

on the like as soon as it's appended to the ⁓ well we update the hash table as well. So the offset of it, correct? The actual value. Yeah. ⁓ And ⁓ whenever a read comes, we simply go to the offset in the well ⁓ and ⁓ return that read it because in a pen only file you're allowed to do random reads, but you're not allowed to do random reads. Yeah. Okay. So ⁓ just to ⁓ cover it end to end.

What would happen is let's say I got put k1 v1 I first write it to the disk over here and note with the offset and then I update my in-memory hash table such that k1 points at o1 then I received k2 v2 I update my in-memory with k2 o2 over here then let's say k3 goes to o3 over here and then let's say I got put k1 v1 dash now what will I do? ⁓ I will first write k1 v1 dash over here

Then go to my hash table and I updated K4 over here. So it goes over here and the first entry removes over here. Now let's say K2 is deleted. Now what deletes would look like? If I delete K2, I'll not point it over here. I just remove the entry from my memory hash table. If the entry does not exist, it does not ⁓ exist. Now what if ⁓ given that we have the structure,

First, let's look at how get works. If you want to converge or if you just want to consult on how gets would work. What would be the flow you get the request then? So you get the request you look it up in the memory hash table. ⁓ If the ⁓ entry is not there, you simply return 404 or something. ⁓ If it is present, you go to the well and you go to the offset and read that value. Perfect. Right. Okay. Now what if

Sneha Mehra (02:13:53)  
your machine died and because this is in memory, it is gone. Now what? So you'll probably have to rebuild the indexes when you restart the server, which might not be a good idea for a very low. That's fine. That's fine. When you restart, you just go through this entire file and rebuild this index. Yeah. ⁓ Straightforward. So get order one, you are doing the bare minimum in get, you're doing bare minimum input, you're doing bare minimum in delete. ⁓ Yeah. So it's the max that we could get.

out of this ⁓ on a cheap magnetic disk storage. ⁓ So we just did read, write, update, ⁓ read, write, update, everything with full persistence. Right. But what else? What could go wrong over here? Siddish, we have this one file that putting these entries, what could go wrong?

Sneha Mehra (02:14:50)  
I mean availability is less because when you come back again, take a lot of time. That's fine. We are still okay with that. There are ways to make it better, but at least we have an eye approach to get started with. ⁓ Another assumption is we can hold on to all the indexes in memory. That's an assumption that ⁓ because you ⁓ now you are limited by the amount of data you can fit in them, amount of index ⁓

fit in RAM. It's not the values, but the keys. So you can add only as many keys as many you can keep in memory. Correct? So that's a limitation. Okay. But what else can go wrong? ⁓ Another one is we are just appending to a file. So there is no disk garbage that we are collecting. We'll have to also figure that out. Like eventually our system will

go out of space because we are not collecting anything from the disk. Correct. So file grows out to be huge. ⁓ What could go wrong when file grows to be huge? How do you solve it? Yeah. So basically our disk space will be ⁓ out. So how do I solve ⁓ it? ⁓ Basically, ⁓ do I do is, ⁓ idea is ⁓ don't just have one file on the disk. ⁓ So you will have to create multiple files. ⁓

And indexes we can maintain to multiple files. ⁓ Now, whenever you change, yeah, whenever you change a file, ⁓ you will have to start a process to like we did last time to scrub, like consider a file, a partition, ⁓ that we are not actively appending to, but we are just still reading. ⁓ You start reading from that file and create a new file out of that partition to remove the deleted keys and just write the keys which still are required in the system. Got it.

So at any point in time, right will only go to one file. Yes. Other file becomes immutable. ⁓ But when will you make this switch? When, when will you make, Hey, now I'm done writing through this file. Now start writing to this file. How will you do? ⁓ this depends on the disk space because how many files we want to keep in the disk, right? At max at any given time, ⁓ I think we would need a space of three files. ⁓ One that we are writing to ⁓ one that we are reading from.

Sneha Mehra (02:17:13)  
and one that we are going to create. No, reading and writing are from the same file right? ⁓ No, ⁓ reading and writing won't go to the same file. ⁓ Whichever file my key is, ⁓ it will go to that

Yeah, but what you have to do is the file that you're reading, ⁓ has a lot of offsets, right? Like which are deleted. ⁓ I won't like if our disks are serially better performance, ⁓ I would just read one file, right? To create a completely new file out of it by just picking up the keys that are still available. ⁓ that's unnecessary. That's unnecessary. Right? Just, just for that one read.

then you can just club it in a different process that we'll talk about but first let's decide what's the criteria ⁓ of making the switch to a new file what's the criteria? each disk space it will be either divided by 2 or 3 based on the approach that we so if your disk space is let's say 100 terabytes you'll have files worth of 10 terabytes each 10 files ⁓ in this naive approach yes now there are more problems to this no what's the problem with that?

Yeah, the problems are basically the way you're scanning, right? When you build up indexes one, ⁓ if you're just scanning a single file, you can't concurrently build indexes in memory. So that's one problem. So that's an motivation to create more partitions. Number two is also the garbage collection that you're doing. Right? So if in smaller partitions, you can be more consistent, like just scrub the biggest one of all the biggest one of all file system limitation. It is very much possible that the file system does not support more than one terabyte.

Yeah, quite possible. Right? Also consider that. Right? PTRFS has its limitation, NTFS has its limitation. Right? Plus waiting for that huge file to pile up, then writing to the next one versus keep smaller files. Let's say 100 MB files or let's say 1 GB files. As soon as that threshold checks off, you move forward, right? You start writing to the new file. So the idea being

Sneha Mehra (02:19:21)  
that let's this is file 001, then it is file 002, then this is file 003, this 004 and you keep on writing. Now at any given point in time, ⁓ only one file will be active. So let's say this is the active file. So at any given point in time, one file will be active where all the writes are going. ⁓ Everything else becomes read-only.

Right, so you are still allowed to read this. Now given this in the in memory structure that we had, we were just storing offset. Now we would want to store that this key is present in this file in this offset.

So now when the read comes in, ⁓ get comes in, we get the key, we find and this file and this offset is there. You go to that, read the content and send it back. ⁓ And every time your file crosses that limit, you create a new one and start writing it. So this is typically called as file rotation.

⁓ Once you are done with that, it's very evident that out of all of this file, are lot of stale entries, lot of deleted entries. So this K1 got overwritten over here, so this is stale. Why do you want to waste that much of space? This K2 was deleted over here, so we can anyway delete it. So what do you want to do is when you have multiple of these files, read only files, active and let it continue to support it. You'd run a process called Merge.

and compaction. Where what will you do ⁓ is you will take these parts, take these files, ⁓ iterate through them, ⁓ remove or don't or and basically create a new file. So it would be like you have this three file using this three file you create a fourth file which is combination of this three.

Sneha Mehra (02:21:20)  
in which you are not copying the stale entries, you are not copying the deleted entries ⁓ and once this file is written, you delete these three files forever. And then these two becomes your active file. Like this becomes an active file, this is where the read comes in. But when you do this, you have to also update your in-memory index because now the offsets would have changed. Earlier you would have used stored that this

is present or rather this key is present in this file at this offset. Now when these three files get merged into this fourth file, what you have to do is you also have to update this in memory index for the entries that are just moved to this new file. ⁓ This is a classic way to do merge and compaction. ⁓ You can very clearly see how you can do merge and compaction. ⁓ Just iterate through this file, keep keeping track of data that you are having. Okay, this is the key that I saw now. This gets replaced periodically flush it to the disk and basically do the simple merge and compaction.

When you are merging all these files into one. ⁓ So if you look carefully at your disk utilization, ⁓ this is it would look like. ⁓ Periodically, ⁓ your file, your disk usage will go high high high. ⁓ When merger compaction kicks in, ⁓ it would suddenly drop. And then it would grow high high high and then suddenly drop. It would grow and then suddenly drop. ⁓ So this pattern that you seeing because

When the rights are coming in your disk would grow like the disk utilization would grow more data you are adding when your merger compaction runs your rate would suddenly decline then it would grow back again then the next runs and then runs and then the next runs you will see this sawtooth pattern in your disk consumption now why I am talking about this because this would help you configure what needs to be there like what should be the

flush frequency or rather margin compaction frequency because if this is a disk utilization, this shows this line, this line would tell you this is your max, this line would tell you the maximum capacity of disk. So if your disk is let's say one terabyte big, you cannot let this grow beyond one terabyte. ⁓ So this pattern like how you'd want to sort to determine how your performance your disk database would be. ⁓

Sneha Mehra (02:23:43)  
⁓ Then how frequently you'd want to dump if you let's say dump too frequently, it would have too much of garbage collection happening too much of things that you're worrying about trying to frequently your CPU not getting CPU to do actual work. Right? If you do it less frequently, then a lot of work gets accumulated. ⁓ So your flush frequency that you are having should be smooth enough for you to not drop down on your performance. ⁓ These two factors you would be considering in order to configure. You only have two configurable parameters over here. First,

is margin compaction frequency and second ⁓ is disk rotation threshold sorry file rotation threshold like how big you want your file to be

Sneha Mehra (02:24:29)  
how big you want your file to be and how frequently you would want to merge it. These two are the configurable parameters. And how do you configure it? Looking at the sort of pattern that you're seeing. So, configure this when you're writing, when you're building this part, have these things in check, monitor them, and then take a call that, what should be this frequency? And this, when you're designing this database, you provide this as a configurable parameters to your folks, to whoever you're shipping. ⁓ But this is how you get the best

⁓ read, write, update, delete along with best or most efficient disk utilization not wasting memory having this merge and compaction. This is a pretty standard approach and this is the best we can do ⁓ on disk on a cheap magnetic disk storage. ⁓ On SSD you'll still get benefits but not as significant because SSDs don't go with disk based approach to or rather ⁓ the circular disk approach to write it. They have

They have flash memories and whatnot. ⁓ does very optimally there. But on a cheap magnetic disk storage, this gives you wonderful experience almost as fast as SSDs. ⁓ So on a cheap storage, you get the max performance out of the system. ⁓ But it's still not done. ⁓ I want to go into bit more granularity. ⁓ On to how exactly because here we just casually said, I'll write key value over here. ⁓ No, no, it's much more than that. ⁓

Let me go into details of it or like how will you actually represent and how should you actually represent when you are defining a disk based structure, right? That this is my key. This is my value. How do you define it? Even in dictionary problem, we just wrote word comma meaning and I'll read file line by line, but that's not how files are read files are read by blocks. So how do you actually define a format or record format?

This is how all database work. ⁓ Any database, MySQL, Postgres, MongoDB, Redis, Fedis, any database, whenever they're serializing things to disk, they have their own format in which they write it so that it becomes very efficient to read. We'll walk you through that. ⁓ So, how one entry in the file looks like. We always say that in this file, the entry is looking like a key value path. We'll store key, we'll store value. ⁓

Sneha Mehra (02:26:54)  
That's what we are thinking here. I'll just store key here and value over here. But now imagine when you are doing this, you cannot just read until new line because every time you're doing a disk IO, ⁓ let's say you read one byte at a time until you hit a slash end. How many disk IOs are you making? That would make things so slow for you. ⁓ You should exactly know if I'm at this offset, how many bytes do I have to read to complete my entire key value? ⁓

which is what you do. ⁓ You don't just say I'll read until new like that is literally 11th standard kids would write code like that. Right? Read until new like no you have to read the exact number of bytes you are supposed to so that you get that entire data in one shot. So that's what you do. You instead of just write key and value you write key size and value size as well. So first you write key size

which tells you how many bytes is key. Then you write value size which tells how many bytes is value. ⁓ When you read it, you read these eight bytes over here. These eight bytes, four integer, four byte integer. These eight bytes, you interpret the first four byte as key size, interpret second four byte as value size. Then you read these ⁓ many bytes in total, you get key and value both. Out of it, you get these many bytes that becomes key.

and these very bytes becomes value. ⁓ It gives you the must-read predictability. So that in one disk IO, you get this part, the metadata, kind of header that we discussed in the previous bullet. You read this eight bytes, interpret first four bytes as this, second four bytes as value size. Then you sum it up and you read one more disk IO to read this entire thing together.

in memory and then you process this as key and this as value. ⁓ Because from this you would know exactly how much key takes and how much value takes. ⁓ So this is what you would store on the disk. This is similar what we will do even in word dictionary also. Word size, meaning size. ⁓ Same thing, we just wrote it in CSU because it's easy but now we are to details of it. So now you can update that system design with respect to this approach. This is how every entry in your file would look like. Very predictable. ⁓

Sneha Mehra (02:29:19)  
Okay, a lot of people just stop here. But always remember your database needs to be fault tolerant. What if you're writing a gigantic value and while writing the value the machine crashed. You're just writing one one zero and then machine crashed. Now, how will you know you cannot fix it? If write as failed, your data is corrupt. How will you know that an entry that you return is corrupt?

That's what you have to know that entry that you because first you have to know that it is corrupt. Then you fix it. There is no way to fix it because the data is gone, but you cannot just serve it because after this point, let's say you wrote this entry and then let's say another entry is written. If you read value size number of bytes over here, it would penetrate into other entry that you have. And that's very bad. You need to know that whatever you wrote has lost its integrity. The only way to do it is

You put CRCs, you put checksums, cyclic redundancy checks. ⁓ And that will be the first thing that you would write to the test. Now what are CRCs? Simple hash of all the values. Key size, value size, key value, everything. You'd have key and value in memory. You create this record format in memory, compute the hash of it and store that checksum or hash or checksum, whatever you'd want to use as the first thing over here.

Similar to how in distributed ID generators you saw timestamp on the leftmost side of it in any good format of the world in good record format of the world that is persistence. You would always say CRC is the leftmost side. Postgres ⁓ commit log file CRC on the leftmost side. MySQL binlog CRC on the leftmost side. Why? You are first the first thing that you write that you ensure that is at least flushed is CRC.

because that is the bare minimum thing you need to know that something is corrupted. You cannot write CRC in the end. You write CRC in the beginning. Pick any network format. You'd always say CRC at the first. Because that's the first byte or that's the first set of bytes that you should receive because that is what would tell you that the data that you're receiving is not corrupted by any means. So this is what your final structure would look like. CRC, time-sum just for conflicts, key size, value size, key and value.

Sneha Mehra (02:31:45)  
First disk Ion you will write this much then in one disk Ion you write key and then another disk Ion you write the gigantic value that you have. So in case the first disk Ion itself fails this entry itself is not there. Problem solved. But if let's say this succeeds and while writing key it failed when you read the data you check the CRC and you see that the CRC doesn't matter because there is some corruption over here. So you tell the user that the entry is corrupted. ⁓

You're not serving corrupted data to your users ever. And then you can have your own recovery mechanism to recover from it. That's heuristics mostly. You ignore that entry in margin compaction and whatnot. You assume that the right never came your way because corrupted data you cannot recover from it anyway. This is what I want all of you to think about. you're to say, don't just take it as your CSV. Is that the best way to do it? And have a record for better. This is where you need to think.

Let's say saying that I read only until new line, ⁓ not good. Key size value size key value, all of this with CRCs, integrity checks, durability, fault tolerance. Very important when we are designing a storage HHL. ⁓ Okay. For the get, this is what we discussed. Key value, you store it in your in-memory hash table, which points to the corresponding entries in the file. ⁓ Given that we are having an up and only file, all the operations are key value pairs in this

format are getting appended one after another. Now what would happen because of this? Our one file will grow out to be very big. That's why we have to do ⁓ file rotation, which means whenever a file reaches a certain threshold, we start writing it to the new file, then a new file, then a new file, then a new file. At any given point in time, only one file would be active. Everything else would be just available for reads. This means that in the in-memory hash table,

what we were storing, we are storing just key and offset. Now we need to store that this key is present in this file. The entire record size is this as this offset ⁓ and timestamp for conflicts and meta information. Right? This way your entries can point to any file, but the writes will always happen to the most, to the active file that you have. Right? This is how your in-memory index would look like. Okay. Now we have a lot of files.

Sneha Mehra (02:34:08)  
can we optimize given that we have a lot of files having a lot of stale and deleted entries. Let's merge and compact. ⁓ So what we write, we write very simple merge and compact. We add it through all the files. They are not sorted. It's not similar to merge of merge sort. Because here the order can be anything. It's not sorted. So to load it in a temporary structure, merge them, replace the entries that you are seeing like ⁓ the further.

presence of those entries you are replacing it in the hash table and then flushing it to the disk. Simple. Try writing a code, it's not really that difficult. So read entries from this file, open it in the hash table and then flush it periodically to the disk by ⁓ not writing the stale entries or the deleted entries there. Right? But now you would see this sorted pattern emerging and this pattern would help you determine the maximum

this size or the maximum size of a file and the flush frequency that you need so that it stays within the limit. This is the max size of the disk that you have. This is what you would get. ⁓ Doing very frequent merges also, ⁓ we're doing very frequent margin compaction, not a good option. ⁓ Very infrequent, you would breach past this, right? So keep it to a, ⁓ keep it at optimal state, ⁓ that no matter what, this never breaches, right? This should never cross the cap off.

As soon as it crosses the cap of it, your system would crash. Right? Okay. Now, the limitation of this, whatever we designed sounds very fancy. He gave me order one, put order one, delete everything order one. But what is the limitation? The limitation of the system is all the keys must fit in memory because in the in-memory structure that we have, we are storing keys and basic information. So it is bounded by the number of keys I can fit in them. Values are flushed on the disk, but keys are there in memory.

So on disk there are keys and values as well, but in memory you have keys and offset. ⁓ the size of the index, ⁓ how big the number of keys that you can fit in memory determines how big your index is determines how many keys you can store ⁓ in this particular database. ⁓ But the strengths of this DB is it gives me order one reads, writes and reads. When I say order one, it's a bare minimum that your disk has to do when it is there to write, when you request it to write something.

Sneha Mehra (02:36:31)  
you get very high write throughput. You get very low latency. Write very low latency, you are doing the bare minimum disk IO. Reads ⁓ one memory lookup, one random read, get the data, send it back and move it back to continue writing. Bare minimum that you are doing that. So bare minimum to write, bare minimum to read. You get very high write throughput. You actually get IO saturation. ⁓ Your physical device ⁓ IO will be saturated and that's a good thing.

Because in most cases, it's not that you are waiting for IO to happen. You're doing the bare minimum to read and write and whatnot. So you'll saturate your network call, your saturated network, saturate your disk IO. You're getting the max out of an underlying hardware. And a very simple thing. Backups are easy. Control C, control V, these files that you have backup. It's normal file that you are just storing them. Right? You can literally do that. The database that we just designed is called BitCast.

It's an embedded database. ⁓ I previously shared the video on embedded database. What embedded database are? ⁓ This is embedded database. ⁓ Where it is not a central server that is doing it. ⁓ Right? BitCast is not like a MySQL server that was spinning up and everyone connects to that. ⁓ BitCast is an embedded database, which is kind of a library. ⁓ Right? ⁓ Like SQLite is an embedded database. ⁓ BitCast is an embedded database as well. ⁓ So BitCast is used by Uber in production for the main ride-hailing service.

For the main service, use BitCast. because BitCast is embedded database, it does not have a separate server server. So ⁓ Uber, what Uber does is Uber uses React as their main database. React is a simple key value store. And React uses BitCast as a backend. ⁓ like how Elasticsearch is an HTTP wrapper over Lucene. Similarly, React

is a wrapper over bitcast. So React's responsibility is to make bitcast distributed. You'd still have like this is mutually exclusive set of data, but React gives you this additional thing, master, slave, architecture, distributedness, consensus, whatnot. Right? It gives you all these features to correct. So it gives you proxy to seamlessly connect and route the request, hash space, consistent hashing, data movement, whatnot.

Sneha Mehra (02:39:00)  
That is responsibility of Riyadh. But in the backend, what it uses, it uses BitCast to get the maximum out of when you are storing and retrieving the data from the disk. That's the best you can do over there. This is what Uber uses in production. If you look carefully, the BitCast, the way we design, the append-only file that we see is the magic behind getting the max out of cheap commodity hardware. So wherever you see anything that says we are high throughput,

It means that using append-only file. If they are not using append-only file, they are append-only file or something similar. They are not high throughput at all. ⁓ Not at all. I devoid that claim because it's not high throughput. You can get, you can go beyond that as well. Right? So that is not very powerful. Whenever you are seeing very high throughput service, it has to be high throughput and persistence. It has to be done in append-only file. And this is where you see it in. Right. ⁓ And this is what this

very fabulous database BitCast is all about. It's really simple for everyone to learn but it covers those very intricate details, data integrity checks, record format, merge and compaction, getting the max out of cheap commodity hardware. ⁓ And it's very heavily used in production by React. Every company that uses React, most of them have BitCast as their backend. So we designed BitCast and React just makes it distributed.

Nothing much. It just like what elastic search did to Lucene react does it with BitCast. Right. And this is what I wanted to cover as part of building the super fast key value store. Right. There is a paper on BitCast, which you can refer ⁓ just a Google search away. But what I would personally recommend implement BitCast. And you know how, ⁓ how much, how big of a promoter I am of implementation.

I would highly recommend you to implement BitCast, which is not really difficult, but when you build it, you have much deeper understanding of file IOs, how to get max out of performance and whatnot. I did a personal benchmark, BitCast versus key value store, BitCast versus on-desk key value store, in-memory key value store and try to get a performance on magnetic disk, SSDs and whatnot. That's why I'm so sure on how we could get the max performance of bit, how IOS saturates and all of that.

Sneha Mehra (02:41:24)  
Set it up, it's not difficult. You just need the Prometheus on a machine continuously monitoring it. Run your BitCars, key value, which is why I just focus on key value use case. Nothing fancy on that. Right? Just key value use case, but build it. Spend two weeks, two weeks per day, one week of your time. Implement it. You'll have a time of your life. Trust me on that. When I first read it, I got my feeling was all over the room. Six, six years it has been. Yeah, six years. Just yeah. No, five years it has been.

and I implemented BitCast. ⁓ Do that, do that. That's where the fun lies. ⁓ Okay, that's all what I wanted to cover. Folks, if you want to drop off, you can drop off. I'll upload the recording by evening. ⁓ yeah, ⁓ next week is all about high throughput. Next week is all about high throughput systems. We'll extend the storage week. ⁓ We'll cover S3, video encoding, YouTube view scouting and whatnot. And one more database we'll cover. But...

Those will form the core of how high throughput systems are built and why are we doing some things in certain way. ⁓ This is where we got high throughput on storage side. There we'll get high throughput on compute side. But before that, ⁓ next week Saturday we'll first discuss huge discussion on S3. We'll go into those nitty gritties of S3 and how to think about it. Everything that we studied will come in handy when we design S3. What we design in S3 will come in handy in the next one. So it's all stringed together. ⁓ So see you folks next week.

⁓ Any questions on BitCast? Before I take other questions.

Sneha Mehra (02:43:00)  
No? Do you to the beach?

Sneha Mehra (02:43:05)  
Right, was that simple. Super. Go on, Hemal. ⁓ Just ⁓ want to understand the in general, ⁓ what they say LSM ⁓ trees and bitcats is like ⁓ completely this process only or something ⁓ extra is there when people talk about saying SS table, MAM table. It looks like whatever we have built on bitcats. ⁓

Similar to that. ⁓ Is there some? In SS table in memory, we will be touching about SS table next week. But in SS table, you also store values in memory. Here we are not storing values in memory. That's the one difference. That's the big difference. And the sorting of those values. In SS table, we sort it because we leveraging something else. ⁓ We will go in depth of SS table next week. And why are we doing something that we are doing? But beyond that, that's a big difference. ⁓ Other than that, it looks same.

other beyond that, because we have that flow wise, ⁓ you have a file where we have the content and that is used by our merging and compaction to create. ⁓ similarly for deletion, ⁓ we go and do the same thing. ⁓ So beyond that, whatever they're in the MAM table as a sorted key values, ⁓ that's the only difference. ⁓ That's the only difference. ⁓ But with that, but here we get persistence there, there is still loss of data. ⁓ If our memory goes down.

There is still a lot of data. ⁓ Here we are seeing immediate persistence on the disk. ⁓ That's big difference. ⁓ Memtable has to be saved maybe a local on the system or something. ⁓ But every time, on every write. Every time. ⁓ That's too costly. ⁓ Here it's exactly. Here you get it by default. ⁓ Because you are prioritizing persistence. ⁓ So you are getting the max out of your underlying disk by prioritizing persistence. In Memtable you buffer and then you write.

I think then it makes sense if we understand this and then go towards that. That becomes simpler to understand. ⁓ So which is why this particular order I have taken. Next week we will be touching upon SS table, mem table, that part, LSM trees, ⁓ Hyderabad systems when you do counting and all. will touch upon that. ⁓ But this becomes the foundation for that. ⁓ I got about BitCast in the storage chapter from Martin Kleppman book and I think today we

Sneha Mehra (02:45:32)  
most of the things we covered, but the ⁓ understood because they, he tried to cover the bit cask and it was the first thing which I understood. ⁓ it costs, then ⁓ it is easy to understand other things. ⁓ And ⁓ he covered that ⁓ with cask only in his ⁓ book. I think that's why I had this. had this in my college paper reading. I took a course on storage evidence and it was in, it was part of my paper reading thing.

And that's where I stumbled over like, my God, this is such a simple thing, but it, but it covers it so beautifully. Okay. Then ⁓ I understood the bit because maybe that's good. But I do, read the paper in case you have it. That's a fabulous paper. I read it a little bit, but not ⁓ completed. ⁓ Some part I, but reading from what the Martin book give me that summary. ⁓ I skipped that.

⁓ That's what happens. ⁓ yeah, thanks. ⁓ Yeah, ⁓ I bet like my question may sound naive to you, but like on the append only files, like generally, what's the way of creating those files? Like I've never done this. ⁓ So that's why my knowledge is when you say what like creating a file.

⁓ Yes. Let's say you want to create a 1GB append only file. So like, ⁓ would you kind of create a file and upfront grab 1GB space in that? And so that that file is maybe contiguous on the file system. later on, like you keep on ⁓ checking that how much data out of that 1GB you have written. And then you, once the 1GB has completed, then you basically switch the new file.

Or is it more incremental in nature that you would grab? ⁓ Like you would ⁓ keep on doing just say simple file APIs, like right to a file these many bytes. ⁓ Okay. So this is what is abstracted by the file system. ⁓ What you do when you open a file, you can specify file options, read like R, W, R plus W plus ⁓ there is an option.

Sneha Mehra (02:47:54)  
called a, a is an append only option. Okay. So when you do that, the underlying file system, so when you do file IO, what happens? The file system system call gets involved, right? ⁓ That takes care of structuring this file, moving it here and then broken it into pieces in case that's how it is implemented. You don't have to add on upfront reserve one GB of file. You just open a file and start writing.

without thinking, without worrying about it. That's what is abstracted by the file system. Right? Now it does not mean that they would have done it very poorly. In most cases, when a file is allocated, it's allocated like ⁓ a huge chunk is marked as blocked ⁓ and it's responsibility of the file system, BTRFS, NTFS, CXT 2, 3, 4, whichever you are using, that when you're creating new files, it finds a decent enough free space for it to be there.

Right. So it's not a sequential allocation of multiple files, but it just spaced out a bit. This way, even if a file grows out slightly larger, there is no, there's no problem. Right. So when you are, when you want to create an append only file, when you write, when you open a file, just pass in an option A for append only mode. And you have now just created an append only file. Okay. Right. Thanks. Yeah. Amrita.

Sir, my question is regarding the delete deletion of the keys. So I understand that we are putting some default value for deletion, but then how we handling it in the merge? ⁓ In merge and compact, ⁓ when you merging the two immutable files, you skip the deleted entries. ⁓ That is the file that you writing. So let's say you are having three files, one active and two read-only. ⁓ Those two read-only you will merge. ⁓ While you merging and creating a third file,

you skip the stale entries and the deleted entries. So that would never be copied on the disk. ⁓ basically for the entries which have for which we have that default value, will avoid the first entry and the ⁓ any entry will avoid, right? Because as soon as you detected, you would not just store that anyway. Okay. Right. Because see you are having three files in total. One is active file and all the rights are happening. ⁓ having two historical files which were rotated.

Sneha Mehra (02:50:20)  
Right now you will be merging and comparing these two files. Right? ⁓ When you merging these two files, you would skip the stale entries, which means let's say have written key k1 twice, I'll skip the first entry. So I'll append, ⁓ when I'm writing to the new file, what I'm doing is I'm first buffering the key value pair in memory ⁓ and I'm just replacing whatever I find it there. If I get something new out of it, sorry, if I get a key which is deleted,

I would just remove the entry from the in-memory hash table and then I flush this entire hash table as this new file and so on and so forth. This hash table is different from the index that we spoke about. Index is separate. This is a separate process that you are running, which is all about merging this one. ⁓ Yeah. ⁓ So ⁓ that's what I'm highly recommending to write this very small bit cascabri repetition. Very nice. Right.

You'll get like how then then you'll start sampling upon this like, Hey, how would you merge? ⁓ How do you merge this to file in the same process in different process? How will you obtain memory has to be in memory index that you have? Like all those nitty gritties, not those, those, those devils that lies in the details. You would explode. ⁓ And it's not, to be honest, it's not really difficult. ⁓ It's just start writing code one week of time, ⁓ spare that one week of time, ⁓ but build this thing. ⁓

In case any difficulty drop me a note I'll more than happy to help.

⁓ Siddish. So I have an implementation specific question. Can you go back a slide? ⁓ am wondering on the checksum thing. ⁓ The checksum that we added on the value. ⁓ So ⁓ like what we are saying here is writing the whole value is not atomic, but writing the checksum and the header is. Yes.

Sneha Mehra (02:52:19)  
this thing, writing this is atomic. So yeah, my question, my question is coming from, ⁓ like one is what, so different file systems and kernels now will give you different guarantees, right? I just want to understand a pattern here. So is the pattern that like, let us suppose we are assuming that this is under the block size. So it is guaranteed by the kernel to us. So we write that first and then write the rest of the things later.

because this value can be one MB big. Now flushing one MB in one shot, ⁓ your kernel might not give that guarantee. ⁓ Right? That's right. ⁓ Yeah. That's what I was trying to understand because now when you try to install this software and different kernels, ⁓ behaviors may be different, but we have to be and that's how this whole wall thing goes. Right? So, ⁓ But in any case, if you look carefully, this is eight bytes, ⁓ four bytes, four bytes, four bytes, ⁓ which means roughly four, four, five, 20 bytes. ⁓

20 parts no matter. ⁓ It will definitely be there. ⁓ Yeah. Got it. Got it. Yeah. Because see, ⁓ like generally there are also more paranoid systems where like kernels don't guarantee this. ⁓ So you do have a separate wall system, but I get the design. That's why I was trying to understand the pattern here. Now this seems to be very interesting pattern because it can be used because this is more performant. ⁓ Yeah. ⁓ But this is, this is exactly where these

This interesting integrity is coming, right? Like we just tend to say, will write it in a wall file, but how, right? ⁓ This is what makes us better every day because a lot of people just skip this. No one in almost any system design you pick up on any system, any overview, that's why they are called overview. No one talks about integrity at all. No one talks about security and integrity. These are like lone wolves out there. Just someone, ⁓ one passionate engineer thinks about integrity. ⁓

Otherwise it's all gone for a toss. We assume it would work just fine, but in reality it does not. But when you go into databases details, would see all this. ⁓ And by the way folks, Hussain Nasser is on fire. He is posting Postgres internals like anything. Like going into record level format and what not. My God, I don't know what happened to him. He is just going through the Postgres source code. So go through his medium logs. They are very fascinating.

Sneha Mehra (02:54:43)  
and understand the things that we just discussed in the digital. He touched upon that that this is a Postgres record format. This is how pages there he found that one very interesting comment and that blew my mind. ⁓ number of pages a Postgres takes up to store meta information versus actual data. That just blew my mind. So read that he's on fire for some reason. He's very obsessed with Postgres now.

Diving really deep into it. Every day he's putting out videos and medium blogs and whatnot. Not sure he came up on job or what. He's on fire. So once you have a bit of understanding, like implementing BitCast, once you have this bit of understanding of how FileIO works, highly recommend you to check out Hussain Nasir's. He's not created videos on it, but on medium blog post, ⁓ on medium, he has posted a ton of amazing stuff on Postgres Internets. ⁓ But very into

implementation detail, like actual source code snippet. This is where this is happening. This is how the page is getting created. The artist is getting flushed. This is what the format is. Those part, right? It's really deep into Postgres. So highly recommend you to check that out. Once you have this understanding, if you have this understanding already, watch that or read that fascinating stuff is doing then. ⁓ Thanks for that. So one more question. ⁓ Can we go back to the initial design of the S3 thing? So

⁓ There is one question that I have. ⁓ Maybe next time. I'm not aware of that. Yeah. So this S3 thing, right? So one is ⁓ the calculation part. I am still not able to guess how come even if you get 200 key words, right? ⁓ This one terabyte to me like sound. ⁓ Yeah, because I was checking per entry. is like a MBs of record. I was blown by that. So I don't know what's going on there, but

Yeah, ⁓ you're great. Like you should ask this critical question. So I just added one terabyte. But now, now, now I'll I'll just justify the size. ⁓ Right. ⁓ Okay. Yeah. So this one terabyte I added so that it becomes impossible. Like, so that I bring up this point of cost that you cannot just copy this dictionary on every API server. Right. ⁓ That's why I wanted to add it. Otherwise, you'd say, no, it's so small. I can just copy it over all the, that's what I was going to say. ⁓

Sneha Mehra (02:57:09)  
I wanted to do that so that I touch upon network attached storage, then I touch upon this, I touch upon portability, right? But now in order to justify it, now this meaning, assume it's multimedia baked in, ⁓ audio video visualization of it, baked in this thing, right? That's ⁓ For me, it's just, ⁓ I just wanted to ⁓ enforce that part that you cannot possibly copy it on every API server. So that's why multi, right?

No, no. Yeah. Thanks. No, no, just ⁓ like wondering what happened then, because I was thinking I'll just download and try this out. And that's what I was thinking. man, like this, ⁓ but the records, maybe I'm missing something, ⁓ but the question that I have, so the way what we have designed here, let us suppose I actually want to expose it. Okay. ⁓ Now per API, I am doing one S3 call. So how costly is it? That is what I wanted to understand from you, ⁓ like a point call.

And I understand I can add caching ⁓ and then my, all of those stuff will come down. But then when I was thinking alternatively, I was trying to think that, okay, I can have some kind of frequency. then eventually what will happen is I will store my frequency based words and all of that. But I just wanted to understand is that over optimization, if I'm hosting a real service or is like how, how do you have to do that? This is where I'm like directly querying the estate to get it.

So my intention for that was I wanted to point out that it is possible to make pointed reads on S3. This way you can build a of others like AWS S3 Athena, will build a multi-tiered storage cost efficient auto solution for Amazon next week. ⁓ All of that would fall into place there. Because we are able to now literally do a pointed read on S3 of a particular order ID. You can now move your partial cold storage, like this is a classic cold storage mechanism.

We moving data from hot year to cold year to save cost. But when you move to cold year, you're not just archiving, but still giving a seamless ability to query the data on S3. This enables us to do that. ⁓ Given it is infrequent, it's okay to have higher latency. Cost would be there. S3 is not cheap. Network IO is not cheap. S3 will take its own sweet time. It's not very performant as well. But it is very cheap. So if you're having infrequent re-use,

Sneha Mehra (02:59:30)  
you can make it an essay. use word dictionary as an example because it's easy to understand. ⁓ If I say key values, will be filled with key values. I just use word dictionary so that I can pinpoint this. We can actually measure how small the indexes can be and not get overwhelmed by this number. ⁓ So 170,000 seems a huge word, ⁓ like a huge set of words. But in reality, ⁓ when you multiply it by your index size, like the per index entry turns out to be MB. ⁓

We take MBs very lightly, but stands for a million, a million bytes. ⁓ So we just take million and billion for like, we'll take it very lightly. They are not very lightweight. They still contain a huge amount of information. Those were my motivations to do it so that I set it up for other parts, which should become very easy to understand. And last question on the, like ⁓ the same ⁓ part where we said that we will just trigger redeployment.

I just want to understand one detail there. So when you say redeployment, ⁓ basically you are spinning up new machines. The load balancer will be intact, right? You'll just add EPS servers. ⁓ Okay. That's By the way, Cloud Router gives you this out of the box. Rolling deployment out of the This is a ⁓ super one. I think this is coming from more experience because ⁓ as engineers, we would write some code, but there is already a way without even writing a single line of code.

So this is what I actually did for a couple of my services. I like why to have like, thought of adding radius, but what not and have a test expose and what not. ⁓ are so accustomed to write code to solve a problem. You never think out of the box. I wanted to just cover this part that infrastructure is ours. We can do whatever we want with it. So just make a call. This was a ⁓ great idea. mean, like

without writing a code and then you don't have all those have a pop then maintaining and all that stuff. You just use what you have done. Thanks, ⁓ Arif. Superb. Anyone else? Any other question?

Sneha Mehra (03:01:33)  
No, then no pop-up allowed. I have one. Sorry. ⁓ Just a basic one. So we are storing the offsets. So offsets are the line numbers or the actual byte. Byte, byte. ⁓ From where that record is starting that offset. ⁓

Okay, sure. ⁓ And ⁓ what was the other question? I actually forgot. One was offset. Sorry, can you please go to that page we were discussing about the offset and the ⁓ bitcast part of the dictionary. ⁓ The word dictionary, Okay, offset. We're discussing offsets over here. Yeah.

Yeah. So here ⁓ the index where you're mentioning the index. So initially you are storing the header where, know, by, by offset of the data, basically the bytes or the versions or all that information. So the size of the index is basically the complete index file, which we're storing in this file. ⁓ Correct. the, one. Yes. The ⁓ size of the, because initially my understanding was it is just the ⁓ location of that file stored.

and not the actual file store. ⁓ This is the actual size of the number of bytes that you are storing in the index that gets stored there. ⁓ Right. Right. ⁓ Yeah. you first read the header, ⁓ you get the index size, ⁓ then you read these many bytes here, you get the index, then you start serving this data. But now whenever, ⁓ so let's say whenever we are doing any updates, so this size might be, you know, changing every time. So every time we are updating this header, right?

Yeah, we are recreating this header. ⁓ we are doing this margin compaction. We're doing this margin compaction. We're creating this dictionary. We're creating index also. We're creating header also. This new file is a new file that you are ⁓ creating. Right. Now the question is because initially we have never discussed this thing that you know, this file has to be sorted only when we started discussion discussion about the merger. Then only we discussed that, you know, it has to be sorted. So what is the idea of sorting there because we never discussed

Sneha Mehra (03:03:47)  
how we sort from which we making a key value store. ⁓ Yeah, we are not storing it. ⁓ We did not touch upon that because we did not have a need. ⁓ When we went into intricate details, we realized that keeping it sorted makes our merging very easy. ⁓ Like merging change lock into actual dictionary, ⁓ the original dictionary becomes very easy. Very efficient. ⁓ But the thing is that now

Keeping the file as sorted in act. mean, keeping this file sorted will would create another level of complexity. So how would we approach that? is that complex? Because he's storing 170,000 records and now you're trying to insert into every time you have to do what, what, what search algorithm would you use here? A binary search or a linear search. don't have to search anything, right? It's direct lookup. You have index index as offset to directly go to the software. that light. ⁓ Right.

But inserting a new entry, ⁓ how would you decide that? Where do I need to insert? ⁓ So let's say you want to do it right. ⁓ Here it is. ⁓

This is what, you're creating this new dictionary, right? ⁓ Change. ⁓ Okay. Okay. Got it. Got it. Got it. Okay. So this idea only will keep on, keep, keep this thing sorted every time. Correct. Correct. This is sorted. ⁓ This is sorted. This This is sorted. ⁓ Right. Right. And got it. Got it. Got it. I thought you were pellets. So I have changed a of the way she did a sign. ⁓ So we are just ensuring that original dictionary started change lock is sorted. ⁓ The new one would be sorted. Now this would become the old for the next updates. ⁓ Next. ⁓ Right. ⁓ Always be sorted.

But because of the sortedness of the words that we have merging of dictionary in change log becomes order n operation. That's it. Just one iteration over this I or both the files and you have this new file created. You don't have to do n square complexity. ⁓ Actually, I started looking into this complete record, but I forgot that from the very first record, we are sorting this thing. So yeah, so yeah, that's good. Yeah, thank you. the details I want to thank for asking this question. These are the things that you know,

Sneha Mehra (03:05:51)  
brings that clarity like sparks that interests alike. This is the optimization that we were talking about like why things needs to be sorted and which is why the entire domain of computer science is obsessed with sorted data. This is why it just makes life so simple. makes processes so simple. ⁓ Thanks. Done. ⁓ Okay. Any other question anyone? No? Done? That's it? No more what?

⁓ Thanks for tuning in, I will upload the recordings by the of the day. See you folks next week. I will also share the preview. ⁓ Thanks folks, see you next week. Bye bye.

—---------------------------

—----------------------------------------

—---------------------

