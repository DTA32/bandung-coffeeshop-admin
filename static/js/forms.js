// Small progressive enhancements that replace the console's React state:
// list search, the Leaflet coordinate picker and geoman polygon editor, the orderable image list,
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

  function osmTiles() {
    return L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      // OSM blocks tile requests without a Referer; the page sends
      // Referrer-Policy: same-origin, so override it for tiles.
      referrerPolicy: 'no-referrer-when-downgrade'
    })
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
    osmTiles().addTo(map)

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

  // Polygon editor (location form, types area / district): leaflet-geoman edits
  // the layers and mirrors them into the GeoJSON textarea; editing the textarea
  // and blurring it redraws the map.
  function initPolygonEditor(el) {
    if (typeof L === 'undefined' || !L.PM) return
    var geojson = document.querySelector(el.dataset.geojsonTarget)
    var errBox = el.parentNode.querySelector('[data-geojson-error]')
    if (!geojson) return

    var map = L.map(el).setView(DEFAULT_CENTER, 13)
    osmTiles().addTo(map)
    var group = L.featureGroup().addTo(map)

    map.pm.addControls({
      position: 'topleft',
      drawMarker: false,
      drawCircleMarker: false,
      drawPolyline: false,
      drawRectangle: false,
      drawCircle: false,
      drawText: false,
      drawPolygon: true,
      editMode: true,
      dragMode: true,
      cutPolygon: true,
      removalMode: true,
      rotateMode: false
    })
    map.pm.setGlobalOptions({ allowSelfIntersection: false })

    function setError(msg) {
      if (!errBox) return
      errBox.textContent = msg || ''
      errBox.hidden = !msg
    }

    function liveLayers() {
      return group.getLayers().filter(function (l) {
        return map.hasLayer(l)
      })
    }

    // Textarea <- map
    function write() {
      var polys = []
      liveLayers().forEach(function (l) {
        var g = l.toGeoJSON(7).geometry
        if (g.type === 'Polygon') polys.push(g.coordinates)
        else if (g.type === 'MultiPolygon') g.coordinates.forEach(function (c) { polys.push(c) })
      })
      if (polys.length === 0) geojson.value = ''
      else if (polys.length === 1) geojson.value = JSON.stringify({ type: 'Polygon', coordinates: polys[0] })
      else geojson.value = JSON.stringify({ type: 'MultiPolygon', coordinates: polys })
      setError('')
    }

    function track(layer) {
      group.addLayer(layer)
      layer.on('pm:edit pm:dragend pm:markerdragend pm:vertexadded pm:vertexremoved pm:cut', write)
    }

    function fit() {
      map.invalidateSize()
      if (liveLayers().length) map.fitBounds(group.getBounds(), { padding: [20, 20], maxZoom: 17 })
    }

    // Map <- textarea. Returns false (map untouched) when the text is not a
    // Polygon / MultiPolygon.
    function load(quiet) {
      var text = geojson.value.trim()
      var polys = []
      if (text) {
        var g
        try {
          g = JSON.parse(text)
        } catch (e) {
          if (!quiet) setError('Invalid JSON: ' + e.message)
          return false
        }
        if (g && g.type === 'Polygon' && Array.isArray(g.coordinates)) polys = [g.coordinates]
        else if (g && g.type === 'MultiPolygon' && Array.isArray(g.coordinates)) polys = g.coordinates
        else {
          if (!quiet) setError('Expected a Polygon or MultiPolygon geometry.')
          return false
        }
        try {
          polys = polys.map(function (c) {
            return L.polygon(L.GeoJSON.coordsToLatLngs(c, 1))
          })
        } catch (e) {
          if (!quiet) setError('Invalid coordinates.')
          return false
        }
      }
      setError('')
      group.clearLayers()
      polys.forEach(track)
      if (!el.closest('[hidden]')) fit()
      return true
    }

    map.on('pm:create', function (e) {
      track(e.layer)
      write()
    })
    map.on('pm:remove', function (e) {
      group.removeLayer(e.layer)
      write()
    })
    map.on('pm:cut', function (e) {
      if (!group.hasLayer(e.layer)) track(e.layer)
      if (e.originalLayer && e.originalLayer !== e.layer) group.removeLayer(e.originalLayer)
      write()
    })

    geojson.addEventListener('change', function () {
      load(false)
    })

    el._polygonEditor = {
      // Called when the editor becomes visible or the type changes.
      show: function () {
        load(true)
        fit()
      }
    }
    load(true)
  }

  // Location form: poi uses the point picker, area / district the polygon editor.
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
        if (!poi) {
          var m = n.querySelector('[data-polygon-editor]')
          if (m && m._polygonEditor) m._polygonEditor.show()
        }
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

  // Touch button area: only while the tag is assigned and visible in the form.
  function syncTouchBox(row) {
    var box = row.querySelector('[data-tag-touch-box]')
    if (!box) return
    var assigned = row.querySelector('[data-tag-assign]').checked
    box.hidden = !(assigned && row.querySelector('[data-tag-visible] input').checked)
  }
  document.addEventListener('change', function (e) {
    var row = e.target.closest && e.target.closest('[data-tag-row]')
    if (row) syncTouchBox(row)
  })

  // After a touch, that tag is the newest: drop the "current remark" badge elsewhere.
  document.addEventListener('htmx:afterSwap', function (e) {
    var el = e.detail.elt
    var mine = document.querySelector('[data-tag-touch][id="' + (e.detail.target && e.detail.target.id) + '"]')
    var box = mine || (el && el.closest && el.closest('[data-tag-touch]'))
    if (!box) return
    document.querySelectorAll('[data-tag-touch] [data-current-remark]').forEach(function (b) {
      if (!box.contains(b)) b.remove()
    })
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

  // Selects with data-autosubmit submit their GET form on change (cafe list sort).
  document.addEventListener('change', function (e) {
    var sel = e.target.closest && e.target.closest('[data-autosubmit]')
    if (sel && sel.form) sel.form.requestSubmit()
  })

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
    document.querySelectorAll('[data-polygon-editor]').forEach(initPolygonEditor)
    document.querySelectorAll('[data-location-type]').forEach(initLocationType)
    document.querySelectorAll('[data-images-editor]').forEach(initImagesEditor)
    document.querySelectorAll('[data-rating]').forEach(initRating)
    document.querySelectorAll('[data-list-filter]').forEach(initListFilter)
  })
})()
