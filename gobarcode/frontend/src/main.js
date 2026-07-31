import './style.css';
import './app.css';

import * as go from '../wailsjs/go/main/App';

const CANVAS_WIDTH = 600;
const CANVAS_HEIGHT = 360;
const DEFAULT_PAGE_WIDTH_INCHES = 8.5;
const DEFAULT_PAGE_HEIGHT_INCHES = 11;
const DEFAULT_PPI = 150;
const MIN_ELEMENT_WIDTH = 60;
const MIN_ELEMENT_HEIGHT = 32;
const STORAGE_KEY = 'gobarcode-compositor-layout-v3';

const PLACEMENTS = {
  barcode_placement: {
    type: 'barcode',
    label: 'Barcode',
    source: () => upcSelect.value || 'UPC column',
  },
  title_placement: {
    type: 'title',
    label: 'Title',
    source: () => titleSelect.value || 'Title column',
  },
};

const filebtn = document.getElementById('filebtn');
const workName = document.getElementById('sheetname');
const setRow = document.getElementById('btnsetRow');
const headerInput = document.getElementById('headerinput');
const upcSelect = document.getElementById('header-1-select');
const titleSelect = document.getElementById('header-2-select');
const filterColumnSelect = document.getElementById('filter-column-select');
const filterTextInput = document.getElementById('filter-text');
const skipMissingUpcInput = document.getElementById('skip-missing-upc');
const padOddUpcInput = document.getElementById('pad-odd-upc');
const submitbtn = document.getElementById('submitbtn');
const canvas = document.getElementById('label-canvas');
const status = document.getElementById('editor-status');
const properties = document.getElementById('element-properties');
const noSelection = document.getElementById('no-selection');
const propertyKind = document.getElementById('property-kind');
const propertySource = document.getElementById('property-source');
const layoutSummary = document.getElementById('layout-summary');
const pageSizeSummary = document.getElementById('page-size-summary');
const labelWidthRuler = document.getElementById('label-width-ruler');
const labelHeightRuler = document.getElementById('label-height-ruler');
const saveLocationButton = document.getElementById('save-location-button');
const saveLocationDisplay = document.getElementById('save-location');
const applyLayoutButton = document.getElementById('save-layout');
const generatePdfButton = document.getElementById('generate-pdf');
const propertyInputs = {
  origin_x: document.getElementById('property-x'),
  origin_y: document.getElementById('property-y'),
  width: document.getElementById('property-width'),
  height: document.getElementById('property-height'),
};
const outputInputs = {
  page_width: document.getElementById('page-width'),
  page_height: document.getElementById('page-height'),
  ppi: document.getElementById('layout-ppi'),
};

let layout = createDefaultLayout();
let selectedPlacement = null;
let interaction = null;
let saveLocation = '';

filebtn.addEventListener('click', selectFile);
setRow.addEventListener('click', setHeader);
submitbtn.addEventListener('click', submitColumns);
upcSelect.addEventListener('change', renderLayout);
titleSelect.addEventListener('change', renderLayout);
filterColumnSelect.addEventListener('change', updateFilterAvailability);
applyLayoutButton.addEventListener('click', applyLayout);
generatePdfButton.addEventListener('click', generatePDF);
document.getElementById('reset-layout').addEventListener('click', resetLayout);
saveLocationButton.addEventListener('click', selectSaveLocation);
canvas.addEventListener('pointerdown', handleCanvasPointerDown);
window.addEventListener('pointermove', handlePointerMove);
window.addEventListener('pointerup', endInteraction);
window.addEventListener('pointercancel', endInteraction);
document.addEventListener('keydown', handleKeyDown);

Object.entries(propertyInputs).forEach(([property, input]) => {
  input.addEventListener('change', () => updateSelectedGeometry(property, input.value));
});
Object.entries(outputInputs).forEach(([property, input]) => {
  input.addEventListener('change', () => updateOutputProperty(property, input.value));
});

loadLayout();

function createDefaultLayout() {
  return {
    image_height: CANVAS_HEIGHT,
    image_width: CANVAS_WIDTH,
    barcode_placement: {
      height: 100,
      width: 280,
      origin_x: 160,
      origin_y: 190,
    },
    title_placement: {
      height: 54,
      width: 360,
      origin_x: 120,
      origin_y: 92,
    },
    page_height: DEFAULT_PAGE_HEIGHT_INCHES,
    page_width: DEFAULT_PAGE_WIDTH_INCHES,
    ppi: DEFAULT_PPI,
  };
}

