import { QueryEditorProps } from '@grafana/data';
import {
  Combobox,
  ComboboxOption,
  InlineField,
  InlineFieldRow,
  Input,
} from '@grafana/ui';
import React, { ChangeEvent } from 'react';

import { DataSource } from '../datasource';
import { DEFAULT_QUERIES, Options, Query, QueryType } from '../types';
import {
  getAggregationOptions,
  getFiltersOptions,
  nameOptions,
} from '../utils';
import { ZoneField } from './ZoneField';

interface Props extends QueryEditorProps<DataSource, any, Options, Query> { }

export function VariableQueryEditor({
  datasource,
  query,
  onChange,
  onRunQuery,
}: Props) {
  return (
    <>
      <InlineFieldRow>
        <InlineField label="Variable Type" labelWidth={25}>
          <Combobox<QueryType>
            width={25}
            value={query.queryType}
            options={[
              {
                label: 'Zones',
                value: 'zones',
              },
              {
                label: 'Filter Values',
                value: 'filtervalues',
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
      {query.queryType === 'filtervalues' && (
        <>
          <InlineFieldRow>
            <ZoneField
              isInline={true}
              datasource={datasource}
              zone={query.zone}
              onZoneChange={(value) => {
                onChange({ ...query, zone: value });
                onRunQuery();
              }}
            />
          </InlineFieldRow>

          <InlineFieldRow>
            <InlineField data-testid="metric" label="Metric" labelWidth={25}>
              <Combobox<string>
                width={25}
                value={query.name}
                options={nameOptions
                  .filter(
                    (name) =>
                      name !== 'httpRequests' && name !== 'firewallEvents',
                  )
                  .map((name) => ({ value: name }))}
                onChange={(option: ComboboxOption<string>) => {
                  const aggregationOptions = getAggregationOptions(
                    option.value,
                  );

                  onChange({
                    ...query,
                    ...DEFAULT_QUERIES['metrics'],
                    name: option.value,
                    aggregation: aggregationOptions
                      ? aggregationOptions[0]
                      : undefined,
                    zone: query.zone,
                    limit: query.limit,
                  });
                  onRunQuery();
                }}
              />
            </InlineField>
          </InlineFieldRow>

          <InlineFieldRow>
            <InlineField data-testid="field" label="Field" labelWidth={25}>
              <Combobox<string>
                width={25}
                placeholder="Field"
                value={query.field}
                options={getFiltersOptions(query.name!).map((field) => ({
                  value: field,
                }))}
                onChange={(option: ComboboxOption<string>) => {
                  onChange({ ...query, field: option.value });
                }}
              />
            </InlineField>
          </InlineFieldRow>

          <InlineFieldRow>
            <InlineField label="Limit" labelWidth={25}>
              <Input
                width={25}
                onChange={(event: ChangeEvent<HTMLInputElement>) => {
                  onChange({
                    ...query,
                    limit: parseInt(event.target.value, 10),
                  });
                }}
                placeholder="100"
                value={query.limit || ''}
              />
            </InlineField>
          </InlineFieldRow>
        </>
      )}
    </>
  );
}
