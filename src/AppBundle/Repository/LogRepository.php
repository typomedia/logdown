<?php

namespace AppBundle\Repository;

use AppBundle\AppBundle;
use AppBundle\Helper\DateHelper;
use Doctrine\DBAL\Exception;
use Doctrine\ORM\EntityManagerInterface;
use Psr\Cache\InvalidArgumentException;
use ReflectionClass;
use ReflectionException;
use Symfony\Component\Cache\Adapter\AdapterInterface;
use Symfony\Component\Config\FileLocator;
use Symfony\Component\DependencyInjection\ParameterBag\ParameterBag;
use Symfony\Component\HttpFoundation\RequestStack;
use Symfony\Contracts\Cache\CacheInterface;

class LogRepository
{
    public const FETCH_ALL = 0;
    public const FETCH_ONE = 1;

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
    /**
     * @var RequestStack
     */
    private $request;

    public function __construct(
        EntityManagerInterface $entity,
        ParameterBag $param,
        RequestStack $request,
        CacheInterface $cache
    ) {
        $this->entity = $entity;
        $this->cache = $cache;
        $this->param = $param;
        $this->request = $request;
        $this->path = __DIR__ . '/Queries/';
    }

    /**
     * @param string $name
     * @param int $mode [optional]
     * {@see FETCH_ALL} Fetch all assoc
     * {@see FETCH_ONE} Fetch one assoc
     * @return array|false|mixed
     * @throws Exception
     * @throws ReflectionException
     * @throws InvalidArgumentException
     * @throws \Doctrine\DBAL\Driver\Exception
     */
    public function get(string $name, int $mode = 0)
    {
        $query = $this->getContent($name);
        $variables = $this->getVariables($name);

        $requests = $this->filter($this->request->getCurrentRequest()->query->all(), $variables);

        $stmt = $this->entity->getConnection()->prepare($query);

        foreach ($requests as $key => $value) {
            $key = strtolower($key);
            $value = trim($value);

            switch ($key) {
                case 'date':
                    $stmt->bindValue($key, $value . '%');
                    $format = AppBundle::VIEWFORMAT[$mode];
                    $view = DateHelper::analyzeDate($value);
                    $stmt->bindValue('view', $format[$view]);
                    break;
                case 'search':
                    $stmt->bindValue($key, '%' . $value . '%');
                    break;
                default:
                    $stmt->bindValue($key, $value);
                    break;
            }
        }

        $key = md5($query . serialize($this->getProperty($stmt, 'params')));
        $item = $this->cache->getItem($key);

        if ($item->isHit() && $this->param->get('cache')) {
            $result = $item->get();
        } else {
            $result = $mode ?
                $stmt->executeQuery()->fetchAssociative() :
                $stmt->executeQuery()->fetchAllAssociative();
            $item->set($result);
            $this->cache->save($item);
        }

        return $result;
    }

    /**
     * @param string $name
     * @return false|string
     */
    public function getContent(string $name)
    {
        $file = new FileLocator($this->path);
        $file = $file->locate($name);

        return file_get_contents($file);
    }

    /**
     * @param string $file
     * @return array|mixed
     */
    public function getVariables(string $file)
    {
        $pattern = '/:(\w*)/';
        preg_match_all($pattern, $this->getContent($file), $matches);

        return $matches[1] ?? [];
    }

    /**
     * @param object $object
     * @param string $name
     * @return mixed
     * @throws ReflectionException
     */
    public function getProperty(object $object, string $name)
    {
        $reflection = new ReflectionClass($object);
        $property = $reflection->getProperty($name);
        $property->setAccessible(true);

        return $property->getValue($object);
    }

    /**
     * @param array $assoc
     * @param array $list
     * @return array
     */
    public function filter(array $assoc, array $list)
    {
        $data = [];
        foreach ($list as $value) {
            $data[$value] = $assoc[$value] ?? '';
        }
        return $data;
    }
}
