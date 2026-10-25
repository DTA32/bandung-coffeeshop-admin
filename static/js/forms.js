// Small progressive enhancements that replace the console's React state:
// the Leaflet coordinate picker and geoman polygon editor, the orderable image list,
// the per-day hours editor, rating / tag toggles, /changes batch selection, confirm prompts and the "Saving..." button label.
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
    var errorBox = root.querySelector('[data-image-error]')
    var form = root.closest('form')

    function newRow() {
      var row = tpl.content.firstElementChild.cloneNode(true)
      list.appendChild(row)
      return row
    }

    // The preview follows the URL input (typed or uploaded).
    function syncPreview(row) {
      var img = row.querySelector('[data-image-preview]')
      var url = row.querySelector('input[name="image_url"]').value.trim()
      if (!img) return
      img.hidden = !/^https?:\/\//.test(url)
      if (img.hidden) return
      // Same small variant as the server-rendered thumb(u, 160).
      var base = root.dataset.imageBase + '/'
      img.src = url.indexOf(base) === 0 ? url + (url.indexOf('?') < 0 ? '?' : '&') + 'w=160' : url
    }

    function setStatus(row, text) {
      var st = row.querySelector('[data-image-status]')
      st.textContent = text
      st.hidden = !text
    }

    function showError(text) {
      errorBox.textContent = text
      errorBox.hidden = !text
    }

    function slug() {
      var input = form && form.querySelector('[data-image-slug]')
      return input ? input.value.trim() : ''
    }

    // Uploads one file into row; the URL input gets the CDN URL on success.
    function upload(file, row) {
      var data = new FormData()
      data.append('folder', root.dataset.imageFolder)
      data.append('slug', slug())
      data.append('file', file)
      row.dataset.uploading = '1'
      setStatus(row, 'Uploading ' + file.name + '…')
      return fetch('/images', { method: 'POST', body: data, credentials: 'same-origin' })
        .then(function (res) {
          return res.json().then(
            function (body) { return { ok: res.ok, body: body } },
            function () { return { ok: false, body: { error: 'upload failed (' + res.status + ')' } } }
          )
        })
        .then(function (r) {
          if (!r.ok) throw new Error(r.body.error || 'upload failed')
          row.querySelector('input[name="image_url"]').value = r.body.url
          syncPreview(row)
          setStatus(row, '')
        })
        .catch(function (err) {
          setStatus(row, file.name + ': ' + err.message)
        })
        .finally(function () {
          delete row.dataset.uploading
        })
    }

    root.querySelector('[data-image-add]').addEventListener('click', function () {
      var input = newRow().querySelector('input[name="image_url"]')
      if (input) input.focus()
    })

    // File inputs sit inside labels (templates.ImagesEditor): "Upload images"
    // ([data-image-file], several files, a new row each) and each row's own
    // ([data-image-row-file], replaces that row's URL). Delegated, so rows
    // added later work too.
    root.addEventListener('change', function (e) {
      var input = e.target
      var rowInput = input.hasAttribute('data-image-row-file')
      if (!rowInput && !input.hasAttribute('data-image-file')) return
      var files = Array.prototype.slice.call(input.files)
      input.value = ''
      showError('')
      if (!files.length) return
      if (!slug()) {
        showError('Set the ID / slug before uploading images.')
        return
      }
      var target = rowInput ? input.closest('[data-image-row]') : null
      // One at a time keeps the -2, -3 suffixes of same-named files in order.
      files.reduce(function (p, file, i) {
        var row = target && i === 0 ? target : newRow()
        return p.then(function () { return upload(file, row) })
      }, Promise.resolve())
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

    list.addEventListener('change', function (e) {
      if (e.target.name === 'image_url') syncPreview(e.target.closest('[data-image-row]'))
    })

    // Don't save the form while an upload is still running.
    if (form) {
      form.addEventListener('submit', function (e) {
        if (list.querySelector('[data-uploading]')) {
          e.preventDefault()
          e.stopImmediatePropagation()
          showError('Wait for the uploads to finish before saving.')
        }
      }, true)
    }
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

  // Batch select on /changes: row checkboxes (also rows appended later by
  // infinite scroll), "select all loaded", sticky action bar. Delegated, so
  // new rows start unchecked and everything re-syncs after any htmx swap.
  function syncBatch() {
    var bar = document.getElementById('batch-bar')
    if (!bar) return
    document.querySelectorAll('[data-change-group]').forEach(function (g) {
      if (!g.querySelector('[id^="proposal-"]')) g.remove()
    })
    var rows = document.querySelectorAll('[data-batch-row]')
    var n = 0
    rows.forEach(function (c) {
      if (c.checked) n++
    })
    var all = document.getElementById('batch-all')
    if (all) {
      all.checked = rows.length > 0 && n === rows.length
      all.indeterminate = n > 0 && n < rows.length
    }
    bar.hidden = n === 0
    var label = n + ' selected'
    bar.querySelector('[data-batch-count]').textContent = label
    var a = bar.querySelector('[data-batch-approve]')
    var r = bar.querySelector('[data-batch-reject]')
    a.dataset.confirm = 'Approve ' + n + ' selected change(s) as proposed? Informational ones are skipped.'
    r.dataset.confirm = 'Reject ' + n + ' selected change(s)? The scraper will not re-propose them.'
  }
  document.addEventListener('change', function (e) {
    var t = e.target
    if (t.id === 'batch-all') {
      document.querySelectorAll('[data-batch-row]').forEach(function (c) {
        c.checked = t.checked
      })
      syncBatch()
    } else if (t.matches && t.matches('[data-batch-row]')) {
      syncBatch()
    }
  })
  document.addEventListener('click', function (e) {
    if (e.target.closest && e.target.closest('[data-batch-clear]')) {
      document.querySelectorAll('[data-batch-row]').forEach(function (c) {
        c.checked = false
      })
      syncBatch()
    }
  })
  ;['htmx:afterSwap', 'htmx:oobAfterSwap', 'htmx:afterSettle'].forEach(function (ev) {
    document.body.addEventListener(ev, syncBatch)
  })
  window.addEventListener('pageshow', syncBatch)

  // Per-day hours editor (templates/hours.templ): add / remove time ranges and
  // special dates, and "Closed" disables the group's ranges (disabled inputs are
  // not posted, so hours_group / hours_open / hours_close stay aligned).
  // Delegated, so editors inside HTMX-swapped proposal rows work too.
  var hoursSeq = 0
  function syncHoursClosed(group) {
    var box = group.querySelector('[data-hours-closed]')
    if (!box) return
    group.querySelectorAll('[data-hours-range] input, [data-hours-range] button, [data-hours-add-range]').forEach(function (i) {
      if (i.closest('[data-hours-group]') === group) i.disabled = box.checked
    })
  }
  function syncAllHours() {
    document.querySelectorAll('[data-hours-group]').forEach(syncHoursClosed)
  }
  document.addEventListener('click', function (e) {
    var btn = e.target.closest && e.target.closest('button')
    if (!btn) return
    var editor = btn.closest('[data-hours-editor]')
    if (!editor) return
    if (btn.hasAttribute('data-hours-add-range')) {
      var group = btn.closest('[data-hours-group]')
      var row = editor.querySelector('[data-hours-range-template]').content.firstElementChild.cloneNode(true)
      row.querySelector('input[name="hours_group"]').value = group.dataset.hoursGroup
      group.querySelector('[data-hours-ranges]').appendChild(row)
      row.querySelector('input[name="hours_open"]').focus()
    } else if (btn.hasAttribute('data-hours-remove-range')) {
      btn.closest('[data-hours-range]').remove()
    } else if (btn.hasAttribute('data-hours-add-special')) {
      var key = 'n' + Date.now().toString(36) + '-' + ++hoursSeq
      var special = editor.querySelector('[data-hours-special-template]').content.firstElementChild.cloneNode(true)
      special.dataset.hoursGroup = key
      special.querySelectorAll('input[name="special_key"], input[name="hours_group"], input[name="hours_closed"]').forEach(function (i) {
        i.value = key
      })
      editor.querySelector('[data-hours-specials]').appendChild(special)
      special.querySelector('input[name="special_from"]').focus()
    } else if (btn.hasAttribute('data-hours-remove-special')) {
      btn.closest('[data-hours-special]').remove()
    }
  })
  document.addEventListener('change', function (e) {
    if (!(e.target.matches && e.target.matches('[data-hours-closed]'))) return
    syncHoursClosed(e.target.closest('[data-hours-group]'))
  })
  document.body.addEventListener('htmx:afterSettle', syncAllHours)
  window.addEventListener('pageshow', syncAllHours)

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
    syncAllHours()
    syncBatch()
  })
})()
