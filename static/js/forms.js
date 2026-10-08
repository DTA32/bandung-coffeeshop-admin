// Small progressive enhancements that replace the console's React state:
// list search, the Leaflet coordinate picker, the orderable image list,
// rating / tag toggles, confirm prompts and the "Saving..." button label.
;(function () {
  'use strict'

  // Bandung city center — default view when no coordinates are set yet
  // (same as the console's CoordinatePicker.tsx).
  var DEFAULT_CENTER = [-6.9017, 107.6178]

  function pinIcon() {
    return L.divIcon({
      className: '',
      html:
        '<div style="position:relative;width:28px;height:28px;transform:translate(-50%,-50%)">' +
        '<div style="position:absolute;inset:0;border-radius:50%;background:#6A9E52AA"></div>' +
        '<div style="position:absolute;top:50%;left:50%;width:14px;height:14px;border-radius:50%;background:#4A7038;border:2px solid white;box-shadow:0 1px 4px rgba(0,0,0,.3);transform:translate(-50%,-50%)"></div>' +
        '</div>',
      iconSize: [0, 0],
      iconAnchor: [0, 0]
    })
  }

  function pointFromGeojson(text) {
    try {
      var g = JSON.parse(text)
      if (g && g.type === 'Point' && Array.isArray(g.coordinates)) {
        return { lat: g.coordinates[1], lng: g.coordinates[0] }
      }
    } catch (e) {
      // not valid / not a point
    }
    return null
  }

  function initCoordinatePicker(el) {
    if (typeof L === 'undefined') return
    var latInput = document.getElementById('coord-lat')
    var lngInput = document.getElementById('coord-lng')
    var geojson = el.dataset.geojsonTarget ? document.querySelector(el.dataset.geojsonTarget) : null

    if (geojson) {
      var p = pointFromGeojson(geojson.value)
      if (p) {
        latInput.value = p.lat
        lngInput.value = p.lng
      }
    }

    var lat = parseFloat(latInput.value)
    var lng = parseFloat(lngInput.value)
    var has = !isNaN(lat) && !isNaN(lng)

    var map = L.map(el).setView(has ? [lat, lng] : DEFAULT_CENTER, has ? 16 : 13)
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      // OSM blocks tile requests without a Referer; the page sends
      // Referrer-Policy: same-origin, so override it for tiles.
      referrerPolicy: 'no-referrer-when-downgrade'
    }).addTo(map)

    var marker = null

    function write(ll) {
      latInput.value = Number(ll.lat).toFixed(7)
      lngInput.value = Number(ll.lng).toFixed(7)
      if (geojson) {
        geojson.value = JSON.stringify({ type: 'Point', coordinates: [Number(lngInput.value), Number(latInput.value)] })
      }
    }

    function place(ll) {
      if (marker) {
        marker.setLatLng(ll)
        return
      }
      marker = L.marker(ll, { icon: pinIcon(), draggable: true }).addTo(map)
      marker.on('dragend', function () {
        write(marker.getLatLng())
      })
    }

    if (has) place([lat, lng])

    map.on('click', function (e) {
      write(e.latlng)
      place(e.latlng)
    })

    function onInput() {
      var a = parseFloat(latInput.value)
      var b = parseFloat(lngInput.value)
      if (isNaN(a) || isNaN(b)) {
        if (marker) {
          map.removeLayer(marker)
          marker = null
        }
        return
      }
      if (geojson) write({ lat: a, lng: b })
      place([a, b])
      map.panTo([a, b])
    }
    latInput.addEventListener('change', onInput)
    lngInput.addEventListener('change', onInput)
    if (geojson) {
      geojson.addEventListener('change', function () {
        var p = pointFromGeojson(geojson.value)
        if (!p) return
        latInput.value = p.lat
        lngInput.value = p.lng
        place([p.lat, p.lng])
        map.panTo([p.lat, p.lng])
      })
    }

    // The picker may start hidden (location type != poi); fix its size when shown.
    el._leafletMap = map
  }

  // Location form: the map is only for points; polygons are pasted as GeoJSON.
  function initLocationType(select) {
    var form = select.form
    function sync() {
      var poi = select.value === 'poi'
      form.querySelectorAll('[data-poi-only]').forEach(function (n) {
        n.hidden = !poi
        if (poi) {
          var m = n.querySelector('[data-coordinate-picker]')
          if (m && m._leafletMap) m._leafletMap.invalidateSize()
        }
      })
      form.querySelectorAll('[data-polygon-only]').forEach(function (n) {
        n.hidden = poi
      })
    }
    select.addEventListener('change', sync)
  }

  function initImagesEditor(root) {
    var list = root.querySelector('[data-image-list]')
    var tpl = root.querySelector('[data-image-template]')

    root.querySelector('[data-image-add]').addEventListener('click', function () {
      var row = tpl.content.firstElementChild.cloneNode(true)
      list.appendChild(row)
      var input = row.querySelector('input[name="image_url"]')
      if (input) input.focus()
    })

    list.addEventListener('click', function (e) {
      var btn = e.target.closest('button')
      if (!btn) return
      var row = btn.closest('[data-image-row]')
      if (btn.hasAttribute('data-image-remove')) row.remove()
      else if (btn.hasAttribute('data-image-up') && row.previousElementSibling)
        list.insertBefore(row, row.previousElementSibling)
      else if (btn.hasAttribute('data-image-down') && row.nextElementSibling)
        list.insertBefore(row.nextElementSibling, row)
    })
  }

  // Ratings: the score/override inputs are disabled until the dimension is ticked.
  function initRating(box) {
    var toggle = box.querySelector('[data-rating-toggle]')
    toggle.addEventListener('change', function () {
      box.querySelectorAll('input:not([data-rating-toggle])').forEach(function (i) {
        i.disabled = !toggle.checked
      })
    })
  }

  // Tags: "visible" only shows while the tag is assigned (rows can be added by HTMX).
  document.addEventListener('change', function (e) {
    var assign = e.target.closest('[data-tag-assign]')
    if (!assign) return
    var row = assign.closest('[data-tag-row]')
    var vis = row.querySelector('[data-tag-visible]')
    vis.hidden = !assign.checked
    if (assign.checked) vis.querySelector('input').checked = true
  })

  // Inline "Add tag": clear the inputs and the error after a successful create.
  document.addEventListener('htmx:afterRequest', function (e) {
    var box = e.target.closest && e.target.closest('[data-new-tag]')
    if (!box || !e.detail.successful) return
    if (e.detail.xhr.getResponseHeader('HX-Retarget')) return
    box.querySelectorAll('input').forEach(function (i) {
      i.value = ''
    })
    var err = document.getElementById('new-tag-error')
    if (err) err.textContent = ''
  })

  // List search (cafes, locations): filter rendered rows by name / id.
  function initListFilter(input) {
    var list = document.querySelector(input.dataset.listFilter)
    var empty = list.querySelector('[data-list-empty]')
    function run() {
      var q = input.value.trim().toLowerCase()
      var shown = 0
      list.querySelectorAll('[data-search]').forEach(function (row) {
        var hit = row.dataset.search.indexOf(q) !== -1
        row.hidden = !hit
        if (hit) shown++
      })
      if (empty) empty.hidden = shown > 0
    }
    input.addEventListener('input', run)
    run()
  }

  // Confirm prompts for destructive buttons.
  document.addEventListener(
    'click',
    function (e) {
      var btn = e.target.closest('[data-confirm]')
      if (btn && !window.confirm(btn.dataset.confirm)) {
        e.preventDefault()
        e.stopImmediatePropagation()
      }
    },
    true
  )

  // "Saving..." label + disabled buttons while a full-page form posts.
  document.addEventListener('submit', function (e) {
    var form = e.target
    if (form.hasAttribute('hx-post')) return
    var label = form.querySelector('[data-saving-label]')
    if (label && (!e.submitter || e.submitter === label)) label.textContent = label.dataset.savingLabel
    setTimeout(function () {
      form.querySelectorAll('button').forEach(function (b) {
        b.disabled = true
      })
    }, 0)
  })

  // Re-enable buttons when the page comes back from the bfcache.
  window.addEventListener('pageshow', function (e) {
    if (!e.persisted) return
    document.querySelectorAll('form button[disabled]').forEach(function (b) {
      if (!b.closest('[data-rating]')) b.disabled = false
    })
  })

  document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('[data-coordinate-picker]').forEach(initCoordinatePicker)
    document.querySelectorAll('[data-location-type]').forEach(initLocationType)
    document.querySelectorAll('[data-images-editor]').forEach(initImagesEditor)
    document.querySelectorAll('[data-rating]').forEach(initRating)
    document.querySelectorAll('[data-list-filter]').forEach(initListFilter)
  })
})()
