$('.btn-dupcheck').click(function(){
  var id = $('.id-input');
  $.get("/iddup/"+id.val(), function(dup) {
    console.log(dup)
    if (dup.Duplicated == true) {
      $('.id-input').css({'background-color':'#ffd1d1'});
    } else {
      $('.id-input').css({'background-color':'#d6ffd1'});
    }
  });
})

$('.btn-submit').click(function(){
  var pw1 = $('.pw-input1');
  var pw2 = $('.pw-input2');
  if (pw1 != pw2) {
    $('.pw-input1').css({'background-color':'#ffd1d1'});
    $('.pw-input2').css({'background-color':'#ffd1d1'});
    return;
  } else {
    $('.pw-input1').css({'background-color':'#d6ffd1'});
    $('.pw-input2').css({'background-color':'#d6ffd1'});
  }
  
})