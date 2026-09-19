Sneha Mehra (00:00:00)  
Welcome to Frontend System Design Sensions. Before we get into patterns, tools, and code, I want to set the stage for why the codes exist and why those topics matter for you as a front-end developer. In most of our day-to-day work, we focus on building features, implementing a new form, adding a search box, or tweak a page layout. But as applications grow, the real challenge shift.

Keeping performance consistent at scale, ensuring your UI reflects real time changes, making your code base easy to maintain months or years later. Prepare for nightwork issues, slow APIs, and unexpected user behavior. These are not just coding problems, they are system design problems. The main challenge is that many front end developers keep the component first mindset.

Focusing narrowly on the UI they are building without thinking about data flow from the back end to front end and back affiliate cases instead of just a happy path. Trade-offs between performance, flexibility and speed of delivery. In this course I want to help you shift to front end architecture's mindset, thinking about the entire system, not just the individual components. The course is built around seven essential areas of front end system design.

Each of these addresses common pitfalls and pain points that middle-level developers face when systems get bigger. The first one is data mutation. ⁓ How do you handle changes from the UI to the server? We'll cover patterns for real-time updates, optimistic UI for instant feedback, ⁓ safe rollback strategies, and state synchronization. You will see how to make changes feel faster and reliable for the user. The second one is data fetching.

Fetching isn't just a fetch and render. We will dive into caching strategies, both in memory and persistent, pagination, infinite loading, and request optimization, including deduplication, cancellation, and debuncing. This is how you reduce network load and keep your UI responsive. The third one is rendering strategies. ⁓ Where rendering happens ⁓ matters for both speed and ICEO. We will cover client-side rendering, server-side rendering.

Sneha Mehra (00:02:24)  
static site generation, island architecture and streaming ISSR, with guidance on when each makes sense. The fourth one is data modeling and state management. Poorly structured states lead to duplication, bugs and complexity. We will explore normalization, selectors, memorization and persistence strategies using tools like local storage, indexed DB, Redux and other state management tools.

The fifth one is infrastructure. ⁓ A first front end depends on solid infrastructure. We'll look at CDNs, ⁓ HTTP caching, compression, CI, CD pipelines, and automation. So your builds are always reliable and deployable. Sixth one is cross-functional requirement. They are not just extra work, they are core to a good system. Accessibility, internationalization, security,

observability, ICEO and error handling. All of these affect user experience and business outcomes. The seventh one is performance optimization. We will explore perceived performance techniques like skeleton screens, preload stretches, lazy loading and bundling optimizations. The goal is to make your application feel instant, even with real work is happening in the background. But none of this course you will have

Mental framework for designing front-end that skill, not just in terms of performance, but also in reality and maintainability. Instead of fixing problems after they appear, you will learn to anticipate and prevent them. Instead of building features in isolation, you will see how every decision fits into the bigger picture of your application's architecture. This is a foundation that will make the rest of the codes click.

And it's why front end system design issues is about much more than just writing code.

—--------------  
Sneha Mehra (00:00:00)  
In the previous modules, we discussed several common techniques for performance enhancement. We have done code splitting, perfeting, normalization, pagination, all the typical application level optimizations that improve runtime, efficiency, and reduce bundle size. But sometimes even after all these improvements, users still feel that the application loads slowly. So what is missing here?

Generally speaking, we optimize performance at two levels. ⁓ Application level and infrastructure level. So for application level, this is what happens in your code base. You do the code splitting, preloading, data normalization, virtualization, and so on so forth. ⁓ Well, infrastructure level ⁓ is ⁓ about how your application is delivered to ⁓ users. Things like compression, modification, caching, and using a CDN.

Most foreign end engineers spend the majority of their time optimizing at the application level, but the infrastructure level can often make a much bigger difference, especially for perceived performance, which is what users actually feel. In this lesson, we will focus on one of the most powerful infrastructure level optimization, HTTP caching headers. With just a few configuration changes, you can make your application feel significantly faster without touching.

most of your code. We will go through four essential headers the cache control, e tag, last modified, and vary. Each plays a different role, and together they form the foundation of efficient browser caching. Let's start with cache control, which defines who can cache a response and for how long. It's most important header when it comes to controlling browser and CDN behavior. Here is a real example you might see

You might find in your browser for some of the response you can see things like this cache control public ⁓ max age 300 still well revalidate 60 seconds. Let's break this down. Public means the response can be cached by anyone in a network. The browser shared proxies or CDN nodes. MaxH 300 means the content is considered fresh for 300 seconds or five minutes and still well revalidate.

Sneha Mehra (00:02:25)  
sixty seconds tell the browser it can temporarily serve an expired copy of the data for sixty seconds while it's fetching a new one in the background. So in our board code base, the user profile data doesn't change very often. We can safely cache it for five minutes, for example. So when a user opens a board, the profile appears instantly from the cache. ⁓ So we don't have to do the cache call

server side. And for these directives, here are some typical patterns you will find useful. So for static assets, ⁓ we can make it much longer max age value. ⁓ And ⁓ that means the file ⁓ rarely changed and aversion by their file names normally. So we can safely catch them for a month, for example. So for ATI responses we normally can set the catch control as part of it and we give it a shorter max age.

That means the cache can only be ⁓ stored in browser, not all the proxies or CDNs. So for sensitive data, we'll set the cache control as no store, so we'll never cache authentications or user specific information ⁓ in the browser or any nodes in the network. And for dynamic or frequent changing content, we will set the cache control as no cache, then still using caching but forcing a validation check ⁓ with the server before use the cache that.

data. So no-cache option is ⁓ often confusing people. No cache doesn't mean don't cache. It means you can you can store it but always confirm with the server before show using the data or showing the data. Here it's a very simple example of how to set the cache control header ⁓ to the user API. We set it as private, that means we only ⁓ use the cache or store the data in the browser side, not only proxies

or in the CDN nodes and we will save data for 300 seconds, ⁓ five minutes, basically. Once the browser has a valid cached copy, it serves that content instantly. There will be no network delays, ⁓ no CPU cost. It's what makes the interface feels ⁓ lightning faster. So for example, if you have this cache control public, max H300 still will revalidate ⁓ SXT. So in the timeline at time zero

Sneha Mehra (00:04:51)  
We send all the requests to user two and serve responses and then we will save the cache in the browser. And at 200 seconds the client send request to get the same UIL. The browser will serve the data from the cache. That will be much faster. And at point three hundred fifty, that is still in the three hundred plus sixty range. So the browser will serve the still data. So it will think the data is already

you know outdated but still it can serve it. But at the same time it will fetch data in the background. This is where the HWR means here. And then at four hundred seconds we send request, but because the data has already been refreshed, we will use that data from the cache. So for the board API, I'm setting the ⁓ cache control handler as public and I will save that for a minute and I will do the HWR as ⁓

30 seconds. Add this refresh button temporarily here. So when you click the button, we will send the request to API board one. So if the data is already in the cache, we will just serve it from the browser. Otherwise, after 90 seconds, we will actually fetch the data again. So let's see the in action. So this is our application up and running, and I will open the dev tools and do a hard refresh.

And you can see that we are sending out the API board one, the server return 200, okay. And also in the response header here, it has this cache control, ⁓ public max860 seconds and then the ⁓ SWR 30 seconds. And now if I click this refresh button, you will see the size is this cache, that means it returns from the cache directly.

And also you can notice that it's a slightly different color here. ⁓ and if I click the refresh again, it will always serve from the ⁓ the cache. And after 90 seconds I can click the button again and we'll see it will get the data from the server again. But before the expired time, we will always get the data very quickly from the cache. And if we open up this detail, the browser will tell you it's using the desk cache. And please note that I have unticked the disabled cache.

Sneha Mehra (00:07:15)  
option here. So normally when you're doing your local development, you probably have this ticked and that will ⁓ distribute cache. That means it will always get data from the back end. That's probably ⁓ not what you want to test this out. Just keep in mind that we need to untick this one if we want to test this caching behavior. Alright let's do the refresh again. And this time it's not from the cache anymore. ⁓ You can see it's sending actually a request to the server side.

and we get this data back again. And if we do refresh again, it will start from the cache ⁓ like before. So this is a simple demonstration of how to use the cache control ⁓ in your server side. And next let's talk about e tag, which stands for entity tag. You can think of it as a fingerprint or unique identifier for a specific version of a resource ⁓ user or board. It helps the browser know

whether something has changed since the last time it was fetched. Here's how it works in practice. So on the first request, the server returns the response body along with the e tag header. Something like user ⁓ ABC 123, things like that. The browser stores both the response data and the E tag. Later when it makes a next request for the same resource, the browser includes the ETAG in the request header

like this, it will be in a if non match with that tag ⁓ user ABC123. Now the server compares the value with the latest version it has. If nothing has changed, it simply replies a 304 not modified. And there's nobody or no extra data transferred. So if something has changed, it returns a new 200 response with updated data and the new E tag.

And the broader will store the data and the e tag again. So this mechanism saves bandwise, reduces the server load, and gives users the perception of ⁓ instant updates, because most of the response can be served directly from the cache. Let's look at a practical example in our board application. In our board API, I have added a E tag generation. So we'll use a correctly hash MD5 algorithm and we will use the field data ⁓ to calculate a MD5 of that.

Sneha Mehra (00:09:40)  
a filter data and we will click this e tag and ⁓ we will set that as a e tag header and then later on when another request comes in the browser will have this if no match and will compare the e tag we calculate at this runtime and if they are matching we just return a 304 without any data attached. So the server will use the catch the data. Otherwise it will go to the E tag generation and then serve the data from the normal response.

Which is a 200\. Let's see how it works. Actually, let me ⁓ temporarily decrease this one as ⁓ zero ⁓ and ⁓ zero. So we will not store it ⁓ very long. And then in the browser let's see how it works. And we have this 200 ⁓ response, and if we open it up ⁓ for the details in the response header, we can see that it's a e tag. This is a fingerprint of the data in the response like this one.

And ⁓ if we now do a refresh, you can see that 304\. And if we open up this inner request, we are sending out if now match ⁓ with this e tag, which is a previous one we get from the server side. So the server now can compare these values and because the data is exactly the same, so it will simply return a 304 without the extra data. So in here you can see that we're getting point two kilobytes.

which is just a header. And if we do another refresh, you will notice that it is sending out a request, but it only gets very small data from the server side, which is a header. And if we do let's say research ABC, now we get a 200 response. And if we open it up you can see the payload is different. It has no card with a ABC. But now it will have a new different E tag.

And then if we remove this one, it sends another request, which is the old one, it has the old tag. So if I type ABC again, you can see we got a 304\. And in the request header you can see that it's the if no match ⁓ with this e tag. ⁓ And that e tag is ⁓ from the original ABC request. if we look at the response header ⁓ the A tag is

Sneha Mehra (00:12:09)  
this one. That means we are still sending requests to the server side, but the server will do the comparison and then just send a a header. So with this e tag in place, the browser can quickly check if cached version is still valid and skip downloading the full payload if nothing has changed. Please note that there are many ways to generate the E tags. We are using MD5, but you can still using a number of other ⁓ algorithms like a timestamp or things like that. What matters here is that

It changes only when the outline data changes. So together with a cache control and a E tag, we can form the foundation of the conditional caching. The browser doesn't have to guess whether data is fresh, it simply asks the server and reuse ⁓ cache data whenever possible. The next header is Lust Modified, which works in a similar way to e tag, but it's simpler to implement. Instead of a hash or

Fortune identifier it relies on a timestamp that represents when the resource was last changed. Here's a typical example you might see. When server returns the data, it has a header set as last modified the timestamp. And when the browser makes the following up request, it includes this timestamp in the request header if modified since. The server then compares this date with the current update time ⁓ for the resource.

If the data hasn't changed since the timestamp, it responds to 304 not modified. And if it has changed it send a 200 response ⁓ as normal with the updated content and ⁓ new lust modified value. This approach is great when your data already has a clear lost updated field. For example, ⁓ a user has updated that name at some point.

or a post blog post published data or a database record updated at field. Use Lust modified when ⁓ your data naturally includes ⁓ a modification timestamp ⁓ or something that doesn't change frequently. Or you maybe you just need a simple lightweight validation mechanism. In many cases, this is all you need to keep your API response ⁓ fast and efficient. The Lust header we will cover here is very

Sneha Mehra (00:14:32)  
This one doesn't control freshness or ⁓ validation like the others. Instead it tells caches ⁓ that what makes two response different. You can think of it as a hate to cache systems, ⁓ a way of saying store separate version of this response depends on ⁓ certain request headers. So here's an example. ⁓ the very header is set as accept accept ecoding authentication.

This means the cache should treat request differently based on three facts. The accept header, for example, whether the client wants a JSON or ⁓ XML, the accept encoding header, ⁓ zip broadly or none, and the authentication header, meaning the response might depend on the logged in user. Without this information, cache can accidentally serve the wrong version of a response.

For example, imagine user A fetches their personal ⁓ board ⁓ data, and if that response is cached ⁓ and the cache isn't aware it depends on authentication, user B might receive the same board data, which is obviously not what we want. This is why very header is so important whenever response depends on user identifier, ⁓ language or format. Let's see a few examples here.

And in this example we vary by accept language. So English or Spanish or Chinese user each get their correct localized content. So in a content MPI endpoint, we just set get the content and then we set the vary as accept language. And then in collect side, the browser can decide to store different versions of the data in different language. So the main idea here is simple. Whenever a request header affects a response,

You should include that header name in Vary. It is indeed a small configuration detail, but it makes a huge difference in keeping caches accurate and making the user experience reliable.

Sneha Mehra (00:16:41)  
Alright, let's take a step back and summarize what we have covered. We have talked about the performance from two perspectives, the application level and the infrastructure level. At the application level, you have already implemented techniques like code distribution, ⁓ normalization, virtualization, pagination, things like that. And that makes your application logic more efficient. ⁓ And at the infrastructure level, you control how your application is ⁓ delivered using

Compression, ⁓ modification, CDNs, and as we focused on today, HTTP caching ⁓ headers. So those four headers, ⁓ caching control, e tag, lust modifying, and vary. And they work together as a complete caching system. So this is the real world example that has all these four headers are used. So the cache control defines how long response can stay fresh.

ETAC and last modified handle ⁓ validation, letting the browsers checking whether data has ⁓ changed before downloading again. And the very ensures the cache ⁓ doesn't mix up responses that depend on different headers, such as user ID or language. Together they create a fast and reliable caching flow. So the key takeaway here is that you don't always need to rewrite your front end.

To make your application feel faster. Sometimes the biggest gain comes from how you configure your backend and the delivery. HTTP caching is one of the simplest and most powerful ways to achieving that.

—----------------

Sneha Mehra (00:00:00)  
In the previous modules, we discussed several common techniques for performance enhancement. We have done code splitting, perfeting, normalization, pagination, all the typical application level optimizations that improve runtime, efficiency, and reduce bundle size. But sometimes even after all these improvements, users still feel that the application loads slowly. So what is missing here?

Generally speaking, we optimize performance at two levels. ⁓ Application level and infrastructure level. So for application level, this is what happens in your code base. You do the code splitting, preloading, data normalization, virtualization, and so on so forth. ⁓ Well, infrastructure level ⁓ is ⁓ about how your application is delivered to ⁓ users. Things like compression, modification, caching, and using a CDN.

Most foreign end engineers spend the majority of their time optimizing at the application level, but the infrastructure level can often make a much bigger difference, especially for perceived performance, which is what users actually feel. In this lesson, we will focus on one of the most powerful infrastructure level optimization, HTTP caching headers. With just a few configuration changes, you can make your application feel significantly faster without touching.

most of your code. We will go through four essential headers the cache control, e tag, last modified, and vary. Each plays a different role, and together they form the foundation of efficient browser caching. Let's start with cache control, which defines who can cache a response and for how long. It's most important header when it comes to controlling browser and CDN behavior. Here is a real example you might see

You might find in your browser for some of the response you can see things like this cache control public ⁓ max age 300 still well revalidate 60 seconds. Let's break this down. Public means the response can be cached by anyone in a network. The browser shared proxies or CDN nodes. MaxH 300 means the content is considered fresh for 300 seconds or five minutes and still well revalidate.

Sneha Mehra (00:02:25)  
sixty seconds tell the browser it can temporarily serve an expired copy of the data for sixty seconds while it's fetching a new one in the background. So in our board code base, the user profile data doesn't change very often. We can safely cache it for five minutes, for example. So when a user opens a board, the profile appears instantly from the cache. ⁓ So we don't have to do the cache call

server side. And for these directives, here are some typical patterns you will find useful. So for static assets, ⁓ we can make it much longer max age value. ⁓ And ⁓ that means the file ⁓ rarely changed and aversion by their file names normally. So we can safely catch them for a month, for example. So for ATI responses we normally can set the catch control as part of it and we give it a shorter max age.

That means the cache can only be ⁓ stored in browser, not all the proxies or CDNs. So for sensitive data, we'll set the cache control as no store, so we'll never cache authentications or user specific information ⁓ in the browser or any nodes in the network. And for dynamic or frequent changing content, we will set the cache control as no cache, then still using caching but forcing a validation check ⁓ with the server before use the cache that.

data. So no-cache option is ⁓ often confusing people. No cache doesn't mean don't cache. It means you can you can store it but always confirm with the server before show using the data or showing the data. Here it's a very simple example of how to set the cache control header ⁓ to the user API. We set it as private, that means we only ⁓ use the cache or store the data in the browser side, not only proxies

or in the CDN nodes and we will save data for 300 seconds, ⁓ five minutes, basically. Once the browser has a valid cached copy, it serves that content instantly. There will be no network delays, ⁓ no CPU cost. It's what makes the interface feels ⁓ lightning faster. So for example, if you have this cache control public, max H300 still will revalidate ⁓ SXT. So in the timeline at time zero

Sneha Mehra (00:04:51)  
We send all the requests to user two and serve responses and then we will save the cache in the browser. And at 200 seconds the client send request to get the same UIL. The browser will serve the data from the cache. That will be much faster. And at point three hundred fifty, that is still in the three hundred plus sixty range. So the browser will serve the still data. So it will think the data is already

you know outdated but still it can serve it. But at the same time it will fetch data in the background. This is where the HWR means here. And then at four hundred seconds we send request, but because the data has already been refreshed, we will use that data from the cache. So for the board API, I'm setting the ⁓ cache control handler as public and I will save that for a minute and I will do the HWR as ⁓

30 seconds. Add this refresh button temporarily here. So when you click the button, we will send the request to API board one. So if the data is already in the cache, we will just serve it from the browser. Otherwise, after 90 seconds, we will actually fetch the data again. So let's see the in action. So this is our application up and running, and I will open the dev tools and do a hard refresh.

And you can see that we are sending out the API board one, the server return 200, okay. And also in the response header here, it has this cache control, ⁓ public max860 seconds and then the ⁓ SWR 30 seconds. And now if I click this refresh button, you will see the size is this cache, that means it returns from the cache directly.

And also you can notice that it's a slightly different color here. ⁓ and if I click the refresh again, it will always serve from the ⁓ the cache. And after 90 seconds I can click the button again and we'll see it will get the data from the server again. But before the expired time, we will always get the data very quickly from the cache. And if we open up this detail, the browser will tell you it's using the desk cache. And please note that I have unticked the disabled cache.

Sneha Mehra (00:07:15)  
option here. So normally when you're doing your local development, you probably have this ticked and that will ⁓ distribute cache. That means it will always get data from the back end. That's probably ⁓ not what you want to test this out. Just keep in mind that we need to untick this one if we want to test this caching behavior. Alright let's do the refresh again. And this time it's not from the cache anymore. ⁓ You can see it's sending actually a request to the server side.

and we get this data back again. And if we do refresh again, it will start from the cache ⁓ like before. So this is a simple demonstration of how to use the cache control ⁓ in your server side. And next let's talk about e tag, which stands for entity tag. You can think of it as a fingerprint or unique identifier for a specific version of a resource ⁓ user or board. It helps the browser know

whether something has changed since the last time it was fetched. Here's how it works in practice. So on the first request, the server returns the response body along with the e tag header. Something like user ⁓ ABC 123, things like that. The browser stores both the response data and the E tag. Later when it makes a next request for the same resource, the browser includes the ETAG in the request header

like this, it will be in a if non match with that tag ⁓ user ABC123. Now the server compares the value with the latest version it has. If nothing has changed, it simply replies a 304 not modified. And there's nobody or no extra data transferred. So if something has changed, it returns a new 200 response with updated data and the new E tag.

And the broader will store the data and the e tag again. So this mechanism saves bandwise, reduces the server load, and gives users the perception of ⁓ instant updates, because most of the response can be served directly from the cache. Let's look at a practical example in our board application. In our board API, I have added a E tag generation. So we'll use a correctly hash MD5 algorithm and we will use the field data ⁓ to calculate a MD5 of that.

Sneha Mehra (00:09:40)  
a filter data and we will click this e tag and ⁓ we will set that as a e tag header and then later on when another request comes in the browser will have this if no match and will compare the e tag we calculate at this runtime and if they are matching we just return a 304 without any data attached. So the server will use the catch the data. Otherwise it will go to the E tag generation and then serve the data from the normal response.

Which is a 200\. Let's see how it works. Actually, let me ⁓ temporarily decrease this one as ⁓ zero ⁓ and ⁓ zero. So we will not store it ⁓ very long. And then in the browser let's see how it works. And we have this 200 ⁓ response, and if we open it up ⁓ for the details in the response header, we can see that it's a e tag. This is a fingerprint of the data in the response like this one.

And ⁓ if we now do a refresh, you can see that 304\. And if we open up this inner request, we are sending out if now match ⁓ with this e tag, which is a previous one we get from the server side. So the server now can compare these values and because the data is exactly the same, so it will simply return a 304 without the extra data. So in here you can see that we're getting point two kilobytes.

which is just a header. And if we do another refresh, you will notice that it is sending out a request, but it only gets very small data from the server side, which is a header. And if we do let's say research ABC, now we get a 200 response. And if we open it up you can see the payload is different. It has no card with a ABC. But now it will have a new different E tag.

And then if we remove this one, it sends another request, which is the old one, it has the old tag. So if I type ABC again, you can see we got a 304\. And in the request header you can see that it's the if no match ⁓ with this e tag. ⁓ And that e tag is ⁓ from the original ABC request. if we look at the response header ⁓ the A tag is

Sneha Mehra (00:12:09)  
this one. That means we are still sending requests to the server side, but the server will do the comparison and then just send a a header. So with this e tag in place, the browser can quickly check if cached version is still valid and skip downloading the full payload if nothing has changed. Please note that there are many ways to generate the E tags. We are using MD5, but you can still using a number of other ⁓ algorithms like a timestamp or things like that. What matters here is that

It changes only when the outline data changes. So together with a cache control and a E tag, we can form the foundation of the conditional caching. The browser doesn't have to guess whether data is fresh, it simply asks the server and reuse ⁓ cache data whenever possible. The next header is Lust Modified, which works in a similar way to e tag, but it's simpler to implement. Instead of a hash or

Fortune identifier it relies on a timestamp that represents when the resource was last changed. Here's a typical example you might see. When server returns the data, it has a header set as last modified the timestamp. And when the browser makes the following up request, it includes this timestamp in the request header if modified since. The server then compares this date with the current update time ⁓ for the resource.

If the data hasn't changed since the timestamp, it responds to 304 not modified. And if it has changed it send a 200 response ⁓ as normal with the updated content and ⁓ new lust modified value. This approach is great when your data already has a clear lost updated field. For example, ⁓ a user has updated that name at some point.

or a post blog post published data or a database record updated at field. Use Lust modified when ⁓ your data naturally includes ⁓ a modification timestamp ⁓ or something that doesn't change frequently. Or you maybe you just need a simple lightweight validation mechanism. In many cases, this is all you need to keep your API response ⁓ fast and efficient. The Lust header we will cover here is very

Sneha Mehra (00:14:32)  
This one doesn't control freshness or ⁓ validation like the others. Instead it tells caches ⁓ that what makes two response different. You can think of it as a hate to cache systems, ⁓ a way of saying store separate version of this response depends on ⁓ certain request headers. So here's an example. ⁓ the very header is set as accept accept ecoding authentication.

This means the cache should treat request differently based on three facts. The accept header, for example, whether the client wants a JSON or ⁓ XML, the accept encoding header, ⁓ zip broadly or none, and the authentication header, meaning the response might depend on the logged in user. Without this information, cache can accidentally serve the wrong version of a response.

For example, imagine user A fetches their personal ⁓ board ⁓ data, and if that response is cached ⁓ and the cache isn't aware it depends on authentication, user B might receive the same board data, which is obviously not what we want. This is why very header is so important whenever response depends on user identifier, ⁓ language or format. Let's see a few examples here.

And in this example we vary by accept language. So English or Spanish or Chinese user each get their correct localized content. So in a content MPI endpoint, we just set get the content and then we set the vary as accept language. And then in collect side, the browser can decide to store different versions of the data in different language. So the main idea here is simple. Whenever a request header affects a response,

You should include that header name in Vary. It is indeed a small configuration detail, but it makes a huge difference in keeping caches accurate and making the user experience reliable.

Sneha Mehra (00:16:41)  
Alright, let's take a step back and summarize what we have covered. We have talked about the performance from two perspectives, the application level and the infrastructure level. At the application level, you have already implemented techniques like code distribution, ⁓ normalization, virtualization, pagination, things like that. And that makes your application logic more efficient. ⁓ And at the infrastructure level, you control how your application is ⁓ delivered using

Compression, ⁓ modification, CDNs, and as we focused on today, HTTP caching ⁓ headers. So those four headers, ⁓ caching control, e tag, lust modifying, and vary. And they work together as a complete caching system. So this is the real world example that has all these four headers are used. So the cache control defines how long response can stay fresh.

ETAC and last modified handle ⁓ validation, letting the browsers checking whether data has ⁓ changed before downloading again. And the very ensures the cache ⁓ doesn't mix up responses that depend on different headers, such as user ID or language. Together they create a fast and reliable caching flow. So the key takeaway here is that you don't always need to rewrite your front end.

To make your application feel faster. Sometimes the biggest gain comes from how you configure your backend and the delivery. HTTP caching is one of the simplest and most powerful ways to achieving that.

—-------------------------------  
Sneha Mehra (00:00:00)  
One of the biggest complaints users have about web applications is slowness. You click the button, wait for a spinner to show up and wonder why the application keeps reloading the same data you just saw a second ago. As developers, we know exactly what's happening. Every time the UI component mounts, it's ferry another network request. That means ⁓ wasted time, wasted bandwidth, and a frustrating user experience.

What if we instead of asking the server again and again, ⁓ we keep a local copy of the data and reuse it? And that is caching. So what exactly is a cache? A cache is just a temporary storage. It remembers data you have already fetched, so you don't need to fetch it again. So instead of going to a station every time you check the departure board, you take a photo of the schedule. You can look at it on your phone whenever you want.

until it changed and then you will need a new photo. On the front end that node might just be a JavaScript map. Each cache has a few basic operations. You can set a value, give the cache a key and some data, you can get a value back, you can check if the key exists, and you can delete or clean entries. That's a foundation. So for example, you can declare a new map simply like a verbal declaration and you can set the key with a value associated with it.

And the value can be different types, a number, a string, or even object. And you can check if the key exists in the cache. ⁓ If it does, you can do some ⁓ operations, you can even clean the cache completely. Over time, developers have discovered patterns that make cache more effective. The first pattern is still well revalidate or HWR. The problem itself is simple. The UI feels slow if we always wait.

For fresh data. And the solution is reserve the cached data instantly and refresh it in the background. That gives the user a faster experience while still keeping the data up to date. The second common pattern is time to leave or TTL. The problem is that the cached data can stay around forever, even when it's no longer valid. So we need to give every entry an expiration time. Once it's expired,

Sneha Mehra (00:02:24)  
We refresh before showing again. Those are very common and classical caching patterns, and we will use them those patterns in our applications ⁓ in the following lessons. There are also strategies you will ⁓ almost always need when you build a cache. One is the request deduplication. If two components are asking for the same data, the cache should avoid firing two network calls. Instead, it ⁓ reuse the same.

in flight request and share the result in a single global ⁓ cache. And another strategy is menu invalidation. When the user adds a new to do or update their profile, we cannot rely on the old cache anymore, so we need to ⁓ either clean it or patch it. So the UI stays consistent. Those are not formal patterns like SWR or TTL. They're just practical strategies ⁓ that keep your cache useful in

real world applications. Let's see this in action with a small example. So basically we are going to build something similar to React Query or Ten Stack query, ⁓ but it will be an oversimplified version of it. So basically we'll need a cache in a global ⁓ store and then we will have a customer hook that can access this state. So firstly we will need to define a cache entry type. So each entry in a cache will have this format. We have data, we have tem step,

will have a flag that indicated if the cache is in loading ⁓ and if there's some arrow happened in the cache entry. And ⁓ a instance of that entry will be something like this. The data is the ID, ⁓ name, ⁓ list, and the time step is a number. We can use that to validate ⁓ if the cache ⁓ is expired and ⁓ the loading and arrow is indicating the response status.

And ⁓ you can imagine that at some given time the cache ⁓ is ⁓ having some data like this. We have API users that has data of the users list. ⁓ API slash me has ⁓ the current user data, and the posts will have the corresponding post list. And we will need to define a provider component or context that has a access to this map.

Sneha Mehra (00:04:48)  
And correspondingly we'll need to expose a customer hook. ⁓ we call it use query. And the use query accepts two parameters, the key and a fetcher. So basically we'll inside the ⁓ use query, it can access the query context, which is a provider we defined above. Basically, it does two things normally. It will first check if the key exists in the cache. If it does, we simply return whatever it is already in the cache.

Otherwise we will trigger fetch and then save the data into the ⁓ corresponding key. And because we always have one single instance of a cache, every component erupted by this ⁓ query contacts or query provider will have the access to the cache. So for example, in if in one component we triggered a use query ⁓ to fetch a data for users ⁓ two, and then on the other component, the sibling component, it triggered the sim query.

And we can deduplicate it in this ⁓ global space because we have already got the information ⁓ for the key in the cache. It's already there, we don't have to ⁓ refetch it, we just return whatever ⁓ it is. Otherwise, we can cancel the previous one with a board controller and ⁓ only send one request. And when the data returns, ⁓ all the components will be updated correspondingly. So our application will be changed to something like this. ⁓ the application will be wrapped

By the query provider and the user list component can use a user query to do the data fetching, the board view can do its own data fetching, and their children can do their own data fetching. But under the next they are all accessing the same cache so they can share the same data consistently. So basically each time we call a user query ⁓ with the correct parameter, it will insert the ⁓ key into the cache or update that entry in the cache.

