<?php
/**
 * @see       https://github.com/zendframework/zend-expressive-helpers for the canonical source repository
 * @copyright Copyright (c) 2015-2017 Zend Technologies USA Inc. (https://www.zend.com)
 * @license   https://github.com/zendframework/zend-expressive-helpers/blob/master/LICENSE.md New BSD License
 */

declare(strict_types=1);

namespace Zend\Expressive\Helper;

use Psr\Container\ContainerInterface;

use function sprintf;

class ServerUrlMiddlewareFactory
{
    /**
     * Create a ServerUrlMiddleware instance.
     *
     * @throws Exception\MissingHelperException
     */
    public function __invoke(ContainerInterface $container) : ServerUrlMiddleware
    {
        if (! $container->has(ServerUrlHelper::class)) {
            throw new Exception\MissingHelperException(sprintf(
                '%s requires a %s service at instantiation; none found',
                ServerUrlMiddleware::class,
                ServerUrlHelper::class
            ));
        }

        return new ServerUrlMiddleware($container->get(ServerUrlHelper::class));
    }
}
