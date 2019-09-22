package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	log "github.com/sirupsen/logrus"
)

var tclient *twitter.Client

type creds struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessTokenKey    string `json:"access_token_key"`
	AccessTokenSecret string `json:"access_token_secret"`
}

func main() {
	credsFile, err := os.Open("creds.json")
	if err != nil {
		panic(err)
	}
	bv, err := ioutil.ReadAll(credsFile)
	if err != nil {
		panic(err)
	}
	var crds creds
	if err = json.Unmarshal(bv, &crds); err != nil {
		panic(err)
	}

	config := oauth1.NewConfig(crds.ConsumerKey, crds.ConsumerSecret)
	token := oauth1.NewToken(crds.AccessTokenKey, crds.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	tclient = client
	// todo: change the year, month, & day to be cli
	deleteTweets("tweets.json", time.Date(2017, time.January, 1, 0, 0, 0, 0, time.Local))
}

func deleteTweets(fileName string, deleteBefore time.Time) error {
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return err
	}

	bv, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var ids []struct {
		ID string `json:"id"`
	}
	if err = json.Unmarshal(bv, &ids); err != nil {
		return err
	}
	// ids = ids[0:5] // testing purposes

	for _, id := range ids {
		time.Sleep(time.Millisecond * 500)
		i, err := strconv.Atoi(id.ID)
		if err != nil {
			log.Errorf("error converting id: %s to int64: %v", id.ID, err)
			continue
		}
		tweet, _, err := tclient.Statuses.Show(int64(i), &twitter.StatusShowParams{})
		if err != nil {
			log.Errorf("could not get tweet %v: %v", id.ID, err)
			continue
		}

		created, err := tweet.CreatedAtTime()
		if err != nil {
			log.Errorf("error getting tweet time for id: %v : %v", id.ID, err)
			continue
		}

		if created.Unix() >= deleteBefore.Unix() {
			log.Infof("tweet %s not old enough, skipping delete", id.ID)
			continue
		}
		// log.Debugf("deleting %s", id.ID)
		// delete retweeted tweet
		if tweet.Retweeted {
			log.Infof("Tweet %v is retweeting, unretweeting...", tweet.ID)
			_, _, err = tclient.Statuses.Unretweet(tweet.ID, &twitter.StatusUnretweetParams{})
			if err != nil {
				log.Errorf("error unretweeting %s", tweet.ID)
				continue
			}
			continue
		}

		log.Infof("tweet %v is being deleted", tweet.ID)
		_, _, err = tclient.Statuses.Destroy(tweet.ID, &twitter.StatusDestroyParams{})
		if err != nil {
			log.Errorf("error deleting tweet %v", tweet.ID)
			continue
		}
	}
	return nil
}