And over time we will have all the data we needed to run the the ⁓ the application ⁓ from different entries ⁓ that triggered from different components. So that way our component can become very simple. We don't manage local state, we don't check the loading flags, ⁓ we just declare what we want, and then the cache handles ⁓ the rest. So why go through the effort of adding a cache?

Sneha Mehra (00:07:14)  
The biggest benefit is performance enhancement. So we now need to refetch what we already have. ⁓ The number two is user experience. So now with the cache the data appears instantly. Instead of waiting for a spinner, it just shows up ⁓ right away. And number three is consistency. Different parts of the application share the same sort of truths, so they will all have the same data, same value. ⁓ And number four is reduced server load. We don't have to always

request the same data. ⁓ with fewer duplicate requests means less pressure on the back end. This is why libraries like RedQuery, SWR or Relay exist. They are essentially maintaining big cache under the hood. They're robust product ready ⁓ caches. Underneath they are built on top of the same principle we just saw in this ⁓ lesson. So let's wrap it up. A cache is just a temporary storage

For data we want to reuse. The basic operations are just set, get, has, and delete. We also discussed a few ⁓ common patterns, the still will revalidate and time to leave TTL ⁓ patterns. The benefit of using cache are speed, smoother ⁓ user experience, consistency, and fewer requests. Even a small cache can transform how your front end feels. And once you see the value, you will understand the tools like RedQuery.

are really just the smarter caches under the hood. So but caching alone isn't enough. Sooner or later, users won't just read data, they will want to change it. ⁓ They will add new cards to the board, they will change the column name, they will update their profile or delete a comment. And that raises new questions. How do we keep the cache in sync after a mutation? Do we patch data locally or refetch from the server

And how do optimistic updates fit in? That's exactly what we will explore in the next lesson. Data mutation. We'll look at different strategies for updating the cache when the underlying data changes, so the UI stays responsive and consistent.

—--------------------------------

Sneha Mehra (00:00:00)  
So far we have looked at the data modeling and the data stretching. This gives us a way to structure our application and load the data we need. But here's the thing. Real applications are not just about reading. ⁓ Users want to change data and that's a data mutation. And when we talk about data mutation, we don't just mean ⁓ one operation. It actually covers three main actions: ⁓ add, update, and delete. In Rust of APIs, ⁓

Those usually a map to post, patch or put and delete. ⁓ So once they create, we post the data to the backend and then ask it to create some resource. And update is ⁓ about put or patch. depends on the usage. put is for like completely replace the ⁓ data, well patch is partially change the fields. For example, you just change the username, ⁓ you probably will use the patch. And if you are

update everything for the user you will use the put. And for delete you can send a ⁓ delete request to delete the ⁓ entity from the back end. So typically in your ⁓ replic application you can use a fetch API with the URL you want to ⁓ request and you prepare the header, ⁓ the body and the method for this fetch API and that will talk to the backend API to ⁓ do the operation. And each of those operations has its own patterns and

problem to solve. So for add ⁓ or create we need to think about how new IDs are generated, where it's generated and how to place the new item into our UI. And for update, often it's a partial update, which means deciding how to merge the new fields with the old state ⁓ that will be a challenge. And for deletion it's pretty straightforward ⁓ on the surface, but it's kind of tricky in the UI.

Do we need to remove the item right away or wait for the user to confirm? And also ⁓ how do we update local storage to make it think ⁓ in components? So basically a mutation isn't just ⁓ sending data, it's about handling the nuance of each type of change. And when we have cache in front end and ⁓ the interaction of the cache and the mutation will be ⁓ very interesting. In modern front-end applications, we

Sneha Mehra (00:02:23)  
really just ⁓ wait for the backend ⁓ response. We combine the mutation with caching ⁓ and that allows us to show change instantly in the UI, keep different components in thing, and ⁓ reduce unnecessary ⁓ fetches. And without the connection of your cache and mutation, your application will feel sluggish and inconsistent. So to make this concurrent, let's start with the most common action, adding a card to a board

as example. So let's ⁓ have a look at how we can implement it ⁓ step by step. So in our codebase again in the mocks we're using MSW to mock the API. So I have added a new endpoint about the post method for API slash cards and it accept a request object which contains a title and a column ID. So basically we want to create card object

And we want to insert this card into that column. As the title is the only thing we will need for now. So if the title or column ID is not defined, we just throw a 400 saying that request is invalid. Once we have this ⁓ title and column, we will find that column ⁓ from the board ⁓ first. And if a column cannot be found, we will return a full four. And once we have that ⁓ column, ⁓ we will create this new card, we will give it

ID and the ID is tickets dash total cards which is ⁓ number of cards at the moment in a board and we ⁓ plus one so we'll have a new card ID and then we use this to create a new card and push that to the column of the board and then we simply return the new card object with the status 201 which means the resource is created ⁓ and then in front end we will

need to call this API in the column. So on the UI we'll have a component here. It's basically input box and a button attached to it and each column will have this create card form kind of component. So we do ⁓ a new card and we click this button and it will insert this into the current column. And if we do it in another column, let's say here

Sneha Mehra (00:04:47)  
⁓ hello word and if we trigger the action it will insert this into the corresponding column. So if we look at the column component we have a handle create card action. Basically we'll send a post request to API slash cards and we will prepare the body which is column ID plus the title and title is from this input.

That's a very simple implementation and we will handle the key down event and then we'll handle create card ⁓ and basically we'll just check the title, make sure it's not empty, and then we set it's creating, and then we send a request, and if response is okay, ⁓ we just unpack the response and ⁓ call the add card action. That's a new action we add to our contacts.

Which is manipulating the normalized store. And that store is a global store we use in all the other places for our UI. For example, all the cards, columns, board, the assignees, users, everything is leaved inside that ⁓ store. So ⁓ if you look at the add card action here, which is from the contacts, ⁓ the border contacts, and in here we have the add card. Add card action will simply update the normalized.

Store. So basically we'll update all the cards. We'll insert the new card object into the cards list. And we will also find the column and update the columns ⁓ array and making sure the cards ⁓ in the column is updated as well. So basically if we are inserting ⁓ a card new card to column one, we'll need to find column one and then update the card IDs for that column.

And add insert the card ID at the end of the card IDs. And then we will put back the cards ⁓ and columns object back to the store. ⁓ So we will have the updated store and then that will re-trigger the UI to re-render. And then we'll get the hydrated cards and for the columns. It will run these ⁓ cards in a column. So we'll have all the updated version of the UI. And then as you can see, we will get the ⁓ cards inserted through a particular column.

Sneha Mehra (00:07:10)  
interst did the refresh and we can do that again. In here, hello word, ⁓ and we can insert card again. So that's the simplest version of a mutation. The user triggered action ⁓ that will call the API and we will update the store and up the store will update the UI. ⁓ if there is anything wrong, we will handle the arrow as well. So now that you have the this floor for adding a card, it's your turn. At the checkpoint you will implement a deleting

A card, ⁓ the reverse of what we just did. ⁓ after that, we will move on to one of the most interesting patterns in front end system ⁓ updating, which is optimistic updates. ⁓ we will start by assigning a user to a card and showing ⁓ the change instantly, even before the server response. Later we will combine ⁓ that with ⁓ real-time updates. So changes ⁓ sync across multiple clients.

So to recap, ⁓ mutation isn't just one thing. It's add ⁓ update and delete is more complicated than the data fetching, ⁓ each with its own challenges. And when you combine mutation with caching and optimistic updates, you know, real-time updates, your application feels smoother, faster, and consistent.

—---------------------------------------------

Sneha Mehra (00:00:00)  
When you hear server-side rendering or SSR, it can sound a little bit intimidating. But really, it's just a way of saying, instead of letting the browser build the initial HTML, we'll build it on the server side and set it down, ready to go. Let's step through the life cycle of a single request. From the moment user type the URL, the browser sends a request to your server. Let's say the user goes to board one.

If the server receives that request, instead of just serving a static HTML, it kicks off a rendering process. On the server, your React code runs, it builds the component tree, fetches whatever data needed, and produces HTML for that route. The server sends back a full HTML document. This HTML already contains the markup for the board, not just a blank div.

the browser starts parsing and painting that HTML right away so user sees content quickly meanwhile the bundled JavaScript arrives in the background when it runs React hydrates the existing HTML meaning it attaches event listeners and makes the page interactive without rerunding from scratch so the key idea is ⁓ SSR gives the user real content immediately and hydration

change the static HTML into a live Reactor application. So for our board application, we have a running instance in this port, 5173\. It's a default white port. And you can see when I refresh, there is a slight flash here. You can see the page turns from blank to full content. And if I make the network slower in here, with the faster 4G as a stuttering strategy,

And if I do a refresh, ⁓ you see there's a blank page for two seconds maybe and then it starts to render the full application. ⁓ And on another tab, I have a server-side rendering, ⁓ same application. And I will turn a throttle to fast 4G as well. And when I hit refresh, ⁓ immediately you can see the HTML and then later on there will be some JavaScript loaded.

Sneha Mehra (00:02:25)  
So that means in slower network, the SSR definitely has a better user experience. So if I turn the start link to slow 4G and I do a refresh, and almost like immediately you can see the HTML, but I will try to click this list button while it's loading, let's say. ⁓ But you go see it's not inactive. I cannot click that until very late, but the initial render will be way faster.

And on the other side, if we go to the client side rendering, if I turn a slow 4G, and now if I do a refresh, basically that will be empty. I cannot do anything here, but wait for ⁓ after, ⁓ I don't know, five seconds already passed. ⁓ And yeah, it's still ticking, still loading. But once it's returned and rendered,

It's immediately interactive. I can click the list button without delay, but you can see that initially there is a blank page for several seconds. There's nothing the user can see. And if we look at the generated HTML in the client side, let me turn off this no throttling. And ⁓ if I do a refresh and if I on the page, I view page source, you can see in the root element there.

a lot of things already in the content. And at the bottom, is a window, initial data is some data and the board ID is 1\. If we go to the client side rendering, ⁓ let me turn off this throttling. If I go to the board, view page source, and you can see there's only ⁓ one ⁓ empty element ⁓ ID root in the browser.

And there is no actual content in the HTML document. So you can see how different they are. The server-side rendering has all the HTML tags, the CSS class names, and everything. So now how do we actually enable SSR in a typical front-end application? So there are four main steps. Step one, request handling. You need a server that can receive HTML requests and decide what page to render.

Sneha Mehra (00:04:50)  
Often that's an express server in Node.js. I will see that in a minute. Step two, rendering on the server. So instead of calling create root in the browser, you use Reactors ⁓ server API, render to pipable stream or render to stream on the server. You wrap the application component with the right provider, pass it in initial data and generate HTML.

Step 3 is sending HTML back. ⁓ The server writes out a full HTML shell with a header, ⁓ a root div containing the rendered reactor markup, ⁓ and a script tag that abandoned any initial data the client will need. Then it appends a script tag that loads your client bundle. ⁓ Step 4, ⁓ hydration on the client. So on the browser side, instead of creating root, you call hydrant root.

React takes the HTML, it sits in the DOM and attached interactivities. From that point on, the app behaves like a normal client-side React application. So there are a few critical moving parts here. The server is responsible for data fetching. If your page needs user info or board data, you fetch it before sending or running so the HTML already contains it. Initial data has to be shared.

Whatever you fetched on the server must also be passed down to the client, usually by serializing them into a script tag. ⁓ That way the client can hydrate with the same state the server used. ⁓ Hydration must match. The HTML from the server and ⁓ the render logic on the client must be in sync. If you don't match, React will warn about mismatch ⁓ during hydration. ⁓ Streaming is possible with render to

pipeable stream API, you don't have to wait for the entire page to render before sending it. You can stream HTML chunks as they are ready, which makes the larger pages feel even faster. Now let's have a look at how it works in the code level. ⁓ So in last lesson, I have already extracted a Express server in a server folder, so that can handle all the API endpoints.

Sneha Mehra (00:07:13)  
But to enable the SSR, I have added a new endpoint called board ID, meaning we will fetch the data, rendering in ⁓ the backend, and then send out the full HTML document to the client side. It happens in here, we are mixing the endpoint and the API altogether in one Express server. But in real world, the API might live in a remote server. We will need to fetch the data from here.

And once we have this board data and user data, we can normalize the data. Once we have this normalized data, we can use the data to ⁓ render the application. To make that happen, I have modified the application slightly. I have modified the board provider to let it accept an initial state. That means we can pass in a parameter to the provider so it has all the data it needed for all the children to render.

So initially we don't have this one. The initiative is only fulfilled ⁓ in the board page. So previously in the board page, we will send the request to the server side to get the user information and the board information. And once we get this data, we will ingest the board and the users. So we will have the full ⁓ store in front end. But this use effect will only happens in client side. ⁓

backend or server-side, cannot execute the user effect. That means we need to prepare the data earlier and then pass it to the board provider. If we go back to the server, so once we have this initial data and the board ID, we will call the render application, which is defined in the entry server. So this is a function that will be called on the server-side. We'll have the data and we will render a board provider with the data we prepared. And then this will be

used to generate the HTML. ⁓ The key point here is using render to pipeable stream from React on server package. So basically the render to pipeable stream will take our reactor tree, which is what we prepared above here, and ⁓ it will render that into a stream. And we can use that stream in the ⁓ end point here. So once we have this stream, we can write to this response with a status if there's anything wrong from the stream.

Sneha Mehra (00:09:37)  
we will return 500, ⁓ otherwise it will be 200\. And then we will write the headers to the response. ⁓ And in the bottom here, ⁓ we will pipe our stream to this writer, which will write to the response stream correspondingly. That means as the reactor render, ⁓ we have some data available and we will write to the response. So the UI can see that HTML.

as soon as possible. And once all the data render is done, it will do the final call here. It will send a end HTML that is the data will be needed for the hydration process. And this is all generated in the backend by the Roundup application. And this is pretty much about the server side. And for the front end, we will also need a new entry, which is entry client.

As you probably remember before, we have a main.tsx. It has a correct root. It will get the root element from the page and then it will render the application component, which is this one. It has a board provider and a board page. So instead of doing that, we have correct a new entry point here, which is entry client. And we're not create the root. We are hydrate the root and we will get the root element from the page.

But the difference here is, by the time when we call the hydrant root, the document.getElementById, the root should have already got all the HTML. So the hydrant will hydrant the data into this HTML. And when you call the create root, normally this element is empty. It has nothing. And when you call the hydrant root, it has everything already. And in here, you kind of like rerun the whole thing again.

And the result of this one ⁓ and the content in the root should be the same ⁓ because you have the initial data and this data should have come from the server-side rendering as well and you can use this initial data to rerun the whole application in front-end. ⁓ is smart enough to compare the HTML at this stage. This should be exactly the same. And the only thing React does here is

Sneha Mehra (00:11:59)  
it will attach all the event listeners to the HTML, ⁓ which is already on the page. So the application turns to interactive. ⁓ And lastly, we have modified the build script in the package.json. So build will build the client and the server. ⁓ So the client side build is still doing the wide build, which is the existing behavior. And for the server build, it will use the SSR and it will use the entry server as an entry point.

and it will put the output into a dist server. And if we open up the dist ⁓ server, entry server, this is the backend bundle. And we will use this render application bundle, which essentially has the JSX runtime export and the compiler enabled. So when we call the initial data to the render application, it will call these entry server to JS and that will handle all the JSX ⁓ syntax here, which compels the

JSX into HTML and execute all our RectLogic code in there. And then we have this stream object we can pipe down to the response writer. And then the front end can get this HTML for the following renders. SSR might feel like a big complex feature, but once you break it down into these steps, the flow is actually very clear. The server will do the data fetching and the render and the browser only display and the hydrant.

Understand this cycle is the key to build faster, ICU-friendly applications and it gives you a solid foundation for more advanced strategies like ⁓ streaming or partial hydration.

—-----------------------------------

Sneha Mehra (00:00:00)  
We have covered a lot in this module rendering strategies, skeletons, profetching, lazy loading, and you have even implemented some of these patterns yourself. But here's the thing, in system design is not enough to say this feels faster, we need to back it up with numbers. That's what this lesson is about. How do we actually measure performance? And which numbers should we care about? Let's start with the core web vitals. Because those map closely to real user experience.

Let's start with largest content for paint or LCP. This is how long it takes for the main content to appear. Imagine you load a news site. LCP is when the big headline and hero image finally show up. If this loading helped shrink your initial bundle, LCP usually improves. The next one is communicative layout shift or CLS. This measures how much the page jumps around after it starts rendering.

Think about a sign up button that suddenly moved down because an image loads above it. That shift is frustrating and CRIs captured. Skeleton screens helps here because they reserve space and prevent that jumping. And next we have interaction to NextPint or IMP. This is about responsiveness. It measures how quickly a website updates ⁓ or shows changes after user interactive with it. For example, you type into a search box and

and the UI doesn't update until a second later. This lag is what IMP highlights. If you cut down the main thread work, INP usually gets better. So those three vital cover when the page looks ready, whether it stays stable, and how responsiveness it fails. Apart from this metrics, there are a lot of other useful metrics as well. The first one is time to first bite or TTFB. This is the time from click a link

To the server, send the very first byte back. For server-side rendering, this tells you if the backend is responding ⁓ quickly enough. If your SSR page takes too long to stream, you will see a slow TTFB. And the next important one is time to interactive or TTI. This is very common term you see in ⁓ performance measurement. This is when the page is not just visible but actually usable.

Sneha Mehra (00:02:26)  
For example, maybe the home page looks ready, but buttons don't work, you cannot click them until a big JavaScript bundle finishes running. That's high TTI. And normally we want to reduce that number. Now we understand that a web voto is important, but how do we measure them? There are normally two common approaches. First, Lab Tools. Chrome DevTools has a lighthouse built in, you can run it against your page and it will give you ⁓ LCP, CLI, CNP and so on.

So Lighthouse can give you a very brief and ⁓ insightful chart based on your application performance. So it will give you a score. You can also drill down to more specific specs like the speed index, the cumulative ⁓ layout shift, largest content for paint, and so on. And for each individual matrix you can have a few more details like this one. The largest content for paint is takes almost three seconds, which is very slow.

And that's why it gives you ⁓ a low number yeah. And you can also use ⁓ page speed if your website is already published. And you give it basically a URL of your website and it will run and analysis the performance of your website and give you a report similarly to what we just did in a lighthouse. For example, if I want to test my own website, I can ⁓ type it in the URL and it will run against my site.

I can test against ⁓ a mobile device or a desktop device and it can diagnose the performance issues for me. And this is a report for the mobile device and also I can switch to the desktop and it can give me ⁓ a bit more details on the metrics which installed. And ⁓ you can see at what point the screen shows something meaningful and they will give you more details on like you probably can reduce some of the requests.

or preload the resource or even remove this completely. So my workflow is once I have the page or application stable enough, I have done the functionality, I will run that against my local build and see how the score tells me. So because the DevTools is in local in my browser, it's very handy to compare the before and after. For example, I can run once ⁓ with the both list view bundle ⁓ up front and once after you lazy load it.

Sneha Mehra (00:04:51)  
And check how the LCP or TDI changes. So this is our bird view. So once I have done my feature development, I would launch my local DevTools inside the Chrome. And if I go to the Lighthouse tab, ⁓ you can run a report inside the ⁓ browser right away. So normally I would do the d a default option for the navigation and I ⁓ normally test against the desktop version first.

And based on the actual usage or the ⁓ application's audience, this board view probably don't fit to the ⁓ mobile users that much. It's mainly used by ⁓ people who use a desktop. And then I can check the performance ⁓ accessibility best practices or SEO, but for now I just say using the performance one and ⁓ run the analysis page load. So as you can see, ⁓ in the ⁓

Dev mode we are running in 5173\. ⁓ we can run that page load ⁓ analysis, but it's not as accurate because it's in local ⁓ dev mode. It's not working the same way as the production version. And after a few seconds I can see these ⁓ score overall and I can see the breakdowns for CLIs, cumulative layout shift, total blocking time, and the first content paint and the largest content paint.

And in my local you can see these two metrics are not as good. That a lot of things can be improved. And if I scroll down to the inside section, there's some suggestions. Network dependency tree. And because I'm running in local, the JavaScript is not bundled. It's purely used as a native module. So this number will definitely be ⁓ a little bit high. And you can see when not minifying the JavaScript and estimated saving of a megabyte. And in a

Diagnostic section there have some suggestions. we can reduce the unused JavaScript. We can minify the JavaScript as well. And we need to avoid the document write, things like that. And if I run this ⁓ same thing with my ISISI abode site, which is running in 4000, ⁓ and if we do that again, hopefully we can get a better result. Yeah, we got a 100, that's great. ⁓ I guess the ⁓ for one we have a relatively small application.

Sneha Mehra (00:07:16)  
and secondly we ⁓ using the server-side rendering for this board view. So initially we have already got the ⁓ the well structured document and we just ⁓ hydrate it later on. So we get a ⁓ relatively good result. And in the inside there only use efficient cache lifetimes, ⁓ partition can see 400 key and for other things it looks relatively good. And here there is a reduced unused JavaScript.

There's some JavaScript we're not using. ⁓ it seems like it mainly comes from the Chromed extension. That means we need to either disable these ⁓ extensions or we use that in ⁓ incognito mode to run this performance test. And this is only for the local. And once we publish the site to a public domain, we'll need to run that against the page speed inside as well to get the better understanding of the performance in the real world.

And that's a typical workflow I'm using for feature delivery. Because sometimes you feel fast in your local doesn't mean it will perform as well ⁓ in a real world scenario. And another category is the field data, that's real world numbers from the actual users. Analytics libraries can capture web vitals in the browser and send them back. And this is important because the Lab Tools runs on your laptop, which is usually very fast, and your Wi Fi, your local Wi Fi, which

is fast as well, but your users might not ⁓ have the similar setup. They might run in a mid-range phone on 3G. Field data shows you what they really experience. So for example, if I open up the PageSpeed Insights ⁓ to test my web page and I got some data scores, ⁓ performance, accessibility, etc.

But on the top of the section you can see there's no data available here that says discover what the real users are experiencing. And if you look at the information here, it says the Chrome cannot ⁓ get the sufficient real world speed data for this page. And that's okay. For my side it doesn't have much traffic. And if we go to a page that definitely has some data, and then you can see there is a dashboard generated based on the real world data usage. So ⁓ we can see the LCP. ⁓

Sneha Mehra (00:09:35)  
INP, CRS and so on. And if you expand the view and you can see more details. On the chart you can see there is a indicator here. ⁓ it's about seventy-five percentile is two point one second. And also for this one for seventy percentile you can see ⁓ it's in the need improvement section. That means the interaction to next paint is not as good. we can probably enhance that. And the COS looks good.

And the time to first buy it is not as good as fit to one second. Well the good condition is less than 0.8 seconds. This is the real user data ⁓ collected by Google and it tells you ⁓ what the page performed in a real world scenario. And that data is very useful for you to keep improving your page's performance ⁓ in the production environment. But obviously you need to let your ⁓ page running in production for a little while. that allows

from to collect the sufficient data to ⁓ generate a meaningful report for you. And if you are working in React you can use a React Profeller. ⁓ It doesn't give you the viral vitals, but it is great for spoiling visited renders. For example, ⁓ if you click a filter and it re-renders the whole page, the whole board of view that is unnecessary. So the Profeller can make that visible. Now here is a key point.

Measuring isn't about chasing perfect scores. It's about closing the loop. You try a strategy, maybe lazy loading, maybe server-side rendering, maybe ⁓ perfeting. Then you measure. Did LCP improve? Did INP drop? Did CLS stay stable? That's how you know if your change actually helped. Because sometimes a pattern that seems clever can backfire. For example,

Covering every section with a skeleton might reduce CRS, but it can also make LCP worse ⁓ if the skeleton loads faster than the actual content. But without measuring, you wouldn't know, right? And here is my challenge for you. Pick a feature in your own project ⁓ and you run a baseline measurement, not the LCP, INP, or maybe TDFB, then apply one of the patterns we have talked about in this module, maybe perfection or

Sneha Mehra (00:11:59)  
Suicide rendering, leads loading, and so on. And then you compare what worked and what didn't. An analysis why that's the case. And then you repeat that process until you get a relatively good or good enough score for your product in your ⁓ environment.

—----------------------

Sneha Mehra (00:00:00)  
Up until now, all of our data mutation has been one way conversation. The user clicks, we send a request and we wait for the server's reply. That works fine if you are the only person touching the data, but what happens when other people are changing it at the same time? Imagine we are both working on the same board. I assign you a task. You shouldn't have to refresh your browser to say it. The update should just appear.

And that's the premise of real time updates, keeping everyone's view of the data in sync. So why real time matters? Realtime isn't just about being fancy or flashy, it's about trust and collaboration. If you are in a shared system, whether it's a project board or chatroom or even a livestock ticker, you cannot afford to say steal data. Think about these cases. A teammate adds a new card to the board.

Another tab delete a card you are looking at. Your manager assigns you a new task. If your app doesn't reflect those changes quickly, people get confused. They might override each other's work or worse, making decisions on outdated information. That's why real time updates are not just nice to have. They are fundamental to collaborative apps. There are three main ways to get real time behavior in ⁓ web applications polling,

ServerSen events and WebSocket. Let's start with polling. Polling is the simplest idea. The client just asking the server every few seconds, are we there yet? Are we there yet? It's that easy to implement. Basically we can have a interval ⁓ and a ⁓ seconds delay. So every few seconds we will do the fetch and check if the ⁓ date has the expected state. And then we can update the UI correspondingly. But it is wished for.

Most of these requests return nothing or not return the thing we are expecting. And if you pull too slowly, updates feel laggy, and you pull too quickly, you overload the server. Pulling works, but it's a blunt tool. You will say it's used in basic dashboards or as a fallback when other methods are not possible. And next we have server send events or SSE. With SSE the clients make a single HTT request and keep it open.

Sneha Mehra (00:02:27)  
The server can then stream down events whenever something changes. ⁓ It's one way and it's only from server to client. And for many applications that's exactly what you need. It's lightweight, reconnects automatically and scales really well for a thousand users. And it's perfect for our board application because the server only needs to tell clients that something is happening.

So SSC has two parts. In client side, we will need to define an event source object and that will connect to a server side endpoint. And on that event source we can add event listener to a customer event. And when that event happens, that means the server is pushing us some new data and we can pass the data ⁓ and use that data correspondingly. And also we can have error handling, ⁓ like we lost a connection and we can ⁓ do the reconnection.

Things like that. And on the server side we will have a global event ⁓ bus or event emitter. And then for example, if we have a card inserted into the database and we can send an event to this event bus. And that handle card assigned will be triggered and we can write data to the response. So in the client side we can get this notification. And note here we are sending the text-based response to the client side. Basically the event, the event name.

and ⁓ a new line and the data and the JSON ⁓ string fied data of the card assignment object we will need to pass to the client side. And finally WebSockets. This is the four two-way channel. The client and the server can both send messages at any time. That makes it incredibly powerful. You can build chat applications, collaborative document editing like Google Doc or some whiteboard application.

and multiplayer games. But it comes with complexity. You need to manage reconnections, handle load balancing, and sometimes even introduce message brokers so multiple servers can share state. And just like a server send event, WebSocket has two parts as well. So in client side you will need to ⁓ create a new WebSocket object and you will need the backend to support the WebSocket protocol. And once you have this WebSocket object you can

Sneha Mehra (00:04:49)  
Register open event handler, the message event handler or arrow event handler. And when server sends you any message, you can pass the data and ⁓ do the corresponding action based on the data. And on the server side, you will need a web socket server and it listen on a particular port. And then when connection happens, you can send messages through that port to the client side. It's basically like a broadcast. And for small applications, WebSocket might be an over queue.

But for the heavy, interactive ones, it's essential. So how do these approaches stack up in practice? Pulling is the simplest, ⁓ but because you are sending constant requests just to hear ⁓ nothing new, it's inefficient. It's creating extra load on your server and depends on your interval, ⁓ can feel slow to users. And the service end event or SSE hit a sweet spot for many applications.

They are lightweight, reliable, and perfect when a server just needs to broadcast updates. Imagine it like turning into a radio station, the server announces change and your app licens to these changes. ⁓ That makes it great for things like notifications, live feeds or our board application, where one person's change should instantly show up for everyone else. And WebSockets are the heavy heaters. They open up a two way.

channel where both client and server can talk freely almost like a phone call. That's essential for chat, ⁓ collaborative editing or multiplayer games. But with the power comes more work. You need to handle scaling, reconnections, and often extra ⁓ infrastructures to make it reliable across servers. And here is the thing, real world applications don't pick just one, they mixed and matched

For example, Slack uses ⁓ WebSocket for chat messages, ⁓ SSE for lightweight presence updates, and even polling as a SIFT net ⁓ if connection drop. The important takeaway isn't to pick the fanciest option, it's to pick the tool that matches the job your application is doing.

—--------------------------

Sneha Mehra (00:00:00)  
When users assign a task to someone, they expect the list of people to appear right away. But if the app only starts fetching data after they click, what they get is a spinner and a short but frustrating delay. In this lesson, I will show you how prefetch solves this problem perfectly. We'll talk about what prefetch is, why it's useful, and when to use it. Then we'll build it to our application so that the assign user drop-down list feels instant.

even on a slow network. So what is perfetch? Perfetch means loading data before the user explicitly asks for it. Instead of waiting until the button is clicked, we start fetching when the user shows intention. That could be hovering over the button, focusing it with the keyboard, or touch it on the mobile. By the time they actually click, the data is already in the cache ⁓ and the UI renders instantly.

This doesn't make the network faster, but it shifts the waiting time to the moment when the user isn't ⁓ paying attention yet. Profesh is useful because it eliminates the dead time between a click and the moment content appears. Instead of clicking and waiting, the content is already there. It works especially well for features like pop-ups, menus, drawers, and detail pages.

Those are places where users expect the interface to respond instantly. It also works nicely with skeleton screens. In fact, if the Perfetch data is already warmly in cache, you don't need a skeleton at all. So when should you consider use Perfetch in your case? So use it when the data is safe to cache and not too large. Use it when you have clear signals of intent, such as hover or focus. And use it when

a short time to live makes sense, maybe a minute or two where the data is still fresh and reliable. Normally in the board application, for example, when you assign a task to a user, potentially there is a user list and the user list won't change as often. And normally it's acceptable to keep this ⁓ user list into your local cache for a minute or two. Let's look at the assign user drop down in our board application.

