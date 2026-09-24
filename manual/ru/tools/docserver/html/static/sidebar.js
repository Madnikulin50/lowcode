document.addEventListener('DOMContentLoaded', function () {
  document.querySelectorAll('.sidebar .dir-label').forEach(function (el) {
    el.addEventListener('click', function () {
      var li = this.parentElement
      if (li) li.classList.toggle('expanded')
    })
  })

  var menu = document.getElementById('docs-menu')
  var backdrop = document.getElementById('docs-backdrop')

  function toggleNav () {
    if (window.innerWidth < 1024) {
      document.body.classList.toggle('nav-open')
    } else {
      document.body.classList.toggle('nav-collapsed')
    }
  }

  if (menu) menu.addEventListener('click', toggleNav)
  if (backdrop) {
    backdrop.addEventListener('click', function () {
      document.body.classList.remove('nav-open')
    })
  }

  var root = document.documentElement
  var saved = localStorage.getItem('docs-color-mode')
  if (saved === 'dark' || saved === 'light') {
    root.setAttribute('data-color-mode', saved)
  }

  var themeBtn = document.getElementById('docs-theme')
  if (themeBtn) {
    themeBtn.addEventListener('click', function () {
      var next = root.getAttribute('data-color-mode') === 'dark' ? 'light' : 'dark'
      root.setAttribute('data-color-mode', next)
      localStorage.setItem('docs-color-mode', next)
    })
  }
})
