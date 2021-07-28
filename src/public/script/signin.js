$('.btn-submit').click(function(){
  fetch('/signin', {
    method: 'post',
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
        id: $('.id-input').val(),
        password: $('.pw-input1').val(),
      })
  });
  location.href="/html/signin.html"
})