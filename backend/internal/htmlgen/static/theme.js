// AI Daily — theme.js
(function(){
  // 恢复主题
  var t = localStorage.getItem('theme');
  if (t === 'light') document.documentElement.classList.add('light');

  // 主题切换
  window.toggleTheme = function(){
    document.documentElement.classList.toggle('light');
    localStorage.setItem('theme',
      document.documentElement.classList.contains('light') ? 'light' : 'dark');
  };

  // 图片加载失败时显示占位
  document.addEventListener('error', function(e){
    if (e.target.tagName === 'IMG' && e.target.dataset.fallback !== '1') {
      e.target.dataset.fallback = '1';
      e.target.style.display = 'none';
      var fb = e.target.nextElementSibling;
      if (fb && fb.classList.contains('fallback')) fb.style.display = 'flex';
    }
  }, true);
})();
