var ws;

const MSG_KEY_VALUE_SEP = "=";
const SHAPE_KEY_VALUE_SEP = ":";
const SHAPE_ENTRY_SEP = ";";
const TEXT_MIN_SIZE = 8;

// ==================================================
// FONCTIONS D'ENCODAGE/DECODAGE DES MESSAGES
// ==================================================

// Encode un objet en message protocolaire sous la forme ";:cle:valeur" concatenee.
// Exemple: TODO
// Paramètres : TODO
// Retourne : TODO
function encode(obj) {
  return Object.entries(obj)
    .map(
      ([key, value]) =>
        `${SHAPE_ENTRY_SEP}${SHAPE_KEY_VALUE_SEP}${key}${SHAPE_KEY_VALUE_SEP}${value}`,
    )
    .join("");
}

// Decode un message protocolaire sous la forme ";:cle:valeur" concatenee en un objet.
// Exemple: TODO
// Paramètres : TODO
// Retourne : TODO
function decode(str) {
  return Object.fromEntries(
    str
      .trim()
      .split(SHAPE_ENTRY_SEP + SHAPE_KEY_VALUE_SEP)
      .filter(Boolean) // pour enlever les éléments vides résultant du split
      .map((pair) => {
        const [key, value] = pair
          .split(SHAPE_KEY_VALUE_SEP)
          .map((s) => s.trim());
        return [key, value];
      }),
  );
}

// ==================================================
// WHITE BOARD DATA
// ==================================================
let shapes = {}; // { id: shapeObject }
let selectedId = null;
let tool = "select";

// Etats d'interaction
let isMouseDown = false;
let mouseStart = { x: 0, y: 0 };
let lastMousePos = { x: 0, y: 0 };
let drawState = null; // while creating a shape by drag
let pendingState = null; // lorsqu'on a modifié le whiteboard et qu'on attend la confirmation du serveur
let dragState = null; // while moving / resizing

// ==================================================
// MISE EN PLACE DU WHITE BOARD
// ==================================================
const HANDLE_SIZE = 7; // taille des poignées de redimensionnement
const GRID_CELL_SIZE = 20;

const canvas = document.getElementById("canvas");
const ctx = canvas.getContext("2d");
const wrap = document.getElementById("canvas-wrap");

function resize() {
  canvas.width = wrap.clientWidth;
  canvas.height = wrap.clientHeight;
  redraw();
}
window.addEventListener("resize", resize);
resize();

// ==================================================
// DESSIN SUR LE WHITE BOARD
// ==================================================

// Dessine la grille de fond du canvas, avec des lignes fines tous les GRID_CELL_SIZE pixels et des lignes plus épaisses tous les 5 * GRID_CELL_SIZE pixels.
// Paramètres : aucun
// Retourne : aucun, mais dessine la grille directement sur le canvas.
function drawGrid() {
  ctx.lineWidth = 1;

  // Dessiner les lignes de la grille fine
  ctx.strokeStyle = "#19191950";
  ctx.beginPath();
  for (let x = 0; x < canvas.width; x += GRID_CELL_SIZE) {
    ctx.moveTo(x, 0);
    ctx.lineTo(x, canvas.height);
  }
  for (let y = 0; y < canvas.height; y += GRID_CELL_SIZE) {
    ctx.moveTo(0, y);
    ctx.lineTo(canvas.width, y);
  }
  ctx.stroke();

  // Dessiner les lignes de la grille épaisse tous les 5 cellules
  ctx.strokeStyle = "#1f1f1f50";
  ctx.beginPath();
  for (let x = 0; x < canvas.width; x += GRID_CELL_SIZE * 5) {
    ctx.moveTo(x, 0);
    ctx.lineTo(x, canvas.height);
  }
  for (let y = 0; y < canvas.height; y += GRID_CELL_SIZE * 5) {
    ctx.moveTo(0, y);
    ctx.lineTo(canvas.width, y);
  }
  ctx.stroke();
}

// Dessine une forme donnée sur le canvas, en appliquant les styles appropriés.
// Paramètres : TODO
// Retourne : TODO
function drawShape(s, selected) {
  ctx.save();
  if (s.cmd === "rect") {
    ctx.fillStyle = s.color || "#4488ff";
    ctx.fillRect(+s.x, +s.y, +s.w, +s.h);
    if (selected) {
      drawSelectionRect(+s.x, +s.y, +s.w, +s.h);
      drawHandlesRect(+s.x, +s.y, +s.w, +s.h);
    }
  } else if (s.cmd === "circle") {
    ctx.fillStyle = s.color || "#ff8844";
    ctx.beginPath();
    ctx.arc(+s.x, +s.y, Math.abs(+s.r), 0, Math.PI * 2);
    ctx.fill();
    if (selected) {
      ctx.strokeStyle = "#00ff88";
      ctx.lineWidth = 1;
      ctx.setLineDash([4, 3]);
      ctx.beginPath();
      ctx.arc(+s.x, +s.y, Math.abs(+s.r) + 1, 0, Math.PI * 2);
      ctx.stroke();
      ctx.setLineDash([]);
      drawHandlesCircle(+s.x, +s.y, +s.r);
    }
  } else if (s.cmd === "text") {
    const size = Math.max(+(s.size || 20), TEXT_MIN_SIZE);
    ctx.fillStyle = s.color || "#000000";
    ctx.font = `${size}px 'Courier New', monospace`;
    ctx.fillText(s.value || "", +s.x, +s.y);
    if (selected) {
      const w = ctx.measureText(s.value || "").width;
      drawSelectionRect(+s.x, +s.y - size, w, size);
      drawHandlesRect(+s.x, +s.y - size, w, size);
    }
  }
  ctx.restore();
}