Sneha Mehra (00:02:21)  
So right now the list of users is fetched only when the drop down opens. That means every click can cause a noticeable delay and we can improve this by using the prefetch technique. So at the moment we have a user assigning on each ticket. When you click it, there will be a user list shows up. Currently we only fetch the first five and when you scroll it will fetch the next five and so on. Because we are running everything in local, there is no much noticeable delay and we can stimulate the

slow network by making the API users a little bit slow. We can add a small delay here, ⁓ with a delay for two seconds. And now if we go to the UI, if you see here, ⁓ the ⁓ set it's loading, and then after two seconds, you can see the actual result. And every time when you scroll, it will trigger another load, and the loading indicator shows up again. And when you click it again, because it's fetching at the time when you click,

So you will always see that loading status whenever you click the button. And there's a loading status whenever you scroll, it will always be another loading status. So if we open up the network inspector and we click the button and you will see there is a users page there or page size five is ⁓ triggered ⁓ and ⁓

You can notice the time here waiting for the server response is two seconds. That's exactly what we set for the simulation. ⁓ And during this two seconds, we cannot see anything. It's just a empty drop down list. And that's something we need to improve. ⁓ So ideally, ⁓ whenever we hover on this button, ⁓ we should trigger the fetch and when you click it, there is a small delay between the click and hover. And ideally, ⁓ we should get enough data to

show in a drop down already. And also we want to keep the data in a cache for the rest of the application. So whenever we have already fetched one user list and another card, when you hover or click, we shouldn't do another fetch because the data has already been in a cache. So we can run that immediately unless the data is expired. So in this fetching module, we have introduced a local cache, which is defined in the components. We have defined a

Sneha Mehra (00:04:42)  
query a provider, but we never use this in our application yet. So basically the query provider is a cache. It's maintaining a map internally for every API call. We will check the local cache first. If the key authority in the cache, we just use the data and by default we will give each entry in the cache a TTL. But for each entry in the cache, we give it a time step, ⁓ which is a time we lost to fetch the data.

And every time when we get a use query call, we will check if the entry is already in the cache and it's still fresh. So we don't have to make another request to the server and we will return the cache entry immediately. ⁓ So to use this query provider, ⁓ we need to put that to the utmost position of the component tree. So all the component can access the cache globally. And also we will provide a hook with the use query.

If you are familiar with React query, ⁓ it's pretty similar. ⁓ It will be used as a key in a cache. ⁓ And then we have a fetch function that will send the request to the server side. And we have also a TTL time, ⁓ which is by default a minute. On each card, we have a user selected component. ⁓ And in that component, we will try to fetch data, use the use query with this key. So basically it has the users, ⁓ the keyword.

that's name you want to type for search. By default it's empty string. And the page size is how many items you want to show for that page. And this is the page number, which is the index. So now we have this key ⁓ users empty five zero. ⁓ And the query function is fetch user for that page with page size five. And then the debunk is again the username you want to search. And we want to

keep the item in cache for a minute. And then once we have the data, we will set the options, which will be the ⁓ user names in the list. This is how it works by default. And the user select is used in card. And in the card component, we have a pop-up defined in here. So basically we have a trigger that is open a signing picker and there's a button. We want to define an event handler for mouse enter. That means when we move the mouse into this ⁓ button,

Sneha Mehra (00:07:00)  
It will trigger the PerfetchedUser call and the PerfetchedUser is simply calling the usePerfetch which triggers a fetchedUser function call and it will keep the item in the cache for a minute. And it will only fetch the users with empty username search query and we only get the first five items. So it will always get the first page of the user list and these PerfetchedUser will be happened only when the

mouse enter on the existing avatar. So now if we open up the network tab and if I hover on, let's say the install dependency button and you see immediately it will send the request for the first page of the user and it get response of the file items. Let me do that again. I will do refresh and when I hover it has less waiting time for the page to show up. And as we have already got that into the cache,

Next time when we open up or hover on another ticket, the request will not be sent because we have that in cache already. And when we click this one, we have already got these first item. And when we scroll, ⁓ it's getting the next page because the next page is not in the cache. So it will show in the dot dot dot indicator and the same thing for the next page ⁓ and ⁓ the rest of the content. But as we have already

get all the data needed for the whole list. And if we go to another ticket and click that, the data will be shows up immediately and when we scroll, it's already there in the cache. So you can feel that instantly shows up all the users. And that's great. That's enhanced the overall perceived performance significantly. So at the moment, the only drawback is that one minute cache, but based on the product requirement,

it's acceptable to have this item in the cache for a minute. So the important part here is that the perfetch runs when the user shows intent. Hovering, focusing, ⁓ or touching the button is enough to trigger it before the user actually click. And when you are using perfetch, there are a few best practices to follow. First, make sure your perfetch keys exactly match the query keys. If they don't match, ⁓ the cache data won't be reused. ⁓ Second, trigger perfetch only from

Sneha Mehra (00:09:24)  
clear intent signals such as hover, focus, or touch on mobile devices. In some cases, you might want to avoid calling the backend too often because the hover can happen quite easily. So you will need to add a few milliseconds ⁓ delay. Like ⁓ when the user moves demos into the existing avatar, you don't trigger the purfetch immediately. Instead, you wait for a hundred milliseconds just in case they moved away. That's a...

very common action like I'm just moving my mouse around, but as soon as I moved into that and immediately I moved out, we don't have to trigger the prefetch because they are not showing the intentions enough. And third, set a sensible time to life. If it's too long, the data can become stale. If it's too short, you will lose the benefit of caching. And fourth, guard against spamming the server. You can debounce prefetch calls or only allow it to run once.

per session. And finally, always provide a fullback. Not every interaction will be per-fetched, so your UI must still handle the keys where the data isn't already in cache. So as you can see, per-fetch is a small change, but it has a big impact. ERV user picker is simple. Hover time request completely remove the spinner and made the drop down feel instant. And that's what perceived performance is all about. Shift the weight away.

from the moment when users are paying attention.

—---------------------------------

Sneha Mehra (00:00:00)  
In this lesson we are going to talk about something that often treated as an afterthought in front end architecture. Accessibility. Accessibility isn't just about compliance checklist or adding you know alternate text. It's about designing systems that work for everyone, regardless of ability, device or context. Good design is accessible design. We're going to talk about some common tools, practices.

⁓ in front end applications and using our board application as a demonstration again. So the key idea of designing a accessible system is simple. Don't retrofit accessibility. Build it in from the start. When you design inclusively you automatically create better products. Accessibility improves ⁓ usabilities for everyone, not just users with disabilities. Here is why it matters. So many regions now

legally require WCEG 2.1 AA compliance. ⁓ Accessible application can reach up to 50% more users, which is good for your marketing or your ⁓ profit. And third, semantic mockups improve SEOs and secret reader support. And also accessibility patterns make interaction cleaner for all users. And they feature proof your design across devices and the input method.

Sneha Mehra (00:01:29)  
Let's work through the foundation patterns that makes a accessible system. Progressive enhancement. Start from a reliable baseline. Your app should still work if JavaScript or CSS fails. Then layers ⁓ on richer ⁓ features progressively. This approach is draws gracefully a degradation. ⁓ Semantic HTML foundation. Use the right elements for the right purpose. Buttons for actions, ⁓ nav for navigation.

Main for content. That structure gives meaning to a UI, helping assistive technologies ⁓ interpretive it correctly. Everything must be usable by keyboard. Set the logic tab order, manage forks carefully, and make sure virtual forks ring or indicator are always visible. And last but not least, automate accessible testing. Automate checks catch common issues earlier and teach developers good patterns over time.

Tools like a JestX and YesLint plugins should be part of your ⁓ toolbox or your normal workflow.

Sneha Mehra (00:02:40)  
So let's first add automated accessible checks directly into our development workflow. So in our board application, I have already added just X, ⁓ which is a tool that will enable us to run automatic checks in the unit test. ⁓ so for example, if I go into my card test, we have defined a few ⁓ tests here. So basically we're still doing doing the render with the provider just like before. ⁓ the

Difference here is that we can use X ⁓ to render. It will automatically check if the container follows the accessibility rules. And it has a customized ⁓ assertion. ⁓ We can do that because we have the test setup. ⁓ We are extending the expand with this to have no validation. ⁓ And that assertion is from ⁓ just X. So with that, if we run a test ⁓ in our local

We can see it's all passing, but if we let's say go to our card component, ⁓ let's make it ⁓ fail by introducing some incorrect pattern here. So for example, list item should be used inside a list container or list, right? You normally do ul ⁓ things like that, A, B, and C, right? This is ⁓ correct usage of the

UL L I and that's fine if we run the test it will pass in the accessibility check. But if we use the ILI element without a UL or OL as a container, ⁓ it will ⁓ fail because ⁓ obviously you should use that with a you know container element. So now if we run the test again, you can see that we have this failure. It expects the HML ⁓ to have ⁓

No validation but received ⁓ this element must be contained in the UL or OL ⁓ list item. So this is the validation. And because we have this check in our test, we can always catch that before it's slipped into the production. And we can remove this. And next we can integrate the yes lint rules ⁓ with these ⁓ accessibility checks as well. So in our yes lint config

Sneha Mehra (00:05:03)  
I have already installed a package called YesLint plugin GSX Accessibility and I'm applying this plugin to all the files with TS, TSX, GSX, GS. And the rules are ⁓ it has some recommended rule ⁓ following and then it has some automatic rules as well. And during our development process with this plugin it will always highlight the arrows for us. So the benefit of using the

Testing tools and yes linting ⁓ configurations is that it can enable the earlier detection, consistent enforcement and better team awareness. And next we can talk about the semantic HTML. ⁓ semantic HTML isn't just for screen readers, it's a foundation of maintainable and meaningful front end. The schematic architecture has clear document structure, ⁓ it's a correct component.

semantics. It has well defined area rules, the logical fruits order and the proper leave regions for dynamic updates. So for the card component we could have this HTML structure. With the CSS help ⁓ it might render as correct ⁓ on the surface in the browser you cannot tell the difference. But a semantic version of the implementation will be something like this. We'll have articular ⁓ that means a isolated area for card.

And the article has a R labeled by ID and this ID is the H3 with a heading and then is a title. And with this setup, screen readers can announce the card content meaningfully. You can also improve the focus management like ⁓ you get the next card when you deleting the current one. So you will do some search and found the next card on the UI and set the focus to the next ⁓ card on the column.

When the item is deleted, fricks automatically move to the next card. So there is no dead ends for keyboard or screen reader users. You might also want to check the IRA labels for interactive elements, ensure all images has alternative text, fixing color contrast issues, adding skip links for keyboard navigation, and verify visible focus indicator everywhere. For example, in our application, when I tap, you can see that the focusing ring on the avatar.

Sneha Mehra (00:07:29)  
and ⁓ tab again it will be focusing on the ⁓ search and the bird view list view and I can use the keyboard to change that order and I can shift tab to go back to the board view and then tab again to do the refresh tab to the first card and the first avatar assignment ⁓ okay set there's something wrong now I do a tab again ⁓ it can go through the card and I can also

adding a new card, new card and tab to add a button and I can space to add it to the column and tap again ⁓ things like that and that introduces us to the next section keyboard accessibility. So it is not an option. It's critical part of the front end architecture, especially in ⁓ complex applications. The principles here are the same ones used in professional grade systems like Jira ⁓ or

Notion, keyboard shortcut for power users, logical tab order and focusing states, escape patterns to closed models, ⁓ arrow key navigation between items, things like that. This will make your application feel fluent and ⁓ professional. It's faster for power users and also usable for everyone. ⁓ So for example I'm currently at the Jira Board View. I can use a tab to ⁓ do a lot of things.

like ⁓ going to summary and I can also go into timeline ⁓ and ⁓ doing all kind of things and if I press command key I got a panel that I can type quickly to things like ⁓ the one is go to summary I can go to summary ⁓ here it will be much faster I just simply press the comment key and if I want to go to the calendar view I just press four

And it can go to ⁓ the column view. And lastly we can use some DevTools ⁓ during our development process to verify the accessibility issues very quickly. For example, we can always use a lighthouse in Chrome ⁓ DevTools and we can ⁓ check the accessibility part particularly and run an analysis page load and it will give us a better understanding of the overall ⁓

Sneha Mehra (00:09:57)  
accessibility analysis. We can do that with our application now. Yeah it looks like our application is pretty accessible and it has all these categories. ⁓ If we have some ⁓ violations you can also get the ⁓ category and analysis and the diagnosis and the items that you need to take care of but at the moment it looks pretty good.

And you can install a few other tools like X plugin in your ⁓ browser and it will give you a quick scan and give you a better understanding of how the performance ⁓ goes in your ⁓ application.

Sneha Mehra (00:10:40)  
So accessible design is good design. It's not a checklist you complete at the end. It's a mindset that shapes every component and interaction. When you design with accessibility in mind, you don't just help a small group of users. You make your system stronger and more reliable and more inclusive for everyone.

—--------------------------------

Sneha Mehra (00:00:00)  
In this lesson we are going to talk about something that often treated as an afterthought in front end architecture. Accessibility. Accessibility isn't just about compliance checklist or adding you know alternate text. It's about designing systems that work for everyone, regardless of ability, device or context. Good design is accessible design. We're going to talk about some common tools, practices.

⁓ in front end applications and using our board application as a demonstration again. So the key idea of designing a accessible system is simple. Don't retrofit accessibility. Build it in from the start. When you design inclusively you automatically create better products. Accessibility improves ⁓ usabilities for everyone, not just users with disabilities. Here is why it matters. So many regions now

legally require WCEG 2.1 AA compliance. ⁓ Accessible application can reach up to 50% more users, which is good for your marketing or your ⁓ profit. And third, semantic mockups improve SEOs and secret reader support. And also accessibility patterns make interaction cleaner for all users. And they feature proof your design across devices and the input method.

Sneha Mehra (00:01:29)  
Let's work through the foundation patterns that makes a accessible system. Progressive enhancement. Start from a reliable baseline. Your app should still work if JavaScript or CSS fails. Then layers ⁓ on richer ⁓ features progressively. This approach is draws gracefully a degradation. ⁓ Semantic HTML foundation. Use the right elements for the right purpose. Buttons for actions, ⁓ nav for navigation.

Main for content. That structure gives meaning to a UI, helping assistive technologies ⁓ interpretive it correctly. Everything must be usable by keyboard. Set the logic tab order, manage forks carefully, and make sure virtual forks ring or indicator are always visible. And last but not least, automate accessible testing. Automate checks catch common issues earlier and teach developers good patterns over time.

Tools like a JestX and YesLint plugins should be part of your ⁓ toolbox or your normal workflow.

Sneha Mehra (00:02:40)  
So let's first add automated accessible checks directly into our development workflow. So in our board application, I have already added just X, ⁓ which is a tool that will enable us to run automatic checks in the unit test. ⁓ so for example, if I go into my card test, we have defined a few ⁓ tests here. So basically we're still doing doing the render with the provider just like before. ⁓ the

Difference here is that we can use X ⁓ to render. It will automatically check if the container follows the accessibility rules. And it has a customized ⁓ assertion. ⁓ We can do that because we have the test setup. ⁓ We are extending the expand with this to have no validation. ⁓ And that assertion is from ⁓ just X. So with that, if we run a test ⁓ in our local

We can see it's all passing, but if we let's say go to our card component, ⁓ let's make it ⁓ fail by introducing some incorrect pattern here. So for example, list item should be used inside a list container or list, right? You normally do ul ⁓ things like that, A, B, and C, right? This is ⁓ correct usage of the

UL L I and that's fine if we run the test it will pass in the accessibility check. But if we use the ILI element without a UL or OL as a container, ⁓ it will ⁓ fail because ⁓ obviously you should use that with a you know container element. So now if we run the test again, you can see that we have this failure. It expects the HML ⁓ to have ⁓

No validation but received ⁓ this element must be contained in the UL or OL ⁓ list item. So this is the validation. And because we have this check in our test, we can always catch that before it's slipped into the production. And we can remove this. And next we can integrate the yes lint rules ⁓ with these ⁓ accessibility checks as well. So in our yes lint config

Sneha Mehra (00:05:03)  
I have already installed a package called YesLint plugin GSX Accessibility and I'm applying this plugin to all the files with TS, TSX, GSX, GS. And the rules are ⁓ it has some recommended rule ⁓ following and then it has some automatic rules as well. And during our development process with this plugin it will always highlight the arrows for us. So the benefit of using the

Testing tools and yes linting ⁓ configurations is that it can enable the earlier detection, consistent enforcement and better team awareness. And next we can talk about the semantic HTML. ⁓ semantic HTML isn't just for screen readers, it's a foundation of maintainable and meaningful front end. The schematic architecture has clear document structure, ⁓ it's a correct component.

semantics. It has well defined area rules, the logical fruits order and the proper leave regions for dynamic updates. So for the card component we could have this HTML structure. With the CSS help ⁓ it might render as correct ⁓ on the surface in the browser you cannot tell the difference. But a semantic version of the implementation will be something like this. We'll have articular ⁓ that means a isolated area for card.

And the article has a R labeled by ID and this ID is the H3 with a heading and then is a title. And with this setup, screen readers can announce the card content meaningfully. You can also improve the focus management like ⁓ you get the next card when you deleting the current one. So you will do some search and found the next card on the UI and set the focus to the next ⁓ card on the column.

When the item is deleted, fricks automatically move to the next card. So there is no dead ends for keyboard or screen reader users. You might also want to check the IRA labels for interactive elements, ensure all images has alternative text, fixing color contrast issues, adding skip links for keyboard navigation, and verify visible focus indicator everywhere. For example, in our application, when I tap, you can see that the focusing ring on the avatar.

Sneha Mehra (00:07:29)  
and ⁓ tab again it will be focusing on the ⁓ search and the bird view list view and I can use the keyboard to change that order and I can shift tab to go back to the board view and then tab again to do the refresh tab to the first card and the first avatar assignment ⁓ okay set there's something wrong now I do a tab again ⁓ it can go through the card and I can also

adding a new card, new card and tab to add a button and I can space to add it to the column and tap again ⁓ things like that and that introduces us to the next section keyboard accessibility. So it is not an option. It's critical part of the front end architecture, especially in ⁓ complex applications. The principles here are the same ones used in professional grade systems like Jira ⁓ or

Notion, keyboard shortcut for power users, logical tab order and focusing states, escape patterns to closed models, ⁓ arrow key navigation between items, things like that. This will make your application feel fluent and ⁓ professional. It's faster for power users and also usable for everyone. ⁓ So for example I'm currently at the Jira Board View. I can use a tab to ⁓ do a lot of things.

like ⁓ going to summary and I can also go into timeline ⁓ and ⁓ doing all kind of things and if I press command key I got a panel that I can type quickly to things like ⁓ the one is go to summary I can go to summary ⁓ here it will be much faster I just simply press the comment key and if I want to go to the calendar view I just press four

And it can go to ⁓ the column view. And lastly we can use some DevTools ⁓ during our development process to verify the accessibility issues very quickly. For example, we can always use a lighthouse in Chrome ⁓ DevTools and we can ⁓ check the accessibility part particularly and run an analysis page load and it will give us a better understanding of the overall ⁓

Sneha Mehra (00:09:57)  
accessibility analysis. We can do that with our application now. Yeah it looks like our application is pretty accessible and it has all these categories. ⁓ If we have some ⁓ violations you can also get the ⁓ category and analysis and the diagnosis and the items that you need to take care of but at the moment it looks pretty good.

And you can install a few other tools like X plugin in your ⁓ browser and it will give you a quick scan and give you a better understanding of how the performance ⁓ goes in your ⁓ application.

Sneha Mehra (00:10:40)  
So accessible design is good design. It's not a checklist you complete at the end. It's a mindset that shapes every component and interaction. When you design with accessibility in mind, you don't just help a small group of users. You make your system stronger and more reliable and more inclusive for everyone.

—--------------------------  
Sneha Mehra (00:00:00)  
Last time we talked about why real time matters. Now let's build it. I will show you what changed on the client and a tiny express ⁓ endpoint on the server, how the two talk to each other and what to watch out for in production. Previously we fetched the data, made a change, and waited for the next fetch to catch up. If a teammate assigned you to a card in their screen, your screen stayed stale. And now we open one long lived connection to

API board events. ⁓ whenever a server has news like card assigned, it pushes a tiny message down the pipe and our UI updated immediately. So for example, I now have two tab opened. Let's say one is my teammate's screen and one another is my screen. So my teammate has ⁓ seen the same thing ⁓ initially, and when they assign the second ticket in the list to someone else.

So currently this assignment is Charlie. Let's change that to ⁓ Sam. And when I click this one, you notice that Sam is changed or like shows up on the other ⁓ tab as well. So similarly, if we change this one, the third one to another person, let's say Hannah. And you will notice that the Hannah is shows up on the ⁓ left hand as well on that screen. So this all happened in real time.

And our new latest change will be broadcast to all the screens that listen to this board. So that's all about the real time updates. So if you remember that we in the last video we talked about three different approaches to implement the real time updates. The polling, service and events, and the web circuits. In this lesson we're going to use ⁓ SSE ⁓ to implement these real time updates. So basically in front end we will need to listen to a endpoint.

that has this capacity to broadcast, ⁓ we'll need to implement a ⁓ event source. The URL for the resource is API board the board ID and events. So that means in the back end we'll need to implement this endpoint. ⁓ And ⁓ that way we can connect to that endpoint and then get notification once we have some new data in that channel. And then ⁓ we will add a event handler to the event source.

Sneha Mehra (00:02:27)  
we are listening to this particular customer event, the card assigned, and we will have that event, so that will be some data carried ⁓ in the event. So we'll ⁓ pass the data into the object we can understand and then call the API or the API in the store ⁓ in our context and ⁓ to ⁓ make the change in the front end. So the UI is refreshed. And to make that happen we will need to make some change in the back end

So currently we are using MSW to mock the net network ⁓ in our browser, but because they are my MSW process running inside the browser, inside that tab, or inside that browser process, it cannot talk to another process directly. That means we will need to use the real server side code. I have moved the mocks into a server folder. So basically it's a express an endpoint. And this server running in my local 4000 ⁓

And if I want to get something from that server, we can simply do a lochost four thousand API board ⁓ one. And then we can format the output into a beautified JSON. And we can see the response ⁓ pretty much like what we saw in the JSON file. And for users we can get that as well. We can get all the users. this is the first page. you know, we can use the REST for API ⁓ as

what we do in a browser like ⁓ offset equals to ⁓ five ⁓ and ⁓ limits equals to ⁓ five as well. And ⁓ that will give us the next page. So basically we have this mock server running in my local on port four thousand. And in the front end configuration we have a ⁓ proxy that will dispatch all the requests sent to slash API to

The 4000 server instead of the 5173 and the mocks folder will be ⁓ not used anymore. Awesome, so we now have both the front end and back end up and running. What I have changed is I have added a event emitter ⁓ object ⁓ to the server. So essentially event emitter is a event listener or pop sub kind of implementation. basically you can register a

Sneha Mehra (00:04:51)  
event and the event handler by using the on interface and you can unregister a event by using the off and you can trigger a event with the data with the image function. So basically once we established a channel we will add a few ⁓ event handlers to a particular event we want to listen on. For example we want to listen on card assigned event. And when the card API is head

We will emit this event. And at that point we can call the callback in the listener. So for example, if we go to the index.js, which is a mock server, and if we go to the ⁓ event ⁓ endpoint, this is the new added endpoint. And ⁓ we will get the board first the parameter and then we will write header to the response. So basically the most important part is the connection here, the keep alive connection.

That will make sure we don't close this connection. We will use this channel all the time. And when something happens, we can use this channel and send message to the front end so they can update their UI correspondingly. And then we define a function called sendEvent. So basically that we use the response ⁓ channel to write the event, event type and the data that associate with the event. And this is the protocol that we communicate with the client side.

And once we have this connection, we will send the event to the client side ⁓ that says ⁓ the board is connected. And it's up to the client side to ⁓ either ignore the event or you know doing something interesting with that connected ⁓ event. And here the most important part, the mock event emitter, which is the instance that we created ⁓ of our ⁓ implementation here. So we basically define a class that doing the ⁓ map and it can ⁓

you know, regist, unregist, and ⁓ unnotify ⁓ all the listeners. And in our index we have a one single instance of the event emitter. So that means all the client ⁓ all the browsers like ⁓ my browser instance or my colleagues ⁓ browser instance they all using the same instance. So that way they can register themselves as a listener to our notification.

Sneha Mehra (00:07:16)  
And then when something happened, we can always use this single instance to notify all of them. And if I go back to the event endpoint, and in here we will register a card assigned event handler. ⁓ that means when this card assigned event happened, we will call the handle card assigned ⁓ handler, which essentially will send a event. ⁓ the event name is or event type is card assigned, and the data is whatever passed from the

caller place which is the emate function. So this is only doing the registration or like a listening, but nothing happened at this stage. And what triggers the card assigned ⁓ event is in our API slash card slash ID. So this is the existing ⁓ endpoint and once we find the card, you know, doing the ⁓ assignment and only at that point we call the emate ⁓ function

That will trigger the actual event. The event type is card assigned. And the data attached to the event is the new card object, which has the the card itself, it has a board ID, and it has a new assignee. ⁓ and then in the front end we can get this data and unpack it and so we can use it to update our UI. So this is how it ⁓ all connected. So this is the back end change. We have this image change in our ⁓

assignment endpoint. So if we go to the board page, ⁓ so basically we'll need to define new USF block. It will ⁓ connect to the API slash board slash b ID slash events. And on open we can do something interesting but we are not doing anything interesting here. And for the card assigned event and that's what we just saw in the back end, we'll add the event listener. Whenever the back end trigger the card assigned event we can get that data

And ⁓ this callback will be ⁓ called. So we will unpack the data first and there is a card ID and assignee in a ticket and we will check if the user exists. If it does, we will absert the user to the local store. And ⁓ we'll also update the card with the card ID that returned from the event, and also we will update the ⁓ assignee ID to this ⁓ user ID. And this handles the case that we are removing a

Sneha Mehra (00:09:43)  
assignee from a card. And then let's go back to our board and ⁓ let me do a refresh and in this ⁓ screen you can see the board as before. In the network tab we have this event endpoint. You can see in the event stream there's already something happened that is ⁓ connected type and the data is board ID one which is sent out by the backend. So now if we trigger a event in here, so for example if I

assign the user to Alice and you can see there is a card assigned event triggered and the payload is exactly what we are expecting. It has ID, title ⁓ and assignee which is a user object from the back end. And then we get that data we can do the ⁓ UI updating. And this event is not visible only to us. It's all the client. So for example I have a few tab opened

If I open this one, even here, and ⁓ you can see the board is here. So if I now ⁓ change the assignment to Charlie, and you can see the ⁓ card assigned event happen in this Chrome screen and also the one underneath, you can see that event is being triggered in here as well. And all the UI will be updated correspondingly. So it works fine, but that are the few things to keep in mind when you use the SSE. ⁓

So, firstly, we'll talk about the direction. So, SSE flows from server to browser. So, when the browser needs to talk back, you need to use a normal fetch or a post. And if you really need a two-way ⁓ communication, you will need a web socket ⁓ for these keys. So, staying alive is ⁓ another issue. Some proxies close idle connections. So, we need to ⁓ hardbeat every a few seconds to keep that connection alive. And another limit is

Connection limit. So under HTP ⁓ one point one or one point zero, browser can only allow a handful of open connections per site. About six. So if you have several stream ⁓ or multiple tabs ⁓ that can easily hit the keep. ⁓ but under HTP two streams are multiplexed on one connection and the server and the browser can negotiate ⁓ how much the limit should be. So the keep is

Sneha Mehra (00:12:08)  
normally not a big issue anymore. So you will always need to keep one shared event source for your application and carry multiple event types over it.

—--------------------------  
Sneha Mehra (00:00:00)  
Welcome back. In this lesson, we are going to cover building confidence in our system by introducing tests. Testing is how we catch problems earlier, reducing release anxiety, and move faster with less fear of breaking things. Before we dive into setup, let's start with a quick mental model that guides how we structure our tests, the test pyramid. A test pyramid is a simple idea. At the bottom, we have many small fast unit tests.

In the middle, we have fewer integration tests, and on the top, we have a small number of four end-to-end tests. Each layer gives us confidence at a different level and cost. Unit tests are cheap and immediate. Integration tests verify the components work together, while end-to-end tests are slow but simulate real user flow. When we get this balance right, we spend less time debugging and more time shipping. So for unit tests

It's a foundation. A good unit test isolate a small piece of logic and verify its behavior. But modern unit tests are often integration style. We still use the real code we wrote, like our provider or hooks. ⁓ We only mock things outside the system such as API codes or the browser APIs. That gives us faster, reliable coverage without making our test brittle. Well, integration test connecting the pieces

⁓ together and they check multiple parts of the system work together. For example, a form that updates a list ⁓ or a component that renders data fetched from a hook and so on. Those tests runs on the simulated environment by keeping ⁓ most dependency real. They are slightly slower than pure unit test, but they catch the kind of varying errors that only appear when the parts interact.

Well the Anton test is the top of the pyramid. Those run the app in a browser interact like a real user and hit the server for real. They are powerful but slower, so we keep them focused on the most important flow like login, the checkout, and the architects ⁓ manipulated cards on a board. Testing is definitely a very deep topic, one that really deserves its own book. In fact I have written one ⁓ book on this topic.

Sneha Mehra (00:02:26)  
Testing driven development with React and TypeScript. It's published by EPRES. ⁓ It goes much deeper ⁓ into how to design reliable, maintainable front end with TDD and with real world examples, patterns, and workflow. If you'd like to explore testing beyond this course, you can find it on Amazon or on the Epress homepage. Alright, back to the lesson. We'll start from the bottom, ⁓ the first layer, and add integration still.

unit tests using white test and the testing library. So why do we need this layer at all? As our front end grows, refactor becomes common. Without a safety net, a small change can break behavior in unexpected places. Unit tests give us quicker local confidence. We can fix the bug once and prevent it from ⁓ returning. So first we'll need install a few packages. The white test configure JSDOM for realistic rendering

Add a few browser API streams and test the card component in our board application ⁓ using its real provider. We will verify that it renders correctly following accessibility best practices and behaves ⁓ properly when deleting a card, including a failure handling when the server rejects the request. Even in our unit test, we will focus on the behavior. Traditional unit tests often mock everything: the contacts, hooks, utilities.

They look neat but break when you rename a function or change internal wiring. Instead we test the behavior, not the implementation details. We render the component within its real context, interact with it like a real user would, and assert on visible outcomes. And we mark only what's outside of our control, like the network request, analysis ⁓ or brother APIs ⁓ that we don't have in the GStor. So why this matters for production?

