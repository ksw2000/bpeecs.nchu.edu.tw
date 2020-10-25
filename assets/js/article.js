function stripHtml(html) {
  var tmp = document.createElement("div");
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || "";
}

function hash(str) {
  var h = 0;

  for (var i = 0; i < str.length; i++) {
    h = (h << 5) - h + str.charCodeAt(i);
    h |= 0;
  }

  return h;
} // input: string(json)
// output: string(html code)


function loadAttchment(str) {
  // load attachment
  if (str === "") {
    return "";
  }

  try {
    var attachment = "";
    var parse = JSON.parse(str);
    var len = parse.client_name.length;

    for (var i = 0; i < len; i++) {
      attachment += '<li><a href="' + parse.path[i] + '">' + parse.client_name[i] + '</a></li>';
    }

    return attachment;
  } catch (e) {
    return "";
  }
}

function appendMoreInfo(obj) {
  $(obj).next().slideToggle();
}

function articleTypeDecoder(key) {
  var dict = {
    "normal": "一般消息",
    "activity": "演講 & 活動",
    "course": "課程 & 招生",
    "scholarships": "獎學金",
    "recruit": "徵才資訊"
  };
  return dict[key];
}

function loadNews(scope, type, from, to) {
  if (type === undefined) {
    type = 'normal';
  }

  if (from === undefined) {
    from = '';
  }

  if (to === undefined) {
    to = '';
  }

  return new Promise(function (resolve, reject) {
    $.ajax({
      url: '/function/get_news',
      data: {
        'scope': scope,
        'type': type,
        'from': from,
        'to': to
      },
      type: 'GET',
      success: function success(data) {
        resolve(data);
      },
      error: function error(err) {
        reject(err);
      },
      dataType: 'json'
    });
  });
}

/*
function loadNewsOnlyTitle(scope) {
  return new Promise(function (resolve, reject) {
    loadNews(scope, 'normal').then(function (data) {
      data = data.NewsList;
      var len = data.length;
      var ret = "";

      for (var i = 0; i < len; i++) {
        data[i].Publish_time = $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd");
        ret += "<div class=\"article\" data-id=\"".concat(data[i].Id, "\" style=\"cursor: pointer\" onclick=\"window.location='/news?id=").concat(data[i].Id, "'\"><div class=\"candy-header\"><span class=\"single\">").concat(data[i].Publish_time, "</span></div><h2 class=\"title\">").concat(data[i].Title, "</h2></div>");
      }

      resolve(ret);
    }).catch(function (err) {
      reject(err);
    });
  });
}
*/