async function selectFile() {
  try {
    const fileInfo = await go.SelectFile();
    workName.textContent = fileInfo.selected_sheet_name;
    setStatus('Spreadsheet loaded');
  } catch (error) {
    setStatus(error, true);
  }
}

async function setHeader() {
  const row = Number(headerInput.value);
  if (!Number.isInteger(row) || row < 1) {
    setStatus('Header row must be 1 or greater', true);
    return;
  }

  try {
    const headers = await go.GetHeaders(row);
    populateHeaderSelect(upcSelect, headers);
    populateHeaderSelect(titleSelect, headers);
    populateHeaderSelect(filterColumnSelect, headers, 'No filter');
    updateFilterAvailability();
    renderLayout();
    setStatus(`Loaded ${headers.length} columns from row ${row}`);
  } catch (error) {
    setStatus(error, true);
  }
}

function populateHeaderSelect(select, headers, emptyOption = '') {
  select.replaceChildren();
  if (emptyOption) {
    const option = document.createElement('option');
    option.value = '';
    option.textContent = emptyOption;
    select.appendChild(option);
  }
  headers.forEach((header) => {
    const option = document.createElement('option');
    option.value = header;
    option.textContent = header;
    select.appendChild(option);
  });
}

function updateFilterAvailability() {
  const enabled = Boolean(filterColumnSelect.value);
  filterTextInput.disabled = !enabled;
  if (!enabled) filterTextInput.value = '';
}

async function submitColumns() {
  if (!upcSelect.value || !titleSelect.value) {
    setStatus('Choose both a UPC and title column', true);
    return;
  }

  try {
    const filterText = filterTextInput.value.trim();
    await go.SetColumns(
      upcSelect.value,
      titleSelect.value,
      filterColumnSelect.value,
      filterText,
      skipMissingUpcInput.checked,
      padOddUpcInput.checked,
    );
    renderLayout();
    const filterApplied = filterColumnSelect.value && filterText;
    setStatus(filterApplied ? `Spreadsheet filtered by ${filterColumnSelect.value}` : 'Spreadsheet columns saved');
  } catch (error) {
    setStatus(error, true);
  }
}

async function selectSaveLocation() {
  saveLocationButton.disabled = true;
  try {
    const location = await go.SetSaveLocation();
    saveLocation = location;
    saveLocationDisplay.textContent = location;
    saveLocationDisplay.title = location;
    setStatus('Save location selected');
  } catch (error) {
    setStatus(error, true);
  } finally {
    saveLocationButton.disabled = false;
  }
}

function renderLayout() {
  updateOutputPanel();
  canvas.replaceChildren();
  Object.entries(PLACEMENTS).forEach(([key, definition]) => {
    const placement = layout[key];
    const elementNode = document.createElement('div');
    elementNode.className = `design-element ${definition.type}-element`;
    elementNode.dataset.placement = key;
    elementNode.style.left = `${placement.origin_x}px`;
    elementNode.style.top = `${placement.origin_y}px`;
    elementNode.style.width = `${placement.width}px`;
    elementNode.style.height = `${placement.height}px`;
    elementNode.classList.toggle('selected', key === selectedPlacement);

    if (definition.type === 'barcode') {
      const bars = document.createElement('div');
      bars.className = 'barcode-bars';
      const caption = document.createElement('span');
      caption.className = 'barcode-caption';
      caption.textContent = `{${definition.source()}}`;
      elementNode.append(bars, caption);
    } else {
      const title = document.createElement('span');
      title.className = 'title-placeholder';
      title.textContent = `{${definition.source()}}`;
      elementNode.appendChild(title);
    }

    const resizeHandle = document.createElement('button');
    resizeHandle.className = 'resize-handle';
    resizeHandle.type = 'button';
    resizeHandle.dataset.action = 'resize';
    resizeHandle.setAttribute('aria-label', `Resize ${definition.label.toLowerCase()}`);
    elementNode.appendChild(resizeHandle);
    canvas.appendChild(elementNode);
  });
  updatePropertiesPanel();
}