// Dessine un rectangle de sélection autour d'une forme sélectionnée, avec une bordure en pointillés verts.
// Paramètres : TODO
// Retourne : TODO
function drawSelectionRect(x, y, w, h) {
  ctx.strokeStyle = "#00ff88";
  ctx.lineWidth = 1;
  ctx.setLineDash([4, 3]);
  ctx.strokeRect(x - 1, y - 1, w + 2, h + 2);
  ctx.setLineDash([]);
}

// Dessine les poignées de redimensionnement pour un rectangle.
// Paramètres : TODO
// Retourne : TODO
function drawHandlesRect(x, y, w, h) {
  const pts = getHandlePointsRect(x, y, w, h);
  for (const [hx, hy] of pts) {
    ctx.fillStyle = "#00ff88";
    ctx.fillRect(
      hx - HANDLE_SIZE / 2,
      hy - HANDLE_SIZE / 2,
      HANDLE_SIZE,
      HANDLE_SIZE,
    );
    ctx.strokeStyle = "#001a0d";
    ctx.lineWidth = 1;
    ctx.strokeRect(
      hx - HANDLE_SIZE / 2,
      hy - HANDLE_SIZE / 2,
      HANDLE_SIZE,
      HANDLE_SIZE,
    );
  }
}

// Dessine les poignées de redimensionnement pour un cercle.
// Paramètres : TODO
// Retourne : TODO
function drawHandlesCircle(cx, cy, r) {
  const pts = getHandlePointsCircle(cx, cy, r);
  for (const [hx, hy] of pts) {
    ctx.fillStyle = "#00ff88";
    ctx.fillRect(
      hx - HANDLE_SIZE / 2,
      hy - HANDLE_SIZE / 2,
      HANDLE_SIZE,
      HANDLE_SIZE,
    );
    ctx.strokeStyle = "#001a0d";
    ctx.lineWidth = 1;
    ctx.strokeRect(
      hx - HANDLE_SIZE / 2,
      hy - HANDLE_SIZE / 2,
      HANDLE_SIZE,
      HANDLE_SIZE,
    );
  }
}

// Calcule les positions des poignées de redimensionnement pour un rectangle, en fonction de ses coordonnées et dimensions.
// 8 positions de poignées pour un rectangle (T = top, M = middle, B = bottom, L = left, C = center, R = right) :
// TL TC TR ML MR BL BC BR.
// Paramètres : TODO
// Retourne : TOD
function getHandlePointsRect(x, y, w, h) {
  return [
    [x, y],
    [x + w / 2, y],
    [x + w, y],
    [x, y + h / 2],
    [x + w, y + h / 2],
    [x, y + h],
    [x + w / 2, y + h],
    [x + w, y + h],
  ];
}

// Calcule les positions des poignées de redimensionnement pour un cercle, en fonction de son centre et de son rayon.
// Paramètres : TODO
// Retourne : TODO
function getHandlePointsCircle(cx, cy, r) {
  return [
    [cx, cy - r], // Nord
    [cx + r, cy], // Est
    [cx, cy + r], // Sud
    [cx - r, cy], // Ouest
  ];
}

// Redessine tout le canvas en effaçant d'abord le contenu, puis en dessinant la grille de fond, les formes non sélectionnées, la forme sélectionnée (si elle existe) ou la forme en cours de dessin (si applicable) avec une opacité réduite.
// Paramètres : aucun
// Retourne : TODO
function redraw() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  drawGrid();

  // All non-selected and non-pending shapes first
  for (const s of Object.values(shapes)) {
    if (s.id !== selectedId) drawShape(s, false);
  }
  // Selected shape on top
  if (selectedId && shapes[selectedId]) {
    if (dragState) {
      ctx.globalAlpha = 0.33;
      drawShape(dragState.editedShape, true);
      ctx.globalAlpha = 1;
    } else if (!pendingState || selectedId !== pendingState.previewShape.id) {
      drawShape(shapes[selectedId], true);
    }
  }

  // Preview while drawing
  if (drawState?.previewShape) {
    ctx.globalAlpha = 0.33;
    drawShape(drawState.previewShape, false);
    ctx.globalAlpha = 1;
  }

  if (pendingState) {
    ctx.globalAlpha = 0.67;
    drawShape(pendingState.previewShape, pendingState.selected);
    ctx.globalAlpha = 1;
  }
}

