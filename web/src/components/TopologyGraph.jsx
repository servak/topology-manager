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

function TopologyGraph({ topology }) {
  const cyRef = useRef(null)
  const containerRef = useRef(null)

  useEffect(() => {
    if (!topology || !containerRef.current) return

    if (cyRef.current) {
      cyRef.current.destroy()
    }

    const elements = [
      ...topology.nodes.map(node => ({
        data: {
          id: node.name,
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
          id: `${edge.source}-${edge.target}`,
          source: edge.source,
          target: edge.target,
          localPort: edge.local_port,
          remotePort: edge.remote_port,
          status: edge.status
        }
      }))
    ]

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
            'width': (ele) => ele.data('isRoot') ? 60 : 50,
            'height': (ele) => ele.data('isRoot') ? 60 : 50,
            'border-width': (ele) => ele.data('isRoot') ? 4 : 2,
            'border-color': (ele) => ele.data('isRoot') ? '#f39c12' : '#333'
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
      
      const info = [
        `Device: ${data.label}`,
        `Type: ${data.type}`,
        `Layer: ${data.layer}`,
        data.hardware ? `Hardware: ${data.hardware}` : '',
        `Status: ${data.status}`,
        data.isRoot ? 'ROOT DEVICE' : ''
      ].filter(Boolean).join('\n')
      
      alert(info)
    })

    cy.on('tap', 'edge', function(evt) {
      const edge = evt.target
      const data = edge.data()
      
      const info = [
        `Connection: ${data.source} → ${data.target}`,
        data.localPort ? `Local Port: ${data.localPort}` : '',
        data.remotePort ? `Remote Port: ${data.remotePort}` : '',
        `Status: ${data.status}`
      ].filter(Boolean).join('\n')
      
      alert(info)
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