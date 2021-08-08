$.get("/user", function(user) {
  $("#user-id").html(user.id)
});


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

var projectList = $("#project-list")
var appList = $("#app-list")

var addProject = function(item) {
  projectList.append("<a class='collapse-item project' id='"+item.name+"' href='#none'>"+item.name+"</a>");
};

var addApp = function(item) {
  appList.append("<a class='collapse-item application' id='"+item.name+"' href='#none'>"+item.name+"</a>");
};

$("#test").click(function(){
  projectName = $(this).attr("id");
  console.log(projectName);
  $("#project-name").text(projectName);
  console.log("clicked");
  alert("clicked");
  location.href="../html/project.html"
});


var projectName="ricky";
$("#project-name").text(projectName);

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