// ==================================================
// TESTS DE COLLISION
// ==================================================

// Effectue un test de collision pour déterminer si les coordonnées (x, y) se trouvent à l'intérieur d'une forme. Parcourt les formes dans l'ordre inverse de leur création (du plus récent au plus ancien) pour détecter la forme la plus haute sous le curseur.
// Paramètres : TODO
// Retourne : TODO
function hitTest(x, y) {
  const ids = Object.keys(shapes).reverse();
  for (const id of ids) {
    const s = shapes[id];
    if (s.cmd === "rect") {
      if (x >= +s.x && x <= +s.x + +s.w && y >= +s.y && y <= +s.y + +s.h)
        return id;
    } else if (s.cmd === "circle") {
      const dx = x - +s.x,
        dy = y - +s.y;
      if (Math.sqrt(dx * dx + dy * dy) <= Math.abs(+s.r)) return id;
    } else if (s.cmd === "text") {
      ctx.font = `${+(s.size || 20)}px 'Courier New', monospace`;
      const w = ctx.measureText(s.value || "").width;
      const h = +(s.size || 20);
      if (x >= +s.x && x <= +s.x + w && y >= +s.y - h && y <= +s.y) return id;
    }
  }
  return null;
}

// Effectue un test de collision pour déterminer si les coordonnées (x, y) se trouvent à proximité d'une poignée de redimensionnement d'une forme donnée. Retourne l'indice de la poignée touchée (0 à 7 pour un rectangle, 0 à 3 pour un cercle) ou -1 si aucune poignée n'est touchée.
// Paramètres : TODO
// Retourne : TODO
function hitHandle(x, y, shape) {
  if (!shape) return -1;
  const hs = HANDLE_SIZE / 2 + 2;
  if (shape.cmd === "rect") {
    const pts = getHandlePointsRect(+shape.x, +shape.y, +shape.w, +shape.h);
    for (let i = 0; i < pts.length; i++) {
      if (Math.abs(x - pts[i][0]) <= hs && Math.abs(y - pts[i][1]) <= hs)
        return i;
    }
  } else if (shape.cmd === "circle") {
    const pts = getHandlePointsCircle(+shape.x, +shape.y, +shape.r);
    for (let i = 0; i < pts.length; i++) {
      if (Math.abs(x - pts[i][0]) <= hs && Math.abs(y - pts[i][1]) <= hs)
        return i;
    }
  } else if (shape.cmd === "text") {
    // NOTE : le y du texte correspond à la ligne de base (bas du texte)
    ctx.font = `${+(shape.size || 20)}px 'Courier New', monospace`;
    const w = ctx.measureText(shape.value || "").width;
    const h = +(shape.size || 20);
    const pts = getHandlePointsRect(+shape.x, +shape.y - h, w, h);
    for (let i = 0; i < pts.length; i++) {
      if (Math.abs(x - pts[i][0]) <= hs && Math.abs(y - pts[i][1]) <= hs)
        return i;
    }
  }
  return -1;
}

// ==================================================
// REDIMENSIONNEMENT
// ==================================================

// Calcule les nouvelles propriétés d'une forme en fonction du déplacement d'une poignée de redimensionnement. Applique les règles de redimensionnement spécifiques à chaque type de forme (rectangle, cercle, texte) et impose une taille minimale pour éviter que la forme ne s'inverse ou ne disparaisse.
// Paramètres : TODO
// Retourne : TODO
function applyHandleMove(cmd, handleIdx, dx, dy, orig) {
  if (cmd === "rect") {
    let { x, y, w, h } = orig;
    // Au cas où les valeurs seraient des strings, on les convertit en nombres pour éviter les concaténations lors des calculs.
    x = +x;
    y = +y;
    w = +w;
    h = +h;
    switch (handleIdx) {
      case 0: // TL
        x += dx;
        y += dy;
        w -= dx;
        h -= dy;
        break;
      case 1: // TC
        y += dy;
        h -= dy;
        break;
      case 2: // TR
        y += dy;
        w += dx;
        h -= dy;
        break;
      case 3: // ML
        x += dx;
        w -= dx;
        break;
      case 4: // MR
        w += dx;
        break;
      case 5: // BL
        x += dx;
        w -= dx;
        h += dy;
        break;
      case 6: // BC
        h += dy;
        break;
      case 7: // BR
        w += dx;
        h += dy;
        break;
    }

    // Pour éviter que le rectangle ne s'inverse si on dépasse du côté opposé, on impose une taille minimale de 5 pixels. Si la largeur ou la hauteur devient inférieure à 5, on la remet à 5 et on ajuste la position en conséquence pour que le handle suive le curseur.
    if (w < 5) {
      x = x + w - 5;
      w = 5;
    }
    if (h < 5) {
      y = y + h - 5;
      h = 5;
    }
    return {
      x: Math.round(x),
      y: Math.round(y),
      w: Math.round(w),
      h: Math.round(h),
    };
  } else if (cmd === "circle") {
    const delta = [-dy, dx, dy, -dx][handleIdx] ?? 0;
    return { r: Math.max(5, Math.round(+orig.r + delta)) };
  } else if (cmd === "text") {
    // Pour le texte, on empêche la modification de taille par les côtés (TODO : faire en sorte que les handles de côtés modifient la taille aussi)
    let { x, y, w, h } = orig;
    x = +x;
    y = +y;
    w = +w;
    h = +h;
    // NOTE : le y du texte correspond à la ligne de base (bas du texte)
    y -= h; // On convertit le y de la ligne de base au y du coin supérieur pour réutiliser la même logique que pour les rectangles.
    switch (handleIdx) {
      case 0: // TL
        x += dx;
        y += dy;
        w -= dx;
        h -= dy;
        break;
      case 1: // TC
        y += dy;
        h -= dy;
        break;
      case 2: // TR
        y += dy;
        w += dx;
        h -= dy;
        break;
      case 3: // ML
        break;
      case 4: // MR
        break;
      case 5: // BL
        x += dx;
        w -= dx;
        h += dy;
        break;
      case 6: // BC
        h += dy;
        break;
      case 7: // BR
        w += dx;
        h += dy;
        break;
    }
    return {
      x: Math.round(x),
      y: Math.round(y + h),
      size: Math.max(8, Math.round(h)),
    };
  }
  return {};
}

