$('.btn-dupcheck').click(function(){
  var id = $('.id-input');
  $.get("/signup/idcheck/"+id.val(), function(dup) {
    if (dup.duplicated == true) { // duplicated
      $('.id-input').css({'background-color':'#ffd1d1'});
    } else {
      $('.id-input').css({'background-color':'#d6ffd1'});
    }
  });
})

$('.btn-submit').click(function(){
  var pw1 = $('.pw-input1').val();
  var pw2 = $('.pw-input2').val();
  if (pw1 != pw2) { // not equal
    $('.pw-input1').css({'background-color':'#ffd1d1'});
    $('.pw-input2').css({'background-color':'#ffd1d1'});
  } else {
    $('.pw-input1').css({'background-color':'#d6ffd1'});
    $('.pw-input2').css({'background-color':'#d6ffd1'});
    fetch('/signup/register', {
      method: 'post',
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
          id: $('.id-input').val(),
          password: $('.pw-input1').val(),
          email: $('.email-input').val()
      })
    });
    alert("회원가입 완료!")
    location.href="/html/signin.html"
  }
})