import React, { useEffect, useRef, useState } from 'react';
import cytoscape from 'cytoscape';
import nodeHtmlLabel from 'cytoscape-node-html-label';
import dagre from 'cytoscape-dagre';
import coseBilkent from 'cytoscape-cose-bilkent';
import './CytoscapeTopology.css';

// Extensions registration
cytoscape.use(nodeHtmlLabel);
cytoscape.use(dagre);
cytoscape.use(coseBilkent);

const CytoscapeTopology = ({ topology, selectedDevice, onDeviceSelect }) => {
  const containerRef = useRef(null);
  const cyRef = useRef(null);
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({ nodes: 0, edges: 0, layers: 0 });

  // Layer configurations - 順序を逆転させて上位レイヤーが上に表示されるようにする
  const LAYER_CONFIG = {
    10: { name: 'Border Router/Leaf', color: '#e74c3c', order: 13 },  // 最上位
    11: { name: 'Security Appliances', color: '#c0392b', order: 12 },
    20: { name: 'DC Core Interconnect', color: '#8e44ad', order: 11 },
    30: { name: 'Fat-Tree Core Spine', color: '#2980b9', order: 10 },
    31: { name: 'Fat-Tree Agg Spine', color: '#3498db', order: 9 },
    32: { name: 'Spine Switches', color: '#5dade2', order: 8 },
    40: { name: 'Fat-Tree Edge/Leaf', color: '#27ae60', order: 7 },
    41: { name: 'Leaf Switches', color: '#2ecc71', order: 6 },
    42: { name: 'Aggregation Switches', color: '#58d68d', order: 5 },
    43: { name: 'Access Switches', color: '#85c1e9', order: 4 },
    50: { name: 'Servers', color: '#f39c12', order: 3 },           // 最下位
    51: { name: 'Storage Devices', color: '#e67e22', order: 2 },
    52: { name: 'Other Appliances', color: '#d35400', order: 1 }
  };

  const getDeviceIcon = (type) => {
    const icons = {
      'dc_core_interconnect': '🌐',
      'border_leaf': '🚪',
      'core_spine': '🏢',
      'agg_spine': '🏗️',
      'edge_leaf': '🍃',
      'spine': '🏛️',
      'leaf': '🌿',
      'core': '⚡',
      'aggregation': '📊',
      'access': '🔌',
      'server': '💻',
      'storage': '💾',
      'firewall': '🛡️',
      'router': '📡',
      'switch': '🔗',
      'unknown': '❓'
    };
    return icons[type] || icons.unknown;
  };

  const buildCytoscapeElements = (topology) => {
    const elements = [];
    const layerGroups = new Set();
    
    // Group nodes by layer
    const nodesByLayer = {};
    topology.nodes.forEach(node => {
      const layer = node.layer || 50; // Default to server layer
      if (!nodesByLayer[layer]) {
        nodesByLayer[layer] = [];
      }
      nodesByLayer[layer].push(node);
      layerGroups.add(layer);
    });

    // Create a mapping for visual ordering - 直接LAYER_CONFIGのorderを使用
    const getVisualOrder = (layer) => {
      const config = LAYER_CONFIG[layer];
      if (config) {
        return config.order;
      }
      return 99; // 未定義レイヤーは最下位
    };

    // Create layer parent nodes (compound nodes)
    layerGroups.forEach(layerId => {
      const config = LAYER_CONFIG[layerId] || { 
        name: `Layer ${layerId}`, 
        color: '#95a5a6', 
        order: 99 
      };
      
      elements.push({
        data: {
          id: `layer-${layerId}`,
          label: `${config.name}\n(${nodesByLayer[layerId].length} devices)`,
          type: 'layer',
          layer: layerId,
          visualOrder: getVisualOrder(layerId), // Add visual ordering
          isParent: true
        },
        classes: 'layer-node'
      });
    });

    // Add device nodes as children of layer nodes
    topology.nodes.forEach(node => {
      const layer = node.layer || 50;
      const config = LAYER_CONFIG[layer] || { color: '#95a5a6' };
      
      elements.push({
        data: {
          id: node.id,
          label: node.id,
          parent: `layer-${layer}`,
          type: node.type || 'unknown',
          hardware: node.hardware || '',
          status: node.status || 'up',
          layer: layer,
          isRoot: node.is_root || false,
          deviceType: node.device_type || node.type
        },
        classes: `device-node ${node.is_root ? 'root-device' : ''}`
      });
    });

    // Add edges
    topology.edges.forEach((edge, index) => {
      elements.push({
        data: {
          id: `edge-${index}`,
          source: edge.source,
          target: edge.target,
          label: edge.local_port && edge.remote_port ? 
            `${edge.local_port} ↔ ${edge.remote_port}` : '',
          status: edge.status || 'up',
          weight: edge.weight || 1
        },
        classes: `edge ${edge.status === 'up' ? 'edge-up' : 'edge-down'}`
      });
    });

    return elements;
  };

  const initializeCytoscape = (elements) => {
    if (cyRef.current) {
      cyRef.current.destroy();
    }

    const cy = cytoscape({
      container: containerRef.current,
      elements: elements,
      
      style: [
        // Layer nodes (compound/parent nodes)
        {
          selector: '.layer-node',
          style: {
            'label': 'data(label)',
            'background-color': '#f8f9fa',
            'border-width': 2,
            'border-color': '#6c757d',
            'border-style': 'dashed',
            'text-valign': 'top',
            'text-halign': 'center',
            'font-size': 14,
            'font-weight': 'bold',
            'color': '#495057',
            'padding': 20,
            'compound-sizing-wrt-labels': 'include',
            'min-width': 200,
            'min-height': 100
          }
        },
        
        // Device nodes
        {
          selector: '.device-node',
          style: {
            'width': 30,
            'height': 30,
            'label': node => {
              const icon = getDeviceIcon(node.data('type'));
              return `${icon}\n${node.data('label')}`;
            },
            'background-color': node => {
              const layer = node.data('layer');
              const config = LAYER_CONFIG[layer] || { color: '#95a5a6' };
              return config.color;
            },
            'border-width': 2,
            'border-color': node => node.data('isRoot') ? '#ff6b6b' : '#fff',
            'border-style': node => node.data('isRoot') ? 'solid' : 'solid',
            'color': '#2c3e50',
            'text-valign': 'bottom',
            'text-halign': 'center',
            'font-size': 8,
            'text-wrap': 'wrap',
            'text-max-width': 80,
            'overlay-opacity': 0.1,
            'transition-property': 'background-color, border-color, overlay-opacity',
            'transition-duration': '0.3s'
          }
        },
        
        // Root devices
        {
          selector: '.root-device',
          style: {
            'width': 40,
            'height': 40,
            'border-width': 3,
            'border-color': '#ff6b6b',
            'box-shadow': '0 0 10px rgba(255, 107, 107, 0.5)'
          }
        },
        
        // Edges
        {
          selector: '.edge',
          style: {
            'width': 2,
            'line-color': '#95a5a6',
            'target-arrow-color': '#95a5a6',
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier',
            'arrow-scale': 1.2,
            'label': 'data(label)',
            'font-size': 6,
            'text-rotation': 'autorotate',
            'text-margin-y': -10,
            'edge-text-rotation': 'autorotate'
          }
        },
        
        // Active/up edges
        {
          selector: '.edge-up',
          style: {
            'line-color': '#2ecc71',
            'target-arrow-color': '#2ecc71'
          }
        },
        
        // Down edges
        {
          selector: '.edge-down',
          style: {
            'line-color': '#e74c3c',
            'target-arrow-color': '#e74c3c',
            'line-style': 'dashed'
          }
        },
        
        // Hover states
        {
          selector: 'node:active',
          style: {
            'overlay-opacity': 0.3,
            'overlay-color': '#3498db'
          }
        },
        
        // Selected state
        {
          selector: '.selected',
          style: {
            'border-width': 4,
            'border-color': '#3498db',
            'box-shadow': '0 0 15px rgba(52, 152, 219, 0.7)'
          }
        }
      ],

      layout: {
        name: 'cose-bilkent',
        animate: true,
        animationDuration: 2000,
        fit: true,
        padding: 50,
        nodeDimensionsIncludeLabels: true,
        uniformNodeDimensions: false,
        packComponents: true,
        stepSize: 10,
        
        // Hierarchical layout options
        idealEdgeLength: 150,
        edgeElasticity: 0.45,
        nestingFactor: 0.1,
        gravity: 0.8,
        numIter: 2500,
        tilingPaddingVertical: 100,
        tilingPaddingHorizontal: 100,
        
        // Forces for compound nodes
        gravityRangeCompound: 1.5,
        gravityCompound: 1.0,
        gravityRange: 3.8
      },

      // Interaction options
      wheelSensitivity: 0.1,
      minZoom: 0.1,
      maxZoom: 3,
      boxSelectionEnabled: true,
      autoungrabify: false,
      autounselectify: false
    });

    // Event handlers
    cy.on('tap', 'node[!isParent]', (evt) => {
      const node = evt.target;
      const deviceId = node.id();
      
      // Remove previous selection
      cy.elements('.selected').removeClass('selected');
      
      // Add selection to clicked node
      node.addClass('selected');
      
      if (onDeviceSelect) {
        onDeviceSelect(deviceId);
      }
    });

    // Highlight selected device
    if (selectedDevice) {
      cy.elements('.selected').removeClass('selected');
      const node = cy.getElementById(selectedDevice);
      if (node.length > 0) {
        node.addClass('selected');
        cy.center(node);
      }
    }

    // Store reference
    cyRef.current = cy;

    // Update stats
    const nodeCount = cy.nodes('[!isParent]').length;
    const edgeCount = cy.edges().length;
    const layerCount = cy.nodes('[isParent]').length;
    
    setStats({
      nodes: nodeCount,
      edges: edgeCount,
      layers: layerCount
    });

    setLoading(false);

    // Fit to viewport after layout
    cy.ready(() => {
      cy.fit();
    });

    return cy;
  };

  useEffect(() => {
    if (topology && topology.nodes && topology.edges && containerRef.current) {
      setLoading(true);
      const elements = buildCytoscapeElements(topology);
      initializeCytoscape(elements);
    }

    return () => {
      if (cyRef.current) {
        cyRef.current.destroy();
        cyRef.current = null;
      }
    };
  }, [topology]);

  useEffect(() => {
    if (cyRef.current && selectedDevice) {
      cyRef.current.elements('.selected').removeClass('selected');
      const node = cyRef.current.getElementById(selectedDevice);
      if (node.length > 0) {
        node.addClass('selected');
        cyRef.current.center(node);
      }
    }
  }, [selectedDevice]);

  const handleFitToView = () => {
    if (cyRef.current) {
      cyRef.current.fit();
    }
  };

  const handleResetLayout = () => {
    if (cyRef.current) {
      cyRef.current.layout({
        name: 'cose-bilkent',
        animate: true,
        animationDuration: 2000,
        fit: true,
        padding: 50,
        nodeDimensionsIncludeLabels: true,
        uniformNodeDimensions: false,
        packComponents: true,
        stepSize: 10,
        
        // Hierarchical layout options
        idealEdgeLength: 150,
        edgeElasticity: 0.45,
        nestingFactor: 0.1,
        gravity: 0.8,
        numIter: 2500,
        tilingPaddingVertical: 100,
        tilingPaddingHorizontal: 100,
        
        // Forces for compound nodes
        gravityRangeCompound: 1.5,
        gravityCompound: 1.0,
        gravityRange: 3.8
      }).run();
    }
  };

  return (
    <div className="cytoscape-topology">
      <div className="topology-controls">
        <div className="stats">
          <span>Nodes: {stats.nodes}</span>
          <span>Edges: {stats.edges}</span>
          <span>Layers: {stats.layers}</span>
        </div>
        <div className="controls">
          <button onClick={handleFitToView} title="Fit to view">
            🔍 Fit
          </button>
          <button onClick={handleResetLayout} title="Reset layout">
            🔄 Reset
          </button>
        </div>
      </div>
      
      <div 
        ref={containerRef} 
        className="cytoscape-container"
        style={{ 
          width: '100%', 
          height: '600px', 
          border: '1px solid #ddd',
          borderRadius: '8px',
          backgroundColor: '#f8f9fa'
        }}
      />
      
      {loading && (
        <div className="loading-overlay">
          <div className="loading-spinner">Loading topology...</div>
        </div>
      )}
    </div>
  );
};

export default CytoscapeTopology;