// ==================================================
// GESTION DES INTERACTIONS
// ==================================================

// Calcule les coordonnées relatives au canvas à partir des coordonnées de l'événement de souris, en tenant compte de la position du canvas dans la page.
// Paramètres : e (événement de souris)
// Retourne : un objet avec les propriétés x et y représentant les coordonnées relatives au canvas
function getPos(e) {
  const r = canvas.getBoundingClientRect();
  return { x: e.clientX - r.left, y: e.clientY - r.top };
}

// Gère l'événement de mousedown sur le canvas pour initier une interaction en fonction de l'outil sélectionné.
// Pour l'outil de sélection, vérifie si une poignée de redimensionnement est cliquée pour entrer en mode "handle",
// sinon vérifie si une forme est cliquée pour la sélectionner et entrer en mode "move".
// Pour l'outil de texte, crée une nouvelle forme de texte à la position du clic et invite l'utilisateur à saisir le contenu du texte.
canvas.addEventListener("mousedown", (e) => {
  if (pendingState) return; // On ignore les clics de souris tant qu'on attend la confirmation du serveur pour éviter les conflits d'état.

  const { x, y } = getPos(e);
  isMouseDown = true;
  mouseStart = { x, y };

  if (tool === "select") {
    // On vérifie si on a cliqué sur une poignée de redimensionnement d'une forme déjà sélectionnée.
    // Si oui, on entre en mode "handle" pour redimensionner la forme.
    const sel = shapes[selectedId];
    if (sel) {
      const h = hitHandle(x, y, sel);
      if (h !== -1) {
        // Clic sur une poignée de redimensionnement — calcul manuel de la boîte englobante pour les formes de texte
        let origExtra = {};
        if (sel.cmd === "text") {
          ctx.font = `${+(sel.size || 20)}px 'Courier New', monospace`;
          origExtra = {
            w: ctx.measureText(sel.value || "").width,
            h: +(sel.size || 20),
          };
        }
        dragState = {
          type: "handle",
          handleIdx: h,
          startX: x,
          startY: y,
          origShape: { ...sel, ...origExtra },
          editedShape: { ...sel, ...origExtra },
        };
        return;
      }
    }

    // Si on n'a pas cliqué sur une poignée de redimensionnement,
    // on teste si on a cliqué sur une forme pour la sélectionner et éventuellement la déplacer.
    const hit = hitTest(x, y);
    if (hit) {
      selectShape(hit);
      dragState = {
        type: "move",
        startX: x,
        startY: y,
        origShape: { ...shapes[hit] },
        editedShape: { ...shapes[hit] },
      };
    } else {
      // Clic sur le fond du canvas, on désélectionne tout.
      selectShape(null);
    }
  } else if (tool === "text") {
    const val = prompt("Enter text:", "Text");
    if (val !== null) {
      const id = crypto.randomUUID();
      const s = {
        op: "create",
        cmd: "text",
        id,
        x: Math.round(x),
        y: Math.round(y),
        value: val,
        size: 20,
        color: currentColor(),
      };
      operation = encode(s);
      sendOut(operation);
      pendingState = { op: operation, previewShape: s, selected: false };
      redraw();
    }
  } else {
    drawState = { type: tool, startX: x, startY: y, previewShape: null };
  }
});

// Met à jour le style du curseur en fonction de l'outil sélectionné
// et de la position de la souris par rapport aux formes sur le canvas.
function updateCursorStyle(x, y) {
  if (tool !== "select") {
    canvas.style.cursor = "crosshair";
    return;
  }
  const sel = shapes[selectedId];
  if (sel) {
    const h = hitHandle(x, y, sel);
    if (h !== -1) {
      const cursors = [
        "nw-resize",
        "n-resize",
        "ne-resize",
        "w-resize",
        "e-resize",
        "sw-resize",
        "s-resize",
        "se-resize",
      ];
      canvas.style.cursor = cursors[h] ?? "pointer";
      return;
    }
  }
  canvas.style.cursor = hitTest(x, y) ? "move" : "default";
  return;
}

