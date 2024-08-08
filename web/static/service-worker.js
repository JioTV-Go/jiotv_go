const CACHE_NAME = 'jitovgo-cache';
const urlsToCache = [
  '/',
  '/views/index.html',
  '/static/index.js',
  '/static/images/logo/32.png',
  '/static/images/logo/64.png',
  '/static/images/logo/72.png',
  '/static/images/logo/128.png',
  '/static/images/logo/144.png',
  '/static/images/logo/180.png',
  '/static/images/logo/256.png',
  '/static/images/logo/512.png',
  '/static/images/logo/1024.png',
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        return cache.addAll(urlsToCache);
      })
  );
});

self.addEventListener('fetch', event => {
  const url = new URL(event.request.url);
  
  // Check if the request is for an image in the /static/images folder
  if (url.pathname.startsWith('/static/')) {
    event.respondWith(
      caches.match(event.request).then(response => {
        return response || fetch(event.request).then(fetchedResponse => {
          return caches.open(CACHE_NAME).then(cache => {
            cache.put(event.request, fetchedResponse.clone());
            return fetchedResponse;
          });
        });
      })
    );
  } else {
    event.respondWith(
      caches.match(event.request).then(response => {
        return response || fetch(event.request);
      })
    );
  }
});


self.addEventListener('activate', (event) => {
  const cacheWhitelist = [CACHE_NAME];
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (!cacheWhitelist.includes(cacheName)) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});
