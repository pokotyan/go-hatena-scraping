package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	db "firebase.google.com/go/db"
	"github.com/PuerkitoBio/goquery"
	"google.golang.org/api/option"
)

type FireBase struct {
	opt    option.ClientOption
	ctx    context.Context
	app    *firebase.App
	client *db.Client
}
type HatenaLink struct {
	Users string
	Title string
	Link  string
	Desc  string
}

// https://firebase.google.com/docs/database/admin/save-data
func initFireBase() *FireBase {
	opt := option.WithCredentialsFile("./credentials/serviceAccountKey.json")
	ctx := context.Background()
	config := &firebase.Config{DatabaseURL: "https://hatena-75088.firebaseio.com/"}
	app, err := firebase.NewApp(ctx, config, opt)

	if err != nil {
		fmt.Printf("error initializing app: %v", err)
	}

	client, err := app.Database(ctx)

	if err != nil {
		log.Fatalln(err)
	}

	return &FireBase{opt, ctx, app, client}
}

func (f *FireBase) isExists(category string, date string) bool {
	var hl HatenaLink
	err := f.client.NewRef("hatena/hot-entry").Child(category).Child(date).Get(f.ctx, &hl)

	if err != nil {
		log.Fatalln(err)
	}

	if hl.Title == "" {
		return false
	}

	fmt.Printf("%v(%v)のデータはすでに保存ずみ\n", category, date)
	return true

}

func scraping(category string, date string) HatenaLink {
	doc, err := goquery.NewDocument("http://b.hatena.ne.jp/hotentry/" + category + "/" + date)
	if err != nil {
		fmt.Print("url scarapping failed")
	}
	users := doc.Find("#container > div.wrapper > div > div.entrylist-main > section > div > ul > li > div > div.entrylist-contents-main > span > a > span").Text()
	selection := doc.Find("#container > div.wrapper > div > div.entrylist-main > section > div > ul > li > div > div.entrylist-contents-main > h3 > a")
	title := selection.Text()
	link, _ := selection.Attr("href")
	desc := doc.Find("#container > div.wrapper > div > div.entrylist-main > section > div > ul > li > div > div.entrylist-contents-main > div.entrylist-contents-body > a > p.entrylist-contents-description").Text()

	data := HatenaLink{Users: users, Title: title, Link: link, Desc: desc}

	time.Sleep(1 * time.Second)

	return data
}

func main() {
	f := initFireBase()

	ref := f.client.NewRef("hatena/hot-entry")

	now := time.Now()
	newestDate := now.AddDate(0, 0, -1)
	oldestDate := time.Date(2012, 01, 01, 0, 0, 0, 0, time.Local)
	categories := []string{"it", "general", "all"}

	for !newestDate.Before(oldestDate) {
		formatedDate := newestDate.Format("20060102")

		for _, category := range categories {
			isExists := f.isExists(category, formatedDate)

			if !isExists {
				fmt.Printf("%v(%v)のデータをスクレイピング、保存します\n", formatedDate, category)

				hl := scraping(category, formatedDate)

				targetRef := ref.Child(category)

				err := targetRef.Child(formatedDate).Set(f.ctx, &hl)

				if err != nil {
					log.Fatalln("Error setting value:", err)
				}
			}
		}

		newestDate = newestDate.AddDate(0, 0, -1)
	}

	log.Println("Done")
}