// Gère l'événement de mousemove sur le canvas pour mettre à jour la position de la souris
// et effectuer les actions appropriées en fonction de l'état actuel de l'interaction
// (par exemple, déplacer une forme, redimensionner une forme, ou mettre à jour le style du curseur).
canvas.addEventListener("mousemove", (e) => {
  if (pendingState) return; // On ignore les mouvements de souris tant qu'on attend la confirmation du serveur pour éviter les conflits d'état.
  const { x, y } = getPos(e);
  lastMousePos = { x, y };

  // Cursor when not dragging
  if (!isMouseDown) {
    updateCursorStyle(x, y);
  } else if (dragState) {
    const s = shapes[dragState.origShape.id];
    if (!s) return;
    const dx = x - dragState.startX;
    const dy = y - dragState.startY;

    if (dragState.type === "move") {
      dragState.editedShape.x = Math.round(+dragState.origShape.x + dx);
      dragState.editedShape.y = Math.round(+dragState.origShape.y + dy);
    } else if (dragState.type === "handle") {
      const patch = applyHandleMove(
        s.cmd,
        dragState.handleIdx,
        dx,
        dy,
        dragState.origShape,
      );
      Object.assign(shapes[s.id], patch);
    }
    updatePropsPanel();
    redraw();
    return;
  }

  if (drawState) {
    const x1 = Math.min(drawState.startX, x),
      y1 = Math.min(drawState.startY, y);
    const x2 = Math.max(drawState.startX, x),
      y2 = Math.max(drawState.startY, y);
    if (drawState.type === "rect") {
      drawState.previewShape = {
        cmd: "rect",
        id: "_preview",
        x: x1,
        y: y1,
        w: x2 - x1,
        h: y2 - y1,
        color: currentColor(),
      };
    } else if (drawState.type === "circle") {
      const r = Math.max(x2 - x1, y2 - y1) / 2;
      drawState.previewShape = {
        cmd: "circle",
        id: "_preview",
        x: (x1 + x2) / 2,
        y: (y1 + y2) / 2,
        r,
        color: currentColor(),
      };
    }
    redraw();
  }
});

// Gère l'événement de mouseup sur le canvas pour finaliser une interaction en fonction de l'état actuel
// (par exemple, terminer le déplacement ou le redimensionnement d'une forme, ou créer une nouvelle forme à partir du dessin en cours).
// Envoie les mises à jour nécessaires au serveur et réinitialise les états d'interaction.
canvas.addEventListener("mouseup", (e) => {
  if (pendingState) return; // On ignore les clics de souris tant qu'on attend la confirmation du serveur pour éviter les conflits d'état.

  const { x, y } = getPos(e);

  if (dragState) {
    const s = dragState.editedShape;
    const orig = dragState.origShape;
    const diff = {};
    // On compare les propriétés pertinentes de la forme modifiée avec celles de la forme originale
    // pour construire un objet "diff" qui ne contient que les propriétés qui ont changé.
    // Cela permet d'envoyer au serveur uniquement les données nécessaires pour mettre à jour la forme,
    // plutôt que d'envoyer l'intégralité de la forme même si seule une partie a été modifiée.
    for (const k of ["x", "y", "w", "h", "r", "size"]) {
      if (s[k] !== undefined && String(s[k]) !== String(orig[k]))
        diff[k] = s[k];
    }
    if (Object.keys(diff).length > 0) {
      operation = encode({ op: "update", id: s.id, ...diff });
      sendOut(operation);
      pendingState = { op: operation, previewShape: s, selected: true };
    }
    dragState = null;
    redraw();
  } else if (drawState?.previewShape) {
    // Cas de création d'une forme
    const ps = drawState.previewShape;
    const ok =
      (ps.cmd === "rect" && +ps.w > 5 && +ps.h > 5) ||
      (ps.cmd === "circle" && +ps.r > 5);
    if (ok) {
      const id = crypto.randomUUID();
      const s = {
        op: "create",
        ...ps,
        id,
        x: Math.round(+ps.x),
        y: Math.round(+ps.y),
        ...(ps.cmd === "rect"
          ? { w: Math.round(+ps.w), h: Math.round(+ps.h) }
          : {}),
        ...(ps.cmd === "circle" ? { r: Math.round(+ps.r) } : {}),
      };

      operation = encode(s);
      sendOut(operation);
      pendingState = { op: operation, previewShape: ps, selected: false };
    } else {
      addToLog("Shape too small, ignoring");
    }
  }

  isMouseDown = false;
  drawState = null;
  redraw();
});

// =========================================
// SÉLECTION ET PANNEAU DE PROPRIÉTÉS
// =========================================

