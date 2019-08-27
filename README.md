# tweets-stream
publish, search tweets, downstream the tweets

1. set mysql=root:root@tcp(127.0.0.1:3306)/test
2. set HTTP_PROXY=127.0.0.1:8380

3. set CONSUMER_SECRET=yous
4. set CONSUMER_KEY=xxx
5. set TOKEN_SECRET=xxxxdfd
6. set TOKEN=xcxdf

--------------------------------------------------------
#get specified user's twitter
tweets_stream.exe -person -q somename -c 10 -mysql %mysql%

#query twitter according the keyword
tweets_stream.exe -s -q trump -c 10 -mysql %mysql%

#post twitter [update status]
tweets_stream.exe -p -m "make earth great again"

#stream filter the twitter according the keyword
tweets_stream.exe -stream -q trump -mysql %mysql%
--------------------------------------------------------
##if you specify mysql, the tweet will store in mysql
--------------------------------------------------------

#please contact davidwang2006@aliyun.com
