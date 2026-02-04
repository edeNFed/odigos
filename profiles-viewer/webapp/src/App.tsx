import { useState, useEffect } from 'react';
import styled, { ThemeProvider, createGlobalStyle } from 'styled-components';
import { AppSelector } from './components/AppSelector';
import { TimeRangeSelector } from './components/TimeRangeSelector';
import { FlameGraph } from './components/FlameGraph';
import { AppInfo, BucketInfo, FlameGraphNode } from './api/types';
import { fetchApps, fetchBuckets, fetchFlameGraph } from './api/client';

const darkTheme = {
  colors: {
    primary: '#7B61FF',
    background: '#0D0D0D',
    surface: '#1A1A1A',
    surfaceLight: '#2A2A2A',
    text: '#FFFFFF',
    textSecondary: '#A0A0A0',
    border: '#333333',
    error: '#FF6B6B',
    success: '#4CAF50',
  },
};

const GlobalStyle = createGlobalStyle`
  body {
    background-color: ${(props) => props.theme.colors.background};
    color: ${(props) => props.theme.colors.text};
  }
`;

const Container = styled.div`
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  padding: 24px;
`;

const Header = styled.header`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid ${(props) => props.theme.colors.border};
`;

const Title = styled.h1`
  font-size: 24px;
  font-weight: 600;
  color: ${(props) => props.theme.colors.text};
`;

const Controls = styled.div`
  display: flex;
  gap: 16px;
  align-items: center;
`;

const MainContent = styled.main`
  flex: 1;
  display: flex;
  flex-direction: column;
`;

const EmptyState = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: ${(props) => props.theme.colors.textSecondary};
  font-size: 16px;
`;

const ErrorMessage = styled.div`
  color: ${(props) => props.theme.colors.error};
  padding: 16px;
  background: ${(props) => props.theme.colors.surface};
  border-radius: 8px;
  margin-bottom: 16px;
`;

const LoadingSpinner = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: ${(props) => props.theme.colors.textSecondary};
`;

function App() {
  const [apps, setApps] = useState<AppInfo[]>([]);
  const [selectedApp, setSelectedApp] = useState<AppInfo | null>(null);
  const [buckets, setBuckets] = useState<BucketInfo[]>([]);
  const [selectedTimeRange, setSelectedTimeRange] = useState<{ start?: string; end?: string }>({});
  const [flameGraphData, setFlameGraphData] = useState<FlameGraphNode | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch apps on mount
  useEffect(() => {
    const loadApps = async () => {
      try {
        const data = await fetchApps();
        setApps(data.apps || []);
      } catch (err) {
        setError('Failed to load applications');
        console.error(err);
      }
    };
    loadApps();

    // Poll for new apps every 10 seconds
    const interval = setInterval(loadApps, 10000);
    return () => clearInterval(interval);
  }, []);

  // Fetch buckets when app is selected
  useEffect(() => {
    if (!selectedApp) {
      setBuckets([]);
      return;
    }

    const loadBuckets = async () => {
      try {
        const { namespace, kind, name } = selectedApp.appId;
        const data = await fetchBuckets(namespace, kind, name);
        setBuckets(data.buckets || []);
      } catch (err) {
        setError('Failed to load time buckets');
        console.error(err);
      }
    };
    loadBuckets();

    // Poll for new buckets every 5 seconds
    const interval = setInterval(loadBuckets, 5000);
    return () => clearInterval(interval);
  }, [selectedApp]);

  // Fetch flame graph when time range changes
  useEffect(() => {
    if (!selectedApp) {
      setFlameGraphData(null);
      return;
    }

    const loadFlameGraph = async () => {
      setLoading(true);
      setError(null);
      try {
        const { namespace, kind, name } = selectedApp.appId;
        const data = await fetchFlameGraph(
          namespace,
          kind,
          name,
          selectedTimeRange.start,
          selectedTimeRange.end
        );
        setFlameGraphData(data);
      } catch (err) {
        setError('Failed to load flame graph data');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    loadFlameGraph();
  }, [selectedApp, selectedTimeRange]);

  const handleAppSelect = (app: AppInfo | null) => {
    setSelectedApp(app);
    setSelectedTimeRange({});
    setFlameGraphData(null);
  };

  const handleTimeRangeSelect = (start?: string, end?: string) => {
    setSelectedTimeRange({ start, end });
  };

  return (
    <ThemeProvider theme={darkTheme}>
      <GlobalStyle />
      <Container>
        <Header>
          <Title>Continuous Profiling</Title>
          <Controls>
            <AppSelector
              apps={apps}
              selectedApp={selectedApp}
              onSelect={handleAppSelect}
            />
            {selectedApp && (
              <TimeRangeSelector
                buckets={buckets}
                onSelect={handleTimeRangeSelect}
              />
            )}
          </Controls>
        </Header>
        <MainContent>
          {error && <ErrorMessage>{error}</ErrorMessage>}
          {!selectedApp && (
            <EmptyState>
              <p>Select an application to view its CPU profile</p>
            </EmptyState>
          )}
          {selectedApp && loading && (
            <LoadingSpinner>Loading profile data...</LoadingSpinner>
          )}
          {selectedApp && !loading && flameGraphData && (
            <FlameGraph data={flameGraphData} />
          )}
          {selectedApp && !loading && !flameGraphData && !error && (
            <EmptyState>
              <p>No profile data available for the selected time range</p>
            </EmptyState>
          )}
        </MainContent>
      </Container>
    </ThemeProvider>
  );
}

export default App;