// Met à jour l'id de la forme sélectionnée puis rafraîchit le panneau de propriétés et redessine le canvas pour refléter la sélection.
// Paramètres : TODO
// Retourne : TODO
function selectShape(id) {
  selectedId = id;
  updatePropsPanel();
  redraw();
}

// Met à jour le panneau de propriétés en fonction de la forme actuellement sélectionnée.
// Affiche les propriétés pertinentes pour le type de forme sélectionné (rectangle, cercle, texte)
// et remplit les champs avec les valeurs actuelles de la forme.
// Si aucune forme n'est sélectionnée, affiche un message indiquant qu'aucune sélection n'est active.
// Paramètres : TODO
// Retourne : TODO
function updatePropsPanel() {
  const s = shapes[selectedId];
  document.getElementById("no-selection").style.display = s ? "none" : "block";
  document.getElementById("shape-props").style.display = s ? "flex" : "none";
  if (!s) return;

  document.getElementById("prop-id").value = s.id;
  document.getElementById("prop-x").value = Math.round(+s.x);
  document.getElementById("prop-y").value = Math.round(+s.y);

  document.getElementById("wh-props").style.display =
    s.cmd === "rect" ? "block" : "none";
  document.getElementById("r-prop").style.display =
    s.cmd === "circle" ? "block" : "none";
  document.getElementById("text-prop").style.display =
    s.cmd === "text" ? "block" : "none";
  // s.cmd === "text" ? "flex" : "none";

  if (s.cmd === "rect") {
    document.getElementById("prop-w").value = Math.round(+s.w);
    document.getElementById("prop-h").value = Math.round(+s.h);
  }
  if (s.cmd === "circle") {
    document.getElementById("prop-r").value = Math.round(+s.r);
  }
  if (s.cmd === "text") {
    document.getElementById("prop-value").value = s.value || "";
    document.getElementById("prop-size").value = +(s.size || 20);
  }

  document.getElementById("prop-color").value = cssToHex(s.color) || "#4488ff";
}

