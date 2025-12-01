import { useState, useEffect, useCallback } from 'react';
import { apiService } from '../../../services/api';
import { log } from '../../../utils/logger';
import { DatabaseSchema } from '../types';

export const useDatabaseSchema = () => {
  const [schema, setSchema] = useState<DatabaseSchema | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadDatabaseSchema = async () => {
      try {
        setLoading(true);
        const data = await apiService.getDatabaseSchema();

        // 如果data为null或undefined，设置默认值
        if (!data) {
          setSchema({
            name: 'unknown',
            type: 'unknown',
            tables: [],
            relationships: []
          });
          setError(null);
          setLoading(false);
          return;
        }

        // 转换后端数据格式到前端格式
        const transformedSchema = {
          name: data.name || 'unknown',
          type: data.type || 'unknown',
          tables: (data.tables || []).map((table: any) => ({
            id: table.name,
            name: table.name,
            type: 'table' as const,
            schema: data.name,
            rowCount: table.rowCount,
            size: table.size,
            columns: (table.columns || []).map((col: any) => ({
              id: `${table.name}.${col.name}`,
              name: col.name,
              type: col.type,
              nullable: col.nullable,
              primaryKey: col.primaryKey,
              unique: col.unique,
              defaultValue: col.defaultValue,
              description: col.description,
              position: { x: 0, y: 0 },
            })),
            indexes: (table.indexes || []).map((idx: any) => ({
              id: `${table.name}.${idx.name}`,
              name: idx.name,
              columns: idx.columns || [],
              unique: idx.unique,
              type: idx.type,
            })) || [],
            foreignKeys: [],
            position: { x: 0, y: 0 },
          })),
          relationships: (data.relationships || []).map((rel: any) => ({
            id: rel.name || `${rel.sourceTable}_${rel.targetTable}`,
            source: rel.sourceTable,
            target: rel.targetTable,
            type: 'foreign_key' as const,
            label: `${rel.sourceColumn} → ${rel.targetColumn}`,
            style: {
              color: '#1890ff',
              width: 2,
              style: 'solid' as const,
              arrowType: 'arrow' as const,
            },
          })) || [],
        };

        setSchema(transformedSchema);
        setError(null);
      } catch (err) {
        log.error('数据库表结构加载失败', err, 'database', 'Dashboard', err instanceof Error ? err.stack : undefined);
        setError(err instanceof Error ? err.message : '加载数据库表结构失败');
      } finally {
        setLoading(false);
      }
    };

    loadDatabaseSchema();
  }, []);

  const onTableSelect = useCallback((tableName: string) => {
    log.info(`用户选择数据库表: ${tableName}`, { tableName }, 'database', 'Dashboard');
    // 可以在这里添加表详情处理逻辑
  }, []);

  return { schema, loading, error, onTableSelect };
};