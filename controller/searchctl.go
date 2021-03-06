// Pipe - A small and beautiful blogging platform written in golang.
// Copyright (C) 2017, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package controller

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/b3log/pipe/model"
	"github.com/b3log/pipe/service"
	"github.com/b3log/pipe/util"
	"github.com/gin-gonic/gin"
	"github.com/vinta/pangu"
)

func searchAction(c *gin.Context) {
	if "GET" == c.Request.Method {
		showSearchPageAction(c)

		return
	}

	result := util.NewResult()
	defer c.JSON(http.StatusOK, result)

	args := map[string]interface{}{}
	if err := c.BindJSON(&args); nil != err {
		result.Code = -1
		result.Msg = "parses search request failed"

		return
	}

	blogID := getBlogID(c)
	key := ""
	if nil != args["key"] {
		key = args["key"].(string)
	}
	page := 1
	if nil != args["p"] {
		page = int(args["p"].(float64))
	}
	articleModels, pagination := service.Article.GetArticles(key, page, blogID)
	var articles []*model.ThemeArticle
	for _, articleModel := range articleModels {
		var themeTags []*model.ThemeTag
		tagStrs := strings.Split(articleModel.Tags, ",")
		for _, tagStr := range tagStrs {
			themeTag := &model.ThemeTag{
				Title: tagStr,
				URL:   getBlogURL(c) + util.PathTags + "/" + tagStr,
			}
			themeTags = append(themeTags, themeTag)
		}

		authorModel := service.User.GetUser(articleModel.AuthorID)
		if nil == authorModel {
			logger.Errorf("not found author of article [id=%d, authorID=%d]", articleModel.ID, articleModel.AuthorID)

			continue
		}

		mdResult := util.Markdown(articleModel.Content)
		article := &model.ThemeArticle{
			Title:    pangu.SpacingText(articleModel.Title),
			Abstract: template.HTML(mdResult.AbstractText),
			URL:      getBlogURL(c) + articleModel.Path,
			Tags:     themeTags,
		}

		articles = append(articles, article)
	}

	data := map[string]interface{}{}
	data["articles"] = articles
	data["pagination"] = pagination
	result.Data = data
}

func showSearchPageAction(c *gin.Context) {
	t, err := template.ParseFiles(filepath.ToSlash(filepath.Join(util.Conf.StaticRoot, "console/dist/search/index.html")))
	if nil != err {
		logger.Errorf("load search page failed: " + err.Error())
		c.String(http.StatusNotFound, "load search page failed")

		return
	}

	t.Execute(c.Writer, nil)
}
