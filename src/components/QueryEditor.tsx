import React, { ChangeEvent } from 'react';
import {
  Combobox,
  ComboboxOption,
  Field,
  Input,
  Box,
  useStyles2,
  IconButton,
  Stack,
  MultiCombobox,
} from '@grafana/ui';
import { GrafanaTheme2, QueryEditorProps } from '@grafana/data';
import { css } from '@emotion/css';

import { DataSource } from '../datasource';
import { DEFAULT_QUERIES, Options, Query } from '../types';
import {
  filtersOptions,
  getDimensionsOptions,
  getOrderByOptions,
  nameOptions,
} from '../utils';
import { ZoneField } from './ZoneField';

type Props = QueryEditorProps<DataSource, Query, Options>;

export function QueryEditor({
  datasource,
  query,
  onChange,
  onRunQuery,
}: Props) {
  const styles = useStyles2((theme: GrafanaTheme2) => ({
    marginTop: css`
      margin-top: ${theme.spacing(2)};
    `,
  }));

  return (
    <div className={styles.marginTop}>
      <Stack direction="row" gap={1} wrap={true}>
        <Field data-testid="metric" label="Metric">
          <Combobox<string>
            width={25}
            value={query.name}
            options={nameOptions.map((name) => ({ value: name }))}
            onChange={(option: ComboboxOption<string>) => {
              onChange({
                ...query,
                ...DEFAULT_QUERIES['metrics'],
                name: option.value,
                zone: query.zone,
                limit: query.limit,
              });
              onRunQuery();
            }}
          />
        </Field>

        <ZoneField
          datasource={datasource}
          zone={query.zone}
          onZoneChange={(value) => {
            onChange({ ...query, zone: value });
            onRunQuery();
          }}
        />

        {query.name && query.name.split('_')[0] in filtersOptions && (
          <Field label="Filters">
            <Box display="flex" direction="column" grow={0} gap={1}>
              {query.filters?.map((filter, index) => (
                <Box key={filter.field} display="flex" grow={0} gap={1}>
                  <Combobox<string>
                    width={25}
                    placeholder="Field"
                    value={filter.field}
                    options={filtersOptions[query.name!.split('_')[0]].map(
                      (field) => ({
                        value: field,
                      }),
                    )}
                    onChange={(option: ComboboxOption<string>) => {
                      const newFilters = [...(query.filters || [])];
                      newFilters[index] = {
                        ...newFilters[index],
                        field: option.value,
                      };
                      onChange({ ...query, filters: newFilters });
                    }}
                  />
                  <Combobox<string>
                    width={10}
                    value={filter.operator}
                    options={[
                      { value: '=' },
                      { value: '!=' },
                      { value: '>' },
                      { value: '<' },
                      { value: '>=' },
                      { value: '<=' },
                    ]}
                    onChange={(option: ComboboxOption<string>) => {
                      const newFilters = [...(query.filters || [])];
                      newFilters[index] = {
                        ...newFilters[index],
                        operator: option.value,
                      };
                      onChange({ ...query, filters: newFilters });
                    }}
                  />
                  <Input
                    width={25}
                    placeholder="Value"
                    value={filter.value}
                    onChange={(event: ChangeEvent<HTMLInputElement>) => {
                      const newFilters = [...(query.filters || [])];
                      newFilters[index] = {
                        ...newFilters[index],
                        value: event.currentTarget.value,
                      };
                      onChange({ ...query, filters: newFilters });
                    }}
                  />
                  {index === 0 ? (
                    <IconButton
                      name="plus"
                      aria-label="Add Filter"
                      onClick={() => {
                        const newFilters = [
                          ...(query.filters || []),
                          { field: '-', operator: '=', value: '' },
                        ];
                        onChange({ ...query, filters: newFilters });
                      }}
                    />
                  ) : (
                    <IconButton
                      name="minus"
                      aria-label="Remove Filter"
                      onClick={() => {
                        const newFilters = [...(query.filters || [])];
                        newFilters.splice(index, 1);
                        onChange({ ...query, filters: newFilters });
                      }}
                    />
                  )}
                </Box>
              ))}
            </Box>
          </Field>
        )}

        {query.name && query.name.startsWith('httpRequests_') && (
          <Field label="Dimensions">
            <MultiCombobox<string>
              width={25}
              value={query.dimensions}
              options={getDimensionsOptions(query.name).map((dimension) => ({
                value: dimension,
              }))}
              // eslint-disable-next-line @typescript-eslint/array-type
              onChange={(option: ComboboxOption<string>[]) => {
                onChange({
                  ...query,
                  dimensions: Array.from(option.values()).map(
                    (value) => value.value,
                  ),
                });
              }}
            />
          </Field>
        )}

        {query.name && query.name.startsWith('httpRequests_') && (
          <Field label="Order by">
            <MultiCombobox<string>
              width={25}
              value={query.orderBy}
              options={getOrderByOptions(
                query.name,
                query.dimensions || [],
              ).map((orderBy) => ({
                value: orderBy,
              }))}
              // eslint-disable-next-line @typescript-eslint/array-type
              onChange={(option: ComboboxOption<string>[]) => {
                onChange({
                  ...query,
                  orderBy: Array.from(option.values()).map(
                    (value) => value.value,
                  ),
                });
              }}
            />
          </Field>
        )}
      </Stack>

      <Stack direction="row" gap={1} wrap={true}>
        <Field label="Legend">
          <Input
            width={25}
            placeholder="{{label}}"
            value={query.legend || ''}
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              onChange({ ...query, legend: event.currentTarget.value });
            }}
          />
        </Field>
        <Field label="Limit">
          <Input
            width={10}
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              onChange({ ...query, limit: parseInt(event.target.value, 10) });
            }}
            placeholder="100"
            value={query.limit || ''}
          />
        </Field>
      </Stack>
    </div>
  );
}