function updateOutputPanel() {
  Object.entries(outputInputs).forEach(([property, input]) => {
    input.value = layout[property];
  });

  const pageWidthPixels = Math.round(layout.page_width * layout.ppi);
  const pageHeightPixels = Math.round(layout.page_height * layout.ppi);
  const labelWidthInches = layout.image_width / layout.ppi;
  const labelHeightInches = layout.image_height / layout.ppi;
  pageSizeSummary.textContent = `${pageWidthPixels} × ${pageHeightPixels} pixels`;
  layoutSummary.textContent = `${layout.image_width}×${layout.image_height} px label · ${formatInches(layout.page_width)}×${formatInches(layout.page_height)} in page · ${layout.ppi} PPI`;
  labelWidthRuler.textContent = `${formatInches(labelWidthInches)} in`;
  labelHeightRuler.textContent = `${formatInches(labelHeightInches)} in`;
}

function handleCanvasPointerDown(event) {
  const elementNode = event.target.closest('.design-element');
  if (!elementNode) {
    selectedPlacement = null;
    renderLayout();
    return;
  }

  event.preventDefault();
  selectedPlacement = elementNode.dataset.placement;
  const placement = getSelectedPlacement();
  if (!placement) return;

  canvas.querySelectorAll('.design-element.selected').forEach((node) => {
    node.classList.remove('selected');
  });
  elementNode.classList.add('selected');
  interaction = {
    mode: event.target.dataset.action === 'resize' ? 'resize' : 'drag',
    pointerId: event.pointerId,
    startX: event.clientX,
    startY: event.clientY,
    originX: placement.origin_x,
    originY: placement.origin_y,
    originWidth: placement.width,
    originHeight: placement.height,
  };
  canvas.setPointerCapture?.(event.pointerId);
  updatePropertiesPanel();
}

function handlePointerMove(event) {
  if (!interaction || interaction.pointerId !== event.pointerId) return;
  const placement = getSelectedPlacement();
  if (!placement) return;

  const deltaX = event.clientX - interaction.startX;
  const deltaY = event.clientY - interaction.startY;
  if (interaction.mode === 'drag') {
    placement.origin_x = interaction.originX + deltaX;
    placement.origin_y = interaction.originY + deltaY;
  } else {
    placement.width = Math.max(MIN_ELEMENT_WIDTH, interaction.originWidth + deltaX);
    placement.height = Math.max(MIN_ELEMENT_HEIGHT, interaction.originHeight + deltaY);
  }
  keepPlacementInBounds(placement);
  updateElementNode();
  updatePropertiesPanel();
}

function endInteraction(event) {
  if (!interaction || interaction.pointerId !== event.pointerId) return;
  if (canvas.hasPointerCapture?.(event.pointerId)) {
    canvas.releasePointerCapture(event.pointerId);
  }
  interaction = null;
}

function keepPlacementInBounds(placement) {
  placement.width = clamp(Math.round(placement.width), MIN_ELEMENT_WIDTH, layout.image_width);
  placement.height = clamp(Math.round(placement.height), MIN_ELEMENT_HEIGHT, layout.image_height);
  placement.origin_x = clamp(Math.round(placement.origin_x), 0, layout.image_width - placement.width);
  placement.origin_y = clamp(Math.round(placement.origin_y), 0, layout.image_height - placement.height);
}

function updateElementNode() {
  const placement = getSelectedPlacement();
  const node = canvas.querySelector(`[data-placement="${selectedPlacement}"]`);
  if (!placement || !node) return;
  node.style.left = `${placement.origin_x}px`;
  node.style.top = `${placement.origin_y}px`;
  node.style.width = `${placement.width}px`;
  node.style.height = `${placement.height}px`;
}

function updatePropertiesPanel() {
  const placement = getSelectedPlacement();
  const definition = PLACEMENTS[selectedPlacement];
  properties.hidden = !placement;
  noSelection.hidden = Boolean(placement);
  if (!placement || !definition) return;

  propertyKind.textContent = definition.label;
  propertySource.textContent = definition.source();
  Object.entries(propertyInputs).forEach(([property, input]) => {
    input.value = placement[property];
  });
}