Because those tests fail when behavior changes, not when code structure changes, they are faster, reliable, and provide an earlier warning system for regressions. This is the most cost-effective layer of our test pyramid. So that's enough theory, let's dive into the code ⁓ to implement the test in our board application. So firstly, we need to install a few packages. ⁓ We will install the white test, testing library React, testing library JSDOM, and the testing library.

Sneha Mehra (00:04:51)  
the user event and also the JSDOM that simulates the browser environment in ⁓ memory so we can test the full rendering results. So we simply install these packages and if we go to our code base, the first thing we need to do is invite the config.ts we need to define a new section. So we will need to set up the environment as JSDOM. It's basically a JavaScript implementation of DOM running headless. You don't have a real

Browser to run in our tests. And we want to export a few APIs to global so we can use a describe eight expect ⁓ functions in the test. And we also need a setup file. I have already defined a test setup TS in ⁓ source folder. That setup file includes small ⁓ browser shims for popovers and responsive components. And we don't want to scan or fan test in the following

Folders like the node modules or dist and so on. So we'll skip these folders. And let's back to our test file. It's going to components ⁓ test card test TSX. That's a pattern that the gest ⁓ deferred using and a test file should end with test.tsx and a good practice is using the same name as your component that on the test. So we are testing the card component, so it makes sense to have a

card.t test.tsx under the test folder. And I have added two tests here. The first one is that we are rendering the card with title and ID and we want to verify the content or the text and the ID is showing up on the screen. I want to verify it's not only in the DOM but visible. And the second test I want to verify ⁓ is we want to make sure the articular is used for rendering a card.

So all the details of the card will be rendered inside a articular. ⁓ And if we go back to our implementation, you can see we are using articular. So we want to verify that it's used correctly. It's a more semantic articular element. And once I have this all set up, and I can go to command line to run ⁓ test and it will find this test ⁓ and ⁓ run the assertions. If it's passing, it will report back to us like this.

Sneha Mehra (00:07:16)  
And it's waiting for the file to change. We can add in more test and the white test will detect the change and run the test again. And npm run test works because I have defined a script ⁓ in the script section of package.json. So when we run test, it will run a white test, which will use a white config to find our test and run against the code base. And back to the card test.tsx.

You notice that we are not using the render itself, we're using a render with provider. And we will pass in a UI, we'll wrap that UI with the query provider and the board provider, just as what we did in the application. But obviously we don't have to provide the real data, we just give it empty data. ⁓ And ⁓ also we are mocking the fetch because we don't want to fetch actual APIs in the unit test level. That means in the test, whenever we use the fetch to

Get anything from the server side, we will use the mock data when not it ⁓ sending the real request. Obviously the render is the simplest ⁓ scenario. We definitely need to add a few more to cover the interactions like the deletion ⁓ assign user and also the accessibility part as well. Alright, well I have already added a few more tests for different scenarios. Like the first one ⁓ is rendering and then we want to check the

assignee displayed correctly. So in first test we don't have assignee and pass n so we'll definitely show a question mark. And in next test we have assignee, we pass that to the card component and then we are expecting the ⁓ j to be showed as the initial of the user if the advocate URL isn't provided. And then we have some accessibility related test. For example we want to say the label text for open card menu

to present ⁓ on the card and also we want to make sure the card is articular, the heading is used correctly, the button used correctly. And then we have an interaction ⁓ that is deleting. And we want to find the menu first and we click that menu to open the pop up and then we click the archive card text and we click that button and we want to expect that URL is triggered with the delete method. And lastly we want to

Sneha Mehra (00:09:39)  
Check the delete failure gratefully. So we are marking the response of the server when we delete, we return 500\. And when we click the menu and click archive and delete, and we wait for the title ⁓ of the card is still visible to the users, and that means the delete is failed. And we just roll back that deletion.

Sneha Mehra (00:10:04)  
We now have a lightweight safety net ⁓ that runs quickly and give real confidence in behavior. It supports refactoring and reinforces good accessibility practices. So that's our foundation. We have built the base of testing pyramid, faster, realistic tests that verify real behavior. In the next lesson, we'll move up a layer and test our system that we user experience it end-to-end, from the browser to the server and back.

—---------------------------  
Sneha Mehra (00:00:00)  
When we first learn front-end development, data modeling doesn't feel that important. You grab some JSON, put it into a state, and run date. For a small or medium application, that often works fine. But as soon as your application grows, cracks begins to show. You start duplicated state, updating one part of the UI, doesn't update another. Or you realize that you cannot easily represent a feature.

the product manager just asked for. That's when you discover split model is not just the back end concerns. It's fundamental to front end architectures too. Take a chat application like Slack, for example. At first it's easy to see some ⁓ data models, ⁓ a user object or model, a message, maybe an attachment to a thread, but as you explore the domain, more concepts appear.

threads that group conversations, private groups ⁓ with restricted ⁓ membership, broadcast messages that cannot be replied to, archived messages that change how history is shown. None of these show in the first adjustment response. You only discover them ⁓ when you really understand the domain or the business logic. Or we can consider a online course platform. When you do the data modeling, you might start with

course, modules, lessons, ⁓ student and instructors. But that's not enough. A student isn't just a student. They are enrolled in specific courses. Courses might have coupons, expert dates, ⁓ or ⁓ a drip schedule where lessons unlock over time. You might need to track progress or issue a certificate. ⁓ And the deeper you go, the more the model reflects the reality of the domain.

And let's bring it back to our start project, a board application. It's similar to Jira board view. At first glance you can see some board, column, ⁓ card, and ⁓ soon you will realize that cards has assignees, the board has permissions, ⁓ column represent a workflow ⁓ stage. For example, in some cases you can move a card from backlog to in-progress.

Sneha Mehra (00:02:28)  
but you cannot move it back. That is controlled by ⁓ a workflow underneath and you can edit the off workflow and then you can enable that moving back action. And cards ⁓ needs comments or attachment, it has fields ⁓ on it. If we just nesting everything inside a race in state, the code quickly ⁓ falls apart. Instead we need a model that captures relationships.

Board has many columns, columns has many cards, cards belongs to users. Once we agree on that shared language, the code gets easier to design, the logic gets cleaner, and bugs become rare. So here is the key point I want you to take away. Data modeling isn't just about ⁓ shaping objects in JavaScript. It starts with understand the domain itself. In practice, this means before writing code, ⁓ ask

What's the process end to end? Who is doing what in the system and how do they do it? What rules and exceptions do we need to capture? That's how you move from obvious nonce like user or messenger to deeper concepts like enrollment, threads, or drip schedule. For front end engineers, there's one more dimension. Our models ⁓ should reflect how the UI actua consumes the data.

APIs shouldn't just dump everything up front. They should be shaped by the access patterns of the AUI, what data is shown when and in what context. That's why, for example, for rendering a board, we might want to nested the structure, but for updating a single card, we might ⁓ normalize everything into entity tables. The shape depends on how data is used. To make this more concurrent.

in next lesson we'll look into a simple sidebar navigation example. At first, the UI decided which feature to show based directly on user's plan type. That works initially, but it doesn't scale well. And we will use the case study to explore how data model helps us ⁓ move business logic out of the front end, let the API return clear entitlements.

Sneha Mehra (00:04:49)  
And ship response around the UI's accura access and patterns. By the end you will say how even a small shift in modeling can make a application more consistent, scalable and maintainable.

—--------------------------

Sneha Mehra (00:00:00)  
When we talk about performance in front end applications, we usually think about speed in absolute terms. How quickly does the server respond? How fast does JavaScript run? But there is another dimension that just as important: the perceived performance. This is about how fast your application feels to the user, regardless of what's actually happening under the hood. Today we're going to look at the four common techniques that can dramatically improve.

perceived performance. Skeleton screens, optimistic updates, prefetch and server-side rendering. Let's start with skeleton screens. Instead of showing a spinner while content is loading, we render placeholders that matches the final layout. For example, if we're loading a list of articulars, you might show a few green rectangles where the headlines and images will appear. The benefit here is psychologic. Users feel like

Progress is happening immediately because the structure of the page is already visible. It feels faster than starting at a blank screen or a spinner. In React, you might use a simple component that renders a few green boxes with a pulse animation. Then when the data arrives you replace the placeholders with the real content. So while the load time hasn't actually changed, the experience feels smoother and faster.

Next, let's talk about optimistic updates. We have already covered this ⁓ in details in the previous module. So this is a technique where you immediately show the result of the user action in a UI before the server has confirmed it. Take posting a comment as example. Instead of waiting for the server respond, ⁓ you add a comment to the list right away. In the back end you send a request to the server. If this stays ⁓ great, nothing changed.

And if it fails, you just roll back to and show an error message. This creates a snappy experience. The user feels like the application is responding ⁓ instantly, even though the server round trip is still happening. So that's like a reactive query or relay make optimistic updates ⁓ easy to implement. But you can also roll your OR logic and local state ⁓ and rollback handling. The trade-off here is that you need to

Sneha Mehra (00:02:28)  
Handle arrows gracefully, but the payoff is huge, use the feel like your application is lightning faster. Another simple but powerful trick is Profetch and Hover, ⁓ which we will cover in next lesson in details, but here is how it works. When the user hover over a link or a button, you start fetching the data in the background. By the time they actually click, the data is already cached, and then the next screen loads ⁓ almost instantly.

Think about how you browse a product listing. You hover over a product card, ⁓ maybe thinking of click ⁓ in tune it. If the application starts preloading that product detail on hard, then when you do the click, it feels instant. Frameworks like Next.js support this out of box with next ⁓ link and its per fetch option, but you can also implement it manually. For example, calling the query client.fetchquery.

In React query, ⁓ when the link ⁓ gets forks on or hover. It's a very small touch, but it can transform the experience of navigating your application greatly. Finally, let's visit server-side rendering, which we have already covered in details. ⁓ but it's a very good technique for enhancing the perceived performance. So with SSR, the server generates HTML for the page and ⁓ sends it down to the browser.

That means users say something meaningful almost immediately, even before JavaScript has finished loading and hydrating. From a perceived performance perspective, this is powerful. Instead of saying a blank page, which we have already said in last lesson, user can see the content they are looking for right away. And this is common used in the product page. ⁓ You want user to see the product details immediately for SEO and for experience.

Interactivity comes a ⁓ fraction later, but the perception is that the page loaded fast. And of course, just like other patterns, SSR comes with trade-offs, there's more server-side ⁓ complexity and need to handle hydration carefully. But when downright, it's one of the best tools you can use for improved this perceived performance. So to recap, skeleton screen gives user an immediate sense of progress.

Sneha Mehra (00:04:50)  
optimistic updates make interaction feel instant, while perfetion on hardware the weight ⁓ for the next screen, and sufficient rendering ensures the user says something right away. So together those techniques don't just make your application faster in absolute terms, they make it feel faster, which is what the user actually notice. In front end system design, perceived performance is just as important as the real performance.

If your app feels smooth and responsive, user are more likely to stay engaged even if the backend still takes a second or two to respond.

—------------------------

Sneha Mehra (00:00:00)  
In this lesson, we are going to look at one of the most effective ways to make your application fails faster. Optimistic updates. Instead of waiting for the server before showing the results, we update the UI right away and then reconcile when the server responds. It's a small change in mindset, but it makes a huge difference in how smooth your application fails. Normally when you update something, maybe assign a task to a user ⁓ or deleting a tag from the board.

The floor looks like this. You click the button, the request goes to the server, and only after the server replies, maybe after two seconds, you update the UI to show the change. It works, but the user is stuck ⁓ waiting during that round trip. Even a short delay feels slow. Multiply that by lots of small interactions and your application feels sluggish. And that's when optimistic updates kicked in.

Optimistic updates flips that order around. We update the UI first, assuming the server call will succeed. And then if the request works, we don't have to do anything else. And if it fails, we just root back to the old state and let the user know ⁓ that something went wrong. The benefit is instant feedback. The app feels like it's keeping up with you. Let's start with a simple example. Assigning a task ⁓ to a teammate.

So in last lesson, we have already seen this feature that when you assign a user ⁓ to a ticket, ⁓ let's say Charlie, it will happen instantly. That's because we are doing ⁓ the local ⁓ mock and there is no delay. But in real world, the network might be very slow. So let's mimic this behavior by simply update our handler in the card assignment. So in here we can

⁓ add a extra a wait for let's say delay a second right and if we go back to our board when we assign a user to chart again and you can see that the pop-up show hangs on the screen when I click the chart here. So let me do this again. So when I assign user to stem and you can see this pop-up will hang there for a little bit like one second. ⁓

Sneha Mehra (00:02:26)  
during this ⁓ request. It's definitely a small thing, but just imagine that how we s let's say switch to the list view, add a few delays, moving cards across columns, let's say I move this one to this down column and they're adding a few more seconds. So here and there that these small details can make your application over or feel s slow. And that's something we want to avoid.

So let's say we want to change the ⁓ user experience for assigning a user to a card. Let's go to our card component. So currently we have this handle assign user. ⁓ We will set the updating to true and then we'll send the request. ⁓ And this request will have one second delay. ⁓ We just add this change in the ⁓ mock.

And only when we get the response and we see that successfully ⁓ processed, we will update the UI and then we will close the pop-up. And ⁓ this is a quite normal process. So for ⁓ a optimistic update, we will do this update ⁓ first and before we send out the request. And the very first step will be we need to remember the current ⁓ assignee.

So we need to declare verbal ⁓ let's see previous ⁓ assignee ⁓ is the assignee. ⁓ we will need to remember that one and then we will move this part ⁓ at the beginning of the handler function. And then if it's success, ⁓ we don't have to do anything. And if it's failed, we will need to revert this process.

in here basically we will need to ⁓ use the the previous assignee ⁓ to ⁓ update the user with the previous assigning here ⁓ and ⁓ in ⁓ the update card as well. So this way if that is a previous assigning we will ⁓ update the update user like a ruleback and then we will need to update the card as well and this way it will fuse faster let's have a look at the UI

Sneha Mehra (00:04:48)  
⁓ if I go to take it ⁓ one and give it ⁓ Charlie again. So you can instantly it close off the pop-up and the user after change instantly and during this updating there is a small opacity change but that's okay because we have already updated the UI and close off the ⁓ pop up and you can see that immediately it enhanced the ⁓ user experience and that's a happy pass. So what if

that is something went wrong when we updating a user assignment here. ⁓ So let's go to our handler and say if the ⁓ ticket ID is ticket two for this one we'll need to return ⁓ a ⁓ 500 like ⁓ internal server arrow.

So for ⁓ only for ticket two we will throw an arrow and then if I go back to our UI for the ticket two, if I assign it to ⁓ someone, it will throw and you see that that's a flash kind of thing. ⁓ it's assigned to ⁓ Sam, the avatar changed, and then after a second it drove back. So that means something went wrong. So normally we will need also to have a notification pop up

in the corner to see that okay we didn't make it something went wrong and you can check the ⁓ log or ⁓ retry. And as you can see that only a small change can make your UI ⁓ feel faster even the network is slow. Now let's look at the deleting which is a little bit trickier. As for deletion we will need to not only remember the card detail but also the ⁓

position or location of that ticket. For example, if we want to delete this one, we need to remember the card detail for the card ID, the description, the avatar, and so on. But we also need to remember ⁓ the position ⁓ of that ticket in the column and if is something went wrong, we will have enough information to roll it back. And then I made some change in the delete ⁓ in the API level. So we still adding a some delay to make it

Sneha Mehra (00:07:13)  
⁓ more realistic and for it take the one we will return a 500 that will indicate the deletion is failed and for others it just ⁓ works as before. And in the card card component we will set the menu open to false which is close off the menu ⁓ about the deletion. So basically that means this menu will be hide as we click the delete

And also we will get the data from the state and remember the card details and the location. And then we'll try to remove this card from the local store. And then we do the data fetching, ⁓ which is basically ⁓ send the request to the backend and then we will check if the result is successful. If it's not successful, we will try to restore the card. And ⁓ the restore and the find card location basically is what we just knew, ⁓ added

⁓ function that will help us to insert the card back to the original position. Now if we look at the ⁓ UI ⁓ and ⁓ if the the ticket one will fail, let's see the ticket two, if we archive it, ⁓ it's just immediately gone without the delay. And ⁓ for other tickets as well it will super fast. even there's a delay there.

and for these drawback keys, let's open up the console and if I delete this one and and after one second you can see a failed delete ticket one, so that ticket is added back. So what do we actually get from optimistic updates? First, the UI reacts instantly, ⁓ users don't feel the lag, they see the result straight away. ⁓ That makes the app feel faster and smoother even if the network is slow.

And in apps where people make a lot of small changes, it just feels much more natural. But to make it work well, there are a few things to keep in mind. So always keep a copy of the older state so you can roll back if something fails. That way the user isn't left with the incorrect data. And if two users are editing the same data, you might run into conflicts. A common fix is to refetch or revalidate after the server call succeeds, so your UI gets ⁓ back in sync.

Sneha Mehra (00:09:35)  
For destructive actions like deletion, don't just feel silently. If it doesn't go through, ⁓ restore the item and show a clear arrow so the user knows what happened. And finally, if the same piece of data shows up in different parts of your app, make sure your state management updates them consistently. Tools like a React query, relay, or even a shared context as we are doing in our application can help with that. So to ramp up,

Optimistic updates are one of the simplest ways to make your application feels faster and delightful. We update the UI first, assume success and drawback if we have to. They're perfect for quick changes like assign a user, inline editing, add an item to a favorite list, add a comment to a post, and so on. But optimistic updates only cover what you do as a user.

In next lesson we will flip it around and look at what happens when the server sends updates back to us using real time updates with server send events.

—-------------------------  
Sneha Mehra (00:00:00)  
In this lesson we are going to make our application feel more interactive and enjoyable to use by adding drag and drop. Right now our board works but moving card between columns take extra clicks. With drag and drop, users can simply grab a card and move it where they want. It's faster, smoother, and feels much more natural. We will use Pragmatical Drag and Drop library here, keeping the implementation simple.

Focusing on the user experience, not on ⁓ reinventing the complex frameworks. And once we get the basic drag and drop working, we will take it a step further, making it accessible. Because a great user experience isn't just about animation, gestures, it's about making sure everyone can use the application, including the keyboard users and people who rely on assistive tools. So in this video I will walk you through the code I have already written.

Explain the key parts and show how it all come together to create a smoother, more inclusive experience. Let's dive in.

Sneha Mehra (00:01:09)  
So before we can start implementing the drag and drop in front end, we will need some preparation work to be done first. ⁓ we will add a back end API first. That API will enable us to actually move the card in the database. ⁓ obviously we are mocking the database, but you get the idea. And then we will in the front end board contacts we'll add a API to move the card in the local store or in the browser. And once we have this low level API ready, we can move on to the actual drag and drop.

So now if we go to our server side and go to the index RTS and I have already added a new API called API slash card ID move and in here we will need a few parameters ⁓ for us to locate the card to be moved. So we'll need the form column ID, the two column ID if we are moving card across columns. And also we will need the form index and the tool index if we are moving the card inside a column.

And once we have this parameter, we will find the column ⁓ the forum and two column. If the column doesn't exist, we'll just throw a 404\. And then we will find the card ⁓ in the column. ⁓ And ⁓ if we cannot find the card, we just return a 404 again. And then we will need to remove the card from the source column and then put that card into the targeting column, and then we return ⁓ the card we want to move or the we g the card we're moving. So

This is about a back-end ⁓ change, it's very straightforward. And we're not using database here, we're just using the mock ⁓ as before. And another change we'll want to make is in ⁓ board contacts, which is our front end store. So if we go to ⁓ the board contacts, and we're adding a new API called move card. So basically it does the same thing. We will need to find the the from column, to column.

And we will check like if that doesn't exist, which simply return the state. ⁓ And ⁓ similarly we'll need to handle the same column movement and ⁓ a different column. For same column we just update the form column. ⁓ Otherwise for the ⁓ you know cross column we need to update both the form column and the two column. And in here we just move in cards in the store. It doesn't have anything to do with ⁓ remote.

Sneha Mehra (00:03:37)  
And in the actual moving functionality we will call ⁓ both correspondingly. And at that point we can decide if we want to do the optimistic updates or we want to update the server first and then once we get the success we call the move card in the board context to make it happen in the front end. And that's pretty much about the preparation work.

Sneha Mehra (00:04:03)  
And now let's move on to implement that actual drag and drop operation in the front end. And we're going to use the pragmatic drag and drop library to implement the actual drag and drop. ⁓ It's a very popular ⁓ library ⁓ developed by Atlantic Design System team. I have using that in many projects. And the document is pretty good. You can read on the examples or tutorials ⁓ in the Atlantic Design System document.

And I have made a free course in my YouTube channel already. ⁓ If you are interested in ⁓ hands-on, you know, building the drag and drop ⁓ experience, you can have a look at this course as well. It has a few details videos ⁓ you can look into. And in here I'm not going to dive into all the details ⁓ 'cause that's not ⁓ the main point here, but I will cover some of the key points here. So let's go to our board application and in the board we have columns.

And column we are integrating the cards belongs to that column. ⁓ Remember that. And now I'm adding ⁓ a draggable card as a wrapper of the actual card. So draggable card is using the programmatic drag and drop and enable the drag and drop operation. So basically ⁓ for drag and drop you will need to ⁓ use effects. And the first use effect block is making sure our card is draggable. That means you can move it

And the other one is making your ⁓ card as a drop target. So that means you can move card on top of it. So you can accept the ⁓ event ⁓ or the data combined with the event at that point. So to make it a draggable is very simple. You basically need a element reference ⁓ that is the DOM element of the card, and then you can attach few ⁓ data on top of this element. So when you drop

We can get the data from this ⁓ initial data. Also, you can set these states and then use them to stale your component. You can rotate it a little bit and ⁓ to indicate it's moving. And on the other hand, ⁓ we want to make the card also a drop target. And the API is pretty straightforward as well. So we will need the element that we want to make a drop target. And ⁓ there are the few callbacks you can use, like the can drop.

Sneha Mehra (00:06:30)  
That means ⁓ can we drop a element on top of this card. And the most important callback here is the on-drop. So basically in here we can access the moving object and the target. The moving one is the source, that's the one you are dragging now. And the self here is the drop target, that means the the target element. So in here you can ⁓ get the source ⁓ card ID, column ID, index and everything.

And once we have this information, we will call the onMove callback, which is passed down from the column component. And the column component has a handle move event handler. And in here, we will actually call the context API to actually move the card in a local. So that will do the optimistic updates. And in the meanwhile, we will make sure the back end change also happens. And we will send a patch request to the new added move ⁓ API.

And we will prepare for all the you know parameter for the API and that will change the data in the backend database. And let's have a look at ⁓ how it works in ⁓ browser. So in the browser we have the ⁓ application up and running. We can drag this setup project structure and move that to down. For example, we can move it here and we release. That card is moved, and if we open up the console.

open up network as well. So let's say we want to move the ⁓ create the board UI back to ⁓ to do. Let's move that ⁓ and we can drop it beneath the configuration yes lint. And you can see that when we drop it will send the request to the move and that'll move in the ticket ⁓ number five. And because we are using optimistic updates, so before we actually move we just update the UI first and then we send the request.

It's great for us to be able to move the card around by just using the mouse drag and drop. That's great. But for people who not using mouse or they are purely rely on keyboard navigation, ⁓ it's impossible for them to actually move card ⁓ into different position. That's really bad. So we want to avoid that. We want to build a application that more inclusive and ⁓ it can reach out to more users. So that means we'll

Sneha Mehra (00:08:50)  
Provide a alternative way for users who rely on the keyboard to be able to move the card across the board. We can put some operation into the context menu. ⁓ That means the user can ⁓ use the keyboard to navigate to that button, they can focus on it, they can expand it, and they can use the m-up and down to select the menu, and then it can use the space to trigger the menu, then the card will be moved correctly.

To make that happen, I'm exporting one more ⁓ data from the contacts, which is the columns. ⁓ In the columns we can access all the columns, like the columns name. And in the card, we want to integrate through the columns to generate these menus dynamically. I will show you how it works. So your card component, we will access the columns from the contacts, and then we will generate a few buttons for each column.

And for the button we will define a handle move to column event handler and then we will set the column title as the ⁓ button ⁓ text so user can see what columns we are moving the card to. And for handle move to column callback, basically we will call the move card and ⁓ call the fetch, which is pretty much what we do for the mouse ⁓ drag and drop we just saw.

So now let's see how it looks like in browser. So now I'm going to use the keyboard only. I'm using keyboard to navigate to a card. Let's say I want to move the config yesnint pretty tier to the ⁓ you know in progress one. So I fix down the button and I use a space to expand it. And now I can you know navigate to different ⁓ columns or actions in the drop down menu and then I can press the space to move that to in progress.

And you can see, I'm purely using the keyboard. So let me move into another one. For example, this one. I want to move that to down. So the setup projective structure is done. That is great. Our application is now not only smoother, user can drag and drop, but it's more inclusive. They can simply use the keyboard to finish what they want to do.

—-----------------------  
Sneha Mehra (00:00:00)  
In this lesson we are going to add multiple pages to our application, things like your work page, board view, and settings. And we will do it in a way that works with both clients-side routing and server-side routing. We have already talked about lazy loading, code splitting, and SSR in earlier modules, and now we will simply apply in those same ideas at the page level, which is where routing really start to matter for performance and structure.

So we will use the React router ⁓ first to set up the multiple pages and then gradually we'll build the lazy loading for the pages. So why do we use the React router? ⁓ React router gives us a clean, declarative way to describe our routes, and it lets us navigate in between pages without full page reloads. But it also plays nicely with the server side rendering, which we have already covered earlier, which means the server can render the right page.

when someone lent to the URL directly. The tricky part here is making sure the server and the client are using the same routing setup. So let's look at how we solve that. So what we are going to implement is that the we have a your work page. When you click that it will lend to the your work and we have a list of boards and you can click one of them and go into the board detail. ⁓ For example this one. ⁓ I'm not setting up the

mock data for all the boards yet well reusing the same board. ⁓ But we can set up easily for each board we have different data. And then I'm adding another page called settings. ⁓ Basically you can go to the user editor and continue the settings and that will ⁓ land to the settings page you can customize user preferences as normal application. ⁓ But the purpose here is that to demonstrate the multi-page routing and lazy loading based on this setup.

So again, let's go back to the ⁓ your work. ⁓ we have this board list and we can go to let's see the first one. To make this happen, obviously we need the board list API. That's pretty easy to set up. In our server side, we are adding a boards JSON. And in the API we have these ⁓ API boards and we will return all these ⁓ boards data. So we have a ID, a name, description, and the card inside the board and the

Sneha Mehra (00:02:27)  
the last updates. So in the source code we are splitting this page by folder. So the board page has its own folder and we have your work. It has a separate ⁓ page and also settings has its own page ⁓ setting page. So firstly we will install the reactor router DOM and with that we will define the different router for the front end and back end.

So in client side we're using a browser router which understands the history API in ⁓ browser in the client side. And on the server side we don't have such history. So we'll use a static router instead. The static router, on the other hand, understands the URL. So it doesn't have the concept of history. So we'll wrap it the application around the static router. Let's have a look at the app. So instead of putting the router inside app.tsx, we keep the app

router agnostic, it just render roads as we can see here. basically when you ⁓ on the slash your work it will render your work page and if it's on settings it will run the setting page and similarly for the ⁓ board ⁓ it will render the board page by default it will run the the your work page. And this pattern keeps our roads defined in one place. Well each environment provides the rotor it needs

It's also exactly how Rector Router expects you to structure your server-side rendering. And now let's have a look at the individual page. For example, if we are on the router of slash your work, we are going to render your work page. And if you go to that page, it has the ⁓ use effect to get the boards list. And ⁓ if it's loading, it will show in a skeleton. So we have a skeleton here as well, a top bar skeleton and ⁓ the board list skeleton. And once we have this data ⁓

Fetched, we'll run the top bar and then we will enter it through the boards. Each of them is a link and it has a name, description, and ⁓ the metadata. And when you click that link, it will go to that particular board slash board ID, which will go back to the application.txt and matching the route here, which will render the board page ⁓ in standard. And then here is the old board page. Basically render all the you know the columns and cards.

Sneha Mehra (00:04:52)  
And similarly, if we go to the top bar and ⁓ we have a link that points to the slash your work, which will you know change the router and we can render the your work ⁓ page. And in the user profile, we have the another link that goes to the settings page, ⁓ which is what you see here. When you go to the settings, it will change the router to settings and this is ⁓ new setting page. So basically we have three pages now. The your work page

Board detail page and the settings. And right now each page fetches its own data ⁓ amount. That means the board page will fetch the board data. ⁓ Your work page will ⁓ get all the border list and the settings will fetch the personal information. We're not using SSR data fetching yet. The server just renders the shell. ⁓ It's simple, ⁓ it works with our existing code and ⁓ it's you know good enough for now. And later we can add in the server side per fetching to speed up the initial render.

And with this setup, if we want to test the server-side rendering, we run NPM run build. So we have these client side artifacts and the server-side artifacts. And then if we go to our browser and go to the 4000, which is ⁓ where the ⁓ express server runs, and ⁓ we can go to you know individual pages and ⁓ see the board detail and we can drag and drop the card around ⁓ or add it to the details by clicking the card.

And do some customization here, assign it to ⁓ some other ⁓ user, and then doing all this normal activity, you can add a card and so on. And that's how we set up the ⁓ multiple pages with the current code base. So currently all the pages are bundled into one single JavaScript bundle. That means the entry client is relatively big. It's around 433 kilobytes. Let's say we're on the border list page.

There's no point for us to load all the board details page JavaScript or the settings. ⁓ we should, you know, on the border list page only load JavaScript that are needed for the border list. And on the border details, obviously we don't need these two other you know two pages. So the page is a natural boundary for us to split the JavaScript bundles. But how can we do that with our current code structure?

Sneha Mehra (00:07:15)  
It's pretty easy. We have talked about the lazy loading and the code split in the previous modules, and we will use the exact same technique here to make that ⁓ split by pages. Let's go back to our code. So in the roads here, you remember that when you load the your work, we will go to the your work page. But we can make this ⁓ your work page as a lazy load component. So we will import it ⁓ with you know React lazy. So

In the bundling process, this will be created as a separate bundle. Similarly, for the board page and the setting page, they all will be split into separate bundles. And at runtime, when you go to that, your work, and only that bundle is downloaded. With this setup, if we run npm run build, you can see that for the client side ⁓ it has different pages. Let's say the board page has its own bundle.

settings page has its own bundle, your work page has its own bundle. Similarly for the server side rendering it has settings page, ⁓ board page, and ⁓ your work page. And you can see the entry client after build it's around 200 kilobytes and all the others are put out into established bundles. That's a great achievement. So let me go to the browser and ⁓ open up the network inspector. So a 4000 is where the Express server runs.

