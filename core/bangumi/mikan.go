package bangumi

import (
	"GoBangumi/model"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/golang/glog"
	"net/url"
	"strconv"
	"strings"
)

const (
	MikanBaseUrl         = "https://mikanani.me"                                     // Mikan 域名
	MikanIdXPath         = "//a[@class='mikan-rss']"                                 // Mikan番剧id获取XPath
	MikanBangumiUrlXPath = "//p[@class='bangumi-info']/a[contains(@href, 'bgm.tv')]" // Mikan番剧信息中bangumi id获取XPath
)

var MikanInfoUrl = func(id int) string {
	return fmt.Sprintf("%s/Home/Bangumi/%d", MikanBaseUrl, id)
}

type Mikan struct {
	Info *model.Bangumi
}

func NewMikan() Bangumi {
	return &Mikan{}
}

func (b *Mikan) Parse(opt *model.BangumiParseOptions) *model.Bangumi {
	b.Info = &model.Bangumi{
		Name: opt.Name,
	}
	glog.V(3).Infof("获取「%s」信息开始...\n", opt.Name)
	b.parseMikan1(opt.Url, b.Info)
	b.parseMikan2(b.Info)
	b.parseBangumi(b.Info)
	glog.V(3).Infof("获取「%s」信息成功！更名为「%s」\n", opt.Name, b.Info.FullName())
	return b.Info
}

// parseMikan1
//  Description 解析mikan rss中的link页面，获取当前资源的mikan id
//  Receiver b *Mikan
//  Param url_ string
//  Param bgm *model.Bangumi
//
func (b *Mikan) parseMikan1(url_ string, bgm *model.Bangumi) {
	glog.V(5).Infof("步骤1，解析Mikan，%s\n", url_)
	doc, err := htmlquery.LoadURL(url_)
	if err != nil {
		glog.Errorln(err)
		return
	}
	miaknLink := htmlquery.FindOne(doc, MikanIdXPath)
	href := htmlquery.SelectAttr(miaknLink, "href")
	u, err := url.Parse(href)
	if err != nil {
		glog.Errorln(err)
		return
	}
	query := u.Query()
	if query.Has("bangumiId") {
		id, err := strconv.Atoi(query.Get("bangumiId"))
		if err != nil {
			glog.Errorln(err)
			return
		}
		bgm.SubID = id
	}
}

// parseMikan2
//  Description 通过mikan id解析mikan番剧信息页面，获取bgm.tv id
//  Receiver b *Mikan
//  Param bgm *model.Bangumi
//
func (b *Mikan) parseMikan2(bgm *model.Bangumi) {
	url_ := MikanInfoUrl(bgm.SubID)
	glog.V(5).Infof("步骤2，解析Mikan，%s\n", url_)
	doc, err := htmlquery.LoadURL(url_)
	if err != nil {
		glog.Errorln(err)
		return
	}
	bangumiUrl := htmlquery.FindOne(doc, MikanBangumiUrlXPath)
	href := htmlquery.SelectAttr(bangumiUrl, "href")

	//fmt.Println(href)
	hrefSplit := strings.Split(href, "/")
	bgmId, err := strconv.Atoi(hrefSplit[len(hrefSplit)-1])
	if err != nil {
		glog.Errorln(err)
		return
	}
	bgm.ID = bgmId
}

// parseBangumi
//  Description 从bangumi网站获取信息
//  Receiver b *Mikan
//  Param bgm *model.Bangumi
//
func (b *Mikan) parseBangumi(bgm *model.Bangumi) {
	glog.V(5).Infof("步骤3，解析Bangumi，%d\n", bgm.ID)
	bangumi := NewBgm()
	newBgm := bangumi.Parse(&model.BangumiParseOptions{
		ID: bgm.ID,
	})
	newBgm.ID = bgm.ID
	newBgm.SubID = bgm.SubID
	bgm = newBgm
}
