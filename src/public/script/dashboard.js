

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
        location.href="/";
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
      location.href=document.referrer;
    } else {
      alert("삭제 중 문제가 발생했습니다.")
    }
  });
});

// build app
$("#build-app-btn").click(function(e){
  e.preventDefault();
  if ($(this).hasClass('clicked')) { 
    return false;
  } else {
    $(this).addClass('clicked').trigger('click');
  }

  $("#build-progress-bar").css({'width':"0%"});
  $("#build-progress-bar").attr('area-valuenow', '0');
  $("#build-percent").text("0%");
  var appName = $("h1").text();
  fetch('/app/build', {
    method: 'post',
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
      var path = res.job.split('-');
      $("#build-detail").attr('href', 'http://jenkins.3.35.25.64.sslip.io/job/'+path[0]+'/job/'+path[1])
      var status = 0;
      var running = false;
      var ok = false;
      var delay = 500;

      function building() {         //  create a loop function
        setTimeout(function() {   //  call a 3s setTimeout when the loop is called
          running, ok = false;
          $("#build-progress-bar").css({'width':String(status)+"%"});
          $("#build-progress-bar").attr('area-valuenow', String(status));
          $("#build-percent").text(String(status)+"%");
          if (status <= 20) {
            $("#build-progress-bar").attr('class', 'progress-bar bg-danger');
          } else if (status <= 50) {
            $("#build-progress-bar").attr('class', 'progress-bar bg-warning');
          } else {
            $("#build-progress-bar").attr('class', 'progress-bar bg-success');
          }
          if (status == 100) {
            return
          }
          $.ajax({
            async: false,
            type: 'GET',
            url: "/app/build/status/"+res.job+"/"+res.id,
            success: function(now) {
              if (now.status == true && now.running == true) {
                running, ok = true;
                if (status >= 70) {
                  delay = 3000;
                } else if (status >= 80) {
                  delay = 4000;
                } else if (status >= 90) {
                  delay = 5000;
                }
              } else if (now.status == true) {
                ok = true;
                status = 99;
              } else {
                console.log("ERROR");
                $("#build-progress-bar").css({'width':'0%'});
                $("#build-progress-bar").attr('area-valuenow', '0');
                $("#build-percent").text("ERROR!");
                return;
              }
            }
          });
          status++;
          if (status <= 100) { 
            building();
          } 
        }, delay)
      }
      
      building();

      if (running == false && ok == false) {
        $("#build-progress-bar").css({'width':'0%'});
        $("#build-progress-bar").attr('area-valuenow', '0');
        $("#build-percent").text("ERROR!");
      }
    } else {
      location.href = "/html/404.html"
    }
  });
});

// deploy app
$("#deploy-app-btn").click(function(e){
  e.preventDefault();
  if ($(this).hasClass('clicked')) { 
    return false;
  } else {
    $(this).addClass('clicked').trigger('click');
  }

  $("#deploy-progress-bar").css({'width':"0%"});
  $("#deploy-progress-bar").attr('area-valuenow', '0');
  $("#deploy-percent").text("0%");
  var appName = $("h1").text();
  fetch('/app/deploy', {
    method: 'post',
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
      // 여기부터 수정
      $("#deploy-detail").attr('href', 'https://3.35.25.64:31286/applications/'+appName)
      var status = 0;
      var running = false;
      var ok = false;
      var delay = 100;

      function deploying() {         //  create a loop function
        setTimeout(function() {   //  call a 3s setTimeout when the loop is called
          running, ok = false;
          $("#deploy-progress-bar").css({'width':String(status)+"%"});
          $("#deploy-progress-bar").attr('area-valuenow', String(status));
          $("#deploy-percent").text(String(status)+"%");
          if (status <= 20) {
            $("#deploy-progress-bar").attr('class', 'progress-bar bg-danger');
          } else if (status <= 50) {
            $("#deploy-progress-bar").attr('class', 'progress-bar bg-warning');
          } else {
            $("#deploy-progress-bar").attr('class', 'progress-bar bg-success');
          }
          $.ajax({
            async: false,
            type: 'GET',
            url: "/app/deploy/status/"+appName,
            success: function(now) {
              if (now.status == true && now.running == true) {
                running, ok = true;
                if (status >= 70) {
                  delay = 1000;
                } else if (status >= 80) {
                  delay = 1500;
                } else if (status >= 90) {
                  delay = 2000;
                }
              } else if (now.status == true) {
                ok = true;
                status = 99;
              } else {
                console.log("ERROR");
                $("#deploy-progress-bar").css({'width':'0%'});
                $("#deploy-progress-bar").attr('area-valuenow', '0');
                $("#deploy-percent").text("ERROR!");
                return;
              }
            }
          });
          status++;
          if (status <= 100) { 
            deploying();
          } 
        }, delay)
      }
      
      deploying();

      if (running == false && ok == false) {
        $("#deploy-progress-bar").css({'width':'0%'});
        $("#deploy-progress-bar").attr('area-valuenow', '0');
        $("#deploy-percent").text("ERROR!");
      }
    } else {
      location.href = "/html/404.html"
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
      var count = 0;
      if (items.length == 0) {
        $("#app-list").append("<p>There is no application on project '"+parameters.name+"'.</p>");
      } else {
        items.forEach(e => {
          if (e.project == parameters.name) {
            count = count + 1;
            $("#app-list").append("<li><b>"+e.name+"</b><p>"+e.description+"</p>"+"<a rel='nofollow' href='../html/application.html?name="+e.name+"'>See Detail &rarr;</a>");
          }
        });
        if (count == 0) {
          $("#app-list").append("<p>There is no application on project '"+parameters.name+"'.</p>");
        }
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
