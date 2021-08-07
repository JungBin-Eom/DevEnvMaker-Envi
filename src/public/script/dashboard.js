$("#createproject").click(function(){
  fetch('/project', {
    method: 'post',
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
        name: $('#newprojectname-input').val(),
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

$.get("/user", function(user) {
  console.log(user)
  $("#user-id").html(user.id)
});

var projectList = $("#project-list")
var appList = $("#app-list")

var addProject = function(item) {
  projectList.append("<a class='collapse-item' href='#'>"+item.name+"</a>");
};

var addApp = function(item) {
  appList.append("<a class='collapse-item' href='#'>"+item.name+"</a>");
};

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