That's server side rendering. If we do a hard refresh, and as you can see, your work is the initial HTML. It has the ⁓ shell of the HTML and it's loading the entry client. And when the entry client is using for the hydration and your work page is loaded separately ⁓ here, there is also a user profile per bundle, which is another separate bundle. And if we go to let's say the board

detail page we click that and ⁓ you can see the very first JavaScript is that ⁓ board page bundle it will be loaded and if we go to the settings here it will be a settings page loaded at runtime and now if we go to the your work again ⁓ because the bundle has already been downloaded and it's cached by the browser so we don't have to re-download it again and all the following requests to for the server side rendering it will be relatively faster.

Sneha Mehra (00:09:38)  
And the functionality works just like before. We can move card around, you know, create a new card, ⁓ new ticket, like that. It's just ⁓ working as before. Or you can edit the card details just like before. But the bundle size is much smaller for the individual pages. So let's say you are in the board details page, you only download what you need to render the board detail. And similarly for other pages, you only download what needed to present to the user.

And also we define a suspense boundary that wrapped in the ⁓ individual page and we have defined a application layout which has the top bar defined. So the top bar will always ⁓ stick on the top of the page and only the children will change ⁓ when we change the router.

Sneha Mehra (00:10:29)  
With this setup in place, we can easily ⁓ scale the application without a big performance regression. So basically the increase will be linear. If we want to add a new page, we just define a new router here and we will lazy load that new page. And if that page is super big, it won't impact any other pages because we are bundling them separately.

—-------------------------

Sneha Mehra (00:00:00)  
In this part of preparing for production, we are going to address one of the most frustrating production issues in React applications, a component crash that brings down the entire UI. You have probably seen this before. Everything works fine. You click something and suddenly the whole page goes blank. No error message, no recover, no anything, just a white screen of death. This is what happens when a React component through an arrow.

and there's no safety net in place. In this lesson, we'll learn how to prevent that by using arrow boundaries. ⁓ That's a react built-in mechanism for catching random arrows and keep the rest of your application alive. We'll look at how they work, ⁓ why they are essential for production retines, and how to place them strategically for graceful recovery. So what every boundaries are and what problem they are resolving, think of error boundary

As a tri-cache for React components. In JavaScript we use tri-cache around fragile code to prevent one failure from crashing the whole program. Every boundary does the same thing, but instead of wrapping functions, ⁓ they wrap parts of your React tree. When the components throw during rendering, React by default unmounts the entire tree below that arrow. That's why you see a blank screen. And arrow boundary catches the arrow before it reached to the root.

Instead of unmounting everything, it renders a fallback UI, something like something went wrong or a retry button or close action that let the user to recover. It isolates the damage, the rest of the application keeps working. And the key idea here is content instead of collapse. So let's make it concurrent. So let's say we are in our board application and we want to assign a user to a card. So when we click the avatar, a pop-up shows up and we can select a user and so on.

But imagine that when there's something went wrong when we show the user select component or the pop-up and ⁓ the pop-up crashes. If we don't have any cache mechanism, that will break down the whole entire application. So in a card component we have a trigger button ⁓ and we will show the avatar and or if there is no avatar we show the initial. But for now let's say we have the ticket tool and we want to click that one and the pop-up will

Sneha Mehra (00:02:25)  
show a user select component if everything goes well. But in user select let's simulate a arrow keys here. So let's say if the page ⁓ is zero, we will throw ⁓ something like a ⁓ new arrow. Oops ⁓ page cannot be arrow.

Sneha Mehra (00:02:50)  
Zero. then if we go back to our page and now if I go to here, because when I click it, the page will definitely be zero, ⁓ it's a default value and it's throw. So you can see the whole application is gone. And ⁓ if we go to the console we can see that arrow is on caught and ⁓ in user selector component, we can see that console log

But for the user's experience, it's just a blank screen. That's very bad. So surely we need to handle this case. So let's build a simple arrow boundary class component. I have already defined a class component. That's probably the only case you will need a class component in React. So React requires a class syntax here because arrow boundary relies on left cycle method. So we have three important parts here. GetDivide state from arrow.

update local state when something fails. And component did catch ⁓ logs the arrow ⁓ to the console at the moment. ⁓ Or we can send it to Sentry or other libraries. And we have a render at the bottom. So the render decide if there is arrow we show a fullback UI. ⁓ Otherwise it will run in the children ⁓ normally. So that means when the arrow is thrown we can catch it in

the error boundary and then we set the state as arrow and then it will trigger a rerunder and then we check the has arrow and if it does have has a fullback we will use that fallback otherwise we will use the generic fullback ⁓ component and then ⁓ otherwise we just render its normal children. So that's a default keys. To prevent that user selector to crash, the current component see here I just wrap the user select with the error boundary and I give it

Define a fullback UI just to make sure it's fit to the setup. And when the user selects throw, we will simply fall back to this UI instead of crashing the whole application. So let's have a look at the behavior now. We go to the browser and then we do the same thing if we click the after and a throw. And then we ⁓ just fall back to a arrow UI. So the rest of the application is still usable.

Sneha Mehra (00:05:12)  
It's just showing this arrow. And ⁓ we can still like delete the card or adding a new card and a new card and so on. Only this part is through, but we can catch that and forget to a UI that is more gracefully. And that's great. But normally we will have different layers of error boundaries. ⁓ for example in here, we have already wrapped the user selector around the error boundary, but the card itself could crash, that will ⁓ break down the

whole application again. So we can define an error boundary around the card. And if for some reason inside the card that anything went wrong and there is no error handler, ⁓ it will catch in the cut level. Let's have a look at how it works. Let's go to the board column component. And in here we are rendering the cards ⁓ and we can wrap that card around the error boundary. And also introduce the boundary. And now we have an error boundary

wrapping around the card generation and because we have already catched the arrow in the card level so if we temporarily commit this out we're just using the use select as before let's remove this part and using select so when we click the avatar and use select will throw and because there is no catch here it will populate the arrow to the parent and let's see how it goes. Now if we click this one

It will throw and because we don't have the card level arrow boundary, so it will throw to the column level. And in the column we have this boundary around each card. So each card is isolated. ⁓ And if we click this one and it will crash this card only. And that's great. If the card crashes, the rest of the board remains inactive. You can use a can still add in new cards. That's great for degradation, partial failure result for downtime.

And similarly, at the page level or application level, we can define a global boundary around the entire application as a final fallback, the last line of defense. We can show a reload button or just a ⁓ link that point the user to check the status and so on. Do those layers work together? The component level boundary handles small local issues, or the page level boundaries ⁓ protect the feature areas.

Sneha Mehra (00:07:39)  
And in the top level or application level, ⁓ the boundary protects the entire shell. Each one provides an appropriate recovering path.

Sneha Mehra (00:07:52)  
Arrow boundaries are critical for production readiness because they make failure survivable. They give users a way out instead of a forced ⁓ refresh. They let developers log and fix issues without breaking the rest of the experience. There are a few notes to keep in mind. Use specific fullbacks ⁓ when possible, like close or retry buttons, and wrap dynamic or third party components individually. And always log the arrow in

component did catch so you can choose real instance later. This isn't just for good user experience, it's a good system design. It's about resilience.

—--------------------------------  
Sneha Mehra (00:00:00)  
The last time we saw the problem, a user update in top bar didn't reflect ⁓ in the board. This happened because the user data was nested and duplicated. In this video, we'll fix that by implementing ⁓ a normalization in React context. So the first step is to normalize the payload we get back from the server. Instead of keeping users ebanded ⁓ inside each card, I mapped them out into a users by ID dictionary.

Cards now only hold an assigned ID and the columns just hold a list of card IDs. That way all the entities are flattened into their own collections. But we still keep a column order array so the layout of the board is preserved. So I define a function called normalized board and it will get a payload ⁓ of the ship to the board payload ⁓ ship, which is exactly what returns from the backend. So if we look at the

mock ⁓ board. So basically we have columns, column is an array and for each item in it it has ID title and cards array. Cards array has card type ⁓ which basically ID title assigne. Assigning itself is a user object. So we have defined a ⁓ type for their server ⁓ returned payload. So basically we have columns ⁓ it's a server ⁓ column. So column has cards.

server cards and the assign is the user and user is ID name and avatar URL. So this is pretty much about the type. So I have defined a function called normalized bird. So we accepting the bird and then return the normalized bird. The normalized bird is something like this one. So have column I order which is the order of the ⁓ different stage of the column. Columns by ID is all the column definition. This is column ID and this is the column object.

You can imagine that for our particular keys there are three key value pairs inside this object. Cards as well, it has old cards, and users has old users. So when we call this normalized board, we convert the payload from a nested structure to a flattened ⁓ normalized structure. And the final result of the function contains users by ID, cards by ID, columns by ID, and ⁓ column holder. So next I create the context that holds this normalized state.

Sneha Mehra (00:02:26)  
The context gives us a couple of key actions. ⁓ the ingest, board, and the observed user and also the state, which is a store. This means any component, ⁓ whether it's board or list view or the top bar, can subscribe to the same store and stay consistent. So the board contacts will be the root level contacts provider, and all the children elements will listen to the change happen on the store. So we define a board ⁓ provider component.

That uses the context provider and provide the state ⁓ and ingest board and observed ⁓ user actions. So for the ingest board, we simply call the normalized board ⁓ and whatever ⁓ we get from the backend, we just normalize it and save at the state into the contacts. And then we define an absurd user.

That will be called when we update the user information from the top bar and that will change the state and ⁓ all the children will re-render with the new state. And for the observed user, we will change the states. Update the users by ID object. we basically unwrap the what exists and for that particular ID we want to update, we will change the user object for whatever called this upsert user ⁓ function.

So that way we can have this updates in place. And also we expose the contacts through a hook, use board contacts. So we can use this hook to get the state or get the ⁓ absolute user action, so we can call that in the right place. Because the store is normalized, it's not immediately in the shape our components want. So for example, the card needs the ID title and assignee, which should be a user object in a normalized

⁓ state we don't have this assignee anymore we only have the assignee ID so we need to hydrate the data meaning I need to map the assigning ID back to the actual user and the card IDs back to the card object. This extra mapping step is the trade-off of normalization. You pay a small cost at the render boundaries but you get the consistency across the whole application. And because React is already efficient at re-rendering

Sneha Mehra (00:04:47)  
This trade-off is usually worth it. If we look at the board page, ⁓ immediately we will treat growth a ⁓ fetch, and we have that fetch result. We will ingest the board listed from the board contacts. That's basically doing the normalization. And ⁓ then in the top bar we can assume that we get the user ⁓ from the ⁓ contacts, and then we can use the absurd user from the contacts.

And when we handle save name, ⁓ which is form submission, and we call the fetch with the patch ⁓ just as before. ⁓ But after we got the updated result, we observed user that is from the board contacts that will update to the store. And ⁓ then it will re-trigger the ⁓ re-render and all the places that using the user will update it. So we go back to the board page and go to the board.

and then we have a border view. Inside the border view we can get the states out of the context and then in here we do the ⁓ hydration step. So basically we will integrate through the older array and we will get the column object by ID. So we have the column and then we will map through the cards. We will hydrate the card first. So we will need to from the cards ⁓ object

We will get that card by ID. We get the card for the assignee, we need to hydrate it as well from the state. So basically we need to get the users ⁓ from the store by the assigning ID and then we hydrate that card. Once we have this hydrated cards list, we can use that pretty much like what we have before. So now if we look at the behavior. So in a user avatar here, if I hover it's Charlie, if I

⁓ update the user in here which will send a patch request to the back end that make the change and then if I hover ⁓ it's changed to chartly new and for this one it's chart new as well. ⁓ So because they are referenced into the same object so all the data are consistent and if I do a refresh we can see in the console log ⁓ there is a object of the

Sneha Mehra (00:07:08)  
⁓ normalized data you can see it's pretty much like a user bar ID, it's a user object, a columns by ID is a column object, cards by cards object, ⁓ and also the column order, which is column one to three. And the ticket is referencing the user by an ID, and we can use this ID to find the correct user ⁓ from the user object. So because both the top bar and cards reference the same dictionary.

the normalize the data, ⁓ they all update in sync, no duplication, no steel data. And as you have noticed, normalization isn't free. We need display to always read entities from the store instead of holding on to abandoned copies in local state. And we need to hydrate at the right level so components don't all repeat the same logic, the mapping logic. But the payoff is huge, one sort of truth. Predictable updates and easier

Reasoning about state as the application grows. What I've shown here is React Context, which is a good fit for demo or a small or median project. If you are working on the larger team or need ⁓ time travel debugging, Redux ⁓ with correct entity adapter gives you the normalization patterns out of the box. If you want something lighter, DustRand is a nicer ⁓ option with a small API surface. And if you are working heavily with server-side data.

labor like a record query or relay already normalize the cache for you. You can still combine this with a small ⁓ UI store if you need more control over the view state. So that's normalization in practice. Reflecting the board data, centralized it in the context, hydrate it ⁓ for rendering, and apply the updates in one place. And next we will have a small hands-on ⁓ exercise to consolidate what we have learned in this lesson.

—----------------------

Sneha Mehra (00:00:00)  
When a record application starts small, structure doesn't seem important. You just create a few files, maybe a couple of components, a page or two, and everything feels fine. But as project grows, things quickly get messy. You start to see duplicate components, unclear dependencies and inconsistent naming. At some point even small change becomes frustrating. That's why a well organized structure matters.

It helps you to find what you need, understand relationships between modules, and scale the application result chaos. In this lesson we will look at how to structure a regular application from the simplest setup to a more mature scalable form. A regular application isn't just about components. A real-world front end project includes source code with components, hooks and sales. The assets, things like images, fonts,

and icons and configurations which is the build setup environment ⁓ variables and so on and tests ⁓ unit test integration test and end to end tests and also documentation things like readme style guides and ⁓ different nodes and also development tools like linters and the CI CD configurations all these files need to leave somewhere predictable so the goal of structuring your code is to make it easy to locate

modify and extend any part of your application. There are several ways teams to organize React applications. Let's go through the four most common approaches. ⁓ So first ⁓ the feature based structure. In a feature based structure you organize code around features or modules. Each feature has its own folder, for example, the home page, the card, the checking out and the profile. Inside each module you might have components, hooks,

Services and tasks related to this feature. This structure keeps the related code together, which makes it easier to maintain and scale. It's a good default choice for most medium to large projects. The downside is that you might duplicate code between features. For example, each feature might define its own model or button. You will handle that later with a shared layer. And the second approach is called component-based structure.

Sneha Mehra (00:02:26)  
The component based structure focused on reusability. You place all components in a century component folder, organized by type, maybe buttons, models, cards, and so on. This makes sense if you are building a UI library or a design system. It's clear, modular, and easy to test in isolation. The drawback is that as application grows, it becomes harder to tell which components belong to which feature.

So it's great for shared UI, but not ideal for representing business features on its own. And the third one is called Atomic Design Structure. Atomic design breaks down the UI into five levels: ⁓ automs, molecules, organisms, templates, and pages. Atoms are the smallest unit like buttons or inputs. Moleculars combine items like a search box made of input and a button. Organisms are larger components, for example, a product card.

⁓ or header, templates define layouts and the page represent four screens. It's a systematic way to design UIs and maintain consistence across teams, but it adds extra structure and works best when you have a design system team or a large design driven project. And the fourth one called MVVM model. MVVM stands for model, view, view model. You can think of it as a

Separating data, logic and UI. So modules your data, for example, a product model ⁓ or card item model. While view models are hooks or classes that hold state and business logic. Views are react components that renders the UI. This separation keeps logic out of components and ⁓ makes testing easier. However, it adds complexity, so it's most suitable for large state heavy application.

So in my book React Adding Patterns, I have a dedicated chapter ⁓ to talk about all these ⁓ structures and the benefits of using different ⁓ patterns. And apart from the patterns we just explored, it also discussed how you can evolve your front-end code structure over time. Regardless which structure you start with, the key is evolution. You don't need to get it perfect from day one.

Sneha Mehra (00:04:49)  
So you start simple and as your application grows, extract ⁓ shared logic into new layers. So in next lesson we'll apply the feature-based structure into our board application. So we can see that in live. And once your structure is in place, consistency is what keeps it healthy. For example, if styles live next to component, keep that rule everywhere. And if you are using khaburb keys, if first it across the teams. Tools like yesLint or FordLint

Can automatically check names, the naming conventions, and ⁓ you know folder commissions. This kind of consistency helps the teams navigate the project without constantly wondering where things belong. ⁓ And the last thing to notice that the project structure is not static. It's something that you continuously refine as your application evolves. So you always start with the simple, you know, feature-based layout, introducing shared layers.

when duplication appears, and keep naming conventions consistent across the team. So good structure doesn't just make your code base ⁓ look clean, it improves developer velocity, reduces confusion, and makes scaling the system much easier over time. And that's it for this lesson. And in the next one, we'll apply the feature-based structure into our board application. And I will see you there.

—-----------------

Sneha Mehra (00:00:00)  
Alright, this is the final video of the course and I want to talk about something that affects all of us now, how to work with AI as a front end engineer. AI can write code incredibly fast. Sometimes it feels like it can generate a entire feature in just a few seconds. But speed is not the same as understanding. And direction still has to come from you. So in this video I want to share a few

practical strategies I use every day to work with AI effectively.

Sneha Mehra (00:00:35)  
AI is really good at details. It can write type references, fixing setx arrows, expanding bioloplate, and handle repetitive code very quickly. It's pretty much like having a junior developer who never gets tired. But it also makes things up sometimes. You will get a confident answer that completely wrong and or a code snappy that doesn't even compile. So AI is great at filling in small pieces.

But it's not good at holding the whole system in its hand. ⁓ That part is still your job.

Sneha Mehra (00:01:13)  
Another thing to keep in mind is that AI struggles when you ⁓ throw a big vague problem at it. If you say building this whole feature, it often gets lost or produce something that doesn't really work with your system at all. When that happens, the best strategy I found is to step back a little bit, you go back to the original question, break it down into smaller pieces, give it more context and start with

the clear steps. This is exactly like guiding a junior developer in your team. If they reach to that end, you don't push them deeper, you pull back, clarify the requirement further and break it down to these manageable ⁓ steps. For AI, those steps might look like generate the type based on the schema we have, or write the code to handle the loading status, adding a test for this particular scenario

reflect the hook into a smaller function, things like that. So basically you are stay in control. AI helps you to execute the steps more quickly.

Sneha Mehra (00:02:26)  
Some of the biggest productivity wins comes from letting the AI handle the supporting tasks, the things that normally take time but don't require deep design decisions, like fixing linking arrows, infer types, writing the accessibility checks, or generating mock data, getting a small variation of a component, ⁓ or apply a particular pattern to a component. Those tasks are perfect for AI.

And offloading them frees up more of your brain to focus on the co design work.

Sneha Mehra (00:03:05)  
One approach I have found really helpful is to category the work you do into four simple buckets. Basically you have things like you are good at and you are not good at and AI are good at and AI are not good at. The swift spot is that the overlap between what you are good at and what AI is good at. This is where collaboration becomes really powerful. You set the direction and the AI handles the detailed implementation.

And you only do the review when it's done. Then there is the area where you are not good at, but AI is, this is great for learning. Ask it to explain something, ⁓ generate examples or showing a different approach than you currently have. And of course, avoid the area where neither you nor the AI are good at. This is where most mistakes and wrong assumptions happen. This simple categorization helps you decide.

what to delegate to G AI, what you can learn from it and what to keep firmly under your own control.

Sneha Mehra (00:04:13)  
And before finish, I just want to say thank you. Thank you for going through the course with me and for supporting my work. I really appreciate it. AI isn't the end of front end engineering. It's not replacing people who understand what they are doing. If anything, AI makes the fundamentals even more important. The framework we use in this course, the mental model, the pillars, thinking across the build, deployment at runtime,

timeline, all of these are still highly relevant. And when you combine this understanding with AI, you become the person who actually knows what's going on. You are not just a generator the code, you are guiding the direction, making decisions and shaping the system. Use these tools wisely, keep practicing and you will keep growing into someone who truly understands their craft. Thanks again for learning with me.

I hope to see you in the next course.  
—---------------

Sneha Mehra (00:00:00)  
In last video we saw how content security policy acts like ⁓ powerful Safety Net blocking malicious script at the browser level. But CSP is your last line of defense. What about your first? That's what we will explore in today, input sanitization. It's a process of cleaning user input before it even reached your backend storage. This is especially critical when your application needs to display rich text.

or HTML content like comments, BIOS or Markdown. Today we will see why React built-in protection isn't always enough when you might have to use the dangerously set inner HTML and how DOM Purify keeps you safe when you do. Let's dive in. So React normally keeps you safe automatically when you run the content like this, in the H3 header you run the title and you got the benefits automatically.

Even if someone tries to inject the script, like alert something, Reactor doesn't execute it, it just shows the text itself, and that's your first line of defense. But what if you actually need to run the HTML? ⁓ A rich text editor where user can bolt or link text? A markdown viewer converting Markdown ⁓ to a HTML document, ⁓ a user description that supports simple formatting, or a comment section that allows tags.

In this case you want to render the real HTML and if you skip everything, users just say the skipped version instead of the bold or the meaningful HTML. And that's when many developers reach for this API, the dangerously set in HML. And the name of the API says it all already. React even labels it dangerous because it's your responsibility to make sure the HML you're using in a tag is safe.

So let's say you want to let the user add a formatted description to a card. And a malicious user could submit this instant. You can see this is a text description, but it has some ⁓ HML in it. And it's not good. And because we're using dangerously set in HTML, we get these pop-ups. And if we don't have the CSP, we will definitely get the attack. And every user who opens the card runs the attacker's code.

Sneha Mehra (00:02:27)  
Their cookies, ⁓ sessions or private data are stolen, which is definitely not what we want. This is where DOM Purify come in. It's a small bottle tested library that sanitize HTML. Think of it as a bouncer for your content. It lets the good tags and through all the bad ones. Essentially it allows the thief tags like strong, emphasize ⁓ or a anchor, and it removes dangerous tags like script, ⁓ iframe.

And it also strip on safe attributes like on arrow, on click, these even handlers.

Sneha Mehra (00:03:06)  
So firstly we need to install the DOM purify and the types into our ⁓ board application. And to use it, we need to import the DOM purify API from the package. Let me show you how. So in here we are using the DOM purify the sanitize and we passed in a content, ⁓ which is a string. ⁓ Normally ⁓ it's our description of the card or title. And then we define the allow tags, ⁓ the b tag, i tag, strong, emphasize, ⁓ link.

And the paragraph and the library and we allow only these attributes, the HREF and the title. And for other tags or the attributes of the tags ⁓ that not listed here, we just remove them. For example, if we pass in the malicious rich text into ⁓ the standardized HTML, and because the image is not listed here, it will be removed, and the script is not listed here, it will also be removed.

And also the iframe. And then if we run the dev instance ⁓ and we go to the browser ⁓ and obviously this is the vulnerable version of it. And if we go to the protected and if we inspect this description, you can see ⁓ in here we have this this bug is strong and needs emphasize attention, please see the document stuff. So the save content are kept on the screen.

So we can still see this content, but for the malicious ones, they are all removed by the synetized HTML, which we are using in the card. So if the mode is safe, we just synthesize the HML. The content is ⁓ the malicious rich content, and then we use this synthesized version. And even we use that V dangerous set in the HTML, because the content has already been synitized, we have removed all the

malicious parts. So it's works perfectly. So we have the both benefit of rendering the HML content, but we're not executing the malicious ⁓ tags. And that's some certainly ⁓ best practices we need to remember. So firstly try to avoid dangerously set in HML API whenever possible. Use normal reactor rendering or markdown libraries that escape safely. And secondly if you must use it, synthesize first.

Sneha Mehra (00:05:32)  
Always pass your HML through DOM purify and strictly control allowed tags and allowed attributes. And third, you combine it with CSP. Even if something slips through, your CSP header will block script execution. So in short, never trust the user input, always synetize. So let's wrap it up. ⁓ React already is keep HTML by default, but when you need to render rich text.

You have to take responsibility. DOM Purify lets you safety display user-generated HML by stripping out the malicious script while keeping a legitimate formatting intact. And when you combine synthesization with CSP, you get a strong layered protection.

—-------------------------

Sneha Mehra (00:00:00)  
In the last video we looked at how XSS attacks work and how a tiny mistake like using dangerously set in HTML can let attacks run JavaScript right inside your user's browser. And today we'll look at the solution side, how modern browsers help us protect against these attacks using something called content security policy or CSP. By the lesson you will see how reactors ESKPI CSP work together to form a strong

Layer defense. So as demoed before, if we run the access demo in a ⁓ vulnerable version, we can see this attack ⁓ alert or pop up. And that is a problem. And if I go to the back end and make some small change and if I relaunch the application again.

And now if I go to the browser and do a refresh, you can see the message is gone. And also you can notice there are some errors in the console. It says refuse to execute inline event handler because it violates the following content security policy directives. That means even we have this malicious code, they just cannot run in the browser. So if we do a refresh with this vulnerable E-board.

And you can see there is no pop-up anymore, there is no XS attack triggered message anymore. And that's where content security policy becomes our safety net. So you can see refused to execute inline event handler stuff when we load the page. The browser shows a warning. So what just happened? If we open up the network tab, go to all, and if we look at this HTML request.

And in the response we can find something here says content security policy ⁓ is some configuration. And that's what we set in the backend. So basically we define a security header, which is content security policy. And the value of this policy could be a very long string concocted by semicolon. So we are defining a customer middleware ⁓ function here. That means we'll add the security header to all the responses.

Sneha Mehra (00:02:14)  
That's why the content security policy is added to all the responses ⁓ from our 4000 server. So basically CSP is a special HTTP header that you server send to the browser. It tells the browser what's allowed to run and what's isn't. So when the server sends a header content security policy security source self, the browser interprets it as only execute ⁓ JavaScript that comes from the same origin, the same domain as this page.

Anything else ⁓ like inline code, remote script, injected code should be blocked. So CSP doesn't stop the injected itself, the malicious HTML code are still there, it's it still appears, but it blocks the execution of any scripts that doesn't meet the rules. So when attackers inline on arrow handler tries to run, the browser checks the rules and say ⁓ no.

This JavaScript doesn't come from the trusted source and it blocks the execution completely. So without CSP, ⁓ when the HTML injected, browser renders and the on arrow fires. And then JavaScript runs, so we got the attack. And with the CSP, when the HTM injected, the browser renders and when the on arrow fires, the CSP blocks the script, so we are safe. And that's the magic of the CSP, it doesn't stop the injection itself.

But it prevents injected JavaScript from doing any damage. So you you might be wondering if SSP is so powerful, why do we still need reactors escaping or sanitization? That is because good security is always about defense in depth, ⁓ having multiple layers that protect each other. So our first line of defense is prevention. Make sure untrusted content is never interpreted as HTML. React already helps you here, so when you do

H3 with this bracket display title, React automatically skips the title. If someone tried to inject the image tag, it rendered as harmless text. And your second layer is medication. Even if HTML somehow slipped through, maybe due to a bug or a third-party library, ⁓ or maybe people just using the dangerous set inner HML incorrectly. In this case, CSP prevents the malicious script from running. Think of it this way.

Sneha Mehra (00:04:40)  
React escaping is your SIF belt, it prevents accidents. While CSP is your airbag, it saves you when something goes wrong. And you really need both. Let's wrap up what we learned today. ⁓ So XSS attacks inject malicious HML that runs JavaScript in your users' browsers. And React Auto Escaping prevents HTML from being executed in the first place. While CSP adds the powerful

Browser level safety net that blocks any untrusted or inline JavaScript. Together they form a two-layer defense prevention plus protection. CSP is very easy to set up and it gives you ⁓ enormous security benefits for almost no cost. Every production web application should have it. In the next video, we'll go one step further, sanitizing the user input so we can eliminate some of the problem from the source.

—---------------  
Sneha Mehra (00:00:00)  
Last time we fixed the risk conditioning in our search with a board controller. In this lesson we'll make our search cheaper and smoother using two tiny timing tools, D bounce and throat. We will cover the idea in Planned JavaScript, weld it into our application, and then prove it in Chrome DevTools. So let's talk about the concept first. The D bounce weeds for silence. So when events ⁓ come quickly, like keystrokes, ⁓

Debunce reset a timer on every event. Only after there has been no event for let's say 300 milliseconds do we run the function. If we type ⁓ INST quickly, ⁓ D bounce fires only ⁓ once after I stop. On the other hand, throat limits frequency. During the long bust like a scroll, drag and drop, ⁓ or faster typing. throat allows the function to run at most

once every n milliseconds. You still get updates during the ⁓ bust, ⁓ just not for every single event. There are tiny ⁓ framework-free helpers that you can paste into static page and it will work. So let's look at the DBounce implementation. As you can see in the DBounce ⁓ we ⁓ will accept a function and a delay as a parameter. ⁓ we can give the default ⁓ delay.

So we define a time ID available living in a closure, it remembers the latest timer between calls and return new function, the debunster wrapper, and you call this function instead of the ⁓ ifn directly. On every call to the wrapper, we clean up the timer, cancel the previous ⁓ scheduled run, and then we reschedule the fn to run after delay milliseconds with the latest argument.

If calls keep coming faster than delay, the timer keeps getting reset and nothing runs yet. Once there is a pause longer than the delay, the last scheduled timer fires and even runs ⁓ with the most recent arguments. Here's a simple animation to demonstrate how it runs. So let's say we have a timeline and ⁓ initially at time zero the function is called, so we debounce that.

Sneha Mehra (00:02:26)  
Instead of calling the FN directly, we will call it ⁓ after three hundred milliseconds. As we wait, there's another event triggered on ⁓ T fifty, so the timer is rescheduled to three hundred and fifty, but it's still not called yet. And then another event comes ⁓ at one hundred fifty. So we will reschedule the timer again to four hundred and fifty milliseconds. At until at this point there is nothing.

been called yet. So we keep rescheduling the timer instead of coding the function itself. And then ⁓ nothing happened during this time, the latest event to the delay, then the 450 millisecond meet and then we will call that function eventually. So you can see that even during this time there are many events happened, we only call the ⁓ function once.

We will continue to watch the event happened. If it's ⁓ less than the delay, we simply ignore the call and reschedule the timer and until we have a pause and longer than the delay. We will eventually call that latest wrapper function. DBunce delays ⁓ running a function until there's been a pause in events for a set time. Each new event resets the timer so if input keep arriving faster.

