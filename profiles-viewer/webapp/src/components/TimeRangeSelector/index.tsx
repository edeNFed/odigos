import React, { useState } from 'react';
import styled from 'styled-components';
import { BucketInfo } from '../../api/types';

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
  min-width: 200px;
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

interface TimeRangeSelectorProps {
  buckets: BucketInfo[];
  onSelect: (start?: string, end?: string) => void;
}

export function TimeRangeSelector({ buckets, onSelect }: TimeRangeSelectorProps) {
  const [selected, setSelected] = useState<string>('all');

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    setSelected(value);

    if (value === 'all') {
      onSelect(undefined, undefined);
      return;
    }

    const bucket = buckets.find(
      (b) => b.startTime === value
    );
    if (bucket) {
      onSelect(bucket.startTime, bucket.endTime);
    }
  };

  const formatTime = (isoString: string) => {
    const date = new Date(isoString);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  };

  return (
    <Container>
      <Label>Time Range</Label>
      <Select value={selected} onChange={handleChange}>
        <option value="all">All Available ({buckets.length} buckets)</option>
        {buckets.map((bucket) => (
          <option key={bucket.startTime} value={bucket.startTime}>
            {formatTime(bucket.startTime)} - {formatTime(bucket.endTime)} ({bucket.sampleCount} samples)
          </option>
        ))}
      </Select>
    </Container>
  );
}
