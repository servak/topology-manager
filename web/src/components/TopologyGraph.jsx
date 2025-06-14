import React, { useEffect, useRef } from 'react'
import cytoscape from 'cytoscape'
import dagre from 'cytoscape-dagre'
import coseBilkent from 'cytoscape-cose-bilkent'

cytoscape.use(dagre)
cytoscape.use(coseBilkent)

const DEVICE_COLORS = {
  internet: '#e74c3c',
  firewall: '#e67e22',
  router: '#f39c12',
  core: '#3498db',
  distribution: '#2ecc71',
  access: '#9b59b6',
  server: '#95a5a6',
  group: '#95a5a6',
  unknown: '#95a5a6'
}

const LAYER_POSITIONS = {
  0: 100,   // internet
  1: 200,   // firewall
  2: 300,   // router
  3: 400,   // core
  4: 500,   // distribution
  5: 600,   // access
  6: 700    // server
}

function TopologyGraph({ topology, onObjectSelect, onGroupExpand }) {
  const cyRef = useRef(null)
  const containerRef = useRef(null)

  useEffect(() => {
    if (!topology || !containerRef.current) return

    console.log('TopologyGraph re-rendering with topology:', {
      nodes: topology.nodes.length,
      edges: topology.edges.length,
      groups: topology.groups?.length || 0
    })

    if (cyRef.current) {
      cyRef.current.destroy()
    }

    const elements = [
      ...topology.nodes.map(node => ({
        data: {
          id: node.id,  // node.name ではなく node.id を使用
          label: node.name,
          type: node.type,
          hardware: node.hardware,
          status: node.status,
          layer: node.layer,
          isRoot: node.is_root
        }
      })),
      ...topology.edges.map(edge => ({
        data: {
          id: edge.id,  // エッジIDも統一
          source: edge.source,
          target: edge.target,
          localPort: edge.local_port,
          remotePort: edge.remote_port,
          status: edge.status
        }
      }))
    ]

    console.log('Cytoscape elements:', {
      nodes: elements.filter(e => !e.data.source).length,
      edges: elements.filter(e => e.data.source).length,
      sampleNode: elements.find(e => !e.data.source),
      sampleEdge: elements.find(e => e.data.source)
    })

    const cy = cytoscape({
      container: containerRef.current,
      elements: elements,
      style: [
        {
          selector: 'node',
          style: {
            'background-color': (ele) => DEVICE_COLORS[ele.data('type')] || DEVICE_COLORS.unknown,
            'label': 'data(label)',
            'text-valign': 'center',
            'text-halign': 'center',
            'color': '#fff',
            'font-size': '12px',
            'font-weight': 'bold',
            'text-outline-width': 2,
            'text-outline-color': '#000',
            'width': (ele) => {
              if (ele.data('isRoot')) return 60
              if (ele.data('type') === 'group') return 80
              return 50
            },
            'height': (ele) => {
              if (ele.data('isRoot')) return 60
              if (ele.data('type') === 'group') return 50
              return 50
            },
            'shape': (ele) => ele.data('type') === 'group' ? 'round-rectangle' : 'ellipse',
            'border-width': (ele) => ele.data('isRoot') ? 4 : 2,
            'border-color': (ele) => ele.data('isRoot') ? '#f39c12' : '#333'
          }
        },
        {
          selector: 'node[type="group"]',
          style: {
            'background-color': '#95a5a6',
            'border-color': '#7f8c8d',
            'border-width': 3,
            'font-size': '10px',
            'text-wrap': 'wrap',
            'text-max-width': '70px'
          }
        },
        {
          selector: 'edge',
          style: {
            'width': 2,
            'line-color': '#ccc',
            'target-arrow-color': '#ccc',
            'target-arrow-shape': 'triangle',
            'curve-style': 'bezier'
          }
        },
        {
          selector: 'node:selected',
          style: {
            'border-width': 4,
            'border-color': '#f39c12',
            'background-color': (ele) => {
              const color = DEVICE_COLORS[ele.data('type')] || DEVICE_COLORS.unknown
              return color
            }
          }
        },
        {
          selector: 'edge:selected',
          style: {
            'width': 4,
            'line-color': '#f39c12',
            'target-arrow-color': '#f39c12'
          }
        }
      ],
      layout: {
        name: 'dagre',
        rankDir: 'TB',
        nodeSep: 100,
        edgeSep: 50,
        rankSep: 150
      }
    })

    cy.on('tap', 'node', function(evt) {
      const node = evt.target
      const data = node.data()
      
      console.log('Node clicked:', data)
      
      if (onObjectSelect) {
        onObjectSelect({
          type: 'node',
          data: data
        })
      }
    })

    // グループノードのダブルクリック処理
    cy.on('dblclick', 'node[type="group"]', function(evt) {
      const node = evt.target
      const data = node.data()
      
      console.log('Group double-clicked for expansion:', data)
      
      // グループ展開を実行
      if (onGroupExpand) {
        onGroupExpand(data)
      }
    })

    cy.on('tap', 'edge', function(evt) {
      const edge = evt.target
      const data = edge.data()
      
      console.log('Edge clicked:', data)
      
      if (onObjectSelect) {
        onObjectSelect({
          type: 'edge',
          data: data
        })
      }
    })

    // 背景クリックで選択解除
    cy.on('tap', function(evt) {
      if (evt.target === cy && onObjectSelect) {
        onObjectSelect(null)
      }
    })

    cyRef.current = cy

    return () => {
      if (cyRef.current) {
        cyRef.current.destroy()
      }
    }
  }, [topology])

  return (
    <div className="topology-graph">
      <div 
        ref={containerRef} 
        id="topology-cy" 
        style={{ 
          width: '100%', 
          height: '600px',
          minHeight: '600px'
        }} 
      />
    </div>
  )
}

export default TopologyGraph