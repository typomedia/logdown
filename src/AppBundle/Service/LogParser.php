<?php

namespace AppBundle\Service;

use Generator;
use RuntimeException;

/**
 * Parses the W3C Extended Log File Format used by IIS.
 */
class LogParser
{
    /**
     * Maps W3C extended log field identifiers to Log table columns.
     * Fields not listed here (e.g. sc-bytes, cs-bytes) are ignored.
     * The "date" and "time" fields are handled separately and merged
     * into the "date" column.
     */
    public const FIELD_MAP = [
        's-ip'            => 'server',
        'cs-method'       => 'method',
        'cs-uri-stem'     => 'request',
        'cs-uri-query'    => 'param',
        's-port'          => 'port',
        'cs-username'     => 'user',
        'c-ip'            => 'client',
        'cs(User-Agent)'  => 'agent',
        'cs(Referer)'     => 'referer',
        'sc-status'       => 'status',
        'sc-substatus'    => 'substatus',
        'sc-win32-status' => 'win32',
        'time-taken'      => 'duration',
    ];

    /**
     * Field identifiers in column order, taken from the "#Fields:" directive.
     *
     * @var string[]
     */
    private $fields = [];

    /**
     * Parse a whole logfile, yielding one associative row per data line.
     *
     * @param string $path
     * @return Generator
     */
    public function parseFile(string $path): Generator
    {
        $handle = @fopen($path, 'r');
        if ($handle === false) {
            throw new RuntimeException(sprintf('Cannot open log file "%s".', $path));
        }

        try {
            while (($line = fgets($handle)) !== false) {
                $row = $this->parseLine($line);
                if ($row !== null) {
                    yield $row;
                }
            }
        } finally {
            fclose($handle);
        }
    }

    /**
     * Parse a single line.
     *
     * Directive lines (starting with "#") update the parser state and return
     * null. The "#Fields:" directive defines the column order used for all
     * following data lines. Data lines return an associative row keyed by Log
     * column name, or null when no "#Fields:" directive has been seen yet.
     *
     * @param string $line
     * @return array|null
     */
    public function parseLine(string $line): ?array
    {
        $line = trim($line);
        if ($line === '') {
            return null;
        }

        if ($line[0] === '#') {
            if (stripos($line, '#Fields:') === 0) {
                $this->setFields(substr($line, strlen('#Fields:')));
            }

            return null;
        }

        if ($this->fields === []) {
            return null;
        }

        $values = preg_split('/\s+/', $line) ?: [];

        return $this->mapValues($values);
    }

    /**
     * Define the column order from the content of a "#Fields:" directive.
     *
     * @param string $directive
     * @return void
     */
    public function setFields(string $directive): void
    {
        $this->fields = preg_split('/\s+/', trim($directive), -1, PREG_SPLIT_NO_EMPTY) ?: [];
    }

    /**
     * The field identifiers currently in effect, in column order.
     *
     * @return string[]
     */
    public function getFields(): array
    {
        return $this->fields;
    }

    /**
     * Map a row of raw values to an associative row keyed by Log column name.
     *
     * @param string[] $values
     * @return array
     */
    private function mapValues(array $values): array
    {
        $row = [];
        $date = null;
        $time = null;

        foreach ($this->fields as $index => $field) {
            $value = $values[$index] ?? '-';

            if ($field === 'date') {
                $date = $value;
                continue;
            }

            if ($field === 'time') {
                $time = $value;
                continue;
            }

            if (isset(self::FIELD_MAP[$field])) {
                $row[self::FIELD_MAP[$field]] = $value;
            }
        }

        if ($date !== null || $time !== null) {
            $row['date'] = trim($date . ' ' . $time);
        }

        return $row;
    }
}
