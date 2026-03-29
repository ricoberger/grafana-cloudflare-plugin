import { QueryEditorProps } from '@grafana/data';
import {
  Combobox,
  ComboboxOption,
  InlineField,
  InlineFieldRow,
} from '@grafana/ui';
import React from 'react';

import { DataSource } from '../datasource';
import { DEFAULT_QUERIES, Options, Query, QueryType } from '../types';

interface Props extends QueryEditorProps<DataSource, any, Options, Query> { }

export function VariableQueryEditor({ query, onChange, onRunQuery }: Props) {
  return (
    <>
      <InlineFieldRow>
        <InlineField label="Variable Type" labelWidth={25}>
          <Combobox<QueryType>
            value={query.queryType}
            options={[
              {
                label: 'Zones',
                value: 'zones',
              },
            ]}
            onChange={(option: ComboboxOption<QueryType>) => {
              onChange({
                ...query,
                ...DEFAULT_QUERIES[option.value],
                queryType: option.value,
              });
              onRunQuery();
            }}
          />
        </InlineField>
      </InlineFieldRow>
    </>
  );
}