function loadNewsForWhat(what, scope, type, from, to) {
  var self = this;
  this.from = from;
  this.to = to;
  this.len = to - from + 1;

  this.load = function () {
    return new Promise(function (resolve, reject) {
      loadNews(scope, type, self.from, self.to).then(function (data) {
        var ret = self.render(data.NewsList);

        if (data.HasNext) {
          ret += "<div><button style=\"margin:0px auto;\" onclick=\"loadNext(this)\">More</button></div>";
        }

        resolve(ret);
      }).catch(function (err) {
        reject(err);
      });
    });
  };

  if (what == 'management') {
    this.render = function (data) {
      if (data == null) return "No articles";
      var len = data.length;
      var ret = '';

      for (var i = 0; i < len; i++) {
        var isDraft = data[i].Publish_time === 0 ? true : false;
        data[i].Create_time = $.format.date(new Date(data[i].Create_time * 1000), "yyyy-MM-dd HH : mm");
        data[i].Last_modified = data[i].Last_modified === 0 ? '-' : $.format.date(new Date(data[i].Last_modified * 1000), "yyyy-MM-dd HH : mm");
        data[i].Publish_time = data[i].Publish_time === 0 ? '-' : $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd HH : mm");
        var newContent = stripHtml(data[i].Content);

        if (newContent.length > 50) {
          newContent = newContent.slice(0, 80);
          newContent += "<a href=\"/news?id=".concat(data[i].Id, "\">...More</a><p></p>");
        }

        var attachment = loadAttchment(data[i].Attachment);
        var draftIcon = isDraft ? '<div class="draftIcon">draft</div>' : '';
        var draftColor = isDraft ? 'border-color:#fe6c6c;' : 'border-color:#14a1ff;';
        ret += "<div class=\"article\" data-id=\"".concat(data[i].Id, "\" style=\"").concat(draftColor, "\">\n                    <h2 class=\"title\">").concat(draftIcon).concat(data[i].Title, "</h2>");
        ret += "<div class=\"header\" onclick=\"javascript:appendMoreInfo(this)\">";
        ret += "    <div class=\"candy-header\"><span>\u5206\u985E</span><span>".concat(articleTypeDecoder(data[i].Type), "</span></div>");
        ret += "    <div class=\"candy-header\"><span>\u6700\u5F8C\u7DE8\u8F2F</span><span class=\"orange\">".concat(data[i].Last_modified, "</span></div>");
        ret += "</div>";
        ret += "<div style=\"display: none;\">";
        ret += "  <div class=\"candy-header hide-less-500px\"><span>\u5EFA\u7ACB\u65BC</span><span class=\"red\">".concat(data[i].Create_time, "</span></div>");
        ret += "  <div class=\"candy-header hide-less-500px\"><span>\u767C\u4F48\u65BC</span><span class=\"green\">".concat(data[i].Publish_time, "</span></div>");
        ret += "</div>";
        ret += "\n                    <div class=\"content\">\n                        ".concat(newContent, "\n                    </div>\n                    <div id=\"attachmentArea\">\n                        <ul>").concat(attachment, "</ul>\n                    </div>\n                    <div class=\"buttonArea\" style=\"text-align: right;\">\n                        <button id=\"read\" onclick=\"window.location='/news?id=").concat(data[i].Id, "'\"\" class=\"border\">\u95B1\u8B80</button>\n                        <button id=\"delete\" onclick=\"javascript:delete_what(this, 'news', ").concat(data[i].Id, ")\" class=\"red\">\u522A\u9664</button>\n                        <button id=\"publish\" onclick=\"javascript:edit_news(").concat(data[i].Id, ")\" class=\"blue\">\u7DE8\u8F2F</button>\n                    </div>\n                </div>\n                ");
      }

      return ret;
    };
  } else if (what == 'brief') {
    this.render = function (data) {
      if (data == null) return "No articles";
      var len = data.length;
      var ret = "";

      for (var i = 0; i < len; i++) {
        data[i].Publish_time = $.format.date(new Date(data[i].Publish_time * 1000), "yyyy-MM-dd");
        var newContent = stripHtml(data[i].Content);

        if (newContent.length > 30) {
          newContent = newContent.slice(0, 80);
          newContent += "...<a href=\"/news?id=".concat(data[i].Id, "\">\u7565</a>");
        }

        var attachment = loadAttchment(data[i].Attachment);
        ret += "\n                <div class=\"article\" data-id=\"".concat(data[i].Id, "\">\n                    <h2 class=\"title\">").concat(data[i].Title, "</h2>\n                    <div class=\"header\" onclick=\"javascript:appendMoreInfo(this)\">\n                        <div class=\"candy-header\"><span>\u767C\u4F48\u65BC</span><span>").concat(data[i].Publish_time, "</span></div>\n                    </div>\n                    <div style=\"display:none;\">\n                ");
        ret += "<div class=\"candy-header\"><span>\u5206\u985E</span><span class=\"green\">".concat(articleTypeDecoder(data[i].Type), "</span></div>");
        ret += "<div class=\"candy-header\"><span>\u767C\u6587</span><span class=\"cyan\">@".concat(data[i].User, "</span></div>");
        ret += "\n                    </div>\n                    <div class=\"content\">\n                        ".concat(newContent, "\n                    </div>\n                    <div id=\"attachmentArea\">\n                        <ul>").concat(attachment, "</ul>\n                    </div>\n                    <p></p>\n                    <div class=\"buttonArea\" style=\"text-align: right;\">\n                        <button id=\"attachment\" onclick=\"window.location='/news?id=").concat(data[i].Id, "'\"\n                                style=\"display: inline-block;\">\u95B1\u8B80\u5168\u6587</button>\n                    </div>\n                </div>\n                ");
      }

      return ret;
    };
  }

  this.next = function () {
    self.from = self.to + 1;
    self.to += self.len;
    return self.load();
  };
}

function loadNewsById(newsID) {
  return new Promise(function (resolve, reject) {
    $.ajax({
      url: '/function/get_news?id=' + newsID,
      type: 'GET',
      success: function success(data) {
        resolve(data);
      },
      error: function error(err) {
        reject(err);
      },
      dataType: 'json'
    });
  });
}

/*
function loadPublicNewsById(id) {
  return new Promise(function (resolve, reject) {
    loadNewsById(id).then(function (data) {
      var ret = {};
      ret.text = "";
      data.Publish_time = $.format.date(new Date(data.Publish_time * 1000), "yyyy-MM-dd");
      data.Content = data.Content;
      var attachment = loadAttchment(data.Attachment);
      ret.text += "\n            <div class=\"article\" data-id=\"".concat(data.Id, "\" style=\"border:0px;\">\n                <div class=\"header\" onclick=\"javascript:appendMoreInfo(this)\">\n                    <div class=\"candy-header\"><span>\u767C\u4F48\u65BC</span><span>").concat(data.Publish_time, "</span></div>\n                </div>\n                <div class=\"header\" style=\"display: none;\">");
      ret.text += "<div class=\"candy-header\"><span>\u5206\u985E</span><span class=\"green\">".concat(articleTypeDecoder(data.Type), "</span></div>");
      ret.text += "<div class=\"candy-header\"><span>\u767C\u6587</span><span class=\"cyan\">@".concat(data.User, "</span></div>");
      ret.text += "</div>\n                <div class=\"content\">\n                    ".concat(data.Content, "\n                </div>\n                <div id=\"attachmentArea\">\n                    <ul>").concat(attachment, "</ul>\n                </div>\n            </div>\n            ");
      ret.json = data;
      resolve(ret);
    }).catch(function (err) {
      reject(err);
    });
  });
}
*/