than the delay, nothing runs. ⁓ well when things go quiet it fares ⁓ once with the latest argument. This makes the ideal for do it after the user stops ⁓ action like ⁓ search queries, ⁓ form validations or autosave. That reducing wasted work and ⁓ the server load. And now as we are in React

we can either keep using the vanilla helper on the event handlers or using a tiny hook ⁓ for convenience. And with this hook in our search ⁓ scenario, we don't set the value directly and we just put a delay there. During the span of the delay, ⁓ we won't change the value. And when we stop ⁓ and after three hundred milliseconds the value will be set and then the search will be performed.

Sneha Mehra (00:04:43)  
So we get the result for the value altogether, not for every keystroke. So we prefer to use DBounce for searching. If you really want result to update during typing, you can swap to use throat ⁓ the same way. I will leave that to you. So now let's have a look at how we can use this use debounce hook ⁓ in our search scenario. So let's copy this ⁓ use DBounce hook into our application.

So we have this UDBounce here, and we will ⁓ define a new variable that uses a debunced version of the search. If the state search changed ⁓ very quickly, we will debunce it. ⁓ And if the change is is happened within 300 milliseconds, we will ignore the change until we stop. And then we get the latest value of the ⁓ search altogether. So then we will watch this debunced keyword ⁓ in our ⁓

use effect in here and when it changed we will append that to a query string and ⁓ prepare the parameter just like ⁓ before and then we do the search and if we go to the application in here so if I type very fast let's say IEST ⁓ you can see it only happened once ⁓ and if I remove them and after ⁓ a 300 millisecond it will ⁓ send out.

So if we type let's say IN, ⁓ it just ⁓ happened once. ⁓ and if we type very quickly ⁓ on to say something like this, you see during the typing nothing happened and all of sudden the query is contacted into a long query string here altogether. And that's how you use a use D bounce hook.

Throat ⁓ limit how often a function can run during a burst, at most once every n milliseconds while still providing periodic updates. Instead of waiting for silent, it emits an on a fixed ⁓ condense, often with a final trailing calls that captures the latest state. This is great for continuous interactions where you want responsiveness without spamming ⁓ work.

Sneha Mehra (00:07:07)  
like scrolling, resizing handler, drag and drop, the sliders, or live typing purviews. So threat tool is a little bit ⁓ complicated as you can see, it's more live code. So first we define a verbal last and trading ID. So check the time span of the last time the iPhone actually run. And the trailing ID holds the ID of the scheduled trailing call, one that should happen when the interval ends. And we also return new function

The through third ⁓ wrapper. On every call to the wrapper function, we compute the now and the remaining if the remaining less than zero, we have passed the interval, so we'll run the function immediately and ⁓ set the last to now. In the else branch, ⁓ we cancel any previous scheduled trailer call ⁓ and we schedule the new trailing call to run ⁓ after remaining milliseconds. When it files, it sets last equals now.

I'll run Fn with the latest argument received. So during the burst of the events, you get at most ⁓ one emitted call at the start of the interval and at most one scheduled call at the end of the interval. If events continue non-stop, you will effectively get about one call per ⁓ interval milliseconds. Here is a simple animation as well. So you can see yellow dots indicate potentially the calling point, the interval. So at T0.

we have event triggered, so nothing will be called at this point. And then another event triggered, nothing happened. Because we are still in the remaining, and then at the 200 millisecond, the function will be called at once because and then moving forward, ⁓ still there might be multiple actions happened after the 200 mark, and we just counted that one at 400 millisecond that will be called once. So that will be continue processing.

And then after 400 mark, if there is no more new event happened, there will be no more function calls with the throw tooling. So through basically means don't call more than once ⁓ per n milliseconds with a final call at the end of each window. Sorto where it lets you choose a leading ⁓ and trailing behavior. Combining both gives a quick first response plus a catch-up at the end. Like D bounce.

Sneha Mehra (00:09:34)  
certainly reduce how often you start ⁓ expensive task. ⁓ but for network requests you should ⁓ still protect against still responses with cancellation. And with the start you will see a handoff request spaced ⁓ approximately 200 milliseconds apart. So in all cases we keep ⁓ a board controller. If two requests overlaps and ⁓ the old one returns last, it's aborted and cannot override the new result.

just like before. So to wrap up ⁓ D bounce wait for quiet, ⁓ best for search after typing, ⁓ form validation, ⁓ auto-saving, we can start the ⁓ delay around 200 to 300 milliseconds. Well on the other hand, a throat cap frequency, best for scroll or drag, ⁓ resize the window, ⁓ you probably will use 100 to ⁓ 200 milliseconds ⁓ for the delay.

And you will also want to keep the abort controller at guards ⁓ against out of order responses. You can combine these ideas all together. They are indeed two tiny timing tools, but they also do have a big impact on performance and ⁓ the experience of your application.

—----------------------  
Sneha Mehra (00:00:00)  
Alright, we have reached the final part of this course, and instead of giving you a long summary, I want to leave you with something much more useful, a simple way to think about front end system design so you can use it every day at work. We have covered a lot of techniques, but you don't need to memorize all of them. What matters ⁓ is having a mental model you can lean on whenever you design or review something in your front end system.

Sneha Mehra (00:00:32)  
One of the easiest ways to organize everything we learned is to map it onto the software lifecycle. So we have build, deployment, and runtime. Let's work through it quickly. So for build, this is everything that happens before your code reaches the browser. Things like ISISR, rendering strategies, bundling, code splitting, ⁓ resource per processing, linting, and testing. For example,

Code splitting decide how your application is broken into smaller chunks so users don't download everything upfront. And static analysis and tests help catch issues earlier, long before they reach to the production. Next we have deployment. This is how your code is delivered to users. Here we think about CDNs, basically global serves that catch in your static assets so users get them faster no matter where they are.

we also think about HTTP caching, which let you the browser reuse what is downloaded already instead of fetching it again. Those decisions directly affect your loading speed and reliability. And finally we have this technique at runtime. This is everything that happens when user interact with your application. Normalization, state management that help you keep your data consistent. Preload, pre-fetching let you

load data and code before it's needed, making the UI feel much more smoother. Pagination and visualization help your handle the large lists without freezing the UI. And you also have error handling, logging, accessibility, localization and security. The things every real world application needs. This left circle is not something you need to memorize either. It's just a helpful map, a way to

See where your decisions live.

Sneha Mehra (00:02:35)  
And another structure we can use is the structure we have in this course. Now let me connect this back to the pillars we used throughout this course. And as you can see we have several parts in the framework. We start from the data modeling and state management. ⁓ We discuss the techniques like normalization, the ⁓ tools and libraries like Redux and the React Context API. And then we talk about data fetching.

In there they have paginations, ⁓ request management, and ⁓ then we have the data mutation. So from there you have real time updates, optimistic updates, ⁓ the signalization, nested structure and so on. And ⁓ for the performance part, this big part, we touched the rendering strategy ⁓ in the performance parts, ⁓ like server side rendering, the static side generation and the ⁓ hybrid approach.

And also we talk about the perceived performance like skeleton pattern, loading indicator, ⁓ the code splitting patterns, ⁓ preloading, lazy loading, and how to bundle them separately. And finally we have these ⁓ cross functional requirements or product and ready techniques. We talk about strength and security, accessibility, and also the error handling, ⁓ testing.

in different layers the testing strategy to make sure our code is always ⁓ in a maintenance state. And again we don't need to remember everything. You just need to remember the shape of the model.

Sneha Mehra (00:04:14)  
The most important thing I want you to take away ⁓ is this. System design becomes simpler when you use it as a thinking framework, not a list of techniques. When you pick up a new feature, even a small one, work through the model. What happens at build time, what decision affect the deployment, what matters at runtime, or you can think of the pillars we used here. This habit is what makes you consistent.

It helps you to move faster and it helps you reason about problems long before they run into defects.

Sneha Mehra (00:04:54)  
For example, say you are building a feed list. At build time you might split the feed into separate chunks so the home page loads ⁓ faster. And during the deployment you rely on CDN caching for images so the scrolling feels smooth. ⁓ And at runtime you use infinite loading, ⁓ request management, error handling, and maybe even optimistic updates for user actions. And in cross functional requirements

You track analytics ⁓ events so you understand how people engage with the feed list. And when you actually designing or reviewing the design, you are using the map to help you to navigate ⁓ through the system. The best way to get comfortable with all this is to ⁓ use it. Pick one or two ideas and apply them into your next feature. Review your teammates PR using the structure, using the SIM framework to explain decisions when you

⁓ talk to a designer or backend engineer. So as we always said, the practice is what turns the ⁓ cult knowledge into instinct. And before we finish, I want to point you to the next video, which is about working with AI as a front end engineer. AI can write a lot of code very quickly, much faster than any of us, but you are still the one in charge. You are still responsible for the decisions, the structures and the final outcome.

And as a person, your real advantage is learning how to break things down to smaller tasks, how to guide the AI with cleaner ⁓ prompts and how to check whether the outcome actually fits your system.

—---------------  
Sneha Mehra (00:00:00)  
In the last lesson, we talked about why data modeling matters. Now let's make it concurrent with a small case study. We'll look at something simple, like sidebar navigation, and say how different modeling choices affect our design. On the UI, on the left hand side, there is a feature section of the menu. ⁓ There will be some common features users can navigate into, ⁓ and for some cases there will be a fancy features, and this feature is only

⁓ enabled for the premium users. And there is a small decoration icon shows up on the right hand of the feature. And if we go to the code ⁓ for the sidebar and accepting a user as a input and if we cut check the user type as the ID and as a name and the plan type ⁓ and ⁓ register that which is a date. ⁓ And if we check the ⁓ sidebar component

A checking the user type is equal to premium or not. If it is a premium user, we'll use that as a flag to show the nav item for the fancy features. And if we look at the API, so basically we will ⁓ fetching API for API slash me and it will return the user ⁓ which is defined as a hard coded ⁓ user details here. the plan type is premium.

And that's why we can see this menu here. And if I change that to free user, for example, and if we go back the feature is gone. Right now our API just tells us the user's plan type. In the front end we decide ⁓ that if the plan is premium we should show the fancy feature and so on. This approach works, but notice the problem here. So every front end, including the web or mobile, it has a

⁓ copy of the same logic. If the business changed the rule, we have to redeploy them in all ends. What if we add another condition to the current code? So let's say I have another ⁓ fancy feature that only enabled when the user is premium and ⁓ the user has been registered over three years. And similarly we'll duplicate this one and using the show another fancy feature, another.

Sneha Mehra (00:02:26)  
fancy feature. And then we will need to change the plan type back to premium and see if it's showing up. ⁓ and it shows the another fancy feature here and if we change that to let's say twenty twenty four and that feature will be height. So now our UI is handling business rules. The plan logic and the time calculation is put in front end. Imagine this duplicated in other clans like in desktop, mobile or even

the mobile app, this is brittle and hard to maintain. So instead we can push entitlement logic to the API. ⁓ The back end knows the rules, it can return ⁓ feature flags directly. The UI just renders what is told. So I have made some ⁓ changes in the back end API. So now the API ⁓ API slash me ⁓ returns not only the user but ⁓ a features object.

And we have a utility function called plan to features. If the plan is premium, then ⁓ show fancy feature is true. And if the user regists over three years, we will have a show another fancy feature ⁓ to true as well. And then ⁓ when we get this user from the front end, we return now this one. And in front end, when we hit that API and there is a features object that has the old feature flags.

show fancy feature true show another fancy feature false and then we in the front end we can rely on this feature and then in front end the sidebar component we don't have to put the logical here to do the check. Instead we directly use the user the features show fancy feature ⁓ and also we will use something like this show another fancy feature.

And then because the flags already in the response directly, so we don't have to put the logic to do the calculation in front end. So front end just to render whatever it's ⁓ told. Obviously there is some type checks here, we can fix that later. And now the UI is declarative, the backend owns the business logic and the rules, and all the client side stay consistent. And of course not all conditions belongs in the back end.

Sneha Mehra (00:04:51)  
Some are purely about ⁓ presentation, things like highlight new features or showing a badge in the UI. So the split of the responsibility is that the back end decide which feature a user has and the front end decide how to present them. There's also a main ground in raw systems the API you are consuming isn't just for your front end.

It might also serve as a mobile application or a partner integration or internal tools. That means you cannot just change it freely every time the UI needs something new. This is where GraphQL or backend for front end BFL for short comes in. They allow you to keep your bin through ⁓ centralized in the back end while shipping the response ⁓ specifically for the UI. So for example, this is ⁓ GraphQL Curry, ⁓ we can get the

features ⁓ just by specify what we want from the back end. And in the back end might have logic in a Rust API or rust for API expose the the raw data or whatever. But in GraphKL resolvers ⁓ layer we can resolve these fields only for this particular UI. So here the front end asks for exactly the field it needs, nothing more, nothing less. The business logic still lives in one place, but the response is titled

⁓ to the consumer's access patterns. This means we can evolve our backend service without breaking other clients and still keep the front end simple and declarative. This is a small example, but the same principle applies across your whole system. ⁓ Good data modeling helps you put logic in the right place, keep your code clean and make your application scalable. So next we will bring this thinking back to our starter project

Instead of just a hard code state, we will actually fetch data from ⁓ two endpoints, the API slash me for the user and ⁓ the API board for the board and the cards. But here is a twist. If we just store those response as is, we'll quickly run into duplications and inconsistency. ⁓ for example the user ⁓ might appear on a board and ⁓ on a card. But if we update them in one place

Sneha Mehra (00:07:13)  
The other doesn't change. This is why in the next lesson we will introduce data normalization in the front end. We will take those API responses and model them in a way that keeps entity consistent across the app. And that's where data modeling really starts to pay off. Not only help you ⁓ understand the domain, but also give us the right structure to build scalable UIs.

—---------------------

Sneha Mehra (00:00:00)  
Today we'll dive into one of the most dangerous and honestly most common security issues in modern web applications. Curse site scripting or XSS. We'll look at how XSS attacks actually work, why they are so dangerous, and what makes them tricky to prevent. I will even show you a live demo and attack happening right inside our application. By the end, you will clearly understand how attackers can run code inside your

user browsers and why we need stronger defense like content security policy to protect real world applications. Let's jump in. So what exactly is cross siting scripting? In simple terms, XSS happens when an attacker manages to inject their own script, JavaScript code, into your web application. That code then runs inside another user's browser as if it was trusted code from your site.

Let's make this concrete. Imagine you are building a task application like what we are doing. Users can create cards with titles and its descriptions. Pretty harmless, right? Now what if one of the titles looks like this? At first glance, that looks weird, but maybe not as dangerous. But wait until you see what happens when another user loads that card. ⁓ I have prepared a demo mode in our application. When I visit

XSS demo equals to vulnerable, it switched to a version that uses on SIF code. See this red banner on the top? This means we are in vulnerable mode. And for the card, ticket one, it's highlighted because it rendered using dangerously set inner HTML. ⁓ Now let's reload the page.

And immediately you can see a pop-up shows up, XSS attack successfully. And if we look up the dev tools, and you can see that in the console, attacker can catch this ⁓ cookie and save that to their own instance, and they can use this cookie to impersonate this ⁓ particular user. So let's break it down what happened step by step. So the malicious HTML is injected into the DOM, the browser passes it and sees the image tag.

Sneha Mehra (00:02:16)  
It tried to load the image from the source equals X. The image fails to load, which triggers the onArow handler, and that handler runs JavaScript, the attacker's code. I have defined a variable code malicious payload in ⁓ XSS demo. So basically what happens is this malicious HTML is injected into the DOM. The browser passes it and see the image tag here, and it tries to load this X. Obviously it doesn't exist, browser will code this on arrow handler.

And that's why we have these logs printed out. And in this demo we just log the cookie. In real attack, it could send your session cookie to a attacker's website. And once that happens, the attacker can impersonate you, read your data, even perform actions on your behave. And here is the scary part, it spreads automatically. So one single malicious card can compromise every user who views it.

You might be wondering, can we just block the image tag? Unfortunately, no. Image tags are totally legitimate. We use them everywhere, and that's what makes this attack so sneaky. Here is why it's clever. First, it doesn't use script tag, so it bypasses basic features that only looking for boost. ⁓ Second, it always triggers because the image source is invalid, the own error event fires every time.

And third, there are endless variations of these tricks. You can use the iframe, ⁓ SVG, even div or body. And even if you block one pattern, attackers just switch to another. And if we look at how the malicious payload is actually used in our component, you will find that in the card, ⁓ we are checking if we are in demo mode. So for example, we are using the ticket one as a demo card, and when the mode is vulnerable, we're

Using the dangerously set inner HTML ⁓ to the display title, which is a malicious payload, and that will showcase the actual arrow. So you might be wondering why would React even provide an API code dangerously set inner HTML? That sounds like asking for trouble, right? Well, sometimes you actually need it. Here are some legitimate user cases. So for CMS or markdown output.

Sneha Mehra (00:04:36)  
You store user content in markdowns and run it as HTML. You need the paragraph that links the ULI tag to appear, not just the skip text. And also sometimes you use the server-size rendering HTML snippet. You backend provide safe, a pre-synchronized fragment, like documents or changelogs. They all could include HTML and third-party widgets, some abandoned or email templates come in as HML as well.

And there are also cases like a legacy content or what do you say is what you get editors when you migrate old HTML content or handle rich text from content editable field. In all these cases you have to inject real HTML and in React, that means you need to use a dangerously set in the HTML. So it's not evil by itself, it's just ⁓ dangerous if you are not careful. Let me go back to the demo again. So you can see ⁓ in a vulnerable

mode you can see this ⁓ dangerously set inner HTM is used. And if we use protected version, which will not use the dangerously set inner HML, so you can see the actual content is fun bug, image, ⁓ things like this. This is skipted HTML so it doesn't execute. So by default it's SIF. And this happens because in a SIF mode we are using the display in this brick it and the React will

Handle the escape for us, that happened automatically. And ⁓ only when you use the dangerously set in the HTML at that point the browser will need to pass the HTML and ⁓ trigger the malicious code.

Sneha Mehra (00:06:18)  
So, to recap, XSS happens when untrusted content like user input end up being executed as code in someone else's browser. It's sneaky, it's powerful, and it can spread very fast. And even legitimate APIs like dangerously set in HTML can open the door to it if you are not ⁓ sanitizing properly. In the next lesson, we will talk about how to defend against these attacks, ⁓ specifically using content security policy.

or CSP in short, and HML sanitization to keep your users safe.

—------------------

Sneha Mehra (00:00:00)  
Alright, we have reached the final part of this course, and instead of giving you a long summary, I want to leave you with something much more useful, a simple way to think about front end system design so you can use it every day at work. We have covered a lot of techniques, but you don't need to memorize all of them. What matters ⁓ is having a mental model you can lean on whenever you design or review something in your front end system.

Sneha Mehra (00:00:32)  
One of the easiest ways to organize everything we learned is to map it onto the software lifecycle. So we have build, deployment, and runtime. Let's work through it quickly. So for build, this is everything that happens before your code reaches the browser. Things like ISISR, rendering strategies, bundling, code splitting, ⁓ resource per processing, linting, and testing. For example,

Code splitting decide how your application is broken into smaller chunks so users don't download everything upfront. And static analysis and tests help catch issues earlier, long before they reach to the production. Next we have deployment. This is how your code is delivered to users. Here we think about CDNs, basically global serves that catch in your static assets so users get them faster no matter where they are.

we also think about HTTP caching, which let you the browser reuse what is downloaded already instead of fetching it again. Those decisions directly affect your loading speed and reliability. And finally we have this technique at runtime. This is everything that happens when user interact with your application. Normalization, state management that help you keep your data consistent. Preload, pre-fetching let you

load data and code before it's needed, making the UI feel much more smoother. Pagination and visualization help your handle the large lists without freezing the UI. And you also have error handling, logging, accessibility, localization and security. The things every real world application needs. This left circle is not something you need to memorize either. It's just a helpful map, a way to

See where your decisions live.

Sneha Mehra (00:02:35)  
And another structure we can use is the structure we have in this course. Now let me connect this back to the pillars we used throughout this course. And as you can see we have several parts in the framework. We start from the data modeling and state management. ⁓ We discuss the techniques like normalization, the ⁓ tools and libraries like Redux and the React Context API. And then we talk about data fetching.

In there they have paginations, ⁓ request management, and ⁓ then we have the data mutation. So from there you have real time updates, optimistic updates, ⁓ the signalization, nested structure and so on. And ⁓ for the performance part, this big part, we touched the rendering strategy ⁓ in the performance parts, ⁓ like server side rendering, the static side generation and the ⁓ hybrid approach.

And also we talk about the perceived performance like skeleton pattern, loading indicator, ⁓ the code splitting patterns, ⁓ preloading, lazy loading, and how to bundle them separately. And finally we have these ⁓ cross functional requirements or product and ready techniques. We talk about strength and security, accessibility, and also the error handling, ⁓ testing.

in different layers the testing strategy to make sure our code is always ⁓ in a maintenance state. And again we don't need to remember everything. You just need to remember the shape of the model.

Sneha Mehra (00:04:14)  
The most important thing I want you to take away ⁓ is this. System design becomes simpler when you use it as a thinking framework, not a list of techniques. When you pick up a new feature, even a small one, work through the model. What happens at build time, what decision affect the deployment, what matters at runtime, or you can think of the pillars we used here. This habit is what makes you consistent.

It helps you to move faster and it helps you reason about problems long before they run into defects.

Sneha Mehra (00:04:54)  
For example, say you are building a feed list. At build time you might split the feed into separate chunks so the home page loads ⁓ faster. And during the deployment you rely on CDN caching for images so the scrolling feels smooth. ⁓ And at runtime you use infinite loading, ⁓ request management, error handling, and maybe even optimistic updates for user actions. And in cross functional requirements

You track analytics ⁓ events so you understand how people engage with the feed list. And when you actually designing or reviewing the design, you are using the map to help you to navigate ⁓ through the system. The best way to get comfortable with all this is to ⁓ use it. Pick one or two ideas and apply them into your next feature. Review your teammates PR using the structure, using the SIM framework to explain decisions when you

⁓ talk to a designer or backend engineer. So as we always said, the practice is what turns the ⁓ cult knowledge into instinct. And before we finish, I want to point you to the next video, which is about working with AI as a front end engineer. AI can write a lot of code very quickly, much faster than any of us, but you are still the one in charge. You are still responsible for the decisions, the structures and the final outcome.

And as a person, your real advantage is learning how to break things down to smaller tasks, how to guide the AI with cleaner ⁓ prompts and how to check whether the outcome actually fits your system.

—--------------------  
Sneha Mehra (00:00:00)  
In the last lesson, we talked about why data modeling matters. Now let's make it concurrent with a small case study. We'll look at something simple, like sidebar navigation, and say how different modeling choices affect our design. On the UI, on the left hand side, there is a feature section of the menu. ⁓ There will be some common features users can navigate into, ⁓ and for some cases there will be a fancy features, and this feature is only

⁓ enabled for the premium users. And there is a small decoration icon shows up on the right hand of the feature. And if we go to the code ⁓ for the sidebar and accepting a user as a input and if we cut check the user type as the ID and as a name and the plan type ⁓ and ⁓ register that which is a date. ⁓ And if we check the ⁓ sidebar component

A checking the user type is equal to premium or not. If it is a premium user, we'll use that as a flag to show the nav item for the fancy features. And if we look at the API, so basically we will ⁓ fetching API for API slash me and it will return the user ⁓ which is defined as a hard coded ⁓ user details here. the plan type is premium.

And that's why we can see this menu here. And if I change that to free user, for example, and if we go back the feature is gone. Right now our API just tells us the user's plan type. In the front end we decide ⁓ that if the plan is premium we should show the fancy feature and so on. This approach works, but notice the problem here. So every front end, including the web or mobile, it has a

⁓ copy of the same logic. If the business changed the rule, we have to redeploy them in all ends. What if we add another condition to the current code? So let's say I have another ⁓ fancy feature that only enabled when the user is premium and ⁓ the user has been registered over three years. And similarly we'll duplicate this one and using the show another fancy feature, another.

Sneha Mehra (00:02:26)  
fancy feature. And then we will need to change the plan type back to premium and see if it's showing up. ⁓ and it shows the another fancy feature here and if we change that to let's say twenty twenty four and that feature will be height. So now our UI is handling business rules. The plan logic and the time calculation is put in front end. Imagine this duplicated in other clans like in desktop, mobile or even

the mobile app, this is brittle and hard to maintain. So instead we can push entitlement logic to the API. ⁓ The back end knows the rules, it can return ⁓ feature flags directly. The UI just renders what is told. So I have made some ⁓ changes in the back end API. So now the API ⁓ API slash me ⁓ returns not only the user but ⁓ a features object.

And we have a utility function called plan to features. If the plan is premium, then ⁓ show fancy feature is true. And if the user regists over three years, we will have a show another fancy feature ⁓ to true as well. And then ⁓ when we get this user from the front end, we return now this one. And in front end, when we hit that API and there is a features object that has the old feature flags.

show fancy feature true show another fancy feature false and then we in the front end we can rely on this feature and then in front end the sidebar component we don't have to put the logical here to do the check. Instead we directly use the user the features show fancy feature ⁓ and also we will use something like this show another fancy feature.

And then because the flags already in the response directly, so we don't have to put the logic to do the calculation in front end. So front end just to render whatever it's ⁓ told. Obviously there is some type checks here, we can fix that later. And now the UI is declarative, the backend owns the business logic and the rules, and all the client side stay consistent. And of course not all conditions belongs in the back end.

Sneha Mehra (00:04:51)  
Some are purely about ⁓ presentation, things like highlight new features or showing a badge in the UI. So the split of the responsibility is that the back end decide which feature a user has and the front end decide how to present them. There's also a main ground in raw systems the API you are consuming isn't just for your front end.

It might also serve as a mobile application or a partner integration or internal tools. That means you cannot just change it freely every time the UI needs something new. This is where GraphQL or backend for front end BFL for short comes in. They allow you to keep your bin through ⁓ centralized in the back end while shipping the response ⁓ specifically for the UI. So for example, this is ⁓ GraphQL Curry, ⁓ we can get the

features ⁓ just by specify what we want from the back end. And in the back end might have logic in a Rust API or rust for API expose the the raw data or whatever. But in GraphKL resolvers ⁓ layer we can resolve these fields only for this particular UI. So here the front end asks for exactly the field it needs, nothing more, nothing less. The business logic still lives in one place, but the response is titled

⁓ to the consumer's access patterns. This means we can evolve our backend service without breaking other clients and still keep the front end simple and declarative. This is a small example, but the same principle applies across your whole system. ⁓ Good data modeling helps you put logic in the right place, keep your code clean and make your application scalable. So next we will bring this thinking back to our starter project

Instead of just a hard code state, we will actually fetch data from ⁓ two endpoints, the API slash me for the user and ⁓ the API board for the board and the cards. But here is a twist. If we just store those response as is, we'll quickly run into duplications and inconsistency. ⁓ for example the user ⁓ might appear on a board and ⁓ on a card. But if we update them in one place

Sneha Mehra (00:07:13)  
The other doesn't change. This is why in the next lesson we will introduce data normalization in the front end. We will take those API responses and model them in a way that keeps entity consistent across the app. And that's where data modeling really starts to pay off. Not only help you ⁓ understand the domain, but also give us the right structure to build scalable UIs.

—----------------------------

Sneha Mehra (00:00:00)  
Today we'll dive into one of the most dangerous and honestly most common security issues in modern web applications. Curse site scripting or XSS. We'll look at how XSS attacks actually work, why they are so dangerous, and what makes them tricky to prevent. I will even show you a live demo and attack happening right inside our application. By the end, you will clearly understand how attackers can run code inside your

user browsers and why we need stronger defense like content security policy to protect real world applications. Let's jump in. So what exactly is cross siting scripting? In simple terms, XSS happens when an attacker manages to inject their own script, JavaScript code, into your web application. That code then runs inside another user's browser as if it was trusted code from your site.

Let's make this concrete. Imagine you are building a task application like what we are doing. Users can create cards with titles and its descriptions. Pretty harmless, right? Now what if one of the titles looks like this? At first glance, that looks weird, but maybe not as dangerous. But wait until you see what happens when another user loads that card. ⁓ I have prepared a demo mode in our application. When I visit

XSS demo equals to vulnerable, it switched to a version that uses on SIF code. See this red banner on the top? This means we are in vulnerable mode. And for the card, ticket one, it's highlighted because it rendered using dangerously set inner HTML. ⁓ Now let's reload the page.

And immediately you can see a pop-up shows up, XSS attack successfully. And if we look up the dev tools, and you can see that in the console, attacker can catch this ⁓ cookie and save that to their own instance, and they can use this cookie to impersonate this ⁓ particular user. So let's break it down what happened step by step. So the malicious HTML is injected into the DOM, the browser passes it and sees the image tag.

Sneha Mehra (00:02:16)  
It tried to load the image from the source equals X. The image fails to load, which triggers the onArow handler, and that handler runs JavaScript, the attacker's code. I have defined a variable code malicious payload in ⁓ XSS demo. So basically what happens is this malicious HTML is injected into the DOM. The browser passes it and see the image tag here, and it tries to load this X. Obviously it doesn't exist, browser will code this on arrow handler.

And that's why we have these logs printed out. And in this demo we just log the cookie. In real attack, it could send your session cookie to a attacker's website. And once that happens, the attacker can impersonate you, read your data, even perform actions on your behave. And here is the scary part, it spreads automatically. So one single malicious card can compromise every user who views it.

You might be wondering, can we just block the image tag? Unfortunately, no. Image tags are totally legitimate. We use them everywhere, and that's what makes this attack so sneaky. Here is why it's clever. First, it doesn't use script tag, so it bypasses basic features that only looking for boost. ⁓ Second, it always triggers because the image source is invalid, the own error event fires every time.

And third, there are endless variations of these tricks. You can use the iframe, ⁓ SVG, even div or body. And even if you block one pattern, attackers just switch to another. And if we look at how the malicious payload is actually used in our component, you will find that in the card, ⁓ we are checking if we are in demo mode. So for example, we are using the ticket one as a demo card, and when the mode is vulnerable, we're

Using the dangerously set inner HTML ⁓ to the display title, which is a malicious payload, and that will showcase the actual arrow. So you might be wondering why would React even provide an API code dangerously set inner HTML? That sounds like asking for trouble, right? Well, sometimes you actually need it. Here are some legitimate user cases. So for CMS or markdown output.

Sneha Mehra (00:04:36)  
You store user content in markdowns and run it as HTML. You need the paragraph that links the ULI tag to appear, not just the skip text. And also sometimes you use the server-size rendering HTML snippet. You backend provide safe, a pre-synchronized fragment, like documents or changelogs. They all could include HTML and third-party widgets, some abandoned or email templates come in as HML as well.

