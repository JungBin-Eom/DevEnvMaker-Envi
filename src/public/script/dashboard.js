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

$("#user-id").click(function(){
  $.get("/user", function(user) {
    $("#user-id").html(user.id)
  });
})