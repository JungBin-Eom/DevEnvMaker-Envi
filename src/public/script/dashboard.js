// 오른쪽 상단 사용자 이름 표시
$.get("/user", function(user) {
  $("#user-id").html(user.id)
  $("#now-token").html(user.github_token)
});

// 프로젝트 생성 버튼 클릭
$("#createproject").click(function(){
  if ($('#newprojectname-input').val().indexOf(' ') != -1) {
    alert("프로젝트 이름에는 공백 문자를 포함할 수 없습니다.");
    $('#newprojectname-input').val('');
  } else {
    fetch('/project', {
      method: 'post',
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
          name: $('#newprojectname-input').val(),
          description: $('#newprojectdesc-input').val(),
        })
    })
    .then(res => res.json())
    .then(res => {
      if (res.success == true) {
        alert("프로젝트 생성 완료!");
        location.href="/";
      } else if (res.count != 0) {
        alert("같은 이름의 프로젝트가 존재합니다.");
        $('#newprojectname-input').val('');
      } else {
        location.href="/html/404.html";
      }
    });
  }
});

// 프로젝트 삭제 버튼 클릭
$("#delete-project-btn").click(function(){
  var projectName = $("h1").text();
  fetch('/project', {
    method: 'delete',
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
        name: projectName,
      })
  })
  .then(res => res.json())
  .then(res => {
    if (res.success == true) {
      $("#"+projectName).remove();
      alert("프로젝트 삭제 완료!")
      location.href="/"
    } else {
      alert("삭제 중 문제가 발생했습니다.")
    }
  });
});

// 애플리케이션 생성 버튼 클릭
$("#createapplication").click(function(){
  if ($('#newapplicationname-input').val().indexOf(' ') != -1) {
    alert("애플리케이션 이름에는 공백 문자를 포함할 수 없습니다.");
    $('#newapplicationname-input').val('');
  } else {
    var beforeParams = get_query(document.referrer)
    var projectName = beforeParams.name;
    fetch('/app', {
      method: 'post',
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
          name: $('#newapplicationname-input').val(),
          description: $('#newapplicationdesc-input').val(),
          project: projectName,
          runtime: $('#application-select option:selected').val(),
        })
    })
    .then(res => res.json())
    .then(res => {
      if (res.success == true) {
        alert("애플리케이션 생성 완료!");
        location.href=document.referrer;
      } else {
        location.href="/html/404.html"
      }
    });
  }
});

// 애플리케이션 삭제 버튼 클릭
$("#delete-app-btn").click(function(){
  var appName = $("h1").text();
  fetch('/app', {
    method: 'delete',
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
        name: appName,
      })
  })
  .then(res => res.json())
  .then(res => {
    if (res.success == true) {
      $("#"+appName).remove();
      alert("애플리케이션 삭제 완료!")
      location.href="/"
    } else {
      alert("삭제 중 문제가 발생했습니다.")
    }
  });
});


// 토큰 등록 버튼 클릭
$("#register-token").click(function(){
  fetch('/user/token', {
    method: 'post',
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
        token: $('#token-input').val(),
      })
  })
  .then(res => res.json())
  .then(res => {
    if (res.success == true) {
      alert("토큰 등록 완료!")
      location.href="/"
    } else {
      location.href="/html/404.html"
    }
  });
});

// 프로젝트, 애플리케이션 목록 가져오기
$.get("/project", function(items) {
  if (items.length == 0) {
    projectList.append("<h6 class='collapse-header'>There is no project.</h6>");
  } else {
    items.forEach(e => {
      addProject(e)
    });
  }
});

// 프로젝트마다 태그 추가
var projectList = $("#project-list")

var addProject = function(item) {
  projectList.append("<a class='collapse-item project' id='"+item.name+"' href='#none' onclick='goProjectPage(this)' >"+item.name+"</a>");
};

// $.get("/app", function(items) {
//   if (items.length == 0) {
//     appList.append("<h6 class='collapse-header'>There is no application.</h6>");
//   } else {
//     items.forEach(e => {
//         addItem(e)
//     });
//   }
// });

// var appList = $("#app-list")

// var addApp = function(item) {
//   appList.append("<a class='collapse-item application' id='"+item.name+"' href='#none'>"+item.name+"</a>");
// };

function goProjectPage(obj){
  projectName = $(obj).attr("id");
  location.href="../html/project.html?name="+projectName;
};

// redirect시 url 파라미터 쿼리 얻기
function get_query(url){
  var qs = url.substring(url.indexOf('?') + 1).split('&');
  for(var i = 0, result = {}; i < qs.length; i++){
    qs[i] = qs[i].split('=');
    result[qs[i][0]] = decodeURIComponent(qs[i][1]);
  }
  return result;
}

var parameters = get_query(document.location.href);
if (parameters.name != undefined) {
  var itemName = parameters.name;
  $("#name").text(itemName);
  if (document.location.href.indexOf("project") != -1) {
    $.get("/project/"+itemName, function(project) {
      $("#description").text(project.description);
    });
    $.get("/app", function(items) {
      if (items.length == 0) {
        $("#app-list").append("<p>There is no application on project '"+parameters.name+"'.</p>");
      } else {
        items.forEach(e => {
          if (e.project == parameters.name) {
            $("#app-list").append("<li><b>"+e.name+"</b><p>"+e.description+"</p>"+"<a rel='nofollow' href='../html/application.html?name="+e.name+"'>See Detail &rarr;</a>");
          }
        });
      }
    });
  } else if (document.location.href.indexOf("application") != -1) {
    var beforeParams = get_query(document.referrer)
    var projectName = beforeParams.name;
    $.get("/app/"+projectName+"/"+itemName, function(app) {
      $("#description").text(app.description);
    });
  }
}

var addApp = function(item) {
  appList.append("<a class='collapse-item application' id='"+item.name+"' href='#none'>"+item.name+"</a>");
};

// github link 클릭
$("#github-link").click(function(){
  fetch('/github/name', {
    method: 'get',
    headers: {
      "Content-Type": "application/json",
    }
  })
  .then(res => res.text())
  .then(res => {
    location.href="https://github.com/"+res
  });
});