And there are also cases like a legacy content or what do you say is what you get editors when you migrate old HTML content or handle rich text from content editable field. In all these cases you have to inject real HTML and in React, that means you need to use a dangerously set in the HTML. So it's not evil by itself, it's just ⁓ dangerous if you are not careful. Let me go back to the demo again. So you can see ⁓ in a vulnerable

mode you can see this ⁓ dangerously set inner HTM is used. And if we use protected version, which will not use the dangerously set inner HML, so you can see the actual content is fun bug, image, ⁓ things like this. This is skipted HTML so it doesn't execute. So by default it's SIF. And this happens because in a SIF mode we are using the display in this brick it and the React will

Handle the escape for us, that happened automatically. And ⁓ only when you use the dangerously set in the HTML at that point the browser will need to pass the HTML and ⁓ trigger the malicious code.

Sneha Mehra (00:06:18)  
So, to recap, XSS happens when untrusted content like user input end up being executed as code in someone else's browser. It's sneaky, it's powerful, and it can spread very fast. And even legitimate APIs like dangerously set in HTML can open the door to it if you are not ⁓ sanitizing properly. In the next lesson, we will talk about how to defend against these attacks, ⁓ specifically using content security policy.

or CSP in short, and HML sanitization to keep your users safe.

—-------------------------

Sneha Mehra (00:00:00)  
When a network is slow, every second feels longer to the user. If all they say is a blank screen or a spinner, frustrations grows quickly. A better solution is to use a skeleton screen. Instead of saying wait, we give users a preview of the layout, green boxes where the content will appear. That way they feel progress immediately, even though the data hasn't arrived yet. Today I will show you how the skeleton pattern works ⁓ in our board application.

And we will test it on the simulated slow network. Skeletons follow three simple principles. So firstly, we need to match the final layout. Placeholders should have the same size or about the same size and shape as the real UI. Secondly, we will keep it lightweight, just rectangles or circles, no completed details, because we don't want the skeleton itself to slow down the UI. And third, we will add subtle animations, a softer pulse or shimmer.

makes it feel alive like it's in progressing. And here's a minimum example of our top bar component. So we have this top bar, there is a user editor here. Well in loading will only show a simple grain cycle here. So if I refresh the page, you see here there's a small cycle. I will do that again. So you see here. I have already added a few delays in the back end so the loading will take some time. So the code looks like this in the top bar

We are checking if the user exists or not. If the user hasn't returned from the server side, we'll simply show a top bar skeleton. And a skeleton is a very simple component, it's very static, there's nothing fancy in it. it has the same layout as a header and has a div and also a avatar. And if you look at the actual ⁓ top bar component, it has the same structure which has a header, a div, and if it's hydrated we will show a more fancier ⁓ feature.

Otherwise it's a simple div. We don't have to show the image itself, we only show this ⁓ container div, which is the 10 byton ⁓ cycle placeholder. And if we go to the top bar skeleton, you can see ⁓ the header div and the 10 byton placeholder. So we have a simple animate pose class added. So if we look closely at the UI, you will notice there is a simple shimmer kind of thing here happening when we're loading.

Sneha Mehra (00:02:26)  
So the animate post utility gives us a simple breathing effect. If you want a shimmer, you can add a gradient overlay and animate it across. The important thing is the skeleton has the same dimension as the final after. So when the real image loads, nothing jumps around. And similarly, we will add a few more skeleton components for the rest of the application. So the perceived performance of the whole application flows faster. So now if you notice the board.

section is still empty when we're loading and there's the ugly search bugs showing up ⁓ as it's loading. So we need to fix that as well by adding a few more skeleton components. So in the board page I will check if the a call moder exists or not. If it doesn't exist we will show a fallback. So basically we'll use the same structure as the actual component and use two a skeleton component. The top bar will be these

part and the rest of this one including the control and the board view itself, the columns cards, ⁓ they are all in a board skeleton. The board skeleton is a little bit complicated because that's what a UI is. We need to make sure the column name ⁓ is in the placeholder and the ⁓ column itself will place a few random placeholders in the board. So in the first column we'll have three cards and the second one will have two and third one will have one.

We can always like adjust this ⁓ structure to make the layout more dynamic or kind of realistic. And then when we load the page you can see the page is more dynamic and ⁓ the layout is more accurate ⁓ as we load. And now you notice there is a small jump here, that's because we have more cards in ⁓ actual scenario. We can add this like a one, two, three, four, five, maybe ⁓ to simulate this kind of structure.

or the layout change. So for example if we go to our code here we just ⁓ add a few more and when we load you can see the screwbar is even correct. The skeleton pattern is used quite often in other places. Like in Jira product page ⁓ if I slow down the nightwork to fast 4G, let me do a refresh. And you can see the skeleton pattern ⁓ is applied in a

Sneha Mehra (00:04:52)  
lot of places. If you go to summary, you see these ⁓ dashboard has also got the skeletons and go to list there is a least skeleton as well. ⁓ So the good thing is there is no blank page, no spinners, no layout shift, just a very smooth transition. This is the core of the perceived performance. Even though the actual request take like two seconds, the experience feels responsive. Here is a few a quick tips.

Don't over animate the UI. You should always keep the motion as subtle as possible. You don't block the whole screen, show the header and navigation immediately ⁓ when you can, and skeletonalize ⁓ only the dynamic content. You also should time box the skeleton to avoid flashes. If your data usually arrives in like ⁓ one hundred ⁓ milliseconds, consider delaying the skeleton slightly so it only shows for slower responses.

And you should always test them under slotling. That means you always simulate bad network conditions to see the real user experience. And you can use that in DevTools to slow down the network request and then test it before you release. Skeletons are one of the simplest ways to improve perceived performance. They don't actually make the network faster, but they make the wait feel shorter. And that's what really matters for user experience.

—------------------------

Sneha Mehra (00:00:00)  
Welcome to Front End System Design Essentials. In this first module, we are going to set the foundation for the entire course. We'll start with a simple start project, a lightweight taskboard application similar to ⁓ Jira or Trello. It's not perfect. In fact, I have intentionally left a few bad smells because this project will be our baseline. We'll keep coming back to it, reflecting and improving ⁓ step by step.

Throughout the course you will see how to take this rough starter and turn it into a clean, ⁓ scalable and maintainable front end system. We will explore topics like request handling, pagination, data normalization, optimistic updates, code splitting, catching and more. And here's the key. This isn't a work-only tutorial. You will set up the project locally on your machine, ⁓ work alongside me.

And experience firsthand how design decisions ship ⁓ rail applications. So before we dive in, make sure you have Node.js version 20 or higher installed. And also you have comfortable running a React ⁓ Plus 5 application locally. If you don't check the setup instructions in the Ripple, I have included a short guide there. This way we can focus our time in this course on the system design, not just the installing tools.

Sneha Mehra (00:01:29)  
So on the UI it's very similar to what you have in Jira or Cello. So basically it's a board application. a board has columns ⁓ and ⁓ in columns it has ticket, and on ticket you can have a ⁓ assignee, and then you can select a user from the list and then assign that to a ticket. And also ⁓ we might add the features like you can move the card from one column to another or change the status ⁓ in some other way.

And also you can change the ⁓ assigning to another one, ⁓ move across the board, and you can also search the board by ⁓ the ⁓ title of the ticket, like ⁓ implement a user list, ⁓ and also you can switch to a list view ⁓ from a board view. Then ⁓ you probably will add a feature ⁓ for like only see the issues that assign to a particular user. And on the top bar ⁓ there is a user after

Which indicate the current user, and we can go to the profile page, making some edit there. ⁓ And then we can go to the board again to see some ⁓ date ⁓ changing. And as a starter project, ⁓ it's has a very limited ⁓ functionality. Over time we'll enhance the features. You can see how the decision is making, like what's the person accounts for each option. for example.

if we want to introduce the pagination to this application, we'll need to consider both the cursor based ⁓ offset based pagination with the processing accounts. And when we talk about the data consistent or data modeling, ⁓ for example, the user in here ⁓ will have the similar structure in avatar here. So how do we maintain the consistency between this data and how do we do the data fetching and data mutation? Well we adding

⁓ new features on top of this baseline. So I have already put the code into ⁓ a ⁓ repo ⁓ in my GitHub so you can download it and there is a clear instruction in the ⁓ README so you can follow along and set it up. And if you run npm rundev you should be able to see ⁓ a ⁓ same page ⁓ layout as this one in your local

Sneha Mehra (00:03:52)  
And from there we can gradually make the application more real life, ⁓ by introducing more system design concepts.

Sneha Mehra (00:04:02)  
So the start project is simply a React application. ⁓ We are using White as a build system. ⁓ So ⁓ if you look at the README, you will notice that we have a few command line tools like npm install, npm rundev. And if you launch that rundev, after you install all the dependencies, you will be able to see a application running on your local browser. And if you go to your browser, you will be able to see the application running on your local. And then if we look at the project structure,

You will notice that there is a mocks ⁓ folder which ⁓ mocks all the network layers. And in handlers ⁓ TS that is where we define all the API endpoints. ⁓ this ⁓ user endpoint, the board endpoint, the card endpoint as well. That's where all the data comes from. And for the UI part, we start from the main.tsx. And at the very beginning, ⁓ if we are using in a dev environment, we'll use the MSW to

Intercept all the network ⁓ request and the return ⁓ correspondingly whatever we are fetching in the application. So then we go to the application, ⁓ we have this mock-up, and in mockup I have top bar which is the avatar, and then we have a board. Board need ID, and if we go to the board it will use the effect to fetch the board data by ID ⁓ and ⁓

The inside the board ⁓ we have a board control which corresponding the ⁓ search bar and the switch for different views. You can you can switch to ⁓ border view or list view and then there will be ⁓ either the border view or list view correspondingly. And if you go to the border view, there will be a big map ⁓ through all the columns and for each column there will be a title and inside the column there is ⁓ card list.

And for each card, ⁓ it has a title of the card, there's ID, there's a pop-up for the assignee. And if you click the button, ⁓ basically that's what this part is. And it will pop up the selector. ⁓ and ⁓ inside this ⁓ popover there's a content area and there's a user selector component. And if you go to that it will fetch the users.

Sneha Mehra (00:06:29)  
So every time when you click the icon here, you will see a list of the users and you can select from. ⁓ As you can see, we are using this very simple, naive way to do the data fetching and we are not handling the the arrows or the loading standards. ⁓ that is intentional, so we will enhance them over time in the following lessons. And you may notice that the old components are just ⁓ put in a components folder.

We probably can organize them even better to add a few sub folders. But as a start project, we haven't done that yet. And again, in the following lessons we'll talk about the ⁓ points and cons for how do we ⁓ organize them ⁓ properly in a real-world scenario. ⁓ So for now, don't worry about making everything perfect. These imperfections are here for us to learn from. By the end of this course, you will have not only built

a working application but also developed the skills to design furniture systems with confidence. Alright, let's move on to the next lesson.

—--------------------

Sneha Mehra (00:00:00)  
So to make our application more realistic, we would like to enable the user to edit the card details ⁓ from a model dialogue so for example if we ⁓ want to edit the title or the description for this card, we want to click the card and it will pop up a model dialogue. Then we can change the title and the description or even the assignee of the ⁓ card, and then we have a button to save.

And in that scenario we will practice ⁓ what we learned in the previous chapters about the least loading, the skeleton pattern, and the mutation. So let's get started. So firstly we'll need to install the Redix model dialog, because we are using that headless component in our repo already. Basically, we need to ⁓ npm install a Redix UI React dialog. ⁓ And ⁓ after installation we can create a model based on that.

So basically it will be something like this. We cre click a button, which is our card in this case, and we click it, we will launch this ⁓ module dialogue and you can use keyboard navigation to you know make the change and then save. And ⁓ the usage is relative simple. So basically we need to import the dialog from the RedX UI and we need to define a dialogue route. So basically there are two ways of using this ⁓ module dialogue. You can use that with the trigger.

Like you define a button and then you click that and then it launched the model dialog. But for our case we want to use a card. When you click the card it will open up this dialogue. ⁓ So it's more like a programmatically make that happen. So we will need to create a component called editing dialogue and we have this portal and we will show the title and description as well. ⁓ But the trigger in our case will be a little bit different than the trigger defined here. So basically that means in our case it will

Be doing the on click and here we'll set the ⁓ model open to true and when that state comes to true and then we can programmatically open the model dialog from this place. Alright, let's start with creating a new folder on the card. Now let's call it edit ⁓ model here and then instantly we can have a card edit model ⁓ and ⁓ in the card

Sneha Mehra (00:02:26)  
We just need a new state to maintain the it's open. ⁓ It's model open ⁓ set model open ⁓ and basically use state ⁓ as false. So by default it's obviously it's false and we will need to register a handle ⁓ click, let's say. Basically we will need to ⁓ set the model open to true and then at the bottom of the card we just use that flag to is

⁓ model open. ⁓ If it does, we just open the ⁓ card edit model. Let's do that export const ⁓ card edit model. ⁓ obviously we'll need to implement that later ⁓ model ⁓ and then in here we will use that card ⁓ edit model ⁓ right so

Ideally we should be able to see that when we click this articular to set it at true. Let's do that here. ⁓ Let's say on click ⁓ is handle click. So pretty much like that. ⁓ when we click this articular, ⁓ that will be the model shows up inside the card. Let's give it a try. So in the browser we have this card shows up and when we click it, ⁓ it's just shows a model and we click it again. it's not highlighted, so we need to basically

How could ⁓ open open and ⁓ if we click that it shows model and we click again it hide it basically. Obviously we need to use a Redix model in here so we can make it real. I have already implemented the card edit model in details, but if you want to give it a try, ⁓ feel free to pause and ⁓ just type whatever the model that should be looked like using the Red X UI. ⁓ But I will skip that typing part.

because it's not the main focus here. Welcome back. So my implementation here is using the Red X rector dialog. The UI itself is very simple. We're using the user selector component and we are setting the title. And I have added a description editing, even though we don't have the description yet, but we'll make the change in the back end as well so we can save ⁓ description as part of the card. ⁓ So basically it's a form, we have a title.

Sneha Mehra (00:04:50)  
I have the description and we have this user avatar you can pick up and then you can save that change to the backend. And for the actual saving we are sending a patch request to cards and we put that ID and string by the ⁓ payload. So basically we'll have the title, description and assigning to the back end so backend can figure out what needs to be updated. And once we have this change, we will set the ⁓ on open change to false.

So we'll close off the module dialog ⁓ after the update is done. And in the card, if model open, we will open up the ⁓ card edit model, which is which is the importing from the edit model, and we should be able to see the pop-up correctly. let's go back to the browser ⁓ and if we click ⁓ one, it can open up this ⁓ editing dialogue ⁓ and obviously we don't have this handler yet.

And if the mode open will show this model dialog. ⁓ So if we go to the UI, we click, we can see this mode dialogue. But there is a small bug here. When you click anywhere, the mode dialogue will close. ⁓ That is because ⁓ we are not handling the events correctly. And let's go back to the card and fix that. So in our article unclick, it will handle the click event and it will flip the model open status, which is not accurate because when we edit in

We don't want to handle this while the model dialog is open. So we'll need to skip that if we are editing. So let's do that. That means we will need to handle the ⁓ cases where we don't want to close the model dialog. We are going to fix that by saying if the tag name is not a button, if we are not editing, we're not deleting, we just we will set it as true. And ⁓ in that case ⁓ if we open this up and if we are inside this ⁓ area, ⁓ we just skip that check.

So we can still click inside, it will not trigger the close. But yeah, obviously we can close that by console. And now let's give it another try. It will do the update. Let's say instructional document for onboarding ⁓ and so on. Let's give it another user, ZAM again. ⁓ save change. Yeah, we have already successfully updated this one. It has the title correctly set and also the avatar changed.

Sneha Mehra (00:07:14)  
That's great. Before we move on, I would like you to implement the lazy loading plus the skeleton pattern for the model dialogue. ⁓ Basically we need to split that model dialogue into a lazy loading ⁓ component, wrap it around a suspense, and the fallback will be a skeleton. You can pause the video and ⁓ start to implement our weight.

Sneha Mehra (00:07:38)  
Alright, welcome back. I have created a card edit module skeleton and in the card component I'm lazy loading the card edit module from the edit module ⁓ card edit module and then on the bottom ⁓ we have a suspense and it will fall back to card edit module skeleton ⁓ when it's loading and then we will ⁓ render the card edit model and then if I go back to the UI when we click ⁓ you can see a skeleton shows up or this one. ⁓ it's already loaded, let me do a hot refresh.

So you can see it more clearly a skeleton and if we edit in anything here, hello ⁓ word, ⁓ and it will make the change and ⁓ update the back end.

Sneha Mehra (00:08:23)  
Awesome. So in this lesson we have learned how to use the RedX dialog to implement the model dialogue and we have practiced the ⁓ lazy loading the fullback and also use the existing API, the patch API, to update the title of the card. So now we not only can read the data but also we can mutate, we can change the data in the backend.

—----------------------

Sneha Mehra (00:00:00)  
Welcome back. Now that we have faster behavior focused unit test at the base, ⁓ we're going to add a small layer ⁓ of end-to-end tests. The goal here is different. Want to verify the critical user journeys work from the UI through to the server and back. That way the real user experience all of them. So end-to-end tests are slow compared to the unit tests, so we keep them viewed and focused. We'll cover three flows, loading the board.

assign a user and delete a card. That's enough to catch broken wiring across multiple layers without slowing the pipeline. So because n twin tests change real state, if we don't reset between tests, ⁓ one run can present the next and the flickness creeps in. The fix is a small server only reset endpoint that reloads our mock data. We call it before ⁓ each test and every run starts clean.

So on the server side we added a new endpoint, API test reset, and we will try to reset mock data. And once it's done, and we return to a 200 to the test. So in the reset mock data, we just ⁓ reload the data and ⁓ reset everything. And after the reset, the next test is the ⁓ clear start. Let's now talk about the steps that we're gonna take. And we're gonna use playwright ⁓ for the end test. So first we're gonna install it.

And configure it properly. So let's go to command line to and ⁓ install the ⁓ playwright test. And then we will install the playwright itself. It will include a few headless browsers like the Firefox ⁓ and ⁓ Chrome. And we are going to use Chrome in our test here. So basically it's a browser running in your memory, but it's a real browser, just headless, and ⁓ it will using that browser.

To ⁓ access the URL we provide, ⁓ which is the localhost 5173, and then the test can access our application, the memory ⁓ browser, and verify the elements is on the page, ⁓ and it can interact with the ⁓ page so we can verify if something ⁓ exists on the page or like a user click a button and the word shows up and so on so forth. I've already got the installed, I will skip that part.

Sneha Mehra (00:02:27)  
And after the installation, in the project route, I will create the playwright configure.ts and I basically will define some configurations here. So we just like set up a HTML reporter. So after each run, it can generate a HTML report. And in the report it will contain all the critical information for us, like how many tests we run, how many tests passes, the time span on the test, and so on. And we're going to use ⁓ Chrome and only for the desktop Chrome.

And second, we add a web server command so test can boot the application automatically. And this about the configuration is pretty standard. The base URL is 5173\. And then as you have already seen, we have implemented a restart endpoint on the server and call it before ⁓ each test to reload the fixtures. ⁓ Now let's look at the actual test cases. So in the playwrite config, I have specified the test directory is end to end.

so in the entry I have created board spec.ts and in here we just write the basic test. Before each test we want to restart the API on port 4000\. So in each test case we will have a clean start board and then want to go to the root of the application and we want to wait for the articular to show up. we give it a 10 second ⁓ timeout, which is in most cases enough, and we can use a small

Check here for all the cards to be loaded. For example, we want to make sure ⁓ in the document we have articulars, span. We want to make sure the articular ⁓ is exist and we want to make sure we have more than one card. And this happens ⁓ before each test. And let's look at the first test keys. We should load board and display cards. So we are expecting the page to have a header and we want to see the card exists.

And also we want to see that we have more than one card. And the first card has a H3 header. We probably can be more specific on like what's the title of the first card. I will skip that for now, but I will give you a quick ⁓ look at the what's the filling of the running playwright test. So we go to the command line and run npm playwright test. And the playwright will launch the headless browser in memory.

Sneha Mehra (00:04:52)  
And then run the tests against our application. We're just making sure our application is running here. ⁓ So then we have a report here. We can run MPX playwright show report. ⁓ And then it will launch the browser. And we can see ⁓ three tests are passing. And you can check what is passing and each step how it's going. And also we can run the playwright in the interactive mode. So we just run the ⁓ test in the UI.

mode so that will launch the playwright in a interactive mode. So we can run each test case or the whole test suite. Let's run the first one for now. ⁓ If I click here, it will launch the application and assert the H3 ⁓ shows up. You can see that it passes correctly and if you want to see the second one, this one. ⁓ And it will find the card and click the assignee list.

and ⁓ wait for a bit and then assign the user to correct user. As you can see the end to end test running in the browser and testing the four paths from ⁓ browser to the server side and so on. So with this test we get end to end confidence on the passage user care about the most. ⁓ Because we reset a state each time, so this test we stays ⁓ isolated, ⁓ we can actually add a few smoke tests that catch in

if the server is dead problem before deployment. So let's have a quick look at them. So I have added a few smoke tests that is basically checking if the page is alive. That is no significant failures like the page can load at least. So we want to make sure if we go to the home page, we can see the body at least exist and we want to make sure the critical API works okay. We can have the health check for API endpoints.

And also we want to make sure at least when we go to the board one, we can see a articular shows up on the page, but we don't care about the details of of the ⁓ cards. That should be enough for us to ⁓ validate the survey is still working.

Sneha Mehra (00:07:08)  
So the key insight here is that we should keep the end-to-end test small in number but high in value. And when you write a test, avoid the brittle selectors and prefer waiting for meaningful content over arbitrary timeouts. And if you found a flow become flicky, ⁓ step back and choose a better user visible signal. And that's our approach. We add the reset endpoint for clean isolation, we verify ⁓ the critical flows end-to-end.

And we also run a tiny smoke test ⁓ in CI that can protect the deployment. So with our both layers in place, first unit test and focused end-to-end test, we have built a programmatic safety net that supports rapid ⁓ iteration without compromising production readiness.

—--------------------------  
Sneha Mehra (00:00:00)  
In this module we took everything we have learned throughout the course and applied them into a real growing application. We started by reorganizing the project structure and refactoring the code base so future changes would be easier and cleaner to make. Then we brought the drag and drop to improve the user experience and extended it with proper accessibility so keyboard users can move card just as easily.

After that we introduced multi page using regular router and enhanced those roads with page level lazy loading to keep the application fast as it scales. All these pieces together show that modern front end system design looks like in practice, not just the theory, but the actual patterns and decisions that makes a code base maintainable over time. And with that,

We have reached the end of the content portion of this course. In the next lesson, we will talk about what you can do next and how to continue growing your front-end assistant design skills. And I will see you there.

—----------------

Sneha Mehra (00:00:00)  
Hey everyone, welcome back. In this module, we are going to talk about one of the most important aspects of front end system design the performance. When most people hear the word performance, they think about shaving milliseconds of a load time or squeezing one more point in Lighthouse score. But in practice, performance is really about how your users experience your application. Sometimes a side can be

Technically fast but still feel slow. Other times with the red patterns, a site can feel instant even if some work is still happening behind the scenes. And that's what this module is about. Learning the strategies and patterns that help you design systems where speed is not just measured but filled. So let me give you a roadmap of what we are going to cover together in this chapter. We'll start with rendering strategies. You have probably heard about

Client side rendering, server side rendering, static side generation, maybe even streaming, SSR. We will look at each of these from system design perspective and ask what does it do for performance and when does it make sense to use it. Then we will go hands-on with SSR in action. Instead of just talking about them, we will work through how server rendering works in our ⁓ board application.

and see what actually changed in terms of metrics and the user experience. After that we will shift gears to something more subtle, the perceived performance. This is where design patterns really matter. Even if we cannot make the network faster, ⁓ we can change what the user says and when to make the experience feel smoother. One of the most powerful ways to do this is skeleton pattern. You have probably seen it in apps like Facebook, LinkedIn, Twitter,

Where placeholders appears ⁓ instantly while rail content is loading. We will implement this in our application and see why it matters. And from there we will look at how to get ahead of user with perfection. By warming up the cache before someone even clicks, we can make the next page or interactions feel instant. And as always, we will look in ⁓ how to do that in action, how to implement that in as part of our ⁓ board application. Then we will move to lazy loading.

Sneha Mehra (00:02:26)  
Which is the opposite idea. Instead of putting everything in advance, we hold back what isn't needed immediately. So the initial load is lighter and faster. You will even have a checkpoint exercise here where you will implement a simple lazy load yourself to practice the concept. And finally we'll talk about how to measure performance because in system design it's not enough to guess. We will look at tools like Lighthouse, ⁓ Web Vitals, and React Portfolio.

and see how to check whether the strategies we have learned are actually improving the numbers that matter. By the end of this module you won't just know the collection of ⁓ tricks, tips, you will have a way of thinking about performance as a design decision. You will know how to look at a feature, evaluate trade-offs, measure the results, and choose the right strategy for the work. So ⁓ let's get started.

—-------------------------  
Sneha Mehra (00:00:00)  
Most of front-end applications rely on data fetching, and search ⁓ is one of the most common places where things can go wrong. Today we will look at a subtle but important issue. What happens when your application sends multiple overlapping requests and they don't come back in the right order? This is called risk condition and we will fix it with a very simple tool build write in the fetch API, the board controller. So let's start with our board application.

It has a search box ⁓ and let me find the cards by title. Behind the scenes every keystroke triggered a fetch to the ⁓ mock API. So now if I get to the search box and tap very fast into the IN IST. ⁓ You see there is a flash here. ⁓ If you notice the down column, let me do it again. ⁓ There is a quick flick like ⁓ it was one and then turns out to two cards in the down column.

So what happens if I open up the network tab ⁓ and ⁓ do it again, and you can see when I type ⁓ i it will send the query with ⁓ q equals i ⁓ and if I do n, it's send the in request, and if it does s it will ⁓ in s ⁓ and the t. Right? That's pretty normal, and that's okay when I type

pretty slow, but if I type ⁓ fast enough, let's say f type INST and you can see the request is sending out in order, but it comes back out of order. the reason is that the ANS request takes longer. That's because I have ⁓ dedicatedly in the Mocha API ⁓ set up a significant delay. so this INS request returns in seven hundred milliseconds

Well the others are just like one hand or three hundred. ⁓ so if we look at this bar here, ⁓ you can see that it's a green bar here ⁓ in this area. And if I go to this one, you can see it's ⁓ the timing is much shorter and it returns at four hundred mark and this third one returns at nine hundred and fifty milliseconds. That means the

Sneha Mehra (00:02:25)  
response of this in st written faster than the ⁓ second last one. So that means the response will re-render again with the stale data of the ANS. That's why ⁓ if I include this one and if I do a refresh and if I do a search ⁓ ANST, ⁓ you notice that it's the insights here, which doesn't include the INST, it's just the INS

I but is a incorrect result from the ⁓ back end. That's the risk condition. ⁓ Whichever request resolved last wins, not the one the user actually wanted. And if we look at the code inside the board where the query the search happened, ⁓ so we have prepared a parameter and we ⁓ append the search ⁓ which is the content in the input box into this parameter ⁓ and then we will do a fetch.

And when we have this JSON, we will ingest the word just like what we did a normalization ⁓ lesson. And ⁓ because there is no control here, we just re-render whenever we need. On the surface it is very hard to spore the the problem because if you read that in sequence, ⁓ you cannot see the problem easily. So how do we solve this? The browser already gives us a tool, a board controller. So it lets us cancel a request, a fetch request ⁓ that's no longer needed.

This is how it works in Planned JavaScript. So firstly we will need to create a abort controller object. And then when we do the fetch, ⁓ we will put the additional option for the fetch call. Basically it will be signal, ⁓ the control.signal and ⁓ then we can do the normal fetch. ⁓ And at some point we can call the controller abort ⁓ that will cancel this ⁓ ongoing fetch request if it's not resolved yet or like either in pending ⁓ or

Backend is not responding or whatever, when we call the controller abort, that fetch request will be cancelled completely and we don't rely on the data at all. The fetch will reject with an abort arrow. This means you can start a request and if you know you don't need it anymore, maybe the user typed another calculator. ⁓ you just abort it. In React, the use effect ⁓ cleanup function is the perfect place for this ⁓ call.

Sneha Mehra (00:04:53)  
Every time the search changes, React runs a cleanup for the previous effect ⁓ before running the new one. So if we can abort the cleanup at that point, the previous in-flight request is ⁓ cancelled automatically. So basically we need a few steps here. So firstly we'll need to define a controller, which is a new abort ⁓ controller. And we can use this controller now ⁓ in our fetch here and pass it into signal.

And the control signal. So this way we hooked it up to the fetch and we can then do a ⁓ cache here. So we catch the arrow and we'll cancel log out the ⁓ arrow, whatever it is. And at the bottom of the user effect block, we need to return a callback that will do control about. ⁓ you can also put a reason here. We can see that ⁓ query changed.

⁓ and then ⁓ whenever it's on mount, basically when we ⁓ change the search, it will about this previous in-flight query ⁓ and ⁓ the previous cache will catch this ⁓ arrow and it will see the result in the console log. Now let's go back to our browser and if we open up the ⁓ network inspector. So now if we do ⁓ inst ⁓ and you can immediately see that when I type in.

The previous one is canceled here. And when I type I in ⁓ S, the previous one is canceled. So we always cancel the previous request and rely on the whatever ⁓ in the latest one. So even ⁓ that one is slow, it's already canceled, so we don't have to worry about the response return at some point after the ⁓ latest one. And that way we can completely ⁓ resolve this problem. So to recap.

Multiple in-flight requests can cause risk conditioning in search, and a bot controller lets us cancel old requests we don't need anymore. In React, combining it with the use effect cleanup is the clearest way to handle this. In the next lesson we'll go ⁓ one step further. Instead of just canceling the old request, we will reduce the number of requests ⁓ we made in the first place using D bounce and throat.

Sneha Mehra (00:07:22)  
And this is another key ⁓ technique in front end system design for data fetching.

—-----------------------  
Sneha Mehra (00:00:00)  
So I have modified the user endpoint to support pagination. So now I'm accepting a additional ⁓ parameter from search query. so basically we need a page and page size. ⁓ the page will be ⁓ the current page, ⁓ start from zero, and then there will be page size, like how many items on the page we want to return to the front end. ⁓ And ⁓ once we got the data and we will slice it.

