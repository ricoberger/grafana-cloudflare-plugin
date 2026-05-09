import { Combobox, ComboboxOption, Field, InlineField } from '@grafana/ui';
import React from 'react';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';

interface Props {
  datasource: DataSource;
  zone?: string;
  onZoneChange: (value: string) => void;
  isInline?: boolean;
}

export function ZoneField({ datasource, zone, onZoneChange, isInline }: Props) {
  const state = useAsync(async (): Promise<ComboboxOption[]> => {
    const result = await datasource.metricFindQuery({
      refId: 'zones',
      queryType: 'zones',
    });

    const zones = result.map((value) => {
      return { value: value.value as string, label: value.text };
    });
    return zones;
  }, [datasource]);

  if (isInline) {
    return (
      <InlineField label="Zone" labelWidth={25}>
        <Combobox<string>
          width={25}
          value={zone}
          createCustomValue={true}
          options={state.value || []}
          onChange={(option: ComboboxOption<string>) => {
            onZoneChange(option.value);
          }}
        />
      </InlineField>
    );
  }

  return (
    <Field label="Zone">
      <Combobox<string>
        width={25}
        value={zone}
        createCustomValue={true}
        options={state.value || []}
        onChange={(option: ComboboxOption<string>) => {
          onZoneChange(option.value);
        }}
      />
    </Field>
  );
}