function updateSelectedGeometry(property, rawValue) {
  const placement = getSelectedPlacement();
  const value = Number(rawValue);
  if (!placement || !Number.isFinite(value)) return;
  placement[property] = value;
  keepPlacementInBounds(placement);
  renderLayout();
}

function updateOutputProperty(property, rawValue) {
  const value = Number(rawValue);
  const validPpi = property !== 'ppi' || Number.isInteger(value);
  if (!Number.isFinite(value) || value <= 0 || !validPpi) {
    outputInputs[property].value = layout[property];
    const requirement = property === 'ppi' ? 'a positive whole number' : 'greater than zero';
    setStatus(`${property.replace('_', ' ')} must be ${requirement}`, true);
    return;
  }
  layout[property] = value;
  renderLayout();
  setStatus('Output settings updated');
}

function handleKeyDown(event) {
  const placement = getSelectedPlacement();
  const editingField = ['INPUT', 'SELECT', 'TEXTAREA'].includes(document.activeElement?.tagName);
  if (!placement || editingField) return;

  const movement = event.shiftKey ? 10 : 1;
  const directions = {
    ArrowLeft: [-movement, 0],
    ArrowRight: [movement, 0],
    ArrowUp: [0, -movement],
    ArrowDown: [0, movement],
  };
  if (!directions[event.key]) return;

  event.preventDefault();
  placement.origin_x += directions[event.key][0];
  placement.origin_y += directions[event.key][1];
  keepPlacementInBounds(placement);
  renderLayout();
}

async function applyLayout() {
  try {
    await go.SetLayout(layout);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(layout));
    setStatus('Layout validated and applied');
  } catch (error) {
    setStatus(error, true);
  }
}

async function generatePDF() {
  if (!saveLocation) {
    setStatus('Choose an output file before generating the PDF', true);
    return;
  }

  setGenerationBusy(true);
  try {
    await go.SetLayout(layout);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(layout));
    setStatus('Generating PDF…');
    await go.Start();
    setStatus(`PDF saved to ${saveLocation}`);
  } catch (error) {
    setStatus(error, true);
  } finally {
    setGenerationBusy(false);
  }
}

function setGenerationBusy(isBusy) {
  [generatePdfButton, applyLayoutButton, saveLocationButton, filebtn, setRow, submitbtn]
    .forEach((button) => {
      button.disabled = isBusy;
    });
  generatePdfButton.textContent = isBusy ? 'Generating…' : 'Generate PDF';
  generatePdfButton.setAttribute('aria-busy', String(isBusy));
}

function loadLayout() {
  try {
    const saved = JSON.parse(localStorage.getItem(STORAGE_KEY));
    if (isValidLayout(saved)) {
      layout = saved;
      Object.keys(PLACEMENTS).forEach((key) => keepPlacementInBounds(layout[key]));
      setStatus('Restored saved compositor layout');
    }
  } catch (error) {
    console.warn('Could not restore saved label layout', error);
  }
  renderLayout();
}

function resetLayout() {
  layout = createDefaultLayout();
  selectedPlacement = null;
  localStorage.removeItem(STORAGE_KEY);
  renderLayout();
  setStatus('Layout reset');
}

function isValidLayout(candidate) {
  return candidate
    && Number.isInteger(candidate.image_width)
    && Number.isInteger(candidate.image_height)
    && Number.isFinite(candidate.page_width)
    && candidate.page_width > 0
    && Number.isFinite(candidate.page_height)
    && candidate.page_height > 0
    && Number.isInteger(candidate.ppi)
    && candidate.ppi > 0
    && Object.keys(PLACEMENTS).every((key) => isValidPlacement(candidate[key]));
}

function isValidPlacement(placement) {
  return placement
    && ['origin_x', 'origin_y', 'width', 'height']
      .every((key) => Number.isFinite(placement[key]));
}

function getSelectedPlacement() {
  return selectedPlacement ? layout[selectedPlacement] : null;
}

function clamp(value, minimum, maximum) {
  return Math.min(Math.max(value, minimum), maximum);
}

function formatInches(value) {
  return Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/0+$/, '').replace(/\.$/, '');
}

function setStatus(message, isError = false) {
  status.textContent = String(message?.message || message || 'Unknown error');
  status.classList.toggle('status-error', isError);
}