from start to end and then we will return the items just like before. But in addition we will add a page info section and we will put a total number of the items we have, ⁓ what the current page is and what's the page size. And we will add a addition flag, ⁓ basically a boolean flag saying ⁓ if it's reached the end, the last one. And ⁓ then ⁓ in the front end we will

still call the ⁓ API slash users ⁓ inside our user select component. And basically we will ⁓ need to fetch initially when the component mount. And also we will need to refetch when we are reaching to the bottom of the ⁓ selector. So we're using the React Selector component. It has option here basically ⁓ saying a menu scroll to bottom. So when you user scroll

inside this drop down list and reach to the bottom we will call this function. In your case you probably will need to either use the instructional observer or some other similar mechanism to trigger the refetch. So when we do the ⁓ fetch page ⁓ inside the component will maintain a page ⁓ state at the very top. ⁓ so we will start from zero and as we scroll we will increase that number

set page basically. ⁓ once we have got the data from the fetch page. So this is an updated version of the ⁓ fetch and we will prepare the page info, the page number and the page size. And then we will do the ⁓ fetch and we once we have the data, we unpack the items ⁓ into the user list and then we have the page info ⁓ and then we'll maintain this page info

Sneha Mehra (00:02:28)  
and use that in our ⁓ UI component. And ⁓ for the other rendering part is still ⁓ the same as before. So options ⁓ is what we have to show in drop down list. So if it's the first page we simply use the response from the back end. ⁓ otherwise we'll append this item from the back end to the previous list. ⁓ So we have the four ⁓ user list. And if we open up the application in browser

And do a refresh. ⁓ We can see here if I click any of the ⁓ user selector here, ⁓ you can see it affetched a user ⁓ endpoint that page 0 and page size is 5 at the moment. And we can see five users in the list. And if we scroll ⁓ and you can see a scroll bar increase as well, and then there is a new page request sending out, it's asking for page one.

And if we keep scrolling and it will reach the page two, ⁓ the page size is five, and we do it again, ⁓ it will send out to page three, because we only have twenty users, it's already reached the end and ⁓ it will not fetch anymore because we have this hazmore flag here. If it has more we will fetch. And the hazmore is set from the back end ⁓ in the page info.

if you remember that we have this handler here and if the n is less than total is set as f true, otherwise it would be f ⁓ force. And if we don't have more atoms, we just ⁓ stop fetching and that's exactly what it behaves in the UI. And for individual response gear, if I resize that in here, you can see the result is formed as items.

which is a user object itself and also it has a page info, has has more page size, page number and total size. As we scroll ⁓ to the bottom ⁓ latest one here, ⁓ it doesn't has ⁓ more items, so it's four. So we don't have to fetch more users ⁓ in the last page. And yeah that's pretty much about the very simple implementation of the page nation.

Sneha Mehra (00:04:56)  
with the React selector. And in your case, you're probably using other library or other UI patterns. For example, if you are ⁓ searching and showing the search result in a in a page, you probably need to show the ⁓ page navigation and pagination component, like one, two, three. And when you click number three, you ⁓ set that three as the page ⁓ number and send a request to the back end and get the

response correspondingly. And in other cases you might use cursor based pagination. For example, like ⁓ a endless ⁓ infinity ⁓ feed list. When you scroll, you will always maintain the cursor and then set that cursor to the next one you get from the response. And when you get the the item on the next page, you append that to the ⁓ previous list. And you always maintain this ⁓ big list until you

reach to some logic limit. For example, if you there are too many items on the page already, you can stop and showing a button and ask a user to explicitly click that button to load more pages. Otherwise there might be too many elements on the DOM and that will slow down the page significantly unless you are using some other techniques like ⁓ visualization. But yeah that's pretty much about the pagination. And in next lesson we'll look into the caching front end

And we want to implement a very simple version of the cache to manage the request and response. And later on we can replace this with a tenstack query to make the request management and caching more robust and product ready. I will see you in the next lesson.

—----------------------

Sneha Mehra (00:00:00)  
Data fetching plays a huge role in front end system design. Any meaningful application needs some kind of data from the back end. And once we have the data, we need to manage it, keep it fresh, and run data consistently on screen. But this is not as easy as using Effect. There are a number of problems we need to watch out for. First, risk conditions. When multiple requests are in flight, responses can come back out of order. The user ends up

saying stale or incorrect results. Second, overfetching and on fetching. Sometimes we grab ⁓ way too much data and waste bandwise. Other times we fetch too little and have to make multiple requests to just show a single screen. Third, performance issues. Every request costs time. If we fire off requests on every key stock, the application slows down ⁓ and the back end gets overloaded

and the experience feels broken. Fourth, data stillness. even if we fetch the right data once, the world changes. ⁓ users expect their view to stay ⁓ relatively fresh. This means we need strategies for caching, ⁓ invalidating and refetching at the right time. And finally consistency. Different parts of the application might depend on the same data ⁓ without purple management

Our component might ⁓ show outdated information while another shows ⁓ something newer, creating conflict views of reality inside the same app. In this module we will cover the fundamentals of handling data well. First, ⁓ request management, ⁓ how to avoid risk conditions and reduce wasted request with techniques like aboard controller, debouncing, and throat. Then pagination.

both offset based and cursor-based ⁓ pagination and how to wear them into the UI. After that we will talk about caching and ⁓ staleness. ⁓ we'll start by building a simple cache by hand and then we'll introduce React query as a product ready solution. And finally we will zoom out and look at data fetching ⁓ from a system design perspective, how to design EPIs and front end contract ⁓ that skill ⁓ clearly.

Sneha Mehra (00:02:24)  
But in this module you will not only know how to fetch data, but also how to manage it ⁓ effectively, ⁓ keep it fresh and design your front-end systems to handle real world complexity.

—--------------

Sneha Mehra (00:00:00)  
Alright, now that we have the how to add a card with the for mutation cycle from the user action to the API call and to updating our normalized store, it's time for you to put this into practice. As a checkpoint, I'd like you to implement deleting a card. The floor is almost identical, just ⁓ in reverse. We want to trigger a delete ⁓ request to the backend. If it succeeds, ⁓ remove the card from the ⁓ both.

cards by ID and the corresponding columns ⁓ card IDs. If it fails, make sure to handle error gracefully or you can just console log for now because we will talk about error handling in the coming modules. This exercise will give you a hands-on practice with the same mutation loop we just built and prepare you for the next big step, optimistic updates and real-time updates. So go ahead give it a try.

And once you have done, we will dive into how to make mutations ⁓ feel instant and responsive. So you can pause the video now and I will show you my solution after this.

Sneha Mehra (00:01:13)  
Alright, how's it going? So I will show you ⁓ quickly of my implementation. So basically we'll have a midball button in here and there's a archive card action ⁓ for each of the ticket. And if you click the button it will delete the card from that cards section in the store and it also will update the cards IDs for this particular column.

So basically we'll trigger action that will update ⁓ these two places inside the board context. So let's have a look at my implementation. So basically, I have added a new delete API. So for API slash ⁓ card slash ID, ⁓ I will get the ID from the parameter and then I will use a board ⁓ from our ⁓ board JSON. ⁓ this is our database mock. So here ⁓

we'll for loop through all the columns and find the cards ⁓ for that particular ID and if we can't find it we will ⁓ splice it and then we mark the removed to true. And if we cannot remove it for any reason, we just say ⁓ you know four four ⁓ the card cannot be found and then we will return ⁓ toll four ⁓ just to indicate that the resource is deleted.

And then we will correspondingly in a card ⁓ have to define a new ⁓ handler ⁓ for the deletion. So basically we'll send a delete request to this new added endpoint. And if the ri response is not okay, we just ⁓ throw the arrow and then ⁓ print it in the console log. We will talk about the error handling in the following ⁓ modules. But for now we just assume the deletion works well and then we will call the remove card.

action which is again from the board contacts. If we go to the board contacts ⁓ goes to the remove card. So we basically we'll unpack the cards by ID list and we get rid of this ⁓ the one that we want to delete and ⁓ just receiving the rest cards ⁓ to the new cards by ID section and then also we want to update the columns

Sneha Mehra (00:03:36)  
as well to remove that card from the card IDs list in a particular column. And that will trigger the application to re-render ⁓ and then the card will be ⁓ removed from the ⁓ both places. So if we ⁓ go to our application again ⁓ in here let's say hello ⁓ word we add this one and then let's add another one ⁓ and in here we can simply do ⁓ archive card

I have this one or maybe some others. don't worry, we won't actually delete any of them. It's just you know the mock data. And in the console, if we delete this one, you can see that API slash cards ticket one ⁓ deleted and responsive is two or four, no content, that means everything goes well. And if we ⁓ add another card here, hello new card. ⁓

And you can see the post which is the feature we added in last video ⁓ about the API slash card is created ⁓ with ⁓ new ID ticket eight. ⁓ and if we ⁓ do it again to do the deletion, ⁓ you can see it's get removed as well. Alright, that's about the deletion. ⁓ And in next lesson we'll look into a ⁓ small change

For assign a user to a ticket. ⁓ So currently we have this ⁓ user list. ⁓ You know, we added the pagination to the ⁓ list, but we haven't hooked it up with this ⁓ card. So now if you assign a user, nothing happens. ⁓ In next video, we're looking to ⁓ implement it with the optimistic updates.

—-------------------  
Sneha Mehra (00:00:00)  
At the end of the last video we have made some good progress on the ⁓ storage, normalized data and use that data to render the application. But we accidentally ⁓ introduced the defect into the ⁓ search feature. So ⁓ because our user is all from the cards that has the assignee, ⁓ that means we will do a search here. As we type there will be less users available

And in that ⁓ last ⁓ user list, if we don't have this login user, because we hard coded it in ⁓ start of the application, ⁓ so the top bar will disappear. So for example, if I go to the search box and type insert ⁓ or anything that doesn't exist, you will notice that the top bar is ⁓ disappeared. That is because in our code we are hard coded the ⁓ user in the top bar. So if we go to a top level component, the campaign mockup.

⁓ the current user is get from the state user's ⁓ number two with the ID number two. And if the user doesn't exist will ⁓ disappear the top bar. We just make sure the top bar always shows up ⁓ initially. So this is definitely not a correct assumption, especially when the users ⁓ our user if we look at the ingest board, when we do the normalization here, our board are all from

The assignee ⁓ field, we kind of rely on the cards that has assignees and we get the assignee into a list of users. So that has this ⁓ potential defect. So imagine that all the cards don't have the assignee, the top bubble never shows up. So this is a defect here. So ideally we should fetch another endpoint like a current user endpoint and get that user, save that into the same store. ⁓ And when we type or search, we make sure that.

User's ⁓ list or object will never be cleaned because we need the data from that user object to render the top bar. ⁓ Or we can have a dedicated section in a store to say this is the current user as reference to this ID, and we have this ⁓ object in the users and we will never clean up that user because it's a current user. So to fix the problem we need two steps. One, we need to get the current user from

Sneha Mehra (00:02:26)  
a dedicated endpoint. And secondly when we do the ⁓ search, instead of clean all the things in a store, we just clean the column section and we don't touch the users ⁓ section. So ⁓ the columns will always reflect on the UI and the users are always ⁓ persistent in a local store. So let's make the change. I have fixed the issue already in

the ⁓ top level component. So now we remove the hard coded user from the store kind of selector. So initially we will have two requests sent out. So we will fetch the user board ID too and also we will get the board for the ID. And ⁓ after we get the data we will ingest the user and also ingest the board. So in ingest users ⁓ this is a new added action in the context. So what it does basically

⁓ passing the users from the payload from the back end response into the users ⁓ in the store. And for the ingest board core, we will normalize it first and then we will make sure we only override the existing ones while not cleaning it up, especially when did we search. So if we go to the UI, let's say we want to search for ⁓ implement

As you can see, it's only shows up the implement user list API ⁓ and the user section is still there. ⁓ And if we go to the console log, ⁓ we can see clearly what is happening when we type. So if we say something that doesn't exist, you will notice that ⁓ in here the board context 51, ⁓ which is updated data after we do the normalization and also

After we do all the overrides and you can see in our data store the users are still there. It won't be erased at all. And the cards are still there in the store. What is actually changed is the columns. ⁓ And in columns you can see the card ID is ⁓ empty because there is no cards available after this search. And the actual cards are still there.

Sneha Mehra (00:04:48)  
Unless we are made some changes on the card itself. So for example, if we have a endpoint that modify a card by ID, we change the title or description. That will change the card by ID section. But that will be a different action in the context. So in the top bar now we will ⁓ use the user ID ⁓ to get that from the user state. Because users by ID section in state or in store always

persistent so we don't have to worry about it be removed or something. And we use that ID to get the ⁓ user from the store, ⁓ from this users by ID section. And then when we do a l logout action, we can ⁓ trigger a action in the store or in the context that will clean up the user section and remove this ⁓ corresponding ⁓ user object from that store.

And only at that point the user is removed or disappeared. So we will need to run a fullback like ⁓ login button or something shows up to indicate to the user that you need to log in to see the user after.

—--------------------

Sneha Mehra (00:00:00)  
So before we dive into the main lessons, let's do a quick checkpoint to make sure your environment is set up and everything is working. We will add a tiny feature together, pulling the user avatar from the API into the top bar. As you may have noticed, we have a top bar in our application. There is a gray ⁓ cycle at the top, which is a fake. If we go to the code and the top bar component, we have a div here.

which basically a gray cycle and ⁓ it doesn't actually render the user data reflecting to the current logged in user. Let's not worry about the login ⁓ or ⁓ the current user stuff for a minute. ⁓ And ⁓ so assume we have a new endpoint added, the API user slash ID and the ID is passed from the ⁓ ID parameter dynamically and we will use that ID to do a search based on the

Current users list, ⁓ which is from the users.json, and it has all the current users we like ⁓ mock the database layer from here. And we will return that user ⁓ for that ID. And if we cannot fund it, we just return a full four. And then in the top bar ⁓ component, we can fetch this ⁓ endpoint ⁓ and get the result ⁓ and then render. And here is your task. In top bar.

Instead of using the placeholder image, fetching the user from the API and then display the avatar. Go ahead and try it now. If you get stuck, don't worry, I will walk you through my implementation next. Post the video here.

Alright, ⁓ so if everything goes well you will be able to see a user shows up in here and you will probably have the title set as well for the username. And then if we go back to my implementation, I'm adding a very simple use affect block here and using the internal state to manage the ⁓ user state. And when ⁓ I assume it everything goes well, it can get the response correctly with JSON.

Sneha Mehra (00:02:10)  
format and then I will set the user when it success and I will y render the avatar URL name ⁓ for the alternate and title of the image and then we can see that image shows up correctly. So this is the very rough implementation. We are fetching inside user effect. Well we are not handling arrows or loading status properly yet. That's okay for now. ⁓ the goal of this access is just to make sure your environment is working

And you can connect to the API. In the next module, we will improve this a lot. We will talk about how to better structure ⁓ data fetching, ⁓ managing shared states, and keep things consistent across application. So if you got your avatar showing your all set for the rest of the course, if not, use my implementation ⁓ as a reference and make sure your setup works before moving on.

—--------------  
Sneha Mehra (00:00:00)  
Alright, let's pause here and practice what I have learned. We just worked through lazy loading, the user menu, pop over on the top bar. Now I want you to try the same idea on another part of the application, the board list view. This component isn't especially heavy. The point isn't that it makes our first run slow. The point is ⁓ it's simply not part of the initial experience. When the user first lands the application

They don't see the bird list straight away, so there is no reason to ship that code as part of the initial bundle. So here is what I like to you to do. First, move the bord list view into its own component file so it has a clear boundary. Second, make it a lazy loaded component using React Lazy with a dynamic import. That way React won't download its code until it's actually needed. And third, wrap it around ⁓ the suspend boundary with a simple fallback.

The fallback doesn't need to be fancy, you can just put a loading da da da or something like a skeleton to keep the layout stable while the code loads. And that's it, three steps, very similar to what we just did with the popover. So currently in our board component ⁓ we have this data fetching all the other logic, the controller that can switch the board from list to ⁓ the board view, and that is a conditional rendering.

When support we run the board view, the columns, cards, and the another is a list view. I want you to make changes on this list view, to make that as a dynamic import and the lazy loading module. And here is your checkpoint. You post the video, make those changes in your project and then come back when you are ready.

Sneha Mehra (00:01:51)  
Alright, welcome back. Here is how I implemented the board list view itself is in a simplifier already, so I just moved that to a async folder. In the main application flow, I import it with ReactLeasy, which turns that dynamic import into a component that React will fetch on demand. Then I wrap it with a ⁓ suspense and fallback is loading. I could implement that as a you know skeleton, but we'll keep it that as simple here.

So after this change, if we go to run the build again, and I notice that the new ⁓ bundle corrected is around 2K. It's not a big win here in terms of the ⁓ bundle size, but I did ⁓ make sure we are not loading the unnecessary bytes when we don't need it. And later on when we add in more features in the list view, this size will make more sense. So now if I open up the browser ⁓ and if we switch to the list view.

You will notice that when I click and ⁓ switch, the list view bundle is downloaded at that runtime. Actually, when we ⁓ click that faster, you can see that loading ⁓ text here, and then it will be replaced by the actual component. So that means the list view isn't initially downloaded anymore. It only loads later once the user negative to the board list. So the benefit here is that our initial bundle is smaller.

We're not shipping code that is not needed yet. And also the hydration is a little lighter and faster because there is less JavaScript to parse and ⁓ run up front. And users don't pay for code they never look at. And since the list view is ⁓ its own chunk, it can be cached and reused across different navigations. Again, the key here isn't just about cutting out the big dependencies, it's about

Deferring the code until it's actually relevant. That mindset is what keeps your application faster as it grows.

—----------------  
Sneha Mehra (00:00:00)  
Up until now we have been using React Context to demonstrate normalizations and data modeling. Context is simple, built in and works great for tissue concepts, but in real project we often need something more powerful. Let's take a step back and think about state in React. Some state is local, like the value of an input box or whether a dropdown is opened or closed. That belongs

in use state or reuse reduce hooks ⁓ that we can maintain the state inside a component. Keep it close to the component that is owns ⁓ the state. But some state is shared. For example in our GR like application, the board, the columns and all the cards, multiple components need to access to this particular object. That's when we lift state into a context or into a store, like what we did for

our application for now. So what are our options? First, the React Context API. It's lightweight, built-in, and good for small applications or simple shared data, but it doesn't give you ⁓ the developer tools or performance features for large projects. Second, libraries like Redux or Zustand, ⁓ these are centralized stores. Zustand is smaller, faster to set up and great for modern React applications.

Redux is more opinionated and it has quite a lot of concepts to learn. ⁓ They slice, reducer, action, ⁓ things like that. But it also has a big ecosystem of middlewares and tools. Both are good choices when you need a global, predictable and scalable state. And finally, tools like Redux query or relay, ⁓ these focus on server state, data that come from your backend and they handle the caching, backend updates.

Pagination deduplications ⁓ all of them features all together, they don't replace the normalized client state. ⁓ instantly complemented it. So here is a key idea. No matter what store you choose, the principle stays the same. Normalize your date to avoid duplications and inconsistency. The structure of your state matters more than elaborate you put in. So start simple.

Sneha Mehra (00:02:21)  
Use context to understand the principles. When your application grows, reach for Redux or Susan, or read query depends on your needs. And that sets us up ⁓ perfectly for the next module. We'll dive into the data catching, we talk about caching strategy, pagination, and how to manage requests efficiently. And I will see you in the next module.

—---------------------  
Sneha Mehra (00:00:00)  
The current behavior looks fine at first glance. We have a project board where each card displays its assignee and in a top bar we also show the current user profile. But notice what happens ⁓ when I update Charlie's name in the top bar. So I have add a small ⁓ updating form in a ⁓ top bar, so I can update the user's name.

Before I do that, I want to quickly show you the behavior of the title. And you can see when I hover on the avatar, there is a two tip showing the full name of the user. And also if the user is assigned to a ticket, ⁓ and if I hover, that will be the user's name as well. And ⁓ I will quickly show you the issue here. So when I click the username and I can update it from here, and when I save, it will send the request to the backend.

to persistent data and then I hover ⁓ it will show the new name when it's successful proceeded. And when I move on to the ticket, ⁓ you can see that the name is still the old one. This happens because of the way our data is structured. In the board API response each card abandoned a for user object inside its assignee field. That means charge information is duplicated

Of course multiple places. When we patch one copy, the others remain still. And I can quickly show you what happens in the code level. So as you can see, we have a patch API that accepting the ID and ⁓ the username. And we will fund the user by ⁓ ID that passed from the request. And we get that user from our ⁓ user database ⁓ kind of mock.

And then we find the user and then we change that user's name, basically unwrap the user by index, get that user by ID here, and then we update the payload which is name, and then we directly rewrite the index ⁓ of that user and return the updated user. And if we go to the top bar, we have a handle save name function. At that point we will send a patch request.

Sneha Mehra (00:02:23)  
To the user API and the content will be the name, which is a new name from the input box ⁓ in here. So we get that name and we send that to backend and we wait for the API to resolve. And once it's done, we update the user the local copy. So we will have this ⁓ new updated title. And if we look at the payload from our Chrome, if I do a refresh.

Now it's all reset and I update the name here. So I say Charlie New and I send the request. You can see the payload I'm sending is ⁓ name Charlie New ⁓ and when I get a response it's ⁓ updated the ⁓ the user to the new one. But the problem of this issue is that ⁓ in the board we have ⁓ the nested structure like this.

⁓ We have this ⁓ column. Column has cards. Cards has assignee. Assignee has a four ⁓ user object attached. And so we changed one copy, but we didn't change all the places. The way to solve this problem is through normalization. So instead of duplicating a user object everywhere it's referenced, we replace it ⁓ with a stable identifier.

Like assigning ID, all the user details are stored once in a century dictionary, usually key bar ID. So this approach ensures that both the top bar and the card reference the same source of truth. If Charlie's record is updated, the change propagates automatically everywhere he is referenced. So that means initially we have this nested structure, column card is assignee, it has a full object.

And after the normalization, the data structure will be look like something like this. So we have ⁓ columns by ID dictionary, the key is the column ID, and the object is the ⁓ column object itself. And also we have card by ID object or dictionary and the user by ID ⁓ dictionary as well. So now if we want to find a card in a column, we don't have the full detail of the card, we only have the reference or the ID.

Sneha Mehra (00:04:50)  
And if we want to see the details of a card, we need to use the key to find that in a card ⁓ map or dictionary, and we find this ⁓ four object. And the assignee itself is ⁓ a ⁓ ID as well. If we want to unwrap or ⁓ hydrate the four object, we need to use the ID in a user's object to reference that and get the full object. So in a normalized form.

User data exists in one place. Each card that referenced Charlie points back to the same record. In the next lesson, we will implement this normalized structure using React Context. This will allow the board view, the list view, the top bar all remain consistent without extra duplication or menu updates.

—----------------------------  
Sneha Mehra (00:00:00)  
In this checkpoint, we are going to make the assigning list component real. Right now it's just ⁓ showing the hard-coded gray cycles ⁓ in the next to the board list switch. Let's connect it to the our normalized store so it renders actual users. So you see these three ⁓ green dots here. It's now hard-coded in a assigning list. So here's the current implementation: three static spans, no matter what happens in your data.

This will never change. The goal here is to replace these placeholders with real users. This means reading from the states.users by id, take the first three users and rendering their avatars with proper accessibility attributes. If a user doesn't have an avatar, you can fall back to their initials. You can post a video here and start to implement with a normalized store.

Sneha Mehra (00:00:57)  
Here is the result of the random map. Node F test comes from the store. And the best part, when I update the chart's name in the top bar, this list automatically re-renders ⁓ with the updated value. No extra work needed. This is the pattern we are keep using. The store holds normalized data and components read from it to render denormalized chips. This gives us consistency across the board.

—---------------------  
Sneha Mehra (00:00:00)  
Welcome to the final content module of Front End System Designers. This is where everything we have learned so far comes together. We'll build our KeepStore project step by step in this module. We'll start with a quick cleanup. As the application grows, it's important to keep things organized. So we will organize a code structure, remove unused or demo components, and talk about a few patterns that can help us keep the code base health and scalable.

After that, we will build a modal dialog for editing card details. That's a great chance to practice skeleton loading, handling mutations, and using lazy load to improve performance. Then we will add drag and drop functionality to make the board more interactive and fun to use. And of course, we will make it accessible so everyone can use it, even with a keyboard or screen reader. Next, we will create a board list page and

a dashboard page. That's where we will explore routing and page level lazy loading. That is very useful for applications that are larger than single page. And finally we will wrap up the module with a summary of what we built and what we have learned along the way. But just to be clear, the goal here isn't to launch a production ready application. Well not in setting up the infrastructures or backend systems. This project is about understanding

Front-end system design, the concepts, principles, and patterns that make a front-end application maintainable at scale. So let's dive in and start building our Capstore project.

—---------------------  
Sneha Mehra (00:00:00)  
So let's do some cleanup before we can move on to the next step. ⁓ So currently, if I go to our repo, we have a relatively mixed like structure in ⁓ components. ⁓ it's very flat at the moment. We have this ⁓ components folder and it has everything in it because our current feature is relatively small, but we need to do some cleanup and the structure to make it ⁓ more maintainable. So firstly we will need to clean up the

⁓ some leftover in the card. ⁓ So if you remember last time we have introduced a few demos for the X security concerns. We have XSS ⁓ demo mode, synthesize demo mode as well. we're going to apply synthesize anyways so we don't only apply it in demo mode and also we want to get rid of the XSS because we will evaluate that in the back end API level. So every response in the header has these ⁓

Headers so we don't have to do that in ⁓ demo mode anymore. That means I will clean this up as well. And also if you remember that last time we in the user selector ⁓ we have this throw. ⁓ that means whenever you click a ⁓ avatar, that card will throw. We'll fix that as well. So let's just get rid of this line ⁓ and ⁓

Now it's just working as normal ⁓ and we can do other ⁓ cardboard testing and it works fine, so that's good. And then we need to go to the card and ⁓ kind of revert the mode here. ⁓ So currently we are using this mode ⁓ for the vulnerable ⁓ and ⁓ we will just render the you know normal states ⁓

Let's go to here. This is normal render ⁓ and this ⁓ is runner title. So in run the title, ⁓ we're just using whatever we have ⁓ at the default happy path, which is H3. ⁓ We will just do that in the render ⁓ title. ⁓ And for hydrated we will keep that as is. And for the ⁓ description

Sneha Mehra (00:02:25)  
Let's go find the render description and ⁓ we will ⁓ basically using ⁓ now. So okay, we currently don't have a description. ⁓ that's great. So we don't need this ⁓ anymore. Let me remove that for now. So we don't have the description, and ⁓ then the description function can be safely removed. ⁓ So in ⁓ web storm I can just simply use ⁓ the option enter and it will launch this ⁓

pop-up menu you can remove the unused function ⁓ easily and also for the run title I think we have already got rid of that ⁓ as well. So we just remove that ⁓ and you can see the display title is not used. ⁓ We will remove that one and so all of these three are removable. ⁓ Just get rid of them and in the import we will get rid of these things as well.

So let's do that. And in the future, if you want to ⁓ see how the you know sanitize demo or the XS demo, you can always check out to this branch. Currently we are in the ⁓ main branch and if you log you can see I have added a few tags for all the ⁓ important nodes. So for example, if you want to go to the error boundary, you can always check out to the error boundary. And if you want to go to the playwright, you can always check into that ⁓ point.

Or you can check out the ⁓ specific demo ⁓ to have a look at the synitizing and the excesses. But for now we will just get rid of them ⁓ to make our code a little bit more cleaner ⁓ before we add new features on it. And that means I will clean up the ⁓ synthesis banner ⁓ because we're not going to use that anymore. And also the excess banner, ⁓ we'll just remove them ⁓ and in hooks ⁓

This is used for ⁓ but in utilities we have this demo we can get rid of them and access demo we can get rid of them as well. And in the application we don't need these banners. And also from the file import we'll get rid of them. Then in the board page we ⁓ rendering as normal. ⁓ And if we go back to the page ⁓ to refresh, ⁓ everything should just working. And we can test the query parameter.

Sneha Mehra (00:04:52)  
⁓ like before ⁓ that shouldn't has any impact. That's great. ⁓ we have ⁓ removed these demo pages. ⁓ and ⁓ for other things we will just keep as is and the card component doesn't have any noisy anymore. We can make a commit at the moment just to say okay well just to get rid of these ⁓ demos. ⁓ We'll do the restructure later. ⁓ Okay, so you can see we have ⁓ deleted this bunch of things.

And if you want to check the div, we can do ⁓ git div and ⁓ you know check all the details. ⁓ we have got rid of the utility for the sanitizer and ⁓ XSS demos. ⁓ that's all good. ⁓ So we can add and make a commit, get rid of the demos. ⁓ okay. And now it's reached to a point that we need to do some, you know, ⁓ restructure. If we go to the application.sx ⁓

we have this board page. So I have created a board page in a root level and in it we have this board page. And the board page has two components it's used directly a top bar and a board. ⁓ So I have created two subfolders, one is top bar and other is folder. So in the top bar we have top bar itself, a skeleton and the use profile that will be shown on the top bar. And basically this pop up when you hover and click.

And also on the board ⁓ I have created a board control which is this area, the search input and this ⁓ toggle and also the ⁓ avatar list or assign list. And then I have grouped all the card related components into card subfolder. So for example the user select is used only when you pop up this ⁓ user avatar, so that components ⁓ moved into card.

And also the card ⁓ related test is moved into a card as well. And ⁓ ideally we probably need to create a folder called column. And if we have ⁓ column skeletons, we will move ⁓ the column ⁓ related ⁓ component into that folder. But for now I think we are fine. We don't have to move that at this stage. But if we have more complicated code in the ⁓ future, we can definitely create if a subfolder.

Sneha Mehra (00:07:18)  
under the board and then we move the column into its own folder. And ⁓ the board has a list view and a board view, which is this one. You can you can toggle them. And ⁓ currently the list view is very simple and if we have to make any adjustment or we want to enhance these ⁓ list, like pagination or whatever, ⁓ we can always create a list view on the board.

And then we can enhance that folder until we reach to its limit. But for now, I think the new structure looks pretty good. We can make folder changes based on the new structure. For example, adding the new feature about the editing model ⁓ which will happen in the card. So we will need to click ⁓ a component in the card. And when we click it, it will you know pop up the model dialogue. ⁓ If the model dialogue has skeletons, we'll click subfolders under a card.

And so on. This is pretty much about this lesson and I will see you in the next one.