// Convertit une valeur de couleur CSS en format hexadécimal.
// Permet de choisir une couleur directement dans le navigateur.
// Gère les différentes notations de couleur CSS (noms de couleurs, rgb(), rgba(), hsl(), hsla())
// et retourne une chaîne de caractères représentant la couleur au format hexadécimal (#RRGGBB).
// Si la couleur est déjà au format hexadécimal, elle est retournée telle quelle.
// Si la couleur est invalide ou non définie, une couleur par défaut est retournée.
// Paramètres : TODO
// Retourne : TODO
function cssToHex(c) {
  // Si la couleur est undefined ou null, on retourne une couleur par défaut (bleu clair dans ce cas).
  if (!c) return "#4488ff";
  // Si la couleur est déjà au format hexadécimal, on la retourne telle quelle.
  if (/^#[0-9a-fA-F]{3,6}$/.test(c)) return c;

  const d = document.createElement("div");
  d.style.color = c;
  document.body.appendChild(d);
  const comp = getComputedStyle(d).color;
  document.body.removeChild(d);
  const m = comp.match(/\d+/g);
  if (!m) return "#ffffff";
  return (
    "#" +
    m
      .slice(0, 3)
      .map((n) => (+n).toString(16).padStart(2, "0"))
      .join("")
  );
}

// Gère les changements de propriétés dans le panneau de propriétés en mettant à jour la forme sélectionnée et en envoyant les modifications au serveur.
// Paramètres :
//  - key (la clé de la propriété modifiée),
//  - val (la nouvelle valeur de la propriété)
// Retourne : TODO
function propChanged(key, val) {
  if (!selectedId || !shapes[selectedId]) return;
  // const orig = { ...shapes[selectedId] };
  shapes[selectedId][key] = val;
  sendOut(encode({ op: "update", id: selectedId, [key]: val }));
  redraw();
}

// Ajoute des écouteurs d'événements "changement" à tous les champs de propriétés (position, dimensions, rayon, taille)
// pour détecter les modifications apportées par l'utilisateur.
// Lorsqu'une propriété est modifiée, la fonction propChanged est appelée avec la clé de la propriété
// et sa nouvelle valeur pour mettre à jour la forme sélectionnée et envoyer les changements au serveur.

// Gestion des champs de valeurs numériques
["prop-x", "prop-y", "prop-w", "prop-h", "prop-r", "prop-size"].forEach(
  (elId) => {
    const el = document.getElementById(elId);
    if (!el) return;
    const keyMap = {
      "prop-x": "x",
      "prop-y": "y",
      "prop-w": "w",
      "prop-h": "h",
      "prop-r": "r",
      "prop-size": "size",
    };
    el.addEventListener("change", (e) => {
      let key = keyMap[elId];
      let val = +e.target.value;
      if (key === "size" && val < TEXT_MIN_SIZE) {
        val = TEXT_MIN_SIZE;
        e.target.value = TEXT_MIN_SIZE;
      }
      propChanged(key, val);
    });
  },
);

document.getElementById("prop-value").addEventListener("change", (e) => {
  propChanged("value", e.target.value);
});

document.getElementById("prop-color").addEventListener("input", (e) => {
  propChanged("color", e.target.value);
});

// Récupère la couleur actuelle sélectionnée dans le panneau de propriétés,
// ou retourne une couleur par défaut si aucune couleur n'est définie (ici bleu clair).
// Paramètres : aucun
// Retourne : une chaîne de caractères représentant la couleur actuelle au format hexadécimal
function currentColor() {
  return document.getElementById("prop-color").value || "#4488ff";
}

// ==================================================
// BARRE D'OUTILS
// ==================================================

// Gère la sélection de l'outil actif dans la barre d'outils en ajoutant
// un écouteur d'événements "click" à chaque bouton d'outil.
document.querySelectorAll(".tool-btn[data-tool]").forEach((btn) => {
  btn.addEventListener("click", () => {
    tool = btn.dataset.tool;
    document
      .querySelectorAll(".tool-btn[data-tool]")
      .forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    if (tool !== "select") selectShape(null);
  });
});

// Gère la suppression de la forme sélectionnée en envoyant une opération de suppression au serveur,
// en supprimant la forme du tableau local et en désélectionnant la forme.
// Paramètres : aucun
// Retourne : TODO
function deleteSelected() {
  if (!selectedId) return;
  sendOut(encode({ op: "delete", id: selectedId }));
  delete shapes[selectedId];
  selectShape(null);
  redraw();
}

document.getElementById("delete-btn").addEventListener("click", deleteSelected);
document.getElementById("clear-btn").addEventListener("click", () => {
  shapes = {};
  selectShape(null);
  sendOut(encode({ op: "clear" }));
  redraw();
});

// Gère les raccourcis clavier pour les outils de dessin, la sélection, la suppression et l'échappement.
window.addEventListener("keydown", (e) => {
  if (e.target.tagName === "INPUT") return;
  const map = {
    v: "select",
    V: "select",
    r: "rect",
    R: "rect",
    c: "circle",
    C: "circle",
    t: "text",
    T: "text",
  };
  if (map[e.key]) {
    document.querySelector(`[data-tool="${map[e.key]}"]`).click();
    updateCursorStyle(lastMousePos.x, lastMousePos.y);
  }
  if (e.key === "Delete" || e.key === "Backspace") deleteSelected();
  if (e.key === "Escape") selectShape(null);
});

// ==================================================
// GESTION DE LA COMMUNICATION AVEC LE SERVEUR
// ==================================================

// Gère le clic sur le bouton "Connecter" pour établir une connexion WebSocket avec le serveur.
// Récupère les valeurs de l'hôte et du port à partir des champs de saisie,
// crée une nouvelle instance de WebSocket, etdéfinit les gestionnaires d'événements pour
// l'ouverture, la fermeture, la réception de messages et les erreurs.
// Affiche les événements dans les logs avec des préfixes appropriés.
document.getElementById("connecter").onclick = function (evt) {
  if (ws) {
    return false;
  }

  var host = document.getElementById("host").value;
  var port = document.getElementById("port").value;

  addToLog("[INFO] Tentative de connexion");
  addToLog("[INFO] host = " + host + ", port = " + port);

  ws = new WebSocket("ws://" + host + ":" + port + "/ws");
  ws.onopen = function (evt) {
    addToLog("[INFO] Websocket ouverte");
    document.title = "🌐 " + port;
  };

  ws.onclose = function (evt) {
    addToLog("[INFO] Websocket fermée");
    ws = null;
    document.title = "🖼️ Collaborative Whiteboard";
  };

  // TODO : cacher le canva si aucune ws ouverte pour éviter les interactions sans connexion,
  // ou au moins empêcher les modifications du white board local tant que la connexion n'est pas établie,
  // et afficher un message d'erreur si l'utilisateur essaie d'interagir sans connexion.

  ws.onmessage = function (evt) {
    addToLog("[IN] " + evt.data);
    handleReceive(evt.data);
  };

  ws.onerror = function (evt) {
    addToLog("[ERROR] " + evt.data);
  };
  return false;
};

// Gère le clic sur le bouton "Fermer" pour fermer la connexion WebSocket avec le serveur.
// Vérifie d'abord si la connexion est établie avant de tenter de la fermer, et affiche un message dans les logs.
document.getElementById("fermer").onclick = function (evt) {
  if (!ws) {
    return false;
  }
  ws.close();
  // TODO : gérer la fermeture de la connexion côté client en désactivant les interactions avec le white board et en affichant un message d'information à l'utilisateur.
  return false;
};

// Traite un message reçu du serveur en le décomposant en préfixe et contenu,
// puis en appelant la fonction appropriée pour appliquer les changements
// au white board en fonction du type de message (par exemple, mise à jour des données).
// Paramètres : msg (une chaîne de caractères représentant le message reçu du serveur, encodé selon le protocole défini)
// Retourne : TODO
function handleReceive(msg) {
  var [prefix, ope] = msg.split(MSG_KEY_VALUE_SEP);
  switch (prefix) {
    case "data": {
      applyMsg(ope);
    }
  }
}

// Applique une opération reçue du serveur en mettant à jour les données du white board en conséquence. Gère les différentes opérations possibles (création, mise à jour, suppression de formes, effacement total) et met à jour la sélection si nécessaire.
// Paramètres : ope (une chaîne de caractères représentant l'opération reçue du serveur, encodée selon le protocole défini)
// Retourne : TODO
function applyMsg(ope) {
  const d = decode(ope);

  if (pendingState && pendingState.op !== ope) {
    // L'opération reçue ne correspond pas à celle en attente, ce n'est pas censé avoir lieu.
    addToLog("[WARN] Received operation does not match pending operation");
  }

  if (d.op === "delete") {
    delete shapes[d.id];
    if (selectedId === d.id) selectShape(null);
  } else if (d.op === "update") {
    if (!shapes[d.id]) return;
    const { op, id, ...rest } = d;
    // Comme les valeurs dans le message sont des chaînes de caractères,
    // on convertit en nombres les propriétés qui en ont besoin pour le dessin et les calculs (x, y, w, h, r, size).
    for (const k of ["x", "y", "w", "h", "r", "size"]) {
      if (rest[k] !== undefined) rest[k] = +rest[k];
    }

    if (pendingState && pendingState.op === ope) {
      // Si on reçoit une opération de mise à jour alors qu'on est en train d'attendre la confirmation d'une opération de mise à jour précédente (pendingState), on vérifie si l'opération reçue correspond à celle en attente (en comparant les opérations encodées). Si c'est le cas, cela signifie que le serveur a confirmé la modification de la forme que nous avons initiée, et nous pouvons alors effacer l'état d'attente (pendingState) pour permettre de nouvelles interactions.
      shapes[d.id] = pendingState.previewShape;
      pendingState = null;
    }

    Object.assign(shapes[d.id], rest);
  } else if (d.op === "clear") {
    shapes = {};
    selectShape(null);
  } else if (d.op === "create") {
    const { cmd, id, ...rest } = d;
    for (const k of ["x", "y", "w", "h", "r", "size"]) {
      if (rest[k] !== undefined) rest[k] = +rest[k];
    }
    shapes[id] = { cmd, id, ...rest };

    if (pendingState && pendingState.op === ope) {
      // Si on reçoit une opération de création alors qu'on est en train d'attendre la confirmation d'une opération de création précédente (pendingState), on vérifie si l'opération reçue correspond à celle en attente (en comparant les opérations encodées). Si c'est le cas, cela signifie que le serveur a confirmé la création de la forme que nous avons initiée, et nous pouvons alors effacer l'état d'attente (pendingState) pour permettre de nouvelles interactions.
      pendingState = null;
      selectShape(id);
    }
  }
  redraw();
}

// Envoie une opération déjà encodée au serveur via la WebSocket, après avoir vérifié que la connexion est établie. Affiche l'opération dans les logs avec le préfixe "[OUT]".
// Paramètres : operation (une chaîne de caractères représentant l'opération à envoyer, déjà encodée selon le protocole défini)
// Retourne : TODO
async function sendOut(operation) {
  if (!ws) {
    return false;
  }

  addToLog("[OUT] " + operation);

  // TODO : gérer l'envoi d'un message temporaire et éditer le white board uniquement
  // à partir de la réception de l'accusé de réception du serveur pour éviter les problèmes
  // de synchronisation dus au délai de transmission. Par exemple, on peut ajouter une propriété
  // "pending" à la forme en cours de création ou de modification, et ne la rendre définitive qu'à
  // la réception du message du serveur confirmant l'opération.
  await updateRemote(operation);

  return false;
}

const updateRemote = async (newOperation) => {
  ws.send("data" + MSG_KEY_VALUE_SEP + newOperation);
};

// ==================================================
// LOGS
// ==================================================

// Ajoute un message au journal des logs en l'affichant dans la console
// du navigateur et en l'ajoutant à la section des messages dans la page.
// Fait défiler automatiquement les logs vers le bas pour afficher les messages les plus récents.
function addToLog(message) {
  console.log(message);
  var logs = document.getElementById("log-messages");
  var d = document.createElement("div");
  d.textContent = message;
  logs.appendChild(d);
  logs.scroll(0, logs.scrollHeight);
}

// ==================================================

// document.getElementById("debut_sc").onclick = function (evt) {
//   if (!ws) {
//     return false;
//   }
//   ws.send("section_critique");
//   return false;
// };
// document.getElementById("fin_sc").onclick = function (evt) {
//   if (!ws) {
//     return false;
//   }
//   ws.send("f_section_critique");
//   return false;
// };

// ==============
//    SNAPSHOT
// ==============

document.getElementById("snapshot").onclick = function (evt) {
  if (ws) {
    addToLog("[OUT] SNAPSHOT");
    ws.send("snapshot" + MSG_KEY_VALUE_SEP + "true");
    return;
  }
  addToLog("SNAPSHOT IMPOSSIBLE : ws close");
};
