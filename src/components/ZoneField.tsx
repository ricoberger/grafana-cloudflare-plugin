import React from 'react';
import { Combobox, ComboboxOption, Field } from '@grafana/ui';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';

interface Props {
  datasource: DataSource;
  zone?: string;
  onZoneChange: (value: string) => void;
}

export function ZoneField({ datasource, zone, onZoneChange }: Props) {
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
