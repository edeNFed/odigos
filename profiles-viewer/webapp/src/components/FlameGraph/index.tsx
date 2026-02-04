import { useEffect, useRef, useState } from 'react';
import styled from 'styled-components';
import { flamegraph } from 'd3-flame-graph';
import { select } from 'd3-selection';
import 'd3-flame-graph/dist/d3-flamegraph.css';

const Container = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  background: ${(props) => props.theme.colors.surface};
  border-radius: 8px;
  padding: 16px;
  overflow: hidden;
  min-height: 400px;
`;

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`;

const Title = styled.h2`
  font-size: 16px;
  font-weight: 500;
  color: ${(props) => props.theme.colors.text};
`;

const Controls = styled.div`
  display: flex;
  gap: 8px;
`;

const Button = styled.button`
  padding: 6px 12px;
  border-radius: 4px;
  border: 1px solid ${(props) => props.theme.colors.border};
  background: ${(props) => props.theme.colors.surfaceLight};
  color: ${(props) => props.theme.colors.text};
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    background: ${(props) => props.theme.colors.primary};
    border-color: ${(props) => props.theme.colors.primary};
  }
`;

const ChartContainer = styled.div`
  flex: 1;
  overflow: auto;
  min-height: 300px;

  .d3-flame-graph rect {
    stroke: ${(props) => props.theme.colors.background};
    stroke-width: 1px;
  }

  .d3-flame-graph-tip {
    background: ${(props) => props.theme.colors.surfaceLight};
    color: ${(props) => props.theme.colors.text};
    border: 1px solid ${(props) => props.theme.colors.border};
    padding: 8px;
    border-radius: 4px;
    font-size: 12px;
    max-width: 400px;
    word-wrap: break-word;
  }
`;

const ErrorMessage = styled.div`
  color: ${(props) => props.theme.colors.error};
  padding: 16px;
`;

interface FlameGraphNode {
  name: string;
  value: number;
  children?: FlameGraphNode[];
}

interface FlameGraphProps {
  data: FlameGraphNode;
}

export function FlameGraph({ data }: FlameGraphProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const chartRef = useRef<any>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!containerRef.current || !data) return;

    // Reset error state
    setError(null);

    try {
      // Clear previous chart
      containerRef.current.innerHTML = '';

      // Calculate dimensions - ensure minimum width
      const width = Math.max(containerRef.current.clientWidth, 600);

      // Create flame graph
      const chart = flamegraph()
        .width(width)
        .cellHeight(18)
        .transitionDuration(300)
        .minFrameSize(5)
        .inverted(false)
        .sort(true)
        .title('')
        .selfValue(false);

      // Store chart reference for reset
      chartRef.current = chart;

      // Render
      select(containerRef.current)
        .datum(data)
        .call(chart);

      // Handle resize
      const handleResize = () => {
        if (containerRef.current && chartRef.current) {
          const newWidth = Math.max(containerRef.current.clientWidth, 600);
          chartRef.current.width(newWidth);
          select(containerRef.current)
            .datum(data)
            .call(chartRef.current);
        }
      };

      window.addEventListener('resize', handleResize);

      return () => {
        window.removeEventListener('resize', handleResize);
      };
    } catch (err) {
      console.error('FlameGraph error:', err);
      setError(err instanceof Error ? err.message : 'Failed to render flame graph');
    }
  }, [data]);

  const handleReset = () => {
    if (chartRef.current) {
      chartRef.current.resetZoom();
    }
  };

  if (error) {
    return (
      <Container>
        <Header>
          <Title>CPU Flame Graph</Title>
        </Header>
        <ErrorMessage>Error: {error}</ErrorMessage>
      </Container>
    );
  }

  return (
    <Container>
      <Header>
        <Title>CPU Flame Graph</Title>
        <Controls>
          <Button onClick={handleReset}>Reset Zoom</Button>
        </Controls>
      </Header>
      <ChartContainer ref={containerRef} />
    </Container>
  );
}
