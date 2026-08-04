document.addEventListener('DOMContentLoaded', function() {
  document.querySelectorAll('.sidebar .dir-label').forEach(function(el) {
    el.addEventListener('click', function() {
      var ul = this.nextElementSibling;
      if (ul) {
        ul.style.display = ul.style.display === 'block' ? 'none' : 'block';
      }
    });
  });
});