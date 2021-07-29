$('.btn-submit').click(function(){
  fetch('/signin', {
    method: 'post',
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
        id: $('.id-input').val(),
        password: $('.pw-input').val(),
      })
  })
  .then(res => res.json())
  .then(res => {
    console.log(res)
    if (res.success == true) {
      location.href="/"
    } else {
      alert("입력한 정보가 정확하지 않습니다.")
    }
  });
})