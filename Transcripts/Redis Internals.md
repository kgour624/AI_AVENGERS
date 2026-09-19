11

Sneha Mehra (00:00:00)  
So in the previous video, we implemented delete, expire and saw how auto-deletion would happen or other auto-expiration would happen. ⁓ And in this one, we would look at eviction. So why does a database like Redis needs eviction? Because all the keys that you put in Redis ⁓ are stored in memory. ⁓ Because RAM is limited, what we would want to ensure is that when we hit a certain limit, then we cannot

because we cannot go beyond that because our RAM would be full. And now let's say, ⁓ if you insert a large amount of keys ⁓ and if there is not enough RAM to allocate, what would happen? Your process would crash. ⁓ And you don't want your database to crash because of OOM, which is out of memory error. ⁓ So what do you do? That's where you do eviction. So for example, hypothetically, let's say we have a limit of 1 GB.

⁓ And if we have exhausted that 1GP limit and if we are trying to set a new key, we should do something with that. So which is where we have different eviction strategies, ⁓ which is what we would be looking at this one. ⁓ So how does ⁓ Redis evict? So first, now we know ⁓ that why do we need eviction? And now in this, we'll see how Redis evicts. So Redis has a configuration parameter called MaxMemory.

in which it says that, you can configure it in the main redis.conf file, you can configure it saying that, hey, my max memory consumption should be 1 GB. So it would not let it go beyond that. ⁓ And when redis hits that particular limit, it evicts some of the old data and there could be multiple strategies to evict. ⁓ And depending on the configured strategy that you have, it would do the eviction. So depending on which strategy you put in, it would...

try to find the best key that it can evict and it would throw it out to make a space for the new one. ⁓ So in this one, we would first take a look at all possible ⁓ ways that Redis supports eviction of keys. And then we would be implementing an extremely simple out of the box, simple first eviction strategy. It's not part of Redis, but just to understand it and how we would place our code in our own Redis implementation, just as a quick example of that.

Sneha Mehra (00:02:18)  
we would implement a very simple first eviction strategy. ⁓ So first let's take a look at different eviction strategies and different performance optimizations that comes with it. First strategy is called no eviction. No eviction implies that when your Redis hits the max memory limit, we would not evict. So after hitting the max memory limit, if we try to put something in Redis, that new value that is being written will not be written.

So you'd be discarding the incoming rights that are happening. That is your no eviction strategy or no eviction policy. ⁓ Second one is all keys LRU. LRU is a very famous least recently used strategy that you have studied that you might have heard of or you would have studied in an operating system course. It's a very common cash eviction strategy. ⁓ So what LRU says is that I would evict the key that is least recently used. So it goes by the saying that if my key is not recently used,

it should be safe for me to evict that. So if a key is not accessed in let's say last three hours, there are very less chances of that to be accessed again. If that is what your behavior is going to be, you can configure all keys in RU. So what it would do is it would have to maintain some way to find when a key was recently used, and then it would be evicting it. It would be evicting the one that is least recently used. sorry, the third one,

is all keys LFU. LFU stands for least frequently used. So along with each of the key that you have put in, it would maintain a frequency count that how many times was this key accessed. And when it comes to eviction, I would evict the key that was least frequently used because if it is least frequently used, so you would be like this strategy would come in handy when you would want to evict a key that is least frequently used, which means if a key is very

less frequently used, is highly unlikely for me to access it again. That is when you would adopt a least frequently used. Another strategy is volatile LRU. It's same as LRU, but instead of considering all keys, volatile LRU considers keys having some expiration set, which means it would keep the keys having no expiration set as is. So those keys would not be evicted. There are so many use cases for that. Let's say if you want to put a key

Sneha Mehra (00:04:39)  
If you want to put a key permanently in your redis and you don't want it to be evicted even if your cache is full, this is where it comes in handy. Because all keys LRU would be evicting the key irrespective of expiration is set or not. But volatile LRU will be deleting volatile keys which means keys on which some expiration is set and it would find the least recently used out of it and would be evicting it. Then comes volatile LFU similar to LFU but instead of considering all the keys we are considering keys

having some expiration set, right? Then the, my personal favorite is all keys random. Now this random eviction strategy is very interesting. So it would pick some keys at random and evict it. No LRU, no LFU, nothing. So what it does is, what it says is that, let me just, like if my cache is full, let me just evict one key at random. And you'd see, why, why would you want to do that? Like you might be evicting a key that might just be used.

Yeah, it's true, but you don't need additional data structures to manage which is the least recently used or least frequently used. No extra data structure overhead. Just picking one key at random and evicting it. You would be using this strategy when you have uniform access across all the keys. If you know that your access is going to be uniform across all the keys, ⁓ keys ⁓ random or volatile random works really well. By the way, volatile random again, all key random, but with the keys.

that has some expiration set, right? And the final one is volatile TTL, which means that it would pick a key that has the shortest time to leave, which means the key that is about to be expired, it would be evicted. Very interesting strategies that you would see with Redis. Now let's go deeper. Let's go deeper on how Redis actually implements LRU. So that is where, because if you think about it,

⁓ implementing LRU you would have studied in a data structures course like you would require a doubly linked list and all. But Redis says, I don't want to have extra memory overhead. So what Redis does is Redis does approximated LRU and it was introduced in Redis version 3.0. What approximated LRU does is that it is not an exact algorithm. It means that it is approximated, which means that it would not evict the best possible candidate, but it would do a decent job at it. So the core idea is

Sneha Mehra (00:07:02)  
Sample. ⁓ Sampling works the best. Here you see when it comes to approximation, we are relying on sampling. So you sample some of the keys and from them you pick the one that is least recently used. ⁓ Right? Maybe with each key you are storing, hey, when was it last accessed? And when you are sampling a bunch of keys to be evicted, out of that, the one that is least recently used. That's a core idea. You don't need to maintain additional data structure. You are just sampling a bunch of keys and applying LRU on top of

such a simple, elegant, very memory efficient implementation. ⁓ But why Redis does not use exact LRU? Because just imagine the amount of extra memory that it would require to maintain the data structure. Just imagine maintaining a doubly linked list to maintain accesses across which keys was accessed when and then evicting from one. So much of memory overhead. Redis would choose to use that extra memory to store the data rather than managing pointers. That's why

caches to be extremely efficient, they don't rely on additional data structures at all. They rely on sampling, they rely on approximation. ⁓ So in approximated LRU, what we do is instead of taking just one sample of K keys, we pick N samples having K keys. Now both of them are configurable. ⁓ So if we take N samples, each having K keys and out of that, we pick the one that is least recently used. So this way, if you were just picking one sample,

of let's say five or 10 keys, right? And then you would be doing least recently used out of it. You might still evict the best, like you might still not evict the best one. But if you're taking multiple samples like this, you're increasing your chances to be as close to actual LRU, the best case LRU as possible. So this is the power of sampling. ⁓ these configurations are, sorry, these parameters are configurable. So you can check the redis.conf file to see these parameters in action.

But these are such interesting, amazing, cost efficient, space efficient algorithms that you would love to explore. Right? Okay. Then if we would configure LFU, like we saw a volatile LFU and all keys LFU, right? How it is implemented. So LFU mode, the new LFU mode was introduced in Redis 4.0. And this is again very interesting part, very interesting algorithm. So you say, ⁓ what's so interesting in managing frequency? ⁓ I'll just have a key and I'll just set up.

Sneha Mehra (00:09:28)  
Frequency variable to that and I'll do plus plus every time it is accessed. But think about it every single key in Redis for that you are having an integer and that integer you can store a million value in that right so up to a million range like 0 to 1 million It would require 4 bytes imagine 4 bytes across all the keys. You are wasting so much of space. Can you do it better? This is where ⁓ This is where this approximation comes in

So what it does is what it does is it uses something called as a Morris counter. It does not use normal int++ because if it would be using int++ it means for every key it would require additional 4 bytes to just store this frequency. Now the idea is can you because redis has to be space efficient can you approximate your counting. So for example if I want to represent 1 million

If I want to represent 1 million, why do I require four, ⁓ basically why do I require four byte integer? Can I not do it with a shorter integer? ⁓ Can I not be extremely space efficient when I'm doing that? This is where Maurice counter comes in. It's an extremely simple, but mind blowing implementation of an approximate counter. So this is where your advanced data structures part comes in.

And Morris Counter is a classic, classic, classic example of approximate data structure. I have written a very exhaustive block, very exhaustive block on internals of Morris Counter back in 2020\. I would highly encourage you to read that. It's presented at arpithbhany.me slash blog slash Morris Counter. Read that in case you are interested in how probabilistic data structures are built. So for you to store a million value,

1 million as a value, 1 million as an integer, you'll not require four bytes. You can do it with two or three. And that's the magic behind it. Every byte saved with Redis, you can use it to store additional data. Because just imagine the overhead of storing four byte integer with every key. In most cases, you might be just storing pennies, like two value, three, four, 10 value. Can you not be extremely space efficient there? This is where Morris counter comes in. Right? And just to cover one part,

Sneha Mehra (00:11:42)  
Where would you use LFU? The key idea behind using LFU, you would use LFU when even if you see a dip in the axis of a particular key, let's say this is an axis pattern of a key. It is, and this is the height of it shows the frequency of axis. Even if it is not frequently accessed, but you still want to hold it, that is where you would be using LFU. So this is like your stock market, even if a stock is down because you believe in it that this sometime in the future, it would give me the return.

That's why you're still holding it. That same analogy you can apply over here. So in case you have a use case where you are willing to take that bet that even if this key, because it was frequently used earlier, even if right now there is not much usage on it, I would still want to hold it. That's where you would use LFU. ⁓ And what Morris, although we studied like, hey, let's Morris counter we are using to do approximate counting, but ⁓ what if your key is never accessed again?

Would you still want to keep it in memory? No. So that is where what do you do ⁓ is your frequency has a cap, let's say a million. It caps at a million, your counter caps at a million. So at max it can go to one million. And then every minute there would be decay with it. There would be a logarithmic decay. You can find details of it. It's simple logarithmic decay function that you would find it in the documentation. ⁓ A lot of math involved there. That's why not covering it. But the idea is this decay is that every minute

the value of this counter would be decaying so that if even if we are having ⁓ if we see large frequent axis of that particular key, but if we are not seeing frequent axis ⁓ over a large duration, the value would be reducing to zero. Eventually it would be reducing to zero making it eligible for eviction. So that is where your counter saturates at million and then decay and the value is decayed every one minute. This is the default configuration. Obviously you can configure it in the redis.conf file.

⁓ This is the theory behind different eviction strategies to very critical approximated algorithmic implementation of LRU and LFU. ⁓ Again, highly encouraging you to see the internal working of Maurice Counter. Back in 2020, I have already written a very detailed blog on it. I went through the research paper, understood it in depth and with very illustrations, mathematical explanation, everything is there on my website. Do check that out. Now that we have this part.

Sneha Mehra (00:14:10)  
Let's say we implement something extremely simple just for us to understand where this code would be placed in our repository. ⁓ So here I have our classic implementation like always. Now here what I'm doing is I'm creating, I'm writing ⁓ an extremely simple key eviction algorithm called simple first. ⁓ What we would do is whenever our cache is full, we would be evicting the first key that we find.

Right? Because we just are creating this low level structure because implementing LRU and LFU and all of that with exactly how it is does is too hectic. ⁓ Not a lot of you might be interested into that. That's why we're just keeping it simple to just see where would we place our code. You can make it as complex as you want. Implement LRU, LFU, approximated LRU, approximated LFU and whatnot. Right? But we are just keeping it simple before now implement a simple first eviction strategy. Right? So what we would do

is you would start with the store file ⁓ and in the store file the first thing that we would do is first of all ⁓ when we would want to first of all we need to decide that when we would be triggering an eviction right so when we would be triggering an eviction comes as a fact that when I'm hitting a limit what kind of limit redis does max memory right for us to estimate how much memory we are requiring it's a little hectic so can we do something simpler

Let's say I put a limit on the number of keys that we have. ⁓ Right? ⁓ So let's say I take in as a config that at max my cache has, my cache could hold at max five keys, not more than that. ⁓ Right? So let's say in this key limit, I'm storing key limit to be five. ⁓ So this is just a demonstration purpose. ⁓ Right? So let's say if I'm doing this and I'm storing my key limit as five, which means at max my cache would hold five. You can change it to 5,000, five million, ⁓ whatever you'd want to like. ⁓ You can also change it to memory.

⁓ But that's where you would have to keep track of every time you're allocating the memory, what is the size of the object, and then you would be keeping track of it. ⁓ But just to even establish a low level code, a simple low level code, we can just write a simple first eviction thing, which works on maximum keys that we can store. ⁓ So if len of store, so while putting a key in the hash table that we have, if the len of store is greater than equal to keys limit, we evict one. ⁓ So this is where we are triggering our eviction.

Sneha Mehra (00:16:38)  
And then we are storing the key or storing the newly key and object mapping in our hash table. Right? This is where we are triggering it. And what is written in evict? In evict we are returning or we are invoking the evict first strategy. ⁓ As a future extension, we can support multiple eviction strategy as we just saw, LRU, LFU, volatile, LRU, volatile, LFU, but no need of implementing it. We are just adding them as placeholder right now implementing the simple first eviction strategy. Right? So what we are invoking is evict first.

What evict ⁓ first would do is that it is just eye treating through my store, deleting the first key that it is finding and returning it. Simple. So we are just eye treating over the hash table the first key that we find, we triggering a delete of it and returning it. This way whenever we are hitting a limit, we are simply evicting it. And once eviction is done, I have one place free and I putting my new key there.

⁓ A very simple ⁓ eviction first, like first key that we iterate, we are evicting it. ⁓ Extremely simple implementation. Let's see this in action, just to ensure that everything is working just fine. ⁓ So here, ⁓ I don't think we need normal Redis server for this because we cannot mimic eviction at a large scale. So what I'm doing is, I'm just running my Redis server, or our Redis implementation per se, which is running on port 7379\.

connecting it over ⁓ through our classic Redis CLI. ⁓ Now, I have it I am doing a set k 1 v 1 then I am doing a set k 2 v 1 k 3 v 1 k 4 sorry k 5 v 1 I have added and k 4 v 1\. I have k 1 k 2 k 3 k 4 k 5 on the top you can see the total number of keys that you have are 5 here.

⁓ And now what would happen if I add k6 to that? If I add k6 to that, the total keys are still 5 because it triggered eviction. Because it triggered eviction in the end, my number of keys are capped over there. It would have evicted one of these keys at random. ⁓ And the first one that it could iterate to, and it is obviously not in a particular order, it depends on the hash key. But it would be deleting the first one that it encounters while iterating.

Sneha Mehra (00:19:00)  
and then it would be evicting it making space for K6, K7, K8. And if you look at this, the total keys are limited to five. Obviously, this is not a production grade database. We are mimicking it, but at least what we know is how to structure, where to put our code. And now you can make it as complicated as you want. You can make it as simple as you want, right? So just add more eviction strategies because all we have to do is just add those particular counters.

add those particular last access time and whatnot so that you can very well do this LRU, LFU like implementation if you'd want to do that. Right? ⁓ And that is it. That is it for this one. What we'll do in the next one. So we saw auto expiration. We saw eviction. Now what next? ⁓ The most interesting, like one of the most interesting features of Redis is pipelining. Right? So in the next video, we would look at pipelining. Where from our client, we can pass in multiple commands.

and redis would compute all of them and return result in one shot. ⁓ It's not request and response. ⁓ Instead, we get a large number of commands together and then we would be computing them, ⁓ buffering the results and sending the results in one shot. This is what we would be implementing in the next one. ⁓ So yeah, that is it for this one. I hope you liked this. I'll see you in the next one. Thanks a ton.

—------------------------------------------------------

13  
Sneha Mehra (00:00:00)  
So it is a very popular misconception that Redis is an in-memory data store and it does not give any persistence. No, Redis gives persistence. ⁓ So the idea here is whatever data you have in memory, you can optionally flush it to the disk. So Redis gives us persistence in two flavors. First is the RDB files and second is AOF files. AOF basically append-only mode, RDB is basically Redis database file. ⁓

We'll take a look at what RDB is, what AOF is, and we would be implementing the AOF version of it in our code base. ⁓ So RDB persistence. So RDB is nothing but a point in time snapshot of the data set. ⁓ So when you are creating an RDB or when you are telling Redis server that, hey, go ahead and create an RDB file for me, what Redis server would do is that it would take that point in time snapshot and dump it onto the disk. You can then take that file and put it anywhere you want.

be it S3, be it your Google Drive literally, like it's just a single file output, right? Extremely compact, extremely simple, single file that you can literally port here or there, right? So to do this snapshot, ⁓ RDB is a very good way to do that. It's ⁓ extremely space efficient to be honest, right? But how does Redis create a fork? ⁓ how does, sorry, how does Redis create this RDB file? So the idea is when you trigger or basically,

The way you can configure it, you can configure it a flush frequency. And you may say that, hey, every five minutes, create an RDB file for me. It would take whatever data it has in in-memory data structures that it has, it would take it and put it into an RDB file and flush it. ⁓ So that is the idea. But imagine if Redis is single threaded, ⁓ and you issue this command, ⁓ what would happen?

your because it is single threaded the client who issued this command would have to wait for this process to complete and while that ⁓ is happening it cannot process any other request coming from any other clients. So that is where what you do is or rather what redis does is it creates a new process ⁓ when it is dumping the rdb file it basically forks out a new process and that process

Sneha Mehra (00:02:20)  
⁓ has the access to this internal hash table and it flushes it in the rdb format on the disk. This has ⁓ no impact on Redis's performance, it can still continue to accept incoming requests, handle them and respond to them. ⁓ While other forked out process does all the heavy lifting of dumping and basically creating a snapshot out of it. ⁓ That is great, rdb sounds great, then what's the problem?

Problem is dumping the entire RDB file again and again and again and again becomes costly as the data size would increase. So that is where what you do ⁓ is you obviously define a flush frequency. Let's say this first flush frequency is ⁓ one minute. That's very frequent by the way, very frequent. Let me take a realistic number. Let's it's five minutes, ⁓ once every five minutes. So if you do that, every five minutes you are flushing it, which means every five minutes you are creating a fresh RDB file. But then what might happen?

Let's say you created a fresh file four minutes back, right? In that next four minutes, you accepted a lot of updates in memory. And after that, your process crashed. So before the next flush could happen, the process crashed, which means that ⁓ from the last flush till this time, nothing is written on the disk because Redis is in memory. Nothing is written on the disk. So that volatile data ⁓ is gone. Problem, right?

So this is where the problem of RDF or the problem with RDB files coming that because it's a full snapshot, it would take ⁓ you cannot do it extremely frequently. You have to take your time with that. ⁓ If you do it very quickly, the problem is that ⁓ your ⁓ machine would only be doing this, right? And you don't want that to happen. So that is where what you do is there is you use another format to get durability.

that AOF. AOF is append-only file. It is primarily like a commit log or a bin log that you might have heard of. MySQL has bin log. Similarly, Redis has AOF file. So the idea is AOF logs every single write operation that happened on Redis server. Write operation, not read. So get would not be logged. But if you're doing set something, that would be logged in this append-only file. ⁓ And the best part is

Sneha Mehra (00:04:40)  
the append-only file that would be created, that append-only file ⁓ will be a raw dump of the incoming command that you are getting. ⁓ It would be basically RESP encoded, which means that for you to replay the log, it is extremely simple. You can literally open the file in your favorite text editor, update if you'd want to update something, right? And then you can actually see what has happened. There is no binary format. It's literal, the command that you are getting, it is flushing it there.

just one change, just one change. For example, if you are updating something, instead of that update operation being recorded, it actually records a simplified set version of it. ⁓ It's just like, for example, if you do ⁓ INCR, that would be incrementing a value by one. Let's say your key was there, key's value was four, and then you did INCR. INCR made value five. So instead of logging INCR, it would log set k five. ⁓ Instead of doing

INCR instead of logging INCR it would directly log the set commands. So if a set or if INCR resulted in ⁓ the value being changed from 4 to 5, it would be writing set K 5\. ⁓ This is what it does behind the scenes and this is what we would be implementing today. ⁓ But before we do that, let's understand it a bit more. So periodically the entire EOF file, so now here if you can think about it,

your AOF file, if it is continuously writing a lot of commands, then what would happen? The commands that will be written, for example, if I have a particular key K, and I'm setting new, new values on it, let's say it was V1, then I change it to V2, then I change it to V3, then I change it to V4. So, what is happening is if I log every command in this append-only file, what would happen is your append-only file will grow to be very big.

But your data set, now here I've put set operations, all set operations happen on the same key k. The problem with that is your data set only contains one key, but your AOF file contains four entries. So what Redis does periodically, is periodically does an entire rewrite of AOF file in the most efficient way. So for example, ⁓ if my data set, like although I've hired four entries, after which my value of k became v4, when I'm triggering a particular command, the name of the command is bg rewrite AOF.

Sneha Mehra (00:07:00)  
So in background, rewrite the AOF file. So if I issue this command, any client issues this command, what your Redis server would do is that it would go through the data set and create a new AOF file. So if my data set has KV4, maybe after hundreds of set operations, it would just register set KV4. That's it. So periodically, because it is rewriting the AOF file, AOF file size remains in check.

⁓ So, advantage of AOF, AOF are much more durable. Here you can imagine every single write operation that is happening is continuously recorded in a file, which means even if your machine crashes, even if your redis server crashes, ⁓ can, ⁓ while booting up, it can load this AOF file and reconstruct the entire in-memory data set that it had. ⁓ So, if you are writing, let's say if you are flushing AOF once every second,

⁓ Which means for one second you buffering all the write operations and then flushing it once every one second. ⁓ So what would this mean? At max your data loss would be one second. Similar to where with RDB your data loss was roughly five minutes, here it just reduced to one second. ⁓ And now let's say you wrote something and while writing something, something crashed or rather write was incomplete or write was corrupt.

How do you check that? So Redis gives us a tooling called RedisCheckAOF. This is a command line utility that Redis ships with its binaries. And you can use this to check if the AOF is valid. If not, you can actually fix it as well. ⁓ This is the idea behind ⁓ RedisCheckAOF, which is what we would be using to validate if R, the file that we created is proper or not. So RedisCheckAOF says it's valid, which means it's valid. Your Redis would be able to understand it if you'd want to load it.

Now, one very critical thing that I want to talk about. Now here, we talked about ⁓ when we are triggering a function like bg rewrite aof. What typically happens is writing in background. ⁓ The command starts with bg. ⁓ It is writing in background. But imagine that you already have a file in which some reads and writes are ⁓ happening.

Sneha Mehra (00:09:22)  
and then you are triggering a background rewrite or you are rewriting the entire AOF. So, how would you ⁓ do that like how would that happen? So, the idea here is that whenever you trigger a background rewrite of AOF, it is written in a temporary file. It is written in a temporary file and then ⁓ once the entire write is done, then it is just renaming it to your configured file. So, the by default of configured name is append only dot AOF file right or.

It's up to you. can actually change it through redis.conf, but whatever the default name is, it would be written to that. Like that would be renamed to this. Right? So temporary file, once it is flushed, it would be renamed to appendonly.aof. Right? And this is the idea. This is how it would do that switch so that your existing writes don't break, but you're still rewrite becomes successful and it takes its place.

And then after few seconds or few minutes, it would do again rewrite, like entire full rewrite. So everything would come in place. AOF files, one big disadvantage is AOF files are bigger than RDB files. See, RDB is a very compressed format. It's a single file that is output. AOF literally stores the commands that you have fired. ⁓ Given that it is going to do that,

That is where your file size would be continuously growing. And this would be much, much, much larger than your RDB files. For the same data set, ⁓ RDB files would be much compressed, much compacted as compared to AOF files. Right? So enough of this theory. Let's ⁓ build this. ⁓ Let's see this in action on how it would look like. Again, the code is available at github.com. ⁓ The code is, sorry, my bad. The code is available at github.com slash diceDB slash dice. ⁓

Let's go through the source code. here, ⁓ not many files have changed. ⁓ Most of the changes happened in the previous video. ⁓ Instead of accepting one command, we had to accept multiple commands. A lot of changes have been done there. ⁓ Now, focusing on persistence. ⁓ So what do we do? What do we change? We first implement a command. ⁓ We start with the eval file. We implement a command called bg rewrite aof, ⁓ whose job would be to

Sneha Mehra (00:11:39)  
rewrite AOF right so passed it to eval function eval function I'm invoking a function called dump all AOF I'm not making it asynchronous I'm not creating I'm not forking out a new process that is a to-do right that is a to-do that we might do in some other case or you might just while implementing you can make it asynchronous I'm just focusing on the core functionality of it enhancements can take its own time right so now let's do this dump all AOF what it does dump all AOF is written in a new file called AOF.go

So in the code base you'll find file AOF.go, which contains everything about append-only files. ⁓ So what do we have here? So dump all AOF, opens a new file in which config.AOF file. In the config we have written dice master, like dice-master.AOF file, which would be written over here. ⁓ So we are creating this, if file doesn't exist, create it, open it in write-only mode, and continuously append. ⁓ Then ⁓ in case we can't open it, we'll... ⁓

print an error and then we log on the server side that rewriting the AOF file at this location and then we do dump something and then we say that write is complete. I'm just doing a server dump. Nothing. It will not be sent to the client, but on the console log, you would be able to see that. ⁓ But what we are doing, so while dumping AOF file, as I said, I'm keeping it simple because we ⁓ up until now we have only implemented get and set.

So, we will just keep it simple. idea is go through all the keys and the values that we have and then ⁓ dump it in the AOF format. And the AOF format is literally the format in which you are issuing a command. ⁓ So, for example, if my data set has key k value v, instead of writing key k and value v in the AOF format, I will write set k v ⁓ in RESP-invited format such that it is something that is given as an input, ⁓ as a command. ⁓

How ⁓ are we giving command to the users or from ready CLI if I am doing set kv to my ready server what would be said it would be sent as an array of strings ⁓ right it would be sent as an array of strings where first value would be set then second would be k third would be v right so we would be encoding it exact same way so I am accepting what I accepting over here is right to this file this key and this object so and I creating a command out of it called set ⁓

Sneha Mehra (00:14:03)  
kv and this command I am splitting it by space and then encoding it ⁓ and ⁓ dumping it to the file. Literally nothing more. If I go here at encode, in encode I have just added it an array of string as a type and I literally creating a new buffer encoding it, ⁓ basically encoding the string each value and then I am just appending it over here. Here look at this.

% a star %d because now for example if I'm dumping set kb right set kb because it is an ⁓ array of string it would be dumped as what it would be dumped as an array and array is encoded as ⁓ star with the number of elements slash r slash n and Then followed by n resp encoded elements right so here we have created n resp encoded elements by iterating through all the values and

encoding them as strings and then what I doing is I am doing star %d %d ⁓ will pass len of v and because it is the length of the array so star 3 slash r slash n and then resp encoded strings set k v this is exactly what we would be dumping so aof file is nothing but imagine your ⁓ redis cli dumping those commands at one place so that you can replay it to your redis server

as simple as that. ⁓ This is exactly the changes that we would have to make, nothing more nothing less, as simple as this. So, let us quickly take a look at the implementation part of it or rather let us see it in action. ⁓ So, let me quickly run this server, ⁓ run main dot go, ⁓ but before that let me clear if there is any file. ⁓

rm ⁓ do we have an ai file no we don't have an ai file so what i will do ⁓ is i will ⁓ go run main.go right i've started the server then i'm connecting it to my redis cli ⁓ in redis cli i'm doing set k1 v1 ⁓ and then i'm ⁓ storing it set k2 ⁓ v2 ⁓ and then k3 ⁓

Sneha Mehra (00:16:29)  
⁓ For example, I dump 3 keys k 1 v 1 k 2 v 2 let me change this k 3 v 3 you will also see this in action right. So, first I write k 3 v 4, but now I override it with k 3 v 3 right. So, my data set has 3 values although I made 4 operations, but it has 3 values. ⁓ Now, let us say if I ⁓ my bad I executed it on 6 3 7 9 instead of that I should do it on 7 3 7 9 ok my bad. ⁓ So, let me do a set k 1 v

v1 set k2 v2 ⁓ set k3 v4 and now I am again rewriting it to set k3 v3 ⁓ or set k3 v3. So, now I have made four sets but my value which is written is only three like my data set only has three values ⁓ k1 k2 k3 right. Now, let me fire this command vg rewrite aof.

If I fired this, you saw in the response, okay, but here you see an output. Rewriting AOF file at .slash dice master dot AOF, AOF file rewrite complete. Now let's ⁓ see what is written in the file ⁓ AOF. Cat dot slash dice master AOF, and I'm doing a less of it. Wait, let me just open this in this different place. ⁓ CD workspace, ⁓ then you have.

dice d v ⁓ dice tail minus f like tail sorry instead of tail let me do cat and then dot slash dice master ⁓ with less.

Sneha Mehra (00:18:11)  
Let me just do cat slash ⁓ dice master. ⁓ permission denied. I have to do chmod 666.slash dice master. So, ⁓ edge case, I should have handled that. But let me do a cat of it. Now, here in the cat, you can literally see the commands which we fired. ⁓ So, let me just do that. Cat with ⁓ less. See, the first thing that we see is star 3, dollar 3, set.

⁓ $2 k1, $2 v1, ⁓ $3, then star 3, ⁓ $3 set k2 v2 and then set k3 v3. Although we fired four commands, we are just getting three, right? Because our data set contains three keys, right? And this is a genuinely valid AOF file. How do I know it? Let me fire a quick command. Red is tools, ⁓ red is ⁓ stable.

and in source I have redis ⁓ check AOF tool and in which I am passing dot slash dice db. So, this is the default AOF check AOF tool that is issued by redis. I am just putting it and ⁓ passing r generated AOF file to that and we get response AOF is valid. ⁓ So, the output that we generated or the AOF file when we did bg rewrite AOF the AOF file that was generated

is indeed a proper one. ⁓ So now if we are rebooting our redis server, we can pass this AOF file and ask it to reconstruct it with this. And it would just work fine. ⁓ And this is about persistence. This is how we could, should and would implement AOF. Just one change, we did it synchronously, instead we should have forked a new process and done it.

But that is a to-do item. That's an interesting thing to implement. it would be really fun to implement it that way. ⁓ But I hope you got the essence. I hope you got the idea of how persistence works. We saw what ⁓ the two ⁓ modes of persistence that Redis provides, or two types of ⁓ persistence that Redis provides. One is RDB file. Second is AOF file. RDB is the compact data representation of the data set. It's a single file used as a point in time snapshot.

Sneha Mehra (00:20:28)  
it cannot be extremely frequent. AOF on the other hand can log every single command that we are getting. We have not implemented continuous logging, but we implemented complete dump of the AOF file like bg rewrite AOF. ⁓ So, this is all about persistence that I wanted to cover. You saw how easy it was to implement persistence in it and we are still heavily Redis compliant. We are literally using Redis tools to test our changes.

⁓ It's working perfectly fine. I hope you found it amazing. That is it for this one. I'll see you in the next one. Thanks a ton.

—-----------------------------  
1  
Sneha Mehra (00:00:00)  
So Redis has to be the most amazing, most versatile modern database out there. Although it is heavily known to be used as a cache, but it is much, much, much, ⁓ much more than that. I'm sure you would have wondered how it works internally, why it is so fast, how it can handle large number of TCP connections while being single threaded. So hey folks, my name is Arpit Bhaiyani and I am bringing you this course on Redis internals where we will learn all key things about Redis, not just theoretically, but by re-implementing them.

in Goliath. So let me give you a very quick walkthrough of this course. So this is the page of my course, which you can find at arpitbhani.me slash redis-internals. I put the link in the description down below, but let me just give you a very quick walkthrough. So the course is a self-paced course. It's a paid self-paced course in which we will be understanding Redis internals by actually re-implementing its core features. Core features like event loop. We will be writing our own event loop, our own serialization protocol.

our own persistence, our own pipelining, eviction, transactions and whatnot in Golang. So the database that we'll be designing will be a drop-in replacement of Redis, which means that while you building this database, you can connect it with normal Redis CLI. It will understand the Redis CLI will understand what your server is doing without having to change anything, which is the best part, which means that we are not just building any key values too.

We are doing things exactly how Redis does behind the scenes. ⁓ So this is where you will find all details about the course. Let me still give you very quick walkthrough. By the way, the database that we are designing, I'm naming it as DiceDB. This database, the source code is open source. You can find it on GitHub. GitHub.com slash diceDB slash dice, which is where you can find the entire source code. The course actually goes through and re-implements the source code step by step so that you can understand why we are doing what we are doing. ⁓

The course curriculum is very very very fascinating split across 8 chapters. The total time that it covers is 9 hours. So 9 hours of really deep engineering redis internals. ⁓ I had a ball while shooting it. So the first 3 videos of it are publicly available on YouTube so that it gives you a glimpse on how the course is structured. The first chapter is about starting up where we lay the foundation of our database. So course introduction is this particular video.

Sneha Mehra (00:02:25)  
What makes Redis special and writing a simple TCP Ecoserver. These three are available on YouTube. The second chapter is about ⁓ the serialization protocol of Redis called RESP, PING and Event Loop. This is where we will be implementing our own Event Loop. first we will be understanding the serialization protocol, RESP implementing PING, IO multiplexing Event Loops, handling multiple concurrent clients. Then chapter three about implementing get, set and delete.

the ⁓ most basic thing that any in-memory key value store or cache needs to support, we'll be doing that. Then we'll be touching upon eviction strategies, auto-expiration. ⁓ Then command pipelining, one of the most amazing features of Redis, we'll be implementing that along with AOF persistence. Then we will go deeper into how every single object is represented in Redis around objects, encodings, and we'll be implementing INCR command. Then we'll be implementing...

the statistics, the most basic statistics of the database ⁓ and the most fanciest LRU algorithm called the approximated LRU algorithm that Redis supports and we'll be implementing that. Then we'll be touching upon the memory management of Redis on how it actually does. Then we'll be doing graceful shutdown, music signal handling, implementing transactions. These are, we'll be implementing database transactions and we'll end the course with a bunch of theory, a bunch of theory on how

fanciest data structures that you know that Redis supports like list, set, geospatial queries, strings, hyperlog, log and approximate counting. How it uses the actual science and engineering behind it. How Redis makes things extra efficient. We'll be going through that. The entire thing is roughly nine, roughly it's, it's around nine hours content, right? So nine hours of really deep tech things. Now here I'll not be wasting time.

typing the code. I'll be giving you exhaustive code walkthrough because I don't want to ⁓ spend a lot of time doing like hitting back spaces and changing my code. So the code base is implemented step by step. You can take a very detailed walkthrough of that. I'll be giving a very detailed walkthrough of the code that I've written. Right. It would help us save time and we will be covering the most crisp like this will be covering all the topics in a very crisp, crisp manner. Okay. So why should you end?

Sneha Mehra (00:04:41)  
First, we'll be touching upon the internals. I'm sure you would love to know how Redis works internally. So that's a key highlight of that. Then, knowing the unknowns, you become a better engineer. You would know so much about implementing database transactions, single-threaded systems, event loops, IO multiplexing, persistence, f-syncs, signal handling and whatnot. Then, you get doubt resolution. I'm sure when you go through the course, you will have a bunch of doubts. And it's not that I would be leaving you stranded. The doubts will be resolved. Two modes of doubt resolution. First is asynchronously.

through Discord. So if you like when you enroll into the course, you will also get access to the Discord community, the Asli Engineering Discord community, where you can find a relevant channel and you can post your doubts. I'll be there to answer that. ⁓ And second, over synchronously, obviously there are some doubts, some explanations which cannot happen asynchronously. Right? It takes a long time to explain something. So which is where the doubt resolution will also happen synchronously every once every two weeks for over 30 minutes Zoom call that happens.

Now this call would be common. It would not be one on one. It would be common to all. The link would be scheduled. A calendar invite will be extended to everyone. It's optional for everyone. I'll be there on that call. ⁓ And basically whatever doubts you have, you can ask them synchronously on that. Right? And obviously network and community, you become part of the Ascle engineering community. Right? Okay. Glimpse of the course. This is where the first three videos of the course, the course introduction, what makes it special and writing a TCPco server.

are publicly available on YouTube, I really really really urge you to check them out. Right? So if you find it helpful, do go ahead and enroll into the course. It's going to be fascinating if you want to be a better journey. If you are curious about how radius works internally, this would be, you would have a time of your life 100%. Okay. Then program prerequisites. Obviously because we are implementing it in GoLang, you need to have basic familiarity with GoLang. You just cannot say I don't have any programming experience and I'll jump into it.

It's not a very beginner friendly. You need to have basic familiarity with Golag and you need to have ⁓ very much curiosity to learn like you should be able to put in a bit of effort into implementing it. You need to have Linux based development environment. So what I'm doing like why Linux based specifically? Because we'll be writing our own event loop. Now writing an event loop requires you to integrate with basically kernel with system calls. Now here I've used Linux based system calls to interface with IO ports. To do that,

Sneha Mehra (00:07:07)  
we need Linux based environment, right? So if you have Windows, you can use WSL, which is Windows Subsystem Linux. If you are on Mac, I would highly recommend you to spin up a Docker container and build this. having a Linux based development environment is needed for this. And third one, have a Google account because I'm too lazy to support any other login method except sign in with Google, right? So you need to have, you absolutely need to have a Google account so that

I can like so that you can sign in into the portal and access the course content. Okay. So once you meet all the prerequisites, this is how you can enroll into it. This is the fees. Fees might change depending on what time you are looking at. I typically would be revising it once every year. So I would highly encourage you to see the fees that you see on the web page and not rely on this video. Right. But what you'll get, you'll get nine hours of Redis internals, curated resources to explore internals better every single video or every single topic that I'll be covering.

would have an exhaustive list of external resources that you can refer to to get a deeper understanding and a better understanding of it apart from everything that I've explained. Source code of our GoLang-based implementation called DiceDB. It is open source. can find it on GitHub, github.com slash diceDB slash dice. Then biviquely doubt solving every Tuesday, 7.30 PM to 8 PM. Time is fixed, ISD. Join that synchronously to get your doubts answered around anything around radius internals.

Then you'll get lifetime access to course videos and notes, lifetime access to us in engineering discord community. Right? So click here, make the payment. You'll get all the details over email on where to access. By the way, you can access the course on courses.arpithbhani.me. The link will be shared in the sign up email that you will receive. Right? And this is about me ⁓ and FAQ. If you have any common questions, those are all answered in the frequently asked questions. You can find it ⁓ all of them right here. Right? So this is a very

brief walkthrough of the Redis internals course. Just one cave yard. Here, what I am doing. First of all, we are not going to implement every single command of Redis. The thing that I have specified in the course are the exact things that I'll be implementing. For example, get, set, TTL, info, LRU, INCR, those, because everything else is just fancy features, fancy data structures. So which is where instead of spending a lot of time implementing data structures, I'll be covering them explicitly over here.

Sneha Mehra (00:09:30)  
while implementation of basic commands that you know will be, you can see this. This is the exact curriculum that I'll be covering. Second, I'll not be live coding things. I'll be giving you detailed code walkthrough of our GoLang-based implementation. Why? Because I don't want you to feel that your time is wasted, that in nine hours, if I'm just pressing a lot of bake space and writing code, instead of that, it's theory, and then we would be implementing. And in the implementation, I'll be giving an extremely detailed code walkthrough.

This way we all utilize our time really really really well. But in any case, would highly highly highly encourage you to go through Redis internals to see the things and experience the magic, the engineering, the science behind it. It's so fascinating when I was deep diving into it. I loved every single minute of it. Right? So three videos are already live on YouTube. The first one is what you are watching right now. What makes Redis special? Writing a TCP code server. Get a gist. Get a feeling on what I'm going to

If you find it interesting, you find it amusing even a bit, highly encourage you to enroll into this course and make the best ⁓ use of your time and know the internals of world's most amazing database out there. So yeah, I hope, ⁓ I really, really, really hope that I would have sparked a bit of curiosity in you to explore Redis internals. If I did, go ahead, enroll into the course and I really, really, really hope to see you on the other side. Thank you so much.

—-----------------------------

18  
Sneha Mehra (00:00:00)  
So in our implementation of Redis, we capped the number of keys that we can ingest. ⁓ As soon as we ingested that many keys, and then if we try to add one more, we trigger eviction out of it. ⁓ So that we basically evict an existing key so that a new key can be ingested in that. But what Redis does? Redis does not block at the number of keys level. It blocks ⁓ at the level of amount of memory consumed. Which means that there has to be a way through which

the database knows the amount of memory that it has consumed. To understand this, we go through the Redis source code and look at a file called ZMalloc. This is the one that does all of that magic, right? So ZMalloc, what it is, it is basically the total amount of allocated memory aware version of malloc. So the idea is pretty simple. They've taken malloc and they have wrapped all of the functions of malloc with ZMalloc and ZMalloc is their implementation. So in which the idea is,

Whenever we are allocating something, if the allocation is successful, we just keep a counter and this counter would keep a track of amount of used memory. So every single allocation that happens, it key, value, any data structure, anything, instead of using malloc, they use ZMalloc. ⁓ The function signature remains exactly the same, but the idea being that you need to keep a track ⁓ of amount of memory that you are consuming so that

when you'd want to see how much of memory is myRED is actually consuming, you just need to check that particular variable rather than firing a system call to see what's the memory usage of this process. You can just refer to that integer variable that would tell you the amount of memory that you are consuming. ⁓ And this is the idea behind ZMalloc. And which is where, which is what you'll find in the file ZMalloc.c. ⁓ Here you'll see all the implementations ⁓ of the actual malloc functions.

Like instead of actual implementing model, it's all about wrapping it. But one very interesting thing we'll see over here. Just a minute, let me scroll through. You'll find something very interesting that what it does. So here we'll start from the first one. The first thing that we encounter, the first interesting thing that we encounter is called ZMallocDefaultOM. OM stands for out of memory. So here, this is the default OM handler in which that we are just printing.

Sneha Mehra (00:02:22)  
that ZMalloc out of memory trying to allocate these many number of bytes. ⁓ So whenever we are hitting that particular limit, we can trigger this function and that would evoke or rather that would print this, fprint this particular thing on my server console. So your ready server crashes, unable to handle it and that's where you are aborting it. ⁓ Then another interesting thing is over here, ZTRIMALLOC underscore usable. ⁓ This is where you see a lot of changes being done. ⁓

So this is where the actual allocation happens. So I'll just quickly tell you this, check the line 110\. ⁓ This is where you see the actual malloc function called happening, ⁓ in which you are specifying the minimum malloc size plus prefix. This is specific to radius, but you basically get the idea that instead of directly invoking malloc, you're just wrapping it behind ZMalloc so that you can keep a track of the amount of memory that you have allocated. So here you requested this size of memory and some pointer which needs to be updated.

like which needs to point to that allocated memory is what you are passing over here. And then you are doing some asserts on basic size checks and then you are triggering the actual malloc, right? So for the given size, you are triggering that, hey, this is the amount of size that I want to allocate. Malloc would literally allocate it from the operating system and put it over here. And you put it in the pointer. ⁓ If there is no pointer, which means allocate like malloc failed, then return null.

So return null implies that your allocation itself failed and you would throw that out of memory error. ⁓ And in case there is a specific limit, have malloc sizes that particular macro. In case you do that, this is where you are computing the size of the point of the amount of memory that you have allocated in size and you are updating ZMallocStat allocation. If you check this particular ⁓ macro, you see you're doing an atomic increment on the used memory. This used memory is keeping track ⁓ of

the amount of memory that your redis is consuming. ⁓ underscore underscore n is the number of bytes that you would want to increment or decrement. So whenever you are allocating or you freeing, you are also atomically incrementing and decrementing this particular counter. So now if you would want to check what's the total number of memory consumption at this moment, all you have to do is just check what's the value of used memory. That's it. That's all you have to do. Right? Okay.

Sneha Mehra (00:04:42)  
So then we do this, we do this usable pointer and stuff, ⁓ Then this is the actual usable like this ztry malloc is basically the function name says that we are trying to allocate a particular chunk of memory and if it fails, it does something, if it passes, does something. That's why this is the helper function for every other function called. So now here we see zmalloc. So it's like malloc equivalent with the exact same signature sizeT and size. It would take this internally, it would invoke ztry malloc underscore usable.

and then it would pass in the pointer, it would get it. If it is not there, you are triggering ZMalloc out of memory handler, right? And then you have ZTRIMALLOC, ZMALLOC usable ⁓ and ⁓ no cache. Like this will come to in a couple of minutes. And this is all the functions which are actually, ⁓ these are the functions that are actually over, not really over, but basically wrapped. So they have wrapped up all of the common utilities that malloc only be called through ZMALLOC.

so that you can keep a track of this thing. This file contains all of these functions, right? So, reallocation, calloc, all of that stuff goes over here. The free would internally invoke free and then it would invoke update ZMalloc stat free to do that, right? Now, given here, you can very clearly see what we are not doing. So, ZMalloc is not trying to say, hey, beyond this, you cannot do it, right? That is what is very important.

and interesting that ZMalloc is not trying to dictate how much you can use it. It is just keeping track of the number of bytes that you are using. Now it is your responsibility or your code's responsibility to do that that hey am I going far beyond that? Because if your malloc is failing which means your operating system does not have enough RAM. But if let's say you are capping at some place then it's your responsibility to handle that. That is where one place that I would want to show this in action.

is here. If I go through the file evict.c and if it starts skimming from the top, first thing that we encounter is creation of pool. Here evict pool allocation. In this evict pool allocation, here we are doing Z malloc. We are not calling normal malloc because we want to keep a track of the amount of memory which is consumed. Right? So that is where we are doing a Z malloc over here. Now,

Sneha Mehra (00:07:00)  
This is where if we allocate the pool, the amount of memory that used memory ⁓ integer variable would be doing plus plus plus plus plus plus every time. And when you do three, it would do minus minus, right? And this is not just for this eviction pool. This ZMalloc is used across the code base. If I just do a simple command F across the code base, you'll find a ton of places where anything that we are allocating, a small integer, a string, anything, it is all done using ZMalloc.

⁓ Every single thing, look at the amount of references that you would find in the code base for this. So many references because they're not invoking malloc at all because you want to keep track of every single thing, not just keys and values, but additional data structures, additional pointers, eviction, everything, everything depends on that. ⁓ So this is where we see that in action, ⁓ how ZMalloc is used. And, but I just want to add one more thing to that where what we are doing is if I just scroll a little below,

⁓ This is where you'll find, okay. ⁓ Let me do that. Z melloc on this file here, right? So over memory ⁓ after allocation, right? This is that function. So it says it returns one if used memory is more than the max memory after allocating more memory, which means basically it says that the more memory that we have tried to allocate.

if after allocating this much of memory, if you have gone past max memory. So for example, if my Redis has 1 GB of RAM that I'm setting up as max memory, ⁓ obviously your operating system might have 2 GB, 3 GB of RAM, but you are saying your Redis that hey, you are capping it at 1 GB. And if after allocating you go beyond 1 GB, then you do something. So this is where it's doing. So this is where what it is doing is, ⁓ if server memory is not set, which means there is no limit,

you are allowed like your redis server is allowed to go and allocate as much as you want you would return zero because it has not gone past that. Right? And then you check it. So ZMalloc used memory. Now if you check this function, this function simply returns used memory. That's all it does. That is all it does. It just gets the used memory and it puts it and it returns it. And what we have to do is we have to use it because ZMalloc you try to allocate. ZMalloc will go through it if there is memory available.

Sneha Mehra (00:09:21)  
Because if malloc is successful, ZMalloc will allocate it. Right? But then it's you, your logic has to check if you are crossing the limit or not. Right? So ZMalloc has this very clear boundary that what I'll do, I'll wrap existing memory allocation functions and I'll just keep a track of a variable that would tell you the exact memory utilization at this moment in time. That's all I do. Now here you'll see that after the allocation, if it has gone past it, so memused is more than max memory.

If it is more than, sorry, here we are doing that. If it is more than max memory, then do something. If it is less than, then do something. It's basically, this function is simply returning true or false around that. Return 0, return 1 around it. ⁓ But this is where we see that we are not invoking anything. We are just, your ZMalloc is just keeping track of things and then we have to take a rest of, okay, now what else we have to do. ⁓ Now if you'd want to evict, by when would we evict? How would we evict? How much would we want to evict?

So this is where you will find all of those functions over here that keep on evicting it until you find or basically until your memory here it's there. Like you continuously start evicting the stuff until your freed memory is less than or rather sorry if your used memory is less than this thing or your used memory is less than the max threshold that you set. Right? You would be running this loop up until that time. This is what it is doing.

Right, so you'll find this code over here in avic.c and this shows how to wrap an existing functionality, that existing function call that we have like malloc, you are wrapping it up into that so that you can keep track of the amount of memory that is used and just setting up the boundary. Most people would have coded this way that, hey, let my ZMalloc itself fail. That ⁓ is not a good way to break it because then it should basically be returning.

you the amount of memory use and then you can take a call if you'd want to proceed or abort or whatever you'd want to do with that. Right? That's a very interesting design decision that we saw by doing this. Right? And here you'll in this one just look for ZMalloc ⁓ and especially in the file ZMalloc.c you'll find this particular code and in evic.c this is where you'll see the evictions happening when it crosses that particular limit. Right? So yeah, I would highly highly highly encourage you to check that out. Here this

Sneha Mehra (00:11:39)  
Now you'll say, hey, why are we not implementing this in Golang? Because Golang does not give us an easy way to implement this. So that is where we are relying on Golang's memory management module to take care of that when we are still restricting it to primarily using something called as or rather the default memory allocation that is we are simply using that and we would be capping it at number of keys for this implementation. But there are ways through which Golang allows us to do manual memory management.

For example, you can embed C code in Golang to do this exact same part. ⁓ But unnecessarily over-complicating that particular code does not make sense for demonstration purpose. So this is more about knowing the internals of it and how it actually functions rather than actually doing that grunt work of implementing it. ⁓ But there is a way in Golang using JEMalloc through which you can do that. That is one library you can use. can...

use native C bindings to use malloc inside your goal encoder you can use TC malloc as well to do that but not touching upon that part ⁓ a little out of scope but this is how it happens right so if you are ever building a project slash software slash product where you would want to cap on the amount of memory that you are using this is how you can do that right and yeah that is it that is it for this one i hope you found it amusing ⁓ it had a lot of nitty-gritty details

⁓ I would highly encourage you to go through ZMalloc source code. Very simple, just wrap over it, but you'll see the beautiful implementation of it and how it has drawn that very beautiful line between where it fails and where it returns an error. Right? So yeah, that is it for this one. I'll see you in the next one. Thanks, Atta.

—------------------------

17

Sneha Mehra (00:00:00)  
In the previous video, we theoretically understood the approximated LRU algorithm. ⁓ And in this one, we would be re-implementing it in Golang. But before we do that, let's go through the actual Redis source code to see how they have implemented because there are a lot of interesting nuggets out of it. ⁓ So in the Redis source code, you'll find a file called evic.c. In this file, you'll find the implementation of approximated LRU and LFU algorithms. So if you start skimming through this file, the first macro you encounter is called ev pool size set to 16\. This is that constant pool size that we were talking about in the previous video that

What happens it samples a bunch of keys, ⁓ puts into this pool of size 16\. So it at max samples 16 elements and sorry it basically samples five elements and keep pushing it into this pool. And during eviction the worst or rather the best candidate out of this is evicted, the one with the most idle time. ⁓ So this is what determines the maximum pool size. If we skim a little more we get a function called getLRUclock. ⁓ This is where we talking about the 24 bit clock. So here

When we spoke about it, how are we computing the current clock? The current clock is computed as my epoch seconds last 24 bits. The last 24 bits of epoch second is what we are considering. Correct? So here it's how it's computed. So MS time would give you ⁓ Unix time in milliseconds divided by LRU clock resolution. This LRU clock resolution is nothing but 1000\. So Unix time milliseconds divided by 1000 is equal to Unix time second, which is epoch seconds. Bitwise AND with LRU clock max. Now this LRU clock max is nothing but

one left shift by 24 bits which is the LRU bits that we have minus one which will give us F F F F F F. ⁓ Right? So this is where that particular magic is happening and this is where anywhere someone calls getLruClock it is getting the 24 bit resoluted clock of my current timestamp. And then if we skim through even further we come to this function called estimateObjectIdleTime. So for given object what is the idle time of that? ⁓ Right? The algorithm that we discussed

that if the current clock or if ⁓ the current clock is greater than my objects last accessed at the idle time of this object is the difference between the two. But if my current clock is smaller than that then the idle time is the max minus the last accessed at plus the clock. This is exactly what is written over here. You get the LRU clock. If LRU clock is greater than object error LRU that is the difference between the two. Otherwise it is LRU clock

Sneha Mehra (00:02:25)  
plus object minus or objects or sorry, LRU clock max minus objects LRU plus the clock. Right? So this is where that particular logic is written. Right? And then if we skim through it below, we find the approximated LRU algorithmic implementation. The first thing we find is the eviction pool allocation. This is the function that would allocate the eviction pool of size 16\. Right? To see with to see like from where it is invoking this.

We go to the file server.c and we find this part eviction pool allocation and this is executed during the init server function. So as soon as my server boots up, the first thing it does is eviction pool allocation in which it is allocating a fixed pool of ⁓ 16 elements where it would hold this particular information. Right? Okay. Then what's next? Next is eviction pool population. So whenever my pool is empty or whenever my pool does not have enough candidates, we would be invoking this function to populate this pool. So population of this pool happens how?

We sample like we know what it does. We know it samples a bunch of keys. It picks five, ⁓ around five keys from there. It samples them and puts it into eviction pool if the ⁓ keys that we are inserting are better than what we already have in the pool. Right? So this is where that happens. So get, ⁓ so dict get some keys is basically sampling keys from the dictionary. In the count, we get the number of keys that we have sampled. We iterate through that ⁓ and then we try to insert it.

But here there is one very interesting part. Look at this comment. This comment says, if the dictionary we are sampling from is not the main dictionary, but the expires one, we need to look up the key again in the key dictionary to obtain the value object. Now this is something new because here they are talking about two dictionaries. We thought Redis was storing all the keys in one ⁓ dictionary, wasn't it?

but it looks like there are two dictionaries in which Redis is storing stuff and out of which there is something called as an expires dictionary and other dictionary is the key dictionary. So now this is where what Redis is actually doing is that Redis because the most common use case if you think about Redis is that it wants to very quickly access the key having some expiration.

Sneha Mehra (00:04:44)  
So that it can just iterate over the keys that have some expiration set and do the needful job, be it eviction, be it sampling, be it what not. ⁓ And there would be one key where it would be storing key and the value. So the other dictionary is what is the expires dictionary in which it is storing to simplify it the key and the time at which it would expire while the value is stored in the first dictionary. The key dictionary holds all the keys and all the values, but the second one holds the keys that have some expiration set.

This is what the key, this is what the expiration dictionary is all about. So Redis does not have just one dictionary behind the scene. It has two dictionaries behind the scene. One for normal key value, other for keys that has some expiration set. It would not store value for that. Every key value stored in the key dictionary itself. But this expiration dictionary would help you identify which of the key has some expiration set so that you can make your algorithms much better. Let's say when you are doing eviction, like in this case.

you would want to sample from the keys that have some expiration set. ⁓ So having that specialized dictionary would help you because now you have to sample from this particular dictionary only. ⁓ So this is what we found a very golden nugget of how Redis is internally doing things. ⁓ So what next? Then we've some business logic written over here. If we skim through that, that is where we find that, hey, now here what we are trying to do is,

Because this is insertion, are populating the pool. How would you populate? We would want to keep this pool in ascending order. So which means we would have to support insertions in the middle of this list. So for example, if my eviction pool has 5 elements, the maximum length is 16\. Now if I insert an element, this element can go at the first position, at the last position or somewhere in the middle. Because it is an array, whenever it goes

at any of this place you might have to shift some elements. So if it is inserted somewhere in the middle, you would have to shift the other elements to the right and insert this one element. So this is exactly what it is doing. Here it is finding the place where it would want to insert and then ⁓ moving the elements. So for example, there is a case where it's not the worst element, not the best element. So basically the idea is we would be inserting the element ⁓ only when it is worse than the worst element of the array.

Sneha Mehra (00:07:08)  
or of the eviction pool, then only it would be inserting it. ⁓ Otherwise there is no point inserting because other elements are much much much worse. So that is where think about it, if you keeping your array sorted by idle time, the worst element or the last element of the array that you have is having the most idle time. Given that you would be evicting from that spot. So whenever you are inserting a new element, if it is not worse than the last element, then there is no point inserting into this. So some edge case and obviously implementation could vary, your implementation could vary. But

The idea is that you would be inserting only when it's worthy for it to be inserted. And which is what you doing. And once you find a spot, what you are doing is you are putting it into this by invoking memmove. Memmove would be literally moving, literally copying the bytes from one place to another place. And then you are creating that spot and then adding your spot there. So ⁓ instead of applying a for loop to do a copy element by element, it just triggered a memmove call.

and moved the entire array and then plugging the stuff in. ⁓ Which is what it is doing over here. ⁓ And this is what is about populating the pool and it is being done in the loop. So your pool size would still remain 16, it would never exceed 16\. ⁓ Right? Okay. Then the other part is LFU. We would skip for that. You can very well explore it on your own. ⁓ So it uses the same LRU bits, but it stores something else in there. It does not store the last access that it stores the frequency.

within that same set of 24 bits. ⁓ You can go through this code to understand how LFU works. And then if you skim through it, that is where you find the actual function, which does eviction. I'm just scrolling it through and here you find perform evictions. This where this function is the one that is actually doing the evictions from the eviction pool. So it would be the one who would be finding the worst key. Now worst key would be what? Worst key would be always at the end. you're placing it in the ascending order of the idle time,

it would be stored at the end. So it would start iterating from the end and would be continuously keep on evicting the key. But up until when? It would be keep on evicting the key until you have ⁓ memory. Like until it brings down the memory beyond that threshold. Now here, ⁓ our implementation assumed the maximum number of keys that Redis could hold. ⁓ In the code that we wrote, we saw that maximum keys we are setting it to 100\. So our Redis implementation would not go beyond 100\. But that is not a good way to do it.

Sneha Mehra (00:09:35)  
because imagine a key or imagine a key having a value which is 1 MB big versus the key having a value just 1 KB big having a key having just 4 bytes of integer stored. Right? So you don't you cannot give everyone equal weightage. Right? So that is where the utilization of Redis is more about the memory that it consumes ⁓ rather than the number of keys that it holds. So you can specify the max memory that your Redis can consume. Right? So which is what is important. So now

when you are breaching that memory, the difference of that would be the amount of memory you would have to free. So your loop or your eviction loop would run ⁓ when, like up until when, when your memory freed is less than memory to free, which is where we are computing over here. The two things that how much memory do we have, how much memory do we have to get to and what the difference between them is the amount of memory that we would have to free.

And given that we have that memory to free, are iterating this loop again and again and then evicting the keys from the eviction loop. And during and while we are evicting the keys, we have to make a lot of checks because Redis is much more sophisticated database. It just does a bunch of checks there, best key and then puts it somewhere and cached and what not. But the idea is still the same. It evicts the key, which means that it would know the key that needs to be evicted and then you have to go to the corresponding dictionary in the key store and remove that. You would have to...

that trigger a multiple hooks that are there. ⁓ So expiration set eviction happens and notification needs to be sent. All of that code is written over here. Right? And some edge cases being handled over here. Right? This is where that particular logic is written. ⁓ Where you are actually evicting the particular key from it. And this is where you see you got a

triggering the propagate expire that hey this is where the expirers happened on this particular key so you may want to notify other systems like other components of the not really system but other components of the code so that it can log the information put it into server log and somewhere around that right and this is how it basically does that this is what the implementation of eviction the actual eviction looks like right feel free to go through this we basically got the idea on how it actually does that and now we can go ahead

Sneha Mehra (00:11:49)  
and reimplement this exact same thing in our own implementation. So here we go. So now here, what we are doing is we are taking the inspiration from Redis's code base and doing something very, very, very similar. So we start with the file object.go. So earlier we used to have expire set over here. We don't need expiration over here. What we do is we are storing last access that. Now, Golang does not support bit fields.

Right, we saw it last time, it does not support bit-pitch. So cannot have a 24-bit, unfortunately I cannot have a 24-bit storage, so I am just going for Uint32. So 32-bit unsigned integer is what I going over here. But obviously you can implement it by clubbing type encoding, last access that both here, but we are just not over complicating it for now. Right, so last access that would be storing every time the key gets accessed, ⁓ you would be updating this particular thing. So if I go to store,

you'll find places where you are storing that or where you updating it. But before that, here we have defined another map, the expires dictionary, similar to how Redis does that. Redis has a key value dictionary, which is this, and it has expires where it is setting all the keys that are expired. So here what we are doing is we are storing the actual key value over here, and expires is a dictionary where we are storing the object pointer along with the expiration time. So maybe if you want to do something with that, we can do it, right?

So this where our expiration is set. So now we don't need to store expiration at object level. We are just storing it over here. ⁓ Then initialization, we initialize that setExpire is the helper function that we wrote in which we take the expiration duration in milliseconds and we set the expires on the object to be that particular value. So ⁓ current time units millisecond plus the duration will give us the expiration time of that particular object. Then when we are creating a new object,

We are also specifying now specifying the last access stat which is my get current clock. Now get current clock is doing what? It's just doing the 24 bits of that. So un32 of Unix, Unix will give me time Unix epoch seconds ⁓ bitwise and with F F F F F F. It would give us the 24 bit clock resolution over here. It's written in the file eviction.co. Now here when that has happened, I'm just setting the expiry and I have this function of set expiry which sets the expiry to the object. When we say set expiry,

Sneha Mehra (00:14:07)  
What does this mean? This means that we have to create an entry of that particular object in the expires dictionary, which is exactly what we are doing. So for this object, this is the expiration duration. So expires of object is equal to the absolute expiration time in milliseconds is what we have stored over here. Right? So that's done. So if when we are creating a new object, if some duration is passed, what we do? We call set expiry. Set expiry on that object for that duration is what we are storing. And then input.

So whenever my object is created, you set the last access that whenever my object is put, which means updated, I set the last access that to something. And then when you're accessing the object get, ⁓ then here what we are doing earlier, we used to have that logic. ⁓ If your expired set is not set and something, something, I've just simplified it with this function called has expired. So has my object already expired or not is this function does. This function takes in an object.

and it returns if this object has no expiry set, it returns false, which means object has not expired. Otherwise, it returns expiration is less than equal to U64, like current milliseconds. How less is my expiry than that? Because AXP, the expiration is stored ⁓ as the absolute milliseconds time. That's what we are storing. So this is what would empower us to do this. ⁓ So the expiration that we are setting is less than equal to

current Unix milliseconds, if it is then my key has expired. Right? Okay, that's there. So if the key has already expired, we trigger delete. Otherwise, we set the last access that to my current time block. So every time my object is put or object is get, I'm updating my last access that. And then during deletion, we were only deleting from store first. Now we are also deleting from the expires dictionary. And because delete is a no-op operation, which means if the key does not exist, it does not do anything.

We are just triggering a normal delete on the expired dictionary. Everything else remains the same. ⁓ Right? Okay. Then what takes? Given that we have changed a bunch of things, I've just made alterations in eval.go file. I'll just quickly skim you through that. Where instead of doing that re-computation again and again, I'm just invoking this function has expired and then doing that part. Similarly, during TTL that same change happened over here and then over here. Right? Nothing fancy. It's just about ⁓ we change the definition.

Sneha Mehra (00:16:29)  
So that's why I'm just using the helper functions everywhere. Here we were setting the expiry explicit when we're calling the expire command. Instead of the time just invoking set expiry over here. So object and duration seconds multiplied by thousand would be the duration, like expiration duration in milliseconds, right? And which is exactly what I have changed in this file. Nothing else has been changed. Now comes the interesting file, which is eviction file. Now in the eviction file, ⁓ we do, ⁓ now we implement approximated LRU.

implementation. Okay, so now we start skimming through that we do something very similar get current clock gives us a 24 bit resolution of the clock which is the least significant 24 bits of my current epoch seconds. Then get idle time does that exact same thing that we discussed it gets the current clock if my current clock is greater than last accessed at so for any timestamp for any timestamp in the world it would give us this particular information. ⁓ So if current clock is greater than last accessed then the difference otherwise max minus last accessed at

plus clock. ⁓ This is my idle time, which is what we would be used or we would be using to evict the element or put the element in the eviction pool and then eventually evicting it. ⁓ Then function to populate eviction pool. ⁓ Function to populate eviction pool is we are going by the sample size would be like my sample size is five, at max I am sampling five elements. ⁓ Assuming that my iteration on the map or iteration on the dictionary is in random order, ⁓ I would be iterating through that.

And then what I'm doing is I'm adding it to my eviction pool. We'll look at the implementation of eviction pool. Right? So here, Epool is my eviction pool. Within this eviction pool, I'm pushing the key ⁓ and my last access rat over here. So whatever my last access rat is, I'm just pushing it. Let me quickly open that. The eviction pool is a simple array. It's not an optimized implementation. Right? What I'm doing is I'm repeatedly sorting it to keep things simple because basic implementation to understand is much more important than very fancy implementation. We know

how to very fancily implement it. So here in the eviction pool, what we have? So pool has two things, the array of pool items and keyset. Keyset implies the keys that exist in my eviction pool so that I can very quickly check how many are there, which are like if key is present here or not, like basic check. Then I'm having a custom sorting criteria on idle time. This is the comparator function where I'm just using it to sort this time. You'll see it in action here.

Sneha Mehra (00:18:54)  
Now when we are pushing it into my eviction pool, what do we know? If the key already exists, I don't have to push it again in my eviction pool. Correct? So this is that check. So pq is my eviction pool object dot keyset of key. If this key exists, then do nothing. Because if key already exists, why to push it again? Because it's an array. Why to add that key again? That would be problem. So we are not adding it. Right? Then we create a pool item. Now eviction pool contains

⁓ is an array of pool item. Pool item just contains the key and the last access tag. That's all it contains. The key and the last access tag. So I created this item. Now we check. If my len of pool is less than equal size max, which means there is some space for me to do that. Some space for me to add this sampled element. I would be adding it. So I'll add, I'll make an entry in the key set. I'll be updating the pool over here. I'll be appending it to the pool. But remember ⁓ eviction pool needs to be sorted.

by ⁓ idle time, but we don't have it sorted. That's where I'm just always, this is a performance bottleneck, but I'm just always sorting my pool by idle time. I'm just sorting my pool by idle time every time. It's a very naive implementation, but that's okay. We can improve it later once, right? But you can just run an insertion sort to insert it at the right place, not denying that. But for now we are just sorting it again and again, right? So if there is space in the connection pool, we are adding it.

if there is no space in the connection pool. ⁓ But, but my element that I am just sampled is worse than the worst element that I have, which means one with the most idle time because I keeping it sorted because I might have the worst element. So then what I can do is I will create a space for that and would be adding it to my loop, to my pool, sorry. ⁓ Right, so this is where we are removing the first element.

The first element is the smallest one. keeping it sorted. The first element is what I'm removing and I'm otherwise ⁓ adding the new element into this and I'm just appending it over here. This way we ⁓ would be ensuring that my pool always contains the best possible candidates to be evicted. Right? That's what I'm doing over here. And then pop function simply does error like edge case handling over here and picks the zeroth element.

Sneha Mehra (00:21:14)  
and then sets the pool and basically removes it, ⁓ returns it and then resize my array ⁓ from first element to the last element. So it's basically kind of like we removed that particular part from this. ⁓ Then this is how you would be implementing eviction loop over here. Anything else that we changed, we changed expiry, checked get expiry, there are helper functions that we added. ⁓ But now where do we invoke this? This is where ⁓ our expire function comes in.

and object is done, expire is done, ⁓ store is done, ⁓ anything else that we would want to, ⁓ we just changed this eviction policy, so it would automatically be invoked. Okay, here, this is the function that is remaining. So eviction pool is populated, ⁓ this is where we work, eviction pool populated, and then all keys, LRU implementation. So these are what we are doing is, ⁓ anytime I hit my limit, ⁓ anytime I hit my maximum keys limit, what I would need to do,

is that first invoke populate eviction pool. And this is again a nime you can make it much better. I'm ⁓ just explaining it so that you know the base of it. Then you can make it as complex, as efficient as you can. ⁓ So I'm first populating my eviction pool and then I'm finding the number of keys to be evicted. I'm iterating through that. I'm popping out from the pool and doing it until ⁓ like if I don't have anything from my pool, it returns nil. So I'm just not doing anything. I'm returning from that.

Otherwise, if I have some item element from ⁓ some item popped from my pool, I'm invoking a delete on that. So delete would be deleting it from the dictionary itself and dictionary which means both store as well as expires. ⁓ So this is where it would be doing both. ⁓ And this is where we have configured the ⁓ LRU all keys LRU implementation. So evict all keys LRU through this every time the evict function is triggered, it would be triggered.

when we hit that limit, every time it is triggered, it would be invoking all keys LRU over here. Right? And then this is how the entire flow would look like. So now that we saw how to implement approximated LRU, let's see it in action by going through the actual graph where we see that the number of keys does not go beyond 100\. Here you can very clearly see that the maximum number of keys that my cache ever holds is 100\. It goes below, it goes to 96 and then it grows back.

Sneha Mehra (00:23:38)  
up again and it goes below and goes back up. This is what we see. ⁓ So as soon as my LRU cache hits that upper limit, eviction is triggered. It basically evicts the keys. The worst, the best possible candidate that could evict, it evicts that on the basis of the idle time. ⁓ And this is how you would be implementing your all keys LRU implementation. It's extremely simple, extremely fascinating, extremely interesting algorithm.

to implement and I would highly highly highly highly encourage you to implement it. So what next? ⁓ While we were skimming through Redis's source code, we saw a very interesting word. So while we going through eviction pool malloc, here we saw a word called ZMalloc. I thought malloc was the function that we use to allocate memory in C language. But what is ZMalloc? So ZMalloc is something which Redis has implemented, which is basically wrapping up the actual malloc function.

so that we can keep a track of the amount of memory used. Remember, we made it simplified. We said that, hey, at max we would have 100 keys, but Redis works on the memory utilization. Redis says that at max we would allow ⁓ 1 GB of RAM or 2 GB of RAM, which is configurable. So it needs to somewhere keep track of the memory which is used by Redis. This is where it does that. Right? So in the next video, we will go through ⁓ how Redis does memory management so that it could cap its utilization.

⁓ So yeah, that is it for this one. hope you found it interesting. hope you found it amusing. would highly highly highly urge you to implement this. You can find this source code at github.com slash dice db slash dice our goal and best implementation. You can skim through the commits and find this one relevant commit and go through the changes that we made. So yeah, that is it for this one. I'll see you in the next one. Thanks a ton.

—-----------------------------

16

Sneha Mehra (00:00:00)  
So in the previous video, we implemented all keys random key eviction algorithm, right? In which we were picking keys at random, giving equal chances to any keys, whichever we iterate first, we were just evicting it, right? We were just deleting it, right? But is that fancy? No, LRU is fancy because ⁓ for your cache to be efficient, which means that your cache should be holding the keys that are very likely to be accessed again. So if a key

is recently accessed, it is much likely to be accessed again, then LRU is the best algorithm that you can use. Where the idea is that a key that is not recently used should be evicted. Now, how does Redis implement it? You will say that LRU is a very standard algorithm, why are we spending time talking about this? LRU. Redis does not use exact LRU implementation. It uses an approximated LRU algorithm. ⁓ Now, first let's understand a KNIVE implementation.

An IIM implementation would say that, ⁓ let's store time. ⁓ Like for example, if I have 10 keys, I'll store the time at which my key was accessed. ⁓ Every time my key gets accessed, I'll be updating this time. Let's say I store it in a field called last accessed at or something. ⁓ And every time my key gets accessed, I'll be updating this timestamp. Then what's the problem there? ⁓ Then I'll be just doing a sort of the keys, ⁓ finding the one that is least recently used and I'll be evicting it.

This is slow, right? Doing that at scale, at ⁓ Redis's speed is extremely slow. You have to make it fast. Second, to keep them ordered, let's say you are not sorting every time, then you would have to maintain an order of these keys, which is ordered by your last access that every time a key is accessed, you would be have to moving it to head or tail depending on your implementation, right? A classic way to implement this is using a double-link list because every time a key gets accessed, you would be taking that node and placing it at the head.

at the head of the list and eviction will always happen from the tail because your head would have the most recently accessed keys and the tail would have the least recently accessed keys. So you would be evicting when you are triggering an eviction you will just start editing from the tail up until the point you would want to evict and you can start evicting that. Right? What's the problem there? Multiple problems. First of all, W-link list requires extra memory. Why? Because

Sneha Mehra (00:02:21)  
a doubly linked list node not just stores your key but also the pointer to the next object, pointer to the previous object and the pointer to the key itself ⁓ if not the actual key. So node itself becomes heavy. ⁓ And for a database for an in-memory database like Redis this becomes catastrophic because instead of using the space to store data you are using the space to store pointers and manage this thing unnecessary. Second,

is whenever your last access time changes, it requires you to shuffle a lot. And that would be ⁓ taxing on the CPU because CPU has to do this ⁓ pointer manipulations every time. Plus random memory access is going to happen. Problematic. It would reduce your throughput. You would want this to be extremely simple. Third, storing last access time itself takes up 32 bits per object because for every object that you are storing, you would need to store the time at which that object was last accessed.

Given that if you are storing time at which it was accessed, time, if you consider four byte integer, then you are requiring 32 bits to store that particular information. Again, extremely, because imagine for all the objects, for all the keys that you are putting in, you have to store this extra time for every single one. Problem. So that is where this naive approach does not work for Redis. What it does, it does approximation. But how? So first of all, let's quickly jump through the source code once.

so that we see where it is making this particular choice. So ⁓ a few videos back we saw what Redis object was, right? So every single object that we are storing in Redis is stored in the Redis object where you are having pointer to that object type object encoding that we studied and it stores this LRU bits. So LRU ⁓ is stored. This is where it would store time. The time at which your key was last accessed. This is where it would be storing it.

⁓ So first it store four bits in type, four bits in encode, not bytes, bits. Four bits here, four bits for encoding and 24 bits for LRU. So it means that to store this LRU information, like the time at which it store, it was last accessed, it is just using 24 bits and not 32 bits. ⁓ You think the 32 bit integer is small, it's just using 24 bits to store this information. ⁓ But if you look at this, this total becomes,

Sneha Mehra (00:04:43)  
4 bits plus 4 bits plus 24 bits equal to 32 bits. So which means in one 32 bit chunk you are storing type encoding and LRU. Then you have a reference count and reference pointer. ⁓ Right? So this, ⁓ so then how it is actually doing it? It's not even utilized. So you basically cannot store time there. The exact time that you get like time.now.unix milliseconds you cannot store it. You don't have enough space to do so. Right? So how does Redis solve this problem? That's

what the beauty of this algorithm is. ⁓ Very interesting design decisions that these guys have taken. I loved while I was going through it. Let me ⁓ walk you through on this amazing, amazing optimization. Okay. So we'll start with the first one. First important key design decision to store last access time in 24 bits. Now you'll say, at saving 8 bits, what would it give you? Right? 8 bits per object is huge. 8 bit is one byte. If you have a million keys,

then you're saving one MB of space in which you can fit in a lot of other stuff. ⁓ So eight bits per object is huge for Redis. Plus, if you look at the Redis object, you have type encoding LRU bits taking up 32 bits, then integer reference count and a void star. So 32, 32, 32\. So it is basically taking up 12 bytes of space. ⁓

a multiple of four becomes easy for your CPU cache lines to access it. ⁓ So no random, ⁓ because it's very efficient for you, for your CPU to access objects which are in the multiple of fours. ⁓ So that's basic computer level, ⁓ basic computer architecture level optimization. But think about it, if it's multiple of four, it becomes easier for your CPU cache, or sorry, for your CPU lines to read and write. So now, how does it store it?

because it does not have 32 bits. Let's say when we do time.now, we get an integer object, a 32-bit integer. But we are not having 32 bits, we are having only 24 bits. So what do we do? That is where we take the least significant 24 bits out of it. So basically time.now gives us the time in seconds, for example, or it gives us time in seconds. So time in seconds, that integer value ⁓ and

Sneha Mehra (00:07:02)  
⁓ 0x00ffff so 6 times f right so 24 bits when I do AND of 00FFFF it would mask all of those first 8 bits to 0 and everything else would be copied as is so now if I'm having a large time which is let's say 32 bit time I have 32 bit time and I'm just absorbing the last 24 bits of it everything else is trimmed down to 0 right so this is how I would be storing the last access stat

but we are losing on exact time because we are not storing exact time. This is still incremented every second. So the number would increase every second, but this is not a classic clock that you are thinking of. This is a trimmed down version of our own LRU clock. Let's call it LRU clock. Right? So this is ⁓ very like this has one to one almost one to one correspondence with the actual time that is there, but we are running our own clock kind of that.

Right? So whatever time we get, we ⁓ mask the first 8 bits out of it, we mask it to 0 and everything else is considered as is. Right? So which means our value, our time is now fitting in 24 bits. It's not exact time that we are storing. We cannot, it's a lossy thing. We are removing the first 8 bits out of the picture and just using the last 24 bits out of it. Now ⁓ imagine this as a duration. The value will range from 0 to 2 raised to 24 minus 1\.

If I treat each value as one second, I would be covering 194 days, like almost three, like basically greater than three months. Right? So that is the duration that we are talking about. So ⁓ if my last accessed at of a particular object is zero, right? So then every time if I move forward in time, let's say every time that object gets accessed, the clock is moving forward every second.

⁓ And the total duration that my timeline can span is 194 days. ⁓ Second key decision. Now this will come in extremely handy. We'll take an example to understand this. Second key decision is idle time. So instead of having a classic LRU implementation, look at it with a twist. So LRU evicts the key that is least recently used. A least recently used key would have the most idle time. ⁓

Sneha Mehra (00:09:29)  
So which means if my key is accessed at a particular time ⁓ and then let's say my clock moves forward, your clock always moves forward. Let's say your clock move forward by five seconds. So which means your key is idle for five seconds. Right? So I'm just twisting the terminology from LRU that key was key when it was last used. I'm using the terminology of idle time. This will come in handy. Right? So now what do we do? Let's say

We simplify this case and instead of having 24 bits, I work with only 5 bits so that we understand it well. ⁓ Let's say if I work with 5 bits. Now what would happen? 5 bits implies my clock will range from 0 to 2 raised to power 5 minus 1 which is 31\. So possible values that I have is 0, 1, 2, 3, 4, on and so forth till 31\. ⁓ And if your clock is moving forward after it hits 31, it would again come back from 0

and so on and so forth. So the cycle would repeat continuously. ⁓ Right? So your actual clock is using normal time. ⁓ No problem then. But because of this 24 bits, ⁓ you would have to circle back automatically. ⁓ It's normal plus plus when you do an AND on that, ⁓ you're automatically masking it. Right? So this is how your circle would move when you do plus plus on that. As your time moves forward, this moves forward and these are the possible values that you would be getting. ⁓ Right? Now let's say I have two keys.

Let's say your clock was at time 16, that 16 seconds. It was at 16 seconds and you set a particular key, let's say K1. So the class access time of K1 becomes 16\. And then the clock move forward and forward and forward and at 24th, you set a key K2. So K2 is set as 24\. The last access time of K2 is 24\. Now let's say a clock move forward and at time t equal to 27, what happened? Your LRU kicked in.

and now you want to do eviction. So eviction you want to do least recently used key. Which is least recently used? It's K1 because it has the highest idle time. So let's compute the idle time of that. So idle time for key K1, idle time for key K1 is 27 minus 16\. 27 minus 16 is 11\. So it's 11 seconds of idle time from my current timestamp. Right? And for K2 it is

Sneha Mehra (00:11:55)  
27 minus 24, which is three seconds. So here you can very clearly see that idle time for my key K2 is smaller than key K1, which means K1 is least recently used key and I can evict it. Pretty straightforward algorithm. So you can just take the minimum of the time and just evict it, solves the problem, right? But what happens after five seconds?

You evicted key K1 because it was least recently used having the most idle time. Let's say after five seconds, what would have happened is your time, your clock from 27 would have moved forward to 31 and it would have started over this. ⁓ Now let's say your clock, because it is moving, your LRU clock, because it is moving ⁓ from 31 to zero, then one, two, three, four again. Let's say when it was at four again, you accessed or you set the key K3.

And so then last access time of K3 is equal to 4\. And then a couple of seconds later when your time is 6, ⁓ right? Your current time is 6\. What's happening is you are triggering LRU eviction. Now you cannot just use minimum of these two. Because now your key K2 has been sitting in for a very long time. Key K3 is recently written. If you take the minimum of 4 and 24, you would be evicting K3 because it is smaller than these two.

So your last access time, you cannot just take minimum of last access time to evict it. Right, because K3 is definitely not least recently used. It is in fact the most recently used key. The least recently used key is K2. You should be evicting K2 and not K3. But if you take minimum of these two, like your classic LRU implementation, you would be evicting K3, which is wrong. So how do you solve this problem? So you cannot take minimum. So that is where the, like,

Thinking of LRU with a twist of idle time solves your problem. So now what do we do? Let's think of the ⁓ idle time of each of this key. So idle time of each of this key is that for K3, the idle time is what? K3 was last accessed at four and my current time is six. So my key was, my idle time is what? Two seconds, right? But for K2, it's circled back.

Sneha Mehra (00:14:14)  
So for K2 last access time was 24\. So it circled back. So the idle time for that is 24 to 31 and then from 0 to 6 because our current time is 6\. So it's 31 minus 24 plus 6\. So which means the maximum value minus my last access time plus clock is what my idle time for that particular key is. For my key K2 the idle time is what? Because it ⁓ remained idle till your ⁓ clock reached 31\.

and then up until your clock reach the current time which is 6\. It is sitting idle. Right? So your idle time for key k2 is equal to the maximum value minus that which means this frame plus your current clock. Right? This is what your idle time is. Now if you think about it, now your idle time, if you consider idle time which is same as least recently used but just with a twist so that understanding it becomes simpler, you will be evicting k2. Now how did we know that it had to circle back?

Look at this position of time, your current time. Your current time is 6, right? But your K2 is at 24\. 6 is less than 24\. So whenever your current clock is less than your last axis at time. So which means you have to go full circle, right? But if it is your last axis time is smaller than your current clock, then you can just take the difference. That's the idea behind it. So that is where

you can compute your idle time as clock minus last accessed at if your clock is greater than last accessed at the normal use case. ⁓ If your clock move forward and last accessed at is behind that the difference of weight is your idle time. But if your clock is less than last accessed at which means if your clock is behind but your last accessed at is in the future according to if you just do normal comparison which means you have to circle back. So it would be max minus lat

plus clock that would be your idle time. And now you can compare this idle time and now you can evict which k2 in this case. ⁓ This is how you can implement LRU with just 24 bits. You don't need 32 bits to store it. You are saving one byte every object. Now you think about worst case. What's the worst case over here? The worst case is that your circle like your clock again moved forward and it crossed k2.

Sneha Mehra (00:16:41)  
But in that case your K2 is still idle but you would be a victim K3 that's perfectly fine because the duration for 24 bits ⁓ the duration becomes 194 days. Now imagine for a practical use case for 194 days a key in Redis is never accessed. How unlikely is that? ⁓ Because it is not very likely it's very impractical.

Like a key is just lying there without getting accessed for 194 days and you still want to keep that. Right? Very rare, very very very rare. That is why for a real world scenario this would work just fine. That ⁓ because you have enough large span, we took example of 31 seconds. Here it's 194 days that we are talking about. That the edge case would happen or the worst case would happen.

when your circle, when the key was inserted or when the key was accessed and after 194 days also it is not accessed, then ⁓ your LRU algorithm would start spitting out wrong answers, but that's very rare. So we should not be worrying about that. that's what Redis also does. Redis exactly, exactly implements this. So now you see how we implemented LRU, not with 32 bits, but with 24 bits. ⁓ So now what's the algorithm? Now here, very interesting. Again,

very memory efficient because Redis have to be memory efficient. So that is where what it has ⁓ is the approximated LR algorithm works on approximation ⁓ given. So what it does is it samples N keys from the data set where N is less than equal to five. So from the entire Redis data set, it picks five keys at random ⁓ and it populates in an eviction pool. Eviction pool think of it like your possible candidates to evict. This is your eviction pool.

Right, eviction pool has a fixed size of 16\. So at max you would have 16 candidates over there. Right, you are sampling at max five keys, about five keys from your entire data set and you would be adding it to the pool only if these are better candidates than what you have in the pool. So for example, if in the pool you have a minimum, like for example, if you have your, if you are putting the elements as per your last access or as per your idle time,

Sneha Mehra (00:18:57)  
⁓ If you putting your elements into this eviction pool, you have to take care of is that they are much much much worse than what you already have in the pool. So that when you inserting it, you are only having good candidates to be evicted from this particular pool. ⁓ So keys from the sample are added only when they are better than the existing pool. The pool is kept sorted by the idle time. Hence the insertion can happen in the middle. ⁓

The implementation of it, you may think that, I can use linked list for that, but again, don't use linked list because as soon as you use linked list, you're doing random memory access number one, having additional pointers here and there, number two back, like previous and the next. Instead, what Redis does is it literally stores it in an array and does array copy, right? To move the number and create space for someone in between. It literally does that. We'll walk you through the, I'll walk you through the code in the next video. But this is where our,

our approaches have to be counterintuitive. You think that WLINKLACE would give me order one time, but it's not always about time. When you have a smaller size array, it would not matter. It's all cache by the way, right? It's all cache on the CPU cache itself. It would be lightning fast. So ⁓ no need to worry about that. So you are maintaining this pool of good candidates that are very worse when it comes to or rather that have very high idle times and it's all sampled.

It's not going to be the best match. We don't have to do best match. Here we are optimizing for not using or not overusing memories to maintain doubly linked list and all. We are keeping it extremely simple. 16 ⁓ elements big, you are having an eviction pool. You keep inserting the elements if your existing elements are like, if the elements that you found in the sample are worse than what you have in the array so that you can evict the worst candidates or rather the best candidates out of it.

whenever you are triggering an eviction. So you're using constant amount of memory to do this eviction. You literally have this one array, you're iterating through that and starting from the best candidate to delete your evicting the elements out of it. Right? So during eviction, the best key is deleted from the pool and the keys are continuously evicted until we meet the target memory consumption. For example, ⁓ if, so we were talking about the number of keys there.

Sneha Mehra (00:21:21)  
Right? So in the previous video when we implemented, we saw we put a limit on the number of keys that we have. We put it to 100\. But Redis does it with memory. You can specify the memory that, hey, I want my Redis to use only 1 GB of RAM. Right? So it would work on that. So if it goes beyond 1 GB, let's say it goes to, ⁓ let's say 15 MB beyond 1 GB. So it would continuously, it would keep on evicting the keys until your memory use comes below 1 GB. We will also take a look at how that is implemented.

But the idea is your eviction from this pool would continuously happen until your memory consumption comes below the threshold that you would want for. ⁓ Add up until then that would continuously happen. And this is how Redis implements LRU like approximated LRU algorithm. This is the theory behind it. Very important or rather very important design decisions that they have taken. Very interesting design decisions they have taken. They have kept it extremely memory efficient.

not using doubly linked list, not using fancy data such as to get order one time, they're using approximation, sampling and normal array copy to do that. Brilliant decision, right? This is what you should be thinking about that, what are the challenges with that? A naive implementer, you think that, hey, I'm writing an order one algorithm over here, that would not suit Redis because it requires you to have extra space complexity, right? That does not suit Redis at all because Redis would want to use memory for to store the data set.

rather than managing this peripheral buffers. ⁓ Right? And yeah, that is it for this one. In the next one, we will be going through Redis's source code to see how it implements approximated LRU algorithm and we would be re-implementing it in Golang. Right? So that's what we're going to do in the next video. So that is it for this one. I hope you like this. I hope you found it amusing. I'll see you in the next one. Thanks a ton.

—---------------------------

12

Sneha Mehra (00:00:00)  
So Redis has a very interesting feature called pipelining. A typical way to fire commands to Redis server is that first your client connects to the Redis server. ⁓ It's a TCP connection. Then your client fires the command, server computes the response, it basically does the needful and it sends back the response. ⁓ So first you do set k comma b, then your Redis response, okay. Then your client sends get k and then it sends v.

Right? So here, if I'm just setting a particular key and then getting the value for that particular key, ⁓ I'm doing a lot of round trips because a lot of packets needs to be sent across the network from your client to your server, from your server response back to client, then the second request, then the response to that. So here, ⁓ here, ⁓ if it's within the same machine, you would not see a difference because it's just local connection, it's just local host. But as it goes in an infrastructure over the internet maybe,

this round trip times become extremely critical, right? Because this is the major, basically this is the major cost that would be involved when your client is firing a lot of requests to Redis and your server is responding back. So a lot of time goes into this round trip. So can we optimize this? This is what pipelining actually does. So the idea of pipelining is pretty simple. That hey, instead of just sending one command ⁓ over the TCP network,

What if I send multiple commands, multiple commands in one request? So for example, if I'm firing three commands, right? So instead of sending first command, then getting the response, then fire second command, then get a response, then fire third command, then get a response. Instead of that, what if I fire three commands back to back in one request? So in one round trip, I'm able to send ⁓ command one, command two, and command three, and your server

can independently execute command one, then command two, then command three, and then in one time response, it sends me back the result. So for example, if I do ping, set kv and get k, in response I'll get pong, ⁓ okay, and v. Here, one key thing to note, it is not a transaction. It is just command pipelining in which you are multiplexing a lot of command in one request, sending it to the server. Server will...

Sneha Mehra (00:02:21)  
execute them one after another and it would send you n responses. So if you send n commands, ⁓ you would be getting n responses out of it. So it is just about clubbing the request. ⁓ It is not about transaction at all. ⁓ So a server is independently executing those three, ⁓ those n commands that you pipeline, ⁓ it would compute n responses and it would send you back n responses, ⁓ multiplexed in one response. ⁓ So

Overall, you just do ⁓ one request and one response. In that one request, you're multiplexing a lot of commands. ⁓ So this is the idea of pipelining. Now with pipelining, you'd say, why do we want to do that? But before we do that, if you think about it, when your client is sending multiple requests to the server, now your server, instead of evaluating ⁓ one and then immediately sending back the response,

What it has to do is it has to buffer because your server can only send one response. In that one response, it has to send everything. ⁓ So in request, if I send three commands, in response I have to send three responses. ⁓ So which means that your server has to buffer the responses. So if I'm sending three commands, then our server is evaluating one, output of which is buffered in memory, then evaluate second, output is buffered in the memory, evaluates third, output is buffered in the memory.

And then once everything is evaluated, it is sending back three responses. So here what would happen is it would lead to a higher memory consumption. So if you're doing a lot of pipelining, agreed you'll get great throughput, but it would have an impact on memory consumption. So you have to be wary of that fact. ⁓ But obviously, pipelining helps you improve throughput because now the number of operations that you can run are much larger.

Otherwise, a lot of time of your Redis server goes into waiting for getting the request and then sending the response. So lot of time goes into that because system calls are blocking. So overall commands per second that you are executing is substantially higher when you use pipelining. ⁓ So for Redis, ⁓ and why so? ⁓ for Redis, because all the operations that you are doing, they're in memory, which means that computing them would not take much time.

Sneha Mehra (00:04:44)  
when it receives a command for it, execution literally takes microseconds, if not nanoseconds, right? It would be much, much, much faster. A lot of time of Redis server also goes into making the network call, sending back the response or receiving the response from that, right? Which is where your power of pipelining comes in. With pipelining, your Redis has to do fewer context switch from your in-memory processing then to network IO. Because now imagine if you pipeline 100 commands, right?

So here, instead of like get one command, compute it, or basically make the necessary changes and send it as a response. So, ⁓ IO, you are doing network IO, then doing some CPU processing or some basically memory ⁓ input output, and then you are responding back. And then again, the cycle come back for hundreds of requests. Instead, if you are pipelining hundreds of requests in one request, it gets it, it can quickly fire 100\.

It can quickly evaluate those hundred commands without having to context switch to network IO and then immediately send back. This is the power. This is why it betters the throughput. Right? Now that we understand what pipelining is, let's quickly take a look at it. How do we implement it in our implementation? The implementation seems pretty straightforward, but just a few hunches here and there that would come our way.

⁓ But before we do that, it's always better to see how pipelining indeed works. ⁓ So let me quickly do this where I'll show you, ⁓ like always, top right is where our normal Redis server is running. ⁓ And what I'm doing is, I'll be putting in commands. I'll show you how basically commands are fired, how pipelining is done so that we know how to implement it. ⁓ So here I'm connecting to normal Redis client, not here.

But here bottom left is you can see red is CLI. Let me, yeah, okay. Red is CLI and I'm doing minus P, port 637 because I'm connecting to normal Redis server. ⁓ Now if I connect to that, now let me fire a command. So here I'm sending ping ⁓ and in response I get pong. ⁓ This is normal command that I'm sending. But with pipelining what happens is you have to send multiple commands in one shot.

Sneha Mehra (00:07:03)  
So what we would do is we would be firing a simple netcat through netcat we would be sending the request or the normal bytes to the server. So here out ⁓ here how it would look like. So in pipelining what we would do is we would be just doing this. So here what I'm doing is I'm doing a printf it's a shell printf in which I'm passing command one command two command three pipe with ⁓ netcat localhost

6379\. So, 6379 is our normal Redis server. Right? So, here I am literally putting command 1 after command 2 after command 3 and then sending this encoding this in bytes and sending it as a request to 6379\. Right? This is what the pipelining is all about. So, in one

in one request in one network packet or rather not really network packet but in one request in one IO call I'm sending this entire request in which it has multiple commands literal concatenated if I fire this I'll get some response because it is not proper but let me ⁓ do this right it send a response but it did not get any response because this was invalid let me put a proper command here so if I do this it would look something like this so here

Don't be scared of this. It's normal that we have always ⁓ written the RESP protocol. It's the exact byte that we are sending. ⁓ Star one implies it's an array. ⁓ you, like ⁓ remember when we send commands through Redis CLI, it is sent as an array of strings. So let's say if I'm sending ping, set K comma V and get K, right? If I'm sending that here, you can see we are doing star one slash R slash N dollar four slash R slash N

ping slash r slash n. This is your ping command that you are sending. Right? It's a bulk string encoding. Right? This is how we are sending ping command. Then the second command that we would send is set kv. Right? So set kv is three arrays. sorry. The three element array. So your first thing becomes ⁓ star three, right? Array of a three elements slash r slash n, dollar three slash r slash n set slash r slash n. So set and then dollar one slash r slash n.

Sneha Mehra (00:09:21)  
k then it's slash r slash n then dollar 1 slash r slash n v slash r slash n this became your k v right set k v and then the third command is dollar star 2 slash r slash n dollar 3 get slash r slash n dollar 1 slash r slash n k slash r slash n right and i'm passing this to 637

Here you see the response. ⁓ Right? So we passed in three commands back to back. ⁓ Right? So it's just one request that has went in. And now we are getting in one response, we are getting output of three commands. ⁓ Right? ⁓ So here output of first command is pom. ⁓ Set k, v is okay. ⁓ And then what we are, and then when we fire the next request is get k. Which means we are, because

This is a normal network protocol. That is where you see dollar one, then V. It's basically dollar one slash R slash N V slash R slash N, and which is why it is printed like this, right? So we send three commands back to back and in response, we got three commands or the output of those three commands, right? This is what pipelining is all about. And this is what we have to make changes in our code base to implement this, right? Okay. Now let's move to our code base. ⁓

And this is what our code base actually looks like up until now. So what we would do is we would start somewhere. We would start, the fundamental thing is changing. Like earlier, we wrote a code such that it was accepting in one network IO, ⁓ one request, and then we were sending one response to that, right? Let's just quickly take a look at what would have changed. So I'm in my async underscore TCP file.

And here if you look at nothing has changed up until now, but here something has changed. So our function readCommand which was reading one command because our Redis CLI was sending one command and we're getting one response. That's what we implemented. That now changes to readCommands. ⁓ So we're accepting that same communication object ⁓ and instead of getting one command, I'm getting multiple commands. And in response, instead of passing one command, I'm passing multiple commands.

Sneha Mehra (00:11:42)  
That is the fundamental change we are doing. ⁓ So first let's take a look at read commands. ⁓ So read command was written in sync underscore tcp file, the first ever implementation that we had. Now here instead of sending one, we are sending many. So if you look at this, I've created a new type called rediscmds, which is just an array of star redis command. So the normal redis command that we had, now changes to array of redis star. So that's why I'm just naming it, right? It just multiple commands that we are passing.

The idea here is pretty simple. Earlier we used to do decode one, right? Instead of just doing decode one, what we are doing is we are continuously decoding it. The idea here is that we would want to accept multiple commands through that. So which means that like we just saw when we send the request, we just got multiple commands back to back, literal concatenated. So earlier we were, when we got a body,

⁓ Earlier we got the value from the Redis Client or anywhere. We just ⁓ parsed one object out of it. We never looked forward. Now we have to. ⁓ So that is where what we are doing is we are literally iterating through this. So we are decoding, we get values and values we are iterating through this and each value would be I'm converting it into array string that gives us token and then I'm appending it and I'm creating normal Redis command out of it, which we already did.

⁓ We already created this and now this two array string is doing nothing but accepting an array of interface and converting it into strength. Nothing fancy here. ⁓ So the idea here is I'm decoding everything that I have. ⁓ Like here if I look at this, ⁓ the main function decode now changes. Earlier decode function used to decode that one value. But now instead of returning an interface, it is returning and slice of interface.

This function is that main thing that got changed. So now what we are doing is the idea is pretty simple. A decode function instead of returning one object, it will be returning an array of objects, right? So for example, if I'm passing one, because each command that we are sending from ready CLI to this, each command that we are sending is an array of string, right? And we have multiple search. So it would be array of array of strings, right? But how do we get that? The idea here is pretty simple. So,

Sneha Mehra (00:14:02)  
Here I am creating values, ⁓ decode can iterate through, instead of just decoding one value, we are iterating through multiple such concatenated values until the data ends. ⁓ So here I creating values, having some index, ⁓ decode 1 and decode 1 returns object, ⁓ the delta, like the number of bytes read and error if any. ⁓ Remember we used it during encoding and decoding, ⁓ third, fourth video.

in which when we were understanding RESP, that same thing will be utilizing over here, the delta becomes our friend. So now what we are doing is, let's say I'm passing in three values back to back concatenated. So I'm reading first value, and then my delta would be the number of bytes that I've written over here. So I'll start from here, and then read the second value, then I'll start from here, read the third value. So my decode function, instead of returning one object, it will be returning multiple objects, each decoded back to back, and value is appending over here.

This is the main change that we are doing. ⁓ So decode was returning one object. Now it is returning an array of objects. And then each object, if we are doing pipelining, this would be an array ⁓ of strings. ⁓ So that's the idea behind it. ⁓ Now that we have sorted out our decode function, now instead of having one command, we are getting array of commands. And now this array of commands can now very well be sent for response.

This is where I've read the commands and in respond will do that exact same thing. So, respond instead of accepting ⁓ one radius cmd it is now accepting multiple radius commands. Even if we pass one it would be just array of array of strings. ⁓ So, now we are doing this and I'm passing it to eval and respond. eval and respond now instead of getting one radius command it is getting multiple radius commands. Because it is getting multiple radius commands so instead of having ⁓ direct switch case we are putting it within the for loop.

Right? So, our main logic does not change. It's just that the way we are consuming, we were consuming the input, now it changes. Earlier we were just expecting one command, now we can have array of commands. Right? So, now we are iterating through every single command, evaluating it the way we used to. But now, instead of directly returning it, we are adding it to a buffer. So, earlier the code that we wrote used to, in the evaluation, we used to pass IOReadWriter, IOReadWriter,

Sneha Mehra (00:16:28)  
and then needs to write to it directly, right? But now instead of that, the eval functions that we have, they would be returning the response, they would be returning the output. And this would be added to the buffer and at the end of this eval and respond, we would be sending it over the socket. That's the idea now. So the peripheral has changed, everything else has remained the same, right? So now let's just take a look at how set would work, right? So just taking example of,

Let me just go through all of them. So I just created some helper constants that were there so that we don't have to re-implement it again and again and just copy, like just use this variable instead of ⁓ writing this complex part and rewriting it every time. ⁓ Now, eval ping, instead of returning an, it used to return an error, but now it is just returning a slice of bytes, which is the actual response. So now in case there is an error, we are encoding the error and returning it over here. Otherwise everything else remains the same and output is bytes.

Similarly for set error, are encoding the, earlier we used to directly send error, but now we are encoding the error and sending it. Then we come over here again encoding the error and sending it, encoding the error and sending it, encoding the error and sending it. But at the end, we are just creating this new object and returning response, okay. Right? So here instead of doing socket IO, we are just returning this slice of byte, which will then be sent over the socket by eval and respond. Right?

That's the idea. Similarly, all the changes have been made at every other level. ⁓ Instead of returning error, we are returning encoded version of error, which is a slice of bite. Encoded in RESP format like this. ⁓ If it's an error, we are doing minus percentages slash r slash n that we all learned. ⁓ That same thing we are doing over here. ⁓ And for eval ttr also exact same thing. ⁓ So the idea here is instead of

accepting. So now we are making our code generic enough that instead of just being restrictive that we'll get one command and we have to compute and respond. What we are doing is we are accepting concatenated commands, understanding it, interpreting it, evaluating multiple commands and buffering the output and sending it back. So the eval and respond command that we have, this is where we are buffering all the output because

Sneha Mehra (00:18:50)  
Here we are doing buff.write which will write to this buffer. The output of ping command will be written to this buffer. Then we are doing set, output of set command will be written to this buffer. Then we are doing get, output of get will be written to this buffer. ⁓ And when we pass in these three command, evaluation is done, then we'll do this. C.write buff.bytes. ⁓ So the entire thing would be sent over the socket to my Redis client. ⁓ And this is how you would be implementing the pipelining logic.

We have covered every single file that we altered. Everything is right here. Nothing has changed much. It's just that now we have made our code generic enough that we are supporting pipeline. ⁓ Initially, even when Redis was built, they never thought that they would have to do that. They also started something similar except one command responded back. And then the new feature request came in where they have to make that code as generic as possible. And this is how you would be implementing pipelining. ⁓ We covered all the files that we had to.

And this how you implement Python. Let's quickly take a look at the implementation or ⁓ when we are doing it on our own server. Go run main.go. And now here I'm passing that exact same thing, but instead of passing it to 6379, I'm passing it to 7379\. I get pong, I get plus pong, plus okay, which are simple string that in this response that we sent. And then $1, ⁓ V.

slash r slash n, right? So we just implemented, are just, we have just started supporting pipelining in our own Redis implementation, right? Fascinating, isn't it? So start small and then eventually build complex features on top of it. ⁓ Keep code as generic as possible, right? For us to make this change, we didn't have to change a lot of stuff. It's only the input output part that we had to change because every other thing, every other function was encapsulated.

⁓ And this is how you should be structuring your code so that your life becomes simpler when ad-hoc requirements come in. And that is it. That is it for this one. I hope you found it amusing. I hope you understood how pipelining, like what pipelining is, how it is useful, how useful it is rather, and how we implemented it in our own Redis implementation. ⁓ The idea would remain the same, keeping the code generic, accepting multiple concatenated commands, buffering the response and sending it in one shot. ⁓

Sneha Mehra (00:21:16)  
Earlier we thought that we would be streaming the response that would be much helpful, but not at all. Pipelining changed that flow. Now we have to buffer the output and send it in one shot. Right? Great.

That is it for this one. ⁓ I hope you liked this. I hope you learned something interesting. ⁓ would really, really, really appreciate you to implement. I would really urge you to implement this. ⁓ The source code of this again, github.com slash dice, db slash dice. can go through this particular commit to see how it is implemented. ⁓ name of the commit messages are self-explanatory. I would highly encourage you to check out till this commit and see the changes that we made. ⁓ Right? So that is it for this one. I'll see you in the next one. Thanks a ton.

—-----------------------------------------

22

Sneha Mehra (00:00:00)  
So one of the data structures that Redis supports is list. Now, if I would want to create a list, it's as simple as I can just directly start writing L push, which is basically pushing something to the list. I can specify a key. Let's say I pass in klist as my key. ⁓ And I can pass in the member that I would want to store. Let's say I want to store s in that. ⁓ And this created an element. This created an object of type list within that. ⁓ Now, if I want to see what kind of data structure it is using, ⁓ I can issue a command called

debug object and I can pass in the key over there. So let's say I pass in klist. ⁓ If I do this, here you can see the output of this. In this output you can see that encoding that it is using is quicklist. So this tells us that Redis internally uses an encoding called quicklist to store list. Now let's take a look at what it does, what encoding is. We'll go through the code step by step and understand what it is actually doing, where it is actually doing all of this. Right? Okay.

So let's jump right into the source code. And now in order to find like where is this quick list being created, how it is implemented, ⁓ we take a very simplistic route. ⁓ We fired a command called L push, right? So which means that a place to create a new quick list should be starting from L push command. So here what I do is I can just do a control F of L push command and I'll get its implementation. L push command.

and I will get its implementation. ⁓ Here I can see in tlist, here I can see an elpush command, which is what gets executed when you pass in elpush. And then it invokes get generic command, if I select that, ⁓ here you can see it creates this. If object does not exist, ⁓ then do create quicklist object. This is where it is creating it, which means that this function should be invoking something around quicklist. ⁓ This is where quicklist is there, ⁓ quicklist create. ⁓ This is where we get quicklist instance and we are in the file quicklist.c.

If you scroll back up at the top, ⁓ it's just written one line information about it, a doubly linked list of list packs. We'll talk about this, what these are, how it actually stores. ⁓ So a bunch of theory. So how Redis stores list? As we saw, ⁓ it stores it in something called as Quicklist. And Quicklist is composed of Ziplist, ⁓ or what they call it, Listpack. There's a slight difference between that, but you'll basically get the answer on what it is actually trying to do. ⁓

Sneha Mehra (00:02:27)  
Okay, so regular implementation of list that you would think of is hey, if I want a list, let me just use a wlink list. ⁓ It's the best data structure out there to implement list, isn't it? Like you can create as long of a list as you want. You can do forward traversal, you can do backward traversal, all good, right? But.

Red is, being an in-memory database, what it has to do is it has to be very memory efficient. Because it needs to be extremely memory efficient, which means that it cannot have a lot of overheads. I'll say, what kind of overhead would it have? Which is where...

what we, if we just compute the overhead that a linked list has, it's significant. So let's just do a bit of number crunching there. So for a doubly linked list, ⁓ if let's say I store value as a pointer, ⁓ for every node I need to have a back pointer and a forward pointer, ⁓ right? So if I assume for a 32-bit machine, two pointers and one for value pointer, so between three pointers in each object, bare minimum, ⁓ so it will four into three, 12 bytes, or if it is a 64-bit machine, it would be eight into three, 24\.

24 bytes. ⁓ In Redis, as we saw in the previous videos, that in Redis, every value that it is entering is a Redis object. And our Redis object contains a little bit of metadata, reference count and pointer to a value, which ⁓ roughly is around 16 bytes. ⁓ Now, ⁓ this 16 bytes, on top of that we are adding 24, which means that it has a 40 byte overhead. And we have not even stored the value.

So for example, if we store a five character string in a list, it would be around 10 bytes plus 40 bytes of metadata just to store this Redis object with back pointer and forward pointer or whatnot. ⁓ And how did this 10 byte come in? It came in from the string representation which we'll ⁓ see in a couple of videos later. ⁓ So just to store this five character, this five character string in a list, ⁓ I have to store...

Sneha Mehra (00:04:29)  
I have 40 bytes of overhead plus an extra overhead to store string. If we even discard that 40 bytes of overhead is huge for a 10 byte information that we are storing.

That is where the problem is. So, Redis cannot be efficient when it is using a simple doubly linked list to implement list data structures. So, which is where what Redis does ⁓ is it uses something called as ZipList. But before we do that, we also need to understand other disadvantage of using doubly linked list, a pure doubly linked list implementation. So, ⁓ while we are drawing, well, we always draw doubly linked list, we draw something like this.

We draw three boxes next to one another and connect them via lines. We think, hey, this is so simple. ⁓ We think that these are closely, like they are close to each other. We don't even consider the fact that because it's a random allocation in the heap, it is very much possible that the elements are allocated at random locations in the heap. So when you are traversing through the list, be it backward or forward, you would be jumping from one memory location to another, to another, to another. It might go haywire.

which means that this ⁓ is not cache efficient. Now this cache is what I'm talking about is a CPU cache. So for example, whenever you are accessing an element in memory, that ⁓ memory segment, that 4KB memory page is brought on the CPU cache. Now let's say if you iterate through list one by one, one element by another, and if they're randomly allocated like this, what would happen? For each one, it would have a CPU cache miss, then it would have to go to memory.

the actually memory page fault would happen. It would bring that in CPU cache and then access it. And then that would get evicted and then the other one would get evicted. So that's where the problem starts to creep in. So these kind of random allocation, it sounds very cool in theory. We all learned it in our college days during data success algorithms. But in practice, they are not at all CPU cache efficient because...

Sneha Mehra (00:06:28)  
of a lot of CPU cache misses that would happen. So which is where an alternate comes in called ZipList. Now ZipList is ready specific. So what ZipList does? ⁓ ZipList is a solid block of memory where all the elements of the list are smooshed together. It means that we don't do random allocation here and there. Imagine that you are allocating space equal to five elements. You are not allocating space for ⁓ one element at a time.

you are, when you are invoking malloc, you are allocating space of five elements at a time. So when you are doing append, you don't have to do reallocation. ⁓ That is the idea behind it. So if you can smoosh in a lot of elements in one

array or one zip list that would make your life much simpler. So which is what the idea of zip list is. So it does not obviously hold good because if you put in all the elements of a linked list into that, it becomes an array because the place is a solid block of memory. You can just consider it as an array. If you put all the results, then there is no difference between a link list and an array. So which is where you would want to use zip list as an encoding when you have smaller list. So smaller list, smaller hash table, smaller sorted sets are all implemented using

⁓ And once a threshold is reached, now you'll say, how long should my zip list be? ⁓ Should it only hold five elements or 10 elements or 50 elements? So it's up to you, your configuration, how small you would want your zip list to be. If you put it at a granularity of one entry,

it implies that it would be exactly like a doubly linked list. But if you store, if you say configure it to five or 10 or 15, that means that in each shot, it would be allocating that big chunk of memory. All the append operations would happen in that same block without having to reallocate anything. One that is exhausted, then the next one would be allocated if you'd want to. ⁓ So the way you can check this configuration is fire config get star ziplist star. It would match all the configurations on the Redis console to see where the ziplist configurations are. ⁓ So ziplist is just one

Sneha Mehra (00:08:29)  
It is not a linked list, right? So, ZipList is just one list in which we are keeping in a lot of elements. A lot of it is in the configured number of elements, ⁓ right? So, ZipList is not a replacement for WLinkedList. It's just a small, ⁓ efficient implementation of a linked list, ⁓ which is basically kind of an array.

Right, so how it is able to smoosh a lot of elements together, which is where we now take a look at the layout information of that. So how ⁓ each ZIP list is encoded and stored on in memory. ⁓ So what it has is each ZIP list has some size. It's a solid block of memory. So it will be a solid, contagious block of memory. Right? And in the first four bytes, which means first 32 bits would hold the total size, the total length of the bit list of

the zip list. It could be 1 MB, 1 KB, ⁓ however big you'd want it to be, that would be stored over here. Right? After this four bytes of this, it would store the tail offset. Tail offset is in, when is like the place, the offset in this list, where my last entry is stored. Because you are storing a list in a zip list, a list would have multiple elements. The tail offset would start, would point to the first byte of my last entry in my list. Now this

would be helpful for you to do quick pop operations. So if you moving out of, if you are popping out elements from the list, you have to remove it from the tail and this would help you achieve that very efficiently because you already have a tail pointer. There is not any pointer but you are actually storing the offset within the zip list. ⁓ Then the next element what you are storing, next ⁓ item is the total number of elements which is 16 bits. ⁓ So 16 bits you are using to store total number of elements.

which means at max here you can have 2 raised to 16 elements in a ziplist. That's the maximum. Then you are storing multiple entries, however entries you are there, however entries you may have. Here if you have only three entries, your total elements would be stored as three and then entry 0, entry 1, entry 2 you would be storing and then a specific byte ⁓ which is your end, which signifies that the ziplist ended. It's basically FF which tells us that your ziplist terminates over here.

Sneha Mehra (00:10:40)  
Right? So this way you may allocate a large ziplist but use a small fraction of it.

⁓ And then when you're inserting, you're elongating it or you're appending after that. You don't have to do reallocation there. So this end byte is important for that, which signifies the end, the termination of my ZIP list, which means again, you can allocate a large block of memory, but use a small part of it. As and when you get entries, you keep on adding into that. ⁓ Okay, so this is how your ZIP list is stored. Now, how does each entry within the ZIP list store, which is where what do you have is each entry within the ZIP list, because a list can contain integer also.

string also, another list, it can contain anything. ⁓ So here you need to know how or other it can only contain integer and string, but you need to know how it is actually stored there. So each entry in the ZIP list, like each entry in the ZIP list holds three information. First is it holds the length of previous entry, the length, not the offset, the length of the previous entry, which means if I, now this is such a beautiful optimization. Now if each entry holds

the length of the previous entry if I want to go back, ⁓ can literally do, I know would know the current offset of this, I can do minus that length, I would reach this offset, right? And I can read this entry. So I can literally do backward movement in this list without having to do anything, right? So that is how it would store length of the previous entry. Second, the encoding of the current entry. So let's say if I'm storing integer or string, I would be storing it in a certain format, right? So whether I'm storing int or I'm storing integer, how it is structured,

that information would be stored as part of encoding so that I know how to parse the ⁓ data. So this is how each of the ZipList entries stored. Now here, couple of very good optimizations. First of all, if you are storing small integers,

Sneha Mehra (00:12:30)  
Right? Now you would think that if I'm just storing small integers, like let's say less than 250, less than 255, right? Let's say eight bits integer if I'm storing that. So then you would think that, why I'm wasting so much of space? Let's say I'm just having a list of integers. You would think that we are wasting a lot of space into storing this meta information about it because small integers are of fixed length. Why do I need to store bunch of information with that? So that is where in that case, each ZipList entry is just

previous length and encoding and in encoding itself you are storing the data. You storing the small integer within that. So you don't need extra field for data. ⁓ And you saving space of just reusing encoding field to store your small integers. ⁓ Such a brilliant idea. ⁓ This is what makes Redis ultra special. ⁓ It thinks of every byte.

to be saved so that you can put in large amount of data into the database. Nothing comes for free. That is where it is doing so much of optimizations to ensure that we are working at the full capacity. ⁓ Then here if you look at this, length of the previous entry. Now, how big your previous entry can be? This could be. ⁓

one KB big as well or 100 bytes or 1000 bytes or 500 bytes, something. But given that we have to store this information, like how long the previous entry is. How would you store that? How many bytes would you allocate? Now imagine for each entry, for each Zip List entry, you're allocating four bytes to store the length of the previous entry. That's the basic thing that we could think of. ⁓ If you assign four bytes, four bytes is four billion. I can store very huge amount of data, but...

Imagine most of the entries are not going to be that big. So it will be wasting a lot of space again. So what Redis does, it optimizes that. So what it does is, ⁓ if the previous length is less than 253, which would be true in most cases. So the length of the previous element is less than equal to 253, which means your data can be fit in eight bits. It just uses one byte to store previous length, ⁓ and then encoding, then data. ⁓ And in case,

Sneha Mehra (00:14:37)  
the length of my previous element is more than or equal to 254\. That is where it stores first byte as a special value which signifies F E. Like it's a hexadecimal representation 11110, right? That's what it would be storing F E, right? And then it would be using 4 bytes to store the previous length. Now it is allowing you to store 4 bytes which means you can have 4 billion up until 4 billion.

length of previous entry stored over here. Right? So just using that, so unless it is essential for you to allocate four bytes, you are just, you're using what you have like in one byte itself. Right? That is what the idea is. And why did you, and why do you think it is chosen 253 and 254? Because this is exactly what the special value is. Right? Because you cannot have 254 because it is taken, it is the special marker. It cannot use FF. ⁓

Because FF means end of the zip list, right? So, what is left with? FE. So, that is where it draws that line. Where if it is less than equal to 253, it means that it is just one byte, which is where I storing my length. If it is less than equal to 254, which means I have another four byte, which would be storing the length of my previous entry. And then the encoding and then the data.

such beautiful piece of optimization that you would see throughout the Redis code, the entire code base of Redis. Right? Okay. Now let's talk about the encoding field in the Redis entry. Now, sorry, Zip List entry. So in each entry of Zip List, we know that we stored previous length.

Second, we store encoding and third we store data. Now let's talk about encoding field. Now encoding field is about, depend, ⁓ basically the layout of encoding field depends on what we are actually storing. So this is where ⁓ what encoding field would hold is encoding field would hold if the string, if it is string, the first two bits is the encoding we use to store the length of it. So here the zero byte and the first byte would be used to store the, like it would be used to store

Sneha Mehra (00:16:45)  
the encoding that it is using and then the other bytes are used to store the length. We'll talk about that. Let me dive deeper into this. So for example, ⁓ if in the list, ⁓ one of the element, ⁓ in the zip list, if one of the element that I'm adding is a string whose length is less than equal to 63 bytes. Let's say I'm just storing hello. ⁓ Just storing hello, length is five bytes. ⁓ If I'm storing that small of a value, so what I would have is I would use 00\.

PPPP, now PP sends for anything. ⁓ These are eight bits in my one byte information. Because I'm storing the ⁓ length of the string, 00 signifies that I have a very short string, less than 63 bytes length. So here I can store the length of my list. And so 00PPPP, it's reserved for the length. Okay, then here, ⁓ if it is more than 63 bytes, so 63...

So, why 63 bytes? Because in six bits, you can go to two raised to six, which is 64\. ⁓ That's less than equal to 63 bytes. ⁓ So, that is the length that you would be storing over here. ⁓ And then the data comes in. ⁓ Then let's say if your length is greater than equal to 64 bytes, ⁓ but less than...

⁓ Sorry ⁓ 16383 which is 14 bits 2 raised to 14\. ⁓ So in that case you would be doing is you would be using ⁓ 01 as the first bit first two bits to define the encoding that it would mean that ⁓ this is less than 14 bits so I would want to read the 6 bits over here and 8 bits over here and I would be interpreting it as a length ⁓ of my string that I am storing. So if my string is more than ⁓ 64 bytes it would be using this

thing to store the length and then the actual data comes in. ⁓ In case your length is more than that, ⁓ what's more than that? Four bytes. So which is where you are doing is ⁓ one, zero, zero, zero, zero, zero, zero, zero. Now if it starts with one, this is a special value. It does not support anything like there is no four type here. Then the next four bytes store the actual length of the string. ⁓ This is what it is using to store the length of the string. So the first couple of bits tell the encoding that it is using.

Sneha Mehra (00:19:01)  
and then the other bits after that depending on what the encoding is the other bits after that would signify the length of the string followed by the data. Right? Okay. This was string. Now what if it is an integer? Now here you see that 0 is taken. Right? 0 0 0 1 1 0\. So now everything that is 1 1 is integer because it has to piggyback everything into that one place because there is no way to distinguish between it. So the first couple of bits that you have is all you have.

to see if it's string or int or what not. So it use 0, 1, sorry it use 0, 0 for a short string, 0, 1 to use medium size string, 1, 0 for this. So now, you also have to support integer within that. So every integer starts with 1, 1\. 1, 1, 1, 1, 1, 1, 1, 1 are all integers. So what it does? It basically does that the first two bits are set to 1\. Following two bits would determine how int is stored. Now this is where the first two bits are 1\.

After that, depending on what the next two bits are, would determine how the integer is stored. Now, you have many types of integers. Now, you'll say what type of integers you have. For example, if you have 16-bit integers, why would you want to allocate four bytes of memory for 16-bit integers? That's a waste of space, right? Which is what you would have to do, as you should do, you should be very frugal with your interpretation, should be very frugal with your implementation. So, if you need 16-bit integer, just use 16 bits to store that. Don't use more than that. Which is what we are doing? ⁓

0 0 then the next two bytes would be representing a 16-bit integer. If it is 0 1 then it is 32-bit integer. Then it is 1 0, that is 64-bit integer. Then it is 1 ⁓ 1, ⁓ then it is 24-bit sine integer. Then if it is this 1 1 1 1 1 1 1 it is 8-bit sine integer. And it's Redis specific implementation but you see how frugal ⁓ Redis is while even allocating and it's not just blindfoldedly allocating four bytes for each integer.

It is leveraging every single, it is using every single bit to its fullest. ⁓ This is what makes Redis Ultra special. ⁓ We understood what ZIP List is. ⁓ But now let's see this in action so that you understand this complete picture. So now let's say if my list is 2 and 5, both are strings. 2 as a string and 5 as a string. Now here how my ZIP List would look like. My ZIP List started with 4 bytes ⁓ of what? My total length.

Sneha Mehra (00:21:25)  
Right? So 4 bytes of total length stored over here. Then I had my tail offset. 4 bytes stored over here for my tail offset. Then 2 bytes to store number of entries. Then my two entries and then followed by FF which denotes the end of the list.

Right? Okay. So, ⁓ how many bytes do I have in total? ⁓ 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, which is why here it is stored as 15\. Right? Now, what is my tail offset? Where is my last entry being written? Over here. So, the offset is 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12\.

which is where 12 is stored over here. How many entries do I have? Two entries. So, 0 to 00\. Right? So, here 0f, 0c, 02\. Then, for each of my entry, the first thing is we store the previous length. Previous length is what? Because it's the first length, the previous length is 0\. Right? And then 2 as a string is f3. Right? And then, for my next element, my first byte, the previous length is what? 2 bytes. So, I'll store 02\. And then my...

5 is stored as F6. F6 is the AXE representation of 5 followed by FF. So this is how your ZIP list is stored. Now if you want to append anything to that, you would be writing over here, changing this and changing this. ⁓ It's all operation happening in the same block of memory that you have. ⁓ This is how your ZIP list would work. Now, advantages.

The advantage of the ZipList is there is a sequential memory allocation. You are not doing random allocations. So if you are iterating through the ZipList, ⁓ you are literally iterating through contiguous memory locations, which means you are very CPU cache efficient. ⁓ Number one. Second, there are no internal pointers. There are no random pointers here going here and there because each pointer in a 64-bit machine takes up 8 bytes. ⁓ Huge amount of memory gets wasted. You literally storing the offset within that, the bare minimum that you have to.

Sneha Mehra (00:23:18)  
then this is cache efficient and order one access to head and tail. This is the most common application to access the head of the list or tail of the list to push or pop. That's most common thing that you do with a list. So zero one head and tail access, order of one head and tail access. ⁓ But obviously there has to be, you cannot get best of everything. So disadvantage.

the insert and the delete operations, ⁓ they require reallocation, moving elements within that, obviously. So now imagine if you are adding something in between over here, you would have to ⁓ move ⁓ these two elements or these two parts after that and plug your things in. So you have to do mem copy.

of these bytes into the forward location and then write your entry over there and then update the total length and this part over there. ⁓ And in case you run out of your total allocated zipless size, you would have to do reallocation of a big one.

Right? That would be challenging. that's where insertion and deletion requires the reallocation and reshuffling of data and movement and moving elements within that. That would be required. Then, Ziplist size cannot be huge because first of all, getting such a big chunk of memories may not be always possible. Second, if you're having that big of a memory, why would you want to do that? That purely becomes an array. Right? So you would not want to have that. So that way, your Ziplist has to be short enough. It cannot be like very miniscule, but decent enough

100, 200 elements if you'd want to keep that depending on the amount of data that you're inserting. So that's why it's a configuration that you can run by. ⁓ This is how ZipList are implemented. But we still did not touch upon what are QuickList. We talked about ⁓ random element doubly linked list, we talked about ZipList. But we saw QuickList. Now what is QuickList? QuickList is pretty interesting. QuickList is linked list of ZipList. ⁓ So that's what it is. So we know

Sneha Mehra (00:25:08)  
linked list has its own advantage. ⁓ Right? Linked list gives us advantage of having dynamically allocate random places infinitely growing. But the challenge with doubly linked list was that if I do it for each element one shot or like one allocation for each element then my lot of random allocations would happen. So what do you do? You merge both the approaches. So ⁓ each element of your quick list is a zip list. Once that is full, you allocate one more and then you chain it like a normal doubly linked list.

That's what your quick lists are all about. So you still get order one head and tail access because you would store it at each node. You get cache efficiency because while you are iterating through the list, you are iterating through the zip list first. So this would be very cache efficient. Then you start adding through this, this would be loaded in the cache like CPU cache that would be efficient and so on and so forth.

This is compact yet extensible. So this is very compact. are not wasting a lot of space over there. But still you can add as many ziplist nodes as you want in the quick list. ⁓ And inserting and deleting elements. This is what the disadvantage is. Now imagine if you delete an element over here, ⁓ you can just allocate, ⁓ smoosh them together if you want to. Otherwise you may want to split the ziplist into two half. ⁓ And then chain them with... ⁓

like like a normal quicklist node right so inserting an element or splitting an element would require you to join the ziplist through your and basically through your quicklist nodes you would have to do that.

⁓ If you inserting in the middle, let's say this ziplist is full and you are inserting over here, you cannot. So either you move this element over here, but that would be worse because then you would have to do this rep, like then this would have a transcending effect. A better implementation would be if you are inserting in the middle and this ziplist is full, you split this ziplist into two half and chain them like a normal quicklist node. That's what actually Redis does. Right? So yeah, that is all about quicklist, ziplist and how Redis implements list in

Sneha Mehra (00:27:06)  
in general, you can go through quicklist.c, which is the source file, ⁓ and go through the source code and see how beautifully that they have implemented this. You have a basic idea on how it does that, how encoding happens. I would highly, highly, highly encourage you to check the Redis source code out to understand those ⁓ finer granular details if you wish to. ⁓ But this is how Redis becomes, ⁓ Redis acts extremely frugal while implementing.

something as simple as a list without relying heavily on normal classic implementation of a doubly linked list. Right? So yeah, that is it. That is it for this one. ⁓ I hope you found it interesting, amusing. Again, I would highly, highly, highly, highly recommend you to implement this thing on your own. I have not implemented in my Golang based code base. It was more of a theoretical overview of how it does that because implementation is a lot of grunt work. It would have taken a lot of time for us to implement it. But again, if you'd want to dive deeper into this, there is no better way

but to re-implement this part. ⁓ I have it in my GitHub issues list. If you find that thing still open, I would highly recommend you to implement this and contribute back to the code base. ⁓ So that is it for this one. I'll see you in the next one. Thanks, Atan.

—----------------------------------------------  
28

Sneha Mehra (00:00:00)  
Thank you so so so so much for enrolling into the course and congratulations for completing it. I really hope you found this course helpful and derived a solid value out of it. It would mean the world to me if you can spread the word about this course amongst your peers and friends. And I would be forever grateful if you can also leave me a testimonial. The link to a small form having few questions is in the description down below. Now that being said, to amplify your learnings,

I would highly encourage you to implement what we just studied, just like how we did throughout the course. Right. Apart from that, I would again highly encourage you to go through the register source code to understand those nitty gritties of that. ⁓ Or you can directly start contributing to DICEDB. Now, in order to do that, you can directly go through the open issues available on GitHub over DICEDB repository and pick the one that interests you the most and start implementing.

⁓ I'll be available to scope out features and implement the functionality. Just drop me a message over Discord, I'm always there. ⁓ So, looking forward to collaborating with you and make DiceDB extra special. Thank you once again for enrolling. Keep learning, keep growing.

—----------------------------------

27  
Sneha Mehra (00:00:00)  
So apart from supporting LRU eviction strategy, ⁓ Redis also supports LFU eviction strategy, which is least frequently used. ⁓ And as the name suggests, we have to evict the element that is least frequently used. So somewhere we have to keep a track of the frequency of access, like how many times was this element accessed? Then only you be able to tell, hey, this is the element which is least frequently used, so I would be evicting it when my memory is full.

Right? So ⁓ obviously Redis being very frugal in its implementation, it cannot, cannot, cannot just store normal integers and do plus plus every time, there has to be magical about it. Which is where what Redis does, it uses approximate counting. It does not use standard count plus plus. It does something approximate, which is a very space efficient counting strategy. Right? Okay. So the core idea of this is that we evict the, the core idea of this is

We saw in the Redis object where it had a field called LRU bits, which were 24 bits, right? So Redis says, hey, why to waste that? I can just use that to implement LRU. So Redis uses the same LRU bits, the 24 bits to implement LRU. Now a naive user would think that, ⁓ I'll just use the entire 24 bits to do count plus plus every time, what's wrong with that? So what's wrong with that is the fact that imagine if you have a key placed.

Imagine if you have a key placed and that key was accessed very much during the initial stages. Very much, ⁓ like really huge number of times. If it was accessed for a very, like very highly at the early stage and then it is not accessed at all. But given that initial access was really high, you are continuously keeping it there, ⁓ not evicting it. So that is wrong. ⁓ Which is where you have to decay depending on the time that is being gone by.

⁓ So, which is where the 24 bits, the 24 LRU bits are used to implement LFU and they are split into two parts. First is the 16 bit, second is the 8 bit. So, in 16 bit, it stores the last decrement time ⁓ and in the 8 bits, it stores lock counter. We'll look at both of them in detail.

Sneha Mehra (00:02:21)  
Okay, so first take a look at a simple one which is logarithmic counter. Now what it does is 8 bits, you'll say 8 bits I can just load 2 raised to it which is 256\. But my frequency would be much higher. I cannot just leave it 256 which is where the power of approximate counting comes in and Redis does not just do normal count plus plus it uses something called as a Morris counter to implement this. We'll take a very detailed look into its implementation. Right? And then comes the first part which is first 16 bits. Now 16 bits is that

you would want to decay ⁓ the count, the access count ⁓ depending on when was my key last accessed. You will think that A isn't this LRU. No, no, no, no. We are still evicting on the basis of frequency, but as time progresses, you should be decaying the count so that if a key is not long enough accessed, then you would be like that would become a potential candidate for eviction.

Right? So which is what it does into this. First 16 bits for time, and then 8 bits for actual counter. Right? Okay. So what does it store in 16 bits? 16 bits is such ⁓ less. What would it store? So in 16 bits, it stores Unix time in minutes. Not milliseconds, not seconds, but in minutes. And it just takes the last 16 bits of it. Right? Because that's the amount of space that you have. It cannot take more than that. So every time a key is accessed,

If the strategy is LFU, it would store the Unix time in minutes in the first 16 bits ⁓ of my LRU bits, my 24 bits LRU field that I have. The first 16 bits are the Unix time in minutes. Okay. Then let's go in depth of Morris Counter, the eight bits that we're talking about. So now here, ⁓ excuse me. So here, instead of storing the actual value, because you have only eight bits,

8 bits is 2 raised to the power 8 which is 256\. But instead of storing the actual value, what we store is we store the logarithm of it, which is why it is called as logarithm counter. So, ⁓ I store, so V is the value that I store and N is the actual value that I know ⁓ has existed. We are not storing it anywhere. All we have is V with us. ⁓ So, what we store is V is equal to ⁓ logarithm of 1 plus N. This is what we are storing. Now, here,

Sneha Mehra (00:04:45)  
Because it's a log function. It's a very beautiful function. Why? Because for a smaller value, you see a significant growth and then the growth tapers off. There's a classic log ⁓ graph or rather log curve. So why is this important? This particular distribution or this particular pattern, why is it important? For keys that are accessed less frequently.

You cannot just say frequency of 1 is same as frequency of 2 is same as frequency of 5\. ⁓ These are significantly different. These are significantly far apart. So that's where 1, 2 and 5, even a small change is significant. But if I talk about keys or frequency like 10,000, 10,100 and 11,000, there is not much of a difference. So, ⁓

we need higher precision for smaller values ⁓ and we can lose precision for larger values because there is not much difference between 1 million and 1 million 1000\. All that is equal to 1000 access but in comparison of 1 million there are peanuts. ⁓ But if you think about 1, 2, 5 for smaller values having a higher precision is very important which is why we use a log function over here which is v is equal to log of 1 plus n. We don't know what n is. ⁓ All we have is v by the way.

Okay, so that's the idea. That's what we are going to store. So the key challenge over here is that we just cannot increment V just like every time we see we cannot just do plus plus over here because number one, ⁓ the range of V is just eight bits, it means zero to 255\. After that, you cannot go beyond that. It will circle back to zero, ⁓ right? And if we are storing logarithmic counter and because we are storing the value V is equal to log of one plus and if we are trying to approximate the actual count,

it will go to a very huge value which is not true, right? Which is not appropriate. That's not a very good estimate. So, that is where how what we need is we need a good way to determine when to increase V. Because for every event we cannot increase V. We need to find a way by which we can say, hey now it's time for us to increase V. So, what do we do for that? So, a good metric ⁓ is always cost. Now, here the cost is inaccuracy.

Sneha Mehra (00:06:55)  
So, for example, what we can do is we can find the inaccuracy that we might incur if we make the wrong decision. So, for example, ⁓ if I make a wrong decision, then if I do v equal to v plus 1 at this stage, what is the cost that I am incurring? That hey, I should not have increased, but because I did, my final counter value is this, but it should have been this.

It is not difference of 1, it is the actual cardinality. The actual cardinality would be what? ⁓ N. ⁓ N is what? ⁓ N will be equal to, if you look at this equation, ⁓ N will be equal to ⁓ e raise to power v. It would be e raise to power v minus 1\. ⁓ That would be e raise to power v minus 1\. That is a huge number. Huge, huge number. ⁓ So, ⁓ if you do not count e, if you take it by the base 2, would be 2 raise to power v minus 1\. That is a huge, huge number.

So, you have to be very careful on when you are incrementing V. So, that is where you can decide on incrementing V depending on the cost. Now, the cost is ⁓ as the difference ⁓ between N of V plus 1 and N of V increases. So, if I choose to increase what would be my final value, if I not choose to increase what would be my final value, ⁓ if the difference is larger, ⁓ I should be less probable to make that switch.

If the difference is smaller, I should be highly likely to make that switch. As simple as this. So we work with simple probability. So here, ⁓ if the jump is smaller, from nv to nv plus one, right? So which means, ⁓ if I change v to v plus one, if my estimated value is nv and nv plus one, what I have to do is, I take the difference of this. If this difference is larger, then I should be less likely to make that switch.

if the difference is smaller I should be more likely to make that switch which means V1 to V1 V plus 1\. So, for example if I moving from 10 to 20 my D is equal to 1 upon 20 minus 10 is equal to 1 by 10 equal to 0.1. If I am moving from 100 to 5000 then D is equal to 1 upon 5000 minus 100 is equal to 1 upon 4900 is equal to 0.0002.

Sneha Mehra (00:09:13)  
So here you look at that smaller the difference, I'm highly likely to make that jump because the cost is very less for me, like the inaccuracy is very less, right? Because it's indeed very less. But if I do this, the chances of me getting wrong because it's logarithmic function, the estimation that I'll be making of my cardinality, sorry, not cardinality, if I'm going to make the estimation of my count that I have, right? That would be very ⁓ high.

Given that it would be significantly high. So that's why I have to be very cautious when I deciding to make that jump. So given that we have D as this metric. So how do we do this? This you can treat this as a probability. That probability of you making the jump. So now if you look at this, what would it look like? The idea is pretty simple. We pick a random number given that we have D. ⁓ So at every stage when you are getting something, when you are getting to do count plus plus, you first check.

that hey if I change v to v plus 1 what would be the final value if I change v if I keep v as is what would be the final value you take the difference of this you do 1 divided by this difference that would give you the probability you take that and you pick a random number ⁓ if the distance or rather if d is greater than r you make that switch ⁓ if d is otherwise you keep it as is. So, which means that if you are highly likely

If you are highly likely to make that switch, make that switch.

If you not likely, don't make that switch. That's the idea. Right? ⁓ So, this is the idea of Morris Counter. ⁓ Now, if you look at the graph, ⁓ if you implement this Morris Counter on your own and if you plot the graph between the actual value when you do count plus plus every time and the approximative, you can see it very beautifully following it. And every time you see this, this step happening, ⁓ which is where it flipped the coin or rather it chose a random function. ⁓ And so the random number was within the limit and which is where it chose to increase. ⁓ Right? And which is what you see over here.

Sneha Mehra (00:11:17)  
it's a step like increase. So, this entire duration the counter would remain the same and once it hit that limit or that probability kicked in that random number the one that was chosen was more than D or rather D was more than the random number we chose to increase it and we move forward right. So, what Morris did it's there in the paper what Morris did is instead of using just log of like V is equal to log of n plus ⁓ log of n plus 1\. So, the function that we discussed was V equal to log of 1 plus n.

Right? So instead of that, it shows it to the base two log of n plus one divided by log of base two. This is what it is storing in. It's a ⁓ like after doing a lot of tests, he came to this conclusion, which is very similar to what we started with. Instead of using basically natural logarithm, it used like log n plus one divided by log two. Just a small tweak there. Okay. So what Redis does, ⁓ and we'll also take a look at the source code by the way. So what Redis does, it saturates the counter to one million.

for sure because ⁓ what are the odds that a key is accessed more than a million and you want precision over this. So, saturates at 1 million. It decays the counter every 1 minute. Now, this is where we are talking about decay. It means that if it has been 5 minutes since my key was accessed, my V not N, V would be reduced by 5\. ⁓

This is for the decays. If I am decaying by 1 minute every time or rather every 1 minute if I decaying it by 1, it implies that my v would be we do not have n. We ⁓ do not have n at all. All we have is v which is 8 bits that we have. Right? Now, you will say how do I do decay? The first thing that would come to your mind is say let me run a thread that would do this decay. No, you do need to. So, what do you do? Is you adjust the counter while accessing it. So, whenever you are accessing it,

you see that, hey, what is the time that you access the key last? How much time has elapsed? Get the difference of it and reduce and adjust the counter accordingly. That is exactly what you need to do. You don't need to run a thread onto doing this. That's a very naive way to do it. You don't. While accessing it, you can check it. Right? And why is this an exponential decay? Because changing v,

Sneha Mehra (00:13:25)  
from let's say 256 to 254 it would be changed from 2 raised to 256 to 2 raised to 254 that's a huge change it's literally half of that and then half of that and then half of that so that's why we say that every time with every minute your ⁓ value your actual value the actual frequency would be halved every time

which is why because we are reducing V by one because it's a logarithmic function, the effective value N would be reduced by half because you are doing a runtime evaluation of that to see what's the total value is like what's the big value is. Right? Okay. Let's quickly take a look at the source code and see this in action.

So here what I have is the file evic.c which we saw during LRU implementation. So if you scroll at the bottom half of this, you can see LFU implementation. So this is where LFU implementation starts. Here we say that, yeah, we know that we have 24 bits for LRU. We use that same bits, ⁓ same things for LFU. 16 bits to store last decrement time, eight bits to store log c, which is a logarithmic counter.

which is what it has written over here. Now, ⁓ the best part ⁓ is in the code. ⁓ Here you see, LFU get time in minutes. ⁓ This is what we are talking about. gets ⁓ Unix milliseconds, Unix seconds, no, Unix minutes. gets that. So server.unixtime slash 60 and 65535, which is the least, like the most significant, sorry, the least significant 16 bits of ⁓ Unix time in minutes, right?

So it is getting that it is using this to store something like sorry to store the last decrement time. Then the second function we come to is LFU time elapsed similar to how we found for LRU time elapsed we do that same thing. So what's the time elapsed for LFU? Time elapsed for LFU is you get the now time ⁓ if now is greater than last ⁓ like last decremented time. If now is greater than that then the difference is now minus LDT.

Sneha Mehra (00:15:26)  
Otherwise it is 65535 minus LDT plus now the exact same logic that we applied for LRU that same logic applies here because you want to find the elapsed time like circling back. ⁓ Okay, that's it. ⁓ Then this is the function where we do logarithmic increase. ⁓ Here you can see rand function being invoked and you do count and you base well and all of that over here that we just discussed. ⁓ If r is less than p we do count plus plus. So if p is greater, v is greater than r that we found out it's exactly that implementation.

the exact Morse counter implementation you can see it right here. Okay, then if we scroll back, scroll down, we see this another beautiful function called LFU decrement and return. This function is the one that adjusts the counter that I was talking about. You don't need to run a thread that does that, that minus minus every time or that does that decay every time. This one what it's doing is, it is doing LFU decrement and return for an object. So it gets the LRU, ⁓ the ⁓ LFU is also set in the LRU bits, it takes the

like it's ⁓ let shifts, sorry, it right shifts by eight, which means the most significant, sorry, the least significant right eight bits are gone, which is log C, your logarithmic counter. What you get is the significant, ⁓ is the 16 bits, which is the last decremented time. It got that. Then it took the counter. Counter is what? The counter is ⁓ the LRU and 255, which means the least significant eight bits. So that is counter. So it got LDT and counter both. Now, here it is finding the number of seconds.

sorry the number of minutes that has elapsed. So, number of periods is equal to server dot lfu decay time assuming it is 1 minute. ⁓ If this is set then use lfu decay time ⁓ divided by lfu decay time. ⁓ sorry here ldt last decremented time you get the time elapsed. Time elapsed

you get this divided by decay time. If you are having a decay time of 1 minute, so every 1 minute it would ⁓ reduce it by 1\. So that is what we are doing. We are finding the number of seconds elapsed. So decay time is typically set to 1, there is a default configuration. So it's typically the value that you want. It's the idle time that you have that gets over here, but you can change it. That's why it's just making it ⁓ cognizant of the fact that you can change it and tune it the way you'd want. So it gets the number of periods. It changes the counter.

Sneha Mehra (00:17:44)  
to if number of periods are any then your counter equal to counter, sorry counter equal to number of periods greater than counter than 0 which means if the number of periods are greater than counter then your counter value is 0, right? Because your counter is 0 now, because decay is complete then you would not be left with anything. Otherwise it is counter minus number of periods. ⁓ So whatever the counter value is minus the number of periods and return the counter. Here one key thing to notice it is not making changes to the counter. ⁓

it is simply returning you the effective counter value. That's all it is doing. Right? So that's the idea behind it. So you can leverage this function for a lot of use cases. So we'll take a look at this. So let me first go through db.c file to tell you where it is updating it. So here you have a function called updateLfu in which it is actually using the same function, ⁓ Lfu decrement and return. You get...

for the object, you get decrement and return. It is not actually decrementing and setting it. It is giving you the new counter value for that, the effective counter value for that. Then you are incrementing because you accessed it, right? So here you are getting the updated. So depending on how many time or like how many periods went past from the last access time, you get that, you decay that, you get the effective value of counter, and then you ⁓ have to...

you have to do increment because this is the time ⁓ when you access it. So update LFU when the object is accessed. This is the function that is getting ⁓ invoked upon every key access. If your eviction status is set to LFU. And then you do log ⁓ LFU log increment counter and you set over here, this is where you are setting value of LRU is equal to which is setting in the object. ⁓ LRU get time in minutes.

left shift by eight, which means you get that minute and set the most and you left shift by eight, which means you set the most significant 16 bits over there. And then, or with counter, which is your log counter that you have. ⁓ This is where it is reconstructing the LRU bits.

Sneha Mehra (00:19:49)  
⁓ And then every other logic remains the same, nothing fancy over there. But like every other thing is about normal eviction that we did with LRU, you can implement it in the very same way. But here you see the exact LRU, sorry, the exact LRU implementation using Morris counter, the logarithmic counter, how that bit moves. Such a beautiful piece of implementation. So, so optimized ⁓ on the space and everything. Right, so what I've done ⁓ is I've created a GitHub issue.

on implementation of LRU. I will link that in the description down below. If that is unassigned, I would highly encourage you to pick that up and implement it. It's such a simple algorithm to implement, right? But it does such a fabulous job, such a fabulous job with such minimal memory requirement. This is where the power of approximate algorithms comes in. It just uses a fraction of memory to do gigantic things, right? Without losing a lot, like without inducing a lot of error. So yeah, great. I hope you found it interesting. I hope you found it amusing. That is it for this one. Thank you so much for watching.

—------------------------------

20  
Sneha Mehra (00:00:00)  
In this video, we will take a look at graceful termination. So like always, here are four terminals that we have on the top right, we have a normal Redis server running. Now let's see what happens when we press control C over here. So right now the state of my Redis server is ready to accept connections, right? Now if I hit control C, see what happens. So it said that received SIGINT scheduling shutdown, user requested shutdown, saving the final RDB snapshot before exiting.

DB saved on the disk and then Redis is now ready to exit. Bye bye. So this is what is called as graceful shutdown. So the idea here is whenever you kill a particular process, your operating system sends a signal. It's a sigint, sigterm. There are many such signals like this ⁓ and you have to gracefully handle it so that when your machine is going for a shutdown, for example, if you're rebooting a machine or if you are

shutting down a particular machine or someone wants to explicitly kill your process, what happens is your operating system sends you a signal. You have to trap that signal and write your own handler for that. This way you can ensure that if you are shutting down your database, the existing user requests are handled before it actually shuts down.

⁓ So this is the idea behind signal handling. And this is what happens across every single web server database process. Any process has to handle like ⁓ for a good behavior of that process, it needs to handle these operating systems signals so that it ensures graceful termination of the processes. ⁓ So before we implement that, let's first quickly understand what this is all about. ⁓

Okay, so what happens when you press control C? This is what happens. Your database actually shuts down. ⁓ And whenever we are shutting down the database, it always, always, always has to be graceful. Because for example, in case of Redis also, when we are shutting down, it is creating a final RDB snapshot of the database and persisting it on the disk. Even your database, anything that you are implementing has to take care of that. For example, some final cleanup, deletion of files, dumping of registers and whatnot.

Sneha Mehra (00:02:15)  
⁓ So there are ton of things that needs to happen when your database is shutting down gracefully. Obviously there are chances where you are just abruptly shutting down your machine and your database would not get that chance. Perfectly fine. ⁓ For example, if someone literally plugged out your server ⁓ from the electricity supply, it would crash. And in that case, none of the signals would execute. ⁓ But that's fine. But whenever there is a chance of...

you handling those events gracefully, you should definitely do that. And this is what is called a signal handling. So what are signals? Signals are the notifications that are sent to you by the kernel to the process. ⁓ So whenever some important thing happens, it notifies it through a signal. So that is where every process needs to trap that signal and do whatever they would want to do. ⁓ Now for our database, we do that to ensure that we are doing a proper closure. For example, you bound, like for example, your database server

was running on port 7379, at that point of time, you got a SIGKILL or SIGINT or a SIGTERM. In that case, you would want to unbind that particular socket so that some other process can use that particular port number. ⁓ That is one. Then you might have to flush the file. You might have to flush the buffers. You might have to save. You might want to take the final snapshot. You might just want to send an email, for example, just giving a random example. So anything that you would want to do, you can do that at that time.

Second is to ensure data integrity and consistency. So for example, if you have an in-memory buffer and you ⁓ would want to do a graceful shutdown, then it's important for you to store that buffer on the disk so that your data is not lost. ⁓ So data integrity, consistency, all of that needs to be checked when you are gracefully shutting down your system just to ensure that there is nothing abruptly written on the disk or anything related to data loss. ⁓ So there are... ⁓

Operating system, every single operating system has a different, like has some common set of signals ⁓ that it supports, but every system has its own set of signals that you would want to handle. So we'll walk through a few very important one and see how Redis does it so that we can re-implement it in DiceDB. Okay, so the most important ones that any database has to handle is SIGTERM and SIGINT. These are the ones that triggers that graceful shutdown. So whenever you are...

Sneha Mehra (00:04:35)  
⁓ doing a control C or whenever you are doing a kill, ⁓ like using the kill command, you are trying to kill the process, sick term and second, ⁓ they are the, like these are the signals that is passed from ⁓ the kernel to the process, right? So this tells Redis to shut down gracefully. ⁓ does, and what Redis does, it does cleanup of whatever it wants to, like files, temporary files and whatnot. Then the files are flushed.

socket is closed and data is persisted. Almost everything is configurable, but data persistence definitely is configurable. So we just saw when we press Control C, it ⁓ typed that a final snapshot is being persisted on the disk, right? That was a one time final snapshot, all of the keys being dumped into the RDB files and persisted on the disk, right? So this in end, it also ensures that for example, if there is an existing command running, it can wait for that particular command to wrap up.

and then it would terminate. So for example, if a client because Redis is single threaded, there would be always one command from one client which is actively running.

while that is happening, you would not want to kill that particular thing because that might lead your data into go on inconsistent state or some weird thing might happen. Plus our client would get a very poor experience. That's what you would want to ensure is the fact that when your server is executing something, you would want to wait for this command to complete and then you would shut down, right? So these are sick terms and sick ints that we would want to handle. There are other in which what happens will first, let's first talk about the signals, ⁓

important ones are SIG SEGB, SIG BUS, SIG FP and SIG ILL. sorry, SIG ILL, So, ⁓ SIGIL. So what happens here ⁓ is the final one is where you are executing an instruction that does not exist on the CPU. ⁓ Very well chance that for example, you do some reason.

Sneha Mehra (00:06:32)  
the machine level code that is generated by your programming language environment is firing a CPU instruction that does not exist. Then this signal is raised. ⁓ SIG FPE is the floating point error. So if you're doing division by zero, if you're doing integer overflow, anything SIG FPE is raised. You might want to handle that because that is a case where your database is not handling those edge cases really well. Then SIG bus is when you're accessing an address that does not exist. So SIG bus is basically the bus error.

and when you're accessing a particular memory and that memory address itself does not exist because it's too much to like, for example, if you are having a 4 GB RAM and you are accessing something that is 8 GB, right? Then like at that particular location that would not exist at that time you get a bus error.

and SIG SEGV is where you get segmentation fault. So when your process is trying to access the data, so address is valid, but it is trying to access the location that is owned by some different process, you get segmentation fault. At that time SIG SEGV is raised. Now, whenever any of this is raised, any of this, what Redis does ⁓ is it basically, these are the edge cases if you would want to put it.

So it basically takes ⁓ all possible information and flushes it to the disk. For example, stack trace, ⁓ state of the register, state of the client, ⁓ everything goes onto the disk so that if anyone would want to debug, they would have enough information to debug that particular situation. ⁓ So these signals are there.

So that there's edge cases that you would not have handled that led to your operating system raising this particular signal and giving a process that interrupt, you would want to do that so that your database becomes resilient by the day. And then there are two more, SIG HUP when you would want your process to run in background or when you are...

Sneha Mehra (00:08:24)  
Basically when your terminal session is gone to some other process, SIG HUP is raised and SIG pipe is for piping. When you're writing to a pipe that does not exist. But there are ton of details that you can read about it. But what Redis does is it ignores both of them because you want your database to be continuously running. So it ignores both of them, Enough of the theory. Let's jump into Redis source code and see what it actually does so that we can implement it on our own.

Okay, so now here what I have is I have opened up the actual ready source code and here the first thing I would do is I would do in it server or rather let we know that the two signals that we would want to handle SIGINT and SIGTER. So I'll just search for. ⁓

SIGINT, we'll find something. So here I see lua.c does not seem relevant. Then I see sig action eval.c kind of relevant. Some action is there, some action is set over here. Let us look for sig action. What do we get? We get some file redis log, some server. Server.c is where everything started, isn't it? So this would have some information. ⁓ here we get setup signal handlers, right? And if I search for setup signal handlers, it should be there in server.c. ⁓

Here we get, here, init server. Init server is the first thing that your Redis invokes, right? And now, why did I take you to this path? This is a secret source to ⁓ read and understand any open source code base. And the way we started with the hint, it was sigint, right? We wanted to know how Redis does that.

the only word we had was SIGINT. We know that something to SIGINT has to be there. We took SIGINT and we traced this particular part by doing normal heuristics search, search and search, right? So this is, I took you through this part so that you understand in case you don't know up until now that how to navigate through a gigantic code base, right? This is what I've been using.

Sneha Mehra (00:10:16)  
throughout my life. I don't use any sophisticated tool to be really honest. This simple thing helps us understand the flow and you would obviously, ⁓ at the end you would be able to connect the dots. ⁓ Right? Okay. So, we reached in this point and here we see what it does.

Here, because it's C language, ⁓ like this is a particular syntax that you would have to write, ⁓ signal and then you write which signal you would want to handle and what's the handler. ⁓ SIG IGN is ignore and ignore. And then you invoke setup signal handlers and in setup signal handlers, you see SIG term and SIG int being passed ACT. Now some handler would be here. If I go through this, some SIG action, it would set and what not. It does a bunch of stuff there.

⁓ So here we saw how it does see here crash log enabled, SIG-SIGV, SIG-BUS, SIG-FP, SIG-ILL ⁓ and SIGABOT. ⁓ So the signals that we just spoke about, these are the place like, this is how they are handled in the source code. ⁓ So we understand how Redis does that.

how to navigate through the code base. We exactly saw where to find that part. You can obviously navigate through the source code and understand it in much more details. But now it's time to implement, right? We know what we are supposed to do. So what we would do for now ⁓ is we would keep it a little simple. So here what we do, or here what we want to do is we would want to somehow listen to the interrupt signals coming in, ⁓ SIGINT and SIGTERM. Whenever we get it, we would want to execute something that this is our ⁓ end job.

Basically, this is our shutdown phase. For example, let's say ⁓ like how Redis does. Redis basically does persistence, the final RDV backup and it persisted to the disk, which is exactly what we also want to do.

Sneha Mehra (00:11:58)  
Let's say we call it the shutdown function. So in the events.go file, I've written a function called shutdown. And this function ⁓ is invoked, or this function internally invokes ⁓ bg rewrite aof, the command that we implemented, right? And what it would do is it would take the in-memory hash table that we have, it would dump it into an aof file, which is append-only mode file.

Right? ⁓ Or it like that's the idea. We would want this to execute ⁓ whenever we receive an interrupt, a sigint or sigter. Now, that's where we would want to have that setup. But how do we set up? So, in Golang, we have something called as channels. The way we can understand or rather the way we can ⁓ basically listen to interrupts or we can listen to signals is through a channel. So, here what we are doing is we are creating a channel whose

can accept objects of type os.signal and we are registering it to listen to sigterm and sigint. ⁓ So, whenever there is sigterm or sigint, I would get a message or I would get an event into this particular channel. ⁓ This is what we are doing over here. And then what we would want to do is, we are spinning up two go routines now.

So we are spinning up the first go routine for a normal Redis engine that we just wrote. Earlier it was the main, the only routine that we had, the only thread that we had. But now we are creating a go routine and we are running it over there. And we have another go routine, which is ⁓ there to handle the signal, which is wait for signal. And you'll say, hey, we didn't, ⁓ or basically isn't Redis single threaded. Our actual execution is still single threaded. Our actual execution.

which is this run async TCP server is still single thread. It would read one command at a time, execute it and then pick another one and then pick another one. We are not doing something like this that we accept multiple commands and for each command we are spinning up a new thread. We are not doing that. It is still single threaded, right? But what we are doing is we are creating two separate go routines and then waiting it over here for the completion of it, right? This is what we would do with wait group, okay?

Sneha Mehra (00:14:01)  
Now let's take a look at waitForSignal because this is where the magic is happening. Now waitForSignal accepts a wait group. We are marking it as done, ⁓ deferred done. So when the function would return, it would do wg.done so that your main function can exit. And then SIGs. SIGs is the channel that we created. Now this, we are trying to extract something out of the SIGs and this is a blocking call.

⁓ So, until we get a message or until we get an event in the channel, we would not be moving forward. So, this is a blocking call. Our program control flow would come over here only when there is a signal. ⁓ So, what would we do when we get SIGINT or SIGTUM? We would want to shut down. We would want to exit.

⁓ the database, we want to shut down the database, we want to kill the process, which is where we are writing os.exit0. So with exit code zero, or sorry, with status code zero, we are exiting it, right? Then, but before we do that, we would want to trigger a core.shutdown. This is the shutdown function that we just saw. So just before we die, or just before we close our server, we would want to do a graceful shutdown, which means whatever we have, we would want to serve that, right? So these two things done. Now,

This is where this would happen that whenever we are doing control C, it would do one final time persistence of it, ⁓ BG rewrite AOF, and then we are exiting zero, right? Problem solved. But we would want to handle one more case. That one more case ⁓ is that it should wait.

for the existing command. If there is any command which is being executed at the moment, we would want to wait for its completion, which is what we are doing over here. ⁓ Now, how are we waiting over here? The plan is pretty simple. The plan is in case when the program flow would come ⁓ beyond this over here, ⁓ only when ⁓ I've received the signal interrupt, up until then nothing would happen.

Sneha Mehra (00:15:58)  
⁓ So, which is what we are doing is we are running an infinite loop. So, in case now this infinite loop what it is doing is it is waiting for your engine status to turn into waiting. So, if the engine status is busy I would want to keep this for loop running forever. So, up until the time your engine status is busy I would want to continue. But where are we maintaining this? It is maintained in E status object. Now, this E status is nothing but an integer value.

which ⁓ basically holds if my server is, or if my Redis engine is waiting or busy or shutting down. So busy is where it is like we wrote epol wait. Epol wait is a blocking call. This blocking call would end. This blocking call would return ⁓ only when any one of the file descriptors is ready for ⁓ an IO.

because this is a blocking call, which is where we would have to handle this case of graceful shutdown. So now this is a blocking call. Now here what would happen is if we go beyond this, it means that there is some file descriptor which is ready for an IO. So that is where we become busy. Before that, we are waiting, right? We are waiting.

for some file descriptor to be ready for an IO. That is a waiting stage and below that is the busy stage, right? But we would have to maintain this state somewhere. So that is where we are using the variable E status and which is where we would be setting the value. So wait for signal would be setting the value to shutting down whenever we do that. And it would be set to busy and waiting when the command is being like as and when the commands are being processed, right? Okay. So ⁓ as soon as we receive the signal,

The first thing we do is we wait in case, in case because now wait for signal is a separate go routine and your asynchronous server is running in a separate go routine, right? Now here case number one, when you receive a signal, your server ⁓ is in the waiting state, which means that it has not picked, like there is no file descriptor ready for an IO. It is in the waiting stage, right? So in that case, it is safe for us to

Sneha Mehra (00:18:14)  
terminate, correct? Because there is no one which is executing anything at the moment, correct? So it is okay for us to terminate. So in that case, what would happen is, here we are checking if E status is equal to equal to busy or not. If it is not busy, this forlook would terminate. It would come down. As it would come down, it would mark the status as shutting down. Invoke the shutdown and OS.exit. Problem solved. This is the happy part. So where your TCP server

is in the waiting state, not executing any command, it is okay for us to terminate and which is what we are doing. Because in that case, this for loop would never, like this for loop would not even iterate. So, ⁓ when ⁓ I, just to give you a gist on that, when we invoked the shutdown, when we got that particular signal, if my server is not executing any command, we can directly go and exit it. Pretty clear on that.

Then the next part. The next part is when we issued this particular ⁓ control C or a sign term or a sign ⁓ int and my server is executing something. So how do we know if server is executing something? That is where the server would have status of busy. ⁓ So if your server is executing something, which means after epol wait it is there. Then we have to wait for the process to terminate.

for the command to complete its execution. Correct? So this is where this infinite for loop would come in. So now here what we are doing is, as soon as we receive the signal, we first check if my server is busy, wait. If my server is busy, wait, wait, wait. And this would go on and go on until my ⁓ engine status is busy, which means your engine is waiting for some command to execute.

⁓ while it is executing it. At that point of time, it would be continuously waiting. ⁓ And this for loop would terminate when the status becomes back to waiting because busy is when your command is getting executed. As soon as command execution is done, we have to turn it to waiting. So, which means that this for loop would end. This is a separate for loop, this is in a separate go routine, which is your wait for signal go routine. Now, this for loop would terminate when

Sneha Mehra (00:20:39)  
your asynchronous server is done executing the command and has changed the status to waiting. ⁓ That is when this would terminate. And once that is there, we have to mark my server status or my engine status to shutting down. And then I have to shut down and exit, which is what we are doing over here. So if I just walk you through the source code, and now we'll also talk about one edge case over here, which is very fun one to talk about.

So here, what we are doing is, we'll start with that. Okay. So here, the for loop that we had, earlier we used to have a for, we literally had an infinite for loop, right? A literal infinite for loop. But instead of that, what we are doing is, we are checking if E status is not equal to shutting down.

If it is not equal to shutting down, then only you move in. Now this infinite for loop that we are the infinite for loop back then that we had, it was continuously waiting for connection to accept or any IO to happen and it would move forward, right? Which is what here E-POL wait was doing, right? E-POL wait was a blocking call. It was waiting for all the file descriptors, for all the file descriptor, sorry, for any one of the file descriptor to be ready for an IO and then it used to move past that.

Correct? So here what we are doing basically is that we are waiting for our server where it is not shutting down because if it is while your server is not shutting down, which means it is either busy or waiting, it would continue to move forward. This for loop would break when my server status is shutting down. So as soon as my wait for signal, my wait for signal ⁓ made the server status to shut down this, this line executed, then my for loop would terminate.

⁓ When my control flow would read back at the top of the for loop at that point of time, it would terminate the for loop. ⁓ Okay, that is done. So then we change this particular for loop. And now this is where our epol wait would come in. But just before we do that, I would want to just go ⁓ here in the middle, just after epol wait, because before epol wait, we are waiting for any file descriptor to be ready for an IO. After epol wait, we have some file descriptor which is ready for an IO.

Sneha Mehra (00:22:49)  
Correct? So, before E-POL wait is your waiting zone and after E-POL wait is your busy zone. Right? So, as soon as my control flow came over here, it means ⁓ that I am in the busy state. Right? Now, which is where you would want to set your engine status to busy? ⁓ This is where we are setting it. I'll just come back to it in a minute. And then, where would we set to wait? ⁓ Once my execution is complete, ⁓ I would be changing my engine status to waiting.

So, once the execution is complete, I am changing my status to waiting, ⁓ but as soon as my E-Poll wait succeeded, I am changing my status to busy. ⁓ Correct? So, as soon as this becomes waiting, ⁓ then your E-Poll wait, in case your wait for signal is waiting,

on it being okay. this is where see here the command just a minute here the command execution is complete and then I setting my status to waiting right. So, which means that in case there was anything that was running my engine status would have been busy continuously running then it would have changed to waiting which means this for loop would have broken it would have come down.

then change the status to shutting down, core.shutdown and os.exit. As soon as ⁓ my server is free, ⁓ as in my server is not busy, moved into the waiting state, immediately, what I could do is, I could move this thing forward. In case I received a signal, I could move this forward and I can do an os.exit zero. Because my ⁓ existing command execution is complete. I can move past that.

⁓ So, this handles that, but there is one edge case that we have to handle that one edge case that very notorious. So, now, here imagine a situation where your server status is busy and then what you are doing is you are setting it to shutting down. Now, just imagine that if your server was not busy when you received the interrupt, when you received the signal. ⁓

Sneha Mehra (00:24:57)  
Right? Your server was not busy. Right? Which means it was in the waiting state. But ⁓ just before you could execute this shutting down, where you could change the status to shutting down, what if your server becomes busy again? You don't want that to happen. You really don't want that to happen. Because like you ⁓ know, like after this...

line like you are written this particular for loop to ensure that if it is in the busy, you are continuously waiting, right? So you're waiting for your like you're waiting for the server to complete the execution. Then you would do a shutdown. You're waiting on that. So what you have to ensure is ⁓ once you are done waiting ⁓ one time, then you are allowed to shut down. But imagine if

between here, so before it could execute, but before it could change the status to shutting down, if before this, your server again went to busy state, and now why would that happen? That ⁓ it went into waiting, it again circled back up, executed the epol wait, and then it got something, and then it started executing, and then this thread got scheduled. And then what could have happened is, here your server accepted the request of one of the client.

one of the client and then you are invoking shutting down and then you shut down then OS exit zero, which means this client existing command execution is going to be halt or it's going to break, which means, ⁓ which is exactly what we don't want to do, right? We don't want to have any existing client connection to abruptly break, but this edge case will break it, which means just before you could set the status to shutting down, ⁓ if you accepted

any of the client and started executing the command, you don't want that to happen. Which is where you have to ensure ⁓ that it should ⁓ never ever be possible for a server to go back to being busy ⁓ after shutdown is happened. Understand this. ⁓ Once the shutdown has happened, you cannot go back to being busy.

Sneha Mehra (00:27:19)  
You can only make your server busy only when it is in the waiting state. Otherwise, you cannot. ⁓ So which is exactly what we would do, ⁓ we would be handling now. So here, this is what we spoke about, that your server is running an infinite loop over here so that, ⁓ sorry, it's running a finite loop which would break when your server status is shutting down, where your engine status becomes shutting down. ⁓ So then. ⁓

What we have to do? We know where we marked my existing status, where we marked my existing status to become waiting, ⁓ right? So it was busy, ⁓ command execution done, we are making it waiting, ⁓ right? So when would we want to make it busy? So this is where we would be handling that edge case. So here instead of directly setting, ⁓ because imagine if I directly set over here, if I directly set engine status to being busy, what could happen?

So if I set my engine status to busy, in my wait for signal state, imagine this case, where ⁓ I am done this, it has set ⁓ my status to shutting down, and then that command got executed there because your server could like your asynchronous TCP server reached till this point, and then it said to busy, and then busy which means it would move forward and it would start executing.

You don't want this to pass. You don't want your server to go from shut down to busy ever. Which is where we are doing is we are using compare and swap atomic instruction. These would handle things atomically, which means that what we are doing is we are changing the engine status to, we are changing the ⁓ engine status to busy if the current value is waiting. Otherwise we cannot.

⁓ So, this transition, so if we are ever in the state, ⁓ ever in the state that we are ⁓ in the, ⁓ let's say if we are ever in the shutdown state, we don't want to go back to busy. ⁓ So, here what we are doing, if we are in the waiting state, then only we can become busy, otherwise we cannot.

Sneha Mehra (00:29:29)  
⁓ So here what we are doing is we are doing compare and swap each status if it is waiting then become busy. In case, and now what compare and swap returns? It returns true if it ⁓ was able to set the new value, otherwise it would return false. ⁓ So if waiting to busy, ⁓ my engine status became busy after this, which means that existing status was waiting and it could become busy, it would return true. Then all good, that's a happy path. But what's the sad path?

that if my engine status was shut down, say my engine status was shut down and I was changing to busy, this would return false. So if not atomic.comparant swap, which means my engine status is shutting down, I can directly return because your current status is shutting down. So you don't want to accept any more connections or you don't want to even proceed further, you can directly return. Now you'll say, what about

E-POL weight succeeding. See, here only E-POL weight succeeded. We did not read from the socket at all. ⁓ E-POL weight implies some file descriptor is ready for IO. This does not mean that we have read the value from that particular socket. So if we even terminate it over here, there is no command which is being executed and you're terminating in between. That's not gonna happen. This ⁓ one line,

is going to ensure the correctness of the logic. This is the edge case. This is the beauty of concurrent programming and amazing one. Like when you would write it on your own and you would ⁓ make your system come into such situation and you would see that breaking and then you would find this particular fix, which is why atomic, ⁓ basically atomic instructions really important like compare and swap. These are the tools that would help you build a very robust concurrent system.

Right? So just with these two threads, what we are trying to do is we're trying to do just graceful shutdown and all of this we encountered. Right? Okay. But how do we test it? So just before we do that, we skim through this entire code. We skimmed ⁓ async.tcp file. We went through engine status. We went through wait for signal. Then we saw how to, ⁓ how we change our infinite for loop to a finite for loop. And then we did a compare and swap to do busy. Once execution is done, we are changing it to waiting.

Sneha Mehra (00:31:48)  
⁓ So, as soon as DCP file done, events file is the shutdown that we are doing. So, here as soon as ⁓ that would happen, as soon as the shutdown would happen.

then you are invoking code or shutdown, which means this is where the actual save would happen. And then you're doing OS or exit zero. So once your graceful shutdown is done, you can do ⁓ an exit zero and shutdown is where you can write all sorts of things that you would want to do when you'd want to shut down your database. Right here I've just done this, but you would want to typically close your socket, flush your files, clean up your buffers, ⁓ delete temporary files if any, all of that would go over here.

Okay, and then just to test it what I have done is I have added a command called sleep. I'll just quickly scroll through the ⁓ sleep command and what it does it it just takes one argument which is the number of seconds it want to sleep. It just to test it does not exist in radius it just to test so that we can test the correctness of the system that any existing command which is running your server you press ctrl C your server waits

for this command to complete and then it terminates, right? So, I just implemented a simple sleep command and what it would do is it would simply go through this ⁓ and whatever the integer value is, it would just do a sleep for that duration, nothing much, right? But this is what we would be using in demonstration purpose. Okay, so let's jump right into a very quick demo of graceful termination. So, on the top left, we have our own implementation running go run. ⁓

main dot go, ⁓ my bad, go run main dot go. ⁓

Sneha Mehra (00:33:24)  
Here what would we have is our server is running on port 7379\. If I press control C, we see something else happening. We see rewriting AOF file. Earlier it used to just kill itself up. But now it is doing rewriting AOF file at dot slash dice master dot AOF, AOF file rewrite complete and then it is exiting. Something like this is the, ⁓ basically this is the signal handling that we have done, right? So we are invoking, when you press control C, we are invoking this particular persistence that, hey, go ahead and save it in the AOF format.

Right? Okay. Now, what we would want to test, we have tested this normal signal handler. Now, let's see, I fire this command, call dot slash redis ⁓ cli minus p 7379\. ⁓ Now, here I invoke command sleep of 10 seconds.

⁓ So it would return after 10 seconds. Now if I press control C, nothing happens. But as soon as 10 seconds would complete, see your server is not shutting down. When this execution would complete, then it would shut down. See, as soon as 10 seconds are over, then ⁓ it received that instance, it basically received that interrupt, it rewriting AOF file over here and then AOF file complete and then your server shut down. ⁓ So here we clearly see that how we send a request. ⁓

and we press Control C, it waited for the existing command to finish. ⁓ Once that command completed, ⁓ then only it shut itself up. Up until then it was just waiting. ⁓ Right? See, this is exactly what is important when we want to do graceful termination. ⁓ Right? We are not just letting your client connection get abruptly terminated. We are waiting for the existing command to complete and then we are shutting down our server. ⁓ And yeah, this is it about...

doing graceful termination and signal handling. Right? So, here again just to reiterate on that part it might seem.

Sneha Mehra (00:35:14)  
that we are having a multi, that we are having a code that is multi-threaded. It is not multi-threaded. ⁓ Execution still happens over single thread. It's just a separate single thread in which everything is happening. One for signal, one for this. And really important to understand how we handled this case, this edge case, right? That's what is important with concurrency. So ⁓ skim through the discussion once again. It's extremely simple. Just work it out on the page in case you are finding it difficult.

⁓ The crux highlight is here we are infinitely ⁓ waiting for the things to like your server to complete its execution. ⁓ Just adjust your server to complete its execution. Once it is done, we are setting the signal to ⁓ like we are setting the status to shutting down. ⁓ After this no matter what your server should never go into busy again. ⁓ So ⁓ once it is shut down

your server should never become busy again, which means your once a shutdown is there, your busy set over here, you would never want your server to become busy over here. So only from waiting to shut down is allowed, everything else is disallowed. ⁓ So when that happens, we return nil. ⁓ This is that one edge case, for this we have to add it, we have to add compare and swap over here. But this is the highlight, this is the highlight, this is the essence of implementation.

And this would ensure that the way we are doing it is correct. ⁓ And ⁓ we would never have our client break its connection or server break a client's abruptly. ⁓

⁓ That is it ⁓ for this one. ⁓ I hope you found it interesting. I hope you found it amusing. The world of concurrent programming is really insane. You have to handle a lot of cases. Here with two threads, two simple go routines, we had to handle this particular edge case, but I hope it was fun. ⁓ Great. That is it for this one. I'll see you in the next one. Thanks a ton.

—-------------------------------------

19

Sneha Mehra (00:00:00)  
So does Redis really use malloc? Although we saw in the code base in the previous video that hey, Redis use ZMalloc and it wrapped malloc implementation, but does Redis seriously use malloc? Can there be something better out there? ⁓ So we are all taught that hey, malloc is the ultimate thing that we use to do memory allocation in C language. But, but, but there are libraries. There are libraries that do manual memory management for you. Now, this is where JEMalloc

and TCMalloc comes in. So, Jemalloc and TCMalloc are libraries by Google and Facebook that does memory management for C language. Like you will find the binding of it in multiple languages, but what the idea is pretty simple. ⁓ It does manual memory management for you. Now here, why do you think you need to do this? By the way, where do you find this code? Under ZMalloc, you'll, sorry, under ZMalloc.h, you'll find something. ⁓ This is where you'll find,

that although we're using malloc, but there is a place where it is overridden. So here you see if defined use TC malloc and then you define Z malloc lib something if JEMalloc version major. So here elif defined use JEMalloc. So either you're using TC malloc then do this. If you're using JEMalloc then do this. Something like this exists over here. Now here, what are these? Right? So this is where it's very fun to understand what we are actually trying to do over here. So.

Redis is an in-memory DB. So when it is an in-memory database, which means you are doing a lot of in-memory reads and writes. When you are doing a lot of in-memory reads and writes, what you want, you are allocating and deallocating a lot of heap data. ⁓ When you doing lot of allocation and deallocation of heap data, you might be writing a lot of small objects because key is small, value is small, doing basic incrementation, like increment and what not. So doing a lot of small transactions on it. Now what happens there?

is that when you are directly invoking operating systems implementation, basically operating system deals with pages and what not. A typical page size of an operating system is 4096 bytes, basically 4KB. ⁓ So when you doing a malloc for operating system, it would bring a 4KB page for your allocation, but you are not using that much. ⁓ So it would try to fit in multiple things that operating system level page management won't go into details of that. But when you do a lot of allocation and free, your memory

Sneha Mehra (00:02:26)  
gets a lot of vacuum spaces, which means that let's say you allocated a big chunk of memory and then you freed it up or rather let's say you allocated four small chunks of memory and then you deleted the middle two. Now that middle two ⁓ is fragmented memory because you have some allocation over here, some allocation over here and middle allocation is free. Now you cannot just fit in anything over there, right? So that's a classic case of fragmented memory. So what happens there is that there are libraries like JEMalloc and TCMalloc that does

memory management for us similar to what malloc does. So it's a malloc implementation. ⁓ So they allocate a bunch of memory in junk ⁓ and then it does memory management within that. And it does defragmentation, concurrency control and what not. Which is where JEMalloc and TCMalloc, they perform very well as compared to traditional malloc because traditional malloc typically deals with operating system directly. ⁓ But JEMalloc and TCMalloc has a lot of configurations ⁓ through which

you can make that memory management the way you'd want for your use case to perform the best. ⁓ So that is where you'll find in ZMalloc.h where you'll see a lot of flags called JEMalloc and TCMalloc. This is what it is all about. If you are compiling this code on Ubuntu, you'll find JEMalloc's implementation as the native one. Otherwise, you can change it to TCMalloc as well. Both work just fine. And here in the code ZMalloc.c, ⁓

you would find references to JEMalloc at a lot of places. You'll see how they're actually using it. ⁓ But the idea is pretty simple. For us also, JEMalloc is just a bunch of APIs, a bunch of memory management APIs that are exposed. It's nothing more than that. ⁓ And then it's that. which one to use? So JEMalloc and TCMalloc provides us with some interfaces, some APIs through which we can do memory management. We are wrapping it all under the same name malloc. ⁓

you or by the way you would at couple of places you would also see macros defined in which it does a specific control flow if this then use JEMalloc otherwise use tcmalloc but the idea is pretty similar that ⁓ native memory or sorry native malloc implementation is not performant because it leads to fragmented memory that's where Redis would use something called as JEMalloc slash tcmalloc any one of that to do better memory management which is much which is works very well with concurrency

Sneha Mehra (00:04:47)  
does very minimal fragmented memory, does frequent defragmentation so that you use your memory very wisely. ⁓ The best way to understand how performant JEMalloc and TCMalloc are, I've shared a couple of links in the description, ⁓ you can find that. But I would highly encourage you to check those projects out. You would understand so much more about manual memory management and why it is essential and how to do that. ⁓ So JEMalloc, TCMalloc is what Redis uses, the default one is JEMalloc on Unix operating system and you can actually

you can actually convert them or you can actually switch them to different implementation or you can even use a normal mallop and then run a benchmark, whichever works the best for you, can change it and tune it to the way you want it. ⁓ And yeah, this is what I wanted to bring up as a highlight that ⁓ at scale projects like Redis where it needs to be extremely efficient, they cannot rely on the native things that your operating system provides or like native G-Lib implementation of mallop.

So that is where TC malloc and J malloc comes at the rescue. ⁓ Going deep into that it's extremely out of scope for this one but I would highly highly highly encourage you to check those things out you'll learn so much more about manual memory management. ⁓ In case you are interested do check out that particular thing. ⁓ And yeah that is it for this one in the next one we'll look at signal handling and graceful termination and I'll see you in the next one. Thanks. ⁓

—-------------------------------  
24

Sneha Mehra (00:00:00)  
So Redis supports geospatial queries, which means that you can literally ingest latitude and longitude of a lot of people and lot of entities and you can ask them geospatial information. For example, find people who are near or who are within two kilometer radius from my current location. Redis would be able to give you that. Now imagine the using, like imagine the applications where this can be used. Find people near me, find store near me, I'm building something like Tinder, find people near me who I can swipe left, swipe right, all of that, right?

everything around location and geo proximity. This is how you can use it with Redis. ⁓ The command that Redis exposes is called geo add in which you can pass in for all the, for the key. Let's say I put it riders for riders for a particular latitude and longitude. Let's say my rider ID is her pit. ⁓ I do this, this gets added into that particular key. And if I see the encoding of that debug object riders, ⁓ I see it is encoding with list pack. Hey, what?

it shouldn't be something else. So what redis does is that it wants for a particular key would want all the things, ⁓ all the entities within that. So it is adding elements within this. ⁓ So that's why it's a list and each lat long are encoded in a certain format that makes operations like finding k near me very efficient. ⁓ which is where the algorithm or the representation that comes our way is called as geo hash. Now what we'll do

is we'll build theoretical understanding of it, the intuition behind it, and then we'll work you, and then we'll do a very quick browse of source code. ⁓ To see what it does, how it does. ⁓ So, ⁓ Redis supports geospatial queries, helps us find all proximity things, everything around proximity you can build on top of Redis. But the core requirement is the proximity distance is given on query time. So it's not a fixed thing that, hey, I will always ask within KK, K kilometer radius. It would be given at runtime.

It would be given when you are firing it. That hey, give me within five kilometers, give me within 10 kilometers, give me within 20 kilometers. Right? So given that, you need to accept that parameter on the runtime and be very fast in computing it. Which is where it uses an encoding called GeoHash. Now what GeoHash does is GeoHash, because our location are represented in latitude and longitude. ⁓ Latitude is this.

Sneha Mehra (00:02:22)  
Why longitude is this? So every point, you, me, every single device, the GPS coordinate that we get, that is lat and long. That is our unique position on the earth. Right? So that is a latitude longitude thing that we know. Now, you cannot be efficient ⁓ in finding proximity within that. For example, if I just translate it into two-dimensional geometry, x, coordinate. Right? So then if you would want to find out, hey, let's say this has some x, y.

And you would want to find all the points that are within k kilometer radius from this x, y. You would have to match it against all the points to compute the distance and then you can derive and say, hey, this is within k kilometer, this is within not. ⁓ Which means that this kind of operation is extremely costly when you are dealing with n dimensional geometry where n is greater than equal to 2\. So which is where you need a simplistic approach to solve this problem and solve it for one dimension.

So which is where GeoHash comes in. Now this is the beauty of this algorithm. So what GeoHash does, takes the output is a 64 bit integer, literal 64 bit integer on which you can apply normal range queries. That's the fun part. So it takes a lat long, an x, y, it transform it to a GeoHash function and generates a GeoHash for that particular x, y, for that particular lat and long. Now the core idea behind this is that it splits the world in half every time.

It splits the world in half every time and then figures out where do what. ⁓ So this is the idea behind it. So let me do this. Let me split my world in half first vertically. Now if I split my world in half vertically, let's say everyone who resides in the left half is assigned zero. Everyone sitting in the right half is assigned one. ⁓ So depending on where like, see your coordinate would be constantly changing and this

has to be lightning fast to compute. This is just an intuition behind it, right? On how are we achieving what we would want to achieve, right? So left half is zero, right half is one. Then I split it in horizontally and I say top is zero, bottom is one, right? So because my point first lied on the right half of it, ⁓ I would put one and then it is at the top. So I'll put zero. Then I split it vertically.

Sneha Mehra (00:04:45)  
So if I split it vertically, ⁓ again my left becomes 0, right becomes 1\. My point is on the left side. Because my point is on the left side, it is 0\. Then I split my world horizontally. I split it over here. I do my top as 0, bottom as 1\. ⁓ X is on the bottom. So I put 1\. And then I split my world vertically. Here, this region. Now here, your left is 0\.

right is 1 and x is on the right side. So I put my 1 over here. So this ⁓ is my ⁓ geo hash for x. Now here I have stopped it at 5 bits. I did not go beyond that. But here you can see to a very small fraction also you can go. Like you can take your ⁓ you can take a granularity to 32 bits because let's say your integer is 32 bit long you can go till 32 bits. I will say how much will I be able to serve that like at

26th bit you would be at a 19 meter error rate like within 19 meters or rather 10 meter if you want to round it off. ⁓ So 10 meter accuracy you would get at 26 bits ⁓ 26 bits 10 meters or rather it's much more less it's I think it's 2 meters only at 26 bits. I'll just link the resources ⁓ in the description below. Right so it's that efficient right so at

that much of granularity what you would get is you would get much much much higher precision that you would want. ⁓ Right? Okay. So now why did we do this? ⁓ Now here the core thing is that any point that lies in this region would get the exact same geocode. Correct? So let's say if I have a point ⁓ y over here. Let's say x is over here and y is over here. Now here what would happen is my y

Because I'm only dealing with 5 bits, right? If it would be much more, then it would be much more granular. But here if I'm only dealing with 5 bits, ⁓ all the elements that lie over here will get the same x, y. No, sorry, will get the same geohash. So my y's ⁓ geohash would also be 1 0 0 1 1\. Right? So now if I want to know people who are closer to me, what I can do is I can just see that hey, people who have the same geohash would be

Sneha Mehra (00:07:08)  
same would be like near me and because they are all belonging to the same region. ⁓ But if we do this at a very granular level at a meter square level, ⁓ then that would not be matching. ⁓ So now let's say if I'm exhausted, let's say we are done exploring this part, ⁓ that this region. Now, ⁓ let's say we building Tinder and we ⁓ exhausted all the matches that we could get over here. Now, what do want to do?

we want to zoom out. Right? We want to go a step beyond. Now, which means that I would want to match ⁓ Z that lies over here. What would be the geocode of Z? The geohash of Z would be 10010\. Right? Because it's on the right, then the top, then the left, then the bottom, then the left. Right?

So that is what we are doing with Z. If you look closely, if let's say X was doing that search, then what X would find? X would first forgot Y, it exhausted all the options. Now it increased the zoom out, it increased the distance. I don't want to skim through 10 meters, 10 kilometers, I want to make it 20\. So then what it would need to do, I would need to match Z in that case. How do I do that? If you look at this, they share the same prefix. Look at this prefix. ⁓

For x the prefix is 1001, for z also it is 1001\. So what this shows, this shows that if I ⁓ remove the bits from the right and I just match the prefix, I was first matching over here, now I'm matching in this one region. If I want to zoom out further, I can skip one more bit and I'll be zooming in and I'll be ⁓ choosing this big of a region. And then this big, and then this entire right half, and then the entire one.

Right? So here, the precision, the more you go to the right, the more the precision is. The more you go to the left, the higher the bird's eye view is. So you're zooming out and then zooming in. So now, your problem statement has changed to just a simple prefix match. So for example, if I just put all of these elements, ⁓ all of the locations that you are get in a try.

Sneha Mehra (00:09:34)  
as simple as a try. So I can just put all the hash codes just a dummy implementation just all the geo hash in this try 10011 something like this if I do that and if I want to just find out if let's say I am at this point x is at this point and it is done ⁓ finding it has done exhausting all the options this and I want to zoom out what I can do is I can then ⁓ remove one bit from the right and I match over here so then

all of this would be matched. If I'm done with this, I can just ⁓ go a step further and then it would match all of this. Right? So your problem just reduces to prefix match problem. You can very well implement it with trie or a red XT. Anything. ⁓ It's a normal substring match. So people having closer prefixes or having similar prefixes are closer to each other. ⁓ Greater the length of shared prefix, closer the

entities are. This is the intuition behind GeoHash and this is the beauty of GeoHash. Right? That it makes zooming, zoom out, matching so, so, so, so simple. Right? Okay. So now what do we have? We have two entities that are closer to each other if they share a longer GeoHash prefix that we established. Thus our problem statement has now reduced to just finding a prefix match. So which is why all geospatial databases, they use GeoHash to find this particular, like to do anything around proximity.

So how do we compute? For us it was easy to explain it with half half half half half half and all right. But you don't have to run a loop to implement this. You can literally take a relative offset. So for example if I'm at this point ⁓ and if I'm only doing it for latitude right. So latitude is this horizontal lines that you have and longitude are the vertical ones. So for latitude if I'm doing it the range is minus 90 to plus 90\. ⁓ So here

what I would do is for the minimum value and the maximum value that I have I just want to find out the relative offset of this the fraction of it so this is where what I would do is ⁓ I would just be computing this divided by this it's nothing but that left right left right ⁓ only but for latitude so splitting it into half every time and going until that position that is exactly what you would get when you do A by B

Sneha Mehra (00:11:56)  
⁓ So this is A and the maximum difference is P. So the main difference between the ⁓ least latitude and the lowest latitude and the highest latitude that becomes a denominator and the numerator is the distance or the latitude from your point till the max latitude. The difference between that divided by total latitude coverage is the relative positioning of this on the latitude. Similarly we do it for longitude. Now when you do this, this would be a fraction from 0 to 1\.

Right, because you are doing A divided by B, B is the maximum length while A is the fraction that you would have to cover. So that would be in the range of 0 to 1\. Similarly, we do it for, ⁓ we do it for longitude as well. Right, we get fractions and then we shift it by the things that we would want to like. If let's say we ⁓ are going for granularity of 26 steps or rather 26 such iterations, right, left, right, left, right, like that thing, ⁓ then we would be left shifting it by 26 if that's the precision that we are going for.

If you are going for precision of 20, we would lift it by 20\. We will talk about that. We will see in the implementation. ⁓ So then we have one number for latitude and one number for longitude. So we have GeoHash independently computed for lat and log. ⁓ So now what do we do? Because there we did half, half, half, half every time. Here you see we always first split.

vertically and then horizontally and then vertically and then horizontally. It's like splitting by longitude first and then latitude and then by longitude and then by latitude. So this is like interleaving ⁓ of latitude geohash and longitude geohash. So which is what we do ⁓ is instead of storing two, because two values were the problem because we could not do efficient matching. So that is what we are doing is we are recomputing or rather we are computing geohash for latitude and longitude separately and then interleaving them.

such that odd bits are longitude and even bits are latitude so that we can reconstruct independent latitude and longitude if we want to. ⁓ So in the final 64 bit because this would be a 64 bit represent because both are 32 bits the representation we need is 64 bits so we are interleaving 32 bit and 32 bit so total would be 64 bit like we are interleaving it so odd bits are latitude even bits are longitude. So this is how GeoHash does that part. ⁓ So this is how you would be computing

Sneha Mehra (00:14:13)  
Geo hash and interleaving them for better performance. Now you'll say how is it better? Now ⁓ here we saw how are zooming out and zooming in worked, right? We just removed one bit and we matched ⁓ the prefix n minus one and then n minus one and then n minus one, such that we are matching that. So we are zooming out constantly and zooming in constantly. So we can typically navigate through the map like this, right? So now this kind of ability we also want to get with Geo hash. Now you think that hey,

If we have 64 bit limit ⁓ and I have two 32 bit integers, why can't I have first 32 bits as longitude and then second 32 bits as latitude? Won't that work? That would not work. Why? Because then if you are removing the bits from the right, you are only removing it from the latitude. Longitude is remaining the same. That is what the problem is. But if you interleave it such that here you have latitude and longitude interleaved, when you removing bits from the right to do better to

do a larger to do a shorter prefix match or ⁓ increasing the scope, you are removing the lowest significant bits from it, which is like exactly like zooming out if you are thinking of that. So which is why you still don't lose out on abruptly. That's why you don't ⁓ just throw 32 bits of longitude and 32 bits of latitude. You are interleaving them so that your logic of removing out from the right side to zoom out and then add to the right side to zoom in, you are not losing out on that. ⁓

Second, makes it more precise to zooming in, that's what we saw. ⁓ And then modern architecture has optimized 64-bit CPU computations. Because modern architecture, ARM 64, AMD 64, we see 64-bit machines, so operations of 64-bits are much more efficient than 32-bit operations, because of which, ⁓ if you store it in 64-bits, it would be much, much, much faster for CPU to compute that, versus two 32-bit operations. ⁓ Plus, you get all the benefits of proximity the way we exactly discussed.

⁓ And this is the idea of GeoHash. You can use this thing or any database, any Geo special database in the world uses GeoHash for this kind of competition. This is the idea, the core intuition, the logic behind the GeoHash implementation. ⁓ Okay, ⁓ enough of this. Let me walk you through the actual GeoHash source code of Redis to see how they have actually implemented. So we are in the file geohash.c and this is where you would find that same exact same thing that I just discussed.

Sneha Mehra (00:16:38)  
but this is where you'll find an interleave function. ⁓ before we do that, ⁓ do we have any, but let's start with interleave function. So here, this is that one thing that I would want to highlight that when we are interleaving the bits, we are one bit of latitude, one bit of longitude, one bit of latitude, one bit of longitude and so on and so forth, right? So you may think an idea to do this is using for loops and all, but this is a very highly optimized way of doing it. It's literally five, five, five, five, three, three, three, some.

looks like a magic number some sort of magic happening but this actually this actually does bit interleaving because it is literally doing that and that's highly optimized it's order one computation you don't need any for loop and nothing it's lightning fast just bunch of bit operations but it actually does bit interleaving the way we expect it to do one bit latitude one bit longitude one bit latitude one bit longitude and so on and so forth right so which is where it gets longitude ⁓ it gets x law

y long x long something like that some nomenclature un32 and it interleaves into 64 bit and then puts it there. ⁓ Right? So this is what it is doing and get Geo hash coordinate range. This is the max mean that we talked about. If I just click on that, ⁓ you see minus 85 to plus 85 and minus 180 to plus 180\. So minus 85 to plus 85 minus 180 to plus 180\. Right? That's a typical latitude longitude range that we know of.

So, GeoHash encode, this is where it is encoding it in GeoHash and this is where we see the relative offset, the relative positioning of that. So, lat offset is latitude, the current latitude minus lat range of min, right? Lat range of min divided by lat range of max minus lat range of min, right? That it basically computes similarly for longitude. And then here what we are doing is because this would be in range of 0 to 1, but

we want 32 bit integers, right? We cannot just use float for that. So that is where what we are doing is we are left shifting it. We are left shifting it. ULL is 64 bit unsigned long long, 64 bits left shifting it. So let offset into equal to ⁓ one ⁓ ULL, which means 64 bit integer left shifted by step. So if I'm going at a granularity of 26 steps, which means 26 times I'm doing this for latitude and longitude both, which means if that is a granularity that I'm operating with,

Sneha Mehra (00:19:02)  
This is where what you'll get is you'll get a 32 bit like, ⁓ like N bit like for N like for 26 times we are shifting, we are left shifting it by 26, which is what we are doing over here. And that value would be populated, which is what we are interested in, right? So left shifting them and then similarly left shifting the longitude. And then for these two, they are just interleaving at 64, right? So we are interleaving two 32 bit integers that we have into one 64 bit that we just saw.

And this is exactly where that logic is. Apart from that, it's about geo hash encoding latitude, longitude. This is where the encoding is happening that we just saw. And then other helper functions here and there. But the core idea still remains the same. Here you can find all the code and I would highly, highly, highly encourage. And just one thing, just one thing on that. The way I explained it was about splitting, splitting, splitting into half and then you zooming and zoom out. Just one hunch over there. Given that the way we are doing it, you don't have to only remove one bit.

from the right side. You can remove two bits from the right side and you can literally double your search space or other four times of search space. ⁓ Right? So you can literally do that every time four times of search space, but you can also deliberately choose to move in some direction, this direction, this direction, this direction, this direction, ⁓ on what you can set and unset. ⁓ So instead of setting or unsetting or removing one bit at a time or two bits at a time, you can choose to set one to zero and another to one or one to one and another to zero.

to move in that specific direction. Such a beautiful piece of halkurtam. Which is what you will find at the bottom of this code where it talks about ⁓ move ⁓ x, ⁓ move y. This is what it is doing. Right? And then GeoHash neighbors, would get the neighbors and all. So it's all about basic computation that it is doing. But the idea of GeoHash remains the same of inter like computing latitude offset or longitude offset interleaving them. The intuition was about splitting the world into half and half, zooming out by removing the bits.

zooming in by adding more bits. Problem statement reduced to just doing a prefix match, right? As simple as that, right? So yeah, that's all about GHASH. And again, I've not implemented this, but I would highly, but I've raised an issue, raised an issue in our GitHub. I'll link it in the description down below. I would highly, highly, highly encourage you to implement this and contribute back to this repository. It's a very interesting algorithm to be really honest and very simple one to implement. But if you're really interested into contributing back,

Sneha Mehra (00:21:26)  
to this code base, would highly, highly, highly encourage you to do so. A very simple upgrade for ⁓ Brigadier use cases, right? So yeah, great. That is it for this one. I'll see you in the next one. Thanks a ton.

—--------------------------------

14

Sneha Mehra (00:00:00)  
So let me show you something very interesting. ⁓ Now on the top right, I have a normal Redis running on port 637 and this is a standard Redis server. And on the bottom left, we are connecting it through CLI. ⁓ Now what I'm doing is, ⁓ let's say I do, ⁓ I don't have any key set. ⁓ Let's say I do INCR of, let's say my key I'm using Arpit. ⁓ Let's say I take key Arpit, ⁓ I'm doing literally doing INCR on top of that.

gave me 1\. Arpith ⁓ did not even exist, but it still returned ⁓ 1\. Right? Ok. Let me do something else. I am doing set k to 10\. ⁓ So, now my value for k is set to 10\. Now, if I do incr k, ⁓ it does 11\. Right? So, and here if you look at this, it is returning me an integer 11\. Right? But now when I do ⁓ get of k, ⁓

is returning me string 11\. So, ⁓ what's going on over here? Right? At some time when we do a set of key k and then 10 and then I am doing increment it is giving me integer then I am doing get it is giving me string. What's what's happening over here? The ⁓ answer to this lies into how Redis stores integer. Right? And this is where I would want to bring to your notice

something called as a standard Redis object. So now the file that I'm open, it's not part of our GoLang project. It's the actual source file from the actual Redis source code, where I want to highlight a couple of things, right? So that we can implement this on in our logic. And this is very fascinating, very fascinating. So in the hash table in Redis, ⁓ basically your Redis is a key value store. So everything gets stored in a hash table. In that your key is hashed goes to a particular location and the value is a Redis object.

Now this redis object is this. ⁓ So this redis object encapsulates all the types like sets, lists, hashes and what not everything goes over here. ⁓ So now when you are putting it in the hash table ⁓ everything is being created as a new redis object. Now what is the structure of this redis object? This redis object contains type, encoding, ⁓ LRU, reference count and pointer. ⁓

Sneha Mehra (00:02:27)  
This colon 4, if you remember your C classes, this is bit fields. So which means what it says is, it says that I have a type called unsigned integer ⁓ whose variable name is also type and it should be allowed 4 bits. It should be assigned 4 bits out of this. Then I have encoding, which should also be assigned 4 bits. ⁓ So this together ⁓ gives us 8 bits, which is 1 byte.

Then I have LRU bits, these are 24 bits. ⁓ Here it's declared. ⁓ LRU bit is 24\. And then a reference count which is 4 bytes. LRU bits is 24 which means 3 bytes. So 3 bytes plus 1 byte is 4 bytes. And then reference count is integer which means 4 bytes. And void is 4 bytes, void pointer. ⁓ So in all it becomes ⁓ 1 byte, 3 byte is 4 bytes, 8 byte and 4 byte 12 bytes.

So each Redis object is a 12 byte object. Right? Okay. So now here, the key point that I would want to highlight is type store something called as type, encoding store something called as encoding will go into that. Then LRU is used for key eviction strategies that we saw a couple of videos back on what key eviction strategies are and how it should, we'll be implementing this in some like in basically future videos. Then reference count is used for freeing up the objects.

It means that this is a classic garbage collection strategy. When the reference count becomes zero, we basically free the particular object would be used for that. And finally is void pointer. Now here, this pointer can point to a set, a list, a linked list, a bloom filter, whatnot. ⁓ So this is the pointer to that particular object. ⁓ But the core object is Redis object. Now let's see in this video, we'll be focusing on what this type and encodings are. And we will be implementing increment.

the way it does it, right? And which is what the fascinating part is. So here what do we do? We first see what these types are, what these encodings are. So if I, in this file itself, if I search for obj underscore string, ⁓ and I would be coming over here, line number 597\. So the Redis has few limited set of types. First is string, then is list, then set, then sorted set,

Sneha Mehra (00:04:50)  
hash, ⁓ then module and string. These are standard objects. ⁓ These are the actual objects that you have. ⁓ Here if you clearly see there is no int type object. ⁓ So int is not a type in Redis. ⁓ So the object types that it supports are string, list, ⁓ set, zset, ⁓ or sorted set, ⁓ hash, ⁓ module and string. Which means int must be stored somewhere in this, ⁓ which is where encoding comes in.

So now when I talk about encoding, ⁓ so if I do this ⁓ obj underscore string ⁓ and then ⁓ encoding, if I check the encoding will come over here. So here you can see all the encoding. So encoding implies a concrete implementation. So for example, obj underscore encoding underscore raw, which means raw is in bytes.

it just bytes. Now here your type would be string but encoding is raw which means I am storing a string. ⁓ String is just an array of bytes. So I can interpret it the way I want. So if you are implementing bloom filter, your type would be string, encoding would be raw. If you are implementing a normal string, type would be string, encoding would be string. ⁓ But if you want to store integer, your type would be string, your encoding would be integer.

⁓ So what you are storing, you're just storing bytes and the way you'd want to be interpreted ⁓ as an integer. ⁓ So always when you're doing an integer operation on Redis, it is literally converting it from string to int and then back to string and then persisting it. Not even joking, that's exactly what it is doing behind the scenes. ⁓ And then you have ⁓ a hash table, zip map, link list, zip list and whatnot. So these are different types of encodings.

that it has and types that it has. This is the actual Redis source code. ⁓ Okay. Before we jump to the actual code, let me just quickly, quickly, quickly walk you through basic theory that I want to cover so that you understand when we are implementing it, how, how like the way we have approached it. Right? So every single object we put in Redis is wrapped in a Redis object so that we keep some meta information with it like, like a reference count, white pointer and whatnot. And

Sneha Mehra (00:07:15)  
This is that one generic object that we are passing to all the functions to get things done. ⁓ And this is how the structure of Redis object is which we just saw. ⁓ instead of assigning four bytes, now here interesting design decisions. First of all, ⁓ it has assigned four bits to this. Four bits to this. Like you think, ⁓ I can just store type as an integer and encoding as an integer. Remember integer takes up four bytes.

But if your types are limited, let's say at max you would have 16 types and 16 encoding, right? So given that you have only this much, why would you want to look at four bytes and four bytes for this? You can just save that space. So that is where what it does is, instead of assigning four bytes each, it is assigning four bits each. And this way Redis saves memory per object because that's his heart and soul. So instead of having it like four byte and four byte and four byte,

and 4 byte and 4 byte instead of having 20 byte object it is just doing it in 12\. So almost 40 % saving out of the box on Redis objects. ⁓ Very very important design decision. Then second is void star pointer allows us to refer to any object malloc that could be bloom filter, big string, ⁓ basically Redis pop-up, everything goes there. Right? Then reference count is a place where some object is referenced it becomes plus plus and then

When it is dereferenced it becomes minus minus as soon as it becomes zero you can free it so that your space is freed for you to do other objects allocation there. Classic garbage collection strategy. Then LRU takes up 24 bits which is 3 bytes. Now we will go into LRU detail sometime in the future but it stores it in Redis Object now you know where it stores that. The total size is 12 bytes. Now here what is type and encoding? So Redis supports 7 types out of which module is not a type but it is using it to

⁓ to basically specifically reference it. But string list set, sorted set, hash and string, these are the critical types. Each type has multiple encoding. So string has raw int and embedded string. Embedded string is, if a string is less than 44 bytes, ⁓ curl line, ⁓ length of which is less than 44, it is an embedded string. If it is more than 44, it is a raw string. Then integer ⁓ stored as a string, representation of it. Then list has ziplist and linked list.

Sneha Mehra (00:09:36)  
set as insert and hash table. So for example, list can be implemented as a zip list or a linked list. At smaller scale, it is implemented as a zip list. ⁓ it basically crosses a certain threshold, let's say, hypothetically, let's say the threshold is 100 elements. If your list crosses 100 elements, your zip list will be converted to a linked list. ⁓ And that is where the change of encoding would happen. So list implementation has two things, zip list and linked list. Similarly, sorted as this zip list or skip list.

hash is ziplist or hash table, right? And ⁓ as soon as a certain threshold is crossed, a generic type becomes, sorry, a specialized ziplist or a very specific small scale use case, ziplist, or it becomes a generic type, which is hash table or skip list or basically whatever it grows into, right? So that is the idea. This is why types and encodings are important. Because of this one critical decision, we can...

like literally implement anything that is just a stream of bytes and just encoded a string as simple as this, no overcomplication. This is the beauty of Redis, right? Where it made this very ⁓ thoughtful design decisions, ⁓ right? Now it's time for us to implement it. Let's see how we would implement it. Again, we'll do an exhaustive code walkthrough of implementing types and encodings. We are taking a very similar approach because right now we are re-implementing Redis.

but there could be some decisions that you could challenge that, hey, why are we doing it this way? Why not the other way? Obviously it's the designer of the database who has made the decision. There would be some rationale behind it. ⁓ Okay, so again, this source code is presented at github.com slash dice, db slash dice. You can ⁓ go through the commits and you will find one commit having all of these changes. ⁓ Okay, let's do an exhaustive code walkthrough on what happens. So first of all, I created a file called object.go.

and in this object.go file, the object that we defined in some other file, I moved it over here. Now, instead of it only having value and expireSAT, earlier it used to have only two values, value and expireAt. I'm keeping value as is, as an interface, ⁓ and because interface in golang is same as void star ⁓ in C++ or C. Then expireSAT as an integer, this is something we have to change. We saw how Redis uses LRU with. ⁓

Sneha Mehra (00:12:00)  
only 24 bits, I don't want to store it 64 bits over here and expire SAT. There is ⁓ a great hack that ⁓ ready supplies to approximate LRU. We'll be talking about it in the future one. And then I'm storing type encoding. Now here, one key thing. ⁓ CEC++ gives us ways to apply bit fields on set. Which means that each element in the set, I can assign it to allocate or to have a limited number of bits.

For example, type had four bits and encoding had four bits. Golang does not have that. So which means what we would have to do is, I'm defining a field called type encoding, ⁓ which would be of type uint8, which means an eight bit integer. First four bits, I'll interpret it as type. And second four, like last four bits, I'll be interpreting as encoding, right? Which is what we would be doing. So object type string whose value was zero there.

which value is zero over here, but just left shift by four. And because first four bits I want to set it to here. So if I add multiple types over here, zero, one, two, three, four, I'm just left shifting it by four bytes. ⁓ sorry, by four places, by four bits, ⁓ And my encoding goes as is. My encoding raw is zero, int is one, and embedded string is eight. ⁓ It is exactly what I've stored over here. So if I'm storing...

Type encoding in C, C++, I could do it with bit fills, it would work just fine, but I don't have assembler support over here. So what I'm doing is, I'm primarily just doing an, I'll be doing an or of this, object type string and object encoding. So if I'm storing an integer, I'd be storing type as string ⁓ or like a bitwise or an object encoding int, right? So this is how I type and encoding would be defined. So just to show you,

on how that would happen is let me open a new file that I created called typeencoding.com. So here you can see if from a particular type encoding object, if I want to get a type of it, type is what? The first four bytes, right? So the most significant, sorry, first four bits, the most significant four bits is there. So what I'm doing is in order to extract that, I'm doing a right shift by four and then left shift by four again.

Sneha Mehra (00:14:14)  
Right? Because I want to set my most significant bit to zero that I could have also done with doing an add. So I am just showing different ways to do it. Other way to implement this would be just doing a te and 0b 1111 0000\. ⁓ This would also work. Right? Because what we are interested in is we are interested in the first four bits. The most significant four bits. And I am just using it.

And given that in the object also we have stored it ⁓ as left shifted element. So if I'm storing ⁓ one, it would be one arrow arrow four, which means one left shifted by four, four bits, right? So first four bits this, second four bit this, right? So this would be how we would be defining our time because Golang does not support bit fields. That's why we have to do that, right? So to get the type, we are doing this to get the encoding, we can do this. So I'm just reverting the changes so that nothing else.

Okay, fine. ⁓ So to get the encoding I'm just or I'm just doing a bitwise and with 0 0 0 0 1 1 1 1 I'll be getting encoding out of it Then I have another function called assert type if the type is not there I'll be returning an error if it is there good enough so I'm just checking if the type encoding has this particular type or not ⁓ and if Assert encoding I'm taking a type encoding object and just checking that if the encoding of that is same as the encoding that we pass or not just for

edge case handling or just to check for that because for example if I am doing an increment operation it can only happen on an int object. It cannot happen on any other object. ⁓ That is one. ⁓ Okay. Now that we went through object file we went through type encoding file. So now let's see how store file has changed. So store file what we are doing is when we creating a new object we were only taking first we were only taking value duration. That's it.

because that's what we were storing. But now we are also storing type encoding. So I'll be taking one for object type, one for object encoding. Object type is Uint8 and encoding is Uint8. Unfortunately, we can't do anything else over here, right? So that is where, like this is the bare minimum that we can allocate ⁓ and what everything else remains the same. And I'm just adding this over here. Type encoding is object type bitwise or object encoding.

Sneha Mehra (00:16:36)  
So first four bits set to this, last four bits set to this, like first four bits type, second four bits set to encoding. ⁓ And this is how your store file would change. Everything else remains the same. So we saw store file and we saw this. Now let's take a look at eval. So to understand this, let's first ⁓ see implementation of increment function. ⁓ So what did we, or rather no.

Let's start with set's implementation. Set is very interesting over here. So, ⁓ what did we see when we saw that example? We saw that I was doing set. Set k space 10\. And then I did increment. So, which means that somehow, although I was like, I can use the same set to set a string like set k abc. It would have stored abc. But if I do set k 10,

it stores 10 as an integer. So it understands that 10 is an integer. So there would be a place where it would be deducing the type. So basically that's what happens when we do set. So the first step is we extracted key value like always, but then we try to deduce the type encoding from the value. So the value that we are storing against the key, I would want to deduce what type and what encoding it belongs to. And I'll be passing that exact same thing over here. So here I'm just, I just removed the

bracket error thing, nothing else has changed. I'm just basically making the code cleaner. And here you can see in the new object, I've passed in object type and object encoding. So now let's see what deduce type encoding does. So deduce type encoding does extremely simple thing, extremely simple. So the idea is if I'm able to convert the value to an integer, then my object type is string and object type is string and object encoding is integer, right?

If it is less than 44 bytes, ⁓ then my object encoding is embedded string. If it is not 44 bytes, more than that, my default is raw. As simple as this. So every time we are doing set, ⁓ we would be trying to convert it. We would be trying to parse the value to see if it is an integer. If it is an integer, I'm storing encoding as an integer. ⁓ Right? This does not mean we are literally storing an integer object. ⁓ This just means we are converting it and we are storing integer as a string.

Sneha Mehra (00:19:00)  
inside my hash table inside this redis object. Right? So that is the idea. Now let's we saw how set changed. Now let's look at increment implementation. So I come back to eval file and eval may be go to this incr because that's what we are here to implement, right? Incrementing. So now with increment what we would do is we would get the only argument that it gets ⁓ is the key that we would want to increment. Right? But the beauty of increment is

even if the key does not exist, it starts with zero. ⁓ If the key exists and it is an integer, it would increment a value. And the value that it will be returning would always be an integer. But when I do get, I should get strict. ⁓ So that exact same behavior we have to mimic. So eval increment first takes the arguments. If argument is not one, insufficient arguments, then we get the key, we go to the store, try to get the object. If we get the object and see if the object does not exist, we create a new object.

and put that new object. We create a new object with value 0 with no expiry set, string and encoding as integer. Because we know that we are storing 0 over there. Right? And now here after this we would have object set to either the existing object or this object. Right? Then we do checking. ⁓ If the type encoding of this object is string, if it is not string then return an error. ⁓ If

the encoding if the type encoding has encoding of an integer if it is not ⁓ if it is not an integer then return an error right but if it is string an integer which means type string encoding is integer which means the value is indeed an integer what do we do we convert it we parse it back from that value convert it into an integer do an i plus plus set the value by formatting it to string again and then encode the integer and respond

This is exactly what you have to do. Now you'll think, hey, it's so costly. ⁓ But this was the decision that anti-res or rather that basically the creators of Redis took. ⁓ Because they wanted to simplify or they want to focus on the user experience while thinking of performance. If you would create new types for everything, might become problematic. But again, no one's stopping us from implementing it that way. ⁓ Maybe if tomorrow we take this implementation and say, no, integer is a type for us. We can very well do

Sneha Mehra (00:21:23)  
That would save us a ton of computation. But it's perfect. It's their decision, it's their database. We're just re-implementing it. We're just understanding ⁓ how to create a database with so much of futuristic mindset. That's the idea behind it. Nothing more, nothing less. ⁓ This is how you would be implementing increment. Now that we went through the source code, let's quickly go through ⁓ our own implementation and see ⁓ our implementation in action.

it indeed works. So if I remove this, minus p 7379, but before I do that, let me run this job ⁓ 7379, connected to that. I don't have any key set over here. So let me do increment k. ⁓ It returned 1\. ⁓ If I do ⁓ get k, ⁓ it returned 1 as a string.

Now if I do increment k again, it gives me two, three, four, five, and all of them are integer. Look at the type integer, right? But if I do get k, it gives me six in quotes, which is a string. Type is string, encoding is integer. Every time we are incrementing, we are checking if the encoding is int or not, and then going forward, right? So what did we just implement? We implemented increment. That's good, right? But what did we add? We add support for types and encodings.

This makes our database allowed to be extensible. So we are promoting extensibility. So if tomorrow we would want to implement Bloom filter, it can be implemented because Bloom filter is just a stream of bytes, nothing more than that. So what we can very well do is we can use ⁓ object as string, encoding as raw, or we can create our own encoding called Bloom filter. And it would work just fine. It's our encoding. ⁓ So string.

and encoding as Bloom filter, which is just storing as a stream of bytes or a stream of bytes as simple as that, right? So making database futuristic as futuristic as possible or as expensive as possible was the key idea behind it. And now you understand how Redis is ⁓ multiplexing multiple encodings for the same object. Now, depending on implementation, now here also one key thing is we saw how list is implemented either as a linked list or a zip list.

Sneha Mehra (00:23:41)  
you may come up with a third approach, maybe as a simple array, right? Where you have a bounded list. If you'd want to implement that, you can very well do that, right? So multiple implementation of the same object so that it's your module, it's your extensibility that you can work for. So if you are looking to make some substantial changes in Redis, you can just take one data structure and let's say you found a data structure, sorry, you found a ⁓ heavily optimized version of set.

You can implement it as a new encoding for set ⁓ and hopefully people adopt it. That's the idea behind it. So there are key types that they have defined and bunch of encodings that they have defined. You can define your own encoding and commit to Redis. ⁓ And this is the beauty of Redis as such a simple database. You obviously could challenge it, hey, this is so expensive and all, but that's fine. That's fine. It's their decision. It's their database. If we would want to rethink it, we might take a different route this time. ⁓

And great that is it for this one. ⁓ I hope you found it interesting. I will see you in the next one. Thanks.

—----------------------------------------------

26

Sneha Mehra (00:00:00)  
So Redis supports a bunch of approximate data structures ⁓ and one of them is Hyperloglog. So what are Hyperloglog? Hyperloglog is ⁓ an advanced data structure which is probabilistic in nature and it does an estimation of cardinality. To put it simply, let's say you are given a very large stream of elements ⁓ and what you want to do is you want to find the number of unique elements within that. ⁓ This is a very, very, very common use case.

for most of the in-memory database. ⁓ So what Redis does is Redis uses Hyperloglog to power that and you can use that thing through command called pf something. So pf add, pf merge are the two common utilities that you can use and basically pf count. So pf add adds an element ⁓ to Hyperloglog, pf merge merges to while pf count counts the cardinality of it. ⁓

Let's say if I'm given a stream of numbers 1, 2, 3, 2, 3, 4, 1, 2, 3, 1\. Right? And now if I want to know the cardinality, I can just put all the elements in the set and I can just count the length of the set. Isn't it? Pretty simple. So then why Redis has to use HyperLogLog? Because this is not at all memory efficient. Imagine the amount of data that you are ingesting, although duplicate, but you would want to add and keep all of the unique elements at least in the set. Isn't it?

So if you want to keep that, that is memory bound that would take up a lot of memory and the only operation that you are doing is just count on it. You not even doing anything else. So given that, given this, can you make it ⁓ better because where all you are interested in is the cardinality of it. Right? So then, which is where instead of going for an absolute correctness, what if I could approximate the answer? If I can approximate, let's say there are four elements, four unique elements.

And even if I say phi, there is no harm in that. Given that is a use case, you can go for an approximate data set that would save you a ton of memory while giving you pretty ⁓ near accurate result. Right? So which is where we take a look at Flajolet-Martin algorithm. So Flajolet-Martin is very similar to Bloom filter, which means it's just a huge chunk of memory, ⁓ a filter, if you may call it, and we would keep hashing and storing data within that. And it's an approximate in nature.

Sneha Mehra (00:02:24)  
So it's very similar to Bloom filter. We use a similar concept hash byte array for quick computation, very similar and storage required. This is the best part. So here the storage required for you to store or rather to ⁓ do this cardinality estimation is order log ⁓ where are the number of unique elements. So if you have ⁓ unique elements to be ⁓ 1 million, it would be log ⁓ log 1 million to the base two. It's that.

tiny of a space you would require to count that to do this estimation. So how does it work? The idea is pretty fascinating. Now look at this pattern. Look at this pattern when let's say for a uniform distribution of numbers starting from 0 to 7 0 1 2 3 4 5 6 7 if I write binary representation of it we get 0 0 0 0 0 1 0 1 0 0 1 1 so on and so forth till 1 1 1\. Now let's say

I keep a track of row. Row will be the position of the rightmost set bit. So in 000 the rightmost set bit is not applicable. 001 the rightmost set bit is at position 0\. In 010 the rightmost set bit is 0, 1 at position 1\. ⁓ 011 is rightmost set bit is at 0\. ⁓ 100 the rightmost set bit is at 2 and so on and so forth. Right?

Now if I want to compute in this scale in this range if I want to compute what is the probability of me getting a right most set bit 0\. ⁓ The probability is number of cases out of which how many are at 0th position so it's 4 by 8 is equal to 1 by 2\. What is the probability that I will get rho equal to 1? It would be 1 to 2 by 8 equal to 1 by 4 and then probability where I get 2 it's 1 by 8\.

Right? Okay. So we can generalize this. We can generalize this that for any number of such sequence, any large of the sequence, my probability of finding ⁓ right most set bit to kth position will be equal to, it will be equal to 1 upon 2 raised to power k plus 1\. That's what here it's all about. 1 upon 2 raised to power rho plus 1\. If rho equal to 0 is what we are going for. So it's k.

Sneha Mehra (00:04:50)  
So 1 upon 2 raised to power k plus 1\. ⁓ So that's what we are going over here. ⁓ So if we do this, here what you can, ⁓ in order to visualize this, you can look at this, that my probability at LSB, like me finding ⁓ the rightmost bit set at the least significant bit here is equal to half. That's what we saw. ⁓ At each step,

it reduces by half again so 1 by 2, 1 by 4, 1 by 8 and so on and so forth if I have 32 bits it would be 1 upon 2 raised to power 32 so 2 raised to power 1 till 2 raised to power 32 right so I can have as big of a filter as I want but I would be able to put that and all I'm keeping track of is the right most set bit right so which is what we are doing so here the probability so given

the stream of elements that I getting if I do something with the rightmost set bit this is the expected probability distribution that I would be seeing half 1 by 4, 1 by 8, 1 by 16, 1 by 32, 1 by 64, 1 by 128 so on and so forth till the length that you have. So now what do you do? So here the idea is for all the incoming elements that you are getting you pass it through the hash function some hash function ⁓ that kind of

splits it into a uniform distribution kind of hash you use a hash function you hash the value out the incoming value ⁓ it is an integer right this value that you would now get will be an integer and the distribution is uniform which you'll get that's what the job of the hash function is so you'll pick something like murmur hash or a or a multiplicative hash function something like that and then you record the rightmost set bit of it that's your job so given that you get a hash value out of it you find the rightmost set bit of it

Once you find the rightmost set bit of it, in the filter that you have, set that corresponding bit to 1\. So for example, if I put an apple, let's say apple got hashed at 7, right? So 7 is the integer value that it got hashed to. Now 7 is represented as 1 1 1, the rightmost set bit is 0, right? Because it's at the 0th position. So I'll mark the 0th position as 1\. Then banana, banana hashed to 4\. 4 is represented as 1 0 0\.

Sneha Mehra (00:07:10)  
So my rightmost set bit is set to 2 because 4 is 0, 1, 2\. So here I am setting it at 2\. Let's say got banana again. It would again hash at 4 because hash function for the same input will get the same output. ⁓ 1 0 0 position is 2\. It would be setting it over here and all other fields are 0\. Right? So now given this as the information, given this, given this filter, what can you say about the number of unique elements ⁓ that you would have?

⁓ So here if you look at this, if you look at the probability of like a simple probability if you take a look at it, given that our distribution is random because for any given input we are passing it to the hash function, hash function would be near ⁓ uniformly distributing it. So for a uniform distribution we know that the probability of V getting the right most set bit at LSB at bit 0 is half, then it's 0.25, it's 0.125 and so on and so forth. Which means that

at certain stage it would become zero. ⁓ Like when it becomes zero, that would be the place ⁓ where you would be having those many like two raised to that bit ⁓ number of elements. ⁓ I'll say why, why, why this thing? Why it's two raised to be only? The idea is pretty simple. ⁓ If there were, let's say you have filled in, if I take this example itself, if let's say this

I start from this probability half ⁓ 1 by 4, 1 by 8, 1 by 6 and so on and so forth. When it hits zero at a certain state where I have not seen any rightmost set bit at that location, right? What this helps us deduce? This helps us deduce the fact that there would not be ⁓ more than these number of elements because anyway we are doing approximate, we are not accurate in the answer but we are approximating with a very less space. So,

If there were more than ⁓ unique elements then the bit B would have been set. ⁓ Because if we are finding that bit B where the rightmost bit, where the bit is set to 0 in my filter, it implies that if that, if I had more than elements then that bit would have been set and I would be going to the next bit after that. ⁓ That's the idea or rather that's the reason behind it that for the first rightmost unset bit

Sneha Mehra (00:09:34)  
that I find is the best approximation of the cardinality of the elements that you are looking at. Because your distribution is there once you hit 0 which means you have not seen any more number. Right? So if you look at this because you are going by the probability half 1 by 4, 1 by 8, 1 by 6 and so on and so forth the number of elements that you would have seen would be 2 raised to power this bit. That's the best estimation over here. Right? So now if you plot this thing so and let's say if you are this is your true

unique counts then if you use pleasure let martin it would vary and obviously it's not approximate but it closely follows the curve unknowingly because ⁓ at scale probability works right estimation works this is what you we see when we use this in practice at a very fractional amount of space they are able to approximate the cardinality ⁓ of my incoming stream of numbers so that's what and that's how

And this is the foundation of Redis's HyperLogLog. This is what it uses to implement it, or rather to add support for HyperLogLog. So this is a very crude optimization. Like this is a very crude implementation. Obviously a lot of things have changed. ⁓ Some optimizations that I could think, or some optimizations that there were published in the papers were that obviously there is an error rate. We can see this fluctuating very much. And it...

can or not always be a power of 2\. So there has to be an approximation to that. So to do that error, ⁓ to account for that error, we don't just say it's 2 raised to power b. We said 2 raised to power b divided by 0.77351. It has a huge theoretical point. We can find it in the original research paper, in the Maurizis research paper. But, ⁓ sorry not Maurizis, but basically Fragilette Martin's research paper. I'll link it in the description down below. ⁓ It's one of the things that it does.

to approximate. Other thing is instead of using just one hash function, ⁓ use multiple hash functions and they take average of the estimations. And third is instead of taking average which is susceptible to large values, you can just take mean of it. ⁓ The mean value of, let's say you use three hash functions, take the mean of it. ⁓ That is what could be your estimation for that. ⁓ But this is the idea behind hyperlogarism. This is how it works. This is such a beautiful piece of algorithm. ⁓

Sneha Mehra (00:11:58)  
connects ⁓ like your probability with your set cardinality to seemingly unrealistic and unrelated concepts coming together so beautifully well. And what Redis does is we'll not go into Redis, it's highly complex. But what it does is it has two implementations dense and sparse. The algorithm still remains the same, right? But depending on how it stores, there are a lot of nitty gritties into that. But understanding the core idea behind hyperloglog is very essential, which is what I've covered over here.

You can feel free to skim through the source code of Redis. It's there in the file called hyperloglog.c. You can find it there. Right? And obviously, where do it store? Where does it store hyperloglog? Because it's just a random set of bits. It can store it in the raw encoding. Right? Which is what the beauty is. That's why raw encodings are so important. You can put in any kind of byte filter within that. It's just a bunch of bytes for your system to process. Right? So this is how Redis does it.

Again, you can check the source code ⁓ at hyperloglog.c file. I'll link ⁓ the Redisr source code you anyway would have downloaded up until now. It's there in the hyperloglog.c file to see the actual implementation. But the algorithm ⁓ of doing hyperloglog is exactly this. ⁓ I'll put a couple of links for you to understand this in depth in the description down below. So I hope you found it interesting and amusing. ⁓ That is what I wanted to cover this one and I'll see in the next one. Thanks, Hattar.

—---------------------------------------

8  
Sneha Mehra (00:00:00)  
So first of all, let's see how our synchronous TCP server is not concurrent. So that when we make it asynchronous, we see that in action. So on the top left, what you are seeing ⁓ is our own Go language implementation, which I'll start with gorun-main.go. ⁓ On the top right, you see that classic Redis server running on port 6379, our server runs on 7379\. So if I move down and I connect,

my Redis CLI to port 6379\. ⁓ I am able to connect and I am able to send ping and in response I get pong. So here I connecting to normal Redis server. So if I connect to port 6379, ⁓ get, I am able to connect to this. So both the clients at the same time are connected to the same Redis server, the actual Redis server. Right? Now let's see what happens when ⁓ we try to connect it.

to our own implementation. So our implementation runs on port 7379\. If I run that, first client got connected. If I fire ping, I get pong. And on the second one, if I connect it to port 7379, the connection did not happen. This is where it got stuck. ⁓ This shows that we can support one concurrent client only. So when the first client disconnects, then the second client got connected.

This is the limitation. This is the classic limitation of a single threaded application where your, because the process is single threaded, that one thread ⁓ is blocking, is being blocked because while it is executing, while it is reading, ⁓ while it is basically continuously connected to a client, continuously reading a command ⁓ over that socket, it's not getting a chance to accept other connection. Right? So this is what, this limitation is what we have to sort out.

So now first understand why it is blocking there, why it could not move forward. So if I come to sync TCP implementation, here I was talking about the infinite loops, right? So the previous in ⁓ a couple of videos back, we saw this very implementation. So here you can see that when I am running my synchronous TCP server on port 7379, you see that we are listening on a particular host at a particular port, right? So basically on port 7379 we are listening.

Sneha Mehra (00:02:25)  
And then I have this infinite for loop where we are accepting the connection. Now this accept is the system call that lets your server accept a socket correction from one of the client. So unless this gets executed, unless this gets executed, I will not be able to accept any more client. Right? But when I accepted the first client, it went in another infinite loop which says that read a command ⁓ and respond.

So as I said, read and write both are blocking calls, both are blocking IO calls. So this is blocking and this is blocking. So now what would happen? It ⁓ is transactional. that client send me the command, then server sends the response, then client sends me the command, then server sends the response. So this for loop would break only when my client disconnects or only when ⁓ I issue a quit command or something. We have not yet implemented, but that would be the general implementation detail on that.

Right? So that is where you see the program flow because it is signal threaded. It is first stuck in the first infinite for loop where it is accepting the client. Once the first client accepted the culture, it's stuck into the second infinite for loop. Right? This is what we would want to solve, which means we would still need to have an infinite loop, but this listener.except that we have, we want our control flow to reach here ⁓ as soon as possible. Like it should not be waiting over this. So this

single threaded synchronous implementation would not work because we want to large number of concurrent requests. This is where we would be implementing our async TCP server. So first let me change this from synchronous to asynchronous. Now I'll take you where is this implemented. This is implemented in the async TCP file, ⁓ async underscore TCP file.

All of this source code is you can access it on the GitHub repository, github.com slash diceDB slash dice. Go to the fourth commit. In the fourth commit, you'll see this implementation there. So this async TCP server is what a normal TCP server is doing, but in an async and a swap. This is where you will see IO multiplexing in action. ⁓ Right? Okay. So let's start from the function execution itself. So here,

Sneha Mehra (00:04:41)  
we would see epoll, epoll-wait, epoll-create, epoll-ctl inaction. These are raw system calls. And again, this is a Linux based implementation. ⁓ So if you are using Mac, you have to use KQ. If you are using Windows, you have to use IOCP. Syntax would change. I'm urging you to do it on Linux because epoll is the easiest one of them to understand. ⁓ Very simple. You'll find a ton of tutorials and the man pages are excellent about it. Right? ⁓ Okay. ⁓ So run async TCP server. So now here, ⁓

in the main function, we are, instead of starting a sync TCP server, we are starting an async TCP server. So as my run async TCP server starts, we are first printing the command, it says starting an asynchronous TCP server on a specified host and a port. Then we are setting a variable called maxline. This would come in handy in a couple of minutes. Then we are creating something called as an epole event. Now this epole event ⁓ is the event that we need to like,

as I mentioned in the last video that we have to register our file descriptor in E-Pole. Now, when we are registering our file descriptor, we have to specify some parameters. For example, ⁓ what are we trying to monitor? How are you trying to monitor? All of that. when some file descriptor is ready for an I.O., we would be getting it in this buffer. These are the file descriptors that are available. So, I setting this 20,000 limit.

for now ⁓ and what we are doing is this buffer would be holding the file descriptors that are ready for an IO by E-Poll system call. This events is for us to hold that information that hey these are the IO's that are ready. So you may register 100 but if only two are ready this buffer would hold two events that are ready. ⁓

Now here we do raw socket handling because we cannot use abstraction because we want access to raw file descriptors. So that is where what we are doing is we are creating a socket, a normal socket, which is an IPv4 socket, which is non-blocking and a socket stream. So socket stream typically tells that I don't want to disconnect my TCP connection ⁓ as soon as I got a reply. I want to keep it open.

Sneha Mehra (00:06:55)  
This is a streaming connection. want to keep this TCP connection open. don't want because this is what Redis client ⁓ or any Redis CLI also does. It keeps the connections open. ⁓ And I want this to run in a non-blocking mode. This is not a single node. This is non-blocking TCP level non-blocking mode, ⁓ which you can find documentation on its own system call. But the idea for this is while ⁓ one TCP connection is sending the data, other TCP connection ⁓ can actively receive the data, but it would move forward. ⁓

user space would still remain blocked. And then in case I'm not able to pursue or in case there is an error while creating a socket, ⁓ I'll return an error because I cannot run this server anymore. ⁓ And I'm doing a close. I'm doing a deferred close, which means that once the function execution is done, please close the socket. ⁓ Then I'm setting my socket to non-blocking. This is my server FD, which is the socket server. The reddish server's file descriptor is what we are getting in the response. This is what we would want.

to be monitored by E-POL. This is the server FD, which means that the TCP server that we running on port 737 and this is the file descriptor of that. Any incoming request on that server would come to this file descriptor. ⁓ So now then what we are doing is, ⁓ this is my slight problem with syntax. It's just that I want to bind the port.

bind the host and the port. ⁓ Now here what it expects is it expects a four byte representation. So basically whatever the IP that you let's say you say that I want to listen like I am host local host and I want to listen to port 7379\. So local host is 127 001\. So this is a string for us but in the system call when we pass we have to pass an integer array of length four and with the first byte is 127 then zero then zero then one.

So this is just a way to do that. ⁓ So we are saying that bind our server socket, bind our file descriptor of the socket that we just created ⁓ on port 7379 ⁓ and on address 0000 or 127001\. And then we listen. Here our server starts to listen ⁓ on this file. Now this file descriptor is binded with a port and an address. So this is where it is start to listen. So when you typically write

Sneha Mehra (00:09:14)  
⁓ Start my server listening on a particular port. We directly passing that addressing behind us in all of this happens, right? So here my server would start to listen ⁓ and with a backlog of max client. So at max this is the backlog that we are putting in. Again a little heavy on the networking side. You can read the man pages to understand this particular part in depth. But with respect to asynchronous, I were going much more in depth on how it actually happens. Once your server is starting to listen over here. Now what?

Then your async IOS starts. Now your server is ⁓ ready to listen to incoming sockets. So we create an E-Pole over here. So we create an E-Pole using the system called E-Pole create one ⁓ and we get an E-Pole file descriptor. So as I said, everything has a file descriptor. So E-Pole itself has a file descriptor of its own. This would uniquely like you may create multiple E-Poles ⁓ and each of the E-Pole would have a unique file descriptor for you to

Identified right so if all file descriptor holds that so this is one equal that you created and now what do you want to do you want? The server that we just created a file descriptor it to be to be monitored as well because this server itself will be getting Connection from the client so you want this server to be monitored so you have your server file descriptor You are saying that for any incoming

events for any incoming events. For example, for your server, the critical event that you would want to monitor is whenever a client is connecting to the server. That is where you would want it to be triggered. Right? So this is what we are doing. We are creating an E-POL event over here, where we are saying that the events that I would want to monitor is E-POL in, which means whenever there is an incoming, where even there is a read, I want to be notified about it. And the file descriptor is server empty.

which means anytime an incoming request ⁓ is coming to my server, which means anytime a client is connecting to my server, trigger, basically notify me. ⁓ So this is the event object that we created and here we are registering it in the epoll. So using epollctl, you are registering that in this epoll, which is epollfd, in this epoll ⁓ add epollctl ⁓ add implies

Sneha Mehra (00:11:36)  
register this file descriptor to be monitored, add this file descriptor to be monitored. Which file descriptor? Server FD. And which event? This event. So the server FD will be monitored for E-POL-N events. Any other event, you'll not get notified. You'll only be notified for E-POL-N. Right? And this is where you're registering your main server. Your main TCP server running on port ⁓ 7379 is now getting monitored. So whenever a client is connecting to it,

your epoll if you fire an epoll wait it would return you whenever a client is ready to connect to this server. ⁓ Okay, so now this is where you have set your server to be monitored by epoll and now you have this humongous ⁓ humongous but a good for loop and now what this for loop would do is what we have to do is we have to constantly monitor if any ⁓ IO is ready. So for example,

I am the basic that is why I invoking epol wait which is a blocking call. So epol wait ⁓ is saying that hey let me monitor. So epol wait is monitoring epol fd the one epol that we created we are monitoring it and we are saying that whenever any event is there please put it in this events buffer. Right so epol wait it monitors for any io which is ready and it would put it in the buffer. ⁓ Which buffer? The events buffer. If there is none no io is ready.

this call would block. It would remain blocked as is. But once this completes its execution, whenever some IO is ready, in the response you get n events, which is the number of events. Because you may ⁓ send a buffer of length 20,000 like we did, but let's say only two are available, right? So this n events would tell you how many are available within that so that you don't read the other one. So you may provision a large enough buffer size, but you will only

get or you can only you should only read the two out of it. This is what any events would do. So now that you know that after the epal wait completes its execution, you would know how many events are there that you would want to consume because how many IOs are ready. So you would fire a normal for i equal to zero, i is less than any events i plus plus. ⁓ Once you do that, you are accessing each event. So you are accessing each event and each event

Sneha Mehra (00:13:59)  
has a file descriptor, which means which file descriptor is ready for IO. Correct? So this is where what you are doing is if the file descriptor, now here you will check that because ⁓ if it's a client who is ready to be connected, so then your server's E-POL would be triggered. So if your server is ready for an IO, which means that any new client is willing to connect to this server, this is when this would be triggered because you monitor it for E-POL in. Right?

So if events of i.fd ⁓ double equal to server fd, which means that if it's my server is ready for an IO, which means my socket server 7379 ready server is ready for an IO, then I have to accept ⁓ the incoming socket. ⁓ Then we'll do the normal part, but let me quickly check on the else part because else is ⁓ a little simple then we'll come back to the if part.

So if it's my server, then accept the socket and then do something. And if it is not my server, which means that some client that is already connected to the server, right? If that is one, then do something, right? Because if my server itself is ready for an IO, which means that there is some kind, which means that I'm yet to expect, sorry, I'm yet to accept the connection from the client. This is where the first part would execute.

that if my server itself is ready for an IO, right? And if client is ready for an IO, then we'll do something else. Now we'll see the first one now. So when my server is ready, which means that ⁓ there is a new client willing to connect to this server. Then what you do? You accept. You accept the incoming socket connection ⁓ and you get in return, you ⁓ get the file descriptor. This is the socket file descriptor between your server

and the new client that got connected. And now what you want to do is you want to monitor this socket as well. Yeah, like you have to monitor this is the server to client socket. You want to monitor this file descriptor ⁓ for any IO, which means that when client would send you something, your E-Poll should know that, hey, there is some data that a client wants to send me. How would it know? That's why you would want to register it in E-Poll. Just like how we registered it, the server FD, we would register this particular FD.

Sneha Mehra (00:16:24)  
in our epoll. So this is what we are doing. We creating a socket client event in which we are creating an epoll event called epoll in and in 32 FD and with epoll CTL we are adding it to be monitored which the socket connection that just got established. So this way the first thing what would happen is when your server starts up we are adding the server socket in the epoll to notify it whenever a client is ready to connect your epoll would trigger and epoll wait would succeed it would give us that hey now your server is ready for IO.

When server is ready for IO, it means that there is a new client to be connected. When there is a new client to be connected, you are establishing the socket connection with XSafed and whatever file descriptor you've got, you are adding it into E-Poll so that it also gets monitored. Now next time, two possibilities can happen. Either your server is ready for an IO or this client is ready for an IO, which means either your server file descriptor is ready for an IO or this socket file descriptor is ready for an IO.

So again, if a server is ready for IO, which means there is one more new client who is willing to connect. And if your socket is ready for an IO, which means that there is a client who wants to send data to you. And that is what this else part is all about. So here in the for loop, where we are checking for any events that are coming in, if it's a server event, then accept the connection. That's the first priority, accept the connection right there and then. If not, then this must be some client who is doing it. So some client wants to send data to you.

Right? So now, when that is the case, what do we do? Because when there is a client who wants to send data to us, what do want to do? We want to read the command and respond. That's what we do because client has some IO to do, which means clients want to send data. What client would send? Client would send a command. And what would we respond? The output of that command. Right? This is what we are doing. We are reading the command ⁓ and we are responding. Like how we did with our synchronous TCP function.

Right, that same flow happens over here. We take read command ⁓ from the connection ⁓ and respond to the connection. Right, but just one small change that I did ⁓ in the previous code where instead of because here we are dealing with raw file descriptors. In here when we see we see we get events of i.fd is a file descriptor we are getting. In the synchronous TCP call if you see what we are getting we are getting an abstraction called

Sneha Mehra (00:18:51)  
⁓ Here we are getting a net.con object which is the abstracted socket connection. But here what we are getting with asynchronous IO, we are getting a file descriptor events of I.fd. So which means that the read command that we wrote

we are reusing the same command, the read command, here you can see the file is synchronous underscore TCP, we are using the same thing because at that it just IO that we are doing, either we are reading it from the socket through the socket abstraction or through the file descriptor, both are exactly the same. That is where instead of accepting a net dot socket, sorry, net dot con object, I am accepting IO dot read writer object, which means anything that could read or write, be it file descriptor who can read or write.

or be it socket hook and read or write. I've just changed this type so I can reuse the same function for my file descriptor based IO ⁓ or my normal abstracted socket based IO. Right, just that one small change that I did and IO read writer is nothing but just a simple like anything that could read or write. So I created one more file called com.go in which I created a file descriptor communicator. The simple thing is it can write and it can read has the exact same signature.

Right, so that I can pass in a normal socket connection though your net dot con object also ⁓ and with async IO, I just have to create a new object called file descriptor communicator where I'm passing this file descriptor and internally your read command does that exact same thing what it did it read from it c dot read.

I just expose that same function over there. C.red so that it read from it and everything else remains the same. So read write signature, I've implemented it for file descriptor implementation. Socket implementation was already there. File descriptor I implemented by firing that same system call. I'll just quickly show you that. Here you can see that the write function ⁓ on my FD communicator does just syscall.write. So because I do not have a socket connection, I have to make a system call of write ⁓ over

Sneha Mehra (00:20:55)  
the file descriptor with the bytes that I got. I just had to re-implement the read and write command by ⁓ encapsulating the write and the read system call. Everything else remains the same. This is just I add so that I can reuse the functionality that I already built, the read command and the respond command as is. So here, when my client is ready for an IO, I'll get the command, I'll read the command and then...

I'll respond. in the respond, I'm again doing the same thing, eval and respond, eval and respond gets the command, gets the IO writer. If it's a ping command, it evals the ping. And what ping would do is ping gets the same IO writer and it would just invoke dot write function on top of that. So C dot write of B, ⁓ no code changes there. The three or four files that got changed are main.go where we changed from sync server to a sync server.

Then we added this gigantic file async underscore TCP where we implemented our own event loop. And then we added this com.go file in which we implemented the read and write interface. And in eval.go, we just instead of sending the net.con object, the abstracted socket connection, we are sending simple IO read writer, which means anything that can expose a read or a write system call would work just fine. Right? So this is the bare minimum that

the changes that we did to make our server concurrent. Now, let's see this action. Enough of this code thing. Let's see this action on how this thing ⁓ beautifully works. So now here what I have ⁓ is what I have is I'll just have to restart my server because now we are starting our asynchronous TCP server. Here we are starting our asynchronous TCP server. You can see here it's written asynchronous TCP server. And now let's see. Earlier we saw we were not able to correct

concurrently with 7379, but now it would work. ⁓ I connected my first client over here. I passed in ping and in response I got pong. In the second one when I connected earlier it was not getting connected. So let me reconnect. ⁓ Second connection also got accepted earlier. was not also getting accepted at all. But now here

Sneha Mehra (00:23:10)  
You can see that both my clients are concurrently talking to my same TCP server, the 7379\. It's not 6379, it's 7379\. So we just made a single threaded asynchronous Redis implementation where we are still able to listen to multiple clients and handle multiple requests in parallel. ⁓ Now, just to see how we fare against Redis.

So let me quickly fire a Redis benchmark. We love Redis benchmark, don't we? ⁓ So Redis benchmark. ⁓ So here what I'm doing is let's say I fire Redis benchmark for this specific use case. So again, what I'm doing, I'm triggering a Redis benchmark that fires 10,000 requests of ping request because that's the only command that we implemented till now, ping request with number of concurrent clients as 200\. 200 concurrent clients on port localhost colon 6379, which is the actual Redis server.

So if I fire this, this request goes to actual ready server and it handle 37,735 requests. Now instead of 6379, I do 7379\. And this completed in 36,231. So this shows that with our server, the Go language implementation that we wrote is able to handle 36,000 requests per second. ⁓ And it's all our sync credits, it's all single threaded. And this is what event loops.

asynchronous IO is all about. What we just did is we had a single threaded Redis server and we added asynchronous IO using normal E-Poll system call and made it to handle multiple TCP connections or multiple clients at the same time. ⁓ We get command from one, we execute it and we respond and so on and so forth. And this happens in such a beautiful way. And this is what is all about asynchronous IO, asynchronous network programming, event loops and in general.

So I would highly, highly, highly recommend you to build the same stuff again, like re-implement this. ⁓ You can find the source code at github.com slash diceDB slash dice. Fourth commit that you will see, ⁓ like fourth commit from the start, you will see this, all of these changes in one shot there. But I would highly encourage you to implement this on your own. It's extremely simple, but when you'll do it,

Sneha Mehra (00:25:29)  
You will understand so much about network IO. You will learn so much about IO, how to design, how to encode, how to decode, how to do. And this is all about low level implementation as well. You are understanding the smaller implementation details. ⁓ And that is it. That is it for this one. I hope you found it amusing. We just created our concurrent Redis server. I was so excited while building this and I hope you felt the same way. In the next video, now that we have

concurrency in check, we have IO multiplexing in check. The next video will be implementing the simple get and put. So now we'll be going in the direction of building a simple key. Up until now, the only command we support is ping. In the next video, we'll be supporting get and set. This way, we would have our own key value store, our own in-memory key value store, which works just like Redis. Right? And that is it for this one. I hope you found it amazing ⁓ and I'll see you in the next one.

Thanks a ton.

—-------------------------------------  
7

Sneha Mehra (00:00:00)  
So the end of the previous video, we saw how we were not able to connect concurrent clients to our Redis implementation. So what can we do so that our single threaded Redis implementation can support concurrent clients? The answer to this is the buzzword asynchronous IO. So let's see how can we handle concurrent clients? Like what is the traditional way to do it and how to do it with asynchronous IO? So the most common way, the most common way of supporting concurrent clients is that

Whenever a client connects to the server via socket, you initialize a new thread and let that thread handle things for you. ⁓ Because threads can be scheduled on multiple CPUs, they would run concurrently. So ⁓ in most cases, your systems would not be blocked. So while one thread is doing its job, if it's blocked, it goes into the wait state, then the other one gets scheduled and so on and so forth. This way you'd be able to support large number of concurrent clients. ⁓ But...

The key challenge with this is if you are using thread, you have to make your code thread safe. Like a very classic example is that let's say you have a global variable called count and this count variable is not at all atomic because count plus plus is not atomic. So what would happen is if two threads try to update or try to fire count plus plus accessing the same global variable count, if your count value is 10, if two threads are doing plus plus.

it is very much possible that the end value could be 11 or 12 because two threads can read the same old value, do plus plus on its own and then write. So both would read 10, both would create 11 and then both would write 11 to that variable. And this is where you know that count plus plus is not at all threads. So that is where what you have to do is you have to use apply explicit blocking like mutexes and semaphores to protect your critical section so that your

shared variables, your shared memory does not go in an inconsistent state. That is number one problem. The number two problem is threads have to unnecessarily wait. So for example, ⁓ for IO, your thread is anyway waiting. For example, disk IO, network IO, any kind of IO, your thread is waiting. Once it receives something, it is free to move forward.

Sneha Mehra (00:02:19)  
But because you have a critical section, it would be stuck there again because in that critical section only one thread is allowed at a time. example, count plus plus, a classic example, right? A classic example of count plus plus where all threads are waiting and only one is allowed to do count plus plus at a time. So even though your threads were very well ready to be executed concurrently, you are not letting them to do so because you have a good amount of critical section. And this would slow things down.

this increases the code complexity because you have to everywhere where you are writing it, have to keep in mind that the code is multi-threaded and now you have to ensure that nothing goes wrong because your data if your in-memory data went into an inconsistent set, it's so, so, so difficult because you'd never know what the truth is. ⁓ So the answer to this is asynchronous IO or IO multiplexing. So ⁓ the

IO multiplexing is how every single event loop that you have heard of is implemented. So think about it. Be it Python async IO, be it JavaScript. Why do you think that like it's only for IO? Why can't you do CPU? Like why can't you do count plus plus in a separate event loop? Why are you only able to do IO in a separate event loop? So most people think, hey,

Then is my event loop a separate process? It cannot be because when you run Python space something or you run node space something, you don't see two process spin up. You don't see one node.js thread and one event loop thread spin up. So then ⁓ your event loop is definitely not a process. Similarly, if you see that, if you say that your Python JavaScript is single threaded. So then is event loop a separate thread? If it is a separate thread, then your language is not a single threaded language.

So then that's wrong. And if it was a separate thread, why can't we schedule normal CPU instructions to that? Which means that this is neither, your event loop is neither a separate thread nor a separate process. It is just ⁓ so thin layer that it's able to support only IO. That is it. And this functionality of you being able to support multiple IO is

Sneha Mehra (00:04:45)  
has to be provided by your kernel via system call interface. So kernel needs to implement something so that you get notified when something's about to happen on an IO. This is the core idea of an event loop. And be it Python, be it JavaScript, be it lib event, lib UV, you have heard of any and every event loop out there. It's implemented with this exact same system call because that's what your operating system exposes, system calls.

and that's what you can use. So everything that you have seen around event loops asynchronous programming, it's all, it all boils down to this. Right? So how it is actually implemented, the system calls that we are talking about, there are typically three depending on which operating system or which flavor of operating system you are using. You can either use E-Pole. E-Pole is typically Linux, Unix based, ⁓ like it's basically way to get node-wide, we'll go into details of it.

but Epole is Linux specific, KQ is a BHD specific, basically Mac, ⁓ if you are using, you will have to use KQ and if you using Windows, you have to use IOCP, right? So depending on which language or depending on which operating system you are using, you will picking one of these three. And throughout this course, we'll be focusing on Epole ⁓ because we are assuming that it's Linux based, but your code and the flow remains exactly the same, just system called interfaces.

and its corresponding arguments change. Everything else remains the same. So if you're using ⁓ Linux, go for E-POL. If you're using Mac, go for KQ. If you're using Windows, go for IOCP. So what is the ⁓ idea ⁓ of E-POL? What it actually does, because E-POL is the one that is doing all the magic of making your single thread application support large number of concurrent clients. So what E-POL does? E-POL does one very simple thing. It can...

monitor a lot of file descriptors for new IO. That's it. That's the simplest explanation of it. So for example, ⁓ sorry, in Unix, in Linux and in most operating system, you would see everything is a file. If everything is a file, then it would have a file descriptor. So the idea is in E-Poll, you make like you pass in all the file descriptors that you would want to monitor. ⁓ And E-Poll would tell you,

Sneha Mehra (00:07:11)  
which one of them are ready for new IO. This is how it leverages. We'll go into detail of the flow and build intuition on how we can make a single threaded Redis implementation concurrent. ⁓ But before we jump on to that, let's talk about how IO typically happens because this is where the magic, like this is why it is possible for E-Poll to even exist. ⁓ So let's say if we talk about pure networking terms, where you have a client,

who is connecting via a socket to your server. Now socket is a logical entity. In reality, the client is connecting to a network card, maybe your ethernet port, maybe your Wi-Fi card, right? Your client is connecting to that. Now that is in hardware. From this, the data needs to go into your kernel buffer because the data would always be first received in the kernel.

So for example, if I'm sending, if you're sorry, if a client is sending any data to your server via socket, it comes to the network card, then it goes into the kernel buffer. But how it goes into the kernel buffer? There is an interrupt. As soon as the packet hits the network card, your network card creates an interrupt or it basically triggers an interrupt so that your kernel stops doing everything else and reads from the network card so that the data now is placed in a kernel buffer.

It's a kernel space. We like a normal user application code would not have access to this kernel space. ⁓ Right? Now the data is present in kernel buffer. This is what it is up until now. And now what happens? You might have hundreds of process running in your application space. Maybe your VLC player, maybe your Node.js code, some Python code, some this packet, that packet, something or browsers and whatnot. And your CPU has limited cores, maybe four core CPU.

⁓ if you are using. So which means at max concurrently four processes can move forward. Right? So now what would happen? Whenever now this data came over the socket and is placed in the kernel buffer. And now what would happen is when your process ⁓ is getting the CPU to execute itself. So when your process is scheduled on the CPU, ⁓ the data will be copied from your kernel space to your user space.

Sneha Mehra (00:09:33)  
and this is how your socket that you are managing in your code gets the data because from the kernel buffer it is copied to the user space and then you are able to access it. Right? So the first one is an asynchronous thing where your client some client connected to your server is sending the data and your kernel buffer has the data but until and unless your process gets scheduled it will not get the data obviously because while you are executing then you will check

Like when you are accepting the connection, when you are reading from the kernel buffer, ⁓ this is where your application code is doing. If there is any IO available in the kernel buffer, then also if there is any data available in the kernel buffer, I will copy it for me. Right? And this is how your IO happens. Now, what does this tell us? This tells us that because of this copy step in between, ⁓ there is a time where your kernel knows that for this process, I have some data with me.

because your kernel knows that the data that it receives over this socket for this particular process. Your kernel knows that. Correct? Because your kernel knows that your kernel can tell if some IO is ready because this IO is not with respect to your end client but with respect to the data availability of the process in the kernel space. And when this happens your kernel can tell

if some IO is ready. This forms the heart and soul of EPOL, KQ and IOCP and asynchronous programming in general. So this is how EPOL, KQ, IOCP works internally that they know that some IO is ready because this is a system call interface. You can use EPOL and EPOL would tell you ⁓ if there is an IO ready or not. Right?

And how would it tell you? Because if kernel buffer has data for a particular process, E-POL ⁓ job is because E-POL is a system called interface, which means it can, it is interfacing with your kernel and your user space. So it has access to both. It would check if there is anything for this, ⁓ for ⁓ your application process. If there is, it would say that IO is ready, right? So first flow is also kernel, second flow is it is checking. E-POL has access to do that. And this is how your

Sneha Mehra (00:11:56)  
I O works and how system call interfacing is so important in powering a safe I O. So let's see the core idea. The core idea ⁓ of we implementing the, or we making a ready implementation concurrent is the fact that we will continue to do ⁓ our work. This is what we do. Continue to doing our work. ⁓ Once a while, check if some ⁓ I O is ready, right? Using E-POL.

We see how it is actually done. You check if some IO is ready. If some IO is ready, do the IO. If it is not, continue. This is the idea. Now, if someone's ready for an IO is given to or is told to us by E-POL. ⁓ And how E-POL will tell us, this is where the concept of file descriptors come in. So in Unix, everything is a file. So ⁓ everything is a file and

Every file has a file descriptor. So your socket has a file descriptor, your disk IO because you're writing to the normal file that also has a file descriptor. Any memory buffer has a file descriptor. Anything related to IO, devices, external devices, USB devices, ⁓ every single thing is a file. So which means that they have a file descriptor. So now what, and a file descriptor is simple integer. It's just a simple 32-bit integer which tells that, ⁓ which uniquely, which basically uniquely identifies the file

within your system. It's not a normal text file that you creating. ⁓ It looks and feels like that. Even a socket that you creating is a temporary file that is created with a file descriptor. So that's the abstraction that it plays with. So ⁓ if you don't understand this detail, just look at this. ⁓ Take this abstraction that everything in Unix is a file. Even a socket is a file ⁓ and any IO that you are doing goes through this file. ⁓ And every file has a file descriptor.

So now what are E-Poll does? Now E-Poll, the job of the E-Poll is pretty simple. What you should be doing is, whenever you are connecting or whenever a client is connecting to your server, you get this file, you get because it's a socket. Socket is a file, file is a file descriptor. You send this file descriptor, you register it with E-Poll saying it, hey, please monitor this file descriptor for me. And that is what E-Poll would do.

Sneha Mehra (00:14:22)  
So you'll create an epoll for yourself using epoll underscore create one system call. This is the actual system call by the way epoll underscore create one. You are creating an epoll and then you are asking your epoll to monitor a bunch of file descriptors. So I've written client socket one, client socket two, client socket three and so on and so forth. When server socket. So here what you are doing is you are asking epoll to monitor file descriptors for any IO. So

Whenever a client connects to the server, you are registering it in an e-Poll saying, ⁓ people, please tell me if any one of them is ready for an IO. ⁓ And along with all the clients, your server itself, like your socket server that you are writing, your 7379 ready server also needs to be monitored because it, this server is getting an incoming connection from the client, the socket connection from one of the newer clients. So this

file descriptor also needs to be monitored. So you add the server socket as well and all the client sockets over here. Right? So this is how, this is what you will do to register and deregister a file descriptor and the system call you'll use is called epoll underscore CTL. Right? epoll underscore CTL ⁓ using this, you would be registering an IO, or sorry, you will be registering a file descriptor to be monitored. Right? And now how would you know

if some of them is ready for an IO. This is where the third system call comes in called epoll underscore wait. Now this is a blocking call, which means that let's say if you send in 10 file descriptors to be monitored and if you call epoll underscore wait, you call epoll underscore wait on an epoll. That epoll has multiple file descriptors. Let's say 10 file descriptors. ⁓ If any one of those 10 file descriptors have some IO to do and how would epoll know it?

because E-Pole knows the kernel buffer. E-Pole knows that for this socket is there anything in the kernel buffer. Then that's how E-Pole would know if there is ⁓ IO, like if some IO is ready for you to consume, right? So what E-Pole would do is it is a blocking call, which means that if for the 10 things that you have, right? If for the 10 file descriptors that you have added in the E-Pole, if any one of them is ready for an IO, if it is ready, then it would

Sneha Mehra (00:16:48)  
list those file descriptors and proceed so that you can consume those file descriptors. So you will know at the end of epol wait if this call if this function call completes you would know that hey these are the file descriptors who are ready for an IO. Let's say you got client socket one two and three everyone else is blocked ⁓ everyone else is not ready because your client is yet to send the data and you are waiting on that. So one two and three right because you have this.

You got that ⁓ client socket 1, 2 and 3 are ready for IO. Then your code will move forward. You will get these three file descriptors and you can then read from this file descriptors like you normally do. And this is how you proceed. Now, for example, if this exhausted, let's say ⁓ none of the file IO or rather none of the IO were ready. Like you are just waiting and waiting and waiting. No one's there. E-POL is a blocking call, which means the code will not move forward until there is one or many

⁓ file descriptors who are ready for IO. So that's a blocking call. And this is the core idea. So now what you can do is because of this, you can run a simple infinite loop that is continuously checking for or that is continuously firing EPEL wait to check for any of the file descriptor if they are ready for execution. If they're ready for execution, you're taking the file descriptor, reading that value over the socket and then processing it.

Right, so even though your code flow is single threaded within that you are just continuously monitoring file because what as a ready server what you are doing you are reading a command from the client executing it in your server and returning it back. Right, so this is exactly what you are doing. So now in the infinite loop that you have you just check if there is any file IO that I can do or any IO which is ready if it is ready read that execute the command return the response.

and done. ⁓ Right? So this way you can support large number of clients over here. ⁓ Right? ⁓ Without having to have any kind of multi-threading because the system call interface is telling you which IO is ready. Otherwise if you have used, if you have not used E-Poll and you, ⁓ every IO call or every system IO call that you making is blocking. So read, write, ⁓ receive all of that are blocking system calls. Because of that what would happen is that

Sneha Mehra (00:19:14)  
If you are waiting, if you reading from one socket, you cannot read from multiple file descriptors at once in a single threaded system, which is what the problem is because that is blocking. So until and unless your client sends you something, you would not be able to move forward. But with epol wait, it is very beautifully telling you on which file descriptors, if you make a read call or a write call, you would get something. This is the beauty of epol. So this has made our ⁓ single threaded

⁓ are basically single thread implementation ⁓ to be able to support large number of concurrent users. And this is the idea of asynchronous IO, event loops in general. So JavaScript event loop exactly implemented this way, Python async IO exactly implemented this way, Redis implemented this way. Any place where you see event loops, this is how it is implemented. And that is not a separate thread. Here we never discuss threads. Event loop is not a separate thread, nor a separate process.

It is just a simple single threaded thing ⁓ using system call interface ⁓ that tells it which all file descriptors are ready for an IO. It's such a simple but so so so powerful concept. And this is what Redis uses internally. So yeah, that is it. I hope it was an exhaustive one, but I hope you understood the essence of asynchronous IO. And now if you like you should be very confident on what event loop actually is. In the next video,

we will implement this event loop and make our Redis server be able to handle concurrent users. Like two people would be able to directly talk to Redis server at the same time. We would also run benchmark on connecting on basically simulating 10,000 commands over 200, 300 concurrent clients and we would match the performance of our Go language implementation to the actual Redis implementation. Right? So that is it for this one. I'll see you in the next one.

Thanks a ton.

—----------------------------------  
10

Sneha Mehra (00:00:00)  
So the previous video we implemented get, put and TTL. In this video, we would be implementing delete, expire and we would see how Redis does auto expiration of keys, which means auto deletion of expired keys. Right? So first, like always, let's see what the behavior of a normal Redis server is and then we would be re-implementing it in our own Golang based implementation. Like always, top right contains the normal Redis server running on port 6379 and bottom left is where we are connecting it through Redis CLI.

Now, let say I set a particular key k with value v ⁓ right. Now, if I do a delete here you can see that delete is taking an argument which is key if I pass in k ⁓ the key would be deleted. The return value is 1 what does this 1 mean? This 1 is the number of keys which are deleted. So, delete if I do delete k again key k is already deleted what would be the return value? Return 0\. So, 1 is the number of keys that were successfully deleted.

because key k did not exist like after we deleted it does not exist anymore when we hit the delete ⁓ again on the same key it return 0 right. Now, if I ⁓ set multiple keys key k 1 v 1 ⁓ set key k 2 v 2 and now if I do delete k 1 k 2 k 3 k 4 two keys exist k 1 and k 2, but I am triggering delete of k 1 k 2 k 3 k 4 ⁓ it returns me 2\.

2 implies that out of this 4, 2 keys were deleted and 2 keys didn't exist. It's a no-op operation. So, now if I do run this command again, we are getting 0 because k1, k2, k3, k4 all 4 doesn't exist. Right? That's why it is returning me 0\. This is what we would have to re-implement. Right? So, this is how delete works. Second is let's see how expire works. Let me set a particular key k with value v and now I am triggering now what a use case of expire is. Let's say I set a particular key.

And while setting I didn't provide any expiration. ⁓ But then I realized hey now I want to set an expiration to this key. So that's where what I can do is I can pass in expire function with key k and the second argument is the number of seconds. Let's say I want to expire this key by 10 seconds. As I fire this now if I do TTL of k you can see the time elapsed or the time to live for that is consistently decreasing and decreasing and decreasing and decreasing.

Sneha Mehra (00:02:26)  
So, this is how your TTL was set with expire right. ⁓ Now, if I what happens if I pass in a key that does not exist. ⁓ Let us say I pass in a b c this key does not exist. ⁓ First because I did not pass second argument it give me syntax error. Now, if I pass in argument it returns me 0\. What is this 0 mean? ⁓ If the this 0 implies that the operation was unsuccessful. It could be because key

does not exist. It could be because any reason. if your Redis was able to set an expiration, it would return 1\. If it was unable to set an expiration, it would return 0\. So, here if I do a set k comma v and if I do expire k in 10 seconds, ⁓ it returns me 1 because setting like setting the expiration was successful. Right? This is what the return value of expire is.

Now given that we know how delete and expire behave, we would be re-implementing it on go-rank-base. But this is not done yet because we also like after we go through that, we would be taking a look at how Redis does auto-deletion. So if the key is expired, it would need to auto-delete the key, right? We would be implementing that as well. So that is a very interesting algorithm that we would look at at the end. So first let's start re-implementing this and see and do an extremely detailed code walkthrough. So what we would do is,

We would start with our normal eval.go file. Eval.go file is where we are implementing all the commands and here we add delete and expire command. Delete again all the arguments passed over here and an io read writer similarly for expire. So here if I check eval delete what do we see.

We will delete whatever argument we have passed are the keys and we just have to iterate over them and explicitly delete it from the hash map that we have the hash table that we have. We are just doing that. We are iterating over all the keys.

Sneha Mehra (00:04:21)  
and we are triggering a delete. This delete is a store function written in store.go file. The objective of this is so that we mask the any complexities that might aware, ⁓ sorry that might come up. So what it does is it just triggers a delete from the hash map store, delete the key k. Right? That's what it is doing. And it is returning true if the delete was successful and is returning false if the delete was unsuccessful.

and unsuccessful because key does not exist as simple as that. So if key exists, we are deleting it and returning true. If key does not exist, we are returning false because this is what we need to do count plus plus because the return value of delete is the number of keys that are deleted. Right? So we just added over the keys trigger a delete. If the delete was successful, we do delete plus plus. Otherwise we don't increment it there. And at the end we just encode and return. So return value is an integer which states the number of keys that are deleted.

Right? And now that we have this, if I check the encode function because it's an integer, just one change I've made over here is we earlier just added int64. Now I added all flavors of integer because count deleted was a simple integer type. That's why I've added all flavors of integer over here encoded in the exact same way, ⁓ colon %d slash r slash ⁓ n.

Right? And this is how you would be implementing delete operation. Iterating over all the arguments, deleting it from the hash map, if deleted, count plus plus and at the end return the integer count. Right? Now let's see how expire works. Expire function sets the expiration. First of all, we do a basic check on the arguments because we require exactly two arguments over here. So if the length of arguments is less than equal to one, it's an error condition. Right? Second is the first argument becomes the key.

And the second argument you have to convert, you have to consider it as an integers of 64, which would be the duration in seconds of expiry. Right? We set it over here. ⁓ In case it is not an integer, error would be set. So we would can type error value is not an integer out of range. This is the exact redis error that it throws. Then what we are doing is we are getting the key because to set the expiration, you have to get the object from the hash map. We got the object over here.

Sneha Mehra (00:06:33)  
in case the key does not exist, in case that object doesn't exist, what do we do? We return zero. Because ⁓ if with expiration, if key, we saw that if key didn't exist, we returned zero, right? Because the operation was unsuccessful. We were not able to set. So what Reddy's document says that zero if timeout was not set, example key doesn't exist or operation skip due to provided arguments. Any reason if timeout was unsuccessful, setting timeout was unsuccessful, we return zero. Otherwise,

we set the expiration. So expiration is set to what? We store expires at the absolute time at which the key would be expired. So it would be time.now.unix millis plus expiration duration in seconds into 1000 because we store it in milliseconds. Right? And then we just return one as an integer response. A classic implementation of delete and expire. Right? So this is how delete and expires are implemented.

So now let's look into how auto deletion happens. We saw how delete is implemented. We saw how expire is implemented. Now let's talk about how auto deletion would happen because now here up until this implementation, we deleted something from the hash table only when the delete command was issued, right? While getting it, we were checking if it is expired, we were returning not found, right? But we did not explicitly delete it. So let's see how deletion, how Redis does.

key auto deletion because if you have set an expiry ⁓ after that expiration time is reached I would want to actually delete the key from the hash table. How do we do that? That is where let's quickly take a look at how Redis expires the key. ⁓ So in Redis we can set expiration to the keys by typing expire k 10 expire key and the seconds right. Now here ⁓ why do we have to set expiration so that there is no burden for manual deletion. You can just set and forget.

You just set the expiration and forget your Redis would do the cleanup. We have to now write that cleanup job. So how does that happen? So the first intuition that you might think, hey, it could be another thread that is running. But remember Redis is single threaded. That is where this is such a beautiful implementation. So Redis expires or Redis auto deletes the key in two modes. First is the passive mode. So, and the second one is the active mode. So first let's take a look at the passive mode. So what passive mode says is that

Sneha Mehra (00:08:57)  
a key is passively expired when some client tries to access it and the key is found to be expired. Which means that if you getting an object from the hash table and if that object has some expiration set, right? And if that key is found to be expired, you would be immediately deleting it. Right? So you are accessing an object from the hash table, you got that object if the expiration time is already passed.

then you would be triggering an explicit deletion and then returning null to your user that key does not exist. That is a passive word which means that ⁓ if your expired key is ever accessed, if an expired object is ever accessed you are explicitly deleting it. Problem solved. But ⁓ what about the keys that are never accessed which means a key ⁓ expiration is set but that key is never get for some reason. Then key would be lying there only.

So that is where you have another mode which is an active mode. So what active mode does? This is a beautiful probabilistic or rather beautiful statistical algorithm that it does. Now, ⁓ bit of math, ⁓ slight bit of math, but very interesting algorithm. So here ⁓ what it does is given that the passive mode of deletion would take care of most the keys, like most of the keys that needs to be deleted would be taken care from the passive mode, right?

because a key is very likely that you put it in the cache is very likely to be accessed and that would be deleted. So now the keys that are in question are the ones that are never accessed. That would not be much, right? That would not be many keys like this. So that is where what active mode does is active mode works with a very simple ⁓ with a very simple statistical theory that if you take a random sample from a population that random sample ⁓ would represent

the distribution of the population. Right? So here, if you randomly pick a sample of 20, ⁓ If you pick a, if you randomly pick a sample of 20, the number of expired keys or the percentage of expired keys in this sample of 20 would be very close to the percentage of expired keys in the actual population. So that is what it does is Redis runs a loop.

Sneha Mehra (00:11:14)  
A simple loop that executes 10 times every second which means every 100 milliseconds it executes a loop. It's an interrupt that happens. We will see how it is implemented. ⁓ 10 times every second it runs. What it does is it tests random keys having some expiration set. So out of all the keys that you have added in the redis, it would pick 20 keys at random whose expiration is set.

out of those 20 keys that it has picked at random, it would see how many keys of them are expired. It would first of all delete them. ⁓ If these deleted keys or if these expired keys are more than 25 percent, then it would repeat the process. Right? This is what would, this is what is going to happen. So this would be a continuous process that is running whose job is to pick 20 key or pick 20 keys at random.

out of these 20 keys, find out how many of them are expired, first of all delete them, see if these are more than 25 percent, if it is, which means that if your sample contains more than 25 percent of expired keys not deleted, your population would also contain more than 25 percent of keys that are expired but not deleted. This is what it plays with, right? And this is such a beautiful algorithm, why? Because this is not the only way to delete it, you are anyway having a passive word.

which would be doing ⁓ the bulk of the deletion. The active mode is only taking care of the keys that are never accessed, that are expired but never accessed. So, that would not be huge. ⁓ For a normal generic use case, this would not be huge. So, this is where ⁓ it plays very beautifully with statistics. Where it says that if my

Sample has 25 % if my sample is more than 25 % of expired keys that are not deleted my population will also have more than 25 % of expired keys which are not deleted. So I have to repeat this process. Now, why is it doing so? First of all, it does not have to iterate through all the keys because that is extremely costly. ⁓ Redis is trying to minimize the overhead burden. You might think hey, ⁓ let me have another data structure in which I'm keeping all the expired keys to make it faster or something.

Sneha Mehra (00:13:34)  
Why to add that extra memory burden? This is already an in-memory database. It needs memory to store keys and values, not some additional data structures. ⁓ The second reasoning you can give is, hey, if I'm not creating another data structure, then why can't I just iterate through all the keys? That is costly. Imagine you have million keys having expiration set of one year or something. Why are you wasting time doing that? So that is where this quick sampling, see, in most cases, the passive mode will do the job. ⁓ For edge cases, where you

would have a chance of memory leak. You're just playing by it. So no additional data structures, no additional space requirements, just a random sampling, testing it. And here it does this exactly 20, exactly 20 keys. And how does it comes to this number 20? It's normal central limit theorem, right? Detail like proof of it is a little out of the scope, but if you Google it, you'll find an amazing resource on it, right? ⁓ Okay. So.

The 20 keys it samples out of the 20 sees how many are expired but not deleted. It deletes them and repeats the process if it is more than 25 percent as simple as this. Right. It does not have to go through ⁓ all this part. Now where do we put this? A normal intuition would say hey let this be a separate thread. Redis is single threaded. You cannot have a separate thread. So then how would you do that? That is where it's all about structuring the code.

Redis does it in a different way, but to make the explanation simpler, we would put it in a place that ⁓ makes it very sensible, very easy to understand. And then I'll tell you how to make it complicated. ⁓ Okay. So now let's see how this fairs. So again, we'll do an extremely deep code walkthrough on CN how ⁓ would you write an auto deletion. So given that this cannot be a separate thread, where do we put it?

the only place where we can put it in our event loop. Right? The event loop that we wrote that is where we have to put it. The infinite for loop that is testing that is waiting on epol wait just before that we would add it. But what's the frequency? Although Redis runs it 10 times every second will not run it very frequently. Let's say we run it at a frequency of at max or at a frequency of once every second. Right? So I'm setting up a chrono frequency. This could be a configuration if you want to.

Sneha Mehra (00:15:52)  
a cron frequency of 1 second. So, time dot duration of time 1 dot second and we would maintain last time when it ran. Right? So, let's say last time it ran was now when my server is starting. This is exactly the time when it was executed first. Right? So, what we can do is end the async server that we have written. The gigantic for loop that we wrote which is where the first thing we would do is execute this cron. Where the thing is we would check if time dot now

is after your last cron execution time add cron frequency which means that now that my time so when this loop would execute when my control flow would come over here it would check that hey is this like is my current time more than last execution plus cron frequency which means if the time if the enough time is passed then i'm running this delete expired keys as a simple cron and then i'm setting my last cron execution time equal to time dot now

Very simple and then our normal flow would work which is where we are doing epoll wait and reading the events and accepting the connections and doing that part right that would work just fine. We are just plugging our logic at the first thing in my infinite for loop that we are running right. So if enough time has passed delete the expired keys and this does not guarantee that it would run exactly once every second. Every time the control flow would come over here it would run it right.

a better way. Now here you can see that you cannot because it is single threaded. It's not separate thread that would run exactly one ⁓ second after every second. It is single threaded until and unless my code execution flow does not come over here or my control flow does not come over here. This would not execute which is fine. It is perfectly fine. Right. There is nothing wrong in it. Right. How Redis does it? It uses interrupts. Right. And then

you can configure a kernel level interrupt and then it would run it not making overly complicated this would work just fine because we are just understanding how single threaded high performance database systems could be built. Right? So now let's see what my delete expired keys function look like everything else is pretty clear when I'm doing time.now if my time.now is after last cron time execution plus cron frequency then I'm triggering this. Right? So now what it would do it is another infinite for loop

Sneha Mehra (00:18:09)  
in which although I should have had it bounded but that's fine we are just writing a dummy implementation of it what we are doing is we are writing an infinite follow up whose job is to do an expire sample right and it is just doing expire sample and getting the fraction if it is fraction is less than 0.25 then break so which means if I have deleted more than 25 percent of the keys which means in my sample if there were more than 25 percent of the keys that were expired but not deleted

then I would have to continue otherwise I am breaking the loop. ⁓ Right? And I am just printing in the log deleted the ex... deleted expired but undeleted keys total number of keys which are present right now after the deletion ⁓ is stored is the length of store which is the length of the I am just printing it for our purpose that ha indeed the deletion has happened. Right? Okay. So now that delete expired cases set over here now what expire sample does? Expire sample does

this part. what we do is again writing a sampling code versus Golang's map iteration is pretty random. The order in which a hash map or ⁓ a map of Golang is iterated is random depends on the hashing function that it is used. So that is why let's assume that ⁓ iteration of a hash is not in any order but in a random order. So what we are doing is because we have a limit of 20 we would be and we can sample 20 keys

that are expired, that have some expiry set. So the flow is I'm iterating through the hash map, right? I'm iterating through the hash map. If object.expiredAt is not equal to minus 1, which means some expiry is set. I'm doing limit minus minus. And if expiredAt is less than equal to time.now.unixmillisecond, which means the key is expired, already expired, I'm triggering a delete. And I'm doing expired count plus plus.

And if I have found 20 such keys ⁓ whose expired is set, this limit would become zero. I'm breaking the loop here. Right? So which means that this loop would run ⁓ up until I find 20 such keys whose expiration is set. Right? And then because expired count would hold the number of keys that are actually deleted, then I can just take a fraction of it. Expired count divided by 20\.

Sneha Mehra (00:20:32)  
⁓ So, if this fraction is less than 25 percent which means less than 0.25 more than 0.25 we would be taking up this call. ⁓ This is the active mode of deletion. Now, you see how it would work. ⁓ In the single threaded that are we in the event loop the first thing that we are doing is seeing if my last execution of this cron was before or rather it was done well before which means I have

I'm well past my cron frequency. I would trigger this in which I'm sampling 20 and see how many of them are expired. I'm deleting those expired and returning the fraction. If this fraction is less than 25, I'm breaking the loop. Otherwise the loop would continue a normal active deletion ⁓ flow would happen like this. Right? Now what about passive deletion? Now here you see the beauty of it. That's why we have abstracted the implementation. So here if you check the store.go file, what we know that

every time a file, every time a key is accessed, we check for the expiration. If it is already expired, we are deleting it or we have to delete it. Right? So that is where in the get function, ⁓ which is one of the reasons why I didn't expose my hash map literally in my eval file.

so that we can have, can mask this thing over here. So in the get function, we are getting the value for a particular key. If value is not equal to nil, which means value exists, I am checking the expiration. If expiration is less than time.now.unix.mil is, then I deleting the key and returning nil, which is what would be written over here if the key does not exist. If the value exists and it is not expired, I am just returning v, which means an unexpired key would delete as is. If the key is mid-expiration, we are act, we are

passively deleting it. ⁓ So this is like a lazy deletion. ⁓ So with this lazy deletion, if a key is accessed and found to be expired, it's deleted. ⁓ If key ⁓ is not expired, it is moved forward. And then periodically we are running this job, which is an active deletion cycle in a single threaded event loop whose job is to periodically randomly sample 20 keys and see how many of are expired but not deleted.

Sneha Mehra (00:22:40)  
you delete them and once you have deleted it you repeat the loop until that fraction comes down below 25 percent. This is how Redis does expiration of keys and this is how you would be implementing or you would be re-implementing it in GoLang. So now let us look at a very quick demonstration of this. I am running our server on port 7379 and now what do we do is I am just connecting the Redis client on post 7379\. Here you can see

the first output is deleted expired but undeleted keys total keys equal to zero right now I have nothing there are no keys to be deleted the output of this is total number of keys are zero right now let me do a set of key k with value v no expiration set it does okay if I check TTL of k I get minus one right because no expiration is set and now what I am doing is I setting expire of key k to value 10 which means in 10 seconds this key would be expired if I check this

Here you can see the timer starting and on ⁓ the top you can see total number of keys is 1, keys is 1, keys is 1, keys is 0\. ⁓ So, we did not delete it. The loop automatically deleted the particular key and here you can see how by total number of keys are 0\. ⁓ Because the loop that we are running, its job is to ⁓ actively delete the keys. ⁓

Why? If the key is getting accessed it would be automatically deleted if it is expired. So, let us take that as an example as well. ⁓ So, let us say I set k comma v do expiry 10 and I am doing normal get. Here you would see as soon as I am doing get get get get and if the key would have expired and I would have done a get on it, it would start returning me nil. So, without me accessing it because I am doing a get of it.

which means I'm accessing an expired key. I'm accessing a key which is found to be expired. It is automatically deleted, lazy deletion. So passive deletion in action, loop in action. And this is how you would be implementing auto deletion, right? And now that we are on the topping, let's quickly see how we can set expiration of it with all the edge cases set K comma V. If I forgot to set expiration, I can just do this. I can type in. ⁓

Sneha Mehra (00:24:57)  
Let me first do a deletion of key k1, k2, k3, k4. I should see 0 because key k1, k2, k3, k4 doesn't exist but key k exists. If I do this, it returns 0\. ⁓ But as soon as I am adding k to it, integer is 1 because 1 key was deleted. ⁓ Delete functionality working. If I do this again, output is 0\. ⁓ Similarly, with expire, if I passing a key that does not exist,

it would return me zero because operation is not completed. There was no key to set. Right? But if I do set k comma v ⁓ and then I set expire 10 and output is 1\. Right? A classic implementation of delete ⁓ and ⁓ expire. Right? ⁓ Great\! We implemented delete, expiration, saw how Redis does auto deletion, active mode, passive mode, we implemented everything. Right? Okay. So what's next?

Now here you can see a very ⁓ nice thing which is pending. The nice thing which is pending is although we deleted it, right? We expired, we deleted the keys explicitly by deleting. We did auto-deletion for the keys which are expired. But what if because your Redis stores everything in RAM, what would happen if I'm just inserting keys bluntly, right? Someday your RAM would get filled.

What happens when there is no more space to allocate? Like you are running a malloc, but that has no space to allocate something. What would happen? In that case, your program crashes, but you don't want your server to crash, right? Because your end user request will get failed. So what do you do? A classic cache thing is you do eviction, which means that when your cache is full, you do eviction. So in the next video, we would look at how Redis does eviction.

we would implement a very simple eviction strategy which works on the number of keys. ⁓ We will not make it overly complicated. We will just implement it that hey, Atomax my Redis server or our Redis implementation can have let's say five keys just as an example, right? And we would see as soon as we are trying to add the sixth key, it would evict one and put our

Sneha Mehra (00:27:11)  
the new key that we are trying to insert in it. A classic eviction strategy. ⁓ This is what we would be implementing in the next one. I hope you found it amazing. That is it for this one. I'll see you in the next one. Thanks a ton.

—-------------------------------------

23

Sneha Mehra (00:00:00)  
So Redis supports sets. And to add an entry into the set, you just file the command s add ⁓ for a particular key. Let's say my key is k1 and I'm just adding a member. Let's say I added a string A into that. Right? And if I'd want to see how it is implemented, the encoding of it, I can just do like always, debug object, and I would do K1. Here you see the encoding that it is using is hash table, which means that sets are implemented through.

Hash tables in Redis. But but but there is a very interesting catch. Here we try to insert a string. What if I create another key and I pass in an integer? So I'm creating a set having integer values. So if I do GitHub debug object on k2, I created a set k2 in which I added integer 1, integer 2, integer 3\. ⁓ If I do this, here I get another encoding which is int set. So

In Redis, sets are implemented in two ways: a normal set and an and an integer set. A normal set is implemented through hash table, right? A classic hash table implementation. ⁓ And another thing ⁓ is around like if your set is ⁓ a specific integer specific set which only holds integer elements, then it stores the encoding it uses is an int set. Right. And now as soon as just to give you this, if I add

⁓ string into an int, it accepts that. And if you now check the encoding of it, it becomes hash table. So which means that it does change the encoding depending on the kind of value that we have. Right. So as soon as we know that all of our values are integer, the encoding would be an int set. As soon as I added a non-integer value into that, it elevated itself to become to be implemented via hash table. So all of the reallocation and shuffling would happen by the way. Okay. So let's take a look.

And how it is actually implemented. So here I have a file called T underscore set. ⁓ The way you can trace it is again through the same. If you check for S add command, you will land up to this file. ⁓ So if I scroll through this for the first few lines, here only you see what ⁓ how where it is creating an int set and a normal set. The first command itself is the set type create, which is creation like

Sneha Mehra (00:02:24)  
To create a set of a particular type, it just checks if SDS. SDS is basically simple dynamic string. We'll touch upon that in a couple of videos later. It imagines it's just a string. So is SDS representable as long long? Which means if my string ⁓ is representable ⁓ as a long long. So string 2 long long if it is possible, then okay, otherwise error and then it would set the value there. If it is representable as a long long, which means that my set

That I'm trying to create right now with the first value that I am trying to create. Let's say if I'm doing a set add, I'm doing an S add on a particular key and value is 1\. This would be true. Because this would be true, I am creating an int set object. If this would be false, if I'm inserting the first element as string, it would be creating a set object. The normal set object will hash table implementation, while int set would be an integer set, which is much more interesting. Hash table is pretty ⁓ like you'll find a ton of resources on a hash table like.

Link a couple of like I have a very detailed free playlist available on YouTube on hash table internals. A lot of that applies as is in this implementation of hash table with set. I'm just linking them in the description down below. Watch that if you are interested in knowing hash table internals, resize, resize, and everything I've covered there. But what we are more interested in is the int implementation. So intet here, if I check the int, here it is creating this object of an encoding type is set to int. ⁓

And then here it is creating this inset new, which is allocating size insert like classic implementation. But the fun part begins like how it is actually storing it. So let me walk you through on what it actually does, and then we'll come back to the code to understand how it is actually doing, right? We'll actually go through the source code to understand the nitty-gritties of it. Okay. So just to start from scratch, Redis supports set as a data structure. It is implemented using hash table.

But if the set only contains integer values, it is implemented using int set, which is much more optimized for integer specific values. So, how does int store the value, which is so optimized? ⁓ Int sets are storing the values as a sorted list of integers. That is it. Because if you think about it, sets does not contain any duplicate. And if you just keep the list sorted, ⁓ that just could be your set implementation.

Sneha Mehra (00:04:43)  
Right. So what insets are? Insets are just a sorted list of integers, but only it is valid only till certain limit. ⁓ after that it moves to hash table because a long list of integers is an optimal. Right. So that is where you have a configuration called set max inset entry, which says that hey, at max I'll have 50 entries, 100 entries, 200 entries in the inset. After that, I would not be, I would be moving to hash table. Right. This is a configuration that Redis provides. So after this configure limit is set, ⁓ Redis starts using hash table to hold the data.

Now, if I start with just three elements one, two, three, I'll get ⁓ my implementation using inset. And insert and data would be or now obviously if you think about complexity, because it's a sorted list of it, it's not a linked list, it's a sorted list of integers, a literal array of integers, if you may want to put it. So insertion and deletion would take order in because worst case would be what? When you're inserting, because it's a sorted list, you may get in elements in any order.

You are inserting in the middle, so then you have to move the elements and then insert something in the middle. Or if you are removing something, you would have to move back elements once again. So that is that. But search that's the most common operation for set. Is does this exist in the set? You're doing set union set intersection. You most common operation. Because it is a sorted list of integers, ⁓ it's just a binary search. So your search operations on a set becomes order login. Right? And this is why I'm talking about int sets. Right? It just order login. Well, look at this.

Code ⁓ right now, how it is actually doing. But like let me let me do that. Let me do that later. Let me first walk you through how it is doing it. We'll definitely go through the source code and see how it does it in action because you'll understand the layout of it. So, like we talked about in the previous video, we talked about zip list and its layout. ⁓ Excuse me, because of like similar to that, you would have layouts for ⁓ inset as well. So, an inset is defined something like this: the layout is the first thing is encoding, then length.

Then int 1, int 2, int 3, int 4, and so on so forth. Integers are fixed with integers. It does not mean it is 4 bytes, by the way, because you may have a 16-bit integer or 32-bit integer or a 64-bit integer. So, what it does, what it is instead of allocating a 4-byte integer without utilizing 4 bytes is a waste of space. So that is where encoding says that what kind of integer are we storing? Are we storing 16-bit, 32-bit, 64-bit? Depending on which it would allocate the space or it would be

Sneha Mehra (00:07:08)  
Placing the elements according to that. ⁓ So because each integer is fixed with you can very well know if I want to access a third integer and if I know from my encoding that my each integer is 16 bit long, I can directly go to that space and read the fourth integer if I want to. ⁓ Similarly, if let's say today I am only storing 1, 2, and 3\. But let's say tomorrow I store 4 billion in that same array. So what would happen is initially it would have allocated to int 16, right? Because that's the smallest that it can allocate.

Would allocate to N16, but as soon as we are trying to insert a big one, it would elevate and reallocate a different thing, a different inset ⁓ of encoding N64. Right? So that is where ⁓ by like in most cases, the normal user ex the normal user behavior with Redis would be to put homogeneous types, they would not vary pretty heavily. ⁓ So typically the first few values would determine what it is, but Redis is flexible enough that in case it sees

That this value is unfit for this particular encoding, it would elevate it, reallocate, reload all the data, and then put the new entry into that. So that it would be an expensive operation, but worth doing it because in most cases you would be saving a bunch of space. ⁓ So this is exactly how you would be storing ⁓ an int set in memory. Now, here you can clearly see that this is very similar to an array structure. ⁓ But instead of implementing it with array, it's just a block of memory in which you are physically.

Knowing where to go to the memory offset is you are managing it. So with encoding, you that length you know how many elements are there. So you can directly jump to any element that you would want to. Right. Now adding something to an inset is that you would first check if the value exists or not. Right. ⁓ So check if the value, sorry, first thing, check if the value can be added to the insert or not. Which means that if you are inserting 32-bit in a 16-bit insert, that would not happen. So you have to do reallocation and all. If not, upgrade the encoding, reallocate and whatnot, you'll do that.

Do binary search and see if int exists. If yes, do nothing. Because if element already exists, you do nothing. Because it is sorted, you can do binary search on that. Then with binary search, you will get output as a position of insertion. So then if you got a particular position as insertion, that this is where I would want to put in, you would mem copy, literally fire that sys the you literally fire that called mem copy to copy the elements from here to one place like mem copy mem equal mem copy plus four.

Sneha Mehra (00:09:35)  
It would literally copy the elements like this, and you would be inserting your element or you would be writing your element at that specific location. Right. So this is what you would do with memcopy or mem move. Pick your favorite, like memcopy ⁓ copies, memmove moves, mem move would be suiting better. So it uses mem move internally. Right. And then set the integer value the position that you are interested in. Right. Similarly, for delete, you would find a position to delete memov and shift elements to the left. Right. And because inset is just a sorted array.

Set operations are hyper optimized because for integers, if you don't set union, it's just adding through sorted list of integers and creating a new out of it. Highly optimized for ⁓ this thing. for intersections, also very simple because it is sorted. You can literally it's ordered operation, you can literally go through them and see which one are the same and then move the point of like simple to merge of merge sort kind of logic that you can apply over there. Right? Okay, so as said, let me walk you through.

The actual source code of that and see this in action, like how simple yet beautiful this implementation is. So let me jump to the most interesting part. So here inset new we allocated. Inset resize is where we know that hey, this is what we would want to resize. Like the size is small, I want to enlarge it. That is for that. This is where the search part is. In set search, where I want to search something, some value, and hold the position in which we are searching. Like binary search output would be what? Would be the position at which the element either exists or

Or if it does not exist in the position at which it would be inserted. So that's why here we are passing the pointer so that we don't have to do two lookups there. And then this is where we are doing the binary search. A very familiar code. While max is greater than or equal to min, you compute the mid, then you get the element. If value is greater than current, then you do this, otherwise do this and continue the loop. This is where you can see your binary search that we studied so much in depth in colleges, in our data structures, algorithms, practices and whatnot.

Here you see that in action that how Redis user said to optimize int sets, ⁓ right? Okay. And this is the code base, or rather, this is the file where you will find a lot of other operations that they just to walk you through one. This is how inset add happens. This is where you can see upgrade and add where you have to upgrade the encoding and add. You are doing search if success, then something, if not, then do resize and you add. Otherwise, inset.set at this position, this value, and you are updating the length of the inset. ⁓

Sneha Mehra (00:12:00)  
A very standard implementation, but the nitty-gritties of it will be coming when you re really implement it. You'd re you'd understand the challenges that would come with that. ⁓ But obviously, just a couple of points to highlight. You are not implementing, you not allocate, you will never be allocating the memory only that much that you need. You would be allocating with some buffer. Once you hit that limit, you would be reallocating a big chunk of memory and then re refill the data into that and then add the incoming entries to that.

Right, so yeah, these are the couple of points that you have to keep in mind. But this is how int sets are implemented in Redis. These are specialized sets that hold only integers. Generic implementation is with is with hash table. You can find a very extremely detailed playlist on hash table internals by me on YouTube. ⁓ and I'll link them in the description down below as well. Right, and like always, ⁓ I have not implemented this particular part. This is for you to understand how Redis actually does it, how it is very frugal.

I have opened up a GitHub issue. I'll link them in the description down below. If you find it still open, no one is working on that. I would highly encourage you to pick that thing up, ⁓ use it to implement this on your own. You will learn so much about encodings, being frugal, allocating memory and whatnot through that. ⁓ I would highly, highly, highly, highly encourage you to do that. So, yeah, that is it. That is it for ⁓ like basically that is it for this one. I'll see you in the next one. Thanks at all.

—-----------------------------

4

Sneha Mehra (00:00:00)  
So at the end of the previous video, we saw that when the Redis CLI connected with our ECO TCP server, some messages were exchanged, where we got to see in the ECO server, something like star $2.07 was received. What exactly was that? The message that we received from the CLI was the first message that the CLI sends to the server. ⁓ It sends to the Redis server, but because we connected the CLI to our server, we saw the message that it sent. ⁓ The message was sent in a protocol.

And this protocol is nothing but a way through which we would be sending the data and the server would be understanding it. So here, the Redis server understands Redis serialization protocol, which dictates how different data types, how different information would be exchanged between the client and the server. ⁓ Redis.

or RESP supports common data types like integer, strings and arrays. You don't need anything then because Redis is an extremely simple database. So it does not have very fancy features. Okay, now I'll give a simple example. Let's say we want to put a key in the Redis. What do we type? We type put space K, space V. And we type this exact same thing ⁓ and your Redis server understands that we have to put the key K, ⁓

and ⁓ corresponding to that the value V. Okay, so when we would want to do this, so what Redis CLI, you type this in Redis CLI, what it sends to the server, it basically sends to the server an array of strings that contains put, k and v, such that this information right now on the screen it looks like a JSON or a list of string, ⁓ But when you're sending it over the network, what

way are you serializing it to send it over the network depends. With RESP, you encode it in a way that your Redis server is extremely efficient to understand and process. ⁓ Now, what does this specification look like? A general rule whenever you are reading this, a general rule to understand is that,

Sneha Mehra (00:02:07)  
The RESP is used as a request response protocol, which means that the request that you would send to Redis will be RESP encoded and the response you'll get will also be RESP encoded. Every data type, like because it is supporting multiple data types like int, string, array, every data type starts with a special character and the data ends with CRLM, which is basically your carriage ⁓ return.

line feed slash basically slash r slash n is through which your data ends. This is a general rule. Now we'll go into each of the types that Redis supports and see how it is encoded and transmitted. So we'll start with the first one, ⁓ simple string. We all would have tried that when we connect our CLI to the Redis server with the first command we send is ping and in respond we get pong. ⁓

So here the first data type we'll see is called simple string. So simple string starts with a plus because it has to start with a special character. Starting with plus implies that it is a simple string. It is followed by the string that we would want to send. For example, if you want to send pong, you would write plus P O N G followed by C R L F, which is slash R slash N. So when you send this data ⁓ over the network,

to your Redis server or rather let me take another example. If we send ping instead of pong, if we send plus PING slash R slash N to the Redis server, the Redis server would understand, hey, ⁓ I received because the message is starting from plus, hey, I received a simple string ⁓ and it will read until it gets a slash R slash N. It read till that and while reading it, it got PING. So then it will interpret that, hey, now I received a simple string from my client.

and it says ping. Now it would do whatever the business logic is written. It would return pong in the similarly encoded fashion. These simple strings are extremely efficient.

Sneha Mehra (00:04:05)  
because it has very ⁓ low memory overhead because if you'd want to send a four character ping slash pong, something like that, the extra overhead you get, you are reading a plus and a slash R and a slash N. So just N plus three bytes that you require to do that. So that's why for simple requests and simple responses, simple strings are used and that's why the name simple strings. Okay, the second data type that we see are integers. Now integers in Redis starts with a colon.

⁓ So for example, if I want to send 1729, ⁓ RESP encoded to the Redis server, would do my integer starts with a colon followed by the integer and then your CRLF. ⁓ So if I want to send 1729, it would look colon 1729 slash R slash N. This way, when this goes to Redis, it says, hey, it starts with a colon.

If it starts with a column, which means it has to be an integer. It will be a 64-bit integer. So whatever I get, ⁓ whatever digits I get up until slash r, I can interpret safely as an integer. This is RESP encoding for integers. The third type is very important. It's called bulk strings.

⁓ So bulk strings are strings which are binary safe. We'll just in a couple of minutes we'll take a look at what bulk string is and why it is binary safe. But let's understand the specification for bulk string. Bulk string starts with a dollar followed by the number of bytes that we would send. For example, if I'm sending pong, so then my string starts with a dollar.

then four because pong contains four bytes, P-O-N-G. So dollar four followed by C-R-L-F, which means dollar four slash R slash N. Then the actual string, so P-O-N-G that I'll send and then followed by C-R-L-F slash R slash N. This is the specification. Now this you'll say, we just send simple string, it was just N plus three because just so only slight overhead. Here you are adding so much of stuff, why?

Sneha Mehra (00:06:13)  
because the key highlight of ⁓ bulk string is it is binary safe. So with simple string, you could not have sent a slash R character within the string itself. Because let's say if your string itself contains the byte slash R, then while parsing the simple string, you would be stopping at that point unnecessarily. Although you are yet to read a lot of other data. That's why your strings are not binary safe.

So here what we are doing is because of this bulk string, are doing, we are specifying the length of the string in bytes at the first part itself. We exactly know how many bytes to read, irrespective if it contains slash r slash n, no matter what, it would read till that point and then you'll get slash r slash n as the end of the string. This is why it is binary safe. ⁓ Right? That's an extremely important point because now what would happen is this means that

we can store and send any binary information, any binary. For example, you can even store a PNG image in Redis.

Although you should not, but if you want to you can because an image data can contain any type of bytes. It can also contain null character. In normal languages, you might see null as an end of string or as an end of input, or you might see slash R as an end of command as we just saw. But because you have to be binary safe, you have to prefix it with the length that you gonna read so that you read those many bytes and basically process it however you want.

This is why binary strings are important. This is why bulk strings has its own significance. Okay. Some specific example. So let's say if I want to represent empty string. An empty string is what? A string with zero length. So I can write $0, slash r slash n, slash r slash n. Classic. Because what was our protocol for bulk string was $, followed by the length, which is 0, followed by slash r slash n.

Sneha Mehra (00:08:12)  
followed by the data because it's an empty string, is no data at all, ending with slash r slash n. So this is a classic representation of an empty string. Now, if I want to represent a null value, a null string, right? So lack of data, if I want to represent that, it could be dollar, then minus one, slash r slash n. So minus one is a special value that indicates lack of data or null.

So it would interpret, so your client and your server can interpret if you send $-1 slash r slash and it would interpret it as a null value. ⁓ Okay, time to look at next data type. The next data type is arrays. So here we just saw that when your Redis CLI connects to the server and we would want to do a put k v, ⁓ you are passing put space k space v and we are writing that.

but the Redis CLI would be converting it, would be RESP encoding it and sending it to the server. So any command that you fire ⁓ through any of your client is RESP encoded and sent to the server. And it is encoded as array of strings. So that's why arrays are a very important data type in RESP. ⁓ So an array starts with star followed by the number of elements in the array, then followed by a CRLF.

and then followed by the resp encoded elements. I'll give a simple example. Let's say if I have an array that contains ⁓ a string a, an integer 200 and a string cat. So how will I encode this? I'll start because this is an array of some elements, it's integer and string all combined. ⁓ So I'll start with a star three because I have three elements, star three slash r slash n. Then the first choice

⁓ string with one character a. So, I will represent it as a bulk string. So, I will write $1 slash r slash n a slash r slash n. Then second is 200 integer. Integer starts with a colon colon 200 slash r slash n.

Sneha Mehra (00:10:15)  
And then cat, so $3, slash R slash N, C-A-T, slash R slash N. I've represented it one in each line. There is no extra new line that is put over here. It's all it's one bit string. ⁓ So this is the beauty of how you can represent array with different data types ⁓ as R-E-S-P encoded strings. This is the R-E-S-P encoded string that you would send over the network and your server or your client would interpret it ⁓ depending on what the data that you have sent.

⁓ The null arrays, if you want to send null arrays similar to how null strings are represented, you can type star minus 1 slash r slash n, minus 1 is a special value that denotes null. And empty arrays again, ⁓ star 0 slash r slash n, because you don't have any elements, so you are terminating it with slash r slash n. So, the protocol ⁓ is very well followed. ⁓ And this is how arrays are represented. Now, here you can very clearly see because an array

can also contain an array. So an array within which an element itself is an RESP encoded array can also happen. So this is how you can create complex representation of data in Redis using RESP encoding. You can send and receive the data in RESP encoded forms.

⁓ So these are the three or rather four key data types, simple string, string, integer and array. ⁓ But what you also know that you may sometime fire a command that does not exist. So you get an error. So we need a way to represent errors. So error messages starts with a minus sign.

minus sign followed by the message followed by slash r slash n. ⁓ the starting with minus is important because anything that starts with a minus implies error. So now if you are writing something on your own, you can send this to the server and server would interpret it as an error or the client would interpret it as an error. So if I'd want to encode key not found, what I'll do? I'll start with, it starts with minus, minus K E Y space N O T space F O U N D slash r slash n.

Sneha Mehra (00:12:17)  
This is how you'll represent an error message. ⁓ Okay, now some key highlights, although we discussed four types and we also discussed how errors are represented. Now the key highlights of RASP, like why did they come up with their own protocol? Can't they have used JSON? JSONs are very bulky. You have to parse until you get a...

quote and then read until you get another quote while escaping the quotes in case you'd want to do that then a colon and then something. It's extremely complicated and inefficient. Redis wanted to keep it extremely simple. They wanted it to be human readable. This is indeed extremely human readable.

They wanted it to be simple because if it is simple, then it would have fewer bugs because you can't have bugs with your something as simple as your serialization protocol. They wanted to keep it extremely simple. So they went with this format. They wanted it to be performant as I just explained while converting JSON to ⁓ like ⁓ JSON to map or something or back. It's extremely complex. Complex as in it is extremely time consuming. It's like array, array, ⁓ we have JSON, we can just do stringify and the reverse way.

But it really someone is doing the parsing for you and that consumes CPU. Redis wanted it to be extremely lightweight and which is where we see how RESP is extremely performant. It does not have a lot of mumbo jumbo. It keeps it extremely simple. It is performant, it is efficient and it is extremely small. So no extra network overhead, the bare minimum data that it has to send it sends and just with slight extra bits here and there.

⁓ And the best part of this is RASP is prefixed length, which means that in most types that we see, bulk strings, arrays, ⁓ the first thing that we know is the number of elements or the size that we are reading. This way, we exactly know how much to read. And this makes life so simple because over the wire, the data that you are receiving is pure stream of beads.

Sneha Mehra (00:14:23)  
Given that you are just receiving stream of bits, you need to know how many bytes you are going to read because reading ⁓ is a blocking call. So that is why if you already know that, let's say if I want to read hello, if I'm rather, ⁓ let's say if I'm getting the string, it starts with dollar, which means it's a string. I read till slash r, I get five, five is the length.

Then I'm making a call to read five bytes. I'll write H-E-L-L-O. As soon as I got five bytes, then it would end with slash R slash and I would definitely know that. So it makes it extremely simple. Number one, to allocate the buffer that if I'm reading a string of length five, I'll only allocate a buffer of length five. So you can do memory optimization because of this. Because it is prefix length, you exactly know how much you'd want to read. So no extra reading, no extra blocking on the network and keeping it extremely simple.

Now this prefix length also holds true for array because the first thing that you do is it starts with a star and then the number of elements in the array. That's why having this prefix length, you'd see this in most databases out there that they're always prefix length because you exactly want to know how many bytes to read, how many things to expect. ⁓

And this is what makes RASP extra special. ⁓ So that is it for this one. In the next video, we will go through the source code or we will go through the code where we implement this RASP in GoLine and see a few things in action. ⁓ So thank you so much for watching and I'll see you the next one. Thank you.

—------------------------------------  
5

Sneha Mehra (00:00:00)  
So in the previous video, we saw what RESP specification is and how every single critical data type is represented, encoded and serialized in the RESP specification. In this video, we will do one very exhaustive walkthrough of the RESP specification in our own Golang based re-implementation of Redis. So for this, I've created a module called core within which I have created a file called resp.core.

resp.go will contain anything and everything around resp specification. Basically encoding and decoding of values will go over here. Right? So the function that I have exposed is called decode. Decode takes in a slice of byte called data and it returns the actual object. So if we get in resp.engodate form an integer, this would return an actual int object of golang and it would return an optional error. Right? In this we are invoking another function called decode1. So what happens is we might get a large

slice of byte an extremely large slice of byte but the decode one function will ⁓ decode the first ⁓ resp encoded value out of it so i might have three values encoded back to back but decode one function would just decode the first resp value which is what we are invoking it is like a helper function think of it as an helper function so decode one takes in the slice of byte it decodes the value out of it

and it returns two more things. One is error and one other is integer. We'll take a look at it in a couple of minutes. So let's see what decode one function does. It takes in a slice of byte ⁓ and it returns an interface int and an error. Interface is like generic object that we are returning. And this could be anything, int, string, array, map, anything. ⁓ And what we are doing is because we know that in Redis, every single ⁓ RSP encoded value or every single data type

starts with a special character, which is what we are trying to switch. Switch on the data of zero. So a simple string starts with a plus and error starts with a minus and integer starts with a colon. A bulk string starts with a dollar and array starts with a star. So that's what this switch is all about. Now let's start understanding it one by one. So if we go and read about simple string. Now simple string ⁓ is that whatever value is there in the slice of bite, we pass it over here. Once we get that,

Sneha Mehra (00:02:17)  
What it does, what this function does, it reads the RASP encoded simple string from data and returns the string, the delta and the error. Now we'll understand what this middle value is all about that I was talking about. So your slice is very large, right? Your slice can be very large. You might have three or four values back to back encoded in the message, but you want to just interpret the first value out of it. So while interpreting the first value, you would gather that value

Apart from that, you would want up until which location have you parsed it. So the second value is the Delta, which is the length ⁓ of the value that you have parsed and processed so that if someone wants to encode or decode the next value out of it can start from that location. This is the Delta. So this helps us move forward one after another. It comes in very handy when we are building complex data types, which we'll look at in a couple of minutes. So if I take a look at the simple string, what do we see?

So simple string in redis, what it does is, simple string in redis ⁓ is just, it starts with a plus, then you have the string and then you have slash r slash n, right? So plus okay slash r slash n implies okay as your normal string. So how do we implement this? We start with pos equal to one because first is the first character, the zeroth index is plus. So we start from the first index. This is what your pos, this is what your delta is. ⁓ And we continue to parse.

or sorry we continue to iterate through the array until we encounter a slash r. As soon as we do that, we mark this location, interpret it as a string and return it. So this is the string that we are returning, string of data of one colon pos. This is the string, the actual string, the okay that we just got. And the second value that we are returning is pos plus two. What is pos plus two? So for example, if your string is plus okay,

slash r slash n, the number of bytes required to encode this entire value is ⁓ 1, 2, 3, 4 and 5\. ⁓ So slash r slash n, that plus 2 is to accommodate for slash r slash n, which means that you have parsed till that location. So now if there were three strings back to back attached, I would do, I have to tell the other function that I've read till this point, right? Because slash r slash n and since then you can start.

Sneha Mehra (00:04:44)  
So the number of bytes that you have read is five plus OK slash R slash N. ⁓ The number of bytes that you have read is five, which is what this would handle. So string of data one plus plus plus two is what you're returning and an error is because there is no error. Right? Now let's take a look at the second type, which is error. So error is represented, starts with a minus and then followed by the error message followed by the slash R slash N. Right?

Similar to like if you look at this the difference between an error and a simple string is just of plus and minus nothing else. ⁓ Given this we can just invoke read simple string out of this and return whatever it returns. Dead simple. The third one is an integer 64 value. So an integer starts with a colon followed by the integer then slash r slash n. So integer is literally string representation of an integer. ⁓

So if I want to send 1000, I'll be sending colon 1000 slash r slash n. So if I'm sending that particular value, our iteration would be very similar. So what we'll do is read int takes in slice of bytes and returns an int 64, ⁓ delta and an error. So we'll start with the first, like pause equal to one because the first character is colon. And from there we start our iteration because now if I'm sending colon 1000, so I have to start from the first index.

because the first one is already done. So pos equal to one and I'm creating a variable called value in which I'll be reconstructing the integer because I'm getting integer in the string format or in the byte format and I have to convert it because it's a string to int conversion. ⁓ So while I'm iterating, ⁓ for pos equal to pos data of pos not equal to rl slash r. So until I discover slash r, I'm reconstructing this integer using value equal to value into 10 plus int 64 of data of pos minus zero, ⁓ right?

So minus zero because I'll be getting zero as a string or zero as a byte. I have to convert it into zero as value zero. So this minus byte zero, I'd be reconstructing this integer all over again in value. And I'm returning that int 64, pos plus two to accommodate that slash r slash n at the end and a nil because there is no error over here. Right? That's what your read int 64 would do. Then the next one is bulk string. ⁓ Bulk string is starts with a dollar.

Sneha Mehra (00:07:07)  
Bulk string starts with a dollar followed by the length, followed by the length of the string, followed by slash r slash n, followed by hello, the actual string and ending with slash r slash n, right? Now to write a decoder for this, the first thing that we have to write is we have to write like similar to all other places, we'll start our iteration from the first. So pos equal to one. And we are passing this into a read length function. So read length would

traverses through from the current offset, it would traverse through until it can gather 0 to 9 digits, until it gets slash r typically, until it gets a slash r it would continue to ⁓ iterate and reconstruct the integer and that would be the length and the delta as in how many bytes you really require, that same logic goes over here. So if I look at what read length does, it does that exact same thing. So I am iterating from that position extracting that byte.

If it is ⁓ out of integer range, means that byte, which means that while parsing through my byte array, I encountered a position which is not a digit, which is 0 to 9\. Then what I have to do is I have to directly return the length because the integer that is constructed, I'll be returning it and I'll be returning pos plus 2\. I read and pill this point n slash r slash n. So plus 2 is the adjustment for that. Right?

If it is not there, I'm just reconstructing my integer over again. Length equal to length into 10 plus int of b minus zero. Right? Really simple read length function that you have over here. Now, now that you have read the length, what's the next part? You have to adjust the pause. That the actual pause you need to adjust that you have read till this position. Now from that location, you have to read the length number of bytes, which is our actual string and then a slash r slash n.

So which is exactly what we are doing. We are returning the string representation of data of starting from the pause to pause plus length. I'm interpreting it as a string and returning pause plus length plus two and a nil. So pause to pause plus length because pause was pointing to the first character of the string. Length is the number of characters in the string. So from pause to pause plus length will be my actual string. So hello, it's H.

Sneha Mehra (00:09:28)  
to LELLO, hello, that is my string. And the number of bytes that I have parsed is till that plus slash r slash n. So pos plus len plus two is the number of bytes that you have processed. So that your further, your other functions can use that and jump after that, right? So this is how you will read bulk string. And then the final one, the big one is the array. So array ⁓ is, array typically starts with a star. Array starts with a star.

followed by integer, which represents the number of elements in the array, followed by a slash r slash ⁓ n, ⁓ then the actual elements back to back. This is where your delta, this is where your ⁓ decode one function comes in handy. Because here, what you are having is, you are having values which are stuck after each other ⁓ back to back. So dollar five slash r slash n, hello.

slash r slash n is how your string is encoded and then dollar five slash r slash n world is how your world is encoded. It's back to back. So this is why you'll get this entire slice but you have to iterate one value at a time. Right? So this is where it would come in extremely handy where a read array will start with plus, which starts with one because of like obviously array starts with a star. So here boss starts with one, then we do a read length.

then we adjust the pos pos equal to pos plus delta because we have read the length and now ⁓ now that we have the length or sorry we have the count of the number of elements right we have the count of the number of elements in the array now we will allocate an array ⁓ allocate an array that can hold any type ⁓ of having that many elements right so if I am returning if I am getting 2 which means I creating an array of 2 over here that can hold any type of element then I am iterating

And for those number of elements, I'm decoding one value at a time. Right? So I'm invoking decode one, starting from that particular position, I'm getting whatever the value it is, followed by a delta and an error. If error is not nil, I'm returning. Otherwise I'm setting that element into the array and I'm incrementing pos equal to pos plus delta. Now, no matter how big my array is, my decode one function beautifully moves forward one after another. It moves forward one after another.

Sneha Mehra (00:11:51)  
covering my entire value that I've received. Right? And now if you look at this, I have just ⁓ used all the examples that was given ⁓ in the Redis's official specification and I put that and it really very beautifully parsed the entire array. For example, this array star 2, star 3 something, it is a nested array, array within an array. So you have ⁓ one big array within which you have one sub array having three elements with all integers.

and you have another array where one is string or rather where two are string, one is string and one is error, right? And it still decodes them very well. This is why ⁓ writing a simple looking code is really important, right? Where we just decoded one value at a time, returning that we process these many bytes and now we have to move forward. We process these many bytes and now we have to move forward, right? And this is how you would be ⁓ parsing and decoding an array which can obviously contain

any kind of value because we are just ⁓ calling decode one. Now array within an array within an array you can have n you can have infinite literally infinite number of nesting it would still work fine right because the code flow is so simple right. Now again here the critical decision was that we have to return the number of bytes that we processed to parse that particular value so that it can be leveraged by other function to start from that location and so on and so forth.

Right, so here you can see how we are always starting from data of post colon. Right, so this is a critical low level design decision that we made, which makes our code very clean, very neat, very easy to read, nothing complicated at all. ⁓ And this is all what I wanted to cover. So in this video, just ⁓ to sum it up, we went through the Golang based implementation of RESP specification.

where we covered five or six data types and we wrote and we saw how to write a decoder for the RESP encoded values. ⁓ We saw how to make a low level critical design decisions that would make our code extensible and ⁓ simple to read and extensible for the future. ⁓ In the next video, now that we have RESP with us, in the next video, what we will do ⁓ is we will ⁓ write our first command, ping and pong. So on Redis CLI, when you send ping,

Sneha Mehra (00:14:16)  
In the response Redis server returns pong. We would implement this very same thing and see and actually connect with the Redis CLI. We would send ping and in the response we would get pong. So we will be implementing our first command of our own, ⁓ our very own Redis implementation. So thank you so much for watching and I'll see you in the next one. Thanks once again.

—-------------------------------------  
2

Sneha Mehra (00:00:00)  
So what makes Redis special? Redis is an open-sourced in-memory data structure store that can be used as a database, a cache, a message broker, and a streaming engine. ⁓ Redis provides some of the most amazing data structures out of the box, which are hashes, list, set, sorted sets, bitmaps, hyperloglog, geospatial indexes, and streams. These data structures that Redis gives us

they help us build a variety of applications like real time chats, message buffers, gaming leaderboards, authentication sessions store, media streaming, real time analytics and whatnot. If you look at the trend of how the popularity of Redis has grown, it is an exponential curve because people ⁓ love Redis because of the flexibility and more importantly, the simplicity of it. It finds its application across plethora of data.

across basically clathora of use cases, right? And that makes Redis special. But Redis, the key highlight, the key highlight of Redis is that every operation that you fire on Redis is atomic in nature. By atomic, what I mean is that when a particular command that you fired on Redis is executing, your other commands, let's say you have accepted three or four connections, it's not that your command execution would

pause in between and some other command will be scheduled. No. When one of your command is executing, ⁓ after that execution is complete, then only another command would be executed. This is the beauty of Redis. This is where it makes its stamp, its authoritative stamp saying that, I am atomic in nature. Any command you fire, you don't have to worry about concurrency at all. Because every single thing that you are putting in is atomic in nature.

While that is executing, there is no context switch that would happen. So for example, when you're putting a key, atomic, adding to a list, atomic, doing set union, intersection, atomic. Just imagine when you're doing set union and not having to worry about some other thread or some other parallel thing, putting the value while you are doing this union. That's the best part. ⁓ This is the power of atomicity that Redis exploits.

Sneha Mehra (00:02:22)  
and more importantly, ⁓ incrementing the value. We know that count++ is not thread safe. ⁓ Here incrementing the value is atomic, which means that even if you accept requests from 10 TCP connections or from 10 clients, all of them trying to do incrementing a particular key, ⁓ the final value will be 10\. ⁓ It will not be nine or eight or seven.

So every command that you fire on Redis is atomic in nature, which is brilliant part on the Redis. ⁓ Second, the data is stored in memory. This ⁓ is the best part. Why? Because ⁓ that's why you would see Redis being used as a cache in most of the applications. And most people, unfortunately, don't know that Redis can be used for other purposes as well. This is where the in-memory storage comes in very handy. But...

But Redis also provides a configurable persistence, which means that you can configure when you start your Redis server, ⁓ can configure it such that you can still ⁓ periodically dump the data onto the disk. So whatever the data is there in the memory, it would dump it onto the disk.

⁓ It would not be deleting anything from memory, but it would be dumping whatever it has periodically to the disk so that if in case your Redis process crashes and it reboots, it does not start from scratch. It can load the last file that it dumped and it can load it and start from the point of the last checkpoint. ⁓ It can also provide partial persistence to write ahead logging, which means that every command that you fire, every update command that you fire will be logged.

in an append-only file, which is an AOF file, which basically Redis calls it AOF file. It logs it everything into that, which you can use to reconstruct your Redis. ⁓ Or you can set up a synchronous replication to another Redis server. So this is where your write ahead log file comes in very handy. And you can also configure the persistence thing. I don't want persistence at all. Just keep everything in memory. Whenever you crash, I'm okay losing all the data. ⁓

Sneha Mehra (00:04:35)  
Other key features that Redis provides are ⁓ transactions, which means that while that one transaction is executing, nothing else would do. ⁓ We'll talk about rollback in detail on how it handles that, but it does provide transactions. It provides PubSub, which means you can have a publisher and a lot of consumers listening to the same Redis ⁓ on a particular topic. When a publisher publishes, ⁓ automatically all the consumers listening to that same topic

receives the message. It's a push-paste message that they would receive. So you can build an entire PubSub and you can use Redis as a PubSub for that. Other key feature is that there is TTL or expiration on the keys, which means that you can say that I'm putting this key in Redis and after five minutes, automatically delete that. It's an excellent feature that would help you keep your usage in check. ⁓ You are putting in value and without any expiry, which means it would keep it there permanently.

you would have memory leaks in your system. With expiration, you can automatically say that, hey, let me put my authentication on that with an expiration of 30 minutes. After 30 minutes, authentication like that particular key would be automatically deleted, which means that user would be automatically locked out. So for anything temporary, this is an excellent place to build. And ⁓ LRU eviction. Key eviction strategy. So when your cache is full, you need to evict the key.

and it has its own configurable key eviction strategy, which means that it would still be accepting the rights. It will not crash, it would accept the rights, but if the cache is full, it would be evicting the key. And that's a brilliant highlight of that, that you don't have to ⁓ manage anything. Redis does that thing automatically on its own. Now, we'll talk about the ⁓ programming models. So...

This is where we need to understand how Redis is able to handle large number of connections. We are still laying the foundation while implementation will get much deeper understanding on how it is actually implemented. But let's talk about the concurrent programming models over here.

Sneha Mehra (00:06:39)  
So here, what the given state is that you have a single process, like we have MySQL Postgres process, you have a singular process. Within that process, you want to handle multiple requests concurrently. How do you do that? So there are multiple ways to achieve concurrency in a single process mode. Now, this is all about concurrent programming. There are two possible approaches. The first one is multi-threading, the most common one.

So here what you can do is, whenever you get a command to execute, whenever a client connects to your database, you spin up a new thread and let that thread handle the request coming in from that client. So for example, if you have request one coming in that wants to do increment k, you would spin up a new thread that does increment k.

If you get request two that does increment k, you would spin up another thread that does the same operation increment k. So every client that is connecting to your database, you are creating new threads for every client and let that thread handle anything and everything. This leads to a problem. This leads to a problem where we have to ensure data correctness. Because what could happen ⁓ is that if let's say the value of k is 10 and I have two threads,

if value of k is 10 and I have two threads they both executing k plus plus right k plus plus on the same value k or the same variable k then what would happen ⁓ is if you do not have proper guardrails what would happen ⁓ is both the threads would read the old value of k 10 both of them would increment to 11 and they both would write 11 over there so even though I have two threads

but when I am doing increment twice because two threads can execute on two different cores or even on the same core that same problem exists that there is a possibility where my final value can also be 11 or 12 and it is an unpredictable behavior. You never know when you would get 11 and when you would get 12\. This is a classic problem of concurrency where your count plus plus is not thread safe. So, ⁓ one way to solve this problem ⁓ is

Sneha Mehra (00:09:00)  
to make ⁓ one thread wait while other executes, which means that if I have two threads coming in, I'll write the code such that they would need to acquire some lock and only one thread ⁓ is allowed to acquire the lock while the other one would be continuously waiting on that. So I initially have two threads coming in and they would both be waiting to acquire the lock. One of them would get it

and it would move forward while the other one continues to wait. When the one thread executes, so your K++ is a critical section. One thread would be executing K++ while the other one waits. ⁓ Then this thread goes to release lock. As soon as it releases the lock, the other thread is allowed to move in and it would do K++. Because of this exclusivity, you are...

putting that guardrails so that you never end up with 11\. You will always, always, always end up with 12\. This is what you would have to do to ensure the data correctness. And the way you can achieve this is through mutexes and semapers. Very common strategies to do that. ⁓ These fall under pessimistic locking. ⁓

out of scope of this part because we are building a single threaded database. But if you are interested deep dive into mutexes and semaphores. ⁓ And through this you can implement ⁓ locking or basically pessimistic locking. Now let's talk about the more interesting part which Redis adopts. It's called IO multiplexing. ⁓ So ⁓ this ⁓ is an apparent.

concurrency model, which means that you are not achieving true concurrency. You are achieving apparent concurrency. So the idea here is that with threads, it is possible that two threads will be scheduled on two different cores ⁓ of your computer or of your server. So you are getting true concurrency. ⁓ Here it's apparent concurrency. So here the idea is that any time, like why

Sneha Mehra (00:11:10)  
Are we even talking about this? Because, ⁓ I-O, whenever you're doing a disk I-O or a network I-O, system calls happen in the background. Happen in the background as in, let's say you want to read from a socket, you have to fire the read system call. Read system call, it's same in all operating system. Every single operating system exposes a set of system calls that you have to use to do I-O operations, be it disk I-O, network I-O or anything.

Right, so now when you would want to read something from the socket, you would initiate a read system call. And this read system call is blocking in nature, which means that if I have two devices or two servers connected over a TCP connection, and if one of them fires a read, let's say my first server fired a read system call that it wants to read from the socket, right? This is a blocking call, which means that the call would not proceed

until my connected client sends something over the network. This is painful. This is extremely painful because with I.O., if you are waiting on something, you are not making any progress and you are just blocked. So, can we ⁓ do better over there? That's the idea behind I.O. multiplexing. So, can we multiplex when we are waiting for I.O. to complete?

So here, hence we cannot just invoke read request on a socket unless we know that there will be someone sending the data to us. For example, ⁓ if you don't know ⁓ if the other party is going to send you the data and if you fire a read system call, then you are just blocked there for a particular time or you are just blocked unnecessarily. So that waste of CPU memory and whatnot, right? So this is what

makes us use IO multiplexing in such cases. So ⁓ to handle this, a popular approach that we just discussed is multi-threading, where while one thread is waiting, we can have other threads getting scheduled on the CPU to get things done, right? That is where multi-threading based approaches comes in handy. That is one way to solve that problem, that each thread would fire its own read call

Sneha Mehra (00:13:38)  
while it is waiting for the IO to complete, other thread who are ready to execute can be scheduled on the CPU. ⁓ And this is the core idea behind achieving true concurrency using multi-threaded system.

That's why most multi-threaded, most all multi-threaded web servers are implemented exactly like this. Whenever a new connection is formed, it it's spun up a new thread and that thread does take care of the IO that needs to be done. This way, if one thread is blocked, other thread can be scheduled because it goes in the block state, operating system schedules another, ⁓ another thread which is ready to be executed. Right?

But when we do this, as we just saw, we have to use mutexes and semaphores to protect our critical section. And which makes things slow because ⁓ my two threads. Now, why is it making things slow? Because what we saw was when you have two threads, you applied mutexes or semaphores so that you can protect your critical section. Now, these threads were not waiting for any IO.

these threads were ready to be executing. They were not waiting for IO. They ⁓ just wanted to do count plus plus or K plus plus. They were all ready and roaring to be executed. But what you did is you added this mutex, this pessimistic blocking, which allows only one to move forward. This is the pain point. So a thread that is ready to be executed or can move forward, you are unnecessarily blocking it.

because you want to ensure correctness of the data and that is also important. ⁓ So, can we do something? An idea over here is can ⁓ something notify us when there is some movement in the thread or sorry in the ⁓ IO. So, let say you have a TCP connection open ⁓ ping me when there is something happening there.

Sneha Mehra (00:15:41)  
then only I will fire a read request. ⁓ That's a very simple analogy, ⁓ right? But this is the world of IOMultiplexing. So the core idea here is that we use IOMultiplexing, we use IOMonitoring calls to monitor the socket and fire read request on the one that has some data. ⁓ This is the core idea that we not fire read request until we are 100 % sure that there is some data to be read. And this is the world of IOMultiplexing.

⁓ We'll go in detail. So for example, here, what and this is what your event loop also looks like. This is where Redis also operates. So any single threaded database or process that claims to be supporting ⁓ large number of concurrent IOs, this is what they do behind the scene. This is how the execution looks like. Everything is still a single thread.

There is nothing changed over there. It's not multi-threading. ⁓ Event loop is not a separate process. It's not a separate thread either. Everything is running in this one thread ⁓ only. Right? And the idea is ⁓ that, let's say, you ⁓ get a lot of TCP requests at the moment. Your server just started, you got a lot of TCP requests at the moment. So what did you do? You accept a lot of TCP connections using system. We'll talk about implementation in the...

in the next video, ⁓ but just understand or just absorb the essence of the idea. ⁓ So you'll accept large number of TCP connections over here. ⁓ And you read from one of the sockets. So let's say you accepted connection from four, then you read from one, the one that you read from, like you got to know, you got to know that one of them has data that you can read. You read from that one socket.

and then you execute because you read a command. Let's say we are building redis, you read a command and you're executing that command, right? And let's say took you three time units to execute it. And then you again check, is there any other socket which has some data for me to ⁓ use ⁓ or ⁓ it has some data for me? And you found there is one more socket and you read the data that came in, which is the command and then you executed it.

Sneha Mehra (00:18:04)  
And then you check, are there any socket that has data? No. You check, are there any incoming TCP connections? Yes, there are two. So you accepted them as well. And then you check, are there any sockets that I can read from? You found one. So then you started executing and so on and so forth. ⁓ Here, you can very clearly see how we are accepting multiple TCP connections, picking one that has the data immediately that it has sent you.

using that incoming data, are executing the command that has been provided and then moving to the next. ⁓ This is the idea of a single threaded event loop. This is what event loop does. Any IO operation, whenever there is some data, something, then only you are taking this actions and system calls tell you when there is data available for that. You don't have to do anything. System calls are there to use. You just have to use the corresponding system call.

⁓ And this makes Redis extra special. This is how your Redis evaluation actually happens. ⁓ So, just to summarize, Event Loop is not a separate process. It is neither a separate thread. Everything just happens in this one single thread. Accepting the connection, reading from the socket, executing the commands, everything happens on this one single thread. Now, you'll say, but then how is it fast?

How is it Redis being able to give us such high throughput? The idea is pretty simple. What Redis beautifully exploits is the fact that network IO is slow because ⁓ someone sending you the data over the network is a very slow process. You reading the data is a very slow process. In most cases, when you are doing an IO heavy server, let's say your Redis is an IO heavy use case. ⁓ Anything that you build,

Your Redis is a classic example for that. Let's say you're building a web server that is also IO heavy. You get a request and then you process it, right? So in those cases, you're waiting for a very long time to receive commands or request. ⁓ But in case of Redis, ⁓ your operations that you do, ⁓ increment, adding to the list, they are all memory operations because Redis is an in-memory data store. And because your...

Sneha Mehra (00:20:23)  
operations when you receive a command, the operation that you have to do, are all in-memory operations, they are extremely fast. It's just about adding a new value in the array or updating a value in the array or adding a node in the tree, something like that. It's extremely fast. So, upon receiving the commands from the client, Redis can very quickly execute them. So, the time it takes for Redis to execute the incoming command, it's extremely fast. This is what Redis beautifully exploits.

So this is why Redis made this conscious decision of keeping itself single threaded so that you don't need the complexities of mutexes, semaphores and unnecessarily your threads getting blocked. The threads that are ready to be executed, they are getting blocked unnecessarily. That's why no complexity of mutex semaphores or data correctness and whatnot. Everything is single threaded.

And by doing IO multiplexing, it is still accepting large number of TCP connections concurrently, evaluating them one by one, given that execution after receiving the command is just an in-memory operation, it would happen extremely fast. That's why Redis is able to support large number of TCP connections while being extremely fast at executing them, because every operation is just an in-memory operation.

And this is what makes Redis extra special. So that's it for this video. I hope you had fun understanding what makes Redis extra special. In the next video, we'll be building our own TCP server in GoLine and laying the foundation of our own Redis implementation. ⁓ Thank you for watching and I'll see you in the next one.

—------------------------------------------------  
21

Sneha Mehra (00:00:00)  
In this video, we will be finally implementing transactions. ⁓ One of the most important features of any database out there is transaction. ⁓ So Redis has a different take on transactions, which is what we would look at. ⁓ And the idea is very interesting, very simple, but very fun. And it will be extremely fun to implement it. Our code would go into a bit of restructuring, but we have to do that. But it's still most of reuse.

but you'll see how ⁓ interesting it becomes to implement transactions, right? Okay, so first, like always, let's look at ⁓ Redis's behavior ⁓ and how, and then we would be basically mimicking it. So on top right, we have normal Redis server running on port 6379\. I use Redis CLI, Redis CLI hyphen, ⁓ I can directly connect to that. And now the way Redis transactions happen is through a command called multi.

⁓ Multi implies beginning of transaction. So if I fire multi, you see Tx there. ⁓ This Tx implies that we are in a transaction mode. Now, whatever command you fire, whichever command you fire, your command would not execute. It would be queued up. So for example, if I do set k,v, ⁓ it's not doing actual set. It is writing queue. So your command is queued for execution. Then I do incr k1, ⁓ incr k2.

all of the commands that you fire, it will all be queued. Right? Okay. Then to commit a transaction, you can do exec, which means execute this multiple things at one shot. Then it would execute and you would get response for all the three commands individually in an array response. So you first do set kv, you return okay, then you do incrk1, it returned 2, then do incrk2, it returned 1 because k1 was already set to 1 or something like that. Right?

So you get the three commands that we sent executed in one shot. You got three response in one shot, right? ⁓ Okay. Now if you fire, ⁓ let's say I fire exact again, ⁓ it gets an error that error exact without multi because our transaction ended. ⁓ So then we to initiate another transaction, we have to do multi again. And let's say I do set K comma V, ⁓ it did set. And then I can do, let's say I want to abort it. I can fire discard and it would be aborted. ⁓ Your transaction is aborted.

Sneha Mehra (00:02:24)  
Right? Okay. ⁓ One hunch with Redis transactions is that when you do exec at that point of time, all the commands get executed on the like all the commands get executed. Right? But in case in middle of that something fails, right? Then there would not be rollback. There is no concept of rollback. Redis wants to keep it extremely simple. So all the commands would execute in one shot. Right?

When these commands are executing, a transaction is executed, you may have enqueue hundred commands there. When a transaction is executing, other client's command would not be evaluated as well. Like they would be waiting for this transaction to complete and then they would be doing it. So it's like literal sequential execution and anyway it is a single threaded so it is going to be sequential but just being explicit about it that it can enqueue a lot of command together but it would not execute them.

until you fire an exec command. And when the exec command is executed, it would execute all of that in one shot. And then some other clients request would be shared. ⁓ So it's going to be sequential, very simple. But now we have to implement it in our code base. So first of all, what we have to do over here is we have to identify each of the client because we are in, ⁓ my bad. We should also see what happens when multiple

when basically multiple CLI connects. ⁓ Now, ⁓ you fire multi, ⁓ so ⁓ here also transaction is started and when I fire multi, here also transaction is started. I can do set k,v. ⁓ Here ⁓ also I can do set ⁓ k,v. Here we can see it's not that when multi started, other clients cannot even send the commands. ⁓

two clients can concurrently enqueue their commands in the transaction, then one execs and another execs. Right? That's the idea. So which means that we need a way up until now our implementation, our DiceDB implementation was very static, extremely static, which means, stateless, not static, stateless, which means we got the command, we evaluated and we send a response. Right? But now we have to do that ⁓ multiple clients can be connected

Sneha Mehra (00:04:47)  
and we would want to accept the commands from them and keep them enqueued. When exec is ⁓ invoked, at that time, we have to execute them in one shot. Right? So, that is how our changes would begin. So, first change we would be making is to have this ability to store ⁓ every single thing. Okay. So, here we'll start with that. Connected Clients. So, this Connected Client is a map of int to core client.

Now I've just created an object. So first of all, let me tell you what this is. So every time we know that when a client is connecting to us, ⁓ right? When a client is connecting to us, it is, ⁓ we get a file descriptor, right? So when our socket connection is accepted, we get a file descriptor. This key is a particular file descriptor. And this core.client is what we have written. The FD.com that we had written gets now changed to client.

it is still a would expose the same read write interface, but apart from that, it would store other information as well. Because now what we have to do is, we have to not just do read and write, but also when we are in the transaction mode, we have to enqueue all the commands in it. So that is where we would want to store that information somewhere. So which is where we are doing it? We are storing FD, CQ is the command queue and is transaction which means is this client.

executing the transaction. Is it in the transaction mode or not? Right? So this is where I've changed the FDCOM object to client. So whenever any client connects to this, an object, a unique object is created for that client and we are storing it in the connected client's hash map or hash table. Right? And then we are ⁓ adding functions like transaction begin, transaction execute will basically come to that. Right? So this gives us a very simplistic view of our code.

So connected clients is there and I've written an init function which creates a map out of it, right? So ⁓ which is the actual map object here. We only declared the object, here we added memory to that. Okay, then what next? So here what we would do is we would want to add it ⁓ every time a user is connected. So user or when a client is connected, we get this file descriptor over here. We fire this accept call, we got a request, we fire this accept call, we get this file descriptor.

Sneha Mehra (00:07:12)  
This is where we are storing it. ⁓ ConnectedClient of FD is equal to core.newClient of FD. So if I do this core.newClient just creates a new object out of it and returns that object. ⁓ Right? So we storing for each file descriptor, we are storing this new client object for each file descriptor, which holds the file descriptor and empty command queue. ⁓ Right? So this is where that happened. ⁓ And then now comes the part where we were when

any command we are receiving, ⁓ we were simply executing it. But now what we would do is, ⁓ because we have to maintain a state, what we would do is, from our connected clients hash map, ⁓ for the given connected client, we would get the object. ⁓ And this was the object, this was the client object that we want. ⁓ I just kept a nomenclature comm because the last object was also named comm, but this is the actual

client object that we want. If the client does not exist, we continue because that may be terminated, something it does not exist in the hash map due to any reason. We are just continuing it with the other clients that are connected. And then in the read commands, I'm passing that same client object. And then when ⁓ a client terminates the connection, here ⁓ we were ⁓ closing the file, here we close the file descriptor, and now we have to also remove it from the hash map that we had.

So delete from connected clients the event FD that we have, right? So the file descriptor of the socket, we have to delete that and we did it over here, ⁓ right? So whenever a client gets connected, a new objects get created over here, pushed into the hash map. ⁓ Every time we are about to receive a command from the client, what we are doing is we are receiving the command from the client. We are first getting the unique client object for that TCP for that particular client so that we can maintain that particular state.

and then on that particular same object in case our client disconnects, we are just deleting it from the hash map. Right? Okay. So this is how we did ⁓ that unique state management. So now if a same client is connected, because what we had to do was we wanted ⁓ to identify that, hey, we received a particular number, is this from the same, ⁓ from which client did we receive this command? ⁓ And because we are queuing them, we would want to enqueue them in that one list, which is where our client objects would.

Sneha Mehra (00:09:37)  
come in. Okay, so this is where we are doing this state management of my sockets. What next? Now, let's look at implementation ⁓ of ⁓ the transaction. So multi. Now we know that we required a response called queued, so I just created a global object for that, which is written plus queued slash r slash n. And then I'm creating an object called transaction command. This is just about if the transaction is in exec or discard, otherwise it...

you need to queue. So the idea is we saw that our transaction when we start the transaction using multi then we can pass in any command. Any command we pass it gets queued ⁓ and ⁓ we can either fire exec or discard. When you fire exec your transaction commits and if I discard everything gets discarded. Right? So anything except this so these are my transaction commands which is exec and discard which is what we would be using to check if it is in this then do this otherwise that. Right? Okay.

So then we'll skim through that and we will go at the bottom. Now here a bit of restructuring needs to happen. So first of all, let me talk about a bit of restructuring over here. The first thing that we did is earlier we used to execute a command and directly add it to buffer. I've just created a separate function called execute command, which takes in a command and a client object. And it basically executes it and returns us the response in bytes.

this function takes this and returns a response in bytes. So if I fire ping, ⁓ get pong, ⁓ will literally eval ping will literally send me the bytes instead of writing it to the buffer directly, I'm returning it. This function would come in handy when we are writing transactions. ⁓ So execute command, I've just taken this command a bit out, right? So now we have execute command like this, ⁓ and I've written another function called execute command to buffer. All it does is executes the command, ⁓ gets

byte array in the response and it adds it to the buffer, buff.write into this, which we already did. ⁓ It just created two separate functions for that because we would need to revalue this again. That's why. Okay. And then eval and respond. Okay. This is where the fun begins. So eval and respond earlier, this entire code was written in eval and respond. ⁓ Entire code was written in eval and respond. Now, our eval and respond looks something like this.

Sneha Mehra (00:12:02)  
This was their response. We added the byte buffer here. Now what we are doing over here is we are iterating through all the commands that we received. Right? All the commands that we received. Now these ⁓ are normal client commands that we always used to receive. We got that and now we are checking. If my client is not in the transaction mode, right? If my client is not in the transaction mode, what does this mean? This means we have to do normal execution like we always used to do.

Execute command to buffer, we pass in the command, we pass in the buffer and the client object and we do continue. Right? So for each command, we are just executing it and adding it to buffer like how we used to do it. Right? But if we are in the transaction, then it would come over here. If we are in the transaction, we would have to check. If it is a transaction command, which means either exec or discard, we have to, sorry, if it is not a transaction command, then we have to enqueue it.

Right? And send the response queued. That's what we did. Right? We enqueued it and we send the response queued in it. And if it is a transaction command, then we have to execute it. Execute command to buffer like we always used to do. Right? Okay. Now, here let's see what this transaction queue is. But before that, how our transaction begins? Our transaction begins when we fire multi. Right? So that is where in the handler, we are handling multi, exact and discard.

So multi in which I'm invoking c.transaction begin multi and eval multi. Eval multi simply responds okay. That's all it did. We are also doing that same thing over here. And multi, ⁓ just doing the c.transaction begin. Now c.transaction begin simply sets c.isTransaction equal to true, which means this particular transaction, ⁓ sorry, this particular client is in the transaction mode. Right? Okay. So that is there. And then exec.

Now that is where the part begins execution. How is my transaction executed? So before execution, let's see enqueue. This is where we are enqueueing it. So every time we get a command, we are just appending it in the list that we have. Right? So C.CQ. CQ is the list of commands that are queued in that transaction. We are just appending it to that. That's all we do. Right? We are simply appending it to this particular queue. So whenever we are in the...

Sneha Mehra (00:14:26)  
transaction stage multi has been fired then any command that we fire we are continuously enqueuing it in this particular queue, right? Okay, it's there. Then what do we do? We do exec when someone invokes exec we do this. Now how execution would happen? Execution is doing nothing but ⁓ iterating through all the enqueued commands executing them and adding it to the buffer. Simple, right? Because that's what we are actually doing. So

when I hit exec if let's say transaction enqueue three commands we have to execute all three of them and all three of them we have to add it and return it as an array so this is what we are doing so buffer.write string remember how arrays are encoded with resp star and then the length of the array and then followed by n resp encoded values so star %d %d is length of array now length of array would be what

the number of commands that we have, those many things. So len of c.cq. So we are doing that and then appending those nResp encoded values. So this is what we are doing. We are executing the command and doing buffer.write. It would continuously be adding it to this bytes buffer. And then when we have that, we are just returning that particular response. That is the added response that we have, right? Okay. So this is why we wanted execute command to be separate because what execute?

Earlier the thing that we wrote was directly adding it to the buffer. But now in some cases we have to add to buffer. In some cases we don't want to add to buffer. Right? So that is where we are just splitting that thing into half. Like first where it's all about executing the command and returning us a byte array. And the second one where we can append it to the buffer if we would want it. So this is where we are just executing the command ⁓ and we get that particular response and then we move forward.

Right? Okay. So now here what we are doing is we are when we are firing exec, we are getting a particular response and then, ⁓ we fire the exec if it is not in the transaction. So firing exec without multi, it's an error, right? So firing discount without multi is an error. It's just error handing here. And then we do transaction.exec. Transaction.exec will do that. It would return in the response the byte array that we have, the array of values.

Sneha Mehra (00:16:52)  
But while it is doing that, once it is done evaluating it, it reduces its command queue to zero because now the transaction is done. The queue can be boiled down to zero. This is where I'm just discarding it. Golang would do garbage collection and I'm just creating a new fresh queue out of that and I'm setting transaction to false and I'm returning the response over here. For transaction discard, I'm just creating a new queue and then setting transaction to false. ⁓ This is how our exec and discard would work. ⁓ As simple as this.

And now, whenever you'll get any request, like whenever we are getting any request over here, when it is in multi. Now here, because we are having this hash map of connected clients, multiple clients can be connected over, ⁓ can be connected to the same server, can initiate multi, their corresponding commands would be encued into their corresponding list, right? And whenever someone execs it, right? So exec would also be a command.

when that person or when that client executes that, then only that would execute because now we are not doing a context switch because it's single threaded. We are not doing a context switch until the first one completes and then ⁓ the for loop would move forward and we would accept another connection and get read and then respond. Right? So this is where that particular state management would automatically happen that at one point in time only one transaction would execute. While multiple can accept

the request, accept the commands and enqueue them. But when one is exec, ⁓ when one is executing, literally exec command being fired, other would be waiting. It is almost, it is by definition it is happening. We don't have to explicitly handle it anywhere. It is by definition that it is happening that way. And because that's the beauty of single threaded systems. You don't have to do literally anything. And we just implemented transaction. Now, what about rollbacks? Rollbacks, Redis does not do rollback, right?

Redis says that if your transaction while executing breaks due to any reason, your data will be in inconsistent state and you have to be okay with it because Redis wants to keep it simple. So there is no concept of rollback. ⁓ There is only that when exec happens, ⁓ commands of only one transaction would happen. ⁓ And then once that exec is done, then the other transaction would be scheduled. ⁓ And we automatically handle that without any hiccups. ⁓

Sneha Mehra (00:19:15)  
And this is how you can implement transactions. Let's see this in action as a very quick demonstration of this. Like always on the bottom ⁓ left, we would have our own implementation running. And now on the bottom right, what we would do is we would connect to port 7379\. ⁓ When we did that, I'll first fire multi and then I can fire set K1 V1. ⁓

queued set k2 v2 then INCR k3 ⁓ I did that and then I do exec ⁓ you get ok ok 1 right response exactly how redis gave us response exactly how we got like literal one-to-one correspondence response as an array which is exactly what we got over here as well right and if I can would want to just test it against multiple clients I can very well do that I'll just need to change the port over here ⁓

and I can pass in first of all multi set k1 ⁓ v1 set k2 v2 inc r k3 and ok here also we do the same thing set k1 v1 ⁓ set k2 v2 ⁓ set k2 v3 and then we will do inc r k3 ⁓ ok now if I do increment sorry if I do exec

on let's say my right side client if I do exec over here we are getting ok, ok and 2 INS here k3 is 2 k2 k2's value is v2 and then if I do ⁓ exec over here ⁓ we get ok, ok 3 because we incremented one more but now if I do get of k2 we get v3 because this transaction happened afterwards right normal simple transaction concept nothing really fancy

⁓ And that is how you implement transactions in DiceDB. can find this entire source code on github.com slash diceDB slash dice. We did an exhaustive code walkthrough on how you can implement transactions. It was really simple to implement transactions, but ⁓ it actually worked. It actually worked in the first shot, right? Pretty interesting. And yeah, that is it. That is it for ⁓ this one. I hope you found this transactions amusing.

Sneha Mehra (00:21:35)  
I would highly, highly, highly, highly, highly encourage you to implement this on your own so that you understand how databases are built because nothing is really hard. You just need to have a very ⁓ simplistic approach towards building databases the exactly way, exactly how we did it over here. You can find this commit right there in the source code. I'll also link this in the description down below. Thank you so much for watching.

—------------------------------  
6  
Sneha Mehra (00:00:00)  
So, in the previous video, we saw the implementation of RESP in Golang. In this video, we would be implementing the most basic command that Redis has called ping. ⁓ So, before we jump into the code, let's see how ping fares with normal Redis server. And then we would try to replicate that same thing ⁓ on our own implementation. So here I am starting a normal Redis server listening on port 6379 in the terminal.

Below on the left side, you can see that I am just connecting it with the normal Redis CLI that we have on port 6379\. ⁓ And now my Redis CLI is connected to the Redis server. Now, if I fire ping we and we hit enter, we get Pong. If you clearly notice Pong does not have a double quote, which means that it is a simple string, not a bulk string. Then we can fire ping.

With a message, let's say hello, ⁓ and if I fire this, I get hello with a double quote, which means this is a bulk string. Right. And what if if I fire ping h e double space W O R L D? ⁓ Enter, we get wrong number of arguments for ping command. So this means because we have passed hello and world, these are two different arguments. Correct. So arg1, arg2. So here it clearly shows that we have

A normal ping implementation that returns pong, ping with a hello returns hello in double code, then ping with hello with world as two arguments. If we pass, we get this response which says that wrong number of arguments in the ping command. This is what we'll be implementing today. Right? Okay, so let's jump right into the code. ⁓ And here's this code again, it's the third commit in this part. The repository is github.com/slash dice db slash dice ⁓ and go.

And basically check out the third commit. ⁓ This is where this entire code of ping implementation can be found. Okay. Let's start with the main file. Your sync underscore TCP file where we are accepting the commands, or we are accepting not really command, we are accepting a stream of bytes from the client. Right? Okay. So here the first thing that we change is that we introduce an object called Redis CMD. This Redis CMD is because we are building

Sneha Mehra (00:02:24)  
Our own drop-in replacement. What we are giving is we are giving it a proper structure, which means that in any command that we fire on the Redis, we have the command, let's say ping, and then there are n number of arguments with it. When we are doing put k value, we have put as the command and k and v as two arguments that we are passing. ⁓ This structure will basically hold this part. You have command and you have args with that.

Right. And arcs would be normal string arcs, and then depending on the command, we would be converting it to whatever we want. Right. Okay. So now that that is cleared up. Now our read command that was reading things from it. Now instead of just reading a normal byte from that, what we are doing is we are reading it in the buffer, all the bytes that are coming in from the client, and then we are decoding it into array string. So

Redis, when Redis CLI or any Redis client wants to issue a command to the Redis server. So a command typically has a root command and arguments. ⁓ All of these are sent to the server ⁓ as an array of strings. ⁓ So if I am doing put K value, no matter what type of keys, what type of value it is, it will be sent as an array of strings. ⁓ So that is why we have just written a small decoder that is specifically converting.

That is specifically converting your array of like basically whatever array of strings you get into instead of doing a normal array of interfaces, we are specifically typecasting into array of string. This is just doing that. So here you see we invoked the normal decode function on the stream of bytes, ⁓ and this we are doing in RESP file because it is all related to Redis serialization protocol. We are doing a normal decode over here and then we are typecasting it into interface like

Array of interface, and then we are analyzing or we are initializing a ⁓ an array of strings of length equal to the number of elements that the array of interface has, and then we are just typecasting it into string and putting it over here and returning it. So this function is doing a normal decode just on top of that, a little more ⁓ like a little more beautification of it, a little more a cleaning of it so that we get nice array of string instead of array of interface and we doing it every time.

Sneha Mehra (00:04:46)  
Just writing a simple function to take care of that. ⁓ So this is your decoding to array of strings. ⁓ Once we have that token, now what we are doing is we are creating a Redis command object in which we are passing the root command, in which the tokens of zero is a root command and everything else becomes the argument. So tokens of zero converted to uppercase becomes your CMD and tokens of one colon, which means every other thing that subslice of it becomes the argument.

Right. So this is what will be changed in the read command part. And if we scroll down below ⁓ in the respond function, now what we are doing? While we are responding it, we want to respond with the connection. So with I'll just skim through down. Yeah, okay. So this is the main infinite loop that we are running. Right. Where we are reading the command in case there is an error, your client got disconnected. Done. But then we are responding. So depending on what command we got. Now read command is returning.

The ready CMD object, which has command and arguments. And this command is what we are passing to the respond function because what we are doing is we are getting a command and we want to respond to that. So in respond function, we are passing in the TCP connection ⁓ and the command that we got. And what respond command would do is respond command is evaluating and responding.

So eval and respond. So this is what we have written in the core module because it's the core functionality of Redis. Eval and respond. Its job would be given a Redis command, evaluate it and respond. So it takes in the same CMD and it takes in ⁓ the TCP connection and whatever error we got over here, in case any error, in case for example, number of arguments that we just saw with the ping command, this is where we would be capturing that error. If error is not equal to nil, we would be responding an error.

So let's take a look at respond error function. Respond error function does nothing but writing it on the TCP socket stream of bytes where we are doing string formatting. So we know that ⁓ whenever we send an error ⁓ over the TCP connection to the ready CLI, it needs to be encoded. Encoded into RESP format. How is error encoded? Error starts with an ⁓ it basically starts with the hyper, it starts with the minus sign, then the error string, then slash r slash and

Sneha Mehra (00:07:03)  
This is exactly what we are doing with ⁓ fmt.s printf. Right? So the normal s printf command, this would output a string where this error string gets pasted over here just like normal printf statement. f printf like s printf would just iterate would just put out a string instead of writing it to the console. And this string is converted to the bytes and is written over the socket. This is a normal responding to the error. So in case there is any error.

We are capturing it and we are responding the error back to the client. Right? Okay. Now the only part that remains is to check the eval and respond function. And I if I click on the eval and respond function, this is where you see the input to this eval and respond function ⁓ is the Redis command and the TCP connection. So our job would be depending on what command or depending on which command is sent to us, we would be triggering the corresponding eval function. For example,

The for the ping command that we got, we would be triggering an eval ping and default. I am just for now, for now, I am just triggering an eval ping. Ideally, it should be an error, but I am just triggering a simple eval ping just to not overcomplicate the things for now. We'll be handling it later. Right? So, in both cases, we are just returning the response of the ping function. But what eval ping is taking? ⁓ It is taking the arguments of the rediscount because now that we know that we are

⁓ handling the ping over here, we just have to send the arguments whatever is sent to us because first thing like the ping itself does not matter, the arguments matter to us during evaluation. So every function or every single command that we are handling will have the function, the eval function of that will have the exact same signature. So if you'd want to generalize it, we can do that. Right? Okay, let's take a look at what eval function does. Eval ping function does. So it takes the arguments. So here basic checks. ⁓ First of all.

We are just declaring a variable which would be sent as a response. Right. If the length of argument is greater than 2, we saw that when the length of the argument was greater than 2, it threw an error called error wrong number of arguments for ping command. I just copy pasted it over here. Right. So if ⁓ the ready CLI passes us more than one argument, we would be responding with this particular error function. If the argument is zero, we are responding pong.

Sneha Mehra (00:09:25)  
And if the ⁓ the length of the arguments is one, which means some argument is passed, we are taking that particular argument. Right. But what we have to do, we cannot just respond this normal string, we have to encode it into RESP. So that is where we are using the encode function. This is now what we would write. So encode function job is to take the raw type ⁓ and convert it into and rather encode it into RESP format. First, in the last one, we wrote the decoders for that.

Now we are writing encoders because now the server has to respond in the RESP specification so that our client also understands it. ⁓ So now what we are doing is we are triggering encode. So if length is 0, we are encoding normal pong, and if the length is 1, we are encoding the first argument that is passed to us. But what is this true and false? So here you saw when we fired ping with no arguments, we got pong, which was a simple string because it did not have codes in it. ⁓

If we pass in some argument, we saw it in double codes. Just to quickly show you that here you can see if I pass in ping, I get pong, but if I pass in ping hello, you get hello with double code. So first one is a simple string, second one is a bulk string, which is what we would also be need to handle. So this typically is for us to handle that. So the encode function takes any value.

Which is I'm just accepting interface, right? Which means any value can be passed over here and an argument which says is simple or not. Because string is the only type in Redis which has a bulk string and a simple reply. Just for the sake of that, I'm I have to pass that. The default value of this would be false. Like in most cases, it would be ignored. Only if it is string, we would check if it is simple or not. Otherwise, because they have two different specifications, right? Okay. So now given value, what we would do is we would apply a switch case. Switch case on the

Type of it, what or which type ⁓ is the value that is passed to this encode function. If the type ⁓ is string, we do something. If the type is string ⁓ and we ask to encode it in a simple string format. The simple string format starts with a plus, then the string that we would want to send followed by slash r slash n. That is exactly what we have passed over here in Sprintf, which would split out the string.

Sneha Mehra (00:11:50)  
We are converting it to bytes because encode function takes any value and converts it into bytes that you can send over a stream of socket. Simple. Bulk ⁓ string encoding is you send a dollar, then followed by the length, then slash r slash n, then followed by the string, then slash r slash n. That is exactly what we are doing over here with s printf where we are passing the length.

And the actual string that we are encoding. For now, I've kept it just simple. Our encode function is just handling the strings because that's what we are implementing. Ping, the only command that we are implementing is ping. So that is why currently our encode function is super simple, it only handles ping. When we implement other data types, we would be adding those specific types over here. ⁓ And this is it. This is what you need to power a simple ping command. Right now.

To see this in action, let's take a look at how it fares. So, up until now, we were connected to port 6379\. Let me quickly first of all run our server. Our DIE server is now running on the top left at port 7379\. And on the right side, you see the Redis server running. Now, at the bottom of this, now what do we have? ⁓ Is I am now connecting to port number 7379\. Enter.

You can see upwards where it says that one client got connected, which means now our C L I our ready CLI is connected to ⁓ our server 7379\. Now if I send ping ⁓ we get pong with no codes, which means it's a simple string. If I send ping with ⁓ H E L L O, ⁓ we get H E L L O with codes, which means it's a bulk string. And if I pass in ping hello ⁓ world, we get error.

Error wrong number of arguments for ping command exactly the way your normal ready CLI with normal ready server was working. The same Redis CLI with our server is running in the exact same way, which means that for the ping command, ⁓ our server, our implementation of server is a drop-in replacement of ready. So you can take out Reddit server, put this server, and it would work just fine. Right? And this is how you would implement ping command. But

Sneha Mehra (00:14:13)  
We are not done yet. ⁓ Now, what if we would want to compare how good our server is? How fast our server is. Like because we know Redis is mighty fast. Let's say if I would want to find out how fast our server is. So Redis by default ships with a Redis benchmarking tool. The Redis benchmarking tool comes with Redis out of the box. You can fire a command called Redis hyphen benchmark with some arguments ⁓ and it would fire it on the

On the Redis server, you ask it to. ⁓ So, this is what we would be doing. We would be using because because our server, the server that we implemented is a drop-in replacement of Redis, we can reuse the tooling that Redis already has to test how good our server is. Right. Okay. Now let's do this. In the Redis benchmarking tool, we say that my ⁓ in the first argument that we pass is hyphen n 100,000 or rather basically 10,000, which means that tested for

10,000 times fire a command. Which command, which ⁓ benchmark or which test you need to run? It's called ping underscore bulk, which means firing the ping command. As simple as that. So because the only command that we implemented is ping. So that's all we can test. Right. So 10,000 times fire the ping command with one concurrent connection. Up until now, we have not connected, we have not built anything around multiple users. It's only one concurrent connection that we have. In the first second video, we saw how it was.

Doing it synchronously when one connection terminates, then the second was starting. So that's why we cannot test it with multiple connections at the moment. We will do it in the future video. But so far now we are running this benchmarking with one concurrent connection on localhost 6379\. Now if I run this, you can see this running and it shows that 10,810.81 request per second is what normal ready server handled.

Now, if I change the port, our server is running on 7379\. Now, if I run it, this was 10,800. ⁓ If I run this, our server took 11,695. We bit like we did beat the Redis benchmark, the official Redis benchmark ⁓ on Redis server versus our own implementation. R on implementation right now it's faster. It looks like it is faster than the ⁓ usual Redis implementation. We have implemented in GoLang the original Redis code.

Sneha Mehra (00:16:39)  
Is written ⁓ in basically C language, right? But obviously, Redis is much more complicated. This does not mean that we beat Redis. Redis is much more complicated, it does much more error handling than what we handled. It is much more efficient in most cases, right? So we are not claiming that we beat Redis, but this is why we chose to build a drop-in replacement because now we can use existing Redis toolings to test how good our code is. We can use Redis' ⁓ unit test cases to test how.

Correct our implementation is as well. So if I just one more time just for the sake of it, because I cannot see Redis loose, we we have written in Golang. Golang is not as fast as C. If I just run it once again, 6379 shows 1224 and 7379, it is 1157547\. So now we are lesser than Redis. And this happens because I'm running it on my local machine, depending on which other process is running, how it is fed, how CPU is scheduled.

Like, sorry, how you how your process is getting CPU to execute it, a lot of factors are there. So that is where you see a comparable performance. Be it Redis, the actual Redis or R on implementation, you see a comparable thing. Right? And this is how you implement ping command in Redis with that handles all cases. No arguments, one argument, an error case. And just ⁓ to reiterate on the part.

That the source code is there on github.com/slash dice db slash dice. Check out the third commit. This is what we have implemented in this one. So that is it for this video. In the next video, we will be like this Redis, ⁓ the R on Redis implementation right now ⁓ is not at all concurrent. If I run, if I make two clients talk to each other, just a simple example. If I take instead of passing hyphen c equal to hypens C1, if I pass hypens C2. ⁓

The execution would stop. This would go forever and ever and ever. Because first connect because it is initiating two connections in parallel, and one connection is just waiting for it to be executed. The first executed, but it did not terminate. It is waiting for the second one, it is waiting for the first one to terminate, then it would proceed. This would go on forever. Right? This is the problem. Our server is right now not concurrent at all. So in the next one, we would be understanding how to make it concurrent.

Sneha Mehra (00:19:01)  
Without multi-threading, we would now start looking at into IO multiplexing so that we know how event loops are built, how single-threaded Redis is made faster. We will take a look at a bunch of system calls that help us achieve that, why they exist, what they do, and how they do. And in the next one, we would be implementing our own event loop, which would make our Redis server concurrent. So that is it for this one. See you in the next one. Thanks.

—---------------------------  
9  
Sneha Mehra (00:00:00)  
So in the previous video, we made our Redis implementation concurrent, which means that now we are able to connect multiple clients concurrently to our own server. ⁓ So now in this one, what we would be doing is we would be implementing three of the most critical commands, get, set, and TTL. So before we jump into our implementation, let's first see how a normal Redis server reacts to get set and understand what kind of edge cases that we would need to handle. ⁓

On the top right, have my ⁓ Redis server running on port 6379 ⁓ and on the client what I would do is I would be connecting Redis CLI on port ⁓ 6379\. ⁓ And now when I run it, now what I can do is I can set a particular key. I can pass a key k with value v ⁓ and it will be setting the key k with value v and if I do get on key k, I will get a value v. ⁓ Apart from just setting a key and a value,

Now if you, but before you jump that, if you clearly see, ⁓ if I do a get key or get key k, what I'm getting is I'm getting a string v, which means it's a bulk string, right? Now in the set, you would also see a bunch of parameters that I can take. The parameter that we are interested in is ex. Ex is basically setting an expiry to it. So you can specify that, hey, I would want to, ⁓ I would like, I'm basically setting this key with value v.

And now what I want to do is I want to also set an expiry to that. ⁓ Let's say if I set 5, which means after 5 seconds, the key should be deleted. So if I do a set k comma v expiry of 5, if I do get k, it would give me the value v. But if I do it repetitively, as soon as 5 seconds I will have, you see it returned a nil. Right? So as soon as key is expired, if I pass in or if I do a get on an expired key, it gives me

⁓ It basically gives me nil right okay, so in set what if I don't pass an integer value? What if I pass some random string? It says that value is not an integer or is out of range so basic syntax check and then if I let's say don't pass anything it gives me syntax error So these are basic checks that we would have to apply when we are implementing set right okay now when we do get what do we do what do we get we get

Sneha Mehra (00:02:24)  
key k but key k is expired so now we didn't get anything so when I am getting a key that does not exist it returns me null. Right? Okay. But if I am doing a set set ⁓ key k with value v ⁓ and if I do a get I am getting ⁓ if I do a get on key k I will get the value v. Right? But what if I want to see what the expiration time is like for example if I set up value if say I set up key k with value v

and expiration of 10 seconds and I want to see the expiration the time to live remaining of that key there is a command called TTL it returns an integer that decrements like that shows the amount of time after which it would be elapsed and you saw it returned 5, 3, 2, ⁓ 1 and then it returned minus 2 so what TTL command does is TTL command responds with the amount of seconds or the number of seconds left

for key to get expired. ⁓ That is its job. So, if I am doing a TTL on a key that does not exist, it returns me minus 2\. But what if I do a TTL ⁓ on a key k1 whose expiration is not set? If I do a TTL on key k1, so key exists, but there is no TTL set on it, it returns a minus 1\.

So we see three typical return values of TTL. The first one is minus two, which means if I do a TTL on a key that does not exist, we should return minus two. ⁓ If I do a TTL on a key whose expiration is not set, it returns minus one and otherwise it returns the time that is elapsed or sorry, the time that is remaining for the key to get expired, to get auto expired. So now let's implement get ⁓ set and TTL in our own implementation. Okay, so now what we would do.

⁓ is we would go through and now we would do an exhaustive code walkthrough on how it is actually implemented. So first of all, what we would need to do is we would need to go to the eval.go file in which we write the implementation of all the command and add support for set, get and TTL. Similar to how we did it with pink, now we do it with set, get and TTL. All of them takes arguments as an input and a writer which could be a socket connection or a file descriptor.

Sneha Mehra (00:04:41)  
and we would be writing the code for that. So first let's start with the simple set implementation. So now set, what would set do? Set as a command would get in arguments, key, value and any other extra arguments that we pass in. And if you remember in the second, third video, I talk about that when your Redis CLI sends a command to your Redis server, it is always sent as a string, as an array of strings, right? So this is what we are getting in the argument. So here what we would know is,

In case my number of arguments are ⁓ less than equal to 2, which means it is definitely not having the key or the value, we would want to send an error. So that is the first check that we are making that if the number of arguments are less than equal to 1, so then it is definitely that ⁓ we are not passing the required arguments. And then we define variable key value and expiration. The only thing that we are implementing is basic expiration in seconds. ⁓

Now what? We know that argument 0 will be the key and the argument 1 will be the value. We know this part. Right? So that is what we are setting over here. Key is argument 0, value is argument 1\. And then what we do is for now, although we are just implementing expiration, but set also implements a lot of other functionality or lot of other optional arguments. So I'm just writing this loop so that in future we would be able to accommodate those requirements. Right? But

The idea is for now we just have to support basic expiration which starts with EX, a capital EX or a small EX. This is what we would want to start with. So now, given that this is an optional argument, it is very much possible that some people would just pass set key and value. They would not pass in expiration. Which means that our loop when it starts with 2, it is very much possible that your length of RS is itself 2 and this loop would not even execute. So that's fine. That is where we are having a default value.

And as we saw, the default value of expiration ⁓ is minus 1 because I setting something and if I not passing any expiration it is set to never which means that your key is never expired so that's why I setting the value to minus 1 which is a special value that denotes that this key should never be expired. Okay, now what I doing ⁓ is I got the key and the value. Now everything else are optional arguments. So to extract those

Sneha Mehra (00:07:06)  
I'm iterating through all the arcs starting from the index 2\. So 0 and 1 is key and value and starting from 2, I'm checking if it is X or something X or basically something else. So if ⁓ I put a switch case, if it's EX or capital EX or small ex, which means that user has passed in some expiry while firing that command. So with expiry, what I would want to do is I would want to move forward and whatever is passed next to expiring is the time, is the duration after which the key should be expired.

That's a TTL. So I'm doing an I++ and I'm just making an extra check in case user has not passed that extra argument like after ex if user has not passed anything which means that that's a syntax error. We saw it in the edge case and if there is anything whatever user is passing it needs to be an integer a 64 bit integer if it is not an integer which means that that value is not an integer it is out of range which is what we saw when we passed in string in the argument so

expiration and some random string when we passed it. This is that extra check. And ⁓ if it is, if all's good and we, and the user with expiration passed a valid integer, we would get x duration seconds, which means that, that the duration after which the expiry should happen in seconds. So if user sends in five, we would be storing or we would be getting five over here. And then what do we do is we define a variable because for us,

when we are creating or we ⁓ are storing this key value in our hash table where we would be storing it, would want to store, we want to associate an expiry with that. And now this expiry, the granularity that we would operate in is a millisecond granularity. So ⁓ even if we get in in second, we would store it in milliseconds. So that is why I'm just doing this duration, multiplying by 1000, and I'm just storing this and I'm just preparing this duration to be that.

⁓ So, this is the expiration duration. So, if user passed in 5, I will be getting 5000 in expiration duration milliseconds. ⁓ And if no ex is passed, this would not execute and all good. ⁓ And in case user passes an argument which we don't support yet, we don't support px or nx up until now. So, which means that we would be retaining syntax error. ⁓ So, now that after this thing, we would have key and value definitely set.

Sneha Mehra (00:09:28)  
and x duration millisecond could be anything or rather it is optional. The default value is minus one or it could be anything that user passes as a valid integer. And then what we are doing is we are creating a new object. Now this is where we would have to do what? We would have to create a custom object, a custom class sort of thing, a custom struct. So the role of that would be, so I created a file called store.go. This store file is responsible for holding

anything and everything in our Redis implementation. Now this store.go file is basically what our Redis is simple key value store to hold key value the best data structure is a hash table. So we are just using a map of string to star object. Now this object is something that we have defined. What this object says this object holds two things for now. The first is a value which is an interface which means we can put in literally any value any data structure in this and it would work just fine. So because

Object within which you are storing value. This could be anything That is why we have given a type as interface and the second parameter is expires at now this expires at is not storing the duration But it would store the absolute time at which our object ⁓ will be expired So for example if right now the time is 5 p.m Some epoch millisecond if I take it and if I pass in 5 second as my TTL

So I would be storing that absolute time at which it needs to be expired. ⁓ Rather than storing just five, I would be storing the absolute time. So basically five seconds past five is what I would be storing over here as part of epoch melees. ⁓ Okay, so when this file loads, I'm just creating a map over here, which we are calling it as a store. And the new object, this new object is what this function is doing is, this function is creating a new object.

setting with all the default values that we would want and it will be returning a reference to that object. So it would accept a value and a duration and an expiration duration. If tomorrow we add more functionalities to that, we can just pass it over here. ⁓ The plan is that because we would want to store an absolute expiration time rather than duration, we don't want to repeat the same logic at hundreds of other places. So that's why this function call would help us keep this one logic at this one place only.

Sneha Mehra (00:11:50)  
So, what we would do is in case we are given duration milliseconds over here, we are calculating the absolute expiry at. So, absolute expires at starts at minus 1\. In case some duration millisecond is passed, then we are taking time dot now dot unix millis plus duration milliseconds. Because from that function also we will get it in millisecond, here also we are taking it in millisecond, we are just adding it. So, we would get an absolute expiration time for this particular value. Right?

And we are just creating a new object and we are returning a reference of it. This is what the new object function would do. ⁓ And apart from that, what's the next step? Now that we have created a new object, we are invoking put on it. This put is also part of the ⁓ store.go file whose job is to ⁓ add it to this hash map. Now I've created a separate function for this because tomorrow we might have multiple hash maps rather than just one due to some implementation if two.

like in order to support a specific kind of implementation. So that's why just to keep things extensible for the future or just be future ready on that. What we are doing over here is we are primarily storing the put function is just an abstract function whose job is to take a key ⁓ and an object, not a value, but an object and then make an entry in the hash map in the hash table that we have. Store of k equal to object. That's it. And the second is get where given a key it returns a reference to this object.

and it just returns store of k. If the value does not exist or, sorry, if the key does not exist in the hash map, it will return a nil, which is exactly what we want. Right? This is how our store.go file would look like. So now, coming back to the set implementation, we created an object, we created an object and we put it in the hash, in the hash table, in the hash map that we have. For a key, for this object, make an entry. And when the entry is done, we are returning to the client. What?

and OK reply. OK reply is a simple string. So it is like here it is already res encoded plus OK slash R slash N. So without ⁓ redoing the ⁓ OK thing or without redoing the encoding again, I'm just hard coding it Z send plus OK slash R slash N and done. So this is how our set implementation would look like. ⁓ Okay. So we talked about set. Now let's talk about get. What do we get? So in get,

Sneha Mehra (00:14:14)  
the argument there will there has to be exactly one argument which is the key. If it is no argument is given or more than one argument is given that is an error. So that is what we check over here. And then the first argument that we are given is the key. Then what do we do is we invoke get on our store to get the key we would get an object reference. ⁓ If the object is nil we return a response or a RESP nil which is a

ready serialization protocol encoded nil. Now this nil response, if you remember, ⁓ when we were discussing ⁓ RESP specification, a nil is nothing but a string with minus one length. This is what we are doing so that we don't have to write this nil again and again. We are just created a constant object out of it and we are just referencing it everywhere. So if the object does not exist in get, we would return nil, which is what we are doing over here. And in case the object ⁓ exists,

in the hash map or in the hash table and its expiration is set. So which means if object.expires at a something, it is not minus one, some expiration is set. I check if object.expires is less than time.now.millisecond, which means if the object exists in the hash map, but it is expired, I would return a nil again. Because if the object ⁓ is already expired, it does not make sense for me to return it. Right? Because expired objects,

for your end user, for your client, an expired key is same as deleted key. Right? So that is where what we are doing is we are returning nil over here. In case there is no expiration set or in case your expiration is yet to reach, that's where what we are doing is we are ⁓ returning the value, but in an RESP encoded form. Right? So whatever the value is, if it's string, we would return string. In most cases, it would be string. That's what it handles. So we are taking the value, we are encoding it, and we are sending it back.

⁓ And this is what our get implementation would look like. ⁓ Okay. And now the final implementation, the TTL. So what TTL does is TTL returns the time to live of an object. So from the current time, how many seconds are left for the key to be expired? So it also expects exactly one argument, which is the key. If that key, if that argument is not given, that's in syntax error. ⁓ And my bad, instead of get, this has to be TTL. Okay. For the TTL command.

Sneha Mehra (00:16:41)  
and key we are setting args of 0 to be the key ⁓ and I am doing a normal get on the store we will get an object reference if this object reference is nil which means we have to return minus 2 remember if object does not exist or sorry the key does not exist and we file ttl on that we got minus 2 ⁓ if the ⁓ key exists but its expiration is not set then we return minus 1 this is what we are doing if object not expired double equal to minus 1 we would return minus 1 over here

⁓ And because it is an integer, it starts with a colon minus 1 slash r slash n, an RESP encoded integer. And in case some expiration is set, then what do we do is we compute the TTL, we compute the duration left. The duration left in millisecond would be object.expiresat, because here we are storing everything in millisecond, minus my current time, my current epoch millisecond. The difference would be the number of milliseconds left.

And then if it is less than zero, this is the final check that we are making. In case I am getting an object which is already expired, which means yet to be deleted, but expired, if duration is less than zero, I am returning minus two. Minus two implies that ⁓ key does not exist, which is exactly what we want. Because the object is already expired while we were doing this. If object is expired, return minus two. And in case there is enough duration left, then what do we do?

is we convert this millisecond into second, convert it into an integer and encode it. And when we encode this, we are doing a normal RESP encoded integer over here and returning the response. This way, when our client invokes a TTL command to our server for a particular key, it would get and return the number of seconds left for the key to be expired.

And just a hunch, just the final file that we touched upon is the encode. In encode up until now, we only implemented string type, but now we implement one more type, int64. Now for int64, we know what RASP encoding of an integer looks like. RASP encoding of an integer starts with a colon, followed by a string representation of an integer. So for example, if I'm storing value 1000, it will show 100, and then followed by slash r slash n. So which is exactly how I'm returning, I'm encoding it.

Sneha Mehra (00:19:02)  
and converting it into bytes and returning. Now this is exactly what we would be returning as part of the response, which is where on the Redis CLI, you would see that we are indeed getting an integer minus two in the response. And these are all the files that we would have to change and this is how the code would be structured for get, set and TTL. Let's just quickly go through and see a quick demonstration of it. So I'm running my own server on gorun main.go. ⁓

It is running on port 7379\. through redis cli. slash redis cli minus p 7379\. ⁓ So here ⁓ our redis cli is now connected to our server. Now if I do ⁓ set space k comma v, it's written ok, a simple string. If I do get k, ⁓ it returns me v in double quotes, which is a bulk string. Now if I do ttl of k, because expiration is not set, we would get minus 1\.

Now let's say I do set k comma v with expiration of 10 and if I do syntax error, ⁓ great catch, ⁓ ex and then 10\. ⁓ So which means that we are setting a key k with value v expiration 10\. If I do get k, we would get this thing, the value v and if I do ttl k, ⁓ we just missed. Okay, let me just do it again. Set k and expiration to 10\. Let me check ttl, nine, ⁓ seven, ⁓ six.

4, ⁓ I am just waiting for a random time ⁓ 1, 0 and then minus 2, minus 2 because a key is expired is same as key does not exist so we would be returning minus 2\. It's not like ever decreasing beyond that because key does not exist anymore we are returning minus 2\. Right and this is exactly what how Redis CLI or Redis Client also does. We have just re-implemented get set and TTL up until now. Right. Okay.

So that is it for this one, but there is a great thing to be implemented. What would we implement next? Now here, if you see here, we added a lot of stuff, but here, although we are telling our client that key does not exist or key is expired, which was because of which we are returning minus two. If you look at closely, we have not implemented delete yet. ⁓ although we say that key is expired, but we are never explicitly deleting a key. So how would that happen? So in the next video.

Sneha Mehra (00:21:31)  
And in the next part of this thing, what we would be doing is we would be looking at how to expire a key, how keys are expired in Redis, how we would write. So we would be implementing delete command and we would be implementing key eviction. Right? So that is the agenda or that's going to be the agenda for the next video. So I hope you found this amusing. I would highly, highly, highly encourage you to implement this. This entire source code is present at github.com slash dice db slash dice. Check out the fifth commit from the beginning.

to see this exact changes in place. ⁓ So yeah, that is it for this one. I hope you liked it and I'll see you in the next one. Thanks a lot.

—------------------------------

25

Sneha Mehra (00:00:00)  
So strings is the most basic data type that any database needs to support. We set a value as a string, we can put strings in the list, we can do a lot of operations on string and Redis. ⁓ So how does ⁓ Redis store strings? So let me start with writing a simple set command in which I'm doing set k ⁓ a. So which means I'm setting the key k with value a. Now if I check the object encoding of that, debug object k. ⁓

we get something called as EMBSTR. EMBSTR is the embedded string. Right? So if I do set k to a very large value, very large, we'll come to this then what's the limit over here? If I set a very large value, if I do debug object, I get raw. And if I do set k to 10, ⁓ I get encoding as it. So which means that depending on what we are storing, how long it is, it is doing something around this and deciding the encoding of it.

Right, so let's quickly take a look at the source code and see what is happening behind the string. So here we start with because we are playing with set command. Let us start, starting point should be the set command that we have. So ⁓ if I do set command, if I look for that, I'll find something interesting. So in the t, t underscore string file, which is the type string file, you can find the implementation of the set command. It got the object and does something, something, something.

One thing that we are interested in is try object encoding. And then there is this set generic command. So what is this try object encoding? Because it is changing the encoding. So this is the function that needs to have something. So if I go through this function, what it is doing, it is first setting this value, like long, it is declaring this long value. ⁓ SDS of S equal to O arrow PTR. Here it is creating. So SDS is the simple dynamic string type that we're talking. It's like the way

Redis implement string, it does not use native string, but it uses SDS, which is much more optimized. Then size T is length. Then here we see something happening. So length of SDS length of S, which means whatever the length of the input is, let us evaluate what that it is. If the length is less than 20 and string to L, string to L means string to, it converts string to long. If it is possible to convert string to long, then we do something, which means here.

Sneha Mehra (00:02:18)  
this if would execute if the value that I given is ⁓ a long value, it's an integer value. ⁓ So if I leave this if, if we come to the else part, here we see they are setting ⁓ object encoding to int. So if it is integerable, like if I may put it, it is setting the encoding as int and setting the pointer to value. ⁓ So the value ⁓ in the radius object, when we have the pointer, it points to something. This is where it is pointing to

the value that we are setting in over here. So value was declared over here. You would say, but where did the user set like, where did the code set it? It's set over here. It is passed by reference and this string to length, ⁓ sorry string too long is actually updating it here in this part. It is not returning it, but rather it is updating it on that same location. So these are integers are stored there. Now there is another thing, which is if it is here,

If length is less than object encoded EMBSTR size limit. So this is what the size limit we are talking about. So when I did set KA, it took encoding as EMBSTR. And then when I did a very long string, it went beyond that and it used raw there. ⁓ So if my length is less than equal to the limit that I have, if it is less than the limit that I have, I can create an embedded string object over here.

Otherwise, create a general string object. ⁓ So, which is a raw. So, whatever bytes we are given, we are simply storing it. So, it's not a string string. It's just a normal set of bytes that we are storing in the array. Right? So, this is where the changes are happening. So, now let's see this in action. Like what, like how exactly Redis stores strings efficiently. Okay. So, we'll start with the normal theory part where we

Take a look at how it is actually doing it. ⁓ Sorry. So strings in Redis are implemented using SDS, which is called Simple Dynamic String. And why do they have to write their own implementation? Because number one, ⁓ recall your C. When you want to compute the length of the string, you should literally iterate byte by byte until you hit slash zero. And then you do count plus plus. And that was the length that you had. So you don't want

Sneha Mehra (00:04:37)  
to do that because computing the length again and again, you would do complete iteration again and again, that's very slow. You don't want to do that. Then efficient, then appends are efficient when we write our own implementation, we can do it because we can allocate a larger chunk of memory and then append it would be much efficient versus doing something else. Third, ⁓ our implementation would be binary safe. So what do we mean by binary safe? It means that if you use a general string, the string ends when you see a null character.

⁓ Sorry, when you see a null character. So what if my binary object that I'm trying to store holds the null character? It would terminate at that point unnecessarily. So we want to keep things binary same, which is why we are writing our own, or rather Redis wrote their own implementation called Simple Dynamic Stream. So here ⁓ Redis has five types of layout, just like any other type, it has layouts. Now, first layout is ⁓ SDS HDR5, HDR is header.

So simple dynamic string header five, simple dynamic string header eight, and there are other like simple dynamic string header 16, 32, and 64\. You could have very clearly guessed what this is all about, but still, there is one very interesting optimization that I would want to talk about here. So if I go back to the source code again, and I check for the header in sds.h, if I come to this file, here you see the headers defined.

SDS HDR5, SDS HDR8, SDS HDR16 and so on and so forth. So here you see that if I'm creating a small string, five bits are enough. So which is where what you do ⁓ is you have two attributes in this structure. First one is unsigned char flags. Char means eight bits and unsigned char flags, it means it would be ⁓ storing something, information about encoding of my string and which type I have chosen and whatnot.

So what it does is it uses the three least significant bits for type as in what type of string it is or what kind of information that we are storing and five most significant bits for the string length. And now because of this five most significant bit, it is called SDSHDR5. ⁓ So which means that because it is five ⁓ bits, the maximum length that we can go for is 2 raised to 5 which is 32\.

Sneha Mehra (00:06:56)  
⁓ So that is where it would be capping at. So if your strings are small, can use SDS HDR. Otherwise you can go for SDS HDR 8 in which you have an explicit field for length, you have explicit field for allocation and flags and then buffer. So now length ⁓ is used to store the length of the string. So when you are computing SDS length or you want to compute the length of the string, you can directly use this variable. You will get it in order one because it's storing, it's keeping track of the length of the string.

⁓ Then alloc, it means that you may allocate a large chunk of memory out of that you are using just a small one for string. So allocation would be the total size that you have allocated for it. Flags would be last three bytes for type and five unused bits there. And then this is what I want to highlight is care buff. Care buff blank array. ⁓ Keep this in mind. This is a great place to save space. Right? Okay. Let me jump to the part.

Theory where we were ⁓ taking a look at what was happening behind the scenes. Okay, so we spoke about ⁓ The SDS HDR5, SDS HDR8 and what not right? So here the significance of this char buff ⁓ blank array the significance of this is it does not occupy any space You would have thought that hey if I'm storing a string, let me just use a char star Whatever string size is there. I'll use that

Remember, char star is a pointer. It would require bytes to ⁓ use. Like it would ⁓ take up some bytes in the structure. But when you use char buff blank array, it does not hold anything. ⁓ It just, whatever you allocate, it just uses the last part of it as buffer array. Right? So that's the beauty of this. So when you do, this is not occupying any space at all. But you can store as much because it would be interpreting that much.

as a buffer array, which is where you would be storing the string into. That's a very, very, very interesting and important optimization because this would lead to a lot of critical decisions. Okay. So here, what we know is we can store all kinds of data in SDS, the simple dynamic string, because it has nothing to do with anything. It just stores bytes over there, right? So the three encodings that it supports is raw, emb, str and int. Let's take a look at what emb, str and raw are.

Sneha Mehra (00:09:22)  
So EMBSTR is embedded string is a string whose length is less than equal to 44 bytes. So ⁓ when we did set KA, at that point of time it used the encoding EMBSTR and when it used that particular encoding, what it did is it did some different kind of optimization. And then when we added a prolonged string, a large number of A's, it used raw. So the golden rule over here is,

If the length of the string is less than or to 44 characters, it uses EMBSTR. Now you would have guessed by the name embedded, it means it has to do something with embedding it into some place. Otherwise it would be mallowed. That's exactly what it is about. So remember the Redis object structure that we had? A Redis object structure has four bits of type, four bits of encoding, 24 bits of LRU, right? So in all these are four bytes. Then a reference counter.

Reference counter is integer which means 4 byte. A void star pointer is 8 byte. So now if you sum this up, so this is 4 plus 8, 12 plus 4, 16 bytes. ⁓ Then because we are storing string, we would be storing an SDS header. Now SDS header would be how much? It would be how big? SDS header, ⁓ if you talk about 8, SDS header 8, it would have a Uint8 length.

uint 8 alloc which means 8 bits over here, 8 bits over here, so 1 byte 1 byte. Then unsigned char flag is 1 byte. So this is 3\. So this was 16 plus 3\. Now remember char buff does not take up any space. So 16, so 64 minus 16 minus 3 is 45\. The embedded string limit is 44 which means your string length is 44 plus 1 for null character. Which would place your string right there and then over here without you needing to do anything else.

So for short strings it does not have to do malloc for another thing, it can put it at the same Redis object directly. ⁓ So this is why the embedded string length is 44\. ⁓ And because this is 64, now you would say why are you selecting it 64 bytes? ⁓ Because the way Redis does memory management, the least amount of memory that it allocates is 64 bytes for anything.

Sneha Mehra (00:11:38)  
So this is where it is leveraging it, that hey, because I'm anyway allocating 64 bytes, let me just use embedded string for that and use the extra space that I have to do something with this. ⁓ Right? Okay. So this is how embedded and raw are stored. Now raw would be storing a pointer rather than storing it within that object. Right? Then ⁓ embedded done, done. The third one is integer encoding. Now integer encoding is if string is representable as integer.

Then it converts and holds it into pointer. We just took a look where we took string to L and we passed in that thing. So if it was convertible, then we set the encoding to end and then we changed it. Right. And then we basically used it to change the encoding of it to integer. Right. That same thing we are talking about. So if it's an integer, it's still strong. It's still stored into the reddish object, but then it is stored into void pointer. This void pointer here. And it's basically typecussed the value and all, and it is stored over there.

⁓ And what we did in first few videos when we implemented this, like 6th 7th video when we implemented this, we implemented integer and we always serialized it into the string and persisted it rather than storing it in pointer. ⁓ That was one thing that we missed. So what I've done is I've created a GitHub issue on that. I would also put that link in the description down below if you're interested to implement the fix up. If you still see.

the issue open, would highly encourage you to fix that. It's a very small fix, but you would have contributed to this DiceDB. ⁓ So do check that out, very simple implementation, but this would give you ⁓ a way to think about it and basically make changes to the actual database. ⁓ Then the final thing about raw encoding. Now this is what makes Redis very extra special because now here we are not limiting to anything. We can literally store a set of bytes in this.

It does not care if it's a string or not because it's binary save it does not have to terminate with slash zero, right? ⁓ So raw in raw, you can store literally any ⁓ strange of bytes within that. So you can use it to store large string. You can do use it to store anything. For example, you want to store bloom filter within that hyperlogal log within that. You want to serialize and store sets and lists and whatnot. You can do that. You can write your own implementation that uses that to allocate it. And that's why you would see

Sneha Mehra (00:14:03)  
Throughout the Redis source code, you'll see a lot of instances of SDS because everything that gets allocated as a raw bytes if you want to allocate, it all happens through SDS. And you'll see a ton of, ton of, ton of places where you see SDS referred in the entire Redis's code base. Right? So yeah, this is how Redis implements string. Here you see how typical design choices that they made where they started storing the length of the string. This way we made

⁓ string length access lightning fast. ⁓ are writing our own implementation because we wanted, ⁓ rather Redis wrote their own implementation because they wanted appends to be efficient, which we'll get. We want to access string length in order one, which we are doing, and we want binaries so that we can store any kind of data within that. So if you'd want to write bloom filters for Redis, there is a community plugin for that. If you want to like hyperlog log, ⁓ all of that is pretty much done because you can just dump a gigantic byte array into whatever you want.

And that's the beauty of writing generic, like supporting raw type in which you are literally storing literally anything, right? So yeah, that is all about Redis strings internal. I would highly, highly, highly encourage you to go through the source code. The file is sdh.c and sdh.h. You will see a ton of implementations of common string functions there and a lot of optimizations throughout, right? And again, ⁓ I've created a GitHub issue where you can use integer encoding and change and instead of doing strconv every time,

we would want to store it as a pointer reference, right? So that we don't do serialization, deserialization again and again. So if you find that as an open issue, I would highly, highly, highly encourage you to go and contribute back to this, right? But nonetheless, ⁓ I hope you found it interesting. I hope you found it amazing. You see how it opens up door for so many things, right? So yeah, that is it. That is it for this one. I'll see you in the next one. Thanks, Atal.

—------------------------  
15  
Sneha Mehra (00:00:00)  
So in this one, let's circle back to an eviction strategy called all keys random. Last time we implemented simple first eviction strategy. The idea was to evict the first key that we can iterate to. ⁓ But now we take it to the next level, all keys random. The idea is pretty simple. Out of all the keys that exist in the RedisDB or in our database, we would be picking keys at random and we would be evicting it. ⁓ So let's take a quick look at implementation for that. So here I've changed the eviction strategy to all keys random.

which will be implementing. ⁓ But apart from that, I'm setting my keys limit to 100\. We are still relying on the maximum number of keys as our threshold that our database at a max supports 100 keys. ⁓ Apart from that, and this is mostly for demonstration in production, this would not happen like this, but this is for demonstration purpose. Then I'm defining a thing called eviction ratio. This means that it would dictate that whenever my eviction is triggered, ⁓ how many keys would I be evicting?

This is the idea behind it. How many keys would I be evicting whenever my eviction would be kicking in? So 0.4 implies that I would be evicting 40 % of the keys. ⁓ Although that's not a production-grade thing, you would be evicting only few keys and not 40 % of it. But this is just for demonstration purpose so that we can see our eviction algorithm working. ⁓ Now let's take a look at the eviction strategy. We'll go to eviction.go file and see how we can implement this. So first of all,

What we do is we look ⁓ at evict all keys at random, which is what we have added in our switch statement as well. All keys random, evict all keys random over here. And what it does ⁓ is it first identifies the number of keys that it needs to evict, which is the total number of keys that it has, the maximum limit and the eviction ratio. Eviction ratio is 0.4, so 40 % of the keys limit is the number of keys that we would want to evict. And we iterate through the hash table.

Once we iterate through that, we would be deleting every count number of keys from that by invoking delete on it. Right? As simple as this. Now here what we relying on is we are relying on the randomness of the iteration of it. It's not entirely random, but the idea is that the way our hash table or Golang's hash table is adjusted, whenever we are inserting a particular key, it is passed through the hash function depending on which it would get any one of the slot. Now,

Sneha Mehra (00:02:22)  
given that we don't know what the output of the hash would be, we can treat this as a fairly random distribution. ⁓ So this is how we would be implementing the all keys random strategy. Here we are relying on the randomness of the hash function. ⁓ But depending on your implementation, you can implement it however you want, but this is the easiest one to implement. ⁓ And this is how you support that as simple as this. But now, ⁓ how do we know that ⁓ our

eviction is really working, which means that when we are doing a large number of key sets, ⁓ like we are putting a lot of keys in our database, what needs to happen is that let's say, ⁓ given that we have set a key's limit to 100, how do we know that we are not breaching that? So that is where any and every database supports statistics. ⁓ So does Redis. ⁓ So let's take a look at what stats are all about in Redis.

We'll take a very quick example over here. So like always, top right is where we are running our regular Redis server, right? And on the bottom left, what do I have? Is on the bottom left, I have ⁓ our, ⁓ we'll be connecting through our CLI. So . slash Redis CLI is what I'm passing over here. ⁓ And I'm connecting to port 6379, which is normal Redis port. ⁓ I've connected to it.

⁓ So, let's say if I do set k1 ⁓ v1 ⁓ a key set. And now if I fire info, ⁓ you get some output. Now this info is the redis command that when you fire it gives you the statistics about the database at that given instant. ⁓ So, if I fire info right now, what is the info about my database at this moment is what is outputted over here. The section that we are interested in is key space.

So this key space section that we have over here, sorry, so this key space section that we have over here ⁓ is what it holds or it tells the number of keys that are there. ⁓ So now in database zero, the number of keys are one, expire zero, average CTL zero. We don't care about expires and average CTL, we are interested in number of keys. ⁓ This is what now we would be implementing. ⁓ So we want to see how the stats look like. We want to support the stats so that we can see

Sneha Mehra (00:04:46)  
if we are ever like our key eviction in action, right? So before we do that, let's take a quick look at what stats are all about and what's the response format of it. So stats are extremely crucial for database. helps us ⁓ monitor our database really well, set up alerts, set up like give, it gives us transparency. It gives us confidence that our database is working fine, right? And we can use info command to get the statistics we just saw as an example. Now, what does info command returns?

info command returns information about the database. Information is classified into multiple sections. Each section starts with a section title, which starts with hash space section title. So for example, what we are interested in is key space. So hash space key space is what the section title would be, then a new line, then you would get key and value followed by slash r slash n.

⁓ So, for key space it would be hash space key space db0 colon some value. Now, here value itself is key is equal to 4\. Here your key is db0 and the value is key is equal to 4\. ⁓ Here your key is db1 and value itself is key is equal to 7\. ⁓ So, you can have as much as you want. The idea here is pretty simple. Your keys are important.

And values is like depending on the format, you would be writing the parser to understand it. ⁓ And then you had other metrics like server metrics, like what version it is running, what's the memory consumption of it, what's the cache miss ratio and whatnot. All of that would be visible there. All arranged in key value, key value, key colon value, key colon value followed by slash r slash n. This is it is done. But now when you'd want to visualize it, you would want to, given that Redis gives us

point in time statistics. So when you fire an info command, you'll get output of that. You'll not get historical output. So that is where you need a way to see the historical trends of it. So how do you do that? That is where tools like Grafana comes in. Grafana is a visualization tool, right? So Grafana is a visualization tool. You can plot very fancy charts on Grafana and see how your database fared over time. We have a RedisDB running. RedisDB supports info command that gives us point in time response.

Sneha Mehra (00:07:05)  
So what do we want? We want something that gets this information from Redis persisted some database which Grafana can use to render it. ⁓ That is why we are using Prometheus. Prometheus is a time series DB that pulls the data out from any database and it puts it in its time series DB format which Grafana can understand and plot the graph. ⁓ So this is where but Prometheus cannot support all the DB. So that's where you have plug-ins. That's where you have exporters. ⁓

So we'll use a regular Redis exporter, nothing fancy. We'll use existing Prometheus, we'll use existing Redis exporter because we are building Redis compliant database. It was much easier for us. So RDB implementation would be implementing info command that exports the information in the exact same format that your Redis exporter like that your normal Redis DB exports in, which means that Redis exporter would not even know.

that is it talking to diceDB or redis. It would mean exactly the same thing for it. So redis exporter would read, would continuously fire info and would be collecting this metric and exposing it on an endpoint. Prometheus will read this input and persist it in its time series db and Grafana will use it for visualization. ⁓ So this is what we would be mimicking on the local so that we can see how our number of keys is changing over time. ⁓

So this is just about visualization, but when we do that, we'll see the importance of making a DB, a DiceDB compliant to Redis, that we are not have to re-implement anything, we're just using existing toolings to just see our implementation in action, ⁓ Okay, so let's take a look at the implementation part of it. So here what we do ⁓ is we implement the part where we just looked at eviction, I'll close the eviction. Now we'll start with stats, because for stats we are implementing info command. So let me take you through.

the eval file where we are defining all the commands. So if I go below, you can see three commands being implemented. The main one is info. What info command does ⁓ is we ⁓ saw the format. What we are interested in is number of keys. Other sections we don't care, right? All we care is key space sections, which is what we would be responding. So hash space, so we created a buffer where then we are writing hash space key space slash r slash n. And then for all the key space stats that I have,

Sneha Mehra (00:09:26)  
I'm just concatenating it, putting it into buffer, DB0 colon keys equal to ⁓ %d, expires zero, average ttl zero. We don't care about expires and average ttl. All we are interested in is keys, right? So that's what we are implementing. And then here we are doing key space of I of keys. Well, what is this? Here, it's just a global object. So for example, I support four databases within my Redis. So in Redis, you can have 16 databases within Redis itself.

By default it goes into DB 0\. ⁓ So DB 0 to DB 15 is what you can have. So here I am just, because we building a small one, am just sealing it at 4\. ⁓ We are having 4 DB. So by default when we are doing gets and sets and everything, by default it goes to DB 0\. ⁓ So an update DB stat is one function that I have exposed whose job is to for which DB, for which metric, what's the value. So for DB 0, I want to set keys. ⁓

to be 1, 2, 3, 4 and so on and so forth. ⁓ This is what we are trying to store. ⁓ So that is what I creating this object for. But where would we do that? That is where our store file comes in. Every time we ⁓ are putting an object, what we would do is we would do key space stat of 0 of keys++. So for the 0th index, because it's a map of string of int, setting the key string to value

plus plus which means if zero becomes one, one becomes two, three, four and so on and so forth. So every time we are putting it this global dictionary gets updated and whenever we are deleting it the same thing gets minus minus. So this way at the global dictionary level we would have the number of keys at that given instant. ⁓ So which means that our info command when it is iterating through that it would get the exact number of keys at that given moment and we can render it.

or we can send it to the client in a bulk string response. ⁓ So let's see this in action. So here, what do I have? I have ⁓ our classic implementation of our terminal where top right normal red is, we don't get about that right now because we have implemented our own info. So on the top left, what do we have? Is we have ⁓ our implementation. So we are doing go run main.go. Now our DB is running on port 7379\.

Sneha Mehra (00:11:50)  
On port 7379, I'll use my Redis CLI to connect to port 7379, which is RDB. And the client got connected. Now here, I do, let's say I first fire info command. Here you will see everything is zero. DB0, DB1, DB2, DB3, number of keys are zero. Let's say now I do set K1 V1. ⁓ As soon as I did that, ⁓ firing info, you can see DB0, keys equal to one, everything else is zero. We have not even implemented that, right?

So it goes to zero, my keys are changing. ⁓ Let's say I do set k2 v1, ⁓ I'll say number of keys is two, ⁓ right? Similarly, k3 v1, number of keys is done, ⁓ right? So now the number of keys are increasing, ⁓ right? So now here, how do we visualize it? So here we can visualize this particular part with our browser.

So here, this is the Grafana instance that I have running in which I'm plotting the metrics for last, ⁓ for basically last 15 minutes. ⁓ So here I have this Grafana metrics running in which you can see the graph. Let me just clear it up so that it looks neater and cleaner. So now you can see that this is refreshed. ⁓ Now you can see the number of keys is three. ⁓ Let me just add more ⁓ keys to that.

⁓ Let me just add more keys to that and we can see that the chart will now shoot up. Now if I do set k4 v1 and then I do set k5 v1 and I do set k6 v1. ⁓ So here if I do this what do I get ⁓ is I get our charts when I refresh it I can see this particular thing shooting up. ⁓ So you can see the number of keys getting increased. So here you can see that graph.

on how the number of keys are there. So I've written a very quick utility where I can bombard this server with a lot of set requests. ⁓ And what we would do is we would be firing a lot of set requests over here so that we can see our eviction strategy in action. So here I've written a storm utility. It's again checked in our code base, normal bulk request that we are firing. Go run storm set main.go.

Sneha Mehra (00:14:05)  
it will continuously keep on firing random keys in our database. And here you can see the number of keys are increasing. You can see the chart and see that the number of keys are increasing. I'm just constantly refreshing it. Now the number of keys have written, have reached ⁓ 100, almost about to reach 100\. The latest value is 71\. Here you can see at the bottom, the latest value is 71\. Now you can see this is increasing and then decreasing again.

The last value is what you are seeing there. The yellow dot is the last value ⁓ which is being the number of keys. Now here you see even as the time increases, my number of keys would never go beyond 100\. But if you see, the yellow dot goes up and then comes down. It goes up and then comes down. Why it is coming down? Because our eviction ratio is 40%. ⁓ So every time our LRU is running,

Every time our eviction is running, it is evicting 40 % of the keys. ⁓ So it would come down, go back up, come back down, go back up and come back down. ⁓ And this is how you would see your eviction, but you would never see it go beyond 100\. ⁓ Now see, the graph is again rising up and the value that it is plotting is once like, I'm refreshing it again and again, you would not see point in time things, but you would see once every second it is registering one data point.

But here you can see the number of keys will never cross 100\. The yellow thing, ⁓ you can clearly see a sawtooth pattern emerging. You are setting a large number of keys, it would increase. Then when it hits 100, it would be evicting 40 % of it. It would drop down to around 60 and then again adding, again dropping, again adding. ⁓ So this is what you would see a sawtooth pattern. But at no stage,

the number of keys in our database will cross 100\. Although we are constantly bombarding it. We are constantly bombarding it. Here you can see on the bottom right, my script is running, is constantly running set some key and some value. Constantly bombarding it. But it has no impact on our DB. Our DB is running just fine, capping it at 100, ⁓ basically 100 keys. Right? ⁓ And this is where you see your eviction strategy in action.

Sneha Mehra (00:16:28)  
We implemented all keys at random with a certain threshold, ⁓ our eviction ratio. And why we choose eviction ratio as 40 %? ⁓ So that we can see this clear pattern. If you reduce your eviction ratio to let's say 10 % or 5%, you would see that just slight swiggles there.

⁓ But with this 40%, it would be sharply declining. The number of keys would sharply decline and then rise up and then decline and then rise up and then decline. So you can tweak your eviction ratio as per your use case. ⁓ But typically Redis also does very small number of evictions, five to 10 keys at max. It does not evict a large number of keys at all. ⁓ But this is where you see this in action. Now here, what have we achieved? What have we achieved is we implemented two things. All keys random eviction algorithm.

second statistics. ⁓ Given that, I didn't have to build any single thing because we are re-implementing Redis. We are using existing ⁓ Redis tooling to visualize this part. ⁓ And how beautiful is this? Because we see, given that ⁓ for us it's understanding, but we being able to reuse existing tooling makes our life so simple. We are using existing Redis CLI itself. We are not have to write explicit CLI for ourselves.

Right? And which is what I would highly, highly, highly, highly encourage you to re-implement it. So easy, so simple and so fun to be honest. Right? ⁓ So yeah, again, the source code, the entire source code can be found on github.com slash diceDB slash dice. Go through the commit, go through ⁓ comment number eight or nine. will like the, like the commit messages are self-explanatory. Go through this commit, see the changes that we have made. We might not have did the best implementation for key space stat, but

it's enough for us to see our eviction strategy in action. And then over time, when a database matures, we can ⁓ make it as complex, as fancy as we would want to. ⁓ But idea remains the same. Understanding is much more important than anything else. And we think we did a very decent job there. So yeah, that is for this one. Now what's next? The next is about approximated LRU algorithm in detail. We would know...

Sneha Mehra (00:18:40)  
like we would learn about how Redis actually does that. A couple of videos back when I was talking about Redis object, we saw LRU bits set there. There was 24 bits for that and that is used for approximation algorithm. So next video we will understand how it does that. Theoretically we'll understand what is exactly happening behind the scenes. And then the video after that we would be implementing it by going through the actual Redis's source code to see

that exact same thing in action and that's such a beautiful piece of code, a very interesting algorithm, very simple to implement, makes LRU super, super, super efficient at scale. ⁓ So yeah, that is it for this one. I hope you found it amusing. You saw how we implemented all keys random eviction strategy. We implemented our own stats, visualized it through Grafana stored in Prometheus so that we can see this nice wavy pattern. So yeah, that is it for this one. I'll see you in the next one. Thanks a ton.

—----------------------------------  
3  
Sneha Mehra (00:00:00)  
So in this video, I'll be giving you a walkthrough of a simple TCP server written in Golang that will be an eco server, which means that whatever message we send to the server, it will be responding back with the exact same message. ⁓ This is a simple, simple Golang based project in which we have a go.mold file having no external dependency, which means we are writing everything from scratch. This asserts that fact. And then ⁓ the execution starts from

classic main.go file in which we have the entry point main function. Now, when we run go ⁓ run main.go, this is where the execution starts. The first line of this says setup flags. Now, setup flags in this, we are setting up two flags. These are the command line flags that we would be giving when we are executing something. So for example, when we start a database, we would want to specify the port at which our database would be listening. So the default value for this would be 7379\.

Basically, Redis is 6379 and DiceDB is 7379\. With this, I am also taking an input for the host, which means from which ⁓ connections or from which IP should I be accepting the request? Should it only be local? I am using 0.0.0.0, which means accept incoming connections ⁓ from anywhere in the internet. ⁓ Okay, when that is done, doing a simple print about rolling the dice because it's DiceDB. And then this is where the execution starts.

So here, this is what we have written where the plan is that I'll be running a synchronous TCP server, which means I'll be starting a server that would accept TCP connections on the specified port. Let's see what it does. So if I open this, we see this particular function. So what it does, first log that just information, ⁓ then ⁓ I'm having a variable called ⁓ con underscore clients. This variable is an integer.

which would hold a number of concurrent clients that are connected at the moment. It's just some extra information that we are collecting so that we can see that, hey, indeed we are having large number of concurrent connections being accepted to our server. Now, this is where ⁓ our actual socket programming starts. Now, the first thing that we are doing is we want our server to listen. So we are calling net.listen. ⁓ We want to listen onto a TCP connection.

Sneha Mehra (00:02:23)  
at corresponding host and corresponding port. So as soon as this is executed, would be, our server would start to listen on a particular port. And which means that any of the client can talk to the server ⁓ on the port that it is listening on. ⁓ Right? Okay. ⁓ Once we, once our server is started, that's where we are running this infinite for loop. This is an infinite for loop.

whose job is that, I'm infinitely waiting for new connections to be accepted to my server. Right? So now ⁓ any client, any client ⁓ will be able to connect to this server. Now, for us to tell that, hey, I'm waiting for a new connection to be accepted, this is where we are making this blocking call called listener.accept. Now, listener was the instance of the server.

that we started and we are saying that, for this over the socket, I'm accepting a connection. So this is a blocking call. So until a client connects to the server, ⁓ my code execution would be waiting over here, ⁓ right? Listener.except. As soon as a client connects to the server on the port, ⁓ the control flow moves forward. And then we would have this. ⁓ This is where we are doing concurrent clients++.

This typically means that, now I have more than like, now I have one, two, three and clients connected. Right? Okay. When that is done, I'm just printing that message. This is an important one where I'm printing the message that, hey, when my client connected to the server, this is the eco that is, or basically this is the server log that is going that, hey, a new client is connected to our server. ⁓ Right? Okay. Now when that is done, I'm adding another infinite for loop within this.

Now the job of this for loop would be that because what Redis would be doing or when we're writing our eco server in any case, we would want our clients to continuously send us commands. For example, put this key, delete this key, get this key. This should be continuously happening, right? So this is what we are doing. We are running an infinite for loop for now in which we would have a client continuously sending us commands. And what would be doing? We would be

Sneha Mehra (00:04:38)  
⁓ responding with the command that was sent. That is simple ECO server. So if we get hello, we would be sending hello. If we get world, we would be sending world. ⁓ Right? Whatever we get as an input, we are sending it back as an output. Right? So this is what this infinite for loop will do. So once we accepted the incoming connection from the client and then this infinite for loop runs, it is continuously reading the message over the socket. Right? Okay. So now what are we doing? We are invoking

or we are written this function called read command. Now, what does this read command does? So read command takes this socket connection ⁓ and basically fires this system call called read. Now this read system call, what it is doing is it is listening over this socket and it is trying to read a message ⁓ over this socket, right? So ⁓ if there is nothing that is coming in from my client, this is a blocking call.

this would block until I get something from the client. So, I am continuously ⁓ listening onto something, onto the incoming messages that is there, and this is a blocking call. So, until there is something, this call, the execution would halt over here, right, for that particular specific client. So, when we read it, we put it in a buffer and we get the number of bytes read, right? So, if there is any error, we return the error. Otherwise, whatever the incoming message we are getting,

we are converting it to the string and we are returning back. So read command is basically making a blocking call, waiting for the client to send us a command and then we are converting it to string and basically sending it forward for our execution. So now as the read command is done, what are we doing? As the read command is done, in case there is any error, which means that in case your client disconnects the connection, in case there is any issue with the socket, your error would not be equal to null.

which means that I want to close my socket connection. I want to reduce my number of concurrent TCP connections that I'm handling. And I'm printing this message that, hey, a client got disconnected. Which client? The client whose remote address is this. And now my concurrent clients are these many. Right? That is what I'm printing. And in case it's the graceful termination where your client is sending an EOF to terminate the connection, I'm simply breaking out of my for loop. So which means that your...

Sneha Mehra (00:07:01)  
your connection would be killed and whatnot. ⁓ Right? Okay. And ⁓ if your error was nil, sorry, if your error was not nil, which means that there was some error, you ⁓ basically terminating the socket connection, allowing someone else to pull that connections off. Right? Okay. So once this is done, ⁓ Which means that there is, ⁓ your error is, like where your error is nil, which means that you got something from the client. Maybe put,

maybe get, maybe something. You got some command from the client. I'm just logging this command on my server, the command that I got. And I am triggering this respond. Now respond is also what we have written in which we are accepting this command ⁓ and the socket connection. And we would be responding something ⁓ as like we would understand the command, process it and then send the response over the socket. For now, because we are doing simple echo, what our respond function would do is given the command, given the socket connection, I'm just

writing it back over the socket. So, whatever we got, we are sending it back to the client. As simple as this. Right? And once that is done, we are done with the response. Right? Because this is what we are trying to build. We building an eco server. Right? Whatever we get from the client, we are sending it back to him. Right? Okay. So, once that is done, this ends our infinite for loop and which is wrapped in the, ⁓ under another for loop.

And this is an extremely simple ECO server that we have written in Golang. Right? Now, if I were to execute this, now let's see what happens if we execute this. Now, just switching things back, if I execute this particular part, if I run gorun main.go, you can see that it is printing, rolling the dice, starting a synchronous TCP server on 00007379\. Right? ⁓ Now, because it is an ECO server, let's say,

What I do is I connect to this TCP server and I send something. I should be getting something back in return, right? So let's say I connect to this TCP server using Netcat. I'll fire localhost ⁓ 7379\. If I fire this, let's say what you get. In the console log, what we are getting is we are getting that client connected with address 127.001 with some ⁓ socket. So this is ephemeral port that it is using.

Sneha Mehra (00:09:24)  
So a client got connected to this TCP server and now my concurrent clients is one. So this asserts the fact that hey, one client got connected. Now let's say we send the server something. ⁓ Let's say I send hello. ⁓ What do I get? So in the server you log you see that I received a command hello, which is what we intended. And in the response, we also got hello back, right? So we send hello, we got hello. This is a simple ECO server, right? So if I send world.

We get hello, we send word, we got word back. ⁓ Okay. Now if I send hello space word, we send it, we got it back. Ready started? Right? Okay. We connected one client. Now let's see what happens when we connect another client over there. So let, so now the concurrent connection should be two. Let me fire that localhost 7379\. So if I do that, now let's see what happens.

No movement on the server. Earlier when the first client got connected, your server said, hey, a new client is connected with some address. This time nothing happened. Why? Because our server is single threaded. We had for loop within for loop. So until your first client disconnects, ⁓ when your first client disconnects, then your second client will get a chance. Right? Now let's see this in action.

If this is true, ⁓ if I close this connection, which means I fire control C, then the other client should be getting something. Right? So let me fire control C. And we saw that as soon as we fire control C, what we got a client got disconnected this, this particular client got disconnected. And now my concurrent clients is zero. And then a client connected with this address, this is the second client that got connected. So if I just fire this, if I send

Now my second client sends the message, hello. Now it will get back hello, right? But now if my first client again tries to connect, now this is blocking because our code is stuck in this infinite for loop where it can only accept one TCP connection at a time. It is not able to accept multiple TCP connections at a time. So what we have just built is we built a simple eco server that accepts one TCP connection at a time, runs an infinite loop,

Sneha Mehra (00:11:51)  
where whatever we get as a command, we can do some processing and we send some response back. Right? This is what we have done. So now if my second server sends, let's say it sends Arpit. ⁓ Right? Nothing happened because your server did not even accept the connection because it's not there in the for loop. Your server is not waiting on accept call. It is stuck in the for loop where it is continuously reading the message or it is trying to read the data over the socket.

but your second client is not sending the data so it is blocked there. So if my second client sends word, I'll get back word. But as soon as the second client drops off, now what happens? Now from the first client, what happened? Your second client terminated, which means your first client got chance to get accepted. The TCP connection is accepted. Your server now says a client with this port connected and now my concurrent client is one.

and the arpith command that we sent from the first client now after reconnecting, it got that command and we responded it back. So this clearly shows that we just built an extremely simple TCP ECO server that is single threaded and can handle just one connection at a time. Okay. So now that we have this, let's have some fun ⁓ and see what happens ⁓ when we connect an

a Redis client to this server, ⁓ right? See, Redis client or the Redis CLI that you have, ⁓ it is also like your Redis server is also a TCP server. And your Redis client can is rather your Redis CLI is connecting to that. It should be a normal TCP connection. ⁓ With this ECO server, we'll get to know what happens when a Redis client connects to our server. So let me fire that. ⁓ SRC, I'll fire.

Redis CLI, Redis CLI on port 7379\. So on this port, this is the port of our server. So now what happens when Redis Client connects to it? Let me just close this server so that I clear the screen and things become clearer for us. Okay. Now I'm connecting Redis Client on port 7379\. ⁓

Sneha Mehra (00:14:12)  
Let's see what happens. So as soon as Redis client connected, some messages were exchanged. So Redis client sends something because we wrote an ECO server, we would exactly know what your client sent when your server received that particular message. ⁓ So when Redis ⁓ CLI connected, it sent message ⁓ star one because this command was already printed. ⁓ We got star one, some new line.

and then ⁓ $7 and then command. What is this? This ⁓ is what your Redis serialization protocol is all about. This is a command that Redis CLI sent to the server when the connection got established. Okay, so now let me do a put K comma V. I'm trying to, so this is typical command that we fire on Redis to put a key and a value. ⁓ Let's see what happens. When I fired this command, your Redis client, I got a message on the server that said star three,

and then new line and then ⁓ $3 and then put and then $1k and then $1v. So this is what your CLI sends to the server and what your Redis server needs to understand. This is called as Redis serialization protocol. And now in the next video, we will understand what this protocol is, which means that when we are starting our Redis server, how would our Redis client

talk to this Redis server because Redis server does not expose HTTP endpoint. The connection or the communication happens over raw TCP connection. So there has to be a serialization protocol format in which your client or your CLI talks to the Redis server and which is what we would be building or we would be understanding in the next video and then implementing it in the next one. Right? So that is it for this one. If you guys found this amusing, I would highly recommend you to code this thing out. I gave an extremely detailed code walkthrough.

The repository is already there, you know, github.com slash diceDB slash dice. Go to the first commit, check the source code out and highly recommend you to implement it on your own. Right? So that's it for this one. I'll see you in the next one. Thank you so much.

