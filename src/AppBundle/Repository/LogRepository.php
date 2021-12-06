<?php

namespace AppBundle\Repository;

use Doctrine\DBAL\Exception;
use Doctrine\ORM\EntityManagerInterface;
use Psr\Cache\InvalidArgumentException;
use Symfony\Component\Cache\Adapter\AdapterInterface;
use Symfony\Component\DependencyInjection\ParameterBag\ParameterBag;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Contracts\Cache\CacheInterface;

class LogRepository
{
    /**
     * @var string $path
     */
    protected $path;

    /**
     * @var EntityManagerInterface $entity
     */
    private $entity;

    /**
     * @var AdapterInterface $cache
     */
    private $cache;

    /**
     * @var ParameterBag
     */
    private $param;

    public function __construct(
        EntityManagerInterface $entity,
        ParameterBag $param,
        CacheInterface $cache)
    {
        $this->entity = $entity;
        $this->cache = $cache;
        $this->param = $param;
        $this->path = dirname(__DIR__, 3) . '/var/scripts/';

    }

    /**
     * @param $name
     * @return false|string
     */
    public function getContent($name) {
        return file_get_contents($this->path . $name);
    }

    /**
     * @param $file
     * @return array|mixed
     */
    public function getVariables($file) {
        $pattern = '/\s:(\w*)\s/';
        preg_match_all($pattern, $this->getContent($file), $matches);

        return $matches[1] ?? [];
    }

    /**
     * @param array $assoc
     * @param array $list
     * @return array
     */
    public function filter(array $assoc, array $list) {
        $data = [];
        foreach ($list as $value) {
            $data[$value] = $assoc[$value] ?? '';
        }
        return $data;
    }

    /**
     * @param array $assoc
     * @param array $numeric
     * @return array
     */
    public function filter3(array $assoc, array $numeric) {
        return array_intersect_key($assoc, array_flip($numeric));
    }

    public function filter2($array, $allowed) {
        return array_filter(
            $array,
                static function ($key) use ($allowed) {
                    return in_array($key, $allowed, true);
                },
                ARRAY_FILTER_USE_KEY
            );
    }

    /**
     * @param string $name
     * @param Request $request
     * @return array|false|mixed
     * @throws \Doctrine\DBAL\Driver\Exception
     * @throws Exception
     * @throws InvalidArgumentException
     */
    public function get($name, $request) {
        $query = $this->getContent($name);
        $variables = $this->getVariables($name);
        $requests = $this->filter($request->query->all(), $variables);

        $stmt = $this->entity
            ->getConnection()
            ->prepare($query);

        foreach ($requests as $key => $value) {
            $key = strtolower($key);
            $value = trim($value);

            switch ($key) {
                case 'date':
                    $stmt->bindValue($key, $value . '%');
                    break;
                case 'search':
                    $stmt->bindValue($key, '%' . $value . '%');
                    break;
                default:
                    $stmt->bindValue($key, $value);
                    break;
            }
        }

        $key = md5($query . serialize($requests));
        $item = $this->cache->getItem($key);

        if ($item->isHit() && $this->param->get('cache')) {
            $result = $item->get();
        } else {
            $stmt->executeQuery();
            $result = $stmt->fetchAll();
            $item->set($result);
            $this->cache->save($item);
        }

        return $result;
    }
}