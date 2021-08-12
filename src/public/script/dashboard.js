// 오른쪽 상단 사용자 이름 표시
$.get("/user", function(user) {
  $("#user-id").html(user.id)
});

// 프로젝트 생성 버튼 클릭
$("#createproject").click(function(){
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
    console.log(res)
    if (res.success == true) {
      alert("프로젝트 생성 완료!")
      location.href="/"
    } else {
      alert("입력한 정보가 정확하지 않습니다.")
    }
  });
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
    console.log(res)
    if (res.success == true) {
      $("#"+projectName).remove();
      alert("프로젝트 삭제 완료!")
      location.href="/"
    } else {
      alert("삭제 중 문제가 발생했습니다.")
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

$.get("/app", function(items) {
  if (items.length == 0) {
    appList.append("<h6 class='collapse-header'>There is no application.</h6>");
  } else {
    items.forEach(e => {
        addItem(e)
    });
  }
});

// 프로젝트, 애플리케이션마다 태그 추가
var projectList = $("#project-list")
var appList = $("#app-list")

var addProject = function(item) {
  projectList.append("<a class='collapse-item project' id='"+item.name+"' href='#none' onclick='goProjectPage(this)' >"+item.name+"</a>");
};

var addApp = function(item) {
  appList.append("<a class='collapse-item application' id='"+item.name+"' href='#none'>"+item.name+"</a>");
};

function goProjectPage(obj){
  projectName = $(obj).attr("id");
  location.href="../html/project.html?name="+projectName;
};

// redirect시 url 파라미터 쿼리 얻기
function get_query(){
  var url = document.location.href;
  var qs = url.substring(url.indexOf('?') + 1).split('&');
  for(var i = 0, result = {}; i < qs.length; i++){
    qs[i] = qs[i].split('=');
    result[qs[i][0]] = decodeURIComponent(qs[i][1]);
  }
  return result;
}

var parameters = get_query();
console.log(parameters)
if (parameters.name != undefined) {
  var itemName = parameters.name;
  $("#name").text(itemName);
  $.get("/project/"+itemName, function(project) {
    $("#project-description").text(project.description);
  });
}

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