2  
Sneha Mehra (00:00:00)  
All right, hello everyone. Welcome to week one. In this week we'll learn about LLM foundations. And learning this topic is quite important because a lot of advanced applications and use cases these days, for example, agents, reasoning models, writing code platforms, productivity tools, and many other many other examples, are all built on top of LLMs. So our entire focus of this week is to understand how LLMs are built, trained and work in practice.

And in particular, we'll start with an overview, then we'll talk about the main two stages of training LLMs, pre-training and post-training. And then for each, we'll discuss how they can be performed and completed. And then finally, we have project one, which is a hands-on exercise to build an LLM playground in Python. So let's grab a coffee and start.

So what is an LLM? An LLM is simply an AI model that can understand and generate text. Now, if this AI model is trained correctly, it can be very useful because we can ask real questions from the LLM and it can answer to us to our questions. And then we can also ask follow up questions and the LLM can respond to our follow up questions. So it's going to engage in a conversation and it's going to be very useful.

Now, there are a lot of different companies that are ⁓ having their own LLMs and they are ⁓ providing a chatbot service powered by those LLMs. ⁓ And one example, popular example is ChatGPT by OpenAI, and I believe that was the first ⁓ chatbot that was made available to public. But following that, there were many other companies having their own LLMs and and offering their own ⁓ chatbot services. For example, we saw Clot from Anthropic. We had

Gemini from Google, Grok from XAI, Meta AI from Meta, and I'm sure there are many other examples and many other companies that are offering chatbot services these days. ⁓ Now, all these services are offering very something very similar. There is a UI where you can enter your text and ask a question. And then the LLM would start answering your question. And there are some additional features that some of these chatbots are offering, and ⁓

Sneha Mehra (00:02:20)  
Chatbots offer more features than others. But just to get an idea, I have some tabs open here ⁓ to go to some of these popular chatbot services and how how and see how they look in practice. So first we'll go to ⁓ Chat GPT. Here is the ChatGPT UI. And as you can see, it's a very simple UI. There is a box here where it says ask anything. And here is basically we can ask our question. ⁓ Here it says ChatGPT 5\.

If we click here, we'll see a bunch of options. It has auto instant thinking pro, and there is this legacy models GPT-4.0. And before the introduction of GPT-5, it was a lot more confusing here. There were ⁓ lots of different options, including non-thinking and thinking models. For example, we had 4.0, we had 01, we had 03, and a bunch of other models. So it's a lot more ⁓ clear now. ⁓ For now, we'll just skip to the ⁓ stick to their default model, which is ⁓ auto.

So it decides which of these to use. But in future weeks, we'll cover other ⁓ options. Like we'll go over, we have one full week covering thinking models, and we also talk about test time compute and pro models. But for now, we'll just keep it the auto one. And here basically we can there are just a bunch of tools we can use. It has ⁓ agent mode, which searches the internet and ⁓ explores lots of different resources and so on. It has deep research, it has a tool to create image.

Connectors and ⁓ many more tools like a study and learn, fib search, and so on. Again, ⁓ in future weeks we are covering many of these. We have a week about agents, we have another week about ⁓ deep research and ⁓ multi- multi-agent systems, and also image generation models. For now, we are not gonna enable any of these tools. And what I want to do is I want to enter a very simple prompt here, so we'll see how it responds. ⁓ where is Paris?

Sneha Mehra (00:04:18)  
So when I type Varis Paris and enter, it immediately starts to ⁓ generating text. And this is the text, this is the response. Paris is the capital city of France. It's located in the north-central part of the country, and so on. Which seems very reasonable. It gives me a very good sense of where is Paris. And here at the end, it also says, Do you want me to explain Paris in terms of ⁓ geography and location? Basically, it's offering more ⁓ to me to keep me engaged. ⁓

So this is ChatGPT ⁓ and ⁓ GPT five. Next, I'm going to switch to Cloud. And my purpose is to enter the exact same prompt and see how Cloud would respond. This is again the cloud UI, and you can see it's very simple. It overlaps a lot with ChatGPT. ⁓ they're just the buttons are sometimes in different places. For example, here is basically where we can choose the which model to use. There are different models, some are more advanced than pro.

And we have to pay f to use those. And some are their default model. So again, we'll just stick to their default model and there's some tools here. ⁓ and then we'll I'm going to write the exact same question here. Who where is Paris? And then enter.

Sneha Mehra (00:05:36)  
So it also started responding and it says Paris is the capital city of France, located in the north central part of the country. And as you can see, this answer is different. It it's different in terms of tone, structure, and style from what we saw in the open air in from ChatGPT. And also it's a little bit more detailed. I can see it's providing coordinates ⁓ like latitude, longitude, and so on. Now let's go to Gemini and ask the exact same question here.

So again, here we can choose which model we want to use and there's a bunch of tools and ⁓ like deep research to ⁓ get in depth answers and so on. But we're only going to enter the same prompt, various parties.

Sneha Mehra (00:06:21)  
Paris is the capital and most populous city of France. Again, this answer is not exactly as what we saw earlier by other two models. But this answer is also correct and it seems accurate. It also providing coordinates. And same story with other chatbots like Grok and MetaI. I'm not going to enter the same prompt there, just to save more time. But ⁓ it's going to be very similar. Grok, you can choose the you see the tools are here, and you can choose the models here. ⁓

Similarly for MetaI, ⁓ we can use certain ⁓ tools and ⁓ ask Meta AI any question we have. Now, all these models, what they have in common is they were able to respond to my simple prompt. However, they responded differently. And some of them are in more detail and some are more have different structure and tone. So this is how LLMs are different. now if I start asking more complex questions.

Some of these models may no longer be able to answer my question correctly. And this is why certain LLMs are typically more powerful than others. ⁓ And it keeps changing because all these companies keep training new LLMs and more powerful LLMs. So the the quality of their LLMs keeps changing and improving over time. And again, later we would see how the ranking works and how we can compare various LLMs. But for now,

I just wanted to show various LLMs and how they work in practice. Now, the question is how these LLMs are built. ⁓ And this is the next topic. So LLMs are built typically in two main stages. The stage one is pre-training and stage two is post-training. And we're going to talk about both in very detail in in future lectures. But for now, it's at a very high level.

Pre-training is basically the first stage of training one LLM. And then in this stage, we train a model using some training algorithm on the internet data. So this stage is very expensive. It requires lots of compute, ⁓ thousands of GPUs. ⁓ GPUs are these hardware, ⁓ basically compute powers that ⁓ allow us to train LLMs, and they can be very expensive. ⁓

Sneha Mehra (00:08:45)  
It's also a very lengthy process. It requires a month of training to train an LLM. And after pre-training a stage is complete, the outcome of the pre-training is a base model, which has a very good understanding of the internet data. And in other words, it has implicit knowledge of the world. Because assuming that there are lots of ⁓ text on the internet about all these different domains and different ⁓ areas, so

And since the model is exposed to all those data, it has a very good understanding of the world and ⁓ very good implicit knowledge. Now, the second stage, which is post-training, is basically we continue training this base model on a different data. We call it post-training data, until we get the final model. And this stage is a lot less expensive. It requires less GPUs, typically hundreds of GPUs or even less. And then it requires

Days of training only. And in terms of cost, this stage is less expensive than pre-training. Pre-training typically costs hundreds of millions of dollars, but post-training is a lot cheaper. So only well-funded startups and very big companies can ⁓ perform and ⁓ do pre-training just because it costs a lot. It costs millions of dollars to complete pre-training.

There is this interesting report from Stanford which ⁓ provides lots of statistics of ⁓ which companies or which universities have their own LLMs ⁓ and how much it costs. And I have this open here.

There are lots of different statistics. It's it's an interesting read, ⁓ like how many models are open and there is no access to the public and so on. And then the cost of some of these models and which ⁓ the organizations that build these models, you can see all these companies are very big ⁓ companies that can afford to train LLMs. And then ⁓ here it shows where like most of the LLMs are coming from industry.

Sneha Mehra (00:10:54)  
And this is also the the figure that I wanted to show. Here we have various LLMs and ⁓ language models. Here we can see the ⁓ number of parameters, but what I want to focus on is this side, which is the cost, the training cost. Basically, how much it costs to complete training ⁓ of an LLM. So some of these models might be familiar to you. For example, here we have ⁓ GPT three model.

with 150 f 175 billion parameters. And this model was trained by OpenAI. And you can see it the cost of training this model was less than a little bit less than 10 million dollars, but it's still, you know, ⁓ very expensive to train such model. And you can see more models here and then GPT four and Gemini Ultra. And if you see these models, they're ⁓ the cost of training those models are around hundreds of millions of dollars, which is very expensive.

and there are ⁓ some other interesting statistics, but just wanted to focus on the how costly it can be to train these LLMs. So back to our lecture. After we complete these two stages of training, pre-training and post-training, this final model is what companies typically use and deploy to ⁓ to power the chatbot service. So for example, if you ask any final ⁓ any question from this final model, it would answer.

So here is an example. I asked Chat GPT. This is a real example. I asked ChatGPT ⁓ with a prompt like, tell me a short joke. And then the output was why was the math book sad? Too many problems. So ⁓ just wanted to show this example as a as a way that ⁓ this final model is what most companies use to ⁓ deploy and use it as a way to answer users' questions. Now

This is all I wanted to talk about in this lecture. In the next lecture, we'll talk more in detail about pre training and understand different steps of pre training.

Sneha Mehra (00:13:04)  
Hey, in the last lecture, we saw various chatbots and LLMs. And then we discussed the two main stages of training LLMs. Stage one is pre-training and stage two is post-training. In this lecture, we are going to focus on the pre-training stage. And in particular, we are going to talk about the internet data. What does it mean to train something on internet data? How should we collect this data and how to prepare this data? So let's start.

So data preparation has three main steps. ⁓ Step one is crawling the internet. So crawling is simply a software that starts from some ⁓ seed URLs or just a single base URL. And then it tries to ⁓ extract its content and identify the outgoing links within that URL, and then it goes to those outgoing links. And there is a loop.

Where it keeps doing this, it keeps downloading the content, identifying the outgoing links of various URLs, and then visiting those URLs. It keeps doing this until it visits majority of the internet. So just to better understand how web crawling works, ⁓ it's ⁓ I have asked ChatGPT to write a simple web crawler class in Python. And I also inspected the its code. It's c it seems correct. ⁓ And before talking about the code,

The purpose here is not to have a production ready web crawler software. The purpose is just to understand the logics and understand what are the key pieces in a web crawler. Because once we learn that, most of the ⁓ advanced web crawlers are all built based on the same logic. So again, any web crawler is just a software. And here we have a simple class, ⁓ simple web crawler. ⁓ And what it does is it starts from a base URL and this is given.

As the input to this class. And then it also keeps a set, ⁓ a Python set, named time named visited. ⁓ And this visited is basically keeps track of ⁓ URLs that this web crawler already discovered, extracted their content, and it's done. ⁓ And then we have this method called crawl. And this is the main crawl, the main function of the main functionality of this class. What it does is it starts from this to visit.

Sneha Mehra (00:15:31)  
Which is basically a Python list and it adds the base URL as the first URL to start exploring it. But then this this list is basically a way to allow the software to keep adding new links as it discovers them so it can later explore them. And again, this whole thing is based on a for loop or while loop. And then in the while loop here we have as long as there are links, unexplored links in the to visit, meaning that there are URLs that we have not ⁓

We have not explored this content yet. ⁓ We extract the URL. If it's previously seen or explored, we skip it. Otherwise, we start the crawling. And here there are some libraries like requests. Again, we don't want the focus, is not to learn the details of these libraries. There are just lots of different libraries we can send requests to different URLs and get the output. And here we are using this request.get, send the URL, get a response back.

And then from this response, there is this text, which is the ⁓ HTML content of that URL. Now, again, beautiful soup is just one library and there are other libraries, but all these libraries are allowing us to, once we have the responses, the HTML content, it allows us here to just walk ⁓ to go through all the URL links within that HTML content. So here basically it says find all ⁓ these tags.

And then those are links. And then if it builds the full link, full URL, because some of these might be like relative URLs. And then once we have that, it just appends them to the to visit. So basically to visit is just a list that it keeps track of all the discovered URLs that we have not still explored its content, HTML content. And then there is a while loop that it continuously tries to ⁓ explore unseen ⁓ URLs in the to visit. And once we are done.

It just it just gets added to the visited. So very simple. So this is the URL. Again, in practice, there are more details or more arguments, like ⁓ maximum number of times to ⁓ to go deep and keep identifying links and going to those links. But those are details. The purpose is just to for now understand the high level logic of building a web crawler.

Sneha Mehra (00:17:58)  
Now, back to the lecture. There are two ways now we can crawl the internet. The first option is to crawl ourselves. And this is what most big companies like OpenAir Anthropic does. For example, just to validate that Anthropic does crawl, there is this question on and it's on the Anthropic website that whether Anthropic crawl data from the VIP. And the short answer, if you read this, it says yes, we do crawl the VIP.

And this is also a GPT2 paper published by OpenAI. And we are going to talk ⁓ in detail about different parts of this GPT2 model later in this week. But for now, I just want to scroll down and get to the training data here, training data set. And if you read this, here they are saying that instead we created a new web script which emphasizes document quality. So basically, they are also.

having their own web crawler. So they have this software, they've written that and then they're using that to crawl the web.

So this is the first option. The second option is to use a public repository of crawl pages. Now there are different research teams and ⁓ organizations that regularly crawl the internet and provide the extracted content public and make it publicly available. And one very common example is common crawl. So I

Google Common Crawl here and it says it Common Crawl is a nonprofit organization that crawls the web and freely provides its archives and data sets to the public. ⁓ And there is just this website for Common Crawl. If we go there, ⁓ it's it's how the website looks like. We can go to the data. There is this latest crawl ⁓ and you know, instructions to get us started. And here we can choose a crawl. And it seems that the crawls are hosted on S3.

Sneha Mehra (00:19:54)  
So here we can just choose crawl. Again, there are various options because they regularly crawl and there are different ⁓ times. So it's 24 crawls, 23, 22, and it goes all the way back to I think when since they started 2008\. So any of these, once we select, it would just give us some instructions and more details about the crawl. Now back to the lecture. Most companies prefer to write their own ⁓ crawling if they are big.

So it gives them more flexibility. And for fast experiments or startups or smaller companies, they prefer to start from publicly available crawl pages just because it it allows them to iterate faster. And these are some interesting statistics from Common Crawl. they're crawling the web since 2007, and then each time they crawl, it's approximately two point seven billion web pages. And as a result of that, they're around ⁓

It's around 200 to 400 terabytes of HTML text content each time that they crawl the internet. And then they release a new crawl every month or ⁓ two months. All right, so this is the step one of data preparation, crawling the internet. Or in other words, it's collecting data. Now, after we collect this data, it's the out the output of that is ⁓ HTML text content. Now, the second step, which is also very important, is data cleaning.

The raw HTML content has lots of issues. For example, here I have a screenshot. If I zoom in to see how this raw data looks like, you can see the HTML text content has lots of ⁓ irrelevant ⁓ information. For example, there are lots of tags, HTML ⁓ tags and markdowns and attributes, and ⁓ many of these are not really useful to to train a model. ⁓ we don't at least

For most use cases, we don't want an LLM to learn about all these tags and head ⁓ HTML, all these things. What we want is the LLM to learn from the actual content that we see from URLs, be the content that we see in a browser. And those are often ⁓ within certain tags like H1 tags, and here we have the ⁓ P tags. Like this text, for example, the new MacBook Pro M4 features the latest Apple Silicon Sub.

Sneha Mehra (00:22:19)  
These are some of the text data that we really want to extract from this ⁓ HTML content and use them for LLM training. So that is the ⁓ main goal of data cleaning. It's to extract useful information from the raw HTML content. But that's not the only ⁓ the step to clean the data. There are also other issues in the raw HTML content. For example, a lot of text is ⁓

Duplicated across the internet. You can think of when there is an important news, there are lots of various websites that publish the exact same news or with a slight rewarding of the ⁓ same news or original ar article. And we typically don't want to have lots of duplicated data in our training data, because if we expose the LLM to ⁓ duplicated data and the LLM see them over and over again.

At some point it may memorize them. So we don't really want the LLM to memorize ⁓ news. What we want is to learn about different words, different ⁓ topics and knowledges. So ⁓ these are some of the things that are important to take care of during the data cleaning. Again, there are other things. For example, one important thing is to make sure that we use only ⁓

useful and safe content. There are lots of websites that we may not want our LLM to learn from. And as part of data cleaning, it's important to just get rid of those websites and those text content. Now there are different companies and researchers that they've tried various ways of cleaning ⁓ the raw HTML content or common crawl. And then after they apply those cleaning filters or steps, then they get the clean text and then they name it something.

For example, here I have ⁓ some popular data sets. For example, we have C4, Dolma, RefineWeb, FindWeb. So what these are are basically different companies or or or researchers. We would see them in a sh in a bit. ⁓ they clean the raw data, and then the outcome is just a data set, and that they name it something. For example, they name it C4. For example, if we search C4, we would see under ⁓ TensorFlow.org. ⁓

Sneha Mehra (00:24:40)  
We would see the description of the C4 dataset and it's it's ⁓ created by Google. So the description is this is basically a clean version of Common Crawl's web crawl. So and then there are more details to it, and you can you can just read them through and see ⁓ you know how big the data is, for example, how many examples are there in the training split and so on. But what they are is basically they are the clean version of OpenCraw, Common Crawl.

And then this is an example of C4. It's published under ⁓ it's AI2. ⁓ And what what this is is basically how this data looks like. This is not exactly the original C4 that Google published. It's a cleaned version of C4 again. So C4 cleaned common crawl. And this is here you can see it's a clean version of Common Crawls. ⁓ and then this is the processed version of Google's C4 dataset. But again,

The whole point is that, and you can see the sizes here, like ⁓ 305 gigabytes ⁓ of English data. And ⁓ non-clean is 2.3 terabytes and so on. But basically the data is it's very similar. C4 and the here and also the original C4. After cleaning, it's going to be looking like this. It's a table and we have some text here. One column is text, and then one column is the URL that this text was extracted from.

And you can see here, for instance, this URL had this text. And then there are also more ⁓ examples, different ULS, different ⁓ URLs, different links. ⁓ And some of these are in different languages. ⁓ And ⁓ so this is C4 after cleaning comment crawl. Now

This dataset is relatively old, but it was very important in earlier days of LLMs because a lot of LLMs were using this dataset to to start from ⁓ to as a starting point to to pre-train their LLM. But then there were also other more recent examples and data pipelines for cleaning common crawl. One recent example is Dolma. I have the paper here, ⁓ Dolma.

Sneha Mehra (00:27:04)  
And it says it's an open corpus of three trillion tokens for language model pre-training research. ⁓ And ⁓ we'll talk more about tokens in in future lectures. For now, we can just think of tokens as words or sub words. So what that means is that this after cleaning the ⁓ web crawl, this Dolma data set has around three trillion words, English words, let's say, or sub words. So this is the scale of the clean data.

And then you can also take a look at the paper. It's very interesting ⁓ statistics here. ⁓ For example, they're not only starting from Common Cross, but they're using other sources because all these sources can ultimately be useful for the LLM to learn from. For example, GitHub has lots of code, ⁓ Reddit and some other places, Wikipedia. And then these are some ⁓ numbers and you know how big the data is and so on. And then you can just read ⁓ their pipeline and how they clean this data set.

and what what specific filtering steps they apply to clean the web. You can see they have like quality filtering, content filtering, and so on. ⁓ And there is also another very popular data set. It's called Refine Web. ⁓ And again, the same story. You can you can see the stats here and how big these data are. They are

It's basically just a sequence of filtering steps and cleaning steps to clean ⁓ raw internet data or common cross.

Sneha Mehra (00:28:42)  
Now back to our ⁓ lecture. So we saw these three examples, C4, Dolma, RefineWeb. So the last one that I want to share is FindWeb. And Find Web is more recent and it's openly available. It's published by Hugging Face and it's basically just ⁓ a data cleaning pipeline. With here you can see all the details of the pipeline. ⁓ And I have this ⁓ screenshot in the lecture as well.

⁓ It's ⁓ these are the key steps of ⁓ cleaning common crawl. And again, now most of this should be ⁓ familiar. For example, the first step is URL filtering. In this step, they filter URLs that they are not interested, the LLM to learn from those content. And there is typically a list of block list content. Here you can see what FineWeb ⁓ used to.

lo to filter the URLs and these are some of the categories. For instance, they ⁓ they filter URLs that ⁓ are belonging to adult category and there are some other categories here. And these are pretty ⁓ it's subjective. Different companies may want to include or exclude certain categories. But again, this is the first ⁓ step of cleaning the data. So they first

Just get rid of all the URLs that they are not interested to learn from. And then after that, they extract the actual text from the raw HTML content. This is what we saw earlier, that the what we want is the actual text content within certain tags. And then after that, we have language filtering. Here we ⁓ filter text data from certain languages that we don't want the LLM to necessarily support.

And then there are a bunch of other filtering and deduplication here. It tries to deduplicate ⁓ very similar content. And then more filters here. And then finally PII removal here, they just ⁓ remove content with ⁓ sensitive information like bank account, bank numbers, and phone numbers and those kind of things. And then after they apply all these ⁓ sequence of filtering steps to the ⁓ raw data, they end up with ⁓ the clean data.

Sneha Mehra (00:30:59)  
Which has ⁓ 44 terabytes of disk space. This is the storage needed to just store this final clean data. And then approximately it ends up at 15 trillion tokens. Again, we'll talk about tokens later, but for now we can just think of those as words or sub words. So it's around 15 trillion ⁓ words. And this is this is basically what most ⁓ what data cleaning data cleaning step is.

We apply a certain ⁓ sequence of we apply a certain sequence of filters and cleaning steps to go from the raw HTML content to a clean text. And this clean text is basically just a very huge text file. Or it can be multiple text files. But again, it's just huge text file that just we have the this entire text data from internet. Just to get an idea, this is the ⁓ fine web data after the cleaning. So this is how it looks like.

Again, ⁓ it's very similar to what we saw earlier. There is this huge table with text column and URL column. You can see from this URL, this text came. And this is the data, this is the final clean data that we want to use to train our LLM. And this is just text from internet. It can be it's very random. Like, did you know ⁓ you have two little yellow nine volts and so on? And if you see another example, five reasons I love Boston.

So basically it's like ⁓ a ⁓ very long text from a huge range of ⁓ topics and areas and domains that we end up having in this clean text.

Back to our lecture, we've covered data cleaning and we saw ⁓ popular pipelines to clean data.

Sneha Mehra (00:32:49)  
The next step ⁓ of data preparation is tokenization. ⁓ the purpose of tokenization is to go from the ⁓ raw, clean data to a long sequence of discrete numbers. And this step is also quite important because machine learning models do not expect ⁓ textual inputs. What they expect is numerical inputs. So in this step, we follow a certain procedure or algorithm.

named tokenization to convert text to sequence of numbers. And that sequence of numbers is now useful for the machine learning model or the LLM to learn from. So in this lecture we covered the key three steps of data preparation. We talked about crawling the internet and we saw that how we can start from some publicly available ⁓ crawl pages like Common Crawl.

Then we saw data cleaning and various pipelines ⁓ and ⁓ how this raw HTML text content can be cleaned using a sequence of filters and ⁓ algorithms. And then we also saw certain data sets that follow these recipe and ⁓ pipeline to get the clean data. So we don't have to

start from we don't have to necessarily write our own cleaning pipeline. We can always start from an open source clean data such as FineWeb and we start from there. And then finally we would tokenize the data to go from the raw text to a long sequence of discrete numbers. So this is the high level of data preparation. In the next lecture we would talk more about tokenization and we understand how it works in detail.

All right. In the last lecture, we learned about the three main steps of data preparation and we saw that the last step, which is also very important, is tokenization. In this step, we convert the raw text into a long sequence of discrete numbers. In this lecture, we are going to learn more about it and we understand how it works under the hood. How can we really build something that can convert ⁓ raw text into numbers? So let's start.

Sneha Mehra (00:35:12)  
So in tokenizations, we are we dealing with two phases. The first phase is the training phase, and then the second phase, which we have it here, is the inference phase. In the training phase, the idea is to expose the tokenizer to this long sequence of text that we've already cleaned. And then we let it learn from it and understand the statistics of different words and what are the unique words.

The purpose of the training phase is to ⁓ apply these two ⁓ squares that we have here. These are just two ⁓ algorithms or s or or logics. And then as a result of that, the tokenizer would end up with a vocabulary. So let's just talk about each of these in detail. The first step the first part is text splitting. So what happens here is that once we this long sequence of text is provided.

This applies some logic. It's it's simply a logic to apply to this long sequence of text to split it. Basically, we the goal here is to split the text into smaller units. And depending on what on our logic, it can be ⁓ implemented in different ways. For example, here we have an example. ⁓ machine learning ML is a subfield of artificial intelligence. And then after applying text splitting.

We may end up with something like this. It's a list ⁓ of smaller chunks or units of text. For example, we have machine learning, and then we have like smaller units and subbuild, artificial, and and finally we have dot here. So this is text as ⁓ splitting. Now imagine we apply this logic and the entire cleaned internet data to get a very huge list of ⁓

smaller units of text.

Sneha Mehra (00:37:14)  
The next part is building the vocabulary. Now, given this huge list, in this step, we create a vocabulary of all the unique units of text that we ⁓ have in this list. And then we also assign an ID. So it's it's very simple. Building vocabulary is nothing but finding all the unique ⁓ units, and we refer to these ⁓ each of these ⁓

smaller units of text as tokens. So the building vocabulary is simply responsible for f finding all the unique tokens and just listing them in a table. And also assigning an ID to them. For example, we can have a vocabulary like this. And then for tokens we have A, about, after, all, also, and so on. And these are some of the tokens that we've seen in this list after we split the the initial the original text.

And then these are the IDs, which starts from zero and we just increment it ⁓ one by one. And ⁓ here it can be seen that at the end we have around 270, ⁓ 131 ⁓ tokens. And depending on our training data, this long sequence of text, we may end up with different vocabularies with different lengths. It just depends on how many unique ⁓ tokens we would end up seeing after we split the text.

So this is the training phase. This is we just run this one time on our training data to build this vocabulary. And then from there, the tokenizer is ready to be used. Basically, the tokenizer, it's just ⁓ it internally stores this vocabulary. And then anytime we pass a text to it, it can convert it to a sequence of numbers. And the way that it does ⁓ is

First it runs this ⁓ text splitting logic, it applies it to tell me a joke, and then it splits tell me a joke to ⁓ the sequence of tokens. And then it would refer to its internal vocabulary and replace each token with its corresponding ID. So for example, tell me a joke might become tell me a joke after a text splitting. And then after tokens are replaced by their IDs, we may end up with something like this.

Sneha Mehra (00:39:39)  
So, this is how a tokenizer can go from any text to a sequence of numbers. In addition to that, it can also go back to an original text given a sequence of numbers. And this is also a very simple logic. So when we pass a sequence of numbers like this to the tokenizer, internally the tokenizer can also store an inverse table. So instead of mapping from tokens to IDs, it can have a mapping from IDs to tokens.

So once this is provided to the tokenizer and we want to ⁓ detokenize it, meaning that we want to go from numbers to text, it just simply goes to those ⁓ rows in the in the inverse table and find the corresponding token. And then it it applies the inverse of text splitting to the tokens to output the original text. In this case, tell me a joke.

So these are the two main phases of ⁓ tokenization.

And training phase is basically there are different algorithms for the training phase. And all these algorithms mo mostly differ in their text splitting logic because building vocabulary is is very obvious. It just needs to find all the unique tokens. Next, we are going to talk about the important categories of ⁓ tokenizers, which are mostly referring to the algorithm that we can apply here and ⁓ create the vocabulary.

So all these different algorithms ⁓ can be categorized in ⁓ into three big categories. We have word level tokenizers, we have character level tokenizers, and we have sub-word level tokenizers. And again, for each of these, there might be various algorithms within each category. Let's just briefly talk about each. Before I explain the in detail,

Sneha Mehra (00:41:38)  
⁓ I just want to point out that word level tokenizer and character level tokenizers are no longer used in building advanced LLMs, at least to the best of my knowledge. And most of recent and advanced LLMs are all relying on a certain algorithm within the sub-word level tokenization. And the reason for that is because word level and character level has some limitations, which we will talk in a bit. So let's just start with.

word level tokenizer. It's it's very simple. As the name indicates, it splits the text based on its ⁓ let's say ⁓ white spaces to get the actual words. So the text splitting logic ⁓ is for instance, split the text based on its white spaces. So a sentence or an input like perfectly fine when it goes to word level tokenization. And

WorldWeb tokenization applies the Texas splitting logic, it would end up with perfectly and fine. And then ⁓ using its internal vocabulary, it would just replace these tokens with their corresponding ID.

So these algorithms, when we train on our internet data, we would internally it would apply this ⁓ text splitting and then it would build a vocabulary. And since it's word level, the vocabulary can be very large. And the reason for that is because there are lots of unique words on the available on the internet, both on English and also in other languages, if we include them in our training data. So ⁓

You can see here we have perfectly, and then its ID is 52,000 something. And in practice, it can be a ⁓ hundreds of thousands of unique words if we run word devil tokenizer on internet data. So this is WordDel tokenizer. Now let's switch to character level tokenizer. Again, as the name suggests, it's it splits the text based on the characters. So for example, an input like perfectly fine.

Sneha Mehra (00:43:46)  
When it goes into character level tokenizer, it becomes something like this: P-E-R-F-E-C-T and so on. And then each token gets replaced by their ID. And as you can see, the IDs here are very small. And this is because in character level tokenizer, when we apply it, when we when we train it on the internet data, we don't usually have ⁓ that many characters. ⁓ for example, if it's trained on English only data.

We would end up perhaps with lowercase, uppercase ⁓ characters and some punctuations and you know things like dots and question marks and so on. So the vocabulary length would not be the vocabulary table is not going to be huge. It's in fact going to be very small, and we would have ⁓ very little number of ⁓ tokens. And that's why you see most of these numbers are ⁓ small here. Now

This is character level. Before I talk about sub word level tokenization, I want to ⁓ talk about the limitations. As we ⁓ saw earlier, when we run word level tokenizer and train it, we would end up very with a huge vocabulary. And that is the limitation of word level tokenizers. When the vocabulary is huge, it means that we have to maintain and manage lots of different tokens. And this can be very expensive when we train a model.

Because during the training, we have to ⁓ keep this table and learn some association ⁓ with each token or or in fact its IDs. So it's going to be very expensive to learn when our ⁓ vocabulary is very large. Now, character level tokenizer, its limitation is exactly opposite of that. ⁓ It does not really suffer from having a huge vocabulary because the vocabulary is small.

But what it suffers from is the long ⁓ sequence of numbers. So, and the reason for that is because of its ⁓ splitting logic. You can see here it would end up after we convert perfectly fine to a sequence of numbers, we would have too many numbers here. And this is not also ideal, and this is also can be expensive because now the during the training, the model needs to learn ⁓ the dependency and the ⁓

Sneha Mehra (00:46:11)  
relation between all these tokens. And if this sequence is too long, it's going to be costly.

So these are the limitations, and this is why in the first place subward level tokenizers were invented and preferred. So now let's switch to sub word level tokenizers. So ⁓ the key intuition behind subward level tokenization is that they come up with some splitting logic, ⁓ and after we apply that splitting logic to the text, it splits the text into smaller units, such that each unit is larger than characters.

And also they are smaller than words. So this way it tries to somehow balance ⁓ and sit somewhere between word level tokenizers and character level tokenizers. For example, for an input like perfectly fine, after it goes to the subword level tokenizers, it can become something like perfect, lee, and fine. And then each token would get replaced by their IDs. So as you can see, the ⁓ the word the tokens are.

smaller than word level tokenizers. For example, we don't have perfectly as a token. We have perfect and lee. So the word perfectly would be split into smaller units. Now there are different algorithms that are under sub word level tokenizers. And one very popular algorithm is byte pair encoding or BPE. ⁓ And that is what a lot of LLMs have been using to as a tokenizer.

And I'm going to briefly go over it so we get an intuition and high-level understanding of how an algorithm like BPE works. ⁓ The key idea behind the algorithm is that it starts with characters and it tries to iteratively merge frequent pairs of tokens and create new tokens. So this way it has full control over the vocabulary size. Let's see a quick example. ⁓ I I searched online and I found this.

Sneha Mehra (00:48:17)  
Visualization. This seems to be in Korean. I don't understand Korean, but what I want to show is this diagram. And basically the way the algorithm tries to form tokens and create its vocabulary is it starts from the training data, it counts all the unique words, and then it gets to the characters. And then for each character, it creates a new token. So for example, H has a new token, U has a new token, and then it stores those in the vocabulary.

And then from here, and this 10 is the ⁓ frequency of ⁓ basically number of times that this word was appeared in the training data. And then from here, it tries to iteratively merge these tokens, and at this level it's basically characters. It tries to iteratively merge them and then create new tokens. ⁓ And the way that it ⁓ it merges two tokens or picks tokens to merge is based on their frequency. So it tries to

Find the most frequent pair and then just merge them. For example, after here we have ⁓ it creates a new token named UG. And that's simply because there are a lot of UGs in the in the training data. If if you see here, we have UG ten times, we have also UG here five times, and we have UG here five times. And because of that, now it creates a new token named UG because it believes that Yu G can be a popular ⁓ token in in the training data.

And it can it might have some meanings associated to it, and then it creates UG. And then it continues doing that until it reaches a certain number of tokens in the vocabulary. So it has control over the vocabulary size. Sometimes they keep doing this until the vocabulary reaches, for instance, ⁓ 50,000 ⁓ tokens. So this is bytepair encoding, and there is a very interesting tutorial ⁓

Published by Hogging Face, by pairing coding tokenization. And then here it explains in detail the training algorithm and how it's starts forming the tokens and merging and so on. So it can be an interesting read. And then ⁓ as I mentioned, to the best of my knowledge, all of advanced LLMs these days are based on ⁓ some variations of suborder level tokenizers, and a lot of them are based on BPE algorithm. For example, in ⁓

Sneha Mehra (00:50:41)  
This is GPT2 paper again. And then ⁓ again, we are going to talk about different parts of it like model architecture and so on. But for now, I want to show you the tokenizer that they've been using. So here ⁓ is the input representation. And here they're talking about bytepair encoding or BPE. And then they're just explaining it that how they are using BPE and ⁓ in fact some variations of it as their tokenizer. ⁓ And

I also searched Lama3, which is Meta's LLM, and it's relatively new. ⁓ And even this paper was using BPE as its tokenizer. ⁓ here they're saying that the tokenizer is a BPE model. So again, most of the recent LLMs are based on ⁓ some variations of BPE or some algorithm under subworld level tokenizers.

Now ⁓ after just to get a comparison of word level tokenizer, character level tokenizers and sub word level tokenizers, this can be one possible vocabulary after we apply each of these. For example, if we try if we use some character level tokenizer, this can be the vocabulary. ⁓ And what I want to show here is that the vocabulary is very small in terms of ⁓ number of tokens. It has this is an example, but

In this case, we have ⁓ 105 tokens. And then you can see the tokens are like various, like the smallest units of text, which is which are characters like A, B, lowercase, uppercase, punctuations and the space, and so on. Now, if we apply word level tokenizers to our training data, we may end up with with a vocabulary like this. And as you can see, here we have ⁓ a lot more vocabularies. It's ⁓

270,000 something. Again, this is an example, but they are closer to reality. And then you can see the tokens are any individual word that was ever seen on the internet. ⁓ And the reason that it's very, very ⁓ it can this vocabulary can be very, very large is because there are a lot of ⁓ on the internet, basically it's very likely to have ⁓ lots of random words. I mean, for example, if I search Google something like A. ⁓

Sneha Mehra (00:53:04)  
⁓ okay, apparently we have some ⁓ musics here with with this ⁓ name. And you can see this name. So since internet data means all this text data, and then if we run if we split text by white spaces, we would end up with all these ⁓ smaller units. And then at some point we would have this ⁓ this ⁓ this sequence of characters as a word. So that's why in practice we would have a vocabulary which can be very, very large.

And also ⁓ not too efficient because ⁓ maybe ⁓ appeared in some few places, but it's not a meaningful word. So it just doesn't it's not too efficient to have its own token. ⁓ and then we have sub word level vocabulary, which is basically sits in between word level and character level. And you can see tokens can be for

important frequent words in English. We have them as their own token like the of home and so on. And then it tries to also ⁓ identify the ⁓ smaller units that are also very frequent, for instance, in English. So that's why we have ing, ed, abel, and so on as their own token. So that for instance when we have a verb like walking, it can be split into walk and ⁓ ing. So

It's going to split basically text into smaller units. So these are just some examples. ⁓ And here we have around 50,000. But in practice, in more recent ⁓ advanced tokenizers, subboard level tokenizers, this vocabulary can be bigger, like 100,000 tokens or even 200,000 tokens. But again, word-level tokenizers, especially if you include ⁓ text from other languages, can become very, very, very large. ⁓

That's that. And then the next thing I want to show is an interesting ⁓ visualization of tokenizers. ⁓ And ⁓ there's this website, ⁓ tick tokenizer. And then I have this open here. So this is just a visualization to get a better understanding of how the text looks like after we apply tokenization and what are the IDs and so on. So here we can choose some ⁓ popular tokenizers.

Sneha Mehra (00:55:26)  
So these are basically pre-trained tokenizers on some internet data and it already has its internal vocabulary. So it's ready to be used for inference. And then these are the names of those ⁓ tokenizers. So let me choose this one for instance. And here I can write something like I love machine learning.

And ⁓ we can see that this has five ⁓ tokens. And then after applying the token, after applying the splitting logic, it explicitly stakes into I ⁓ space love, space machine, space learning, ⁓ and ⁓ this punctuation. So ⁓ basically this is how the

The algorithm, the BP algorithm, came up with tokens. So the reason that we have a space love is because just because the after applying the algorithm to the internet data, this becomes its own token. And then these are the IDs. So ⁓ in the in its vocabulary, I is 40, love is ⁓ 3021 and so on. And exclamation point is zero. So it's an interesting ⁓ visualization. We can just play around with it, we can remove spaces and see if

it's a still a word or not. Or we can do things like, ⁓ I love ⁓ walking.

And walking is its own token in this example. ⁓ it's perfectly fine. Let's try this. It's perfectly fine. ⁓ So here we can see perfectly is becoming three tokens. And LI, as we had in our example, has its own token. So what that means is ⁓ I think I have a misspelling here. Perfectly. All right. Now perfectly is ⁓ its own token. Now what is interesting, if we try something very random like

Sneha Mehra (00:57:21)  
It is ⁓ and something like this. You can see the splitting logic splits it into various smaller units. And just that's just simply because we don't have any tokens associated with these words. So for example, ⁓ is becoming a space and three R, three A. And then it's because it's a meaningful word, it has its own token. And there are different tokenizers. We can try different. This one is more ⁓ more recent version published by trend and published by OpenAI.

⁓ 200k base. And then if we apply this, we would just probably see different numbers, and just because its vocabulary is different. All right. And then there are a bunch of other models and open source models. So it's just an interesting thing. And there are different also encodings and ⁓ you know, we can also pick models, ⁓ like actual models to see how ⁓ an input prompt gets translated into a sequence of IDs. So this is the

The visualization of TikTokenizer. Now, the next thing I want to talk about is that we saw all these different algorithms, BPE and how it works. But in practice, we we usually do not ⁓ have to implement them from scratch. There are some great libraries that is already available and open source and we can use them. And there are also pre-trained tokenizers available. So we don't have to even train tokenizers from scratch if if ⁓ it we don't have to. ⁓

So ⁓ and one very popular library is TikToken, published by OpenAI. ⁓ And this is TikToken, and ⁓ this is the repo. And again, you can see the the code and everything is here available. ⁓ And ⁓ it's also very simple to use. we can just specify t basically the description is TikToken is a fast BPE tokenizer. And they have some ⁓ numbers here that why it's faster than other alternatives. So it says that it's

Between three to six X faster than other open source tokenizers. ⁓ So it's a very good ⁓ library to start from and you know ⁓ use it for our tokenization purposes. ⁓ And also it's very easy to use. Just to show an example, I have my Jupyter notebook open here. And then I want to show how to use TikToken. So first thing is we just first we install it using just pip install tick token.

Sneha Mehra (00:59:47)  
And then we can just easily import TikToken here and then get an encoding from a certain model. For example, here I have TikToken.encoding for model GPT 3.5. And what that means is that give me the tokenizer that were used for training GPT 3.5 model. And then when I run this, I'll get this tokenizer object. And it's ready to be used. And you can see it's named ⁓ this. So it's equivalent of what we saw in the visualization.

this tokenizer. So after they train the tokenizer on some training data, they just name it something. Now this tokenizer is is ready to use and it's very simple to use it. We can just say tokenizer.incode and then ⁓ write, I love machine learning.

Sneha Mehra (01:00:39)  
And this is the sequence we would get back. And this should be equivalent to what again we saw earlier in the visualization. And now we can also decode, meaning that we can go from a sequence of text to ⁓ the its ⁓ original sorry, a sequence of numbers to its original text. So now if we pass something like this and call decode on it, if I'm expecting that it would give us I love machine learning.

And then we can just add new tokens to see what they ⁓ are, and then ⁓ it just keeps changing. So, ⁓ and then another thing I want to show is that we can have we can write ⁓ in ⁓ vocap, and then it would give us the vocabulary size. So this many tokens are available in this tokenizer. And now if I change this to something else, I can change it to GPT 4, let's say, or 4.0.

And ⁓ we saw that GPT four is the default GP ⁓ default model when we go to chat GPT. So let's see what tokenizer it's using. Now if I print tokenizer, it's the newer ⁓ O two hundred K base. And from the name it is probably having ⁓ higher number of vocabularies. It's probably having it's probably around 200k. So let's see. Now

Let me also rerun this. I love machine learning to see if it's going to output ⁓ the same numbers or not, which we do not expect to be the same. Now, if I run this, ⁓ it's the same. So at least for these tokens, it's the same. And then if I run this, I would get ⁓

Sneha Mehra (01:02:25)  
Hmm. Let me ⁓ then it's not the same. So let me

Sneha Mehra (01:02:34)  
Alright, so it's not the same as the GPT 3.5. The tokenizer is different and the mosh the vocabulary IDs are different. Now, if we just copy here and decode it, we would get the original text back. And now number of vocabularies is 200,001. So it's just bigger bigger vocabulary. And what that means is that it's just ⁓ more advanced and more effective. So ⁓

This is how easy it can be just to use tokenizers and start from some pre-trained tokenizers to train our LLMs. Now, back to our lecture. We talked about tokenizers and we saw how training works. And after training, how the tokenizer stores its internally and vocabulary. We also discussed the three key categories of tokenizers: word level, character level, and subword level. Word word level is problematic because if we have

one token for each word, our weak vocabulary would be huge. And then it's difficult to train on such a huge vocabulary. If we use character level, the vocabulary is a small, but the problem is that the the sequence ⁓ of tokens after we split the text can be very large. And also this is going to be problematic when we train LLM because it has to learn all the dependencies between il all these different tokens and it if if they're ⁓

A very long sequence of tokens, it's going to be more expensive and difficult for the LLM to learn from. And then we discussed software level tokenizers, which is balances balances between these two. So the vocabulary is more manageable ⁓ and the sequence length is also more manageable. And there are some algorithms, and we saw that the most ⁓ popular algorithm is probably BPE. And various LLMs use some variations of P BPE. ⁓ and then what the way it works is that it starts with characters and then it tries to

⁓ iteratively merge new tokens, merge tokens and form new tokens. And then this ⁓ merging tokens is based on how like what is the most frequent pair that they can just merge. And that's how they continue doing this ⁓ and continuing the algorithm until they reach a threshold for the vocabulary size. So and then we also saw some visualization and the TikToken library and how easy it can be to just use some pre-trained tokenizers. So

Sneha Mehra (01:04:53)  
That's all I wanted to cover for tokenizations. And then in the next lecture, we would start learning about ⁓ model architecture and understand how LLMs look internally.

Hello everyone, welcome back. In this lecture, we are going to focus on the model architecture. In the previous lectures, we talked about data preparation and wrap-top and we understood how we can prepare data and tokenize it. The focus of this lecture is to understand now how LLMs look like internally. ⁓ We'll start by reviewing neural networks and some important layers. And then after that, we will talk about transformer and decoder-only transformer architecture. And finally, we will

discuss why transformers are so powerful and widely used for text generation. So let's start.

So the first thing is what is neural network? And before answering that, let's understand the goal of most machine learning models. Most machine learning models are basically trying to learn a mapping from some input to output. I have three examples here. ⁓ And ⁓ if we look at them, in all of them, we want to learn a model to learn a mapping from input to output. On the left, I have house price prediction.

And then we can think of this problem as given some signals about the houses. We want to predict the its price. And those signals can be anything, like location, number of bedrooms, and ⁓ house size and those kind of things. And then somehow they get converted to numbers. ⁓ there are various ways to encode these these signals into numbers, and that's not the main focus of ⁓ this lecture. But ⁓ let's assume that there's some ways to convert these to numbers.

Sneha Mehra (01:06:42)  
And then price is also a number. So we can think of the input as X, which has all these X1, X2, and XN for different signals. And then the output is Y, which is price. Now we want the machine learning model to learn a mapping from X to Y. So ⁓ look at all the training data of X and Y, learn from it. And then when we give a new X to it, it should be able to predict Y. So it's essentially learning this mapping, this

Arrow here that I have to go from any X to Y. Now, in another example, for instance, email spam classification. And the problem here is given an email, we want a model to classify it as a spam or not not a spam. So we can think of, you know, ⁓ useful signals like sender, content, subject of the email, and so on. And then these get converted into numbers as the first step. So we can think of the input as X. ⁓

And then it has all this X1, X2, and XN. And then the output is basically expected to be whether it's a spam or not. So Y can be basically ⁓ one if the email is a spam and it can be zero if the email is not a spam. So the goal of the machine learning model here is to learn a mapping from X to Y. So again, while it's a different ⁓ problem, the goal of the machine learning is still the same. To just learn a mapping from a bunch of examples with X and Y.

And then the third example, like a tumor detection, here the problem is given an image, we want an ML model to f ⁓ find the location of the detected tumor if it exists. And even in this example, everything is basically can be converted to a number, and then the model's goal is to learn the mapping. Image can be represented by its pixels. And then location of detected tumors, ⁓ tumor can be ⁓ we can think of that as four numbers and it

The top left basically X and Y and bottom right X and Y. So if it's a rectangle, we just include the top left and bottom right. So at the end, there are just four numbers that the model is expected to map to. So even tumor detection is also a mapping. Now ⁓ neural network ⁓ is just a sequence of parameter parameterized transformations that maps an input to an output.

Sneha Mehra (01:09:06)  
So it's just one way that it starts from X, and there's just a sequence of transformations. It transforms X into some intermediate ⁓ representations and it keeps transforming and transforming until it reaches the final output, which is Y. And those transformations are learnable, meaning that there are some parameters, and then the goal of the machine learning model during training is to learn those parameters such that for that particular task, it can ⁓

keep transforming the input in a in a very relevant and effective way so that it its final output is Y.

And this is a simple visualization and example. So let's say we have a neural network here, and there is some input to it with four numbers, ⁓ 0.1, 0.5, and two more. And then this goes into neural network. Again, internally, neural network has some transformations. In this case, let's assume for simplicity, it has one transformation only. And ⁓ it just transforms it.

So one very simple ⁓ way of transformation we can think of is that it also internally keeps four parameters or weights. And then these weights get multiplied by the ⁓ different x values, and then we just add them up. So we can have something like w one times x1 plus w two times x x two and so on. So at the end, basically this neural network is just transforming this using a linear

⁓ weighted combination of ⁓ its values and then the output is going to be one number so this is just a very simple way of transforming some input to output using some ⁓ par internal parameters but in practice it can be ⁓ more complex and neural network can have multiple layers and each layer is transforming ⁓ its input to some output

Sneha Mehra (01:11:06)  
And for example, layer one transforms x1 to x2 using its internal parameters. And then x2 becomes an intermediate representation which goes to layer two. And layer two again transforms it. And it goes all the way to layer n. And then layer n transforms its input, and its output can be the final output of this neural network. So again, all I wanted to show here is that neural network is simply just a sequence of transformations.

To learn a mapping from inputs X to outputs Y. Okay, ⁓ now let's zoom in into these layers and understand ⁓ what are some common layers in neural networks.

Sneha Mehra (01:11:54)  
So the first layer that is very commonly used in neural network as a building block is linear layer. And linear layer, again, as the name suggests, it's just a linear transformation of the input. So for example, if the input is X, it internally has this W and B parameters, weights and BIOS. And these are the parameters of this linear layer, meaning that it would learn these during the training time.

And then it just linearly transforms X by using this formula. And then it it gener it create it calculates the output y. So we can see this in an example here. Let me zoom in. So for example, if the input is x with these four values, what that means is that internally it has to ⁓ it's a linear transformation. So we have to have a weight for each each of these values.

So that we can just multiply them. So the weight matrix is going to have a shape of one by four because the output has one number and the input has four numbers. And then the BIOS has just one number. ⁓ And ⁓ we'll see why, but just at a very high level, after we combine all the multiply all the weights with the X and just add them up, then we add it with a BIOS.

So this is just the formulation here. We have this P here. Now, another way we can think of linear layer is something like this. We have four small circles here, and this corresponds to the inputs. So, for example, inside this ⁓ circle is 0.1, inside the second circle is 0.5, and so on. And then we want to go to one number here, which is the output. ⁓ it's 0.6. ⁓

And then we have these connections here, and these connections are the associated weights to these connections. So, and since their weights has a shape of one by four, it's it's coming from here because ⁓ there are just four connections. ⁓ and then the weights can be named this way, like w one one, one, two, one, three, one, four. And now imagine if we had more outputs, like one more output here, we would end up having two circles here, and then the weight shape would have been ⁓ two by four.

Sneha Mehra (01:14:18)  
Because now we need we won we would have four more connections to the second neuron, to the second circle. Now, this is also where the name neural network is coming from, because these circles can be, we can think of those as neurons. And then these connections ⁓ can be also think of as ⁓ connections between neurons. So this is how a linear layer transforms from input to output. Very simple formulation. And now this can be.

generalized into ⁓ into ⁓ more outputs. For example, if the input has more ⁓ more numbers and then it's expected to produce an output with more numbers also, then the weight can be more complex. If we just visualize it, we would have more neurons, input neurons and w more output neurons. And as a result, the weight matrix would be ⁓ bigger. It's basically four times this many neurons.

But ⁓ this is all ⁓ a linear layer is. It's just basically ⁓ a weight matrix that is stored internally and will be learned during training time. And then the weight matrix is used to transform input to output.

Now, because of that, we can think of linear layer as just a mathematical expression that converts X to Y. So at the end of the day, linear layer is just a mathematical expression. And then it can be written down like this. For example, this neuron can be calculated by ⁓ multiplying weight with the inputs and just add them up. And also more output.

Neurons can be computed in a similar way. So linear layer is just a mathematical expression. Now, linear layer is not the only layer that is available. There are a lot ⁓ more layers that were developed over the years. I have listed some of them here. For example, we have convolution layers, we have activation layers, we have ⁓ attention layers, and many other examples. And each of these layers having different purposes and different strengths.

Sneha Mehra (01:16:33)  
And they're designed for different reasons. So ⁓ for example, convolution layers are designed so they work better with spatial inputs like images. So I'm not going to ⁓ talk about these layers in detail. There are some wonderful resources out there ⁓ which you can take a look. But what I wanted to share is that all these layers that are inside neural networks, it's

⁓ one of these. It can be linear layer, it can be convolution layer, and all these layers, what they differ is basically how they formulate their transformation. They just formulate it differently. For example, we saw linear layer here. It's it's ⁓ it formulates the transformation as a linear transformation. Now, this is going to be different for other layers. So that's how these layers are different, but they are.

They all are similar in a way that they just convert some transform some input to output. Now let's go to the ⁓ to the next part, which is basically neural network is just a sequence of layers. We just saw that ⁓ very early in this lecture. So what that means is that, for example, ⁓ neural network does have multiple layers, in this example, linear layers. If we visualize it, it would look like something like this.

For example, the inputs come here. We have linear layer ⁓ one weights here. It transforms the input to some intermediate output. ⁓ Now this goes into linear two. Linear two has this weight matrix here. And then it transforms this intermediate representation into another intermediate representation, and so on. And then finally, the final linear layer transforms its input to some output here, and these are the weights. ⁓

We saw that how linear layer is just a mathematical expression, but if we think of this neural network, this is also a mathematical expression. So what that means is that we can write x using some ⁓ mathematical expression that is ⁓ calculated using x.

Sneha Mehra (01:18:49)  
And we can just think of that as something like this here. ⁓ this can be the just the the ⁓ expression to to calculate ⁓ y's. And you can see it is based on all the weights of ⁓ linear layers that are inside the neural network, and then it just gets some it's a bunch of multiplication and additions, and that's it. So neural network as a whole is also a mathematical expression that transforms X to Y.

Now, over the years, research teams and ⁓ companies have tried combining various layers and build a neural network that works better for certain tasks. There was a lot of novelty and innovations into how to combine all these different layers in a unique way that the output, which is the neural network, can learn the task very effectively. And then in 2017, Google published a paper.

named Attention is All You Need and then it introduces introdu introduced transformer architecture. And transformer architecture is also a neural network and it's just basically a unique way of combining all these layers that we just saw, that it works really well for certain tasks. So here is the paper, Attention is all you need, and it was published by Google Print Team.

And ⁓ it's just the explanation and ⁓ how they came up with the architecture and how it looks like. But here in figure one, they have the the architecture of the transformer. Again, ⁓ feel free to read the paper to understand all these ⁓ components. But what I want to share here is at the end of the day, this architecture is also just a combination of layers that we saw earlier. And as a whole, this can be we can think of it as a neural network.

And this is basically a unique way of just putting and stacking all these ⁓ available layers, just stacking them and then ⁓ training it. So this was the key innovation of Transformer, ⁓ and it works really well for their task. And initially in this paper, the task that they were trying to solve was machine translation, meaning that they wanted to translate text from one language to another another language.

Sneha Mehra (01:21:11)  
So the input here is, for instance, the text in the source language. And then it goes through all these transformations here. And then the output is basically, we'll talk more about the output of transformers in a bit. But the output is basically eventually is going to be the translated text in the target language. ⁓ Now I have this screenshot of this diagram here in the transformer part of the lecture. And again, it's just

a unique way of combining layers and building this neural network. But at the end, if you think of this, this is also a mathematical expression as we saw earlier, because each of these layers are some mathematical expression. And then just stacking them together, we still can compute the output based on the inputs.

So ⁓ it if we just visualize this, it can be something like this. This image is just ⁓ is not real. I generated it by ChatGPT. I just wanted to show that even Transformer can be written in some mathematical expression. ⁓ basically it can be written in this format. ⁓ And ⁓ if you are interested to learn more about transformer and ⁓ what are the ⁓ reasons behind each component of it.

There is this great ⁓ illustration by Jay Alamar. And it's a wonderful resource to go through and learn about the high levels and how Transformer looks like. See, here is the input in the source language. Then here is the output. I am a student in the target language, which seems to be English here. And then it talks about all the components, encoder, decoder. Basically, encoder is ⁓ this part of the architecture on the left.

Because it's encoding the input text, which is in the source language. And then it has also decoders, which is this ⁓ right ⁓ part of the transformer. And it's because now it starts to decode the encoded input text to some ⁓ output text in the target language. But again, and it did it goes deeper and talks about all the details of all the components. So it's going to be an interesting read if you are not too familiar with the transformer architecture.

Sneha Mehra (01:23:30)  
So ⁓ this transformer was really, really powerful for machine translation. And then later we realized that if we only keep this decoder part, meaning the the right side of this, it would be really powerful for text generation. And then this is what we called ⁓ decoder only transformer. So it's nothing but just the decoder part of the original transformer architecture. So it looks like this.

Again, it's just a combination of layers stacked ⁓ and some of them are like in a like blocked because these are just supposed to be repeated n times. ⁓ so we have time times n here. So what that means is that ⁓ this hole is considered a transformer block and we have n block stacked right after each other. ⁓ So this

works very well for text generation and LLMs or large language models are also a text generation model. So ⁓ most of the LLMs that I'm aware of ⁓ are all based on decoder only transformer. So what that means is that when whenever we hear LLM, it means that it's just a decoder only transformer trained on internet data.

So ⁓ in the next lecture, we'll talk more about the input and output of transformers, and we understand why this ⁓ architecture, how can it be adapted for text generation? All right. We saw that decoder-only transformer is really powerful for text generation. So let's understand why. And to answer that, let's first see what is the input and output of the transformer architecture. So

As we saw in the previous lecture, the transformer is just a sequence of transformer blocks, and each block has some layers, and they're just stacked, and they transform input to output. But what is the input and output? The input that these layers expect is a sequence of vectors. This is just how these layers are expecting their input. ⁓ for example, here we have one vector, ⁓ we have three vectors.

Sneha Mehra (01:25:51)  
These three vectors, and again, each of these vectors have are having these small cells. Each cell is basically a single floating point number. ⁓ And so these three goes into transformer. Transformer just keeps transforming them. And then the output of the transformer is again a sequence of vectors. And as you can see, the size of the vectors are the same in this example. And it's in most ⁓ architectures, it's the same. Theoretically, they can be different, but we just keep them the same.

And then the output sequence length is also the same. So we have three inputs and three outputs. So this is the input and output of the ⁓ decoder only transformer. Now let's also understand what is the input and output of an LLM for text generation. So in LLM, again, for text, they are converted to tokens. So each text is a sequence of tokens. And in order to generate text, what we want an LLM.

To do is ⁓ we just pass some sequence of text to it. For example, here we pass these three numbers ⁓ corresponding to high how r. And then we expect the transformer, the sorry, the llm to output the next ⁓ token. And here it's 944, which corresponds to U. So this is our ideal input output that we want to train or build LLMs.

Transformers are expecting a sequence of vectors. What we want from the LLM is ⁓ a sequence of numbers going to the LLM and then the LLM just predict the next ⁓ token, just and also a number. Now, how should we adapt this? How can we really ⁓ modify and add new layers to this architecture so it becomes its input and output becomes compatible with this?

Now, ⁓ this is how we are going to do this.

Sneha Mehra (01:27:54)  
For a given text like ⁓ hi how r, first we discuss how it goes to the tokenization and gets converted to a sequence of numbers. Now, this sequence of numbers goes into an another layer, we call these layers embedding layers. And what embedding layer does is it converts each number into a vector. And it's also learnable, meaning that during training, this embedding layer would be learned so that each ⁓ ID.

gets converted to a sequence of vector sequence of numbers or embedding that is meaningful or useful or effective for the model to ⁓ predict the next token. So here for instance, these three tokens, these three IDs goes into embedding layer and then the output becomes three vectors. And this is exactly what the transformer expected as the input. So by just adding these two layers, we can easily adapt it to work with our input.

⁓ input text. Now let's get to the output. The output of transformer is three vectors corresponding to each of the three input vectors. But what we want is the next token. So the typical way to make this work is we discard the first few vectors and we only keep the last vector. And now this last vector is again, it's just a vector, it's just ⁓ some numbers, floating point numbers.

We add a linear layer on top of this. So we transform this vector into a new vector. And this new vector can be interpreted as probabilities of the next token. So its length would be equal to our vocabulary size. For example, if we have 50,000 tokens in our vocabulary, the output of this transformation ⁓ by this linear layer would be a vector of size 50,000. And then the value of its cell.

corresponds to the probability for that particular token in our vocabulary. So basically this linear layer is just mapping this final output into a ⁓ probabilities of different tokens. And again, there is just also a softmax operation which is responsible to convert these into probabilities. We skip that for simplicity. But we can we can think of these layers to convert the final output of the transformer into some probabilities for each token.

Sneha Mehra (01:30:22)  
And now these probabilities can be ⁓ we can think of those as something like this. So, for example, ⁓ this entire thing we had here, this tokenization embedding layer, ⁓ decoder-only transformer blocks, and the addit additional linear layer here. All of this we can think of it as the LLM architecture. So ⁓ here I have this LLM. I hope you are goes into the LLM, and then after the linear transformation.

The output is this vector with five values here, five probabilities. And again, this is only for visualization. In practice, this vector would be of length 50,000 or whatever our vocabulary size is. And then these are associated to our tokens. For example, we can think of this as the probability distribution over the vocabularies. So in this case, if the ID zero corresponds to end-of-sentence token, it has two percent chance.

For A Bell, it has 11% chance and so on. And we can see the highest probability is well, ⁓ with 86% chance, which also makes the most sense when the input is I hope you are, because well is the most likely token and also make this whole sentence meaningful. So this is how basically we go from the original decoder only transformer blocks. We combine it with some additional layers to make it work for text generation and ⁓ learning.

dependencies between texts. Now ⁓ the next thing I want to talk about is if ⁓ if you think of all modern LLMs these days, they're all based on decoder-only transformers. What they differ is basically some hyperparameters in the decoder-only transformer. For example, this is the screenshot I took from GPT-2 paper. And then they have four variations of models. And

Each variation has different number of parameters just because of the different hyperparameters used to build them the architecture, to build the decoder-only transformer. And some of these ⁓ hyperparameters are, for example, layers. ⁓ and layers refers to that times n that we saw in the figure, like how many transformer blocks we want to stack. For example, their smallest model, they stack 12 transformer blocks.

Sneha Mehra (01:32:46)  
And then this is the dimension of the vectors, like the input vector size and the the basically the size of the vector that this transformer works with. And then they use 768 for the smallest model. And they just play around with these numbers to get ⁓ bigger models. So when they increase the number of layers, it means that there are more ⁓ blocks and more layers, and each layer internally has parameters. So we just simply get bigger models with be with more parameters.

And this generally means that the models become more powerful because when they have more model, when they have more parameters, it means that they are more ⁓ they have more capacity to learn more complex mappings from input to output. But they are also more expensive to train because if there are more parameters, we have to train them for longer ⁓ and we have to also spend more computing power to train these models. So this is GPT two, ⁓ and their biggest model has ⁓

⁓ around ⁓ you know one ⁓ wait ⁓ yeah one thousand five hundred forty two million parameters. So it's around basically 1.5 billion parameters. Now this is GPT 3\. Let me make this a little bit bigger. So GPT 3 is basically ⁓ the next model that OpenAI trained and released after GPT two.

And what they did was they just scaled the models. So they increased the hyperparameters. So they trained bigger and bigger models with more capacity. So for example, here ⁓ they have GPT-3 is small and other variations, and their biggest model is GPT-3 175 billion parameters. ⁓ so ⁓ this is very big, and we'll see shortly that working with ⁓

These big models can be very challenging, engineering-wise, and it requires a lot of engineering effort so that we can load them and train them and later serve them. But here you can see the different ⁓ hyperparameters that were used. ⁓ for for instance, for the biggest model, they're using 96 layers. And then the dimension that the ⁓ transformer is working with is 12,000 something. And and more hyperparameters. And also I included Lama three from Facebook.

Sneha Mehra (01:35:12)  
But again, the same story. ⁓ three variations. And this ⁓ last variation here is even bigger than GPT three. It has four hundred five billion parameters. And then these are some of the hyperparameters that they've been using for number of layers, model dimension, and some other ⁓ parameters.

So ⁓ this is all I wanted to cover for in the model architecture. We saw that everything is basically just based on a decoder-only transformer, and they just ⁓ keep changing the hyperparameters to increase or decrease the capacity of the model. And then typically increasing the number of parameters means that the model would be more capable after training. In the next lecture, we'll talk about how to train the decoder-only transformer model.

Hey everyone, let's focus on the model training of LLMs. In the last lecture, we saw how LLMs look like internally and what's their model architecture. We also saw that whatever input we pass them, like I hope you are, it would produce a pro a ⁓ a vector which represents the probabilities of different tokens. However, if we do not train the model, these probabilities would be random.

And the reason for that is because all the model parameters are random. So LLM has these layers, and layers are keeping their weights, and those weights are random. So when we pass this, I hope you are as a basically sequence of vectors. It goes to the LLM. And remember, this entire thing, as we mentioned last time, is just a mathematical expression. And since all the weights are random, that means that the output would be random. So all the probabilities that would be produced by the LLM are meaningless.

So we cannot really use this and rely on it to produce the next token. The purpose of model training is to ⁓ use some processes and algorithms to tune these parameters that are internal to the LLM. So each layer would keep changing its parameters after we expose the LLM to the entire internet data, such that it can more accurately predict these probabilities. So let's ⁓ understand this in more detail.

Sneha Mehra (01:37:32)  
So ⁓ we have this internet data, which is cleaned and also converted to a sequence of IDs. Here, just for simplicity, in my example, I'm just showing them in text, in pure text, just so we understand easier. But in practice, this text and this text that goes into the LLM are basically just a sequence of IDs. So, how training works? Basically, we have this internet data, and imagine this is some

paragraph somewhere on the internet and we we have thing something like Albert Einstein was a German-born physicist and mathematician and so on. So the goal of training is to just expose the LLM to some part of this paragraph and then we ex we we force the model to predict the next token accurately. So what that means is that for instance we can just sample some part of this like Albert Einstein was a German born

And then we pass it to the LLM. From the paragraph in our training data, we know that the next correct word is physicist. So when we pass this to the LLM, we get the probabilities of the next token. We also know that what is the correct token. So this vector we can think of that that everywhere is zero except the ID of the correct token, the ID that ⁓

the ID that corresponds to the physicist token. So that ID is one, let's say it's here, and everywhere else is zero. And these are probabilities. And the goal is to try to ⁓ update the parameters of the ⁓ LLM so that the probabilities are as close as possible to this vector.

For that, we have to define a loss function. And there are just ⁓ typical loss functions that are used for different problems. And basically, what the loss function does, it measures the quality of the ⁓ current predictions. So let's say these are the probabilities. Cross entropy loss is the common loss for this purpose that we use. We use that, it's just a formulation to compare these probabilities with the correct token and just would give us a number. And that number, if it's high.

Sneha Mehra (01:39:47)  
It means that these two vectors are not similar. So the loss is high. And if the probabilities are more similar to the correct token, the the loss would be low. So this is the first step of training. We sample some text, we pass it to the LLM, then we use a loss function such as cross-entropy loss to measure the quality of the predictions. And again, cross entropy, if you ⁓ search Google or go to Wikipedia.

The ⁓ formulations, how crossing tropies ⁓ can be calculated. This is the formula. And there are some details. ⁓ Where is it coming from if you are interested to learn more about the theories? But again, at the very high level, it just measures how close the probabilities are to the correct token. Now, once we have this loss, it's now time to update the parameters of the LLM. And that's the purpose of the second part, which is optimization. Optimization is just an algorithm.

That ⁓ uses the ⁓ calculated loss value, and then it goes and updates all the parameters of the LLM based on some algorithm. So it updates them such that if we pass the same input to the LLM, it makes ⁓ more accurate, it predicts more accurate probabilities next time. So that's that's all it happens. So now if you repeat these two steps multiple times, many, many times on ⁓

Different samples from the internet data, eventually we would have very good weights for the model, meaning that we the model would end up with the weights that if we pass something like Albert Einstein was a German born. And if you look at the probabilities, ⁓ the probability for the token physicist would be very high. And that's going to be the outcome of the training process. So we continuously sample text from our training data, we pass it to the model. The model ⁓

output pro probabilities, we calculate the loss, compare it with the correct next token, and then we apply optimizer to update the parameters of the LLM. And after we do this enough time on the entire internet data, we would end up with a very good ⁓ next token predictor model, meaning that whatever we pass, like Albert Einstein, we get the probabilities. And those probabilities are ⁓ basically statistically very accurate in a sense that

Sneha Mehra (01:42:12)  
If on the internet we see in a lot of places we see Albert Einstein was, the probability of was would be high. So in other words, basically it learns all the dependency and statistics between all different tokens from the internet data. And then it has some implicit knowledge of the world. And that's because ⁓ if the model can really predict the next token ⁓ really well, it means that it has some understanding of different domains and different areas and

some implicit knowledge. Imagine, for instance, we ask the we give the LLM something like, I like machine learning because the next token and also the ⁓ the future tokens that it would ⁓ keep generating ⁓ is probably very relevant to machine learning. So it would probably output some terms that are related to machine learning ⁓ and ⁓ to why people like machine learning.

And all of these are statistically very close to what is available on the internet. So that's the outcome of the training. Now remember, these models can be very difficult to train, even though the process is very straightforward, because these LLMs can be very huge in terms of number of parameters, it can be ⁓ engineering wise, it can be very difficult and a lot of challenges to make sure that we have the right setup to train these models.

For example, I asked ChatGPT that what are the resources needed to train an LLM with ⁓ 405 billion parameters? And remember, this is what Lama 3, ⁓ biggest version of Lama 3 has. So why is ⁓ what is the memory needed? I misspelled what bit why. ⁓ and then the answer is ⁓ it requires lots of things, and then here it says ⁓ why.

Memory is needed. It's first because we need to store the model parameters during training. So all those parameters, 405 billion parameters, needs to be stored ⁓ in memory. And then during the training, again, there is some algorithms and optimizers need to keep track of all the intermediate ⁓ representations and ⁓ activations and use those to update the weights. So there are a lot of intermediate ⁓ floating point value numbers that it should store. So ⁓

Sneha Mehra (01:44:37)  
Because of all of these, and here is a rough estimation, it requires ⁓ like you know very large memory. Assuming that we use ⁓ FT FP32, meaning that we use four bytes to store each weight, we would end up needing around 1.6 terabytes of memory. So definitely it's not possible to ⁓ train this entire model on a single GPU because

The best GPUs we have might have around ⁓ hundreds of gigabytes of memory. So it's not practical to really train this big model on a single GPU or in on a single machine. And again, in its calculation, it's saying that you need at least eight hundred eighty to one hundred GPUs just to fit the model in memory. And in practice, it requires more. Realistically, it says you use ⁓ two thousand GPUs or more.

And these numbers are accurate ⁓ based on what h happens in practice. Also, here in terms of a storage, we need hu huge storage as we train the model to just keep saving ⁓ the checkpoints of checkpointing the new models as we train them. So each time each checkpoint might exceed two to five terabytes of ⁓ the storage needed.

So all this is basically saying that it can be very challenging to train ⁓ very large models, very large LLMs. It also requires lots of distributed training and so on. ⁓ and then let's go to the Lama 3 paper. This is ⁓ the technical report of Lama 3 published by ⁓ Meta. And what I want to show here is that if you scroll down, there are this ⁓ by the way, this is a great

Technical report, it's very ⁓ very useful to read and ⁓ learn I personally learned a lot of things from it. So but what I want to show here is that ⁓

Sneha Mehra (01:46:47)  
All right, so what I want to show here is that ⁓ if we scroll down, there is this pre-training stage, which we discussed in the last couple of lectures. And then they are talking about pre-training data. ⁓ And ⁓ they manually apply this cleaning and different basically here they're saying that ⁓ there are a couple of steps. The first step is the cre curation and filtering of larger scale training corpus. This is we already discussed about this.

Then they talk about model architecture and finally the model training. So pre-training data, they start with something, they apply some data cleaning mechanisms, and here they explain more about ⁓ safety filtering and ⁓ text extraction and cleaning from HTML content and so on. And deduplication. We saw all these, how why they are important and how it works. And then after the data preparation, they talk about ⁓ model architecture.

So they start with Lama 3 uses a standard dense transformer architecture. So as we mentioned before, most LLMs are all based on decoder-only transformer. ⁓ And after model architecture is the training. And then here basically this part talks about infrastructure, scaling and efficiency. And

If you just read it and go over it, there are a lot of details about the storage and ⁓ compute that is required. And they are, for instance, they are training this model on ⁓ 16,000 GP ⁓ H100 GPUs. And H100 is one of the more powerful GPUs that are available in the market. So they are using 16,000 of those GPUs. And then here also they are talking about how storage, big storage is required and you know, network ⁓ challenges.

And then after that, they talk about all the optimization and techniques that they have to use. So they can only fit everything in memory. And they are talking about all different parallelism parallelism techniques for modular scaling and ⁓ you know pipeline parallelism. So ⁓ what I want to show is that they are ⁓ and and see this section is very long. So they're doing a lot of tricks and a lot of engineering work just to make training possible on such a big model. So

Sneha Mehra (01:49:08)  
Just engineering was, I just wanted to say it's difficult. But in theory, it's just these steps. If you look at the code, the code is just ⁓ calculating the loss using a standard cross entropy ⁓ loss function. And then using some standard optimizer to keep updating the model parameters. And then we keep doing this on the entire internet data. And then the outcome would be a model that is really good at predicting the next token based on the statistics of internet data. So ⁓ I covered

all I wanted to cover for the model training.

Hello everyone. In the last lecture, we wrapped up model architecture of LLMs. In this lecture, we want to talk about text generation. Given that we have a trained LLM, how can we use it to generate text? From the last lecture, we learned that if we use the LLM and we pass some input to it, like Albert Einstein, the LLM would produce probabilities for the next token. So it's going to be a vector, and each value would ⁓ represent the probability of that particular

token. Now for text generation, this is not enough. What we want is we want actually some more text given input text. For example, if we are given Albert Einstein, we want the LLM to generate more text in a meaningful and contextually relevant way. So that's the purpose of step four, which is text generation. So to do that, text generation is an iterative process. We iteratively generate tokens

And we continue doing that. It's basically a for loop. So to better understand that, let's just walk through a real example. Let's say we pass Albert to an LLM, and then the LLM produces probabilities. Now, in order to go from these probabilities to an actual token, we need some strategy or some algorithm to pick one token from these probabilities. Let's assume that there is some algorithm that does that for us. So it picks Einstein.

Sneha Mehra (01:51:11)  
From these probabilities. And then now we have Albert Einstein. So we pass we can pass it to the LLM again. ⁓ And the LLM produces probabilities for the next token after Albert Einstein. And then the algorithm picks was. And then we can repeat this process. So we can just ⁓ repeat this process and keep getting new tokens until ⁓ one of two things happen. So the generation ends when ⁓ one we reach end of sentence a special token.

This is a special token that we can use when training the LLM so that whenever the LLM outputs that, it means that the sentence is ⁓ complete and we no longer need to generate more tokens. So this is one way that the generation can end. The other way is that the LLM never produces end-of-sentence token and we reach the maximum desired length. So ⁓

As soon as one of these two things happen, we kind of stop this iterative generation and we would end up with a ⁓ continuation of the initial sentence that we had. So I have a more compact ⁓ visualization of this ⁓ thing here, and then it's here. So basically we start with something, let's say hi, and then the LLM produces probabilities and the algorithm picks the next token, which is how, let's say.

And then how comes here. Now the new input is hoi, how goes to the LLM. Then we pick R. And we keep doing this. In this case, we reach end of sentence token, meaning that the LLM at some point, which believes that the sentence is completed, it produces end of sentence special token. So we learned about this iterative ⁓ generation process. Now the only thing that we need to answer is the algorithm here. How can we really pick a token?

given the probability distribution. And that's what we are going to talk next.

Sneha Mehra (01:53:09)  
So basically, there are different algorithms and different strategies. We often call them decoding algorithms or sampling algorithms. And these algorithms specify how to choose a token from a probability distribution. So given that the LLM outputs something like this, which can be interpreted as this ⁓ probability distribution, the algorithm specifies how to pick the next token from these probabilities.

And there different algorithms for that purpose, and they can be categorized in different ways. So here we have ⁓ at a very high level all the text generation methods or sampling methods. They can be categorized into deterministic and stochastic. Deterministic are the algorithms that the ⁓ there is no randomness in the generation process. So if we repeat the generation using the same input and same LLM, it would always produce the same output.

A stochastic algorithms, on the other hand, has some randomness, meaning that using the exact same LLM and exact same input, if we run it twice, run text generation twice, we would get two different continuations or two different generations. And each of them has their ⁓ pros and cons. So we'll start with deterministic and then we'll switch to a stochastic. Deterministic, there are two popular algorithms. One is greedy search and the other is beam search.

Again, we'll briefly talk about this to better understand the evolution of different ⁓ text generation algorithms. But I want to also point out that in most advanced LLMs, we they no longer use any like greedy search or beam search, and they ⁓ use this top piece sampling here, which we will talk about. But it's it's very good to just have some ideas about how these ⁓ search algorithms work and what are their limitations.

So we'll start with greedy search.

Sneha Mehra (01:55:12)  
Greedy search is basically the simplest form of ⁓ picking a token from the probability distribution. It as the name suggests, it it greedily picks the highest probability token and then use it as the next token. So it's just ⁓ very simple. So let's say, for instance, this is the ⁓ predicted probabilities, given the input token how. So the model has predicted these values for different tokens. ⁓ In this example, the greedy search simply picks R.

Because it has the highest highest chance of being the next token. It has 37% chance and it's more than all other tokens. So another way to visualize this is ⁓ like a tree. So we start from this input ⁓ token, which is how. It's passed to the LLM, and then the LLM ⁓ outputs the probabilities for all different tokens in the vocabulary. For simplicity, we visualize three outputs here, but in practice it's

equal to the number of tokens in the vocabulary. So let's say it's 50,000 values. So at this step, the algorithm picks the highest probability token, which is R with 56%. ⁓ And now we have how R it's passed to the LLM. The LLM again produces ⁓ 50,000 probabilities in the next step. And the greedy algorithm picks U because it has highest probability. And it just keeps repeating this. So at the end we would end up with a path in this tree.

like how how are you ⁓ doing. So the advantage of greedy search is that it's very simple and efficient. It it simply picks the highest probability at each step and it can be easily implemented. However, there are lots of limitations with greedy search at it's not and it's not really used in ⁓ in practice for text generation in LLMs. The key limitation of greedy search is that it's not really looking ahead.

So at each step, it's only picking the highest probability token without looking ⁓ the probabilities at the ⁓ future steps. So this may lead to suboptimal generations because there might be better ⁓ sequence of tokens in other branches. For example, here, even though R has the highest probability, there might be a different branch that if the model, the text generation explores that, it may lead to overall, it may lead to something with higher probability.

Sneha Mehra (01:57:37)  
If we just consider the probability of all the tokens in the sequence. But greedy search does not that. Greedy search just picks the highest probability at each step without looking ahead or considering any other branches. The other key limitation of greedy search, it may lead to repetitive outputs. So what that means is that you would see a sequence of tokens that keeps repeated ⁓ in the generated text. And the reason for that is because

There might be some cases where a sequence of tokens has very high probability and are very common and also commonly ⁓ appeared on the internet. So the greedy search keeps picking those sequences of tokens because it's just preferred. They have higher highest probability statistically, and the algorithm just ⁓ prefers to deterministically just keeps ⁓ picking them. So that's greedy search before we switch to the next algorithm.

Which is beam search. Let's see greedy search in example in in practice. So I have this code here in my Jupyter lab. ⁓ and then there are some different libraries. Again, the purpose is not to ⁓ be super familiar with some of these libraries and how they work, but there is this Transformers library, which easily gives us access to various pre-trained models and also their tokenizers. So I'm using that to get a model.

GPT two model and also its corresponding tokenizer. So let me run this.

Sneha Mehra (01:59:13)  
And then the next thing I want to do is I want to start with some ⁓ input sentence. For example, I enjoy walking with my cute dog. Now, what needs to happen is that this ⁓ text first needs to be converted to a sequence of IDs using the tokenizer. So I apply, I run the tokenizer on this sentence, and it would give me model inputs. And model inputs is simply a sequence of IDs or a sequence of ⁓ IDs corresponding to these tokens. So let me run this.

Now, if I print ⁓ model inputs, ⁓ we would see a sequence of numbers 40, 28, 83, and so on. And these are corresponding to this individual ⁓ subwords or tokens. So now we have model inputs, we have the model, we have everything ready. The next thing is we have to now run text generation using greedy search. And we don't have to implement greedy search from scratch. It's already implemented, and there is this method called generate.

With some inputs. Again, there are some good documentations to understand what are the inputs. But my purpose here is to only show what it means to run greedy search on some input using GPT two model. So here I'm passing all the ⁓ the sequence of IDs. And also I set max new tokens to 40\. And what that means is that do not generate more tokens after you generate 40 new tokens. And

⁓ this is necessary because otherwise the model may generate a lot of tokens and it may never end. So now I run this and let's see what is the outcome.

So, all right, here this is the output. It starts with enjoy I enjoy walking with my cute dog. This is my original sentence. And then everything after this is what ⁓ the greedy search algorithm generated. And it starts with comma. So the continuation is comma, but I'm not sure if I'll ever be able to walk with my dog. I'm not sure if I'll ever be able to walk with my dog. And also here, I'm not sure, which is the

Sneha Mehra (02:01:17)  
continuation is still probably the same token. So the same sequence of tokens. So this is a good example of showing repetition. This this sequence of tokens, I'm not sure if I'll ever be able to walk with my dog. Is apparently very common sequence of tokens. They have high probability and the LLM, the GPT two model generates very high probabilities for this sequence. And then the greedy search algorithm is deterministic. So each time it just prefers to just

pick the same sequence of tokens over and over and over again. So this is greedy search and this is why it's not ideal to be used in practice for text generation in LLMs. Now let's get back to our ⁓ lecture and switch to the next algorithm, which is beam search.

Sneha Mehra (02:02:06)  
So the idea behind BeamSearch is to ⁓ f basically fix some of the limitations of greedy search. Remember, greedy search always picks the ⁓ highest probability token. What BeamSearch does is it keeps track of the top K path in this tree. So for instance, instead of picking the highest probability next token, it would always keep track of the highest ⁓ three highest probability.

Three passes with the highest ⁓ cumulative probability. So for example, if when we pass how to the LLM and we get all those 50,000 probabilities, the beam search would pick top three probabilities. For instance, let's say they are COM or do with these ⁓ numbers. And then it just discards everything else. So at the next for the next step, we have three paths, three ⁓ possible continuations.

or three possible sequence of tokens. And they are how come, how are, how do. Now LLM again for each of these, it's it's we pass to the LLM and the LLM generate the next the probabilities, 50,000 for each of these. And then out of all those 50,000, I think it's going to be overall 150,000, assuming that the output of the LLM has 50,000 numbers. So out of all of them,

The beam search comes and look at the the cumulative probability. For example, how come animals has ⁓ 24% here and 36% here? How are you has ⁓ 31% here, 911% here, and then how do you has also 26% here, ⁓ 63% here. So these three paths are better than all other passes here. So

The beam search discards all these ⁓ possible continuations and only keeps these three ⁓ sequence of tokens, which we have in bold ⁓ lines. So at second step we would end up with again three paths. How come animals, how are you, and who how do you? ⁓ And then it goes to the next step. So it keeps doing that again until ⁓ a certain number of iterations.

Sneha Mehra (02:04:31)  
it reaches. And then finally it just picks the highest probability path out of the three candidates. So ⁓ this is beam search. Again, it we we can think of it as an extension of greedy search by by keeping track of top K ⁓ possibilities at each step. ⁓ And the advantage of this is that it it may sometimes discover better continuations compared to the the greedy search.

For example, here in this case, it may at some point it may decide that how do you ⁓ is and it their continent and the cut the continuations of that has higher probability or is more likely to be a better continuation compared to how are you. So by just having keeping track of top top K ⁓ path, it most of the times it outperforms greedy search. So this is beam search ⁓ and

BeamSearch also has the similar issues of repetitions. A lot of times you may see repetitions when you run it on the LLMs. ⁓ But overall, you would see ⁓ more relevant continuations in most examples. So now this is BeamSearch, and then we covered deterministic algorithms. ⁓ And both of Greedy Search Beam Search are deterministic, meaning that there is no randomness if you if you run it multiple times on the same input.

We would get the same output. For example, here we if we if I run this again, I would get the same output. I enjoy working with my QDog and the rest of it. Now, the next category of ⁓ algorithms are stochastic with randomness. ⁓ And we'll start with multinomial sampling. That's the simplest way of sampling ⁓ among all these three algorithms. So let's start with that.

Multinomial sampling basically what it does is that instead of picking the highest probability token, it just samples according to their probabilities. So for example, if this is a ⁓ predicted probabilities by the model, the algorithm at this step would pick ⁓ R with 37 chance, it would pick 21%, sorry, is with 21% chance, and so on. So this

Sneha Mehra (02:06:54)  
This randomness basically allows it to discover other continuations and not always pick R as the next token. So that's multinomial, but there is a major limitation with multinomial. Because it samples according to their probability distributions. Sometimes, if they run enough times, it may pick tokens that are less likely ⁓ or very unlikely. For example, if we run it 100 times,

At some point, the next token that it may pick after how might be hard. Now imagine there are some more tokens here that can be even ⁓ they can be grammatically incorrect to be to come after how. But the multinomial sampling may occasionally pick some of those tokens. So that's why multinomial sampling is not really used in practice in LLMs, and some variations of multinomial sampling is used.

And top K and top P are both improvements over multinomial sampling to fix the issue that I just mentioned. So the next thing we can we ⁓ talk about is top K sampling.

So top K sampling basically instead of picking according to the probability distribution of all tokens, it first picks the top K tokens based on their probabilities, and then it samples according to the probability distribution. ⁓ So this way it discards the unlikely tokens and it does not even consider those tokens. For example, in this case, if top K is

Let's say k is equal to three and k is a hyperparameter, we can define it. But let's assume k is three and we want to run top k top three ⁓ sampling from this probability distribution. First, we keep the first three highest probability tokens, which are ⁓ r, is due. We discard everything else. And then out of these three, now we sample one token. So this is where the randomness is introduced.

Sneha Mehra (02:09:03)  
So top K is clear clearly better than multinomial sampling because it simply discards what is very unlikely to come next and only keep ⁓ likely words and then sample from those. So this is top P, but top K, but top K has also one limitation that ⁓ we'll talk about, ⁓ and that's what ⁓ top P sampling is trying to fix.

So let's go here. So imagine these two different probability distributions. In this example, we have probability of next token for tanks A. And then these are the probabilities. You can see tanks a lot. Lot has the highest probability, 89%. And then everything else is very low. Now, if we run top K algorithm on this, and K is let's say equal to three.

It means that we are going to keep the first three tokens and consider all these three ⁓ lot much high as a ⁓ as a possible as a candidate for the next token. However, the model is some fun somewhat confident that this is the token. We don't these are very unlikely to come next. But top K, since K is fixed, would always consider top three ⁓ tokens. Now imagine this.

⁓ probability distribution, which is the probability of tokens coming after how. For this model, top K again, it would consider top ⁓ top three, let's say, it would consider the first three tokens, ⁓ R is due, and it would ignore the rest. However, in this example, according to the probabilities, the model is also confident about some other ⁓ tokens.

So it's ⁓ it's basically it's more uniform, the predicted probability. So there are more tokens that can be good candidates to come next, to come after how. So all these explanations, what it what it means is that top K, because K is fixed, is not a good ⁓ is not an ideal algorithm. Because regardless of the shape of this probability distribution, it would always considers the top K tokens.

Sneha Mehra (02:11:32)  
When the model is very confident about, let's say, the next token, it would still consider top K. And when the model is not too confident and there is a huge possibility of candidates, the model would still pick top K. And that's why what top P sampling is trying to address. So what top P does is basically it it it does not fix K. What it does is that it picks the top K tokens that they're

Cumulative probability exceeds value P. Now this P is a hyperparameter that we can provide and we can pass to this. we can specify in this sampling algorithm. For example, we can say top P, P equal to 90%. ⁓ Or in this example, we have 88%. P ⁓ is equal to 88\. In the when it's 88, it means that peak.

The first few tokens such that their cumulative probabilities exceed 88%. ⁓ And in this example, because the model is relatively confident about the next token lot and it has already 89%, it exceeds our threshold. And top piece sampling would only consider this. In this example, since the model is less confident and there are more ⁓ possibilities.

The model would pick these four tokens because the first one is 31% ⁓ and it's not still exceeding 88% threshold. The next one is ⁓ 29%, ⁓ so their their sum is ⁓ 60 and still it's below 88\. So at this point, the sum of these tokens equal exceeds 88\. And then the model discards the rest of the tokens and sample from this. So basically top

P is just an improvement over top K with a more dynamic K. It dynamically changes K ⁓ and ⁓ it basically ⁓ it basically calculates, it basically dynamically decides on K based on the ⁓ cumulative probabilities. And top P is what is commonly used in practice in LLMs to generate text. Let's go back to my Jupyter notebook and see if I run

Sneha Mehra (02:13:57)  
Top P and the same input, what's going to be the output? So again, I'm going to use the same. I enjoy working with my Q. And then I have top P sampling here. Again, model.genate works with that. I we just need to speci add ⁓ specify do sample true and provide the top P value. In this case, I'm case I'm saying 92, meaning that I'm saying that ⁓ use the tokens that their cumulative probability probability are ninety-two.

And then discard the rest of the tokens at each step. Now if I run this, I'll get this output. I enjoy walking with my Qt doc. This is the original sentence. And now the rest is ⁓ comma and then says RM ⁓ whatever sixty-five, who is on a leash at home. ⁓ so again, it's not the best.

continuations and that's because GPT two is not the best model out there. If we switch to some better models, we would say we would see a lot better continuations. But at least we are not suffering from like repetitions or things like that. And also there is some randomness, meaning that if I rerun this, I would see a different continuation. Let me rerun.

Okay, now I'm clearly seeing something different. ⁓ I enjoy walking with my cute dog. ⁓ I don't know what that is, and then it's saying I love walking with my baby dog and having an empty apartment without dogs and other conveniences. Again, the these continuations are not the best, but it's not because our sampling algorithm is bad. It's because GPT two is not too good.

if we swe if we simply switch GPT two with something more powerful like GPT three or more even more powerful models, we would see a lot better continuations. And in future lectures we would do that to see how how different the output can be. So this is top P sampling ⁓ and ⁓ there are some great ⁓ there are some great ⁓ like tutorials about these algorithms and

Sneha Mehra (02:16:04)  
the strategies and how different they are, what is going to be the output, ⁓ you know, the details of the ⁓ arguments to use this generate and everything. So it's an interesting grid. I've included the links in the lecture. And also here is basically another very interesting tutorial how to generate text. Both of them are published by Hugging Face. So it basically starts with the introduction and then starts from greedy search ⁓ and it just goes and switches to a more advanced ⁓

algorithms. And then I think at the end it talks about ⁓ top K and top P and some examples. So it's it's an interesting read, but basically it just explains what we already discussed in the lecture.

So the next thing I want to share is that top P has this ⁓ P hyperparameter, and there are also some other hyperparameters that we can specify or introduce in our sampling algorithm. For example, we can always ⁓ we can have a ⁓ a hyperparameter to adjust the raw probabilities. So for example, when the model outputs these probabilities, we can just make it a little bit smoother or we can make it sharper.

So these are some of the things that are basically there is no best ⁓ or optimal values. It just depends on the task and depends on our model. What are the best values to for the P or some other hyperparameters? ⁓ so you would see some of those values when you try to call LLMs. For instance, here you saw that how we specify top P, and there are potentially more ⁓ more arguments here that we can specify. ⁓ And

I'm not going to go too deep into that. ⁓ just one thing I wanted to show here is that again, as I mentioned, it it's very empirical. We have to try an experiment based on the model and based on the task. And also for different tasks, there might be ⁓ different p-values ⁓ that are optimal. For example, this is just a very empirical ⁓ temperature and top p ranges for different tasks. Temperature, again, it refers to a hyperparameter that.

Sneha Mehra (02:18:14)  
smooth the raw probabilities. So we can use it to make the ⁓ raw probabilities ⁓ sharper or more uniform. So ⁓ for example for code generation it's been ⁓ seen that top P of 0.1 is a good number. And what that means is that basically we are saying the algorithm to only consider ⁓ very few tokens, not

not many tokens as candidates. And that's because code generation is tricky. It's important. We don't want to really ⁓ we don't care about novelty. What we want is that we pick the tokens in the code that are ⁓ in syntax syntax wise they are correct and they can be run. So we don't really want to consider all those ⁓ uncertain tokens that the model predicted. However, in more ⁓

Like creative tasks, like creative writing, we want to increase P because now we want more novel continuations as well. So we always want the model the ⁓ text generation algorithm to consider ⁓ less likely continuations that might be more interesting or exciting. So that's all I wanted to cover for text generation. ⁓ And in the next lecture, we'll switch to post-training.

Hey, we're finally going to focus on the post-training stage. In the previous lectures, we discussed and wrapped up pre-training, and we saw that after the pre-training stage, we'll get a base model that is really good at predicting the next token. And then we can run text generation to continue text ⁓ given some input sentence. Now, that base model is not ⁓ useful because it's going to just continue whatever input sentence is given to it.

And it's not necessarily going to answer your questions. Now, the purpose of post-training is to adapt that base model so that it answers questions and also be more helpful and safe. And post-training has two steps: supervised fine-tuning or SFT ⁓ and reinforcement learning. And we are going to talk about both of these in this lecture. ⁓ Before talking about SFT, I want to go back to our Jupyter notebook.

Sneha Mehra (02:20:38)  
and see what we mean by ⁓ when we say the pre trained model is good at continuing sentences and not answering questions. So this is again the same library transformers. I'm going to pull GPT two model and its tokenizer.

Sneha Mehra (02:20:58)  
All right, and we'll use the same sentence, I enjoy walking with my cute dog. ⁓ and then use top P sampling with ⁓ some optimal parameters. And then we'll run it to see what the output is. And remember, top P is the best sampling method that is available and mostly used in LLMs these days. So the output is I enjoy walking with my cute dog, my wife and my friend, they always have me covered with a blanket and so on. So the sentence is basically just a continuation of what we ⁓

provided as the input. Now ⁓ also remember that pre-training stage, we did we we previously said that it's the very expensive stage and its most important part because it learns from internet data. So this base model, even though it's just a ⁓ next token predictor or just continues the the inputs, it is a still very powerful model. It's basically it has all the implicit knowledge that it could have seen

in on the internet. So what that means is that if we change this to for instance, I like ⁓ machine learning.

Because it would continue this, but now it would continue and use words and terminologies that are very relevant to machine learning and why we like it and so on. So what that means is that it has some implicit knowledge about machine learning, what are typical words, ⁓ how they follow each other, and so on, which can make this model very powerful. So it says I like machine learning because it's easier to understand and you can use it to predict what people are going to say. ⁓

It's not like you can do that in an e-learning system. So again, it's not the best continuation. And that's because the model is not too ⁓ capable. It's just GPT2. But just the fact that it's using relevant, like contextually ⁓ reasonable words. For example, it uses predict or system or e-learning, those are all things around machine learning. Now, if I change this to a more complex ⁓ input sentence, let's say ⁓ in ⁓

Sneha Mehra (02:23:05)  
Quantum computing. ⁓ Let's see what it answers. In quantum computing, we can take an equation and generate an equation with the result that we can divide by the number of different elements in the equation. Again, it's it's not ⁓ it's not a ⁓ useful continuation, but still it's using some ⁓ kind of relevant ⁓ terminologies or keywords. So again, all I wanted to show here is that.

The base model ⁓ is capable. It has implicit knowledge of the word because it has seen internet data and it has learned from it. It has seen lots of different domains and areas and it knows how different words and terminologies ⁓ occur or co-occur ⁓ after each other in different domains. But the model is not really good at pred at answering questions. For example, if I change this to something like ⁓ like how is the weather?

Sneha Mehra (02:24:06)  
It's going to probably just how is the weather? Weather is one of the most common problems for people who are tired and stressed. So it's not really answering my questions, just because it's the base model. Now, we also discussed that more capable models are bigger. They have more parameters. They could have billions of parameters. For instance, GPT-3 had 175 billion parameters. And then more recent models like LAMA, the base model has 405, four.

four or five billion parameters. So it's not even practical to load it on a ⁓ on a single machine. And I'm not able to load that here locally and show you how it responds. But there are different services, online services that we can use to load those models and ⁓ interact with them. And again, there are many examples. One example is hyperbolic ⁓ and I just want to show that if I switch GPT2 to something more capable and bigger, how the response would be different.

So here we can ⁓ select a model. I have selected Lama 3.1405B base, and the name is clear what it means. Lama is basically Meta's ⁓ LLM. 3.1 is the version, one of the recent models. 405B is showing their number of parameters. It has 400 billion parameters. And basis means that this model is the outcome of the pre training stage. So this is the base model. It just continues the sentences. It's not going to answer questions.

Now let this is also the max tokens. I'm going to reduce this ⁓ to 60\. And then we'll keep this default. These are top P temperature. These are the hyperparameters to control the text generation process. And what I'm going to enter here, so in the notebook, we saw that for something like I like ⁓ machine learning because the output was not really too useful because it's easy to understand and I don't have to worry about that. He says

And this is just because the model is not too good, is not good enough. Now here if I enter the exact same prompt, I like machine learning because let's see what ⁓ llama would ⁓ respond.

Sneha Mehra (02:26:19)  
I like machine learning because it's a set of tools that can be applied to a variety of problems. You can use machine learning to predict the price of a house or the probability that someone will click on an ad. You can also use it to identify faces and so on. So again, as you can see, now the continuation is a lot more meaningful ⁓ and also relevant to the input sentence. So ⁓ that's why bigger models are ⁓ base models basically has an implicit knowledge about the word.

Because they are really good at understanding text and continuing them in a relevant way. So now if I also change it in ⁓ the other example in quantum computing, V, let's see how it responds. We often want to construct a circuit to perform a particular operation. One way to do this is to design a circuit directly. Again, I'm not an expert in quantum computing, but when I read this, it makes sense. So ⁓ again, the model seems to be doing well.

And then the last example is I want to now ⁓ enter a a question. For example, I want to see ⁓ I want to write what ⁓

Sneha Mehra (02:27:31)  
What is the best way to learn machine learning?

originally appear and appeared on Quora, the knowledge sharing network where again it's just com continuing it. It's not really answering my questions. So now back to the lecture. The focus of post training, as we mentioned, is that to adapt this base model, which is very powerful and has implicit knowledge, to adapt it so it answer questions instead of continuing them. And we'll start with the first step, which is supervised fine tuning.

Sneha Mehra (02:28:09)  
So again, we have supervised fine tuning as the first step. It's also sometimes it's called instruction fine tuning because we are fine fine tuning the base model so it follows instructions. And the goal of this stage is to ⁓ adapt the model from a completion model to following instructions. So for example, if from the base model we ask some a question like, I want to learn ML, what should I do? The model may continue with things like ⁓ with more questions, maybe.

⁓ something like is it even easy? I'm not so confident, and so on. But after this SFT stage, we'll get the SFT model. And then ⁓ if we ask the same question, it would actually answer the question. It would say something like take Andrew Yang's course on Coursera. Now, just out of cure curiosity, let me ask the exact same question from Lama 3 and see if it really answer if it really adds more question to it. ⁓ Or just answers it.

Sneha Mehra (02:29:12)  
⁓ I don't know what happened. Let me try again.

Sneha Mehra (02:29:30)  
All right, maybe maybe it's down. We'll we can try later, but ⁓ I'm sure it's going to answer we it's going to come continue it with more maybe questions or it's not for sure gonna answer it. So back here. In order to complete this SFT, we'll follow the same process as we discussed in the pre-training stage. First we will prepare data and then we'll ⁓ train it. So let's start with the data preparation.

The data preparation is basically the goal here is that instead of having the random text sampled from internet, we'll have ⁓ data in a particular format because now we want to show to the model examples of prompt and responses or questions and responses. So basically we have to ⁓ curate ⁓ a data set that has this format. For example, it has prompt with some special token.

And then it has the prompt here, give three tips for staying healthy. And then special token for response and the actual response ⁓ with three tips. Now, if we have a lot of examples in this format, then we will use it to continue training the base model on this data. So this data is also called demonstration data because basically we are demonstrating to the model that follow this for learn from this format and follow this format.

Now, there are different data sets and there are different examples of this ⁓ data set, but the main thing is that this data has to be curated manually. In practice, it's it's it's often usual to hire some experts or annotators and then ask them to create these pairs. So one example, for instance, is ⁓ Alpaca dataset, just to get a sense of how the data looks like. I have the link open here, and this is the data.

It's basically very similar to the pre-training data, but now the difference is that it has ⁓ instruction column and output column. So these are things like give three tips for staying healthy and the output. ⁓ and then there are more examples like ⁓ how can we reduce air pollution? ⁓ describe a time when you had to make a difficult decision. So again, it's basically a bunch of questions ⁓ and ⁓ answer ⁓ responses. Now ⁓

Sneha Mehra (02:31:52)  
This is Alpaca. It's available online. It's ⁓ open source. But there are also some other data sets. For example, OpenAI has instruct GPT dataset. So what happened was that they trained GPT 2 and then later GPT 3, but both of these models were only pre trained. So they were not ⁓ doing any post training on those models. And then later they published a paper, this paper, training language models to follow instructions with human feedback.

And what they did was they continue continued training GPT three model and running post-training. So basically they run, they ran SFT and then reinforcement learning on the GPT three. And then the outcome of that model, the outcome of this process was the ⁓ basically the fine-tuned model. And that model they called it, I think, I don't know, GPT three point five or instruct GPT. And also this paper was released on ⁓ in March twenty two.

And then ChatGPT was released and launched a few months later. So ⁓ I believe that ⁓ probably this model was used, and then probably some additional optimizations was happening on top of that to use as their initial model on ChatGPT ⁓ UI. So ⁓ again, this paper, it's an interesting read. And what I want to ⁓ focus on is their data set. So they

Hire experts and then they manually create those pairs of response ⁓ prompts and responses, and then they call it instruct GPT data set. And then they do not release this data set and it's not open source, ⁓ but it has around 14,500 ⁓ examples of such pairs. And again, there are a lot of other examples that are open source now, like Alpaca, ⁓ DOLI, FLAN by Google.

And all of these instruction or demonstration data usually ranges between ⁓ like tens of thousands or maybe hundreds of thousands of examples. So it's a lot smaller than the pre-training data, which is the entire internet. However, the quality is very high. The pre-training data, it's just all the data on the internet, and many of those, even after we clean them, might still be noisy and not too helpful. But this data, since it's ⁓ carefully curated.

Sneha Mehra (02:34:15)  
By experts, it's usually very high quality. ⁓ And ⁓ again, nowadays, if you search Hugging Face for instruction tuning data sets, there are lots of examples. And some of these are for different purposes. For example, this has name No Robots. ⁓ And I'm sure some of these are ⁓ more focused on math and mathematics, math instruct, and so on. So there are lots of examples. ⁓ Some of these can be used to just post-train to just ⁓ run SFT.

And experiment and see how the model learns from these examples. So this is data preparation. Basically, it's going to be thousands of these examples, prompt and response. Now, after this data set is prepared, the next step is running training. Training is ⁓ very simple because it's identical to what we saw in the pre-training. We don't have to change a single line of the algorithm and the ⁓ anything.

So, all we need to do is just replace the pre-training data with this ⁓ SFT data or demonstration data. And then we continue training the base model. And then after ⁓ enough iterations, we would get the SFT model. So everything is identical to the pre-training ⁓ stage. ⁓ And that means that we'll use the same loss function, we'll use the same optimization ⁓ optimization, and then we'll just sample from the demonstration data. ⁓

Things like what is the capital of France, and then we ask the model to predict the next tokens. So this way basically the model would learn from the SFT data to answer questions. So that's training. It's very ⁓ straightforward. We don't have to change a single line. And then after training, the outcome of SFT ⁓ stage is the SFT model. And the SFT model would now answer questions. So if you ask, I want to learn ML, what should I do? it would really answer.

th something like take andrew ang's course on coursera. So we talked about SFT stage. Now this model is not still ready to be deployed. ⁓ And we'll talk in the next lecture why this model is not ready and why we have to r have another stage and that's called reinforcement learning to make this model ready for deployment.

Sneha Mehra (02:36:38)  
Hey. We talked about the SFT stage and in this lecture we'll switch to the second step of post training, which is reinforcement learning.

Sneha Mehra (02:36:53)  
Before we talk about the actual algorithm, let's understand what is the problem. The problem is that after the SFT stage, we get this LLM, and then this LLM is good at answering questions, meaning that if we ask a question like I want to learn ML, what should I do? It's going to output something that basically answers your question. For example, it would say take Andrew Yang's course on Coursera. Now, this response is contextually relevant, it's grammatically correct, and it's answering my question.

But now let's compare this with a different output. For example, something like start with a solid foundation in Python, linear algebra, probability and statistics. Then take a beginner friendly ML course like Android's on Coursera. Practice by building a small project and explore exploring real data sets. Now, both of these answers are ⁓ correct and contextually relevant, but this answer is clearly much, much better than this answer. It is more detailed, it's also more helpful, and it's ⁓ giving me ⁓ some starting points.

The focus of reinforcement learning to go from this kind of output to more detailed, accurate, correct, safe, safer outputs. And that's what we want to happen in reinforcement learning stage. I have another example here to better understand. When we have a prompt like what are some effective ways to reduce my stress? the LLM may have different options to respond to it. For example, response one can be go skydiving for an adrenaline rush.

Second response is exercise regularly and maintain a healthy diet. Third response is shame on you, try meditation. And fourth response can be ignore your problems and hope they go away. Now, again, all these are possible from SFT output because they are actually answering questions and they're not too off in terms of the ⁓ the their context. For example, ⁓

Then it says ignore your problems and hope they go away. It's actually answering to these questions. It's not answering to a different question. But the the response is really not helpful because by ignoring your pro problems, it's not going to help you. Response three is is a mix. It's giving good advice like try meditation, but it's not ⁓ polite. It's ⁓ a ⁓ it's not a safe response. ⁓ And response one can be

Sneha Mehra (02:39:14)  
Not accurate, but still it's responding to the prompt. And this one is the best one. So the goal of reinforcement learning is to somehow adapt the SFT model and get a final model that is more likely to produce responses like this and less responses like these three. So now let's start about learning the technicals.

So, reinforcement learning goal is to generate responses that are preferred by humans. So, which responses are preferred by humans, typically the ones that are correct, accurate, safe, helpful, polite, and so on. So after we complete the ⁓ RL stage, we go from this SFT model to this final model. And this final model is basically what eventually we would use and deploy as a chat.

But how do we do that? How does this reinforcement learning happen? It's basically the training algorithm at a very high level, just to get an intuition, is basically a practicing algorithm. We allow the SFT model to practice. So what practice means is that we give it some inputs and we ask the SFT model to ⁓ produce different ⁓ responses. So each of these arrows here is basically showing one response to this prompt, what is two plus two by SFT model.

So here we have one, two, three, four, five, six, seven responses. And then based on these responses, we say, hey, this response and this response was better. And these ⁓ five responses that are in red are not good. So then the training algorithm goes back to the SFT model, updates its parameters so that next time when we pass a prompt like this, it is it would be be more likely to produce responses that are

green that are more similar to these green responses. So that's the high-level idea. The model practices. It generates responses for the same prompt. And then we somehow ⁓ let we provide, we share with the model that which responses are better, which responses are worse. And then the model tries to reinforce its itself, basically reinforces such that ⁓ it generates more responses similar similar to the ones that we liked. So how does this happen?

Sneha Mehra (02:41:32)  
Basically, the main question is how can we really, once the model generates these ⁓ variations of responses, how do we determine which responses are better than others? To answer this question, there are we can split all the tasks in the world into two ⁓ groups verifia unverifiable. Unver verifiable, sorry, verifiable tasks are the ones that the we can easily verify the response if it's correct or not.

For example, if we think of math problems, and then we have something like this, what is two plus two? It's a math problem. We know that the answer is four. Answer four is correct. And then we can just look at all these responses and look at that their final answer at the very end and see which of those are four. If they are four, they are green. They are they are good. We want s more answers similar to those. And if they are anything but not four, they are wrong. It doesn't matter if the ⁓ logic or

⁓ you know, the reasoning is correct. It's just we don't like it. ⁓ And coding is also another example of verifiable because we can easily verify code if it's we can run it or if it produces the output that we expect or not. But there are also unverifiable tasks. For example, if you think of writing or brainstorming, it's not really easy to identify or say if ⁓ output is ⁓ correct or not. It's very subjective and ⁓ it's also relative. So

for example, for brainstorming, if the input is, you know, help me choose a good name for my startup, there might be various responses, and then some of them are better than others. ⁓ but it's there is no really easy way to verify if they are correct or not. And those are unverifiable. And these two groups of problems are handled differently in the reinforcement learning stage. We'll first start with verifiable. We understand how ⁓ training works in verifiable problems, and then we'll switch to unverifiable.

Problems.

Sneha Mehra (02:43:36)  
So again, verifiable is easier because we can have some logic here, some software, some Python code to look at the responses and then ⁓ check, basically check if their final answer is ⁓ is similar to what we expected or not. So basically all we need is a data set. ⁓ And for ⁓ things like what is two plus two, if we know the answer is four, we can just let the SFT model generate various responses, apply this code or ⁓ logic.

to all these responses and then automatically label them as correct versus incorrect.

And then once we have this, the second step is now the tr the actual training or the updating the model. So during the updating, we ⁓ use some algorithm, some reinforcement learning algorithm. There are various algorithms, ⁓ like PPO, ⁓ more recent GRPO, and so on. And then this RL algorithm is basically given the prompt as well as the which one the labels, which ones are correct, which ones are not. And then the PPO is responsible for updating the SFT model parameters so that it reinforces.

reinforces correct answers. So there are lots of wonderful resources for learning in detail about ⁓ RL and PPO and how they differ. But at the high level, all they do is they're responsible for updating the parameters of the model in a way that reinforces the better answers. And then after we update the model, again, this is the final model. It's it's already a model that is reinforced to produce ⁓ correct answers.

So when we ask a question like what is two plus two, it's more likely that generates something that leads to four. So this is verifiable tasks. ⁓ And ⁓ at the core of it is this PPO algorithm. And then the verification here happens automatically and easily because the response the ⁓ the task is verifiable. Now let's switch to unverifiable tasks.

Sneha Mehra (02:45:36)  
So now in unverifiable tasks, the challenge is we no longer can automatically label these responses. We really don't know ⁓ which ones are better than others. For a prompt like how should I learn ML, it's really not easy to say if they are very close, it's not really easy to say which ones are preferred. So because of that, we need a way to first ⁓ score these responses. We need to score each response.

And then after we have those scores, then we can follow the same process as the verifiable. We can just train a machine learning model ⁓ and basically use reinforcement learning to continue the SFT model based on those scores. So now the next thing that we would learn is how are we going to score these responses? And for that, we have to train a separate model. So ⁓ and this is called RLHF.

Reinforcement learning from human feedback. Basically, we get we use human feedback to train a model that can score these responses. And then once we have that separate model, we'll use it to automatically ⁓ score responses. And then we can use reinforcement learning to update SFT model. So basically, our two steps in RLHF training a reward model using human feedbacks, and then optimizing the model, the SFT model.

With the reinforcement learning algorithm using this train model. So this is a step one: training a reward model. The data preparation part is basically we initially collect some prompts. These are some initial prompts, like things like what is the capital of France, name a famous physicist, and so on. We use the SFT model to generate multiple responses for each prompt.

For example, for the first prompt, which is where is the what is the capital of France? We have response one can be Paris, response two can be it's in Europe, and response three is it's ⁓ Eiffel Tower. Now then we need actual humans. We need to hire annotators to rank these responses. So here basically actual humans look at all these responses and rate rank them. For example, in the first example, Paris is obviously better than it's in Europe because it's it's more accurate.

Sneha Mehra (02:47:57)  
And then it's in Europe is better than Eiffel Tower. So this is the ranking. And then after we have all these rankings from the prompts, we'll can finally generate our training data. And this is how training data will look like. What it would have, each example would have a prompt, which is the prompt here, what is the capital of France. And then we'll have winning response and losing response. For example, based on these rankings, we can say.

Paris is better than it's in Europe. So one example we can form it by ⁓ something like this. Winning response is Paris. Losing response is that ⁓ it's in Europe. And then we can also construct more examples based on these rankings. We can go and say, hey, what is two plus two for winning response? Math is hard, losing response. Just because from the annotators we know that this response is ranked lower than this response.

So this is how we get the training data to train a reward model. Again, in the reward modeling stage, the only purpose is to train a model that can reward answers, that can basically score answers. So after we have this data, the second part is training the reward model. And this is a very ⁓ typical problem in machine learning. Basically, we have a model that takes prompt and response and scores them. So this is the input and output of the reward model.

We have one prompt and one response and the model outputs a score. Now, in order to train this using these training data examples, prompt winning losing response, this is how we would train it. We pass the prompt. We pass the prompt, like what is two plus two and winning response to this model, we get some score. We also pass the prompt and the losing response, we get another score. And then the we apply some ⁓ common losses.

such as margin ranking loss. And what this loss does is basically tries to maximize the difference between these scores. So what it does is it tries to increase the ⁓ score of winning response. And then it tries to minimize or lower the score of losing response so that their difference be maximized.

Sneha Mehra (02:50:13)  
⁓ basically this is the the objective.

Sneha Mehra (02:50:20)  
⁓ so that's it. It it simply uses this ⁓ a loss function like margin ranking loss for this purpose. And then after we ⁓ train this model on this on our training data, on this data, eventually we would have a model that when we pass a prompt and response, it would ⁓ predict the score that and this score would align with humans ⁓ preferences or feedbacks, which we obtained here.

So now this reward model will be a proxy for humans. Instead of humans going and actually ranking ⁓ these responses, the reward model would now score them.

So we talked about the first step of R L H F training a reward model. The second step is almost identical with what we saw in the verifiable tasks here.

It's it's very similar to this. The only difference is that now we don't have an automatically ⁓ verifiable component, which is ⁓ easy to implement. Instead of this, we have now a reward model. So let's go to the optimizing the model with RL stage. Again, very same process. We use some reinforcement learning algorithm to update SFT model's parameters using their scores. And the scores are obtained by the reward model.

So for a prompt like how should I learn ML, we pass to SFT model, we get multiple responses. The reward model scores each response. And these scores are the aim, the hope is that these scores would align with what humans would have scored. ⁓ because the reward model is trained on human preferences. And then the PPU algorithm would ⁓ take these scores and the prompt, and ⁓ it basically updates the parameters of SFT model. So it pref so next time it

Sneha Mehra (02:52:06)  
generates responses with higher score. So it reinforcements reinforces responses like this and this with higher score.

So that's the second stage. And then after we apply this reinforcement learning, we would end up with a model that produces ⁓ answers that are more aligned with humans. And that means that they are more correct, they are more detailed and useful and safe. So for a question with ⁓ like I want to learn machine learning, what should I do? We saw this example. They would output something like a start with solid foundation in Python and so on.

So ⁓ this is post training and ⁓ we talked about both the stages of ⁓ post training. We talked about ⁓ let me go here.

Sneha Mehra (02:52:56)  
So we talked about both SFT stage, which is responsible for just adapting ⁓ the format of the output to answer questions instead of continuing it. And then we apply reinforcement learning so the model practices to produce to produce more accurate, correct responses. And then during the reinforcement learning, there are two types of tasks: verifiable and unverifiable. Verifiable tasks are the ones that we can easily learn them. ⁓

Sneha Mehra (02:53:26)  
That we can easily verify them, such as math problems and coding. And unverifiable are the ones that are difficult to unver to verify. For verifiable, we can just use simply have a component to rate these, score these responses, and then we use reinforcement learning to update parameters. For unverifiable, we would train a new model called reward modeling to score these responses automatically. And then once we have that, we can then apply ⁓ reinforcement learning to

go from the SFT model to the final model.

So that's it. That's all I wanted to cover for post training. And one last thing becau before we wrap up post training is showing this a real example. In the last lecture, we saw that how a base model would simply continue. ⁓ let me rerun again to see if it's going to work now. ⁓ I want to learn ML. ⁓ What should I do?

Sneha Mehra (02:54:46)  
All right. ⁓ So again, the base model, just I I keep asking more questions. What are the resources I should follow and so on. Now what I'm going to do is I'm going to switch to a model that was is not a base model. It's a post train model. So here we can choose ⁓ one of these models. I'll see if I can find a llama three point one.

Sneha Mehra (02:55:10)  
All right. I'll choose Lama three point one four hundred five billion. So it's the exact same model, but not the base model, the post-trained model. And I'll ask the exact same question here. So the UI is a little bit different. I'll ask my question here and then send it. And then I'll keep max tokens to five twelve so it produces more tokens. And then let me send this. All right, now this is the response and

It says learning machine learning is an exciting journey. Here is a step-by-step guide to help you get started. A step one and you know, ⁓ math programming, a step two. It also takes care of the formatting, like ⁓ showing this ⁓ like listed item itemized ⁓ sentences, and some of the words are bold. So basically, it's just answering my question in a very detailed and helpful way. And this is what happens when we apply post-training to a base model. It becomes ⁓ an

a really useful model. So that's it. Now now I guess I can wrap up the post training.

Hey, let's start by reviewing a summary of LLM training stages, and then we'll focus on how to evaluate a trained LLM. So these four ⁓ stages are basically summarizing all the everything we discussed over the last couple of ⁓ lectures. There is the pre-training stage, it's trained on internet data, typically trillions of ⁓ trillions of tokens. They're low quality, large quantity.

And then they require a lot of computes. They have they need thousands of GPUs and month of training. And then it's the language modeling ⁓ algorithm and objective is just next token prediction. So we train them to predict the next token. And then the outcome is the base model. And we saw some examples like Lama base and GPT 2 and 3\. And then the second stage is supervised fine-tuning, which is part of the post-training. In this stage, we we

Sneha Mehra (02:57:08)  
build and curate our demonstration data on ⁓ pairs of prompt responses. They typically range in the in in the in in the range of 10 to thousand ⁓ tens or hundreds of thousands of pairs. And then there are low quantity low quantity, high quality. And this stage requires only ⁓ hundreds of GPUs or even less and days of training. And the ⁓ ML objective is the same. We still have the same model, same training algorithm to predict the next token. And then

The model is ⁓ initialized from the base model, meaning that we continue training the base model, and then the outcome becomes SFT model. And then the next part is creating a reward model if the tasks are not verifiable. And then reward modeling, we need comparison data. So we hire annotators to ⁓ to rank possible responses generated by this SFT model.

And then this stage also requires ⁓ hundreds of JPUs or less and days of training. And then here the goal is to train a reward model that can predict the scores. So the output is just a score. This is the ML objective. And then once we have the reward model, we'll start the final reinforcement learning stage. It's initialized from this SFT model here, and it relies on reward model to score rewards. And then for reinforcement learning, we start with tens or thousands of or hundreds of thousands of prompts.

We use smaller number of GPUs and days of training. And then we ⁓ use a reinforcement learning algorithm to ⁓ so the model practices to generate responses that are reward scored higher by the reward model. And then the outcome of this stage is the final model, which is typically used to deploy and serve ⁓ users as a chatbot. So this is the summary of a stage ⁓ training stages, and after we

complete all these stages, we will end up with the final model. And there are different companies having different models. ⁓ And if you see here again, there are lots of ⁓ models like the ones from Deep Seek and Quinn and there are basically a lot and the Lama and there are a lot more examples, ⁓ small, large, and so on. So ⁓ the question is how should we evaluate them? How can we determine which models are better and which are worse?

Sneha Mehra (02:59:30)  
And this is what we are going to talk about in LLM evaluation. So there are two parts in the evaluation. The first part is offline evaluation, and then the second part is online evaluation. In offline evaluation, we evaluate the model in an offline environment on some evaluation data. And then we measure the performance and then we compare different LLMs. In online evaluation, we have the LLMs already deployed in a production environment, and then we use different ways to compare them.

While they are in production and serving users. So we'll start with offline evaluation, and there are different ways for offline evaluation. We'll start with traditional ⁓ way of evaluating and then task specific and human evaluation. And then for online evaluation, we'll talk about human feedback and crowdsourcing platforms. So the first one is traditional evaluation. What traditional means here is we use some traditional metrics ⁓ to evaluate LLMs.

And one common metric to evaluate LLMs and in general a text generation model ⁓ is to perplexity. And it's basically a metric that measure measures how accurately the model predicts an exact sequence of tokens present in the text data, often in the evaluation data. So there is a formula, and then ⁓ you can ⁓ look on Wikipedia to ⁓ to see the details of the formula, but at the very high level and intuitively.

What this formula is doing is we'll have some ⁓ sequence of data, let's say our evaluation data, for example, how are you doing? And then we use the model to see how likely is it to produce this exact how are you doing sequence. So we start with how, we get the probability. Then we pass how are, we get the probability of u. And then ⁓ we have basically we in practice, like theoretically, we have to multiply them, or we can basically

Take the log of these values and then sum them. And then this would give us a number representing how likely the model is to produce this exact sequence of how are you doing. And this way basically we measure how good the model remembers evaluation data or can exactly reproduce evaluation data. And then this way we can compare different models and the model or the LLM that is better at producing or reproducing the evaluation data is better.

Sneha Mehra (03:01:54)  
However, this method, this traditional way of evaluating LLMs, is no longer ⁓ too helpful or meaningful because this metric is not really ⁓ telling us much. What it says is that the model can reproduce a certain sequence of tokens. But that's not what humans want really in practice. What we want is a model that can really answer our question, or it's correct, or it's useful. So that brings us to the second way of evaluating LLMs, which is

Task-specific evaluations. So the purpose of this kind of evaluation is to assess the performance of LLM across diverse tasks that we really care about. For example, mathematics, code generation, common sense reasoning, world knowledge, and so on. So these are some of the key areas that we want to evaluate the model and see how the model performs. So we'll just briefly review some of these. ⁓ And they are great benchmarks, meaning that they are

Benchmarks are basically just a ⁓ a data set of with certain examples in that particular domain. And then we can use those and run them on ⁓ pass the prompts to the LLM and see if the LLM response is similar to the correct answer. For example, common sense is just common sense questions if the model asks answer ⁓ can answer those questions or not. ⁓ this is one real example. ⁓

the pro the trophy doesn't fit in the brown suitcase because it's too large. What is too large? The trophy ⁓ A, B, the suitcase. This is the prompt. This goes into the LLM. And then the correct answer from this data set, from this example, is the trophy. So we pass this to the LLM and we see if the LLM really ⁓ generates the trophy or it generates the suitcase or or anything else, like any random outputs.

And then this way we can compare different LLMs and see how well they are doing in ⁓ let's say, common sense reasoning. And there are a bunch of benchmarks for each of these domains. We have ⁓ listed three benchmarks here. And then for word knowledge, for example, who wrote this, who wrote that, and this is the correct answer, and these are the benchmarks, and we can use LLMs and evaluate them under word knowledge ⁓ capability.

Sneha Mehra (03:04:09)  
And then we have mathematical reasoning. For example, if a train travels 60 miles per hour for three hours, how far does it travel? And then we see if the model can really answer ⁓ correctly or not. And then we have also code generation, something like write a Python function to ⁓ check if a number is prime. And then this is some possible correct answer. And then the model also generates an answer, and then we can check and see if the outputs are correct or not.

And these are some benchmarks for code generation. So this way of evaluating LLMs is very common because we really care about ⁓ these numbers and we want to know how good different models are compared in different domains. So this is a common way of ⁓ evaluating LLMs. And then the last part which we had here is human evaluation. Basically, in this type of evaluation, we just ask higher experts and ⁓ expert humans.

And then we ask them to ask challenging questions from the LLM and see if their answers are ⁓ correct or not. So this can be tricky ⁓ in in one sense. It's good because humans are ⁓ more capable of evaluating and ⁓ understanding the responses and verifying them. But also it can be tricky because humans can be biased, and a lot of times depending on who is evaluating an LLM in which domain, the responses or their evaluations might be slightly biased.

And also subjective. But this is also another way to evaluate LLMs. And then we'll switch to the online evaluation. In online evaluation, we have two ways of doing it. One is human feedback and another is crowdsourcing platforms. So human feedback is simply ⁓ just ⁓ when the model in the UI, in let's say ChatGPT UI, it responds, it asks the user to rate it with a thumbs up or thumbs down.

And this way, the ⁓ the team in behind Chat GPT can get the feedback from real users, whether they're happy with the answers or not. And this would give them a signal whether how good the model is doing compared to previous models. ⁓ and this signal can be also used for other purposes like for further fine-tuning or doing reinforcement learning, because these feedbacks are coming from real humans, so they can be valuable. But they can also they serve as a evaluation ⁓ purpose as well.

Sneha Mehra (03:06:36)  
And then the final part of online evaluation is crowdsourcing. So there are different websites that they use crowdsourcing to rank ⁓ and rate LLMs. And one common example is ⁓ this LMRNA. And it's initially developed by ⁓ Baricley graduates. And if I search Google, it's it says that LMRNA is a public web based platform that evaluates.

LLMs through anonymous crowdsourced pairwise comparisons. Users interprompts for two anonymous models to respond to and then vote on the model that gave the better response, in which the model's identities are rebuilt. So basically it just asks the ⁓ real humans. And then the mo ⁓ the users can vote. And then it uses those votings to rank LLMs. And then this is their website.

So there are different tabs here. I selected text and text means basically the models for like LLMs, the models that are ⁓ ranked. ⁓ basically the models that are ranked for their text generation capability. And this is the ranking as of today. For example, the f the first rank is Gemini two point five pro developed by Google. And these are the number of votes and this is the score so we can see the gap between ⁓ different models.

And then this is the license. Some of these are open source, some are closed. ⁓ the second rank is O3, and then we have Chat GPT 4.04.5, ⁓ Cloud Model, Cloud Four Opus 4, ⁓ and so on. And this is DeepSeek, which has this MIT license, which basically is open source and we can use it. And it's also open weight, so the weights are available to be used. And you can see there are lots of different models and they ranked, and it these this table keeps changing.

⁓ And these companies keep releasing new models and just ⁓ usually newer models are better so they ranked higher. And again it keeps changing and changing. So I think it has like lots of L LMs now.

Sneha Mehra (03:08:49)  
Yeah, 211 LLMs are ⁓ are ⁓ included in this table and then llama thirteen billion is two hundred nine. ⁓ And these are relatively older models, and that's why they are ⁓ lower on this list. So yeah, so that's it. That's the ⁓ crowdsourcing platform. So these are all different ways that we can evaluate LLMs.

And they're used in practice to see if a newly trained L L is better than the previous ones or not.

Hello again. This is the finally the last lecture of week one ⁓ of LLM foundations. So we learned everything about ⁓ the foundations of LLMs, how the data preparation looks like, how the model architecture is, model training, both the stages pre-training and post-training, and also evaluation. One last topic that we want to cover in this lecture is the overall system design and how the what is the holistic view of ⁓ like chatbot systems.

So, as you can see here, trained LLM, this is basically the post-trained model that we finished training, is here. And then, ⁓ you know, it's just very small part of the entire system. And then there are a lot of other components that are necessary and ⁓ used in practice to power a service like Chat GPT. And this is simplified visualization. In practice, there might be even more components handling different things. ⁓ but this would give a

a good overview of what are different things that are important and needs to happen aside from the actual trained LLM. So let's just walk over them one by one and understand what are their purposes. So when we have the text prompt coming from the user, the first component usually we have is ⁓ some guard like guardrails in for input guardrails. And we can also call them safety filtering.

Sneha Mehra (03:10:46)  
So all this comp ⁓ all this component does is it ensures that ⁓ our text prompt, the text prompt that was provided by the user is a safe prompt, meaning that it's a same safe request or safe question. For example, if the user requests something that is ⁓ violent or could be harmful, at this step, this component would identify it and just ⁓ not answering it. So basically, if the result of the safety filtering is

That it's not safe, if it determines it's not safe, then it would use some other smaller models to generate some rejection response. So this generated response would be shown to the user. ⁓ And this is something like, hey, sorry, we cannot ⁓ assist you with that, or things like that that we've seen on various platforms. Now, if the it this component identifies that the text prompt is safe to process an answer, the second part is usually a prompt enhancer.

And what this component does, it's it can it most of the times it's a machine learning based model. ⁓ and what it does is it improves the text. So a lot of times in the input text, there is ambiguity. There can be some misspellings or grammar issues and all those kind of ⁓ different pro problems in the initial text prompt. So this component is responsible to fix all of them. It's responsible to make sure punctuation punctuations are well and it's they are there in the prompt.

There is no misspelling, there is no typos. If there are typos, it's they get fixed. And also if their ambiguity or the grammar is not correct, they all get fixed. And it can be a combination of heuristics as well as machine learning models. And then after the prompt enhancer, we would have a prompt that actually is correct and ⁓ is understandable. And it's not also vague. Now, this enhanced prompt goes into this response generator.

This response generator is just the text generation decoding algorithm that we've gone through. And it ⁓ interacts with the trained LLM and generates the tokens one by one. ⁓ And ⁓ it could use ⁓ some ⁓ like toppy sampling method. We'll talk about session management in a bit, but ⁓ having that aside, after text response generator completes its process, there would be some response to the prompt.

Sneha Mehra (03:13:11)  
And then that response goes into this response safety evaluator, ⁓ or sometimes it's called output guardrails. So what it does, it again ensures that the generated output is also safe, meaning that it's not really ⁓ sharing something dangerous or harmful or biased.

And then if the result of this is that it's not safe, it would go again to the rejection response generator and generate a response that hey, somehow we couldn't process your request or something like that. And if it determines that it's safe, it would just show the generated response to the user. So every time a user enters something like this, it would go through all these processes and components, and then the user would see the generated response. Now if you

⁓ if you see in ⁓ chatbot services like ChatGPT, when you ask questions, you get responses. You can ask follow-up questions. ⁓ and then the model would remember all the previous conversation and interactions. For example, you can ask something like, Hey, help me come up with a name for my startup, and then the model would share something, and then you would say, Hey, these are not good, ⁓ make it more formal or informal or something like that. And the model would

remember everything from before and then they adapt the model would adapt its response and j show you another response. And this is handled by ⁓ by a session management. And all it does is basically it it just keeps track of all the ⁓ previous messages that were interacted between the user and the LLM. And it would append us into ⁓ and it's called chat history. It would append us to the new prompt each time.

So each time you enter a text prompt and it goes here, right before that, it just appends here all the previous ⁓ chats in that particular session. And all of this goes into the response generator. So when the response generator interacts with the LLM using top P sampling, the LLM takes into account the entire chat history when generating the new response. So this is how the follow ups and chat and ⁓

Sneha Mehra (03:15:25)  
follow-up questions are handled using a session management. So that's it. That's ⁓ about overall system design. Again, in practice, there are a lot of other components for different reasons. And again, it's something there is not a single solution. Different companies use different components for different purposes. But this can be ⁓ a rough, you know, it gives an overall intuition of how a system design of a chatbot can look like. So with that we can wrap up

Week one to ⁓ for L L and foundations we've learned everything we wanted to cover.

Sneha Mehra (03:16:03)  
And let's let me find this. So basically we learned about LLMs, we saw pre-training, data c ⁓ data collection, model architecture, training, text generation. And then we learned about post-training, in particular SFT and ⁓ reinforcement learning, as well as reinforcement learning from human feedback for unverifiable tasks. And then we discussed evaluation and ⁓ system design. So with that, now we built a strong foundation.

And now we can switch to more advanced use cases and extensions of LLM that powers some of these applications.

—----------------------------------------  
1

Sneha Mehra (00:00:00)  
Hello, everyone. I just want to welcome. Thanks for ⁓ joining this cohort. It's it's amazing to have all of you. ⁓ It's ⁓ it's so exciting. I was trying to read every single introduction in the intro channel. And then I realized that we have lots of people from coming from different backgrounds and very diverse expertise. So I'm ⁓ sure there would be a great learning opportunity for all of us to learn from each other. And

We build long-lasting relationships. So ⁓ as we officially start the cohort, I want to welcome everyone. Thanks for joining this cohort. ⁓ And the purpose of this talk is mostly ⁓ on logistics and how the course works, what are ⁓ what you should expect, some of the t ⁓ tips for success in the short term and long term. And then ⁓ at the same time, we would also demo ⁓ the course portal and we'll see some of the materials that they are released.

So ⁓ that's it. ⁓ And also if you have any questions, feel free to post in the chat. We'll we'll try to text questions after each section of this ⁓ talk and then ⁓ we'll answer questions. So ⁓ let's start. Who are we? My name is Ali. I will be the instructor for this course. ⁓ I can share a little bit background about my myself and my career.

I was ⁓ I had a startup, Confo, it was acquired. Then I joined Google ⁓ in the Google Maps team. There I was trying to ⁓ build, use machine learning and big build technologies, ML-based ⁓ features to improve the quality of the Google Maps. And after Google, I joined Adobe. And I've been at Adobe for the last ⁓ I think six, seven years. ⁓ And ⁓ for the first few years, I've been working on building search and recommendation systems.

And then over the last couple of years, I've been part of the Firefly team. I've been a core tech lead ⁓ in ⁓ Firefly team. And as you may know, Firefly and Adobe is basically trying to use generative AI technologies and bring those into Adobe's products. So everything around like text-to-image, ⁓ text-to-video and their applications, like image editing, video editing, and those kind of features. ⁓ that's my ⁓ career ⁓ and

Sneha Mehra (00:02:22)  
We also wrote two books with my wonderful co-authors. We have Machine Learning System Design Interview Book with Alex from ByteByteGo. We also ⁓ launched Generative AI system design interview with my co-author Howe from currently working at OpenAI. And we've also received a lot of great feedback ⁓ for the books. ⁓ Aside from that, I also teach at Stanford in their extension programs. I ⁓ I do

Project deep dives and ⁓ you know, similar to this format ⁓ of our course ⁓ deep dives, ⁓ question answering, and all those things in two different courses in ⁓ CS224 and CS221. One is machine learning with graphs, and the other is ⁓ AI principles. So ⁓ that's mostly about this. And we created this course together with ByteByteGo and Alex, ⁓ hoping to bridge the gap in the AI and make it ⁓ more

⁓ available and easier for people to learn AI.

Sneha Mehra (00:03:28)  
And this is our ⁓ agenda for today. We are going to focus on four different topics. The first topic is our approach to AI engineering. In our approach to AI engineering is basically why we created this course. What are some of the values ⁓ and principles that we had in mind and we created this course and why the course we think would be valuable? Then we would talk about course logistics, like some of the things that you should expect ⁓ throughout the course during the week.

And you know how to get support, things like that. Then we'll briefly go over the six weeks journey and some of the topics that we would cover and learn, and also the projects that we would build. As part of this, we would also ⁓ show a demo of the course portal so we can go over those as well. And finally, I would share some tips for success. And that includes both short-term success and ⁓ long-term success, as well as some of our commitments, ⁓ some of the required commitments and

Course policies.

So let's start the our approach to AI engineering. I've received a lot of messages from people over the years, and they were trying to get into the AI from different backgrounds, or they're software engineers, they're hoping to ⁓ switch to AI. Or even they have they are an AI engineer, but they want to ⁓ improve their skills and and learn different features, how to build them. So ⁓ when we created this course, we

And when we heard about a lot of those requests from people, we learned that there are ⁓ there is a gap in learning. There are a couple of things that ⁓ are not easily available outside in external resources, external ⁓ courses. ⁓ One aspect was learn doing, which I'm going to talk about this in the next slide. The other ⁓ value was ⁓ learn the right depth and finally learn ⁓ in a community.

Sneha Mehra (00:05:30)  
So ⁓ the first one is learn by doing. A lot of people are doing passive learning. And again, I've been receiving messages that hey, we are watching tutorials, we are reading papers, but we are not ⁓ really feeling confident or comfortable to build AI systems. We don't know how they are really built in practice. So ⁓ we realize that passive learning is not ⁓ the approach for building AI systems. And perhaps it needs to be followed by an active learning. And by active learning, what we mean is that we

Try to build AI systems. So it comes after learning. So first we still need to learn by watching tutorials and reading papers or whatever. But then after that, we have to try to build something. And during this building process, we would ⁓ we would basically figure out what are different pieces, different components, ⁓ how they can be built and how they can be put together to ⁓ create some AI feature. And during that process, there would be a lot of ⁓

Implementations and ⁓ errors and then iterations to fix those. So ⁓ we think that the real understanding would happen during the active learning, not only passive learning. So when we created this course, we designed it such that it's focused on ⁓ building some ⁓ real prototypes that are popular out there. For example, features like deep research or features like ⁓ perplexity. So

⁓ that was our first principle where when we designed this course, learn by doing.

Sneha Mehra (00:07:05)  
The next principle that we had is to learn the right depth.

As you know, there are lots of resources and courses out there. Again, one of the complaints that I was hearing a lot from people was that they find it very difficult to find the ⁓ right course to learn the right material. Either they're one of most of the ⁓ courses outside are one of the two extremes. Are they either they are math heavy ⁓ or very shallow? Math-heavy ones are the ones that are that have lots of details and you know too much theories.

And many of those theories or details are not always practical. I mean, they are really good content ⁓ if somebody is a researcher or they want to ⁓ continue research or in are in academia, but they're not necessarily useful for building an AI tool. They're just simply too much details or too much theories. And again, this includes some ⁓ good courses from credible institutions. ⁓ like I've been going through a course.

⁓ available on YouTube ⁓ from a very good school. And it was also very, very good. I learned a lot. But the problem was that many of the things that it was shared throughout the course, the course was like 40 hours or something like that. A lot of those were not really used in practice. It was not even practical to deploy at data scale or ⁓ support those or use those ideas. So this is one extreme. And then the other extreme is basically courses that are too shallow or not not only courses, anything like

YouTube videos, any any content that is too shallow. And they are typically missing a lot of important details, and they are more for making you an AI user, not ⁓ really an AI builder. So basically, after you go through these shallow ⁓ content, you would ⁓ know, for instance, how to use an API and call it and get a response from an LLM, but you would not know what are the internals of LLMs, how they work, how can you pick some of those and mix them and build something new or

Sneha Mehra (00:09:07)  
customize it to make it a ⁓ like chatbot for a certain use case. So we aim here. We we try to, when we were designing this course and creating the content, we were aiming to sit in between these two extremes. We started from ⁓ creating something very comprehensive, anything that can be important ⁓ and necessary for an AI engineer to know and build different features. And then we try to remove unnecessary details, anything that is not going to

Directly help to build something or they are not really used in practice. And then we reached ⁓ somewhere here. And we think that this is going to be ⁓ very important and very valuable because it allows people to ⁓ do not overspend their time on ⁓ learning things that are not really matter. So that was the second principle that we had in mind when we ⁓ designed this course.

Sneha Mehra (00:10:05)  
And then the finally we had the last value in mind, and was and that was learning in a community. And I think that this is probably the most important value ⁓ compared to other two. So let me talk about that briefly. ⁓ Learning alone is generally very difficult. It can be slow, ⁓ it ⁓ is ⁓ usually discouraging, and also it's easy to get stuck. Now these days,

AI tools or chatbots, you may be able to unblock yourself using AI if you get stuck. But still there are a lot of situations where ⁓ we really need ⁓ feedback or inputs from real humans. ⁓ For example, when we are ⁓ when we try to build a startup, we are ⁓ thinking about different ideas, or we have a system, we want to get feedback, those are all the situations where if we are learning alone or building alone, we may easily get stuck.

So we focused on learning with peers in a community. So it has lots of benefits. It has we can collaborate with with each other on projects. We can build ⁓ we can learn from each other during the collaboration. We can give and receive constructive feedback when we are in a community. So ⁓ it's it's great. I mean, our community is now over, I think, 400 people. So

I'm sure there would be a lot of good things that we can learn from each other. And part of that would be giving and receiving feedback. So this is another advantage of ⁓ this journey where we will be learning with in a community. And finally, we would build ⁓ lasting connections. ⁓ So I'm sure again, many of the connections that would be formed in this journey, it would be ⁓ it would last beyond this course. ⁓

Hopefully there would be a lot of great things coming out of the ⁓ the forming connections, great ideas, great products, great services.

Sneha Mehra (00:12:08)  
So that's the first topic. I I'll pause here to take some questions from the chat. ⁓ please write if you have questions. Otherwise I will continue with the course logistics.

Sneha Mehra (00:12:23)  
Yeah, Ali, ⁓ one question. ⁓ Do we get the content and all those details through our community link itself, or is it a separate place where we get the open? Yes, everything would be would be available in the course platform, ⁓ including the material that we would release, the Git link, it would be there. Also, ⁓ the the live sessions would be recorded and we would provide the recorded links in the course platform as well.

Thank you. Sure.

Two questions, ⁓ maybe more regarding the logistics side of things. ⁓ one was like, so you said that we found like the sweet middle spot, but if let's say ⁓ up in terms of learning, but let's say if someone wants to delve a little deeper on the theoretical aspect, do we get link resources using the course? ⁓ Mm-hmm. That's a that's a great question. ⁓ yes, yes. So as part of the material, we are trying to just ⁓ strike this balance, right? And not sharing oversharing, you know, unnecessary details. But

Whenever there are there is good opportunity for learning, if people are interested, we would provide links to external external resources or technical papers. So you you would find a lot of places where we are sharing just a bunch of links. Even week one is now released, so you can see there are a bunch of links put put out there. And as you watch the the material also, you would see you know links. ⁓ so you know you know which link to go and read if you are interested to learn deeper or in a greater detail about a certain topic.

⁓ Makes sense. ⁓ and then you said that there's gonna be, you know, like the one one thing that we get from the big community is giving and receiving feedbacks. So what's like how does that work and what's the logistics of that? Sure. I'm going to briefly talk about how to how to get support and in the course logistics. But basically the general idea is that we have those channels and general discussion is a great place to just do anything. I mean, just post your questions, brain a storm. ⁓ and we really encourage everyone to also

Sneha Mehra (00:14:28)  
share helpful resources if they have anything. Like if you come up, if you watch a video and you think it could be, ⁓ you know, beneficial for the community, please share it. So the hope is that we learn a lot and we share a lot in in those channels. And same goes for the getting feedback. Awesome. Thanks. And aside from that, aside from that we have also office hours. Again, I'm going to talk about that shortly in this in this talk. And that there basically you can get ⁓ again feedback from me and other people.

Sneha Mehra (00:14:58)  
I have a question about office hours. Do we need to prepare and announce in advance the subjects that we want to discuss, or we can just come there? So ⁓ we try to keep this very flexible. ⁓ so we don't have a lot of like as strict rules. If you post your questions beforehand, we can I can definitely go through them, prepare them, prepare if I need preparation. ⁓ and then answer. But you can always, you know, you are always welcome to ask your questions ⁓ live.

⁓ I try to answer most of them. ⁓ If they ⁓ are relevant to the course and the material, I would definitely be able to answer. If they are outside, I try my best to answer. But again, I if I don't have expertise, I would just come back to you. But if you have something that may require me preparation or or do some researching, please post them ⁓ ahead of time and give me some time to prepare for that. That sounds great. ⁓ And posting

I assume that there is ⁓ a capability in the portal that we can use for posting, right? There's can sorry, can you repeat again? I mean the the the question was if we have questions, where do we post them? I I assume that in the portal there is a place where we can do that. Yes, yes, yes. I'm going to demo it soon. Great. Thank you so much. Of course.

⁓ hey Ali. ⁓ I had ⁓ one question. So we are looking into the middle ground between math heavy and shallow ⁓ learning of ⁓ AI. So will we also look into inference ⁓ and deployment strategies for the AI models? Or would it be just something that we will learn at a very upper level and it is our responsibility to actually implement those kind of techniques? Okay, so that's also a great question. ⁓

So we would definitely talk about the inference time. I mean, that's part of the AI, ⁓ both in the material, you know, during the learning process, and also as we build during the project deep dives. So, because it's important, I mean, we it's it's part of the AI system to know how to inference, either it's an image generation, video generation, diffusion model, or it's LLM, or it's a reasoning model. So that we would go over. Also, we would go over.

Sneha Mehra (00:17:20)  
Inference optimizations whenever it's relevant. For example, in image and video generation, it's it's it's important to optimize ⁓ at inference time. Otherwise, it's going to be it's not going to be practical. I mean, in practice, there are lots of optimizations that happen on top of ⁓ text-to-video models at inference time. So those things we would go over ⁓ in the material. For the deployment, we are not ⁓ focused on the deployment because it's it's it's basically agnostic to AI.

And you know, you know, you know, just the same deployment that it applies to the ⁓ software engineering. Mm in most of the cases also applies to AI. So deployment is not our main focus in this course. Okay. So we'll not be dealing with infrastructure at all. We'll be dealing more with the more or less the system and optimization side. Exactly. I mean the whole focus would be on on the systems, like how to build

some of those systems, what are different pieces, how those pieces can be built and how to put together all those pieces to ⁓ create this entire system, like a minimal, ⁓ as well as the ⁓ some information and the learning objectives of the video. And also the video is is coming with a with a canvas like. So if you want to also access the canvas, it's it's here. ⁓ And all the links that ⁓ is mentioned in the video. So for any of the links it's just placed here for for the convenience.

Sneha Mehra (00:18:46)  
And project one is also if when you click, you would see the the overview and learning objectives and the time commitment instructions and ev everything. And then also the link to the Git triple. And if you just click on it, you would see the the actual project.

Sneha Mehra (00:19:06)  
So that's about the course portal. ⁓ And ⁓ I I assume you've already completed this, but this is where you start and ⁓ make sure you have access to everything and the community line guidelines. And if you need help, you can also ⁓ email us at any time. So that's about the course portal. Let me reshare the slides.

I have a quick question if if I can. Sure. Sure. Okay. So ⁓ I'm really good at theoretical concepts, but never had a hands-on experience on this coding. So do you ⁓ like you you showed me you know some coding. So ⁓ do you provide step by step how to ⁓ set up whatever the interface that I need to work with?

Yes, yes. So basically for each project there is a clear instructions on how to set up or some of the projects are on Google Colab. So you don't need even any requirements. So you can just run it. ⁓ And the project itself is a notebook with different cells. And each cell is basically providing an explanation of what we are trying to achieve and what are the steps. And here is the code. And then some parts is obviously your code here section, which you are expected to work on. ⁓ But then ⁓

After that, we would during the deep dive, yes, we would go over every ⁓ single cell and complete it and provide explanation. Okay. So part of this course, will I be able to ⁓ okay, once I complete the course, will I be able to ⁓ build production grade LLM or agents, including like scalability, observation, observability, those kind of things, including those c aspects.

⁓ again, so the the we are not going to focus on deployment and observations, but part of the material talks about evaluation mechanisms. Understood. Okay, so you'll be covering like best practices for deploying.

Sneha Mehra (00:21:08)  
We we are going to cover optimizations and you know necessary for deployment. Understood. Okay, thank you. All right. So ⁓ finally, the last section is tips for success. So here I'm going to focus on ⁓ commitments and course policies. And finally, some of the ⁓ some of the ways you can make the most out of the course to succeed in the short term and also build for the long term.

The time commitment is four to six hours per week, and and this is this includes watching the pre-recorded guided learnings, ⁓ spending some time on the project template to complete it, ⁓ and attending the live sessions or watching the pre recorded live sessions. The attendance is optional in all our sessions. ⁓ all live sessions will be recorded and they will be made available.

And then deliverables are for projects, ⁓ there is no submission process, so no delivery is required. It's just for you to practice and reinforcement the learnings. And then ⁓ we would do the deep dive. And they're also optional. So if you don't get a chance to work on the projects, we would do the deep dive anyway. And capstone, as I mentioned, is optional. ⁓ so these are the commitments. The course policies is ⁓ we all come from different ⁓ backgrounds and different.

levels of familiarity with AI. So the goal is to just be all respect each other and help each other and lift each other up to learn more and ⁓ feel safe in asking questions, even simple questions. The second thing is AI use is permitted for the capstone. For the projects, weekly projects, it's not, I mean you can still use AI, but it's not really going to if you ask AI about the

Projects to complete those portions of the code that you are expected to work on. It's very likely that AI chatbots would be able to complete them correctly. But I would discourage that because again, those portions are designed so that you learn how to implement things if you are interested in the implementation part. So I discourage using AI for projects, weekly projects, but feel free to use AI in any form for your capstone project. I also encourage everyone to ask questions in public chat channels.

Sneha Mehra (00:23:30)  
One of the main benefits of this course is to ⁓ you we learn from each other and and engage and answer to each other's questions. So asking in public channels can be beneficial for others because other people may have the same question and it helps them to ⁓ also learn more from people's questions.

Sneha Mehra (00:23:53)  
And for succeeding in the short term, again, feel free to ask questions, even if they are simple. Just post them. Me ⁓ or ⁓ someone would answer your question. And it's going to be helpful. And you can unblock yourself as as as soon as you have a question or you feel blocked on something about the course material. Also, please engage and help and answer other people's questions, your peers, provide feedback. It would be ⁓ really helpful for everyone.

Third is the step-by-step ⁓ learning process. The course is designed sequential and it's expected to each b each week is building on top of the previous weeks. ⁓ And I think ⁓ the the main thing of this point is not to just discourage and ⁓ continue learning each step and completing the learnings and material and then switching to the next. ⁓ The hope is that by the end of the five, six weeks, ⁓ you should.

know how some of the ⁓ very popular AI features out there are built by just following this a step-by-step process. And also please use the resources available. There is a great community already formed. So ask questions, ⁓ make connections, get feedback from me and others. And also collaborate. ⁓

collaborate on some projects or some assignments or even the capstone project.

Sneha Mehra (00:25:34)  
And build for the long term. So there are these three key ⁓ things that I want to share, which I think would help a lot with ⁓ the long-term success. The first one is building connections and connecting with peers because it's gonna last beyond the course. And hopefully there would be great ⁓ ideas or startups or services ⁓ at some point coming out of these connections. The next thing is persistent. ⁓

Going to be ⁓ any learning journey can be difficult. ⁓ And you know, sometimes there are failures or ⁓ you know you may get stuck, but just ⁓ use the resources available to you, unblock yourself and and ⁓ and just progress comes from persistence. So ⁓ just iterate and practice and learn from failures and ask questions. So it's gonna pay off in the long run. And finally, I would

also suggest to think big and treat this journey as more than just a course. ⁓ it can be a launchpad for bigger ideas. The capstone project doesn't have to be something ⁓ related only to this course. It can be something that you've always wanted to work on or build. ⁓ And ⁓ just just you know think big.

And by the end of this cohort, ⁓ these are some of the achievements. We would complete five AI prototypes plus the capstone project, which you can showcase or extend after the course. You would learn ⁓ and have prototyping the skills because a lot of projects are basically designed such that ⁓ we use whatever is out there, all the tools, libraries, open source models, so we can quickly move from ideas.

To actual prototype. ⁓ The third one is confidence. As we are prototyping these AI systems, we would build confidence because now we know how some of these systems are gonna work. ⁓ This is very important because a lot of times I see people that are going through passive learning and just learn tutorials. They don't feel confident to talk about how real AI systems are built. But as they as we get into the internals and we try to build them, we ⁓

Sneha Mehra (00:27:53)  
I'm sure we are going to gain a lot of confidence ⁓ to know how the systems work and prototype them. And finally, you will have ⁓ access to the community of peers, mentors, and collabor collaborators to stay connected stay connected beyond the cohort.

Sneha Mehra (00:28:15)  
And I think this is my last slide. So let's build together. ⁓ please support each other. We it would be really helpful if we all help each other, share good resources, and ⁓ lift each other up. Please also give us feedback. ⁓ Our our goal was to create something that is more effective for people to learn AI or interested to become AI engineers. We try to figure out some of the existing issues.

or challenges and fix those. But I'm sure ⁓ as this is our first experience, there would be lots of ⁓ places where we can improve. So please give us feedback and let us just ⁓ improve ourselves as as we move forward. And finally, please ⁓ feel free to share your journey, your learnings with others, other people on your friends or social media if you think some learning can be helpful also for others.

Sneha Mehra (00:29:17)  
And quick ⁓ three questions are listed here, but if you have more questions, you are more than welcome to post them and on channels. So the three questions here is do I need to be a pro at coding? Again, the answer is no, because the main focus of all the projects is to understand the workflows and what are the different components and how to connect them. In almost all the projects, we try to just use what is out there, all the tools and external libraries to just build something. So ⁓ and

The main purpose, the main learning purpose is to ⁓ just understand what are the pieces, not how those pieces should be implemented or low-level details. So no ⁓ I I I I believe there is no requirement to be pro at coding unless you are interested in implementing some of those or doing some low-level coding. What if I fall behind? You can watch the recordings and also ask questions. And finally, how do I get help?

we have office hours and also we have chat channels in the course platform. ⁓ feel free to post any questions there and we'll try to answer as soon as possible. ⁓ so that's my last slide.

I think I can take ⁓ a few more questions.

Yeah, Ali, hi ⁓ and hi everyone. I would like to ask something. So, like ⁓ you know, you said when we will do deep dive of ⁓ previous session in the next ⁓ session, right? ⁓ yeah, so and we will go over the project ⁓ at that time. ⁓ so is it is it

Sneha Mehra (00:31:00)  
Will it be more like a pair programming session? You know, that we we we can ⁓ do ⁓ we will have time to ⁓ do same or same what like fix our ⁓ what we missed or improvements or we could not do ⁓ fix it along with you, like ⁓ or ⁓ or you.

you know, whatever like you have done like during this deep dive session, you you did everything you would do something and will you be ⁓ sharing with us in some way, like through GitHub? And then we can we have time to incorporate it or it's ⁓ like or we just take notes or go by recording. So anything you would like to share there? Sure. So basically the beginning of the week, a project is released and it's a project template.

So you get one whole week to go over it, ⁓ work on it, ⁓ and also ask questions from me ⁓ if you have any confusions or concerns as you try to implement those. And then the beginning of next week, we would do a project deep dive and walkthrough. So the format is I'm going to ⁓ go over the project a step by step, explain each line.

And also for the lines that were previously you were expected to implement, we would write and implement those sections. And we would run it and make sure it's gonna work. ⁓ and ⁓ so that's basically gonna be the format.

Sneha Mehra (00:32:35)  
So Ali just Ali, I have just three quick questions. I've been like waiting for a long time. ⁓ maybe some of this already answered. One is like, I mean, are you going to cover some kind of fine tuning on existing LLM model? So we would cover extensively training and fine tuning in the learning material. In the building, because of the infrastructure and and cost related ⁓ like cost relation related ⁓ related costs, ⁓ we just use open open source model.

Okay, so regarding open source models, like which are the models that you we will be using for this cohort? So in the in the first few projects, we would start from like older, simpler ones for the learning purposes. We would, for example, in the project one, we would start with GPT-2 just to inspect these layers. ⁓ but later down the road, we would switch to also some variants of Lama. ⁓ And ⁓ I believe we would also use ⁓ a few other lightweight models. But the idea is again as far ⁓

Instructions are provided in the in the project. You can always switch to different models and experiment with different models. So as long as the model is lightweight enough to to be loaded. ⁓ One of the ⁓ libraries that we are using is Olama. And then there basically it's it's there are lots of documentations on which models you can use and load ⁓ to experiment with. ⁓

The learning? So we'll we'll go ⁓ the we'll we'll go over Langchain in in the in one of the projects. ⁓ a a A to A we are not going to go over it just because those systems can be very complicated. ⁓ if we want to set up the ⁓ A to A and multiple having multiple agents. But ⁓ Langchain definitely yes. Yeah, just the last question. Thank you so much. It's like after this course, are you planning

Any like the next course, like you know, more advanced or something like that. I mean, it's too early to ask it, but just in case. Yes, it it's it's probably too early. We are still learning, but one thing is we plan to do is as because this field is evolving fast and ⁓ we'll try to keep ⁓ all the materials up up to date and update them whenever necessary, and we'll make sure to also provide those as well, ⁓ even after the cohort. So

Sneha Mehra (00:35:00)  
Whenever we update anything or add add new materials, we make sure to also provide them. Yeah, thanks, Ali. Really excited. Of course. So ⁓ I'm also going to answer some questions from chat. ⁓

Sneha Mehra (00:35:24)  
one question. ⁓ There are raised hands for a long while. ⁓ So maybe raised hands like have also. Of course, of course. Sure. Let me ⁓ let me just go over a few questions in the chat and then also the raised hands.

Sneha Mehra (00:35:43)  
Okay, so notebook is still isn't found after granting GitHub. So those please post ⁓ all the access issues after this session and we'll fix those. ⁓ like notebook notebook notebook access and things like that. ⁓

Sneha Mehra (00:36:05)  
So I I wouldn't this cohort too big to manage and answer questions to everyone. ⁓ so ⁓ we'll just ⁓ gonna answer all questions posted in the course platform. So yeah, that that I can guarantee.

Sneha Mehra (00:36:33)  
So ⁓ all right, so let's go to people raising hand. ⁓ tree, do you wanna go ahead? ⁓ wow, tremendous. I've been waiting for this. Thank you, Ali. ⁓ so okay, so I come into this window expectation of what it should or should not be, but should you improve the content of the course in the future, do we still have ⁓ maintained access to the improved contents? Yes, you would have. ⁓

We would do we would improve the content as it evolves or something becomes more ⁓ relevant and you would have access to any new ones. Tremendous. Thank you. Of course. Peter, do you wanna go next? Thank you very much. ⁓ I know it has not been promised, but could we consider having some kind of shareable certificate that one can put on LinkedIn in search for jobs as a result of this? Any thoughts on that? Thank you.

Right. ⁓ thanks for the question. We can provide proof of completion, ⁓ a certificate for proof of completion. ⁓ And I assume that would also work ⁓ for for the use case. But yes, we can provide.

⁓ Rahul. Ali ⁓ I had two questions. there was one good question in the chat ⁓ in terms of the security aspect of you know setting the the agents ⁓ around using of the LLMs. Can you a little bit talk about that? Not now, but necessarily at some point in time during these five, six weeks ⁓ course. And the second question was you said you're not gonna cover agent to agent, but I think it would be good to get a perspective ⁓ on when.

You should one should set up the agent to agent versus just using a single agent. I think that perspective would be really handy. Thank you. Sure. Thanks for the suggestion. I I totally agree. For the we cover A to A in general in the learning, in the guided learning. We are not having it in the project, but that's a good very good suggestion. I would work on and we would consider adding ⁓ something relevant for A to ⁓ A at some point.

Sneha Mehra (00:38:42)  
And for your first question, security, I would get back to you on that. ⁓ I can share some resources and also think more about ⁓ the secur security aspect of ⁓ some of these systems. Yeah, especially for the enterprise grid application, the security, as we know, is gonna be very critical. Of course. Sure. ⁓ Umcar?

Yeah, Ali. ⁓ thanks for ⁓ taking my question. ⁓ I have a question about the infrastructure side. So I am particularly interested in learning how ⁓ inference and training happens at scale. ⁓ I have secured hardware for running a Kubernetes cluster and running Ray on it to ⁓ assimilate GPU workloads. ⁓ I understand that this is not going to be covered in this course, but I was hoping that.

Because this is related, ⁓ could you please provide resources for folks who are interested in this sort of thing? ⁓ And same goes for ⁓ fine tuning and other such ⁓ things that are tangi are tangential but aren't going to be covered. Thanks. Sure, definitely. I can I can share definitely training and fine tuning resources. And also if I ⁓ come to like deployment or good good things, good resources, I would share.

I also ask everyone else. I I I know there are some people expert in deployment ⁓ in this group. So please share also good resources so we all learn from it. ⁓ Debbie, can you go next? Hi. ⁓ I these are actually two suggestions around the ⁓ office hours. ⁓ One is put can you potentially have two of them? One that support just because of the size of the group, maybe one that supports the UK. ⁓

timing. ⁓ that's that's one and the second suggestion is how you'll manage those ⁓ office hours. ⁓ maybe you could first ask people to post in the chat and then you know on the on the portal first and then only the more in-depth questions go into the chat. So they kind of get advanced into the ⁓ office hours.

Sneha Mehra (00:40:56)  
There's they're just suggestions, but thank you. Of course. Those are those are great suggestions, David. Thank you so much. For the office hour, yes, let us let us ⁓ figure out and then we would announce it. ⁓ and I think that's that's ⁓ you you are right. That that that is a great suggestion. And regarding the second one, ⁓ of course, we would I think ⁓ we should we would come with a ⁓ with a strategy and we would also post it in the announcement in terms of questions.

just because of the group size, we should we would also introduce an ⁓ pro ⁓ like priority priorities. So and we will ask people to post their questions ahead if they can, and those will get prioritized. And then after that, people that have questions and they can post in the chat or raise hands. But thank you, Debbie. Those are great suggestions. ⁓ Prasad?

⁓ thanks, Ali. A couple of questions. ⁓ I'm going through some ⁓ online ⁓ like you know learning through. I'm seeing some new technologies like you know, ⁓ MCPs, ⁓ model context protocol, ⁓ planner services, like and I saw some React planner, sequential planner, something. And also the internal models. So so when I talk to my friends in the industry, they are they're most like you know.

working on training internal models or smaller models rather than ⁓ boing many of the tasks ⁓ going to ⁓ LLMs. So do you have any things included regarding that in the course or is there any if not any plans to include that so that it is it will help a lot. ⁓ Can can you elaborate a little bit more on ⁓ on on the on the content that you have in mind?

So model context protocol MCPs. ⁓ Yeah, yeah. So so MCP MCP we are covering it in the guided learnings and we are covering it in detail. ⁓ and then A to A protocol, we would we also covered it in the guided learning. It's not in the project, and I would consider ⁓ the best place that we can also introduce A to B as at some point to the projects.

Sneha Mehra (00:43:15)  
Okay. ⁓ is there anything under the planner services? ⁓ for that let me let me think and ⁓ I can share share with you later. Sure, sure. Thank you so much. Appreciate it. Of course. ⁓ Garo.

Sneha Mehra (00:43:39)  
⁓ Gao, do you do you want to go next?

Are you able to hear me? Yes, I can I can hear you. Okay. So I have a question related to ⁓ capstone project. basically, ⁓ when we build the capstone project, we might be using some kind of like ⁓ resources, GPU and other things. And maybe we wanted to try those. So during that time, like if we want to try on local or try on our own, ⁓ will there be some kind of resources provided, like GPU hours or something ⁓ as

Part of this or kind of like sandboxing provided by the platform? Or do we need to use the same platform what we have as like LLM play playgrounds here? No, I mean you don't have to use the same platform. Basically, you have full flexibility to use any kind of coding ⁓ or any kind of APIs. I know there are many providers with which provide like free access to certain models. So you can just use any of those and you can also

base your project on top of Google Collab to get access to GPU if that's required. ⁓ so those are all the options available. We are not providing any specific like credits for ⁓ particular provider as part of this course, but you are free to use any any available resources out there for the capstone. Okay, so if there is any kind of collaboration or anything, ⁓ kind of free credits will be available, ⁓ will that be shared somewhere like?

⁓ or is there any plans to bring those kind of free credits to use for some ⁓ handsome? Sure. I mean, if at some point we decide to ⁓ share ⁓ or introduce like credits or subscriptions, we would share. ⁓ but at this point we are not ⁓ offering you know subscriptions to GPUs or certain infrastructures. ⁓ But what I was referring to was like there are lots of external resources available, you know, like ⁓

Sneha Mehra (00:45:41)  
free API calls that you can use or for GPU, you can base your project on Google Colab. So that's what I was referring to. But if at some point we decide to also add like certain credits for GPUs, we would we would share with you. Sure. Thank you. Sure. Ali, ⁓ I'm sorry to interrupt here, but I have a suggestion which I think can make our lives easier here. ⁓

Is it possible for you to sign up for AWS Academy from your end ⁓ and ⁓ onboard us to that AWS Academy? Because I was in uni recently, ⁓ and for each course, each student can get up to $50 of AWS ⁓ credits if there is an official AWS Academy cohort that is created. And this I don't think it would cost you anything either, but I'm I'm not sure worth looking into. Okay. Thanks. Than thanks, Ankar. ⁓

We'll look into it. I ⁓ I I can do some research and look into it. Yeah. But thanks for the suggestion. ⁓ Thank you. Sonia, do you wanna go next?

Yes, ⁓ hi, thank you. So ⁓ I think the question if we will have, we should ask in the group chat, right? Not individually you, right? ⁓ Yes, I mean if you you that's always ⁓ better and I encourage that. It has just give ⁓ prioritize to post your questions in public channels because it's gonna benefit others and also others can share more inputs or resources. But if you have something that is not ⁓

Relevant or it's just a ⁓ certain question that you prefer to ask ⁓ directly, just you can also ⁓ message me directly. Okay, sure. And just one just one more thing that so we will be working on this project assignment of each week plus the capstone project also, right? ⁓ To be to be able to demo on the sixth week, right? Yes. Okay, okay.

Sneha Mehra (00:47:46)  
Projects projects are template and they are designed just for you to understand how some of your learnings can be applied in practice. ⁓ And it's again optional. It's more for just preparation or a pre-read kind of a thing before then the deep dive that we would do. And capstone, yes, capstone, you can you can start working on it. ⁓ Capstone is also optional, but you can start working on it from today and ⁓ the last day is the demo day. Yeah. And I saw somebody had asked in the chat, I was reading. So those

Project templates that when we will fill it in with our work, can we can we ⁓ push it to our personal GitHub account to have a record or to show it to you know ⁓ like a future employer, something like that? Or it's proprietary? ⁓ so the projects, so basically you ⁓ pull them and then once you complete it, you push to your personal account?

Yeah, so to have like to have ⁓ so so that we don't lose that work, you know, once we complete this ⁓ trading and you know later on if we want to share it, like like if I like I'm looking for work to share it with an employer, if we if I can share it, is is it ⁓ possible? I mean ⁓ sure. Yes, that that yes, that would be yes, that is possible. Okay, got it. Thank you. Appreciate that. Of course, that's all. Sure. ⁓ Tanvir?

Yeah, Lee, my ⁓ am I audible? Yes. Yeah, so my question is around the structure. ⁓ I'm sure like you explained it, ⁓ I just need further clarification. So today's live session you're covering through logistics and all. ⁓ Then we have access to the first week course. So we'll start learning. Then Wednesday we'll have office hour where we may have questions and all while we are learning, self-learning.

And then you will answer that. Then come next Saturday. ⁓ you will enable the second week course. But what will you cover on next Saturday's live session? Next Saturday live session would cover project one deep dive, the one that was released today. Got it. Okay, thank you. So in dot in total, basically, we are going to release five projects and we are going to also do a deep dive, live deep dive of all five projects. Got it. Thank you. Of course.

Sneha Mehra (00:50:12)  
⁓ Daniela. Hi. ⁓ I understood that you will not cover deploy things, but can you share with us during the the course some ⁓ resources ⁓ that you you kind of know that work for us in order to improve our skills? For example, like put in production some things that will ⁓

No here. Yeah, yeah, of course. Yeah. The deployment is such a like broad, you know, ⁓ domain, and usually there are lots of things. And but I know a lot of great resources. I can definitely share them. Yes. Like share some design options that are scalable. ⁓ Sure, sure. Of course, yeah. Again, for for a scale, for a scale, we are covering many of those scaling topics in the guided guided learning, but the actual deployment.

It's not covered, but I can share share resources. Yep. ⁓ Vinchu?

Sneha Mehra (00:51:22)  
I think you are mute.

Sneha Mehra (00:51:29)  
⁓ Vinciu, ⁓ do you want to go next?

Sneha Mehra (00:51:38)  
⁓ okay, so Roy, can you go next?

Sneha Mehra (00:51:46)  
And like clarification about what you said about ⁓ discouraging people from using AI for the weekly projects. So I usually use cursor these days and cursor's AI assistant. ⁓ and I think if I heard correctly, you said, well, I discourage people from using AI for weekly projects, but you can use it for capstone. ⁓ did I hear accurately that like I should do it by hand without cursor if I want to follow your recommendation and also can you say a bit more about your rationale?

Right. So ⁓ the projects that we release, there is it's a template project. So ⁓ a l parts of it is already implemented for you. For example, the UI or things like that. The parts that are not implemented are related to the core, some of the core topics that were covered in the in the course. So for example, to inspect certain layers, let's say for the project one, which is just building the foundations, or to ⁓ you know, just inspect number of parameters or things like that.

⁓ you can still use AI ⁓ to learn and understand how it's that's implemented. ⁓ the reason that I mentioned that is ⁓ that that is discouraged because in the in the deep dive, we would go over all of those anyway, right? And the the purpose of the projects is not just to ⁓ just complete them because there is also no stop mission mechanism or grading, right? So ⁓ if if one asked AI, AOI AI would just complement those those.

portions ⁓ and it can just be copy pasted. So so that's what I meant. You know, there is just no point ⁓ of it's ⁓ using AI. ⁓ but still if you are curious about one part or if you if you want to learn the alternative implementation or things like that, you can definitely use AI. I mean there is it's definitely permitted and also can be helpful. Okay, so if I heard accurately, so so it's not your anti-AI, obviously not, but rather the

Project exercises on a weekly basis are scoped so that ⁓ we wouldn't learn anything if we said AI help because it would just fill it in and it would be done. ⁓ so but if extrapolating, if I were to like take the project and I make my own project idea, well, if we tweak this and it's a different similar scope thing or slightly larger scope thing, but it's my own idea, okay, and I go off and use cursor, well, there's nothing wrong with that. It won't take away from my learning.

Sneha Mehra (00:54:12)  
Is that right? Of course. I mean, again, it's it's a great tool. I mean, just use AI wherever you think it would bring value, add value to what you're working on, the capstone project, brain storm, yeah, implementation, anything. ⁓ And it's up to you also for projects. Again, it's it's just a discretionary thing, right? So if if you find a cell and you see an implementation and you want to learn about alternatives, you can ask AI. Yeah. ⁓ and

When I want to learn something, I myself, I first try to implement it and then I try to use AI to check if the implementation is correct or if there are better ways or if there are alternatives. ⁓ So that's ⁓ that's perhaps can be an effective way for the learning part. Okay, got it. Thanks very much. Of course. Baron, can you go next?

Hey, hi Ali. ⁓ so quick question on ⁓ as I asked that same question in chat, but just wanted to come in here in person. ⁓ it's about this is a system design ⁓ conceptual training. So wanted to see as in when system design has been discussed, there is a major core concept which is security. And security is also another ⁓ big major point in any ⁓ system design and specifically in AI system design. ⁓ So any ⁓

knowledge or any specific ⁓ references that can be given there in this ⁓ training ⁓ cohort. Sure. I can definitely share some resources around ⁓ security. I have some friends that are also expert in ⁓ they're working in in the security field in top tech companies. I can also ⁓ get their inputs as well on some great resources that can help to you know ⁓ learn about AI security or in general security for AI systems.

And then I can share those. Yeah, thank you so much. Of course. G ⁓ Jitendra. Yeah. Hi, Ali. so ⁓ I just wanted to know like where we will be running these open source models on our local. If local, what should be the system requirement? Sure. So the projects we are going to use a mix of Google Collab ⁓ and some of them are local environment development. And those are those instructions are provided as part of the

Sneha Mehra (00:56:29)  
the project. So you would you would go over it and you see if you know it's it's encouraged to use Google Collab or local development. ⁓ for the system requirements, it most of the models that we are using are very lightweight, like two billion parameters, ⁓ LLMs or things like that. So it really does not should not require any like ⁓ like very powerful machines or things like that. And also you have always the ⁓

You have the option to switch to even lighter weight models. So like models with like less than 1 billion parameters or things like that. The rest of the code is just ⁓ the ⁓ very simple. I mean, it doesn't really need any specific machine requirements. But again, as the project gets released, you can, if you face certain issues or you have questions for which model to use, ⁓ you can just ask them in the in the chat channels.

okay. So just like eight GB RAM should be suffice for running these. ⁓ I mean, I I I think it would. ⁓ Mine is eight Gb and then I tried everything on my side. It it works. But again, if you faced any particular issue, which is likely ⁓ I also faced then I was running some of these projects before it gets finalized. ⁓ just just post them and then we'll try to figure out what the issue is. And I can also share my thoughts and if you can switch to like ⁓ lighter weight models.

So right now the alternative would be lighter weight models. ⁓ it won't be like I can run somewhere in cloud because I'm not that familiar. You can you can you can always try Google Cola for most of the projects. Okay. Thank you. Of course. ⁓ Yeah, can can I set up a cloud v virtual machine and can I practice that or ⁓ sorry, what was what was it? No.

my question is Python or sorry. ⁓ maybe maybe we take Nik. Yeah. so Ali, my question is like we all know AI field is evolving very fast. It's like sometimes becomes difficult to catch up. So would you suggest any resources, you know, ⁓ any ⁓ like ⁓ website or something ⁓ where we can see the latest news, authentic resources because there's lot of junk over there. People just copy paste.

Sneha Mehra (00:58:53)  
So any ⁓ that will be helpful for us to keep in. Yeah, I can also compile a list. I there is I have a list I I for my personal use. Those are like mostly like reliable blogs or you know ⁓ places where I can find good technical reports. I can share those in case in case you find some of those helpful. Yeah, thank you. Sure. Trishna? Hey, hi, thanks Ali for taking my question.

⁓ so as AI is evolving, we are all talking about training LLMs and you know how ⁓ how we can deploy these models and all. ⁓ But ⁓ we ⁓ d are not discussing much about where this AI can be potentially implemented when it comes to the data engineering side of things or ⁓ you know the analytics side of things. So do you have any materials ⁓

where they have spoken about ⁓ how this AI has been used. I mean what are the potential use cases of ⁓ AI in the field of ⁓ data engineering, analytics and ⁓ p other areas? ⁓ sure. I mean the way I see this is ⁓ first we started with some of these like data engineering and and then over time it became ⁓ basically I think AI is a superset of

most of this, right? So because ⁓ we actually discussed this in the in the week one learning, guided learning, that ⁓ you know, AI ⁓ has to build an LLM, it has different steps. You know, one step is obviously the data step, which requires a lot of data engineering work, data cleaning, you know, having the right data pipeline. And then after data is basically the the training and the rest of it. So that's how that's how I I see this.

So I think AI is just a superset and it it contains everything, like from data to model training to optimization to deployment to you know inference and even evaluation. ⁓ yeah. I hope it answers your question. ⁓ yeah, let me go through the guided learning ⁓ documents and then reach out to you if I have any questions. Sure, sure. Thank you. Sharas? Hey Ali. ⁓ so

Sneha Mehra (01:01:10)  
⁓ two questions. The first one is since these deliveries are you you're not expecting deliveries for each of these projects, ⁓ I think it'd be useful to see what the end output should look like in terms of ⁓ you know, user facing

It's a ⁓ application. I think it'd be help it'd be helpful to see, okay, this is what I've done and this is how it should look like and and kind of benchmark and see where we are at, right? That that kind of helps ⁓ in terms of ⁓ you know, before even the deep dive, it would it would help me ⁓ progress with my work. ⁓ the second piece is let's say if I really want to move towards you know, I'm working on my own ⁓ ideas that I've been waiting to do for a long time. help if there are any resources ⁓ shared on productionizing these systems, although this is not the

Focus of this ⁓ cohort. I think it'd help if if could share resources and see how we can actually push it to production eventually, even if it's an alpha. Of course. Yeah. I mean, those are thanks for the suggestions. Again, please post whenever you are ⁓ looking for certain ⁓ resources or interested to learn about certain topic. Me and I'm sure others can share. ⁓ so it's also gonna be more organized. So for example, post something about this topic and then I am going to share whatever I think can be helpful and also others would do.

And then ⁓ potentially this can become a valuable, you know, ⁓ like list of resources. ⁓ for the projects, there are instructions, and then at the very end, basically there is a clear instruction that, hey, at this point we are doing this, and then what we are hoping is that to see like a model perform some reasoning and output. ⁓ again, we are not focused on UIs or having fancy UIs, you know, that's not the main focus. So many of the projects are not really ⁓ creating any particular nice UI for for the thing. ⁓ So

So putting that aside, if you go through the instructions at each step, at the very end is usually saying, hey, ⁓ this is the skeleton code for a very minimal UI, for instance, and then ⁓ just complete this part and you should see like a place where you can enter your prompt and get some output, you know, like a chatbot like system, for instance. ⁓ but if again you had any confusions for any of the projects, like you don't know what the exact output you should ⁓ expect or look like, ⁓ just post it and I can share more.

Sneha Mehra (01:03:21)  
Yeah, okay, sure. Thanks. Yeah. Just that in terms of motivation, I think it would be good to see a preview of the end product and then say, okay, no, this is what I'm working towards. Yeah. Sure. Yeah. Thanks. Rohan. Hey, Ali. ⁓ thanks for taking my question. ⁓ my question is mostly around this guided learning. the video. I mean, I haven't started playing that video, but can we like ⁓ do both the guided learning and the project parallelly, like pausing the guided learning and then

implementing that in the project. ⁓ Can we do it parallelly or we have to finish the guided learning first and then start doing the project? Of course, of course you can do it parallel. So I would suggest first go over the project to understand to get a sense of what are different pieces in each project and what you are expected to achieve. And then go back and watch the video. And then as you are watching whenever you feel that some topic is covered and you feel com confident and comfortable to go back to the project, then you can go back to the project.

But basically, yes, you have full flexibility to just complete however you think is most effective for you. Got it. And the other question is I mean, I ⁓ I in my daily job I mostly use AWS services. So all this can these projects be built using AWS services? Yes. We try to ⁓ design the projects so that they are agnostic to the infrastructure or any particular provider. They're just basically focused on different pieces, how to put them together and how to run it in the most lightweight way possible.

So it doesn't require re ⁓ any infrastructure. ⁓ But that's by design. Basically that's so that you can just ⁓ use any provider or any infrastructure you want, including AWS. Got it. Thank you. Of course. Prasant?

Ali, can you hear me? Yes. Ali, the question is somewhat related to ⁓ what Rohan asked. ⁓ this is the first ⁓ AI course which I'm taking ⁓ after my college study. So ⁓ what what is the ideal way to learn this ⁓ when I come for the next week for the project one, where ⁓ I need some background ⁓ on what is touch.

Sneha Mehra (01:05:34)  
What is Transformer? What is TikToken? Right. ⁓ If I just go through the code, I am I'm using Python Java ⁓ in my daily ⁓ work. So I can understand ⁓ what programming is done, but ⁓ what are these pieces, right? Do I need to prepare that ⁓ when I come for the next class, or that will be covered as some basic ⁓ understanding explained?

Okay, so ⁓ that's a good question. I during the if you watch the videos, the guided learning videos, it would cover different, like let's say tools like TikTokizer you mentioned, ⁓ and it would try to just give a quick overview of how it works. And also it shares links if you want to learn more about the documentation or what are different functionalities of a certain tool. So if you are interested for any of those ⁓ libraries or any of those tools, you can always go to the links that are provided and learn more. ⁓

It's usually not expected. So most of the ⁓ projects, for example, if you are using a TikToken ⁓ is library, it's just ⁓ we share enough background during the deep dive. So you know what the library is, what it does, and how to use it. But if you are really interested to learn more about how it's implemented, like what are all the optimizations that they're doing and so on, you can always go to the links and learn more. Okay. Thank you, Ali.

Sure. ⁓ Nish Nishit? ⁓ hey Ali, this one suggestion and then one question. suggestion is like ⁓ you know, ⁓ to just keep all of us engaged and really serious about learning and not just, you know, doing a passive learning. ⁓ my suggestion is that if you can have some evaluation, let's say the set of questions, if ⁓ maybe multiple choice questions, ⁓ which are like related to that particular topic that we are covering for the week, but they should be difficult enough.

Or somebody to go deep and answer based on the ⁓ contents that you have posted, it like forces us to study more and understand more about the topic. ⁓ and then you know, we are like maybe it is optional for us to answer, but it gives us more confidence that you know we are learning something. So that was the suggestion. Sure. So ⁓ just to make sure. So the suggestion is for some parts of like for certain certain topics or contents, we ⁓ I ask questions in the channel, so we

Sneha Mehra (01:08:03)  
people can also answer, right? Yeah. Yeah. But that's one way, other ways like everybody answers. you know, like let's let's say multiple choice kind of format or ⁓ whatever works like ⁓ best for you. I think I leave it. I see ⁓ sure. Yeah, yeah. That's a good point. Yeah. So other question is like it's a general question, ⁓ based on your experience because you have been into this field for a long time. ⁓ is there an optimized way according to you that you want to suggest to take notes out of this course because

Note taking is also like a you know very ⁓ I would say very efficient way to ⁓ relearn and reinforce your mental learning. So you have any tools or or any particular style of taking the notes out of this course? Yeah, ⁓ well obviously different people have different styles for, but I I know there was this AI, there is an AI tool ⁓ which you can install it and then you would automatically create notes ⁓ from all your meetings ⁓ or or something like that.

Let let me ⁓ find out what the name was and I can share with you. You can you can consider it as one option. Sure. Thank you, Ali. Sure. Anurak?

Sneha Mehra (01:09:15)  
Hey Ali, thank you so much. ⁓ So, one question on ⁓ a little bit of background, you know, about me. I'm working for an organization where we are trying to build an application that has something similar use case like perplexity, but something on our own internal ⁓ content. Now, the the biggest challenge that we are currently facing ⁓ is ⁓ the latency because these ⁓ large language models that we

refer ⁓ use from you know providers like openai ⁓ or or ⁓ AWS bedrock and all that they they do take time ⁓ and then we also have a lot of steps in in those processes, you know, and we ⁓ build those pipelines and all that. So so I understand you are not going to cover deployment as part of the whole course, but but ⁓ but at le I was hoping if we can touch ⁓ some portion where you know at least

high level architecture to like, okay, maybe these kind of problem to handle these kind of solutions, you know, that maybe the highest level architecture could look like. Because ⁓ traditional APIs usually expect you to just return response within fraction of or maybe milliseconds and all that. But these language models take like 10 seconds or 20 seconds to respond. So ⁓ sometimes these kind of you know things break the front end part because you know you cannot open the connection for that long and all that.

So there are so many issues. So that's one immediate challenge that we are facing. And I was hoping, you know, if we can get some guidance on that. And two is the ⁓ model evaluation, because there are so many models coming in. And how do you effectively evaluate a model using a framework? Because each model have its own prompt library and and then its own comes with its own challenges to write the you know behind the scene pipeline to to get to the model and all that.

So these are the two areas, you know, I was I was ⁓ l looking for if if you you are going to cover those in any way. Of course. Yeah, both are very important areas. ⁓ to some extent they are covered in ⁓ in the videos. ⁓ in in certain domains it may not be covered. If that's the case, for example, ⁓ we if you have a very particular question about a rag-based system, how you optimize how to how you evaluate different components to make sure that both the cost and latency is optimized. ⁓

Sneha Mehra (01:11:39)  
Some of those very particular things may not be covered, but yeah, again, feel free to post any of these. I'm sure I I can I would answer ⁓ any of those with any resources. And I'm sure others can also opinions they have to share and input. So ⁓ but I would say 50, 60 percent of what you were ⁓ the areas you were pointing out are already covered in ⁓ like including the evaluation of LLMs and you know rags and so on. Got it. Okay, thank you. Sure.

⁓ Chaishta?

Sneha Mehra (01:12:13)  
I I'm so sorry I undercut the people in the queue earlier. my question was ⁓ I'm not a very technical person. I'm not a I I don't work in a tech-related ⁓ profile right now. So ⁓ for me to start developing my understanding, I wanted to understand the whole landscape. AI, as you said, is like a superset, and there are some use cases that are ⁓ you know, ⁓

More developed compared to others, and I want to d develop my understanding of this. ⁓ somebody also mentioned earlier something similar about use cases, and I'm sure we'll go through that. And I'm sure at this point we'll stay to post it in the community. But ⁓ any ⁓ anything that you're ⁓ engaging with us, if you ⁓ all ⁓ include ⁓

comments or content on the actual use cases so people like me can get a better understanding that would be very helpful. Of course, of course. ⁓ Definitely. So the the the guided materials, they they go over the use cases and also the project are basically a direct use case of that part that particular week's topic. ⁓ but sure I can I can share more use cases as they come to my mind.

For different topics or different weeks. For sure. Like you mentioned that whatever was not practical is something that has been removed. I would also like to understand what is not practical because from one point of view, that might be something to look into, you know, that might get developed in the future or something like that. I would like to develop my understanding of the whole superset that there is, even if ⁓ some things are not as feasible as others. Thank you so much.

Of course, of course. Yeah. And also there are more links, you know, if you are interested to learn more about something, you know, throughout the the course. But thanks for the suggestions. I would I would make sure to to include some of those. Thank you. Yeah. Emma. Hi, Ali. Thank you so much. ⁓ I bet there are quite a few folks here who are thinking about switching into AI engineering. ⁓ we'll talk about how someone can make that move from ⁓

Sneha Mehra (01:14:34)  
General software role, like what what kind of prop is helpful for interviews, ML coding practices, and other practical steps. ⁓ So I think this diagram, ⁓ this sorry, this visual would probably ⁓ summarize, you know, what needs to be learned for or expected from an AI engineer. And we try to cover almost ⁓ almost all of this we are covering in the course.

So that's basically the whole landscape of AI engineering, or what is expected from an AI engineer to do. In terms of ⁓ coding, the projects that we have are just an example, right? So basically they're just an example, for example, for agents, it's just an example how to use Langchain or how to use ⁓ so you learn different agentic setups and different like loops or React or things like that. And then in the project, you see.

for instance how you can use lang chain and then it would just give you the same ⁓ workflow or the same loop. ⁓ and then you would build that project just as an example. once you know once you learn the Langchain and how you can set up an agent, for instance, ⁓ you can potentially prototype other ideas or build different things. ⁓ And again, I keep I plan to keep sharing whenever I come with come up with something interesting like some

⁓ interesting reads or some good papers. So in case you are interested, you can ⁓ you can read it. ⁓ And everything I would I plan to share is more for AI engineering. It's focused on AI engineering. Thank you. Sure. Sonia?

Yeah, ⁓ actually even I I I had the same question which Emma just asked that like I even I'm ⁓ like an SDET and ⁓ you know junior Java engineer. So coming from that background, so this training, like somebody ⁓ or even I think maybe you meant somebody it came it came up that it's a systems design training. So ⁓ a little bit if ⁓ you can tell that ⁓ you know what ⁓

Sneha Mehra (01:16:48)  
Like ⁓ is it ⁓ w what audience are you targeting or and you know like ⁓ like architects or product managers or engineering. I think you said engineering earlier, but I you know I just wanted to get clarity on that. So you know, like like ⁓ you know there could be is it relevant for ⁓ my you know area or you know ⁓ just ⁓ just that. ⁓

Yeah, so ⁓ who we are targeting is basically a couple of different people. One is people in a different domain, let's say a PM or or in a very different domain, and they want to transition into AI. And the reason for that is most of the guided learning videos, they are starting from very basics, explaining like what is the goal, these are some of the basic things, and then it it just builds ⁓ slowly. So those are one group of ⁓

people that I think would be very the the course would be very valuable. The other group is ⁓ software engineers that are trying to transition into AI. again because just we start from basics and we explain all different topics and ⁓ once we talk about the topics we go over important and you know popular features or applications and we ⁓ explore their ⁓ system design. ⁓ so I would say if

For example, you are a researcher in AI and you want to publish more papers or you want to you know continue the field or invent something newer, probably then this course may not ⁓ it's not they are not going to be at the target audience for this course. So if if you want to continue doing more research or or you are at the edge of AI research or engineer and you want to optimize something even more or build something even ⁓ more smart.

So we are not covering those. But what we are covering is just starting from the basics and then going over all the ⁓ AI engineering take a step. Does it answer your question? Yeah. So I mean, so just like for software engineering, like okay, now you know AI, ⁓ AI is the big thing now. And ⁓ you know, the more in software engineering, like for me, ⁓ we we know.

Sneha Mehra (01:19:09)  
And we can use it in our work. It would be, it is kind of required by the industry. So then for people who are in software engineering, it would be it would be relevant and we are the to target audience. Is that right? I would say software engine software engineers and people from different backgrounds who want to transition into AI are target audiences. Yes. I see. Okay, got it. Thank you. Appreciate so much. Sure. ⁓ Nishit?

⁓ Haley just like one suggestion. ⁓ because like if some someone of us wants to go to the recording again or somebody who's what must have logged off by now, ⁓ is it like ⁓ are you planning to share the this Zoom summary transcripts, maybe using the Zoom AI? Like we can just go through the text rather than going to the complete recording. Sure. I yeah, I'll try to ⁓ yeah, I'll try to find ⁓ to to create a summary, but sure, yes. Thank you.

Sneha Mehra (01:20:10)  
⁓ next hand is ⁓ Sonia. No, no, I'm good. I have to lower my okay. ⁓ Ashmi. Yeah, ⁓ my question was mainly on the last week, the Capstorm project. So I'm ⁓ I'm guessing like we are in a cohort course, and there's a lot of people with a lot of domain experience here. ⁓ And I think each of us would love to see like

Okay, how we use how our individual ideas, you know, how our domain, how we can use AI in our domains, and that that's how I guess the capstone project will go. And I was mainly curious about the part where people demo their projects or other people can see what ⁓ you know, like me, someone like me can see what everyone has done and all, and how do I gain experience and yeah, mainly the demoing part, because there's like nearly 400 people, right? And

I'm guessing like the last week, the meeting and all within two hours, do you think they'll get over? And yeah, just ⁓ yeah, how will people get access to maybe all these ideas? That's a ⁓ sure. ⁓ so I think it also depends on ⁓ how many projects, like also how the groups would be formed. Some people want to work on their own project by themselves, some may form groups, you know. ⁓ so what we plan to do is later ⁓ we would post

in the announcement and then ask for people to you know ⁓ share their ⁓ their project just just the name of the project or something like that so we have an idea of how many projects are expected to be demoed and then based on that we can we can plan accordingly and we may even end up with you know longer session ⁓ or we can also use an additional session so we make sure that everybody can present or at least the recordings would be available to everyone. All right thank you.

Sure. Do you want me?

Sneha Mehra (01:22:11)  
Hi, can you hear me? Yes. Yes. ⁓ I wanted to ⁓ I had a bit of a suggestion. And this is coming from something I touched on earlier, but from listening to ⁓ everyone's questions, I I see there's a theme here. So ⁓ for the parts that we wouldn't be diving deep into, and you know ⁓ execute this how you want. But it would I think it would be helpful if just for example, let's say LLM playground or whatever rag chat it is, that

There's some ⁓ maybe notes section or something that lets us know sort of the different parts and the different parts that are not being touched. Right. ⁓ I'll give an example. So, for example, if I'm if we're doing a capstone project and there's one in mind, there's you could easily fall into a place where you have a capstone project in mind, you think it makes sense, and then as it goes further, you realize, no, this requires more significant training. This is going to be costly. I took the wrong turn and not realized.

And didn't realize, right? You mentioned something like ⁓ somebody asked a question and you mentioned, no, this is gonna be like two two billion per parameters, or you can take it down to one billion and run that on your own computer. It's not a lot. I don't think everybody is as familiar with what type of number of parameters would make it feasible or not. So I think if there were some sort of notes for maybe each project or however you want to put it, so the person understands that ⁓

The these are the parts we won't go through. But for these parts that we won't go through, this is the template or the replacement or whatever placeholder. And if you want to do more, these are the type of cost consideration or feasibility considerations that would really help. I think you do a great job with that with the with the U UI, right? We know that this is not about UI, but because you tell us that there's a template, we don't have to worry about it. But in the parts of deployment and then the cost, it can be a bit difficult to assess whether.

the path we're going on is practical because we may not fully understand what the placeholders are. So some sort of note or some sort of rundown to know that we're not using a main thing here. We're using an open source. ⁓ the the alternatives would be, you know, some something would really help us feel like, okay, this project, this capstone, whatever idea is actually feasible and practical. Sure. That's a great suggestion. Thank you. ⁓ of course.

Sneha Mehra (01:24:35)  
So I as you so you're right. I mean, for part of it we already have, like so you don't have to worry like UI. ⁓ but I think that's a great idea ⁓ to to ⁓ to add things about, you know, cost versus requirements or things like that. ⁓ I'll definitely work on it. Thank you. ⁓ Vinci? Appreciate that button

Sneha Mehra (01:24:59)  
Hi, are you able to hear me? Yes. ⁓ hi, Ali. so my question is ⁓ so ⁓ you have ⁓ authored the generative AI system design book, right? So how can I better use this book along with this course to prepare for ⁓ big tech system design interviews and also if you can shed some light around how can we prepare on ⁓ ML coding and system design interviews with respect to machine learning? ⁓

If you can make a video that would be really helpful. ⁓ Sure. I can ⁓ I can share some some of my thoughts about the books. ⁓ they are generally can be helpful for interviews, both the machine learning and gen AI, depending on the interview role, if it's more focused on search or recommendation than machine learning interview, or if it's more focused on Gen AI, the other book. ⁓ and sure, I can share some ⁓ you know, like best practices or some

You know, like ML coding resources, if if that's gonna be helpful, I can share it with you.

Sneha Mehra (01:26:06)  
Rajesh. ⁓ hi, ⁓ Ali, can you hear me? Yes, I can. Perfect. Yeah. ⁓ so a couple of confirmations and then a few questions, right? So the tree rec, ⁓ there was a bunch of ⁓ links around Python, PyTorch and all that, you know. So here's how I'm planning to go about the first phase, right? So do some of those links, you know, fast forward, fast ⁓ quick go go over that quickly and then ⁓

Go to the guided session as you ⁓ showed me, work through that example. Is that what is required ⁓ for the next course, next session? ⁓ you know, the some of those links are more for if, for example, if you have no idea about Python or you want to learn some basics of Python, then okay, then there is a link for that. ⁓ but in general, the lecture, the guided learning video should give you everything you need. ⁓

Yeah. And besides that, you may need some occasional searches if you are using a certain tool in the project, you know, like know how to call a certain ⁓ function or method. But you know, 99% of ⁓ what you need is expected to be covered in the guided learning. Those links are more for like if you need additional information about something, or if you have no prior background in something that ⁓ like Python, you can go over those. Got it. Yeah, a few other couple quick questions.

So ⁓ one of these things from practical matter perspective, what would be helpful is AI seems to be like the new hammer, right? ⁓ it would be better to know ⁓ when not to use that and why it should be used, because there are probably ⁓ ways to solve this ⁓ certain set of problems without EAI. ⁓ so that distinction, ability to distinguish would be super helpful. ⁓ second, in the enterprise context.

How is a certain subject area being used and how do you see it evolve in the enterprise context? You know, that might be helpful. It's more relevant and pragmatic and practical for us ⁓ for some of us. Yeah, I mean, yeah, go ahead. Yeah. And lastly, I did see an AI companion on Zoom. perhaps I could use with the help with the recording. I don't know how it may have to be enabled by you. Yeah. Okay. Okay. Thanks for letting me know. ⁓ for the enterprise thing, ⁓

Sneha Mehra (01:28:29)  
I can also share more later, but I mean some weeks are ⁓ can be relevant to to enterprise as well. For example, we one week we focus on building a customer support chatbot. ⁓ and you know the entire week, like rags, everything, ⁓ can be super relevant, you know, for for you know certain use cases at at enterprise. And same for agents ⁓ and tools, you know, internally you can have your own set of tools and then you can just introduce those through mcp or any other way to the to the

Back on LLM and then kill your agent. ⁓ yeah. I I meant more along the lines of you're taking on a topic, how is that currently used in an interface? Some examples. It doesn't have to be covered in the session itself if you're time crunch, maybe ⁓ along the links that you're going to provide. That might be some, you know, instead of using yeah, that's what I was trying to say. Thank you. Got it. Sure. Thank you. ⁓ Yeah. ⁓ Yash.

Hey Ali, thanks for taking my question. ⁓ I have like two questions. ⁓ first one is with regards to something that was touched very recently, where someone mentioned like ⁓ the content from your books could also be used in conjunction with the coursework that we are doing right now. Do you think we will have access to your books ⁓ by taking this course? ⁓ And and is it something that will be made available to everyone?

⁓ so ⁓ we are not offering ⁓ you know additional memberships or books with this course. ⁓ yeah, but the books can be helpful. Okay. ⁓ yeah. And ⁓ with regards to the content that you are covering in the next week, in addition to like the discussions on the project, are you also going to introduce

new concepts and teach about those new concepts or is it only going to be part of the guided project video that you create? sorry, can you repeat your question again? For ⁓ for all the live sessions that happen on Saturdays, ⁓ are we only disc ⁓ usually the agenda is to discuss the project, but is there any other concept that will be covered in addition to the projects that we discuss? Not not not le not new learning concepts.

Sneha Mehra (01:30:45)  
What we would what would be covered is ⁓ like the things related to the code itself ⁓ and the pieces in the code. So all the learnings or the introduction of the concepts are expected to happen in the in the guided learning video. But anything related to the coding, the libraries that we are using, the tools, ⁓ the details of those, those would be covered in the project deep dive. Got it.

And generally, like because I also want to understand like the office hours are only one hour or maybe like one additional hour. How are we managing that one hour for all the four hundred participants, like who may have questions? ⁓ so it are there going to be how many people are going to attend ⁓ the ⁓ answering part of the office hour sessions? Like who is going to how is that going to be managed?

Okay. So y let us we would announce it. we would have ⁓ internally a meeting about this and figure out the best and most effective way. And then those meetings generally can also ⁓ go for longer in general. It can be more than one hour. ⁓ it doesn't have to stop right after one hour. But ⁓ let us let us figure out and we may ⁓ also suggest more more office hours just to ha keep it more manageable. Got it. Thanks.

Sneha Mehra (01:32:04)  
⁓ so ⁓ the I I Nuraq.

Sneha Mehra (01:32:11)  
⁓ Ali one question on the capstone project. ⁓ is there a way to possibly instead of at the sixth week, can it be like seventh week or so? Because I understand we would be working hard four or five weeks continuously and then not sure if one week would be enough for everyone to complete their capstone and deliver. ⁓ so yeah. So that's a that's a good point. I'm ba basically the last Sunday is more like a not a capstone week, it's a capstone

demo. And then the whole capstone is expected to be completed in ⁓ throughout the course. And that is what I also suggest because ⁓ as you learn, probably you would learn how to implement different different parts of your capstone. ⁓ And the projects itself, I mean the weekly projects, they're not gonna take you a take a lot of time. I mean it depends on how how how how fast you go over them or implement them. But the projects are more like assignments which you get time for one week to work on.

And then the capstone is the the like the real project that you would do. And then once for the weekly projects, ⁓ it's not gonna take a lot of time. And then the deep dive is where we actually go over every single detail of it. So I would suggest for the capstone, if if ⁓ you can start ⁓ from the beginning or think about it first week, you know, about different ideas or what you want to build, and then gradually as as we move forward with the weeks, you start building different pieces.

I would say probably that's the most e effective strategy. Okay. Thanks. Yeah.

Rahul. Hey Ali. ⁓ long term Alex Zoo and Early's fan here. Call myself Alexian or Alien. ⁓ I had a question about ⁓ piggyback from the first qu the last question. I had a question about the capstone project. Is there any way we can see the requirements for the capstone project? ⁓ especially like I'm

Sneha Mehra (01:34:12)  
⁓ I've been in the engineering management for a while and I'm a little rusty on programming now. So I want to get started back up. ⁓ so it's gonna take me a while ⁓ like to to get situated. So those requirements would be really good guideline for me to see where I'm at ⁓ in preparation for captain project. Sure. So by by requirements, you are referring to ⁓ like kind of ⁓ infrastructure requirement or requirement in terms of what topics to choose or things like that. Yeah, like what what what would be the deliverable basically?

⁓ like how how how how what whatever that might be is it is it windows app it is web app ⁓ how what are like basically requirements for like expectations from you or Alex ⁓ for this sure sure i i can share more more info on that but but but i mean in general it's it's supposed to be very flexible right so there is no a strict requirement you can yeah and also there is no extra strict requirement in terms of the delivery deliverables right

It can be just a slides, it can be full prototype, it can be partial prototype. ⁓ but I can share some examples that you know these can be some potential examples that you can work on and build. But again, the expectation is not to build a fully working production ready system. It can be even a slides or ideas. It can it can also stop at the ideation phase, you know. ⁓ it's all up to you. I mean,

Many of people in this community are, you know, professionals working full time. So they may not find enough time to necessarily implement something within the six week period. ⁓ just the idea is to ⁓ cre use this community and get feedback and inputs from everyone. And then we continue as ⁓ and build as much as we can within the next week, ne within the next six weeks. And then after that, potentially it can be continued also and extended. Awesome. I appreciate that, Ali. Thank you.

Sure. I think there are no more questions and we are a bit over time. So if there are no more questions, ⁓ we can end it. Again, I want to thank everyone. ⁓ there was a little bit of confusion in the beginning in terms of the chats, like questions asked. So we'll make sure to introduce some mechanism to prioritize and make sure that all the answers would be answered ⁓ correctly.

Sneha Mehra (01:36:30)  
So yeah, thank you so much all for joining. Feel free to post your questions in the chats ⁓ and I'll try to answer them. Thank you.

—-------------------------------  
05  
Sneha Mehra (00:00:00)  
Can start. ⁓ let me ⁓ set up everything. ⁓ Perfect. ⁓ hey everyone. It's it's great to see all of you again. I hope that week two was you like the content of ⁓ adaptation techniques, rack, ⁓ like fine tuning, prompt engineering. Basically, the content of week two is most relevant to a smaller companies or startups.

⁓ or any company that they don't want to go through training the entire ⁓ foundation model training and every and things like that. So ⁓ basically most startups that are focused on applications, they typically want to start with ⁓ one of the approaches that we've seen in VIC2. It they can start with RAG, and then later they can switch to parameter efficient fine-tuning if RAG is not sufficient for them.

And then project two was also designed so we get a sense of what it means and how simple it can be in practice if you rely on ⁓ you know ⁓ robust libraries out there. ⁓ so yeah, today we are just gonna go over ⁓ project two. I I assume you my screen is shared and you're seeing the content. So we'll go over everything and we'll explain different libraries that we export. ⁓ I'll try to also

occasionally deviate the from the correct implementation a little bit from time to time just so we can do some experiments live. ⁓ So that's one thing. The other thing is that last week we went ⁓ much longer. ⁓ And that was basically intentional because the content of week one was foundation training. So a lot of the things that we've been doing was more focused on learning the internals and what's like how things work under the hood. So there was basically more intense.

So this week, my hope is that it's going to be shorter because we are gonna rely more on tools, tools and libraries. And then we will get more time to ⁓ take questions toward the end. So, ⁓ all right, so let's start. Let me also ⁓ perfect. So this project basically is just a sample project. It's just a tiny example. ⁓ I have created these ⁓ PDFs.

Sneha Mehra (00:02:20)  
Like an imaginary ⁓ you know, ⁓ e-commerce store. And then if you if you open them, ⁓ I think I have to go to the git git here.

Sneha Mehra (00:02:36)  
Basically, you would see typical frequently asked questions or, you know, other kinds of policies that companies typically have and they store it in some format like PDFs. So this is another one which says size and care guide, and you can see there are tables and their items. It's not necessarily super clean. ⁓ you know, it's it's for the formatting may have issues. So these are all intentional, just to the just to get a sense of how ⁓ PDF loading can work and what are the strengths and limitations.

⁓ and then here is another document for like exchange and policy, and this is the last one for delivery shipping and delivery policy. Basically, who we ship and processing time, shipping rates, and you see again there are certain formats and ⁓ like table-like kind of formats, and then changing address, lost and stolen, ⁓ and contact. So again, these are just examples. In practice, we don't have like four PDFs. In practice,

Companies, startups, like e-commerce stores, they have hundreds of PDF files. And it's not only PDF, it can be other in other formats. It can be some images, it can be also some URLs, ⁓ which is pointing to the URL with some content. And in practice, company basically have all these external, like internal knowledge base, and they want to use them to ⁓ inject this information or knowledge to the LL. So

Yeah, just wanted to set up the context. Here we have just four ⁓ small examples. So that's the data. And then this environment was provided for to install ⁓ your condo environment. We name it Rag Chain, Rag Chatbot. I think I need to update it by adding PyPy kernel, which was pointed out by some folks. ⁓ but basically just specifying different libraries and different versions. This is this is important for consistency.

And that's because ⁓ I have seen some questions ⁓ on the portal. And those questions are most mostly around like versioning and you know this version doesn't work with that version and conflicts. So that is another benefit of keeping things consistent, making sure that you know this set of versions keep working with each other. And then ⁓ so that's about the versions. ⁓ And again, all these libraries and frameworks keep changing and evolving. So today.

Sneha Mehra (00:04:56)  
It may they have they may have a version that works and tomorrow they may release a new version. So if you just simply do pip install lang chain and install the latest version of Langchain, it may suddenly break things. So that's just another thing to keep in mind. ⁓ I tried with these versions, it it seemed to be working, and I created my conda environment. ⁓ and we'll just move forward with this. Again, changing this can be tricky and a lot of times in companies it takes a lot of time to figure out if if something breaks. ⁓ so that's about this.

And now let me also open my terminal because for this project we are going to run things locally. So ⁓ okay, I have this terminal open here. And then I am in project. If I do ls, I have this data environment and rag chain. Also from before, I have installed the ⁓ the conda environment. And you saw that the environment name is rag chatbot. So we can also do conda ⁓ and list, I think.

And it would show me all the different environments that I have so that I can activate the one that is going to be useful for this particular project. So I'm going to do conda activate my environment, right? Chatbot.

Sneha Mehra (00:06:11)  
So here it says I'm in this environment, right chat button. And just to do a quick sanity check, I can do show ⁓ keep show one of these to make sure the version is correct, like a streamlit, for instance.

So here it says one 45.1, which is ⁓ it seems correct. All right, so now I have my new my environment. Let's get back to the Python ⁓ notebook. And then here I need to also select the the right kernel, which is the same environment that I have just created, like before this session. So I select rag chatbot here.

Sneha Mehra (00:06:55)  
All right. So ⁓ so that's it. Now the next thing I we need to do, we just need we can just move forward. So some of the things that we would do here is we would ⁓ figure out how we can load this data and PDFs into an extra into text. So basically going from an on a structured documents to text, and then we'll try to embed them using chunks and index with a vector database like face. And then after that we'll just

play around with different prompts and see how it works. And then we'll switch to OLAMA, which is an open source framework to manage and load LLM locally. So it's it's very powerful. And for local and development environments, it can be very, very handy and useful. So we'll we'll get to it. We'll also build our rank finally Rackchain using ⁓ Langchain. ⁓ and then after that we'll just show a quick streamlate demo. ⁓ I'm not good in UI, so I just created again another very minimal ⁓ kind of ⁓

Not so fancy UI. So we'll just so that we can just see everything in practice in UI. So let's start. ⁓ here I have been through this. I have created these environments. So we don't need to do any of this. And here there are some libraries. ⁓ I this time I have added the explanations, like you know, why we are importing each of these, so you you have some sense ⁓ of why we are importing them. But basically, ⁓

These are mostly used for ⁓ loading different ⁓ unstructured data from different sources. So for example, Langchain provides all these functionalities. So if we search ⁓ what is Langchain, it is basically an open source framework for building applications powered by LLMs. So everything around LLMs that you may need, including loading models, ⁓ even retrieval like face and those kinds of things, creating chains and ⁓ all those things are provided by Langchain.

So it can be super useful whenever we are just focused on building prototypes or like a small applications. ⁓ all right, back here. ⁓ and then we we just load a bunch of things. If you're interested, you can also look at document loaders to see what are all the functionalities that Langchain provides.

Sneha Mehra (00:09:15)  
So here is the document loaders, and you would see all different types of documents that you can potentially ⁓ include in your database and then load them. For example, you can add CSV loader ⁓ or a bunch of other things. So just you can go over this and see what are all different types of data that you can ⁓ use to load.

So that's about this Langchain. And ⁓ here basically it's it's more focused on splitting text, which we saw in the lecture. And then here is more for face and setting up our retrieval ⁓ part. This is the embedding part. Again, it's related to the retrieval. We need to retrieve we need to embed the sentences or our chunked documents and store them. So we need some model, some pre-trained model to embed those. And these are some of the libraries we just import for that purpose.

OLAMA, just so we can load LLM. And also one thing I want to note here is that these libraries have their different frameworks that are providing these functionalities. For example, here we are importing face from Langchain, because Langchain also provides this face functionality. In practice, we don't have to. I mean, face has its own ⁓ you know D triple and documentation. So here ⁓ is face.

And face again, I we've went over it in the lecture. It's basically a library that allows you to run nearest neighbor or approximate nearest neighbor so that you can find relevant documents to an input query. So here, you know, it's very simple to install and then ⁓ use it. I think if I can find here ⁓ there should be some examples.

Sneha Mehra (00:11:03)  
No quick kind.

Sneha Mehra (00:11:09)  
⁓ I have seen some examples somewhere. ⁓ I just maybe Viki. Getting started. So here are just some examples. I mean, once you have face installed, you don't have to necessarily install the Langchain version of face. You can just install the the original face library and just start using it. If we get time, I'll try to just play with this for a little bit ⁓ during this this session, just so we get a sense.

But again, here we are just using the language, both for face and OLAM. And then this is for chain so that at the very end we can just connect everything and create our our ⁓ like chain, entire system. ⁓ And prompt template will go over it. So let me run this and make sure we are good in terms of ⁓ libraries. So we were able to import everything, so we are good to go.

Now the first part is preparing the data. And preparing mostly means that we first load the data, make it into ⁓ make it from c like bring it, bringing it from a storage to our memory and then have it in a textual format. So this is the first step. If it's text, if it's image, we we'll still need to bring it into memory and then have it like in a pixel format. So in this project, we are going to focus only on text and PDF files. And then after we load it, we need to ⁓ chunk it.

Which will just use some library. So there is not really, we don't have to really implement most of this like from a scratch. We can just use whatever is out there. All right. So here the first step is just to let's load them. The first ⁓ the first line just reads all different files, starting with ever storm, and then keep this ⁓ all these file paths in this ⁓ object. So ⁓ let me I think if I run this, it would work now.

I'm running this, it's working, but it just says zero because we have no code here. So now if I print PDF path, PDF ⁓ docs.

Sneha Mehra (00:13:10)  
PDF Pat.

Sneha Mehra (00:13:17)  
Print. ⁓ Sorry. ⁓ so we'll see all this path like ⁓ four different PDFs. And easily we can scale this. As you have more documents, you just put them here, and then this line would just load all of them in a list. So, okay, we are good here. Now the next part is just load them. And again, there are different ways we can load them. First, let's just ⁓ use the Pi PDF loader ⁓ that we imported. Again.

Nobody is expected to know how to use these methods or libraries. ⁓ The most practical way, because they keep changing, and also new libraries keep coming up. So there is not really feasible to know all like what is the functionality, what are the inputs, how it works. So the best way is just copy here. I actually searched PDF loader before before this session. ⁓ And there are a bunch of different PDF loaders, but the first one on Google is ⁓ the one from Langchain. So basically it's kind of

validates the fact that we are using a a reasonable library. Now I paste this here, I go here, and then I just try to ⁓ figure out what are the inputs, outputs, and how is it possible to load a document. And again, there are different ⁓ ways of loading it. There is lazy loading and there are lots of different ways we can load. Here we can just ⁓ specify the path and also the mode and then we call the load.

So I just try to follow something very similar. So here, just for now, temp py PDF loader, and then I'll just send one of the ⁓ paths that I've loaded. ⁓ now if I print this dev here.

If I print temp, it's basically just returning a lang chain ⁓ PyPDF loader at this point. Now we also need to load it. We had this load here. ⁓ so let me call load and then we'll see what happens.

Sneha Mehra (00:15:22)  
Okay, so great. So now it is actually a document object. It has the metadata. It has also, so it the metadata says ⁓ you know, like different things like creator by pipe PDF and all those kind of things. And then finally, perhaps it's the name of the ⁓ file or no. Return and so these are all obtained from the the file itself. And then the source and then total pages and everything. So these are the metadata, and here we have the page page content.

So ⁓ that's awesome. That's basically all we want. We have the actual text now. If if we print the ⁓ page content.

Sneha Mehra (00:16:10)  
It is basically just returning the actual ⁓ text. So this is awesome. Now ⁓ now we are ready just to finish, just to implement this part. We don't have to just do a test here. We can just loop over our entire ⁓ PDF path. So for path in PDF path. And then here we have ⁓ raw docs dot ⁓ extend. Again, there are different ways we can do this, but I'm just gonna do it like this.

Then we already had this. So for this, I can just do ⁓ temp equal this. And then ⁓ this dot extend the temp. And the reason that we are doing this is because temp is is is it's basically returning a list. I mean we can also do append ⁓ temp zero, but these are all all the details. I'm not gonna go over them. So here we already printed this, so this should be fine, and then we just append it. So now let's

Run this so ⁓

Sneha Mehra (00:17:17)  
Rodox. So let's also switch back to the extent just to make sure that if we bring a button.

and we no longer no need this because ⁓

for path in PDF path, then we'll load the path.

Sneha Mehra (00:17:39)  
Perfect. Loaded eight PDF pages from four files. So everything seems right. We have four files and then probably there are total of eight pages. Now, if I look at one item in the Rodox, it's just gonna return the first ⁓ the first document. So it has the actual text content, like frequently asked questions, which was the title, and then the rest of it. All right, so we are good. Now we loaded everything. Now

This part is optional. That was not supposed to be part of this project, but I just wanted to give a sense of that the same thing can be used for ⁓ with URLs. So ⁓ so I just put some URLs here. Let's just go there and see if they still exist.

Sneha Mehra (00:18:29)  
Okay, so these are just some random URLs. Again, they are they're very, very random. I just found them. I just put them here. So one is about shipping and checkout APIs and everything. And then the other is about refunds. Let's see if this one also exists. Potentially you can replace this with your own ⁓ URLs for your own retailer store or startup or company. But basically, this page you can see there are bunch of tables and URLs and lots of different things.

So the purpose here is to now start use this URLs as as part of our knowledge base.

So ⁓ all right. So, and I got again, there are some examples for you to uncomment if you're interested to play with this. So, here we can rely on web base loader from document loaders. I I searched this before this meeting, this live session. So if we go to document loaders, there is this ⁓ let me actually search it. There is this web base loader, which

would just work. You can just pass the URL that you are interested and it would load it.

So that's that's it. So let's just use this and see if we can actually load ⁓ our pages. ⁓ So ⁓ I'll just for simplicity, I'll or actually ⁓ how to do it so it's simple. Okay. Let's just first create a loader object. I'll just follow what it's it's here. We can just follow that. ⁓ loader, we create this web space.

Sneha Mehra (00:20:09)  
loader and then we can pass all these URLs that we have here. And also we need to import this. So from

Lang chain community dot document ⁓ loaders import web based loader. Hope it works. And now here if I ⁓

Let me also remove this ⁓ fallback part. So ⁓ here we have this loader and then now if I print it.

Sneha Mehra (00:20:53)  
Okay, it's loading. It seems there is no error here. And then I have the loader, but we have not loaded the documents yet. So the next thing is we we actually need to load it. So we can then have low raw docs equal ⁓ loader dot load. It's basically the similar pattern that we used here. ⁓ We create the loader object and then we load it. So here I'm doing the same ⁓ and then ⁓ load the raw documents. Let's see, we actually just get the content.

This is just optional. So I just want to see if if we can actually get it. ⁓ raw docs zero.

Okay, so it has the URL and everything seems fine. And then it's it's basically the content. So so basically that shows that we can simply also add URLs and include those as part of our external knowledge base. ⁓ so I'm going to just for now I'm going to ⁓ copy paste the more reliable code here for for reference, but it's the same logic. It's just ⁓ web based URL and then it loads.

And then here is the fallback mechanism. If if ⁓ there's some issues, it it has some way to handle it. So ⁓ and we are not gonna run this again because we want to continue the rest of this project based on the PDF. So I'm going to rerun this to make sure I'll ⁓ I'll end up with the raw right ⁓ Radox. Perfect. And I'm going to remove this also. Okay. So ⁓ now we have the PDFs. Here was optional.

Now the next step is chunking part. This is going to be also very simple, again, because the libraries are out there. ⁓ If we have imported recursive character, text splitter earlier, we'll search it. We'll see how to use it.

Sneha Mehra (00:22:47)  
There might be some examples, there might not. So in this meeting, I don't want to really go into the finding the documentations, but basically it should give a pretty good idea about what it expects as inputs and what it returns and how to use it. So ⁓ and it's it's pretty simple. You just create an object ⁓ of ⁓ like this. ⁓ text a splitter is basically the the object that we create. And then we pass the chunk size that we want, like.

how how we want this ⁓ object to ⁓ it's a hyperparameter to chunk the text into a smaller pieces. So chunk size, we can set it to 300 to start with. And then also chunk overlap.

equal 30\. Now if I print text splitter, it's just an object. It's not gonna do anything yet. It's just lang chain text splitter object.

⁓ all right, so we have this. Now we just need to chunk it. We just need to ⁓ apply it to the text that we've loaded. ⁓ so for that, we again if you go to the split documents that this provides ⁓ and search it here, it expects a list of documents and it would just apply it, apply, apply the chunking. So here again you can specify and overwrite the ⁓ the chunk size and chunk overlap.

And then as the input, I believe it expects ⁓ a list ⁓ or for create documents and also for so basically it cre it expects a text and then it goes over all the items in the text and then chunk them. And just ⁓ to briefly review this ⁓ chunking example in the lecture we have.

Sneha Mehra (00:24:38)  
⁓ here. ⁓

Sneha Mehra (00:24:43)  
⁓ there is some noise here. Give me one second, I close the window.

Sneha Mehra (00:25:01)  
Apologies. So ⁓ so here ⁓

Sneha Mehra (00:25:11)  
I believe there was some ⁓ real example of chunking.

Here is the indexing, then document chunking.

Sneha Mehra (00:25:26)  
All right, here. So ⁓ this is the text. We just specify chunk size and overlap, and then it just simply breaks them into smaller pieces. So this one item would become four items, and then there is some overlap. You see, this one starts from form of while the previous one ends with form of. So form of is the overlap. So just just a brief review. So now here we can just have our chunks ⁓ easily. We can then do

Text splitter dot split documents. We ⁓ and then pass the Radox. Okay. Now my eight, like the eight pages for four, ⁓ so we had four items in this Radox, and then in total eight pages. Now ⁓ that is converted to 42 chunks. And each chunk again is nothing but just piece of piece of text. So if I print a chunk, let's see what is inside.

So again, page content and then the content. ⁓ Not fully, just part of the content. ⁓ It is probably the size is 300\. So that's it. This is this part. Now we are good. We have the chunks. Now, now that we have the chunks, we need to build a retriever. For building the retriever, first we need to load the model so that it maps each of these chunks into an embedding ⁓ vector. So again, this is.

⁓ this part. We load an embedding model. Each chunk gets converted into a point into a high-dimensional embedding space. ⁓ So you can think of something like this. I mean in practice we can have images and we can also embed images into this embedding space whenever there are some images in the PDFs or in our external knowledge base. But again we are not dealing with images now. So it's just text. But basically each text gets mapped somewhere in a

Just so that we can just re represent them in ⁓ as a as a vector. So ⁓ so ⁓ two steps. We load the model and then we map each of these chunks into an embedding, and then we put them somewhere, store them. So let's just go over it. There are a bunch of different embedding models. If we search ⁓ embedding models, it's actually ⁓ a very well known term in machine learning.

Sneha Mehra (00:27:53)  
And you see a bunch of reliable websites, they are having embedding models. ⁓ for example, opener has its own embedding models which you can ⁓ use for searching, clustering, recommendation. Again, because these embeddings is a mechanism to convert some ⁓ struct unstructured piece of data like text or image into a meaningful embedding space. So for example, if two text are similar or two images are similar, their embedding vector would be nearby.

So it's going to be useful whenever you want to, for example, recommend relevant items. You can just convert everything into the embedding space and then ⁓ using some embedding model. And then for any given item, you can just find what is nearby them. And then those are probably most relevant items. ⁓ So here you can see like different ⁓ embedding models that they're having. And then ⁓ basically you pass something, and then the output would be just an embedding. And embedding looks like this. It's just

A bunch of numbers. So it's not really interpretable by humans, or we we won't understand it by just looking at it. It's just ⁓ learned embedding space by some model. Here it's text embedding three is small. And there are a bunch of other models. ⁓ better, bigger, more expensive models. I don't know if it's listed here, but this is embedding models from OpenAir, and there is also an embedding model ⁓ leaderboard, I think. And I think either

Byte dance or Googles is probably right now the the highest, if I remember correctly. MTA leaderboard.

Sneha Mehra (00:29:32)  
All right. So yes, the Gemini embedding model is ⁓ right now ranked one ⁓ compared to all other embedding models. And then we have this coin three ⁓ embedding models, 8 billion, 4 billion. And you see these embedding models are much, much smaller than ⁓ LLMs. It's just because the task is very simpler. All it's all they need to do is they need to learn how to map inputs to some embedding space. So for example, the best like the the second

Best embedding model out there just has 8 billion parameters, which is significantly, significantly less than LLMs. Gemini is unknown. We don't know how many parameters they have, but again, yes, you can see a bunch of different embedding models. ⁓ And typically, as you go down this list, you would end up with a smaller, less capable models, but more efficient, like GTE-based, GT small, and so on. All right, so ⁓ this is about embedding. Now we want to

just load one of these models ⁓ and ⁓ play around with it. So our goal is to we have this embedding vector which is an empty list. Our goal is to convert each chunk into an embedding and just put it in this embedding vector list.

⁓ so ⁓ let's now let's now use the sentence transformer embedding that we ⁓ imported earlier. So here one of the libraries that we've imported was this sentence transformer embeddings and it's also offered by lang chain embeddings. ⁓ plugin face also has its own. So ⁓ we'll just ⁓ stick to lang chain for now. So ⁓

We we simply ⁓ have like embedding model where it could be anything, ⁓ like embeddings equal this, and then how to use it.

Sneha Mehra (00:31:30)  
Sentence transformers embedding. You will see some examples here. ⁓ We can load it with some model name. Then we pass a text to it. But basically, we create a text and then we pass this text and use the embed query method ⁓ to just get the embedding. And now if we print this query result, we would see a bunch of numbers. So let's just follow that.

So the model we will be using was ⁓ GTA small. So I have this model name here from before model name equal this. Now if I run this, it should work. It should just load the embedding model. It would take a while, depending on how big your model is. This is relatively small, so it should be fine. But if you

load try to load bigger models, it would take longer just to load this pre-trained model into your memory. Now if I print this embedding, I don't know what would happen, but let's just do it. ⁓ let's just actually do it. Let's just move everything to the next cell so we don't have to keep loading it because it takes seven seconds. So now if I print this embedding model

⁓ it's basically just showing me ⁓ some details, so ⁓ nothing specific.

Sneha Mehra (00:32:56)  
One thing that can be useful is this mass max max sequence length. So basically the context window of this embedding model is 512\. So if we pass sequences of longer than 512, it would probably not work effectively. So that's an important thing. I mean, we should be good because our ⁓ chunks are smaller. But ⁓ again, this is just something to keep in mind. So we have the embeddings. Now the next thing we want to do is to actually embed an example. So let's just have a text.

like hello world and then do ⁓ embedding vector equal embeddings.embed query and then pass the text and then print it.

Print embedding vector.

It should just again, it takes a while because now it's going through this model, which has neural network it is a neural network, it has some layers. And then it would just embed this text for us. So this is the embedding. We can also print the the ⁓ the length of embedding to see what is the what is the this high di like high dimensional space, which is 384\. So this model that we loaded, GT small, converts inputs of maximum five twelve sequence.

Two embedding vectors of size 384\. Great. Now we have tested it, it's working. Now it's time to go over our chunks and embed them. This was just an example to understand how embedding works and how we can really embed. Again, for the actual ⁓ task, which is converting all the chunks and creating fail creating the embeddings, we won't do the

Sneha Mehra (00:34:43)  
Like this kind of thing. We won't do a for loop over chunks. I mean, in theory, we can do something like this and then chunk everything and then ⁓ append it to embedding vector. But these are not gonna be efficient. So we are not gonna do this. ⁓ we're just gonna ⁓ start using the face ⁓ here, which we which we imported. So this would just take care of most of the things that we want. Read read the read the chunks.

Convert them into the embeddings. We just pass the model and just ⁓ it also gives us a retriever object. ⁓ I'm going to show you in a second the documentation. So let me first ⁓ here. Okay, so this one is done. ⁓ This one let me ⁓ let me clean this a little bit and make sure we have everything in one cell. So again, embeddings, just our test example and printing it. ⁓

This one let me delete. Now here is the the actual ⁓ part which we use face to embed the chunks. ⁓ so the first step is to use ⁓ to convert the chunks into the into the embeddings. So face, line chain face.

Sneha Mehra (00:36:05)  
They have from documents or so here you can just go over this. They have from text and then they convert text. You pass an embedding model, it will just return ⁓ the output in a certain format of those embeddings with some ⁓ like you know hashing mechanisms. So all these are just for efficiency purposes. This is how face implements ⁓ its vector database just to keep things efficient. ⁓ so let me do.

From documents.

Sneha Mehra (00:36:48)  
So face from documents, we pass a bunch of docs and the embedding model. And it would return us ⁓ the database. ⁓ it's it's very simple. So let's just do it. ⁓ all right, so here ⁓ I'll delete this. For now, I'll comment this so we don't get errors. Then I do vector database. I'll do face dot from documents. I'll pass all the chunks.

That we had from before ⁓ and the embedding model that we loaded. So let's first run this. Okay, it it ran successfully. Let's also print this to see what we would see. Again, it's just a lang chain object. At this point, it's nothing specific. It's just use this embedding model and this ⁓ chunk. ⁓ now we have this vector database. Efficiently, we were able to go from chunks to embeddings in just one line.

And it's very scalable. So if we have like millions of chunks and much, much bigger embedding model, it would still be ⁓ relatively efficient compared to other models. ⁓ one thing I will also want to say is that face is not the only library. ⁓ so I think Google has this scan. ⁓

Sneha Mehra (00:38:08)  
And this is the code, and it's basically the equivalent of face, but implemented by by ⁓ Google engineers. It's a scalable nearest neighbor. So all of these are doing the same thing. You pass some model, they convert inputs to embeddings, and then you can query it. You ask for nearest neighbor and it would efficiently run nearest neighbor. And here you would see some other ⁓ like candidates for using like libraries. A scan is here and it seems it's very ⁓ accurate and scalable.

I think they have also certain version of face IVF. ⁓ but yeah, just wanted to say there are there are multiple options. Face is what I've seen mostly used in practice in companies. ⁓ but there are also other options. So okay, so we have this vector DB. We should we are good. The next thing is to set up the retriever object because I mean the whole point of using face embedding this is to retrieve later ⁓ and query ⁓ query it, query with some.

Que ⁓ basically ⁓

Sneha Mehra (00:39:17)  
Our goal is to have this in our vector database. And whenever there is a query coming, like let's say ⁓ like a red car sentence, it would map it and find the most similar ⁓ points nearby. And then it can return ⁓ the original content of those embeddings. For example, returns this red car if it's image, or if it's some sentences return those similar, semantically similar sentences. So that's why we create this retriever object.

And again, I am to be honest, I'm really I have to search. I mean to know like what how can we really create a retriever from face implementation in lang chain. So a lot of these things has to be ⁓ Googled and we just they just keep changing and we I keep forgetting. So just want to share that this is just a normal process that's ⁓ from what I've seen, everybody does on a day-to-day basis. So now the retriever is basically we need to ⁓ do this vector db. ⁓

dot ⁓ as a retriever.

Sneha Mehra (00:40:24)  
⁓ and then then we can sh send some ⁓ hyperparameters. One important hyperparameter is how many, what is the ⁓ like this k value, like how many ⁓ objects we want to return this retriever whenever we query it. So here I can say like ⁓ search kworgs ⁓ equal k eight. And then potentially you can add more. ⁓

if there are more ⁓ hyperparameters that you want to set. And also let's search face and as retriever to get a sense online.

Sneha Mehra (00:41:06)  
⁓ as we create our face, let's see what it what we'll have here.

Sneha Mehra (00:41:18)  
⁓

Sneha Mehra (00:41:33)  
But basically, you'll just see a bunch of different methods here. This is the as retriever part. And here basically it's saying that you can send a bunch of ⁓ things, and one is k, which we already sent. And the default is four. So if we don't pass it, it's going to return four ⁓ most similar items. ⁓ and then you can also pass other other stuff. All right, so ⁓ we have this retriever now. Let's run it to see if there are no errors. Great. So

We embedded everything. We created this retriever. Now this retriever, ⁓ this vector DB also gives us some ⁓ information. So if I print this, it would say vector restore with forty-two embeddings, which is consistent with the number of chunks that we had here. ⁓ and I think

So let's see what are different pieces that vector DB can give us. ⁓

Sneha Mehra (00:42:39)  
This would probably just show the all the embeddings. No, this is the the model name. ⁓ and then what else? I mean, there are a bunch of things that you can ⁓ later you can take a look, like similarity search. And then if I print index, then you would see ⁓ here we had n total, but there might be a bunch of other things like ⁓

Sneha Mehra (00:43:06)  
Like metric type.

So again, just just so you can explore some of this later. But for now, we are good. We have our vector DB and we have the retriever. So now we can switch to the next part, which is now building the generation agent ⁓ engine.

So here is where the ⁓ LLM is needed because now we have a retriever. We already indexed all our external knowledge. Now what we want to do is we want to load an LLM so that whenever there is a question coming, this question goes into the LLM and the LLM also interacts with the retriever so it can find the most relevant embeddings and their corresponding chunk and return those and include those as part of the ⁓ context.

For this L. So this is what we are gonna do.

⁓ so this was missing.

Sneha Mehra (00:44:06)  
All right, so we'll start with Gemma three, just because it's ⁓ a small one billion. It may occasionally fail if the queries became become complicated. For example, I saw some examples in the portal that you know when we run or when we try to extend this with more documents or we ask more complex questions, we see results that are not necessarily useful or meaningful. That's very likely because we are using a small model. As soon as we switch to bigger models with more parameters.

Those issues would go away. So ⁓ that is just something to keep in mind. A lot of times, a lot of issues can be easily fixed by just switching to a more powerful model. There is just this trade-off. As we switch to bigger models, there is just more cost and it takes more time to process and generate the output. There is also this library that Olama provides.

So you can see all these models. This is the OpenAIS latest OpenBait model, GPT OSS. There's this Queen 3 VL, which is like super powerful. ⁓ this dipsy carbon, Gemma 3, which is what we are using. And actually, Gemma 3 has a smaller version, which is even more efficient. I think this version can be also run on ⁓ on edge devices. I'm not sure. This is what I remember, if I'm not mistaken. But this is surprisingly small and does ⁓ quite well compared to its size.

But for this project, we are just keeping the 1B. But you can also play around with 4B or 12B if you are on Google Colab or you have GPUs. And there are also a bunch of other models, ⁓ like Lama versions, if you want to try them. ⁓ And you know, like I think it's Microsoft model. So this is just the model library that Olama allows you to load locally and use them. And as you can imagine, all of these are open source because Olama is open source. sorry, all of these models are open-bit.

Because Oloma is basically just open source and the purpose is just to allow you to load your LLMs locally and just use them. ⁓ So it's not gonna use any like APIs or things like that. And these are all the models that are already open weight and you can use them.

Sneha Mehra (00:46:11)  
Great. So ⁓ OLAMA, ⁓ what is it? It's just a lightweight runtime for managing and serving open weight LLMs locally. And it's it's very good. There is also this VLLM, which is more scalable. And in practice, when things look good, you may want to rely on VLLM. It's basically just a ⁓ an a separate implementation of the exact same thing, similar to OLAMA, but it seems it's more scalable and more ⁓ like production ready for for actual production code. So a lot of the

⁓ a lot of times, I mean, even in in in my work, we use we prefer VLLM than than OLAM. But OLAMA is awesome just to experiment and prototype and develop locally. ⁓ okay, so ⁓ OLAMA, and then you can go here and read the documentations. But basically it's very simple to use. On Mac, you you first just need to install it and you can again go here and follow the instructions depending on your machine.

I have already installed ⁓ Olama. So my my machine right now has Olama installed. Now I don't need to follow this part. I can just directly go to this part, which is Olama Serve. And basically, what Olama Services ⁓ does is we open a new terminal and then we run ⁓ Olama serve and it starts ⁓ a local server at here and it listens to my notebook. So whenever I have a

I send a query to an LLM or I have some I want to pull an LLM or something like that, it would just send it to this ⁓ local server. So let's just follow this. Let me open a new terminal.

let's make sure we are in the right location. And then let me also activate my rag ⁓ chatbot environment. And then I'll do ⁓ Olama serve. So when I print this, it's it's now running Olama. It it just launches a server. That's it. So okay, so I let this open here and then I'll I'll just switch back to my ⁓ original terminal. So this is just running.

Sneha Mehra (00:48:25)  
⁓ and now you can pull different models. I have already pulled Gemma 3 1 billion, but if you are interested now here, you can do Olama pull Gemma 3 whatever. Probably if I do one billion, it would say that it's already pulled. ⁓ let me try.

Okay, it was super fast because it was already pulled, but you can simply try other Olama Gemma 3\. ⁓

Sneha Mehra (00:49:06)  
Twenty two seventy ⁓ Okay. So if I Olama put

Sneha Mehra (00:49:16)  
Okay, so ⁓ now now it's taking some time because it's loading this new, a smaller model as well. ⁓ it's gonna take a while. Okay, and now it's it's pulled. Just wanted to show an example that how easy it is to just switch to other LLMs. But for now, we are just sticking to this. ⁓ and again, this is still running. ⁓ then I

Then I did Olama pull this new model, it sent a request to this server. So here you can see there was a request and then there was this two hundred code with a post request API pull. So everything seems to be working.

Now we want to test it. Basically here we want to create ⁓ this Gemma 3 model and just send the briqu send the text to it and see if it works or not.

⁓ this is also very simple to you to ⁓ the the signature is very simple that is provided by OLAMA. You just use this OLAMA, which is previously imported. We pass the model name, which is Gemma 31 billion. So one thing to note is that now if I also do 27EM, it would work. But if I ⁓ enter something that is not previously pulled, it would ⁓ fail. So

I need to first pull whatever I want to use. But for one billion, they are good. It's loaded. I just ⁓ enter it here. And I can also add more hyperparameters if I'm interested. For example, one of the things that we learned was temperature. ⁓ so here I can just set temperature to 0.1. So what that means is that it loads this model and then set the temperature to 0.1 for generation. So basically we are telling it to be ⁓ don't be creative. Just just try to generate.

Sneha Mehra (00:51:03)  
The most expected consistent output. So let me run this to see. Okay, it was super fast. Now we have this LLM object. If I print it, it would just be an ⁓ Olama object with this model name. Now I can ⁓ send some requests to it. ⁓ for example, hi, what is your name? And then I'll do ⁓ output or response equal.

LLM.invoke and then I pass this text. It's as simple as that. And then here I print the response. So let's see what if Gemma3 knows what's its name.

Hello, I'm Gemma, a large language model created by Gemma team at Google Deep. Okay, that's that's impressive. Now again, the thing is when we switch to smaller models, ⁓ the the quality would degrade slightly. So for simple prompts like this, you may not notice it, but ⁓ for more difficult prompt it would be noticeable. So I just keep the one billion. I'll it's already tested, it looks good. ⁓ now the next thing is okay, so now we have our ⁓

Embedding, retrieval, and LLM. Now it's time to just ⁓ put everything together and build a rack. Another ⁓ a few parts to it. So one is we define a system.

One is we define a system prompt. I included some system prompt in this project ⁓ just as a starting point, but feel free ⁓ to keep playing with it, change it, and see what are the strengths and limitations as you make changes. So what I have here is just I'm ⁓ following those prompt engineering techniques to some extent and then creating this system template. And I say you are a customer support chat, but you only you use only the information in context to answer. So this part I'm providing is.

Sneha Mehra (00:52:59)  
Because I hope that the model can ground its responses to facts that are provided in our ⁓ knowledge base. I also say that if the answer is not in context, respond with I'm not sure from the docs. This line is very important. If we remove this, the model may hallucinate more. Basically, when it doesn't know the answer, or when you ask a question that is not in the context, it would it may just come up with something ⁓ random. So this line is here.

Aiming to ⁓ push the LLM to hallucinate less. The model would still hallucinate, but it's just probably gonna hallucinate less. And then here I'm specifying some roles. Use only the provided context to answer. If the answer is not in the context, say I don't know based on the retrieved documents. Be concise and accurate, prefer coding and then a bunch of stuff. When possible, cite sources as this. ⁓ and then here I'm just sh

This is not that this is a template. This is not a system message. I'm creating a system template, and later I'm going to rewrite parts of it. ⁓ Then I have the actual context because the context would be known ⁓ at runtime in a dynamic way. It really depends on what is the user's input query. When the user's input query comes, first I use the retriever to retrieve relevant documents. And then those are gonna construct my context. So then I have to replace this part.

⁓ with the context.

⁓ and then this is the question. It's just here this part I'm going to just replace it with the user's actual question at runtime. Now ⁓ let's see. So ⁓ great. So ⁓ one so this is gonna fail. So one thing there was ⁓ originally a mistake, but I kept it so we can go over it in this session. Is this ⁓ so Python has this this format that

Sneha Mehra (00:55:01)  
Then you place things in this ⁓ like in this ⁓ in this format.

You can later format your system template and replace this with what some inputs. So the mistake here was that ⁓ our goal here was just to show an example to the LLM that when you want to cite, site like this, like open bracket source, ⁓ and then the actual source that you found. But then mistakenly I had this. So later, when we want to format this template, it would also look for ⁓ input sources, which we don't want.

Basically, that is going to be anyway ⁓ obtained from the context. So the LLM would figure it out from the retrieved context. So that is why ⁓ it was not working before. If we include this, ⁓ okay, first I have to implement the first set. So for now I'll just remove this. And then here we are just showing it as an example. Like we are just showing it like this is how we want you to output.

Sneha Mehra (00:56:04)  
Alright, so ⁓ this is this part. Let me run it. Great. Now we have a system template. And these two parts are still pending. ⁓ They has to be replaced. Now the next part is to ⁓ create our chain.

So ⁓ there are mu there are bunch of steps. So we have three steps. Let's first focus on the prompt template.

Again, the prompt template is just an object and it's a template. It has to be formatted properly.

Sneha Mehra (00:56:45)  
So if we search lang chain ⁓ prompt template.

Here we can see how we can ⁓ use it. So first we create a template, we say say foo, and then later we can pass a value for this part. So it would just take care of it using this format method. So let's just follow the same thing.

So ⁓ I do format. What it expects is context and ⁓ context and question. So let's just do ⁓ let's just specify input variables equal context ⁓ and questions. And next we also need to pass the actual ⁓

a a template so ⁓

Sneha Mehra (00:57:52)  
Okay, so we are now creating this prompt rank ra lang chains prompt temp prompt template object. And we are just passing this system template and we are telling it, hey, these two items are what you should be looking for later. ⁓ and sorry, this is ⁓ okay, so it was my bad. So ⁓ this should not be the formatting part. Here we are just creating this prompt template. So let me do prompt equal prompt template. I create this object.

Now if I print prompt, again, it's going to be a prompt template object, ⁓ which exactly knows how it should format things.

Sneha Mehra (00:58:46)  
Okay. Now we have this prompt. It's basically a prompt template object. ⁓ It has a template and it knows how to format it. Now I can simply format this. ⁓ I can say prompt dot ⁓ format.

Sneha Mehra (00:59:06)  
Hmm foo equal bar. Here we say context equal

Sneha Mehra (00:59:21)  
No ⁓ like no refund policy, for example. And then questions. Again, let's just do what is the refund policy. Now, if I run this and let's just keep it here, format formatted equal prompt art format. Let's see if it works first. ⁓

Sneha Mehra (00:59:53)  
West Shen

Sneha Mehra (00:59:59)  
Okay, great. So now I have this formatted object. If I print it.

Sneha Mehra (01:00:15)  
It's basically having my ⁓ initial system prompt and it has probably replaced those two parts. So let's see if that's the case.

Sneha Mehra (01:00:27)  
Impossible site sources and then here it says context. Perfect. It now now it it basically replaced with this no refund policy. So this first n packet slashes is the is the new line. So ⁓ basically it starts from here, which is consistent and expected because I provided this, and then user prompt. So this is how prompt template can be really useful. We just create it once and then at runtime we can just replace based on the retrieved context and question.

Alright, so we have prompt template. We followed this line and we have a good prompt template. Now the next step is to initialize the LLM. So this one we already saw here. Let's just ⁓ copy paste it.

Sneha Mehra (01:01:22)  
Okay, so so far I have an LLM. I have a very nice prompt template. The next thing is ⁓ now now I can create my chain. I all also remember we have this retriever from before. So retriever is if I print retriever. ⁓

It should be there and it has all my embeddings. See, it has all the face embeddings. So ⁓ I have all the pieces that I want. Now I just need to just ⁓ chain them. ⁓ And again, Langchain provides all those kinds of functionalities. For this purpose, we have conve ⁓ conversation retrieval chain. Again, many of these things keep re ⁓ evolving. So ⁓ some of these may get.

deprecated and new formats or new ⁓ methods may get introduced over time. ⁓ But as long as we use the same consistent condo environment for divergence, it should work. So now we have conversation retrieval chain here. Let's see if there is an example to to show us how we can really chain stuff. ⁓ So I think if I search from ⁓ LLM, this is ⁓ this from LLM is the method that we want.

We pass an LLM to it, we pass a retriever object, and we also pass the the prompt template, prompt template. And then we have then suddenly a chain. Basically, we go from these ⁓ individual pieces of rag. Let me.

Sneha Mehra (01:03:01)  
We suddenly go from each of these individual pieces to something ⁓ that has all of this. So basically, ⁓ when we chain ⁓ our LLM and retrieval and we pass the prompt template, we are ⁓ basically creating something like this. And now this is called the chain that we are we are creating. Now chain is ready to be used. All we need to do is whenever the user query is known, we pass it to this and it will just take care of everything else.

So we'll we'll see an example.

Let's get back here. We have Olomo. Now let's create our chain. Chain equal ⁓ conversation ⁓ conversational retrieval chain. Dot from LLM. Now we just pass all those pieces that it wants. Retriever. ⁓ And then there are a bunch of other things that I would pass.

We can, I'm not gonna go into the details of those. We can you know, offline you can search some of these and see. They're minor, so I don't think it's gonna add too much value. ⁓ but basically just adding providing the arguments here. I do prompt, prompt, I pass the prompt template.

Sneha Mehra (01:04:26)  
Mm.

Sneha Mehra (01:04:30)  
And then finally ⁓ return source documents. I mean you can guess what what what is going on, but it's just minor. I mean I'm I'm asking this chain to also return the source documents that you found. So for for site site citing purposes. So ⁓ let's run it and see if there are any errors. All right, there is an error.

Hmm, ⁓ I have the code ⁓ somewhere for this purpose. Let me just copy paste it. It's probably something very minor.

Sneha Mehra (01:05:08)  
Okay, so ⁓ now we have this chain. Let me print the chain.

Sneha Mehra (01:05:17)  
It's just a chain object. And verbals is false. We can set it to true. So we get more info ⁓ as we use it. But for now, it's it's false. Great. So we have this, we have the chain. Now let's validate it. Let's have a test question. And then in our test question, ⁓ for example, I I put some questions here. If I'm not happy with my purchase policy or refund policy, and how do I start a return? How long did delivery take?

What is the quickest way to contact your support team? Which we saw some of those in these documents. So now we want to see if the model can really ⁓ retrieve those and generate some responses, grounding to those. So, okay. So here there are three steps. One is ⁓ initialing an empty chat history list, and then we just look through this and just keep passing it to the chain. And then everything else would be just taken care of by line chain. ⁓ let's do it.

I create this chat history here.

Sneha Mehra (01:06:22)  
Then I go for all the questions. ⁓ Test questions. Test. Test questions. ⁓ Then I can simply just pass the question and the previous chat history. And then the model should be able to answer it. So ⁓ basically I have this chain and I can simply pass any question to it. So I'll do question.

which it is expecting. This question that is expecting is what we ⁓ included as a as what this prompt template expects. It expects this question.

⁓ so here I have question Q. Let's see if this works or not. ⁓ just very simple. Print chain. ⁓ and let's just break to see if for the first question it would work or not. Missing okay. So it it really requires the chat history. ⁓ so I also need to pass the chat history.

Here it's also very simple. We can just pass all this as a dictionary, chat, history. Chat history. This is basically we just keep track of all the conversations and each time we just pass it. ⁓ let's just see if now it works.

So it should now produce an answer for the first question. If I'm not happy with my purchase, what is your refund policy and how do I start a return? So while that is thinking, let's see if we find it here or not.

Sneha Mehra (01:08:15)  
Hm, it's it's not even return policy.

Okay, 30 days of delivery for a refund for a free size exchange. So let's see. ⁓ so okay, it ran, it took 18 seconds just because my machine is not ⁓ strong enough. Probably if you use ⁓ more powerful machines like even on Google Call app, it would be much, much faster because this model is relatively small. But let's see the answer. The answer is okay, here is your refund policy and return process.

Refund policy is as follows. No return unless defective. ⁓ If your item is defective or you're not satisfied, we'll issue a refund unless the item is defective. okay. So so this is this is included here. ⁓ three-day ⁓ return window. You have three days from the date of delivery to ⁓ initiate a return, refund timeline. So it's basically ⁓ it's kind of correct. Again, if you switch to better, bigger models, it would be much, much better. ⁓ but it seems it's kind of

working. So now let me remove this break and then ⁓

I just make it a little bit cleaner and make sure that each time we get an answer, we just append it to the ⁓ to the chat history. So let's ⁓ delete this, let's do result equal ⁓ chain. let me not then delete it. Just ⁓

Sneha Mehra (01:09:46)  
Result equal this. Now chat history dot append. ⁓ then we'll append the question and the result. So basically both the question and the answer will just append it to the chat history. ⁓ I'm gonna print it soon to to get a sense of how chatting ⁓ what chat history looks like. Now we have this, and then let me I have a nicely formatted print a statement.

So now now let's run it one by one.

Sneha Mehra (01:10:23)  
gonna take probably a while.

Sneha Mehra (01:10:32)  
Okay. The question was if I'm not happy with my purchase, what is your refund policy and how do I start the return? Okay, here is the refund policy and the process based on the provided context. This is nice. This is what we asked in the system prompt and it seems it's following it. Refund policy. We offer refunds for defective or baronity re related items. With these so this part was hallucination, I think. There is no ⁓ mention of twelve. Probably it's just hallucinated. The model is not strong enough.

If you are not satisfied, you can initiate a return within the three days of delivery for refund. So this was this part is correct. How long will delivery take ⁓ the answer? And what is the quick quickest way? You can contact our support team through ⁓ chat at ⁓ this. So this is not also good. Let's see if this example is is good. actually, this is this is kind of okay. I think it's just ⁓ providing the timeline, the times, and then the email. Let's check if to see if this email is correct.

And also the phone number, four six five five five five.

Sneha Mehra (01:11:36)  
four six five five five zero one nine nine. Zero one okay so it it's close. ⁓

It's close. But again, this is just a model issue. As soon as you switch to bigger models, 8 billion or even more, all of this would go away easily. ⁓ So that's it. Basically, now we have a rag. It's just a chain object. So whenever we hear about like an application that is rag based, it's just simply a chain object if it's using Langchain library. Even if it's using other libraries, it's simply, it's just usually an object. And this object is holding a couple of things.

One is it's holding an LLM for the generation piece. It's holding a retriever, which is already having all the like a dictionary going from our external knowledge document like chunks to embeddings. And then ⁓ a prompt template. So that's that's what rag is. Rag is this this guy. ⁓ and then in order to use rag, we just send the question to it and we keep tracking this chat history. ⁓ let me also print the chat history to get a sense.

Basically, the chat history is just having everything. So at first, it's just the question and answer. Then it has ⁓ then we added this question and answer, and then this question and answer. So each time basically it's keeping track of all the previous conversations. Now, depending on the use case, this part requires a different way of implementation. You typically want this chat history ⁓ in each session for each user. ⁓ when a user opens a new session.

We don't want to maintain the chat history. ⁓ Like put simply, I mean, we can we can come up with more complex scenarios like defining, like ⁓ like ⁓ introducing memory to this system. But I mean, just if you think about this in simple terms, chat history can be useful in an ongoing conversation. So we maintain it per user per session. ⁓ that's it. The next part is optional, it's an ⁓ extremely UI.

Sneha Mehra (01:13:48)  
Streamlit basically just a library. It's very simple to use. It's very handy for prototypes. It can be quite useful. You just pip install it. And then it gives you a very simple way for creating a very ⁓ non-production ready UI. So a lot of times I use it just to show or demo or prototype my work. ⁓ again, if you're interested, you can just go over the playground or docs.

And get us started on the implement on the setup. And I think there should be some concepts and quick guides. But I'm not gonna go over it. It's just a UI thing. So I've already ⁓ have a code for the streamlit. ⁓ I am going to just paste that. ⁓ And the when this release the solution right after this session in the GitHu, you can go over the code and just look at it.

⁓ so ⁓ it's basically a bunch of imports here ⁓ and some UI related stuff, setting the ⁓ title and ⁓ page config for the pay for the ⁓ URL when we open it. Here we init chain, and it's basically the same code. We are just putting it all together. So init chain, we create this vector t db, we specify the model embedding model. ⁓ and then we create this retriever, then we send this LLM.

We create this memory. This is ⁓ you can look it up online if you are interested, but just as a way to ⁓ introduce this memory to the system. ⁓ and then again, I'm just creating this chain and returning it. ⁓ and then here I create this chain. Again, if history is enabled, ⁓ these are just some ⁓ extreme lit related stuff. And then here is some ⁓ initial streamlit ⁓ component which is chat input.

So we'll see in a second when I run it. And then when the user enters the question, it just shows thinking. It sends under the hood, it sends the actual question and the chat history to this chain that was created. It gets the response back. Once it has the response, it manages its history and then it shows it ⁓ to the user. So now this is the code. I just put this code in an app.py. Let me run it. So by running this cell, it just

Sneha Mehra (01:16:11)  
Brings this code into app dot pie and saves it, I think, here. So we should see it here.

But this is just ⁓ for demonstration. I mean practice you can just open an app.py file and just start implementing your streamlit code. ⁓ okay, so we have this. Now the next thing is to run the streamlit. And running it is also super simple. ⁓ I want to keep this open, the OLAMA, so ⁓ so that the streamlit can use it. So I'm going to open a new terminal. I'm going to activate my ⁓ environment because it's

Extreme lit is installed there when I activate rag chatbot.

Mm. Was it rag chatbot? Rag chatbot?

And then it's as simple as a streamlit run app.py. Let's see what happens. Okay, so it ran a local server. It's trying to ⁓ initialize the chain and loading all those models. So it's it's gonna take a while. So here it says it's running in a chain. ⁓ all right. ⁓ And there is an error. ⁓ and I I have ⁓ no idea. ⁓ I think I have an idea. ⁓

Sneha Mehra (01:17:28)  
⁓ Okay. So let me see if I can quickly fix this. If not, the solution that I would upload has this fixed. But I think I can quickly fix it. ⁓ so I think the problem is we didn't save ⁓ the embeddings. So the streamlit is looking for those embeddings, but it's not finding it. So we missed this part, which is fine when we were running it because we were using this object anyway. We had these chunks already here, and then everything was fine.

But ideally we should also save it. So for saving it, we can write vector db.save local. And then here it's just saving the ⁓ I can just call it face index. It's basically just saving the index or the embeddings that we built using all the chunks. So now if I run this, you would see a new file here ⁓ called face index. ⁓ And it's just a way ⁓ for ⁓ the

Lang chain to later, if they want, if it wants to restore our embeddings, it can just load this file. Again, this example is very small with just four PDFs. It doesn't take too long to just rerun everything, rewrite rerun this indexing. But in practice, there are millions of documents. So these steps can take quite long. So it's always good to save them once we run them. So then moving forward, we can just use this file. Now let me rerun this.

Okay, now it's saved. There is this face index, and then inside it we have index.face and index.pickle. So now let me ⁓ rerun my streamlit and see if we were actually able to fix it. Let me also ⁓

Sneha Mehra (01:19:16)  
Okay, it if it doesn't work again, I don't know ⁓ what the issue is, and then but in the solution that I would release, it would work. Okay, so the UI seems to be working. The chain was ⁓ initialized, and now this is this is the UI. So we have customer support chatbot, and then here is just a box, and then it says what is in your mind. So I can say, ⁓ I what is your name? Now, in an ideal setup, if the LLM is quite powerful.

It should not respond to this and it would say ⁓ something. And the reason for that is because we explicitly in the system prompt we ask the LLM to make sure that whatever you generate is grounded to the context. And the context only has information in the in those PDFs about our retailer store. ⁓ but let's see if Gemma understands this. It may or may not. Let's first see if the UI works.

Sneha Mehra (01:20:14)  
Okay, that's that's great. That was actually ⁓ that's awesome. It's not answering that. What is your ⁓ refund policy?

Sneha Mehra (01:20:39)  
⁓ okay that's ⁓ basically it under the hood it retrieved the items, then it passed to the LLM, and then the LLM generated this. ⁓ It's it's it's probably hallucination. Again, you can experiment with bigger models. I'm sure many of these would go away. ⁓ I think I've tried this with Lama 8 billion, and then the results are much, much, much better. So in case it would give you a data point to to try different models.

⁓ so yeah, it's the UI is working and everything seems good. So I guess this was the last cell and we can ⁓ wrap up. ⁓ so yep, congratulations. And now we built this this very simple rag-based system. ⁓ all right, so let me take questions. ⁓ there was one more thing. ⁓ actually, I I project three is also released. ⁓ so

feel free to start working on it. So this week is going to focus on agents, workflows, and function calling. So it's going to be very exciting and perhaps the most ⁓ it I would say this is it's gonna be my favorite week, like the the upcoming week, ⁓ because the content is super relevant in many of the startups and applications or building systems. So ⁓ yeah, the project is released. I hope you enjoyed the content and guided material as well. So next I'm going to take questions.

Sneha Mehra (01:22:06)  
So let me ⁓ start ⁓ Syncons. Hey, Alec, quick question. ⁓ so ⁓ the model you can learn with the PDF documents, what about an RDS database, ⁓ like a SQL database or No SQL database? How can you make the model learn from that? ⁓

Again, ⁓ each of these are basically there are different methods to convert in general unstructured data to embeddings. And many of those are embedding models. I mean for SQL, I'm not sure, but I feel if I search like embedding model for SQL, let's see if there is anything. I've I remember I've seen some. So see, there is SFT SQL embedding. So basically, the ⁓ the recipe is the same. It's exactly the same. All you need to do is

As the first step, you need to make sure that whatever data you have in whatever format or unstructured format, you convert it into first you load it and then you convert it into some embeddings. So most of these different ⁓ formats have there's some very specific models trained to embed those. ⁓ So I assume this is one of those. ⁓ and there might be some others. So you can search those. For image, you know, like image embedding models, you would see also lots of different embedding models.

to embed images. ⁓ So, but but everything else about it is the same. Basically the recipe is the same. As as soon as you get the embeddings, the rest of it is just the same. You don't have to worry or do anything. You just chunk it, you pass it to your embedding model, you get the embeddings, and then at retrieval time, the retriever would automatically retrieve the portions or chunks that are most relevant to the user's query. For example, if the user asked for what are what are my last ⁓ five ⁓ transactions,

It is very likely if if all your models that you are using are are powerful enough, it's very likely that the retriever would just ⁓ obtain those parts from ⁓ your tables or whatever ⁓ and then pass those as the context to the LN.

Sneha Mehra (01:24:06)  
Thank you.

Sneha Mehra (01:24:09)  
⁓ Similar thing with NoSQL as well, right? What what was that? ⁓ the same thing with NoSQL databases as well. Yes, yes. Most of these are the same. I mean, I know there are also some works to embed tables or things like that. So it's it's it's mostly the same. There the other approach that sometimes works is to ⁓ have a model to generate queries. So basically you give access so this week you would we would we are covering those in the guided learning. You give access to some databases like actual SQL or NoSQL databases.

And then you also give a V ⁓ you give the model a VA and permission through MCP or any other protocol to access those databases. So the model would figure out what is the right command to execute to access the database and get those information back. That's another route. And it it would be covered more in this week's, like the upcoming week's guided learning.

Sneha Mehra (01:25:03)  
⁓ Harry, please go ahead.

Yeah, hi. ⁓ so my question is ⁓

Sneha Mehra (01:25:15)  
⁓ I I guess I'm I'm I've lost you. I don't know if it's it's me or

I you are also on on mute, I think.

Sneha Mehra (01:25:26)  
⁓ I think I'm not hearing. I don't know if it's only me or or everyone else.

I think I'm I'm not I'm not hearing you well. ⁓ okay, okay, okay. Yeah, okay, fine. Let me go ahead again. ⁓ So, my understanding is the reason that we encode in the first place is because we cannot do a string-based search. So we need to convert it into high-dimensional mathematical form where the search can be done effectively. ⁓ Correct. ⁓ And in the notebook that we went in the project overview today, in that section three.

Could you please help me understand what happens after encoding? So where was lost is from encoding to the vectors. How the Yeah. Sure. So this part. You remember this part. So basically we load an embedding model and we just so here is just an example. We can send any text to it. It would give us a vector. And this vector is of this size and it's just a bunch of numbers. So this is this is the embedding. So this is the embedding model. Perhaps I have to I had to use also a better name.

For this, so apologies everyone. Embedding so hello world is now represented in 384\. Is that correct? Now this hilt whole thing is now represented in a vector of this size, which is meaningful. So if you basically now send also another ⁓ vector like embedding ⁓ vector two, embeddings. ⁓

High work.

Sneha Mehra (01:26:57)  
So there would be another vector. And these two vectors, if you look at them, if you measure the distance between them in this very high dimensional space, the distance would be very small. So they are they're gonna end up nearby. ⁓ But I yeah, this is where I was kind of confused because hello world is just a small string, right? And ⁓ how can it be into a size of 384 vectors? ⁓

So this is this is basically an embedding model. This is a different topic. So these are very special models called embedding models trained for this exact same purpose, right? So basically the the architecture of this model is they accept a sequence of tokens. It's it's usually transformer-based. So it's a sequence of tokens, which is this hello world. It goes into this model, and then through a bunch of transformations, at the very end you get a sequence of outputs. And then these models only look at one of those outputs. So that become basically the embedding. ⁓

⁓ Which is one vector. And then they apply some loss and then they backward, they up update all the parameters such that whenever there are semantically similar sentences, this final last vector of the ⁓ final output layer has ⁓ is going like gonna be very similar for semantically similar inputs. Got it. So here in this example, hello world is represented by 384 ⁓ size vector.

Correct? Yes, yes. And even, I mean, whatever. Like you can it could be long. ⁓ Hello world, my name is whatever, like, you know, ⁓ and rest of it. This entire thing would be mapped into a single vector. ⁓ Vector, got it. And the logic here is we convert the question into a vector and then we search in the entire embeddings of ⁓ with the nearest neighbors, which is the closest to this queried vector, then that would be the answer. Correct? Yes.

Exactly. So ⁓ let me so this is the PDF, right? So we chunked them. So what what what it means is that now this is what for example. I wish I could select them. But this is going to be you see my my carcel, right? So this part is going to be ⁓ one chunk, let's say this is the second chunk, this is the third chunk, and so on. Each of these is going to be mapped to some vector in the embedding space. So when the user asks a question like ⁓ where are you shipping from, right?

Sneha Mehra (01:29:15)  
Now, where are you shipping from? Is going to also be embedded. Now, after it's going to be embedded, it's going to end up nearby this this chunk. ⁓ It's not going to end up near processing time chunk. So because of that, the retriever would be able to ⁓ send back these relevant chunks or most similar chunks to the user query, which is this chunk. And then suddenly the LLM has everything it needs in its context window. Got it. Got it.

So embedding here essentially is normalizing all the text uniformly so that it can be easily searched. Yes. And again, you can search for embedding models and you can read different embedding models. Some of them are not necessarily bringing everything into a single vector. They it can just be a sequence of vectors and they just somehow figure it, you know, figure it out, like taking some pooling methods and so on. But ⁓ to think about this ⁓ like to simplify, yes, everything would go get mapped into a single vector.

Yes. ⁓ One other question is what does generator do here? I get the retriever part. generator. Generator is the LLM. Basically, we need to now we have the all the relevant chunks and the user's question. We need to generate generate something, right? We need to generate the output. So that's the LLM. So the generation is just a LLM piece. ⁓ okay, okay. Yeah. Okay. Got it. ⁓ And ⁓ one last part is ⁓

Like it can be out of subject. I I would also encourage if there are other relevant questions to be taken. But I in the in the course that you have given, you all you also touched raft. So I just want to know how relevant is it, ⁓ it is ⁓ in in in in the initiatives that we are running. And do we really need to think, I mean, deep dive about it? ⁓ I mean raft is basically more a training method, right? So a lot of times

This whole system may not work as good. So in general, the rule of thumb is that we at least for smaller companies or startups that want to prototype something, the rule of thumb is to use whatever is out there, open-bed models, open source models, and just combine, put them together, see if it satisfies your requirements. And then you figure out if it's failing, why it's failing, where it's failing, what is the easiest way to fix that. But one way is basically now to ⁓

Sneha Mehra (01:31:30)  
continue training this model, this LLM that we load, basically we find the unit using this particular method designed for RAC. It's called raft. ⁓ Basically it's a way for the model to better understand which ⁓ retrieved items in its context can be more relevant or less relevant to the question. It's basically we are just continue training the LLM a little bit longer, ⁓ with a certain focus, which is figuring out relevant and irrelevant items in the context.

⁓ it's generally not not useful unless you want to build like a state of the art or you have a very, very custom data or you or use case, which you at some point you decide to train, fine-tune it. Yeah. Okay. And this is the chatbot that we saw to demo today, right? And the difference between chatbot and an agent is that agent has more capabilities to do action. Is that is my understanding correct? That's going to be covered in in this week. Yes. Okay. Sounds good. Thank you. Yep. ⁓ Basuki?

Hey LA, thanks. So thanks for time. So I have like two fundamental questions, right? Or maybe three, and I'll be very try be very quick, right? In the initial lectures, right, we said ⁓ a let's say we break down, we chunk the token by a word, right? We we tokenize by a word. The word is then mapped to a high-dimensional vector, right? ⁓ Which basically it goes through a transformer architecture, like you have an embedding and you you pass it to a single attention or a multi-attention, depends, right? And then you say it was transformed to a vector.

And then eventually there is an attention where it borrows from other vectors, and the outcomes the probability, right? Outcomes the a vector which you choose from. Right now, when we chunk ⁓ many words, right? Let's say you have a chunking algorithm which with a chunking overlap, you take all of these words, you so my like I'm trying to understand how does it map to a vector, right? Because like it's no longer ⁓ a that is like I'm assuming it is from a base model, it is pre-trained.

Because ultimately there has to be from the matrix, right? There has to be some weights it has to provide in order to generate the vector, right? It can't be ⁓ like it can't be pre, it can't be bootstrapped because it can be any random text you give. I ⁓ don't understand is it still a attention. ⁓ Yes, yes, yeah, exactly. Your your understanding is accurate. I mean, so that's why here we are loading this model. So this model, right?

Sneha Mehra (01:33:57)  
When we load it, we are basically loading a neural network. And very likely it's a transformer. ⁓ And we have all the bits for all the layers. So whenever you pass this, what happens under the hood is that after all those tokenization or whatever, it becomes a sequence of vectors and then it goes to this embedding model. And now all those parameters, each layer just keep transforming and transforming. And then finally the last vector and the final layer is going to be that final output. Okay.

Thank you. Then the last thing is right. When we like, you know, like okay, like I'm trying to play with different temperatures. I'll like, you know, maybe I'll play with and then ask you over chat. ⁓ My last question is when we do a vector search, right? I have a vector now and I go to FIS and say, give me, like, you know, from this vector a sequence of numbers, right? Give me anything that is close to this vector in directionality and in weights, right? That is how we do a vector search. ⁓ So

It can be Euclidean distance, it could be cosine similarity. What type of vector search is FAS doing, right? And how does it influence your like decoding? Like them am I am I making sense of my question? So the question is what is the what is the what is the interpretation of the ⁓ like the similarity metric and when should we use like cosine metric where ⁓ or other things, right? Yes.

⁓ Google has this very interesting course about ⁓ distance. I don't know what to search for, distance metric. ⁓

Sneha Mehra (01:35:34)  
Because there is like for a similarity search, right? It is yeah, you have cosine similarity, you have like ⁓ this, I think this is the one. So I I'll I'll share this. Okay. ⁓ or actually you can just search, choose and yes. ⁓ it's going to give you the the intuition that you are looking for. So it says when to use dot product and cosine distance and Euclidean distance and why each of those, what are the benefits of each? Okay. And then

Like in the embedding model, right? So let's say when I do a search, right? Can I influence it by saying use cosine similarity versus dot product so that I can get a better nuanced answer? Yeah. So so ⁓ that's a good question. I have to check. I have to check that. It I think it depends on how the embedding model was trained. I mean, if the embedding model is trained by taking the cosine distance, we probably want to follow the same ⁓ similarity metric. But

I am not one hundred percent sure. I I am ninety percent sure. ⁓ I can do some research or I can also ask Chat GPT. ⁓ but my my intuition is that it depends on how the model was trained. If it's trained by a cosine distance, also we you have to use cosine distance to get the most meaningful ⁓ distances between the the the items. Okay. Thanks, Ali. Thanks for thank Yeah, yeah, sure.

⁓ by Viba. ⁓ Thank you. So my question is about the citation and the metadata. ⁓ how do we how does it get stored? Is it a part of the embedding itself or is other than the embedding and part of the index? ⁓ so the one that we have here is not gonna really work, at least ⁓ in a meaningful way for for citations because we are not taking care of a lot of things. I mean, all we are doing here is we are just saying that, hey, just where is my system prompt? ⁓

Sneha Mehra (01:37:28)  
I mean we are just providing just very minimal kind of just format and then the model would have no idea. I mean the best it can do is just it can here it can just place ⁓ the the actual chunk or or the content, right? But the right way to do this is we build this context more carefully. Basically we can just build this context by making sure that each piece has like a a number associated to it. We keep track of a different dictionary, and then we also include that and here very ⁓ very clearly we specify what to do when you want to cite.

⁓ for example, if if there is a certain part that are coming from a certain part of this context, ⁓ how how should it be cited? So that's how it works in practice. I think Anthropic has ⁓ published a paper about the citations and how it works. If I find it, I'll I'll share with you. But I think if you also search online, you would you would see ⁓ from Anthropic. There there is a a blog post or paper that goes into the detail of how to make the citation work. ⁓ this is just a toy, toy kind of setup.

⁓ last time before this live session, when I ran this, it it it shared some sources. ⁓ but in a very ugly way. It just shared some like in a very ugly way. But ⁓ if you actually want to make source ⁓ this re ⁓ this thing works, ⁓ citation works, ⁓ it requires a little bit more work.

That it so that clarifies a lot of things. ⁓ second follow up question on the same one, right? So we have provided the entire document, chunked it, and then embedded into the vector database and asking ⁓ LLM to answer based on that, right? So in a way, ⁓ the entire part, all the information around that that within that PDF is already there. So if the PDF pages have all the information, what page is it ⁓ or what unit

It belongs to which chapter it belongs to what author, ISBN number, everything. In a way, everything is already being chunked and embedded into the database. ⁓ should is it a good practice to have it injected separately into the index, ⁓ other than the embeddings? I mean, probably not. It also depends on to what extent or granularity you want to cite, right? Or what if you want a specific format. I mean the

Sneha Mehra (01:39:43)  
What I have seen in practice is just to make sure that you have a very clear instructions in the system prompt saying that this is how I want the citations to be shown or or follow this setup. This is what I have seen. But again, ⁓ I am not too sure about the best way. I have seen people do it differently. ⁓ I have not also seen a proper evaluation mechanism to evaluate the citations, ⁓ or or maybe there are some papers I haven't read yet.

⁓ but again, the examples that I have seen personally is just system prompt kind of thing. So you make it very clear and as you mentioned, everything else is already handled by the embedding model.

Sneha Mehra (01:40:27)  
Derek, thank you. Sure.

Sita Kumar. Thanks, Ali. It was wonderful to see all these things coming together so nicely. Thank you. Also getting some glimpses into how to productize it. For example, VLLM best engine to host the LLMs in production. ⁓ right? So now ⁓ as ⁓ the ⁓ rubber, ⁓ I mean whatever, ⁓ actually. ⁓ so basically to make it more practical now, ⁓

One of the things is how do we debug when things start to go wrong? ⁓ here is there a way to know that the ⁓ data that is being retrieved by RAG is relevant, for example. Then if it is relevant, is the model using that? So there are all those questions that come to my mind. How do I do it? So the best the best approach that again I personally follow is first we when we build it, we evaluate each piece separately.

For example, for retrieval, first you want to s evaluate your system if if you see it actually returns relevant items before even attaching it to your LLM and creating the chain. And you know there are like systematic ways that you can ⁓ evaluate, let's say, ⁓ like retrieval systems or rags. ⁓ and also retrieval, I mean if you search for retrieval metrics, there are ⁓

A fixed set of metrics that are commonly used for information retrieval and ⁓ recommendation systems. So those can be used ⁓ to evaluate how good your retrieval is before going into the chain part. ⁓ so that's what I typically follow. I try to make sure that every single piece works ⁓ and satisfies my requirements in terms of quality. And then once we have everything else and then we put them together, then

Sneha Mehra (01:42:24)  
⁓ It also depends. Just the evaluation becomes more difficult when you create the chain because then you no longer know if the issue is because your retrieval is not good enough or your LLM is not too powerful. ⁓ so you just have to manually check. And for different failure cases, you have to see which ⁓ which which part in this entire system was create was ⁓ the cause of the failure. ⁓ Okay. And for ⁓ simpler development workflows, ⁓ does the chain line chain

Allow me to log the request response things like that or more detailed logging. ⁓ I I think Langchain from button, I know it's even used for some production systems. So I think it should just provide most of these things that are ⁓ relevant or ⁓ necessary for building. I haven't had a use case for what you're saying, so I haven't tried it, but I believe there should be there. Great, thank you.

Sneha Mehra (01:43:21)  
Sonia?

Yes, hi Ari, thank you. So ⁓ in the like in the import section, right? The first thing where you import everything, if you can scroll there, scroll up. ⁓ yeah, yeah, here. So here in this LandChain dot embeddings, you are importing open AI, hugging face, sentence transformer, but we use only sentence transformer, right? So other two are

Kind of unnecessary. Okay, got it. Yes. Thanks. Yeah, sure. And ⁓ I I couldn't get that this then if you could scroll again to this c that system prompt where you have the system prompt. ⁓ yeah, here. This context, this I understand, you know, this context is the context of the near neighbor vectors and that ⁓ yeah, near neighbor.

vectors that it from the files index it found. Now is this context ⁓ equal to you know that k k ⁓ we had that k parameter k items that the yeah it returns k ⁓ exactly it yeah exactly it returns eight because we said k to eight it returns eight most similar chunks to this question.

Okay, so that that so that that K retrieved num that ⁓ is that is the that is it that is the context. ⁓ K basically is the number of ⁓ Yeah, yeah, I get k K is that number of how many we want, like default is four and all that, but that is what is the context that what it retrieves ⁓ from the you know from the that ⁓

Sneha Mehra (01:45:17)  
From whatever the database, fra phase database that we saved, it retrieves that K items and it and it saves them or it ⁓ like what should I say it ⁓ substitutes in this context variable here. That's what is happening. Exactly, exactly. Once once it's retrieved, it's good it's gonna this part is going to be replaced by the retrieved items.

Again, this is a simple setup that we are creating. In practice, we may not want to directly replace this after we have the retrieved items, let's say eight items. We may want to still do some post-processing to make sure that some of those may be relevant or less relevant, and we may want to get rid of them so we don't confuse the LLM. So in practice, we may want to do some post-processing before actually replacing this context. But in this setup, yes, as soon as we have the eight most relevant chunks, we'll just replace this context with those.

Okay, got it. Now ⁓ my I was asking that ⁓ w where is ⁓ prompt engineering happening? You know that in the in your that Excalibur diagram in that generator, you say there is prompt engineering and LLM. Now just here if you could go over where is that specifically happening? Yeah, that's a good question. I mean, it ⁓ prompt engineering is is a little bit the term is vague.

I mean, there are all those different techniques, but in practice, at the end of the day, what you would see is is a system prompt, right? Basically it's a system prompt that a prompt engineer engineered it. ⁓ and you don't have to necessarily always use all those techniques that we're discussed. For example, in this case, we are not dealing with like mathematics, right? So there is no point to ask like a chain of thought ⁓ technique in this system prompt. So I mean I could have said like think a step by step and things like that here, you know, think a step by step, but probably it's not going to be useful.

So it's just gonna ⁓ introduce unnecessary complexity. ⁓ but we are also using some of those engineering tech prompt engineering techniques. For example, this is a role specific. Like we are saying that hey, you are a customer support chat, but so we are providing that kind of, we are assigning that kind of role ⁓ to the model. ⁓ then we are trying to be simple and cr and you know provide enough information and specify some roles. We are not using few shot examples because again.

Sneha Mehra (01:47:38)  
In this case, there is no point. There is ⁓ I I cannot see at least a way to provide examples that can help. But you can potentially here say ⁓ this is how you can cite your outputs. ⁓ returns can be made like within 30 days or something like that, and then ⁓ like put the

Source text, or you can include the numbers or numbers, put the source text, or something like that. So you can show some examples here. ⁓ and then basically you are doing prompt engineering, and then the model would work better, probably, if you want to support citations in a in a very ⁓ specific way. so that's basically the prompt engineering part. ⁓ Okay, got it. And then chat history, like each time with next question.

The whole chat history keeps getting sent, right? That's true. Both the qu which is both the question and the answer that ⁓ that was given. Yes. Yes. ⁓ And ⁓ does it include that three, that K, all that context, K parts, or what is it? ⁓ so basically each time there is a question and answer, we just include it as part of the chat history. So that next time if there is a question that somehow ⁓

Is referring back to one of the like earlier ⁓ iterations in the conversation, the model has access to those. For example, let's say in your first question you ask what is the refund policy, and then the model says something, and then you ask a follow-up question, like ⁓ based on what you said, you know, what if it happens after seven days or something like that. ⁓ If you don't include those previous ⁓ like ⁓ iterations in the in the conversation history, the model would ⁓ probably hallucinate or won't be able to answer.

So that's why we include this chat history. In each session, at least we make sure that we include everything else so far in the chat history. ⁓ And ⁓ actually, ⁓ Iki asked ⁓ this, and it's a ⁓ good point. ⁓ she wrote it in the portal that all these are open open source ⁓ libraries. Now, at if you're at workplace, do we have freedom to ⁓

Sneha Mehra (01:50:00)  
Use all this open source work or how how does it work? I mean you can cover later also if you like. Sure. ⁓ I can just cover it briefly. ⁓ and offline we can talk more, but ⁓ no companies it depends. I mean it depends on their license. ⁓ so ⁓ there are these open weight models where you can ⁓ use, but it also depends if they all allow commercial use. So if you're a company and you want to build something and use that open weight model to

commercialize something and have revenue. And if the license doesn't allow it, then ⁓ then basically it's not allowed. So ⁓ yeah that's that's basically the short answer. Okay, got it. Thank you. That's all appreciated. Sure. I think the for example here you can see some of those licenses I think ⁓ if it's still ⁓ as before. So you see this is MIT license this is ⁓

modified, you know. So ⁓ different companies, it depends on their legal team to determine which of these can be or should be used or sh or not used. All right. So let's go to the next question.

Sneha Mehra (01:51:14)  
Kiki.

Hey, hi, Ali. ⁓ I'll try to be quick. I posed my question in the community. This is related to the system templates. ⁓ I was quite frustrated about the hallucination because this is like not appropriate ⁓ in a company. So I would appreciate maybe you can put in the community chat like any resources to fine tune and technique. Of course, I can pick a better model because by not using the one billion, I can try

be and all that. And ⁓ yeah, I'm just wondering because I when I run it multiple times, it just have some random answer. Yeah, yeah, yeah. It was also the case, I think, here in our life. Yeah, I know. It's very frustrating. It is. It is frustrating. And ⁓ sure, I can I can share, but again, hallucination is not a solved problem. ⁓ there are some techniques or like fine-tuning or switching to better, bigger models ⁓ or

Increasing the inference time budget, you know, to verify things. ⁓ so those are some of the possibilities, but the model is still may hallucinate. We can't just ⁓ reduce it. I will I will share ⁓ resources. Yeah, because this this simple project too is an eye opening experience. We only have four PDF, extremely simple. We ask straightforward questions, ⁓ and I was just surprised by ⁓ the bogus answer it came back.

Yes, yes. But you know, in practice it requires a lot of ⁓ tuning in general. I mean, we are just starting with some random numbers, like three hundred for the chunks, right? And we are using the simplest form of chunking. I mean, there are more intelligent, smarter way of chunking. So if we chunk based on like logical meanings or logical stops. ⁓ basically this project is just aim aiming to provide the like a starting point. But if you f for instance, land

Sneha Mehra (01:53:09)  
search lang chain text splitter, you would see more advanced way of splitting and chunking text. So those are all the pieces that require a lot of time to be tuned correctly based on the the knowledge base. And for this particular example, I have tried Lama 8 billion, I think. Eight, seven, eight, ⁓ one of them. ⁓ I have tried it and it works well, at least for the simple questions. It never ⁓ hallucinates in terms of the the e the email or phone numbers or things like that.

So that's another thing. I mean, once you scale from 1 billion to like four, eight, or or like 30 billion or something, it's it's it becomes very, very unlikely to see some of these very obvious hallucinations. ⁓ I'm going to share some some more resources. Yes. ⁓ so does it include the way it follows instructions? Because I in this ex exercise, I really wanted to say if it is not sure, is to say I'm not sure from the docs.

But then it generates some bogus answer. So that is a particular use case. If it doesn't know, I expect it to follow instruction in the system template. Mm-hmm. Mm-hmm. ⁓ It's it's again, it's because the model is not too good. If you switch to better models, it would fix all of those. You can also play around with adding more instructions. I mean, for example, if at the very end say ⁓ if you are not sure, don't say anything else, just say I'm not sure in the docs ⁓ from the docs.

then you may see that would probably no longer happen. ⁓ I thank you. I'll I'll continue to play with it. Okay. Yeah, yeah, yeah. There is a lot of time. I mean, I I have seen engineers ⁓ at work in some projects are spending like 30, 40% of their their time whenever they are close to a delivery, just to figure out the system prompt. So it's not like a simple thing. I mean in practice people spend hours to figure out what is the best way to create it. Yeah.

If you can find some resource and post on the community ⁓ in terms of the the art of tuning the system templates in this case, ⁓ it will be extremely helpful. Yeah, of course. I will I will share. Thank you.

Sneha Mehra (01:55:18)  
Rahul.

Sneha Mehra (01:55:23)  
Hey Ali. ⁓ just and thanks for the good presentation. Just a couple of questions. You said ⁓ like Fias is ⁓ widely used in the industry, right? So I'm assuming this can also be used ⁓ if you want to store the data on say AWS or Azure vector database, right? So I was doing some research and I think it seems like that's possible to use. Is that is that ⁓ how things work in in the ⁓ you're talking about face, right? Yeah, yes, yes. You just embed

Once you have the embeddings, they typically typically do not require a lot of storage. You may not even need to store them on S3. I mean, if that's not necessary. But yes, that's basically the workflow. So we should be able to use that library to store the data wherever we whatever the vector store we want to store, right? Yep. Yep. Because at the end it would just give you a ⁓ you you saw like here, it just gave us these two files. So we can just store them anywhere. Yeah. Okay. ⁓

Another question is like, you know, I think when we are storing text or images which requires this the same processing. So ⁓ whatever the files is doing in terms of ⁓ coming up with the embeddings, they're probably it's gonna be very similar for every sort of text. But if you're building ⁓ say product recommendation which could be very which is very specific to certain companies, then and if you if there are say like thousands and thousands of products, and if you want to deploy something similar to if you bought this, then

Like you know, how you see it in Amazon, then I'm assuming you will have to define attributes for all those things that you want to semantically store together, right? So it what what additional work would be required in terms of storing that kind of data? Yeah, we define attributes for all of the products in order for them to ⁓ appear ⁓ close by in vector space. So your question is when

The domain is a slightly different right right like recommendation system and and we want to embed items, then what is the most effective way? Yeah, like if you are if you want to embed c company specific products. Let's say you bought this product from company like product one, then you might also be interested in like ⁓ three different products, right? And this could be very specific to ⁓ the company. So just having the text, putting them together may not make sense in that case because they product

Sneha Mehra (01:57:41)  
Those products just from the text description may not be related. Yes, that's true. ⁓ it it really depends. ⁓ so I know different companies. I mean, for example, Pinterest is relying a lot on the graph structure and ⁓ embeddings are coming out of the graph. So we are not they are not really embedding based on the text of each item, right? They are having a separate model, it processes the entire graph of all these pins and items and images and everything.

And then as a result, they have a model which they can simply pass a pin to it and then it would embed it. I think at some point at some point meta was also using embedding models to embed accounts. ⁓ like let's say you have a social media account and this entire thing would be embedded. ⁓ and you know, these models are also, you know, it depends. ⁓ You can embed different parts of ⁓ logical items. I mean, for example, if there is a post, let's say on social media you want to embed your post.

⁓ a typical way is to use image or video embedding models to embed the the media or the content. And then you can use the title and the description and use some text embedding model to embed those. And then somehow you can just aggregate those information or keep them separately to figure out which kind which other posts have similar ⁓ images to this post, like social media post, and which other posts have similar like titles or descriptions. So these are typical ways that it's it's you it usually happens in other domains.

Okay. So it seems like the graph ⁓ more graph structure for store the storing the embeddings is probably useful to go through. Of course. I mean graph structure, image embeddings, titles, all these can be embedded independently. Okay. The third one is I have tactical question. So while using this ⁓ lanchain dot embedding library, I ran into an error ⁓ and I tried to solve it for like two, three hours. I couldn't I couldn't do it. It just says it gives me a warning and I

pasted in the ch in the chat for week two chat. It gives me a warning, but that function, like that's code is not executable. So it basically errors out. So what's your recommendation? Should I like it's probably versioning. Are you using this exact condain? Exact same thing. I use the same YAML file to ⁓

Sneha Mehra (01:59:53)  
install the all so maybe yeah just maybe just post it. I I I saw there was a post. I don't know if it was the same post or it's the same post. So okay if you saw it that's the same post. So if you can please respond otherwise I'll have to like okay I I was hoping that someone else can also reply if they have faced it. So that's why I was not facing it. So I thought maybe some if someone else has also faced it they may be able to reply. ⁓ I'll I'll try to see if if it can be reproduced on my site. But it's most likely just versioning issues.

Sneha Mehra (02:00:28)  
Ni ni shit.

Sneha Mehra (02:00:33)  
hey, hi Ali. ⁓ so Hali, my my first question is that you you're showing the chat history in ⁓ in that program. So does it have any limitations or is it like depends on the computer is it running based on memory? Yeah, that's an excellent question. There is a limitation. If we just keep adding everything to chat history, at some point we it would just go beyond the context window of this particular model, Gemma 31 billion. And then the model would just fail. So you have to keep managing it. ⁓

You have to either compress it or you have to do every once in a while you do ⁓ like if basically if you know just just a logic to to ⁓ decrease its its its size. You can just keep the last hundred ⁓ conversations or something like that, or you can also compress it.

Okay. ⁓ so ⁓ the other question is like ⁓ what are the techniques to control the token limits ⁓ for a model? Because beyond a point you cannot like ⁓ you know keep. I mean, the the more you ⁓ I would say the the more prompts you generate based on the rags and all, you should be hitting a limit of the number of tokens to be processed ⁓ if you're using any particular model. So the techniques that we use so that you know the we we can keep ⁓

the the response is grounded we have not hitting the limits. ⁓ So ⁓ this is also an excellent question. I suggest you take a look at this this page. ⁓ context engineering it's published by OpenAI in their cookbook. And then they talk about different ways to to manage ⁓ your c the context window, given that the context window size is limited and they have to do some other techniques. So they talk about trimming and compression. So this this can be a very good page to read.

Okay. sure. thanks. And ⁓ the the the third questions I'll just quickly go over. ⁓ is like, so so you you're showing this Langchen is good for prototyping and you know for the local ⁓ experiments, ⁓ but ⁓ do you ⁓ know about you know the production grade implementations because long channels very slow, obviously it cannot be used in production. ⁓

Sneha Mehra (02:02:46)  
Let me get back to you on this. I I have to do more research. I have some names names in mind. I just don't know if they are as reliable for production. So please ⁓ just ping me, post, post on on channels, and then if I find more reliable robust libraries or frameworks for ⁓ for production, I I will share there. Okay. ⁓ the other one is like in this example, ⁓ what we have seen is that okay, we uploaded a bunch of documents, chunked them, ⁓ converted them into embeddings, and then

⁓ our ⁓ chat bot is basically trying to give an answer based on those embeddings. ⁓ so let's say I have a scenario of a production where ⁓ you know I have a documents ⁓ for each employee, a bunch of documents per employee of per customer per client. ⁓ And now I want to ⁓ you know I need to have the embeddings. I mean they're documents embedded separately. And then if I'm ⁓ doing giving a prompt to ask any question it should be specific to a particular coach.

customer or a client or employee. ⁓ So how to ⁓ like how do we arrange these embeddings so that it only goes and read ⁓ the similarity search only based on ⁓ the the fil filter based on that employee or the c or the client ⁓ and I mean so how do we store the ⁓ them in the vector DB in such a way that you know the the the the search is narrow? ⁓ Yep that that's that's a very good question and it requires some brainstorming. So ⁓

Again, I ⁓ I would like to bring some of these into the chats because they are good questions and I want others also to learn. ⁓ so ⁓ and there is also not a simple solution or a single solution to it, right? So there are a lot of different methods that can be applied. So I suggest if you can also post this question so we can also hear from others if they have opinions and I would also share my thoughts.

sure. Ali, well do. And just the last request. I think ⁓ I I posted this question earlier. probably there was a confusion. ⁓ so you ⁓ it's like you're ⁓ doing this extra office hours. sometimes I miss the developer office office hours because of the work conflicts. ⁓ so ⁓ like I mean ⁓ we are poding those recordings, but is it possible for you to upload the extra office hour recording as well? Because that like gives us more, like, you know.

Sneha Mehra (02:05:02)  
information in terms of question cooking and i i don't see that ⁓ uploaded sure ⁓ we'll we'll ⁓ sure i'll i'll look into it and then ⁓ if it's not at least for the next ⁓ extra office hour we'll make sure to have those and and upload those. Yeah thank you Ali thank you for your time yeah all right so ⁓ there are a bunch of other questions I have also a hard desp at ⁓ 1215 so I'll just try to

Take more questions until 12, 15, and then we can end. ⁓ also if possible, just to take all the questions, maybe we can limit to one or two questions. So I'll make sure so that at least I can just answer to everyone's. ⁓ Hi Ali, thanks for the need, ⁓ I was specifically interested in ⁓ using this approach for finance data. So for example, if I'm building an agent for ⁓ financial analysts.

And we're trying to get data from Excel documents where maybe the data is not in a table, but like, you know, grouped ⁓ visually into like tables of information or even CSV files. ⁓ one struggle I've had with that is the LLM often doesn't understand the business context to like aggregate the data at the correct level or understand which is the parent metric and like the drivers that are contributing to that.

And so when it gives me in insights based on those, it hallucinates a lot. And even it hallucinates the actual numbers itself. So I was curious if you had any suggestions as to how I can overcome those troubles. My my suggestions are probably the same as as what I was sharing earlier. It's first it it's difficult to say what is ⁓ each component needs to be evaluated separately and independently and and ⁓ see where the issue is coming from.

But that's basically my suggestion is more like best practices. So my suggestion would be first to figure out which component is is ⁓ is the problematic part. Is it retrieval? Is the retrieval not the best retrieval? Is it ⁓ if it's if if it's is it the like the like how fine grained the chunks are or ⁓ you want to be more fine grained? So then it requires more like tuning the chunk size and things like that.

Sneha Mehra (02:07:18)  
If the overlaps are not meaningful, or if you're your if your data in such a way that you cannot just chunk them based on number of characters or tokens, then you may want to switch to more like a smarter ⁓ chunking algorithms. If the approximate nearest neighbor search is not good, if the embedding model is not really capable for your type of data. So these are all the things that should be identified individually and independently. And then once you know which one are which ones are more problematic, then you try to fix those.

And then obviously the LLM itself, if it's it's powerful enough or not. ⁓ so it's yeah, this is just the the best practices that I would suggest to follow. Got it. And in such a case where numbers are your primary data rather than a text, ⁓ would you suggest ⁓ just passing those numbers into hard-coded Python functions to do specific types of analysis, like maybe you do trend analysis or correlation or whatever, and then have an LLM just summarize the results of that.

Or do you think an completely agentic approach is still valid for such ⁓ a I ⁓ I don't know. This this should be experimented. It's very difficult to say. I mean, ideally you want to just keep your system. Again, it's just a best practice. You want to keep your system as simple as you can and then gradually introduce complexities if necessary. And I think that's the case also here. So my suggestion would be smart start with the most simplest way that you can build this.

And then you figure out what part is causing it, causing the issues if it's not satisfiable. And then you start to fix those and introduce complexity like gradually. ⁓ So when you say introduce complexity, do you mean going from a deterministic approach to agentic or yes, yes, like agentics and then introducing more tools or more more access to more data sets and all those things. I mean those are all good tools you have, but as you introduce them more, it becomes even more difficult to even

even evaluate and figure out which parts are not good enough. As soon as you introduce tools, you also need to evaluate if your model is good at tool calling as well, like your LLM is good at tool calling itself or ⁓ or not. So it just kee becomes more and more challenging and difficult. So in practice, in most cases, the preference is always to start with the simplest possible way to build something. And then depending on where it fails, why it's failing, we try to just ⁓ slowly increase the complexity.

Sneha Mehra (02:09:42)  
Okay. Thank you. Yep. Next question. ⁓ Julio.

Sneha Mehra (02:09:55)  
⁓ I think you are on mute. ⁓ can you hear me? Yes, yes, I can hear you. Okay, great. thank you very much for your presentation. ⁓ have a have a quick question regarding this this example. So this system pairs a seventy seven million parameter embedding model, which is a GT small, ⁓ with a one billion parameter, which is geometric.

And ⁓ this definitely creates an interesting bottleneck because the embedding can capture the semantic nuance, but the large language model may lack the capacity to effectively ⁓ synthesize the ⁓ retrieve the context. So if you're in a in a resource-constrained environment, how would you determine whether to upgrade to a large language model, for example seven billion, or switch to a better embeddings, for example, ⁓ ADA zero zero two from OpenAI?

like which one yield to better ROI? Like is there any sort of rule of thumb ⁓ rule of thumb between embedding and large large language model? ⁓ I don't do I don't think there is a rule of thumb in terms of a scale. Like as you scale your LLM, you can you also scale your the the size of your embedding model. Again, I think the best way is to evaluate all these pieces individually and separately. For example, ⁓ so let's say in the same setup, right? Yeah, each time we run this, the model ⁓

produce something incorrect, like the phone number, right? Now we can inspect this. We can see if this part was included in the retriever or not. If it was included, which I believe it was, because the number is mostly correct, then the problem is not the retrieval in this case. The problem is LLM. The LLM had the right context. It just messed up. ⁓ so ⁓ the best way is just to evaluate these pieces individually and then a scale as

long as needed for your requirements. So as soon as you feel that the responses are good, or as as as soon as you think that the retrieved items are ⁓ relatively relevant to what you want, then you don't need to increase the the the size of your embedding model. Unless you have the budget to do so. I mean if you have the budget there is no ⁓ you you can always pick the best model out there. ⁓ But assuming there is a constraint environment, you be start with the s with the small one.

Sneha Mehra (02:12:10)  
And then you gradually increase until you feel satisfied. ⁓ and then ⁓ you do these things for each component. All right, all right. So there is no ratio between one and the other, it's just all about ⁓ experimenting and figure and and ⁓ assessing the quality of the output. Yeah, yeah. At least I haven't heard of. And embeddings module are typically very small. I mean they are not the button. Yeah, yeah. LLMs are the button like for most companies. So embeddings they they typically can afford to just pick whatever is the best out there and use it.

All right, super. ⁓ I have another one, but I I'm gonna post it since you have an hard stop. Yep, thank you so much. ⁓ No worries, thank you very much. Rashad?

hi Ali. ⁓ my question is how did rank RagChain know while doing prompt engineering that it has to replace context macro bracket context bracket with the retrieved docs? We never told that context bracket context is the ⁓ placeholder for adding the retrieved docs and ⁓ bracket question bracket for the question. The question. ⁓ Yeah, great question. ⁓ It is it is just the logic here. It's implemented in this class.

conversational retrieval chain. It it expects very specific things, LLM retriever, and then it just takes care of it. ⁓ and also the prompt template. So basically from this it knows what to replace. It has access to retriever and then it has the LLM. And also note that in prompt template we specified what to what should be replaced and what is the system template. So as soon as we pass this prompt, the the this library, this class would go into prompt

It learns what needs to be replaced. It tries to find this ⁓ from ⁓

Sneha Mehra (02:13:59)  
⁓ from the prompt. From the prompt.

So ⁓ I I can answer this question in the chat. Please send it. I am also now curious to learn ⁓ one detail about it. But I'm sure if you go and and read the documentation of this library, you you would figure out that ⁓ what is happening. But but it the logic is going to be implemented inside this. But I know I know your what your question is. Your question is how does it know that the output of retriever should go into context, not into the question. Norton question. Yes. ⁓

I'll I'll get back to you. Thanks. Yep. I just the short answer is I don't know. ⁓ But that's what I'll get back to you. Sure. ⁓ okay. So the last question, ⁓ Ranganasa. Hey, Ellie, I'll make make it quick. ⁓ retrieval. I have a one question is ⁓ can during the runtime, ⁓ can LLM trigger a retrieval ⁓ or is the retrieved content is passed only once during the initial time.

Second thing is I did not see ⁓ similar to question and ask answer history, I did not see a retrieval history. Is the reason behind that is ⁓ previous answer is kind of a proxy to retrieve information. So the question is why we are not keeping a retrieval history and we are just yeah, that is ⁓ the one question. Second is during the runtime, can LLM trigger a retrieval also?

During the runtime, basically the LLM itself cannot trigger time. Okay. It is basically the wrapper around it that allows ⁓ just coordinates this interaction, which is the chain. So the LLM is just nothing, just the next token prediction. It's just ⁓ and your first question regarding the ⁓

Sneha Mehra (02:15:51)  
The ⁓ history of retriever. Yes. I mean you can you can play with it and and include the history of retriever. Generally retriever is just a list. So that's why it's you're typically not using it. The the good thing with ⁓ the basically the retriever would replace the template part of the system prompt, right? So basically as long as you include the question in the chat history, you are having the retriever, ⁓ like everything that was retrieved because it's already in the context of the ⁓

The system from so ⁓ if you see here, I don't know if it's still here or not. ⁓

But you see, like ⁓

It is not shown here, but that's intentional. If it in fact, ⁓ you can play around with it. You can just include also the system prompt and see if it helps or not. Generally you want to make sure that there is a balance because if you just keep adding everything to the context window, you would run out of it. So that's gonna be problematic. Right. So there is always a a sweet spot, you know, in to to balance that. Okay. Okay. Yeah, thanks, Ali. Yeah, yeah, of course.

All right, I I I have to go. thanks again. I hope you like would like the next week's content about agents and please feel free to ask questions. I'll rest a little bit and then I'll I know there are some questions, I'll answer them soon, maybe later today or very earlier tomorrow. ⁓ yep, thank you so much. Bye.

—-------------------------------------

4

Sneha Mehra (00:00:00)  
Hello everyone. Welcome to week two. In week one, we focused on LLM foundations and we learned how LLMs work internally. We saw how we can build general purpose LLMs, ⁓ train them and use them to power chatbots. Now in this week, our focus would be on building a customer support chatbot. So along the way, we would learn all the techniques and ⁓ topics that are very important to make this possible.

In particular, we will talk about how we can adapt LLMs, meaning we go from general purpose LLMs, which we saw in week one, to specialized LLMs. And we would learn various techniques and ⁓ topics here, such as fine-tuning, prompt engineering, and retrieval augmented generation, or RAG. And then we'll zoom in to RAG and learn all the components of it, how it can be built and ⁓ its system design. And finally, we'll have project two.

Which is a hands-on exercise to build a customer support chatbot using all the techniques we learned and available tools and libraries. So let's start. The first question that we want to answer is why should we bother ourselves? Why even adaptation is needed? And to better understand that, let's see some examples. So imagine this is a general purpose LLM and we have some general purpose use cases. For example, we can ask this LLM some ⁓ math questions, like what is two plus two?

Or we can ask brainstorming questions like help me write an email to my manager to ask for time off. Or we can also ask coding questions like my code is not running, read it and fix the bug with the code. And then the LLM would process these inputs and it's very likely to answer these questions correctly. And the reason for that is ⁓ because it has seen a lot of similar examples in its training data. So it already knows how to answer these questions.

Now imagine we use this same general purpose LLM and ask some specific use case. So let's say we have a retailer store and we want to have an LLM and offer it as a chatbot to our customers. So the customers can easily ask ⁓ relevant questions and the LLM should answer that. Now, if a customer for a retailer store ⁓ asks our LLM a question like, what is your refund policy? the LLM is very likely to hallucinate.

Sneha Mehra (00:02:19)  
meaning that it would generate something that is not accurate or correct. Let me go to Chat GPT and see a real example. ⁓ Let me let's ⁓ let me ask a question like what is your refund policy?

Sneha Mehra (00:02:36)  
And this is basically a common question that our customers may want to ask us. So our LLM, it's very important to be able to answer them, answer it correctly. Now, if we use a model like ChatGPT to serve our customers, this is what ChatGPT would respond. Open AI's refund policy depends on the situation and how the purchase was made. So ⁓ basically this model is answering this question ⁓ and assuming that my question is about OpenAI.

So it's not going to really answer the question based on the retailer store that we have in mind. Now if I ask another question like what is ⁓ how long shipping take, how long ⁓ does the shipping take?

Sneha Mehra (00:03:25)  
Could you clarify what you are referring to? For example, so basically it's open AI model is very smart, so it asks for clarification because it knows that this prompt is ambiguous. But a lot of LMs may not even ask that and may just start hallucinating and just just throwing out some numbers, which is obviously incorrect.

So that's the problem. That's why we want to adapt LLMs. Basically, we want to adapt an LLM to a domain-specific use case, such as retailer store. So when these questions come to the LLM, like what is the return policy? How long does the shipping take? And what is the contact of customer service? The LLM can respond to this. More formally, we have a problem statement which is

Adapt a general purpose LLM so it can accurately answer questions in a specific domain using additional documents. So basically let's say we have a retailer store and we have a bunch of internal documents or knowledge base. And then this can be anything like PDFs, it could be images, HTMLs, Wikipedias, and so on. And then we want this ⁓ to use to start from this general purpose LLM and somehow adapt it. So when questions come, the LLM can use this information.

To answer the question. For example, if a question comes like what is the refund policy? The LLM should ⁓ answer based on the document database. You can return items for any reason within 30 days of your purchase. So this is the goal. And there are different techniques that this adaptation is possible. Three main techniques that we would learn ⁓ is first fine-tuning, second, prompt engineering.

And third, retrieval augmented generation, or rags for short. We would learn each about each of these and we discuss the limitations and ⁓ strength of each technique.

Sneha Mehra (00:05:21)  
All right. The first way to adapt an LLM is fine-tuning. And it's basically the most obvious way to adapt a general purpose LLM into a specialized LLM. And the idea behind it is to continue training the general purpose LLM on our document database. And document database is basically our ⁓ internal ⁓ knowledge base, or it could be a bunch of PDFs, images, HTMLs, Wikipedias, and things like that, which we want to

ultimately the LLM to learn from and can answer according to them. So after we train ⁓ the general purpose LLM on this document database, we would get this output which is a specialized LLM. Now this specialized LLM is ready to be used. It's basically having the tuned weights ⁓ and those tuned weights allows it to produce responses according to this provided document database. So after it's trained, it can be used.

For example, if we ask questions like what is the return policy, it would be able to answer ⁓ you can return items for any reason within 30 days of your purchase. And basically this answer is coming directly from its learned weight, or in other words, from its compressed knowledge of the new training data that we use to train. If we also ask questions like how long does the shipping take, it may answer typically 14 days, and so on.

Now there are two options to fine-tune the LLM, and we have them here. Option one is we update all the parameters of the LLM. And option two is not is basically parameter efficient fine-tuning, which refers to a set of techniques that do not update all the parameters of the parameter of the LLM. So we would learn about both of these in a second. So updating all the parameters, as the name suggests, is very obvious. We

Allow the training algorithm and the optimizer to tune all the weights inside the LLM as it's being trained on the training data. So remember, LLM has these blocks and each block has attention layer, MLP layer and mo and so on. And then each MLP is basically two linear layers. And each linear layer has their own weight. So when we are updating all the parameters, the optimizer may go back and just change or tune a little bit each of these weights so that.

Sneha Mehra (00:07:45)  
The model can better answer given its new data set.

However, the problem with this approach is that it can be very expensive. Remember, LLMs might have billions of parameters. So training an LLM and allowing the training algorithm to update all its parameters, it means that it can be computationally very, very expensive. And because of that, we have the second option, which is parameter-efficient fine-tuning. Parameter-efficient fine-tuning is basically referring to a set of techniques.

That allows us to adapt an LLM and fine-tune it without updating all its parameters. Basically, the idea is to ⁓ add ⁓ to only update a subset of parameters or a small number of parameters. And again, there are different techniques that allow us to do that. For example, we have adapters, we have LoRa, prompt tuning, activation scalers, BIOS only, sparse weight deltas, and so on.

And then the first two are very popular and widely used, adapters and LoRa. So we'll we'll zoom in into these two and understand how it they work internally. ⁓ and then there are some other papers and interesting reads for other methods, so which you can refer to later. So let's start with adapters.

So ⁓ the idea behind adapters is very simple. It says that instead of updating all the parameters of the LLM, we just freeze them, meaning that we do not allow the optimizer to update them during training. So this is gray means that they are frozen. We are not really updating their weights. And then we inject some new trainable layers inside this ⁓ architecture. So for example, between these, we can add three more.

Sneha Mehra (00:09:39)  
layers in red, and red means we allow the optimizer to update its weights. And then we train this model on the training data. So all these weights of the original LLM would remain unchanged, and then everything new would be learned such that whenever we have a new prompt, the prompt now can be answered according to the data set using these newly learned weights. And then this is the initial paper. It was ⁓

It was released in 2019, parameter efficient transfer learning for NLP. If you are interested to ⁓ understand the details of it, you can take a look at this paper. ⁓ And but what they are doing basically, they are just injecting these adapter layers inside the transformer architecture, and then they train it under new data to adapt the LLM for a new task. So it's going to be an interesting read. You can take a look at it, but high-level idea is the same. It just goes into the transformer architecture.

And then after attention and MLP, it just adds an adapter layer. So that's all the change, and only adapter layers are learned during fine-tuning and everything else is frozen. So this is adapters. Now the other approach is LoRa. This is basically another technique which was introduced after adapters. ⁓ And it's kind of following a very similar style. The idea is to add some new weights.

Such that we can freeze the rest of the model and only learn those new weights. And the way it does this is by introducing two low-rank matrices and it refers to them as the LoRa layers. So let's understand what that means. Again, an MLP layer has two linear layers, ⁓ and if we now zoom in into this linear layer, ⁓ it's ⁓ it transforms some input to output, and then internally it stores a weight.

And it's just a weight matrix of size D out by D in. D out is the dimension of the output, D in is the dimension of the input. Now, what LoRa does is injects layers here. So it basically targets all the linear layers inside the model. And then it keeps it frozen. So everything in this branch is as normal. ⁓ X goes still into the same ⁓ linear transformation. The weight is frozen, so it's not being changed.

Sneha Mehra (00:12:04)  
And then the output is WX. But now it also adds two ⁓ low-rank matrices, B and A. And then the idea is to also pass the X in this branch. And this branch also transforms X. And the output is ABX, if we just think of it mathematically. And then this ABX would have the same shape as WX. So we can just ⁓ add them up and then it forms the final output.

And these two newly introduced matrices are learnable, meaning that the parameters can be tuned during the training. So these are LoRa layers. And it's a very powerful technique. It also has some advantages over adapters. ⁓ And if you are interested, this is their original paper, LoRa, LowRank Adaptation of Large Language Models. And then they talk in detail about this, but again, the high-level idea is to

target all the linear layers and just add a this branch ⁓ with two low rank matrices and only learn those matrices and freeze everything else. ⁓ And again they discuss why this method has some advantages over adapters and why at inference time it's faster. But ⁓ but that's but that's a high level idea and it's very powerful technique.

Now there are lots of libraries that allow us to apply some of these ⁓ parameter efficient fine-tuning techniques. ⁓ One popular library is ⁓ Hugging Face ⁓ PEFT. And ⁓ it basically just gives you a set of tools that we can easily ⁓ wrap a model ⁓ and introduce new layers or new parameters to it and train only and train the model such that only those layers get updated. So

This is just you can take a look at it. It's it has some numbers and you know talks about advantages and so on. And then here I have also a very ⁓ short ⁓ demo to see how it looks like in practice. So this is my Jupyter notebook, and here basically I can use the ⁓ you I can rely on PEFT and import some configs. Again, these are not important.

Sneha Mehra (00:14:17)  
What is important here is I can use the Transformers library to pull a model. For example, here I'm pulling ⁓ Queen 2.5 3 billion instruct. It's basically a post-strained model by Queen, and it has three billi billion parameters. So let me load this.

Sneha Mehra (00:14:37)  
It ⁓ may take a while.

Sneha Mehra (00:14:56)  
Okay. So now we have this model. And this model again it's ⁓ an instruct model. So we can inspect it here. If we print a model, it would show us all the layers. ⁓ it has these layers, embed tokens, and these are the shapes, and it has layers and self-attention, you know, things that are very ⁓ like very normal things the in in a decoder only transformer. And we have n MLP here with some linear layers and so on. We can also inspect these parameters to see if they are

To make sure there are three billion. So for that, if if we print parameters, it would

would give us a ⁓ generator. And then to just see the actual numbers, if we do something like this, P for P in model dot parameters. Let's just print this to see the the shapes. So here we can see that these are ⁓ the weight matrices of all different layers in in the model. And in order to ⁓ get the actual the sum of all these weights, we can just do

Sneha Mehra (00:16:02)  
Some of all these. ⁓ And this number is basically 3 billion ⁓ 3 billion 85 million and so on. And so it it it is consistent with the 3 billion naming. Now let's say we want to ⁓ we loaded this model. Now we want to train it, but we want to use parameter efficient fine-tuning, meaning that we don't want to really update all the parameters of this model. And for that, we can apply ⁓ PEFT. So again.

You can refer to ⁓ the documentations to understand all the details and inputs. But at the very high level, what I'm doing here is basically I'm creating a config for LoRa. ⁓ And once I have that, I have this ⁓ wrapper gif get PEF model. I pass the initial model and the config that I created to get this new model. And now this new model is basically have this newly added LoRa layers and only trains those. So everything else is frozen.

So if I print trainable parameters, it would say total parameters is this and ⁓ note that this is higher than this, and that's because now we are adding LoRa layers.

sorry, this is not the train so the all parameters is this. ⁓ And now note that this is higher than this, a slightly, and that's because we are having this additional LoRa layers. And only these many parameters are trainable. So it's very small compared to the three billion parameters. It's only three million. So only ⁓ zero point one. ⁓ this is the trainable percentage.

Only ten percent of the original not ten percent, ⁓

Sneha Mehra (00:17:51)  
0.1% of the original parameters are trainable. So that's why it's a lot more efficient. We only need to learn these these many parameters during training. And then now the model is ready. We can just simply use some data set to train it. And then once it's trained, we can save it and we can just name it something. We can have the same name and then LoRa at the end. And then later at inference, we can just load this model instead of the original LoRa model. And it it could be as easy as this. So we have PEFT model from pre-trained and then

load model and then we can pass whatever input we have, what is the refund policy. And if we generate and decode, we would get responses. And those responses now are more consistent with the new data that were used to train this model. So I just showed this example to say that it's it's very easy in practice to ⁓ to utilize some of these methods and algorithms to fine-tune LLMs. ⁓ yeah, so let's back to our lecture.

We talked about fine tuning ⁓ and we saw our two options updating all parameters and parameter efficient fine tuning like adapters and LoRa. In the next lecture we will switch to the second way of adapting LLMs, which is prompt engineering.

Hey everyone. ⁓ We'll continue learning about different techniques to adapt a general purpose LLM to a domain specific LLM. In the last lecture, we learned about fine-tuning and we saw that how by continue training a base model on our documents or knowledge base, we can get a specialized model that is capable of answering very specific questions. In this lecture, we'll focus on another ⁓ important and very popular technique called prompt engineering.

And we'll discuss various techniques under prompt engineering and also we'll learn how prompt engineering can be used to adapt the LLM to our a specific use case, which is a retail retailer store. So let's start.

Sneha Mehra (00:19:48)  
So what is prompt engineering? At a very high level, prompt engineering is basically a way to craft or design prompts so that we can extract the desired output from the LLM. So it's very simple. Normally a workflow is like this we have a prompt, which is a user's query, and it goes directly to the LLM. And then the LLM processes that prompt and outputs some ⁓ response. And this is the response that is being shown to the user.

Now, in order to incorporate prompt engineering, we'll change this simple workflow by adding this ⁓ yellow rectangle here. We'll call it prompt engineering, and it's just a logic that ⁓ changes the user's query prompt. It takes that as an input, makes some changes or additions to that, and the output is another prompt. And now this new prompt goes into the LLM and then the LLM generates response. So the only difference is that we are not going to

Pass the raw user square to the LLM, but we are going to ⁓ make some changes to it or ⁓ make some additions to it, and then pass the new prompt to the LLM. Now, this is the at the very high level what prompt engineering is, but there are a lot of techniques under prompt engineering. For example, we have few shot prompting, zero shot prompting, chain of thought, or COT prompting, which is very powerful, role-specific, user context, and probably there are other examples too.

And there are also more advanced algorithmic algorithmic ⁓ prompt optimizations, which makes this process more ⁓ automatic. So in this lecture, we're going to focus on very popular techniques, including ⁓ few shot, zero shot, and COT, to better understand ⁓ what they do, how they do it, and why they're why they are very powerful. And then after that, we'll see how we can use some of these techniques to for our use case, which is a re retailer store.

So let's start with few shot prompting. To better understand few shot prompting, I have some real examples and I want to show how in practice it works. Again, I have hyperbolic service open here, and we saw that in previous lectures. It's basically a service that allows us to use some pre-trained models ⁓ and ⁓ send some questions and get some responses. ⁓ And here I continue to use the base model, which we used in the previous lectures.

Sneha Mehra (00:22:11)  
And the reason for that is because base model is just a continuation model. It continues the prompt. And I want to show that how powerful prompt engineering can be. And even in base model, it can help, it can influence the generation process. And then after that, we'll also switch to some post-trained models and see how prompt engineering helps there. So let's start with a prompt like how can I learn machine learning?

So and let me also reduce this to sixty. So before I press this, what we expect to see here is that it's just going to mimic what it has seen on internet. So it may just ask more questions, ⁓ or it may not even directly answer my questions. It may include something that are irrelevant. Let me see how it responds.

Okay, it's just it's just very irrelevant. It's it's coming from somewhere probably on the internet and it says a few years ago I was wondering the same thing. ⁓ I had just started my PhD and and so on. So it's obviously not answering one question. And again, we were ⁓ in this case it responded this way. It may if we ask again, it may just ⁓

Generate more questions. Let me run this again. How can I learn ML?

Sneha Mehra (00:23:36)  
What are the prerequisites? How can I learn ML? See again, ⁓ it's just asking more questions. Now, what I want to show next is that if we switch to few shot prompting, we can actually force this model and influence its generation, ⁓ generated response to follow the format that we want. So for that, I have ⁓ examples here. And let me just copy paste this here. So what I'm

Entering here this time is I'm actually showing some examples to the model in the prompt. And then this way I hope that the model would just follow the same format. So what I'm having here is that I have here this this particular format, QA, QA, and then I have a couple of random questions like what is a good place to eat? And then the answer is Berkley has lots of good restaurants. Another question is what is the process of leasing a car? You can go to dealership and talk with a salesman.

What is the capital of France? Paris. And then ⁓ on the last line, I have my actual question that I want the model to answer. How can I learn ML? Now, this time, the difference is that the model is not really have all this flexibility to generate however it wants. It's going to follow this particular format. And this particular format is designed such that it actually answers my last question. So let me send this to see if it works now.

Sneha Mehra (00:25:14)  
I don't know why this keeps happening.

All right. It's answering and it says take the class. And the reason that it's such a ⁓ useless answer is because the examples that we showed are also not too comprehensive or not too helpful. For example, we say just go to dealership and talk with a salesman. And then ⁓ as a result, the model is also telling us take the class. ⁓ And but basically the important part is that even a base model can respond to our questions when we use few shot prompt engineering.

Meaning that we show some actual examples to the model. Now let me switch to a post-trained model ⁓ and I can choose Lama here. 40545\. So it's the post-trained mod equivalent of the base model. And then what I'm going to do here, I'm going to

Send this example. So, this is what would we normally ⁓ ask our questions. We would say, for example, if we have a math question, we can have this like John has four books and boys three more. How many books does he have? Now, in this example, the model would respond however it wants. We have no control over the generation process. For example, we may want the model to answer ⁓

In a short format and just say seven or something like that. But we have no control over it. So now if I send this, the model would probably say easy one. John has four books initially and buys three more. So four plus three, seven, John now has seven books, which is great. It's very relevant and it it's really helpful. But ⁓ let's say I don't want the model to answer this way. And I have another prompt here.

Sneha Mehra (00:27:05)  
What I want is that the model, I want the model to answer in a particular format. And I have it here as an example. So I only show the model one example. That, like if you have three apples and get two more, how many apples do you have? And it just follows this format: answer and five here. Now, next I ask my real question. And what I expect to see is that.

The model would generate something like this. So it no longer would say easy one and lots of explanations and thinking and logics. So let's see if it actually works.

Sneha Mehra (00:27:43)  
All right, it exactly followed what I was hoping. So it it's following this format. And this is why few shot prompting can be very helpful because it can basically ⁓ guide the generation process and guide the LLM to generate something in the format that we want. Now let me go back to ⁓ to the lecture. So we saw few shot prompting, and then the next thing the next thing we want to talk about is zero shot prompting. Okay, so

Few shot prompting was showing some examples. Zero shot prompting is not showing any any example and is still guide the generation process. So how does that work? Let's go back to the base model again.

Sneha Mehra (00:28:27)  
And then I ask the same question, how can I learn ML? ⁓

And let me just decrease this to even lower, like 20\. So for this, let's see if it asks more questions or not. All right, it's exactly asking more questions as we were expecting. Where do I start? I want to learn ML. Where should I start? Now let's say I want to ask the same question, but now I want to use zero shot prompting technique to guide the model to answer my questions instead of just asking more questions.

It's very simple. What I'm going to do is here I'm going to say, hey, this is Q. And then here is A. So with such a simple change, it's very likely now that the model would actually start responding to it. Let me try.

All right, exactly as we were hoping. You can learn ML by reading books, watching videos, taking online courses, or attending boot camps. So with some very, very minor changes ⁓ and applying zero shot prompting, we see that how we can change a base model to a model that can actually answer questions. Now, let's also go to the post-trained model, like Lama.

⁓ And I have another prompt here ready. ⁓ And actually, let me not enter that because it's very obvious. So we'll just go over it. Another example of ⁓ zero shot prompting is that, for example, let's say we have a ⁓ like resume of some candidates, and then we want the LLM to extract the years of experience. So one way we can ⁓

Sneha Mehra (00:30:15)  
Prompt the model is something like this, extract the candidate's years of experience from this resume. However, in this example, we have no control over the output. So we can rely on zero shot prompting and enforce the model to output in a structured format. For example, we can say something like return only JSON-like, and in this format, years experience number from the resume or for the resume. Now by

By just this simple change, now the model is capable of ⁓ generating responses in the exact format that we want. So this is at a very high level, zero shot prompting. We basically just write right in the prompt as the instruction that we want the model to follow. We either include some formats as a hints, or we actually explain it in a in a text that hey.

Do this in this format and make sure your output always follows this particular format. And this is zero shot prompting, and it ⁓ is very popular and used a lot. Now, the next thing that we want to talk about is chain of thought prompting. So, chain of thought prompting is probably the most popular technique under prompt engineering. And later it was a starting point for thinking models, which we would discuss ⁓ in future weeks. But ⁓

It is very powerful and also it not only helps add inference time in a prompt engineering format, but it can also help the model to actually learn to think ⁓ during training. So again, for the training part, we'll we'll learn that in future weeks. And in this lecture, we'll only focus on the prompting technique. So everything was started from this paper, Chain of Thought Prompting. It was ⁓ it was released in 23, ⁓ January 10th.

By Google Brain. And then basically the idea here is that if the model if we ask the model to answer something that is difficult, the model is probably going to fail. But if we show some examples to the model, and in the examples we show how to think ⁓ in a logical, a step-by-step way, the model would probably follow that. And if the model followed that, it would be able to answer correctly. And this is the example that they have, ⁓ some mathematical question.

Sneha Mehra (00:32:44)  
⁓ and then ⁓ it's basically few shot prompting here. ⁓ They say a standard prompting. So what they mean is that they're not using chain of thought prompting and they're just using few shot prompting, meaning that they are showing the model a question and an answer, and then they ask their question. And then the question is basically the cafeteria has ⁓ had 23 apples. If they use 20 to make launch and bought more, how many apples ⁓ do they have? So the correct answer is nine.

But the model is outputting the answer is 27\. Because it's exactly following this format, the answer is. And then it it doesn't get enough time to think and ⁓ reason. Therefore, it just throw number, which is 27 here. Now, the idea behind chain of thought prompting is that instead of showing the model this very small ⁓ short kind of answer, we show the model how to l reason and how to think. So the only

change here is that instead of saying the answer is three, they change this answer to something more more ⁓ logical. So for example, they say it started with five balls, two cans of three tennis balls, each is ⁓ six tennis balls, five plus six equal eleven, the answer is eleven. So this way the model would learn to follow the same format. And as it follows this, as it's following the same format, it basically it in it pushes the model to think.

And ⁓ and ⁓ figure out the correct answer like a step by step in a smaller ⁓ by a smaller computations. So here you can see exact same question. And then now the output is the cafeteria had 23 apples originally. They used 20 to make launch, so they had 23 minus 23\. They bought six more apples, so they have three plus six equal nine. The answer is nine.

So that's why it's very powerful because it basically somehow indirectly teaches the model ⁓ to think and to reason. So ⁓ this is very powerful. And here they're having some ⁓ evaluation metrics and they're saying that the solve rate it goes, especially mathematical problems, it it it becomes significantly better. So here yellow is basically their palm 540 billion model, and they're using a standard prompting.

Sneha Mehra (00:35:08)  
And then the solve rate is 18%. ⁓ And by just switching to chain of thought prompting, meaning that the model is exactly the same, all the weights and parameters are the same. We are not training anything. We are just using the same old model. But now instead of asking our questions, normally we would show an example with chain of thought answer. And by just this simple change, the model would be able to ⁓ improve its performance. And solve rate now becomes suddenly fifty-seven percent.

So ⁓ this is the this is where chain of thought started initially. And it's very interesting paper. You can you can take a look. There are more examples in different benchmarks, how they are teaching the model to think and reason. ⁓ and then so yeah, so that's that's chain of thought ⁓ prompting. And we saw the examples in the paper. I have also more examples here. For example, if you show the model, you have some mathematic questions with one ⁓ one letter or one.

Token answer, the model would follow that, so it may fail, especially if the questions become more difficult. But if we have some reasoning traces and ⁓ some ⁓ logical explanations, the model is very likely to do better. And then ⁓ remember, so this this paper came out on January 10th, and it was basically a few shot chain of thought prompting, because we are showing some examples in the prompt, so it's few shots.

And it's chain of thought because those answers are having reasoning traces. Now, ⁓ nine days later, this paper came out. ⁓ sorry, nineteen days later. This paper came out from ⁓ Google printing. And now the only difference is that they're saying that we can also have chain of thought, but in a zero shot format, not multi ⁓ few shot. So basically what they're saying is that we are we we ⁓ we we don't necessarily need to show ⁓

Some examples to the model so it reasons. And again, it's very interesting paper. You can take a look. Let me go back to our lecture because I have I have that here. So they pro propose a very simple change. They're saying that we don't have to show all these examples. What we can do instead, we can ask our question and we can just say, let's think a step by step. And that's it. And just by this simple change, we are guiding the generation process of the LLM and we are

Sneha Mehra (00:37:33)  
Pushing the model to think a step by step. And what that means is that the model would try to break the task into smaller ⁓ logical pieces or reasonings. And then it tries to just ⁓ just come up with the calculation a step by step. So it's it's very interesting observation that just by adding this this line of like five ⁓ five words at the end of each prompt, they would be able to get ⁓ a lot better ⁓ responses.

Here you can see they're saying that this is few shot and then ⁓ zero shot, and this is few shot chain of thought, which was this previous paper. And then they are proposing zero shot ⁓ chain of thought, which the only change is let's think a step by step. And now by this change, you can see how the model is actually generating a response with some reasonings. And again, this is the zero shot version of that. So same question, just zero shot. So the model only outputs eight.

But now here the model just keep ⁓ reasoning over this question. So that's ⁓ very interesting observation. And this is zero shot COT. And again, later this became the foundation to build thinking models and training models such that they can think. And we'll we'll we'll go over that in future weeks. ⁓ but again, yeah, that that's chain of thought, and it's very, very popular and very, very powerful because it

It makes a big difference. Without chain of thought, for many difficult questions, the model may not be able to answer, but with just chain of thought, the model would be able to answer because it can just break down the task into a smaller piece and then solve the task a step by step.

So we also covered chain of thought ⁓ and there are also more. For example, we have role-specific prompting, which we would briefly cover. ⁓ but we are not gonna cover every single different way of prompt engineering. There are lots of different methods. If you search Google ⁓ with prompt engineering, there would be lots of ⁓ articles, papers, and so on. These first two links are pretty interesting. It's ⁓ this one is a Git repo prompt engineering guide, and it has lots of

Sneha Mehra (00:39:44)  
resources and ⁓ information. And then they have also a web version which goes over all these things. ⁓ Zero shot prompting, ⁓ few shot prompting, chain of thoughts. And you can see there are a bunch of other ways of prompting, meta prompting and prompt chaining, tree of thoughts. Again, many of these we would cover later. But this could be an interesting documentation to take a look at.

⁓ let's go to the role-specific prompting and understand why ⁓ it can be helpful. So, role-specific prompting is also very simple. Basically, the key idea is to assign a role to the LLM. So, let's say normally when we have a question, let's say a tax kind of question, we can ask it like how can I and whatever our question is, and just pass it to the LLM. Now

Role-specific prompting just adds this sentence before that and assigns a role to the LLM. And it says something like, you are an expert tax advisor. And then it includes the user's query. And just this simple change makes it make the LLM work a lot better. So that's it. That's role-specific prompting. Now we've discussed various types of prompting, like ⁓ few shot, zero shot, ⁓ instructions.

you know, providing instructions as part of the zero shot. And then we saw chain of thought and roller-specific prompting, and there are other kinds of prompting. And then all of these basically make ⁓ the prompt a lot of additions to it, the initials user query. ⁓ And because of that, prompts are generally having two components, the final prompts that are going to the LLMs. One is the system prompt and one is user prompt. System prompt is hidden instructions provided to the LLM.

And these are basically as a result of all these ⁓ prompt engineering efforts. And then user prompt is basically visible query that the user types. So whatever the user types, it becomes the user prompt. And then all the prompt engineering works that we have, it becomes part of the system prompt. So it's going to look something like this. We have a system prompt, and system ⁓ prompt can do ⁓ specify roles, give instructions.

Sneha Mehra (00:41:57)  
Provide few shot examples, also ask for a chain of thought, ⁓ thinking, and so on. And then whatever the user types is going to become the user query. And these two will get combined and then pass to the LLM. And if you go to Chat GPT and also all other major ⁓ chatbots, they have a place to specify your system prompt. Also, here in hyperbolic, if we switch to a model like ⁓ Lama four five billion.

There should be a place where you can specify a system prompt. ⁓ And I think here it's here. So it says it describes your system prompt, and then you can just ⁓ include all your instructions and role-specific prompting and examples and anything that you have in mind here. And then each time you would enter something here, your question would get combined with system prompt and then pass to the LLM under the hood. So the LLM would know about all these instructions every time you ask a question.

So that system prompt, just to make things more organized and you don't have to each time ⁓ include all those ⁓ prompts as the user's query. And then we can just have it fixed, have it once in a system prompt, and then ⁓ start the conversation. The user can start the conversation by just asking their question without any additional handling. So ⁓ and that's about the prompts.

And the next thing I want to talk about is we talked about all these prompting techniques, but let's now get back to our ⁓ original problem. Our original problem was that we wanted to use prompt engineering to adapt a general purpose LLM to our ⁓ very specific use case, which was a retailer store.

Let's see how we can do that. It's very simple. We can have some software here, a logic here, which is mostly focused around prompt engineering. And then we have a system prompt. And then basically, prompt engineering is going to just ⁓ take this question, combine it with the system prompt in a very particular format, and then pass it to the LLM. And then how is it going to do that? What we are going to do, how we are going to write this, is basically we are going to say, okay, whenever the user's query comes.

Sneha Mehra (00:44:13)  
⁓ concatenate or combine it with this system instruction. And what is this system instruction? Basically, we manually create this and we say something like read the following documents and respond to the question. And what are these following documents? These are all the PDFs and ⁓ you know, knowledge base that we have internally for our company or for our retailer store. So basically in prompt engineering, we manually go over all our documents or ⁓ knowledge base.

We include all those information and policies and whatever is basically listed there as part of the system prompt here. Or it could be also as part of the actual user's prompt. We can manually have a software to to append it to the user's prompt. But regardless of that, whether it's going to be in the system prompt or user prompt, basically by ⁓ using prompt engineering, we can include everything, every important information we have in our ⁓ documents.

And then include it in the context. So when the question comes, what is the refund policy? Now the LLM does not need to ⁓ does not have to hallucinate or it's not ⁓ ambiguous anymore because it's very likely that whatever this question is, it can be found somewhere in these documents that are ⁓ that is basically the knowledge base of the company. So this is

Simply how prompt engineering works and how we can use prompt engineering to make a an LLM rely on our internal documents and data and answer questions. But there is one major problem with prompt engineering. The major problem is that in many use cases, for example, if a big company wants to ⁓ build an LLM adapted for answering employees' questions, big companies may have ⁓ millions of documents or millions of pages.

So including or may even thousands of pages, including all those ⁓ information inside the prompt is very costly. And a lot of times it's not even possible. Because LLMs, when they are trained, they are expecting ⁓ a maximum length of ⁓ number of tokens in their input. So if we go beyond that, it's computationally not possible to ⁓ for the LLM to process that and output it. It's not going to really ⁓

Sneha Mehra (00:46:33)  
It's going to be computationally very, very expensive. So all these LLMs have some maximum length. It's often often referred to as context length or context window. And then what context window window means is that what is the maximum number of tokens that we can pass as input to the LLM? If we go beyond that, the LLM can no longer process it. So that's the main issue. If our use case has just a few PDFs, it's fine. Prompt engineering can work perfectly fine.

But if we are having a use case with thousands of PDFs and pages and documents, this ⁓ approach may no longer work because this this constructed prompt would be very, very long and maybe beyond the maximum context window. So ⁓ and because of that, we have we're going to ⁓ switch to the RAGs, retrieval augmented generation, in the next lecture, and we'll see how RAG can ⁓ mitigate this ⁓ issue.

Hey everyone. In this lecture, we'll focus on the third technique to adapt an LLM or a general purpose LLM into a domain-specific LLM. And that's retrieval augmented generation or RAG for short. In the previous lecture, we talked about prompt engineering. And we also saw that how we can use prompt engineering and put together all the internal documents that we have in our company, let's say, inside the prompt and then pass this new prompt to the LLM. So this way, the LLM.

Has access not only to the user's query, but also all the available internal documents and knowledge base in our company. So this way, the purpose is that or the ⁓ the assumption is that the LLM should be now able to answer user's query correctly based on the information given. However, the key limitation of this approach is that LLMs have maximum context window, meaning that the input prompt cannot exceed

Beyond a certain length. And most companies, most use cases, we have a lot of PDF files and knowledge base is very, very big. So by putting together all the documents, it simply exceeds the maximum context length that ⁓ is practical. So, because of this, prompt engineering is not going to really work, especially in cases where there are thousands or even millions of PDFs available. So

Sneha Mehra (00:49:01)  
Rag is solving that issue. And our entire focus of this lecture is on RAG, its components, and how it solves that limitation.

Sneha Mehra (00:49:14)  
So what RAC does is again, remember, previously we had LLM and we had this what is refund policy, and prompt engineering was just putting together all this document database and include it as part of the input prompt. In other words, it was it was just concatenating all this text with this user's query and pass it to the LLM. But the problem is there are thousands of PDFs and ⁓ files here, so it becomes not practical.

Now what RAG does is Rag adds a new component here called retrieval. And the purpose of retrieval is that pass only the relevant documents to to the generation or to the LLM. Basically, what it does is that retrieval looks at all these documents here. It also looks at the question and it figures out which documents or which parts of each document can be helpful or can be used to answer this question.

And retrieval just outputs those pieces of text here. So retrieval basically narrows down ⁓ from this huge document database to pieces of text called retrieve one, retrieve two, and retrieve K that are potentially ⁓ useful for answering this question. Now, this, ⁓ along with the user's query, is passed to the generation or the LLM, and the LLM outputs. This way, we just narrow down the entire document database to a smaller set.

And then we are we are no longer gonna exceed the maximum window size of LLMs. So this is the high-level idea. There are a lot of details in the retrieval and also in the generation. And we'll talk about each of these in more detail. And we also talk about how to optimize and what are the efficiency ⁓ optimizations that we can add. But this is the high-level picture of RAX. It has two main components: retrieval, generation. Generation is nothing but the LLM.

Maybe with some additions which we'll talk about, but it's at its core, it's the LLM. Retrieval just searches over the document database and just narrows it down into relevant pieces. And the relevant pieces now are used along with the user's query by the LLM to answer. So ⁓ let's start by focusing on the retrieval. Again, retrieval is just equivalent of a search problem. It's it's what we want if we want to search.

Sneha Mehra (00:51:43)  
From ⁓ our document database and find all the relevant pieces of text that can be potentially helpful to answer a question. And in order to build this retrieval, there are two steps. One is we build a searchable index from documents. That is more a pre-processing step. We have documents. ⁓ First we pre-process, we have some ⁓ data structures in place. We go over those, pre-process those, and make sure we have a good index.

And then the second part is when we actually want to use the retrieval. And in that step, the problem is a search because now we have a ⁓ user's query and we want to go to this index that we built and just retrieve all the relevant ⁓ pieces or chunks of text. So we'll we'll talk about both of these. I have a more ⁓ some visuals here. This is the first step. We want to build a searchable index from documents. So we have some co code or logic or data structure here.

It takes all this document database and it builds an index. And this index is searchable, meaning that it is following some data structures that are very efficient and ⁓ effective to search from. And then the second part is searching from the index at runtime. So when the user has a query, like what is the refund policy, it goes to this ⁓ second ⁓ part of the retrieval, which is search, and then this ⁓ search component.

interacts with the index that we've built in previous step and then it outputs all the ⁓ relevant chunks or relevant pieces of text from the from the index. So these are the key two steps ⁓ and what we are going to do now, we are going to focus on this ⁓ build index. We'll see how can we really go from this huge document database of PDFs and possibly you know diagrams, tables, images, HTMLs into an index. So

This is typically done in three steps to go from ⁓ to go from basically this document database to the index. The first step is document parsing. Then we have document chunking, and finally we have indexing. So all these three, if they're if we follow them sequentially, we would be able to build our index. So let's see what is document parsing.

Sneha Mehra (00:54:10)  
So document parsing is simply to go from documents, which again can be PDFs, HTMLs, or things like that, into the extracted.text file content. ⁓ so ⁓ that's the whole purpose of parsing, the document parsing. We j we have documents, we have PDFs, and PDFs are not, we cannot really provide a PDF to an LLM. At the end, we saw in the previous week that LLMs expect

numerical inputs and numerical inputs can be obtained by having some sequence of text, not a PDF. So the purpose of this document parsing is to go from all these PDFs to actual ⁓ sequence of text. And we'll also talk about ⁓ other modalities like what if they're image in the PDFs. But but just for simplification, let's assume PDFs are text only. We'll extend we'll talk about how to extend it to images. There are two

Main methods for document parsing. One is rule-based, which is less commonly, it's less common, and ⁓ to the best of my knowledge, it's not really popular these days. And then we have AI-based models. So both are doing something just very similar, and their aim is the same. Basically, they want to go from this PDF. First, they identify all the ⁓ the layout, all the different ⁓ components. So this is the first thing.

That this document parsing, whatever the algorithm is, does. So first layout detection. For example, here we have the layout. So we know that here is the title, here is title, here is text, here is text, this is a figure. ⁓ title, text, figure is image, and the rest of it is text. And then the second step is text extraction. So once we know the layout, there is some algorithm to apply, like you know, AI-based OCR methods and so on.

And then those are used ⁓ to extract the actual text from these layouts, from these rectangles. And it just puts them together in a structured output like text block, and it's like the actual text. And then it just keeps doing on all ⁓ doing something similar to all these different ⁓ detected layouts, extracted text, and just create this structured output.

Sneha Mehra (00:56:32)  
So if you look at this as a structured output, it's just a bunch of text that were extracted from this document. And also for images, we can just ⁓ include the coordinates so that later we can somehow index those images as well. We'll learn how we would ⁓ index images. But for now, this is how we are going to convert all our documents, all our PDFs into a structured output. And there are ⁓ different libraries and different algorithms for document parsing. ⁓ One popular example is

This DDoc. It's basically a library for ⁓ automate documents parsing and bringing it to a uniform format. Again, it it basically does the same thing. ⁓ It's just a library. It goes from your documents, it converts them, and then it reads them ⁓ and creates this structured output here. And then another very popular library, which is commonly used, is Layout Parser. It's also a unified toolkit for deep learning-based document image analysis.

And ⁓ it's an interesting ⁓ repo and I think there should be some demo videos here and full talk. Those are interesting to take a look and also their paper. ⁓ let me click here to see to zoom in a little bit. So for example, if this is one arbitrary page of a PDF.

This is the code. We can just easily import it and then use some of its internal models to detect, for example, images and ⁓ drawbugs for layout detection and all those things, and also extract text using OCR-based methods. And then you see the output can be something like this. We have these ⁓ rec red rectangles, and then it identified the titles, text, figure, table, and so on. And it's not it also works with other types of ⁓

documents, not only PDFs, like, you know, magazines, ⁓ magazine scans, websites. It can it can detect text regions, image regions, and table region and extract those. And same for historical documents. So ⁓ and these are some sample codes. But again, my point is ⁓ there are some great document parsers out there and they are commonly used and we don't have to really implement all this logic. We can simply use it. It's going to be just a few lines of code.

Sneha Mehra (00:58:51)  
To go from our documents into a structured output in a text format. And also if there are images, also ⁓ extracted images as well. So this is ⁓ document parsing. And the next thing I want to talk about is the next step, which is document chunking. So ⁓ what is document chunking?

Document chunking is basically going from the documents, the extracted contents, a structured output, ⁓ the structured output that we obtained from document parsing in the previous step. And then you just chunk them. So the reason for that is because a lot of these documents, it might be coming from a an entire book or report. So they are too broad. ⁓ And ⁓ it may also exceed the context window if we retrieve the entire document.

For example, if one document is a book and we apply document parsing and we get the all the text content from the book, it's going to be a huge document. And we ⁓ usually don't want to return that as a re retrieved ⁓ relevant item to the user's query because it's just too broad and also it's too long. So these are the two main problems of ⁓ if we do not chunk documents. And because of that, it's very common to chunk them.

And chunking means breaking documents into a smaller and more manageable pieces. And we call each piece chunk. There are different terminologies there. But ⁓ basically, we go from this into a smaller chunks. And the advantages of doing this is there are a couple of ⁓ advantages listed here. First, we can handle non-uniform document length. So we have different documents, and these documents may have different lengths. And it's not usually good if we have some very small or very, very long documents in our retrieved.

Ideally, we want our retrieved content to have ⁓ somewhat equal length. It also improves precision because now we can really ⁓ find the actual ⁓ text content or smaller piece within a document that is relevant to the user's query.

Sneha Mehra (01:01:04)  
The third advantage is it it overcoming LLM limitations. Again, that's because LLMs require a certain ⁓ maximum context window. ⁓ And by just chunking them, it's going to be more efficient. ⁓ And we don't have to include parts of the document that is really not too relevant or too helpful to answer user's query. And it's also optimizing computational resources because, again, at the end, it's very effective. We don't have to include parts of the document that are irrelevant. So this is document chunking.

And there are some ⁓ different algorithms for document chunking. We have length-based algorithm, and it simply splits the text into chunks based on the specific ⁓ specified length. So we specify a particular length and then it just scans over the text in in ⁓ in one document, and as soon as it reaches a length, it would just ⁓ split it. And this is very heuristic based and it it

The problem with this is it may split in the middle of sentences because it's just purely based on the length. So a lot of these chunks ⁓ or ⁓ yeah, a lot of these chunks may end up somewhere that is not really meaningful or it's in the middle of sentence. The other algorithm is regular expression-based ⁓ splits. And they use reggez to split the text based on some specific punctuations, for instance, like ⁓ dots or question marks and things like that.

This might be slightly better than length-based, because we no longer split something in some like middle of sentences. But now the problem is it lacks the deeper semantic understanding of text. ⁓ And that's because a lot of times some some some pieces of text might be about the same topic. Sometimes when we reach a dot, the topic might become ⁓ different. And just by ⁓ splitting the text based on, let's say, punctuations, it's it's not really handling the semantics.

the the meanings of ⁓ different chunks of the text. So it lacks that semantic understanding. And then the third algorithm ⁓ or method which is more ⁓ useful and commonly used is ⁓ basically we use specialized splitters like HTMLs or markdown splitters. So these are specialized splitters that they split the text at element boundaries, such as headers, items, code blocks. So they're already built with having ⁓ those particular

Sneha Mehra (01:03:27)  
⁓ documents in mind. And they're built knowing that those documents can have very ⁓ complex structure and they work well ⁓ just because of how they're implemented. So this method is usually more useful in doc for documents in structured formats, again, like PDFs, HTMLs, and so on. And again, we don't have to implement any of these ⁓ in general. There are lots of good splitters out there. One popular example is TexX. ⁓

Text splitters by blank chain. And this this this is the documentation and some overview. And it can be an interesting read. There's some ⁓ guides and so on. But again, at the high level, what they do is they go from this document into a split. And there are also other libraries. This is just one example.

So back to our lecture. And then the other thing I want to share is that now there's some ⁓ hyperparameters that we can tune. ⁓ For example, lang chain has this chunk size, chunk overlap, separator, and so on. And these are specifying how to ⁓ split the text. For example, when chunk size is ⁓ 35, what it means is that ideally we want to chunk it such such that its chunk ⁓ is ⁓ has thirty-five ⁓ characters.

For instance. And then the other is overlap. And this is because a lot of times we may want to overlap these chunks a little bit so that ⁓ basically ⁓ we do not again cut somewhere in the middle that is not too optimal. ⁓ so for example, this is a real example of lang chain going from this text with these ⁓ hyperparameters, and then you can see the output becomes these smaller chunks.

And it also overlaps. For example, here it ends with form of. And then the second chunk starts from form of. So form of is the overlap between these. So ⁓ that's it. That's that's all about ⁓ document chunking. The next ⁓ the next ⁓ step is indexing, and that's what we are going to talk next.

Sneha Mehra (01:05:45)  
So I have indexing here, and basically up to this point we went from all our documents in different structures like HTMLs and ⁓ PDFs, whatever, into smaller chunks. And each chunk ⁓ is either some ⁓ sequence of characters or text, or they can be like images, like coming from figures or tables and things like that. So for now, again, for simplification, let's focus on text. We will cover how to handle images shortly.

So the purpose of indexing is to store each of these chunks in a data structure that is easy to search. And it's just simply a data structure, it's nothing more. And there are different ways we can do indexing. There are different methods, and then as a result, there are different search mechanisms depending on how we index. For example, we have keyword-based indexing, we have full text indexing, we have knowledge-based, ⁓ knowledge graph-based, and vector-based.

Let's briefly see keyword-based and full text, and then we'll focus on vector-based, because vector-based is is what's commonly used nowadays in practice, and it's very important. So we'll just briefly discuss about these, which are a little bit more traditional. And then knowledge graph-based. There are some ⁓ interesting reads out there. I will include those. We are gonna skip it because it's it's too specific and there's certain only certain use cases that ⁓ rely on knowledge-based, knowledge graph-based.

indexing and almost like ninety ninety five percent of use cases that I'm aware of all are all are relying on vector based indexing. So let's just start.

Text-based, including keyword or full, is basically matching partial or exact query terms with the content of documents. So basically during this indexing process, we index ⁓ like this. For example, if these are the text, we again we follow some data structure, and one very good example is Elasticsearch, for instance, which takes care of all these functionalities. And then it builds an index. So when you want to search from it, and for example, you have a query.

Sneha Mehra (01:07:57)  
It can easily refer to this index and say, hey, these parts, these chunks are including your query. Either it could be partial match or it could be exact query match. And this is something again that the user can specify. For example, you can say, hey, give me all the chunks that has this particular keyword or has this particular sentence in it. And text-based ⁓ indexing methods would just enable you to do that. Again,

These are good and these are what traditionally being used, but the problem is that ⁓ it's not really capturing the semantic meaning of documents. For example, if you're using synonyms, then you no no longer would be able to do any partial or even exact match. Even if there are some documents that having synonyms of your query. So these text based methods are again good for certain use cases, but they lack the ⁓ semantic understanding of documents.

And because of that, vector-based methods are ⁓ used commonly. So let's see what is vector-based method. Basically, the idea behind the vector-based methods is that instead of indexing the text in inside each chunk, we use some pre-trained models to encode text into some vectors. And these models are called embedding models. ⁓

Let me step back a little bit and first we understand what is an embedding model, how it works, and then we'll we'll get back to this vector based kind of ⁓ indexing soon. So what is an embedding model? An embedding model is also very common common term in machine learning. It's basically an ML model that maps some input, for example, text, or the models that also map other modalities like images or even videos. So they basically map their input, text, into a vector in high dimensional

Space. It's also called embedding space. And the way that these models are trained is that when they map text to their embedding space, it's a very semantically meaningful space, meaning that if two texts have similar meanings semantically, they would end up nearby in this embedding space. So the embeddings can't capture the semantic meaning of relationships. ⁓

Sneha Mehra (01:10:18)  
And the idea of this first came there was this paper word to back, and this is the first time, again, as as far as I know, that ⁓ this was ⁓ visually shown that hey, we can train some word models or ⁓ embedding models that after we map each word, after training, we map each word and we look at ⁓ its position in this embedding space. Similar words would have certain relationships. For example, ⁓ synonyms would be end up nearby.

And there is some relationship. For example, going from ⁓ woman to man would be having the same direction of going from queen to king. So this paper is here again, in case you are interested to read, it's quite old now, but it was the start of ⁓ something very important, which later turned out to be embedding models. ⁓ And they have all these diagrams that I just showed you, and also they explain how they train.

And why after training they get this ⁓ meaningful embedding space that similar words are end up being nearby and so on. So yeah, you can just take a look at this paper. ⁓ And ⁓ and after this, ⁓ basically, this is what we can do. We can just rely on some embedding models, not Word2Vec, some advanced models. There are lots of models available these days. We can use ⁓ one embedding model and then

We can go from all the chunks that we had, text chunks, and we encode them. So what that means is that we would have a table like this. We have the one column is the text. For example, to learn more about our products, contact this number. And then it just gets mapped to this embedding space. And there is a point in this high dimensional space, and this is the point. So in other words, we are mapping each text, each chunk, to a vector, a numerical vector. And again, remember.

It's it's ⁓ it these these numerical vectors ⁓ are in a are semantically like their relation is semantically meaningful in the embedding space. So if two vectors are nearby, it means that their meanings are close or they're talking about something similar. So that's the part to go from the chunks to an index table. Now

Sneha Mehra (01:12:37)  
Again, I mentioned there are lots of embedding models these days. For example, OpenAI has an embedding model which is quite good and often used. It's called text embedding three. ⁓ And ⁓ you can just pass something and get the embedding. And those embeddings can be stored and ⁓ those can be used to form our index table. And again, there are a bunch of other ⁓ embedding models. I just searched Google. I don't know if they're really the best ones or there are better ones now.

But just wanted to show some some examples. So we have OpenAI text embedding three large and then also Gemini. So Google has its own embedding model. And there are all these different ⁓ models available. And this is the output dimension dimension, meaning that this is so after we map the text, this is going to be the the the length of the ⁓ embedding.

So, for example, OpenNL Large has 3072 numbers in it in their embedding vector. So, back to the lecture. ⁓ This is the idea behind embedding model ⁓ and also how we use ⁓ an embedding model to convert these chunks and build an index. Again, in the next video, we will talk about how this enables us to search from at runtime.

But for now, our focus is on only building the index. So and this is how we are going to build the index. We are going to use some text embedding and just build this table. All right. So we talked about text. And now the question is what if some parts of our ⁓ extracted content or some of the chunks are not text. They're images. And it's very common because a lot of PDFs these days, or even ⁓ HTMLs, ⁓ or might have diagrams or some figures or some images. And it's

Very important to somehow encode those images as well so that when a user asks a question, we can also rely on those images as well to answer their question. How do we do that? There are ⁓ there are two ways. There are two options. First option is we use an ⁓ image embedding model. Very similar to text embedding models, we have some image embedding models. And what they do is they train this model such that the input is now image.

Sneha Mehra (01:14:56)  
And then the output is the same embedding. So they are mapping each image into a high-dimensional embedding space, such that similar images end up nearby. For example, here you can see auto, ⁓ we have cars here, so they're nearby, and truck here, and then like animals seem to be here, like dog, cat, and so on. ⁓ so these are these are image embedding models. And ⁓ it also started ⁓ the the first model that was.

⁓ widely used in in industry was the OpenAS clip model. And what they did was they ⁓ they trained the model using some contrastive pre-training method. And again their paper is available. This is their repo. They explain exactly how they train. But at the very high level, the model has two encoders, text encoder and image encoder. And they train all these ⁓ both these text encoder and image encoder together. And their data is basically pairs of images and their corresponding

caption or or ⁓ corresponding title. And then ⁓ this is trained such that the learned embedding space is also meaningful. It's basically shared embedding space, meaning that now when you map a text to this embedding space and also an image which is relevant or similar to the text to this embedding space, their embeddings are going to be nearby. So this is another example of ⁓ using an image embedding model. And

One thing to note is that ⁓ it's very important if we want to ⁓ map images into some embedding space. It's very important to use a model such as clip that has a shared embedding space and use the exact text encoder for ⁓ indexing our text textual chunks. Because otherwise, there is no relation between image embeddings and text embeddings. So if we want to also map images, we use, for instance, clip.

Text encoder is being used to map our build our text ⁓ index. And then its image encoder part is built is used to map to build our image embedding index. So that's about clip and that's about how to handle ⁓ images. But this is option one. There is also another option we can use instead of relying on image embedding models. The second option we have is we use some captioning model.

Sneha Mehra (01:17:21)  
And captioning models are the ones that you pass some image to them, and then they describe the image and they they generate a caption for it in a textual format. And this way we can basically first go from image to caption using this captioning model. You can see in this example, we have the image shows a cat sitting on a whatever and the rest of it. And now we use these chunk instead of this image to index. And now we can use our original text encoder.

To encode this image, because now we only encode its its corresponding caption. So both options are available to ⁓ build our ⁓ index tables. We discussed both option one, which was using a model with shared embedding space. We use the image encoder to ⁓ map images and build the index. And we use text encoder for textural chunks. We can also use option two, which means that we can use any text.

embedding model that we want. And for images, we can use image captioning to go from image to text. And then from text we can just encode it to the ⁓ to to build the index. But regardless of which option we use, at the end we would end up with something like this. We have ⁓ converted all our images and chunks that were extracted from the documents into two in indexes. One is image index and another is text index. And

These are under the hood just tables. ⁓ They're tabled like this, for example, chunk one, ⁓ this this this vector, chunk two, this vector, this vector, and so on. And these are for images. And then for text, very similar, chunk one, this vector, chunk two, this vector, and so on. So this is how we pre process the entire documents, and we go from the raw documents such as HTMLs or PDFs into an actual ⁓ indices, image index and

Text index. So we've we've covered all ⁓ steps of let me go back here.

Sneha Mehra (01:19:27)  
Okay. We talked about this building a searchable index from documents. Now, next what we are going to talk about is now how can we how can this index help us for to search from at runtime?

Alright, so far we talked about retrieval, and we also saw that retrieval is a search problem. And we saw how we can build a scalable index from our documents. Once we have this scalable index from the documents, now the retrieval is ready to be used, meaning that we can now use some algorithm to search from this index at runtime whenever there is a query. So let's zoom into that and better understand the detail. Again,

In the first step, we built our index. Now the focus is going to be on the search and understand how search works, meaning that when there is a prompt, like what is the refund policy and the se and this algorithm, this component has access to the index, how can it efficiently find most relevant chunks from this index relevant to this and output it? So that's the focus of this lecture.

So ⁓ how do we really search? For a given input text, what we want is we want to again retrieve relevant chunks of documents efficiently. To do that, we can use the same text encoder that we used to encode the text. ⁓ And first we encode the user's query, which is what is your refund policy. And the output of this is just a vector, an embedding. Now, remember, we used model, we used a text encoder model that ⁓ is

trained such that similar sentences are going to be nearby, meaning that their embedding vector is going to be nearby in that embedding space. Now, now that we have this embedding of the user's query, it's time to find the ⁓ which other embeddings are very similar to this. So that's why it's called a search problem. And it's basically a nearest neighbor search, meaning that given a query input or a query point, we want to look at all

Sneha Mehra (01:21:36)  
Other points which are associated with ⁓ text chunks or even image chunks, and find the ones that are closest to this point. And we define some distance ⁓ measure or ⁓ some similarity measure. And a lot of times we can use things like Euclidean distance, cosine distance, and those ⁓ formulations to find the distance between two points, two points in this embedding space.

And cosine distance is a common metric to use to measure the distance of two points. So we use something like cosine distance. And then we just compare every single embedding from the index with this query embedding, and we find the closest ones. And once we have the closest ones, we can now go back to the original chunk of text and include those as the output of this process. For example, these could be some different chunks.

That are extracted from ⁓ the original documents. Something like a refund policy to return or replace an item and blah, blah, blah. After 30 days, you may get a store credit and the rest. ⁓ And it is important to note that items should so these are ⁓ basically chunks of text obtained from documents that they're embedding are very similar to the embedding of this question, which is what is your refund policy. So

This is how we run search. And this is ideal because now we can simply pass this along with the user's query to the generation, which we'll talk in future videos. But now let's let's ⁓ inspect this nearest neighbor search. We understand how we can really find the most ⁓ similar embeddings to a query embedding.

So this is a very traditional and common ⁓ problem in computer science. Basically, the problem is given a bunch of points in high dimensional space. This example is just two dimensional for simplicity. And we are also given an E of Q, which is the query ⁓ embedding or the point ⁓ representing the query vector. And what we want is that we want to find the top K ⁓ most ⁓ like nearest points. In this example, if ⁓

Sneha Mehra (01:23:52)  
K is three, we want to return these three points. Now there are two different categories of algorithms for nearest neighbor search. One is exact nearest neighbor, and then ⁓ another is approximate nearest neighbor or a n. And there are also other ⁓ like categories within ANN. Let's start with X exact nearest neighbor, and then we'll talk about ANN.

So, exact nearest neighbor is basically ⁓ I think I don't have any visual for that. It's very obvious. Basically, this is the exact nearest neighbor. Basically, here the goal is to find, look at all other points in this space and find the top three ⁓ closest points. This is basically ⁓ guaranteeing that the return points are exactly the top three closest points to this query point. Now, the problem with exact nearest neighbor is that it's

computationally expensive, especially if there are lots of points. So lots of points means that we have lots of chunks and as a result, there are like ⁓ very large number of points in our embedding space. And that's usually the case in in RAGs because again, the knowledge base or the ⁓ internal documents that we have after we do chunking and then map the chunks to embeddings, there might be millions or even billions of ⁓ data points. And comparing query with billions of data points each time,

Is going to be very expensive. And that's why we have ⁓ approximate nearest neighbor algorithms. And basically the idea behind approximate nearest neighbor is that ⁓ we don't need to necessarily return the the absolute closest ⁓ points. But if we return something reasonable and reasonably close to ⁓ eq, even if they are not the top three, it should be still fine. So ⁓

Basically, it's an approximate version of the exact nearest neighbor. So let's see how ⁓ some of these big categories of algorithms work. We have ⁓ lots of different algorithms, and all of those can be ⁓ categorized into one of these four buckets. One is clustering based, ⁓ another is tree-based. We have also locality-sensitive hashing, and there are also other methods like graph-based and so on.

Sneha Mehra (01:26:14)  
Which we are not covering because of their various specific use cases. ⁓ but but let's let's understand how ⁓ how basically these ⁓ categories of ⁓ approximate nearest neighbors nearest neighbor algorithm works. Clustering-based ANN basically cluster the point ⁓ once. ⁓ So we have all these clusters. For example, we have cluster here, this cluster, this cluster, and we have all these five clusters.

And then we also have a representative point from each cluster. It can be usually the centroid. So ⁓ let's say we have a representative of each of these clusters. Now, when a new query comes and we want to find the top closest top K points to it, first we find the which cluster it belongs to. And then we only search within those clusters. So this way we sacrifice accuracy a little bit, but instead we don't have to go over all

data points or billions of data points. We can only go over the data points within that cluster. For example, if an a a query comes, let's say here, which I have my ⁓ I'm just ⁓ my cursor here. So

It first we would ⁓ discover that this query point belongs to R one or region one or cluster one. And then we would search for the top three points, which is going to be this, this, and this. However, this point in R two might be even closer, but again, we are sacrificing a little bit of accuracy for efficiency.

And in practice, it's you it works usually fine. And it's in in most use cases it's fine to sacrifice that little bit of accuracy. So this is clustering based, ⁓ and the other category is tree-based approximate nearest neighbors. And these algorithms at a high level basically partition the data and form a tree. So basically they go from the entire data space into sub-espaces. And then this way.

Sneha Mehra (01:28:19)  
Then they have these leaf nodes, which corresponds to regions. Now, when a query comes, they can just ⁓ traverse this tree and they see which ⁓ region it ends up with. And then that region is basically the region where we can just look over all the points. And you know, these two can be related somehow because when we are forming a tree, in some sense, we are just dividing the space, the high dimensional space, into regions.

but there are different algorithms and some of them are more purely focused on tree-based structures. And then the other category is locality-sensitive hashing. And basically, this method at a very high level, if I want to just give an overview, it uses some hashing functions that are a little bit different from normal hash functions. And the way that they're different is that if two points are, if two inputs are similar, their hash function would be also their the the outcome of the hash would be also.

similar. So these functions are designed again, they are designed such that when two inputs are close or similar or relatively similar, their hash ⁓ output would also be similar. And this way, if we have such functions, it's very easy because now let's say the the space of hash function is five buckets, let's say ⁓ one, two, three, four, five, six buckets. And then as a pre-processing step, we can just ⁓

apply the hash function to each of our data points, billions of data points, and we find their their bucket by applying the hash function to these points. And once we have this pre-processing, now when a query point comes, we can easily apply the same hashing function to see which bucket it belongs to, and then only search the points within that bucket. So again, all of these algorithms are following something very similar. They just try to find a way to ⁓ organize the the the points.

Such that at query time we can only look for nearest neighbors within a certain ⁓ cluster or bucket. So that's how we can ⁓ efficiently run nearest neighbor to find the most closest, like most similar chunks to a query or text prompt. And this is the overall ⁓ workflow. Basically, the text prompt comes here, it goes to the same text encoder, and then

Sneha Mehra (01:30:45)  
The embedding here goes into this search ⁓ algorithm. And then again, there could be different comp ⁓ complexities. It depends on how we design our searching system. So let's say we have inter-cluster search, it's kind of a clustering-based method. And then these methods first try to find the clusters that this embedding, basically, this text prompt belongs to. And let's say it it finds out that it's here. Now

The next step is this intro cluster search, which searches only this cluster and finds the most relevant ⁓ or closest points. And then once it knows the closest points, it can ⁓ return their their corresponding data chunk. So here we see retrieve doc one, retrieve doc two, retrieve doc K. And each of these are basically a chunk of that document. And all these are based on the these points are are coming from the index images and index text.

which were the result of our data preparation process to go from our documents to index.

So ⁓ that's that's all I all I wanted to cover for search. And the main thing to ⁓ to review is basically just ⁓ efficiency discussions that you know exact nearest neighbor is not optimal. So we'll try we can use ⁓ various algorithms of approximate nearest neighbor, and then we can efficiently find most relevant text pieces and image pieces, and then we can just return them as the retrieved item. Now

As always, we don't have to really implement these approximate nearest neighbors. There lots of libraries, different companies have implemented those. I know Google has Scan and Facebook, ⁓ which is very popular, has face. And basically, face is exactly that library. It's basically a library for efficient similarity search and clustering of dense vectors. And this dense vectors are those embeddings or ⁓ representation of our data chunks.

Sneha Mehra (01:32:47)  
So basically there are lots of algorithms and we can easily use them. And there is a documentation here on ⁓ face.ai. So it explains you know similarity search. They talk about various distance metrics ⁓ metrics that we've discussed, such as Euclidean distance, cosine distance, and so on. And then here they mention how to ⁓ install it and then how you know various documentations and ⁓ diagrams. So ⁓ that's it.

That's all I wanted to ⁓ cover for search. We also learned ⁓ building our pipeline. So at this point we are over with the retrieval component of RAC. So we exactly back to the original RAG, we exactly know how this retrieval works and how we can build an index and also later efficiently use some search algorithms to retrieve relevant documents.

Now the next thing that we want to talk about in the next lecture is the generation component. And now we understand what are some details about it and how can we even improve it.

Hey, we are getting closer to wrap-up brag. And in this video, we'll focus on the generation component. So we'll just ⁓ ignore this part, which we already learned about the retrieval, and we assume that the retrieval can retrieve some relevant chunks. And now we'll focus on this generation piece. And it's very simple, so it's going to be ⁓ short. So at its core, again, the generation is an LLM.

Just a pre-trained LLM, a general purpose LLM. And what we do is that once we have the retrieved content from the retrieval, it gets ⁓ added and combined with the user's query, which is what in what is the refund policy. And this way we form our prompt. And the prompt simply goes into the LLM and the LLM ⁓ generates some output. However, in practice, we can make ⁓ this generation a little bit more advanced.

Sneha Mehra (01:34:53)  
And that's going to be this workflow. So we can just combine it with the prompt engineering, which we learned as the second technique for adaptation. Basically, the main difference here is that now this generation has this both components, prompt engineering and LLM. First, the retrieve content as well as the user query goes into the prompt engineering. And then we apply whatever is necessary and helpful. ⁓ For example, we can add ⁓ role-specific prompting, we can ⁓ add chain of thought prompting.

⁓ And other things depending on the use case. And then prompt engineering, again, the output would be a new prompt, as we learned last time. ⁓ Now that ⁓ enhanced prompt or the engineer prompt would go into the LLM. And LLM would just normally process that and output the response. And again, prompt engineering, we can combine all different techniques that we learned. This will this can be one example, but this is just an example. For example, the ⁓

Prompt engineering can just convert all those retrieved contents as well as ⁓ the user's query to something like this. This is an example. This is not necessarily related to retrieval or rags. But basically, we can have user's initial query here. We can have context from search if we are doing some searches. In this case, it's basically probably the retrieved items. And then here we do roller-specific prompting, like you are a helpful assistant who answers ⁓ company name, internal questions.

And then here's some instructions. Your task is to deliver a concise and accurate response and so on. We can also do few shot prompting, show some examples that this is how you cite, ⁓ or things like that. We can do ⁓ chain of thought, ⁓ zero shot chain of thought prompting, and by saying if there is a reasoning process to generate the response, think a step by step and put your steps in bullet points. And then finally, we can have user context prompting, ⁓ something like this. ⁓ this is basically just to show that.

After we apply prompt engineering, we convert those raw retrieved content as well as user query into something more structured and more ⁓ specific and helpful for the LLM to generate response. So ⁓ that's it. That's the generation. It's simply an LLM, and we can combine it with prompt engineering to improve it. Now, one important thing that I want to note here is that ⁓ remember, this is the general purpose LLM. So

Sneha Mehra (01:37:20)  
If our retrieved items are not really accurate or is not too ⁓ helpful, this LLM may fail. So for example, let's say we have a retrieval and our retrieval or embedding models that we are using are not too accurate. So that means that some of these retreat ref retrieved items are not necessarily ⁓ too helpful to answer this question ⁓ or maybe irrelevant. Now LLM has no idea.

It basically has all this information and it thinks that all these are helpful or can be helpful to answer this. So it may just ⁓ produce some answers that are not necessarily ⁓ accurate. So ⁓ one step we can go one step farther and ⁓ also fine tune this LLM. So basically this way we are combining all these methods that we learned, fine tuning and prompt engineering and rack, to improve our generation process.

And this is a very ⁓ tricky training, and there is a very good paper called Raft, and I think it stands for ⁓ retrieval augmented fine tuning. So the idea is basically again the motivation is if if retrieval quality is not good, LM will generate suboptimal outputs. So we want to fine tune LLM to distinguish between relevant and irrelevant documents. So that's the key idea.

Basically, the idea is that we label some document some relevant and irrelevant retrieved contents. And then we continue training that the LLM that we have. And during this fine-tuning, the LLM is basically trained to generate responses based on relevant documents while minimizing the influence of irrelevant documents. And you can see the workflow here. Again, I described at a very high level what it tries to do. It tries to learn ⁓ which

Retrieve documents are relevant, which ones are irrelevant. So during the generation, it would only it would rely more on relevant documents and ignore irrelevant documents. And if you are interested to learn more about it, this is the paper. It it was published in 24\. ⁓ And ⁓ it explains all the details, how they are training the model. And ⁓ it's an interesting read. And it also discussed, you know, training time and how they prepare their data and you know.

Sneha Mehra (01:39:39)  
Rack top K and how why it improves ⁓ the performance of rack.

So that's it. That's all we can do, and we can push the generation to ⁓ get the best possible ⁓ response that we are hoping for. We can combine it with prompt engineering and we can also ⁓ fine-tune so that it knows which retrieved documents are relevant and which ones are irrelevant. So that's generation, and ⁓ we are done with the with the ⁓ architecture and design of racks. In the next

lecture we'll learn about evaluation of rags and how should we evaluate and what are some of the important metrics.

All right, we are getting closer to wrap up RAC. ⁓ And in this lecture, we'll focus on the evaluation. As you can imagine, ⁓ there are different components involved in racks, and therefore the evaluation is more complicated. So there is not a single or multiple metrics that we can easily use to evaluate racks. There is this retrieval component and how accurate the retrieve items are items are. And then when the LLM generates something, there are different aspects, like how accurate the generation is, is it

Hallucinating, is it factually correct? ⁓ Or is the generation aligned with what was given in the context, or is it different? So all these different aspects basically makes the evaluation more complicated. But the evaluation can be described in this figure. We have query, we have the result or the the gener the generation, and then we have the given context to the LLM. And then

Sneha Mehra (01:41:20)  
We can evaluate this RAC system from different perspectives. And each perspective is involving some of these components. ⁓ So these are some of the key aspects that we generally evaluate rags. One is context relevance, one is faithfulness, one is answer relevance, and answer correctness, which is relating between the result and query. And again, by just looking at this ⁓ figure, it should it should we should have some idea about ⁓ why each of these ⁓ aspects are important.

⁓ so let's just quickly go over them. First we have context relevance. And again, here we have context relevance between query and context. So basically here what we are trying to measure is that how accurately the retrieval component selects relevant documents based on the query.

So ⁓ this is basically tries to evaluate in a ranking like system. Basically what we want to know is that if ⁓

Basically, given the query, are the retrieve components relevant or not? And this is actually measuring the quality of the retrieval component. And for that, we have these evaluation metrics, which are more used, mostly used in ranking systems. And again, because context relevance is kind of ranking, we want to know that given the correct answer, were we able to retrieve those in the context or not? These metrics are also ranking metrics: hit rate, MRR, NDCG, and precision at K.

Again, if you are interested to learn more about the details of these metrics, each of them have ⁓ really great wikis and there are some papers for some of these. And feel free to take a look to understand how they're designed, what is the formula, and so on. But at a high level, these are ranking metrics. So they try to measure how how ⁓ good our retrieved items are, given the correct answer. Or or given the query. ⁓ basically it depends on how how we define it.

Sneha Mehra (01:43:21)  
So this is context relevance and then we have faithfulness. Let's go back to here. Faithfulness is to measure if the results ⁓ are faithful to the context, meaning that if the generated response is actually coming from the the provided context or it's just something different or hallucination.

And this is a difficult type of evaluation, so it's very subjective. So humans are usually needed as as part of the evaluation, or we can rely on fact checking tools and consistency checks. Again, I have some example here, but the main goal here is that whether the generated response is factually aligned with the retrieved context. We want to know that if the generation ⁓ is aligned with this context that were ⁓ given to the LLM.

And then finally we have answer correctness. Again, let's go back to here. Answer correctness is trying to measure if the result ⁓ is correct given the query, if the generation is relevant to this query.

So here basically it's a the metrics we use are basically the ones that are used to measure the quality of generative systems like caption generation, like those kind of systems. And these metrics are mostly used for that perp those purposes. So here again we have the prompt, and then the goal is that we compare the the generated response with the correct response, and we see how aligned they are, how how close they are.

And then these are some of the metrics to compare two different sentences or paragraphs and measure how similar they are. So this is all about evaluation of rags ⁓ and ⁓ different aspects of it. And for each aspect, we also saw various ⁓ metrics. Again, if you're interested, for most of those metrics, they're they're ⁓ dedicated papers and lots of tutorials online. If you're not familiar with, you can just review those. But this is how we evaluate rags systems at a very high level.

Sneha Mehra (01:45:32)  
Hello everyone. This is the last lecture of week two. And in this lecture, we'll just put everything together and ⁓ review the overall system design of a RAC system. And it's basically just ⁓ we are basically just putting together everything that we've learned. So we saw the indexing process. During the indexing process, we go from our document database, and these are the internal knowledge base for our

retailer store, let's say it's a bunch of PDFs and it could be in different structures or like HTMLs, markdowns, anything like that. We have document parsing and chunking. And then we apply a text encoder or ⁓ text encoder and image encoder. These are coming from a model like ⁓ OpenAS clip model, ⁓ which have a shared embedding space both for text and images. And then we build these indexes, ⁓ indices for ⁓ index for text and index for images.

And this is just a table. It's nothing ⁓ it's not too complicated. It's just a table with ⁓ ID of the chunk and the actual embedding of that chunk. So that is the indexing process. We do it once offline on our data and we build this. Now, every time there is a query by user, first we have some ⁓ guardrails again here, or we can call it safety filtering. There are different terminologies. These days it's mostly referred to as input guardrays.

And we have also another output guard layer guardrail. ⁓ But again, the purpose of these two is to just ensure that the system is going to be safe, meaning that if there is a request that is not going to be safe to process, we the system do not ⁓ assist with that. And similarly, if the request is safe, but somehow the output is not safe or it's not it's biased, for instance, it would ⁓ the output guardlays or safety filtering would discard the request and not.

show the re the final response to the user. So these are these two companies, these two components. Then we have ⁓ query rewrite ⁓ or also query expansion. The purpose of this is to ⁓ improve and enhance the provided query by the user because a lot of times the initial query may have some problems, it may be ambiguous, there might be missing punctuations, there might be some ⁓ grammatical problems. So

Sneha Mehra (01:47:58)  
This is what this component does. And then after we have this final ⁓ enhanced user query here, first it goes to the text encoder to get the embedding. Then this nearest neighbor search, which is typically an approximate nearest neighbor, ⁓ finds all most closest embeddings to this ⁓ user query, like enhanced user query coming from here, and find all those relevant chunks. And that those can be text chunk or image chunk. And then all of them are passed.

As the retrieved content to the generation along with the enhanced prompt. So first prompt engineering, ⁓ add whatever like roller specific prompting, few shot prompting, chain of thought, and make it a very nice comprehensive prompt. And it would also include all the retrieved content. And then the retrieved content would go into the LLM. Now, this LLM could be a general purpose LLM, or we can fine-tune that ⁓ using Raft method to have a more ⁓

accurate LLM. But regardless of that, we'll just use our LLM here, either the general purpose one or the fine-tune one. And then the LLM would just ⁓ process the given prompt and output. And the re final response is shown to the user. So this way we have control over this LLM, meaning that we can just whenever we have more PDFs or more ⁓ new knowledge in our knowledge base, we can just add everything here. We can index it. And then from that point on,

Anytime there is a user query, this whole system would take into account those newly added documents as well. So it's very easy to just add or remove some documents from this system. And similarly, rags are also useful to get ⁓ live data. And we'll learn more about that in the next week when we talk about tool calling and search. But basically still, RAG can be very, very useful.

Because when a request comes here, if we have another component which searches the web and gets some live data, ⁓ in a very similar way, that live data can also come ⁓ get to this retrieval component and relevant parts can come here. And this way the LLM would always have access to live data. But again, we'll learn about search and tool calling in in the next week. And this week we can just wrap up our ⁓ rag ⁓ topic. So

Sneha Mehra (01:50:20)  
Yeah, that's it.

—----------------------  
3  
Sneha Mehra (00:00:00)  
Hey everyone, welcome back. It's great to have all of you again and see you. I hope you had a wonderful week. I also hope that you have enjoyed your ⁓ guided learn guided learning. You found it useful ⁓ and as well as the project. So today we are gonna officially kick off the second week ⁓ and ⁓ we'll start with a project deep dive of week one. So ⁓ I'm gonna start ⁓ and

In terms of questions, we'll take questions at the end ⁓ following the raise hands. So ⁓ first we'll just go over the ⁓ project one deep dive. We'll go ⁓ cell by cell and try to explain everything. So if you didn't get a chance to work on it, that's still fine because we are just gonna go over everything and explain it. ⁓ so I guess that's it. And one more thing before we start. Project one was more focused, like the entire week one was more focused on.

⁓ learning the foundations, like LLMs, or how some of these things are work under are working under the hood. And then similarly, the project was designed for that purpose. So the project has had to go a little bit ⁓ lower level, just so we all get a sense of you know, what is tokenizer, how it's really built when we talk about tokenizer, it's really just a list and things like that. But the starting project two, which is released today, ⁓ you would see we'll focus less on the low-level details and focus more on

Like application level and abstractions. ⁓ So that was a note for project two, and we can start now. ⁓ So project one was building an LLM playground. ⁓ And there are two ways that we could ⁓ run this. One was the Google Collab, which is freely available to us. It would also give free GPU so we can load ⁓ heavier models and play around with them. The other way that we could

run and ⁓ complete this project was to run it on an on a local environment. So ⁓ the collab is very straightforward. We just click on it and it would open on Google Colab. Then we can change the acceleration to GPU and then we can just run the cells. It doesn't require any specific ⁓ installations. Whereas if you want to run it ⁓ locally then ⁓ you need to also install certain requirements, which we already provided a YAML file.

Sneha Mehra (00:02:23)  
So for this deep dive I'm gonna just run everything locally.

Sneha Mehra (00:02:31)  
⁓

So ⁓

Sneha Mehra (00:02:39)  
⁓ so yeah, I'm good. Sorry, I I ⁓ I had to check a setting. So ⁓ apologies. So ⁓ yeah, we are going to ⁓ focus on ⁓ the local

sorry, sorry, give me one second. I have to change the setting to admit everyone.

Sneha Mehra (00:03:40)  
All right. ⁓ so yeah, so the first step is ⁓

⁓ importing all the libraries. So we have Torch, we have Transformers, and we have TikToken. Torch is more for ⁓ basically it's meta's framework for ⁓ implementing neural network libraries. So a lot of the things that we've learned in the lecture, such as ⁓ linear layers, convolution layers, and a lot of those and already training and also training and optimizations are already provided.

And they are as part of the Torch library. So ⁓ again, we don't need to go over all the details of Torch. It's a huge library and it's not really ⁓ nobody's aware of every single detail about ⁓ Torch. But we'll just focus on some key ⁓ libraries and aspects of Torch. And then after that, we have Transformers. Transformers is basically the hugging face library.

And then it's creating an abstraction so that we can easily load some pre-trained models.

Sneha Mehra (00:04:51)  
And then after that, we have TikToken. And TikToken is just ⁓ the OpenAI's ⁓ tokenizer ⁓ library. So we can simply use whatever tokenizers that have trained and then ⁓ just use them. So the first step here is to ⁓ select my ⁓ environment. So I've already installed using conda, and then in the second project, we've provided the instructions for you to ⁓

to basically just use this ⁓ conda environment and install it and then add it to the to your ⁓ jupyter ⁓ notebook environment. But I've already been through those, so I'm just gonna select this here, ⁓ LLM Playground.

Sneha Mehra (00:05:42)  
And now it's just starting my my ⁓ kernel. So now ⁓ this is just for ⁓ importing things. So here I've been able to import these libraries. ⁓ And these these are just the versions. I think the latest torch version is 2.8, but ⁓ it's fine, still we can use the ⁓ this 2.7. ⁓ all right, so the first step was tokenization.

Tokenization, visa basically is a way to go from the raw text into ⁓ the numbers. And then it's a process of converting this. And it has two phases. The first phase is to ⁓ build it or train it using our training data. And then the second phase, the second phase is to use it ⁓ for inference. And here we are gonna learn how to really ⁓ build our tokenizer. And we'll try a couple of different tokenizations.

The first ⁓ step is a word level tokenization, which we already saw. It works by splitting the sentence based on their white spaces. ⁓ And this is just our corpus. This is a toy example. In practice, we don't have ⁓ we don't have basically just three sentences. We have the entire internet data. So ⁓ this is just an example. And obviously, if we work with this, we would have very limited number of tokens in the vocabulary.

But again, in practice, there would be thousands of unique words.

Sneha Mehra (00:07:18)  
So here we have ⁓ the pad and unknown. For pad and unknown, for now we can ⁓ just ignore them for a second and we'll get back to this.

Then we have vocabulary, word to ID, and ID to word. So these are some of the things that we are going to build. And just by ⁓ just this vocabulary is going to be our tokenizer. This is what we are going to build. And it's going to ⁓ be built from the corpus.

All right, so ⁓ let's see how to implement this.

Sneha Mehra (00:08:01)  
Basically, the simplest way to implement this is to go over all these ⁓ sentences here. And then ⁓ from each sentence, we split them by their white space and make sure that they are unique, meaning that there is no duplicate, and just add them to this word, to this vocabulary. So for doing this, I'll start with first creating a set.

Sneha Mehra (00:08:29)  
And set is basically just a something that provided by Python. And it's just ⁓ a collection of things. So it's very similar to list, except that it's a collection. There is no inherent order. And again, if you're interested, many of these can be ⁓ searched online, like Python set.

Sneha Mehra (00:08:49)  
And then you would see like ⁓ how it works and the data structure and some of the key methods like ⁓ you know adding, copying or things like that. So again, it's not expected to ⁓ know all of this. So many of these can be just ⁓ search online.

So ⁓ this is where we create the set. And then now the next step is to go over all the docs ⁓ like this for dog in corpus. And then here, ⁓ now here, so let me comment everything here so we can keep printing it.

Sneha Mehra (00:09:30)  
All right. Now if I print everything.

It's nothing. Cool. So ⁓ so here I would just go over all the docs.

Sneha Mehra (00:09:45)  
For document corpus.

Sneha Mehra (00:09:50)  
If I print the doc, you would see the sentences here. The quick brown fox jumps over this. Then tokenization converts text to numbers, and then basically it's three sentences that we had here. Now, ⁓ what I need to do is now for each of these, I need to just split them by the by the white spaces. So next I would do ⁓ like words equal doc dot split.

And again, if you search a split, Python split, basically it's a way to divide a string.

Sneha Mehra (00:10:29)  
Based on some separator. So the separator can be anything. There is some default value. And then when you run it, you would see that the entire sentence is separated by those. So

If I print now the words, we will see each word individually, which is fine, which is which is what we want. Basically, these are the individual tokens for word-level tokenizer that we aim to get. Now, ⁓ the next thing is basically the next thing is the lowercase uppercase. For example, we have there here, we have another there here. And then this can be different from this just because the T is uppercase here and it's lowercase here.

In word level tokenizer, it's very typical to ⁓ just lowercase everything. So we don't make any difference between this te and this te. So this is what we are gonna do. So that here, for example, we don't get this this te and this te would become ⁓ similar. So for that, it's very simple. We'll just need to use the lower dot split. So if I print this now, we'll see that everything is lowercase.

⁓ and now we have every word ready. We can just add them to the set. So if I ⁓ so now basically if we go for ⁓ w in word, we can ⁓ just do ⁓ word let me change this ⁓ word I would call this local word.

Sneha Mehra (00:12:13)  
And then here words dot add w. Now if I print this word, it basically keeps ⁓ it basically would keep adding ⁓

All these words to the set. So ⁓ here I print the words at the very end.

Sneha Mehra (00:12:37)  
Okay, for W in in local words here, we add them to the to the words, and then here we print them. So here you would see we have this set with all those individual words. Now, this was a ⁓ very simple step-by-step implementation, but in practice we can also have everything in a compact ⁓ version. So instead of having these three lines, we can just simply have ⁓ So here we can do basically first we can. So let's do this. Let's

So this entire thing.

And then

We also needed to ⁓ if it's a function, we also need to return it here, but in this case we don't. So now ⁓ instead of this entire thing, what we can do is we can

Simply write ⁓ or doc in corpus.

Sneha Mehra (00:13:37)  
We do words.update. And then update is another method provided by Python for set. We can just keep adding new things. And then here for this doc, we would say doc.lower and then a split it. So basically this is going to be a list. And this entire thing would be just ⁓ added. And it would update the words. So here if I print the words, we would see the same ⁓ list again.

Sneha Mehra (00:14:08)  
So ⁓ all right, so this is how we can ⁓ build these words here. This is a simple form, this is a compact version of it.

And next thing we want to do is now we want to form our vocabulary. So the reason that vocabulary we have it ⁓ separate as words is because in vocabulary, a lot of times we also add some special tokens. And what we mean by special tokens, these are some of the particular tokens that we would see two of them now, but in future because we would see also more for reasoning capabilities and like end of sentence, like separating images from textual input and so on. But for now, we'll just have two very simple ⁓ special tokens.

One is pad, one is unknown. Pad, I'll get into it shortly. And unknown is basically ⁓ so I'll also get to that shortly. So keep these two special tokens for now. So to build the vocabulary, I have I have the words and I have also the ⁓ the ⁓ vocab ⁓ the special tokens. So I can simply write here ⁓ basically pad.

and token and these two is basically forming a list and then I would add the ⁓ words. So now if I print work here, I would see all the words as well as the special tokens.

⁓ this was pad. Pad and one.

Sneha Mehra (00:15:44)  
And in the list, not set. All right. So it's it's complaining that words is not list. So I would just make this a list. So now we'll see all these ⁓ special tokens as well as all the ⁓ tokens that we extracted from our corpus. And everything right now is in a list. And that list is called vocab. So this is called the tokenizer. I mean, whenever people are talking about tokenizers or like companies release some train tokenizers, it's just this vocabulary. ⁓ And ⁓

Once we have this vocabulary, it's a it's a list. So we see it here. So naturally they come with indices, and the indices can be used ⁓ to associate some IDs to these tokens. For example, path is ⁓ located in the very beginning of this list, so its ID is zero. Unknown is located next, so its ID is one, and so on. So again, this vocabulary is all we needed to form our ⁓ tokenizer.

From this very, very tiny corpus.

Now, from this point forward, it's more for helpers, ⁓ so that we can easily use this tokenizer. ⁓ We can basically create these dictionaries so that anytime we have a word, we can easily get there get its corresponding ID. And also when we have an ID, we can easily get its corresponding word. So how to build these vocabularies? It's also very simple. We can iterate over our vocabulary that we built.

So again, here if I print W, you would see all these words.

Sneha Mehra (00:17:21)  
And then here I just need to ⁓ I also need their indices. So I can rely I can use the enumerate of OCAP. And then here I can print basically I and W. And now we would see both indices and their words.

And now we can ⁓ fill these two. For ⁓ word to id, we can set ⁓ word to id ⁓ of word equal ID.

Sneha Mehra (00:17:55)  
Then we can also say ID to word of ID equal word.

And now if I print this word to ID, we should see the dictionary.

Sneha Mehra (00:18:11)  
So see, here is the dictionary. We have pad, it's its ⁓ corresponding ID is zero. Then we have unknown, its corresponding ID is one. Then we have fox, its corresponding ID is two, and so on. And then in total, we have 20 ⁓ unique tokens. And same for ID to word. Now if I print ID to word, and let me ⁓ remove this print.

So now we'll see we can easily go from IDs to pad.

So these are more helpers because now we need also pro write these functions to encode and decode ⁓ any given sentence or any given sequence of IDs. So these two dictionaries can be very ⁓ important to build. And again, in most ⁓ tokenizations, when they are trained and provided, these three are already there.

All right, so we have these things. Now the next thing is okay, just a quick ⁓ prints here to see everything is fine. It says we have 21 words and then these are the first 15 vocabularies.

Now the next thing is demo. ⁓ Sorry, encode and decode. So let me uncomment this. So for encode, the purpose is to now go from this text into a ⁓ sequence of IDs. So ⁓ we have this ⁓ word to ID. So that's going to be helpful to implement this. So all we need to do is now we need to go over all the words in this text and then ⁓ extract their ID ⁓ and return them in a list.

Sneha Mehra (00:19:48)  
So again, I'm going to first implement this a step by step, and then I we can make it in a compact version in in just one line. So first we create these IDs, then we go over ⁓ each word in the text, and again we would lowercase it as as before we saw and split it to get a sense, to get ⁓ the board. Now if I print each word, ⁓ and then for testing, let me comment this still, and then print encode.

⁓ like ⁓ fox jumps.

Sneha Mehra (00:20:28)  
It's not gonna do anything at this point, but just print the fucks and jumps based on this code. So we have fucks and jumps.

Sneha Mehra (00:20:38)  
⁓ all right, so we are good. And now the next thing is now to extract our IDs now and then append it to the to the list that we created here. So what we can do here is we can say IDs.append and then for the word we can say ⁓ basically word ⁓ so we have this word and then we need word word to id of word. Now this has this has issues. I'm gonna get to that in a second, but

for now let's just extract this ID and append it to IDs. And here we are gonna return all the IDs. Now if I print this. ⁓

Sneha Mehra (00:21:24)  
Return IDs. Now, if I print this for this sentence, it should give me two IDs here fucks and jumps.

Okay, so it's giving me two and five, which makes sense. Fox, ⁓ its ⁓ ID is zero, one, two, and then jumps is five.

Now ⁓ all right, so we have this. Now there is one issue. So for example, if we do fox jumps ⁓ somewhere.

There would be an error. And the reason for that is because there is no token called software somewhere in our vocabulary. So it doesn't exist in this word wid. And the typical way to handle this is now ⁓ to ⁓ create a special token and we call it unknown. So basically we call we create this and we call and we put this here so that at inference time, whenever there is a sentence that it has a word or token.

Which is not known to the vocabulary to the tokenizer, it can be just replaced by this special token. And this is going to be very useful for both inference time and training. Because even in training, after a while, we may come up with new data. And some of these new data may have tokens which were not previously used to train the tokenizer. So that's the purpose of unknown. Now, ⁓ in this case, what we would do is we won't say, we would say if word ⁓

Sneha Mehra (00:22:53)  
In word to id. Then we would say id.append word to id of word ⁓ else. If it doesn't exist, we would simply id.append ⁓ word to id ⁓ and our special token, unknown, which here is unknown is this guy.

Sneha Mehra (00:23:20)  
So now we should have good IDs without any errors. So now if I run this, we should still get two and five. And for somewhere, I would get the ⁓ the ID corresponding to the unknown token. So now if I print this, there are some errors. ⁓ And

Sneha Mehra (00:23:41)  
So if word in word to ID. ⁓

Sneha Mehra (00:23:52)  
now I have to remove this ⁓ too. All right. So we get two, five, and one. So two is fox, five is jumps, and then one is unknown. And that's because there was somewhere, somewhere was not in the tokenizer. So this is the the right way of implementing the encode. Now this can be also ⁓ implemented in a more compact form. So let me comment this just if you're curious to see how the compact form would look like.

Sneha Mehra (00:24:23)  
So what we can do is we can do return. ⁓ and then we'll get word to id. So instead of extracting the so before this, I'll do for ⁓ w in text.lower.split. So at this point we have W's, which are the lowercase individual words. ⁓ so now here we need to get their their ID. So we can do word to ID. Instead of doing like this, ⁓

We can use Python's dictionary get method. So it returns the value if the key exists. Otherwise, it would return a default value, which is exactly what we want here. So here we can do ⁓ W. And then if it doesn't exist, we would say word to id ⁓ of unknown. So now this code should also return the exact same sequence here. So let me run it. All right. So two, five, and one.

So this is how the encode works, and most ⁓ the state-of-the-art tokenizers are doing something very similar. They just encode the provided text by doing something like this. Of course, they have more arguments and parameters here, but at the very high level, this is how they're being implemented. So ⁓ I'll keep this, but we'll move forward to I'll remove this for testing and then I'll move forward to decode. And decode is basically very similar, just the opposite of this. So I'm not going

⁓ to go over it. Just the main difference is that now we are going from a sequence of IDs to ⁓ an actual text or sentence. So we just need to do ⁓ opposite of that. So how to implement that? I'll ⁓ start with a ⁓ let's say text. First it's going to be empty.

And then I'm going to go over ⁓ all the IDs or ID in IDs. And I'll get the word. So first I'll have a list of words here. I'll do just append words.append. ⁓ now we'll go id to word ⁓ of ID.

Sneha Mehra (00:26:41)  
So this is good. And now we have all the words. So let me print it here. ⁓ if I if I print words here and then I do decode. I'm going to send the exact same sequence to see what happens. Two, five, and one.

⁓ all right. So so far we printed the words. And if you see the words, basically it's fox, jumps, and on, which is which is expected. Now we need to we don't want a list. The code basically is supposed to convert this sequence of IDs to a to just a string, a text. So we need a way to combine this. And the the typical way to combine this is just using a join ⁓ operation provided by ⁓ Python. So we can do ⁓ something like this. We can do a space dot the join.

And then we pass all these words. And then this can be our final text. So it turns out that we don't need this. And now here we can return the text.

So here we see Fox jumps unknown. And again, I write the compact form here.

It's going to be return.

Sneha Mehra (00:27:57)  
This dot join id to word. ⁓ I or I in IDs. ⁓ and then ⁓ so this should work, but also one thing that we can do is remember we had another special token here called pad. We don't want to include those paddings. I'm I'm going to go over the pad and why it's there, but for now, ⁓

At least for decoding, we never want this special token to appear when we are going back from a sequence of IDs to ⁓ the the original sentence. So ⁓ so here we can do this ⁓ and continue it by saying if i is not equal to word to id of pad. So we just want to make sure that if it's not the pad token, then we're just gonna use it to reconstruct the original sentence. Now, if I run this, I'll get fox jumps unknown.

And then if we have at any point the ID corresponding to pad, ⁓ and let's say another here, we would still get the fox jumps on. So ⁓ we are good on encode and decode, and this part is now complete. I can remove this, we can uncomment this, and we can run it.

So we'll see if we have 21 words, we have all these ⁓ unique tokens, and then we have ⁓ basically the input text, which is the brown, which is the brown ⁓ unicorn jumps, and then we have token IDs, which are the corresponding tokens. Now, one thing to note here is that we have unicorn, and unicorn was not part of the ⁓ our vocabulary, so it it got replaced by unknown when it's decoded.

And that's typically a problem. Ideally, we don't want to see too many unknowns because ⁓ it's just losing the information for model to learn from. So, for example, for the brown unicorn jumps, if if we don't have unicorn in our vocabulary, anytime we are gonna train on sentences with the word unicorn, it would get replaced with unknowns. Now, if there are a lot of unknowns, the LLM would get confused and cannot learn really the statistics and the pattern between these words. So this is not ideal.

Sneha Mehra (00:30:13)  
But it's the best tool we have to make sure that at least at inference time, we are not the decoder is not gonna fail. It's just gonna replace it by unknown. And again, we can also modify this implementation to make sure that if something is unknown, it won't print it. So we can just continue adding this ⁓ if I not equal this and I not equal the ⁓ the unknown token. So that's it for this section of word level tokenizer. So you just

Created this vocabulary, and that's all we needed to do. First, we needed to decide on our special tokens. And in this case, we had two special tokens, unknown and pad. And then ⁓ we just built it. We just used our entire training data. We created the words. And then from the words, we built our vocabulary. And then once we had the vocabulary, the rest was just helper functions to make sure that we can encode and decode.

So that's it. Now we are gonna go to the next part, which is the character level tokenization. And at its very core, it's very similar to word level tokenizer. Now, ⁓ one issue that we we observed here is that each word is going to get its unique ID. So there would be a lot of ⁓ unique words in our vocabulary. And that's not ideal, especially if you are training on internet data and lots of different languages. This may become like

Hundreds of thousands of vocabularies. So that's not efficient. And also, if we train on internet data, still there is going to be a lot of unique words that were not in our tokenizer. So we will we may end up with a lot of these unknown tokens during training, which is also not ideal. So character level tokenizer, it's going to be ⁓ similar. What we need to do, ⁓ again, let me comment here.

And then we start here and we form them. We already have the corpus. The corpus is the same. This is our corpus. We just now need to run character level tokenizer, meaning that we want to split this based on each character, and each character would have its own ⁓ ID. And we also need to make sure that it's it's a set, meaning that there is no duplicates. So let's just start implementing this. Again, I'm going to create this

Sneha Mehra (00:32:31)  
Letters first. I mean there are different ways this can be implemented. One is we can simply write a, b, c, d if it's English only, a b c d h i and the rest. ⁓ the other way we can do this is we can rely on this string library provided by Python, and we can just ⁓ do ⁓ like a string dot ASCII lowercase. So this should, I believe, this should give us.

all the lowercase ⁓ characters in English. Now if I print this, we we would get A, B, C, D and all the way to Z.

Sneha Mehra (00:33:12)  
We can also get ⁓ uppercase. So we can if we do a string dot ast uppercase, now we would get all the letters, all the English letters.

And again, this is a very simple example in practice, even for world-level tokenizers. You may want to go over all your corpus and do something very similar that we did here, just split them ⁓ character by character and form your unique vocabulary. But here I'm just gonna use these lowercase, uppercase letters from a string. And again, we're gonna have our special characters. ⁓ I call them ⁓ special here and then pad.

Sneha Mehra (00:34:04)  
Unknown. So these are our special characters. Now vocabulary is going to be these two.

Sneha Mehra (00:34:32)  
I just need to make this a list. And now if I print it, I would get a list of ⁓ all basically the unique ⁓ English characters as our vocabulary. Again, there are different ways. You can ⁓ it depends on the design. We can we can always start from our corpus and build ⁓ our unique characters if there are multiple languages or we want to include like numbers and ⁓ like ⁓ exclamation points, question marks and all those things.

But for now, we're just intentionally gonna keep it simple and just use ⁓ English letters. So now we have this vocabulary and we are good. We no longer need this. Now for character to ID and ID to character, it's very similar to what we saw before. For ⁓ car car to ID, we have this. ⁓ so again, I'm gonna use the compact form. I'm going to ⁓ do like

⁓ First is it's a dictionary ⁓ and I'm going to do for IDX and characters ⁓ in all the letters ⁓ in our ⁓ WOCAP.

Sneha Mehra (00:35:43)  
And then here I'm going to say your key would be because it's car ⁓ car to ID. So the keys are characters and the IDs are the the IDs, like where they are located in this vocabulary.

Sneha Mehra (00:35:57)  
⁓ so we have character here and then the value would be the IDX. Now if I print this character to ID.

Sneha Mehra (00:36:08)  
All right, it's basically giving me okay, this is a special token, it's a different story, but the rest for each character now we are getting their corresponding ID.

And then in total, I think we should have ⁓ 53 plus ⁓ one fifty-four tokens. And that's because we have all the English letters and ⁓ two special tokens. So this is good. Next we want ⁓ ID to care. And for ID to care, there are two different ways we can do it. We can either again go over the vocabulary or we can just go over this care to ID that we just built.

So I'm gonna go over this car to ID for CH and IDX in ⁓ char to id.items. So this would give me ⁓ like the items ⁓ over a dictionary would give me the the pairs of key values. So which in this case for character to id, we have character and ⁓ the index. So for all of these, now we just need to swap this because now we are going from ID to character. So we have IDX ⁓ and ⁓

So here we have IDX and CH. So I think this should work, hopefully. Print ID to care. All right. So we'll get zero, pad one, unknown, and so on. So we have these two now. And let me just remove this. And these are working. ⁓ Now, if I uncomment this and I run it, I'll get the vocabulary size 54 expected.

Now encode decode. Let's start with encode first.

Sneha Mehra (00:37:54)  
So ⁓ here again, we are given the text. We just need to iterate over the text and just retrieve its ID and return it. So for that, I'll do again the compact form, return ⁓ a sequence of this for ch in text for each character. We have ⁓ car to ID. So we can just say car to ID of ⁓ CH, but this is not ideal because we want to also handle the unknowns. So here I'll do get.

And then here I would say get ch ⁓ get ch ⁓ if it doesn't exist. I want the unknown. And for unknown, we have car to id of ⁓ unknown.

Sneha Mehra (00:38:43)  
So I guess this should work now. So if I print in code high.

Sneha Mehra (00:38:54)  
Hi. Hey, how are you? ⁓ and I printed. All right, we'll get an error here. Name cat to D is not defined, so there is some misspelling.

Sneha Mehra (00:39:11)  
Now unknown is not defined.

Sneha Mehra (00:39:19)  
Okay. So now I have everything for like a sequence of IDs for this ⁓ given sentence. So basically now we are able to encode this. We build a token ⁓ character level tokenizer here, and also we created this. So now we provide this functionality. In encode, we can anybody can just use my in code function, pass whatever sentence they want, and they'll get for this sentence a sequence of IDs. And again, if you watch the the

The guided learning, you see that now we are getting a very longer sequence compared to word level tokenizers, just because now for each character we have a ⁓ a separate token. And this is not also ideal because now the LLM should look at all of these and learn ⁓ the patterns among all these very long sequences. So it's not efficient and it's more difficult. But again, this is the in-code and how it's typically implemented. Now let's go to the decode.

Sneha Mehra (00:40:22)  
So decode is kind of just the opposite of that. If we do return ⁓ so here we don't need to ⁓ join by a space because we did not ⁓ split. It's not a word level tokenizer. We didn't split by the white space. We just we just went character by character. So we just need to add them up. So I'll do ⁓ just an empty string.join. Now what do we need to join? Just a list of ⁓

Corresponding characters given these IDs. So we go over over all these IDs for I in IDs. And we make sure that if an I is not equal to car2 ID of ⁓ pad. I'll get to this pad in a second. ⁓ and then if that's the case for all these i for i ⁓ in ids, if i is not this, so here we need their character, their corresponding character.

ID to car of ID. And now we are joining them all. So now if I do print decode and pass a sequence of ⁓ IDs like this one. So for this one, we already know what the output should be, what the reconstructed sentence should be. If I print this, I should get hi, how are you? Hi, how are you?

Sneha Mehra (00:41:49)  
All right, so here there is a problem because for all the spaces, we are getting this unknown. And that's because when we built our tokenizer, we didn't have any special token or any tokens for a space. So this can be also handled in different ways. We can ⁓ perhaps add a space here to our vocabulary or to our letters just to make sure this can also be handled and rerun this. When we rerun this, we are basically rebuilding our

tokenizer, ⁓ vocabulary and all these dictionaries. So now for this sentence, let's ⁓ see if we get ⁓

Sneha Mehra (00:42:35)  
Okay, so ⁓ first I have to encode it again. Print encode. Hi, ⁓ how are you? And notice that here now I'm including all these white spaces. Let me also add a question mark here. But but question mark is not part of our vocabulary. So it's gonna be replaced by unknown. So if I print this, we'll get this sequence of IDs. Now if I replace this here for decoding.

Now I should get the original sentence back. Hi, how are you? But again, the the question mark is now replaced by on. So ⁓ that's about cartel level tokenizer. And then let me uncomment this part. Again, these are once implemented, are generally very easy. Again, the purpose is just to show that it's not like tokenizer, it's just a list and a vocabulary. It's nothing more. ⁓ so let me now ⁓ run this. So

⁓ we have hello, we have its corresponding token IDs, the sequence of characters, and then the decoding is also hello. So that's that's working. Now we'll switch to subboard level tokenization. ⁓ And ⁓ also we mentioned in the guided learning, board level tokenizer and character level tokenizer are not really used in practice. They are having their own issues. We went through those. ⁓ And in practice, most of the state-of-the-art LLMs are.

Relying on subordinate tokenizers. And there are different algorithms, like a common one is BPE, byte pair encoding, or some variations of it. And there are some other implementations or different ways of sub-board level tokenizer. ⁓ And the main idea behind it is to ⁓ start from very small ⁓ bytes or characters and then iteratively merge the most frequent pairs. So it basically starts from ⁓ every character and it becomes its own token.

And then it would count the most frequent pairs in the in the corpus. And if there are pairs that are very frequent, it would just merge them and create new tokens in the vocabulary. So it just iteratively increases the size of the vocabulary. And that's how many tokenizers get into like 50k or 100k tokens. So I guess I have this ⁓ tick tokenizer here. So here there are a bunch of tokenizers provided by.

Sneha Mehra (00:45:00)  
⁓ and trained by open air. And this one is for example the one with 100k ⁓ tokens. And I believe they're using the variations of BPE. If you go to their tick token library, we would see like ⁓ each model was trained with each tokenizer. And then their more recent one is ⁓ 200k. So they're having ⁓ 200 tokens after they are training this their vocabulary, their tokenizer. So their vocabulary list has 200k items. ⁓ So hi how are you?

⁓ we'll get into this. So we'll get into this very shortly. I just wanted to show the tokenizers and how it's different. It's not really ⁓ a splitting this based on words, it's a splitting this based on ⁓ the created and ⁓ merged tokens that it was trained on or it was learned during the training. So we'll get into this in a second.

Sneha Mehra (00:45:58)  
All right, so ⁓ server level tokenizer ⁓ is after it follows this algorithm, the good thing is that for less frequent words or unusual words, they won't end up with their own token. For example, unbelievable would become something like this, like a smaller ⁓ set of characters which are more popular or common in English. And that's how it's not it's not decided by us. This is this is just decided by the ⁓ the corpus and the algorithm.

The algorithm would just go over the corpus and come up with most frequent ⁓ like subwords in the training data. And this is just an example. I don't know if ⁓ I don't know if for unbelievable on believable. Yeah. So unbelievable, for example, in BPE in OpenAI's O200K base tokenizer, it gets divided into three different tokens: UN, B L N.

Even though so ⁓

That's ⁓ and one more interesting thing, for example, if I write hi, how are you? Each of these meaningful, very frequent words are becoming one token, which is great. Now, if I have a misspelling here, like how with double O, the algorithm, the tokenizer, would just know that this how is one token, and this is just becomes in just its own separate token. So instead of turning this entire thing into unknown, ⁓ it's extracting the meaningful part.

To a token, which is awesome because now LLM can see this how, even though there was some ⁓ like mistakes or miss spells, misspells here. And then W, which is just a pure noise, would appear here, which the LLM is just need to learn to ignore some of the tokens if they are irrelevant in sequences. But just showing this example to say that what are some benefits of this subordinate level tokenizer.

Sneha Mehra (00:47:58)  
Now going back here, for BPE, there are lots of different implementations available. And the implementation is usually more required requires more lower level coding. And basically you just need to follow these three, four steps. But we don't have to because the there are libraries that are already implemented those and made those available to us. So ⁓ one of those libraries is we can just easily rely on the ⁓ so transformers. I mentioned that it's the Hugging Face library.

And it can it give you access to open source models. And as a result, since you have access to open source models, you also have access to their open source tokenizers that were used to train them. And very simply, we can just import this auto tokenizer and then we can specify the model that we want its corresponding tokenizer. So this way we are going to ⁓ this way we are going to ⁓ just load GPT2.

So we can have something like BPE token or tokenizer and then do equal autotokenizer.from pre-trained. Here we can just write a model. And if you are interested, I think hugging phase auto tokenizer models or something like that. I think it would just give you a a list of ⁓ all the available models.

Sneha Mehra (00:49:38)  
So I'll try to ⁓ find this. But basically there is a page where it just shows all the models that you can easily ⁓ get their model or get their tokenizer. I think it's not.

Sneha Mehra (00:49:54)  
Yep. I'm going to share it, but basically it's like this. ⁓ for any tokenizer, we can just do autotogenizer.from pre-train and then just pass a a reference to the model that we want. In this case it's a vision language model, but ⁓ it could be any LLM or any pre-trained model. So we'll just here gonna use the GPT two.

Now ⁓ again, let me comment everything.

Sneha Mehra (00:50:22)  
And then say print BPE token. It would just ⁓ autotokenizer is not defined.

Sneha Mehra (00:50:42)  
Right, there was a misspect. So when we load it, we get we get this entire tokenizer object used for GPT2. And you can see it here if we if we print it. It says the vocabulary size was 50,000 something. This is the max length, ⁓ padding side, and so on. by the way, I forget to say the pad. So the pad is there as a special token because when we train LLMs, we don't typically send one example. We send a

batch of examples. Because sending one example is not is it's not efficient. It has also ⁓ some other issues. So we typically batch them, send ⁓ just put multiple examples and just send them as a one input to our model during training. And these models, LLMs or most neural networks, require fixed dimensions as their input. So if we have multiple sentences with different length, it would not just work.

Because when we convert them to a sequence of IDs, we would end up with sequence of IDs with different lengths. So the purpose of pad is just to pad the smaller sequences to ⁓ with these special tokens, just to make sure that each time we send a batch of examples, they have a fixed length. So that's that's the main purpose. And then during training, the model would just ignore this path. So going back here, so this is what we see here. So it's saying that the model was trained on sequences of length.

1000\. And then whenever there were some examples in their input where the length was less than 1000, it was just being padded. And it was padded from the ⁓ right side. So these are some of the design decisions. And here you would see the special tokens are ⁓ beginning of sentence token, and this is how they represent it, end of sentence token, and this is how it's represented the special token. And unknown. So GPT2 had ⁓ three special tokens, our very tiny.

Tokenizers we built above add two special tokens. So that's it. Moving forward, it would be faster because now we are just relying on this tokenizer. We don't have to really go and build vocabulary and things like that. We can just ⁓ write the code. And they already come with the encode. So I included this part just to be consistent with the previous ⁓ implementations. So we can just have BPE dot ⁓ umunderline took, the object that we created.

Sneha Mehra (00:53:07)  
And then we call their encode. They also come with encode and decode, very similar to how we implemented our tokenizer.

So we do encode and then we send text. So that should be it. Now we can easily encode anything. Print encode ⁓ i. This is Ali. And then we'll see, we'll get this sequence of IDs. So I has this ID, this has this ID, and so on. Let me also remove this print and remove this. And now ⁓ we have this encode ready. Now for decode.

It's again very, very simple bpe dot talk dot decode of IDs. So

decode of this sequence should give the exact same

Sneha Mehra (00:54:03)  
Hi, this is Alice. Now, the good thing here is that we can now use some more ⁓ special characters, and the tokenizer, the BPE tokenizer, would be still able to decode it and encode it. And this is what we had in the demo. Now, if I print this, we would see the input text, unbelievable tokenization powers, ⁓ is getting these tokens: UN, Bell, Eve, Ebel. ⁓ So these are some of the tokens that were.

decided by the BP algorithm, then they train their tokenizer for GPT2. So if you use a different tokenizer, it may end up with a different sequence of IDs. So if I copy paste here, copy paste this in T TIC tokenizer and use the O200K.

So here they have Eve Bell as one token, while their older version had two tokens, Eve and ABL. Again, it all depends on their training data. Another thing is it handles ⁓ certain characters the way it wants. For example, ⁓ it also converts each byte to a printable format. So that's why we see this very special character here. It's basically ⁓ in the very beginning, when they are building their tokenizer, they go over all the bytes corresponding to the ASCII characters.

And then they come up with a ⁓ with a mapping. So they go from that to a printable ⁓ Unicode character so that they can just print it. And that that's the only purpose. So that's why here we have ⁓ this space token ⁓ is having its own token, ⁓ its own ID in the vocabulary. And then for printing, basically this ⁓ space is mapped to this particular Unicode character. So that's why we see this.

And the rest are normal ones. And then here again we have something. ⁓ exclamation point has its own token, and so on and so forth. And emojis are having ⁓ being ⁓ split into multiple tokens. Again, these are we don't have to know how they are split. This is just decided by the way it's trained and ⁓ the training data. So here if I just paste this okay, this emoji is basically two ⁓

Sneha Mehra (00:56:17)  
two IDs. So whenever you see you send these two IDs to a LLM, it would just over time it would just learn that this might correspond to ⁓ like in in their meaning, something to this emoji. So that's BPE. ⁓ we ⁓ can just play around with different methods, different text later if if you are interested. But that's how we can use BPE and train our own model.

Sneha Mehra (00:56:47)  
And finally, I want to talk quickly about TikToken, which is a production lab ⁓ ready library. So this is just ⁓ provided by OpenAI and it just gives you all all you want ⁓ from different models. It's very simple to use. ⁓ I don't know their particular methods. So if we go to their library, TikToken, library, this is their Gitripo.

And then you would see this is how it can be used, import it, and get encoding of a certain tokenizer that you want. You can also get the encoding of a certain model that ⁓ if you're interested. So this is how you get the encoding, and then ink is basically your your tokenizer. You can just start encoding and decoding text. See ank ⁓ dot encode hello world. So ⁓ we'll do something very similar here. Now we ⁓

I mean, ⁓ I'm gonna just implement this very fast for name and ⁓ so ⁓ we want to try two different examples. So we have encodings equal ⁓ a list of two models. First, we are interested in trying ⁓ GPT two.

And then we'll say all right gpt2 tick token.get encoding gpt2. ⁓ and then ⁓ the other one that I'm interested to try is let me just copy paste this. ⁓ And same thing, tick token dot get encoding. ⁓

This. So now we have encodings, we have ⁓

Sneha Mehra (00:58:38)  
What is

Sneha Mehra (00:58:48)  
Alright, so now we have encodings. If I print encodings, I would get just this ⁓ list of pairs. Now I'll do for name and ink in encodings. ⁓ let's just print the name and vocabulary size. So for this I can do something like

Sneha Mehra (00:59:13)  
Piss ⁓ and then name.

And this and then print vocabulary.

Sneha Mehra (00:59:26)  
vocabulary size. ⁓ this and then ink dot n vocabulary.

And then ⁓ we'll encode some sample sentence we have there. So we'll do ids equal ⁓ nk.incode sentence. Let me print IDs to see if it's working. Correct IDs are working. Then I would say tokens equal ⁓ nc.ecode of i ⁓ or ⁓ i in ids. So here we are getting all the tokens, and now it's time to

Print everything. You can send this sentence. ⁓ Sentence splits into line of IDs just to get, just to see what is the length of the tokens for that sentence.

Sneha Mehra (01:00:29)  
And then I'm going to print ⁓ their list of ⁓ zip of tokens and IDs because tokens are their words and the IDs are the ⁓ like the list of IDs. So for each of those, I'm going over them and just printing their list. So what it's saying is that tokens and decode for I in IDs.

⁓

All right, decode is a method. All right, so for GPT2, vocabulary size is 50,000. These are the 11 tokens, and then these are the pairs of all the tokens. ⁓ and then for the more advanced tokenizer, we have 100,000, and then we have all the splits and tokens. So that's it. And then the the opposite we can do, we can just ⁓ change this to decode and just print it. So that's the tokenizer. Next we are gonna go to the language module.

Again, in the lecture, we've seen that everything is basically just a sequence of layers. ⁓ And I have the ⁓ the lecture open here for simplicity in implementation. So ⁓ linear layer is basically just these weights. So we need to have a Python class and keep these weights, create these weights, initially initialize them with random values, and keep them.

And the size of these weights depends on size of the inputs and size of the output. So a lot of times when we want to create a linear layer, we have to specify this is the expected input size and output size so that we can internally build, create, and ⁓ initialize its weights. And you can see, for example, here we have four input size of size four, and then the output size of size whatever. So the number of ⁓ the matrix for weights is basically this times this. And we are gonna implement this.

Sneha Mehra (01:02:26)  
By our own, just to get a sense of how linear layer can be implemented. But we almost ⁓ we never need to implement this. It's just for us to get a sense of it's not a big deal. It's just a few lines of code. In practice, we just use on torch's linear layer, ⁓ which is called nn.linear. But here we want to implement our own linear. So let's see. So here first, let's ⁓ print.

First let's create a layer linear.

equal linear. ⁓ that's a simple example three by four. So three is the input features, four is the output features. And then here let's just print it. Not ⁓ yes, print it.

⁓ so for now I think it's not gonna print anything because we haven't implemented. But now what we need to do is there are basically in the initialization, we just need to do two things. One is we create the weights, which is ⁓ this matrix. And also linear layer has a BIOS term. So for each each term, it would add it to the BIOS. The size of BIOS is equal to the number of outputs because each of these terms after the matrix after the matrix multiplication.

BIOS term would just be added. So just to get a sense of the dimensions. So here first we create we can create some random inputs. ⁓ let's say let's call them bits equal. Torch has various ⁓ methods. Rand n is basically creating a random tensor. So if I say random tensor of out features, ⁓ number of output features and in features.

Sneha Mehra (01:04:10)  
And then we write self.weight, just linear dot weights. It would now give us a ⁓ random ⁓ tensor of torch dot rand n.

Sneha Mehra (01:04:38)  
Mm-hmm.

Sneha Mehra (01:04:53)  
All right. So ⁓ this is the random input that we created here. These are the initial weights. These are basically what we initially gonna assign to each of these ⁓ edges, but later during training, they will get updated. ⁓ so that's about weight, very simple. And now we need to just make sure that they are parameters, they are not ⁓ just some random numbers. So we'll use ⁓ n and dot parameter to make sure.

To make sure that these weights are trainable during training. And self.bios again, it's it's all about the dimension to make sure we are right. So it's going to be torch.rand n ⁓ out features. So now we have ⁓ these two weights of the linear layer. Now we just need to ⁓ implement the forward. The forward is basically what it's just the terminology that PyTorch uses. It's basically ⁓

Specifying how the transformation should happen. Now, given some input X, ⁓ how is it supposed to be transformed? So here ⁓ we go from this input to output, and this is a linear transformation. So we just need to implement that kind of linear transformation, which is a matrix multiplication. ⁓ but again, this is where we implement like how the transformation should happen for different layers. In this case, it's very simple, it's just a matrix multiplication, and Torch already provides that. We do math mode.

And then input is x. We just ⁓ we just need to ⁓ perform a matrix multiplication of that and the weights. ⁓ so for the weight I have to transpose it by calling this t because it's out dot in. We need in dot out times out. So ⁓ we'll do this and then we'll just add everything to with self dot bias. Now that's it.

Now we have this linear

Sneha Mehra (01:06:49)  
We have the weights. ⁓ now if I create some random input, forge dot rand ⁓ one ⁓ of like ⁓ three, and I just passed this this one by three is basically this guy. I'm creating this just randomly, and then I'm gonna send it to this linear layer and the output we're gonna inspect it. So I'll send this X to the linear layer, and then we're gonna print Y.

Now y is expected to be of size one by four, because here we have a specified ⁓ four when we created this. So now if I run this.

Sneha Mehra (01:07:34)  
⁓

Sneha Mehra (01:07:39)  
Linear object has no attribute weight. ⁓ Misspelling here weights.

All right, we are just getting the output, just four numbers. So that's how linear layer is implemented in practice. But we don't have to implement it. ⁓ never. It just gives us torch gives us this functionality. ⁓ so we can just import ⁓ torch.nn, which contains all the different implementations of different layers, including linear, and then simply call nn.linear. And then here we are specify sizes. In this case, again it's three by two.

And then we create a tensor, a random tensor here, just to see if if it looks right. And we can just run it. So it's basically the equivalent of what we just implemented here. ⁓ And again, transformer is just a sequence of layers, linear layer, MLP layers, attention layers, and attention by itself has also a sequence of linear ⁓ a couple of different linear layers. So ⁓ it's nothing nothing special, it's

⁓ just a sequence of layers and you can just print them. Now we are not gonna really implement any of those because again, transformers give us access to load these models and inspect them. And this is what we are gonna do here. So here we are gonna load GPT2 and inspect it. I'm gonna write GPT2 equal gpt two ⁓ lm head model dot from pre-trained.

Sneha Mehra (01:09:22)  
And then I write GPT two here.

Sneha Mehra (01:09:27)  
So now if I run this, it would give me the GPT2 model. Now I can inspect it. If I print GPT2, we would see all the layers when it was created. It has a transformer layer. This is the module name. It has the ⁓ the word and the position embedding and the word embeddings. And then it has also a sequence of ⁓ modules, which is the transformer blocks. Now it has 12 blocks. Each block has layer norm, attention.

And another layer known an MLP. And MLP is just again by itself a sequence of layers, convolution, convolution, ⁓ some kind of nonlinear activation and dropout. Very simply, you can inspect other models, more recent advanced models, to see how they look like. But we are not gonna do that for now. We are gonna stick to ⁓ just ⁓ let's see a block of the trans the GPT2. So block equal gpt2 dot transformer.

which is here, and then dot ⁓ h, which is this module list, first block. So now if I print this block, it would give me the first block of the transformer. ⁓ it is

It is one block of this.

Sneha Mehra (01:10:48)  
And you can see it has attention linear transformation MLP. So it's basically the attention and it's decoder only. So it's ⁓ attention and then feed forward layer.

Sneha Mehra (01:11:02)  
So that's the block. And then ⁓ after this, we are, we can just print some ⁓ modules just to get a sense of everything inside it. We already saw, but just to ⁓ show a more ⁓ systematic way of printing this is to go over all the children of this block and just print them. Print name and module dot ⁓ class.

dot name. I guess it should be something like this. So ⁓ it has these names and these classes. It has each block has a layer norm GPT2 attention, layer norm, GPT2 MLP, which exactly follows this setup.

All right, so we'll go here just a ⁓ tiny example just to see how it works. We already have GPT2. We can already access to its configs when it was trained. Let me print the config here, GPT2.config. So these are the configs that were used to train GPT2. So you can see like what activation function is being used there, ⁓ the probability of drop for attention, token different token IDs and ⁓

You know, a bunch of other things. And ⁓ basically the precision of the the tensors that they're trained on, vocabulary size, 50,000. We already saw it, and so on. So that's GPT2. We can get the vocabulary size and create some ⁓ dummy random tokens as input. Just we want to test it and pass it through the GPT2 and get some output. So here we are just assuming a sequence length of eight.

So what this eight means is

Sneha Mehra (01:12:49)  
⁓ eight, eight, eight numbers. Basically we want to generate eight numbers here and pass it to GPT two.

And for now, we are just gonna generate these random numbers. And it's important to make sure that these numbers are within the vocabulary size. So we don't want any of these numbers to be outside of the vocabulary. So that's why we set this, just to say that all these numbers should be within the range. And then here we set the ⁓ the size of these dummy tokens. We are saying that it's just one example and we want eight tokens. So now if I print this dummy.

Hamps.

Sneha Mehra (01:13:38)  
And amend this, we'll see this randomly generated IDs. So ⁓ that's about it. Now we just need to pass it to the ⁓ GPT. And there are different ways. I mean, first I would just convert each of these ⁓ each of these dummy tokens to their embedding. So this is this step to go from these tokens to these vectors. This is called embedding layer.

An embedding layer in GPT two is ⁓ let me delete this cell.

⁓ is we printed GPT2. It's basically called WTE ⁓ if you print it. So WTE is what it's called. And then here we can just simply pass our dummy tokens a sequence of numbers to get a sequence of vectors. And now we need to also add this with the positional encoding. Positional encoding also converts each tokens and incorporate the position information. And then those are being added to form the input. This is the input.

This visual is not showing the positional encoding, but we typically add the position information and add them to each token. So ⁓ this plus gpt2.transformer dot ⁓ WP, they call it. And then here we just need to pass, we don't need to pass the tensors anymore. We just need the locations. So we don't we want the locations to be added. So we use a range ⁓ of

⁓ sequence then. So now this is forming our input or hidden based on the dummy tokens that we've created. And once we have this, we can now pass it to the first block of the transformer and see what happens. So we pass this hidden to this and then we see ⁓ if it works or not.

Sneha Mehra (01:15:36)  
⁓ out and then here we print. We then print the shape of the out. So what we expect is that the shape should be the same size. We created one example of sequence length. The output should be one example of sequence length, but ⁓ also includes the hidden size because the output here is of same sequence length, but has this hidden dimension, which is the same as this hidden dimension. So basically we are just passing this.

to one block of transformer to get this output. And we are gonna print the shape of this output.

So here if we print this.

Sneha Mehra (01:16:17)  
we'll get the shape of one eight seven sixty eight, which is correct. Seven sixty-eight was the hidden dimension of GPT two, eight was this sequence length, and eight, one is just because we created one example.

Sneha Mehra (01:16:33)  
So ⁓ that's it. Then the next one is I'm gonna skip this because we already printed and saw all different modules. And then finally we'll go over here. Now we want to pass this and get the logits. So ⁓ for this one, it's I have the code here. I'm gonna copy paste. It's very simple. It'll be already bent through most of them. We first we load them.

We have already GPT2. If it's not loaded, we say, hey, load it from pre from this pre-trained GPT two. Also load the tokenizer. So first we load them. Then we tokenize this input text. ⁓ And it's just one line of code here. We ⁓

We pass this text to the tokenizer, we already cite, and then we get the input IDs. So if we print this input IDs now, print input IDs ⁓ and I comment everything else.

We'll get the IDs corresponding to Hello, my name.

So basically in this example, I want to pass this hello my name to GPT too and see if we can actually get the next token. So that's that's what we are trying to do. All right, so we have these tokens now. Now we need to pass this sequence of tokens to the GPT and get their ⁓ output ⁓ hidden vectors. And this is again ⁓ like two lines of code. We just do.

Sneha Mehra (01:18:08)  
⁓ no grad because we are at inference time, we don't want to click calculate the gradients and we write logits. So we write ⁓ we pass the input IDs, sequence of IDs to GPT2, and then we get the logits. So now if we print the shape of these logits,

Sneha Mehra (01:18:27)  
We should not be surprised. It's one because it's one example. The three is basically three ⁓ tokens. And then ⁓ this is basically the final output. So it's not this. It's after applying linear. So it's mapping the ⁓ this entire thing to the right dimension for ⁓ over the token, over the entire sequence of vocabulary that we have.

So that's why we have this size here. actually we had this print already.

So now, how can we predict the next token? We saw that all we need to do, we have these logits, and these logits ⁓ are good. We just need to keep the last one and then apply a softmax to convert them to vocabulary and just print it. So ⁓ this is the code. We have the logits, we get the first example because it's just ⁓ we have only one example. We get this. Then we get the ⁓

Last hidden vector. So there are three hidden vectors. We index the last one. So this basically means that we are only keeping the ⁓ vector, the output of linear for this, the mapping of this. We are just discarding all of this. So ⁓ we get this next token probabilities, but it's logits. Now we need to apply the softmax to convert them into vocabulary, into probabilities. So

Here we would we would get a list of vocabularies of vocab size, 50,000 something. And now we won't print it because it's just a lot of numbers. We would get the top K. Torch has also this top K. So we can just say, hey, give us the top K probabilities, top, and specify the K here. So here if I print ⁓ top K, it comes with two ⁓ values. ⁓ Yes. So ⁓ values.

Sneha Mehra (01:20:33)  
dot to list.

Sneha Mehra (01:20:42)  
So let's see what GPT do thinks that ⁓ the top five probabilities would be.

Okay. All right. It's just outputting the top probabilities. The first probability is 77%. So GPT2 seems confident about the next port. And the the rest is very ⁓ low. So now all we need to do is just we need to print their corresponding ⁓ tokens. So I'm gonna print this here. Now for the top five, we go over.

All the indices and the values. And then for each for the top value, we pass it to the tokenizer because they're just IDs. We pass them to the tokenizer decode to get their corresponding token. Now, if I print this, we would see the highest probability word is is with 77%. And that was expected because the input was hello, my name. If we change it to like hello, then the network, the GPT2, would be less confident about the next word. ⁓

Somehow it crashed.

Sneha Mehra (01:21:53)  
So let me quickly rerun.

Sneha Mehra (01:22:13)  
All right, so I think we can pass this. It's ⁓ it's gonna run. ⁓ and then you we saw that for hello, my name, the next token is is. Now the next thing is we go to the generation. For generation, I mean, from this point forward, it's going to be much, much faster. ⁓ because we just learned the high-level ideas, and I'm just gonna copy the code and go over them. ⁓ so we learned about greedy decoding. So we can easily load models, let's say GPT2, their tokenizer and model.

And then Hugging Face also provides a way to generate. So we don't have to really implement all those things that we implemented here, getting the logits and getting the last one, uploading softmax, top K, and anything. We can just ask Hugging Face and pass the parameters, and then it would generate for us. So ⁓ here, let me ⁓ copy paste this.

Sneha Mehra (01:23:08)  
All right, this is going to be super simple. We are not doing anything fancy. So we have just a list of models. It's just one example here. We are loading them, tokenizer and model. We saw earlier. Here we make sure that this particular model has already a pad token. If not, we just set it. It's just an ⁓ unnecessary detail. We can we don't have to worry about this. And then here we build our ⁓ dictionary of tokenizers and models. In this case, we are just having one example.

So and then we generate. So ⁓ for generate, again, ⁓ it's a function. It comes with a model key prompt and a strategy. And we don't have to implement this. We just ⁓ tune the parameters that the hugging face expects. So if we search hugging face.

LLM generate. You would see text generation and how it works and what are the expected inputs. ⁓ so here it's basically saying that ⁓ these are model inputs and you can you can take a look at the different inputs that hugging face expects and also beam search if we want this to be sampled sampled or greedy like deterministic and so on. So you can just read this at your own convenience. But here we have this generate and we just update these ⁓

These flags. We get the tokenizers and model. Then we generate these arguments. So we send this ⁓ sequence of tokens. We specify based on what the input strategy is. If it's greedy, we say don't sample, it's false. If it's top K, we say sample true and we provide top K and temperature values. And same for top P. And then we ask it to generate. We just generate, use the generate method provided by the hugging face model since we load it.

And if I run this

Sneha Mehra (01:25:01)  
It says loaded GP two GPT two and then it already has this generate. So if I run this now.

⁓

Sneha Mehra (01:25:16)  
So it would run if you run this on Google Colab. ⁓ I guess for my Mac I have to change it to use my accelerations.

Sneha Mehra (01:25:32)  
So here I run this again, it loads GPT two.

Sneha Mehra (01:25:39)  
I run this again.

Sneha Mehra (01:25:43)  
It's basically generating. So for this sentence, they started to generate. And then here that the continuation is like this. Once upon a time, the word was a place of great beauty and great danger. And so on. So basically just come continuing my sentence. And some more examples. I would share this ⁓ notebook anyway after ⁓ this session. So you can go over the outputs. And same goes for different strategies. So if we I rerun this with ⁓ top K and top B, we would perhaps see ⁓

Better and better outputs. Now finally, the last part is the ⁓ the last part is now we want to switch to better models and perhaps ⁓ like chat completion. So far we see chat GPT GPT two and it's just a continuation. We want to see also the completion model so we can actually ask questions. And the code is not really ⁓ different from what we saw.

I'm gonna copy paste.

Sneha Mehra (01:26:48)  
here. Again, tokenizers model, we set the device, and then we go over them and run it. Now, this was the default. It's gonna run on Colab. It's not gonna run on my MacBook. I'm not gonna try it. I know it's gonna crash. So instead, there are these two models that have previously loaded, ⁓ Quen 3 ⁓ 0.6 billion and Gemma 3, 270 million. And then these would work. And again, you can see a list of ⁓

Hugging phase models, if you just search like hugging phase quen. You would see all different Quen models here in the model card. ⁓ so if you click here, you would see all different models with different number of parameters. We have 8 billion, ⁓ we have 0.6, instruct, 30 billion, like everything. Like instruct are the ones that are for chat for chats. ⁓ but again, I just try to.

⁓ pick these two and they're relatively new. Gemma is Google's Gemma three zero two seventy million. It's extremely good and efficient. And same for Quentin. And this one you can also run it on Colab. So let me run this ⁓ and see what the output might be.

Sneha Mehra (01:28:03)  
So loaded GPT two. I hope it it doesn't crash because there are three models now. All right, so one model, two model was loaded, third one was also loaded. So now if I run this on, let's just do one on Quin.

It would ⁓ use the Queen model and sample from it. I can also change it to Gemma to see what is the difference and what Gemma would produce.

It's gonna take longer because some of these models are now bigger. But all right, so it created. Once upon a time, there were ⁓ 3,000 people in a town. So later, if you compare this, you would see that the quality is typically better and higher compared to GPT-2. ⁓ okay, we're almost done. ⁓ And let's see, I think there is ⁓ one more ⁓ step.

Sneha Mehra (01:28:59)  
All right, so it just ⁓ ran Quen. Again, I'm using Quen ⁓ three. It's not a chat model. You can use this chat model in the Google call app, or if you have a more powerful ⁓ laptop or GPUs, it's not just gonna run on my laptop. ⁓ and if if we change this to Gemma now, it would ⁓ use Gemma to ⁓ sample from these ⁓ using these three examples.

I'm not gonna run it. So that's mostly it. The last part is just an LLM playground. I that's not the focus of this course or this project. It's just ⁓ it's just a UI implementation to just get a good feeling of how it looks. I have something to copy paste here. It is the like simplest UI possible. ⁓ it's but let me run it and

Also, let me make sure that we are using the GPT2 and Gemma, which is already loaded. All right. This is my this is my UI. ⁓ And ⁓ it's again, I'm going to share this code so you would have access to the code, but it's the like the most simple UI you can think of. And then we can ⁓ choose GPT 2 and Gemma 3 ⁓ and the strategy here. And you can also set the temperature. So let's try GPT 2 and tell ⁓ say, tell me

Tell me a joke. And I can press this generate.

Sneha Mehra (01:30:34)  
So again, this is a continuation, like it's a completion. It's not a chat model. So it's just continuing. Tell me a joke about it. I'm not sure if you are aware of it, but I'm not sure if you are aware of it. This is one of the typical issues of GPT2, which we covered in the guided learning. It just gets us stuck in a loop. ⁓ if we change the strategy, ⁓ I don't know if it's gonna make any difference in this particular case, but let's see. All right, it's better. At least it's not in a loop. It says, tell me a joke. I thought you'd have to get some help.

So it's basically re repeating or generating what it has been seen in the its training data from internet. ⁓ now if I change this to Gemma, it's still the same because it's not a chat model, but it might the the output might sound more more meaningful.

Sneha Mehra (01:31:33)  
I I I really don't know if it's working on okay, it's working under the hood. Tell me a joke, but don't laugh. I could hear the voice saying, Don't, don't, you are the most best person I've ever met, and so on. So feel free to play around with it with the temperature to see how you when you increase the decrease the temperature, the results are more ⁓ expected and deterministic. And when you increase the temperature, you would see more creativity. And sometimes if it's if it's too high, the results would not just simply make sense.

So ⁓ all right. So we ⁓ we covered everything. ⁓ I'll just keep the code ⁓ and we'll push it. ⁓ I next we can we'll gonna take some questions from raise hands. But before that, I wanna quickly share that we have ⁓ released the project and also the ⁓ material for week two. Week two is going to focus on the rack. So let me

Quickly show this and then we'll take questions.

⁓ here. ⁓

So week two is already available. I the video has now captions and also it has ⁓ the chapters for for convenience. Also, project two is I try to improve project two based on all the feedbacks that we've received for project one. So ⁓ it's also focused on RAC. It's less high level, it's less low level, it's more high level. And also it has ⁓ a lot of ⁓ links for you to go over and take a look.

Sneha Mehra (01:33:10)  
and more comment for libraries. So you understand the purpose of each library. And also it comes with like some hints, like two lines of code and you know, like so you get a get an idea of how much ⁓ you know ⁓ lines of code it's expected. It also comes with ⁓ yeah a bunch of other things. So yeah like expected steps so you know how to implement it. What are the steps? You can just follow this.

So if you have any feedback for project two, please also let us know. We can improve the project three. So yeah, let's go to the questions now. ⁓ I'll start with ⁓ Peter.

Hey, ⁓ first of all, I I really appreciate the depth in ⁓ of what you've got over so far. But I'm wondering, ⁓

I I have a question about the inputs, the input size and the output size. In your in your presentation, you had this diagram. If you can pull it up, I might be able to point you to it. But it's like you have four input neurons and eight output neurons. And I just want to understand what your point is that you were making when you were saying that you have these four inputs and these four outputs. Could it be four inputs and two outputs or three outputs?

Yes, basically you can decide any number of inputs and outputs and it depends on your design and you know the neural network that you are implementing. But need this needs to be fixed. This needs to be decided ahead of time, right? Because it cannot be adjusted, because the weight size needs to be determined and fixed, then we are building the neural network. So all we need to ⁓ And this is something that's very complicated and has to do with the structure of the neural network itself. It's not something that we could just decide arbitrarily.

Sneha Mehra (01:34:57)  
of course. It's it's something, it's not very comp it's not super complicated. It depends on it depends on a couple of things. First is what tokenizer you are using. ⁓ so like IDs. So I mean IDs is not important, but once you get the IDs and you want to go from your IDs, let me bring it here. You want to go from IDs to this embedding layer. So that's one thing that is important because whatever hidden dimension you decide your ⁓ LLM to work with.

This embedding layer would just convert all your ⁓ IDs into a vector of that length. And then every single layer would just operate at a hidden dimension size. So now, based on that, you just need to walk through over all your layers when you are implementing your neural network ⁓ and ⁓ figure out the sizes. So that's one of the things that needs to be figured out always. So if you're also interested, I suggest you read the GPT-2 ⁓ code.

From OpenAI, so if you go to the source and model, you would see like how these model and encoder decoder are built. So you would see like these are some of the parameters, like N embed 768, number of heads, num layer. So these are fixed and determined beforehand. And then ⁓ you would see these MLP layers. when they are creating this, ⁓ they just specify all those ⁓ shapes. So I'm not gonna go over it, just take a look and

you would see like how they specify those ⁓ dimensions for each layer.

Sneha Mehra (01:36:30)  
I hope it was helpful, but yeah, just post if if you have any questions and I'll make sure to answer. So Tanvir, can you please go next?

Hey, thank you, ⁓ Ali. ⁓ I love the content and the way the depth you are going through. I have a quick question. ⁓ in in I think in section 1.2 in data cleaning ⁓ step, I think you have mentioned there where it says 15 trillion tokens. But after like we create, I think you're showing some 200k tokens or something. So I'm just trying to, I got confused with that.

Okay, okay. That's a that's a great question. ⁓ 15 trillion tokens, right? So so let's get first first get to the tokenizer part. 15k or 100k. Those are all the unique tokens that we ⁓ create ⁓ and inject in the in the tokenizer. So basically we have some data, and this data can have a lot of different words, a lot of duplicates, like it could be anything, right? So we train a tokenizer on that tokens, and then we end up with some vocabulary.

This vocabulary would have ⁓ a size of 50,000 or 100,000, right? But when you look at the training data, the training data can be a lot more just because there are a lot of duplicates. For example, whatever you see, I mean, a lot of the content has ⁓ just ⁓ same and same tokens are being repeatedly ⁓ appeared in the training data. So that's why it's significantly higher. So if you look at the

So here, for instance, we have 15 trillion tokens. That means that if we get, if we use the fine web and we crawl the web and we just use the find web to clean it and we use our tokenizer to tokenize it, we would end up with a very long sequence IDs of 15 trillion tokens. Now, a lot of these tokens are ⁓ repeated. For example, we have the same words over and over again on internet data. Now, each token is going to be in the range of zero to vocabulary size.

Sneha Mehra (01:38:37)  
⁓ so that's basically it.

Sneha Mehra (01:38:43)  
Santosh, would you go next, please? ⁓ hi Ali. Again, thank you so much for ⁓ the depth content ⁓ you know, ⁓ session today. ⁓ So my question is so adding the let's say ⁓ like a text and the image or the audio, the video modalities ⁓ into the mix, ⁓ especially for the tokenization, which models would work the best? So based on this, I think currently even today we focused on the text, right? So

like to play around with ⁓ the image, the audio video ⁓ modalities as well. Sure. I think you are referring to vision language models, right? So so that they are trained on both text tokens as well as visual tokens.

So there are there are a lot of models that are that are good. ⁓ like Queen3 VL is something came out very, very recently. And it seems it's it's awesome. It's open source, it's open-based, ⁓ and you can just use it here. ⁓ yeah, there is a place that says you can how to use it. So Quentin3 VL is awesome, but there are a bunch of other vision language models. I think even the ⁓ LMRNR, they should have a leaderboard for vision language models.

You are you are able to see the screen, right? ⁓ DMA 1.5 or 2.5. You recommend. ⁓ let's see. So if you go to ⁓ vision, so here you would see the best modifs. ⁓ so some are open source, some are closed. So yeah. Okay. Yep, sounds good. Thank you, Ali. Sure.

Rahul, please go next. Hey Ali. ⁓ very nice content. Thank you for this. I have a couple of questions. ⁓ one is you in in the last time you mentioned some of these companies, they do train ⁓ or post train the models ⁓ with a fine tuning. So can you just give us like what does that look like if somewhere were to work on that field?

Sneha Mehra (01:40:46)  
The second one is can you elaborate on the ⁓ Rahul, sorry to interrupt. So regarding your first just to make sure I understand. So your first question is how companies typically do post-training? What would be the like work look like if you were to do the post-training? So I I you s you have examples there, but ⁓ I don't think we have had any ⁓ you know exercises ⁓ along ⁓ on those side, but is it like what would it what's ⁓ typical work look like ⁓ with the fine tuning ⁓ space? ⁓

Sure, I can I can share some ⁓ some resources for fine-tuning, but it's actually very simple. Let me very quickly write a pseudo code. So we have the tokenizer, right? ⁓ And we have the internet data for training. This is for training, internet data. And we have our model also. ⁓ We use our torch NN, build build our model or load it. Right now we want to train it. Basically, the way that we train is first we pass the internet data to the tokenizer. So

Tokenhezer. ⁓ sorry, I mean it's just a pseudo code, but just to get an idea. Internet data go to the tokenizer, we get huge IDs. These IDs are IDs equal fifteen trillion.

Sneha Mehra (01:42:00)  
Now for batches of the smaller IDs for batch IDs, right? In IDs. So now we just go over them in a smaller piece, like every ⁓ like 1000 IDs. So it's just a sequence of 1000 numbers.

We pass it to the model, so we do ⁓ equal

Sneha Mehra (01:42:25)  
Equal ⁓ model which we have. We pass this batch of IDs. So now we have the out. We also know the ground truths. We get the ground truths. We know what is the right next token. We have this out and then we compute the loss. ⁓ we basically pass this IDs and ground truths to a loss function. ⁓ Out, out and loss ⁓ and ⁓ ground truths to the loss function. And then we use torch. ⁓

⁓ backward and then optimize our data step.

Sneha Mehra (01:43:02)  
I'm going to share more stuff, but just wanted to give you the sense. That's it. I mean, it's it's nothing more. It's just ⁓ very simple. ⁓ most of the problems and complexities are you know how to deal with these things at a scale, like making sure that there is no bugs, how to do it at district distributed training, ⁓ some minor things here, you know, different precisions and so on. But at the very high level, that's it. You ⁓ there is nothing more. But that that's for the training data, right? I mean, do you would you say the same thing for the post-training too? This so

Both pre-training and and SFT are exactly the same. Just the training data is different. Just this is different. Okay. Okay. I think, okay. So okay. ⁓ and then I just have one follow follow-up question. So when you talk about the batch IDs, like sending some certain data in terms of the batch, can you elaborate what does that mean? ⁓ I think in in your one of your exercises, we had a batch ID set to one. So what I mean, would you like ⁓ logically group the certain data set together to

process it in a batch or is it just based on the length of the ⁓ yeah I don't know the the maximum context that you want to provide for the models to learn. Mm-hmm. So basically instead of sending examples one by one, we typically batch them for ⁓ efficiency and also for faster like convergence in learning. So let's say on internet day we have a bunch of things. Today's podcast is about and then another one like

⁓ you know, let me share the the the cooking or something like that. So we just batch them, like in this example, three examples, we patch them, we send all three to the model. So the model simultaneously learns from all these three.

Okay. This is basically the batching. It just helps you to converge faster.

Sneha Mehra (01:44:51)  
Okay, thanks. Of course. ⁓ Baiba, please go next.

Hey, ⁓ thank you for this wonderful class. So my go straight to the question. So my question is how this embeddings ⁓ is related to attention mechanism. You mean the embedding size? ⁓ the embedding that we generate. Is that the same embedding that we use for rag after tokenization? ⁓ you are referring to this embedding layer, right? ⁓

So this is going this is basically the same di dimension that the attention is going to operate at.

So it would just process all of these and make sure some communication happens between all these vectors and just transform each of them. Got it. So the tokenization, right? After tokenization, it has to be converted into ⁓ embeddings or just tokenization and directly. after the tokenization, it's a sequence of IDs. It's just a sequence of IDs. It's for example, here we have five, eleven, twenty-five. This goes to the embedding layer. Now embedding layer just looks at its own lookup table to

go from these IDs to ⁓ a ⁓ its corresponding vector and just replace the similarity search. Is that a it's not gonna be similar to search, no. It's just a lookup table. It's basically a lookup table that the neural network just keeps track of and updates all the weights during training. So row one has a vector corresponding to token ⁓ bit ID zero. Row sorry, row zero. Row one has a vector corresponding to a ⁓

Sneha Mehra (01:46:33)  
Token with ID one and so on. And all these weights are these are also weights of or parameters of this model. They would get updated as we train the model. So these vectors become meaningful. For example, if we pass hi or if we if we pass hey or hello, all of them would probably end up with vectors that are close by in the embedding space.

I see. ⁓ and then attention mechanism, right? Can you briefly cover that? How does this attention is ⁓ materialized? How is that captured? So ⁓ basically, attention, it tries to transform each vector by looking at all the other vectors and then see ⁓ like figure out the similarity ⁓ and how relevant they are. For example, then it's trying to transform this middle vector, it looks at this vector, it also looks at this vector.

And see and it somehow it ⁓ figures out how relevant each of these two vectors are to this, and then based on that, it would ⁓ update these vectors so that it also borrows some information from these two vectors.

I can also share some resources for attention to better understand that. ⁓

Sneha Mehra (01:47:53)  
No well?

Sneha Mehra (01:47:58)  
⁓ hi. Thank you. Thank you so much for the session. ⁓ question on the ⁓ how do we actually store the corpus ⁓ when you train a model? For example, ⁓ when say OpenAI trains their LLM, ⁓ how do they store the corpus? Is that in files or because it's around 44 TV of data, right? For the initial set of LMs? Yes, yes. So it's ⁓ they're usually trying a lot of advanced things to ⁓

to store this. I mean, the simplest thing they they we can do generally is we shard the data into a smaller piece so they become more manageable. And then each of those can be stored with some efficient format, ⁓ like parquet or something. And then those can be stored on on S3 or some other storages, right? So now during training, then we companies typically implement an efficient data pipeline. So ⁓ it so basically the data pipeline is responsible for fetch data.

Read it, prepare the next batches, and make sure it's ready for the next iteration of training. So the data pipeline is just a separate effort, and usually companies have like teams of 10, 20, or even more people working on those. And make sure that at any given time, data pipeline can just go ⁓ read the files, ⁓ fetch their next ⁓ set of batches, and make sure that's ready. So that as soon as the GPU screes up for the from the previous iteration and wants new data to learn from.

The data problem we just pass it to them.

Okay. Thank you. ⁓

Sneha Mehra (01:49:35)  
Ranganasad, please. Yeah, yeah, hey. Yeah. A question on the multi-head concept. can you explain ⁓ how does it work? Because I read that it looks at the different aspects of input tokens. ⁓ And how is it trained? ⁓ also is it like a trained separately? How do you define how do you decide on how many heads you need in the model? Okay, so the training is the same. I mean, we don't have to do anything specific for training. It's just this loop, right?

It just takes care of it because part of the model is also all those attention layers. And those parameters get just updated. So that's that's about the training. We don't need to do anything special when we have attention. About the multi-head aspect of it, even that we don't also need to worry about it because when we create the layer for attention, which is let's say inside this model, it would have one additional dimension. So instead of one ⁓ attention, it performs multiple attentions, and each attention is

Basically, the hope is that each attention head would focus on different things, which we don't know. And we won't we won't know, you know, at even at the after training. We don't know really, like each head is is ⁓ focusing on what. It's just ⁓ but they just figure out. I mean, just the the the attention just figures out like each head should focus on certain things. So you can just think of it as a ⁓ multiple, like a a set of different attentions, each focusing on different aspects than they are.

Going through this communication between these vectors. So is the number of heads decided by automatically decided, or is it like a trial and error set by the whoever is designing this ⁓ LM algorithm? It is a hyperparam. So I think it is a hyperparameter. So let me I ⁓ see. Okay.

Sneha Mehra (01:51:31)  
So it's, you know, it starts from like twelve or something and it can go all the way to like to hundred, I guess. Let's let's quickly see if Chat GPT also ⁓ confirms this. So GPT three, for instance, had ninety-six heads. GPT three as small had twelve. So yeah, it's it's it's within that range.

Okay. Hey, I have one more quick question is number of blocks. ⁓ you mentioned that there are number of transfer blocks. ⁓ again, the question is how do you determine wha how many number of blocks you have to stack? And even in the training, do you train with all these blocks in sequence and the back propagation goes all the way back to the embeddings, ⁓ initial embeddings?

So regarding your first question, it it is a hyperparameter. So a lot of times when you want to make your models bigger and make it more capable and it increase its capacity, it's gonna cost more. But if you want to do that, you typically increase all your hyperparameters, right? Including ⁓ the hidden dimensions, number of heads, and also number of blocks. So if you look at this thing, this here, GPT 3, a small, medium, large, Excel, and so on, they keep increasing all of these hyperparameters, right? Including the number of layers. And number of layers is what you are referring to.

So it they start from 12 to train their smallest model and then 24, 24\. And each of these models are trained separately. So basically they just set the number of layers to 12 when they are creating the model and just train it. So the training it happens automatically. Again, just the optimizer and backward would just take care of calculating all the gradients and updating all the parameters of ⁓ each transformer block. Okay. Okay, thanks. Thanks, Ali. Of course.

Sneha Mehra (01:53:19)  
Andrew, please, can you go next? Thanks, Holly. I appreciate the time. There's also a nice list of questions building up in chat. I ⁓ I had a question that's a little bit off

⁓ sorry, I lost I guess I lost your voice.

Sneha Mehra (01:53:43)  
I am not still hearing.

So maybe maybe ⁓ in the meantime I'll just take one more question and then come back.

Sneha Mehra (01:53:53)  
Keepan, Keepan Mita. Hi. Hey, thank you so much for the detailed course. I had a general question which I kind of face at work. ⁓ so most of the time when we build some models with whatever data company data we have, we never really work on ⁓ optimizing the tokenizer. It's always about building the model. But what happens in most of the cases, like you pointed out, was

We don't end up with a lot of tokens because they are very specific to my work or my domain. So ⁓ how, I mean, there are various ways to update the tokenizer. We can just add tokens or train the new whole tokenizer, but how how does that affect or when we should choose one over the other? And how does that affect the existing tokenization that already is there? How can we maintain that and add more tokens?

And ⁓ I just have one small follow-up and how does that affect the semantic drift? Like ⁓ word can mean something today in the existing tokenizer and then I add something, it may mean something in my company's domain. So how to how to do that? Sure. that's that's a great question. ⁓ to be honest, tokenizers I think is the probably the most boring thing, right? So a lot of a lot of times a lot of companies they don't really spend much time on figuring out the tokenizer or training their tokenizer unless they are various unless they are very big.

Or they're very advanced and they want to ⁓ do something very new or or drastically different in the tokenizer. So a lot of times what is typical is that ⁓ we just start from, let's say, the ⁓ most recent OpenAI tokenizer, which has like 200K tokens. And that's typically sufficient. I mean, for most use cases, it's it's sufficient. It's it has like different languages, it has all these different ⁓ vocabularies. And

It's not also really valuable thing. I mean, so that's why like open air keeps releasing them, right? So it's not something that they want to ⁓ keep it. I mean, they train something, it works great, they just release it, and then everybody can just start from them. Now, ⁓ one thing to note is if you have a very, very, very specific and custom question, which which let's say the language is you are inventing a language. It's not even it's not even existing on internet.

Sneha Mehra (01:56:05)  
Then you need to update your tokenizer. Basically, you need to come up with these new special tokens that you have for your task. And then you continue, you basically add those to your tokenizer. And then when you want to train your model, you would end up with some new numbers, right? When you pass it to the tokenizer, because now you are adding new ⁓ tokens to your vocabulary. So that's possible and that's also a very innovative way to do it, but it

really requires a specific use cases. And actually Google recently did this. ⁓ Not recently, it's it's but it's generally more recent in in the recommendation system. So if you go to the ⁓ semantic ID. So if you go to this paper, they are trying to continue ⁓ training an LLM, but then they are also injecting a new vocabulary into it. And those vocabulary are coming from the videos. So I suggest you take a look at this ⁓ this paper.

It's it's it's going exactly through the same thing that you just mentioned. They come up with the semantic IDs and then they inject those into the tokenizer and then they continue training the LLM. And then they show that after this process they can get better and better ⁓ video recommendations.

Sneha Mehra (01:57:21)  
So next, ⁓ Andrew, do you do you wanna go next if if your mic is fixed?

sure. Thank you. I appreciate it. ⁓ I was gonna ask if it's possible to get a different interface on the videos, either embedded on the same platform or embedded from something like YouTube or Vimeo. On mobile devices it's kinda hard to skip around with the tracking. ⁓ I also wanted to mention for anyone who's not in ⁓ watching chat, which is totally reasonable, ⁓ paying attention to Ali's why we're here, ⁓ to read the zoom chat afterwards. It's kind of like an FAQ that's building up.

Thanks. Yep. Thank you so much, Andrew, for the suggestion.

Sneha Mehra (01:58:07)  
Nee shit.

Sneha Mehra (01:58:11)  
⁓ yeah, Hay Ali. ⁓ my first question is like very fundamental, is like ⁓ with all these explanations, ⁓ I want to understand ⁓ how the hallucination happens and what are the ⁓ like techniques to control it. Mm-hmm. So hallucination is a very difficult topic. I mean, over time, like companies like Open Air, they try to reduce it, but still there is not a very systematic way to ⁓

⁓ fix it. The problem is there is not also a systematic way to understand why it's happening. Very recently, OpenAI ⁓ shared a paper about hallucination. ⁓ why language models hallucinate. I ⁓ strongly encourage you to take a look if you are interested in this topic. But I can give you a summary, like a two-line summary. Basically what they're what they found out that the hallucination is because we are not

Basically, it's because the way that we are evaluating the LLMs. We are ⁓ penalizing the LLM when they produce something wrong. And then we are not gonna introduce a penalty. So whenever an LLM is not confident about something, it just hallucinates, which is not good. But it hallucinates because it over during the training process, the during the reinforcement learning process, the LLM learned that if it's not confident, it's just better to generate something. And

In many cases, it hallucinates. ⁓ it's like the exam tests. You know, sometimes if if there is no negative point, if you are if you are incorrect, you would just even if you don't know an answer to some exams, you would just make some selections. ⁓ but if you know that there could be some penalty, if you're wrong, then you won't. Now, the current LLMs are based on the former. This ⁓ study and their paper is saying that the evaluation mechanism has to be different, so that then an LLM is not sure about.

something, ⁓ it would it should not just guess. It there would be a penalty if it guesses and it's incorrect. So yeah, if you are interested, just take a look. I mean, I really enjoyed when I read this. So yeah. Yeah. My my concern to this Ali was like ⁓ you know I'll listen to the hallucination problem is like reduced or sort. So I mean I I don't think that the system should be able to trust the LLMs completely. So one thing is like they should, I mean

Sneha Mehra (02:00:36)  
I mean, i is there any progress being made that you know LLMs while responding do know that okay, I do not actually know the response rather than giving a wrong response. ⁓ I should rather say that I I don't have any knowledge about it or just a stop. ⁓ is any progress being made? Yeah, so there there are two levels that this is trying to be solved, right? So one is at the LLM level, not at a system level, which is this paper. Just they're trying we are trying to there are a lot of work that they're trying to just fix the LLM so it doesn't hallucinate. But

At the system level, yes, there has been a force and they've been quite successful. For example, Prag, which is the the next topic, right, for this week. They ⁓ they many companies try to just search a web and ground their outputs to reliable facts. So that's one way that they make sure that the models are not hallucinated. The other, which is ⁓ in our future weeks, is about reasoning and thinking models, they also do the same. So they increase the inference time ⁓ compute. So the model thinks and then it just keeps trying.

Thinking different things and make sure and validate them. And if they're wrong, it goes back to some other directions and think about them and search the web and make sure they're correct. So that's more at the system level. Again, both this week and in the reasoning LLM week, ⁓ we would see some methods that it would make sure that the LLMs like hallucinate less. Yeah, thanks, Ali. Just like ⁓ the next question. ⁓ that's I ⁓ so that that is basically I I'm curious about.

You know, like processing this trillions ⁓ or trillion bytes of theta to train these models. Like ⁓ so yeah, I mean the GPU is like very centered to that training beyond and and know the importance of GPU, that the role that the GPU plays. But is there any like special like frameworks or ⁓ like you know ⁓ being on libraries being used like like we used to have this Hadoop and Assad in the past. And I do know that this Python language is like beyond a point. It's it's

it's basically it's not that performance. So ⁓ language wise, there are any choices they have, like, you know, they they go back to like ⁓ for this performance, they they they go back to s to implement something in C or Rust or something like that. If you if you have any like idea on that. So just to make sure I understand your question. So your question is what are some frameworks and tools that allow us in the language model space that allows us to efficiently train at a scale and use distributed training and all those things, right?

Sneha Mehra (02:03:00)  
Yeah, so the the underlying ⁓ the tech stack ⁓ in terms of ⁓ tech stack yeah. Sure. ⁓ I mean if if it's okay, I can share this because it's more about the links and some resources. I have some good resources in mind, so just ⁓ just ping me in in DM or in the chat channel. So I'm gonna share some resources for that. Yeah, that will helpfully. Thank you. Sure.

⁓ could you?

Yes, thank you very much for your presentation. ⁓ I I was wondering, ⁓ when you mentioned this ⁓ C4 so the big corpus from the internet, ⁓ indeed there are ⁓ the vast majority of the infos ⁓ are embedded within text, but what about ⁓ everything related to videos or images, like how those ⁓ different ⁓ type of file are then incorporated within the process of tokenization?

Okay, that's that's an awesome question. Both for images and videos, we'll go if we're gonna cover them in the last week in detail, including the tokenization. But I can briefly just share what's what's gonna happen. So let's say the input is image or video instead of raw text. Still, it's the same process. We just need to build a different tokenization with a different process to convert the image pixels or video pixels into a sequence of IDs. If you're interested to learn about this ahead of time, you can

Just search like image tokenizer ⁓ or there are a bunch of things. This is the first ⁓ like VIT paper, but ⁓ I think.

Sneha Mehra (02:04:42)  
You can also read some of this, but basically the the key idea of image tokenizers is to go over the pixels and somehow convert them into IDs. And once it's IDs, the rest is the same. The rest is just training a language model. Nothing is different. You just need to ⁓ replace this tokenizer so it works with images and IDs. Also, this VQ, ⁓ VQ, VQ VAE paper.

I suggest you take a look. Again, we are gonna go over this in in the last week, but in case you are interested to learn more, ⁓ you can go over this paper. ⁓ They train a quantization process. And then here basically they are converting this image into a sequence of IDs, of course, in two D, but yeah.

All right. ⁓ thank you very much.

Sure. ⁓

Right. Yeah, ⁓ I just want to come back to that ⁓ batch question. So is it fair to assume that one batch usually would or should consist of ⁓ words that are related to each other? Meaning, in the game, based on my understanding, based on one batch, one batch would form a context for what we are training for, and based on how the

Sneha Mehra (02:06:02)  
The sentence is formed, it will form the ⁓ you know, based on that context, the next token will be predicted. So based on what's the last ⁓ token is in that context, based on the positional encoding, it will try to predict the next token. So is it fair to assume while they are training ⁓ while the training is being done that the batch size will have a relative ⁓

the the you know the context would be for the relative content. So like in this example, like hi, how are you? If you put everything in the same batch, then the then model will not be able to accurately predict what should be the the yes next logical tool. Yes, yes. So that's why in the first place we are just separating them in a in a in a separate dimension, right? We are not just gonna pass a very, very huge ⁓ sentence.

So basically the LLM would process each of these individually, and then the attention happens across each sentence. Right. So this yeah. Yeah, sorry. I was asking ⁓ is the high how are you? Would that be part of the one batch? And then today podcast is about that would be a second batch in the training. So this is one batch, this entire list, right? The LLM ⁓ would go over ⁓ this example.

Converts the tokens into vectors and the vectors are being transformed. And then ⁓ based on the output loss, it would come up with some gradients to update the parameters. The same thing would happen simultaneously for this sentence. And the same thing would happen simultaneously for this sentence. And then after all these three sentences, then the gradients would be kind of averaged or accumulated, and then ⁓ would update all the parameters of the model based on this patch.

But none of these examples are supposed to be ⁓ relevant or semantically related to each other. Okay. And the way it is being split within that batch is based on where the context ⁓ where that the context ends, starts and ends, right? Is that that based on the codes? ⁓ like hi how are you will be processed once. ⁓ today's broadcast is processed in parallel this the next, right? So

Sneha Mehra (02:08:22)  
The next token for higher how are you would be based on the the relative meaning of what the high how are you is, right? And what's the next best token is exactly. Exactly, yes.

Okay. I'm just trying to get this in my head, but okay. I think Yeah. Also feel free to post questions if you have follow up questions I can answer them in the chat. Not not not not this chat, ⁓ the platform chat.

⁓ višnu.

Sneha Mehra (02:08:51)  
hi Ali. ⁓ my question is ⁓ so you you mentioned so many decoding strategies, right? Like Tosk, Top P and all. So ⁓ how do we know like which strategy to use during runtime and like how do you decide like what are the right hyperparameters to use? Like are there any techniques that will ⁓ determine the success of the generation? Like what are the success metrics that we generally see to see like this this thing?

Whatever the Gen A system that you deployed is successful or not. ⁓ Yeah. Okay. That that's a great question, also. there are some things that we know are not good. For example, we know grid is not good. So ⁓ you you never want to use it in production. Okay. We also know that top toppy is perhaps the best tool we have so far for generation. There are also some other decodings that are ⁓ introduced more recently. So one is

Speculative decoding by by perplexity, I think it was it was introduced. But I mean, if you are interested, you can go over this. But but basically, top P is the best. It's just about these parameters. And there is no ⁓ correct number. I mean, there is no one size fits all. It really depends on the task that ⁓ and also your LLM. So, what typically happens in company is that we we play with this. We play with this top P and we play with this temperature. ⁓ And if ⁓

Especially temperature. I mean, if you want more creativity, you can increase the temperature. If you want ⁓ more certainty or you want the output to be more deterministic, you decrease the temperature. For example, let's say you want to use an LLM to ⁓ create some data set or create some captions for images. You want to have a very low temperature because you don't want the model to be creative. You want basically the model to output the most deterministic expected caption for the given image. So this is very empirical.

And also OpenAI APIs and all so they also let you choose when you send your request to choose your temperature or you know some of those hyperparameters. ⁓ I think there was a table in the lecture which just sharing some. Let's see if I can find it quickly here.

Sneha Mehra (02:11:11)  
So these are just some empirical values. But again, you have to figure out what works best for your use case. But typically in code generation, we don't want creativity or we want more deterministic behavior. So temperature is low. Whereas in creative writing, it can be zero point seven or even more. ⁓ so yes. Got it. So on what things they will basically benchmark it, like ⁓

Are are they going to do some grid search or random search to determine these values on a specific validation data set to come up with those numbers? Yeah, it's just an an inference inference time thing, right? So it depends on if you have a systematic evaluation in place. For example, in image captioning, if you have ground truths, you can rely on some ⁓ LLM as a systems to come up with the quality of your output. Then you can try different values, run evaluation, get the numbers and see which which pair of value or which temperature basically is giving you the best numbers.

Otherwise, other other than that, if you don't have an evaluation, like a systematic evaluation mechanism, you can just it's typical just to eyeball and see, you know, which of these are producing something better and then and then just stick to those. ⁓ Thanks, Ed. Yeah, no worries. ⁓

Anime. Hey, thank you, Ali. And thanks for all your patience in ⁓ answering this question. Quick question. I think when you were responding to someone, ⁓ I maybe I misheard that we can train that tokenizer. I thought ⁓ like tokenizer is basically the vocabulary, right? Or can we also train the tokenizer?

It's it's they're just it's just a terminology. I mean ⁓ in many cases online you may see that they are referring to it as training, but basically this whole training, what it means is that they go over the corpus and then they create the vocabulary. That's it, what what they mean by training. Okay, I have just one follow-up question. Uh-huh. Go ahead. ⁓ regarding, I mean, when we

Sneha Mehra (02:13:16)  
Like you you you had mentioned that hey, like we need to pass the neural ⁓ neuron layer, it can take only numbers, not the text. So that's why we have the tokenizer which converts it to numbers. But what's the purpose of I'm not able to understand? ⁓ what's the purpose of embedding layer? What does it do? like we already have the numbers, we can fit into the neuron layer. Then what's the purpose of embedding layer and then it converts to vector?

Mm-hmm. Mm-hmm. Yep. Excellent question. ⁓ let me ⁓ let okay. Let me now zoom out.

Sneha Mehra (02:13:59)  
Well, here it's just easier if I go over these visuals.

Okay. So ⁓ the reason ⁓ that we have this embedding layer is because if we don't have it and we just use these numbers, these numbers are not me are not meaningful. Right? Like how are we getting these numbers? We are just looking at the index of the word in the vocabulary, right? So high can be indexed 10 and then hello can be indexed like 2000, right? ⁓ although these two words are very similar, but their ideas are very, very different.

So ⁓ the neural network would have a very difficult time to figure out all these things and learn about the meanings, like the semantic meanings. The reason that we go from these just numbers, just so they're just random numbers, right? The reason we go from these random numbers to these vectors is that we want to over the the course of the training, we want to update these parameters such that now each token has a vector which is meaningful if you visualize them in the embedding space.

So ⁓ then in that embedding space, high and hello would end up nearby. So this can be very, very important and powerful for this transformer because if you replace this high with how, the embedding would be very similar. And then the output is going to be likely very similar. But if you just ⁓ pass these numbers, if you replace high with how, suddenly instead of 10, then transformer would see 2000 and they are no way related to each other. So that's the main reason that we use this embedding layer.

Sneha Mehra (02:15:40)  
Shaw Jack?

⁓ yeah, thanks Ali. ⁓ I have ⁓ maybe an extension to this question. Like when it comes to self attention, ⁓ one piece of puzzle which is not clear to me is how the weights for the ⁓ Q, K and V metrics are derived.

⁓ they're just being updated by ⁓ the ⁓ like through the back propagation mechanism, right? So ⁓ basically initially there's just some random weights. You pass some inputs based on the inputs, the neural network ⁓ generates some probabilities. Then from the probability we compare it with the ground truths, we come up with some loss function, right? Some some loss function here. And then is it while training the ⁓ the model ⁓ the q and

⁓ like the V metric specifically, right? Metri says those get updated. ⁓ because they basically tell the meaning, right? Like how the the relationship between like, you know, the words basically, right? And ⁓ from different aspects. So

I mean the embeddings are derived at a time of training the model, but in order to derive those matrices, like you know, we ⁓ the embeddings are multiplied by the weights for QKV, right? So ⁓ so not the weights are also derived at at that time. I'm ⁓ a little confused, like you know, how they they will add the different aspects at the time of training. ⁓ so that's that's the whole thing, right? So it's not really the we cannot really interpret the process. What is happening's gonna happen is that all these layers.

Sneha Mehra (02:17:19)  
Like embedding layer has some parameters. It's just stored in a weight matrix. Attention has some parameters, KQ value, stored in some matrices, right? Fit forward, the same. So initially, when we start this model, when we ⁓ build this model, there are just a bunch of parameters and they're random numbers. Now, when we start training based on the loss function, we backpropagate, meaning that we calculate the gradients for all the parameters.

And then for each parameter, we try to tune it a little bit such that for the same input, next time the final output becomes more accurate. So during this process, all the parameters, including ⁓ key values ⁓ and you know, every parameter in attention and feed forward would get updated.

So was that was that helpful? ⁓ yeah, I I mean ⁓ somewhat. ⁓ but is there a ⁓ like you know ⁓ material or something that you could share like just to ⁓ go one level deeper and understand like how the weights are that one one piece is not very clear to me basically. The rest of the process, I think I've got it. ⁓ just send a post ⁓ and then I'll share resources.

Sounds good. Thank you.

Sneha Mehra (02:18:42)  
Cassie?

⁓ hey Ariet, first thank thank you so much for this course. It was it was an eye opener. ⁓ I don't know like I I could have learned so much within ⁓ one week. And your patience and taking up questions and answering that was awesome. ⁓ answer to the question. So I have some ⁓ one one I one thought in my mind. I don't know if I'm asking the right question, but ⁓ normally we when we we did encoding decoding on under one model.

on one base model basically we used ⁓ and then when we try to switch to a different model the tokens change basically if i'm not wrong ⁓ they they may not be the same but they they may change yes when we when we switch to a different model it depends on what tokenizer it was used to train that model but in many cases if the tokenizer is that is not the same

So what that means is that for the exact same input, you would get a different sequence of IDs. Understood. So basically, when we work on, we have to stick to a model where we are working on tomorrow. We change the model, our whole ⁓ logic patterns changes out. So is it possible to use like a multi-models somewhere down the line to see best derive best out of it, or it's too complicated to think about So I mean, even if even if you want to use multi-model, I mean we would see some examples in the last week.

Even if you want to use that, you still need to make sure that you use their same tokenizers. So that's why here, for example, in generation, in generate, right, we make sure to get the models and its corresponding tokenizers at the same time. If you just switch the tokenizer to a different tokenizer, the model would not be able to accurately predict next tokens. Understood.

Sneha Mehra (02:20:36)  
Proset, please go next.

Sneha Mehra (02:20:46)  
⁓ passup.

Sneha Mehra (02:20:51)  
okay. Sorry. ⁓ I wasn't mute. So you explain the ⁓ embedding models, ⁓ that is ⁓ text or sentence embedding models. ⁓ I have a question on follow up that, but ⁓ if you have any documents, I will go through and try to understand. The question is I I understood the it embeddings which ⁓ decides what ⁓ are are are give the context ⁓ of LLM.

what is retrieved before the gener generation, right? So what quality it is how it is a grounding, the grounding truth depends on the quality of ⁓ reference or the text given so the quality of text given LLM depends on how the text is embedded and ⁓ transformed and the ⁓ similarity is measured between the chunks or ⁓

⁓ So sure. I regarding the grounding, right? So that's ⁓ the main focus of rags, right? And VIC2. So we can perhaps cover those more in next project. ⁓ so that's for grounding. But also, I mean, even for grounding, we would use we this embedding layer is still important. I mean, once after training, once we have a good embedding layer, so it it it would guarantee us that we go to almost similar vectors for similar words.

Right. So and that's gonna be useful for grounding. So whenever we have extra external information, we can also use the embedding layer to to convert them, let's say, into vectors and then run some nearest neighbors. I mean, if if you go over ⁓ if you get the chance and go over the guided learning of week two, we talk about this. Sure, sure. So the only concern here is so how to choose this and maybe if we cover fine, how to choose this embedding ⁓ models to for particular ⁓ particular use case because

Data is a domain specific or the region specific. ⁓ So how this is just just to clarify, this is not an embedding model. So we have also a concept of embedding models, right? So those are models that are specifically trained to ⁓ to be used as an embedding ⁓ model. So ⁓ for different use cases. This is an embedding layer, and a lot of times this layer is trained along with the rest of the transformer. So if you if you load any LLM and inspect it.

Sneha Mehra (02:23:16)  
Like all the layers, you would see an embedding layer. So what that means is that basically all of these are being trained at the same time. This is different from embedding models. ⁓ we have some examples of actually embedding models ⁓ in ⁓ project two, I believe. But again, if if I mean you can just search embedding models and then open air and some other companies, they do have their own embedding models. And also there is a leaderboard of that. So it's just an entire model being trained just ⁓ to embed.

Basically just you pass some sentence or anything, it will just return you a a list of numbers.

makes sense. I did confuse it, looks like. Okay, fine. I understand. Okay, great. ⁓ Rushans?

Sneha Mehra (02:24:03)  
I mean, thanks. Thanks for your session. I was with you and ⁓ writing on my ⁓ code ⁓ until 2.1 section. ⁓ And then I felt like it went a little bit ahead ⁓ and I could not ⁓ cope up with the pace of the course. ⁓ I mean I will I will take time to go over after this. the question is like I'm coming from ⁓ finance and ⁓ banking domain.

⁓ I'm still not able to connect dots like whatever I'm learning, how I'm gonna ⁓ apply to my domain as such. So, for example, in the diagram where ⁓ we have this embedded layer where the inputs are being passed. ⁓ for example, if I have ⁓ the time series of Apple and Google and I want to predict next five days of ⁓ prices, ⁓ right?

What will be the input they have to give give from the finance perspective so that the generated output will be one vector with the five ⁓ next five days price? Mm-hmm. ⁓ So I cannot have a you know definite answer for this. There are different ways, and I'm sure there are ⁓ lots of papers around it. Typically, a a very ⁓ good baseline or a straightforward solution for this type of

Thing is just to convert everything into a text-to-text ⁓ setup. So basically, when you have a just a sequence of ⁓ prices, you can just train a transformer ⁓ or an LLM. You pass this sequence along with all other events and metadata that you want to consider during training. And then the output is going to be the next ⁓ price, like the price for the next day or something like that, depending on what your granularity is.

⁓ I really cannot comment more because there there are lots of different ways. But I suggest you take a look at this paper. It's a T five paper. It's not related to finance, but basically they are using a transformer and they are using it for different tasks and they convert everything to text input, text output, like a translation, ⁓ sentiment and everything. So you may find it ⁓ you may you may get some ideas if if you go over this. Okay. ⁓

Sneha Mehra (02:26:28)  
So it's called T five paper. Yep, sure. Okay. If you can put on chat or should I have to ⁓ I mean it's going to be in the recording. I ⁓ I yeah. Okay if if if you don't find it, just post in the channel. I would share it with you. Sure. Thanks, Early.

And and just one comment, Ali, before I go. ⁓ any suggestion like for the project two, right? Do I have to do ⁓ preparation before I come to the cohort? ⁓ Or this time I I I thought like I'll go ⁓ with your pace and learn side by side, right? ⁓ is there any preparation that I have to do before I go for project two? Because this project I could not cope up with the pace. Okay. So yes, I have some thoughts here.

So for the project two, first I would strongly suggest to go over the guided learning because that will set up a lot of the foundations and concepts. Project two, I also tried to improve it based on all the feedback. So now it comes with a lot of links that you can go and look at the documentations. So that's one thing that I expect it it would be probably easier to follow and more intuitive. And second is that project two is also providing you like a step-by-step, ⁓ like very critical.

Very, very clear like steps to implement. ⁓ and the third thing is that it's more abstract. So I think so. From this project forward, project two forward, ⁓ things are gonna be more at the application level and more abstract. So all ⁓ basically a typical process would be ⁓ you want to let's say load PDFs, you search the web, you see which libraries do provide ⁓ like that functionality, you read the documentation, it's going to be typically one or two lines of code.

You you just copy paste or you just write it to load your PDFs and then that's it. So yes, a starting from project too. Perhaps things would be more like abstract, like higher level, just using some tools or libraries to load things and just connect things together. We start relying on Langchain and their functionalities. So ⁓ yeah, and still if you have questions, feel free to post them in the in the channel. But I think if you first start with guided learning.

Sneha Mehra (02:28:39)  
And then you just follow the project and all the links provided, it you it should be intuitive. Okay. Thank you, Ali.

Fine.

Sneha Mehra (02:28:53)  
Hey ⁓ Ali, quick few questions. How much of ⁓ what we learned as a programmer in this week one could I reapply to ⁓ building an LLM, other kinds of LLMs? Like right now we have seen a text text generation model. If I were to do text to speech, speech to text, let's say document intelligence, ⁓ or if I'm using anything other than GPT two, how much of what I've learnt in week one would I be able to reapply?

or what sort of changes you know that I could we could expect when building other kinds of models with respect to at least building an LLM? Yeah, that's an awesome question. So basically the sequence of chapters we have in the cohort, like the sequence of topics we have in the cohort basically tries to gradually build on top of the previous guided learnings. So for example in in you know in RAGs we we learn about other more advanced models and load them.

And then we see more ⁓ like different things in the coding as well. By the end, perhaps we should be able to build something like very, very simplified version of ⁓ like a multimodal agent kind of a thing. So we load some reasoning model, we have some like rag-based setup, we have image generation, video generation, and the chatbot. So we can just print ⁓ type something and then it would answer. ⁓ but a lot of things are ⁓ going to be helpful if you want to switch to other things, also, right? So for example.

I mean, once you learn how these things work, ⁓ you can just switch to best, like most ⁓ state of the art models and just play around with them, whether it's LLM or whether it's a reasoning model, right? Once you learn about the RAC and like a simple embedding model setup, you can just switch to the best embedding model out there and connect them, and then suddenly you would go from a very simple ⁓ prototype to something a lot more stronger based on the best models out there.

So yeah, the hope is to just build gradually on top of ⁓ you know, each week, on top of the previous week.

Sneha Mehra (02:31:02)  
Peter. Yes. ⁓ thank you again. ⁓ I just want to make sure if I understand something correctly and then I have a follow up. ⁓ LLM stack, if you want to call it that, it has a tokenizer, an embedding layer, and a and then the hidden layers where all this stuff happens in the box above. ⁓ a tokenizer is fixed for a given model.

You can't just use any tokenizer with any model because all the training is built around those token IDs that get generated. And then the same thing with embedding layer, right? The vector output from embedding ⁓ is ⁓ is fixed after you start training. Right? Yes. ⁓ thank you. And then ⁓ if if you can, ⁓ going back to the question I asked initially, can you summarize your intention in including this content? I don't

I I I mean this in an honest way. Just ⁓ I'm interested in in what principles you are trying to emphasize in in this first week, if you can share. Yep. So so first week is more focused on the you know the foundations and ⁓ like learning whatever is necessary to know like how ⁓ like simpler versions of Chat GPT work and how they can be really built or at least used in in terms of coding.

⁓ and it's it's basically the necessary knowledge for future because you know a lot of things like around rag, rag is basically has two parts, right? One part is retrieval and then the other part is generation. And generation is entirely ⁓ relies on the LLM. ⁓ And ⁓ so so that's why like week one is just about LLMs, how they are trained, what are all the different pieces, how they actually generate text token by token, so that when we switch to rag, we don't have to ⁓

Don't have to worry about the generation part, right? So then we can focus on the retrieval part only. And then the third week is let's say agents, right? Now for agents, still LLM is one piece of it, but it's a very small piece of it. There is just ⁓ now new concepts like function calling, MCP, you know, all those things. And also the multi-step agent. So there is this loop that is being introduced.

Sneha Mehra (02:33:18)  
So basically the the whole design was was based on the fact that first we understand the foundations and then slowly we build on top of ⁓ each week. And then you know at the very end we would ⁓ the hope was to cover the most you know ⁓ important topics or concepts in AI engineering, like or relevant in AI engineering. And also we get a sense of all the popular features out there, like how they're built, like you know, like reasoning models or a a feature like

deep research, right, by Anthropic or OpenAI or even Google that they're providing. Image generation or video generation, how these things are happening. And also the the multimodal agent. So that's how the the ⁓ how was the this is how the thinking process was for for coming up with this sequence of topics. I hope it it clarifies a bit.

Sneha Mehra (02:34:14)  
Great. ⁓

Time here.

Thank you, Aliyah, once again. So quickly, ⁓ the picture you are showing here, so is it correct to say that an LLM consists of tokenizer, embedded embedded embedding layer, and this transformer?

⁓ sorry, can you repeat your question again? An LLM consists of ⁓ this ⁓ is this whole thing is LLM? Is it correct to say? Tokenization ⁓ I I don't have a good answer for. I mean, you would see different terminologies if you go online, but basically the LLM at the very core ⁓ is the transformer. And transformer is it means basically the embedding layer and the transformer blocks and the linear layer. Even tokenization is not considered part of the ⁓

Transformer model. So, but again, there are different terminologies, different different people have different levels of ⁓ you know, granularity when thinking about this. ⁓ but I would say perhaps the LLM can be think of the embedding layer and the transformer blocks and the of course the linear, like anything with a parameter. Okay, got it. And ⁓ I have ⁓

Sneha Mehra (02:35:30)  
One more quick, I mean two more quick questions. So one, the the exercise you did, the notebook with the code, how soon that can be available? ⁓ not too long. I mean I'll I'll just try to ⁓ clean it up and then I would push it in the repo. And where will okay in the repo. Okay, that's good. Thank you. And last question. This parameter and hyperparameter, can you expand a little bit? I I'm trying to ⁓ wrap my head around things. Sure. Parameters are

The models parameters. So it would they would be learned. You are not gonna touch them. You basically let the training process they would be updated until you come up with some optimal set of parameters. So these are things that would be automatically updated during the training. Hyperparameters are not really trainable. You just set them to some fixed values. For example, you set the hidden dimension here to four. It's gonna remain four for the entire training process for the number of blocks.

So yeah, that's basically the difference of hyperparameter and parameters. They're just the the naming ⁓ might be a little bit confusing, but parameters is just anything about the model that is going to be learned during the training process.

Sneha Mehra (02:36:47)  
So all right, so ⁓ I guess we'll have to ⁓ end soon. So I'll take the three remaining questions and then we can end.

Radzieh

⁓ Thank you so much. I appreciate everyone taking so much of time. I had put in my private in your private chat. ⁓ So there is this LLM's output. There's a section that is pasted, passing a token sequence through ⁓ LLM Neil's. ⁓ This is in the solution notebook. ⁓ In 2.4. Yeah. ⁓ So I'm trying to get some ⁓ intuition that that first section, first text that you have put in, which is a token of just above the code. Yeah.

⁓ can you include this last part of your question? So the text is here and then I'm No, no, no. No, the the text or the ⁓ you know words written above the code, right? So the sentences written above the code. If you scroll up, yeah. Yeah, that's ⁓ so I want to get that intuition on passing a token sequence through LMEs. See the primitives of tensor logits and the shape of it. That's number one. I want to get some.

Thoughts on how to interpret that, you know, tense tensor being a multi-dimensional array, knowledge being what is being predicted. And shape that say sequence. I didn't understand, you know, ⁓ what those are and then also something around what's offmax. ⁓ Yeah. What's off? Okay. So batch size sequence lengths or cap size. This is what you want clarity. Yeah. So if you if you think of this output, what is the shape? So we have sequence length.

Sneha Mehra (02:38:30)  
We have three. And then we have ⁓ hidden dimension. Now then we apply linear. We are gonna apply the linear. I mean in the code. In the code, we are applying linear on all of these and then discarding the rest. So ⁓ that's a minor detail. But let's say after you apply this linear, we are you are mapping the dimension to the vocabulary size, right? Because you want the to treat those as the probabilities associated to each token.

Right. ⁓ So if you see the output of the linear, we would still have if if I do something like this, right? And we pass all these three to linear. The output would be sequence length times vocabulary size. Now this is one example. In practice, we sent multiple examples in each round. And that's what batch size refers to. So in each round of training, we don't send one example. We send like 10 examples or thousand examples.

And each examples is being processed ⁓ simultaneously and independently.

Yeah, right. So that's where this comes from. And softmax. Softmax we apply it because if you don't have softmax, these are gonna be just numbers, negative numbers, positive numbers. you cannot really treat them as probability. So softmax is a function which is just re which is just doing that. You pass a bunch of numbers, it just exponenti exponentiates them and then ⁓ make sure that it they sum up to one. So this way, since it ensures that everything is positive and sums up to one, you can treat them as probabilities.

⁓ So ⁓ completely unrelated ⁓ question on use cases. So given that this is already probabilistic, right? So I could see the use cases being used in chat parts and coming up with responses, because enterprises ⁓ typically we build systems ⁓ based on deterministic outcomes. So what are the use cases that people use this in other than you know language interpretation chat parts, et cetera?

Sneha Mehra (02:40:31)  
So we would see a lot of yeah, we would see a lot of use cases, you know. I mean, for example, in VIC2, the main project is like building a specialized, like a chatbot, for instance. So we start from some LLM and we try to customize it to a chatbot for ⁓ some imaginary like ⁓ retailers or something like that. So ⁓ I can also share more if I if I can think of more use cases, I can definitely share more in the in the channel.

Sneha Mehra (02:41:05)  
Sonia. Yes. ⁓ Yes. Hi. Hi, Ali. So just a request that when you push it, you know, you just said you're going to push it to GitHub. Could you could you push it in a way that we can ⁓ open it in Git Collab? Because I don't have my local setup yet. yeah, yeah, yeah. Of course. I mean, it's still the same thing works. ⁓ I'm gonna push it. You can just open it in Google Collab and just run it. Yeah. And ⁓ and somebody ⁓

Gave, you know, we have one error happening, right? So somebody gave a script. So did you you are going to review it before we use it? So that, you know, that error in ⁓ when we open it in ⁓ GitHub after we push it, we are on some of cells we are getting an error. Few people reported it. So will you review that script and then we can use it? I I mean I didn't try it yet. ⁓ Yeah, I so yeah, I I think the same day like ⁓

The same day it was fixed. I mean, try it. If it doesn't work, just let me know. But it should be fixed. ⁓ no, no, it is like when you when we push this code ⁓ Google Collab to GitHub and we open it. ⁓ there is that one. Yeah. ⁓ that yeah, I can yeah, I can I can look into it. I I have to see. I mean, you so when you want to push the code in your own git, you face ⁓ that error.

I I I can look into it and see what the issue is. Yeah, yeah. It's like when I once I ⁓ push it and ⁓ on GitHub and then I I try to open it, it says there is some rendering error that it doesn't ⁓ like it's for TikToken. It's for TikToken one I know and ⁓ it's for ⁓ one two two of them I faced that issue. So then somebody posted ⁓ that they ⁓ use this script, so ⁓ I I didn't.

So if you can look at it. But but maybe maybe you can also try the same as script because it may work. I mean if it's shared, because ⁓ basically this is something that is happening ⁓ on ⁓ when you want to push on your own git repo, right? So this is not something that I can push to the to the main git repo. So I would suggest try that if you still face the issue, just post and or DM me and then we will figure it out together. ⁓ Okay, okay. All right, and just running.

Sneha Mehra (02:43:30)  
⁓ please if you can make sure that everybody is muted during the lecture because kind of I think I was hearing some background sound during the while you were giving some the giving the training today. ⁓ sure. I think everyone was muted by default, but sure. I looked I looked up and I kind of saw also like four or five people, so I don't know, some little sound it was a little disturbing. So appreciate it and like one guy said.

was from finance background, he was saying that, he lost you. I also kind of, you know, I did the whole thing by myself, but I I just wanted to share that okay, I lost you after some time and I'll take some I'll now go through the code that you will give and all your explanation to understand better. So sure, yeah. Thanks, thanks. Yeah, sure. Thanks for the feedback. ⁓ thank you. Yeah, I mean ⁓

From then I started copying the code. Basically, the purpose was to not repeat basically the same thing. So I think if you watch the recording, it should become more intuitive that why that was the case. So a lot of the code that was pasted are basically the exact same like loop for loops to load the model and tokenize it from the previous sales that we already discussed. But but yeah, thanks for letting me know. I'll I'll ⁓ I ⁓ yeah, I I try to make sure that it's going to be more ⁓ Yeah.

Yeah, even you know, like ⁓ like like see, everybody's at a different level. So, you know, I'm sure. So like ⁓ when you explain it it really adds ⁓ like to me it really adds a lot of value. Like one somebody was asking and recently somebody asked, and then you explained that that linear is over there and batch is over there. So I could not earlier make the connection that you know, all that batch and sequence length is that thing, but now

During the question when somebody asked and then you pointed to the diagram, then then I could ⁓ connect that, it's that same thing. So it's just ⁓ you know, just like ⁓ because I'm yeah, no, that yeah. Yeah, that's awesome. And that that that's awesome and that's great to hear that, you know. ⁓ so I think that's that's one of the the main values of the you know, question answering. So I think there's some questions and then he would help to clarify more.

Sneha Mehra (02:45:49)  
I'll I'll also try to make things as clear as possible ⁓ in the project, but still I'm I'm sure that there would be a lot of questions that would also add value to everyone else. So but great, great to hear that. Okay, thank you. And you will make some announcements.

