import React from 'react';
import styled from 'styled-components';
import { AppInfo } from '../../api/types';

const Container = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
`;

const Label = styled.label`
  font-size: 12px;
  color: ${(props) => props.theme.colors.textSecondary};
`;

const Select = styled.select`
  padding: 8px 12px;
  border-radius: 6px;
  border: 1px solid ${(props) => props.theme.colors.border};
  background: ${(props) => props.theme.colors.surface};
  color: ${(props) => props.theme.colors.text};
  font-size: 14px;
  min-width: 250px;
  cursor: pointer;

  &:hover {
    border-color: ${(props) => props.theme.colors.primary};
  }

  &:focus {
    outline: none;
    border-color: ${(props) => props.theme.colors.primary};
    box-shadow: 0 0 0 2px ${(props) => props.theme.colors.primary}33;
  }
`;

interface AppSelectorProps {
  apps: AppInfo[];
  selectedApp: AppInfo | null;
  onSelect: (app: AppInfo | null) => void;
}

export function AppSelector({ apps, selectedApp, onSelect }: AppSelectorProps) {
  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    if (!value) {
      onSelect(null);
      return;
    }

    const app = apps.find(
      (a) => `${a.appId.namespace}/${a.appId.kind}/${a.appId.name}` === value
    );
    onSelect(app || null);
  };

  const selectedValue = selectedApp
    ? `${selectedApp.appId.namespace}/${selectedApp.appId.kind}/${selectedApp.appId.name}`
    : '';

  return (
    <Container>
      <Label>Application</Label>
      <Select value={selectedValue} onChange={handleChange}>
        <option value="">Select an application...</option>
        {apps.map((app) => {
          const key = `${app.appId.namespace}/${app.appId.kind}/${app.appId.name}`;
          return (
            <option key={key} value={key}>
              {app.appId.namespace}/{app.appId.name} ({app.appId.kind})
            </option>
          );
        })}
      </Select>
    </Container>
  );
}
