package bangumi

import (
	"GoBangumi/models"
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
}

func NewMikan() Bangumi {
	return &Mikan{}
}

func (b *Mikan) Parse(opt *models.BangumiParseOptions) *models.Bangumi {
	info := &models.Bangumi{
		Name: opt.Name,
	}
	glog.V(3).Infof("获取「%s」信息开始...\n", opt.Name)
	b.parseMikan1(opt.Url, info)
	b.parseMikan2(info)
	res := b.parseBangumi(info)
	if !res {
		glog.Errorln("获取Bangumi信息失败，结束此流程")
		return nil
	}
	res = b.parseThemoviedb(info)
	if !res {
		glog.Errorln("获取Themoviedb季度信息失败，默认SE01")
		info.Season = 1
	}
	glog.V(3).Infof("获取「%s」信息成功！更名为「%s」\n", opt.Name, info.FullName())
	return info
}

// parseMikan1
//  Description 解析mikan rss中的link页面，获取当前资源的mikan id
//  Receiver b *Mikan
//  Param url_ string
//  Param info *models.Bangumi
//
func (b *Mikan) parseMikan1(url_ string, info *models.Bangumi) {
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
		info.SubID = id
	}
}

// parseMikan2
//  Description 通过mikan id解析mikan番剧信息页面，获取bgm.tv id
//  Receiver b *Mikan
//  Param info *models.Bangumi
//
func (b *Mikan) parseMikan2(info *models.Bangumi) {
	url_ := MikanInfoUrl(info.SubID)
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
	info.ID = bgmId
}

// parseBangumi
//  Description 从bangumi网站获取信息
//  Receiver b *Mikan
//  Param info *models.Bangumi
//
func (b *Mikan) parseBangumi(info *models.Bangumi) bool {
	glog.V(5).Infof("步骤3，解析Bangumi，%d\n", info.ID)
	bangumi := NewBgm()
	newBgm := bangumi.Parse(&models.BangumiParseOptions{
		ID: info.ID,
	})
	if newBgm == nil {
		return false
	}
	info.ID = newBgm.ID
	info.Name = newBgm.Name
	info.NameJp = newBgm.NameJp
	info.AirDate = newBgm.AirDate
	info.Date = newBgm.Date
	info.Eps = newBgm.Eps
	return true
}

// parseThemoviedb
//  Description 从Themoviedb网站获取当前季度
//  Receiver b *Mikan
//  Param info *models.Bangumi
//
func (b *Mikan) parseThemoviedb(info *models.Bangumi) bool {
	glog.V(5).Infof("步骤4，解析Themoviedb，%s\n", info.NameJp)
	tmdb := NewThemoviedb()
	newBgm := tmdb.Parse(&models.BangumiParseOptions{
		Name: info.NameJp,
		Date: info.AirDate,
	})
	if newBgm == nil {
		return false
	}
	info.Season = newBgm.Season
	return true
}